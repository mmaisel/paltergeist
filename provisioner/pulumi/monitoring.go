package pulumi

import (
	"fmt"
	"log/slog"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/logging"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createLoggingSink creates a GCP logging sink and grants the necessary permissions.
func createLoggingSink(ctx *pulumi.Context, p *Provisioner, sinkName string, filterBuilder func() string) error {
	sink, err := logging.NewProjectSink(ctx, sinkName, &logging.ProjectSinkArgs{
		Project:     pulumi.String(p.projectId),
		Destination: pulumi.String(fmt.Sprintf("logging.googleapis.com/projects/%s", p.paltergeistProjectId)),
		Filter:      pulumi.String(filterBuilder()),
	})
	if err != nil {
		slog.Error("failed to create sink", "err", err, "sinkName", sinkName)
		return fmt.Errorf("failed to create sink %s: %w", sinkName, err)
	}
	// Allow the sink to write to the paltergeist project.
	if _, err := projects.NewIAMMember(ctx, sinkName+"-log-writer", &projects.IAMMemberArgs{
		Member:  sink.WriterIdentity,
		Role:    pulumi.String("roles/logging.logWriter"),
		Project: pulumi.String(p.paltergeistProjectId),
	}); err != nil {
		slog.Error("failed to add sink identity to paltergeist project", "err", err, "sinkName", sinkName)
		return fmt.Errorf("failed to add sink identity %s to paltergeist project: %w", sinkName, err)
	}
	return nil
}
