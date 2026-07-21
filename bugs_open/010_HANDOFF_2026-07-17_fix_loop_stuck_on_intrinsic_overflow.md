# HANDOFF — the fix loop does not converge on layout-intrinsic mobile overflow

> **STATUS 2026-07-21 — both candidates now BUILT AND LIVE; (b)'s live proof in
> flight.** Read the two corrections below before acting on the original account:
> the diagnosis in §"Root cause" was **half wrong**, and the half that was wrong
> is the half that felt most obvious.
>
> - **(a) drill-down attribution — LIVE and PROVEN.** Shipped v1.0.1135, still
>   in v1.0.1140 (pod-verified). The signal went from the useless `fieldset
>   (419px)` to naming `div.ltb-row-grid` and why it will not shrink, and
>   tool-improver visibly re-aimed: its next attempt root-caused to the GRID
>   ("grid items couldn't shrink") where the two previous attempts had both said
>   "constrain the fieldset".
> - **(b) convergence guard — LIVE in v1.0.1144, proof IN FLIGHT.** Built
>   `b13238be6`, then two council-driven fixes: `b88da540b` (DO UPDATE refresh
>   instead of DO NOTHING) and `905fbbeef` (MERGE the spec on re-escalation, not
>   full-replace — so a human's triage keys survive). v1.0.1144 shipped the
>   guard with the *pre-merge* form; the `905fbbeef` merge fix needs the NEXT
>   image. The guard is LIVE and unfired (the benchmark's last verdict was 07-18,
>   inside the 7-day cooldown). **A manual acceptance run was queued 2026-07-21
>   11:51 to induce the first escalation** (the benchmark is genuinely past
>   threshold — 4 complete cycles at `mobile-fit`, verified live) — [UNPROVEN
>   until it resolves]; watch for an `acceptance_stuck:tool-loot-table-balancer:*`
>   row. Council: 3 rounds, all REVISE, no veto; rounds 1 & 2 each produced a
>   real fix (above), round 3 in flight. Details in §"Fix candidates" (b).
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

**(b) A convergence guard (judge/loop) — BUILT 2026-07-20, inert until the next
chassis image.** `judge_acceptance_results` now counts the fix cycles that
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

**(c) NOT recommended: hand-fix this one tool.** It would clean the trial site
but erase the benchmark and prove nothing. The tool is left overflowing on
purpose.

## Verify after fixing

1. With (a): re-run acceptance on `tool-loot-table-balancer`; the acceptance-fail
   note's routed culprit should name a grid child inside `#ltbRows`, not the
   fieldset; a targeted fix should then turn `mobile-fit@mobile` green.
   **DONE 2026-07-18** — the note names `div.ltb-row-grid` with a reason, and
   tool-improver re-aimed at the grid. The tool did NOT go green, for the
   unrelated reason in the correction at the top (024: the page never re-rendered).
2. With (b): force two non-converging cycles on any tool; the third fail should
   produce `needs_human_review`, not a third identical improve_tool item.
   **NOT YET DONE — this is the outstanding verification.** The guard is
   committed but inert until a chassis image rolls. The benchmark tool is
   already past the threshold (4 prior cycles, counted live), so the *first*
   Tier-4 verdict after the roll should escalate rather than raise a fifth
   improve_tool item — which makes this cheap to verify, but note it also means
   the escalation will fire on a tool whose real defect is 024's delivery chain.
   After the roll:
   ```
   -- the guard's own count, as the judge runs it (expect >= 2)
   SELECT count(*) FROM site_work_items w
   WHERE w.site_id='e33263f4-74f8-494f-b191-546845dbbddf'::uuid
     AND w.item_type='improve_tool'
     AND w.item_key='acceptance_fail:tool-loot-table-balancer:e33263f4-74f8-494f-b191-546845dbbddf'
     AND w.status IN ('complete','failed')
     AND jsonb_typeof(w.spec->'failing_checks')='array'
     AND EXISTS (SELECT 1 FROM jsonb_array_elements_text(w.spec->'failing_checks') e
                 WHERE e = ANY('{"mobile-fit"}'::text[]));
   -- after the next verdict: an escalation, and NO new improve_tool item
   SELECT item_type, status, spec->>'why_escalated' FROM site_work_items
   WHERE item_key LIKE 'acceptance_stuck:tool-loot-table-balancer:%';
   ```
   Verify the deploy with a **discriminating** pod-grep — a string the change
   CREATED, not one it merely uses:
   `kubectl exec -n ai-persona-system <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "is not converging on this defect"'`
   (positive control: `acceptance_stuck`; negative control: any string not in
   this change).

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` T24 (the run).
- The T15 chrome-attribution precedent (same class of signal-refinement) — same handoff, T15.
- Adapter overflow logic: `internal/adapters/browserrunner/run_checks_action.go`
  (`HorizontalOverflow`); judge routing: `platform/orchestration/actions/tool_acceptance_actions.go`.
