# 256 — mobile full-page screenshot at 3x DeviceScaleFactor exceeds the vision API's 8000px dimension cap

**Filed 2026-08-12** by `staged_component_build`, surfaced while firing a routine acceptance
run at a newly-authored PLAN (`tool-llm-cost-calculator`, work item `a8cc2fef-8dbf-44dc-
a636-c7ffd55acdd4`, correlation `5a5f08b2-89d3-4433-8d9d-e4b729689e4a`). **Not this lane's
tool at fault** — the CHECKS half of that run passed cleanly (16/16, 0 failed, matching the
offline proof), so this is filed as a separate, cross-cutting finding rather than folded
into that PLAN's own record.

**Status: OPEN. Symptom and one root-cause link MEASURED first-hand; breadth across other
pages NOT measured. No fix proposed yet.**

## Declared substitute for the 090 loop (owner ruling 2026-07-31)

Not run. Offered in its place: the failing API response was read verbatim (below), the
actual rendered PNG was fetched from its signed S3 URL and its pixel dimensions read
directly from the PNG header (not inferred from viewport config), and the code path that
produces it was read and cited at the exact line. What is **not** independently confirmed
is how many *other* pages/tools this affects — this file measures the mechanism on one
page, not its fleet-wide reach. Flagged as `[UNMEASURED]` in the section below; a thread
that wants the blast-radius number should run 090 or a direct census.

## Symptom (measured, could have come out otherwise)

`execute_vision_prompt` failed with:

```
step look failed: failed to execute action execute_vision_prompt: execute_vision_prompt:
vision call failed: API request failed with status 400:
{"type":"error","error":{"type":"invalid_request_error","message":
"messages.0.content.1.image.source.base64.data: At least one of the image dimensions
exceed max allowed size: 8000 pixels"},"request_id":"req_011Cdxwy1hQq4LP8W7NoST6N"}
```

`current_step` ends `complete_no_look` — the checks half completed and the vision half did
not, same shape as `bugs_open/243` but **a different root cause**: 243 is `no storage
client` (the download never starts); this run got past that (243's fix is live) and failed
at the actual API call, so the two are sibling failure modes of the same step, not the same
bug re-occurring.

## Root cause, so far as measured

The run's two `landing`-stage renders (`orchestration_states.collected_data->'request_run'
->'response'->'renders'`) were fetched from their signed S3 URLs and their PNG headers read
directly:

| profile | viewport (config) | actual PNG pixel size |
|---|---|---|
| desktop | 1366x900@1x | **1366 x 2108** — comfortably under 8000 |
| mobile | 390x844@3x | **1170 x 10059** — height exceeds the 8000px cap |

The mobile PNG's width (1170) is exactly `390 x 3`; its height (10059) is the full page's
rendered CSS height at 390px width, also multiplied by 3. The 3x comes from
`run_checks_action.go:267` (`mobileScale = 3 // DeviceScaleFactor for the mobile context,
and the @Nx in Viewport`), applied at `run_checks_action.go:903`
(`ctxOpts.DeviceScaleFactor = playwright.Float(mobileScale)`) — Playwright's full-page
screenshot (`page.Screenshot(true)`, called from `run_checks_action.go:365`, in the
LANDING state per the TL-035(d) comment at line 354 — i.e. before any check interacts with
the page) captures at that device scale, so a page whose mobile-width CSS height is
anywhere above ~2667px (8000 / 3) will produce a PNG that trips this cap regardless of what
the page's own content is.

**`[UNMEASURED]`**: whether 2667px of mobile CSS height is a low, easily-crossed bar for
this estate's pages in general, or whether `tool-llm-cost-calculator`'s landing page
(header + hero + the tool's own input form + footer, at 390px width) is unusually tall. No
other page's render was checked. `bugs_open/243`'s own count (26 `complete_no_look` runs in
its retained window, all on the storage-client cause) predates this cause becoming
reachable, so it cannot be used to estimate this one's frequency.

## What is NOT lost

Exactly as 243 states: the check results (`request_run`) land and are judged independently
of `look`. This run's 16/16 pass stands regardless of the vision failure. What is lost is
the vision half's own judgement for this run.

## Fix candidates (not attempted, not evaluated against each other)

1. **Downscale before sending to the vision API**, not before storing — keep the stored S3
   evidence at full fidelity for humans, resize the copy `execute_vision_prompt` sends so
   its longest edge is under 8000px. Contained to one action.
2. **Cap full-page screenshot height at capture time** (`run_checks_action.go:365`) rather
   than at the point of failure — e.g. clip to N viewport-heights, or fall back to a
   viewport-only (non-full-page) shot when the full-page height would exceed the cap.
   Changes what gets stored as evidence, not just what reaches the API.
3. **Lower `mobileScale`** — cheapest, but changes every mobile screenshot's fidelity
   platform-wide for a defect that (per the `[UNMEASURED]` note above) may only bite a
   minority of unusually tall pages.

## How to re-verify

```sql
SELECT correlation_id, current_step, collected_data->'__step_error'->>'message'
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'work_item_id' = 'a8cc2fef-8dbf-44dc-a636-c7ffd55acdd4';
```

Renders (signed URLs expire after 7 days from 2026-08-12):
```sql
SELECT jsonb_pretty(collected_data->'request_run'->'response'->'renders')
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'work_item_id' = 'a8cc2fef-8dbf-44dc-a636-c7ffd55acdd4';
```
Read a fetched PNG's own header rather than trusting the `viewport` field in the row above —
that field states the requested viewport, not the full-page screenshot's actual output size.
