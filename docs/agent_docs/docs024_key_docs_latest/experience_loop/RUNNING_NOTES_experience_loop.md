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
