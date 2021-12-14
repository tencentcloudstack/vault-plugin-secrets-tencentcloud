package tencentcloud

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/vault-plugin-secrets-tencentcloud/clients"
	"github.com/hashicorp/vault/sdk/logical"
)

func setup() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// All responses below are directly from AliCloud's documentation
		// and none reflect real values.
		action := r.Header.Get("X-TC-Action")
		switch action {

		case "AddUser":
			w.WriteHeader(200)
			w.Write([]byte(`{
			  "Response": {
				"Uid": 5648765,
				"Uin": 100000546533,
				"Name": "test124",
				"Password": "test123456",
				"SecretId": "faweffewagwaegawe",
				"SecretKey": "fawef23rjhiuaefhuaifhiuawef",
				"RequestId": "b46d2afe-6893-4529-bc96-2c82d9214957"
			  }
            }`))

		case "DeleteUser":
			w.WriteHeader(200)
			w.Write([]byte(`{
			  "Response": {
				"RequestId": "b46d2afe-6893-4529-bc96-2c82d9214957"
			  }
            }`))

		case "CreatePolicy":
			w.WriteHeader(200)
			w.Write([]byte(`{
				"Response": {
                   "PolicyId": 17698703,
                   "RequestId": "89360f78-b1dd-4e43-aa91-ecb2c8b8f282"
                }
			}`))

		case "DeletePolicy":
			w.WriteHeader(200)
			w.Write([]byte(`{
				"Response": {
					"RequestId": "1a21f666-d00e-4df8-92f7-7121f9012e43"
                }
			}`))

		case "AttachUserPolicy":
			w.WriteHeader(200)
			w.Write([]byte(`    {
				"Response": {
                   "RequestId": "1a21f666-d00e-4df8-92f7-7121f9012e43"
                }
			}`))

		case "DetachUserPolicy":
			w.WriteHeader(200)
			w.Write([]byte(`    {
				"Response": {
                   "RequestId": "1a21f666-d00e-4df8-92f7-7121f9012e43"
                }
			}`))
		case "ListPolicies":
			w.WriteHeader(200)
			w.Write([]byte(`{
				  "Response": {
					"ServiceTypeList": [],
					"List": [
					  {
						"PolicyId": 16313162,
						"PolicyName": "QcloudAccessForCDNRole",
						"AddTime": "2019-04-19 10:55:31",
						"Type": 2,
						"Description": "des",
						"CreateMode": 2,
						"Attachments": 0,
						"ServiceType": "cooperator",
						"IsAttached": 1,
						"Deactived": 1,
						"DeactivedDetail": [
						  "cvm"
						],
						"IsServiceLinkedPolicy": 1
					  }
					],
					"TotalNum": 239,
					"RequestId": "ae2bd2b7-1d55-4b0a-8154-e02407a2b390"
				  }
				}`))
		case "CreateAccessKey":
			w.WriteHeader(200)
			w.Write([]byte(`    {
				"Response": {
				   "AccessKey": {
				     "AccessKeyId": "ABBD8GFED7sSr33rSq9KK7h5ISSEoQrFXkmb",
				     "SecretAccessKey": "iDVjy9Mdr289A7d5efdBIMMIAqqKtNzX",
				     "Status": "Active",
                     "CreateTime": "2020-03-03 18:00:26"
				   },
				   "RequestId": "f8423e9b-a7da-488d-9539-333f1955ca78"
			    }
			}`))

		case "DeleteAccessKey":
			w.WriteHeader(200)
			w.Write([]byte(`    {
				"Response": {
                   "RequestId": "99d650e2-10fa-4c8f-819f-874578039641"
                }
			}`))

		case "AssumeRole":
			w.WriteHeader(200)
			w.Write([]byte(`    {
				 "Response": {
					"Credentials": {
					  "Token": "da1e9d2ee9dda83506832d5ecb903b790132dfe340001",
					  "TmpSecretId": "AKID65zyIP0mpXtaI******WIQVMn1umNH58",
					  "TmpSecretKey": "q95K84wrzuEGoc*******52boxvp71yoh"
					},
					"ExpiredTime": 1543914376,
					"Expiration": "2018-12-04T09:06:16Z",
					"RequestId": "4daec797-9cd2-4f09-9e7a-7d4c43b2a74c"
                 }
			}`))
		}
	}))
}

func teardown(ts *httptest.Server) {
	ts.Close()
}

func newIntegrationTestEnv(testURL string) (*testEnv, error) {
	ctx := context.Background()
	b, err := proxiedTestBackend(ctx, testURL)
	if err != nil {
		return nil, err
	}
	return &testEnv{
		SecretId:  "fizz",
		SecretKey: "buzz",
		RoleARN:   "acs:ram::5138828231865461:role/hastrustedactors",
		Backend:   b,
		Context:   ctx,
		Storage:   &logical.InmemStorage{},
	}, nil
}

// This test thoroughly exercises all endpoints, and tests the policy-based creds
// sunny path.
func TestDynamicPolicyBasedCreds(t *testing.T) {
	ts := setup()
	defer teardown(ts)

	integrationTestEnv, err := newIntegrationTestEnv(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", integrationTestEnv.AddConfig)
	t.Run("read config", integrationTestEnv.ReadFirstConfig)
	t.Run("update config", integrationTestEnv.UpdateConfig)
	t.Run("read config", integrationTestEnv.ReadSecondConfig)
	t.Run("delete config", integrationTestEnv.DeleteConfig)
	t.Run("read config", integrationTestEnv.ReadEmptyConfig)
	t.Run("add config", integrationTestEnv.AddConfig)

	t.Run("add policy-based role", integrationTestEnv.AddPolicyBasedRole)
	t.Run("read policy-based role", integrationTestEnv.ReadPolicyBasedRole)
	t.Run("add arn-based role", integrationTestEnv.AddARNBasedRole)
	t.Run("read arn-based role", integrationTestEnv.ReadARNBasedRole)
	t.Run("update arn-based role", integrationTestEnv.UpdateARNBasedRole)
	t.Run("read updated role", integrationTestEnv.ReadUpdatedRole)
	t.Run("list two roles", integrationTestEnv.ListTwoRoles)
	t.Run("delete arn-based role", integrationTestEnv.DeleteARNBasedRole)
	t.Run("list one role", integrationTestEnv.ListOneRole)

	t.Run("read policy-based creds", integrationTestEnv.ReadPolicyBasedCreds)
	t.Run("renew policy-based creds", integrationTestEnv.RenewPolicyBasedCreds)
	t.Run("revoke policy-based creds", integrationTestEnv.RevokePolicyBasedCreds)
}

// Since all endpoints were exercised in the previous test, we just need one that
// gets straight to the point testing the STS creds sunny path.
func TestDynamicSTSCreds(t *testing.T) {
	ts := setup()
	defer teardown(ts)

	integrationTestEnv, err := newIntegrationTestEnv(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", integrationTestEnv.AddConfig)
	t.Run("add arn-based role", integrationTestEnv.AddARNBasedRole)
	t.Run("read arn-based creds", integrationTestEnv.ReadARNBasedCreds)
	t.Run("renew arn-based creds", integrationTestEnv.RenewARNBasedCreds)
	t.Run("revoke arn-based creds", integrationTestEnv.RevokeARNBasedCreds)
}

func proxiedTestBackend(context context.Context, testURL string) (logical.Backend, error) {

	profile := clients.NewClientProfile()
	transport := &http.Transport{}
	capturer, _ := newURLUpdater(testURL)
	transport.Proxy = capturer.Proxy
	profile.HttpTransport = transport
	profile.HttpProfile.Scheme = "HTTP"
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: time.Hour,
			MaxLeaseTTLVal:     time.Hour,
		},
	}
	b := newBackend(profile)
	if err := b.Setup(context, conf); err != nil {
		panic(err)
	}
	return b, nil
}

/*
	The URL updater uses the Proxy on outbound requests to swap
	a real URL with one generated by httptest. This points requests
	at a local test server, and allows us to return expected
	responses.
*/
func newURLUpdater(testURL string) (*urlUpdater, error) {
	// Example testURL: https://127.0.0.1:46445
	u, err := url.Parse(testURL)
	if err != nil {
		return nil, err
	}
	return &urlUpdater{u, nil}, nil
}

type urlUpdater struct {
	testURL *url.URL
	request *http.Request
}

func (u *urlUpdater) Proxy(req *http.Request) (*url.URL, error) {
	req.URL.Scheme = u.testURL.Scheme
	req.URL.Host = u.testURL.Host
	return u.testURL, nil
}
