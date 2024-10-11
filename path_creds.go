package tencentcloud

import (
	"container/list"
	"context"
	"fmt"
	camLocal "github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk/tencentcloud/cam/v20190116"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"math/rand"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/clients"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/jsonutil"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/spf13/cast"
)

const credsPath = "creds/"

const timeLayout = "2006-01-02T15:04:05Z"

func pathCreds(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: credsPath + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "The name of the role.",
			},
			"external_id": {
				Type: framework.TypeString,
				Description: `The external ID is a string of characters that you define for this role. To use this role, a user needs to pass in this external ID as you set.
This improves the security of role assuming by preventing unauthorized use of the role when the role information is leaked or guessed.
You're advised to enable external ID verification if you will allow a third-party platform to use the role to be created, or if the account and role information is easily accessible by other users.`,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathCredsRead,
			},
		},
		HelpSynopsis:    pathCredsHelpSyn,
		HelpDescription: pathCredsHelpDesc,
	}
}

func checkData(roleName string, ctx context.Context, req *logical.Request) (
	role *roleEntry, creds *credConfig, err error) {
	if roleName == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	role, err = readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, nil, err
	}
	if role == nil {
		return nil, nil, fmt.Errorf("role is nil")
	}
	creds, err = readCredConfig(ctx, req.Storage)
	if err != nil {
		return nil, nil, err
	}
	if creds == nil {
		return nil, nil, fmt.Errorf("unable to create secret because no credentials are configured")
	}
	return role, creds, nil
}

func (b *backend) roleTypeSTSFunc(creds *credConfig, req *logical.Request,
	role *roleEntry, roleName, externalId string) (*logical.Response, error) {
	client, err := clients.NewSTSClient(b.profile, creds.SecretId, creds.SecretKey)
	if err != nil {
		return nil, err
	}
	assumeRoleResp, err := client.AssumeRole(generateRoleSessionName(req.DisplayName, roleName), role.RoleARN, externalId)
	if err != nil {
		return nil, err
	}
	expiration, err := time.Parse(timeLayout, *(assumeRoleResp.Response.Expiration))
	if err != nil {
		return nil, err
	}
	resp := b.Secret(secretType).Response(map[string]interface{}{
		"secret_id":  *(assumeRoleResp.Response.Credentials.TmpSecretId),
		"secret_key": *(assumeRoleResp.Response.Credentials.TmpSecretKey),
		"token":      *(assumeRoleResp.Response.Credentials.Token),
		"expiration": expiration,
	}, map[string]interface{}{
		"role_type": roleTypeSTS.String(),
	})
	ttl := expiration.Sub(time.Now())
	resp.Secret.TTL = ttl
	resp.Secret.MaxTTL = ttl
	resp.Secret.Renewable = false
	return resp, nil
}

func addUserFunc(req *logical.Request, roleName string, failList *list.List, client *clients.CAMClient) (
	createUserResp *cam.AddUserResponse, err error) {
	userName := generateUsername(req.DisplayName, roleName)
	failList.PushBack(&addUserFail{&userName})
	createUserResp, err = client.AddUser(userName)
	if err != nil {
		return nil, err
	}
	return createUserResp, nil
}

func inlinePolicyFunc(createUserResp *cam.AddUserResponse,
	role *roleEntry, failList *list.List, client *clients.CAMClient) (inlinePolicies []*remotePolicy, err error) {
	inlinePolicies = make([]*remotePolicy, len(role.InlinePolicies))
	for i, inlinePolicy := range role.InlinePolicies {
		policyName := *createUserResp.Response.Name + "-" + inlinePolicy.UUID
		policyDoc, err := jsonutil.EncodeJSON(inlinePolicy.PolicyDocument)
		if err != nil {
			return nil, err
		}
		createPolicyResp, err := client.CreatePolicy(policyName, string(policyDoc))
		if createPolicyResp != nil && createPolicyResp.Response != nil {
			failList.PushBack(&createPolicyFail{createPolicyResp.Response.PolicyId})
		}
		if err != nil {
			return nil, err
		}
		inlinePolicies[i] = &remotePolicy{
			PolicyId: *(createPolicyResp.Response.PolicyId),
		}
		failList.PushBack(&attachUserPolicyFail{
			createPolicyResp.Response.PolicyId,
			createUserResp.Response.Uin})
		if err := client.AttachUserPolicy(createPolicyResp.Response.PolicyId,
			createUserResp.Response.Uin); err != nil {
			return nil, err
		}
	}
	return inlinePolicies, nil
}

func (b *backend) pathCredsRead(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("name").(string)
	externalId := ""
	if raw, ok := data.GetOk("external_id"); ok {
		externalId = raw.(string)
	}
	role, creds, err := checkData(roleName, ctx, req)
	if err != nil {
		return nil, err
	}
	switch role.Type() {
	case roleTypeSTS:
		return b.roleTypeSTSFunc(creds, req, role, roleName, externalId)
	case roleTypeCAM:
		b.profile.HttpProfile.ReqTimeout = 600
		client, err := clients.NewCAMClient(b.profile, creds.SecretId, creds.SecretKey)
		if err != nil {
			return nil, err
		}
		failList := list.New()
		success := false
		// 5> clean up data
		defer func() {
			// Operation failed, delete data
			if failList.Len() > 0 && !success {
				for failList.Len() > 0 {
					fail := failList.Back()
					deleteForFail(fail.Value, client, b)
					failList.Remove(fail)
				}
			}
		}()
		// 1>AddUser
		createUserResp, err := addUserFunc(req, roleName, failList, client)
		if err != nil {
			return nil, err
		}
		// 2> inlinePolicy
		inlinePolicies, err := inlinePolicyFunc(createUserResp, role, failList, client)
		if err != nil {
			return nil, err
		}
		// 3> remotePol
		for _, remotePol := range role.RemotePolicies {
			policyId, err := getPolicyIdByRemotePol(remotePol, client)
			if err != nil {
				return nil, err
			}
			failList.PushBack(&attachUserPolicyFail{policyId, createUserResp.Response.Uin})
			if err := client.AttachUserPolicy(policyId, createUserResp.Response.Uin); err != nil {
				return nil, err
			}
		}
		// 4> CreateAccessKey
		accessKeyResp, err := client.CreateAccessKey(createUserResp.Response.Uin)
		if accessKeyResp != nil && accessKeyResp.Response != nil {
			failList.PushBack(&createAccessKeyFail{
				accessKeyResp.Response.AccessKey.AccessKeyId,
				createUserResp.Response.Uin})
		}
		if err != nil {
			return nil, err
		}
		resp := b.makeResp(accessKeyResp, createUserResp, inlinePolicies, role, roleName)
		if role.TTL != 0 {
			resp.Secret.TTL = role.TTL
		}
		if role.MaxTTL != 0 {
			resp.Secret.MaxTTL = role.MaxTTL
		}
		success = true
		return resp, nil
	default:
		return nil, fmt.Errorf("unsupported role type: %s", role.Type())
	}
}

func (b *backend) makeResp(accessKeyResp *camLocal.CreateAccessKeyResponse, createUserResp *cam.AddUserResponse,
	inlinePolicies []*remotePolicy, role *roleEntry, roleName string) *logical.Response {
	return b.Secret(secretType).Response(map[string]interface{}{
		"secret_id":  *(accessKeyResp.Response.AccessKey.AccessKeyId),
		"secret_key": *(accessKeyResp.Response.AccessKey.SecretAccessKey),
	}, map[string]interface{}{
		"role_type":       roleTypeCAM.String(),
		"role_name":       roleName,
		"username":        *(createUserResp.Response.Name),
		"uin":             cast.ToString(*(createUserResp.Response.Uin)),
		"secret_id":       *(accessKeyResp.Response.AccessKey.AccessKeyId),
		"inline_policies": inlinePolicies,
		"remote_policies": role.RemotePolicies,
	})
}

func getPolicyIdByRemotePol(remote *remotePolicy, client *clients.CAMClient) (*uint64, error) {
	req, err := client.ListPolicies(remote.PolicyName, remote.Scope)
	if err != nil {
		return nil, err
	}
	list := req.Response.List
	if list == nil || len(list) <= 0 {
		return nil, fmt.Errorf("policy_name:%s,scope:%s no data", remote.PolicyName, remote.Scope)
	}
	policyItem := list[0]
	return policyItem.PolicyId, nil
}

func generateUsername(displayName, roleName string) string {
	return generateName(displayName, roleName, 64)
}

func generateRoleSessionName(displayName, roleName string) string {
	return generateName(displayName, roleName, 32)
}

func generateName(displayName, roleName string, maxLength int) string {
	name := fmt.Sprintf("%s-%s-", displayName, roleName)
	if len(name) > maxLength-15 {
		name = name[:maxLength-15]
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s%d-%d", name, time.Now().Unix(), r.Intn(10000))
}

const pathCredsHelpSyn = `

`

const pathCredsHelpDesc = `
 
`

type addUserFail struct {
	userName *string
}

type createPolicyFail struct {
	PolicyId *uint64
}

type attachUserPolicyFail struct {
	PolicyId *uint64
	Uin      *uint64
}

type createAccessKeyFail struct {
	AccessKeyId *string
	Uin         *uint64
}

// Rollback Data
func deleteForFail(fail interface{}, client *clients.CAMClient, b *backend) {
	switch fail.(type) {
	case *addUserFail:
		f := fail.(*addUserFail)
		if err := client.DeleteUser(f.userName); err != nil {
			if b.Logger().IsError() {
				b.Logger().Error(fmt.Sprintf("unable to delete user %s", *(f.userName)), err)
			}
		}
	case *createPolicyFail:
		f := fail.(*createPolicyFail)
		if err := client.DeletePolicy([]*uint64{f.PolicyId}); err != nil {
			if b.Logger().IsError() {
				b.Logger().Error(fmt.Sprintf("unable to delete policy %d", *(f.PolicyId)), err)
			}
		}
	case *attachUserPolicyFail:
		f := fail.(*attachUserPolicyFail)
		if err := client.DetachUserPolicy(f.PolicyId, f.Uin); err != nil {
			if b.Logger().IsError() {
				b.Logger().Error(fmt.Sprintf(
					"unable to detach policy id:%d,  from user:%d", *(f.PolicyId), *(f.Uin)))
			}
		}
	case *createAccessKeyFail:
		f := fail.(*createAccessKeyFail)
		if err := client.DeleteAccessKey(f.AccessKeyId, f.Uin); err != nil {
			if b.Logger().IsError() {
				b.Logger().Error(fmt.Sprintf(
					"unable to Delete accessKey id:%s,  from user:%d", *(f.AccessKeyId), *(f.Uin)))
			}
		}
	}
}
