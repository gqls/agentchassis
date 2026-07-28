package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// The helper's two decision branches, without a DB. The rewrite/unlink
// semantics themselves stay covered by datahelpers' link_repair_test.go — what
// is pinned here is that the rerender seam actually APPLIES them, and that its
// degrade ships the input untouched (the branch whose failure mode is silent
// damage: an over-eager "repair" with no index would strip every internal link
// on the page).

func TestRepairOutboundPageLinksFailsOpenWithoutIndex(t *testing.T) {
	in := `<a href="/never-built">x</a>`
	got := repairOutboundPageLinks(context.Background(),
		ActionParams{ExecutionContext: &types.ExecutionContext{}},
		uuid.Nil, "d.com", "p", "/p.html", in, nil, false, zap.NewNop())
	if got != in {
		t.Fatalf("no index must mean NO change (fail-open), got %q", got)
	}
}

func TestRepairOutboundPageLinksRepairsAgainstIndex(t *testing.T) {
	idx := datahelpers.NewPageURLIndex([]string{"/about.html", "/cases/index.html"})
	in := `<a href="/about">fix</a> <a href="/cases">alias</a> <a href="/never-built">gone</a>`
	got := repairOutboundPageLinks(context.Background(),
		ActionParams{ExecutionContext: &types.ExecutionContext{}},
		uuid.Nil, "d.com", "p", "/p.html", in, idx, true, zap.NewNop())
	if !strings.Contains(got, `href="/about.html"`) {
		t.Fatalf("extension-omitted href must be rewritten to the stored url: %q", got)
	}
	if !strings.Contains(got, `href="/cases"`) {
		t.Fatalf("a section-index alias is a VALID target (the index normalises index.html away) and must stay byte-identical: %q", got)
	}
	if strings.Contains(got, `href="/never-built"`) {
		t.Fatalf("phantom href must be unlinked: %q", got)
	}
	if !strings.Contains(got, ">gone<") && !strings.Contains(got, "gone") {
		t.Fatalf("unlinking must keep the anchor text (body prose is content): %q", got)
	}
}

func TestRepairOutboundPageLinksLeavesCleanPagesAlone(t *testing.T) {
	idx := datahelpers.NewPageURLIndex([]string{"/about.html"})
	in := `<a href="/about.html">fine</a> <a href="https://example.org/x">external</a>`
	got := repairOutboundPageLinks(context.Background(),
		ActionParams{ExecutionContext: &types.ExecutionContext{}},
		uuid.Nil, "d.com", "p", "/p.html", in, idx, true, zap.NewNop())
	if got != in {
		t.Fatalf("a page with no dead internal links must pass through byte-identical, got %q", got)
	}
}
