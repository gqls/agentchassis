# PLAN — Behavioral QA loop for tools & games

Status: proposed (2026-06-06). A standalone maintenance/QA loop, separate from the adoption and build pipelines. Extends the **existing** three-tier tool-quality model (020_tool_lifecycle, 019_tool_library) by building out Tier 3 properly, and adds the missing **games** parity. Reuses the check → work-item → improver pipeline; does not introduce a parallel system.

---

## 1. Why

The tools and games are good on first contact and fail under deeper use. Observed on gamesdesign.co.uk:

- **Jelly Invaders** plays correctly, then **skips rows and behaves erratically after a while** — a *temporal/runtime* degradation, invisible to a single render or screenshot.
- **P2P networking**: Firefox↔Chrome on desktop works; a phone browser connects to the host (star topology) and can send to the host, but **the host's replies only reach the desktop, not the phone** — a *directional relay* failure to a *specific device*, invisible to single-context testing.
- The rest "looked OK" but were **not tested hard, not on mobile, not across variants, not across browsers**.

None of these are catchable by Tier 1 (structural Go regex) or Tier 2 (one-shot LLM code review), and they go beyond even the *planned* Tier 3 (render + screenshot + console at fixed viewports). They need: running the artifact **over time**, **driving** it with synthetic input, **multiple simultaneous browser contexts** across **engines and devices**, and **per-variant** coverage — with an oracle that knows what "still working" means.

---

## 2. What already exists (reuse, don't rebuild)

| Piece | Reuse for QA loop |
|---|---|
| `site_work_items` + `build-dispatch-loop` + `build-pipeline-trigger` | Same pipeline carries QA findings and dispatches fixes. |
| Tier 1 `tool_health` (`check_tool_health.go`) → `improve_tool` | Keep as the cheap front gate; QA loop is the layer *after* it. |
| Tier 2 `tool-auditor` (Sonnet code review) → `improve_tool` / `needs_human_review` | Unchanged; behavioral findings feed the same `improve_tool`. |
| `tool-improver` (`load_tool`, spec `component_id`/`issue`/`check`/`fix_suggestion`) | The fix handler for tool defects we find. No new fixer needed for tools. |
| Fork model (site owns its tool/game copy) | QA tests and fixes the **site's fork**, so fixes don't cascade unexpectedly. |
| Discovery-check action pattern (Go check → emits work item) | New checks (`check_*_qa.go`) follow the same shape. |
| `improvement-sweep` (600s) / scheduling pattern | The QA loop is a **separate sweep** with its own (slower) cadence, modelled on this. |
| Planned headless-browser pod (020 §Three-tier: isolated K8s, Chromium, screenshots→S3, `visual_test_tool`) | This plan *is* the build-out of that pod, generalised. |

**New pieces required:** an isolated Playwright pod (3 engines), a QA orchestrator agent, a behavioral test harness (duration + drive + invariants), a multi-context (multiplayer) harness, a **games** lifecycle mirroring the tool lifecycle, and the oracle layers in §5.

---

## 3. Where it sits in the flow

A dedicated loop, parallel to (not inside) the build/adoption flow:

```
qa-sweep (scheduled, slow cadence)              ← own sweep, queue-depth gated
  └─ picks least-recently-QA'd site with deployed tools/games
  └─ qa-orchestrator  (every agent is an orchestrator)
       ├─ spawn tool-behavioral-tester  (per tool fork, per variant)
       ├─ spawn game-behavioral-tester  (per game fork, per variant)
       └─ spawn multiplayer-tester       (per networked artifact)
            └─ each tester drives a session in the Playwright pod,
               returns a structured QA report, and emits:
                 improve_tool / improve_game   (high-confidence, auto-fix)
                 needs_human_review            (ambiguous)
  └─ triage items → call build-dispatch-loop  (same as improvement-loop step 10)
```

Triggers (mirror 004): **scheduled** `qa-sweep` (slow — see §8); **post-deploy hook** (QA an artifact once, shortly after it first deploys); **manual** `./trigger-qa.sh <site> <domain> [tool|game|all]`. Conventions held: workflows stay thin, complexity in Go actions; sub-agents spawned rather than sub-workflows; agent creation + messages logged; testers reply to the orchestrator's (parent) response topic.

---

## 4. The headless pod (infra)

- Separate Kubernetes deployment, isolated, with its own CPU/memory limits (Chromium/Firefox/WebKit are heavy and run **arbitrary generated HTML** — treat as untrusted: locked-down network egress, no host mounts, per-session timeouts, recycle contexts).
- **Playwright** (not Puppeteer) so the same harness drives **Chromium, Firefox, and WebKit** from one API.
- Exposes a session API the agent calls: `{ url, engine, viewport, deviceProfile, driveScript, durationMs, sampleIntervalMs, contexts }` → returns `{ consoleLog, jsErrors, screenshots[], videoTrace, stateSamples[], metrics }`. Screenshots/video/traces stored in S3 (as the planned Tier 3 already intended); report rows in Postgres.

---

## 5. The oracle problem (how we know it's still working)

Generated artifacts have no hand-written test suite, so correctness comes in three layers, cheapest first:

**(a) Generic invariants** — hold for *any* interactive artifact; deterministic, no LLM:
- No uncaught JS errors / unhandled promise rejections during the session.
- No `NaN`/`Infinity`/`undefined` rendered into the DOM where a value is expected.
- Console error count does **not grow unbounded** over the session (catches slow leaks/log spam).
- DOM node count and JS heap stay **bounded** over time (catches runaway allocation).
- Frame budget: `requestAnimationFrame` cadence doesn't **collapse** (catches the "erratic after a while" class).
- No layout overflow / clipped content at the target viewport.

**(b) Type-specific assertions** — derived from the artifact's declared spec/intent (the same spec the generator used):
- *Calculator/estimator*: output changes when inputs change; output finite; matches a reference computation for a few known input pairs.
- *Simulator*: monotonic/!monotonic series behave as declared; bounds respected.
- *Game*: entity/row counts evolve within declared rules (Jelly Invaders: row structure changes only by the rules — **no sudden row skips**); score behaves as expected; the loop keeps advancing; lose/win states reachable and terminal.

**(c) LLM-as-judge over a time-series** (Tier 2.5, extended to *temporal*): feed the ordered screenshots/short video + console log to Sonnet — "does behaviour glitch, freeze, teleport, or degrade across this capture?" Catches the subtle cases invariants miss. Higher cost, so gated to artifacts that passed (a)/(b) or that (a)/(b) flagged ambiguously.

**Auto-fix gating:** only **high-confidence deterministic** findings (a JS error; a hard invariant violation; a missing relay path) create `improve_tool`/`improve_game` directly. LLM-judge "possible" findings → `needs_human_review`. This protects against "fixing" behaviour that was actually correct.

---

## 6. The hard cases, concretely

**Temporal degradation (Jelly Invaders).** Load → drive with a synthetic autoplay script → run 30–120s (or N frames) → sample state + screenshot every interval → evaluate invariants across the whole series. Acceptance: row/entity structure only changes per declared rules; no frame-cadence collapse; no error growth. This is the cell that fails today.

**Cross-browser.** Run the same session in Chromium + Firefox + WebKit; diff pass/fail; flag engine-specific failures (the P2P bug surfaced partly as engine/device-specific behaviour).

**Mobile.** *Tier A (now):* emulated context — viewport, `deviceScaleFactor`, touch events, mobile UA — in Playwright (cheap, catches layout + touch-target + hover-only bugs). *Tier B (later):* real-device cloud (BrowserStack / Firebase Test Lab) for what emulation can't reproduce (real touch, real GPU, real NAT for WebRTC).

**Multiplayer / networked (P2P).** A dedicated multi-context harness: spin up **N contexts** (mix of engines + emulated devices) in the pod, join them into one session (host + clients), then assert the **full relay matrix** in every topology the artifact claims:
- star: every `client→host` delivered **and** every `host→client` delivered (the observed failure is the `host→client[mobile]` cell);
- mesh (if claimed): every `client↔client` delivered.
WebRTC needs signalling — provide a tiny in-pod signalling rendezvous (or use the artifact's own) for same-pod contexts. Caveat: same-pod contexts validate **relay/topology logic** but not real-NAT traversal; real-network correctness is Tier B (device cloud / real endpoints). Acceptance for the gamesdesign case: 3-context session (host-desktop, client-desktop, client-mobile) delivers all host→client and client→host messages.

**Variant coverage.** Enumerate variants from the artifact definition (input_schema options, modes, difficulty). Run the suite per variant, but **cap combinatorics** (representative sampling, not full cartesian). Today only the default variant is ever exercised.

---

## 7. Games parity (mirror the tool lifecycle)

Games currently have no quality lifecycle. Add the analogues, reusing tool shapes wherever possible (confirm first whether games are modelled as `component_level='game'` / `page_type='game'` so the tool pipeline can be forked rather than rewritten):

| Tool piece | Game analogue (new) |
|---|---|
| `tool_health` check | `check_game_health.go` (deployed, html, `<script>`, canvas/loop present, responsive) |
| `tool-auditor` | `game-auditor` (LLM review: loop integrity, state machine, input handling, entity bounds, win/lose, memory) |
| `tool-visual-tester` (planned) | `game-behavioral-tester` (shares the §4 harness; game-specific invariants + autoplay drivers) |
| `tool-improver` | `game-improver` (fixes `improve_game`; `load_game` mirrors `load_tool`) |
| `evaluate_tools` / `improve_tool` / `audit_tool` | `evaluate_games` / `improve_game` / `audit_game` / `behavioral_test_game` |

---

## 8. Scheduling & cost

QA is expensive (browser pods, multi-context, LLM-judge), so it is **not** on the 600s improvement cadence:
- `qa-sweep`: slow cadence (e.g. daily), queue-depth gated, picks least-recently-QA'd site; tracks `last_qa_at` per artifact (e.g. in `content_components` settings or a `qa_runs` table).
- **post-deploy**: QA a tool/game **once** shortly after first deploy (highest value — catches the worst before anyone sees it).
- **manual** `./trigger-qa.sh` for targeted runs.
- Concurrency: cap simultaneous browser sessions; rate-limit per node; reuse the atomic `claim_work_item` so qa-sweep and improvement-loop don't collide.

---

## 9. Reporting

- Per-artifact QA report: pass/fail per engine × viewport × variant, invariant violations, LLM-judge notes, links to S3 screenshots/video/trace.
- Surface in the admin dashboard (012): per-site QA status, last run, open `improve_*` items.
- Findings flow as work items (§3), so existing dashboards/triage already see them.

---

## 10. Phasing (each phase shippable on its own)

- **Phase 0 — pod + baseline (delivers the long-planned Tier 3).** Isolated Playwright pod (Chromium first). A `qa-runner` action: load URL, run a drive script, capture console + screenshots, return structured result. Wire `behavioral_test_tool` → emits `improve_tool` on hard failures (JS errors, missing script execution).
- **Phase 1 — temporal + interactive invariants (Chromium, desktop).** Run-for-duration + synthetic drive + the §5(a) generic invariants. **Closes the Jelly-Invaders class.**
- **Phase 2 — cross-browser + emulated mobile.** Add Firefox + WebKit; add emulated mobile contexts (viewport/touch/UA). Closes engine- and mobile-layout divergence.
- **Phase 3 — multiplayer/networked harness.** N contexts, relay matrix, topologies, in-pod signalling. **Closes the P2P directional class** (emulated; real-NAT deferred).
- **Phase 4 — variants + temporal LLM-judge.** Per-variant runs (sampled) and §5(c) time-series judgement.
- **Phase 5 — games parity agents.** `check_game_health`, `game-auditor`, `game-behavioral-tester`, `game-improver`, `improve_game` — most of the harness reused from tools.
- **Phase 6 (optional/future) — real-device cloud.** Real mobile + real WebRTC NAT traversal for the cases emulation can't reach.

Recommended first cut: **Phase 0 + Phase 1** — smallest infra investment that already catches the runtime-degradation class and finally lands the Tier 3 the lifecycle has been pointing at.

---

## 11. Open questions (decide before building)

- Are games already modelled (`component_level`/`page_type`, fork model) so the tool pipeline can be forked rather than rewritten? (Check schema first.)
- Signalling for WebRTC tests: in-pod shim vs the artifact's own signalling server vs a stub.
- Device-cloud budget/appetite for Phase 6 (real mobile + real NAT).
- How variants are declared per artifact, so the runner can enumerate them.
- Oracle confidence thresholds: where exactly the auto-fix vs HITL line sits per finding type.
- Where `last_qa_at` and QA reports live (extend `content_components.settings` vs a new `qa_runs` table).

---

## 12. Risks

- **Resource cost & isolation** — browser pods are heavy and render untrusted HTML; sandbox hard (egress lockdown, no mounts, recycle, timeouts).
- **Flaky tests** — timing/animation nondeterminism; mitigate with retries, tolerance windows, and a quarantine state before a finding becomes an `improve_*`.
- **Oracle false positives** — never auto-"fix" correct behaviour; gate auto-fix on deterministic high-confidence signals, route the rest to HITL.
- **Test maintenance** — type-specific assertions and drive scripts are per-archetype; keep them in the artifact archetype definitions so they evolve with the generators.
