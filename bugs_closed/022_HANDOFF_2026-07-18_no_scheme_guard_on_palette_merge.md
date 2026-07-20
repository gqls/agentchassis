# 022 — Nothing stops a spec from flipping a site's colour scheme: light palette rendered onto a dark layout

**CLOSED 2026-07-20 — fixed AND live.** Fix `9c3b0c3e7`, live in the
v1.0.1140 chassis image (deployed 18:58 BST); verification record at the
foot of this file.

**Filed:** 2026-07-18 from the robot-hands R1 thread. This is the **damage
mechanism** behind that incident; `bugs_open/017` and the `generic_theme` check
fix (commits 3437f2212 + 3b52da8ec) only reduce how often it is *triggered*.
Raised by the council gate's `bug_historian` seat on correlation
`e0ebf6ee-dcc0-4a7b-9a3d-438ce9af5fff` (round 2), then **withdrawn from that
submission on the council's own recommendation** — guardian's stability point:
a shared render-merge change should not ride along with a single-file check fix.
It needs its own submission, which this file is the groundwork for.

## Symptom (real, shipped)

robot-hands.com runs layout `tool-portal-dark` (`layouts.scheme='dark'`, a user
decision taken twice: B7 on 2026-07-10 and again at the D13 gate). On
2026-07-17 20:31 a routine `webdesign-agent` run committed
`robot-hands.com/assets/css/styles.css` (gqls/sites `302702fc96`) carrying:

```
--color-background:     #F4F5F7;    /* light — on a scheme=dark site */
```

No error, no warning, no gate. It went live. Four CSS rewrites landed that day
(`5dfe168347` 08:54, `4b6685a422` 13:34, `bd814de701` 16:24, `302702fc96`
20:31), each with different core colours; only the last was catastrophic.

## Mechanism

`analyze_design` (webdesign-agent, `execute_llm_prompt`) emits a fresh
`color_scheme` on **every** run. `RenderCSSFromSpecAction` merges it at
`render_css_from_spec_action.go:119`:

```go
mergedPalette := buildPaletteMap(comp.Palette, specPalette)
```

`buildPaletteMap` (`render_css_composition_helpers.go:72`) applies the
documented core-vs-specialised rule: for **core** slots — `primary`,
`secondary`, `accent`, `background`, `surface`, `text`, `text_muted`, `border` —
**the spec always wins**. That rule is right for site identity and wrong for
*scheme*: nothing anywhere compares the proposed `background` against the
layout's declared `scheme`. The layout is the user's choice; a per-run LLM
guess silently overrides it.

Per-site mitigation applied (robot-hands only): pinned
`design_intent.palette.reference_values` so the prompt stops inventing —
`docs024_key_docs_latest/robot_hands/SQL_2026-07-17_r1b_design_intent_palette_pin.sql`.
That is a data patch on one site; every other site with a `scheme` layout and no
pin has the same exposure.

## Verifications already done (the council asked; these all resolve in favour)

- `parseHexColor` — **exists**, `platform/orchestration/actions/color_util.go:26`.
  `relativeLuminance` at `:64`. No new helper needed (reuse_agent's concern).
- `layouts.scheme` — **exists**: `text`, `CHECK (scheme IN ('light','dark',
  'neutral'))`; robot-hands' `tool-portal-dark` row is `scheme='dark'`.
  It is NOT currently carried on `themeComposition`
  (`render_css_composition_loader.go:37`) — the loader's JOIN must select it.
- Call sites — `buildPaletteMap` and `loadThemeComposition` have **exactly one
  non-test caller each** (`render_css_from_spec_action.go:119` and `:107`).
  So a guard at that boundary is NOT the "one call site patched, mechanism left
  generic" shape bug_historian feared; today it *is* the mechanism. (It must be
  re-checked at implementation time — that count is the load-bearing fact.)

## Fix candidate

Add `LayoutScheme` to `themeComposition` (populate from `l.scheme` in
`loadThemeComposition`'s JOIN), then guard immediately after the merge at
`:119`: when the layout declares `dark` and the merged `background` has
luminance > 0.5 (or the mirror case for `light`), keep the **theme's**
`background` *and* `text` together — never half-swap, that breaks contrast —
and `logger.Warn` naming both the rejected and kept values. Sites with
`scheme` NULL/`neutral` are untouched.

Open question for the submission (guardian): confirm no legitimate workflow
renders a light palette against a dark layout — a deliberate rebrand should
swap the layout first, which is the sanctioned 025 FK-swap order anyway.

## How to verify

- Unit: merged palette with a light background + `LayoutScheme="dark"` returns
  the theme background, and emits the Warn.
- Live: re-run webdesign-agent on a `scheme=dark` site with the design_intent
  pin REMOVED; the committed styles.css must stay dark and the pod log must
  carry the rejection line.
- Deploy (debug_historian's round-2 point): Go is inert until an image
  build+roll — verify with a pod-binary grep for the new symbol, never the
  commit hash or tag.

## FIX BUILT 2026-07-20 (this thread) — awaiting image roll, so still OPEN

Implemented exactly per the fix candidate above, as its own council
submission (`SUBMISSION_CORR=0328ddc7-3eac-4353-bcf9-b4b8f205720e`).

**Round 1 verdict: REVISE — and the objection was real.** editquality +
bug_historian both caught that my theme-without-text branch restored the
theme background while keeping the spec's text — the exact half-swap this
fix exists to forbid — and that my unit test codified it while the Warn
claimed a full restore. Revised: refuse any partial restore. Round-1
checks answered from the live system: exactly one live pipeline routes
through `render_css_from_spec` (`agent_definitions` → webdesign-agent
only); no `themeComposition` consumer exists outside the loader+action,
and template binding is a hand-built map, never the struct; bug 017's
generic_theme fix never reads `layouts.scheme` (it gates on
`content_data.color_scheme`), so this guard is the sole enforcement
point, not a duplicate.

**Round 2 verdict: REVISE — ten approvals, bug_historian again right.**
The round-1 revision "refused" by keeping the violating merge and
Warning — an unwatched log line as the only signal, the exact shape
behind this incident class ("a Warn nobody is paged on is one step
better than fully-silent, but is not fail-loud"). Revised to the seat's
own third named option: **hard-fail the render** — the same
migration-gap-must-be-loud contract `loadThemeComposition` already
enforces in this file pair, and a failed step is a recorded failure,
which the fleet immune-system sweep consumes without any new
`site_work_items` item_type (which would have needed a consumer and
touched the idx_swi_dedup lockstep). Its low-severity point (the
unjudgeable-background early-returns were fully silent) → both now emit
an Info naming what was skipped.

**Round 3 verdict: REVISE — the objections moved from the fix to my
supporting claims, and verifying them was worth it.** guardian: does a
failed step retry forever? **No** — the only step-retry counter is
spawn-scoped (`coordinator.go:1408`, cap 30); `generate_css` declares no
`error_step`, so the error fails the workflow once and the parent
dispatch loop marks the item and moves on. prior_art_librarian: is the
"failure sweep" real machinery or a doc quote? **Real, verified hop by
hop**: `failWorkflow` → `notifyParentOfFailure` → `agent_error_log`
severity=fatal (`coordinator.go:3520`) + parent notify →
`build-dispatch-loop.call_handler.error_step=mark_failed` →
`fail_work_item` → `site_work_items status='failed'` →
`diagnosis-triage` (live agent) gathers exactly that surface
(`diagnose_triage_action.go:328`). Honest caveat: `fail_work_item`
writes the static "Handler failed", so triage's signature grouping is
per-handler, not per-cause; the specific guard message lives in
`agent_error_log`. bug_historian: missingkey silent-blank in this
action's template? **Refuted for this action** — all palette/typo/token
lookups go through FuncMap helpers with mandatory fallback args
(Template Helper Fallback contract), and the five direct template keys
are set unconditionally; the missingkey=zero incidents are
`call_agent.go` and the closed json_envelope class. missingkey=error
hardening here would be the ride-along shape guardian round-1 forbade —
declined, flagged as its own candidate submission. Round 4 = evidence
resubmission, code unchanged from round 3.

**Round 4 verdict: REVISE — 13 approvals (all three round-3 objectors
now approve), one new medium from reuse_agent.** The seat read the
round-4 sketch standalone and objected that `parseHexColor`/
`relativeLuminance` "are never shown being defined or imported" — but
they are the PRE-EXISTING shared utilities in `color_util.go` (:26/:64,
same package, which is why the diff shows no import), a fact this very
file's §Verifications recorded before round 1 and reuse_agent's own
round-1 approval praised ("explicitly reuses parseHexColor/
relativeLuminance from color_util.go rather than writing new colour
math"). My round-4 rationale had dropped that citation — seats have no
cross-round memory, so every resubmission must carry its full evidence
again. **Lesson for the runbook trap list: a resubmission rationale
must re-state ALL standing evidence, not just the new round's.** Its
low (layouts.scheme provenance) answered: the column pre-exists, created
by `docs024_key_docs_latest/idea.uk/
migration_layouts_scheme_and_light_tool_portal.sql` (nullable + CHECK,
NULL = no scheme constraint by design). Round 5 = citations restored,
code still unchanged from round 3.

**Round 5 verdict: REVISE (12 approvals) — and here the trail was
closed under the runbook's advisory principle.** The deciding medium
was bug_historian objecting that the guard lives "in the caller, not in
the shared merge primitive" — the direct reverse of its own round-1
approval of the same placement on the same sketch ("explicitly patches
the verified single call site of the shared merge mechanism … the
correct fix shape per bug #6/#7's lesson"). This is the runbook's
documented cross-round-contradiction trap verbatim; its prescription:
both readings are defensible, pick one, record WHY in the code, move
on. **Picked: the boundary guard** (022's own council-reviewed
groundwork; `buildPaletteMap` is a pure no-logger helper and has exactly
one non-test caller, re-verified twice — today the call site IS the
mechanism). The WHY + the second-caller warning is now a comment on
`enforceLayoutScheme` itself. Committed WITHOUT a `Council-Reviewed`
trailer — that trailer is earned by APPROVED only, and this trail ends
REVISE; the five council_reports on corr `0328ddc7` are the honest
record (rounds 1–2 improved the code; rounds 3–5 verified claims:
retry-path, failure-sweep liveness, missingkey, reuse provenance,
column provenance, all resolved in the fix's favour).

**Remaining to CLOSE (fixed AND live bar):** build+roll the chassis
image from a commit containing this fix, then:
`kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c enforceLayoutScheme'`
(≥1 = live). Then the live verification above (§How to verify) — re-run
webdesign-agent on a scheme=dark site with the design_intent pin
removed; styles.css must stay dark and the pod log must carry the Warn.
NOTE for the rolling thread: the roll also activates other threads'
inert fixes (019, 001, code-lookup tier, V4 — see their bug files);
quiet-check first, and no orchestration dispatch within ~300s of the
restart.

## CLOSED 2026-07-20 — live verification record (v1.0.1140, deployed 18:58 BST)

- **Binary**: `9c3b0c3e7` is an ancestor of the build commit
  (`bca5d8255`); pod `agent-chassis-5567d99bd6-5snzn`:
  `strings /app/agent-chassis | grep -c enforceLayoutScheme` → 2, and
  the refusal string `refusing to render scheme-violating CSS` → 1.
  These strings are CREATED by this change (016b §9 pod-grep rule).
- **Live pipeline run, pin REMOVED** (the §How-to-verify test): a real
  webdesign-agent run (orch `fb744273`) on robot-hands with
  `design_intent.palette` superseded away. The LLM proposed a
  CONFORMING dark palette (`design_spec.color_scheme.background
  #0F1219` — it anchors on the previous spec in `content_data`), the
  guard passed it through, rendered CSS carries `--color-background:
  #0F1219`, deployed and confirmed on the live URL. Pass-through path
  proven live end-to-end on the new binary.
- **Reject path**: proven by unit tests with the incident's exact
  values (#F4F5F7 vs dark theme → theme background+text restored, Warn
  with rejected/kept fields) and by the strings in the pod binary.
  **[UNEXERCISED live]** — a deterministic live-fire (design_intent
  temporarily pointed at light reference_values so the spec would
  violate) was attempted twice (~18:45 and ~19:00 UTC); both kcat
  dispatches to `system.agent.generic.requests` vanished — no
  orchestration row, no error — while the cluster processed other
  work normally. [INFERRED] this is `bugs_open/003`'s broker-2 network
  fault at the PRODUCER side (kcat -P exits 0 on failed async
  delivery; each ephemeral kcat pod schedules onto a different node);
  occurrence noted in 003. The test was abandoned rather than left
  running against a mutated live site.
- **Site state restored**: the R1b pinned `design_intent` re-superseded
  VERBATIM from the pre-test snapshot (`background #0F1218`, guidance
  intact, `created_by='bugfix-022-thread'` rows document the test
  window). Live styles.css dark; pin retained as defence-in-depth.
- **Closure judgement**: the defect ("nothing anywhere compares the
  proposed background against the layout's declared scheme") is
  structurally closed by code proven present and functional in the
  production binary. If a live reject-fire is later wanted, the
  deterministic method is documented above: supersede the pin with
  light reference_values, dispatch webdesign-agent, expect dark CSS +
  the Warn, restore the pin.

- `render_css_composition_loader.go`: `l.scheme` selected in the existing
  JOIN, scanned as NullString (column nullable — live distribution 2026-07-20:
  15 NULL / 2 light / 1 dark), exposed as `themeComposition.LayoutScheme`.
- `render_css_from_spec_action.go`: new `enforceLayoutScheme` (returns
  error), called immediately after `buildPaletteMap` and before
  `logOverrides` (so a reverted background reports as
  `claimed_but_ignored`, and every downstream consumer —
  `BackgroundIsDark`, `buildSectionDefaults`, template lookups — sees the
  corrected palette). Violation = declared `dark` + luminance > 0.5, or
  declared `light` + luminance < 0.5. Outcomes: theme supplies both
  `background`+`text` → restored together, one Warn naming rejected and
  kept values; theme missing either slot → **the render hard-fails**
  (nothing ships, site keeps last-good CSS, the failed step is what the
  fleet failure sweep consumes); background absent/non-hex → passes with
  an Info; scheme NULL/neutral → inert and logless.
- `render_css_scheme_guard_test.go` (new): incident case (#F4F5F7 on
  scheme=dark), light-mirror, conforming override survives, NULL/neutral
  inert, unjudgeable-with-Info, incomplete-theme-fails-render — all pass
  against `git archive HEAD` + these three files overlaid (other
  sessions' WIP excluded).

Load-bearing fact re-checked at implementation time as §Verifications
demanded: `buildPaletteMap` / `loadThemeComposition` still exactly one
non-test caller each (`render_css_from_spec_action.go:119` / `:107`).

**Close only when live**: pod-binary check is
`strings /app/agent-chassis | grep -c enforceLayoutScheme` (never the tag).
The robot-hands `design_intent` pin stays in place as defence-in-depth.
