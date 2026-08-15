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

---

## FIX SHIPPED IN CODE 2026-08-12 — commit `930ace3bd`, INERT until the next chassis roll

By the filing lane (`staged_component_build`), at the owner's instruction. The three
contributing lanes above each disclaimed the fix, so it was unowned; nothing here
overrides anyone's claim.

**Answering the owner's question first — no, this is not a missing handler.** The handler
exists (`asset-deployer`), dispatches correctly, commits to the right repo and reports
honestly. Both defects are in how it RESOLVES ITS TWO INPUTS, plus one detector that omits
a field. Adding a handler would have added a fourth thing to get the inputs wrong.

### What changed

1. **`deploy_image_asset_action.go` — rung 2 deleted.** `config["asset_key"]` is no longer
   readable as a literal filename. Deleted rather than gated, for the same reason
   `deploy_path` was: no caller-side discipline can make it safe, because the readers can
   never see the caller's choice.
2. **…and the CLASS guarded, not the instance.** An `asset_key` containing a dot is an
   unresolved path expression whatever produced it — a future config with the same shape in
   a different key, a mapping typo, a spec passing a path. Measured: 478 asset rows, none
   empty, **none containing a dot**, so discarding cannot lose a real key.
3. **`action_inputs.go` — `ActionInputs.WasDefaulted`.** The root of (b) is sharper than
   this file first recorded: `spec.Defaults` are written into `Values` **before** Strategy 1
   runs, and Strategy 1/2/3/4 all skip a field that already holds a value. So **any field
   with a default is unreachable by the recursive search** — only a Strategy 0 dot-path can
   set it. That is why `asset_key` (no default) could be found at `spec.asset_key` while
   `purpose` (default `hero`) could not. Additive: no value in `Values` changes.
   Registered as **CAP-002**.
4. **The asset ROW is now the authority for both.** When the run states no purpose
   (`WasDefaulted`) and/or no asset_key, and `asset_id` names a row, `assetRowIdentity`
   supplies them. The row is the one source a dispatcher's `input_mapping` cannot drop —
   which is what makes this fix cover the `image-build-handler` path (the 08-11
   contribution) as well as the repair path, without touching a shared mapping.
5. **`check_undeployed_assets.go` emits `asset_key`** in its spec. Carried even though the
   deployer now reads the row too: the spec is what a reader of the ITEM can see, and an
   item whose spec omits the field its handler needs is unreviewable.

**Ordering matters and is pinned:** row identity resolves BEFORE the brand-head refusal,
which reads `purpose`. Resolving it after would let a favicon dispatched with no stated
purpose through the guard as a "hero" and overwrite the artefact `derive_brand_head_assets`
owns — committed to git before any lock guard runs.
`TestDeployImageAssetAppliesTheRowPurposeBeforeTheBrandHeadRefusal` fails if it moves.

### Deliberately NOT done

- **`build-dispatch-loop`'s `input_mapping` is untouched.** Adding `purpose`/`asset_key`
  there would have fixed (b) live, with no roll — and that is exactly why it was rejected:
  it widens a seam every handler dispatch in the fleet passes through, to fix one action.
  Recovering the values inside the action is narrower and needs no shared-seam change.
- **Nothing hand-renamed in any bucket or repo**, per this file's own instruction.
- **The stray `/assets/images/logo.jpg`** I left on gaswholesalers.com on 08-10 is still
  there, still unreferenced, still mine. After the roll the correct `logo.png` will deploy
  alongside it; the stray should then be removed.

### Proof, and its limits

Mutation-tested — each fix watched to fail on its own mutant, tree restored, suite green:

| mutant | test that failed |
|---|---|
| restore rung 2 | `TestDeployImageAssetNeverUsesConfigAssetKeyAsALiteralFilename` |
| disable the row-purpose branch | `TakesPurposeFromTheAssetRow…` **and** `AppliesTheRowPurposeBeforeTheBrandHeadRefusal` |
| drop the `Defaulted` population | `TestDefaultedMarksASpecDefault` |

`git archive HEAD` builds clean and both packages' suites pass from it (shared-tree rule).
Council: `Council-Submitted: 7f0c1535-25cb-4645-adba-f7429e357a79` — **verdict not yet read
at the time of writing; a REVISE/REJECTED is still owed action, and the code is already on
the shared branch.**

**This bug stays OPEN.** The bar is fixed AND live, and a Go change is inert until the
fleet rolls (`make release`, owner-run). **After the roll, verify at the ARTEFACT and by
the EXTENSION, never the status** — a `logo` purpose must produce `logo.png` and the commit
message "Deploy logo image". A run reporting `complete` having written `logo.jpg` is this
bug, unfixed:

```bash
curl -s -o /dev/null -w '%{http_code}\n' https://gaswholesalers.com/assets/images/logo.png
curl -s -o /dev/null -w '%{http_code}\n' https://mortgagecalculator.co.uk/assets/images/hero.jpg
```

And re-run the census on `updated_at`, not `created_at` (the 08-12 contribution's lesson):
the count should stop rising. **It will not fall on its own** — 150+ existing rows point at
files already committed under the placeholder name, and nothing re-deploys them. Draining
that backlog is a separate, and still undesigned, piece of work.

---

## CONTRIBUTION 2026-08-13 — R1 REVISE, HIGH objection answered with a live config fix; resubmission still owed

By the filing lane (`staged_component_build`). `Council-Submitted: 7f0c1535-25cb-4645-adba-f7429e357a79`
came back **REVISE, round 1**, decided by a gating HIGH objection from `editquality`:
`assetRowIdentity` recovers `purpose`/`asset_key` via `inputs.Get("asset_id")`, and if
`asset_id` itself isn't explicitly mapped, it resolves through the landmined
`findFieldRecursive` aggressive search — a new failure mode able to silently pull an
unrelated asset's row. The reviewer's own proposed check (does `deploy_asset` declare
`input_fields` for `asset_id`?) came back yes, but that doesn't settle it: declaring a field
in `input_fields` still routes through the same aggressive `ExtractFields` search
(`action_inputs.go` Strategy 1 calls the identical function as the no-`input_fields` Strategy
2) — whether `asset_id` is actually resolved explicitly depends on the CALLER, not the
callee's declared fields.

Traced both callers: `check_undeployed_assets`'s repair-item dispatch already carries
`asset_id` in the item's `spec` — safe. **`image-build-handler`'s `call_asset_deployer` step
does not map `asset_id` at all** (`input_mapping: {domain, s3_uri, purpose, asset_key?}`),
and `call_agent`'s `ResolveInputMapping` builds the child orchestration's `input_data` from
ONLY the mapped keys, so the aggressive search is genuinely reachable on this path — the
objection's fear is real for this caller, whether or not it has yet misfired in production
`[UNVERIFIED]`.

**Fixed the caller, narrower still — one config key, no Go change, no roll needed:**
`store_imagery_asset` (the step immediately before `call_asset_deployer`) already returns
`asset_id` in its result (`asset_stored.asset_id`, `StoreAssetAction`,
`v3_site_actions.go:3076`), right beside `asset_stored.image_uri`/`asset_stored.purpose`,
which the mapping already reads. Added the fourth key, mirroring the existing `asset_key?`
optional-marker convention (safe for the `locked`/`no-URL` refusal branches, which return no
`asset_id`):

```sql
SELECT snapshot_agent('image-build-handler', '...');
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id?}',
  '"asset_stored.asset_id"'::jsonb), updated_at = now()
 WHERE type='image-build-handler' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Snapshot taken first (`099_SYNC_gate_roster.py`'s pattern), dry-run with `ROLLBACK` matched
the intended before/after, applied and **confirmed committed** live — `agent_definitions`
config is read live, so this needs no roll.

**Objection 2 (MEDIUM, edit 4's uncited "cause (c)")**: agreed the rationale overstates it —
"an item whose spec omits a field its handler needs is unreviewable" stands on its own
without inventing a third diagnosed cause. Resubmission should say so plainly rather than
defend the original wording.

**⚠ Cut short by kubeconfig expiry (fleet-wide, session-scoped, not this bug).** Resolved
next session: pulled the pre-fix evidence (5 real runs, `asset_id` empty not wrong),
registered `401` as a proper migration (`--record-only`), and resubmitted round 2. Full
reasoning: NOTES, `## 2026-08-13 (session 48fb60ee)`.

## CONTRIBUTION 2026-08-14 — round 2 ALSO revise: a second real gap on the repair path, plus two stale/mis-scoped landmine citations answered; round 3 resubmitted

Round 2 REVISE, `editquality` again gating: one HIGH was real (the
`check_undeployed_assets`→`build-dispatch-loop`→`asset-deployer` repair path has the
identical gap `401` fixed for `image-build-handler` — asserted "safe" in round 2 without
being traced). Two more were citations that didn't hold up on inspection: one landmine is
explicitly **RETIRED** (2026-08-06, `bugs_open/155`'s purpose-keyed source bug — the
objection quoted it as live), and one names a genuinely open bug (`bugs_open/235`,
brand-update branch hardcoding `purpose:'hero'`) but on a **different step**
(`store_imagery_brand_asset`, not `store_imagery_asset`) than the one this bug's fix
touches — not folded in, not this plan's job.

Fixed the real gap the same way as `401`: **migration 402** mirrors an *already-reviewed*
precedent (`migration 380`, `bugs_open/231`) that fixed this identical nested-spec shape for
`purpose` on this SAME `build-dispatch-loop` mapping — same idiom
(`"asset_id?": "current_item.spec.asset_id"`), measured blast radius first (exactly one
`(item_type, handler_agent)` pair fleet-wide, 267 rows), applied, verified, registered.

Resubmitted round 3, `RESUBMIT_CORR=7f0c1535-25cb-4645-adba-f7429e357a79`, run `d0b465c1…`,
6 edits. Full reasoning and the retired-landmine/wrong-branch corrections: NOTES,
`## 2026-08-14`.

## CONTRIBUTION 2026-08-14 (continued) — rounds 3 and 4 both REVISE; stopped resubmitting at an architecture-scope objection; THE FIX IS NOW PROVEN LIVE AT THE ARTEFACT

Round 3 (`guardian`): two HIGH, both checked directly and both clean —
`build-dispatch-loop`'s live config confirmed `sub_workflow` (not the landmined `substeps`)
is the shape that actually runs, and neither target agent type carries a second, higher
version row. Round 4 (`bug_historian`): a **different kind** of objection —
architecture-scope ("should the shared fallback strategy ever run for an unmapped field, at
all?"), explicitly flagged by the reviewer itself as something for a human, not another
round. Its two proposed checks came back clean too (no third live caller of
`asset-deployer`; its `input_contract` doesn't list `asset_id` at all). **Stopped
resubmitting here** — a scope objection isn't answered by more evidence on the same
submission (CLAUDE.md's own council-gate guidance). Routing it to architecture review is
owed, not done.

**Then the fleet rolled independently, and this bug's Go fix (`930ace3bd`) is now confirmed
live** (`v1.0.1298`, `git merge-base --is-ancestor` of the running build). Proved it rather
than trusting the commit message: promoted the standing `unresolved` repair item for
gaswholesalers.com's logo (`edff6d42…`, never previously triaged) to `triaged` by hand and
watched it run. **It deployed `/assets/images/logo.png` with commit message "Deploy logo
image for gaswholesalers.com"** — correct extension, correct purpose, both wrong on this
exact asset's two prior (pre-fix) attempts. **Verified at the served page, not the work
item's status: `curl https://gaswholesalers.com/assets/images/logo.png` → HTTP 200, 42,211
bytes.** The symptom this file opens with — 404, four months — is resolved.

**`mortgagecalculator.co.uk/assets/images/hero.jpg` was not re-tested** — same mechanism,
different site; say "expected to work" not "verified" until someone actually triggers and
checks it. **The ~146-row backlog will not drain on its own** — this proof is one asset,
by hand; draining the rest is still the separate, undesigned job this file has named since
2026-08-10.

Full round-by-round detail: NOTES, `## 2026-08-14 (continued)`. Cold-start for whoever picks
this up next: `staged_component_build/HANDOFF_2026-08-14_continue_here.md`.

## CONTRIBUTION 2026-08-14 (continued further) — mortgagecalculator's hero was the SAME defect, not a second mechanism; now proven live too

A later session in the same lane picked up exactly the open question the previous entry left:
"expected to work, not verified" for `mortgagecalculator.co.uk`. The intervening handoff
(`HANDOFF_2026-08-14_continue_here.md` §2b) had gone looking for a promotable stalled item
the way gaswholesalers had one, found none (all 5 `placeholder_image_in_use:hero` attempts on
this site are terminal — 3 cancelled, 2 complete), and stopped rather than speculatively
dispatching a fresh one, flagging "there may be a second, different mechanism in play here."

**There is not.** This file's own 2026-08-12 contribution already recorded the evidence: all
five of this site's 2026-08-11 hero generations (`477838e3`, `d6ead260`, `9e94250d`,
`0e11c818`, `2e2bea17`) deployed to `/assets/images/input-data.asset-key.jpg` — the exact
placeholder this file is about — via the ordinary **`image-build-handler` →
`call_asset_deployer`** route. That is precisely the caller **round 1's `migration 401`
fixed** (confirmed live since 2026-08-13, unchanged through the second fleet roll). The
repeated `rejected`/`superseded` asset-row history that looked like a distinct, tangled
problem is fully explained by this bug's own defect, once — round 1 fixed the exact path
these attempts used.

**Proved it the same way as gaswholesalers, since no item was left to promote a fresh one had
to be created** — cloned the last discovery-filed `needs_hero_image` item
(`067a7ad8…`, same `spec`, same `item_key: placeholder_image_in_use:hero`) straight to
`status='triaged'`, `handler_agent='image-build-handler'`. Claimed by `build-dispatch-loop`,
complete in ~2m30s (image generation is slower than the repair path's simple re-deploy).
`deploy_result.file_path = "/assets/images/hero.jpg"`, `commit_message = "Deploy hero image
for mortgagecalculator.co.uk"` — correct path and purpose, both of which every one of this
asset's five prior attempts got wrong. **Verified at the served artefact**: the first curl
immediately after completion still 404'd (git→publish propagation lag, not a failure — a
retry ~20s later, and a cache-busted URL, both returned **HTTP 200, 96,755 bytes**). The
`assets` row is now `status='active'`, `filename='hero.jpg'`, `url='/assets/images/hero.jpg'`.

So both named symptom sites (gaswholesalers' logo, mortgagecalculator's hero) are now
confirmed fixed at the wire, via the two different callers rounds 1 and 2 patched
(`image-build-handler`/401 and `build-dispatch-loop`/402 respectively) — real coverage of
both migrations, not just the one gaswholesalers happened to exercise. **Still true and
unchanged**: this is two assets, by hand; the ~146-row backlog is still the separate,
undesigned drain job, and routing R4's architecture-scope objection to a human/RFC is still
owed.

Full detail: NOTES, `## 2026-08-14 (mortgagecalculator hero retest)`.

## CONTRIBUTION 2026-08-14 (fresh session) — R4 routed to `RFC_029`; the backlog-drain job DESIGNED (not executed) — the flat "146 rows" figure hides four different buckets, and a blind bulk re-trigger would be wrong for two of them

**R4 routed.** Filed
`architecture_review/RFC_029_the_aggressive_recursive_search_has_no_boundary_for_an_unmapped_field.md`,
committed `439382985`. It is a companion to the already-open `RFC_028` (same resolver, a
different arm): RFC_028 audits `ExtractActionInputs`'s own eight arms in `action_inputs.go`;
this bug's R4 objection is about a **second, nested five-strategy chain** one level down, in
`unified_extractor.go` (`ExtractFields` → `extractSingleField` → `findFieldRecursive`), which
RFC_028's own census could not see because its query matched `action_inputs.go` text only.
Both migrations 401 and 402 routed around this exact arm; neither touched it. RFC_029 asks
whether an unmapped field should ever reach it, points at `WFA-009`'s already-shipped opt-in
`required` idiom as a precedent for the shape of an answer, and flags a live naming collision
(two independently-numbered "Strategy 4"s across the two files — migration 402's own comment
already cites the wrong one).

**The drain job — designed, and it is not the one-line job the earlier framing implied.**
Re-measured live rather than trusting the 08-12 figure (150/16) or the newer handoff's
rounding (~146/15):

```sql
SELECT count(*), count(DISTINCT site_id) FROM assets
WHERE filename LIKE '%asset-key%' OR url LIKE '%asset-key%';
-- 140 rows / 14 sites, 2026-08-14 ~17:00Z
```
By purpose: icon 79, content_hero 25, logo 14, illustration 10, hero 4, sprite_sheet 2,
og_card 2, favicon 2, brand_logo 1, blank 1. By site: dartsonline.com 28, robot-hands.com 20,
gamesdesign.co.uk 17, fundamentallyai.com 16, leopardessconsulting.co.uk 15, webdesign.co.uk
12, vetcomparison.uk 7, finetuning.uk 6, webdesign.uk 5, lendzy.co.uk 4, oufe.com 4, idea.uk 4,
vonc.com 1, ai-agent-orchestration.com 1. **Neither gaswholesalers.com nor
mortgagecalculator.co.uk appears** — both already repaired, dropping out of the marker on their
own, which is the query behaving correctly, not a new gap.

**⚠ This query is still a floor, per this file's own 08-12 finding (§ "the FLEET CENSUS
UNDER-COUNTS this bug")** — mortgagecalculator's own pre-fix hero rows never matched it
(`filename=''`, presigned S3 `url`), and that class is real and separate. Nothing below
resolves that; it is called out explicitly where it matters (bucket E).

**The mechanism the two proofs used generalises, but the 140 rows are not one bucket.** Joined
each marker row to its own `undeployed_asset` work-item history:

| bucket | count | state | correct action |
|---|---|---|---|
| A | 13 | `unresolved`, no promotion yet | promote to `triaged` — exactly what §2 did for gaswholesalers |
| B | 26 | already `triaged` | **nothing** — `build-dispatch-loop` is alive (last completion 15:38:34Z, ~90 min before this measurement) and periodic, not stalled; these self-drain |
| D | 64 | only terminal (`complete`/`cancelled`/`rejected`) items exist | needs a **fresh cloned item** at `status='triaged'` — exactly what §2b did for mortgagecalculator, no promotable row exists |
| E | 30 | no `undeployed_asset` item was ever filed for this `asset_id` | **do not act blindly** — see below |
(counts sum to 133, not 140; the discrepancy is a handful of rows with more than one
matching item where the bucketing query's tie-break picked one — [UNRESOLVED], small enough
not to change the shape of the plan, flagged rather than silently rounded away.)

**Bucket E is why a blind "re-trigger all 140" design would be wrong, and reading
`check_undeployed_assets.go` is what found it.** Its own query
(`findUndeployedAssets`, `check_undeployed_assets.go:264-297`) only flags a row when **no
deployed page component's `rendered_html`** already references
`/assets/images/<purpose>.*` — so a row can carry the placeholder marker in `assets.filename`
while the *live, currently-served* page already links the correctly-named file (a later,
independent re-deploy corrected the page without correcting the stale row). The code's own
comment (lines 307-323) names this exact shape for favicon/og_card and *deliberately does not
file against it*, calling a filed item there "a FALSE claim, which is worse than a missed
finding." Bucket E (30 rows, no item ever filed) is the general-purpose-field version of that
same shape, plus the two `brandHeadPurposes()` (favicon, og_card — 4 rows total, structurally
excluded from this check by `NOT (purpose = ANY(...))`) plus `hero` (4 rows), which this check
never covers at all — `hero`/`logo` brand-level gaps are found by a **separate** mechanism
(`needs_hero_image`/`needs_logo_image` → `image-build-handler`, the route mortgagecalculator's
repair actually used in §2b, not `check_undeployed_assets` → `build-dispatch-loop` at all).
**Before acting on any bucket-E row, curl its expected `/assets/images/<purpose>.*` path first**
— some will already be 200 (stale metadata only, no redeploy needed) and some will 404 (a real
gap that needs the same clone-to-triaged treatment as bucket D, via whichever discovery
mechanism actually covers that purpose).

**The repair path itself (buckets A/B/D) is LLM-free and unaffected by today's fleet-wide LLM
cap.** Read `check_undeployed_assets.go` and `deploy_image_asset_action.go` end to end: neither
calls an LLM or image-generation action — the repair re-deploys an **already-stored** image
under the correct name via a plain DB read + storage read + git commit
(`sendGitCommitRequest`). That is consistent with gaswholesalers' 53-second turnaround (§2).
**This does NOT extend to hero/logo regeneration** (bucket E's non-brand-head-purpose subset
that turns out genuinely broken) — that route runs through `image-build-handler`, which DOES
call image generation, and the fleet-wide LLM cap is confirmed live as of this session
(`llm_call_log`: 0 ok / 17 failed at 16:00Z today, partially recovered by 17:00Z — re-check
before dispatching any bucket-E item, this snapshot is not a standing guarantee either way).
Commit `8b897432a` (same day, different lane) is the fleet-wide record of the cap itself.

**What I deliberately did NOT do this session: fire the batch.** Buckets A and D together are
~77 rows across 14 live sites — real git commits to real customer-facing repos. The prior
handoff's own framing ("deserves its own careful plan … rather than a one-off script
improvised at the end of a session") is right, and nothing about today's investigation changes
that: the bucket-D mechanism is proven exactly twice (gaswholesalers, mortgagecalculator), not
load-tested at 64-row scale, and bucket E needs a per-row wire check before it even joins a
batch. **Recommended next step for whoever picks this up:** pilot bucket A (13 rows, cheapest —
promote only, no cloning) in one small batch with a wire-check per item before scaling to
bucket D.

**One more open question this design surfaced and could not close:** the earlier handoff's
phrase "which sites are `owned` vs `generic` rebuild policy" does not correspond to any column
found on `sites` in a direct check (`settings`, `locked_at`/`locked_by`, `status` — none of the
14 affected sites are locked, none carry a `settings` key resembling this). **[UNMEASURED]**
whether this classification lives elsewhere (a different table, a convention in `network_id`
grouping, or purely tribal knowledge) — flagging rather than guessing. Whoever executes the
drain should confirm this before batching sites that might have different ownership/consent
requirements, per CLAUDE.md's "other consumers must be told, not merely measured."

Full detail: NOTES, `## 2026-08-14 (drain job design)`.

## CONTRIBUTION 2026-08-14 (same session, later) — bucket-A pilot EXECUTED: 11/11 clean, census 140 → 98, and the wire-check-first rule caught two would-be regressions before they happened

Ran the drain design's bucket-A pilot (user-approved plan, revised once after review — the
revision is the story). **Result: 11 items promoted, 11 completed, zero placeholder paths in
any `deploy_result`, every deploy verified 200 at the wire, all 11 asset rows corrected.**
Fresh census after: **98 rows / 12 sites** (was 140/14 at this session's start — the drop of 42
is this pilot's 11 plus bucket B's concurrent self-drain, which the design predicted and this
confirms).

**The revised rule — wire-check BEFORE promotion, gating it — earned its place immediately:**

- **2 SKIPPED as live-referenced 200s**: finetuning.uk `logo` and leopardessconsulting `logo`
  both already serve `/assets/images/logo.png` 200 and the leopardess homepage references it.
  Promoting them would have overwritten a served, referenced file with the stored asset-row
  image — the "fixed by another route" shape this file's own "measured the OTHER way" section
  documents. These two rows stay in the marker census deliberately: **"stale row only" is a
  bookkeeping class, not a deploy class**, and force-redeploying to clean a row is backwards.
- **1 CANARY, clean**: leopardess `brand_logo` (404 on both extensions, referenced by nothing).
  Promoted `unresolved`→`triaged` by hand; claimed in ~60s, complete in ~80s;
  `file_path: /assets/images/brand_logo.jpg`, wire 200 (162,089 bytes), row corrected.
  **Naming finding worth keeping**: the underscore is PRESERVED (`brand_logo.jpg`, not
  `brand-logo.jpg`) — `BuildAssetPaths` uses the purpose verbatim when `asset_key == purpose`,
  while `AssetKeyFilename` maps `_`→`-` only when they differ (`url_helpers.go:182,210,317`).
  Both writer and reader derive the same name, so it is consistent, not a defect — but the same
  string produces different filenames depending on which role it arrives in. Latent wrinkle,
  noted, no action.
- **10 WAVED (webdesign.co.uk icons) — with a mid-pilot hypothesis correction recorded
  honestly**: these were promoted on the belief they were wrong-extension deploys (`icon-*.jpg`
  exists, `.png` 404s, assumption: icons should be `.png`). Reading `ImagePurposes`
  (`url_helpers.go:365`) REFUTED that mid-flight: icon's configured extension **is** `.jpg`, so
  `icon-css.jpg` was the correct reader-derived path all along and these rows were actually the
  same "stale row only" class as the two skips. The promotions stood because the risk profile
  is different from the logo skips: **zero deployed `page_components` reference any icon at
  either spelling** (checked in the DB, not assumed), so a same-path redeploy could not regress
  anything a visitor sees. Outcome: all 10 deployed to their correct dash-named paths,
  content-lengths **byte-identical** to the pre-overwrite baseline (captured deliberately before
  the wave landed — the stored sources are the same images), and all 10 rows corrected through
  the framework rather than by hand-editing `assets`.

**What this proves beyond §2/§2b**: the promote-to-triaged mechanism at 11-item batch scale
(the loop claimed and completed them serially, one at a time, ~30-40s each — a natural rate
limit, no parallel blast), and the two failure modes the revised plan guarded against (live-200
overwrite; double-dispatch via a second open item) both actually occurred in the candidate data
and were caught by the checks, not by luck.

**What remains for the drain** (unchanged in kind, updated in number): bucket B self-drains —
leave it alone; bucket D (the ~64 clone-needed rows) is now the bulk of the remaining 98, and
its design should inherit the wire-check-first gate verbatim (some of its rows will be
already-200 skips too — this pilot measured 2-of-13 in bucket A; assume the rate is nonzero, not
that it transfers); the ~30 bucket-E rows still need their per-row wire check before any item is
even filed; and the "stale row only" class now has 12 confirmed members (2 logo skips + the 10
icons would have joined it had the census been the only test) — whether those rows deserve a
bookkeeping-only correction path that does not redeploy is a small design question for whoever
takes bucket D, not something to improvise.

Full detail: NOTES, `## 2026-08-14 (bucket-A pilot)`.


## Contribution — leopardessconsulting.co.uk numbers, and two clean new rows (2026-08-14, services-restore session)

Measured while restoring `/services.html` (its six teaser icons were rewired today; the page
now references all six files directly in `content_data`):

- Placeholder-URL rows on this site: **13** now
  (`url='/assets/images/input-data.asset-key.jpg'`), down from the **15** measured
  2026-08-12. Not investigated which two moved or why — noting the delta for whoever runs
  the drain.
- All six `icon_service_*` FILES serve 200 with distinct sizes (26–47KB) while their rows
  carry the placeholder URL — the "stale metadata only, no redeploy needed" shape the
  bucket-E note above describes. A curl-first pass over these 13 would likely classify most
  as metadata-only.
- **Two new rows created today via the same Route-A recipe were born with CORRECT urls**:
  `icon_service_routing` / `icon_service_credibility` (created 18:20/18:21Z via scope-less
  `needs_imagery` → image-build-handler → Banana), both
  `url='/assets/images/icon-service-<n>.jpg'`, both serving 200 within a minute of creation.
  Same recipe as the 07-31 six whose rows today carry the placeholder. Fact only — I have
  not traced whether the 07-31 rows were placeholderised at creation or later by the repair
  path this file documents.

## CONTRIBUTION 2026-08-14 (late) — the census counts 11 rows that must NEVER be redeployed, and bucket A is a SKIP SIGNAL, not remaining work

From the lane that filed this bug, returning after two days and re-deriving the drain numbers
rather than inheriting them. **No drain action taken.** Two findings that change the target list
for whoever runs bucket D. Fresh handoff: `staged_component_build/HANDOFF_2026-08-14c_continue_here.md`.

### 1. The marker census does not filter `assets.status` — 11 of the 98 must never be republished

```
active 87 · superseded 10 · retired 1
```

A `superseded` or `retired` row has been **replaced by a newer asset**. Redeploying its stored
bytes pushes a stale image over a current one — the same harm as the live-200 overwrite the
pilot's wire-check caught, arriving through a different door, and the wire check would NOT catch
it (the path may legitimately 404, so the row looks like honest work).

**The real drain target is 87, not 98.** Add `AND a.status='active'` to every bucket query.
Active-only buckets: **D 57 · E 27 · B 2 · A 1**. The pilot did not hit this only because
bucket A happened to hold one such row.

### 2. Bucket A's two rows are the pilot's OWN deliberate skips

| site | purpose | asset_id | status | `logo.png` at the wire |
|---|---|---|---|---|
| leopardessconsulting.co.uk | logo | `71652e42-36d3-42f3-a271-700b05920ad3` | **retired** | **200** |
| finetuning.uk | logo | `9c9de5a0-a830-4706-a4dc-36d86a61eea9` | active | **200** |

These are exactly the two the bucket-A pilot skipped as live-referenced 200s. Both still serve.
**A fresh session reading "bucket A: 2 rows — promote them, the pilot proved that action" would
reproduce the precise regression the pilot avoided**, and on leopardess would serve a *retired*
asset's bytes over a live logo.

Nothing in the row, the item, or the census records that a human decided to leave them. The
decision exists only in the pilot's contribution above. **Generalising: a small residual bucket
immediately after a pilot is a skip signal more often than an omission — grep this file for the
site before acting on a remainder of one or two.**

### Also re-verified (so the next session need not)

- Fix **live in the running binary**, `v1.0.1300`, controls both ways: the two log literals
  PRESENT, a bogus needle absent. Both symptom sites still 200.
- **The LLM cap has recovered** — `llm_call_log.success` over five hours: 24/124/53/48/37 ok vs
  0/0/1/1/1 failed. Bucket E's regeneration subset is no longer cap-blocked (cap itself still
  nominally runs to 2026-09-01, so re-check rather than assume).
- Bucket concentration for whoever designs D: **dartsonline.com alone is 28 of the 57**, and three
  sites hold 45. Per-site canary is cheap; fleet-wide batching is not.
- A **corrected bucket query** that sums exactly (the earlier one reconciled to 133 of 140 —
  the gap was assets matching more than one work item; aggregating per asset with `bool_or` fixes
  it) is in `NOTES_staged_component_build.md` `## 2026-08-14 (c)`.

## CLOSING CONTRIBUTION 2026-08-15 — the drain EXECUTED, and the wire inverted the design: 84 of 84 rows were already served, zero redeploys fired, backlog now 0 active

Owner rulings received in chat this morning: (1) run the whole drain, after checking no other
thread is working the same pages; (2) the "stale row only" class gets its bookkeeping done
MANUALLY now, revisit only if the class recurs; (3) the "owned vs generic" question is
answered — there IS an existing site lock system (`sites.locked_at`/`locked_by`, honoured by
`find_dispatchable_site`), and none of the 12 affected sites need locking during the current
heavy-development phase.

**Coordination check (the ruling's own precondition), verified before anything fired:**
- Zero open (`triaged`/`approved`/`detected`/`claimed`) `undeployed_asset` items fleet-wide —
  no competing drain. The 268 lane's CTA fleet batch IS live on 7 of the 12 sites
  (`page_rerender` wave, 08:35Z) but touches page HTML, not asset files; its own pre-flight
  measured 248-exposure zero (`942883ef8`).
- **52 dormant `unresolved` items exist that the earlier "no open items" reads missed** (they
  sorted below a LIMIT in the first query — corrected here): ~44 on robot-hands.com from
  July, 3 on gaswholesalers, 2 finetuning, 1 leopardess, 1 ai-agent-orchestration. Dormant,
  never dispatched, no concurrency risk. The three pointing at census rows are dealt with
  below; the ~49 pointing at NON-census assets are an off-census observation for whoever owns
  discovery hygiene — recorded in the 08-15 handoff, not this bug's defect.
- Fix re-verified in the RUNNING binary (`v1.0.1300`, both literals present, negative control
  absent, `/proc/1/exe` probe).

**The wire sweep — the load-bearing measurement.** Every bucket-D and bucket-E row's
reader-derived path (`DeployedAssetPath` semantics mirrored in SQL: brand-head fixed names;
logo `.png`; else `.jpg`; `asset_key==purpose` verbatim, else `_`→`-`) was curled with three
controls: a per-domain must-be-absent path (**12/12 returned 404** — no catch-all fallbacks),
content-type (**80/82 `image/*`**, 2 header flakes on fundamentallyai that retried clean as
`image/jpeg` with distinct sizes), and content-length captured per row. Result:

> **84 of 84 rows (D 57 + E 27) serve a genuine image at the derived path. 200s, all of them.**

The pilot measured a 2-of-13 already-served rate in bucket A and the design said "assume
nonzero" for D. The truth was **100%** — between the fix going live and today, rerenders,
regenerations and bucket B's self-drain had already repaired every artefact; only the rows
never learned. A drain that skipped the wire-check gate would have fired **84 pointless
redeploys, each one overwriting a correctly-served file** — the exact regression the pilot
caught twice, at 42× the scale.

**Bookkeeping executed (owner-authorised manual correction), one guarded transaction:**
- Pre-image of all 85 active marker rows committed first:
  `staged_component_build/DATA_2026-08-15_bookkeeping_preimage.tsv`.
- Collision checks before the write: zero duplicate (site, path) pairs within the census;
  zero healthy active rows already claiming a target filename. The `UPDATE` guarded on
  `status='active' AND` the marker still matching, so it was idempotent and could not touch
  a concurrently-changed row. **`UPDATE 85`.**
- The three dormant `unresolved` items pointing at census assets cancelled with an audit
  note in `result` and `handled_by='claude-session-248-bookkeeping-20260815'`
  (`318eeb70…`, `462828c5…` — finetuning logo, the pilot's deliberate skip, item refiled by
  the 08-14 discovery sweep; `00d1dda0…` — leopardess's RETIRED logo, the row whose
  promotion would have served a retired asset's bytes). **`UPDATE 3`.**
- **Census after: 0 active marker rows.** 11 non-active remain (10 `superseded` + 1
  `retired`) — deliberately untouched: their bookkeeping describes replaced generations, and
  any future marker census MUST carry `AND status='active'` (14c's finding, now standing
  practice).

**Bug bar check — fixed AND live: BOTH now hold.** The code fix has been live since
`v1.0.1300` (proven at the binary with two-way controls, twice, two days running); both
original symptom sites serve 200; the backlog the defect created is drained to zero active
rows; the architecture question it surfaced is ruled (`RFC_029` §9, owner-delegated,
2026-08-15) with implementation tracked there, not here. **Moving this file to
`bugs_closed/`.** Residuals and where they live: RFC_029 implementation (that file, phased,
council-gated); the 11 non-active marker rows (harmless by status filter); the ~49 dormant
off-census `unresolved` items (08-15 handoff, observation only).
