# RUNBOOK — bugfix_387 (deployed-and-404 headline + the NNN+ stand-in)

Every command here was hard to get right at least once. Gotchas attached.

## Curl a page the honest way
```bash
# NEVER compose the URL from pages.name — read it. (LANDMINES.md:7890; 3rd occurrence was this bug.)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F' ' \
  -c "SELECT s.domain, p.url, p.build_status FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='<domain>' AND p.name IN ('<name>')"
curl -s -o /dev/null -w '%{http_code}\n' "https://<domain><url-verbatim>"
# TWO controls, per domain, same run:
#   invented URL (catch-all detector):   /zzz-control-not-a-page-$RANDOM.html  -> must be non-200
#   known-good page at the SAME FORM:    a deployed sibling's pages.url        -> must be 200
# The 387 filing ran only the first; both its claim and its control were extensionless, so the
# control varied the PAGE while holding the URL FORM fixed — and the form was the whole defect.
```
Automated as `scripts/probe-page-url.sh` (WP4).

## Fleet sweep: every deployed page at its recorded URL (the 0/709 measurement, 2026-08-25)
```bash
S=<scratch>; kubectl ... psql -At -F' ' -c "SELECT s.domain, p.url FROM pages p JOIN sites s ON s.id=p.site_id \
  WHERE p.build_status='deployed' AND p.status='active' AND s.status IN ('active','deployed') \
  AND s.domain NOT LIKE '%.internal' AND p.url NOT LIKE '%#%' AND p.url NOT LIKE '%?%'" > $S/pages.txt
cat $S/pages.txt | xargs -P 8 -n 2 sh -c 'printf "%s %s %s\n" "$0" "$1" "$(curl -s -o /dev/null -w "%{http_code}" --max-time 20 "https://$0$1")"'
# Gotcha: sporadic `000` under concurrency = transport (TLS) flake, NOT a 404 — re-probe before counting.
# Gotcha: webdesign.uk 302s off-domain BY DESIGN (7 pages); idea.uk /privacy.html 301s. Neither is damage.
```

## The stand-in census (the detector's disconfirmable control)
```sql
-- regex deliberately EXCLUDES bare XX (siglo XX, relojistas) and [number] (quoted template, idea.uk guide)
SELECT s.domain, p.name, pc.slot_name FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND pc.rendered_html ~ '(\mN{2,}\+|\mNNN\M|\mX{2,}\+|\mN,NNN\M)';
-- 2026-08-25 pre-fix: exactly 1 row (ai-agent-orchestration.com model-directory hero). Post-WP2 expect 0.
-- Also run over site_components (chrome): 0 on 2026-08-25.
```

## The copy-rate measurement (how often the writer copies the exemplar)
```sql
SET statement_timeout='170s';  -- prompt_rendered is ~30KB/row; unbounded LIKE over weeks WILL time out
SELECT count(*) AS instructed,
       count(*) FILTER (WHERE response_text ~ '\mNNN\M') AS copied,
       count(*) FILTER (WHERE response_text ~ '\d{2,3}\+ (AI )?agents') AS stated_value
FROM llm_call_log
WHERE created_at > '2026-08-22' AND agent_type='page-content-writer'
  AND step_name LIKE '%generate_content%' AND prompt_rendered LIKE '%NNN+ AI agents%';
-- 2026-08-25: 137 | 14 | 0.  Zero \mNNN\M in ANY writer response before 08-22 (557's apply date).
```

## Read exactly what one writer call saw (by id — never re-derive from the spec)
```sql
SELECT position('NUMBERS — the facts list' in prompt_rendered) > 0 AS handwritten_block,
       position('NUMBERS (state only these' in prompt_rendered) > 0 AS composed_block,
       (SELECT count(*) FROM regexp_matches(prompt_rendered, '\m200\M', 'g')) AS value_occurrences
FROM llm_call_log WHERE id='<call id>';
-- Gotcha: regexp_matches() returns CAPTURE GROUPS — wrap the whole pattern in an outer () or you
-- get the innermost group and read garbage as your answer (cost this session two goes).
-- For 9ba94176 (the hero, 08-25 06:29Z): t | f | 0  — the writer never saw the value.
```

## Apply the WP2 migration (scoped!)
```bash
# LANDMINE: MIGRATIONS_DIR on its own line scopes NOTHING and applies ~100 other threads' files.
MIGRATIONS_DIR=<one-file dir> ./scripts/migration/run-migrations.sh            # dry-run first
MIGRATIONS_DIR=<one-file dir> ./scripts/migration/run-migrations.sh --apply
```

## Rebuild a tracker page via the framework (never hand-edit — owner ruling 2026-08-04)
Copy the shape of a real refresh item first (`SELECT * FROM site_work_items WHERE id='c73d25b1-bff7-4738-8da9-d3bc54d253d4'`),
then insert a `needs_page` row for the page. The 6-hourly refresh does this itself at ~00:24/06:24/12:24/18:24Z.
Gotcha: a `page_rerender` CANNOT fix this — the defect lives in `content_data`, and a rerender regenerates from it.

## Date a build against the binary before calling it evidence (the wrong call this lane made)
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'  # scrolls! see CLAUDE.md fallback
# A build that STARTED before the roll ran on the OLD code. orchestration_states.execution_started_at vs the roll time.
```
