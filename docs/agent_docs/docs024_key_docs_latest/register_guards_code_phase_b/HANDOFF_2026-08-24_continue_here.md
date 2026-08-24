# HANDOFF 2026-08-24 — continue here

**Lane:** `register_guards_code_phase_b` (`bugs_open/288`, the class behind
`bugs_closed/225`). **Supersedes `HANDOFF_2026-08-16_continue_here.md`**, which is
accurate for everything up to 2026-08-17 and silent on everything since.

**State: four phases built and committed 2026-08-24. ALL GO, therefore ALL INERT until
the next chassis roll.** Phase 1 council-**APPROVED at round 1** (`67643b47`, both real
advisories acted on, not banked). Phases 3a/4 are at `041b3026`, **round 1 REVISE
answered, round 2 in flight** — read the verdict before doing anything else.

## Read this first: three claims in the OLD handoff and bug file were false

1. **"P11 refuses a malformed declaration where it is written"** — P11 had never once
   run on a tool fence. Its only production caller is the *experience-pattern* register.
   Fixed; the write door now validates.
2. **`parseCriteriaFacts` fail-open** returned no ids *and no issues* on a broken fence,
   which also disarmed the round-3 zero-rows warning. Fixed.
3. **§5.5's deferral cites `bugs_open/093`, which CLOSED 2026-08-17** — and the way it
   closed (a second SURFACE for the existing scanner, never a second scanner) is the
   template for the work it defers.

## What is built (all inert until the roll)

| phase | what it makes unrepresentable | commit |
|---|---|---|
| 1 | a `facts` declaration that silently means nothing — refused at the write door, and reported as a `fact_declaration_broken` doc_note | `995b5fbbe` |
| 2 | RFC_025 stage 2b: `artifact_check` unreachable for citation facts, and an address that dies on decomposition | `eecd99b0a` |
| 3a | a declared pair the sweep knows nothing about — the byte probe, **annotation only** | `abe14bd38`, `8c9a0cf05` |
| 4 | a tool whose code carries a registered figure with no binding and no suggestion on record | `bf8a9bd35`, `8c9a0cf05` |

## THE THREE RULES THAT ARE NOT NEGOTIABLE

1. **Probe SCRIPT TEXT, never the whole page.** The register's own `writer_line` puts
   the figure in the page's PROSE. Measured on the live stamp-duty page: `500,000` is in
   both script and prose, `500000` in the script only, `625000` (the expired cap) in
   neither. **A whole-page check would have certified bug 225's page daily for sixteen
   months.**
2. **Never tokenize a `string_agg` of `rendered_html`.** Partial fragments; one unbalanced
   `<script>` leaks `inScript` into the next component's prose. Same certification, one
   level down. Extract per fragment, combine the results.
3. **The distinctiveness floor (1000) is MEASURED, not chosen.** 32.75% / 3.79% / 0.06% /
   0.03% / 0.00% false positives at 1–5+ digits over 161 real tool pages with invented
   probes. Do not lower it by argument; re-derive it (RUNBOOK).

## WHAT IS OWED, in order

1. **Read the round-2 verdict on `041b3026`.** The round-1 gating objection was right and
   found a real defect. Expect `debug_historian` to gate again if anything is still wrong.
2. **The roll, then prove at the binary** — recipe in the RUNBOOK, with controls. Do not
   verify at the tag.
3. **Retype ONE real fact** so RFC_025 finally has a consumer (it still has **0 of 294**).
   `sdlt-ftb-relief-cap` with `artifact_check{subject_key:"stamp-duty", pattern:<context-
   bearing, never bare digits>}`, as a CONTRIB to the mcalc lane. Then induce: make the
   pattern unmatchable inside a `trap … EXIT` window. **A dry run that reports nothing
   after an induced change is the failure.**
4. **Read the first full sweep's probe distribution.** That measurement is the entire
   precondition for Phase 3b (letting presence settle an item), and 3b is its own round.
5. **`mortgages-stamp-duty` on loanandmortgagecalculator declares nothing** and is the
   estate's second SDLT calculator. The suggester will now propose its bindings; their own
   08-17 triage says they are unblocked. CONTRIB filed.

## Landmines this session added to the lane's own list

- **A test that calls a guard directly cannot see whether anything CALLS it.** Three
  times this session, all green under deletion of the call site. For every guard, **delete
  the CALL, not just the body.**
- **A fixture built to DEMONSTRATE a rule will not TEST it.** The headline test for the
  script-only rule passed under the whole-page mutation. Assert the premise — that the
  wrong version really is wrong — before asserting the fix.
- **A trailing comma is a LIST SEPARATOR.** Excluding it from the boundary guard hid
  `{ upTo: 1500000, rate: 0.10 }` — the real band table, on the page the rule was for.
  And **Go's RE2 has no lookaround**, so this is hand-rolled byte checks.
- **An objection can be sound with the wrong reason attached.** Two seats' same-named-tool
  scenario cannot happen (134 PLANs, 134 distinct keys) — but one global PLAN resolves on
  many sites (6 do), so their fix was right anyway. Check the premise, keep the fix.

## Still NOT done, and not this lane's to close alone

Piece 4 (the oracle — the only thing answering *is the figure RIGHT*; a tool and register
wrong together still agree silently). The prose half of the class (`bugs_open/288` §5.5,
now with a corrected citation). `bugs_open/033` — the architecture seat recorded that this
lane has now built the **second** bespoke doc_notes surface to route around the dead
review queue, and asked that it not be silently absorbed by a third.
