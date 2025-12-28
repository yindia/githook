package internal

import "testing"

func TestRuleEngineEvaluate(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "action == \"opened\"", Emit: "pr.opened"},
			{When: "action == \"closed\" && merged == true", Emit: "pr.merged"},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider: "github",
		Name:     "pull_request",
		Data: map[string]interface{}{
			"action": "opened",
			"merged": false,
		},
	}

	topics := engine.Evaluate(event)
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
	if topics[0] != "pr.opened" {
		t.Fatalf("expected topic pr.opened, got %q", topics[0])
	}
}

func TestRuleEngineEvaluateMissingField(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "missing == true", Emit: "never"},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider: "github",
		Name:     "push",
		Data:     map[string]interface{}{},
	}

	topics := engine.Evaluate(event)
	if len(topics) != 0 {
		t.Fatalf("expected no topics, got %d", len(topics))
	}
}
