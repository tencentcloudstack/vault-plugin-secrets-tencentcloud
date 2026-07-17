package clients

import "testing"

func TestChainedCredsPrefersConfigurationOverEnv(t *testing.T) {
	t.Setenv("TENCENTCLOUD_SECRET_ID", "env-secret-id")
	t.Setenv("TENCENTCLOUD_SECRET_KEY", "env-secret-key")

	creds, err := chainedCreds("config-secret-id", "config-secret-key")
	if err != nil {
		t.Fatalf("chainedCreds returned error: %v", err)
	}

	if got := creds.GetSecretId(); got != "config-secret-id" {
		t.Fatalf("expected config secret id, got %q", got)
	}
	if got := creds.GetSecretKey(); got != "config-secret-key" {
		t.Fatalf("expected config secret key, got %q", got)
	}
}
