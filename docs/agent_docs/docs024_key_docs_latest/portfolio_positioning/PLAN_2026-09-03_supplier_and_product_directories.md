# INITIAL PLAN — supplier directories with a selection of their products

**Written 2026-09-03 by the `portfolio_positioning` lane, at the owner's request, to be handed to
another thread.** He asked for this after correcting a wrong hold of mine: *"Directories will be a
frequently used go to for many sites including directories of other suppliers information and a
selection of their products — this part isn't yet built, please provide an initial plan for it."*

**This is a plan, not a design decision.** Everything measured is dated and marked. The taking
thread should re-derive any figure before quoting it — the counts move weekly.

---

## 1. What already exists, measured 2026-09-03

Directories are a working mechanism on this estate, not a gap. **165 entries across 6 kinds**, all
compiled by `directory-researcher` / `finance-directory-researcher`, most recent refresh 2026-09-01:

| kind | entries | active |
|---|---|---|
| `mortgage-lender` | 51 | 49 |
| `model` | 36 | 36 |
| `company` | 33 | 32 |
| `savings-provider` | 28 | 28 |
| `health-insurer` | 12 | 12 |
| `protocol` | 5 | 5 |

**The enablement recipe for a NEW kind** (repeated six times, so it is a recipe and not a project):

1. A `scheduled_tasks` row `"<kind>-directory-discovery"`, `target_agent_type` =
   `directory-researcher`, `interval_seconds` 604800, `input_data.research_query` carrying the
   deep-research prompt. This is the compilation step.
2. On each consuming site, `classification.content_features.<name>_directory` =
   `{kind, recommended: true, separate_page: true, reason}` — the worked example is
   `farmerinsurance.uk`'s `health_insurer_directory`.
3. Entries land in `directory_entities`; the page renders via `render_directory_action`, and
   `directory-json-exporter` / `directory_export_action` publish the machine-readable copy.

**Supporting machinery that already exists:** `directory_claims` (claims on entries),
`entity_state_log`, `directory-freshness`, `directory-build-handler`,
`directory-export-orchestrator`, `feed_directory_recommendation_action` (decides whether a site
should have one), `evaluate_directory_features`.

## 2. The actual gap — the entity model is FLAT and GLOBAL

`directory_entities` columns: `id, kind, slug, name, owner, summary, links, attributes, status,
discovered_by, created_at, updated_at`.

- **No hierarchy.** There is no `parent_id` and no child kind. A supplier cannot own products.
- **`attributes` is a small flat map**, used today for exactly one or two scalars per kind —
  `region`, `sector`, `category`, `modality`, `steward`, `status`. It is not carrying structured
  sub-records anywhere, and nothing reads it as a list.
- **`links` is a flat map** of one or two URLs (`{"docs": "..."}`).
- **Entities are GLOBAL BY KIND, not per site.** Two sites naming the same kind share the same
  entries. That is a feature for reuse and a problem for "a *selection* of their products", which
  is inherently per-site curation.
- **No product-shaped fields at all**: no price, spec, SKU, availability, currency, or
  as-of date.

## 3. Three options, with a recommendation

**Option A — products as a second entity kind with a parent reference.** Add `parent_id uuid` (or
`parent_slug`+`parent_kind`) to `directory_entities`; a product is an entity of kind
`<supplier-kind>-product`. Reuses every existing seam: the researcher, freshness, claims, export,
render. Costs one nullable column and one index.
**Recommended.** It is the smallest change that makes the model expressive, and it keeps one
compilation path rather than two.

**Option B — products nested inside the supplier's `attributes` jsonb.** No schema change. But
nothing today reads `attributes` as a list, so every consumer (render, export, claims, freshness)
needs to learn a new shape; products become invisible to `directory_claims` and to any per-product
freshness check; and a product cannot be cited independently. **Not recommended** — it buys a
migration and pays for it in every reader.

**Option C — a separate `directory_products` table.** Cleanest conceptually, largest surface: a
second researcher path, a second freshness path, a second export path, and a join that every
renderer must learn. **Only worth it if products need fields entities genuinely cannot hold** — a
price history, stock, or per-site pricing. Worth revisiting if Option A's `attributes` proves too
thin in practice.

**Per-site selection, needed by all three:** a join table (e.g. `site_directory_selection`:
`site_id, entity_id, position/rank, included`) so each site curates which suppliers and which
products it shows, without forking the shared entity. Without this, "a selection of their products"
cannot be expressed at all, and every site showing a kind shows everything in it.

## 4. Phasing a thread can actually execute

- **Phase 0 — decide the shape.** Options above. **This is architecture-scope** (see §6) and should
  not be built before that is settled.
- **Phase 1 — one supplier kind with no products**, on one site, using the existing recipe
  unchanged. Proves the vertical end to end and produces a real page. No schema change.
- **Phase 2 — the parent column plus the per-site selection table**, with products for that one
  kind. Verify at the served body, not at row counts.
- **Phase 3 — the researcher prompt for products**, which is where the quality risk lives (§5).
- **Phase 4 — generalise**: a second supplier kind on a second site, to prove the recipe rather
  than the instance.

## 5. Risks, each with a live precedent on this estate

1. **The researcher produces category-shaped non-entities.** Already happened: two
   `mortgage-lender` rows were archived as *"category-shaped entity, not a named firm"*, and a
   named-firm rule was added to `extract_claims` (migration 423). **Products are more susceptible,
   not less** — "a range of widgets" is the product-shaped equivalent. Whatever rule is written
   must be enforced in the extractor, not in the prompt.
2. **An empty listing page ships as prose about itself.** This is `bugs_open/444`, measured across
   five pages on 2026-09-02. The plan gate (migration 720) now holds a listing page whose item
   source resolves to zero — **treat that gate as the safety net and do not work around it.**
3. **Product data goes stale and prices are claims.** Anything with a number needs an as-of date
   and needs to pass the estate's claims gating. `directory-freshness` exists but has never been
   exercised on per-product data.
4. **A directory page that promises entries it does not have.** Enable the compilation BEFORE the
   page is planned, per the runbook's pre-enablement recipe.
5. **Third-party sourcing.** Compiling suppliers or products from another marketplace's listings
   raises attribution and terms-of-service questions, and a removal request has to be actionable.
   The copyonline directory (seeded from bark.com, owner's direction 2026-09-03) is the first case
   and its open questions are recorded in that site's brief rather than assumed.
6. **Entities are shared, so a bad entry is shared too.** One wrong supplier record appears on
   every site naming that kind. There is no per-site override today; the selection table in §3 is
   what would provide one.

## 6. Why this needs review before it is built

Adding `parent_id` to `directory_entities`, or a per-site selection table, **alters a shared
mechanism every site passes through** — the 2026-07-28 owner ruling's definition of
architecture-scope, and the accumulation concern `RFC_022` exists to catch. It is also the shape
`RFC_010`'s ruling addressed: new authority on a shared seam ships as an **opt-in field whose unsafe
default is OFF**, so a site that has not opted in behaves exactly as today.

**Recommendation to the taking thread:** file it as an RFC with §3's options costed, get the shape
ruled on, then build Phase 1 (which needs no schema change and can proceed in parallel).

## 7. What the owner still has to decide

- Which option (§3), or delegate that to the RFC.
- The first supplier vertical and the first site — this plan deliberately does not choose.
- Whether products carry prices at launch. Prices are the highest-value field and the highest
  claims risk, and launching without them is a legitimate first cut.
- Third-party sourcing policy (§5.5), which applies beyond copyonline.

## 8. Cross-references

`bugs_open/444` (empty listing pages; the gate) · `bugs_open/450` (tool pages, same
producer-absence family) · `RFC_010` (opt-in default-OFF on a shared seam) · `RFC_022` (optional-key
accumulation) · `RFC_037` (the classifier reading the register) ·
`portfolio_positioning/RUNBOOK_remake_release.md` §6 (pre-enablement) · migration 423 (the
named-firm rule) · `DIR-001` in the concept register.

---

## Worked instance, 2026-09-04: the copywriter directory for copyonline.co.uk — what it ACTUALLY takes

Yesterday's plan said the platform "already builds directories" and that enabling a kind was a recipe.
Measured today `[NOTES (jjj)]`: the recipe is wrong for any kind that is not AI models, financial
providers, or (via the adoption researcher) companies/protocols. Each existing kind has its OWN
prompt-hardcoded producer. Two routes exist and neither is configuration-only.

### Route A — a copywriting vertical in the per-site business-intel directory (what the planner chose)
- **Pages:** none needed — `directory-listing` + `query.business_directory` exist; BLD-028 will release
  `uk-copywriter-directory` and `get-copy-written` the moment copyonline has a `directory-export-json`
  scheduled task (vetcomparison's row is the template: domain, vertical, outputs, `business_type_ilike`).
- **Data:** a `business_intel.business_verticals` row (`copywriting`) and a COLLECTION chain. Today's
  chain (`process_area_vet_sweep.go` → `scan_discovery_candidates.go` → Companies House match/review →
  `vet-practice-verifier`) is vet-specific in code and its orchestrator `vet-pipeline-orchestrator` has
  been disabled since 2026-03-17; `collection_config` is `{}` on all three verticals, so the intended
  generalisation was never built. **Cost:** generalise or clone the sweep + verifier (Go, prompts,
  Companies House SIC filter for copywriting agencies), a roll, then a supervised first sweep. Weeks-shaped.
- **Gives:** verified businesses with addresses, Companies House numbers, `publication_optout` (the
  owner's removal question answered by a column), claimable listings. The right long-term shape for
  "directories of suppliers and their products".

### Route B — a `copywriter` kind in the global verified-claim register (this lane's Phase-B shape)
- **Producer (config only):** a researcher seed cloned from
  `finance_directory_pipeline/SEED_finance_directory_researcher_agent.sql` (same six steps) with an
  extraction prompt for UK copywriting ORGANISATIONS emitting `entity_kind: "copywriter"` and a stated
  field vocabulary (specialisms, sectors, supplier_type marketplace|direct, website, established_year,
  location). Registration accepts unknown kinds with an open vocabulary (`directory_claims.go`:
  "Kinds absent from this map … are unaffected") — **no Go for the data**.
- **Task:** retarget `copywriter-directory-discovery` to the new agent; keep the query **under 200
  characters** (the `web_search` cap — LANDMINES).
- **Pages (Go, roll):** two `queryresolve` arms (`copywriter_directory`, `_full`) — the map is
  hardcoded per kind — and a `directoryCheckProfiles` entry so BLD-028 and the opt-in checker know the
  kind; a components seed (snippet + listing pair, the finance generator's shape); the site's
  `classification.content_features.copywriter_directory` opt-in (which the classifier rewrites — set it
  and expect to re-set it, or make the planner read it from strategy).
- **Replan:** the two held pages are typed for `business_directory`; a `needs_site_plan`
  (`manual-replan`) after the kind exists lets the planner type them to the register source.
- **Cost:** ~a day: three seeds + one small Go change + council + roll + a supervised first run.
- **Gives:** cited, re-verified facts per organisation (the model/finance shape); randomised listing is a
  render option; no addresses/Companies House unless the prompt asks.

### Recommendation
**B first**, because it unblocks the two pages within a day using three worked precedents, and it is
the shape the owner described ("a directory… randomised listing… point leads to listed companies"). A
is the better destination for the wider "suppliers and their products" programme and should be planned
as its own lane once the vet chain's generalisation is scoped. **Neither is to be started without the
owner's word** — both add a shared producer and one adds Go.
