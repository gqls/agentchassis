# 291 — `tool-auditor` files dispatchable work at `hitl-review`, an agent that has never existed, and it is bleeding now

**Filed 2026-08-17** by the `bugs_open/284` lane, which found it while verifying its
own fix at the live table. **Status: OPEN, evidence measured, producer identified,
root cause NOT diagnosed beyond the producer** — per the 2026-07-31 owner ruling
this file asserts no structural root cause; it records what was measured first-hand
and names the `090` run as the next step. Grepped `/bugs_open/` and `/bugs_closed/`
for `hitl-review` first: **zero hits**, so this is not a duplicate. The
`needs_diagnosis` queue was empty at filing.

This is `bugs_closed/077`'s class alive again — *"a check that filed items at an
agent which had never existed, so every one was blocked at claim time"* — reached
this time from a **config** producer rather than Go.

## `[MEASURED]` 2026-08-17 — 14 rows, 2 sites, three days, still arriving

```sql
SELECT count(*), count(DISTINCT site_id), min(created_at)::date, max(updated_at)
FROM site_work_items WHERE status='blocked' AND error LIKE 'Handler agent not registered%';
-- 14 | 2 | 2026-08-14 | 2026-08-17 06:55:35
SELECT EXISTS(SELECT 1 FROM agent_definitions WHERE type='hitl-review' AND deleted_at IS NULL);
-- f   ← the handler every one of them names does not exist, and never has
```

**It is growing, not settled**: the same census on 2026-08-16 returned **5**; a day
later it is **14**. Newest arrival 2026-08-17 06:55.

Every row: `item_type='needs_human_review'`, `handler_agent='hitl-review'`,
`created_by='tool-auditor'`, `source='tool-auditor'`, `spec->>'check'='tool_auditor'`.
Their summaries are real, specific and useful findings about live tools — *"None of
the four number inputs (#w1, #h1, #w2, #h2) have associated labels"*, *"The entire
tool depends on `policy-generator.js` via a bare …"*. This is good audit output
being recorded as a routing failure.

## `[MEASURED]` They are BORN dispatchable — no promoter is involved

```sql
SELECT count(*) FILTER (WHERE spec ? 'original_pipeline') AS via_a_promoter,
       count(*) FILTER (WHERE triaged_at IS NOT NULL)     AS triaged_at_set
FROM site_work_items WHERE status='blocked' AND error LIKE 'Handler agent not registered%';
-- 0 | 0
```

Both zero, so nothing promoted them: they were filed straight into a dispatchable
status, claimed by the dispatch loop, and refused by
`claim_work_item_action.go`'s **handler-not-registered** branch (:191-214) — the
sibling of the empty-handler branch that `bugs_open/284` is about.

**So 284's fix does NOT cover this, and that is the point of filing separately.**
284 put the claim path's routability test at the *promoter*; these rows never go
near the promoter. It is the "born dispatchable" second path 284 records as
unclosed — with a machine producer instead of a hand insert, and a **named but
unregistered** handler instead of an empty one. Note also that the CHECK constraint
284 proposes as the eventual closer would **not** catch this either: a CHECK cannot
subquery `agent_definitions`, so "this handler exists" is not expressible there.

## Producers naming `hitl-review` — two, and one is Go

- **`tool-auditor`** (live `agent_definitions` config) — the producer of all 14 rows.
  `SELECT type FROM agent_definitions WHERE deleted_at IS NULL AND default_config::text
  LIKE '%hitl-review%'` → one row: `tool-auditor`.
- **`resolve_composition_layout_action.go:390`** (`handlerAgent: "hitl-review"`) —
  has produced none of these 14, but is the same mistake waiting to fire. Its
  comment at :375 says the pairing *"matches the existing"* convention, which is how
  a wrong value spreads: it was copied because it looked established.

## Why it matters

1. **The findings are lost.** `blocked` is not terminal, so each row holds its
   `(site_id, item_key)` slot in `idx_swi_dedup` — the auditor's later findings for
   the same key are dropped silently. The count understates the loss.
2. **Nothing will ever release them.** `feasibility-recheck` promotes a blocked row
   only `WHERE EXISTS (… ad.type = wi.handler_agent …)`; `hitl-review` will never
   satisfy that until an agent of that name exists.
3. **It is the live half of this family.** 284's class is currently paused because
   its driver (`improvement-sweep`) is disabled; this one is arriving daily.

## Fix candidates, ordered by what closes the door (pending diagnosis)

1. **Decide what `hitl-review` was meant to be** — a real agent, or the empty-handler
   HITL-terminal idiom the sibling checks use (`check_unverified_claims`,
   `check_voice_tells`: `HandlerAgent: ""` + `status: needs_human_review`, which is
   never dispatched and never blocked). If the latter, both producers are simply
   wrong and the fix is to stop naming a handler at all. **Measure before choosing:**
   `needs_human_review`'s other rows (5 cancelled, 2 at `needs_human_review`) carry
   an empty handler, so the idiom already exists in the data.
2. **Refuse an unregistered handler at the WRITE door**, not at claim — the check
   claim already performs, moved to `writeWorkItem`. Covers every future producer
   that uses the shared door; ~20 raw `INSERT INTO site_work_items` sites bypass it,
   so it is partial. (284's `workItemHandlerRegisteredSQL` is the predicate to reuse
   — do not write a fourth copy.)
3. **A minting ratchet for `handler_agent`**, the shape `bugs_open/279` built for
   `item_type`: a closed set plus a CI test, so a config that names a nonexistent
   agent fails before it ships.
4. **Repair the 14 rows** once (1) is decided — they are useful audit findings, so
   the repair is a re-file at the right shape, not a cancel.

**Next step: a `090` diagnosis run** — point it at `tool-auditor`'s live
`default_config`, `claim_work_item_action.go`'s handler-not-registered branch,
`resolve_composition_layout_action.go:375-390`, and the `needs_human_review`
population split by `handler_agent`. Check the queue first.
