# Google GitHub Actions - Example Workflows

This repository holds several references to example workflows and demonstrates how to use the Google GitHub Actions for common scenarios. Each action should be represented as a sub-folder under the `workflows` folder in this repository, e.g. the `workflows/auth` folder will hold examples for the `google-github-actions/auth` action.

**This is not an officially supported Google product, and it is not covered by a
Google Cloud support contract. To report bugs or request features in a Google
Cloud product, please contact [Google Cloud
support](https://cloud.google.com/support).**

**NOTE: This is currently a work in progress**

## Available Examples

### [create-cloud-deploy-release](workflows/create-cloud-deploy-release/README.md)

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
|[cloud-deploy-to-cloud-run](workflows/create-cloud-deploy-release/cloud-deploy-to-cloud-run.yml) |  | Build a Docker container, publish it to Google Artifact Registry, and use Cloud Deploy to deploy to Google Cloud Run. |

### [deploy-cloudrun](workflows/deploy-cloudrun/README.md)

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
|[cloudrun-buildpacks](workflows/deploy-cloudrun/cloudrun-buildpacks.yml) | ✅ | Build a container image with Buildpacks, publish it to Google Artifact Registry, and deploy to Google Cloud Run. |
|[cloudrun-declarative](workflows/deploy-cloudrun/cloudrun-declarative.yml) |  | Build a Docker container, publish it to Google Artifact Registry, and deploy to Google Cloud Run using a declarative YAML Service specification (KRM). |
|[cloudrun-docker](workflows/deploy-cloudrun/cloudrun-docker.yml) | ✅ | Build a Docker container, publish it to Google Artifact Registry, and deploy to Google Cloud Run. |
|[cloudrun-source](workflows/deploy-cloudrun/cloudrun-source.yml) | ✅ | Deploy to Google Cloud Run directly from source. |

### [get-gke-credentials](workflows/get-gke-credentials/README.md)

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
|[gke-build-deploy](workflows/get-gke-credentials/gke-build-deploy.yml) | ✅ | Build a Docker container, publish it to Google Container Registry, and deploy to GKE. |


