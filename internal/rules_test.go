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
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"action":"opened","merged":false}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(matches))
	}
	if matches[0].Topic != "pr.opened" {
		t.Fatalf("expected topic pr.opened, got %q", matches[0].Topic)
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
		Provider:   "github",
		Name:       "push",
		RawPayload: []byte(`{}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 0 {
		t.Fatalf("expected no topics, got %d", len(matches))
	}
}

func TestRuleEngineWithDrivers(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "action == \"opened\"", Emit: "pr.opened", Drivers: []string{"amqp", "http"}},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"action":"opened"}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if len(matches[0].Drivers) != 2 {
		t.Fatalf("expected 2 drivers, got %d", len(matches[0].Drivers))
	}
}

func TestRuleEngineJSONPathDot(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "$.pull_request.draft == false", Emit: "pr.opened"},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"pull_request":{"draft":false}}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestRuleEngineJSONPathIndex(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "$.pull_request[0].draft == false", Emit: "pr.opened"},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"pull_request":[{"draft":false}]}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestRuleEngineJSONPathFilter(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "$.pull_request[?(@.draft==false)][0].draft == false", Emit: "pr.opened"},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"pull_request":[{"draft":false},{"draft":true}]}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestRuleEngineBareJSONPath(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "action == \"opened\" && pull_request.draft == false", Emit: "pr.opened"},
			{When: "pull_requests[?(@.draft==false)][0].draft == false", Emit: "pr.any"},
		},
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"action":"opened","pull_request":{"draft":false},"pull_requests":[{"draft":false}]}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestRuleEngineStrictMissing(t *testing.T) {
	cfg := RulesConfig{
		Rules: []Rule{
			{When: "missing_field == true", Emit: "never"},
		},
		Strict: true,
	}

	engine, err := NewRuleEngine(cfg)
	if err != nil {
		t.Fatalf("new rule engine: %v", err)
	}

	event := Event{
		Provider:   "github",
		Name:       "pull_request",
		RawPayload: []byte(`{"action":"opened"}`),
	}

	matches := engine.Evaluate(event)
	if len(matches) != 0 {
		t.Fatalf("expected no matches in strict mode, got %d", len(matches))
	}
}
