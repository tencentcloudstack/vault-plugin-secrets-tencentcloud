package vault_plugin_secrets_tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathCreds() *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "The name of the role.",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.operationCredsRead,
			},
		},
		HelpSynopsis:    pathCredsHelpSyn,
		HelpDescription: pathCredsHelpDesc,
	}
}

func (b *backend) operationCredsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}

	if role == nil {
		// Attempting to read a role that doesn't exist.
		return nil, nil
	}

	cred, err := readCredential(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if cred == nil {
		return nil, errors.New("unable to create secret because no credentials are configured")
	}

	client, err := sdk.NewClient(cred.AccessKey, cred.SecretKey, cred.Region, b.debug)
	if err != nil {
		return nil, err
	}

	var resp *logical.Response

	generateName := generateName(roleName)

	switch role.CredentialType {
	case assumedRoleCredential:
		credentials, expiration, err := client.AssumeRole(generateName, role.RoleARN, role.TTL)
		if err != nil {
			return nil, err
		}

		resp = b.Secret(secretType).Response(map[string]interface{}{
			"name":         generateName,
			"access_key":   credentials.TmpSecretId,
			"secret_key":   credentials.TmpSecretKey,
			"secret_token": credentials.Token,
			"expiration":   expiration,
		}, map[string]interface{}{
			"credential_type": assumedRoleCredential,
		})

		resp.Secret.TTL = time.Until(expiration)
		resp.Secret.Renewable = false

	case userCredential:
		policyId, err := client.CreatePolicy(generateName, role.Policy)
		if err != nil {
			return nil, err
		}

		var success bool

		defer func() {
			if !success {
				if err := client.DeleteUser(generateName); err != nil {
					if b.Logger().IsError() {
						b.Logger().Error(fmt.Sprintf("delete user %s failed: %v", role.Name, err))
					}
				}

				if err := client.DeletePolicy([]uint64{policyId}); err != nil {
					if b.Logger().IsError() {
						b.Logger().Error(fmt.Sprintf("delete policy %d failed: %v", policyId, err))
					}
				}
			}
		}()

		uin, uid, accessKey, secretKey, err := client.AddUser(generateName)
		if err != nil {
			return nil, err
		}

		if err := client.AttachUserPolicy(policyId, uin); err != nil {
			return nil, err
		}

		success = true

		resp = b.Secret(secretType).Response(map[string]interface{}{
			"name":       generateName,
			"access_key": accessKey,
			"secret_key": secretKey,
			"expiration": time.Now().Add(role.TTL).Format("2006-01-02T15:04:05Z"),
		}, map[string]interface{}{
			"name":            generateName,
			"credential_type": userCredential,
			"uin":             uin,
			"uid":             uid,
			"policy_id":       policyId,
			"access_key":      accessKey,
		})

		resp.Secret.TTL = role.TTL

		// we can reset access key and secret key
		resp.Secret.Renewable = true
	}

	return resp, nil
}

func generateName(name string) string {
	const prefix = "vault-token"

	dateStr := time.Now().Format("2006-01-02")

	return fmt.Sprintf("%s-%s-%s-%s", prefix, name, dateStr, randString(8))
}

func randString(n int) string {
	const letters = "0123456789abcdefghigklmopqrstuvwxyz"

	if n < 1 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, n)

	for i := range b {
		b[i] = letters[r.Int63()%35]
	}

	return string(b)
}

const pathCredsHelpSyn = `
Generate a user API key or STS AssumeRole credential using the given role's configuration.'
`

const pathCredsHelpDesc = `
This path will generate a new user API key or STS AssumeRole credential for
accessing TencentCloud. For example, if this backend is mounted at "tencentcloud",
then "tencentcloud/creds/my-role" would generate access keys for the "my-role" role.

The user API key or STS AssumeRole credential will have a ttl associated with it. User API keys can
be renewed or revoked as described here: 
https://www.vaultproject.io/docs/concepts/lease.html,
but STS AssumeRole credentials do not support renewal or revocation, it will be revoked automatically after timeout.
`
