# PLAN — the `contrast_ratio` acceptance check (bugs_open/131's framework half)

**Lane opened:** 2026-08-22 · **Bug:** `bugs_open/131_HANDOFF_2026-07-28_vonc_gauntlet_usability_audit_from_a_real_visitor.md`
(⚠ 131 is an AMBIGUOUS NUMBER — `bugs_closed/131` is an unrelated og-image case. This lane is the
**vonc gauntlet usability audit** slug, and specifically its FRAMEWORK residue.)

## What this lane is, in one paragraph

Bug 131's items A–G were all fixed in July. On 2026-08-22 this lane re-measured the live pages and
found the contrast fixes have **decayed** (evidence below), because the platform still has **no check
that can fail a page for unreadable text** — the exact gap 131's own closing section names, the
`contrast_ratio` check proposed in `experience_loop/HANDOFF_2026-07-28_appeal_dimension.md` and never
built by anyone. This lane builds that check. It does NOT repaint vonc (the gauntlet lane's surface),
does not touch the render-audit sweep's parked queue (`bugs_open/296`), and does not decide the
component-painted-ground question (`bugs_open/212` §8, an owner decision).

## Why the robust fix is a CHECK, not another repaint (the decision, with its reasons)

1. **The point fixes decayed within three weeks, silently.** 131-A was fixed 2026-07-28 at a
   measured 3.31:1 (`--color-stage` #f59e0b on #6d28d9). Measured 2026-08-22: the section background
   has churned to `rgb(124,60,255)` (#7c3cff) and the same accent now reads **2.48:1 — below the 3.0
   large-text bar it was fixed to pass**. Item D (readable column) was fixed at 83% of a phone
   viewport; today it measures **65.6%** — worse than the 74% the owner originally complained about.
   Nothing noticed either regression, because nothing watches.
2. **Every owner-visible incident of this class was found by a person.** fundamentallyai 07-27, vonc
   07-28 (this bug), ai-agent-orchestration 08-06 (hand-run script), idea.uk 08-19. Four sites, zero
   machine catches at the gate.
3. **The detection that exists cannot GATE.** The weekly render-audit rotation (VIZ-015) measures
   live pages and files `contrast_failure` tickets — but it cannot fail an acceptance run, its drain
   is parked behind a disabled sweep (`bugs_open/296`: 225 items), and its repair half has a
   destructive history (`bugs_closed/198`, the idea.uk clobber). The Tier-4 acceptance ladder — the
   only machinery that can fail a page — has eight check types and **not one can see colour**
   (`LANDMINES.md:1523`; register `DES-054` "aspirational"; verified 2026-08-22 at
   `run_checks_action.go:554-557`).
4. **The precedent is this bug's own item B.** `no_horizontal_overflow`'s clipped-overflow clause was
   filed in 131, built in `run_checks_action.go` in one commit (`5042d5ecb`), council-APPROVED,
   rolled, witnessed on a real page, and has been catching real cuts since. `contrast_ratio` follows
   that path exactly.

## Design (settled 2026-08-22, this session)

New Tier-4 check type **`contrast_ratio`** in `internal/adapters/browserrunner/run_checks_action.go`.

- **Measurement**: one in-page probe via the existing `browserPage.Evaluate` seam (the seam built for
  `render_audit`). The WCAG maths + effective-background walk (`parseRGB/lum/ratio/over/effBG`) is
  **factored out of `auditJS` into one shared JS const** used by both the audit and the check —
  a third drifting copy of the maths is the alternative and is refused (there are already two:
  `auditJS` and `scripts/render_audit.py`).
- **Semantics** (mirrors `no_horizontal_overflow`, deliberately — no new patterns):
  - Measures every visible text-bearing element on the page (same filters as the audit: skips
    `display:none` / `visibility:hidden` / `opacity:0`, zero-size rects, <2 chars of direct text).
  - Per-element threshold: WCAG AA — 4.5:1 body, 3.0:1 large (≥24px, or ≥18.66px bold). An explicit
    `min_ratio` on the check replaces both (a fence-visible, deliberate relaxation or tightening —
    the honest place for a design exception).
  - **`over_image` grounds never fail the check** (backdrop under an image/gradient is approximate —
    the same rule the audit's `ContrastFirm` applies, and the false-positive containment 131-B's
    history demands: a false acceptance failure becomes an `improve_tool` fixer aimed at a correct
    page, `bugs_open/126`). Their count is reported in the detail as unmeasured.
  - **Attribution, not scoping**: like overflow, the probe measures the whole page and attributes
    each failure inside/outside the tool container (`toolContainer(doc, ch)`, per-check `selector`
    override). `pass:false` on any firm failure; `Scope`/`Culprit`/`CulpritSelector` name the worst
    offender so the judge routes a tool ticket or a chrome one.
- **Vocabulary**: `experienceCheckTiers["contrast_ratio"] = 4` (Tier 4 ONLY, necessarily — computed
  colour does not exist in static HTML; same class as `has_visible_area`);
  `experienceCheckTypeFields["contrast_ratio"] = {"min_ratio"}`; `MinRatio float64
  \`json:"min_ratio"\`` on `criteriaCheck` (an untagged field is silently dropped at unmarshal —
  LANDMINES:8645). The lockstep tests harvest the new `case` literal and hold the tables to it.
- **NOT Tier-2**: no static evaluator arm, no confirming/refuting classification needed.

### Architecture-scope assessment (RFC_022 / 2026-07-29 ruling #1)

Not architecture-scope, measured against both rulings: it is **opt-in** (a check runs only where a
criteria fence names it), the unsafe side is the **default** (absent = not run), and **zero live
consumers name it** — measured 2026-08-22, three queries: `doc_plans` fences using it as a check
type (`body ~ '"type":\s*"contrast_ratio"'`) → **0**; active `agent_definitions` whose config names
it → **0**; `sql_for_agents/*.sql` seeds naming it → **0**. (Four `doc_plans` rows contain the bare
substring — they are the smart-contrast and oklch-picker TOOLS' own page prose, not consumers of a
check type; the discriminating query is the typed one.) It adds a refutation
axis to a lane whose job is already refutation (Tier-4 fails pages today); it changes no existing
guarantee. Normal council gate, submitted before/alongside the commit.

## Phasing — and why the rollout order is load-bearing

**Phase 1 (this session): code + tests + register + council + commit.** The check exists in the repo
and in the vocabulary tables, registered as TL-049. INERT until the browser-runner-adapter image is
rebuilt and rolled — and that inertness is a LANDMINE (`LANDMINES.md:512`): a fence naming a check
type the running binary does not know is **skipped**, and a wholly-skipped fence can record a green
verdict. **So Phase 1 deliberately does NOT advertise the type to any planner prompt or author any
fence.**

**Phase 2 (after the adapter rolls): witness, then advertise.**
1. Prove the deployed binary knows the type — grep the pod for the long detail sentence the new arm
   emits (recipe in RUNBOOK; `strings` does not exist in this container, and short literals never
   reach rodata — use the long marker + positive/negative controls, 131-B's own discipline).
2. **Witness on a known-bad live page**: vonc's own gauntlet page carries `gi-eyebrow` at 1.66:1 and
   `gi-rules-label` at 1.76:1 today (measured 2026-08-22, by probe AND by screenshot). A manual
   acceptance work item there (the `manual_131b_witness` precedent, shape `043bfe1d`) must return
   `pass:false` with an attributed culprit. A clean page in the same batch is the control.
3. Only then: seed updates advertising the type — the planner vocabulary lists and the
   `259_experience_approval_council.sql` deferral-honesty list (which hard-codes the executable
   check types; a new type absent from it makes the council object to a legitimate deferral).
   Phase 2's seed edits are a separate coherent task with its own council round if they change agent
   behaviour materially.

**Phase 3 (separate decisions, NOT this lane):** whether standing tool fences gain a `contrast_ratio`
check by default (an estate-wide authoring decision — owner/council), and vonc page-side repairs
(gauntlet lane's design pass owns the surface; this lane's 08-22 measurements are contributed into
`bugs_open/131` and the design-pass handoff so the pass knows what regressed).

## What closes what

- This lane closes when the check is **live and witnessed** (131-B's bar: fixed, live, AND
  demonstrated on a page that IS bad).
- Bug 131 itself does NOT close with this lane: D needs an owner/design-pass decision (it was never
  decided — three records contradict; see the 08-22 section appended to the bug), and H's residue is
  the owner's own distribution leg. Who closes 131 is recorded there.

## Decisions taken and their reasons (running)

- 2026-08-22 · **Reuse `Evaluate` rather than adding a browserPage method** — the interface comment
  says Evaluate exists for exactly this shape ("whole measurement is one in-page probe"); a new
  method would churn the fake page in every test for no capability.
- 2026-08-22 · **Fail on chrome failures too (attribute, don't exempt)** — mirrors overflow; a person
  cannot read chrome text either, and `Scope` is how the judge routes it. Inventing a
  tool-only-fails rule would be a new pattern with a silent pass class.
- 2026-08-22 · **Default AA, not an invented floor** — the rail in the appeal handoff: a check must
  reduce to a published number. 2.x "invisibility floors" are made-up; a fence that wants one writes
  `min_ratio` where a reviewer can see it.
- 2026-08-22 · **`over_image` cannot fail** — approximate grounds produce plausible-but-wrong ratios;
  the audit already splits `ContrastFirm` for the same reason, and a false failure arms a fixer at a
  correct page (126).
