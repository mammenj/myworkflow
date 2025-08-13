package main

// Workflow represents a sequence of steps.
type Workflow1 struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	StartStep   string        `yaml:"start_step"`
	Transitions []Transition1 `yaml:"transitions"`
}

// Transition defines a move from one step to another based on a rule.
type Transition1 struct {
	FromStep     string `yaml:"from"`
	ToStep       string `yaml:"to"`
	RuleName     string `yaml:"rule"`
	FallbackStep string `yaml:"fallback_to"`
}

// WorkflowState represents the current state of a workflow instance.
type WorkflowState struct {
	CurrentStep string
	Data        map[string]any
}
