# RUNBOOK — bugs_open/081

Every command here was needed to get something right. Gotchas attached.

## Find a bug nobody is working on (the only check that worked)

`scripts/who-owns.py` reads COMMITS — it cannot see a session mid-fix, and with
~30 concurrent sessions it says "OWNED or recently active" for nearly everything.
Counting bug FILENAMES in transcripts is worse: every session runs
`ls bugs_open/`, so listings swamp the signal and every bug looks equally hot.

Count `bugs_open/NNN`-shaped references and rank **ascending**:

```python
import os,glob,re,time,collections
d='/home/ant/.claude/projects/-home-ant-projects-agentchassis/'
me='<your-own-session-id>.jsonl'
now=time.time()
tx=[f for f in glob.glob(d+'*.jsonl')
    if os.path.basename(f)!=me and now-os.path.getmtime(f)<240*60]
pat=re.compile(r'bugs_open/(\d{3})|bugfix[_ ](\d{3})|bugs_closed/(\d{3})|bug[s]?[( ](\d{3})')
heat=collections.defaultdict(int)
for f in tx:
    txt=open(f,errors='ignore').read()
    for m in pat.finditer(txt):
        for x in m.groups():
            if x: heat[x]+=1
# then: sorted(open_bug_numbers, key=lambda n: heat[n])
```

**Then confirm with the CODE SYMBOLS**, not the number — a lane working your file
may never type the bug number:

```bash
grep -c 'resolveStorageURIFromAsset\|apply_gap_plan_action' ~/.claude/projects/*/[0-9a-f]*.jsonl
```

That is what caught `155` (57 hits, live session) and `162` (`repair_step`×86)
after both had been picked as "unowned".

## Is 081 still valid? (both halves — code AND data)

```bash
# the upsert must still lack page_type in DO UPDATE
grep -n "ON CONFLICT (site_id, name)" -A 6 platform/orchestration/actions/apply_gap_plan_action.go
# the retype arm must still exclude deployed
grep -n "build_status, '') <> 'deployed'" platform/orchestration/actions/discovery_checks/check_news_feed.go
```

```sql
-- the live instances
SELECT s.domain, p.name, p.page_type, p.url, p.build_status, p.status, p.sections
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.sections @> '["news-listing"]'::jsonb AND p.page_type <> 'news-index'
ORDER BY 1,2;

-- the loop
SELECT s.domain, swi.status, swi.created_at::date, swi.spec->>'page_name',
       (swi.spec ? 'retype_candidates') AS has_cands
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.item_type='missing_news_page' ORDER BY 1, swi.created_at;
```

**GOTCHA — `sections @> '["news-listing"]'` inside a `kubectl exec ... psql -c`
needs the inner double quotes escaped** (`'[\"news-listing\"]'`), or the shell
eats them and the JSONB literal is malformed. The error is not obvious: psql
reports a type error on the `@>`, not a quoting problem.

## The scope measurement (this is what set the fix's boundary)

```sql
SELECT COALESCE(build_status,'(null)') AS build_status,
       count(*) FILTER (WHERE sections @> '["news-listing"]'::jsonb
                          AND page_type <> 'news-index') AS news_mistyped
FROM pages GROUP BY 1 ORDER BY 1;
--  deployed 5 · needs_rebuild 0 · planned 0
```

Re-run this before widening the guard. If a `planned` row ever appears, the
argument for scoping to `deployed` has expired.

## PREPARE the SQL before trusting the build

`go build` cannot parse a string literal. Every statement in the change was
checked against the LIVE schema first:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 <<'SQL'
PREPARE p081_item AS INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, page_id, priority, status, created_by, item_key, parent_item_id
) VALUES ($1,'content-gap-planner','build','mistyped_deployed_page','medium',$2,
          $3::jsonb,$4,40,'needs_human_review','content-gap-planner',$5,$6)
ON CONFLICT DO NOTHING;
SELECT name FROM pg_prepared_statements;
SQL
```

`handler_agent` is `NOT NULL DEFAULT ''::text` — omitting it from the column list
is fine; passing NULL is not.

## Tests

```bash
go test ./platform/orchestration/actions/ -run 'TestApplyNewPage' -v
go test ./platform/orchestration/actions/discovery_checks/    # the item-type ratchet
```

`sqlmock` drives the conflict branch by returning **no rows** from the INSERT —
that is exactly what `DO NOTHING ... RETURNING id` does when the name is taken.
The growth-budget SELECTs that run first are deliberately unexpected:
`CheckPageGrowthBudget` swallows its own Scan errors, so an unmatched call leaves
the counts at zero and the budget allows the page.

## OWED — verify live after the next chassis roll

A green build proves nothing: the build path is the one that already works.

```bash
# 1. confirm the binary carries the fix (grep a string the change ADDED,
#    with a positive control in the SAME exec — a roll is not evidence)
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "deployed_page_type_conflict"; \
   strings /app/agent-chassis | grep -c "content-gap-planner"'
# want: >=1 and >=1. A 0 on the first with a non-zero control means the image
# predates the commit — see bugs_open/153.
```

Then induce BOTH branches:

```sql
-- FIRING: re-drive the aao item (news is deployed as page_type='content')
UPDATE site_work_items SET status='triaged'
WHERE item_type='missing_news_page' AND status='detected'
  AND site_id=(SELECT id FROM sites WHERE domain='ai-agent-orchestration.com');
-- afterwards REQUIRE ALL THREE:
--   pages.title/sections for name='news' UNCHANGED (snapshot them first)
--   the item is status='blocked' with an error naming the conflict
--   a mistyped_deployed_page row exists at needs_human_review
```

```sql
-- CONTROL: a gap plan naming a page that does NOT exist must still create it.
-- dartsonline.com took this branch on 2026-07-29 and completed; any site with no
-- news-index page will do. REQUIRE: page created, page_created:true, item complete.
```

**Snapshot `title` and `sections` BEFORE re-driving.** The whole claim is that
they do not change, and without a before-value you cannot tell "unchanged" from
"changed back".

## Council

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
# SUBMISSION_CORR for this change: ccd4384c-aff9-45ed-80b2-01c3ced573bb
```

Find the run by PAYLOAD, not by the printed id, and budget ~30 minutes:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id'
    = 'ccd4384c-aff9-45ed-80b2-01c3ced573bb';

SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='ccd4384c-aff9-45ed-80b2-01c3ced573bb' AND kind='council_report'
ORDER BY created_at;
```
