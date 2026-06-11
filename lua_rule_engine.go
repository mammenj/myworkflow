package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// LuaRuleEngine implements the RuleEngine interface using Lua scripts
type LuaRuleEngine struct {
	luaPool  *LuaStatePool
	rulesDir string
	rules    map[string]lua.LValue
	cache    map[string]*lua.FunctionProto
	mu       sync.RWMutex
}

// NewLuaRuleEngine creates a new Lua-based rule engine
func NewLuaRuleEngine(pool *LuaStatePool, rulesDir string) *LuaRuleEngine {
	return &LuaRuleEngine{
		luaPool:  pool,
		rulesDir: rulesDir,
		rules:    make(map[string]lua.LValue),
		cache:    make(map[string]*lua.FunctionProto),
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

	// Get or compile the script
	proto, err := l.getOrCreateProto(ruleName)
	if err != nil {
		return false, err
	}

	// Push the function onto the stack
	lfunc := state.NewFunctionFromProto(proto)
	state.Push(lfunc)

	// Execute the script to define functions (like 'check')
	if err := state.PCall(0, 0, nil); err != nil {
		return false, fmt.Errorf("failed to execute rule script: %w", err)
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

func (l *LuaRuleEngine) getOrCreateProto(ruleName string) (*lua.FunctionProto, error) {
	l.mu.RLock()
	proto, ok := l.cache[ruleName]
	l.mu.RUnlock()

	if ok {
		return proto, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring lock
	if proto, ok := l.cache[ruleName]; ok {
		return proto, nil
	}

	// Load the rule script.
	rulePath := filepath.Join(l.rulesDir, fmt.Sprintf("%s.lua", ruleName))
	ruleFile, err := os.ReadFile(rulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule file: %w", err)
	}

	// Compile the script
	reader := strings.NewReader(string(ruleFile))
	chunk, err := parse.Parse(reader, ruleName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lua script: %w", err)
	}

	compiled, err := lua.Compile(chunk, ruleName)
	if err != nil {
		return nil, fmt.Errorf("failed to compile lua script: %w", err)
	}

	l.cache[ruleName] = compiled
	return compiled, nil
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