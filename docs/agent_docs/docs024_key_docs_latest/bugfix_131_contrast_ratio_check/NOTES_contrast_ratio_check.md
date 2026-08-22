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
guardian's two-consumer worry: an inert recomposition needs no coordinated rollout whatever the
two deployments' tags are. (⚠ **I wrote "`render-audit-adapter` runs this package ~80 tags behind"
here and it is NOT today's state** — see the correction further down: ONE binary compiles the
package and both overlays pin `v1.0.1323`. The argument does not depend on the skew; the claim was
carried from a landmine without re-measuring it.)

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

### Misstep: `landmines-sync.py --apply` run before `landmines-verify-dispatch.sh` (2026-08-22)

CLAUDE.md documents this exact ordering trap and I hit it anyway when appending the re-measurement
to the 80-tags entry: `--apply` consumes the "new entry" status, after which the verifier never
checks the entry. Recovered with the documented per-entry remedy,
`./scripts/trigger-landmine-verifier.sh '<slug>'` (corr `e567ad52`); the earlier blind-pass entry
went through `landmines-verify-dispatch.sh` correctly (corr `b19e22da`). The rule in one line:
**append → `landmines-verify-dispatch.sh` (never `--apply` first)**, and if you already applied,
trigger per slug.

### The two landmine-verifier verdicts read NEEDS_HUMAN_REVIEW — and that is INDEX STALENESS, not doubt (2026-08-22 11:13–11:15Z)

Both entries this lane armed came back `NEEDS_HUMAN_REVIEW`. Read the reason before treating either
as suspect — a future reader scanning verdicts will otherwise discount two sound entries:

- **the blind-pass entry** (corr `b19e22da`): *"the core remediation file `contrast_check.go` and its
  symbols (`runContrastRatio`, `contrastProbeMarker`)"* were not found. They exist — I committed them
  today. The verifier's own footer says it answers about **indexed commit `1b4f836f`, committed
  2026-08-21 19:02 UTC, "the last pushed tip, not the present tree"**. A code index a day behind
  cannot see a file created this morning; the verdict is about the index's reach, not the entry.
- **the 80-tags entry** (corr `e567ad52`): the useful half — *"the Go-level footprint resolves and is
  **consistent with the entry's structural claim (one binary, two deployments)**"*. That is
  independent corroboration of today's correction, arrived at from the code index rather than from
  my grep. What it could not check — makefile targets and kustomize `newTag` pins — is because the
  corpus is `.go` only (8,417 rows, "kinds with NO rows"), and those live in YAML and the makefile.

**The transferable bit:** this verifier can only ever confirm or refute claims whose evidence is Go
symbols at the last PUSHED tip. For an entry whose load-bearing facts are YAML pins, shell targets or
same-day code, `NEEDS_HUMAN_REVIEW` is the *correct and uninformative* answer — it is the tool
declining to guess, which is the behaviour we want, but it means **the verdict carries no information
about the entry's truth** and must not be quoted as if it did. Both entries' own evidence is in this
lane's NOTES and the commits.

## 2026-08-22 (evening) — the adapter rolled, the check is LIVE, and the witness earned its keep

**Deployment PROVEN at the artefact, not inferred.** `browser-runner-adapter` and
`render-audit-adapter` both on `v1.0.1326`, pods started 15:10:5xZ (overlay bumped 1323→1326 by
another session's build). Provenance line: `git_commit 27b932acad15740da850d71799e01191010a3713`.
`git merge-base --is-ancestor` TRUE for `2611b0b16` (the fail-closed fix) and `de7806e55` — **with a
negative control**: HEAD is correctly NOT an ancestor, so the test can say no. Binary needles in one
exec (no `strings` in this image): new marker **1**, probe sentinel **1**, 131-B positive control
**1**, nonsense negative control **0**.

**The witness.** Fired `run_checks` DIRECTLY at the adapter topic (`witness_contrast_ratio.sh`) with
the reply on a throwaway topic — deliberately NOT a tool-acceptance run, because acceptance invokes
the judge, which files `improve_tool` items, and the target is the gauntlet lane's live surface with
a design pass queued for exactly these tokens. Reply topic did not auto-create (only 3 KafkaTopic
CRs exist), so the result came from the adapter's own completion log:

```
run_checks complete  run_id=65c90571…  function=witness-131-contrast  profiles=[mobile]
passed=3  failed=0  skipped=0
```

**`skipped:0` is the deployment proof** (the binary knows the type — an unknown type is skipped).
**`failed:0` was WRONG**, on a page I had screenshot-confirmed as unreadable that morning.

### What the witness found: the check could not see its own founding case

Diagnosed with the probe **extracted from source at runtime** (`scripts/dump_probe_test.go.txt` →
`scripts/run_deployed_probe.py`), so scan-vs-deployed drift is impossible by construction — the
131-B session's own method. Before the fix, on the live page: **scanned 33, failures 10, firm 0,
overImage 10** — `gi-eyebrow` 1.66:1, `gi-title-accent` 2.48:1, `gi-rules-label` 1.31:1, ratios
matching the morning's independent Playwright scan. **All ten discarded.**

Cause, measured by dumping the ancestor chain: `effBG` sets ONE flag for "a background-image
appeared anywhere", and over-image can never fail. The gauntlet section is an **opaque**
`rgb(124,60,255)` beneath `radial-gradient(rgba(139,0,0,0.35) → rgba(0,0,0,0))` and the rules card
adds `linear-gradient(rgba(251,191,36,0.08), rgba(220,38,38,0.08))` — **`url()` appears nowhere.**
So the backdrop was never unknown, only decorated, and the check threw away the very defect it was
built for. **That is the PASSES-WHILE-BLIND family again — the second instance in this lane after
round 1's nil-result defect, and this one I introduced by inheriting the audit's rule without
asking what it conflates.**

### The fix, and why it is conservative rather than merely less blind

Split by what is knowable. **UNBOUNDED** (a `url()` image, a gradient whose stops are named/hex
colours, or no opaque base so `effBG` substituted mid-grey) → unchanged, reported, never fails.
**BOUNDED** (opaque base + translucent rgba stops) → the true backdrop lies in a known range, so
judge on the reading **most favourable to the page** (best contrast against the base or any single
stop composited over it) and fail only when nothing in that range saves it, reporting that best
case. Probe-local, NOT in `contrastMathsJS`: the shared kernel stays byte-identical, so the render
audit's behaviour is provably untouched (sha-pinned golden still green).

After, same page and method: **9 FIRM, 0 unbounded** — and visibly generous where it should be
(`gi-eyebrow` 1.66 → **2.37** best case; `gi-title-accent` now clears 3.0 on its best reading and is
**correctly not flagged**).

**Four induced controls** (`scripts/induced_backdrop_controls.py`), each able to come out otherwise:

| case | expected | got |
|---|---|---|
| flat opaque, bad colours | FIRM flat | FIRM flat 1.34:1 |
| gradient over opaque, bad | FIRM bounded | FIRM bounded 1.78:1 |
| `url()` image, bad | approx, never fails | approx 1.34:1 |
| gradient over opaque, good | not flagged | not flagged |

The discriminating pair is rows 1 and 3: **identical colours, identical 1.34:1, opposite verdicts.**
The classification is doing the work, not the colours — which is the control that would have caught
this defect at design time had I run it then.

**Submitted as council round 3** (`RESUBMIT_CORR=7e2391ec`, run orch `956e2326`) because it widens
what an already-approved mechanism refuses.

**Still owed after this lands:** a rebuild of `browser-runner-adapter` (the refinement is inert on
`v1.0.1326`), then re-run the witness expecting `failed:1` with an attributed culprit, then the
Phase-2 seed/planner vocabulary work.

### Blast radius of the bounded-backdrop refinement — measured fleet-wide 2026-08-22

Ran the refined probe (extracted from source) against **every live fleet domain** — 24 domains, list
taken from `sites JOIN pages WHERE deployed_at IS NOT NULL`, NOT from recall (see the WRONG_CALLS
row below about what a hand-typed list cost me). Homepage per domain, mobile 390×844.

**The refinement is almost inert fleet-wide: 3 newly-judged elements, all on one site.**

| | count (as of 2026-08-22) |
|---|---|
| firm failures across 24 live homepages | **145** |
| of those, NEWLY judged by the refinement (`gradientBounded`) | **3** — all on vonc.com |
| sites gaining zero new failures from the refinement | **23 of 24** |
| the diagnosed page itself (`vonc.com/tools/gauntlet/`) | 9 firm, all 9 newly judged |

So the answer to "this widens what an approved mechanism refuses" is: measured, the widening is
**3 elements on 1 of 24 sites** plus the page it was diagnosed from. The new branch is rare and
specific — the same shape 131-B found for its own clipped-overflow clause (86 clean / 8 flagged /
exactly 1 on the new branch), and for the same reason: it keys on a narrow structural condition.

**The much larger finding is the 145, and it is NOT about this refinement** — 142 of them fail under
the rule as already shipped in `v1.0.1326`. **If `contrast_ratio` were added to standing fences
today, it would fail most sites.** That is a Phase-3 input the owner/council will need: adoption is
not a switch to flip quietly. Worst offenders on this sample: `loanandmortgagecalculator.co.uk` 30,
`vonc.com` 33, `gamesdesign.co.uk` 15, `idea.uk` 14, `loancash.co.uk` 13.

⚠ **145 is a FLOOR, and homepage-only is exactly the sampling error `bugs_closed/122` warns about**
(*"a homepage is not a sample of a site"*; its own homepage-vs-sitemap runs differed by ~2 orders of
magnitude — robot-hands 3 → 193, dartsonline 1 → 125). Do not quote 145 as a fleet total; it is
24 pages, one per site, on one profile, on one date.

**`https://pool-energy-utilities.internal` was excluded** (not publicly resolvable), and that
exclusion is stated rather than silent — a dropped row in a scan table reads exactly like a clean
one.
