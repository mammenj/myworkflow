package main

import (
	"context"
	"fmt"
	"log"
)

// LoggingEventHandler implements EventHandler to log workflow events
type LoggingEventHandler struct{}

// OnWorkflowStart logs the start of a workflow
func (l *LoggingEventHandler) OnWorkflowStart(ctx context.Context, workflowName string, state *WorkflowState) error {
	log.Printf("Workflow '%s' started with data: %+v", workflowName, state.Data)
	return nil
}

// OnWorkflowEnd logs the end of a workflow
func (l *LoggingEventHandler) OnWorkflowEnd(ctx context.Context, workflowName string, state *WorkflowState) error {
	log.Printf("Workflow '%s' ended at step: %s", workflowName, state.CurrentStep)
	return nil
}

// OnStepTransition logs a step transition
func (l *LoggingEventHandler) OnStepTransition(ctx context.Context, workflowName string, fromStep, toStep string, state *WorkflowState) error {
	log.Printf("Workflow '%s' transitioned from '%s' to '%s'", workflowName, fromStep, toStep)
	return nil
}

// ValidationErrorHandler implements EventHandler to validate workflow data
type ValidationErrorHandler struct{}

// OnWorkflowStart validates the workflow data
func (v *ValidationErrorHandler) OnWorkflowStart(ctx context.Context, workflowName string, state *WorkflowState) error {
	// Example validation - check if required fields are present
	requiredFields := []string{"name", "age"}
	
	for _, field := range requiredFields {
		if _, exists := state.Data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing from workflow data", field)
		}
	}
	
	return nil
}

// OnWorkflowEnd does nothing for validation
func (v *ValidationErrorHandler) OnWorkflowEnd(ctx context.Context, workflowName string, state *WorkflowState) error {
	return nil
}

// OnStepTransition does nothing for validation
func (v *ValidationErrorHandler) OnStepTransition(ctx context.Context, workflowName string, fromStep, toStep string, state *WorkflowState) error {
	return nil
}