package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

// NoWildcardActions checks that policy and boundary JSON files do not contain
// service-wide wildcard actions (e.g., iam:*, s3:*, kms:*).
// Corresponds to SEC-003 from the byoc security scanner.
type NoWildcardActions struct{}

func (r *NoWildcardActions) ID() string          { return "no-wildcard-actions" }
func (r *NoWildcardActions) Description() string { return "Policy/boundary JSON must not contain service-wide wildcard actions" }

var serviceWildcardRe = regexp.MustCompile(`^[a-zA-Z0-9-]+:\*$`)

func (r *NoWildcardActions) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	allFiles := append(
		copyJSONFiles(ctx.App.BoundaryFiles),
		copyJSONFiles(ctx.App.PolicyFiles)...,
	)

	for _, f := range allFiles {
		findings = append(findings, r.checkFile(ctx.Dir, f.Path)...)
	}

	return findings
}

func copyJSONFiles(files []appconfig.JSONPolicyFile) []appconfig.JSONPolicyFile {
	out := make([]appconfig.JSONPolicyFile, len(files))
	copy(out, files)
	return out
}

func (r *NoWildcardActions) checkFile(appDir, path string) []rule.Finding {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var doc struct {
		Statement []struct {
			Effect string `json:"Effect"`
			Action any    `json:"Action"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		return nil
	}

	var findings []rule.Finding
	relPath, _ := filepath.Rel(appDir, path)
	if relPath == "" {
		relPath = path
	}

	for _, stmt := range doc.Statement {
		wildcards := extractServiceWildcards(stmt.Action)
		for _, wc := range wildcards {
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityWarning,
				Message:  fmt.Sprintf("policy %q contains service-wide wildcard %q — scope to specific actions", filepath.Base(path), wc),
				File:     relPath,
			})
		}
	}

	return findings
}

func extractServiceWildcards(action any) []string {
	var wildcards []string
	switch v := action.(type) {
	case string:
		if serviceWildcardRe.MatchString(v) {
			wildcards = append(wildcards, v)
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && serviceWildcardRe.MatchString(s) {
				wildcards = append(wildcards, s)
			}
		}
	}
	return wildcards
}
