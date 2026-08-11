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

---

## 2026-08-10 — fix candidate 1 PROVEN LIVE, candidate 2 IN FLIGHT, and the site list was wrong

### Candidate 1 (migration 360) is behaviourally proven, not merely applied

One `needs_imagery` item at cookly.uk (`brand_update:true, asset_key:"logo",
purpose:"logo"`, item `3c1c7c65-…`). **The routing is the disconfirmable part**,
and the run recorded it in its own `collected_data`:

```
check_imagery_brand_update = {"condition_met": true,
                              "next_step_override": "store_imagery_brand_asset"}
asset_stored               = {"purpose": "logo",
                              "logo_url": "/assets/images/logo.png", …}
```

Pre-360 that identical input stored the logo with `purpose:"hero"`. Served
artefact: `logo.png`, PNG, 400×218, 189,044 B, sha `3fb6ad54…` (replacing
`e38781c2…`); **no `logo.jpg` created**. This is the **first time the brand
branch has ever produced a correct logo** — cookly's previous, correct logo came
from the LEGACY `needs_logo` path (`store_logo_asset`), a different step, which
is exactly why this had to be fired rather than reasoned about.

> **⚠ TWO CORRECTIONS TO THIS FILE'S OWN "How to verify a fix" SECTION.** Both
> would fail a CORRECT run, so fix them before anyone else tests against them.
>
> 1. **Not "400×400 PNG".** Logo processing fits the image inside a 400px box
>    preserving aspect ratio, so a wide wordmark comes out **400×218** — as does
>    idea.uk's known-good logo. The disconfirmable properties are **PNG not
>    JPEG**, **≤400px not 900×900 / 1408×768**, and **`purpose='logo'` on the
>    stamped row**.
> 2. **No new `assets` row is created.** `store_asset` UPDATED row `5c351ebc`
>    in place (its `storage_path` moved to a `20260810/` key). "Assert the row
>    the deploy stamped" is right; "assert the NEW row" finds nothing.

### The affected-site list was wrong in two places `[MEASURED]`

Cross-checked three ways — `assets.purpose` for `asset_key='logo'`, what each
domain actually serves, and what its homepage HTML references:

- **`idea.uk` is NOT affected** and must come off the list: row `purpose='logo'`,
  serves `logo.png` 400×218, homepage references the PNG. It has only a stale,
  **unreferenced** `logo.jpg`.
- **`relojistas.com` IS affected and was missing**: row `purpose='hero'`, serves
  `logo.jpg` (JPEG 646×275), homepage references it.

Corrected picture: **9 sites with visitor-visible damage** (gamesdesign.co.uk,
vonc.com, dartsonline.com, robot-hands.com, vetcomparison.uk,
fundamentallyai.com, oufe.com, lendzy.co.uk, relojistas.com), plus
**webdesign.co.uk** (bad row, serves the JPEG, but its homepage references no
logo at all) and **webdesign.uk** (bad row, 302s so unverifiable over HTTP).

Two dimension families among the damaged: 900×900 and 1408×768 — worth a glance
if anyone wants to date when the deploy's hero processing changed.

### Candidate 2 — 9 items filed 2026-08-10 12:00Z, dispatching now

`created_by='bugfix-235-logo-repair'`, one per standard-deploy affected site,
each spec built **from that site's own current `site_plan_imagery` logo row**
(the framework authored those prompts — I did not write nine new ones), mirroring
`imageryplan.BuildSpec`. Filed at `status='triaged'` because
`build-pipeline-trigger`'s pre_query requires exactly that plus `pipeline='build'`
— **filing at the discovery default `detected` goes nowhere**, which is worth
knowing before anyone files these by hand again.

**TWO SITES DELIBERATELY NOT FIRED**, both needing a decision rather than a retry:

1. **`relojistas.com` and `webdesign.uk` are `github_repo='vm-sites'`** — a
   different deploy path from the bucket-served sites. Re-driving the logo would
   regenerate and store the asset correctly, but I have not verified the deploy
   half reaches a VM-served site, and firing at an unverified deploy path is how
   you get a green work item and an unchanged page. Route via whoever owns the
   VM-sites lane. (`idea.uk` is also `vm-sites`, and is fine, so being VM-served
   is not itself the defect.)
2. **`webdesign.uk` is additionally BLOCKED from all dispatch.** It holds work
   item `8793da9a` inserted directly as `status='claimed'` with NULL
   `claimed_at`/`claimed_by`. The selector skips any site with a claimed item, so
   **nothing on webdesign.uk can ever dispatch until that row is resolved.** Not
   this lane's row to clear — but it must be cleared first, and it is silently
   starving that site of every kind of build work, not just this one.

### Still owed after the artefacts are re-made

The order in this file's original candidate 2 still holds and is not yet done:
re-deploy the correct `logo.png` → **re-render the pages that reference
`logo.jpg`** (9 homepages do) → only then delete the stale file. A deploy writes
the derived path only; nothing removes the old object, and the pages keep
pointing at it until they are re-rendered.

### RESULT 2026-08-10 12:25Z — 8 of 9 repaired and verified at the artefact; 3 sites remain, each for a DIFFERENT reason

`[MEASURED]` `assets.purpose` for `asset_key='logo'` across the affected set:
**8 now `logo`** (dartsonline, fundamentallyai, gamesdesign, lendzy, oufe,
vetcomparison, vonc, webdesign.co.uk) · **3 still `hero`** (relojistas.com,
robot-hands.com, webdesign.uk).

Verified at the served artefact, not at the work-item status:

| site | logo.png now | stale logo.jpg |
|---|---|---|
| gamesdesign.co.uk | PNG 400×400 | JPEG 900×900 |
| vonc.com · dartsonline.com | PNG 400×218 | JPEG 900×900 |
| vetcomparison.uk · fundamentallyai.com · oufe.com · lendzy.co.uk · webdesign.co.uk | PNG 400×218 | JPEG 1408×768 |
| robot-hands.com | **404 — not repaired** | JPEG 1408×768 |

Note gamesdesign came out **400×400** and the rest **400×218**: further evidence
the rule is "fits a 400px box, aspect preserved", not a fixed size.

#### robot-hands.com — FAILED SAFELY on a permanent approval lock. Do not retry it.

`assets.locked_at = 2026-07-11`, `locked_by = user-b6-approval`,
`lock_type = permanent`. `store_asset` refused, returning
`{stored:false, locked:true, reason:…}` with no `image_uri`, and the next step
then died on `source path 'asset_stored.image_uri' not found for field 's3_uri'`.

**The lock did its job.** But note the failure surfaces as a missing-path error
three steps downstream, not as "asset is locked" — so the next person to hit this
will debug an input_mapping bug that is not there. Worth a cheap improvement:
`spawn_asset_deployer`/`call_asset_deployer` should branch on
`asset_stored.stored == false` rather than fail on the absent field.

**Why a retry is the WRONG move here, and this is the substantive point:** the
locked asset *is* the defective one, but the lock exists because someone approved
the **artwork**. Re-driving it through the generator produces a *different image*
and throws the approved one away. The defect on robot-hands is how the image was
processed and deployed (JPEG 1408×768 instead of a small PNG), **not what it
depicts.** The correct repair is to re-deploy the EXISTING source object at
`purpose='logo'` — preserving the artwork, fixing the encoding — which is a
different operation from the one the other eight ran. **Owner call.**

#### relojistas.com — locked too, and VM-served

`locked_by = 'owner via relojistas-5 session (bugs_open/131)'`, 2026-07-29. It was
excluded from this run for being `github_repo='vm-sites'`, so the lock was never
tested — but it would have refused for the same reason. Same owner call as
robot-hands, plus the VM deploy-path question.

`[MEASURED]` **no other affected site carries a lock**, so nothing owner-approved
was overwritten by this run. That was checked, not assumed.

#### webdesign.uk — unlocked, but cannot dispatch at all

Still `purpose='hero'`. Blocked by work item `8793da9a`, inserted directly as
`status='claimed'` (NULL `claimed_at`/`claimed_by`); the selector skips any site
with a claimed item. Clear that row and this site can be repaired like the other
eight — subject to the same `vm-sites` deploy-path question.

#### Still owed — the pages have NOT changed yet

All 9 homepages still reference `/assets/images/logo.jpg`, so **a visitor sees no
difference yet.** The deploy DOES update `sites.content_data.logo_url` to
`/assets/images/logo.png` (verified on gamesdesign and cookly), so a re-render
will pick the new file up. Remaining, in order: re-render → then delete the stale
`logo.jpg`. Prepared statement and its safety reasoning are in the lane's
scratch SQL; the site-level `needs_rerender` is the right entry point because
`get_pages_for_rerender` is configured `include_statuses = [deployed, active]`
and so **cannot** disturb the 26 `needs_rebuild` / 17 `planned` pages that
predate this work.

### 2026-08-10 16:33Z — re-render drain COMPLETE; the 8 sites are DONE end to end; deletion step re-scoped

`[MEASURED]` all **255** fanned-out `page_rerender` items complete, **0 failed**
(survived the v1.0.1279 fleet roll mid-drain). Verified at the served HTML, not
the item status: gamesdesign, vonc, dartsonline, vetcomparison, fundamentallyai,
oufe, lendzy all reference `/assets/images/logo.png`; webdesign.co.uk references
no logo on its homepage (as before — its row is fixed, impact was always
unclear). **For these 8 sites the defect is fully remediated for visitors.**

#### ⚠ The stale-file deletion step is WRONG as written in this file — re-scoped

This file's candidate 2 ordering ("… then delete the stale file") assumed a
`logo.jpg`'s readers are its own site's pages. **Measured otherwise:**
`fundamentallyai.com`'s served portfolio hot-links OTHER domains' logos directly —
`https://idea.uk/assets/images/logo.jpg` and
`https://relojistas.com/assets/images/logo.jpg`. So `idea.uk`'s `logo.jpg`,
which the 08-10 handoff called "stale, unreferenced" (true on idea.uk itself),
**has a reader on another domain**, and deleting it breaks fundamentallyai's
page.

Deletion therefore needs an **estate-wide** served-HTML reference audit (every
live domain, absolute + relative URL forms), not a per-site one — and only URLs
with zero references anywhere may go. The files cost nothing meanwhile; this
step is optional and LAST. Do not delete on the strength of this file's earlier
wording.

#### B1 feasibility note (robot-hands)

Its locked asset's `storage_path` is a plain public HTTPS URL, **no query
string** (not presigned): `…backblazeb2.com/personae-prod-uk001-images/images/
demo_client/20260509/d321c4f2-….png` → converts mechanically to
`s3://personae-prod-uk001-images/images/demo_client/20260509/d321c4f2-….png`.
So the re-deploy-existing-source repair is feasible via `deploy_image_asset`
with explicit `spec.s3_uri`. Unknown: whether the deploy's stamping step also
refuses on the lock — run and read, do not assume.

### 2026-08-10 ~16:45Z — owner decisions taken; robot-hands lock OVERRIDDEN on owner instruction

Owner answered the four open decisions in-session:

1. **robot-hands.com: REGENERATE, overriding the `user-b6-approval` lock** —
   owner chose this over re-deploying the existing source, accepting that the
   approved artwork is replaced. Lock state before override, for the record:
   `locked_at=2026-07-11 16:38:11.647284+00, locked_by=user-b6-approval,
   lock_type=permanent`, asset id `75164b4c-a417-430b-81d2-4d1e85578a33`.
   **The approved artwork itself is NOT destroyed** — its S3 object remains at
   `images/demo_client/20260509/d321c4f2-edc5-4770-833a-3a2563b420ba.png`
   (an asset-row update never deletes the object). Restorable from there.
2. **webdesign.uk: malformed `claimed` row `8793da9a` RESET to `triaged`** —
   done; the site can dispatch again.
3. **240: owner chose C2 (MetadataTopics fix) + C3 (GOMEMLIMIT/headroom) + C4
   (scheduled sweep stopgap); C1 (per-topic reaper) deliberately NOT taken.**
4. **relojistas.com: attempt in this session**, including verifying the
   `vm-sites` deploy path first.

### 2026-08-11 ~10:05Z — 11 of 11 DONE: relojistas repaired through migration 380, proven at the served artefact

The last site is fixed, and by the class fix rather than a workaround. The
shape mismatch that beat both correct values on 2026-08-10 (item spec AND
asset row `purpose='logo'`, deploy still "hero") is closed by **migration
380**: build-dispatch-loop's `call_handler.input_mapping` gains
`"purpose?": "current_item.spec.purpose"`, the same idiom
site-work-orchestrator's `fix_items_loop` already carried. Full mechanism and
the refutation of the earlier `purpose_field` candidate: `bugs_open/231`
(2026-08-11 contribution) and NOTES_209.

Proof at the artefact, not the status: item `6084d849` reset to `triaged`,
dispatched 10:00Z, committed **"Deploy logo image for relojistas.com"**
(both pre-fix runs said "Deploy hero image"), asset row restamped
`/assets/images/logo.png`, and the served object is
`https://relojistas.com/assets/images/logo.png` → 200, `image/png`,
**PNG 400×170 RGBA** (alpha restored), last-modified 10:00:50Z.

Follow-ons executed the same hour:
- relojistas site-level `needs_rerender` (`07051741`, refresh_site_components)
  completed 10:05Z and fanned out its page_rerender items — homepage reference
  flip to be verified at the served page once the fan-out drains.
- fundamentallyai's portfolio hot-link: the `.jpg` URL lived in ONE
  `page_components` row (index / portfolio-showcase, `28348227`), in BOTH
  `content_data` and `rendered_html` — a rerender alone would have reproduced
  it. content_data patched `.jpg`→`.png` (backup:
  `page_components_bak_20260811_fundai_logolink`), then `page_rerender`
  `ffe2bd7e` filed for its index.
- The old `relojistas.com/assets/images/logo.jpg` is deliberately NOT deleted
  — it still serves, and deletion waits on this file's estate-wide reference
  audit (still open).

**What keeps this file OPEN:** the estate-wide logo.jpg reference audit before
any deletion, and verification of the two rerenders at the served pages.

### 2026-08-11 ~11:45Z — audit DONE, rerenders VERIFIED at the served pages; one site deferred to its owning lane; deletion now needs only the owner's word

Estate-wide audit of renderable `/assets/images/logo.jpg` references (queries
in RUNBOOK_209): **three sites only**, all now handled —

- **relojistas.com**: chrome flipped by the site rerender; served homepage
  verified `2× logo.png, 0× logo.jpg`.
- **robot-hands.com**: chrome (head+header) still carried `.jpg` — its own
  detector's `needs_rerender` (render_inputs_drift, `add3a661`) promoted to
  triaged; chrome flipped 10:23Z; served homepage verified `2× logo.png, 0×
  logo.jpg`; interior pages draining.
- **webdesign.uk**: same — its `missing_structure` item (`58f88922`)
  promoted; chrome flipped 10:25Z, 5 pages complete. Deliberate 302 — rows
  are the evidence, both slots `png only`.
- **fundamentallyai.com**: BOTH hot-links (relojistas + idea.uk, one
  portfolio component) patched in content_data AND rendered_html; but the
  site's index is under **active rebuild by the brochure_component_library
  lane** (`needs_page:index:151census:20260811`, filed 10:27Z) — my assembly
  item races their full rebuild benignly (both sources patched, and the
  stale `.jpg` still serves at its origins). Served-page verification
  deferred until their rebuild lands; do not fight over the page.
- content_components templates: 0 references. No other site references any
  logo.jpg, own or cross-site.

**Deletion of the stale logo.jpg files is now gated only on (a) the owner's
say-so and (b) fundamentallyai's index actually serving `.png` links after
the brochure-lane rebuild.** Until both: delete nothing.

> **SUPERSEDED same hour:** fundamentallyai's verification is no longer
> deferred. The assembly item (`ffe2bd7e`) completed 11:21:36Z, the page
> redeployed 11:21:48Z, and the served index now shows ALL THREE portfolio
> logos as `.png` (relojistas, idea.uk, leopardess). Every affected site is
> now verified at the served artefact. Deletion is gated on the owner's word
> alone — though note the brochure lane's `needs_page:index` rebuild is still
> queued and will regenerate this page again from the (patched) content_data.

> **ADDENDUM 2026-08-11 ~11:45Z (from the brochure_component_library lane, whose rebuild
> your last note anticipated):** the deferred check has an answer and it is the one that
> BLOCKS deletion. My `needs_page:index` full rebuild deployed 11:26:58Z (content
> regeneration, not assembly) and the served index now references
> `https://relojistas.com/assets/images/logo.jpg` and `https://idea.uk/assets/images/logo.jpg`
> again — only leopardess kept `.png`. Your content_data patch did not survive REGENERATION:
> the portfolio card fields are re-resolved from their source on a full rebuild, so a patch
> at content_data level holds through assembly/rerender but dies at the first `needs_page`
> (the bugfix-238 save-vs-rerender family, this time on the resolver path). All referenced
> URLs serve 200 today, so nothing is visibly broken — but **do not delete the origin
> `.jpg` files** while fundamentallyai's served index references two of them, and your (b)
> gate condition is NOT met. The durable fix needs to land where the resolver reads the
> logo URL from (the portfolio source), not in the resolved copy. fundamentallyai's own
> `/assets/images/logo.png` serves 200/157KB, so the blocked `image_url_404:logo.png`
> item's "no active asset" premise looks stale as well — worth re-verifying before anyone
> acts on it.
