# NOTES — `bugs_open/199` render-seam envelope guard

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-05, session start — ownership check

`scripts/who-owns.py 199` returned **OWNED or recently active**, naming
`docs024_key_docs_latest/bugfix_190_content_data_envelope` (4 commits/14d). That is the lane
that **filed** 199, not one working it — it deferred this deliberately.

`who-owns` reads commits and is therefore lagging, so I also grepped live `.jsonl` transcripts
modified in the last 6h. Four hits: my own session, two that had only seen `199` in an `ls`
listing, and `fae13c79` (the `190` lane) with 24 hits — **whose last entry, at 10:08 UTC, is an
`away_summary` reading "Next action is bugs_open/199, the render-side follow-up, starting with
its unmeasured census."**

So the lane that filed it intended to take it next and then stopped. The owner opened this
session pointing at 199. Proceeding, and recording the risk here: if `fae13c79` resumes on its
stated next action there are two lanes on one bug.

## The census — and it decided the file

The bug filed its own claim `[UNMEASURED]` and said the census decides bug-or-note. Run live:

| class | components | active | live `page_components` rows |
|---|---|---|---|
| no/empty `input_schema` (gate skipped outright) | 56 | 35 | 44 |
| unrecognised dialect (`SchemaContentFields` `!ok`) | 8 | 1 | 1 |
| v2, no required `source:"llm"` field | 64 | 48 | 270 |
| **gate CAN speak** | 102 | 98 | 855 |

**315 / 1212 = 26% gate-blind.** And the one envelope still live in `content_data`
(gaswholesalers.com `how-pricing-works`, slot `pricing`, keys `{result,type}`, 1363-byte
`result`) sits on component `pricing` — **v2 dialect, no required llm field, gate-blind.** The
mechanism has fired against exactly this population.

## MISSTEP 1 — I classified 67 rows by a join that could never resolve

I wanted the historical envelope rows attributed to components:

```sql
FROM page_component_history h LEFT JOIN content_components c ON c.id = h.component_id
```

Answer came back clean: **67 rows, all `GATE-BLIND (no schema)`** — precisely the result that
would have supported the bug I was writing up.

It is meaningless. `page_component_history.component_id` is a FK to **`page_components(id)`**,
not the component library, and it is `ON DELETE SET NULL` while `save_page_sections` DELETEs and
re-INSERTs on every save. **67 of 67 are NULL.** Nothing joined. And my `CASE` tested
`c.input_schema IS NULL` *before* `c.id IS NULL`, so every failed join was relabelled as the
class I was counting.

Two errors, either sufficient: the wrong join key, and an arm order that hid it. The arm order
is the one to remember — it turned "I measured nothing" into a confident row of the expected
answer.

**Cheap check:** `\d page_component_history` before writing the join (the FK target and
`ON DELETE SET NULL` are both printed — I wrote the join from the column *name*), and put the
`IS NULL` arm **first** in any `CASE` over a `LEFT JOIN`. Logged in `WRONG_CALLS.md`; the
distilled version is now a `LANDMINES.md` entry.

## MISSTEP 2 — "no live agent uses render_component"

My first pass over `agent_definitions` for steps with `action='render_component'` returned
**zero rows**, which would have meant the whole bug was unreachable. Two compounding causes:

1. `default_config->'workflow'->'steps'` is an **object keyed by step name**, not an array —
   `jsonb_array_elements` errored with *"cannot extract elements from an object"* on the first
   attempt, and the fixed version still scanned only the top level;
2. the two `render_component` steps are **nested** inside
   `process_sections_loop.config.sub_workflow.steps`.

A `default_config::text LIKE '%render_component%'` scan found `page-content-writer` immediately.
**A structured path read that returns zero is not an absence — it is a shape assumption.**

## The correction that moved the fix

Reading `extractContentWithFallbacks` against the live `content_from` value, not in isolation:
the bug file (and the `016b` §10 row written from it) say the **"last resort"** branch returns
the envelope. It does not, on the live path. For `content_from: "generated_content.result"`:

- `pathsToTry[0]` = `generated_content.result` → the envelope's `result` **string** → the
  `map[string]interface{}` assertion fails → **the loop continues**;
- `pathsToTry[1]` = `generated_content` → the envelope **map** → passes the bare `len(m) > 0`
  test → **returned**.

`hasContentFields` guards only the last-resort branch and is never consulted. But that branch is
a **second** real leak: a superset envelope like finetuning's `{content,result,type}` passes
`hasContentFields` precisely on its `content` key. **Two doors, and the documented one is the
quieter.** That is what settled D5 — guard the caller, where one call covers both.

## Trigger rate: zero, and the query could have said otherwise

0 of 62 runs carrying a `generated_content` step output were envelope-shaped in the ~25h
`orchestration_states` window. The same query returns 111 for `compose_note`, 10 for
`generate_css`, 3 for `generate_tool_html` — so it discriminates.

Before recording the zero I checked the query could **see** the trigger at all:
`collected_data ? 'generated_content'` → 62 runs. Without that, a zero means "my filter cannot
see this", which is a different claim entirely.

So: **the door is open and currently unused.** Recorded in the bug file, the guard header, the
register entry and the `016b` row, because it changes how the post-roll check reads — a count of
zero is the *expected* result and proves nothing alone.

## False positives: zero, measured

114 live `render_context` maps in `orchestration_states`; **not one carries a `type` key at
all**. So `render_from_template` (whose `content_from` *is* `render_context`) cannot trip the
predicate. Every envelope-shaped object anywhere in live `collected_data` has keys exactly
`{result,type}` and is an LLM step output.

## MISSTEP 3 — I named a mutation, ran it, and it PASSED

Four tests, each naming the mutation that must break it. I ran all four rather than asserting
them. Three behaved. `TestRenderNonEnvelopeContentByteIdentical`, carrying *"predicate on the
presence of `type` alone"*, **passed under its own mutation.**

Cause: my seam's fast-exit is `if !isLLMTransportEnvelope(contentData) { return contentData, nil }`,
and the function it then calls — `normalizeContentDataEnvelope` — **opens by checking the same
predicate** and returns `(m, false, nil)` for a non-envelope. Weakening my local check changes
nothing. **Guards in series.** The local check is an optimisation and a log-suppression, not the
decision; byte identity is inherited from the storage normaliser's no-op contract.

Verified the real one: weakening the shared `isLLMTransportEnvelope` **does** fail the test.

I recorded the series relationship on the test rather than quietly relabelling the mutation to
one that fails — the relabel would have left a test whose stated guarantee was false for the
predicate a reader would actually go and weaken. Logged in `WRONG_CALLS.md`.

## Fable's plan corrected two of my assumptions, and I re-measured both rather than taking them

1. **The last-resort branch is not dead** — the superset case reaches it. Adopted; it is now the
   argument for caller-side placement.
2. **The save seam's identity paths are empty here.** I re-ran it myself: across every stored
   `page-content-writer` run (n=110 at my run; fable saw 115 — retention moved between us),
   `site_record.site_id` **0/110** and `current_page.name` **0/110**, against
   `input_data.site_id` **110/110** and a page-name union of **110/110**. Had I copied the save
   seam's paths the guard would have fired correctly and written an unattributable row every
   time. Now a `LANDMINES.md` entry.

## Blast radius, traced not assumed

`shouldContinueLoopOnError` (`loop_error_handler.go:70-90`) returns false unless the step config
carries `continue_on_error: true`. The live `page-content-writer` sets it on neither
`process_sections_loop` nor any substep. **So a refused render fails the whole page, not one
section.**

Judged proportionate: it is already the disposition for the 855 rows the existing gate covers,
and with `190` live a REFUSE-tier payload fails the run anyway at the save — later, and after
more LLM spend. Section-skip is a one-key config change, live immediately, and deliberately not
pre-empted.

## Not touched on purpose

`content_data_envelope_guard.go` — fable's plan had two comment-only edits there. Another
session had just renamed `require_sections_metadata` → `refuse_save_without_sections_metadata`
in it. A pathspec commit **cannot** exclude a same-file passenger, so I left the file alone and
put the second-seam documentation in my own file and the register entry instead. Checked before
committing: `git status` showed it clean by then (their change was already committed), so the
risk had passed — but the cost of avoiding it was two comments.

## Council verdict — APPROVED round 1, `dfb87f5e-6f01-42d4-8a01-6c59a4640c08`

"approved with 2 advisory objection(s) — none high-severity". 12 seats fired, 4 abstained.
Three of the objections were **checkable claims rather than opinions**, so I checked them
instead of banking the approval.

**`editquality` (medium) — "does the fast-exit predicate actually catch the SUPERSET shape?"**
Fair, and the sharpest objection of the round: it decides whether "one call covers both
branches" is true for the second leak, and **nothing at my seam pinned it**. It holds —
`isLLMTransportEnvelope` is signature-not-arity by deliberate design — but "it holds" is not a
test. Added `TestRenderGuardCatchesSupersetEnvelope` (both arms: an unrecoverable superset is
refused; a recoverable one decodes **and keeps the accreted sibling**), with the named mutation
`add a len(m)==2 arity test` **run and verified to break it**.

**`guardian` + `render_guardian` (medium) — "the single-consumer claim is asserted, not
proven; and is `RenderComponentAction` reachable from the scoped-rerender path, where a refusal
could be swallowed by its carry-stored-HTML bail-out?"** Both correct that I had argued this
from `agent_definitions` alone. Checked structurally, and it comes back stronger than I claimed:

- `grep -rn "RenderComponentAction"` returns **no Go call site at all** — every hit is a comment
  except `registry.go:940`, the `"render_component"` registry entry. So the action is reachable
  *only* by that action name, and `page-content-writer` is its only live user. The claim is now
  structural, not a census.
- `rerender_page_sections_action.go` **does not call it** — it calls `RenderTemplate` directly
  (`:485`). So the swallowed-refusal path the `render_guardian` describes does not exist for
  this guard. Their reasoning was right; the structure rules it out.

**`bug_historian` + `guardian` (medium) — no independent throttle on fail-loud.** *"If the
envelope class ever fires in volume this converts N silent blank sections into N failed page
builds with no per-run signal distinguishing 'guard working as intended' from 'guard now taking
down builds'."* This one I cannot close by checking, and I am not going to pretend otherwise.
It is disclosed, argued, and both seats flagged it explicitly **for a human**. Recorded in the
bug file's fix section and here; `continue_on_error` is a one-key, live-immediately workflow
config change if the owner wants it, and deliberately not shipped by this lane.

**`reuse_agent` (low) — a second log writer for one defect class.** Accepted as a tracked
follow-up; the seat's own note says the duplication was named with a concrete reason rather than
silently introduced, which is what it wanted. Unifying `writeRenderEnvelopeLog` and
`writeContentDataEnvelopeLog` into one seam-parameterised writer is a clean, separate change.

**`llm_reliability`, unprompted and worth keeping:** the envelope can itself be the downstream
symptom of a `max_tokens`-truncated reply whose JSON never closed. The refuse branch is a safe
default there, but this seam **cannot** tell "the model replied in prose" from "the model was
cut off" — that distinction lives at the `GenerateText`/`stop_reason` layer. Named as a
residual, not a gap here.

## Post-roll, 2026-08-05 20:47 UTC — LIVE on v1.0.1254, and the verification found something

Pod-grep on both `agent-chassis` replicas: all three added symbols present (2 each), refusal
text present, positive control `sanitizeSectionsContentData`=2, negative controls 0. Pods
started 20:41 UTC, nine hours after the commit, so the image cannot predate the fix.

The negative control is **weak by necessity** — the change is purely additive, so there is no
removed string to grep for at 0. The strongest available substitute is a **test-only symbol**
(`TestRenderGuardCatchesSupersetEnvelope` → 0), which at least proves the grep discriminates on
binary content rather than always returning 0. Recording that rather than claiming a control I
do not have; the `190` lane hit the same limit and said so too.

### The verification landmine fired, and checking it is what made "live" true

`kubectl get pods -l app=agent-chassis` returns 2 pods. **41 pods in the namespace run the
agent-chassis image, and 34 of them were still on `v1.0.1252` — no guard.** My own memory note
("`-l app=<subsystem>` may be the WRONG SERVICE — one image, every label") is the only reason I
looked past the label.

Two candidate conclusions, opposite in consequence:

1. *the release fragmented the fleet and the fix is live on 7 pods of 41* — alarming, wrong;
2. *those pods cannot reach the code* — correct, but it needed proving, not assuming.

`ownerReferences[0].kind` on all 34 is **`Job`**, not `Deployment` — per-work-item pods pinned
to whatever tag existed when they were spawned, so `redeploy-agents` was right not to restart
them and they age out. Then the question that actually decides it: of their five agent types
(`content-feed-orchestrator`, `feed-ingester`, `feed-triage`, `model-directory-publisher`,
`vet-practice-verifier`), can any invoke `render_component`? All five return **false**, against
a positive control of `page-content-writer` = **true**. Combined with `RenderComponentAction`
having no Go call site at all, the action name is the only route and its one live consumer runs
on the two guarded replicas.

**So: live everywhere it can be reached.** That is a stronger statement than "both replicas
have it", and it is the one the label-scoped check could not have made.

### The gap I am closing the file WITH, not around

**The production no-op path is unexercised.** No `page-content-writer` run since the roll — the
estate is idle for that handler (nothing since 15:00, no pending `site_work_items`), so
legitimate content passing through byte-identical has not happened on a live page. It is
covered by mutation-verified action-level tests that drive `RenderComponentAction` itself, but
a test is not a page.

I considered driving a run to close it and decided against: a `page-content-writer` dispatch
rewrites a real page's sections on a live site, which is an outward-facing change and
disproportionate to the residual risk on a path this well covered by tests. The check is left in
the bug file and the register for whoever builds the next page. **Closing on the stated bar
(fixed AND live) with the gap named — not pretending it was proven end to end.**

### No second SUMMARY today, deliberately

The cadence rule says a summary is written when the five headings would genuinely differ, and
warns against a shelf of near-identical files. This morning's `SUMMARY_2026-08-05` already named
"after the roll, verify and close" as where we were going. Completing a step that was already
named is not a new understanding, so the read-out would repeat with one status line changed.
Recorded here so the absence reads as a decision rather than an omission.
