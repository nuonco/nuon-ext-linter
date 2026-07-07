package rules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

type ExpectedDirectories struct{}

func (r *ExpectedDirectories) ID() string          { return "expected-directories" }
func (r *ExpectedDirectories) Description() string { return "Expected directories exist for referenced resources" }

func (r *ExpectedDirectories) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	// If permissions.toml references boundary files, permissions/ dir should exist
	if ctx.App.Permissions != nil && hasBoundaryFiles(ctx.App.Permissions) {
		findings = append(findings, r.checkDir(ctx.Dir, "permissions")...)
	}

	// If components are defined, components/ dir should exist
	if len(ctx.App.Components) > 0 {
		findings = append(findings, r.checkDir(ctx.Dir, "components")...)
	}

	// If actions are defined, actions/ dir should exist
	if len(ctx.App.Actions) > 0 {
		findings = append(findings, r.checkDir(ctx.Dir, "actions")...)
	}

	// If inputs has groups, input_groups/ dir is recommended
	if ctx.App.Inputs != nil && len(ctx.App.Inputs.Groups) > 0 {
		findings = append(findings, r.checkDirWarning(ctx.Dir, "input_groups")...)
	}

	return findings
}

func hasBoundaryFiles(perms *appconfig.Permissions) bool {
	for _, role := range []*appconfig.Role{perms.ProvisionRole, perms.DeprovisionRole, perms.MaintenanceRole} {
		if role != nil && role.PermissionsBoundary != "" && isLocalPath(role.PermissionsBoundary) {
			return true
		}
	}
	return false
}

func (r *ExpectedDirectories) checkDir(appDir, name string) []rule.Finding {
	dirPath := filepath.Join(appDir, name)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return []rule.Finding{{
			RuleID:   r.ID(),
			Severity: rule.SeverityError,
			Message:  fmt.Sprintf("directory %q is expected but does not exist", name),
		}}
	}
	return nil
}

func (r *ExpectedDirectories) checkDirWarning(appDir, name string) []rule.Finding {
	dirPath := filepath.Join(appDir, name)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return []rule.Finding{{
			RuleID:   r.ID(),
			Severity: rule.SeverityWarning,
			Message:  fmt.Sprintf("directory %q is recommended but does not exist", name),
		}}
	}
	return nil
}
