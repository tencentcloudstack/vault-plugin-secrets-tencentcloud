// Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v20190116

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2019-01-16"

type Client struct {
	common.Client
}

// Deprecated
func NewClientWithSecretId(secretId, secretKey, region string) (client *Client, err error) {
	cpf := profile.NewClientProfile()
	client = &Client{}
	client.Init(region).WithSecretId(secretId, secretKey).WithProfile(cpf)
	return
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewCreateAccessKeyRequest() (request *CreateAccessKeyRequest) {
	request = &CreateAccessKeyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cam", APIVersion, "CreateAccessKey")
	return
}

func NewCreateAccessKeyResponse() (response *CreateAccessKeyResponse) {
	response = &CreateAccessKeyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAccessKey
// 为CAM用户创建访问密钥
//
// 可能返回的错误码:
//  FAILEDOPERATION_ACCESSKEY = "FailedOperation.Accesskey"
//  INVALIDPARAMETER_PARAMERROR = "InvalidParameter.ParamError"
//  INVALIDPARAMETER_USERNOTEXIST = "InvalidParameter.UserNotExist"
//  OPERATIONDENIED_ACCESSKEYOVERLIMIT = "OperationDenied.AccessKeyOverLimit"
//  OPERATIONDENIED_SUBUIN = "OperationDenied.SubUin"
//  OPERATIONDENIED_UINNOTMATCH = "OperationDenied.UinNotMatch"
func (c *Client) CreateAccessKey(request *CreateAccessKeyRequest) (response *CreateAccessKeyResponse, err error) {
	if request == nil {
		request = NewCreateAccessKeyRequest()
	}
	response = NewCreateAccessKeyResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteAccessKeyRequest() (request *DeleteAccessKeyRequest) {
	request = &DeleteAccessKeyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cam", APIVersion, "DeleteAccessKey")
	return
}

func NewDeleteAccessKeyResponse() (response *DeleteAccessKeyResponse) {
	response = &DeleteAccessKeyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteAccessKey
// 为CAM用户删除访问密钥
//
// 可能返回的错误码:
//  FAILEDOPERATION_ACCESSKEY = "FailedOperation.Accesskey"
//  INVALIDPARAMETER_PARAMERROR = "InvalidParameter.ParamError"
//  INVALIDPARAMETER_USERNOTEXIST = "InvalidParameter.UserNotExist"
//  OPERATIONDENIED_ACCESSKEYOVERLIMIT = "OperationDenied.AccessKeyOverLimit"
//  OPERATIONDENIED_SUBUIN = "OperationDenied.SubUin"
//  OPERATIONDENIED_UINNOTMATCH = "OperationDenied.UinNotMatch"
func (c *Client) DeleteAccessKey(request *DeleteAccessKeyRequest) (response *DeleteAccessKeyResponse, err error) {
	if request == nil {
		request = NewDeleteAccessKeyRequest()
	}
	response = NewDeleteAccessKeyResponse()
	err = c.Send(request, response)
	return
}
