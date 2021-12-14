package tencentcloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/clients"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/jsonutil"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/spf13/cast"
)

const secretType = "tencentcloud"

func pathSecrets(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: secretType,
		Fields: map[string]*framework.FieldSchema{
			secretId: {
				Type:        framework.TypeString,
				Description: "Secret Id",
			},
			secretKey: {
				Type:        framework.TypeString,
				Description: "Secret Key",
			},
		},
		Renew:  b.operationRenew,
		Revoke: b.operationRevoke,
	}
}

func (b *backend) operationRenew(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleTypeRaw, ok := req.Secret.InternalData["role_type"]
	if !ok {
		return nil, errors.New("role_type missing from secret")
	}
	nameOfRoleType, ok := roleTypeRaw.(string)
	if !ok {
		return nil, fmt.Errorf("unable to read role_type: %+v", roleTypeRaw)
	}
	rType, err := parseRoleType(nameOfRoleType)
	if err != nil {
		return nil, err
	}

	switch rType {

	case roleTypeSTS:
		// STS already has a lifetime, and we don'nameOfRoleType support renewing it.
		return nil, nil

	case roleTypeCAM:
		roleName, err := getStringValue(req.Secret.InternalData, "role_name")
		if err != nil {
			return nil, err
		}

		role, err := readRole(ctx, req.Storage, roleName)
		if err != nil {
			return nil, err
		}
		if role == nil {
			// The role has been deleted since the secret was issued or last renewed.
			// The user's expectation is probably that the caller won'nameOfRoleType continue being
			// able to perform renewals.
			return nil, fmt.Errorf("role %s has been deleted so no further renewals are allowed", roleName)
		}

		resp := &logical.Response{Secret: req.Secret}
		if role.TTL != 0 {
			resp.Secret.TTL = role.TTL
		}
		if role.MaxTTL != 0 {
			resp.Secret.MaxTTL = role.MaxTTL
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unrecognized role_type: %s", nameOfRoleType)
	}
}

func (b *backend) operationRevoke(ctx context.Context,
	req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	roleTypeRaw, ok := req.Secret.InternalData["role_type"]
	if !ok {
		return nil, errors.New("role_type missing from secret")
	}
	nameOfRoleType, ok := roleTypeRaw.(string)
	if !ok {
		return nil, fmt.Errorf("unable to read role_type: %+v", roleTypeRaw)
	}
	rType, err := parseRoleType(nameOfRoleType)
	if err != nil {
		return nil, err
	}
	switch rType {
	case roleTypeSTS:
		return nil, nil
	case roleTypeCAM:
		creds, err := readCredConfig(ctx, req.Storage)
		if err != nil {
			return nil, err
		}
		if creds == nil {
			return nil, errors.New("unable to delete access key because no credentials are configured")
		}
		client, err := clients.NewCAMClient(b.profile, creds.SecretId, creds.SecretKey)
		if err != nil {
			return nil, err
		}
		userName, err := getStringValue(req.Secret.InternalData, "username")
		if err != nil {
			return nil, err
		}
		uin, err := getStringValue(req.Secret.InternalData, "uin")
		if err != nil {
			return nil, err
		}
		secret_id, err := getStringValue(req.Secret.InternalData, "secret_id")
		if err != nil {
			return nil, err
		}
		apiErrs := &multierror.Error{}
		uinInt := uint64(cast.ToInt64(uin))
		if err := client.DeleteAccessKey(&secret_id, &uinInt); err != nil {
			apiErrs = multierror.Append(apiErrs, err)
		}
		inlinePolicies, err := getRemotePolicies(req.Secret.InternalData, "inline_policies")
		if err != nil {
			return nil, err
		}
		for _, inlinePolicy := range inlinePolicies {
			if err := client.DetachUserPolicy(&(inlinePolicy.PolicyId), &uinInt); err != nil {
				apiErrs = multierror.Append(apiErrs, err)
			}
			if err := client.DeletePolicy([]*uint64{&(inlinePolicy.PolicyId)}); err != nil {
				apiErrs = multierror.Append(apiErrs, err)
			}
		}
		remotePolicies, err := getRemotePolicies(req.Secret.InternalData, "remote_policies")
		if err != nil {
			return nil, err
		}
		for _, remotePolicy := range remotePolicies {
			policyId, err := getPolicyIdByRemotePol(remotePolicy, client)
			if err != nil {
				return nil, err
			}
			if err := client.DetachUserPolicy(policyId, &uinInt); err != nil {
				apiErrs = multierror.Append(apiErrs, err)
			}
		}
		if err := client.DeleteUser(&userName); err != nil {
			apiErrs = multierror.Append(apiErrs, err)
		}
		return nil, apiErrs.ErrorOrNil()

	default:
		return nil, fmt.Errorf("unrecognized role_type: %s", nameOfRoleType)
	}
}

func getStringValue(internalData map[string]interface{}, key string) (string, error) {
	valueRaw, ok := internalData[key]
	if !ok {
		return "", fmt.Errorf("secret is missing %s internal data", key)
	}
	value, ok := valueRaw.(string)
	if !ok {
		return "", fmt.Errorf("secret is missing %s internal data", key)
	}
	return value, nil
}

func getRemotePolicies(internalData map[string]interface{}, key string) ([]*remotePolicy, error) {
	valuesRaw, ok := internalData[key]
	if !ok {
		return nil, fmt.Errorf("secret is missing %s internal data", key)
	}

	valuesJSON, err := jsonutil.EncodeJSON(valuesRaw)
	if err != nil {
		return nil, fmt.Errorf("malformed %s internal data", key)
	}

	policies := []*remotePolicy{}
	if err := jsonutil.DecodeJSON(valuesJSON, &policies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s internal data as remotePolicy", key)
	}
	return policies, nil
}
