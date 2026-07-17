# HANDOFF — the fix loop does not converge on layout-intrinsic mobile overflow

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

**(b) A convergence guard (judge/loop) — the safety net.** Track fix attempts
per (function, criterion): after N (say 2) improve_tool cycles that did not
turn the criterion green, stop auto-raising improve_tool and raise a
`needs_human_review` item instead. The fixer already loads prior NOTES; the
loop should *act* on the repetition the notes record. Without this, any defect
the one-shot fixer can't solve becomes a weekly no-op forever.

**(c) NOT recommended: hand-fix this one tool.** It would clean the trial site
but erase the benchmark and prove nothing. The tool is left overflowing on
purpose.

## Verify after fixing

1. With (a): re-run acceptance on `tool-loot-table-balancer`; the acceptance-fail
   note's routed culprit should name a grid child inside `#ltbRows`, not the
   fieldset; a targeted fix should then turn `mobile-fit@mobile` green.
2. With (b): force two non-converging cycles on any tool; the third fail should
   produce `needs_human_review`, not a third identical improve_tool item.

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` T24 (the run).
- The T15 chrome-attribution precedent (same class of signal-refinement) — same handoff, T15.
- Adapter overflow logic: `internal/adapters/browserrunner/run_checks_action.go`
  (`HorizontalOverflow`); judge routing: `platform/orchestration/actions/tool_acceptance_actions.go`.
