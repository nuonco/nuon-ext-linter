package rules

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/nuonco/nuon-ext-linter/internal/nuonconfig"
	"github.com/nuonco/nuon-ext-linter/internal/rule"
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

func TestConfigParse_InvalidInputUnknownGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := testdataDir(t, "invalid", "input-unknown-group")
	_, err := nuonconfig.Load(context.Background(), dir)
	require.Error(t, err)
	msg := nuonconfig.FormatLoadError(err)
	assert.Contains(t, msg, "group that does not exist")

	ctx := &rule.LintContext{
		Dir:         dir,
		ParseErrors: []string{msg},
	}
	findings := (&ConfigParse{}).Run(ctx)
	require.Len(t, findings, 1)
	assert.Equal(t, "config-parse", findings[0].RuleID)
	assert.Equal(t, rule.SeverityError, findings[0].Severity)
	assert.Contains(t, findings[0].Message, "does_not_exist")
}
