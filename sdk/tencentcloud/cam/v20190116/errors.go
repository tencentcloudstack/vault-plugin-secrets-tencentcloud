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

const (
	// 此产品的特有错误码

	// 操作访问密钥错误。
	FAILEDOPERATION_ACCESSKEY = "FailedOperation.Accesskey"

	// 非法入参。
	INVALIDPARAMETER_PARAMERROR = "InvalidParameter.ParamError"

	// 用户对象不存在。
	INVALIDPARAMETER_USERNOTEXIST = "InvalidParameter.UserNotExist"

	// 每个账号最多支持两个AccessKey。
	OPERATIONDENIED_ACCESSKEYOVERLIMIT = "OperationDenied.AccessKeyOverLimit"

	// 子用户不允许操作主账号密钥。
	OPERATIONDENIED_SUBUIN = "OperationDenied.SubUin"

	// 被操作密钥与账号不匹配。
	OPERATIONDENIED_UINNOTMATCH = "OperationDenied.UinNotMatch"
)
