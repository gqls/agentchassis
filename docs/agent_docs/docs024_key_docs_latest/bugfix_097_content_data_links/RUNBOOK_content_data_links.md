# RUNBOOK — content_data link resolution (bugs_open/097)

Every command here was run on 2026-08-02 and each carries the gotcha that made it
hard to get right. Change a command HERE, not in your scrollback.

---

## R1 — Measure the standing debt with the SHIPPING CODE (authoritative)

**Do this rather than trusting R2.** A SQL predicate can count fields; only the
walker decides what is a candidate and what the value classifier throws out. This
is the same reasoning `bugs_open/093` used ("run with the SHIPPING code — not a
SQL approximation of it").

Two steps: dump the corpus, run the real function over it.

```bash
# 1. dump. The pages predicate MUST match loadValidPagePaths, which uses the
#    shared linkablePageStatusPredicate (prepare_link_context_action.go:54) —
#    status NOT IN ('deleted','archived'). Getting this wrong is the whole
#    measurement: robot-hands' /learning-center/index.html is ARCHIVED, so
#    including it would make a correct rewrite look like a false positive.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT jsonb_build_object(
  'components', (SELECT jsonb_agg(jsonb_build_object(
        'site_id', p.site_id, 'domain', s.domain, 'page', p.name,
        'component', COALESCE(cc.function, pc.slot_name, '(unknown)'),
        'data', pc.content_data))
     FROM page_components pc
     JOIN pages p ON p.id = pc.page_id
     JOIN sites s ON s.id = p.site_id
     LEFT JOIN content_components cc ON cc.id = pc.component_id
    WHERE pc.content_data IS NOT NULL AND pc.content_data::text <> '{}'),
  'pages', (SELECT jsonb_agg(jsonb_build_object('site_id', site_id, 'url', url))
     FROM pages WHERE status NOT IN ('deleted','archived'))
);" > /tmp/corpus.json
```

```go
// 2. run. A ~40-line main.go that unmarshals the above, builds one
//    datahelpers.NewPageURLIndex per site_id, and calls
//    datahelpers.RepairContentDataLinks(cp.Data, ix) per component.
//    NOTE it mutates its own copy — that is the point, it exercises the
//    rewrite arm — so never point it at anything but a dump.
```

Result 2026-08-02: `885 audited · 13 with findings · 19 rewrite · 33 phantom · 52 total`.

> **GOTCHA — the corpus is ~1.7 MB of JSON and psql will not truncate it, but a
> `head -c` while eyeballing it will.** Check `wc -c` before assuming a short file
> means few rows.

## R2 — The SQL approximation (for a quick fleet number only)

Useful for a one-line answer; **not** what to quote. It has to hand-reimplement
`NormalizePagePath` and `ClassifyLinkScope`, and the two will drift.

> **GOTCHA — a recursive CTE may reference itself only ONCE.** The obvious
> formulation (one `UNION ALL` branch for objects, another for arrays) fails with
> *"recursive reference to query "walk" must not appear within its non-recursive
> term"*. Put both child sources in a single `LATERAL` with a `UNION ALL` inside
> it, switching on `jsonb_typeof` with an empty literal for the other type:

```sql
UNION ALL
SELECT w.pc_id, c.k, c.v, w.depth+1
  FROM walk w,
       LATERAL (
         SELECT e.key AS k, e.value AS v
           FROM jsonb_each(CASE WHEN jsonb_typeof(w.val)='object' THEN w.val ELSE '{}'::jsonb END) e
         UNION ALL
         SELECT w.key AS k, a.value AS v
           FROM jsonb_array_elements(CASE WHEN jsonb_typeof(w.val)='array' THEN w.val ELSE '[]'::jsonb END) a
       ) c
 WHERE w.depth < 6
```

`NormalizePagePath` in SQL, mirroring `datahelpers/links.go`:

```sql
CREATE OR REPLACE FUNCTION pg_temp.norm_page_path(href text) RETURNS text AS $$
    SELECT CASE
      WHEN q = '' THEN '/'
      WHEN q = '/index.html' THEN '/'
      WHEN q LIKE '%/index.html' THEN COALESCE(NULLIF(rtrim(left(q, length(q) - 11), '/'), ''), '/')
      ELSE COALESCE(NULLIF(rtrim(q, '/'), ''), '/')
    END
    FROM (SELECT lower(split_part(split_part(href, '#', 1), '?', 1)) AS q) a;
$$ LANGUAGE sql IMMUTABLE;
```

## R3 — Which components hide a url inside an array?

The census that sized the class. **`content_components` has no `deleted_at`
column** — `\d content_components` first; the obvious `AND deleted_at IS NULL`
fails with a bare `ERROR: column "deleted_at" does not exist`.

```sql
WITH f AS (
  SELECT cc.function, cc.is_active, fk AS field_name, fv AS field_spec
  FROM content_components cc,
       LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) AS e(fk, fv)
)
SELECT f.function, f.field_name, string_agg(ik, ',' ORDER BY ik) AS nested_url_fields
FROM f, LATERAL jsonb_each(COALESCE(f.field_spec->'items','{}'::jsonb)) AS i(ik, iv)
WHERE f.field_spec->>'type' = 'array' AND ik LIKE '%url%'
GROUP BY 1,2 ORDER BY 1;
```

2026-08-02: **25 active component functions**.

## R4 — Read the durable record after a save

```sql
SELECT created_at, domain, agent_type, step_name,
       context->>'rewritten' AS rewritten, context->>'phantom' AS phantom,
       jsonb_pretty(context->'findings')
FROM agent_error_log
WHERE error_code = 'CONTENT_DATA_LINK_AUDIT'
ORDER BY created_at DESC LIMIT 5;
```

> **GOTCHA — this is a THIRD code.** `CONTENT_LINK_REPAIR_DETAIL` (what the gate
> changed in the markup) and `CONTENT_LINK_REPAIR_SKIPPED` (a pass that declined)
> are untouched. Querying the old codes will NOT show this pass, by design: they
> answer a different question.

## R5 — Prove the deploy, with a NEGATIVE control

`bugs_open/153`: a roll is not evidence your fix shipped, and a mis-cased grep
reads exactly like "not shipped".

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✓ NEW — 0 before the roll, >=1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'audited content_data internal links before persist'"
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'CONTENT_DATA_LINK_AUDIT'"
# ✓ POSITIVE CONTROL — the markup pass's marker, live since v1.0.1187, must stay 1
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'repaired dead internal links before persist'"
# ✓ NEGATIVE CONTROL — invented, must be 0 on every binary
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'CONTENT_DATA_LINK_INVENTED'"
```

Run it on **every replica**, not `deploy/agent-chassis` — `logs deploy/X` reads
one pod of N.

## R6 — Prove a guard is load-bearing (the mutation loop)

An assertion that a guard exists is not a demonstration that it does anything. For
each guard: break it, watch the named test fail, restore.

```bash
cp platform/orchestration/datahelpers/content_data_links.go /tmp/cdl.orig.go
# e.g. neuter the value-scope guard
python3 - <<'PY'
p='platform/orchestration/datahelpers/content_data_links.go'; s=open(p).read()
open(p,'w').write(s.replace('if ClassifyLinkScope(href) != LinkScopePage {','if false {',1))
PY
go test ./platform/orchestration/datahelpers/ -run ValueClassifier -count=1   # must FAIL
cp /tmp/cdl.orig.go platform/orchestration/datahelpers/content_data_links.go
```

> **GOTCHA — a mutation that does not COMPILE proves nothing.** Two of the five
> first attempts failed to build (a `case []int:` that broke type inference; a
> removed `sort.SliceStable` that orphaned the `sort` import) and a build failure
> is not a red test. Keep the import satisfied — `_ = sort.Strings` — and re-run.

> **GOTCHA — `-count=1`.** Without it a restored file can serve a cached PASS and
> you will conclude the guard is inert when you never re-ran it.

## R7 — Verify the change against committed HEAD, not the working tree

The tree carries other sessions' WIP; `make build-*` builds from `HEAD`.

```bash
rm -rf /tmp/headtree && mkdir -p /tmp/headtree && git archive HEAD | tar -x -C /tmp/headtree
for f in <your files>; do cp "$f" "/tmp/headtree/$f"; done
cd /tmp/headtree && go build ./... && go test ./platform/orchestration/{datahelpers,actions}/ -count=1
```

## R8 — Which live agents actually persist through this chokepoint?

**The obvious query under-counts by half, and looks authoritative doing it.**

```sql
-- ✗ NARROW — returns 3. Misses every agent that reaches the action via a loop step.
SELECT ad.type, s.key, s.value->>'action'
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS s(key,value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'save_page_sections';
--   page-build-handler | page-rerender | tool-recreation-handler
```

```sql
-- ✓ HONEST — text-match the whole config, then look at WHERE each one matched.
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%save_page_sections%'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Nine rows on 2026-08-02. Six are persisters — the three above plus
`pageflow-builder`, `page-rebuild` and `site-work-orchestrator`, which each reach
it inside a `build_*_loop` step (`action=loop`). The other three
(`council-gate`, `diagnose-agent`, `fix-proposer`) match on **prompt text** — a
footprint map in `select_panel` and a reviewer prompt — and are not persisters.

> **GOTCHA — always look at WHERE the text matched before counting it.** Both
> mistakes are available here: the narrow query silently drops the loop-dispatched
> half, and the broad one silently adds three agents that merely *mention* the
> action in a prompt. The figure "six agent types" that this family's docs repeat
> is **re-verified correct on 2026-08-02** by doing both and reconciling them.

## R9 — Is the standing debt actually converging, or just sitting there?

The claim "it clears as a side effect of ordinary operation" is only true at the
cadence the affected pages are actually re-saved. **Query it; do not assert it**
(the council's `debug_historian` seat caught exactly this, citing the matching
`WRONG_CALLS.md` precedent).

```sql
SELECT s.domain, p.name, p.build_status,
       count(pc.id) AS slots,
       count(pc.id) FILTER (WHERE pc.locked_at IS NOT NULL) AS locked_slots,
       max(pc.updated_at)::date AS last_write,
       (now()::date - max(pc.updated_at)::date) AS days_since
FROM pages p JOIN sites s ON s.id = p.site_id
LEFT JOIN page_components pc ON pc.page_id = p.id
WHERE (s.domain, p.name) IN (('gaswholesalers.com','supply-terms-and-eligibility'), ('idea.uk','about'), ...)
GROUP BY 1,2,3 ORDER BY days_since DESC;
```

2026-08-02: **2 to 23 days**, 0 of the 13 affected components locked. So
convergence is **opportunistic, not scheduled** — the fast pages clear in days,
gaswholesalers may sit for weeks, and nothing guarantees any page is re-saved at
all.

> **GOTCHA — a locked slot is NOT a blocker here, and the reason is worth
> knowing.** `save_page_sections_action.go:576` discards the *whole rebuilt
> section* for a locked slot (the row is kept out of the DELETE and the insert
> loop `continue`s), so the audit's rewrite is thrown away with it and frozen
> content can never be mutated. The same discard means the audit still RECORDS a
> finding for copy that was never persisted — a shared property with the markup
> pass, which does exactly the same thing.
