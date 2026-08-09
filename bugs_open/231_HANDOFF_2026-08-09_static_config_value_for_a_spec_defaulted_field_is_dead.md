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
