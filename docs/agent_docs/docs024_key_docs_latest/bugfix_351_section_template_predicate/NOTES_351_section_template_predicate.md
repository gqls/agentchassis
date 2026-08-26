# NOTES — `bugfix_351_section_template_predicate`

Running technical record. **Append-only, newest at the bottom.** Evidence, commands, what the
system actually said, and every misstep. The missteps are not an appendix — they are the point.

Created 2026-08-23, late: this lane ran from 2026-08-21 to 08-22 with only a `HANDOFF`, which is
not the standing five. The entries below therefore start at 08-23; the 08-21/22 record lives in
`bugs_open/351` itself and in the handoff, and is not restated here (CLAUDE.md: point at bugs,
do not fork a second account that drifts).

---

## 2026-08-23 — picking the lane back up: is the bug still real?

**It is half real, and the half that is fixed had not been recorded as fixed.** Working through it
in order.

### The fix had rolled and both docs still said "inert"

`bugs_open/351`'s "IMPLEMENTED" section and the handoff's §3 both said the Go change was inert
until the next chassis roll. It had rolled. Verified at the binary rather than at git or at a tag:

```sql
SELECT DISTINCT service, git_commit, max(last_seen_at) OVER (PARTITION BY service, git_commit)
FROM service_binary_capabilities WHERE kind='build' ORDER BY 1,3 DESC;
--  agent-chassis | f5eaabe3342a906b0392f3cb0d77a67765da6955 | 2026-08-23 17:40:25Z
```

```bash
git merge-base --is-ancestor 97c337371 f5eaabe33   # → yes
git merge-base --is-ancestor <a later commit> f5eaabe33   # → no   (the control)
```

The `build provenance` startup line was **not** in `--tail=3000` on either pod, exactly as
CLAUDE.md warns; the RFC_040 table is what answered it.

### MISSTEP 1 — I started to date a PAST event from a table with a two-hour retention window

A `needs_new_component:loans-credit-health-check` row was filed at **12:08:49Z**, and I wanted to
know whether the fix was live when it was. The capability table showed the older of its two commits
with `min(started_at) = 10:53Z` — before 12:08 — and that commit contains the fix. I was one step
from concluding *"the fix was live and the defect recurred anyway"*, which would have reopened a
bug that is in fact working.

What stopped it: `kubectl get rs` shows a **rollout at 11:51:18Z** whose pods appear **nowhere** in
the capability table. `RetentionWindow` is `2 hours` and ephemeral job pods are *designed* to age
out, so the table is a **window, not a history** — it cannot speak to 12:08 at all, whatever
`started_at` says. Filed as a `LANDMINES.md` entry (no symptom; the rows are all real and correctly
dated, which is what makes it dangerous) and referenced from the bug file.

The right resolution of the 12:08 row turned out to be simpler and is below: the same page bound the
incumbent at 14:23, so the item was superseded, not evidence of failure.

### MISSTEP 2 — a line grep counted a CATEGORY as a LEVEL, and the wrong number was plausible

Counting the exported corpus by a sentinel header `@@@BEGIN:<id>|<level>|<category>|<function>`:

```bash
grep -c '^@@@BEGIN:.*|section|'   # 148
grep -c '^@@@BEGIN:.*|tool|'      # 132   ← wrong, the answer is 129
grep -c '^@@@BEGIN:'              # 277
```

148 + 132 = 280 ≠ 277. Three section-level components have `category='tool'` and were counted
twice. Caught only because I had printed the total in the same breath. `awk -F'|' '$2=="section"'`
asks the question I meant and cannot make the mistake. Logged in `WRONG_CALLS.md` with the general
form: *a measurement whose wrong value sits in the same range as its right value cannot be checked
by reading it.*

### Re-calibration against the live corpus — the flip SET held

Isolated tree (`git archive HEAD`, **on disk, not `/tmp`** — another lane found that tmpfs at 100%
the same day), throwaway `zz_tmp_351_calib_test.go` inside `platform/orchestration/actions` because
the predicates are unexported, deleted with the tree:

```
read: section=148  tool=129  calculators=22     (all asserted non-zero — the vacuity guard)
sectionTemplateValid   rescued=22  regressed=0
endsCleanly flips = 2   SET = {3f946437 case-studies-grid, 6c41404d about-commercial-block}
calculators FAILING the live predicate = 0      calculators with unbalanced markup = 0
```

Mutation controls in the same harness, so the assertion could have failed: a real mid-tag cut, a cut
immediately after a complete mid-template action, and a bare `}}` suffix are all still refused;
nested trailing `{{end}}` wrappers are accepted.

**The corpus moved again** — 150/124 on 08-22, 148/129 on 08-23. Re-running was not ceremonial.

### DEMAND PROOF — the bug file's own closing condition, met

The condition was *"a site planning a calculator section that RESOLVES to a library incumbent with
no `needs_new_component` item filed at all"*.

| bound (UTC) | site | page | component | `section_type` |
|---|---|---|---|---|
| 13:57:41 | loanzy.uk | `tool-is-a-loan-right-for-me` | `loans-damage-checker` | NULL |
| 14:07:15 | loanzy.uk | `tool-eligibility-checker` | `loans-credit-health-check` | NULL |
| 14:23:29 | loanzy.uk | `tool-credit-health-check` | `loans-credit-health-check` | NULL |

All three incumbents were born on `loanandmortgagecalculator.co.uk`, so this is cross-site reuse of
the library — the thing that was impossible before. No `needs_new_component` was filed for any of
them. **Attribution checked rather than assumed:** `824e3309`'s `html_template` was last written
**2026-08-20**, so the data did not move under us; the code did.

---

## 2026-08-23 — two findings the bug file does not contain

### A. `usage_count` counts ONE of the two resolution paths, and the selector scores on it

This started as "why do all 22 incumbents read `usage_count = 0` when three are bound to live
pages?" and ended somewhere more general.

`IncrementUsageCount` has **exactly one non-test caller** as of **2026-08-23**
(`grep -rn "IncrementUsageCount" platform/ internal/ pkg/ cmd/`):
`plan_sections_action.go:1957`, inside `resolveSectionComponent` — which is **Path 2 only**, the
`section_type` selector. A component resolved by **Path 1** (direct `function`/`name` match,
`plan_sections_action.go:1258`) is bound to the page and **never counted**.

And `usage_count` is a scoring input on the path that does count it —
`component_selector.go:181` and `:235`, both queries:

```sql
+ LEAST(COALESCE(usage_count, 0)::float / 50.0, 1.0) * 0.1
```

with the file header calling it *"battle-tested components score higher, with diminishing returns"*.

`[MEASURED 2026-08-23]` over active, non-forked, `component_level='section'` rows:

| | count |
|---|---|
| have any `usage_count` at all | **12** of 149 |
| `usage_count = 0` **and** ≥1 live `page_components` binding | **96** of 149 |
| page bindings invisible to the scoring term | **1,802** |

So the "battle-tested" term sees 12 components' worth of history and is blind to 1,802 bindings. It
is not merely noisy — it is **systematically** biased towards whatever was resolved by
`section_type`, because that is the only thing it can see. A component's score reflects the *path*
it was reached by, not its merit.

**Why this matters for the residual and not just as trivia:** it is a direct argument against the
"just backfill `section_type`" candidate. A backfill would put the incumbents (`usage_count = 0`,
because Path 1 is all they have ever had) into Path 2 scoring **against their own diverted twins**,
which carry 1–4 — a number that measures nothing but which path resolved them. The selector would
prefer the site-suffixed twin over the generic incumbent, on evidence that is an artefact.

`[UNMEASURED]` whether the 0.1 weight is large enough to flip a real pairing. The other terms
(site-type relevance, page-type relevance, quality, specificity) may dominate. **That is the query
to run before anyone acts on this**, and it is not run yet.

### B. A `section_type = function` backfill appears to be a NO-OP for selection

Reasoned from the code, and the reason it is worth writing down is that the bug file's open question
assumes the opposite.

`plan_sections_action.go:1258–1300` runs the paths in a fixed order: Path 1 (`components[sectionName]`,
built by `loadComponentSchemas` → `loadSectionComponents`, which matches **`name` then `function`**,
each against the **raw and kebab-normalised** form, `v3_site_actions.go:4958`), then Path 2
(`resolveSectionComponent` → `SelectComponentByType`, `WHERE section_type = $1`), then Path 3.

If a backfill sets `section_type := function`, then the only key by which the incumbent becomes a
Path-2 candidate is a string that Path 1 **already resolves** — and Path 1 runs first. So the
incumbent can never actually be *reached* through the new key. The backfill adds a row to a query
whose answer is never consulted for that string.

**Consequence:** the ordering question the bug file left open ("which wins depends on path order")
may be the wrong question. The real one is whether these components should answer to a **more
generic vocabulary term** than their own function name — which is a taxonomy decision, not a
mechanical backfill.

`[INFERRED]` — this is read off the code, not demonstrated by running it. The disconfirming
experiment is stated in the plan; do not quote this paragraph as measured.

---

## 2026-08-23 (later) — putting a number on the damage: 25 of 30 filings were avoidable

Over the **whole history** of `needs_new_component` (first row 2026-08-05, last 2026-08-23),
`[MEASURED 2026-08-23]`:

| | count |
|---|---|
| items ever filed | **30** |
| whose `section_type` exactly matched a live component's `function` | **27** |
| …and that component carried `section_type IS NULL` | **25** |

Each of those 25 is a paid LLM generation for a section the library already owned and could have
resolved by name on Path 1 — the platform commissioning a second copy of its own work.

**The obvious way this measurement could have lied, and the control that rules it out.** The
matching component might have been created *by* the item it appears to indict, which would make the
join circular and the number meaningless. Re-run with `AND c.created_at < w.created_at`:

```
30 | 27 | 25      -- identical
```

Identical, so every match genuinely predated its item. (The twins do **not** contaminate this: their
`function` values are site-suffixed — `loans-credit-health-check-loancalculator-co-uk` — so they
cannot satisfy `lower(function) = section_type`.)

`[CAVEAT]` `is_active` is evaluated as of today, not as of the item's date, so a component that was
inactive when the item was filed and activated later would be counted wrongly. Not checked — it
would move the number by at most a couple either way and does not change the shape.

**This makes a disconfirmable prediction, which is the point of recording it rather than admiring
it:** if the predicate fix is doing what we think, the *rate* of `needs_new_component` filings whose
`section_type` matches a live function should fall to near zero from 2026-08-23 onward. It has not
been long enough to test. **Whoever next picks this up should run the query above windowed on
`created_at > '2026-08-23'` — a continued high rate refutes the fix, and that is the cheapest
available way to be wrong about it.**

It also sharpens finding B above: 27 of 30 filings named an **exact function match**, so Path 1 is
the route that actually matters for this population and Path 2's `section_type` key is close to
irrelevant to it. A `section_type := function` backfill would be adding a key to the path that was
never the bottleneck.

---

## 2026-08-23 (later still) — the council round, and three things I got wrong

### Council `7b662d65`: APPROVED round 1, and the objections were worth more than the verdict

Both advisories were acted on rather than banked. Recorded in full in `bugs_open/351`; the part
worth repeating here is that **the medium-severity objection found a hole in my own calibration that
neither I nor the submission had seen.** `bug_historian` asked whether every
`componentTemplateValid` dispatch arm had been audited. There are only two — and because the second
is a **default** rather than a `section` case, **12 active rows at other levels
(`site` 6, `header` 4, `footer` 1, `element` 1) took the changed predicate and had never been
calibrated.** Measured: `read=12, rescued=6, regressed=0`, and six site-level headers/footers turn
out to have been wrongly dropped all along.

**The lesson is not "audit the arms".** It is that *I calibrated the populations the bug was about,
not the populations the changed function serves.* The blast radius of a predicate is its callers'
inputs, not the rows that motivated it — and a `default:` arm is the easiest way for those two sets
to differ without anything looking wrong.

### MISSTEP 3 — I answered "which path resolved it?" by inference when a positive signal was available

`editquality` objected, correctly, that Path-1 resolution was *asserted, not shown*. I had marked it
`[INFERRED]`, which is honest, but the marker was doing work a query could have done: the selector's
`IncrementUsageCount` side-effect gives a **positive** signal, not merely the absence of one, and it
was two minutes away. **Marking a claim unverified is not a substitute for verifying it when the
check is cheap** — the marker is the floor, not the ceiling.

### MISSTEP 4 — my census of the manual route answered a slightly different question than I asked

I gave the planning agent a date histogram of `created_from='manual'` section-level births
(08-15 ×12, 08-14 ×2, 08-13 ×14, …) as if it described **the 25 NULL rows**. It does not — it is the
histogram of **all** manual section births and sums to 35. The 25 NULL rows' own histogram is
08-13 ×13, 08-14 ×2, 08-15 ×6, 07-28 ×2, 07-31 ×1, 08-02 ×1 (as of 2026-08-23). Caught by the agent,
which reconciled the sum against the population — the **same** check that caught misstep 2 earlier
today, now two for two. Nothing was published with the wrong figure; the claim that reached the docs
("all 25 are manual, last manual write 2026-08-15") is correct.

### The planning agent was wrong about one thing, and it is the thing I had just filed a landmine on

It reported that `service_binary_capabilities` *"has no rows at all before 15:30Z on 08-23 (the
table's writer is that new)"*. The emptiness is real; **the explanation is not** — the writer has run
since RFC_040 on 2026-08-20, and the rows had simply been **pruned**. I had cited
`min(started_at) = 10:53Z` in this lane's own landmine only ~25 minutes earlier, and by 18:09Z that
row was gone, with `min(last_seen_at)` sitting exactly on `now() - 2 hours`.

**So two readers drew opposite conclusions from the same emptiness on the same afternoon, and both
were wrong in the same way** — treating a pruned window as evidence about the past. That is a better
argument for the landmine than anything I wrote in it, and it is now recorded inside the entry.
A subagent's report is another doc: its figures were sound, one of its *explanations* was not, and
the two are not distinguishable by tone.

### Housekeeping: my landmine correction was swept into another session's commit

`d0930af6f` (the `326` lane, pathspec on `LANDMINES.md`) took my uncommitted one-line correction as
a **same-file passenger** — roughly ten minutes after I read this bug file's own implementation note
warning to expect exactly that. Nothing is lost: the text is committed and correct, just under
someone else's message. Recorded because the file's warning is evidently not folklore.

---

## 2026-08-24 — the birth guard: migration 581 / CLC-029, written and HELD

The owner chose to close the birth door (option 1 of four put to them). Written, tested against the
live DB, council-submitted, **not applied**.

### What it is, and the one line that decides its shape

```sql
IF NEW.component_level = 'section' AND NEW.forked_from IS NULL AND NEW.section_type IS NULL
```

`BEFORE INSERT` on `content_components`, `ERRCODE 23514` (the class the table's existing kebab CHECKs
already use). **It is a trigger and not Go because the producer is unreachable from Go**: all 28 such
rows ever born are `created_from='manual'`, the `generated` route has produced zero, and the three Go
INSERT sites are each provably outside the predicate.

### Three scoping decisions, each verified rather than reasoned about

1. **`forked_from IS NULL` is load-bearing and looks like decoration.** `deploy_tool_action.go:326`'s
   fork INSERT lists 16 columns and **`section_type` is not among them** — so a section-level fork is
   *legitimately* born NULL. Widen the predicate and tool deployment breaks at runtime.
2. **INSERT-only, which is also why it is not a CHECK.** A CHECK — even `NOT VALID` — is enforced on
   UPDATE of pre-existing rows, so every template-repair write to the 25 standing rows would start
   failing.
3. **Not a generated column, not a `COALESCE(section_type, function)` default.** 35 active rows
   deliberately carry a `section_type` that differs from `function`; and `function` DEFAULTs to
   `generic-text-block`, so a silent COALESCE would pour unlabelled rows into the commonest selector
   pool — where a wrong match is least likely to be noticed.

### The verify block INDUCES, and it is mutation-proven

A verify of bare `SELECT`s cannot stop a `COMMIT`, so it is `DO`/`RAISE`: attempt the refused INSERT
and **fail the migration if it is accepted**, then four controls that must all succeed (labelled
section, fork born NULL, tool, and an UPDATE of a standing NULL row — the NOT-VALID-CHECK trap,
induced). All probe rows are written inside a subtransaction deliberately aborted.

Run against the **live** DB with `COMMIT` → `ROLLBACK`:

```
NOTICE:  581 VERIFY: PASS — refusal induced, and 4 controls ... all behaved as required.
NOTICE:  581: 25 active section-level non-forked rows still carry a NULL section_type ...
ROLLBACK
```
then `0` triggers present and `0` probe rows left behind.

**Two mutations, each caught by the assertion that should catch it** — and this is the part that
makes the verify block evidence rather than ceremony:

| mutation | caught by |
|---|---|
| predicate made inert (`IF FALSE AND …`) | the induce probe: *"a section-level row with NULL section_type was ACCEPTED — the trigger is inert"* |
| `forked_from IS NULL` dropped (too wide) | the **fork control**, failing with the trigger's own refusal text naming `function=zz-probe-fork` |

The unmutated file still passes. Mutation 2 is the important one: it demonstrates the fork trap is
real *and* that the control detects it, which is the only reason to trust the scoping.

### MISSTEP 5 — I took migration number 580 and so did someone else, one minute earlier

Wrote `580_refuse_selector_invisible_section_birth.sql` at 11:32; `580_database_cleanup_…` had been
created at 11:30 by another session. Caught by `ls` immediately after writing, and renumbered to
**581**. The number was checked *before* writing and was free then — **on this tree "the next free
number" is a fact with a shelf life of minutes.** Check it again in the same breath as the write, and
`grep` the renamed file for the old number afterwards: mine had **two** stragglers the filename sed
did not reach (a `COMMENT ON FUNCTION` body and an internal `RAISE` tag), and a migration that names
the wrong number in its own error messages is a bad afternoon for whoever reads the log.

> **⚠ ADDENDUM — misstep 5 was ALREADY IN `LANDMINES.md`, three times over, and I did not look.**
> *"The next free migration number is only free until someone commits — two sessions can hold the
> same NNN for hours and the ledger will happily apply both"* (line ~11243), plus two more entries
> at ~3479 and ~12287, one of which names the exact command I used (`ls … | tail`). It cost me only
> a rename, but I had the answer on disk before I had the problem.
>
> **Why I did not see it, which is the transferable part:** the `SessionStart` hook only surfaces
> landmines whose footprint matches a file **already dirty** in the tree, and
> `docs/agent_docs/sql_for_agents/` was clean when this session started. A landmine for a directory
> you are about to write your *first* file into is structurally invisible to that hook. The standing
> remedy is the one in MEMORY — **grep LANDMINES for the path, table or symbol you are about to
> touch, before you touch it** — and "I am about to create a numbered file in a shared directory" is
> exactly the shape that deserves it.

### MISSTEP 6 — the trailer gate stopped me writing a join key of `pending`

I drafted the commit with `Council-Submitted: pending`, intending to fill it in. The `commit-msg`
hook refused: the trailer is a **join key** for the 098 coverage report, a non-UUID resolves to
nothing, and forward-only forbids the amend that would fix it. Correct behaviour, and the right
order is simply the other one — submit first (the trigger prints `SUBMISSION_CORR` in seconds), then
commit with the real id. Recorded because "I'll fill it in after" is a natural thing to type and
there is no second chance at it.

### Fresh chassis build — the fix is still live on it

`v1.0.1332` = `0b262ed5e`, pods started 2026-08-24 09:37Z. `git merge-base --is-ancestor 97c337371
0b262ed5e` → yes, with a control that correctly reports NOT an ancestor. The demand proof of
2026-08-23 was taken on `f5eaabe33`; the predicate half remains live on the successor build.

---

## 2026-08-24 (close) — 581 applied, 351 closed

Council `f0cd2420` **APPROVED** round 1, 3 advisory objections. Two were medium and **both changed
the SQL** — recorded in `bugs_closed/351` and CLC-029, not repeated here. The short version is that
the reviewers found a door I had reasoned about and then not closed (an UPDATE could clear
`section_type`, reproducing the defect through another write path) and a re-run failure I had made
*worse* with my own guard.

**Applied by hand, then recorded the same minute** — `--apply` takes every pending file and two other
sessions' `_HOLD` halves were sitting in the directory. `--record-only` with the provenance in its
`--note`, because a hand `psql` run writes no ledger row and the number stays unclaimed for ever
(the ledger confirmed `581` was unclaimed before I applied, and `580` was **not** mine).

**Proven behaviourally on the live table, which is the whole point of the verify-later:** birth
refusal induced, clear refusal induced on a real row (`faq`), a repair UPDATE to standing row
`824e3309` **succeeded**, a labelled birth **succeeded**. All rolled back.

Post-check: `pg_trigger` shows `trg_cc_refuse_null_section_type` enabled (`tgenabled = 'O'`), and
`schema_migrations` carries the row.

### The one thing I would tell the next person

**Five of the six missteps in this file are the same mistake wearing different clothes: a number or a
result that could not have come out otherwise.** The vacuous `0/22`, the flip *count* that could not
distinguish a rescue from a substitution, the field-blind grep whose wrong answer was in the
plausible range, the pruned table that answered a question about the past with the present, and the
census that reconciled only against itself. The council's medium objection was the same shape from
outside: I had calibrated the population the *bug* was about, not the population the *changed
function* serves.

The habit that actually caught things here was not care — it was **arranging for a wrong answer to
look different from a right one**: printing the total beside the breakdown, asserting the flip SET
instead of its size, running a control that had to fail, and inducing every refusal rather than
asserting a guard exists.

## 2026-08-26 — the lane resumed once more: 541 discharged, and nothing now remains

Session resumed on the closed lane (owner named it `remortgagecalculator.co.uk` — that domain does
NOT exist; `sites` holds only `remortgagecalculator.uk`, the portfolio_positioning lane recorded the
same misremembering on 08-19). The one debt in the CLOSE banner — migration `541` — was found
UNBLOCKED and was discharged today. Evidence trail, controls inline:

- **Roll condition:** every live chassis pod (217/217 seen within 30 min, ONE distinct build commit
  `2fb40a960`) carries `e34b33a36` (the check's birth) as an ancestor; control — today's HEAD is
  correctly NOT an ancestor.
- **Capability condition, superseding the exec-grep recipe:** the binary now self-reports checks in
  `service_binary_capabilities` (`kind='discovery_check'`) — `stylesheet_gutted` on 217/217 live
  pods, negative control `stylesheet_gutted_NOTREAL` = 0 rows.
- **Re-calibration, with the check's OWN code** (the 08-21 proxy misstep is why this is stated):
  exported the corpus with the check's own predicate SQL (shared builders inlined verbatim from
  `datahelpers/links.go`), then drove the real `Run()` via a throwaway in-package test — sqlmock fed
  the real rows, `fetchSiteStylesheet` left UNstubbed so fetches were live HTTP. **0 filed /
  29 resolved / 2 declined-to-judge over all 31 deployed sites** (corpus was 25 on 08-21). The two
  declines: lampenkap.com's linked stylesheet 404s (asset_reference_404's finding, not ours);
  loanandmortgagecalculator.co.uk sheet serves 200/16,277 B so the skip branch is one of
  over-cap/external/other-sheet-failed — `[UNVERIFIED]` which, and it files nothing either way.
  Throwaway test deleted in the same command that ran it.
- **Applied** 09:07Z by hand (`UPDATE 1`, DO/RAISE verify passed, COMMIT); live row independently
  re-read: 24 checks, `stylesheet_gutted` present. Ledger row `541_enable_stylesheet_gutted_check.sql`
  record-only; checksum re-synced after the discharge-header edit (the ledger stores content md5, so
  header edit → checksum update, in that order). Files renamed to drop `_HOLD` (475 precedent), name
  added to `liveConfiguredChecks`, all one commit `6531e694b`, `Council-Reviewed: d3187418` (verdict
  re-read from `doc_notes` today before writing the trailer).

**Missteps, this visit:**
1. My corpus export loop ran ONE iteration and exited 0 — `kubectl exec -i` inside `while read`
   consumed the loop's stdin. Caught immediately because the script printed a per-site line and
   asserted the site-list count first. The class is already covered (`check_stdin_eater` in
   `pattern-check.py`; LANDMINES has the meta-entry) — it could not fire here because the script
   lived in the scratchpad, outside the pre-commit's reach.
2. My calibration harness printed zap Int fields as `<nil>` (observer `Field.Interface` is nil for
   ints), so the two SILENT sites' skip counters were unreadable and one skip branch stays
   `[UNVERIFIED]`. Cosmetic, but it is why the loanandmortgagecalculator question is open above.

**What the next reader should know:** `go test ./platform/orchestration/actions/` is RED at HEAD
today for a reason that is NOT this lane's — `WORK_ITEM_STATUS_OVERRIDE_REFUSED` (commit
`2b46afbe6`, bugs_open/396 lane) is undeclared in the finding-code registry. Two other lanes have
already noted it in `bugs_open/396_…undispatchable…`; do not re-report it, and do not mistake it for
a 541 failure — `discovery_checks` and the filtered registration/gutted tests pass.

### Same day, 13:xx — the FIRST LIVE EXERCISE, verified at the artefact within the hour

Timing gift: the design rotation (`site-discovery-rotation-design`) was re-enabled at **09:20Z**
after 15 days off (the 08-11 cost-scare pause, `bugs_open/401`) — thirteen minutes after 541
applied. Its first visit, agritec.uk, ran orchestration `18fe7caa` at 09:20:38Z, and I verified the
row first-hand rather than trusting the peer report that flagged it (webdesign-tool-rebuilds seat):
`discovery_result.checks_run` = **24 names with `stylesheet_gutted` present (last, the appended
position)**, `checks_failed` = **[]**, `failed: 0`, run COMPLETED, `error` NULL. So the
unregistered-name failure mode is disproven in production, not just at the capability registry, and
the zero-findings calibration held on the first real visit. `remortgagecalculator.uk` carries NO
design stamp (checked `site_discovery_rotation`) so it is in the rotation's first six — expect its
first-ever design visit within ~18 h of the re-enable.

Two cautions from the same hour: (1) my first read of the run used `jsonb_pretty | grep`, and grep
ADJACENCY put `stylesheet_gutted` under `checks_failed` — extracting the arrays with `#>>` showed
`checks_failed` empty. Read the array, not the neighbourhood. (2) The peer session initially
reported the check "had already run on several sites" overnight via improvement-loop design children
(e.g. webdesign.co.uk 03:46Z) — those runs PRE-DATE the 09:07Z enable, so their zero
`stylesheet_gutted` items is VACUOUS (23-check roster; the guard `NOT … ? 'stylesheet_gutted'` on
the 541 UPDATE proves the name was absent until 09:07). Corrected with the peer so their NOTES do
not bank it as coverage.
