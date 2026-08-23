# RUNBOOK — bugs_open/326

Every command that was hard to get right, with its gotcha attached. When one
changes, change it HERE.

## Build and test while the shared tree does not compile

**The working tree did not compile for the whole of 2026-08-23** — another session's
mid-signature-change to `applyCTARecompute` (`rerender_page_sections_action.go:550,552`).
`make build-*` is unaffected because it builds committed HEAD, so **anyone building an
image will not notice**, while anyone running `go test` on that package is blocked. That
asymmetry reads exactly like "my test setup is broken".

```bash
go build ./platform/orchestration/actions/   # exit 1 — theirs
D=$(mktemp -d); git archive HEAD | tar -x -C "$D"
(cd "$D" && go build ./platform/orchestration/actions/)   # exit 0 — so it is not yours
```

The workaround is `git archive HEAD` + overlay only your own files:

```bash
mkoverlay() {           # usage: mkoverlay <outdir> <repo-relative files...>
  OUT=$1; shift
  rm -rf "$OUT"; mkdir -p "$OUT"
  git archive HEAD | tar -x -C "$OUT"
  for f in "$@"; do mkdir -p "$OUT/$(dirname "$f")"; cp "$f" "$OUT/$f"; done
}
```

⚠ **Your own copy of a shared file carries the OTHER session's uncommitted hunk**, so
overlaying it onto HEAD can *still* fail to build — on 2026-08-23 the `bugs_open/345`
hunk in `load_work_item_actions.go` was **half-written** (caller updated to 8 args,
callee still at 7 in HEAD). Strip their hunk **from the build copy only, never from the
tree**.

⚠ **`cd` PERSISTS BETWEEN COMMANDS in this harness.** A `cd` into a subdirectory for one
check leaves every later relative path wrong. Combined with `2>/dev/null` on a `grep`,
that produces "no matches" from a file that was never opened — which is how the coverage
ratchet got a false all-clear here before being redone. **Never `2>/dev/null` a grep whose
absence you intend to believe**, and `cd /home/ant/projects/agentchassis` at the top of
any command that follows one.

## Mutation-proving a test (the only thing that makes a green run mean anything)

Apply the mutation in the OVERLAY, never in the tree:

```bash
mkoverlay /tmp/mut platform/orchestration/actions/load_work_item_actions.go ...
python3 - /tmp/mut/platform/orchestration/actions/load_work_item_actions.go <<'PY'
import sys; p=sys.argv[1]; s=open(p).read()
old = "case newestAge < antiChurnWindowHours:\n\t\t\t\tif legacy {"
new = "case newestAge < antiChurnWindowHours:\n\t\t\t\tif true {"
assert old in s, "MUTATION DID NOT APPLY — pattern absent"   # <-- the check that matters
open(p,'w').write(s.replace(old,new,1))
PY
(cd /tmp/mut && go test ./platform/orchestration/actions/ -run '...' -count=1)
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
