package rules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon-ext-linter/internal/appconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

type ComponentNuonToml struct{}

func (r *ComponentNuonToml) ID() string          { return "component-nuon-toml" }
func (r *ComponentNuonToml) Description() string { return "Components should have a nuon.toml in their source directory" }

func (r *ComponentNuonToml) Run(ctx *rule.LintContext) []rule.Finding {
	var findings []rule.Finding

	for _, c := range ctx.App.Components {
		srcDir := r.resolveSourceDir(ctx.Dir, &c)
		if srcDir == "" {
			continue
		}

		nuonToml := filepath.Join(srcDir, "nuon.toml")
		if _, err := os.Stat(nuonToml); os.IsNotExist(err) {
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityWarning,
				Message:  fmt.Sprintf("component %q source directory %q does not contain a nuon.toml", c.Name, srcDir),
				File:     filepath.Base(c.File),
			})
		}
	}

	return findings
}

func (r *ComponentNuonToml) resolveSourceDir(appDir string, c *appconfig.Component) string {
	if c.Source != "" && isLocalPath(c.Source) {
		return filepath.Join(appDir, c.Source)
	}

	if ref := c.PublicRepo; ref != nil && isLocalPath(ref.Directory) {
		return filepath.Join(appDir, ref.Directory)
	}
	if ref := c.ConnectedRepo; ref != nil && isLocalPath(ref.Directory) {
		return filepath.Join(appDir, ref.Directory)
	}

	return ""
}

func isLocalPath(p string) bool {
	if p == "" {
		return false
	}
	return p[0] == '.' || p[0] == '/'
}
