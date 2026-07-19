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
	if policy != "Defaults:codex-fixture timestamp_timeout=0\n" {
		t.Fatalf("policy must contain only the dedicated no-cache default, got %q", policy)
	}
	if strings.Contains(policy, " ALL") || strings.Contains(policy, "pam_exec") || strings.Contains(policy, "sudo -K") {
		t.Fatal("policy contains a command grant, PAM hook, or imperative cache clear")
	}
	if !regexp.MustCompile(`(?m)^Defaults:codex-fixture\s+timestamp_timeout=0\s*$`).MatchString(policy) {
		t.Fatal("dedicated timestamp timeout default is missing")
	}
}

func TestUnauthorizedIdentityCannotUseDedicatedPolicy(t *testing.T) {
	policy := readPolicy(t)
	if strings.Contains(policy, "ALL") || strings.Contains(policy, "@") { // no broad user/group selector
		t.Fatal("policy broadens beyond the dedicated identity")
	}
	if strings.Contains(policy, "codex-fixture ALL=") || strings.Contains(policy, "root") {
		t.Fatal("policy grants commands or root access")
	}
}

func TestProductionPolicyHasNoPAMOrClientInvocation(t *testing.T) {
	policy := readPolicy(t)
	for _, forbidden := range []string{"pam", "codex-authority-sudo", "command", "NOPASSWD", "runas"} {
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
