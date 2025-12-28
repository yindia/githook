package internal

import (
	"log"

	"github.com/Knetic/govaluate"
)

type Rule struct {
	When string `yaml:"when"`
	Emit string `yaml:"emit"`
}

type compiledRule struct {
	emit string
	expr *govaluate.EvaluableExpression
}

type RuleEngine struct {
	rules []compiledRule
}

func NewRuleEngine(cfg RulesConfig) (*RuleEngine, error) {
	rules := make([]compiledRule, 0, len(cfg.Rules))
	for _, rule := range cfg.Rules {
		expr, err := govaluate.NewEvaluableExpression(rule.When)
		if err != nil {
			return nil, err
		}
		rules = append(rules, compiledRule{emit: rule.Emit, expr: expr})
	}

	return &RuleEngine{rules: rules}, nil
}

func (r *RuleEngine) Evaluate(event Event) []string {
	if len(r.rules) == 0 {
		return nil
	}

	matches := make([]string, 0, 1)
	for _, rule := range r.rules {
		result, err := rule.expr.Evaluate(event.Data)
		if err != nil {
			log.Printf("rule eval failed: %v", err)
			continue
		}
		ok, _ := result.(bool)
		if ok {
			matches = append(matches, rule.emit)
		}
	}
	return matches
}
