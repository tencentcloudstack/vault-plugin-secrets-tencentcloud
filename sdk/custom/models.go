package custom

import (
	"encoding/json"

	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

type DeleteAccessKeyRequest struct {
	*tchttp.BaseRequest

	// 指定用户Uin，不填默认为当前用户删除访问密钥
	TargetUin *uint64 `json:"TargetUin,omitempty" name:"TargetUin"`

	// 指定需要删除的AccessKeyId
	AccessKeyId *string `json:"AccessKeyId,omitempty" name:"AccessKeyId"`
}

func (r *DeleteAccessKeyRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *DeleteAccessKeyRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteAccessKeyResponse struct {
	*tchttp.BaseResponse

	Response *struct {
		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteAccessKeyResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *DeleteAccessKeyResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateAccessKeyRequest struct {
	*tchttp.BaseRequest

	// 指定用户Uin，不填默认为当前用户创建访问密钥
	TargetUin *uint64 `json:"TargetUin,omitempty" name:"TargetUin"`
}

func (r *CreateAccessKeyRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *CreateAccessKeyRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type CreateAccessKeyResponse struct {
	*tchttp.BaseResponse

	Response *struct {
		// 访问密钥
		AccessKey *AccessKeyDetail `json:"AccessKey,omitempty" name:"AccessKey"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateAccessKeyResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *CreateAccessKeyResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AccessKeyDetail struct {
	// 访问密钥标识
	AccessKeyId *string `json:"AccessKeyId,omitempty" name:"AccessKeyId"`

	// 访问密钥（密钥仅创建时可见，请妥善保存）
	SecretAccessKey *string `json:"SecretAccessKey,omitempty" name:"SecretAccessKey"`

	// 密钥状态，激活（Active）或未激活（Inactive）
	Status *string `json:"Status,omitempty" name:"Status"`

	// 创建时间(时间戳)
	CreateTime *string `json:"CreateTime,omitempty" name:"CreateTime"`
}
