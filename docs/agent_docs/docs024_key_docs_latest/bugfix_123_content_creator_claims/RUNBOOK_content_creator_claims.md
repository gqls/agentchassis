# RUNBOOK — bugs_open/123, content-creator claims

Every command that was hard to get right, with its gotcha attached. Change it
HERE when it changes, not in your scrollback.

---

## Is another thread on this bug? (the check `who-owns.py` cannot do)

`scripts/who-owns.py <n>` is advisory and, on this tree, **returns "OWNED or
recently active" for almost every open bug** — its verdict fires on any commit
touching the bug file inside 14 days, and the 2026-07-27 triage sweep touched most
of them in one commit. Read its *owning workstream* block; ignore the verdict line.

It is also **lagging by construction** — it reads commits, so a session mid-fix is
invisible. Ask the live transcripts instead:

```bash
cd /home/ant/.claude/projects/-home-ant-projects-agentchassis/
# what is each LIVE session actually working on? (top bug per transcript)
find . -name '*.jsonl' -mmin -360 -size +10k | while read f; do
  echo "=== $(basename $f | cut -c1-8)"
  grep -oE 'bugs_open/[0-9]{3}' "$f" | sort | uniq -c | sort -rn | head -3
done
```

**Gotcha: read the counts, not the presence.** Every session that runs
`ls bugs_open/` has *all* the numbers in its transcript. A count of 2–8 is that
listing; a lane shows 40–400. And check whether another session is running the
same sweep prompt you are — two were, on 2026-08-03:

```bash
# the first real user message of each live session
python3 - <<'PY'
import json,glob,os
for f in sorted(glob.glob('*.jsonl'), key=os.path.getmtime)[-30:]:
    for l in open(f):
        try: d=json.loads(l)
        except: continue
        if d.get('type')=='user' and not d.get('isMeta'):
            c=d.get('message',{}).get('content')
            t=c if isinstance(c,str) else ' '.join(p.get('text','') for p in c if isinstance(p,dict) and p.get('type')=='text')
            if 60<len(t.strip())<1200 and 'command-caveat' not in t:
                print(f[:8], t.strip().replace('\n',' ')[:160]); break
PY
```

## Is the bug still valid?

```bash
# 1. does content-creator still have no validation at all?
grep -rniE "validate|claims|evidence|fabricat|banned" internal/agents/contentcreator/
# expect: NO OUTPUT. One file, 828 lines.

# 2. is it still invisible to per-agent usage queries?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT agent_type, count(*), max(created_at) FROM llm_call_log
WHERE created_at > now() - interval '14 days' AND agent_type ILIKE '%content%' GROUP BY 1;"
# expect: content-quality-auditor / page-content-writer / content-reviewer /
# content-gap-planner. NO content-creator row.
```

**Does anything actually dispatch it?** This is the measurement 123 left untraced,
and it decides the severity:

```sql
SELECT type, is_active, COALESCE(is_snapshot,false) snap, deleted_at IS NOT NULL del
FROM agent_definitions WHERE default_config::text ILIKE '%content-creator%';
```

2026-08-03: three rows, **all `is_active=false` and all soft-deleted**. No live
workflow calls it. Do not filter this query on `is_active AND deleted_at IS NULL`
— that returns zero rows and reads as "nothing references it", which is a
different and much weaker statement than "everything that references it is dead".

## Reading the existing claims engine (do not reimplement it)

```bash
sed -n '90,250p'  platform/orchestration/datahelpers/claims_global.go   # the 10 fleet-wide patterns
sed -n '1,85p'    platform/orchestration/actions/save_sections_claims_guard.go  # the severity precedent
grep -rn "ScanAllBannedClaims" --include=*.go . | grep -v _test.go       # every enforcement surface
```

**Gotcha — the trap that cost the first hour.** The three link checks are called
`phantom_internal_links`, `dead_controls`, `misdirected_cta`; the WORK ITEMS they
raise are **singular** (`phantom_internal_link`, `dead_control`) or named something
else entirely (`cta_names_unknown_destination`). Querying `site_work_items` by
check name returns zero rows and looks like proof the check never ran. List what
the producer actually produced first:

```sql
SELECT item_type, count(*), max(created_at)::date
FROM site_work_items WHERE created_by='<the agent>' GROUP BY 1 ORDER BY 2 DESC;
```

## Where a site-less finding can be recorded

`agent_error_log.site_id` **is NULLABLE** — checked with `\d agent_error_log`,
2026-08-03 — and a NULL-site row is an established shape, not an invention:

```sql
SELECT site_id IS NULL AS site_is_null, count(*), max(occurred_at)::date
FROM agent_error_log WHERE occurred_at > now() - interval '7 days' GROUP BY 1;
--  f | 201 | 2026-08-03
--  t | 283 | 2026-08-02
```

So a producer with no site can still write a durable record. Read it back by code:

```sql
SELECT occurred_at, severity, error_message, context
FROM agent_error_log WHERE error_code = '<the new code>' ORDER BY occurred_at DESC LIMIT 20;
```

## Blast radius before shipping any pattern

Never argue about a pattern's false-positive rate — measure it over the real
corpus with the same engine:

```bash
go run ./cmd/claimscan -h        # read the flags first; -no-global isolates a new set
```

## Proving the deploy (Go changes only — DB config is live immediately)

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=content-creator-agent -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/<binary> | grep -c "<a string the change ADDS>"'   # expect > 0
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/<binary> | grep -c "<a string it REMOVED>"'        # expect 0
```

**Both directions, every replica.** A positive control alone proves the pipeline,
not your spelling (`bugs_open/153`). `content-creator-agent` builds from its own
service target — confirm the binary path inside the image before trusting the grep;
a `grep -c` against a path that does not exist returns 0 and reads as "not shipped".
