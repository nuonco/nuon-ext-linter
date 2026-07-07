package rules

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

var defaultBlockedPolicies = map[string][]string{
	"aws": {
		"AdministratorAccess",
		"PowerUserAccess",
	},
	"gcp": {
		"roles/owner",
		"roles/editor",
	},
	"azure": {
		"Owner",
		"Contributor",
	},
}

type NoAdminPermissions struct {
	ExtraBlockedPolicies []string
}

func (r *NoAdminPermissions) ID() string          { return "no-admin-permissions" }
func (r *NoAdminPermissions) Description() string { return "IAM permissions must not include admin or overly broad policies" }

func (r *NoAdminPermissions) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	if ctx.App.Permissions != nil {
		findings = append(findings, r.checkPermissions(ctx.App.Permissions, ctx.Platform, "permissions.toml")...)
	}
	if ctx.App.BreakGlass != nil {
		for _, role := range ctx.App.BreakGlass.Roles {
			findings = append(findings, r.checkRole(&role, ctx.Platform, "break_glass.toml")...)
		}
	}

	// Check individual permission role files from permissions/ directory
	for _, pr := range ctx.App.PermissionRoles {
		role := &appconfig.Role{
			Name:                pr.Name,
			PermissionsBoundary: pr.PermissionsBoundary,
			Policies:            pr.Policies,
		}
		relFile := filepath.Join("permissions", filepath.Base(pr.File))
		findings = append(findings, r.checkRole(role, ctx.Platform, relFile)...)
	}

	return findings
}

func (r *NoAdminPermissions) checkPermissions(perms *appconfig.Permissions, platform, file string) []rule.Finding {
	var findings []rule.Finding

	roles := []*appconfig.Role{perms.ProvisionRole, perms.DeprovisionRole, perms.MaintenanceRole}
	for _, role := range roles {
		if role != nil {
			findings = append(findings, r.checkRole(role, platform, file)...)
		}
	}
	for i := range perms.CustomRoles {
		findings = append(findings, r.checkRole(&perms.CustomRoles[i], platform, file)...)
	}

	return findings
}

func (r *NoAdminPermissions) checkRole(role *appconfig.Role, platform, file string) []rule.Finding {
	var findings []rule.Finding
	blocked := r.blockedPolicies(platform)

	for _, policy := range role.Policies {
		if policy.ManagedPolicyName != "" {
			for _, b := range blocked {
				if strings.EqualFold(policy.ManagedPolicyName, b) {
					findings = append(findings, rule.Finding{
						RuleID:   r.ID(),
						Severity: rule.SeverityError,
						Message:  fmt.Sprintf("role %q uses blocked managed policy %q", role.Name, policy.ManagedPolicyName),
						File:     file,
					})
				}
			}
		}

		if policy.Contents != "" {
			findings = append(findings, r.checkInlinePolicy(role.Name, policy, file)...)
		}
	}

	return findings
}

func (r *NoAdminPermissions) checkInlinePolicy(roleName string, policy appconfig.Policy, file string) []rule.Finding {
	var findings []rule.Finding

	var doc struct {
		Statement []struct {
			Effect   string `json:"Effect"`
			Action   any    `json:"Action"`
			Resource any    `json:"Resource"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal([]byte(policy.Contents), &doc); err != nil {
		return nil // not JSON or has template variables — skip
	}

	for _, stmt := range doc.Statement {
		if !strings.EqualFold(stmt.Effect, "Allow") {
			continue
		}
		if isWildcard(stmt.Action) && isWildcard(stmt.Resource) {
			policyName := policy.Name
			if policyName == "" {
				policyName = "(inline)"
			}
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityError,
				Message:  fmt.Sprintf("role %q has inline policy %q with Action:* and Resource:* (full admin)", roleName, policyName),
				File:     file,
			})
		}
	}

	return findings
}

func isWildcard(v any) bool {
	switch val := v.(type) {
	case string:
		return val == "*"
	case []any:
		for _, item := range val {
			if s, ok := item.(string); ok && s == "*" {
				return true
			}
		}
	}
	return false
}

func (r *NoAdminPermissions) blockedPolicies(platform string) []string {
	var blocked []string

	if platform != "" {
		blocked = append(blocked, defaultBlockedPolicies[platform]...)
	} else {
		// Check all platforms if unknown
		for _, policies := range defaultBlockedPolicies {
			blocked = append(blocked, policies...)
		}
	}

	blocked = append(blocked, r.ExtraBlockedPolicies...)
	return blocked
}
