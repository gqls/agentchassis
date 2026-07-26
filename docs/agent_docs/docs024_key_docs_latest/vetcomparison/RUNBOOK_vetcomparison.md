# RUNBOOK — vetcomparison.uk operations

Last updated 2026-07-16. Operator steps only; context in SUMMARY/PLAN (same dir).

## Deploying site changes (the sites repo trap)

Local `/home/ant/projects/sites` is ~1,700 commits behind, dirty with other sites' work; render
bots push master continuously. NEVER push from it, never `reset --hard` it, never force-push.

```
cd /home/ant/projects/sites && git fetch origin -q
git worktree add --detach <scratch>/sites-<task> origin/master
# edit + verify INSIDE the worktree (verify and push as SEPARATE steps —
# set -e does not reliably abort here)
cd <worktree> && git add <paths> && git commit
for i in 1 2 3; do git fetch origin -q && git rebase origin/master && \
  git push origin HEAD:master && break; done
# verify live with curl (python urllib gets Cloudflare 403), then:
git worktree remove --force <worktree>
```

## Enabling the directory exporter (after next chassis deploy)

Preconditions: chassis image containing `directory_export_json` deployed; then:

```sql
-- 1. bump the agent images to that build
UPDATE agent_definitions SET image_tag='<new tag>', updated_at=NOW()
WHERE type IN ('directory-json-exporter','directory-export-orchestrator');
-- 2. smoke once via kcat trigger (topic system.agent.business-intel.requests,
--    payload = scheduled_tasks.input_data for 'directory-export-json'),
--    check the sites repo commit: directory identical to live (2,389),
--    price-aggregates.json present, claimed/attributed empty ([]).
-- 3. enable
UPDATE scheduled_tasks SET enabled=true WHERE name='directory-export-json';
```

Fail-closed checks built in: refuses without explicit domain+vertical; attributed prices only
with per-row source URL + date on the practice's own host; opted-out businesses excluded;
`source='seed_import'` excluded unconditionally; aggregates only where n ≥ 3.

## Claim / opt-out handling (manual V1)

Requests arrive by email from the homepage routes (`#claim`), pre-formatted with the fields
below. Never action a request without verifying the requester represents the practice — an
unverified claim would publish someone else's prices under their name.

**Claim.** Verify (email domain matches the practice's recorded website, or call the number
published on their own site — not one supplied in the email), then log the request WITH the
consent wording they agreed to, mark the business claimed pointing at that request, and enter
their figures:

```sql
-- 1. log the verified request (consent_text is snapshotted deliberately)
INSERT INTO business_intel.claim_requests
  (business_id, request_type, requester_name, requester_email, requester_role,
   evidence_method, evidence_note, status, verified_by, verified_at,
   consent_text_version, consent_text, consent_given_at)
VALUES ('<business uuid>', 'claim', '<name>', '<email>', '<role>',
        'email_domain_match',            -- or callback_published_number
        '<what you actually checked>', 'verified', 'operator:<you>', NOW(),
        'claim-consent-v1-2026-07-16',
        'I confirm I am authorised to act for this practice and I agree that VetComparison.uk may publish the prices I provide, attributed to the practice and dated. I understand I can correct or withdraw them at any time by email.',
        NOW())
RETURNING id;  -- use below

-- 2. mark claimed, linked to that request (also reverses any prior opt-out)
UPDATE business_intel.businesses
SET is_claimed=true, claimed_at=NOW(), claimed_by='<claim_requests id>',
    publication_optout=false, optout_at=NULL, optout_note='Reversed by verified claim.'
WHERE id='<business uuid>';

-- 3. enter their figures against the CMA taxonomy (repeat per item)
INSERT INTO business_intel.product_prices
  (product_id, business_id, price_gbp, includes_vat, pet_band, product_url,
   source, observed_at, is_current)
SELECT p.id, '<business uuid>', <price>, true, '<any|cat|dog_small|dog_medium|dog_large|dog_xlarge|dog_giant>',
       '<their price page URL, or NULL>', 'claimed_listing', NOW(), true
FROM business_intel.products p WHERE p.slug='cma-1-first-consultation';  -- see: SELECT slug,name FROM business_intel.products WHERE cma_item ORDER BY cma_category;
```

Then re-run the exporter. Claimed prices supersede scraped ones automatically.

**Opt-out.** Same verification standard, then log and action — re-run the exporter promptly
afterwards, because "we'll remove them" is a promise with a clock on it:

```sql
INSERT INTO business_intel.claim_requests
  (business_id, request_type, requester_name, requester_email, evidence_method,
   evidence_note, status, verified_by, verified_at, notes)
VALUES ('<business uuid>', 'optout', '<name>', '<email>', 'callback_published_number',
        '<what you checked>', 'actioned', 'operator:<you>', NOW(), '<their words>');

UPDATE business_intel.businesses
SET publication_optout=true, optout_at=NOW(), optout_note='<why>'
WHERE id='<business uuid>';
```

Opt-out removes per-practice price display only. The practice stays in the directory and still
counts toward unnamed aggregates. A later verified claim reverses it (step 2 above).

## Price provenance rule (hard)

No price row is publishable per-practice unless product_prices.product_url is a real URL on the
practice's own domain and observed_at is set. Historical rows (762) have empty URLs → aggregates
only, forever. Any new scraper/verifier MUST persist per-price source_url or its output will
never be published. Never un-quarantine `source='seed_import'`.

## DB access

```
PW=$(kubectl get secret postgres-clients-secret -n ai-persona-system -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)
kubectl exec -i -n ai-persona-system postgres-clients-0 -- env PGPASSWORD="$PW" psql -U clients_user -d clients_db
```
Migrations 006–009 in this dir are all idempotent; apply with `-v ON_ERROR_STOP=1`.

## Dates that matter

- **30 Jul 2026 23:59** — funding-order consultation closes (draft response in this dir; owner
  submits via connect.cma.gov.uk portal; case team VetsMI@cma.gov.uk).
- **~Jul 2026** — substantive draft Order consultation expected; check the case page:
  https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review
- **23 Sep 2026** — statutory deadline for the Order; compliance clocks start.
- **~Dec 2026 / ~Mar 2027** — mandated price lists appear (large/small) → compliance-watch
  scraping + claim funnel.

## Editing a rendered component in the DB (fabrication / copy fixes)

The renderer builds pages from `page_components.rendered_html`. **Fixing the published file is
not enough** — the DB copy is republished on the next render. Fix the DB, then lock it.

Audit every component on the site first — the fabrication was in the `hero` slot, not the one
whose name suggested it:

```sql
SELECT p.name, pc.slot_name, LENGTH(pc.rendered_html) AS len, COALESCE(pc.lock_type,'-') AS lock,
       (pc.rendered_html ~ 'Mulberry32|makePostcode|PREFIXES|buildData')::text AS generator,
       (pc.rendered_html ~ 'representative sample|ownership data|independently owned|Price: Low to High')::text AS false_claim
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='vetcomparison.uk' ORDER BY p.name, pc.position;
```

**Read every hit before changing it.** Of three hits outside the homepage, only one was a false
claim about us; the other two accurately described the CMA's findings and obligations. A blanket
regex sweep would have corrupted correct content.

To write large HTML into the column: generate dollar-quoted SQL **locally** (a short python
script writing `UPDATE ... SET rendered_html = $tag$<html>$tag$, lock_type='permanent', ...`)
and pipe the resulting file into psql via stdin.

**GOTCHA that cost me the hero component:** `\set html` backtick-`cat file`-backtick inside a
piped psql runs `cat` *in the pod*, which has no such file. It wrote 0 bytes and still printed
`UPDATE 1`. Always re-query `LENGTH(rendered_html)` afterwards.

`page_components.data_path` exists but is empty fleet-wide — vestigial, do not build on it.

## Triggering a render — UNSOLVED

`rerender-pages` is `experimental`; `page-rerender` is active. Neither
`system.agent.site-builder.requests` nor `system.agent.page-rerender.process` produced an
orchestration state from a kcat trigger (2026-07-18). Renders currently only happen when the
build loop raises them. **If you solve this, record it here** — it is the missing verification
step for the bug-020 fix.

## Standing safety rails

- Nothing may export without an explicit domain (Go fail-closed + blanked DB configs). We do
  NOT own vetcomparison.co.uk — never reintroduce it.
- Scheduled tasks: `directory-export-json` is **ENABLED** (48h). The med retailer pipeline is
  being re-enabled one task at a time (owner direction 2026-07-23; see §Med retailer pipeline
  below). vet-sweep-continue, ch-vet-collect, vet-batch-verify remain disabled.
  Enable deliberately, one at a time.
- Before re-enabling vet-batch-verify: the sweep deny-list is now extended IN CODE
  (`scan_discovery_candidates.go` skipDomains, +5 families, v1.0.1151) — the REMAINING
  prerequisite is making the verifier persist per-price source_url. Note: sweep workers run as
  spawned pods, so their `agent_definitions.image_tag` must be ≥ v1.0.1151 when re-enabled or
  the deny-list additions won't be in the running code.
- 176 practices sit in `pending` with wheree.com websites awaiting genuine re-verification.

## Med retailer pipeline (revived 2026-07-23, provenance-first)

Three tasks, enabled one at a time in this order (discover → scrape → export; the export's
hard 14-day freshness window makes exporting stale data structurally impossible):

| task | target | payload | interval |
|---|---|---|---|
| `med-discover-urls` | med-url-discover-orchestrator (spawn) | `{}` | weekly |
| `med-scrape-prices` | med-price-scrape-orchestrator (spawn) | `{"batch_size":20}` | 21600s (~4-day full coverage of 304 listings) |
| `med-export-json` | med-json-exporter (in-process on business-intel) | `{"domain":"vetcomparison.uk"}` at enable; blank = fail-closed refusal | 48h |

- **v1.0.1151 changes** (`f82f8b425`): `typical_vet_price` STRIPPED from all export outputs
  (the fabricated family, LEGAL record `vet_price_est`); fail-closed provenance gate — any
  price row without a source URL + capture date is withheld and COUNTED, surfaced as the
  always-present `skipped_missing_provenance` field in `data/price-metadata.json`. That
  field's PRESENCE is the deploy proof (the stripped field was omitempty, absence proves
  nothing).
- **Two deploy artefacts must BOTH be current** or workers silently run old code:
  the business-intel deployment image (in-process export) AND
  `agent_definitions.image_tag` for the 8 med-* types (spawned pods use the def tag, not the
  deployed image — `spawn_actions.go`). Verify:
  `kubectl -n ai-persona-system exec <business-intel pod> -- sh -c 'strings /app/agent-chassis | grep -c skipped_missing_provenance'` (≥1)
  and `SELECT type, image_tag FROM agent_definitions WHERE type LIKE 'med-%'`.
- Force a run: `UPDATE scheduled_tasks SET last_triggered_at=NULL WHERE name='<task>';`
- Verify scrape by the artefact: fresh `business_intel.med_price_snapshots.collected_at` +
  `med_scrape_evidence` rows, then spot-check 2–3 prices against the live retailer pages.
- Verify export by the artefact: `curl https://vetcomparison.uk/data/price-metadata.json` —
  fresh `exported_at` AND `skipped_missing_provenance` present; every option in
  `medicine-prices.json` carries `url` + `collected_at`.
- Pre-flights before any enable: `FIRECRAWL_API_KEY` in the business-intel pod env
  (passthrough to spawned workers; secret `personae-default-secrets`); ollama-adapter up with
  `mistral-small3.1` (LLM fallback parser); git-adapter pods Running.
- Known cosmetic: export logs may Warn `view refresh failed` — `med_price_current` lacks the
  unique index CONCURRENTLY needs; harmless, the export queries snapshots directly.
- Dormant leak paths for the fabricated-figure family (recorded, deliberately not fixed):
  scrape still persists retailer-claimed `typical_vet_price` into `med_price_snapshots` (a
  DIFFERENT provenance class — extracted from the retailer's own page with evidence); the
  matview and snapshot `raw_data` expose it to any future consumer. Un-persisting it is a
  separate owner decision.
- Site surface: data files only. The medicine pages/calculator rebuild is a SEPARATE task
  (bug-020 class: a rebuilt tool must not invent data).

### Fidelity sweep — stored price must appear in its own evidence (bugs_closed/061)

The LLM fallback fabricated 212 snapshots over two eras (all quarantined in
`med_price_snapshots_quarantine_061`, 2026-07-24 — **retained deliberately**, it is the
reversible record; do not drop it).

> **UPDATED 2026-07-26 — the guard is LIVE (`v1.0.1165`), so this is a standing audit, not
> a stopgap.** It used to read *"until the parse-fidelity guard is live in a rolled image,
> re-run this after scrape activity"*. Now the write path refuses a price absent from its
> own evidence, so a PRICE_ABSENT row should be **impossible** — which is exactly why the
> sweep is still worth running: a hit now means the guard has a hole, not that the LLM
> misbehaved. Last run 2026-07-26: **2,577 OK / 0 PRICE_ABSENT / 0 UNCHECKABLE.**

Gotchas baked in: pair snapshot to SAME-RUN evidence (gap ≤120s, nearest `created_at`);
use `FM999999990D00` — the zero-less `FM999999D00` renders 0.42 as `.42` and false-OKs
sub-£1 fabrications; strip commas from the markdown, not the price.

```sql
WITH snaps AS (
  SELECT ps.id, ps.listing_id, ps.size_variant, ps.price, ps.collected_at
  FROM business_intel.med_price_snapshots ps
), ev AS (
  SELECT DISTINCT ON (s.id) s.id AS snap_id, e.markdown_content,
         abs(extract(epoch from (e.created_at - s.collected_at))) AS gap_s
  FROM snaps s
  JOIN business_intel.med_scrape_evidence e ON e.listing_id = s.listing_id
  ORDER BY s.id, abs(extract(epoch from (e.created_at - s.collected_at)))
)
SELECT s.id, l.retailer_product_name, s.size_variant, s.price, s.collected_at,
       CASE WHEN ev.snap_id IS NULL OR ev.gap_s > 120 THEN 'UNCHECKABLE'
            WHEN replace(ev.markdown_content, ',', '')
                 LIKE '%' || to_char(s.price,'FM999999990D00') || '%' THEN 'OK'
            ELSE 'PRICE_ABSENT' END AS verdict
FROM snaps s
LEFT JOIN ev ON ev.snap_id = s.id
JOIN business_intel.med_retailer_listings l ON l.id = s.listing_id
WHERE s.collected_at > now() - interval '14 days'   -- drop for full-table
ORDER BY s.collected_at DESC;
```

- Attribute a PRICE_ABSENT before deleting: the fabricated values sit verbatim in
  `llm_call_log` (`provider='ollama'`, `prompt_rendered LIKE '%Extract all product
  size/price variants%'`) at a timestamp matching the snapshot to the second.
- **Never purge by price value** — 19 genuine £17.48 rows exist alongside the 79
  example-echo fabrications; only the evidence check separates them.
- LLM-extracted rows carry `collection_method='scrape_llm'`; regex rows stay `'scrape'`.
- **To re-prove the guard on a live fabrication** (how it was verified 2026-07-26 — the
  Advocate category page still makes the model invent a whole price table, so this is a
  repeatable fault, not a one-off):

  ```sql
  UPDATE business_intel.med_retailer_listings SET last_scraped_at = NULL
  WHERE id = '0b50fd2d-a129-4edd-85a0-75843181fe0c';   -- petdrugsonline /advocate
  UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'med-scrape-prices';
  ```

  `loadMedListingsForScrape` orders `last_scraped_at ASC NULLS FIRST`, so nulling it puts
  the listing at the head of the next batch — necessary, because 247 of 305 listings are
  stale and the April-era backlog is ahead of anything scraped this month. **Budget ~10
  minutes, not 5**: the fallback call to the local CPU Mistral measured **495 s** on its
  own. Then assert `variants_found > prices_stored` on the new `med_scrape_evidence` row
  and grep the spawned worker's log for `fidelity guard dropped variant` — capture that pod's
  logs while it lives, the pod is ephemeral and GC'd shortly after.
