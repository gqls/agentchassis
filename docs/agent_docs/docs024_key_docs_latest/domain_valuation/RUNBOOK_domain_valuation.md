# RUNBOOK — domain valuation

## Inbound data contract (what registrar lanes were asked to deliver)

Each registrar lane writes to
`docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/`:

- `<registrar>_domains_<YYYY-MM-DD>.csv` — one row per domain:
  `domain,expiry,auto_renew,nameservers` (minimum: `domain,expiry`).
- `<registrar>_valuations_<YYYY-MM-DD>.csv` — if the registrar offers any
  appraisal/valuation: `domain,valuation,currency,source`.
- `<registrar>_listings_<YYYY-MM-DD>.csv` — if domains are listed for sale on
  that registrar's marketplace: `domain,price,currency,status`.
- Afternic feed extras (their lane, agreed 2026-09-02): 5th column
  `price_source` (buy_now|floor|min_offer|none); currency cell is the literal
  `USD-assumed` until an export confirms the marking, then plain `USD` —
  ingest must accept BOTH.

Commit by pathspec, message `domain valuation inbound: <registrar> list`.

**FILE DATES ARE PRODUCTION DATES — the day the file was WRITTEN, never the day
it is meant to be used.** This is the convention every inbound and output file
here follows, and it is stated because breaking it cost another lane real work:
two queue files were dated for the window they were built *for*, and on
2026-09-03 the sedo lane read `..._2026-09-04.csv` in this directory, concluded
that was today's date, and mis-dated a stretch of its own work and a LANDMINES
entry before catching it. Both files were renamed to their production date.
A future-dated filename is indistinguishable from a clock you can trust.

**Re-pull every registrar list the day the pricing sheet finalises** (owner
adds domains on occasion — proven 2026-09-02, when Dynadot grew 451→453 between
snapshots): a registrar count is stale by ADDITION, never by loss, and a fresh
name missing from the sheet is invisible unless you re-pull. Dynadot re-pull is
one ask to that lane; same for the others.

## Nominet list (owner-run; the lane's session cannot touch credentials)

    ! python3 scripts/domains/nominet.py login
    ! python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt

Gotcha: the first walk (2026-09-02 18:48) produced a 0-byte file on a
connection blip — check `wc -l` before trusting it. `--months 120` not 12:
ten-year registrations exist and a 12-month lookahead misses them.

## Finding the live registrar/lane sessions

`ListAgents` → send with `SendMessage` to the bare lane name (dynadot,
porkbun, nominet, spaceship, afternic). Replies arrive as
`<cross-session-message>` turns in this session.

## Dynappraisal daily window (quota 300/day, 429 at the cap; any session may run)

Sequence for each day until the estate is appraised (dynadot lane's script,
their `956708e70` — reads the "domain" column BY HEADER NAME, so both files
below feed it unchanged; a file with no domain column is refused):

1. One non-Dynadot test call (answers whether Dynappraisal takes foreign
   domains — unanswered as of 2026-09-02; 429-on-exhausted-quota is
   inconclusive, only a fresh window answers it).
2. `scripts/domains/dynadot-appraise-all.sh docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/dynadot_domains_2026-09-02.csv docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/dynadot_valuations_2026-09-02.csv`
   (finishes the 151 Dynadot stragglers; idempotent — skips rows already present).
3. If the test passed: same command with
   `inbound/appraisal_priority_2026-09-03.csv` as input and the SAME output
   file — appends, skips duplicates, financial/home-garden land top-down until
   the day's 429. Reset timezone unstated; the first successful call dates the
   window.

## The appraisal queues (as of 2026-09-03 — 8 windows of work left)

588 of 2,945 owned domains appraised. Two queues, in this order, one 300/day
window each. **Say in the dynadot lane's channel before starting a window** —
three lanes share one 300/day account and a collision is only discovered as a
429.

1. `inbound/appraisal_queue_direct_2026-09-03.csv` — **1,482 rows**, appraise
   the domain itself (.com/.net/.uk). Ordered financial → home-garden → … →
   generic-word/misc, so a part-window still finishes whole high-value blocks.
2. `inbound/appraisal_queue_proxy_2026-09-03.csv` — **875 rows** of
   .co.uk/.org.uk/.me.uk, which Dynappraisal refuses. Appraise the
   `proxy_domain` column (the .com string equivalent) and record the value
   **against `domain`, marked as a proxy** — it measures the keyword in the
   .com market, not the UK market, and must never be presented as a direct
   appraisal.

⚠ 12 domains on other TLDs (org, cv, vin, biz, ai, io) are untested — try one
of each before queueing them.

Rebuild both queues after any window: they are derived from
`WORKING_table.csv`, so they shrink as coverage grows.

## Prior-conversation mining

Transcripts live at `~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`.
Candidate sessions for the .co.uk/.uk valuation discussion (found by grep count
of valuation terms, 2026-09-02): 460a5226, 839df212, db85f55f, 48fb60ee,
7fe0cd84, a107ab07, e9ad9395 — all clustered 2026-08-05..08-14.
Gotcha: session titles go stale after /rename — 460a5226 is titled
"provenance step by step build tools" yet holds 84 valuation mentions; judge by
content, not title.
