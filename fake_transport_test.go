package vault_plugin_secrets_tencentcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk/custom"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

type fakeTransport struct{}

func (f *fakeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	rawAction := request.Header["X-TC-Action"]
	if len(rawAction) == 0 {
		return nil, errors.New("no X-TC-Action is set")
	}

	action := rawAction[0]

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)

	switch action {
	case "AddUser":
		response := cam.NewAddUserResponse()

		response.Response = &struct {
			Uin *uint64 `json:"Uin,omitempty" name:"Uin"`

			Name *string `json:"Name,omitempty" name:"Name"`

			Password *string `json:"Password,omitempty" name:"Password"`

			SecretId *string `json:"SecretId,omitempty" name:"SecretId"`

			SecretKey *string `json:"SecretKey,omitempty" name:"SecretKey"`

			Uid *uint64 `json:"Uid,omitempty" name:"Uid"`

			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		}{
			Uin:       common.Uint64Ptr(1),
			Name:      common.StringPtr("test"),
			Password:  nil,
			SecretId:  common.StringPtr("alkjfaj12lj434lqj25qlajga"),
			SecretKey: common.StringPtr("AKID11223344556677889900"),
			Uid:       common.Uint64Ptr(1),
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		}

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "CreatePolicy":
		response := cam.NewCreatePolicyResponse()

		response.Response = (*struct {
			PolicyId  *uint64 `json:"PolicyId,omitempty" name:"PolicyId"`
			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		})(&struct {
			PolicyId  *uint64
			RequestId *string
		}{
			PolicyId:  common.Uint64Ptr(1),
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		})

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "DeleteAccessKey":
		response := custom.NewDeleteAccessKeyResponse()

		response.Response = (*struct {
			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		})(&struct {
			RequestId *string
		}{
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		})

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "AttachUserPolicy":
		response := cam.NewAttachUserPolicyResponse()

		response.Response = (*struct {
			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		})(&struct {
			RequestId *string
		}{
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		})

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "CreateAccessKey":
		response := custom.NewCreateAccessKeyResponse()

		response.Response = (*struct {
			AccessKey *custom.AccessKeyDetail `json:"AccessKey,omitempty" name:"AccessKey"`
			RequestId *string                 `json:"RequestId,omitempty" name:"RequestId"`
		})(&struct {
			AccessKey *custom.AccessKeyDetail
			RequestId *string
		}{
			AccessKey: &custom.AccessKeyDetail{
				AccessKeyId:     common.StringPtr("AKID11223344556677889900"),
				SecretAccessKey: common.StringPtr("alkjfaj12lj434lqj25qlajga"),
				Status:          common.StringPtr("Active"),
				CreateTime:      common.StringPtr("2020-08-24T12:42:32Z"),
			},
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		})

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "DeleteUser":
		response := cam.NewDeleteUserResponse()

		response.Response = (*struct {
			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		})(&struct {
			RequestId *string
		}{
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		})

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "DeletePolicy":
		response := cam.NewDeletePolicyResponse()

		response.Response = (*struct {
			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		})(&struct {
			RequestId *string
		}{
			RequestId: common.StringPtr("test-111-2222-33333-444444"),
		})

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	case "AssumeRole":
		response := sts.NewAssumeRoleResponse()

		response.Response = &struct {
			Credentials *sts.Credentials `json:"Credentials,omitempty" name:"Credentials"`

			ExpiredTime *int64 `json:"ExpiredTime,omitempty" name:"ExpiredTime"`

			Expiration *string `json:"Expiration,omitempty" name:"Expiration"`

			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		}{
			Credentials: &sts.Credentials{
				Token:        common.StringPtr("token"),
				TmpSecretId:  common.StringPtr("AKID11223344556677889900"),
				TmpSecretKey: common.StringPtr("alkjfaj12lj434lqj25qlajga"),
			},
			ExpiredTime: nil,
			Expiration:  common.StringPtr("2020-08-24T12:42:32Z"),
			RequestId:   nil,
		}

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)

	default:
		return nil, fmt.Errorf("unknown action %s", action)
	}

	return recorder.Result(), nil
}
