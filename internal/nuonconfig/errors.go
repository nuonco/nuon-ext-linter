package nuonconfig

import (
	"errors"

	nuoncfg "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
)

// FormatLoadError returns the most specific message from a nuonconfig.Load error.
func FormatLoadError(err error) string {
	if err == nil {
		return ""
	}

	var cfgErr nuoncfg.ErrConfig
	if errors.As(err, &cfgErr) && cfgErr.Description != "" {
		return cfgErr.Description
	}

	var parseErr parse.ParseErr
	if errors.As(err, &parseErr) {
		if parseErr.Err != nil {
			if detail := FormatLoadError(parseErr.Err); detail != "" {
				return detail
			}
		}
		return parseErr.Description
	}

	return err.Error()
}
