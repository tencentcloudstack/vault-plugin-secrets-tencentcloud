package tencentcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	rolePath = "role/"
)
const (
	roleTypeUnknown roleType = iota
	roleTypeCAM
	roleTypeSTS
)

type roleType int

// String
func (t roleType) String() string {
	switch t {
	case roleTypeCAM:
		return "cam"
	case roleTypeSTS:
		return "sts"
	}
	return "unknown"
}

type roleEntry struct {
	RoleARN        string          `json:"role_arn"`
	RemotePolicies []*remotePolicy `json:"remote_policies"`
	InlinePolicies []*inlinePolicy `json:"inline_policies"`
	TTL            time.Duration   `json:"ttl"`
	MaxTTL         time.Duration   `json:"max_ttl"`
}

type inlinePolicy struct {
	UUID           string                 `json:"hash"`
	PolicyDocument map[string]interface{} `json:"policy_document"`
}

type remotePolicy struct {
	PolicyName string `json:"policy_name"`
	Scope      string `json:"scope"`
	PolicyId   uint64 `json:"policy_id"`
}

// Type
func (r *roleEntry) Type() roleType {
	if r.RoleARN != "" {
		return roleTypeSTS
	}
	return roleTypeCAM
}

func parseRoleType(nameOfRoleType string) (roleType, error) {
	switch nameOfRoleType {
	case "cam":
		return roleTypeCAM, nil
	case "sts":
		return roleTypeSTS, nil
	default:
		return roleTypeUnknown, fmt.Errorf("received unknown role type: %s", nameOfRoleType)
	}
}

func pathListRoles(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "role/?$",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.pathRolesList,
			},
		},
		HelpSynopsis:    pathListRolesHelpSyn,
		HelpDescription: pathListRolesHelpDesc,
	}
}

func pathRole(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "role/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "The name of the role.",
			},
			"role_arn": {
				Type: framework.TypeString,
				Description: `ARN of the role to be assumed. If provided, inline_policies and
remote_policies should be blank. At creation time, this role must have configured trusted actors,
and the secret id and key that will be used to assume the role (in /config) must qualify
as a trusted actor`,
			},
			"inline_policies": {
				Type:        framework.TypeString,
				Description: "JSON of policies to be dynamically applied to users of this role.",
			},
			"remote_policies": {
				Type: framework.TypeStringSlice,
				Description: `The name and type of each remote policy to be applied.
Example: "policy_name:QcloudAFCFullAccess,scope:All".`,
			},
			"ttl": {
				Type: framework.TypeDurationSecond,
				Description: `Duration in seconds after which the issued token should expire. Defaults
to 0, in which case the value will fallback to the system/mount defaults.`,
			},
			"max_ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "The maximum allowed lifetime of tokens issued using this role.",
			},
		},
		ExistenceCheck: b.pathRoleExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathRoleWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathRoleWrite,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathRoleRead,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathRoleDelete,
			},
		},
		HelpSynopsis:    pathRolesHelpSyn,
		HelpDescription: pathRolesHelpDesc,
	}
}

func (b *backend) pathRoleExistenceCheck(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (bool, error) {
	entry, err := readRole(ctx, req.Storage, data.Get("name").(string))
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}

func roleInlinePolicies(policyDocsStr string, role *roleEntry) (err error) {
	var policyDocs []map[string]interface{}
	if err = json.Unmarshal([]byte(policyDocsStr), &policyDocs); err != nil {
		return err
	}
	role.InlinePolicies = make([]*inlinePolicy, len(policyDocs))
	for i, policyDoc := range policyDocs {
		uid, err := uuid.GenerateUUID()
		if err != nil {
			return err
		}
		uid = strings.Replace(uid, "-", "", -1)
		role.InlinePolicies[i] = &inlinePolicy{
			UUID:           uid,
			PolicyDocument: policyDoc,
		}
	}
	return nil
}

func roleRemotePolicies(remotePolicies []string, role *roleEntry) (err error) {
	role.RemotePolicies = make([]*remotePolicy, len(remotePolicies))
	for i, strPolicy := range remotePolicies {
		policy := &remotePolicy{}
		kvPairs := strings.Split(strPolicy, ",")
		for _, kvPair := range kvPairs {
			kvFields := strings.Split(kvPair, ":")
			if len(kvFields) != 2 {
				return fmt.Errorf("unable to recognize pair in %s", kvPair)
			}
			switch kvFields[0] {
			case "policy_name":
				policy.PolicyName = kvFields[1]
			case "scope":
				policy.Scope = kvFields[1]
			default:
				return fmt.Errorf("invalid key: %s", kvFields[0])
			}
		}
		if policy.PolicyName == "" {
			return fmt.Errorf("policy name is required in %s", strPolicy)
		}
		if policy.Scope == "" {
			return fmt.Errorf("policy scope is required in %s", strPolicy)
		}
		role.RemotePolicies[i] = policy
	}
	return nil
}

func (b *backend) pathRoleWrite(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)
	if roleName == "" {
		return nil, fmt.Errorf("name is required")
	}
	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil && req.Operation == logical.UpdateOperation {
		return nil, fmt.Errorf("no role found to update for %s", roleName)
	} else if role == nil {
		role = &roleEntry{}
	}
	if raw, ok := data.GetOk("role_arn"); ok {
		role.RoleARN = raw.(string)
	}
	if raw, ok := data.GetOk("inline_policies"); ok {
		policyDocsStr := raw.(string)
		err = roleInlinePolicies(policyDocsStr, role)
		if err != nil {
			return nil, err
		}
	}
	if raw, ok := data.GetOk("remote_policies"); ok {
		remotePolicies := raw.([]string)
		err = roleRemotePolicies(remotePolicies, role)
		if err != nil {
			return nil, err
		}
	}
	if raw, ok := data.GetOk("ttl"); ok {
		role.TTL = time.Duration(raw.(int)) * time.Second
	}
	if raw, ok := data.GetOk("max_ttl"); ok {
		role.MaxTTL = time.Duration(raw.(int)) * time.Second
	}
	if role.MaxTTL > 0 && role.TTL > role.MaxTTL {
		return nil, fmt.Errorf("ttl exceeds max_ttl")
	}
	if role.Type() == roleTypeSTS {
		if len(role.RemotePolicies) > 0 {
			return nil, fmt.Errorf("remote_policies must be blank when an arn is present")
		}
		if len(role.InlinePolicies) > 0 {
			return nil, fmt.Errorf("inline_policies must be blank when an arn is present")
		}
	} else if len(role.InlinePolicies)+len(role.RemotePolicies) == 0 {
		return nil, fmt.Errorf("must include an arn, or at least one of inline_policies or remote_policies")
	}
	err = saveRole(ctx, role, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	resp := &logical.Response{}
	if role.Type() == roleTypeSTS && (role.TTL > 0 || role.MaxTTL > 0) {
		resp.AddWarning("role_arn is set so ttl and max_ttl will " +
			"be ignored because they're not editable on STS tokens")
	}
	if role.TTL > b.System().MaxLeaseTTL() {
		resp.AddWarning(fmt.Sprintf("ttl of %d exceeds the system max ttl of %d, "+
			"the latter will be used during login", role.TTL, b.System().MaxLeaseTTL()))
	}
	if len(resp.Warnings) > 0 {
		return resp, nil
	}
	return nil, nil
}

func (b *backend) pathRolesList(ctx context.Context,
	req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, rolePath)
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(entries), nil
}

func (b *backend) pathRoleRead(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)
	if roleName == "" {
		return nil, fmt.Errorf("name is required")
	}

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}
	return &logical.Response{
		Data: map[string]interface{}{
			"role_arn":        role.RoleARN,
			"remote_policies": role.RemotePolicies,
			"inline_policies": role.InlinePolicies,
			"ttl":             role.TTL / time.Second,
			"max_ttl":         role.MaxTTL / time.Second,
		},
	}, nil
}

func (b *backend) pathRoleDelete(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, rolePath+data.Get("name").(string)); err != nil {
		return nil, err
	}
	return nil, nil
}

func saveRole(ctx context.Context, role *roleEntry, s logical.Storage, roleName string) error {
	entry, err := logical.StorageEntryJSON(rolePath+roleName, role)
	if err != nil {
		return err
	}
	if err = s.Put(ctx, entry); err != nil {
		return err
	}
	return nil
}

func readRole(ctx context.Context, s logical.Storage, roleName string) (*roleEntry, error) {
	role, err := s.Get(ctx, rolePath+roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}
	result := &roleEntry{}
	if err := role.DecodeJSON(result); err != nil {
		return nil, err
	}
	return result, nil
}

const pathListRolesHelpSyn = "List the existing roles in this backend."

const pathListRolesHelpDesc = "Roles will be listed by the role name."

const pathRolesHelpSyn = `
Read, write and reference policies and roles that API keys or STS credentials can be made for.
`

const pathRolesHelpDesc = `
This path allows you to read and write roles that are used to
create API keys or STS credentials.

If you supply a role ARN, that role must have been created to allow trusted actors,
and the secret id and key that will be used to call AssumeRole (configured at
the /config path) must qualify as a trusted actor.

If you instead supply inline and/or remote policies to be applied, a user and API
key will be dynamically created. The remote policies will be applied to that user,
and the inline policies will also be dynamically created and applied.

To obtain an API key or STS credential after the role is created, if the
backend is mounted at "tencentcloud" and you create a role at "tencentcloud/roles/deploy",
then a user could request access credentials at "tencentcloud/creds/deploy".

To validate the keys, attempt to read an secret after writing the policy.
`
