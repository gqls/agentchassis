# HANDOFF 2026-08-03 (rev 2, ~20:20 BST) — shrink-guard thread CLOSED; 178 root cause CONFIRMED; next task is 178's FIX, fully scoped below

**Read this first; then `NOTES_…` (2026-08-03 tail, both entries after
"root-cause 090 run dispatched") for the full evidence trail.**
`SUMMARY_2026-08-03_work_item_routing_columns.md` is the read-aloud milestone
account of the lane up to the shrink-guard close. This revision's news: a
**fresh chassis build (v1.0.1243) was deployed and pod-verified**, and the
178 root cause is now CONFIRMED — **the fix itself is unimplemented and is
the clear next task.** Full design direction below so a new session can start
writing code immediately rather than re-deriving it.

## Fresh build verified 2026-08-03 ~20:15 BST — v1.0.1243, both replicas

Per this repo's rule ("a roll is not evidence your fix shipped" —
`bugs_open/153`), verified by pod-grep, not by tag or git log:

```
kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}'
  -> docker.io/aqls/agent-chassis:v1.0.1243
kubectl exec <pod> -- strings /app/agent-chassis | grep -c raiseToolContentItem   -> 7  (both pods)
kubectl exec <pod> -- strings /app/agent-chassis | grep -c recurrenceExpected     -> 1  (both pods)
```

**What shipped in this build that touches this lane's tracked items:**
`bugs_open/177` is **FIXED, council-APPROVED, and now confirmed LIVE**
(commit `74655b709`, Council-Reviewed `982507b0-2e18-4457-a354-85a809012bbd`,
a different session's lane — `bugfix_177_tool_content_items`). Their fix
(`tool_content_item.go`, a shared `raiseToolContentItem` helper used by both
tool emit paths, read-only section-satisfiability check before raising the
item) is the pod-grepped symbol above. **Their sweep deliberately did NOT
release the two blocked `content_rewrite` dependents** — their commit
message names why: *"dependents stay blocked, their dispatch is destructive
today"* — citing THIS lane's completed 178 diagnosis (dispatching a
crosslink-class item still destroys content until 178 is actually fixed).
That is this session's diagnosis work being read and correctly acted on by
another lane — nothing further to do there. Item 3 in the old priority list
(below) is therefore **done, not "unstarted"** — struck out.

**Checked, does NOT touch 178's mechanism**: no commit since the 178
diagnosis completed (11:00Z) touches `load_existing_content_action.go`,
`save_page_sections_action.go`, `create_tool_cross_link_items.go`, or
`apply_gap_plan_action.go`. The confirmed root cause below is still exactly
where this lane left it.

## State: DONE, LIVE (v1.0.1238, pod-proven both replicas), and APPROVED — unchanged from rev 1

- `bugs_closed/154` (routing columns) + `bugs_closed/176` (selector↔loader,
  FIFO fairness) — closed earlier; unchanged.
- **Per-slot shrink guard** (`2da3e08e5`) — live since 1233, **proven by live
  induction** (refused twice, honestly FAILED, refusal item emitted+deduped,
  zero bytes written; prediction recorded before outcome — NOTES 08-03).
  Council `e64f8576`: REVISE → **APPROVED r2** (08-02 23:11Z).
- **Locked-slot exclusion** (`5f00dcba9`) — live in 1238, proven by ANCESTRY.
- **Refusal wording** (`77b58fd4d` + advisory `0913d5754`) — council
  `98aa9103` **APPROVED** (08-02 23:26Z). Live in 1238, both replicas.
- Restores (`287`) artefact-proven earlier.

**Nothing on the shrink-guard thread is owed. No post-roll checks
outstanding on it.**

## NEXT TASK — implement the 178 fix (root cause is confirmed; this is new code)

### The confirmed mechanism (full citations in `bugs_open/178`'s final update)

`content_rewrite` work items never set `spec.mode`. `page-build-handler`'s
`load_existing_content` step is a hard gate —
`load_existing_content_action.go:64-69`:
```go
mode := inputs.Get("mode")
if mode != "recreate" {
    return map[string]interface{}{"has_existing": false, "reason": "not_recreate"}, nil
}
```
— gated on the **adoption-crawl** path only (its own doc comment: *"loads
existing page content from research_results"* — the ORIGINAL crawl, never
current `page_components`). `call_content_writer`'s `input_mapping` passes
that no-op result as `existing_content?`, and the only other page-shaped
input, `current_page: page_record`, carries only "sections, title,
page_type" per `load_page_record`'s own step description — no prose.

**So `page-content-writer` gets the item's guidance text and NOTHING to
edit**, for any `content_rewrite` item against an already-built page. It
fabricates a full replacement section that satisfies the instruction's
shape — well-formed HTML, the requested addition present, but the prior
prose gone. Confirmed on `93f2a3b7` (robot-hands gripper page): `spec.mode`
absent, `create_tool_cross_link_items.go` never sets one (grepped, zero
hits). `apply_gap_plan_action.go`'s `content_rewrite` emission (:243) is
the other live source and was **not checked** for a `mode` key — check it
before assuming the fix only needs to touch the crosslink path.

### Why fix candidate 1 (a real edit channel) is the only one the current plumbing supports

Fix candidate 2 (bug file's original second choice: "raise it only when the
spec declares sections") doesn't apply here — 178 isn't about whether to
raise the item, it's about what the writer receives once raised. The
diagnosis also **eliminates the tempting cheap fix**: setting
`spec.mode="recreate"` on a crosslink item would NOT work, because that path
loads `research_results` (the original adoption-crawl snapshot), not current
`page_components` — it would feed the writer STALE pre-edit content, which
is arguably worse than none. **There is today no workflow channel that
passes a page's LIVE stored section content to its own writer for
editing.** Candidate 1 — build that channel — is the one to implement.
Candidate 3 (emit the delta so the loss is visible even when allowed) is a
weaker, complementary safety net, not a replacement.

### Design direction — read before writing code

- **This is a shared-seam change** (`page-build-handler`'s `call_content_writer`
  step, used by every page on the platform) — the **2026-08-02 owner ruling**
  applies: *"new authority on a shared seam ships as an OPT-IN FIELD, not a
  documented contract... make X a field with the unsafe default OFF."* Do
  NOT change `call_content_writer`'s behaviour for every caller. The shape
  that fits: a new, explicitly-named input (e.g. `edit_mode` or reuse `mode`
  with a THIRD value alongside `"recreate"`) that a `content_rewrite` emitter
  opts into, loading current `page_components.content_data` for the page's
  existing sections and passing it to the writer as material to edit —
  default OFF so every other caller of `page-build-handler` (fresh builds,
  `needs_content_page`, adoption `recreate`) is provably unaffected.
- **Likely NOT architecture-scope** under the 2026-07-29/08-02 rulings if
  scoped this way: it's an opt-in capability on an existing step, reachable
  by nothing until an emitter names it — same shape as RFC_010 D1's ruling
  on opt-in fields. Re-read
  `architecture_review/RFC_010` and the two 2026-08-02 narrowings in
  CLAUDE.md before submitting, to write the submission in the right register
  (still submit to the council gate regardless — this touches `platform/`).
- **Where the new content-loading logic should live**: a new action file,
  following the `raiseToolContentItem`/`tool_content_item.go` precedent the
  177 lane just set (a shared helper, unit-tested with the platform's sqlmock
  patterns, reused by every `content_rewrite` emitter rather than duplicated
  per call site). Load current section content_data by slot_name, matched
  against `section_plan`, read-only.
- **Both known emitters need the new field set**: `create_tool_cross_link_items.go`
  and `apply_gap_plan_action.go` (the latter unchecked — read its `content_rewrite`
  emission at :243 first to see if it already has a legitimate reason to
  omit it, e.g. if it always targets brand-new pages).
- **Verify per the bug file's own test**: raise a crosslink item against a
  page with a long prose section; assert `content_data` length is unchanged
  apart from the inserted anchor (`bugs_open/178`'s "How to verify a fix"
  section has the exact query). Do NOT verify by the item reaching
  `complete` — that already happened on the broken path.
- **The shrink guard (already live) is your safety net during development**:
  if the fix is wrong, the guard will refuse the save loudly
  (`save_refused_incomplete`) rather than silently losing content — check
  that queue, don't just watch for `complete`.
- **Item 2 stays untouched**: the council's tracked rule is that a FOURTH
  bespoke floor on `save_page_sections` (beyond the existing three) triggers
  the unified content-loss detector design as its own submission — this fix
  is not that floor (it's upstream, at the writer's input), but if the
  implementation is tempted to add a fourth guard instead of fixing the
  input, stop and re-read `bugs_open/178`'s tracked-deferral note.

## OPEN — in priority order for whoever continues

1. **THE FIX ABOVE.** Root cause CONFIRMED 2026-08-03 (this session);
   nothing shipped against it in v1.0.1243. This is genuinely new code —
   start a fresh session/context for it given the size (new action file +
   tests + two call-site edits + register entries + council submission +
   roll + pod-verify, comparable in scope to the 177 lane's `74655b709`).
2. **Sibling writers unguarded** (`ApplySectionEditAction`,
   `rebuild_blog_listing_action.go`, `apply_gap_plan_action.go`,
   `deploy_tool_action.go`) — named in 178; the council's tracked rule: a
   FOURTH floor on `save_page_sections` is the trigger for a unified
   content-loss detector as its own design, NOT another bespoke guard. Not
   the same as item 1 above — this is about the SAVE-time guard's coverage,
   item 1 is about the WRITER's input.
3. ~~**`bugs_open/177`**~~ **DONE, verified live 2026-08-03** — see the fresh-build
   section above. No longer routed to this lane at all; owning lane
   (`bugfix_177_tool_content_items`) closes it out.
4. ~~**relojistas' deleted DefinedTermSet slot**~~ **CLOSED 2026-08-03** — a
   deliberate back-out by the traffic_probe lane, not a bug. Do not restore.
   Full trail in `bugs_open/178`'s 2026-08-03 update and NOTES.
5. Watch list: `bugs_open/169` part A (spawn hang, unmeasured) · loader's
   dependency subquery is SITE-SCOPED (fix = Go + both queries together,
   unmeasured) · scheduler pre_query/selector asymmetry
   **[MEASURED 2026-08-03: inert]** — zero `approved`-status rows have ever
   existed fleet-wide; real in code, dormant until an approval flow is
   built. Evidence in NOTES 08-03 tail.

## Landmines specific to this lane (carry-forward + new)

- **The refusal queue is live now** — `save_refused_incomplete` items are the
  guard WORKING, not a new bug. Read `spec.reason` (names slot + before/after
  sizes) before routing anything at one. A legitimate large cut is resolved by
  `section_shrink_floor` on the step config (0 disables); a measurement-error
  refusal is transient — retry, do NOT tune floors.
- 284+285 only safe TOGETHER; never revert one alone.
- A dependency releases ONLY on complete/verified — wont_fix blocks for ever.
- Dispatch quiet spells: read `collected_data->'load_items'`
  (`item_count:0` + `rows_dropped:0` = the 176 signature), never
  time-since-last-claim.
- Induction recipe (if ever needed again): inflate the STORED
  `rendered_html` of one slot (md5-guarded, back up first), queue a
  `page_rerender` item with `reason='section_data_resolved'` +
  `page_name` from `pages.name` (NOT derived from a subdirectory URL), wait
  out the ~300s post-restart window, and mind idx_swi_dedup.
- `orchestration_states` retention is ~24h — both this session's diagnosis
  runs (`aece2920` for 178, `da59941f` for 177) were captured to
  `EVIDENCE_2026-08-03_*_verdict_*.json` in this directory before the reaper;
  read those rather than re-querying `orchestration_states`.
- **NEW — `page-build-handler`'s content writer has no edit channel at all**
  (the mechanism this handoff exists to fix). Full landmine entry:
  `LANDMINES.md#page-build-handler-s-content-writer-never-sees-a-page-s-own-stored-prose-unless-`
  (verifier dispatch `98ca06a2` — check its verdict in `doc_notes` categories
  `landmine-verification` before relying on the entry unread).

## Cold-start pointers

- This file → `NOTES_…` (08-03 tail, two entries: "root-cause 090 run
  dispatched" through "178 root cause: CONFIRMED") → `SUMMARY_2026-08-03_…`
  for the shrink-guard-era story (predates the 178 confirm).
- `bugs_open/178` carries the full root-cause writeup, citations, and the
  tracked consolidation rule for a hypothetical fourth floor.
- Council trail: correlations `e64f8576` (guard, 2 rounds), `98aa9103`
  (wording, 1 round) — both APPROVED, both pre-date this revision.
- Diagnosis trail: run `aece2920` (178, 5 iterations, UNVERIFIABLE — closed
  by first-hand code read per the 07-31 escape hatch) and `da59941f` (177,
  1 iteration, REFUTED — independently correct, superseded by the owning
  lane's fix). Both evidence JSONs in this directory.
