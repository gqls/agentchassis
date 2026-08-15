# 231 — a STATIC step-config value for a spec-DEFAULTED field is silently dead: `ExtractActionInputs` applies Defaults first and every later strategy skips a populated field

**Filed:** 2026-08-09 by the `bugfix_209_deploy_purpose_keyed_source` lane, discovered
while costing the "bring pageflow-builder / site-work-orchestrator into line" option
for `bugs_open/209` at the owner's request.
**Status:** OPEN, owned by `bugfix_209_deploy_purpose_keyed_source` for the
deploy_image_asset instance; the fleet-wide class is `[UNMEASURED]` and has been
handed to the diagnosis loop (090 fired 2026-08-09, correlation in NOTES).
**Severity:** medium-high for the named instance (wrong deploy path + wrong resize
class, silent); unknown for the class.

## Method note (per the 2026-07-31 owner ruling)

090 was **not** run before filing the *instance* claim; substituted first-hand
verification, stated plainly: the mechanism was **executed, not inferred** — the real
`ExtractActionInputs` + the real `DeployImageAssetInputSpec` + a replica of the
action's own fallback, in committed characterisation tests that pin the behaviour
(`deploy_image_asset_purpose_source_test.go`:
`TestLegacyLogoStep_StaticPurposeIsShadowedByDefault`,
`TestPurposeFieldBridge_DeadForDefaultedField`,
`TestStrategy0DottedPaths_DefeatTheDefaultAndTheRecursiveSearch`), plus a read of the
deciding arms (`action_inputs.go:457-460` Defaults-first; `:499`/`:511`/`:523` the
three already-resolved skips; `deploy_image_asset_action.go:92-99` the unreachable
fallback; `url_helpers.go:190` the consequence). The **class** claim is exactly what
090 is for, and that half went through it rather than being asserted here.

## The mechanism

`ExtractActionInputs` (`platform/orchestration/datahelpers/action_inputs.go`):

1. `spec.Defaults` are copied into `result.Values` **first** (`:457-460`).
2. Strategy 0 resolves config values as explicit paths — but **only multi-segment
   dotted strings** (`strings.Contains(pathStr, ".")`, `:478`). A static value like
   `"logo"` is invisible to it.
3. Strategies 1, 2 and 3 each **skip any field already in `Values`**
   (`:499`, `:511`, `:523`) — which a defaulted field always is.

Net: for a spec-defaulted field, a static (non-dotted) config value can never be
read through the inputs path, and the deprecated `*_field` bridge for it is equally
inert. If the action then has a fallback of the form `if inputs.Get(f) == "" { read
config }`, that fallback is **unreachable** — the default guarantees non-empty.

## The proven instance

`pageflow-builder` and `site-work-orchestrator` (both live, owner 2026-08-09:
*"not dead, but not being worked on"*) each carry `deploy_logo_image` with static
`"purpose": "logo"`, no `input_fields`. `DeployImageAssetInputSpec` has
`Defaults{purpose: "hero"}` — since **`34d2315ce`, 2026-02-20**.

Measured (tests above): the logo step's effective purpose is **"hero"**. Consequences
if the step runs: resize by the hero dimension class, and — because the deploy path
is `BuildAssetPaths(purpose, ext)` → `filename = purpose + ext` when no `asset_key`
is supplied — the logo's bytes commit to **the hero's path**, clobbering the hero
image, while the store step's own `content_data` write promises `logo_url:
/assets/images/logo.png`, which is then never written. Silent at deploy time; the
symptom is a broken/hero-shaped logo on the site.

`[UNMEASURED]` Whether this ever fired live. The pair has no dispatcher among live
definitions and did not run on 08-08/09; completed `orchestration_states` rows are
reaped at ~24h, so history is gone. Do not assert a live occurrence; equally, do not
assume the pair "worked fine" recently — the honest state is **broken-if-run today,
at the resolver level, proven by executing the resolver**.

`[UNMEASURED]` The fleet class: ~10+ other `ActionInputSpec`s carry `Defaults`
(grep `Defaults: map` under `platform/orchestration/actions/`). Any live step config
carrying a static value ≠ the default for one of those fields is silently taking the
default. Sizing this needs the spec-by-spec field list joined against live
`agent_definitions` step configs — that is the 090's job, not this file's claim.

## Why it stayed hidden

- The shadow only *changes* behaviour where config-static ≠ default — for
  `deploy_hero_image` (static "hero" = default "hero") it is invisible.
- The affected workflows are frozen; nothing runs them, so nothing surfaces it.
- Every individual piece looks correct in review: Defaults-first is reasonable,
  skip-already-resolved is reasonable, the action's `if == ""` fallback is
  reasonable. The defect is the composition.
- `bugfix_136`'s landmine ("`Deprecated` cannot alias a renamed SETTING — it resolves
  the value as a data path, so it takes the default and silences the detector") is
  the **same family**, found from the aliasing side. This file names the general
  rule: **against a defaulted field, only a Strategy-0 dotted path can win.**

## Fix candidates, ordered by what closes the door

1. **Per-instance repair via Strategy-0 dotted paths** (config-only, live on apply):
   point the field at data that exists — for the legacy pair,
   `"purpose": "logo_stored.purpose"` etc. This is also the 209 "into line" Phase 1;
   one migration fixes both. Proven deterministic by the third test above.
2. **Make the composition honest in `ExtractActionInputs`**: let an explicit
   config-static value for a spec field **override** the default (defaults are
   *defaults*; an explicit config value is not an absence). One ordering change +
   one rule ("config-static beats Default, Strategy 0 beats both") — but it changes
   behaviour for every action at once wherever a config-static value was being
   silently ignored, i.e. it *activates* config that has been dead for months.
   Blast-radius census required first (the 090 delivers it). Council, definitely.
3. **Detect-only**: extend the offline config-key audit (`CheckConfig`) to flag a
   static config value for a defaulted spec field as "shadowed — will not take
   effect". Cheap, catches future authors; repairs nothing existing.

## How to verify a fix

For candidate 1 (the pair): apply the migration, then one sacrificial-domain run of
each workflow; assert the committed files — `hero.*` and `logo.*` both present with
**different** bytes, and `content_data.logo_url` resolving 200. For candidate 2: the
census first; then the three pinned tests flip and must be updated deliberately,
citing the change.

## Related

- `bugs_open/209` — found while costing its into-line option; Phase 1 there = fix 1 here.
- `bugfix_136_config_key_aliases` — same family, aliasing side.
- `bugs_closed/155` / `bugs_open/152` — the asset-source-identity thread this all serves.
- LANDMINES 2026-08-09 entry (footprint `ExtractActionInputs`, `ActionInputSpec`,
  `Defaults`) — the prospective trap for config authors.

---

## STATUS 2026-08-09 (same day, later) — instance REPAIRED LIVE (migration 348); 090 verdict in, recorded honestly

**Fix candidate 1 executed for the proven instance.** Migration
`348_pageflow_swo_deploy_steps_resolve_by_identity.sql` (applied + recorded
09:41:53, ROLLBACK sidecar alongside) replaced the four deploy steps' dead static
keys with Strategy-0 dotted paths (`{p}_stored.purpose/s3_uri/asset_id`,
`site_record.domain`) and added `input_fields` that **deliberately exclude
`s3_uri`** so a store failure degrades to a safe skip instead of an aggressive
search that can cross to the sibling asset. Discipline followed: post-verify is
DO/RAISE and was **induced first** (run standalone against unmigrated rows — it
raised, 0/4); dry-run in a doomed transaction; applied scoped via
`MIGRATIONS_DIR` (the runner's `--apply` takes every pending file — others'
342/345/346 were pending); verified at the rows by content, with the store steps
as a negative control (untouched). Two pre-existing config keys surfaced by the
full-row read — `domain_field` (now inert: Strategy-0 `domain` wins) and
`output_mapping` on the hero steps — both preserved by the surgical merge.
Harness extended to the **exact live shape** including the store-failure corner:
8/8 tests pass (`TestMigration348Shape_StoreFailureResolvesNoURI_NeverTheSibling`
is the door-stays-shut proof).

**The 090 on the fleet class returned UNVERIFIABLE (scope-not-narrowing)** — run
`e952039b`. Read precisely, it *strengthens* the mechanism claim and declines the
instance: the loop **independently confirmed the helper-level mechanism** (its
words: the Defaults-first + has-value-skip chain "is real and would silently drop
a static non-dotted config literal for a defaulted field") but could not cite
`DeployImageAssetInputSpec`'s `Defaults` because **its code lookup only surfaced
the spec's two USE sites, never the `var` declaration — `bugs_open/223`'s
`code_symbols` var-blindness biting the diagnosis loop's own lookup layer** (second
consumer of that gap; noted in 223). Gap 2 — no runtime row of the bug firing —
was already `[UNMEASURED]` here and stays so: the workflows do not run. The
declaration gap is answered outside the loop by the committed test that
**executes** the real spec (`TestLegacyLogoStep_StaticPurposeIsShadowedByDefault`
— its header marks the config shape as pre-348 history while noting the resolver
behaviour it pins is still current for any config authored that way).

**What keeps this file OPEN:**
- The **fleet-class census** (which other live configs carry a static value for a
  spec-defaulted field) is still undone — the 090 stopped before enumerating.
- Fix candidates 2 (make config-static beat Defaults in the helper) and 3
  (CheckConfig flags shadowed statics) are unaddressed; 3 is cheap and catches
  future authors.
- The **behavioural proof** for the repaired instance is owed: one
  sacrificial-domain run of each workflow, hero.* and logo.* both committed with
  different bytes. The bar for closing is fixed AND live AND proven at the
  artefact.

---

## CONTRIBUTION 2026-08-09 (afternoon), `bugfix_209_deploy_purpose_keyed_source` lane — this fired LIVE on 11 sites, and 348 did not fix that

Two of this file's own statements are corrected below. Full evidence, method and
missteps: `docs024_key_docs_latest/bugfix_209_deploy_purpose_keyed_source/NOTES_209…`
(2026-08-09 afternoon section).

### 1. `[MEASURED]` The behavioural proof for the repaired instance: DONE, and it passes

Sacrificial domain `cookly.uk`, full `pageflow-builder` build (correlation
`0562a667-…`, orchestration `22fb157a-…`). Both assets committed —
`cookly.uk/assets/images/hero.jpg` and `…/logo.png` — and the logo is a **400×400
PNG**, the first correct one this system has produced since 2026-03-02. The hero
deploy step (the only one running with BOTH assets in `collected_data`) resolved
every field through Strategy-0 dotted paths, downloaded the hero's OWN storage
object and stamped the hero's OWN asset row, while the aggressive search ran in the
same step and had its result discarded.

> **The verification bar written in this file is not disconfirmable — do not rely on
> it.** "hero.* and logo.* committed with different bytes" cannot fail: the deploy
> re-encodes per purpose (hero→jpg, logo→png), so the outputs differ in bytes even
> when the wrong source is fetched. Use the **downloaded object key** and the
> **asset row stamped** instead; both are in the NOTES with timestamps.

### 2. `[MEASURED]` "Whether this ever fired live" — IT DID. 11 sites are serving it now

This file records that as `[UNMEASURED]`. Census of every logo committed to
`gqls/sites`, joined to its producing commit subject:

- **Correct — `logo.png`, 400×400 PNG, subject "Deploy _logo_ image":** 4
  (ai-agent-orchestration 02-22, finetuning 03-02, leopardess 07-10 hand-made,
  cookly 08-09 = the post-348 run above).
- **Wrong — `logo.jpg`, JPEG at 1408×768 / 900×900 / 646×275, subject "Deploy
  _hero_ image":** 11 — gamesdesign 06-06, idea.uk 06-21, vonc 06-23, dartsonline
  07-06, robot-hands 07-10, vetcomparison 07-17, fundamentallyai 07-21, oufe and
  webdesign.co.uk 07-25, relojistas 07-29 (since hand-replaced), lendzy 08-02,
  webdesign.uk 08-04.

`purpose == "hero"` at those deploys is established three independent ways:
`deploy_image_asset_action.go:579` builds the subject from the resolved `purpose`;
`DeployedAssetPath` takes the extension from `purpose` and the filename from
`asset_key` (`url_helpers.go:317-330`) and `ImagePurposes["logo"]` is
`{400,400,90,"png"}` (`:364`), so `.jpg` is unreachable with purpose "logo"; and
`DownloadOptimizeAndPrepare` resizes by `purpose` — none of the 11 is 400×400.

**Impact is not cosmetic:** JPEG has no alpha channel, so each of those logos is
served with an opaque background at up to 1408×768 instead of 400×400.

### 3. The predicted failure SHAPE was wrong, and that matters for who still owns the exposure

This file predicts the logo's bytes committing to **the hero's path**, clobbering the
hero — reasoning from "no `asset_key` is supplied". The artefacts say otherwise: the
files are named `logo.jpg`, so an `asset_key` of "logo" **was** supplied. The historic
producer is therefore **not** the legacy-pair shape this file modelled, and:

> **Migration 348 did not fix the fleet's logo problem.** It repaired the two
> workflows that nobody dispatches. The exposure that actually shipped came through
> another door.

`[UNVERIFIED]` Which caller. `assets` points at **`asset-deployer`**: four rows are
named literally `input-data.asset-key.jpg` — an unresolved `input_data.asset_key`
config path leaked in as the asset key, with hero's extension again, which is this
file's mechanism wearing a different hat (`asset-deployer` reads
`purpose: input_data.purpose`; a caller that omits it gets the `"hero"` Default and
nothing says so). Filing a 090 rather than asserting it — the cause is not where the
symptom is, which is exactly that loop's case.

### 4. What this changes about closing

The instance repair (348) is proven; **this file must not close on it.** The live
damage is 11 shipped artefacts plus whatever keeps producing them, and neither is
addressed. Candidate 3 (CheckConfig flags a static value for a defaulted field)
would have caught the producer at authoring time and is still the cheap win.

---

## CONTRIBUTION 2026-08-09 (later) — the producer question is ANSWERED, and it is NOT this file's mechanism: see `bugs_open/235`

Same lane, closing the `[UNVERIFIED]` left in the previous contribution. The 090
(run `fd7ef7a9-93fb-4e20-9956-f8913bd4ab89`) returned UNVERIFIABLE
(scope-not-narrowing) — it could not fetch `asset-deployer`'s live step config
(it read the empty `task_workflow`/`orchestrator_workflow` columns, not
`default_config`) nor the `ImagePurposes` var declaration (`bugs_open/223`'s
var-blindness, **third consumer**). Its named gaps were answered first-hand; the
full chain is filed as **`bugs_open/235`** with every link quoted.

The short version: `image-build-handler.store_imagery_brand_asset` carries a
**static `purpose: "hero"`** on the brand-update branch, whose own description
says it handles *"logo or canonical index hero"*. A `needs_imagery` item with
`brand_update:true, asset_key:"logo", purpose:"logo"` (the spec says exactly what
it is) stores the logo as purpose "hero"; `call_asset_deployer` forwards
`asset_stored.purpose`; the deploy derives `logo.jpg` + hero processing.
`[MEASURED]` by work-item ↔ commit date join for lendzy (08-02) and webdesign.uk
(08-04); `[INFERRED]` for the older nine.

**What this corrects in THIS file's frame:** the wrong logos are NOT an instance
of the Defaults-shadow — the static resolves fine; it is simply the wrong value
for one of the branch's two arms. So candidate 3 here (CheckConfig flags a
shadowed static) would NOT have caught the producer, and closing this bug must
not be read as closing the logo damage. The two doors are neighbours in the same
wall (`ExtractActionInputs` config statics), but they are different doors.

Also for the record, both proof arms are now done: the site-work-orchestrator run
(correlation `aab47560-…`, 13:50–13:56) re-made both cookly.uk assets from its own
store outputs — fresh commits `b56599fe0` ("Deploy **logo** image", 400×400 PNG,
sha `e38781c2…`) and `d47cf0315` ("Deploy **hero** image", sha `be35ba8d…`), both
byte-different from the pageflow run's artefacts. On this arm the pod-level
object-key log line was NOT captured (the capture was pinned to a dead pod and
these pods keep ~11s of logs) — the assertion rides on the artefact properties,
the asset-row stamps, and the mechanism being the identical config shape on the
same binary as the pageflow arm, where the log-level proof exists.

---

## 2026-08-10 17:05Z — a LIVE specimen, caught in the wild: `deploy_image_asset`'s `Defaults{purpose:"hero"}` beats a correct value in BOTH the spec and the asset row

While completing 235's relojistas repair, the census's quarry fired twice in one
hour, reproducibly:

- The `undeployed_asset` dispatch path (build-dispatch-loop → asset-deployer)
  delivers item spec fields at `input_data.spec.*`. The asset-deployer
  definition's `deploy_asset` step binds `"purpose": "input_data.purpose"` — a
  dotted path that resolves NOTHING on that dispatch shape.
- `DeployImageAssetInputSpec` (`deploy_image_asset_action.go:36-38`) carries
  `Defaults: {"purpose": "hero"}`. The unresolvable dotted path falls to the
  Default **before anything finds `input_data.spec.purpose`**.
- `[MEASURED]` two deploys for relojistas.com, 17:01 and 17:03: item spec
  `purpose='logo'`, and (round 2) the asset ROW also corrected to
  `purpose='logo'` first — **both runs derived and committed as "Deploy hero
  image"**. The only value that ever won was the Default.

So the class this census hunts has a second face: not just "a static config
value shadowed by a Default", but **"a dotted path that resolves nothing on one
dispatch shape falls to the Default, shadowing the correct value sitting one
level down"**. The step config LOOKS explicitly wired (`input_data.purpose`
reads like a real binding) and works on the image-build-handler path (which
maps purpose to the top level) — it is only this dispatch shape that dies, and
it dies silently into a plausible wrong artefact (a JPEG hero-derivation of the
right source image).

Candidate fix (needs its own round, it edits a shared definition): the
asset-deployer `deploy_asset` step should bind
`"purpose_field": "input_data.spec.purpose"` alongside — the spec's Deprecated
bridge maps `purpose_field`→`purpose` — or the dispatch loop should map purpose
to the top level for `undeployed_asset` items. Either way, **231 candidate 3
(CheckConfig flagging a shadowed static) would NOT catch this**: nothing here is
a static; it is an unresolvable dotted path. The census query needs a third arm:
config dotted paths that cannot resolve against their dispatch shape, on
actions whose spec carries a Default for the same field.

---

## 2026-08-11 — the dispatch-shape face is FIXED LIVE (migration 380), and the handoff's `purpose_field` candidate is REFUTED by this file's own mechanism section

The 2026-08-10 specimen's candidate fix ("deploy_asset gains
`purpose_field: input_data.spec.purpose` via the spec's Deprecated bridge")
**could never have worked, and this file already said so**: the bridge is
Strategy 3, which skips any populated field, and Defaults populate first —
"the deprecated `*_field` bridge for it is equally inert" (The mechanism, §3),
pinned by `TestPurposeFieldBridge_DeadForDefaultedField`. The candidate and
its refutation sat in the same file, written for different readers. Caught by
reading the resolver before implementing; logged in WRONG_CALLS.

**What shipped instead — migration 380** (applied + recorded 2026-08-11
~10:00Z, ROLLBACK sidecar alongside, council corr `a46a4421`):
build-dispatch-loop's `call_handler.input_mapping` gains
`"purpose?": "current_item.spec.purpose"`. Only a Strategy-0 dotted path on
`purpose` itself can beat the Default, and the deploy step's existing binding
(`input_data.purpose`) is already exactly that — it just had nothing to
resolve against on the `undeployed_asset` dispatch shape. The mapping gives it
something, on the same idiom site-work-orchestrator's `fix_items_loop` already
carries (`"purpose?": "current_fix_item.spec.purpose"`) — which is the
evidence the omission was never a design choice.

Blast radius measured before applying (queries in RUNBOOK_209): exactly two
live definitions bind `input_data.purpose` — asset-deployer (the fix target)
and image-build-handler's `check_logo_or_hero`, whose `purpose == 'logo'` arm
was half-dead on this shape and now activates (needs_imagery brand-update
logo items route down the logo branch: the condition's stated intent, this
family's 235 fix). The 11 no-mode favicon/og_card items flip from latent
hero-deploys to clean 179-B refusals — the guard fires on the RESOLVED
purpose. Items without spec.purpose: the `?` mapping skips silently.

Behavioural proof, same hour: relojistas item `6084d849` re-dispatched
through the repaired mapping committed **"Deploy logo image"** (pre-fix: two
runs of "Deploy hero image" against identical row state), asset row restamped
`logo.png`, served artefact `https://relojistas.com/assets/images/logo.png`
200 `image/png` PNG 400×170 RGBA.

**What keeps THIS file open, unchanged in kind:**
- The fleet-class census, now with its three arms (shadowed static ·
  unresolvable-dotted-path-with-Default · the original 61-spec sweep) — the
  third arm has its first confirmed-and-fixed instance but has not been
  enumerated.
- Candidates 2 (config-static beats Defaults in the helper — needs the census
  first) and 3 (CheckConfig flags shadowed statics — cheap, catches future
  authors; would NOT have caught either live face, but still worth having).

---

## 2026-08-11 (afternoon) — the CENSUS IS DONE, all three arms; candidate 3 is BUILT; and the census found a THIRD live face: four auditors' `audit_source` is dead and every finding ships as "design-audit"

Method note (per the 2026-07-31 owner ruling): 090 was not re-run for the new
instance; substituted first-hand verification, stated plainly below — the
mechanism half was already twice through the loop (both runs UNVERIFIABLE
scope-not-narrowing, both independently CONFIRMING the helper mechanism), the
checker that found the instance **executes** nothing it did not mirror from the
resolver, and the instance claim is proven at the read site and at the artefact
rows, each quoted below.

### The instrument (candidate 3, built and calibrated)

`cmd/config-key-audit --default-shadowed-keys` (wrapper:
`scripts/audit-default-shadowed-keys.sh`), same idiom as its sibling modes:
asks the **binary** for every registered spec (164 registered, **62 carry
Defaults, 232 defaulted fields** — the bug's "~10+" estimate was out by 6×),
walks live definitions with `validation.WalkSteps`, and classifies every config
entry bound to a defaulted field by the exact resolver arm that kills it:
`static_string` · `non_string_literal` · `composite_literal` ·
`deprecated_bridge` · `unextractable_field` (a defaulted field outside
Required+Optional — dead to EVERY config shape, dotted included; strategies
0/4/5 iterate Required+Optional only, which this file's mechanism section
predates) · `dotted_conditional` (arm 2 — reported, never fatal, because
resolvability is a runtime fact). Calibrated on BOTH live faces as committed
tests: the pre-348 static fires as `static_string`/mismatch against the real
registry; asset-deployer's `input_data.purpose` reports as
`dotted_conditional`. Exit 1 only on a dead entry whose value mismatches its
default.

### Census results, 2026-08-11 (182 live agents, snapshot in scratchpad + re-runnable any time)

- **24 dead-mismatched** (config says one thing, inputs path delivers the
  default) · **75 dead-matching** (value restates the default — invisible
  today, inert on first edit) · **96 dotted_conditional** (34 where the config
  path differs from the default).
- **Read-path verification of all 24 mismatched** — the caveat this file's 235
  correction taught ("an action can read config directly and honour the
  static") turned out to be the RULE, not the exception: **20 of 24 are
  honoured through direct `config[...]`/`GetIntField(config,…)` reads** in the
  action body (`diagnose_council_decide` max_rounds:544,
  `revalidate_review_queue` dry_run:287, `diagnose_persist_fix_plan`
  max_plan_bytes:150, `diagnose_build_gate`, `execute_vision_prompt`,
  `checkpoint_for_review`, `diagnose_dormant_agents`, `diagnose_route`,
  `normalize_to_feed_items`, `diagnose_code_lookup`, `analyse_repo_local`,
  `feed_normalize` — every one verified at its read line). No live damage
  there; for those actions the spec Default is documentation that must be kept
  in sync by hand with the in-body fallback.
- The `*_field` indirection family (`append_doc_note`, `write_doc_plan`, the
  diagnose plan/council fields) also reads config directly, so its 96-strong
  presence in the dotted/matching buckets is extractor-irrelevant.

### The third live face — `[MEASURED]` at mechanism, read site AND artefact

**Four live auditor agents set a distinctive `audit_source` static on
`write_audit_findings`, whose spec carries `Defaults{audit_source:
"design-audit"}`. All four statics are dead; every finding all four have ever
written is labelled `design-audit`.**

- Config: `brief-fidelity-auditor` → `'brief-fidelity-audit'` ·
  `content-quality-auditor` → `'content-quality-audit'` · `site-review-agent`
  → `'site-review'` · `visual-design-auditor` → `'visual-design-audit'`.
- Read site: `write_audit_findings_action.go:495` — `inputs.Get("audit_source")`,
  followed by the signature unreachable arm `if auditSource == "" {
  auditSource = "design-audit" }` (the Default guarantees non-empty). No direct
  config read anywhere in the action.
- Artefact: fleet-wide, **zero** `site_work_items` rows carry any of the four
  intended labels; 136 rows carry `design-audit` (2026-04-09→2026-08-11,
  still being written); 2 rows are the internal contradiction that settles
  attribution — `item_type='audit_finding_brief_fidelity'` (2026-07-24) with
  `spec->>'audit_source'='design-audit'`: the type names brief-fidelity, the
  label claims design-audit. The 2026-08-04→08-11 cluster mixes
  content-quality-shaped (`content_rewrite`, `tone_shift`),
  visual-design-shaped (`needs_design_review`, `spacing_fix`,
  `hardcoded_section_colors`, `dark_section_audit`) and site-review-shaped
  types, all under the one label. Per-row attribution beyond that is
  **impossible in principle** — the discriminator is the thing that is dead,
  which is the damage.
- Blast: `bugs_open/213`'s regression guard names `spec->>'audit_source'` as
  THE producer-measurement key ("NOT item_type, NOT created_by"). That
  instrument currently sees ONE producer where there are at least five. The
  two correct `tool-acceptance-tier4` rows come from a Go literal in
  `tool_acceptance_actions.go:1267`, not through the shadowed config — which
  is why they escaped. Noted in 213's file (consumers must be told, owner
  ruling 2026-07-29 §3).

### What the census changes about the candidates

- **Candidate 3: DONE** (this instrument). It caught the third face on its
  first fleet run, which the original file predicted it could ("catches future
  authors") but undersold — it also catches the *existing* ones.
- **Candidate 2's blast radius is now MEASURED, and it is startlingly small:**
  activating "config-static beats Default" would change live behaviour for
  exactly the dead-mismatched entries whose actions read via the inputs path —
  which today is **the four `audit_source` entries and nothing else** (75
  matching = no-op by equality; the other 20 mismatched = no-op because the
  action never consults the inputs value for that field — re-verify the
  read-path table at implementation time, it is a point-in-time census).
  Design questions for its round: composites (Strategy 5 deliberately excludes
  them), and the explicit-empty-string case (`changed_files_field: ''` exists
  live on feature-implementer, authored as "disable").
- **The instance repair has NO config-only path.** Unlike 348/380 there is no
  data path holding the wanted value — it is a per-agent constant, and this
  file's own rule says only a resolving dotted path can beat a Default. Even
  REMOVING the Default does not work: a dotless static on a non-defaulted
  field is read as a single-segment collectedData reference (Strategy 4),
  resolves nothing, and the action's `== ""` fallback re-imposes
  "design-audit". The real options: (a) the action reads `audit_source` from
  config directly with the inputs value as fallback (the idiom 20 of its
  sibling actions already use — four lines, one action); or (b) candidate 2.
  Not implemented here — needs its own round, and `write_audit_findings` is
  a file the `bugfix_213` lane is actively working.

### Arm 2 completion (same session, later) — every diverging dotted binding assessed

The 34 dotted_conditional findings where the config path differs from the
default decompose completely:

- **5** — `deploy_image_asset` `purpose` bindings: the 348/380 repairs
  themselves, live-verified previously. Correctly-authored arm-2 shape.
- **28** — the `*_field` indirection family (`append_doc_note`,
  `write_doc_plan`, `write_experience_pattern`, `load_doc_context` via the
  shared `docResolveSubject` helper at `write_doc_plan_action.go:148`, and the
  diagnose `plan_field`/`loop_scope_field`/`changed_files_field` set): every
  one verified at its read line to consume config DIRECTLY, so the extractor's
  behaviour is irrelevant to them. Not arm-2 exposure.
- **1** — `asset-deployer.derive_card_asset_step` binds
  `entity_type: input_data.spec.entity_type` (default `page`), read via the
  INPUTS path (`derive_card_asset_action.go:84`) on the same dispatch geometry
  as the 380 face. Assessed BENIGN today by the action's own guard: v1
  supports only `page` (`:89` errors loudly on anything else), so an
  unresolvable path falls to the only value that works, and a resolved
  non-page value cannot fail silently. **This row becomes the 380 shape the
  day phases I5/I6 add `news`/`products`** — whoever builds those must give
  the dispatch shape a top-level mapping or verify `input_data.spec.*`
  arrives on every path that can reach the step. Recorded here so that
  session finds it by grep.

The 62 matching dotted bindings (config restates the default path) are no-ops
by equality. **The census is closed: every live config entry bound to a
defaulted field is now enumerated and classified, and every diverging one has
a read-path verdict.** Remaining open on this file: candidate 2's decision
(blast radius measured above), and the audit_source instance repair (owner /
213-lane call).

---

## OWNER RULINGS 2026-08-11 (evening) — candidate 2 SHIPS; the instance fix goes first

Three decisions put to the owner after the census closed (costing above):

1. **The `audit_source` instance: option (a) NOW, option (b) LATER.** The
   four-line direct-config read in `write_audit_findings` ships as its own
   task, and candidate 2 still ships afterwards. The two are not alternatives.
2. **Candidate 2: SHIP.** "An explicit config value beats a Default" becomes
   the resolver's rule — `ExtractActionInputs` stops treating a spec Default
   as though it were data. This is the fix that makes the whole class
   unrepresentable rather than merely detectable.
3. **The stale `logo.jpg` files: LEAVE.** Not deleted, off the work-list.
   (Recorded here because a future reader will otherwise re-propose it: zero
   renderable references remain, so this is a tidiness question the owner has
   answered "no" to, not an open exposure.)

**What ruling 2 means for this file's own pinned tests.** Candidate 2 inverts
the behaviour that `TestLegacyLogoStep_StaticPurposeIsShadowedByDefault` and
`TestPurposeFieldBridge_DeadForDefaultedField` exist to pin, and it changes
what the detector's `static_string` / `non_string_literal` classes MEAN (they
describe entries that are dead under the old rule and live under the new one).
Both must flip in the SAME commit as the resolver change, citing this ruling —
a resolver that beats Defaults while a detector still reports those entries as
"dead" is a tool lying about production, which is the exact failure this lane
spent two days measuring. Sequencing, design questions (composites, the
explicit empty string, whether the Strategy-3 bridge also beats a Default) and
the council requirement are in the handoff's Task B.

**Re-measure before implementing.** The activation-set figure (4 entries)
was measured 2026-08-11 and **goes to zero once option (a) ships** — which is
the intended order. Re-run `scripts/audit-default-shadowed-keys.sh` plus the
read-path check at implementation time rather than quoting the number from
here; a census is a snapshot, and this one is deliberately about to change.

---

## CANDIDATE 2 SHIPPED — 2026-08-14, commit `d3edb5b89` (Go: INERT until the next chassis roll)

Ruling 2 executed. `Council-Submitted: 41a01378-1211-4987-966d-f8b6e2fddce1`
(platform/ is in scope this time; verdict not yet read at time of writing —
whoever reads it owns acting on a REVISE, because the code is already on the
shared branch). Registered as **CTS-059** in the same commit, per the ordering
exemption's surviving condition (2).

### The mechanism section of this file was RIGHT but understated the reach

This file said Strategy 0 was the only arm that could beat a Default. Reading
every arm line-by-line to implement the fix showed *why*, and it is stronger than
"strategies 1/2/3 skip populated fields": **nothing anywhere ever deletes from
`result.Values`.** `delete(result.Defaulted, field)` at Strategy 0 touches the
provenance map only. So the has-value skip at the head of Strategies 1, 2, 3, 4,
5 AND the nested-object backward-compat block can never pass for a defaulted
field. Six arms, not three.

**That is the blast-radius proof, and it is why this change did not need an
estimate.** A field Strategy 6 can touch is one no other arm could reach ⇒ no
behaviour that works today can change. Everything else is arithmetic on the dead
set.

### What shipped

- **Strategy 6** (`action_inputs.go`): for a field still holding only its Default,
  an explicit dotless config SCALAR of the Default's kind becomes the value, and
  `Defaulted` is cleared. Guards: a dotted string stays dead (a dot means
  REFERENCE — `bugs_open/248` finding (a), 150+ page-visible 404s named
  `input-data.asset-key.jpg`, same discriminator, measured against 478 asset
  rows); composites still refused; a kind guard (`datahelpers.LiteralKind`)
  refuses a scalar whose type differs from the Default's; an explicit `""` cannot
  override a Required field's Default.
- **Strategy 3 bridge** now beats a Default when its path resolves — it performs
  Strategy 0's operation under a deprecated spelling. Zero live definitions carry
  a Deprecated alias for a defaulted field, so this arm is correctness for the
  next author, not a live change. The alias is still REPORTED in
  `DeprecatedUsed`, asserted by a test, so beating the Default does not quietly
  bless the old spelling.
- **Detector re-specified in the same commit** (`defaultshadow.go`): `static_string`
  and `non_string_literal` → `live_override`; `type_mismatch` and
  `required_empty_string` are the new dead classes; `deprecated_bridge` moved from
  dead to conditional. The binary now emits a per-finding **`verdict`** and the
  wrapper script groups on it instead of re-deriving `class != "dotted_conditional"`
  — that second copy of the rule would have printed 99 working entries as dead
  keys while the binary exited 0.
- **Two pinned tests flipped**, both now also asserting provenance is cleared,
  because `deploy_image_asset` reads `WasDefaulted("purpose")` before letting an
  asset row's purpose win (248 finding (b)): an override that set the value but
  left `Defaulted` intact would still read as "nobody said anything".

### Re-measured at implementation time, as this file's own last paragraph demanded

`[MEASURED 2026-08-13/14 — one session, spanning midnight; the BEFORE census is 08-13, the AFTER census and the demand control are 08-14]` — 184 live agents, 61 specs with Defaults (was 62; the
`write_audit_findings` spec dropped out when `bugs_open/264` removed its only
Default):

| | before | after |
|---|---|---|
| dead mismatched | 21 | **0** |
| dead matching | 78 | **0** |
| conditional (dotted + bridges) | 96 | 96 |
| live overrides | — | **99** |
| exit | 1 | **0** |

**The 24 → 21 drop between 08-11 and 08-14 was not this lane.** `bugs_open/264`
took the four `audit_source` entries out (migration 399 + `audit_source` made
Required with no Default), which is −4, and **one NEW entry arrived**:
`render-audit-agent steps.audit request_render_audit max_pages=60 (default 25)`,
from migration `392` (another lane's weekly-rotation work). Investigated before
anything else, per the previous handoff's cold-start rule: **benign** — read
directly at `request_render_audit_action.go:98`
(`datahelpers.GetIntField(config, "max_pages", 25)`), so the live cap was already
60. It is the 21st member of the direct-read false-positive family, not a fresh
instance of the class.

So candidate 2's **activation set was empty by construction** — exactly the order
ruling 1a intended, though reached by 264's route rather than the four-line
config read that ruling described.

### The zero has a demand control, because I edited both the resolver and its detector

A post-fix zero from a detector the same commit re-specified is the shape that
should be distrusted. One live `max_pages: 60` was mutated to the string `"60"`
in the same live export and fed back through the same binary: **exit 1, 1 dead
mismatched, classed `type_mismatch`, live overrides 99 → 98.** The zero is a real
zero, and the pipeline (live query → binary → script → exit code) can still
return non-zero.

### The remaining exposure, unchanged and deliberate

The **96 dotted_conditional** entries still fall back to their Default silently
when a path does not resolve — this file's second face, and it stays open by
design: resolvability is a runtime fact an offline check cannot decide. The one
latent row named earlier in this file (`derive_card_asset entity_type`, benign
until phases I5/I6) is unaffected.

### The wart, registered rather than hidden (CTS-059's landmine)

The same dotless string now means different things either side of a Default:
`analysis_field: "repo_analysis"` on a **defaulted** field is a LITERAL
(Strategy 6); on a **non-defaulted** field it is a single-segment REFERENCE
resolved against collected_data (Strategy 4). No live config can depend on the
other reading — Strategy 4 was never reachable for a defaulted field — and all 48
live dotless statics are plainly values their authors typed (`repo_name:
'agentchassis'`, `ref: 'main'`, `country: 'GB'`, `severity: 'high'`). But you
cannot answer "is this string a value or a path?" from the config alone any more:
you must check whether the action's spec defaults that field.

**Open review question, named at registration and deliberately NOT taken:**
whether a resolving dotless string on a defaulted field should instead resolve as
a collected_data reference. It is the one arm that could replace a typed Default
with a resolved object of unknown shape (`ActionInputs.Get` returns `""` for a
non-string), and zero live entries want it. Whoever revisits it owns re-measuring
the `*_field` family first — 28 of them read config directly today, which is
precisely what makes the change look free.

### What is left on this bug

1. **Read the council verdict** (corr above) and act on a REVISE/REJECTED.
2. **Post-roll proof**, once a chassis image carrying `d3edb5b89` ships: expect
   the `Strategy 6: explicit config value beat the spec default` Info line from a
   step carrying a dotless static on a defaulted field, and
   `scripts/audit-default-shadowed-keys.sh` still at 0 dead. Expect **no**
   `Strategy 6: config value's type differs` Warn anywhere — no live entry
   mismatches kinds today, so one would mean new config arrived.
3. The dotted_conditional census (96) remains this file's open half.

---

## POST-ROLL, 2026-08-14 — LIVE on `v1.0.1298`, and the behavioural check FAILED for a reason worth more than the check

**The code is running, proven at the artefact, both replicas.** Stamp `bc39e7bf5`
**PRESENT** in both pods' `/proc/1/exe`; a later commit `d11fb2a44` **ABSENT** in
both, so the probe discriminates rather than matching anything; and
`git merge-base --is-ancestor` true for `d3edb5b89` (the seam) and `14e4333f7`
(the REVISE round). Post-roll census: 185 agents, **0 dead**, 96 conditional, 99
live overrides, exit 0.

**The behavioural half could not be read, and the control is what says so.** I went
to the logs for `Strategy 6: explicit config value beat the spec default` and found
**zero**. That reading is worthless: the pod retains **243 lines spanning 92
SECONDS** (13:51:23Z → 13:52:55Z) on a pod started **08:58:03Z**. The discriminating
check is that **Strategy 0's PRE-EXISTING Info line is also absent** from the same
window — so the absence measures the retention window, not the resolver. 241 of the
243 lines are `level:info`, so the level is not filtered either.

**This is `bug_historian`'s council objection, confirmed by measurement, while
verifying the change it was raised against.** The seat said a log-only rejection
surface is not durable on this fleet because chassis logs rotate within minutes. It
was right, and the evidence is that I could not verify my own change through them.
`--report` (round 2) exists for exactly this; it is built and **undriven**, and the
CronJob that would drive it is deliberately not shipped until the image exists
(ImagePullBackOff reports as a Job still RUNNING on this fleet).

**So the one check still owed is: watch `Strategy 6` fire.** Not from `--tail` on an
old pod — `logs -f` on both replicas, or drive a step carrying one of the 99 live
overrides. Absence of the type-mismatch Warn is a genuine expectation (no live entry
mismatches kinds), but it inherits the same blindness until the window question is
settled, so do not report it as a pass on its own.

### Round 1's verdict and what it changed

REVISE on corr `41a01378`, 11 seats, 6 abstained, no truncation, decided by a gating
objection from `guardian`. **Four objections were real defects**, closed in
`14e4333f7`: the unreachability proof is now an ENFORCED invariant with its own
vacuity control; the canonical config key now beats a deprecated alias (Strategy 3
ran first and would otherwise win); both non-defaulted bridge arms are pinned; and
`--report` gives a rejected override a durable `doc_notes` surface. Two seats' prior-art
questions were answered with greps rather than code (`LiteralKind` duplicates nothing;
candidates 1 and 3 both SHIPPED and candidate 3's council attempt was refused
CLIENT-SIDE on scope, never vetoed). The architecture seat's `needs_rfc` was **not**
argued down — it is routed to **RFC_028**, which supplies the measurement it asked for
and printed as unknown: **27 council rounds have touched this resolver, 8 drew
`needs_rfc`, 1 was ever vetoed**. Round 2 is with the council under the same
correlation.

---

## 2026-08-14, later — the owed check is CLOSED, `--report` has written a real row, and rounds 2 and 3

### Strategy 6 observed firing, both replicas — streaming, not tailing

`[MEASURED 2026-08-14 13:56–14:30Z]` **six** `Strategy 6: explicit config value beat
the spec default` lines (`action_inputs.go:923`) across **both** pods:

| pod | field | value | default | action |
|---|---|---|---|---|
| dphbw ×2 | `max_plan_bytes` | 65536 | 32768 | `diagnose_persist_fix_plan` |
| dphbw ×2 | `max_rounds` | 3 | 2 | `diagnose_council_decide` |
| 6tfxf | `max_plan_bytes` | 65536 | 32768 | `diagnose_persist_fix_plan` |
| 6tfxf | `max_rounds` | 3 | 2 | `diagnose_council_decide` |

Both are **entries from this file's own 21-strong dead-mismatched census**, which means
**the council that reviewed this fix was running on the configuration the fix
repaired.** Zero `type differs` and zero `empty on required` Warns in the same window —
meaningful only because `Strategy 0: Resolved config path` fired throughout as the
liveness control.

**LIMIT, so nobody over-reads it:** `diagnose_council_decide` and
`diagnose_persist_fix_plan` both read their value from step config DIRECTLY in the
action body (this file's read-path table). So this proves Strategy 6 **fires and applies
the value**; it does NOT show a behaviour change. That is consistent with the central
claim of zero net live behaviour change, not evidence against it.

**The method matters more than the result.** `kubectl logs --tail=200000` returned
**zero** — and that reading was worthless, because the pod retains **243 lines spanning
92 seconds** and Strategy 0's *pre-existing* line was missing from the same window.
Streaming `logs -f` on both replicas, with a pre-existing line in the same filter as the
control, is what answered it.

### `--report` has produced a real `doc_notes` row

Run against the live fleet: `subject_key='default-shadowed-keys'`,
`source='default-shadowed-keys-check'`, `created_at 2026-08-14 16:34:31+00` — 185
agents, **0 dead**, 96 conditional, 99 live overrides, exit 0. So the durable surface is
real, not prospective. **Nothing schedules it yet**, so `bug_historian`'s objection is
**MITIGATED, not closed**: between runs a mistyped setting is still refused with no
fleet-visible signal. The CronJob needs its image built and pushed first — applying the
overlay before the image exists gives an ImagePullBackOff, which this fleet reports as a
Job still RUNNING and never FAILED, so shipping the yaml early would swap a known gap
for a silent one.

### Round 2: REVISE, and its lesson is about me, not the code

12 seats, 5 abstained, no truncation, **gated by `prior_art_librarian`**. All four
findings were about argument or an inert control; none disputed the code.

- **HIGH, and it is right: I argued a gating objection down with a citation the
  reviewers cannot read.** Council seats have no access to `CLAUDE.md`, so "the owner
  ruled to ship this" is, from their side, an unverifiable appeal to authority used to
  dismiss a gate. **A reviewer who cannot check a claim is right to refuse it.** Round 3
  makes no architecture argument at all — the signal is routed to `RFC_028` for a human
  and left explicitly OPEN.
- **MEDIUM: my 27-round figure had no query attached**, given to the one seat whose job
  is verifying exactly that class of claim. Round 3 ships the SQL and states its limit
  (the `needs_rfc` count matches report TEXT, so it is a floor).
- **LOW: the "no duplicate `LiteralKind`" proof was a content grep**, unreliable here
  (`bugs_open/108`). Replaced with a **declaration** search over signatures: 21 funcs in
  datahelpers match `(interface{}) → string`, none classifies a value's kind; the
  detector declares none. Nearest neighbour named rather than hidden —
  `InterfaceToString` switches on the same type set but returns the value **rendered as
  text**, mapping both `25` and `"25"` to `"25"`, collapsing the very distinction the
  kind guard exists to make.

### Round 3 submitted

Same correlation `41a01378`, `RUN_ORCH_ID=f80e528b-f24c-4544-bfae-bae08d9a7815`.
Verdict **not read at time of writing**. No `Council-Reviewed:` trailer exists on any
commit in this lane and none may be written until an approved verdict is read.

### Round 3: died on the account cap, refired after the restore, **APPROVED** 17:29:51Z

The round submitted at 16:36Z never reached a verdict. It passed `persist_submission` and
died at `complete_invalid` on `review_editquality` — the Anthropic account usage cap
(`bugs_open/243-anthropic-cap`, third exhaustion in 15 days), not a fault in the
submission. Refired **unchanged** at 17:10:27Z once the cap was lifted (`a41e1677`).

**Verdict: APPROVED** — 13 reviewers, 4 abstained, `gated_by_truncation: false`,
**5 advisory objections, none high-severity**. Trailer, now legitimately writable:
`Council-Reviewed: 41a01378-1211-4987-966d-f8b6e2fddce1`.

The objections are recorded here with dispositions, because two of them are claims about
**this lane's evidence** and one of those is right.

1. **`architecture` — MEDIUM, `needs_rfc`, DEFLECTIONS: 8. Recorded OPEN, as it asked.**
   It explicitly did not object to the patch shipping ("narrowly scoped, blast radius
   bounded by the unreachability proof, now enforced by a test rather than prose") and
   recorded the signal so the disposition "stays visible rather than being silently
   absorbed into an approve". That is `RFC_028`'s question and it remains a human's.
2. **`guardian` — MEDIUM + LOW. Recorded OPEN, as it asked.** Not a veto; it asked that
   the record show blast-radius certainty is "bounded by one test's coverage, not by
   pipeline enumeration", and that the Strategy-3 bridge precedence rests on a
   **point-in-time census** rather than a forward constraint — any future
   `agent_definitions` row carrying both a deprecated alias and its canonical key on one
   defaulted field silently gets the new precedence, with the unit test as the only
   backstop. **Both stand as written. The second is the more durable and belongs in
   RFC_028's scope, not in a closed bug.**
3. **`bug_historian` — MEDIUM: enumerate the newly-live overrides that sit in a
   rebuild/rerender/render pipeline. DONE, and the answer is zero.** `[MEASURED
   2026-08-14 17:40Z]` Of the 99 live overrides, **21 change a value and 78 equal their
   own default** (inert by construction — honouring them cannot alter anything). Of the
   21, **none writes to `page_components`, `rendered_html`, `content_data` or
   `site_components`**. The three nearest rendering are read/audit-side bounds:
   `render-audit-agent request_render_audit max_pages=60` (how many pages an audit
   reads), and `tool-acceptance-agent execute_vision_prompt images_field='browser_run'`
   / `max_images=4` (which render array the vision step looks at, and how many).
   **And all three read their step config DIRECTLY** — `GetIntField(config,"max_pages",25)`
   at `request_render_audit_action.go:98`, `GetStringField(config,"images_field",…)` /
   `GetIntField(config,"max_images",16)` at `execute_vision_prompt_action.go:132-133`,
   `config["severity"].(string)` at `checkpoint_for_review_action.go:109` — **so they
   never went through the resolver and Strategy 6 does not change them at all.** This
   generalises the limit round 3 already stated for the two council entries: `live_override`
   is a statement about what the RESOLVER would now honour, and it **over-counts
   behaviour change** wherever an action reads `params.StepConfig.Config` itself.
   > **CORRECTED 2026-08-15:** that limit was NOT discovered here. It is documented in
   > `scripts/audit-default-shadowed-keys.sh`'s header, at `defaultshadow.go:90`, and in the
   > **report text the tool prints** (`defaultshadow.go:413`), all naming `bugs_open/235`.
   > What this round adds is only the **enumeration** — which three of the 21 are in that
   > class, with file:line. `WRONG_CALLS.md` 2026-08-15.
4. **`debug_historian` — MEDIUM: the `/proc/1/exe` stamp citation. CHECKED, does not
   survive.** The landmine it cites (*"the provenance recipe is INOPERATIVE on
   agent-chassis"*) is about the **startup LOG line** rotating away, and was itself
   REFINED 2026-08-13 to "TIME-LIMITED, `INOPERATIVE` is too strong". Its "trap inside
   the trap" bullet condemns probing `/proc/1/exe` **with your own commit's sha**, because
   the binary carries exactly one commit — the build point. This lane probed
   `bc39e7bf5`, which **was** the build point, and ran the mandated ancestry comparison:
   `git merge-base --is-ancestor d3edb5b89 bc39e7bf5` → yes; `14e4333f7` → yes; a commit
   made after the roll → correctly **not** an ancestor. Two-sided, control passes. The
   seat read the entry's title, not its arms.
5. **`debug_historian` — LOW: "both replicas" is a sample. CORRECT, and now MEASURED
   rather than caveated.** `[MEASURED 2026-08-14 17:36Z]` **17 pods run this binary**, not
   2 — `agent-chassis` ×2 plus `vet-intel`, `business-intel`, `agent-diagnose-orchestrator`,
   `agent-image-build-handler` and nine ephemeral `agent-build-dispatch-loop` pods. The
   2-pod check was a sample of 17. **The population is uniform and has moved on: all 17
   are on `v1.0.1299`, not the `v1.0.1298` this file certified.** Stamp of the live build,
   read from a pod 3 seconds old whose log demonstrably reached back to startup
   (`logs | head -1` = "Logger initialized successfully", the landmine's own
   discriminator): **`6f8efa158`**. `d3edb5b89` and `14e4333f7` are both ancestors of it;
   a commit made after the roll is not. **So the seam survived the re-roll and is live
   fleet-wide on every pod running this binary** — a stronger claim than the one the seat
   objected to, and it took the objection to go and get it.
6. **`editquality` — MEDIUM: CTS-059 may be registered at the wrong path. REFUTED.** It
   inferred from a landmine quoting `003_contracts_and_standards.md` that the register uses
   numbered filenames. It does not: `ls docs/agent_docs/docs026_concept_register/register/`
   shows exactly one numbered file, `000_concept_index.md`, and CTS-059 is in
   `contracts-and-standards.md` **and** in the index.
7. **`tooling_provenance` — LOW: the prior-art declaration search should have been a code-index
   bundle query.** Fair and not blocking; it explicitly accepted the reasoning about the
   index's content-search limitation. Noted for the next round in any lane.

**Seats approving outright:** `reuse_agent`, `guidelines`, `diagnosis_guardian`,
`render_guardian`, `constitution`, `mission`, `prior_art_librarian`. `render_guardian`
independently reached objection 3's conclusion by a different route — "no edit modifies
`rerender_pages_actions.go`, `rerender_single_page_action.go`, `save_page_sections_action.go`,
`component_library.go`, or any CSS/template file".
