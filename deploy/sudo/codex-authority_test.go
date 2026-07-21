package sudo

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const policyPath = "codex-authority"

func readPolicy(t *testing.T) string {
	t.Helper()
	policy, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(policy)
}

func TestPolicyDisablesTimestampCachingDeclaratively(t *testing.T) {
	policy := readPolicy(t)
	want := "Defaults:coding-agent pam_service=codex-authority\n" +
		"Defaults:coding-agent pam_login_service=codex-authority\n" +
		"Defaults:coding-agent pam_askpass_service=codex-authority\n" +
		"Defaults:coding-agent timestamp_timeout=0\n" +
		"coding-agent ALL=(root:root) PASSWD: ALL\n"
	if policy != want {
		t.Fatalf("policy does not match the dedicated full-sudo contract, got %q", policy)
	}
	if strings.Contains(policy, "pam_exec") || strings.Contains(policy, "sudo -K") || strings.Contains(policy, "NOPASSWD") {
		t.Fatal("policy contains a PAM hook, imperative cache clear, or authentication bypass")
	}
	if !regexp.MustCompile(`(?m)^Defaults:coding-agent\s+timestamp_timeout=0\s*$`).MatchString(policy) {
		t.Fatal("dedicated timestamp timeout default is missing")
	}
}

func TestUnauthorizedIdentityCannotUseDedicatedPolicy(t *testing.T) {
	policy := readPolicy(t)
	if strings.Contains(policy, "Defaults:") && strings.Contains(policy, "Defaults:ALL") {
		t.Fatal("policy contains a global selector")
	}
	if strings.Contains(policy, "codex-fixture") || strings.Contains(policy, "(ALL:ALL)") {
		t.Fatal("policy uses a fixture identity or unrestricted runas list")
	}
}

func TestProductionPolicyHasNoPAMOrClientInvocation(t *testing.T) {
	policy := readPolicy(t)
	for _, forbidden := range []string{"pam_exec", "codex-authority-sudo", "NOPASSWD", "!authenticate", "exempt_group"} {
		if strings.Contains(strings.ToLower(policy), strings.ToLower(forbidden)) {
			t.Fatalf("production policy contains forbidden %q", forbidden)
		}
	}
}

func TestPolicyFixtureScaffoldingDoesNotMutateHost(t *testing.T) {
	// The actual sudo/PAM setup is Main-owned and must run in a private mount
	// namespace. This test only verifies the production fragment is inert text.
	if info, err := os.Stat(policyPath); err != nil || info.IsDir() {
		t.Fatal("policy fixture is not a regular file")
	}
}
