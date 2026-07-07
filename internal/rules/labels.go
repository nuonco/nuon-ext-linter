package rules

import (
	"fmt"
	"path/filepath"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

type RequireLabels struct {
	RequiredKeys []string
}

func (r *RequireLabels) ID() string          { return "require-labels" }
func (r *RequireLabels) Description() string { return "Components, actions, and runbooks must have labels" }

func (r *RequireLabels) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	for _, c := range ctx.App.Components {
		findings = append(findings, r.checkLabels(c.Labels, c.Name, c.File, "component")...)
	}
	for _, a := range ctx.App.Actions {
		findings = append(findings, r.checkLabels(a.Labels, a.Name, a.File, "action")...)
	}
	for _, rb := range ctx.App.Runbooks {
		findings = append(findings, r.checkLabels(rb.Labels, rb.Name, rb.File, "runbook")...)
	}

	return findings
}

func (r *RequireLabels) checkLabels(labels map[string]string, name, file, kind string) []rule.Finding {
	var findings []rule.Finding
	relFile := filepath.Base(file)

	if len(labels) == 0 {
		findings = append(findings, rule.Finding{
			RuleID:   r.ID(),
			Severity: rule.SeverityWarning,
			Message:  fmt.Sprintf("%s %q has no labels", kind, name),
			File:     relFile,
		})
		return findings
	}

	for _, key := range r.RequiredKeys {
		if _, ok := labels[key]; !ok {
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityWarning,
				Message:  fmt.Sprintf("%s %q is missing required label %q", kind, name, key),
				File:     relFile,
			})
		}
	}

	return findings
}
