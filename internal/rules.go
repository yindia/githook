package internal

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/PaesslerAG/jsonpath"
)

type Rule struct {
	When    string   `yaml:"when"`
	Emit    string   `yaml:"emit"`
	Drivers []string `yaml:"drivers"`
}

type compiledRule struct {
	emit    string
	drivers []string
	vars    []string
	varMap  map[string]string
	expr    *govaluate.EvaluableExpression
}

type RuleEngine struct {
	rules  []compiledRule
	strict bool
}

type RuleMatch struct {
	Topic   string
	Drivers []string
}

func NewRuleEngine(cfg RulesConfig) (*RuleEngine, error) {
	rules := make([]compiledRule, 0, len(cfg.Rules))
	for _, rule := range cfg.Rules {
		rewritten, varMap := rewriteExpression(rule.When)
		expr, err := govaluate.NewEvaluableExpression(rewritten)
		if err != nil {
			return nil, err
		}
		rules = append(rules, compiledRule{
			emit:    rule.Emit,
			drivers: rule.Drivers,
			vars:    expr.Vars(),
			varMap:  varMap,
			expr:    expr,
		})
	}

	return &RuleEngine{rules: rules, strict: cfg.Strict}, nil
}

func (r *RuleEngine) Evaluate(event Event) []RuleMatch {
	if len(r.rules) == 0 {
		return nil
	}

	matches := make([]RuleMatch, 0, 1)
	for _, rule := range r.rules {
		params, missing := resolveRuleParams(event, rule.vars, rule.varMap)
		log.Printf("rule debug: when=%q params=%v", rule.expr.String(), params)
		if r.strict && len(missing) > 0 {
			log.Printf("rule strict missing params: %v", missing)
			continue
		}
		result, err := rule.expr.Evaluate(params)
		if err != nil {
			log.Printf("rule eval failed: %v", err)
			continue
		}
		ok, _ := result.(bool)
		if ok {
			matches = append(matches, RuleMatch{Topic: rule.emit, Drivers: rule.drivers})
		}
	}
	return matches
}

func resolveRuleParams(event Event, vars []string, varMap map[string]string) (map[string]interface{}, []string) {
	if len(vars) == 0 {
		if len(event.RawPayload) == 0 {
			return event.Data, nil
		}
		return nil, nil
	}

	params := make(map[string]interface{}, len(vars))
	missing := make([]string, 0)
	for _, name := range vars {
		if path, ok := varMap[name]; ok {
			value, err := resolveJSONPath(event, path)
			if err != nil {
				missing = append(missing, path)
				log.Printf("rule warn: jsonpath error path=%s err=%v", path, err)
				params[name] = nil
			} else {
				if value == nil {
					missing = append(missing, path)
					log.Printf("rule warn: jsonpath no match path=%s", path)
				}
				params[name] = value
			}
			continue
		}
		if value, ok := event.Data[name]; ok {
			params[name] = value
		} else {
			missing = append(missing, name)
			params[name] = nil
		}
	}
	return params, missing
}

func resolveJSONPath(event Event, path string) (interface{}, error) {
	if event.RawObject != nil {
		value, err := jsonpath.Get(path, event.RawObject)
		if err != nil {
			return nil, err
		}
		return normalizeJSONPathResult(value), nil
	}
	if len(event.RawPayload) == 0 {
		if event.Data != nil {
			value, err := jsonpath.Get(path, event.Data)
			if err != nil {
				return nil, err
			}
			return normalizeJSONPathResult(value), nil
		}
		return nil, nil
	}
	var raw interface{}
	if err := json.Unmarshal(event.RawPayload, &raw); err != nil {
		return nil, err
	}
	value, err := jsonpath.Get(path, raw)
	if err != nil {
		return nil, err
	}
	return normalizeJSONPathResult(value), nil
}

func normalizeJSONPathResult(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return value
	}
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		return items[0]
	}
	return items
}

func rewriteExpression(expr string) (string, map[string]string) {
	var out strings.Builder
	out.Grow(len(expr))

	varMap := make(map[string]string)
	inString := false
	var stringQuote byte

	for i := 0; i < len(expr); {
		ch := expr[i]

		if inString {
			out.WriteByte(ch)
			if ch == '\\' && i+1 < len(expr) {
				out.WriteByte(expr[i+1])
				i += 2
				continue
			}
			if ch == stringQuote {
				inString = false
			}
			i++
			continue
		}

		if ch == '"' || ch == '\'' {
			inString = true
			stringQuote = ch
			out.WriteByte(ch)
			i++
			continue
		}

		if ch == '$' || isIdentStart(ch) {
			token, next := parseJSONPathToken(expr, i)
			if isKeyword(token) {
				out.WriteString(token)
				i = next
				continue
			}
			path := token
			if token[0] != '$' {
				path = "$." + token
			}
			safe := safeVarName(path)
			varMap[safe] = path
			out.WriteString(safe)
			i = next
			continue
		}

		out.WriteByte(ch)
		i++
	}

	return out.String(), varMap
}

func parseJSONPathToken(expr string, start int) (string, int) {
	i := start
	bracketDepth := 0
	parenDepth := 0
	var quote byte

	for i < len(expr) {
		ch := expr[i]

		if quote != 0 {
			if ch == '\\' && i+1 < len(expr) {
				i += 2
				continue
			}
			if ch == quote {
				quote = 0
			}
			i++
			continue
		}

		switch ch {
		case '\'', '"':
			quote = ch
			i++
			continue
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '(':
			if bracketDepth > 0 {
				parenDepth++
			}
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		}

		if bracketDepth == 0 && parenDepth == 0 && isTerminator(ch) {
			break
		}

		i++
	}
	return expr[start:i], i
}

func isTerminator(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', ',', ';':
		return true
	case '+', '-', '*', '/', '%':
		return true
	case '=', '!', '<', '>', '&', '|':
		return true
	case ')':
		return true
	default:
		return false
	}
}

func safeVarName(token string) string {
	var b strings.Builder
	b.Grow(len(token) + 2)
	b.WriteString("v_")
	for i := 0; i < len(token); i++ {
		ch := token[i]
		if isIdentStart(ch) || (ch >= '0' && ch <= '9') {
			b.WriteByte(ch)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isKeyword(token string) bool {
	switch token {
	case "true", "false", "null":
		return true
	default:
		return false
	}
}
