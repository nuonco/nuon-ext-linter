package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Settings Settings                `toml:"settings"`
	Rules    map[string]RuleConfig   `toml:"rules"`
}

type Settings struct {
	CustomRulesPath string `toml:"custom_rules_path"`
	MinSeverity     string `toml:"min_severity"`
}

type RuleConfig struct {
	Enabled            *bool             `toml:"enabled"`
	Severity           string            `toml:"severity"`
	RequiredLabels     []string          `toml:"required_labels"`
	ExtraBlockedPolicies []string        `toml:"extra_blocked_policies"`
	Settings           map[string]any    `toml:"settings"`
}

func (rc RuleConfig) IsEnabled() bool {
	if rc.Enabled == nil {
		return true
	}
	return *rc.Enabled
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Settings: Settings{
			MinSeverity: "warning",
		},
		Rules: make(map[string]RuleConfig),
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("lint.toml: %w", err)
	}

	return cfg, nil
}

func DefaultPath(appDir string) string {
	return filepath.Join(appDir, "lint.toml")
}
