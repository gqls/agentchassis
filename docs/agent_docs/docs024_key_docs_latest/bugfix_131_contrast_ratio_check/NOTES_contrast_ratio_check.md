# NOTES — contrast_ratio check lane (append-only, newest at the bottom)

## 2026-08-22 — session 1: validity re-check, research, design

**Task**: owner asked for a look at `bugs_open/131` (resolved BY SLUG — the vonc gauntlet audit;
`bugs_closed/131` is the unrelated og-image case), a robust framework-preferring fix plan, council,
and a check that the bug is still valid and unowned.

**Ownership check** (`scripts/who-owns.py vonc_gauntlet`, 2026-08-22): owning lane is
`gauntlet_dead_cta`; its last 131 activity is 2026-07-31, NOTES/README end 2026-08-03, and the
2026-08-12 design-pass handoff does not mention 131/D/H at all. The recent commits on that lane are
the robot-hands gripper dossier routing through tools-api — unrelated to 131 (proof: zero "131" hits
in that dir). Diagnosis queue (`needs_diagnosis`/`awaiting_diagnosis`): empty. Open work items
matching contrast/vonc/131: none that touch this lane's target (the contrast-adjacent rows are
`improve_tool`/`audit_tool` items on other sites' tools).

**Live re-measurement** (playwright venv `~/.venvs/vonc_pw`, script `check_131_validity.py`, output
archived below in this dir's `p_sources` note; screenshots `shot_gauntlet_sealed_390.png`,
`shot_home_stats_390.png` in the session scratchpad):
- **A REGRESSED**: `.gi-title-accent` `rgb(245,158,11)` on section bg `rgb(124,60,255)` = **2.48:1**
  (< 3.0 large-text bar; fixed 07-28 at 3.31:1 against `#6d28d9`). The background token has churned
  back to `#7c3cff` — the very palette the bug's correction paragraph describes as the pre-07-28
  value. `[INFERRED]` the churn came from a palette regeneration/re-render, not a hand edit — the
  WHEN and WHAT of the churn are not yet pinned to a row; the MEASUREMENT of today's state is firm.
- **D REGRESSED past its filed state**: `.gi-challenge-text` = 256px of 390 = **65.6%** (was 74%
  filed, 83% after the 07-28 fix). Measured in the revealed state (one real round consumed).
- New firm sub-2:1 failures live today: `div.gi-eyebrow` **1.66:1**, `div.gi-rules-label` **1.76:1**
  (gauntlet page); homepage `a.gauntlet-btn-primary` **1.76:1** (= bug 122 Finding 3, recorded there
  as "STILL LIVE, the Gauntlet workstream's surface"), `span.stats-eyebrow` **1.63:1** (= bug 212
  §8's component-painted-ground case, recorded there as vonc 1.63:1). Screenshot-confirmed by eye.
- Disconfirmability: the same scan returns clean elements on most of the page and returned **zero**
  sub-2:1 rows for the provocation card and step controls — the instrument can and does say "fine".

**Why no 090 run for these claims** (the 2026-07-31 owner ruling's named escape hatch): the
regression claims are first-hand live measurements with by-eye confirmation, reproducible from the
archived script; the "no check can see contrast" structural claim was verified by reading the
dispatch switch (`run_checks_action.go:554-557` — eight types, none colour) and is independently
recorded by `LANDMINES.md:1523` and register `DES-054`. The one claim NOT verified — what exactly
churned the palette and when — is marked `[INFERRED]` above and deliberately NOT asserted as a root
cause anywhere.

**Research** (four parallel readers; full reports in session transcript):
- Contrast machinery map: maths single-sourced in `platform/colour/contrast.go`; generation-side
  prevention live (`--color-*-ink`, v1.0.1298); DB-level detector `check_palette_contrast` (files
  capability_gap to nobody, deliberate); live-page AUDIT `render_audit_action.go` + hourly rotation
  (VIZ-015) filing `contrast_failure` → css-patch-agent; **no check type can fail a page on
  contrast**. The check-type surface to touch: `run_checks_action.go` (allowlist :554, eval switch,
  struct :211, header), `experience_criteria.go` (:57 tiers, :108 type-fields), lockstep tests
  (whole-file case-literal harvest — every new `case "x"` demands a tiers entry), NOT Tier-2.
- History: the `contrast_ratio` check was proposed in
  `experience_loop/HANDOFF_2026-07-28_appeal_dimension.md` (spec at :141: computed colour vs nearest
  non-transparent ancestor, 4.5/3.0) and never built; the experience_loop dir has zero follow-up.
  `cmd/contrastscan` is a PHANTOM (built+deleted 2026-07-28, LANDMINES:174). 122 closed 08-15
  (owner ruling) on the generation-side fix; 113 effectively closed on its mechanism (no commits
  since 08-12); 296/242/352 are the render-audit lane's open defects — none of them owns a GATE.
- Gauntlet lane state of 131: **D was never decided** — `HANDOFF_2026-07-28_continue_here.md:66`
  says FIXED at 83%, the bug header says "open (design decision)", §D has no update banner; no
  parking note exists. **H**: owner ruled 2026-07-29 "3 leading to 2" (distribution experiment
  first); engineering delivered in full 07-29→07-31 (opinion ledger, share card, published round
  record `round.html?r=<slug>`); the outstanding leg is the OWNER'S OWN distribution experiment,
  which has no recorded occurrence. Nobody has recorded who closes 131.

**Consumer enumeration for RFC_022** (queries in RUNBOOK): typed fences 0 · active agent configs 0 ·
seeds 0. Four `doc_plans` rows match the bare substring — they are the smart-contrast /
oklch-picker tools' own prose. First query attempt used `content` column (does not exist — it is
`body`); recorded here so the RUNBOOK query stays the corrected one.

**Design settled**: see PLAN. Notable: reuse `browserPage.Evaluate`; factor `auditJS`'s maths into a
shared JS const (refusing a third WCAG-maths copy); overflow-style attribution (fail on any firm
failure, `Scope` routes tool vs chrome); `over_image` can never fail; default AA 4.5/3.0, explicit
`min_ratio` overrides both; Phase 1 does NOT advertise the type to planners (LANDMINES:512 window).

## 2026-08-22 — session 1 (continued): council submitted, check built, all green

- Council: DRY_RUN admission passed, then submitted for real. **SUBMISSION_CORR =
  `7e2391ec-47d0-4820-afde-b4cc475714e5`** (JSON alongside this file). Committing before the verdict
  with `Council-Submitted:` per the 2026-07-30 rule; 098 resolves at report time.
- Built: `contrast_check.go` (shared `contrastMathsJS` kernel + probe + Go-side verdict
  `runContrastRatio`), the four `run_checks_action.go` edits (header, `MinRatio` field, allowlist,
  eval arm), `auditJS` recomposed onto the shared kernel (append/split only — `TestAuditJSComposition`
  pins the fragments and that `effBG` appears exactly once), `experience_criteria.go` tier +
  type-field entries, 11 tests in `contrast_check_test.go` (verdict table, wiring through
  `splitByProfile`/`evaluateOnPage`, probe embedding, composition guard). fakePage grew an
  `evalScripts` recorder; its stale "run_checks never calls Evaluate" comment corrected.
- Proof: `go build ./...` exit 0; full suites green (`browserrunner` 19.5s, `actions` 3.7s —
  the two lockstep tests harvested the new case literal and passed against the new table entries).
- Register: **TL-049** appended to `tool-lifecycle.md` (numstat 9/0 — append-only verified per the
  deleted-bullet landmine). No `102_coverage_ratchet.txt` line matched.
- NOT done, deliberately (Phase 2, after the browser-runner-adapter image rolls): pod-grep proof,
  the vonc witness (`div.gi-eyebrow` 1.66:1) + clean control, seed/planner vocabulary updates
  (incl. 259's deferral-honesty list). The check is INERT until that roll — LANDMINES:512 is why no
  fence names it yet. The adapter is its OWN image (a chassis roll does nothing for it); single-
  service target precedent `35c8277a8`.

## 2026-08-22 — council round 1: REVISE, and the gating objection found a REAL defect I had shipped

Verdict `revise`, `decided_by: gating objection from bug_historian`, 4 abstained, 9 approve / 4 object.
Report: `diagnosis_artifacts` where `correlation_id LIKE '7e2391ec%' AND kind='council_report'`
(⚠ the table is `diagnosis_artifacts`, NOT `fix_artifacts` — I guessed the latter twice and got
"relation does not exist"; the RUNBOOK query is corrected).

**The gating objection was RIGHT, and the defect was already committed** (`b32aa9cd9`):
`runContrastRatio` treated an Evaluate failure/nil/garbage result the same as "zero sub-threshold
records found". A nil result decoded to a zero-value `contrastScan` and returned **pass**. That is
`render_audit.py`'s own landmine — *"prints 0 contrast failure(s) and exits 0 for a page it never
measured… as a gate it PASSES WHILE BLIND"* — reproduced one rung higher, inside the check written
to replace that very failure mode. My own tests did not catch it because every fixture supplied a
well-formed scan: **I tested the paths I built, not the path where the probe never runs.**

Fix (`2611b0b16`): the probe stamps `probe: "contrast_ratio/v1"` and the verdict refuses to grade
any payload lacking it; the probe counts what it measured and `scanned == 0` fails closed as
vacuous; eval/decode errors stay explicit FAILs. Four tests pin exactly those paths.

**Second real objection** (bug_historian + guardian, medium): my `TestAuditJSComposition` asserted
substring presence, which proves inclusion and not behaviour — a lost semicolon or duplicate
binding at the join would pass it. Replaced with **byte-identity** against a golden extracted
mechanically from `git show b32aa9cd9~1` (never transcribed) into
`internal/adapters/browserrunner/testdata/audit_js_golden_2026-08-22.txt`. Equality with a string
that demonstrably executed in production is stronger than any syntax check, and it settles the
guardian's two-consumer worry: `render-audit-adapter` runs this package ~80 tags behind
`browser-runner-adapter`, and an inert recomposition needs no coordinated rollout.

**Third** (debug_historian, medium): my witness plan never named WHICH pod. It must be
`-l app=browser-runner-adapter` (the gating service) — proving it at `render-audit-adapter` proves
nothing about Tier-4 acceptance. The RUNBOOK recipe already targeted the right pod; the plan did
not say so, and the marker sentence it greps is now pinned by a unit test so a rewording cannot
silently break the recipe.

Lows answered in the resubmission rather than by code (each verified, not asserted): the lockstep
harvest regex anchors `^\s*case` so a `//` line cannot match; `MinRatio` follows
`MinWidth`/`ExpectValues`' own per-type convention with P7's type-fields table as the collision
control; `toolContainer` at `run_checks_action.go:252` IS the shared helper (two call sites, :659
overflow + contrast_check.go:173). ⚠ I first wrote `:238` and `:645` from memory of the pre-edit
file — my own edits had shifted them. Grep before quoting a line number in a submission.

Round 2 dispatched on the same trail (`RESUBMIT_CORR=7e2391ec…`), run orch `c047a44b`.

### Mutation proofs of the round-2 guards (2026-08-22, run in a scratch copy — the real tree never mutated)

A guard I have not seen fire is the thing this lane's own landmine is about, so each new guard was
defeated in a throwaway copy of the tree and the intended test had to catch it:

| mutation | killed by | evidence |
|---|---|---|
| drop the semicolon at the `auditJS` join point (the EXACT class guardian + bug_historian named) | `TestAuditJSComposition` | `composed auditJS diverges from the pre-refactor literal at byte 1274: "rast:[],images:[],overflow:null},seen={}\n  var all=…"` |
| disable the probe-marker guard (`if false && scan.Probe != …`) | `TestContrastRatio_NilResultFailsClosed` **and** `_ForeignPayloadFailsClosed` (both, independently) | two FAILs, exactly the round-1 gating scenario |
| disable the vacuity guard (`if false && scan.Scanned == 0`) | `TestContrastRatio_ZeroMeasuredFailsClosed` | one FAIL |

Note the second mutation kills TWO tests, which is what a guard in series looks like from the other
side — nil and foreign payload reach the same arm by different routes, and both are pinned.
Scratch copy deleted after; `git status internal/` clean throughout.

## 2026-08-22 — council round 2: **APPROVED** (and what the advisories asked for)

`decision: approved`, `decided_by: approved with 1 advisory objection(s) — none high-severity`,
4 abstained. **`bug_historian`, the seat that gated round 1, approved with ZERO objections** — the
fail-closed identity answered it. 11 of 14 seats approve clean.

Advisory objections and what was done with each:

- **editquality [low] — the golden's extraction was described but not itself verified**; if it were
  off by a byte the test would fail permanently, "or, worse, [be] silently re-generated". **ACTED
  ON, and it was worth acting on.** Proof: parsed the pre-refactor file with `go/parser` +
  `strconv.Unquote` (the COMPILER's view of the const, not an awk approximation) and compared to
  the golden — `len=2861`, `sha256=4ec6cb73…258da7`, **IDENTICAL: true**. That digest is now pinned
  as a constant in `TestAuditJSComposition`, so a golden regenerated from post-refactor code fails
  instead of vouching for the refactor against itself. Mutation-proven: appending ONE space to the
  golden produces `sha d24adb69… — it has been regenerated, not verified`.
- **reuse_agent [medium] — a second contrast mechanism alongside `contrast_failure`** (the audit's
  CSS-source-reading producer) with no unification plan. **ACKNOWLEDGED, deliberately not merged.**
  They answer different questions: the audit SWEEPS deployed sites and files tickets (async,
  parked behind `improvement-sweep`, repair half known-defective per `bugs_closed/198`); this GATES
  one page at acceptance time and can fail a build. Merging them would put a fleet sweep on the
  acceptance path. Unification is the render-audit lane's call and belongs with `bugs_open/296`'s
  queue decision — recorded, not silently ignored.
- **reuse_agent [low] — `scripts/render_audit.py` is a THIRD copy of the WCAG maths.** True, and
  named in the code comment already (`contrast_check.go`: "two implementations of this maths
  already exist in this repo"). Out of scope here: it is a hand-run Python script on the other side
  of a language boundary; this change reduced the Go-side copies from 2 to 1 rather than adding a
  third, which was the objection's actual concern.
- **guardian [low] ×2 — a possible THIRD binary building this JS; and the shared criteria struct.**
  Enumerated properly (see the correction below): **exactly ONE binary compiles this package** —
  `cmd/browser-runner-adapter` is the only `cmd/` importer. The struct concern is answered by
  byte-identity plus the additive-only shape (new case arm, new optional field, no signature
  change) that the seat itself noted.

  > **CORRECTED, same session, minutes after writing it.** I first wrote here that "`grep -rl` for
  > the package shows `browser-runner-adapter` and `render-audit-adapter` only" — **and I had not
  > run the grep.** That is precisely the receipt-before-the-query error I logged in `WRONG_CALLS.md`
  > this morning, committed again the same day, in a paragraph answering an objection ABOUT
  > unverified claims. Caught by re-reading my own commit. What the grep actually shows:
  > `grep -rl "internal/adapters/browserrunner" --include=*.go cmd/` returns **`cmd/browser-runner-adapter`
  > and nothing else**. `render-audit-adapter` is **not a second binary at all** — it is a second
  > *deployment of the same image*: `deployments/kustomize/services/render-audit-adapter/base/deployment.yaml:58`
  > runs `docker.io/aqls/browser-runner-adapter`, tag pinned per overlay. **Both overlays currently
  > pin `v1.0.1323`** (browser-runner-adapter `:18`, render-audit-adapter `:19`), so the "80 tags
  > behind" landmine describes a state that is not today's — the skew is possible, not present.
  > The conclusion I asserted was right and the method was absent; one binary is a *stronger* answer
  > to the seat than two, and I would have known that ten seconds earlier by typing the command.

**Trailer**: the code commits carry `Council-Submitted: 7e2391ec…` (correct for pre-verdict, 098
credits them automatically); this follow-up carries `Council-Reviewed:` because the approved
verdict has now been READ, not assumed.
