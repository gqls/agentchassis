# RUNBOOK — gamesdesign.co.uk

Site id `e33263f4-74f8-494f-b191-546845dbbddf` · domain `gamesdesign.co.uk` ·
49 active pages as of 2026-09-02.

## Find the brand string anywhere on the site (case-sensitive)

```sql
-- specs (current only)
SELECT aspect, id FROM site_specs
WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND is_current
  AND data::text LIKE '%GameDesign.uk%';

-- pages + plan + components in one sweep
SELECT 'pages' t, count(*) FROM pages
 WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
   AND (title LIKE '%GameDesign.uk%' OR meta_description LIKE '%GameDesign.uk%')
UNION ALL
SELECT 'plan_pages', count(*) FROM site_plan_pages spp
 JOIN site_plans sp ON sp.id=spp.plan_id
 WHERE sp.site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND spp.title LIKE '%GameDesign.uk%'
UNION ALL
SELECT 'components', count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
 WHERE p.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
   AND (pc.content_data::text LIKE '%GameDesign.uk%' OR pc.rendered_html LIKE '%GameDesign.uk%');
```

Gotcha: `LIKE` is case-sensitive, which is what you want — lowercase `gamedesign.uk`
is legitimate (adopted_from fact + the P2P simulator cross-link on
guide-p2p-architecture). A case-insensitive sweep convicts both.

## Supersede a spec (never UPDATE in place)

Retire then insert, **separate statements, one transaction** — a chained CTE hits the
partial unique index `idx_site_specs_current (site_id, aspect) WHERE is_current`.
Worked example: `SQL_2026-09-02_brand_rename_APPLIED.sql` in this directory (already
applied — never re-apply). Manual-write conventions: `source='operator'`,
`created_by='claude-session-<slug>-<date>'`, `notes` names what superseded what and why.

## Rerender one page (deterministic re-assembly; does NOT rewrite content)

Publish with `scripts/kafka-publish-lib.sh` (OPP-009) — **never hand-rolled
`kubectl run -i … kcat -P`**, which drops ~4 in 5 at exit 0:

```bash
. scripts/kafka-publish-lib.sh
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --payload '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"e33263f4-74f8-494f-b191-546845dbbddf","domain":"gamesdesign.co.uk","page_id":"<PAGE_ID>"}}' \
  --header "orchestration_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "message_type=request" --header "client_id=demo_client" \
  --header "action=orchestrate" --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-gamesdesign" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Exit 0 = published (receipt seen) · 10 = retry now · 11 = verify, don't retry
(`kafka_verify_landing <correlation>`). Batch form:
`dispatch_rerenders.sh` pattern in the 2026-09-02 NOTES entry.

Pre-flight, both estate-standard: no dispatch within ~300s of a chassis pod restart
(`kubectl -n ai-persona-system get pods -l app=agent-chassis`), and check
`site_work_items` for open items touching your target first.

## Verify at the served artefact (never at a status)

`bugs_open/315` is live on THIS site's history (deployed_at written without
publishing; one page skipped by four rerenders) — so a `complete` orchestration or a
fresh `deployed_at` proves nothing. Per page, cache-busted, with a demand control and
a narrowness control:

```bash
u="https://gamesdesign.co.uk<PAGE_URL>?cb=$(date +%s%N)"
curl -s "$u" | grep -c "GameDesign.uk"        # want 0 (old brand, case-sensitive)
curl -s "$u" | grep -c "GamesDesign.co.uk"    # want >0 on branded pages (demand control)
# and on /guides/p2p-architecture/index.html only:
curl -s "$u" | grep -c "gamedesign.uk/games"  # want 1 (the deliberate cross-link survived)
```

Inside the deploy window (deployed_at within ~2 min) a fetch proves nothing — poll
until the condition holds rather than sampling once.
