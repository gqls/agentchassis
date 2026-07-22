# PLAN — bugs_open/026 structural half: the input_schema dialect fail-open

**Started 2026-07-21.** Owner directive: *build the fail-open fix* (chosen over
close-on-diagnosis). This workstream owns only **Defect B part 2** of bugs_open/026 —
the durable structural cause. Defects A and B-part-1 are already fixed live by the
relojistas/027 thread (seed 179); this workstream does not touch them.

## The problem (corrected framing)

> **CORRECTION to the originating bug's framing.** bugs_open/026 wrote Defect B as
> *"validation treats `''` as present."* That is not the mechanism. The real cause is
> schema-**dialect** blindness, and it is upstream of any validator. Recorded here so the
> correction is not silently edited away.

A component declares its content fields in `content_components.input_schema`. Two dialects
exist:
- **v2** (current): `{"fields": {"<name>": {"source","required","type","llm_guidance",...}}}`
- **legacy JSON-Schema** (pre-v2): `{"type":"object","required":["x"],"properties":{"x":{...}}}`

Two independent readers each did `inputSchema["fields"].(map)` and on a miss returned
"no fields" — the **same blind spot in both**:
1. **Generation** — `plan_sections_action.go:1182` → on miss, "all fields from LLM" with an
   empty field-spec list → the page-content-writer is never told the field exists → it is
   never generated.
2. **Enforcement** — `missingRequiredLLMFields` (`json_envelope.go:193`), called by the
   render gate (`v3_site_actions.go:1631`) and rerender path (`rerender_page_sections_action.go:210`)
   → on miss returns nil → an empty required field is never caught.

So `news-listing`'s `required` headline (legacy dialect) was **neither requested nor
enforced** and served as an empty `<h1>`. Reading one dialect and returning "nothing found"
on every other dialect **fails open**: "I can't read this contract" becomes "there is no
contract."

## Design decision — normalise, don't reject

Two options were on the table:
- (A) *Fail loud* when input_schema is non-empty but unreadable.
- (B) *Normalise* the legacy dialect onto the v2 field view both readers already consume.

**Chose B.** It is strictly better: an old-shape component becomes **understood** (planned
for AND enforced) rather than merely blocked, and it reuses the existing machinery on both
sides (no second parser bolted onto each call site — the platform convention). One shared
reader, `schemaContentFields()`, is the single point that knows both dialects.

**Safety boundary that makes B correct:** the 7 legacy core sections
(hero/header/footer/cta/features/social-proof) carry **bare example-value** schemas
(`{"headline":"string"}`) with *no* machine-readable requiredness. `schemaContentFields`
returns `ok=false` for those (no `fields`, no `properties`), so their existing
"all-fields-from-LLM" path is preserved unchanged. Only the JSON-Schema dialect (which
carries `properties`/`required[]`) is projected. Empty `{}` also stays `ok=false`.

## Scope

> **WIDENED 2026-07-21 by the council's round-1 REVISE (bug_historian).** The original scope
> below was "the two readers that caused 026". bug_historian correctly asked whether two was
> really all of them — it was not (9 genuine `input_schema["fields"]` readers). Scope widened
> to the correctness readers + the direct safety-net audit + a fail-loud tripwire; see the
> round-2 classification.

**In (round 2):**
- shared reader **relocated to `datahelpers.SchemaContentFields`** (the home both `actions`
  and `discovery_checks` import — no cycle), returning `(fields, ok, fromLegacy)`.
- **rewired** (dialect-tolerant): `plan_sections` (generation), `missingRequiredLLMFields`
  (render gate), and **`check_required_fields_missing`** (post-deploy audit — the render
  gate's direct companion; rewiring it was bug_historian's completeness point).
- **fail-loud tripwire** `WarnLegacyDialect`, fired by the generation path (every build) and
  the audit (post-deploy) when a legacy dialect is actually projected.
- tests: reader unit tests in `datahelpers`, enforcement regression in `actions`.

**Left direct, on purpose (different legacy-consequence than "required field ships empty"):**
`compute_component_quality` (field count), `store_generated_component` (sync flags),
`load_existing_component` (field-name print), `check_image_source_unsatisfiable`, `ctafields`
(CTA derivation), `expectedItemFieldsFromComponentSchema` (array item-fields). A wrong
metric/CTA/array is a different, lesser bug class outside 026; rewiring them would shift
quality scores/sync flags with no benefit while the dialect is extinct. The tripwire makes any
legacy component visible fleet-wide regardless, so none can silently absorb a regression.

**Not ours (routed to bugs_open/015):** the two residual empty/stale news sections — `idea.uk`
(`headline=''`, `page_type='section-index'`, page 404s) and `ai-agent-orchestration.com/news.html`
(`page_type='content'`, old 2026-05-01 page). Both are mistyped-`page_type` orphans the news
renderer never touches — the 015 class, not this component's defect.

## Why this is defensive (stated plainly)

The legacy dialect is **extinct fleet-wide today** (0 of 173 components — verified live). So
this change fixes nothing currently broken; it removes a latent fail-open that re-arms if the
old shape returns via a config re-seed, a restored snapshot, or a `component-creator` run that
emits JSON-Schema. The owner weighed this and chose to build it: it is the structural half the
027 thread explicitly left open as 026's, and "structural fixes over patches" is a platform
value. Cost is one small Go change + test + one image roll.

## Phasing

1. **Diagnosis** — DONE, committed `428c3cc82` (bug file addendum + 016b §9 pattern).
2. **Code + test** — DONE, committed `fd87c8ebf` (4 files, actions package green, no v2
   regression). Inert until an image roll.
3. **Council review** — submitted `SUBMISSION_CORR=a85c1220-7174-41fe-8892-64009eadcf47`
   (see RUNBOOK for the verdict query). Advisory; deploy is gated on it by choice given the
   high blast radius of the two files.
4. **Build + deploy + verify** — after APPROVED: `make build-agent-chassis` from committed
   HEAD, bump IMAGE_TAG, push/deploy, pod-grep the new marker. Verify behaviourally (a
   legacy-dialect component with an empty required field is refused), not just by pod-grep.
5. **Close 026** — Defect A + B-part-1 closeable now (fixed live); B-part-2 closes when this
   ships. Move to /bugs_closed/ then.
