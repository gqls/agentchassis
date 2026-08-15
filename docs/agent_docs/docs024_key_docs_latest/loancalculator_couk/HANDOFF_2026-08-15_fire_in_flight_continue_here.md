# HANDOFF — loancalculator.co.uk · REBUILD FIRE IN FLIGHT (2026-08-15 ~11:00Z)

> Supersedes `HANDOFF_2026-08-14_post_canary_continue_here.md`. The owner answered
> all three questions; PHASE 1 of a TWO-PHASE fire is dispatched and mid-pipeline.
> Full mechanics + evidence: NOTES 2026-08-14 (late night, new session) onward.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
fire      PHASE 1 corr 2d950ecc-4919-441b-a4fb-e6aa47663ad9, DB fire time
          2026-08-15 07:54:41Z  (stamp baselines from the DB clock ONLY — the
          local machine ran 10h behind until NTP snapped it mid-session)
pages     29 active = 27 built+serving (verified post-roll) + about + guides-index
          (restored not-built, zero components). tool-standard-calc stays ARCHIVED
          (08-03 artifact, unrelated).
locks     12 calculator locks ONLY (decompose_20260802_proven_calculators).
          The 8 non-calculator locks are RELEASED (chrome 3 + css carriers 4 +
          homepage prose-0); pre-release state: NOTES 08-11 §step 6.
chassis   v1.0.1301 (rolled 10:14Z 08-15). RECOMPOSE tell live since 1295.
state     PLAN 34b1b056 LANDED 10:59:33Z · checkpoint RUN (Q2: ZERO invention —
          the pin held; identity md5 unchanged e6dd8fb8…; un-defer trio moved;
          calc locks 12/12) · REBUILD RUNNING: 15 needs_page (built non-tool
          pages) + about + guides-index + hero_about + 1 needs_rerender open,
          11 owned_page_review at the human gate ALREADY (phase 1, not 2).
```

> **CORRECTED ~11:15Z, same session — read this before the phase framing below:**
> the reconciler emitted the FULL rebuild set in PHASE 1 (the mission prose alone
> sufficed; mechanism note in NOTES — plan sections for built pages are again
> EMPTY, yet items emitted; sync_pages-writes-empty-sections suspected, unchased).
> **Phase 2 is ON HOLD and probably unnecessary** — its regeneration purpose is
> served; the placement test moved to BUILD time and is running ungated on
> `index` (locked tool-loan-repayment + identity arm are the floor). The
> `phase2_recompose_26.sh` script stays valid if plan-level composition records
> or the RECOMPOSE tell are ever wanted. D1/D3 below are REVISED accordingly.

## The owner's three answers (2026-08-14/15, applied)

1. **Regeneration:** explicit `recompose_pages`, all 26 pre-fire built pages.
2. **Keep-pages pin:** trusted, WITH the immediate post-planner new-active-rows
   check; anything invented gets archived before it can build.
3. **Invented pages:** BOTH wanted. Rows restored to active; their build items
   un-defer at the checkpoint (NOT before — see timing rule below).

## Why the fire is TWO-PHASE (measured, not preferred)

082 carries no spec: the pipeline-minted `needs_site_plan` has NO spec field and
the mint→dispatch window is seconds — `spec.recompose_pages` cannot ride phase 1.
- **Phase 1 (in flight):** 082 rebuild. Mission installed (and corrected — see
  wrong-calls below), research/vertical/strategy COMPLETE (07:59/08:27/08:31Z),
  briefing → plan → the two restored pages build → chrome/design regenerate.
- **Phase 2 (`phase2_recompose_26.sh`, in this dir, committed):** direct
  build-site-planner dispatch with `input_data.spec.recompose_pages` = the 26.
  Fire ONLY after phase 1 settles and verifies. The redesign intent PROSE already
  stands in the mission spec (both-places rule of the recompose no-op landmine).

## Two gates phase 1 hit — BOTH by design, both re-driven (NOTES 08-15)

1. **B2 refresh-safety gate** (strategist `gate_next_item`, migration 341): a
   deployed site's strategy refresh completes WITHOUT chaining briefing→plan.
   Remedy (precedent: webdesign.uk 08-09): manually file the withheld
   `needs_briefing` item. **082 fresh-mode alone will NEVER re-plan a deployed
   site — this manual step is REQUIRED, every time.**
2. **The polling dispatcher cannot see `detected`**: `find_dispatchable_site`
   takes ONLY `triaged`/`approved`; chain-created items are dispatched by their
   creating workflow (that's why they claim in ~30s), and the fleet triage sweep
   is backlogged to 08-13. **Manual redrive items must be born `triaged`.**
   Item `00bd5eff` was moved there 10:51Z.

## THE CHECKPOINT — run `./checkpoint_postplan.sh` the moment a NEW current plan lands

(Script in this dir; a watcher loop polling for the plan is trivial to recreate:
`SELECT id FROM site_plans WHERE site_id='0162cde4…' AND is_current AND
created_at > '2026-08-15 07:54:00Z'`.) The script, in order: Q2 invention check
(new page rows after fire — expect NONE; archive any, guarded, zero-components) ·
new plan shape + roles · compositions for about/guides-index (expect >0 sections
each — never-shipped pages ARE composable, `realisedPageHasShipped`) · tool
placements in plan · RECOMPOSE tell (expect none in phase 1) · page-identity md5
of the 27 · **UN-DEFER the trio** `222ecf94` (needs_page:about), `a52e59d8`
(needs_page:guides-index), `ad289c0e` (hero_about imagery) — they are the ONLY
build path for the two pages ('deferred' blocks both the dedup index and the
reconciler's re-mint) and must move ONLY after the mission plan is current, else
the ~90s dispatcher builds them from the canary's bare plan · calc locks 12/12 ·
new items listing.

## Then: phase 1 verification, phase 2, final verification

1. Two pages build → serving 29/29 (`./check_site_serving.sh`; fetch at
   `pages.url`, NEVER name-derived paths). Rerender/design cascade may follow —
   let it settle.
2. Fire `./phase2_recompose_26.sh`. EXPECTED output, not failure: ~15 needs_page
   (auto-rebuild) + **11 `owned_page_review` for the tool-role pages (human gate,
   NO handler — TP-004: the generic builder clobbers tool pages)** + 1
   needs_rerender. The 11 go to the OWNER.
3. Judge phase 2: RECOMPOSE tell rows (`proposed_verbatim` = the no-op landmine
   fired; `absent_from_plan` = dropped/renamed — investigate) + plan-sections
   diff vs the pre-phase-2 baseline the script prints.
4. Final battery (handoff-08-14 step 5, unchanged): purity of the 12 locked rows
   vs 08-11 backups (`loancalc_bak_20260811_*`, snapshot `0d1b55f0` — NB revert
   re-applies CURRENT lock state, not the snapshot's), pre/post URL diff EMPTY on
   the 27, toolgolden 11/11, calculators IN PLACE, then RE-LOCK judgement:
   whether the 8 released locks (or successors) are re-armed is an OWNER call —
   the rebuilt chrome/css may not need carriers at all.

## OWNER DECISIONS OPEN (REVISED ~11:15Z; nothing blocks the running rebuild)

- **D1 (RE-REVISED ~14:30Z — PHASE 2 FIRED, scoped 26→12):** the hold reversed
  when the homepage escalation decoded — landing/tool builders REQUIRE plan
  compositions (`spec_sections count 0, source none` → `mark_no_ready_sections`);
  guide builders self-compose. Phase 2 supplies exactly those compositions and
  mutates no live page. Fired corr `2f74a975-1a87-40a8-af88-a9bd2ecc1510`
  14:23:32Z, scope = index + the 11 tool pages (the 14 rebuilt guides+legal
  excluded to avoid churn). Judge queries in the script's output; the open
  review items are the vehicles that realise the compositions.
- **D2 — the 11 tool-page reviews (ALREADY MINTED, `owned_page_review:*`):**
  per page or in bulk — realise the redesign deliberately / leave the built page
  as-is (locks keep the calculators regardless) / authorise a bulk-apply gated
  by the acceptance checks (toolgolden 11/11 + serving + locked-row purity).
- **D3 (EXPANDED) — the owner-parked 08-12 items now GATE the visual refresh:**
  `needs_design` + `needs_composition` (site-wide keys!) + 2
  `needs_brand_head_assets` + 3 imagery items (logo `003b98cc`, guide-loan-faqs
  hero `15d6323f`, homepage hero `34b3ace1`) all sit deferred and DEDUP-BLOCK
  fresh mints. Consequence: the planner's emit_design produced NOTHING, so the
  released chrome/css will NOT regenerate until these free up. Un-deferring
  reverses the owner's own 08-12 parking → owner call. (**CORRECTED ~14:30Z: the born-triaged
  rule applies to EVERY hand-moved item** — the previous sentence here said
  'detected' was fine for some types and that was WRONG, proven by the trio
  sitting invisible for 3 hours. Un-park = set `'triaged'`, always.)
- **D4 (housekeeping, later):** 13 lock_blocked_change + 6 content_rewrite at
  needs_human_review + the remaining parked items + blocked bugs_open/189
  rerender — mostly superseded by the rebuild; sweep once it settles.

## Standing cautions (carried)

- Query runs BY CORRELATION, never now()-interval; planner rows purge in ~2 days.
- No orchestration dispatch within ~300s of a chassis pod (re)start.
- A roll kills in-flight council/orchestration runs — check pods before firing.
- `lock_blocked_change` = "the lock was exercised", never "the copy differed".
- The stale `reconcile_rerender:8d7c…` (canary plan) stays deferred FOREVER.
- Wrong calls this fire (both in WRONG_CALLS 08-15): the homepage carries
  `tool-loan-repayment` (NOT the credit roadmap — mission + spec rows corrected
  pre-reader); baselines stamp from the DB clock, never local.
