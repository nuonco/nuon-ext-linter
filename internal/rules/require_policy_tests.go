package rules

import (
	"fmt"
	"path/filepath"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

// RequirePolicyTests checks that every OPA/Rego policy file has a corresponding
// _test.rego file for validation.
type RequirePolicyTests struct{}

func (r *RequirePolicyTests) ID() string          { return "require-policy-tests" }
func (r *RequirePolicyTests) Description() string { return "OPA/Rego policies must have corresponding test files" }

func (r *RequirePolicyTests) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	for _, policy := range ctx.App.OPAPolicies {
		if !policy.HasTest {
			relFile, _ := filepath.Rel(ctx.Dir, policy.File)
			if relFile == "" {
				relFile = policy.File
			}
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityWarning,
				Message:  fmt.Sprintf("policy %q has no test file (%s_test.rego)", policy.Name, policy.Name),
				File:     relFile,
			})
		}
	}

	return findings
}
