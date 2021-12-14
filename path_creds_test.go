package tencentcloud

import (
	"testing"
)

func TestGenerateUsername(t *testing.T) {
	result := generateUsername("displayName", "roleName")
	if len(result) > 64 {
		t.Fatal("too long: " + result)
	}
	result = generateUsername("displayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayName", "roleName")
	if len(result) > 64 {
		t.Fatal("too long: " + result)
	}
	result = generateUsername("displayName", "roleNameroleNameroleNameroleNameroleNameroleNameroleNameroleNameroleNameroleName")
	if len(result) > 64 {
		t.Fatal("too long: " + result)
	}
}

func TestGenerateRoleSessionName(t *testing.T) {
	result := generateRoleSessionName("displayName", "roleName")
	if len(result) > 32 {
		t.Fatalf("too long: %d, %s", len(result), result)
	}
	result = generateRoleSessionName("displayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayNamedisplayName", "roleName")
	if len(result) > 32 {
		t.Fatal("too long: " + result)
	}
	result = generateRoleSessionName("displayName", "roleNameroleNameroleNameroleNameroleNameroleNameroleNameroleNameroleNameroleName")
	if len(result) > 32 {
		t.Fatal("too long: " + result)
	}
}
