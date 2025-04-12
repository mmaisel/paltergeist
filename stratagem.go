package paltergeist

import (
	"context"
	"log/slog"

	"palter.io/paltergeist/generator"
)

const (
	defaultTrapCount = 3
)

// Stratagem plans a collection of related traps for specific use case.
type Stratagem func(ctx context.Context, g *Graph) error

func FollowTheYellowBrickRoad(ctx context.Context, g *Graph) error {
	slog.Info("planning Follow The Yellow Brick Road stratagem")
	serviceAccounts := Select[*ServiceAccount](g, IsTarget)
	// generate service account traps
	traps, err := generator.GenerateTraps(ctx, serviceAccounts, defaultTrapCount)
	if err != nil {
		return err
	}
	for _, trap := range traps {
		// add the trap to the graph, and mark it.
		slog.Info("adding service account trap", "trap", trap)
		if err := g.AddTrap(trap); err != nil {
			return err
		}
	}
	return nil
}

func CrownJewelGravityWell(ctx context.Context, g *Graph) error {
	slog.Info("planning Crown Jewel Gravity Well stratagem")
	buckets := Select[*Bucket](g, IsTarget)
	// generate bucket traps
	traps, err := generator.GenerateTraps(ctx, buckets, 1)
	if err != nil {
		return err
	}
	for _, trap := range traps {
		// TODO: constrain generation to supported regions and storage classes.
		trap.Location = "us-central1"
		trap.StorageClass = "STANDARD"
		// add the trap to the graph, and mark it.
		slog.Info("adding bucket trap", "trap", trap)
		if err := g.AddTrap(trap); err != nil {
			return err
		}
	}
	return nil
}
