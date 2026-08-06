# RUNBOOK — checks used to action the 2026-08-05 code review

Every command here was needed to get something right. Gotchas attached. When one changes,
change it HERE.

---

## R1 — Is a finding's lane still the owner? (re-run; the answer expires in minutes)

`scripts/who-owns.py <number>` is the start, not the end — it reads COMMITS, so a session
mid-fix is invisible. Two extra checks that actually decided this lane's work:

```bash
# Has the bug closed since the triage was written?
ls bugs_open/NNN* bugs_closed/NNN* 2>&1

# Did the owning lane ever pick the review up? (zero hits = the findings are unowned)
grep -rniE 'code.?review|triage|F1\b|F7\b|F15\b' \
  docs/agent_docs/docs024_key_docs_latest/bugfix_NNN_*/
```

**Gotcha:** a lane can close between the triage being written and being read. Both 194 and
195 closed within one minute of the triage commit.

## R2 — Which lane owns a FINDING (not a file)

```bash
git log -1 --format='%h %ad %s' -- <file>        # WRONG when two lanes share a file
git blame -L <start>,<end> --date=short -- <file> # right: blame the cited lines
```

**Gotcha:** this is not a refinement, it changes the answer. `save_page_sections_action.go`
last-committed by the 156 lane; lines 624-628 written by the 194 lane, 26 minutes earlier.

## R3 — Proving a symbol has no callers (compiler, not grep)

A grep is what a council reviewer cannot verify and what a reviewer objected to. Rename it and
let the compiler answer:

```bash
F=platform/errors/errors.go
[ -n "$(git diff --numstat -- $F)" ] && { echo "ABORT: dirty"; exit 1; }   # fail CLOSED
sed -i 's/func IsRetryable(/func IsRetryableZZPROOF(/' $F
go build ./... ; go vet ./platform/...          # any caller => undefined symbol
git checkout HEAD -- $F && git diff --numstat -- $F   # must be empty
```

**Gotcha:** proves it for THIS MODULE only. An external importer would not be seen. State that
scope when you cite it.

## R4 — Mutation testing that fails closed

```bash
F=<path>
[ -n "$(git diff --numstat -- $F)" ] && { echo "ABORT: dirty"; exit 1; }  # 1. clean at HEAD
sed -i 's/OLD/NEW/' $F
grep -q "NEW" $F || { git checkout HEAD -- $F; echo "ABORT: mutation did not land"; exit 1; }  # 2. it landed
go test ./<pkg>/ -run <TheGuard>                                          # 3. must FAIL
git checkout HEAD -- $F                                                   # 4. restore
git diff --numstat -- $F                                                  # 5. empty
```

**Gotcha:** steps 1, 2 and 5 are the whole point — a mutation that silently failed to apply
gives a green run that looks like proof. Mutate only files clean at HEAD so `git checkout`
restores exactly. (`WRONG_CALLS.md` records a session that lost a backup to a cleared
`$SCRATCH` and left a shared file mutated.)

## R5 — Reading agent_error_log without being fooled

```sql
-- WRONG: count(domain) counts the EMPTY STRING as present. So does `domain IS NOT NULL`.
-- And `WHERE domain IS NULL` sees only 1.3% of the no-domain rows [11:2xZ]: 3 of the 20
-- writers store '' instead of NULL, and one of the three is the coordinator's GENERIC writer,
-- so they produce most of the table. Use COALESCE(domain,'') = '' for "no domain". Full
-- diagnosis + the writer census in NOTES §16 / R13. (Do NOT repeat the older gloss that
-- site_id's NULLIF shows intent: that NULLIF is COMPELLED by the ::uuid cast.)
SELECT CASE WHEN domain IS NULL THEN '(NULL)'
            WHEN domain = ''    THEN '(empty string)'
            ELSE domain END AS domain_value, count(*)
FROM agent_error_log GROUP BY 1;

-- Which writer wrote a row: use ->> on the key, never jsonb::text LIKE '%"k":"v"%'
-- (jsonb renders a SPACE after the colon, so that pattern matches NOTHING).
SELECT context->>'classification', context->>'dropped_at', count(*)
FROM agent_error_log WHERE occurred_at >= '<t>' GROUP BY 1,2;
```

## R6 — Is a table actually reaped? (the check I got wrong first)

```bash
# The reaper is SQL in a DB column, NOT Go. Searching *.go returns clean and proves nothing.
grep -rniE "delete from <table>" --include=*.sql --include=*.sh --include=*.py .
```

```sql
-- enabled + last_triggered_at answer "is the SCHEDULER firing it".
-- last_completed_at does NOT: the agent's own workflow writes it, keyed by name.
SELECT name, enabled, interval_seconds, last_triggered_at, last_completed_at
FROM scheduled_tasks WHERE pre_query LIKE '%<table>%';

-- Then TEST the boundary. "oldest row is 30 days old" is produced BOTH by a working
-- 30-day reaper and by no reaper at all — it cannot discriminate.
SELECT min(occurred_at) AS oldest, now() - interval '30 days' AS line,
       min(occurred_at) > now() - interval '31 days' AS reaper_working
FROM <table> WHERE resolved = false;
```

## R7 — Finding every live carrier of a config key (walk the steps, don't guess)

```sql
SELECT ad.type, step.key AS step_name, step.value ->> 'action' AS action,
       step.value #>> '{config,<the_key>}' AS value
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config #> '{workflow,steps}') AS step
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND step.value #>> '{config,<the_key>}' IS NOT NULL;
```

**Gotcha:** select the `action` too. A key collision is only visible when you can see that two
carriers sit on *different actions* — that is what turned F7 from "a key exists twice" into
"one definition holds both meanings". (For nested `sub_workflow` steps a top-level-only walk
under-reports; see `bugs_open/144`.)

## R8 — Council submission schema (the trigger refuses a wrong shape, cheaply)

`plan` is an OBJECT, not an array:

```json
{"rationale": "...", "submitter": "...",
 "plan": {"summary": "...", "edits": [{"file","symbol","operation","rationale","sketch"}],
          "grounded_in": ["verbatim quotes"], "risks": "..."}}
```

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <file.json>
```

**Gotchas:** a top-level `plan` array fails with `ERROR: .plan missing`. Save the printed
`SUBMISSION_CORR`. Read the verdict **by your own correlation** — `doc_notes ... ORDER BY
created_at DESC LIMIT 1` returns whichever session landed last, which cost a confused minute
here:

```sql
SELECT metadata->>'decision', left(body, 4000) FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;   -- column is `body`, not `content`
```

## R9 — Before editing any file, is it another session's WIP?

```bash
git diff --numstat -- <file>          # non-empty => someone is mid-flight
git diff -- <file> | head -50         # is the code the finding cites on the + side?
```

**Gotcha:** a pathspec commit CANNOT exclude a same-file passenger. If the lines a finding
names are in someone's uncommitted diff, the finding is not yours to fix — verify it, write
down the corrected values, and leave it. (F13 was left this way; `WRONG_CALLS.md` was
committed WITH a declared passenger because append-only made that the lesser harm.)

## R10 — Did `save_page_sections` actually RUN? (never ask the error log)

`agent_error_log` holds **errors**. A save that succeeds writes nothing, so
`WHERE action='save_page_sections'` returning 0 means "no errors", **not** "no traffic" — and
the check it is meant to support is precisely a traffic denominator. Ask the save's side effect
instead; it DELETEs+INSERTs its rows, so `created_at` is fresh on every save:

```sql
SELECT count(*) AS rows_inserted, count(DISTINCT page_id) AS pages,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS rows_with_content_data,
       min(created_at), max(created_at)
FROM page_components WHERE created_at > '<roll>';
```

Then identify WHICH caller ran. **Use `owner_agent_type`** — it is a real column and it is
exact:

```sql
SELECT o.owner_agent_type, o.status, count(*), min(o.created_at), max(o.created_at)
FROM orchestration_states o
WHERE o.created_at > '<roll>'
  AND jsonb_path_exists(o.workflow_plan, '$.**.steps.*.action ? (@ == "save_page_sections")')
GROUP BY 1,2 ORDER BY 3 DESC;
```

> **CORRECTED 2026-08-06.** This step first fingerprinted callers by their
> `sections_metadata_field` value. **Do not** — `orchestration_states.owner_agent_type` states
> the caller outright. (A truncated `\d` is how the column got missed: it is the 30th of 36,
> so `\d … | head -25` hides it. Read the whole schema or `grep` it.) The fingerprint scheme
> was also wrong twice: `page_content.response.sections_metadata` is shared by **four**
> definitions, so it disambiguates only `page-rerender`; and `count(*)` over the LATERAL walk
> counts **step occurrences, not runs** — it reported 2 where the truth was 1 orchestration
> carrying 2 steps. If you must walk steps, `count(DISTINCT o.orchestration_id)`.

**Gotchas, four, each of which cost something here:**
1. `execution_metadata->'completed_steps'` is a **NUMBER, not an array of step names** — a
   `@> to_jsonb(step_name)` containment test is type-mismatched, returns false for every row,
   and **cannot come out otherwise**. It is also dead: 0 on runs that are `status='COMPLETED'`.
   Use `status`, not that counter.
2. Only the **last** save per page survives in `page_components` (29+1 runs → 13 pages here),
   so this is a floor on traffic, and intermediate saves leave no trace.
3. **`orchestration_states` reaps terminal rows at ~24h.** If the check is scheduled T+24h and
   its denominator lives here, the denominator ages out *as you read it* — **take it early.**
4. **`workflow_plan` is the RUNTIME plan, not the definition** — loop `sub_workflow` steps are
   already unrolled into `<loop>_iter_N_<step>` clones carrying the parent's config. So it
   over-counts *configuration* sites while being the only honest source for *traffic*. Census
   config in `agent_definitions.default_config` (R11); census traffic here. Getting these two
   the wrong way round is the same error as NOTES §11, from the opposite end.

## R11 — The F3 invariant, measured as config (a log grep cannot do this)

A pod-log grep only sees since the last restart — 35 minutes, the morning this was run. The
property is a config invariant, so query config and let the warn condition be a **column**
rather than something you eyeball across six rows:

```sql
SELECT ad.type, coalesce(step.value #>> '{config,sections_metadata_field}','(unset)') AS meta_field,
       coalesce(step.value #>> '{config,expects_no_sections_metadata}','-') AS expects_none,
       coalesce(step.value #>> '{config,html_field}','(unset)') AS html_field,
       CASE WHEN step.value #>> '{config,html_field}' IS NOT NULL
             AND step.value #>> '{config,sections_metadata_field}' IS NULL
             AND coalesce(step.value #>> '{config,expects_no_sections_metadata}','false') <> 'true'
            THEN '*** WOULD WARN ***' ELSE 'explicit' END AS f3_status
FROM agent_definitions ad,
     LATERAL jsonb_path_query(ad.default_config, '$.**.steps') AS steps,
     LATERAL jsonb_each(steps) AS step
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND step.value ->> 'action' = 'save_page_sections' ORDER BY 1;
```

**Gotcha:** the nested `jsonb_path_query(..., '$.**.steps')` walk is mandatory — a top-level
`jsonb_each` finds **3 of 6** (NOTES §11, `LANDMINES.md`). Expect exactly **6 rows, all
`explicit`**. Note `grep -c` exits **1** when the count is 0, so a `for` loop over pods "fails"
on the good result; read the printed number, not `$?`.

## R12 — Re-probe the pod after ANY roll, including one you did not do

`bugs_open/153` says a roll is not evidence your fix shipped. The converse bit this lane: a
**later** roll retires the pod names your proof cites (`v1.0.1254`'s pods were gone by the next
morning), so a successor quoting them is quoting nothing.

```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime'
# then, ONE exec per grep — batching times out on this binary and the image has no `strings`
kubectl exec -n ai-persona-system <pod> -- grep -ac "<literal your change ADDED>" /app/agent-chassis
kubectl exec -n ai-persona-system <pod> -- grep -ac "<deliberate misspelling>" /app/agent-chassis  # expect 0
```

**Gotcha:** re-probing proves *presence* only. Where no removed-unique-literal is available
(NOTES §12 explains why this change set has none), the misspelling control proves the probe
discriminates — it does not prove the old code is gone. Say which one you have.

## R13 — Census every writer of a shared table, and tie stored shape to writer shape

Two halves. Neither alone is enough: the code half tells you what *should* differ, the data half
tells you whether it actually does — and the data half is what could come out otherwise.

```bash
# 1. EVERY writer, not the one you were sent to. Exclude tests; 20 exist as of 08-06.
grep -rn "INSERT INTO agent_error_log" --include=*.go platform/ internal/ | grep -v _test.go

# 2. Read each one's VALUES clause — the column list and the values are on different lines,
#    so a one-line grep cannot pair them. Dump a window per hit:
grep -rn "INSERT INTO agent_error_log" --include=*.go platform/ internal/ | grep -v _test.go \
| while IFS=: read -r f l _; do echo "=== $f:$l"; sed -n "${l},$((l+13))p" "$f"; done
```

```sql
-- 3. THE DISCRIMINATING HALF. Group by a column each writer sets distinctively (here
--    error_code) and show all three shapes side by side. A clean partition confirms the
--    mapping; a per-code MIX refutes it. That is what makes it evidence.
SELECT error_code,
       count(*) FILTER (WHERE domain = '')     AS empty_str,
       count(*) FILTER (WHERE domain IS NULL)  AS is_null,
       count(*) FILTER (WHERE domain <> '')    AS real_dom
FROM agent_error_log GROUP BY 1 ORDER BY 2 DESC;
```

**Gotchas, in the order they bit:**

- **`grep -F 'NULLIF($'` — the `$` needs fixed-string mode.** Unquoted in a double-quoted bash
  string it becomes a bare `$`, and grep exits **2** with an empty result that a `| wc -l`
  happily reports as `0`. An error mistaken for a measurement of zero.
- **Do not classify the hits with `sed`.** I tried reducing each line to its `NULLIF(...)` form
  and the regex silently dropped `::uuid` on most, producing a tidy table that said the opposite
  of the truth. Print the real lines and read them.
- **Check whether the pattern you are calling a convention is COMPELLED.** `NULLIF($n,'')::uuid`
  is mandatory on a uuid column (`SELECT ''::uuid` → `ERROR: invalid input syntax`), so its
  presence there says nothing about intent. Test the cast before inferring a house style —
  and count the fleet before inferring one from a single site (I inferred "no convention for
  text columns" from one INSERT; there are **32** `NULLIF($n,'')` uses on text across 24 files).
- **A refactor is not a review of what it moves.** RFC_012 relocated this exact INSERT to a leaf
  package on 2026-08-06 and carried the defect verbatim. Any file:line for a shared writer is a
  *snapshot* — re-locate it (`grep -rn "INSERT INTO <table>"`) before citing one.
