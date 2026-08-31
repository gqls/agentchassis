# agritec.uk — continue here (updated 2026-08-31, afternoon)

Cold start: read this, then `SUBJECT_LEDGER.md`, then `NOTES_agritec_uk.md` from the bottom.
Mechanism maps and traps: `HANDOFF_2026-08-25_continue_here.md` §5–§6 (still valid). The
2026-08-30 handoff's "next: build the five calculators" is SUPERSEDED — the framework built
them, ungated, and that changed the work (below).

## 1. State — measured 2026-08-31

- **Phase 1 core is live and healthy**: 23 pages serve 200, light palette incl. all imagery,
  SFI stacker correct + fenced + reconciled (24/24), cap-claim garble fixed, GTM back,
  guides hub 6/6. The Tier-4 acceptance verdict (first ever): **8/9, the mobile-fit failure
  and a real AHW2/CAHL2 arithmetic edge both found AND fixed by the loop same-day (08-26).**
- **Zero `needs_human_review` items open.** All ruled with evidence in `result`
  (`SEED_2026-08-26b`, `SEED_2026-08-31_close_vision_and_claims.sql`).
- **The tools hub is at `/tools.html` now** — nav/sitemap/internal links all agree; the old
  `/tools/index.html` 404s and nothing internal references it.
- **The unresolved churn (84 `undeployed_asset`, 14 `deactivated_component`, 24
  `page_rerender`…) is a broken METER cycling daily, not damage** — the named artefacts serve
  200 (probed 08-30 + 08-31). Do not "fix" the site for it.

## 2. ⚠ THE OPEN PROBLEM: five tools shipped WITHOUT the evidence discipline

Between 08-26 15:24 and 08-27 17:32 the `evaluate_tools → add_tool` route built and deployed
**5 calculators + 5 companion guides** (T2–T5 + a gas-unit converter). Measured 08-31:
register untouched (104 facts), **0 external source links on all five**, no fences.
`blue-carbon-estimator` encodes unsourced empirical constants in JS (species dry-matter %,
carbon %, sink/burial fractions) — the claims scanner cannot see JS. Full per-tool detail:
ledger T2–T5 rows + NOTES 08-31. **T1 (vertical energy) is a false green** — its `add_tool`
row completed having persisted nothing; T1 is still unbuilt, and should be built
evidence-first when its evidence completes.

**The mechanism question is ROUTED, not ours**: CONTRIB filed into the loanzy lane
(`loanzy_uk_example_site/CONTRIB_2026-08-31_agritec_add_tool_executed_ungated.md`) — they own
the growth-refusal decision on the owner's desk (RFC_056 FOLLOW-UP #1); agritec is their
fourth worked case. Session `loanzy.uk` messaged 08-31.

**The site-side repair IS ours** (owner rule 08-21: no unsourced figure anywhere), in order:

1. **Evidence run #1 is DONE and REVIEWED** (08-31, NOTES tail): register 104 → **111** —
   kelp carbon 30% of DW (NASEM ×2) + the sequestration-fraction set (~11% of NPP reaches
   long-term storage; 88% of that via deep sea; ~70% via recalcitrant-DOC export). Clean on
   four of §9's five checks; the DRY-MATTER half came back silently empty, so run #2
   (dry-matter fraction) was dispatched. ~~run #2's FIRST dispatch silently dropped (the
   kcat trap)~~ **CORRECTED 2026-08-31, same session: it did NOT drop — both dispatches
   landed and completed (rows 13:01:18 and 13:52:45). My absence check filtered
   `created_at > 13:05`, which excluded the 13:01 row, so the query could only ever say
   "missing"; the refire was a DUPLICATE run, the exact cost the runbook says a retry
   incurs. Harmless here only because they ran sequentially (no lost update; register
   111 → 116, no duplicated facts). Logged in WRONG_CALLS.** Run #2 is DONE and REVIEWED
   (NOTES tail): dry matter covered — 10% DW/FW reference factor, measured 6.3–17.4%
   across systems, moisture 77.5–89.8% (Frontiers + PMC, peer-reviewed). ⚠ CIT-7691202a's
   quote is a raw TABLE ROW (§9.2) — corroborated by three sentence-quoted siblings; cite
   those first. **The register now covers every constant class the blue-carbon tool
   encodes; the BINDING step is next** (§2 item 2). ⚠ ONE evidence dispatch at a time
   (§10) — and before concluding a dispatch dropped, prove your query's time window
   actually contains the dispatch time.
2. After reading it: bind the blue-carbon tool's constants to the registered facts — cited
   visible copy (`content_rewrite` route works, proven on the cap fix) + `artifact_check`
   fences per the SFI pattern (`SEED_2026-08-25b` is the worked example).
3. Then the same for the others: hydroponic figures, the BSF help-copy trial ranges
   (re-source or reword to uncommitted language), VPD crop-band targets if any. The
   gas-unit converter is likely pure physics — verify, then say so in the ledger and move on.
4. T1 last, evidence-first (its remaining needs: LED efficacy, grid carbon intensity — ONE
   question per run, under 200 chars; the trigger script enforces the ceiling).

## 3. Everything else, in order

- Phase 2 (IoT cluster) after Phase 1 tools are sourced and verified.
- Later: news, editorial, directory (needs a new `directory_entities` kind).
- Standing cautions: `HANDOFF_2026-08-30` §5 (all still true — two discovery carriers,
  benign CTAs, 397's chrome item, the meter-vs-artefact undeployed_asset check).

## 4. Owner decisions on record

D1–D8 (`PLAN_2026-08-21` §1) · cite-every-figure with a visible link (08-24) · light palette
default (08-25) · no cannabis · **no unsourced figure anywhere (08-21) — the rule §2 enforces**
· every site through the framework (08-04).
