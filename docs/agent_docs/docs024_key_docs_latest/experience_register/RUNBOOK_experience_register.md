# RUNBOOK — experience register

Commands/queries proven useful so far. Update HERE when one changes.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Travelling-docs substrate

Current plan for a subject (exact-key only — doc_plans has NO metadata column, NO structured
search; that is why the register needs its own table):

```sql
SELECT subject_type, subject_key, is_current, created_at, left(body, 200)
FROM doc_plans WHERE subject_type='<type>' AND subject_key='<key>' AND is_current;
```

Read the LIVE subject_type CHECK (do not trust docs — 163/184 both changed it):

```sql
SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint
WHERE conname IN ('doc_plans_subject_type_check','doc_notes_subject_type_check');
```

Gotcha: the DB CHECK is only one of the enforcement points — it was FOUR when this was
written (2026-07-24) and is now **two**. `bugs_closed/064` (fixed, live on v1.0.1156, closed
2026-07-25) single-sourced the Go side into `validDocSubjectTypes` in
`platform/orchestration/actions/doc_subjects_common.go`, consumed by both `docResolveSubject`
(shared by write_doc_plan + append_doc_note + load_doc_context) and the
`persist_diagnosis_note` gate. A migration-lockstep test fails the build if a migration
widens the CHECK without the Go entry — so adding `experience-pattern` means **both**, in one
change, or the build gate stops it.

## Bug filing

Next free number = max across BOTH dirs + 1 (numbering is one shared sequence, never
reassigned; several numbers are duplicated by historical accident — resolve by slug):

```
ls bugs_open/ bugs_closed/ | grep -oE '^[0-9]+' | sort -n | tail -1
```

Ownership check before routing work at an existing bug: `scripts/who-owns.py <number|slug>`.

## Later phases (do not fire yet)

- Experience-plan compose+council for one site experience (P3+, gauntlet session owns the
  vonc pilot): `092_TRIGGER_experience_plan.sh <domain> <experience_key>` — parked rule: only
  after tools-api is deployed + smoke-POSTed, liveness evidence via the 197 compose-decisions
  block.
- Council gate for the P2 platform change-set:
  `097_TRIGGER_council_review_v1.sh <submission.json>` — budget ~30 min (dispatch queues
  behind the fleet); find the run by payload correlation, not the printed id.
- Component-selector scoring reference (the shape the register's selection copies):
  `platform/orchestration/actions/component_selector.go` SelectComponentByType.

## Harvest — verifying a live journey before taking a pattern from it (added 2026-07-26)

**Rule that makes the rest of this section worth running: harvest from the LIVE artefact, not
from the source in the repo and not from another session's verification log.** Both are
usually right and neither is the thing a visitor loads.

```bash
UA='Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36'
curl -s -A "$UA" --max-time 25 https://<domain>/<page> -o page.html -w 'HTTP %{http_code} bytes=%{size_download}\n'
curl -s -A "$UA" --max-time 25 https://<domain>/data/<feed>.json -o feed.json -w 'HTTP %{http_code}\n'
diff -q feed.json <repo copy of the feed>     # "is the live feed the one we think it is?"
```

Gotchas, each one real:
- **Send a browser User-Agent.** A 403 from a Cloudflare-fronted host has two senders: the
  application's own JSON `{"error":"origin not allowed"}` and Cloudflare's plain-text
  `error code: 1010` for a non-browser fingerprint. Without the UA you diagnose the wrong one.
- **Grep the served JS bundle for strings the behaviour CREATES**, never ones it merely uses.
  A snippet lives inside a concatenated bundle, so its presence must be proven by something
  unique to it — a style-element id it injects, a class it adds — with a positive control:
  ```bash
  curl -s -A "$UA" https://<domain>/assets/js/snippets.js -o snippets.js
  for s in '<id the loader injects>' '<class the loader adds>' '<internal symbol>'; do
    printf '%-40s %s\n' "$s" "$(grep -c "$s" snippets.js)"; done
  ```
- **A page that renders its list from a feed shows the TEMPLATE row in static HTML**, not the
  rendered rows. Counting `__item-title` in the fetched HTML counts slots in the hidden
  template, not entries on screen. Anything about rendered state needs a real browser (the
  owning workstream's Playwright harness, e.g. `gauntlet_dead_cta/p4_sources/verify_live.py`).

**Ground the component identities in the DB before writing them into an entry** — and read
`usage_count` while you are there, because it is the register's own case:

```sql
SELECT id, name, section_type, quality_score, usage_count, suitable_site_types
FROM content_components WHERE name IN ('<component>', …) ORDER BY name;
-- note: there is NO component_type column; suitable_page_types/section_type carry that duty
```

**Read an approved EXPERIENCE_PLAN (the level-4 unit a pattern feeds into):**

```sql
SELECT body FROM doc_plans
WHERE subject_type='experience' AND subject_key='<key>' AND is_current;
-- psql -At and redirect to a file; the body is ~14 KB of markdown with a ```criteria fence
```

**The criteria vocabulary — read it from source, do not trust any doc including this one:**

```bash
grep -n 'case "' platform/orchestration/actions/discovery_checks/check_tool_acceptance.go
grep -n 'case "' internal/adapters/browserrunner/run_checks_action.go
grep -n 'stepDelay\|settleDelay\|runDeadline' internal/adapters/browserrunner/run_checks_action.go
```
Tier 2 = `selector_exists|selector_count|interaction|asset_loads|page_status_ok`; Tier 4 adds
`no_horizontal_overflow`; steps are `fill|click|select` only; assertions run 300 ms after the
last step. Our own design doc got this list wrong on 2026-07-24 by quoting a summary instead.

**Before writing to an existing bug file, re-check WHERE it lives** — contents in context say
nothing about status, and a closed case moves directory:
```bash
ls bugs_open/ bugs_closed/ | grep '^064'
```

## Applying a migration when 19 other threads' files are also pending (2026-07-27)

**Never `--apply` in this shop without reading the pending list first.** The runner applies
*every* pending file in order, and on 2026-07-27 that was 20 files, 19 belonging to other
threads — some parked on purpose (`229`'s own probe says the state it targets was not found).

```bash
./scripts/migration/run-migrations.sh            # dry run — READ the pending list
```

Apply one file, then register it (an unrecorded hand-apply stays "pending" for ever and is
eventually replayed — `bugs_open/007`):

```bash
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/NNN_x.sql

./scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/NNN_x.sql \
  --note "applied by hand <date>; <what you verified>"
```
The ledger row is `(filename, md5 checksum of the file AS APPLIED, applied_by='record-only',
notes)` — so if you edit the file before applying, checksum *after* the edit.

**Every migration needs a guard block, and the guard needs proving.** README convention:
`DO $$ … RAISE EXCEPTION … $$` inside the same `BEGIN/COMMIT`, so a partial apply rolls itself
back. A trailing `SELECT` after `COMMIT` reports but cannot prevent. Prove it bites by running
it alone *before* the migration, when it should fail:

```bash
sed -n '/^DO \$guard\$/,/^\$guard\$;/p' docs/agent_docs/sql_for_agents/NNN_x.sql | \
  kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f -
# expect the RAISE, naming exactly what is not yet true
```

**Verify a widened CHECK with a negative control**, inside a transaction you roll back — a
positive alone cannot distinguish a widened constraint from a dropped one:

```sql
BEGIN;
SAVEPOINT s1; INSERT INTO doc_plans (subject_type, subject_key, body) VALUES ('<new>','__probe__','x');
ROLLBACK TO s1;                                     -- expect ACCEPTED
SAVEPOINT s2; INSERT INTO doc_plans (subject_type, subject_key, body) VALUES ('site','__probe__','x');
ROLLBACK TO s2;                                     -- expect REJECTED
ROLLBACK;
```

## Reading a council verdict — there are THREE outcomes, not two (2026-07-27)

A run can end with **no verdict at all**: killed by the stale-step reaper after 4h. It is a
`FAILED` row with no objections and nothing to act on, and it is **not** a REVISE.

```sql
SELECT current_step, status, error, created_at, updated_at FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
-- error LIKE 'reaper: stale EXECUTING_STEP%'  →  no verdict; resubmit, do not interpret
```
Check whether it is systemic before assuming it was you:
```sql
SELECT date_trunc('day', created_at) AS day, count(*) FROM orchestration_states
WHERE error LIKE 'reaper: stale EXECUTING_STEP%' AND created_at > now() - interval '7 days'
GROUP BY 1 ORDER BY 1;
```

**Keep the submission JSON in the workstream directory, not the scratchpad.** A session
scratchpad does not survive; rebuilding a submission means re-deriving every `grounded_in` quote.
Verify quotes are byte-exact before spending credits:

```bash
jq -r '.plan.grounded_in[]' <submission.json> | while IFS= read -r q; do
  grep -rqF -- "$q" platform/ internal/ || echo "MISSING: $q"; done
```

## Pod-grep for a symbol that has no caller yet

`ValidateExperienceCriteria` greps **0** in a binary that certainly contains the change: with no
caller, the linker drops it. **A dead-code symbol is not a deployment check.** Grep a string that
is reachable — for P2a, the `experience-pattern` literal in `validDocSubjectTypes`, consumed by
`docResolveSubject` — and always with a positive and a negative control.
