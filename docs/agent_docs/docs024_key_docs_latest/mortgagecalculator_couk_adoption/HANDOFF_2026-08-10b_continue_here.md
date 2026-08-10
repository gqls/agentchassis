# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-10b)

> **SUPERSEDED the same evening by `HANDOFF_2026-08-10c_continue_here.md` — start
> there.** §3 action 1 (the twelve tool PLANs) is **DONE for the eight tools that
> can have one**, and all eight have since passed a real Tier-4 run. Four things
> in §3's framing of that action were wrong — see 10c §2. The two that would have
> cost the most: **the ladder's subject key strips `tool-`** (a PLAN filed under
> the page name is never read, with no error anywhere), and **three of the
> "twelve" are not ladder-eligible at all**. §3 action 1's parenthetical "zero
> acceptance runs have ever happened on this site" was already false when written
> — two completed 08-09 evening.

**Supersedes `HANDOFF_2026-08-10_continue_here.md`** (read it second: its §0 owner
rulings stand, its §1 verdict table stands except stamp-duty, and its §4 actions 1
and 4 are **DONE**). Site: UNLOCKED. The design this lane executes:
**`PLAN_2026-08-09_facts_into_tool_acceptance.md` — read it before writing code.**

## 0. Owner rulings in force (unchanged)

1. Correctness beats fidelity — never copy a wrong original; improve past it.
2. The checker proves results don't differ on identical inputs, and catches wrong
   results. 3. Site stays unlocked. 4. Everything runs from the framework.
5. Both-right-differently → supply BOTH, the alternative on its own signposted page.

## 1. What landed on 08-10 evening

**(a) Migration 366 is PROVEN on a real rebuild — A3 is done for
`tool-recreation-handler`.** But **not by the test the last handoff prescribed**,
and the reason matters more than the result.

> **The prescribed test could not have failed.** "Read the generated JS for
> £500,000 rather than £625,000" was owed — but the **pre-366** build of 08-08
> already carried `FTB_RELIEF_LIMIT = 500000`. The model knows SDLT. That test
> measures its memory, not our register, and returns green either way.

The discriminator is **attribution**: does the register's own wording reach the
artefact? Measured over the two builds of the same tool, same agent, spec
identical but for one id-contract clause — each of the four composed
`writer_line`s appears **0 times before, exactly once after**, as a comment beside
the constant it licenses. Positive control (`Stamp Duty`) present in both, so the
search could fire on either file. **Strip `//` markers and collapse whitespace
before matching** — the generator wraps a long line across two comment lines, and
a verbatim grep reports 0 for BOTH, which reads as "the change did nothing".

**(b) THE FINDING OF THE DAY, and it was not the one being looked for: the
register is load-bearing in BOTH directions.** The same rebuild **dropped the
£40,000 additional-property surcharge floor** — `SURCHARGE_FLOOR = 40000`, twice
in the old build, zero times in the new. True law, correctly implemented before,
absent from a register that held four SDLT facts. 366's own text says *"Do NOT
state a rule that is not in the register"*, and it obeyed. Nothing failed; the
tool was silently wrong below £40,000. **A partial register is not a neutral one,
and every register is partial.** Filed fleet-wide to `LANDMINES.md` with the
prospective check; registered as CLM-021 in the concept register.

**(c) A1 DONE — 4 facts → 13**, one per band edge and per rate, `value` scalar
throughout, each with its own verbatim GOV.UK quote — including the £40,000 floor,
registered precisely because (b) had just shown what omitting it costs. Retired
`sdlt-standard-bands` and `sdlt-ftb-nil-rate` (checked: zero references in
`doc_plans`, `site_work_items`, `page_components`). `pinned` carried forward.
**The SQL and its generators are now IN THE REPO** — `evidence/` — closing the gap
the PLAN recorded (the first four facts existed only in the database).

**(d) The induced proof, run forward — this is the strongest evidence the lane
has.** Re-filed the recreation with a spec **byte-identical** to the previous one
(diffed as parsed JSON before firing), so the register was the only changed input.
`const ADDITIONAL_THRESHOLD = 40000` came back with the register's new
writer_line as its comment, **and is read at the branch**, not merely declared.
All ten granular writer_lines appear beside their constants. Arithmetic unchanged
and correct. `[n=1 on a non-deterministic generator — this evidences the
mechanism; it does not prove it.]`

**(e) Stamp-duty option VALUES are aligned — action 4 done, REPLAY-FAIL cleared.**
`#buyerType` now carries `next`/`ftb`/`additional` in the original's order with
`next` selected on load. The comparator judges the tool at last and returns
**DIVERGED with the ORIGINAL wrong**: £595k FTB → golden `£14,750`, rebuilt
`£19,750`; the £350k vector agrees at `£2,500`. Reports:
`acceptance/COMPARE_2026-08-10b_…txt` and `…-10c_stamp_duty_register_driven.txt`.

**(f) Fleet weather, measured not assumed.** Every LLM call in the fleet failed
**14:51:45Z–17:02:12Z** on the account usage cap whose message promises access
back on 2026-09-01. **It cleared the same day** — 70 successful calls in the 18:00
hour across 3 agent types. The `LANDMINES.md` entry has been corrected: as written
it told every lane its council obligation was unsatisfiable for three weeks.

## 2. Where the plan stands

- **Phase A — data and config.** A1 **DONE** · A2 **DONE** · A3 **DONE for
  `tool-recreation-handler`**; `tool-generator` and `tool-improver` still TODO
  (they need a `read_site_spec` step added first — the action exists and is
  registered, so still config, not Go) · A4 **TODO** · A5 **DONE**.
- **Phase B** — Pieces 2 and 3 (the `facts` declaration in the criteria doc; the
  drift fan-out). Go + validator + one seed. Council gate.
- **Phase C** — Piece 4, the oracle. **RFC first**, then council. **A1 has removed
  its blocker**: the band table is now readable off the register as scalars.

## 3. Next actions, in order

1. **A4 — create tool PLANs for the twelve recreated tools.** This is now the
   single biggest blocker: no `doc_plans` row ⇒ no criteria ⇒ no Tier 2, no Tier 4,
   and **zero acceptance runs have ever happened on this site**. It gates fences
   and all of Phase B. Stamp-duty is ready for one first (ids AND option values
   are now pinned).
2. **Finish A3 — `tool-generator` and `tool-improver`.** Add the `read_site_spec`
   step plus the prompt clause, mirroring 366. Config only. **Before you do, read
   §1(b)**: these agents will also delete what the register omits.
3. **Sweep the other tools for unregistered constants BEFORE any of them is
   rebuilt.** §1(b) generalises past stamp-duty — every calculator on this site
   encodes rules, and the register only carries SDLT. The check is in LANDMINES:
   enumerate the constants, ask which the register carries, register the gaps
   first. Do not file a recreation for a fact-bearing tool until then.
4. **Report the stamp-duty ORIGINAL defect to the owner** — under-quotes by £5,000
   at £595k FTB, now confirmed by the automated comparator rather than by hand.
   `CONTRIB_2026-08-09_original_stamp_duty_fixed_by_owner_decision.md` suggests
   this may be in hand; confirm before re-raising.
5. **Id-alignment on the three stragglers** (affordability, fact-finder, portfolio).
6. Phase B, then the Phase C RFC.

## 4. Landmines live on this work

- **All of 08-08b §4, 08-08c §3 and 08-10 §5** still stand — including the
  shared-id comparator trap (a rebuild specified to ADD outputs reads DIVERGED by
  construction) and the typed-struct round-trip that deletes every citation.
- **NEW — absence from the register is an instruction too** (§1b). Fleet-wide in
  `LANDMINES.md`; CLM-021 in the concept register.
- **NEW — a `complete` work item can be an API-limit death.** One read `complete`
  **52 seconds** after filing with `result.response` holding an early step's
  output; the orchestration was at `complete_error`. A real recreation here takes
  ~5 minutes, so a sub-minute `complete` is the tell. Judge at `page_components`,
  never at the item. RUNBOOK §12.
- **NEW — a verbatim grep against generated source is a claim about the
  generator's line width.** Strip comment markers before matching, and always
  carry a positive control (§1a).
- **`doc_plans` has no `site_id`** — the criteria document is fleet-global. A fact
  id is per-site, so it must resolve against the site of the PAGE being driven.
- **A neighbouring lane is active in the same machinery.** Re-measure `doc_plans`;
  do not carry a count from any commit message, including this handoff's.

## 5. Files of record

This dir: `PLAN_2026-08-09_facts_into_tool_acceptance.md` (**the design**) ·
`NOTES` (08-10 evening = today's measurements and the misstep) ·
`README_where_we_are` (owner's log) · `RUNBOOK` §11 (legislation watch), **§12
(filing a recreation item, and why `complete` is not success), §13 (lifting
quotes with the real Go extractor)** · `evidence/` (**the register's SQL and its
generators — reproducible now**) · `acceptance/COMPARE_2026-08-10b…`, `…10c…`.
Fleet: `LANDMINES.md` (+1 new, +1 corrected), `WRONG_CALLS.md` (+1),
concept register **CLM-021** (+ index row).
Migrations: `sql_for_agents/366_*` (+ ROLLBACK).
Bugs: `bugs_open/218`, `bugs_open/222`, `bugs_open/225`, `bugs_open/178`.
