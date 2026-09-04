# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-09-03b, evening)

**Supersedes `HANDOFF_2026-09-03_continue_here.md`**, which was written at midday and patched five
times as the day moved under it. Nothing in it needs reading: everything true is restated here, and
its §2a scoreboard and §4a hazard are both **spent** — the hazard fired, and the scoreboard was
built on fences I had broken without knowing.

---

## 0. State in one paragraph

18 tool pages, all serving 200, **all 18 now instance-scoped** (only `investor-index` still renders
bare ids, and it is a section-index, not a tool). **All 18 are ladder-eligible** (was 9 yesterday).
**13 of 18 resolve to a live fence, and as of this evening all 8 of this lane's fences are correct
against the pages as actually served** — they were not, for most of today, and that was my doing.
8 Tier-4 runs are in flight. Imagery is another lane's. `449` is another lane's. Open here: **`441`
and `448`**.

## 1. ⚠ Read this before you trust any PASS or FAIL on this site today

**Five Tier-4 failures recorded between 12:18 and 13:46 are MINE, not the calculators'.** The tools
are fine. The fences were broken by an incomplete repair I made this morning, and I did not catch it
because my verification shared the repair's defect.

What happened, precisely, because the shape is worth more than the incident:

- The fence format holds selectors in two places: `steps[].selector` (the inputs to drive) and
  **`expect_values`** — a map whose **KEYS are selectors** and whose values are the expected text
  (`run_checks_action.go:238`, `ExpectValues map[string]string \`json:"expect_values"\``).
- My morning transform matched a key named **`values`**. That key does not exist. So it prefixed
  every input selector and **left every assertion selector bare**.
- The result passes every eye test: the tool fills the right boxes, clicks the right button, and
  then asserts `#monthlyResult` on a page that now calls it `#c-tool-simple-monthlyResult`.
- **My verification walked the same wrong key name**, so it reported ALL EIGHT SATISFIABLE while
  four of the eight were half-broken. A checker built from the same misunderstanding as the thing it
  checks cannot fail.

**Fixed this evening, and verified differently on purpose:** the new check reads the raw fence JSON
**textually** (`"#..."` occurrences) rather than walking typed keys, so no key name can hide from it.
Control: the superseded fences, tested the same way, are absent on **1–9 selectors each**. Applied
under an in-transaction assertion; **verified again by reading the fences back out of `doc_plans`**
rather than trusting my local files. **No expected VALUE was changed at any point today** — only
addresses.

⚠ **If you see the 12:18–13:46 failures in `doc_notes`, do not treat them as evidence about the
tools.** They are `#monthlyResult`, `#pay1`, `#displayMonthly` etc. "absent from the live DOM" —
which is exactly what a correct calculator looks like through a fence pointed at its old ids.

## 2. What the day actually did to this site

| | before | after |
|---|---|---|
| ladder-eligible tool pages | 9 | **18** |
| pages resolving to a live fence | 6 | **13** |
| tool pages rendering instance-scoped ids | 0 | **18** |
| this lane's fences correct against the served page | 8 | 0 (morning) → **8 (this evening)** |

Three actors, none of them wrong, composed all of it:

1. **`701`** (owner-applied 09-02 ~21:06) retyped 11 adopted rows to `component_level='tool'`,
   adopting the bodies **verbatim with bare ids**. This is what made all 18 eligible.
2. **The instance-scope sweep** (09-03 07:40) correctly found those components unconverted, filed 11
   `instance_scope_conversion` items, and rewrote the templates 08:36–08:46.
3. **The rerender wave** published the result — 5 pages at 08:46–08:49, the remaining 6 by 16:08.

⚠ **Do not attribute the id rewriting to `701`.** I did, all morning, in six documents including a
note to the 357 lane telling them their md5 guarantee had expired. It had not; retracted in full.
`701` did exactly what it said.

## 3. What is in flight right now

> **⚠ STALE AS OF 2026-09-04 — these eight verdicts no longer describe the tools they name.**
> **25 `improve_tool` fixes landed on this site 05:34–12:04 on 09-04**, and **all eight** of these
> tools have had both their `content_components` row and their `page_components` rendering rewritten
> since. Several fixes were ARITHMETIC ("the final-month interest correction is mathematically…",
> "the basisText logic is inverted", "computeBands stores `r.to` as `band.upTo`"), and these fences
> pin exact values — so a pinned value may legitimately have moved.
> **The fences themselves survived: 0 of 8 broken**, checked against the live pages 2026-09-04 —
> the rewrites preserved the `c-tool-…` ids. So this is a stale VERDICT, not a stale fence.
> **Re-fired 2026-09-04** as `773d6f9c` simple · `a1db08c0` repayment · `dd57a4e7` equity-release ·
> `5fb4488e` fee-analyser · `2b17ba03` rate-forecaster · `d7f86c80` bridging-loan · `2e58d5f2`
> overpayment · `d09cce26` stamp-duty. **Read those, not the ones below.**
> ⚠ **A failure in the re-run may be CORRECT** — if a fix changed a value the fence pins against an
> independent oracle, the fence is doing its job and a human decides which number is right. That is
> what `no_auto_fix` is for. Do not assume a fail means the tool broke.

**LANDED (2026-09-03) — 8 of 8 PASS**, all terminal by 17:21:52Z, each run ~50s. simple, repayment,
equity-release, fee-analyser, rate-forecaster, bridging-loan, overpayment, stamp-duty.

That closes the loop three ways: the calculators were correct throughout, the re-point is right, and
the 12:18–13:46 failures are now positively explained as mine — nothing about the tools changed
between the two runs, only the fence addresses.

⚠ **Quote the scope line, not the word PASS.** The verdict says it itself: *"Scope of this verdict:
this fence compares 4 exact values"* — arithmetic, desktop only, nothing about boot, console, status
or mobile.

**Read the verdict in `doc_notes`, not on the work item.** These fences carry `no_auto_fix: true`, so
a failure reaches a human and never `tool-improver` — verified: `parseNoAutoFix` is honoured at
`tool_acceptance_actions.go:876/907/923`, and Tier 2 ignores it entirely (one comment, `:292`).

⚠ **Expect queueing, and do not re-dispatch to hurry it.** Site selection ranks by oldest-waiting
row across ~18 sites; this site went 08:49→16:08 with nothing claimed earlier today. That is
rotation, not a fault — checked every arm (not locked, no `claimed` row held, no `retry_after`,
governor admits the type, we ranked 5th and eligible). Contributed as data to the
`dispatch_throughput` lane, who own that shape.

## 4. The scoreboard, and what a PASS is worth here

| state | tools |
|---|---|
| **PASSING** on a verdict that predates today's fence churn | `tool-overpayment-priority`, `tool-rate-scenarios`, `tool-bridging-compound` |
| **PASSED 09-03 on arithmetic — ⚠ VERDICTS STALE, tools rewritten 09-04, re-fired** | `tool-simple`, `tool-repayment`, `tool-equity-release`, `tool-fee-analyser`, `tool-rate-forecaster`, `tool-bridging-loan`, `tool-overpayment`, `tool-stamp-duty` |
| **FAILING for real** — `441` stale fence, fixer blocked by `448` | `tool-deposit-tracker`, `tool-remortgage-savings` |
| **no fence at all** | `tool-affordability`, `tool-btl-investor`, `tool-credit-health-check`, `tool-portfolio`, `tool-rate-stress-test` |

⚠ **A PASS here is narrower than it sounds** — this is `bugs_open/449`, now owned by the
`bugfix_449_fences_assert_no_number` lane. This lane's 8 fences assert **arithmetic only** (no boot,
console, status or mobile check); the generator's fences assert **liveness only** and no number at
all. **No tool on this site is verified for both.** Say which kind when you report a pass.

## 5. Next, in order

1. ~~Read the 8 verdicts.~~ **Done — 8 of 8 PASS.** 11 of 18 tools now hold a passing Tier-4
   verdict (4 this morning, 1 usable yesterday). If a future run fails, check the named selector
   against the served page **before** believing it — a fence can be wrong in a way that looks exactly
   like a broken tool, and did today.
2. **`bugs_open/441`** — the platform fix (`anchorPresent` + the Tier-4 selector path accept both
   spellings). Council gate; **nobody holds it**, and today is the argument for it: five fences broke
   by rerender at 08:46 and three more by 16:08, none of it intended by anyone.
3. **`bugs_open/448`** — 62% of every failed `improve_tool` fleet-wide; the sound code is 300 lines
   below the broken code in the same file.
4. **Fences for the 5 unfenced tools.** ⚠ `tool-portfolio` needs an ORACLE first —
   `verify_criteria.py` reports *"NOT VERIFIED (no independent model): fact-finder, portfolio"*.
   Never pin a value that was only emitted from the page.
5. `install_fences.py` still keys on the criteria **filename**, which `701` invalidated — it now
   SKIPs all 10 as *"not ladder-eligible"*. Rename the files to `tool-<slug>.criteria.json` and sync
   their selectors to what is installed, or the next person re-installs bare selectors over today's
   repair. **The files on disk are now three revisions behind `doc_plans`.**

## 6. Owned elsewhere — do not start these

- **Imagery** → `bugfix_114_imagery_wiring` (owner ruling 09-03). Our `hero-tool` finding is in
  `bugs_open/114`; the ten pages that had "nowhere to put a picture" now do, since `701`.
- **`bugs_open/449`** → `bugfix_449_fences_assert_no_number`. They extended it (186 fences / 115
  blind; 91 drive inputs and 55 of those assert nothing) and hold the framework half. We keep this
  site's fences. Their three questions are answered in
  `CONTRIB_2026-09-03_from_the_449_lane_…`.
- **Dispatch rotation** → `dispatch_throughput`.

## 7. Still open, unchanged

The copy-quality CONTRIB of 08-26 **still needs an answer**. `/scorecard-simulator.html` is still the
one dead internal link. The 13 `fact_drift_review` items, the contact-page business email and the
"Contact" title question remain the owner's.

## 8. Files of record

`NOTES_mortgagecalculator_couk.md` — `## 2026-09-02 (b)/(c)/(d)`, `## 2026-09-03` ·
`README_where_we_are.md` (owner's log) 09-02 ×2, 09-03 ×2 ·
`RUNBOOK` §14 (corrected), §15 eligibility, §16 extractor traps, §17 demand control ·
`bugs_open/441` (re-framed as a live generator), `448` · `bugs_closed/357` (post-close CONTRIB +
retraction) · `WRONG_CALLS.md` 09-02 ×3, 09-03 ×4.
