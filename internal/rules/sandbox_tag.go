package rules

import (
	"fmt"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

type SandboxUseTag struct{}

func (r *SandboxUseTag) ID() string          { return "sandbox-use-tag" }
func (r *SandboxUseTag) Description() string { return "Sandbox repository should use a tag instead of a branch" }

func (r *SandboxUseTag) Run(ctx *rule.LintContext) []rule.Finding {
	if ctx.App.Sandbox == nil {
		return nil
	}

	var findings []rule.Finding

	if ref := ctx.App.Sandbox.PublicRepo; ref != nil {
		findings = append(findings, r.checkRef(ref, "public_repo")...)
	}
	if ref := ctx.App.Sandbox.ConnectedRepo; ref != nil {
		findings = append(findings, r.checkRef(ref, "connected_repo")...)
	}

	return findings
}

func (r *SandboxUseTag) checkRef(ref *appconfig.RepoRef, refName string) []rule.Finding {
	if ref.Tag != "" {
		return nil // using a tag, good
	}

	if ref.Branch != "" {
		return []rule.Finding{{
			RuleID:   r.ID(),
			Severity: rule.SeverityWarning,
			Message:  fmt.Sprintf("sandbox %s uses branch %q instead of a pinned tag", refName, ref.Branch),
			File:     "sandbox.toml",
		}}
	}

	return nil
}
