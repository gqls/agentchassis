# HANDOFF — session "bugsearch 2", 2026-07-28. Two bugs closed, nothing blocked.

Entry point for a fresh chat. **Nothing here is a blocked deliverable** — both bugs
are closed, live and verified. What follows is the state to preserve, what is left
open and for whom, and what I would pick up next.

---

## 1. What was done

### `bugs_closed/124` — every manual diagnosis was running twice
Workstream: `bugfix_124_double_dispatch/` (PLAN · RUNBOOK · NOTES ·
README_where_we_are · SUMMARY · **HANDOFF_2026-07-28_continue_here.md** ·
REVIEW_2026-07-28_ctx_namespace.md).

`090_TRIGGER_needs_diagnosis_v1.sh` wrote its intake at `awaiting_diagnosis` **and**
published its own envelope, while `diagnose-pipeline-trigger` was enabled — so
`diagnose-dispatch-loop` claimed the same row ~60s later and ran a second full
diagnosis. Fixed three ways: the claim is the ticket; dispatch authority read from
live state; and `$ctx.`, a generic execution-context param namespace for
`query_database` (concept register **WFA-002**, chassis ≥ v1.0.1191, migration 258).

Verified on a real run: 0 orchestrations under the intake correlation, exactly 1
`diagnose-agent`, item closed itself. **Both filed mechanisms were refuted** and
corrected in 124 and 029.

Council **REJECTED** on scope; **owner ruled Option A** — keep the code, fix the
precedent. The rule is now in `CLAUDE.md` §"Platform seams and the ordering
exemption".

### `bugs_closed/106` — the register's coverage sensor had no cadence
Workstream: `bugfix_106_register_coverage_cadence/` (PLAN · RUNBOOK · NOTES ·
README_where_we_are).

The sensor existed and worked but ran only when a human remembered — the same
detected-by-coincidence mechanism the bug was about. Closed with
`check_register_coverage`, the 9th check in `scripts/pattern-check.py` (advisory,
pre-commit path, concept register **OPP-004**): a commit creating a workstream the
register has never heard of now says so, to the person creating it. Measured 4
fires / 1,500 commits (0.27%), all inspected, induced-gap verified in three arms,
and demonstrated on its own commit.

---

## 2. The one thing you must not break

**Migration 258 binds `$ctx.correlation_id`. A chassis below v1.0.1191 resolves it
to nil, fails `claim_item`, and the diagnose lane stops dispatching — silently,
with no failed row to find.**

**After every chassis roll:**

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "unknown execution-context field"'   # must be 1
```

Verified on **v1.0.1192, v1.0.1193→1194** — three consecutive rolls by other
sessions, all passed. **That is the committed-HEAD build rule working, not luck:**
every build since `af0cde87d` structurally contains the fix, so only a deliberate
`REF=` to an older commit or a rollback can break it. The protection is upstream of
the check; the check is what tells you the protection stopped applying.

A rollback below 1191 must revert 258 too. The pre-update snapshot is in
**`agent_definitions_backup`** — *not* `agent_definitions`; 258's own rollback
comment names the wrong table and the correction is in the 124 RUNBOOK (258 is
checksum-recorded and must not be edited).

---

## 3. Open, and whose it is

| # | what | owner |
|---|---|---|
| a | **`029`'s rate is inflated** by 124's duplicates. Do not re-derive it from `failed` `needs_diagnosis` rows without splitting on 2026-07-28. | whoever re-measures 029 |
| b | **The register can be complete in coverage and stale in CONTENT** — nothing detects that. Two instances on 07-28, incl. `SCH-012`'s `verify-later` stating an expected answer false for weeks, which helped hide 124. **Sample of two; deliberately not filed.** A third instance is the signature to act on. | unowned |
| c | **`092_TRIGGER_experience_plan.sh` is a LATENT 124.** Intake at a private status + its own publish; safe only while nothing claims `awaiting_experience_plan`. **A warning block sits in the script itself** above its INSERT. | experience_register |
| d | **Whether `$ctx.` gets a second consumer** — explicitly not ruled on. Nothing depends on it. | unowned |
| e | **Other services' overlays may have the same replica drift** as agent-chassis did (imperative `kubectl scale`, overlay says otherwise). Not checked. | unowned |

---

## 4. What I would pick up next

`bugs_open` is ~47 and moving. Checked as genuinely unowned this session and still
open: **`104`** (no fleet-wide claim patterns; the same watcherless-precondition
shape as 106, which 106's own file names as its sibling) and **`091`** — but 091 is
the bugs-sweep thread's declared next, so leave it.

`104` is the natural follow-on: 106 was *"a freeze with no watcher becomes
permanent"*; 104 is *"a deferral with no watcher becomes policy"* — a decision that
said "revisit at two sites" and was never revisited at eight. Same fix shape, and
106's closure gives you the sensor+ratchet+commit-trigger precedent to copy.

Run `scripts/who-owns.py <n>` first regardless — it reads commits, so a session
mid-fix is invisible; re-check `git log` too.

---

## 5. Mistakes from this session, so they are not repeated

- **My deploy silently scaled the chassis 2 → 1.** Imperative `kubectl scale` has a
  half-life of one deploy by anyone. Fixed in the overlay; **other services not
  checked**.
- **My deploy killed my own council round.** The gate runs its seats *inline on the
  chassis*, so a rollout restart ends any round in flight — it does not go FAILED,
  it sits at `EXECUTING_STEP` forever. Get the verdict, then roll.
- **My first claim guard could not fail.** `psql -t -A` prints `UPDATE 0` on a
  zero-row `RETURNING`, so the emptiness test always passed. Found *only* by
  staging the failure. **Two other threads had already hit this and fixed it
  privately in their own script comments** — now in 016b §9, where it should have
  been the first time.
- **A `[VERIFIED]` tag in 124 was earned off a print statement.** Logged in
  `WRONG_CALLS.md`.
- **I asserted "fleet-wide" and "every lane grew a bespoke action" without
  counting.** Both corrected against measurements — 0 live instances, and 32 of 34
  correlation-readers are not about SQL at all.
