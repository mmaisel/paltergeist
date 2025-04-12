
setup:
	brew install pulumi/tap/pulumi

gcp-auth:
	gcloud auth login
	gcloud auth application-default login
