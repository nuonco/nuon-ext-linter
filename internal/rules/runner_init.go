package rules

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

type RunnerInitScript struct{}

func (r *RunnerInitScript) ID() string          { return "runner-init-script" }
func (r *RunnerInitScript) Description() string { return "Runner init script should match the platform" }

func (r *RunnerInitScript) Run(ctx *rule.LintContext) []rule.Finding {
	if ctx.App.Runner == nil {
		return nil
	}

	runner := ctx.App.Runner
	if ctx.Platform == "" {
		return nil
	}

	var findings []rule.Finding

	if runner.InitScriptURL == "" {
		findings = append(findings, rule.Finding{
			RuleID:   r.ID(),
			Severity: rule.SeverityInfo,
			Message:  fmt.Sprintf("runner_type is %q but no init_script_url is set", runner.RunnerType),
			File:     "runner.toml",
		})
		return findings
	}

	url := strings.ToLower(runner.InitScriptURL)

	// Check for platform mismatches
	mismatches := map[string][]string{
		"aws":   {"azure", "gcp", "google"},
		"azure": {"aws", "s3", "gcp", "google"},
		"gcp":   {"aws", "s3", "azure"},
	}

	if wrongPlatforms, ok := mismatches[ctx.Platform]; ok {
		for _, wrong := range wrongPlatforms {
			if strings.Contains(url, wrong) {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityWarning,
					Message:  fmt.Sprintf("runner_type is %q but init_script_url appears to reference %q", runner.RunnerType, wrong),
					File:     "runner.toml",
				})
				break
			}
		}
	}

	return findings
}
