# RUNBOOK — loanandmortgagecalculator.co.uk

Commands that were hard to get right, with the gotcha attached. Update HERE, not in
scrollback.

## 0. THE ONE COMMAND — verify everything, against the live origin

```bash
python3 docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/verify_site.py
python3 .../verify_site.py --disk      # fast offline pre-flight before a push
```

Checks every internal reference, every sitemap URL, every canonical (resolves **and**
self-names), every `ld+json` block parses, head essentials, no old-domain leakage, and
all 52 files byte-identical live. **It defaults to `--live` on purpose.**

**GOTCHA — never verify a link graph against `python3 -m http.server`.** It resolves
directory indexes; the production worker cannot. That single asymmetry shipped 3 dead
references on all 42 pages, 3 dead sitemap entries and 3 canonicals naming a 404. Worse,
the first link checker was "fixed" to resolve `/loans/` → `loans/index.html`, i.e. taught
the same forgiveness, which turned a true positive into silence. `--disk` mode therefore
models the worker's **single** `/` → `/index.html` rewrite and refuses every other
directory path, so both modes agree about what a valid reference is.

**GOTCHA — `robots.txt` legitimately differs live** (198 B → 2,034 B). Cloudflare's
zone-level *Managed robots.txt* prepends content-signal directives; the control domain
does the same and our own rules survive at the tail. It is `verify_site.py`'s single
sanctioned byte exemption. Do not "fix" it.

## 1. The serving path (DONE 2026-07-31 — kept for the diagnostic)

The zone and Workers Route are live. `worker-health` answers and all 52 files serve.
Compare against a healthy zone **in the same breath**, or a misconfigured zone is
indistinguishable from a slow network:

```bash
curl -sS https://loanandmortgagecalculator.co.uk/worker-health   # "Worker is running!"
curl -sS https://mortgagecalculator.co.uk/worker-health           # control
```

Two fleet-wide gaps remain, both confirmed identical on the control domain, both
owner-only (no Cloudflare credentials on this machine):

- **"Always Use HTTPS" is off.** `http://` returns **200**, not a redirect, so the same
  content is reachable under two schemes — at odds with the duplicate-content goal.
- **`www` does not resolve at all.** No site in the fleet has it, so this is consistent
  rather than broken, but it deserves a deliberate choice.

Historical signature, when it was still parked at the registrar:

```bash
dig +short loanandmortgagecalculator.co.uk A     # 76.223.54.146, 13.248.169.48 (AWS/registrar)
curl -sS https://loanandmortgagecalculator.co.uk/worker-health
# -> <script>window.location.href="/lander"</script>   i.e. a registrar lander, not our worker
```

Compare with a healthy zone **in the same breath**, or you cannot tell a
misconfigured zone from a slow network:

```bash
curl -sS https://mortgagecalculator.co.uk/worker-health   # -> "Worker is running!"
```

Two steps, both owner-only:

1. **Add the zone to Cloudflare** and change the nameservers at the registrar.
2. **Workers Routes** → `loanandmortgagecalculator.co.uk/*` → the same worker every
   other site uses (`scripts/cloudflare/worker.js`, which maps `{hostname}{path}` →
   `b2://portfolio-sites/{hostname}{path}`).

**GOTCHA — there are no Cloudflare credentials on this machine.** The only token is
the `CF_API_TOKEN` GitHub Actions secret in the sites repo, and it is used solely
for cache purge. `env | grep -i CF_` is empty, `~/.cloudflared` does not exist. So
no amount of scripting gets past this step.

**The B2 keys are already correct and do not need redoing after DNS.** The worker
keys off hostname, so `b2://portfolio-sites/loanandmortgagecalculator.co.uk/…`
was populated by the push on 2026-07-31 and simply starts being served the moment
the route exists. Verified: 52 of 52 files uploaded, run `30618165416`.

**Expect the purge step to have done nothing** on that run — `ZONE_ID` resolves to
`null` when the zone does not exist, and the step still prints
"Purging cache for …" and exits green. That is the documented signature, not a
failure.

## 2. Rebuilding the site

```bash
cd docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
python3 build_site.py            # the 23 ported calculators
python3 build_pages.py           # home, 2 section hubs, guides hub, 13 guides, legal, 404, robots, sitemap
python3 build_site.py --check    # assert only, write nothing
```

Both builders are idempotent and write into `~/projects/sites/loanandmortgagecalculator.co.uk`.

**GOTCHA — `build_site.py` rebuilds `<head>` from scratch, so anything the source
kept in its head is gone unless the builder puts it back.** This bit once already:
`bridging-loan`, `equity-release` and `fee-analyser` keep
`<script src="js/calculators.js">` in the `<head>` rather than the body, and the
first version dropped it. `bridging-loan` then threw `formatGBP is not defined`;
**the other two carried on looking fine**, which is the dangerous half. There is now
a second assertion for exactly this (see §3).

**GOTCHA — `credit-roadmap.html` is deliberately NOT ported** and the reason is a
comment in the LOAN table in `build_site.py`. It is 1,816 bytes with zero controls
and zero script. If a future run "restores" it, read that comment first.

## 3. The FOUR build assertions, and proving they can fail

> **CORRECTED 2026-07-31:** this section said "the two build assertions". There are now
> four. Properties 1-2 live in `port()` and cover the ported calculators only; **3-4 live
> in `write()`, the one function BOTH builders funnel through**, so they cover all 42
> pages. Each of the four exists because it has already caught a real defect on this site.

The build refuses to write if any fails:

1. **every inline `<script>` block is byte-identical to the source** — the
   calculators' logic must not change;
2. **every external script the source referenced is still referenced** — a
   byte-identical script with a missing dependency is still a broken page;
3. **no emitted `href`/`src` names a directory** (anything ending `/` other than the site
   root) — an object store cannot resolve a directory index, so such a reference is a
   live 404;
4. **every `ld+json` block parses** — all 13 guides once shipped structured data that no
   parser could read, because `html.escape()` escaped the quotes JSON needs.

Mutation-test them rather than trusting them (this is `features_open/027`'s S2 rule
— at least one mutant red per assertion class). Measured 2026-07-31, all four red:

| mutant | expected abort |
|---|---|
| `hub()` returns `/{section}/` | `reference "/mortgages/" names a directory` |
| JSON-LD headline re-escaped | `ld+json does not parse` |
| `parseFloat` → `parseFloatX` | `inline script blocks changed` |
| drop every external `<script src>` | `dependency /assets/js/calculators.js lost` |

```bash
SP=/tmp/…/scratchpad
sed 's|    # ── the safety property|    out = out.replace("parseFloat", "parseFloatX")  # MUTANT\n    # ── the safety property|' \
    build_site.py > $SP/build_mutant.py
python3 $SP/build_mutant.py --check ; echo "exit=$?"   # MUST be non-zero
```
Measured 2026-07-31: `exit=1`, `ABORT mortgages/simple: inline script blocks changed`.

## 4. Verifying the calculators in a real browser — against the LIVE domain

> **CORRECTED 2026-07-31:** this section used to run the audit against
> `python3 -m http.server` on port 8765. **Do not.** The site is live, and the local
> server resolves directory indexes that production cannot — the asymmetry that hid three
> dead URLs for a day. Audit the live origin. (The local server remains fine for a
> pre-DNS site, but then you must not conclude anything about the link graph from it.)

```bash
cd docs/agent_docs/docs024_key_docs_latest/webdesign_tools_repair
D=https://loanandmortgagecalculator.co.uk
python3 toolaudit.py \
  $(for p in simple repayment affordability stamp-duty overpayment investor portfolio \
             equity-release bridging-loan fee-analyser rate-forecaster fact-finder; \
     do echo $D/mortgages/$p.html; done) \
  $(for p in standard-calc compare-loans consolidation car-finance-calculator \
             credit-health-check damage-checker interest-rate-stress-test loan-vs-savings \
             overpayment-calculator settlement-calculator application-tracker; \
     do echo $D/loans/$p.html; done)
```
Last result (2026-07-31, live, after the hub-URL and JSON-LD fixes): **RESPONDS=23**.

**GOTCHA — a `HARNESS-ERROR` is not a site verdict.** One run returned
`HARNESS-ERROR … [Errno 32] Broken pipe` on `mortgages/overpayment.html`; re-run alone it
was `RESPONDS`. Across this site, **4 of 5** adverse verdicts have been the instrument.

**GOTCHA — when this harness says a tool is DEAD, first ask whether it did the
thing it claims to have done.** Three of the four non-passing verdicts on this
site's first run were harness faults, not site faults (recorded as faults ten,
eleven and twelve in the `webdesign_tools_repair` NOTES; all three now fixed).
The twenty-second check:

```bash
python3 evalpage.py <url> "(()=>{ /* drive the control yourself and return before/after */ })()"
```

## 5. Link graph and leakage checks (run before every push)

```bash
cd ~/projects/sites/loanandmortgagecalculator.co.uk
# dead internal refs + orphans — resolve '/' to index.html or you get 60 false hits
# (see NOTES; the first version of this check reported every href="/" as dead)
# true old-domain refs: the EXCLUSION must be case-insensitive, because
# "LoanAndMortgageCalculator.co.uk" CONTAINS "MortgageCalculator.co.uk"
grep -rnoiE "(^|[^a-z])(mortgagecalculator\.co\.uk|loancalculator\.co\.uk)" --include=* . \
  | grep -viE "loanandmortgagecalculator"        # blank = clean
grep -rl "cloudflareinsights\|nav-placeholder" --include=*.html .   # must be empty
```

## 6. Deploying (the shared sites repo)

```bash
cd ~/projects/sites && git pull --rebase     # --rebase, NOT plain pull. See below.
git add loanandmortgagecalculator.co.uk
git commit loanandmortgagecalculator.co.uk -m "..."     # explicit pathspec
git push
```

> **CORRECTED 2026-08-09 (bugfix 224 session): 120 is FIXED — this gotcha is
> retired.** The live workflow diffs `github.event.before..github.sha` and its
> own log line says "Tip is a merge commit — the range diff spans BOTH sides
> (bugs_closed/120)". `git pull --rebase` remains good hygiene, not load-bearing.
> The paragraph below is kept for history only.

**GOTCHA — `git pull --rebase`, never a merge (`bugs_open/120`, still unfixed).**
`.github/workflows/deploy-to-b2.yml` picks its domains with
`git diff --name-only HEAD~1 HEAD`. On a merge commit `HEAD~1` is the **first
parent — your own commit** — so the diff returns only the *other* side and your
domain is never named. The run is **green** and ships nothing. Confirmed still
present in the workflow on 2026-07-31. A rejected push is the normal case on this
repo (three other pushes landed within minutes of mine), so this fires often.

Assert the tip before pushing:
```bash
git log --oneline -1                          # must be YOUR commit
git cat-file -p HEAD | grep -c '^parent '     # must be 1, not 2
```

**GOTCHA — `b2 sync --delete`**: a file removed from the repo disappears from the
bucket. That is how dead files get cleaned up, and also how an accidental deletion
becomes a live outage.

Verify the deploy at the run, since the domain is not yet fetchable:
```bash
gh run list --limit 3
gh run view <id> --log | grep -E "Changed domains|upload " | head
# "Changed domains: loanandmortgagecalculator.co.uk" must appear, and the
# upload count must equal `find loanandmortgagecalculator.co.uk -type f | wc -l`
```

## 7. Adopting into the framework (after DNS)

```bash
./scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh \
  loanandmortgagecalculator.co.uk --from https://loanandmortgagecalculator.co.uk \
  --fidelity locked --email uk@websy.uk
```

Source/destination separation **is** wired, contrary to
`FUTURE_adoption_source_destination_separation.md` (which is stale):
`ensure_site_record.config.domain_override_field = "input_data.destination_domain"`,
consumed at `site_db_actions.go:131`; `crawl_site.config.url_field =
"input_data.target_url"`. So `<destination> --from <source>` genuinely differ.

**GOTCHA — DO NOT let the crawl be the byte source** (LANDMINES; the
loancalculator lane's G10). firecrawl's `rawHtml` is the **serialised
post-JavaScript DOM**, not what the origin sent. It mutated 3 of their guides in
production. This site is less exposed than theirs — its nav is static, so there is
no `nav.js` output to bake in — but absolutised URLs, `<meta charset>` →
`http-equiv`, collapsed whitespace and `&` → `&amp;` all still apply.

**So gate it before letting any `page_rerender` item drain:**
```sql
-- every component's stored bytes must match the repo file
SELECT p.url, length(pc.rendered_html) AS stored
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='loanandmortgagecalculator.co.uk')
ORDER BY p.url;
```
Compare each against `sha256sum` of the repo file. On a mismatch, load the **served
bytes** in (base64 → `convert_from(decode(...),'UTF8')`; note
`sha256(text::bytea)` is invalid — `decode(...,'base64')` already IS bytea) and
cancel the queued items. That is what the other lane did, and a subsequent
`page_rerender` then redeployed **byte-identically** (empty diff), which is the
rebuild-safety property working.

**The per-page split this site wants** (the owner asked for the two sites to
*evolve*, so a wholly-verbatim site is wrong):
- 23 calculators → `rebuild_policy='owned'`, verbatim, never regenerated;
- 13 guides → framework-managed, so content generation can move them.

```sql
-- reuse this component; NEVER seed a second (import.go-style lookups take
-- ORDER BY created_at DESC LIMIT 1, so a sibling silently repoints every port)
SELECT id, name, function FROM content_components WHERE function='ported-page';
-- a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef  "Ported Page (webdesign.co.uk)"
```

Find the run **by payload, not the printed id**, and budget ~30 minutes:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'destination_domain' = 'loanandmortgagecalculator.co.uk'
ORDER BY created_at DESC LIMIT 5;
```

Must be EMPTY under `fidelity=locked` — if either appears, the locked branch did
not take and an LLM is about to rewrite 23 working calculators:
```sql
SELECT item_type, count(*) FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='loanandmortgagecalculator.co.uk')
  AND item_type IN ('needs_content_page','needs_tool_recreation') GROUP BY item_type;
```

## 8. Confirming the chassis actually has the locked path

A roll is not evidence. Grep a string the change **added**, plus a positive control
in the same exec:

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $POD -- sh -c '
  strings /app/agent-chassis | grep -c "Verbatim adoption complete"      # added by e6a8bb63b
  strings /app/agent-chassis | grep -c "verbatim_adoption_deploy"        # added
  strings /app/agent-chassis | grep -c "apply_adoption_plan"'            # pre-existing control
```
2026-07-31 on `v1.0.1211`: `1 / 1 / 2` — live.

**GOTCHA — `grep -c fidelity` proves nothing.** There were already ~10 unrelated
`fidelity` hits in the tree before the mend (a vet-med parse guard, a gemini
comment), so the word appears in a binary that lacks the feature entirely.

## 9. The adoption byte gate — MANDATORY after adoption, and after every builder change

```bash
python3 gate_component_bytes.py            # report
python3 gate_component_bytes.py --repair   # load repo bytes into the mismatches
python3 gate_component_bytes.py            # re-run: the RE-RUN is the evidence
```

Measured 2026-07-31 on the real adoption: **41 of 41 components mismatched**, and the
repair wrote 41 with 0 lock-suppressed. What firecrawl's `rawHtml` had actually done,
diffed against the repo:

- every relative `href`/`src` **absolutised** to `https://loanandmortgagecalculator.co.uk/…`
- `<meta charset="UTF-8">` → `<meta http-equiv="Content-Type" …>`
- `&#9776;` decoded to a literal `☰`
- **the skip link became `https://…/loans/damage-checker.html#content`** instead of
  `#content` — the first thing a keyboard user hits, turned into a full page reload.
  That is an accessibility regression on all 41 pages, and it is the concrete reason
  this gate is not pedantry.
- `mortgages/repayment.html` came back **+11,432 bytes** — 8× every other page — because
  its script builds a 24-row amortisation table on load and the crawl serialised the
  result. Served has 2 `<tr>`; the post-JS DOM has 26.

**GOTCHA — a 3-page sample is not the site.** An earlier note here concluded "no script
injects DOM on load" from three pages with matching `<option>` counts. `repayment.html`
falsifies it. Sample by mechanism (one page that BUILDS markup, one that fills a select,
one static), not by convenience.

**GOTCHA — `content_data.sha256` is NOT the gate.** It is computed over the stored
`rawHTML`, so it is self-consistent by construction and says nothing about fidelity to
the origin. The gate digests in Postgres with `sha256(convert_to(col,'UTF8'))` —
`sha256(text::bytea)` is invalid SQL.

**GOTCHA — TWO WRITERS, and this is the one that will bite later.** A `page_rerender` on
an owned/verbatim page **commits to the shared sites repo** (`Rerender: <file>`, into
`github.com/gqls/sites`) and thence to B2. So `build_site.py` → repo → Actions and
`page_components` → rerender → Actions now write the same 41 files. They agree only
while something keeps them in step. **After ANY builder change, re-run
`gate_component_bytes.py --repair`**, or the DB still holds the old bytes and the next
rerender silently reverts your change.

## 10. Holding the rerender queue while you gate

`adopt_verbatim` creates one `page_rerender` per page **inside the adoption transaction,
already `status='triaged'`**, and `build-pipeline-trigger` is scheduled at
`interval_seconds=120`. So the window between the items existing and mutated bytes
deploying is under two minutes. **Hold first, then look.**

```bash
# flip them out of the dispatchable set the instant they appear (poll every 2s)
UPDATE site_work_items SET status='deferred', updated_at=NOW()
 WHERE site_id=(SELECT id FROM sites WHERE domain='loanandmortgagecalculator.co.uk')
   AND item_type='page_rerender' AND status IN ('triaged','approved');
```
Measured: the poller caught **41 in one second**. Release with `status='triaged'` after
the gate passes.

`deferred` is deliberate: it is **not** in `workItemTerminalStatuses`, so the row still
holds its `idx_swi_dedup` slot (nothing can create a duplicate behind it) and the release
is a plain `UPDATE`.

**GOTCHA — `sites.locked_at` does NOT hold dispatch, despite migration
`213_dispatch_gate_matches_dispatcher.sql:106` containing `AND s.locked_at IS NULL`.**
The **live** `build-pipeline-trigger` row has no such clause. Read the live
`agent_definitions` row, not the migration — a migration file is no better evidence than
a seed.

**GOTCHA — the queue will not drain on your schedule.** The dispatcher takes
`DISTINCT ON (site_id) … LIMIT 1` ordered by `site_id`, one site per invocation, and this
site ranked **14 of 14** with pending work. A released item sat unpicked for 10+ minutes.
To prove a single page, fire `page-rerender` directly instead of waiting:

```bash
kubectl -n kafka run -i --rm kcat-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -H correlation_id=$(uuidgen) -H orchestration_id=$ORCH \
  -H message_type=request -H client_id=demo_client -H action=orchestrate \
  -H sender_agent_type=cli -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"$SITE_ID","domain":"loanandmortgagecalculator.co.uk","page_id":"$PAGE_ID"}}
JSON
```
`kcat -P` exits 0 having sent nothing, so **verify by the orchestration row**, never the
exit code. Acceptance test result 2026-07-31: `COMPLETED`, deployed 1 file, live sha256
unchanged, and `git diff` against the platform's own commit was **empty** — the
rebuild-safety property working.

## 11. The divergence specs (D3/D6)

```bash
python3 set_divergence_specs.py            # dry run, shows the byte delta
python3 set_divergence_specs.py --apply
```

**GOTCHA — the `audience` aspect is read by NOTHING.** It is the most widely-populated
aspect in the database (29 of 33 sites) and no agent, prompt or Go path consumes it. An
earlier version of this workstream's plan named it as one of three targets. Same for
`editorial`, `voice`, `content_standards`, `terminology_and_positioning`.

The three seams that do reach a prompt: `identity.target_audience`,
`identity.key_differentiators`, and **`content_direction.formatted` — the only field of
`content_direction` the writer reads.** New keys therefore go *inside*
`content_direction`, never as new aspects.

**GOTCHA — regenerate `formatted` or the edit is invisible.** A hand-written
`content_direction` that omits it looks applied and changes nothing. The script ports
`datahelpers.FormatContentDirection` and **gates the port**: it regenerates the current
spec and aborts unless the result matches the stored `formatted` as a multiset of
**lines** — not as a string, because Go map iteration is random so section order is
arbitrary. Verified equal at 143 lines / 18,005 bytes, then +1,280 bytes after the write.

**GOTCHA — `site_specs`' JSONB column is `data`, not `spec_data`** (`spec_data` is a
workflow config key naming a path into `collected_data`). And `idx_site_specs_current` is
`UNIQUE (site_id, aspect) WHERE is_current`, so the supersede must precede the insert.

**No cross-site duplicate-content machinery exists in this platform.**
`check_content_duplication` is single-site, `remove_duplicate_page_sections` is
single-page, and `cross_site_contamination` detects another site's `company_name` in
rendered HTML — not topical overlap. The spec is the whole mechanism; nothing will warn
you if the two sites converge again.

## 12. The voice rebuild — decompose a page and re-voice it (2026-08-05/06)

**Read `PLAN_2026-08-05_voice_rebuild_and_decomposition.md` first**; this is the
command sequence only. `LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk`,
and every step wants `DECOMP_WORK` pointing at a scratch dir that holds
`manifest.json`.

⛔ **§2's `build_site.py` route is DEAD for any page that has been decomposed.**
The DB is the render source from that moment; rebuilding a decomposed page from
the build scripts and pushing it would be overwritten by the next rerender, and
the two would fight silently. §2 still applies to pages not yet converted.

```bash
# 0. chrome — ONCE per site, before the first page (it is inert until then)
python3 $LANE/load_chrome.py --check     # refuses on a 404 asset, a nav link
python3 $LANE/load_chrome.py --apply     # that misses pages, a moved splice literal

# 1. decompose from the PINNED sites-repo ref (never the working tree)
python3 $LANE/decompose_lmc.py $DECOMP_WORK/manifest.json --verbose

# 2. author overlays -> $LANE/voice_overlays/<page>.json  {"blocks":{"0":"<html>"}}
DECOMP_WORK=... python3 $LANE/show_blocks.py --list          # names + block shapes
DECOMP_WORK=... python3 $LANE/show_blocks.py <page>          # the prose to rewrite
DECOMP_WORK=... python3 $LANE/voice_apply.py --dry-run --pages <a,b,c>   # per author
DECOMP_WORK=... python3 $LANE/voice_apply.py                 # writes manifest_voiced.json

# 3. write the rows (THIS CHANGES A LIVE PAGE)
DECOMP_WORK=... python3 $LANE/load_lmc.py --check <page>     # prediction only
DECOMP_WORK=... python3 $LANE/load_lmc.py --apply <page>

# 4. deploy: assemble-only page_rerender, then diff against the prediction.
#    FULL insert shape (2026-08-15: the "§8" pointer that stood here was stale —
#    §8 is the locked-path section. Three more columns are REQUIRED: a bare
#    insert bounces on NOT NULL source, then created_by, and with both set the
#    item sits 'blocked' for ever with "No handler_agent set — item cannot be
#    routed to any agent". Nothing retries a blocked item.):
#    INSERT INTO site_work_items (item_type, status, page_id, site_id, spec,
#      summary, source, created_by, handler_agent)
#    VALUES ('page_rerender','triaged','<page_id>','<site_id>',
#      jsonb_build_object('page_id','<page_id>'),   -- NO spec.reason
#      '<why, one line>', 'lmc_decompose_voice',
#      'claude-session-<slug>-<date>', 'page-rerender');
diff $DECOMP_WORK/predicted/<page>.html <(curl -s -A Mozilla/5.0 https://…/<url>)

# 5. the calculator must still compute
PYTHONPATH=docs/agent_docs/docs024_key_docs_latest/webdesign_tools_repair:docs/agent_docs/docs024_key_docs_latest/loancalculator_couk \
  python3 $LANE/golden_compare_post.py $LANE/acceptance/GOLDEN_2026-08-05_prechange.json <url>

# rollback, per page, from the pre-change rows
DECOMP_WORK=... python3 $LANE/load_lmc.py --restore <page>
```

### The gotchas, each one measured

- **`toolgolden --compare` reports RED on every decomposed calculator and the
  page is fine.** The golden fingerprints every id-bearing element; the old
  wrapper carried `id="content"` and thus the whole page text, while the new
  one is the header's empty span. Use `golden_compare_post.py`, which asserts
  that field is exactly `|inline` and everything else matches. **Run its
  `--self-test` first** (one tool's golden vs another tool's live page) — it
  must FAIL, or the comparator is inert.
- **`mortgages/investor.html` cannot be certified by `toolgolden` at all** — it
  computes only ratios, so uniform x1/x2/x0.5 vectors leave every output
  unchanged and the inert-tool guard refuses. Use `investor_golden.py`
  (staggered vectors, one field at a time).
- **Use a tab field separator with psql on this site.** Every page title
  contains " | LoanAndMortgageCalculator.co.uk", so the default `|` splits
  inside the data and reads as truncation.
- **Fetch the served page ~90s after the item says `complete`.** Inside the
  deploy window B2 returns a 7-line `NoSuchKey` JSON at HTTP 200, and every
  grep against it returns 0, which reads as a clean pass. Guard on byte count
  and a leading `<!DOCTYPE`.
- **A page that is still verbatim ignores `site_components` entirely**, so
  chrome can be broken, rerender 41 pages, report success and change nothing.
  Check the chrome RESOLVES, never that rows exist.

## 13. Arithmetic validation — is the tool doing the RIGHT thing? (2026-08-08)

§4 and `golden_compare_post.py` prove a calculator still does what it did.
**They cannot prove it was ever right.** This section is the independent oracle:
expected answers recomputed from the annuity formula and the published HMRC
bands, in code that has never read these pages' JavaScript. Full account:
`REPORT_2026-08-08_arithmetic_validation.md`.

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
cd $LANE       # the modules import each other by name; run from HERE

# 1. all 18 class A/B tools, ~55 boundary vectors, against the LIVE site (~12 min)
python3 oracle.py --json /tmp/oracle.json

# 2. one tool while iterating
python3 oracle.py --tools stamp-duty
python3 oracle.py --tools standard-calc,compare-loans,consolidation

# 3. class C: monotonicity / bounds / determinism / round-trip / portfolio aggregate
python3 invariants.py

# 4. re-dump each tool's user-facing interface (labels, not script)
python3 inventory.py --out /tmp/inventory.json
```

### RUN THE CONTROLS IN THE SAME SESSION — a green oracle run alone is not evidence

```bash
python3 oracle.py --selftest-parse                      # no browser; ~1s
python3 oracle.py --mutate expectation --tools simple,repayment
python3 oracle.py --mutate crosstool  --tools simple,repayment,stamp-duty
python3 oracle.py --mutate parse      --tools simple
```

Each must print `CONTROL OK`. **The criterion is "no check may PASS", not "some
check must FAIL"** — `--mutate parse` legitimately produces N/A, because an
`<input type=number>` rejects `£200,000`, the field holds `''`, and `set()`
refuses to drive on rather than comparing against an empty field. That refusal
IS the property the control tests.

### The gotchas, each one measured

- **Do not author a tool spec by reading the page's `<script>`.** An oracle
  transcribed from the code it is checking agrees with the bug. Use
  `inventory.py`, which reports the visible `<label>` bound to each control and
  the caption above each result box — the site's own claim about what each
  number means, read the way a user reads it. Open the arithmetic only AFTER a
  check has failed; that is diagnosis, not authorship.
- **Tolerance is derived from the tool's own display precision, never chosen.**
  `£1,390` cannot be checked to the penny by anyone, so it is asserted at ±£0.50
  and every result line prints its resolution. A single global ±£0.01 would
  convict every whole-pound tool on this site.
- **A boundary vector must be one where a BROKEN implementation gives a
  different answer.** `consolidation` passed a 0%-APR-*debt* vector because its
  guard returns 0 and 0 is the right answer there; the defect only shows on a
  0% *new loan*, where 0 means a £0.00 monthly payment. Testing where a broken
  guard coincides with the truth yields a green tick and no information.
- **`--mutate crosstool` will leave a few PASSes and they are not failures of
  the control.** This suite deliberately packs vectors onto adjacent boundaries
  and £1,500,000 vs £1,500,001 of SDLT differ by 12p, inside any tolerance —
  a borrowed expectation equal to the true one cannot fail. Those print as
  `NON-TEST`. A separate case prints `MUTATION DID NOT BITE`: at £39,999 the
  borrowed expectation equals what the *buggy* page shows. Read the labels;
  do not loosen the bar to make the control green.
- **Exclude reset-ish buttons when walking any wizard.** `credit-health-check`'s
  result panel contains "Start Over" → `location.reload()`, so a naive
  "click the first button in the active step" walk answers every question,
  reaches the verdict and immediately destroys it — reporting a
  non-deterministic tool. `toolgolden.PRESS_JS` carries this exclusion and the
  comment explaining it; reuse the exclusion, not just the browser.
- **Wait for a tool's own save confirmation, never a fixed sleep.**
  `application-tracker` debounces its notes field by 1000 ms and reports
  `#save-status` = "✓ Saved to browser memory". Reloading immediately reports
  that notes do not persist — a harness fault wearing the tool's clothes.
- **Two routes to one vector is the check that catches a STALE answer.** A
  calculator gated `if (rate > 0)` writes nothing at 0% and the previous answer
  stays on screen looking fresh. Comparing against a single primed reading MISSES
  it, because the stale figure is whatever the last ACCEPTED vector produced —
  including an intermediate state created halfway through typing the new one.
  Drive the same final vector from two different priming vectors and compare;
  the readings must agree whatever the tool computes.

## 14. Unlocking a site for full framework editing (2026-08-10)

**The lock is `pages.rebuild_policy`** — `'generic'` | `'owned'` (migration 164's
CHECK allows only those). Unlock = `owned → generic`. Do it as a migration with a
`DO`/`RAISE` verify block, applied by hand, then `--record-only`.

```bash
# state, both axes at once — policy AND whether the page carries a calculator
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.domain, COALESCE(p.rebuild_policy,'generic') policy, tl.has_tool, count(*)
FROM sites s JOIN pages p ON p.site_id=s.id
JOIN (SELECT pc.page_id, bool_or(pc.rendered_html ~ 'onclick=|addEventListener') has_tool
      FROM page_components pc GROUP BY 1) tl ON tl.page_id=p.id
WHERE s.domain = '<domain>' GROUP BY 1,2,3 ORDER BY 1,2,3;"

# apply BY HAND (never --apply: it takes every pending file, including other threads')
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/367_*.sql
bash scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/367_*.sql --note "..."
```

### The gotchas, each one measured

- ⛔ **NEVER flip a page whose calculator lives in a single verbatim component.**
  `assemble_page → deploy_page(git_commit) → save_sections` commits LLM-written
  HTML to the sites repo **before** the DB guard refuses, so the calculator is
  replaced by prose and shipped. `rebuild_policy='owned'` is the only thing
  preventing it (TL-001 / migration 164). Check
  `slot_name='ported-page'` + a `onclick=|addEventListener` match first.
- **Decompose, lock the tool row, RE-SLOT it, THEN flip — in that order.**
  `matchLockedRow` matches `slot_name` against the incoming section name; a
  positional `tool-1` never matches and `save_page_sections_action.go:855` moves
  the unmatched locked row to `len(sections)+1` — the calculator lands at the
  BOTTOM of the page with only a `lock_blocked` item raised. Silent.
  > **CORRECTED 2026-08-10 (evening): the mechanism is real, the trigger condition
  > above is wrong, and the line numbers have drifted.** `matchLockedRow` is now
  > `save_page_sections_action.go:1043` and the reposition is `:928`; the lock-aware
  > DELETE is `:757`, not `:708`. It matches exact `slot_name` first, then
  > kebab-normalised — and `loans-consolidation`'s live `pages.sections` is already
  > `["prose-0","tool-1","prose-2"]`, so an incoming composition carrying `tool-1`
  > **does** match on the first branch. The trap fires when the composition **omits**
  > the tool slot, which is what a seeded site plan would do if it names sections
  > semantically. So the precondition is *"the composition must name the tool slot"*,
  > not *"positional never matches"*. `[INFERRED from the code + the live row;
  > UNMEASURED end to end.]` Caught by reading `matchLockedRow` rather than
  > repeating the claim. Full working: `HANDOFF_2026-08-10c` §6.
- **"Locked" never meant uneditable.** Migration 164: re-assembly of existing
  `page_components` is deliberately un-gated, *"it is how owned pages deploy"* —
  `page-rerender` and `section-editor` work on owned pages. `owned` blocks the
  GENERIC pipeline's wholesale rebuild. Do not report an unlock as "now editable".
- **Unlocking is not sufficiency: check `site_plans` too.** `rebuild_policy` says
  whether the pipeline MAY rebuild; `site_plans`/`site_plan_pages`/
  `site_plan_sections` are what it builds FROM. Both these sites had **zero** plan
  rows, so unlocking made 39 pages eligible and undriven. Read `bugs_open/204`
  before seeding a plan for a decomposed page.
  > **CORRECTED 2026-08-10 (evening): 204 is FIXED AND LIVE, and so is `bugs_open/189`**
  > — both since 2026-08-06, re-verified at chassis v1.0.1280 by pod-grep on both
  > replicas with a negative control (`stored_slot_name`, `load page slot
  > identities`, `slot_name repeats with different component_ids` = 1/1/1; nonsense
  > string = 0) plus the live config half (`slot_name_from` in
  > `page-content-writer`). A decomposed page **can** now be rebuilt from a plan, and
  > 189's *"never fire a build-path run on a page holding locked rows"* is lifted.
  > **They are still in `bugs_open/` per the owner direction of 2026-08-06 — a file's
  > directory is not its status; read its tail.** The `site_plans`-is-zero point
  > above still stands and is still the real gap.
- **Induce the verify block before trusting it.** A verify block made of `SELECT`s
  cannot stop the `COMMIT` — `ON_ERROR_STOP` ignores a non-empty result set. Use
  `DO`/`RAISE`, and run it once with a deliberately wrong expectation to watch it
  abort. 367's negative control (tool pages must still be `owned`) was induced by
  removing the tool filter, and it named the clobber.
- **Stamp what you changed.** 367 writes every changed page into
  `_mig367_unlocked_prose_pages`, so its ROLLBACK re-locks exactly those rows.
  A domain-wide re-lock would silently undo a concurrent thread's legitimate flip.

### ⛔ CORRECTED 2026-08-10 (late evening) — the state query above CANNOT see a calculator, and 367 unlocked six because of it

The `has_tool` expression in §14's state query and in migration 367 —
`bool_or(pc.rendered_html ~ 'onclick=|addEventListener')` — **misclassified six live
calculator pages as prose**, and they were unlocked for ~7 hours until migration 377
put them back. Do not reuse it. Details in `HANDOFF_2026-08-10c_continue_here.md` §2b.

**Use this instead — three independent spellings, OR'd, so no single blind spot
decides.** Expect **0 rows**; any row is a verbatim page carrying tool machinery that
the generic pipeline is currently allowed to rebuild.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -F$'\t' -A -c "
SELECT s.domain, p.url, p.build_status,
       (pc.rendered_html ~ 'onclick=|addEventListener|oninput=|onsubmit=|onchange=|onkeyup=|onblur=|onkeydown=|calculators\.js') AS handlers,
       (pc.rendered_html ~ '<input |<select |<textarea |<form ') AS form_controls,
       (pc.rendered_html ~ 'getElementById|querySelector') AS dom_addressing
FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
WHERE s.domain IN ('loanandmortgagecalculator.co.uk','loancash.co.uk')
  AND p.sections::text = '[\"ported-page\"]'
  AND COALESCE(p.rebuild_policy,'generic') = 'generic'
  AND (pc.rendered_html ~ 'onclick=|addEventListener|oninput=|onsubmit=|onchange=|onkeyup=|onblur=|onkeydown=|calculators\.js'
    OR pc.rendered_html ~ '<input |<select |<textarea |<form '
    OR pc.rendered_html ~ 'getElementById|querySelector')
ORDER BY s.domain, p.url;"
```

Why three: two of the six keep their `addEventListener` calls in the shared
`/assets/js/calculators.js`, so **the give-away string is not in the page at all**. A
calculator can be fully working with neither `onclick` nor `addEventListener` in
`rendered_html`. Over all 38 generic verbatim pages on the two sites the six match all
three axes and the other 32 match none — there is no borderline case, so a
disagreement between the axes is a signal to stop, not to average.

### Three further corrections to §14 above, each measured

- **A negative control must not be built from the same expression as its filter.**
  The bullet above ("Induce the verify block before trusting it") is right and
  insufficient. 367's control asserted "tool pages must still be `owned`" using its
  filter's own `onclick=|addEventListener`, was induced, fired on the induction — and
  was blind to exactly the six pages the filter was blind to. **Cross-check the
  resulting population against a source that never read your SQL.** Here that is
  `decompose_lmc.py`'s hand-authored `CALCULATOR_URLS` (2026-08-05), and migration 377
  asserts against it directly.
- **Key set assertions on `domain || '|' || url`, never `url` alone.** `pages.url` is
  not unique across these two sites — both have `/guides/jargon-buster.html` and
  `/legal.html`. A url-only assertion in 377's first draft passed a deliberately
  over-locking induction, reporting nothing missing and nothing unexpected while
  having stamped two rows it never matched. Caught by the induction, not by reading it.
- **"`deploy_page` commits before the DB guard refuses" is NOT what happened on the
  `page-build-handler` path.** §14 and both 08-10 handoffs state the loop commits
  LLM-written HTML to the sites repo one step before the refusal. On 2026-08-09,
  20 `needs_page` runs on this site reached step `save_sections` and died there — and
  `git log --since '2026-08-08 20:00' --until '2026-08-09 03:00' --
  loanandmortgagecalculator.co.uk/` in the sites repo shows **only** the 224 APR fix
  and one consolidation rerender. No clobbering commit. `[MEASURED for that path and
  that window only]` — the other two composition loops are unmeasured, and this
  changes nothing about the rule: the guard is what saved the pages, so keep a
  verbatim tool page `owned` until it is decomposed.

## D6 planner loop — the commands (added 2026-08-17)

`SITE='ed633ada-f8af-424b-b4d4-8af79160dbcd'`, sibling
`SIB='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'`,
`K="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

**The floor, four queries.** Re-run these in the SAME session as any plan you judge — this
site's page count moved by 4 within ten minutes of a handoff being written (NOTES 08-17).

```bash
# 1. active vs archived (a bare count(*) reads one too many — archived demo row)
$K -A -F$'\t' -c "SELECT status, count(*) FROM pages WHERE site_id='$SITE' GROUP BY 1;"
# 2. pages carrying a tool slot
$K -A -F$'\t' -c "SELECT count(*) FROM pages WHERE site_id='$SITE' AND status='active' AND sections::text LIKE '%tool-%';"
# 3. the tool slots BY PAGE — the list, never the count (23 B2 slots as of 08-17)
$K -A -F$'\t' -c "SELECT p.name, s.slot FROM pages p, jsonb_array_elements_text(CASE WHEN jsonb_typeof(p.sections)='array' THEN p.sections ELSE '[]'::jsonb END) AS s(slot) WHERE p.site_id='$SITE' AND p.status='active' AND s.slot LIKE 'tool-%' ORDER BY 1;"
# 4. page-type census + URL shape split (39 flat / 6 nested as of 08-17)
$K -A -F$'\t' -c "SELECT CASE WHEN url LIKE '%/index.html' THEN 'nested' ELSE 'flat' END, page_type, count(*) FROM pages WHERE site_id='$SITE' AND status='active' GROUP BY 1,2 ORDER BY 1,2;"
```

**The two structure-spec flags, on this site and the sibling.** ⚠ `evidence_base`,
`structure` etc. are `aspect` values in `site_specs` — there is **no** `evidence_base`
table, and asking for one gets a bare `relation does not exist` that reads like "the
feature is missing".

```bash
$K -A -F$'\t' -c "SELECT s.domain, ss.data->>'plan_includes_tools' AS tools, ss.data->>'honour_realised_identity' AS identity, ss.data->>'url_shape' AS shape, jsonb_array_length(ss.data->'pages') AS n_pages FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE ss.is_current AND ss.aspect='structure' AND ss.site_id IN ('$SITE','$SIB');"
```

**Read the READER, not the flag's documentation.** What `plan_includes_tools` actually
gates is in the live step config, and it keys on `component_level` — the sibling's
calculators are tool-level, ours are section-level, so the sibling's rationale does NOT
transfer (WRONG_CALLS 2026-08-17):

```bash
$K -A -t -c "SELECT default_config->'workflow'->'steps'->'load_components'->'config'->>'query' FROM agent_definitions WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
# and the whole workflow, so you know a replan writes `pages` (sync_pages is one of 14 steps):
$K -A -F$'\t' -c "SELECT s.key, s.value->>'action' FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s WHERE ad.type='build-site-planner' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL ORDER BY 1;"
```

**Is each live page a fixed point of the canonicaliser?** (38 of 45 are not, 08-17.) Call
the REAL helper — a re-derivation of the rule in SQL is the drift `bugs_open/215` was
written about. A scratch module keeps the shared tree clean; `identcheck/main.go` is in the
session scratchpad and reproduces from this recipe:

```bash
# go.mod:  module identcheck / go 1.24.0 / require github.com/gqls/agentchassis v0.0.0
#          replace github.com/gqls/agentchassis => /home/ant/projects/agentchassis
# main.go: datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
#            Role: page_type, Slug: name, ParentSection: parentSectionFromURL(url), FlatURLs: false})
#          — the descriptor write_site_plan_action.go:487 builds; Slug is firstNonEmpty(slug,name)
#            and a realised page carries no slug, so it is the stored NAME.
$K -A -F$'\t' -t -c "SELECT name, url, page_type FROM pages WHERE site_id='$SITE' AND status='active' ORDER BY page_type, name;" | grep -v '^$' > pages.tsv
cd identcheck && GOFLAGS=-mod=mod go run . < ../pages.tsv
```

⚠ **Build the controls INTO the harness, not into your reading of its output.** It asserts
(a) a canonical tool page comes back unchanged and (b) a legacy one moves, and exits 2 if
either fails — otherwise "nothing moved" and "the harness canonicalises nothing" print
identically.

**The seed** (`SEED_2026-08-17_identity_and_tools.sql`): apply with
`$K -f -` < file — it is a site seed, not a schema migration, so do **not** put it through
`run-migrations.sh` (an unscoped `--apply` takes every other thread's pending file too).

**The canary dispatch** — reuse the sibling's script rather than writing one:
`loancalculator_couk/canary_replan_407.sh` (swap SITE_ID/DOMAIN). It fires ONE
`build-site-planner` orchestration and prints its own judging queries. ⚠ `kcat -P` exits 0
having sent nothing, so prove dispatch by the orchestration row **by correlation**, never by
`now()`-interval; and no dispatch within ~300 s of a chassis pod restart
(`kubectl -n ai-persona-system get pods -l app=agent-chassis` → check `startTime`).
