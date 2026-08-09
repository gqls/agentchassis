# HANDOFF — 2026-08-09 (afternoon). PHASE 2 IS COMMITTED (not rolled). 231 IS LIVE DAMAGE. COLD-START HERE.

Supersedes `HANDOFF_2026-08-09b_phase1_live_phase2_next.md`. That file's "next
work" list is done except item 3 (the fleet census) and item 4 (Phase 3).
`NOTES_209…` = evidence + missteps (read the 08-09 afternoon section first);
`RUNBOOK_209…` §10 = how to run a proof build; `README_where_we_are.md` = the
owner's plain-prose log. Contribute INTO `bugs_open/209` / `231` — shared accounts.

## State

| | |
|---|---|
| Phase 1 (migration 348) | **DONE, LIVE, and now PROVEN IN PRODUCTION** — re-verified by CONTENT after the v1.0.1274 roll |
| Phase 2 (delete `findStorageURI`) | **CODE COMMITTED `91dda3243`, NOT ROLLED.** Go-only ⇒ inert until the next chassis image. `Council-Submitted: cc4909c6-de80-4e40-910f-9186e9da8762` — **verdict NOT yet read; reading it is owed** |
| behavioural proof | **pageflow-builder arm DONE** (cookly.uk, correlation `0562a667-…`). site-work-orchestrator arm **attempted twice**, see below |
| 231 | **UPGRADED from shadow to LIVE DAMAGE: 11 sites.** 090 filed for the producer, run correlation `fd7ef7a9-93fb-4e20-9956-f8913bd4ab89` — **verdict owed** |
| commits this session | `e93fe691d` (proof + 231/210 contributions), `91dda3243` (Phase 2 + register) |

## The finding that matters most — read this before anything else

**Migration 348 did not fix the fleet's logo problem, and 231 is not a shadow.**

Every logo committed to `gqls/sites`: **4 correct** (400×400 PNG, commit subject
"Deploy **logo** image") and **11 wrong** (JPEG at 1408×768 / 900×900, subject
"Deploy **hero** image") — gamesdesign, idea.uk, vonc, dartsonline, robot-hands,
vetcomparison, fundamentallyai, oufe, webdesign.co.uk, relojistas, lendzy,
webdesign.uk. JPEG has no alpha channel, so those logos ship with opaque
backgrounds at up to 3× the intended size. Three independent signals agree that
`purpose == "hero"` at those deploys (commit subject built from `purpose` at
`deploy_image_asset_action.go:579`; extension comes from `purpose` and filename
from `asset_key`; the resize keys off `purpose` and none is 400×400).

**But the producer is `[UNVERIFIED]` and is NOT the shape 231 modelled.** 231
predicted the logo landing on the *hero's* path (assuming no `asset_key`); the
artefacts are named `logo.jpg`, so an `asset_key` WAS supplied. 348 repaired the
two workflows nobody dispatches. `assets` points at **`asset-deployer`** — four
rows are literally named `input-data.asset-key.jpg`, an unresolved
`input_data.asset_key` config path leaked in as the asset key. **Do not assert
the producer; the 090 above is answering exactly that.**

## Next work, in order

1. **Read the two verdicts you already own** (both were in flight when this was
   written; neither is optional):
   - Council on Phase 2: `SELECT created_at, metadata->>'decision' FROM
     diagnosis_artifacts WHERE correlation_id='cc4909c6-de80-4e40-910f-9186e9da8762'
     AND kind='council_report' ORDER BY created_at;` — the code is already on the
     shared branch, so a REVISE/REJECTED must be acted on, not filed.
   - 090 on the logo producer: `SELECT current_step, status FROM
     orchestration_states WHERE correlation_id='fd7ef7a9-93fb-4e20-9956-f8913bd4ab89';`
     then its `diagnosis_artifacts`. **Record the verdict into `bugs_open/231`
     whether it confirms or refutes** — a REFUTED is the cheap place to be wrong.
2. **The 11 wrong logos are unfixed and nobody owns them.** Once the 090 names the
   producer: fix the producer first (a repair that re-runs through the same broken
   caller just re-makes the artefact), then re-deploy the logos. Note
   `bugs_open/210`'s `needs_logo` file has a contribution from this session
   correcting its §6, which had read `logo.jpg` as normal.
3. **Finish the SWO proof arm** (item 1 of the previous handoff, still open).
   Three dispatch failures, NONE of them deploy failures and NONE the handshake
   race (I claimed that for attempt 1 without reading the error column — logged
   in `WRONG_CALLS.md` 2026-08-09; the `error` column names the cause on every
   FAILED row, read it first):
   - attempts 1+2 (`4263b54d-…`, `4faf3162-…`): my message mapped
     `reviewed_brief` from `input_data.reviewed_brief`, a path my own
     `input_data` never carried.
   - attempt 3 (`04beb727-…`): mapping removed — failed on the agent CONTRACT:
     `missing required fields: [input_data reviewed_brief]`.
   - attempt 4 (`aab47560-…`): mirrors the WORKING pageflow shape — outer
     `load_site` step, `reviewed_brief: site_record.content_data`,
     `input_data: input_data` (the fixed message is in `fire_209_proof.sh`,
     RUNBOOK §10). Fired 13:49; **check its outcome and verify the deploy steps
     by object key + stamped asset row, as §10 says.**
   The mechanism is already proven on the pageflow arm and the SWO config rows are
   byte-identical in shape, so this arm is corroboration, not the load-bearing
   evidence. Do not let it block Phase 3.
   Related but separate: the pageflow proof run itself later FAILED at
   `apply_site_design` — `CHILD_ORCHESTRATION_FAILED`, request timed out after 3
   retries, NO design-agent child row ever appeared. The asset proof is unaffected
   (it completed and is committed); cookly.uk simply has no site-wide design pass.
   Left uncancelled; if you pick it up, that one may genuinely be the spawn/call
   family — but READ ITS ERROR first, per the wrong call above.
4. **231's fleet-class census** (unchanged, still undone): enumerate the **61**
   `ActionInputSpec`s carrying non-empty `Defaults` (`grep 'Defaults: map' under
   platform/orchestration/actions/`, list is in NOTES), then per defaulted field
   query live `agent_definitions` step configs for a static non-dotted value ≠ the
   default. Candidate 3 (CheckConfig flags a shadowed static) is the cheap win and
   would have caught the logo producer at authoring time.
5. **Phase 3 (optional):** retire the `{purpose}_uri` writers. After `91dda3243`
   the writer at `v3_site_actions.go:2852` has **no reader at all** — that is new
   as of this session and makes Phase 3 cheaper than when it was written up.

## Traps paid for — do not re-derive

- **The per-agent pods keep ~11 SECONDS of logs.** `--since=20m` returned 11
  seconds. Attach `logs -f` BEFORE dispatching (RUNBOOK §10). Only `agent-chassis`
  is a Deployment; `agent-pageflow-builder-*` / `agent-site-work-orchestrator-*`
  are spawned per run, so poll for the pod rather than naming it.
- **"hero.* and logo.* differ in bytes" is NOT a disconfirmable test** and it is
  written into BOTH bug files as the verification bar. The deploy re-encodes per
  purpose (hero→jpg, logo→png), so they differ even when the wrong source is
  fetched. Assert the **downloaded object key** and the **stamped asset row**.
- **`s3_uri` is in `DeployImageAssetInputSpec.Optional`, so Strategy 0 resolves a
  config dotted path REGARDLESS of `input_fields`.** 348's "deliberately excluded
  s3_uri" means excluded from `input_fields` only — it suppresses the aggressive
  search, not the explicit path. Reading it the other way inverts the Phase 2 risk
  assessment.
- A **fleet roll landed mid-session** (v1.0.1274, 12:23 UTC) between a config check
  and a dispatch. Re-verify config by CONTENT after any roll; `updated_at` is
  re-stamped fleet-wide without content changing.
- `sites` has **no `site_id` column** (it is `id`); `content_data` is an **array**
  on some sites, so `jsonb_object_keys` errors there.
- The package's one failing test (`TestValidDocSubjectTypes_Lockstep…`) is the
  064-shape recurrence owned by `idea_uk_vm_site` — pre-existing, not this lane's.
- `apply_site_design` on the pageflow proof run stalled in `AWAITING_RESPONSES`
  with no child orchestration — same spawn→call family. **Left uncancelled
  deliberately** (memory: never cancel a failing handshake row pre-diagnosis).

## Cold-start checks

1. `git log --oneline 91dda3243..HEAD -- platform/orchestration/actions/deploy_image_asset_action.go bugs_open/209_* bugs_open/231_*` — empty = ground unmoved.
2. `go test ./platform/orchestration/actions/ -run 'TestExtractActionInputs_|TestDeployImageAsset_|TestLegacyLogoStep_|TestPurposeFieldBridge_|TestStrategy0DottedPaths_|TestMigration348Shape_'` — 7 expected (was 8; two findStorageURI tests became one action-level guard).
3. **Has Phase 2 rolled?** `kubectl exec -n ai-persona-system <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "Found URI at purpose_uri"'` → **0 means it shipped**; pair it with a positive control (`grep -c "Resolved source object from asset row"` ≥ 1) so a zero cannot come from a broken grep.
4. Ownership sweep (RUNBOOK §9) — another thread may have taken the 231 producer work.
