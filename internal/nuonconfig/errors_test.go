package nuonconfig

import (
	"errors"
	"testing"

	nuoncfg "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/stretchr/testify/assert"
)

func TestFormatLoadError_ErrConfig(t *testing.T) {
	err := nuoncfg.ErrConfig{
		Description: "input bad_field specified a group that does not exist does_not_exist",
		Err:         errors.New("group does_not_exist does not exist"),
	}
	assert.Equal(t, "input bad_field specified a group that does not exist does_not_exist", FormatLoadError(err))
}

func TestFormatLoadError_ParseErr(t *testing.T) {
	err := parse.ParseErr{
		Description: "error parsing config",
		Err: nuoncfg.ErrConfig{
			Description: "input bad_field specified a group that does not exist does_not_exist",
		},
	}
	assert.Equal(t, "input bad_field specified a group that does not exist does_not_exist", FormatLoadError(err))
}
