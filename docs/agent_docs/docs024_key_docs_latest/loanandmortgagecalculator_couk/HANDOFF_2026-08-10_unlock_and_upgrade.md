# HANDOFF — unlocking loanandmortgagecalculator.co.uk + loancash.co.uk for full framework editing

> **⚠ SUPERSEDED 2026-08-10 (evening) — START AT
> `HANDOFF_2026-08-10c_continue_here.md`.** §1–§3 (what the lock is, what migration
> 367 did, why the 20 tool pages were refused) remain accurate and are still the
> reference for the lock mechanism. **§4 IS WRONG — see the correction below.**
>
> **CORRECTED 2026-08-10 (evening): §4's blocker no longer exists, and did not exist
> when this file was written.** §4 says `bugs_open/204` means *"a decomposed page can
> never be rebuilt"* and to read it before seeding a plan. **204 and its sibling
> `bugs_open/189` were both fixed, rolled and behaviourally verified on 2026-08-06**
> — four days before this file. They remain in `bugs_open/` only because of the
> owner direction of 2026-08-06 to leave found bugs there, which overrides
> CLAUDE.md's bar. Re-verified at chassis v1.0.1280 on both replicas by pod-grep with
> a negative control, plus the live config half. **What caught it:** checking the
> tail of the bug file instead of inferring its status from its directory.
> The build path now resolves sections by `page_components.component_id` first, so
> positional slot names resolve — and 189's *"never fire a build-path run on a page
> holding locked rows"* is lifted too. Full working: `HANDOFF_2026-08-10c` §2.

**Owner request, 2026-08-10:** *"unlock them both and make their components and
tools fully editable and upgradable … do it all through the framework."*

**Read this first if you are picking the work up.** Part 1 is DONE and live. Part
2 is the substantial half and is NOT started — it must not be done by flipping a
flag, and this file explains exactly why and what to do instead.

Chassis live at time of writing: **v1.0.1277** (both replicas, started
2026-08-09T21:35Z).

---

## 1. What "locked" actually was, measured

`pages.rebuild_policy` — `'generic'` | `'owned'` (migration 164's CHECK allows
only those two). `'owned'` means "not the generic pipeline's to rebuild" and is
enforced in four places:

| where | what it does |
|---|---|
| `get_pages_to_build_actions.go:128` | excludes owned pages from the build queue |
| `reconcile_site_plan_action.go` | emits `owned_page_review` instead of `needs_page` |
| `save_page_sections_action.go` | hard-refuses a generic section save |
| `owned_page_guard.go` (`assemble_page`) | refuses in the three composition loops |

**Both sites were 100% owned**: 41 pages on loanandmortgagecalculator.co.uk
(`ed633ada-f8af-424b-b4d4-8af79160dbcd`), 18 on loancash.co.uk
(`ee4a8199-4f5b-4e2e-88ce-01e600721b74`). Component locks were almost absent —
exactly **one** `page_components` row carried `lock_type='permanent'`
(consolidation's `tool-1`); everything else was unlocked.

⚠ **Editing was never gated by this flag.** Migration 164 says in terms that
re-assembly of existing `page_components` "is deliberately NOT gated — it is how
owned pages deploy", so `page-rerender` and `section-editor` /
`apply_section_edit` already worked on owned pages. What `'owned'` blocks is the
generic pipeline REBUILDING a page wholesale. Do not describe the before-state as
"uneditable"; it was "not upgradable by the generic pipeline".

## 2. DONE — 39 prose pages unlocked (migration 367, applied + recorded)

`docs/agent_docs/sql_for_agents/367_unlock_prose_pages_loanandmortgage_and_loancash.sql`
(+ `_ROLLBACK`). Applied by hand — **not** `--apply`, which would have swept up
other threads' pending files — then `--record-only` with the verification note.

```
loanandmortgagecalculator.co.uk   generic  24 prose   |  owned  17 tool
loancash.co.uk                    generic  15 prose   |  owned   3 tool
```

The migration stamps every page it changes into `_mig367_unlocked_prose_pages`,
so the rollback re-locks **exactly those rows** rather than "everything on these
domains" — another thread may legitimately flip a page after 367 ran.

**Its verify block is DO/RAISE, and BOTH assertions were induced before applying**
(a verify block made of `SELECT`s cannot stop a `COMMIT` — `ON_ERROR_STOP`
ignores a non-empty result set):

```
induced population mismatch (39 -> 40)        -> ERROR, transaction aborted
induced tool pages INTO the target set        -> ERROR: "NEGATIVE CONTROL FAILED
                                                 — a calculator page has been
                                                 unlocked and would be clobbered"
state after both inductions: stamp table absent, 59 still owned  (nothing leaked)
```

## 3. NOT DONE, and DO NOT SHORTCUT IT — the 20 tool pages

**Flipping these to `'generic'` destroys 20 working calculators.** The reasoning,
because it is not obvious from the flag's name:

- 19 of the 20 are **single-component verbatim** pages: ONE `page_components` row,
  `slot_name='ported-page'`, holding the entire page including the calculator's
  inline `<script>`.
- The three composition loops run
  `assemble_page → deploy_page (git_commit) → save_sections → update_page_status`.
  `assemble_page` is fed freshly LLM-written HTML and `deploy_page` **commits it to
  the sites repo, which the site deploys from — one step BEFORE** the database
  guard refuses. `owned_page_guard.go`'s header documents exactly this ordering.
- So `rebuild_policy='owned'` is the only thing standing between these pages and a
  generic rebuild that replaces the calculator with prose. This is the vonc arena
  clobber (TL-001) that migration 164 was created to prevent.
- Two of these calculators (`mortgages/stamp-duty`, and the `loans/*` family) are
  the ones `bugs_open/225` and `bugs_open/224` were fixed and oracle-verified on
  over 2026-08-08/09/10.

### The 20, with their current shape

```
loanandmortgagecalculator.co.uk  (17)
  loans/application-tracker            ported-page   (class C, localStorage)
  loans/car-finance-calculator         ported-page
  loans/consolidation                  prose-0,tool-1,prose-2   <- ALREADY DECOMPOSED
  loans/credit-health-check            ported-page   (class C, wizard)
  loans/overpayment-calculator         ported-page
  loans/standard-calc                  ported-page
  mortgages/affordability              ported-page
  mortgages/bridging-loan              ported-page
  mortgages/equity-release             ported-page
  mortgages/fee-analyser               ported-page
  mortgages/investor                   ported-page
  mortgages/overpayment                ported-page
  mortgages/portfolio                  ported-page   (class C, localStorage)
  mortgages/rate-forecaster            ported-page
  mortgages/repayment                  ported-page
  mortgages/simple                     ported-page
  mortgages/stamp-duty                 ported-page   (regulatory — bugs_open/225)
loancash.co.uk  (3)
  tools/complaint-deadline-calculator  ported-page
  tools/price-cap-checker              ported-page   (regulatory — see §6)
  tools/true-cost-calculator           ported-page   (regulatory — see §6)
```

### The correct route, per page

`loans/consolidation.html` is the worked example — it is already prose + a locked
tool row, which is what "fully editable **and** the tool survives an upgrade"
looks like.

1. **Decompose** the page into prose components + one tool component.
   `decompose_lmc.py` + `load_lmc.py` in this lane already do this for
   loanandmortgagecalculator (RUNBOOK §12 is the command sequence). loancash has
   no such tooling — that is new work.
2. **Lock the tool row**: `lock_type='permanent'`. This is what makes the
   lock-aware DELETE in `save_page_sections_action.go:708`
   (`pageComponentAgentWritableSQL`) spare it — measured 2026-08-06 inside
   `BEGIN…ROLLBACK` with `DELETE 2` as the control proving the statement was live.
3. ⚠ **RE-SLOT the tool row BEFORE flipping the page.** `matchLockedRow`
   repositions a locked row by matching `slot_name` against the incoming section
   name. Our slots are positional (`tool-1`), so a writer's sections never match,
   and `save_page_sections_action.go:855` moves an unmatched locked row to
   `len(sections)+1` — **the calculator lands at the BOTTOM of the page**, under
   all the new prose, with only a `lock_blocked` item raised. Silent unless
   somebody looks at the rendered page. This is a *precondition*, not a
   preservation tactic.
   `[UNMEASURED end-to-end]` — read from the code plus the SQL test above; no full
   writer run has been driven against a live page. **Measure it on ONE page first.**
4. **Then** flip that page to `'generic'`.
5. **Re-run the checks on that page** (§5) before doing the next.

Do them **one page at a time with a check between**, not as a batch. The failure
mode is silent and the blast radius is a live consumer-finance calculator.

## 4. ALSO REQUIRED, and separate — there is no site plan

```
site_plans / site_plan_pages / site_plan_sections  =  0 rows for BOTH sites
```

`rebuild_policy` governs whether the generic pipeline **may** touch a page;
`site_plans` is what it builds **from**. Migration 367 removed the refusal; it did
not create the demand. So the 39 unlocked prose pages are now *eligible* and still
undriven.

Seeding a plan is the next step for "upgradable" to mean anything, and it is not
free: `bugs_open/204` is exactly the trap here — `plan_sections` resolves a
section by NAME/FUNCTION only, so a decomposed page can never be rebuilt from a
plan, and the build path asks the fleet to manufacture junk components. Read 204
before seeding.

## 5. The checks, and their state right now

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
cd $LANE

python3 oracle.py                       # 18 tools, ~55 boundary vectors, ~12 min
python3 invariants.py                   # class C: monotonicity/bounds/round-trip
python3 zero_rate_sweep.py --from-repo /home/ant/projects/sites/<domain>
# ALWAYS in the same session, or a green result is not evidence:
python3 oracle.py --selftest-parse
python3 oracle.py --mutate expectation --tools simple
python3 oracle.py --mutate crosstool --tools stamp-duty
```

Measured 2026-08-10, **after** migration 367:

| check | result |
|---|---|
| `oracle.py` full sweep, loanandmortgagecalculator | **PASS 170 · FAIL 0 · CONV 6** |
| controls, same session | parse OK · expectation 4 FAIL/0 pass · crosstool 13 FAIL/0 pass |
| `zero_rate_sweep` loancash (3 tools) | 0 of 3 affected |
| `zero_rate_sweep` loanandmortgagecalculator | 0 of 6 affected |

The unlock changed no arithmetic, which is expected — `rebuild_policy` does not
affect rendering — and is worth having measured rather than assumed.

## 6. NEW FINDING TO CHASE — loancash.co.uk has no arithmetic oracle, and two of its three tools are REGULATORY

`zero_rate_sweep` cleared all three, but **that is a negative from a detector with
two known blind spots** (deterministic-zero, and accumulator false-positives) and
it says nothing about whether the arithmetic is right. loancash has **no oracle at
all** — the same gap loanandmortgagecalculator was in on 2026-08-07.

Two of the three are the highest-value class, because their right answer is
external and DATED — the same shape as the SDLT bug (`bugs_open/225`), which was a
tax rule 16 months out of date:

- **`tools/price-cap-checker`** — the FCA high-cost short-term credit price cap.
  The page states its own rule in its meta description: *"0.8% per day, a £15
  default fee limit, and 100% total cost cap"*, and its code carries `0.8`, `15`
  and `100%` literals with `breaches.push(...)` messages.
- **`tools/true-cost-calculator`** — same `0.8` literals.

**Do exactly what worked on SDLT**: verify the cap figures and their effective
dates against the FCA's own source (CONC 5A), *not* against the page; write the
oracle from that; drive boundary vectors on the cap edges (exactly 0.8%/day,
£15.00 of default fees, total cost exactly 100% of principal, and one penny either
side of each). A dated regulatory literal sitting in a page with no external check
is the exact shape that produced the £5,000 under-quote.

`tools/complaint-deadline-calculator` is a date calculation (FOS six-month /
six-year rules) — also externally checkable, and a boundary-heavy one.

## 7. Docs updated in this pass

- `NOTES` — 2026-08-10 entry (the mechanism, the induction, the refusal to flip tool pages)
- `RUNBOOK` §14 — the unlock recipe and its danger
- `README_where_we_are` — plain-prose entry for the owner
- `LANDMINES.md` — the clobber trap (unlocking a verbatim tool page)
- `bugs_open/224` — carries the 2026-08-10 seventh-instance block from another session; not edited here

## 8. If you are starting fresh, read in this order

1. this file
2. `REPORT_2026-08-08_arithmetic_validation.md` — what the checks are and why
3. `bugs_open/224` and `bugs_open/225` — the two fixed defect families
4. `RUNBOOK` §12 (decompose/re-voice) and §13 (the arithmetic checks) and §14 (unlock)
5. `bugs_open/204` before seeding any site plan
