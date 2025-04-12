package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/caarlos0/env/v11"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	provisioner "palter.io/paltergeist/provisioner/pulumi"
)

var (
	workspace auto.Workspace
	ctx       context.Context
	cfg       config
)

const ()

type config struct {
	// GCP project name
	ProjectName string `env:"PROJECT_NAME,required"`
	// GCP Billing Account Id
	BillingAccountId string `env:"BILLING_ACCOUNT_ID,required"`
	// GCP Seed Project Id
	SeedProjectId string `env:"SEED_PROJECT_ID,required"`
	// Email address to create as IAM user.
	Email string `env:"EMAIL,required"`
	// GCP region to deploy resources by default.
	GcpRegion string `env:"GCP_REGION,required"`
}

func init() {
	ctx = context.Background()
	workspace = provisioner.Workspace(ctx)
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
}

func createProjectStack(projectName string, billingAccountId string) auto.Stack {

	deployFunc := func(ctx *pulumi.Context) error {

		project, err := organizations.NewProject(ctx, projectName, &organizations.ProjectArgs{
			Name:           pulumi.String(projectName),
			BillingAccount: pulumi.String(billingAccountId),
			DeletionPolicy: pulumi.String("DELETE"),
		})
		if err != nil {
			return nil
		}
		ctx.Export("projectId", project.ProjectId.ToStringOutput())

		for _, conf := range []struct {
			name    string
			service string
		}{
			{
				name:    "storage",
				service: "storage.googleapis.com",
			},
			{
				name:    "iam",
				service: "iam.googleapis.com",
			},
		} {
			if _, err = projects.NewIAMAuditConfig(ctx, fmt.Sprintf("%s-audit", conf.name), &projects.IAMAuditConfigArgs{
				Project: project.ProjectId,
				Service: pulumi.String(conf.service),
				AuditLogConfigs: projects.IAMAuditConfigAuditLogConfigArray{
					&projects.IAMAuditConfigAuditLogConfigArgs{
						LogType: pulumi.String("ADMIN_READ"),
					},
					&projects.IAMAuditConfigAuditLogConfigArgs{
						LogType: pulumi.String("DATA_READ"),
					},
					&projects.IAMAuditConfigAuditLogConfigArgs{
						LogType: pulumi.String("DATA_WRITE"),
					},
				},
			}); err != nil {
				return nil
			}
		}

		for _, api := range []string{"run.googleapis.com", "sql-component.googleapis.com", "sqladmin.googleapis.com", "compute.googleapis.com", "containerregistry.googleapis.com"} {
			if _, err := projects.NewService(ctx, api, &projects.ServiceArgs{
				DisableOnDestroy: pulumi.Bool(true),
				Project:          project.ProjectId,
				Service:          pulumi.String(api),
			}); err != nil {
				return err
			}
		}
		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, "bootstrap", projectName, deployFunc, auto.WorkDir(workspace.WorkDir()))
	if err != nil {
		log.Fatalf("failed to create stack: %v", err)
	}
	if _, err := stack.Refresh(ctx); err != nil {
		log.Fatalf("failed to refresh stack: %v", err)
	}
	return stack
}

func createAppStack(projectName string, projectId string) auto.Stack {

	deployFunc := func(ctx *pulumi.Context) error {
		// create application buckets
		for _, bucketName := range []string{"data", "analytics", "report"} {
			bucket, err := storage.NewBucket(ctx, bucketName, &storage.BucketArgs{
				Name:                     pulumi.String(fmt.Sprintf("nomaladies-%s", bucketName)),
				Location:                 pulumi.String(cfg.GcpRegion),
				UniformBucketLevelAccess: pulumi.Bool(true),
			})
			if err != nil {
				slog.Error("failed to create bucket", "err", err)
				return err
			}
			ctx.Export(fmt.Sprintf("%sBucket", bucketName), bucket.ID())
		}

		for _, serviceName := range []string{"ui", "api"} {
			sa, err := serviceaccount.NewAccount(ctx, fmt.Sprintf("%s-service-sa", serviceName), &serviceaccount.AccountArgs{
				AccountId:   pulumi.Sprintf("%s-service-sa", serviceName),
				DisplayName: pulumi.Sprintf("%s Service Account for Cloud Run", serviceName),
			})
			if err != nil {
				return err
			}

			if _, err = serviceaccount.NewIAMMember(ctx, fmt.Sprintf("admin-%s-sa", serviceName), &serviceaccount.IAMMemberArgs{
				Member:           pulumi.String(fmt.Sprintf("user:%s", cfg.Email)),
				Role:             pulumi.String("roles/iam.serviceAccountUser"),
				ServiceAccountId: sa.Name,
			}); err != nil {
				return err
			}

			if _, err = serviceaccount.NewIAMMember(ctx, fmt.Sprintf("impersonate-%s-sa", serviceName), &serviceaccount.IAMMemberArgs{
				Member:           pulumi.String(fmt.Sprintf("user:%s", cfg.Email)),
				Role:             pulumi.String("roles/iam.serviceAccountTokenCreator"),
				ServiceAccountId: sa.Name,
			}); err != nil {
				return err
			}

			if _, err = projects.NewIAMMember(ctx, fmt.Sprintf("storage-admin-%s-sa", serviceName), &projects.IAMMemberArgs{
				Role:    pulumi.String("roles/storage.admin"),
				Member:  sa.Email.ApplyT(func(email string) string { return "serviceAccount:" + email }).(pulumi.StringOutput),
				Project: sa.Project,
			}); err != nil {
				return err
			}

			srv, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
				Name:     pulumi.String(fmt.Sprintf("%s-srv", serviceName)),
				Location: pulumi.String(cfg.GcpRegion),
				Template: &cloudrun.ServiceTemplateArgs{
					Spec: &cloudrun.ServiceTemplateSpecArgs{
						ServiceAccountName: sa.Email,
						Containers: cloudrun.ServiceTemplateSpecContainerArray{
							&cloudrun.ServiceTemplateSpecContainerArgs{
								Image: pulumi.String("us-docker.pkg.dev/cloudrun/container/hello"),
							},
						},
					},
				},
				Traffics: cloudrun.ServiceTrafficArray{
					&cloudrun.ServiceTrafficArgs{
						Percent:        pulumi.Int(100),
						LatestRevision: pulumi.Bool(true),
					},
				},
			})
			srv.ToServiceOutput()
			if err != nil {
				return err
			}
			ctx.Export(fmt.Sprintf("%s-endpoint", serviceName), srv.Statuses.Index(pulumi.Int(0)).Url())
		}

		return nil
	}

	stack, err := auto.UpsertStackInlineSource(ctx, "app", projectName, deployFunc, auto.WorkDir(workspace.WorkDir()))
	if err != nil {
		log.Fatalf("failed to create stack: %v", err)
	}
	if err := stack.SetConfig(ctx, "gcp:project", auto.ConfigValue{Value: projectId}); err != nil {
		log.Fatalf("failed to set project config: %v", err)
	}

	if err := stack.SetConfig(ctx, "gcp:region", auto.ConfigValue{Value: cfg.GcpRegion}); err != nil {
		log.Fatalf("failed to set project config: %v", err)
	}
	if _, err := stack.Refresh(ctx); err != nil {
		log.Fatalf("failed to refresh stack: %v", err)
	}

	return stack
}

func main() {
	slog.Info("deploying nomaladies stack")

	projectStack := createProjectStack(cfg.ProjectName, cfg.BillingAccountId)
	if err := projectStack.SetConfig(ctx, "gcp:project", auto.ConfigValue{
		Value:  cfg.SeedProjectId,
		Secret: false,
	}); err != nil {
		log.Fatalf("failed to set seed project id: %v", err)
	}

	stdoutStreamer := optup.ProgressStreams(os.Stdout)
	res, err := projectStack.Up(ctx, stdoutStreamer)
	if err != nil {
		log.Fatalf("failed to deploy project stack: %v", err)
	}
	slog.Info("project stack deployed", "summary", res.Summary, "outputs", res.Outputs)

	projectId, ok := res.Outputs["projectId"].Value.(string)
	if !ok {
		log.Fatalf("failed to get project id")
	}

	appStack := createAppStack(cfg.ProjectName, projectId)
	res, err = appStack.Up(ctx, stdoutStreamer)
	if err != nil {
		log.Fatalf("failed to deploy app stack: %v", err)
	}
	slog.Info("app stack deployed", "summary", res.Summary)
}
