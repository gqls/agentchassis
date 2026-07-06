package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// Tests for buildTokenAliases (R6f vocabulary bridge, step 11 of
// RenderCSSFromSpecAction). The contract: every tokenAliases name not
// already DEFINED in the css (name followed by a colon) is appended in a
// trailing :root block; names the theme defines are left alone; var()
// usages and longer sibling names must not count as definitions.

func TestBuildTokenAliases_EmptyCSSAppendsAll(t *testing.T) {
	out := buildTokenAliases("", zap.NewNop())
	if out == "" {
		t.Fatalf("expected alias block for empty css, got empty string")
	}
	if !strings.Contains(out, "/* renderer-enforced compatibility aliases") {
		t.Errorf("missing header comment in: %q", out)
	}
	for _, a := range tokenAliases {
		if !strings.Contains(out, a.Name+": "+a.Value+";") {
			t.Errorf("alias %s not appended", a.Name)
		}
	}
}

func TestBuildTokenAliases_DefinedNameIsSkipped(t *testing.T) {
	css := ":root {\n  --border-radius: 6px;\n}\n"
	out := buildTokenAliases(css, zap.NewNop())
	if strings.Contains(out, "--border-radius:") {
		t.Errorf("--border-radius already defined but was re-appended: %q", out)
	}
	if !strings.Contains(out, "--shadow: ") {
		t.Errorf("--shadow should still be appended when only --border-radius is defined")
	}
}

func TestBuildTokenAliases_AllDefinedReturnsEmpty(t *testing.T) {
	var b strings.Builder
	b.WriteString(":root {\n")
	for _, a := range tokenAliases {
		b.WriteString("  " + a.Name + ": red;\n")
	}
	b.WriteString("}\n")
	if out := buildTokenAliases(b.String(), zap.NewNop()); out != "" {
		t.Errorf("expected empty output when all names defined, got: %q", out)
	}
}

func TestBuildTokenAliases_UsageDoesNotCountAsDefinition(t *testing.T) {
	css := ".card { box-shadow: var(--shadow); border-radius: var(--border-radius); }"
	out := buildTokenAliases(css, zap.NewNop())
	if !strings.Contains(out, "--shadow: ") || !strings.Contains(out, "--border-radius: ") {
		t.Errorf("var() usages must not suppress the aliases: %q", out)
	}
}

func TestBuildTokenAliases_SiblingNameDoesNotCount(t *testing.T) {
	css := ":root { --shadow-md: 0 1px 2px rgba(0,0,0,.1); }"
	out := buildTokenAliases(css, zap.NewNop())
	if !strings.Contains(out, "--shadow: ") {
		t.Errorf("--shadow-md definition must not suppress the --shadow alias: %q", out)
	}
}

func TestBuildTokenAliases_Idempotent(t *testing.T) {
	first := buildTokenAliases("", zap.NewNop())
	if second := buildTokenAliases(first, zap.NewNop()); second != "" {
		t.Errorf("second pass over own output should append nothing, got: %q", second)
	}
}
