# Vault Plugin: TencentCloud Platform Secrets Backend
This is a backend plugin to be used with Hashicorp Vault. This plugin generates unique, ephemeral user API keys and STS AssumeRole credentials.

## Quick Links
- [Vault Website](https://www.vaultproject.io)
- [Tencentcloud Secrets Docs](https://github.com/tencentcloudstack/vault-plugin-secrets-tencentcloud/blob/master/docs/Tencent%20Cloud%20Secrets%20Engine.md)
- [Vault Github](https://www.github.com/hashicorp/vault)
- [General Announcement List](https://groups.google.com/forum/#!forum/hashicorp-announce)
- [Discussion List](https://groups.google.com/forum/#!forum/vault-tool)

## Usage

This is a [Vault plugin](https://www.vaultproject.io/docs/internals/plugins.html)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works. Otherwise, first read this guide on
how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

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

If you are testing this plugin in an earlier version of Vault or
want to develop, see the next section.

## Developing

If you wish to work on this plugin, you'll first need [Go](https://www.golang.org)
installed on your machine (whichever version is required by Vault).

Make sure Go is properly installed, including setting up a [GOPATH](https://golang.org/doc/code.html#GOPATH).

### Get Plugin
Clone this repository:

```sh

mkdir $GOPATH/src/github.com/hashicorp/vault-plugin-secrets-tencentcloud`
cd $GOPATH/src/github.com/hashicorp/
git clone https://github.com/hashicorp/vault-plugin-secrets-tencentcloud.git

```

(or use `go get github.com/hashicorp/vault-plugin-secrets-tencentcloud` ).

You can then download any required build tools by bootstrapping your environment:

```sh
$ make bootstrap
```

To compile a development version of this plugin, run `make` or `make dev`.
This will put the plugin binary in the `bin` and `$GOPATH/bin` folders. `dev`
mode will only generate the binary for your platform and is faster:

```sh
$ make
$ make dev
```

### Install Plugin in Vault

Put the plugin binary into a location of your choice. This directory
will be specified as the [`plugin_directory`](https://www.vaultproject.io/docs/configuration/index.html#plugin_directory)
in the Vault config used to start the server.

```hcl

plugin_directory = "path/to/plugin/directory"

```

Start a Vault server with this config file:
```sh
$ vault server -config=path/to/config.json ...
```

Once the server is started, register the plugin in the Vault server's [plugin catalog](https://www.vaultproject.io/docs/internals/plugins.html#plugin-catalog):

```sh
$ vault write sys/plugins/catalog/tencentcloudsecrets \
        sha_256="$(shasum -a 256 path/to/plugin/directory/vault-plugin-secrets-tencentcloud | cut -d " " -f1)" \
        command="vault-plugin-secrets-tencentcloud"
```

Any name can be substituted for the plugin name "tencentcloudsecrets". This
name will be referenced in the next step, where we enable the secrets
plugin backend using the tencentcloud secrets plugin:

```sh
$ vault secrets enable --plugin-name='tencentcloudsecrets' --path="tencentcloud" plugin
```

### Tests

This plugin has integration tests and acceptance tests.

The integration tests only test offline without real network API request.

The acceptance tests will send real API requests to TencentCloud.

#### Run the integration tests:

```sh
$ make test
```

#### Run the acceptance tests:

- provide your credentials via `TENCENTCLOUD_SECRET_ID` and `TENCENTCLOUD_SECRET_KEY` environment variables
  and set your CAM role arn via `TENCENTCLOUD_ARN` environment variables

```sh
export TENCENTCLOUD_SECRET_ID=AKID12l4j5ljqatgaljgalg
export TENCENTCLOUD_SECRET_KEY=alkfj23lkraljq5lj532lr32l4
export TENCENTCLOUD_ROLE_ARN=qcs::cam::uin/12345678:roleName/test
```

- Run acceptance tests

```sh
make test-acc 
```

## Other Docs

See up-to-date [docs](https://github.com/tencentcloudstack/vault-plugin-secrets-tencentcloud/blob/master/docs/Tencent%20Cloud%20Secrets%20Engine.md)
and general [API docs](https://github.com/tencentcloudstack/vault-plugin-secrets-tencentcloud/blob/master/docs/Tencent%20%20Cloud%20Secrets%20Engine%20(API).md).