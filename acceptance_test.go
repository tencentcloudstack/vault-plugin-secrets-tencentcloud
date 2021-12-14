package tencentcloud

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

const (
	envVarRunAccTests = "VAULT_ACC"
	envVarSecretId    = "TENCENTCLOUD_SECRET_ID"
	envVarSecretKey   = "TENCENTCLOUD_SECRET_KEY"
	envVarRoleARN     = "TENCENTCLOUD_ROLE_ARN"
)

var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

func TestAcceptanceDynamicPolicyBasedCreds(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add policy-based role", acceptanceTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based creds", acceptanceTestEnv.ReadPolicyBasedCreds)
	t.Run("renew policy-based creds", acceptanceTestEnv.RenewPolicyBasedCreds)
	t.Run("revoke policy-based creds", acceptanceTestEnv.RevokePolicyBasedCreds)
}

func TestAcceptanceDynamicRoleBasedCreds(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add arn-based role", acceptanceTestEnv.AddARNBasedRole)
	t.Run("read arn-based creds", acceptanceTestEnv.ReadARNBasedCreds)
	t.Run("renew arn-based creds", acceptanceTestEnv.RenewARNBasedCreds)
	t.Run("revoke arn-based creds", acceptanceTestEnv.RevokeARNBasedCreds)
}

func newAcceptanceTestEnv() (*testEnv, error) {
	ctx := context.Background()
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: time.Hour,
			MaxLeaseTTLVal:     time.Hour,
		},
	}
	b, err := Factory(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &testEnv{
		SecretId:  os.Getenv(envVarSecretId),
		SecretKey: os.Getenv(envVarSecretKey),
		RoleARN:   os.Getenv(envVarRoleARN),
		Backend:   b,
		Context:   ctx,
		Storage:   &logical.InmemStorage{},
	}, nil
}
