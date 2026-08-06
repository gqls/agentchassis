# HANDOFF 2026-08-06b — 204 is LIVE (one gated step from closing); 189 is committed and awaiting a roll

Supersedes `HANDOFF_2026-08-06_204_fixed_189_next.md` (written a few hours
earlier, before the v1.0.1257 roll and before 189 was written). Same standing
owner brief as `HANDOFF_2026-08-05_next_bug_pickup.md` — re-read it for the
next pickup.

## Where the two bugs stand

**`bugs_open/204`** — build path could not resolve a positional slot name.
Fix `13252f714`, council **APPROVED** (corr `d3e232b8`). **LIVE at v1.0.1257**,
verified two ways and recorded in the bug file's final section: pod-grep of
three added strings on one pod per chassis-image deployment (agent-chassis,
business-intel, vet-intel) with a negative control returning 0; and a
read-only data proof that **57 of 57** loancalculator sections unresolvable by
name ARE resolvable by the stored-component_id route the fix reads. **Still
OPEN for one reason only: the behavioural canary has not run**, because it is
gated on 189.

**`bugs_open/189`** — a resolving save renames a positional slot, so the
locked-row guard misses and duplicates the row. Fix committed `92e14493b`,
council **submitted, verdict pending** (corr `87444080` — read it before
closing; it is the `Council-Submitted:` trailer on that commit, so `098`
credits it automatically on approval). **NOT live** — needs the next roll past
v1.0.1257. `stored_slot_name` greps 0 in the current binary, which is the
ready-made negative control for 189's own post-roll verification.

## The exact next steps, in order

1. **Roll.** The owner runs the whole-fleet release
   (`make release redeploy-agents ENVIRONMENT=production REGION=uk001`);
   `IMAGE_TAG` in the makefile must be bumped past v1.0.1257 first. Nothing
   below works until `92e14493b` is in the image.
2. **Pod-grep 189.** `strings /app/agent-chassis | grep -c "stored_slot_name"`
   → expect ≥1 (it is 0 today, measured). One pod per DEPLOYMENT running the
   chassis image, not per replica — one image serves 44 pods across at least
   three deployments.
3. **Apply the writer config.** `docs/agent_docs/sql_for_agents/023_page_content_writer_agent.sql`
   now ends with a targeted `jsonb_set` block adding
   `slot_name_from: "current_section.name"` to `render_section` and
   `render_from_template`. **Without it the BUILD half of 189 is inert** (the
   re-render half works regardless). It has a `DO $verify$ … RAISE EXCEPTION`
   guard, so a silent no-op fails loudly.
4. **Verify 189** (its own §how to verify): fire `section_data_resolved` on
   `tool-loan-vs-savings` (page `558f9f3f-…`, site `0162cde4-…`); assert
   **exactly 4** page_components rows, `tool-2` still locked at position 3 with
   `id`/`locked_at`/`locked_by` unchanged, and the `prose-*` slot names intact.
   Then close 189 → `bugs_closed/` (name BOTH paths on the `git mv` commit).
5. **Then 204's canary, un-gated**: fire the `voiceh-canary`-shaped
   `content_rewrite` at `guide-how-loans-are-calculated`; assert the prose
   changed against the baseline quoted in 204's §How to verify, and that
   **zero** `needs_new_component` items were filed (sweep query in the bug
   file; baseline today is zero open). Then close 204 the same way.
6. Unblocks after both: the owner's 2026-08-05 instruction to rerun
   loancalculator's copy through the framework in the H voice.

## What this session learned (beyond the two fixes)

1. **Read the verdict you cite.** I called 182's semantics "council-reviewed";
   only its SUBMISSION exists — no `council_report` ever landed for corr
   `80fbbe7d`, and `bugs_closed/182` says "verdict pending at close". The
   guardian seat caught it. Check: `SELECT kind FROM diagnosis_artifacts WHERE
   correlation_id='<corr>' AND kind='council_report'`. Logged in WRONG_CALLS.
2. **"Sole consumer" is a query, not a memory.** I asserted one consumer of
   `plan_sections` from the file header; live `agent_definitions` has **two**
   (page-build-handler and page-content-writer's 087 fallback). Same class of
   error, same round.
3. **After making a dead path live, trace where it terminates.** 204's fix made
   positional sections resolve; the SAVE they then reach had a filed, unfixed
   defect (189) whose trap 204 newly armed. Found by tracing compile→save
   AFTER the council round — disclosed in the PLAN doc and both bug files.
4. **A pathspec commit cannot save you from a same-file edit, in either
   direction.** My `v3_site_actions.go` half of 189 was swept into another
   session's `1d11827c1` (fix(152+155)) while I was writing the tests. Nothing
   was lost — it is at HEAD and forward-only holds — and my own commit says so,
   because a reader of either commit alone sees half a change. Earlier in the
   session the same file carried THEIR WIP as MY passenger; I had drafted a
   declaration for it before their commit made it moot.
5. **The trailer gate is real and it is right.** A placeholder
   `Council-Submitted: pending` was refused at commit time: the trailer is a
   join key, and forward-only forbids fixing a bad one by amend. Submit first
   (the correlation prints in seconds), then commit.
6. **Council latency today was ~8 minutes, not 29.** Both 204 rounds landed
   fast. Poll by payload correlation regardless; the 29-minute figure in
   CLAUDE.md is the budget, not the norm.
7. **Standing architecture question, twice raised (medium) and recorded in
   `doc_notes d9d67807`:** the tri-state id-resolution judgement now exists
   inline at two call sites (`rerender_page_sections` and `plan_sections`) with
   call-site-specific consequences. If a THIRD site needs it, factor the
   DECISION into one shared helper before writing it a third time.

## Other state worth knowing

- `doc_notes d9d67807` (subject `action/plan_sections`, category `decision`)
  records the id-first decision, the 039/041/095/182/204 family, the open
  architecture question and the 189 gate — written for the next fixer at the
  tooling_provenance seat's request.
- The seed file `023_page_content_writer_agent.sql` is a running log, not an
  idempotent script: it holds three historical full-workflow UPDATE blocks and
  a pasted transcript of a later `prompt_template` patch that those blocks do
  NOT carry. Re-running block 3 would revert that patch. This is why step 3
  above is a targeted `jsonb_set`, not a re-run.
- 204's own closure step for pod-grep was corrected mid-session: "one pod per
  ReplicaSet" under-counts because one image serves many deployments.
