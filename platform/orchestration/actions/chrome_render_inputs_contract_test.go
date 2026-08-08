// FILE: platform/orchestration/actions/chrome_render_inputs_contract_test.go
//
// Pins for the chrome render-inputs fingerprint (bugs_open/117). The
// fingerprint's promise is that the STAMP (render_site_components) and the
// RECOMPUTE (stale_site_components) are the same expression, and that the
// expression's coverage tracks the code it models. None of that is enforceable
// by the compiler — the checker lives in discovery_checks, which cannot import
// this package — so, like TestDiscoveryChromeLockFilterMatchesSharedPredicate
// beside it, these tests build the expected text from one side and look for it
// in an artefact that side cannot influence.

package actions

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// TestChromeFingerprintSpecsAspectsMatchResolver pins the fingerprint's
// site_specs aspect list to the one resolveConfigPath actually consults. The
// fingerprint covers the config.* schema-fill sources ONLY because the two
// lists agree — a resolver that gains an aspect the fingerprint does not hash
// reopens the invisible-drift hole this whole mechanism closes (a GTM id or
// colour scheme change that no check can see).
//
// The slice is extracted from the source ANCHORED at the function name, not
// grepped bare, so a comment or another function's list cannot satisfy it
// (source-scan tests make first occurrences load-bearing otherwise).
func TestChromeFingerprintSpecsAspectsMatchResolver(t *testing.T) {
	src, err := os.ReadFile("plan_sections_action.go")
	if err != nil {
		t.Fatalf("cannot read plan_sections_action.go: %v", err)
	}

	fn := regexp.MustCompile(`(?s)func \(r \*sourceResolver\) resolveConfigPath.*?range \[\]string\{([^}]*)\}`)
	m := fn.FindSubmatch(src)
	if m == nil {
		t.Fatal("resolveConfigPath's aspect slice not found — if the resolver was restructured, re-pin this test AND re-check the fingerprint's specs subquery")
	}
	var aspects []string
	for _, q := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(string(m[1]), -1) {
		aspects = append(aspects, q[1])
	}
	if len(aspects) == 0 {
		t.Fatal("resolveConfigPath's aspect slice parsed empty")
	}

	// The fingerprint's specs subquery must name exactly the same aspects.
	specsClause := regexp.MustCompile(`(?s)'specs',.*?aspect IN \(([^)]*)\)`).FindStringSubmatch(datahelpers.ChromeRenderInputsSQL)
	if specsClause == nil {
		t.Fatal("ChromeRenderInputsSQL has no specs aspect IN (...) clause")
	}
	var sqlAspects []string
	for _, q := range regexp.MustCompile(`'([^']+)'`).FindAllStringSubmatch(specsClause[1], -1) {
		sqlAspects = append(sqlAspects, q[1])
	}

	if len(sqlAspects) != len(aspects) {
		t.Fatalf("aspect lists differ in size: resolveConfigPath %v vs fingerprint %v — change both sides together, and note that widening the fingerprint restamps nothing until a render or a version bump", aspects, sqlAspects)
	}
	for _, a := range aspects {
		found := false
		for _, s := range sqlAspects {
			if s == a {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("resolveConfigPath consults aspect %q but the fingerprint does not hash it — config.* drift under that aspect is invisible to stale_site_components", a)
		}
	}
}

// TestRendererStampsTheSharedFingerprint asserts the chrome store UPDATE
// stamps render_inputs with the SHARED expression — not a hand-written copy,
// and not nothing. A renderer that stops stamping makes every row read as
// stale forever (permanent rebuild churn); one that stamps its own spelling
// drifts from the checker and manufactures false positives or negatives.
func TestRendererStampsTheSharedFingerprint(t *testing.T) {
	src, err := os.ReadFile("render_site_components_action.go")
	if err != nil {
		t.Fatalf("cannot read render_site_components_action.go: %v", err)
	}
	s := string(src)
	idx := strings.Index(s, "render_inputs = (")
	if idx == -1 {
		t.Fatal("the chrome store UPDATE no longer stamps render_inputs — every site_components row will compare stale forever")
	}
	window := s[idx:min(idx+220, len(s))]
	if !strings.Contains(window, "datahelpers.ChromeRenderInputsSQL") {
		t.Fatalf("render_inputs is stamped with something other than datahelpers.ChromeRenderInputsSQL — stamp and recompute must be ONE expression:\n%s", window)
	}
}

// TestStaleCheckUsesTheSharedFingerprintAndLockPredicate asserts the checker
// side of the same contract, in the file this package cannot import: the
// recompute embeds the shared expression, rows are filtered on the SAME
// writable predicate the render path enforces (a locked slot can never be
// restamped, so firing on one churns items to `unresolved`), and the item key
// is the site-level `stale_chrome` this check is the sole producer of.
func TestStaleCheckUsesTheSharedFingerprintAndLockPredicate(t *testing.T) {
	src, err := os.ReadFile("discovery_checks/check_integrity.go")
	if err != nil {
		t.Fatalf("cannot read discovery_checks/check_integrity.go: %v", err)
	}
	s := string(src)

	if n := strings.Count(s, "datahelpers.ChromeRenderInputsSQL"); n < 2 {
		t.Errorf("stale_site_components should embed the shared fingerprint in both SELECT and predicate (found %d references)", n)
	}
	if !strings.Contains(s, `datahelpers.AgentWritableSQLFor("sc.")`) {
		t.Error("stale_site_components no longer filters on the shared writable predicate — locked slots will churn items to unresolved")
	}
	if !strings.Contains(s, `ItemKey:      "stale_chrome"`) {
		t.Error("stale_site_components no longer files the site-level stale_chrome key — check the register entry (improvement-loop) before changing the key shape")
	}
}

// TestChromeFingerprintCorrelatesOnAliasedRow documents the fragment's one
// composition rule: it correlates on a site_components row aliased `sc`. An
// embedding without that alias would let each subquery's own site_id column
// (pages, assets, site_specs all have one) capture the reference — every site
// compared with itself, silently.
func TestChromeFingerprintCorrelatesOnAliasedRow(t *testing.T) {
	for _, ref := range []string{"sc.site_id", "sc.component_id"} {
		if !strings.Contains(datahelpers.ChromeRenderInputsSQL, ref) {
			t.Errorf("ChromeRenderInputsSQL no longer correlates via %s", ref)
		}
	}
}
