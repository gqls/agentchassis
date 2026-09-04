# HANDOFF 2026-09-04 — Spaceship thread: continue here

The "spaceship" session (2026-09-02) delivered the last of the three registrar
keys owed this lane and built the client for it, then spent 2026-09-03/04
investigating the owner's account-details security flag alongside the
dynadot/domain_valuation lanes. This file is the cold-start for continuing
the Spaceship slice specifically. Lane-wide cold-start remains the lane
directory itself, `RUNBOOK_domains_cloudflare_rollout.md` first (its
"## Spaceship" section now has a SellerHub subsection — read it before this
file, this file is narrative/pointers, RUNBOOK is the mechanics).

## State — every claim re-verified 2026-09-04 unless dated otherwise

- **Credentials: IN and WORKING.** `~/.config/spaceship/credentials` (mode
  600, `API_KEY=`/`API_SECRET=` lines). Owner created the key in a separate
  terminal, never entered the transcript (his own 08-23 ruling on this).
- **Client: `scripts/domains/spaceship.py`** (commit `ef3157cec`) — `domains`
  / `info` / `ns` / `set-ns` / `dns` / `dns-put` / `dns-delete` / `raw`. Same
  family as `porkbun.py` / `dynadot.sh`; never prints key material.
- **Read paths proven, including SellerHub** (`sellerhub:read` is granted on
  this key — discovered by probing, not assumed). **Write paths
  (`set-ns`, `dns-put`, `dns-delete`) are COMPLETELY UNEXERCISED** — nobody
  has attempted a real Spaceship repoint. That is the one open item on this
  slice with no dependency on the security flag below: whenever the rollout
  is ready to touch a Spaceship-registered domain, that first repoint is
  this thread's next real step, and per the lane's established practice
  (Dynadot's first-write pattern) it should be tried on one low-stakes
  domain first and the result recorded here.
- **Domain inventory, re-pulled and diffed daily**: 203 (09-02) → 247
  (09-03, +44) → 251 (09-04, +4). API `total` field agreed with the fetched
  count all three days — no pagination undercounting. **All 251 are `.com`,
  zero `.co.uk`, on every pull** — a `.co.uk` question is never a Spaceship
  question, it's Nominet's. Registrant contact ID (`GET /v1/domains`'s
  `contacts.registrant` field) is the **same single value across every one
  of the 251 domains**, checked fresh each day including every new arrival,
  resolves via `GET /v1/contacts/{id}` to the account holder's own name and
  email — no exception found, ever. Durable CSVs (not the raw scratch JSON):
  `domain_valuation/inbound/spaceship_domains_2026-09-0{2,3}.csv` (09-03
  adds the `registrant_contact_id` column).
- **SellerHub: 831 listings, byte-identical across all three pulls** (09-02
  → 09-04, zero drift in status or price on any row, no new domain ever
  appears in it). Only **36 of the 831** names are actually registered at
  this account — the rest are listings for names registered elsewhere or
  since dropped; **never treat a SellerHub row as proof of ownership.** 7
  listings are live `onSale` (one keyword family, min-offer $10,000 each).
  **The SellerHub domain object has no seller/payee/bank/account-linkage
  field anywhere in the documented API** — checked against the full
  docs.spaceship.dev section list (Domains / Contacts / DNS / SellerHub /
  Hyperlift), not just the endpoints already called. This ceiling is
  permanent, not a today-only gap: whatever payout question comes up next,
  the answer is dashboard-only, every time.

## ⚠⚠ THE SECURITY FLAG — this slice's negative results were load-bearing, but the resolution and the open action live in the valuation lane's file

Owner, of 50 `.co.uk` domains: *"They have someone else's account details in
the listing and some of them have only just been added."* **RESOLVED as to
WHERE, still OPEN as to WHAT TO DO** — do not re-litigate this from scratch,
read `domain_valuation/LISTING_ACCOUNT_2026-09-03_finding.md` in full. This
slice's contribution, for orientation only:

- Ruled out that the registrant-contact field explains it (uniform, matches
  the owner's own identity, checked above).
- Established that SellerHub cannot carry a seller/payee field at all — this
  negative result is what forced the eventual approach (checking which
  *nameservers* the 50 flagged names actually use), which is what found the
  NamePros/Spaceship-launch-pair delegation and the real for-sale landers.
- The 50 do **not** appear in this account's SellerHub export, even though
  that export demonstrably carries other externally-registered domains (the
  795-of-831 figure above) — so their absence is not explained by them
  being Nominet domains, and the listings sit under a **different Spaceship
  account** this estate has no visibility into. Real payout/transfer risk.
- **What only the owner can do**: log into Spaceship directly (not via this
  API key) and identify the actual account. As of this handoff no reply has
  landed on that.
- **DO NOT price or re-list any of the 50** until he has — this instruction
  lives in the finding file too; repeated here because a session skimming
  the Dynappraisal queues (see the main HANDOFF) could otherwise treat the
  14 newest of the 50 as ordinary `.co.uk` proxy-queue candidates.

## If you're picking this up cold, do this first

1. `git log --oneline -3 -- docs/agent_docs/docs024_key_docs_latest/domain_valuation/` —
   check whether the owner has replied since this was written; if he's named
   the account, that's the actual resolution and this file's open item closes.
2. Re-pull before trusting any count above — one line: `./scripts/domains/spaceship.py domains --json | jq length`
   (SellerHub: swap `domains` for `raw GET '/sellerhub/domains?take=100&skip=0'` and paginate, or just diff against
   the last scratch pull if this session still has it).
3. If a real Cloudflare repoint is finally due for a Spaceship-registered
   domain: this is the FIRST write this client has ever made. Pick one
   low-stakes domain, run `set-ns`, re-read with `ns` to verify (the client
   already does this), and write the result into NOTES the way Dynadot's
   first-write was recorded (`541c61ba2`) — a write success on domain A does
   not mean domain B's works, same lesson as this lane's Cloudflare
   read/write-token split (RUNBOOK, 2026-08-25).

Full narrative: `NOTES_domains_cloudflare_rollout.md` (newest at the bottom
— the 2026-09-02 and 2026-09-04 "spaceship" entries are this thread's work;
the 2026-09-03 "resolved" entry is the dynadot lane closing the loop this
thread fed into). Mechanics/commands: `RUNBOOK_domains_cloudflare_rollout.md`,
"## Spaceship" section.
