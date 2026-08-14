# HANDOFF 2026-08-14 — bugfix 213, continue here (supersedes the 08-13 handoff)

**Read this, then the 08-13 handoff's §"ROUND 1'S OBJECTIONS" and §"ITEM 2's DESIGN IS NO
LONGER OURS TO INVENT" — those two sections are still the reference and are not repeated
here.** `NOTES` (last four sections) has the measurements and every misstep; `RUNBOOK` (last
three sections) has the commands.

---

## THE ONE-PARAGRAPH STATE

D3 (the class detector) is live. **D1 half one — completion gate 1b — is BUILT, council
APPROVED, and LIVE on `agent-chassis` `v1.0.1298`, proven at the artefact on both replicas —
and it has never executed.** It cannot: `improvement-sweep` is switched off, and it is the only
thing that dispatches this item type. **D1 half two is designed and deliberately unbuilt** — its
shape is now dictated by `WII-016`, which the `bugfix_122` lane shipped mid-week, including a
recorded architecture ruling that a third hand-rolled copy (ours) must extract a shared helper
first. **Nothing is on fire and nothing is urgent**, for a reason worth stating up front: the
false-green bleed stopped on 2026-08-12 when that sweep was disabled for unrelated cost
reasons — *not* because of anything this lane shipped.

---

## ⚠ THE THING MOST LIKELY TO BE MISREAD

> Gate 1b did not stop the bleed. **A switched-off sweep did.**

Detection still runs — a 16th site was filed on 08-13 — so `dark_section_audit` items keep
accumulating; they simply are not dispatched, so nothing completes, so no false greens are
minted and the gate never fires. **Gate 1b's value is that it makes re-enabling
`improvement-sweep` safe.** Any sentence claiming it fixed the live bleed is false, and no doc
in this lane says one.

---

## WHAT IS LIVE, AND HOW TO CHECK IT WITHOUT RE-DERIVING ANYTHING

| thing | state | how to check |
|---|---|---|
| WII-013 instance fix (`Grades`) | LIVE since `v1.0.1290` | 08-10 handoff's probe recipe |
| WII-015 detector | LIVE, daily CronJob `25 7 * * *` | `doc_notes WHERE source='verifier-remit-check'` |
| **WII-017 gate 1b** | **LIVE `v1.0.1298`, NEVER EXECUTED** | three-needle binary probe, RUNBOOK §"Proving gate 1b shipped" |
| gate 1b's effect | **0 rows, correctly** | `improvement-sweep.enabled = false` — see below |
| WII-016 (122's retraction) | LIVE since 08-13, also unexercised | their entry; first fire ~08-17 14:54Z |

```sql
-- the whole explanation of the zero, in one row
SELECT name, enabled, last_triggered_at FROM scheduled_tasks WHERE name='improvement-sweep';
--  improvement-sweep | f | 2026-08-12 16:16:22+00
```

**Two probe traps, each of which cost a run on 08-14** (full detail in RUNBOOK): a per-needle
`grep -aq` loop **times out at 2 minutes** because each call rescans ~100 MB — use one
`grep -aoE 'a|b|c' … | sort -u`; and **never grep for your commit SHA** to prove your code
shipped, because the provenance stamp is a single sha (the build's HEAD), so unless the build
was cut at exactly your commit your sha is absent while your code is present. Also: this
service's `build provenance` startup line was gone from `--tail=6000` on **four-hour-old**
pods. Its absence means "not in range", never "unstamped".

---

## DECISIONS OWED BY THE OWNER — three, and two of them answer together

**1. How to get gate 1b's behavioural proof.** No unit test covers the wiring (it needs a
`*sql.DB`), and waiting will not obtain it because nothing dispatches these items.

- **(a) Re-enable `improvement-sweep`.** Not this lane's call — it was disabled after a
  measured 3.2x LLM cost surprise, and the `bugfix_122` lane owns a dated 08-16 pricing action
  against it. Cheapest in effort, most expensive in spend, and it restarts the bleed on every
  *other* ungated item type at the same time.
- **(b) Dispatch the ONE waiting row deliberately** —
  `6fe8a0fc-b9e5-4c96-b14d-9227a7827fa9`, `mortgagecalculator.co.uk`. A real dispatch of a real
  fixer at a real site, so it needs an explicit yes. **The cheapest real proof:** the handler
  will report `total_fixed 0` (0 of 61 bodies change, measured), the gate should block, and the
  row should land `triaged`/`failed` with *"completion blocked: the handler reported it changed
  nothing"*. ⚠ Assert **both** directions in one window — a gate that never fires and a gate
  that is not wired look identical.
- **(c) Accept presence + unit + mutation proof**, record the wiring as unexercised, move on.
  Honest, and below the estate's own bar, which is why it is listed last.

**2. `bugs_open/213`'s closure criterion — unchanged and still unanswered since 08-12.** The
recorded criterion needs a `hardcoded_section_colors` item with no `spec.check` to reach
completion, but Half A permanently moved that producer to `dark_section_audit`: **the fix
removed the traffic that would have demonstrated the fix.** (a) accept the unit + mutation
proof and close, recording the unexercised branch; (b) exercise it with one synthetic row on a
throwaway site; (c) leave OPEN, accepting the file no longer describes a reproducible defect.
**This is the same shape as decision 1 and should be answered with it** — one deliberate
dispatch could serve both.

**3. Whether up to 15 items landing `failed` is acceptable**, once the sweep is back on. That
is the honest state for a route that provably cannot repair them, but it is a population that
did not exist before. The alternative is a handler that can actually fix them, which does not
exist and is nobody's task.

---

## D1 HALF TWO — designed, not started, and the design is constrained

Full reasoning in the 08-13 handoff §"ITEM 2's DESIGN IS NO LONGER OURS TO INVENT". In one
paragraph: **site-scoped retraction after N ≥ 3 consecutive silences, built on a shared helper
extracted from `WII-016`, with the still-failing set taken BEFORE the filing filters.**

- **Site-scoped**, because `spec.page_name` is free prose (`all`, `global`, `sitewide`,
  `index / about`, three comma-joined slugs) and cannot be resolved to pages.
- **N ≥ 3**, because the audit re-reported the defect **7 of 7** times on post-closure
  re-visits, which bounds the per-run miss rate at ~35% — under 5% only at three. Not a safety
  margin; arithmetic.
- **Shared helper first**, because `WII-016`'s own architecture objection records that a third
  inline copy of "still-failing set before locks/caps, retract via `resolveWorkItems`" should
  extract one. We would be the third, and that objection is already on the record.
- **Set before the filters**, because `WriteAuditFindingsAction` drops findings through page
  classification, the dedup key and a cap; a set built after them reads "not filed" as "fixed".

**Talk to the `bugfix_122` lane before submitting.** They own the helper's shape by precedent,
and their entry explicitly hands this item type to `bugs_open/213`.

---

## STILL OPEN, NOT THIS LANE'S TO ADOPT

- **`RFC_024`** — nine CronJob meta-checks, no shared harness. Unclaimed.
- **12 live item_types in neither half of `verifier_coverage_test.go`** (89 rows) —
  `bugs_open/021` §INSTANCE 2.
- **The 10-of-14 payload split** (`bugs_open/213` §D) — mechanism NOT ESTABLISHED. Gate 1b now
  instruments it; the `agent_error_log` stream after the first dispatch is what will settle it,
  and it is expected to DOMINATE the gate's early output. That is the gate working.
- **Two `design-audit` rows vanished from `site_work_items` between 08-11 and 08-12.** No
  pruner exists in code. Unexplained.
- **`WII-016`'s index row still reads "verdict unread"** while its entry records APPROVED and
  LIVE. Another lane's row; left alone. Do not read the index as current for it.

---

## COMMITS, 08-13 and 08-14

`96c53bc18` gate 1b · test commit for the extraction (guardian's gating objection) ·
`4de91ad59` council follow-through (reuse the fleet path resolver, `WII-017`, index row,
`Council-Reviewed:`) · `ee5065b37` the WII-016 constraints on half two · plus the deploy-state
commit carrying this file. Evidence commits from 08-12: `5c27a85a2`, `13d0bc588`.
**`96c53bc18` carries no council trailer and cannot gain one** — the first submission attempt
died on an expired token before dispatching, and forward-only forbids the amend. Correlation
`0c8e7f5b-e510-4d24-893d-e3abb0bbb7b6` (APPROVED) is the join.
