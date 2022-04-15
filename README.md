# Google GitHub Actions - Example Workflows

This repository holds several references to example workflows and demonstrates how to use the Google GitHub Actions for common scenarios. Each action should be represented as a sub-folder under the `workflows` folder in this repository, e.g. the `workflows/auth` folder will hold examples for the `google-github-actions/auth` action.

**NOTE: This is currently a work in progress**

## Available Examples

### [deploy-cloudrun](workflows/deploy-cloudrun/README.md)

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
|[cloudrun-docker](workflows/deploy-cloudrun/cloudrun-docker.yml) | ✅ | Build a Docker container, publish it to Google Artifact Registry, and deploy to Google Cloud Run. |
|[cloudrun-source](workflows/deploy-cloudrun/cloudrun-source.yml) | ✅ | Deploy to Google Cloud Run directly from source. |

### [get-gke-credentials](workflows/get-gke-credentials/README.md)

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
|[gke-build-deploy](workflows/get-gke-credentials/gke-build-deploy.yml) | ✅ | Build a Docker container, publish it to Google Container Registry, and deploy to GKE. |


