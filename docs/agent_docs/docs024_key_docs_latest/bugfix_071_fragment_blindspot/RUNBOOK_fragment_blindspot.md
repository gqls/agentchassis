# RUNBOOK — fragment blind spot (071's arm)

Every command here was got right once and is written down with its gotcha.

## Census: fragment-bearing links fleet-wide

Two shapes, and you need both — `/page.html#x` and bare `#x`. A single regex for
`href="(/[^"]*#[^"]+)"` finds only the first and reports 5 where there are 66.

```sql
-- path#fragment, on active shipped pages
SELECT s.domain, m[1] AS href, count(*)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL regexp_matches(COALESCE(pc.rendered_html,''), 'href="(/[^"]*#[^"]+)"', 'g') m
WHERE p.status='active' AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed')
GROUP BY 1,2 ORDER BY 1;

-- bare #fragment (exclude the no-ops: '#' and '#!' are dead_controls' remit)
SELECT s.domain, m[1] AS href, count(*)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL regexp_matches(COALESCE(pc.rendered_html,''), 'href="(#[^"]+)"', 'g') m
WHERE p.status='active' AND m[1] NOT IN ('#','#!')
GROUP BY 1,2 ORDER BY 1;
```

**Gotcha:** `site_components` has **no `component_type` column** — it is
`slot_name`. Chrome is `site_components`, unfiltered by page status.

## Does a fragment actually resolve? Probe the SERVED page

```bash
curl -fsS https://idea.uk/tools.html -o /tmp/p.html
grep -cE 'id=["'"'"']audience-check["'"'"']' /tmp/p.html     # 1 = resolves
```

**Gotcha:** grep for `id="x"` with the quotes. A bare `grep -c audience-check`
matches the href itself on the linking page and always returns ≥1 — the same
family as LANDMINES' "a grep for the class alone always returns at least 1".

## The blast-radius harness — run the SHIPPING code over real data

Do **not** re-implement the predicate in SQL: the SQL has to hand-copy
`NormalizePagePath` and `ClassifyLinkScope`, and the two answers have already
differed once on this estate (LANDMINES, 2026-08-02).

```bash
S=<scratchpad>
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c "
SELECT json_agg(row_to_json(x)) FROM (
  SELECT s.id::text AS site_id, s.domain,
    (SELECT COALESCE(json_agg(json_build_object('page_id',p.id::text,'name',p.name,'url',COALESCE(p.url,''),'status',p.status,
        'never_deployed',(p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed'))),'[]'::json)
       FROM pages p WHERE p.site_id=s.id AND p.status NOT IN ('deleted','archived')) AS pages,
    (SELECT COALESCE(json_agg(json_build_object('page_id',pc.page_id::text,'slot',COALESCE(pc.slot_name,''),'html',pc.rendered_html)),'[]'::json)
       FROM page_components pc JOIN pages p2 ON p2.id=pc.page_id
       WHERE p2.site_id=s.id AND pc.rendered_html IS NOT NULL AND pc.rendered_html<>'') AS page_components,
    (SELECT COALESCE(json_agg(json_build_object('slot',COALESCE(sc.slot_name,''),'html',sc.rendered_html)),'[]'::json)
       FROM site_components sc WHERE sc.site_id=s.id AND sc.rendered_html IS NOT NULL AND sc.rendered_html<>'') AS site_components
  FROM sites s) x;" > $S/fleet.json     # ~7.5 MB, 38 sites

go test -tags fleetharness ./platform/orchestration/actions/discovery_checks/ \
  -run TestFleetFragmentBlastRadius -v -fleet=$S/fleet.json
```

**Measured 2026-08-06: 67 fragment-bearing hrefs, 0 findings.**

**A zero is not evidence until you have induced a non-zero in the same run.**
Plant two controls (one bare, one cross-page) into a copy of the dump and re-run
— expect exactly 2:

```python
pc["html"] += '\n<a href="#zzz-no-such-anchor-control">x</a>'
pc["html"] += '\n<a href="/tools.html#zzz-planted-target">y</a>'
```

## Prove the tests can fail (mutation)

Back the file up, mutate, run, restore. Each must fail a **distinct** test:

| mutation | test that must fail |
|---|---|
| `Satisfies` returns true for any non-empty id | the firing test + both document tests |
| per-page doc built from `html` only, no chrome | `TestBareFragmentResolvesAgainstItsOwnPageAndChrome` |
| chrome surface judged with `resolvesOnPage` | `TestChromeFragmentJudgedAcrossTheWholeSite` |

## The claim-timeout lockstep

Registering a verifier is only half the gate. `go test
./platform/orchestration/actions/discovery_checks/` fails immediately with the
other half — the item type must also be excluded from the claimed-item-timeout
sweep, in **two** places:

```bash
# 1. the DECLARED list the Go test reads
grep -n "item_type NOT IN" docs/agent_docs/sql_for_agents/220_claimed_item_timeout_generic_evidence.sql
# 2. the LIVE column — read it BEFORE writing the migration; a replace() must
#    name the exact current string, and if it has drifted you would re-encode
#    someone else's missing entry (this is what happened to 305)
```
```sql
SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)')
FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
```

Apply `322` with the migration runner (dry-run first, scope the directory — the
runner takes EVERY pending file otherwise).

## Verify the arm is live after a roll

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c "
  strings /app/agent-chassis | grep -c dead_fragment_link;        # 0 before, >0 after
  strings /app/agent-chassis | grep -c phantom_internal_link;     # POSITIVE control: 9 today
  strings /app/agent-chassis | grep -c zzz_no_such_string_control # NEGATIVE control: 0"
```

Run it on **every replica** — measured on `v1.0.1257` (both pods) at 10:0x on
2026-08-06: `dead_fragment_link` 0, control 9, negative 0, i.e. the pre-fix state.

## Item-type spelling trap

The check is `phantom_internal_links` (plural); the item type is
`phantom_internal_link` (singular), and one check files **three** types. Querying
the check name returns 0 rows and reads exactly like "never fired". Take the
spelling from the `ItemType:` literal in the check's source.

## Make a verifier actually RUN (added 2026-08-08)

`VerifyDeadFragmentLinkResolved` is reachable only through `CompleteWorkItemAction`.
Two ways to get there; the first is deterministic and side-effect-free.

### A. Drive `complete_work_item` directly, with a one-shot agent

**The literal-in-config trap comes first, because it costs you a whole round.**
`ExtractActionInputs` resolves a config value only if it is a **multi-segment
dot-path** (`action_inputs.go:472-488`). A literal UUID in
`config.work_item_id` is silently not a value, and the action fails with
`missing required fields: [work_item_id]` — which reads like you omitted it.
**Pass values through the scheduled task's `input_data` and point config at them.**

```sql
INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
VALUES ('oneshot-frag-verify-probe','…','…','orchestrator','coordinator','experimental', true,
'{"workflow":{"start_step":"complete_item","processing_mode":"orchestrator","timeout_seconds":300,
  "steps":{
    "complete_item":{"action":"complete_work_item",
      "config":{"work_item_id":"input_data.work_item_id","result":"input_data.handler_result"},
      "next_step":"done","output_field":"completion"},
    "done":{"action":"complete_workflow","config":{"output_fields":["completion"]}}}}}'::jsonb);

INSERT INTO scheduled_tasks (name, interval_seconds, target_agent_type, target_topic, input_data, enabled, fire_message)
VALUES ('oneshot-frag-verify-probe-20260808', 86400, 'oneshot-frag-verify-probe',
        'system.agent.scheduled.requests',
        '{"work_item_id":"<uuid>","handler_result":{"probe":"…"}}'::jsonb, true, true);
```

Then **disable it the moment it fires** (`UPDATE scheduled_tasks SET enabled=false …`);
to re-fire, `SET enabled=true, last_triggered_at=NULL, last_completed_at=NULL`.

**The `result` you pass must carry no `response.status` of `failed`/`failure`/`error`,**
or completion gate 1 (`handlerReportedFailure`) short-circuits and the verifier never
runs — the guard is deliberately *before* the verifier. A result with no `response`
key at all passes.

Read the verdict at the item, not at the workflow:

```sql
SELECT status, attempt_count, left(error,160), jsonb_pretty(result->'_verification')
FROM site_work_items WHERE id='<uuid>';
```

Confirm the function body ran, not just the policy wrapper — this line is inside it
(`check_phantom_internal_links_fragments.go:290`) and past both queries:

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  kubectl -n ai-persona-system logs $p --since=10m | grep -o '"msg":"dead_fragment_link verifier[^"]*"[^}]*item_id[^,]*'
done
```

**Gotcha:** run it over **every** replica — `logs deploy/X` reads one pod of N, and
the probe lands on whichever pod took the message.

### B. Let the real dispatcher do it — note it may happen WITHOUT you

A refusal sets `status='triaged'` (`complete_work_item_verification.go:224-231`),
which makes a `detected` item **dispatchable**. `build-pipeline-trigger` runs every
120s and picks the oldest `triaged` item's site fleet-wide, so a refused item can be
picked up and handled a minute later on its own. Budget for that before refusing an
item you are not ready to have handled.

`build-dispatch-loop`'s `load_items` carries **no** pipeline/domain/handler filter —
`item_pipeline` is applied only `if pipelineFilter != ""` (`load_work_item_actions.go:635-673`).
It loads any pipeline for the site; `status IN ('triaged','approved')` is the only gate.

### ⚠ Before you dispatch `nav-link-fixer` anywhere

Its last two steps render a JS snippets bundle and `git_commit` it to **`gqls/sites`**
under `<domain>/assets/js/snippets.js`, and `render_js_snippets_for_site_action.go:86-94`
returns that `files` map **even when the site has zero snippets** — so the commit is
unconditional, and for a pool/internal domain it is junk in a shared repo that then
deploys to B2. Check what you are about to push:

```bash
gh run list --repo gqls/sites --limit 5 --json displayTitle,createdAt,conclusion
gh api repos/gqls/sites/contents/<domain> --jq '.[].path'
```

Also: `render_site_components` does **not** need site-assigned templates. On a site
where `fix_nav_link_templates` reports `"no header/footer component templates assigned
to site"` it still rendered both slots from generic templates. Do not predict a no-op
from the absence of templates.

## Assert a chrome-surface finding, with its own negative control

One `site_components` row carrying both anchors, so the silent half is disconfirmable
in the same run:

```sql
INSERT INTO site_components (site_id, slot_name, rendered_html, build_status)
VALUES ('<site>', 'footer',
  '<footer><nav><a href="#zzz-page-anchor-live">live</a>' ||
  '<a href="#zzz-induced-chrome-dead">dead</a></nav></footer>', 'deployed');
```

The site needs **at least one page with rendered `page_components`**, or rule 2 declines
to judge at all (`resolvesAnywhere` returns `judged=false` when `len(byPageID)==0`) and
you get a zero that means "no verdict", not "clean".

Expect exactly one item: `surface=site_component`, `pipeline=build`,
`handler_agent=nav-link-fixer`, `priority=30` (40 from `routeBySurface`, minus the
fragment arm's 10).
