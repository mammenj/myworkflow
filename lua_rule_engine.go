package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// LuaRuleEngine implements the RuleEngine interface using Lua scripts
type LuaRuleEngine struct {
	luaPool  *LuaStatePool
	rulesDir string
	rules    map[string]lua.LValue
}

// NewLuaRuleEngine creates a new Lua-based rule engine
func NewLuaRuleEngine(pool *LuaStatePool, rulesDir string) *LuaRuleEngine {
	return &LuaRuleEngine{
		luaPool:  pool,
		rulesDir: rulesDir,
		rules:    make(map[string]lua.LValue),
	}
}

// Evaluate executes the Lua script and returns the boolean result.
func (l *LuaRuleEngine) Evaluate(ctx context.Context, ruleName string, data map[string]any) (bool, error) {
	// The "pass" rule is handled in-memory
	if ruleName == "pass" {
		return true, nil
	}

	// Get a state from the pool.
	state := l.luaPool.Get()
	defer l.luaPool.Put(state)

	// Load the rule script.
	rulePath := filepath.Join(l.rulesDir, fmt.Sprintf("%s.lua", ruleName))
	ruleFile, err := os.ReadFile(rulePath)
	if err != nil {
		return false, fmt.Errorf("failed to read rule file: %w", err)
	}

	if err := state.DoString(string(ruleFile)); err != nil {
		return false, err
	}

	// Get the 'check' function from the Lua script.
	checkFunc := state.GetGlobal("check")
	if checkFunc.Type() != lua.LTFunction {
		return false, fmt.Errorf("rule '%s' does not have a 'check' function", ruleName)
	}

	// Push the data onto the stack as a Lua table.
	luaData := LMapToTable(state, data)
	state.Push(checkFunc)
	state.Push(luaData)

	// Call the Lua function.
	err = state.PCall(1, 1, nil)
	if err != nil {
		return false, fmt.Errorf("failed to call lua function 'check': %w", err)
	}

	// Get the result from the stack.
	result := state.Get(-1)
	state.Pop(1)

	// Return the result as a boolean.
	if result.Type() != lua.LTBool {
		return false, fmt.Errorf("lua function 'check' did not return a boolean")
	}

	return lua.LVAsBool(result), nil
}

// RegisterRule registers a new rule with the engine
func (l *LuaRuleEngine) RegisterRule(name string, rule any) error {
	// In a more advanced implementation, this could register functions directly
	// For now, we'll just validate that it's a function if it's a Lua value
	if lv, ok := rule.(lua.LValue); ok {
		if lv.Type() == lua.LTFunction {
			l.rules[name] = lv
			return nil
		}
		return fmt.Errorf("rule must be a Lua function")
	}
	
	return fmt.Errorf("unsupported rule type")
}