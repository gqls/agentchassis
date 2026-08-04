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

> **CORRECTED 2026-08-02 — three rows now exist, and a non-empty result is NOT
> the precondition after all.** Something (not this lane) wrote head/header/footer
> at `2026-08-01 08:02:07` and queued a `page_rerender` for all 27 pages five
> seconds later. Every one completed, and the site did not change by one byte,
> because `loadVerbatimPageHTML` short-circuits before assembly. So the chrome
> arrived, was never exercised, and was **wrong in three ways at once**:
>
> | what it emitted | reality | consequence when assembled |
> |---|---|---|
> | `<link href="/assets/css/styles.css">` | the real file is `style.css` — plural is **404** | every page unstyled |
> | header `<ul>` with no `<li>` | 12 tools + 13 guides live in `nav.js` | no navigation at all |
> | `favicon.png`, `og-card.png` | both **404** | dead icon + a social card that 404s |
>
> **The lesson is the check, not the incident: "are there rows?" was the wrong
> question.** It is satisfiable by chrome that would take the site down, and it
> read as satisfied. Ask instead whether the chrome RESOLVES — fetch every asset
> it references and require 200, and count the nav links:
> ```bash
> for u in $(grep -oE '(href|src)="(/[^"]+)"' <<<"$HEAD$HEADER" | grep -oE '/[^"]+'); do
>   printf '%-34s %s\n' "$u" "$(curl -s -o /dev/null -w '%{http_code}' https://loancalculator.co.uk$u)"
> done
> ```
> This is the same shape as the fleet landmine about a roll not proving a deploy:
> the presence of the thing is not the working of the thing. See the next section
> for what replaced it.

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

## Proving a REWRITTEN tool before it ships (2026-07-31)

`toolgolden.py --compare` drives LIVE urls, so it can only judge a rewrite after
it is deployed. On this tree committing IS deploying, so that order means shipping
an unverified calculator to a public money page. `verify_rewrite.py` runs the same
comparison against the rewrite first.

```bash
cd /home/ant/projects/agentchassis            # it resolves toolprobe by its own path
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 $LANE/rewrite/verify_rewrite.py                       # every rewritten tool
python3 $LANE/rewrite/verify_rewrite.py settlement-calculator # one
python3 $LANE/rewrite/verify_rewrite.py --keep                # keep the staged site
```

**PASS = `all N rewritten tool(s) reproduce their golden values exactly`, exit 0.**

- **The site source is `~/projects/sites/loancalculator.co.uk`.**
  ⚠ `~/projects/sites2/loancalculator.co.uk` also exists and **is a different
  copy** — it differs from the served bytes on every page checked. Verify before
  trusting either: `curl -s https://loancalculator.co.uk/tools/X.html | md5sum`
  against `md5sum ~/projects/sites/.../tools/X.html`. Rewriting from the wrong
  copy produces components faithful to the wrong original, all passing review.
- **A cut pattern must match EXACTLY ONCE or the tool is refused.** A pattern
  matching zero times leaves the original widget on the page beside the
  replacement; the page then passes while proving nothing (two ids, two scripts,
  first wins, numbers identical). This fired for real on `standard-calc`.
- **Use `DIV:<regex>` for anything div-delimited.** It locates the opening tag and
  BALANCES `<div>`/`</div>`. A plain regex cannot count, and every page here uses
  `<div class="card">` for its ARTICLE sections too.
- **`allow_new_keys` is per-tool and must be justified in the spec.** It permits
  ADDED fingerprint keys and pairs renames BY VALUE; every pre-existing key is
  still compared strictly. Only for a tool whose controls had no ids and therefore
  no numeric coverage at all. A re-baseline is owed once it ships.
- **GOTCHA — `display:flex` blockifies its children**, and computed display is part
  of the fingerprint. Two tools here need opposite layouts because their original
  stylesheets differed. Do not harmonise them without re-baselining both.
- **GOTCHA — never interpolate copy into a `<script>`.** `render_tool.go` refuses a
  template that puts a quote-bearing schema value inside a script block, because a
  fallback containing `"58-day"` produced a syntax error that killed a whole
  calculator while it still passed every structural check. Copy goes in the markup;
  the script writes only the number.

## Emitting platform criteria, and the gate before installing one

```bash
python3 $LANE/toolgolden.py --emit-criteria $LANE/acceptance/criteria $URLS
```

Writes one `<slug>.criteria.json` per tool: `computed_values` checks the platform's
own Tier-4 runner executes. It REFUSES a tool whose controls cannot be named by a
selector rather than emitting steps that drive it differently from the capture.

> ### ⛔ RUN `INSTALL_GATE.sh` BEFORE PUTTING A FENCE IN `doc_plans`
> ```bash
> ./$LANE/acceptance/criteria/INSTALL_GATE.sh   # exit 0 = safe to install
> ```
> **A check type the running browser-runner does not know is SKIPPED, and an
> all-skipped fence PASSES.** Install one before the roll and it reports green
> having asserted nothing — the exact false-green `computed_values` exists to
> eliminate. Four council seats raised this; one gated the submission on it.
>
> - **`browser-runner-adapter`'s image has NO `strings` binary**, so CLAUDE.md's
>   `strings /app/X | grep -c` recipe fails silently and returns 0 for everything,
>   which reads as "your change did not ship". Use `grep -ac '<sym>' /app/<binary>`.
>   The gate carries a **positive control in the same exec** and refuses to
>   conclude anything when the control is also 0 — that is how this was found.
> - Carrying the string is necessary, not sufficient: run each fence ONCE
>   in-cluster and confirm no `not implemented` in its skips. Watch the **120s
>   whole-request deadline** (TL-036 hit it at 36 evaluations in-cluster while the
>   same fence took 10.6s locally); these fences carry three vectors each.

## Loading the components into `content_components` (2026-07-31)

```bash
cd /home/ant/projects/agentchassis
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 $LANE/rewrite/load_components.py --check    # validate only, no writes
python3 $LANE/rewrite/load_components.py --apply    # one transaction
```

**Safe by ORDERING, and that is the whole design.** A component with no
`page_components` row is inert — nothing renders it, nothing serves it, and the
site keeps serving its verbatim pages byte-for-byte. Attaching one to a page is a
separate, deliberate step. Verified after loading: all 11 stored templates are
md5-identical to the files on disk, and `/index.html`,
`/tools/standard-calc.html`, `/tools/consolidation.html` still serve unchanged.

- **It will not overwrite.** An existing `function` is skipped and named; there is
  no `--force`. A clobber here destroys another lane's component with no diff and
  no warning. Re-running is a no-op.
- **Validation runs over ALL of them before the transaction opens**, so a bad
  template cannot leave a half-loaded set behind.
- **GOTCHA, and it fired on the first run — prose mentioning a script tag breaks
  the tag-balance guard.** `tool-early-settlement`'s comment explained the
  quote-in-JavaScript bug and, in doing so, wrote `<` + `script` + `>` twice in
  prose. That is 3 opens against 1 close, and it fails not just this loader but
  `hasUnbalancedStructuralTags` — the platform's own birth-write gate. Say "a
  script block" in words inside a template comment; never the literal tag.
- **Every tool needs its tool-doc header before loading** (`add_tool_doc_headers.py`).
  Without it `check_tool_health` raises an `improve_tool` item per tool, pointing
  `tool-improver` at the most thoroughly verified tools on the fleet. The header is
  stripped at deploy assembly so it is output-neutral — proven, not assumed, by
  re-running `verify_rewrite.py` over all 11 afterwards.
- These are **novel** components (`forked_from IS NULL`), matching 27 of the 35
  tool components attached to pages fleet-wide.

## Chrome for an assembled site (2026-08-02)

Authored in `chrome/` and loaded by `decompose/load_chrome.py`. Three rows, and
each one exists to answer a specific way the generic chrome was wrong.

**`head.html`.** Links the REAL stylesheet, drops the two 404 icon links and the
404 `og:image` (a social card that 404s is worse than no card, so `twitter:card`
drops to `summary`), and keeps the platform's own injection points intact:
`<title></title>` and `content=""` must both survive verbatim, because assembly
rewrites them by literal string match —
`titleRe.ReplaceAllString` and `strings.Replace(head, 'content="">', …, 1)`.
Reorder the head so another tag holds the first `content=""` and the page's meta
description silently lands in the wrong tag.

It also carries the **assembled-layout block**, which is the only CSS the
decomposition adds:
```css
main { max-width: 960px; margin: 0 auto; padding: 20px; }
main .container { max-width: none; margin: 0; padding: 0; }
```
Every original page wrapped its body in one `.container` (960px, 20px padding)
and decomposition dissolves it, so assembly would otherwise emit a bare `<main>`
with no width — full-bleed text at a 19px base. Making `main` the container is
exact. **Giving each section its own `.container` is NOT equivalent**: N
containers stack N sets of vertical padding, so gaps grow with section count
where the original had zero. The second rule neutralises the `.container` that
survives inside a section — thirteen guide pages are one prose block and that
block is itself `<article class="container">`.

**`header.html`.** The nav LIFTED VERBATIM from `nav.js`, extracted
programmatically rather than retyped:
```python
m = re.search(r"const navHTML = `(.*?)`;", open("nav.js").read(), re.S)
assert "${" not in m.group(1)   # no interpolation, so it can be lifted as-is
```
Server-rendering it is what lets decomposition drop `<script src>` from the page.
Still a hand-maintained list of 25 links — generating it from `pages` is the next
step and is deliberately not bundled here.

**`footer.html`.** An ADDITION: the hand-built pages had none. It earns its place
because `/legal.html` is orphaned — zero inbound links from the 27 pages or from
`nav.js`. ~~`/tools/standard-calc.html` is the other orphan~~ **RETIRED 2026-08-03
on the owner's decision** (404 on the wire, sitemap 27→26); it duplicated the index
calculator while carrying two dated rate claims the homepage deliberately omits.

⚠ **The footer's own comment used to say all of this, and it shipped on all 27
pages** — an HTML comment is not stripped from a served page. The moment the
orphan was retired the comment became FALSE, site-wide. It now points here
instead. Keep engineering prose out of authored chrome: the chrome has no
tool-doc sentinel mechanism, so anything you write in it is published.

⚠ **Do not write a literal HTML tag inside a comment in any of these.** The first
head carried `<div class="container">` inside a CSS comment explaining the rule,
which is inert to a browser and trips the 5-pair structural balance predicate
that `load_components.py` and the birth-write guard both use. Same class as
template braces in a comment. Check before loading:
```bash
python3 - <<'PY'
for f in ("head.html","header.html","footer.html"):
    s=open(f).read().lower()
    for op,cl in (("<script","</script>"),("<style","</style>"),("<section","</section>"),
                  ("<div","</div>"),("<fieldset","</fieldset>")):
        assert s.count(op)==s.count(cl), (f,op,s.count(op),s.count(cl))
print("balanced")
PY
```

## Proving the decomposition BEFORE writing a row (2026-08-02)

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/decompose
./prepare_work.sh /tmp/decomp-work            # dump 27 pages, fetch assets, decompose
export DECOMP_WORK=/tmp/decomp-work
python3 verify_assembled.py --no-drive        # static only, seconds
python3 verify_assembled.py                   # + drives all 12 calculators, ~7 min
python3 verify_assembled.py --keep --pages index   # leave the staged site to look at
```

`prepare_work.sh` is scripted rather than pasted because the session scratchpad
does not survive the session and this had to be retyped once already.

**Compare against `octet_length`, never `length()`.** `length()` counts
CHARACTERS, so every `£` makes the dumped file read one byte long and 20 of the
27 pages look corrupt when they are byte-perfect.

Six assertions, and what each catches:

| | assertion | the failure it exists for |
|---|---|---|
| A | every calculator reproduces its golden across 3 vectors | the components were proven in the ORIGINAL page context; assembly changes the wrapper, the chrome and the document order, so that proof does not carry |
| B | no visible text lost | a block classified as neither prose nor tool |
| C | every script target still in the DOM | a calculator computing into an element that was dropped |
| D | no internal link goes nowhere | the header's hand-maintained link list drifting from `pages` |
| E | no section silently dropped | `sectionHasVisibleContent` discards any row with ≤10 visible chars, and returns success |
| F | no prose text ADDED | see below — B alone has a blind side |

**F is the one that earned its place.** The calculator component serves both
`/tools/standard-calc.html` and `/index.html`, and it carried standard-calc's risk
warning and two market-rate claims onto a homepage that had never shown them. B
passed (nothing lost). A passed (none of the three has an id). **A screenshot
caught it.** Check both directions, and screenshot the result — a fingerprint over
`[id]` elements is blind to every word that is not inside one.

**"Added" and "moved" are separated, and the distinction is checkable.** A string
the ORIGINAL held in a `<script>` literal and wrote in at runtime is the same
words relocated, not new copy — the rewrites moved a lot of copy into markup on
purpose. So a node absent from the original's text but present in its script text
is reported as MOVED. Without the split the report is 12 lines of noise and you
stop reading it; with it, 2 lines that both matter.

⚠ **Strip HTML comments before extracting text nodes.** `re.split(r"<[^>]*>")`
tears a comment in half at any `>` inside it, and these components quote markup
in their comments — F reported a paragraph of a component's own commentary as
page content.

⚠ **Compare with whitespace REMOVED, not collapsed.** The original marks up part
of a sentence (`…Personal Loans is <strong>7.9%</strong>.`) which collapses to
`… is 7.9% .`, while the component holds the same sentence as one text node. With
spaces intact that reads as text lost AND text added simultaneously, and it is
neither.

## Shipping the decomposition (2026-08-02)

Three writes, in increasing order of consequence. **Each is verified before the
next**, and the first two are inert by ordering, so only the third can change a
live page.

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/decompose
export DECOMP_WORK=/tmp/decomp-work

python3 load_chrome.py --check            # fetches every referenced asset, requires 200
python3 load_chrome.py --apply            # then confirm all 27 pages still byte-identical

python3 update_component.py --check tool-loan-repayment
python3 update_component.py --apply tool-loan-repayment

python3 load_decomposition.py --init                       # prose component + backup of ALL 27
python3 load_decomposition.py --check <page>
python3 load_decomposition.py --apply <page>               # THIS changes the page
python3 verify_shipped.py <page>                           # after the rerender lands
python3 load_decomposition.py --restore <page>             # if it does not
```

⚠ **`verify_shipped.py` can report a false mismatch for a page fetched inside its
deploy window.** The work item goes `complete` when the render and git commit
succeed; the bytes are still a minute or two behind through the sites-repo Action,
`b2 sync` and the Cloudflare purge. Seen once: 1 of 18 "did not match", the leading
hunk was chrome identical on all 27 pages, and a re-run gave 19 of 19 EXACT.
**When one page in a batch mismatches and the rest pass, re-run before believing
it** — a real fault is per-page-shape and reproducible, a propagation lag clears.
Do not put a sleep in the verifier: that hides the distinction instead of showing
it.

**Prove the chrome load changed nothing before going further.** It is only inert
while every page still ships verbatim, so this is a real check, not a formality:

```bash
while IFS='|' read -r name url; do
  live=$(curl -s "https://loancalculator.co.uk$url" | md5sum | cut -d' ' -f1)
  s=$(stat -c%s "$DECOMP_WORK/stored/$name.html"); head -c $((s-1)) "$DECOMP_WORK/stored/$name.html" > /tmp/s
  [ "$live" != "$(md5sum /tmp/s | cut -d' ' -f1)" ] && echo "DIFFERS $url"
done < "$DECOMP_WORK/pages.txt"   # expect: nothing
```

**`update_component.py` is not optional and not cosmetic.** The shipped page is
built from the file either way, but `content_components.html_template` is what the
NEXT re-render reads. Leave it stale and the homepage silently regains two dated
market-rate claims the owner was told it would not carry — invisible until it
ships.

### Getting the page actually rendered — the queue is ~325 items deep

Writing the rows does not deploy anything. Two routes, and on 2026-08-02 neither
was quick:

| route | how | why it did not work that day |
|---|---|---|
| `page_rerender` work item | insert with `page_id` in **the spec AND the column** | see below |
| direct orchestrate | `cta_link_integrity/scripts/049b_deploy_single_page.sh <page_id> <site_id> <domain>` | needs permission to `kubectl run` a kcat pod in the `kafka` namespace |

⚠ **A work item filed `status='detected'` is NEVER dispatched.** The dispatcher
selects `status IN ('triaged','approved')`. Nothing promotes `detected` — 31
`page_rerender` items filed by `discovery` have sat in it since 2026-07-14. File
as `triaged`, or promote immediately:
```sql
UPDATE site_work_items SET status='triaged' WHERE id='<id>' AND status='detected';
```

⚠ **`priority` is very nearly dead in the site selector.** `find_dispatchable_site`
orders `created_at ASC, priority ASC, id ASC`, so priority is consulted only to
break an exact `created_at` tie. A new item goes to the BACK regardless of
priority, behind every older one fleet-wide:
```sql
SELECT count(*) FROM site_work_items WHERE status IN ('triaged','approved')
  AND created_at < (SELECT created_at FROM site_work_items WHERE id='<id>');
```
On 2026-08-02 that returned **325**. `dispatch-queue-depth.sh` reported the lane
CLEAR at the same moment, and it was right: this is queue POSITION, not a stall,
and the two need different responses. Do not diagnose a stall on this evidence.

> ~~items being processed had been created ~19 hours earlier — so a same-day item
> is a next-day deploy~~ **WRONG, corrected same day: the item completed in about
> three hours.** The age of the items completing right now is the age of the
> queue's TAIL, not the wait a new arrival faces — they are the oldest rows by
> construction, and most of a 325-item backlog belongs to a few sites the
> dispatcher clears in one visit each. Use the count as a depth signal; do not
> convert an observed completion age into an ETA.

**Assemble-only is the branch you want.** Fire with NO `reason`: the workflow
takes `render_page`, which stitches the STORED `rendered_html` — no LLM, authored
content untouched. A `reason` of `section_data_resolved`/`image_landed` takes
`rerender_sections`, which re-renders every section from `content_data` through
the CURRENT template; that is safe here (every decomposed row has non-NULL
`content_data`, checked) but it is not the minimal path.

## Re-baselining the golden after decomposition (owed since the rewrite)

**Do this only when all 12 interactive pages are live and verified**, never
partially: `toolgolden.py` refuses a partial capture for a reason — the missing
pages would silently never be compared again.

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 toolgolden.py --out acceptance/GOLDEN_2026-08-02_decomposed.json \
  https://loancalculator.co.uk/index.html \
  https://loancalculator.co.uk/tools/{standard-calc,overpayment-calculator,consolidation,compare-loans,car-finance-calculator,loan-vs-savings,settlement-calculator,interest-rate-stress-test,credit-health-check,damage-checker,application-tracker}.html
```

**Why the new golden is strictly better than the one it replaces, and it is not
just freshness.** Three tools — consolidation, credit-health-check,
application-tracker — had **no numeric acceptance coverage at all**, because their
controls carried no `id` and the emitter could not name anything to drive. The
fingerprint keyed those inputs positionally (`input#0 … input#11`), which is
useless as an assertion: insert a field and every key after it means something
else. The rewrites gave them ids. So a golden captured now can drive every control
on every tool, which the 07-31 one structurally cannot.

That is the whole content of the "41 diverging keys" seen on 2026-08-02: 28
APPEARED (now named), 13 VANISHED (the same controls, formerly positional), and
**zero changed values**.

⚠ **THE GOLDEN SET IS NOW 11 TOOLS, NOT 12 (2026-08-03).**
`GOLDEN_2026-08-03b_after_orphan_retired.json` is the forward baseline;
`/tools/standard-calc.html` is retired and 404s, so any earlier golden — including
`GOLDEN_2026-08-03_defects_fixed.json`, captured hours before — now names a page
that does not exist and will fail. `tool-loan-repayment` keeps its coverage through
`/index.html`, which renders the same component, so nothing was lost. Self-verified:
11 of 11 reproduce.

⚠ **KEEP `GOLDEN_2026-07-31c` — do not delete or overwrite it.** It is the only
record of what the HAND-BUILT site computed, and every equivalence claim made
during the rewrite is stated against it. The new file is the forward baseline; the
old one is the evidence. Same rule as the summary series.

⚠ **Re-verify before capturing, and capture from a settled site.** A page fetched
inside its deploy window serves the previous bytes (see the warning above), and a
golden captured then would pin the OLD page's values as the new baseline — the one
error in this whole sequence that no later check could catch, because the golden
IS the reference.

## After a chassis roll: re-run the byte check (2026-08-02)

The decomposed pages are assembled by Go, so **a chassis roll can change what they
render to** — `assemblePage`, `getPageSections`, `StripToolDocHeader`,
`injectPageJSONLD` and `repairOutboundPageLinks` are all in the image. Nothing on
the wire moves until a page re-renders, so the risk is silent: the site is correct
until the next rebuild of any page, and then that one page differs.

```bash
export DECOMP_WORK=/tmp/decomp-work        # must still hold predicted/
python3 decompose/verify_shipped.py --all  # cheap, no browser
```

**This is a real check, not a ritual — but it has already passed one roll.** The
chassis rolled at `2026-08-02T18:39:14Z` in the middle of this rollout: **26 of the
27 pages rendered after it and 1 before, and all 27 matched the same predictions.**
So the assembly path is stable across at least one image change, which is evidence
rather than hope.

To decide whether a PENDING roll can move the bytes, diff the render path against
what the running pods were built from:

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{.items[0].status.startTime}'
git log --oneline --since=<that time> -- \
  platform/orchestration/actions/rerender_single_page_action.go \
  platform/orchestration/actions/rerender_link_repair.go \
  platform/content/tool_doc_header.go
```
Empty means the roll cannot move them. Note `save_page_sections_action.go` is NOT
on that list: it is a WRITER, not the renderer, so a change there cannot alter what
an already-stored page assembles to — it can only change what a section save
writes. (Checked 2026-08-02: the pending roll's only render-adjacent change was
exactly that, a per-slot shrink guard in the writer.)

If the predictions are gone, regenerate them — `prepare_work.sh` then
`load_decomposition.py --check --all` rewrites `predicted/` without touching the
database.

## Changing a LOCKED calculator (2026-08-03)

The 12 tool rows carry `lock_type='permanent'`. The lock is real: `save_page_sections`
guards its writes with `pageComponentAgentWritableSQL`, which is
`locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`, and
`rerender_page_sections` routes through `save_sections` as **the only writer of
`rendered_html`**. So a locked row's re-render is computed and thrown away.

Six steps. Steps 1–3 are inert while the row stays locked, which is what makes
step 4 the only consequential one.

```bash
cd /home/ant/projects/agentchassis
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk

# 1. PROVE IT OFFLINE, BOTH WAYS. Neither run alone is sufficient.
python3 $LANE/rewrite/verify_rewrite.py            # nothing moved that should not have
python3 $LANE/rewrite/defect_vectors.py --both     # the thing that SHOULD move, did

# 2. Validate what is about to be stored, then store it (still inert).
python3 $LANE/rewrite/load_components.py --check
python3 $LANE/decompose/update_component.py --check <function>
python3 $LANE/decompose/update_component.py --apply <function>

# 3. ⛔ BACKFILL ANY NEW SCHEMA FIELD, or it renders EMPTY. See below.
python3 $LANE/decompose/backfill_content_data.py --check <function>
python3 $LANE/decompose/backfill_content_data.py --apply <function>

# 4. Unlock + file the re-render in ONE transaction, to keep the window shut.
#    reason=section_data_resolved is REQUIRED — see the branch note below.
# 5. Verify against the SERVED page.
# 6. Re-lock, and re-baseline the golden.
```

### ⛔ COMMIT THE TEMPLATE **BEFORE** STEP 2, NOT AFTER

`update_component.py --apply` and `render_tool_row.py --apply` both write the
database **from the file on disk**. That the file is right is the only thing they
check — neither notices, or could notice, that it is uncommitted. So "it is live"
and "it is in the repo" are independent facts, and the apply path establishes only
the first.

Missed on 2026-08-03: `tool-loan-repayment` ran in production for most of a day
while HEAD still held the previous template —

```
working tree  756d8d0e2622  9595 b
live DB       756d8d0e2622  9595 b     <- production matched the WORKING TREE
HEAD          66a8e45df078  8905 b     <- and the repo did not
```

Anyone building or restoring from HEAD would have reinstated the very comment the
change removed. On a shared tree it was also one `git add -A` away from being swept
into another session's commit, half-described. Caught by an end-of-session sweep,
not by any step in this procedure.

**The sweep, worth running before you call a session done** — it compares every
component file against its live row and needs no arguments:

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk && python3 - <<'EOF'
import sys, hashlib
sys.path.insert(0,'rewrite')
from load_components import TOOLS, psql, validate
rows = dict(l.split('|') for l in psql(
  "SELECT function, md5(html_template) FROM content_components WHERE function LIKE 'tool-%';"
  ).splitlines() if l.strip())
for fn in sorted(TOOLS):
    got, err = validate(fn)
    if err: print("%-28s INVALID %s" % (fn, err)); continue
    f = hashlib.md5(got["template"].encode()).hexdigest()
    print("%-28s %s" % (fn, "match" if f == rows.get(fn) else "*** DIVERGED ***"))
EOF
```
2026-08-03: 11 of 11 match. A DIVERGED line means the repo and production disagree
— decide which is right before doing anything else.

### ⛔ A NEW SCHEMA FIELD RENDERS EMPTY UNTIL `content_data` CARRIES IT

`input_schema` carries a `fallback` per field and **the renderer never consults
it.** The render context is `base ⊕ page_components.content_data ⊕ resolved_data`;
`RenderTemplate` resolves an unknown key to the **empty string** and logs a `Warn`.

So schema + template is only two thirds of a change. On 2026-08-03 this would have
shipped the loan-vs-savings accessibility badge as an empty element — the whole
fix, absent, on a page that would still have passed acceptance, because the
element is present and no number moved. `backfill_content_data.py --check` names
exactly the fields at risk. Also in the fleet `LANDMINES.md`.

### ⛔ CORRECTION, SAME DAY — `rerender_sections` DOES NOT WORK ON THIS SITE

**Everything below about `spec.reason` is accurate about the platform and USELESS
here, and it cost this lane a full deploy cycle to find out.** Written before the
first re-render was fired; the re-render then completed, reported success, and
changed nothing.

`rerender_sections` resolves each section's component by passing
`page_components.slot_name` to `loadComponentSchemas`, which matches
`content_components` **by name or function**. This site's slots are POSITIONAL —
`prose-0`, `prose-1`, `tool-2` — so nothing resolves, every section takes the
`component not found, carrying stored HTML` branch, and the run is a no-op that
looks exactly like a success:

```sql
SELECT collected_data->'rerender_sections'->>'rerendered',
       collected_data->'rerender_sections'->>'carried'
FROM orchestration_states WHERE correlation_id='439489b6-73fa-4755-be37-2f3982a9cef9';
-- rerendered 0 | carried 4      <- with all four fixes already live in content_components
```

Measured across the fleet: **63 of 63 slots here**, 78 across 6 sites. Filed as
`bugs_open/182`. Until that is fixed, **updating `content_components.html_template`
on this site changes nothing that will ever reach a page.**

### ✅ THE ROUTE THAT WORKS HERE — render offline, write the row, assemble

This is not a workaround bolted on; it is the route all 27 pages were originally
shipped through, and it is proven byte-exact.

```bash
cd /home/ant/projects/agentchassis
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk

python3 $LANE/decompose/render_tool_row.py --check <function>   # runs a CONTROL first
python3 $LANE/decompose/render_tool_row.py --apply <function>
# then an ASSEMBLE-ONLY rerender (spec with NO `reason`) per page, and:
python3 $LANE/rewrite/defect_vectors.py --live                  # drives the SERVED pages
```

`--check` re-renders the row from a baseline ref and requires the **currently
stored bytes** back before it will write anything. If that control fails, the
offline renderer and the live row disagree and nothing should be written.

⚠ **The locks do NOT need lifting for this route.** `render_tool_row.py` writes by
SQL (the deliberate act the lock exists to force, not automation), and
assemble-only `render_page` does not write `page_components` at all. The 2026-08-03
unlock was unnecessary in hindsight — it was done for the `rerender_sections` route
that turned out not to work.

⚠ **`render_component`'s cache key must include `rewrite_dir`** (fixed 2026-08-03).
It did not, so rendering one component from two directories in one process returned
the first render twice and reported them identical — wrong exactly where a wrong
answer reads "nothing to do, your fix is already live".

### The `reason` decides the branch, and the wrong one is expensive

| `spec.reason` | branch | what it does |
|---|---|---|
| *(none)* | `render_page` | **assemble-only** — stitches the STORED `rendered_html`. This is what the decomposition rollout used. It will NOT pick up a template change. |
| `section_data_resolved` / `image_landed` | `rerender_sections` → `save_sections` → `render_page` | re-renders EVERY section from `content_data` through the CURRENT template. **This is the one a template fix needs.** |
| `cta_links_stale` | as above + CTA recompute | not this. |

**Before firing `rerender_sections` at a decomposed page, check the prose is
byte-stable** — it re-renders all 51 prose rows too, not just your tool:

```sql
SELECT count(*) AS prose_rows,
       count(*) FILTER (WHERE pc.rendered_html =
         '<section class="ported-prose" data-component="ported-prose">' ||
         (pc.content_data->>'content') || '</section>') AS reproduces_exactly
FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND cc.function='ported-prose';
-- 2026-08-03: 51 / 51. If these ever diverge, STOP — a rerender would rewrite prose.
```

`{{.content}}` is not escaped: `executeGoTemplate` imports `text/template`
(`call_agent.go:12`). Worth re-checking if that import ever changes — `html/template`
would turn all 51 prose sections into visible markup in a single pass.

### Unlock and file, in one transaction

```sql
BEGIN;
UPDATE page_components pc SET locked_at=NULL, locked_by=NULL, lock_type=NULL
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND cc.function='<the tool>';

INSERT INTO site_work_items
  (id, site_id, page_id, item_type, status, source, pipeline, severity, priority,
   handler_agent, created_by, summary, spec, created_at, updated_at)
VALUES
  (gen_random_uuid(), '0162cde4-633e-45e9-8ca6-87a6b2fe1d26', '<page_id>',
   'page_rerender', 'triaged', 'decompose-loancalculator', 'build', 'low', 80,
   'page-rerender', '<your tag>', '<why>',
   '{"domain":"loancalculator.co.uk","page_id":"<page_id>","filename":"tools/x.html",
     "page_name":"<page name>","reason":"section_data_resolved"}'::jsonb,
   now(), now());
COMMIT;
```

⚠ **`page_id` goes in the spec AND the column.** ⚠ **File `triaged`, never
`detected`** — the dispatcher selects `('triaged','approved')` and nothing
promotes `detected`.

### Re-lock as soon as it is verified

The window between unlock and re-lock is real exposure: the site is
`rebuild_policy='generic'` and open to the improvement loop for text.

```sql
UPDATE page_components pc
   SET locked_at=now(), locked_by='decompose_20260802_proven_calculators',
       lock_type='permanent'
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND cc.function='<the tool>';
```

Then **re-capture the golden**, or the next comparison reports your own fix as a
regression.

## ⛔ NAV: do NOT run `nav-updater` on this site (2026-08-03)

**It would rebuild the nav down to roughly one link, on all 27 pages, and ship it
immediately.** This is not a general warning copied from the fleet landmine — it is
measured here.

```sql
-- the state that makes it fatal, re-run before trusting this
SELECT count(*) FILTER (WHERE in_header) AS in_header,
       count(*) FILTER (WHERE in_footer) AS in_footer,
       count(*) AS pages
FROM pages WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26';
--  0 | 0 | 27      <- every page declares NEITHER flag (explicitly false, not NULL)

SELECT count(*) FROM site_nav_items WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26';
--  1                <- against 25 links in the authored header
```

`classifyPagesForNav` omits a page when it is never-primary (every `tool` page, and
every `/guides/` child-path page) **and declares neither flag**. That is all 27.
`populate_nav_tables` opens with `DELETE FROM site_nav_items WHERE site_id = $1`.

**The chrome is now LOCKED** (`lock_type='permanent'`,
`locked_by='loancalculator_authored_chrome_20260803'`, backup in
`site_components_bak_20260803_chromelock`), so a forced chrome re-render is refused
and emits `lock_blocked_change` instead of overwriting. That blocks the second half
of the damage, not the first — `site_nav_items` would still be rebuilt.

### If you need a nav change

1. **Edit `chrome/header.html` (or `footer.html`) and reload it.** The chrome is
   authored, not generated. `load_chrome.py --check` refuses on an asset that does
   not resolve — that refusal is the point (see § "Chrome for an assembled site").
2. Unlock, load, re-lock:
   ```sql
   UPDATE site_components SET locked_at=NULL, locked_by=NULL, lock_type=NULL
    WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND slot_name='header';
   -- ... load_chrome.py --apply ...
   UPDATE site_components SET locked_at=now(),
          locked_by='loancalculator_authored_chrome_20260803', lock_type='permanent'
    WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND slot_name='header';
   ```
3. Propagate with **assemble-only** `page_rerender` (no `spec.reason`), `page_id` in
   the spec AND the column.
4. `nav-link-fixer` (refreshes chrome from the EXISTING nav tables, no populate
   step) is the platform-native alternative — but it is only correct here once
   `site_nav_items` actually describes this site, which today it does not.

### Before generating the header from `pages` — the precondition

Set `in_header`/`in_footer` to match what the authored chrome already links.
Until that is done, every platform nav mechanism reads this site as having no
navigation at all. ⚠ `in_header` has a second consumer —
`buildServicesHTML` (`render_site_components_action.go:1156`) selects
`in_header = true OR in_footer = true` for a footer "Our Services" column — so the
flag change needs its own verification, not a bulk UPDATE.

⚠ **`in_header` has THREE states.** `render_site_components.go:1044` reads
`in_header = true OR in_header IS NULL` as included, while `classifyPagesForNav`
reads NULL as excluded. This site's pages are explicitly `false`, so they are
excluded by both — do not "tidy" them to NULL.
