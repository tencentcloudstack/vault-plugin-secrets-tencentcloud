package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	tencentcloud "github.com/hashicorp/vault-plugin-secrets-tencentcloud"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
)

func main() {
	apiClientMeta := new(api.PluginAPIClientMeta)
	flags := apiClientMeta.FlagSet()
	_ = flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	if err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: tencentcloud.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	}); err != nil {
		logger := hclog.New(new(hclog.LoggerOptions))

		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
