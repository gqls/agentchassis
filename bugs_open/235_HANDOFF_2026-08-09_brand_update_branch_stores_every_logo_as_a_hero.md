# 235 — `image-build-handler`'s brand-update branch stores EVERY asset as purpose "hero", so a logo item ships as `logo.jpg` at hero processing

**Filed:** 2026-08-09 by the `bugfix_209_deploy_purpose_keyed_source` lane, as the
answer to the producer question its own `bugs_open/231` contribution left
`[UNVERIFIED]` this morning.
**Status:** OPEN, unowned (`who-owns.py image-build-handler` → no match; grep of
`/bugs_open/` + `/bugs_closed/` for the mechanism → none).
**Severity:** medium-high — silent, wrong artefact on live sites; **11 sites serve
one now**, and every future brand-update logo item re-makes it.

## Method note (per the 2026-07-31 owner ruling)

A 090 ran first (run correlation `fd7ef7a9-93fb-4e20-9956-f8913bd4ab89`) and
returned **UNVERIFIABLE (scope-not-narrowing)** — it could not fetch
`asset-deployer`'s live step config (it read the empty `task_workflow`/
`orchestrator_workflow` columns, not `default_config`) nor the `ImagePurposes`
**var declaration** (`code_symbols` has no var kind — `bugs_open/223`'s
var-blindness, now on its **third** consumer). Its "still needed" list is
answered first-hand below, item by item; the live-config reads and the work-item
join are one query each and are quoted.

**And the 090's hypothesis was WRONG in its mechanism detail — recorded plainly.**
The filed hypothesis said "a caller supplies no resolvable purpose, so the spec
Default 'hero' fires" (bug 231's mechanism). In fact the purpose RESOLVES fine —
to a static `"hero"` a config author wrote on a branch that handles two purposes.
Same observable, different door. The 231 shadow remains real but is NOT what
produced these artefacts.

## The mechanism, every link cited from the live rows

1. A discovery check files `needs_imagery` with `spec.brand_update: true`,
   `spec.asset_key: "logo"`, **`spec.purpose: "logo"`** — the spec says exactly
   what it is. (Live rows: lendzy.co.uk 08-02 `complete`, webdesign.uk 08-04 and
   08-08 `complete`, mortgagecalculator.co.uk 08-02 `deferred` — that deferred
   item will produce a wrong logo when it drains.)
2. `image-build-handler.check_imagery_brand_update`:
   `condition: input_data.spec.brand_update == true` → `store_imagery_brand_asset`.
3. `store_imagery_brand_asset` config: **`"purpose": "hero"` (static)**,
   `asset_key_field: input_data.spec.asset_key`, output `asset_stored`. Its own
   description names the two-purpose reality the static ignores: *"logo or
   canonical index hero — rule b"*. One static value cannot be right for both;
   it is right for the hero case and wrong for the logo case, and the item's own
   `spec.purpose` is available and unread.
4. `call_asset_deployer` maps `purpose: asset_stored.purpose` (= "hero"),
   `asset_key?: input_data.spec.asset_key` (= "logo") →
   `deploy_image_asset`: `DeployedAssetPath("logo","hero")` takes the FILENAME
   from asset_key and the EXTENSION + resize class from purpose
   (`url_helpers.go:317-330`; `ImagePurposes["hero"]={1600,900,85,"jpg"}`,
   `["logo"]={400,400,90,"png"}` at `:364`) → **`logo.jpg`**, hero-encoded,
   commit subject `"Deploy hero image for <domain>"` (built from the resolved
   purpose at `deploy_image_asset_action.go:579`).

`[MEASURED]` for lendzy.co.uk (item complete 08-02, `logo.jpg` committed 08-02 by
"Deploy hero image") and webdesign.uk (08-04, same shape). `[INFERRED]` for the
other nine `logo.jpg` sites (gamesdesign 06-06 → oufe/webdesign.co.uk 07-25 —
census in `bugs_open/231`'s 08-09 contribution): same artefact signature, but
their work-item rows predate what I joined; a different vintage of the same
static may have produced them. Separate shard, same family: four `assets` rows
named literally **`input-data.asset-key.jpg`** (02-04 → 07-10) — an asset_key
config value that was a dot-path STRING never resolved, again with hero's
extension.

## Why it stayed hidden

- The branch is correct half the time by construction — the canonical-index-hero
  case wants "hero", so the static looks right in review and works in tests that
  exercise that arm.
- The deploy succeeds loudly and greenly: `success:true`, a real commit, a real
  file. The only tells are the extension, the dimensions and the commit subject
  — all of which look like a hero deploy, which is what everyone greps AROUND
  when hunting a logo problem (`bugs_open/210`'s needs_logo file read
  fundamentallyai's `logo.jpg` as the site's normal logo — corrected 08-09).
- JPEG-with-no-alpha at hero size *renders*; nothing 404s.

## Fix candidates, ordered by what closes the door

1. **Read the spec's own answer**: `store_imagery_brand_asset` drops the static
   and maps purpose from the item — `purpose_field: "input_data.spec.purpose"`,
   exactly as its sibling `store_imagery_asset` already does. One config key,
   live on apply, correct for BOTH arms of rule b (the hero item's spec says
   "hero"). Guard: `store_asset` must fail loudly if `spec.purpose` is absent
   rather than default anywhere.
2. **Then re-make the 11 artefacts**: re-drive each site's logo through the
   fixed branch (or the `needs_logo`/direct asset-deployer path with explicit
   purpose) so `logo.png` 400×400 replaces `logo.jpg`, and remove the stale
   `logo.jpg` (the deploy writes the derived path only; nothing deletes the old
   wrong-named file — check pages for references to `logo.jpg` BEFORE removing,
   robot-hands' pages reference `/assets/images/logo.jpg` today).
3. **Detect the class**: 231's candidate 3 (CheckConfig flags a static value for
   a spec-defaulted field) does NOT catch this — purpose is not shadowed here,
   it is wrong. What catches THIS class: a check that a `store_asset` step whose
   branch is selected by a spec field does not hard-code the field the spec
   carries. Harder to state mechanically; at minimum a landmine entry
   (added 2026-08-09) so a config author touching the branch reads the trap.

## How to verify a fix

Fire one `needs_imagery` item with `brand_update:true, asset_key:"logo",
purpose:"logo"` at a sacrificial domain (cookly.uk exists and is disposable);
assert the committed artefact is `logo.png`, 400×400 PNG, commit subject
"Deploy **logo** image", and `assets.purpose='logo'` on the row the deploy
stamped. The artefact properties are the disconfirmable part — a wrong purpose
cannot produce them.

## Related

- `bugs_open/231` — the neighbouring mechanism (spec-Default shadowing), whose
  08-09 contribution holds the fleet census of all 15 logo artefacts. This file
  answers its `[UNVERIFIED]` producer question.
- `bugs_open/209` / `bugs_closed/155` — the asset-source-identity thread; the
  deploy path/extension derivation this rides through is IMG-067.
- `bugs_open/210` (needs_logo file) — its §6 read the artefact as innocent;
  corrected by this lane 08-09.
- `bugs_open/223` — `code_symbols` var-blindness, third consumer (the 090 could
  not fetch `ImagePurposes`).

---

## FIXED AT SOURCE 2026-08-09 — migration 360 applied, live immediately (config, no roll)

Fix candidate 1 taken. `store_imagery_brand_asset` no longer carries a static
`purpose`; it reads `purpose_field: "input_data.spec.purpose"`, matching its own
siblings `store_imagery_asset` / `store_variant_asset`.

**Both keys had to move together, and this is the part worth remembering:**
StoreAssetAction resolves purpose literal-first
(`v3_site_actions.go:2662-2670`) — `config["purpose"]` wins, and
`config["purpose_field"]` is consulted only when it is empty. Adding the field
without deleting the static would have changed nothing while looking like a fix.

`[MEASURED]` before applying, the safety question this file raised — **do the
hero-arm items carry `spec.purpose`?** Yes, all of them:

| brand_update | asset_key | spec.purpose | items | window |
|---|---|---|---|---|
| true | `logo` | `logo` | 4 | 08-02 → 08-08 |
| true | `hero_home` | `hero` | 4 | 08-02 → 08-09 |

No `brand_update=true` item is missing `spec.purpose`, so the hero arm keeps
resolving `"hero"` and is behaviourally unchanged; only the logo arm moves, which
is the defect. If the key were ever absent the purpose resolves to `""` and the
asset row gets a NULL purpose — visibly wrong rather than silently mislabelled,
which is the failure mode to prefer.

`[MEASURED]` the verify block was **induced against the unmigrated row first**
(raised "0 of 1") before the apply raised nothing and logged "1 of 1"; live row
then re-read BY CONTENT, showing `purpose` absent and `purpose_field` set on the
brand step while the two legacy single-purpose steps (`store_hero_asset`,
`store_logo_asset`) correctly keep their statics.

### STILL OPEN — the 11 published artefacts are NOT repaired

This migration stops new ones. It does not touch the JPEGs already served by
gamesdesign, idea.uk, vonc, dartsonline, robot-hands, vetcomparison,
fundamentallyai, oufe, webdesign.co.uk, lendzy, webdesign.uk. Fix candidate 2 is
still owed, and note two traps in it: a deploy writes the DERIVED path only, so
the stale `logo.jpg` is not removed by re-running anything, and **pages currently
reference it** (robot-hands' HTML points at `/assets/images/logo.jpg`) — so the
order is re-deploy the correct `logo.png`, re-render the pages that reference the
old name, then delete the stale file.
