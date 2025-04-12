package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"palter.io/paltergeist/generator"

	"github.com/urfave/cli/v3"
	palter "palter.io/paltergeist"
	"palter.io/paltergeist/engagement"
	"palter.io/paltergeist/provisioner/pulumi"
)

var (
	engage *engagement.Engagement
)

type config struct {
	// Target stack to sample resources from.
	TargetStack string `env:"TARGET_STACK,required"`
	// Target project id to deploy resources.
	TargetProjectId string `env:"TARGET_PROJECT_ID,required"`
	// Paltergeist project id to deploy trap monitoring.
	PaltergeistProjectId string `env:"PALTERGEIST_PROJECT_ID,required"`
	// Name of the engagement to deploy use in Pulumi project.
	EngagementName string `env:"ENGAGEMENT_NAME,required"`
}

func main() {
	cmd := &cli.Command{
		Name:  "paltergeist",
		Usage: "A tool for generating and deploying cloud-native traps that lure and detect attackers.",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			slog.Info("starting paltergeist...")
			err := godotenv.Load()
			if err != nil {
				panic(err)
			}
			var cfg config
			if err := env.Parse(&cfg); err != nil {
				panic(err)
			}
			// Construct provisioner for the engagement, pass in Target Stacks to sample resources.
			p, err := pulumi.New(ctx,
				pulumi.WithTargetStack(cfg.TargetStack),
				pulumi.WithProjectId(cfg.TargetProjectId),
				pulumi.WithPaltergiestProjectId(cfg.PaltergeistProjectId),
			)
			if err != nil {
				panic(err)
			}
			// Create engagement with name, provisioner, and stratagems.
			engage = engagement.New(cfg.EngagementName, p, palter.FollowTheYellowBrickRoad, palter.CrownJewelGravityWell)
			generator.Initialize(ctx, cfg.PaltergeistProjectId)
			return ctx, nil
		},
		Commands: []*cli.Command{
			{
				Name:  "deploy",
				Usage: "Deploy the paltergeist resources to the target stacks",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return deploy(ctx)
				},
			},
			{
				Name:  "destroy",
				Usage: "Destroy the paltergeist resources from the target stacks",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return destroy(ctx)
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// deploy plans and deploys engagement.
func deploy(ctx context.Context) error {
	// Plan engagement, execute all stratagems in sequence, generating traps, and adding them to the graph.
	err := engage.Plan(ctx)
	if err != nil {
		slog.Error("failed to plan engagement", "error", err)
		panic(err)
	}
	// Up engagement, execute all stratagems in sequence, generating traps, and adding them to the graph.
	err = engage.Deploy()
	if err != nil {
		slog.Error("failed to up engagement", "error", err)
		panic(err)
	}
	return nil
}

// destroy destroys the engagement.
func destroy(ctx context.Context) error {
	err := engage.Destroy()
	if err != nil {
		slog.Error("failed to destroy engagement", "error", err)
		panic(err)
	}
	return nil
}
