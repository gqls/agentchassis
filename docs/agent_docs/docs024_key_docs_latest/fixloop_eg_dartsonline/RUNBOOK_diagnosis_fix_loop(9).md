# RUNBOOK — Diagnosis→Fix Loop (v2 of the diagnosis loop)

## THE TASK (read this first if you are new)

The platform already has a working, read-only **diagnosis loop**: given a bug
symptom, an agent forms a hypothesis, gathers scoped evidence (real code
bodies from an indexed corpus, read-only database rows, runtime records),
issues a verdict that must CITE evidence or ABSTAIN, and re-scopes by
FOLLOWING what the evidence names — until it confirms a cause with citations
across all three tiers (static code / live data reads / runtime records). It
is deliberately human-gated: it emits a diagnosis and changes nothing.

**This workstream develops it into a diagnosis→fix system** with, in order:
1. **An easy, documented route in and out**: one clear way to input a task or
   bug; live monitoring of what the loop is doing and why (per-iteration —
   and per-step — reasoning written to a task-specific running-notes file);
   and a usable result out, including the ability to
   **consume/download/fetch the bundles** the loop builds each iteration
   (today they are ephemeral, in-memory).
2. **Fixes on a branch**: the confirmed diagnosis drives a proposed fix
   committed to a separate git branch, so the human can amend, ditch, or
   apply it. The loop's core stays read-only; the write surface is isolated.
3. **A council of reviewers** before any fix is finalised: independent
   specialist agents each judging the proposal from their own perspective and
   sending opinions to a **decision-maker** that weighs them all. Initial
   roster (from the problem owner): a **guidelines agent** (does the fix
   adhere to guidelines 000-0xx — or did the guidelines fall short?); a
   **reuse agent** (are we building a new route where a tried-and-tested
   solution exists — checking BOTH code and docs); a **bug-historian**
   (catch early; record bug categories so the same class never repeats); a
   **compliance/legal eye**; **pipeline guardians** — one per master
   workflow/pipeline (seeded from the builder thread's relay map) — checking
   the fix doesn't infringe on another workflow; and **specialist knowledge
   agents** (e.g. a trigger expert, a site-work-items triage expert) that
   answer "we already have one of these" — the motivating example being a
   chat that composed a trigger + triage SQL which already existed.
4. **Architecture-change visibility**: make it loud when a proposed change is
   accidentally fundamental — touching platform contracts, message shapes,
   many packages, exported signatures — before it ships.
5. **Learning**: recorded bugs, proposed guideline amendments, corpus and doc
   enrichment feeding back in.

**Mission for the tool**: use everything available to reach the right result
— the code corpus, schemas, runtime records, the guidelines themselves — with
checks, balances and second opinions built in.

## What already exists (do not rebuild)

- The **live loop** (chassis `pkg/diagnose` + diagnose_* actions): three-tier
  CONFIRMED diagnosis achieved; engine guards (named-scope narrowing, capped
  call-graph expansion, cite-or-abstain, SQL guard read-only allowlist);
  the §7D resolver (fuzzy scope → real symbols, incl. basename
  canonicalisation). RUNBOOK_code_retrieval_route.md is the closed record.
- **contextkit CLI** (`cmd/bundle` + `cmd/analyse`, RUNBOOK_31_.md): manual
  paste-ready bundles from explicit -scope flags (example_bundle.txt is a
  real invocation). The live loop's assembler is its descendant.
- The **code_symbols corpus** (3,7xx symbols, single current commit; the
  index-orchestrator reindex route) + vector/trigram lookup.
- The **work-item relay + immune system** (builder thread §B2/§B3): a proven
  intake/dispatch/retry mechanism a diagnosis task could ride.
- The **tools chat's travelling-docs infrastructure** — REV-22 READ 2026-07-07:
  doc_plans/doc_notes tables LIVE (Stages 0–2 shipped); the diagnose-agent
  workflow is ALREADY rewired by them: `emit → persist_note
  (config.error_step="complete") → complete`, result_from untouched; the
  subject gate is the action's first check ("no explicit subject — skipping
  (do not guess)"); their 3b (threading subject_type/subject_key through
  call_diagnoser.input_mapping + the trigger) is IN FLIGHT. `load_runtime`
  error-routing is APPLIED (anchorless runs survive via routed degrade —
  ~26 min / 5 iterations observed). Canonical trigger:
  drafts/084_TRIGGER_diagnose_v1.sh (subject fields commented until 3b).
  Their Stages 5–6 define a TIERED ACCEPTANCE system for tools: a static
  Tier-2 contract-presence check and a Tier-4 **browser-runner adapter**
  (Chromium+Playwright, Kafka request/response per the 035 Adapter Guide) —
  their "loop for complicated tools" is acceptance/verification + docs, NOT
  a rival diagnosis loop.
- The **builder thread's pipeline map** (RUNBOOK_builder_route.md §B0–§B3):
  the seed material for pipeline guardians.

## Phased plan (thin slices; pre-registered criteria per slice)

**F0 — Intake, observability, egress (first; test on the user's next real bug)**
- F0.1 Bundle egress: persist each iteration's bundle durably + one
  documented fetch route. (Design question Q-A below.)
- F0.2 Task input: one documented way in. (Q-B.)
- F0.3 Per-task running notes: the loop writes its reasoning per iteration
  AND per step (hypothesis, scope chosen and why, requests issued, verdict
  grounds, resolver substitutions) to a task-specific notes doc. (Q-F —
  likely REUSE the tools chat's doc_notes.)
- Success criteria: a bug goes in via the documented route; the human can
  watch reasoning appear; every bundle is fetchable afterwards; the diagnosis
  lands as today.

**F1 — Fix on a branch**
- F1.1 Fix-proposer (Q-C: who/where) turns a CONFIRMED diagnosis into a
  patch on a new branch via the git adapter; PR opened; human amends/ditches/
  applies. Write-token security isolated to the proposer (the spawn token
  gate pattern exists).
- F1.2 The per-task notes gain the proposal rationale + diff summary.

**F2 — The council**
- Independent reviewers (roster above), each a small agent with its own
  curated context (Q-G), producing a structured opinion (verdict-wire-style
  contract: verdict + citations + objections + suggested alternative);
  a decision-maker aggregates; the human sees diagnosis + proposal + council
  report (Q-H format). Architecture-change detector runs as one reviewer
  (Q-E signals).

**F3 — Learning**
- bug_records (category taxonomy, recurrence checks feeding the historian);
  guideline-amendment proposals routed to the human; corpus/doc enrichment.

## Boundaries
- Tools chat: owns doc_plans/doc_notes + tool docs + its diagnose_load_runtime
  draft — F0.3 reuses rather than reinvents; align next turn on their notes.
- Builder thread: owns the relay/spine; the pipeline map is INPUT here;
  guardian findings that imply relay changes route back through it.
- Quality thread: a future consumer of fixes; no overlap now.

## QUESTIONS — decided vs open (discussion log in NOTES_running_fixloop.md)

DECIDED 2026-07-07 (owner): **Q-A** diagnosis_artifacts table, written
through inside assemble (unified-table refinement: kind ∈ {bundle,
iteration_note}); **Q-F** shape (c) — own working-notes storage, terminal
note only in doc_notes via the tools chat's wiring (envelope must carry
subject_type/subject_key); **Q-B** needs_diagnosis item in a NEW
`pipeline='diagnose'` namespace (null-site allowed; envelope extends 084;
manual trigger retained); **Q-C** separate fixer agent (isolated write
token; constrained edit plan; gofmt+build in a spawned job pre-PR); **Q-D**
flag-based hard veto + guideline-gap SIDE-TASK (amendment PR against the
guideline docs, human terminal, fix unblocked, F3 recurrence record).

STILL OPEN (F2-phase; fine to settle in the new thread):
- **Q-A egress medium** for bundles: diagnosis_artifacts table vs
  orchestration-state collected_data vs object storage vs a git branch of
  artefacts. (Constraint memory: cd bloat already caused a 1.27MB Kafka
  incident; bundles are ~60KB × ≤5 iterations.) doc_notes CONSIDERED AND SET
  ASIDE: notes are prose for humans/agents, bundles are machine-replayable
  evidence with different retention — separate table stands as the position.
  PLACEMENT REFINEMENT (collision-avoidance): implement egress INSIDE the
  assemble action (write-through per iteration) — zero workflow-shape change,
  zero contact with the tools chat's emit-adjacent wiring.
- **Q-B intake**: a `needs_diagnosis` work item (rides dispatch + immune
  system + claims; wrinkle: pure code bugs have no site_id — pipeline
  namespace or null-site allowance needed) vs the manual trigger only vs both.
  ENABLER CONFIRMED 2026-07-07: anchorless (site-less) diagnosis runs now
  SURVIVE — the tools chat's load_runtime error-routing is applied — so the
  LOOP side of code-only bugs is done; only the item-namespace decision
  remains. The envelope adopts their 084 trigger's shape + subject fields.
- **Q-C the fixer**: new agent (distinct responsibility, isolated write
  token) vs extending diagnose-agent; how the diff is produced (LLM patch vs
  constrained edit plan) and validated (gofmt/build in a spawned job?).
- **Q-D council topology — VETO SEMANTICS DECIDED (owner, 2026-07-07)**:
  parallel reviewers → decision-maker BY DEFAULT (all opinions advisory,
  weighed together); a **hard_veto flag**, attachable at multiple scopes
  (a reviewer agent, a pipeline, a specific tool/component), converts that
  reviewer's negative verdict into a BLOCK — accessibility and legal are the
  motivating hard-veto cases. Sub-questions travelling to the new thread:
  where the flag lives concretely (reviewer definition column vs per-pipeline
  council config vs both, most-specific-scope wins?); and whether a
  guidelines-reviewer "the guideline itself fell short" finding blocks or
  spawns a side-task (leaning side-task — it is a gap, not a violation).
- **Q-E architecture-change signals**: packages touched breadth; platform/ vs
  actions/; exported-signature diffs vs the corpus; message/topic/schema/
  contract changes; migration presence. Which are load-bearing?
- **Q-F per-task notes — DIRECTION SET (2026-07-07)**: REUSE doc_notes. The
  terminal-diagnosis note already exists on their side (pending their 3b
  subject threading); F0.3 shrinks to (a) our intake carrying
  subject_type/subject_key first-class (adopt/extend THEIR 084 trigger as the
  canonical envelope — do not write a rival), and (b) per-iteration/per-step
  entries as ADDITIONAL doc_notes rows — VOLUME AND SHAPE NEED THEIR
  SIGN-OFF (relay question logged). Category convention: `diagnosis`.
- **Q-G reviewer context**: per-reviewer docselect/contextkit bundles vs one
  shared bundle + role prompts vs curated RAG corpora per specialist.
- **Q-H the human-facing result**: what exactly lands (PR link + diagnosis +
  council report + task notes link) and where.

## LOOP-WORTHINESS TEST (doctrine — apply before every intake)
A task is loop material when ALL hold: (1) it is a SYMPTOM about system
behaviour, not a feature request — "why did/didn't X happen", not "build X";
(2) a causal mechanism plausibly exists in code + data + runtime; (3) it is
not answerable by one or two direct queries (run the cheap pre-check first —
the loop is for when the mechanism is genuinely unclear); (4) it is bounded
to one symptom (decompose lists; pick one); (5) the symptom is verified
CURRENT at intake — symptoms are perishable (the first pilot, missing chrome,
was fixed between selection and start). Feature absences go to build
routes; quality judgements go to the council/auditors; one-query questions
get the query.

## ★ F0 PILOT — CONFIRMED 2026-07-07: nav links to a guides section that has no content

**SYMPTOM (the intake string):** "dartsonline.com published a Guides nav link
and a /guides/index.html page, but the page is blank and no guide pages exist
— while gamesdesign.co.uk, on the same platform, has working guides (and
games and tools), and gaswholesalers.com has a working news feed."

**WHY THIS IS LOOP-WORTHY (all five criteria, unlike candidates 1 and 2):**
1. behaviour symptom, not a feature request ✓ (the system LINKED to something
   it did not build — something is broken, regardless of whether guides are
   desirable);
2. mechanism plausibly exists — and PROVABLY exists, because other sites on
   this platform have working guides ✓;
3. not one-query-answerable: the guides work item, the site_plan, the nav
   source, and the page_type routing table must be cross-read ✓;
4. bounded to one symptom ✓; 5. current ✓.

**IT CARRIES A DIFFERENTIAL — the strongest evidence shape.** Two sites, same
platform, opposite outcomes. The loop should be pointed at the difference, not
just the failure. (Likely differing variable: gamesdesign was probably built
via ADOPTION or pageflow-builder; dartsonline via the fresh-build relay.
Do not assume — let the loop establish it.)

**STANDING HYPOTHESIS (for the loop to confirm/refute FROM CODE, not from us):**
reconcile_site_plan's routing table (load_work_item_actions.go, read in §B3 of
the builder runbook) maps content|index|landing|blog-index|blog-post →
page-build-handler, with "tool" commented out. **"guide" is not in that list
at all**, while blog-index is. If the planner emitted guide pages (or a
guides section-index), reconcile may have silently dropped them while nav —
generated from the PLAN, not the built set — still published the link.
CORRECTION TO AN EARLIER READ: my "no guide-writing agent is wired in"
conclusion was drawn from FRESH-BUILD relay dumps only; gamesdesign proves a
working route exists somewhere. The dumps were not the whole platform.

**PRE-CHECK (mandatory, cheap; may sharpen but will not dissolve the symptom):**
```sql
-- what did the planner actually plan for guides, and what got built?
SELECT name, page_type, url, build_status, in_header
FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
  AND (name ILIKE '%guide%' OR url ILIKE '%guide%') ORDER BY name;

-- did any guide work item ever exist?
SELECT item_type, wi.status, handler_agent, attempts, spec->>'page_name' AS page_name
FROM site_work_items wi WHERE wi.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
  AND (spec->>'page_name' ILIKE '%guide%' OR summary ILIKE '%guide%') ORDER BY wi.status;

-- THE DIFFERENTIAL: how do gamesdesign's guide pages exist? (site_id via domain)
SELECT s.domain, p.name, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='gamesdesign.co.uk' AND (p.name ILIKE '%guide%' OR p.page_type='guide')
ORDER BY p.name LIMIT 10;
```

**PILOT CRITERIA (pre-registered):** (1) intake via the documented route;
(2) per-iteration bundles fetchable from diagnosis_artifacts; (3) per-iteration
notes written; (4) a CITED mechanism at static tier naming the code that drops
or fails to route guide pages (or refuting the standing hypothesis with
something better); (5) stretch: F1 emits a constrained edit plan on a branch.
ORDER: F0.1 plumbing lands first, then this runs through it as F0's acceptance
test. If the fix is "add guide to the routing table + ground nav in the built
set," it also closes builder-queue item 7's guides strand.

## SUPERSEDED: CANDIDATE 2 (fork) — blank guides-index, mechanism claim unverified
User: /guides/index.html on dartsonline also renders blank; no tools on the
site either; claim "we certainly have the mechanism to make guides" — NOT
YET VERIFIED (searched all 9 uploaded dumps incl. the full build-site-planner
workflow: page_type "guide"/"blog-index" exist as CLASSIFICATION vocabulary
in the adoption-analysis prompt, but no guide-WRITING agent, tools-suggestion
pipeline, or games mechanism was found WIRED into the relay dartsonline
travelled; blog-content-planner exists in the census but never appears as a
handler_agent in any dumped workflow — may belong to the news-feed pipeline
(006), not fresh-build planning, or be another orphaned front like
site-classifier).
FORK QUERIES (pre-registered; run before committing to this pilot):
```sql
SELECT item_type, wi.status, handler_agent, attempts, LEFT(summary,60),
       spec->>'page_name' AS page_name
FROM site_work_items wi
WHERE wi.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
  AND (spec->>'page_name' ILIKE '%guide%' OR item_type ILIKE '%guide%')
ORDER BY wi.status;

SELECT type, version, status FROM agent_definitions
WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (type ILIKE '%guide%' OR default_config::text ILIKE '%blog-content-planner%'
       OR default_config::text ILIKE '%tool-suggester%');
```
DECISION RULE: item EXISTS + status=complete + thin/empty result ⇒
GENUINELY DIAGNOSIS-SHAPED (mechanism fired, produced too little — this
becomes the F0 pilot, run it). Item MISSING entirely, or no agent anywhere
references guide-writing/tool-suggestion as a relay handler ⇒ ANOTHER
ABSENT-CONNECTION finding (same family as the roadmap gap) ⇒ builder queue,
not this thread; pilot slot stays open pending a third candidate or this
same fork on tools.

## PREVIOUS: F0 PILOT #1 — DOWNGRADED 2026-07-07 (root cause found; not diagnosis-shaped)
User correction, CONFIRMED IN CODE: the nav/undelivered-pages symptom is not
a mystery mechanism — 082_submit_domain_unified.sh has NO --roadmap path at
any entry point (mission/mission-file only), and build-site-planner's prompt
has no ELSE branch for the roadmap block — absent a roadmap, the phase rules
don't degrade, they vanish. This is a KNOWN, named platform gap (an absent
decision point), not a hidden causal mechanism → FAILS worthiness criterion 2
in its strong form (a citing loop would find what we already found by
reading two files). MOVED to RUNBOOK_builder_route.md item 6 (promoted to
MAIN) as a relay-wide fix: a new post-classification scope-decision hop +
three enforcement points, fixing every future site by construction.
REMAINING PILOT CANDIDATE for THIS thread (narrower, genuinely diagnosis-
shaped once the platform fix lands): after the new hop deploys, run ONE
diagnosis on a fresh test domain to CONFIRM the fix closes the loop
end-to-end (roadmap written → planner honours phase-1-only → nav reflects
only built pages) — a verification diagnosis, not a discovery one. Until
then this thread's F0.1 plumbing work proceeds unblocked; the PILOT SLOT is
open — propose a genuinely mechanism-unclear bug when one is in hand, or use
the verification run above as F0's acceptance test instead.
Imagery: handled in ANOTHER chat (boundary updated — off this thread).
**NEW PILOT: plan/delivery integrity — "the published navigation links to
pages that were never rendered"** (dartsonline: shop-index, product-detail,
brand-detail, brands-index, …). Two strands the trail must separate: (a) why
the needs_page items for those pages did not deliver; (b) why nav published
links regardless (nav generated from the PLAN, not from built reality).
MANDATORY PRE-CHECK before intake (may narrow the symptom to pure
nav-sequencing if items are merely queued):
```sql
SELECT item_type, wi.status, attempts, LEFT(summary,50)
FROM site_work_items wi
WHERE wi.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
  AND item_type='needs_page' AND wi.status <> 'complete'
ORDER BY wi.status;
```
Pilot criteria unchanged (intake route; fetchable bundles; per-iteration
notes; a CITED mechanism; stretch F1 edit plan). THEMATIC PAYOFF: the
diagnosis lands on the grounding principle below — its fix is the natural
first guideline-amendment + enforcement candidate.

## ROOT CONTEXT FOUND IN THE DOCS (2026-07-07 search, user-requested)
Guidelines 001 (~1503–1560) ALREADY defines the mechanism the owner asked
for: Tier-3 intake = mission + **roadmap with PHASES**; roadmap_brief names
the CURRENT phase; **"Future phases get a one-line summary each with a note
that they're not for building now."** 002 lists affiliate-link-manager as
phase 2. dartsonline was submitted WITHOUT a roadmap spec ⇒ the planner had
no phase constraint ⇒ aspirational full plan + nav to match. WHERE IDEAS
LIKE THIS LIVE (owner's question, answered): per-site staging → site_specs
`roadmap`/`roadmap_brief` (exists; write a 3-phase roadmap for dartsonline:
1 content/guides/tools; 2 affiliate-approved product introduction — legal
gate; 3 catalogue); platform principle → guideline amendment ("plans
grounded in deliverable capability; nav never links unbuilt pages;
product_category domains default content-first pending affiliate/legal");
enforcement points → strategist prompt (phased roadmaps for
product_category), planner validate (page-type deliverability vs live
handlers), nav grounded in built set (nav-updater exists). NOT the
thin-slice constitution (dev method, not product strategy). Dynamic assembly
(scrape+feeds+imagery composing brand/product pages) → builder-thread
phase-2 feature, compliance reviewer's first concrete job.

## F0 PILOT ORIGINAL RECORD (superseded) — dartsonline site chrome
The pilot bug (doubles as F0's acceptance test, per "test it on my next
task/bug"): **every deployed page on dartsonline.com lacks site chrome — no
header, no footer, no <nav> in any rendered HTML (index/about/contact/
new-arrivals measured) — despite populate_nav completing during planning.**
Site 5fe8785b-223d-41a3-88ee-c07187622381; pages built by page-build-handler,
deployed by rerender-pages; hypothesis on file (quality runbook): the relay
path may lack pageflow-builder's render_site_components step, or the
assembler doesn't inject chrome.
ORDER: land F0.1 first (artifacts table + assemble write-through + the
needs_diagnosis envelope), THEN run the pilot so it exercises the new
plumbing. Pre-registered pilot criteria: (1) intake via the documented
route; (2) per-iteration bundles fetchable from diagnosis_artifacts;
(3) per-iteration notes written; (4) diagnosis reaches a CITED mechanism
(which workflow/step/action omits chrome); (5) stretch: F1 produces a
constrained edit plan on a branch. Subject fields: align subject_type/
subject_key with the tools chat's taxonomy when their 3b lands.
BOUNDARY NOTE: the pilot runs ON quality-thread LEG 1 with the owner's
direction; findings and any fix feed RUNBOOK_site_quality.md; the quality
thread keeps the other legs. PRE-CHECKS handed to the owner for LEG 2/3
(may dissolve them without the loop): design/imagery item states + the
handlers' image_tag/is_active (the stale-tag class).
DISPOSITIONS: shop/catalogue = NOT diagnosis — a new build leg (product-data
mechanism beneath the already-planned shop/product/brands pages; the
vertical_landscape's catalogue-depth finding is its justification) — route
to the builder thread's queue. Thin content = council/auditors, later.

## COLLISION SURFACE + SHARED COMPONENTS (from the rev-22 read)
- The diagnose-agent workflow is THEIR active surface (emit→persist_note→
  complete; 3b in flight). RULE: any fix-loop change to diagnose workflows is
  fetch-first against the CURRENT JSON and coordinated; our egress lands
  Go-side in assemble (above) precisely to stay off that surface.
- SHARED, reuse-not-duplicate: their 084 trigger (canonical intake envelope);
  doc_notes (per-task notes home); the Stage-6 browser-runner adapter (a
  future verification service for F1 fixes touching pages, and a council
  reviewer's instrument); their criteria-fence pattern in doc_plans (a shape
  our fix proposals can carry acceptance criteria in).
- New gotcha ADOPTED (their 001 §16 finding): `error_step` belongs INSIDE a
  step's `config` — step-LEVEL error_step is silently ignored (dormant bug
  instances exist in tool agents); correct adjacent instances when touching a
  workflow, as its own noted change. Also: idle pods reap at ~3600s — the
  post-completion STATE DUMP (ProcessingHistory) is the accepted evidence
  substitute.

## CURRENT POSITION — 2026-07-07
Documents created 2026-07-06; tools-chat rev-22 ALIGNMENT DONE 2026-07-07
(Q-F direction set; Q-B loop-side enabler confirmed; Q-A refined to
assemble-side write-through; collision surface + shared components recorded).
DISCUSSION COMPLETE for F0/F1 (2026-07-07): Q-A/B/C/D/F all decided —
**CUTOVER-READY**. Q-E/G/H remain as F2-phase questions the new thread can
open with. The tools-chat relay is now a courtesy FYI only (we write nothing
into doc_notes; the envelope carries subjects for their gate). First slice:
F0.1 — the diagnosis_artifacts migration (kind column) + assemble
write-through + the needs_diagnosis envelope/namespace.
