# HANDOFF 2026-08-19 — bugs_open/260 renderer half: designed and evidenced, NOT YET CODED

**Read this first, then `PLAN_2026-08-19_260_render_fallback.md` (the design, decided) and
`NOTES_260_render_fallback.md` (the evidence, with controls).**

## State in one paragraph

The bug is **valid, unclaimed by anyone else, and accelerating**. Root cause was already proven
by the filing lane; what this lane added is the evidence that makes the fix safe to make, three
corrections to the bug file, and a thirteenth render seam nobody had named. **No code has been
written. The tree is clean of my work** — everything committed is docs. Next actions are: council
submission, then implementation.

⚠ **A chassis roll landed 2026-08-19 and it does NOT contain this fix.** Nothing here is shipped.
The seam was re-verified unchanged at HEAD after that roll (`component_library.go:965`, fallback
still at `:1032-1057`, both signatures still error-free).

## What is committed (all docs, all narrow-pathspec, no passengers)

| file | what |
|---|---|
| `bugs_open/260_…md` | headline corrected; **§13** (today's census + measurements), **§13d** (candidate 2 unblocked), **§13g** (the 13th seam), and **§5 corrected** |
| `…/bugfix_260_render_fallback/PLAN_2026-08-19_…md` | **the design — start here for implementation** |
| `…/bugfix_260_render_fallback/NOTES_260_render_fallback.md` | evidence, harnesses, controls, missteps |
| `…/bugfix_260_render_fallback/README_where_we_are.md` | owner-facing plain-prose log |
| `…/copy_quality_two_stage/CONTRIB_2026-08-19_…md` | consumer notice (owner ruling 07-29 §3) |
| `…/LANDMINES.md` | the ungated-path entry; synced + verifier armed |
| `…/WRONG_CALLS.md` | my misstep (item TYPE names the trigger, not the event) |

Probe harnesses are **committed** at `probes/` in this directory — `parseprobe.go`,
`execprobe.go`, `contactprobe.go`, with `probes/README.md` carrying the two `psql` dumps they
need, the `kubectl exec` truncation flake, and why each probe's control is load-bearing. The
JSON dumps themselves are not committed (~13MB); regenerate them. If these earn a permanent
home they belong in `cmd/component-render-check/` rather than in a lane directory.

## The evidence that decides the design — do not re-derive, but DO re-date

All `[MEASURED 2026-08-19]`, each with a control that could have come out otherwise:

- **0 of 251 active templates fail to Parse.** Controls: unclosed `{{if}}` must fail; valid
  nested must pass. **So every occurrence is an EXECUTE error, never a parse error.**
- **0 of 1,778 stored sections fail to Execute** against their own `content_data`. Controls: the
  bug's own A/B pair. Faithful without a `RenderContext` replica because `contextToInterfaceMap`
  merges `ContentData` at the **top level** and `missingkey=zero` makes absent site fields safe.
- **0 of 253 active components use the fallback's dialect** (no `{{#`, no `{{nav_items_html}}`,
  no `{{quick_links_html}}`).
- ⇒ **Deleting the fallback changes the behaviour of nothing that currently works.**
- Census: **26 events · 7 domains · 25 work items · 08-11 → 08-18 23:36Z**, 24 parked at
  `needs_human_review`. Re-checked after the roll: **no new events**.
- Stored damage **0 of 1,789** page_components and **0 of 72** site_components; 1 benign brace
  row (the prompt-library one).

**Re-run the census before the submission if more than a day has passed** — it moved 11→26 in
three days. Query is in NOTES §"The census has moved".

## Three corrections this lane made — they matter to the submission

1. **§5's "the page-BUILD path has no schema gate at all" is FALSE at HEAD.**
   `missingRequiredLLMFields` has **two** callers: `rerender_page_sections_action.go:396` and
   **`v3_site_actions.go:2252`**, the latter inside the build render step. The build path has a
   **presence** gate; it lacks the **type** half — which is exactly why a present-but-mistyped
   field sails through, consistent with all 26 events. **Do not submit the old grounding; it is
   the sentence a reviewer checks first.**
2. **Candidate 2 is no longer inert.** The legacy `properties` dialect is **extinct (0)**, not 4;
   `mechanism-flow` itself now carries the house `fields` dialect. Of the **110** exposed
   components, **107 carry a `fields` schema**. Acute set: **14 llm-authored `array` fields, all
   declaring `items`**.
3. **A 13th render seam** — `RenderTemplateWithMap` (`rerender_pages_actions.go:782`), no FuncMap,
   no `missingkey=zero`, returns `""` on error, and its caller `ReplaceAllString`s the **live
   contact-info block** with it. So an error **deletes** the section. `[MEASURED]` latent today
   (1 active `contact-info` component, renders clean, both controls fired) but one ordinary
   `{{safe}}`/`{{default}}` edit away. **Fix it in the same change** (return input HTML
   unchanged) — see PLAN.

Line numbers in the bug file have drifted (`:2179`→`:2273`, `:333`→`:396`), and
`render_content_envelope_guard.go:13` carries a stale citation the other way. **Re-cite at commit
time; code comments are not a safe source here.**

## NEXT ACTIONS, in order

### 1. Council submission (do this first, or alongside the commit)

Rationale is **drafted and committed**: `RATIONALE_draft_for_council_submission.md` in this
directory — copy it into the submission JSON. Schema is `{rationale, submitter, plan:{summary, edits:[{file, symbol, operation,
rationale, sketch}]}}` plus `grounded_in` evidence quotes; worked example is
`bugfix_277_required_fields_repair/submission_277_required_fields_router.json`.

The `edits` array (≤8) comes straight from PLAN's three edit sets. **Include in the submission:**
- the **rejected alternative** (keeping `RenderTemplate(…) string` returning `""`) and why the
  compile-breaking signature is the feature;
- the **RFC_022 consumer enumeration** for `refuse_mistyped_llm_fields` — *run the query*;
  asserting an opt-in field has no live consumer without it **is itself the objection**;
- the coverage-honesty caveat (**75 of 253 actives declare no schema at all**).

```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
Save `SUBMISSION_CORR`. **Budget ~30 minutes, not 2** — a missing orchestration row is latency,
not a dropped dispatch; find the run by payload, not by the printed id.

### 2. Implement (PLAN has the full table)

Order that keeps the tree compiling in the smallest steps: seam signature + fallback deletion
first (it will break every caller — that is intended), then the 13 call sites, then
`ContentTypeViolations`, then tests.

**Do not skip:** `items` recursion in the checker is **mandatory** (the live case is a *nested*
violation — a top-level check misses the very instance that motivated this).

### 3. Register entry IN THE SAME COMMIT

Ordering-exemption condition 2. One entry in
`docs026_concept_register/register/styling-render-pipeline.md`: the seam's new guarantee ("a
component render either executed or errored; there is no third state"), the retired fallback
dialect, the opt-in key, and the coverage caveat. Update the index count; drop any matching line
from `102_coverage_ratchet.txt`. (`pattern-check` already flagged this lane as register-blind.)

### 4. Build, roll, verify

Commit → bump `IMAGE_TAG` → `make build-agent-chassis` (builds from committed HEAD) → roll.
**Ask the binary what it is running, per service, not the fleet.** Then the arming migration
(`_HOLD`, image first then seed).

### 5. Tell the lanes that are waiting

- **`loanzy_uk_example_site`** — will run a **clean greenfield build** after the roll and report
  either way, including if it still fires. This is the whole-route after-test. Ping when rolled.
- **`portfolio_positioning`** — holding remortgagecalculator.uk **locked and reproducing** as a
  stable specimen; will re-arm their four items on request.
- **`copy_quality_two_stage`** — already notified; they own the writer half and asked nothing
  blocking, but two questions are open in the CONTRIB (do they want a machine-readable error
  shape for a repair loop; and candidate 2 is now viable for them).

## Success criterion — state it this way or it will be misread

**The build fails EARLY, naming `branches` — not "the page builds."** The 24 parked items still
hold mistyped content; making them build is the *writer* half's job (`copy_quality_two_stage`).
What this change buys is that the failure is honest, immediate, names the field, and cannot reach
a live page through the two ungated paths.

## Traps this lane already walked into or past

- **`validate_content` protects the page-BUILD path only.** The section-editor path writes
  `rendered_html` to a live page with no gate, and its `if rendered == ""` guard **cannot fire**
  (the fallback returns mangled HTML, never empty). LANDMINE appended.
- **The 20-blocker count is a regex cap** (`FindAllString(html, 10)`), not a measurement — never
  fingerprint a component from it. My own token-set table inherits the same ceiling; it reads as
  *consistent with* `mechanism-flow`, not proof.
- **A brace literal is not a leak.** Tool pages legitimately carry `{{ }}`; one of the 26 events
  is exactly that. The design deliberately contains **no output brace-scan**, and the acceptance
  test needs a good tool page as a positive control that must **pass**.
- **A `090` on `component_library.go` returns bundles and no verdict** — the file is ~94KB, over
  the limit, and that looks exactly like a run still in progress. Do not spend a round on it.
- **`git mv` + pathspec commit ships a copy** — if 260 ever moves to `bugs_closed/`, name **both**
  paths and verify with `git ls-tree -r --name-only HEAD`.
