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

## Standing safety rails

- Nothing may export without an explicit domain (Go fail-closed + blanked DB configs). We do
  NOT own vetcomparison.co.uk — never reintroduce it.
- All vet scheduled tasks are disabled: med-export-json, med-discover-urls, vet-sweep-continue,
  ch-vet-collect, vet-batch-verify, directory-export-json. Enable deliberately, one at a time.
- Before re-enabling vet-batch-verify: extend the sweep/verifier deny-list (wheree.com,
  bestlocalrated.co.uk, yelp.*, starofservice, threebestrated, allvets.co.uk, calmshops.co.uk)
  and make the verifier persist per-price source_url.
- 176 practices sit in `pending` with wheree.com websites awaiting genuine re-verification.
