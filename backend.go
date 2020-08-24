package vault_plugin_secrets_tencentcloud

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type backend struct {
	*framework.Backend

	transport http.RoundTripper
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	debug := conf.Logger.IsDebug()

	if !debug {
		env := strings.ToLower(os.Getenv("VAULT_LOG_LEVEL"))
		debug = env == "trace" || env == "debug"
	}

	b := newBackend(&sdk.LogRoundTripper{
		Debug: debug,
	})

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

// newBackend allows us to pass in the sdkConfig for testing purposes.
func newBackend(transport http.RoundTripper) logical.Backend {
	var b backend

	b.transport = transport

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: []*framework.Path{
			b.pathConfig(),
			b.pathRole(),
			b.pathListRoles(),
			b.pathCreds(),
		},
		Secrets: []*framework.Secret{
			b.pathSecrets(),
		},
		BackendType: logical.TypeLogical,
	}

	return b
}

const backendHelp = `
The TencentCloud backend dynamically generates TencentCloud access keys for a set of
CAM policies. The TencentCloud access keys have a configurable ttl set and are automatically
revoked at the end of the ttl.

After mounting this backend, credentials to generate CAM keys must
be configured and roles must be written using
the "roles/" endpoints before any access keys can be generated.
`
