# 178 — a "add a link to X" work item regenerates the WHOLE section and silently loses most of its prose; the item reports `complete`

**Filed 2026-08-02** (session "bugfix 19"), immediately after the triage in
`sql_for_agents/286` released the item that caused it. **Status: prevention
LIVE + APPROVED; handler root cause CONFIRMED 2026-08-03** (see the final
update) — `content_rewrite` items never set `spec.mode`, so
`load_existing_content`'s adoption-only gate no-ops and the writer gets the
item's guidance text with no existing prose to edit. Effect measured and
certain on one page, `[UNVERIFIED]` on three others with the same signature
(now down to three — the relojistas row is explained, see below).

**Nothing is lost permanently — the prior content is in `page_component_history`.**
That is what makes this safe to leave open rather than hot-fix.

## What happened

Work item `93f2a3b7`, `item_type = content_rewrite`, summary:

> **"Add Gripper Safety Factor Calculator tool reference to how-to-specify-a-gripper page"**

It completed successfully in ~2 minutes (`status=complete`, `attempt_count=0`,
`error` NULL) and **it did add the link, correctly** — one contextual anchor to
`/tools/gripper-safety-factor-calculator/index.html` in each of the page's three
slots, with sensible surrounding prose.

It also **rewrote `generic-text-block` from scratch**, cutting it from 7
paragraphs to 4 and changing its heading:

```
content_data length   4439 -> 1806   (-59%)
heading  "Gripper Specification: A Systematic Approach"
      -> "Defining the Specification Parameters"
```

The HTML is well-formed and closes cleanly, so **this is not the `bugs_closed/012`
truncation signature** (a cut completion saved as a fragment). It is a shorter
*regeneration*. Substantive content that existed before and does not now:

- workpiece-first methodology (envelope dimensions, mass, surface condition,
  fragility threshold in Newtons, tolerance ranges);
- kinematic requirements (grip force at contact vs actuator force, stroke range,
  repeatability, with the worked ±0.05 mm / 150 mm-stroke counter-examples);
- cycle-time vs actuation-technology trade-off (pneumatic vs electric, compressor
  infrastructure, programmable force profiles);
- integration parameters (**ISO 9409-1** flange pattern, payload budget, I/O
  count, wrist force-torque sensing, collaborative-cell TCP mass limits);
- the closing synthesis paragraph.

The page is titled *"How to Specify a Gripper | Engineer's Reference"*. The lost
paragraphs were the reference.

## Why this is worse than it looks

The item's own summary says **add a reference**. Nothing in it asks for the
section to be rewritten. So the blast radius of a `content_rewrite` item is not
what its summary describes, and the difference is invisible downstream: the item
reports `complete`, the validator passes, `build_status` stays `deployed`, and no
record anywhere says content was dropped. This is the `WDS-004` family — *"we
finished" treated as "the work is right"* — in a path that 056 did not cover.

**Distinguish from the neighbours before quoting this file:**
- `bugs_closed/012` — tool-improver truncating a component into a fragment.
  Different: that is a CUT completion; this HTML is complete and well-formed.
- `bugs_closed/056` — regeneration dropping content that tripped a validation
  blocker. Different: nothing was blocked here; the rewrite passed validation.
  056 is closed and its fix does not address this.

## Fleet-wide signature `[PARTLY UNVERIFIED]`

Pages whose `content_data` shrank >25% against their most recent
`page_component_history` snapshot, last 7 days:

```
domain               url                                   before   now   slots   pct
fundamentallyai.com  /tools/review-council-simulator.html   35344   2900   3->3   -92
vonc.com             /about.html                            16030   8015  12->6   -50
relojistas.com       /glosario/index.html                    5899   3083   3->2   -48
vetcomparison.uk     /about.html                            13138   6803   4->4   -48
robot-hands.com      /how-to-specify-a-gripper.html          5795   3323   3->3   -43
gamesdesign.co.uk    /tools/bayesian-ranking.html            7926   5751   4->4   -27
```

**Read the slot column — it discriminates.** `vonc.com/about.html` (12→6) is
almost certainly the legitimate duplicate-component removal from
`bugs_closed/156`, and `relojistas` lost a slot too (> **CORRECTED 2026-08-03:**
explained — a deliberate back-out by its owning lane, see the final update). The four with an **unchanged
slot count and a large content loss** are the ones matching this bug's signature.
Only `robot-hands.com` is CONFIRMED (I read both versions); the other three are
`[UNVERIFIED]` — same shape, cause not established, do not cite them as instances
without reading their before/after.

**Explicitly NOT an instance:** `bugs_closed/154`'s witnessed `tool-improver` run
the same morning. Its component *grew* (`component_versions` 7944 → 9158 →
10520), so that verification stands.

## Recovery (the exact rows)

```sql
-- pre-write snapshot, captured by the writer itself at 10:41:22.93696Z
SELECT content_data FROM page_component_history
WHERE id::text LIKE 'ecb4b420%';        -- the 4439-char generic-text-block

SELECT id, source, length(content_data::text), created_at
FROM page_component_history
WHERE page_id='5a385981-c2fd-4edb-bc4d-927b93177281' ORDER BY created_at DESC;
```

**Restoring is an owner decision, not a mechanical one**, because the naive
restore also removes the tool link the item was legitimately raised to add. The
merge — old prose plus the new anchor — is an editorial act on a live customer
page. Not done by the filing session for that reason.

## Fix candidates, ordered by what closes the door

1. **A link-insertion item should insert a link.** If `content_rewrite` is being
   used as the carrier for "add a crosslink", the crosslink case wants a handler
   that edits rather than regenerates. Makes the loss unrepresentable.
2. **Refuse a regeneration that shrinks a section past a floor, and record it.**
   The `prune_floor.go` / CTXA-025 machinery from `bugs_closed/135` is the
   existing pattern for exactly this shape — a destructive operation that must
   prove it saw the corpus. Reuse before building.
3. **Make the loss visible even when allowed**: the writer already snapshots to
   `page_component_history` immediately before overwriting, so the delta is
   computable at write time and is currently thrown away. Emitting it would have
   surfaced all four fleet cases on the day they happened.
4. Sweep/restore the affected pages. Cleanup, not a fix; needs 1–3 first or it
   recurs.

## How to verify a fix

Raise a crosslink item against a page with a long prose section and assert the
section's `content_data` length is unchanged apart from the inserted anchor:

```sql
SELECT length(content_data::text) FROM page_components
WHERE page_id=<target> AND slot_name='generic-text-block';
-- expect: prior length + ~90 chars, NOT a wholesale replacement
```

Do **not** verify by the work item reaching `complete` — it did that here, and
that is the whole problem.

---

## UPDATE 2026-08-02 (same session, owner-directed) — RESTORED + guard COMMITTED (inert until roll)

**Restores — `sql_for_agents/287`, applied, every byte from `page_component_history` by row id:**
- robot-hands `generic-text-block` 1806→**4655** = the 4,439 reference PLUS the
  tool anchor merged into the workpiece paragraph (both truths kept).
- fundamentallyai tool slot `content_data` NULL→**32,444** (f321055b). ⚠ its live
  render was the only good copy — **no rerender queued for it, deliberately**;
  do not rerender that page until the restored shape is checked vs the renderer.
- vetcomparison `faq` 2197→**7206**, `about-content` 1882→**3043**. Stated
  trade-off: three small same-day edits discarded; a re-broken CTA re-detects,
  lost prose had no detector.
- **NOT restored**: gamesdesign (−27% was a legitimate rewrite — verified: old
  blob was a context dump, new fields purposeful); vonc (156 md5-dedup);
  relojistas glosario (DefinedTermSet slot DELETED and absent site-wide, but
  history records no slot_name — snapshot `b0e119a4`, 2,816 chars, awaits this
  bug's fix or an owner slot-name decision). (> **CORRECTED 2026-08-03:** no
  decision needed — deliberate back-out, do not restore; see the final update.)
- 2 rerender items `restore_287:%` queued; verify at the artefact (ISO 9409-1 +
  the anchor both present in robot-hands' rendered slot; vet faq ~3×).

**Prevention — fix candidate 2 SHIPPED (commit `2da3e08e5`), INERT until the
next image roll:** per-slot shrink floor in `SavePageSectionsAction`
(`save_sections_shrink_guard.go`). The page-total guard's blindness was
measured twice over (25% wipe threshold; page altitude). New rule: a
same-named prose slot (≥500 stripped chars) keeping <50% of its text refuses
the WHOLE save, fails closed, emits a refusal work item. Config
`section_shrink_floor` (0 disables, clamped ≤0.95). 9-case table test on the
real numbers. Council `e64f8576` (Council-Submitted trailer; verdict pending).
Pod-grep marker after the roll: `strings /app/agent-chassis | grep -c
"SECTION SHRINK"` (expect ≥1) with positive control `"CONTENT REGRESSION
BLOCKED"` (pre-existing).

**Root cause at the HANDLER is still undiagnosed** — the guard stops the
damage; nothing yet explains why a link-insertion item regenerates the whole
section. Candidates 1 (edit-not-regenerate) and 3 (emit the delta) remain open.

## UPDATE 2026-08-03 — guard LIVE, PROVEN by induction, council APPROVED; one tracked deferral

- **Prevention half is done.** The guard shipped, went live (v1.0.1233, survived
  the 1234 roll, pod-grep both replicas), and was **proven by a live induction**
  on dartsonline with the prediction recorded first: refused twice, orchestration
  FAILED honestly (not masked), refusal item emitted+deduped, zero bytes written.
  Council `e64f8576`: round 1 REVISE → round 2 **APPROVED** (2026-08-02 23:11Z,
  4 advisory objections, none high). Locked slots excluded (`5f00dcba9`); the
  refusal's queue wording made its own (`77b58fd4d`, council `98aa9103`
  **APPROVED** 23:26Z; + advisory `0913d5754`). **All of it LIVE in v1.0.1238,
  pod-proven both replicas 2026-08-03.**
  Evidence: `docs024_key_docs_latest/bugfix_154_work_item_routing_columns/NOTES_…`.
- **TRACKED DEFERRAL (the council's ask, 4 seats):** `save_page_sections` now
  carries THREE bespoke content-loss floors — page-total wipe, completeness/drop,
  per-slot shrink — each with its own key, threshold and blind spot. The
  architecture seat: "the deferral itself should be tracked so it doesn't recur a
  fourth time." **The tracking is THIS entry: if you are about to add a FOURTH
  floor to this chokepoint, stop — that is the trigger for the unified
  content-loss detector design, as its own submission, not another rider.**
- **Still open here (unchanged):** root cause at the handler (why does a
  link-insertion item regenerate the whole section — run 090); candidates 1
  (edit-not-regenerate) and 3 (emit the delta); relojistas' deleted
  DefinedTermSet slot (> **CORRECTED 2026-08-03:** resolved, not open — see
  below); other rendered_html writers unguarded (named above).

## UPDATE 2026-08-03 (154 lane) — root-cause 090 DISPATCHED; relojistas slot RESOLVED (deliberate back-out, do not restore)

**The 090 run for the handler root cause completed** (5 iterations, final
UNVERIFIABLE — intake `needs_diagnosis:178-crosslink-regenerates-whole-section`,
RUN correlation `aece2920-f85a-46e2-a53f-235a4b6e9ab1`) and named the exact
config it needed but didn't have in its bundle. Reading that config directly
CONFIRMED the mechanism — see the new section immediately below. Dispatch
record in `bugfix_154_work_item_routing_columns/NOTES_…` 2026-08-03.

**The handler root cause is now CONFIRMED**, closing the gap the 090 run
identified but could not reach in 5 iterations (it stopped UNVERIFIABLE,
naming exactly the missing piece: *"the page-build-handler writer/
content-generation step definition (absent from this bundle)"*). Read that
config directly from live `agent_definitions` after the run terminated —
first-hand verification substituting for a 6th automated pass, per the
2026-07-31 ruling's escape hatch, stated here plainly:

- `page-build-handler`'s `load_existing_content` step is a strict gate:
  `load_existing_content_action.go:64-69` — `mode := inputs.Get("mode"); if
  mode != "recreate" { ... return {"has_existing": false, "reason":
  "not_recreate"} }`. Its own doc comment: *"For non-adoption pages (no mode:
  recreate), returns empty — no-op."*
- `93f2a3b7.spec` carries **no `mode` key at all** (verified against the live
  row), and `create_tool_cross_link_items.go` **never sets one** — grepped,
  zero hits. So for every item this emitter raises, the gate is closed.
- `call_content_writer`'s `input_mapping` passes `existing_content?` (the
  output of the gated step above — empty here) and `current_page:
  page_record`. `load_page_record`'s own description: *"Load page record from
  DB — sections, title, page_type"* — no prose content; that lives in
  `page_components.content_data`, which this workflow never loads for the
  writer at all outside the gate.
- **So `page-content-writer` receives the item's guidance text
  ("weave a natural reference … into the existing content … do NOT add a new
  section") and NO existing content to weave it into**, for any
  `content_rewrite` item on an already-built page. It must fabricate a
  replacement section that satisfies the instruction's shape — which is
  exactly the observed defect: a well-formed, correctly-linked, shorter,
  restructured section with a changed heading.
- **This generalises beyond cross-links**: the mechanism is the emitter never
  setting `mode`, not anything specific to `create_tool_cross_link_items.go`.
  Any `content_rewrite` item from any source, against any page that already
  has content, hits the same gate. `apply_gap_plan_action.go`'s
  `content_rewrite` emission (:243) is the other live source and was not
  checked for a `mode` key — flagging for whoever picks up the fix.
- **Correction to fix candidate 2's premise**: `load_existing_content`, even
  when `mode="recreate"` IS set, sources `research_results` — the original
  adoption-crawl snapshot — never the page's CURRENT `page_components`. So
  setting `mode="recreate"` on a crosslink item would not fix this; it would
  feed the writer stale, pre-platform-edit content instead of none. There is
  today **no workflow channel that passes a page's live stored section
  content to its writer for editing** — candidate 1 (a real edit path) is not
  merely preferred, it is the only candidate of the three that the plumbing
  can currently support without adding a new one.
- Full verdict JSON (5 iterations, final UNVERIFIABLE with the gap named)
  preserved before the ~24h reaper:
  `bugfix_154_work_item_routing_columns/EVIDENCE_2026-08-03_178_verdict_aece2920.json`;
  bundles in `diagnosis_artifacts` under `aece2920%`.

**relojistas' "deleted DefinedTermSet slot" is closed by its owning lane's own
record**, found by tracing the writer instead of the row:

- `page_component_history` has NO slot_name column — the pointer is
  `component_id`, FK `ON DELETE SET NULL`, which is why the snapshot looks
  anonymous after the slot row went.
- The content was a `structured-data-block` component (JSON-LD DefinedTermSet,
  the 8 glossary terms) built by the traffic_probe lane on 2026-07-28 — and
  **backed out by that same lane, same day, deliberately**:
  `sectionHasVisibleContent()` strips `<script>` then requires >10 visible
  chars, so a JSON-LD-only section is dropped by the page assembler by design
  and can never reach the served page. Evidence:
  `traffic_probe/relojistas_rebuild_running_notes.md` 2026-07-28 (3) plus its
  same-session CORRECTED entry; component `b51dbc8f` sits in the library
  `is_active=false` with the full reason in its description. Snapshot
  `b0e119a4` (16:21Z, source `save_page_sections_overwrite`) is the back-out's
  own pre-write copy.
- **Do not restore it by any route** — the section is structurally
  undeliverable. The glosario row (3→2 slots, −48%) leaves this bug's
  fleet-signature list entirely: it was the back-out, not a regeneration.

## Contributed by the bugfix_177 lane, 2026-08-03 — two more crosslink items are DELIBERATELY parked behind a dead dependency, for you to release

`9e9ec430-ff92-4264-83cc-6072840faad8` and `18bc832c-c937-4608-9a05-718772d44c88`
(both `content_rewrite`, item_key `tool_crosslink:tool-cma-obligation-checker:…`,
status `triaged`) depend on `a5cabea0` — a 177 zombie now `wont_fix`, which can
never release a dependency. **177's approved plan cleared those deps; the apply
was NARROWED after reading this bug's own diagnosis note** (doc_notes
2026-08-03: dispatching `93f2a3b7` — released 08-02 on the "stands on its own
merits" reasoning — regenerated whole slots and dropped paragraphs). So the
dead dependency is being used as a VISIBLE interlock: the items cannot dispatch
until someone clears `depends_on`, and that someone should be this lane, in the
same breath as its fix, when dispatching a crosslink item stops being
destructive. `UPDATE site_work_items SET depends_on = NULL WHERE id IN (…the
two ids…);` is the whole release. If this bug is closed another way (e.g. the
crosslink path is retired), cancel them instead — do not let the interlock
outlive the reason for it.

## UPDATE 2026-08-03/04 (154 lane) — FIX IMPLEMENTED, LIVE (v1.0.1247), content-matching PROVEN; end-to-end blocked by an unrelated, newly-discovered bug (192)

**The fix is candidate 1 from this file's own list**: a third `spec.mode`
value, `"edit_live"`, alongside `"recreate"`. New Go action
`load_current_section_content_action.go` (registered as workflow step
`load_current_section_content`, inserted between `check_has_ready_sections`
and `spawn_content_writer`) joins each ready section against
`page_components.slot_name` for the page and attaches the current
`rendered_html` as `existing_content_html`. Both live `content_rewrite`
emitters (`create_tool_cross_link_items.go`, `apply_gap_plan_action.go`'s
`applyAddToPage`) now set `mode="edit_live"` — both always target an
already-existing page. Default OFF for every other caller (2 passthrough
unit tests prove no query fires without the field). Full design writeup,
citations and the shared-seam ruling this follows: see this workstream's
`NOTES_work_item_routing_columns.md` 2026-08-03/04 entry and register entry
PBP-028 (`docs026_concept_register/register/page-build-pipeline.md`).

Committed (`08d0515f3`), submitted to council (`Council-Submitted:
97ebadcf-bbe6-485f-8231-ff16fc4e679f` — **that run stalled mid-review and
never produced a verdict**, unrelated to the change itself; advisory only,
not blocking), built as `v1.0.1244`, deployed by the owner's whole-fleet
release as part of `v1.0.1247`, pod-verified on both replicas (binary
symbol counts, not the tag). Migration `299` applied and recorded.

**Live verification, using the two items parked above** (their `spec` had
no `mode` key — predating this fix — so they were patched with
`mode="edit_live"` before release, not released as-is, which would have
reproduced the original damage). Fired `build-dispatch-loop` directly for
the target site rather than waiting.

> **CORRECTED 2026-08-04 (same day, owner asked "check cron as well as
> scheduled tasks, think hard"):** the line originally here claimed
> `build-dispatch-loop` "isn't scheduled anywhere ... no matching row in
> `scheduled_tasks`". **That was wrong, and the check that would have caught
> it is cheap: I filtered `scheduled_tasks` on `name ILIKE '%dispatch%' OR
> target_agent_type ILIKE '%dispatch%'` and stopped there.** The real row is
> `build-pipeline-trigger` (target_agent_type `build-pipeline-trigger`,
> `interval_seconds=120`, `enabled=true`, firing on schedule via the
> `kafka-scheduler` service, not a k8s CronJob) — its own workflow's
> `find_dispatchable_site` step picks the oldest-eligible site fleet-wide and
> its `call_dispatch` step calls `build-dispatch-loop` for it, every two
> minutes. So the build pipeline **is** scheduled; my manual fire was
> unnecessary (the two items would have dispatched within a couple of
> cycles regardless, subject to the fair-queue ordering other sites compete
> for). The lesson: a substring filter on a name is not a search for a
> concept — I should have listed every row and read it, not grepped for the
> word I expected the row to contain.

**Confirmed working, at the data level**: the dispatched orchestration's
`section_plan.sections_ready[0].existing_content_html` held the page's
exact, complete, unmodified current `generic-text-block` prose (the CMA
compliance content, verified word-for-word against what's live) — proof the
join by slot name found the right row and handed it to the writer intact.

**Could not complete the full before/after `content_data` length check**:
both dispatches then failed one step later, inside `page-content-writer`'s
OWN `select_sections`/`process_sections_loop` — filed as `bugs_open/192`.

> **CORRECTED 2026-08-04, by the `bugfix_192_select_sections_wrapper` lane
> (see their NOTICE below), then re-checked here after their fix**: the
> claim that 192 was "a separate, pre-existing bug... confirmed not caused
> by this fix" was **wrong for the failure I personally triggered and
> diagnosed** (the 08-04 08:26 dispatches). It WAS this fix — my own
> `load_current_section_content_action.go` returned a wrapper object on
> every path including its "pass-through" ones, and because its
> `output_field` reused the `section_plan` key, that wrapper silently
> replaced the real plan on **every page build in every mode**, fleet-wide,
> from the moment migration 299 was applied (~08:20 on 08-04) until the 192
> lane's fix landed (~09:01Z). What I got right: the timing check itself was
> a sound instinct; what I got wrong was trusting an AGGREGATE count
> (`current_step='process_sections_loop' AND status='FAILED'`) as proof of a
> single cause, without reading the actual error text of the historical
> (08-03 21:00) batch to confirm it matched what I was looking at. It
> didn't — that earlier wave was a different, still-undiagnosed failure at a
> different step (`iter_N_generate_content`). Two different problems looked
> like one because I compared counts, not messages.**

No content was lost either way (the failure was upstream of
`save_page_sections`), and the 192 lane's fix is now live+committed. Full
account: `bugs_open/192`, its NOTICE to this file below, and this
workstream's `NOTES_work_item_routing_columns.md`.

## UPDATE 2026-08-04 (154 lane, after the 192 fix) — verification completed: ONE page confirms the fix, ONE page reproduces the ORIGINAL bug for a DIFFERENT reason

Both parked items ran to `complete` after the 192 fix landed (`18bc832c` at
09:05Z, `9e9ec430` shortly after). Checked `content_data` on both, per this
file's own "how to verify a fix" test:

**`guide-cma-compliance` (`d8c51ace-...`, item `9e9ec430`) — SUCCESS.**
`generic-text-block` content_data: **6034 → 6240** (+206 chars, roughly the
inserted link anchor plus surrounding adjustment). `slot_name` unchanged.
This is exactly what the fix was built to do, on a real production page.

**`guide-independent-strategy` (`2a347990-...`, item `18bc832c`) —
REPRODUCES THE ORIGINAL BUG, by a route the fix does not cover.** The
section's stored slot was `generic-text-block` (3637 chars, confirmed in
`page_component_history` id `c5769938`, snapshot taken at write time). After
this run, the page has **no `generic-text-block` row at all** — its only
section is now `article-body`, 3262 chars, and the text is wholly different
prose (different opening, headers, a fabricated "Last reviewed: October
2023" line, none of the original sentences). Diffed directly, not inferred.

**Root cause of this instance, read from the orchestration's own
`section_plan`**: for THIS page's build, `plan_sections`' component
resolution (Path 2, the section-type selector) resolved the section to a
**different, generic fallback component** — literally the same
`article-body` component minted earlier that day for an unrelated page,
`tool-gripper-payload-calculator-guide` (its own `description` field says
so verbatim: `"Component needed for section type \"article-body\" on page
\"tool-gripper-payload-calculator-guide\""`). So `section_plan.sections_ready[0].name`
was `"article-body"`, not `"generic-text-block"` — **`load_current_section_content`'s
join correctly found no `page_components` row named `article-body` for this
page (there isn't one), so `existing_content_html` was legitimately never
attached, and the writer did exactly what it did before this fix: fabricate
a fresh section from the guidance text alone.**

**This is a real, undiscovered-until-now limitation of the fix, not a bug in
the join logic itself**: matching by exact section name only works when the
CURRENT build's resolved component identity for a section agrees with what
is actually STORED under that page. `generic-text-block` and `article-body`
are both generic fallback components (`generic-text-block`'s own
description: *"Fallback component for any unmatched section"*) — the
selector can evidently pick a different one build to build for what a human
would call "the same section", and nothing before this incident said slot
identity for a generically-resolved section is stable across rebuilds.
`[UNVERIFIED]` how often this happens fleet-wide — this is one observed
instance, not a measured rate. The old content is not lost (recoverable from
`page_component_history` id `c5769938`, same as the original bug), and the
shrink guard did not fire because there is no matching slot_name to compare
against — a whole-slot rename bypasses a same-slot shrink check entirely,
which is itself worth someone's attention.

**Status: fix candidate 1 is real, live, and proven to work for the case it
was designed for (stable slot/component identity across builds) — but it is
NOT a complete fix for this bug. A page whose section resolves through the
generic-fallback selector path can still lose its content exactly as
before. Leaving this file OPEN rather than closing it.** Next step for
whoever picks this up: either make `load_current_section_content`'s match
tolerant of a component-identity change (e.g. match by `page_id` + "this
was the only/most-recent prose slot" when the resolved name has no
corresponding stored row), or fix the underlying instability in how
generic-fallback components get assigned identity across rebuilds — the
options are not evaluated here, this update is evidence, not a design.

---

## NOTICE 2026-08-04 from the `bugfix_192_select_sections_wrapper` lane — your action's RETURN SHAPE changed; your fix's substance did not

Told rather than merely measured, per the owner ruling of 2026-07-29 §3. **You own
`load_current_section_content_action.go`; I have changed how it returns.** Read this
before your next edit to it.

**What happened.** The step is wired with `output_field: section_plan`, deliberately
reusing the key `plan_sections` writes so `call_content_writer`'s `input_mapping`
needed no change — your seed `299` says exactly that, and the reasoning is sound. But
`coordinator.go:1859-61` (`storeActionResult`) stores an action's return value
**wholesale** under `output_field`, and the action returned
`{section_plan, applied, reason|matched}` on **every** path, including all eight it
documents as pass-throughs. So `collected_data.section_plan` became a *wrapper* on
**every page build in every mode**, not only `edit_live` ones — and
`page-content-writer.select_sections` could no longer find
`input_data.section_plan.sections_ready`. **Every page build in the fleet failed** from
~08:20 on 08-04 until a config seed landed at 09:01:35Z. Filed as `bugs_open/192`.

**What I changed, and what I did NOT.** The enrichment logic — matching `slot_name` to
`sectionPlanItem.Name` and attaching `existing_content_html` — is **untouched and still
correct**; it was verified working by the lane that found this, and my end-to-end
re-dispatch exercised it again. What changed is only the envelope:

- every path now returns **the plan itself**. Your header always promised this
  ("gets the section_plan it was handed, **byte-for-byte unchanged**") — the code did
  not do it, and now does;
- `applied`/`reason` go to the **log**; on the applied path only, `applied`/`matched`
  are kept in one namespaced key **inside** the plan, `edit_live_meta`, so you keep a
  DB-visible signal that the channel fired;
- **your test file changed** and you should know why: it asserted `result["applied"]`,
  `result["reason"]` and `result["section_plan"]`, i.e. it encoded the wrapper as the
  contract, so it **passed on the code that caused the outage** — while its own comment
  two lines up said "must leave section_plan byte-identical". It now asserts the
  contract, and the pass-through cases assert `reflect.DeepEqual` identity, which is
  strictly stronger than what was there. It is mutation-proven: reverting the action to
  the wrapper fails all five cases.

**Nothing is owed by you.** Committed `2b9d84072`, council `7afbf531-…`, registered as
WFA-009. If you disagree with `edit_live_meta` living inside the plan, that is the one
judgement call worth arguing — say so in `bugs_open/192` and I will follow it.

## UPDATE 2026-08-04 (154 lane) — component-identity-drift gap: MEASURED, then closed for the unambiguous case; still OPEN for the ambiguous one

Picked up this file's own top-priority open item: "measure how often this
actually happens fleet-wide before investing heavily; this session has one
observed instance, not a rate."

**Measured.** Of 127 pages with a current `site_plan` and existing
`page_components` rows, **3 (2.4%) already carry a stored `slot_name` absent
from their own current plan's component-name list** — a lower bound on
exposure, since it only counts pages already mismatched today, not pages that
could newly mismatch on their next rebuild:

- `gripper-cycle-time-estimator` (site `00ff3af5-…`): slots `faq` and
  `tool-gripper-cycle-time-estimator`, plan lists only
  `call-to-action,generic-text-block,hero`.
- `guides-index` (site `1244516d-…`): slot `guide-list`, plan lists
  `content-listing,hero`.
- `index` (site `5fe8785b-…`): slot `image-hover-card-grid`, plan lists six
  other names.

Query (run against the live DB, 2026-08-04):
```sql
WITH plan_names AS (
  SELECT sp.site_id, sps.page_name, array_agg(DISTINCT sps.component_name) AS planned_names
  FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id = sp.id
  WHERE sp.is_current = true GROUP BY sp.site_id, sps.page_name
), stored AS (
  SELECT p.id AS page_id, p.site_id, p.name AS page_name, pc.slot_name
  FROM pages p JOIN page_components pc ON pc.page_id = p.id
  WHERE pc.slot_name IS NOT NULL AND pc.build_status <> 'removed'
)
SELECT s.site_id, s.page_name, s.slot_name, pn.planned_names
FROM stored s JOIN plan_names pn ON pn.site_id = s.site_id AND pn.page_name = s.page_name
WHERE NOT (s.slot_name = ANY(pn.planned_names));
```

**Read the three cases before generalising from them**: none is actually "the
selector flipped between two equally-scored fallback candidates for the SAME
`section_type`" — I checked, and `article-body`/`generic-text-block` (this
file's own instance) each have exactly one active `content_components` row
under their own distinct `section_type`, so that specific mechanism was not
the cause here. The three fleet cases above look instead like a page's stored
slots simply including things the plan's own skeleton never named (a tool
component, a listing component) — a **different**, and probably more common,
route to the same failure: `content_components` is a fleet-shared library
with no `site_id`/`page_id`, so nothing has ever required a page's plan and
its actually-attached components to agree on names. **The "why do names
drift" root cause (fix candidate 3 in this file's original list) is still
open** — this update closes the symptom for the case that matters
operationally, not the mechanism.

**Fix shipped, for the unambiguous case only**: `load_current_section_content_action.go`
(the same file `bugs_open/192`'s NOTICE names you as owning) now falls back
from the exact `slot_name` join to "the page's current single prose slot"
— but ONLY when there is exactly one ready section that missed the exact
match AND exactly one still-unclaimed `page_components` row on the page is
prose-sized (reuses `save_sections_shrink_guard.go`'s own
`minShrinkGuardChars`/`shrinkGuardTagStripper` rather than inventing a second
definition of "prose-sized"). Two or more of either side and it leaves both
unmatched — no guessing, never worse than before this fallback existed.
Recorded structurally as `edit_live_meta.fallback_matched`, kept separate
from the exact-match count `matched` — which doubles as the first fleet-wide
instrumentation of this drift class's real rate, on every future `edit_live`
build, instead of requiring another ad hoc SQL census.

Three new test cases in `load_current_section_content_action_test.go`
(fallback fires on the single-unmatched/single-candidate case; does NOT fire
when two sections are unmatched — ambiguous; does NOT fire when the leftover
slot is small/hero-sized). Mutation-proven: `git stash` of just the action
file, then running the three new tests against the pre-fallback code fails
all three with the exact expected/got mismatch; restore verified, full
package `go test ./platform/orchestration/actions/...` green. Committed in the
same commit as this update (see git log for this file's own path); council
submission `Council-Submitted: 8a3e0315-4576-4829-bf42-c0c8cdfc4e3a` (verdict pending —
advisory only, does not block). **Not yet built or deployed** — this is
source + tests only, following the same pattern as this file's own prior
updates (built/deployed as part of a later whole-fleet release rather than a
one-off image).

**Still open, in priority order:**
1. The ambiguous case (two-or-more unmatched sections, or two-or-more
   candidate prose slots on one page) is explicitly NOT covered — the fallback
   refuses to guess, which is safe but leaves the original bug reachable
   there. Not yet observed in the fleet measurement above (no page showed more
   than one unclaimed prose-sized slot at once), so unknown severity.
2. The actual mechanism causing a page's stored slot names to diverge from
   its current plan's names is still undiagnosed — candidate 3 from this
   file's original list (fix the underlying instability) remains open, now
   with three additional fleet instances as raw material if someone picks it
   up.
3. Build, deploy, pod-verify, and live-verify this fix the same way the
   `edit_live` channel itself was verified (dispatch a real parked item
   against a page in this file's mismatch list above, diff `content_data`
   before/after).
4. Watch-list item 4 from the prior update (shrink guard blind to a
   whole-slot rename) is unchanged by this fix and still open.

## UPDATE 2026-08-04, later — deploy verified LIVE; council resubmitted after the first attempt silently dropped; root-cause 090 dispatched; a self-caught correction on how to live-verify

**Deploy verified at the artefact, not the tag.** `v1.0.1251`, both replicas
(`agent-chassis-5455ddcdcc-crnb6`, `agent-chassis-5455ddcdcc-gpr92`),
pod-grepped: `strings /app/agent-chassis | grep -c "single-unmatched-prose-slot"`
= 1 on both, plus the existing `"SECTION SHRINK"` positive control = 2 on
both. **The fallback fix is live.**

**The first council submission (`8a3e0315-…`) never reached the
orchestrator** — checked ~8.5h after submitting: zero `orchestration_states`
rows, by exact id AND by payload search on `fix_correlation_id`, far past
the ~29-minute measured queue latency this repo's own docs cite. Likely
cause: the trigger script's `kcat -P` publish landed right on top of a
transient cluster connectivity blip this session hit at submission time
(`Unable to connect to the server: dial tcp 10.10.10.10:443: i/o timeout`),
and `kcat -P` is a known silent-failure shape (exits 0, sends nothing).
**Resubmitted the identical JSON** (not a `RESUBMIT_CORR` — there was no
actual prior round to link, since nothing ever ran): new correlation
`Council-Submitted: 56f9a5a2-4d37-4114-9442-239861acd36e`, confirmed
in-flight this time (`council-gate-orchestrate-0804-1926`, progressed
`review_editquality` → `review_constitution` within the same minute).
Verdict not yet landed as of this update — check before assuming either
submission's status.

**Dispatched the root-cause diagnosis** this file's own open list has
carried since the previous update: why does a build's resolved component
name ever disagree with an earlier build's stored `slot_name`, given
`plan_sections`' Path 1 should match an existing, correctly-named component
directly? Filed via `090_TRIGGER_needs_diagnosis_v1.sh` with a fresh slug
(`178-component-identity-drift-mechanism` — deliberately distinct from the
already-completed `178-crosslink-regenerates-whole-section` run, so it
doesn't dedupe against stale results) and the full evidence trail for
`guide-independent-strategy` in the symptom text. **Correlation:
`2bcf9359-4603-472f-ba00-4d1d5f33f6c8` (intake), run correlation
`167d2cc2-0b98-405c-a1d7-d54d80ed37c9` (use this one for artifacts).**
Diagnose-dispatch-loop claimed it within the 180s the trigger waited.
**Not yet read** — whoever picks this up next should read the verdict
before doing anything else with it. Note the trigger's own advisory: local
HEAD was 50 commits ahead of `origin/087_towards_multiple_domains` at
dispatch time, so the diagnosis reads the PUSHED tree, not this session's
commit — irrelevant to this particular question (it's about pre-existing
history, not this session's fix) but worth knowing for any diagnosis that
depends on very recent commits.

**A correction, caught before it became a wasted live-verification
attempt**: the previous update's fleet measurement (3/127 pages, 2.4%) is
NOT a ready-made set of test cases for THIS fallback specifically, and
citing it as one would have been a wrong call. Re-examined the mechanism:
`load_page_sections_from_spec_action.go` sources a page's section list from
`site_plan_sections` (the CURRENT plan) — so `plan_sections` only ever
builds `sections_ready` entries for names the plan actually lists. The three
measured pages' mismatched slots (`faq`, `tool-gripper-cycle-time-estimator`,
`guide-list`, `image-hover-card-grid`) are **not in their pages' plans at
all** — they are extra components attached via some other route (tool
creation, FAQ addition), sitting on the page but never entering
`sections_ready` for ANY build, edit_live or otherwise. Dispatching a
`content_rewrite` item at one of those pages would not exercise the
fallback — the mismatched slot simply never becomes a "ready section" to
attempt a match against. This is a genuinely different shape from
`guide-independent-strategy`'s case, where the PLAN's own named position
resolved to a component that disagreed with what was stored — and that page
no longer reproduces it either, because the run that discovered the bug
already rewrote its stored slot to agree with the plan (`article-body`).

**So there is currently no known live page that would exercise the
fallback**, and manufacturing one on production data was judged out of
scope for this session. What exists instead: solid unit + mutation-test
coverage (previous update), a live deploy, and **instrumentation** —
`edit_live_meta.fallback_matched` is now written into every `edit_live`
build's `section_plan`, so the fallback will self-report the first time it
fires naturally. **Next session: check for it** rather than trying to force
a repro:
```sql
SELECT id, created_at, collected_data->'section_plan'->'edit_live_meta'
FROM orchestration_states
WHERE collected_data->'section_plan'->'edit_live_meta'->>'fallback_matched' = '1'
ORDER BY created_at DESC LIMIT 10;
```
(subject to the usual `orchestration_states` retention — terminal rows live
~24h, so this only sees recent activity; if nothing has fired yet fleet-wide,
that is itself informative about how rare the unambiguous case actually is.)

**One thing that IS still yours, and it is good news:** `192`'s filing notes your own
end-to-end verification was blocked by this outage ("the remaining check — does the
writer actually preserve it end to end — is blocked on this bug, not on 178's own
code"). **That block is gone**; builds complete again. The `content_rewrite` item
`18bc832c` (vetcomparison, `guide-independent-strategy`) ran to `complete` at 09:05Z
and is a ready-made subject for your `content_data` length check.

## UPDATE 2026-08-04, later still — both in-flight results read; council REVISE answered with code evidence and resubmitted; root-cause diagnosis came back UNVERIFIABLE

**Council verdict on the fallback fix (round 1, `56f9a5a2-…`): REVISE**, decided by
a HIGH-severity gating objection from `bug_historian`. Two further MEDIUM
objections (`architecture`, `prior_art_librarian`). The other 8 reviewers
approved, several explicitly endorsing the narrow opt-in scoping. Full verdict:
`diagnosis_artifacts` kind=`council_report`, correlation `56f9a5a2-4d37-4114-9442-239861acd36e`.

**`bug_historian`'s objection checked against the actual code, and it does not
hold for either file it named:**
- `rerender_page_sections_action.go:234-276` does **not** do the naive exact
  slot_name join this fix patches. Since `bugs_open/182` it resolves
  `page_components.component_id` FIRST — immune to naming drift — falling back
  to the name/function map only when `component_id` is absent, and logging
  observe-only when the two disagree (13 live sections fleet-wide, measured
  2026-08-03). Already hardened, differently, before this bug existed.
- `save_page_sections_action.go` is not a stale-row lookup at all — it deletes
  the page's components and inserts a fresh row per section of the new
  composition every time. Its only slot_name-keyed lookup is `matchLockedRow`,
  a narrow, already-disclosed exception (`bugs_open/058`, human-locked rows
  only).
  Neither file shares the "exact-name join, silent miss, fabricate" shape this
  fix targets. On the evidence read, that shape is unique to
  `load_current_section_content_action.go`'s pre-fallback code.

**`prior_art_librarian`'s objection checked too**: `datahelpers.SectionIdentityKey`
(cited as possible prior art) is explicitly the wrong tool for this — its own doc
comment (`datahelpers/section_text.go:93-163`) states slot equality is a
**necessary** condition ("Two different components in two different slots are
not the same section rendered twice, whatever their content_data happens to
contain") plus byte-identical content, built for exact-duplicate deletion
(`bugs_open/156` round 2). This fallback answers the opposite question — same
section, different slot, no byte-identity available — for read-only enrichment,
not deletion. No existing mechanism solves it.

**`architecture`'s objection (root cause still open) stands** — see the
diagnosis result below — but is now backed by a completed, if inconclusive,
investigation rather than an untried gap.

**Resubmitted** with these three answers folded into the rationale, no code
change: `SUBMISSION_2026-08-04b_component_identity_drift_fallback_revise_response.json`,
`RESUBMIT_CORR=56f9a5a2-4d37-4114-9442-239861acd36e` (same trail id, so round 2
accumulates against round 1 per this repo's own practice). Confirmed genuinely
dispatched, not a repeat of the earlier `kcat` silent-drop (`council-gate-orchestrate-0804-2021`,
seen executing `review_editquality` seconds after submitting). **Verdict not yet
read as of this update** — check before assuming either round's status:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='56f9a5a2-4d37-4114-9442-239861acd36e' AND kind='council_report'
ORDER BY created_at;
```
If the second row is APPROVED, commit trailer `Council-Reviewed: 56f9a5a2-4d37-4114-9442-239861acd36e`.

**The root-cause 090 diagnosis (`167d2cc2-…`) came back UNVERIFIABLE, not
silent.** `status: "UNVERIFIABLE"`, `stopped_by: "scope-not-narrowing"` after 5
iterations. It never obtained the actual body of `plan_sections_action.go`'s
Path 1/Path 2 resolution logic despite that being the hypothesis's central
claim, and — the real dead end — **no `page_component_history` row predates
the 2026-08-04 09:04:48 overwrite event under investigation**, so there is no
forensic trace of what wrote `slot_name='generic-text-block'` for this
position, or when. Pulled the one pre-overwrite snapshot that does exist
(`page_component_history` id `c5769938…`) directly: its `content_data` holds
real prose (`heading`/`content` keys, the actual pre-overwrite article text) —
confirming the position DID hold real content, consistent with this bug's
original finding — but `page_component_history` has no `slot_name` column (see
this file's earlier note), so it cannot answer the diagnosis's specific
question either. **The mechanism remains genuinely unknown**, not merely
unread — candidate 3 stays open with no lead.

**One claim from the diagnosis loop itself needed checking before repeating
it, and it was wrong**: its `NeededEvidence` field asserted `page_component_history`'s
snapshot INSERT is "a real bug" because it writes `pc.id` instead of
`pc.component_id`. Checked against schema: `page_component_history.component_id`
**FKs to `page_components(id)`**, not to `content_components` — so `pc.id` is
correct, not a bug. The diagnosis loop misread what the column identifies
(the page's own slot row, not the fleet-shared library component). Not acted
on; recorded here so it isn't repeated by whoever reads that artifact next.

**Still open, in priority order (updated):**
1. Council round 2 verdict — check before doing anything else with this bug.
2. The root-cause mechanism (candidate 3) — now genuinely investigated and
   still unknown, not just deferred. A future occurrence with a
   `page_component_history` row that predates its own overwrite is the only
   evidence class that could settle it; this one didn't have one.
3. The ambiguous case (two-or-more unmatched sections/candidate slots) is
   still unhandled by design — unknown severity, not yet observed.
4. Watch for the fallback's first natural firing (query in the previous
   update), subject to ~24h `orchestration_states` retention.
5. Watch-list item unchanged for four updates: the shrink guard doesn't fire
   on a whole-slot rename.

## UPDATE 2026-08-05 — council round 2 APPROVED; a fresh chassis build re-verified at the pod; the shipped commit's trailer names the wrong (dropped) correlation

**Council round 2: APPROVED.** `56f9a5a2-4d37-4114-9442-239861acd36e`,
decided 2026-08-04 20:27:31 — the resubmission answering `bug_historian`'s
gating objection and `prior_art_librarian`'s pointer with code evidence
(previous update) landed. The fallback fix is now both **live and reviewed**.

**A fresh chassis build was rolled** (owner-initiated). Verified at the pod,
not the tag: `v1.0.1252`, both replicas (`agent-chassis-5b64b888f5-4j2bc`,
`agent-chassis-5b64b888f5-fs4dq`), `strings /app/agent-chassis | grep -c
"single-unmatched-prose-slot"` = 1 on each, plus the `"SECTION SHRINK"`
positive control = 2 on each. **The fallback fix survived the rebuild.**

**Found a loose end while checking this:** the commit that actually shipped
the fix (`4b3f9f89b`) carries `Council-Submitted: 8a3e0315-4576-4829-bf42-c0c8cdfc4e3a`
— the FIRST submission, the one that silently dropped and never ran (see
the 2026-08-04 update above). The verdict that actually landed is under a
DIFFERENT correlation (`56f9a5a2-…`, the resubmission after the drop was
caught). `098`'s coverage report resolves a `Council-Submitted:` trailer by
looking up ITS OWN correlation's verdict — `8a3e0315` will never have one,
so `4b3f9f89b` would read as permanently unresolved if left as-is, despite
the change being reviewed and approved under a different, valid trail.
Forward-only forbids amending `4b3f9f89b` — corrected with a follow-up
commit carrying `Council-Reviewed: 56f9a5a2-4d37-4114-9442-239861acd36e`
(this update's own commit) so the coverage report has a trailer that
actually resolves.

**Checked for the fallback's first natural firing — none yet**
(`fallback_matched = '1'` query from the previous update: 0 rows). Not
informative on its own given `orchestration_states`' ~24h retention; still
the only outstanding piece of end-to-end evidence for this fix.

**Status of the fallback fix (candidate 1 of 3) specifically: DONE** — live,
pod-verified across two separate builds now (`v1.0.1251`, `v1.0.1252`),
council-approved. **This bug file stays OPEN** because the other two items
below are unresolved, not because of anything about the fallback itself.

**Still open, in priority order:**
1. The root-cause mechanism (candidate 3) — investigated (090 diagnosis,
   `167d2cc2-…`, UNVERIFIABLE) and still unknown. No live lead; needs a fresh
   occurrence with an intact `page_component_history` trail.
2. The ambiguous case (two-or-more unmatched sections/candidate slots) is
   unhandled by design — unknown severity, not yet observed.
3. The fallback's first natural firing — watch, don't force (query above).
4. Watch-list item unchanged for five updates: the shrink guard doesn't fire
   on a whole-slot rename.
