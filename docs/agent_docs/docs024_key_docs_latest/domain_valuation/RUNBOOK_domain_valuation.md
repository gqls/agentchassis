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

Commit by pathspec, message `domain valuation inbound: <registrar> list`.

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

## Prior-conversation mining

Transcripts live at `~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`.
Candidate sessions for the .co.uk/.uk valuation discussion (found by grep count
of valuation terms, 2026-09-02): 460a5226, 839df212, db85f55f, 48fb60ee,
7fe0cd84, a107ab07, e9ad9395 — all clustered 2026-08-05..08-14.
Gotcha: session titles go stale after /rename — 460a5226 is titled
"provenance step by step build tools" yet holds 84 valuation mentions; judge by
content, not title.
