package vault_plugin_secrets_tencentcloud

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/logical"
)

func newAcceptanceTestEnv(t *testing.T) (*testEnv, error) {
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

	b := newBackend(&sdk.LogRoundTripper{Debug: true})
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

func runAcceptanceTest() bool {
	env := strings.ToLower(os.Getenv("VAULT_ACC"))

	return env == "1" || env == "true"
}

func TestAcceptanceConfig(t *testing.T) {
	if !runAcceptanceTest() {
		t.SkipNow()
	}

	t.Parallel()

	acceptanceTestEnv, err := newAcceptanceTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("read config", acceptanceTestEnv.ReadConfig)
	t.Run("update config", acceptanceTestEnv.UpdateConfig)
	t.Run("read updated config", acceptanceTestEnv.ReadUpdatedConfig)
	t.Run("delete config", acceptanceTestEnv.DeleteConfig)
	t.Run("read empty config", acceptanceTestEnv.ReadEmptyConfig)
}

func TestAcceptanceCamUserCreds(t *testing.T) {
	if !runAcceptanceTest() {
		t.SkipNow()
	}

	t.Parallel()

	acceptanceTestEnv, err := newAcceptanceTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)

	t.Run("add policy-based role", acceptanceTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based role", acceptanceTestEnv.ReadPolicyBasedRole)
	t.Run("update policy-based role", acceptanceTestEnv.UpdatePolicyBasedRole)
	t.Run("read updated policy-based role", acceptanceTestEnv.ReadUpdatedPolicyBasedRole)
	t.Run("delete policy-based role", acceptanceTestEnv.DeletePolicyBasedRole)

	t.Run("add policy-based role", acceptanceTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based creds", acceptanceTestEnv.ReadPolicyBasedCreds)
	t.Run("renew policy-based creds", acceptanceTestEnv.RenewPolicyBasedCreds)
	t.Run("revoke policy-based creds", acceptanceTestEnv.RevokePolicyBasedCreds)
}

func TestAcceptanceAssumedRoleBasedCreds(t *testing.T) {
	if !runAcceptanceTest() {
		t.SkipNow()
	}

	t.Parallel()

	acceptanceTestEnv, err := newAcceptanceTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)

	t.Run("add arn-based role", acceptanceTestEnv.AddARNBasedRole)
	t.Run("read arn-based role", acceptanceTestEnv.ReadARNBasedRole)
	t.Run("update arn-based role", acceptanceTestEnv.UpdateARNBasedRole)
	t.Run("read updated arn-based role", acceptanceTestEnv.ReadUpdatedARNBasedRole)
	t.Run("delete arn-based role", acceptanceTestEnv.DeleteARNBasedRole)

	t.Run("add arn-based role", acceptanceTestEnv.AddARNBasedRole)
	t.Run("read arn-based creds", acceptanceTestEnv.ReadARNBasedCreds)
	t.Run("renew arn-based creds", acceptanceTestEnv.RenewARNBasedCreds)
	t.Run("revoke arn-based creds", acceptanceTestEnv.RevokeARNBasedCreds)
}

func TestAcceptanceMultiRoles(t *testing.T) {
	if !runAcceptanceTest() {
		t.SkipNow()
	}

	t.Parallel()

	acceptanceTestEnv, err := newAcceptanceTestEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)

	t.Run("add policy-based role", acceptanceTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based role", acceptanceTestEnv.ReadPolicyBasedRole)

	t.Run("add arn-based role", acceptanceTestEnv.AddARNBasedRole)
	t.Run("read arn-based creds", acceptanceTestEnv.ReadARNBasedCreds)

	t.Run("list two roles", acceptanceTestEnv.ListTwoRoles)
	t.Run("delete arn-based role", acceptanceTestEnv.DeleteARNBasedRole)
	t.Run("list one role", acceptanceTestEnv.ListOneRole)
}
