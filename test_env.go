package tencentcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

/*
	testEnv allows us to reuse the same requests and response-checking
	for both integration tests that don't hit TencentCloud's real API, and
	for acceptance tests that do hit their real API.
*/
type testEnv struct {
	SecretId  string
	SecretKey string
	RoleARN   string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	MostRecentSecret *logical.Secret
}

// AddConfig
func (e *testEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"secret_id":  e.SecretId,
			"secret_key": e.SecretKey,
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

// ReadFirstConfig
func (e *testEnv) ReadFirstConfig(t *testing.T) {
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
	if resp.Data["secret_id"] != e.SecretId {
		t.Fatal("expected secret_id of " + e.SecretId)
	}
	if resp.Data["secret_key"] != e.SecretKey {
		t.Fatal("secret_key should not be returned")
	}
}

// UpdateConfig
func (e *testEnv) UpdateConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"secret_id":  "foo",
			"secret_key": "bar",
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

// ReadSecondConfig
func (e *testEnv) ReadSecondConfig(t *testing.T) {
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
	if resp.Data["secret_id"] != "foo" {
		t.Fatal("expected secret_id of foo")
	}
	if resp.Data["secret_key"] != "bar" {
		t.Fatal("expected secret_key of bar")
	}
}

// DeleteConfig
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

// ReadEmptyConfig
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

// AddPolicyBasedRole
func (e *testEnv) AddPolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "role/policy-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"remote_policies": []string{
				"policy_name:QcloudAFCFullAccess,scope:All",
				"policy_name:QcloudAFFullAccess,scope:All",
				"policy_name:QcloudAMEReadOnlyAccess,scope:All",
			},
			"inline_policies": policyDocument,
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

// ReadPolicyBasedRole
func (e *testEnv) ReadPolicyBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "role/policy-based",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["role_arn"] != "" {
		t.Fatalf("expected no role_arn but received %s", resp.Data["role_arn"])
	}

	inlinePolicies := resp.Data["inline_policies"].([]*inlinePolicy)
	for i, inlinePolicy := range inlinePolicies {
		if inlinePolicy.PolicyDocument["version"] != "2.0" {
			t.Fatalf("expected version of 2.0 but received %s", inlinePolicy.PolicyDocument["version"])
		}
		stmts := inlinePolicy.PolicyDocument["statement"].([]interface{})
		if len(stmts) != 1 {
			t.Fatalf("expected 1 statement but received %d", len(stmts))
		}
		stmt := stmts[0].(map[string]interface{})
		action := stmt["action"].([]interface{})[0].(string)
		if stmt["effect"] != "allow" {
			t.Fatalf("expected allow statement but received %s", stmt["effect"])
		}
		resource := stmt["resource"].(string)
		if resource != "*" {
			t.Fatalf("received incorrect resource: %s", resource)
		}
		switch i {
		case 0:
			if action != "af:*" {
				t.Fatalf("expected af:* but received %s", action)
			}
		case 1:
			if action != "afc:*" {
				t.Fatalf("expected afc:* but received %s", action)
			}
		}
	}

	remotePolicies := resp.Data["remote_policies"].([]*remotePolicy)
	for i, remotePol := range remotePolicies {
		switch i {
		case 0:
			if remotePol.PolicyName != "QcloudAFCFullAccess" {
				t.Fatalf("received unexpected policy type of %s", remotePol.PolicyName)
			}
			if remotePol.Scope != "All" {
				t.Fatalf("received unexpected policy type of %s", remotePol.Scope)
			}
		case 1:
			if remotePol.PolicyName != "QcloudAFFullAccess" {
				t.Fatalf("received unexpected policy type of %s", remotePol.PolicyName)
			}
			if remotePol.Scope != "All" {
				t.Fatalf("received unexpected policy type of %s", remotePol.Scope)
			}
		}
	}

	ttl := fmt.Sprintf("%d", resp.Data["ttl"])
	if ttl != "0" {
		t.Fatalf("expected ttl of 0 but received %s", ttl)
	}

	maxTTL := fmt.Sprintf("%d", resp.Data["max_ttl"])
	if maxTTL != "0" {
		t.Fatalf("expected max_ttl of 0 but received %s", maxTTL)
	}
}

// AddARNBasedRole
func (e *testEnv) AddARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "role/role-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"role_arn": e.RoleARN,
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

// ReadARNBasedRole
func (e *testEnv) ReadARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "role/role-based",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["role_arn"] != e.RoleARN {
		t.Fatalf("received unexpected role_arn of %s", resp.Data["role_arn"])
	}

	inlinePolicies := resp.Data["inline_policies"].([]*inlinePolicy)
	if len(inlinePolicies) != 0 {
		t.Fatalf("expected no inline policies but received %+v", inlinePolicies)
	}

	remotePolicies := resp.Data["remote_policies"].([]*remotePolicy)
	if len(remotePolicies) != 0 {
		t.Fatalf("expected no remote policies but received %+v", remotePolicies)
	}

	ttl := fmt.Sprintf("%d", resp.Data["ttl"])
	if ttl != "0" {
		t.Fatalf("expected ttl of 0 but received %s", ttl)
	}

	maxTTL := fmt.Sprintf("%d", resp.Data["max_ttl"])
	if maxTTL != "0" {
		t.Fatalf("expected max_ttl of 0 but received %s", maxTTL)
	}
}

// UpdateARNBasedRole
func (e *testEnv) UpdateARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "role/role-based",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"role_arn": "qcs::cam::uin/100021543443:roleName/firingrole001",
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

// ReadUpdatedRole
func (e *testEnv) ReadUpdatedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "role/role-based",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["role_arn"] != "qcs::cam::uin/100021543443:roleName/firingrole001" {
		t.Fatalf("received unexpected role_arn of %s", resp.Data["role_arn"])
	}

	inlinePolicies := resp.Data["inline_policies"].([]*inlinePolicy)
	if len(inlinePolicies) != 0 {
		t.Fatalf("expected no inline policies but received %+v", inlinePolicies)
	}

	remotePolicies := resp.Data["remote_policies"].([]*remotePolicy)
	if len(remotePolicies) != 0 {
		t.Fatalf("expected no remote policies but received %+v", remotePolicies)
	}

	ttl := fmt.Sprintf("%d", resp.Data["ttl"])
	if ttl != "0" {
		t.Fatalf("expected ttl of 100 but received %s", ttl)
	}

	maxTTL := fmt.Sprintf("%d", resp.Data["max_ttl"])
	if maxTTL != "0" {
		t.Fatalf("expected max_ttl of 1000 but received %s", maxTTL)
	}
}

// ListTwoRoles
func (e *testEnv) ListTwoRoles(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "role",
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

// DeleteARNBasedRole
func (e *testEnv) DeleteARNBasedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "role/role-based",
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

// ListOneRole
func (e *testEnv) ListOneRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "role",
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

// ReadPolicyBasedCreds
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

	if resp.Data["secret_id"] == "" {
		t.Fatal("failed to receive secret_id")
	}
	if resp.Data["secret_key"] == "" {
		t.Fatal("failed to receive secret_key")
	}
	e.MostRecentSecret = resp.Secret
}

// RenewPolicyBasedCreds
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
	if resp.Secret != e.MostRecentSecret {
		t.Fatalf("expected %+v but got %+v", e.MostRecentSecret, resp.Secret)
	}
}

// RevokePolicyBasedCreds
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

// ReadARNBasedCreds
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

	if resp.Data["secret_id"] == "" {
		t.Fatal("received blank secret_id")
	}
	if resp.Data["secret_key"] == "" {
		t.Fatal("received blank secret_key")
	}
	if fmt.Sprintf("%s", resp.Data["expiration"]) == "" {
		t.Fatal("received blank expiration")
	}
	if resp.Data["token"] == "" {
		t.Fatal("received blank token")
	}
	e.MostRecentSecret = resp.Secret
}

// RenewARNBasedCreds
func (e *testEnv) RenewARNBasedCreds(t *testing.T) {
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
	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

// RevokeARNBasedCreds
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

const policyDocument = ` [
    {
        "version":"2.0",
        "statement":[
            {
                "action":[
                    "af:*"
                ],
                "resource":"*",
                "effect":"allow"
            }
        ]
    },
    {
        "version":"2.0",
        "statement":[
            {
                "action":[
                    "afc:*"
                ],
                "resource":"*",
                "effect":"allow"
            }
        ]
    }
]`
