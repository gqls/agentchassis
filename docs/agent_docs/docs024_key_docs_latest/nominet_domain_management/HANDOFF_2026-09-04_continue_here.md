# HANDOFF 2026-09-04 — nominet_domain_management — CONTINUE HERE (first handoff, no prior to supersede)

**Read first, in this order:** this file → `SUMMARY_2026-09-04_nominet_domain_management.md`
(the arc read aloud) → `NOTES_nominet_domain_management.md` (the 2026-09-03 entries — the
3-bug walk saga, the cross-lane corroboration, the release/transfer episode) → `RUNBOOK_
nominet_domain_management.md` (the commands, §2 especially) → `PLAN_2026-09-02_nominet_
domain_management.md` (phases + cross-lane division) → `README_where_we_are.md` (owner's log,
last two entries). Paths relative to
`docs/agent_docs/docs024_key_docs_latest/nominet_domain_management/`.

## The one-paragraph state (2026-09-04 ~12:00Z)

**The lane is healthy and nothing is currently broken or blocking.** Every verb of
`scripts/domains/nominet.py` (probe/login/list/walk/check/info/set-ns) is live-proven against
the real Nominet registry, not just self-tested. The tag's full domain inventory has been
built (**1,606 domains**, `walk --months 120`) and delivered to two other lanes as a CSV,
independently corroborated by a completely separate export. A day-one incident (four domains
gone dark from a Nominet NS batch outrunning Cloudflare zone creation) was found and fully
resolved, including a public site launch. Two LANDMINES entries were filed for genuinely
dangerous bug classes (a proven login does not prove a command; a parser that has never seen
a real response can return `[]` forever with no error). One near-miss was caught and did NOT
turn into a bad write: the owner asked to "release" 50 domains, the registry showed they were
already on the tag, and it turned out he meant "transfer" — a different Nominet operation
entirely.

## Owning this lane — the model, in one paragraph

Sessions CANNOT run credentialed EPP or Cloudflare writes — the permission classifier refuses
them (correctly; a session piping a credentials file is indistinguishable from exfiltration).
**Stage the exact command, the owner runs it** with a `!` prefix in their own prompt. Read-only
checks (registry `dig`, Cloudflare GETs, EPP `probe`) run fine in-session. `nominet.py login`
was refused once and allowed once in the same evening — inconsistent, not relied upon; always
be ready to hand a command to the owner instead.

## NEXT — nothing urgent; two open items, both external

1. **Second Nominet TAG** — applied for 2026-08-11 (Channel Partner type, for customer-domain
   registrations under `site_delivery_and_editor`'s domain programme). Nothing heard since.
   Nothing to build; just ask the owner if it's ever natural to.
2. **Reconcile the 1,606-domain inventory against where each is parked** (`classify_
   nameservers.py`, already proven elsewhere in the estate) — low priority, PLAN P1's stated
   remainder. Not blocking anything.
3. Standing operational work as it arises: renewals, transfer-outs (the TWO-operation process,
   RUNBOOK §4), keeping the webdesign.co.uk `/domains/` registrar page factually honest against
   what the tag actually holds.

## Traps a fresh session needs, that are NOT obvious from a casual read

- **`epp.pl` and the three `idea_uk_vm_site/box/nominet-epp-*.py` scripts are FALLBACKS, not
  the primary tool.** They were fixed alongside `nominet.py` (same 3 bugs, same commit
  `316d83c4c` region) but the go-forward path is `nominet.py`. No retirement decision has been
  made — don't delete them, and don't be surprised they still exist.
- **`nominet.py register` deliberately refuses.** Domain creation costs real money and carries
  registrant rulings (owner-until-sale, D1) — that stays in
  `idea_uk_vm_site/box/nominet-epp-domain-register.py` (VMB-017) on purpose. Don't "fix" the
  refusal.
- **Cloudflare zone pair is NOT predictable from sibling zones.** New zones on this account
  get `betty.ns.cloudflare.com`/`ivan.ns.cloudflare.com`; the 32 older zones sit on
  `alexis`/`leah`. Always take the pair from the zone-create response
  (`cf-zone-bootstrap.sh` already does this correctly) — never assume from a neighbour.
- **A `list`/`walk` call's own `noDomains` attribute is a free correctness check** — `nominet.py`
  now asserts it matches the parsed count on every call (`assert_list_parse_matches`). If a
  future session ever sees this raise, STOP and read LANDMINES "A parser that has never matched
  a REAL response…" before touching the parser — it means the schema shape has drifted again,
  the same way it did on 2026-09-03.
- **"Release" (sponsoring tag) and "transfer" (registrant/legal ownership) are DIFFERENT Nominet
  operations**, however the owner phrases a request. Confirm which one is meant before writing
  anything — checking `domain:info`'s `clID` first (read-only, cheap) is how the 2026-09-03
  near-miss was caught before any action was taken.
- **A domain with an active Cloudflare zone is NOT necessarily a live site** (`apis.uk`,
  `ugg2.com` have zones + routes and nothing behind them). When another lane asks "which
  domains do you manage", give the zone list AND flag this caveat — don't let it be read as
  "which domains are live sites".
- **The published Nominet registry lags the EPP database by seconds-to-minutes** after a
  `set-ns` write. A stale first `dig` right after an EPP `SUCCESS` verify is propagation lag,
  not a failed write — re-check before concluding a write failed.

## Register + memory state (already current, no action owed)

- Register: **OPP-014** (`cf-zone-bootstrap.sh`) and **OPP-015** (`nominet.py`) in
  `docs026_concept_register/register/operator-practice.md`, both status lines reflect
  everything proven as of 2026-09-03. Index rows current.
- Memory: `nominet-domain-management-workstream.md` (topic file) and the `MEMORY_workstreams.md`
  index row are both current as of the 2026-09-03 evening's work. Nothing to update on pickup
  unless new facts land.
- Both 2026-09-03 LANDMINES entries came back `UNVERIFIABLE` from the automated verifier (it
  only covers Go code-symbols; ours are Python/Perl/docs) — expected, not a concern, nothing
  owed.

## Cross-lane state (all quiet, no threads waiting on this lane)

- **domains_cloudflare_rollout** — owns Cloudflare + the three retail registrars. Told them
  their read-only Cloudflare token (`~/.config/cloudflare/token`) is dead (9109); `portfoliotoken`
  is the live one. No reply owed back.
- **portfolio_positioning** — told them the dangling-delegation incident was resolved and that
  three of their held-brief domains (designblog/seotools/websitepromotion.co.uk) are serving
  content predating their remakes. Theirs to act on, not ours.
- **domain valuation** and **sedo** (peer sessions, not repo lanes) — both received the
  1,606-domain CSV, both confirmed receipt, both closed their threads. The Sedo lane also asked
  for and received the exact 46-domain Cloudflare-zone list (not ~40 — that figure was corrected
  live) to fence out of their import sheet.

## Files in this directory

`PLAN_2026-09-02` (phases + cross-lane division) · `RUNBOOK` (commands, §2 = the family client,
§4 = per-domain ops, §5 = Cloudflare ordering, §6 = token state) · `NOTES` (append-only technical
log — the 2026-09-03 entries are the ones worth reading in full) · `README_where_we_are` (owner's
plain-prose log) · `SUMMARY_2026-09-04` (this milestone, read aloud) · this HANDOFF.
