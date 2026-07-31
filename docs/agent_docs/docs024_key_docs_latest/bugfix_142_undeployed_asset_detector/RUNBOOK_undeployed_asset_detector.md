# RUNBOOK — bugs_open/142

Every command here was needed and had a gotcha. Gotchas are attached.

---

## R1 — Run the SHIPPING check as SQL, to size what it would raise today

The check is Go, so the only honest way to size it is to run its own predicate.
Note `NOT (purpose = ANY(...))` is the **new** exclusion — drop that line to
reproduce the pre-fix behaviour.

```sql
WITH q AS (
  SELECT a.site_id, a.id, COALESCE(a.purpose,'unknown') AS purpose
  FROM assets a
  WHERE NOT EXISTS (
      SELECT 1 FROM page_components pc JOIN pages p ON pc.page_id=p.id
      WHERE p.site_id=a.site_id AND pc.build_status='deployed'
        AND (pc.rendered_html LIKE '%/assets/images/'||COALESCE(a.purpose,'')||'.%'
          OR pc.rendered_html LIKE '%/assets/images/'||COALESCE(a.purpose,'')||'-%')))
SELECT s.domain, count(*), string_agg(DISTINCT q.purpose, ',')
FROM q JOIN sites s ON s.id=q.site_id GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **Do not "fix" the `_` in the concatenated purpose.** It is a LIKE wildcard and
it is load-bearing — see R4.

## R2 — Prove a brand-head artefact's state on the wire, not in the DB

The DB cannot tell you whether the file serves. Cache-bust, or Cloudflare
answers for a version you did not deploy.

```bash
curl -s -o /dev/null -w '%{http_code}' "https://$d/assets/images/og-card.png?cb=$RANDOM"
```

⚠ Both filenames are irregular: the purpose is `og_card` and the file is
`og-card.png`. `favicon` is the only one where purpose and filename agree.

## R3 — Where the brand-head reference actually lives

```sql
SELECT s.domain,
  count(*) FILTER (WHERE sc.rendered_html LIKE '%/assets/images/favicon.png%') AS fav,
  count(*) FILTER (WHERE sc.rendered_html LIKE '%/assets/images/og-card.png%') AS og
FROM site_components sc JOIN sites s ON s.id=sc.site_id
WHERE sc.slot_name='head' GROUP BY 1 ORDER BY 1;
```

⚠ **Never add `AND sc.build_status='deployed'` here.** `site_components` rows are
`'rendered'` and never `'deployed'` — all 42 of them. The predicate returns zero
rows and looks like a clean pass.

⚠ The head reference is **not** evidence the artefact exists.
`injectBrandHeadTags` emits the `<link>`/`<meta>` unconditionally, so 13 of 14
heads advertise a card whether or not one was ever generated (idea.uk advertises
a 404).

## R4 — Measure the LIKE underscore before touching it

Run both forms side by side. This is the query that stops the "obvious" fix:

```sql
WITH a AS (SELECT id, site_id, COALESCE(purpose,'') AS p FROM assets WHERE purpose LIKE '%\_%')
SELECT a.p, count(*) AS assets,
  count(*) FILTER (WHERE EXISTS (SELECT 1 FROM page_components pc JOIN pages pg ON pc.page_id=pg.id
     WHERE pg.site_id=a.site_id AND pc.build_status='deployed'
     AND pc.rendered_html LIKE '%/assets/images/'||a.p||'.%')) AS deployed_UNESCAPED,
  count(*) FILTER (WHERE EXISTS (SELECT 1 FROM page_components pc JOIN pages pg ON pc.page_id=pg.id
     WHERE pg.site_id=a.site_id AND pc.build_status='deployed'
     AND pc.rendered_html LIKE '%/assets/images/'||replace(a.p,'_','\_')||'.%')) AS deployed_ESCAPED
FROM a GROUP BY 1;
```

Result 2026-07-31: `content_hero` 38 / 38 / **0**. Escaping the wildcard
manufactures 38 false findings.

## R5 — The handler's real precondition

`derive_brand_head_assets` refuses with `{"derived": false, "reason": "no active
logo asset"}` unless this returns a row:

```sql
SELECT 1 FROM assets WHERE site_id=$1 AND asset_key='logo' AND status='active';
```

⚠ **`asset_key`, not `purpose`.** They disagree badly: 15 of 15 sites have a logo
by `asset_key`, only 4 by `purpose`. Reading the wrong one makes 11 sites look
unserviceable.

## R6 — Find the provenance exception (a row that records the wrong url)

```sql
SELECT s.domain, a.purpose, count(*) AS active_rows,
       count(*) FILTER (WHERE a.url = bh.p) AS rows_at_published_path,
       string_agg(DISTINCT a.url, ' | ')
FROM assets a JOIN sites s ON s.id=a.site_id
JOIN (VALUES ('favicon','/assets/images/favicon.png'),
             ('og_card','/assets/images/og-card.png')) AS bh(pu,p) ON bh.pu=a.purpose
WHERE a.status='active' GROUP BY 1,2
HAVING count(*) FILTER (WHERE a.url = bh.p) = 0;
```

Returns gamesdesign.co.uk and robot-hands.com, whose only rows carry
`/assets/images/input-data.asset-key.jpg` — an unresolved template literal — while
both sites serve the real files 200.

## R7 — Build and test when the working tree does not compile

Other sessions leave in-flight edits in shared packages, so
`go build ./platform/...` in the working tree tells you nothing about your change.
Build HEAD plus your files only:

```bash
SB=<scratchpad>; rm -rf $SB/headtree && mkdir -p $SB/headtree
git archive HEAD | tar -x -C $SB/headtree
cp <each file you changed> $SB/headtree/<same path>
cd $SB/headtree && go build ./platform/... && go test ./platform/orchestration/actions/discovery_checks/
```

## R8 — PREPARE new SQL against the live schema before trusting it

`go build` cannot parse a SQL string. Every query in the change was proven with
`PREPARE name(uuid, text[]) AS <query>;` against `clients_db` — it type-checks
parameters and column references without executing anything.

## R9 — Verify after the roll

The check is only reachable via `design-discovery-agent`, which is driven by
`improvement-sweep` — **disabled since 2026-05-02** (`bugs_open/083`). So a live
firing is not available on demand. Pod-grep the discriminating strings instead,
with a positive control in the same exec:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'brand_head_provenance_url_unexpected'"   # 0 before, >=1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'has never been generated'"               # 0 before, >=1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'Asset .%s. generated but not deployed'"  # POSITIVE CONTROL, must stay >=1
```

⚠ Run it on **every replica**, not `deploy/agent-chassis` — that reads one pod of N.
