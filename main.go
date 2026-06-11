package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var (
	engine      *WorkflowEngine
	storage     WorkflowStorage
	ruleStorage RuleStorage
)

func main() {
	// Load config
	cfg, err := LoadConfig("config.txt")
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = &Config{
			WorkflowsDir: "./workflows",
			RulesDir:     "./rules",
			StatesDir:    "./states",
			LuaPoolSize:  10,
		}
	}

	// Initialize engine
	engine, err = NewWorkflowEngine(EngineOptions{
		WorkflowsDir: cfg.WorkflowsDir,
		RulesDir:     cfg.RulesDir,
		LuaPoolSize:  cfg.LuaPoolSize,
	})
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// Initialize and set storage
	storage = NewFileWorkflowStorage(cfg.WorkflowsDir)
	ruleStorage = NewFileRuleStorage(cfg.RulesDir)
	engine.SetStorage(storage)
	engine.SetStateStorage(NewFileStateStorage(cfg.StatesDir))

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
	workflowNames, err := storage.ListWorkflows(r.Context())
	if err != nil {
		http.Error(w, "Failed to list workflows", http.StatusInternalServerError)
		return
	}

	var workflows []Workflow
	for _, name := range workflowNames {
		wf, err := storage.LoadWorkflow(r.Context(), name)
		if err != nil {
			continue
		}
		workflows = append(workflows, *wf)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflows)
}

func getWorkflow(w http.ResponseWriter, r *http.Request, name string) {
	wf, err := storage.LoadWorkflow(r.Context(), name)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
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

	if err := storage.SaveWorkflow(r.Context(), wf); err != nil {
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

	if err := storage.SaveWorkflow(r.Context(), wf); err != nil {
		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wf)
}

func getRules(w http.ResponseWriter, r *http.Request) {
	rules, err := ruleStorage.ListRules(r.Context())
	if err != nil {
		http.Error(w, "Failed to list rules", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func getRule(w http.ResponseWriter, r *http.Request, name string) {
	rule, err := ruleStorage.LoadRule(r.Context(), name)
	if err != nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
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

	if err := ruleStorage.SaveRule(r.Context(), rule); err != nil {
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

	if err := ruleStorage.SaveRule(r.Context(), rule); err != nil {
		http.Error(w, "Failed to save rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}
