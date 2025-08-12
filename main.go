package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type CustomerData struct {
	Age          int
	CustomerType string
	Name         string
}

// ReadCustomerDataCSV reads customers from a CSV file with headers: age,customer_type,name
func ReadCustomerDataCSV(filename string) ([]CustomerData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open customer data file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read customer data file: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("customer data file should have a header and at least one row")
	}

	var customers []CustomerData
	for i, row := range rows {
		if i == 0 {
			// skip header
			continue
		}
		if len(row) != 3 {
			return nil, fmt.Errorf("row %d has %d columns, want 3", i+1, len(row))
		}
		age, err := strconv.Atoi(row[0])
		if err != nil {
			return nil, fmt.Errorf("invalid age at row %d: %v", i+1, err)
		}
		customers = append(customers, CustomerData{
			Age:          age,
			CustomerType: row[1],
			Name:         row[2],
		})
	}
	return customers, nil
}

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

	// 4. Read customer data from CSV file
	customerFile := "customers.csv"
	customers, err := ReadCustomerDataCSV(customerFile)
	if err != nil {
		log.Fatalf("Could not read customer data: %v", err)
	}

	// 5. Run the workflows with context for cancellation
	timeout := time.Duration(config.WorkflowTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, customer := range customers {
		wfState := &WorkflowState{
			Data: map[string]any{
				"age":           customer.Age,
				"customer_type": customer.CustomerType,
				"name":          customer.Name,
			},
		}
		fmt.Printf("--- Running Workflow for %s ---\n", customer.Name)
		err = engine.RunWorkflow(ctx, "CustomerOnboarding", wfState)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Printf("Final step for %s: %s\n\n", customer.Name, wfState.CurrentStep)
	}
}
