# Council `4d430ca8` — APPROVED r1, 13 seats, 5 advisory objections. Answers, with the checks run.

**Verdict: APPROVED, none high-severity, `gated_by_truncation: false`, 4 abstained, 0 unreadable.**
Trailer for the answering commit: `Council-Reviewed: 4d430ca8-7e34-479a-95f3-71fdc12fdef6`.

Every objection below was **checkable, so it was checked** — not filed as a follow-up. Two of the
five changed something I had written down.

---

## 1. Second producer — raised by FOUR seats independently (editquality, bug_historian, guardian, prior_art_librarian), all medium

**The objection.** I claimed a single producer (`check_voice_tells.go:142`) from reading code. The
landmine on co-dedup'd item_types says a retraction seam adopted against one producer's scanner will
**silently close a second producer's findings**, and `count(DISTINCT created_by)` cannot detect one
because the field is free text. Four seats wanted the codebase-wide search shown, not asserted.

**Answered — single producer, confirmed three ways.** The seats were right that I had not shown it.

| check | result |
|---|---|
| Go call sites filing this type | **ONE** — `check_voice_tells.go:114` (`ItemType: "voice_tells"`). No other `.go` file outside my own new one mentions it |
| Config producers (a seed or agent definition can file without Go) | **1** `agent_definitions` row matches `voice_tells` — that is the check's **enablement** in a `checks` array, not a second filing path |
| Live provenance | two `created_by` values — `generic` (25) and `quality-discovery-agent` (7) |

⚠ **The provenance census looks like two producers and is not.** `created_by` is `dctx.AgentType`,
so it records **which agent ran the check**, not which code filed the row. Both values have
`source='discovery'` and there is exactly one call site. **This is precisely why the landmine says
`created_by` cannot answer the question — it produced a false positive here, in the direction of
alarm.** The Go census is the primary evidence; the config census closes the non-Go path.

## 2. ⚠ `p.status IN ('active','deployed')` — debug_historian, medium. THE SEAT WAS RIGHT ABOUT THE MECHANISM.

**The objection.** That predicate is the shape of the documented `pages.status` vs `pages.build_status`
landmine, where `'deployed'` **never occurs** as a value of `pages.status`. If the value never
matches, the revalidator finds no page, returns `unknown`, and **the whole change reads as shipped
while being inert** — a WRONG_CALLS-shaped failure. The seat noted I never enumerated the values.

**Enumerated, 2026-08-08, live `clients_db` — the seat's claim is CONFIRMED:**

```
 status   | count
----------+-------
 active   |   585
 archived |    29
```

**`'deployed'` does not occur in `pages.status` at all. Half of that disjunct is dead code.**

**But the feared consequence is REFUTED:** all **32** affected items resolve to a page with
`status='active'`, which the predicate matches. The revalidator is not inert.

**Why the dead literal stays, deliberately.** It is inherited **verbatim** from the emit side
(`check_voice_tells.go`, pre-change), which is the point of sharing the scanner: the revalidator
cannot be narrower than the scan that filed the item. Removing `'deployed'` would change the emit
side's predicate too, in a commit about retraction, for no behavioural gain — a dead disjunct in a
shared predicate is a documentation problem, not a defect. **Recorded rather than silently fixed.**
The `archived` case behaves correctly by construction: an archived page returns no scan row, which
the ladder answers `unknown`, never `resolved`.

## 3. Crowd-out under the shared cap — guardian (medium) + guidelines (advisory note)

**The objection.** Registering the type widens `coveredItemTypes()`, and the sweep takes the oldest
N across the whole covered set. The `voice_tells` rows are dated **2026-07-17** — among the oldest in
the queue — so they could crowd out `unresolved_cta` / `required_fields_missing` /
`needs_section_data` / `needs_page` for one or more passes.

**Refuted by measurement.** Today's run: `scanned 151 · capped_at 1500 · cap_binding false`. Adding
32 rows takes it to ~183 against a cap of 1500 — **12% of the budget, with `cap_binding` false and
~1,300 slots unused.** Crowd-out requires a binding cap; this one has not bound since the stopgap
raised it on 2026-08-06, and `cap_binding` is the exact field to watch if that ever changes. The
guardian's reasoning is sound and would be correct at the old cap of 500; it is the raised cap that
makes it moot, which is a dependency worth knowing.

## 4. The generic gap — bug_historian, medium. ACCEPTED, NOT ACTIONED HERE.

**The objection.** This patches one map entry when the underlying shape is generic: an item_type
filed `HandlerAgent: ""` has no dispatch, no handler, no revalidator, and parks for ever.
`bugs_open/033` counts ~175 such "deliberate, documented escalations". The seat wants the
**registration contract** fixed — a lint or test requiring every `HandlerAgent: ""` type to have
either a revalidator or a documented exemption — rather than a fifth registration.

**Accepted as correct, and deliberately out of scope for this commit.** A test that constrains what
every discovery check may file is a change to a **shared contract**, which under the platform-seams
ruling is architecture-scope and does not belong inside an adopter. Two seats disagreed about this
in the same round — `architecture` explicitly recorded `point_fix` and *"uses the extension point as
designed … does not meet the RFC bar"* — and that disagreement is itself the signal that it needs a
human, not a resubmission. **Filed as a note in `bugs_open/033`, whose ~175 figure is the evidence
for it**, rather than actioned unilaterally here.

## 5. What became of the earlier flawed adopter — reuse_agent, advisory `missing`

**The objection.** The rationale says this lane previously shipped a retraction adopter on a false
premise; the plan does not say what happened to it, so a reader cannot rule out two closers for the
same class of finding.

**Answered from the lane record.** It was **reverted** — owner chose option A on 2026-08-04
(`b4c64f433`), which reverted `check_required_fields_missing`'s adoption and instead taught
`revalidate_review_queue` the `unresolved` status, so the better closer (the one with the
`content_data`-empty refusal and the stable `(page_name, slot)` key) owns the type alone. There is no
duplicate closer left standing. `HANDOFF_2026-08-04_continue_here.md` §4.

---

## ⚠ CORRECTION — a figure in this lane's own docs went stale within hours of my writing it

**I recorded "25 items" this afternoon. It is 32.** Seven more were filed **today**
(`2026-08-08`, `created_by='quality-discovery-agent'`) while I was building the revalidator, and I
found them only because objection 1 sent me back to the provenance table.

- **The check is actively filing, so this type grows** — it is not a fixed backlog of 25 to drain.
  That strengthens the case for a revalidator and weakens any sizing argument built on 25.
- **Churn re-measured over all 32: still 13** with a component edited since filing. The 7 new rows
  were filed today, so none of them can have changed since — the ratio moved (13/25 → 13/32) while
  the absolute count did not.
- The "25" in `NOTES`, `CQ-020`, the commit message and the submission's `grounded_in` was **true
  when measured and is now wrong**. Corrected where it is load-bearing; left in the submission JSON,
  which is an immutable record of what was submitted.

**The lesson is the one this estate keeps paying for: a population figure is a measurement with a
timestamp, not a property of the system.** Mine survived about four hours.
