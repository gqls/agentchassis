# RUNBOOK — bugs_open/326

Every command that was hard to get right, with its gotcha attached. When one
changes, change it HERE.

## Building your change against committed HEAD — USE THE SCRIPT

> **⚠ CORRECTED 2026-08-24 — THIS SECTION USED TO PASTE `git archive HEAD | tar` AND THAT WAS
> ACTIVELY HARMFUL. Do not restore it; the sanctioned tool already exists.**
>
> ```bash
> scripts/verify-head-builds.sh                          # does committed HEAD still build?
> scripts/verify-head-builds.sh --with <file> [--test]   # build YOUR change against HEAD first
> ```
>
> **What the hand-rolled recipe cost, measured in this lane.** Each extract is ~**450 MB**, and
> the `rm -rf` in every pasted copy is the *setup* half — it clears the tree that run is about to
> use, so it only ever reclaims a tree of the *same name*, and each variant picks a new name. This
> session left **eleven** of them (`ov`, `ov2`, `base`, `mut`, `verify`, `headonly`, `headtest`,
> `trio`, `all3`, `pairtest`, `headkfTT`) = **5.0 GB**, reaped only when CLAUDE.md's new
> `scripts/verify-head-builds.sh` note pointed it out. **73 documents still spell the recipe out;
> 66 of them never delete anything.** This RUNBOOK was one of them.
>
> `/tmp` here is a **16 GB tmpfs, i.e. RAM**. A full one presents as
> `link: mapping output file failed: no space left on device`, which reads like a compiler fault
> and is not one. The script writes to disk, refuses a tmpfs target by filesystem type, and
> deletes its tree on exit. Reap abandoned scratch on **both** filesystems with
> `scripts/scratch-report.py [--days N] [--reap]`.

**Why you need HEAD at all, rather than `go build`:** the working tree is the union of every
concurrent session's uncommitted work. A commit that leans on another session's *untracked* file
compiles perfectly for you and breaks HEAD for everyone — and `make build-<service>` builds
committed HEAD, so the gap is invisible exactly when the missing piece is someone else's.

**Worked example from this lane, 2026-08-23.** The tree would not compile at all
(`applyCTARecompute`, mid-signature-change by another session — since landed, tree builds clean
as of 2026-08-24 09:00). Separately, my own copy of `load_work_item_actions.go` carried another
lane's **half-written** hunk: their caller passed an 8th argument to `applyWorkItemFailureLadder`
while HEAD's callee still took 7. So `--with` on that file fails against HEAD **through no fault
of your own change** — that is the tool working, and the answer is to coordinate, not to strip
their hunk and pretend. Confirm whose it is before you conclude anything:

```bash
git diff --numstat platform/orchestration/actions/<file>   # in the SAME breath as the commit
git show HEAD:<file> | sed -n '/^func <symbol>(/,/^) /p'   # what HEAD actually declares
```

⚠ **`cd` PERSISTS BETWEEN COMMANDS in this harness.** A `cd` for one check leaves every later
relative path wrong. Combined with `2>/dev/null` on a `grep`, that yields "no matches" from a file
that was never opened — how the coverage-ratchet check got a false all-clear here before being
redone. **Never `2>/dev/null` a grep whose absence you intend to believe**, and prefix with
`cd /home/ant/projects/agentchassis &&` after any `cd`.

## Mutation-proving a test (the only thing that makes a green run mean anything)

Apply the mutation in the OVERLAY, never in the tree:

Mutate the file **in the working tree**, run the test, then revert the mutation — the tree
compiles again as of 2026-08-24, so no extract is needed for this at all. (If you must isolate
from HEAD, use `scripts/verify-head-builds.sh --with <file> --test`; never a hand-rolled extract.)

```bash
python3 - platform/orchestration/actions/load_work_item_actions.go <<'PY'
import sys; p=sys.argv[1]; s=open(p).read()
old = "case newestAge < antiChurnWindowHours:\n\t\t\t\tif legacy {"
new = "case newestAge < antiChurnWindowHours:\n\t\t\t\tif true {"
assert old in s, "MUTATION DID NOT APPLY — pattern absent"   # <-- the check that matters
open(p,'w').write(s.replace(old,new,1))
PY
go test ./platform/orchestration/actions/ -run '...' -count=1   # then REVERT the mutation
```

**Assert the pattern applied.** A mutation that silently fails to apply produces a
passing run that reads exactly like "the test did not catch it" — same output, opposite
meaning.

The five mutations this lane proved, all caught: remove the `retry_after` append;
restore the legacy drop on arm A; restore the `unresolved` brand on arm B; delete the
kill switch; use a flat 3h interval instead of the window remainder.

## Dating a re-submission after `orchestration_states` has been reaped

Retention is ~24h. `domain-submitter` writes a `submission` spec **before** the
deduping step, so `site_specs` outlives the orchestration:

```sql
SELECT created_at, is_current FROM site_specs
WHERE site_id = '<site>' AND aspect = 'submission' ORDER BY created_at;
```

## Telling a real dedup from a brake suppression

⚠ **Do not ask whether an open row holds the key NOW.** That is a present-tense predicate
about a past event and it returns 0 both when the brake fired and when the index did.
Read the lifecycle columns alongside the event timestamp instead — full query in
`LANDMINES.md`, "…`deduped: true` does NOT mean an open item holds the key".

⚠ **The window keys on `created_at`, not `completed_at`.** The probe is
`MAX(created_at)` over terminal siblings. Measured 2026-08-23 on `garden-tools.uk`:
created 17:17:15Z, completed 17:44:59Z — **27m44s apart**. Reasoning from `completed_at`
puts the boundary in the unsafe direction.

## Auditing the classification

```bash
./scripts/audit-undeclared-recurrence.sh          # human-readable
./scripts/audit-undeclared-recurrence.sh --json   # for a diff
```
Exit 1 = findings, 2 = could not determine. **Read refusal from EMPTY stdout, never the
exit code** — `go run` folds the tool's exit 2 into its own 1. Baseline at the fix
commit: **19 findings over 194 live agents**.

## Council submission

⚠ **`097` prints its `SAVE: SUBMISSION_CORR=` receipt at `:260` and publishes at `:263`.**
The pod is named `kcat-cgate-$(date +%s)` — one-second resolution — so two sessions
submitting in the same second collide, `kubectl run` fails `AlreadyExists`, and **nothing
is published** after a convincing summary has already printed. This happened on
2026-08-23 (`94c196fa`, never dispatched; re-run as `f610741f`).

```sql
-- the only proof a submission is in flight
SELECT correlation_id, current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

**Read the trigger's LAST line, not its summary block.** Empty result + clean tail =
latency (~29 min under load), do not retry. Empty result + a non-zero `kubectl run` =
dropped dispatch, retry immediately and treat the first correlation as never existing.

## Verifying the fix at the artefact

Assert **at the row**, never at the orchestration status, which reads COMPLETED either way.

```sql
SELECT id, item_key, status, created_at, retry_after
FROM site_work_items
WHERE site_id = '<site>' AND item_key LIKE 'research%' ORDER BY created_at;
```

The negative control the bug itself demands: submit again while the first is still
`triaged` and assert **no** second row appears.

⚠ **"Nothing ran" is not "nothing was queued."** `build-pipeline-trigger`'s
`find_dispatchable_site` is FIFO by work-item `created_at`, one site per ~90s tick, and a
site with ANY `claimed` item is invisible until it clears — measured time-to-first-agent
on a live greenfield build was **24m52s**. So snapshot whether the site had a claimed
item at that instant, or a negative result is uninterpretable.

⚠ **An earlier version of this note said the picker walks sites in ascending `site_id`.
That was REFUTED by its own author** — 14 consecutive ordered samples that stopped being
ordered twenty minutes later; the selector never mentions `site_id` except as a final
tie-break.

## Applying the migrations

```bash
# 572 — safe against the CURRENT binary, live on apply
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db < docs/agent_docs/sql_for_agents/572_build_chain_declares_recurrence_expected.sql
```

⚠ **573 is `_HOLD` and must NOT be applied before the roll carrying `on_dedup`.**
`create_work_item` is `StrictConfig` and `ValidateWorkflow` runs on **every message**, so
applying early fails *every* domain-submitter run for as long as it is applied — it takes
the front door down. Confirm the binary first by asking the service, not by inferring
from a roll:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <the 326 commit> <the stamped sha>
```
An empty grep means "not in range" (it is a startup line and it scrolls), **not**
"unstamped" — fall back to the binary probe with a present-and-absent control pair.
