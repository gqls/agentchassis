# HANDOFF — the fix loop does not converge on layout-intrinsic mobile overflow

> **STATUS 2026-07-25 — STAYS OPEN. A third defect was found IN THE GUARD
> ITSELF, and the "confirmed live" evidence for (b) was miscounted.** Both (a)
> and (b) remain live and pod-verified on the CURRENT images (chassis
> v1.0.1159, browser-runner-adapter v1.0.1159 — greps below). What changed
> today is that the guard's own justification did not survive being re-run.
>
> - **CORRECTED — the "four terminal improve_tool cycles" figure was wrong, and
>   it was the whole live-evidence base for (b).** It is the count of
>   `acceptance-fail` **doc_notes**, not of the **work items** the guard counts.
>   Re-running the guard's own query on 2026-07-25 finds exactly **ONE** row that
>   has ever existed under `acceptance_fail:tool-loot-table-balancer:<site>`
>   (`ce06e06e`, 07-18 10:38, `complete`). At `max_fix_cycles=2` the guard would
>   have counted **1** on the benchmark and **would not have escalated** — on the
>   very case it was built for. "Confirmed live" was never true. Evidence and the
>   exact queries: §"Defect 3" below. Logged in `WRONG_CALLS.md`.
> - **NEW defect 3 — the judge reports work as queued that it never queued.**
>   Both keyed inserts (`improve_tool`, and site-chrome `responsive_fix`) set
>   their "created" flag from `err == nil` while the statement is
>   `ON CONFLICT DO NOTHING` — which returns **no error when it inserts nothing**,
>   because `idx_swi_dedup` already holds an open item for the key. The
>   acceptance-fail note is the loop's own durable record and was explicitly
>   rewritten on 07-20 to be written "from the outcome, not the intent" — but the
>   outcome variable is derived from the wrong signal, so the note can still
>   assert an `improve_tool` item was created when none was, and can overstate
>   the chrome count `routed separately as N responsive_fix item(s)`. **FIXED and
>   committed today; INERT until the next chassis roll**, which is why this case
>   stays OPEN under the `/bugs_closed/` bar (fixed AND live).
> - **The escalation branch is PROVEN LIVE — 2026-07-26, first time ever**, by an
>   induced fault rather than by waiting for the fleet. It raised
>   `acceptance_stuck:tool-drop-rate-tuner:<site>` at `needs_human_review`
>   (handler `human-review`, priority 20) carrying `fix_cycles_spent: 2` and the
>   full `why_escalated`, and raised **NO** third `improve_tool`. The note's
>   escalation line rendered correctly too ("…instead of a 3rd identical
>   attempt"). Full method, evidence and revert: §"Verify" §4. All test state was
>   reverted in the same session. **This retires the last open question on (b).**

> **STATUS 2026-07-21 — both candidates BUILT AND FULLY LIVE; (b)'s escalation
> branch is live-UNEXERCISED (the benchmark self-resolved before it could fire).**
> Read the two corrections below before acting on the original account: the
> diagnosis in §"Root cause" was **half wrong**, and the half that was wrong is
> the half that felt most obvious.
>
> - **(a) drill-down attribution — LIVE and PROVEN.** Shipped v1.0.1135, still
>   in v1.0.1140 (pod-verified). The signal went from the useless `fieldset
>   (419px)` to naming `div.ltb-row-grid` and why it will not shrink, and
>   tool-improver visibly re-aimed: its next attempt root-caused to the GRID
>   ("grid items couldn't shrink") where the two previous attempts had both said
>   "constrain the fieldset".
> - **(b) convergence guard — FULLY LIVE in v1.0.1146** (pod-verified: guard
>   marker + spec-merge form both present). Three commits: `b13238be6` (guard),
>   `b88da540b` (DO UPDATE refresh, not DO NOTHING — refreshes a re-escalation's
>   count), `905fbbeef` (MERGE the spec, `site_work_items.spec || EXCLUDED.spec`,
>   so a human's triage keys survive a refresh). Logic proven by 6 unit tests,
>   each mutation-verified. **The escalation branch has NOT run live**, and the
>   reason is the honest headline: I queued a manual acceptance run on the
>   benchmark 2026-07-21 11:51 to watch it escalate — and it **PASSED** (all 9
>   checks, `mobile-fit@mobile` green). **The benchmark is no longer RED.** Its
>   `.ltb-row-grid` is now `display:flex; flex-wrap:wrap` (was an unshrinkable
>   grid), so it wraps instead of overflowing, and the page re-rendered 11:33 —
>   024's delivery chain (migration 180 + the Go half, v1.0.1144) finally worked.
>   So the run took the all-pass branch and never reached the count/escalate
>   code. The guard therefore rests on unit tests + deployment verification; its
>   live escalation awaits the fleet's next genuinely-stuck tool (see §"Verify"
>   §2 for how to confirm it then). **I mis-set out to prove it on a tool that
>   had already been fixed** — the count (4 historical cycles) told me nothing
>   about whether the tool was *currently* failing, and I did not re-check the
>   live overflow before firing. Council: 3 rounds, all REVISE, no veto; rounds
>   1 & 2 each produced a real fix (above), round 3's run was killed 5s in by the
>   v1.0.1146 roll (fix_plan at 12:15:15, pod restart 12:15:20). Every actionable
>   objection is addressed or accepted-and-deferred (parent_item_id reuse,
>   generic-mechanism audit). Details in §"Fix candidates" (b).
> - **CORRECTION (2026-07-19, from the travelling-docs thread): the
>   non-convergence evidence in this file does not mean what it says.** The two
>   RED re-verifications and the "materially identical" second fix were not
>   proof that the fixer cannot aim. **The live page had not changed since the
>   tool was born** — tool-improver's fix reached `content_components`, and the
>   re-render that would have put it on the page never ran. The loop was
>   re-testing an unchanged page and correctly reporting it still RED. That is a
>   separate defect chain, filed as **`bugs_open/024`** (five defects; its Go
>   half is live in v1.0.1140, migration 180 pending). **So this bug's title
>   overstates its case:** what was proven is that the loop repeats without
>   bound, not that the fixer cannot solve intrinsic overflow. Caught by the
>   thread that went to watch the benchmark go green and found it could not.
> - **Consequence for (b):** the guard is still worth having — a loop that
>   cannot converge should stop and ask a human — but note it would have fired
>   here on a **delivery** bug while reporting a **fixer** bug. That risk is
>   disclosed in the council submission and is the sharpest open question about
>   this change; see §"Fix candidates" (b).
> - **The benchmark tool is still RED and still the benchmark.** Do not
>   hand-fix it (§ candidate c). Another thread has an open `improve_tool` item
>   on it (`216ea5fe`, a 024 proof) — check before firing anything at it.

**Created 2026-07-17 from the travelling-docs / self-verifying-tools workstream** (its
HANDOFF_2026-07-10 T24 has the surrounding context). This is a **loop-quality**
finding, not an outage: the self-verifying loop ran end-to-end and *correctly*
caught that its own automated fix was insufficient — twice. What needs
attention is why the fixer cannot converge, and two candidate improvements.

**Testbed:** the defect lives on a real tool — `tool-loot-table-balancer`,
gamesdesign.co.uk (`site_id e33263f4-74f8-494f-b191-546845dbbddf`), page
`f25dd4d8-6e25-44eb-a021-689d3057d7a3`. It is a ~29px horizontal overflow at a
390px viewport: a genuine but minor mobile defect. **Left live and unfixed on
purpose** — it is the reproducible benchmark for this finding. DB:
`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## What happened (all autonomous, all on claude-sonnet-5)

Tier-4 acceptance on the tool found `mobile-fit@mobile` failing: the page
overflows at 390px, widest offending element **`fieldset` (419px)**, attributed
**inside the tool** → routed to `improve_tool` (correct). tool-improver
(Sonnet 5) fixed it, redeployed, and the loop **re-verified — still RED, same
419px fieldset.** A second improve_tool item was raised; tool-improver ran
again (Sonnet 5, and it loads PLAN+NOTES first, so it *saw its own prior fix
note*) — and produced a **materially identical fix**, re-narrating "constrain
the fieldset." Still RED.

The behavioural tier did its job perfectly both times ("passed checks ≠
working", now demonstrated on the *fixer*). The problem is the fixer cannot
find the real cause.

## Root cause of the non-convergence (two linked defects)

**1. The attribution signal names the widest ANCESTOR, not the forcing
descendant.** The adapter's `HorizontalOverflow` reports the widest element
crossing the viewport edge — here the single `<fieldset>` (419px). But the
fieldset is 419px only *because a descendant forces it there*. The fieldset's
structure (verified in the template, not guessed):

```
<fieldset>
  <legend>…</legend>
  <div id="ltbRows"> … .ltb-row → .ltb-row-grid (display:grid) … </div>
  <div class="ltb-actions"><button id="ltbAddItem">…</button></div>
</fieldset>
```

`.ltb-row-grid` is `grid-template-columns: 2fr 1fr 1fr auto` (desktop),
`1fr 1fr` (mobile). The overflow comes from a grid child's intrinsic width, not
from the fieldset itself — which already carries `width:100%; min-width:0`
(applied by fix attempt 1). *(NB: an earlier CSS-grep guess blamed
`.ltb-summary div { min-width:140px }`; that div is NOT inside the flagged
fieldset — disregard it. Pin the true culprit in a browser: computed widths,
not source CSS.)*

Because the fixer is told "the FIELDSET overflows", it applies generic width
constraints to the fieldset and its direct "inner elements" — which cannot
resolve an intrinsic min-content width deeper in a grid/flex layout. Both
Sonnet-5 attempts did exactly this. This is the **same class** of problem the
workstream already solved once: T15 refined the document-level overflow signal
to name the culprit + scope (tool vs chrome). The next refinement is to drill
*inside* the offending container.

**2. Nothing bounds a non-converging fix loop.** Each acceptance-fail raises a
*fresh* `improve_tool` item (the item's own `attempt_count` stays 0; the
3-attempt cap never engages because each try is a new item). The only thing
gating re-tries is the 7-day Tier-4 verdict cooldown (`source='tool-acceptance'`),
so this will silently re-fail on a weekly cadence, each time producing an
identical insufficient fix, forever — no escalation, no "we have tried this."

> **This half of the diagnosis held up** (defect 1 did not — see the correction
> at the top). Confirmed live 2026-07-20 by the count the guard now runs:
> **four** terminal `improve_tool` cycles against the same criterion
> (`mobile-fit`) on this one tool between 07-17 and 07-18, and **zero**
> `acceptance-run` notes — the tool has never once passed Tier 4, and nothing
> anywhere had noticed the repetition.
>
> > **CORRECTED 2026-07-25 — the figure above is wrong and it was load-bearing.**
> > **Four** is the number of `acceptance-fail` **doc_notes**. The guard does not
> > count notes; it counts **`site_work_items` rows** under
> > `item_key='acceptance_fail:<fn>:<site>'` with `status IN ('complete','failed')`.
> > There has only ever been **ONE** such row for this tool:
> > ```
> > SELECT id, status, created_at, spec->'failing_checks'
> > FROM site_work_items
> > WHERE item_key='acceptance_fail:tool-loot-table-balancer:e33263f4-74f8-494f-b191-546845dbbddf';
> > --  ce06e06e-55e1-4ff6-ade2-c1aeaaba1b9d | complete | 2026-07-18 10:38:37 | ["mobile-fit"]   (1 row)
> > ```
> > Running the guard's own `convergenceAttempts` SQL verbatim against prod on
> > 2026-07-25 returns **0** today (the 07-21 12:46 `acceptance-run` pass note
> > correctly resets the tally) and **1** with only the reset-bound removed — the
> > positive control that says the query discriminates rather than matching
> > nothing. So at `max_fix_cycles=2` the benchmark's real history yields **1**,
> > and the guard **would not have fired** on the case that motivated it.
> > (The second `mobile-fit` improve_tool row, `216ea5fe`, is another thread's
> > 024 proof under the hand-written key `improve_tool_024proof_3862f72f`; the
> > guard scopes to its own key, so it is correctly not counted.)
> > **What caught it:** re-running the query instead of re-reading the sentence.
> > The repetition itself was real and is not in question — only the number, and
> > the claim that the guard had been shown to catch it.

## Fix candidates

**(a) Drill-down overflow attribution (adapter) — BUILT + VALIDATED 2026-07-17.**
`HorizontalOverflow` now descends from the widest offender through the children
that themselves cross the viewport edge, then along that chain names the
outermost layout container (grid / flex-nowrap) as the fix target — else the
deepest crossing leaf — and states why it will not shrink. New fields
`forced_by` / `forced_reason` ride the CheckResult, the fix-ticket detail, and
the improve_tool / chrome specs (`overflow_forced_by`, `overflow_fix_hint`).
**Probed against this exact live tool before shipping:** the signal went from
`fieldset (426px)` to *"the width is forced by `div.ltb-row-grid` [grid layout
(grid-template-columns: 228px 123px) — a grid item is not shrinking; set
min-width:0 on the items or let the grid wrap]"*. **Confirmed root cause:** the
`.ltb-row-grid` grid items keep their default `min-width:auto` and refuse to
shrink below content, so the two tracks (228+123px) + gap + padding exceed
390px. Code: `internal/adapters/browserrunner/run_checks_action.go`
(`HorizontalOverflow` + `evaluateOnPage`), judge threading in
`platform/orchestration/actions/tool_acceptance_actions.go`. Ships on the next
browser-runner-adapter image (+ chassis image for the judge). **Not yet
re-verified end-to-end** — needs the images, then a fresh acceptance run so
tool-improver receives the pointed hint and can target the grid.

**(b) A convergence guard (judge/loop) — FULLY LIVE in v1.0.1146 (guard +
DO UPDATE refresh + spec merge).** `judge_acceptance_results` now counts the fix cycles that
already failed at the criteria failing *now*, and at N (config `max_fix_cycles`,
default 2) raises ONE `acceptance_stuck` item at `needs_human_review` carrying
`why_escalated`, instead of an identical N+1th `improve_tool`. Below the
threshold nothing changes. Code: `platform/orchestration/actions/tool_acceptance_actions.go`
(`convergenceAttempts` + the judge's failure branch), tests in
`tool_acceptance_convergence_test.go`. Commit `b13238be6`.

The count is bounded three ways, and each bound is load-bearing:
- **terminal attempts only** — an open item is the *current* cycle, not a past
  failure. This is not theoretical: an open `improve_tool` item exists on the
  benchmark tool right now (another thread's 024 proof) and counting it would
  escalate a loop mid-flight;
- **only since the tool last PASSED Tier 4** — a green verdict resets the tally,
  so a regression weeks later is a new defect, not an inherited count;
- **only attempts overlapping the criteria failing now** — a fixer that fixed X
  and left Y has not failed at Y twice.

Fail-open by design: a counting error raises `improve_tool` exactly as before,
because fail-closed would turn a transient DB error into a silently dropped fix.
An open escalation is held by `idx_swi_dedup` (`needs_human_review` is not in
its excluded-status list), so the weekly re-verdict leaves the one item standing
rather than piling up duplicates.

**The open question, stated honestly:** this guard stops the loop, and stopping
is only right if someone reads the escalation. `acceptance_stuck` has no handler
— it lands in the generic `needs_human_review` list. If nobody reads that list,
an escalation is a no-op that *also* stops the retry, which is worse for that
tool than the status quo. **Measured 2026-07-20: that list is 302 items deep**
(`SELECT count(*) FROM site_work_items WHERE status='needs_human_review'`). The
admin dashboard's old hardcoded 50-item cap has since been fixed — it pages
properly (`site_admin_handlers.go:520`, `parseBoundedQueryInt`) — so this is a
backlog-depth problem now, not a visibility bug. An escalation arriving as
item 303 is not obviously read sooner than a weekly retry is noticed. Related: on this very benchmark the guard would have
fired on a **delivery** defect (024) while reporting a **fixer** defect. Both
risks are in the council submission (`submission_010_convergence_guard.json`,
correlation `eeeccdaa-f14b-49cb-b11f-06e7f053add8`).

## Defect 3 — the judge reports work as queued that it never queued (found 2026-07-25)

**FIXED, committed, INERT until the next chassis roll.**

`ON CONFLICT DO NOTHING` returns **no error when it inserts nothing**. Both keyed
inserts in `judge_acceptance_results` derived their success flag from `err == nil`:

```go
_, err := params.DB.ExecContext(ctx, `INSERT INTO site_work_items … ON CONFLICT DO NOTHING`, …)
if err != nil { logger.Warn(…) } else { itemCreated = true }   // ← true even when 0 rows inserted
```

`idx_swi_dedup` is `UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND
status NOT IN (<7 terminal statuses>)`, so whenever the previous cycle's item is
still **open**, a failing verdict inserts nothing — and the judge recorded that
as a queued fix. Two consequences, both in the loop's own durable record:

- the acceptance-fail note asserts `improve_tool item created carrying the
  criteria as acceptance_test` when no item exists. This is the exact failure the
  note was rewritten on 07-20 to prevent — its own comment says the note is
  "written LAST and from the outcome, not from the intent … a note claiming a fix
  was queued when the insert missed is the 'trust the status, not the artefact'
  failure this codebase keeps paying for". The note *was* moved to the outcome;
  the outcome variable was still computed from the wrong signal;
- `chromeRouted` (site-chrome `responsive_fix`) counts the same way, so
  `routed separately as N responsive_fix item(s)` can overstate N.

**Evidence, and its limits — stated because this is where the previous "confirmed
live" went wrong.** The live asymmetry is real: 4 acceptance-fail notes, 1 work
item (queries in the CORRECTION above). It is **consistent with** dedup
suppression but it is **[NOT PROOF] of the current code's behaviour**, because
the 07-17/07-18 notes were written by *older* code (`f59590e32`) in which the
`Fix:` line was a **hardcoded literal** printed unconditionally — so those notes
would have claimed creation no matter what happened. The defect in the **current**
code is established by **code reading plus the new unit tests**, not by those
notes. Whether dedup suppression is what actually produced the 1-row/4-note gap
is **[UNVERIFIED]** — a plausible alternative is that the tool had no
`content_components` row for the first verdicts (there is a `needs_content_page`
item for it from 07-17 09:38, and that branch creates no item either).

**Fix** (`platform/orchestration/actions/tool_acceptance_actions.go`): a
`rowsAffected(res, err)` helper (tolerates the nil `Result` of a failed Exec);
both call sites branch on rows, not on the absent error; a new `itemDeduped`
state so the note says plainly *"no new improve_tool item — one for <criteria> is
ALREADY OPEN under this key (the previous fix cycle has not finished)"* instead
of falling through to the "none could be created" default, which reads as a
defect and would send the next reader hunting a broken insert. Tests in
`tool_acceptance_convergence_test.go`: `TestImproveToolNotReportedCreatedWhenDedupSuppressesInsert`,
its positive control `TestImproveToolReportedCreatedWhenInsertAffectsARow`, and
`TestRowsAffectedToleratesFailedExec`. 9/9 in that file pass; the whole
`platform/orchestration/actions` package passes.

**Council: APPROVED round 1** — corr `a5b0eb25-ae09-4755-b818-c2259e4f322b`,
2026-07-25 18:13, "approved with 3 advisory objection(s) — none high-severity"
(13 seats, abstained 4, `unreadable: null`). **The trailer could not be carried**:
the verdict post-dates the commit (`20dc63716`) and this repo is forward-only, so
the 098 report will bucket it as un-reviewed — a known false negative, not a
missing review. Dispositions, all checked rather than waved through:

- *editquality (low): confirm the guarded `created` counter is the `chromeRouted`
  the diagnosis named.* **Confirmed** — `chromeRouted := routeChromeFailures(…)`
  (line ~520) and that function's `return created` (line ~991) are one value.
- *reuse_agent (low): show a search was done for an existing rows-affected
  helper before adding one.* **Done** —
  `grep -rniE "func +[a-z]*rowsaffected" platform/ internal/ pkg/` returns only
  the new one. None existed.
- *bug_historian (medium): the anti-pattern is not shown to be unique to this
  file; `idx_swi_dedup` is shared, so audit other call sites before calling it
  closed.* **Partly done, and the objection is right.** Measured: **43**
  `ON CONFLICT DO NOTHING` sites, **10 files** discarding the `Result` — listed
  and marked `[UNTRIAGED]` in 016b §9. Only the two here were fixed; a discard is
  a defect only where the flag feeds a durable claim or a counter, and triaging
  the other ten is a separate task, not this bugfix.
- *guidelines (medium): the platform's WORK-ITEM DEDUP convention says
  "use DELETE+INSERT, not ON CONFLICT" — the fix leaves the banned pattern in
  place without flagging it as scoped-out.* **Accepted as deferred scope, stated
  here rather than silently.** `ON CONFLICT` is in fact the fleet-wide norm for
  `site_work_items` (43 sites), including `insertWorkItem` itself, so migrating
  one insert would make it the outlier; the reviewer's own note says that if so,
  "the rule is stale/aspirational". Not migrated. **The drift risk the rule
  guards is real but currently retired** — see §"Verify" §3, where the Go
  terminal-status list was checked element-by-element against the live index.
- *guardian (low): rule out consumers pattern-matching on the note's `Fix:` text.*
  **[UNVERIFIED]** — the guard itself counts `site_work_items` rows, not note
  text, and no other consumer was found, but this was not exhaustively searched.

**Note the interaction with (b), which is the reason this matters here.** The
guard counts rows that this same dedup can prevent from existing. A failing
verdict that queued nothing is not a deferred cycle — it is a cycle that will
**never** be counted. That is a second, independent reason the benchmark's four
failures scored 1. Counting terminal rows only is deliberate and still right
(§"bounded three ways"); what is **[UNRESOLVED]** is whether the threshold of 2
is reachable in practice on the cadence real tools fail at. Do not raise the
threshold or loosen the bounds without measuring that first.

**(c) NOT recommended: hand-fix this one tool.** It would clean the trial site
but erase the benchmark and prove nothing. The tool was left overflowing on
purpose. **UPDATE 2026-07-21: the benchmark self-resolved anyway** — not by a
hand-fix but by 024's delivery chain finally rendering tool-improver's layout
change (grid → wrapping flex) to the live page. So this tool is no longer a
valid test subject for either candidate; it passes Tier-4 now.

## Verify after fixing

1. With (a): re-run acceptance on `tool-loot-table-balancer`; the acceptance-fail
   note's routed culprit should name a grid child inside `#ltbRows`, not the
   fieldset; a targeted fix should then turn `mobile-fit@mobile` green.
   **DONE 2026-07-18** — the note names `div.ltb-row-grid` with a reason, and
   tool-improver re-aimed at the grid. The tool did NOT go green, for the
   unrelated reason in the correction at the top (024: the page never re-rendered).
2. With (b): force two non-converging cycles on any tool; the third fail should
   produce `needs_human_review`, not a third identical improve_tool item.
   **STILL UNEXERCISED LIVE, and the benchmark can no longer test it.** The guard
   is FULLY live in v1.0.1146 (pod-verified) and its logic is unit-tested with
   mutation checks, but the escalation INSERT has never run against prod. I tried
   to induce it (manual `acceptance_run` on the benchmark, 2026-07-21 11:51) and
   the run **PASSED** — the benchmark self-resolved (its `.ltb-row-grid` is now
   `display:flex; flex-wrap:wrap`, page re-rendered 11:33, 024's delivery chain
   working), so the judge took the all-pass branch and never reached the count.
   **Lesson banked:** the count (4 historical cycles) says nothing about whether
   the tool is *currently* failing — always re-check the live overflow state
   before assuming a past-threshold tool is still RED.
   - **To verify when the fleet next produces a genuinely-stuck tool** (a tool
     with ≥2 terminal `improve_tool` cycles at a criterion it *still* fails): on
     the next Tier-4 verdict, expect an `acceptance_stuck:<function>:<site>` row
     at `needs_human_review` carrying `why_escalated`/`fix_cycles_spent`, and NO
     new `improve_tool` for that function. The count the judge runs:
   ```
   SELECT count(*) FROM site_work_items w
   WHERE w.site_id=$SITE::uuid AND w.item_type='improve_tool'
     AND w.item_key='acceptance_fail:'||$FN||':'||$SITE
     AND w.status IN ('complete','failed')
     AND w.created_at > COALESCE((SELECT max(created_at) FROM doc_notes
           WHERE subject_type='tool' AND subject_key=$FN AND source='tool-acceptance'
             AND categories @> '["acceptance-run"]'::jsonb),'-infinity'::timestamptz)
     AND jsonb_typeof(w.spec->'failing_checks')='array'
     AND EXISTS (SELECT 1 FROM jsonb_array_elements_text(w.spec->'failing_checks') e
                 WHERE e = ANY(<the criteria failing now>));
   ```
3. **Structural risk of the never-run escalation INSERT — RETIRED read-only,
   2026-07-25.** The main hazard in a statement that has never executed is that
   it cannot execute: a constraint or an ON CONFLICT arbiter mismatch. Checked
   against the live schema without writing anything:
   - `site_work_items.item_type` is plain `text` with **no CHECK constraint**, so
     the new value `acceptance_stuck` is structurally accepted (`\d site_work_items`);
   - every `NOT NULL` column the statement omits has a default, and the ones it
     must supply — `site_id`, `source`, `item_type`, `summary`, `created_by`,
     `pipeline`, `status` — are all supplied literally;
   - the arbiter `ON CONFLICT (site_id, item_key) WHERE item_key IS NOT NULL AND
     status NOT IN (…)` interpolates `workItemTerminalStatuses`, which is exactly
     the 7-element set in `idx_swi_dedup`'s predicate
     (`complete, failed, verified, rejected, wont_fix, unresolved, cancelled`) —
     verified element-by-element against `\d site_work_items`. Postgres normalises
     `NOT IN (list)` to `<> ALL (ARRAY[…])`, the index's form, so it matches; and
     this same list feeds the fleet-wide `insertWorkItem` path that runs daily,
     so the arbiter form itself is continuously exercised. No 42P10 risk.
   - the guard's counting query `convergenceAttempts` was **executed verbatim
     against prod** (read-only) and returns cleanly, with a positive control —
     see the CORRECTION above. Before today neither it nor the escalation had
     ever run in production, because the only acceptance verdict since the guard
     shipped (2026-07-21 12:46) **passed**, returning at the all-pass branch.
4. **Induced-fault run to exercise the escalation end-to-end — RUN 2026-07-26,
   PASSED.** Correlation `044059b4-67e7-49b6-84b2-7ddd34c4795b`, orchestration
   `ee6b451b`, against chassis **v1.0.1159**. The fleet had still not produced a
   genuinely-stuck tool, and waiting for one is what kept this branch unproven
   for five days — so it was induced, on the trial site, without touching any
   live page. **Observed:**

   ```
   item_type        | acceptance_stuck
   item_key         | acceptance_stuck:tool-drop-rate-tuner:e33263f4-74f8-494f-b191-546845dbbddf
   status           | needs_human_review      handler_agent | human-review    priority | 20
   spec.fix_cycles_spent | 2
   spec.why_escalated    | "2 improve_tool cycle(s) since the last passing Tier-4 verdict left
                            convergence-probe still failing; the one-shot fixer is not converging
                            on this defect"
   ```
   and **no new `improve_tool` row** (the only rows under
   `acceptance_fail:tool-drop-rate-tuner:…` remain the two `cancelled` ones from
   07-12/07-16). The acceptance-fail note took the escalation `fixLine`:
   *"NOT auto-fixed — 2 previous improve_tool cycle(s) failed to turn
   convergence-probe green, so this is escalated to human review
   (acceptance_stuck) instead of a 3rd identical attempt"* — which also exercises
   `ordinalSuffix`. The tool's own 9 real checks **passed** in the same run; only
   the injected probe failed, so nothing about this tool regressed.

   **All test state was reverted in the same session** and verified zero:
   seeded rows, the escalation row, the synthetic acceptance-fail note (removed
   so no future `tool-improver` hunts a phantom selector — `load_doc_context`
   feeds the latest 10 NOTES to it), and the PLAN (restored to `fd7c8af9`, 3046
   bytes, `is_current=true`, `superseded_at` NULL). An accurate
   `["verification"]` note was left on the tool in its place.

   **The method, for the next time a never-run branch needs proving:**
   - add a criterion that cannot pass to the tool's PLAN `criteria` block
     (`doc_plans`, `subject_type='tool'`), e.g.
     `{"id":"convergence-probe","type":"selector_exists","selector":"#doesNotExist"}`.
     **Supersede** (`is_current=false` on the old row + INSERT a new one, in one
     transaction — `idx_doc_plans_current` is UNIQUE on `(subject_type,subject_key)
     WHERE is_current`); restore by reversing, so the original body is never rewritten;
   - seed **two** rows shaped exactly as the judge writes them: `item_type='improve_tool'`,
     `item_key='acceptance_fail:<fn>:<site>'`, `status='failed'`,
     `spec->'failing_checks'=["convergence-probe"]`, `created_at > ` the last
     `acceptance-run` note for that tool;
   - fire `087_TRIGGER_tool_acceptance.sh` with `SEND=1 SPEC_FUNCTION=<fn>`.
     **Expect: an `acceptance_stuck:<fn>:<site>` row at `needs_human_review`
     carrying `why_escalated`/`fix_cycles_spent=2`, and NO new `improve_tool`.**
     Because it escalates, nothing is dispatched to tool-improver — no live tool
     is modified, which is why seeding both prior rows is preferable to letting
     real cycles create them;
   - then delete the seeded rows and the escalation row, and restore the PLAN.
   - **Do NOT use `tool-loot-table-balancer`:** it passes Tier 4 now, and its
     07-21 `acceptance-run` note resets the tally to 0 anyway.
   - **Deploy check (already done for v1.0.1146):** discriminating pod-grep on a
     string the change CREATED, not one it merely uses:
     `strings /app/agent-chassis | grep -c "is not converging on this defect"`
     (=1) and `grep -c "spec = site_work_items.spec || EXCLUDED.spec"` (=1 —
     confirms the merge fix, not just the guard).

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` T24 (the run).
- The T15 chrome-attribution precedent (same class of signal-refinement) — same handoff, T15.
- Adapter overflow logic: `internal/adapters/browserrunner/run_checks_action.go`
  (`HorizontalOverflow`); judge routing: `platform/orchestration/actions/tool_acceptance_actions.go`.
