# BUG 449 — no acceptance fence the tool-generator writes ever asserts a NUMBER, so a calculator that computes garbage passes Tier 4

**Filed 2026-09-02** by the `mortgagecalculator_couk_adoption` lane, answering the owner's "verify
the tools" — the answer being that where the platform's verification runs at all, it is not checking
what the tools are for. **Status: OPEN. Nothing changed, nothing dispatched.**

---

## 1. The claim

The acceptance runner implements a check type, `computed_values`, that drives a tool with fixed
inputs and asserts the **exact text** of each output (`run_checks_action.go:708`, `:809`). It works —
this site's `simple` calculator passes four of them. **No agent that authors fences has ever been
told it exists**, so no generated fence uses it, and the generated fences that do exist check that
the page loads, logs no console errors, fits a phone, and produces *an element* after a click.
**A calculator that renders a confidently wrong figure passes every one of them.**

## 2. Measured 2026-09-02, live DB

| fences by author | count | assert **no expected value at all** | use `computed_values` |
|---|---|---|---|
| `tool-generator` | **170** | **107 (63%)** | **0** |
| `operator:bugfix224-session` | 16 | 0 | 16 |
| `operator:mortgagecalculator-lane-a4` | 8 | 0 | 8 |
| `operator:staged_component_build` | 8 | 0 | **6 — and 6 also carry `interaction`** |
| `webdesign_couk_thread` | 14 | 4 | 0 |

Two things that keep this honest:

- **The claim is NOT "generated fences never assert anything".** 63 of the 170 use
  `interaction.expect.text_matches`, and some of those patterns are real values (`\$1000\.00`,
  `40.0%`, `Checkout for 121`). ⚠ **My first version of this finding said "zero assert a computed
  value" and that was too strong** — it was true of the check *type* and false of the behaviour.
  The accurate figure is the middle column: **107 of 170 assert no expected value anywhere.**
- **`operator:staged_component_build` is the existence proof.** 6 of its 8 fences carry
  `computed_values` **and** `interaction` — correctness and health in one fence. So this is not a
  limitation of the format or the runner; it is a gap in what the authors are told.

**The two check families are otherwise disjoint**, which is the sharp version of the problem:

| family | asserts | blind to |
|---|---|---|
| `computed_values` (operator-authored) | the numbers are right | whether the page loads, errors, or fits a phone |
| `interaction` + `no_console_errors` + `no_horizontal_overflow` + `page_status_ok` + `selector_exists` (generated) | the tool is alive and responds | **whether any number it prints is correct** |

## 3. The cause, and it is one place

```sql
SELECT type, (default_config::text LIKE '%computed_values%') AS knows_computed_values
  FROM agent_definitions
 WHERE default_config::text LIKE '%```criteria%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

| agent | knows `interaction` | knows `selector_exists` | knows `no_console_errors` | **knows `computed_values`** |
|---|---|---|---|---|
| `tool-generator` | yes | yes | yes | **NO** |
| `experience-planner` | yes | yes | no | **NO** |

Both fence-authoring agents were taught the liveness vocabulary and neither was taught the
correctness one. The 0-of-170 is not a modelling failure or a hard case — **the type is absent from
the prompt**, so it is never a candidate.

## 4. What this means for a "PASSED" verdict — read before quoting one

A Tier-4 PASS is only as strong as its fence, and the two families make opposite promises:

- `tool-overpayment-priority` and `tool-rate-scenarios` (this site) are **PASSING**. Their fences
  carry no value assertion of any kind. The verdict means *the page loads and something appears when
  you click* — **it says nothing about the arithmetic.**
- `simple` (this site) is **PASSING** on four `computed_values` checks and nothing else — no boot
  check, no console check, no status check, and every check pinned to `profiles: ["desktop"]`, so
  its doc-level `["desktop","mobile"]` produces a mobile pass in which **all four checks are
  skipped**. The verdict means *four sums are right on desktop.*

**Not one tool on this site is verified for both correctness and health**, and the same will be true
of any site whose fences came from `tool-generator`.

## 5. Fix candidates

1. **Teach both authoring agents the type (recommended, and small).** Add `computed_values` to the
   criteria vocabulary in `tool-generator` and `experience-planner`, with the instruction that a
   tool which computes anything must assert at least one worked example. `staged_component_build`'s
   6 fences are the template to copy.
   ⚠ **The generator must not invent expected values.** The RUNBOOK §14 rule for this site's own
   fences applies fleet-wide: *an emitted value is not an expected value*. A generated
   `computed_values` check whose expectation is whatever the tool printed at birth pins the bug
   along with the behaviour — `run_checks_action.go:775-781` says so in the code that does it. The
   generator can propose the inputs; the expectation needs a source that is not the tool.
2. **Report the gap rather than assume it.** A `needs_criteria`-style note when a fence for a
   computing tool carries no value assertion. Cheap, and it makes the 107 visible without anyone
   having to re-run this query.
3. **Backfill the 107.** Real work, tool by tool, and it needs (1) first or the next generated fence
   re-creates the gap.

## 6. How to verify a fix

- **Induce the red, which is the whole point here:** take a passing tool, change one constant in its
  JS so it computes a wrong figure, re-run acceptance. **Today it still passes.** After the fix it
  must fail. A fix that only adds checks which pass is indistinguishable from no fix.
- Re-run §2's query: `assert_no_value` for `tool-generator` must fall for **newly created** fences.
  Compare by `created_at` window, not by total — the 107 existing ones do not change themselves.

## 7. Related

- `bugs_open/441` — the fences that DO assert selectors are naming pre-conversion ids on
  instance-scoped tools. 441 is "the check cannot find the element"; 449 is "the check never asked
  whether the number was right". Independent, and both must be fixed for a PASS to mean much.
- `bugs_open/126` — gated-tool acceptance / `no_auto_fix`.
- RUNBOOK §14 of `mortgagecalculator_couk_adoption` — the emitted-vs-expected discipline, and
  `verify_criteria.py`, which re-derives every pinned value from a non-page source at three labelled
  strengths. That is the machinery candidate 1 needs and it already exists.

---

## 8. RESUMED 2026-09-03 by the `bugfix_449_fences_assert_no_number` lane — P1+P2 SHIPPED, cause deliberately untouched

**Status: OPEN, and it stays open.** The two halves below are committed (`0b9a5c9e1`) and **inert
until the next chassis roll** — they are Go. The *cause* (neither authoring agent knows the type)
is unchanged, so this bug is still reproducible and does not move to `bugs_closed/`.

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_449_fences_assert_no_number/`
(PLAN / NOTES / RUNBOOK / README). Ownership: the mcalc lane keeps the **site** half (its 8 fences,
`441`, `448`, `install_fences.py`, `verify_criteria.py`); this lane took the **framework** half.
Agreed in `CONTRIB_2026-09-03_from_the_449_lane_…` in their directory — **not yet acknowledged**.

### 8a. Re-verified, and it is WORSE — this is a live intake, not a backlog

§2 was measured 2026-09-02. Re-run 2026-09-03 (query in the lane RUNBOOK §1b):

| | 09-02 (§2) | 09-03 | |
|---|---|---|---|
| `tool-generator` fences | 170 | **186** | +16 in ~24h |
| …asserting no value at all | 107 | **115** | +8 |
| …using `computed_values` | 0 | **0** | unchanged |

`max(created_at)` is **today**. Every hour the fleet runs mints more.

**Two columns §2 did not have, and they are the sharper cut:** **91** of the 186 **drive inputs**
(a `fill` or `select` step) and **55 of those assert no value of any kind.** That subset is
identifiable *from the fence alone* — no classifier for "is this a calculator", which matters
because a guarantee conditional on a classifier inherits the classifier's gaps.

⚠ **A trap on §3's own query.** Both authoring agents' `updated_at` now reads
`2026-09-03 08:56:53.045885+00`, which looks like "the prompt was revised this morning". It was
not: **208 rows share that exact second** — a bulk touch. A timestamp is not a diff.

### 8b. §5 candidate 1 is UNSAFE as written, and the ordering had to invert

The bug's own ⚠ is right and I have promoted it to a decision. `computed_values` is a **regression**
check by construction — `runComputedValues`' docstring says the values *"are CAPTURED from the tool
while it is known good … and then defended"*, and that *"a golden captured from an already-wrong
tool pins the wrong answer"*. At generation time there is no known-good state, and the generator is
handed `{{.generated_html}}` — the tool's own code — so any expectation it derives **shares a
failure mode with the implementation it is meant to police**. That is `224`/`225`: an expired £625k
FTB cap certified green for sixteen months.

**And a second reason, from `441`, which its own lane re-framed today as a LIVE GENERATOR of stale
fences.** `runComputedValues` **fails rather than skips** on a missing element, by design. So
teaching the generator to assert values while 441 is live converts *silent blindness* into *loud
false failures*, aimed at `tool-improver`. Strictly worse.

**So the honest-record halves ship first and the authoring change ships last.** A pinned value the
generator could not derive is worse than none: blindness is uninformative, a defended wrong number
is believed.

### 8c. What is now in HEAD (commit `0b9a5c9e1`, `Council-Submitted: 8745ad9e-…`)

**P1 — the verdict states the scope of its own claim.** This is §4 made machine-readable. The
Tier-4 PASS `doc_note` gains a `Scope of this verdict:` line, and the action result gains
`assertion_grade` (`none` | `pattern` | `exact`), `value_assertions`, `exact_value_assertions`, plus
`verdict_scope: "liveness_only"` **present only** when nothing was compared, so a consumer branches
on presence rather than parsing a string. **No outcome changes — a passing tool still passes.**
Chosen first because it is the only change that covers **all 186 fences at once, today**, with no
backfill and no author cooperation, and it cannot go stale: the grade is derived from the fence at
*run* time, so a fence that gets weaker gets a weaker verdict with nobody remembering anything.

**P2 — the door records a blind fence where it is born.** `write_doc_plan_action.go` is the **only**
production Go writer of a PLAN body (enumerated, not assumed: `grep` every Go INSERT/UPDATE against
`doc_plans`; three live agents reach it). A fence that drives inputs and asserts nothing now writes
a `fence_asserts_no_value` doc_note naming the author. This is §5 candidate 2, placed at the door.
**It RECORDS, it does not REFUSE** — a tool with *no* PLAN is inert at **both** tiers, so refusing
today trades a blind check for no check. Refusal is written up as a deferred P5 **with its opening
trigger stated**, so it does not become an "inert until X" line nobody revisits.

⚠ **Three grades, not two.** §2 corrected itself on exactly this ("my first version said zero
assert a computed value and that was too strong"). A hand-authored `text_matches` is real evidence
*and* weaker evidence — `/£[\d,]+\.\d\d/` is satisfied by the £0.00 an unwired tool prints — so it
grades `pattern`, between the two. Collapsing it either way is a lie in one direction.

### 8d. §6's bar was met by MUTATION, not by green

All three mutations were **run**, not described: (1) delete the judge's scope line → the
liveness-only and pattern tests went RED; (2) delete the door's block → the door test went RED;
(3) credit `computed_values` by TYPE rather than by having expectations → the empty-`expect_values`
subtest went RED. Files then restored byte-identically (`diff -q`) and re-verified green.
Every silence assertion is paired with a **demand control** in the same test — a blind counter that
always returned zero would pass "reports liveness_only" and mean nothing.

`scripts/verify-head-builds.sh --with <4 files> --test` reports **`ok`** for
`platform/orchestration/actions` against HEAD `48bd6c5b6`. (The one failure in that run,
`discovery_checks/TestStylesheetGutted_…`, is **pre-existing at HEAD** in a package this never
touches — the 458 lane's paired-ink work; verified failing with my change absent, and reported to
them in `bugs_open/458` §11.)

**§6's *other* half is still owed and is NOT claimed here:** nobody has yet taken a passing tool,
broken one constant in its JS, and re-run acceptance. P1/P2 cannot make that fail — by design, they
grade rather than strengthen — so that induction belongs to the authoring phase.

### 8e. What is still open, and where it is blocked

1. **The cause.** Both prompts still lack the type. Planned as a `732`-shaped surgical migration on
   `{workflow,steps,compose_plan,config,prompt_template}` — **sequenced after 441**, and its
   load-bearing half is a *refusal* arm: if the generator cannot derive an expectation independently
   of the code it was shown, it must emit **no** `computed_values` check rather than a guessed one.
2. **Where an expected value may come from, in general.** `verify_criteria.py`'s three labelled
   strengths (DEFINITION / REGISTER / CONVENTION) are the honest answer but live in one lane's
   directory, and may only work because mortgages have published formulae and SDLT has a legal
   register. **Asked of the `loancalculator` lane** (owner of `toolgolden.py`) 2026-09-03 — unanswered.
3. **§5 candidate 3 (backfill the 107, now 115).** Not doing, and P1 removes the reason to rush:
   once a value-less PASS reports itself as `liveness_only`, the 115 are honest-but-weak rather than
   false-and-strong. Backfill is then a per-site decision with a per-site oracle, which is where it
   belongs.
4. ⚠ **`no_auto_fix` (LANDMINES §8626).** Tier 2 ignores it entirely and appends three built-in
   shell failures *outside* the criteria loop, so `computed_values` beside the four existing health
   checks is exactly the combination that lets Tier 2 aim `tool-improver` at a **shared** component.
   The authoring phase must answer this. The mcalc lane has been asked for its judgement.
