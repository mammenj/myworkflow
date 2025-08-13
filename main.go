package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

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

// Rule represents a business rule
type Rule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Content     string `json:"content"`
}

func main() {
	// Create HTTP server
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/workflows", workflowsHandler)
	http.HandleFunc("/rules", rulesHandler)
	http.HandleFunc("/analytics", analyticsHandler)
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/workflows/", workflowDetailHandler)
	http.HandleFunc("/rules/", ruleEditorHandler)

	// API endpoints
	http.HandleFunc("/api/workflows", workflowsAPIHandler)
	http.HandleFunc("/api/workflows/", workflowAPIHandler)
	http.HandleFunc("/api/rules", rulesAPIHandler)
	http.HandleFunc("/api/rules/", ruleAPIHandler)

	// Static file serving
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Start server
	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handlers for HTML pages
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.ServeFile(w, r, "./static/404.html")
		return
	}
	http.ServeFile(w, r, "./static/index.html")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func workflowsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/workflows.html")
}

func rulesHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/rules.html")
}

func analyticsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/analytics.html")
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/settings.html")
}

func workflowDetailHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/workflow-detail.html")
}

func ruleEditorHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/rule-editor.html")
}

// API handlers
func workflowsAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getWorkflows(w, r)
	case http.MethodPost:
		createWorkflow(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func workflowAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Extract workflow name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	if path == "" {
		http.Error(w, "Workflow name required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getWorkflow(w, r, path)
	case http.MethodPut:
		updateWorkflow(w, r, path)
	case http.MethodDelete:
		// According to requirements, users should not be able to delete workflows
		http.Error(w, "Delete operation not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func rulesAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getRules(w, r)
	case http.MethodPost:
		createRule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func ruleAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Extract rule name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/rules/")
	if path == "" {
		http.Error(w, "Rule name required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getRule(w, r, path)
	case http.MethodPut:
		updateRule(w, r, path)
	case http.MethodDelete:
		// According to requirements, users should not be able to delete rules
		http.Error(w, "Delete operation not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Implementation of API functions
func getWorkflows(w http.ResponseWriter, r *http.Request) {
	// Read workflow files from the workflows directory
	workflowsDir := "./workflows"
	files, err := ioutil.ReadDir(workflowsDir)
	if err != nil {
		http.Error(w, "Failed to read workflows directory", http.StatusInternalServerError)
		return
	}

	var workflows []Workflow
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			filePath := filepath.Join(workflowsDir, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			var wf Workflow
			if err := yaml.Unmarshal(data, &wf); err != nil {
				continue
			}
			workflows = append(workflows, wf)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflows)
}

func getWorkflow(w http.ResponseWriter, r *http.Request, name string) {
	// Construct file path
	fileName := fmt.Sprintf("%s.yml", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	if _, err := os.Stat(filepath.Join("./workflows", fileName)); os.IsNotExist(err) {
		// Try with .yaml extension
		fileName = fmt.Sprintf("%s.yaml", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	}

	filePath := filepath.Join("./workflows", fileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		http.Error(w, "Failed to parse workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wf)
}

func createWorkflow(w http.ResponseWriter, r *http.Request) {
	var wf Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save workflow to file
	fileName := fmt.Sprintf("%s.yml", strings.ToLower(strings.ReplaceAll(wf.Name, " ", "_")))
	filePath := filepath.Join("./workflows", fileName)

	data, err := yaml.Marshal(wf)
	if err != nil {
		http.Error(w, "Failed to marshal workflow", http.StatusInternalServerError)
		return
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wf)
}

func updateWorkflow(w http.ResponseWriter, r *http.Request, name string) {
	var wf Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save workflow to file
	fileName := fmt.Sprintf("%s.yml", strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	filePath := filepath.Join("./workflows", fileName)

	data, err := yaml.Marshal(wf)
	if err != nil {
		http.Error(w, "Failed to marshal workflow", http.StatusInternalServerError)
		return
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wf)
}

func getRules(w http.ResponseWriter, r *http.Request) {
	// Read rule files from the rules directory
	rulesDir := "./rules"
	files, err := ioutil.ReadDir(rulesDir)
	if err != nil {
		http.Error(w, "Failed to read rules directory", http.StatusInternalServerError)
		return
	}

	var rules []Rule
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".lua") {
			filePath := filepath.Join(rulesDir, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			ruleName := strings.TrimSuffix(file.Name(), ".lua")
			rules = append(rules, Rule{
				Name:     ruleName,
				Language: "Lua",
				Content:  string(content),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func getRule(w http.ResponseWriter, r *http.Request, name string) {
	// Construct file path
	fileName := fmt.Sprintf("%s.lua", name)
	filePath := filepath.Join("./rules", fileName)

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	rule := Rule{
		Name:     name,
		Language: "Lua",
		Content:  string(content),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func createRule(w http.ResponseWriter, r *http.Request) {
	var rule Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save rule to file
	fileName := fmt.Sprintf("%s.lua", rule.Name)
	filePath := filepath.Join("./rules", fileName)

	if err := ioutil.WriteFile(filePath, []byte(rule.Content), 0644); err != nil {
		http.Error(w, "Failed to save rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func updateRule(w http.ResponseWriter, r *http.Request, name string) {
	var rule Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save rule to file
	fileName := fmt.Sprintf("%s.lua", name)
	filePath := filepath.Join("./rules", fileName)

	if err := ioutil.WriteFile(filePath, []byte(rule.Content), 0644); err != nil {
		http.Error(w, "Failed to save rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}
