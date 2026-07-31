# RUNBOOK — loanandmortgagecalculator.co.uk

Commands that were hard to get right, with the gotcha attached. Update HERE, not in
scrollback.

## 1. OWNER ACTION — the serving path (blocking, not scriptable here)

The domain is registered but **parked at its registrar**. Signature:

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

## 3. The two build assertions, and proving they can fail

`build_site.py` refuses to write if either fails:

1. **every inline `<script>` block is byte-identical to the source** — the
   calculators' logic must not change;
2. **every external script the source referenced is still referenced** — a
   byte-identical script with a missing dependency is still a broken page.

Mutation-test them rather than trusting them (this is `features_open/027`'s S2 rule
— at least one mutant red per assertion class):

```bash
SP=/tmp/…/scratchpad
sed 's|    # ── the safety property|    out = out.replace("parseFloat", "parseFloatX")  # MUTANT\n    # ── the safety property|' \
    build_site.py > $SP/build_mutant.py
python3 $SP/build_mutant.py --check ; echo "exit=$?"   # MUST be non-zero
```
Measured 2026-07-31: `exit=1`, `ABORT mortgages/simple: inline script blocks changed`.

## 4. Verifying the calculators in a real browser

The site uses root-relative paths, so it needs a server rooted at the site dir —
`file://` will not do.

```bash
cd ~/projects/sites/loanandmortgagecalculator.co.uk
python3 -m http.server 8765 --bind 127.0.0.1 &
curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8765/assets/css/style.css   # 200 = paths resolve

cd docs/agent_docs/docs024_key_docs_latest/webdesign_tools_repair
python3 toolaudit.py --json /tmp/…/audit.json \
  $(for p in simple repayment affordability stamp-duty overpayment investor portfolio \
             equity-release bridging-loan fee-analyser rate-forecaster fact-finder; \
     do echo http://127.0.0.1:8765/mortgages/$p.html; done) \
  $(for p in standard-calc compare-loans consolidation car-finance-calculator \
             credit-health-check damage-checker interest-rate-stress-test loan-vs-savings \
             overpayment-calculator settlement-calculator application-tracker; \
     do echo http://127.0.0.1:8765/loans/$p.html; done)
```
Last result (2026-07-31): **RESPONDS=23**, i.e. all of them.

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
