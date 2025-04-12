# Nomaladies Health

This is an example scenario to demonstrate the usage of `paltergeist`.

## Scenario

> Healthcare technology company that runs cloud-native telehealth applications for symptom monitoring and 
> preventative intervention. We store and correlate patient, healthcare, and insurance provider data to reduce 
> costs of chronic disease management.

## GCP Infrastructure

+ Cloud Run frontend application.
+ Cloud SQL for patient information— PHI and HIPPA.
+ GCS for wearable sensor blob and analytics warehouse.
+ IAM principals for engineering and business analyst users alongside service accounts.

## Usage

```bash
go run main.go
```