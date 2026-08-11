# 151 — the per-section content writer has no memory of facts already used elsewhere on the site

Filed 2026-07-29 by session "fundamentallyai.com 4" (`brochure_component_library` lane), after the
owner flagged that fundamentallyai.com's home and capabilities pages "seem very similar" and asked
how the framework can avoid this in future. **Structural, cross-cutting, unfixed** — the symptom on
this one site was fixed by hand (7 sections removed across 5 pages, this session); the mechanism
that produced it is untouched and will reproduce on the next site built the same way.

## Symptom

Measured on fundamentallyai.com before any edit (per-section fact census, `site_specs.evidence_base`
9 facts as the ruler): **18 sections across 5 of the site's 10 pages each independently restated 3 or
more of the same 9 approved facts.** Home had six capability-listing sections out of eight, three of
them consecutive, two sharing the literal heading "What this platform demonstrably does" with a third
instance on `/capabilities.html`. The `info-card-grid` component instance on home vs. capabilities was
only **18% textually similar** (`difflib.SequenceMatcher`) while asserting the identical six facts —
independently-generated near-duplicate content, not a copy-paste.

## Root cause, grounded in the actual write and plan paths

**Per-section copy is written in total isolation from sibling sections.** The `page-content-writer`
workflow's `process_sections_loop` calls `generate_content` (an `execute_llm_prompt` action) once per
section (`docs/agent_docs/docs001_flow_general/100_content_page_build_handler_flow.md:18-21`). No
other section's output — same page or a different page on the same site — is passed into that call.

**Every one of those isolated calls receives the identical, undifferentiated fact pool.** The live
prompt (`docs/agent_docs/docs024_key_docs_latest/brochure_component_library/sql/page_content_writer_prompt_v3_2026-07-25.txt:68-72`)
injects `site_specs.evidence_base.writer_block` verbatim. `writer_block` is built **once per site**,
not once per section or per page, by `composeWriterBlock`
(`platform/orchestration/actions/refresh_evidence_base_action.go:582-637`): it walks every fact with a
`writer_line` and concatenates them into one NUMBERS/CAPABILITIES block, with no per-fact record of
where it has already been used. `EvidenceFact` itself
(`platform/orchestration/datahelpers/claims.go:74-96`) has fields for `ID, Claim, Value, Kind, Source,
VerifiedAt, Tolerance, ContextTerms, Observations` — **no field tracks assignment or usage**. This is
not a dead-but-present field (the pattern this codebase usually finds) — the field does not exist.

**The planner, which is the one place with full cross-page visibility, doesn't guard against this
shape of duplication.** `build-site-planner` runs once per site and emits every page's section list
together (`docs/agent_docs/sql_for_agents/053_build_site_planner.sql`), so it structurally *could*
spread facts across sections. Its only duplication guard is page-level topic duplication — "never
duplicate, replace, or rename an existing page" (`053_build_site_planner.sql:2461`, `:3228-3229`) —
nothing about facts or component shape repeating within or across pages.

**Component selection doesn't know a component's shape either.** `SelectComponentByType`'s scoring
(`platform/orchestration/actions/component_selector.go:150-193`) weighs only site-type/page-type match,
`avg_quality_score`, specificity, and `usage_count` — there is no signal for "this component enumerates
the whole fact roster" vs. "this component is a single-topic block", so nothing stops the planner
picking two roster-shaped components for one page, which is exactly what happened three times on home
alone. (Separately, `features_open/017` already records that for this site the planner never selected
any of the newer interactive components at all — every placement was hand-done — so its stock choices
here are precisely the "familiar, repeatedly-picked" default that file describes.)

No existing bug or concept-register entry names this class. `evidence_base`/`writer_block` have several
closed bugs (`043`, `073`, `074`, `104`, `105`) — all about accuracy, staleness, or verification of
individual facts, none about the same fact being independently restated across sections.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Scope facts to sections at plan time (closes the door).** `build-site-planner` already runs once
   per site with full visibility across every page — the natural place to solve this without adding a
   new stateful pass. Extend its output so each section slot carries an explicit subset of fact IDs
   (a new `site_plan_sections` column, e.g. `assigned_fact_ids jsonb`), instructed to spread facts
   across sections rather than repeat them. `page-content-writer`'s per-section prompt then builds a
   *filtered* `writer_block` from only that section's assigned facts (reusing `composeWriterBlock`,
   given a subset instead of the full array) instead of the whole-site block. A section that is never
   told a fact exists cannot restate it.
2. **Give `content_components` a shape tag** (e.g. `is_fact_roster boolean` or an enum alongside
   `section_type`) and teach the planner/selector not to place two roster-shaped components on one
   page, or the same roster-shaped component on sibling pages. Weaker than (1): it stops two card-grids
   colliding but does nothing about two *narrative* sections independently restating the same fact —
   the actual failure mode on `about.html` this session (`about-content` prose vs. `differentiators`
   grid, both full restatements).
3. **A post-build fact-repetition census as a gate**, analogous to the claims gate: after a page (or
   whole site) build, flag any fact appearing in 3+ sections, or any two sections sharing 3+ facts. This
   is the check this session ran by hand in ~15 minutes of SQL + a small Python script — cheap to make
   permanent. It does not prevent generation, but it catches the failure before a site is treated as
   done, and it is the only candidate here that also protects the 9 already-deployed sites while (1) is
   being built.

**(1) is the structural fix; (3) is worth building regardless of (1) shipping**, on this codebase's own
established pattern of pairing a correct-by-construction fix with a gate (evidence seeding + claims
gate; link repair + phantom-link gate). (2) is a partial mitigation, not a substitute for either.

## How to verify a fix

Re-run this session's own measurement — per-section fact census against `site_specs.evidence_base`,
counting sections that share 3+ facts, plus a `difflib.SequenceMatcher` textual-similarity check
between same-shaped components on sibling pages — against a newly-built site. Zero sections sharing
3+ facts is the target; the fundamentallyai.com fix (this session) is a hand-done existence proof of
what "zero" looks like, not a fix to the mechanism.

---

## Second site measured: vonc.com, 2026-07-30 — and the method has a floor

Run by the `gauntlet_dead_cta` lane at the owner's request ("check that we're not
writing the same things all over the site"), using this bug's own method, now a
reusable script: `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/dedup_census.py`
(fact census + `difflib` + exact-block matching in one pass, any site).

**This bug's shape reproduced.** vonc's eight `/archetypes/*.html` pages restate
each other: `call-to-action` sections at **0.90** (oracle↔surgeon), 0.83, 0.79,
0.74, and `hero` sections at 0.67–0.71 across ten more pairs. Same cause as
recorded here — per-section copy written with no sibling context, one
undifferentiated fact pool per site. Contributing the instance rather than
refiling.

**But the fact half of the method has a FLOOR, and vonc is under it.** vonc's
`site_specs.evidence_base` holds **4** facts (8 archetypes / 3 tools / 2 guides /
18 pages) against fundamentallyai's 9. Result: of 55 sections, **50 assert zero
facts, 5 assert exactly one, and none assert 2 or more** — so this bug's "3+
shared facts" threshold **cannot fire on this site at all**, and the duplication
that is really there is textual.

*The trap to record: a clean fact census reads identically to a site with no
duplication problem.* On vonc it means the ruler is too short to measure with, not
that the site is clean — and the site is in fact the worse of the two, because it
has a page rendering every section twice (below). **Any future gate built from
this method needs a stated minimum evidence-base size, or a fallback to the
textual half when the fact pool is too small to discriminate.**

**Unrelated to this mechanism, found by the same run and filed separately:**
vonc's `/about.html` carries 12 component rows that are **6 identical pairs** (same
`slot_name`, consecutive positions, similarity 1.00) and **paints twice** on the
served page — `'The rules are simple'` ×2, `'Game Master'` ×16. That is duplicate
ROWS, not independently-generated near-duplicate copy, so it is a different defect
with a different cause. Handoff D in the `gauntlet_dead_cta` directory has the
evidence and the first query to run; whoever picks it up should establish why two
rows exist before deleting either, since an insert-instead-of-upsert path would be
platform-wide.

> **RESOLVED 2026-07-30 → `bugs_open/156`.** Diagnosed and the site fixed. It is
> **not** an insert-instead-of-upsert path: the save does DELETE-then-INSERT, and the
> row **positions were 1..12 distinct and monotonic**, which refutes both a
> concurrent-save race (two writers each number 1..6, giving two rows *per position*)
> and a whole-loop re-run (which gives 1,2,3,4,5,6,1,2,3,4,5,6). It was one loop over
> an input list that already held 12 entries. Fleet-wide it is **not** platform-wide
> in the state it leaves behind: 17 duplicate `(page_id, slot_name)` groups exist and
> **11 are legitimate** repeated slots with differing content, so ⚠ **a unique index
> on `(page_id, slot_name)` would break 11 real pages.** The producer is
> `[UNRECOVERABLE]` — past retention. 156 stays open as a detection gap.

### vonc's instance of THIS bug: hand-fixed 2026-07-30, mechanism untouched

Owner asked for the archetype copy to be rewritten, so vonc is now in the same state
as fundamentallyai — **site fixed by hand, mechanism still open.** All 8 heroes and
all 8 CTAs rewritten (`content_data` and `rendered_html` together), rerendered and
verified live. On this bug's own census: worst cross-page pair **0.90 → 0.64**, pairs
≥0.70 **→ 0**, hero pairs no longer appear at all, zero identical blocks.

Two things from that rewrite worth keeping for whoever fixes the mechanism:

- **There is a legitimate floor, so a gate built on this method must not target
  zero.** The eight CTAs keep identical *button labels* on purpose — they must agree
  with the URL they point at, and one consistent action label across sibling pages is
  correct UX, not duplication. Measured on the worst remaining pair: **0.64 with the
  labels included, 0.43 on prose alone.** `dedup_census.py`'s `SKIP_KEYS` drops
  url/asset keys but **keeps** `*_cta`/`cta_*` label fields, so it scores that floor
  as similarity. Either exclude label-ish keys or set the threshold above the floor.
- **A rewrite is the moment unbacked claims get laundered.** The sentences being
  replaced carried "best prediction accuracy on the platform", "highest remix rate in
  the Arena", "dominates the Stage", "Thousands of takes already on record" — none
  backed by vonc's 4 approved facts. They were dropped rather than reworded, because
  rewording an unverifiable claim makes it the rewriter's own. **If this mechanism is
  ever fixed by regenerating sibling-aware copy, the fix must not carry the existing
  claims forward as context** — that would relaunder them at scale.

---

## STATUS UPDATE 2026-07-31 (gauntlet_dead_cta lane) — candidate 3 is BUILT, APPROVED, guarded, and INERT

Candidate 3 (the post-build census as a gate, with a deterministic repair) exists.
Council-APPROVED round 3 on trail `da3f2d9b-ae6f-492d-ad3b-748323b66367` (12 approve /
2 advisory), after rounds 1–2 caught real defects — including a measured false positive
that narrowed the in-remit identity to **same page + same slot + byte-identical
`content_data`** (two different vonc components share the byte-identical site-context
boilerplate blob; the prose rule would have deleted a live home-page row).

Same day, on the owner's decision, the repair gained a **plan guard** (commit
`5c4dc317f`, council corr `6c5d1491-08dd-4e66-ab10-e82dab78dc82`): a duplicate group
whose slot repetition the effective plan source (`site_plan_sections` →
`site_specs.site_plan` aspect → `pages.sections`) itself specifies is skipped and
reported, never filed and never deleted — closing the `bugs_closed/058`/`069` shape
before the checker is ever enabled.

**It is all INERT**: zero discovery agents name `content_duplication` (verified across
all five workflow-bearing columns of `agent_definitions`). Enabling is the
`brochure_component_library` lane's decision; their cold-start handoff carries an
INCOMING section and `CONTRIB_2026-07-31_151_candidate_3_is_built.md` has the full
account. **Candidate 1 (facts assigned at plan time) remains the structural fix and
remains unbuilt** — the residue capability_gap items are the population it should
clear when it lands. Full design rationale travels with the mechanism:
`doc_notes` `subject_type='action'`, keys `remove_duplicate_page_sections` /
`check_content_duplication`.

**UPDATE 2026-08-02 (gauntlet_dead_cta):** candidate 3's code **shipped in chassis
v1.0.1231 and is pod-verified on both replicas** (added-string, positive-control and
reconciled negative-control greps — evidence in
`gauntlet_dead_cta/HANDOFF_2026-07-31_continue_here.md` §7). Still INERT, re-verified
same day: zero agent references, zero duplication work items. Enabling is now the
single config step and remains the `brochure_component_library` lane's call.

---

## STATUS UPDATE 2026-08-06 (brochure_component_library lane) — candidate 1 is BUILT, SUBMITTED, and INERT

Candidate 1 (facts assigned to sections at plan time) is implemented end to end and
committed; council-submitted alongside the commit, corr
`902a8563-2200-4771-ac0f-55dab0839a02` (verdict not read at commit time — the
trailer is `Council-Submitted:`, which asserts nothing). Registered as **PBP-037**
in the same commit, per the ordering-exemption's condition 2.

The shape, in one line each: migration `327` adds nullable
`site_plan_sections.assigned_fact_ids` (NULL = unscoped = every existing row);
`build-site-planner`'s prompt (seed `329`) shows the fact roster and takes
object-form section entries `{"name", "facts"}`; `validate_plan` resolves objects
by name and normalises back to string sections + an aligned per-page
`section_facts` array (facts travel INSIDE the entry, never positionally — the
imagery-keying scheme mis-keys silently when an entry is dropped);
`write_site_plan` persists the assignment; `load_page_sections_from_spec` emits it
only from the authoritative tier; `plan_sections` (opt-in step-config key, seed
`328`) attaches `facts_scoped`/`assigned_fact_ids`/`assigned_writer_block` to each
ready item, composing via the existing `composeWriterBlock` filtered to the
assignment; writer prompt v4 (seed `330`) branches on `current_section.facts_scoped`
with an explicit factless instruction for `[]`. Values substitute at compose time —
the assignment pins WHICH facts a section states, never their numbers.

**INERT until: image roll → mig 327 → seeds 328/330 → seed 329.** Each half
tolerates the others' absence (measured/mutation-tested, 9 tests). Acceptance =
replan fundamentallyai (pool 15, 9 fact-overlap pairs) and watch the census pair
count fall; the five fact-blind sites are expected NOT to move. The 7 flag-only
capability_gaps stay open as the drain population — draining means replan+rebuild
through the claims-gated path and is sequenced behind `bugs_open/189`'s config half
for locked-row pages. Candidates 2 (component shape tag) remains unbuilt and
unclaimed.

> **CORRECTED 2026-08-06 (same evening, post-verdict): the council REJECTED the
> submission — hard guardian veto, round 1 (corr `902a8563`) — on BREADTH, not
> correctness (6 of 11 seats approved; the veto is about three fleet-wide prompt
> changes + four Go files reaching production as one slice). Per the standing
> owner rulings the code stays; the ROLLOUT ORDER above is superseded:**
> **Slice A** = image roll → mig `327` → seed `329` ONLY (planner emits
> assignments, nothing consumes them) → inspect real plan rows; **Slice B** =
> seeds `328`/`330`, its own council round, piloted on a small cohort, and a
> human reads the v4 prompt plaintext before `330` applies. Full contract +
> veto record: `architecture_review/RFC_016`. Two objections were fixed in code
> the same night (empty-composition degradation + durable
> `FACT_SCOPING_EMPTY_COMPOSITION` record — a broken assignment can no longer
> render as a deliberately factless section).

> **2026-08-07: SLICE A OBSERVED LIVE** — replan of fundamentallyai (corr
> `801b0732`, plan `8ee5807b`): wiring end-to-end correct (71/71 emitted
> entries persisted, tri-state intact, both consumption negatives re-verified
> while the replan's own builds ran), planner uptake PARTIAL — object form on
> 5/24 pages (all newly-composed ones), 2/9 offered facts assigned (both
> topically exact), every carried-over page unscoped. **Consequence: against
> this plan the acceptance above (census pair-count falls) would NOT move —
> the overlap pairs live on the unscoped pages.** The Slice B council round
> must choose: require object-form for every page (prompt change, same seed
> 329 already touched) or re-scope the acceptance to engaged pages. Options +
> evidence: RFC_016 §3a; full trail in the lane NOTES 2026-08-07. Side-find:
> the imagery positional-keying defect this lane dodged is live in production
> — filed `bugs_open/214`.

> **CORRECTED 2026-08-08 (the day after): the 08-07 note above attributes the
> wrong mechanism.** "Planner uptake PARTIAL — object form on 5/24 pages (all
> newly-composed ones)" was read from validate_plan's output, which is
> downstream of `reconcilePlanWithRealised` Pass B2 — the pass that (correctly)
> restores realised sections over the LLM's for every DEPLOYED page, discarding
> fact assignments with them. The 08-08 compliance replan under seed 333 (rule
> 17 mandatory-per-page, owner decision RFC_016 §5) proved at the RAW emission
> that the planner assigns richly on built pages (index stat-band = F1+F2+F4) —
> and that Pass B2 throws it away. **So: prompt half DONE (333 live); candidate
> 1 cannot reach deployed pages without candidate 1b (RFC_016 §3b: planner
> re-emits realised sections for deployed pages + Pass B2 carries facts by
> name-match); Slice B round HELD until 1b is designed in.** The replan itself
> died on `bugs_open/215` (canonicalised page-name collision kills the plan
> write — pre-existing, unrelated to facts, gates clean observations). No data
> damage (transactional write; plan 8ee5807b still current). Incident trail:
> WRONG_CALLS 2026-08-08; evidence in RFC_016 §3b.

> **CANDIDATE 1 MEASURED END-TO-END, 2026-08-11 (brochure_component_library lane — the
> census round, owner-authorised).** Replan of fundamentallyai (corr `e74974b3`, plan
> `40a66d3a`): all 18 singular built pages' compositions preserved by name+order (seed 362);
> all 9 writer-visible facts assigned, each to exactly ONE section; zero
> `FACT_CARRY_UNMATCHED_SECTION`/`FACT_ASSIGNMENT_ABSENT`. Rebuilds through the new path:
> handler stamps `facts_scoped`/`assigned_writer_block` per section (seen live, orch
> `b3ec646d`), writer states assigned facts in register-mandated form (floors: "10+"/"12+"/
> "0") and NOTHING unassigned — writer-prose fact overlap on pages rebuilt today: ZERO.
> Whole-site fact-overlap pairs 34 → 9 (same instrument, pre/post same morning; the recorded
> 9 from 08-05 had staled to 34 via the 08-07 cascade). The residual 9 pairs are all
> non-writer: 3× the same evidence-chart data served on three pages (composition question),
> 2× portfolio card resolver metrics, 4× stale pre-round copy on
> production-backend-engineering (drains at its next rebuild). The 8 fact-blind sites: zero
> writer/build orchestrations all round — the unscoped arm untouched. **The mechanism this
> bug asked for is live, consumed, and measured. Remaining scope is composition-level
> (charts/resolver data), not the writer.** Evidence: lane NOTES 2026-08-11 entries;
> census inputs/outputs pinned in the session scratchpad.
