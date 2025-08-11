package main

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// LuaStatePool manages a pool of lua.LState instances.
type LuaStatePool struct {
	pool chan *lua.LState
	lock sync.Mutex
}

// NewLuaStatePool creates a new pool with a given size.
func NewLuaStatePool(size int) *LuaStatePool {
	p := &LuaStatePool{
		pool: make(chan *lua.LState, size),
	}

	for range size {
		p.pool <- lua.NewState()
	}

	return p
}

// Get retrieves a Lua state from the pool.
func (p *LuaStatePool) Get() *lua.LState {
	return <-p.pool
}

// Put returns a Lua state to the pool.
func (p *LuaStatePool) Put(l *lua.LState) {
	// A basic check to prevent a nil state from being returned.
	if l == nil {
		return
	}
	p.pool <- l
}

// Close closes all Lua states in the pool.
func (p *LuaStatePool) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()

	close(p.pool)
	for l := range p.pool {
		l.Close()
	}
}

// LMapToTable converts a Go map to a Lua table.
func LMapToTable(l *lua.LState, data map[string]any) *lua.LTable {
	table := l.NewTable()
	for k, v := range data {
		table.RawSetString(k, GoValueToLua(l, v))
	}
	return table
}

// GoValueToLua converts a Go value to a Lua value.
func GoValueToLua(l *lua.LState, v any) lua.LValue {
	switch val := v.(type) {
	case int:
		return lua.LNumber(val)
	case float64:
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case bool:
		return lua.LBool(val)
	case map[string]any:
		return LMapToTable(l, val)
	// Add more type conversions as needed.
	default:
		return lua.LNil
	}
}