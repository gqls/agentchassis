# HANDOFF 2026-08-08 — the 207 lane is DONE; the work continues at 216 (then 217)

Written for a fresh session. Everything below is committed; nothing is uncommitted WIP.

## Where things stand

- **`bugs_open/207` — FINISHED.** Fixed (`9fa6f923b`, `messaging.RetryDisposition` /
  RSH-007), council APPROVED r1 (`Council-Reviewed: 155f4730-…`), LIVE on v1.0.1262,
  proven by induction (`1ada9cca1`). All three close criteria met; stays in `bugs_open/`
  per the owner's 08-06 ruling. Nothing left to do on it.
- **`bugs_open/216` — THE NEXT WORK.** The coordinator's response-driven recoverable arm
  re-arms (`retry_version`++) then refuses its own replay (`RETRY_PAYLOAD_UNAVAILABLE`)
  and fails the parent terminal. Until fixed, every upstream classifier win (195/197/207)
  buys a bookkeeping write instead of a retry — the ~30% deadline-exceeded prize is gated
  here. Mechanism, evidence, and fix candidates (ordered by what closes the door: **pass
  the claimed row through** — `ClaimAwaitedRequest` already RETURNS it) are in the bug
  file. The 090 verdict has been read and recorded IN the file: no clean verdict, filing
  stands — do not re-run 090 for the same claim; read the verdict block first.
- **`bugs_open/217` — WITH OR AFTER 216.** `notifyParentOfFailure`
  (`coordinator.go:3924/3937`) hardcodes `error_unrecoverable` and answers the parent
  FIRST on child-orchestration failures. Converge-or-decline decision, same shape as
  207's. Fixing it before 216 just converts hardcoded-terminal into re-arm-then-terminal.
  Sibling suspect named in the file: `TimeoutMonitor.sendTimeoutResponse`
  (`helpers.go:409`) stamps child TIMEOUTS terminal — liveness `[UNMEASURED]`.

## Unfiled by-product (grep bugs before filing)

The 090 run surfaced completed workflows whose results fail **"message validation
failed"** on `complete_workflow` delivery to the parent (its own corr `aee5853d`, several
`*/complete` steps: page-build-handler, page-rerender, page-content-writer). Not
explained by 216/217's mechanisms. Nobody has filed it.

## Gotchas the next session needs (all earned this lane)

1. **A retry_version bump is NOT a retry** — the bump is written by the same function
   that then refuses (`WRONG_CALLS.md` 2026-08-07; 016b §9 entry). Any acceptance
   criterion for 216's fix must assert the replayed request REACHES the target topic
   (consume it), not the counter.
2. **090 verdict recovery once the orchestration rows reap (~24h):** `llm_call_log`,
   `step_name='verdict'`, the RUN correlation (`spec.dispatch_correlation_id`). No
   verdict artifact is written to `diagnosis_artifacts` for subject-less runs, and the
   loop's outcome labels grade each round's REVISED hypothesis — read what was refuted.
3. **The induction recipe** (parked parent + flat failing child) is
   `SEED_test_207_probe.sql` in this dir + NOTES 2026-08-07: void topic must be created
   (auto-create off), dispatch payload must carry `child_agent_type`, the forged child's
   reply headers are DROPPED at intake so the response lands on legacy
   `system.generic.responses` (capture there; re-publish to
   `system.agent.generic.responses` to drive the parent). A transient child is an INLINE
   workflow with `query_database` + `local_action_timeout_seconds: 0.001`.
4. **216's fix touches `platform/orchestration/`** — council gate before/alongside the
   commit; name the coordinator's consumers; the claimed-row-pass-through is the
   candidate the file argues for. Check `who-owns` + live transcripts first — this lane
   (session `e68e25cd`) is ENDING with this handoff and will not compete.
5. Bug numbers move fast — two collisions in two days. Re-`ls` immediately before
   creating a numbered file, and commit quickly to claim the number.

## The paper trail

- Lane docs: this dir (PLAN, NOTES — the full induction story with timestamps and ids).
- Commits: `9fa6f923b` (fix), `95fa81d79` (notes), `1ada9cca1` (proof + 216/217 filings),
  plus this handoff's commit.
- Memory: `bugfix-207-sender-convergence.md` (auto-loads via MEMORY.md index).
- Register: RSH-007 (+ RSH-006 status correction and landmine-2 resolution).
