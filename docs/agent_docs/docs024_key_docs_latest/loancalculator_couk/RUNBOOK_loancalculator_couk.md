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

## 10. The adoption crawl config (changed 2026-07-30)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'crawl_site'->'config')
FROM agent_definitions WHERE type='site-adoption-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

`formats` must contain **`rawHtml`** or verbatim adoption has nothing to preserve —
it is there (`["markdown","rawHtml"]`), and `only_main_content: false`, both correct.

**`scrape_config.limit` raised 30 → 60** (this lane, 2026-07-30). The site has 29
HTML files and the old cap was 30: `/` and `/index.html` can each be crawled as a
distinct URL, so a 29-file site can legitimately present 30–31 URLs and lose pages
to the cap. **A page dropped by the cap is silently absent, not an error** — it just
never appears in the crawl index, so it is never adopted.

Snapshot taken first, and verified to hold the **pre-change** value:
```sql
SELECT snapshot_agent('site-adoption-agent', '<reason>');   -- two-arg form!
SELECT snapshot_taken_at,
       (default_config #>> '{workflow,steps,crawl_site,config,scrape_config,limit}')
FROM agent_definitions_backup WHERE type='site-adoption-agent'
ORDER BY snapshot_taken_at DESC LIMIT 1;                    -- must show 30, not 60
```
**GOTCHA (LANDMINES.md):** `snapshot_agent(text,text)` writes to
`agent_definitions_backup`; the one-arg form writes an `is_snapshot=true` row into
`agent_definitions`. Check the wrong table and a good snapshot looks like a no-op.
**GOTCHA (LANDMINES.md):** `jsonb_set(..., create_if_missing := false)` on a wrong
path is a **silent no-op**, not an error — assert `UPDATE 1` and re-read the value.

## The acceptance gate: does every calculator still compute? (2026-07-31)

The bar is *"starts similarly enough with working tools"*. This is how it is measured
— in a real browser, not by inspection.

```bash
SP=<scratchpad>
# 1. PIN THE HARNESS — both files. It lives in another lane and is edited daily.
cp docs/agent_docs/docs024_key_docs_latest/webdesign_tools_repair/toolaudit.py  $SP/toolaudit_pinned.py
cp docs/agent_docs/docs024_key_docs_latest/webdesign_tools_repair/toolprobe.py  $SP/toolprobe.py
sha256sum $SP/toolaudit_pinned.py   # baseline of 2026-07-31 used e7607680…

# 2. Run it from the pinned directory (11 interactive tools + credit-roadmap)
cd $SP && python3 toolaudit_pinned.py --json $SP/after.json \
  $(for t in application-tracker car-finance-calculator compare-loans consolidation \
             credit-health-check credit-roadmap damage-checker interest-rate-stress-test \
             loan-vs-savings overpayment-calculator settlement-calculator standard-calc; do
      echo https://loancalculator.co.uk/tools/$t.html; done)
```

**PASS = `RESPONDS=11  NO-CONTROL=1`.** Compare against
`acceptance/BASELINE_2026-07-31_calculators.json` per URL, not just on the totals —
a swap (one tool dies, another revives) preserves the count.

- **GOTCHA — pin BOTH files.** `toolaudit.py` does `from toolprobe import CDP,
  start_chrome`, so a lone copy dies on `ModuleNotFoundError: toolprobe`.
- **GOTCHA — the harness version is part of the result.** HEAD (`f38f5bf7f`,
  `1ea6740b…`) scores **`damage-checker` and `credit-health-check` DEAD when they are
  working**: a checkbox-only tool cannot be driven by assigning `.value` (a tick is a
  `click()`), and a wizard that reveals `<div id="step-N">` by moving a class is
  invisible to `innerHTML` diffing. The 2026-07-31 baseline used the *working-tree*
  fix; the delta is saved as `acceptance/harness_wip_vs_f38f5bf7f.diff`. **Re-pin the
  same harness before comparing, or the comparison measures the harness.**
- **NOT a collision risk:** the port and profile dir are randomised per run
  (`--remote-debugging-port=<rand>`), so a concurrent audit in another lane is fine.
- `--all` is useless here: `tool_urls()` builds from `DOMAIN =
  "https://webdesign.co.uk"`. Pass explicit URLs.
- `NO-CONTROL` on `credit-roadmap.html` is **correct, not a failure** — it is a static
  prose page under `/tools/` with no controls at all (see NOTES 2026-07-31).

## Comparing stored bytes against a file — the operator that matters

```sql
-- RIGHT: bytes, and an exact identity
SELECT octet_length(rendered_html), md5(rendered_html) FROM page_components …;
```
```bash
wc -c < file.html ; md5sum < file.html      -- compare against these
```

**GOTCHA:** `length(text)` counts **characters**, `octet_length(text)` counts
**bytes**. `standard-calc` is 5,730 characters and 5,734 bytes — four `£` signs, each
2 bytes in UTF-8. A gate written with `length()` reports a mismatch on a byte-exact
page, and can offset a real difference against a multi-byte character. Use `md5`/
`sha256`, or `octet_length`.

## Is there site-level chrome yet? (the generic-flip precondition)

```sql
SELECT slot_name, octet_length(rendered_html) FROM site_components
WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' ORDER BY slot_name;
```
**0 rows as of 2026-07-31.** While this returns nothing, `rebuild_policy='generic'`
would deploy every page with no head, no nav and no footer — `assemblePage` reads
chrome from this table (`rerender_single_page_action.go:660`). Decomposition must
populate it. Re-run this before any flip; a non-empty result is the precondition.

## The tool-equivalence gate: does a rewrite still COMPUTE the same? (2026-07-31)

`toolaudit.py` cannot answer this — `RESPONDS` is satisfiable by a page with an input
and no logic at all (see LANDMINES). Use `toolgolden.py` for any rewrite, re-port or
re-style of a calculator.

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
URLS="https://loancalculator.co.uk/index.html $(for t in application-tracker \
  car-finance-calculator compare-loans consolidation credit-health-check damage-checker \
  interest-rate-stress-test loan-vs-savings overpayment-calculator settlement-calculator \
  standard-calc; do echo https://loancalculator.co.uk/tools/$t.html; done)"

# BEFORE any change — already captured as acceptance/GOLDEN_2026-07-31_tool_values.json
python3 $LANE/toolgolden.py --out $LANE/acceptance/GOLDEN_<date>_tool_values.json $URLS

# AFTER — exit 0 = every tool computes identically; exit 1 = divergence, with values
python3 $LANE/toolgolden.py --compare $LANE/acceptance/GOLDEN_2026-07-31c_tool_values.json $URLS
```

**PASS = `all 12 tools reproduce their golden values exactly`, exit 0.**

> **⚠ USE `GOLDEN_2026-07-31c`, not `GOLDEN_2026-07-31`.** The earlier file is superseded
> and kept only as the record of what the earlier harness saw. It contains a
> **destructively wrong** baseline for `application-tracker`: the press selector clicked
> "Clear All Progress", wiping the state the driver had just set, so it recorded the
> post-wipe state as the tool's behaviour. It also predates `localStorage` being cleared
> between vectors, so that tool's vectors 2 and 3 started contaminated by vector 1.
> Everything else in the two files is identical field-for-field (1,653 compared, 21
> drifted, all of them that one tool). A `NUMBER`
divergence is an arithmetic regression; a `text/display` one is usually cosmetic but
read it before waving it through.

- **It refuses to write a golden file** when any tool is inert, input-independent, or
  fails to capture. That is deliberate: such a file certifies nothing and would mark a
  broken rewrite as correct. Fix the cause; do not record it.
- **GOTCHA — `dom_shape` (e.g. `"2:13804"`) is part of the record.** It is
  `script-count:serialised-DOM-length`. A capture taken mid-parse silently records
  £0.00 for everything while reporting success — `settle()` guards it by requiring both
  numbers to stop moving, and storing them makes a bad capture visible in every later
  diff. If `dom_shape` differs between golden and a comparison run, distrust the whole
  comparison before reading the values.
- **GOTCHA — `vary=0` is CORRECT for `application-tracker`, `credit-health-check` and
  `damage-checker`.** They have no numeric field to scale, so gate B is exempt by
  construction. Do not "fix" it by inventing vectors for them.
- **GOTCHA — modal dialogs BLOCK the renderer**, and CDP just times out with no
  indication why (`application-tracker`'s remove button calls `confirm()`). The harness
  stubs `confirm/alert/prompt` after settle. That is a stated behaviour change: a tool
  gated on `confirm()` proceeds as if accepted, identically for every implementation.
- **Run it from the repo root** — it resolves `toolprobe` relative to its own path.
- Verify the gate still has teeth if you change it: break one constant
  (`(APR/100)/12` → `/11` on `standard-calc`) against a local copy and confirm exit 1.
