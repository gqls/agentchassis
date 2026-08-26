// Pins JudgeCTALabel — the one definition of "does this button's copy name the
// page its destination is?", shared by the misdirected-CTA detector and the
// write-time audit (bugs_open/399).
//
// THE DISCONFIRMING CASE IS THE POINT. bugs_open/399 §Verification says it in
// terms: "a check that only ever sees matching pairs proves nothing, which is
// this estate's most-repeated lesson". On the live fleet most pairs DO agree,
// so a suite built from real rows passes by default and measures nothing. Every
// table row below therefore names the outcome it would have had if the
// judgement were broken.
package datahelpers

import "testing"

// dartsonlineCandidates is the bug's own site, reduced to the pages that decide
// its evidence row. Built through NewLabelMatchCandidate rather than by literal
// so the unexported token maps are populated the way production populates them.
func dartsonlineCandidates(t *testing.T) []LabelMatchCandidate {
	t.Helper()
	rows := []struct{ id, name, title, url, nav string }{
		{"1", "news-index", "Darts News | Darts Online", "/news/index.html", "News"},
		{"2", "brands-index", "All Brands | Darts Online", "/brands/index.html", "Brands"},
		{"3", "setup-builder", "Dart Setup Builder | Tools", "/tools/setup-builder/index.html", "Setup Builder"},
		{"4", "shaft-length", "Shaft Length Guide", "/guides/shaft-length.html", "Shaft Length"},
	}
	out := make([]LabelMatchCandidate, 0, len(rows))
	for _, r := range rows {
		c, ok := NewLabelMatchCandidate(r.id, r.name, r.title, r.url, false, r.nav)
		if !ok {
			t.Fatalf("candidate %q produced no distinctive tokens — fixture is wrong", r.name)
		}
		out = append(out, c)
	}
	return out
}

func TestJudgeCTALabel(t *testing.T) {
	pages := dartsonlineCandidates(t)

	cases := []struct {
		name        string
		label       string
		destination string
		onPage      string // page the button sits on
		want        CTALabelVerdict
		wantNamed   string // URL, when the verdict carries one
		wantAmbig   bool
		wantSilence CTALabelSilence
		why         string
	}{
		{
			// THE BUG'S OWN LIVE ROW, page_components.content_data on
			// dartsonline.com/news-index/hero, updated_at 2026-08-25 20:58:09Z:
			// cta_text "Catch up on this week's darts news" beside
			// cta_url "/brands/index.html". Re-minted 17 minutes AFTER the bug
			// was filed, still wrong.
			name: "the filed defect: copy names the news page, link goes to brands",
			// NB the button sits on news-index, so the label naming its own page
			// would be a self-link refusal — that is why this row uses the
			// SECONDARY slot's framing, on a page that is not the news index.
			label: "Catch up on this week's darts news", destination: "/brands/index.html",
			onPage: "shaft-length",
			want:   CTALabelContradicts, wantNamed: "/news/index.html",
			why: "if this returns Agrees or NoOpinion the audit is blind to the bug it was written for",
		},
		{
			name:  "the same copy pointing where it says",
			label: "Catch up on this week's darts news", destination: "/news/index.html",
			onPage: "shaft-length",
			want:   CTALabelAgrees, wantNamed: "/news/index.html",
			why: "a correct button must never be recorded — this is the false-positive guard",
		},
		{
			name:  "normalised paths are one destination, not a disagreement",
			label: "Catch up on this week's darts news", destination: "/news/",
			onPage: "shaft-length",
			want:   CTALabelAgrees, wantNamed: "/news/index.html",
			why: "/news/ and /news/index.html are the same page; convicting here would fire on every site using directory URLs",
		},
		{
			name:  "generic copy has no opinion",
			label: "Get Started", destination: "/brands/index.html",
			onPage: "shaft-length",
			want:   CTALabelNoOpinion, wantSilence: SilenceNamesNothing,
			why: "LabelTokens reduces this to nothing; a generic button is a content question, not a misdirect",
		},
		{
			name:  "copy naming the page it sits on names nothing",
			label: "Read the shaft length guide", destination: "/guides/shaft-length.html",
			onPage: "shaft-length",
			want:   CTALabelNoOpinion, wantSilence: SilenceNamesItsOwnPage,
			why: "bugs_open/308's self-link rule; without it this would read as Agrees and bless a button that links to its own page",
		},
		{
			name:  "copy matching no page at all",
			label: "Wibble the frobnicator", destination: "/brands/index.html",
			onPage: "shaft-length",
			want:   CTALabelNoOpinion, wantSilence: SilenceNamesNothing,
			why: "no candidate shares a token; the matcher must be silent, not convict",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := JudgeCTALabel(tc.label, tc.destination, pages, tc.onPage, "")
			if got.Verdict != tc.want {
				t.Errorf("verdict = %v, want %v\n  why it matters: %s", got.Verdict, tc.want, tc.why)
			}
			if tc.wantNamed != "" && got.Named.URL != tc.wantNamed {
				t.Errorf("named %q, want %q", got.Named.URL, tc.wantNamed)
			}
			if tc.want == CTALabelNoOpinion && got.Named.URL != "" {
				t.Errorf("NoOpinion carried a named page %q — callers may not act on it", got.Named.URL)
			}
			if got.Silence != tc.wantSilence {
				t.Errorf("silence = %v, want %v — the reason a verdict is silent is the seam a later "+
					"destination-KIND check hangs on (391 lane, 2026-08-26); collapsing the reasons "+
					"puts 95 of 186 live mismatches back into undifferentiated residue", got.Silence, tc.wantSilence)
			}
			if got.Ambiguous() != tc.wantAmbig {
				t.Errorf("ambiguous = %v, want %v", got.Ambiguous(), tc.wantAmbig)
			}
		})
	}
}

// TestJudgeCTALabelRefusesAnAmbiguousTie pins RFC_047's owner ruling
// (2026-08-23): when the copy names two pages equally well the matcher REFUSES
// rather than guessing, and the refusal is reported as Ambiguous so a record
// can say "this button is undecidable" without anyone acting on it.
//
// MUTATION: make JudgeCTALabel return Contradicts on !ok. This fails.
func TestJudgeCTALabelRefusesAnAmbiguousTie(t *testing.T) {
	// Two pages tied on every ranking key that carries signal: same identity
	// overlap, same total overlap, same interactivity. Only alphabetical order
	// separates them, and RFC_047 says that is not a reason to rewrite a button.
	a, _ := NewLabelMatchCandidate("1", "alpha-report", "Annual Report", "/alpha.html", false, "")
	b, _ := NewLabelMatchCandidate("2", "beta-report", "Annual Report", "/beta.html", false, "")

	got := JudgeCTALabel("Read the Annual Report", "/somewhere-else.html",
		[]LabelMatchCandidate{a, b}, "index", "")

	if got.Verdict != CTALabelNoOpinion {
		t.Fatalf("verdict = %v, want NoOpinion — an alphabetical tie must never convict a button", got.Verdict)
	}
	if got.Silence != SilenceAmbiguous {
		t.Errorf("silence = %v, want SilenceAmbiguous", got.Silence)
	}
	if !got.Ambiguous() {
		t.Error("Ambiguous = false; the tie refusal must be distinguishable from plain silence, " +
			"or 'this button is undecidable' reaches nobody (RFC_047 §10's stated gap)")
	}
	if got.Named.URL != "" {
		t.Errorf("a refused match carried a named page %q", got.Named.URL)
	}
}

// TestJudgeCTALabelSurvivesTheHyphenFamily pins the failure that
// bugfix_203/CALIBRATION_2026-08-11 measured: on gaswholesalers.com all NINE
// already-correct CTA labels flipped to the wrong tool because "Break-Even"
// tokenises differently from "breakeven". That calibration is the standing
// evidence against judging these pairs by raw token comparison, and it is why
// this predicate delegates to the ranked matcher instead of comparing the label
// to *_target_title.
//
// The bar here is deliberately modest and honest: a correct button must not be
// CONVICTED by a tokenisation artefact. Silence is an acceptable outcome;
// Contradicts is not.
func TestJudgeCTALabelSurvivesTheHyphenFamily(t *testing.T) {
	c, ok := NewLabelMatchCandidate("1", "tool-breakeven-volume-calculator",
		"Breakeven Volume Calculator", "/tools/tool-breakeven-volume-calculator.html", true, "Breakeven Calculator")
	if !ok {
		t.Fatal("fixture produced no tokens")
	}
	got := JudgeCTALabel("Open the Break-Even Volume Calculator",
		"/tools/tool-breakeven-volume-calculator.html", []LabelMatchCandidate{c}, "index", "")

	if got.Verdict == CTALabelContradicts {
		t.Errorf("a correct button was convicted by hyphenation alone (named %q) — "+
			"this is the 2026-08-11 calibration's nine-CTA flip returning", got.Named.URL)
	}
}

// TestJudgeCTALabelIsBlindToTheLabelLockedDefect is a test that PASSES AND IS
// WRONG, on purpose, so that nobody later reads this mechanism as covering more
// than it does.
//
// bugs_open/391: the resolver picks a destination by nav_order alone, then
// stampCTADestinationGuidance tells the writer to name that destination, and
// the writer complies. Copy and destination then AGREE while the button is
// pointed at an off-topic page. Measured 2026-08-25: 16 of 17 resolver-minted
// fields on the password-entropy family were in exactly this state, including
// all three buttons the owner reported.
//
// Agreement between two framework-written strings is evidence of CONSISTENCY,
// never of CORRECTNESS. No comparison at this seam can see it — only a ranking
// fix (bugs_open/391) or a copy pass (cta_target_content_pass) reaches it.
func TestJudgeCTALabelIsBlindToTheLabelLockedDefect(t *testing.T) {
	entropy, _ := NewLabelMatchCandidate("1", "password-entropy",
		"Password Strength Physics", "/tools/password-entropy.html", true, "Password Entropy")
	contact, _ := NewLabelMatchCandidate("2", "contact", "Contact Us", "/contact.html", false, "Contact")

	// The framework wrote BOTH sides: it picked password-entropy, then had the
	// writer name it. On a consultancy site whose real CTA should be "book a
	// call", this button is a defect.
	got := JudgeCTALabel("Try the Password Strength Physics tool",
		"/tools/password-entropy.html", []LabelMatchCandidate{entropy, contact}, "our-approach", "")

	if got.Verdict != CTALabelAgrees {
		t.Fatalf("verdict = %v, want Agrees — this fixture exists to document that the "+
			"label-locked defect PASSES this judgement", got.Verdict)
	}
	// Deliberately no further assertion: the point of this test is the comment.
}
