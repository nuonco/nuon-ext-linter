package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/config"
	"github.com/nuonco/nuon-ext-linter/internal/custom"
	"github.com/nuonco/nuon-ext-linter/internal/output"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
	"github.com/nuonco/nuon-ext-linter/internal/rules"
)

func newLintCmd() *cobra.Command {
	var (
		configPath string
		format     string
		severity   string
		ruleFilter string
	)

	cmd := &cobra.Command{
		Use:   "lint [app-config-dir]",
		Short: "Lint a Nuon app config directory",
		Long:  "Run lint rules against a Nuon app config directory to check for best practices and common errors.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("resolving path: %w", err)
			}

			// Load lint config
			cfgPath := configPath
			if cfgPath == "" {
				cfgPath = config.DefaultPath(absDir)
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			// Override severity from flag
			minSeverity := cfg.Settings.MinSeverity
			if severity != "" {
				minSeverity = severity
			}

			// Load app config
			app, err := appconfig.Load(absDir)
			if err != nil {
				return fmt.Errorf("loading app config: %w", err)
			}

			// Detect platform
			platform := detectPlatform(app)

			// Build lint context
			ctx := &rule.LintContext{
				Dir:      absDir,
				App:      app,
				Platform: platform,
			}

			// Register built-in rules
			registry := rule.NewRegistry()
			for _, r := range rules.DefaultRules(cfg) {
				registry.Register(r)
			}

			// Discover and register custom rules
			customSettings := make(map[string]map[string]any)
			for id, rc := range cfg.Rules {
				if strings.HasPrefix(id, "custom:") && rc.Settings != nil {
					customSettings[id] = rc.Settings
				}
			}
			customPath := cfg.Settings.CustomRulesPath
			for _, r := range custom.Discover(customPath, customSettings) {
				registry.Register(r)
			}

			// Filter rules
			var activeRuleFilter map[string]bool
			if ruleFilter != "" {
				activeRuleFilter = make(map[string]bool)
				for _, id := range strings.Split(ruleFilter, ",") {
					activeRuleFilter[strings.TrimSpace(id)] = true
				}
			}

			// Run rules and collect findings
			var findings []rule.Finding
			for _, r := range registry.All() {
				// Check if rule is enabled in config
				if rc, ok := cfg.Rules[r.ID()]; ok && !rc.IsEnabled() {
					continue
				}

				// Apply --rule filter
				if activeRuleFilter != nil && !activeRuleFilter[r.ID()] {
					continue
				}

				findings = append(findings, r.Run(ctx)...)
			}

			// Filter by severity
			minSev := rule.ParseSeverity(minSeverity)
			var filtered []rule.Finding
			for _, f := range findings {
				if f.Severity >= minSev {
					filtered = append(filtered, f)
				}
			}

			// Output
			switch format {
			case "json":
				if err := output.PrintJSON(os.Stdout, filtered); err != nil {
					return err
				}
			default:
				output.PrintText(os.Stdout, filtered)
			}

			// Exit code
			for _, f := range filtered {
				if f.Severity == rule.SeverityError {
					os.Exit(1)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to lint.toml")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "output format: text, json")
	cmd.Flags().StringVarP(&severity, "severity", "s", "", "minimum severity: info, warning, error")
	cmd.Flags().StringVarP(&ruleFilter, "rule", "r", "", "run only specific rule(s), comma-separated")

	return cmd
}

func newRulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "List all available lint rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.Config{
				Rules: make(map[string]config.RuleConfig),
			}
			registry := rule.NewRegistry()
			for _, r := range rules.DefaultRules(cfg) {
				registry.Register(r)
			}

			// Also discover custom rules
			for _, r := range custom.Discover("", nil) {
				registry.Register(r)
			}

			fmt.Println("Available rules:")
			fmt.Println()
			for _, r := range registry.All() {
				fmt.Printf("  %-25s %s\n", r.ID(), r.Description())
			}
			return nil
		},
	}
}

func detectPlatform(app *appconfig.AppConfig) string {
	if app.Runner == nil {
		return ""
	}
	rt := strings.ToLower(app.Runner.RunnerType)
	switch {
	case strings.HasPrefix(rt, "aws"):
		return "aws"
	case strings.HasPrefix(rt, "azure"):
		return "azure"
	case strings.HasPrefix(rt, "gcp"), strings.HasPrefix(rt, "google"):
		return "gcp"
	default:
		return ""
	}
}
