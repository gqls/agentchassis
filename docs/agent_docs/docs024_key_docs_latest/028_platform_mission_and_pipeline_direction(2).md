# 028 — Platform Mission and Pipeline Direction

**Status:** Living document. Second revision (2026-04-22). Iterating.

This document anchors the *why* of the website-building pipeline so that future sessions, agents, and changes can check themselves against it. It sits alongside `021_site_spec_and_classifier.md` (the "one spec, not two" architecture), `002_system_architecture.md` (agent layout), and `007_adoption_pipeline_v2.md` (adoption mechanics). Where those explain *how*, this explains *what we're trying to achieve*.

---

## The mission

Given a list of domain names — empty, partially developed, or already hosting a live site — produce the best possible website for each, end to end, with minimal human input. "Best" means the site that most probable visitors will find most useful, measured through real engagement, and that also generates the best revenue for us through whatever model genuinely fits the domain.

The pipeline is a single pipeline. A blank domain, a domain with an existing site, a domain with a user-provided mission, and a domain whose owner wants faithful replication all travel through the same agent graph. What differs is the input material the strategic brain has to work with and the fidelity dial that controls how aggressively to extend that material into a full aspiration.

This is the frame every other architectural decision sits inside. When an agent writes or reads a spec, it should be doing so in service of this mission. When a prompt makes an assumption, the assumption should be compatible with this mission. When behaviour drifts, we check against this document.

---

## Commercial viability is not the same as a "business site"

We want to make money from the sites we build. That is a given. It does not follow that every site is a consultancy with an about page, a services page, and a contact form. The revenue model shapes the site, not the other way round.

Revenue can come from direct services (the consultancy shape), from ad-supported tools and content (the `gamedesign.uk` shape — reference material and interactive tools, no services page), from affiliate-driven comparison and research (the `vetcomparison.uk` shape — data-driven listings and reviews), from software-as-a-service with marketing copy supporting a product, from publication and media, from subscription-gated content, or from mixed combinations. Each shape is a legitimate answer to "how does this site earn?" and each produces a completely different site. Defaulting to the consultancy shape when the signal is absent is a failure mode, not a safe fallback.

The classifier's job includes deciding which revenue model fits the domain — informed by what's already there if anything, by the vertical's norms, by competitive research, and by the domain name itself. Downstream planning should then commit to the shape that fits. Mixing shapes (a tools site with a "Start a Project" CTA, for example) is a sign that either the classification is vague or a downstream agent is ignoring it.

---

## The classifier is the strategic brain — it always runs in full

The classifier — currently `domain-research-classifier`, later likely richer — is the agent that decides what this site *should be*. It runs on every site entering the pipeline, and it always does its full job: research, synthesis, aspirational spec writing. Adoption does not shortcut it. An operator-provided mission does not shortcut it. They are inputs that weight its reasoning, not bypasses that skip it.

The classifier is not constrained to what can be built today. If the best version of this site requires a component the library doesn't yet contain or an agent that hasn't been written, the classifier describes it anyway. Those items are marked `blocked` in the spec. The `feasibility-recheck` task promotes them to `planned` when the necessary capability comes online. The spec describes where we're going; the build is whatever subset of the spec the system can currently deliver and is configured to deliver given the site's fidelity setting. This is the "one spec, not two" principle from doc 021, extended with the fidelity mechanism described below.

The classifier's outputs are read downstream as direction, not suggestion. When the planner, composition agents, content writer, or design agent can't fully implement an item, they build what they can, mark the rest, and surface the gap — they do not substitute their own preference. Disagreement with the classifier is resolvable through notes, HITL review, or an explicit direction update. Silent override is the failure mode we are trying to eliminate.

---

## Input sources and their weight

The classifier reads the same spec shape and produces the same spec shape regardless of how the site entered the pipeline, but the material feeding into its prompt differs. Each input has an associated weight — how strongly it should steer the output. The weights should be controllable rather than hardcoded in prompt phrasing.

**Operator mission or brief.** A human has supplied a strategic mission or objective in the input data. This is the owner's stated intent. Weight: very high. The classifier validates it against research but does not contradict it.

**Adoption output.** The site-adoption-agent has crawled an existing live site at the source URL and captured its current identity, archetype, content direction, design intent, pages, and page content. Weight: very high for "what this site IS" (identity, character, palette). Moderate for "what this site should do next" (services, extensions, features) — because the adopted state describes the present, not the ceiling.

**Operator-provided reference sites.** The operator has flagged specific sites as inspiration. Weight: high for structural and stylistic patterns, moderate for content specifics. References inform what the target should be; they do not become the target directly.

**Competitor and vertical research.** The classifier searches the web and scrapes candidates within the same niche. This is always available and always consulted, even when adoption or mission provides strong signal. Weight: medium — directional rather than prescriptive. It identifies patterns peers demonstrate that this site could adopt, gaps in the vertical that this site could fill, and norms that readers in this space expect.

**Domain name and TLD.** A minimum signal, always present. Weight: low on its own — enough to seed research, not enough to commit to a direction without supporting evidence.

Across all five, the classifier produces one spec. The difference is evidence weight and confidence, not spec shape. A site with all five inputs should use all five. A site with only a domain name should still produce a spec, with the classifier's uncertainty reflected in the confidence field and in how aspirational versus conservative the output is.

The fact that adoption is an input rather than a replacement is important. Even when adoption has run, the classifier still does vertical research. The adopted state describes where the site is; competitor research informs where it could go. Those are different questions with potentially different answers, and the spec captures both. What gets deployed at launch is controlled by the fidelity dial, not by silently trimming the classifier's output.

---

## Fidelity — controlling how much aspiration reaches the first build

Fidelity is the dial that controls how closely the first deployed version of a site matches the strongest input evidence (usually adoption), versus how much aspirational extension the system builds on top of that evidence at launch.

The classifier produces the full spec in all cases — adopted baseline plus aspirational extensions informed by vertical research. Fidelity then controls which subset of that spec the build pipeline actually realises at first deployment, and how aggressively the improvement loop narrows the gap over time.

Five values, with `high` as the default when adoption evidence is present:

**`locked`** — First build matches adopted evidence exactly. No aspirational extensions are built, and the improvement loop does not promote planned items into deployment. The classifier may still record aspirational extensions in the spec for human reference, but nothing moves from planned to deployed without operator action. This is the "I own this, don't change it without asking me" case.

**`high`** (default when adoption is present) — First build matches adopted evidence closely. Aspirational extensions live in the spec marked `planned` but are not built at launch. The improvement loop promotes them slowly — roughly one substantive change per audit cycle, respecting constraints.

**`medium`** — First build preserves adopted character (identity, palette, voice) but includes modest aspirational extensions at launch where they don't conflict with the adopted baseline. A medium-fidelity adoption of gamedesign.uk would launch with the adopted tools and guides, plus perhaps one or two extensions that peers demonstrate are valuable (e.g. a community-links footer, a contributor page). Improvement loop promotes faster.

**`low`** — Adoption is inspiration only. First build draws heavily on aspirational extensions; adopted evidence constrains palette and general character but not structure or feature set. Closest to what the pre-fix behaviour accidentally produced, but now as an explicit choice.

**No adoption** — The concept shifts. There's no adopted baseline to be faithful to, so fidelity instead controls confidence tolerance: high means "build only high-confidence items from research"; low means "build the full aspirational spec including less-certain extensions." Default on a blank domain: medium.

The value lives on the adoption trigger input and in the spec (likely in a new `adoption_meta` or `build_policy` aspect). The classifier reads it and modulates both the aspirational-extension production and the prompt's framing. The build pipeline reads it to decide which spec items to realise at first deployment. The improvement loop reads it to decide promotion rate.

Implementation note: the fidelity dial depends on per-item status on specs (see next section). Without that, fidelity can only operate at the coarse level of "prompt the classifier to be more or less aspirational." With it, fidelity operates correctly as deployed-vs-planned partitioning, which is the proper model.

---

## The spec has status — deployed, planned, blocked

Doc 021 introduced the principle: "one spec, not two. Items have status (deployed / planned / blocked), not separate documents. The dream is the full spec. The build is the non-blocked subset."

This principle is central to how fidelity works. It is not fully implemented yet. Current specs contain fields without per-item status — the build pipeline builds whatever the spec says, and whatever isn't in the spec doesn't exist. Per-item status (planned to be implemented in Phase 2 below) makes the dream-vs-build distinction mechanical rather than conceptual.

Each spec row records `source`, `source_agent`, and `source_item_id`. This lets us reconstruct what each agent contributed and in what order, which is how we diagnose issues like the adoption→classifier clobber that motivated this document. Status per spec item extends this: we'd also know which items are currently live, which are queued, which are waiting on capability that doesn't exist. An aspect rewrite should happen because the spec is changing, not because an agent forgot to check. Status changes are the normal flow; spec rewrites are the exceptional flow.

The versioning mechanism (`is_current`, `superseded_at`) handles intentional spec evolution. Silent overwrite — one agent rewriting another's output without reading it — is the failure mode we are eliminating.

---

## Who writes what, who doesn't override

Every agent in the pipeline either reads the spec, writes to the spec, or both. The spec is the contract between agents. Kafka messages move work items between pods, but the authoritative state lives in `site_specs` rows. This has consequences.

An agent that changes behaviour based on information it didn't put in the spec is a bug, because a different agent won't see that information. An agent that overwrites a spec aspect that another agent is expected to produce is a category error, because it breaks the implicit ownership model. An agent that produces the right output but doesn't write it to the spec is not helpful, because downstream agents rely on reading it back.

Ownership specifically: the classifier owns `identity`, `classification`, `content_direction`, `design_intent`, `seo`, `maintenance` aspects and should be the only writer under normal flow. Adoption owns `site_archetype` and `design_reference` aspects (nobody else writes these). The strategist owns `strategy` aspects. Planners own `site_plan` aspects. When we write a new agent, we commit to which spec aspects it owns, which it reads, and which it merely extends.

When adoption has produced an `identity`, `content_direction`, and `design_intent` and the classifier then runs, the classifier *does* update those aspects — but its prompt now reads the adoption output as ground truth for the adopted dimensions (palette, voice, character) and produces output that preserves those while adding the classifier's own strategic contributions (classification, extended content_direction fields, design_intent guidance). This is not an override; it's a read-and-extend.

---

## Spec evolution over time

Finders find gaps between spec and deployed state. Fixers fix them. Each fix moves a spec item from `planned` to `deployed`. Over months, the deployed state converges on the aspirational spec.

The improvement loop's rate is configurable per site. A locked or high-fidelity site improves slowly; a low-fidelity site improves aggressively. The strategist can also update the spec directly — adding new planned items, retiring ones that no longer make sense, changing direction in response to operator input. Human direction (doc 007) pins specific items as inviolable.

Adoption is a one-time capture. The classifier produces a one-time initial spec. Everything after that is the normal evolution loop — spec changes from strategists and humans, status changes from finders and fixers, with versioning preserving the history.

---

## Phased implementation

This document describes a target state. Not all of it is implemented. The phasing matters because it explains what today's adoption-aware classifier prompt does and doesn't do, and why we can't yet produce the full unified-model output safely.

**Phase 1 (current, 2026-04-22).** Adoption-aware classifier prompt (migration 006). When adoption output is present in site_specs, the classifier reads it and produces a faithful first-draft spec that preserves adopted identity, archetype, content direction, and design intent. No aspirational extensions in the output — because the build pipeline would build them immediately. Fidelity is implicit `high`. Adoption sites launch close to their source; blank sites launch from research. This is correct for the "I own this, want a copy" case but under-uses the classifier's strategic capability.

**Phase 2 (next substantive work).** Per-item status on specs. Schema extension to let individual spec fields, pages, sections, services, and features carry `status: deployed | planned | blocked`. Build pipeline respects status — only builds `deployed` items. Backward compatible: existing specs default every item to `deployed`. The `feasibility-recheck` task already exists conceptually (doc 021) and promotes blocked → planned when capability comes online; this phase makes it real.

**Phase 3.** Explicit fidelity input parameter. Stored on the adoption trigger and in a new spec aspect (`build_policy` or `adoption_meta`). Classifier reads it and modulates prompt framing. Build pipeline reads it to decide what subset to realise at launch. Improvement loop reads it to set promotion rate.

**Phase 4.** Classifier prompt extended to produce aspirational extensions alongside the faithful adopted baseline, marked with appropriate status. This is where competitor/vertical research fully pays off — the classifier becomes the strategic brain described above, writing full aspirational specs that include adopted ground truth and vertical-informed extensions. Until this phase lands, the classifier's aspiration stays conservative when adoption is present.

**Phase 5 and beyond.** Multi-source aspiration — classifier merges adoption, reference sites, competitor research, and mission into a single weighted synthesis. Confidence expressed per field. Downstream agents react to confidence (high-confidence items get built faster, low-confidence items accumulate more audit oversight before deployment).

The current adoption-aware prompt patch (migration 006) lives in Phase 1. It's correct for Phase 1 and will need revision when Phase 2 and 4 land.

---

## Failure modes we want to eliminate

Several concrete failure modes should be avoided, and future changes should be checked against them.

*Silent overwrite.* An agent writes a spec aspect over an earlier agent's output without reading or weighting that earlier output. This was the classifier-over-adoption bug.

*Confident fabrication on thin signal.* On a blank domain with no research hits, the classifier produces an elaborate identity and set of services as if they were evidence-backed. This misrepresents aspiration as certainty. Guesses are fine; unflagged guesses are not. Confidence fields exist for this reason and should be populated honestly.

*Default-to-brochure bias.* When the site shape isn't clear, fall back to consultancy/services/contact. Our tooling is biased toward this shape because that's what most sample prompts and examples assume. This needs active countering — the enum already includes `tools`, `content`, `interactive-platform`, `ecommerce`, `landing`; the prompts should use them when evidence points there.

*Reflexive re-running of upstream agents.* A self-heal mechanism queues the classifier when a spec is missing. This is sensible when the spec is genuinely missing. It is not sensible when the classifier has just run and its output is still current. The self-heal needs to distinguish between "no strategic output exists" (queue the classifier) and "strategic output exists but not in the specific aspect this downstream agent wanted to read" (resolve locally, surface a gap, or migrate the data).

*Schema-level commercial bias.* A prompt schema that always contains fields like `services`, `cta_style.verb_choices`, `persuasion_approach`, or `social_proof_style` forces the LLM to populate them even for sites where they make no sense. Fields that apply only to commercial/persuasive sites should be conditional on the site's determined commercial shape.

*Adoption without strategic analysis.* An adopted site that never receives vertical/competitor research produces a faithful replica but no direction for evolution. The improvement loop has nothing to improve toward. This is the failure mode the unified-model framing prevents — adoption is an input to strategy, not a substitute for it. Phase 4 is when this becomes concrete.

*Aspirational extension without status.* The classifier produces a rich aspirational spec including features that can't yet be built or that the operator hasn't approved, and the build pipeline builds them anyway because there's no status mechanism to filter. Phase 2 prevents this.

*Superseding a spec doesn't undo what that spec installed.* Some agents write more than a spec row. When site-design-planner finishes resolving a composition, it writes four things: the `resolved_composition` spec, a new `css_themes` row, a new `style_collections` row, and a pointer on the `sites` record (`sites.style_collection_id`) that points at that style collection. Page renders don't read the spec — they read the style collection via that pointer. That matters because marking the spec `is_current=false` doesn't touch the other three. The pointer still points at the old theme. Pages still render against it. And because no work item has been created asking the planner to re-run, the planner has no reason to run again.

The gamesdesign.co.uk remediation hit this. Clearing the stale `resolved_composition` spec was one line of SQL; what we didn't do was queue a `needs_composition` work item. Result: specs upstream (classification, content_direction, design_intent) correctly refreshed by migration 006, but every page deployed afterwards was rendered against the pre-migration brochure-formal CSS theme that was still installed at the `sites.style_collection_id` pointer. The build pipeline was working from fresh specs and stale installed artefacts at the same time.

Treat this as a rule: whenever we invalidate a spec that was produced by an agent with install side-effects (composition, nav, pages, assets), we also queue the work item that re-runs that agent. The better long-term fix is structural — the supersession itself should emit the recovery work item, not rely on the operator remembering. For now, manual invalidations need the re-queue step applied by hand, and the remediation SQL should show it.

The test for whether an agent falls into this pattern: does the agent write to tables other than `site_specs`, AND is there a live pointer from a long-lived table (`sites`, `pages`, `navigation_structures`) into what it wrote? If yes, spec invalidation alone will never propagate to the build pipeline.

---

## How this document gets used

When a new agent is being designed, or an existing agent is being changed, the proposal is checked against this document. Does it respect the classifier's role as strategic brain? Does it read what's already in the spec before writing? Does it write to aspects it owns rather than aspects another agent owns? Does it surface gaps rather than silently adjusting the spec to match what it was able to produce? Does it respect the current phase — for instance, not producing aspirational extensions before Phase 2 lands?

When a bug is being diagnosed, this document frames the diagnosis. Is this a spec-contract violation? Is this an overwrite that should have been a read-and-extend? Is this a schema biasing the model toward a shape that doesn't fit? Are we seeing confident fabrication that should have been flagged as low-confidence? Is it a Phase 1 limitation that Phase 2+ will resolve, or a genuine bug in Phase 1's expected behaviour?

When the user points out that the system has drifted, the drift is usually a gap between what this document says and what an agent actually did. The fix targets the gap.

---

## Open questions

Several things this document leaves unresolved. They are flagged here so future sessions address them deliberately rather than assume.

*Build-site-planner writes to `content_direction` and `design_intent`.* Observed in the 2026-04-22 adoption trace. Doc 021 says the planner writes only `site_plan`. Observed behaviour contradicts that. Is the planner recording a committed subset, overriding, or doing something else? Resolve in a focused session.

*Classifier re-run triggers.* Currently the self-heal triggers on missing specs (`validate_composition_inputs`). We haven't specified when a deliberate re-classification is wanted — operator request, vertical shift, significant time passed, user-supplied new direction. Without explicit triggers, re-runs happen only accidentally via the self-heal path.

*Confidence expression and reaction.* The classification output has a `confidence` number. No consumer behaviour is currently keyed off it. Downstream agents should probably treat low-confidence items more cautiously — slower promotion to deployed, more audit oversight, maybe HITL review for first-time deployment. Phase 5 territory.

*Integration of the 15 domain content strategy questions.* Doc 021 maps them to spec aspects, but the current classifier prompt doesn't walk through them explicitly. Integrating them cleanly would sharpen the aspirational output once Phase 4 lands.

*Per-input weighting API.* We currently express input weight through prompt phrasing ("STRONG guidance," "STRONGEST signal"). A structured weight value on each input — available to the prompt and to downstream consumers — would be more controllable. Good follow-up after Phase 3.

*Constraints across aspirational extensions.* The `site_archetype.constraints` array from adoption is inviolable. When the classifier produces aspirational extensions in Phase 4, those extensions must respect the adopted constraints. Mechanism to enforce this (prompt-level, schema-level, or a validation step) is undesigned.

---

## What this document doesn't cover

This is the mission and direction for the website-building pipeline. It is not the place for:

The specific prompt contents (those live in `agent_definitions.workflow.steps[].config.prompt_template` and are edited via SQL with prior backup per doc 009).

The agent graph or messaging topology (doc 002 covers this).

The data flow and schema details (doc 021 covers the spec, doc 011 covers the DB, doc 003 covers message contracts).

The adoption mechanics (doc 007).

The improvement and audit loops (doc 004).

The news, tools, and other vertical-specific pipelines (docs 006, 005, 008, etc.).

This document is read alongside those, not instead of them.
