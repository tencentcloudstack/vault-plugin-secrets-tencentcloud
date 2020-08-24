package vault_plugin_secrets_tencentcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	allowPolicyVersion = "2.0"
	maxTTL             = 43200 * time.Second
)

const (
	userCredential        = "cam_user"
	assumedRoleCredential = "assumed_role"
)

var credentialTypes = []interface{}{
	userCredential,
	assumedRoleCredential,
}

func (b *backend) pathListRoles() *framework.Path {
	return &framework.Path{
		Pattern: "roles/?$",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.operationRolesList,
				Summary:  pathListRolesHelpSyn,
			},
		},
		HelpSynopsis:    pathListRolesHelpSyn,
		HelpDescription: pathListRolesHelpDesc,
	}
}

func (b *backend) pathRole() *framework.Path {
	return &framework.Path{
		Pattern: "roles/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"credential_type": {
				Type:          framework.TypeString,
				Required:      true,
				Description:   fmt.Sprintf("The credential type, allow value contains: %q and %q", userCredential, assumedRoleCredential),
				AllowedValues: credentialTypes,
			},
			"name": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "The name of the role.",
			},
			"role_arn": {
				Type: framework.TypeString,
				Description: `The resource description of the role.
Normal role example:
qcs::cam::uin/12345678:role/4611686018427397919
qcs::cam::uin/12345678:roleName/testRoleName

Service role example:
qcs::cam::uin/12345678:role/tencentcloudServiceRole/4611686018427397920
qcs::cam::uin/12345678:role/tencentcloudServiceRoleName/testServiceRoleName

Note: this only effective when credential_type is user.
`,
			},
			"policies": {
				Type:        framework.TypeString,
				Description: "JSON Policy description. The description rule can be found in https://intl.cloud.tencent.com/document/product/598/10604 and https://intl.cloud.tencent.com/document/product/598/10603",
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Default:     7200,
				Description: fmt.Sprintf("Duration in seconds after which the issued token should expire. Default is 7200, if credential_type is %q, the max is 43200.", assumedRoleCredential),
			},
		},
		ExistenceCheck: b.operationRoleExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.operationRoleCreateUpdate,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.operationRoleRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.operationRoleCreateUpdate,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.operationRoleDelete,
			},
		},
		HelpSynopsis:    pathRolesHelpSyn,
		HelpDescription: pathRolesHelpDesc,
	}
}

func (b *backend) operationRoleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	entry, err := readRole(ctx, req.Storage, data.Get("name").(string))
	if err != nil {
		return false, err
	}

	return entry != nil, nil
}

func (b *backend) operationRoleCreateUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}

	if role == nil && req.Operation == logical.UpdateOperation {
		return nil, fmt.Errorf("no role found to update for %s", roleName)
	} else if role == nil {
		role = new(roleEntry)
	}

	role.CredentialType = data.Get("credential_type").(string)
	role.Name = data.Get("name").(string)

	switch role.CredentialType {
	case assumedRoleCredential:
		if raw, ok := data.GetOk("role_arn"); ok {
			role.RoleARN = raw.(string)
		} else {
			return nil, fmt.Errorf("role_arn can not be empty when credential_type is %s", userCredential)
		}

	case userCredential:
		if raw, ok := data.GetOk("policies"); ok {
			if role.Policy == nil {
				role.Policy = new(sdk.Policy)
			}

			if err := json.Unmarshal([]byte(raw.(string)), role.Policy); err != nil {
				return nil, err
			}
		}

		if role.Policy.Version != allowPolicyVersion {
			return nil, fmt.Errorf("allow policy version is %s, not %s", allowPolicyVersion, role.Policy.Version)
		}

	default:
		credTypes := make([]string, 0, 2)

		for _, credType := range credentialTypes {
			credTypes = append(credTypes, credType.(string))
		}

		allowCredTypes := strings.Join(credTypes, ", ")

		return nil, fmt.Errorf("unknown credential_type %s, allow credential_type contains [%s]", role.CredentialType, allowCredTypes)
	}

	role.TTL = time.Duration(data.Get("ttl").(int)) * time.Second

	if role.TTL > maxTTL {
		return nil, errors.New("allow max ttl is 43200")
	}

	entry, err := logical.StorageEntryJSON("roles/"+roleName, role)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// Let's create a response that we're only going to return if there are warnings.
	resp := new(logical.Response)

	if role.TTL > b.System().MaxLeaseTTL() {
		resp.AddWarning(fmt.Sprintf("ttl of %v exceeds the system max ttl of %v, the latter will be used during login", role.TTL, b.System().MaxLeaseTTL()))
	}

	if len(resp.Warnings) > 0 {
		return resp, nil
	}

	// No warnings, let's return a 204.
	return nil, nil
}

func (b *backend) operationRoleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	var resp *logical.Response

	switch role.CredentialType {
	case assumedRoleCredential:
		resp = &logical.Response{
			Data: map[string]interface{}{
				"name":            role.Name,
				"credential_type": role.CredentialType,
				"role_arn":        role.RoleARN,
				"ttl":             role.TTL / time.Second,
			},
		}

	case userCredential:
		resp = &logical.Response{
			Data: map[string]interface{}{
				"name":            role.Name,
				"credential_type": role.CredentialType,
				"ttl":             role.TTL / time.Second,
				"policies":        role.Policy,
			},
		}
	}

	return resp, nil
}

func (b *backend) operationRoleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return nil, req.Storage.Delete(ctx, "roles/"+data.Get("name").(string))
}

func (b *backend) operationRolesList(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, "roles/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

func readRole(ctx context.Context, s logical.Storage, roleName string) (*roleEntry, error) {
	role, err := s.Get(ctx, "roles/"+roleName)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	result := new(roleEntry)

	if err := role.DecodeJSON(result); err != nil {
		return nil, err
	}

	return result, nil
}

type roleEntry struct {
	Name           string        `json:"name"`
	CredentialType string        `json:"credential_type"`
	RoleARN        string        `json:"role_arn,omitempty"`
	Policy         *sdk.Policy   `json:"policies,omitempty"`
	TTL            time.Duration `json:"ttl"`
}

const pathListRolesHelpSyn = "List the existing roles in this backend."

const pathListRolesHelpDesc = "Roles will be listed by the role name."

const pathRolesHelpSyn = `
Read, write and reference policies and role arn that user API keys or STS AssumeRole credentials can be made for.
`

const pathRolesHelpDesc = `
This path allows you to read and write roles that are used to
create user API keys or STS AssumeRole credentials.

If you supply a role ARN, that role must have been created to allow trusted actors,
and the access key and secret that will be used to call STS AssumeRole (configured at
the /config path) must qualify as a trusted actor.

If you supply policies, a user and API key will be dynamically created. The policies
will be applied to that user.

To obtain a user API key or STS AssumeRole credential after the role is created, if the
backend is mounted at "tencentcloud" and you create a role at "tencentcloud/roles/my-role",
then a user could request access credentials at "tencentcloud/creds/my-role".

To validate the keys, attempt to read an access key after writing the policy.
`
