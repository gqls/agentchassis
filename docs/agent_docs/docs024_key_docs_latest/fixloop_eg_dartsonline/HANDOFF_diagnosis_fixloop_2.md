# HANDOFF → next chat (continues "diagnosis fixloop 2")

*Updated 2026-07-16 (turn 33). Cold-start bootstrap for the next chat. Read
this top to bottom; it is self-sufficient. Deeper detail lives in the companion
docs named at the end — start with `SUMMARY_where_we_are_2026-07-16.md` for the
full journey (from → now → going). A different model may run this chat (Fable
credits low) — nothing here depends on a specific model.*

---

## 1. The immediate next action (this is why you're here)

**The tool is COMPLETE.** All four phases of the triage/escalation design are
built, deployed, and proven in production (see §3). The loop can already take a
report, diagnose it with citations, plan a fix, run a two-reviewer council,
implement it in a caged pod, gate it on build, and open a PR — and PR #1 was
merged. The whole immune-system → fix-loop channel is closed and self-reporting
in the digest.

**So the next move is to point the finished tool at REAL bugs.** The owner
opened the real-case queue on 2026-07-16 and chose the first case:

> ### ▶ FIRST REAL CASE — the image-landing data-loss trap
> Full self-sufficient handoff:
> **`docs/agent_docs/docs024_key_docs_latest/aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`**
>
> **What it is:** a *platform* data-loss trap (not one site). Landing an image
> on a page fires a scoped section re-render; on any page whose article-body was
> never unwrapped from its LLM JSON envelope, that re-render renders the body
> empty and **silently overwrites the good HTML with a blank shell** — the
> article vanishes from the live page. It blanked 9 pages across 5 sites; 4 more
> sit JSON-leaking.
>
> **STATUS UPDATE 2026-07-16 — the bleeding is stopped, the wound is not
> dressed.** The guard (`missingRequiredLLMFields` / "escalating page to writer
> instead of blanking", from the empty_sections thread) **is now LIVE in prod
> (`v1.0.1123`)** — verified in the running pod (2 / 1 / 4 on the three
> symbols). So new blanking is prevented and the old "don't land an image"
> operating rule is **lifted**. But **the 13 broken pages are still broken on
> the live sites** — the guard repairs nothing. Recovery is now the top job
> (004 §4.2), and `ParseLLMJSON` still fails on 14 fixtures (004 §4.3), so
> writer-escalation may not cleanly regenerate every page.
>
> **Why it's the right first case:** it is a genuine, high-severity, already
> hand-diagnosed platform bug with a clear code map (004 §7) — exactly the
> shape the loop was built for, and (like the darts benchmark) diagnosable so
> its output can be graded. It also spans two threads' territory (imagery found
> it; empty_sections owns the guard), so it exercises the loop's cross-thread
> awareness.
>
> **How to start it:** DON'T just run the loop blind. First read 004 top to
> bottom, then decide the intake: either (a) hand-write the `needs_diagnosis`
> symptom (090 contract — see §7 of THIS doc) pointing at 004's mechanism and
> code map and let the loop confirm/plan it, or (b) if these pages are
> surfacing as `page_rerender` failures, let triage route them.
>
> **Note the case has SHIFTED since it was filed.** The guard landed, so the
> live question is no longer "stop the blanking" — it's the remaining half:
> **recover the 13 broken pages** (string-surgery out of `content_data.result`;
> the envelope is NOT valid JSON, and some are truncated → only partial
> recovery), **fix `ParseLLMJSON`'s 14 fixtures** (decide repairable vs
> quarantine-the-truncated), and consider the structural hardening — a
> schema-`required` field should never render empty (`missingkey=zero`,
> `call_agent.go:1152`), which is the same class as the product-page defect.
> That last one is the most loop-worthy piece: a real, platform-wide, code-level
> defect. Frame the intake around what's actually left, not the filed headline.

The other queued cases (dispatch order is the owner's call), all in
`aaa_fails_to_mend/`: `001` replan-clobbers-built-pages, `002` errors-to-fix
list, `003` spawn-lost-child-response.

Everything stays **manual** and human-gated. Each diagnosis run spends credits —
the owner says go, per case.

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
- **Phase 3 — feedback close-out: DONE, LIVE 2026-07-15 (v1.0.1122, turn 32).**
  Each triage sweep recomputes failure-pattern keys over ALL failed items (no
  window) and closes any parked escalation whose pattern has vanished
  (`triageCloseResolved`); re-escalation is automatic via the dedup index.
  Re-driving original items after a fix ships stays a HUMAN action. Proven
  both ways: real sweep closed nothing (all 3 patterns still exist); a
  synthetic probe closed itself while the 3 real ones stayed open. Silent-check
  already does the equivalent for its own findings. **The whole
  triage/escalation design (Phases 1–4) is now live.**
- **Phase 4 — digest escalation section: DONE, LIVE 2026-07-15 (v1.0.1120,
  turn 31).** The digest's "Escalation channel" section shows sweep counts,
  the WHOLE open diagnosis queue every digest (NEW-flagged in-window), silent
  findings open/CLOSED, and standing capability gaps; triage + silent-check
  are in the run roster. First delivery: `docs/fixloop_digests/DIGEST_latest.md`
  (2026-07-15).
- **Later — wider council** (guidelines / reuse / bug-historian reviewers — the
  F2 roster in the runbook); **capability-builder** (features from specs — the
  ambitious, human-gated direction).
  **Now informed by the concept register (search-tab2, `docs026_concept_register/`).**
  Its stage 3 — "build council agents per concept area" — IS this
  council-widening track. `FIX-036` in that register is explicitly the
  wider-council-roster vision (flagged "the seam this concept register is meant
  to fill"), and concepts independently rediscovered 4–6× across doc eras
  (e.g. "adoption writes first, classifier consumes"; the wrapper-orchestrator
  pattern) are the strongest signals for which reviewer seats to build FIRST.
  Stage 2 there is complete (1,627 concepts verified, ~7.6% doc-error rate);
  wiring stage-3 council seats into the live fix-loop workflow is a
  cross-workstream production change the owner has reserved for explicit
  sign-off (that register's RUNBOOK B4).
- **The real-case queue (owner, 2026-07-15):**
  `docs/agent_docs/docs024_key_docs_latest/aaa_fails_to_mend/` holds handoffs
  of real errors to diagnose **once we're happy with the tool** (001 replan
  clobbers built pages; 002 errors-to-fix list; 003 spawn lost child
  response). They enter via the loop's normal intake (090 needs_diagnosis
  contract, or as triage patterns where they surface as failures). Owner
  gives the go — each diagnosis run spends credits.

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

- `SUMMARY_where_we_are_2026-07-16.md` — the workstream JOURNEY (from → now →
  going); read this first for the full arc + the two forward tracks.
- `SUMMARY_where_we_are_2026-07-14.md` — gentle plain-language state (for
  humans; the read-aloud version — covers triage + silent-check going live).
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
