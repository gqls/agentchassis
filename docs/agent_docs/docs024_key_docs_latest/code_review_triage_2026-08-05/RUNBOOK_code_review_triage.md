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
-- WRONG: count(domain) counts the EMPTY STRING as present.
-- LogAgentError inserts domain as a bare $2 (no NULLIF), unlike site_id's NULLIF($1,'')::uuid.
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
