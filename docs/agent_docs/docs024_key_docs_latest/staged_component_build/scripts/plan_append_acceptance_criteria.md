
## Acceptance criteria

Authored 2026-07-31 by lane `staged_component_build`. Until now this PLAN had a
Verification contract naming a local Python probe, and **no criteria fence at all** —
so a fired acceptance run resolved the page, found nothing to assert, and skipped
with `needs_criteria`. A skip is honest but it is not a failure, so the run read as
clean. That is the silent class `CHECK_naming_contract.sh` exists to find, and this
tool was its one BROKEN-B instance fleet-wide.

Every check below was watched to pass against the live URL, and watched to FAIL under a
mutant that breaks the behaviour it names. Neither claim is an assertion about this
document: both are re-runnable, and the commands are in the lane RUNBOOK.

**Read the profile gating before reading the checks.** All 18 were verified on desktop
*and* mobile offline (36/36, three consecutive runs), but only 4 run on mobile in the
cluster, because 36 evaluations exceeded the 120-second run deadline and the first
dispatch failed on it. The four were chosen because their mobile answer is not implied by
their desktop answer; the rest assert profile-independent facts. The reasoning, the
measurement and the misleading error message it produces are in
*"The 120-second run deadline"* below — that section is the most useful thing here for
anyone authoring their next fence.

```criteria
{
  "profiles": [
    "desktop",
    "mobile"
  ],
  "container": ".tool-container[data-component=\"tool-review-council-simulator\"]",
  "checks": [
    {
      "id": "page-serves-200",
      "type": "page_status_ok"
    },
    {
      "id": "tool-container-renders",
      "type": "selector_exists",
      "profiles": [
        "desktop"
      ],
      "selector": ".tool-container[data-component=\"tool-review-council-simulator\"]"
    },
    {
      "id": "roster-is-built-client-side",
      "type": "selector_exists",
      "selector": ".rcs-seat"
    },
    {
      "id": "default-preset-is-our-typical-eight",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-seats-out",
        "text_matches": "^8 seats on the panel$"
      }
    },
    {
      "id": "threshold-starts-where-we-run-it",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-threshold-out",
        "text_matches": "^Only high severity blocks$"
      }
    },
    {
      "id": "headline-is-computed-not-a-placeholder",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-pass1",
        "text_matches": "^[0-9]+\\.[0-9]%$"
      }
    },
    {
      "id": "within-rounds-is-computed",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-passn",
        "text_matches": "^[0-9]+\\.[0-9]%$"
      }
    },
    {
      "id": "mean-rounds-is-computed",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-rounds-needed",
        "text_matches": "^[0-9]+\\.[0-9]$"
      }
    },
    {
      "id": "seats-firing-is-computed",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-firing",
        "text_matches": "^[0-9]+\\.[0-9]$"
      }
    },
    {
      "id": "blocker-chart-is-ranked",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-blockers-list .rcs-bar-row"
      }
    },
    {
      "id": "reality-band-is-drawn",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-reality-legend .rcs-legend-pct"
      }
    },
    {
      "id": "denominator-states-what-it-counts",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "expect": {
        "selector": "#rcs-denominator",
        "text_matches": "count ROUNDS"
      }
    },
    {
      "id": "threshold-lever-updates-the-readout",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "steps": [
        {
          "action": "fill",
          "selector": "#rcs-threshold",
          "value": "0"
        }
      ],
      "expect": {
        "selector": "#rcs-threshold-out",
        "text_matches": "^Any objection blocks$"
      }
    },
    {
      "id": "every-seat-preset-counts-twenty-six",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "steps": [
        {
          "action": "click",
          "selector": ".rcs-preset[data-preset=\"all\"]"
        }
      ],
      "expect": {
        "selector": "#rcs-seats-out",
        "text_matches": "^26 seats on the panel$"
      }
    },
    {
      "id": "no-horizontal-overflow",
      "type": "no_horizontal_overflow"
    },
    {
      "id": "minimal-preset-is-the-measured-pair",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "steps": [
        {
          "action": "click",
          "selector": ".rcs-preset[data-preset=\"minimal\"]"
        }
      ],
      "expect": {
        "selector": "#rcs-seats-out",
        "text_matches": "^2 seats on the panel$"
      }
    },
    {
      "id": "cleared-panel-refuses-to-invent-a-number",
      "type": "interaction",
      "profiles": [
        "desktop"
      ],
      "steps": [
        {
          "action": "click",
          "selector": ".rcs-preset[data-preset=\"none\"]"
        }
      ],
      "expect": {
        "selector": "#rcs-pass1",
        "text_matches": "^n/a$"
      }
    },
    {
      "id": "no-console-errors",
      "type": "no_console_errors"
    }
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

### The 120-second run deadline is a first-class constraint, and it cost this fence a redesign

**Measured, not predicted.** The first cluster dispatch of this fence
(correlation `211dd1d4-6bfc-4418-83f1-4191f6d1e0c1`) **FAILED** after 133 seconds with:

```
run_checks: browser open failed for https://fundamentallyai.com/tools/review-council-simulator.html
[mobile]: context deadline exceeded (code: run_checks_failed)
```

Read that message carefully, because it is misleading in a specific way: it names the
**browser open** and it sounds like infrastructure. It is not. `runDeadline` is
**120 seconds for the WHOLE request** — every url x every profile — and `openChromium`
returns `ctx.Err()` if the deadline expires during its settle wait. So a fence that is
simply too large presents as a browser that would not start.

The fence at that point declared 18 checks on 2 profiles = **36 evaluations**, and it was
**correct**: 36/36 against the live page, three consecutive local runs, ~10.6 seconds each.
**Local timing tells you nothing about the cluster's budget** — in-cluster cost is roughly
3-5 seconds per evaluation against ~0.3 locally. The only comparable data point is the one
other acceptance run that has ever happened (`dc952633`, 2026-07-30): ~21 evaluations, 48
seconds, passed.

**The fix is a design improvement rather than a workaround: assert on mobile only what
mobile can answer differently.** The tool's arithmetic, presets and readout text are
profile-independent — running them twice asserted the same fact twice. Four checks are
gated to both profiles because their mobile answer is genuinely not implied by their
desktop answer:

| check | why it runs on mobile too |
|---|---|
| `page-serves-200` | the mobile UA is a different request |
| `roster-is-built-client-side` | the teaser-reveal-panel class is per-profile — JS can run on one and not the other |
| `no-horizontal-overflow` | **the whole reason a mobile profile exists**; a 26-row roster and a bar chart at 390px |
| `no-console-errors` | a mobile UA can take a different JS path |

The other 14 carry `"profiles": ["desktop"]`. That is **18 + 4 = 22 evaluations**, down from
36, and it loses no assertion — only duplicate ones.

**The 14 profile-gated skips are intentional and are NOT the silent class.** A gated skip
is the fence author's own instruction; `try_fence.go` reports it in a separate bucket from
a `not implemented` skip, and the arithmetic still reconciles to 18 x 2 = 36 slots
(22 evaluated + 14 gated). If you see a skip in a cluster run, check which kind it is
before reading it either way.

**And this is a limitation of the offline harnesses, stated plainly:** `try_fence.go` proves
a fence is CORRECT; it cannot prove it FITS. It runs on a developer machine an order of
magnitude faster than the pod, and it does not model the deadline. **A fence is not proven
until it has completed once in the cluster.** Mobile coverage of all 18 checks remains
established offline (36/36, three runs) — it is simply not part of the cluster gate.

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
