# HANDOFF 2026-09-03 — bugs_open/114 imagery wiring: continue here

**Supersedes `HANDOFF_2026-08-22_continue_here.md`** (read only its §Traps if you read it
at all). Cold-start for whoever picks this lane up. Read this, then the 2026-09-02/03
entries of `NOTES_imagery_wiring.md` (evidence + missteps), then
`PLAN_2026-08-22_imagery_wiring.md`'s three 2026-09-02 addenda (the revised plan, the
follow-ups, the RFC_063 claim). `README_where_we_are.md` is the owner's plain-prose log —
append, never rewrite. Sibling account you MUST read before touching delivery:
`bugs_open/412` §§10–11b (the handover, the build record, the REVISE triage, the
plan-less domain limit).

## State in one paragraph

`bugs_open/114` ("imagery is planned, generated, deployed, and nothing points a page at
it") is **one observation away from closable**. All four closing-bar items are satisfied
as of this morning: the IMG-072 gate and IMG-073 event-derive are live and proven at
fleet scale (193/193 natural convergence); no post-roll poisoning anywhere; and the
detection check (IMG-077) is **live on chassis `v1.0.1356`, enabled by migration 708, and
fired correctly on its first sweep** (idea.uk, 09:25Z today: one `no_image_slot` rollup,
count 6, orchestration COMPLETED). Migrations 562 and 709 are applied and
council-APPROVED; the detector is APPROVED (round 3). The delivery mechanism for the
`unwired` state (IMG-078, bugs_open/412 candidate 1) is **built, aboard the fleet, and
deliberately INERT** — opt-in default OFF, migration 710 HELD — parked behind a
diagnosis-first protocol.

## What closes 114, in order (the only remaining gate is OBSERVATION)

1. **Watch the sweep cover the fleet** (~a day; `site_discovery_rotation` picks sites
   >4h stale every 300s). Confirm rollups appear where the census predicts and
   orchestrations complete:
   ```sql
   SELECT s.domain, wi.item_key, wi.spec->>'count', wi.spec->>'measured_at', wi.created_at::date
   FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE wi.item_type='unrendered_page_imagery' ORDER BY wi.created_at DESC;
   ```
   Predicted `[MEASURED 2026-09-02]`: `unwired` on webdesign.co.uk (largest, 66
   candidates), gamesdesign, loanandmortgagecalculator; `fragment_slot` on the 16
   tool-page rows (mcalc has 10); `no_image_slot` widely (tool 231, blog-post 7 pages).
   ⚠ **A fleet-wide silence after full rotation is an unexercised detector — a bug, not
   a clean fleet** (708's runbook step 3). Also check the design orchestrations keep
   completing.
2. **Move the file**: `git mv bugs_open/114_… bugs_closed/114_…` with a closing appendix
   pointing the residual states at their owners (unwired → 412; fragment_slot → 357;
   no_image_slot → RFC_063 execution + the composition question). Name BOTH paths on the
   commit (LANDMINES: a git mv + pathspec ships a copy otherwise) and verify with
   `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 114` → exactly one line.
3. Update `MEMORY_closed.md` / the memory index per its own rules if a 114 line exists.

## What was LIVE-VERIFIED this morning (2026-09-03, all at the artefact)

| thing | proof |
|---|---|
| chassis `v1.0.1356` carries the lane's commits | capability probe: `unrendered_page_imagery` on 79 pods = positive control, negative control 0; wiring literal in `/proc/1/exe` present=1, absent-control=0 |
| migration 709 applied | `removed icon=16 content_hero=6 illustration=4 sprite_sheet=1`, 18 rows in `bak_site_dead_purpose_urls_20260902` — byte-identical to the pre-approval dry run |
| migration 708 applied + discharged | preflight n=1, UPDATE 1, checks array 24→25, post-apply clean; `_HOLD` off both files; `liveConfiguredChecks` updated in the same commit (`8c4a5789e`) |
| first natural firing | idea.uk 09:25:32Z, `unrendered_page_imagery:no_image_slot` count 6 dated 2026-09-03, orchestration COMPLETED |

## The lane's OTHER work — none of it blocks closing 114

- **IMG-078 / 412 candidate 1 (the `unwired` delivery): PARKED, deliberately.** The
  council REVISE (corr `bd78490d`) was right: the content_data floor loses to the
  resolver's fresh-merge on exactly the GAP-4 cohort. **Round 2's protocol is written in
  `bugs_open/412` §11a**: run ONE disposition-logged rerender canary on a quiet unwired
  page (webdesign.co.uk is busy and its unwired pages are not blog-post typed — pick
  from the census query in `RUNBOOK_imagery_wiring.md`), read the resolver's logged
  eligibility (is `pageName` empty at the resolver? that is the live hypothesis two
  diagnosis attempts could not settle before the logging existed), THEN ship whichever
  of (resolver fix | site_plan_imagery page-row upsert) the evidence licenses. §11b adds
  the domain limit: 6 plan-less sites (203 deployed pages) are unreachable by the
  plan-row design until RFC_063's conversion lands; route 2 (Lane B) is their
  plan-independent fallback today. **Do NOT apply migration 710 before this diagnosis.**
- **RFC_063 execution — this lane CLAIMS the imagery seeding step only** (PLAN, last
  addendum; the RFC's execution line records it). Owner ruled option B (converge the 6
  plan-less sites; hand-insert permitted as a closed backfill). Sequenced strictly
  after (a) the unwaived one-site reconciler-skip proof and (b) the composition half
  creating each site's current `site_plans` row — both UNASSIGNED, not ours. Then: seed
  `site_plan_imagery` page-scope hero rows **from the assets table only** (4 sites have
  rows to seed: ai-agent-orchestration 17, finetuning 14, loancash 9, lampenkap 6;
  gaswholesalers and cookly seed ZERO) — seeding from page enumeration trips
  `check_unfulfilled_imagery_plan` into generation spend. Ship as a council-reviewed
  migration.
- **Tracked follow-up (architecture seat): narrow or retire `check_undeployed_assets`
  half 1** once IMG-077 rollups give the evidence. Its born-`unresolved` backlog
  (1,651 → 1,662 within hours — count rows with `created_at=updated_at`, it moves)
  has a LANDMINES entry; the backlog's DISPOSITION is an owner decision.
- **Owner decisions still unasked:** widening `check_content_image_missing`'s surface to
  `page_type='content'` (fleet generation spend — population query in the old handoff);
  the parked-rows disposition (7 recorded in NOTES with reasons, left parked).

## Council ledger (all verdicts READ)

| what | corr | verdict |
|---|---|---|
| 562 hero_url repair | `4145fcdc` | **APPROVED** (resubmitted after a dropped dispatch — 0 orch rows for 11 days is a drop, not latency) |
| 709 dead purpose-url keys | `151a51db` | **APPROVED** standalone |
| IMG-077 detector + 708 | `3b568104` | **APPROVED round 3** (r1: a landmine cited unread — real; r2: sketches hid committed code; ten advisories adjudicated in NOTES 09-02 close) |
| IMG-078 wiring + 710 | `bd78490d` | **REVISE — deliberately not resubmitted** pending the GAP-4 canary (412 §11a) |

## Traps this lane hit or dodged since the last handoff — read before repeating

1. **Grep LANDMINES for the SYMBOL you are about to trust.** Citing `DeployedWebPath`
   without reading its landmine cost a council round; the landmine turned out fixed and
   brand-head-only, but that was luck, not diligence.
2. **A jsonb path probe returning a uniform result is evidence about your PATH.**
   `spec->>'page_id'` (the key is `entity_id`) produced a false "0 of 193 converged".
   Dump ONE row's spec first. (WRONG_CALLS 09-02.)
3. **Grep the code token before filing a 090.** The mcalc tool-page render mystery was
   one grep from `adopt_fragment_section.go` / bugs_open/357 — already diagnosed.
4. **The shared tree can break your `go test` with someone's dirty WIP, and your commits
   can break theirs.** Both happened same-day (their `CTALabelCandidateRow` WIP; my
   hand-spelled tombstone predicates racing `def8126e3`'s single-sourcing —
   `TestNoHandSpelledTombstonePredicate` went red fleet-wide until `d1cf3aac3`). Use
   `git worktree` for clean-tree tests; NEVER stash. Use `datahelpers.NotRemoved()` /
   `PageWantedLivePredicateFor()` — never spell the predicates.
5. **`pages.build_status='deployed'` is not "has deployed"** — for a fires-at-render
   cohort use `deployed_at IS NOT NULL`; the broadened LANDMINES entry (finetuning lane)
   is the fleet-visible home of the rule; the 17-page delta concentrates on
   gaswholesalers.com, which makes it the self-driving plan-less canary.
6. **A landmine indexed by INTENT is invisible to a reader with a different intent and
   the identical predicate** — grep the symbol, not your intent.
7. **The council-report body in `doc_notes` truncates; the full report is
   `diagnosis_artifacts` WHERE kind='council_report' AND correlation_id=<corr>` (column
   `body`, JSON).** And a 097 sketch that is all comments is REFUSED — sketches carry
   real code lines.

## Files of record

`NOTES_imagery_wiring.md` (2026-09-02 ×4 + 2026-09-03) · `PLAN_2026-08-22…md` (three
09-02 addenda) · `bugs_open/114…md` (RESUMPTION section + UPDATE blocks) ·
`bugs_open/412…md` §§10–11b · `architecture_review/RFC_063…md` (ruled; our consumer
input + claim) · register `imagery.md` IMG-077/IMG-078 (+ corrected 072/073) ·
`LANDMINES.md` (born-unresolved backlog entry) · `WRONG_CALLS.md` 09-02 ·
CONTRIBs in: mortgagecalculator_couk_adoption, editorial_design_uplift,
inline_guide_imagery, bugfix_357_component_identity.
