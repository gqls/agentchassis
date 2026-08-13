# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-13)

**Supersedes `HANDOFF_2026-08-11_continue_here.md`** (read its §11 second — it holds
the full diagnosis this file continues from; its §0 owner rulings all stand).
Written mid-task at the owner's instruction: token load was high, so this is the
continuation state, not a completion report. **The work below is NOT done.**

## 0. The owner's standing asks, verbatim and current

1. **2026-08-12:** hero image gone · top nav says just "Home" · cards have no
   imagery — *"The point is not to do that manually but to figure out why the
   framework didn't do it."* (Diagnosed — 08-11 handoff §11. Not yet repaired.)
2. **2026-08-12:** *"create the handlers that are missing and leave the assignment
   until the rung 2 is fixed"* (DONE — IMG-071, migrations 397+398, both inert.)
3. **2026-08-13:** *"carry on. And also the site currently has no hero image or
   logo etc."* — **this is the live task.** "etc." is real: measured 2026-08-13,
   **every** brand asset 404s: `/assets/images/hero.jpg`, `logo.png`, `logo.jpg`,
   `favicon.png`, `/favicon.ico`, `og-card.png` — all 404. And the `sites` row has
   `logo_url` NULL, `logo_text` NULL, `brand_assets = {}`.
4. Bad copy, recorded not yet fixed: *"and you don't need to sign up for any of
   it."* (voice spec §3 rule: no sentence nobody would say out loud).

## 1. Where the task stands — one paragraph

The three visible symptoms trace to three framework defects (08-11 §11): the hero
was generated and deployed to a placeholder path (`bugs_open/248`, rung 2 —
`asset-deployer.deploy_asset config.asset_key = "input_data.asset_key"`, **verified
still live 2026-08-12**); the header chrome can thin but never thicken (090 filed
twice, BOTH runs died on loop infrastructure, claim unverified); the card-imagery
items sit in `deferred`, a status dispatch cannot select. Two router agents now
exist for the two never-handled item types (IMG-071) but route nothing. **The next
concrete step was, and remains: cut rung 2, then drive hero + logo live through the
framework, then assign the routers.**

## 2. NEXT ACTIONS, in order — the plan I was executing when this handoff was cut

1. **Migration 399 — cut rung 2** (`bugs_open/248` fix candidate 1, config form).
   `asset-deployer`'s `deploy_asset` step carries `config.asset_key =
   "input_data.asset_key"`. When rung 1 (path resolution) misses, rung 2 takes that
   same string as a LITERAL filename → `input-data.asset-key.jpg`. The fix that
   makes the bad state unrepresentable: **DELETE the `asset_key` key from that step's
   config entirely.** Then: a caller that supplies asset_key (image-build-handler
   maps `asset_key?: input_data.spec.asset_key`) is found by ExtractActionInputs'
   recursive search (248 §(b) records this works — asset_key has NO default, so the
   search runs); a caller that doesn't → `assetKey=""` → `DeployedAssetPath("" ,
   purpose)` → the correct purpose path (`hero.jpg`). BEFORE writing it, read the
   actual rung code in `deploy_image_asset_action.go` (~line 230-260, "three-rung
   ladder" comment) — I had NOT yet read the precise rung mechanics when this
   handoff was cut; do not trust my paraphrase of rung 1 vs rung 2 resolution.
   Snapshot first (`snapshot_agent`), guard with RAISE (never bare SELECTs), apply
   by hand + `--record-only` (**NEVER `--apply`** — 12 pending files from other
   threads, 5 flagged replay hazards).
2. **Maybe migration 400 — dispatcher purpose mapping** (248 fix candidate 2).
   `build-dispatch-loop.call_handler`'s input_mapping has no `purpose` key, so on
   the `undeployed_asset` → asset-deployer route the `Defaults: {"purpose":"hero"}`
   default wins and EVERYTHING deploys as a hero (that's why all 150 placeholder
   files are `.jpg`). Needed if you take the redeploy-existing-asset path for the
   hero (below). NOT needed for generation routes (image-build-handler supplies its
   own purpose from `asset_stored.purpose`).
3. **Hero live.** Two paths, pick one:
   - **(a) Redeploy an existing generated hero, no regeneration cost.** The site
     has TWO `active` hero assets (`0e11c818`, `2e2bea17`, both 2026-08-11
     19:10-19:11Z; note `d6ead260` — the one from §11.1's trace — is now
     `superseded`, and `477838e3` is `rejected`: the owner has been curating).
     File an `undeployed_asset` item mirroring `check_undeployed_assets.go:113-119`'s
     spec shape (`check`, `asset_id`, `purpose`, `asset_type`, `url`), status
     `triaged`, and let build-dispatch-loop run it. ⚠ Requires steps 1 AND 2 first.
     ⚠ NEVER dispatch asset-deployer by direct kcat to the standing chassis — no
     storage client there (`bugs_open/245`); only build-dispatch-loop and
     image-generator-adapter pods have the bucket. 248 proved this the hard way.
   - **(b) Let detection re-file.** `check_placeholder_image_in_use` re-files
     `needs_hero_image` → image-build-handler regenerates (a paid generation) and
     deploys — needs only step 1, not step 2. ⚠ BUT the check's asset-existence
     gate was fixed 2026-08-12 from purpose-scoped to asset_key-scoped
     (`hasActiveAssetForAssetKey` — see IMG-069's update note); with two ACTIVE
     hero rows present, whether it still fires depends on those rows' asset_key
     values. **v1.0.1294 rolled 2026-08-13 09:48Z — check whether that fix is
     aboard before relying on this path** (provenance line already scrolled;
     use the OCI-label fallback: `docker pull` the image, read
     `org.opencontainers.image.revision`, then `git merge-base --is-ancestor`).
   Then wire-verify: `curl -o /dev/null -w '%{http_code}'
   https://mortgagecalculator.co.uk/assets/images/hero.jpg` → 200, AND the wrong
   path still serving is fine (it is unreferenced litter, cleanup is 248's owner).
   **Run both curls in one breath — the pair is the evidence** (248 contribution).
4. **Logo.** Harder, because there is an owner AESTHETIC ruling in play:
   - Facts: the site has **NO logo asset row, ever** (5 asset rows, all hero).
     `logo_url`/`logo_text`/`brand_assets` all empty on the site row. The header
     serves a text-only `logo-text` span. A `needs_imagery:site:-:logo` item sits
     `deferred` (priority 70, from build-site-planner 08-02).
   - The tension: on 2026-08-11 the owner reviewed two generated logo candidates
     for THIS site and ruled *"I prefer the old [existing] logo … we can approve
     this route for future attempts"* (IMG-069 update note). So the generation
     ROUTE is approved, but the owner rejected its output for this site once
     already. Now the owner says the site "has no logo".
   - **Do this first: find what "the old existing logo" IS.** Check the original
     adopted site's crawl assets / the site repo `gqls/sites` for a pre-framework
     logo file; check `assets` on OTHER rows (superseded/rejected 08-11 rows were
     heroes, not logos — the two rejected logo candidates may live under a
     different site_id or in `assets` with `origin_type='generated'` elsewhere).
     If an original logo artefact exists: deploy THAT (via the fixed asset-deployer
     route, purpose=logo). If none exists: promote the deferred
     `needs_imagery:site:-:logo` to `triaged` (the route is owner-approved) and let
     image-build-handler generate — but expect the owner may reject the output;
     that is their call, not a defect.
5. **Favicon + og-card.** `needs_brand_head_assets` items completed TWICE
   (08-05, 08-11) yet `favicon.png`/`og-card.png` 404. Brand-head artefacts use
   their own fixed path registry (`storage.BrandHeadAssetPaths`), NOT the
   asset_key derivation — so this may be a DIFFERENT defect from 248, or the same
   one via the deploy step. Diagnose before assuming; grep the completed items'
   `result` for the committed `file_path` first (the 08-11 §11.1 trick — the
   work item's own result usually names both paths).
6. **Assign the routers** (IMG-071) once rung 2 is cut AND wire-proven: the
   assignment SQL is in `397_image_flag_only_routers.sql`'s header. Owner's
   condition ("leave assignment until rung 2 is fixed") is then met. Suggest
   assigning THIS SITE's rows first (3 `image_url_404` + 17
   `image_source_unsatisfiable`), watching one of each route end-to-end, then the
   fleet's remaining ~73. Watch the first `conditional_branch` run — IMG-071's
   verify-later names `asset_facts.backing` resolvability through
   `output_format: "object"` flattening as the most likely failure.
7. **Cards imagery** (owner ask 1) — after hero/logo. The 13 `deferred`
   `needs_imagery` rows include the 4 homepage card icons (priority 98). Promoting
   them post-fix generates icons under `icon-*` names — but **check the render path
   consumes them before spending**: the homepage `tool-list`/`guide-list` items
   carry frozen `"image": ""` with `data_sources` NULL (08-11 §4), so a generated
   icon may be `bugs_open/114`'s class (paid, deployed, referenced by nothing).
   Read how `buildRenderContextFromDB` fills `items[].image` first.
8. **Un-parked leftovers, lower priority:** the nav (chrome thinning — blocked on
   the 090 loop being healthy again, claim still UNVERIFIED, do NOT assert §11.2's
   mechanism as established); the 30 stale `<title>`s (08-11 §8.1); the bad copy
   line (§0.4 above — a section_edit on the hero slot, path per 08-11 §7).

## 3. Fresh chassis build — what it changes and what it does not

- **v1.0.1294, both agent-chassis replicas, started 2026-08-13 09:48/09:49Z**
  (pod list; provenance line already out of `--tail=300` range — busy service).
- Rung 2 is **DB config, unaffected by any build** — assume still live until
  migration 399 lands (it was verified live 2026-08-12).
- **No orchestration dispatch within ~300s of a chassis pod (re)start** — the
  spawn is silently dropped (CLAUDE.md). Both pods restarted 09:48Z; safe by now,
  but re-check if another roll happens mid-work.
- What MAY have ridden the roll and matters here: the
  `check_placeholder_image_in_use` asset_key-granularity fix (IMG-069 note,
  council-submitted `7e839679…`) — decides whether path 3(b) works. Verify
  per-service, per stamp, never per-fleet (`bugs_open/249`).

## 4. What this arc of sessions changed (all committed, chronological)

| commit | what |
|---|---|
| `f1f0d30b2` | The three-symptom diagnosis (08-11 handoff §11, NOTES, 3 LANDMINES) |
| `e8b6c9aa9` | 248 contribution: census under-counts (150/16 by `updated_at`; this site invisible to it — all its hero rows have `filename=''`); rung 2 still live; my WRONG_CALLS entry (wrote it up before grepping bugs_open) |
| `2499e9450` | 090 filed on chrome-nav one-directional claim |
| `e4946f9ed` | That 090 FAILED (timeout, terminal at attempt 1/1); UTC-vs-BST trap recorded (join on `spec->>'dispatch_correlation_id'`, never time) |
| `5ce785e50` | IMG-071: the two routers (migrations 397+398), inert, register entry |
| `0a7d08d58` | 090 retry failed too — loop-wide (5 of 10 runs failed 08-12); stopped at two |

Migrations applied + recorded (`record-only`): `397`, `398`. NOT YET WRITTEN: `399`
(rung 2 cut), possibly `400` (dispatcher purpose mapping).

## 5. Landmines for whoever continues (beyond LANDMINES.md's own)

- **`--apply` on the migration runner sweeps 12 other threads' pending files**, 5
  flagged replay hazards. Apply by hand, then `--record-only` with a note.
- **`snapshot_agent` RAISES on a type that does not exist** — a migration that
  CREATES types cannot open with the README's prescribed line (397 shows the
  conditional pattern).
- **`audit-config-keys.sh` silence is not clearance** — it decoded 184 of 189
  active agents and names only agents WITH findings. Verify strict-key contracts
  directly against `create_work_item`'s spec (12 recognised keys; `spec` is
  RETIRED — use `spec_paths`/`spec_literal`; `site_id` is Required and every live
  caller sets it explicitly).
- **The diagnosis loop was failing ~half its runs on 2026-08-12**
  (`execute_llm_prompt` → "context canceled"). Check
  `SELECT status, count(*) FROM site_work_items WHERE item_type='needs_diagnosis'
  AND created_at::date=CURRENT_DATE GROUP BY 1;` before spending a 090 run.
- **The `sites` row for this site is `pending` build_status with `locked_at` NULL**
  — unlocked, per owner ruling. Do not lock it.
- Correlations that matter: 090 chrome-nav intake
  `853347f4-41f3-4d57-a016-8f0af0ba2763`, runs `2aab9013…` (failed) and
  `044c0a54…` (failed). Both terminal. Re-file only after the loop is healthy.

## 6. Evidence quick-reference (so the next session doesn't re-derive)

- Site id: `62b5978e-4271-4589-8e00-4baebfc0447c`.
- The stranded hero: live at `/assets/images/input-data.asset-key.jpg` (200,
  68,984 B) — unreferenced; the page's inline CSS references `hero.jpg` (404).
- Asset rows (2026-08-13): `477838e3` rejected · `d6ead260` superseded ·
  `9e94250d` superseded · `0e11c818` ACTIVE · `2e2bea17` ACTIVE — all purpose
  `hero`, all `filename=''` (which is why 248's census can't see this site).
- Work-item dead ends on this site: 13 `needs_imagery` `deferred` (10 days,
  attempt_count 0) · 17 `image_source_unsatisfiable` `needs_human_review` ·
  3 `image_url_404` `blocked`. Zero rows dispatch can see.
- The nav data is CORRECT (16 rows, 5 primary); the stored header chrome froze
  at 18:06Z 08-11 with 1 link. Footer serves 16. Do not "fix" the nav tables.
