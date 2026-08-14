package actions

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gqls/agentchassis/platform/colour"
)

// inkVar pulls one emitted custom property out of a :root block.
func inkVar(t *testing.T, css, name string) string {
	t.Helper()
	re := regexp.MustCompile(regexp.QuoteMeta(name) + `:\s*(#[0-9a-fA-F]{3,8})\s*;`)
	m := re.FindStringSubmatch(css)
	if m == nil {
		t.Fatalf("no %s in the emitted block:\n%s", name, css)
	}
	return m[1]
}

// realPalette is a whole site palette, transcribed from its SERVED stylesheet.
type realPalette struct {
	domain                                   string
	primary, accent, background, surface, tx string
	wantPrimaryInk, wantAccentInk            string
}

func (p realPalette) toMap() map[string]string {
	return map[string]string{
		"primary": p.primary, "accent": p.accent,
		"background": p.background, "surface": p.surface,
		"text": p.tx,
	}
}

// TestInkPolicy_ProductionEmissionIsPinnedAtTheLiveThreshold pins what the
// renderer ACTUALLY EMITS, at the threshold production actually runs.
//
// WHY THIS EXISTS RATHER THAN THE platform/colour PIN ALONE, and it is a real
// gap rather than belt-and-braces. TestLegibleVariant_EmittedHexIsPinnedForRealPalettes
// pins LegibleVariant at colour.AANormal, which is the fixed WCAG constant 4.5.
// The moment inkMinContrast moved to 5.0 (owner ruling 2026-08-14) that test
// stayed green while describing seven colours THAT NO LONGER SHIP. A pin whose
// threshold cannot follow production is a pin on a hypothetical.
//
// This one drives buildLegibleInkDefaults — the real emitter, with the real
// four-ground compositing — at defaultInkPolicy(), so it moves when the shipped
// policy moves and fails when the shipped policy changes silently.
//
// EVERY INPUT IS `[TRANSCRIBED 2026-08-14]` from the served stylesheet:
//
//	curl -s https://<domain>/assets/css/styles.css \
//	  | grep -oE -- '--color-(primary|accent|background|surface|text): *#[0-9A-Fa-f]{3,8}'
//
// A bare curl returns 200 on all seven; no UA flag is needed. (LANDMINES.md
// said otherwise for a while and was wrong — measured from two hosts.)
//
// EVERY EXPECTED OUTPUT was computed by a SEPARATE implementation of the
// HSL/WCAG maths before this file was written, not read back out of this
// package. That ordering is the whole point: a table populated from the code it
// checks cannot fail. The separate implementation reproduced all seven of the
// existing 4.5 pins exactly, which is what earns its 5.0 column any trust.
func TestInkPolicy_ProductionEmissionIsPinnedAtTheLiveThreshold(t *testing.T) {
	// "" in a want column means "the source colour already clears the target,
	// so it must come back UNCHANGED" — the no-op branch, which is what makes
	// the derivation safe for the components already repointed.
	cases := []realPalette{
		{domain: "robot-hands.com", primary: "#1A1F2E", accent: "#E8500A", background: "#0F1218", surface: "#1E2535", tx: "#E2E8F0",
			wantPrimaryInk: "#94a0c2", wantAccentInk: "#f77f47"},
		{domain: "dartsonline.com", primary: "#1A1F2E", accent: "#E8311A", background: "#111520", surface: "#1E2436", tx: "#F0F2F7",
			wantPrimaryInk: "#94a0c2", wantAccentInk: "#f18072"},
		{domain: "webdesign.co.uk", primary: "#5c6b5d", accent: "#d4a373", background: "#f9f8f6", surface: "#ffffff", tx: "#2b2b2b",
			wantPrimaryInk: "", wantAccentInk: "#915e2c"},
		{domain: "vonc.com", primary: "#7c3cff", accent: "#fc5c7d", background: "#0a0a0f", surface: "#13121f", tx: "#f0eeff",
			wantPrimaryInk: "#a274ff", wantAccentInk: ""},

		// oufe sets primary == surface (#1B2A3B), so the ink is made legible
		// against its own colour. Thinnest in the fleet at 4.5 (+0.01); the
		// move to 5.0 is what buys it a real margin.
		{domain: "oufe.com", primary: "#1B2A3B", accent: "#C49A3C", background: "#0F1820", surface: "#1B2A3B", tx: "#E8E2D9",
			wantPrimaryInk: "#8ba9ca", wantAccentInk: "#c8a048"},

		{domain: "cookly.uk", primary: "#2C2C27", accent: "#C8502A", background: "#FDFAF4", surface: "#F0E8D5", tx: "#2C2C27",
			wantPrimaryInk: "", wantAccentInk: "#a24122"},
		{domain: "lendzy.co.uk", primary: "#1B2A4A", accent: "#E8700A", background: "#F8F7F4", surface: "#FFFFFF", tx: "#1A1A1A",
			wantPrimaryInk: "", wantAccentInk: "#a85107"},
	}

	for _, c := range cases {
		t.Run(c.domain, func(t *testing.T) {
			css := buildLegibleInkDefaults("", c.toMap(), defaultInkPolicy(), zapNop())
			if css == "" {
				t.Fatalf("%s: no ink block emitted", c.domain)
			}
			for _, slot := range []struct{ name, src, want string }{
				{"--color-primary-ink", c.primary, c.wantPrimaryInk},
				{"--color-accent-ink", c.accent, c.wantAccentInk},
			} {
				got := inkVar(t, css, slot.name)
				want := slot.want
				if want == "" {
					want = slot.src // the unchanged branch
				}
				if !strings.EqualFold(got, want) {
					t.Errorf("%s %s = %s, want %s at inkMinContrast=%.1f.\n"+
						"If this is a DELIBERATE threshold change, recompute every hex quoted in "+
						"bugs_open/122, the lane NOTES, the handoff and any council submission — "+
						"they all name these values, and a stale one reads as measured.",
						c.domain, slot.name, got, want, inkMinContrast)
				}
			}
		})
	}
}

// TestInkPolicy_EmittedInkClearsTheConfiguredTargetNotJustAA measures the
// emitted values rather than comparing them to a table, so it still says
// something true after a deliberate threshold change retires the pins above.
func TestInkPolicy_EmittedInkClearsTheConfiguredTargetNotJustAA(t *testing.T) {
	// dartsonline: the site the owner's canary is pointed at.
	palette := map[string]string{
		"primary": "#1A1F2E", "accent": "#E8311A",
		"background": "#111520", "surface": "#1E2436", "text": "#F0F2F7",
	}
	policy := defaultInkPolicy()
	css := buildLegibleInkDefaults("", palette, policy, zapNop())

	// The four grounds a visitor can actually see: the two declared ones and
	// each under the renderer's own translucent --section-surface overlay.
	grounds := []string{palette["background"], palette["surface"]}
	for _, g := range []string{palette["background"], palette["surface"]} {
		grounds = append(grounds, colour.CompositeOverGround("#ffffff", sectionSurfaceOverlayAlpha, g))
	}

	for _, name := range []string{"--color-primary-ink", "--color-accent-ink"} {
		emitted := inkVar(t, css, name)
		for _, g := range grounds {
			ratio, err := wcagContrastRatio(emitted, g)
			if err != nil {
				t.Fatalf("wcagContrastRatio(%s,%s): %v", emitted, g, err)
			}
			if ratio < policy.minRatio {
				t.Errorf("%s = %s measures %.3f:1 on ground %s, below the configured "+
					"target %.1f. The policy is not reaching legibleInkFor, or a ground "+
					"is missing from pageGrounds.", name, emitted, ratio, g, policy.minRatio)
			}
		}
		// The discriminating half: at 5.0 these must be STRICTLY better than
		// the 4.5 floor, or the threshold is not actually being applied and
		// this test would pass against the old binary.
		worst := 99.0
		for _, g := range grounds {
			if r, err := wcagContrastRatio(emitted, g); err == nil && r < worst {
				worst = r
			}
		}
		if worst < inkFloorContrast {
			t.Errorf("%s worst ground ratio %.3f is below even the AA floor %.1f", name, worst, inkFloorContrast)
		}
	}
}

// TestInkPolicy_MinRatioActuallyReachesTheDerivation closes a gap that every
// other test in this file walks straight past.
//
// defaultInkPolicy().minRatio == inkMinContrast, so a builder that IGNORED the
// policy and read the package constant directly would satisfy the pinned table,
// the composited-ground test and the clamp tests alike. Deleting the wiring
// would leave the suite green — which is precisely the mutation-passes shape
// this lane documented on 2026-08-14 (M5: removing the compositing loop left
// the package green because the assertion lived one layer down).
//
// So this drives the builder at a target it does NOT get by default, and
// requires the OUTPUT to move. Both expected values were computed by the
// separate implementation described above.
func TestInkPolicy_MinRatioActuallyReachesTheDerivation(t *testing.T) {
	palette := map[string]string{
		"primary": "#1A1F2E", "accent": "#E8311A",
		"background": "#111520", "surface": "#1E2436", "text": "#F0F2F7",
	}
	cases := []struct {
		ratio          float64
		wantPrimaryInk string
	}{
		{inkFloorContrast, "#8a97bd"}, // what shipped on 2026-08-14 morning
		{5.0, "#94a0c2"},              // what the owner's ruling produces
	}
	seen := map[string]bool{}
	for _, c := range cases {
		css := buildLegibleInkDefaults("", palette, inkPolicy{resolved: true, enabled: true, minRatio: c.ratio}, zapNop())
		got := inkVar(t, css, "--color-primary-ink")
		if !strings.EqualFold(got, c.wantPrimaryInk) {
			t.Errorf("at minRatio %.2f, --color-primary-ink = %s, want %s", c.ratio, got, c.wantPrimaryInk)
		}
		seen[strings.ToLower(got)] = true
	}
	if len(seen) < 2 {
		t.Error("both targets produced the SAME hex, so this test cannot tell whether " +
			"policy.minRatio is wired through at all. buildLegibleInkDefaults is probably " +
			"reading the inkMinContrast constant instead of its policy argument.")
	}
}

// TestInkPolicy_DisabledEmitsNothing is the kill-switch's whole contract:
// emitting NOTHING is what returns every consumer to the raw palette colour
// through its own var() fallback. Anything else — an empty block, a comment —
// would leave the name defined and the rollback incomplete.
func TestInkPolicy_DisabledEmitsNothing(t *testing.T) {
	palette := map[string]string{
		"primary": "#1A1F2E", "accent": "#E8311A",
		"background": "#111520", "surface": "#1E2436", "text": "#F0F2F7",
	}
	if css := buildLegibleInkDefaults("", palette, inkPolicy{resolved: true, enabled: false, minRatio: inkMinContrast}, zapNop()); css != "" {
		t.Errorf("disabled policy still emitted CSS — the kill-switch does not kill:\n%s", css)
	}
	// Control: the SAME palette must produce a block when enabled, or the test
	// above passes for the wrong reason.
	if css := buildLegibleInkDefaults("", palette, defaultInkPolicy(), zapNop()); css == "" {
		t.Fatal("control failed: enabled policy emitted nothing, so the disabled case proves nothing")
	}
}

func TestResolveInkPolicy_DefaultsToEnabledAtTheShippedTarget(t *testing.T) {
	got := resolveInkPolicy(nil, "any-site", zapNop())
	if !got.enabled {
		t.Error("nil config disabled the ink companions; the default must be ON")
	}
	if got.minRatio != inkMinContrast {
		t.Errorf("nil config gave minRatio %.2f, want the shipped default %.2f", got.minRatio, inkMinContrast)
	}
}

func TestResolveInkPolicy_GlobalAndPerSiteKillSwitches(t *testing.T) {
	const siteA = "11111111-1111-1111-1111-111111111111"
	const siteB = "22222222-2222-2222-2222-222222222222"

	cases := []struct {
		name        string
		config      map[string]interface{}
		siteID      string
		wantEnabled bool
	}{
		{"global off", map[string]interface{}{"legible_ink_enabled": false}, siteA, false},
		{"global on is explicit no-op", map[string]interface{}{"legible_ink_enabled": true}, siteA, true},
		{"per-site off matches", map[string]interface{}{"legible_ink_disabled_site_ids": []interface{}{siteA}}, siteA, false},
		{"per-site off does not match another site", map[string]interface{}{"legible_ink_disabled_site_ids": []interface{}{siteA}}, siteB, true},

		// Hand-edited config under incident pressure. Both of these are the
		// operator having done the right thing; neither may silently fail.
		{"per-site tolerates whitespace", map[string]interface{}{"legible_ink_disabled_site_ids": []interface{}{"  " + siteA + " "}}, siteA, false},
		{"per-site tolerates case", map[string]interface{}{"legible_ink_disabled_site_ids": []interface{}{strings.ToUpper(siteA)}}, siteA, false},

		// An unknown site id must not disable anybody — the safe direction is
		// "the fix stays on".
		{"empty siteID cannot be disabled", map[string]interface{}{"legible_ink_disabled_site_ids": []interface{}{siteA}}, "", true},
		{"mistyped key is ignored", map[string]interface{}{"legible_ink_enabled": "false"}, siteA, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveInkPolicy(c.config, c.siteID, zapNop()); got.enabled != c.wantEnabled {
				t.Errorf("enabled = %v, want %v", got.enabled, c.wantEnabled)
			}
		})
	}
}

// TestResolveInkPolicy_RefusesToConfigureBelowTheAccessibilityFloor is the one
// that matters most in this file. A kill-switch that can be configured to 1.0
// is a supported way to ship illegible text and call it a rollback — strictly
// worse than the bug it was built to undo.
func TestResolveInkPolicy_RefusesToConfigureBelowTheAccessibilityFloor(t *testing.T) {
	cases := []struct {
		requested float64
		want      float64
	}{
		{1.0, inkFloorContrast},
		{4.4999, inkFloorContrast},
		{inkFloorContrast, inkFloorContrast},
		{5.5, 5.5},
		{inkCeilingContrast, inkCeilingContrast},
		{21.0, inkCeilingContrast}, // 21 is max possible contrast; still clamped
		{-3.0, inkFloorContrast},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%.4f", c.requested), func(t *testing.T) {
			got := resolveInkPolicy(map[string]interface{}{"legible_ink_min_contrast": c.requested}, "site", zapNop())
			if got.minRatio != c.want {
				t.Errorf("requested %.4f -> %.4f, want %.4f", c.requested, got.minRatio, c.want)
			}
			if got.minRatio < inkFloorContrast {
				t.Errorf("resolved target %.4f is BELOW the WCAG AA floor %.2f — config must never "+
					"be able to ship sub-AA text", got.minRatio, inkFloorContrast)
			}
		})
	}
}

// A clamp that happens silently is an operator believing a number that is not
// in force. The warning is part of the contract.
func TestResolveInkPolicy_LogsWhenItClamps(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	resolveInkPolicy(map[string]interface{}{"legible_ink_min_contrast": 2.0}, "site", zap.New(core))
	if logs.FilterMessageSnippet("clamped").Len() == 0 {
		t.Error("a sub-floor contrast request was clamped silently; the operator would " +
			"believe 2.0 is in force")
	}
}

// TestPickInkOn_StillUsesTheAAFloorNotTheInkTarget guards a scope boundary that
// a single shared constant used to hide.
//
// pickInkOn serves --color-<x>-TEXT: ink ON a filled control. The 2026-08-14
// ruling raised the target for --color-<x>-INK: links and eyebrows on the page
// ground. Before that day both read one constant, so raising it would have
// silently retuned every button label in the fleet as a side effect of a
// decision about links. If someone re-merges the two constants, this fails.
func TestPickInkOn_StillUsesTheAAFloorNotTheInkTarget(t *testing.T) {
	if inkMinContrast == inkFloorContrast {
		t.Skip("targets currently coincide, so this test cannot discriminate")
	}
	// A palette whose text clears the AA floor on this fill but NOT the raised
	// ink target — the band that only exists because the two differ.
	fill := "#767676"
	palette := map[string]string{"text": "#ffffff", "background": "#000000", "text_muted": "#eeeeee"}

	ratio, err := wcagContrastRatio(palette["text"], fill)
	if err != nil {
		t.Fatalf("wcagContrastRatio: %v", err)
	}
	if ratio < inkFloorContrast || ratio >= inkMinContrast {
		t.Skipf("fixture no longer sits between the two constants (%.2f:1 against floor %.1f / target %.1f); "+
			"pick another fill", ratio, inkFloorContrast, inkMinContrast)
	}

	got, source := pickInkOn(fill, palette)
	if got != palette["text"] {
		t.Errorf("pickInkOn(%s) = %s (%s), want the palette text %s.\n"+
			"It measures %.2f:1 — above the AA floor %.1f, below the ink target %.1f. "+
			"Getting the fallback here means pickInkOn has been repointed at inkMinContrast, "+
			"which silently retunes every filled control from a ruling that was about links.",
			fill, got, source, palette["text"], ratio, inkFloorContrast, inkMinContrast)
	}
}

// TestInkPolicy_UnresolvedZeroValueFailsSafeAndLoud pins the fix for the one
// defect this change's council round actually found.
//
// `inkPolicy{}` is Go's free zero value and it is the DANGEROUS state:
// enabled=false, minRatio=0, i.e. "silently emit nothing". On a shared builder
// that is indistinguishable from a deliberate kill and raises no error, so a
// future call site that forgets to thread a policy would quietly restore
// pre-repair behaviour on everything it renders. Raised independently by the
// bug_historian and guardian seats (round 1 of d60aab29) as the
// "generic mechanism patched at one call site" shape.
//
// The contract: an unresolved policy is a CALLER BUG, so it must (a) still
// emit — failing safe, matching ChromeLinkPolicy's discipline of resolving
// toward the safe direction — and (b) log at Error, so it cannot hide.
func TestInkPolicy_UnresolvedZeroValueFailsSafeAndLoud(t *testing.T) {
	palette := map[string]string{
		"primary": "#1A1F2E", "accent": "#E8311A",
		"background": "#111520", "surface": "#1E2436", "text": "#F0F2F7",
	}

	core, logs := observer.New(zapcore.ErrorLevel)
	css := buildLegibleInkDefaults("", palette, inkPolicy{}, zap.New(core))

	if css == "" {
		t.Fatal("a bare inkPolicy{} emitted NOTHING — the zero value is being honoured as a kill. " +
			"A caller that forgets to resolve a policy would silently restore pre-repair behaviour.")
	}
	// It must fall back to the SHIPPED default, not to some other target.
	if got := inkVar(t, css, "--color-primary-ink"); !strings.EqualFold(got, "#94a0c2") {
		t.Errorf("unresolved policy emitted %s, want the default-policy value #94a0c2", got)
	}
	if logs.FilterMessageSnippet("UNRESOLVED").Len() == 0 {
		t.Error("the unresolved policy was repaired SILENTLY. Failing safe without saying so is how " +
			"a call site stays unwired indefinitely — the log line is half the contract.")
	}

	// Control: a deliberately disabled, RESOLVED policy must still emit nothing,
	// or the guard above has simply broken the kill-switch.
	if css := buildLegibleInkDefaults("", palette, inkPolicy{resolved: true, enabled: false}, zapNop()); css != "" {
		t.Errorf("a resolved+disabled policy still emitted CSS; the guard has broken the kill-switch:\n%s", css)
	}
}
