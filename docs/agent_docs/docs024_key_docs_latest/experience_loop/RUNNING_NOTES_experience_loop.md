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

---

## 2026-08-31 — lane resumed after five weeks; CONTRIB ask 1 built, LIVE, and its proposed rule refuted first

**Where the lane actually was**, read before touching anything (the memory entry was stale by
five weeks and still said "NEXT = T4 MVP build"):
- Last own-lane action was **run 12, 2026-07-23** (corr `fa4b77cd`): REJECTED on a correct
  feasibility veto — the plan gated Journey A on a tools-api that did not exist. That REJECTED
  plan is still `is_current` for `vonc-spark-game` and **must not be built from**. The vonc build
  cycle itself belongs to `gauntlet_dead_cta` (owner revision 2026-07-23).
- `HANDOFF_2026-07-28_appeal_dimension.md` §6 told a new thread to start with `contrast_ratio`.
  **Both its first moves are DONE by other lanes** — `no_horizontal_overflow`'s fix shipped
  `5042d5ecb`, and `contrast_ratio` now exists in the browser-runner
  (`internal/adapters/browserrunner/contrast_check.go`, with tests and cascade attribution).
  Do not restart either.
- 2026-07-31: our 10 seat prompts gained a length budget (another lane's notice, above).
- 2026-08-31: **the CONTRIB from the first paid build** arrived with four asks. That is the live
  inbox, and ask 1 is what this entry is about.

### Ask 1 built: `scripts/audit-listing-class-promise.py` + CronJob `listing-class-promise-check`

**Live and verified at the artefact, not at the apply:** CronJob `25 7 * * *` UTC,
`suspend=false`; manual run `lcp-verify-2`, pod `lcp-verify-2-2qgk9`, **`exitCode=0`**;
`doc_notes` receipt at 17:06:53Z under `subject_key='listing-class-promise'`. One row per run
including clean runs, so a missing row means the job did not run.

**Live finding: 1** — `leopardessconsulting.co.uk/blog.html`, heading "Latest Articles",
**7 of 13** items are `/guides/tool-*-guide` pages. Second site, same shape as boxingonline,
found by the detector rather than by a reader. Not dispatched anywhere yet — see "what I did not
do" below.

### The CONTRIB's suggested rule was REFUTED before a line of it shipped

It proposed: *a listing headed news/articles/latest must not be populated by pages whose
`page_type` is `guide` or `tool`.* `[MEASURED 2026-08-31]`:

```sql
SELECT page_type, count(*) n, count(DISTINCT site_id) sites FROM pages
 WHERE url LIKE '/guides/%' OR name ILIKE '%guide%' OR title ILIKE '%| Guide%'
 GROUP BY 1 ORDER BY 2 DESC;
-- blog-post 246 / 30 sites   guide 72 / 9 sites   section-index 12   content 4   landing 1
```

**Guides-as-`blog-post` is the fleet convention**, so the rule returns **zero on both real cases**
— including boxingonline, the one that motivated it (its four guides were `blog-post` at build
time; the webdesign lane retyped them to `guide` at 17:30 BST the same day, *after* the owner
complained). leopardess's eight are `blog-post` today. The refuted rule now runs as a printed ARM
on every detector run, so its zero stays a measurement instead of a memory. Landmine filed.

### Three missteps of my own, each caught by looking at findings instead of counting them

1. **I built a promise-reader that reads the items and calls it the promise.** First cut took
   heading **+ subtitle**. Of its 8 fleet findings, **4 came from the subtitle naming a class in
   passing** — homegarden's section prose ends "…the Garden Jobs Finder"; lampenkap's was an item
   excerpt, "the companion guide sets out the method". Worse: 139 of 159 listings carry their
   promise only in `rendered_html`, where "the first `<p>`" IS an item's own excerpt. The heading
   alone carries the promise now.
2. **`/blog/` is not "editorial"** — I made the same mistake as the rule I had just refuted, one
   column along, and reported dartsonline's "All guides" and agritec's "Technical explainers" as
   broken for listing `/blog/` items that are, by fleet convention, their guides. Now an explicit
   asymmetry: the check catches guides/tools under an editorial promise; it CANNOT catch articles
   under a guide promise, and says so in its own docstring and in every `doc_notes` row.
3. **`--self-test` passed while `write_doc_note` did not exist.** A patch silently failed to
   apply; no fixture called that path, so every case passed, and the CronJob found it in the
   cluster with a `NameError` on its first run. Split into `note_body()` so a fixture reaches it.
   **A self-test cannot vouch for a path it never calls.**

Plus a fourth, about controls rather than code: **I named a live page as the positive control and
it was repaired underneath me** (read at 16:45Z with four guides; gone by the 16:57:20Z scan; row
`updated_at` 16:57:33Z). Whether the flip landed before my read or between two writes is not
recoverable — `page_components` keeps one timestamp, not a history — and it does not matter: the
control read FAIL on a working detector inside twelve minutes. The fixtures are the controls now;
the live page is a note.

### Coverage, stated rather than implied

`[MEASURED 2026-08-31]` 159 listing instances fleet-wide (keys `articles` + `items`;
`nav_items`, 61 instances, is chrome and out of scope). Of those, **19 carry a heading in
`content_data`** — the other **139 only in rendered markup**, one has neither. A
`content_data`-only reader reported 88% of the corpus as "promises nothing", which is a clean
result that could not have come out otherwise. Verdicts on the current corpus: 41 kept the
promise, 116 named no class, 1 no promise text, **1 mismatch**, 3 suppressed by the
subject-vocabulary guard (loancalculator.co.uk, mortgagecalculator.co.uk, garden-tools.uk — a
class word inside the site's own name is subject matter, and the suppressions are printed, never
swallowed).

### What I did NOT do, and why

- **Asks 2, 3 and 4 of the CONTRIB are untouched** (zero-item index as an experience failure; a
  data-backing criterion for tool selection; routing `content-quality-auditor` into the new-build
  path). Ask 4 in particular is a dispatch question owned elsewhere — the auditor is active with
  49 COMPLETED fleet-wide and zero runs on the paid site.
- **No council submission.** `scripts/audit-listing-class-promise.py` is outside the gate's scope
  (`scripts/council-scope.sh` admits `scripts/pattern-check.py` only); nothing in `platform/`,
  `internal/`, `pkg/`, `cmd/config-key-audit/` or an appliable migration was touched.
- **The leopardess finding is not dispatched.** It sits in the detector's `doc_notes` row and in
  this entry; whoever owns that site should decide whether the guides move out of the listing or
  the heading changes. Either fixes it, and it is not my call.

## 2026-09-02 — round two of the paid-build review: two more checks built and LIVE, one refused

Peer message from the boxingonline session (round two of the owner's review, measured at the
served site 2026-09-02). Three defects; **I verified all three at the DB myself before building
anything**, because a peer report is a document, not a measurement.

| their claim | my independent check | verdict |
|---|---|---|
| two "News" entries in the primary nav | `/articles/index.html` (blog-index, nav_order 2) and `/news/index.html` (news-index, nav_order 100), both `nav_label='News'`, both `in_header` | CONFIRMED |
| the fight calendar page has no calendar | the page carries exactly two components — `hero-tool` and `Generic Text Block`. No listing component, no `articles`/`items`/`events` array, no tool component at all | CONFIRMED, and sharper than reported |
| articles promise specific news, deliver general essay | not re-measured; accepted as reported and REFUSED as a mechanical check (below) | routed, not built |

### Built and live: `experience-promise-check` (`40 7 * * *` UTC)

Verified at the pod (`epc-verify-1-ltcgv`, `exitCode=0`), at its `doc_notes` receipt (10:31:06Z)
and by diffing the deployed ConfigMap byte-for-byte against the committed script.

- **Rule A — two doors, one name.** Two active header entries, same label, different pages.
- **Rule B — a tool page with nothing usable.** `page_type='tool'` serving no control, no inline
  data and no runtime fetch. Live findings: **1** (boxingonline fight-calendar, 6,640 chars of
  prose about a calendar), plus **4** tool pages with no rendered html at all, kept in a separate
  "never built" bucket so they can never inflate rule B.

**Demand control, printed every run:** of 320 tool pages, **314 carry a control, 126 inline data,
12 fetch at runtime.** The check can come out either way; if those reach zero it has gone blind.

### The measurements that decided the rule, and the two that would have made it wrong

- **`cc.name ILIKE 'tool-%'` is not "has a tool".** My first cut counted a page's tool component
  by name and found **74 tool pages with none** — a big, exciting number. Opening them: the tool
  components are called `loans-compare-loans-loanzy-uk`, `funding-fit`, `patent-check`,
  `Ported Page`. **The naming convention was my hypothesis about provenance, not a measurement**
  (`a-subagent-report-is-another-doc`). Rule B judges what the page SERVES instead, and drops
  from 74 to 1.
- **The runtime-fetch escape earned its place on a real page, not on a worry.** vonc's
  `/tools/gauntlet/round.html` has no control and no inline data and is NOT broken — it fetches
  its round from the live API. Only 12 of 320 tool pages fetch at all, so the escape is narrow
  rather than a blanket silencer. Without it, rule B's first live report would have accused a
  working page.

### RULE A RETURNED ZERO ON ITS OWN MOTIVATING CASE — and the rule is fine

`/news/index.html` had `in_header` flipped to **false at 10:24:59Z**, five minutes before my
first fleet run at 10:29:37Z. Another lane repaired it mid-build. **This is the second time in
three days** — the listing-class check's positive control went the same way on 2026-08-31, and
that one I could not even prove the ordering of, because `page_components` keeps one timestamp
and not a history. Filed as a landmine of its own today: on a site under active review, a zero on
the motivating case must be resolved against the row's `updated_at` BEFORE the rule is suspected,
and the control belongs in a frozen fixture. Both of this lane's detectors now name no live page
as a control.

### The transport defect I caused, and the fix I refused

The first fleet run died with `unexpected EOF` — I had asked for every tool page's
`rendered_html` in one statement and the `kubectl exec` stream truncated a 948 KB response
mid-string. **The tempting fix was to push the three regexes into SQL** so only booleans crossed
the wire. I refused it: `--self-test` would then be exercising a Python mirror of a rule that
Postgres actually applies, in a dialect that differs — a test that vouches for something the
fleet does not run. Chunked the transport instead (8 pages a query, chosen by bisection), with a
short-read guard that exits non-zero rather than quietly understating every count. In the
CronJob this is all moot: it dials postgres directly.

### REFUSED: "does this page contain the thing its title asserts?"

`/blog/last-nights-result-underdog-shocks-the-champion.html` contains no result — it is an essay
on why underdogs win, citing 1990 and 2019. The decisive detail is that `/news/index.html` on the
**same site** carries "Filip Hrgovic beats Moses Itauma by stoppage", dated 31 August, which is
precisely the story the article's title promised. The site held it; the article did not use it.

That is a genuine promise-keeping defect and it is **not mechanical**. A proper-noun or date
count fires on every well-written general piece and stays silent on a specific-sounding essay —
it would be a check that looks like this lane's others and is not one. Recorded as refused in the
detector's own docstring and in the register, so nobody re-attempts it as a regex. It needs a
seat that can read; the copy lane already has the writer half.

### One open ask CLOSED by owner ruling

The tools-without-data criterion I was asked for on 2026-08-31 and could only propose now has a
ruling behind it (2026-09-02): *"should contain detailed, fact checked information that prefills
the form for the comparisons"* and *"The research agent should have researched what's on and that
is what should have appeared on this page."* Rule B is its mechanical half. The other half — a
criterion applied at tool-SELECTION time, so a tool whose site-supplied data set would be empty
is never chosen — belongs in the planner and the critics, and is still open.

### Still open from the CONTRIB

Ask 2 (a zero-item index as a first-class experience failure) is **partly** addressed: rule B
catches the tool-page form of it. The listing form — an index whose set is empty and which says
nothing true about that — is not built. Ask 4 (route `content-quality-auditor` into the new-build
path) is untouched and is a dispatch question owned elsewhere.

### 2026-09-02, later — the peer answered my stored-vs-served ask, and confirming it caught a fabrication of mine

**The answer inverted my fear.** I had flagged reading STORED markup as rule B's blind spot.
The boxingonline session probed all five of that site's tool pages, cache-busted, three signals
each: **four agree, one differs, and the one that differs is rule B's only true positive.** The
served fight-calendar page's single control is `<button class="mobile-menu-toggle">` inside
`<header>`; `page_components.rendered_html` excludes chrome, so stored correctly sees zero.
Switch to served bytes and every page in the estate has a control, rule B never fires, and the
fleet result is a clean zero. **Stored is right because the question is "does the page BODY offer
the reader anything", and chrome is not the page.** Recorded in the detector's docstring and in
SQ-005; the entry's one open risk is now closed rather than carried.

**Confirming it, I nearly registered a fabricated mechanism.** I ran
`count(*) FILTER (WHERE rendered_html ~* '<(button|input|select|textarea)\b') FROM site_components
WHERE slot_name='header'` → **0 of 33**, and reasoned that the toggle "exists in no DB row and must
be injected by the publish template" — sharper than the peer's account, and false. `\b` is
**backspace** in Postgres, not a word boundary; `\y` is. The truth is **30 of 30**. What caught it
was an `ILIKE '%mobile-menu-toggle%'` returning **59 rows in the table my regex had just called
empty**, and choosing to chase the contradiction rather than the conclusion I preferred — the
register sentence was already drafted. Landmine + WRONG_CALLS entry filed. Both detectors in this
family run their regexes in Python, so neither check was ever affected; only my design-time
censuses were, and one of them (`\.json\b` in the fetch probe) is why the SQL census said 6
runtime-fetch pages while the Python rule says 12.

**Two things the peer added that changed the artefacts, not just the record:**
- **A rule A finding names two doors and must not imply which to close.** On its own motivating
  case the newer duplicate (`/news/index.html`) holds the site's only real, dated, sourced boxing
  results, and the six `/blog/` essays are the disposable half — so the nav symptom points at the
  page you must NOT delete. That caution is now printed in the check's own `doc_notes` body, where
  whoever acts on a finding will see it, rather than living in a lane doc they will not read.
- **The refusal now has a concrete disproof.** The underdog essay names nine proper nouns and two
  dates while containing no news, so a proper-noun-or-date approximation would score it ABOVE a
  correct short report of the actual result. It inverts the case rather than missing it. That is
  worth more than the refusal argument was on its own, and it is in the docstring.

**Rule A's zero was not a fix by the peer's lane:** site_delivery_and_editor pulled `news-index`
from the header flags as a deliberate INTERIM under the owner's ruling, keeping the page
reachable, pending structural work. So the rule reported zero because the estate was correct at
that moment — which is what a detector should do.

**Their `while read` + `kubectl exec -i` trap is ALREADY DETECTED and needs no new entry**:
`check_stdin_eater` in `scripts/pattern-check.py` (016b §9 #20) fires on exactly that shape, and
its own comments use `kubectl exec -i pod -- psql` as the worked example. It scans committed
`.sh`/`.bash` files, so it structurally cannot see an ad-hoc command typed into a session, which
is where they hit it. Told them rather than filing a duplicate.

### 2026-09-02, round three — rule C built, the peer's stronger formulation refused for want of an INPUT, and my own 08-31 dismissal refuted

The peer brought a third instance of the listing-class family and asked me to judge their two
candidate formulations **on a false-positive census rather than on their three instances** —
exactly the right question, and the reason I could answer it at all.

**Their (B), "compare what the typed query ASKED for against what it GOT", is the better shape
and cannot be built: the estate does not retain the first half.** `[MEASURED 2026-09-02]`
`page_components.data_path` is empty on **every** row fleet-wide; only **6** page_components
mention `query.` anywhere in `content_data`; of **1,082** active pages only **28** carry a
`page_spec` at all and **3** name a query or a source. The resolver runs at build time and stores
the resolved items array alone. So the check has no input for ~99% of listings — a different
refusal from the title/body one (that needed a reader; this one needs a column that does not
exist), and it becomes buildable the day anything starts persisting the query.

**Their (A), refined, is rule C and is now live.** An index-role page whose own directory holds
active pages while its listing shows zero of them. Two escapes, both measured rather than
imagined:
- **Index-role only.** Unrestricted, the rule fires on every tool page and individual guide
  carrying a "related content" block — `loanandmortgagecalculator.co.uk/tools/early-settlement/`
  lists 6 items, none from `/tools/`, and is doing its job. That is the large false-positive
  population the peer correctly worried about; the role restriction removes it.
- **`pages_in_dir > 0`.** An index whose directory is empty has nothing of its own to list.
  Six homegarden month indexes and five `/news/` indexes sit in exactly that state.
A mixed index passes: the rule fires only on **zero**. Findings fleet-wide: **2**.

**AND THE CENSUS REFUTED MY OWN DISMISSAL FROM SUNDAY.** On 2026-08-31 I looked at dartsonline's
`/guides/index.html` listing twelve `/blog/` items and cleared it, reasoning that this estate
files guides under `/blog/` — a convention I had just measured (246 across 30 sites). The
reasoning was sound and the conclusion was wrong: **`/guides/` on dartsonline holds nine tool
guides**, all orphaned, created 2026-07-31 to 2026-09-01, and the index lists none of them. I
used a real convention to settle a question it does not touch, and never asked whether the
directory contained anything. WRONG_CALLS entry filed. The practical cost: SQ-004 shipped with a
stated blind spot ("cannot catch articles under a guide promise") that was avoidable — the
directory answers it with no class inference at all.

**The peer's causal story is narrower than the mechanism, and this matters for whoever fixes
it.** They report that retyping boxingonline's four guides `blog-post`→`guide` (the fix for
instance 1) is what made instance 3 possible, and warn that reverting the retype would restore
instance 1. True for that site. But **dartsonline has the same defect with no retype in its
history** — its nine guides are still `blog-post`. So the retype is one route into the state, not
the mechanism. Whatever the structural fix is, it has to work for a site nobody has retyped.

**One peer claim did not survive my read, and I did not propagate it.** They reported that a
`page_rerender` only re-resolves when `spec->>'reason'` is `section_data_resolved` or
`template_changed`. `platform/livespec/rerender_reasons.go:85` says close to the opposite for the
first of those: `image_landed` and `section_data_resolved` **deliberately** do not set
`StampAlways`, because they carry REB-001's designed degrade — *a reason without a `component_id`
falls back to assemble-only*. So the accurate caution, which is what rule C now prints beside its
findings, is: a completed rerender does not prove the listing was re-resolved; check the item's
reason **and whether it carried a `component_id`** before reading a persistent finding as a
failed fix.

Verified live: `exitCode=0`, ConfigMap `bh76d66662`, `doc_notes` receipt carrying all three rules.

---

## 2026-09-02 — `content-quality-auditor` into the new build path: §2 says STOP, and here is why

Working the HANDOFF of the same date. Its §2 is an explicit gate: *read what a run actually
produces before wiring anything in; if the audit does not name what the owner complained about,
"the job is bigger and the prompt is the work — say so plainly rather than shipping a route to a
seat that will stay quiet."*

**The gate fails. Do not route it as it stands.** Six findings, all `[MEASURED 2026-09-02]`
against live config and the live DB.

### 0. First, a correction to the handoff's own arithmetic — and it is a real signal

The handoff records **44** auditor runs. Today the same query returns **42**. A count of rows
that goes DOWN means `orchestration_states` is being reaped, so every "N runs since" figure
about this seat has a shelf life measured in days. Quote it with the date attached or not at all.

```sql
SELECT count(*) FROM orchestration_states WHERE collected_data ? 'content_audit';  -- 42
```

### 1. The auditor structurally cannot see the pages the owner complained about

`load_page_content` is the whole of its sight, and it reads:

```sql
WHERE p.site_id = $1 AND p.name IN ('index','about','services','contact')
```

Four hardcoded page names. On boxingonline that is **3 of 22 pages** (no `services` page exists).
Fleet-wide it is **92 of 1,196 pages across 36 sites — 7.7%**, averaging 2.56 pages seen per site.
`services` exists on **7 of 36** sites, so a quarter of the budget is usually spent on nothing.

The owner's three complaints were the padded `/guides/tool-*-guide.html` pages, the
`/articles/index.html` manifesto, and the `/tools/fighter-comparator/` form. **All ten guide,
tool and index pages are outside the four names.** The seat is a four-page brochure auditor
pointed at sites averaging 33 pages.

This is the answer to §2 and it is not a matter of prompt tuning: no wording change lets an LLM
review a page that was never put in its context.

### 2. Of the little it sees, it sees the first 1,000 characters

`LEFT(string_agg(...), 1000)` per page. Index pages fleet-wide average **28,180 chars**, so the
landing page is sampled at **4.5%**. about 11.0%, contact 14.2%, services 13.6%.

### 3. …and ~43% of that budget is stylesheet

`rendered_html` carries `<style>` blocks. Fleet-wide over 2,851 components, **42.8%** of
`rendered_html` is CSS. On boxingonline's index the `<style>` block starts at **character 1** and
never closes inside the window — **999 of the 1,000 chars the LLM receives are CSS**. `about` and
`contact` lose 426 and 417 chars respectively to a style block that also never closes.

`content_data` (jsonb, avg **2,145** chars vs 6,297 for the html) holds the same content
structurally with no CSS, and is the better input.

> **MISSTEP, mine, logged because the metric looked fine:** my first pass measured "prose chars in
> sample" by stripping `<style.*?</style>` then `<[^>]*>`, and reported index as **993/1000 prose**
> — the most CSS-choked page of the three scored best. The style block is *truncated*, so there is
> no closing tag to match, the strip no-ops, and the raw CSS body survives as counterfeit prose.
> **A tag-stripping metric silently inverts on truncated markup.** The honest measure is
> positional: where does `<style` start, and does `</style>` appear at all.

### 4. The sample window drifts across runs with no content change at all

`string_agg(pc.rendered_html, ' ')` carries **no `ORDER BY`**, so the aggregation order is
unspecified. This is not theoretical:

- the 12:35 run's stored `page_samples` contains the CTA prose "New fights get announced…";
- that prose sits at char **15,154 of 18,553** — far outside any 1,000-char window taken today;
- all three index components were created 06:37 and have `updated_at = created_at`, i.e. **the
  content did not change between the run and now**.

Same rows, same bytes, different leading component. Fix is `ORDER BY pc.position` (the column is
`NOT NULL`).

> **MISSTEP, and the reason it nearly got recorded the wrong way:** I first tested this by hashing
> the sample **5 times in a row** and got five identical hashes, and wrote that I would *not*
> claim non-determinism. Five consecutive runs share a plan and a heap; the drift shows up across
> **hours**, not across seconds. **A stability test whose window is shorter than the disturbance
> cannot see the disturbance** — the sampling interval is part of the claim, and mine was
> unstated. The stored `page_samples` from the earlier run is what actually settled it, because it
> is a record from *before* the window moved.

### 5. The prompt reviews a dimension the schema cannot express

Prompt REVIEW list: TONE · GAPS · CTA · DIFFERENTIATION · **AUDIENCE**.
Declared enum: `tone|gap|cta|differentiation|`**`content`**.

`AUDIENCE` has no enum value and `content` has no review dimension, so the model improvises.
Across all stored audits, **210 findings**: gap 64, content 45, differentiation 45, cta 43,
**audience 10 (outside the declared enum)**, tone 3.

### 6. Even the findings it does produce have nowhere to go

Beyond `filing_mode='record'`, the categories themselves are unrouted — `site_work_items` carries
`capability_gap` rows saying so in the platform's own words: **30** for `cta`
("no handler for audit category \"cta\""), **3** for `audience`, **1** for `tone`, spanning
2026-08-20 → 2026-09-02.

So §4's warning ("something must READ them") is already the observed steady state for this seat's
largest category, not a risk to guard against.

### What this changes about the task

The handoff's shape — insert a `call_agent` between `apply_site_design` and `update_site_status`
— stays right, and §3's argument for `call_agent` over `spawn_agent` is untouched. But it is now
**step 3 of 3**, not the job:

1. **Widen and clean the input** (the load-bearing fix): drop the four-name allow-list, order by
   `position`, read `content_data` or strip `<style>`, and raise the per-page budget. Without this
   the seat cannot see guides, tools or indexes on any site.
2. **Reconcile prompt dimensions with the category enum**, and add the promise-keeping questions
   the owner actually asked for (does a listing list its own class; does a tool carry data; does a
   guide earn its place beside the tool).
3. **Then** route it into the build path, and answer §4's record-vs-dispatch question.

Doing 3 alone would put a seat in the build path that is blind to 92% of the site by construction
— the same silence as defect 4, moved earlier and harder to see.

**Diagnosis substitution, stated per the 2026-07-31 owner ruling:** no `090` filed. The criteria
that trigger it (cause non-obvious after a quick look; cause not where the symptom is) do not
hold — the cause is four literals in the seat's own SQL, read directly, and every claim above is
first-hand measurement against live config and the live DB rather than inference.

---

## 2026-09-02 (later) — 694 APPLIED and live-verified; and a peer's site exposed a blind spot in BOTH my detectors

### 694 is applied, and the evidence is at the artefact, not the status

Council: **APPROVED at round 2**, corr `d52a0e45-5c64-4d32-a1ab-f73532684d37`. Round 1 was REVISE
and was worth every minute — see the previous entry's correction. Applied by hand 14:36:08Z, then
`run-migrations.sh --record-only` so no other session's `--apply` re-runs it.

Live check, read back from `agent_definitions` **after** the apply: allow-list gone, non-greedy
strip present, `ORDER BY pc.position` present, the four promise dimensions present,
`filing_mode` still `record` (the owner's ruling of today, preserved).

**The measurement that matters is fleet behaviour, not my own assertion.** Every
`content-quality-auditor` run since 14:36:08Z:

| run | pages sampled | distinct page_types | carries the new `page_type` column |
|---|---|---|---|
| 19:09 | 14 | 4 | yes |
| 18:53 | 18 | 6 | yes |
| 17:50 | 18 | 7 | yes |

Before 694 the same field held **3** rows and no `page_type` key at all. On boxingonline the seat
now samples **18 pages across 8 types / 20,974 chars**, against 3 pages of which the index was
999/1000 CSS.

> ⚠ **Do NOT cite a boxingonline audit as proof yet.** Its most recent run is **06:23Z, pre-694,
> 3 pages**. The improvement sweep takes one site per 15-min tick across ~54 sites, so it has not
> come round again, and a manual dispatch I published at ~14:37Z produced **no orchestration row**
> (unexplained; not retried, per CLAUDE.md's "a missing row is almost always latency"). The
> post-694 rows above are OTHER sites. The change is proven fleet-wide and NOT yet on the
> motivating site — which is exactly the shape `a-pass-from-a-blind-check-outlives-the-blindness`
> warns about, so it is written down rather than rounded up.

Also worth recording: the seat's config `updated_at` moved to **15:38:29Z**, an hour after my
apply, with **no snapshot taken** (last `agent_definitions_backup` row is my own 14:35:30). No
migration applied in that window names this agent and the step list is still exactly 694's eight,
so nothing I assert on changed — but on a tree this many sessions share, "my change is still
there" is a query, and I ran it.

### The designblog exchange — my two detectors are blind to the case the owner is complaining about

The `designblog.co.uk` lane messaged: four listing pages serving ZERO items and carrying prose
about their own brief instead (`/glossary.html`, `/inspiration/`, `/the-design-feed/`,
`/uk-studios-directory/`). Owner's words: *"the glossary has text about the brief and is not a
glossary… the directory is empty."*

**First answer — schedule — is true but is NOT the interesting one.** Neither cron had seen the
sites: SQ-004 last fired 07:25:11Z and the three new sites have `created_at` 11:47:22Z (all four
updated 18:34–19:12Z); SQ-005's `lastScheduleTime` is **empty** — the CronJob was created today at
10:30:27Z, i.e. after its own 07:40 slot, so it has **never fired on schedule** and first runs
tomorrow. The three receipts dated today are my own manual triggers from building it.

**The real answer is that running them by hand right now still reports ZERO**, and Rule C cannot
see this case by construction:

```sql
AND jsonb_typeof(COALESCE(pc.content_data->'articles', pc.content_data->'items')) = 'array'
), sized AS MATERIALIZED (SELECT * FROM inst WHERE jsonb_array_length(arr) > 0)
```

Two independent filters drop an empty index before the rule runs. Measured on designblog: all
**11** component rows under those four pages have `arr_type` **NULL** — no `articles`/`items` key
at all — so they die at the FIRST filter, before the `> 0` one is even reached. **Rule C asks
"does this index list things from OUTSIDE its own directory?", which presupposes it lists
something. An index that lists NOTHING is invisible to it.** And `/glossary.html` is
`page_type='content'`, outside Rule C's selector entirely — so the fix needs a content-class
trigger, not merely a page_type widening.

This is the CONTRIB's open ask-2 listing half, now with a **second independent instance**, which
is what turns it from a note into work. Taken.

> **A DEFECT IN MY OWN DETECTOR, found by reading my own output instead of skimming it.** The
> designblog run of `audit-listing-class-promise.py` printed
> `positive_leopardess: FAIL — classifier drifted or site changed`. That is neither a designblog
> finding nor a regression: **the positive control is a leopardess page, and `--site
> designblog.co.uk` filters the control's own case out of the corpus**, so it reports FAIL where
> it means N/A. Re-run with `--site leopardessconsulting.co.uk` and it PASSES and still finds the
> real 7/13 off-class mismatch. So **every `--site` run against any other domain currently prints
> a failed positive control** — which trains a reader to ignore the control line, and that is the
> one line that says whether a zero is trustworthy. Until fixed, SQ-004's designblog zero is
> **UNTESTED, not clean**, and I told the peer so rather than letting them read it as a pass.

Peer context for the rule build, from `bugs_open/444` (portfolio positioning): the four remakes'
feeds have **0** `content_sources` rows, glossary/inspiration have **no item producer anywhere in
the estate**, and the directory kind does not exist. So the rule will initially fire on pages that
**cannot** fill themselves — correct behaviour, and 444's own fix candidate (plan validation
refusing a listing page whose item source resolves to zero) is the upstream door-closer the
detector would then hold shut. Worth stating in the rule's docstring so nobody "fixes" the
detector for over-reporting.

### Open after this session
1. **The empty-index rule** (SQ-005 rule D, or a widening of C): an index that lists nothing, and
   whose page says nothing true about that. Must trigger on content-class, not page_type alone.
2. **SQ-004's `--site` control bug** above — cheap, and it undermines every scoped run until done.
3. Routing `content-quality-auditor` into `site-work-orchestrator` (the original task's step 3),
   held until a post-694 audit is observed on a real build.
4. The planner-side half of CONTRIB ask 3 (refusing to select a tool when we hold no data for it).

---

## 2026-09-02 (close) — fresh chassis rolled; post-roll verification BLOCKED on token expiry

Owner reports a fresh chassis build deployed (~21:0xZ). Two notes, and the second is the reason
this entry exists rather than a tick.

**1. 694 should be untouched by a roll, and I did not get to prove it.** The migration edits
`agent_definitions.default_config` — DB config, live immediately, no Go — so a chassis roll has no
mechanism to revert it. That is a sound argument and it is still an argument. At 21:09Z every
cluster call returned `You must be logged in to the server (Unauthorized)`, including
`kubectl get --raw /version`, which is the fleet-wide kubeconfig-expiry signature (owner refreshes;
context `personae-uk001-prod-agent-chassis-cluster`). So the post-roll check is **OUTSTANDING, not
passed**, and it is written as step 0b of the new handoff with the exact queries and the expected
row.

**The trap I am deliberately not walking into:** with an expired token, every query returns nothing,
and "nothing" is indistinguishable from a clean result. Had I run the verification anyway and read
the empty output as "no problems", that is a false pass built on a dead connection — the same shape
as `a-post-fix-ZERO-needs-a-DEMAND-control`. Nothing verified today after 21:09Z should be quoted.

**2. What still needs proving once the token is back**, in priority order: (a) 694's markers still
present on the live row; (b) the behavioural check — post-roll audits sampling 14–18 pages with
`has_new_columns = t`, because the config reading right is not the same as the seat sampling right;
(c) boxingonline's first post-694 audit, which has **still never happened** (its latest is 06:23Z,
pre-694, 3 pages).

**Handoff written:**
`docs/agent_docs/docs024_key_docs_latest/experience_loop/HANDOFF_2026-09-02b_after_694_the_empty_index_rule_and_the_routing_still_owed.md`
It carries the two outstanding promises to peer lanes (designblog: build the empty-index rule and
re-run; vetcomparison: re-run both detectors plus the widened auditor after migration 701), the
SQ-004 `--site` control bug, and the routing step that is still the original task.

---

## 2026-09-02 (post-roll, token refreshed) — 694 survives the roll, the NEW DIMENSIONS FIRE, and my manual dispatch does nothing

Token refreshed by the owner; the §0b checks from `HANDOFF_2026-09-02b` are now run.

### 0b(a) PASS — 694 survived the chassis roll to v1.0.1355

All markers present on the live row: allow-list gone, `<style[^>]*?>` non-greedy strip,
`ORDER BY pc.position`, all four promise dimensions, `filing_mode` still `record`. Pods
`agent-chassis-8ddbf8958-*` started 20:56/20:57Z on `v1.0.1355`.
Note the row's `updated_at` has now moved three times since my apply (14:36 mine → 15:38 →
**20:55:58**, just before the roll), each time with **no snapshot taken**. Every marker survives
each write, so nothing I changed is being reverted — but the check is cheap and stays worth running.

### 0b(b) PASS — and this is the result the whole task was for

The sweep audited **gamesdesign.co.uk** at 21:27/21:28Z: **21 pages across 6 page_types, 9,482
input tokens**. Pre-694 the same seat sent **1,744** tokens averaged over 3 pages. That is a 5.4x
widening of what the model is shown, live, unattended.

**The four new dimensions are not decoration — they fired, and they produced the owner's own
complaint classes on a site nobody pointed them at:**

| finding | dimension | what it says |
|---|---|---|
| [HIGH] `/guides/index.html` | 6 PROMISE | hero CTA labelled "Launch Cooldown & Resource Cost Analyser" links to `/games/auto-battler/` — names one tool, opens another |
| [HIGH] `/guides/economy-basics/` | 6 PROMISE | body promotes the Sink & Faucet Modeller; both CTAs go to two different tools, neither of them it |
| [HIGH] `/games/p2p-networking/` | **8 TOOL DATA** | *"asks the reader to supply a Host ID obtained from another user in a separate session … A tool that requires external coordination to produce any output is a form stub, not a working tool."* |
| [MED] `/index.html` | **9 GUIDE PROMINENCE** | *"The explainer leads the usable thing it explains, inverting the priority the site's own value proposition demands."* |

The third and fourth are the owner's boxingonline complaints — *"the tool requires the user to
input all the details"* and *"the guide is more prominent than the tool"* — **restated
independently by the seat, on a different site, unprompted.** That is the capability he asked for,
working. It is what a checker that can see the site looks like.

### 0b(c) STILL OPEN — boxingonline has not been reached, and I could not force it

Its latest audit remains **06:23Z, pre-694, 3 pages**. The sweep takes one site per 900s tick over
~54 sites and has not come round.

> **MISSTEP, corrected here in full, because I published the wrong reading twice before catching
> it.** I dispatched the seat by hand at boxingonline twice (corr `30b5a2c5`, then `34cb071c`).
> Both produced **no `orchestration_states` row**. First I read that as latency (CLAUDE.md's own
> advice). Then, seeing chassis logs marked `stateless:true` on the generic topic, I hypothesised
> that generic-topic runs simply persist no row — and *then* I found two successful
> `content-quality-auditor` LLM calls in the window and reported that "the audit ran".
> **All three readings were wrong.** The two LLM calls carry correlation `0f0dc48e`, which is
> **neither of mine** — they are the sweep's run on gamesdesign.co.uk. Checked directly:
> ```sql
> SELECT count(*) FROM llm_call_log       WHERE correlation_id IN ('30b5a2c5…','34cb071c…');  -- 0
> SELECT count(*) FROM orchestration_states WHERE correlation_id IN ('30b5a2c5…','34cb071c…');  -- 0
> ```
> **My dispatches did nothing at all: zero rows, zero LLM calls, twice.** The publish receipt was
> genuine (`kafka_publish_checked` printed PUBLISHED both times) — so a real receipt on a real
> topic proves the MESSAGE LEFT, and nothing whatever about anything consuming it.
>
> **The check that would have saved all three errors is one query, and it is the same query every
> time: does the work carry MY correlation?** I had a correlation id in hand from the moment I
> published and did not join on it — I looked at a time window instead, and a busy fleet always has
> something in a time window. `scratchpad/fire_cqa.sh` is therefore **NOT a working dispatcher** —
> do not copy it, and do not re-run it expecting an audit. Modelled on `fire-offer-analyser.sh`,
> which presumably works for its own seat; what differs for this one is undiagnosed and was not
> worth chasing, because the sweep reaches every site unaided.

**So the honest status of the original task:** 694 is live, proven across the estate, and its new
dimensions demonstrably produce the owner's complaint classes. It has still never run on the site
that prompted it, and that remains 0b(c) for whoever picks this up — by WAITING for the sweep, not
by re-running my script.

---

## 2026-09-03 — the applied migration was NOT the committed migration, for ~7 hours

Caught while checking the tree was clean ahead of a chassis build. `git status` showed
`694_content_quality_auditor_can_see_the_site.sql` **modified and uncommitted** — 45 lines that
exist in the live database and did not exist in git.

**How it happened, and it is a clean example of a rule I had already read.** The sequence was:
round 1 REVISE → fix the greedy regex → **commit `0fa679a28`** → submit round 2 → **APPROVED with
three advisories** → act on two of them (the NULL-safety refusal and the execute-the-embedded-SQL
verify) → **apply to the live DB**. The advisory fixes landed between the commit and the apply, and
I committed nothing in between. Applying felt like the terminal act, so the commit felt already
done — but the apply is what makes it *live*, not what makes it *recorded*.

**The cost, had this session ended there:** the next reader of `694…sql` would have found a file
whose verify block has no NULL guard and never executes the query, and would have concluded the
live seat was applied without those protections. `agent_definitions_bak_694` and the live row would
both have disagreed with git, with no way to tell which was authoritative. The rollback file would
still have worked, so nothing was at risk — but the audit trail was wrong for about seven hours.

**The check, and it costs nothing:** `git status <the file you just applied>` **immediately after
applying**, not at the end of the session. "Live" and "committed" are independent facts and an
apply advances only one of them. This is the auto-memory entry
`live-and-committed-are-independent-facts` — which I had loaded, and which describes this exact
shape ("sweep file-vs-live before you finish"). Reading a rule is not running it.

Committed now. The live DB and git agree again; nothing was lost, forward-only holds.

**Also noted for the incoming build:** a chassis build is in flight, deploying within the hour
(current tag `v1.0.1355`, pods up 2026-09-02 20:56/20:57Z). 694 is DB config so the roll cannot
ship or revert it, but the §0b checks in `HANDOFF_2026-09-02b` are **per-roll, not once** — re-run
them against the new tag. Boxing Online has *still* not been audited post-694 as of this check.

---

## 2026-09-03 09:27Z — v1.0.1356 rolled; 694 verified across a SECOND roll

New chassis `v1.0.1356` (pods up 08:57:46/08:58:07Z). Per-roll checks re-run, because 694 being DB
config is an argument for it surviving, not a check that it did.

- **Markers: PASS.** Allow-list gone, non-greedy strip, `ORDER BY pc.position`, all four
  dimensions, `filing_mode` still `record`. The row's `updated_at` moved again to **08:56:53**,
  immediately before the roll — the fourth such write (14:36 mine → 15:38 → 20:55:58 → 08:56:53),
  every one with **no snapshot taken**, and every one leaving my markers intact. Something touches
  this row at roll time; it is not reverting 694, and the check is cheap enough to keep running.
- **Behaviour over the 18h to the roll: PASS**, from `llm_call_log` — 36 calls, **0 failures**,
  avg **~7,000** input tokens (3,612–11,129) against a pre-694 average of **1,744**.
- **Behaviour since the roll: not yet observable, and I am not calling it either way.** 0 auditor
  calls in 30 minutes. **Demand control run in the same breath: 45 LLM calls across 5 agent
  types**, and `improvement-sweep` last triggered **09:24:43**, 3.5 minutes before the check, on a
  900s interval. So the fleet is live and the sweep is ticking; it takes one site per tick and only
  some need a content audit. **Two ticks is not a sample.** Re-check in an hour.

**Why the control matters here and not as ceremony:** a zero from a seat that runs opportunistically
looks identical whether the fleet is quiet, the sweep is dead, or the seat is broken. Without the
45-calls-across-5-agents line and the sweep's `last_triggered_at`, "0 auditor calls after a roll"
is exactly the shape that gets written up as a regression. It is not one.

**Handoff written and superseding both 09-02 files:**
`docs/agent_docs/docs024_key_docs_latest/experience_loop/HANDOFF_2026-09-03_experience_loop_continue_here.md`
Self-contained, so the next chat reads one file rather than chasing a chain of three. It carries
the 24h-reaping trap (§4) that invalidated the previous handoff's own verification method, the two
outstanding peer promises, the empty-index rule, the SQ-004 control bug, and the routing still owed.
