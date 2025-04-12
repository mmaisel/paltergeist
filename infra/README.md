# Paltergeist Detection Infrastructure

This directory contains the Pulumi infrastructure code for deploying the Paltergeist GCP project containing cloud logging for trap iteractions and Gemini for trap generation.

## Configuration

The following configuration values are required in `Pulumi.pg.yaml`:

- `gcp:project` - The GCP project ID where resources will be deployed.
- `gcp:region` - The GCP region for resource deployment .
- `pg:billing-account-id` - The GCP billing account ID to associate with the project.

## Infrastructure Components

The main infrastructure is defined in `main.go` and consists of:

- A new GCP project named "paltergeist".
- Enabled Google Cloud APIs:
  - Compute Engine API (compute.googleapis.com)
  - Gemini (generativelanguage.googleapis.com)
  - Vertex AI (aiplatform.googleapis.com)

## Usage

Deploy the infrastructure:

```bash
pulumi up
```
