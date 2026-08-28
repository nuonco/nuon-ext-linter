package nuonconfig

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
)

// Load parses a Nuon app config directory using the same pipeline as nuon apps sync.
func Load(ctx context.Context, dir string) (*config.AppConfig, error) {
	return parse.ParseDir(ctx, parse.ParseConfig{
		Dirname: dir,
		V:       validator.New(),
		FileProcessor: func(_ string, obj map[string]any) map[string]any {
			return obj
		},
	})
}
