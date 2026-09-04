# RUNBOOK — client accounts

Commands this lane had to get right, with the gotcha attached. When one changes, change it **here**.

---

## Is the ownership chain populated? (the lane's central question)

```sql
SELECT count(*) AS sites, count(network_id) AS with_network,
       count(DISTINCT network_id) AS distinct_networks
  FROM sites;
SELECT n.id, n.client_id, c.name FROM networks n LEFT JOIN clients c ON c.id = n.client_id;
SELECT id, name, email, external_id, tier, customer_status, created_at FROM clients ORDER BY created_at;
```

⚠ **`count(DISTINCT network_id)` is the load-bearing column, not `count(network_id)`.** "42 of 60
sites have a network" reads like partial attribution and is not: `[MEASURED 2026-09-04]` all 42 share
**one** network. A count of non-nulls cannot see that, and it is the exact shape that makes an empty
chain look half-populated.

## Which fleet sweeps can see a customer's site? (§1 of the OPTIONS paper)

```sql
SELECT count(*) FILTER (WHERE enabled)                                            AS enabled_tasks,
       count(*) FILTER (WHERE enabled AND pre_query IS NOT NULL)                  AS with_prequery,
       count(*) FILTER (WHERE enabled AND pre_query ~* 'network_id|client_id|customer|tier')
                                                                                  AS mention_customer
  FROM scheduled_tasks;
SELECT name FROM scheduled_tasks WHERE enabled AND pre_query ~* 'FROM sites' ORDER BY name;
```

⚠ **A regex hit is not a filter — open every match.** `[MEASURED 2026-09-04]` the two hits were
`evidence-register-absence` matching the word *"compliance-tiered"* **inside a prose comment**, and
`zip-link-refresh` matching the **table name** `customer_access_tokens`. Reporting "2 sweeps consider
tenancy" would have been false and would have weakened the paper's central argument. Show why each
matched:

```sql
SELECT regexp_replace(pre_query, '\s+', ' ', 'g') FROM scheduled_tasks WHERE name = '<task>';
-- then eyeball the surrounding 60 chars of each match
```

## Has a described mechanism ever actually RUN? (the lane's most expensive mistake)

Before repeating any register entry's mechanism as live — **select the column it names.**

```sql
SELECT count(*) FILTER (WHERE external_id IS NOT NULL AND external_id <> '') FROM clients;
SELECT id, status, provider, provider_customer_id, paid_at FROM billing_orders;
SELECT count(*) FROM billing_events;
```

⚠ **`status: deployed` in the concept register refers to the CODE, and reads as the DATA.** PAY-009
correctly describes Stripe's customer id landing on `clients.external_id`; `[MEASURED 2026-09-04]`
the column is empty because a one-off `mode=payment` charge with `customer_creation: "if_required"`
makes no Stripe Customer. Two sessions read that entry the same wrong way on the same day
(`WRONG_CALLS.md`, 2026-09-04).

⚠ **`billing_events.payload` contains the payer's real name, email, country and postcode.** Read it
in place. Never copy `customer_details` into a document, a commit message, or a chat reply.

## Council: did my submission actually get reviewed?

Submit per CLAUDE.md (`097_TRIGGER_council_review_v1.sh <submission.json>`), save `SUBMISSION_CORR`,
budget ~30 minutes. Then, **before writing any `Council-Reviewed:` trailer:**

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

⚠ **A run whose seats were all down ends `status='COMPLETED'`, `error` NULL,
`current_step='complete_invalid'` — which reads as "your submission was malformed" when it was
fine.** (Provider credit outage 2026-09-04 11:21:11–11:56:47 UTC killed 92 council-gate runs this
way; relayed by the `inter thread comms` session.) The distinguishing evidence is only here:

```sql
SELECT left((collected_data->'__step_errors')::text, 2000) FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

**The correlation is then SPENT** — no `council_report` is written, so a `Council-Submitted:` trailer
naming it reads un-reviewed for ever. **Resubmit and record the NEW correlation.**

## Working out what THIS lane has committed on a shared tree

```sh
for h in <your commit hashes>; do git show --name-only --pretty=format: $h; done | sort -u
```

⚠ **`git log --author=cqls` DOES NOT ISOLATE YOUR SESSION.** Every session on this machine commits as
the same author, so an author filter returns the whole tree's recent work and reads as yours. This
lane ran exactly that check to prove it had shipped no Go, got a list of ten Go files belonging to
other lanes, and would have concluded the opposite of the truth. **Keep your own hashes and enumerate
them**, or use the pathspec you committed with.

## Verifying against a fleet roll

- **`git merge-base --is-ancestor <your-commit> <the stamp>`**, where the stamp comes from
  `service_binary_capabilities` filtered by `pod_name` (⚠ two-hour window; see CLAUDE.md).
- ⚠ **"in the cut" ≠ "new in this roll"** — a change may already be an ancestor of the *running*
  binary. Check against **both** refs.
- ⚠ **Verify inert code by ANCESTRY, never by literal.** A built-but-uncalled module is dropped by
  the linker and probes ABSENT with clean controls.
