# RUNBOOK — bugfix 314 (council gate scope)

## Test admission for FREE, in both directions (added by this fix)

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
```
Runs every client-side validation and the scope check, then stops before minting a correlation or
dispatching. Exit 0 = would be admitted; exit 2 = REFUSED; exit 1 = a validation error or a missing
scope fragment. **Refusals were always free** (the filter is client-side, before any dispatch); what
`DRY_RUN` adds is a free **positive** control, so the filter can no longer only be half-tested.

## The §6 control matrix, as actually run (2026-08-19)

Build one template and vary only `.plan.edits[0].file` — it must satisfy the five client-side
validator traps (rationale >40 non-space chars; every edit has file/operation/rationale/sketch;
non-empty `grounded_in`; sketch not comment-only; one whitespace-free repo-relative path per edit).

```bash
T='{"rationale":"scope-control test for bugs_open/314 verification, not a real submission — this exercises the admission filter only and dispatches nothing","submitter":"314-scope-lane","plan":{"summary":"scope control","edits":[{"file":"FILE_HERE","symbol":"x","operation":"config_change","rationale":"scope control","sketch":"UPDATE agent_definitions SET default_config = default_config; -- control"}],"grounded_in":["bugs_open/314 section 6 requires a positive AND two negative controls"],"risks":"none - refused or dry-run, entirely client-side"}}'
echo "$T" | jq --arg f '<path>' '.plan.edits[0].file=$f' > ctrl.json
```

| control | path | expected | got |
|---|---|---|---|
| NEG prose | `…/docs024_key_docs_latest/LANDMINES.md` | REFUSED, exit 2 | ✅ |
| NEG sidecar | `…/sql_for_agents/490_…_ROLLBACK.sql` | REFUSED, exit 2 | ✅ |
| **POS `_HOLD`** | `…/sql_for_agents/417_…_HOLD.sql` | **ADMITTED, exit 0** | ✅ |
| NEG `_SUPERSEDED` | `…/sql_for_agents/470_reaper_initialized_arm_SUPERSEDED.sql` | REFUSED, exit 2 | ✅ |
| NEG stacked | `…/sql_for_agents/494_…_HOLD_ROLLBACK.sql` | REFUSED, exit 2 | ✅ |
| NEG prose-in-migrations-dir | `…/sql_for_agents/README.md` | REFUSED, exit 2 | ✅ |
| POS migration | `…/sql_for_agents/490_internal_linker_….sql` | ADMITTED, exit 0, no FORCE | ✅ |
| POS code (regression) | `platform/orchestration/actions/database_actions.go` | ADMITTED, exit 0 | ✅ |
| **DISCONFIRMING** | POS migration against `git show HEAD:097…` (pre-fix) | REFUSED, exit 2 | ✅ |

That last row is the one that matters. The four negatives prove the check was not deleted; only the
old-script run proves the change **did** something. Get the old script with
`git show <sha>:docs/…/097_TRIGGER_council_review_v1.sh > /tmp/097_OLD.sh`.

## Fail-semantics inductions (each consumer, fragment absent)

A scratch git repo with the three consumers copied in and **no** `scripts/council-scope.sh`:

| consumer | required | got |
|---|---|---|
| 097 (admission) | exit **1**, loud — and NOT exit 2 | ✅ |
| 098 (report) | exit **1**, loud | ✅ |
| nudge (commit-msg) | exit **0**, **zero bytes** of output | ✅ |

097 exiting 1 rather than 2 is the load-bearing part: `exit 2` is "refused", and a missing
definition must not be reported as a scope decision.

## Drift guard, both ways (a control that cannot fail is decoration)

```bash
. scripts/council-scope.sh
council_scope_drift_warn "$(git rev-parse --show-toplevel)"   # expect SILENCE
# negative: doctor a scratch copy of the runner and re-point it
sed "s/SIDECAR_RE='_\[A-Z\]\[A-Z0-9_\]\*\\\\.sql\$'/SIDECAR_RE='_[A-Z][A-Z0-9]*\\\\.sql\$'/" \
  scripts/migration/run-migrations.sh > /tmp/dr/scripts/migration/run-migrations.sh
council_scope_drift_warn /tmp/dr                              # expect the WARN
```
⚠ The anchors are `grep -qF` **fixed strings** against `run-migrations.sh:65` and `:283`. If you
reformat either line in the runner — even harmlessly — this warns until you reconcile the fragment.
That is the intended cost.

## 098 before/after (the report's population changes — check it, don't assume it)

```bash
git show <pre-fix-sha>:docs/…/098_REPORT_unreviewed_commits_v1.sh > /tmp/098_OLD.sh; chmod +x /tmp/098_OLD.sh
NO_DB=1 /tmp/098_OLD.sh 14 | grep -E '^In-scope commits found|^### '
NO_DB=1 ./docs/…/098_REPORT_unreviewed_commits_v1.sh 14 | grep -E '^In-scope commits found|^### '
```
Measured 2026-08-19, 14-day window: **411 → 547** in-scope commits, UNREVIEWED **69 → 149**. Use
`NO_DB=1` for the shape — it skips the verdict lookup, so the buckets are trailer-shape only.

**Verify the post-filter excludes sidecar-only commits** (else the report accuses authors of skipping
a review the gate never offered):
```bash
. scripts/council-scope.sh
for sha in $(git log --since='14 days ago' --format=%H -- docs/agent_docs/sql_for_agents); do
  files=$(git diff-tree --no-commit-id --name-only -r "$sha")
  n=$(printf '%s\n' "$files" | in_council_scope | grep -c . || true)
  [ "$n" -eq 0 ] && echo "correctly excluded: $(git log -1 --format='%h %s' $sha | cut -c1-80)"
done
```
Measured: the three `494_*_HOLD.sql` commits are excluded; `5315c8a19` (migration 490) is included.

## Submitting a config-only change now

Just submit it — no `FORCE=1`, and no explanation owed:
```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
```
⚠ `FORCE=1` is still needed, and still needs saying so in the FIRST paragraph of the rationale, for
anything the widening does not cover: **tooling** (`scripts/`, the 097/098 scripts themselves),
prose, and sidecars. This fix's own submission was in exactly that position.

## ⚠ `_HOLD.sql` is IN scope — and the harness that told me otherwise

Corrected 2026-08-20 after the council's `editquality` seat objected at medium severity in the
round that approved the change. `_HOLD` is a migration held back from the runner **for ordering**
and applied **by hand**; it is live config reaching production. The first cut excluded it by
reusing `run-migrations.sh`'s `SIDECAR_RE`, which answers a different question. Census when
corrected: **157 `_ROLLBACK`, 12 `_HOLD`, 7 `_VERIFY`, 2 `_SUPERSEDED`**.

Edge cases the anchoring must keep getting right — all in the matrix above:
- `482_ROLLBACK_claim_timeout_exclusion.sql` — `ROLLBACK` at the **start** of the name is a REAL
  migration (the runner lists it under Pending, not Sidecars). IN scope. The exclusion regex is
  anchored to the end for exactly this reason.
- `494_…_HOLD_ROLLBACK.sql` — a rollback OF a hold. OUT.

### Two harness traps that stacked, and hid each other

1. **`local got=$?` must be the very next line.** A harness that did
   `printf '%s %s exit=%s' "$label" "$(basename "$file")" "$?"` reported **exit=0 for every
   case**, because the command substitution ran first and reset `$?`. Every negative "passed"
   while actually returning 0. The tell was that nothing failed, including cases that must.
2. **A sourced pipeline needs `{ …; } || true`.** The unclassified-suffix census ends in a
   `grep -v` that matches nothing when the census is CLEAN — exit 1, inherited by the command
   substitution, fatal under the callers' `set -euo pipefail`. **It broke the gate only when
   everything was healthy**, with no output at all. Induce the empty-result case; it is the one
   nobody tries.

```bash
# the induction that catches it — source under set -e with a clean census
bash -c 'set -euo pipefail; . scripts/council-scope.sh; council_scope_drift_warn "$(pwd)"; echo "SURVIVED"'
# and the guard still fires when a suffix is unclassified
touch /tmp/x/docs/agent_docs/sql_for_agents/503_thing_STAGED.sql   # in a scratch tree
bash -c 'set -euo pipefail; . scripts/council-scope.sh; council_scope_drift_warn /tmp/x'
```
