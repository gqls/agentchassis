// FILE: platform/orchestration/actions/check_tool_fabrication_test.go
//
// Tests for the bug-020 fabrication detector. The whole value of this gate is
// PRECISION: it must catch the vetcomparison-style invented directory while
// leaving legitimate games, calculators, and faithful data-backed recreations
// alone. Each negative case below is a real shape the fleet produces.

package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// Regression for the 2026-07-22 induced-fault find: the action must read the
// html_field config as a LITERAL path and resolve it ONCE. If it double-resolves
// (via ExtractActionInputs Strategy 0, which resolves dotted config values against
// collected_data), the path becomes the HTML *content*, the detector sees empty
// input, and returns "uninspectable" — silently no-op'ing on the fail-open detector
// or over-HOLDING every recreation on the fail-safe one. Caught live by the probe;
// this test fails against that bug and passes against the config-read fix.
func TestCheckToolFabricationAction_ReadsDottedConfigPath(t *testing.T) {
	collected := map[string]interface{}{
		"completeness_check": map[string]interface{}{"clean_html": vetcompFabrication},
	}
	params := ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: map[string]interface{}{"html_field": "completeness_check.clean_html"}},
	}
	out, err := CheckToolFabricationAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", out)
	}
	if m["fabricated"] != true {
		t.Fatalf("expected fabricated=true reading the dotted config path, got %v (tier=%v)", m["fabricated"], m["tier"])
	}
	if m["tier"] != "declaration" {
		t.Fatalf("expected tier=declaration (real detection ran), got %v — the action likely double-resolved html_field into content and saw empty input", m["tier"])
	}
}

// The actual fabrication that shipped to vetcomparison.uk (condensed but with
// every real tell: the confessing comment, Mulberry32, the fragment arrays,
// makePostcode, buildData — and no fetch, over an original that fetched a JSON).
const vetcompFabrication = `
<div class="tool-page">
  <div id="directory"></div>
  <script>
  (function(){
    // The original directory holds 2,100+ UK practices. For this recreation we
    // generate a large, realistic, deterministic dataset so search, filtering
    // and pagination behave exactly as they would against the real directory.
    const rng = (function(){ let a = 0x2f6e2b1; return function(){ a |= 0; a = a + 0x6D2B79F5 | 0; let t = Math.imul(a ^ a >>> 15, 1 | a); t = t + Math.imul(t ^ t >>> 7, 61 | t) ^ t; return ((t ^ t >>> 14) >>> 0) / 4294967296; }; })();
    const TOWNS = ['Abbey','Oakwood','Riverside','Kingsway'];
    const PREFIXES = ['North','South','Central'];
    const SUFFIXES = ['Veterinary Centre','Vets','Animal Hospital'];
    function makePostcode(){ const a = 'ABCDEFGHIJKLMNOPRSTUW'; return a[Math.floor(rng()*a.length)] + 'W' + Math.floor(rng()*9+1) + ' ' + Math.floor(rng()*9) + 'PL'; }
    function buildData(){ const out = []; for(let i=0;i<2100;i++){ out.push({ name: TOWNS[i%TOWNS.length] + ' ' + SUFFIXES[i%SUFFIXES.length], postcode: makePostcode() }); } return out; }
    const data = buildData();
    function render(){ /* … */ }
    render();
  })();
  </script>
</div>
<!-- tool-recreation-complete -->`

const vetcompOriginal = `
<div id="practice-search"></div>
<script>
  fetch('/data/vet-full-index.json').then(r => r.json()).then(rows => renderDirectory(rows));
</script>`

func TestDetect_VetcompFabrication_Gated(t *testing.T) {
	res := DetectToolFabrication(vetcompFabrication, vetcompOriginal, false)
	if !res.Fabricated {
		t.Fatalf("expected the vetcomparison fabrication to be gated; signals=%v", res.Signals)
	}
	// Tier A must fire first — the confessing comment and makePostcode are both present.
	if res.Tier != "declaration" {
		t.Fatalf("expected tier=declaration (comment + makePostcode), got %q; signals=%v", res.Tier, res.Signals)
	}
}

// Same fabrication with the confessing comment and makePostcode removed — it must
// STILL be caught, this time by Tier B (corpus + original-was-data-backed + no
// preserved fetch). This proves the gate does not depend on a lucky comment.
func TestDetect_CorpusWithoutDeclaration_CorroboratedGated(t *testing.T) {
	rec := `
<div class="tool-page"><div id="grid"></div>
<script>(function(){
  function mulberry32(a){ return function(){ a |= 0; a = a + 0x6D2B79F5 | 0; let t = Math.imul(a ^ a >>> 15, 1 | a); return ((t ^ t >>> 14) >>> 0)/4294967296; }; }
  const TOWNS=['Abbey','Oak']; const SUFFIXES=['Vets','Clinic'];
  function buildRecords(){ const r=mulberry32(42); const o=[]; for(let i=0;i<500;i++) o.push({name:TOWNS[i%2]+' '+SUFFIXES[i%2]}); return o; }
  const rows=buildRecords();
})();</script></div>`
	original := `<script>const xhr=new XMLHttpRequest(); xhr.open('GET','/data/vets.json');</script>`
	res := DetectToolFabrication(rec, original, false)
	if !res.Fabricated {
		t.Fatalf("expected corroborated corpus to be gated; signals=%v detail=%q", res.Signals, res.Detail)
	}
	if res.Tier != "corroborated_corpus" {
		t.Fatalf("expected tier=corroborated_corpus, got %q; signals=%v", res.Tier, res.Signals)
	}
}

// A legitimate DICE / loot game: uses a seeded PRNG for gameplay, has NO original
// data dependency, does not declare invented data. Must NOT be gated — this is
// the false positive the corroboration gate exists to prevent.
func TestDetect_LegitDiceGame_NotGated(t *testing.T) {
	rec := `
<div class="tool-page"><canvas id="board"></canvas>
<script>(function(){
  function mulberry32(a){ return function(){ a |= 0; a = a + 0x6D2B79F5 | 0; let t = Math.imul(a ^ a >>> 15, 1 | a); return ((t ^ t >>> 14) >>> 0)/4294967296; }; }
  const roll = mulberry32(Date.now());
  function rollDice(){ return 1 + Math.floor(roll()*6); }
  document.getElementById('board');
})();</script></div>`
	original := `<canvas></canvas><script>/* the original dice roller */ function rollDice(){ return 1+Math.floor(Math.random()*6); }</script>`
	res := DetectToolFabrication(rec, original, false)
	if res.Fabricated {
		t.Fatalf("legit dice game must NOT be gated; tier=%q signals=%v", res.Tier, res.Signals)
	}
}

// A plain CALCULATOR: pure arithmetic, no PRNG, no arrays, no data. Not gated.
func TestDetect_Calculator_NotGated(t *testing.T) {
	rec := `
<div class="tool-page">
  <input id="amount"><input id="rate">
  <script>(function(){
    function compute(){ const a=+document.getElementById('amount').value; const r=+document.getElementById('rate').value; return (a*r/100).toFixed(2); }
    document.getElementById('amount').addEventListener('input', compute);
  })();</script>
</div>`
	res := DetectToolFabrication(rec, `<script>function compute(){}</script>`, false)
	if res.Fabricated {
		t.Fatalf("calculator must NOT be gated; signals=%v", res.Signals)
	}
}

// A FAITHFUL recreation of the data-backed directory: it PRESERVES the fetch and
// invents nothing. Even though the original was data-backed, there is no corpus
// signature and a live fetch — must NOT be gated.
func TestDetect_FaithfulDataBackedRecreation_NotGated(t *testing.T) {
	rec := `
<div class="tool-page"><div id="results"></div>
<script>(function(){
  fetch('/data/vet-full-index.json').then(r=>r.json()).then(function(rows){
     window.__rows = rows; renderPage(rows);
  });
  function renderPage(rows){ /* region filter + pagination over rows */ }
})();</script></div>`
	res := DetectToolFabrication(rec, vetcompOriginal, true /* analysis flagged data source */)
	if res.Fabricated {
		t.Fatalf("faithful fetch-preserving recreation must NOT be gated; signals=%v", res.Signals)
	}
}

// The CORRECT empty-state behaviour the new prompt asks for: original was
// data-backed, recreation dropped the fetch but invents NOTHING (renders an empty
// state). No corpus signature → not gated (the model did the right thing).
func TestDetect_HonestEmptyState_NotGated(t *testing.T) {
	rec := `
<div class="tool-page"><div id="results">No data available yet.</div>
<script>(function(){ /* nothing to load without the directory source */ })();</script></div>`
	res := DetectToolFabrication(rec, vetcompOriginal, true)
	if res.Fabricated {
		t.Fatalf("honest empty state must NOT be gated; signals=%v", res.Signals)
	}
}

// A legitimate NAME-GENERATOR tool: fragment arrays are its whole purpose, and
// its original had them too (no fetch). Must NOT be gated — no declaration, no
// synthetic PII, and no data-backed corroboration.
func TestDetect_NameGeneratorTool_NotGated(t *testing.T) {
	rec := `
<div class="tool-page"><button id="gen">Generate</button><span id="out"></span>
<script>(function(){
  const ADJECTIVES=['Swift','Bright','Noble']; const NOUNS=['Fox','River','Peak'];
  function generateName(){ return ADJECTIVES[Math.floor(Math.random()*3)]+NOUNS[Math.floor(Math.random()*3)]; }
  document.getElementById('gen').addEventListener('click',()=>{document.getElementById('out').textContent=generateName();});
})();</script></div>`
	original := `<script>const ADJECTIVES=['Swift']; const NOUNS=['Fox']; function generateName(){}</script>`
	res := DetectToolFabrication(rec, original, false)
	if res.Fabricated {
		t.Fatalf("name-generator tool must NOT be gated; tier=%q signals=%v", res.Tier, res.Signals)
	}
}

// Synthetic PII generator INTRODUCED where the original had none — Tier A on its
// own, even without a confessing comment.
func TestDetect_IntroducedSyntheticPII_Gated(t *testing.T) {
	rec := `<script>function randomPostcode(){ return 'SW'+Math.floor(Math.random()*9)+' '+Math.floor(Math.random()*9)+'AB'; }
	  const list = Array.from({length:300}, () => ({ postcode: randomPostcode() }));</script>`
	original := `<script>fetch('/data/branches.json').then(r=>r.json());</script>`
	res := DetectToolFabrication(rec, original, false)
	if !res.Fabricated || res.Tier != "declaration" {
		t.Fatalf("introduced synthetic PII must be gated as declaration; got fabricated=%v tier=%q signals=%v",
			res.Fabricated, res.Tier, res.Signals)
	}
}

// Corpus signals present but the original was NOT data-backed (a purely
// procedural generative-art tool). Not corroborated → not gated, but the signals
// are surfaced for observability.
func TestDetect_UncorroboratedCorpus_ReportedNotGated(t *testing.T) {
	rec := `<script>function buildData(){ /* procedural terrain seed grid */ return []; }
	  function mulberry32(a){ return function(){ return 0.5; }; }</script>`
	res := DetectToolFabrication(rec, `<canvas></canvas><script>/* generative art, no data source */</script>`, false)
	if res.Fabricated {
		t.Fatalf("uncorroborated corpus must NOT be gated; signals=%v", res.Signals)
	}
	if len(res.Signals) == 0 {
		t.Fatalf("expected corpus signals to be surfaced for observability even when not gated")
	}
}

// FAIL-SAFE: an un-inspectable (empty/missing) recreation must be HELD for review,
// not silently passed to deploy — a safety gate that can't read its input must not
// default to "clean" (the missingkey=zero class bug 020 belongs to). This also
// makes a drifted field path fail loud rather than becoming a silent no-op.
func TestDetect_UninspectableEmpty_FailsSafe(t *testing.T) {
	res := DetectToolFabrication("", "", false)
	if !res.Fabricated {
		t.Fatalf("empty/uninspectable recreation must be HELD for review (fail-safe), not passed")
	}
	if res.Tier != "uninspectable" {
		t.Fatalf("expected tier=uninspectable, got %q", res.Tier)
	}
	// whitespace-only is equally uninspectable
	if !DetectToolFabrication("   \n\t ", "", false).Fabricated {
		t.Fatalf("whitespace-only recreation must be held for review")
	}
}

func TestDataSourceIsExternal(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want bool
	}{
		{"bool true", map[string]interface{}{"has_external_data": true}, true},
		{"bool false", map[string]interface{}{"has_external_data": false}, false},
		{"string true", map[string]interface{}{"has_external_data": "true"}, true},
		{"string false", map[string]interface{}{"has_external_data": "false"}, false},
		{"string false-prose", map[string]interface{}{"has_external_data": "false — purely computational"}, false},
		{"nonempty source", map[string]interface{}{"has_external_data": "", "source": "/data/x.json"}, true},
		{"empty source", map[string]interface{}{"source": "  "}, false},
		{"not a map", "nope", false},
	}
	for _, c := range cases {
		if got := dataSourceIsExternal(c.in); got != c.want {
			t.Errorf("%s: dataSourceIsExternal=%v want %v", c.name, got, c.want)
		}
	}
}
