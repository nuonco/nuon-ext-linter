package rules

import (
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

// ConfigParse surfaces errors from nuon/pkg/config/parse.ParseDir.
type ConfigParse struct{}

func (r *ConfigParse) ID() string          { return "config-parse" }
func (r *ConfigParse) Description() string { return "App config must parse with nuon/pkg/config" }

func (r *ConfigParse) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding
	for _, msg := range ctx.ParseErrors {
		findings = append(findings, rule.Finding{
			RuleID:   r.ID(),
			Severity: rule.SeverityError,
			Message:  msg,
		})
	}
	return findings
}
