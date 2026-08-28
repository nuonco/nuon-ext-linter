package rules

import (
	"testing"

	"github.com/nuonco/nuon-ext-linter/internal/refcollect"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateRefExists_MissingComponent(t *testing.T) {
	cfg := &config.AppConfig{
		Components: []*config.Component{{
			Name:       "consumer",
			Type:       config.TerraformModuleComponentType,
			SourceFile: "components/consumer.toml",
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "ghost",
				Input: "nuon.components.ghost.outputs.id",
			}},
			TerraformModule: &config.TerraformModuleComponentConfig{
				TerraformVersion: "1.11.3",
			},
		}},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateRefExists{}).Run(ctx)
	require.Len(t, findings, 1)
	assert.Equal(t, "template-ref-exists", findings[0].RuleID)
	assert.Contains(t, findings[0].Message, "ghost")
}

func TestTemplateRefExists_MissingAction(t *testing.T) {
	cfg := &config.AppConfig{
		Actions: []*config.ActionConfig{{
			Name: "runner",
			References: []refs.Ref{{
				Type:  refs.RefTypeActions,
				Name:  "ghost_action",
				Input: "nuon.actions.ghost_action.outputs.token",
			}},
		}},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateRefExists{}).Run(ctx)
	require.Len(t, findings, 1)
	assert.Contains(t, findings[0].Message, "ghost_action")
}

func TestTemplateRefExists_ValidReferences(t *testing.T) {
	cfg := &config.AppConfig{
		Components: []*config.Component{{
			Name: "database",
			Type: config.TerraformModuleComponentType,
		}},
		Actions: []*config.ActionConfig{{
			Name: "healthcheck",
		}, {
			Name: "consumer",
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "database",
				Input: "nuon.components.database.outputs.host",
			}, {
				Type:  refs.RefTypeActions,
				Name:  "healthcheck",
				Input: "nuon.actions.healthcheck.outputs.status",
			}},
		}},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateRefExists{}).Run(ctx)
	assert.Empty(t, findings)
}

func TestTemplateScope_SandboxComponentRef(t *testing.T) {
	cfg := &config.AppConfig{
		Sandbox: &config.AppSandboxConfig{
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "nlb",
				Input: "nuon.components.nlb.outputs.dns_name",
			}},
		},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateScope{}).Run(ctx)
	require.Len(t, findings, 1)
	assert.Equal(t, "template-scope", findings[0].RuleID)
	assert.Contains(t, findings[0].Message, "sandbox")
}

func TestTemplateScope_StackComponentRef(t *testing.T) {
	cfg := &config.AppConfig{
		Stack: &config.StackConfig{
			Name: "my-stack-{{ .nuon.components.nlb.outputs.dns_name }}",
		},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateScope{}).Run(ctx)
	require.NotEmpty(t, findings)
	assert.Equal(t, "template-scope", findings[0].RuleID)
}

func TestTemplateScope_StackInstallIDAllowed(t *testing.T) {
	cfg := &config.AppConfig{
		Stack: &config.StackConfig{
			Name: "stack-{{ .nuon.install.id }}",
		},
		Sandbox: &config.AppSandboxConfig{
			VarsMap: map[string]string{
				"region": "{{ .nuon.install_stack.outputs.region }}",
			},
		},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateScope{}).Run(ctx)
	assert.Empty(t, findings)
}

func TestTemplateSyntax_InvalidTemplate(t *testing.T) {
	cfg := &config.AppConfig{
		Stack: &config.StackConfig{
			Name: "broken-{{ if .nuon.install.id",
		},
	}

	ctx := &rule.LintContext{NuonConfig: cfg, Dir: t.TempDir()}
	findings := (&TemplateSyntax{}).Run(ctx)
	require.Len(t, findings, 1)
	assert.Equal(t, "template-syntax", findings[0].RuleID)
}

func TestRefcollect_Sites(t *testing.T) {
	cfg := &config.AppConfig{
		Components: []*config.Component{{
			Name:       "db",
			SourceFile: "/tmp/components/db.toml",
			References: []refs.Ref{{
				Type: refs.RefTypeInstallStack,
				Name: "region",
			}},
		}},
		Sandbox: &config.AppSandboxConfig{
			References: []refs.Ref{{
				Type: refs.RefTypeInstallInputs,
				Name: "cluster_name",
			}},
		},
	}

	sites, _ := refcollect.Collect(cfg)
	require.NotEmpty(t, sites)
	assert.Equal(t, refcollect.ContextComponent, sites[0].Context)
}
