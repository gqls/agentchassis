# RUNNING NOTES — Experience Loop

*Append-only. Newest entry last. Companion to PLAN_experience_loop.md and
RUNBOOK_experience_loop.md. House rule: every entry states what was done, what
was verified (and how), and what it changes for the next actor.*

---

## 2026-07-17 — Workstream ACTIVE: defaults accepted, RUNBOOK drafted (session "vonc4")

**Owner input**: "defaults accepted" — resolving PLAN §7 / HANDOFF §3d:

1. **Gauntlet in MVP = minimal-real** — a playable timed round against the daily
   provocation, client-side scoring, no leaderboard; every fabricated number
   stripped. Demotion to coming-soon only if the feasibility critic proves
   minimal-real can't be honest and small.
2. **Provocation detail pages = static daily-emitted pages** — consistent with
   the daily JSON emitter design. NOTE (verified this session): the daily
   emitter itself is NOT BUILT — `/data/provocations.json` is a hand-committed
   static sample; the pipeline exists only as
   `docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`.
   The council must scope this dependency in the MVP cut.
3. **Pilot fully autonomous on vonc** — artifact-verified checkpoints recorded
   here (CP1–CP5, defined in RUNBOOK §2), no approval gates.

**Done this session**: machinery survey (4 read-only exploration passes over the
acceptance ladder, council pattern, build/render side, claims verification) +
live-DB schema checks; RUNBOOK_experience_loop.md drafted from the findings;
PLAN header flipped PROPOSED → ACTIVE.

**Load-bearing facts established (each verified against code or live DB, refs in RUNBOOK §1):**

- `doc_plans`/`doc_notes` CHECK constraints allow only `('tool','pipeline')` —
  `subject_type='experience'` requires a migration on BOTH tables (RUNBOOK T1.1).
- browser-runner has NO navigation step action; each URL runs in a fresh browser.
  Tier-4 journeys are a genuine extension, not configuration (RUNBOOK T5.1).
- The arena's orphaning is a THREE-way key mismatch: criteria load by
  `content_components.function`, URL resolution by `pages.name = function`,
  travelling docs by `subject_key = function`. The page rename to `tool-arena`
  broke the middle link; re-keying doc_plans alone would NOT reconnect the sweep
  (RUNBOOK T4.2).
- `ReconcileSitePlanAction` emits `needs_page:<name>` → page-build-handler for
  EVERY plan page including tool-owned ones — the tool-role exclusion that
  `check_incomplete_page_group.go` already encodes (TP-004) is missing there.
  That asymmetry is guard rail 1's primary edit (RUNBOOK T2.1).
- No page-ownership marker column exists anywhere; TL-001 protection today is
  heuristic guards in `SavePageSectionsAction` + a manual park step.
- No code re-keys travelling docs on rename; no automated renamer exists either
  (the arena rename was manual). Root cause of the drift class:
  `create_tool_component_action.go:207` hardcodes `/tools/{function}.html`
  instead of calling `CanonicalisePage` (RUNBOOK T2.2).
- Claims verification V1a/V1b code is committed but NOT in any deployed image;
  vonc has NO evidence_base row, so the gauntlet's fabricated numbers are
  currently outside every enforcement lane (RUNBOOK T2.4).
- Migration sequence (`docs/agent_docs/sql_for_agents/`) runs through 161 with
  known collisions at 151 and 157; next free number at this snapshot = 162 —
  RE-CHECK at execution time.
- Concurrent-session state at drafting time: uncommitted WIP by another session
  in `platform/orchestration/actions/create_rerender_items_action.go` (a
  guard-rail-1 touchpoint) and an untracked live-probe test in
  `internal/adapters/browserrunner/` — re-check both before editing.

**Next actor starts at**: RUNBOOK §3 phase table — Phase 1 (foundations
migration) then Phase 2 (guard rails).

---

## 2026-07-17 (later, same session) — Phase 1 DONE, Phase 2 guard rails 1–4 DONE (code), image roll A next

Owner said "please go ahead"; execution began. Everything below is committed
per task and verified as stated; RUNBOOK §8a mirrors this as a state table.

**Phase 1 — DONE.** Migration **163** (not 162 — that number was taken
mid-flight by another session's toolgen-rerender tail, itself good news: it
closes TP-002). Both subject_type CHECKs now allow `experience`; applied +
ledgered; probe insert verified. Commit `378054bad`.

**Guard rail 1 — DONE (DB live, Go awaits image).** Migration **164**:
`pages.rebuild_policy generic|owned`, 38 pages seeded owned (36 `page_type=
'tool'` fleet-wide + vonc `provocations-index`/`provocation`). Go: reconcile
now emits `owned_page_review` (needs_human_review, NO handler) instead of
`needs_page`→page-build-handler for tool/game-role or owned pages — retires
the manual park step; `save_page_sections` hard-refuses owned pages before
its heuristic guards. Rerender/assembly deliberately not gated. Commit
`fb89f1071`.

**Guard rail 2 — DONE (code).** `RekeyTravellingDocs` datahelper (refuses
two-current collision), `rename_tool_identity` action (atomic function +
slot_name + doc re-key + rename note; reports pages.name coupling and stale
js_snippets/nav refs), and `create_tool_component` now derives page identity
via `CanonicalisePage` (kills TL-003 flat-URL drift at birth;
sanitiseFunction's tool- prefix keeps pages.name == function). Commit
`aabd38161`. The arena re-attach (T4.2) is this action's first deliberate use.

**Guard rail 3 — DONE (code).** `IsNoopHref`/`DeadControlAnchors` (+unit
tests, green): bare `#`, `#!`, javascript:void — the class ClassifyLinkScope
correctly files under anchor scope, which is why the gauntlet's dead CTAs
were invisible to phantom_internal_links. Wired as (a) Tier-2 built-in
`shell-dead-controls` and (b) `dead_controls` discovery check
(page_components only — chrome nav toggles are legit #+JS; runtime-fill
shells exempt; emits `dead_control`, needs_human_review, no handler).
Enable SQL **165 written but NOT applied** — image-first ordering. Commit
`f2824a713`.

**Guard rail 4 — DONE (live).** Migration **166**: vonc `evidence_base`
seeded — facts deliberately EMPTY (nothing quantitative is assertable until
EXPERIENCE_PLAN data contracts exist), 9 banned patterns from a live-page
harvest, allowed_entities for the fictional product nouns. **claimscan
baseline: 14 findings across 49 components** — the full known gauntlet set
PLUS three previously unknown fabrications: `14,203 Happy Customers`
(gauntlet-cta component on about AND index), `10K+ Players Scored` (index),
and mangled stat labels on about (`1 Daily Clients Served`, `4 Hours Awards
Won`). All queued for the T4 strip. Also found: `unverified_claims` was
already pre-enabled on quality-discovery-agent by its own workstream —
V1b activates the moment the claims image ships, no further seed needed.
Commit `c437682a6`.

**Concurrency observed this session** (all handled): migration 162 claimed
between drafting and execution; the build default inverted fleet-wide
(default target now builds committed HEAD — RUNBOOK §0 updated); cluster
rolled 1128→1130 by other sessions; `create_rerender_items_action.go` WIP
belongs to the toolgen-tail session (file untouched by us, as planned).

**Next**: T2.5 — image roll A + CP1.

---

## 2026-07-17 (later) — ✅ CP1 REACHED: guard rails live in prod and proven on vonc

**Image roll A came free.** Another session built+deployed **v1.0.1134** from a
HEAD that already included all four guard-rail commits — so no separate roll
was needed (my local 1132 build is moot; I did not push it). Verified in-pod,
NOT by tag: `strings /app/agent-chassis` in pod `agent-chassis-6d85fff446-54jzc`
(container name is `agent-chassis`, NOT `agent` — that's only the intel pair)
shows rebuild_policy×4, owned_page_review×4, rename_tool_identity×4,
RekeyTravellingDocs×2, shell-dead-controls×1, dead_controls×6, CanonicalisePage×2.
Deployment AND agent_definitions.image_tag both v1.0.1134.

**Migration 165 applied** (image-first satisfied — dead_controls symbol confirmed
in-pod): `dead_controls` now in completeness-discovery-agent's checks array.

**CP1 proofs (all live on vonc, artifact-verified):**

1. *Binary*: the 7 symbols above are in the running 1134 binary.
2. *dead_controls check fired* — completeness discovery (corr
   4cedb4fb, completeness-discovery-agent row COMPLETED) emitted two live
   `dead_control` items: index `brief-explanation` "Get Started" → `#` and
   "Learn More" → `#`, both needs_human_review, no handler. Genuine new finds
   (the index had dead CTAs too, not just the gauntlet). The gauntlet's own
   dead CTAs weren't caught here only because tool-gauntlet is
   build_status=needs_rebuild, not deployed — it's covered by the claims lane
   and rebuilt in T4.
3. *owned_page_review routing* — a SCOPED reconcile (corr 4c0c4acf; ran ONLY
   reconcile_site_plan via a one-step envelope, NOT build-site-planner, to
   avoid the re-plan clobber) emitted `owned_page_review` (needs_human_review,
   NO handler) for `provocation`, `tool-gauntlet`, `tool-archetype-taster-quiz`
   — exactly the tool/owned pages that previously went to needs_page →
   page-build-handler. `tool-arena` (deployed at current plan version) correctly
   skipped. **Zero needs_page emitted for any owned page.** The manual park of
   needs_page:provocation is now mechanical; the stale 07-12 park item
   (01674b35) was cancelled as superseded.
4. *save_page_sections refusal*: code-verified in the 1134 binary (the
   rebuild_policy='owned' guard returns before the DELETE+reinsert); not fired
   against a live owned page — the reconcile proof already demonstrates the
   marker is read and enforced, and firing a destructive save purely to see it
   refuse is not worth the risk.

**Guard-rail bonus already banked**: between the claimscan baseline (14 findings,
3 previously unknown) and the dead_controls sweep (2 more on index), the rails
surfaced 5 defects nobody had catalogued — before any experience work began.

**Phase 2 CLOSED.** Next: Phase 3 — experience-planner agent + challenge council,
run to convergence on vonc-spark-game (CP2).

---

## 2026-07-17 (later) — Phase 3 BUILT & STAGED; blocked on image roll B for the council run

Everything for Phase 3 is written, committed, and validated; the ONE thing left
before the council can run for CP2 is a chassis image carrying the
docResolveSubject fix.

**Split contract found & closed.** Migration 163 opened doc_plans/doc_notes to
`subject_type='experience'`, but `docResolveSubject` (shared by write_doc_plan +
append_doc_note) still rejected anything but tool/pipeline in Go — the
experience-planner's persist_plan step would have failed on it. Fixed
(`write_doc_plan_action.go`, commit **66d32477d**) with a lockstep comment and a
positive test. **This fix is NOT in the deployed image** — v1.0.1134 was built
before it (verified in-pod: `strings | grep -c "pipeline' or 'experience'"` = 0).

**experience-planner + council seed** — `167_experience_planner_and_council.sql`
(commit 0059521f9). ONE workflow, modelled verbatim on fix-proposer v6:
load_context → load_schema_hint → compose (sonnet-5, 16k) → persist_plan
(write_doc_plan, subject_type=experience) → 4 critics
(journeys[veto]/feasibility[veto]/honesty[HARD veto]/mvp[advisory], sonnet-5 4k
each) → council_decide (diagnose_council_decide reused verbatim, hard_veto_from
=['honesty'], max_rounds 3) → append_council_note (doc_notes, experience-council)
→ router (approved→complete / veto→reframe-once→escalate / object→run_checks→
recompose). No root ai_service (MDL-039). Syntax validated by a rolled-back
apply (`INSERT 0 1`). **NOT yet applied** — image-first sequencing.

**Two runner traps handled in the seed** (both verified against ai_actions.go):
(1) execute_llm_prompt ALWAYS attempts JSON-parse and falls back to text — so
critic JSON parses to a map (council reads `.result`), and the prose plan lands
as `.result` text; (2) stripMarkdownFromResponse strips a fence only if the whole
response starts/ends with ``` — the plan ends with a `<!-- END EXPERIENCE_PLAN
-->` trailer AFTER the criteria fence so the embedded fence survives storage for
later extractCriteriaBlock. Also: input_data added to compose/recompose/reframe
input_fields so `{{.experience_name}}`/`{{.experience_domain}}` resolve (the
extractor promotes input_data keys to root).

**Trigger** — `092_TRIGGER_experience_plan.sh` (090-style: durable
`needs_experience_plan` intake item + kcat orchestrate on a shared correlation;
passes experience_key/name/domain/correlation via input_data).

**BLOCKED ON**: a chassis image built from a commit ≥ 66d32477d, deployed to
agent-chassis (+ business-intel/vet-intel + agent_definitions.image_tag). Image
deployment is owner/pipeline-driven here (my manual push was interrupted earlier;
1134 was rolled externally). The moment such an image is live:
  1. `strings /app/agent-chassis | grep -c "pipeline' or 'experience'"` ≥ 1 (verify in-pod);
  2. apply 167 (+ ledger row) — image→seed ordering;
  3. wait ≥300s past any pod restart, then
     `bash docs/agent_docs/sql_for_agents/092_TRIGGER_experience_plan.sh vonc.com vonc-spark-game "the Spark daily-provocation game"`;
  4. judge by the experience-planner orchestration row; expect the council to
     sharpen D2's emitter scoping and D1's minimal-real Gauntlet. On escalation,
     the disagreement IS the round-boundary decision menu — surface it. → CP2.

---

## 2026-07-18 — CP2: the council is LIVE and WORKING; it escalated, and its
## reason is an owner decision (D2 is not cheaply feasible)

v1.0.1135 shipped the docResolveSubject fix (verified in-pod, grep=1), seed 167
applied + ledgered, planner fired. **The loop works end to end**: compose →
persist (doc_plans subject_type=experience) → 4 critics → deterministic
council_decide → router → run_checks → recompose, looping and superseding the
plan each round. Six plan versions written; the final is **18,413B, is_current,
with an intact ```criteria fence and END trailer** (the stripMarkdown trailer
defence held).

**Four runner defects found by running it** (all fixed, commit 6c5dc9e13):
1. `ExtractFields`→`UnwrapDeep` strips the `{result,type}` wrapper, so in
   TEMPLATE context an LLM step's output IS the value: `{{.proposal.result}}`
   → `{{.proposal}}`. Config paths (`plan_body_field`, `review_fields`) read RAW
   collected_data and correctly keep `.result` — that asymmetry is the trap.
2. Critic `max_tokens` 4000 → 8000: feasibility's JSON truncated mid-object and
   council_decide **failed closed** ("likely truncated at max_tokens") rather
   than waving a partial review through. The fail-closed behaviour is correct.
3. `councilReview.Edit` is an `int`; a critic emitted a string → critics now
   told `edit` MUST be a bare integer section number. Plus compactness caps.
4. Plan-contract gaps the council itself proved: §4 must be an ORDERED, GATED
   step list with prerequisite DATA steps as step 0, and §3 must define the
   exact computation + honest label of any number a visitor reads as a score.

**Two runs to a terminal, both `complete_escalated`** (max_rounds 3, then 5):
- Run A (3 rounds): revise×3 → "objection from feasibility — revise cap reached".
- Run B (5 rounds, tightened contract): revise×5 → "objection from journeys".
Final round B: honesty **approve**; journeys object (med/low/low); mvp object
(low/low/med); feasibility object (**HIGH**/med/med/med).

**Why it does not converge — two distinct causes:**
- *Structural*: `decideCouncil` maps ANY object → revise, with no severity
  threshold, so one low nit from any of four critics blocks approval and burns
  a round. `diagnose_council_decide` is SHARED Go (fix-proposer, council-gate)
  — changing its semantics unilaterally is out of bounds. Fixed instead at the
  prompt layer, inside my own agent: **verdict discipline** — object only for
  medium/high; low nits go in `notes` with verdict approve. (Applied; not yet
  re-run.)
- *Substantive, and the real one*: **feasibility's HIGH objection says D2's
  default is not cheaply feasible** — "page_type='provocation' has zero prior
  rows and no proven build/render pipeline; the plan folds authoring a whole
  new page-type render path into an MVP step without sequencing it." The
  planner cannot resolve this because D2 is stated in its prompt as an accepted
  owner default it must not relitigate. **This is exactly the round-boundary
  decision menu D3 anticipated: it needs the owner, not another round.**

**DECISION MENU (owner) — per-provocation detail pages:**
- **A. Keep D2's default** (static per-provocation pages) but sequence the new
  page-type render path as its OWN prerequisite round before the MVP. Honest,
  slower, and adds an unproven render path to the critical path.
- **B. (recommended) Switch to PLAN §7's documented alternative** — client-side
  detail rendering on the existing archive page. The archive is ALREADY a
  runtime-fill shell hydrating from the same feed, so there is no new page type,
  no new render path, and the MVP ships on proven machinery; static per-
  provocation pages move to LATER once the daily emitter exists.
The Gauntlet (D1) is NOT the blocker — honesty approved the final plan, and
feasibility confirmed the timer/scoring/reset are genuinely client-side-doable.

**Do NOT build the current is_current plan** — it is the escalated (rejected)
version; that is the standing rule for an escalated experience plan.

**Next**: owner picks A or B → set it in the compose prompt's D2 block → re-fire
(verdict discipline is already in) → expect convergence → CP2 closed → T4.

---

## 2026-07-18 — this council's escalation VALIDATED against bugs_open/016

Bug 016 (found here while building this council, now circulated and largely
fixed by the owning threads) raised a question about MY OWN result: if a
reviser cannot see the objections, a council can look stubborn-but-working
while actually being broken. The feature-builder thread confirmed the
pathology is real on their run `3b084712` — three rounds burned with the
bug-historian's objection UNCHANGED in every one — and named the tell:
**facts improve while objections never get addressed**.

**This council does not show it.** Per-round verdicts from run `6a4710d2`:

| round | journeys | feasibility | honesty | mvp |
|---|---|---|---|---|
| 1 | object(4) | object(5) | approve | approve(3) |
| 2 | object(3) | object(5) | **object(1)** | object(4) |
| 3 | object(3) | object(4) | approve | approve(2) |
| 4 | object(4) | object(3) | approve | approve(2) |
| 5 | object(3) | object(4) | approve | object(3) |

Verdicts flip and objection counts move every round (honesty
approve→object→approve; mvp approve→object→approve→approve→object;
feasibility 5→5→4→3→4). That is a reviser demonstrably reacting to what it
was told — the opposite of 016's tell. Two reasons this council was clean:
its template fix landed BEFORE any run that reached a verdict (the first run,
`cca7ea8c`, died loudly at `{{.proposal.result}}` on a TEXT step rather than
degrading silently), and its `check_results` reference was always
`.results_text` — a field ON the unwrapped value, which is correct.

**So the escalation is trustworthy, and its diagnosis stands**: this is
OSCILLATION across a four-critic panel (each round fixes one critic and trips
another), not a blind loop. That is what the verdict-discipline change targets
(object only for medium/high) — applied, but NOT yet exercised.

**It does not, however, remove the blocker.** Feasibility's objection is
HIGH severity and structural (a page_type with zero prior rows and no proven
render path), so it survives any severity threshold. Re-running without an
answer to D2 would burn another ~25 minutes and escalate again. The owner
decision is the gate, not the round cap.
