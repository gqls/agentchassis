# NOTES — Afternic domain management (append-only, newest at the bottom)

## 2026-09-02 — lane opened (session "afternic", owner-directed)

Owner asked: "help me setup the afternic api so you can manage my domains
there".

**Finding that reshaped the task: there is no Afternic seller API to set
up.** Checked before answering (web, 5 searches + fetches):
- partnerportaldocs.afternic.com — Fast Transfer / DLS APIs, registrar
  partners only.
- blog.afternic.com 2025 review + portfolio-agent post — seller-side
  automation is "Portfolio Agent", an in-dashboard conversational agent, not
  an API.
- domainnamewire.com 2025-03-19 — self-brokering, dashboard-only.
- GoDaddy developer Aftermarket API — allowlisted resellers; endpoint doc
  404s publicly; estate registrars aren't GoDaddy anyway.
- Unofficial route exists (dashboard JSON endpoints, e.g.
  `afternic.com/fosv2/api/v1/sales`, session-cookie auth) — offered to the
  owner as option (a), not chosen.

**Owner ruling (in chat, 2026-09-02): route = no-credential CSV loop;
scope = all three** (listings+pricing, sales+leads, verification+NS).

**Prior art found and reused rather than rebuilt:**
- `scripts/domains/` family — sedo-api.sh landed TODAY (OPP-012, sibling
  lane), registrar helpers dynadot/porkbun/spaceship (domains_cloudflare_rollout
  lane; Dynadot enumeration proved 451 domains today).
- `scripts/domains/classify_nameservers.py` — already classifies
  afternic/dan delegation; covers the NS-state scope as-is.
- Estate shape measured from the live DB (not carried from docs):
  `sites` = 57 rows, 40 real domains (17 `.internal` pool/system);
  `site_specs` aspect `commercial`: **2** rows as of 2026-09-02 —
  finetuning.uk (tier 2, not for sale), relojistas.com (tier 2,
  `for_sale_requested: true`, marketplace_url = forsale.godaddy.com lander).
  So price-by-tier desired state barely exists yet; P4's source needs owner
  decisions.
- Portfolio scale [from classify_nameservers.py's own measured comment,
  2026-07-30 registry export]: 1,567 domains, 998 on dan.com NS — Afternic
  (which absorbed Dan) is where the bulk of the portfolio parks. **1,567 as
  of 2026-07-30** — re-enumerate at the registrars before quoting.

**Built:** `scripts/domains/afternic-csv.py` (OPP-013) — export ingest:
header-name mapping only, refuses short rows (the WRONG_CALLS 2026-07-28
positional-paste shape), `--control` value assertions, `--known`
cross-check, dated snapshots + auto-baseline diff.
- `--self-test`: 13/13 PASS 2026-09-02.
- Mutation check: replaced the cell-count guard with `if False:` → self-test
  FAILS ("short row REFUSED, not padded") — the guard is load-bearing, not
  decorative.
- `[ASSUMED]` the portfolio export is CSV with headers resembling the
  owner's 2026-07-28 dashboard paste (Minimum Offer, Sale Lander, Views,
  Leads, 30-day Searches). The parser REFUSES rather than guesses if wrong;
  first real export locks `ALIASES` — record additions here.

**Open questions:**
1. Real export format/headers — waiting on owner's first export (RUNBOOK §1).
2. Does the dashboard offer a separate sales/leads export, or is the
   portfolio export the only feed? Owner to check while in there.
3. P4 desired-state source for the ~1,500 non-estate domains (owner
   decision, PLAN P4).

## 2026-09-02 (later, same session) — domain_valuation lane hand-off agreed

Cross-session request from the **domain valuation** session (working for the
owner): they are valuing the whole portfolio and repricing for a
**bottom-~500 sale**; our Afternic asking prices are wanted as a comparison
column. Owner's words, relayed: the Afternic prices are *"generally
overpriced and I will want to change them"* — so they are an INPUT to the
valuation, not the answer, and **this lane's bulk-XLSX route is the likely
vehicle for the eventual repricing** (PLAN P4 has a named customer now).

Agreed + built: `valuation-csv` subcommand — after every successful ingest,
write `domain_valuation/inbound/afternic_listings_<date>.csv`
(`domain,price,currency,status,price_source`), commit by pathspec, message
"domain valuation inbound: afternic listings", notify their session. The
`price_source` column is one more than they asked for, deliberately: price
is buy_now-else-floor-else-min_offer, and a valuation comparing against a
FLOOR without knowing it would read low listings as low askings. Flagged to
them in the ack. Self-test now 17/17 PASS (4 new cases); one test fix on the
way in was the TEST's own newline handling (`read_text()` translates the
csv module's `\r\n`), not the writer.

## 2026-09-03 — first real export ingested; headers LOCKED; feed delivered

Owner dropped `inbound/domains-1788424049.csv` (378 KB, **1,634 domains**,
UTF-8 with BOM — `utf-8-sig` handled it) and quoted three dashboard values:
veterinarypractice.co.uk $50k, redesign.co.uk $120k,
mortgagecalculator.co.uk $100k.

**Header set (LOCKED, real export):** Domain, Buy Now Price, Floor Price,
Min Offer, Lease to Own, Max Lease Period, Sale Lander, Show Buy Now
Option, Show Lease to Own Option, Show Make Offer Option, Hidden, TLD,
Date Added (UTC), Listing Status, Fast Transfer, Views, Leads, 30/90/365-day
Unique/Total Searches, GoDaddy NS. All price/status/lander/views/leads
columns mapped by existing aliases; one alias added (`dateaddedutc`);
13 columns deliberately unmapped → extras (search stats are
"Members-only feature" strings anyway).

**Semantics finding: `0` in a price column means NOT SET** (Afternic's own
bulk vocabulary; the quoted domains have Buy Now=0, Floor=0 and the quoted
value in **Min Offer**). Parser updated (0 → None, price fields only);
self-test 18/18 incl. the new case. Without this fix the valuation feed
would have carried a $0 asking for every BIN-less domain.

**Controls 3/3 PASS** on `min_offer` — with one discrepancy flagged to the
owner: the export has **veterinarypractice.uk**, not `.co.uk`; the `.uk`
row carries exactly the quoted $50k, and no `.co.uk` variant exists in the
file.

**Portfolio shape (export dated 2026-09-03, 1,634 rows):** all status
Listed; Buy Now set on **419**, Floor on **420**, Min Offer on **1,634**;
leads on 12 domains (top: kilocars.com 4, makeitaquote.com 4,
nanangmrk.com 4). Estate cross-check: **26/41** estate site domains present;
missing 15: cookly.uk, designblog.co.uk, farmerinsurance.uk, finetuning.uk,
lendzy.co.uk, leopardessconsulting.co.uk, loancalculator.co.uk,
loancash.co.uk, loanzy.uk, oxenunity.com, robot-hands.com, vetcomparison.uk,
vonc.com, webdesign.co.uk, webdesign.uk.

**relojistas.com observation:** today BIN 45000 / floor 28000 / min_offer
12000. The 07-28 owner statement "floor = $12,000" matches today's
MIN OFFER — either repriced since July, or the July label was Min Offer.
Not relitigated; recorded so the next reader doesn't treat the two as a
contradiction.

**Valuation feed delivered:** `domain_valuation/inbound/afternic_listings_2026-09-03.csv`
— 1,634 rows; price_source: buy_now 419, floor 3, min_offer 1,212;
currency cell `USD-assumed` throughout (the export carries no currency
marking — the assumption stands).

**Valuation lane's cross-check back (2026-09-03, after ingesting the
feed):** **689** of the 1,634 Afternic rows are NOT in their **1,339**-domain
retail-registrar estate (as of 2026-09-03) — 688 UK names (340 .uk,
337 .co.uk, 10 .org.uk, 1 .me.uk) + 1 .us. So the Afternic export is also a
**~half preview of the still-unenumerated Nominet portfolio, with minimum
offers attached** — the registrar walk of Nominet (domains_cloudflare_rollout
lane's EPP territory) will be the reconciliation. They hold the 689 as
provisional rows pending that walk; the veterinarypractice.uk-vs-.co.uk flag
is parked for the same reconciliation. Feed semantics confirmed applied:
buy_now → their live_ask, min_offer kept as a floor-like column, all figures
dated 2026-09-03.

## 2026-09-03 (later) — 6 listings queued for removal (owner ruling on NameSilo, relayed via valuation lane)

Owner ruling, relayed: *"namesilo exists but has nothing interesting in it,
we can ignore them. Any domains still listed there were probably lost."*
NameSilo is out of scope; anything traced there counts as LOST, not stock.
Combined with the valuation lane's registry sweep, that names **6 Afternic
listings advertising domains the owner cannot deliver** — verified present
in the 2026-09-03 export (all Listed, Min-Offer-only):
- `chicklets.co.uk` ($10,000), `demisexual.uk` ($10,000),
  `protecty.co.uk` ($30,000) — RDAP 404, registered to nobody
- `cheapbuild.co.uk` ($10,000, → Voove), `enables.co.uk` ($50,000, →
  123-Reg) — registered to someone else, live sites on infra we don't run
- `qlp.us` ($50,000) — at NameSilo per today's ruling

**QUEUED for the next bulk change (P4), not a special trip** — no removal
mechanism exists yet (P4's generate half is unbuilt). Do NOT remove via a
one-off dashboard action either; let it ride the same bulk-XLSX run as the
repricing so it's one auditable file, not a silent edit.

**NOT for removal** (peer's own caveat, keep it attached): `pocketvaginas.com`
(genuinely at Dynadot, just absent from their inventory listing — ours,
unexplained but not lost), `healthinsuranceconsultant.co` /
`studentloandebtsettlement.co` (UNDETERMINED — .co publishes no RDAP and
whois egress is blocked here). None of the three are Afternic listings
anyway (checked: not in the 2026-09-03 export) — the caveat is about a
different sweep, kept here only so nobody in this lane conflates it.

**Spec CONFIRMED by the valuation session (same day)**, with one refinement
adopted: the currency ASSUMPTION must travel in the cell, not sit silently
in our docs — so the default currency value is now the literal
`USD-assumed`; a plain `USD` appears only once a real export confirms its
marking (then `--currency USD` explicitly). They will carry each figure's
date as the export-file date. Repricing division confirmed: desired prices
from their lane, bulk-XLSX from ours, upload click stays with the owner.
