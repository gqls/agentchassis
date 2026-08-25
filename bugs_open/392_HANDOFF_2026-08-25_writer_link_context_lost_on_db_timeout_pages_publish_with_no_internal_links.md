# 392 — a DB timeout while loading link context publishes pages with NO internal links, and nothing notices afterwards

**Filed** 2026-08-25 by the `bugs_open/358` lane, on the owner's ruling of 2026-08-25 (decision 4:
commission a reader). **Routed at the `bugfix_092_writer_link_constraints` lane** (quiet 14d as of
filing — `scripts/who-owns.py 092`), whose own detector is what made this visible.

**Status** OPEN. Not diagnosed via `090` — no causal theory is asserted: the mechanism below is the
writer's own recorded statement of what it did, and every figure is `[MEASURED 2026-08-25]` with the
query inline.

## 1. The defect, in plain terms

When the page-content writer starts, it loads the list of pages it may link to
(`prepare_link_context_action.go`). If that load fails, the run **degrades deliberately**: the
writer is instructed to emit **no internal links at all**, and a durable row records it
(`LINK_CONTEXT_UNAVAILABLE`, born of `bugs_closed/092`'s fix — degrade-not-fail was that lane's
deliberate design, and this file does not argue with it).

The gap is **after** the degrade: the page ships without internal links, the row that says so
expires on the retention clock, and **no mechanism retries, re-renders, or even counts**. The
degradation is permanent-until-coincidence — the page stays link-less until something unrelated
happens to rerender it.

## 2. Evidence

```sql
SELECT occurred_at, agent_type, severity, context->>'site_id' AS site,
       context->>'failure' AS failure, context->>'outcome' AS outcome, context->>'degraded' AS degraded
  FROM agent_error_log WHERE error_code='LINK_CONTEXT_UNAVAILABLE' ORDER BY occurred_at;
```

`[MEASURED 2026-08-25]` — 2 rows, both 2026-08-24 14:21Z, **two distinct sites**
(`0a538b4a-803c-4f82-b298-d916f893fe8e`, `a998349c-6a55-45d5-8558-c0e6b63d915b`), both
`page-content-writer`, both `severity=error`, `degraded=true`, failure
`page query failed: query pages: FATAL: query timeout (SQLSTATE 08P01)`, outcome
*"writer instructed to emit NO internal links"*. Two sites hit within 24 seconds — the failure is a
shared-resource event (DB load), so **bursts are the expected shape**: one bad minute can degrade
every page being written across the fleet at that moment.

Detection lineage: the code was caught **undeclared** by the `finding-code-registry-check` CronJob
on its first live run (2026-08-24, ~2h after the first row — `bugs_open/358`), which is the only
reason anyone read these rows at all.

## 3. What is asked for — a READER, not a redesign

The write side is correct and stays: degrade-not-fail, plus the durable row. Commissioned
(owner ruling 2026-08-25): **something automated that selects `LINK_CONTEXT_UNAVAILABLE` rows and
acts.** Fix candidates, ordered by what closes the door:

1. **Heal**: a consumer in the `cmd/content-loss-check` family (the estate's proven
   reader-with-writer exemplar, `bugs_open/355` A2/A3) that, per unresolved row, checks whether the
   page(s) written in that run now carry internal links, and files a `page-rerender` work item if
   not — then resolves the row (**extract first: `resolved=true` halves remaining life to 14 days**).
   The site_id is in `context`; the writer could also be extended to record the page id, which would
   make the join exact instead of site-wide.
2. **Retry at the source**: one bounded retry of the page query inside `prepare_link_context`
   before degrading. Cheap, halves the incidence, does not remove the class (the second timeout
   still degrades silently) — a mitigation alongside 1, not instead.
3. **Count**: alarm on rows/day. Weakest — it tells a human, it does not fix a page.

**Acceptance**: a new `LINK_CONTEXT_UNAVAILABLE` row leads, without human action, to the affected
page(s) carrying internal links again (candidate 1), verified at the served artefact, not at the
work item's status. Registry follow-up: flip the code's entry to `consumed` with
`reader: <file:line>` and `reader_sink: agent_error_log` in the same commit that ships the reader —
the checker verifies both (`DBG-075`).

## 4. Traps for whoever picks this up

- **The rows expire.** 365 days under migration 567 (the code is declared `human-evidence` today),
  but a consumer that resolves rows drops them to the 14-day clock — extract what you need first.
- **A rerender is not proof.** Verify links at the served page; a rerender that hits the same
  timeout re-degrades and re-records, which is the loop working, not failing.
- The registry entry (`finding_code_registry.json`) carries the disposition and this bug's number;
  keep them in step or the daily check will say so.
