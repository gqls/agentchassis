# Handoff — idea.uk contact-form spam + IP block list

## The problem
idea.uk's report-request form (the `/request` intake) is receiving spam: fake
submissions with test data and generated order ids, e.g.

    New idea.uk report request.
    Requester: test (test@test.com)
    Business: test
    Audience: test
    Notes: test
    Order id: ord_1783948426211007948

The owner wants (a) the existing spam removed and (b) an IP block list started.

## Where the data actually is (corrected)
The submissions ARE reachable — they are not on a third-party form service. Per the VM
deploy runbook (`RUNBOOK_idea_uk_chassis_site_and_vm_deploy_25_.md`): idea.uk's DNS points
at the VM; **nginx on the VM serves the static framework pages and reverse-proxies the
reserved tool paths (the `/request` intake among them) to the chassis Go process on
`127.0.0.1:8080`**. So the form posts THROUGH the chassis Go handler, and the submissions
(with their `ord_...` ids) are written by that handler — almost certainly into a store we
can query.

A first schema search in `clients_db` for table names matching
order/lead/contact/submission/request/form/report returned only internal orchestration
plumbing (`orchestration_requests`, `input_requests`, `http_request_log`, etc.) — i.e. the
intake table is under a name those needles didn't catch, NOT absent. Treat that as a
search miss, not a "no data" result.

## First steps for the next chat (schema-first, read-only)
1. **Find the intake store by its DATA, not its name.** The order id is the strongest
   handle. Search across tables/columns for the literal `ord_1783948426211007948` — e.g.
   enumerate `public` tables and, per candidate, check
   `SELECT count(*) FROM <t> WHERE to_jsonb(<t>.*)::text LIKE '%ord_1783948426211007948%'`.
   Also try table names around the tool-portal/intake domain: `%tool%`, `%intake%`,
   `%order%` was tried — extend to `%event%`, `%job%`, `%portal%`, and inspect the Go
   handler that serves `/request` (the reserved-path handler behind nginx on :8080) to see
   exactly which table/columns it writes.
2. **Read that table's schema** (`\d <table>`) for: the spam-identifying fields
   (email, the all-`test` payload, created_at) and — critically for the block list —
   whether it captures a **client IP / user-agent**. nginx proxies the request, so the IP
   may arrive as `X-Forwarded-For`; whether the Go handler PERSISTS it decides if an IP
   block list is even possible from stored data.

## The likely shape of the fix (decide on the read)
- **Removing existing spam:** once the table + columns are known, a guarded `DELETE`
  targeting the spam signature (`test@test.com`, every field literally `test`, or a burst
  from one IP/time window) — always run as a `SELECT` first to see precisely what it
  removes, then delete. Structural, in the reachable store; no third-party involved.
- **Block list:** only viable if the handler records the requester IP. If it does, see
  whether the spam clusters on a few addresses, then add a small blocklist the `/request`
  handler checks before accepting (a Go change in that handler + a `blocklist` table). If
  the IP is NOT persisted, the honest position is that IP-blocking needs the handler to
  start recording `X-Forwarded-For` first — and the more effective near-term defence is at
  the form/handler: a honeypot field, a minimum time-to-submit check, and/or a CAPTCHA
  (hCaptcha / Cloudflare Turnstile), which stop spam before it becomes a row. Spammers
  rotate IPs, so a honeypot + timing check usually outperforms an IP list.

## Working rules (unchanged)
Go not Python; British English; schema-first (`\d`) before any SQL; a 0-row/empty result is
not decisive until the query itself is cleared; structural over quick patches; reuse
existing tables/handlers before adding new ones; `logger.Info` not `logger.Debug`. DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
The `/request` handler is the chassis Go process reverse-proxied by nginx on the VM at
`127.0.0.1:8080`.

## Not part of this workstream
The chassis colour/imagery/page work is separate and tracked in
`HANDOFF_claude_code_continue.md` (the three open items: re-plan the three uncomposed pages,
the supervised fixer trial on dartsonline, the slice-4a re-seed check). This spam handoff is
self-contained.
