# 035 G6 DRAFT — stored patterns of judged arrangements

**Written:** 2026-09-02, by Fable, at the owner's request — the third ask for
this model on this design, after two lanes declined the section so that it would
not be written by a passing hand. Every figure marked `[MEASURED 2026-09-02]`
was taken by this session against the live DB or the tree today; carried-forward
figures keep their original date and attribution.

**Status: DRAFT for integration.** `035_FEATURE_component_hierarchy.md` is
under active edit by another session; this file edits nothing there. §8 below
says where each piece lands. Nothing here re-scopes P1–P5 or touches anything
above 035 §2.

**The steer this designs** (owner, 2026-08-31, verbatim, recorded in 035's
CANDIDATE G6 block):

> *"There is also a mechanism to have components contain other components that
> we should use and we could store patterns of such combinations after they've
> been through component and experience loops etc."*

---

## 1. The goal, in §2's shape

- **G6 — Stored patterns of judged arrangements.** An arrangement that has
  earned its verdicts — every child family scored by the component loop, the
  whole ruled on by the experience council at the moment of promotion — can be
  stored as a unit, versioned, and applied to a new page whole. A pattern
  stores the ARRANGEMENT (ordered slots, concrete variant bindings), never
  content: application generates fresh content per child through D5.
  *Test: applying a stored pattern to a fresh page yields a parent whose
  ordered slot list and per-child variant bindings diff EMPTY against the
  pattern's spec, with all child `content_data` newly generated and no hand
  edits; and every reference in the pattern's provenance block resolves — the
  council ruling row exists, each child family's `quality_score` is non-NULL
  and dated. A dangling reference refuses the promotion; nothing is stored.*

The rest of this file is the argument for each clause of that sentence, and
the honest statement of where it bends the steer.

## 2. The stored object — a promoted composite definition, not a new class

**A stored pattern is a `content_components` row**: a composite definition in
D3's exact vocabulary — `slots` in `input_schema`, `render_mode='composite'`
derived — whose slot plan is fully **concrete** (exact counts, no ranges;
concrete bindings, not family names), plus a promotion provenance block.

Why this home and not a new table:

- **The identity, versioning and review machinery already exist and are
  live.** `component_versions` snapshots every definition edit (363 rows,
  real producers `[MEASURED 2026-08-22, 035 §3]`); an instance pins a pattern
  version through D6's `pinned_component_version_id`; the library's selector
  columns already speak advisory scoping (`suitable_site_types`,
  `suitable_page_types` — present in the live schema
  `[MEASURED 2026-09-02, \d content_components]`).
- **A second home for arrangement vocabulary is D3's retired drift class one
  level up.** D3 retired `child_components` because two homes for one
  declaration is the drift this repo keeps paying for; a `composition_patterns`
  table would be the same mistake at table grain. Slot declarations live in
  `input_schema`, full stop — a pattern's are simply concrete.
- **The component loop's population includes it by construction.**
  `compute_component_quality`'s sweep mode selects `WHERE is_active = true`
  with no `render_mode` filter (`compute_component_quality.go:119-121`
  `[MEASURED 2026-09-02]`) — so the moment a pattern exists as an active row,
  the loop that the steer names can see it. No routing change buys this; the
  storage choice does.
- **It is where the estate already decided patterns go.** The concept
  register's "Site interrogation & pattern library" entry (category
  `adoption-pipeline`, status **aspirational**, legacy era) proposed minting
  extracted patterns as `content_components` rows. G6's harvest source is the
  opposite — our own judged pages, not scraped external ones — but the
  destination agrees, and the register entry must be related, not duplicated,
  when P6's register entry is written.

**Two small schema deltas, both RFC_022 shape** (opt-in; unsafe default OFF;
zero consumers at birth), both migrations and therefore council-scoped:

1. **`content_components.pattern_provenance` jsonb, nullable, default NULL.**
   NULL = not a pattern; non-NULL = this definition was promoted from a live
   arrangement, and the block holds the resolvable references §3 requires.
   Deliberately NOT inside `input_schema` (that block is generation input; a
   verdict record is not) and NOT `child_components` (stays retired).
2. **`created_from` gains CHECK value `'promoted'`.** The live constraint is
   `manual|generated|adopted|tool|forked` `[MEASURED 2026-09-02]`. Reusing
   `'adopted'` was considered and rejected: it means external-site adoption,
   and the discriminating value is what keeps every future census honest.

**Concrete binding needs one precise extension of the slot vocabulary.** A
schema slot binds a family (`"function": "prose-block"`); D6 variants are
sibling rows sharing a function (`forked_from` set — which is why the
by-function quality lookup filters `forked_from IS NULL`,
`compute_component_quality.go:112-114` `[MEASURED 2026-09-02]`), so a family
name cannot name a variant. A pattern slot therefore MAY carry
`"component_id": "<uuid>"` beside `function` — binding the definition, exactly
as D6's variant swap does. **The pattern binds the definition, not the
version**: `pattern_provenance` records the `component_versions` id each child
was judged at, but application follows the library by default and pins nothing
— pinning stays a per-instance, deliberate act (D6). `[INFERRED]` that
recording-without-pinning is the right default; it is flagged as an owner
decision in §7.

**A consequence to state, not hide: the library is estate-global.**
`content_components` has a UNIQUE constraint on `name` and no site or client
column `[MEASURED 2026-09-02]`. A promoted pattern is therefore visible to
every site by construction. Advisory scoping via the existing
`suitable_site_types`/`suitable_page_types` columns is the recommended
containment; hard per-site scoping would need schema the table does not have,
and is not proposed.

**What a pattern never stores: content.** No `content_data`, no imagery, no
prose. Content transfer between sites is the duplicate-content risk
`features_open/002` exists for, and it would gut G1 (each page's children are
generated for that page, against that page's G5 brief). The pattern is the
shape that earned its verdicts; the words are always new.

## 3. What "has been through the loops" means mechanically

This is the load-bearing clause, and the CANDIDATE block's measurement stands:
**neither loop emits a verdict at arrangement grain** (component loop scores
`content_components`, 126 of 381, zero of this feature's families; experience
council rules on features/experiences/decisions, 80 notes, 13 subjects, none
an arrangement, none editorial — all established 2026-08-26). The design does
not pretend otherwise. It gives the clause a **compositional** meaning plus a
**promotion ruling**, and says plainly where that bends the steer:

- **The component loop vouches for the PARTS, at the grain it already
  emits.** Promotion requires every child family named in the slot plan to
  carry a non-NULL `quality_score`, and the provenance block records each
  score with its `quality_checked_at` date (the dated-count rule applied to
  verdicts). The editorial families' current NULLs are a **dependency to
  discharge** — run the live loop over them once P2 builds them — not a
  redesign of the loop.
- **The experience loop vouches for the WHOLE, at the grain it already
  emits.** Promotion is a design decision, and decisions are precisely the
  object class the council rules on. One council round per promotion, subject
  = the pattern; the ruling's `doc_notes` row id goes in the provenance
  block. This extends the council's **subject-matter** (it has never ruled on
  anything editorial) but not its grain, machinery or roster.
- **Live service is evidence and is recorded with its demand control.** The
  provenance block names the source page, parent instance, dates served, and
  WHICH instruments looked at it and found nothing — a bare "no adverse
  findings" is a zero with no demand control and is worth nothing.
- **No third loop.** A bespoke arrangement-scoring loop would join §3's three
  dormant composition mechanisms the day it shipped: 035 §6.8 is explicit that
  a mechanism without a driver stays dormant, and an arrangement scorer has no
  driver until someone wants to store a pattern. Promotion-as-event carries
  its own driver — the verdict is produced exactly when it is wanted.

**The honest weakening, stated so the owner can overrule it:** the steer reads
as though the loops judge arrangements today. They do not, and this design
satisfies the clause by composition (parts scored at part grain) plus ruling
(whole judged as a decision) plus recorded live service — rather than by an
arrangement-grain instrument, because that instrument does not exist and
building one speculatively is the dormant-mechanism trap. The named
alternative — extend `compute_component_quality` with composite-aware
arrangement checks — remains open, can be added later **without changing the
stored object**, and one piece of it is owed regardless (next paragraph).

**One rubric interaction is owed whether or not the alternative is taken.**
The quality checker's template-variable extractor collects top-level idents —
`{{.slots.lead}}` parses to field `slots`
(`extractTemplateVariables`, `compute_component_quality.go:352+`
`[MEASURED 2026-09-02]`) — and the schema/template sync check compares
extracted idents against declared fields `[INFERRED at the comparison level;
measured at the extractor]`. A composite declaring D3's `slots` block would
therefore read as referencing an undeclared field and mis-score structurally.
Teaching the sync check that a declared `slots` block satisfies a `slots`
reference is a small, named change and belongs to whichever phase first
creates a composite definition (P1/P2) — G6 inherits it, it does not own it.

## 4. The apply path — D7, with the spec retrieved instead of generated

**Yes, D7's propose→apply→approve is the apply path**, exactly as 035's
CANDIDATE block anticipated. A pattern's concrete slot plan IS a composition
spec — D7 already commits to the spec vocabulary being D3's slot shape, so
"what the pattern stores" and "what the page becomes" diff mechanically, the
property D7 promises for freshly generated specs. Application is:

1. retrieve the pattern (a library read, not an agent call);
2. propose it against the target page as a D7 spec (the proposer may be a
   human or a design agent citing the pattern);
3. human review, per application — design D / bugs-126 unchanged: the
   promotion ruling vouched for the arrangement, not for its fit to this
   page's content and purpose;
4. on approval, the pipeline applies it as row operations and D5 generates
   each llm child fresh against the target page's G5 brief.

**No auto-apply authority ships with G6.** If applications of a proven
pattern turn out to be rubber-stamps, the measured approval rate is the
evidence an owner decision to loosen step 3 would rest on — measure first,
decide later, and the loosening would be its own opt-in change (2026-08-02 §2
discipline). §7 lists it.

## 5. Phase and dependencies — P6, after P4, not gated on P5

**G6 is P6.** Its dependencies, each already in §5's chain:

- **P2** — the generation fan-out is the application engine (step 4 above),
  and P2 building the editorial families is what makes their quality scores
  producible at all;
- **P3** — a pattern binds variants; the variant seam and swap guard must be
  live and proven before a binding means anything;
- **P4** — the spec vocabulary and the apply pipeline ARE the pattern
  machinery; P6 adds retrieval and promotion, not a second pipeline.
- **P2-exit addition (recommended, one line):** run
  `compute_component_quality` over the editorial families P2 creates, so the
  component-side provenance §3 requires exists before P6 opens.

**Not gated on P5.** The guides are not a consumer of nesting (established
2026-08-31, and scoped as that lane insisted: eleven guides, one site, judged
from rendered structure — evidence about the guides, not the editorial
corpus). The first harvest candidate is the P2 `insight-article` arrangement
after it has held on a live editorial page for real weeks — the same
"real weeks" bar §5 already sets for P5.

## 6. Falsifiers — the goal's, and the design's

- **Application fidelity** (the G-line test): arrangement diff EMPTY against
  the spec, content fresh, no hand edits.
- **Provenance resolvability**: promotion refuses, storing nothing, on any
  dangling reference — fail-closed, the D4/D6 idiom. A pattern row whose
  provenance cannot be walked back to its verdicts is the failure mode this
  goal exists to prevent (a library of unsupported snippets).
- **Transfer** — the deep premise on trial: apply the first pattern to three
  different pages. If each application needs hand edits to the spec to fit
  the page, arrangements are page-shaped, storage buys a museum, and G6
  collapses. Three is small and stated: it can detect only gross failure to
  transfer, not a subtle rate — say so when reporting it.
- **Gate discrimination, proven by mutation, not by approvals**: the
  promotion gate is not proven until it has REFUSED something. Submit one
  deliberate control — an arrangement the submitting lane believes should not
  be stored — and the council must refuse it. Five clean approvals in a row
  prove the gate unproven, not the patterns excellent.
- **Vocabulary sufficiency**: if storing a real arrangement needs fields
  D7's spec cannot express (so the pattern cannot round-trip
  page → spec → page), the shared-vocabulary claim is false; the spec grows
  or the design is wrong — found at P6's first harvest, cheaply.

## 7. Decisions the owner is owed

1. **The clause's meaning.** §3 satisfies "been through the loops"
   compositionally plus a promotion ruling. If the owner meant "an instrument
   scores the arrangement itself", that instrument must be commissioned
   (composite-aware checks in `compute_component_quality`) and P6 waits on
   it. Recommendation: compositional now, instrument later if wanted — the
   stored object does not change either way.
2. **Library scope.** Estate-global identity is forced by the live schema;
   advisory scoping via `suitable_site_types`/`suitable_page_types` is the
   recommended containment. If hard per-site scoping is wanted, that is new
   schema and should be said now, not discovered at P6.
3. **Version binding on application.** Recommended: record the judged
   versions in provenance, follow the library on application, pin nothing.
   The alternative (apply pins the judged versions) trades drift-with-the-
   library for exact-as-judged; it is one owner sentence either way.
4. **Auto-apply for proven patterns.** Not shipped; revisit on the measured
   approval rate after real applications.

## 8. Integration notes for the session editing 035

- §2: append the G6 bullet (§1 above) after G5; the CANDIDATE block collapses
  to the steer quote + a pointer at the design (or its measurements fold into
  the new section).
- §4: this design lands as **D9** (D8 is taken), touching D3 (slots
  vocabulary, `component_id` binding extension), D5 (application generates),
  D6 (record-not-pin), D7 (retrieval arm).
- §5: **P6** as §5 above, plus the one-line P2-exit addition.
- §9 gains: *do not store content in a pattern — arrangement only*; *do not
  create any `pattern_provenance` row before the P6 promotion path ships*
  (§6.6's image-before-config rule); *do not hand-write a provenance block —
  promotion writes it or it does not exist*.
- Register obligation at P6 ship: one entry, related to the legacy
  "Site interrogation & pattern library" aspiration (adoption-pipeline
  bucket), stating harvest-from-judged-pages as the difference.
