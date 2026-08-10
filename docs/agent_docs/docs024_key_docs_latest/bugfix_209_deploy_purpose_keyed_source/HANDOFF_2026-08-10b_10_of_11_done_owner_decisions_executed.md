# HANDOFF — 2026-08-10 (evening). 10 of 11 logos DONE. Owner decisions executed. One site remains, mechanism fully diagnosed. COLD-START HERE.

Supersedes `HANDOFF_2026-08-10_235_repaired_8_of_11_and_a_fleet_incident.md`.
Evidence + missteps: `NOTES_209…` (08-10 sections). Owner read-out:
`../bugfix_240_kafka_metadata_storm/SUMMARY_2026-08-10_…`. Shared accounts:
`bugs_open/235`, `bugs_open/240`, `bugs_open/231` — contribute INTO them.

## Owner decisions (taken in-session 2026-08-10 ~16:40Z) and their execution

| decision | state |
|---|---|
| robot-hands: REGENERATE, overriding the `user-b6-approval` lock | **DONE + verified**: serves `logo.png` PNG 400×218, row `purpose='logo'`. Old approved artwork survives at `images/demo_client/20260509/d321c4f2-….png` (recorded in 235) |
| webdesign.uk: clear the malformed `claimed` row | **DONE**: row `8793da9a` reset to `triaged`; site dispatches again. Its logo repair then ran: row `purpose='logo'` 16:59Z. Site is a deliberate 302 — HTTP cannot verify, row + item are the evidence |
| relojistas.com: attempt re-deploy of existing source | **BLOCKED by a NEW 231 finding** (below). Row corrected to `purpose='logo'`; artwork untouched; served artefact still the JPEG |
| 240: C2 + C3 + C4 (NOT C1) | C2 **verified and REFUTED in naive form** (below); C3 **committed** (`55e992e8b`, inert until owner rolls); C4 **installed** (crontab, below) |

## The one remaining site: relojistas.com — read this before touching it

Two deploys ran (17:01, 17:03), both committed **"Deploy hero image"** despite
`purpose='logo'` in the item spec AND (round 2) on the corrected asset row.
Cause, proven from source: **`DeployImageAssetInputSpec` has
`Defaults:{"purpose":"hero"}`** (`deploy_image_asset_action.go:36-38`), and the
asset-deployer step binds `"purpose": "input_data.purpose"` — which resolves
NOTHING on the `undeployed_asset` dispatch shape (spec fields land at
`input_data.spec.*`), so the Default wins over everything. Live specimen
recorded in `bugs_open/231` (it is that census's quarry, caught in the wild).

Net damage: none — both runs re-encoded the same approved artwork to the same
`logo.jpg` path. The row is now correct (`purpose='logo'`, lock intact,
`owner via relojistas-5 session`).

**To finish it** (one of, in preference order):
1. Fix the shape mismatch properly: asset-deployer `deploy_asset` step gains
   `"purpose_field": "input_data.spec.purpose"` (the spec's Deprecated bridge
   maps it to `purpose`) — a SHARED-definition edit, so: migration + council
   round, and it fixes every future `undeployed_asset` deploy, not just this one.
2. Or route via image-build-handler's path (maps purpose top-level correctly —
   its runs commit "Deploy logo image") — but that path regenerates.
After deploy: relojistas' homepage references `logo.jpg`, so it needs a
site-level `needs_rerender` (vm-sites re-render loop is proven — idea.uk).
Also note `fundamentallyai.com`'s portfolio hot-links this site's `logo.jpg` —
after relojistas flips to PNG, fundamentallyai needs a re-render too, and the
old `logo.jpg` must NOT be deleted until that estate-wide audit (in `235`).

## bugs_open/240 — state and the standing warning

- Incident ended by the sweep; **root cause confirmed by dose–response**
  (24,131 topics → 121Mi dying · 354 → 15Mi stable).
- **C2 naive form REFUTED from kafka-go v0.4.47 source, both arms** (full write-up
  in 240): metadata requests are served from the transport CACHE; a topic
  outside a static `MetadataTopics` list gets `UnknownTopicOrPartition` with no
  broker round trip (our writers never set `AllowAutoTopicCreation`); with
  auto-create, `refreshMetadata` waits on state fed by `discover()`'s ONCE-built
  static request and burns the produce deadline. **A static list on the SHARED
  transport breaks every `job.*` producer.** Safe subset: scoped transport for
  closed-topic-set services only (the scheduler qualifies); the fleet-wide lever
  at this library version is the TOPIC COUNT.
- Corollary `[INFERRED]` for 029/040: blank list + 6s TTL + no auto-create is a
  clean mechanism for the spawn→call handshake race. Do NOT raise MetadataTTL
  as an optimisation — it widens that race.
- **C3 committed** (`55e992e8b`): GOMEMLIMIT=192MiB + 256Mi limit on
  kafka-scheduler. Inert until the owner rolls (releases are whole-fleet).
- **C4 installed**: crontab entry, twice daily at :17, runs
  `scripts/kafka-orphan-topic-sweep.sh --apply`, logs to
  `~/kafka-sweep-240.log`. KNOWN FAILURE MODE: kubeconfig token expiry (~3 days)
  makes it refuse loudly into the log — check the log, and remove the entry once
  the real fix lands.
- Topic clock at 16:33Z: 1,236 `job.*` (~129/h steady, ~278/h under load).

## What remains, in order

1. **relojistas** — the shared-definition fix (option 1 above) + rerender + the
   fundamentallyai follow-on. This also closes the last of 235.
2. **231's census** — now has three arms: shadowed static · this new
   unresolvable-dotted-path-with-Default face · and the original 61-spec sweep.
3. **240 C2 safe subset** — scheduler-scoped transport (code + tests + council),
   plus the owner question of whether to revisit C1 given the C2 refutation
   shows the topic count is the only fleet-wide lever.
4. **Estate-wide logo.jpg reference audit** before any deletion (in 235).
5. 209 Phase 3 (retire dead writers) and 236 remain open, unowned by this thread.

## Cold-start checks

1. `go test ./platform/orchestration/actions/ -run 'TestExtractActionInputs_|TestDeployImageAsset_|TestLegacyLogoStep_|TestPurposeFieldBridge_|TestStrategy0DottedPaths_|TestMigration348Shape_'` — 7 pass expected.
2. Migration 360 in force (query in 235); survived the v1.0.1279 and 1280 rolls.
3. Topic count (in-pod, file-first, three agreeing reads — piping TRUNCATES):
   `kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- bash -c 'bin/kafka-topics.sh --bootstrap-server localhost:9092 --list > /tmp/t.txt 2>/dev/null; grep -c "^job\." /tmp/t.txt; rm -f /tmp/t.txt'`
4. Scheduler health = MEMORY not restart count (a roll resets the counter);
   ~15Mi good, ~120Mi = the topic count is back.
5. `tail ~/kafka-sweep-240.log` — the C4 cron's voice.
