package vault_plugin_secrets_tencentcloud

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathConfig() *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"access_key": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "Access key with appropriate permissions.",
			},
			"secret_key": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "Secret key with appropriate permissions.",
			},
			"region": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "The region of role and Credentials",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.operationConfigUpdate,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.operationConfigRead,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.operationConfigDelete,
			},
		},
		HelpSynopsis:    pathConfigRootHelpSyn,
		HelpDescription: pathConfigRootHelpDesc,
	}
}

func (b *backend) operationConfigUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Access keys and secrets are generated in pairs. You would never need
	// to update one or the other alone, always both together.

	entry, err := logical.StorageEntryJSON("config", credConfig{
		AccessKey: data.Get("access_key").(string),
		SecretKey: data.Get("secret_key").(string),
		Region:    data.Get("region").(string),
	})
	if err != nil {
		return nil, err
	}

	return nil, req.Storage.Put(ctx, entry)
}

func (b *backend) operationConfigRead(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	cred, err := readCredential(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, nil
	}

	// "secret_key" is intentionally not returned by this endpoint
	return &logical.Response{
		Data: map[string]interface{}{
			"access_key": cred.AccessKey,
			"region":     cred.Region,
		},
	}, nil
}

func (b *backend) operationConfigDelete(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, "config"); err != nil {
		return nil, err
	}
	return nil, nil
}

func readCredential(ctx context.Context, storage logical.Storage) (*credConfig, error) {
	entry, err := storage.Get(ctx, "config")
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	creds := new(credConfig)
	if err := entry.DecodeJSON(creds); err != nil {
		return nil, err
	}

	return creds, nil
}

type credConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

const pathConfigRootHelpSyn = `
Configure the access key and secret to use for Tencentcloud CAM and STS.
`

const pathConfigRootHelpDesc = `
Before doing anything, the TencentCloud backend needs credentials that are able
to manage CAM users and STS AssumeRole. 
This endpoint is used to configure those credentials.
`
