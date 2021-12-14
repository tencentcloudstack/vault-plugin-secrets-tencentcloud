package tencentcloud

import (
	"context"
	"strings"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/clients"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// Factory
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := newBackend(clients.NewClientProfile())
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func newBackend(profile *clients.ClientProfile) *backend {
	b := new(backend)
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: []*framework.Path{
			pathConfig(b),
			pathRole(b),
			pathListRoles(b),
			pathCreds(b),
		},
		Secrets: []*framework.Secret{
			pathSecrets(b),
		},
		BackendType: logical.TypeLogical,
	}
	b.profile = profile
	return b
}

type backend struct {
	*framework.Backend
	profile *clients.ClientProfile
}

const backendHelp = `
The TencentCloud backend dynamically generates TencentCloud secret for a set of
CAM policies. The TencentCloud secret have a configurable ttl set and
are automatically revoked at the end of the ttl.

After mounting this backend, credentials to generate CAM keys must
be configured and roles must be written using
the "role/" endpoints before any secret id can be generated.
`
