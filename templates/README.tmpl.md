# {{.Title}}

This repository holds several references to example workflows and demonstrates how to use the Google GitHub Actions for common scenarios. Each action should be represented as a sub-folder under the `workflows` folder in this repository, e.g. the `workflows/auth` folder will hold examples for the `google-github-actions/auth` action.

**NOTE: This is currently a work in progress**

## Available Examples

{{range .Actions}}### [{{.Name}}]({{.ReadMePath}})

| Name                                                         | Starter                   | Description      |
| ------------------------------------------------------------ | ------------------------- | ---------------- |
{{range .Workflows}}|[{{.RelativeName}}]({{.WorkflowPath}}) | {{ if .Starter}}✅{{end}} | {{.Description}} |
{{end}}
{{end}}
