package scripting

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/dop251/goja"
)

type RequestOverride struct {
	Method  string
	Headers map[string]string
	Body    string
}

type Engine struct {
	runtime *goja.Runtime
	fn      goja.Callable
}

func New(src string) (*Engine, error) {
	rt := goja.New()

	rt.Set("__env", func(key string) string {
		return os.Getenv(key)
	})
	rt.Set("__randomInt", func(min, max int) int {
		if max <= min {
			return min
		}

		return min + rand.Intn(max-min)
	})

	rt.Set("__uuidv4", func() string {
		return pseudoUUID()
	})
	// ---

	prog, err := goja.Compile("script.js", src, false)
	if err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}

	_, err = rt.RunProgram(prog)
	if err != nil {
		return nil, fmt.Errorf("runtime error: %w", err)
	}

	defaultExport := rt.Get("main")
	if defaultExport == nil || goja.IsUndefined(defaultExport) {
		return nil, fmt.Errorf("script must define a function named `main`")
	}

	fn, ok := goja.AssertFunction(defaultExport)
	if !ok {
		return nil, fmt.Errorf("`default` export must be a function, got %T", defaultExport)
	}

	return &Engine{runtime: rt, fn: fn}, nil
}

func NewFromFile(path string) (*Engine, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read script file %q: %w", path, err)
	}
	return New(string(src))
}

func (e *Engine) Call() (*RequestOverride, error) {
	val, err := e.fn(goja.Undefined())
	if err != nil {
		return nil, fmt.Errorf("script execution error: %w", err)
	}

	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil, fmt.Errorf("script main() must return an object")
	}

	exported := val.Export()
	m, ok := exported.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("script main() must return an object, got %T", exported)
	}

	override := &RequestOverride{
		Headers: make(map[string]string),
	}

	if method, ok := m["method"].(string); ok && method != "" {
		override.Method = method
	}

	if body, ok := m["body"].(string); ok {
		override.Body = body
	}

	if headers, ok := m["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if sv, ok := v.(string); ok {
				override.Headers[k] = sv
			} else {
				override.Headers[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return override, nil
}

type ScriptPool struct {
	src string
}

func NewScriptPool(src string) *ScriptPool {
	return &ScriptPool{src: src}
}

func NewScriptPoolFromFile(path string) (*ScriptPool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read script file %q: %w", path, err)
	}
	return &ScriptPool{src: string(src)}, nil
}

func (sp *ScriptPool) Clone() (*Engine, error) {
	return New(sp.src)
}
