package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon-ext-linter/internal/refcollect"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/render"
)

var stackDisallowedRefTypes = []refs.RefType{
	refs.RefTypeSandbox,
	refs.RefTypeComponents,
	refs.RefTypeActions,
	refs.RefTypeInstallStack,
}

var sandboxDisallowedRefTypes = []refs.RefType{
	refs.RefTypeComponents,
	refs.RefTypeActions,
}

// TemplateSyntax validates Go template syntax in stack templated fields.
type TemplateSyntax struct{}

func (r *TemplateSyntax) ID() string          { return "template-syntax" }
func (r *TemplateSyntax) Description() string { return "Template strings must parse as valid Go templates" }

func (r *TemplateSyntax) Run(ctx *rule.LintContext) []rule.Finding {
	if ctx.NuonConfig == nil {
		return nil
	}

	_, stackFields := refcollect.Collect(ctx.NuonConfig)
	var findings []rule.Finding

	for _, f := range stackFields {
		if !strings.Contains(f.Value, "{{") {
			continue
		}
		if err := render.ValidateTextTemplate(f.Value); err != nil {
			findings = append(findings, rule.Finding{
				RuleID:   r.ID(),
				Severity: rule.SeverityError,
				Message:  fmt.Sprintf("%s: invalid template: %v", f.Field, err),
				File:     f.File,
			})
		}
	}

	return findings
}

// TemplateRefExists checks parsed component and action references resolve to declared names.
type TemplateRefExists struct{}

func (r *TemplateRefExists) ID() string          { return "template-ref-exists" }
func (r *TemplateRefExists) Description() string { return "Template references must point at declared components and actions" }

func (r *TemplateRefExists) Run(ctx *rule.LintContext) []rule.Finding {
	if ctx.NuonConfig == nil {
		return nil
	}

	componentNames, actionNames := nameSets(ctx.NuonConfig)
	sites, _ := refcollect.Collect(ctx.NuonConfig)

	var findings []rule.Finding
	seen := make(map[string]bool)
	for _, site := range sites {
		file := relFile(ctx.Dir, site.File)
		for _, ref := range site.Refs {
			key := string(ref.Type) + "|" + ref.Name + "|" + ref.Input
			if seen[key] {
				continue
			}
			switch ref.Type {
			case refs.RefTypeComponents:
				if !componentNames[ref.Name] {
					seen[key] = true
					findings = append(findings, rule.Finding{
						RuleID:   r.ID(),
						Severity: rule.SeverityError,
						Message:  fmt.Sprintf("%s references unknown component %q (%s)", site.Entity, ref.Name, ref.Input),
						File:     file,
					})
				}
			case refs.RefTypeActions:
				if !actionNames[ref.Name] {
					seen[key] = true
					findings = append(findings, rule.Finding{
						RuleID:   r.ID(),
						Severity: rule.SeverityError,
						Message:  fmt.Sprintf("%s references unknown action %q (%s)", site.Entity, ref.Name, ref.Input),
						File:     file,
					})
				}
			}
		}
	}

	return findings
}

// TemplateScope enforces which reference types are allowed in stack and sandbox contexts.
type TemplateScope struct{}

func (r *TemplateScope) ID() string          { return "template-scope" }
func (r *TemplateScope) Description() string { return "Stack and sandbox templates must not reference unavailable state" }

func (r *TemplateScope) Run(ctx *rule.LintContext) []rule.Finding {
	if ctx.NuonConfig == nil {
		return nil
	}

	sites, stackFields := refcollect.Collect(ctx.NuonConfig)
	var findings []rule.Finding

	for _, site := range sites {
		if site.Context != refcollect.ContextSandbox {
			continue
		}
		file := relFile(ctx.Dir, site.File)
		for _, ref := range site.Refs {
			if refTypeDisallowed(ref.Type, sandboxDisallowedRefTypes) {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityError,
					Message:  fmt.Sprintf("sandbox references %s, which is not available when rendering sandbox config (%s)", ref.Input, site.Entity),
					File:     file,
				})
			}
		}
	}

	for _, f := range stackFields {
		if !strings.Contains(f.Value, "{{") {
			continue
		}
		if strings.HasPrefix(f.Field, "custom_nested_stacks") && strings.Contains(f.Field, ".parameters.") {
			if err := config.ValidateStackParameterTemplate(f.Value); err != nil {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityError,
					Message:  fmt.Sprintf("%s: %v", f.Field, err),
					File:     f.File,
				})
			}
		}
		for _, ref := range refs.ParseFieldRefs(f.Value) {
			if refTypeDisallowed(ref.Type, stackDisallowedRefTypes) {
				findings = append(findings, rule.Finding{
					RuleID:   r.ID(),
					Severity: rule.SeverityError,
					Message:  fmt.Sprintf("%s references %s, which is not populated when the install stack is generated", f.Field, ref.Input),
					File:     f.File,
				})
			}
		}
	}

	// Also check non-stack-field sites in stack context if we add stack entity refs later.
	_ = sites

	return findings
}

func nameSets(cfg *config.AppConfig) (map[string]bool, map[string]bool) {
	components := make(map[string]bool)
	actions := make(map[string]bool)
	for _, c := range cfg.Components {
		if c != nil {
			components[c.Name] = true
		}
	}
	for _, a := range cfg.Actions {
		if a != nil {
			actions[a.Name] = true
		}
	}
	return components, actions
}

func refTypeDisallowed(t refs.RefType, disallowed []refs.RefType) bool {
	for _, d := range disallowed {
		if t == d {
			return true
		}
	}
	return false
}

func relFile(dir, file string) string {
	if file == "" {
		return ""
	}
	rel, err := filepath.Rel(dir, file)
	if err != nil || strings.HasPrefix(rel, "..") {
		return file
	}
	return rel
}
