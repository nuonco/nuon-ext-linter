package rules

import (
	"github.com/nuonco/nuon-ext-linter/internal/config"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

func DefaultRules(cfg *config.Config) []rule.Rule {
	labelsRule := &RequireLabels{}
	if rc, ok := cfg.Rules["require-labels"]; ok {
		labelsRule.RequiredKeys = rc.RequiredLabels
	}

	permissionsRule := &NoAdminPermissions{}
	if rc, ok := cfg.Rules["no-admin-permissions"]; ok {
		permissionsRule.ExtraBlockedPolicies = rc.ExtraBlockedPolicies
	}

	return []rule.Rule{
		labelsRule,
		permissionsRule,
		&ComponentNuonToml{},
		&ExpectedDirectories{},
		&SandboxUseTag{},
		&RunnerInitScript{},
		&PermissionsBoundaryScope{},
		&NoWildcardActions{},
		&RequirePermissionsBoundary{},
		&RequirePolicyTests{},
	}
}
