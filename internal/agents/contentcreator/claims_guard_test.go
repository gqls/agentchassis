// FILE: internal/agents/contentcreator/claims_guard_test.go
//
// Every assertion here checks that the mechanism FIRED, not merely that the run
// was quiet — the contract style of save_sections_claims_guard_test.go, and the
// reason is sharper for this guard than for that one. This guard only records
// and annotates, so DELETING ITS CALL SITE BREAKS NO COMPILE AND NO OTHER TEST.
// Nothing else in the package would notice. That is why
// TestHandleMessageActuallyCallsTheGuard reads the source.

package contentcreator

import (
	"os"
	"strings"
	"testing"
)

// The motivating fabrication, verbatim from bugs_open/123.
const fabricatedStat = "Industry data shows that large language models experience hallucination rates between 3% and 10% depending on the task."

func TestBannedClaimInGeneratedTextIsDetected(t *testing.T) {
	banned, _ := scanGeneratedText(
		"Our approach is proven.\n\nOur analysis is always accurate.\n",
		"markdown",
		claimsGuardToggles{CheckClaims: true},
	)
	if len(banned) != 1 {
		t.Fatalf("want exactly 1 banned finding, got %d: %+v", len(banned), banned)
	}
	// Proves the SHARED engine fired rather than a private copy: only the
	// fleet-wide/per-site set carries a Pattern, and a hand-rolled scan in this
	// package would have no reason to populate it.
	if banned[0].Pattern == "" {
		t.Errorf("finding carries no Pattern — that means it did not come from the shared "+
			"banned-claim set, which is the only engine this seam may use: %+v", banned[0])
	}
	if banned[0].Check != "banned_claim" {
		t.Errorf("check = %q, want %q", banned[0].Check, "banned_claim")
	}
}

// THE LEVER PAIR. The OFF leg is what makes the ON leg mean something: a guard
// that fired regardless of its lever, or a test that would pass with the scan
// deleted, fails one leg or the other.
func TestCheckClaimsLeverIsLoadBearing(t *testing.T) {
	const text = "Our analysis is always accurate."

	on, _ := scanGeneratedText(text, "markdown", claimsGuardToggles{CheckClaims: true})
	if len(on) == 0 {
		t.Fatalf("lever ON must produce a finding for %q", text)
	}
	off, _ := scanGeneratedText(text, "markdown", claimsGuardToggles{CheckClaims: false})
	if len(off) != 0 {
		t.Errorf("lever OFF must produce no findings, got %+v", off)
	}
}

// Pins the owner ruling of 2026-08-02 as a named test, so a future "helpful"
// default flip has to delete an assertion that says why it exists rather than
// change a bare literal.
func TestUncitedStatsAreOptInAndDefaultOff(t *testing.T) {
	var empty RequestPayload
	got := togglesFromRequest(empty)
	if !got.CheckClaims {
		t.Errorf("check_claims must default ON (the floor's lever of the same name), got OFF")
	}
	if got.CheckUncitedStats {
		t.Errorf("check_uncited_stats must default OFF — owner ruling 2026-08-02: new authority on a " +
			"shared seam ships as an opt-in field with the unsafe default OFF")
	}

	// And the default is load-bearing on real text, not just on the struct.
	_, uncitedByDefault := scanGeneratedText(fabricatedStat, "markdown", togglesFromRequest(empty))
	if len(uncitedByDefault) != 0 {
		t.Errorf("with the default toggles the uncited-stat scan must not run, got %+v", uncitedByDefault)
	}

	on := true
	var req RequestPayload
	req.Data.CheckUncitedStats = true
	req.Data.CheckClaims = &on
	_, uncitedWhenOn := scanGeneratedText(fabricatedStat, "markdown", togglesFromRequest(req))
	if len(uncitedWhenOn) != 1 {
		t.Fatalf("VACUOUS PASS: with check_uncited_stats ON the bug's own fabrication produces %d "+
			"findings, so the default-OFF assertion above proves nothing", len(uncitedWhenOn))
	}
}

// The format router matters here specifically: this agent emits markdown and
// plain text, and the HTML extractor fuses a whole markdown document into one
// block (see claims_plaintext.go). A fused block can match across a paragraph
// boundary the prose does not have.
func TestMarkdownIsNotScannedAsHTML(t *testing.T) {
	md := "## Our position\n\n" + fabricatedStat + "\n\nWe publish our workings.\n"

	req := RequestPayload{}
	req.Data.CheckUncitedStats = true
	_, uncited := scanGeneratedText(md, "markdown", togglesFromRequest(req))
	if len(uncited) != 1 {
		t.Fatalf("markdown must be split before scanning; got %d findings: %+v", len(uncited), uncited)
	}
	if strings.Contains(uncited[0].Snippet, "Our position") {
		t.Errorf("the heading fused into the finding's block — the splitter did not run: %q", uncited[0].Snippet)
	}
}

func TestEmptyGenerationScansNothing(t *testing.T) {
	banned, uncited := scanGeneratedText("", "markdown", claimsGuardToggles{CheckClaims: true, CheckUncitedStats: true})
	if len(banned) != 0 || len(uncited) != 0 {
		t.Errorf("empty text must produce no findings, got %d/%d", len(banned), len(uncited))
	}
}

// ---------------------------------------------------------------------------
// The negative a mock cannot assert
// ---------------------------------------------------------------------------

// funcBodyOf returns the body of the named top-level function, so a neighbouring
// function can never be read by mistake. Same shape as
// platform/orchestration/claim_recovery_guard_test.go, and used for the same
// reason: the property is control flow no type system can see.
func funcBodyOf(t *testing.T, file, funcDecl string) string {
	t.Helper()
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	body := string(src)
	i := strings.Index(body, funcDecl)
	if i < 0 {
		t.Fatalf("function %q not found in %s — renamed? update this test with it", funcDecl, file)
	}
	body = body[i+len(funcDecl):]
	if j := strings.Index(body, "\nfunc "); j >= 0 {
		body = body[:j]
	}
	return body
}

// THE TEST THAT EXISTS BECAUSE NOTHING ELSE WOULD FAIL.
//
// The guard is annotate-only: remove the call from handleMessage and the package
// still compiles, every other test in this file still passes (they call
// scanGeneratedText directly), and the agent still answers every request
// perfectly happily — with the scan silently gone. That is the precise shape of
// a mechanism rotting unnoticed, and this is the only thing standing between the
// guard and it.
//
// MUTATION-PROVED before commit: comment out the scanGeneratedText call in
// handleMessage, run `go test ./internal/agents/contentcreator/`, and only this
// test fails.
func TestHandleMessageActuallyCallsTheGuard(t *testing.T) {
	body := funcBodyOf(t, "agent.go", "func (a *Agent) handleMessage")

	for _, needle := range []string{
		"scanGeneratedText(",
		"a.recordClaimFindings(",
		"ClaimFindings:",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("handleMessage no longer contains %q — the claims guard has been "+
				"disconnected. Nothing else in this package would have failed. (bugs_open/123)", needle)
		}
	}

	// Ordering is the property, not just presence: the scan must happen BEFORE
	// the payload is assembled, or the findings cannot ride on it.
	scanAt := strings.Index(body, "scanGeneratedText(")
	payloadAt := strings.Index(body, "responsePayload := ResponsePayload{")
	if scanAt < 0 || payloadAt < 0 || scanAt > payloadAt {
		t.Errorf("the scan must run before the response payload is assembled (scan at %d, payload at %d)",
			scanAt, payloadAt)
	}
}

// bugs_open/158 owns both reply-delivery sites in this package. This change must
// not have touched them, and a future edit to this guard must not drift into
// them either — a collision on a shared tree is expensive precisely because it
// looks like ordinary progress.
func TestClaimsGuardDoesNotTouchReplyDelivery(t *testing.T) {
	src, err := os.ReadFile("claims_guard.go")
	if err != nil {
		t.Fatalf("read claims_guard.go: %v", err)
	}
	for _, forbidden := range []string{"sendSuccessResponse", "sendErrorResponse", "DeliverReply"} {
		if strings.Contains(string(src), "a."+forbidden) || strings.Contains(string(src), forbidden+"(") {
			t.Errorf("claims_guard.go references %q — the reply-delivery path belongs to the "+
				"bugs_open/158 lane and this seam must not edit or adopt it", forbidden)
		}
	}
}
