package pulumi

import (
	"context"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"log"
)

var workspace auto.Workspace

// Workspace creates a new local Pulumi workspace and installs the GCP plugin.
func Workspace(ctx context.Context) auto.Workspace {
	if workspace != nil {
		// Already initialized, so return workspace.
		return workspace
	}
	var err error
	// Set the local workspace to use current working directory.
	workspace, err = auto.NewLocalWorkspace(ctx, auto.WorkDir("./"))
	if err != nil {
		log.Fatalf("failed to create workspace: %v", err)
	}
	// Install latest GPC plugin version as of Feb 2025.
	if err = workspace.InstallPlugin(ctx, "gcp", "v8.20.0"); err != nil {
		log.Fatalf("failed to install gcp plugin: %v", err)
	}
	return workspace
}
