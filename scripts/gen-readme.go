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
	"errors"
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

var (
	workflowConfigPath string = path.Clean(path.Join("workflow.config.json"))
	readmeTmplatePath  string = path.Clean(path.Join("templates", "README.tmpl.md"))
	readmeOutputPath   string = path.Clean(defaultEnv("OUTPUT_PATH", path.Join(".", "README.md")))
)

// PropertiesConfig are the object proeprties for the *.properties.json files
type PropertiesConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Creator     string   `json:"creator"`
	IconName    string   `json:"iconName"`
	Categories  []string `json:"categories"`
}

// ReadmeActionConfig is the action template config used for the index README template
type ReadmeAction struct {
	Name       string
	Path       string
	ReadMePath string
	Workflows  []ReadmeWorkflow
}

// ReadmeWorkflowConfig is the workflow config used for the index README template
type ReadmeWorkflow struct {
	Name           string
	RelativeName   string
	Description    string
	Starter        bool
	WorkflowPath   string
	PropertiesPath string
}

// ReadmeTemplateConfig is the template config used for the index README template
type ReadmeTemplateConfig struct {
	Title   string
	Actions []ReadmeAction
}

// Workflow is the object properties for each workflow
type Workflow struct {
	Starter        bool   `json:"starter"`
	Type           string `json:"type"`
	WorkflowPath   string `json:"workflowPath"`
	PropertiesPath string `json:"propertiesPath"`
}

// WorkflowConfig is the object referencing all workflow configs
type WorkflowConfig map[string]Workflow

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := realMain(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context) error {
	configBytes, err := os.ReadFile(workflowConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read workflow config file: %w", err)
	}

	var workflowConfig WorkflowConfig
	if err := json.Unmarshal(configBytes, &workflowConfig); err != nil {
		return fmt.Errorf("failed to unmarshal workflow config: %w", err)
	}

	hasInvalidConfigs := false

	sortedWorkflowsIDs := getSortedWorkflowIDs(workflowConfig)

	readmeActions := map[string]ReadmeAction{}

	for _, workflowID := range sortedWorkflowsIDs {
		workflow := workflowConfig[workflowID]

		if _, err := os.Stat(workflow.WorkflowPath); os.IsNotExist(err) {
			hasInvalidConfigs = true
			fmt.Println(fmt.Sprintf("workflow file does not exist for workflow %s: path - %s", workflowID, workflow.WorkflowPath))
		}

		propertiesBytes, err := os.ReadFile(workflow.PropertiesPath)
		if err != nil {
			if os.IsNotExist(err) {
				hasInvalidConfigs = true
				fmt.Println(fmt.Sprintf("properties file does not exist for workflow %s: path - %s", workflowID, workflow.PropertiesPath))
			} else {
				return fmt.Errorf("failed to read properties file %s: %w", workflow.PropertiesPath, err)
			}
		}

		if hasInvalidConfigs {
			continue
		}

		var propertiesConfigs PropertiesConfig
		if err := json.Unmarshal(propertiesBytes, &propertiesConfigs); err != nil {
			return fmt.Errorf("failed to unmarshal properties file %s: %w", workflow.PropertiesPath, err)
		}

		parts := strings.Split(workflow.WorkflowPath, "/")

		if len(parts) < 2 {
			return errors.New("invalid workflow path %s, path should have at least 3 folders")
		}

		actionName := parts[1]
		actionPath := path.Join(parts[:2]...)
		actionReadMePath := path.Join(actionPath, "README.md")

		if _, err := os.Stat(actionReadMePath); err != nil {
			if os.IsNotExist(err) {
				actionReadMeContents := fmt.Sprintf("# %s examples", actionName)
				if err := os.WriteFile(actionReadMePath, []byte(actionReadMeContents), 0644); err != nil {
					return fmt.Errorf("failed writing content to action README file %s: %w", actionReadMePath, err)
				}
			} else {
				return fmt.Errorf("failed to validate %s README file exists: %w", actionReadMePath, err)
			}
		}

		workflowSubPath := path.Join(parts[2:]...)
		workflowRelativeName := strings.TrimSuffix(workflowSubPath, filepath.Ext(workflowSubPath))

		actionData, hasKey := readmeActions[actionName]
		if !hasKey {
			emptyWorkflows := make([]ReadmeWorkflow, 0)
			actionData = ReadmeAction{
				Name:       actionName,
				Path:       actionPath,
				ReadMePath: actionReadMePath,
				Workflows:  emptyWorkflows,
			}
		}

		actionData.Workflows = append(actionData.Workflows, ReadmeWorkflow{
			Name:           propertiesConfigs.Name,
			RelativeName:   workflowRelativeName,
			Description:    propertiesConfigs.Description,
			Starter:        workflow.Starter,
			WorkflowPath:   workflow.WorkflowPath,
			PropertiesPath: workflow.PropertiesPath,
		})

		readmeActions[actionName] = actionData
	}

	sortedActions := getSortedActionNames(readmeActions)

	readmeTemplateConfigs := ReadmeTemplateConfig{
		Title:   "Google GitHub Actions - Example Workflows",
		Actions: sortedActions,
	}

	if hasInvalidConfigs {
		return fmt.Errorf("failed to process invalid configs")
	}

	if err := renderReadmeTemplate(readmeTemplateConfigs); err != nil {
		return fmt.Errorf("failed to render readme template: %w", err)
	}

	return nil
}

func renderReadmeTemplate(templateConfig ReadmeTemplateConfig) error {
	readmeTemplate, err := template.ParseFiles(readmeTmplatePath)
	if err != nil {
		return fmt.Errorf("failed to create properties template: %w", err)
	}

	readmeFile, err := os.Create(readmeOutputPath)
	if err != nil {
		return fmt.Errorf("failed to create properties file: %w", err)
	}
	defer readmeFile.Close()

	err = readmeTemplate.Execute(readmeFile, templateConfig)
	if err != nil {
		return fmt.Errorf("failed to execute properties file template: %w", err)
	}

	return nil
}

func getSortedWorkflowIDs(workflowConfig WorkflowConfig) []string {
	workflowIDs := make([]string, 0, len(workflowConfig))
	for id := range workflowConfig {
		workflowIDs = append(workflowIDs, id)
	}
	sort.Strings(workflowIDs)

	return workflowIDs
}

func getSortedActionNames(actions map[string]ReadmeAction) []ReadmeAction {
	actionNames := make([]string, 0, len(actions))
	for name := range actions {
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	readmeActionData := make([]ReadmeAction, 0, len(actions))
	for _, actionName := range actionNames {
		readmeActionData = append(readmeActionData, actions[actionName])
	}

	return readmeActionData
}

func defaultEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
