# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-08c)

**Written 2026-08-08 late night. Supersedes `HANDOFF_2026-08-08b_continue_here.md`**
(read it second — its §0 owner rulings, §1 comparator post-mortem and §4
landmines all stand; its §3 steps 1–3 are DONE and the results are below).
Site: UNLOCKED. Queue state unchanged from 08-08b §2 (re-verified this session).

## 0. Owner rulings in force (08-08 night, both recorded verbatim in NOTES)

1. Correctness beats fidelity — never copy a wrong original; improve past it.
2. The checker proves results don't differ on identical inputs + catches wrong
   results. 3. Site stays unlocked. 4. Everything runs from the framework.
5. **Both-right-differently → supply BOTH calculators** — primary page keeps one
   model, the alternative ships as its own well-flagged, signposted page
   explaining when each applies.

## 1. DONE this session — the replay comparator, and the verdict on all 9

`acceptance/compare_rebuilt.py` REWRITTEN: replays the golden's recorded fill
plan (literal values by id; select by VALUE; checkables set not toggled; press
via heuristic since no original button has an id). Verdicts: VERIFIED /
DIVERGED / NEEDS-JUDGEMENT / DOMAIN-DIFF (validation refused an input — not
arithmetic) / REPLAY-FAIL (input didn't take — tool NOT judged). Rounding-equal
= within half a unit of the coarser side's precision. 200-char-truncated ids
are eyeball-only. Harness validated against the 08-08 hand-drive before use
(repayment: £1,389.58 vs £1,390 reproduced). Report of record:
`acceptance/COMPARE_2026-08-08_replay_absolute_inputs.txt`. Full per-tool
evidence: NOTES 08-08 late-night entry.

| tool | verdict on identical inputs |
|---|---|
| simple | VERIFIED (4/4 vectors; accepts fractional years) |
| repayment | VERIFIED where both answer; rebuild refuses fractional terms (`Number.isInteger`) — stricter domain, defensible |
| overpayment | VERIFIED in substance ("0 years" vs "6 months" — units, rebuild finer) |
| investor | VERIFIED in substance; golden's LTV 0% = capture never pressed the SECOND button (see landmine below). Rebuilt LTV 75.0%/35.9% arithmetically exact |
| equity-release | debt roll-up: golden £0s = same unpressed-button artefact; rebuilt penny-exact (100000×1.065^y). Real diff: max-LTV at 65, step table 31% vs linear ramp 30% — both stated approximations, improvement-loop call |
| stamp-duty | **ORIGINAL WRONG, rebuild RIGHT** at FTB £500–625k: original grants relief there (a no-regime hybrid: £300k nil + £625k cap), quotes £14,750 at £595k where post-Apr-2025 rules say standard rates £19,750. Other 3 vectors match exactly (hand-mapped `ftb`→`firstTime`). REPLAY-FAIL until option VALUES are aligned — the id contract missed them |
| rate-forecaster | BOTH RIGHT, different models: original = 3-phase rate path (yrs 1–2 r1, 3–5 r2 on remaining balance, 6+ r3 — reproduced to the penny); rebuild = each rate full-term from day one (textbook-exact). §0.5 candidate — the original's model arguably IS the "forecaster" |
| fee-analyser | BOTH RIGHT, different definitions: original = total repayments + fees over deal (£26,841.44 exact); rebuild = simple interest + fees (£17,759 exact). §0.5 candidate |
| bridging-loan | BOTH RIGHT, different models (known): retained-interest gross-up vs compound variant. §0.5 candidate |

**Scoreboard: zero rebuilt tools compute a wrong number. One ORIGINAL does
(stamp-duty). Three legit model/definition splits → §0.5.** New fleet landmine
filed + synced: "toolgolden presses ONE button per page" (multi-calculator
pages golden the unpressed section's placeholders).

## 2. Next actions, in order

1. **Route the three §0.5 tools to the improvement loop** with both models
   stated (table above + NOTES formulas). Primary-page model proposals:
   bridging → retained-interest (the structure lenders quote); rate-forecaster
   → the 3-phase path model (it is the better "forecast"); fee-analyser → keep
   interest+fees as "true cost" but SHOW the total-outlay figure too. Each
   alternative model = separate signposted page (owner §0.5). Everything
   framework-built — the `add_tool` item shape already exists on this site's
   queue (`add_tool_novel_mortgagecalculator.co.uk`, deferred, is the worked
   example to copy).
2. **Stamp-duty option-value alignment**: fold option VALUES into the id
   contract (spec `interactive_features` should name them like ids) and re-file
   the recreation so `#buyerType` carries `ftb`/`next`/`additional`. Until then
   automated replay REPLAY-FAILs and any emitted criteria (select-by-value)
   cannot drive the page either — fences are blocked on this too.
3. **Report the stamp-duty ORIGINAL defect** to the owner as a finding (his
   site under-quotes SDLT by £5,000 at £595k FTB; the README 08-08 late-night
   entry already states it in prose) — no fix needed on our side, the rebuild
   is correct; do NOT "fix" the original (owner controls it).
4. **Finish id-alignment on the three stragglers** (unchanged from 08-08b §3.4:
   affordability + fact-finder + portfolio — back up + delete `page_components`,
   re-file recreation items copying specs from items `1b74c60c`/`d935c3a4`;
   portfolio carries the `bugs_open/222` comment-style clause).
5. **Wire the replay comparator into the experience/tool loops** as their
   acceptance check (ruling §0.2/§0.4). Note `--emit-criteria` is still gated on
   oracle-checking first (see fleet landmine "a behavioural golden certifies a
   calculator that has ALWAYS been wrong").
6. Fences after ids are stable 12/12 (and after step 2 for stamp-duty).
7. Owner housekeeping list unchanged (08-08b §3.7).

## 3. Landmines live on this work

- **All of 08-08b §4** (comparator cross-page trap; unlocked-site queue checks;
  ~1-day `orchestration_states` purge; full `/index.html` URLs; commit per task).
- **NEW: multi-calculator pages** — a golden id inert at `0%`/`£0` across every
  vector while its inputs were driven = unpressed-section suspect, not a
  measurement (LANDMINES, added 08-08).
- The replay comparator inherits toolgolden's other traps: defaults-
  neighbourhood vectors (boundary defects invisible — the stamp-duty catch came
  from the asym vector straying over £500k, luck not design; an ORACLE from the
  published rules is the real fix, see loanandmortgagecalculator lane), and
  golden-certifies-the-always-wrong.

## 4. Files of record

This dir: `NOTES` (08-08 late-night = per-tool evidence + formulas) ·
`README_where_we_are` (owner log through the stamp-duty finding) ·
`SUMMARY_2026-08-08_arithmetic_verified_on_identical_inputs.md` (the milestone
read-out) · `RUNBOOK` §10 · `acceptance/` (`compare_rebuilt.py` replay mode;
`COMPARE_2026-08-08_replay_absolute_inputs.txt`; the misleading
`COMPARE_2026-08-08_id_aligned_9of12.txt` kept as evidence).
Bugs: `bugs_open/218` (B design call open) · `bugs_open/222` (fabrication
negation-blindness) · `bugs_open/178`. Fleet docs this session: `LANDMINES.md`
(+1, synced).
