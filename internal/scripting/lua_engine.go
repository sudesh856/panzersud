package scripting

import (
	"fmt"
	"math/rand"
	"os"

	lua "github.com/yuin/gopher-lua"
)

type LuaEngine struct {
	state *lua.LState
}

func NewLua(src string) (*LuaEngine, error) {
	L := lua.NewState()

	L.SetGlobal("__env", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		L.Push(lua.LString(os.Getenv(key)))
		return 1
	}))

	L.SetGlobal("__randomInt", L.NewFunction(func(L *lua.LState) int {
		min := L.CheckInt(1)
		max := L.CheckInt(2)
		if max <= min {
			L.Push(lua.LNumber(min))
			return 1
		}
		L.Push(lua.LNumber(min + rand.Intn(max-min)))
		return 1
	}))

	L.SetGlobal("__uuidv4", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(pseudoUUID()))
		return 1
	}))

	if err := L.DoString(src); err != nil {
		L.Close()
		return nil, fmt.Errorf("lua compile/run error: %w", err)
	}

	mainFn := L.GetGlobal("main")
	if mainFn == lua.LNil {
		L.Close()
		return nil, fmt.Errorf("lua script must define a function named `main`")
	}
	if mainFn.Type() != lua.LTFunction {
		L.Close()
		return nil, fmt.Errorf("lua `main` must be a function, got %s", mainFn.Type())
	}

	return &LuaEngine{state: L}, nil
}

func NewLuaFromFile(path string) (*LuaEngine, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read lua script %q: %w", path, err)
	}
	return NewLua(string(src))
}

func (e *LuaEngine) Call() (*RequestOverride, error) {
	L := e.state

	// call main() with 0 args, expect 1 return value
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("main"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return nil, fmt.Errorf("lua main() error: %w", err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	tbl, ok := ret.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("lua main() must return a table, got %s", ret.Type())
	}

	override := &RequestOverride{
		Headers: make(map[string]string),
	}

	if m := tbl.RawGetString("method"); m != lua.LNil {
		override.Method = m.String()
	}

	if b := tbl.RawGetString("body"); b != lua.LNil {
		override.Body = b.String()
	}

	if h := tbl.RawGetString("headers"); h != lua.LNil {
		if htbl, ok := h.(*lua.LTable); ok {
			htbl.ForEach(func(k, v lua.LValue) {
				override.Headers[k.String()] = v.String()
			})
		}
	}

	return override, nil
}

func (e *LuaEngine) Close() {
	e.state.Close()
}

type LuaScriptPool struct {
	src string
}

func NewLuaScriptPool(src string) *LuaScriptPool {
	return &LuaScriptPool{src: src}
}

func NewLuaScriptPoolFromFile(path string) (*LuaScriptPool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read lua script %q: %w", path, err)
	}
	return &LuaScriptPool{src: string(src)}, nil
}

func (sp *LuaScriptPool) Clone() (*LuaEngine, error) {
	return NewLua(sp.src)
}
