# PLAN — vigilant senior designer + offer/benefit analyser

**Opened:** 2026-08-02, at the owner's direction ("a constant vigilant designer… and the offer
analyser and benefit analyser… all of the above can also be made into checkers and handlers
too for continual improvement. It needs a strong focus or we'll keep tripping up on the detail.")

**Organising rule:** every detector ships with its drain (handler + promotion + trigger).
The platform's proven failure mode is detection without consumption — `bugs_open/115`
(a correct brief-fidelity audit predicted the owner's complaints three days early and died at
`status='detected'`), `bugs_open/083`, and the 019 owner ruling that deferred sweep enrolment
for exactly this reason.

## Owner decisions (2026-08-02, at planning time)

| decision | answer |
|---|---|
| Cadence | **Manual triggers for now.** Build everything drainable; fire per site by hand. `improvement-sweep` stays `enabled=false`; flipping it on (G1) is a separate owner go — it reverses the 07-29 deliberate stop. |
| Build order | **Designer first**, then the offer analyser. Shared Phase 0 (drain) first. |
| Critic model | **Trial both** — vision seam implemented for Gemini AND Claude, same critique over 2–3 sites with each, compare actionability, record the ruling (supersedes the untested 07-24 Gemini call). |
| Authority | **Broad autonomy** — auto-apply at every altitude, locks respected. Two platform-constrained exceptions (below). |

## Autonomy model

- Auto-applies: CSS/detail fixes, spacing/responsive, CTA, imagery, bounded page recomposition,
  directed palette change (webdesign-agent + reference_values pinned in the same operation so
  the change sticks instead of churning), content/tone rewrites, strategy refresh.
- **Exception 1 — shared chrome**: 3 head components serve all 14 live domains; an automated
  chrome edit is fleet-wide by construction. Auto path = fork-to-site-scope first (pin the fork
  via `style_collections`, the pin predicate deliberately admits forks), then edit the fork.
  Built as A4.4, last; until then chrome findings file `capability_gap`.
- **Exception 2 — layout-triple swap**: no runtime re-compose path exists by deliberate design
  (DES-048/049). Stays a `capability_gap` review item in v1; a gated auto re-compose is
  architecture-scope, revisit once the critic demonstrates demand.
- Locks always respected (`locked_at IS NOT NULL` never touched, counted as `skipped_locked`);
  byte-preserved adopted sites → `needs_human_review` only.

## The full approved plan

The complete phase-by-phase plan (Programme A phases 0–5, Programme B phases B1–B5, the
item_type/handler/verifier table, verification acceptance list) was approved by the owner on
2026-08-02 and is reproduced in full below. It was designed against three exploration passes +
two planning passes over the live system the same day; the "facts that bind" list is what was
verified then — re-verify anything acted on later, this tree moves in minutes.

### Facts that bind the design (verified live 2026-08-02)

- Promotion is single-owner: migration 286 makes `improvement-loop.triage_findings` the sole
  `triage_detected_items` owner. Never a second triage step; `scripts/audit-single-owner-actions.sh` stays clean.
- `AIService` is text-only (`platform/aiservice/interface.go`). No vision path exists in the
  chassis; the screenshot critic needs a new seam (council + concept-register, same commit).
- `render-audit-agent` (VIZ-012) is live but its browser findings stop in `collected_data` —
  no write tail. `browserrunner/screenshots.go` exists but fires only on check failure.
- The 3-pass audit cap lies: capped sites report "clean" (`bugs_open/171`); the JSON path is
  `sites.settings.maintenance_profile.audit_pass_count`.
- `domain-strategist.create_next_item` unconditionally enqueues `needs_briefing` → rebuild
  chain. Must be gated (B2) before `check_premise_incomplete` ever enables (B3).
- Adopted finance sites (loancash class) have no `strategy` row; `bugs_open/115` is down to
  2 open findings — finding 2 (template repetition) is Programme A's acceptance test.
- Webdesign LLM may only change hue (8 core slots); `install_site_composition` refuses re-runs;
  sites lacking `design_intent.palette.reference_values` re-roll their palette every webdesign
  run (3 sites exposed; the generic_theme repair IS the churn until A4.3).
- Checker recipe: registry init() + sqlmock test + handler/verifier/registration coverage
  tests; check names into an agent's checks array only after the image rolls; IMP-016 order
  (observe-only → handler live → one clean cycle → enable).
- Active adjacent lane — do not collide: portfolio_positioning owns premise→writer wiring for
  new builds (their gap 1, lendzy shadow build). This workstream owns the read-back side of
  the same spec fields. Do not enrol lendzy until their marker test passes.

### Programme A — the vigilant designer

**Phase 0 — make findings flow.**
0.1 Repoint `scheduled_tasks.improvement-sweep.target_topic` → `system.agent.scheduled.requests`
(dead-topic fix), leave `enabled=false` (G1 owner gate). Its pre_query already self-enrols all
active/deployed sites with backpressure, so 019's enrolment is structurally solved by the same
future flip.
0.2 Replace the 3-pass cap with convergence + cooldown: pg function `site_audit_fingerprint(site_id)`
(md5 over ordered `pages.content_hash` + composed palette + chrome rendered_html hashes); state
in `settings.maintenance_profile.last_audit {fingerprint, at, passes_at_fingerprint}`; audit due
iff fingerprint changed OR 14d cooldown expired; **`triage_findings` runs on every path** (the
cap only ever protected LLM spend — promotion is one UPDATE); non-convergence brake at 3 passes
on an unchanged fingerprint → one `audit_not_converging` capability_gap row; `audit_skipped`
recorded so skipped ≠ clean.
0.3 Render-audit write tail: new Go action `write_render_audit_findings` — firm ContrastFinding
→ `contrast_failure` → css-patch-agent (item_key `render:contrast:<page>:<class>`; dedup + next
audit is the de-facto verifier); over_image reported never filed; BrokenImage → existing
undeployed-asset machinery or `needs_imagery` → image-build-handler (never the deliberately
flag-only `image_url_404`); Overflow → `responsive_fix` → component-template-fixer; locked
culprits skipped and counted. Config tail step added to render-audit-agent after the image rolls.
0.4 The ~205 detected items: no bulk promote; per-site hand-fired sweeps starting with one
specimen site; cancel provably-stale rerender rows first; hold 115's two rows for A3.2's
acceptance run.

**Phase 1 — eyes.**
1.1 browserrunner: `capture_screenshots` + `profiles [desktop, mobile]` on the render-audit
path; full-page PNG per (url, profile) unconditionally when asked; uploads under
`design-critique/<site>/<run>/…`; `Screenshots` in the result; `max_pages` caps count.
1.2 Vision seam (new shared seam — council + register same commit): `AIService.GenerateWithImages`
implemented for anthropic AND gemini (the model trial is a config switch; ollama typed
not-supported); new `execute_vision_prompt` action (URIs → bytes via storage client; same
overlay/model config + llm_call_logger path as `execute_llm_prompt`).

**Phase 2 — the 018 critic.** New `design-critique-agent`, third spawn/call pair in
design-audit-agent, internal due-gate (fingerprint/14d → else `complete_not_due`):
`ensure_site_record → check_critique_due → request_render_audit (screenshots, ≤8 pages × 2
viewports) → load_design_context (design_intent, composed palette, section inventory, fleet
homepage-skeleton summary) → critique (execute_vision_prompt) → write_findings
(audit_source='design-critique') → record_critique → complete`. Three-altitude prompt (site max
3 / component max 6 / detail max 3, 12/site cap), closed category vocabulary matching the
router exactly, findings on the proven contract (acceptance_test, affected_component,
max_fix_attempts…). Routing: component_css/detail_css → `design_css_fix` → css-patch-agent
(+ real verifier); spacing/responsive/cta → existing fixers; imagery_density → needs_imagery;
page_composition → `needs_page_recompose` (enable only once A3.1 lands); brand/palette →
`needs_design_review` → webdesign-agent with pin-on-apply; chrome → capability_gap until A4.4;
new `unknown_category_policy: capability_gap` for this audit_source. Model trial across 2–3
sites, ruling recorded here.

**Phase 3 — recompose drain + 016/017.**
3.1 `page-recompose-handler`: claim → load page/sections/candidates (join tables, never
usage_count) → LLM bounded recompose (≤2 section changes, locks retained, archetype constraints)
→ Go `validate_recompose_plan` → apply via existing plan/section path → rerender item → complete.
Real verifier: sections differ as acceptance names. DES-038 boundary: selects from registered
components only.
3.2 Revive 016: routable categories only; due-gate (brief change + 30d); wire into
design-audit-agent. **Acceptance = bug 115's finding 2 travelling detected → triaged →
recompose → a page whose section plan differs.** Then experimental → active.
3.3 017's checks: `check_interactivity_promised` → needs_page_recompose;
`check_generic_section_sole_candidate` → capability_gap; fleet dormancy inventory feeds A4.1
steering + the fixloop digest, not work items.

**Phase 4 — anti-brochure compose-time.**
4.1 Per-archetype structural constraints in build-site-planner prompt (required AND forbidden
sections per site_type; anti-sameness instruction; dormant-component steering by exemplar).
Council/owner-flagged: site_type vocabulary extension, core-slot changes.
4.2 `check_composition_convergence`: homepage section-sequence similarity ≥0.8 vs a different-
proposition site, or shared layout+palette+typography triple → needs_page_recompose naming the
twin; template-level residue → capability_gap.
4.3 Churn fix: `check_palette_unpinned` → `palette-pinner` handler (composed core slots →
reference_values; gated on contrast-clean). Lands before any sweep-driven webdesign volume.
4.4 Chrome fork-to-site-scope handler (broad-autonomy completion; riskiest, last; pin predicate
not pool predicate; LANDMINES :690/:2801 first; council).

**Phase 5 — enablement (manual-trigger mode).** Hand-fire per site; checks into
design-discovery-agent's array only after their image rolls; per-category critic enablement
follows its handler; **G1 (flip improvement-sweep on) deferred to a separate owner go.**

### Programme B — offer/benefit analyser (after A)

B1 Widen `site-review-agent.load_strategic_context` to load strategy/identity/content_direction/
mission_brief/audience; prompt judges against primary_model. Verify by planted marker reaching
the assembled prompt.
B2 De-mine domain-strategist (refresh_mode/deployed → complete WITHOUT create_next_item — else
every refresh enqueues a rebuild chain) + four Q-fields in the strategy shape
(satisfaction_condition, money_flow, recurring_value_hook, trust_threshold) + refresh-preserves
instruction. Verify on loancalculator.co.uk: new strategy row, zero needs_briefing items.
B3 `check_premise_incomplete` → `needs_strategy` → domain-strategist (fills the "none"/"implied"
premise on the live estate); `check_revenue_shape` (doc 028's rule: tools/advertising → no
service-selling CTA lexicon → revenue_shape_cta → component-template-fixer, residue →
capability_gap; lead-gen/direct → nav-reachable contact/quote page with a form →
missing_conversion_path → content-gap-planner; affiliate → one capability_gap, machinery absent;
locked/adopted → needs_human_review only). Into quality-discovery-agent's array after the roll,
observe-only first.
B4 `offer-analyser` agent (config-only): load_premise → load_offer_surface (ALL active pages) →
run_offer_analysis → write_findings (audit_source 'offer-analysis'). Judges need↔offer fit,
the benefit table (promised → delivered? surfaced?), monetisation shape, improvements — closed
vocabulary emitting EXISTING item_types only. Grades against the site's own proposition
coordinates. Prompt forbids "users want…" phrasing — artefact-vs-premise only (zero outcome
data; cheapest future instrument = the disabled intent-collection task, named not scoped).
Due-gated inside improvement-loop before triage_findings.
B5 End-to-end proof on webdesign.co.uk + both planted-marker proofs; enrolment order = owner
calls at the time; lendzy only after the positioning lane's marker test passes.

### Item-type / handler / verifier summary (new or changed)

| item_type | detector | handler | verifier |
|---|---|---|---|
| contrast_failure | render-audit tail | css-patch-agent | classified (next audit + dedup key) |
| design_css_fix | critic css categories | css-patch-agent | NEW (patch CSS contains named property) |
| needs_page_recompose | critic, 016, interactivity, convergence | page-recompose-handler (NEW) | NEW (sections differ) |
| palette_unpinned | check_palette_unpinned | palette-pinner (NEW) | NEW (reference_values present) |
| needs_design_review (directed palette) | critic brand | webdesign-agent + pin-on-apply | existing |
| needs_strategy | check_premise_incomplete | domain-strategist (live) | NEW (strategy row with primary_model) |
| revenue_shape_cta | check_revenue_shape | component-template-fixer | NEW (lexicon re-scan) |
| missing_conversion_path | check_revenue_shape | content-gap-planner | NEW (page+form query) |
| capability_gap | chrome (until A4.4), layout swap, affiliate, sole-candidate, non-convergence, off-vocabulary | none — roadmap by design | classified |
| offer LLM findings | offer-analyser | existing handlers only | existing |

### Programme-level acceptance

1. Bug 115's own test: brief-fidelity finding 2 → a page whose section plan differs.
2. A render-audit contrast failure → css-patch → next render audit clean.
3. A critic design_css_fix → applied → verifier passes → re-critique does not re-raise.
4. Two archetype-different sites structurally diverge after A4.1/A4.2 (section lists compared).
5. One offer finding end-to-end on webdesign.co.uk + both planted-marker proofs.

## Corrections to Programme B (2026-08-08, from `REVIEW_2026-08-08_…`)

Marked here rather than edited into §B above, so the original text stays readable.

- **B2's four Q-fields are a RESTORATION, not an invention.** `satisfaction_condition`,
  `trust_threshold`, `recurring_value` and the money-flow field are the shape
  **gaswholesalers.com** has carried since 2026-04-17 (its whole strategy spec is those six
  keys — `satisfaction_condition, trust_threshold, recurring_value, monetisation,
  primary_intent, visitor_type` — and it is the only site with it; the other 16 carry the
  current 12-key shape). Copy the live row rather than designing from the plan text.
- **`primary_model` is NOT a defect in this plan — do not "fix" §B1 or §B3.** A first pass of
  the 08-08 review claimed the field had never been written and called those two lines
  defective. **That was wrong**: it exists on **16 of 17** sites, nested at
  `revenue_models.primary_model` (the shape `domain-strategist`'s own prompt declares). The
  reviewer had read the top level only. Recorded because a future reader who finds only the
  first claim would "correct" two correct lines. `WRONG_CALLS.md` 2026-08-08.
  - Distribution, worth having: `direct_business` **10**, `saas_tools` 3,
    `display_advertising` 2, `lead_generation` 1, `sponsored_listings` 1, absent 1. So §B3's
    `check_revenue_shape` has a real population on day one.
- **`needs_strategy` already has a live producer** (`vertical-exemplar-researcher` →
  `domain-strategist`; 3 complete rows). §B3 makes this lane the **second** producer on an
  existing `item_type` ⇒ the **owner ruling 2026-08-02 / RFC_010 §1** applies: no architecture
  round, **but the concept-register entry must name the full producer set and state the shared
  `item_key` shape**. Not optional.
- **B2's hazard is verified three links deep, not one**: `domain-strategist.create_next_item`
  (unconditional) → `needs_briefing` → `build-briefing-agent` → `needs_site_plan` →
  `build-site-planner`. All 3 historical `needs_briefing` rows are greenfield builds, where
  the chain is correct; it has never run against a deployed site.
  [UNVERIFIED beyond the planner — `build-site-planner` files no further work items.]

## Decision log

- **2026-08-02** — programme approved by the owner with the four decisions tabled above.
  Chosen over: re-enabling the sweep now (owner kept the 07-29 stop in force); offer-first
  ordering; single-model critic. Broad autonomy chosen over detail-auto/structure-review —
  the two stated exceptions are platform constraints, not authority limits.
- **2026-08-08 — OWNER DECISION: B1 and B2 jump the queue.** This **partially reverses the
  08-02 "designer first" build order**, for these two items only; the rest of Programme B
  stays behind Programme A, and A's own order is unchanged. Reason: neither B1 nor B2 depends
  on anything in Programme A, both are `agent_definitions` config (live on apply, no image
  roll, no committed-but-inert window), and each fixes something already wrong — B1 because
  `site-review-agent` asks the offer question ~16×/fortnight with no premise in context, B2
  because a premise refresh on a deployed site would re-plan it. Taken after
  `REVIEW_2026-08-08_…`; the owner also directed the wider scope be written up as
  `features_open/030_FEATURE_offer_and_benefit_analyser.md`, which now holds the
  correspondence surface, the council question and the open questions.
- **2026-08-11 — OWNER DECISION: A-track next, not B4.** Resolves the "B4 — the analyser
  itself, or A-track. Owner's call, unchanged" fork carried in every handoff since 08-08.
  Driven by an external ask, not this lane's own queue: `portfolio_positioning` wants
  visual design diversity across the finance-domain estate (lendzy reads as generic
  AI-designed; every site so far has run the same design model). Programme A's next step,
  **Phase 2 (the `design-critique-agent` / "018 critic")**, is the machinery that ask
  actually needs — Phase 1's vision seam (proven for both Gemini and Claude) is already
  built. **Scope note for whoever picks this up:** Phase 2 as specced in this file critiques
  a design against the site's OWN declared `design_intent` plus a fleet homepage-skeleton
  summary — it does NOT reference external "well-designed" sites or named designers. If the
  reject-and-retry loop is meant to judge against outside taste rather than internal
  consistency, that's a scope addition to `load_design_context` (a reference corpus), not
  existing plumbing — decide before building, don't discover it after. B3's stuck council
  round and its four findings are UNCHANGED by this decision — still owed, separately.
- **2026-08-11 (evening) — OWNER DECISION REVERSED THE SAME DAY: B4 next, not A-track.**
  Recorded here 2026-08-12, because the entry above was left standing and this file is
  supposed to be where decisions live. The reversal is in
  `HANDOFF_2026-08-11_continue_here.md` §"Owed": *"Next track: B4 — the offer analyser
  itself. Not A-track. B1–B3 are live, the estate is swept, and the inputs the analyser
  needs now exist on every deployed site."* Read the two entries together: the A-track
  argument (portfolio_positioning's design-diversity ask needs the 018 critic) was not
  withdrawn and still stands on its own merits — it was outranked, not refuted. **A reader
  picking up the A track should treat the entry above as live scope, not as a dead decision.**
- **2026-08-12 — the B4 premise that justified the reversal is 32% TRUE, measured.** The
  handoff's *"the inputs the analyser needs now exist on every deployed site"* is right about
  `revenue_models.primary_model` (22 of 22 current strategy rows carry one — that is what B3
  drove to completion) and **wrong about the fields B4 actually needs to judge an offer**.
  §5.4 of `features_open/030` calls these the Q-fields; B2 restored them; they are on **7 of
  22 sites**, and the boundary is not a property of the sites:

  | | sites | source of the current strategy row |
  |---|---|---|
  | HAS `satisfaction_condition` + `trust_threshold` + `recurring_value` | **7** | 6 `domain-strategist` (all written 08-08 or later) + 1 `operator` oneshot |
  | lacks them | **15** | 13 `domain-strategist` (all written 08-02 or earlier) + **2 human-authored** |

  **Every strategist-written row since B2 shipped carries the Q-fields; every one written
  before it does not.** So the gap is a vintage, not a defect, and it closes by refreshing a
  premise — which is exactly the operation B2 made safe on deployed sites (that was B2's
  whole point, and it has not been used for it yet).

  ⚠ **The refresh is NOT a blanket sweep, and the exclusion list is short and load-bearing.**
  Two of the fifteen carry **human-authored** current specs — `mortgagecalculator.co.uk`
  (`source='owner_direction'`, created_by *"owner direction 2026-08-11 (customer voice,
  softer, no clever titles, not competitive)"*) and `leopardessconsulting.co.uk`
  (`source='hitl'`). A `domain-strategist` refresh writes a new `is_current` row and
  supersedes whatever is there, so a 15-site sweep would **overwrite the owner's own
  direction on mortgagecalculator, one day after he gave it**. The 13 strategist-written rows
  are safe; those two need his call, and it is a different question (re-elicit, hand-merge,
  or leave un-analysable).

  The query, so this is re-runnable rather than quoted:
  ```sql
  SELECT (sp.data ? 'satisfaction_condition') AS q_fields, sp.source, count(*),
         string_agg(s.domain, ', ' ORDER BY s.domain)
  FROM site_specs sp JOIN sites s ON s.id = sp.site_id
  WHERE sp.aspect = 'strategy' AND sp.is_current
  GROUP BY 1, 2 ORDER BY 1 DESC, 3 DESC;
  ```

  **What this changes for B4's design, before a line of it is written.** The analyser cannot
  assume its richest inputs are present, and the two honest options are not equivalent:
  (a) **refresh first** — 13 dispatches, then B4 sees a uniform estate, at the cost of 13
  strategist runs and a week's latency; or (b) **B4 degrades explicitly** — judge on
  `value_proposition` alone where the Q-fields are absent and SAY SO in the finding, which
  keeps B4 unblocked but means its verdicts are not comparable across sites. **(b) has a trap
  this lane has already been bitten by twice: a check that examines less on some sites and
  does not say so produces a silence that reads as a clean bill** (WII-014 is the entire
  fix for one instance of it, `bugs_open/255` the other). If B4 degrades, the degraded
  verdict must be a stated field on the finding, not an absence.

  **Recommendation for the owner: (a) for the 13, then B4.** It is the same shape of decision
  he already took on 08-11 for the three sites with no premise at all, the machinery is
  proven (three of three dispatches produced correct premises, two of them unattended), and
  it means B4's first real verdicts are comparable across the estate instead of carrying an
  asterisk per site. The two human-authored sites stay out until he says otherwise.
- **2026-08-12 (evening) — OWNER SAID GO, and the refresh is DONE: 13 of 13, gate held, B4 is
  unblocked.** Executes the recommendation in the entry above. **Q-fields now on 20 of 22
  sites** (19 strategist-written + LMC's operator row); the only two without are exactly the two
  excluded by the `source` filter — `leopardessconsulting.co.uk` (`hitl`) and
  `mortgagecalculator.co.uk` (`owner_direction`), which remain the owner's call.
  Vehicle: `domain-strategist` oneshot envelopes, canary → 6 → 5 → 1, each disabled immediately
  after firing. Evidence and the two missteps are in NOTES (2026-08-12 evening).
  **Three things this establishes that were previously assumed:**
  1. **B2's gate works for the case it was built for.** It had only ever been exercised on
     sites with NO premise; this is the first refresh of an existing one. Zero `needs_briefing`,
     zero `needs_site_plan`, zero work items of any type across all 13 — with today's greenfield
     build (noted.co.uk) as the control proving the `else` arm still fires when it should.
  2. **A refresh is STABLE: 12 of 13 kept the same `primary_model`.** The strategist re-derives
     the same commercial answer from the same site, so the operation adds the Q-fields without
     churning the premise. **This is what makes it repeatable** — nobody had measured it, and a
     third of the estate changing shape would have made this a one-off gamble rather than a
     maintenance operation.
  3. **The one change is informative, not a fault:** dartsonline.com `direct_business` →
     `affiliate`, which re-points `check_revenue_shape` to the affiliate arm on its next
     examination. Predicted consequence recorded in NOTES for the next session to verify.
  **B4's precondition is now met.** The "refresh first or degrade explicitly" fork in the entry
  above is closed in favour of the first, and B4 can assume a uniform estate on the 20 — with
  the two human-authored sites as a stated, named exception rather than a silent gap.
- **2026-08-13 — CORRECTION to the entry above, and it is a correction to my reasoning, not to a
  number.** That entry says *"a refresh is STABLE: 12 of 13 kept the same `primary_model` … **this
  is what makes it repeatable**"*. The measurement stands; the inference does not.
  **Classification stability is not prose accuracy.** I measured whether the strategist re-derives
  the same commercial answer for a site. I did not measure whether the sentences it newly wrote
  are true — and on 08-13 a donor run for leopardessconsulting.co.uk produced a
  `recurring_value` asserting a **twice-weekly technical blog that does not exist** (6 posts in
  ~4 months, on entirely different subjects). So the refresh can be stable in the sense measured
  and still import invented specifics. **Read that entry as "the refresh does not re-roll a
  site's revenue model", which is what it actually established, and not as "the refresh is
  safe".** The 13 refreshed premise records have never been claim-checked; a 3-site eyeball
  found no leopardess-class falsehood, which is a sample, not a check.
- **2026-08-13 — the two hand-written premises: ONE MERGED, ONE REFUSED (owner approved the
  merge).** Method: the **strategist writes and this lane merges only the three Q-fields**,
  discarding the rest of the donor run — because hand-authoring them would breach the 2026-08-06
  content ruling AND, more to the point, would make B4 grade each site against a standard a
  session invented. The merge is one atomic `DO` block with `RAISE EXCEPTION` guards (a verify
  block of bare `SELECT`s cannot stop a `COMMIT`), the load-bearing one being
  **`md5(merged − the three added keys) = md5(protected row)`** — mechanical proof that nothing
  existing was reworded or removed.
  - **mortgagecalculator.co.uk — merged.** Current row still `source='owner_direction'`, now
    carries all three fields, and the stripped-back md5 equals the pinned before-state exactly.
    The owner's voice direction of 08-11 is untouched.
  - **leopardessconsulting.co.uk — refused, restored, and the refusal is the point.** Its donor
    prose failed the claims gate (above). The `hitl` row is current again at its pinned md5;
    donor demoted and retained; ~3-minute exposure window checked clean (one orchestration on
    that site — the donor run itself — and zero work items). **Its Q-fields remain absent, by
    decision.** Options for the owner: merge only the two clean fields
    (`satisfaction_condition`, `trust_threshold` read clean) and leave `recurring_value` absent;
    or leave all three out and let B4 carry a stated exception for that one site.
  - ⚠ **A banned-term screen is not a claims check.** The regex built from the 2026-07-16 ruling
    (`70+`, `N departments`, `managing agent`, least-privilege) passed this prose cleanly. A
    banned-term list records what was already caught; the next invention uses different words.
    Reading it, then checking the one checkable sentence against `pages`, is what found it.
- **2026-08-13 — OWNER DECISION: yes, support affiliate properly, with dartsonline.com as the
  worked example over the coming days.** Answers the largest open item from the 08-12 read-out.
  Scope is **not** "which retailers to link to" (that is dartsonline's own lane, and the owner is
  asking them separately) — it is the platform capability the estate currently lacks and which
  `check_revenue_shape`'s affiliate arm has been filing against since 08-09: link management,
  disclosure blocks, partner config, and enough of a shape definition that the offer checker can
  examine an affiliate site instead of filing *"I have no way to check or fix this"*. **Three
  sites are waiting on it** — dartsonline.com, loancalculator.co.uk,
  loanandmortgagecalculator.co.uk — each currently carrying an undispatchable
  `capability_gap:revenue_shape` row (`gap_kind=handler_missing`). Those three rows are the
  requirement list, and closing the capability is what retracts them. Not started; not this
  lane's next action unless the owner says so (B4 remains the stated next track).
