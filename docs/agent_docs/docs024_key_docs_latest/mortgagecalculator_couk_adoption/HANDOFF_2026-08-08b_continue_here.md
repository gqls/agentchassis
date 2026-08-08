# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-08b)

**Written 2026-08-08 ~19:30 UTC. Supersedes `HANDOFF_2026-08-08_continue_here.md`**
(read it second — its state table, §3 landmines and both addenda stand, EXCEPT
addendum 2's "6 tools diverge on FORMULA" claim, which is REFUTED below).
Chassis at writing: v1.0.1264. **Site: UNLOCKED (owner ruling, below).**

## 0. OWNER RULING 2026-08-08 (night) — read this before anything else

Given verbatim-in-intent after the (wrong) formula-divergence report:

1. **Do NOT copy an original's calculation method if it is wrong — improve every
   tool to the best it can be.** The experience/tool loops own that improvement.
2. **The arithmetic checker's job is to prove results don't differ on identical
   inputs, and to catch wrong results** — not to enforce fidelity to the original.
3. **The site is not to stay locked** — especially not to preserve tools that
   report wrong results. It is UNLOCKED as of this session; the queue is live.
4. **All content and tools are controlled from the framework** so they can be
   improved. The byte-frozen "originals are the contract" posture ENDS for tools.
   The golden remains the drive-plan source and a regression reference — not an
   arithmetic authority.
5. **ADDENDUM (same night, later): both-right → supply BOTH.** "If the two
   calculators are right in different ways then we can explain it and supply
   both calculators for each task — maybe as a separate, but well flagged and
   signposted page (for those that are interested or need one or the other)."
   Divergence routing is therefore three-way: rebuild wrong → fix · original
   wrong → improve past it · **both defensible → primary tool keeps one model,
   the alternative ships as its own signposted page explaining when each
   applies** (bridging-loan is the standing candidate). Framework-built, per §0.4.

## 1. Why the compare said "it's all different" — the actual answer

**Zero rebuilt tools have a demonstrated arithmetic defect.** The comparator
(`toolgolden.py` via `compare_rebuilt.py`) derives every driven value by scaling
the page's OWN markup `value` attributes (`DRIVE_JS`:
`parseFloat(e.getAttribute('value')) × factor`; a fixed **1000** into any numeric
field with NO value attribute; `step>=1` rounds the driven value). It goldens a
page against ITSELF — so the compare fed the original ITS defaults and the rebuild
ITS OWN different defaults, and reported two correct answers to two different
questions as divergence. Proven per tool:

| tool | what the diff really was |
|---|---|
| repayment | rebuild ships defaults 200000/5.0/25 (original 250000/4.5/25). Hand-driven with identical inputs: rebuild **£1,389.58** vs original **£1,390** — same answer, original rounds to pounds |
| rate-forecaster | same class — rebuilt default amount 200000; its £1,111.66 is exactly 0.8× the correct £1,389.58 |
| overpayment, equity-release | different shipped defaults → both-correct different answers |
| investor | rebuild ships NO value attributes → harness drove **1000 into rent AND price** → the "1200% yield (£12,000/yr)" |
| fee-analyser | no value attributes → 1000 driven into a 2-year deal-term field → validation refused → '—' |
| simple | identical defaults; only `half` diverged because rebuilt `years` has `step="1"` → driven 12.5 rounds to 13 → £739.94 is CORRECT for 13y (golden £765 correct for 12.5y) |
| stamp-duty | select options[1] BY INDEX lands on a different buyer type (rebuild reordered options). Both sides' SDLT arithmetic verified correct for what each selected. The historic "£0 after press" was the same select-order effect |
| **bridging-loan** | **the ONE genuine divergence**: identical defaults, different model. Original: retained-interest gross-up `gross = net/(1 − fee% − monthlyRate%×months)` (fee £4,494, interest £20,225 on net £200,000 — the structure bridging lenders actually quote). Rebuild: a compound-interest variant (£19,180.99). Which is RIGHT is a product-knowledge judgement — per ruling §0.1 it goes to the improvement loop, NOT copy-the-original |

Full account: `WRONG_CALLS.md` 2026-08-08 (differential-test entry + addendum) ·
NOTES 08-08 night correction · LANDMINES "toolgolden … cross-page compare".
**The landmine existed before the mistake** ("only ever drives NEIGHBOURHOODS OF
THE SHIPPED DEFAULTS") — grep LANDMINES for the symbol you are about to trust.

## 2. State right now

| thing | state |
|---|---|
| Site lock | **UNLOCKED** (ruling §0.3). The queue is live — open items on the site: `needs_rerender` (detected, 08-05 "rerender after template fix"), several `blocked`/`needs_human_review` rows incl. 2 new `save_refused_incomplete` (below) and 2 portfolio fabrication reviews. Subset-batch discipline (§10c backstop) still applies when you need an exclusive window |
| 12 tools | ALL live. 9 with golden-aligned ids (id contract proven — the spec's `interactive_features` renders as OVERRIDE requirements in the recreate prompt). 3 kept earlier versions: affordability (shrink guard 14,907→6,495 hero), fact-finder (prune guard 1-of-4 sections), portfolio (fabrication gate false positive → `bugs_open/222`) |
| Unsaved payloads | all three preserved and id-complete (5/5, 19/19, 15/15) in session scratchpad `payload_{1b74c60c,d935c3a4,07b7eca3}.html` — scratchpad dies with the session; the id contract in the specs regenerates them, so treat as convenience, not the artefact |
| Originals | all still serve; §10f sweep clean (robots.txt only) as of 08-08 ~17:10 UTC |
| Validator fix (218 A) | LIVE in v1.0.1264, proven on its motivating cases |
| 218 defect B | mechanism corrected in the bug file (routing NOT dead; discard = success-only field + skip-as-success). Design call (save-anyway vs cannot-complete) still owed — not this lane's |
| `bugs_open/222` | NEW: fabrication declaration tier convicts negations ("no fabricated data"); the recreate prompt's own vocabulary manufactures the echo. Fix candidates in the file |
| Arithmetic verification | **0 of 12 proven — but now for the right reason**: the comparator has never yet fed both sides identical inputs. No tool is proven wrong either (one bridging model question open) |

## 3. Next actions, in order

1. **Make the comparator replay ABSOLUTE inputs** (the checker fix, ruling §0.2).
   The golden already records the fill plan per control (`sel`/`action`/`value` —
   recorded for --emit-criteria replay). Change `compare_rebuilt.py` (or add a
   replay mode to `toolgolden.Runner`) to drive the REBUILT page with the
   GOLDEN's recorded literal values by id (select by VALUE, not index; dispatch
   input+change, which the hand-drive proved sufficient). Judge `after_press`
   only. Hand-drive one case first and confirm the harness gets the same answer
   the hand-drive gets — the harness is only trustworthy once it reproduces a
   known-good drive.
2. **Run the replay compare across the 9 id-aligned tools.** Where both sides
   AGREE on identical inputs → arithmetic verified (expect most to pass — spot
   checks already match). Where they DIFFER → compute independently, judge which
   is RIGHT (ruling §0.1: correctness, not fidelity), and route the wrong side:
   rebuild wrong → fix item with the correct formula stated; ORIGINAL wrong →
   record it and improve the rebuild past it — do not regress the rebuild to
   match.
3. **Bridging-loan model judgement** — the one known genuine divergence. Per
   ruling §0.5: both models are defensible answers to different questions, so
   the plan is BOTH — primary page keeps the retained-interest gross-up (the
   structure lenders quote), and the compound-interest variant becomes a
   separate, well-signposted page explaining when each applies. Route to the
   experience/tool improvement loop with both models stated (§1 table);
   framework-built, never hand-rolled.
4. **Finish id-alignment on the three stragglers**: affordability + fact-finder
   → back up + DELETE their `page_components` (§10e precedent — the guards then
   have nothing to compare against), re-file the recreation items (same spec
   shape as the 08-08 batch; copy from items `1b74c60c`/`d935c3a4`); portfolio →
   same, plus the 222 comment-style workaround clause (in the bug file) until
   222 is fixed. Site is unlocked — plain re-file dispatches; use a §10c
   backstop only if you need to keep other queue items out of the window.
5. **Wire tool improvement into the experience/tool loops** (ruling §0.4) — the
   lane hand-drove two batches; the loops should own iteration from here. The
   replay comparator from step 1 becomes their acceptance check ("results don't
   differ on identical inputs, and the answers are independently right").
6. Fences (unchanged, after ids are stable on 12/12) · owner README kept current.
7. Housekeeping calls for the owner, listed not urgent: two portfolio
   fabrication review items (`aca92097` 08-05, `3d11e960` 08-08 — the 08-08 one
   is a proven false positive per 222); 2 `save_refused_incomplete` items
   (superseded if step 4 runs); `pages.build_status` cosmetically stale.

## 4. Landmines active on this exact work

- **The comparator trap that caused tonight's wrong call** — see §1. Never read
  a toolgolden cross-page diff as arithmetic. Absurd magnitudes (1200%) = the
  1000-into-empty-fields branch.
- **The site is UNLOCKED** — automated dispatch can reach it. What protected
  unbuilt pages before the lock was zero `page_components`, and that protection
  is gone for all 12 tools (they have components). A stray rerender is
  harmless-by-design (regenerates from content_data) but a `needs_content_page`
  producer is not — check the queue before assuming nothing is in flight.
- `orchestration_states` purges completed runs within ~a day — read evidence
  same-day (the 08-05 portfolio signals died exactly this way).
- Wire spot-checks: FULL `/index.html` form; bare directory URLs 404 except root.
- `complete` ≠ artefact, and a `complete_error` run may still have deployed
  (spawn→call handshake race at `deploy_page` — overpayment, 08-08 afternoon).
- Commit per task with pathspec; `make build-*` builds committed HEAD.

## 5. Files of record

This dir: `NOTES` (08-08 evening + night entries and correction = the full
story) · `README_where_we_are` (owner log — 08-08 night entry answers the
owner's questions in prose) · `RUNBOOK` §10 (rebuild chain; §10f sweep + new
bare-URL gotcha; §10c/§10g backstop rules) · `acceptance/` (goldens = drive
plans + regression reference per ruling §0.4; `COMPARE_2026-08-08_id_aligned_9of12.txt`
= the misleading raw report, kept as evidence; `EVIDENCE_2026-08-08_rerun_3tools_…jsonl`).
Bugs: `bugs_open/218` (A live+proven; B corrected, design call open) ·
`bugs_open/222` (fabrication negation-blindness, NEW) · `bugs_open/178`
(shrink guard, cited by the refusal). Fleet docs updated this session:
`WRONG_CALLS.md` (two entries + addendum), `LANDMINES.md` (toolgolden
cross-page + bare-URL… synced to doc_notes), 016b §9 (fabrication echo
pattern).
