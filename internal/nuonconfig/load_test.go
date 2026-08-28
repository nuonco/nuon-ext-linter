package nuonconfig_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/nuonco/nuon-ext-linter/internal/nuonconfig"
	"github.com/nuonco/nuon-ext-linter/internal/refcollect"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
	"github.com/nuonco/nuon-ext-linter/internal/rules"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataDir(t *testing.T, parts ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	base := filepath.Join(filepath.Dir(file), "..", "..", "testdata")
	return filepath.Join(append([]string{base}, parts...)...)
}

func TestLoad_InvalidMissingComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "invalid", "missing-component")
	cfg, err := nuonconfig.Load(context.Background(), dir)
	require.NoError(t, err)

	sites, _ := refcollect.Collect(cfg)
	var ghost []refs.Ref
	for _, s := range sites {
		for _, r := range s.Refs {
			if r.Type == refs.RefTypeComponents && r.Name == "ghost" {
				ghost = append(ghost, r)
			}
		}
	}
	require.NotEmpty(t, ghost)

	ctx := &rule.LintContext{Dir: dir, NuonConfig: cfg}
	findings := (&rules.TemplateRefExists{}).Run(ctx)
	require.NotEmpty(t, findings)
	assert.Equal(t, "template-ref-exists", findings[0].RuleID)
}

func TestLoad_InvalidSandboxComponentRef(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "invalid", "sandbox-component-ref")
	cfg, err := nuonconfig.Load(context.Background(), dir)
	require.NoError(t, err)

	ctx := &rule.LintContext{Dir: dir, NuonConfig: cfg}
	findings := (&rules.TemplateScope{}).Run(ctx)
	require.NotEmpty(t, findings)
	assert.Equal(t, "template-scope", findings[0].RuleID)
}

func TestLoad_InvalidStackSandboxRef(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "invalid", "stack-sandbox-ref")
	cfg, err := nuonconfig.Load(context.Background(), dir)
	require.NoError(t, err)

	ctx := &rule.LintContext{Dir: dir, NuonConfig: cfg}
	findings := (&rules.TemplateScope{}).Run(ctx)
	require.NotEmpty(t, findings)
}

func TestLoad_InvalidOPADenyAllManaged(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "invalid", "opa-deny-all-managed")
	cfg, err := nuonconfig.Load(context.Background(), dir)
	require.NoError(t, err)

	ctx := &rule.LintContext{Dir: dir, NuonConfig: cfg}
	findings := (&rules.OPAPolicyBlocksCreate{}).Run(ctx)
	require.NotEmpty(t, findings)
	assert.Equal(t, "opa-policy-blocks-create", findings[0].RuleID)
}

func TestLoad_InvalidInputUnknownGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "invalid", "input-unknown-group")
	_, err := nuonconfig.Load(context.Background(), dir)
	require.Error(t, err)
	msg := nuonconfig.FormatLoadError(err)
	assert.Contains(t, msg, "group that does not exist")
	assert.Contains(t, msg, "does_not_exist")
}

func TestLoad_ValidMinimal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "valid", "minimal")
	cfg, err := nuonconfig.Load(context.Background(), dir)
	require.NoError(t, err)

	ctx := &rule.LintContext{Dir: dir, NuonConfig: cfg}
	assert.Empty(t, (&rules.TemplateRefExists{}).Run(ctx))
	assert.Empty(t, (&rules.TemplateScope{}).Run(ctx))
}
