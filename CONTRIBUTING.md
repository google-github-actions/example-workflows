# Google GitHub Actions - Example Workflows

## Adding Workflows

New workflows should be bootstrapped with the provided go script: `go run scripts/new-workflow.go action-name/workflow-name`. This will generate the following items:

- A new directory if it does not exist
  - `example-workflows/workflows/action-name`
- A blank `README.md` file for the action folder if it does not exist
  - `example-workflows/workflows/action-name/workflow-name/README.md`
- A blank workflow file
  - `example-workflows/workflows/action-name/workflow-name/workflow-name.yml`
- A properties file for workflow metadata
  - `example-workflows/properties/workflow-name.properties.json`
- An entry in the main `workflow.config.json` file

### Prerequisites

- Go verison 1.17+

### Usage

#### Example Workflows

```bash
# Basic example workflow
go run scripts/new-workflow.go auth/auth-simple

# Folder Structure
/example-workflows
  /workflows
    /auth
      auth-simple.yml
```

#### Starter Workflows

```bash
# Starter workflow, default type (deployments)
go run scripts/new-workflow.go --starter deploy-cloudrun/cloudrun-docker

# Starter workflow, with type
go run scripts/new-workflow.go --starter --type="automation" deploy-cloudrun/cloudrun-automation

# Folder Structure
/example-workflows
  /properties
    cloudrun-docker.properties.json
    cloudrun-automation.properties.json
  /workflows
    /deploy-cloudrun
      cloudrun-docker.yml
      cloudrun-automation.yml
```

##### Valid Starter Types:

- automation
- ci
- code-scanning
- deployments     (default)

## Gnerate main `README.md`

The main `README.md` file holds references to all the action folders and the workflows they contain. Run the following command to generate an updated `README.md` file based on the `templates/README.tmpl.md` file:

```bash
go run scripts/gen-readme.go
```

## Pull Request to GitHub Starter Workflows

Updates to starter workflows should be merged into the GitHub Actions `actions/starter-workflows` repository. This can be done automatically by triggering the `Pull Request to GitHub` action or manually by following the steps below.

**NOTE:** The GitHub Action is still a work in progress

### Manual Process

**NOTE:** This process assumes the `actions/starter-workflows` and `google-github-actions/example-workflows` repositories are siblings.

```bash
/some-directory
  /example-workflows
    [...]
  /starter-workflows
    [...]
```

Steps:

1. Fork the `actions/starter-workflows` respository from GitHub
2. `cd` into `starter-workflows`
3. Create a new branch: `git checkout -b <BRANCH_NAME>`
4. `cd` into `example-workflows`
5. Run the go script `go run scripts/release.go` to update the required files in the `actions/starter-workflows` repository
6. Commit and push your changes to the `actions/starter-workflows` repository
7. Create a Pull Request on the `actions/starter-workflows` respository
