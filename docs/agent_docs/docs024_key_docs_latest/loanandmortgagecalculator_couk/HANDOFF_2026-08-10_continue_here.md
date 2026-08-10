# HANDOFF — bugfix 224 and everything it grew into. START HERE in a new chat.

**Written 2026-08-10 by the bugfix-224 session**, superseding
`HANDOFF_2026-08-09_continue_here.md` (kept for history). Covers two live sites
and the platform acceptance fences.

## 0. State in one paragraph

Everything asked for is **done, live and verified**. Ten stale/NaN calculator
defects fixed across two sites; the arithmetic is pinned to an independent oracle;
**all 17 calculators on loanandmortgagecalculator.co.uk are now covered by the
platform's own unattended weekly sweep**, which already caught a real defect on
its first night. Nothing is in flight. **No task is owed.** What follows is
context, the traps, and the things a future session could choose to do.

---

## 1. What is live

### loanandmortgagecalculator.co.uk — `bugs_open/224`, ten instances
Six calculators each had a private copy of the annuity formula and none handled a
0% rate: three printed `£NaN` (compare-loans also **inverted its verdict**),
three left a **stale** previous answer, consolidation quoted **£0.00/month** for
an interest-free loan. Fixed by deleting the private copies and calling the shared
`assets/js/calculators.js`, plus one additive `calculateBalloonAmortization`.
Then **four more instances of the same class** were found and fixed —
equity-release (guarded on AGE), bridging-loan (on VIABILITY), investor ×2 (on a
BLANK input). **Oracle: 23 FAIL → 0; full estate PASS 170 / FAIL 0**, all four
mutation controls green.

> **The class is NOT "0%" and not "a rate". It is a GUARD THAT LEAVES A HANDLER
> WITHOUT WRITING THE DOM.** Six of ten were rate-guarded; grepping `r > 0` finds
> a fraction. `grep -n "return;"` and read three lines up; `alert(...); return;`
> is the highest-yield spelling. In LANDMINES.

### loancalculator.co.uk — the same defect, six components, FIXED + LIVE
Fixed through that lane's own pipeline (template → commit → `update_component
--apply` → `render_tool_row --apply` → assemble-only rerender). Sweep went
**5 of 8 pages affected → 0**; `defect_vectors.py --live` **16/16**; serving guard
26/26. **8 new `defect_vectors.py` cases** cover the 0% rate there — that lane had
none for those tools — and all 8 score **PROVEN** under `--both`.

### bugs_open/225 (SDLT) — fixed by a concurrent session, owner-approved
The expired £625k FTB cap and the missing £40k surcharge floor, on this site and
its byte-identical twin. Not my work; verified live (the expired rule appears
nowhere, stamp-duty reads 17/17).

### Platform acceptance — 17 fences installed, unattended, 17 of 17 covered
Emitted from the live pages, then **every pinned value re-derived from
`oracles.py`** (72/72) so the record cannot enshrine a wrong answer. Installed
into `doc_plans` by `install_fences.py`. **The sweep ran itself at 03:20 on
08-10: 14 runs, 13 passed, 1 real defect found** (equity-release), and
`no_auto_fix` held — **0 `improve_tool` items**, the run's own note saying *"NOT
auto-fixed — this fence declares no_auto_fix"*.

---

## 2. How the coverage works, and the three things holding it safe

Eligibility (`discovery_checks/tool_eligibility.go`) accepts EITHER
`cc.component_level='tool'` (branch a) OR `page_type='tool'` + no tool component
+ **exactly one** active component (branch b).

- **16 pages use branch (b)** — one `page_type` update each, no decomposition.
- **`loans-consolidation` uses branch (a)** — it has prose rows beside its tool,
  so (b) was impossible. Given a tool-level component whose **`function` is
  identical to `pages.name`**, which keeps the subject key stable so the existing
  fence still applies and page lookup still resolves.

Safety, all three verified rather than assumed:
1. **`no_auto_fix: true` on all 17 fences** — a failing arithmetic fence
   escalates to `needs_human_review` instead of dispatching a rewriter. Proven on
   a real failure.
2. **No `page_status_ok` in any fence.** Tier 2 reads the same fence, **ignores
   `no_auto_fix`**, skips every `computed_values` check, and raises `improve_tool`
   only on a failure — so `page_status_ok` was the single check it could fail, and
   its `improve_tool` aims at the `ported-page` shell **shared by ~154 pages on
   three sites** (already corrupted once, 08-04). Removed.
3. **For consolidation only**, which IS audited by `check_tool_health` (that
   check selects tool-level components and does not read `no_auto_fix`): the
   component is the page's own (blast radius 1), the page is
   `rebuild_policy='owned'` (`save_page_sections` refuses), and the row is
   `lock_type='permanent'` (rerender output discarded).

**Cost and cadence:** 7-day cooldown per tool, and each run makes a **Sonnet
vision call**. Budget ~17 vision calls a week.

---

## 3. Commands you will want

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
SIB=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk

# arithmetic truth, this site (live only; deploy first). 23->0 was the fix.
cd $LANE && python3 oracle.py --tools standard-calc,compare-loans,stress,settlement,car-finance,overpayment-calculator,consolidation
python3 oracle.py --selftest-parse && python3 oracle.py --mutate expectation --tools simple,repayment
python3 oracle.py --json /tmp/o.json            # full estate, ~12 min, expect PASS 170 FAIL 0

# the sibling site's 0% cases
python3 $SIB/rewrite/defect_vectors.py --live   # 16/16
python3 $SIB/rewrite/defect_vectors.py --both   # every case PROVEN / CONTROL, none VACUOUS

# fences
python3 $LANE/verify_criteria.py                # re-derive every pinned value from oracles.py
python3 $LANE/install_fences.py                 # dry run; --apply to (re)install
./docs/leopardessconsulting/scripts/tool_acceptance_run.sh ed633ada-f8af-424b-b4d4-8af79160dbcd loanandmortgagecalculator.co.uk <subject_key>

# after ANY repo edit on this site
python3 $LANE/gate_component_bytes.py --repair && python3 $LANE/gate_component_bytes.py
```

Site ids: this site `ed633ada-f8af-424b-b4d4-8af79160dbcd`; loancalculator
`0162cde4-633e-45e9-8ca6-87a6b2fe1d26`. Sites repo `/home/ant/projects/sites`
→ push → GH Actions → B2; wait ~100 s past the run or B2 serves a `NoSuchKey`
blob at HTTP 200 that greps clean.

---

## 4. Traps found this session (all in LANDMINES, all cost real time)

- **`gate_component_bytes.py --repair` would have destroyed a decomposed page** —
  it compared every row against the whole repo file. Fixed: verbatim +
  single-component only.
- **The Tier-4 runner opens ONE page per (url, profile)** and does not reload
  between checks, while the capture used a fresh page per vector. Any fence whose
  clicks BUILD state needs `{"action":"reload"}` first. Consolidation failed 3/4
  until that was added.
- **`render_tool_row.py`'s default `--control-ref` is stale** — it refuses
  correctly; pass the commit that produced the stored rows.
- **`/tools/standard-calc.html` on loancalculator is a 404 with live DB rows**,
  sharing its component with the HOMEPAGE.
- **`zero_rate_sweep.py` is blind to a DETERMINISTIC zero** (consolidation) — a
  clean line there means *unmeasured*.
- **`toolgolden 11/11 exact` cannot refute this class** — vectors scale each
  field's own default, and no scaling of 7.9 is 0.
- **`--emit-criteria` can pin the page's INITIAL MARKUP as an expected answer**
  when the tool writes nothing for a vector. It gates a wholly inert tool, not a
  tool inert on one vector.

## 5. Five times the red result was MY harness, not the site

Fractional-term rounding; a wrapper with no `<meta charset>` (mojibake ✅ read as
a failure); `verify_criteria.py` collecting only `fill` steps so stamp-duty's
`select` was dropped and a **correct** tool was reported £5,000 wrong; a "no NaN
on the page" check that matched **my own fix comment**; and a probe that filled
investor's fields but never clicked its buttons. **On this estate, assume a red
result from a checker you wrote today is the checker.** Print the inputs it drove
and the raw value it compared before believing it. All in `WRONG_CALLS.md`.

## 6. Owner decisions already taken — do not reopen

1. **Unattended fences: YES** — done, 17 of 17.
2. **A shared engine for loancalculator.co.uk: NO, and not worth an RFC.**
   CLOSED. Its only shared JS plumbing is the **fleet-wide** `js_snippets` table
   (no `site_id`), so it is architecture scope. **Do not re-log this as tech
   debt**; the 8 PROVEN defect cases are what stop the copies drifting.
3. **Re-baseline: DONE** — `GOLDEN_2026-08-09_postfix.json` and
   `BASELINE_2026-08-09_stored_md5_at_b26fdc81b.txt`, both NEW files with the
   08-05 pair kept; `load_lmc.py` re-pointed. 17 of 41 pages had moved and every
   one was accounted for first — that guard exists to catch another session's
   write, so a blind regeneration absorbs what it should surface.

## 7. If you want more (nobody has asked)

- **Sweep the fleet for the same defect class.** Ten instances on two sites came
  from one mechanism; `mortgagecalculator.co.uk` and `loancash.co.uk` share this
  family's ancestry and have never been checked for it. `grep -n "return;"` per
  calculator page is the cheap first pass.
- **The other 6 pages on this site are branch-(b) eligible but have no fence**
  (application-tracker, credit-health-check, damage-checker, fact-finder,
  investor, portfolio). They emit nothing while criteria are absent; Tier 2 will
  write a `needs_criteria` note each. Class C in the oracle's own classification —
  no external right answer — so they want INVARIANT checks, not arithmetic.
- **`investor` still has no fence at all**: `toolgolden` cannot certify a ratio
  tool (uniform scaling leaves a ratio invariant — its own LANDMINE), so its
  criteria would have to be hand-written. Its buttons now have ids, so it is
  ready if someone wants to.

**Council:** nothing here was submittable — the gate's scope is
`platform/ internal/ pkg/` and every change was site content, DB rows or lane
tooling. **No Go changed, so none of this needed the chassis build.**
