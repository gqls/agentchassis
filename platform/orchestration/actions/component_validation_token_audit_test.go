package actions

import (
	"testing"

	"go.uber.org/zap"
)

// Tests for AuditTemplateTokens (R6f vocabulary detection net). Contract:
// returns distinct non-canonical var(--…) names in first-seen order, never
// errors, warns only when logger != nil and unknowns exist. Canonical tokens
// and compatibility aliases must both be accepted.

func TestAuditTemplateTokens_AllCanonicalReturnsNil(t *testing.T) {
	tpl := `<section class="x" style="color: var(--color-text); padding: var(--section-pad-y); border-radius: var(--radius);">ok</section>`
	if got := AuditTemplateTokens("features", tpl, zap.NewNop()); got != nil {
		t.Errorf("expected nil for all-canonical template, got %v", got)
	}
}

func TestAuditTemplateTokens_AliasesAccepted(t *testing.T) {
	// the exact names D2a's buildTokenAliases guarantees
	tpl := `<div style="border-radius: var(--border-radius); box-shadow: var(--shadow); color: var(--hero-ink); gap: var(--spacing-section); max-width: var(--container-max-width);">x</div>`
	if got := AuditTemplateTokens("hero", tpl, zap.NewNop()); got != nil {
		t.Errorf("compatibility aliases must be canonical, got unknowns %v", got)
	}
}

func TestAuditTemplateTokens_DetectsUnknown(t *testing.T) {
	tpl := `<section style="color: var(--color-text); background: var(--mystery-bg); outline: var(--totally-new);">x</section>`
	got := AuditTemplateTokens("features", tpl, zap.NewNop())
	if len(got) != 2 || got[0] != "--mystery-bg" || got[1] != "--totally-new" {
		t.Errorf("expected [--mystery-bg --totally-new] in order, got %v", got)
	}
}

func TestAuditTemplateTokens_DeduplicatesAndPreservesOrder(t *testing.T) {
	tpl := `var(--zeta) var(--alpha) var(--zeta) var(--alpha) var(--zeta)`
	got := AuditTemplateTokens("x", tpl, zap.NewNop())
	if len(got) != 2 || got[0] != "--zeta" || got[1] != "--alpha" {
		t.Errorf("expected [--zeta --alpha] first-seen deduped, got %v", got)
	}
}

func TestAuditTemplateTokens_EmptyTemplate(t *testing.T) {
	if got := AuditTemplateTokens("x", "", zap.NewNop()); got != nil {
		t.Errorf("empty template should return nil, got %v", got)
	}
}

func TestAuditTemplateTokens_NilLoggerDoesNotPanic(t *testing.T) {
	tpl := `var(--unknown-token)`
	got := AuditTemplateTokens("x", tpl, nil)
	if len(got) != 1 || got[0] != "--unknown-token" {
		t.Errorf("nil logger must still return unknowns without panic, got %v", got)
	}
}

func TestAuditTemplateTokens_WhitespaceInVarCall(t *testing.T) {
	// var( --name ) with spaces, and var(--name, fallback) forms
	tpl := `a: var( --color-text ); b: var(--nope, 10px);`
	got := AuditTemplateTokens("x", tpl, zap.NewNop())
	if len(got) != 1 || got[0] != "--nope" {
		t.Errorf("expected only [--nope] (spaced canonical accepted, fallback form parsed), got %v", got)
	}
}
