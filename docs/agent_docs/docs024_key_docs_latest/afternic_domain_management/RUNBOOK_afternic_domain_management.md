# RUNBOOK — Afternic domain management (no-credential CSV loop)

Afternic (afternic.com, GoDaddy-owned, absorbed Dan.com) has no self-serve
seller API — see the PLAN. This lane runs on files the owner moves by hand:
portfolio export IN, bulk-upload template OUT. No credentials anywhere.

Lane paths (all under `docs/agent_docs/docs024_key_docs_latest/afternic_domain_management/`):
- `inbound/` — owner drops portfolio exports here, dated names please
  (`portfolio_2026-09-02.csv`).
- `snapshots/` — parsed snapshots the tool writes; newest is the baseline.

## §1 Owner: export the portfolio (each cycle, ~1 min)

1. Log in at afternic.com → Portfolio.
2. Use the export/download control to get the full domain list as **CSV**.
   If it only offers XLSX, open it and save-as CSV — this machine has no
   XLSX reader (no openpyxl, measured 2026-09-02).
3. Put it in the inbound directory — easiest from the Claude prompt with a
   `!` command (runs in your session, output lands in the conversation):

```
! cp ~/Downloads/<export>.csv docs/agent_docs/docs024_key_docs_latest/afternic_domain_management/inbound/portfolio_$(date +%F).csv
```

Then quote ONE value you can see on the dashboard (e.g. "relojistas.com
floor is 12000") so the session can pin it as a `--control` — that is what
proves the parse read the right columns.

## §2 Owner: fetch the bulk-upload template (once)

Portfolio → Bulk upload → download `bulk_upload_sample_v3.xlsx` (Afternic's
own template; blank cell = "don't change"). Save it into this lane's
directory. The generate half (PLAN P4) is not built until it exists.

## §3 Ingesting an export (any session)

```bash
scripts/domains/afternic-csv.py --self-test    # offline, proves mechanics

# known-domains cross-check file from the estate DB (excludes internal rows):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -At \
  -c "SELECT domain FROM sites WHERE domain NOT LIKE '%.internal'" > /tmp/estate_domains.txt

scripts/domains/afternic-csv.py ingest \
  docs/agent_docs/docs024_key_docs_latest/afternic_domain_management/inbound/portfolio_<date>.csv \
  --known /tmp/estate_domains.txt \
  --control relojistas.com:floor:12000
```

- `--control` is DOMAIN:FIELD:VALUE, repeatable; fields:
  `buy_now` `floor` `min_offer` `status` `lander` `verified`. **Refresh the
  value from what the owner quotes THAT day** — a control pinned to a stale
  price fails honestly, but a control never updated stops testing anything.
- `--baseline auto` (default) diffs against the newest snapshot in
  `snapshots/`; pass a path to pin, or `--baseline ''` to skip.
- Exit 1 = refused rows or a failed control. **Do not quote any figure from
  a run that exited 1.**

**STANDING HAND-OFF (requested by the domain_valuation lane, 2026-09-02):
after every successful ingest**, write their normalised feed, commit it by
pathspec, and tell them:

```bash
scripts/domains/afternic-csv.py valuation-csv \
  docs/agent_docs/docs024_key_docs_latest/afternic_domain_management/snapshots/portfolio_<date>.json
# writes docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/afternic_listings_<date>.csv
git add docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/afternic_listings_<date>.csv
git commit docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/afternic_listings_<date>.csv \
  -m "domain valuation inbound: afternic listings"
# then SendMessage to the "domain valuation" session (or note it in their
# lane docs if that session is gone)
```

Columns: `domain,price,currency,status,price_source` — price is buy_now,
else floor, else min_offer, and `price_source` names which (a floor is not
an asking price; the extra column is deliberate so the valuation can tell).
Currency: the cell carries **`USD-assumed`** until the first real export
shows its own currency marking — the valuation lane asked that the
assumption travel in the cell itself, not sit silently in our docs. Once
confirmed, pass `--currency USD` (or whatever the export says) and the
plain code takes over.

## §4 Verification + NS state (any session, no files needed)

Delegation is public DNS; reuse the existing classifier:

```bash
# which estate domains sit on marketplace (afternic/dan) NS vs live DNS:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -At \
  -c "SELECT domain FROM sites WHERE domain NOT LIKE '%.internal'" \
  | scripts/domains/classify_nameservers.py --check-http
```

It reports delegation and HTTP serving in separate columns — do not blend
them into one "parked" verdict (its own header explains why). For the wider
registrar estate, feed it an enumeration from the registrar helpers
(`scripts/domains/dynadot.sh`, `porkbun.py`, `spaceship.py` — the
domains_cloudflare_rollout lane's tools; read their lane docs first).

Afternic-side verification status ("In Verification" etc.) comes from the
portfolio export (§3) — the `verified`/`status` columns, once the first real
export confirms their header names.

## §5 Gotchas

- **Never read a dashboard PASTE positionally** — the tool exists because a
  2-cell paste under an 11-column header produced a false "Minimum Offer=0"
  (WRONG_CALLS 2026-07-28). If the owner pastes rather than exports, quote
  the labels back to him; do not map by position.
- **A parked domain 200s every path** (existing landmine): when probing what
  an Afternic/Dan-delegated domain serves, curl an invented URL as control —
  the lander answers 200 to everything.
- The v3 bulk template is **XLSX**; a generated upload must start from the
  owner's downloaded copy of the real template, not a reconstruction.
- Snapshot filenames carry the date; a second ingest the same day gets an
  `_HHMM` suffix rather than overwriting (append-only, like everything else
  in this estate).
- The relojistas `marketplace_url` in `site_specs` points at
  `forsale.godaddy.com/forsale/relojistas.com` — GoDaddy lander, same
  listing (Afternic listings surface through GoDaddy's network). Not a
  contradiction to find "relojistas" in an Afternic export.
