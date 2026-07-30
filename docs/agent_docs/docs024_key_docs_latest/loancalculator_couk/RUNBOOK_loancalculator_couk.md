# RUNBOOK — loancalculator.co.uk

Commands that were hard to get right, with the gotcha attached. Update HERE, not in
scrollback.

## 1. Is the site serving? (and the outage signature)

```bash
curl -sI --max-time 15 https://loancalculator.co.uk/worker-health | head -3
curl -sI --max-time 15 https://loancalculator.co.uk/ | head -5
```

`/worker-health` returning `200 "Worker is running!"` proves the **edge worker** is
bound to the zone (`scripts/cloudflare/worker.js` maps `{hostname}{path}` →
`b2://portfolio-sites/...`).

**GOTCHA — the outage signature is a HANG, not an error.** On 2026-07-30 before the
fix: DNS resolved to Cloudflare proxy IPs, TLS handshook, HTTP/2 request was sent —
and nothing came back until timeout (`curl` exit 28). No status code, no error page.
A dead origin behind a live proxy looks like a network problem, not a config
problem. **Compare against a healthy zone in the same breath** —
`curl -sI https://gamesdesign.co.uk/worker-health` — or you cannot tell "this zone
is misconfigured" from "the internet is slow".

**The fix (owner, Cloudflare dashboard — not scriptable without a zone-scoped
token):** zone `loancalculator.co.uk` → Workers Routes → add
`loancalculator.co.uk/*` bound to the same worker gamesdesign.co.uk uses. Done
2026-07-30 ~15:10Z; verified 200 at 15:11Z.

Optional follow-ups: `www` is NXDOMAIN (add proxied CNAME → apex + a `www.…/*`
route); confirm the sites-repo Action secret `CF_API_TOKEN` covers this zone — the
failure symptom is only a stale cache plus a null `ZONE_ID` in the Action log.

## 2. Deploying static file changes

The deploy repo is the **shared** monorepo, one dir per domain:

```bash
cd ~/projects/sites && git pull --ff-only        # ALWAYS first — other lanes push here
git status --short loancalculator.co.uk
git add <paths> && git commit <paths> -m "..."   # explicit pathspec, per CLAUDE.md
git push
```

**GOTCHA — this repo is shared by every site and pushed to by automation.** The pull
on 2026-07-30 brought in dartsonline/fundamentallyai/gamesdesign changes from other
lanes. A bare `git commit -m` here would sweep them. Pathspec on `commit` is what
protects you, not the one on `add`.

**GOTCHA — deletions matter.** The workflow runs `b2 sync --delete`, so a file
removed from the repo disappears from the bucket. That is how dead files
(`style.css.1`, `pdf-gen.js`, `search.js`) actually get cleaned up — but it also
means an accidental deletion is a live outage for that path.

Push to `master` triggers `.github/workflows/deploy-to-b2.yml` in the sites repo: it
diffs changed top-level domain dirs, `b2 sync --delete --skip-newer <domain>
b2://portfolio-sites/<domain>`, then purges that domain's Cloudflare zone cache.

## 3. The passthrough component row — reuse it, never seed a second

```sql
SELECT id, name, function, component_level, is_active
FROM content_components WHERE function='ported-page';
```

As of 2026-07-30 there is **exactly one**:
`a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef` — "Ported Page (webdesign.co.uk)",
`component_level='section'`, active.

**GOTCHA — `import.go`-style lookups take `ORDER BY created_at DESC LIMIT 1`.**
Seeding a second `ported-page` row silently repoints every future port at the new
one. Reuse this id. (Its name says webdesign.co.uk; that is historical — renaming it
to something generic is safe and preferable, but adding a sibling is not.)

## 4. Finding the adoption run

```bash
./scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh \
  loancalculator.co.uk --from https://loancalculator.co.uk \
  --fidelity locked --email uk@websy.uk
```

**GOTCHA — budget ~30 minutes, not ~2**, and find the run **by payload, not by the
printed id** (dispatch queues behind the fleet; a missing row is latency, not a
dropped dispatch — do not retry on that evidence):

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'domain' = 'loancalculator.co.uk'
ORDER BY created_at DESC LIMIT 5;
```

## 5. Checking what adoption actually created

```sql
SELECT url, name, page_type, status, build_status, rebuild_policy
FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain='loancalculator.co.uk')
ORDER BY url;

SELECT pc.slot_name, pc.build_status,
       pc.content_data->>'deploy_mode' AS deploy_mode,
       length(pc.rendered_html) AS html_len
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='loancalculator.co.uk');

-- must be EMPTY under fidelity=locked:
SELECT item_type, count(*) FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='loancalculator.co.uk')
  AND item_type IN ('needs_content_page','needs_tool_recreation')
GROUP BY item_type;
```

## 6. The crawl fidelity gate (run BEFORE letting the deploy drain)

Compare what the platform captured against what the site actually serves:

```bash
# per page: sha256 of the stored rendered_html vs the repo file
```
A mismatch means firecrawl's `rawHtml` is not byte-verbatim — stop and report;
do not "fix" it by hand-editing the stored HTML, because the next crawl repeats it.

## 7. Verifying a deploy

**GOTCHA — never verify against git or the image tag; verify at the artefact.**
For pages: `curl -s https://loancalculator.co.uk/tools/standard-calc.html | grep -c 'id="monthly-display"'`
plus a positive control in the same command. For the chassis change (M1/M2):
pod-grep a string the change **added** and one it did not, per CLAUDE.md — a roll is
not evidence the fix shipped.

## 8. Hand-raising a page_rerender item (if one is ever needed)

**GOTCHA (LANDMINES.md, dartsonline lane):** a `page_rerender` item needs `page_id`
in the **spec AND the column**; `needs_page` resolves by `page_name` but this does
not. Copy the spec shape from a **completed row of the same item_type**:

```sql
SELECT jsonb_pretty(spec), page_id FROM site_work_items
WHERE item_type='page_rerender' AND status='complete' LIMIT 1;
```
Symptom if you get it wrong: `rerender_single_page: page_id not found in input`,
retried to `attempt_count = 3` — reads like a flaky handler, is a malformed item.

## 9. This lane's council submission

| what | value |
|---|---|
| commit | `e6a8bb63b` (M1+M2, register, landmines) |
| submission corr | `f9eae63e-05fb-40c8-b60c-1670c5681cbe` |
| trailer on the commit | `Council-Submitted: pending` — **a placeholder, and a mistake**; see NOTES. The join in `098` will miss this commit. Submit BEFORE committing so the real correlation can go in the trailer. |

Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='f9eae63e-05fb-40c8-b60c-1670c5681cbe' AND kind='council_report'
ORDER BY created_at;
```
**GOTCHA — do NOT roll the chassis while a council run is in flight**; a roll kills
it and the review is lost (the run has to be resubmitted, spending credits again).
