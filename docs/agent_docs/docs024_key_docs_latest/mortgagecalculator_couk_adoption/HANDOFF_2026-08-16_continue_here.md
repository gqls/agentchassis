# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-16)

**Supersedes `HANDOFF_2026-08-13_continue_here.md`** — and corrects it: read that file's §0 for
the owner's standing asks (still valid), then treat its §2 plan as HISTORY. Everything it asked
for is either done, superseded, or re-scoped below, with the reasons in NOTES `## 2026-08-14`
and `## 2026-08-16`.

## 0. Owner rulings and decisions in force (verbatim where it matters)

1. 08-12: *"The point is not to do that manually but to figure out why the framework didn't do
   it."* — ANSWERED (§2). 08-13: *"the site currently has no hero image or logo etc."* — DONE (§1).
2. 08-12: *"I prefer the old logo … carry on with the original logo but we can approve this
   route for future attempts."* — the ORIGINAL gold roundel is the logo. Executed.
3. 08-14 (via question, twice): the homepage hero is **hero_v2** (`9e94250d`, the text-free
   navy icons image). Executed. v3 (`3b0cac59`, wordmark baked in) is superseded, kept.
4. 08-14: design discovery returns as a **one-shot for this site only** — done and disabled
   again. **Fleet-wide `site-discovery-rotation-design` re-enable is an OPEN OWNER COST CALL**;
   the 08-10 pause rationale (deploys wasted on the placeholder path) is gone.
5. 08-16: *"let the hero items run"* — they did (10/10 tool-page heroes generated 08-15).
6. Standing: never hand-build; every change through the framework; site row stays unlocked.

## 1. Where the site IS (measured 2026-08-16 ~10:00Z, all at the URL)

| artefact | state | how it got there |
|---|---|---|
| `/assets/images/logo.png` | 200, the original roundel | `scripts/amend-asset.sh` ingest → asset `e766370e` (uploaded/operator-supplied, active) → `undeployed_asset` deploy |
| `/assets/images/favicon.png` | 200, 64×64 roundel | `needs_brand_head_assets` `{"mode":"brand_head"}` → `derive_brand_head_assets` |
| `/assets/images/og-card.png` | 200, 1200×630 | same item |
| `/assets/images/hero.jpg` | 200, 68,984 B = hero_v2 | `undeployed_asset:9e94250d…` |
| header | **renders `<img src="/assets/images/logo.png">`**, nav Home + About | the `stale_chrome` rerender (08-15 18:40Z) consumed the asset — the "chrome can only thin" worry from 08-11 §11.2 did not block this |
| hero copy | *"No sign-up, no upsell, and no personal data collected…"* | `site-review_content_rewrite_index` 08-15 18:50Z (improvement loop, unasked) overwrote my 08-14 `section_edit`; result is fine, layout intact |
| 10 tool-page heroes | 10 active assets, 10 files serve 200 — **and NOTHING references them** (entity link NULL, tool pages re-rendered to the `hero.jpg` fallback, listing cards `image:""`) | `content_image_missing` GENERATE arm; `bugs_open/114`-class, contributed there 08-16. Left in place as the fixture for whoever wires the link |
| placeholder litter | `input-data.asset-key.jpg` still 200, unreferenced | 248's backlog-drain owns it |

Site id `62b5978e-4271-4589-8e00-4baebfc0447c`. `sites.github_repo` empty → B2 route.

## 2. Why the framework "didn't do it" — the answer, in three layers

1. **Detection was OFF.** `site-discovery-rotation-design` `enabled=false` since 08-10 (cost
   pause). Nothing files hero/logo/brand-head gaps while it is off. (One-shot precedent:
   `oneshot-design-discovery-mcalc-20260814`, `enabled=f` now.)
2. **The hero had spent its two strikes** (`placeholder_image_in_use:hero`, two `complete`
   rows 08-11; strikes = complete+failed within 7 days, `load_work_item_actions.go:1276-1284`).
3. **The favicon/og-card items were falsely `complete`** — never ran (bugs_closed/213 §D /
   274: the child's reply never validated, the parent completed the item with a foreign
   payload). AND `derive_brand_head_assets` is logo-gated — no active logo row had ever
   existed for this site. Fixed upstream (274 CLOSED 08-15, live v1.0.1303, verified with
   demand: 0 cannot-deliver vs 859 child completions).

## 3. THE LIVE THREAD — `bugs_open/287`, filed this morning

While verifying 274 is live I found the NEXT defect on the same seam: since the 08-15 10:14Z
roll, ~75% of dispatch-loop completions record the **spawn record** (`{role,topics,agent_id}`)
as the item's `result` instead of the handler's reply — the work is done, the record is wrong.
Zero instances before that roll. `090` run **`fb7ae3bc-e9bf-4a96-b540-d593b91bc79c`** —
**verdict: see §3a below / read it if blank:**
`SELECT body FROM doc_notes WHERE categories ? 'diagnosis' AND body LIKE '%fb7ae3bc%' ORDER BY created_at DESC LIMIT 1;`
⚠ The trigger warned the loop reads ORIGIN and local HEAD is 853 commits ahead — it may not
see WFA-014 (`3ba384c63`). If its citations predate that commit, its mechanism read is of the
wrong binary; say so in 287 §6 rather than accepting or dismissing it.

Not this lane's bug to FIX (coordinator seam, RFC_012 territory) — but this lane found it, filed
it, and owns getting the verdict recorded. Then hand it to whoever owns the coordinator.

### 3a. Verdict
_(fill in when read)_

## 4. NEXT ACTIONS, in order

1. **Read the 287 verdict** (§3), record it in 287 §6, note if it read the wrong tree.
2. ~~Eyeball one tool-page hero~~ DONE 08-16: `/tools/repayment/index.html` does NOT reference
   `content-hero-tool-repayment.jpg`; its hero fields resolved to `/assets/images/hero.jpg` on
   the post-generation rerender; entity link NULL on all 10. **The 10 tool heroes are paid and
   unconsumed** — `bugs_open/114` contribution 08-16 has the measurements. Do NOT promote or
   re-file `content_image_missing` items on this site until 114's link is wired; if the owner
   wants them visible, that is a page-hero MAPPING change (read how the image-landed rerender
   fills `hero_url` — it took the site fallback over the page-scoped asset), not more generation.
3. **Router fleet assignment (IMG-071) — deliberately NOT done, and here is the rule I'd
   apply:** on this site every one of the 20 pre-fix findings was stale by the time the router
   ran (this session's own deploys had fixed them 3 minutes earlier), so routing produced only
   noise I then cancelled. **Per site: fresh discovery pass FIRST, then route** — never route
   a pre-fix finding post-fix. Both routers are wire-proven (escalate branch and mappable
   branch both resolved; the `asset_facts.backing` flatten works). Two known cosmetic defects:
   `summary_from: "input_data.summary"` is a template-string misuse (escalation summaries read
   literally), and 397 has NO assignment SQL for the unsatisfiable arm — write it with a
   `site_id` predicate.
4. **Card icons** — still parked, still 114-class: no per-array-element source resolution
   exists (`plan_sections_action.go:2076`), so a generated icon can never reach the frozen
   `items[].image`. Needs a component-field change (top-level `source: site_assets.icon`) or a
   `card`-purpose derived asset; not a work-item promotion.
5. **The 30 stale `<title>`s** (08-11 §8.1) — mechanical, unchanged.
6. **`images/mortgagecalculatormono.xcf`** (GIMP master, 175 KB, publicly served, flagged 3×
   since 07-31, never actioned) — remove from repo+bucket; keep a local copy first (only
   artefact that may hold a wide wordmark).
7. Fleet design-rotation re-enable — owner's call; do not act.

## 5. Landmines this lane paid for (beyond LANDMINES.md)

- **`idx_assets_site_asset_key_unique`** — one ACTIVE row per (site, asset_key). Supersede
  before you activate, or the swap aborts mid-transaction.
- **A discovery finding routed after the fix files noise** (§4.3). Route only after a fresh pass.
- **`site_work_items.result` is not the handler's report** — first 213/274 (foreign planner
  payload), now 287 (spawn record). Verify EVERY deploy at the URL; never at the item.
- **The improvement loop rewrites hero copy unasked** (`site-review_content_rewrite_index`).
  A `section_edit` you land can be gone next day; check the served page, not your item.
- The asset-amend bucket URL in `assets.url` is PRIVATE (401) — the row's `file_size` +
  `dimensions` + the action's sha256 check are the byte proof, not a curl.
- `check_mode` (brand_head) in LIVE config already tests `input_data.spec.mode`; the 07-11
  seed file does not. Read `agent_definitions`, not seeds (LANDMINE class already on file).
- Read the 08-14 NOTES entry for the `attempt_count` wrong call — a claim does NOT bump it.

## 6. Files of record

- NOTES `## 2026-08-14`, `## 2026-08-16` (this arc); README `2026-08-14 (evening)` (owner prose).
- `bugs_open/287_…` (new, mine); `bugs_closed/213` §D contribution + correction (mine);
  `WRONG_CALLS.md` 08-14 attempt_count entry (mine).
- Commits this arc: `a2a691213 ad86a498d c24ed39fb d360457c2 fbad2d434 cc4670d75 5988e4111 ae5d12048`.
