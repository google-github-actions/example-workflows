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
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
)

var (
	workflowConfigPath string = path.Clean(path.Join("workflow.config.json"))
	outputPath         string = path.Clean(defaultEnv("OUTPUT_PATH", path.Join("..", "starter-workflows")))
	outputPropsDirName string = "properties"
	outputFilePrefix   string = "google"
)

// Workflow is the object properties for each workflow
type Workflow struct {
	Starter        bool   `json:"starter"`
	Type           string `json:"type"`
	WorkflowPath   string `json:"workflowPath"`
	PropertiesPath string `json:"propertiesPath"`
}

// WorkflowConfig is the object referencing all workflow configs
type WorkflowConfig map[string]Workflow

// FileCopyConfig is the source and destination file path for the files to copy
type FileCopyConfig struct {
	Source string
	Dest   string
}

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
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var workflowConfig WorkflowConfig
	if err := json.Unmarshal(configBytes, &workflowConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	isInvalid := false

	filesToCopy := make([]FileCopyConfig, 0)
	for workflowID, workflow := range workflowConfig {
		// skip non-starter workflows
		if !workflow.Starter {
			continue
		}

		if _, err := os.Stat(workflow.WorkflowPath); os.IsNotExist(err) {
			isInvalid = true
			fmt.Println(fmt.Sprintf("workflow file does not exist for workflow %s: path - %s", workflowID, workflow.WorkflowPath))
		}

		if _, err := os.Stat(workflow.PropertiesPath); os.IsNotExist(err) {
			isInvalid = true
			fmt.Println(fmt.Sprintf("properties file does not exist for workflow %s: path - %s", workflowID, workflow.PropertiesPath))
		}

		// add workflow yaml to copy list
		workflowFilename := path.Base(workflow.WorkflowPath)
		workflowDestFilename := fmt.Sprintf("%s-%s", outputFilePrefix, workflowFilename)
		filesToCopy = append(filesToCopy, FileCopyConfig{
			Source: workflow.WorkflowPath,
			Dest:   path.Join(outputPath, workflow.Type, workflowDestFilename),
		})

		// add properties file to copy list
		propertiesFilename := path.Base(workflow.PropertiesPath)
		propertiesDestFilename := fmt.Sprintf("%s-%s", outputFilePrefix, propertiesFilename)
		filesToCopy = append(filesToCopy, FileCopyConfig{
			Source: workflow.PropertiesPath,
			Dest:   path.Join(outputPath, workflow.Type, outputPropsDirName, propertiesDestFilename),
		})
	}

	// handle invalid config messaging and fail
	if isInvalid {
		return fmt.Errorf("failed to process invalid configs")
	}

	// copy all files to destination
	for _, file := range filesToCopy {
		// remove any existing destination files
		os.Remove(file.Dest)
		if err := os.Link(file.Source, file.Dest); err != nil {
			return fmt.Errorf("failed to copy files: %w", err)
		}
		fmt.Println(fmt.Sprintf("successfully copied %s -> %s", file.Source, file.Dest))
	}

	return nil
}

func defaultEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
