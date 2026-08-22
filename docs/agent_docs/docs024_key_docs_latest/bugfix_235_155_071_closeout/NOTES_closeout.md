# NOTES — bugs 235 / 155 / 071 close-out (append-only, newest at the bottom)

## 2026-08-22 — session start: verification pass, ownership, owner decisions

- Ownership: `who-owns.py` on all three + git log on the file paths — nobody else active.
  235's residual lane closed its council trail 08-21; 155/071 quiet since the 08-21 header
  corrections. ⚠ 071 is an ambiguous number (closed `agent_job_cleanup` 071 unrelated).
- `[MEASURED]` 0 `page_components` rows reference `logo.jpg` fleet-wide (rendered_html OR
  content_data) — 235's 08-21 proof still holds.
- `[MEASURED]` live `asset-deployer.deploy_asset` `input_fields` =
  `["s3_uri","purpose","domain","asset_key","asset_id"]` — migration 324 IS applied,
  but the file is untracked in git and absent from `schema_migrations`.
- `[MEASURED]` fundamentallyai.com HAS an active `logo` asset (updated 08-10) and serves
  `/assets/images/logo.png` 200/157165 B — the `image_url_404:logo.png` item's "no active
  asset" premise is false; `image_url_404:logo.jpg`'s premise (pages reference it) is also
  false today. Both `detected`, both stale. (A third item, hero.jpg, left for re-verification.)
- Explore agent (155): `findStorageURI` GONE (91dda3243, 209 Phase 2, live v1.0.1276);
  zero Go readers of `{purpose}_uri`; writers survive at `generate_image_actions.go:1009-1012,
  :1027, :1032` and `v3_site_actions.go:3458-3460` (stale comment :3445-3452); the
  sha256-differ behavioural proof has NEVER run; LANDMINES 155 entry already corrected.
- Explore agent (071): `valid :=` unchanged at :400; persistence hole real and reachable
  (warning-only-no-repair builds write nothing durable); bugs_closed/079 closed via
  `repairSectionsBeforePersist`; 092 closed; **NEW finding**: `section_editor_actions.go:783-785`
  fabricates `cta_url:"/contact.html"` pre-merge — 203's class, unrecorded there.
- Design agent: full design + censuses. Live-config census: ZERO readers of flat
  `{purpose}_uri` in active agent_definitions / workflow_templates / active content_components.
  `store_generated_image` appears in zero agent_definitions rows (doubly dead cache writer).
  Discovery: asset-deployer `input_contract` still `required:["domain","s3_uri"]`
  (enforced by `call_agent.go:1005-1013`) — blocks the asset_id-only closure dispatch;
  needs a small migration. Chassis v1.0.1323 carries `retract_asset_files` (marker probe
  on `/proc/1/exe`, pod -4qlp7; second replica still to check before arming).
- Owner decisions recorded in PLAN. NormalizePagePath + fragment RFC deferred by owner.
- **090 not used this session — substitution stated per the 2026-07-31 ruling:** every
  residual claim acted on here is either already loop-verified (155's CONFIRMED run
  0dd9aee4), council-tested (235: three REVISE rounds each finding something real), or
  first-hand measured this session with the query/file:line inline (the two new claims:
  persistence hole — read at :400/:409/:449 with the warning producers enumerated; CTA
  fabrication — read at :783-785 with the merge order and both template-layer guards
  traced). The deferred NormalizePagePath cross-cutting claim is exactly the shape that
  DOES need a 090 before filing — noted in the plan for whoever takes it.

## 2026-08-22 — Phase 0 ops

- Baseline 209 proof fired (pageflow): CORR `6fd5fc99-434f-4312-a37f-59fce57bb13c`,
  ORCH `e55996f0-e979-406e-bb5f-14372a06bb81`, PUBLISH_OK seen. Queue latency ~30 min is
  normal — find it by payload, not by printed id.
