# HANDOFF — Experience Loop, resume here

*2026-07-19. Supersedes HANDOFF_2026-07-17_experience_loop_start.md (that one
was the bootstrap; this is the live position). Read top to bottom, then start
at §5. Companions: PLAN_experience_loop.md (why), RUNBOOK_experience_loop.md
(how — §8a is the task-state table), RUNNING_NOTES_experience_loop.md (the log).*

---

## 0. What this subproject is

Every check the platform had verified a page or a tool **in isolation**. Nothing
owned the **experience**: the promise a button makes, the journey a visitor
takes across pages, the data a widget needs, the honesty of the numbers on the
page. The link-integrity loop can prove a button reaches a real page; it cannot
notice the page is a mock.

The Experience Loop adds that layer: a machine-written EXPERIENCE_PLAN (journeys,
promise ledger, data contracts, MVP cut, acceptance criteria), attacked by a
council of critics until it converges, then built contract-first and verified by
the existing acceptance ladder. Pilot: the vonc.com Spark game.

## 1. Bootstrap facts (verified 2026-07-18/19)

- **Site**: vonc.com, `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`.
- **DB**: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
- **Chassis**: `v1.0.1135`, single replica. Guard-rail Go + the
  docResolveSubject `experience` fix are IN it (verified in-pod by binary
  string, never by tag).
- **Agent**: `experience-planner` v1 active. Seed
  `sql_for_agents/167_experience_planner_and_council.sql` (re-apply is
  idempotent, `ON CONFLICT DO UPDATE`, snapshots first).
- **Trigger**: `sql_for_agents/092_TRIGGER_experience_plan.sh vonc.com vonc-spark-game "the Spark daily-provocation game"`
- **Migrations applied + ledgered**: 163 (experience subject_type), 164
  (pages.rebuild_policy), 165 (dead_controls check), 166 (vonc evidence_base),
  167 (the agent). Next free number: **168** — re-check at execution time.

## 2. Done and proven (do NOT redo)

**Phase 1 + Phase 2 complete; CP1 proven live on vonc.**

| Guard rail | State | Proof |
|---|---|---|
| Page-ownership marker (`pages.rebuild_policy`) | live, 38 pages `owned` | scoped reconcile emitted `owned_page_review` for provocation/gauntlet/quiz and **zero** `needs_page` for owned pages — the manual park step is retired |
| Rename re-keys travelling docs | live | `RekeyTravellingDocs` + `rename_tool_identity` action; `create_tool_component` now uses `CanonicalisePage` (kills TL-003 drift at birth) |
| Dead-control check | live | `dead_controls` discovery check fired and caught two real dead CTAs on the vonc index |
| Claims lane for vonc | live | `evidence_base` seeded; `claimscan` baseline found **14** findings incl. 3 nobody had catalogued (`14,203 Happy Customers`, `10K+ Players Scored`, mangled about stats) |

**The council machinery works end to end**: compose → persist to `doc_plans`
(`subject_type='experience'`) → 4 critics → deterministic `diagnose_council_decide`
→ router → `run_checks` → recompose, superseding the plan each round, with the
round trail in `doc_notes` (category `experience-council`) + `diagnosis_artifacts`
(`kind='council_report'`).

**Owner decisions in force**: D1 gauntlet = minimal-real playable round. D2 =
**option B**, per-provocation detail renders CLIENT-SIDE on the existing
`/provocations/index.html` runtime-fill shell (no new page_type, no new render
path; static pages + daily emitter are LATER). D3 = autonomous pilot on vonc.

## 3. Where it actually stands — CP2 NOT yet reached

Six council runs. Every escalation was **correct**, and each exposed a defect in
the harness rather than in the plan. All four are fixed:

1. `{{.step.result}}` in prompt templates — `ExtractFields`→`UnwrapDeep` strips
   the `{result,type}` wrapper, so template refs must be bare `{{.step}}` while
   **config** dot-paths keep `.result`. (Became fleet-wide `bugs_open/016`.)
2. Critic `max_tokens` 4000→8000 — a truncated critic JSON made
   `council_decide` fail closed (correctly).
3. **`load_context` was lying by omission** — it filtered
   `component_level='tool'`, hiding `gauntlet-interface`, `gauntlet-cta`,
   `provocation-card`, `lobby-grid` (all `level='section'`). The planner
   asserted components existed; the critic could see no evidence and objected —
   **correctly, five rounds running, across two independent runs.** Now surfaces
   per-page attachments with no level filter, labelled COMPLETE ground truth.
4. **Compose truncation death spiral** — `llm_call_log` showed recompose output
   13303→12599→14138→15499→**16000/16000**. The plan grew each round absorbing
   objections until it truncated *inside the §5 criteria fence*, which produced
   objections that revising could never clear (revising makes it longer). Fixed
   at both ends: 16000→32000 **and** a LENGTH DISCIPLINE rule (tighten and
   replace, never append; criteria fence has absolute priority).

**Both fixes are confirmed working** by the last run (`054b358a`):
- the persisted plan is **complete** — 13,578 B, closed ```criteria fence,
  `<!-- END EXPERIENCE_PLAN -->` present (was 26,522 B and truncated);
- feasibility's objections **changed class** — no more "cannot verify the
  component exists"; now substantive (who authors the feed fields; deactivated
  site components).

**Why that run still didn't converge**: `review_feasibility` died on
`no text content in response (had 1 blocks)` — the workflow routed to
`error_step` and the run terminated at `complete_refused` after round 1. That
is the **known** bug `bugs_open/008` item 5 (undecoded `stop_reason`, likely
`refusal`, on Sonnet 5), which was filed as *optional* and left undone. I
appended real-case evidence to 008; the fix belongs to the fixloop thread.

## 4. Two live findings the council surfaced that are NOT experience-loop work

1. **Site components deactivated fleet-wide on vonc** — feasibility reported
   `header-bold-gradient`, `footer-4-column` and `Document Head` inactive across
   16 pages including `provocations-index` and the tool pages. Not investigated.
   Verify before building anything on those pages.
2. **`bugs_open/016`** (mine, now circulated): council revise prompts dropped
   every reviewer's output. `fix-proposer` and `feature-designer` fixed by their
   threads; `content-creator-hero` is the last `.result}}` in the fleet and is
   affected — its hero prompt renders `<no value>` where the researcher's
   findings belong, so heroes have been written without the research they
   commissioned. Left for its owning thread.

## 5. START HERE — next actions in order

1. **Re-fire the council.** Nothing is in flight (check first — the Kafka topic
   lag has exceeded 10 minutes, and that is how a double-fire happened once):
   ```sql
   SELECT count(*) FROM orchestration_states
   WHERE owner_agent_type='experience-planner' AND status NOT IN ('COMPLETED','FAILED');
   ```
   then run the 092 trigger. Judge by the `experience-planner` row, not the
   generic wrapper. A round is ~5 min; up to 5 rounds.
2. **If it dies again on `no text content`**: that is 008, not the loop. Either
   wait for the fixloop thread's fix, or make the council resilient by pointing
   the critic steps' `error_step` at `council_decide` instead of
   `complete_refused` — `diagnose_council_decide` already treats an absent
   reviewer as an **abstention**, so a single flaky critic would no longer
   destroy a whole run. (Recommended: it is a two-line config change and it
   makes every future round robust.)
3. **If it converges** → CP2 closed. Then RUNBOOK Phase 4 (T4): MVP build round
   — feed to spec, arena re-attach via `rename_tool_identity` then let the
   acceptance ladder drive the rebuild, client-side provocation detail (D2=B),
   gauntlet minimal-real, `claimscan` exit 0.
4. **Phase 5 (T5)** is coordinated with the travelling-docs/tool-acceptance
   thread. Their binding ruling is in RUNBOOK §T5: journeys must be an
   **ADDITIVE** persistent-context path (never rework `Execute`), reuse
   `evaluateOnPage` for free `forced_by` overflow attribution, navigate by
   **symbol not line**, and unify `needs_experience_replan` escalation with
   `bugs_open/010` candidate (b) rather than building it twice.

## 6. Landmines

- **Check for an in-flight run before every fire** (topic lag > 10 min).
- **Verify deployed code in-pod by binary string**, never by git or tag.
- The chassis is a **single replica** and went quiet fleet-wide for ~15 min on
  07-18. Do not restart it to unstick your own run — it disrupts every
  concurrent thread; a healthy pod with no `AWAITING_RESPONSES` backlog is lag,
  not failure.
- **The current `is_current` plan is an escalated/unapproved one. Do not build
  from it** until a run converges.
- Migration numbers collide across sessions — claim the next free one at
  execution time.
- Runtime-fill shells (`provocation-card`, `lobby-grid`) render blank/`#` before
  hydration by design; the dead-control and phantom-link checks already exempt
  them.
- The three `misdirected_cta` `unresolved` triplets on vonc are expected state.
