package custom

import (
	"encoding/json"

	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

type AssumeRoleRequest struct {
	*tchttp.BaseRequest

	// 角色的资源描述。
	// 普通角色：
	// qcs::cam::uin/12345678:role/4611686018427397919、qcs::cam::uin/12345678:roleName/testRoleName
	// 服务角色：
	// qcs::cam::uin/12345678:role/tencentcloudServiceRole/4611686018427397920、qcs::cam::uin/12345678:role/tencentcloudServiceRoleName/testServiceRoleName
	RoleArn *string `json:"RoleArn,omitempty" name:"RoleArn"`

	// 临时会话名称，由用户自定义名称
	RoleSessionName *string `json:"RoleSessionName,omitempty" name:"RoleSessionName"`

	// 指定临时证书的有效期，单位：秒，默认 7200 秒，最长可设定有效期为 43200 秒
	DurationSeconds *int `json:"DurationSeconds,omitempty" name:"DurationSeconds"`

	// 策略描述
	// 注意：
	// 1、policy 需要做 urlencode（如果通过 GET 方法请求云 API，发送请求前，所有参数都需要按照云 API 规范再 urlencode 一次）。
	// 2、策略语法参照 CAM 策略语法。
	// 3、策略中不能包含 principal 元素。
	Policy *string `json:"Policy,omitempty" name:"Policy"`
}

func (r *AssumeRoleRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *AssumeRoleRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AssumeRoleResponse struct {
	*tchttp.BaseResponse
	Response *struct {
		// 临时安全证书
		Credentials *Credentials `json:"Credentials,omitempty" name:"Credentials"`

		// 证书无效的时间，返回 Unix 时间戳，精确到秒
		ExpiredTime *int `json:"ExpiredTime,omitempty" name:"ExpiredTime"`

		// 证书无效的时间，以 iso8601 格式的 UTC 时间表示
		Expiration *string `json:"Expiration,omitempty" name:"Expiration"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AssumeRoleResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *AssumeRoleResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type Credentials struct {
	// token。token长度和绑定的策略有关，最长不超过4096字节。
	Token *string `json:"Token,omitempty" name:"Token"`

	// 临时证书密钥ID。最长不超过1024字节。
	TmpSecretId *string `json:"TmpSecretId,omitempty" name:"TmpSecretId"`

	// 临时证书密钥Key。最长不超过1024字节。
	TmpSecretKey *string `json:"TmpSecretKey,omitempty" name:"TmpSecretKey"`
}

type ListAccessKeysRequest struct {
	*tchttp.BaseRequest

	// 指定用户Uin，不填默认列出当前用户访问密钥
	TargetUin *int64 `json:"TargetUin,omitempty" name:"TargetUin"`
}

func (r *ListAccessKeysRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *ListAccessKeysRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ListAccessKeysResponse struct {
	*tchttp.BaseResponse

	Response *struct {
		AccessKeys []*AccessKeys `json:"AccessKeys,omitempty" name:"AccessKeys"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ListAccessKeysResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *ListAccessKeysResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type AccessKeys struct {
	// 访问密钥标识
	AccessKeyId *string `json:"AccessKeyId,omitempty" name:"AccessKeyId"`

	// 密钥状态，激活（Active）或未激活（Inactive）
	Status *string `json:"Status,omitempty" name:"Status"`

	// 创建时间
	CreateTime *string `json:"CreateTime,omitempty" name:"CreateTime"`
}

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
