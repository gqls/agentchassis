# RUNBOOK — Council Gate (advisory review service)

**Thread:** "fixloop council on every bugfix" (started 2026-07-17 from
`HANDOFF_2026-07-17_council_gate_thread.md`). Design:
`DESIGN_feature_builder_and_council_gate.md` §2. Turn record:
`NOTES_running_council_gate.md`.

## State: LIVE (advisory) — applied 2026-07-17 evening

Applied to `clients_db` (`postgres-clients-0` / `ai-persona-system`) with the
owner's explicit named go. Pre-flight read the LIVE fix-proposer roster from
the DB first (the owner warned seats are being added frequently) and found it
had grown to **7 seats** (adoption-guardian + a code_lookup step since the
v11 files) — mirrored the seventh seat and its gate/footprint from the live
row before applying; `code_lookup` deliberately not mirrored (reproposer-side
machinery; recorded in the seed header). Post-apply verification green: row
active, 27 steps, seven-way review_fields/check_fields, five gated
footprints, all seven prompts intact. First smoke submission fired
(correlation `bd12762a-5b10-416b-a70d-90ee3067ce7d` — a genuine change: the
digest gate-verdicts section the handoff names as the channel to extend).

| # | Component | File | State |
|---|---|---|---|
| 1 | Submission wrapper + trigger | `097_TRIGGER_council_review_v1.sh` | built; validations dry-run tested (single-line payload proven — kcat trap) |
| 2 | Orchestrator seed | `0NN_council_gate.sql` | **APPLIED & LIVE**; roster now maintained by `099_SYNC_gate_roster.py`, not by editing this file. **16 seats as of 2026-07-20**, zero drift vs `fix-proposer`; chassis v1.0.1140 pod-verified (carries the 036/019 council fixes) |
| 3 | Visibility report | `098_REPORT_unreviewed_commits_v1.sh` | built; live-run 2026-07-17: 28 in-scope commits / 3 days, 0 reviewed |
| 4 | PR-mode (enforcement) | — | **not built** — owner's explicit go required (build order rule) |

## Owner decisions on record (collected 2026-07-17, this thread)

1. **Scope:** `platform/`, `internal/`, `pkg/`. Docs/site content never spend
   council credits (the 097 script refuses them client-side; `FORCE=1` overrides).
2. **Mode at launch:** advisory first (steps 1–3). PR-mode stays a later,
   separate owner call.
3. **Credit policy:** one council run per submission = per task/commit,
   matching the commit-per-task rule.
4. **Roster:** WAIT for more concept-register stage-3 seats before launch —
   and subsequently (owner, later the same day, via the concept-register
   thread): **the relevance filter next**. Both conditions are now MET: the
   live council is 6 reviewers (v10 tooling-provenance) with the relevance
   filter wired (v11) on image v1.0.1133 (fleet release, pod-verified), and
   this gate's seed mirrors all of it. **The launch precondition is
   satisfied; the remaining step is the owner's named go to apply the seed.**

**Flag → RESOLVED same day:** the v6-inherited `run_checks.check_fields`
omission (three advisory seats' checks solicited but never run) was flagged
by this thread and fixed by the concept-register thread as **v9, applied
live**. The gate seed matches. Their relevance-filter Go engine
(`select_review_panel` + council_decide abstention) is also now built and
committed (`37468ba65`) but **inert until a chassis image ships** — they are
holding that deploy to sequence with this thread; the gate seed notes the
coming lockstep change in its header.

## Launch checklist (when the owner says go)

0. ~~Roster/filter precondition~~ **MET 2026-07-17**: 6 seats live, filter
   wired (v11), image v1.0.1133 pod-verified. Remaining: the named go.
1. Roster lockstep: any seat or filter change lands in **both**
   `0NN_fix_proposer` (v12+) **and** `0NN_council_gate.sql` in the same
   migration — the two files' reviewer steps are deliberately name-matched;
   letting them drift is the dedup-index/Go-list class of failure. (Synced
   to v11 as of 2026-07-17.)
2. Apply `0NN_council_gate.sql` to clients_db
   (`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`)
   — needs the owner to name the target, per the standing permission gate.
   No image build: every action is registered and live (≥ v1.0.1127).
3. Run the post-apply verification queries at the bottom of the seed file
   (17 steps; `review_fields` lists all three reviewers).
4. Smoke-run: submit a small real change via 097, watch
   `orchestration_state_audit`, read the verdict note
   (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`).
5. Announce in the digest/docs: threads submit before committing; approved
   commits carry the trailer.

## How a thread uses the gate (once live)

1. Write a submission JSON (schema in the 097 header: `rationale`,
   `submitter`, and a fix_plan-shaped `plan` — ≤8 edits, ≤64KB, real diff
   hunks in `sketch`, evidence quotes in `grounded_in`).
2. `./097_TRIGGER_council_review_v1.sh submission.json` → save the printed
   `SUBMISSION_CORR`.
3. Verdicts: **APPROVED** → commit with trailer `Council-Reviewed: <corr>`.
   **REVISE** → objections + the reviewers' checks come back answered
   (doc_note + council_report); revise, resubmit with
   `RESUBMIT_CORR=<corr>` so the trail accumulates. **REJECTED** (guardian
   veto) → do not ship as-is; the guardian's notes name the safest contained
   alternative.
4. `./098_REPORT_unreviewed_commits_v1.sh [days]` shows fleet coverage;
   `PERSIST=1` files it to doc_notes (categories digest+council-gate).

**Strengthen-advisory machinery (owner ruling 2026-07-24: strengthen advisory,
defer PR-mode).** Approval is reachable now (~80% since the `bugs_closed/057` severity
fix, v1.0.1149; case CLOSED 2026-07-24), so the norm is reinforced without enforcement:
- **Commit-time nudge (loud):** `.githooks/commit-msg` runs
  `scripts/council-coverage-nudge.sh` — an ADVISORY note when a commit touches
  `platform/|internal/|pkg/` code with no `Council-Reviewed:` trailer. Never blocks;
  silent on reviewed or docs-only commits. It runs AFTER the D2 direction gate
  (that gate still blocks — do not weaken it when editing the hook).
- **Coverage baseline (regular):** run `PERSIST=1 ./098_… [days]` at the digest
  cadence to file a coverage snapshot. A fully-scheduled run needs a runner with
  BOTH git and cluster (kubectl) access — a cloud cron agent likely lacks the
  latter; that automation is an owner call.

### Reading a verdict: check `unreadable`, NOT `abstained` (contributed 2026-07-25)

**`abstained` on this council is not a health signal, and treating it as one will
make every honest approval look broken.** It counts seats the **relevance filter
skipped** — deliberately, as "not applicable" information. `diagnose_council_decide_action.go:284-307`
says so in as many words, and keeps `unreadable` separate on purpose: *"an
abstention is a seat the relevance filter skipped … an unreadable seat is an
opinion we were owed and lost … conflating them would let a lost opinion read as
a considered non-objection."*

Confirmed against the last 8 gate rounds: **`reviewers + abstained == 16` in every
one** (16 = the live seat count), and in 7 of the 8 *every* review present carried
a verdict — so `abstained` was 4–8 while the number of seats that reviewed-and-declined
was zero.

| round | decision | reviewers | abstained | reviews in body | carried a verdict |
|---|---|---|---|---|---|
| `66d77d4d` | approved | 8 | 8 | 8 | 8 |
| `c67ecb24` | approved | 12 | 4 | 12 | 12 |
| `b896fc22` | rejected | 10 | 6 | 10 | **9** ← one real abstention, and `abstained` does not count it |

So the checks that actually mean something:

```sql
SELECT metadata->>'decision', metadata->>'unreadable',        -- MUST be 0
       jsonb_array_length(body::jsonb->'reviews') AS in_body,
       (SELECT count(*) FROM jsonb_array_elements(body::jsonb->'reviews') r
         WHERE r->>'verdict' IN ('approve','object')) AS voted -- MUST equal in_body
FROM diagnosis_artifacts
WHERE kind='council_report' AND correlation_id='<SUBMISSION_CORR>'
ORDER BY created_at DESC LIMIT 1;
```

**Do not carry the experience-loop's rule over here.** `RUNBOOK_experience_loop.md`
(~line 416) says *"an approved verdict is only trustworthy alongside `abstained: 0`"* —
**correct for that council and wrong for this one.** That council has 4 seats and no
relevance filter, so an abstention there really is a flaked seat. On the gate,
`abstained: 0` would mean all 16 seats fired, which happens for the broadest changes
and says nothing about whether the round was sound. Cost of the confusion: a
genuinely unanimous 8-seat approval that reads as "every seat abstained".

### Traps met in real use (2026-07-18, bugs_open/012 submission)

- **On a resubmit, update the `sketch` fields — not just the `rationale`.**
  Reviewers judge the **sketch**; it is the only view of your code they get. A
  round-2 submission that fixed the SQL and described the fix in prose, while
  leaving the round-1 sketch in place, drew two confident objections about code
  that no longer existed (debug-historian: "no `error_step IS NULL` predicate
  visible anywhere in the WHERE clause" — it was in the file, not in the
  sketch). Those objections cost a whole round and read as real defects.
- **Seats can contradict each other across rounds, and that is not a bug in
  your plan.** Round 1: edit-quality + guardian objected to a refactor as scope
  creep and it was withdrawn. Round 2: the reuse seat objected that there were
  now two near-identical recorders. Both readings are defensible. Advisory means
  advisory — pick one, record WHY in the code, and move on (the withdrawal note
  now lives in `store_generated_component_action.go`).
- **The printed `RUN_ORCH_ID` is not the orchestration the chassis creates.** It
  assigns its own id and its own `generic-orchestrate-*` name, so looking up the
  printed id returns 0 rows and a healthy run looks dropped. Find your run by
  payload instead:
  `SELECT orchestration_id, status, current_step FROM orchestration_states
   WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';`
  This matters beyond debugging: CLAUDE.md offers `RUN_ORCH_ID` as an
  alternative for the `Council-Reviewed:` trailer, and for a **gate** submission
  that id never exists — so the 098 join would silently miss it. Use
  `SUBMISSION_CORR`.

  > **PARTLY CORRECTED 2026-07-19 (council-gate thread) — the conclusion stands,
  > the reason is narrower.** Your advice is right and CLAUDE.md now says
  > "use `SUBMISSION_CORR`". But "that id never exists" is too strong as a
  > general rule: on the commissioning smoke run the printed orchestration id
  > (`72e552df`) *did* resolve — 1 row in `orchestration_states` and **2
  > artifacts carrying it**. So the chassis does sometimes honour the envelope's
  > `orchestration_id` header. The likeliest reading of your 0-row lookup is
  > your own fourth trap immediately below: the run had not appeared yet.
  > Either way the safe rule is unchanged, and now rests on a property that is
  > always true rather than a behaviour that varies: **the correlation is the
  > key artifacts are written under** (`diagnosis_artifacts.correlation_id`), so
  > `SUBMISSION_CORR` always resolves; an orchestration id resolves only if the
  > artifacts happen to carry it. Evidence for the record — the three imagery
  > rounds on `098b29b8` carry orchestration ids `5aa40a0a`, `5cb5b43c`,
  > `82f43425` (one per round), while the correlation is stable across all
  > three, which is exactly why the correlation is the right trailer.
- **Runs can be slow to start.** A submission may sit before its orchestration
  row appears. Absence a minute later is not evidence of a dropped dispatch —
  poll by `fix_correlation_id` before concluding anything (this cost two
  needless resubmissions).

  > **NOW MEASURED, 2026-07-20 (council-gate thread): publish → run start was
  > 29 minutes** under normal fleet load (published 18:51:57Z, run `0b8bcc1b`
  > created 19:20:36Z, completed normally). **Budget ~30 minutes end to end**;
  > the council itself is only 2–5 of them. I ignored this very bullet, called
  > it a dropped dispatch at 13 minutes, resubmitted, and spent ~25 minutes on
  > publish-path forensics — logged in `WRONG_CALLS.md`. If you genuinely need
  > to know whether the message left, read the topic instead of guessing:
  > `kcat -C -b <bootstrap> -t system.agent.generic.requests -o -400 -e -q -f '%T|%s\n' | grep <corr>`
  > — that proves the publish independently of the consumer, and it showed both
  > of mine had been on the topic all along.
- **The status is `COMPLETED`, uppercase.** A poll loop with
  `case "$R" in *completed*) break;;` never fires. Lower-case the subject
  (`${R,,}`) or match both — this cost 10 minutes of a 12-iteration loop.
- **A resubmission must re-state ALL standing evidence, not just the new
  round's** (2026-07-20, bugs_open/022 submission, corr `0328ddc7`). Seats
  have no cross-round memory: round 4's rationale answered round 3's
  objections but dropped the round-1 reuse citation (parseHexColor/
  relativeLuminance are pre-existing color_util.go helpers), and reuse_agent
  — which had APPROVED rounds 1–3, praising that exact reuse — objected
  "introduces new colour-math primitives ... never shown being defined or
  imported" and decided a REVISE on it. Same-package helpers never appear in
  a diff, so the rationale is the only place a reviewer can learn they
  pre-exist. Carry the full evidence base forward every round; each
  resubmission is judged standalone.
- **The plan schema is stricter than the 097 header suggests — validate the
  types before submitting** (2026-07-25, bugfix-006 submission: two runs died
  `complete_invalid` at `persist_submission` before any reviewer fired).
  `fixPlan` (diagnose_persist_fix_plan_action.go): `risks` is a **single
  string**, not an array (an array kills the whole unmarshal:
  `cannot unmarshal array into Go struct field fixPlan.risks`); `edits[].operation`
  allowlist is exactly **modify | add | remove | config_change** — `create` is
  refused (use `add` for a new file); `grounded_in` IS a string array; ≤8 edits,
  ≤32KB single-plan bytes. Also mind `noOpEditReason`: a sketch containing
  "add a comment" / "no change needed"-family phrases is refused as a no-op.
  An invalid run writes **no artifacts at all** — polling
  `diagnosis_artifacts` by corr waits forever; poll `orchestration_states.current_step`
  too and treat `complete_invalid` as terminal (the error is in
  `collected_data->'__step_error'`). Cheap pre-check: unmarshal your JSON
  against the struct fields at lines 60–77 of the action file before firing.
  - **Contributed 2026-07-25 (040-partial-build thread): a THIRD invalid run, same
    morning, hit independently.** Same two causes in one submission — `risks` as an
    array, `operation: "create"` — dispatched 08:34, dead at `complete_invalid` by
    08:34:34, an hour before the bullet above was written. Two threads finding the
    identical trap within an hour, neither able to warn the other, is the case for
    validating **client-side in 097** rather than documenting harder: the struct is
    four fields and the trigger already refuses out-of-scope paths before spending
    credits, so it has the stage to do it in. Noting, not building — the gate
    tooling is the council-gate thread's. Second-order trap worth more than the
    schema itself: a submission that dies this way looks *exactly* like a queued
    one to any watcher keyed on `diagnosis_artifacts` alone, and I lost an hour to
    that (`WRONG_CALLS.md`, 2026-07-25).

## Honest limits (advisory mode)

- The gate cannot intercept a hand-commit to the shared branch; the 098
  report makes the gap visible, nothing more. The first live run of 098
  (3-day window) found **28 in-scope commits, 0 reviewed** — that number is
  the baseline the gate is judged against.
- A council run costs credits and minutes. Submit per coherent task, not per
  iteration; PR-cadence batching arrives only with PR-mode.
- The trailer is self-declared. MISMATCH (trailer without a green report) is
  bucketed separately by 098 precisely so a false claim of review is visible.

## The live roster (re-verify before assuming — it changes often)

As of 2026-07-18, 9 seats. Always-on: **edit-quality**, **guardian** (hard
veto). Relevance-gated advisory seats and their footprints:
**bug-historian** (rebuild/rerender/render paths), **reuse-agent** (new
code/SQL/migrations), **guidelines** (contracts, work-items, agent defs,
schemas), **tooling-provenance** (contextkit/bundle, doc_plans/doc_notes,
registry), **adoption-guardian**, **diagnosis-guardian**,
**improvement-guardian**. Check the truth with:

```sql
SELECT jsonb_array_length(default_config->'workflow'->'steps'->'council_decide'->'config'->'review_fields')
FROM agent_definitions WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Re-sync procedure when the fix-proposer gains a seat: read its live row, copy
the reviewer step verbatim (swap `## The diagnosis` /
`{{.diagnosis_row.conclusion}}` for `## The author's stated rationale` /
`{{.input_data.rationale}}`, and `error_step` → `complete_invalid`), add its
gate + footprint, extend `review_fields`/`check_fields`, re-run the literal +
routing validator, then re-apply this file (`snapshot_agent` backs up first).

**After any seat/roster change, run the parity lint** (read-only, no credits):
`python3 102_LINT_council_seat_parity.py`. It reads the LIVE rosters and flags a
seat that has drifted from its siblings — a config key the rest of the family
carries but it lacks (this is how `bugs_open/019`'s last hole survived:
`review_prior_art` shipped without `tolerate_truncation` while the other 15 seats
had it), a stale model/ceiling, or a seat wired into the workflow but absent from
`council_decide.review_fields`. It is the complement to `099`: 099 compares the
two councils against EACH OTHER and is therefore blind to a key missing on BOTH;
102 compares each seat against its OWN family, one council at a time. Clean on the
whole live fleet as of 2026-07-21 (`--all-families` adds a noisier general audit).

## Cross-links

- `HANDOFF_2026-07-17_council_gate_thread.md` — this thread's cold-start.
- `DESIGN_feature_builder_and_council_gate.md` §2 — the design executed here.
- `../../docs026_concept_register/` — stage 3 = the seat roster (launch gate);
  `PILOT_bug_historian_reviewer.md` is the proven seat-adding pattern;
  `PILOT_reuse_agent_reviewer.md` is seat #4 awaiting sign-off.
- Multi-session coordination workstream — commit-per-task + build-from-ref;
  the `Council-Reviewed:` trailer composes with those rules.
