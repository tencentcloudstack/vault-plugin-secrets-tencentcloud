package vault_plugin_secrets_tencentcloud

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

func newIntegrationTestEnv() (*testEnv, error) {
	b := newBackend(new(fakeTransport))

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
		AccessKey: "AKID11223344556677889900",
		SecretKey: "alkjfaj12lj434lqj25qlajga",
		RoleARN:   "qcs::cam::uin/12345678:roleName/test",
		Backend:   b,
		Context:   context.Background(),
		Storage:   &logical.InmemStorage{},
	}, nil
}

func TestIntegrationConfig(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv()
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

func TestIntegrationCamUserBasedCreds(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv()
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

func TestIntegrationAssumedRoleBasedCreds(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv()
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

func TestIntegrationAssumedMultiRoles(t *testing.T) {
	t.Parallel()

	integrationTestEnv, err := newIntegrationTestEnv()
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
