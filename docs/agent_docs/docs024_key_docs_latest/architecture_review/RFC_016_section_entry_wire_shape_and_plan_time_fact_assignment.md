# RFC 016 — the "section entry" wire shape, and plan-time fact assignment as its first structured field

**Raised by:** `brochure_component_library` (bug 151 candidate 1), 2026-08-06, on the
council's instruction — corr `902a8563-2200-4771-ac0f-55dab0839a02`, REJECTED, hard
guardian veto round 1, with the `architecture` seat independently returning
`ARCHITECTURE_SIGNAL: needs_rfc`. Per the owner rulings of 2026-07-28/29: the code
stays (committed `b882d5abf`; review here is after the fact by design), a scope veto
is answered by recording it where the change lives and routing the seam here — not
by resubmitting with better measurements.

**Status: OPEN — needs a human eye on §5 (the two decisions) before Slice B applies.**

## 1. The contract this RFC writes down

A **section entry** — one element of a page's planned `sections` array — has until
now been a bare component-name string everywhere, held together by convention across
at least these carriers:

| carrier | shape | reader assumptions |
|---|---|---|
| planner LLM output (`plan_site` → `validate_plan`) | historically `string`, object form silently tolerated | `extractSectionNames` took name-ish keys, dropped the rest; validate_plan's chrome-strip and name-resolution loops SKIPPED objects entirely ("brief objects pass through") |
| `site_plan_sections` rows | one row per entry, `component_name` text | authoritative store; the build reads it (pages.sections is a cache) |
| `pages.sections` jsonb | array of strings | read fleet-wide via `jsonb_array_elements_text` and equivalents |
| `spec_sections` (in-flight, `load_page_sections_from_spec` output) | `[]string` | `plan_sections` consumes via config dot-path |
| `sectionPlanItem` / `sections_ready` | structured Go struct → JSON map | writer's loop var `current_section`; template reads under `missingkey=zero` |

**The change `b882d5abf` made** (plus the `v3_site_actions.go` half that reached HEAD
inside `cb7b4d759`): the planner MAY emit `{"name": "...", "facts": ["F1-..."]}`;
`validate_plan` resolves/strips object entries by their `name` and then **normalises
before the plan leaves the step** — `sections` reverts to plain strings, assignments
move to an aligned per-page `section_facts` array — so every carrier below
validate_plan keeps its historical shape. The assignment persists as
`site_plan_sections.assigned_fact_ids` (nullable jsonb; migration 327), flows as an
additive `section_facts` key on `spec_sections` (authoritative tier only), and lands
as three additive fields on `sectionPlanItem` (`facts_scoped`, `assigned_fact_ids`,
`assigned_writer_block`) behind an opt-in step-config key.

**The rule for the next structured field** (imagery variants, testimonial bindings —
the architecture seat expects more): object-form entries live ONLY between the
planner LLM and validate_plan's normalise pass. Everything downstream reads the
string array plus a named, aligned, per-page sibling key. Alignment is intrinsic
(the field travels inside the entry through the LLM leg; positional lookup only ever
against the RAW array, before skips) — the `imagery.sections` `"<page>:<ordering>"`
scheme is the counter-example: **an entry dropped by validate_plan after those keys
are written mis-keys every later ordering silently.** That latent defect in the
imagery scheme is real today and deserves its own bug file.

### 1a. Scope clarification (owner question, 2026-08-08): what "never positional-from-outside" does and does not forbid

The rule targets one specific shape: **a positional cross-reference that is
STORED — written by counting into a sibling structure, persisted separately,
and consumed later, after either side may have changed.** That is the imagery
counter-example (`bugs_open/214`): "about, section 4" recorded at plan time,
wrong by the time anything reads it, and silently so.

It does NOT forbid:

- **Addressing by stable identity from outside.** Editing a particular image
  means naming it — `asset_key` (`icon_fine_tuning`), the plan-imagery `key`,
  a page name, a work-item key. Names survive reorders; ordinals do not.
  Every existing edit path already works this way (`needs_imagery` items carry
  the key; hero lookup goes through `imageryplan.ContentHeroKey(page)`).
- **Ephemeral positional language resolved immediately.** A human telling an
  operator flow "change the second image on the about page" is fine — the
  position is resolved to an identity in the same breath, against the live
  artefact, and the IDENTITY is what gets stored or acted on. Position used as
  a pointing gesture is harmless; position used as an ADDRESS THAT OUTLIVES
  the state it counted is the defect.
- **Alignment INSIDE one atomically-written structure** — the `section_facts`
  sibling array is index-aligned, but it is derived and written in the same
  pass that finalises the section list, and nothing re-derives one without the
  other. Alignment is safe exactly as long as both sides live and die together.

The test for a new design: *if someone reorders, drops or renames an entry on
one side, does the other side follow automatically or lie silently?* If it can
lie, it needed a name.

## 2. Blast radius (measured, not asserted)

- `site_plan_sections` readers: `load_page_sections_from_spec`, `reconcile_site_plan`,
  `PlanSpecifiedSectionCounts` (three callers: dedup repair, census, save collapse),
  `page_section_satisfiability`, `check_section_source_drift`,
  `check_sectionless_pages` — all read named columns; the added nullable column is
  invisible to each.
- `pages.sections`: unchanged shape by construction (the normalise pass exists
  because `SyncPagesToDBAction` serialises the raw array — `site_db_actions.go:1109-1116`).
- `sectionPlanItem`: additive keys under `missingkey=zero`. The bug_historian seat's
  standing concern applies and is fenced as far as this change can fence it: the
  three-way writer-prompt branch means absence falls back to today's block, and the
  ambiguous case (non-empty assignment composing an empty block) is now degraded to
  UNSCOPED with a durable `agent_error_log` row (`FACT_SCOPING_EMPTY_COMPOSITION`)
  rather than rendering the deliberate-factless branch — fixed post-verdict, with a
  pinning test.
- Prompts: `build-site-planner` (13 of 14 current plans) and `page-content-writer`
  (every page build) — which is precisely the breadth the guardian vetoed shipping
  as one slice.

## 3. The veto, and the rollout it dictates (SUPERSEDES the order in the 151 bug file)

Guardian, verbatim intent: split into two independently reviewable slices.

- **Slice A — planner emits, nothing consumes.** Image roll (already-committed Go)
  → migration `327` → seed `329` only. Plans start carrying `assigned_fact_ids`;
  no writer behaviour changes anywhere. Observe real plan output
  (`SELECT page_name, ordering, component_name, assigned_fact_ids FROM
  site_plan_sections WHERE assigned_fact_ids IS NOT NULL`) — are the planner's
  assignments sane, spread, complete?
- **Slice B — consumption.** Seeds `328` + `330`, submitted to the gate as its own
  round once Slice A's output has been inspected on live data, and piloted on a
  named small cohort if practicable (per-site/step config), not a fleet-wide swap
  on day one. **Before 330 applies, a human/compliance pass reads the v4 plaintext**
  (`brochure_component_library/sql/page_content_writer_prompt_v4_2026-08-06.txt`) —
  the compliance seat's ask; the text is committed for exactly that purpose.

### 3a. Slice A observation, 2026-08-07 — the entry evidence the Slice B round must cite

Replan of fundamentallyai (corr `801b0732`, new plan `8ee5807b`, 71 sections /
21 pages; full evidence trail in the lane NOTES, same date). Answers to §3's
three questions:

- **Sane: YES.** Both assignments topically exact (`F7-idea-stripe` → the
  backend-engineering page's prose section; `F8-private-search` → the
  search-infrastructure page's), none on hero/CTA roles, `[]` used
  meaningfully on deliberately factless sections. Tri-state semantics arrived
  intact; 71 emitted entries persisted as exactly 71 rows (zero
  validate_plan drops); `pages.sections` stayed strings. The consumption
  negatives (no `facts_scoped`/`assigned_writer_block` in the live writer, no
  `section_facts` in page-build-handler) were re-verified while the replan's
  own builds ran — the no-op guarantee was exercised, not assumed.
- **Spread: YES, trivially** — no fact assigned twice, so no sharing to judge.
- **Complete: NO.** Object-form on 5 of 24 pages, all pages the planner was
  composing fresh; 2 of the 9 rostered facts assigned. Every carried-over page
  — including index/capabilities/about, which hold the 9 fact-overlap pairs
  that motivated 151 — emitted plain strings, i.e. NULL/unscoped. (Roster note:
  9 of 15 pool facts offered is the deliberate `{{if .writer_line}}` filter,
  mirroring `composeWriterBlock`; the other 6 are chart-only facts.)

**Consequence for the Slice B round:** against this plan, consumption changes
writer behaviour on 5 pages only, and the overlap pairs live on the unscoped
pages — so the acceptance as stated in the 151 bug file (census pair-count
falls on rebuild) would NOT move. The round must therefore choose, and say
which: **(a)** strengthen rule 17 so object-form entries are required for
every page (a fleet-wide planner-prompt change — reviewable in the same round,
it is the same prompt seed 329 already touched), or **(b)** keep the prompt and
re-scope the acceptance to engaged pages, accepting that coverage of legacy
pages arrives replan-by-replan. Option (a) is what the motivating case needs;
option (b) is honest about cost but leaves 151's headline symptom standing on
fundamentallyai.

Side-finding, promoted: §1's imagery `"<page>:<ordering>"` counter-example is
not merely latent — a fleet census the same morning found 5 of 131
section-scope refs orphaned (one minted by this very replan), four of them
with paid-for active assets unreachable by any build. Filed as
`bugs_open/214` with the census query and fix candidates.

## 4. Objections acted on / answered (so the next round is not re-litigated)

- **bug_historian, empty-composition ambiguity: FIXED in code** (degrade + durable
  record + test), see §2.
- **bug_historian, unknown-ID-only-logged: FIXED by the same change** — the
  all-unknown case now writes `agent_error_log`; a PARTIALLY unknown assignment
  still composes the known subset and logs the misses (deliberate: one stale ID
  should not silence a section's remaining facts; the census still measures the
  outcome).
- **prior_art_librarian, verify the absence claims:** re-verified with the
  LIKE-underscore landmine in mind. `evidence_base` contains `_` which LIKE treats
  as a wildcard — that loosens the pattern, so a ZERO under it is a *stronger*
  zero, and the claim stands: neither planner's `default_config` references
  evidence_base in any spelling. `extractSectionNames` single caller: grep over
  `platform/`, one call site (`write_site_plan_action.go:313`), test files aside.
- **editquality, "mutation tests not in the plan":** they exist and ran before the
  submission (`fact_scoping_151_test.go`, both mutations caught) — a plan-authoring
  omission, not a code one.
- **guidelines, is `current_section` a declared contract?** Internal step-to-step
  flow within one agent's workflow (plan_sections → loop → generate_content), not a
  `call_agent` boundary; treated as internal, this RFC is where the shape is now
  written down regardless.

## 5. The two decisions this RFC actually asks for — ALL DECIDED (owner, 2026-08-08)

1. **Ratify the section-entry rule in §1** (object form is a validate_plan-internal
   transient; downstream carriers keep historical shapes; aligned sibling keys, never
   positional-from-outside) as the standing contract for the NEXT structured
   per-section field — so the next lane extends this instead of inventing a rival
   carrier. **DECIDED: RATIFIED (owner, 2026-08-08).** With one scope
   clarification the owner's own question surfaced, now recorded in §1a below:
   the rule bans STORED positional cross-references between separately-stored
   structures; it does not ban addressing a thing by its stable NAME or id from
   outside (asset_key, page name, work-item key), nor a human saying "the
   second image" to an operator flow that resolves it to identity immediately
   against the live artefact.
2. **Approve the Slice A / Slice B order in §3**, including the pilot-cohort and
   the human read of the v4 prompt before Slice B. **DECIDED: APPROVED (owner,
   2026-08-08).** The v4 plaintext read remains an open ACTION (owner or
   delegate) gating seed 330's apply — approval of the process is not the read.
3. **§3a's follow-on: DECIDED for option (a)** (owner, 2026-08-08) — strengthen
   rule 17 so the planner must state fact ownership for EVERY page, with the
   recommended safeguard: one more observed replan before the consuming half
   applies. Shipped as seed `331` in the same Slice B council round.

## 6. Sources

`b882d5abf` (the change; PBP-037 in the same commit) · `cb7b4d759` (the
v3_site_actions.go half, swept as a same-file passenger) · council report:
`diagnosis_artifacts` kind=`council_report`, corr `902a8563-2200-4771-ac0f-55dab0839a02`
(11 seats: 6 approve, 3 object, guardian veto, architecture needs_rfc) ·
`bugs_open/151` · `brochure_component_library/PLAN_2026-08-06_151_candidate_1_fact_assignment.md`
