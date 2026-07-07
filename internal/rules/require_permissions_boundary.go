package rules

import (
	"fmt"
	"path/filepath"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

// RequirePermissionsBoundary checks that every lifecycle permission role has a
// permissions_boundary set. A role without a boundary is unconstrained.
// Corresponds to SEC-004 from the byoc security scanner.
type RequirePermissionsBoundary struct{}

func (r *RequirePermissionsBoundary) ID() string { return "require-permissions-boundary" }
func (r *RequirePermissionsBoundary) Description() string {
	return "Permission roles must have a permissions_boundary to constrain effective permissions"
}

func (r *RequirePermissionsBoundary) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	// Check roles from permissions.toml (single-file format)
	if ctx.App.Permissions != nil {
		for _, role := range []*appconfig.Role{
			ctx.App.Permissions.ProvisionRole,
			ctx.App.Permissions.DeprovisionRole,
			ctx.App.Permissions.MaintenanceRole,
		} {
			if role != nil && role.PermissionsBoundary == "" {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityWarning,
					Message:  fmt.Sprintf("role %q has no permissions_boundary — role is unconstrained", role.Name),
					File:     "permissions.toml",
				})
			}
		}
	}

	// Check roles from permissions/ directory (multi-file format)
	for _, role := range ctx.App.PermissionRoles {
		if role.PermissionsBoundary == "" {
			relFile := filepath.Base(role.File)
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityWarning,
				Message:  fmt.Sprintf("role %q has no permissions_boundary — role is unconstrained", role.Name),
				File:     filepath.Join("permissions", relFile),
			})
		}
	}

	return findings
}
