# TODO — Open items as of 2026-05-15 end of session

Single reference for everything that came out of the Phase 2G+2H verification work. Organised by area, prioritised within each area. Cross-references back to where each item was first discussed.

---

## State of play

**Phase 2G + 2H: verified end-to-end at scale on 2026-05-15.** Seven hero items processed through the full discovery → audit → triage → dispatch → handler → asset → git deploy chain on robot-hands.com without manual intervention beyond initial unsticking. Each ~25-30s through the handler. All seven assets committed to git, visible at `robot-hands.com/assets/images/hero-*.jpg`.

**One outstanding item:** `icon_cross_technology` still failing (state machine corrupted: `status=triaged` with stale `claimed_at`). Two attempts remaining before failed-terminal.

**Pages stale:** the seven new heroes aren't yet rendered into the live pages — last page rerender was 2-3 days ago. Rerender trigger queued.

**Next session entry point:** run `trigger_rerender_robot_hands.sh`, then `trigger_audit_finetuning.sh` for parallel verification on a different site, then return to the icon.

---

## Imagery pipeline (Phase 2G/2H follow-ups)

Direct outputs of the verification work. Sized small-to-medium, none are blockers.

### 1. State machine corruption on failed items
**Severity:** medium. Hygiene bug; doesn't block work, but makes diagnosis confusing.
**Symptom:** the icon work item ended up `status=triaged` with `claimed_at=2026-05-15 14:50:47` and `claimed_by=build-dispatch-loop` populated. That's an invalid state combination — `triaged` should mean "not yet claimed" with NULL claim metadata.
**Cause:** some failure-recovery path resets status without clearing the claim metadata. Probably `mark_work_item_failed` or its caller in image-build-handler's error_step.
**Fix:** trace what UPDATE-ed the row when the image-generator failed at 512×512. Ensure claim metadata is cleared on any non-claimed status transition.

### 2. `updated_at` trigger not firing on site_work_items UPDATEs
**Severity:** medium. Makes the table's audit trail unreliable.
**Symptom:** icon row shows `updated_at=2026-05-13 18:54:34` despite multiple recent UPDATEs (`claimed_at` set on 2026-05-15, `result` shows entries from 2026-05-15).
**Cause:** either the trigger doesn't exist, or it's column-filtered to only fire on `status` changes (and our UPDATEs were of other columns).
**Fix:** check for `update_site_work_items_updated_at` trigger; ensure it fires on any UPDATE.

### 3. Icon dimensions — patch verification or second override
**Severity:** medium. Blocking the icon from completing.
**Status:** user says the icon_dimensions_patch is deployed (kindDefaults["icon"] now 1024×1024). But the failure at 14:50 logged 512×512 to Stability. Two possibilities, both worth checking:
  - The chassis pod restart at ~14:30 ran on a pre-patch image (since pod age says 27min but doesn't confirm image tag)
  - There's a second injection path for width/height that overrides kindDefaults
**Fix:** verify the chassis container image tag (`kubectl describe pod ... | grep -i image:`). If correct, reset the icon and trigger; watch the next orchestration's `collected_data->'request_body'` to see what dimensions were actually sent.

### 4. `constraints.aspect` vs `style_hints.aspect_ratio` schema/code mismatch
**Severity:** low. Doesn't break anything today, but creates author confusion.
**Symptom:** the planner LLM is emitting aspect ratio under `spec.constraints.aspect`. The code only reads `spec.style_hints.aspect_ratio`. So constraint hints from the planner get silently ignored for dimension purposes (they ARE used for negative prompt).
**Fix:** either teach the planner to use `style_hints.aspect_ratio`, or extend `generate_image_actions.go` line 252-pattern to also accept `constraints.aspect`. Pick one canonical location.

### 5. `parseAspectRatio` SDXL v1.0 whitelist validation
**Severity:** low (no concrete failure today since the icon's aspect didn't come through here).
**Symptom:** the function snaps to multiples of 64 (a generic SDXL constraint), not the strict SDXL v1.0 whitelist (1024x1024, 1152x896, 1216x832, etc).
**Fix:** add a post-snap validation step that rounds to the nearest whitelisted dimension pair for SDXL v1.0.

---

## Dispatch architecture (FOCUS_dispatch_diagnostic.md Q3/Q4)

Larger improvements to the dispatch chain, surfaced by this session's deep dive.

### 6. Watchdog: auto-reset stuck claimed items
**Severity:** high. Caused two incidents in this session (both `needs_design_review` zombies blocked the entire site).
**Mechanism:** a single `claimed` item with no progress excludes its entire site from dispatch indefinitely via the `NOT EXISTS` clause in `find_dispatchable_site`.
**Fix:** scheduled job (or extension to build-pipeline-trigger) that resets `status` to `triaged` and clears claim metadata when `claimed_at > now() - interval '15 minutes'` AND no active orchestration exists for the item.

### 7. Pipeline field — fix discovery emission + loosen dispatcher
**Severity:** medium. Latent bug that bites whenever a discovery check writes a non-`build` pipeline.
**Decision recorded in Q4 (2026-05-15, with user):** leave the field as soft, currently-unused routing label.
**Two concrete changes:**
  - Fix `unfulfilled_imagery_plan` (and any sibling checks) to NOT explicitly write `pipeline='design'`. Let schema default ('build') apply.
  - Loosen `build-dispatch-loop.load_items.item_pipeline` config to accept any value (or drop the filter).

### 8. Outer ORDER BY in find_dispatchable_site
**Severity:** low. Affects fairness, not correctness.
**Symptom:** when multiple sites have eligible items, the `LIMIT 1` selects arbitrarily (Postgres scan order). Some sites might wait longer than others purely by UUID lexical position.
**Fix:** add an outer `ORDER BY oldest-pending-item ASC` or similar fairness metric.

### 9. build-pipeline-trigger writes orchestration_states
**Severity:** low. Makes debugging dispatch decisions harder.
**Symptom:** the trigger runs every 30s but doesn't appear in `orchestration_states` (not in the 34 known `owner_agent_type` values). Its decisions are invisible to post-hoc auditing.
**Fix:** trigger should write a row per invocation with the selected site_id (or none) and the reason.

### 10. Per-item-type circuit breaker
**Severity:** medium. The `needs_design_review` failures starved all imagery work on robot-hands.com for hours today.
**Mechanism:** when an item type fails N times in a row on a site, exclude it from `find_dispatchable_site` consideration for that site for a cooldown period. Lets other item types make progress.
**Fix:** new column or table tracking per-site-per-type failure rate; SQL clause in `find_dispatchable_site` excluding sites where ALL pending items are of a recently-failing type.

### 11. needs_design_review handler reliability
**Severity:** medium. Currently a recurring source of zombies on robot-hands.com.
**Status:** the audit run created `needs_design_review` items that consistently fail with "Claim timed out — handler pod likely died". Pattern suggests systematic crash, not transient.
**Fix:** separate investigation needed. Not blocking imagery work — can be punted to `failed` manually if it gets in the way again.

---

## Reconciler / coupling gaps

Closing feedback loops in the pipeline.

### 12. Asset → page rerender coupling
**Severity:** medium. Visible to users — pages show old heroes/no heroes even after assets are generated.
**Symptom:** seven new hero assets generated 1 hour ago. Pages on the site last rendered 2-3 days ago. No automatic rerender was triggered.
**Fix:** when `store_asset_action` inserts a new row, find pages referencing that asset_key (via `content_data.*_url` or rendered HTML) and insert a `needs_rerender` work item for each. Closes the loop.

### 13. Phase 3 — adoption image mirror (deferred)
**Severity:** low. Documented in `PLAN_imagery_loop_closure.md`. Adopted sites lose their crawled imagery; the mirror would persist them as reference imagery for img2img generations.
**Status:** deferred — not blocking the build pipeline.

### 14. Variant chain site_id pass-through
**Severity:** low. Edge case, surfaced earlier in the session.
**Symptom:** when variant chains spawn child orchestrations, site_id sometimes doesn't propagate through `input_mapping`, causing `render_js_snippets` to fail with "site_id not found at site_record.site_id".
**Fix:** confirm the chain's input_mapping includes `site_id`; trace any chain that loses it.

---

## Content quality (robot-hands.com observations from rendered HTML)

Not imagery work, but noticed during verification. Worth a separate cleanup pass.

### 15. Duplicate learning pages
Four pages with overlapping names: `learning-center.html`, `learning-center-index.html`, `learning-center-hub.html`, `learning-center-article.html`. Probably accumulated from discovery iterations. Footer lists three of them as "Learning Center" — confusing for users.

### 16. Duplicate selection guide pages
`selection-guide.html` and `gripper-selection-guide.html` — both linked from footer as "Selection Guide".

### 17. Empty `tools.html`
Listed in main nav and footer. Body is empty in the rendered HTML.

### 18. Broken footer contact display
Footer has `<a href="mailto:"></a>` and an empty `<p></p>` where contact details should be. Either `sites.email` / `site_specs.identity.email` is empty, or the renderer isn't reading them.

These are operator-level cleanups, not pipeline work. Worth a separate "site hygiene" pass when convenient.

---

## What's verified today (for the record)

- ✅ Discovery emits `needs_imagery` items at `status=detected` ✓ (Phase 2G.4)
- ✅ `design-audit-agent` triage step promotes `detected → triaged` ✓
- ✅ `build-pipeline-trigger` selects site with `find_dispatchable_site` query ✓
- ✅ `build-dispatch-loop` claims up to 5 triaged items per tick per site ✓
- ✅ `image-build-handler` runs handler workflow for `needs_imagery` items ✓
- ✅ `generate_image` produces SDXL-valid 1024×1024 hero images ✓
- ✅ `store_asset` writes `assets` row with `purpose='hero'`, `origin_model='sdxl'` ✓
- ✅ `deploy_image_asset` optimises (1600×900 for hero) and commits to git ✓
- ✅ `mark_work_item_complete` transitions items to `status=complete` ✓
- ❌ Page rerender after new asset deploy — gap; needs reconciler (item 12)
- ❌ Icon path — see items 1-5

---

## Sequence for next session

1. Run `trigger_rerender_robot_hands.sh` — pages pick up the 7 heroes generated today. Visual confirmation on the live site.
2. Run `trigger_audit_finetuning.sh` — parallel verification on a different site. Validates the pipeline isn't site-specific.
3. Sort out the icon (items 1-5 above). Most likely just needs a chassis image tag verification and a retry, but might surface a second override path.
4. After icon completes (or determined to be a real bug), declare 2G+2H done and switch focus.
5. Next architectural focus: item 6 (watchdog for stuck claims) or item 12 (asset→page rerender coupling) — both unblock real operational value.

---

## Cross-references

- `FOCUS_dispatch_diagnostic.md` — dispatch architecture deep-dive, Q1-Q4
- `016_debugging_guide.md` — operational reference for state-machine debugging
- `PLAN_imagery_loop_closure.md` — broader plan; this session closes Phases 2G+2H
- `icon_dimensions_patch.go` — patch for kindDefaults["icon"] (status: deployed per user, awaiting verification)
- `trigger_rerender_robot_hands.sh` — next immediate action
- `trigger_audit_finetuning.sh` — second immediate action
