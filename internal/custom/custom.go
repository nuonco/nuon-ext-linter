package custom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

const defaultTimeout = 30 * time.Second

type ExternalRule struct {
	name   string
	path   string
	settings map[string]any
}

func (r *ExternalRule) ID() string          { return "custom:" + r.name }
func (r *ExternalRule) Description() string { return fmt.Sprintf("Custom rule: %s", r.name) }

func (r *ExternalRule) Run(ctx *rule.LintContext) []rule.Finding {
	input := map[string]any{
		"settings": r.settings,
		"platform": ctx.Platform,
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return []rule.Finding{{
			RuleID:   r.ID(),
			Severity: rule.SeverityError,
			Message:  fmt.Sprintf("failed to marshal input: %v", err),
		}}
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, r.path, ctx.Dir)
	cmd.Stdin = bytes.NewReader(inputJSON)

	output, err := cmd.Output()
	if err != nil {
		return []rule.Finding{{
			RuleID:   r.ID(),
			Severity: rule.SeverityError,
			Message:  fmt.Sprintf("custom rule execution failed: %v", err),
		}}
	}

	var result struct {
		Findings []rule.Finding `json:"findings"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return []rule.Finding{{
			RuleID:   r.ID(),
			Severity: rule.SeverityError,
			Message:  fmt.Sprintf("failed to parse custom rule output: %v", err),
		}}
	}

	return result.Findings
}

func Discover(customPath string, settings map[string]map[string]any) []rule.Rule {
	seen := make(map[string]bool)
	var rules []rule.Rule

	// Search custom_rules_path first
	if customPath != "" {
		rules = append(rules, findInDir(customPath, settings, seen)...)
	}

	// Search PATH
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		rules = append(rules, findInDir(dir, settings, seen)...)
	}

	return rules
}

func findInDir(dir string, settings map[string]map[string]any, seen map[string]bool) []rule.Rule {
	matches, err := filepath.Glob(filepath.Join(dir, "nuon-lint-*"))
	if err != nil {
		return nil
	}

	var rules []rule.Rule
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()&0111 == 0 {
			continue // not executable
		}

		name := strings.TrimPrefix(filepath.Base(path), "nuon-lint-")
		if seen[name] {
			continue
		}
		seen[name] = true

		var s map[string]any
		if settings != nil {
			s = settings["custom:"+name]
		}

		rules = append(rules, &ExternalRule{
			name:     name,
			path:     path,
			settings: s,
		})
	}

	return rules
}
