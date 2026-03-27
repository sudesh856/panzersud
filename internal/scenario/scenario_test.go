package scenario_test

import (
	"os"
	"testing"

	"github.com/sudesh856/suddpanzer/internal/scenario"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "sudd-test-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoadScenario_Valid(t *testing.T) {
	yaml := `
name: my-test
stages:
  - duration: 10s
    target_vus: 100
  - duration: 20s
    target_vus: 500
  - duration: 10s
    target_vus: 0
endpoints:
  - url: https://httpbin.org/get
    method: GET
    weight: 70
  - url: https://httpbin.org/post
    method: POST
    body: '{"id": "{{uuid}}"}'
    weight: 30
`
	path := writeTempYAML(t, yaml)
	s, err := scenario.LoadScenario(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "my-test" {
		t.Errorf("expected name 'my-test', got %q", s.Name)
	}
	if len(s.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(s.Stages))
	}
	if len(s.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(s.Endpoints))
	}
	// Parsed durations should be set
	if s.Stages[0].ParsedDuration.Seconds() != 10 {
		t.Errorf("stage[0] duration: expected 10s, got %v", s.Stages[0].ParsedDuration)
	}
}

func TestLoadScenario_MissingName(t *testing.T) {
	yaml := `
stages:
  - duration: 10s
    target_vus: 100
endpoints:
  - url: https://example.com
    weight: 1
`
	path := writeTempYAML(t, yaml)
	_, err := scenario.LoadScenario(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoadScenario_MissingURL(t *testing.T) {
	yaml := `
name: bad
stages:
  - duration: 5s
    target_vus: 10
endpoints:
  - method: GET
    weight: 1
`
	path := writeTempYAML(t, yaml)
	_, err := scenario.LoadScenario(path)
	if err == nil {
		t.Fatal("expected error for missing endpoint URL")
	}
}

func TestLoadScenario_ZeroWeight(t *testing.T) {
	yaml := `
name: bad-weight
stages:
  - duration: 5s
    target_vus: 10
endpoints:
  - url: https://example.com
    weight: 0
`
	path := writeTempYAML(t, yaml)
	_, err := scenario.LoadScenario(path)
	if err == nil {
		t.Fatal("expected error for zero weight")
	}
}

func TestLoadScenario_BadDuration(t *testing.T) {
	yaml := `
name: bad-dur
stages:
  - duration: notaduration
    target_vus: 10
endpoints:
  - url: https://example.com
    weight: 1
`
	path := writeTempYAML(t, yaml)
	_, err := scenario.LoadScenario(path)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestDefaultMethod(t *testing.T) {
	yaml := `
name: default-method
stages:
  - duration: 5s
    target_vus: 1
endpoints:
  - url: https://example.com
    weight: 1
`
	path := writeTempYAML(t, yaml)
	s, err := scenario.LoadScenario(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Endpoints[0].Method != "GET" {
		t.Errorf("expected default method GET, got %q", s.Endpoints[0].Method)
	}
}