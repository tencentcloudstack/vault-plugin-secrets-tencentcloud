
# Tencent Cloud Secrets Engine

The Tencent Cloud secrets engine dynamically generates Tencent Cloud access secret ID/key or tokens based on CAM policies, or Tencent Cloud STS credentials based on CAM roles. This generally makes working with Tencent Cloud easier, since it does not involve clicking in the web UI. The Tencent Cloud access secret ID/key or tokens are time-based and are automatically revoked when the Vault lease expires. STS credentials are short-lived, non-renewable, and expire on their own.

## Installation

### From Sources

If you prefer to build the plugin from sources, clone the GitHub repository locally.

### Build the plugin

Build the secrets engine into a plugin using Go.
```shell
$ go build -o vault/plugins/vault-plugin-secrets-tencentcloud ./cmd/vault-plugin-secrets-tencentcloud/main.go
```

### Configuration

Copy the plugin binary into a location of your choice; this directory must be specified as the [`plugin_directory`](https://www.vaultproject.io/docs/configuration#plugin_directory) in the Vault configuration file:

```hcl
plugin_directory = "vault/plugins"
```

Start a Vault server with this configuration file:

```sh
$ vault server -config=vault/server.hcl
```

Once the server is started, register the plugin in the Vault server's [plugin catalog](https://www.vaultproject.io/docs/internals/plugins#plugin-catalog):

```sh
$ SHA256=$(shasum -a 256 vault/plugins/vault-plugin-secrets-tencentcloud | cut -d ' ' -f1)
$ vault plugin register -sha256=$SHA256 secret vault-plugin-secrets-tencentcloud
$ vault plugin info secret vault-plugin-secrets-tencentcloud
```

You can now enable the TencentCloud secrets plugin:

```sh
$ vault secrets enable -path=tencentcloud vault-plugin-secrets-tencentcloud
Success! Enabled the vault-plugin-secrets-tencentcloud secrets engine at: tencentcloud/
```

## Setup

Most secrets engines must be configured in advance before they can perform their functions. These steps are usually completed by an operator or configuration management tool.

1. [Create a custom policy](https://intl.cloud.tencent.com/document/product/598/35596?lang=en&pg=)
    in Tencent Cloud that will be used for the access key you will give Vault. See "Example
    CAM Policy for Vault".

2. [Create a user](https://intl.cloud.tencent.com/document/product/598/13674) in Tencent Cloud
    with a name like "tc-vault-demo", and directly apply the new custom policy to that user
    in the "User Authorization Policies" section.

3. Create an access key for that user in Tencent Cloud, which is an action available in
    Tencent Cloud UI on the user's page.

4. Configure that access key as the credentials that Vault will use to communicate with
    Tencent Cloud to generate credentials:

    ```shell
    $ vault write tencentcloud/config \
        secret_id="AKIDa0A4h4AXXXXXXXX31jBMGtFLAj14rO" \
        secret_key="HI1TCj25sPhjXXXXXXXXXXXX4ZnmVx95" 
    ```

    Alternatively, the Tencent Cloud secrets engine can pick up credentials set as environment variables,
    or credentials available through instance metadata. Since it checks current credentials on every API call,
    changes in credentials will be picked up almost immediately without a Vault restart.

    If available, we recommend using instance metadata for these credentials as they are the most
    secure option. To do so, simply ensure that the instance upon which Vault is running has sufficient
    privileges, and do not add any config.

   1. Configure a role describing how credentials will be granted.

       To generate access tokens using only policies that have already been created in Tencent Cloud:

       ```shell
       $ vault write tencent/role/policy-based \
           remote_policies='policy_name:ReadOnlyAccess,scope:All' \
           remote_policies='policy_name:QcloudAFCFullAccess,scope:All'
       ```

       To generate access tokens using only policies that will be dynamically created in Tencent Cloud by
       Vault:

       ```shell
        $ vault write tencentcloud/role/policy-based \
           inline_policies=-<<EOF
            [
                    { 
                      "version": "2.0",
                       "statement": [
                         {
                            "effect": "allow",
                            "action": "*",
                            "resource": "*",
                            "condition": {
                                "numeric_equal": {
                                    "qcs:read_only_action": 1
                                }
                           }
                        }
                      ]
                   },
                   {...}        
           ]
       EOF
       ```

       Both `inline_policies` and `remote_policies` may be used together. 
       ```shell
            vault write tencentcloud/role/role-based \
            role_arn="qcs::cam::uin/100021543888:roleName/hastrustedactors"
       ``` 
       Any `role_arn` specified must have added "trusted actors" when it was being created. These
       can only be added at role creation time. Trusted actors are entities that can assume the role.
       Since we will be assuming the role to gain credentials, the `secret_id ` and `secret_key` in
       the config must qualify as a trusted actor.




## Usage

After the secrets engine is configured and a user/machine has a Vault token with
the proper permission, it can generate credentials.

1.  Generate a new access key by reading from the `/creds` endpoint with the name
    of the role:

```shell
    $ vault read tencentcloud/creds/policy-based
    Key                Value
    ---                -----
    lease_id           tencentcloud/creds/policy-based/f3e92392-7d9c-09c8-c921-575d62fe80d8
    lease_duration     768h
    lease_renewable    true
    secret_id          0wNEpMMlzy7szvai
    secret_key         PupkTg8jdmau1cXxYacgE736PJj4cA
```

Retrieving creds for a role using a `role_arn` will carry the additional
fields of `expiration` and `security_token`, like so:

```shell
    $ vault read tencentcloud/creds/role-based
    Key                Value
    ---                -----
    lease_id           tencentcloud/creds/hastrustedactors/lZw7hW3jfscsYKYVvp7m7ERx
    lease_duration     1h59m59s
    lease_renewable    false
    expiration         2021-12-07T09:12:46Z
    secret_id          AKIDHW6K0TXXkZr_XkXXXXXXXXXXXX6ZJQ6khu9danZucl_4HyjHk04UDOw9DSN
    secret_key         yEbUKHizYzTNyaV832P6wnmVU0zmtEyd+TIsvQEBtsM=
    token              eEAZmzBApPoUIgjxgQGxS9SxDZoo298a665df05508487d66cc34068694a84defXaUtsypDE3IZvju0N7u2ZV9i3K8u4zfOMZLth7G8kkuQS2bl7ICpxOQdmSy10m3vkCyh_ktiG0IQL2-zH8i3icZyc71kCl2ojC7BsKZEmQBv2sUAu9VFOP5e5FF21VIQpPnAUGGjNx3Cjj7c-LcA2OU8d8R0dpr1qJpGu-QtV_PX5Fbs2JwD4ZmxTU5RrryA3D9mpBQ3ux4osGAV7bPoJTPeavNEqrgw0_D_CneCHoiM5ybjAIYGJpIRiHrQINVqOU-rWIvmQPwq3Quc17jufZy388WDAOAkJggXKvCuotOrBTmZAPGhpjmsaL3km1gQSIrTcEhxT-rYBANJ0ieMsc2XSfriK4dEwHDfoz5MW6qKrRAycC-hLbR1YipWUDTiCFfsr51fIF1UrJxdQf3CaQ
```

### Example CAM Policy for Vault

While Tencent Cloud credentials can be supplied by environment variables, an explicit
setting in the `tencentcloud/config`, or through instance metadata, the resulting
credentials need sufficient permissions to issue secrets. The necessary permissions
vary based on the ways roles are configured.

This is an example RAM policy that would allow you to create credentials using
any type of role:

```json
{
  "Statement": [
    {
      "Action": [
        "cam:CreateAccessKey",
        "cam:DeleteAccessKey",
        "cam:CreatePolicy",
        "cam:DeletePolicy",
        "cam:AttachPolicyToUser",
        "cam:DetachPolicyFromUser",
        "cam:CreateUser",
        "cam:DeleteUser",
        "sts:AssumeRole"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ],
  "Version": "1"
}
```

However, the policy you use should only allow the actions you actually need
for how your roles are configured.

If any roles are using `inline_policies`, you need the following actions:

- `"cam:CreateAccessKey"`
- `"cam:DeleteAccessKey"`
- `"cam:AttachPolicyToUser"`
- `"cam:DetachPolicyFromUser"`
- `"cam:CreateUser"`
- `"cam:DeleteUser"`

If any roles are using `remote_policies`, you need the following actions:

- All listed for `inline_policies`
- `"cam:CreatePolicy"`
- `"cam:DeletePolicy"`

If any roles are using `RoleArn `, you need the following actions:

- `"sts:AssumeRole"`

## API

The Tencent Cloud secrets engine has a full HTTP API. Please see the
[Tencent Cloud secrets engine API](https://github.com/tencentcloudstack/vault-plugin-secrets-tencentcloud/blob/master/docs/Tencent%20%20Cloud%20Secrets%20Engine%20(API).md) for more
details.




