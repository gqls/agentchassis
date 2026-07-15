# PLAN — vetcomparison.uk rebuild (generic chassis work)

**Date:** 2026-07-15
**Planned on:** high-effort model; **implementation intended for a cheaper model** — every phase
below is written to be executable without re-deriving context. Read the companion docs first:
- `HANDOFF_2026-05-18_vetcomparison_uk_planning.md` (same dir) — prior design; Go-A/Go-B never shipped
- `LEGAL_2026-07-15_vetcomparison_factual_record.md` (same dir) — what happened and the publication policy now in force
- `006_unify_prices_schema.sql` (same dir) — the surviving migration, still unapplied

## Hard constraints (from the owner — do not relax)

1. **Everything on the chassis must be generic.** No `vetcomparison`-named actions, no hardcoded
   domains. Config-driven per vertical + domain, so the same machinery serves the next
   comparison site. The existing bug of this class: `vet_med_export_action.go:56` defaults to
   `vetcomparison.co.uk`, a domain we do not own.
2. **No direct republication of scraped datasets** (copyright/database-right position, see LEGAL
   §7). Publish: (a) directory facts, (b) aggregates, (c) per-practice prices only where the
   practice has claimed the listing (licence) — attributed-scrape publication stays OFF until a
   solicitor signs it off.
3. **No price without provenance.** Source URL + capture date + retained evidence, or claimed
   consent. This is the policy that replaced the fabrication; it is not negotiable.

## Current state (verified 2026-07-15)

- Live site = honest directory of 2,579 verified practices (origin/master `92526ccd`), no prices.
- DB: 803 genuine current price rows (all with source_url) across 330 practices; 997 fabricated
  rows quarantined (`source='seed_import'`, `is_current=false` — never republish these).
- `business_intel.businesses`: 2,767 verified practices; has `is_claimed/claimed_by/claimed_at`.
- `insertPrice` (business_intel_actions.go:~1054) still writes deprecated `business_prices`.
- No `vet_export_json` action exists; no `sites` row for vetcomparison.uk (site not adopted).
- All med/vet scheduled tasks disabled.

## The regulatory schema (grounded — this defines the product's columns)

From CMA final report Part B (24 Mar 2026), Table 3.1 ¶3.74 and ¶3.88–3.95. The mandated price
list all UK practices must publish (~Dec 2026 large groups / ~Mar 2027 small) is **36 items in 5
categories**, **no free text allowed**, VAT-inclusive, typical-case pricing, against **six pet
categories**: cat, small dog <10kg, medium 10–25kg, large 25–40kg, extra-large 40–60kg, giant
>60kg (note: CMA guidance table merges cat with small dog into one <10kg band — verify against
the final Order before freezing; Order due by 23 Sep 2026).

**Category 1 — Consultation and preventative care (12):** first consultation (with duration);
repeat consultation; out-of-hours consultation (with provider name/link if outsourced); nurse
consultation (with duration); nail clipping; anal gland expression; microchipping; animal health
certificate (with multi-animal charges); vaccinations primary course; vaccination booster;
kennel-cough vaccination; pet care plan monthly cost (with inclusions link).
**Category 2 — Prescription, dispensing and administration (6):** prescription fee; additional
items same consultation; repeat prescription fee; dispensing fee; injection administration fee;
insurance administration fees.
**Category 3 — Surgeries and treatments (6):** dental assessment*; castration*; spay
(traditional)*; spay (laparoscopic)*; physiotherapy session; laser therapy session.
(* = standard checkbox disclosure of inclusions: GA/sedation, analgesia, hospitalisation,
pre-op bloods, post-op checks.)
**Category 4 — Diagnostics and laboratory tests (9, all incl. interpretation):** X-ray (incl.
sedation/GA + 3 images); ultrasound abdominal; echocardiogram; cytology ear swab; cytology fine
needle aspiration; basic urine screen; CT per body part (incl. GA); MRI per body part (incl.
GA); pre-anaesthetic blood test.
**Category 5 — End-of-life care (3):** euthanasia; communal cremation; individual cremation.

Other pegs: prescription fee cap **£21** (+£12.50 per additional medicine, same consultation);
ownership disclosure 6 months all sizes; written estimates ≥£500; practices submit data to RCVS
"Find a Vet" within 12 months (~Sep 2027) and RCVS shares it with **approved third parties**
(approval criteria ~Jun 2027; feed API/CSV/JSON ~Sep 2027; approved parties may not show paid
rankings). Large/small threshold = 15 FOPs (premises).

## Phases

Each phase is independently shippable and testable. Follow repo conventions: complex Go in
actions, thin workflow JSON, wrapper orchestrator (spawn → call → complete), `logger.Info`,
`-n ai-persona-system`, variable names in sync between workflow and action.

### Phase 0 — hygiene (small, do first)

**0a. Kill the unsafe export default.** In `platform/orchestration/actions/vet_med_export_action.go`
`parseMedExportConfig`: remove `Domain: "vetcomparison.co.uk"` default; return an error if
`config["domain"]` is empty ("med export requires explicit domain"). Same for repo_name if
defaulted. Acceptance: unit test that empty config errors; grep confirms no `.co.uk` literal.

**0b. Name-quality cleanup.** Some verified `businesses.name` values are scraped page titles
(e.g. "26 Vets in Birmingham - Compare Prices & …"). Write one SQL report to list suspects
(`name ~* 'compare|prices|best|top [0-9]|\|'` or length > 60), then fix by preferring
`trading_name` where sane, else truncate at " - "/"|" separators. Manual-review file for the
remainder. Re-export the directory JSON afterwards (same query as LEGAL §5) and redeploy via the
worktree pattern (below). Acceptance: no directory entry matches the suspect regex.

### Phase 1 — unified price schema + CMA taxonomy

**1a. Go-B from the 2026-05-18 handoff (spec unchanged, still valid):** rewrite `insertPrice` to
upsert `products (kind='service')` + `product_prices`; add the medicine helper reading
`verification_result.medicine_prices[]`; stop writing `business_prices`. Slugs per handoff §2.
**1b. Apply `006_unify_prices_schema.sql`** (idempotent) after Go-B deploys; verify row counts
per the SELECT at its end.
**1c. NEW — seed the CMA canonical taxonomy.** Migration inserting the 36 items above as
`products` rows: `kind='service'`, `slug='cma-<category-n>-<item-slug>'`, plus columns (or a
JSONB attrs field if adding columns is invasive): `cma_item boolean`, `cma_category smallint`,
`checkbox_disclosures text[]` for the 4 starred items, and a `pet_band` dimension on
`product_prices` (enum of the six categories + 'any'). This taxonomy is the mapping target for
all scraped/claimed prices and — later — the RCVS feed. Acceptance: SELECT returns exactly 36
`cma_item` rows in 5 categories.

### Phase 2 — generic directory/price exporter (replaces the never-built Go-A)

**One action, fully config-driven:** `directory_export_json` in a new
`platform/orchestration/actions/directory_export_action.go`, modelled on
`MedExportJSONAction`'s config/git plumbing (`sendMedExportToGit`) but generic:

```
config: {
  vertical: "veterinary_practice",     // business_verticals lookup
  domain: "vetcomparison.uk",           // REQUIRED, no default
  repo_name: "sites", data_path: "vetcomparison.uk/data",
  filters: { verification_status: "verified", require_website: true, require_postcode: true },
  outputs: {
    directory: true,                    // facts only: id,name,location,postcode,website,is_claimed
    aggregates: { enabled: true, min_n: 5, group_by: "postcode_area", items: "cma_item" },
    claimed_prices: true,               // per-practice prices WHERE is_claimed=true only
    attributed_prices: false            // scraped per-practice prices — OFF until solicitor sign-off
  }
}
```

Aggregates read `product_prices` joined to `products` (kind='service'), **excluding** any row
whose source is 'seed_import', grouped by postcode area, published only where n ≥ min_n
(k-anonymity + statistical honesty; emit the n alongside median/min/max). Register as
`"directory_export_json"`, Category "business_intel", IsLocal true. Agent pair
`directory-json-exporter` + `directory-export-orchestrator` per the standard wrapper pattern
(recreate the shape described for lost artefact 003). Scheduled task seeded **disabled**.
Acceptance: smoke-run against staging produces directory.json byte-compatible with the live
file's shape, an aggregates.json with no group under min_n, and claimed.json empty (no claims
yet); nothing in any output lacks provenance or consent.

### Phase 3 — claim flow (the business core, V1 deliberately manual)

- `claim_requests` table (generic, keyed to businesses.id): requester name/email/role, evidence
  (e.g. email domain matches practice website), status, notes, consent_text_version, timestamps.
- V1 process: "Claim your listing" on the site → structured email → operator verifies (call the
  practice's published number / domain-match) → operator marks `businesses.is_claimed`,
  `claimed_by`, `claimed_at`, records consent → operator enters the practice's 36-item price list
  into `product_prices` with `source='claimed_listing'`, `source_url` = practice's own price page
  where it exists.
- Exporter (Phase 2) then publishes those prices automatically. Claimed listings get a "prices
  provided by the practice, <date>" badge. Self-serve portal is V2 — out of scope.
Acceptance: one end-to-end dry run with a friendly/test practice record.

### Phase 4 — adopt the site onto the chassis

Only after Phases 1–2 (data pipeline honest end-to-end): adopt vetcomparison.uk via
site-adoption-orchestrator (handoff Phase 4 steps still apply — the classifier will treat the
now-live honest site as ground truth, which is what we want). Recreate the lost 005 classifier
`content_features` patch ONLY IF the current classifier prompt lacks a content_features block —
check first; other work may have landed it since May. Then fork directory + comparison tool
components from the tool library onto the site (components read the Phase 2 JSON artefacts).
Trap (from memory, fleet-wide): re-running build-site-planner regresses built pages — never
re-plan to fill gaps; the guides are hand-authored and adoption-locked content.

### Phase 5 — compliance watch (post-Order; the growth engine)

When practices' mandated lists go live (~Dec 2026 large / ~Mar 2027 small): a scraper that,
per verified practice, looks for the mandated price list (≤1 click from homepage, fixed 36-item
vocabulary, no free text — parse-friendly by design), stores observations with evidence, and
sets a per-practice compliance flag. The site then shows "price list: published / not found" —
which is both a genuine consumer service and the strongest claim-your-listing motivator.
Separately: when RCVS approval criteria publish (~Jun 2027), apply for approved-third-party
status; decision needed then on paid placement (feed recipients may not run paid rankings —
mutually exclusive with that badge).

## Standing decision points (owner's, not the implementer's)

1. Attributed-scrape per-practice publication: OFF until solicitor sign-off (LEGAL §8).
2. Paid placement vs future RCVS-approved status (Phase 5; no rush).
3. Aggregates min_n (default 5 until told otherwise).
4. Respond to the substantive draft Order consultation when it opens (expected July 2026) —
   argue for low-barrier third-party approval; watch the case page. Funding-order consultation
   (RCVS levy, less relevant to us) closes 30 Jul 2026.

## Deploy trap (repeat offender — follow exactly)

The `gqls/sites` local checkout is ~1,700 commits behind, dirty with other sites' work, and
render bots push to master continuously. Never push from it, never `reset --hard` it. Pattern:
`git fetch` → `git worktree add --detach <scratchpad>/sites-<task> origin/master` → edit/commit
in the worktree → `git push origin HEAD:master` → verify origin + live URL → `git worktree
remove`. If the push races a bot, re-fetch and cherry-pick onto the new head; never force-push.
