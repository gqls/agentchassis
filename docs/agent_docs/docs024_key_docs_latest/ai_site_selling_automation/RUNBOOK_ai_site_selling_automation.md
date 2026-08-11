# RUNBOOK — ai_site_selling_automation

Commands that were hard to get right, with their gotchas. Update HERE when one
changes.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Is the Anthropic account cap still in force? (one command, no cluster access)

```
curl -s -X POST https://preview.webdesign.uk/api/chat -H 'Content-Type: application/json' \
  -d '{"conversation_id":"","message":"how much does a website cost?"}'
```

A real answer → the cap has been lifted (LLM fleet is back; council gate
usable again). The "Thanks for your patience. Please reach us directly …"
contact line → still capped. Gotcha: HTTP status is 200 either way — the
fallback is fail-closed by design, so read the BODY, not the code. (Proven by
the webdesign lane 2026-08-10; see their HANDOFF_2026-08-10c §1 step 2.)

## Client rows, live

```
SELECT id, external_id, name, created_at FROM clients ORDER BY created_at;
```

As of 2026-08-10: 2 placeholder rows ("Default Client", "System Scheduler"),
1 network, 39 sites. Gotcha: today a customer's contact details land on the
SITE row (`sites.email/phone`, written by `082 --email`), not on clients —
until the columns migration in PLAN §2.1 ships, `clients` has no
customer-shaped columns at all.

## Submitting a build by hand (the only intake door)

```
scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh <domain> --email … --mission-file …
```

Gotchas (all from the start-here handoff §3.2): `client_id` is hardcoded
`demo_client` — the exact seam this lane exists to replace; `--fidelity` is
recorded but wired to nothing except `locked`; `hitl_mode=auto` synthesises
answers from classification defaults rather than merely skipping gates; and
dispatch itself is unreliable until `bugs_open/239` lands (queue starved —
the webdesign lane hand-drove every stage).

## Council submission: when a missing row IS a dropped dispatch (2026-08-11)

The standing rule "a missing orchestration row is almost always latency — do
not retry" assumes the publish SUCCEEDED. Distinguish by the kcat pod's exit:

- `% Delivery failed for message: Local: Message timed out` in the trigger
  output + `pod ... terminated (Error)` → the message NEVER reached the
  topic. Retry IS correct — with `RESUBMIT_CORR=<same corr>` so the printed
  correlation stays the one in your commit trailer. (Cause that night:
  kafka broker 0 was 0/1 NotReady; brokers 1-2 carried the retry.)
- kcat pod `Completed` + the trigger printing its full trailer-instructions
  tail → published; a missing row is queue latency, wait it out.

Prove the dispatch (not just the publish) by payload:
```
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```
Gotcha: don't pipe the 097 trigger through `tail`/`grep` when running it in
the background — the pipe buffers until EOF, so the correlation banner
(printed BEFORE the slow kcat step) stays invisible exactly when you want it.
