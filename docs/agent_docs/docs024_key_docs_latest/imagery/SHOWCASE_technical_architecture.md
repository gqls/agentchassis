# agentchassis — technical architecture, illustrated by the imagery pipeline

*The technical companion to `SHOWCASE_one_pager.md` and
`SHOWCASE_imagery_workstream.md`. Written for engineers. All examples are
real, from the 8–10 July 2026 imagery workstream; diagrams are ASCII so they
read anywhere. Verified against production 2026-07-10.*

---

## 1. Platform shape

One generic Go runtime (**agent-chassis**) executes declarative workflows.
An "agent" is a row in Postgres (`agent_definitions`): a JSONB workflow of
steps, each step naming a registered Go **action**. Complexity lives in the
Go actions; workflows stay thin. Agents communicate over Kafka; every agent
is an orchestrator that replies on its caller's response topic. 144 active
agent definitions share the one chassis image.

```
                 ┌────────────────────────────────────────────────┐
                 │                   Kafka (Strimzi)               │
                 │  system.agent.<type>.requests / .responses      │
                 └───────▲───────────────▲────────────────▲───────┘
                         │               │                │
   ┌─────────────┐  ┌────┴─────┐   ┌─────┴─────┐   ┌──────┴──────┐
   │  scheduler / │  │  agent-  │   │  spawned  │   │  adapters   │
   │  triggers    │─▶│  chassis │──▶│  handler  │──▶│ (image-gen, │
   │ (cron, kcat, │  │ (generic │   │  pods     │   │  scrape,    │
   │  dispatch)   │  │ runtime) │   │ (per-job) │   │  git, LLM)  │
   └─────────────┘  └────┬─────┘   └─────┬─────┘   └──────┬──────┘
                         │               │                │
                 ┌───────▼───────────────▼────────────────▼───────┐
                 │            Postgres (single source of truth)    │
                 │  agent_definitions · sites · pages ·            │
                 │  page_components · content_components ·         │
                 │  site_plans/_pages/_sections/_imagery ·         │
                 │  assets · site_work_items · site_specs          │
                 └───────────────────────┬─────────────────────────┘
                                         │ git commit (per site repo)
                                         ▼
                        GitHub → Actions → object storage/CDN
```

Key conventions: enum-like columns are `text + CHECK` (mirrored in Go
constants — both change together); versioned entities keep history rows;
every mutation ships as an idempotent SQL migration with a backup table and
a verification block.

## 2. The work-item economy (how the fleet maintains itself)

Everything the fleet does — build a page, generate an image, fix a bug —
is a row in `site_work_items`. Three populations write them: the build
pipeline (planned work), **discovery checks** (found work), and humans
(decisions). One dispatcher drains them.

```
   site_specs / plans           rendered sites (DB + git)
        │                              │
        │ build pipeline               │ discovery agents run
        │ emits planned work           │ registered checks (Go, no LLM)
        ▼                              ▼
   ┌─────────────────────────────────────────────┐
   │              site_work_items                │
   │  status: detected → triaged → claimed →     │
   │          complete | failed | needs_human_   │
   │          review | wont_fix                  │
   │  dedup: (site_id, item_key) partial index   │
   └──────────────────────┬──────────────────────┘
                          │ build-dispatch-loop (claims triaged,
                          │ one per site at a time)
                          ▼
              handler agents (spawned pods)
       image-build-handler · page-build-handler ·
       component-creator · site-design-planner · …
                          │
                          ▼
              results → DB → git → re-render
                          │
                          └────▶ next discovery pass verifies
```

Two properties matter. **Dedup**: the partial unique index on
`(site_id, item_key)` for non-terminal statuses means checks can re-run
forever without flooding the queue. **Honest state**: real step errors now
route through a `mark_item_failed` step before the workflow's error-complete
— a failed build reports as failed, not as silent success (fixed this week
after catching a Kafka reply race being swallowed).

## 3. The imagery pipeline, end to end

```
 build-site-planner (LLM)
   │  emits structured "imagery" block: site/pages/sections scopes,
   │  kinds logo|hero|illustration|icon|infographic,
   │  one-entry-one-image rule (multi-image prompts = multi-panel garbage)
   ▼
 site_plan_imagery (scope, key, kind, prompt, style_hints JSONB)
   │
   │  emit_imagery_items / check_unfulfilled_imagery_plan
   ▼
 needs_imagery work items ──▶ image-build-handler
   │                            │ composes final prompt (see §4)
   │                            ▼
   │                     image-generator ──▶ dynamic adapter
   │                            │     routes BY KIND:
   │                            │     icon/logo/illustration → Gemini image
   │                            │     photographic kinds     → SDXL
   │                            ▼
   │                     store_asset (assets row; upsert refuses
   │                            │     rows with locked_at set)
   │                            ▼
   │                     asset-deployer (spawned WITH storage env —
   │                            │     the base chassis has none, by design)
   │                            │     optimises + commits to git as
   │                            │     /assets/images/<key-with-hyphens>.<ext>
   │                            ▼
   │                     flag_page_image_rebuild → needs_page @ prio 99
   ▼                            │
 plan_sections resolver ◀───────┘   (page re-resolves through the resolver)
   │  resolves every component image field by ROLE:
   │    literal asset key → image-role alias (background, product_screenshot,
   │    image… → "hero") → page hero → site hero → placeholder
   │  emits storage.DeployedWebPath(asset_key, purpose) — the DERIVED git
   │  path, never assets.url (a presigned URL that expires in 7 days)
   │  and injects hero aliases into resolved_data, which the renderer
   │  merges LAST ("authoritative overlay") so per-page values beat
   │  legacy site-wide context fields
   ▼
 rendered page_components → compiled page → git → CDN
```

The week's three render bugs all lived in the last inch of this diagram —
images generated perfectly but components couldn't receive them: (1) field
names didn't match generated keys → the role-alias layer; (2) resolver
surfaced expiring presigned URLs → `DeployedWebPath` as the single shared
path convention between deployer and resolver; (3) component templates that
were *saved rendered output* with zero `{{vars}}` → regeneration (below).

## 4. Brand consistency as data (Phase I1)

Each site carries a `site_specs` aspect `imagery_style_guide`:

```json
{ "palette": "deep charcoal, electric blue accents (#0080FF–#00B4FF)…",
  "medium":  "industrial photography, dark atmospheric lighting…",
  "mood":    "precise, technical, engineered",
  "avoid":   "stock-photo people, generic technology abstractions…",
  "reference_asset_keys": ["hero_canonical", "hero_home"] }
```

`generate_image` composes per kind — the gating encodes hard-won lessons:

| Kind | Gets | Why |
|---|---|---|
| hero/illustration/infographic | medium + mood + palette, prepended | full brand voice |
| icon | palette ONLY | a photographic direction on an icon prompt makes the model paint a photo around the icon (observed 2026-05-20) |
| logo | nothing | generated once, human-approved, then `locked_at` set — the assets upsert refuses to overwrite locked rows |

`avoid` feeds the **negative prompt** (stronger channel than positive-prompt
pleading). Reference keys resolve to stable `s3://` URIs (presigned URLs are
stripped back to bucket/key so anchors outlive the 7-day signature) and flow
to the Gemini provider as style anchors. Charts/data graphics are explicitly
NOT diffusion work — diffusion models render the *appearance* of a chart
with fabricated values; real data gets code-rendered (go-echarts), with the
LLM restricted to the editorial layer (titles, callouts).

## 5. Case study: the self-healing loop closing on its own history

Fourteen active components fleet-wide had `html_template` saved as rendered
output — literal `<no value>` strings where every variable should be
(damage from a pre-validation era of the component generator, March–April).
The repair used only existing machinery plus one 200-line bridge:

```
 compute_component_quality            (existed: flags "0 template variables")
        │
        ▼
 check_component_template_corrupted   (NEW: the bridge — a discovery check;
        │                              cross-site guard because components
        │                              are fleet-shared; cap 5/pass)
        ▼
 needs_component_regeneration ──▶ component-creator (existed)
        │                            │  pre-store validation rejects:
        │                            │  · templates containing "<no value>"
        │                            │  · schema-field renames that would
        │                            │    strand deployed content_data
        ▼                            ▼
 verified clean template ──▶ affected pages re-render
```

Timeline: first component detected → regenerated → verified **with zero
human involvement**; 10 of 14 healed within a day; the guard now polices
the fleet permanently. One rejection along the way was the *guard working*:
the LLM dropped three schema fields, pre-store validation refused, and a
retry with explicit field-preservation instructions (rendered into the
creator's prompt via the work-item spec) passed.

## 6. Numbers (production, 2026-07-10)

| Metric | Value |
|---|---|
| Deployed sites / active agent definitions | 9 / 144 |
| Work items on the test site, all-time / complete | 1,187 / 1,031 |
| Rebuilt test site | 33 pages, 5 interactive tools, 9 news sources |
| Hero generation, prompt → git commit | ~90 s |
| Distinct per-page heroes after the resolver fix | 16 (0 expiring URLs) |
| Corrupted components found / self-healed in a day | 14 / 10 |
| Whole-site restyle (layout FK swap + CSS re-render + deploy) | ~1 h |
| New permanent discovery checks shipped this week | 2 (+1 error-honesty fix) |

## 7. Honest limitations (current)

- **Dispatch throughput**: one item at a time per site, and a stuck claim
  blocks that site until a reaper cycles — the single biggest time cost in
  verification this week (fix scoped: reaper cadence + per-type circuit
  breaker).
- **No runtime re-composition**: changing a live site's layout is a
  deliberate targeted-migration pattern today, not an agent capability.
- **Vision-quality audit not yet built**: the loop verifies images exist and
  resolve; judging that an image *looks* right (vs the brand guide) is a
  planned phase with a vision-capable auditor.

*Deep dive: PLAN_imagery_best_in_class.md (phases I0–I8),
RUNNING_NOTES_imagery_best_in_class.md (full evidence trail, Turns 1–22),
002_system_architecture.md (platform reference).*
