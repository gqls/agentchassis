# 327 — the build trigger can publish NOTHING and exit 0: one submission in three vanished with no orchestration, no work item, and no error

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane. **Status: OPEN, UNOWNED.**
Sibling of `bugs_open/326` (a re-submission that dedups also reports success) — together they
mean **"I submitted it" is not evidence that anything was submitted**, for two unrelated reasons.

> **On the 090 loop:** not run. The claim is an absence measured against a fleet-wide control in
> the same window, not a theory about a mechanism; the underlying `kcat -P` behaviour is already
> a documented landmine. What is new is that it hit the customer-facing build trigger.

## What happened

`bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh loanzy.uk`
printed its correlation and orchestration ids and exited 0. The kcat pod ran and was deleted
normally. **No `orchestration_states` row for that correlation ever appeared, and no work item
was created** — checked immediately, then again after 60s, then again at 10 minutes.

```
correlation 8fa2a4a6-2af1-4675-bae9-bbd59b702160 → 0 rows in orchestration_states
```

**The control that makes this a drop rather than a stalled fleet:** in the same ten minutes,
**29 orchestrations were created fleet-wide**, so the chassis was consuming normally. A second,
identical invocation minutes later (`f7e4dec3-…`) landed and filed
`needs_domain_research` / `research_loanzy.uk` / `triaged` within seconds.

**Rate on this evidence: 1 of 3 submissions from this lane on 2026-08-18.** Small sample, stated
as such — but the failure is silent, so the true rate fleet-wide is unknown and unknowable from
the logs today.

## Why it matters here rather than in a runbook

`kcat -P` publishing nothing at exit 0 is already in `LANDMINES.md`, and every session is told
to verify the item rather than the exit code. That is a fine rule for a session and a bad one
for a product: this is the **entry point of the one-shot build route**. The operator-visible
outcome of a dropped submission is indistinguishable from a successful one — same printed ids,
same "SAVE: CORRELATION_ID=…", same clean exit — and the customer-facing consequence is a build
that never starts while everyone believes it did.

## Fix candidates, ordered by what closes the door

1. **The trigger verifies its own landing.** After publishing, poll for the
   `needs_domain_research` row (or the orchestration row) for ~30s and **exit non-zero** if it
   never appears, printing the correlation to retry with. A submission whose evidence is the
   row it created cannot be silently dropped.
2. **Stop publishing through a throwaway pod.** The kcat-in-a-pod pattern is what makes the
   producer's failure invisible; a small producer that returns the broker's ack (or an HTTP
   endpoint on an existing service) would surface the failure at the source.
3. **Idempotent re-submission** — depends on `bugs_open/326`, and is the reason the two should
   be fixed together: today, "the submission vanished, try again" and "you already submitted
   this, so nothing will happen" both look like success, and the operator cannot tell which one
   they are in.

## How to verify a fix

Run the trigger with the broker deliberately unreachable (or the topic renamed) and require a
**non-zero exit** and a message naming what did not land. A fix verified only against a healthy
broker has tested nothing — the failure mode IS the unhealthy path, and it must be induced.
