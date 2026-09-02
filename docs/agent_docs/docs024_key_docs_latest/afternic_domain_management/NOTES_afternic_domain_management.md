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

**Spec CONFIRMED by the valuation session (same day)**, with one refinement
adopted: the currency ASSUMPTION must travel in the cell, not sit silently
in our docs — so the default currency value is now the literal
`USD-assumed`; a plain `USD` appears only once a real export confirms its
marking (then `--currency USD` explicitly). They will carry each figure's
date as the export-file date. Repricing division confirmed: desired prices
from their lane, bulk-XLSX from ours, upload click stays with the owner.
