# PLAN 2026-08-19 — bugs_open/311: divert-on-foreign-collision in store_generated_component

Owner brief (2026-08-19, chat): fix `bugs_open/311`, prefer the robust framework-wide
solution over the individual case, diagnosis loop for research, council round, commit for
the next chassis build.

## The mechanism, as refined this session (and sent through 090)

The bug file's account is right about the deadlock and right that the two halves key on
different columns — but its step 2 is incomplete, and that changes which fixes work:

1. `plan_sections` resolves a section name by **function BEFORE the selector**: Path 1 →
   `loadSectionComponents` (`v3_site_actions.go:4366`), pass 2 matches
   `content_components.function` with **no component_level filter**. The incumbent rows
   (`mortgages-repayment` etc.) WOULD resolve there.
2. They don't because `componentInfoFromRaw` → `componentTemplateValid` →
   `sectionTemplateValid` (`plan_sections_action.go:1812`) requires a `</section>`
   substring. All three incumbents are hand-seeded (`created_from='manual'`,
   2026-08-13/15), **tool-shaped** templates ending `</script>` — measured, none contains
   `</section>` — so they are **dropped at load as "truncated"**. THAT drop, not the NULL
   `section_type`, is what routes resolution to the selector.
3. The selector then misses on `section_type IS NULL` (the bug file's step 2) →
   `needs_new_component` → the store finds the incumbent by `function` → the
   field-contract guard correctly refuses → deadlock, forever.
4. Corollary (**this kills fix candidate 2's backfill half**): backfilling
   `section_type = function` on a guard-dropped row makes `SelectComponentByType` return
   a candidate that `loadSingleComponentSchema` then drops → `resolveSectionComponent`
   returns `selector_error` → plan_sections passes the section to the content writer
   **"as-is"** (`plan_sections_action.go:1308-1321`) instead of raising
   `needs_new_component`. That converts a self-healing not_found into a silent degrade.
   For guard-PASSING rows the backfill is a no-op: if the request matched neither name
   nor function at Path 1, `section_type = function` cannot match it at Path 2 either.

090 run on this refinement: intake `1306e72c-c725-4c3b-b0c3-8a63137f35fb`, run
correlation `f1433782-6ba7-4304-a7f9-8bd830dfb7c9` (verdict recorded in NOTES when it
lands).

## Decision: fix candidate 1 (divert), candidate 2 refuted/narrowed, candidate 3 deferred

**The door being closed:** one site's component must never be able to block another
site's build, and the blocking state must be unrepresentable — not "retried harder".

### Ships this round (platform code, one coherent task)

**E1 — `platform/orchestration/actions/component_storage_identity.go` (NEW).**
`resolveStorageIdentity(ctx, db, functionName, requesterSiteID, requesterDomain, logger)`:
- Lookup A: existing base row by function (exactly today's query/order).
- No row → creation (as today).
- Row found → dependent-site census: `page_components→pages` UNION `site_components`,
  joined to `sites` for domains.
- Requester unknown (`site_id` empty: direct invocation) → today's behaviour, stated.
- Dependents ⊆ {requester} (or none) → regeneration of that row, as today (the
  field-contract guard still protects the requester's own pages).
- Foreign dependent(s) → **divert**: final function = `function + "-" + domainSlug(domain)`
  (the `deploy_tool_action.go` convention; domain from input_data, fallback
  `sites.domain` by id; if unobtainable → today's behaviour + durable finding). Lookup B
  on the suffixed name: none → creation; own-site row → regeneration of the site's own
  diverted row; foreign again (cross-site squatting tail) → refuse with a message naming
  the double collision. No double-suffixing on re-runs (HasSuffix check).

**E2 — `store_generated_component_action.go` wiring.**
- Identity resolution moves to before `separateInlineJS` (the extracted-JS
  `<script src="/tools/assets/{function}.js">` reference must carry the FINAL name).
- The diverted creation INSERTs a **base row** (`forked_from` NULL — fresh generation,
  not lineage of the incumbent; and the requesting page recovers via the selector, which
  filters `forked_from IS NULL`) with `section_type = <requested section_type>`
  (unsuffixed — the request vocabulary), which is what makes the row selector-visible
  for the requester's rebuild AND reusable by every later site that plans the same
  section type. The library heals: the first diversion creates the selector-visible
  component the incumbent never was.
- Regeneration UPDATE gains `section_type = COALESCE(section_type, $N)` — every
  successful regen repairs a NULL, so the manual-seed drift class shrinks instead of
  persisting.
- On diversion: durable `LogActionFindings` record (`COMPONENT_COLLISION_DIVERTED`)
  naming incumbent id/function, its dependent domains, requester, and the minted
  function; response map gains `diverted_from` / `requested_function`.

**E3 — `component_storage_identity_test.go` (NEW).** sqlmock, mutation-proof in the
`store_generated_component_source_guard_wiring_test.go` idiom: foreign-dependent case
must reach INSERT with the suffixed function (an UPDATE is an unexpected statement →
fail); own-site case must reach UPDATE; unknown-requester case preserves today's
refusal; UPDATE carries the section_type COALESCE.

**E4 — register entry (same commit, per the 2026-07-29 ordering-exemption ruling):**
`docs026_concept_register/register/component-lifecycle.md` — new CLC entry for the
site-aware storage-identity seam, landmine included.

### Explicitly rejected alternatives (the council will ask)

- **Regenerate-in-place with preserved field names (CLC-004 extended to function
  lookup):** would make ANOTHER site's build swap a live third-party site's hand-built
  calculator template for an LLM-generated one (markPagesPendingRebuild + rerender fan
  out to the incumbent's pages). That is the clobber class CLC-001..008 catalogue, as a
  feature. Also CLC-004's preservation machinery is dormant for exactly these rows
  (lookup keys on section_type, NULL here).
- **Fork with `forked_from = incumbent`:** hides the new row from the selector
  (`forked_from IS NULL` in every selection path), so the requesting page could never
  link it; and the lineage is false — the generation is fresh, not derived.
- **Backfill section_type / NOT NULL constraint (candidate 2):** refuted above; the
  constraint additionally cannot be added while the three tool-shaped rows violate it,
  and "backfill those three" is the harmful case.

### Deferred, recorded in the bug file as residuals

- **Candidate 3 (nothing gates a deploy on planned-sections-present).** Separate seam,
  separate coherent task. Note: the detection half partially exists —
  `discovery_checks/check_unresolved_sections.go` re-arms pages once a matching
  component exists (function OR section_type — a view WIDER than the loader's, which is
  the livelock's retry engine today and converges only once a valid component exists).
- **The three tool-shaped section-level rows** stay mis-shelved (load-invalid as
  sections; sites serve from stored rendered_html). Repair runs through the framework
  (RFC_034 bar), not hand SQL, and belongs to the incumbent's lane.
- **Latent seam:** the selector can offer a row `loadSingleComponentSchema` then drops →
  `selector_error` → section shipped "as-is". Unreachable for our rows since we do NOT
  backfill; noted for the future.

## Consumers of the seam, named (2026-07-29 ruling §3)

- `component-creator` workflow — only registered caller of the action; its failed items
  start succeeding (creation instead of refusal).
- `plan_sections`/selector — gains new base rows; that is the library's normal growth
  path.
- `unresolved_sections` discovery check — its retry loop converges instead of
  livelocking.
- The incumbent lane (loanandmortgagecalculator.co.uk) — gains a guarantee: its rows can
  no longer be regeneration targets of other sites' builds.
- RFC_034 conversion programme — convert-by-id, unaffected.
- No new action input keys (site_id/domain/source are read from input_data as the action
  already does) → no RFC_022 optional-key budget movement.

## Verification (both halves, per the bug file's own test)

After image build + roll (prove at the artefact, per service):
1. Re-drive one failed item (loanzy, `loans-credit-health-check`).
2. Assert: new row `function='loans-credit-health-check-loanzy-uk'`,
   `section_type='loans-credit-health-check'`, base, active; incumbent `824e3309`
   `md5(html_template)` UNCHANGED and its dependents' `content_data` keys untouched;
   loanzy's `tool-credit-health-check` page ends with a non-empty `component_id` for
   that slot after its rebuild.
3. Unit tests green; build from committed HEAD (`git archive`), not the shared dirty tree.
