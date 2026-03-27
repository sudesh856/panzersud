package scenario_test

import (
	"math"
	"testing"

	"github.com/sudesh856/suddpanzer/internal/scenario"
)

func TestPickEndpoint_Distribution(t *testing.T) {
	endpoints := []scenario.Endpoint{
		{URL: "https://a.example.com", Weight: 70},
		{URL: "https://b.example.com", Weight: 30},
	}

	const iterations = 10_000
	counts := map[string]int{}

	for i := 0; i < iterations; i++ {
		ep := scenario.PickEndpoint(endpoints)
		if ep == nil {
			t.Fatal("PickEndpoint returned nil")
		}
		counts[ep.URL]++
	}


	expectedA := 0.70
	actualA := float64(counts["https://a.example.com"]) / iterations
	if math.Abs(actualA-expectedA) > 0.05 {
		t.Errorf("endpoint A: expected ~70%%, got %.1f%%", actualA*100)
	}

	expectedB := 0.30
	actualB := float64(counts["https://b.example.com"]) / iterations
	if math.Abs(actualB-expectedB) > 0.05 {
		t.Errorf("endpoint B: expected ~30%%, got %.1f%%", actualB*100)
	}
}

func TestPickEndpoint_Single(t *testing.T) {
	endpoints := []scenario.Endpoint{
		{URL: "https://only.example.com", Weight: 1},
	}
	ep := scenario.PickEndpoint(endpoints)
	if ep == nil || ep.URL != "https://only.example.com" {
		t.Errorf("expected only endpoint, got %v", ep)
	}
}

func TestPickEndpoint_Empty(t *testing.T) {
	ep := scenario.PickEndpoint(nil)
	if ep != nil {
		t.Errorf("expected nil for empty slice, got %v", ep)
	}
}