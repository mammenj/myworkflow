package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	// 1. Load configuration
	config, err := LoadConfig("config.txt")
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		config = &Config{
			WorkflowsDir:           "./workflows",
			RulesDir:               "./rules",
			StatesDir:              "./states",
			LuaPoolSize:            10,
			WorkflowTimeoutSeconds: 30,
		}
	}

	// 2. Initialize the modular engine with options from config
	opts := EngineOptions{
		WorkflowsDir: config.WorkflowsDir,
		RulesDir:     config.RulesDir,
		LuaPoolSize:  config.LuaPoolSize,
		EventHandlers: []EventHandler{
			&LoggingEventHandler{},
			&ValidationErrorHandler{},
		},
	}

	engine, err := NewWorkflowEngine(opts)
	if err != nil {
		log.Fatalf("Failed to initialize workflow engine: %v", err)
	}

	// 3. Set up storage
	workflowStorage := NewFileWorkflowStorage(config.WorkflowsDir)
	stateStorage := NewFileStateStorage(config.StatesDir)

	engine.SetStorage(workflowStorage)
	engine.SetStateStorage(stateStorage)

	// 4. Create a workflow instance with initial data.
	customerData1 := map[string]any{
		"age":           25,
		"customer_type": "premium",
		"name":          "Alice",
	}

	customerData2 := map[string]any{
		"age":           16,
		"customer_type": "standard",
		"name":          "Bob",
	}

	customerData3 := map[string]any{
		"age":           30,
		"customer_type": "standard",
		"name":          "Charlie",
	}

	wfState1 := &WorkflowState{
		Data: customerData1,
	}

	wfState2 := &WorkflowState{
		Data: customerData2,
	}

	wfState3 := &WorkflowState{
		Data: customerData3,
	}

	// 5. Run the workflows with context for cancellation
	timeout := time.Duration(config.WorkflowTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Println("--- Running Workflow for Alice ---")
	err = engine.RunWorkflow(ctx, "CustomerOnboarding", wfState1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Final step for Alice: %s\n\n", wfState1.CurrentStep)

	fmt.Println("--- Running Workflow for Bob ---")
	err = engine.RunWorkflow(ctx, "CustomerOnboarding", wfState2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Final step for Bob: %s\n\n", wfState2.CurrentStep)

	fmt.Println("--- Running Workflow for Charlie ---")
	err = engine.RunWorkflow(ctx, "CustomerOnboarding", wfState3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Final step for Charlie: %s\n\n", wfState3.CurrentStep)
}
