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
