# HANDOFF 2026-09-03 (rev 3, night) — bugfix 440: PHASE 3 IS BUILT, PROVEN BY EXECUTION, AND HELD AT ONE SIGNATURE

Written for a session with none of this context. Every claim carries its check; cite symbols,
never line numbers. Rev 3 supersedes rev 2 in place (same filename, deliberately: one canonical
"continue here" per lane). **What changed since rev 2: phase 3 is written and tested end to end,
D1 turned out to need a code change, and the flip was found to leave a NEW blind spot in this
bug's own drift class.**

## The bug in one paragraph

`spec.reason` on `page_rerender` items is TWO fields wearing one name: the gate's routing key AND
free human prose. The live gate (`check_rerender_mode`: five-value `==` disjunction on
`spec.reason`, `else_step: render_page`) silently ASSEMBLES anything it doesn't recognise, so a
routing key nobody understands completes green having changed nothing. Refusal is impossible
until the fields split (`spec.routing_reason` = vocabulary-only; `spec.reason` = annotation, free
forever). Split + refusal = RFC_062; evidence = `bugs_open/440_HANDOFF_2026-09-02_…`.

## OWNER RULINGS — settled, do not re-open (full text: RFC_062 §Rulings)

| # | ruling |
|---|---|
| D1 | a refused item routes to **`needs_human_review`** (message names the bad key AND the vocabulary) — not a silent assemble, not a blunt orchestration failure |
| D2 | the **404 lane CO-SIGNS** the phase-3 gate migration (the declarations are theirs) |
| D3 | **YES to the write-door CHECK constraint** (NOT VALID first, validated after census) |
| D4 | **NO policing of `spec.reason`** — the annotation stays free prose forever |
| D5 | phase 1b's courtesy gate on the 404 lane **LIFTED** (acted on: 1b shipped same day) |
| **D6 (new, 2026-09-03 eve)** | **D1's message: use the OPT-IN TEMPLATE, not a static literal.** `fail_work_item`'s `error_message` is a config literal with no interpolation, so a static message can name the FIELD and the vocabulary but never the offending VALUE |
| **D7 (new, 2026-09-03 eve)** | **D2's co-sign: BUILD IT ALL, STOP BEFORE APPLYING.** The 404 lane is dormant since 2026-08-26 and the co-sign CONTRIB is unanswered. Nothing goes live without it |

## State

| what | state | re-check |
|---|---|---|
| phase 1a (livespec foundation, REB-008) | **APPROVED r1 `55def842`; SHIPPED** | `platform/livespec/rerender_routing_key{,_test}.go`, commit `a3758c399`. ⚠ never literal-probe zero-caller code — DCE strips it |
| phase 1b (creator stamps the key) | **APPROVED r1 `934327db`; SHIPPED** | `create_rerender_items_action.go`; stamps in LOCKSTEP with `spec.reason` |
| phase 2 (producer conversion + raw-SQL door) | **APPROVED `c7dab2c1` + `3b484a74`; LIVE AND PROVEN IN PRODUCTION** | `[MEASURED 2026-09-03]` **14** items now carry `routing_reason` (12 earlier the same day) — the converted `check_misdirected_cta.go` path is stamping. One conversion DEFERRED: `refresh_evidence_base_action.go` (another session had 245 uncommitted lines in it) |
| **phase 3 (THE FLIP)** | **BUILT, PROVEN, HELD** | `docs/agent_docs/sql_for_agents/741_refuse_unknown_rerender_routing_key_HOLD.sql` + `_ROLLBACK` + `_VERIFY`; the Go half is `fail_work_item`'s `error_message_template` (**WII-038**). **NOT through council yet** — see "what to do first" |
| narrowing (phase 3b) | **BLOCKED and quantified** | 1,804 pending items route on `reason`, 14 on `routing_reason` |

## WHAT TO DO FIRST, in this order

1. ~~**Dispatch the council round for phase 3.**~~ **DONE 2026-09-03 14:1xZ, after the chassis
   roll landed** (`v1.0.1359`, replicaset `85c4984f77` 2/2 ready, pods 38 min old — well past the
   ~300s window in which a spawn is silently dropped, and after the roll rather than before it,
   because a roll KILLS an in-flight council).
   **`SUBMISSION_CORR = 56047b18-9e0a-4107-a781-007df8ff59bd`** — submission
   `submission_440_phase3_r1.json` (6 edits, 10 `grounded_in` quotes, 7 stated risks).
   **READ THE VERDICT FIRST, before anything else in this list:**
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='56047b18-9e0a-4107-a781-007df8ff59bd' AND kind='council_report'
   ORDER BY created_at;
   SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
   ```
   Budget ~30 minutes, not ~2 — the council itself takes 2–5 min but the dispatch queues behind
   the fleet. A missing orchestration row is almost always latency, not a dropped dispatch; do not
   retry on that evidence. REVISE → resubmit with `RESUBMIT_CORR=56047b18-…`.
   ⚠ **The code commit `83407cd37` carries NO trailer and will list as un-reviewed in the `098`
   report.** That is accurate rather than an oversight: the round was dispatched *after* the
   commit, deliberately, because a chassis roll was imminent and forward-only forbids an amend.
   The `Council-Submitted:` trailer is on the follow-up commit that carries the submission JSON,
   so the correlation is in the git record. Do NOT write `Council-Reviewed:` anywhere until you
   have read an approved verdict — `098` buckets that as MISMATCH, its dishonesty surface.
   ⚠ Two things that already cost this lane something, for the resubmit if there is one: the
   pre-dispatch self-check (every symbol named in `symbol` must appear in its sketch, and every
   symbol NAMED must be visibly defined or explained) caught
   `TestErrorMessageTemplateIsAbsentFromEveryLiveStep` named but shown only as a comment — the
   exact class that drew a medium at round `c7dab2c1`. And `DRY_RUN=1` caught `operation:
   "create"`, which is not in `allowedFixOperations` (a new file is `add`) — free, before credits.
2. **Get the 404 co-sign (D2/D7).** That is the only release condition on 741. The lane is
   `docs024_key_docs_latest/bugfix_404_rerender_reason_vocabulary/`, dormant since 2026-08-26; a
   CONTRIB asking for the co-sign has sat in their NOTES since 2026-09-02. If it stays dormant,
   this is the owner's call, not a session's — and note the premise has WEAKENED: their two
   Declarations turn out not to need editing at all (below).
3. **Apply 741 — AND COMMIT THE LIVESPEC DECLARATION EDITS IN THE SAME COMMIT.** The five
   specific edits are enumerated in 741's own header. They are held back on purpose: the daily
   auditor's own note says to fix a declaration *"in the same commit as the migration that moved
   it"*, and today's row reads `probed 15 live object(s); 0 finding(s)` — committing them ahead of
   the apply turns that clean 0 red every morning and masks real drift.
4. **Then close 440** on an INDUCED unknown routing key landing in `needs_human_review` (recipe in
   the RUNBOOK). A census of zero refusals proves nothing — it is equally consistent with
   "nothing bad written yet" and "the refusal branch is unreachable".

## ⚠ THE FINDING THAT MATTERS MOST: the flip leaves a NEW blind spot, in 404's own drift class

`[MEASURED 2026-09-03, BY EXECUTION]` **both** of the 404 lane's live livespec Declarations stay
**GREEN** through the flip, unchanged and un-edited:

- the `FragmentMatch` on `CheckRerenderModeConditionClause()` holds — the old five-value clause is
  a substring of `TransitionRerenderModeConditionClause()` **exactly once**, so Min:1/Max:1 is satisfied;
- the paired count still reads **5** — it counts `input_data.spec.reason ==`, and
  `input_data.spec.routing_reason ==` does not contain that as a substring.

**That is convenient and it is a defect.** Five new `routing_reason ==` disjuncts arrive asserted
by NOTHING, so a sixth routing value appended to the live gate without touching Go would drift
exactly the way `bugs_open/404` drifted — inside the change built to fix it. The existing
Declaration's own comment states the principle it now fails: *"A fragment sees loss and mutation;
only a count sees ADDITION."* Remedy enumerated in 741's header (step (b) is the load-bearing one).

## Flip-day safety, measured rather than argued (all `[MEASURED 2026-09-03]`, live)

| question | answer |
|---|---|
| pending items that would be REFUSED on flip day | **0** — no pending item carries a present-but-unknown routing key |
| rows in the WHOLE table (every status) that would fail the CHECK | **0** — so VALIDATE can follow the apply immediately, not eventually |
| items where the two keys DISAGREE | **0**, and **nothing enforces it** — the residual below |
| pending items with an in-vocabulary `reason` / with `routing_reason` | **1,804 / 14** |
| `fail_work_item` live steps / agents / already carrying `{{` | **7 / 6 / 0** |

## The residual phase 3 does NOT close, stated rather than left implicit

`keys_disagree` is 0 **as a property of the producers, not of the gate.** An item carrying
`routing_reason='literal_markdown'` with free-prose `reason` would be routed by the transition
clause (it matches either key) while `rerender_sections` is still handed `input_data.spec.reason`
— so `shouldStripLiteralMarkdown` and the CTA recompute would see the annotation and silently
under-deliver: this bug's own shape in miniature. Those readers move to `routing_reason` at the
NARROWING (phase 3b); until then `_VERIFY` counts the state so it cannot arrive unnoticed. Flagged
to the council as a known risk rather than fixed here — inventing a new constraint now would be
re-deciding a ruled design.

## Craft notes (each cost a round or a correction)

- **Submission accuracy is where this lane keeps bleeding — not design.** Five consecutive rounds
  were gated or objected-to on what the submission SHOWED. Attach the query/grep OUTPUT; list
  EVERY file the commit touches; name the prior round's commit when an edit builds on
  shipped-but-unlisted symbols.
- **A single measurement of something PROBABILISTIC certifies whichever answer you got.** The
  `ALTER TABLE ... ADD CONSTRAINT NOT VALID` in 741 took **2 ms** on one dry run and was **still
  waiting after 2 minutes** on an earlier one. `NOT VALID` skips the SCAN, not the LOCK, and a
  queued `ACCESS EXCLUSIVE` blocks every later reader and writer of `site_work_items` behind it —
  so the bad case stalls the work-item pipeline. `SET LOCAL lock_timeout` is in the file; re-run
  in a quiet window rather than raising it. LANDMINES has the general form.
- **Dry-run a config migration by executing it and rolling back** — the migration's own `DO`-block
  verify runs, and the pasted strings can be read back out of `agent_definitions` and diffed
  against the Go renderers (verified IDENTICAL, all four). Recipe in the RUNBOOK. Then PROVE the
  tree is clean rather than assuming the rollback happened.
- **`python3 scripts/pattern-check.py <file>` ignores the filename** and lints the git INDEX — it
  printed nothing on a deliberately bad migration. Call the check FUNCTION, with a positive control
  in the same breath. And `migration_is_lintable()` takes a **basename**: passing a path returns
  False and nearly produced a correction against a TRUE claim.
- **Declaring one config key can arm a detector against every key that was already undeclared** —
  and it then calls keys the action DOES read "silently ignored". Enumerate them all and declare
  them in one move; assert with a live step shape AND a bogus-key negative control.
- **On a shared tree, a test FAIL naming symbols you have never heard of is a neighbour's WIP.**
  `platform/livespec` still FAILS at HEAD regardless of this lane
  (`TestNoNewMigrationFileReadersOutsideTheAllowList`, now firing on another lane's
  `write_audit_findings_origin_test.go`). Theirs — and it is also why this lane did NOT add a
  Go test that reads 741's text.
- **The git INDEX is shared too.** `git diff --cached --name-only` showed three other lanes' staged
  files while this work was in progress. The pathspec on **`commit`** is what keeps them out.

## Key artefacts

| what | where |
|---|---|
| bug file | `bugs_open/440_HANDOFF_2026-09-02_a_routing_key_nobody_understands_completes_green.md` |
| design + rulings + §Phase 3 as built | `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_062_routing_key_annotation_split.md` |
| the flip | `docs/agent_docs/sql_for_agents/741_refuse_unknown_rerender_routing_key_HOLD{,_ROLLBACK,_VERIFY}.sql` |
| the Go half | `platform/orchestration/actions/fail_work_item_message_template{,_test}.go`; register **WII-038** |
| lane docs | this directory — PLAN (phases 3/3b) · NOTES (evidence + missteps, newest at bottom) · RUNBOOK (dry-run, paste-proof, induction, pattern-check) · README (owner-facing prose) |
| council | `55def842` (1a) · `934327db` (1b) · `c7dab2c1` (2a) · `3b484a74` (2b) — all APPROVED r1. **Phase 3 has NO round yet**, and the evaluator `== null` fix rides it |
