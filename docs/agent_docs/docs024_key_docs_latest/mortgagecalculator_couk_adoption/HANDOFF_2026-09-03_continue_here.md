# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-09-03)

**Supersedes `HANDOFF_2026-09-02b_continue_here.md`.** That file's §1 (the refutation of the
render-path theory) and §4 still hold; its §2a scoreboard is superseded by §2 below.

**Two things changed overnight and both change the job:** migration `701` landed, and the owner
handed the imagery work to another lane.

---

## 0. State in one paragraph

18 tool pages, all serving 200. **All 18 are now ladder-eligible (was 9)** — `701` did that.
**13 of 18 now resolve to a live acceptance fence (was 6)** — this lane re-addressed all 8 of its
fences today, because `701` silently orphaned every one of them. **Imagery is no longer ours.**
Open bugs owned here: `441`, `448`, `449`. Nothing on the site is known broken.

## 1. Imagery — HANDED OVER, do not work it

Owner, 2026-09-03: *"hand that lane the whole job."* mortgagecalculator.co.uk's tool imagery now
belongs to `bugfix_114_imagery_wiring`, including the spend decision. Everything we knew is in
`docs/agent_docs/docs024_key_docs_latest/bugfix_114_imagery_wiring/CONTRIB_2026-09-03_from_mcalc_lane_OWNER_HANDS_YOU_THE_WHOLE_TOOL_IMAGERY_JOB.md`,
with a pointer and a correction in `bugs_open/114`.

⚠ **The correction matters if you read the older handoffs:** our 09-02 claim that ten tool pages
*"have nowhere to put a picture"* is **no longer true.** `701` gave each calculator its own
`component_level='tool'` component and freed the hero slot. The composition change is safe as of
today and was genuinely unsafe yesterday. **Do not re-derive this and do not act on it — it is
their call now.**

## 2. Tools — the scoreboard, and what 701 did to it

**Eligibility 9 → 18.** Every tool page now has a tool-level component, so both ladder tiers can
see all of them. That is the real gain from `701` and it is larger than the migration's own notes
claim.

**All 8 lane fences were orphaned by it, and nobody re-keyed them.** The ladder keys on `cc.function`
for a tool-level component, so the subject key moved `<slug>` → `tool-<slug>`. Left alone the
fences addressed nothing, silently — Tier 2 writes `needs_criteria`, Tier 4 emits nothing, no error
anywhere. **Repaired today** (see §3).

| state | tools |
|---|---|
| **PASSING** (verdict on record) | `simple`, `tool-overpayment-priority`, `tool-rate-scenarios`, `tool-bridging-compound` |
| **fence live, re-run dispatched 2026-09-03 10:52** | the 8 re-addressed ones — `tool-simple`, `tool-repayment`, `tool-equity-release`, `tool-fee-analyser`, `tool-rate-forecaster`, `tool-bridging-loan`, `tool-overpayment`, `tool-stamp-duty` |
| **FAILING** — `441` stale fence, fixer blocked by `448` | `tool-deposit-tracker`, `tool-remortgage-savings` |
| **no fence at all** | `tool-affordability`, `tool-btl-investor`, `tool-credit-health-check`, `tool-portfolio`, `tool-rate-stress-test` |

⚠ **A PASS here is weaker than it sounds — `bugs_open/449`.** The 8 lane fences assert arithmetic
and nothing else; the 5 generator fences assert liveness and **no number at all**. Not one tool on
this site is verified for both. Say which kind when you report a pass.

## 3. What was done today, and the checks that could have stopped it

All 8 fences re-addressed in `doc_plans` by supersede-and-insert, one guarded transaction:
subject key `<slug>` → `tool-<slug>`; selectors re-pointed with the `c-tool-<slug>-` prefix on the
**5** pages already re-rendered; **no expected value changed**. Five checks ran first:

1. no key collision (`idx_doc_plans_current` is UNIQUE on `(subject_type,subject_key) WHERE is_current`);
2. no other site consumes those keys (`doc_plans` keys are fleet-wide — this is the caution we gave
   the 357 lane and then owed ourselves);
3. every new selector verified present in the live page, using a **verbatim reimplementation** of
   `selectorAnchor` + `anchorPresent`, not an approximation;
4. **the control**: the OLD fences re-tested against the 5 scoped pages come back 4/4/5/7/6 anchors
   absent — had the transform been a no-op this reads 0;
5. an in-transaction `RAISE EXCEPTION` unless exactly 8 new current / 0 old current. It returned
   `OK: 8 new current fences, 0 old remaining`.

Old rows are **superseded, not deleted** — history intact, provenance in `created_by`.

⚠ These fences carry **`no_auto_fix: true`**, checked *before* dispatching: a failing arithmetic
verdict reaches a human, not `tool-improver`. That is why re-keying could not aim a rewriter at a
working calculator.

**Also fixed: `acceptance/verify_criteria.py` was dead**, and took `install_fences.py` with it. It
float()'d every fact in `evidence_base`; the register now holds 5 value-less `CIT-*` citations among
18 facts. Two defects one line apart (`float('')`, and a JSON-null value making psql print a row
with **no trailing separator**, so `split("\t")` unpacked into two names). Proved both directions:
clean run 84/84 agree 0 MISMATCH with `register: 13 SDLT facts loaded live`; `--mutate
sdlt-ftb-relief-cap=625000 stamp-duty` still reports 1 MISMATCH and **exits 1**.

## 4. ⚠ THE LIVE HAZARD — read before you re-render anything here

`701` created every adopted component with an **instance-scoped template** (`{{.InstanceID}}-`) while
preserving the old rendered bytes. Those two agree only until something renders. This morning's
rebuild wave re-rendered **5 of 10** at 08:46–08:49Z and their ids changed (`amt` →
`c-tool-simple-amt`).

**So the site is half-converted right now: 10 templates scoped, 5 renderings scoped, 5 still bare**
(`tool-affordability`, `tool-bridging-loan`, `tool-overpayment`, `tool-portfolio`,
`tool-stamp-duty` — all still at the `2026-09-02 21:06:35` transaction timestamp, re-checked
10:59Z).

- The tools are **fine** — 0 dangling JS bindings on all ten; the converter rewrites bindings with
  the ids.
- **Each of those 5 renders invalidated that tool's fence**, and the remaining 5 will do the same
  whenever they next render. Three of the fences installed today carry **bare** selectors and are
  correct only until that happens.
- **If you re-render any of those five, re-point its fence in the same session.** The transform is
  a pure `#x` → `#c-tool-<slug>-x` prefix; verify each against the live page before writing.

This is why `bugs_open/441` was re-framed today: **it is not a backlog, it is a live generator of
stale fences.** Only the scope-aware checker (441 candidate 1) is stable.

## 5. Next, in order

1. **Read the 8 verdicts** from this morning's runs — `doc_notes`, not the work item.
   Items: `879ee87a` simple, `076377ba` bridging-loan, `a570b486` rate-forecaster, `36db5755`
   repayment, `b1bdb777` stamp-duty, `2efd13f7` fee-analyser, `89a3cc7a` overpayment, `d9d4dce0`
   equity-release. ⚠ They queue behind fleet fairness — a single run took 5m38s to be claimed, and
   eight had not been claimed after 7m30s. **Slow is normal; a missing row is latency, not a
   dropped dispatch.**
2. **`bugs_open/441` candidate 1** — the platform fix. Council gate; it is `platform/` code.
3. **`bugs_open/448`** — 62% of every failed `improve_tool` on the estate. Small fix, sound code
   already 300 lines below in the same file.
4. **Fences for the 5 unfenced tools.** ⚠ `tool-portfolio` needs an ORACLE first — `verify_criteria.py`
   reports *"NOT VERIFIED (no independent model): fact-finder, portfolio"*, which is why it was never
   installed. **Do not install a fence whose values were only emitted from the page.**
5. `install_fences.py` still keys on the criteria **filename**, which `701` also invalidated — it
   now SKIPs all 10 as *"not ladder-eligible"*. Rename the files to `tool-<slug>.criteria.json` and
   sync the 5 scoped ones' selectors to what is installed, or the next person re-installs the old
   bare selectors over today's repair.

## 6. Unchanged, still open

The copy-quality CONTRIB of 08-26 **still needs an answer**. `/scorecard-simulator.html` remains the
one dead internal link. The 13 `fact_drift_review` items, the contact-page email and the "Contact"
title question are still the owner's.

## 7. Files of record

`NOTES_mortgagecalculator_couk.md` — `## 2026-09-02 (b)/(c)/(d)` and `## 2026-09-03` ·
`README_where_we_are.md` (owner's log) 09-02 ×2, 09-03 ·
`RUNBOOK` §15 (eligibility), §16 (extractor traps), §17 (demand control), §14 corrected ·
`bugs_open/441` (re-framed), `448`, `449` · `bugs_closed/357` (post-close CONTRIB) ·
`WRONG_CALLS.md` 09-02 ×3, 09-03.
