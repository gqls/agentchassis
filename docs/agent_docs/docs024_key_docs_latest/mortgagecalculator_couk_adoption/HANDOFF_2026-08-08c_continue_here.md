# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-08c)

> **UPDATE 2026-08-09 afternoon, read before acting on §2.**
> - **All five items in §2.1 are `complete`** (built 11:08–11:19Z, verified by row
>   id). **The comparator re-run they were waiting on is now OWED** — nothing yet
>   confirms the rebuilds landed the agreed models rather than reporting success.
> - **§2.1b(b) — tools-vs-facts acceptance — is now DESIGNED, not built:**
>   `PLAN_2026-08-09_facts_into_tool_acceptance.md`. Four pieces, phased; the
>   first needs no Go at all. Read its §5 landmines before writing any consumer
>   of `evidence_base`.
> - **The first `evidence-freshness` sweep over our facts has NOT run** — the
>   task last completed 08:58Z, *before* the ~12:30 seed. Due ~08:58Z 08-10.
>   §2.1b's 24h check is still owed and the day-one gotcha still applies.
> - **New structural finding:** this site's twelve recreated tools have no
>   `doc_plans` PLAN, so no criteria, so no Tier 2 and no Tier 4 — zero
>   acceptance runs have ever happened here. That is a prerequisite for §2.6
>   (fences) and for the plan's Phase B.

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

1. **DONE 08-09 — the three §0.5 tools are ROUTED (owner approved the model
   proposals).** Five items filed + armed, FIFO order add_tool first:
   `0dc7a786` add_tool tool-bridging-compound · `0c529013` add_tool
   tool-rate-scenarios · `c9f810a3` recreation tool-bridging-loan
   (retained-interest primary) · `df5c5935` recreation tool-rate-forecaster
   (3-phase path primary) · `ba68c674` recreation tool-fee-analyser (both
   figures, new ids tcTerm/tcOutlay). Each spec embeds the formula AND a worked
   check the implementation must reproduce (NOTES 08-09). **Follow up: after
   they build, re-run the replay comparator** — bridging + forecaster should
   then MATCH the golden; fee-analyser tcTotal stays a recorded definitional
   split (amortized interest, so ~17,385 not ~17,759 on golden defaults).
1b. **Legislation watch (owner ask 08-09): the mechanism EXISTED — the site is
   now ENROLLED.** Daily `evidence-freshness` sweep re-verifies citation facts
   by re-fetch + verbatim-quote match (V5). Seeded `site_specs` aspect
   `evidence_base` (pinned): 4 SDLT facts citing GOV.UK, incl. the £500k FTB
   relief cap the original violates. **CHECK THE FIRST SWEEP (~24h): a
   `citation_lost` on day one = quote-extraction mismatch, NOT moved
   legislation** — fix the quote. Then two design pieces, in order: (a) the
   published "current SDLT rates" page (owner wants it; numbers from
   writer_lines; blocked on confirming how a NEW guide page row gets created
   on this site); (b) tools-vs-facts acceptance — an oracle-from-the-register
   check so a tool encoding a stale threshold FAILS acceptance (platform seam
   → council gate when built; loanandmortgagecalculator's oracle.py is the
   pattern).
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
