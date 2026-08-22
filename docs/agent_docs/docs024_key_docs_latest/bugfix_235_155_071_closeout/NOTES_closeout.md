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

## 2026-08-22 — the day's full trail (Phases 0-3 executed)

**155 (now CLOSED → bugs_closed/):**
- Baseline pageflow proof PASS: corr `6fd5fc99`, hero `de147987…`/logo `47322664…` distinct,
  commits `2bc1888e4`/`f03151d06` at 10:20Z TODAY (verified at the repo, not the report).
- swo arm of `fire_209_proof.sh` had NEVER run: its mapping read `input_data.reviewed_brief`,
  never supplied → strict-mapping hard fail before SWO spawned. Fixed (load_site step mirroring
  the pageflow arm), re-fired (corr `2586036c`): the deploy question is moot but the full build
  failed at `install_site_composition` — "site already has style_collection_id" — a composition
  re-install guard on the sacrificial site, NOT the deploy seam. Not chased; out of scope.
- Migration 324: committed (`8403546ad`) + `--record-only`; no longer pending (runner unblocked).
- Migration 553 (input_contract admits asset_id): written, pre-state guarded, applied (UPDATE 1,
  verify DO passed), recorded, committed. Blast radius measured: ONE contract-validated live
  step (image-build-handler call_asset_deployer), maps both keys.
- asset_id-ONLY proof PASS: `icon_dartboard` corr `4150f72b` → `icon-dartboard.jpg`
  sha `de14fb6c…` commit `1f32bbc40`; `icon_steel_tip` corr `72b3fb29` → `icon-steel-tip.jpg`
  sha `47b0672e…` commit `1c9ad71d9`. Distinct; hashes re-derived from origin/master bytes;
  dartboard opened visually vs origin_prompt: match. 155 moved to bugs_closed (`e626bda11`),
  LANDMINES 155 entry closed out, verifier dispatched (STILL_VALID pending).

**Platform commits (all inert until roll):** A `0ce242d9c` 071 warning persistence (+ r2 delta
`120427766` — failure-path warnings, the council's round-1 catch); B `d59ba32b8` section_editor
CTA deletion; C `69cc0ea7a` 155/209 writer retirement. Mutation proof: `ok && false` mutant
failed 3 tests, restored green. `git archive HEAD` build + tests green.

**Council:** 155 retirement corr `c0e02ad3` APPROVED r1 · 203 CTA corr `dc557fc8` APPROVED r1 ·
071 persistence corr `f30a28e1` REVISE r1 (editquality HIGH: the LogActionEntry landmine — its
own body says FIXED+LIVE v1.0.1268, and the test pins positions 5/8/9/11/12; LOW: sketch
fidelity; MISSING: invalid-build warnings — IMPLEMENTED) → r2 resubmitted same corr · 554
disarm corr `bbf5e418` submitted.

**235 (now CLOSED → bugs_closed/):**
- Pre-deletion audit: page_components 0 · site_specs 0 · content_components 0 ·
  sites.content_data 1 = prose in leopardess's array-typed brief (serves 404; not renderable).
- ⚠ **THE ARMED "DRY RUN" — the day's biggest finding.** Live asset-retraction row carried
  `dry_run:false` (operator edit 08-20, never reverted; description still said dry-run
  default). Ten intended dry runs each DELETED (0 refusals — guards ran; end state was the
  owner-authorised one; pre-audit had run). Caught on first result readback. Disarmed by
  migration 554 (applied+recorded+committed `ce3ca376d`); LANDMINE + WRONG_CALLS filed.
  Cheap check that would have caught it BEFORE: read the LIVE step config (one query).
- 13/13 sites retracted, 0 refusals: 10 sites-repo wire-verified (jpg 404 / png 200,
  cache-busted); idea.uk + relojistas wire-verified after vm propagation (~4 min — the vm
  deletion-propagation question ANSWERED); webdesign.uk repo-verified (commit `85ca602`,
  jpg gone, png present; site 302s deliberately).
- fundamentallyai `image_url_404:logo.png` + `:logo.jpg` cancelled with verification in
  result.reason (premises measured stale); hero.jpg sibling deliberately untouched.
- Register DGH-010 corrected (was "INERT until roll; zero callers" — stale on both counts).

**Shared-tree note:** my WRONG_CALLS append rode ANOTHER session's commit (`62291fa66`) as a
same-file passenger between my write and my commit — the documented behaviour, nothing lost.

**Missteps this session (both in WRONG_CALLS):** (1) trusted four descriptions of a safety
default over the one-row live read — the armed dry-run; (2) a grep filter
(`DOMAIN=|PUBLISH_OK|error`) ate a `Permission denied`, reading as "no output yet".

## 2026-08-22 ~11:10Z — council trail CLOSED: all four tasks APPROVED

- 071 warning persistence `f30a28e1`: r1 REVISE → **r2 APPROVED** (the round-1 catch —
  invalid-build warnings — was implemented, not argued; the HIGH objection dissolved on the
  landmine's own FIXED status + the tests' pinned provenance positions).
- 203 section_editor CTA `dc557fc8`: **APPROVED r1**.
- 155/209 writer retirement `c0e02ad3`: **APPROVED r1**.
- 554 asset-retraction disarm `bbf5e418`: **APPROVED r1**.
All four commits carry `Council-Submitted:`; 098 credits them automatically now the
verdicts are approved — no amends, per forward-only.

## 2026-08-22 ~19:00Z — POST-ROLL VERIFICATION (chassis v1.0.1326)

Fresh chassis deployed (~16:30Z, pods `6bb7b67bd4-*`). Verified at the artefact, not the roll:
- **Stamp**: binary embeds build commit `27b932aca` (found by looping candidate shas over
  `/proc/1/exe` — the provenance log line had rotated out of retention). Ancestry:
  `git merge-base --is-ancestor` → **all four commits IN** (0ce242d9c, 120427766 r2,
  d59ba32b8, 69cc0ea7a).
- **Marker pairs, BOTH replicas, nonsense control clean**:
  `"Failed to store URI"` ABSENT (writer retirement live) · `"Failed to store URL"` PRESENT
  (positive control) · `CONTENT_VALIDATION_WARNING_DETAIL` + `"Valid build carried"` PRESENT
  (071 recorder live).
- **Organic acceptance rows**: zero rows under all three recorder codes in the 4h since the
  roll — a quiet build window, so behavioural acceptance needs the induced/organic run:
  `SELECT count(*) FROM agent_error_log WHERE error_code='CONTENT_VALIDATION_WARNING_DETAIL';`
  (first warning-carrying valid build writes row 1).
- **Regression proof re-fired** (pageflow pair at cookly — exercises StoreAssetAction, the
  changed code, end-to-end): result recorded below when drained.

## 2026-08-22 ~19:35Z — post-roll acceptance CLOSED OUT

- **Regression proof PASS on v1.0.1326**: the pageflow pair at cookly ran a FULL build this
  time (planner + content writer + reviewer, 8 orchestrations, all COMPLETED) and deployed
  fresh distinct artefacts — hero `626aac20…` (174,898 B) vs logo `1df425e8…` (242,987 B),
  new bytes vs this morning's, so StoreAssetAction ran end-to-end on the retired-writer
  binary and each purpose deployed its own asset. Commit C behaviourally proven live.
- **071 recorder: deployed, live row ARMED not proven** — the build's flow never reaches
  `validate_content` (zero ValidatePageContentAction lines in retention), so the zero
  recorder rows mean not-exercised. Armed check + disconfirming shape recorded in 071.
- The two post-roll `apply_edit` FAILEDs re-confirmed NOT ours (required headline/trust_note
  refusals + a slot-floor refusal; the deleted seeds only ever set cta_text/cta_url).
