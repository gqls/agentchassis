# RUNBOOK — the post-delivery follow-up (bugs_open/477)

Every command here had to be got right once. The gotcha is attached to each; when one changes,
change it **here**, not in your scrollback.

---

## Before you enable anything — the three conditions, in order

The schedule ships **disabled** (`775`). All three must be true before the flip:

1. ~~**The owner has ruled on the interval.**~~ **DONE — THREE DAYS** (owner, 2026-09-04, verbatim:
   *"I think the follow up should be 3 days"*), carried in `775`'s agent config. It supersedes the
   "a week or so" of his original suggestion. The action still refuses to run without an explicit
   `followup_after_days`, so this is a recorded decision and never a default.
2. **The owner has been TOLD, in words, that enabling it emails him.** idea.uk is the only selectable
   site and its delivery address is `aaa@designconsultancy.co.uk`.
3. **An image carrying `send_followup_email` has rolled** — a seed naming an unregistered action
   fails at runtime.

```sql
-- (3), per SERVICE, not per fleet. The chassis is what runs actions.
SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
 WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
```
```bash
git merge-base --is-ancestor 0949244e8 <the git_commit above>   # exit 0 = safe to apply 775
```
> ⚠ **THE COMMIT IN THAT CHECK IS `0949244e8`, NOT `f89dfa31d`, AND THE DIFFERENCE CAN COST A
> CUSTOMER THEIR ONE FOLLOW-UP.** This line originally named `f89dfa31d`, the commit that added the
> action — which is necessary and NOT sufficient. `0949244e8` renamed the placeholder to
> `{{instructions_link}}`, and `775`'s seeded template uses that name. A binary between the two
> carries `send_followup_email` and would pass a `f89dfa31d` ancestry check, but does not know
> `{{instructions_link}}` — so the literal survives the fill, trips the post-fill `{{` scan, and that
> scan fires **after `ClaimFollowup` has already stamped `followup_sent_at`**. Result: the customer's
> single follow-up consumed, no email sent, recoverable only by the "stamped but never sent" section
> below. **Config that names a token must never go live ahead of the binary that can fill it.**
> (Raised by the `bugs_open/475` lane 2026-09-04 as a general rule for their own migration; it
> applies here more sharply, because my irreversible step precedes the scan.)
> ⚠ `service_binary_capabilities` is a **two-hour window**, not a history. It answers *what is running
> now*. If you are dating something older than two hours it will silently answer with today's
> survivors — corroborate with `kubectl -n ai-persona-system get rs -l app=agent-chassis --sort-by=.metadata.creationTimestamp`.

Then, and only then:
```sql
UPDATE scheduled_tasks SET enabled = true WHERE name = 'delivery-followup-send';
```

## Who would actually get an email, before you flip anything

**Run this first, every time. It names people.**

```sql
SELECT s.domain, s.handed_over_at, bq.direction->>'customer_email' AS would_email
  FROM sites s
  JOIN LATERAL (SELECT direction FROM build_queue
                 WHERE lower(domain)=lower(s.domain) ORDER BY created_at DESC LIMIT 1) bq ON true
 WHERE s.handed_over_at IS NOT NULL
   AND s.handed_over_at <= now() - interval '3 days'
   AND s.live_link_expires_at > now()
   AND s.transfer_confirmed_at IS NULL
   AND s.followup_sent_at IS NULL
   AND COALESCE(bq.direction->>'customer_email','') <> '';
```

> ⚠ **A ZERO HERE IS NOT INFORMATION ON ITS OWN.** Re-run it with `interval '0 days'` as a **demand
> control**. If the relaxed version also returns nothing, the zero is not about the calendar — it is
> the `build_queue` gap below, and the sender can select nobody at all. This is the only reason that
> gap was found; every other signal was green.

## The recipient gap — check whether it still bites

`build_queue` is the ONLY permitted recipient source (`651`'s header, corrected 2026-08-31 under
`bugs_open/420`; **`sites.email` is forbidden** — since the contract split it is the PUBLISHED
contact only). A site delivered outside the order pipeline has no such row.

```sql
SELECT count(*) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM build_queue bq WHERE lower(bq.domain)=lower(s.domain)
           AND COALESCE(bq.direction->>'customer_email','') <> '')) AS unaddressable,
       count(*) AS handed_over
  FROM sites s WHERE s.handed_over_at IS NOT NULL;
```
`[MEASURED 2026-09-04]` **1 of 1.** If those two numbers are equal, the sender has no reachable
population whatever. `775`'s verify prints this on every apply.

> **Do NOT "fix" it with `orchestration_states`.** It does hold the address a delivery used, and it
> passes every test you write this afternoon. But it is a QUEUE: `sql_for_agents/466` deletes
> `WHERE status IN ('COMPLETED','FAILED') AND updated_at < now() - INTERVAL '24 hours'`, so anything
> due days later reads a reaped row.
>
> ⚠ **And do not size that with `min(created_at)` — it is the wrong column, and it over-states the
> margin.** `now() - min(created_at)` reports the oldest SURVIVOR's birthday while the policy keys on
> `updated_at` (measured 1d02:41, on a survivor last updated 22:24 ago). **Ask the row you care
> about:** `SELECT updated_at + interval '24 hours' - now() AS time_left FROM orchestration_states
> WHERE <predicate>;`

> ⚠ **AND THE COLUMN A READER REACHES FOR IS POPULATED AND WRONG, WHICH IS WORSE THAN EMPTY.**
> `[MEASURED 2026-09-04]` idea.uk carries `sites.email = 'idea.uk@contactforsales.com'` — a site
> mailbox. The delivery went to `aaa@designconsultancy.co.uk`. There is no NULL to warn anyone:
> somebody answering a support or refund question finds a well-formed, plausible address and is
> confidently misled. `sites.email` is the PUBLISHED contact by definition (`bugs_open/420`) and is
> correct as itself; it is simply never the customer.

**FIXED — `site_deliveries` (migration `778`).** The recipient is now recorded in the SAME STATEMENT
that claims the handover (`StampHandover`'s CTE), so the record cannot exist without the delivery and
the delivery cannot happen without the record — `StampHandover` refuses an empty recipient outright.
Owner ruling 2026-09-04: a dedicated table, not `sites.delivered_to`, because `sites` is read by a
great many things and `bugs_open/420` exists to control which address lives where.

```sql
SELECT s.domain, d.delivered_to, d.delivered_at, d.recorded_by
  FROM sites s LEFT JOIN site_deliveries d ON d.site_id = s.id
 WHERE s.handed_over_at IS NOT NULL;   -- a NULL delivered_to here is a delivery we cannot trace
```
`recorded_by` tells you how much to trust it: `delivery-email` = written by the delivery itself;
`backfill-orchestration-states-778` = recovered from the run log before it aged out.

> ⚠ **ORDERING, and it is the one way this change can break production:** `778` must be applied
> **before** an image carrying the new `StampHandover` rolls. The table is not optional to that
> statement — without it every delivery fails at the claim. `778` is live-on-apply and the Go is
> inert until a roll, so applying it the same day closes the window.

## "Stamped but never sent" — finding a site the claim consumed and the send lost

The claim stamps `followup_sent_at` **before** the send, deliberately: for a chase email, not sending
beats sending twice. So a post-claim failure leaves a site marked with no email delivered. That is a
recoverable state, not a silent one — but only if you know where to look, because **triage reads the
work item and the reason is in the log**.

```sql
-- Every post-claim failure names its site in the message (added after the council
-- round asked how an operator would ever find one).
SELECT created_at, agent_type, left(error_message, 300)
  FROM agent_error_log
 WHERE error_message ILIKE '%followup_sent_at IS stamped%'
 ORDER BY created_at DESC LIMIT 20;
```
```sql
-- Cross-check against the stamps: any site here whose customer never replied and
-- which appears above was claimed and not delivered to.
SELECT domain, followup_sent_at FROM sites WHERE followup_sent_at IS NOT NULL ORDER BY 2 DESC;
```
**Recovery is a deliberate operator act**, never a retry: clear the stamp for that one site
(`UPDATE sites SET followup_sent_at = NULL WHERE id = '<site>'`) only after establishing the customer
did NOT receive it, then re-dispatch. Clearing it blind re-emails whoever did get one.

## What is APPLIED and what is not — check this before believing anything else here

`[MEASURED 2026-09-04 15:0xZ]`

| migration | what it is | applied? |
|---|---|---|
| `778` | `site_deliveries` — who a site was delivered to | **YES, live.** 1 row, backfilled |
| `774` | `sites.followup_sent_at` — the follow-up's claim column | **NO** |
| `775` | the follow-up agent + its disabled schedule | **NO** |

```sql
SELECT (SELECT count(*) FROM information_schema.columns
         WHERE table_name='sites' AND column_name='followup_sent_at')      AS mig_774,
       (SELECT count(*) FROM scheduled_tasks WHERE name='delivery-followup-send') AS mig_775,
       (SELECT count(*) FROM information_schema.tables
         WHERE table_name='site_deliveries')                              AS mig_778;
```

> **Why `774`/`775` being unapplied is safe, and why it also means the file may still be edited in
> place.** Nothing calls `delivery.ClaimFollowup` until `775` seeds the agent, so the Go referencing a
> column that does not exist is inert. And because `775` has never run, correcting its text — as was
> done for the owner's three-day ruling — genuinely changes what will be applied. **That stops being
> true the moment it is applied:** rewriting an applied migration changes the FILE and not the LIVE
> ROW, and the estate has landmines about exactly that. After `775` is applied, any further change to
> the interval or the letter is a NEW numbered migration that `UPDATE`s the live
> `agent_definitions` row. (Raised as an objection by the council's editquality seat on the
> assumption `775` was already live; it was not, and the query above is how you check rather than
> assume.)

## ⚠ The ordering hazard, and how to check it at the POD rather than by reasoning

`StampHandover` writes `site_deliveries` inside the claim. **Without the table, every delivery fails
at the claim.** The reasoning "migrations are live-on-apply, Go is inert until a roll" is correct and
is not evidence — a roll can lag or partially complete, and this estate has had real incidents from
trusting that shape. So check the binary, not the argument (council advisory, debug_historian).

```sql
-- Which commit is each chassis pod actually running? Filter by pod_name, NOT the
-- service column: that column also carries rows for other pods sharing the image.
SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
 WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
```
```bash
# Does the running binary carry the recipient write? ANCESTRY, not a literal grep.
git merge-base --is-ancestor 698b144fa <the git_commit above>   # exit 0 = yes
```
Then the ordering is safe iff **`site_deliveries` exists** (query above) **or no running pod is a
descendant of `698b144fa`**. Both true is fine; the dangerous state is a descendant pod with no
table, which today cannot happen because `778` is applied — and which is exactly what a rollback of
`778` would recreate. That is why `778`'s ROLLBACK refuses while the table holds rows.

> ⚠ `service_binary_capabilities` is a **two-hour window**, not a history: it answers *what is
> running now*. Corroborate anything older with
> `kubectl -n ai-persona-system get rs -l app=agent-chassis --sort-by=.metadata.creationTimestamp`.

## Why a new table rather than `customer_access_tokens`

Asked by the council's prior-art seat, and answered by looking rather than by reasoning from the
name. `customer_access_tokens` has thirteen columns and **not one of them is a recipient**:

```
id, site_id, purpose, token_hash, issued_at, expires_at, single_use, used_at,
use_count, revoked_at, created_by, stored_url, stored_url_expires_at
```

`created_by` is the ISSUER (`'delivery-email'`), not the customer. The table tracks a token's
lifecycle, not an identity, so extending it would have meant giving an access-token row a second,
unrelated meaning. `SELECT string_agg(column_name, ', ') FROM information_schema.columns WHERE
table_name='customer_access_tokens'` is the whole check.

## Applying the migrations

```bash
# ALWAYS dry-run first. It cost nothing and caught a real defect in 774's own verify
# block: a PL/pgSQL variable named is_nullable is AMBIGUOUS against
# information_schema.columns' own column, and Postgres refuses the whole block.
{ echo "BEGIN;"; sed -e 's/^BEGIN;$//' -e 's/^COMMIT;$//' <the migration>; echo "ROLLBACK;"; } \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
      psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
```
`774` before `775` — `775` refuses without the column. Both are idempotent
(`ADD COLUMN IF NOT EXISTS`, `WHERE NOT EXISTS`). Rollback sidecars exist and **refuse** if anything
has been sent or the schedule is enabled.

## Proving the claim's SQL semantics again after any edit

The whole safety story is two predicates in one UPDATE. To re-prove them, run the fixture script in
this lane's scratch recipe (five sites: due / confirmed / already-sent / too-early / window-closed),
inside `BEGIN … ROLLBACK`, and **keep the negative control**: re-run the confirmed site's claim with
only `transfer_confirmed_at IS NULL` removed and assert it now DOES claim. Without that control the
five zeros prove nothing — a fixture that could never have claimed anything gives the same result.

## Reading a council verdict for this lane

```sql
SELECT metadata->>'decision', left(body, 400) FROM diagnosis_artifacts
 WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report' ORDER BY created_at;
```
> ⚠ **`COMPLETED` is not a verdict.** A run killed by an estate-wide LLM outage lands on
> `current_step='complete_invalid'` with `status='COMPLETED'`, `error` NULL, and **no
> `council_report` row at all** — the `fix_plan` artifact still persists, because validation runs
> before the first LLM call, so "the artifact is there" proves nothing. Ask `__step_error`:
> ```sql
> SELECT collected_data->'__step_error'->>'failed_step',
>        left(collected_data->'__step_error'->>'message', 300)
>   FROM orchestration_states
>  WHERE collected_data->'input_data'->>'fix_correlation_id'='<corr>';
> ```
> `no reviewer produced a readable opinion` = killed, not judged: resubmit fresh once the fleet is
> healthy. Check health on the SUCCESS side and do **not** grep the message for a phrase — see
> `LANDMINES.md`, the usage-limit entry and its 2026-09-04 addendum (the documented needle
> `'%usage limit%'` returned **0** during a live 117-failure outage; that day's fault said
> `credit balance`).
> ```sql
> SELECT date_trunc('hour',created_at), count(*) FILTER (WHERE success) AS ok,
>        count(*) FILTER (WHERE NOT success) AS failed
>   FROM llm_call_log WHERE created_at > now() - interval '5 hours'
>    AND model ILIKE 'claude%' GROUP BY 1 ORDER BY 1 DESC;
> ```
