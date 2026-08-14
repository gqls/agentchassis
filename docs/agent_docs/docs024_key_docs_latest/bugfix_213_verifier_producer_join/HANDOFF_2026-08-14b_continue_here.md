# HANDOFF 2026-08-14b — bugfix 213, continue here (supersedes 08-14 and 08-13 on STATE)

**This is the cold start. Read this file, then `SUMMARY_2026-08-14…md` for the read-aloud
version.** The 08-13 handoff remains the reference for two things and only those: its
§"ROUND 1'S OBJECTIONS" table, and §"ITEM 2's DESIGN IS NO LONGER OURS TO INVENT". `NOTES` last
six sections carry every measurement and every misstep.

---

## THE ONE-PARAGRAPH STATE

**D1 half one is FINISHED.** Completion gate 1b is live on `agent-chassis` `v1.0.1299`, council
APPROVED, and **both its arms are proven on real production traffic with a 1:1 accounting** —
3 items blocked, 4 abstained, every abstention matched to a completion to the millisecond.
**Nothing is running and nothing is spending:** the `improvement-sweep` that drove that traffic
was enabled at 14:15Z on the owner's instruction, measured at **6.0x baseline**, and switched
back off at **16:41:46Z** on the owner's decision once they had the real number. Two things are
now waiting on other people: a **diagnosis-loop run** on the §D payload mystery (correlation
`266be67d-a6e1-4afc-8fc1-84b553b2ea82`, verdict unread), and **D1 half two**, which is designed
but must be built on a shared helper extracted from another lane's live code, so it starts with
a conversation rather than an edit.

---

## WHAT IS DONE — gate 1b, and the proof is complete

*A fix that changed nothing is not a fix.* The gate refuses to stamp `complete` on a
`dark_section_audit` item when the handler's own report says it altered no page and no template.
Opt-in per `item_type`, unsafe default OFF, no verifier registered, no browser, no page fetch.

| | evidence |
|---|---|
| in the binary | three-needle single-pass probe on **both** replicas: own literal present, long-live control present, nonsense needle absent |
| **block arm** | 3 items. The deliberate one: promoted 14:09:25Z → blocked 14:11:26Z with both handler counters `0` → `failed` at attempt 3, 14:14:41Z. **Terminated rather than churned** — that was a claim in my council risks block and is now measured |
| **abstain arm** | 4 items completed unblocked and **all 4 have a matching `agent_error_log` row**, timestamps agreeing to milliseconds (14:24:46.072 vs .079). No silent holes |
| council | APPROVED round 2, corr `0c8e7f5b-e510-4d24-893d-e3abb0bbb7b6` — **see the note on the trailer below** |

**Before this gate, every one of those items would have read `complete`.**

⚠ **Trailer note:** `96c53bc18` (the gate itself) carries **no** council trailer and cannot gain
one — the first submission attempt died on an expired token before dispatching, and forward-only
forbids the amend. Later commits carry `Council-Reviewed:`. If the `098` report lists that commit
as un-reviewed, this is why; the correlation above is the join.

---

## WHAT IS WAITING ON SOMEONE ELSE

### 1. The §D diagnosis — filed, verdict unread

**Run correlation `266be67d-a6e1-4afc-8fc1-84b553b2ea82`** (use this for artifacts, NOT the
intake correlation the trigger printed first).

The question: **a work item completes carrying the output of a step that did not handle it.**
`dark_section_audit` items routed to `color-variable-fixer` complete with a `result` holding
either a design-system spec (`color_scheme typography spacing design_notes`) or a page-structure
triage decision (`not_actionable retype_existing update_spec add_to_page new_page approach
reasoning`) — neither is that handler's envelope. The abstain arm reproduced the split **live at
3:1**, matching the historical 9:1, so it is **systematic, not a historical artefact**.

What is established first-hand and is in the symptom: **neither `color-variable-fixer` nor
`build-dispatch-loop` declares a `complete_work_item` step** in
`agent_definitions.default_config->'workflow'->'steps'`, yet `build-dispatch-loop` is the
`agent_type` on every abstain row. So the site binding the `result` input is unidentified. **I
stopped there rather than guessing** — read the verdict before forming a theory.

### 2. D1 half two — designed, and the first deliverable is not the feature

**Site-scoped retraction after N ≥ 3 consecutive silences, on a shared helper extracted from
`WII-016`, with the still-failing set built BEFORE the filing filters.** Every clause has a
reason and they are in the 08-13 handoff. The short version:

- **Site-scoped** because `spec.page_name` is free prose (`all`, `global`, `sitewide`,
  `index / about`, three comma-joined slugs) and cannot resolve to pages.
- **N ≥ 3** because the audit re-reported a known fault on **7 of 7** post-closure re-visits,
  which bounds the per-run miss rate at ~35% — under 5% only at three. Arithmetic, not a margin.
- **Shared helper first** because `WII-016`'s own architecture objection records that a third
  inline copy of that pattern should extract one, and ours would be the third.
- **Set before the filters** because `WriteAuditFindingsAction` drops findings through page
  classification, the dedup key and a cap; a set built after them reads "not filed" as "fixed".

**Talk to the `bugfix_122` lane before submitting.** Their entry hands this item type to
`bugs_open/213` explicitly, and they own the helper's shape by precedent.

---

## DECISIONS OWED BY THE OWNER

**1. How `bugs_open/213` closes.** Unchanged since 08-12 and still the real blocker to closing
the file. Its recorded criterion needs a `hardcoded_section_colors` item with no `spec.check` to
reach completion, but Half A permanently moved that producer to `dark_section_audit`: **the fix
removed the traffic that would have demonstrated the fix.** (a) accept the unit + mutation proof
and close, recording the one unexercised branch; (b) exercise it with one synthetic row on a
throwaway site; (c) leave OPEN, accepting the file no longer describes a reproducible defect.
**Today's production proof does not settle this** — it proves the *new* gate, not that original
branch.

**2. Whether `dark_section_audit` items settling as `failed` is acceptable long-term.** Currently
2 `failed`, and more will follow whenever dispatch resumes. Honest for a route that provably
cannot repair them; the real fix is a handler that can, which does not exist and is nobody's task.

---

## THE SWEEP EPISODE — read this before re-enabling anything

`improvement-sweep` was **off** before 14:15Z (the `bugfix_122` lane disabled it on 08-12 after a
measured 3.2x), **on** 14:15:23Z–16:41:46Z at the owner's instruction so gate 1b could be
exercised, and is **off again** on the owner's decision once measured properly.

| | input tokens/hour |
|---|---|
| baseline, the 4h15m before | **56,480** |
| the 2h26m it ran | **331,026** (164 calls, 807,704 in, 198,912 out) |
| ratio | **6.0x** |

**`page-content-writer` was 80% of it** — 57 calls, 541,676 input tokens, and **zero** in the
preceding 4h15m. The sweep triggers page content *rewrites*, not just audits.

Three things to carry, all paid for:

1. **I quoted 3.2x as the basis for the owner's decision and the real figure was 6.0x.** That
   3.2x was another lane's ratio against a ~248k/h baseline; today's baseline is 56k/h. **A cost
   ratio does not transfer between days — the denominator moves.** Quote a ratio with both
   absolutes or not at all.
2. **My first-pass price was 6.5x too LOW** (~52k/h vs 339k/h). A staged pipeline's cheap steps
   complete and log first, so **an early price is a lower bound with unbounded error**, not an
   estimate. One `GROUP BY agent_type ORDER BY input_tokens DESC` found the dominant cost
   immediately.
3. **`enabled=false` is not "spend stopped".** At the moment of the switch there were 1 `claimed`
   item and 3 orchestrations still `EXECUTING_STEP`; they finish on their own and land spend
   afterwards. Check those, not just the flag.

The `bugfix_122` lane has been told all of it — three appended notes in their handoff, the last
confirming their 08-16 measurement window is clean again.

---

## TRAPS THIS LANE PAID FOR (beyond the 08-12 list, which still stands)

1. **A printed correlation is NOT evidence of a dispatch — three times now, three different
   tools.** The council trigger printed `SUBMISSION_CORR` then failed on an expired token
   (nothing queued). The `090` trigger printed `CORRELATION_ID` then failed on
   `invalid input syntax for type json` because my symptom text contained **double quotes**
   (nothing filed). Both print before they publish. **Verify the row exists.**
2. **`orchestration_states` has no `id` column** — it is `orchestration_id`. A poll loop with
   `2>/dev/null` swallowed that error for twelve iterations and printed empty lines, which reads
   exactly like "never dispatched".
3. **Never grep for your commit SHA to prove your code shipped.** The provenance stamp is a
   single sha (the build's HEAD), so unless the build was cut at exactly your commit your sha is
   absent while your code is present. Probe a distinctive literal of your own change instead,
   with both controls. And a per-needle `grep -aq` loop **times out at 2 minutes** — one
   `grep -aoE 'a|b|c' | sort -u` scans once.
4. **This service's `build provenance` line was gone from `--tail=6000` on FOUR-HOUR-OLD pods.**
   Absence means "not in range", never "unstamped".
5. **Three instrument failures in three days, one family** (all in `WRONG_CALLS.md`): a join key
   renamed by my own change, read as absence; a mutation harness scored by a grep that could not
   see top-level failures; a cost window that closed before the cost arrived. **Each was caught
   by looking a second time, none by suspicion.** The rule earned: **when a number will be
   quoted to someone else, compute it a second way before sending it.**

---

## COMMITS (08-12 → 08-14)

Evidence: `5c27a85a2`, `13d0bc588`. Gate: `96c53bc18` (no trailer, see above), the extraction
test, `4de91ad59` (reuse + `WII-017` + index, `Council-Reviewed:`). Constraints on half two:
`ee5065b37`. Then the deploy-state, production-proof, correction and shutdown commits, all
carrying `Council-Reviewed: 0c8e7f5b-e510-4d24-893d-e3abb0bbb7b6`. Register: **WII-017** (this
gate) and the corrected **WII-013** entry; `SUMMARY_2026-08-14` is the read-aloud account.
