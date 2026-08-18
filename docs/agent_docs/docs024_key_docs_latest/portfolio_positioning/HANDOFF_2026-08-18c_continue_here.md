# HANDOFF — ⛔ BUILDS STILL HALTED · two owner decisions open · two new items need the owner's hands — 2026-08-18 late evening

Supersedes `HANDOFF_2026-08-18b_continue_here.md` (still accurate on its own history;
**§1 and §2 below are unchanged from it and remain the gate**). Everything in §3–§6 is new
since that file. Chassis at HEAD (`0b185bad2`, both pods, verified from the pods' own
`build provenance` line).

## 1. ⛔ THE HALT — unchanged

**Owner, 2026-08-18:** *"Stop the builds until we sort out the classifier and which builder
flow we are using."* Implemented with `sites.locked_at`, which is what
`build-pipeline-trigger.find_dispatchable_site` excludes on. Queued work is preserved.

| site | locked | queued HELD | HITL | failed |
|---|---|---|---|---|
| `adversecreditmortgage.co.uk` (build #1) | ✅ | 41 | 1 | 0 |
| `remortgagecalculator.uk` (pilot) | ✅ | 0 | 15 | 6 |

```sql
UPDATE sites SET locked_at = NULL, locked_by = NULL
WHERE domain IN ('adversecreditmortgage.co.uk','remortgagecalculator.uk');
```
**Do not lift without the §2 decisions.**

## 2. THE TWO DECISIONS — unchanged, but (a) got bigger

Write-up: `DECISION_2026-08-18_two_builder_flows_side_by_side.md`. RFC: `RFC_037`.

**(a) Which builder flow.** Flow A (seeded + hand-written mission, 45–60 min/domain) vs
Flow B (prompt only, ~2 min, but no `evidence_base` ⇒ every claims lane silently no-ops).
Recommendation was **flow B + an automatic seed**.
> **⚠ NARROWED 2026-08-18 evening by owner ruling P11** (`REGISTER_positioning.md`): the
> first `loanzy.uk` build was **cleared**, because *"we shouldn't create accredited finance
> broker sites unless asked"*. An auto-seed supplies the guards; it does **not** stop the
> classifier/strategist adopting a regulated identity in the first place. Flow B therefore
> needs a **prohibition on regulated-intermediary positioning** (or a check that refuses the
> plan) ON TOP of the auto-seed — new work, not costed in the decision doc.

**(b) RFC_037 — the classifier reads the register.** Owner chose option 2. Filed, not built.

**Also settled tonight and worth knowing before you decide (a):** the two flows are ONE
flow, measured at the handler rather than asserted from the script. All three sites produced
identical item types handled by identical agents (`build-site-planner`, `site-design-planner`,
`page-build-handler`, `image-build-handler`), in the same order. `pageflow-builder` is a live
agent and the classifier still writes `recommended_builder: "pageflow-builder"`, but **no Go
code reads that field** outside doc-comment examples — a leftover from the older intake route.

## 3. ⚠ TWO THINGS NEED THE OWNER'S HANDS

**(i) Deploy the www redirect — one command.** `scripts/cloudflare/worker.js` now 301s
`www.<domain>` → apex (committed `407e334fb`). The deploy is blocked for a session by the
permission classifier, which is proportionate: a bad PUT of this one file takes **all 39
zones** down.
```sh
./scripts/cloudflare/deploy_worker.sh scripts/cloudflare/worker.js
```
Already done for you: live worker confirmed to match the repo copy **behaviourally** (both
recent changes probed — `webdesign.co.uk/tools/` 200 proves DGH-012's index.html append;
`remortgagecalculator.uk/<missing>` returns the worker's own "Not found" at 404, proving
132's fix), and `node --check` passed as `.mjs` **with a deliberately-broken control proving
the checker is not inert** (README warns it silently passes broken ESM in a `.js` file).
The branch is unreachable until a zone has a www record, so deploying it alone changes nothing.

Then, and only then, the fan-out — **written, dry-run clean, not applied**:
```sh
./scripts/cloudflare/add_www_redirect.sh            # dry run (default)
./scripts/cloudflare/add_www_redirect.sh --apply
```
It classifies every zone live rather than looping: **24 need the DNS record only** (they
already carry a wildcard `*.<domain>/*` route), **12 need DNS + route**, and **4 are skipped
or they break** — `idea.uk` and `relojistas.com` have no route to the worker at all (a
proxied A → 192.0.2.1 there is a 522 black hole), `relojistas.com` already serves www,
`webdesign.uk` deliberately 302s to webdesign.co.uk. `cookly.uk`/`dartsonline.com` already
301 correctly. `robot-hands.com`/`leopardessconsulting.co.uk` have a www that hangs today —
this fixes them. **Correction to 18b's §4:** "www resolves NOWHERE across all 36 zones" was
too uniform — 8 of 39 have a record, in four different states.

**(ii) The pilot's lender page is stale — say the word.** It serves 2 lenders; the register
now holds 25. It refreshes on the directory publish path, but the site is LOCKED under §1,
so nothing will re-render it. A single page refresh is not a build, but it is a change to a
live site under your halt, so it is your call.

## 4. THE MISSING TOOLS — root cause found, CONFIRMED, not fixed

**`bugs_open/311`.** `remortgagecalculator.uk` served with no calculator because **selection
keys on `section_type` and storage keys on `function`**, and the calculator component has a
`function` and no `section_type`. So the selector cannot find it (→ "generate a new one") and
the writer CAN find it (→ "this is a regeneration") and the field-contract guard then
correctly refuses to overwrite loanandmortgagecalculator.co.uk's live field contract. Three
identical rejections; page built, deployed and served one section short, `status='active'`.

- `090` verdict **CONFIRMED, first iteration** (`8aa2e283-129f-41d1-93a0-6dcacbbabeae`).
- **Not a one-off:** `loans-credit-health-check` was retrying live at 18:02→18:25 tonight for
  `loancalculator.co.uk` and `loanzy.uk`, blocked by the same site's component.
- **Class is 26 wide** (active base `section` components with no `section_type`), + 79
  `tool`-level rows the section selector cannot see at all.
- **An active `tool-mortgage-repayment` component has existed since 2026-05-06.** The pilot
  needed no new component.
- **Fix candidates ranked in the bug file.** The door-closing one — fork instead of collide,
  the convention `deploy_tool_action.go` already uses — is **platform-scope on a shared seam
  every build passes through: council round, and the owner's say-so, before it is written.**
- LANDMINES entry added + verifier dispatched (`c06ad655`).

**Tools are NOT "still being built"** — nothing is pending. The attempt failed three times
and stopped. Separately: tools normally arrive AFTER a build via `tool-suggester` →
`add_tool` → `deploy_tool_to_site` (35 completions across 18 sites); **neither new finance
site has ever had an `add_tool` item.**

## 5. THE DIRECTORY — 2 → 25 lenders (migration 471, applied and verified)

**Why it was thin:** yield is decided by SOURCE SHAPE. All-history per-source yield —
savings got **12 of 13 firms from ONE gov.uk page**; health got all 10 from **two** broker
round-ups; mortgage got 1 firm per source from four single-firm pages. An enumeration page
had never reached the mortgage scrape set (its slots went to ukfinance's largest-lenders
table, **bsa.org.uk's HOMEPAGE not its member list**, and two societies' own pages).

Migration **471**: `max_scrapes` 4→10, `num_results` 10→20, `max_snippets` 5→8; four
enumeration-shaped mortgage queries (BSA member list · adverse credit · buy-to-let · FSCS
protected); and the prompt line calling third-party round-ups weak narrowed — it was the
opposite of the measured history. Sizing checked against `bugs_closed/062`: 85 kB for 4 URLs
measured, so ~210 kB at 10 against a ~1 MB ceiling.

**Result (four runs, ~15 min): 57 candidates, 42 registered, 25 active named lenders**,
including the adverse-credit cohort `adversecreditmortgage.co.uk` had nothing to list
(Bluestone, Kensington, Pepper Money, Vida, Foundation Home Loans, Aldermore, The Mortgage
Lender, United Trust Bank). Review queue took 8 mortgage + 5 savings rejects — the verbatim
gate at a normal rate. Rollback sidecar exists. **⚠ Queries must stay < 200 bytes.**

**The citation queue, since it was asked about:** it holds the machine's REFUSALS — a
proposed fact whose quote was not found verbatim on re-fetch. HITL-terminal at the
`system.internal` pseudo-site, one row per kind, refresh-on-conflict. Five rows open;
"4 items" in 18b meant four LISTS, not four facts.

## 6. Everything else from 18b that still stands

Phase C signed off · first-50 build order approved · build #1 planned 19 pages and both
directory proof points passed · B8/B9/I10 rulings recorded · L9/P10/P11 loanzy · cost
**$3.81/domain text today, $4.83 from 2026-09-01**, images unmeasured.
**Traps not to re-pay:** a parked domain returns 200 on every path (read the BODY) · never
verify a fix by grepping the binary for its commit sha · seeded `banned_claims` fail silently
when double-escaped (probe in Go) · cost measured mid-build reads ~70% low.

## 7. Files of record

`bugs_open/311` (missing tools) · `sql_for_agents/471_*` (+ ROLLBACK) ·
`scripts/cloudflare/worker.js`, `add_www_redirect.sh`, `README.md` ·
`REGISTER_positioning.md` (P11) · `NOTES_portfolio_positioning.md` (evidence, newest at
bottom) · `README_where_we_are.md` (owner's log) ·
`DECISION_2026-08-18_two_builder_flows_side_by_side.md` · `RFC_037` ·
`docs026_concept_register/register/directory-pipeline.md` (DIR-001, intake sizing).
