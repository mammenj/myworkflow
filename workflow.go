package main

// Workflow represents a sequence of steps.
type Workflow struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description" yaml:"description"`
	StartStep   string       `json:"start_step" yaml:"start_step"`
	Transitions []Transition `json:"transitions" yaml:"transitions"`
}

// Transition defines a move from one step to another based on a rule.
type Transition struct {
	FromStep     string `json:"from" yaml:"from"`
	ToStep       string `json:"to" yaml:"to"`
	RuleName     string `json:"rule" yaml:"rule"`
	FallbackStep string `json:"fallback_to" yaml:"fallback_to"`
}

// WorkflowState represents the current state of a workflow instance.
type WorkflowState struct {
	CurrentStep string         `json:"current_step"`
	Data        map[string]any `json:"data"`
}

// Rule represents a business rule
type Rule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Content     string `json:"content"`
}
