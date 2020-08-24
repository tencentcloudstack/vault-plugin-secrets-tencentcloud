package vault_plugin_secrets_tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const secretType = "tencentcloud"

func (b *backend) pathSecrets() *framework.Secret {
	return &framework.Secret{
		Type: secretType,
		Fields: map[string]*framework.FieldSchema{
			"access_key": {
				Type:        framework.TypeString,
				Description: "Access Key",
			},
			"secret_key": {
				Type:        framework.TypeString,
				Description: "Secret Key",
			},
		},
		Renew:  b.operationRenew,
		Revoke: b.operationRevoke,
	}
}

func (b *backend) operationRenew(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	internalData := req.Secret.InternalData

	switch internalData["credential_type"] {
	case userCredential:
		cred, err := readCredential(ctx, req.Storage)
		if err != nil {
			return nil, err
		}

		if cred == nil {
			return nil, errors.New("unable to renew secret key because no credentials are configured")
		}

		accessKey := internalData["access_key"].(string)

		var uin uint64
		if raw, ok := internalData["uin"].(uint64); ok {
			uin = raw
		} else {
			uin = uint64(internalData["uin"].(float64))
		}

		client, err := sdk.NewClient(cred.AccessKey, cred.SecretKey, cred.Region, b.transport)
		if err != nil {
			return nil, err
		}

		if err := client.DeleteAccessKey(uin, accessKey); err != nil {
			return nil, err
		}

		accessKey, secretKey, err := client.CreateAccessKey(uin)
		if err != nil {
			return nil, err
		}

		internalData["access_key"] = accessKey

		resp := b.Secret(secretType).Response(map[string]interface{}{
			"access_key": accessKey,
			"secret_key": secretKey,
		}, internalData)

		var ttl uint64
		if raw, ok := internalData["ttl"].(uint64); ok {
			ttl = raw
		} else {
			ttl = uint64(internalData["ttl"].(float64))
		}

		resp.Secret.TTL = time.Duration(ttl) * time.Second

		return resp, nil

	case assumedRoleCredential:
		return nil, fmt.Errorf("when credential_type is %s, doesn't support renew", assumedRoleCredential)

	default:
		return nil, fmt.Errorf("unsupport credential_type %s", internalData["credential_type"])
	}
}

func (b *backend) operationRevoke(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	internalData := req.Secret.InternalData

	switch internalData["credential_type"] {
	case assumedRoleCredential:
		// assumed role type will revoke after ttl automatically
		return nil, nil

	case userCredential:
		cred, err := readCredential(ctx, req.Storage)
		if err != nil {
			return nil, err
		}

		if cred == nil {
			return nil, errors.New("unable to delete secret key because no credentials are configured")
		}

		roleName := internalData["name"].(string)

		var policyId uint64

		// when run unit test, policy_id is uint64, however when run as plugin, policy_id is float64
		if raw, ok := internalData["policy_id"].(uint64); ok {
			policyId = raw
		} else {
			policyId = uint64(internalData["policy_id"].(float64))
		}

		client, err := sdk.NewClient(cred.AccessKey, cred.SecretKey, cred.Region, b.transport)
		if err != nil {
			return nil, err
		}

		if err := client.DeleteUser(roleName); err != nil {
			return nil, err
		}

		if err := client.DeletePolicy([]uint64{policyId}); err != nil {
			return nil, err
		}

		return nil, nil

	default:
		return nil, fmt.Errorf("unsupport credential_type %s", internalData["credential_type"])
	}
}
