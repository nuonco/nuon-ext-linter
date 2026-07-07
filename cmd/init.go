package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultLintToml = `# Nuon App Config Linter Configuration
# All built-in rules are enabled by default.
# Uncomment and modify to customize.

[settings]
# Additional directory to search for custom rule executables (nuon-lint-*)
# custom_rules_path = "./lint-rules"

# Minimum severity to report: "info", "warning", "error"
# min_severity = "warning"

[rules]

# Require labels on all components, actions, and runbooks
# [rules.require-labels]
# enabled = true
# severity = "warning"
# required_labels = ["team", "env"]

# Disallow admin/owner IAM policies
# [rules.no-admin-permissions]
# enabled = true
# severity = "error"
# extra_blocked_policies = []

# Components should have a nuon.toml in their source directory
# [rules.component-nuon-toml]
# enabled = true
# severity = "warning"

# Expected directories exist for referenced resources
# [rules.expected-directories]
# enabled = true
# severity = "warning"

# Sandbox should use a pinned tag, not a branch
# [rules.sandbox-use-tag]
# enabled = true
# severity = "warning"

# Runner init script should match the platform
# [rules.runner-init-script]
# enabled = true
# severity = "warning"

# Permissions boundary must not grant Action:* Resource:* (SEC-002)
# [rules.permissions-boundary-scope]
# enabled = true
# severity = "error"

# Policy/boundary JSON must not contain service-wide wildcard actions (SEC-003)
# [rules.no-wildcard-actions]
# enabled = true
# severity = "warning"

# Permission roles must have a permissions_boundary (SEC-004)
# [rules.require-permissions-boundary]
# enabled = true
# severity = "warning"

# OPA/Rego policies must have corresponding test files
# [rules.require-policy-tests]
# enabled = true
# severity = "warning"
`

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [app-config-dir]",
		Short: "Generate a starter lint.toml in the app config directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			path := filepath.Join(dir, "lint.toml")
			if _, err := os.Stat(path); err == nil {
				fmt.Fprintf(os.Stderr, "lint.toml already exists at %s\n", path)
				return nil
			}

			if err := os.WriteFile(path, []byte(defaultLintToml), 0644); err != nil {
				return fmt.Errorf("writing lint.toml: %w", err)
			}

			fmt.Printf("Created %s\n", path)
			return nil
		},
	}
}
