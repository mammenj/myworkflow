package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the workflow engine configuration
type Config struct {
	WorkflowsDir           string
	RulesDir               string
	StatesDir              string
	LuaPoolSize            int
	LogLevel               string
	LogFile                string
	WorkflowTimeoutSeconds int
}

// LoadConfig loads configuration from a file
func LoadConfig(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &Config{
		WorkflowsDir:           "./workflows",
		RulesDir:               "./rules",
		StatesDir:              "./states",
		LuaPoolSize:            10,
		LogLevel:               "info",
		LogFile:                "workflow.log",
		WorkflowTimeoutSeconds: 30,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "workflows_dir":
			config.WorkflowsDir = value
		case "rules_dir":
			config.RulesDir = value
		case "states_dir":
			config.StatesDir = value
		case "lua_pool_size":
			if size, err := strconv.Atoi(value); err == nil {
				config.LuaPoolSize = size
			}
		case "log_level":
			config.LogLevel = value
		case "log_file":
			config.LogFile = value
		case "workflow_timeout_seconds":
			if timeout, err := strconv.Atoi(value); err == nil {
				config.WorkflowTimeoutSeconds = timeout
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return config, nil
}