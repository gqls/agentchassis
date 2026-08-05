# HANDOFF — staged_component_build D10 continuation (fresh chat starts here)

**Supersedes `HANDOFF_2026-08-05_continue_here.md`** — that file's calibration tranche is
DONE and the owner ruled the same day: **D10, exhaustive clearance (option c)**, recorded
in the PLAN. This file is the production line's state and its remaining work-list.

## 1. Read these first

| doc | why |
|---|---|
| `README_where_we_are.md`, last three entries | plain prose, fastest way in |
| `SUMMARY_2026-08-05_calibration_tranche_done.md` | the milestone this continues from |
| `PLAN_2026-07-30_staged_component_build.md` — **D10** | scope guards: no fence for a pageless tool (BROKEN A); levels beyond section/tool out of scope |
| `NOTES_staged_component_build.md`, all `## 2026-08-05` entries | every batch's table, every line rule, every CID |
| RUNBOOK §8–§10, §13 | unchanged and still binding |

## 2. State — proven and persisted (all S6-green in-cluster with negative controls red)

**35 subjects done end-to-end 2026-08-05:** tools `tool-fuel-cost-estimator`\* and
`tool-loan-vs-savings`; sections hero, call-to-action, generic-text-block, article-body,
features, ported-prose, hero-about, info-card-grid, content-block-about, tool-cta,
hero-contact, differentiators, faq, about-content, contact-form, contact-info,
brief-explanation, hero-case-studies, hero-services, hero-tool, services-grid,
system-stats, testimonials, tool-guide-intro, tool-list, use-cases-list, ported-page,
content-listing, departments-grid, evidence-chart, guide-list, leadership-team,
mechanism-flow. (\*fce's S6 is DEFERRED, stated in its PLAN: gaswholesalers 404s
`/assets/images/logo.png` fleet-wide; dispatch the moment the logo serves.)

## 3. The line (all instruments committed in `scripts/`)

Per batch of ~6–10 subjects (~9 min/subject measured):
1. **Placements**: single-instance per page (`HAVING count=1`), avoid sites with known
   chrome 404s (dartsonline, gaswholesalers, idea.uk hero.jpg/logo class; lendzy origin
   flaky). Curl-verify serving AND probe every relative asset AND every CSS
   `url('/...')` background.
2. **Fences**: python-generate from the template read (root selector; unconditional
   headings asserted `\S`, conditional ones NOT; `{{range}}` grids get ≥1-item checks
   only when the item selector is unambiguous on the proof page; components sharing a
   root class resolve by `data-component` attribute). Mobile = status/overflow/console
   (+ has_visible_area both profiles).
3. **Trial**: `try_fence.go` — all-evaluated-all-passed with reconciled arithmetic.
4. **Prove**: `prove_fence_mutants_file.go <fence> <mutants.json> <url>` — mutants are
   per-subject JSON; every `from` verified to occur EXACTLY ONCE in the served page
   (scripted). Optional declared accommodations in the mutants file: `serve_local`
   (same-origin JS fetches the redirect harness would break) and `strip` (third-party
   beacons that cannot pass CORS from localhost). Both uniform across all runs.
5. **Persist**: batch SQL, supersede-then-insert, DO/RAISE length assert inside the
   transaction (single `%` in RAISE formats!), then read back and diff the fence.
6. **Dispatch**: `DISPATCH_s6_component_run.sh <site> <domain> <fn> <page_id> <bad_page_id>`
   — success lands `neg_control_confirmed_red`; read `acceptance_verdict` and skip
   REASONS.

**LANDMINE (new, synced):** the offline harnesses run HEAD's evaluator — fence vocabulary
newer than the deployed browser-runner passes offline and fails/skips live. Pod-grep step
actions too. If a run fails on vocabulary, the raised `improve_tool` item is a false
positive — cancel with reason in `result` (precedent: `6c06b0ad`, lvs `reload`).

## 4. Remaining work, itemised

- **~17 static sections, 1–2 placements each** (about-commercial-block, archetype-grid,
  featured-content, gripper-spec-sheet, hero-use-cases, patent-check,
  people-feature-block, platform-comparison, social-proof, stat-band, intent-probe,
  portfolio-showcase, pricing, game-master-explanation, archetype-combinations,
  funding-fit, image-hover-card-grid, case-studies-list, hero-case-studies… re-run the
  census, the population moves). Same line, batches of ~8.
- **~16 interactive sections (js_content > 0)**: news-listing, latest-news,
  case-studies-grid, contact-block, ai-readiness-quiz, game-list, report-request-form,
  audience-check-form, blog-listing, tool-ai-agent-roi-estimator,
  tool-ai-vendor-trust-checklist, tool-archetype-taster-quiz,
  tool-gripper-cycle-time-estimator, adoption-tracker-listing, model-directory-listing,
  protocol-tracker-listing. Each needs its JS read (teaser/loan-vs-savings are the
  worked examples for gesture/golden checks). Budget ~30–45 min each.
- **~10 ready tools** (page serves, no PLAN): re-run `CHECK_naming_contract.sh` +
  census; the calibration list had tool-fuel-budget-forecaster, tool-credit-health-check,
  tool-application-tracker, tool-grip-force-friction-calculator, tool-matchmatrix,
  tool-meme-generator, tool-prompt-architect, tool-bg-remover, tool-setup-builder,
  tool-llm-cost-calculator (multi-site — fence must stay site-invariant or golden per
  the shared component's identical JS; check first).
- **Lane-owned, DO NOT fence without coordination**: gauntlet-interface (the arena!),
  gauntlet-cta, lobby-grid, gauntlet-round-record (also a subject_type mismatch —
  its PLAN sits under 'tool'); provocation-card, provocations-archive-list,
  evidence-timeseries, swipeable-insight-carousel (provocation lane, active 08-05).
- **Listings, not fences**: 35 active sections with ZERO active placements (census
  query in NOTES); ~24 tools with no resolvable page (BROKEN-A guard); ported-page's
  58 drift rows on lmc/loancash (served pages carry no component markup).
- **Blocked/deferred**: fce S6 (gaswholesalers logo); tool-gas-unit-converter +
  tool-ab-test-calculator (serve broken — queue-parked items need an owner call);
  tool-equity-release (active row, 404 URL).

## 5. Standing defect list for the owner (found by the line, all verified at the artefact)

1. gaswholesalers.com: every page 404s `/assets/images/logo.png` (assets row exists).
2. dartsonline/gamesdesign/idea/oufe/relojistas/vonc `/assets/images/hero.jpg` 404s
   (bugs_closed/128 measured 07-31) — **still serving 08-05**, at least dartsonline
   confirmed; detection was flag-only, no repair ever dispatched.
3. `article-body` ships no `pre/code` overflow CSS — code-bearing articles scroll
   horizontally on mobile (one live page confirmed; recorded in its PLAN).
4. The two broken tools above; idea.uk serves raw `{{.placeholders}}`.
