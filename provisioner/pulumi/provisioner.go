package pulumi

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	pg "palter.io/paltergeist"
)

type Provisioner struct {
	ctx       context.Context
	workspace auto.Workspace
	// Target GCP projectId
	projectId string
	// Paltergiest GCP projectId
	paltergeistProjectId string
	// Fully qualified names of stacks to add to the provisioner
	stackFullyQualifiedNames []string
}

type Option func(*Provisioner)

func WithProjectId(projectId string) Option {
	return func(p *Provisioner) {
		p.projectId = projectId
	}
}

func WithPaltergiestProjectId(paltergeistProjectId string) Option {
	return func(p *Provisioner) {
		p.paltergeistProjectId = paltergeistProjectId
	}
}

// WithTarget adds a stack to the provisioner to add to Graph and use as Persona examples.
func WithTargetStack(stackFullyQualifiedName string) Option {
	return func(p *Provisioner) {
		p.stackFullyQualifiedNames = append(p.stackFullyQualifiedNames, stackFullyQualifiedName)
	}
}

// getProjectID returns the current GCP project ID.
func getProjectID() (string, error) {
	cmd := exec.Command("gcloud", "config", "get-value", "project")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func New(ctx context.Context, opts ...Option) (*Provisioner, error) {
	workspace := Workspace(ctx)
	p := &Provisioner{
		ctx:                      ctx,
		workspace:                workspace,
		projectId:                "",
		stackFullyQualifiedNames: make([]string, 0),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.projectId == "" {
		credentials, err := getProjectID()
		if err != nil {
			return nil, err
		}
		p.projectId = credentials
	}
	if p.paltergeistProjectId == "" {
		credentials, err := getProjectID()
		if err != nil {
			return nil, err
		}
		p.paltergeistProjectId = credentials
	}
	slog.Info("initialized provisioner", "provisioner", p)
	return p, nil
}

func (p *Provisioner) String() string {
	return fmt.Sprintf("Provisioner{projectId: %s, paltergeistProjectId: %s, stackFullyQualifiedNames: %v}", p.projectId, p.paltergeistProjectId, p.stackFullyQualifiedNames)
}

// Add trap resources in the graph to a stack in the Pulumi project.
func (p *Provisioner) Add(engagementName string, graph *pg.Graph) error {
	runFunc := func(ctx *pulumi.Context) error {
		serviceAccountEmails := make([]any, 0)
		bucketNames := make([]any, 0)
		for _, trap := range graph.Traps() {
			switch trap.Type() {
			case pg.BucketResource:
				b := trap.(*pg.Bucket)
				slog.Info("adding bucket", "bucket", b)
				bucket, err := storage.NewBucket(ctx, b.Name, &storage.BucketArgs{
					Name:                     pulumi.String(b.Name),
					Location:                 pulumi.String(b.Location),
					UniformBucketLevelAccess: pulumi.Bool(true),
					ForceDestroy:             pulumi.Bool(true),
				})
				if err != nil {
					slog.Error("failed to create bucket", "err", err)
					return err
				}
				bucketNames = append(bucketNames, bucket.Name)
			case pg.ServiceAccountResource:
				saTrap := trap.(*pg.ServiceAccount)
				slog.Info("adding service account", "serviceAccount", saTrap)
				// Create the service account.
				sa, err := serviceaccount.NewAccount(ctx, saTrap.Name, &serviceaccount.AccountArgs{
					Project:     pulumi.String(p.projectId),
					AccountId:   pulumi.String(saTrap.Name),
					DisplayName: pulumi.String(saTrap.Description),
				})
				if err != nil {
					return err
				}
				// Add the service account email to the list of emails to filter logs by.
				serviceAccountEmails = append(serviceAccountEmails, sa.Email)
				// Let the target service accounts impersonate the trap service accounts.
				targetServiceAccounts := pg.Select[*pg.ServiceAccount](graph, pg.IsTarget)
				for _, targetServiceAccount := range targetServiceAccounts {
					if _, err := serviceaccount.NewIAMMember(ctx, fmt.Sprintf("trap-sa-user-%s-%s", targetServiceAccount.Name, saTrap.Name), &serviceaccount.IAMMemberArgs{
						Member:           pulumi.String(fmt.Sprintf("serviceAccount:%s", targetServiceAccount.Email)),
						Role:             pulumi.String("roles/iam.serviceAccountUser"),
						ServiceAccountId: sa.Name,
					}); err != nil {
						return err
					}
					if _, err := serviceaccount.NewIAMMember(ctx, fmt.Sprintf("impersonate-sa-%s-%s", targetServiceAccount.Name, saTrap.Name), &serviceaccount.IAMMemberArgs{
						Member:           pulumi.String(fmt.Sprintf("serviceAccount:%s", targetServiceAccount.Email)),
						Role:             pulumi.String("roles/iam.serviceAccountTokenCreator"),
						ServiceAccountId: sa.Name,
					}); err != nil {
						return err
					}
				}
				// Let users impersonate the trap service account.
				users := pg.Select[*pg.User](graph, pg.IsTarget)
				for _, user := range users {
					if _, err := serviceaccount.NewIAMMember(ctx, fmt.Sprintf("trap-sa-user-%s-%s", user.Email, saTrap.Name), &serviceaccount.IAMMemberArgs{
						Member:           pulumi.String(fmt.Sprintf("user:%s", user.Email)),
						Role:             pulumi.String("roles/iam.serviceAccountUser"),
						ServiceAccountId: sa.Name,
					}); err != nil {
						return err
					}

					if _, err := serviceaccount.NewIAMMember(ctx, fmt.Sprintf("impersonate-sa-%s-%s", user.Email, saTrap.Name), &serviceaccount.IAMMemberArgs{
						Member:           pulumi.String(fmt.Sprintf("user:%s", user.Email)),
						Role:             pulumi.String("roles/iam.serviceAccountTokenCreator"),
						ServiceAccountId: sa.Name,
					}); err != nil {
						return err
					}
				}
			default:
				slog.Warn("unsupported trap type", "type", trap.Type())
			}
		}

		// Filter interactions with traps by the service account emails.
		pulumi.All(serviceAccountEmails...).ApplyT(func(inputs []any) error {
			filterBuilder := func() string {
				emails := make([]string, 0)
				for _, input := range inputs {
					emails = append(emails, input.(string))
				}
				regex := strings.Join(emails, "|")
				// Filter logs by the service account emails: `protoPayload.authenticationInfo.principalEmail =~ "(%s|%s)"`
				return fmt.Sprintf("protoPayload.authenticationInfo.principalEmail =~ \"(%s)\" OR resource.labels.email_id =~ \"(%s)\"", regex, regex)
			}
			slog.Info("adding sink", "filter", filterBuilder())
			return createLoggingSink(ctx, p, "paltergiest-trap-sink", filterBuilder)
		})

		// Filter interactions with traps by the bucket names.
		pulumi.All(bucketNames...).ApplyT(func(inputs []any) error {
			filterBuilder := func() string {
				buckets := make([]string, 0)
				for _, input := range inputs {
					buckets = append(buckets, input.(string))
				}
				regex := strings.Join(buckets, "|")
				// Filter logs by the bucket names: `protoPayload.resourceName =~ "(%s|%s)"`
				return fmt.Sprintf("protoPayload.resourceName =~ \"(%s)\"", regex)
			}
			slog.Info("adding bucket sink", "filter", filterBuilder())
			return createLoggingSink(ctx, p, "paltergiest-trap-bucket-sink", filterBuilder)
		})

		return nil
	}
	stack, err := auto.UpsertStackInlineSource(p.ctx, "paltergiest", engagementName, runFunc, auto.WorkDir(p.workspace.WorkDir()))
	if err != nil {
		return err
	}
	if err := stack.SetConfig(p.ctx, "gcp:project", auto.ConfigValue{Value: p.projectId}); err != nil {
		return err
	}
	if _, err := stack.Refresh(p.ctx); err != nil {
		return err
	}
	slog.Info("added stack", "stack", stack.Name())
	// Wire up our update to stream progress to stdout
	stdoutStreamer := optup.ProgressStreams(os.Stdout)
	// Creates or updates the stack in the environment.
	if res, err := stack.Up(p.ctx, stdoutStreamer); err != nil {
		slog.Error("Failed to update stack", "err", err)
		return err
	} else {
		slog.Info("deployed", "engagement", engagementName)
		slog.Debug("deploy summary", "summary", res.Summary)
	}
	return nil
}

func (p *Provisioner) Destroy(engagementName string) error {
	stack, err := auto.UpsertStackInlineSource(p.ctx, "paltergiest", engagementName, func(ctx *pulumi.Context) error {
		return nil
	}, auto.WorkDir(p.workspace.WorkDir()))
	if err != nil {
		return err
	}
	if err := stack.SetConfig(p.ctx, "gcp:project", auto.ConfigValue{Value: p.projectId}); err != nil {
		return err
	}
	if _, err := stack.Refresh(p.ctx); err != nil {
		return err
	}
	if res, err := stack.Destroy(p.ctx); err != nil {
		return err
	} else {
		slog.Info("destroyed", "engagement", engagementName)
		slog.Debug("destroy summary", "summary", res.Summary)
	}
	return nil
}
