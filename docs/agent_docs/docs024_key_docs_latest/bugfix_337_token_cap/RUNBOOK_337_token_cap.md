# RUNBOOK — bugfix 337 (token cap / threshold management)

Every command that was hard to get right, with its gotcha.

## Read the live step config (never the seed)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'generate_template'->'config')
FROM agent_definitions
WHERE type='component-creator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
Gotcha: the budget lives at `config->ai_service->max_tokens`; there is NO
`default_config->prompts` for component-creator (0 rows) — the prompt arrives via
`getPromptWithPriority` (agent `prompt_template` / step `prompt_template`), so do not
conclude "no prompt" from the missing key.

## Census: cap-hit failures per step (all history)

```sql
SELECT agent_type, step_name, count(*) AS truncs,
       min(created_at)::date AS first, max(created_at)::date AS last,
       max(max_tokens) AS cap
FROM llm_call_log
WHERE NOT success AND error_message ILIKE '%reached the configured cap%'
GROUP BY 1,2 ORDER BY truncs DESC;
```
Gotcha: filter failures on `success=false`, never on non-empty `error_message` —
successful RETRY rows carry a marker in `error_message` too (bugs_open/119 precedent).

## Census: headroom on SUCCESSFUL calls (the leading indicator)

```sql
SELECT agent_type, step_name, count(*) AS calls, max(max_tokens) AS cap_sent,
       percentile_cont(0.95) WITHIN GROUP (ORDER BY output_tokens) AS p95_out,
       max(output_tokens) AS max_out,
       count(*) FILTER (WHERE output_tokens >= (max_tokens*9)/10) AS ge90pct
FROM llm_call_log
WHERE created_at > now() - interval '14 days'
  AND output_tokens IS NOT NULL AND max_tokens > 0
GROUP BY 1,2
HAVING count(*) FILTER (WHERE output_tokens >= (max_tokens*9)/10) > 0
ORDER BY ge90pct DESC;
```
Gotcha: truncated FAILED calls can carry NULL `output_tokens` (usage is only sometimes
recovered onto the error path), so this census understates pressure — pair it with the
failure census above; neither alone is the picture.

## The parked items (the live loss)

```sql
SELECT swi.id, s.domain, swi.item_key, swi.status, swi.attempt_count, swi.updated_at::date
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.item_type='needs_new_component' AND swi.error ILIKE '%reached the configured cap%'
ORDER BY swi.updated_at DESC;
```

## Open-work coverage check before routing anything at these targets

```sql
SELECT id, item_type, item_key, status FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected','failed')
  AND (item_key ILIKE '%credit-health%' OR summary ILIKE '%max_tokens%');
```

## Re-drive + verify (from bug file §How to verify; recipe origin RUNBOOK_311_fix.md)

Re-drive `loans-credit-health-check` on loanzy.uk, then assert: item completes with
`attempt_count=0`; `content_components` row for the section has `length(html_template)>0`
and contains `</section>`; and the served page gains real controls —
`curl -s https://loanzy.uk/tools/credit-health-check/index.html | grep -c '<input'`
above a pinned "before" of 0. Verify at the ARTEFACT, not the item status.

## Verifying a repaired tool page — two traps that both bit this lane (2026-08-22)

**1. Get the URL from the site, not from the page name — and always record the status.**
URL shape is PER SITE: `loanzy.uk` serves `/tools/<name>/index.html`, `loancalculator.co.uk`
serves `/tools/<name>.html`. A name-derived guess hit the latter's custom 404, which is 1,201
bytes of real HTML with a stable md5 and zero `<input>` — so it passed both a two-reads
stability check and a content check while being the wrong document.

```bash
curl -s -o /dev/null -w '%{http_code} %{size_download}\n' "$URL"   # refuse anything but 200
```

**2. Check what the tool IS before choosing the success predicate.** The bug file's recipe was
`grep -c '<input'` above 0. The component that actually shipped is a button-driven quiz:

```bash
b=$(curl -s "$URL")
printf %s "$b" | grep -c '<input'          # 0  — and the page is FINE
printf %s "$b" | grep -c '<button'         # 13
printf %s "$b" | grep -oE '<section[^>]*class="[^"]*"'      # the section is there
printf %s "$b" | awk '/<script/,/<\/script>/' | wc -c       # 4593 bytes of real logic
```
Better general predicate: assert the SECTION is present (`tool-<type>-section`) and that its
behaviour is present (inline script bytes, or handler names the component declares), rather
than one tag type borrowed from a calculator-shaped tool.

## Classify the rejections BEFORE naming the class (the query the re-scope needed)

```sql
SELECT CASE
         WHEN error_message ILIKE '%no site carries a site_specs aspect named%' THEN 'phantom_aspect'
         WHEN error_message ILIKE '%removes/renames%'                           THEN 'stranded_fields'
         ELSE 'other' END AS class,
       count(*) AS rows, count(DISTINCT work_item_id) AS items,
       min(occurred_at)::date AS first, max(occurred_at)::date AS last
FROM agent_error_log
WHERE error_code = 'component_validation_rejected'
GROUP BY 1 ORDER BY 2 DESC;
```
Gotcha: `agent_error_log`'s timestamp column is **`occurred_at`**, not `created_at`.
This is the query whose absence let a 3-of-101 class be named as the cause.

## Was the writer told anything? (the blind-advisory tell)

```sql
SELECT o.orchestration_id,
       length(o.collected_data->'existing_component'->>'field_names') AS advised_chars,
       left(regexp_replace(e.error_message, E'[\n\r]+', ' ', 'g'), 90) AS refusal
FROM agent_error_log e
JOIN orchestration_states o ON o.orchestration_id::text = e.orchestration_id
WHERE e.error_code = 'component_validation_rejected'
ORDER BY e.occurred_at DESC;
```
`advised_chars = 0` plus a stranding refusal is the signature. Gotcha: the join is on
`work_item_id`/`orchestration_id`, and older orchestrations are pruned — a small result set
means retention, not absence.

## Census keyed on DEMAND, not on function names (the one that does not over-count)

The wrong version joins `loader(section_type = <function name>)` and calls every miss a
deadlock. It counted 52; 21 of them had regenerated fine. Key on what is actually asked for:

```sql
WITH demand AS (
  SELECT DISTINCT spec->>'section_type' AS s FROM site_work_items
  WHERE item_type = 'needs_new_component' AND spec->>'section_type' <> ''
), loader AS (
  SELECT DISTINCT ON (section_type) section_type AS s, function AS loader_fn
  FROM content_components
  WHERE forked_from IS NULL AND is_active AND component_level = 'section'
    AND section_type IS NOT NULL AND section_type <> ''
  ORDER BY section_type, usage_count DESC NULLS LAST, updated_at DESC
) SELECT count(*) FILTER (WHERE l.s IS NULL) AS blind FROM demand d LEFT JOIN loader l ON l.s = d.s;
```

## Demand control for bugs_open/345's feedback path (do not infer it from the migration)

```sql
SELECT count(*) FILTER (WHERE collected_data->'input_data' ? 'last_error') AS carrying,
       count(*) AS total
FROM orchestration_states WHERE created_at > '2026-08-22 11:08:01+00';
```
Applied ≠ delivering. All-history control: the same predicate with no date filter returned
**0** before migration 555 and **5** after, which is what makes it a measurement.

## Mutation-testing on a SHARED tree — never `git checkout` to revert

`git checkout <file>` / `git restore <file>` restores the **whole file from the index**, so it
destroys every other session's uncommitted work in it. It cost this lane ~75 lines of another
lane's work, unrecoverably (unstaged work was never in git). `git stash` is hook-blocked for
this exact blast radius; the one-path form is not.

```bash
f=platform/orchestration/actions/<file>.go
cp "$f" "$SCRATCH/snap"                 # snapshot THIS instant, not at session start
perl -0pi -e 's/<the guard>//' "$f"     # apply the mutation
go test ./platform/orchestration/actions/ -run '<the tests>' -count=1
cp "$SCRATCH/snap" "$f"                 # restore from YOUR snapshot, never from git
```

## Pick the migration number when you WRITE the file, not when you plan it

Planned 561 (taken), agreed 562 (taken), then 563 and 564 went while the file was being
written. `ls docs/agent_docs/sql_for_agents/ | grep -E '^56' | grep -v ROLLBACK` immediately
before `mv`-ing into place.
