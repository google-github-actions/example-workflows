// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

const (
	readmeTitle              = "Google GitHub Actions - Example Workflows"
	propertiesDirName string = "properties"
)

var (
	starterPtr = flag.Bool("starter", false, "starter workflow")
	typePtr    = flag.String("type", "deployments", "starter workflow type")

	propertiesTemplPath string = path.Join("templates", "workflow.properties.tmpl.json")
	rootWorkflowPath    string = path.Join("workflows")
	workflowConfigPath  string = path.Join("workflow.config.json")
	readmeTmplatePath   string = path.Join("templates", "README.tmpl.md")
	readmeOutputPath    string = path.Join(defaultEnv("OUTPUT_PATH", "README.md"))
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	flag.Parse()

	if err := realMain(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context) error {
	args := flag.Args()
	if len(args) <= 0 {
		return fmt.Errorf("expected command workflow or readme, got none")
	}

	command := args[0]

	if strings.EqualFold(command, "workflow") {
		return generateWorkflow(ctx, args)
	}

	if strings.EqualFold(command, "readme") {
		return generateReadme(ctx)
	}

	return fmt.Errorf("invalid command: %s", command)
}

// generateWorkflow handles the creation of new workflow files
func generateWorkflow(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, got %d: %q", len(args), args)
	}

	var wc workflowConfig
	if err := loadJSONFromFile(&wc, workflowConfigPath); err != nil {
		return fmt.Errorf("failed to load workflow config: %w", err)
	}

	workflowArg := args[1]
	workflowID := path.Base(workflowArg)
	workflowDir := path.Join(rootWorkflowPath, path.Dir(workflowArg))
	workflowFilePath := path.Join(workflowDir, fmt.Sprintf("%s.yml", workflowID))
	workflowDirParts := strings.Split(workflowDir, "/")

	// This should be at least workflows/action-name, but can be longer
	if len(workflowDirParts) < 2 {
		return fmt.Errorf("invalid workflow path %s, path should have at least 2 folders, e.g. action-name/workflow-name", workflowDir)
	}

	actionName := workflowDirParts[1]
	actionPath := path.Join(workflowDirParts[:2]...)
	actionReadMePath := path.Join(actionPath, "README.md")

	if _, ok := wc[workflowID]; ok {
		return fmt.Errorf("workflow exists in %s, please use existing workflow or use a different name", workflowConfigPath)
	}

	if _, err := os.Stat(workflowFilePath); err == nil {
		return fmt.Errorf("workflow file %s already exists", workflowFilePath)
	}

	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	_, err := os.Stat(actionReadMePath)
	if os.IsNotExist(err) {
		actionReadMeContents := fmt.Sprintf("# %s examples", actionName)
		if err := os.WriteFile(actionReadMePath, []byte(actionReadMeContents), 0644); err != nil {
			return fmt.Errorf("failed writing content to action README file %s: %w", actionReadMePath, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to validate %s exists: %w", actionReadMePath, err)
	}

	fileContents := "# TODO: Add meaningful workflow content here."
	if err := os.WriteFile(workflowFilePath, []byte(fileContents), 0644); err != nil {
		return fmt.Errorf("writing content to workflow file: %w", err)
	}

	propertiesFilePath := path.Join(propertiesDirName, fmt.Sprintf("%s.properties.json", workflowID))
	propertiesConfig := &propertiesTemplateConfig{
		WorkflowID: workflowID,
	}

	if err := renderTemplate(propertiesTemplPath, propertiesFilePath, propertiesConfig); err != nil {
		return fmt.Errorf("failed to render properties template: %w", err)
	}

	wc[workflowID] = workflow{
		Starter:        *starterPtr,
		Type:           *typePtr,
		WorkflowPath:   workflowFilePath,
		PropertiesPath: propertiesFilePath,
	}

	newConfigBytes, err := json.MarshalIndent(wc, "", "  ")
	if err != nil {
		return fmt.Errorf("fail to marshal new workflow config: %w", err)
	}

	if err := os.WriteFile(workflowConfigPath, newConfigBytes, 0644); err != nil {
		return fmt.Errorf("failed to write update workflow config: %w", err)
	}

	return nil
}

// generateWorkflow handles the creation of the main readme and individual action readmes
func generateReadme(ctx context.Context) error {
	var wfConfig workflowConfig
	if err := loadJSONFromFile(&wfConfig, workflowConfigPath); err != nil {
		return fmt.Errorf("failed to load workflow config %s: %w", workflowConfigPath, err)
	}

	hasInvalidConfigs := false
	sortedWorkflowsIDs := getSortedWorkflowIDs(wfConfig)
	readmeActions := map[string]readmeAction{}

	for _, workflowID := range sortedWorkflowsIDs {
		workflow := wfConfig[workflowID]
		workflowPathParts := strings.Split(workflow.WorkflowPath, "/")

		// This should be at least workflows/action-name/workflow-name.yml, but can be longer
		if len(workflowPathParts) < 3 {
			return fmt.Errorf("invalid workflow path %s, should be at least workflows/action-name/workflow-name.yml", workflow.WorkflowPath)
		}

		actionName := workflowPathParts[1]
		actionPath := path.Join(workflowPathParts[:2]...)
		actionReadMePath := path.Join(actionPath, "README.md")
		workflowSubPath := path.Join(workflowPathParts[2:]...)
		workflowRelativeName := strings.TrimSuffix(workflowSubPath, filepath.Ext(workflowSubPath))

		if err := validateGenerateReadme(workflow, readmeAction{ReadMePath: actionReadMePath}); err != nil {
			fmt.Println(fmt.Errorf("validation failed for generate readme workflow %s: %w", workflowID, err))
			hasInvalidConfigs = true
			continue
		}

		var properties propertiesConfig
		if err := loadJSONFromFile(&properties, workflow.PropertiesPath); err != nil {
			return fmt.Errorf("failed to load properties file %s: %w", workflow.PropertiesPath, err)
		}

		actionData, hasKey := readmeActions[actionName]
		if !hasKey {
			emptyWorkflows := make([]readmeWorkflow, 0)
			actionData = readmeAction{
				Name:       actionName,
				Path:       actionPath,
				ReadMePath: actionReadMePath,
				Workflows:  emptyWorkflows,
			}
		}

		actionData.Workflows = append(actionData.Workflows, readmeWorkflow{
			Name:           properties.Name,
			RelativeName:   workflowRelativeName,
			Description:    properties.Description,
			Starter:        workflow.Starter,
			WorkflowPath:   workflow.WorkflowPath,
			PropertiesPath: workflow.PropertiesPath,
		})

		readmeActions[actionData.Name] = actionData
	}

	if hasInvalidConfigs {
		return fmt.Errorf("failed to process invalid configs")
	}

	sortedActions := getSortedActionNames(readmeActions)

	readmeTemplateConfigs := readmeTemplateConfig{
		Title:   readmeTitle,
		Actions: sortedActions,
	}

	if err := renderTemplate(readmeTmplatePath, readmeOutputPath, readmeTemplateConfigs); err != nil {
		return fmt.Errorf("failed to render readme template: %w", err)
	}

	return nil
}

// validateGenerateReadme handles validations for generating readmes
func validateGenerateReadme(w workflow, a readmeAction) error {
	if _, err := os.Stat(w.WorkflowPath); err != nil {
		return fmt.Errorf("failed to validate %s exists: %w", w.WorkflowPath, err)
	}

	if _, err := os.Stat(w.PropertiesPath); err != nil {
		return fmt.Errorf("failed to validate %s exists: %w", w.PropertiesPath, err)
	}

	if _, err := os.Stat(a.ReadMePath); err != nil {
		return fmt.Errorf("failed to validate %s exists: %w", a.ReadMePath, err)
	}

	return nil
}

// renderTemplate renders a go template
func renderTemplate(templatePath string, outputPath string, templateConfig interface{}) error {
	template, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	err = template.Execute(file, templateConfig)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// getSortedWorkflowIDs sorts workflowConfig by workflowID
func getSortedWorkflowIDs(workflowConfig workflowConfig) []string {
	workflowIDs := make([]string, 0, len(workflowConfig))
	for id := range workflowConfig {
		workflowIDs = append(workflowIDs, id)
	}
	sort.Strings(workflowIDs)

	return workflowIDs
}

// getSortedActionNames sorts a list of readmeActions by name
func getSortedActionNames(actions map[string]readmeAction) []readmeAction {
	actionNames := make([]string, 0, len(actions))
	for name := range actions {
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	readmeActionData := make([]readmeAction, 0, len(actions))
	for _, actionName := range actionNames {
		readmeActionData = append(readmeActionData, actions[actionName])
	}

	return readmeActionData
}

// loadJSONFromFile loads unmarshals json from a file path
func loadJSONFromFile(config interface{}, path string) error {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	return nil
}

// defaultEnv sets a default value for a missing environment variable
func defaultEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// propertiesTemplateConfig is the go template config used for the workflow properties template
type propertiesTemplateConfig struct {
	WorkflowID string
}

// propertiesConfig are the object properties for the *.properties.json files
type propertiesConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Creator     string   `json:"creator"`
	IconName    string   `json:"iconName"`
	Categories  []string `json:"categories"`
}

// workflow is the object properties for each workflow
type workflow struct {
	Starter        bool   `json:"starter"`
	Type           string `json:"type"`
	WorkflowPath   string `json:"workflowPath"`
	PropertiesPath string `json:"propertiesPath"`
}

// workflowConfig is the object referencing all workflow configs
type workflowConfig map[string]workflow

// readmeAction is the action template config used for the index README template
type readmeAction struct {
	Name       string
	Path       string
	ReadMePath string
	Workflows  []readmeWorkflow
}

// readmeWorkflow is the workflow config used for the index README template
type readmeWorkflow struct {
	Name           string
	RelativeName   string
	Description    string
	Starter        bool
	WorkflowPath   string
	PropertiesPath string
}

// readmeTemplateConfig is the template config used for the index README template
type readmeTemplateConfig struct {
	Title   string
	Actions []readmeAction
}
