package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"
)

// WorkflowEngine is the core orchestrator.
type WorkflowEngine struct {
	workflows     map[string]Workflow
	ruleEngine    RuleEngine
	storage       WorkflowStorage
	stateStorage  StateStorage
	eventHandlers []EventHandler
	luaPool       *LuaStatePool // A pool of Lua states for performance
	mu            sync.RWMutex
}

// EngineOptions contains configuration for the workflow engine
type EngineOptions struct {
	WorkflowsDir  string
	RulesDir      string
	LuaPoolSize   int
	EventHandlers []EventHandler
}

// NewWorkflowEngine creates a new engine and loads workflows from a directory.
func NewWorkflowEngine(opts EngineOptions) (*WorkflowEngine, error) {
	if opts.LuaPoolSize <= 0 {
		opts.LuaPoolSize = 10
	}

	engine := &WorkflowEngine{
		workflows:     make(map[string]Workflow),
		luaPool:       NewLuaStatePool(opts.LuaPoolSize),
		eventHandlers: opts.EventHandlers,
	}

	// Initialize default rule engine
	engine.ruleEngine = NewLuaRuleEngine(engine.luaPool, opts.RulesDir)

	// Load workflows
	if err := engine.loadWorkflows(opts.WorkflowsDir); err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	// Register a simple "pass" rule that always returns true
	if err := engine.registerPassRule(); err != nil {
		return nil, fmt.Errorf("failed to register pass rule: %w", err)
	}

	return engine, nil
}

// RegisterWorkflow adds a new workflow definition to the engine.
func (e *WorkflowEngine) RegisterWorkflow(wf Workflow) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.workflows[wf.Name] = wf
}

// SetRuleEngine allows setting a custom rule engine
func (e *WorkflowEngine) SetRuleEngine(engine RuleEngine) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ruleEngine = engine
}

// SetStorage allows setting custom workflow storage
func (e *WorkflowEngine) SetStorage(storage WorkflowStorage) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.storage = storage
}

// SetStateStorage allows setting custom state storage
func (e *WorkflowEngine) SetStateStorage(storage StateStorage) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stateStorage = storage
}

// AddEventHandler adds an event handler to the engine
func (e *WorkflowEngine) AddEventHandler(handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eventHandlers = append(e.eventHandlers, handler)
}

func (e *WorkflowEngine) registerPassRule() error {
	// For simplicity, we can have an in-memory rule that always returns true.
	// This avoids creating a dedicated file for it.
	passRuleFunc := func(l *lua.LState) int {
		l.Push(lua.LBool(true))
		return 1
	}

	l := e.luaPool.Get()
	defer e.luaPool.Put(l)
	
	l.SetGlobal("pass", l.NewFunction(passRuleFunc))
	return nil
}

// RunWorkflow executes a workflow from a given state.
func (e *WorkflowEngine) RunWorkflow(ctx context.Context, wfName string, state *WorkflowState) error {
	e.mu.RLock()
	wf, ok := e.workflows[wfName]
	e.mu.RUnlock()
	
	if !ok {
		return fmt.Errorf("workflow '%s' not found", wfName)
	}

	// Trigger workflow start event
	for _, handler := range e.eventHandlers {
		if err := handler.OnWorkflowStart(ctx, wfName, state); err != nil {
			return fmt.Errorf("workflow start event handler failed: %w", err)
		}
	}

	// Set the initial step if the state is new.
	if state.CurrentStep == "" {
		state.CurrentStep = wf.StartStep
	}

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Find the transition for the current step.
		var currentTransition *Transition
		for _, t := range wf.Transitions {
			if t.FromStep == state.CurrentStep {
				currentTransition = &t
				break
			}
		}

		if currentTransition == nil {
			fmt.Printf("Workflow finished at step: %s\n", state.CurrentStep)
			
			// Trigger workflow end event
			for _, handler := range e.eventHandlers {
				if err := handler.OnWorkflowEnd(ctx, wfName, state); err != nil {
					return fmt.Errorf("workflow end event handler failed: %w", err)
				}
			}
			
			return nil
		}

		// Evaluate the rule.
		ruleResult, err := e.ruleEngine.Evaluate(ctx, currentTransition.RuleName, state.Data)
		if err != nil {
			return fmt.Errorf("failed to evaluate rule '%s': %w", currentTransition.RuleName, err)
		}

		previousStep := state.CurrentStep
		if ruleResult {
			state.CurrentStep = currentTransition.ToStep
			fmt.Printf("Transitioning from '%s' to '%s' (Rule '%s' passed)\n", currentTransition.FromStep, state.CurrentStep, currentTransition.RuleName)
		} else {
			state.CurrentStep = currentTransition.FallbackStep
			fmt.Printf("Transitioning from '%s' to '%s' (Rule '%s' failed)\n", currentTransition.FromStep, state.CurrentStep, currentTransition.RuleName)
		}

		// Trigger step transition event
		for _, handler := range e.eventHandlers {
			if err := handler.OnStepTransition(ctx, wfName, previousStep, state.CurrentStep, state); err != nil {
				return fmt.Errorf("step transition event handler failed: %w", err)
			}
		}
	}
}

// loadWorkflows reads all YAML files from a directory and loads them as workflows.
func (e *WorkflowEngine) loadWorkflows(workflowsDir string) error {
	files, err := os.ReadDir(workflowsDir)
	if err != nil {
		return fmt.Errorf("failed to read workflows directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			filePath := filepath.Join(workflowsDir, file.Name())
			wf, err := e.loadWorkflowFromFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to load workflow from %s: %w", filePath, err)
			}
			e.RegisterWorkflow(*wf)
		}
	}

	return nil
}

// loadWorkflowFromFile reads a YAML file and unmarshals it into a Workflow struct.
func (e *WorkflowEngine) loadWorkflowFromFile(filePath string) (*Workflow, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, err
	}

	return &wf, nil
}