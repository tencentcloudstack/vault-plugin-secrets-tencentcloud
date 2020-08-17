package custom

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const (
	StsApiVersion = "2018-08-13"
	StsApiService = "sts"

	CamApiVersion = "2019-01-16"
	CamApiService = "cam"
)

type Client struct {
	common.Client
}

func NewClient(credential *common.Credential, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewAssumeRoleRequest() (request *AssumeRoleRequest) {
	request = &AssumeRoleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo(StsApiService, StsApiVersion, "AssumeRole")
	return
}

func NewAssumeRoleResponse() (response *AssumeRoleResponse) {
	response = &AssumeRoleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// 申请扮演角色
func (c *Client) AssumeRole(request *AssumeRoleRequest) (response *AssumeRoleResponse, err error) {
	if request == nil {
		request = NewAssumeRoleRequest()
	}
	response = NewAssumeRoleResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteAccessKeyRequest() (request *DeleteAccessKeyRequest) {
	request = &DeleteAccessKeyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo(CamApiService, CamApiVersion, "DeleteAccessKey")
	return
}

func NewDeleteAccessKeyResponse() (response *DeleteAccessKeyResponse) {
	response = &DeleteAccessKeyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// 删除密钥
func (c *Client) DeleteAccessKey(request *DeleteAccessKeyRequest) (response *DeleteAccessKeyResponse, err error) {
	if request == nil {
		request = NewDeleteAccessKeyRequest()
	}
	response = NewDeleteAccessKeyResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAccessKeyRequest() (request *CreateAccessKeyRequest) {
	request = &CreateAccessKeyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo(CamApiService, CamApiVersion, "CreateAccessKey")
	return
}

func NewCreateAccessKeyResponse() (response *CreateAccessKeyResponse) {
	response = &CreateAccessKeyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// 创建密钥
func (c *Client) CreateAccessKey(request *CreateAccessKeyRequest) (response *CreateAccessKeyResponse, err error) {
	if request == nil {
		request = NewCreateAccessKeyRequest()
	}
	response = NewCreateAccessKeyResponse()
	err = c.Send(request, response)
	return
}
