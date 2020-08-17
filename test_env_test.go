package vault_plugin_secrets_tencentcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/logical"
)

// testEnv allows us to reuse the same requests and response-checking for both integration tests that don't hit
// TencentCloud's real API, and for acceptance tests that do hit their real API.
type testEnv struct {
	AccessKey string
	SecretKey string
	RoleARN   string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	MostRecentSecret *logical.Secret
}

func (e *testEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"access_key": e.AccessKey,
			"secret_key": e.SecretKey,
			"region":     "ap-guangzhou",
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["access_key"] != e.AccessKey {
		t.Fatal("expected access_key of " + e.AccessKey)
	}

	if resp.Data["region"] != "ap-guangzhou" {
		t.Fatal("expected region ap-guangzhou")
	}

	if resp.Data["secret_key"] != nil {
		t.Fatal("secret_key should not be returned")
	}
}

func (e *testEnv) UpdateConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"access_key": "foo",
			"secret_key": "bar",
			"region":     "ap-shanghai",
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadUpdatedConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["access_key"] != "foo" {
		t.Fatal("expected access_key of foo")
	}

	if resp.Data["region"] != "ap-shanghai" {
		t.Fatal("expected region ap-shanghai")
	}

	if resp.Data["secret_key"] != nil {
		t.Fatal("secret_key should not be returned")
	}
}

func (e *testEnv) DeleteConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "config",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadEmptyConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) AddPolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "roles/policy-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"credential_type": userCredential,
			"name":            "policy-based",
			"policies": `{     
        "version":"2.0",
        "statement": 
        [ 
             {  
                    "principal":{"qcs":["qcs::cam::uin/1238423:uin/3232523"]}, 
                    "effect":"allow", 
                    "action":["name/cos:PutObject"], 
                    "resource":["qcs::cos:bj:uid/1238423:prefix//1238423/bucketA/*", 
                                        "qcs::cos:gz:uid/1238423:prefix//1238423/bucketB/object2"], 
                     "condition": {"ip_equal":{"qcs:ip":"10.121.2.10/24"}} 
             }, 
            {  
                 "principal":{"qcs":["qcs::cam::uin/1238423:uin/3232523"]}, 
                 "effect":"allow", 
                 "action":"name/cmqqueue:SendMessage", 
                 "resource":"*" 
            } 
     ] 
}`,
			"ttl": 10,
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadPolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/policy-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["credential_type"].(string) != userCredential {
		t.Fatalf("expected credential_type %s, got %s", userCredential, resp.Data["credential_type"].(string))
	}

	if resp.Data["name"].(string) != "policy-based" {
		t.Fatalf("expected name policy-based, got %s", resp.Data["name"].(string))
	}

	if resp.Data["role_arn"] != nil {
		t.Fatalf("expected no role_arn but received %v", resp.Data["role_arn"])
	}

	policy := resp.Data["policies"].(*sdk.Policy)

	wantPolicy := sdk.Policy{
		Version: allowPolicyVersion,
		Statement: []sdk.Statement{
			{
				Principal: map[string][]interface{}{
					"qcs": {"qcs::cam::uin/1238423:uin/3232523"},
				},
				Effect: "allow",
				Action: []string{"name/cos:PutObject"},
				Resource: []string{
					"qcs::cos:bj:uid/1238423:prefix//1238423/bucketA/*",
					"qcs::cos:gz:uid/1238423:prefix//1238423/bucketB/object2",
				},
				Condition: map[string]map[string]string{
					"ip_equal": {
						"qcs:ip": "10.121.2.10/24",
					},
				},
			},

			{
				Principal: map[string][]interface{}{
					"qcs": {"qcs::cam::uin/1238423:uin/3232523"},
				},
				Effect:    "allow",
				Action:    "name/cmqqueue:SendMessage",
				Resource:  "*",
				Condition: nil,
			},
		},
	}

	policyStr, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}

	wantPolicyStr, err := json.Marshal(wantPolicy)
	if err != nil {
		t.Fatal(err)
	}

	if string(policyStr) != string(wantPolicyStr) {
		t.Fatalf("expected policy is %v, got %v", wantPolicy, policy)
	}

	if resp.Data["ttl"].(time.Duration) != 10 {
		t.Fatalf("expected ttl %ds, got %ds", 10, resp.Data["ttl"].(time.Duration))
	}
}

func (e *testEnv) UpdatePolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/policy-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"credential_type": userCredential,
			"name":            "policy-based",
			"policies": `{     
        "version":"2.0",
        "statement": 
        [ 
             {  
                    "principal":{"qcs":["qcs::cam::uin/1238423:uin/3232523"]}, 
                    "effect":"allow", 
                    "action":["name/cos:PutObject"], 
                    "resource":["qcs::cos:bj:uid/1238423:prefix//1238423/bucketA/*", 
                                        "qcs::cos:gz:uid/1238423:prefix//1238423/bucketB/object2"], 
                     "condition": {"ip_equal":{"qcs:ip":"10.121.2.10/16"}} 
             }, 
            {  
                 "principal":{"qcs":["qcs::cam::uin/1238423:uin/3232523"]}, 
                 "effect":"allow", 
                 "action":"name/cmqqueue:SendMessage", 
                 "resource":"*" 
            } 
     ] 
}`,
			"ttl": 10,
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadUpdatedPolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/policy-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["credential_type"].(string) != userCredential {
		t.Fatalf("expected credential_type %s, got %s", userCredential, resp.Data["credential_type"].(string))
	}

	if resp.Data["name"].(string) != "policy-based" {
		t.Fatalf("expected name policy-based, got %s", resp.Data["name"].(string))
	}

	if resp.Data["role_arn"] != nil {
		t.Fatalf("expected no role_arn but received %v", resp.Data["role_arn"])
	}

	policy := resp.Data["policies"].(*sdk.Policy)

	wantPolicy := sdk.Policy{
		Version: allowPolicyVersion,
		Statement: []sdk.Statement{
			{
				Principal: map[string][]interface{}{
					"qcs": {"qcs::cam::uin/1238423:uin/3232523"},
				},
				Effect: "allow",
				Action: []string{"name/cos:PutObject"},
				Resource: []string{
					"qcs::cos:bj:uid/1238423:prefix//1238423/bucketA/*",
					"qcs::cos:gz:uid/1238423:prefix//1238423/bucketB/object2",
				},
				Condition: map[string]map[string]string{
					"ip_equal": {
						"qcs:ip": "10.121.2.10/16",
					},
				},
			},

			{
				Principal: map[string][]interface{}{
					"qcs": {"qcs::cam::uin/1238423:uin/3232523"},
				},
				Effect:    "allow",
				Action:    "name/cmqqueue:SendMessage",
				Resource:  "*",
				Condition: nil,
			},
		},
	}

	policyStr, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}

	wantPolicyStr, err := json.Marshal(wantPolicy)
	if err != nil {
		t.Fatal(err)
	}

	if string(policyStr) != string(wantPolicyStr) {
		t.Fatalf("expected policy is %v, got %v", wantPolicy, policy)
	}

	if resp.Data["ttl"].(time.Duration) != 10 {
		t.Fatalf("expected ttl %ds, got %ds", 10, resp.Data["ttl"].(time.Duration))
	}
}

func (e *testEnv) DeletePolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "roles/policy-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatalf("expected nil response to represent a 204, got %v", resp)
	}
}

func (e *testEnv) AddARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "roles/role-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"credential_type": assumedRoleCredential,
			"name":            "role-based",
			"role_arn":        e.RoleARN,
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatalf("expected nil response to represent a 204, got %v", resp)
	}
}

func (e *testEnv) ReadARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/role-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["credential_type"].(string) != assumedRoleCredential {
		t.Fatalf("expected credential_type %s, got %s", assumedRoleCredential, resp.Data["credential_type"].(string))
	}

	if resp.Data["name"].(string) != "role-based" {
		t.Fatalf("expected name role-based, got %s", resp.Data["name"].(string))
	}

	if resp.Data["role_arn"] != e.RoleARN {
		t.Fatalf("received unexpected role_arn of %s", resp.Data["role_arn"])
	}

	if resp.Data["policies"] != nil {
		t.Fatalf("received unexpected policies of %v", resp.Data["policies"])
	}

	if resp.Data["ttl"].(time.Duration) != 7200 {
		t.Fatalf("expected ttl %ds, got %ds", 7200, resp.Data["ttl"].(time.Duration))
	}
}

func (e *testEnv) UpdateARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/role-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"credential_type": assumedRoleCredential,
			"name":            "role-based",
			"role_arn":        "qcs::cam::uin/123:role/123",
			"ttl":             100,
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadUpdatedARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/role-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["credential_type"].(string) != assumedRoleCredential {
		t.Fatalf("expected credential_type %s, got %s", assumedRoleCredential, resp.Data["credential_type"].(string))
	}

	if resp.Data["name"].(string) != "role-based" {
		t.Fatalf("expected name role-based, got %s", resp.Data["name"].(string))
	}

	if resp.Data["role_arn"] != "qcs::cam::uin/123:role/123" {
		t.Fatalf("expected role_arn %s, got %s", "qcs::cam::uin/123:role/123", resp.Data["role_arn"].(string))
	}

	if resp.Data["policies"] != nil {
		t.Fatalf("received unexpected policies of %v", resp.Data["policies"])
	}

	if resp.Data["ttl"].(time.Duration) != 100 {
		t.Fatalf("expected ttl 100, got %d", resp.Data["ttl"].(time.Duration))
	}
}

func (e *testEnv) ListTwoRoles(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "roles",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	keys := resp.Data["keys"].([]string)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys but received %d", len(keys))
	}

	if keys[0] != "policy-based" {
		t.Fatalf("expectied policy-based role name but received %s", keys[0])
	}

	if keys[1] != "role-based" {
		t.Fatalf("expected role-based role name but received %s", keys[1])
	}
}

func (e *testEnv) DeleteARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "roles/role-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ListOneRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "roles",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	keys := resp.Data["keys"].([]string)
	if len(keys) != 1 {
		t.Fatalf("expected 2 keys but received %d", len(keys))
	}

	if keys[0] != "policy-based" {
		t.Fatalf("expectied policy-based role name but received %s", keys[0])
	}
}

func (e *testEnv) ReadPolicyBasedCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/policy-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["access_key"] == "" {
		t.Fatal("failed to receive access_key")
	}

	if resp.Data["secret_key"] == "" {
		t.Fatal("failed to receive secret_key")
	}

	e.MostRecentSecret = resp.Secret
}

func (e *testEnv) RenewPolicyBasedCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RenewOperation,
		Storage:   e.Storage,
		Secret:    e.MostRecentSecret,
		Data: map[string]interface{}{
			"lease_id": "foo",
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	e.MostRecentSecret = resp.Secret
}

func (e *testEnv) RevokePolicyBasedCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RevokeOperation,
		Storage:   e.Storage,
		Secret:    e.MostRecentSecret,
		Data: map[string]interface{}{
			"lease_id": "foo",
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadARNBasedCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/role-based",
		Storage:   e.Storage,
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["access_key"] == "" {
		t.Fatal("received blank access_key")
	}

	if resp.Data["secret_key"] == "" {
		t.Fatal("received blank secret_key")
	}

	if fmt.Sprintf("%s", resp.Data["expiration"]) == "" {
		t.Fatal("received blank expiration")
	}

	if resp.Data["security_token"] == "" {
		t.Fatal("received blank security_token")
	}

	e.MostRecentSecret = resp.Secret
}

func (e *testEnv) RenewARNBasedCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RenewOperation,
		Storage:   e.Storage,
		Secret:    e.MostRecentSecret,
		Data: map[string]interface{}{
			"lease_id": "foo",
		},
	}

	wantErr := fmt.Errorf("when credential_type is %s, doesn't support renew", assumedRoleCredential)

	resp, err := e.Backend.HandleRequest(e.Context, req)

	if err != nil {
		if err.Error() != wantErr.Error() {
			t.Fatalf("expected error %v, got error %v", wantErr, err)
		}

		return
	}

	t.Fatalf("expected error %v, got response %v", wantErr, resp)
}

func (e *testEnv) RevokeARNBasedCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RevokeOperation,
		Storage:   e.Storage,
		Secret:    e.MostRecentSecret,
		Data: map[string]interface{}{
			"lease_id": "foo",
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}
