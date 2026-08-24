# PLAN — bugs_open/382: an image request with an EMPTY `kind` routes to SDXL, silently

**Lane opened:** 2026-08-24. **Bug:** `bugs_open/382_HANDOFF_2026-08-24_empty_kind_image_requests_still_route_to_sdxl_silently.md`
**Parent:** `bugs_closed/011` (hero routing). **Filing lane:** vetcomparison (site work done, fleet
half handed off — this lane).

## 0. Why this lane exists

382 was filed with the symptom measured and the *caller-side* half deliberately not diagnosed
(the filing lane said so in §5.1). This lane does that read, then fixes the mechanism rather than
the instance.

## 1. What the bug is, in plain terms

The image adapter picks which AI image model to use from a single string called `kind`
("hero", "logo", "icon", …). A table in `internal/adapters/imagegenerator/routing.go` maps every
known kind to **Banana** (Google's Gemini image model — renders legible text, honours per-site
reference images). Anything the table does not recognise falls through to **Stability/SDXL**, an
older model that cannot render text and ignores reference images.

That fall-through has two doors:

- **A kind the table does not know** (say `"diagram"`). The code notices, sets `UnmigratedKind`,
  and a warning row lands in `agent_error_log`. This door is guarded.
- **An EMPTY kind.** The code deliberately says nothing — the comment calls empty-kind callers
  "legacy … a documented, deliberate Stability path, not an oversight". This door is **not**
  guarded, and it is the whole live defect.

## 2. The root cause on the caller side — FOUND (this lane, 2026-08-24)

`kind` reaches the adapter only if it is present in the callee's `input_data`.
`resolveKind` (`platform/orchestration/actions/generate_image_actions.go:115-123`) reads
`inputData["kind"]`, then `inputData["default_kind"]`, else `""`.

For a `call_agent` step, `input_data` is built **solely** from the step config's `input_mapping`
— `extractDataForAgent` (`platform/orchestration/actions/call_agent.go:974-1018`). Nothing else
in the step config travels. So a `default_kind` key sitting at the step **config** level is
never in `input_data` and **has never done anything**.

Migration `390` (2026-08-11) found this and fixed *two* branches — `call_hero_gen` and
`call_logo_gen` — by mapping `"kind?": "input_data.spec.purpose"`. Its blast-radius paragraph
then claimed:

> "The Phase-2E branches (call_imagery_gen, call_variant_gen) already forward kind."

**That claim is FALSE for `call_variant_gen`, and still false in the live row today.**
`call_variant_gen`'s `input_mapping` is `{prompt, site_plan}` — no `kind`, no `site_id` — while
its config carries `"default_kind": "hero"`, which reads exactly like the thing that would have
saved it. `call_variant_gen` is the ONLY handler of `unfulfilled_hero_variant` work items, i.e.
every per-page hero (`hero_about`, `hero_services`, …).

**The proof is a minute-for-minute match** (both queries in the RUNBOOK):
5 `unfulfilled_hero_variant` items completed on ai-agent-orchestration.com at
16:28:33 / 16:37:23 / 16:38:33 / 16:39:58 / 16:42:01 on 2026-08-11; the 5 SDXL hero assets on
that site are stamped 16:28:04 / 16:36:53 / 16:38:08 / 16:39:30 / 16:41:21, asset_keys
`hero_about / hero_services / hero_contact / hero_tools / hero_case_studies`. Those runs are
**after** 390 was applied that same day at 13:42 BST — so 390 demonstrably did not cover them.

The 15th asset (mortgagecalculator.co.uk, plain `hero`, 2026-08-11 10:35) is **pre-390** and is
the case 390 fixed.

### 2b. The other empty-kind doors, live in `agent_definitions` as of 2026-08-24

| agent | step | kind in input_mapping? | note |
|---|---|---|---|
| image-build-handler | `call_variant_gen` | **NO** | dead `default_kind:"hero"`; all hero variants |
| image-build-handler | `call_imagery_gen` | `kind?` ← `spec.kind` | empty whenever the imagery spec omits `kind` |
| image-build-handler | `call_hero_gen` / `call_logo_gen` | `kind?` ← `spec.purpose` | fixed by 390 |
| pageflow-builder | `generate_hero_image`, `call_logo_generation` | **NO** | no kind, no default_kind at all |
| site-work-orchestrator | `generate_hero_image`, `call_logo_generation` | **NO** | same |

pageflow-builder / site-work-orchestrator show **no runs in the 1-day `orchestration_states`
window** and pageflow-builder has 0 `llm_call_log` rows ever — but `llm_call_log` is the wrong
instrument for an orchestrator, so their reachability is **[UNMEASURED] beyond one day**. They are
doors either way.

## 3. The fix, ordered by what closes the door

### A. CODE — the framework-wide half (`routing.go`) ← the real fix

1. **Empty kind routes to Banana, not Stability.** The "legacy deliberate" premise is refuted by
   measurement: of the 5 sites that got SDXL heroes post-fix, **none** carries
   `provider:"stability"` in its `imagery_style_guide` and **no** live agent definition pins a
   stability model (382 §2). The sanctioned opt-out already exists and is data, not code — a site
   that genuinely wants SDXL says so and still gets it (hint precedence is unchanged).
2. **Make the silence loud.** A new `MissingKind` flag on `routingDecision` plus a
   `MISSING_IMAGE_KIND` reported condition, so the next unmapped caller lands in
   `agent_error_log` with the orchestration context attached instead of surfacing months later as
   an owner complaint about eyes.

Why this altitude and not the config alone: it fixes **every** empty-kind caller at once,
including the two orchestrators nobody has audited and any future one, and it does so without
depending on a config author remembering anything. This is the same reasoning 011 used to replace
a `switch` with a table — applied to the one branch 011 left exempt.

### B. CONFIG — close the live door now (live on apply, no image roll)

A migration mirroring 390 exactly:
- `call_variant_gen.input_mapping["kind?"] = "input_data.spec.purpose"` — the spec always carries
  `purpose:"hero"` (`check_unfulfilled_image_prompt.go:44-60`), and `store_variant_asset` already
  reads `input_data.spec.purpose` from the same place, so the value is proven present.
- `call_variant_gen.input_mapping["site_id"] = "site_record.site_id"` — **a second, separate
  defect on the same step**: with no `site_id`, `getImageryStyleGuideForSite` is called with `""`,
  so variant heroes get no style guide, no `provider` preference, no `avoid` terms, no reference
  anchors and no `design_intent.imagery_direction`. `store_variant_asset` proves
  `site_record.site_id` is in scope at that point. Flagged separately because it changes prompt
  composition for variants (it makes them consistent with `call_hero_gen`, which already does it).
- **Delete the three dead `default_kind` keys** (`call_hero_gen`, `call_logo_gen`,
  `call_variant_gen`). They are read by nothing, and on `call_variant_gen` the dead key is
  precisely why the missing mapping looked covered. Measured: `default_kind` appears on exactly
  **3** live `call_agent` steps, all in image-build-handler (2026-08-24, RUNBOOK query).

### C. DOCS — retract the guidance that prescribes the defect

`016b` §9 "A dispatch table's `default:` branch is a silent bug factory", bullet *"Warn on the
unhandled case, not on the fallback itself"*, currently instructs: *"an empty kind is a documented
legacy path that legitimately uses the fallback, so it must not warn"*. That is the exemption this
bug walked through. It gets a dated correction, not an edit.

### D. NOT proposed (measured and set aside)

A general `config-key-audit` mode for "config keys on a `call_agent` step that nothing reads".
Censused first rather than argued: across all live `call_agent` steps the config-key population is
`target_role` 59, `timeout_seconds` 58, `input_mapping` 57, `agent_type` 42, `output_mapping` 8,
`error_step` 7, `default_kind` 3, `prompt` 1, `input_data` 1 (2026-08-24). So the entire dead-key
surface is ~11 steps and `call_agent` declares no `ConfigKeys` at all, which the audit would need
first. Candidate A makes the `default_kind` instance harmless anyway. Recorded here so the next
lane does not have to re-measure. Two adjacent, **[UNVERIFIED]** smells fell out of the census and
are noted for whoever wants them: `error_step` inside `config` on 7 steps (it is a *step*-level
field), and `prompt` inside a `call_agent` config on 1.

## 4. Verification (the fix's own close condition)

- Unit: `routing_test.go` — `TestRouteProviderEmptyKindIsLegacyNotUnmigrated` **inverts**; it
  currently pins the defect. Replace it with a test that pins empty → banana **and** the
  `MISSING_IMAGE_KIND` condition, and keep the explicit-`stability`-hint escape proven.
- Config: re-read the live row after the migration; `kind?` present on `call_variant_gen`.
- Behaviour, with a **demand control** (a post-fix zero over no traffic proves nothing): drive one
  real `unfulfilled_hero_variant` through the pipeline and read `assets.origin_model` on the
  resulting row. Negative control: banana heroes keep appearing on the healthy path.
- Detector: a `MISSING_IMAGE_KIND` row in `agent_error_log` from a caller that still omits kind
  (the two orchestrators, if they ever run) — the point of A2 is that this becomes queryable.

## 5. Decisions and their reasons

- **2026-08-24 — fix the routing default, not just the one config row.** The owner's standing
  preference is the framework over the individual case, and the census says the individual case is
  not the only door: two orchestrators carry the same shape unaudited. A config-only fix leaves
  them, and leaves the silence.
- **2026-08-24 — do NOT make `default_kind` work.** Teaching `call_agent` to forward selected
  config keys into a callee's `input_data` is a change to a shared seam used by 59 live steps, for
  a key with 3 users, to solve a problem candidate A already solves. Delete the key instead.
