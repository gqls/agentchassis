# RUNNING NOTES — Diagnosis→Fix Loop (v1)

Chronological; newest entries appended under DISCUSSION LOG; decisions
promoted to DECISIONS with rationale.

## 2026-07-06 — Workstream founded

- Origin: the diagnosis loop (read-only, three-tier citing, human-gated) is
  closed and working; the owner wants it developed into diagnosis→fix with:
  documented intake, live per-iteration/per-step reasoning into task-specific
  notes, fetchable bundles, fixes on a git branch, a council of specialist
  reviewers (guidelines / reuse incl. docs / bug-historian / compliance /
  per-pipeline guardians / named specialists e.g. trigger + site-work-items
  experts) feeding a decision-maker, architecture-change visibility, and a
  learning record of bugs. Motivating example: a chat re-invented a trigger +
  triage SQL that already existed — a specialist would have said so.
- Assets inventoried (see runbook): live loop, contextkit CLI (+ the real
  example_bundle.txt invocation), code_symbols corpus, the work-item relay,
  the tools chat's doc_notes infra (coordinate; their docs arrive next turn),
  the builder thread's pipeline map for guardians.
- Three documents created: RUNBOOK_diagnosis_fix_loop.md (task + phased plan
  F0–F3 + open questions Q-A…Q-H), this notes file, HANDOFF_fixloop_thread.md
  (manifest + the cmd/bundle invocation for code context).
- Method carried over: thin slices, pre-registered criteria, evidence first,
  snapshots, reuse before recreate; the loop's READ-ONLY core is preserved —
  the write surface (F1 fixer) is isolated by design.

## DISCUSSION LOG
(appended per exchange in the originating chat; handoff updated in step)

### 2026-07-07 — tools-chat rev-22 absorbed (their runbook + running notes)
- THEIR LIVE STATE: doc tables shipped; diagnose-agent workflow rewired
  emit→persist_note→complete with a skip-don't-guess subject gate (run-3
  verified the skip); load_runtime error-routing applied — ANCHORLESS RUNS
  SURVIVE (≈26 min / 5 iterations); 3b (subject threading) in flight; first
  tool PLAN live 2026-07-07 12:32; Stage 5 = static Tier-2 criteria check;
  Stage 6 = browser-runner adapter (Playwright, 035-conformant contract).
- WHAT IT ANSWERS HERE: Q-F → reuse doc_notes (per-iteration rows pending
  their volume sign-off; intake adopts their 084 envelope + subject fields).
  Q-B → the site-less-bug loop side is DONE by their routing; only the item
  namespace remains. Q-A refined: egress as write-through inside assemble
  (Go-side) — keeps us entirely off their emit-adjacent surface.
- COLLISION RULE ADOPTED: diagnose workflow changes are fetch-first +
  coordinated; their surface is active until 3b closes.
- GOTCHAS ADOPTED: error_step INSIDE config (step-level silently ignored —
  001 §16; fix adjacent instances when touching a workflow, noted); ~3600s
  pod reap → ProcessingHistory state-dump as evidence substitute.
- SHARED-COMPONENT REGISTER: 084 trigger; doc_notes; browser-runner (future
  F1 verification + a council instrument); criteria-fence pattern.
- RELAY QUESTION for the tools chat: per-iteration diagnosis notes in
  doc_notes — acceptable volume/shape? proposed: one note per iteration,
  category 'diagnosis', body = hypothesis/scope/requests/verdict grounds.

### 2026-07-07 (close of thread) — opening paragraph written; thread state at cutover

- Wrote the new chat's opening paragraph (single, self-contained): symptom +
  site_id + why it is a defect not a feature request ("the system linked to
  something it never built") + the gamesdesign/gaswholesalers comparison
  cases + the probable differing variable (build route) flagged as
  TO-BE-ESTABLISHED + the standing routing-table hypothesis stated AS a
  hypothesis to confirm-or-refute from code + the three pre-check queries +
  the work order (F0.1 plumbing, then the pilot through it) + success =
  static-tier citation, stretch = F1 edit plan on a branch.
- Caveat recorded with it: two earlier pilot candidates dissolved on
  inspection (chrome fixed before start; nav-to-unbuilt-pages had a root
  cause a code read found in minutes → builder item 6). The new thread must
  treat the reconcile_site_plan suspicion as a LEAD, not a conclusion. That
  triage history is the worthiness test earning its keep before the loop has
  run once.

## THREAD STATE AT CUTOVER (2026-07-07)

DECIDED AND RECORDED (nothing to re-derive):
- Q-A diagnosis_artifacts table (kind ∈ bundle|iteration_note), written
  through inside the assemble action — Go-side, off the tools chat's surface.
- Q-B intake = needs_diagnosis work item, NEW pipeline='diagnose' namespace,
  null-site allowed; envelope extends the tools chat's 084 trigger; manual
  trigger retained.
- Q-C fixer = separate agent; isolated git write token (spawn-gate pattern);
  constrained edit plan; gofmt+build in a spawned job before any PR.
- Q-D council = parallel reviewers → decision-maker by default; hard_veto
  flag at reviewer/pipeline/tool/component scope (accessibility, legal);
  guideline-gap = SIDE-TASK (amendment PR against the guideline docs; the
  fix is not blocked).
- Q-F per-task notes = our own working-notes storage; only the TERMINAL note
  lands in the tools chat's doc_notes via THEIR persist_note wiring; our
  envelope must carry subject_type/subject_key so their gate opens post-3b.
- ★ PILOT = the dartsonline guides-nav-without-guides defect (differential
  against gamesdesign; standing hypothesis on reconcile_site_plan's routing
  table + plan-sourced nav; three pre-check queries; five criteria).

STILL OPEN (F2-phase; the new thread may open its discussion here):
- Q-E which architecture-change signals are load-bearing (packages touched,
  platform/ vs actions/, exported-signature diffs vs the corpus, message/
  topic/schema changes, migration presence).
- Q-G reviewer context: per-reviewer contextkit bundles vs one shared bundle
  + role prompts vs curated RAG corpora per specialist.
- Q-H the human-facing result package (PR link + diagnosis + council report +
  task-notes link) and where it lands.
- Sub-questions from Q-D: where the hard_veto flag physically lives
  (reviewer definition column vs per-pipeline council config vs both,
  most-specific-scope wins).

ROUTED ELSEWHERE (do not work here):
- Builder queue item 6 — the roadmap/scope-decision gap: no submission path
  produces a roadmap (082 script has no --roadmap; the planner prompt has no
  else-branch), fixed relay-wide by a new post-classification hop + three
  enforcement points.
- Builder queue item 7 — coverage expectation (guides/tools/news/curated
  non-affiliate top-N) + the genuinely-new curated-list mechanism; plus the
  001_development_guide additions (roadmap mandate for commerce-shaped
  domains; "useful-but-unoriginal still counts as best-in-class").
- Tools chat — tool-pipeline internals, tool docs, tool-auditor, their
  diagnose_load_runtime draft. Imagery — another chat. Site quality legs —
  RUNBOOK_site_quality.md.

CARRY INTO THE NEW CHAT: the constitution; RUNBOOK_diagnosis_fix_loop.md;
this file; HANDOFF_fixloop_thread.md; BUNDLE_fixloop_F0.md (199,579B, 11/11
scopes resolved); RUNBOOK_code_retrieval_route.md; RUNBOOK_31_.md;
RUNBOOK_design_diagnosis_loop_7_.md; RUNBOOK_builder_route.md; the tools
chat's rev-22 pair; guidelines 000/001/003; example_bundle.txt.
FIRST ACTION: slice F0.1 with pre-registered criteria.

### 2026-07-07 (final) — ★ PILOT CONFIRMED: guides nav link with no guides content

- User: gamesdesign.co.uk HAS working guides/games/tools; gaswholesalers.com
  HAS a working news feed. My earlier "no guide mechanism wired in" was drawn
  from fresh-build relay dumps only — too strong, corrected in both runbooks.
- The pilot is therefore the STRONGEST candidate yet: the system published a
  Guides nav link and an empty /guides/index.html while an identical platform
  builds guides elsewhere. Passes all five worthiness criteria AND carries a
  DIFFERENTIAL (two sites, same platform, opposite outcomes) — the strongest
  evidence shape for a citing loop.
- Standing hypothesis for the loop to confirm/refute from code (NOT asserted):
  reconcile_site_plan's routing table routes content|index|landing|blog-index|
  blog-post; "guide" is absent, "tool" is commented out; nav is generated from
  the PLAN rather than the built set — so a link can outlive a dropped page.
- Pre-check queries (pages, work items, gamesdesign differential) + the five
  pilot criteria recorded in the fix-loop runbook. Order unchanged: F0.1
  plumbing, then the pilot as F0's acceptance test.
- Builder item 7 narrowed to the coverage/standing-expectation strand + the
  genuinely-new curated top-N list mechanism.

### 2026-07-07 (later still) — TWO CORRECTIONS: amendment path under-specified; bug is platform-wide

- Owner correction 1: "the amendment path" was a future F3 mechanism, not
  usable today. CLARIFIED: 001_development_guide is the right doc (Tier-3
  roadmap section, ~1503-1560) — owner will supply it for a direct edit
  once ready.
- Owner correction 2 (the important one): fixing dartsonline's spec fixes
  one site; the BUG is that no submission path ever produces a roadmap.
  VERIFIED IN CODE: 082_submit_domain_unified.sh — grep confirms ONLY
  --mission/--mission-file exist, no --roadmap anywhere. build-site-planner
  prompt — the {{if .roadmap_brief}}...{{end}} block has NO else; absence
  means the phase-authority instructions vanish, not degrade. This is an
  ABSENT DECISION POINT, confirmed, not a hidden mechanism.
- RECLASSIFIED: this is the builder thread's MAIN queue item now (item 6,
  promoted from background), NOT a diagnosis pilot — fails worthiness
  criterion 2 in its strong form. Fix-loop's role narrows to a future
  VERIFICATION run once the platform fix ships. Pilot slot reopened.
- Fix shape sketched for the builder thread: new post-classification hop
  writing a phased roadmap for commerce-shaped domains; three relay-wide
  enforcement points (strategist/new-step; planner else-branch; built-
  grounded nav) — fixes every site by construction, matching the owner's
  "generalise the enforcement" instruction exactly.
- Enforcement generalisation confirmed as directly achievable: strategist
  prompt / planner validate / nav-updater are ALL relay-level steps already
  (not per-site) — editing them once IS the generalisation.

### 2026-07-07 (latest) — pilot replaced; roadmap mechanism found; homes assigned
- Chrome pilot EVAPORATED (fixed live) → worthiness criterion 5 added
  (verify symptom currency at intake). Imagery → another chat (boundary).
- NEW PILOT: nav links to never-rendered pages (two strands: undelivered
  items; plan-sourced nav). Mandatory pre-check query recorded — queued-slow
  items narrow it to nav-sequencing, still the pilot.
- Docs search ANSWERED the owner's question: 001 Tier-3 roadmap PHASES exist
  ("future phases… not for building now"); 002 has affiliate phase-2.
  dartsonline lacked a roadmap spec ⇒ unconstrained aspirational plan.
  Homes: per-site staging → roadmap/roadmap_brief specs (3-phase dartsonline
  sketch on offer); platform principle → guideline amendment; enforcement →
  strategist prompt + planner deliverability validate + built-grounded nav.
  Constitution ruled out. Dynamic assembly + legal gate → builder queue P2.

### 2026-07-07 (later) — pilot chosen; worthiness doctrine; bundle confirmed
- BUNDLE_fixloop_F0.md generated by the owner and verified clean: 199,579B,
  all 11 scopes resolved (538 files analysed) — ready for the new chat.
- Owner asked whether the loop fits the dartsonline quality problem. Answer:
  decomposed via the new LOOP-WORTHINESS TEST (symptom-not-feature;
  mechanism-plausible; not one-query-answerable; single-symptom). PILOT =
  the chrome/nav defect (mechanism-shaped, every-page value, small likely
  fix = F1's first edit plan). LEG 2/3 get cheap pre-checks first
  (item states + handler image_tag/is_active). Shop/catalogue = build
  territory, not diagnosis (product-data mechanism beneath already-planned
  pages; vertical_landscape justifies it) → builder-thread queue. Thin
  content → council, later. Quality runbook confirmed unchanged since
  handoff; boundary line recorded (pilot runs ON their LEG 1, findings feed
  back).

### 2026-07-07 — Q-D veto semantics decided (owner)
- Flag-based: DEFAULT = decision-maker weighs all opinions; a hard_veto flag
  at reviewer/pipeline/tool/component scope makes that reviewer's negative
  verdict a BLOCK (accessibility, legal = motivating cases). Open detail for
  the new thread: flag placement (definition column vs council config vs
  both; most-specific wins?) and the guideline-gap case (block vs side-task —
  leaning side-task).
- Per-iteration note volume issue explained to the owner + relay message
  DRAFTED for the tools chat (three shape options: filterable category /
  is_current-versioned single note / own table + terminal note only).
- Standing unconfirmed at the time: Q-A/Q-B/Q-C — ALL CONFIRMED/DECIDED by
  the owner on 2026-07-07 (see DECISIONS): diagnosis_artifacts; (c) own
  working-notes table + terminal-only in doc_notes; pipeline='diagnose';
  separate fixer agent w/ constrained edit plan + gofmt/build gate.
  Guideline-gap = side-task, with the mechanism answered (amendment PR
  against the guideline docs; human terminal; F3 recurrence record).
  REMAINING OPEN (F2-phase; may open the new thread's discussion): Q-E
  architecture-change signals, Q-G reviewer context, Q-H result format.

## DECISIONS (with rationale)

### 2026-07-07 — F0/F1 design settled (owner decisions)
- **Q-A — bundle egress = `diagnosis_artifacts` table**, written through from
  INSIDE the assemble action (Go-side; zero workflow-shape change; off the
  tools chat's emit-adjacent surface). Rationale: DB is the durable shared
  memory; one-SQL fetch by correlation; cd/Kafka size class already bitten.
- **Q-F shape = (c)**: working notes in OUR OWN table; only the terminal note
  lands in the tools chat's doc_notes — via THEIR existing persist_note
  wiring (we write nothing into their table). Integration duty: the intake
  envelope carries subject_type/subject_key so their gate opens post-3b.
  REFINEMENT (proposed, start unified): fold working notes into
  diagnosis_artifacts with kind ∈ {bundle, iteration_note}; split later only
  if retention diverges. Relay to tools chat downgrades to a courtesy FYI.
- **Q-B — intake = `needs_diagnosis` work item, `pipeline='diagnose'`**
  namespace (site-less code bugs ride null-site in the new namespace; the
  loop side already survives anchorless via their load_runtime routing);
  envelope extends their 084 trigger; manual trigger retained for ad-hoc.
- **Q-C — the fixer = a SEPARATE agent**: distinct responsibility; the git
  WRITE token isolated behind the existing spawn-gate pattern; diff produced
  as a CONSTRAINED EDIT PLAN and validated by gofmt+build in a spawned job
  before any PR opens.
- **Q-D completion — guideline-gap = SIDE-TASK** (does not block the fix):
  a work item carrying the evidence; handler drafts a concrete amendment and
  opens a PR against the GUIDELINE DOCS (docs-branch symmetry with F1);
  human review terminal; gaps accumulate in the F3 learning record
  (recurrence = restructuring signal). Decision-maker sees "gap raised" as
  context only.
