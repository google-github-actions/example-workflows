# {{.Title}}

This repository holds several references to example workflows and demonstrates how to use the Google GitHub Actions for common scenarios. Each action should be represented as a sub-folder under the `workflows` folder in this repository, e.g. the `workflows/auth` folder will hold examples for the `google-github-actions/auth` action.

**NOTE: This is currently a work in progress**

## Example Workflows vs Starter Workflows

### Starter Workflows

Starter workflows should be considered _"as simple as is needed for the service"_. This is usually a common scenario with best practices so users can use it off-the-shelf with their applications. They are integrated with the GitHub user interface and are presented to users based on the types of files that exist in their repositories, see the categories property [here](https://github.com/actions/starter-workflows/blob/main/CONTRIBUTING.md).

Additionally, starter workflows are reviewed by the GitHub team and have a published [contributing guide](https://github.com/actions/starter-workflows/blob/main/CONTRIBUTING.md).

### Example Workflows

Can be used to showcase any functionality for a given action. This may include examples for documentation or a blog article and may have highly specific use cases that don't make sense to surface as starter workflows.

#### Example

A good starter workflow for Cloud Run is to build a Docker container for your application, upload it Google Container Registry and then deploy the container Cloud Run. This is a common starting place and has everything needed to start using Cloud Run.

A bad starter workflow for Cloud Run may have user specific logic or custom scripts and implementation steps. This could be good for a specific use case or documentation/blog article, but isn't simple or generic enough for all users to start with.

## Available Examples

{{range .Actions}}### [{{.Name}}]({{.ReadMePath}})

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
{{range .Workflows}}|[{{.RelativeName}}]({{.WorkflowPath}}) | {{ if .Starter}}✅{{end}} | {{.Description}} |
{{end}}
{{end}}
