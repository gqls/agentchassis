package actions

import (
	"strings"
	"testing"
)

// F1.1b(c) build gate: the script and naming are pure functions — test the
// real ones. The k8s interaction is exercised live (the gate's first run IS
// its integration test; logs stay inspectable via TTL).

func TestBuildGateScriptScoping(t *testing.T) {
	script := buildGateScript("gqls", "agentchassis", "fix/e08c5b01",
		[]string{"platform/a.go", "platform/b.go"},
		[]string{"./platform/...", "./cmd/..."})

	// Clone targets the branch, depth 1, token from env not baked in.
	if !strings.Contains(script, `--branch "fix/e08c5b01"`) ||
		!strings.Contains(script, "${GITHUB_READ_TOKEN}") ||
		!strings.Contains(script, "github.com/gqls/agentchassis.git") {
		t.Fatalf("clone line wrong:\n%s", script)
	}
	// gofmt scoped to the changed files ONLY (never repo-wide — pre-existing
	// unformatted files must not fail the gate).
	if !strings.Contains(script, "gofmt -l 'platform/a.go' 'platform/b.go'") {
		t.Fatalf("gofmt not scoped to changed files:\n%s", script)
	}
	if strings.Contains(script, "gofmt -l .") {
		t.Fatalf("repo-wide gofmt is forbidden:\n%s", script)
	}
	// build targeted, never ./... at root (docs-dir package clashes).
	if !strings.Contains(script, "go build './platform/...'") ||
		!strings.Contains(script, "go build './cmd/...'") {
		t.Fatalf("build targets missing:\n%s", script)
	}
	if strings.Contains(script, "go build ./...\n") {
		t.Fatalf("root-wide build is forbidden:\n%s", script)
	}
}

func TestBuildGateScriptNoGoFiles(t *testing.T) {
	// No changed .go files → no gofmt stage at all (e.g. a pure-config plan
	// that still wants the repo to build).
	script := buildGateScript("gqls", "agentchassis", "fix/x", nil, []string{"./platform/..."})
	if strings.Contains(script, "gofmt") {
		t.Fatalf("gofmt stage should be absent with no .go files:\n%s", script)
	}
}

func TestGateJobName(t *testing.T) {
	cases := map[string]string{
		"fix/e08c5b01":           "build-gate-fix-e08c5b01",
		"Fix/Weird_Name!!":       "build-gate-fix-weird-name",
		strings.Repeat("a", 100): "build-gate-" + strings.Repeat("a", 49),
	}
	for in, want := range cases {
		got := gateJobName(in)
		if got != want {
			t.Fatalf("gateJobName(%q) = %q, want %q", in, got, want)
		}
		if len(got) > 63 {
			t.Fatalf("job name exceeds k8s limit: %q", got)
		}
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("a'b"); got != `'a'\''b'` {
		t.Fatalf("shellQuote: %q", got)
	}
}
