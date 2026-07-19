# HANDOFF → next chat (continues "diagnosis fixloop 3")

*Written 2026-07-18 evening (turn 40). Cold-start bootstrap for the next chat —
read top to bottom, it is self-sufficient. Supersedes
`HANDOFF_diagnosis_fixloop_2.md` as the fresh-chat entry point. Deeper detail:
`SUMMARY_where_we_are_2026-07-18_evening.md` (operational state),
`SUMMARY_the_immune_system_2026-07-18.md` (the journey), and the DESIGN docs
named in §5. A different model may run the next chat — nothing here depends on
one. FIRST, read the repo-root `CLAUDE.md` — it now carries the load-bearing
coordination rules (commit-per-task, build-from-HEAD, both council policies).*

---

## 0. What changed since HANDOFF_2 (the one-paragraph delta)

HANDOFF_2 said "the tool is complete, go point it at real bugs." We did. It
diagnosed its **first real-case CONFIRMED** (BUG A — silent max_tokens
truncation), **found a second bug by failing** (BUG B — root ai_service shadows
step), and **grew two capabilities from two honest refusals**: agent-state
autogather and the **code-lookup verify tier**. The council widened from 2 seats
to **13** (concept-register stage-3, other threads) with a relevance
panel-selector. We **migrated both councils to Claude Sonnet 5**, fixed the
roster-mirror's drift blind spot, and finished the **F1.2** per-run base-branch
cleanup. Audits then found real correctness bugs in the council plumbing that are
still open (§3).

## 1. The immediate next actions (why you're here)

Pick from these — all are small because the hard parts already exist. Owner
gives the go per item; runs spend credits.

1. ~~**Build the 016-finding-2 reviser fix.**~~ **DONE — see the correction
   below. Do not start here.**

   > **CORRECTED 2026-07-19 (turn 41):** this item said "decision already made,
   > not built". It was already built on fix-proposer, and the claim was checked
   > against the live `agent_definitions` rows, not inferred:
   > - **fix-proposer** — `load_council_reviews` present and wired
   >   `council_decide → load_council_reviews → check_approved`, so both reviser
   >   paths inherit it. Built before this handoff was written.
   > - **council-gate** — 13 seats but **no reviser loop at all**
   >   (`complete_revise` is terminal; objections go back to the human
   >   submitter). The bug class does not apply. The "mirror to the gate via
   >   `099`" instruction above was a step that never needed taking — noted so
   >   the next thread doesn't re-derive it.
   > - **feature-designer** — `PATCH_017` had fixed the REVISE path only. The
   >   VETO path (`reframe`) still named `review_editquality` + `review_guardian`
   >   in `input_fields`: **2 of 5 seats**, blind to bug_historian, guidelines
   >   and reuse_agent. Closed by `PATCH_018_reframe_reads_artifact.sql`, which
   >   mirrors fix-proposer's placement (load step ahead of the routers) rather
   >   than adding a second query step. Verified live: both revisers render
   >   `{{.council_report_row.body}}`, zero residual `review_*` refs, 23/23
   >   steps reachable.
   >
   > The transferable lesson: a fix that covers **one branch of a two-branch
   > router** reads as done and is not. PATCH_017's own header asserted the
   > designer was "complete (5/5/5)" — true of the path it touched. Details:
   > `NOTES_running_fixloop(10).md` turn 41.

2. **Build the diagnosis-side code tier (new capability, planned).** Give the
   DIAGNOSER the code-search the council already has, by REUSING the
   `diagnose_code_lookup` action via a new verdict `code_requests` field. Full
   plan + how it fits the evidence-tool family:
   `DESIGN_diagnosis_side_code_tier.md`. Closes "is this cause elsewhere?" on the
   diagnosis side — the same class of question the council tier was built for.

3. **Prove the 016-finding-1 `.result}}` render fix.** It landed on `fix-proposer`
   at 13:15:11Z (UTC) but no fix-proposer repropose has STARTED since, so it is
   unproven. **TIMESTAMP TRAP:** a repropose's step time can look post-fix while
   its orchestration started pre-fix — join `llm_call_log` to
   `orchestration_states.created_at` and test the RUN START. The first
   fix-proposer repropose whose orchestration starts after 13:15:11Z, with no
   `<no value>` in its rendered prompt, is the proof. (Correction on record: the
   D1 proof run `00a20123` started 13:11:13Z — pre-fix — and DID render
   `<no value>`; its truncation result still stands, its council-injection did
   not.)

4. **BUG A's approved fix → the implementer.** `/bugs_open/008` has a
   council-APPROVED plan on correlation `e505f70f` (covers both provider adapters).
   It can go to the implementer (092 → build gate → PR) — but that opens a real PR
   and is the 008 fixing thread's call; coordinate.

## 2. What the whole thing IS (one paragraph)

A three-tier self-healing system. Tier 1 (build workflows) does the work. Tier 2
(the immune system: triage + silent-check) detects problems and routes genuine
code bugs to tier 3. Tier 3 (the fix loop) diagnoses with citations (cite or
abstain), turns a CONFIRMED diagnosis into a constrained edit plan, has a
reviewer COUNCIL argue it, implements in a caged pod (writes only via the
git-adapter, holds no token), gates on build, and opens a PR for a human to
merge — nothing merges itself. The load-bearing property: **it is trustworthy
because it REFUSES** — no confirm without evidence, no confirm on code alone
without observing the mechanism, no blessing a partial class-fix.

## 3. Live now (shipped + verified)

- **Both councils on `claude-sonnet-5`, reviewers `max_tokens: 8000`** (D1 —
  proven non-truncating; the raise was necessary, two reviewers wrote >3000).
  Backups: `bak_agentdef_fixproposer_sonnet5_20260718`,
  `bak_agentdef_councilgate_sonnet5_20260718`.
- **`099_SYNC_gate_roster.py` now detects config-value drift** (deep JSON
  compare), not just seat-name drift.
- **F1.2 done** — implementer base branch is a per-run input (`input_data.base_branch`,
  default main via 092's `BASE_BRANCH`). Stale `084` literal gone from all three
  spots. `read_current_files` + `create_branch` fully per-run NOW; `prepare`
  completes when the `base_branch_field` Go change ships (next image; safe `main`
  fallback until then). Backup `bak_agentdef_fiximpl_F1_2_20260718`.
- **Code-lookup verify tier (F2.3b(c))** live, with Go-receiver-aware symbol match
  + dedup; proven converting the historian's escalation to an approval.
- **Agent-state autogather** in the diagnosis bundle for config-shaped bugs.
- **Council: 13 seats + a relevance panel-selector** (`select_review_panel`,
  keyword footprints; empty footprint = fail-open/always-run). Other threads own
  the roster; use `099 --apply` to mirror fix-proposer→gate, never hand-patch.
- **CLAUDE.md** carries both policies: council-on-fix (advisory gate) and
  diagnosis-before-debug (opt-in by judgement).

## 4. Open correctness issues (confirmed live) + tensions

- **bugs_open/016 finding 1** — `.result}}` render fix UNPROVEN (see §1.3).
- ~~**bugs_open/016 finding 2** — reviser half-blind.~~ **CLOSED 2026-07-19**
  across all three agents: built on fix-proposer, not applicable to council-gate
  (no reviser loop), and the designer's veto path closed by `PATCH_018`. See the
  corrected §1.1.
- **bugs_open/019 vs D1** — the council gate VOIDS a round if a reviewer overruns
  8000 tokens; substantial submissions push seats toward it, and D1 set the
  ceiling AT 8000. Open question: raise the ceiling, or change 019's
  void-on-overrun? Flagged to the gate thread.
- **The Go half of F1.2** (`base_branch_field`) is committed but inert until the
  next chassis image — verify then: `strings /app/agent-chassis | grep -c
  base_branch_field`.

## 5. Companion docs (read as needed)
- `SUMMARY_where_we_are_2026-07-18_evening.md` — operational state (start here).
- `SUMMARY_the_immune_system_2026-07-18.md` — the journey / standalone overview.
- `DESIGN_diagnosis_side_code_tier.md` — §1-5 the code tier plan, §6 the 016-f2 fix.
- `DESIGN_feature_builder_and_council_gate.md` — the two forward threads' designs.
- `READOUT_problems_we_faced_2026-07-18.md` — plain-language problems read-out.
- `RUNBOOK_diagnosis_fix_loop(10).md`, `NOTES_running_fixloop(10).md` — task +
  turn-by-turn (turns 34-40 this thread).
- `/bugs_open/` — the real-case queue (001-018+, several loop-relevant: 003 spawn
  loss + reaper blind spot, 008/009 BUG A/B, 010 intrinsic overflow, 013
  implementer-skips-gofmt, 017 unregistered-action-marked-complete).

## 6. Gotchas that cost hours — do not relearn
- **The clock trap: DB is UTC, the dev host is BST (+1).** A poll printing
  `date +%H:%M:%S` shows BST; `orchestration_states.created_at` is UTC. This
  masked that the D1 run was pre-fix. Compare like with like.
- **Config re-seeds clobber concurrent config work** (like `git add -A` clobbers
  WIP). fix-proposer was re-seeded ~5× by other threads in 18h. Use PATCH-STYLE
  idempotent seeds, never whole-object writes; back up first
  (`bak_agentdef_<type>_<date>`). Finding:
  `../multi_session_coordination/FINDING_2026-07-17_config_reseed_clobber.md`.
- **`git commit <explicit paths>`** (pathspec on COMMIT, not add) — the shared
  index is full of other threads' staged files; a bare `commit -m` sweeps them in.
- **Same-tag deploy ships stale binary; verify the POD binary, never git/tag.**
  Build from committed HEAD (`make build-agent-chassis-ref`); image FIRST then
  seed; no orchestration within ~300s of a chassis (re)start (spawn dropped).
- **Round-count inflation:** accumulated `council_report` rows on a
  fix_correlation shorten the loop — clear them for a fair-round run (back up
  first). Round counting is orchestration-scoped in source but the accumulated
  rows still bit us; verify per run.
- **Model gotchas (Sonnet 5):** omitting `thinking` runs ADAPTIVE (spends from
  max_tokens); size max_tokens for thinking+output; tokenizer ~30% heavier.
- **BUG B, live rule:** ROOT `ai_service` SHADOWS the step's (ai_actions.go reads
  root first). The old runbook gotcha was INVERTED. fix-proposer has NO root
  block so its step config takes effect; check before editing any agent's model.

## 7. Key triggers, queries, DB
- Triggers (kcat envelope): `090_…needs_diagnosis` (has the coverage check +
  FORCE=1), `091_…fix_proposer`, `092_…fix_implementer` (→ orchestrator; now
  takes `BASE_BRANCH`), `093_…fixloop_digest`, `095_…diagnosis_triage`,
  `097_…council_review` (the gate), `099_SYNC_gate_roster.py` (roster mirror).
- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
- Artifacts: `diagnosis_artifacts` (kind ∈ bundle|fix_plan|council_report|
  escalation, by correlation_id/iteration); `orchestration_states.collected_data`
  (verdict, route, review_*); `llm_call_log` (prompt_rendered, response_text).
- The benchmark: darts `guides-index` blank page, correlation `e08c5b01` — still
  live, escalates by design (fixing it retires the benchmark).

## 8. Operating posture
MANUAL everything; nothing auto-dispatches; each run spends credits, owner says
go per item. Correctness rests on gates + deterministic routing + human
decisions, so a new model can continue safely from these docs.
