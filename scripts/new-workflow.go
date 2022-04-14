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
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

var (
	starterPtr = flag.Bool("starter", false, "starter workflow")
	typePtr    = flag.String("type", "deployments", "starter workflow type")

	propertiesDirName   string = "properties"
	propertiesTemplPath string = path.Clean(path.Join("templates", "workflow.properties.tmpl.json"))
	rootWorkflowPath    string = path.Clean(path.Join("workflows"))
	workflowConfigPath  string = path.Clean(path.Join("workflow.config.json"))
)

// PropertiesTemplateConfig is the go template config used for the workflow properties template
type PropertiesTemplateConfig struct {
	WorkflowID string
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

	flag.Parse()

	if err := realMain(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context) error {
	args := flag.Args()
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d: %q", len(args), args)
	}

	workflowArg := args[0]
	workflowID := path.Base(workflowArg)
	workflowDir := path.Join(rootWorkflowPath, path.Dir(workflowArg))
	workflowFilePath := path.Join(workflowDir, fmt.Sprintf("%s.yml", workflowID))
	workflowDirParts := strings.Split(workflowDir, "/")

	if len(workflowDirParts) < 2 {
		return errors.New("invalid workflow path %s, path should have at least 3 folders")
	}

	configBytes, err := os.ReadFile(workflowConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var workflowConfig WorkflowConfig
	if err := json.Unmarshal(configBytes, &workflowConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if _, ok := workflowConfig[workflowID]; ok {
		return fmt.Errorf("workflow exists in %s, please use existing workflow or use a different name", workflowConfigPath)
	}

	if _, err := os.Stat(workflowFilePath); err == nil {
		return fmt.Errorf("workflow file already exists")
	}

	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	actionName := workflowDirParts[1]
	actionPath := path.Join(workflowDirParts[:2]...)
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

	fileContents := "# TODO: Add meaningful workflow content here."
	if err := os.WriteFile(workflowFilePath, []byte(fileContents), 0644); err != nil {
		return fmt.Errorf("writing content to workflow file: %w", err)
	}

	propertiesFilePath := path.Join(propertiesDirName, fmt.Sprintf("%s.properties.json", workflowID))

	propertiesTemplate, err := template.ParseFiles(propertiesTemplPath)
	if err != nil {
		return fmt.Errorf("failed to create properties template: %w", err)
	}

	propertiesFile, err := os.Create(propertiesFilePath)
	if err != nil {
		return fmt.Errorf("failed to create properties file: %w", err)
	}
	defer propertiesFile.Close()

	err = propertiesTemplate.Execute(propertiesFile, &PropertiesTemplateConfig{
		WorkflowID: workflowID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute properties file template: %w", err)
	}

	workflowConfig[workflowID] = Workflow{
		Starter:        *starterPtr,
		Type:           *typePtr,
		WorkflowPath:   workflowFilePath,
		PropertiesPath: propertiesFilePath,
	}

	newConfigBytes, err := json.MarshalIndent(workflowConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("fail to marshal new workflow config: %w", err)
	}

	if err := os.WriteFile(workflowConfigPath, newConfigBytes, 0644); err != nil {
		return fmt.Errorf("failed to write update workflow config: %w", err)
	}

	return nil
}
