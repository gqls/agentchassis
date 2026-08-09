# HANDOFF — 2026-08-09 (evening). 209 IS DONE AND LIVE. 235 FIXED AT SOURCE. COLD-START HERE.

Supersedes `HANDOFF_2026-08-09c_…` (whose "Phase 2 committed, NOT rolled" is no
longer true) and `…09b_…`. Read `NOTES_209…` (08-09 afternoon + late-afternoon
sections) for evidence and every misstep; `RUNBOOK_209…` §10 for how to run a
proof build. Contribute INTO the bug files — shared accounts.

## State — the lane's original brief is COMPLETE

| item | state |
|---|---|
| Phase 1 (migration 348) | DONE, live, **and behaviourally proven** on both legacy workflows |
| Phase 2 (delete `findStorageURI`) | **DONE, LIVE, POD-VERIFIED on v1.0.1276** — commit `91dda3243`, council **APPROVED** (`cc4909c6-…`, read: 1 low advisory, nothing to act on) |
| behavioural proof | DONE both arms — pageflow `0562a667-…`, SWO `aab47560-…` |
| `bugs_open/209` | defect fixed and live. Open only for **Phase 3** (below) |
| `bugs_open/231` | producer question ANSWERED → `bugs_open/235`. Open for the **fleet census** + candidates 2/3 |
| `bugs_open/235` | **FIXED AT SOURCE** (migration 360, applied+verified). Open for the **11 published artefacts** |
| `bugs_open/236` | FILED today, unowned — nothing checks whether a site serves |

**How Phase 2 was proven live** (the shape to copy for any DELETION): per-replica
`strings /app/agent-chassis | grep -c` on v1.0.1276 — removed string
`Found URI at purpose_uri` = **0** on both, with two positive controls at 1
(`Resolved source object from asset row`, `no storage URI found for`). A zero
without a positive control cannot distinguish "shipped" from "broken grep".

## What remains, in the order I would do it

### 1. Re-make the 11 wrong logo artefacts (`bugs_open/235` candidate 2) — highest value

Migration 360 stops NEW ones; the published JPEGs are untouched. Eleven live
sites serve a logo processed as a hero (JPEG, no alpha, up to 1408×768 instead of
400×400 PNG): gamesdesign.co.uk, idea.uk, vonc.com, dartsonline.com,
robot-hands.com, vetcomparison.uk, fundamentallyai.com, oufe.com,
webdesign.co.uk, lendzy.co.uk, webdesign.uk.

**Order matters, and two traps are already paid for:**
1. Re-drive each site's logo through the repaired branch — file a `needs_imagery`
   item with `brand_update:true, asset_key:"logo", purpose:"logo"`. Verify FIRST
   on **cookly.uk** (sacrificial, already wired): assert the artefact is
   `logo.png`, **400×400 PNG**, commit subject "Deploy **logo** image", and
   `assets.purpose='logo'` on the stamped row. Those properties are the
   disconfirmable part — a wrong purpose cannot produce them.
2. **A deploy writes the derived path only — nothing deletes the stale
   `logo.jpg`.** And pages still REFERENCE it (robot-hands' HTML points at
   `/assets/images/logo.jpg`). So: re-deploy the correct `logo.png` → re-render
   the pages referencing the old name → only then delete the stale file.

### 2. `bugs_open/231`'s fleet-class census — the last of the original brief

Undone. **61** `ActionInputSpec`s carry non-empty `Defaults` (enumerated 08-09,
list in NOTES; regenerate with `grep 'Defaults: map' platform/orchestration/actions/*.go`).
For each defaulted field, query live `agent_definitions` step configs for a
**static, non-dotted** value ≠ the default — those are silently ignored.
⚠ Remember 235's lesson: this census finds the *shadowed* class only. A static
that resolves fine but is simply WRONG for one arm of a branch (235) is a
different door and this census will not see it.

Also open: 231 candidate 3 (`CheckConfig` flags a shadowed static — cheap, catches
future authors) and candidate 2 (make config-static beat Defaults — needs the
census first, and a council round).

### 3. Phase 3 — retire the `{purpose}_uri` writers (`bugs_open/209`, optional)

Cheaper now than when written up: since v1.0.1276 the writer at
`v3_site_actions.go:2852` has **no reader at all**. A writer with no reader is
what the next author will assume is load-bearing. Also `generate_image_actions.go:994`.

### 4. `bugs_open/236` — nothing checks whether a site SERVES

Filed today after `lendzy.co.uk` was found returning **522 to every visitor** with
every internal signal green. Fix candidate 1 (a scheduled probe over every live
domain) would also have caught `loanzy.uk`'s dangling delegation. The probe
helper exists — `discovery_checks/check_asset_reference_404.go:220-236`.
⚠ Any such checker MUST honour the rollout lane's skip-list or it cries wolf:
`idea.uk`/`relojistas.com` are VM-served and `webdesign.uk` is a deliberate 302.

## Domain / infrastructure work done today (owner-requested, mid-lane)

Recorded in `domains_cloudflare_rollout/NOTES` — that lane owns the follow-ups.

- **cookly.uk is LIVE** (zone `ab126cfa…`, active 15:14Z). Owner ran the Nominet
  repoint. Serves 200/29,393 B. Its logo is the **first correct 400×400 PNG since
  2026-03-02** — the post-348 proof artefact.
  Known shortcoming: **no site-wide design applied** — `apply_site_design` failed
  on both builds (see below).
- **www now works** on cookly.uk + dartsonline.com. The CNAME alone never could:
  the shared worker keys the bucket path on the RAW hostname
  (`objectKey = ${hostname}${path}`), so www 404s. Needs three parts — www CNAME,
  a 301 page rule, **and deletion of the `*.<domain>/*` worker route**, which
  otherwise wins. The other ~13 zones still 404 on www; one line in the shared
  worker (`hostname.replace(/^www\./,'')`) would fix all of them, but that script
  backs every site — flagged for the rollout lane, deliberately not taken.
- **loanzy.uk** was a **dangling delegation** (NS at Cloudflare, zone never
  created → authoritative answer pointing at a dead placeholder → timeout, not
  NXDOMAIN). Owner added the zone; now returns a fast 404 because **no site has
  ever been built for it** (0 rows, no bucket folder; HOLD in the positioning
  register). Wired to the estate template and ready if a site is ever wanted.
- **lendzy.co.uk was DOWN (522)** — zone had NO worker routes. Fixed; 200/41,431 B.
  Census of all 38 active zones found no others (3 apparent hits were VM-served
  or a deliberate redirect — tested, not reported).
- Two LANDMINES added + synced: the parked-domain 200-on-every-path trap, and the
  dangling-delegation timeout trap.
- **Token file is misnamed**: `~/.config/cloudflare/token.expired-2026-08` is the
  LIVE write-scoped token (to 2026-09-30); plain `token` is read-only. Its scopes
  lack the Rulesets API — hence Page Rules, not Dynamic Redirects.

## Open defects observed today, NOT this lane's

- **`apply_site_design` fails on every full build.** Both proof runs died there:
  spawn returns `initialized:true` with real job topics, the `call_agent` times
  out after 3 retries (`CHILD_ORCHESTRATION_FAILED`), and **no child
  orchestration row is ever created**. Every OTHER `call_agent` in the same runs
  succeeded — that asymmetry is the narrowing fact. Contributed as two dated
  reproductions to `bugs_open/029` (hung spawns); no webdesign-agent pod or job
  exists afterwards, and the job object is cleaned before you can read its exit.
  **Cheap next check: `kubectl get jobs -w | grep webdesign` across a build.**
- `image-build-handler` ran 3× in the SWO proof run and ended `complete_error`
  each time. Not chased (sacrificial domain).

## Traps paid for — do not re-derive

- **Per-agent pods keep ~11 SECONDS of logs**, and only `agent-chassis` is a
  Deployment (the rest are spawned per run). Attach `logs -f` BEFORE dispatching.
- **"hero.* and logo.* differ in bytes" is NOT disconfirmable** and is written
  into both bug files as the bar. The deploy re-encodes per purpose, so they
  differ even when the wrong source is fetched. Assert the **downloaded object
  key** and the **stamped asset row**.
- **`s3_uri` is in the spec's `Optional`, so Strategy 0 resolves a config dotted
  path REGARDLESS of `input_fields`.** "Excluded from input_fields" suppresses
  only the aggressive search.
- **`StoreAssetAction` resolves purpose literal-first** (`v3_site_actions.go:2662`):
  `config["purpose"]` beats `config["purpose_field"]`. Adding the field without
  deleting the static changes nothing while looking like a fix.
- **Read `orchestration_states.error` BEFORE naming a cause.** I attributed a
  dispatch failure to the famous spawn→call race; it was my own dangling
  `input_mapping` path, and the error column said so. Cost two dispatches
  (`WRONG_CALLS.md` 2026-08-09).
- The SWO dispatch needs the working pageflow message shape: outer `load_site`,
  then `reviewed_brief: site_record.content_data` + `input_data: input_data`
  (the agent CONTRACT requires both). Fixed script: `fire_209_proof.sh`.
- A fleet roll landed mid-session between a config check and a dispatch; re-verify
  config by CONTENT after any roll (`updated_at` is re-stamped fleet-wide).
- `sites` has no `site_id` column (it is `id`); `content_data` is an **array** on
  some sites, so `jsonb_object_keys` errors there.

## Cold-start checks

1. `git log --oneline b8509054a..HEAD -- platform/orchestration/actions/deploy_image_asset_action.go bugs_open/209_* bugs_open/231_* bugs_open/235_*` — empty = ground unmoved.
2. `go test ./platform/orchestration/actions/ -run 'TestExtractActionInputs_|TestDeployImageAsset_|TestLegacyLogoStep_|TestPurposeFieldBridge_|TestStrategy0DottedPaths_|TestMigration348Shape_'` — 7 expected.
3. Migration 360 still in force (config drifts): the brand step must show
   `purpose` ABSENT and `purpose_field = input_data.spec.purpose`, while
   `store_hero_asset`/`store_logo_asset` correctly KEEP their statics (they are
   single-purpose steps). Query in `bugs_open/235`.
4. The package's one failing test (`TestValidDocSubjectTypes_Lockstep…`) is the
   064-shape recurrence owned by `idea_uk_vm_site` — pre-existing, not this lane's.
