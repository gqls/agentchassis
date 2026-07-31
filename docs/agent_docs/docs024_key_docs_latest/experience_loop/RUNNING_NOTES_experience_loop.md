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

---

## 2026-07-18 — D2 answered (B); three more MY-side defects found by running it

**Owner ruling: D2 = option B.** Per-provocation detail renders CLIENT-SIDE on
the existing `/provocations/index.html` runtime-fill shell. No new page_type,
no new render path, no per-provocation static pages in the MVP; those and the
daily emitter move to LATER. Written into the compose prompt WITH its rationale
so the planner cannot re-derive it (commit `5f96a7330`), paired with a hard
constraint so B cannot reintroduce the original defect in a new coat: opening
an archive entry must be a REAL observable outcome (deep-linkable
fragment/param + a detail region populated with that entry's real feed
content); a class toggle or an empty region is the same dead control, and an
entry with no detail content must not be presented as openable.

Four council runs followed. Each escalated, and each escalation was correct —
every one exposed a defect **of mine**, not of the plan:

**1. Verdict discipline worked.** Objection counts per critic fell from 3–5 to
1–2, low-severity noise largely disappeared, honesty approved in every round of
every run. Two runs got within ONE medium objection of approval.

**2. `load_context` was lying by omission** (fixed, `7fa27c401`). It filtered
`cc.component_level='tool'`, but `gauntlet-interface`, `gauntlet-cta`,
`provocation-card` and `lobby-grid` are all `level='section'`. The context
surfaced ONE component when five are active and attached — and `tool-gauntlet`
genuinely DOES have `gauntlet-interface` attached. So the planner asserted the
gauntlet component existed while the critic could see no evidence, and objected.
**Correctly. Five rounds running, across two independent runs.** Now surfaces
per-page attachments with no level filter, labelled COMPLETE ground truth, with
both planner and critics told not to claim what is absent nor object as
unverifiable what it settles.

**3. A compose TRUNCATION DEATH SPIRAL** (fixed, `a751397f3`). `llm_call_log` is
unambiguous — recompose output per round: 13303 → 12599 → 14138 → 15499 →
**16000/16000**. The plan grew each round absorbing objections until it hit the
ceiling and truncated mid-JSON *inside the §5 criteria fence*; journeys and
feasibility then both objected THAT IT WAS TRUNCATED — an objection revising can
never clear, because revising makes it longer. Same class as the article-body
fix (2000→8000), new place. Fixed at both ends, because raising the ceiling
alone only delays it: max_tokens 16000→32000 on compose/recompose/reframe, PLUS
a LENGTH DISCIPLINE rule (revise by tightening and replacing, never appending;
§5 criteria fence has absolute priority and must always be complete, closed and
followed by the END trailer). **The council caught its own document being
truncated** — without it, a plan with an unparseable criteria fence would have
gone downstream to the acceptance ladder.

**4. Operator error, mine: a double-fire.** The first D2=B trigger sat in the
topic ~10 minutes before being consumed, my poll window missed it, I re-fired,
and two councils ran concurrently on the same experience key. Harmless here
(each round-counts by its own orchestration_id; both escalated identically, and
the redundancy was actually informative — two independent runs agreeing) but it
is exactly what the CLAUDE.md coverage rule exists to prevent. **Check for an
in-flight run before every fire; the topic lag can exceed 10 minutes.** Now done
as a matter of course (`SELECT count(*) ... status NOT IN ('COMPLETED','FAILED')`).

**State at handoff**: run `054b358a` fired 15:31 with all three fixes live, and
is QUEUED — the chassis has consumed nothing fleet-wide since 15:27 (single
replica, Running 6h19m, 0 restarts, zero AWAITING_RESPONSES backlog, so nothing
is holding a slot). Deliberately NOT restarting the chassis: it would disrupt
every concurrent thread and the evidence does not support it — a prior run
landed fine after a similar ~10 minute lag. Expect it to land and run ~25 min.

**Judge that run by**: whether the persisted plan ends with
`<!-- END EXPERIENCE_PLAN -->` and a closed ```criteria fence (proves the
truncation fix), and whether feasibility stops objecting about component
existence (proves the context fix). If it converges → CP2 closed → T4.

---

## 2026-07-18 late — run `054b358a`: BOTH fixes confirmed; killed by a known bug

The run landed and both fixes are **proven**:

- **Truncation fixed.** Persisted plan: 13,578 B, closed ```criteria fence,
  `<!-- END EXPERIENCE_PLAN -->` present. (Previous run: 26,522 B, no trailer,
  cut off mid-selector.) The LENGTH DISCIPLINE rule also made it *terser*, not
  just longer-ceilinged — which was the point.
- **Context gap fixed.** Feasibility's objections changed class entirely: no
  more "the only tool component surfaced is tool-arena-interface / cannot
  verify the gauntlet component exists". Round 1 objections are now substantive
  — who authors the `/data/provocations.json` fields, and that
  `header-bold-gradient` / `footer-4-column` / `Document Head` are deactivated
  site-wide across 16 pages.

**It still did not converge, for a reason outside this loop.**
`review_feasibility` failed with:

```
AI call failed with unhandled error: no text content in response (had 1 blocks)
```

The step's `error_step` routed to `complete_refused`, so the whole run
terminated after round 1 and four critics' work was discarded.

This is **`bugs_open/008` item 5**, verbatim: "handle `stop_reason == 'refusal'`
explicitly (Sonnet 5+ returns it; currently it would surface as *no text content
in response*)" — filed as **optional** and left undone. The only other
occurrence in 7 days is 008's own originating case (`diagnose-agent`/`verdict`,
07-16). Real-case evidence appended to 008; the fix belongs to the fixloop
thread.

**Resilience lesson for this loop, worth doing regardless of 008**: a single
flaky critic should not destroy a whole council run. `diagnose_council_decide`
already treats an absent reviewer as an **abstention** (it only fails closed if
ALL reviewers are missing), so pointing the four critic steps' `error_step` at
`council_decide` instead of `complete_refused` would let a run survive one dead
critic. Two-line config change; recommended as the first action next session.

> **CORRECTED 2026-07-19:** the diagnosis above is right, the prescription was
> wrong in two ways — see the 2026-07-19 entry below. (1) Routing a critic to
> `council_decide` skips every critic *after* it, turning one dead seat into
> three abstentions; the fall-through must be to the NEXT critic. (2) It must
> NOT be applied to all four: `review_honesty` is the sole `hard_veto_from`
> seat, so letting it abstain would let a plan reach "approved" with the
> anti-fabrication gate never applied. Caught by reading `council_decide`'s
> config (`hard_veto_from: ["honesty"]`) before writing the change.

**Also surfaced, not ours**: vonc's header/footer/head site components are
deactivated across 16 pages including `provocations-index` and the tool pages.
Verify before building on those pages (T4 touches them).

**Handoff written**: `HANDOFF_2026-07-19_experience_loop_resume.md` — start
there.

---

## 2026-07-19 (session "experience loop 2") — abstention-tolerant critics; CP2 re-attempt

Picked up from the resume handoff. First action was the recommended resilience
fix, but reading the code before writing it changed the shape of the change.

**What the previous session's recommendation got wrong.** Two things, both found
by reading rather than by failure:

1. **Fall-through target.** Critics run in sequence
   `review_journeys → review_feasibility → review_honesty → review_mvp →
   council_decide`. Pointing a failed critic at `council_decide` skips the
   remaining critics as well — one dead seat would have become three
   abstentions. Each critic now falls through to the **next critic**.
2. **Not all four.** `council_decide`'s config is
   `hard_veto_from: ["honesty"]`. Since an absent field reads as an abstention,
   making `review_honesty` fall through would let a plan reach `approved` with
   the anti-fabrication gate never applied — the exact class the loop exists to
   catch. `review_honesty` **keeps** `error_step: complete_refused`; a dead
   honesty auditor must refuse the run. The asymmetry is commented in the seed
   so it does not get "made consistent" later.

**Verified before changing** (`diagnose_council_decide_action.go`): absent review
field → `abstained++`, skipped (`:98-112`); all-absent → hard error, fails closed
(`:141`). So this cannot degrade into silence-means-approval.
`routeToErrorStep` (`coordinator.go:3184`) never writes the step's
`output_field`, so a routed-around critic genuinely leaves its field absent —
which is what makes the abstention path fire.

**Applied**: `sql_for_agents/171_experience_council_abstention_tolerant.sql`
(config-only, live on commit, no image roll), with a `DO` block asserting all
four end states including the deliberate asymmetry. Seed 167 patched in-place
so a re-apply cannot clobber it. Ledgered same sitting. Commit `da2c5dea3`.

**Migration numbering**: the handoff said "next free 168" — 168/169/170 had all
been taken by other sessions overnight. Claimed **171**. The re-check-at-
execution-time rule earned its place again.

**Chassis is now v1.0.1137**, not the 1135 the handoff recorded (rolled by
another session). Re-verified in-pod by binary string that the
`subject_type must be 'tool', 'pipeline' or 'experience'` literal is still
present before firing — it is.

**Run fired**: `fbe12212-1d7a-433b-9f70-d6988ce44d7b` (nothing in flight, pod
12h old so no 300s restart window).

### Result: 5 full rounds, escalated at the cap. CP2 still open — but the failure moved.

`COMPLETED / complete_escalated` at 11:07:34, ~20 min, 5 rounds. **This is the
first run to survive past round 1.** Verdict trail:

| Round | journeys | feasibility | honesty | mvp | decision |
|---|---|---|---|---|---|
| 1 | object ×3 | **approve** | object | object | revise |
| 2 | object | object | **approve** | object | revise |
| 3 | object | object | **approve** | object | revise |
| 4 | **approve** | object | **approve** | **approve** | revise |
| 5 | **approve** | object ×3 | **approve** | object | exhausted → escalate |

**Round 4 reached 3-of-4 approve** — one objection from converging. Nothing in
six previous runs got close to this.

**What is now proven, not just believed:**
- **Truncation fix holds under sustained pressure.** All 5 persisted plans are
  complete (closed criteria fence + `<!-- END EXPERIENCE_PLAN -->`) and stable in
  size — 14392 / 14032 / 14657 / 14594 / 14871 B. No growth spiral; LENGTH
  DISCIPLINE is doing exactly what it was written for. (Compare the previous
  run's first plan: 26,522 B, truncated mid-selector.)
- **`load_context` fix holds.** Feasibility's round-1 verdict was a clean
  **approve**, and every later objection cites the attached-components ground
  truth *correctly* — including using it to contradict the plan ("§3 asserts
  `.system-stats` has two live instances; ground truth shows ONE").
- **The abstention fix did not fire, because nothing flaked.** `abstained: 0` in
  all 5 rounds; all four critics returned every round. So 171 is deployed and
  correct-by-construction but **still unproven in anger** — do not record it as
  battle-tested. `bugs_open/008` item 5 simply did not recur this run.

**Why it did not converge — and it is NOT a harness defect this time.** Two
distinct behaviours, both visible in the trail:

1. **The MVP referee's objection is never acted on.** It said essentially the
   same thing in rounds 1, 2, 3 and 5: *defer the tool-arena rebuild / Journey C;
   the core loop only needs `arena.timer_seconds`.* The composer never cut it.
2. **Scope oscillation.** Each recompose satisfies the previous round's objector
   by ADDING specification, which enlarges scope, which re-triggers the MVP
   referee. Round 4 → 5 is the clearest instance: journeys and mvp both went to
   approve in round 4, then the recompose that answered feasibility pushed mvp
   straight back to object.

**A design contradiction underneath it** (found by reading, worth fixing
whichever way): seed 167 documents the MVP referee as *"ADVISORY (approve|object
only — an MVP opinion must not gate on its own)"*. But `decideCouncil`
(`diagnose_council_decide_action.go:263-267`) returns `revise` on **ANY**
reviewer's `object`. So the advisory seat gates exactly as hard as a veto seat —
the only difference is `revise` vs `rejected` routing. Four of five rounds were
decided by an objection; the seat we designed not to block has been blocking.
Resolve it deliberately in one of two directions, do not leave it ambiguous:
either exclude advisory seats from the revise trigger, or accept that it gates
and make the composer treat its scope cuts as binding.

**The round-5 objections are worth reading on their own merit** — this is the
council doing the job it was built for rather than reporting our own bugs:
- `gauntlet-cta` is a **shared** component on both `about` and `index`; editing
  its `js_content` for index would also execute on about's instance. Real
  cross-page side effect, unaddressed by the plan.
- `index` has `rebuild_policy='generic'` and is queued for header/footer
  reassembly, so component `js_content` edits **risk being clobbered** by that
  pass — the [[replan-clobbers-built-pages]] landmine, reached independently.
- The `.system-stats` instance-count error above.

**Side finding, verified and filed elsewhere.** Chasing the handoff's §4 item 1
("site components deactivated across 16 pages") showed that claim conflated two
tables: all 49 vonc `page_components` attachments are **active**; the
deactivation is 3 rows in `site_components` (`header`/`footer`/`head`). Not a
live breakage — all three serve baked `rendered_html`. But their repair items
have sat at status `detected` since **2026-07-11**, because `detected` is not
dispatchable without a triage promotion that never ran. That is `bugs_open/023`
class G, so the evidence went **there**, not into a new bug file — commit
`1260dd726`. Fleet-wide: `head` inactive on 11/11 sites.

### Run 8 (`17be3962`) — **CP2 CLOSED.** First approved EXPERIENCE_PLAN.

`COMPLETED / complete` (not `complete_escalated`) at 11:50:56, 5 rounds, ~29 min.

| Round | journeys | feasibility | honesty | mvp | decision |
|---|---|---|---|---|---|
| 1 | object | **approve** | object | object | revise |
| 2 | object | object | **approve** | object | revise |
| 3 | **approve** | object | **approve** | **approve** | revise |
| 4 | object | object | **approve** | **approve** | revise |
| 5 | **approve** | **approve** | **approve** | **approve** | **approved** |

**Verified, not assumed** — the checks that matter given what 171 changed:

- `abstained: 0`, `reviewers = 4` in the final round. This is a **genuine
  unanimous approval, not an approval-by-abstention**. Migration 171 made
  approval-with-a-silent-seat structurally possible for the three advisory
  critics, so this check is now mandatory on every approved round; do not skip
  it. `decided_by = "all reviewers approve"`.
- The approved plan is `is_current`, **14,414 B**, closed criteria fence +
  `<!-- END EXPERIENCE_PLAN -->`.
- Plan sizes across the run: 15504 → 13589 → 13722 → 13928 → 14414. It **shrank**
  after round 1 and stayed flat. Contrast run 7 (14392 → 14871, drifting up) and
  the pre-fix run (13303 → 16000, truncated).

**Migration 172 worked, and for the right reason — not by luck.** The evidence is
in the plan text, not just the verdict:
- §3 now carries `arena{status:"coming_soon"}`. The Arena rebuild — the thing the
  MVP referee demanded be deferred in four of five rounds of run 7, and which was
  never cut — is now deferred *and honestly labelled*, which also satisfies the
  anti-fabrication rule rather than dodging it.
- §4 now *explicitly answers* the referee where it keeps something: "Journey D's
  four onboarding CTAs stay as-is; each maps to an already-existing component
  needing only a `content_data`/template-binding fix, no new build, so the cost
  of testing them is negligible and the core loop is unaffected." That is exactly
  the (a)-apply-or-(b)-rebut discipline 172 required, in the composer's own words.
- The referee reached `approve` at round 3 and **held it** through rounds 4 and 5.
  In run 7 it oscillated straight back to `object`. The oscillation is gone.

**Still unproven, and I am not counting it as tested**: the 171 abstention path.
`abstained: 0` in all 10 rounds across runs 7 and 8 — no critic has flaked since
the fix landed, so the fall-through has never actually executed.
`bugs_open/008` item 5 remains the underlying defect and is still open.

**The landmine is now retired**: the previous handoff's "the current `is_current`
plan is an escalated/unapproved one, do not build from it" no longer applies. The
current plan is the round-5 approved one. T4 may build from it.

### 2026-07-19 — the js_content clobber risk, traced to mechanism (T4 precondition)

The council's round-5 objection said component `js_content` edits "risk being
clobbered" by index's pending generic rebuild. Traced it before planning around
it. **The risk is real but it is NOT what the objection describes, and
`rebuild_policy` does not fix it.**

**Ground truth first** (this reframes everything):

| component | level | js_content | on pages | page policy |
|---|---|---|---|---|
| `gauntlet-interface` | section | **3,909 B** | tool-gauntlet | **owned** |
| `provocation-card` | section | *empty* | index | generic |
| `gauntlet-cta` | section | *empty* | **about + index** | generic |
| `lobby-grid` | section | *empty* | index | generic |
| `brief-explanation` | section | *empty* | index | generic |

So the only component that has JS today sits on an **owned** page and is already
protected. The exposure is entirely **prospective** — it appears the moment T4
*adds* js_content to a component on a generic page.

**Who actually writes `js_content`** (`grep` over `platform/`, `internal/`, `cmd/`):
exactly one writer — `store_generated_component_action.go:420-430`, a single
UPDATE setting `html_template`, `input_schema`, `js_content` and `render_mode`
together. That is **component (re)generation**, not page rerender.

**Page rerender never writes it.** `rerender_single_page_action.go:149`
(`collectJSAssets`) only READS it, to emit `/tools/assets/{function}.js`. So the
three real risks are:

- **A — component regeneration overwrites the edit.** The genuine clobber. Any
  re-run of component generation over these functions replaces html_template AND
  js_content in one statement. It snapshots to `component_versions` first, but
  **warn-and-continue** if the snapshot fails (`:405-414`), so recovery is not
  guaranteed. **`pages.rebuild_policy` does NOT gate this** — it gates only
  `save_page_sections_action.go:148` and `reconcile_site_plan_action.go:202`.
  Marking index `owned` would not prevent it.
- **B — bulk rerender silently drops the JS asset.** The sharper one, and it runs
  the opposite way to the objection. `rerender_pages_actions.go:568` states
  outright that the bulk path "has no collectJSAssets equivalent — js_content
  assets are only emitted by single-page rerenders." So a bulk `rerender-pages`
  pass emits HTML referencing a JS file it does not republish: **DB row intact,
  live page broken.** And the stuck `deactivated_component` items
  (header/footer/head) route to `rerender-pages` — that IS the pass the council
  smelled, but the failure mode is a stale/missing asset, not a lost row.
- **C — shared-component blast radius.** `gauntlet-cta` is attached to **about
  AND index**, and js_content is library-level, so one edit executes on both
  pages. Not a clobber; a scope problem the plan must state.

**The truncation guard does not cover this.** `componentRegressionIssues`
(`component_write_guard.go:133`) is wired into `update_component_html_action.go`
only — `store_generated_component` merely references it in a comment
(`:1297-1302`). That is `bugs_open/021` exactly ("covers ONE durable-write path").
Worse for us: the guard compares **HTML only**, so `js_content` is never compared
on *any* path.

**Recommended resolution for T4** (in order):
1. **Prefer `page_components.content_data` over library `js_content`.** It is
   per-page, so it dodges A, B and C at once. The approved plan already leans
   this way — its Step 1 notes the content_data path works without a rerender.
2. **Anything that genuinely needs library-level JS should live on an `owned`
   page** — which is where the only existing js_content already sits.
3. **If a bulk rerender does run over a page with js_content, follow it with a
   single-page rerender** of that page, or the asset never re-emits (B).
4. Durable fix is `bugs_open/021`'s: extend the guard to `store_generated_component`
   and make it compare js_content, not just HTML. Not T4's job — pointed at, not
   forked.

### 2026-07-19 — filed to the diagnosis loop: the two republish paths carry complementary halves

Owner judged the js_content finding cross-cutting and asked for a cited
diagnosis to choose the fix. Grepping both bug dirs first (per CLAUDE.md) turned
up `bugs_open/024` — *"a tool-improver fix is written durably and NEVER reaches
the live page"*, filed hours earlier by the travelling-docs thread. Reading it
before filing sharpened the finding rather than duplicating it.

024 establishes that `RerenderSinglePageAction` is *"Simple concatenation - no
template re-rendering"* — it assembles stored `page_components.rendered_html` and
never re-renders `content_components.html_template`. But that is the same path I
had found calls `collectJSAssets`. Checking the other path closed the shape:

| republish path | re-renders template from `html_template` | emits JS from `js_content` |
|---|---|---|
| `rerender_single_page_action.go` | **no** — stale `rendered_html` (024) | **yes** — `collectJSAssets:121` |
| `rerender_pages_actions.go` (bulk) | **yes** — `:486`, `:680` | **no** — stated at `:569` |

**Neither path does both**, yet `store_generated_component_action.go:420-430`
writes `html_template` and `js_content` in a **single UPDATE**. So a component
change spanning both fields is split across two publish paths that each carry
only one half. That is the platform's own **split-contract-drift** class (the
name comes from the 42P10 incident), one level up: not two lists that must agree,
but two publish paths that must jointly cover one write.

This is a superset of 024's finding, not a rival to it — 024 is the html half.
Filed as one coherent bug, code-only (no `RUNTIME_SITE`: the defect is
structural, and the coverage key for code-only diagnoses is the SEED_SCOPE file
set). Coverage check passed — no other thread holds open work on those files.

- corr `8d86f110-447d-48ec-9857-32e7992326ca`, item_key
  `needs_diagnosis:rerender-paths-split-publish`, REF pinned to `57fc4f484`.
- SEED_SCOPE: `rerender_single_page_action.go:collectJSAssets`,
  `rerender_pages_actions.go`, `store_generated_component_action.go`,
  `update_component_html_action.go`.

**Why this was worth a loop run rather than a direct fix**: there are at least
three defensible repairs (give the bulk path a `collectJSAssets`; make the
single-page path re-render templates; or split the component write so the two
fields publish independently) and they have very different blast radii across
every site on the fleet. This is the "want a cited, auditable diagnosis for a fix
that changes behaviour fleet-wide" case in CLAUDE.md, not the "bug you can see"
case. T4 does not block on it — the recommended T4 route (prefer per-page
`content_data` over library `js_content`) avoids the defect entirely.

### 2026-07-19 — **the diagnosis loop REFUTED my hypothesis, and it was right**

Run `8d86f110` came back `REFUTED` in ~9.5 min with two static citations. I
verified the refutation against the code myself before accepting it. It holds.

> **CORRECTED 2026-07-19 — the "complementary halves" table in the two entries
> above is WRONG in its left-hand column.** The bulk path does NOT re-render
> body sections from `content_components.html_template`.
> `rerender_pages_actions.go:592 rerenderLoadSections` selects
> `COALESCE(pc.rendered_html,'') FROM page_components` — the *same* durable
> cache the single-page path reads.
>
> **How I got it wrong**: I grepped `html_template` in that file, saw hits at
> `:486` and `:680`, and inferred "the bulk path re-renders sections from
> html_template" **without reading either function**. Reading them now:
> `:486` loads the **head-seo-standard** component for the `<head>` block, and
> `:680` loads **contact-info** for a contact-info injection. Neither is body
> section rendering. This is the "read the function before changing it"
> convention, broken by me while writing a table that asserted a structural
> claim about the fleet.
>
> **What caught it**: the diagnosis loop, on the first iteration, by reading
> `rerenderLoadSections` — the function I never opened.

**What survives, and it is the half backed by the explicit code comment**: only
`RerenderSinglePageAction` emits `content_components.js_content` (via
`collectJSAssets`); the bulk path has no equivalent. The loop confirmed this
with the `:569` comment as its own citation. So risk **B** as recorded is real.

**Corrected picture:**

| republish path | body sections | JS assets |
|---|---|---|
| `rerender_single_page` | reads `page_components.rendered_html` | **emits** via `collectJSAssets` |
| bulk `rerender_pages` | reads `page_components.rendered_html` | **none** |

Neither path re-renders `html_template` for body sections. That is a *simpler and
worse* shape than the one I filed: not two paths carrying complementary halves,
but **two paths that both serve a cache, and one open question about whether that
cache is ever refreshed.**

**The loop named that open question**: "whether an `html_template` change ever
reaches `page_components.rendered_html` at all … is not shown here and is the
real open question." **`bugs_open/024` already answers it** — it proves the edits
do not arrive, because `build_status='pending'` is dead state nothing reads and
tool-improver's rerender routes to the stale-assembly path. So 024 is the
substantive bug and mine was a mis-framed superset of it, not a superset.

**Not re-firing the loop.** The surviving asymmetry (JS assets) is already
documented by an explicit comment in the source — there is nothing left to
*diagnose*; what remains is a design decision about what to do, which is not the
loop's job. The refuted framing is corrected here and the real question belongs
to 024's thread.

**The T4 recommendation is unaffected** — I checked rather than assuming. It
rested on risks A (component regeneration overwrites js_content,
`store_generated_component:420`, unchanged), B (confirmed by the loop) and C
(shared `gauntlet-cta`, verified empirically). None of those depended on the
refuted left-hand column. Prefer per-page `content_data` over library
`js_content` still stands.

### 2026-07-19 — T4 STOPPED at Step 0: the approved §3 data contract does not match the live loaders

Started T4 (owner: "please go ahead"). Step 0 is the blocking data gate — make
`/data/provocations.json` conform to §3. Before rewriting the file I went looking
for its consumers, because changing a shape without reading the consumer is the
mistake I had refuted this morning. **The consumers contradict the approved plan.**

**Who actually consumes the feed** (verified by fetching the live assets, not by
grep inference): NOT `gauntlet-interface.js` — that asset contains **zero**
references to `provocations.json` and no `fetch` at all. The real consumer is
**`/assets/js/snippets.js`** (14,293 B), a site-wide asset loaded on all four
pages, holding **three** loaders that each fetch the feed:

| loader | guard, read verbatim from source | needs |
|---|---|---|
| `fillProvocationCard` | `if (!section \|\| !data \|\| !data.today) return;` | `today` |
| `fillLobbyGrid` | `if (!section \|\| !data \|\| !data.arena) return;` then `var entries = Array.isArray(a.cards) ? a.cards : [];` | `arena.cards` |
| `fillArchive` | `if (… \|\| !Array.isArray(data.archive.entries)) return;` | `archive.entries` |

**Against §3 of the approved plan:**

| §3 says | loader needs | result if §3 is implemented literally |
|---|---|---|
| `today{…}` | `data.today` | ✅ compatible |
| `archive[]` — top-level array | `data.archive.entries` array | ❌ `Array.isArray(undefined)` → **early return**, archive never fills |
| `arena` = `{status:"coming_soon"}` **only** | `data.arena.cards` | ❌ `entries=[]`, `n=0` → **no card filled**, lobby grid never fills |
| `lobby[≤4]` top-level | — | ❌ **read by nothing** |

So implementing the approved contract verbatim silently blanks two of the three
runtime-fill regions. **Silently** is the sharp part: all three loaders fail
gracefully by design ("leave the shell as-is"), and G22 means the dead-control
and phantom-link checks *deliberately exempt* runtime-fill shells. The regression
would be invisible to exactly the guard rails built to catch this class.

**This is the council's first substantive escape, and the cause is the same class
we already fixed once.** Fix #3 taught `load_context` to surface component
attachments — but the loaders live in **`js_snippets`**, a table nothing surfaces
to the council. The council could see the components and could not see the
JavaScript that hydrates them, so feasibility approved a data contract whose
consumer it had never been shown. Same shape as the `component_level='tool'`
filter: **context lying by omission**, one table over.

**Also wrong in §3** (same root cause): "`gauntlet-interface` … real runtime
timer, **real fetch**". `gauntlet-interface.js` fetches nothing. Step 1's "ensure
prompt populates from `today` before Start" therefore requires a fetch that does
not exist in that asset today.

**Not proceeding on my own judgement.** A compatible shape exists that satisfies
every *honesty* rule in §3 (editorial-only stats, no counts, no fake live state,
no dead anchors) while keeping the nesting the loaders require — the substantive
rules and the key layout are separable. But quietly building a shape the council
did not approve would make CP2 meaningless. Escalated to the owner.

**Nothing was written.** No file published, no component touched. The live feed
still carries its fabricated stats (`1,284 Positions Filed`, `62% Disagree`,
`3h 12m Until Close`, six arena cards with `312 positions` and `Closes 18:00`).

### 2026-07-19 — runs 9/10/11: the contract seat works, and my rule for it is too strict

**Run 9 (`eae6278b`)** — 5-seat council, first fire. `contracts=object` all 5
rounds, escalated. The seat independently found the exact §3↔loader mismatch I
had found by hand, quoting the alias verbatim: *"lobby-grid-loader never reads
`data.lobby` — it only reads `data.arena.cards` (`var a = data.arena; ...
a.cards`)"*. That alias is precisely why 174 surfaced SOURCE instead of
regex-extracted paths; the choice paid for itself on the first run.

It also found three mismatches **I missed**: the loader reads
`a.eyebrow/a.title/a.subtitle/a.cta_label` which §3 never defines; the loader's
own comment fixes card DOM order at `cards[0..5]` (6 slots) while §3 caps
`arena.cards` at ≤4, so 2 of 6 render as stub against the plan's own "never a
stub" promise; and `entry.tag` has no source in §3.

**Run 10 (`20a1581b`) — died at `complete_refused`, zero council reports. My
regression.** `compose` failed with `stop_reason=max_tokens (output_tokens=32000)`.
174+175 grew load_context 13KB→39KB with real JS source; compose renders it
inline and had **no length rule** — LENGTH DISCIPLINE was added to `recompose`
only, by the run-6 fix, because compose had never needed one. Fixed by **176**
(length + quoting discipline; cap deliberately NOT raised — approved plans are
~14KB). **The fresh build v1.0.1138 is why this surfaced rather than corrupted**:
it decodes `stop_reason` and hard-errors on a capped completion, where the old
image would have persisted a truncated plan and drawn unclearable truncation
objections (run 6's spiral). New image caught my regression on its first run.

**Run 11 (`4f6a3997`)** — 176 confirmed: all 5 plans complete with trailer, and
**tighter** than before (12748/11401/12994/12933/12469 B vs ~14KB). Escalated;
`contracts=object` all 5 rounds again.

**175 confirmed too — "unverifiable" became "provably wrong".** Run 9: *"no
gauntlet-interface script source is in context — cannot verify."* Run 11: *"the
given gauntlet-interface js_content (ground truth) has NO fetch call and NO
reference to these selectors — only objectives/timer/stat-counter code exists."*
That matches my hand check: zero occurrences of all six Journey-E selectors.

> **MY DESIGN ERROR, three runs to see it.** The seat's rule — *"a pair you
> cannot verify from context is itself an objection"* — is correct for an
> EXISTING consumer and **wrong for one the plan will CREATE**. Run 11's
> objections include *"Step 2 creates a new provocations-archive-detail hydrator
> … this consumer's source is not in context and does not exist yet"* and
> *"Step 1 adds a new fetch … cannot be checked against any real source"*. Both
> are true and neither is a defect: a plan proposing new code has no source to
> quote, and that is what a plan IS. So the seat has a **false-positive class
> that blocks every greenfield step**, which is a large part of why runs 9 and 11
> could not converge. The rule needs to split:
> - existing consumer whose source CONTRADICTS the plan → hard objection (correct
>   today, and it is catching real defects);
> - consumer the plan explicitly creates/changes → NOT an objection provided the
>   plan states the exact access path the new code must implement AND §5 carries
>   an acceptance criterion that would fail if it does not.
> Unfixed as of this entry — owner decision pending, since it is a deliberate
> strictness trade-off, not an obvious bug.

**A third context gap the seat named** (it is systematically mapping the
boundary): `content_components.html_template` is not surfaced either, so it
cannot settle whether the archive item template carries a default `href="#"`.
Sharp objection, and one I had not thought of.

**Cost so far**: 3 council runs (~75 min + credits) since the seat landed, none
converged. It is finding real, provable defects every round; it is also blocking
on its own over-strict rule. Both are true and should be weighed together.

### 2026-07-23 — the runs 9/10/11 rule split APPLIED (migration 196) — owner-approved, config-only

The greenfield-strictness fix proposed in the "runs 9/10/11" entry above is now live.
Owner (vonc gauntlet AI-competitor-debate workstream) reviewed the exact rule split and
approved it verbatim. Migration `196_experience_contracts_greenfield_split.sql`
(`docs/agent_docs/sql_for_agents/`) replaces ONLY the strictness paragraph inside
`review_contracts`' prompt_template — everything else about the seat (its 4 judged
pairs, verdict shape, non-veto-but-blocking status, council wiring, recompose
visibility) is untouched and re-asserted by the migration's own `DO $$` block.

**The split, as applied:** for each producer/consumer pair, first decide whether the
consumer is EXISTING code (in context) or NEW code the plan proposes.
- EXISTING consumer: unchanged — must quote source; an unseen consumer is still a hard
  objection (this is what caught the real §3<->loader mismatch on run 9; not weakened).
- NEW consumer: no longer an automatic objection. Approve iff BOTH (1) the plan states
  the exact access path the new code must implement, and (2) §5 carries an acceptance
  criterion that would fail if that path is never built. Missing either → THAT is the
  objection (named precisely), not "the code doesn't exist yet".

Applied out of band via `psql -f` (config-only, no image dependency); ledgered in
`schema_migrations` with the snapshot id for rollback (`e0194bee-3b8e-4a38-a402-a031d4fe7a15`).
Byte-exact match verified against the live prompt before applying (avoids a silent
no-op `replace()`).

**Not yet exercised.** No council run has fired since. The next `092_TRIGGER_experience_plan.sh vonc.com vonc-spark-game` run against the pending debate-gauntlet requirement (a maximally-greenfield plan) is both the real feature work and the first live test of this split — if `contracts` still objects to genuinely-new consumers whose access path IS pinned and IS criterion-backed, the split needs another look before trusting it generally.

### 2026-07-23 — run 12 (corr fa4b77cd): 196's split PROVEN LIVE; REJECTED on correct sequencing

First council fire since the 196 greenfield split. **The split worked exactly as
designed**: `contracts` APPROVED every plan-created pair — its own words: *"New-consumer
pairs (gauntlet API/DOM, arena day-counter, provocations-detail query script) each state
exact access paths with a matching §5 criterion"* — and objected ONLY to two
existing-consumer contradictions it could quote (provocation-card-loader never reads
`data.lobby`; lobby-grid-loader wires clicks via a closure, not the `data-url` attribute
the plan described). No greenfield false-positives. The rule split is behaviourally
verified; CP2-class convergence is now blocked by real findings only.

**Verdict: REJECTED — feasibility veto, and the veto is CORRECT.** The plan (composed
under the 197 debate-gauntlet ruling) gates Journey A on POSTs to
{API_BASE}/api/v1/tools/gauntlet/* — an API that does not exist yet (its build is
in-flight in the feature-builder, corr c2a9fd27). Feasibility: *"a council-review plan
cannot gate an MVP on infrastructure whose existence is unconfirmed"*, and it named the
right sequence: *ship Steps 0–3 (all static-buildable) as the MVP; the live-API Gauntlet
is its own gated follow-up once API_BASE is confirmed reachable.* The loop is enforcing
build-before-promise — the exact honesty discipline it exists for.

Other seats: honesty APPROVED ("unusually disciplined" — Day-N labelled as
publish-count, no client-computed score, honest offline banner, no leaderboard);
journeys/mvp objections are direct build-round input (defence/verdict steps need
acceptance interaction checks; secondary-CTA journey missing; drop the arena rebuild
from the MVP cut). NOTE: the rejected plan is now `is_current` — per the seed's own
rule it MUST NOT be built from until a re-run converges. Re-fire 092 AFTER the
tools-api is deployed + smoke-POST-verified, with the liveness evidence carried into
the compose decisions block (the 197 channel).

---

## 2026-07-31 — NOTICE from the bugs_open/138 lane: your council seats' prompts changed

Not a request. Telling you because I edited config your lane owns, per the 2026-07-29
owner ruling that a shared mechanism's other consumers must be **told, not merely
measured**.

**What changed.** A LENGTH BUDGET paragraph was added to the prompt of all 10 of your
review seats — `experience-planner` (contracts, feasibility, honesty, journeys, mvp)
and `experience-approval-council` (checkability, deferral_honesty, honesty,
observable_outcome, prior_art). Inserted immediately before each prompt's `## Output`
heading. Nothing else in the prompts was touched, and no output schema was reordered.

**Why your councils were in scope.** Eligibility was defined by the council having a
`diagnose_council_decide` step — measured, not assumed — because that is the decider
whose behaviour the block describes. Both of yours do.
`domain-research-classifier` does not, and was excluded for exactly that reason.

**What changed about your guarantee — which is the part that matters to you.** Your
reviewers are now told, in their own prompt, that a reply cut off at `max_tokens` is
recovered as a fragment, marked `degraded`, and that a degraded `object` **gates the
round to REVISE regardless of the severities it assigned**. They are asked to keep
`notes` under ~250 words and to shorten `problem` texts rather than drop objections if
they are running long. So: **expect shorter reviews, and expect the same number of
findings.** The block says explicitly "cut words, never findings", precisely so this
does not quietly trade coverage for brevity.

**The evidence it works, since you should not take this on trust.** On
`review_editquality` (fix lane), measured by round spawn time against an unchanged
cap: 10 rounds spawned before the block peaked at 98.3% of cap (mean 9,848 tokens); 8
rounds after peaked at **55.0%** (mean 6,569). Cap did not move, so the budget alone
did it.

**Relevant to you specifically.** `review_deferral_honesty` truncated **3 of 5 calls at
cap 12000** in the 14 days to 2026-07-29 — the worst rate of any seat anywhere, and
already above the 8000 default, which is direct evidence that a bigger cap alone does
not fix this. `review_checkability` sits at p95 90.6% of 8000 (n=4, so watch rather
than act). Both now carry the block.

**If you want it off**, it is delimited and reversible: the paragraph runs from
`LENGTH — THIS IS A CORRECTNESS CONSTRAINT` to `— end length budget —`. Snapshots of
both your agent rows were taken immediately before the write and are in
`agent_definitions_backup` (`snapshot_taken_at` 2026-07-31 18:16:48 and 18:16:49,
reason `pre-update: seat length budget, bugs_open/138 candidate 4`) — verified to be
genuine pre-update copies, i.e. the block is absent from them.

Mechanism, evidence and the five measurement traps:
`docs024_key_docs_latest/bugfix_138_degraded_gates/HANDOFF_2026-07-31_continue_here.md`.
Tool: `scripts/apply-seat-length-budget.py` (idempotent; `--verify` reports live state).
