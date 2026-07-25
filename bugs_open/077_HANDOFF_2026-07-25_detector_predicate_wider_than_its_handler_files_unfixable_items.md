# HANDOFF — the `hardcoded_section_colors` detector files items its own handler cannot fix

**Filed:** 2026-07-25, by the thread closing `bugs_open/021`. Split out of 021's
INSTANCE 2 rather than left inside a closing file.
**Severity:** LOW. Not an outage, no data loss, no credit spend. It is a
permanent, bounded population of work items that misdescribe reality.
**Status:** OPEN. Needs a design call from whoever owns the check, not a patch.

---

## The mechanism

The **detector** and the **handler** for `hardcoded_section_colors` implement two
different predicates, and the detector's is strictly wider:

| | predicate |
|---|---|
| detector (`countHardcodedColorComponents`, `check_hardcoded_section_colors.go:95`) | `rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'` **AND** `rendered_html LIKE '%<style%'`, `locked_at IS NULL` — ANY hex, 3/4/6/8-digit, light or dark, **anywhere in the component** including inline `style=""` attributes; the `<style>` test is on the component, not on the colour's location |
| handler (`ReplaceHardcodedColors`, same file, used by `fix_hardcoded_colors` / agent `color-variable-fixer`) | dark 6-digit only (`#[0-4][0-9a-fA-F]{5}`) on `background`/`background-color`, plus two-colour `linear-gradient(Ndeg, …)` — and **only inside `<style>…</style>` blocks** |

So the detector can file an item on a site where the handler provably has nothing
to do. The handler then runs, changes nothing, and the item is eventually parked.

## Live evidence (2026-07-25, all counts re-run today)

Detector population, with "inside the handler's remit" computed by running
`ReplaceHardcodedColors` verbatim over each component's `rendered_html`:

| site | detector matches | inside handler's remit |
|---|---|---|
| robot-hands.com | 3 | 3 |
| gamesdesign.co.uk | 4 | 1 |
| leopardessconsulting.co.uk | 4 | 1 |
| **finetuning.uk** | **8** | **0** |
| **gaswholesalers.com** | **6** | **0** |
| **ai-agent-orchestration.com** | **4** | **0** |
| **webdesign.co.uk** | **2** | **0** |
| **dartsonline.com** | **1** | **0** |

**On 5 of 8 sites the handler's remit is empty.** 32 components matched, 5 of them
fixable.

Work items of this type, live today: **13** (4 `complete`, 8 `unresolved`, 1
`detected`). The 8 `unresolved` were parked by the two-strike rule
(`insertWorkItem`, `load_work_item_actions.go:1041`) — whose label is *"handler
had 2 chances and the issue persists"*. On the zero-remit sites that label is
wrong about the cause: the handler did not fail, it was never able to succeed.
Oldest is 2026-04-08.

## What is NOT happening — measured, because the obvious guess is wrong

The 021 INSTANCE 2 note called this **churn** ("correct completions keep
re-detecting; the cycle repeats"). That is **not** what the data shows, and the
correction matters because it changes the severity:

- `idx_swi_dedup` is UNIQUE on `(site_id, item_key)` excluding only terminal
  statuses — and `detected` is **not** terminal. So one open item per site blocks
  any re-file. robot-hands.com has carried a `detected` item since 2026-07-17;
  a design-discovery sweep ran over that very site on 2026-07-24 20:46 (it filed
  `undeployed_asset` ×21, `needs_sprite_css` ×3, `needs_imagery` ×4 — so the
  agent and this check's list both ran) and filed **no** new
  `hardcoded_section_colors` item.
- `hardcoded_section_colors` appears **zero** times in 7 days of discovery output
  fleet-wide.

So the volume is bounded at roughly one open item per site, indefinitely. There is
no repeated dispatch and no repeated spend — and the handler is not LLM-driven
(3 workflow steps, no `execute_llm_prompt`), so a wasted run costs compute only.
**This is a correctness/legibility defect in the backlog, not a cost defect.**

## Why it is worth fixing anyway

1. **The backlog lies.** A human or an agent reading `site_work_items` sees 8
   unresolved "hardcoded colours" items and reasonably infers 8 sites with a
   colour problem a fixer keeps failing on. On 5 of them there is nothing the
   assigned fixer was ever going to change.
2. **`unresolved` is load-bearing elsewhere.** It is the two-strike parking state
   used fleet-wide to mean "needs investigation". Poisoning it with items that
   are *correctly* unfixable devalues the signal for every other type.
3. **The completion gate is now scoped to the handler and discovery is not.** As
   of `34adb171c` (live v1.0.1159) `VerifyHardcodedSectionColorsResolved` asks
   "would the fixer's own transform still change anything?" — deliberately the
   HANDLER's remit (see `016b` §9, *"A 'did the fix work?' check must assert the
   HANDLER's remit, not the DETECTOR's predicate"*). Detection still asks the
   wider question. The two ends of the same item now disagree by design, which is
   defensible as a stopgap and poor as a resting state.

## Candidate fixes (a design call — do not just pick one)

- **A. Narrow the detector to the handler's remit.** Count only what
  `ReplaceHardcodedColors` would change — ideally by *calling it*, as the verifier
  does, so the three ends cannot drift. Cheapest and makes item counts honest.
  **Cost:** the platform stops noticing light/3-digit/inline-attribute hardcoded
  colours entirely. If anyone wants those found, they lose their only detector.
- **B. Widen the handler.** Teach the fixer light hexes, 3/4/8-digit forms and
  inline `style=""` attributes. **Cost:** materially riskier — inline styles and
  light colours are where legitimate design intent lives, and a wrong rewrite is
  a visible site regression. This is the branch that wants a real review.
- **C. Split the item type.** `hardcoded_section_colors` (fixable, dispatch to
  `color-variable-fixer`) and a report-only finding for the rest. Honest, but adds
  a type to a registry that already carries 77.

**Before picking:** the 5 zero-remit sites are the test set. Whatever ships must
leave them with either zero items or items truthfully labelled as report-only.

## How to verify a fix

Re-run the remit table above (method in
`docs024_key_docs_latest/durable_write_guard/RUNBOOK_durable_write_guard.md`,
§"Know the expected verdict"): dump the detector population with `row_to_json`
and run the shipped transform over it. A correct fix makes "detector matches" and
"inside remit" agree, or makes the difference explicit in the item.

## References

- `bugs_open/021` (now `/bugs_closed/`) — INSTANCE 2, where this was found and
  where the verifier's remit-scoping is argued in full.
- `016b` §9 — *"A 'did the fix work?' check must assert the HANDLER's remit, not
  the DETECTOR's predicate"*, the transferable pattern.
- `WRONG_CALLS.md` 2026-07-20 — the held `page_rerender` verifier, the near-miss
  that taught the distinction.
- `platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go`
  — detector, handler transform and verifier, all three in one file, deliberately.
