
## Acceptance criteria

Authored 2026-07-31 by lane `staged_component_build`. Until now this PLAN had a
Verification contract naming a local Python probe, and **no criteria fence at all** —
so a fired acceptance run resolved the page, found nothing to assert, and skipped
with `needs_criteria`. A skip is honest but it is not a failure, so the run read as
clean. That is the silent class `CHECK_naming_contract.sh` exists to find, and this
tool was its one BROKEN-B instance fleet-wide.

Every check below was watched to pass against the live URL on both profiles, and
watched to FAIL under a mutant that breaks the behaviour it names. Neither claim is
an assertion about this document: both are re-runnable, and the commands are in the
lane RUNBOOK.

```criteria
{
  "profiles": ["desktop", "mobile"],
  "container": ".tool-container[data-component=\"tool-review-council-simulator\"]",
  "checks": [
    { "id": "page-serves-200", "type": "page_status_ok" },

    { "id": "tool-container-renders", "type": "selector_exists",
      "selector": ".tool-container[data-component=\"tool-review-council-simulator\"]" },

    { "id": "roster-is-built-client-side", "type": "selector_exists",
      "selector": ".rcs-seat" },

    { "id": "default-preset-is-our-typical-eight", "type": "interaction",
      "expect": { "selector": "#rcs-seats-out", "text_matches": "^8 seats on the panel$" } },

    { "id": "threshold-starts-where-we-run-it", "type": "interaction",
      "expect": { "selector": "#rcs-threshold-out", "text_matches": "^Only high severity blocks$" } },

    { "id": "headline-is-computed-not-a-placeholder", "type": "interaction",
      "expect": { "selector": "#rcs-pass1", "text_matches": "^[0-9]+\\.[0-9]%$" } },

    { "id": "within-rounds-is-computed", "type": "interaction",
      "expect": { "selector": "#rcs-passn", "text_matches": "^[0-9]+\\.[0-9]%$" } },

    { "id": "mean-rounds-is-computed", "type": "interaction",
      "expect": { "selector": "#rcs-rounds-needed", "text_matches": "^[0-9]+\\.[0-9]$" } },

    { "id": "seats-firing-is-computed", "type": "interaction",
      "expect": { "selector": "#rcs-firing", "text_matches": "^[0-9]+\\.[0-9]$" } },

    { "id": "blocker-chart-is-ranked", "type": "interaction",
      "expect": { "selector": "#rcs-blockers-list .rcs-bar-row" } },

    { "id": "reality-band-is-drawn", "type": "interaction",
      "expect": { "selector": "#rcs-reality-legend .rcs-legend-pct" } },

    { "id": "denominator-states-what-it-counts", "type": "interaction",
      "expect": { "selector": "#rcs-denominator", "text_matches": "count ROUNDS" } },

    { "id": "threshold-lever-updates-the-readout", "type": "interaction",
      "steps": [ { "action": "fill", "selector": "#rcs-threshold", "value": "0" } ],
      "expect": { "selector": "#rcs-threshold-out", "text_matches": "^Any objection blocks$" } },

    { "id": "every-seat-preset-counts-twenty-six", "type": "interaction",
      "steps": [ { "action": "click", "selector": ".rcs-preset[data-preset=\"all\"]" } ],
      "expect": { "selector": "#rcs-seats-out", "text_matches": "^26 seats on the panel$" } },

    { "id": "no-horizontal-overflow", "type": "no_horizontal_overflow" },

    { "id": "minimal-preset-is-the-measured-pair", "type": "interaction",
      "steps": [ { "action": "click", "selector": ".rcs-preset[data-preset=\"minimal\"]" } ],
      "expect": { "selector": "#rcs-seats-out", "text_matches": "^2 seats on the panel$" } },

    { "id": "cleared-panel-refuses-to-invent-a-number", "type": "interaction",
      "steps": [ { "action": "click", "selector": ".rcs-preset[data-preset=\"none\"]" } ],
      "expect": { "selector": "#rcs-pass1", "text_matches": "^n/a$" } },

    { "id": "no-console-errors", "type": "no_console_errors" }
  ]
}
```

### What the fence asserts, and why these and not others

Each check asserts a claim this PLAN already makes, rather than a property of the
markup. The three that carry the most weight:

- **`cleared-panel-refuses-to-invent-a-number`** is the fence's best check. The
  Behaviour contract promises "empty roster renders n/a and an explanation, never a
  misleading 100%". Its mutant replaces that `n/a` with `100.0%` and the check goes
  red — so the tool's *honesty* is now a machine-checked property, not a line in a
  document.
- **`roster-is-built-client-side`** is the teaser-reveal-panel gate. `#rcs-seats` is
  served EMPTY (`<div class="rcs-seats" id="rcs-seats"></div>`) and all four result
  figures are served as `--`; every one of them is filled by the IIFE. So these
  checks pass only if the script genuinely ran in the visitor's browser, which is
  precisely what four rounds of static-markup verification failed to establish for
  `teaser-reveal-panel`.
- **`denominator-states-what-it-counts`** asserts the provenance text is actually
  rendered. The tool's numbers are a dated snapshot (Deliberate decision 1), and a
  snapshot whose denominator quietly stops being displayed is the failure that turns
  an honest figure into a misleading one.

### What it deliberately does NOT assert — all four are omissions, not oversights

1. **No `has_visible_area` check, even though the type is live in the running
   binary** (verified 2026-07-31 by long-marker grep of
   `/app/browser-runner-adapter`, build 08:53:36 UTC, with three positive
   controls). `bugs_open/157` is open and unfixed at HEAD
   (`run_checks_action.go:773-774`): the decode asserts `float64`, but
   playwright-go returns `int` for a whole number, so **any axis whose rendered
   size is an integer measures 0** and the check reports "too small to see or
   click". Adding it here would have bought a false FAIL on a correct tool. The
   bug belongs to the leopardess lane, which holds the reproducer. **Add these
   checks when 157 closes** — the 26-row roster's checkboxes are exactly the
   integer-sized controls the type exists to police.
2. **No count assertion via `selector_count`.** That type is behaviourally
   IDENTICAL to `selector_exists` — `run_checks_action.go:497` passes on `n > 0`
   and there is no expected-count field on the check. It reports the count in its
   detail line, which reads like an assertion and is not one. So every count in
   this fence (8, 26, 2 seats) is asserted through the tool's own readout text
   instead, which has the side benefit of asserting the tool can *say* what it did.
3. **`input`-versus-`change` wiring is not distinguishable by any fence.** The
   Behaviour contract says the inputs update on the `input` event. Playwright's
   `fill()` dispatches both events and the criteria vocabulary has no way to send
   one without the other, so a tool wired only to `change` would pass. This was
   found by mutation, not by reasoning: the first version of the mutant killed only
   the `input` listener and the check still passed. The check is therefore named
   `threshold-lever-updates-the-readout`, which is what it proves. The "no
   calculate button" half of the claim IS covered — no click is needed for the
   readout to change.
4. **No aesthetic check.** PLAN D2: every gate is validation, never judgement.

### Order is load-bearing

`evaluateOnPage` runs the whole fence against ONE page instance per profile, in
fence order, so interactions accumulate. The structural and default-state checks
therefore come first, then the threshold fill, then the presets in the order
all -> minimal -> none, because `none` wipes the results panel. `no_console_errors`
is forced last by the runner itself so it catches interaction-triggered errors.
**Reordering this fence changes what it tests.**

### How to re-verify (both halves — one is not evidence without the other)

```
go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/try_fence.go \
  docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/fence_tool_review_council_simulator.json \
  https://fundamentallyai.com/tools/review-council-simulator.html

go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail.go \
  docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/fence_tool_review_council_simulator.json \
  https://fundamentallyai.com/tools/review-council-simulator.html
```

Measured 2026-07-31: **36 of 36 check-evaluations passed** on the live URL across
desktop and mobile with zero skips and the arithmetic reconciled, and **17 of 17
mutants were caught, with all 18 checks watched red** against an all-green
baseline. Both harnesses call the platform's own `RunChecksAction`, not a
reimplementation of it.
