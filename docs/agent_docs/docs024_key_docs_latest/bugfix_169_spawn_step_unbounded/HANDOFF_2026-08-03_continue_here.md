# HANDOFF — 2026-08-03 — bugfix 27 / bugs_closed/169 part A

**Read this first, then `NOTES_spawn_step_unbounded.md` (technical log) and
`README_where_we_are.md` (plain prose for the owner).**

## State in one line

**`bugs_open/169` is CLOSED and moved to `bugs_closed/`.** Both parts fixed and live.
Nothing is owed. **One thing is open and it is an owner decision, not a residual:
`architecture_review/RFC_011`.**

## What was done this session (in order)

| | what | evidence |
|---|---|---|
| 1 | Confirmed part A unowned | `who-owns.py` said "OWNED" — **lagging, it reads commits**. The live session touching 169 was the bugfix_154 lane whose own handoff note says *"169 part A (spawn hang) untouched"*. Nobody mid-edit on `spawn_actions.go`/`coordinator.go`; no commits there in 3 days |
| 2 | Filed `090` as the bug file instructs | `3ca53d45-4826-4935-96a3-a0af4d194d91` — ran 5 iterations, produced **5 `bundle` artifacts and no diagnosis**. Root cause therefore rests on first-hand verification, stated plainly per the 2026-07-31 ruling |
| 3 | Found the cause | no deadline anywhere in `continueExecution → executeStep → executeLocalAction → executeAction → handler(ctx, params)` |
| 4 | Shipped the fix | `fe34fd04f`, council `2c6800e6` **APPROVED**, register **RSH-004** |
| 5 | Routed the scope objection | `RFC_011` — the `architecture` seat was right that this belongs there |
| 6 | Verified live + **induced** | pod-grep both replicas, then made the guard fire against a real agent |
| 7 | Fixed what the induction found | `893cb6483` — the error said `on step ""` |
| 8 | Closed and moved | `e590866cb`, verified at HEAD with `git ls-tree` (exactly one line) |

## The fix, for anyone touching it

`platform/orchestration/coordinator.go` — `localActionContext()` derives a deadline that
`executeLocalAction` applies to the handler call.

- default **600s**; per-step `local_action_timeout_seconds`; **`<=0` disables** (Warn every
  time); **`DISABLE_LOCAL_ACTION_TIMEOUT=true`** kills it fleet-wide with no rebuild.
- a **malformed** value falls back to the DEFAULT, never to unbounded.
- `DeadlineExceeded` is wrapped to name action / step / elapsed, then routed through the
  existing `handleActionError` so `error_step` is unchanged.
- 8 tests in `local_action_deadline_test.go`; the three risky behaviours **proven able to
  fail by mutation**.

## The four traps, if you touch this area

1. **`timeout_seconds` is NOT this bound and must never be wired to it.** It is read,
   stored into `execCtx.TimeoutSeconds`, threaded through — and **read by no action**
   (all 271 checked). 53 of the 64 live steps carrying it are `call_agent`;
   `await_approval` carries 86400. **Grep for the READER, not the key.**
2. **`models.Step.Name` is EMPTY on the live coordinator path.** Use `state.CurrentStep`.
   The coordinator's own log at `:1325` has the same hole. **No fixture can reveal this —
   they all set `Name`.**
3. **A ~4h01m stall is not a hang** — it is `coordinator.go:831`'s
   `maxAge = TimeoutSeconds × 3`, which fires only when a message next arrives and never
   interrupts the goroutine. Print gap **boundaries**, not durations.
4. **A ctx deadline only cuts ctx-AWARE work.** An action that ignores `ctx` still runs to
   completion. Goroutine-abandonment was considered and rejected (leak + late-write into an
   already-failed step).

## If you need to re-verify it live

```bash
# in the binary, both replicas, with a control in the same exec
for P in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  kubectl -n ai-persona-system exec "$P" -- sh -c \
    'strings /app/agent-chassis | grep -c local_action_timeout_seconds; strings /app/agent-chassis | grep -c "Executing local action"'
done
```

**To induce it again** (this is how it was proven; smallest reversible surface):
snapshot `endpoint-health-checker`, set `local_action_timeout_seconds: 0.001` on
`check_health`, wait one 60s tick, revert — and **assert the revert with a `DO`/`RAISE`,
not by eye**. Expect a `FAILED` run whose error ends
`query endpoints: context deadline exceeded`. Full SQL in `NOTES`.

## The one open item — needs YOU, not a thread

**`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_011_a_fleet_wide_execution_deadline_on_the_step_seam.md`**

Should the coordinator impose an execution deadline on **every** local action by default?
Three costed options: (a) default ON at a generous constant — what is running now;
(b) default OFF, opt-in per step — which is what the 2026-08-02 RFC_010 ruling asks for
in general, but which walks straight into the inert-by-omission defect **RFC 006** was
decided on the same day; (c) bound it at a higher layer — which is what
`coordinator.go:831` already is, and it demonstrably did not solve this.

Council **approved** the change; the `architecture` seat's objection was about **how a
fleet-wide guarantee change reached production**, and per the 2026-07-28 ruling a scope
objection is not answered by better measurements. So it was routed, not rebutted.

⚠ **Cite RFCs by SLUG.** `RFC_009` and `RFC_010` are each **two different files** by
different sessions. Same collision hazard as migrations — landmined this session.

## Also from this session, unrelated to 169

- **RFC 006 decided by the owner** (one promoter, one owner) — implemented, live, and its
  detector now runs **daily** as `single-owner-carriers-check` (a CronJob, because a
  pre-commit hook cannot gate live config: at commit time the migration is unapplied).
- Missteps logged in `WRONG_CALLS.md`: shipping a seam one commit ahead of its register
  entry; backticks in `git commit -m` executing as shell (**use `-F <file>`**); and a
  mutation test that only produced a *build* failure, which proves nothing about a guard.

## Tree hazards live right now

Two other sessions have `internal/adapters/git/github_client.go` and
`platform/orchestration/actions/refresh_evidence_base_action.go` **in a non-building
state** in the working tree. HEAD is clean — verify with
`git archive HEAD | tar -x -C $(mktemp -d)` and build there, not in the shared tree.
