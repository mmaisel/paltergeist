# Paltergeist

> Defenders and attackers think in graphs. Defenders think in game theory. Attackers win...or did they?

`paltergeist` is a cyber deception tool for generation, orchestration, and monitoring of cloud-native traps that lure and detect attackers. It's built in Go and intended for security operation and engineering teams exploring the use of cyber deception in cloud environments.

+ *Make tailored, believable traps.* Trap personas are generated with large language models and uses in-context learning with examples sampled from the environment. To improve trap believability and enticingness, generation can use LLM critics and other reasoning models to assess trap quality.
+ *Manage deception engagements.* Orchestration of deception engagements— groups related traps with similar targets. The [Pulumi Automation API](https://www.pulumi.com/automation/) is used to plan and apply infrastructure lifecycles against the cloud provider.
+ *Instrument trap interactions.* Interaction activity is observable through cloud logging. Emitted logs are filtered, enriched, and processed into alerts for security operations teams.

## Stratagems

Stratagems are templates for traps in `paltergeist`. They describe the trap use case, where they are placed in the attack path, and what vulnerabilities they masquerade as to the attacker. Coverage is limited to GCP in this work but extendable to AWS, Azure, and K8s.

+ **Follow the Yellow Brick Road.** "*Security Engineers make Paved Roads for Developers, so let's build Yellow Brick Roads for Adversaries. Shhh...they don't know the Wiz is a lie.*" Detect lateral movement using valid cloud accounts and credentials. Plant trust relationships in IAM that serve as tripwires to attack path enumeration. Create chains of abusable IAM resources that bait attackers.
+ (TODO) **Save Your Pets. Shoot Your Cattle. Watch Your Canaries.** Compute resources that are canaries, traps, or honeypots. They may have preconfigured vulnerabilities or weaknesses.  Create personas that mimic existing VMs, containers, or serverless resources.
+ **Crown Jewel Gravity Well**. "Like data gravity...but for attackers." Detect data exfiltration. Storage resources like container registries, databases, blob storage, or analytics. Populate with synthetic data that entices data exfiltration.
+  (TODO) ***Enter My Cloud Labyrinth*.** Entire projects where everything is trap. Elicit threat intelligence. Traversing and interacting with any resource in the project indicates attack behavior or is highly suspect. 

## Getting Started

### Prerequisites

- Go 1.21 or later
- GCP project with required APIs enabled
- [Pulumi CLI](https://www.pulumi.com/docs/install/) installed and configured
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) installed and configured

### Installation

```bash
go build -o paltergeist ./cmd
```

### Usage

`paltergeist` uses environment variables for configuration. Create a `.env` file in the root directory or export the variables:

```dotenv
# ---- Example Project Variables (used by cmd/example) ----
# GCP project name for the example deployment
PROJECT_NAME=""
# GCP Billing Account Id for the example project
BILLING_ACCOUNT_ID=""
# GCP Seed Project Id (used by Pulumi for bootstrapping the example project)
SEED_PROJECT_ID=""
# Email address to grant IAM permissions in the example project
EMAIL=""
# Default GCP region for deploying resources in the example project
GCP_REGION="us-central1"

# ---- Paltergeist Core Variables (used by cmd/main) ----
# Fully qualified name of the Pulumi stack to sample target resources from (e.g., organization/project/stack)
TARGET_STACK=""
# GCP Project ID where the traps will be deployed
TARGET_PROJECT_ID=""
# GCP Project ID where Paltergeist monitoring resources (e.g., logging sinks) will be deployed
PALTERGEIST_PROJECT_ID=""
# A unique name for this deception engagement
ENGAGEMENT_NAME=""
```

Authenticate with GCP:

```bash
gcloud auth application-default login
```

Deploy traps:

```bash
./paltergeist deploy
```

Destroy traps:

```bash
./paltergeist destroy
```

## Code Layout

- `cmd/`: Main application entry point and CLI command definitions.
- `engagement/`: Manages the lifecycle of a deception engagement, coordinating stratagems and provisioning.
- `generator/`: Handles the generation of trap personas using AI models.
- `provisioner/`: Interfaces with infrastructure-as-code tools (currently Pulumi) to deploy and manage resources.
- `infra/`: Pulumi code for bootstrapping the Paltergeist monitoring project.
- (root): Core data types, interfaces (`Trap`, `Kind`, `Type`), and stratagem definitions (`FollowTheYellowBrickRoad`, `CrownJewelGravityWell`).
