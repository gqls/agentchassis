# RUNBOOK — bugs_open/125

Every command here had a gotcha attached when I first ran it. The gotcha is the point.

## Is anyone else on this bug? (the leading check, not the lagging ones)

`who-owns.py`, `git log` and `git status` all answer *"has anyone FINISHED?"*. They are
blind for the 20+ minutes between another session choosing a bug and its first Write —
which is exactly when a collision is created. This is the only leading signal:

```bash
# Grep the target's CODE SYMBOLS, not its bug number. A bug number appears in
# every session that ran `ls bugs_open/`, so number hits are pure noise.
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(ls -t *.jsonl | head -30); do
  n=$(grep -c "determinePageFilename\|git_deployer_actions" "$f"); [ "$n" != 0 ] && echo "$f: $n"
done
```
Your own session id will appear — subtract it. A count of 1–2 is usually a directory
listing; a count in the tens is a session reading the file. **Run this again immediately
before you start writing**, not only at the start.

## Re-measure the blast radius (do not quote the filed figure)

```sql
SELECT count(*) FILTER (WHERE url <> '/'||name||'.html') AS wrong_path,
       count(*) FILTER (WHERE url =  '/'||name||'.html') AS right_path,
       count(*) AS total
  FROM pages WHERE url IS NOT NULL AND url <> '';
```
Filed 07-28 as 280/151/431; on 07-31 it was **316/156/472**. Three days moved the
denominator by 41.

## Is any URL unsafe to use as a file path?

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE url LIKE '/%')     AS leading_slash,
       count(*) FILTER (WHERE url LIKE '%#%')    AS has_fragment,
       count(*) FILTER (WHERE url LIKE '%?%')    AS has_query,
       count(*) FILTER (WHERE url LIKE '%//%')   AS double_slash,
       count(*) FILTER (WHERE url LIKE '%..%')   AS dotdot,
       count(*) FILTER (WHERE url ~ '\s')        AS whitespace
  FROM pages WHERE url IS NOT NULL AND url <> '';
```

**Gotcha:** a non-zero `has_fragment` is not a "strip it" instruction. Find out what the
stripped value would collide with FIRST:

```sql
SELECT p.name, p.url FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE s.domain='<the site>' ORDER BY p.url;
```
On idea.uk the stripped value `/tools.html` is a **different page's** canonical URL, so
stripping would make one page's rebuild overwrite another's file.

## Which agents actually reach the resolver?

**Gotcha:** the obvious query misses them all. The `deploy_page` steps that carry
`page_field` live inside a LOOP step's sub-steps, so walking `workflow.steps` with
`jsonb_each` returns 19 rows and **zero** `page_field`s. Match the whole config as text:

```sql
SELECT type FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config::text LIKE '%page_field%';
-- pageflow-builder | page-rebuild | site-work-orchestrator
```

## Council submission

**The authority for the schema is the 097 script header (lines 22–40), not the
RUNBOOK prose.** `plan` is an OBJECT:

```json
{"rationale": "...", "submitter": "...",
 "plan": {"summary": "...", "edits": [{"file","symbol","operation","rationale","sketch"}],
          "grounded_in": ["..."], "risks": "a single STRING, never an array"}}
```
`risks` and `grounded_in` are **inside** `plan`. Getting that wrong is refused
client-side with `ERROR: .plan missing`, which does not say what is missing about it —
cheap, but only if you know where to look. `operation` ∈ modify|add|remove|config_change
(`create` is refused). ≤8 edits.

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  <submission.json>
# save SUBMISSION_CORR; budget ~30 min, not ~2
```

Find the run by PAYLOAD, never by the printed id:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report' ORDER BY created_at;
```

## Verify the fix on the pod (not the tag, not git)

`determinePageFilename` is an unexported function name and is *not* in the binary's
strings. Grep the **log message the change added**, with a positive control in the same
exec (`bugs_open/153`: a roll is not evidence your fix shipped):

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  echo "== $p"
  kubectl -n ai-persona-system exec $p -- sh -c \
    'strings /app/agent-chassis | grep -c "deploy path taken from the page.s canonical url"; \
     strings /app/agent-chassis | grep -c "determinePageFilename_negative_control"'
done
# want: 1 then 0, on EVERY pod
```

## Acceptance (the run the bug asks for)

`bugs_open/087`'s re-test, on a page that is **not** `rebuild_policy=owned` (the guard
that stopped it last time is correct and must not be worked around). Assert the **path**,
not merely `success: true`:

```sql
SELECT name, url, rebuild_policy FROM pages
 WHERE site_id = '<site>' AND COALESCE(rebuild_policy,'') <> 'owned'
   AND url <> '/'||name||'.html' LIMIT 5;   -- pick a subdirectory page
```
