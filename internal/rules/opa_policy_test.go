package rules

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeBlocksCreate_UnconditionalDeny(t *testing.T) {
	rego := `package nuon
deny contains msg if { true }`
	assert.Contains(t, analyzeBlocksCreate(rego), "unconditional deny")
}

func TestAnalyzeBlocksCreate_DenyAllManaged(t *testing.T) {
	rego := `package nuon
deny contains msg if {
  some resource in input.plan.resource_changes
  resource.mode == "managed"
}`
	assert.Contains(t, analyzeBlocksCreate(rego), "denies all managed resources")
}

func TestAnalyzeBlocksCreate_ConditionalDenyOK(t *testing.T) {
	rego := `package nuon
deny contains msg if {
  some resource in input.plan.resource_changes
  resource.type == "aws_kms_key"
  resource.mode == "managed"
}`
	assert.Empty(t, analyzeBlocksCreate(rego))
}

func TestValidateOPAPolicy_WrongPackage(t *testing.T) {
	err := validate.ValidateOPAPolicy(`package other
deny contains msg if { true }`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package nuon")
}

func TestValidateOPAPolicy_Valid(t *testing.T) {
	err := validate.ValidateOPAPolicy(`package nuon
deny contains msg if {
  resource.mode == "managed"
}`)
	require.NoError(t, err)
}
