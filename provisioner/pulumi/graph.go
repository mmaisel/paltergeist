package pulumi

import (
	"encoding/json"
	"log"
	"log/slog"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	pg "palter.io/paltergeist"
)

// export Pulumi's Stack Deployment state to use for graph transformation.
func (p *Provisioner) export(fqStackName string) apitype.DeploymentV3 {
	stack, err := auto.SelectStack(p.ctx, fqStackName, p.workspace)
	if err != nil {
		log.Fatalf("failed to select stack: %v", err)
	}
	// Export the Stack to Deployment state.
	deployment, err := stack.Export(p.ctx)
	if err != nil {
		log.Fatalf("failed to export app stack state: %v", err)
	}
	if deployment.Version != 3 {
		log.Fatalf("unsupported deployment version: %d", deployment.Version)
	}
	var state apitype.DeploymentV3
	if err = json.Unmarshal(deployment.Deployment, &state); err != nil {
		log.Fatalf("failed to unmarshal deployment: %v", err)
	}
	return state
}

// ConstructGraph builds a Resource Graph based on the target persona stacks.
func (p *Provisioner) ConstructGraph() *pg.Graph {
	g := pg.New()
	for _, fqStackName := range p.stackFullyQualifiedNames {
		slog.Info("building graph from stack", "stack", fqStackName)
		state := p.export(fqStackName)
		for _, rsc := range state.Resources {
			switch rsc.Type.String() {
			case "gcp:storage/bucket:Bucket":
				bucket := pg.Bucket{
					Name:         rsc.Outputs["name"].(string),
					Location:     rsc.Outputs["location"].(string),
					StorageClass: rsc.Outputs["storageClass"].(string),
				}
				if err := g.Add(&bucket); err != nil {
					panic(err)
				}
			case "gcp:serviceaccount/iAMMember:IAMMember":
				slog.Info("iam member", "outputs", rsc.Outputs)
				if strings.HasPrefix(rsc.Outputs["member"].(string), "user:") {
					user := pg.User{
						Email: strings.TrimPrefix(rsc.Outputs["member"].(string), "user:"),
					}
					if err := g.Add(&user); err != nil {
						slog.Info(err.Error())
						continue
					}
				}
			case "gcp:serviceaccount/account:Account":
				member := rsc.Outputs["member"].(string)
				if strings.HasPrefix(member, "serviceAccount:") {
					// Add ServiceAccount Resource
					sa := pg.ServiceAccount{
						Id:          rsc.Outputs["id"].(string),
						Name:        rsc.Outputs["accountId"].(string),
						Description: rsc.Outputs["displayName"].(string),
						Email:       rsc.Outputs["email"].(string),
					}
					if err := g.Add(&sa); err != nil {
						panic(err)
					}
				} else {
					slog.Warn("unsupported resource", "type", rsc.Type, "member", member)
				}
			}
			if strings.HasPrefix(rsc.Type.String(), "gcp:") {
				slog.Info("resource", "type", rsc.Type, "urn", rsc.URN)
			}
		}

	}
	slog.Info("constructed graph", "size", len(g.Resources()))
	return g
}
