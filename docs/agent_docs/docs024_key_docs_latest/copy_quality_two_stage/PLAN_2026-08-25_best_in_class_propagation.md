# PLAN 2026-08-25 — making "best in class" travel, and wiring research to serve it

**The instruction** (owner, 2026-08-25, ruling 4): *"Somewhere the mission for every site we build
must be to make it the best in class … For that to happen we can research the latest findings, the
latest products, the most trusted product reviews and anything else that will make us best in
class."* Plus, on this plan: *"yes please plan it, if we put it into the golang file would that
break things or make it better?"*

## 1. What exists today `[MEASURED 2026-08-25]`

- The mission text EXISTS in exactly one prompt: `domain-research-classifier`'s **"Build standard"**
  block — *"Aim for best-in-class quality in this site's field. The bar is not 'competent template'
  but 'stands comparison with the strongest sites in this vertical'…"*. It fires once, at domain
  classification.
- **0 of 51 sites' current specs carry the phrase.** The planner, writer, designers and the
  experience loop never see it.
- Research machinery exists but is BIRTH-ONLY: the classifier's web research produces
  `vertical_landscape` (homegarden's is 16,562 chars); `vertical-exemplar-researcher` exists
  (43 calls/30d); the writer has a `{{.research_result}}` hook. The submit script records the gap
  itself: *"Best-in-class design research (a design-research sub-agent + best-in-class mission
  default) is a separate TODO and is not part of this trigger."* Nothing refreshes research after
  birth, and nothing routes it toward "what would make this site best in class".

## 2. The Go question, answered

**What the two homes mean.** Text in a Go file is compiled into the binary: it cannot drift, it is
code-reviewed, and it is identical everywhere — but changing one word means commit → image build →
fleet roll, and on this estate a Go change is inert until that roll. Text in the DB is live within
seconds of a migration applying, and it is what our prompt-audit tooling scans — but each copy is a
separate row that can drift (the house voice existed in eight drifting places before CQ-022).

**The rule this platform already proved:** put the MECHANISM in Go and the WORDS in one DB row.
That is exactly the house voice's design (`platform/voicestyle/voicestyle.go` + the single
`voice_style_block` row, injected wherever a template writes `{{.voice_style}}`): Go guarantees
one source and one injection path; the DB keeps the words editable in a minute with no roll, under
council review via migrations.

**Applied here:** putting the best-in-class text INTO a Go file would not break anything — but it
would make the wrong thing better. It optimises for immutability, when what this mission needs is
to reach five different agents from ONE source and to be TUNED cheaply (the house voice needed
three wordings in six weeks; this will too). Worse, prose in Go is where our own audit tooling sees
least (population D of the census is an upper-bound guess for exactly that reason). So:
**a `build_standard_block` carrier row beside `voice_style_block`, injected as
`{{.build_standard}}` through the same (generalised) voicestyle mechanism.** The one-time Go
change is the small injection generalisation, rolled once; every wording change after that is a
live migration. Go-resident text stays the right choice only for invariants that must never be
runtime-editable (output contracts, schema rules) — the opposite of a mission statement.

## 3. Where the standard must reach (the propagation map)

| surface | how | why |
|---|---|---|
| `build-site-planner` / `content-gap-planner` | `{{.build_standard}}` opt-in | plans decide what a best-in-class site HAS |
| `visual-designer` family | same | his 08-25 review: "it hasn't done its job" |
| site birth (`strategy` aspect) | classifier writes a `benchmark` key: who the strongest sites in this vertical are (it already researches them) + what beating them requires | makes the standard site-SPECIFIC, not a slogan |
| writer | indirectly — through briefs/strategy, not raw | the writer's job is a section; the standard shapes what it is asked FOR |
| experience loop | reads the strategy benchmark | his proposed "happy user" agent maps here |

## 4. Research that serves it (the never-built TODO)

Phase R1 — inventory first (register rule: reuse before building): what `vertical_landscape`
already holds per site, what `vertical-exemplar-researcher` produces, what the writer's
`research_result` hook is fed by, and where `seed_content_sources` routes. Phase R2 — a periodic
research refresh per site (latest findings, products, trusted reviews in the site's vertical) that
writes into `evidence_base`/content sources with dates, so facts arrive through the register
rather than as prose. Phase R3 — route it: the planner's gap pass and the experience loop read the
refreshed benchmark and file work ("the strongest sites carry X; we do not"). Costing and cadence
decided after R1 sizes what exists.

## 5. The stakes split (ruling 4's first half)

Finance acknowledges uncertainty; grass seed does not — the onus is what we DO know and what the
research says. Implementation: a per-site stakes marker (finance/health = high; editorial/hobby =
low) set at classification, read by brief seeding; HIGH keeps the acknowledge-uncertainty trust
device, LOW replaces it with evidence-forward framing. Then a one-time sweep of current briefs'
trust lines (homegarden's *"Explicit honesty about uncertainty is itself a trust signal"* is the
wrong device for that site under this ruling). Ships as: classifier addition + brief-seed rule +
sweep migration, each council-reviewed.

## 6. Order of work

1. Carrier row + injection generalisation (Go once) → planner/designer opt-in migrations.
2. Classifier writes `strategy.benchmark` per site at birth.
3. R1 research inventory → R2 refresh → R3 routing.
4. Stakes marker + brief trust-line sweep.
Each step lands with its own canary and the CQ-032 scanner run over any new prompt text.
