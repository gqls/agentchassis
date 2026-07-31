# RUNBOOK — bugs_open/092, the writer's link constraints

Every command here had a gotcha attached. When one changes, change it HERE.

## Is the bug still live? (the filer's measurement)

```sql
SELECT count(*) AS runs,
       count(*) FILTER (WHERE (collected_data->'link_context'->>'page_count')::int = 0) AS zero_pages,
       max(created_at) AS latest
FROM orchestration_states WHERE collected_data ? 'link_context';
```

**Gotcha:** `prepare_link_context` runs inside page-content-writer's OWN orchestration, not
the page-build-handler's. Filter on `collected_data ? 'link_context'` (the child rows) or
you will inspect the parent, not find the step, and conclude it never ran.

**Second gotcha:** the denominator moves. It was 20 on 07-26, 16 on 07-27, 26 on 07-31 —
`orchestration_states` is on a retention clock, so a falling count is a shrinking window,
not improvement. Quote the ratio, never the numerator alone.

## Where does this path actually keep its site id? (the query that decided the fix)

```sql
SELECT count(*) AS writer_runs,
       count(*) FILTER (WHERE collected_data->'input_data' ? 'site_id') AS has_input_site_id,
       count(*) FILTER (WHERE collected_data ? 'site_record')           AS has_site_record,
       count(*) FILTER (WHERE collected_data ? 'site_id')               AS has_toplevel_site_id,
       count(*) FILTER (WHERE collected_data ? 'db_sync')               AS has_db_sync
FROM orchestration_states WHERE owner_agent_type='page-content-writer';
```

**Why this one and not "does db_sync exist":** a query shaped as "is the configured field
missing?" only confirms the bug. This one also tells you **where the identity IS**, which is
what the fix needs. Wiring the DB read to the package's shared `extractSiteID` would have
resolved nothing on every real run — the fix would have failed the same silent way as the
bug. `input_data.site_id` is present on 26 of 26; the other three on 0.

## Who consumes this action, and what does the prompt actually do with it

```sql
-- consumers (there is exactly one)
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%prepare_link_context%' AND deleted_at IS NULL;

-- the template that interpolates the text, verbatim
SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
FROM agent_definitions
WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Gotcha:** the template lives inside a **loop sub-workflow**, not at the top level. A
`jsonb_each` over `workflow.steps` will not find it; you need the `#>>` path through
`process_sections_loop → config → sub_workflow → steps → generate_content`.

The guard is `{{if .link_context.link_constraint_text}}` and the heading `## Internal
Linking` is supplied **by the template**, so (a) an empty string removes the whole section
and (b) the action must not emit its own heading or you get two.

## The predicate question — are the two candidate page-sets the same?

```sql
SELECT status, count(*), count(*) FILTER (WHERE url IS NULL OR url='') AS no_url
FROM pages GROUP BY status ORDER BY 2 DESC;
```

2026-07-31: `active` 449 / `archived` 23, nothing else; `no_url` 0 everywhere. So the deploy
gate's `status NOT IN ('deleted','archived')` and `loadActivePagesForLinkContext`'s
`status='active'` are **the same set today**. Re-run before relying on it — a third status
appearing is what would silently split them.

## Blast radius before submitting (do not ask the reviewer to measure this)

```sql
WITH w AS (
  SELECT DISTINCT o.correlation_id,
         o.collected_data->'input_data'->>'domain' AS domain,
         NULLIF(o.collected_data->'input_data'->>'site_id','')::uuid AS site_id
  FROM orchestration_states o WHERE o.collected_data ? 'link_context'
)
SELECT w.domain, w.site_id IS NOT NULL AS resolvable,
       (SELECT count(*) FROM pages p
         WHERE p.site_id=w.site_id AND p.status NOT IN ('deleted','archived')
           AND COALESCE(p.url,'')<>'') AS linkable_pages
FROM w;
```

Answers "would the fix actually have fired on the runs we can see, and how big is the list
it would have produced" in one go: 8 distinct runs, resolvable on all, 31 pages each.

## Verifying the fix once the chassis rolls

```sql
SELECT created_at,
       collected_data->'link_context'->>'page_count' AS pages,
       collected_data->'link_context'->>'source'     AS source,
       collected_data->'link_context'->>'degraded'   AS degraded,
       length(collected_data->'link_context'->>'link_constraint_text') AS text_len
FROM orchestration_states WHERE collected_data ? 'link_context'
ORDER BY created_at DESC LIMIT 5;

SELECT occurred_at, severity, error_message, context
FROM agent_error_log WHERE error_code='LINK_CONTEXT_UNAVAILABLE'
ORDER BY occurred_at DESC LIMIT 10;
```

Pod-grep with a **positive control in the same exec** — a roll is not evidence your fix
shipped (`bugs_open/153`), and the image may predate your commit:

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "LINK_CONTEXT_UNAVAILABLE"; \
   strings /app/agent-chassis | grep -c "PrepareLinkContextAction"'
```

`0 0` means the grep is wrong; `0 N` means the image predates the fix; `N N` is the pass.

**Do NOT verify by reading the prompt template.** It is correct and always has been.

## Checking nobody else is on this bug (the number-grep is not enough)

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -name '*.jsonl' -mmin -180 -size +10k); do
  n=$(grep -c -E 'prepare_link_context|PrepareLinkContextAction|link_constraint_text' "$f")
  [ "$n" != "0" ] && echo "$f :: $n"
done
```

Grep the **code symbols**, not the number: sessions working the same class arrive via
`bugs_open/071` and never type "092", while three of the hits you do get are the same
`MEMORY.md` line loaded into three contexts. Then read the context around each hit — a hit
is a lead, not a verdict.

## Testing against HEAD rather than the shared working tree

```bash
SB=<your scratchpad>
git archive HEAD | tar -x -C $SB/headtree
cp <your changed files> $SB/headtree/<same paths>
(cd $SB/headtree && go test ./platform/...)
rm -rf $SB/headtree        # <-- DO NOT SKIP THIS
```

**Gotcha, paid for on 2026-07-31:** `/tmp` is a **16G tmpfs shared by every concurrent
session**, and one checkout plus its build cache is ~220MB. It hit 100% mid-session. The
symptom is Bash returning `the temp filesystem … is full … ENOSPC` — and the command may
still have SUCCEEDED, with only its output capture lost, which reads like a failure that
was not one. Delete the checkout as soon as the test run is done.
