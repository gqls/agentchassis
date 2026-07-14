# HANDOFF → "diagnosis fixloop 2" chat

*2026-07-14. Cold-start bootstrap for the next chat continuing the diagnosis→fix
loop. Read this top to bottom; it is self-sufficient. Deeper detail lives in the
companion docs named at the end. A different model may be running this chat
(Fable credits are low) — nothing here depends on a specific model.*

---

## 1. The immediate next action (this is why you're here)

> **DONE 2026-07-14 (chat "diagnosis fixloop 2", turn 29).** v1.0.1117 verified
> in the pod (triageRoute grep=2); dry-run confirmed the filter (only real
> code bugs in the loop group; timeouts → re-queue; no-signal → hold);
> `dry_run` flipped to false; live sweep escalated 2 patterns
> (`triage-diag:needs_new_component:c4ad0be8a0f2`,
> `triage-diag:needs_component_regeneration:171f7b9c1d60`), parked at
> `awaiting_diagnosis` (inert); a third sweep proved dedup (deduped 2, wrote
> nothing). Known cosmetic defect: dry-run counters mislabel would-be
> escalations as "capped" (fixed in v1.0.1118). **TRIAGE IS LIVE — the
> tier-2→tier-3 channel is closed.** Same day (turn 30): **Phase 2 built and
> LIVE too** (v1.0.1118, `diagnose_silent_check` — see §4/§5). Next = Phase 3
> (feedback close-out) / Phase 4 (digest escalation section); dispatching the
> parked escalations into the loop is the owner's call. The steps below are
> kept for the record.

A new chassis image is shipping (was building at handoff; expected **v1.0.1117**
or later). The moment it's live, finish bringing **triage** online:

1. **Verify the image in the pod** (never trust the tag — same-tag deploys ship
   stale binaries):
   ```
   P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
   kubectl -n ai-persona-system exec ${P#pod/} -- sh -c "grep -ac triageRoute /proc/1/exe"
   # must be >= 1 (triageRoute = the loop-worthiness filter, this build's new code)
   ```
2. **Settle 300s** after the pod (re)started (rebalance-window gotcha), then
   **re-run the triage dry-run**:
   ```
   ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/095_TRIGGER_diagnosis_triage_v1.sh
   ```
3. **Read the report** and confirm the filter works — the "Code bugs → fix loop"
   group should now contain ONLY real handler errors; the claim-timeouts and
   no-signal patterns should have moved to "re-queue" / "hold":
   ```
   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db \
     -t -A -c "SELECT body FROM doc_notes WHERE categories ? 'triage' ORDER BY created_at DESC LIMIT 1;"
   ```
4. **If the preview looks right, flip to live** (one line) and fire again to
   escalate for real:
   ```
   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
     UPDATE agent_definitions SET default_config = jsonb_set(default_config,
       '{workflow,steps,sweep,config,dry_run}', 'false'::jsonb), updated_at=now()
     WHERE type='diagnosis-triage' AND is_active AND COALESCE(is_snapshot,false)=false;"
   ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/095_TRIGGER_diagnosis_triage_v1.sh
   ```
   Then confirm needs_diagnosis items were written:
   ```
   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
     SELECT item_key, summary FROM site_work_items WHERE source='diagnosis-triage' ORDER BY created_at DESC LIMIT 5;"
   ```

That closes the escalation channel: the operational immune system → the fix loop.
Everything stays **manual** (owner: Fable credits low — no auto-cadence).

## 2. What the whole thing IS (one paragraph)

A three-tier self-healing system. **Tier 1** (build workflows) does the work.
**Tier 2** (checkers + handlers — the immune system) detects problems and
applies known remedies. **Tier 3** (the fix loop) is for problems whose cause is
in the CODE: it diagnoses with citations, plans a constrained fix, has a
two-reviewer council approve/revise/reject it (with self-verification and honest
escalation), implements it in a caged dedicated pod (writes only via the
git-adapter, holds no token), gates it on `gofmt`+`go build` in a container, and
opens a **pull request for a human to merge — nothing merges itself**. The new
**triage** router is the bridge from tier 2 to tier 3: a handler *failing* (not a
checker detecting) is what feeds the loop, and only genuine code bugs get
through.

## 3. What is BUILT and LIVE (production)

- **Diagnosis loop** (F0): gated CONFIRMED with citations, durable bundles.
- **Fix-proposer + council** (F1.1a, F2.1–F2.3b): constrained plans; 2-reviewer
  council; decision router (approved | revise→verify→repropose |
  first-veto→reframe-once | rejected/exhausted→escalate); verify step (reviewers'
  read-only SQL checks, run under containment); live schema hint; escalation
  artifact. Workflow v5. **max_rounds=3.**
- **Write step** (F1.1b(c)): `fix-implementer` (dedicated pod via
  `fix-implementer-orchestrator`; reads via GitHub contents API with a read
  token from the spawn gate; writes via the git-adapter) → hard file allowlist
  (`diagnose_prepare_fix_commit`) → branch + commit → **build gate**
  (`diagnose_build_gate`, golang k8s Job) → PR (`git_adapter_request`). git-
  adapter gained create_branch / create_pull_request / branch-commit.
  **PROVEN: opened and the owner MERGED PR #1** (a real one-file defect).
- **Awareness digest** (`fixloop_digest`): deterministic (no LLM) daily-ish
  account of runs, council decisions, gate/PR outcomes, and agent-config
  snapshots. Delivered as a committed file under `docs/fixloop_digests/`
  (`094_pull_digest_to_file.sh` bridges doc_notes → the file).
- **`main` is fixed**: PR #1's fix (stranded on 084 by PR ordering) was
  cherry-picked onto main (commit 998c0b31).

## 4. What is BUILT and — since 2026-07-14 — LIVE (v1.0.1117 triage; v1.0.1118 adds silent-check)

- **`diagnose_silent_check`** (Phase 2, v1.0.1118): the verification checker
  for silent failures — structural invariants violated with NO covering work
  item. Emits inert `silent_failure` items for triage to route; closes them
  when the violation (or its silence) ends. Live, `dry_run` flipped false;
  manual trigger `096_TRIGGER_diagnosis_silent_check_v1.sh`. §5 Phase 2 has
  the detail.

- **`diagnose_triage`** — the router. Scans `site_work_items`:
  - LOUD failures (`status='failed'`) → **loop-worthiness filter** (`triageRoute`)
    → only real code-bug patterns escalate to `needs_diagnosis` (deduped by
    (item_type, handler, error signature); capped at `max_escalations=3`;
    ON CONFLICT dedup on a stable item_key; parked at `awaiting_diagnosis`,
    INERT). Transient/infra (claim timeout, dead pod) → re-queue (surfaced, not
    escalated); no error text → hold (human).
  - CAPABILITY GAPS (`capability_gap` / `deferred`) → surfaced to the roadmap,
    NEVER escalated (a missing handler is a capability decision, not a bug —
    already emitted by `load_work_item_actions.go:245-280`).
  - One doc_note per sweep (categories `triage+fixloop`) — the readable artifact.
  - **Now runs `dry_run=false`** (flipped live 2026-07-14 after the dry-run
    preview verified the routing; the seed still ships `dry_run=true`, correct
    for any re-seed).
- Seed `0NN_diagnosis_triage.sql`, trigger `095_TRIGGER_diagnosis_triage_v1.sh`.

## 5. The roadmap after triage goes live (design in DESIGN_triage_and_escalation.md)

- **Phase 2 — silent-failure verification checker: DONE, LIVE 2026-07-14
  (v1.0.1118, turn 30).** Scope was narrowed by the empty-sections thread's
  reconciliation (DESIGN §silent-failure): their completion gate de-silences
  registered item_types, recurrence rides `insertWorkItem`'s two-strike rule,
  so the checker owns only defects **no work item ever touches**.
  `diagnose_silent_check` (deterministic, no LLM): `nav_linked_never_built`
  emits (the darts class — found it on 2 sites), `deployed_zero_components`
  report-only pending owner review (`emit_checks` promotes it). Findings =
  INERT `silent_failure` items, ONE platform pattern via a ≥140-char error
  prefix; proven end to end (checker → triage → needs_diagnosis
  `triage-diag:silent_failure:fd86fec2c4da`) including live dedup and honest
  close-out. Seed `0NN_diagnosis_silent_check.sql`, trigger `096_…`, notes
  turn 30.
- **Phase 3 — feedback close-out.** After a fix deploys, re-verify the original
  items; fixed → close the escalation, still failing → back to triage.
- **Phase 4 — digest escalation section.** Fold escalations + capability gaps +
  verifications into the awareness digest, so the owner sees the whole immune
  system on one page.
- **Later — wider council** (guidelines / reuse / bug-historian reviewers — the
  F2 roster in the runbook); **capability-builder** (features from specs — the
  ambitious, human-gated direction).

## 6. Gotchas that cost hours — do not relearn

- **Same-tag deploys ship stale binaries.** Bump IMAGE_TAG (makefile line 16)
  and verify strings in the POD binary (`grep -ac <symbol> /proc/1/exe`), never
  the tag, never git.
- **Deploy order:** image FIRST, then the seed. A seed naming an unregistered
  action fails at runtime.
- **Rebalance window:** never fire an orchestration within ~300s of a chassis
  pod (re)start — the spawn is silently dropped.
- **max_tokens** lives INSIDE a step's `ai_service` block; root is dead config
  (client defaults to 2048, truncates JSON).
- **fix-implementer MUST fire via `fix-implementer-orchestrator`** (the 092
  trigger targets it) — fired directly it runs in-chassis and gets no read token.
- **A stale `fix/*` branch must be deleted before re-firing** the implementer
  (create_branch is idempotent and reuses the old base).
- **implementer ref/base are live-set to `084_site_improvements_local_ai`** (the
  active branch; origin/main was stale). Making ref a per-run INPUT is the open
  F1.2 cleanup.
- **Git: forward-only, no resets/amends.** Many concurrent sessions commit to the
  same branch; your changes may land in another session's commit — that's fine,
  nothing is lost. Check `git log` before assuming your commit is HEAD.
- **snapshot_agent** writes to `agent_definitions_backup`
  (`snapshot_taken_at` / `snapshot_reason`).
- **BST vs UTC:** dev host +0100, DB UTC; `orchestration_states.last_activity`
  is timestamp WITHOUT tz.
- **needs_diagnosis contract:** anchors to `system.internal`
  (`eac60db8-b032-432b-b36d-76f37632045d`); the real site travels in `spec`;
  parks at `awaiting_diagnosis`; dedup on (site_id, item_key) where status not
  terminal.

## 7. Key files, triggers, queries

- Triggers (all replay the proven kcat envelope): `090_…needs_diagnosis`,
  `091_…fix_proposer`, `092_…fix_implementer` (→ orchestrator),
  `093_…fixloop_digest`, `095_…diagnosis_triage`.
- Seeds: `0NN_fix_proposer.sql` (v5), `0NN_fix_implementer.sql`,
  `0NN_fix_implementer_orchestrator.sql`, `0NN_fixloop_digest.sql`,
  `0NN_diagnosis_triage.sql`.
- DB: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.
- Artifacts: `diagnosis_artifacts` (kinds bundle | fix_plan | council_report |
  escalation) by correlation_id; `site_work_items` (failures, capability_gap,
  needs_diagnosis); `doc_notes` (categories digest / triage).
- The benchmark: darts `guides-index` blank-page bug — correlation
  `e08c5b01-01ef-42ad-80d0-b77c50ec9e84`. STILL LIVE and unfixed by design (the
  council correctly judges its true fix architecture-level → escalates). Fixing
  it by hand retires the benchmark.

## 8. Companion docs

- `SUMMARY_where_we_are_2026-07-13.md` — gentle plain-language state (for humans).
- `DESIGN_triage_and_escalation.md` — the triage + escalation architecture (three
  flavours, routing, phasing, decisions).
- `RUNBOOK_diagnosis_fix_loop(10).md` — task, phases, every gotcha.
- `PLAN_fixloop_pilot.md` — what's built / next.
- `NOTES_running_fixloop(10).md` — turn-by-turn evidence trail.
- `HANDOFF_CURRENT_fixloop.md` — the prior living handoff (this doc supersedes it
  as the fresh-chat entry point).
- `docs/fixloop_digests/` — the owner's committed-file awareness surface.

## 9. Operating posture

MANUAL everything (owner, Fable credits low): nothing auto-dispatches, no
scheduled cadence. The diagnose-dispatch-loop is shipped-disabled; triage ships
dry-run; the digest is manual-trigger. The whole design is gates + deterministic
routing + human decisions, so correctness does not depend on any one model — a
new model can continue safely from these docs.
