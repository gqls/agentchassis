# PLAN 2026-08-21 — teach the remaining call sites to see a stored slot identity

Design, phasing, decisions **and their reasons**. Corrections to the originating
brief live here, marked as corrections.

Drafted with `fable` from a brief carrying the measured evidence, the live wiring
and the landmine constraints; reviewed, grounded and corrected by this session
before anything was written down as fact. **Where I re-measured one of the plan's
figures and it came out different, the correction is below and the plan's number
is not used.**

---

## 1. What the defect is, in plain terms

Some sites keep a page's make-up as a list of **slot labels** — `prose-0`,
`tool-1` — meaning "the first prose block on this page", not "a component of type
X". The label points at a stored row that records which component actually renders
it.

A checker runs over every proposed site plan and deletes any section name it does
not recognise as a component. It only ever knew component names. It has never been
taught to look at what the page *already has*. So on these sites it deletes every
slot label as unrecognised, the plan comes out with those pages empty, and the
empty list is saved over the real one.

The build path and the re-render path had the same blindness and were both fixed
by consulting the page's stored rows first. This plan applies that same decided
pattern to the surface that was missed, and then closes the door on the save step
that lets an empty list overwrite a real one.

## 2. Scope — four resolver call sites, two write surfaces

All four callers of `loadComponentNameResolver`, verified at HEAD 2026-08-21:

| # | site | location | verdict | reason |
|---|---|---|---|---|
| 1 | `ValidateSitePlanAction`, `validate_components` arm | `v3_site_actions.go:3796`, drop at `:3838` | **FIX — the primary one** | The only site with measured damage: 140/140 recorded drops, 41 pages emptied 08-20. The drop happens BEFORE the object-form→string split, so RFC_016 `facts` die with the entry. |
| 2 | `applyAddToPage` | `apply_gap_plan_action.go:230`, drop at `:244` | **FIX — low urgency** | Latent: zero recorded drops. And its `add_sections` spec key has no reader — grep finds only writers (`apply_gap_plan_action.go`, `check_directory.go:279`) and exactly one live agent config mentions it, `content-gap-planner`, the producer. [MEASURED 2026-08-21] So a drop here loses nothing that is read *today*; the fix is insurance plus honest drop-records. |
| 3 | `applyNewPage` | `apply_gap_plan_action.go:369`, drop at `:374` | **DELIBERATELY LEAVE** | A genuinely new page has no `page_components` rows, so a positional name there points at nothing and dropping it is right. Pinned by a test so nobody "fixes" it later. |
| 4 | `applyRetypeExisting` | `apply_gap_plan_action.go:900`, drop at `:905` | **FIX** | The page exists. A retype naming its stored slots must not have them replaced by `defaultSectionsForPage` — and this one writes `sections = $3::jsonb` straight onto the live row. |

Write surfaces:

| surface | location | verdict |
|---|---|---|
| `sections = EXCLUDED.sections` in `upsertPage` | `site_db_actions.go:1201` (one Go caller, `SyncPagesToDBAction`; **three** live agents: `build-site-planner` `sync_pages`, `pageflow-builder`, `site-work-orchestrator`) | **FIX — Phase C, separate commit** |
| retype `UPDATE … sections = $3::jsonb` | `apply_gap_plan_action.go:924` | **LEAVE** — replacing a stranded page's sections is its licensed purpose, and after fix #4 the stored slots survive into `$3`. Incoming list is never empty, so Phase C's guard shape would be inert here anyway. |
| `applyNewPage`'s conflict arm `UPDATE pages SET title=$3, sections=$4::jsonb` | `apply_gap_plan_action.go:641` | **OUT OF SCOPE, named as a residual.** It overwrites a live page's list wholesale when a "new" page turns out to exist with the same type. That is a name-collision replacement in the `bugs_closed/081` / RFC_010 family, not a resolution defect — the incoming names resolve fine. Phase C does not catch it either (incoming list non-empty). Recorded so it is not rediscovered as 204. |
| `adopt_verbatim.go:478`, `apply_adoption_plan_action.go:548`, `create_blog_posts_action.go:244`, `page_role_upsert.go` | — | **OUT OF SCOPE** — different authorities, none downstream of the resolver. Named so nobody reads Phase C as "sections writes are now guarded fleet-wide". They are not. |

## 3. The design, and the landmine it has to survive

### 3.1 The constraint

`LANDMINES.md`, entry *"Widening a planner's component MENU changes nothing on its
own"* (added 2026-08-16, footprint `loadComponentNameResolver` ·
`componentNameResolver.resolve` · `ValidateSitePlanAction` · `validate_components`):

> **⚠ do NOT "fix the inconsistency" by widening `loadComponentNameResolver`
> itself.** Three of its four call sites are `apply_gap_plan_action.go` …, whose
> menu 407 and register PLAN-049 record as **deliberately** not widened —
> "gap-planning a NEW tool page is a different authority". A resolver-level
> widening hands that path an authority an owner-facing decision withheld,
> silently and with no config to show for it.

The obvious fix — a sixth arm on the shared resolver — is precisely this.

### 3.2 The shape chosen: a **page-scoped stored-slot membership rescue in each drop branch**

Not a resolver widening. Three properties, each **enforced by code, not asserted in
a comment** (CLAUDE.md: *a comment is not a control on a tree this many sessions
share*):

1. **The resolver is untouched** — its query, `validFunctions`, `resolve()`'s five
   arms and `addMenu` are not edited. A widening is a *site-wide grant*: anything
   added to the valid set becomes placeable by every caller on every page. This
   adds nothing to any valid set.
2. **The judgement is keyed per page on realised rows.** A name is kept only when
   the page it is proposed for already carries a stored row under exactly that slot
   name. The arm therefore **cannot authorise placing anything anywhere it is not
   already placed**. The authority granted is *"keep what this page already has"* —
   the opposite of the placement authority PLAN-049 withheld. That scoping is a
   property of the map's key.
3. **It sits in each call site's drop branch, per site, visibly** — and
   `applyNewPage`, the one path whose product is genuinely new surfaces, is
   deliberately excluded.

### 3.3 Source of truth: `page_components`, NOT `pages.sections`

**This is a correction to the bug file's own 08-20 contribution**, which proposed
*"leave an unresolved name in place when the page's realised `sections` already
contains it"*. That names the wrong store, for two measured reasons:

- `pages.sections` is **the column the bug destroys**. After one un-fixed run has
  emptied it, the "source of truth" says the page has nothing and the fix stops
  fixing.
- The realised list reaches validate through the `existing_pages` collected-data
  field, and **`site-planner` has no `load_existing_pages` step** — its step list
  is `complete, load_available_components, load_style_collections, plan_site,
  validate_plan`. A fix keyed on that field is structurally inert on one of the two
  live consumers. [MEASURED 2026-08-21]

`page_components` is the ground truth (the rows the page actually serves), it is
what `plan_sections` chose (`loadPageSlotComponentIDs`), and reading it from
`params.DB` covers both agents uniformly. LOCK-007's rule generalises: **read the
store the damage cannot reach.**

### 3.4 Decisions taken inside the arm

- **Kept entries are kept VERBATIM** — object shape and `facts` intact, name not
  rewritten to the component's function. Rewriting would collapse two `prose-N`
  slots that share a function onto one name, which is the positional naming's whole
  purpose and the bug file's own candidate-3 rejection.
- **Lazy, once per run, on first miss.** Honest-function sites issue zero extra
  queries, the hot path is untouched, and the existing validate sqlmock tests
  (which enumerate expected queries in order) do not all need editing.
- **A slot-query failure SKIPS the drops for that run, loudly** — it must not
  degrade to "drop everything stored", which silently re-lands the incident on any
  transient DB error. This mirrors the arm's own precedent
  (`if len(resolver.validFunctions) > 0` already skips resolution wholesale when
  the base fails to load). A typo surviving into a plan is recoverable downstream;
  an emptied decomposed page is recoverable only from a snapshot.
- **Membership is on slot existence**, regardless of whether the stored
  `component_id` still resolves to an active component. A kept-but-stale slot
  reaches `plan_sections`, which handles exactly that case loudly. Validate's job is
  not to garbage-collect live pages.
- **Observability:** one Info per keep plus a `sections_kept_by_stored_slot` count
  in the action's result, so it lands in `collected_data` where a canary can read
  it. This is the `resolvedViaMenu` lesson — without a positive tell, *"the fix
  works"* and *"the planner happened to propose only honest names"* are
  indistinguishable. No per-keep `agent_error_log` row: a keep is the correct
  outcome, and 100+ rows per replan is noise.

### 3.5 Where the loader lives, and `plan_sections`' private copy

New `platform/orchestration/datahelpers/page_slot_identities.go`, on the LOCK-008
pattern: one exported SQL var (so a test can pin it), per-page and per-site
loaders, two **pure** derivations — `SlotIDMap` (the slot→component_id map, with
`loadPageSlotComponentIDs`' exact conflict rule: repeats that disagree are dropped,
"an ambiguous carry source is no carry source") and `SlotNameSet` (page→set of slot
names, the membership judgement the three fixed sites share). Membership needs no
conflict rule — it decides keep-vs-drop and never picks an id — which is why it is
a separate pure function, not a reuse of the id map. Header carries the
consumers-told list.

**`plan_sections`' loader unifies NOW**, as a thin delegation preserving its exact
log strings and error text (204's own closure evidence pod-greps those strings).
Leaving it creates the two-hand-maintained-copies drift class this estate reviews
for; "unify later" is how a seam becomes folklore. The risk of touching a live
proven path is contained by phasing: the extraction is its own commit,
behaviour-identical, suite-green from `git archive HEAD` before anything
behavioural lands.

## 4. Governance — the rulings, made rather than dodged

**RFC scope — NOT architecture-scope; normal council gate, with the guarantee
sentence rewritten in the same commit.** The sentence that changes is
`validate_components`' own comment: *"every surviving section name is a valid
component function"*. Per owner ruling 2026-07-29 §1 a guarantee change is the RFC
trigger, so this needs arguing, not waving through. Three grounds:

- **The worked RFC case ran the other way.** RFC_002's trigger was an evaluator
  *gaining* destructive power (refute where it had only confirmed). This change
  *removes* a destructive power from validate and grants no new power downstream.
- **Downstream already meets these names, by the estate's own decided design.**
  `pages.sections` carries positional names on 7 live sites / 107 names today, and
  the readers were built for them: 182 (rerender), `13252f714` (build), PBP-035
  (`stored_slot_name`), `revalidateNeedsPage` (matches on `slot_name` precisely
  because functions under-match). Validate was enforcing a stale invariant against
  stores the platform had already moved past. **Preserving restores the status quo
  ante; it does not mint a new shape.**
- **The new guarantee is narrower than it sounds and is checkable:** *"a valid
  component function, **or the slot name of a stored row on that same page**"*. The
  second disjunct is page-local and realised.

What is owed instead, and is in the plan: the council gate before/alongside the
commit; the guarantee comment edited visibly; the concept-register amendment in the
**same commit** (ordering-exemption condition 2); consumers named and told.

**Opt-in field — no config key; the arm ships unconditional.** The 2026-08-02 §2
rule targets a seam *"whose widest branch is licensed by 'callers must all be X'"*
— a caller property enforced by comment. This arm's licence is page-local stored
data enforced by a map key, not a caller property. Further: the two sibling call
sites of this same judgement (`plan_sections` Path 0, 182's rerender) both run
unconditionally and both were council-approved — making the third opt-in would
recreate the half-fixed asymmetry that **is** bug 204. And the unsafe default would
be OFF-ness itself: default-OFF preserves the behaviour that emptied 41 pages,
everywhere, until config names each agent.

RFC_022 is not in play either way — no optional key is added, so the WFA-013 budget
is untouched. Which also matters because **`validate_site_plan` has no
`RegisterActionInputSpec` at all** (207 exist across the estate), so any key added
to it would be invisible to the optional-key audit. Adding keys to an unaudited
action is something this plan declines to do; the missing input spec is noted as a
separate small task, not smuggled in here.

**Consumers of the changed guarantee, enumerated from code and TOLD** (owner ruling
2026-07-29 §3). The output that changes is validate's surviving-sections list:

- `write_site_plan` → `site_plan_sections.component_name` (`write_site_plan_action.go:659`):
  will now carry per-page slot names for decomposed pages. `component_name` stops
  being guaranteed to be a `content_components.function`.
- `sync_pages_to_db` / `upsertPage` → `pages.sections`: round-trips the stored
  names. Restored status quo, no new shape.
- Readers of `site_plan_sections`: `reconcile_site_plan`'s `decideEmit`;
  `load_page_sections_from_spec` tier 1 → `plan_sections` (built for positional
  names, proven live); `plan_section_counts`; `page_section_satisfiability` /
  `revalidateNeedsPage` (already slot_name-matching — this *improves* its match);
  `check_section_source_drift`; `check_sectionless_pages` (stops being handed
  falsely-emptied pages).
- **Told:** register amendments to PLAN-027 and the plan_sections/PBP-035 entries;
  dated notes in `bugs_closed/204` and the reconcile lane; the new datahelpers header
  carries the list. The message is *"a validated plan may carry, for an existing
  decomposed page, that page's stored slot names; resolve identity via
  `page_components`, never by assuming a section name is a function."*

**Ordering by what closes the door:** the resolver arm makes the bad state (stored
identity deleted by validation) unrepresentable at source; the Phase-C write guard
makes the worst *consequence* (empty-over-real persisted) unrepresentable against
any future emptier. Candidate 2 (loudness) is already live since 08-16. **Nothing
here asks an operator to remember anything.**

## 5. The write-side guard (Phase C) — separate commit, and why

Separate because it is a different mechanism with a different blast radius
(`upsertPage` serves three live workflows the resolver fix never touches), because
it is a *class* defence rather than this bug's fix, and because the resolver fix is
complete without it.

**Reconciling it with plan-authority.** The `upsertPage` landmine records that this
helper deliberately carries `page_type = EXCLUDED.page_type` because *"the plan is
the authority on what a page is"*. Sections stays plan-authoritative for every
**non-empty** proposal — a replan that recomposes a page still wins, exactly as
`meta_description`'s 08-19 guard lets a real incoming value win and refuses only a
blank. The guard intercepts only the transition **non-empty → empty**. Sections
differs from page_type in the way the incident proved: a wrong page_type is visible
and rewritable by the next plan; an emptied sections list destroys the only record
of a decomposed page's composition — `page_components` keeps serving, so nothing
looks wrong, and the next build builds an empty page over a live one.

**Zero sections stays representable**, and this is measured, not hypothetical:
**72 of 748 active pages live legitimately at `sections=[]`, 60 of them tools**
[MEASURED 2026-08-21 — re-measured by this session, matches the plan's figure].
Three doors stay open: empty-over-empty untouched; new-page inserts untouched (the
guard is in the `ON CONFLICT` branch only); and a deliberate emptying goes through
the **`recompose_pages` release**, the already-sanctioned explicit-redesign channel
readable from CollectedData — no new config key, no operator memory, because the
intent channel already exists and already means this.

Shape: one CASE clause parameterised by an allow flag, single-statement and
race-free, with `RETURNING` extended so the action detects a refusal and files a
durable `PAGE_SECTIONS_EMPTY_OVERWRITE_REFUSED` finding per page. **A silent keep
would be a new landmine** — the wrong result would look exactly like the right one.

## 6. Phasing

All Go. Nothing is live until an image rolls; there is no config half, so no
config-vs-Go ordering hazard.

- **Commit A — the extraction.** `datahelpers/page_slot_identities.go` +
  `plan_sections` delegation + moved tests + register entry. Definition and caller
  in **one** commit so they land together compiled. (The "commit the definition
  first, ALONE" landmine governs *splitting* an extraction across commits; not
  splitting is strictly better — a caller with no definition takes the whole repo's
  build down.) Behaviour-identical; suite green from `git archive HEAD` before B.
- **Commit B — the behaviour.** Validate arm + gap-plan sites 2 and 4 +
  observables + tests + the guarantee-comment rewrite + register amendments +
  the `bugs_closed/204` fix note. **Council submission 1** covers A+B as one coherent
  task; `Council-Submitted:` trailer if committing before the verdict lands, and
  never `Council-Reviewed:` on a verdict not read.
- **Commit C — the write guard.** **Council submission 2.**
- **Then:** the next whole-fleet roll (owner runs `make release`), then §7. B and C
  are independent to roll; C is protective even alone.

## 7. Tests that could FAIL, and verification at the artefact

Full test matrix, each assertion with its disconfirming result, is in fable's
returned plan and is reproduced in the implementation commits' test files. The ones
that carry the argument:

- **Page-scoping:** slot `prose-0` stored on page A only, proposed on page B →
  **dropped**. Disconfirm: kept. *This is the test that makes the landmine defence
  a property rather than a claim.*
- **Arm ordering:** a stored slot whose name is also a valid function resolves at
  arm 1, the keep-arm never fires, and **no slot query is issued**. Mutation: move
  the check ahead of `resolve()` → this test must fail.
- **Honest drop intact:** a proposed name that is not stored is still dropped and
  still recorded. Disconfirm: kept. This is the regression guard for the checker's
  real job.
- **`applyNewPage` still drops a positional name** — pins the deliberate leave.
- Zero-drop assertions use `recordDroppedSectionNames`' **returned count**, which
  exists precisely because a database mock cannot prove a negative.

Verification (for the later session):

1. **Provenance per SERVICE, never `strings`:** the pod's own `build provenance`
   line → `git merge-base --is-ancestor <commit-B-sha> <stamp>`; the line scrolls,
   so fall back to `grep -aq "<known-sha>" /proc/1/exe` **with a control in the
   same breath** — one sha that must be present, one that must be absent. One pod
   per *deployment* running the chassis image (one image serves many).
2. **Behavioural canary:** the 08-20 replan shape on loanandmortgagecalculator.co.uk
   with that incident's own containment runbook (pre-fire snapshot, digest pinned,
   cancel the emitted queue before any repair). Assert `validate_plan`'s pages carry
   the positional names, `sections_kept_by_stored_slot > 0` in the same row, and the
   post-`sync_pages` `pages.sections` digest equals its pre-fire value.
3. **Demand control on the zero.** Fleet `PLAN_SECTION_NAME_DROPPED` positional rows
   going to zero is the success metric — but a post-fix zero needs proof the detector
   still fires. Induce one drop with an invented name and confirm the row appears.
   **A zero without that control is a blind pass.**
4. **Negative control:** the same replan shape on an honest-function site → keep
   count exactly 0. If the arm fires there, the scoping property failed; stop.
5. **The 107-names / 7-sites census is NOT the fix's metric** — it measures
   decomposition, not damage, and should be unchanged. Saying so here stops the next
   session misreading it.

## 8. Risks, specifically for the sites that are NOT decomposed

> **CORRECTION to the drafted plan (this session, 2026-08-21).** It said *"the 141
> non-decomposed sites"*. There is no such number. `sites` holds **45 rows**
> (23 deployed, 17 pool, 2 active, 2 test, 1 system) and **27 have any active
> page**; 7 of those carry unresolvable names, so the non-decomposed population is
> **about 20 sites**, not 141. The figure is not load-bearing for the design — the
> inertness argument is structural, not a headcount — but an invented number in a
> risk section is exactly the shape that gets quoted onward, so it is corrected
> here rather than repeated. [MEASURED 2026-08-21]

- **Structural inertness:** the arm fires only after all five resolver arms and the
  menu union have missed AND the name equals a stored slot on that page. On honest
  sites the slot names *are* component functions, so they resolve at arm 1 and the
  arm cannot fire. To be measured and recorded before Commit B: stored `slot_name`s
  unresolvable by the resolver, grouped by site — expected non-zero only on the
  known 7.
- **Worst realistic regression:** a planner-invented junk name that coincidentally
  equals a stored slot on the same page is now kept rather than dropped; downstream
  `plan_sections` resolves it by stored id to the component actually at that slot.
  The failure mode degrades from *"section deleted"* to *"existing section
  rebuilt"* — the tolerable direction. Stated so a later session recognises it
  instead of re-deriving it.
- **`decideEmit` stale storm [INFERRED]:** the first post-fix replan writes
  positional names into `site_plan_sections` where prior plan versions have no rows
  for those pages, so pages may classify "stale" and produce one wave of rebuild
  emissions. Bounded by counting at the canary before releasing the queue; the
  reconcile lane is told.
- **Phase C false refusal:** a legitimate flow that empties a non-empty page through
  sync without a recompose release would now be refused. No such flow is known, and
  it is **[UNMEASURABLE from history]** — transitions leave no trace. The refusal is
  durably recorded, so the first occurrence is a one-line diagnosis rather than
  silent damage in either direction.
- **Same-file passenger:** `v3_site_actions.go` is a hot shared file. Commits are
  pathspec-scoped and the scope block is read each time, but a same-file passenger
  from a concurrent session remains possible and no hook prevents it. Re-run
  `git status` immediately before each commit.
