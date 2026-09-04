# HANDOFF 2026-09-04 — Porkbun thread: continue here

The "porkbun" session (2026-09-02) delivered the third of the three registrar keys
this lane was owed and built the client for it. This file is the cold-start for
continuing that thread. Lane-wide cold-start remains the lane directory itself,
`RUNBOOK_domains_cloudflare_rollout.md` first — this handoff only covers the
Porkbun slice and points at where each fact lives.

## State — every claim re-verified 2026-09-04 unless dated otherwise

- **Credentials: IN and WORKING.** `~/.config/porkbun/credentials` (mode 600,
  `API_KEY=`/`SECRET_API_KEY=` lines, the RUNBOOK's stated convention). The
  check: `./scripts/domains/porkbun.py ping` → `SUCCESS — authenticated`
  (passed 2026-09-04). Key deliberately NOT IP-restricted (lane ruling
  2026-08-04 — the office line rotates both families; ping has returned a
  different IPv6 on each of the two days it has run, so the ruling is visibly
  load-bearing).
- **Client: `scripts/domains/porkbun.py`** (commit `5af348ef5`) — `ping` /
  `domains` / `ns` / `set-ns` / `dns` / `dns-create` / `dns-edit` /
  `dns-delete` / `check` / `raw`. Same family as `spaceship.py` / `dynadot.sh`;
  never prints key material; non-zero exit on any non-SUCCESS.
- **Account: 683 domains as of 2026-09-02, all ACTIVE, zero labels, zero listed
  on Porkbun's own marketplace** (whole-marketplace intersection, 43,203
  listings that day). 600/683 delegate to afternic-family NS — the portfolio's
  for-sale surface is Afternic-side, outside this API. Re-derive rather than
  quote: `./scripts/domains/porkbun.py domains` (~1 min, paginated).
- **Docs already carrying the detail** (do not fork second accounts):
  - `RUNBOOK_domains_cloudflare_rollout.md` — credentials table row; Porkbun
    section: global opt-in finding, marketplace mechanics (⚠ any filter param
    switches `/marketplace/getAll` to a mode that silently CAPS at 1,000 — page
    unfiltered, filter locally), no-appraisal-endpoint finding.
  - `NOTES_domains_cloudflare_rollout.md` — 2026-09-02 CONTRIBUTION entry
    (evidence trail), 2026-09-04 status line.

## The ONE open item on this thread

**The global API-access opt-in is OWNER-SIDE and still OFF (re-tested
2026-09-04).** porkbun.com → Account Settings → enable API access for all
domains. Until it flips, EVERY per-domain endpoint refuses — NS read/repoint,
all DNS CRUD — so no Porkbun domain can be repointed to Cloudflare or have
records managed. It does NOT gate `ping`/`listAll`/`marketplace`, so listing
work reads green while the write path is entirely closed; do not mistake one
for the other.

The check, any session, ~2s (exit 0 once flipped):
```
./scripts/domains/porkbun.py ns ai-agent-orchestration.com
```

**When it flips:** (1) run that check; (2) update the RUNBOOK's Porkbun section
and credentials-table row — both say "refused until the opt-in", and a stale
"gated" line will make the correct next action look premature; (3) exercise a
write only when there is a real repoint to do, and record it — reads proving
green says nothing about writes (this lane's own Cloudflare read-only-token
lesson, RUNBOOK 2026-08-25).

## Adjacent, for orientation — not this thread's work

- **domain_valuation** consumed this thread's deliveries (3 CSVs in
  `docs024_key_docs_latest/domain_valuation/inbound/`, all `_2026-09-02`:
  `porkbun_domains`, `porkbun_comps_uk` 774 rows, `porkbun_comps_com` 1,204
  rows) and confirmed COMPLETE for valuation purposes on 2026-09-02. That lane
  has since moved far past those snapshots (estate 3,029 domains as of
  2026-09-03, sell-cut rulings, a Sedo sheet generator with its own guard) —
  **the 683-row CSV is a dated registrar snapshot, not the estate picture; do
  not size anything from it without re-pulling.**
- **sedo** lane sources registrar lists via the valuation lane's inbound
  contract; it was told (2026-09-02) the opt-in gates writes only, not lists.

## Gotchas this thread added to the corpus

- `/marketplace/getAll` filtered-mode 1,000 cap (RUNBOOK Porkbun section).
- Marketplace `price` carries no currency field — USD is `[INFERRED]` from the
  site, and is recorded as such wherever quoted.
- Substring comps: `hire` matches `*shire`/`*hampshire`; `aluminium` AND
  `aluminum` both zero-match (real marketplace gap, not spelling).
