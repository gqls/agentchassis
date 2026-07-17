# PLAN — the Experience Loop: plan → challenge → build → verify, end to end, without manual intervention

*Drafted 2026-07-17 from the vonc Spark game defects (owner screenshots) + the
travelling-docs machinery. Status: PROPOSED — awaiting owner reaction, nothing
built. House style: this PLAN travels; a RUNBOOK and RUNNING_NOTES join it when
work starts.*

---

## 1. What prompted this (diagnosis of the three screenshots, verified live 2026-07-17)

Three broken surfaces on vonc.com, three DIFFERENT root-cause classes:

| Surface | Symptom | Root cause (artifact-verified) | Class |
|---|---|---|---|
| `/provocations/index.html` | clicking an entry returns to the same page | archive items are runtime-filled from a template whose `href="#"` is never given a destination — per-provocation detail pages were never planned or built (`needs_page:provocation` deliberately parked since 07-12) | **journey dead-end: no destination ever specified** |
| `/tools/arena/index.html` | "Loading… DAY 0", empty Floor, unclear what to do | current `tool-arena-interface` component (23KB, stored 07-14 17:02) contains ONE script, four localStorage refs, and **no fetch of `/data/provocations.json`** (the feed itself is live, 200/5.6KB) — the page cannot load a provocation. Its travelling-doc PLAN (with acceptance criteria) exists but under the OLD subject_key `tool-arena-interface`; the page was renamed to `tool-arena` (TL-003 reconcile), so the acceptance sweep never covers the live page — the defect was invisible to the ladder | **broken artifact + orphaned criteria (rename detached the travelling doc)** |
| `/tools/gauntlet/index.html` | bullets strike through, Enter/Preview buttons dead, timer counts to nothing | by construction: `href="#"` on both CTAs, `gauntlet-interface.js` (3.9KB) does only strikethrough/timer/counter effects. Fabricated stats (12,847 competitors, 94,210 challenges, named leaderboard) on a live page. No doc_plan exists — the page predates travelling docs | **marketing mock shipped as product; fabricated numbers; never under acceptance** |

The common factor: **every check we have verifies a page or a tool in isolation.
Nothing owns the experience** — the promise a button makes, the journey a visitor
takes, the data a widget needs, the honesty of the numbers on the page. The
misdirected_cta loop (closed 07-16) guarantees a button reaches a real page; it
cannot know the page is a mock.

## 2. What already exists (reuse, don't rebuild)

From the travelling-docs workstream (all proven in production, STATUS 2026-07-16):
- **doc_plans / doc_notes** — machine-written PLAN with acceptance criteria at tool birth; NOTES from every fix/diagnosis/verdict.
- **The verification ladder** — Tier 2 static acceptance; Tier 4 behavioural (headless Chromium, desktop+mobile, interactions P2, overflow P1, screenshots-on-failure P3).
- **tool-acceptance-agent** — criteria → browser run → verdict → attributed failure → routed ticket (improve_tool / component-template-fixer) → re-verify GREEN.
- **tool_acceptance_due** — the scheduled sweep that makes it continuous.

From elsewhere in the platform:
- **Council/critic pattern** — concept-register council (3 reviewers live in prod), fixloop judge panels: independent lenses, converge-or-escalate.
- **Work-item pipeline** — discovery → triage → dispatch → handler → two-strike escalation.
- **Claims-verification machinery (V0/V1)** — bans fabricated claims; gauntlet's invented numbers are squarely its jurisdiction.
- **Site plan / reconciler, tool-generator, page pipeline** — the build side.

## 3. The Experience Loop (the new workflow)

Five phases. A–B are the "several loops of discussion and challenge"; C–E are
build/verify rounds. Owner intervention: zero during a round; the loop surfaces
one decision menu per round-boundary at most.

### Phase A — Experience spec (new agent: experience-planner)
Writes an **EXPERIENCE_PLAN** (a `doc_plans` row, `subject_type='experience'`,
e.g. `subject_key='vonc-spark-game'`). Contents, all machine-checkable:
1. **Journeys** as first-class objects: "land on provocations → open an entry →
   read the record → Enter today's Arena → file a take → see it on the Floor".
   Every step names: page, control (selector), action, and the OBSERVABLE outcome.
2. **Promise ledger**: every CTA's copy → what the destination must deliver
   (the "Enter the Gauntlet = a playable round starts" rule that would have
   failed today's mock).
3. **Data contracts**: what `/data/provocations.json` must contain, who writes
   it, when; what runs client-side only (static site — no server).
4. **MVP cut**: round-1 scope, and an explicit LATER list. Hard rule carried
   from today: a not-yet feature must be ABSENT or labelled coming-soon —
   never simulated. No dead controls, no fabricated numbers, ever.
5. Per-journey **acceptance criteria** in the Tier-2/Tier-4 grammar the runner
   already executes.

### Phase B — Challenge council (new workflow over existing council pattern)
3–5 critics with distinct lenses attack the spec; planner revises; repeat until
all pass or max N rounds (then escalate the disagreement, not the whole doc):
- **journey-completeness critic** — every clickable element has a named destination and outcome; no step ends in "#".
- **feasibility critic** — buildable with current components/tool-generator? data available at runtime? works with the static-hosting constraint?
- **honesty auditor** — no invented stats/users/social proof anywhere in the spec (claims-verification rules).
- **MVP referee** — cuts scope; challenges anything that isn't needed for the core loop to be playable.
Every round appends to RUNNING_NOTES (machine-maintained, travelling-docs rule).

### Phase C — Contract-first build
Spec rows become site-plan pages/sections/tools **with their acceptance criteria
attached at birth** (exactly what tool-generator already does for tools — extended
to pages and journeys). Build the MVP cut only.

### Phase D — Journey acceptance + self-heal
The existing ladder runs it. Two extensions:
- **Tier-4 journeys**: browser-runner follows multi-page paths (click → navigate
  → assert), not just single-tool interactions. (Its P2 interaction machinery is
  the base; journeys are a sequence of P2 steps with navigation.)
- **New failure scope** alongside tool/chrome: `data` (feed missing/stale) and
  `plan-gap` (a journey step has no owner) → `needs_experience_replan` items
  route BACK to Phase A instead of to a code fixer.
Everything else is the proven flow: attributed failure → routed ticket → fix →
re-verify GREEN → verdict note.

### Phase E — Feature rounds
After MVP is GREEN end-to-end, the loop picks the next slice from the LATER
list and repeats B→D. Images/copy polish are rounds, not blockers.

## 4. Guard rails this incident demands (independent of the pilot)

1. **Page-ownership marker** — a page owned by a tool/widget (`tool-arena`)
   carries a marker that makes generic rebuild/re-render paths REFUSE (TL-001
   became a habit-rule; it must be mechanical). The arena clobber is the proof.
2. **Rename moves the travelling docs** — page/tool renames must re-key
   `doc_plans`/`doc_notes` (or the sweep must resolve aliases). The orphaned
   arena PLAN is the proof.
3. **Dead-control static check (Tier 2)** — an interactive element with
   `href="#"`/no handler on a deployed page is a failure, not a style choice.
4. **Fabricated-numbers check** — page copy with quantitative claims must trace
   to a data contract (claims-verification jurisdiction; gauntlet is the proof).

## 5. Pilot: the Spark daily-provocation game (vonc)

Round 0 (this workflow's first real run) folds in the "fix what's broken" ask:
- Phase A writes the EXPERIENCE_PLAN for the game that vonc's copy already
  promises: daily provocation → file a position → the day's record archives.
- The council will inevitably face the one genuine product decision: **is the
  Gauntlet a real playable round in MVP, or honestly demoted (coming-soon) until
  a later round?** Default recommendation: minimal-real — a playable timed
  round against the daily provocation, client-side scoring, no fake leaderboard;
  demote only if the feasibility critic proves it can't be honest and small.
- Phase C rebuilds: arena widget to spec (fetching the live feed), provocation
  detail pages (un-parks `needs_page:provocation` WITH a spec this time),
  gauntlet per the council's ruling; strips every fabricated number.
- Phase D wires all three under `tool_acceptance_due` with journey criteria, so
  a future clobber/regression is caught by the next sweep, with screenshots.

## 6. What is genuinely NEW to build (everything else is reuse)

| Piece | Size | Notes |
|---|---|---|
| `experience-planner` agent + EXPERIENCE_PLAN schema | new agent, existing doc_plans table | subject_type='experience' |
| Challenge-council workflow (4 critics + revise loop) | new agent defs over the proven council pattern | converge-or-escalate, max rounds |
| Tier-4 journey runs (multi-page sequences) | extension to browser-runner + acceptance-agent | P2 interactions + navigation |
| `needs_experience_replan` routing + `data`/`plan-gap` scopes | small: new item type + judge vocabulary | |
| Page-ownership rebuild guard | small Go guard + page marker | kills the TL-001 class |
| Doc re-keying on rename | small: reconciler hook | kills the orphaned-criteria class |
| Dead-control Tier-2 check | small: one static check | |

## 7. Open questions for the owner (only these; everything else defaults)

1. Gauntlet in MVP: minimal-real (default) or honest demotion?
2. Per-provocation detail pages: static pages built daily by the pipeline
   (default — consistent with the existing daily JSON emitter) or a single
   archive page with client-side detail rendering?
3. Appetite for the pilot to run fully autonomous end-to-end on vonc (default:
   yes — it is the test site), with the usual artifact-verified checkpoints in
   RUNNING_NOTES rather than approval gates?
