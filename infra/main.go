package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	projectName = "paltergeist"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		billingAccountId := conf.Require("billing-account-id")
		project, err := organizations.NewProject(ctx, projectName, &organizations.ProjectArgs{
			Name:           pulumi.String(projectName),
			BillingAccount: pulumi.String(billingAccountId),
			DeletionPolicy: pulumi.String("DELETE"),
		})
		if err != nil {
			return nil
		}
		for _, api := range []string{"compute.googleapis.com", "generativelanguage.googleapis.com", "aiplatform.googleapis.com"} {
			if _, err := projects.NewService(ctx, api, &projects.ServiceArgs{
				DisableOnDestroy: pulumi.Bool(true),
				Project:          project.ProjectId,
				Service:          pulumi.String(api),
			}); err != nil {
				return err
			}
		}
		ctx.Export("projectId", project.ProjectId.ToStringOutput())
		return nil
	})
}
