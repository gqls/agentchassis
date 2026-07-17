// FILE: platform/orchestration/datahelpers/voicetells_test.go
//
// Benchmark corpus from SPEC_voice_tells_check §5 — every TRIP case is real
// copy that shipped on leopardessconsulting.co.uk; every PASS case is the
// owner-approved v2 register that replaced it. V6 guards the design rule that
// the scanner must never reward errors or slang.

package datahelpers

import "testing"

const testGateJSON = `{
  "voice_gate": {
    "enabled": true,
    "expect_contractions": true,
    "banned_phrases": [
      {"pattern": "\\bproduction[- ]grade\\b", "reason": "site ban: production-grade unless stack named"},
      {"pattern": "\\bAI[- ]powered\\b", "reason": "site ban: unqualified"}
    ]
  }
}`

func mustGate(t *testing.T) *VoiceGate {
	t.Helper()
	g, err := ParseVoiceGate([]byte(testGateJSON))
	if err != nil {
		t.Fatalf("ParseVoiceGate: %v", err)
	}
	if g == nil {
		t.Fatal("gate unexpectedly nil (opt-in parse failed)")
	}
	return g
}

func checksIn(fs []VoiceFinding) map[string]int {
	m := map[string]int{}
	for _, f := range fs {
		m[f.Check]++
	}
	return m
}

// V1 — services hero pre-fix: triad + em-dash + 40-word sentence. Must TRIP.
func TestV1ServicesHeroDense(t *testing.T) {
	g := mustGate(t)
	html := `<section><p>Your LLM integration works in staging. The gap between that and a system that holds under real load — with observability, fault isolation, cost controls, and human oversight — is an architecture problem, and these are the engagements we run to close it, on Kubernetes, Kafka, and Postgres, at the speed your roadmap requires.</p></section>`
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	c := checksIn(fs)
	if c["long_sentences"] == 0 {
		t.Errorf("V1: expected long_sentences trip, got %v", c)
	}
	if c["em_dash_density"] == 0 {
		t.Errorf("V1: expected em_dash_density trip (2 dashes in ~55 words), got %v", c)
	}
}

// V2 — old index title register: banned phrase must TRIP via site ban.
func TestV2ProductionGradeBanned(t *testing.T) {
	g := mustGate(t)
	html := `<section><h1>Production-Grade Multi-Agent AI Systems for UK Engineering Teams</h1></section>`
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	if checksIn(fs)["banned_phrase"] == 0 {
		t.Errorf("V2: expected banned_phrase for production-grade, got %v", fs)
	}
}

// V3 — v1-dense homepage hero: honest but dense; flourish + density signals.
func TestV3DenseHonestHero(t *testing.T) {
	g := mustGate(t)
	html := `<section><p>Most of what we build is unglamorous, and that is the point. A pipeline that checks scraped business records against Companies House, and stops to ask a person when it is genuinely unsure, a system that reads across news sources and scores what is worth trusting, and a website that keeps itself current are the shape of the work. Each one runs without anybody watching it, and every decision it made is written down where you can read it back afterwards, which matters more than any demonstration we could stage for you.</p></section>`
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	if len(fs) == 0 {
		t.Errorf("V3: dense register should produce at least one finding")
	}
}

// V4 — v2 homepage hero. Must PASS clean.
func TestV4PlainV2HeroPasses(t *testing.T) {
	g := mustGate(t)
	html := `<section><p>We build systems that take over repetitive work. Each one has a clear job. It knows when to ask a person for help, and it writes down every decision it makes. When it isn't sure, it stops and asks. Nothing happens in a black box.</p></section>`
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	if len(fs) != 0 {
		t.Errorf("V4: v2 hero should pass clean, got %+v", fs)
	}
}

// V5 — who-we-help v2 cards. Must PASS clean.
func TestV5PlainCardsPass(t *testing.T) {
	g := mustGate(t)
	html := `<section><h3>A list that has to be checked against another list</h3><p>Records against a register, invoices against orders, one system against another. The rules are stateable, the volume is real, and a person doing it by hand gets slower and less careful as the pile grows.</p><h3>A report someone assembles by hand every month</h3><p>The steps are the same every time. Nobody chose this as a job; it accreted. A system can run it on schedule and write down exactly what it did, and the person gets their two days back.</p></section>`
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	if len(fs) != 0 {
		t.Errorf("V5: v2 cards should pass clean, got %+v", fs)
	}
}

// V6 — slang/typos must NOT be rewarded: a sloppy page with tells still trips,
// and a clean page does not trip MORE for being grammatical. We assert the
// scanner has no signal that fires on correctness itself.
func TestV6ErrorsNotRewarded(t *testing.T) {
	g := mustGate(t)
	sloppy := `<p>aight so we leverage cutting-edge AI-powered agents, u know what i mean lol</p>`
	fs := g.ScanVoice(ExtractAssertionText(sloppy), false)
	c := checksIn(fs)
	if c["banned_phrase"] < 2 {
		t.Errorf("V6: slang must not mask banned phrases (leverage, cutting-edge, AI-powered); got %v", c)
	}
}

// V7 — long-form thresholds: an essay paragraph that would trip landing-copy
// density passes under longForm.
func TestV7LongFormRelaxed(t *testing.T) {
	g := mustGate(t)
	html := `<article><p>Supervisor architectures fail in production for reasons that rarely show up in a notebook, and the interesting failures are the quiet ones — a worker that stalls without erroring, a queue that backs up behind a slow consumer, a retry that duplicates work because the acknowledgement was lost. The fix is rarely clever. It is usually a boundary drawn more carefully: one job per worker, one owner per state transition, and an explicit decision about what happens when a human needs to intervene.</p></article>`
	strict := g.ScanVoice(ExtractAssertionText(html), false)
	relaxed := g.ScanVoice(ExtractAssertionText(html), true)
	if len(relaxed) > len(strict) {
		t.Errorf("V7: long-form must never be stricter (strict=%d relaxed=%d)", len(strict), len(relaxed))
	}
}

// Strawman shapes trip: the services-failure CTA ("Not a demo environment.
// Not a proof of concept.") and the comma form.
func TestStrawmanShapes(t *testing.T) {
	g := mustGate(t)
	html := `<p>Not a demo environment. Not a proof of concept. If the work fits, we can show you a running example.</p><p>This is not just a framework, but a system that ships.</p>`
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	if checksIn(fs)["strawman"] < 2 {
		t.Errorf("expected both strawman shapes to trip, got %+v", fs)
	}
}

// Opt-in contract: a voice spec without voice_gate parses to nil gate.
func TestOptInContract(t *testing.T) {
	g, err := ParseVoiceGate([]byte(`{"tone": "plain"}`))
	if err != nil || g != nil {
		t.Fatalf("expected nil gate for spec without voice_gate, got %v err %v", g, err)
	}
}

// Stiff register: fifteen-plus contraction-free sentences trip no_contractions.
func TestNoContractions(t *testing.T) {
	g := mustGate(t)
	sent := "We would rather scope the work honestly than promise it. "
	html := "<p>" + repeatN(sent, 16) + "</p>"
	fs := g.ScanVoice(ExtractAssertionText(html), false)
	if checksIn(fs)["no_contractions"] == 0 {
		t.Errorf("expected no_contractions trip, got %+v", fs)
	}
}

func repeatN(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
