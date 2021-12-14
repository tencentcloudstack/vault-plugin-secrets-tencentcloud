
# Tencent Cloud Secrets Engine (API)

This is the API documentation for the Vault [Tencent Cloud secrets engine](/docs/secrets/tencentcloud).

This documentation assumes the Tencent Cloud secrets engine is enabled at the `/tencentcloud` path
in Vault. Since it is possible to enable secrets engines at any location, please
update your API calls accordingly.

## Config management

This endpoint configures the root CAM credentials to communicate with Tencent Cloud. Tencent Cloud
will use credentials in the following order:

1. Environment variables
2. A static credential configuration set at this endpoint
3. Instance metadata (recommended)

To use instance metadata, leave the static credential configuration unset.


Please see the Vault [Tencent Cloud secret engine](/docs/secrets/tencentcloud) for
the policies that should be attached to the access key you provide.

| Method | Path               |
| :----- | :----------------- |
| `POST` | `/tencentcloud/config` |
| `GET`  | `/tencentcloud/config` |

### Parameters

- `secret_id` (string, required) - The ID of an secret key with appropriate policies.
- `secret_key` (string, required) - The secret for that key.

### Sample Post Request

```shell-session
$ curl \
    -H "X-Vault-Token: ..." \
    -X POST \
    -d '{"secret_id":" ... ","secret_key":" ..."}' \ 
    http://127.0.0.1:8200/v1/tencentcloud/config
```

### Sample Get Response Data

```json
{
  "access_key": "..."
}
```

## Role management

The `role` endpoint configures how Vault will generate credentials for users of each role.

### Parameters

- `name` (string, required) – Specifies the name of the role to generate credentials against. This is part of the request URL.
- `remote_policies` (string, optional) - The names and types of a pre-existing policies to be applied to the generate access token. Example: "name: ReadOnlyAccess,type:-".
- `inline_policies` (string, optional) - The policy document JSON to be generated and attached to the access token.
- `role_arn` (string, optional) - The ARN of a role that will be assumed to obtain STS credentials. See [Vault Tencent Cloud documentation](/docs/secrets/tencentcloud) regarding trusted actors.
- `ttl` (int, optional) - The duration in seconds after which the issued token should expire. Defaults to 0, in which case the value will fallback to the system/mount defaults.
- `max_ttl` (int, optional) - The maximum allowed lifetime of tokens issued using this role.

| Method   | Path                        |
| :------- | :-------------------------- |
| `GET`    | `/tencentcloud/role`            |
| `POST`   | `/tencentcloud/role/:role_name` |
| `GET`    | `/tencentclod/role/:role_name` |
| `DELETE` | `/tencent/role/:role_name` |

### Sample Post Request

```shell-session
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/v1/tencent/role/my-application
```

### Sample Post Payload Using Policies





```json
{
  "remote_policies": [
    "policy_name:ReadOnlyAccess,scope:All",
    "policy_name:QcloudAFCFullAccess,scope:All"
  ],
  "inline_policies": "[{\"version\":\"2.0\",\"statement\":[{\"action\":\"name/sts:AssumeRole\",\"effect\":\"allow\",\"principal\":{\"qcs\":[\"qcs::cam::uin/1000215438888:root\"]}}]}]"
}
```

### Sample Get Role Response Using Policies

```json
 {
  "request_id": "3d98e59a-8e57-14e8-43c2-8a5ae348cd64",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "inline_policies": [
      {
        "hash": "182ea48f5a55cbc418e73b047494ceee",
        "policy_document": {
          "statement": [
            {
              "action": [
                "api:Describe*"
              ],
              "effect": "allow",
              "resource": "*"
            }
          ],
          "version": "2.0"
        }
      },
      {
        "hash": "7d1ab0493a1530a5568e1ef1580794b2",
        "policy_document": {
          "statement": [
            {
              "action": [
                "cam:*",
                "cloudaudit:LookUpEvents"
              ],
              "effect": "allow",
              "resource": "*"
            }
          ],
          "version": "2.0"
        }
      }
    ],
    "max_ttl": 0,
    "remote_policies": [
      {
        "policy_id": 0,
        "policy_name": "QcloudAFCFullAccess",
        "scope": "All"
      },
      {
        "policy_id": 0,
        "policy_name": "QcloudAFFullAccess",
        "scope": "All"
      },
      {
        "policy_id": 0,
        "policy_name": "QcloudAMEReadOnlyAccess",
        "scope": "All"
      }
    ],
    "role_arn": "",
    "ttl": 0
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Sample Post Payload Using Assume-Role

```json
{
  "role_arn": "qcs::cam::uin/100021543888:roleName/hastrustedactors"
}
```

### Sample Get Role Response Using Assume-Role

```json
{
  "request_id": "b3afa4f6-671b-abb1-0678-644308813153",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "inline_policies": null,
    "max_ttl": 0,
    "remote_policies": null,
    "role_arn": "qcs::cam::uin/100021543888:roleName/hastrustedactors",
    "ttl": 0
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Sample List Roles Response

Performing a `LIST` on the `/tencentcloud/roles` endpoint will list the names of all the roles Vault contains.

```json
["policy-based", "role-based"]
```

## Generate CAM Credentials

This endpoint generates dynamic CAM credentials based on the named role. This
role must be created before queried.

| Method | Path                    |
| :----- | :---------------------- |
| `GET`  | `/tencentcloud/creds/:name` |

### Parameters

- `name` (string, required) – Specifies the name of the role to generate credentials against. This is part of the request URL.

### Sample Request

```shell-session
$ curl \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/tencentcloud/creds/example-role
```

### Sample Response for Roles Using Policies

```json
{
  "request_id": "705c18c6-350b-7e21-a2b6-9cacb1f9459b",
  "lease_id": "tencentcloud/creds/hastrustedactors/2HGDFyWHQEfXi6iHv34wIWTX",
  "renewable": true,
  "lease_duration": 2764800,
  "data": {
    "secret_id": "...",
    "secret_key": "..."
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Sample Response for Roles Using Assume-Role

```json
{
  "request_id": "c92b6dd2-ade4-29e6-31b8-d0ae6b077484",
  "lease_id": "tencentcloud/creds/hastrustedactors/z8cZkm5PgVtrszoysLM77hhN",
  "renewable": false,
  "lease_duration": 7199,
  "data": {
    "expiration": "2021-12-07T09:57:28Z",
    "secret_id": "...",
    "secret_key": "...",
    "token": "..."
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```
