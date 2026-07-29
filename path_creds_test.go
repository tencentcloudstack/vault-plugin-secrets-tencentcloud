package tencentcloud

import (
	"regexp"
	"strings"
	"testing"
)

var generatedNameSuffixPattern = regexp.MustCompile(`\d+-[0-9a-f]{4}$`)

func TestGenerateUsername(t *testing.T) {
	testCases := []struct {
		name        string
		displayName string
		roleName    string
	}{
		{
			name:        "short name",
			displayName: "displayName",
			roleName:    "roleName",
		},
		{
			name:        "long display name",
			displayName: strings.Repeat("displayName", 10),
			roleName:    "roleName",
		},
		{
			name:        "long role name",
			displayName: "displayName",
			roleName:    strings.Repeat("roleName", 10),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateUsername(tc.displayName, tc.roleName)
			if len(result) > 64 {
				t.Fatalf("too long: %d, %s", len(result), result)
			}
			if !generatedNameSuffixPattern.MatchString(result) {
				t.Fatalf("unexpected generated username format: %s", result)
			}
		})
	}
}

func TestGenerateRoleSessionName(t *testing.T) {
	testCases := []struct {
		name        string
		displayName string
		roleName    string
	}{
		{
			name:        "short name",
			displayName: "displayName",
			roleName:    "roleName",
		},
		{
			name:        "long display name",
			displayName: strings.Repeat("displayName", 10),
			roleName:    "roleName",
		},
		{
			name:        "long role name",
			displayName: "displayName",
			roleName:    strings.Repeat("roleName", 10),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateRoleSessionName(tc.displayName, tc.roleName)
			if len(result) > 32 {
				t.Fatalf("too long: %d, %s", len(result), result)
			}
			if !generatedNameSuffixPattern.MatchString(result) {
				t.Fatalf("unexpected generated role session name format: %s", result)
			}
		})
	}
}

func TestGenerateName(t *testing.T) {
	testCases := []struct {
		name        string
		displayName string
		roleName    string
		maxLength   int
	}{
		{
			name:        "keeps full prefix when under max length",
			displayName: "display",
			roleName:    "role",
			maxLength:   64,
		},
		{
			name:        "truncates long display name",
			displayName: strings.Repeat("display", 10),
			roleName:    "role",
			maxLength:   32,
		},
		{
			name:        "truncates long role name",
			displayName: "display",
			roleName:    strings.Repeat("role", 10),
			maxLength:   32,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateName(tc.displayName, tc.roleName, tc.maxLength)
			if len(result) > tc.maxLength {
				t.Fatalf("too long: got %d, max %d, value %s", len(result), tc.maxLength, result)
			}

			match := generatedNameSuffixPattern.FindStringIndex(result)
			if match == nil || match[1] != len(result) {
				t.Fatalf("unexpected generated name format: %s", result)
			}

			prefix := result[:match[0]]
			expectedPrefix := tc.displayName + "-" + tc.roleName + "-"
			if !strings.HasPrefix(expectedPrefix, prefix) {
				t.Fatalf("expected prefix %q to be a leading substring of %q", prefix, expectedPrefix)
			}
		})
	}
}
