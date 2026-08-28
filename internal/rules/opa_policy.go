package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/open-policy-agent/opa/v1/ast"
)

// OPAPolicyValid checks Rego policies parse and follow Nuon conventions.
type OPAPolicyValid struct{}

func (r *OPAPolicyValid) ID() string          { return "opa-policy-valid" }
func (r *OPAPolicyValid) Description() string { return "OPA/Rego policies must parse and define package nuon with deny or warn rules" }

func (r *OPAPolicyValid) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	if ctx.NuonConfig != nil && ctx.NuonConfig.Policies != nil {
		for _, policy := range ctx.NuonConfig.Policies.Policies {
			if policy.Engine != "" && policy.Engine != "opa" {
				continue
			}
			contents := policyContents(ctx, policy)
			if contents == "" {
				continue
			}
			file := policySourcePath(policy)
			if err := validate.ValidateOPAPolicy(contents); err != nil {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityError,
					Message:  fmt.Sprintf("policy %q: %v", policyName(policy), err),
					File:     file,
				})
			}
		}
	}

	if ctx.App != nil {
		for _, p := range ctx.App.OPAPolicies {
			data, err := os.ReadFile(p.File)
			if err != nil {
				continue
			}
			rel, _ := filepath.Rel(ctx.Dir, p.File)
			if err := validate.ValidateOPAPolicy(string(data)); err != nil {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityError,
					Message:  fmt.Sprintf("policy %q: %v", p.Name, err),
					File:     rel,
				})
			}
		}
	}

	return findings
}

// OPAPolicyBlocksCreate warns when deny rules likely block all managed resource changes.
type OPAPolicyBlocksCreate struct{}

func (r *OPAPolicyBlocksCreate) ID() string          { return "opa-policy-blocks-create" }
func (r *OPAPolicyBlocksCreate) Description() string { return "Warn when OPA deny rules may block all resource creation" }

func (r *OPAPolicyBlocksCreate) Run(ctx *rule.LintContext) []rule.Finding {
	if ctx.NuonConfig == nil || ctx.NuonConfig.Policies == nil {
		return nil
	}

	var findings []rule.Finding
	for _, policy := range ctx.NuonConfig.Policies.Policies {
		if policy.Engine != "" && policy.Engine != "opa" {
			continue
		}
		contents := policyContents(ctx, policy)
		if contents == "" {
			continue
		}
		if reason := analyzeBlocksCreate(contents); reason != "" {
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityWarning,
				Message:  fmt.Sprintf("policy %q: %s", policyName(policy), reason),
				File:     policySourcePath(policy),
			})
		}
	}

	return findings
}

func policyContents(ctx *rule.LintContext, policy config.AppPolicy) string {
	if strings.Contains(policy.Contents, "\n") || !strings.HasSuffix(policy.Contents, ".rego") {
		return policy.Contents
	}
	path := policy.Contents
	if !filepath.IsAbs(path) {
		path = filepath.Join(ctx.Dir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func policySourcePath(policy config.AppPolicy) string {
	if strings.HasSuffix(policy.Contents, ".rego") && !strings.Contains(policy.Contents, "\n") {
		return policy.Contents
	}
	if policy.GetSourceFile() != "" {
		return policy.GetSourceFile()
	}
	return "policies.toml"
}

func policyName(policy config.AppPolicy) string {
	if policy.Name != "" {
		return policy.Name
	}
	return string(policy.Type)
}

func analyzeBlocksCreate(contents string) string {
	module, err := ast.ParseModule("policy.rego", contents)
	if err != nil {
		return ""
	}

	for _, rule := range module.Rules {
		if rule.Head.Name != "deny" {
			continue
		}
		body := strings.ToLower(rule.Body.String())
		if body == "true" || strings.Contains(body, " true") {
			return "contains an unconditional deny rule that will always fail"
		}
		if strings.Contains(body, "resource.mode") && strings.Contains(body, `"managed"`) &&
			!strings.Contains(body, "resource.type") {
			return "denies all managed resources (blocks creates and updates); intentional only for read-only lookup components"
		}
	}

	return ""
}
