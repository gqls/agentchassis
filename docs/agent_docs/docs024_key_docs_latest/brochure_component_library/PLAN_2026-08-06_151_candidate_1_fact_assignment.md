# PLAN 2026-08-06 — 151 candidate 1: facts assigned to sections at plan time

Owning lane: `brochure_component_library`. Bug: `bugs_open/151` (the per-section
content writer has no memory of facts already used elsewhere on the site).
Candidate 3 (post-build census, flag-only) is LIVE and has measured the population;
this is candidate 1, the structural fix the census's own `why_no_handler` names.

## The mechanism, in one paragraph

The planner (`build-site-planner`, author of 13 of 14 current plans) is shown the
site's verified fact roster and asked to say, per section, which facts that section
is responsible for stating. The assignment is persisted on the plan
(`site_plan_sections.assigned_fact_ids`, new nullable jsonb column). At build time
the per-section writer call receives a writer block composed from ONLY that
section's assigned facts, instead of today's whole-site block injected identically
into every call. A section that is never told a fact exists cannot restate it.
Absent assignment (NULL) at any point in the chain = today's behaviour, unchanged.

## Grounding (all verified this session, 2026-08-06)

- Injection point: writer prompt v3 lines 68–72 interpolate
  `{{.site_specs.specs.evidence_base.writer_block}}` — one site-scoped string,
  identical for every section; pure template substitution, no Go reader on the
  build path.
- The planner already RECEIVES `site_specs` (`plan_site` input_fields) — its
  blindness is prompt-level only. Neither planner's live config references
  `evidence_base` (measured on `agent_definitions`).
- `extractSectionNames` (`write_site_plan_action.go:723`) has exactly one caller
  and already tolerates object-form section entries (takes the name, drops other
  keys) — so a new-prompt/old-binary combination degrades to today's behaviour.
- `validate_plan`'s two section loops pass objects through UNVALIDATED
  (`v3_site_actions.go:3088-3091`, `:3131-3134`) — so object-form entries must be
  taught to those loops, and the plan must be re-normalised to string `sections`
  before it leaves `validate_plan`, because `SyncPagesToDBAction` serialises the
  raw array into `pages.sections` (`site_db_actions.go:1109-1116`), whose every
  reader expects strings.
- The only structured per-section payload reaching the writer is
  `section_plan.sections_ready` (`sectionPlanItem`, `plan_sections_action.go:767-790`),
  looped as `current_section`. The channel from plan rows to `plan_sections` is
  strings-only (`spec_sections.sections`).
- `plan_sections_action.go` is in the same package as `composeWriterBlock`
  (`refresh_evidence_base_action.go:592`) — the filtered block can be composed
  there and the writer-side change becomes template-only.
- Fact IDs are stable, human-readable strings (`F1-live-sites`…); only facts with
  a `writer_line` ever reach the writer (`composeWriterBlock` skips the rest).
- Population (live gap specs, 08-06): fundamentallyai 9 fact-overlap pairs
  (pool 15); leopardess pool 18, 0 overlap pairs; five sites are
  `fact_census_blind` (pool < 6) and their residue is TEXTUAL near-duplication.
- Timing: 3 of 14 sites had an evidence_base on plan day; 6 have none at all.

## What this does and does not fix (stated up front, for the council and the owner)

- FIXES: independent restatement of the same verified facts across sections/pages
  — the class 151 was filed on (fundamentallyai's 18 sections / 9 facts).
- Does NOT fix: textual near-duplication on fact-poor sites (five of the seven
  flagged sites). Same root cause (per-section isolation), different content —
  the assignment mechanism gives the planner a differentiation vocabulary only
  where facts exist. The residue census keeps measuring both halves.
- Does NOT rewrite anything existing: this changes what NEW builds and REPLANS
  generate. Draining the seven flagged sites means replanning/regenerating them
  through the normal build path (claims gate at save time applies); that is
  sequenced work, gated today by 189's unapplied config half on locked-row pages.

## Design decisions (with the reason, per the working-docs rules)

1. **Facts travel INSIDE the section entry** (`{"name": "hero", "facts": ["F1"]}`),
   not in a positionally-keyed sibling map (the `imagery.sections` precedent).
   Reason: `validate_plan` drops/strips section entries, which shifts ordering; a
   positional key goes silently misaligned (the imagery scheme has this latent
   defect today). Intrinsic carriage survives every transformation.
2. **Normalise ONCE, in `validate_plan`, after its last section transformation.**
   Object entries are resolved/stripped by their `name`, then split into canonical
   string `sections` + per-page `section_facts` (aligned arrays). Reason:
   everything downstream — `sync_pages_to_db`, reconcile snaps, `pages.sections`
   readers — keeps today's shapes; exactly one place understands both forms.
3. **NULL vs `[]` are distinct.** NULL `assigned_fact_ids` = unscoped, site-wide
   block (today). `[]` = deliberately factless section: the prompt says "no
   verified facts are assigned to this section — state no business numbers or
   named-entity claims". Reason: without the distinction, an explicitly-empty
   assignment would silently fall back to the whole-site block, inverting the
   planner's decision.
4. **The filtered block is composed in `plan_sections`, attached to
   `sectionPlanItem`** (`assigned_fact_ids`, `assigned_writer_block`,
   `facts_scoped`). Reason: it is the only step that builds the per-section items
   the writer loops over; it has DB access, site identity, and package-level
   access to `composeWriterBlock`. The writer's own workflow gains no step and no
   DB read (the `extractSiteID` landmine on that path stays untouched).
5. **The facts channel into `plan_sections` is an explicit config field**
   (`section_facts` path in the step config, wired by seed), default absent =
   feature off. Reason: owner ruling 2026-08-02 — new behaviour on a shared seam
   ships as an opt-in field, not a convention. Also gives provenance: facts are
   only trusted when they came from the same source that supplied the section
   list (`site_plan_tables`), never index-aligned across fallback sources.
6. **Hallucinated fact IDs are inert, not fatal.** `plan_sections` composes from
   the CURRENT evidence_base filtered by assigned IDs; an unknown ID matches
   nothing and is logged. Values are substituted at compose time, so a re-verified
   fact's number is always current. A fact deleted later leaves a dangling
   (inert) ID; a fact added later is stated nowhere until a replan assigns it —
   facts enter copy where planned, not wherever the next rewrite happens.
7. **Rewrite paths inherit scoping only via their items.** A regeneration whose
   `current_section` carries no `facts_scoped` gets the site-wide block — today's
   behaviour. The door closes at plan-time builds; post-hoc rewrite paths are out
   of scope here and stay behind the claims gate.

## The edits (8)

1. **Migration `sql_for_agents/3XX_site_plan_sections_assigned_fact_ids.sql`**
   (number = highest+1 at commit time): `ALTER TABLE site_plan_sections ADD COLUMN
   IF NOT EXISTS assigned_fact_ids jsonb;` — nullable, additive, ordering note +
   `information_schema` type assertion + `COMMENT ON COLUMN`, on the seed-263
   pattern. Column must exist before the binary that writes it.
2. **`write_site_plan_action.go`**: widen the (single-caller) section extraction
   to return `{name, facts}` pairs from the page's `section_facts` key; bind
   `assigned_fact_ids` in the INSERT (NULL when absent). Tests: string-form,
   object-form, mixed, NULL-vs-empty.
3. **`v3_site_actions.go` (`ValidateSitePlanAction`)**: chrome-strip and
   name-resolution loops learn object entries (operate on the object's `name`);
   final normalise pass splits object entries into string `sections` +
   `section_facts` per page. Tests: objects resolved/stripped/dropped correctly,
   output sections always strings, alignment preserved across drops.
4. **`load_page_sections_from_spec_action.go`**: tier-1 query also selects
   `assigned_fact_ids`; output gains `section_facts` (aligned array; emitted only
   by the `site_plan_tables` tier — all fallbacks yield no facts). Consumers of
   `spec_sections.sections` are unaffected (additive key).
5. **`plan_sections_action.go`** (+ a small same-package helper): optional config
   field reads the aligned facts; on each ready item set `assigned_fact_ids`,
   `facts_scoped`, and `assigned_writer_block` composed via `composeWriterBlock`
   over the evidence_base filtered to the assignment. Absent config/facts = no
   new fields. Tests: scoped, empty-scoped, unscoped, unknown-ID inert.
6. **Planner prompt seed (`build-site-planner`)**: show the fact roster (IDs +
   claims, from `site_specs.specs.evidence_base.facts`, template-guarded for
   absence) and instruct: sections may be `{"name", "facts"}` objects; spread
   facts — a fact belongs to the section whose job is to state it; sections
   sharing 3+ facts is the defect being prevented. Anchored `replace()` +
   exactly-once pre-check + backup table + verify, on the seed-206 pattern.
7. **Writer-side seeds**: (a) page-build-handler wrapper step config gains the
   `section_facts` wiring; (b) writer `prompt_template` v4 — a `facts_scoped`
   branch: assigned block / explicit factless instruction / site-wide fallback,
   applied by the established base64 round-trip with backup + em-dash count
   verify. Image-first gate stated in both headers (config half applies only
   after the Go half is pod-proven).
8. **Register entry (`register/page-build-pipeline.md`, PBP-next)** in the SAME
   commit as the seam (ordering-exemption condition 2): the assignment contract,
   NULL-vs-`[]` semantics, the named consumers (below), the landmine (a scoped
   section's block is composed at build time from CURRENT facts — the assignment
   pins WHICH, never WHAT VALUE), and the open question (drain sequencing for the
   seven flagged sites).

## Consumers of the shared mechanisms touched (named, per the 07-29 ruling)

- `site_plan_sections` readers: `load_page_sections_from_spec` (build),
  `reconcile_site_plan`, `PlanSpecifiedSectionCounts` (dedup guard trio),
  `page_section_satisfiability`, `check_section_source_drift`,
  `check_sectionless_pages`. All read named columns; an added nullable column is
  invisible to them. None gains new behaviour.
- `sectionPlanItem` consumers: the writer's `process_sections_loop` /
  `generate_content` template (additive keys; `missingkey=zero` renders absent
  fields falsy in `{{if}}`), and `persistSectionSkips`. Additive keys only.
- `evidence_base` consumers: writer prompt (site-wide block — retained as
  fallback), claims gate (`numberSupported` — unchanged; scoping REDUCES what a
  section asserts, the gate still validates whatever is asserted), census check
  `assignFacts` (unchanged; it measures the outcome this mechanism shapes).
- `plan_site` output consumers: `validate_plan` (edited), then everything
  downstream sees normalised string sections exactly as today.

## Rollout order

1. Commit code + migration + seeds + register (one commit, this task); submit to
   council before/alongside (`Council-Submitted` trailer).
2. Apply the migration (column first — inert).
3. Build + roll chassis; pod-grep added strings with an invariant control.
4. Apply seeds: 065 wiring + writer prompt v4 (inert until items carry facts),
   then the planner prompt (first plan with assignments now possible).
5. Acceptance: replan ONE fact-rich site (fundamentallyai, pool 15, 9 overlap
   pairs — the motivating case), rebuild through the normal path, then re-run the
   census: the fact-overlap count must FALL (9 → target 0 pairs ≥3 shared); the
   five fact-blind sites are expected NOT to move (stated so the check is
   disconfirmable). NOTE: do not fire builds at locked-row pages until 189's
   config half is applied (bug_backlog_clearing lane owns that sequencing).

## Open questions for the council / owner

- Should `validate_plan` also validate assigned IDs against the live
  evidence_base (DB read) and drop unknowns at plan time, or is compose-time
  inertness + logging enough? (Current position: inertness is enough; a plan-time
  read adds a failure mode to validation for a defect that is self-healing.)
- Drain sequencing for the seven flagged sites (replan + rebuild), pending 189
  config and owner appetite for replans on live sites.
