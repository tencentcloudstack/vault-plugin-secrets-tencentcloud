package clients

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

func chainedCreds(secretId, secretKey string) (common.CredentialIface, error) {
	providerChain := []common.Provider{
		common.DefaultEnvProvider(),
		common.DefaultProfileProvider(),
		NewConfigurationCredentialProvider(&Configuration{secretId, secretKey}),
		common.DefaultCvmRoleProvider(),
	}
	return common.NewProviderChain(providerChain).GetCredential()
}

// Configuration
type Configuration struct {
	SecretId  string
	SecretKey string
}

// NewConfigurationCredentialProvider
func NewConfigurationCredentialProvider(configuration *Configuration) common.Provider {
	return &ConfigurationProvider{
		Configuration: configuration,
	}
}

// ConfigurationProvider
type ConfigurationProvider struct {
	Configuration *Configuration
}

// GetCredential
func (p *ConfigurationProvider) GetCredential() (common.CredentialIface, error) {
	if p.Configuration.SecretId != "" && p.Configuration.SecretKey != "" {
		return common.NewCredential(p.Configuration.SecretId, p.Configuration.SecretKey), nil
	} else {
		return nil, ErrNoValidCredentialsFound
	}
}

var (
	ErrNoValidCredentialsFound = fmt.Errorf("no valid credentials were found")
)
