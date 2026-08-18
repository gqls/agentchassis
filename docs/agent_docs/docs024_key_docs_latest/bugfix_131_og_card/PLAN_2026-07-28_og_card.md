# PLAN — 2026-07-28 — bugs_open/131 (og-card slug): every page advertises a social preview that 404s

Case file: `bugs_open/131_HANDOFF_2026-07-28_og_image_points_at_a_card_that_was_never_generated.md`.
**NB `131` is one of the documented ambiguous numbers** — the other 131 is the vonc gauntlet
usability audit, owned elsewhere. Resolve by slug; `git log` the file path, never the number.

Grew out of the relojistas / fleet-discoverability lane
(`traffic_probe/HANDOFF_2026-07-28_continue_here.md` §4 item 2).

## The defect, restated

`render_site_components_action.go:448` writes `og:image` (and `:451` `twitter:image`)
unconditionally, pointing at `/assets/images/og-card.png`. On 11 of 14 live sites that file
does not exist, so every share of every page on those sites renders with no preview at all.

## What measuring changed — the fix ORDER in the case file is wrong for this estate

The case file ranks fix 1 (suppress the tag unless the asset exists) ahead of fix 2 (generate
the card), on "what closes the door" grounds. Measured against the live system on 2026-07-28,
that ordering costs more and delivers less:

| | fix 1 — suppress the tag | fix 2 — generate the card |
|---|---|---|
| outcome | *no* preview | a *working* preview |
| code change | Go, `render_site_components_action.go` | none |
| needs council + build + roll | yes | no |
| needs chrome re-render on 14 sites | **yes** — head is a stored artefact (`bugs_open/117`) | no |
| needs page redeploy | yes | only the asset commit |
| available today | yes | **yes — all 14 live sites have an active `logo`** |

The decisive measurement: **every live site already has an active `logo` asset**, which is
`derive_brand_head_assets`'s only precondition. So fix 2 is available fleet-wide *right now*
and needs no code at all — the tag already points at the right path; the file is simply absent.

**Both still belong.** Fix 1 remains the structural guard for a future site whose logo is
missing or whose derivation failed. It is second, not first.

## Decisions

1. **Fix 2 first, piloted.** Queue `needs_brand_head_assets` for one site (relojistas — this
   lane's own), verify the card on the wire end-to-end, then decide on the remaining 11.
2. **Do NOT re-derive leopardessconsulting.co.uk.** Its card serves 200 and was hand-made from
   an owner-approved logo (`docs/leopardessconsulting/RUNBOOK.md` H4, resolved 2026-07-10).
   Re-deriving would overwrite an approved brand artefact. Its missing `og_card` provenance row
   is a bookkeeping gap, to be backfilled, not a reason to regenerate.
3. **The gate for fix 1 cannot be "an `assets` row exists".** That is the obvious design and it
   is wrong: leopardess has a working card and *no* row, so a row-gate would regress the one
   site whose preview works. Whatever fix 1 keys on must not have that false negative.
4. **Fix 1 follows the sprites.css precedent** already in the same function — Phase I2 gates
   `<link rel="stylesheet" href="/assets/css/sprites.css">` on an active sprite-sheet asset
   "otherwise the `<link>` would 404 on sites without one" (`:704-712`). The og-card case is
   the same question, and the comment at `:701` explicitly waved it away as "harmless if they
   404 until derivation runs". It is not harmless; that is this bug.

## Phasing

- **P1 (done first)** — pilot fix 2 on relojistas; verify card 200 on the wire.
- **P2** — roll fix 2 to the remaining 11 (owner check first: it puts a generated image on 11
  live sites).
- **P3** — fix 1, the code gate, through the council. Separate commit.
- **P4** — backfill leopardess's `og_card` provenance row.
- **Out of scope, filed not fixed** — the `og:title` / `og:description` fallback (case file
  "second defect"); and the `undeployed_asset` detector blind spot found below.

## Decisions added 2026-07-29 (owner, session relojistas-5)

5. **Storage writes go "through the chassis"** — no credentials into any session. Implemented
   for exact bytes as an in-cluster Job (`secretKeyRef` on `personae-storage-secrets`,
   database-backup cronjob pattern); for new images, the adapter pipeline is already
   chassis-native. RUNBOOK has both recipes.
6. **relojistas' corrected crop applies EVERYWHERE** — S3 master (done, row locked), header
   (deploy repo done; edge pending origin sync), card + favicon (pending the 1199 roll).
7. **gaswholesalers + idea.uk belong to the relojistas-4 session** (owner reassignment,
   mid-session). Split + generated candidate logos + landmines:
   `COORDINATION_2026-07-29_who_does_what.md`. This lane no longer touches either site.
8. **Decision 2's mechanism was resized by reading the code:** a locked `og_card` row does NOT
   protect leopardess's card on the live binary — the derive's git commit precedes the lock
   check (only the provenance upsert honours it). So the protection is now two-part: locked
   rows backfilled (P4 done, 2026-07-29) **plus** commit `e9e345464`, which checks locks
   BEFORE compositing/committing. Until that rolls, the malformed logo row remains the real
   guard — leave it alone.
9. **P3 is resized**: the council round in flight (`bfd73f71-…`) covers the favicon
   aspect-distortion fix and the lock guard — NOT the tag gate (fix 1), which stays future
   work with decision 3's constraint intact.

## Found while measuring — a second, separate defect

The `undeployed_asset` detector has fired 5 times for `og_card`, **every one of them on
robot-hands.com — the one site whose card serves 200.** It has never fired for any of the 11
sites that genuinely lack a card, because its denominator is the `assets` table and those sites
have no `og_card` row at all. A detector that can only see assets that were *generated* is
structurally blind to the ones that were *never generated* — which is the actual failure here.
Not fixed in this lane; see NOTES for the evidence.

---

## PHASE 2026-08-18 — the mis-route: items filed without the mode that routes them (new session, picking the lane up after 14 quiet days)

The 2026-08-17 contribution in the case file (from the `idea_uk_vm_site` lane) found the
remaining live defect and explicitly did NOT fix it at source. This phase is that source fix.

**The mechanism, each link verified first-hand today (code read + live DB + live wire):**

1. **Producer** — `check_undeployed_assets.go:182-188` builds the item spec with
   `purpose` but **no `mode`**. Identified by source read; corroborated by the item census
   (`created_by IN ('design-discovery-agent','generic')`, `source='discovery'`, every spec
   carrying `purpose` and none carrying `mode`). It is the ONLY code producer of
   `needs_brand_head_assets` (grep: one `ItemType:` site).
2. **Router** — asset-deployer's live chain (read from `agent_definitions`, not the seed):
   `check_mode → check_sprite_mode → check_card_mode → check_ingest_mode → deploy_asset`;
   every conditional keys on `spec.mode`/`mode` only, so a mode-less item falls through to
   `deploy_asset`.
3. **Guard** — `deploy_image_asset_action.go:229-242` REFUSES brand-head purposes
   (correct, deliberate, council-reviewed) and returns refusal-as-result: workflow completes.
4. **Completion** — `needs_brand_head_assets` has **no registered verifier**
   (`RegisterVerifier` grep: absent), so `verifyBeforeComplete` is a no-op for it and the
   item stamps `complete`. 22 items `complete` today; the artefacts of 4 sites 404 on the wire.
5. **Two-strike knock-on** — `load_work_item_actions.go:1336-1376` counts those false
   completes as strikes (7-day window), so re-detections are born `unresolved`, never
   dispatched (webdesign.co.uk ×4, measured).
6. **Control** — idea.uk: item filed BY HAND with `spec.mode='brand_head'` on 08-17 →
   `check_mode` matched → deriver ran → both artefacts serve 200 today. Routing and
   generator work when the mode is present.

**Diagnosis-loop note (per the 2026-07-31 owner ruling):** not re-run through `090` — the
case is already filed in `bugs_open/131` with the 08-17 contribution's measurements, and this
session substituted equivalent first-hand verification: every link above was re-read in
current code / live config / live rows today, with a positive control (idea.uk) and a live
negative (3 sites 404). Stated here rather than silently skipped.

### The fixes (all three; ordered by what closes the door)

- **A — producer emits the routing key** (`Go`): add `"mode": "brand_head"` to the spec at
  `check_undeployed_assets.go:182`. Closes the door for all future discovery-filed items.
- **B — a completion verifier for `needs_brand_head_assets`** (`Go`): re-runs the deriver's
  fixed point (an active `assets` row whose `url` IS the published path — exactly
  `findBrandHeadAssetGaps`' `rowAtPublishedPath` arm, and exactly what `recordDerivedAsset`
  writes on success, so detector predicate and handler remit COINCIDE — the 016b §9
  "verify the handler's remit" trap does not bite). Grades declared (bugs_open/213):
  speaks for specs naming a brand-head `purpose` or `mode='brand_head'`.
  This REVERSES a recorded decision (verifier_coverage_test.go: "stays catCreation") —
  corrected visibly there: existence is weak proof of QUALITY, but the live failure was
  ABSENCE stamped complete, which an existence verifier catches; the eyeball stays.
  Lockstep: add the type to the claim-timeout exclusion in `220_…_generic_evidence.sql`
  (TestRegisteredVerifiersMatchClaimTimeoutExclusion enforces both directions).
- **C — routing fallback, LAST in the chain** (config, migration 467): new conditional
  `check_brand_head_purpose` between `check_ingest_mode` and `deploy_asset`:
  brand-head `spec.purpose` → `derive_head_assets`. Placed after every mode check so an
  explicit mode always wins and the fallback can only catch items that would otherwise
  reach a branch that unconditionally refuses them — it converts a guaranteed refusal into
  the action the refusal's own text prescribes. NOT a widening of `check_mode`: that would
  hijack ingest/sprite/card-mode items that happen to carry a brand-head `purpose` field.

**Interactions:** A and B are Go (inert until a chassis build rolls); C is config (live on
apply). No ordering constraint — C alone heals the live population on redrive; A alone fixes
future discovery items; B makes any residual mis-route visible instead of silent. Hazard
window accepted and stated: between applying the 220 exclusion (live at apply) and the roll
that registers the verifier, stuck CLAIMS of this type fall to the 40-minute reset instead of
auto-completing — near-zero impact, self-heals at the roll.

**Redrive (after C):** webdesign.co.uk, cookly.uk, lendzy.co.uk get the proven manual item
shape (idea.uk's). **loancalculator.co.uk is deliberately left to its owning lane** (active
today, transcript-verified; its own queue plan already names its 2 deferred brand-head items).
webdesign.uk 302s to webdesign.co.uk — one redrive covers both.

**Deliberately NOT done:** no HTTP probe in the verifier (house rule — see image_url_404's
entry in the coverage map); no change to the refusal-as-result semantics in
`deploy_image_asset` (correct and council-reviewed; the verifier is the framework's designed
place to catch a refusal that changed nothing).
