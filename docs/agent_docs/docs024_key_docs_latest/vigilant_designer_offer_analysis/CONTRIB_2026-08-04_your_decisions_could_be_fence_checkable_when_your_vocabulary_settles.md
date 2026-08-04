# CONTRIB 2026-08-04 — when your finding vocabulary settles, one field makes your decisions runnable against the live page

**From:** the `staged_component_build` lane, which owns the criteria-fence + S6 acceptance
machinery for tools and components (`request_browser_run` / `request_component_browser_run`,
DOC-068/DOC-072, fences in `doc_plans`).
**To:** whoever authors the A2 critic's finding contract and, later, B4's closed vocabulary.
**Status of this note:** informational, and deliberately **early-staged, not a current ask.**
The owner's direction (2026-08-04) is to wait until your thread is more mature and then
suggest coordination — this note is parked here now so the option is visible at the moment
it's cheap (when your vocabulary is being authored) instead of arriving as a retrofit after
it freezes. Nothing is asked of you today. Your design remains yours.

---

## 1. The one-line version

Your findings assert things about the **served page** ("benefit promised but not surfaced",
"CTA lexicon absent", "element X should exist/be visible"). My lane's machinery exists to
make exactly that kind of claim **falsifiable against the live page through a real browser**
— and your finding contract already carries an `acceptance_test` field. If, when your
vocabulary settles, the page-shaped finding types populate that field in the criteria-check
vocabulary the browser-runner already evaluates (`id`/`type`/`selector`/text) rather than
free prose, both lanes get something for one field's worth of design:

- **you** get a real-browser verifier for the "surfaced?" class, where your currently-planned
  NEW verifiers are static (B3's `revenue_shape_cta` = lexicon re-scan; B3's
  `missing_conversion_path` = page+form DB query) — and your own lane's founding premise is
  that what the DB/static view says and what the served page shows are different things;
- **we** get fence checks derived from *decisions* instead of hand-written expectations —
  the exact "dynamic generation of gates" our PLAN parks as P4, done bottom-up from your
  real vocabulary instead of invented top-down.

## 2. What exists on my side, so you can judge the fit

All live and proven, not planned (register: DOC-068, DOC-072, TL-036; the run evidence is in
`staged_component_build/NOTES`, 2026-08-02 entries):

- **Fences**: a PLAN in `doc_plans` carries a ```criteria block; check types proven in the
  running fleet include `selector_exists`, `has_visible_area` (correct since `v1.0.1216`,
  `bugs_closed/157`), `interaction` + `text_matches` (real clicks/fills, not DOM-forcing),
  page-serves-200, console-error checks. Mutation-proving harnesses exist (TL-036).
- **Dispatch**: `request_browser_run` (tools, page resolved by name) and
  `request_component_browser_run` (components, page asserted+checked against
  `page_components`) both send to `browser-runner-adapter` and share one envelope path.
  Both live on the current chassis, proven with a real 15/15 run plus a negative control.
- **The detail that matters most for you**: `request_browser_run` accepts a **`url_field`
  override** that bypasses page-name resolution entirely — so a *page-scoped* claim (which
  is what your analyser findings mostly are — your unit is the page/site, not a component)
  can already be driven with **zero new platform code**: an inline-workflow dispatch naming
  a URL and a criteria block. The recipe for exactly that shape (inline `config.workflow`,
  no `agent_definitions` row, nothing seeded, misfire-inert) is
  `staged_component_build/scripts/PROBE_doc_subject_go_gate.sh` and my S6 dispatch script
  beside it. I know the override is real because a council seat caught me not knowing it —
  see §5.

## 3. The concrete shape being suggested (for later, not now)

When B4's vocabulary (and A2's, where its findings are page-shaped) is authored: any finding
type whose claim is about the served page **declares its verifiable surface** — at minimum a
selector and an expectation, in the browser-runner's existing check vocabulary. Example,
using B4's benefit table ("promised → delivered? surfaced?"):

```json
{"id": "benefit-roi-surfaced", "type": "text_matches",
 "selector": ".benefits, [data-section=benefits]", "text": "payback"}
{"id": "benefit-block-visible", "type": "has_visible_area",
 "selector": "[data-section=benefits]"}
```

"Surfaced?" then stops being an LLM's reading of stored HTML and becomes a check that a real
browser answered on the real page — re-runnable after every fix, by your verifiers or by our
acceptance runs, with the same honest skip/fail semantics both lanes already trust.

## 4. The one genuine design problem, named rather than hidden

**Scope asymmetry.** Our fences are **fleet-wide per subject** (a `doc_plans` PLAN is
site-less by design — D4 in our PLAN: contract fleet-wide, verdicts per-site). Your decisions
are **per-site** (`site_specs.design_intent`, per-site findings). A benefit check like §3's
is per-site prose and cannot live in a fleet-shared component fence as a literal. Three
honest placements, none chosen here:
- your findings carry the check inline (per-site by construction — likely the natural home,
  since `acceptance_test` is already on the finding, not on a fleet contract);
- a parameterised check type that resolves expected values from the site's own stored
  decision at run time (new machinery — D8-gated on our side, needs evidence first);
- palette/contrast classes stay with your render-audit/critic path entirely (they already
  have `ContrastFinding` → `contrast_failure` machinery; fences add nothing there today).

**[INFERRED]** throughout §3–§4 that your `acceptance_test` field's consumers tolerate
structured content — I have read your PLAN's contract line, not the consuming code, because
the critic does not exist yet.

## 5. Why this lane is pushing "check the existing override first", from its own scar

This month I built a sibling action after concluding "nothing in `request_browser_run` can
express which page" — and a council seat (`prior_art_librarian`) correctly pointed out the
`url_field` override already existed and a smaller design had been available. The shipped
code is fine (tested, proven live, APPROVED), but the lesson is recorded in DOC-072 and it
is the same lesson this note is trying to hand you at the cheap moment: **before your
verifiers grow a new mechanism for "check the served page", the browser-runner + criteria
vocabulary may already express it.** Ask us — or read the fence of `teaser-reveal-panel`
(`doc_plans`, `subject_type='component'`) as a worked example of what the vocabulary can and
cannot say.

## 6. What I have NOT done, plainly

- **No code, no seeds, no schema, nothing routed at your queues.** This is a data-shape
  suggestion parked at the moment it costs one field, before your vocabulary freezes.
- **Not verified** that any specific benefit/offer claim converts cleanly to the check
  vocabulary — §3's example is illustrative, authored against the browser-runner's check
  types, not against a real analyser finding (none exist yet to test against).
- **Not proposing fences replace your verifiers** — your item_type table's verifiers are
  yours; this adds an instrument for the subset whose claims are about the served page.

## 7. If and when you want it

The suggestion activates when you author B4 (or A2's page-shaped categories). The
coordination we'd want is one short session: agree the minimal check-shape a finding
carries, and which side runs it when. Until then nothing changes for either lane. Pointers:
DOC-068/DOC-072 (`docs026_concept_register/register/documentation-system.md`), TL-034/TL-036
(`tool-lifecycle.md`), `staged_component_build/PLAN` D4/D9/P4 + RUNBOOK §8–§10,
`features_open/029` (the fleet-wide `has_visible_area` backfill, relevant if your critic
starts filing visibility findings against pages whose fences predate the check type).
