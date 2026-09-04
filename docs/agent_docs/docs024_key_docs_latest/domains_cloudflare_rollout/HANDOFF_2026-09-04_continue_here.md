# HANDOFF — domains → Cloudflare rollout (Dynadot slice), 2026-09-04

Written for a fresh session with zero context. Read this, then the RUNBOOK
section "## Dynadot" for the exact commands, then NOTES only if you need the
narrative behind a decision. Do not re-derive facts already verified below —
re-verify only what has a shelf life shorter than "today" (marked ⏱ below).

## What this lane is

Wiring up registrar APIs (Nominet EPP, Dynadot, Porkbun, Spaceship) plus
Cloudflare, originally to move the whole domain estate onto Cloudflare. It has
grown a second live purpose: the **domain_valuation** and **sedo_domain_management**
lanes (separate sessions, separate directories, same repo) now depend on this
lane's registrar clients to build a portfolio pricing sheet for the owner. All
three sessions coordinate over cross-session messages with an
**announce-before-starting protocol** on the one shared, quota-limited
resource (Dynadot's appraisal API, 300 calls/day account-wide — see below).

## Where things stand — 2026-09-04, verified this session

- **All four registrar keys are in and reads are proven**: Nominet (EPP login
  proven, `nominet.py`, 1,606-domain tag inventory landed 09-03), Dynadot (this
  lane, below), Porkbun (`ping`+`listAll` proven, 683 domains, per-domain
  endpoints blocked pending an account-side API-access opt-in — not this
  lane's job to unblock), Spaceship (read paths proven — but see the ⚠ below,
  its own SellerHub account is now central to an open security question).
- **Dynadot inventory: 473 domains** (⏱ re-pull before trusting — the owner
  adds domains "on occasion", proven twice already: 451→453→472/473 across
  three same-week pulls, zero ever removed). Fresh pull:
  `./scripts/domains/dynadot.sh list_domain`.
- **OWNER RULING (2026-09-03): all Dynadot domains auto-renew, always.**
  Enforced by `scripts/domains/dynadot-ensure-autorenew.sh [--apply]` (dry-run
  default; re-run confirms **clean, 473/473**, verified minutes before writing
  this). ⚠ The account-level default (also set) is unproven for a
  transferred-in or marketplace-acquired domain — the one case that has
  actually failed once already (two 09-02 arrivals came in manual). Fresh
  .com *registrations* have been clean twice. If a domain arrives via
  transfer, run the sweep with `--apply` and don't assume the default caught it.
- **Dynappraisal (Dynadot's domain valuation API) is CONFIRMED to value ANY
  domain string** — not an inventory lookup, not even Dynadot-registered.
  This is why it can price the whole ~2,945-domain estate, not just Dynadot's
  slice. Quota is **300 calls/day, account-wide** (shared by whichever session
  runs it) — confirmed both by an induced stop and by two independent runs
  landing on exactly 300 and 289.
- **Appraisal progress so far: 589 rows** in
  `inbound/dynadot_valuations_2026-09-02.csv` (⚠ filename says -09-02, content
  is current through 09-03 — the directory's convention is dating by
  production date of the FIRST version, not by content date; check the file's
  own `source` column for per-row dates, they carry the real date). **~2,357
  domains still queued.**
- **Queue files ready to run, in `domain_valuation/inbound/`, priority-ordered
  (financial → home-garden → … → generic-word last)**:
  - `appraisal_queue_direct_2026-09-03.csv` — 1,482 rows, appraise the domain
    itself.
  - `appraisal_queue_proxy_2026-09-03.csv` — 875 rows, `.co.uk`/`.org.uk`/
    `.me.uk` names Dynappraisal can't value directly — has a `proxy_domain`
    column (the `.com` equivalent to appraise instead); the walker below
    handles this automatically.
  - `appraisal_queue_PREMIUM_direct_2026-09-03.csv` (69 rows) and
    `appraisal_queue_PREMIUM_proxy_2026-09-03.csv` (73 rows) — a further
    reprioritisation after the valuation lane found its scoring model was
    penalising the estate's best names; not fully investigated by this
    session, ask the domain_valuation lane before assuming these supersede or
    duplicate the main queues.
  - **Suggested run order** (valuation lane's call, this lane agreed):
    the 12 untested-TLD probes below → direct queue → proxy queue.
- **12 domains sit on untested TLDs** (`.org`, `.cv`, `.vin`, `.biz`, `.ai`,
  `.io`) — nobody has spent a call confirming Dynappraisal covers them.
  **Do one test call each before queueing them**, same pattern as the
  `.co.uk` discovery below. Confirmed working: `.com`/`.net`/`.uk`. Confirmed
  NOT working (HTTP 200, `"appraisal_price":"$--"`, not an error): `.co.uk`/
  `.org.uk`/`.me.uk`.

## ⚠⚠ THE ONE THING THAT MUST NOT BE DROPPED

The owner said, of 50 `.co.uk` domains: *"They have someone else's account
details in the listing and some of them have only just been added."*
**RESOLVED as to WHERE, still OPEN as to WHAT TO DO:**

- 44 of the 50 delegate to NamePros nameservers, 2 to Spaceship's own launch
  pair; both serve a live Spaceship for-sale lander (confirmed by fetching
  one: "Listed with spaceship.com", $4,999).
- **None of the 50 appear in the Spaceship SellerHub export from the account
  this estate holds a key for** — 831 rows, and that export does carry other
  externally-registered domains, so the absence isn't explained by these
  being Nominet domains. **A live listing exists under a Spaceship account
  nobody here has visibility into.** That's a real payout/transfer risk, not
  a UI glitch — a sale could pay or transfer to the wrong party.
- Dynadot is **cleared**: the account is provably the owner's, none of the 50
  have a Dynadot listing, and Dynadot's listing API has no seller/payee field
  at all (verified across every listing this session inspected).
- Full evidence: `domain_valuation/LISTING_ACCOUNT_2026-09-03_finding.md`
  (read it in full — the summary above elides the $4,999-pricing-convention
  detail that argues it's not a stranger, which matters for tone but not for
  what to do).
- **What only the owner can do**: log into Spaceship directly (not via any
  API key held here) and identify which account these are listed under.
  **As of this handoff, no owner reply has landed** (checked: zero commits
  mentioning Spaceship since the finding was filed).
- **⚠ DO NOT price or re-list any of these 50 domains until that is settled.**
  This instruction is already in the finding file; repeating it here because
  it is the kind of thing a fresh session skimming the queues could miss and
  accidentally violate — the 14 newest of the 50 ARE `.co.uk` names that
  would otherwise be perfectly normal candidates for the proxy queue above.

## Script inventory (`scripts/domains/`, registered as OPP-016)

- `dynadot.sh <command> [k=v…]` — legacy API3. Key from
  `~/.config/dynadot/credentials`, never printed. Exits 1 unless
  `ResponseCode 0`.
- `dynadot-restful.sh <METHOD> </restful/v2/path> [body]` — signed RESTful v2
  (HMAC-SHA256 over `API_KEY\npath\nrequest-id\nbody`). This is where
  Dynappraisal lives.
- `dynadot-appraise-all.sh <domains.csv> <valuations.csv>` — resumable walker.
  Finds the `domain` column BY NAME (never by position — a positional read
  would feed priority numbers or categories to the API on some of these
  queue files). Auto-detects an optional `proxy_domain` column for the proxy
  queue. Skips domains already in the output file, so re-running after a
  daily 429 resumes exactly where it stopped — **safe for any session to run,
  any day, without coordination beyond the announce-first courtesy.** A
  non-numeric appraisal (the `.co.uk`-shaped `"$--"` case) writes an explicit
  `no_appraisal` marker rather than junk or an infinite retry.
- `dynadot-ensure-autorenew.sh [--apply]` — the auto-renew-always enforcement
  sweep. Dry-run by default.

## Protocol with the other two sessions

- **Announce before starting an appraisal window** — one 300/day account-wide
  quota, three sessions can reach for it. Say so via cross-session message
  first; this has held cleanly so far, no double-spend yet.
- domain_valuation session owns the pricing model, categorisation, and the
  queue files. sedo session owns the Sedo bulk-import sheet. Neither reports
  to this lane in the org-chart sense — this is peer coordination, not
  hand-off-and-forget. If you pick up tomorrow's appraisal window, you are
  doing THEIR work using THIS lane's tooling; tell them what you did.
- Re-pull registrar data before any sheet the owner is about to see finalises
  — this got written into all three lanes' practice after a stale count
  nearly shipped once already (451 → 453 → 472/473, see NOTES).

## If you're picking this up cold, do this first

1. `git log --oneline -5 -- docs/agent_docs/docs024_key_docs_latest/domain_valuation/` —
   check whether the Spaceship account question has been answered since this
   was written; if resolved, that unblocks the 14 co.uk names for the proxy
   queue.
2. Check for a cross-session message already waiting (another session may
   have started today's window before you read this).
3. If clear to proceed: the 12 TLD test calls, then
   `dynadot-appraise-all.sh` against the direct queue, then the proxy queue —
   announce first.

Full narrative, every wrong turn and correction: `NOTES_domains_cloudflare_rollout.md`
(52KB, newest at the bottom — the 2026-09-02 through 2026-09-03 (night)
entries are this session's work). Mechanics/commands: `RUNBOOK_domains_cloudflare_rollout.md`.
