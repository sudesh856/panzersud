package scenario_test

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sudesh856/suddpanzer/internal/scenario"
)

func TestReplaceVars_UUID(t *testing.T) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	result := scenario.ReplaceVars("id={{uuid}}", nil)
	parts := strings.SplitN(result, "=", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected result: %q", result)
	}
	if !uuidPattern.MatchString(parts[1]) {
		t.Errorf("not a valid UUID v4: %q", parts[1])
	}
}

func TestReplaceVars_Timestamp(t *testing.T) {
	before := time.Now().Unix()
	result := scenario.ReplaceVars("{{timestamp}}", nil)
	after := time.Now().Unix()

	ts, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Fatalf("timestamp not an integer: %q", result)
	}
	if ts < before || ts > after {
		t.Errorf("timestamp %d out of range [%d, %d]", ts, before, after)
	}
}

func TestReplaceVars_RandomInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := scenario.ReplaceVars("{{random_int 10 20}}", nil)
		n, err := strconv.Atoi(result)
		if err != nil {
			t.Fatalf("random_int result not an integer: %q", result)
		}
		if n < 10 || n > 20 {
			t.Errorf("random_int %d out of range [10, 20]", n)
		}
	}
}

func TestReplaceVars_Env(t *testing.T) {
	os.Setenv("SUDD_TEST_VAR", "hello_sudd")
	defer os.Unsetenv("SUDD_TEST_VAR")

	result := scenario.ReplaceVars("val={{env.SUDD_TEST_VAR}}", nil)
	if result != "val=hello_sudd" {
		t.Errorf("expected 'val=hello_sudd', got %q", result)
	}
}

func TestReplaceVars_ExtraMap(t *testing.T) {
	result := scenario.ReplaceVars("user={{user_id}}", map[string]string{"user_id": "42"})
	if result != "user=42" {
		t.Errorf("expected 'user=42', got %q", result)
	}
}

func TestReplaceVars_Unknown(t *testing.T) {
	// Unknown placeholders are left unchanged.
	input := "x={{totally_unknown}}"
	result := scenario.ReplaceVars(input, nil)
	if result != input {
		t.Errorf("expected unknown placeholder to be preserved, got %q", result)
	}
}

func TestReplaceVars_Multiple(t *testing.T) {
	result := scenario.ReplaceVars(`{"id":"{{uuid}}","ts":"{{timestamp}}"}`, nil)
	if strings.Contains(result, "{{") {
		t.Errorf("unreplaced placeholder remaining in: %q", result)
	}
}