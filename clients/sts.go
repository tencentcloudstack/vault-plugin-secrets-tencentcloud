package clients

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

// NewSTSClient
func NewSTSClient(clientProfile *ClientProfile, secretId, secretKey string) (*STSClient, error) {
	creds, err := chainedCreds(secretId, secretKey)
	if err != nil {
		return nil, err
	}
	client, err := sts.NewClient(creds, regions.Ashburn, clientProfile.ClientProfile)
	// proxy serve
	if clientProfile.HttpTransport != nil {
		client.WithHttpTransport(clientProfile.HttpTransport)
	}
	if err != nil {
		return nil, err
	}
	return &STSClient{client: client}, nil
}

// STSClient
type STSClient struct {
	client *sts.Client
}

// AssumeRole
func (c *STSClient) AssumeRole(roleSessionName, roleARN string) (*sts.AssumeRoleResponse, error) {
	assumeRoleReq := sts.NewAssumeRoleRequest()
	assumeRoleReq.RoleSessionName = &roleSessionName
	assumeRoleReq.RoleArn = &roleARN
	return c.client.AssumeRole(assumeRoleReq)
}
