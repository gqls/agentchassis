# HANDOFF 2026-08-16 — bugfix_281 (tool audit by instance): CLOSED; what a fresh chat should know

**Read in this order:** this file → `README_where_we_are.md` (plain prose, incl. the corrections) →
`NOTES_tool_audit_by_instance.md` (evidence, newest at bottom) → `RUNBOOK_…` (queries) →
register **TL-042** (`docs026_concept_register/register/tool-lifecycle.md`) → `bugs_closed/281_…`.

## State (all `[MEASURED 2026-08-16]`)

- **bugs 281 is CLOSED** (moved to `bugs_closed/` 2026-08-16): both mechanisms fixed, live in
  `v1.0.1303` (built from `5e075a6f9`; pods now `v1.0.1304`, 10:41Z), measured at the artefact.
- Code: `25f92a967` (fix), `a41d11e30` (gofmt), `d7b2d9994` (council follow-up, `Council-Reviewed:
  360ae540…`), `fa661d5d2` (corrections). Seeds **425/426 applied** 08-15 17:17Z. Council **APPROVED**
  round 2 (round 1 = my submission-schema slip).
- First sweep on webdesign: 13 `ported_tool_fix` + 12 `audit_tool` (cap), 0 `improve_tool` at the
  shared component, negative controls clean. Instance pin proven on a live auditor run (Mind Map,
  orch `a6f7ac42…`). Fence seen refusing (09:59Z, induced by the 285 lane).
- Shared ported-page component: template restored (v3 content, 4,664 chars, `{{.body}}`), 0/115
  pending, poison page restored — all by the `bugfix_285_shared_template_write` lane on 08-15.

## Three things I got wrong (corrected everywhere, dated; WRONG_CALLS 2026-08-16 ×2)

1. "not propagated" — false for ONE page (`learn-ai-builders-content-first`, served the poison ~23.5 h
   via the improver's arbitrary DELIVERY target); the 285 lane found + restored it. My census showed
   the row and I named it away.
2. "0 tool PLANs" — wrong table (`doc_notes`); `doc_plans` has 143 tool PLANs, 14 for the ported 63.
   D1 routing stands on "no per-instance writeback exists", not a PLAN count.
3. "`hardcoded_colors` will catch the Mind Map" (the bug file's premise, adopted unrun) — its styles
   are `var(--…)`; bare-hex census 0. The LLM tier sees its defects (score 5/10, 4 high); the
   structural tier correctly files nothing for it.
   Also: `fix_component_template` is component-scoped, not a page-aware LLM rewriter (guard census
   corrected).

## Open threads that are NOT this lane's (each has a home)

- **tool-auditor runs stall at `create_items_loop_complete`** — PRE-EXISTING (20 RUNNING / 2 FAILED
  there before 425); strongly correlated with findings > `max_iterations` 10 (43 capped runs stuck vs
  1 completing; 3 uncapped also stuck). Items never complete → the claimed-item-timeout sweep
  re-dispatches (a fresh Sonnet audit each ~40 min) until `failed`. This is why the audit_tool lane
  reads 19 failed / 16 unresolved, and it bounds the value of the Tier-2 queue until fixed.
  **Filed `needs_diagnosis` RUN_CORRELATION `815322b9-6a9f-4407-9c82-5d6c7ade43c2` (2026-08-16
  ~15:20Z)** — read the verdict: `SELECT collected_data->'verdict' FROM orchestration_states WHERE
  correlation_id='815322b9-…'` (verdicts live in the diagnose-agent's collected_data). If CONFIRMED,
  the fix belongs in `loop_actions.go` (truncation/loop_complete handoff), council-scope, and is a
  fleet-wide win (every capped loop consumer). NOT started.

  > **CORRECTED 2026-08-16 (later, by the session that read the verdict):** the run came back
  > **REFUTED**, then stopped at its iteration cap (`status: UNVERIFIABLE`) — so it refuted the
  > hypothesis above and confirmed nothing in its place. Two things here were wrong.
  > **(a) "truncation/loop_complete handoff" is not the mechanism.** The diagnoser found
  > `ec046659…` with 14 findings against `max_iterations` 10 — truncated — sitting at
  > `complete`, so truncation plainly does not block the handoff.
  > **(b) The real cause is a 2^N blow-up of `collected_data`, and the correlation with the
  > cap is a symptom of it, not a clue to it** — hitting the cap just means running the
  > maximum number of doublings. A `loop_complete` substep is injected once per iteration and
  > re-runs the WHOLE-loop aggregator from inside each one, nesting every earlier iteration's
  > aggregate into its own. `tool-auditor`'s `collected_data` reaches **22 MB avg / 29 MB max**
  > and the run dies. Measured on three agent types incl. `build-dispatch-loop` (the fleet
  > dispatcher, 13 MB), with a clean control that does not double. Full evidence and fix
  > candidates: **`bugs_open/289_…_loop_complete_substep_re_aggregates_…`**.
  > The diagnoser's own next lead — a Kafka `context canceled` — is a **red herring**: those
  > rows are logged against `process_item_iter_N_call_handler`, the *parent* dispatch loop's
  > step, not `tool-auditor`'s loop, and there is one of them against 31 dead rows.
  > Fresh `090` filed on the correct mechanism: RUN_CORRELATION `12ffad7c-a7b2-4955-b531-554f07650598`.
  > **What caught it:** not the loop — it refuted the wrong idea and ran out of iterations.
  > What caught it was comparing the three same-shaped loop consumers by
  > `length(collected_data::text)` and noticing `tool-auditor` was 1,000x the others.
- **Per-instance fixer for ported tools** (TL-042 gap (b)) — the 285 lane's next item; also
  `audit_review` items are per PAGE (`item_key_suffix_field=tool_data.page_id`), so N findings on one
  page collapse into ONE review row (better than the old per-site key, still lossy; the full findings
  live only in the orchestration's `audit_result` for ~24 h). Per-finding keys or a spec carrying all
  findings is the follow-on — note it if you touch 425's create steps.
- **Tier-4 judge** re-derives the component by function → cannot file a fix for a ported tool
  (281 Finding B). Open, unowned.
- **Decomposition of the 63 ported tools** — owner decision; `PROPOSAL_2026-08-15_decompose_
  webdesign_tools.md`. Precondition 2 (set `component_level='tool'`) handed to the LMC B2 lane with
  the measured 17→4 eligibility drop.
- **LANDMINE candidates not yet appended to LANDMINES.md** (do it if you have the budget, then run
  `./scripts/landmines-verify-dispatch.sh`): (a) "the binary stamps ONE sha (the build HEAD) — grepping
  your own commit's sha in `/proc/1/exe` returns 0 even when your commit is in the image; find the
  build HEAD by probing candidate commits with `grep -acF`, then `git merge-base --is-ancestor`";
  (b) "an audit_tool item that never completes is a Sonnet audit every ~40 min until `failed`".

## If you continue in a new chat

Nothing is owed on 281 itself. Highest-value next steps in this lane's neighbourhood: (1) read the
`815322b9` diagnosis verdict and fix the loop stall (fleet-wide); (2) the per-instance fixer /
per-finding review keys (coordinate with `bugfix_285_shared_template_write` — check their NOTES
mtime first); (3) the two LANDMINE entries above.
