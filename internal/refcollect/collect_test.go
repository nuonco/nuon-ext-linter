package refcollect

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/stretchr/testify/require"
)

func TestCollect_ComponentAndSandbox(t *testing.T) {
	cfg := &config.AppConfig{
		Components: []*config.Component{{
			Name:       "db",
			SourceFile: "components/db.toml",
			References: []refs.Ref{{
				Type:  refs.RefTypeInstallStack,
				Name:  "region",
				Input: "nuon.install_stack.outputs.region",
			}},
		}},
		Sandbox: &config.AppSandboxConfig{
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "nlb",
				Input: "nuon.components.nlb.outputs.dns",
			}},
		},
		Stack: &config.StackConfig{
			Name: "stack-{{ .nuon.install.id }}",
		},
	}

	sites, stackFields := Collect(cfg)
	require.Len(t, sites, 2)
	require.NotEmpty(t, stackFields)
}
