package rule

import (
	"encoding/json"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func ParseSeverity(s string) Severity {
	switch s {
	case "info":
		return SeverityInfo
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	default:
		return SeverityWarning
	}
}

type Finding struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	File     string   `json:"file,omitempty"`
}

type LintContext struct {
	Dir      string
	App      *appconfig.AppConfig
	Platform string // "aws", "azure", "gcp", ""
}

type Rule interface {
	ID() string
	Description() string
	Run(ctx *LintContext) []Finding
}

type Registry struct {
	rules []Rule
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

func (r *Registry) All() []Rule {
	return r.rules
}
