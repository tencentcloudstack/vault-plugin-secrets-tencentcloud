package vault_plugin_secrets_tencentcloud

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

func newIntegrationTestEnv(t *testing.T) (*testEnv, error) {
	id := os.Getenv("TENCENTCLOUD_SECRET_ID")
	if id == "" {
		t.Fatal("miss TENCENTCLOUD_SECRET_ID")
	}

	key := os.Getenv("TENCENTCLOUD_SECRET_KEY")
	if key == "" {
		t.Fatal("miss TENCENTCLOUD_SECRET_KEY")
	}

	arn := os.Getenv("TENCENTCLOUD_ARN")
	if arn == "" {
		t.Fatal("miss TENCENTCLOUD_ARN")
	}

	b := newBackend(true)
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: 7200 * time.Second,
			MaxLeaseTTLVal:     7200 * time.Second,
		},
	}
	if err := b.Setup(context.Background(), conf); err != nil {
		return nil, err
	}

	return &testEnv{
		AccessKey: id,
		SecretKey: key,
		RoleARN:   arn,
		Backend:   b,
		Context:   context.Background(),
		Storage:   &logical.InmemStorage{},
	}, nil
}

func TestConfig(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", integrationTestEnv.AddConfig)
	t.Run("read config", integrationTestEnv.ReadConfig)
	t.Run("update config", integrationTestEnv.UpdateConfig)
	t.Run("read updated config", integrationTestEnv.ReadUpdatedConfig)
	t.Run("delete config", integrationTestEnv.DeleteConfig)
	t.Run("read empty config", integrationTestEnv.ReadEmptyConfig)
}

func TestDynamicPolicyBasedCreds(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", integrationTestEnv.AddConfig)

	t.Run("add policy-based role", integrationTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based role", integrationTestEnv.ReadPolicyBasedRole)
	t.Run("update policy-based role", integrationTestEnv.UpdatePolicyBasedRole)
	t.Run("read updated policy-based role", integrationTestEnv.ReadUpdatedPolicyBasedRole)
	t.Run("delete policy-based role", integrationTestEnv.DeletePolicyBasedRole)

	t.Run("add policy-based role", integrationTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based creds", integrationTestEnv.ReadPolicyBasedCreds)
	t.Run("renew policy-based creds", integrationTestEnv.RenewPolicyBasedCreds)
	t.Run("revoke policy-based creds", integrationTestEnv.RevokePolicyBasedCreds)
}

func TestDynamicRoleBasedCreds(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", integrationTestEnv.AddConfig)

	t.Run("add arn-based role", integrationTestEnv.AddARNBasedRole)
	t.Run("read arn-based role", integrationTestEnv.ReadARNBasedRole)
	t.Run("update arn-based role", integrationTestEnv.UpdateARNBasedRole)
	t.Run("read updated arn-based role", integrationTestEnv.ReadUpdatedARNBasedRole)
	t.Run("delete arn-based role", integrationTestEnv.DeleteARNBasedRole)

	t.Run("add arn-based role", integrationTestEnv.AddARNBasedRole)
	t.Run("read arn-based creds", integrationTestEnv.ReadARNBasedCreds)
	t.Run("renew arn-based creds", integrationTestEnv.RenewARNBasedCreds)
	t.Run("revoke arn-based creds", integrationTestEnv.RevokeARNBasedCreds)
}

func TestMultiRoles(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", integrationTestEnv.AddConfig)

	t.Run("add policy-based role", integrationTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based role", integrationTestEnv.ReadPolicyBasedRole)

	t.Run("add arn-based role", integrationTestEnv.AddARNBasedRole)
	t.Run("read arn-based creds", integrationTestEnv.ReadARNBasedCreds)

	t.Run("list two roles", integrationTestEnv.ListTwoRoles)
	t.Run("delete arn-based role", integrationTestEnv.DeleteARNBasedRole)
	t.Run("list one role", integrationTestEnv.ListOneRole)
}
