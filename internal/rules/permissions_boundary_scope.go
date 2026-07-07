package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

// PermissionsBoundaryScope checks that permissions boundary JSON files do not
// grant Action:* Resource:* (which provides no effective constraint).
// Corresponds to SEC-002 from the byoc security scanner.
type PermissionsBoundaryScope struct{}

func (r *PermissionsBoundaryScope) ID() string { return "permissions-boundary-scope" }
func (r *PermissionsBoundaryScope) Description() string {
	return "Permissions boundary must not grant Action:* Resource:* (no effective constraint)"
}

func (r *PermissionsBoundaryScope) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	for _, bf := range ctx.App.BoundaryFiles {
		findings = append(findings, r.checkFile(ctx.Dir, bf.Path)...)
	}

	return findings
}

func (r *PermissionsBoundaryScope) checkFile(appDir, path string) []rule.Finding {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var doc struct {
		Statement []struct {
			Effect   string `json:"Effect"`
			Action   any    `json:"Action"`
			Resource any    `json:"Resource"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		return nil
	}

	var findings []rule.Finding
	for _, stmt := range doc.Statement {
		if isWildcard(stmt.Action) && isWildcard(stmt.Resource) {
			relPath, _ := filepath.Rel(appDir, path)
			if relPath == "" {
				relPath = path
			}
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityError,
				Message:  fmt.Sprintf("boundary %q grants Action:* Resource:* — provides no effective constraint", filepath.Base(path)),
				File:     relPath,
			})
		}
	}

	return findings
}
