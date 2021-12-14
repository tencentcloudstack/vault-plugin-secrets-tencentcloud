package clients

import (
	camLocal "github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk/tencentcloud/cam/v20190116"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

// NewCAMClient
func NewCAMClient(clientProfile *ClientProfile, secretId, secretKey string) (*CAMClient, error) {
	creds, err := chainedCreds(secretId, secretKey)
	if err != nil {
		return nil, err
	}
	client, err := cam.NewClient(creds, regions.Ashburn, clientProfile.ClientProfile)
	if err != nil {
		return nil, err
	}
	clientLocal, err := camLocal.NewClient(creds, regions.Ashburn, clientProfile.ClientProfile)
	if err != nil {
		return nil, err
	}
	// proxy server
	if clientProfile.HttpTransport != nil {
		client.WithHttpTransport(clientProfile.HttpTransport)
		clientLocal.WithHttpTransport(clientProfile.HttpTransport)
	}
	return &CAMClient{client: client, clientLocal: clientLocal}, nil
}

// cam client
type CAMClient struct {
	client      *cam.Client
	clientLocal *camLocal.Client
}

// CreateAccessKey
func (c *CAMClient) CreateAccessKey(targetUin *uint64) (*camLocal.CreateAccessKeyResponse, error) {
	req := camLocal.NewCreateAccessKeyRequest()
	req.TargetUin = targetUin
	return c.clientLocal.CreateAccessKey(req)
}

// DeleteAccessKey
func (c *CAMClient) DeleteAccessKey(accessKeyId *string, targetUin *uint64) error {
	req := camLocal.NewDeleteAccessKeyRequest()
	req.AccessKeyId = accessKeyId
	req.TargetUin = targetUin
	_, err := c.clientLocal.DeleteAccessKey(req)
	return err
}

// CreatePolicy
func (c *CAMClient) CreatePolicy(policyName string, policyDocument string) (*cam.CreatePolicyResponse, error) {
	description := "Created by Vault."
	req := cam.NewCreatePolicyRequest()
	req.PolicyName = &policyName
	req.PolicyDocument = &policyDocument
	req.Description = &description
	return c.client.CreatePolicy(req)
}

// DeletePolicy
func (c *CAMClient) DeletePolicy(policyIds []*uint64) error {
	req := cam.NewDeletePolicyRequest()
	req.PolicyId = policyIds
	_, err := c.client.DeletePolicy(req)
	return err
}

// AttachUserPolicy
func (c *CAMClient) AttachUserPolicy(policyId *uint64, attachUin *uint64) error {
	req := cam.NewAttachUserPolicyRequest()
	req.AttachUin = attachUin
	req.PolicyId = policyId
	_, err := c.client.AttachUserPolicy(req)
	return err
}

// DetachUserPolicy
func (c *CAMClient) DetachUserPolicy(policyId *uint64, detachUin *uint64) error {
	req := cam.NewDetachUserPolicyRequest()
	req.PolicyId = policyId
	req.DetachUin = detachUin
	_, err := c.client.DetachUserPolicy(req)
	return err
}

// AddUser
func (c *CAMClient) AddUser(userName string) (*cam.AddUserResponse, error) {
	req := cam.NewAddUserRequest()
	req.Name = &userName
	return c.client.AddUser(req)
}

// DeleteUser
func (c *CAMClient) DeleteUser(userName *string) error {
	req := cam.NewDeleteUserRequest()
	req.Name = userName
	_, err := c.client.DeleteUser(req)
	return err
}

// ListPolicies
func (c *CAMClient) ListPolicies(keyWord string, scope string) (*cam.ListPoliciesResponse, error) {
	req := cam.NewListPoliciesRequest()
	req.Scope = &scope
	req.Keyword = &keyWord
	return c.client.ListPolicies(req)
}
