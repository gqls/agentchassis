# HANDOFF — loancalculator.co.uk · continue here (2026-08-08)

**Supersedes `HANDOFF_2026-08-05_continue_here.md`.** That file's §3 (the `plan_sections`
blocker) is CLOSED — 204 and 189 were fixed, reviewed and proven by the
`bug_backlog_clearing` lane on 08-06. Its §4 (voice seeded), §5 (other finance sites
held) and §6 (fleet-wide base prompt) are still accurate and NOT repeated here.

Read order for a cold start: this file → `SUMMARY_2026-08-08_the_site_speaks_in_the_new_voice.md`
→ `PLAN_2026-08-08_voice_h_rollout.md` → the 2026-08-08 sections of `NOTES`.

```
site         loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis      v1.0.1263
voice H      23 of 26 active pages  (3 blocked by bugs_open/219, unchanged and healthy)
live         26/26 HTTP 200 · 12/12 locked calculator rows identical · toolgolden 11/11 exact
```

> # ✅ DONE — 2026-08-08 evening. THE ROLLOUT IS COMPLETE; §1 AND §4 BELOW ARE CLOSED.
>
> Read `SUMMARY_2026-08-08c_the_rollout_is_complete.md` first; this file is kept for
> its still-live §2/§3/§5/§6, which are unchanged and still worth reading.
>
> ```
> voice H      26 of 26 active pages          (was 23; the last 3 landed on v1.0.1266)
> live         26/26 HTTP 200 (swept live)
> calculators  11/11 identical to golden, re-baselined GOLDEN_2026-08-08_voice_h_complete.json
> 219          FIXED, LIVE, PROVEN — kept in bugs_open/ by owner direction, evidence inside
> 221          OPEN, belongs to webdesign.co.uk — same check, different mechanism
> ```
>
> **§1 is closed:** `index`, `tool-car-finance-calculator` and
> `tool-interest-rate-stress-test` were rebuilt through the framework, graded 3/3 PASS,
> and their work items completed after grading. **§4 is closed:** golden re-baselined
> (11 pages, now including the `asym` vector the old one predated) — and the compare
> was run BEFORE the capture, deliberately, because capturing first would have blessed
> whatever the rebuild did.
>
> **NOTHING IS OWED ON THIS LANE.** §5 was answered the same evening — the owner ruled
> **keep the explanatory copy**, so nothing is trimmed and there is no follow-up work.
>
> **If you are here to rebuild a page anyway, the two checks that earned their keep:**
> a roll is not evidence, so grep a positive AND a negative marker on **the pod that
> will run your action** (the fleet was MIXED — the `agent-chassis` deployment had
> rolled while 20 spawned agent pods were still on the old image):
> `strings /app/agent-chassis | grep -c "Schema/pipeline vocabulary in visible page copy"`
> (expect ≥1) and `'the model wrote about its task instead of doing it'` (expect **0**).
> And prove `voiceh_grade.py` can FAIL before believing a pass — mutate the baseline to
> carry the page's CURRENT row ids and re-run; it must flag every row.
>
> **`index` GOES THROUGH THE FRAMEWORK LIKE EVERYTHING ELSE — OWNER RULING,
> 2026-08-08.** Build all three together; no special handling, no holding it back.
>
> I had flagged `index` as "no longer a routine rebuild", because a parallel session
> established that its prose was last written **2026-08-02** by the decomposition — so
> the page still carries **original hand-built copy**, and the owner's *praise* for it,
> as well as his complaint about the opening, are both about writing the framework
> never touched (`ecf50b634`, `SUMMARY_2026-08-08b…`). That observation is correct and
> stays on the record. **The caution I drew from it was wrong and the owner removed
> it:** *"I am happy for the whole site to be built through the framework because that
> is what I am judging."*
>
> The reasoning matters more than the instruction, because it generalises: **preserving
> hand-built copy the owner liked protects the wrong thing.** What is under evaluation
> is the framework's output, so a page that keeps its hand-built paragraphs is a page
> that tells him nothing — and the same standing rule already forbids the alternative
> (`CLAUDE.md`, owner ruling 2026-08-04: every site goes through the framework, never
> hand-built). If a framework rebuild makes `index` worse, **that is the finding**, and
> the pre-run bytes are in `page_components_bak_20260807_voiceh` either way.
>
> Fix detail, the correction it required, and the council's JSON-LD objection:
> `bugs_open/219`. A second live instance of the same check, which this fix does
> **not** solve: `bugs_open/221` (webdesign.co.uk — that lane has been told).

## 1. The rewrite is DONE except for three pages, and those are a platform bug

`index`, `tool-car-finance-calculator`, `tool-interest-rate-stress-test` could not be
rebuilt. **Nothing is wrong with them or with the copy** — they serve fine, in the old
voice. `bugs_open/219`: `validate_page_content`'s meta-commentary check substring-scans
the WHOLE assembled page HTML including comments, and those three pages carry a locked
calculator whose `html_template` holds a developer changelog comment containing the
string `input_schema`. Blocker severity, so the build fails before `save_page_sections`.

**Deterministic, and proven by predictor rather than argued:** exactly the 3 of 12 tool
pages whose template contains that string failed — all three of them, three times over
for car-finance — and all 9 without it passed.

**When 219 ships**, each page is about three minutes:
```bash
./voiceh_batch.sh index tool-car-finance-calculator tool-interest-rate-stress-test
python3 voiceh_grade.py <baseline> index tool-car-finance-calculator tool-interest-rate-stress-test
```
Then re-baseline the golden (see §4). Do **not** work around 219 by editing the
templates — that treats the instance and leaves the class, and it edits a locked
calculator to satisfy a validator.

## 2. The tools in this lane, and what each is for

| script | what it does |
|---|---|
| `voiceh_rewrite.sh <page>` | one page, through the framework. Copies the reviewed prompt **by SQL** from canary item `2517bc4b`; files the item `detected` (undispatchable) so only the direct publish fires it |
| `voiceh_batch.sh <page>…` | fires a batch, waits for every orchestration to go terminal, prints the grade command |
| `voiceh_grade.py <baseline.json> <page>…` | the grader. Row identity, facts, links, heading structure, CSS survival, locked rows, and the served page with a negative control |
| `voiceh_restore_css_slot.sh <page> [slot]` | exact row restore from the backup + redeploy, for the CSS trap in §3 |

Assets: `page_components_bak_20260807_voiceh` (63 rows, pre-run) and
`scratchpad/…/voiceh/baseline_20260807.json` (76 KB — every row's full text, length,
md5 and row id, so it serves fact comparison as well as restore). **If the scratchpad
is gone, re-create the baseline from the backup TABLE**, which is the durable copy.

## 3. ⚠ The trap that cost the most, and it is not obvious

**8 of this site's 51 `prose-*` rows are not prose — they hold the page's `<style>`
block.** Rewriting one deletes the CSS that lays the calculator out, and **every guard
in the platform passes it**: the component's own guidance promises "no element addressed
by any script, so rewriting this prose cannot break a calculator" (true, and silent
about CSS), the lock protects the tool row not the style row, and the validator does not
object. The arithmetic still computes; only the layout collapses.

It bit 4 pages here. **The writer PRESERVED the style block on the other 4** — so it is a
coin flip and a spot-check of one page clears the class wrongly. `voiceh_grade.py` now
fails on any lost selector; keep that check. Full entry in `LANDMINES.md`.

## 4. What is owed

- **The 3 pages**, once 219 is fixed (§1).
- **Re-baseline the calculator golden.** `GOLDEN_2026-08-03b` still matches exactly, so
  nothing is broken — but the prose around 10 calculators has changed and a future
  capture will diff on `dom_shape`. **Not** done yet, deliberately: 3 pages are pending
  and `toolgolden.py` refuses a partial capture because such a file certifies nothing.
- **An owner decision on expansion** (§5).

## 5. ~~The open question for the owner~~ — ANSWERED 2026-08-08: **KEEP THE COPY**

> **OWNER RULING, 2026-08-08 evening: "keep the explanatory copy."** Nothing to do —
> the pages stay exactly as the framework built them. **Nothing was trimmed, and
> nothing should be.** This was the last item owed on the lane; with §1 and §4 closed
> the same evening, the workstream is complete.

The question, kept because the ruling is only legible next to it: several calculator
pages had near-empty prose stubs (32–156 bytes). The framework did not restyle them, it
**filled them**, to 800–1,900 bytes of new explanatory copy. It reads well and nothing
in it trips the claim gate, but it was **new substance on a finance site**, and the
brief said "voice only, preserve every fact, add nothing". Both readings were
defensible — the pages were thin and are now useful; and the writer exceeded its
instruction. The owner took the first.

**What the ruling covers, stated so a later thread does not silently widen or narrow
it.** The expansion itself, explicitly. I am reading it as also settling the two
smaller observations in the same family, which were listed under this question and are
the same act by the same writer — it expands `Consumer Credit Act` to `Consumer Credit
Act 1974` (correct, and still an addition), and it reworded two headings including one
`h1` (`Typical Fees & The Power of APR` → `Hidden Loan Fees & The Power of APR`, which
matches the URL better). If that reading is wrong it is a trim of two headings and a
statute year, not a re-run. **Do not re-open this as "should we trim?" — it is decided.**

⚠ **`page_components_bak_20260807_voiceh` (63 rows) is no longer the undo for this
decision**, since there is nothing to undo. Keep it anyway: §3's CSS trap is what it
actually protects against, and `voiceh_restore_css_slot.sh` reads it.

## 6. Method notes worth keeping

1. **A checker built from the shape of the OLD artefact will convict the new one of
   being different — and it points at the system, not at the checker.** Mine did it
   three times: two distinct bugs behind one identical "the copy never shipped" message,
   and a digit-matcher that called every spelled-out figure a lost fact (the voice
   deliberately writes "four to five years"). Where a check and a spot-read disagree,
   **suspect the check** — probe for any other chunk of the new copy on the same page.
   In `WRONG_CALLS.md`.
2. **An error message names a cause it has not established.** `meta_commentary` says
   *"the model wrote about its task instead of doing it"*. It cannot know that — it is a
   substring scan. I repeated it for an hour before a predictor falsified it.
3. **Prove the grader can FAIL before trusting a pass.** Running it against a
   not-yet-rewritten page (`guide-jargon-buster`, then `index`) correctly failed every
   row on row-identity. A grader that has never failed is not evidence.
4. **Complete a work item AFTER grading, never on the orchestration's own status.** A
   direct dispatch bypasses the loop's `mark_complete`, so `detected` at the end
   honestly means "built, not yet graded".
5. **A jsonb path that misses returns NULL, which is indistinguishable from "someone
   reverted it".** `slot_name_from` read absent because I had the path wrong — and a
   revert of that exact key was a live, documented risk. Prove the path resolves
   (`config::text LIKE '%key%'`) before believing an absence.
