package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileWorkflowStorage implements WorkflowStorage using the file system
type FileWorkflowStorage struct {
	workflowsDir string
}

// NewFileWorkflowStorage creates a new file-based workflow storage
func NewFileWorkflowStorage(workflowsDir string) *FileWorkflowStorage {
	return &FileWorkflowStorage{
		workflowsDir: workflowsDir,
	}
}

// SaveWorkflow saves a workflow to a YAML file
func (f *FileWorkflowStorage) SaveWorkflow(ctx context.Context, workflow Workflow) error {
	filename := fmt.Sprintf("%s.yml", strings.ToLower(strings.ReplaceAll(workflow.Name, " ", "_")))
	filePath := filepath.Join(f.workflowsDir, filename)

	data, err := yaml.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}

// LoadWorkflow loads a workflow from a YAML file
func (f *FileWorkflowStorage) LoadWorkflow(ctx context.Context, name string) (*Workflow, error) {
	filename := fmt.Sprintf("%s.yml", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	filePath := filepath.Join(f.workflowsDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		// Try with .yaml extension
		filename = fmt.Sprintf("%s.yaml", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
		filePath = filepath.Join(f.workflowsDir, filename)
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read workflow file: %w", err)
		}
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}

	return &wf, nil
}

// ListWorkflows lists all workflow files in the directory
func (f *FileWorkflowStorage) ListWorkflows(ctx context.Context) ([]string, error) {
	files, err := os.ReadDir(f.workflowsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflows directory: %w", err)
	}

	var workflows []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			// Convert from snake_case to CamelCase if needed
			workflows = append(workflows, name)
		}
	}

	return workflows, nil
}

// FileStateStorage implements StateStorage using the file system
type FileStateStorage struct {
	statesDir string
}

// NewFileStateStorage creates a new file-based state storage
func NewFileStateStorage(statesDir string) *FileStateStorage {
	return &FileStateStorage{
		statesDir: statesDir,
	}
}

// SaveState saves a workflow state to a JSON file
func (f *FileStateStorage) SaveState(ctx context.Context, workflowName string, state *WorkflowState) error {
	filename := fmt.Sprintf("%s_state.json", strings.ToLower(strings.ReplaceAll(workflowName, " ", "_")))
	filePath := filepath.Join(f.statesDir, filename)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads a workflow state from a JSON file
func (f *FileStateStorage) LoadState(ctx context.Context, workflowName string, stateID string) (*WorkflowState, error) {
	filename := fmt.Sprintf("%s_state.json", strings.ToLower(strings.ReplaceAll(workflowName, " ", "_")))
	filePath := filepath.Join(f.statesDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}