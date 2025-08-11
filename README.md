# Modular Workflow Engine

This is an enhanced version of the workflow engine with improved flexibility and modularity.

## Key Improvements

1. **Modular Architecture**: The engine is now built with clear interfaces that allow for easy extension and customization.

2. **Plugin System for Rules**: The rule engine is now an interface that can support multiple implementations (Lua, JavaScript, etc.).

3. **Event System**: Added an event system with handlers for workflow lifecycle events.

4. **Storage Abstraction**: Workflow and state storage are now abstracted behind interfaces.

5. **Configuration Management**: Externalized configuration for better deployment flexibility.

6. **Context Support**: Added context support for cancellation and timeouts.

7. **Better Error Handling**: Structured error handling with more context.

## Architecture

The engine follows a modular design with the following key components:

- `WorkflowEngine`: Main orchestrator
- `RuleEngine`: Interface for evaluating business rules
- `WorkflowStorage`: Interface for workflow persistence
- `StateStorage`: Interface for workflow state persistence
- `EventHandler`: Interface for workflow events

## Runtime Rule Updates

One of the key features of this engine is that rules are evaluated at runtime. This means:

1. Rule files are read from disk each time they are evaluated
2. Changes to rule files are immediately reflected in workflow execution
3. You can modify business logic without restarting the engine

This is implemented in the `LuaRuleEngine.Evaluate` method, which reads the rule file from disk on each evaluation:

```go
// Load the rule script.
rulePath := filepath.Join(l.rulesDir, fmt.Sprintf("%s.lua", ruleName))
ruleFile, err := os.ReadFile(rulePath)
// ... evaluate the rule
```

To test this behavior:
1. Run the `test_runtime_rules` program
2. While it's running, modify any of the `.lua` files in the `rules/` directory
3. The next time that rule is evaluated, the updated logic will be used

## Files

- `interfaces.go`: Defines all interfaces for modularity
- `engine.go`: The main workflow engine implementation
- `lua_rule_engine.go`: Lua implementation of the rule engine
- `file_storage.go`: File-based storage implementations
- `event_handlers.go`: Example event handlers
- `config.go`: Configuration loading
- `main.go`: Main function

## Extending the Engine

### Adding a New Rule Engine

To add a new rule engine (e.g., JavaScript-based):

1. Implement the `RuleEngine` interface
2. Register it with the engine using `SetRuleEngine()`

### Adding New Storage

To add new storage (e.g., database):

1. Implement the `WorkflowStorage` and/or `StateStorage` interfaces
2. Register them with the engine using `SetStorage()` and `SetStateStorage()`

### Adding Event Handlers

To add new event handlers:

1. Implement the `EventHandler` interface
2. Register it with the engine using `AddEventHandler()`

## Configuration

The engine can be configured using `config.txt`:

```
# Directory paths
workflows_dir=./workflows
rules_dir=./rules
states_dir=./states

# Lua settings
lua_pool_size=10

# Logging
log_level=info
log_file=workflow.log

# Timeout settings
workflow_timeout_seconds=30
```