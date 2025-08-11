package main

import (
	"context"
)

// RuleEngine defines the interface for rule evaluation
type RuleEngine interface {
	Evaluate(ctx context.Context, ruleName string, data map[string]any) (bool, error)
	RegisterRule(name string, rule any) error
}

// WorkflowStorage defines the interface for workflow persistence
type WorkflowStorage interface {
	SaveWorkflow(ctx context.Context, workflow Workflow) error
	LoadWorkflow(ctx context.Context, name string) (*Workflow, error)
	ListWorkflows(ctx context.Context) ([]string, error)
}

// StateStorage defines the interface for workflow state persistence
type StateStorage interface {
	SaveState(ctx context.Context, workflowName string, state *WorkflowState) error
	LoadState(ctx context.Context, workflowName string, stateID string) (*WorkflowState, error)
}

// EventHandler defines the interface for workflow events
type EventHandler interface {
	OnWorkflowStart(ctx context.Context, workflowName string, state *WorkflowState) error
	OnWorkflowEnd(ctx context.Context, workflowName string, state *WorkflowState) error
	OnStepTransition(ctx context.Context, workflowName string, fromStep, toStep string, state *WorkflowState) error
}