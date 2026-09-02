# PLAN 2026-09-02 — bugs_open/443: per-section subjects (and facts) for pages the plan tables do not serve

**Lane:** `bugfix_443_fallback_tier_subjects` (this session, "bugs_open/443"). **Bug:**
`bugs_open/443_HANDOFF_2026-09-02_a_page_built_from_the_fallback_tier_cannot_carry_per_section_subjects_so_repeated_component_types_write_the_same_section.md`
— the shared account; measurements land there too. Filed by the finetuning_uk_service lane,
handed off, resumed here after confirming no fix was in flight (no dirty/recent commits on the
deciding file; the filing lane's handoff says "Filed as bugs_open/443").

## What we are fixing (one paragraph)

`load_page_sections_from_spec_action.go` publishes `section_subjects` and `section_facts` only
when tier 1 (`site_plan_sections`) served the layout. On a site with no current `site_plans`
row every page resolves at tier 2/3/4, so per-section scoping is structurally unreachable, and
every repeated component type on a page receives the identical page-level brief and writes the
same section. `[MEASURED 2026-09-02]` 6 real sites / 203 deployed pages are plan-less; **11
pages repeat a component type, and all 11 serve real subject repetition (at least 8 with
verbatim-identical h2s; curled with per-domain invented-URL 404 controls). All 11 resolve at
tier 3.** Non-adjacent repeats fire too (finetuning `our-position-on-ai`: two blocks separated
by a `features` block) — the trigger is "same type more than once with nothing to distinguish
the instances", not adjacency or count.

## Decisions and their reasons

**D1 — publish scoping from a fallback tier only where alignment is a FACT, not a guess.**
The tier gate's reasoning is sound (index-aligning a subject list against a *different* tier's
section list is a guess); the fix keeps the rule and adds constructional sources:
- **Tier 3 (`pages_table`) — the load-bearing arm** (all 11 damaged pages resolve here): two
  new nullable jsonb columns, `pages.section_subjects` and `pages.section_facts`, read in the
  same query as `pages.sections`, applied only when aligned with the served list.
- **Tier 2 (`site_specs` aspect) — near-free future-proofing, severable if the council
  objects:** read `section_subjects`/`section_facts` sibling keys from the *same page object*
  the sections came from. `validate_plan`'s normalise pass already emits exactly these keys
  page-level. `[MEASURED 2026-09-02]` 0 current aspects carry them, 0 damaged pages resolve
  here — inert on arrival, aligned by same-object construction when it ever fires.
- **Tier 4 (`same_role_sibling`) — never.** A borrowed skeleton's subjects would be another
  page's. Structurally cannot populate the arrays; comment says so.

**D2 — guard on the length invariant, not the source label.** The emission and LOCK-008
merge-nil-insertion guards change from `specSource == "site_plan_tables" && len==len` to
`len==len` alone. The length equality *is* the invariant; the source check was a proxy for
"alignment is constructional", which D1 now enforces at fill time. Tier 4 leaves the arrays
nil so the length guard is never satisfiable there. Extra tier-3 guard: when the sections list
came from `page_sections_fallback` (collected_data) rather than the row, arrays attach only if
the row's stored `sections` content-equals the served list — a same-length different list must
not pass, so equality, not length, is the test on that sub-path.

**D3 — stored misalignment is kept-but-inert-and-visible, never applied, never destroyed.**
`pages.sections` has **19 candidate writer files as of 2026-09-02** (`grep -rln "UPDATE
pages\|INSERT INTO pages" … | xargs grep -l sections`); any of them can rewrite sections
without knowing the sibling columns exist. Options considered:
- CHECK constraint → REJECTED: every such writer's UPDATE would start erroring (blast radius =
  all 19 files).
- BEFORE UPDATE trigger nulling misaligned arrays → REJECTED: silently destroys
  operator-written data on any sections rewrite; keep-but-ignore lets the operator re-align
  instead of re-type.
- Object-form entries inside `pages.sections` (`{"name","subject"}` — alignment travels inside
  the entry, the normalise pass's own argument) → REJECTED: everything downstream of
  validate_plan expects plain string sections (its own comment), and `jsonb_array_elements_text`
  consumers error on objects; that is a shared-column shape change with a fleet of readers.
- **CHOSEN:** read-time guard (D2) + WARN log when stored arrays exist but do not align + the
  D4 detector firing on the resulting subjectless repeat. The 5b cache sync additionally
  *repairs* alignment in the one place we control: when the locked-row merge changes a tier-3
  list whose arrays were served this run, the same UPDATE writes the nil-padded arrays back
  beside the merged sections.

**D4 — a build-side observe-only detector, mirroring the planner-side one.** In
`plan_sections`, after the site-level triple filter: any component name appearing ≥2 times
with ≥1 of its instances subjectless files ONE durable `agent_error_log` finding,
`REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT`, via the action's existing channel. Distinct code
from the planner's `SUBJECT_MISSING_ON_REPEATED_COMPONENT` because the remedy differs (planner:
"replan / write subject rows"; build: "this page's serving tier carries no subjects — see
bugs_open/443 for how to give it some"), and the registry entry names its writer precisely.
Registered in `finding_code_registry.json` in the same commit (358 discipline), disposition
`human-evidence`. Not gated on "plan is playing the subject game" (the planner gate's job is
retro-spam protection; here the point is precisely the pages that cannot play). Bounded:
`[MEASURED 2026-09-02]` 25 repeat-layout pages fleet-wide (14 on plan-carrying sites, 11
plan-less), so worst case is one row per rebuild of one of 25 pages. This converts the silent
predicted failure into a signal for every cause, including tiers 2/4 and causes not yet found.

**D5 — no config migration, no new input keys.** `[MEASURED 2026-09-02]` the live
`page-build-handler.plan_sections` config already wires both `section_facts` and
`section_subjects` from `spec_sections.*` (639 applied 2026-09-02 by apis.uk, pod-verified).
The action input specs are unchanged, so the RFC_022 optional-key budget is untouched.

**D6 — end-to-end effect rides seed 641, and that is unchanged by this fix.** The writer
prompt is v4; v5 (641_HOLD, renders the subject block) is gated on a fresh owner read and NOT
applied. Until it lands, `sectionPlanItem.Subject` is stamped and inert at the writer — for
tier 1 exactly as for these tiers. Verification is therefore staged (below). This fix creates
no new dependency; it joins the fallback tiers to the same one tier 1 already has.

**D7 — the geometry finding goes to the owner as an RFC, not into this commit.** "The plan
tables are becoming the tier where capability lives, and 6 real sites are not in them"
(subjects, facts, and `site_plan_imagery`/IMG-078 all gated on a `site_plans` row). The
convergence alternative — materialise a minimal current plan from live pages — would fix the
class but is a programme: `reconcile_site_plan` emits rebuild items against plan pages, so
giving 6 sites plans touches the birth/rebuild path of 203 deployed pages and must be measured
per consumer (**21 files consume current-plan existence as of 2026-09-02**). RFC_063 files the
question with both options costed; owner decides. This fix is correct under either answer:
sibling-array scoping remains valid for any page a plan does not name.

**D8 — backfill is lane work, not this commit.** Subjects are planner/operator *direction*;
the briefs for finetuning's pages already name one subject per section. RUNBOOK carries the
template UPDATE. The finetuning lane chose `your-own-model` (verbatim ×3) as the acceptance
canary; their playground stays untouched as the demonstrating case.

## Scope ruling check (RFC_022 / 2026-07-29 / 2026-08-02)

Opt-in field, unsafe default OFF (absent columns = byte-identical behaviour), zero live
consumers name it (`[MEASURED 2026-09-02]` no code writes `pages.section_subjects`; the only
readers are this change's own arms) → not architecture-scope by the RFC_022 test. Does it
change what the shared mechanism GUARANTEES? The loader's stated rule — "aligned or absent,
never guessed" — is preserved verbatim; what changes is which sources can *satisfy* it.
Registered in the concept register (PBP-051) in the shipping commit (ordering-exemption
condition 2). Consumers told (2026-07-29 §3): apis.uk (PBP-049 owner), finetuning lane,
copy_quality_two_stage (their experiment preconditions on plan-assigned subjects; this adds a
second source), bugs_open/114 lane (shares the tier geometry; RFC_063 is theirs to weigh in
on).

## The edits (council plan, ≤8)

1. `docs/agent_docs/sql_for_agents/717_pages_section_subjects_and_facts.sql` — two nullable
   jsonb columns + COMMENTs stating the alignment semantics + DO/RAISE verify.
   Column-before-binary, same order 638 stated and for the same reason (the loader SELECTs
   them).
2. `platform/orchestration/actions/load_page_sections_from_spec_action.go` — tier-2 sibling
   keys; tier-3 row read (both sub-paths, content-equality on the collected_data one); D2
   guard restructure; 5b sync array repair; WARN on stored-but-misaligned.
3. `platform/orchestration/actions/plan_sections_action.go` — D4 detector.
4. `docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json` —
   register `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT`.
5. `platform/orchestration/actions/load_page_sections_fallback_subjects_test.go` (new) —
   aligned/misaligned/absent arms per tier, collected_data-mismatch arm, merge nil-insertion
   on tier 3, tier-4 never-emits, detector fires/quiet; mutation-prove the guards (drop the
   equality check → a test must fail; see memory: a mock's bookkeeping cannot assert a
   negative).
6. `docs/agent_docs/docs026_concept_register/register/page-build-pipeline.md` + index — PBP-051.
7. `bugs_open/443_…md` — fix section, the 11/11 measurement, staged verification.
8. `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` — "a writer replacing
   `pages.sections` silently disarms that page's stored subjects/facts" (+
   `landmines-verify-dispatch.sh` after).

Outside the council plan (prose): RFC_063, this lane's standing docs, WRONG_CALLS if earned.

## Verification (staged, controls named)

- **Stage A (after image roll, before 641):** write subjects for `your-own-model` via RUNBOOK;
  rebuild; assert orchestration `sections_ready[].subject` populated for the three
  `generic-text-block` slots and the D4 detector stays quiet on that page (demand control: a
  subjectless repeat page must fire it — that is the detector's own test). Served h2s may
  still repeat — the writer is v4-blind; that is 641's half, not ours.
- **Stage B (after 641 lands for the fleet):** rebuild `your-own-model`; assert the served
  h2s are DISTINCT. Control: a tier-1 page with plan subjects must pass the same assertion,
  before and after this change (the bug file's §6 control).
- **Negative control at every stage:** a plan-less page with no stored arrays must build
  byte-identically to today (unset = today's behaviour, the PBP-049 bar).

## Council

Submit via 097 before/alongside the commit; commit with `Council-Submitted: <corr>` the moment
coherent (the 2026-07-30 trailer rule); budget ~30 min queue latency; find the run by payload
not printed id.
