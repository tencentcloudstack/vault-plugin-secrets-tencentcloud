package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk/custom"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdkError "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

type Client struct {
	customClient *custom.Client
	camClient    *cam.Client
	stsClient    *sts.Client
}

func NewClient(accessKey, secretKey, region string, transport http.RoundTripper) (*Client, error) {
	credential := common.NewCredential(accessKey, secretKey)
	cpf := profile.NewClientProfile()

	customClient, err := custom.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	customClient.WithHttpTransport(transport)

	camClient, err := cam.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	camClient.WithHttpTransport(transport)

	stsClient, err := sts.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	stsClient.WithHttpTransport(transport)

	return &Client{
		customClient: customClient,
		camClient:    camClient,
		stsClient:    stsClient,
	}, nil
}

func (c *Client) AssumeRole(name, roleArn string, ttl time.Duration) (cred *sts.Credentials, expiration time.Time, err error) {
	request := sts.NewAssumeRoleRequest()
	request.RoleSessionName = &name
	request.RoleArn = &roleArn
	request.DurationSeconds = common.Uint64Ptr(uint64(ttl / time.Second))

	resp, err := c.stsClient.AssumeRole(request)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("assume role failed: %w", err)
	}

	expiration, err = time.Parse("2006-01-02T15:04:05Z", *resp.Response.Expiration)
	if err != nil {
		return nil, time.Time{}, err
	}

	return resp.Response.Credentials, expiration, nil
}

func (c *Client) AddUser(name string) (uin, uid uint64, accessKey, secretKey string, err error) {
	request := cam.NewAddUserRequest()
	request.Name = &name
	request.ConsoleLogin = common.Uint64Ptr(0)
	request.UseApi = common.Uint64Ptr(1)

	resp, err := c.camClient.AddUser(request)
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("create user failed: %w", err)
	}

	return *resp.Response.Uin, *resp.Response.Uid, *resp.Response.SecretId, *resp.Response.SecretKey, nil
}

func (c *Client) CreatePolicy(policyName string, policy *Policy) (policyId uint64, err error) {
	policyJson, err := json.Marshal(policy)
	if err != nil {
		return 0, err
	}

	request := cam.NewCreatePolicyRequest()
	request.PolicyName = &policyName
	request.PolicyDocument = common.StringPtr(string(policyJson))

	resp, err := c.camClient.CreatePolicy(request)
	if err != nil {
		return 0, fmt.Errorf("create policy failed: %w", err)
	}

	return *resp.Response.PolicyId, nil
}

func (c *Client) DeletePolicy(policyIds []uint64) error {
	request := cam.NewDeletePolicyRequest()
	request.PolicyId = make([]*uint64, 0, len(policyIds))
	for _, id := range policyIds {
		request.PolicyId = append(request.PolicyId, common.Uint64Ptr(id))
	}

	if _, err := c.camClient.DeletePolicy(request); err != nil {
		if sdkErr, ok := err.(*sdkError.TencentCloudSDKError); ok {
			if map[string]bool{
				"InvalidParameter.PolicyIdNotExist": true,
				"ResourceNotFound.NotFound":         true,
				"ResourceNotFound.PolicyIdNotFound": true,
			}[sdkErr.Code] {
				return nil
			}
		}

		return err
	}

	return nil
}

func (c *Client) DeleteUser(name string) error {
	request := cam.NewDeleteUserRequest()
	request.Name = &name
	request.Force = common.Uint64Ptr(1)

	if _, err := c.camClient.DeleteUser(request); err != nil {
		if sdkErr, ok := err.(*sdkError.TencentCloudSDKError); ok {
			if sdkErr.Code == "ResourceNotFound.UserNotExist" {
				return nil
			}
		}

		return err
	}

	return nil
}

func (c *Client) AttachUserPolicy(policyId, uin uint64) error {
	request := cam.NewAttachUserPolicyRequest()
	request.PolicyId = &policyId
	request.AttachUin = &uin

	if _, err := c.camClient.AttachUserPolicy(request); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteAccessKey(uin uint64, accessKey string) error {
	request := custom.NewDeleteAccessKeyRequest()
	request.TargetUin = &uin
	request.AccessKeyId = &accessKey

	if _, err := c.customClient.DeleteAccessKey(request); err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateAccessKey(uin uint64) (accessKey, secretKey string, err error) {
	request := custom.NewCreateAccessKeyRequest()
	request.TargetUin = &uin

	resp, err := c.customClient.CreateAccessKey(request)
	if err != nil {
		return "", "", err
	}

	return *resp.Response.AccessKey.AccessKeyId, *resp.Response.AccessKey.SecretAccessKey, nil
}
