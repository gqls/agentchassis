# 248 — the `undeployed_asset` repair path deploys every asset as a HERO under a placeholder filename, reports `complete`, and re-triggers itself for ever

**Filed 2026-08-10** by `staged_component_build`, while carrying out the owner's
instruction to fix the gaswholesalers.com logo (standing defect list item 2, "6+ days"
— it is in fact **four months**).

**Status: OPEN. Not fixed.** The immediate repair is blocked on a shared-agent config
change, which is platform-scope and belongs in the council gate, not in a late-evening
patch on a tree this many sessions share.

**Diagnosis loop: CONFIRMED, first iteration** — `090` correlation
`b78e9a04-9a91-4261-af86-fb79f9316a4e`, run `8cb3778d-c3e6-4dd8-9e80-09c0d1b0e594`.
It re-read the same functions independently and cited the same 20 rows. (Filed per the
2026-07-31 owner ruling: a cross-cutting root-cause claim goes through the loop.)

## The symptom, at the artefact

`https://gaswholesalers.com/assets/images/logo.png` → **404** (2026-08-10, and on every
probe since the file was first asked for). The page's markup asks for exactly that path:
`<img src="/assets/images/logo.png">`.

The logo is **not missing**. It is deployed, and correct, at
`/assets/images/input-data.asset-key.jpg` — HTTP 200, 37,221 bytes, and I looked at it:
it is the real Gas Wholesalers wordmark the owner uploaded on 2026-07-29.

## Why nobody caught it: the repair reports success

The detector has raised this correctly and repeatedly — `undeployed_asset` items for
asset `b99c5355-4b3a-430c-9294-56482726be34` on **2026-04-10, 08-03, 08-04 and 08-09**.
Two of them are `complete`. The repair ran, committed a file to the site repo, returned
`success: true`, and the page went on 404ing — so the detector fired again next sweep.

**This is a closed loop that consumes a repair cycle per sweep and can never converge.**

## Root cause — TWO independent defects on the same path

Both live in how `deploy_image_asset` resolves its inputs when dispatched from a work item.

### (a) `asset_key` falls back to the literal path expression

`deploy_image_asset_action.go` resolves `asset_key` down a three-rung ladder, and its own
comment names the rungs:

```go
//   1. inputs.Get("asset_key")   — via input_fields config
//   2. config["asset_key"]       — literal string
//   3. config["asset_key_field"] — JSONPath into collected_data
```

The live `asset-deployer` row sets `config.asset_key = "input_data.asset_key"` — a dotted
**path**, correct for rung 1. When rung 1 resolves nothing, **rung 2 takes that same
string as a literal filename**. `storage.AssetKeyFilename` then maps `_`→`-`, giving
`input-data.asset-key.<ext>`, while every reader resolves the reference through
`storage.DeployedAssetPath` and gets the real name. Writer and reader disagree; the page
404s.

**Fleet-wide: 118 asset rows across 10 sites** carry that placeholder
(`assets.filename`/`url` LIKE `%asset-key%`), spanning purposes `icon`, `logo`, `hero`,
`content_hero`, `illustration`, `og_card`, `favicon`, `sprite_sheet`.

### (b) `purpose` can NEVER resolve on the work-item path, so everything deploys as a hero

`DeployImageAssetInputSpec` declares `Defaults: {"purpose": "hero"}`. The
`build-dispatch-loop`'s `call_handler` step builds the handler's `input_data` from a
fixed `input_mapping`:

```
{spec, domain, issue?, source, site_id, page_id?, item_type,
 page_name?, current_page, work_item_id, component_id?, reviewed_brief?}
```

**There is no `purpose` key and no `asset_key` key.** So `input_data.purpose` has nothing
to resolve against, the `hero` default wins, and `GetImageConfig("hero")` supplies `.jpg`.

That is why every one of the 118 placeholder files is a `.jpg` **regardless of its row's
purpose** — a logo, an icon and a favicon all land as `input-data.asset-key.jpg`.

**The asymmetry between (a) and (b) is the instructive part:** `asset_key` has **no**
default, so `ExtractActionInputs` falls through to a recursive search and *does* find
`spec.asset_key` when a spec carries one. `purpose` **has** a default, so the default
short-circuits before any such search. A field is easier to resolve when it has no
default than when it has one. Compare `bugs_open/231` (a static config value for a
spec-defaulted field is dead) — same mechanism, different field.

### (c) …and the detector's own spec supplies neither

`check_undeployed_assets.go:113-119` builds the repair spec as exactly:

```go
"check", "asset_id", "purpose", "asset_type", "url"
```

`purpose` is in there but cannot be read (see (b)); `asset_key` is **absent entirely**, so
(a) fires. Note also that the spec's `url` is copied from `assets.url` — which for these
rows already holds the placeholder path, so the item carries the corruption forward.

## What I measured, and one thing that came out the OTHER way

- **Only gaswholesalers.com actually 404s today.** The other four sites whose `logo` row
  carries the placeholder (finetuning.uk, leopardessconsulting.co.uk, vetcomparison.uk,
  webdesign.co.uk) all serve `/assets/images/logo.png` = **200**. `assets.url` is not the
  served path — they were fixed by another route. **So the 118-row count is the size of
  the corruption, NOT the size of the outage.** I have not enumerated which of the other
  113 rows are actually referenced by a page; that census is undone.
- Deploys **do** succeed through this path when driven by `image-build-handler`
  (11 logos deployed correctly today by another session's brand campaign, all
  `/assets/images/logo.png`). That handler supplies its own input_data. **The defect is
  specific to the `undeployed_asset` → `asset-deployer` route.**

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Delete rung 2.** A config value that is a path must never be usable as a literal
   filename. If rung 1 resolves nothing, the action should REFUSE (or fall back to
   `purpose`, which is what `DeployedAssetPath` already does for an empty `assetKey` —
   `assetKey == "" || assetKey == purpose` → the purpose-based path, which is exactly
   right for a logo). **This one line would have prevented all 118 rows.**
2. **Add `purpose` and `asset_key` to the dispatcher's `input_mapping`** (they exist on
   the item's spec; the mapping simply does not pass them). Fixes (b) for every handler.
3. **Have `check_undeployed_assets` emit `asset_key` in its spec** — necessary but not
   sufficient on its own, since (b) would still mis-type the extension.
4. Operational, not structural: a post-deploy assertion that the committed path equals
   `storage.DeployedWebPath(asset_key, purpose)` and HTTP-200s. This is D11 4-A item (4),
   "brand-head assets must HTTP-200 before the build reports success" — the same gate,
   generalised. Ranked last because it detects rather than prevents.

**Do not "fix" this by pointing pages at the placeholder filename.** The name is derived
on both sides from one function; the derivation is right and the writer's input is wrong.

## What I did tonight, and what I left behind

- Dispatched the corrected deploy twice by direct kcat publish to
  `system.agent.generic.requests`: both **FAILED — "storage client not available"**. The
  standing chassis has `IMAGE_BUCKET`/`S3_ENDPOINT` unset, so it can never build a storage
  client; only `agent-build-dispatch-loop` and `image-generator-adapter` pods have the
  bucket. **This is the operational half of `bugs_open/245`** (which records the same
  config from the credentials-exposure angle) — contributed there.
- Filed a corrected `undeployed_asset` item carrying `asset_key` explicitly
  (`3866b8a7-c8b1-40bf-aeaf-06271368b791`) and hand-promoted it `detected`→`triaged`,
  because the triage sweep is backlogged (**636 `detected` items, oldest 2026-07-24** —
  worth its own look).
- It was claimed and reported `complete` in 4 minutes, and **deployed
  `/assets/images/logo.jpg` with commit message "Deploy hero image"** — defect (b),
  caught in the act. `asset_key` resolved (my spec supplied it, proving (a)'s remedy);
  `purpose` did not.
- **⚠ I have therefore left a stray `/assets/images/logo.jpg` in `gqls/sites` for
  gaswholesalers.com.** It is unreferenced by any page (the markup asks for `.png`), so it
  is inert, but it is litter and it is mine. It should be removed by whoever takes this
  bug — I have not removed it, because deleting from a site repo is a separate write path
  I did not want to improvise at the end of a session.

**The logo is still 404 and I did not fix it.** The remaining step is a shared-agent
config change (candidate 2) or a Go change (candidate 1); both are platform-scope.

## Verify (for the fixing thread)

```bash
# the outage
curl -s -o /dev/null -w '%{http_code}\n' https://gaswholesalers.com/assets/images/logo.png   # 404
curl -s -o /dev/null -w '%{http_code}\n' https://gaswholesalers.com/assets/images/input-data.asset-key.jpg  # 200

# the corruption, fleet-wide
SELECT count(*), count(DISTINCT site_id) FROM assets
WHERE filename ILIKE '%asset-key%' OR url ILIKE '%asset-key%';   -- 118 / 10 on 2026-08-10

# the two inputs, at the live config
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'deploy_asset'->'config')
FROM agent_definitions WHERE type='asset-deployer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

SELECT jsonb_pretty(default_config->'workflow'->'steps'->'process_item'->'config'
       ->'sub_workflow'->'steps'->'call_handler'->'config'->'input_mapping')
FROM agent_definitions WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

After a fix, the disconfirming check is the **extension and the commit message**, not the
status: a `logo` purpose must produce `logo.png` and "Deploy logo image". A run that
reports `complete` having written `logo.jpg` is this bug, unfixed.

## Relations

- `bugs_open/245` — the same chassis storage config, from the credentials angle. Same lane.
- `bugs_open/231` — a static config value for a spec-defaulted field is dead. Defect (b)
  is that mechanism, on `purpose`.
- `bugs_open/235` — the brand-update branch stores every logo as a hero. Adjacent; this
  file is the *deploy* half of "everything is a hero".
- `bugs_open/155`, `209` — earlier source-resolution defects in the same action.
- `bugs_closed/128` — `image_url_404` detection, which has worked since 07-31 while repair
  never has. **This bug is why**: the repair does run, and it deploys to the wrong name.
  D11 4-A's premise ("items flag-only BY DESIGN") is **too kind** for this item type —
  `undeployed_asset` is not flag-only, it dispatches and it silently mis-deploys.

---

## CONTRIBUTION 2026-08-11 — the placeholder-filename half fires on the IMAGE-BUILD-HANDLER path too, not only the undeployed_asset repair

From the `bugfix_210_needs_logo_unhandleable` lane, in passing; no claim on this bug's fix.

A `needs_hero_image` item for mortgagecalculator.co.uk (`067a7ad8-1c25-4730-9e80-abd27893156f`,
completed 2026-08-11 10:36:20) ran the ordinary **`image-build-handler` → `call_asset_deployer`**
route — not this file's `undeployed_asset` repair path — and deployed to exactly this bug's
placeholder: `file_path: "/assets/images/input-data.asset-key.jpg"` (HTTP 200, 109,803 B, the
freshly generated hero), while the six pages that reference `/assets/images/hero.jpg` go on
404ing. Generation and storage were correct
(`s3://personae-prod-uk001-images/images/system/20260811/8a4d8d09….png`, asset `477838e3…`);
only the deployed FILENAME is wrong.

So whatever fix closes this bug should be measured against **both** producers of the literal —
the repair path this file diagnoses AND the image-build-handler's own deploy step — or the
placeholder will survive on the path that ships every routine hero/logo. (The hero purpose being
"right" here is incidental: the item genuinely was a hero. The filename half is the shared
defect.)

Not hand-renamed in the bucket, deliberately: the owner's framework rule, and this bug's own
"the repair reports success" loop, both argue against another hand-placed artefact. The pages
404 on `hero.jpg` today exactly as they did before the run — nothing regressed; the image waits
under the wrong name for this bug's fix to move it.

---

## CONTRIBUTION 2026-08-12 — the FLEET CENSUS UNDER-COUNTS this bug, and rung 2 is still live

From the `mortgagecalculator_couk_adoption` lane, reached from the owner's report that
"the hero image is no longer there". No claim on the fix. Extends the 08-11 contribution
above, which already named this site.

### 1. `assets.filename LIKE '%asset-key%'` is NOT a census of the defect

It is a census of **the defect AND a successful `assets`-row update**, and those are
independent. `deploy_image_asset_action.go:378-392` writes the local path onto the asset
row **best-effort, after the git commit**, and only when `asset_id` is among its inputs —
its own comment says a failure there "must not fail the deploy". So an instance where the
commit went to the placeholder path but the row update did not take is **invisible to the
census**.

**This site is exactly that case, and it is wire-proven:**

- `/assets/images/input-data.asset-key.jpg` → **200, 68,984 bytes, `image/jpeg`**
- `/assets/images/hero.jpg` → **404** (still referenced by the served homepage's inline
  `background-image`)
- and yet **all five** of this site's `hero` asset rows have `filename = ''` and a
  presigned S3 `url` — **none matches `%asset-key%`**.

So `mortgagecalculator.co.uk` contributes **0 rows** to the 118/150-row figure while
serving the 404 this bug is about. **Whatever number the fix is measured against, it is a
floor, not a count** — and the discrepancy is not small: this one site had **five** hero
generations on 2026-08-11 (`477838e3` 10:35, `d6ead260` 12:46, `9e94250d` 19:07,
`0e11c818` 19:10, `2e2bea17` 19:11) and none of them reached `/assets/images/hero.jpg`.

**A census that can only see instances whose bookkeeping succeeded cannot size a bug whose
symptom is bookkeeping that didn't.** To count the rest you have to ask the site repo or
the wire, not `assets`.

### 2. Re-measured on the right clock: 150 rows / 16 sites, up from 118 / 10

`created_at` is the wrong clock — the placeholder arrives via the post-commit UPDATE, so
grouping by `created_at` shows only **1** row since this bug was filed and reads as "it
stopped". By `updated_at`: **109 rows across 15 sites on 2026-08-11 alone**. Total now
**150 across 16 sites** (2026-08-12), against this file's 118 across 10 on 08-10.
[MEASURED] — the disconfirming result would have been a flat or falling count.

### 3. Rung 2 is UNCHANGED and still live — the `asset_key?` marker is NOT this fix

Both `asset-deployer` and `image-build-handler` rows were updated **2026-08-11 21:52:40Z**,
which is easy to mistake for fix candidate 1 or 3 landing. It is not:

```
asset-deployer      / deploy_asset              config.asset_key       = input_data.asset_key   <-- STILL THE LITERAL
image-build-handler / store_imagery_asset       config.asset_key_field = input_data.spec.asset_key
image-build-handler / call_asset_deployer       input_mapping."asset_key?" = input_data.spec.asset_key
```

The `asset_key?` optional marker sits on the **caller's** `input_mapping`, not on the
deployer's own `config`. `config.asset_key = "input_data.asset_key"` — the exact string this
file's §(a) names as rung 2's source — **is still there today**. Fix candidate 1 ("delete
rung 2", the one line that would have prevented all 118) is **not applied**.

Note the shape difference that makes this confusing: the literal that lands in filenames is
`input_data.asset_key` (from `asset-deployer`), while the newer mappings all say
`input_data.spec.asset_key`. Two spellings of the same intent in one path — so grepping for
the newer one and finding it "fixed" is a trap.

### 4. Also on this site, and probably the same root

`needs_hero_image` / `placeholder_image_in_use:hero` has been filed **five times** here
(3 `cancelled`, 2 `complete`) and `image_url_404:hero.jpg` has been **`blocked` since
2026-08-05**. The closed loop this file describes ("consumes a repair cycle per sweep and
can never converge") is running on this site too, on the `image-build-handler` route.

Nothing hand-fixed, per this file's own instruction and the owner's framework rule. Full
trace: `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/`
`HANDOFF_2026-08-11_continue_here.md` §11.1.

---

## CONTRIBUTION 2026-08-12 — a SECOND independent instance, same placeholder, different content

From the `bugfix_210_needs_logo_unhandleable` lane, in passing; no claim on the fix.

A fresh `needs_hero_image`/`image-build-handler` run for mortgagecalculator.co.uk
(item `68a4faf9-2868-487b-8e19-3fad4ef195c5`, `store_hero_asset` at 2026-08-11 19:08:27,
`origin_model=banana/gemini-3-pro-image-preview`) deployed to the **same** placeholder this
file diagnoses — `/assets/images/input-data.asset-key.jpg` — despite that exact path having
been **deleted from both bucket and repo** the same afternoon (12:36–12:41 UTC) after an
earlier, unrelated generation was rejected. So the placeholder isn't a stale leftover being
repeatedly re-served; the deploy step **reconstructs it fresh, every time**, independent of
prior state. Two independent hero generations for the same site, ~6.5 hours apart, both landed
there — a frequency data point for whoever measures how often this fires.

Also worth a line for the record: this run's `site_work_items.result` column does **not**
contain the image-generation outcome — it holds an unrelated `content-gap-planner`-shaped JSON
(`{"approach":"new_page","new_page":{"name":"guide-mortgage-affordability",...}}`). Checked
against 8 other `image-build-handler` completions from the same night: all 8 show the expected
`{"response":{"asset_stored":{...}}}` shape, so this is **not** the same defect —
isolated to this one row, cause not investigated. Flagging in case anyone chasing 248 also
reads `result` for evidence and gets a stale/wrong read from this row specifically.
