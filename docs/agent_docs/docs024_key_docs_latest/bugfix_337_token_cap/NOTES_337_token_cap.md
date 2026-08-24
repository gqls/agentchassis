# NOTES — bugfix 337 (generate_template blows its 16,000-token cap)

Append-only, newest at the bottom. Technical log: what was tried, what the system said,
and every misstep.

## 2026-08-22 — session start: ownership, validity, first measurements

**Ownership check.** `scripts/who-owns.py 337` names `bugfix_311_component_keys` as the
likely owner (9 mentions), but their `HANDOFF_2026-08-20_continue_here.md:155` lists 337 as
"filed from this lane, 08-20" and the bug file itself says **"Status: OPEN, unowned"** —
the lane's live work is 311/345/351, not 337. Queue check: no open `site_work_items` row
targets 337 or a fix for it (query in RUNBOOK). Taking the bug; contributing findings into
`bugs_open/337_…md`, not a parallel account.

**Bug still valid** [MEASURED 2026-08-22 ~09:55Z]:
- Live `agent_definitions` row: `component-creator.generate_template` still
  `{"model":"claude-sonnet-4-6","provider":"anthropic","max_tokens":16000}`.
- The failed items still parked `failed`, `attempt_count=3`:
  loanzy.uk `needs_new_component:loans-credit-health-check` (08-19),
  loanzy.uk `…_run1` (08-18 — a FOURTH item the bug file predates),
  loancalculator.co.uk `needs_new_component:loans-credit-health-check` (08-15).
- `llm_call_log`: **9 truncation failures** for `generate_template`, 08-15 → 08-19, every
  one `output_tokens=16000`, recovered chars 46,441–48,817. Matches the bug file's
  three-items × three-attempts account.

**New measurement — the cap is tight even on SUCCESSES** [MEASURED 2026-08-22, 14-day
window, `llm_call_log`, successful calls with output_tokens & max_tokens present]:
- `component-creator/generate_template`: 154 calls, p50 8,894, **p95 13,633 of a 16,000
  cap (85%)**, max 15,374 (96%), 6 calls ≥90% of cap. The step runs hot for everything,
  not just `loans-credit-health-check`.
- Same census fleet-wide: `council-gate/review_editquality` max 15,777/16,000 (98.6%);
  `diagnose-agent/verdict` max 31,033/32,000 (97%); `page-content-writer` rewrite_negations
  2 at-cap at 2,000. The near-cap class is fleet-wide, which is what makes "threshold
  management" (the owner's steer) a framework question, not a one-step tweak.
- All-history truncation-failure census per step (error `%reached the configured cap%`):
  top rows `vet-practice-verifier/extract_and_reconcile` 94 (08-05→07, cap 2048),
  `generic+council-gate/review_editquality` 49, `component-creator/generate_template` 9,
  `tool-auditor/llm_audit` 9, `diagnose-agent/verdict` 4, `tool-improver/improve_tool` 3.

**Code read (transport → step → item):**
- `platform/aiservice/truncation.go` — typed `TruncatedError` CARRIES the partial;
  `IsTruncated` is the opt-in salvage hook. Its header states the design position:
  raising caps is NOT a class fix ("whatever the number, the step that writes most
  approaches it on the work most worth doing").
- `platform/aiservice/max_tokens.go` — budget resolution precedence: per-call
  `options["max_tokens"]` wins → `ai_service.max_tokens` at client construction →
  `DefaultMaxOutputTokens` 2048. So a per-call escalated value needs NO client change.
- `platform/orchestration/actions/ai_actions.go:415-553` — `execute_llm_prompt`'s
  truncation path: `tolerate_truncation` opt-in (with the 076 guarded-consumer check),
  5xx transient retries, `AIUnavailableError` back-to-triage. `:743-747`: the 119 JSON
  re-ask DELIBERATELY does not raise max_tokens, citing truncation.go. Distinction that
  matters for this fix: that re-ask is for a *judgement* that can be asked shorter; a
  47k-char *document* cannot be — its length is the work product.
- `platform/errors/permanent_failure.go` — `MatchedPermanentFailure` is deliberately
  closed/conservative; the truncation message matches no needle, so item-level retries
  burn all 3 attempts (candidate 4 territory — noted, not taken).
- `anthropic.go` sends whatever `options["max_tokens"]` says — no clamp.
- `./scripts/audit-optional-key-budget.sh`: `execute_llm_prompt` has **no
  ActionInputSpec** (67 carriers, listed under "NOT COUNTED — surface UNKNOWABLE, not
  zero"). A new opt-in key does not move the RFC_022 counter; the missing spec is a
  pre-existing gap to name, not to fix here.

**Missteps so far:** none recorded yet.

## 2026-08-22 — design decision: escalate-on-truncation + resize + headroom check

Decision and reasons in `PLAN_2026-08-22_token_cap_management.md`. In one line: the
routine cap stays a cost control, a per-step opt-in `max_tokens_ceiling` gives
`execute_llm_prompt` ONE in-call retry at the ceiling when the provider says
`stop_reason=max_tokens`, the step's cap is resized from measurement (16,000 → 24,000),
and a daily headroom check over `llm_call_log` makes every step's cap a monitored
threshold instead of folklore.

> **CORRECTED 2026-08-22 (same session, before any code):** the third leg — "a daily
> headroom check" — was about to duplicate a mechanism that already exists and already
> fired: LCO-007 `fleet-step-token-pressure` (6-hourly) has flagged this exact step in
> `doc_notes` since 08-18. Caught by the prior-art sweep. No new monitor is built; the
> gap is flag→action dispatch, named as a residual in the bug file. See next section.

## 2026-08-22 — implementation + council submission

- Prior-art sweep (subagent, full report absorbed into PLAN) **corrected the draft plan
  before any code**: the "threshold management" monitor I was about to build already
  exists — LCO-007 `fleet-step-token-pressure` (6-hourly, C/T/N/P) — and it has flagged
  `T generate_template@16000 … truncated 9` in doc_notes since 08-18 with no consumer.
  Not a WRONG_CALL (caught pre-assertion, pre-code, by the research pass working as
  intended); recorded here so the near-miss is visible.
- Throughput measured to size the ceiling: the nine cut calls took 165.6–170.2s for
  16,000 tokens (~94–97 tok/s); worst successful throughput 91.7 tok/s. 32,000 is
  clock-safe (~349s worst) against the 600s non-streaming timeout; 40,000 at a
  conservative 60 tok/s is not (667s).
- At-cap SUCCESS rows (output_tokens=16000, success=t) checked before they could
  poison sizing: all April 2026 — pre-TruncatedError history, ignored.
- Shipped: `truncation_escalation.go` (+17-case test, green), wiring in
  `ai_actions.go` (between GenerateText and the error block — the existing block only
  ever sees the FINAL attempt's error), migration 549 (+ROLLBACK), register MDL-042,
  bug-file append. `go vet` clean for my files (one pre-existing warning in
  `load_component_library_actions.go:207`, not mine).
- Council: DRY_RUN passed, submitted for real.
  **SUBMISSION_CORR = 3d531c9a-4351-42bc-806c-17ed25636a8c** — budget ~30 min; find
  the run by payload, not by the printed id.

## 2026-08-22 — before-state pinned; council run located

- Council run found by payload: orchestration `4f031702-0e36-4a4e-a1b1-9eb5d8b00f06`,
  dispatched 09:21:36Z (fast — no queue delay this time), `gate_bug_historian` at
  09:25Z. Watching for the verdict; migration apply waits for it per the PLAN phasing.
- Before-state pinned, TWO reads each, stable:
  - `https://loanzy.uk/tools/credit-health-check/index.html` — 24,323 bytes,
    md5 `bdc997300740612af0625a53d61416d3`, `grep -c '<input'` = **0**.
  - `https://loancalculator.co.uk/tools/credit-roadmap/index.html` — 1,201 bytes,
    md5 `8561e9f73e44ec554f8864deea13833a`, `<input` = **0** (a stub — this page has
    never been built, consistent with `incomplete_page_group` open item).
- Migration runner dry-run running in background (slow — per-file ledger checks).

## 2026-08-22 — same-file passenger disclosure on commit 9e23fb852

My WRONG_CALLS append (the LCO-007 duplicate-monitor near-miss) committed at 68
insertions where my entry is ~35 lines: **the bugfix_342 lane's uncommitted entry ("I
read a VACUOUS zero as a clean census") rode along** — the same-file-passenger case a
pathspec commit cannot exclude. Nothing lost, forward-only holds; their entry is intact
under my commit message. 342 lane: your WRONG_CALLS entry is already committed
(9e23fb852) — do not re-append it.

## 2026-08-22 — council round 1: REVISE, and the gating objection found a REAL defect

Verdict at 09:37:52Z, `decided_by: gating objection from editquality`, 4 abstained,
`gated_by_truncation: false`. Objection on edit 3, verbatim: *"The wiring reads
`options["__sent_max_tokens"]` to get sentCap for the refusal check
(`ceiling <= sentMaxTokens`), but no edit anywhere sets this key. If it isn't already an
established key elsewhere in ai_actions.go, sentCap silently defaults to 0 on every
call, which makes the stated safety refusal … "*

**Checked rather than argued, and the objection is RIGHT — narrowly, and worse than its
own framing.** The key IS established (`grep -rn '__sent_max_tokens' --include=*.go`):
- `anthropic.go:313-318` — set from `requestBody["max_tokens"]`, guarded only by
  `options != nil`, and `options` is never nil on this path (`ai_actions.go:350`
  `make(map…)`). Always set.
- `gemini.go:376` — unconditional inside the same `options != nil`. Always set.
- **`ollama.go:121` — set ONLY when `optionsBlock["num_predict"]` is present as an int.
  And ollama deliberately OMITS `num_predict` when no budget is configured
  (`aiservice/max_tokens.go`: "ollama does NOT use it — it omits the field entirely
  when unconfigured").** So an unconfigured ollama step reaches the escalation with no
  wire number at all.

So the reviewer's failure mode is reachable on one provider, and its consequence is the
inverse of safety: `sentCap` = 0 makes `ceiling <= 0` false for EVERY positive ceiling,
so the refusal I documented as the safety property would have been **vacuous** — it
would escalate against a baseline nobody established. A comment asserting a guard that
the code cannot enforce is exactly the estate's "a doc comment is not an enforcement
mechanism" shape.

**Fix (round 2):** carry the type assertion's `ok` instead of discarding it —
`sentCap, sentCapKnown := options["__sent_max_tokens"].(int)` — and make UNKNOWN a
third state that **refuses to escalate** (fail-closed). Same three-state discipline
`rewrite_negations_action.go` applies to unreported usage. Two new test rows pin it, and
**the guard is mutation-proven, not asserted**: deleting `if !sentKnown { return 0,
false }` fails exactly those two cases and nothing else (run 2026-08-22; the rest of the
table stayed green, so the mutant is not killed by collateral). No behaviour change for
the motivating anthropic case.

**Cost of the round: one REVISE, ~16 minutes, and it bought a defect that no test I had
written could see** — my table set `sent` as a plain int, so absence was unrepresentable
in the fixture. The estate's line holds again: a REVISE round is cheaper than the defect
it finds.

## 2026-08-22 — council round 2 APPROVED; advisories answered; migration 549 APPLIED

**APPROVED** (`decided_by`: "approved with 3 advisory objection(s) — none high-severity",
4 abstained). 8 seats clean; 4 carried objections, all advisory, all answered:

- **prior_art_librarian [low] — RIGHT, and it caught an unmeasured claim of mine.** The
  "067-sweep: generate_template is component-creator's only `execute_llm_prompt` step"
  line carried no `[MEASURED]` tag because I had asserted it from a step-NAME list, not
  a query. Now queried and recorded in the migration header: six steps, exactly one LLM
  action, none of the other five carries a cap. The query could have returned a second
  LLM step and did not.
- **reuse_agent [medium] — checked, declined, reason recorded IN THE CODE.** Unlike
  `aiservice/max_tokens.go` (where the same objection was answered by an import cycle),
  `package actions` can import `datahelpers` freely, so reuse genuinely was available.
  Declined anyway: `GetIntField` handles float64+int only, while the SIBLING key in the
  same config block (`max_tokens`) is read by `aiservice.configMaxTokens`, which also
  takes int64/json.Number. Two coercion rules for two keys of one block is the
  two-readers-of-one-concept drift class. ~10 lines to match the sibling. Noted that if
  257 candidate 2 unifies them, this should follow rather than keep its copy.
- **guardian [medium+low]** — blast radius on a 67-carrier shared action, and the
  in-place `options` mutation. Both already answered by construction: opt-in (no key →
  byte-identical path, pinned by test) and the provider clients read `options` per-call
  inside `generate()`, with `LogLLMCall` reading the map synchronously before its
  goroutine (`llm_call_logger.go:35-39`).
- **bug_historian [medium]** — no remediation step for the two hollow pages. Correct:
  the forward path and the repair are separate, and the repair is PLAN phases 5-7. Doing
  it next; it is the close-out bar, not an optional extra.
- **editquality [low]** — MDL-042 not among the plan's edits. It IS committed (register
  entry + `000_concept_index.md` row, commits `9e89e8ca1`/`c7b2c708e`); the plan's 4-edit
  cap meant docs were described in the rationale rather than listed. No action.

**Migration 549 APPLIED 09:56:36Z**, scoped via `MIGRATIONS_DIR=<scratch dir holding only
549>` — a bare `--apply` would have swept ~12 other threads' pending files. Snapshot
taken (`23720180-…`), `UPDATE 1`, ledger recorded. Verified at the live row by the
RESOLVED value, not the written key: **resolved_cap 24000, ceiling 32000, dead
`config.max_tokens` key NULL, version 2.**

## 2026-08-22 — VERIFICATION BY DEMAND OVERTURNED THE BUG'S DIAGNOSIS (the session's main finding)

Re-drive of both live pages, items filed 09:59Z, both `complete` at `attempt_count=0`.
Three generations at cap 24000: **14,244 / 12,709 / 14,816 tokens, zero truncations.**

**And all three are BELOW the old 16,000 cap** — which cannot happen if the brief
"reliably exceeds" it. That forced the census I should have run at the start, joining
`llm_call_log` to `site_work_items` ON `work_item_id` and filtering on the item's own
`spec->>'section_type'`:

- **82 generations of `loans-credit-health-check`, 73 SUCCESSFUL at cap 16000**
  (8,641–15,374 tokens), **9 cut.** The cap fits ~89% of the time.
- Per item, `attempt_count` 3 vs **13 / 55 / 11 actual LLM calls** — an in-workflow
  regeneration loop that never touches the attempt counter. (The 55-call item is
  `8c8f5de5`, which `bugs_open/345`/migration 533 measured from the other end as "52
  rejections in 3h34m". Two lanes measured the same item and neither saw the other.)
- The real blocker, proven the same hour AT A TALLER CAP: loanzy's fresh generation was
  refused by `store_component` pre-store validation — `field "cta_primary_url" declares
  source "site_specs.ctas.primary_url" but no site carries a site_specs aspect named
  "ctas" … (bugs_open/309)`. No component stored, despite a clean 12.7k-token generation.

**My misstep, logged in WRONG_CALLS:** I inherited "nine cap-hits, zero successes, 100%
reproducible" from the bug file, repeated it in PLAN/NOTES, put it in a council
submission's `grounded_in`, and was APPROVED on it. Thirteen seats could not catch it —
every figure I gave them was true of the population I had selected. The bug file's census
(`site_work_items WHERE error ILIKE '%reached the configured cap%'`) selects on a
LAST-WRITE-WINS column, so it could only ever return items whose final error was the thing
being tested, and the 73 successes were structurally invisible (different table, one row
per CALL not per ITEM). Pattern written up in `016b` §9.

**Disposition — re-scoped, NOT reverted.** 549 + MDL-042 stand on independent evidence
(successful-call p95 13,633/16,000 = 85%, max 96%; LCO-007 flagging since 08-18; 9 real
truncations in 82 calls is a genuine ~11% loss). They must NOT be credited with healing the
pages, and 337 must not close on them. Bug re-scoped in its own file to the validation-driven
regeneration loop, routed at the 309/345 territory.

**Repair state:** loancalculator's component STORED (22,236 chars, closes properly); page
re-render filed to attach it (page had 5 planned / 4 slots). loanzy: no component — blocked
by 309. Both pages still serve 0 `<input`.

## 2026-08-22 — repair outcome: ONE page genuinely fixed, one still blocked; and two verification missteps of my own

**loancalculator.co.uk / tool-credit-roadmap — REPAIRED, verified at the artefact.**
Component stored 10:02Z (22,236 chars, closes properly), page re-render `complete`,
`page_components` 4 → **5 slots**, `build_status` deployed. Served page
`https://loancalculator.co.uk/tools/credit-roadmap.html` (200, 46,594 B) now carries
`<section class="tool-credit-health-check-section">` with a working quiz: 13 `<button>`,
**4,593 bytes of inline script** (`next()`, `showResult()`, `copy()`, listeners, step/points/
meter state).
⚠ The before/after is **[INFERRED, not measured]** on the served page — see misstep 1. What IS
measured: the component row did not exist before 10:02:18Z today and the page had 4 slots, so
the section could not have been served. Sound, but not the same as a pinned before.

**loanzy.uk / tool-credit-health-check — STILL BLOCKED.** No component: `store_component`
refused the fresh 12,709-token generation on the unresolvable-source rule (`cta_primary_url`
→ `site_specs.ctas.*`, an aspect that does not exist). Page unchanged: 24,323 B, 3 sections,
0 inputs, 1 button. This is the re-scoped 337 (the validation-driven loop), not the cap.

**Misstep 1 — I pinned a 404 as the "before" and my own stability check certified it.**
`/tools/credit-roadmap/index.html` was a name-derived guess; that site serves
`/tools/<name>.html`. The 404 is 1,201 B of real HTML with a stable md5 and zero `<input>`, so
two identical reads and a content grep both passed on the wrong document. `curl -s` without
`-w '%{http_code}'` exits 0 on a 404. **A stability check answers "did it change while I
looked", never "is this the thing I meant".** RUNBOOK updated with the status-code form.
Loanzy's pin was valid (200) — URL shape is per-site, which is exactly what made one right and
one wrong.

**Misstep 2 — the inherited success predicate was the wrong shape.** The bug file's bar was
`grep -c '<input' > 0`. This component is a button-driven quiz and scores **0 inputs while
working perfectly**. An inherited predicate encodes the tool someone EXPECTED. Both logged in
WRONG_CALLS.

---

## 2026-08-22 (evening) — second session: the class census, two refutations of my own, and one destructive mistake

**Where I started.** The re-scope handed this on with "start at the validator/loop, not the
ceiling". Correct instruction; wrong class.

**The census that should have been run before the re-scope was written** — at the call level,
`agent_error_log`, `error_code='component_validation_rejected'`, all 101 rows, 08-15→08-22:
**97 field-contract, 3 source-vocabulary, 1 other.** The re-scope named the 3.

**MISSTEP 1 — I repeated the inherited class before counting it.** I opened by telling the
user the loop was 309's unresolvable-source class, because the bug file said so. One
`GROUP BY` refuted it. Same shape as the previous session's own logged error, one level down.

**MISSTEP 2 — "52 deadlocked components", stated to the user before it was checked.** I
built the census on `function` names, joined `loader(section_type=function)` and counted
misses. The disconfirming test I then ran — do any of these have a successful regeneration
since 08-01 — returned **21 of 52**, which killed it. Re-keyed on actual demand
(`site_work_items.spec->>'section_type'` for `needs_new_component`): **14 types requested, 11
with an advisory present, 3 blind-but-safe, 0 blind-and-stranding.** The cause is
`bugs_open/311`'s diversion, live 08-19 16:22:57Z, creating a `section_type`-carrying row as
a side effect. **97 of the 98 field-contract rejections predate it.** Both missteps logged in
`WRONG_CALLS.md`.

**What survived the refutations, and why it is still worth building.** The *structural*
divergence is real and permanent: the advisory diverges from the gate on three predicates
(`section_type` vs `function`, plus `is_active` and `component_level` filters the gate
deliberately lacks — `component_storage_identity.go:157-165`). 311 closed the
NULL-`section_type` arm by side effect and never touched the `is_active` arm. And the
source-vocabulary arm is the **live** blocker.

**The single most useful measurement of the session**, because it made the fix's value
concrete rather than plausible: the failing component declared `site_specs.ctas.primary_url`
and `.secondary_url`; the real aspect is `cta` and it carries **exactly** `primary_url` and
`secondary_url`, on 4 of 26 sites. The writer's intent was right and only the NAME was
invented — a one-character miss on a vocabulary it has never been shown, because the live
`prompt_template` renders no part of `site_specs` at all.

**The natural experiment for "does enumeration work?"** TIER D's query-name list entered the
prompt **2026-05-07** (`25fe1d318`). All five components that ever invented a query name were
created **≤2026-04-16**; none since. Not proof — but it is the closest thing available and it
could have come out the other way.

**The rate control I ran against my own case, and it moderated the claim.** 72 section
components created since 2026-05-07 carry **zero** phantom aspects at rest. So the class is
rare, and the honest argument is about **consequence**, not volume: the birth gate went live
08-18, the first phantom-aspect rejection is 08-21, and before the gate the same output was
stored and rendered silently blank. **The gate closed the silent door and opened a
page-parking one.** I put that framing in the council submission rather than letting a
reviewer find the rate themselves.

**An avenue I tested and closed.** A fable pass proposed replacing the field-contract proxy
with dependents' real `content_data` keys, on the theory that many refusals guarded nothing —
the guard's own comment at `:448-450` pre-authorises it. Refuted in one query: all 10 refused
components have 1–2 live dependents, and `loans-credit-health-check`'s dependent stores 19
keys of which **12 are `button_*`**. Written into the bug file so it is not re-proposed.

**MISSTEP 3, the expensive one — I ran `git checkout <file>` to undo a mutation and destroyed
another session's uncommitted work.** Mid-way through mutation-proving the `is_active` gate I
reverted with `git checkout platform/orchestration/actions/store_generated_component_action.go`.
That restores the **whole file from the index**, so it reverted ~75 lines of the
`bugs_open/345` lane's in-flight `recordRetryFeedback` along with my own change. No recovery
existed — unstaged work was never in git; `git fsck`, editor backups and the stash list all
had nothing. I told that session inside the minute, naming the exact symbols and figures lost
so they could re-type rather than rediscover, and they had it back in minutes. **`git stash`
is blocked by a hook for exactly this blast radius; the one-path form is not, and no hook can
see it because it destroys work before any commit exists.** Landmine written, `WRONG_CALLS`
row written. The replacement habit, which costs nothing: snapshot to scratch immediately
before the mutation and `cp` back — restore from **my** snapshot, never from git.

**A second, gentler passenger event in the same file, in the other direction.** While I was
re-applying my change, that session committed — and their pathspec took my heal with it
(`25df3a19c`). Nothing lost, forward-only holds, and they told me rather than leaving me to
find it. They also caught a real error in my block that `gofmt` flagged: I had inserted my
function **between `recordValidationRejection`'s doc comment and its function**, orphaning
the comment. Moved the whole block above the section banner instead.

**Mutation results, all run rather than asserted:** removing `AND is_active = true` fails both
heal tests; removing the heal call fails both; re-inlining a separate sort in the guard
message fails the one-vocabulary test. Migration 565's double-apply refusal was **induced**
after applying, not asserted — re-running the file returns *"the source-vocabulary block is
already present"*.

**Migration numbering churn worth recording:** I planned 561, found the 345 lane had it,
agreed 562 with them directly — and by the time the file was written 562, 563 and 564 had all
been taken by other lanes. Landed on **565**. On a tree this busy, pick the number at the
moment you write the file, not at the moment you plan it.

## 2026-08-23 — the repair is blocked by something that is NOT this bug: loanzy.uk is not being dispatched at all

The component stored (12:31:38Z) and three `needs_page` re-renders were filed at 12:32Z. Forty
minutes later all three are still `triaged`, never claimed, `attempt_count = 0`, no error.

**This is not queue latency and it is not this bug.** Measured [2026-08-23 13:0xZ]:

- **12 other sites were served by `build-dispatch-loop` in the same hour**, several 15–59 times
  (`dartsonline.com` 59, `ai-agent-orchestration.com` 18, `apis.uk` 17, `robot-hands.com` 17…).
  **loanzy.uk: zero.** Its last touch of any kind is 12:33:13Z; 18 items sit `triaged`.
- Ruled out, each with a check rather than a guess:
  - **not a site lock** — `sites.locked_at`/`locked_by` are NULL, and `status`/`build_status`
    (`deployed`/`pending`) are identical to sites that ARE being served;
  - **not head-of-line blocking by a claimed item** — loanzy has none `claimed`;
  - **not a dead handler** — `page-rerender` has 4,214 completions and `page-build-handler` 262;
    `remortgagecalculator.uk` completed a `needs_page` at 12:39Z;
  - **not the known dropped-row scan failure** that `load_work_item_actions.go:829-853` warns
    about (site selected as dispatchable, every row dropped on scan, loop claims nothing) —
    **zero** `work item row DROPPED on scan error` lines in 60 minutes of chassis logs. That
    hypothesis was the best fit and it is refuted.
- **The 15 items ahead of mine were filed by `component-template-fixer` at 12:32:25–12:33:02Z —
  i.e. triggered by my own component store 47 seconds earlier.** They are not the cause of the
  starvation (loanzy was already not being served), but they do mean loanzy's queue is now 18
  deep and my three re-renders are at the back of it.
- `find_dispatchable_site`, which `RUNBOOK_311_fix.md` and the `single-page-deploy` memory both
  describe as the selector, **no longer exists as a DB function** — only a comment in
  `load_work_item_actions.go:829` still names it. The runbook's model of the dispatcher is
  stale, and anyone debugging dispatch from it will be reading about a function that is gone.

**Not filed as a bug yet, deliberately.** `bugs_closed/029` ("hung spawns saturate dispatch
group and halt builds fleetwide") is the closest known mechanism and its recorded behaviour
self-heals in ~40 minutes — which is exactly the window observed, so filing now would risk
recording latency as a defect. A watch is running. **If loanzy is still unserved well past
that window it is a real starvation defect and should be filed against the dispatch lane, with
the 12-sites-vs-zero census above as the evidence.**

**What this does NOT change:** the 337 fix is live, council-APPROVED, and demand-proven at the
mechanism, and the component this bug is about is stored. The three pages remain unrepaired for
a reason that lives downstream of anything this lane changed.

> **RESOLVED 2026-08-23 13:12:42Z — it was the self-heal window, NOT a defect, and the decision
> not to file was the right one.** loanzy resumed dispatching after **~40 minutes** of complete
> starvation, which is precisely `bugs_closed/029`'s recorded self-heal behaviour. Nothing was
> filed and nothing needs to be.
>
> **The transferable half is the near-miss, not the outcome.** I had a measured, genuinely
> striking census — *12 sites served in an hour, several 15–59 times, loanzy zero* — with four
> plausible causes ruled out by check rather than guess. That is a filing-grade evidence
> package, and it was **describing normal behaviour**. What stopped it becoming a false bug
> report was not the evidence (which was correct) but a single question asked before writing:
> *how long does the closest known mechanism take to clear itself?* Forty minutes. The
> observation window was forty minutes.
>
> **So: before filing a starvation/stall/never-runs claim, get the self-heal interval of the
> nearest known mechanism and check your observation window is longer than it.** An hour of
> zero is not evidence of a defect if the known defect clears in forty minutes. The census
> stays in this file because it is a good measurement — it just does not mean what it looked
> like it meant.

---

## 2026-08-23 (evening) — I picked this lane up to work 253, and instead found that both live items in my own handoff were false. Two of my three near-misses today were wrong theories I nearly wrote down.

### What I set out to do, and why it stopped immediately

The handoff (17:15Z) sent me to §3b: *determine what empties `hero-tool`'s `content_data`
values*, starting with "find a page where it happened and diff `content_data` across history".
I did that first, before anything else, and the diff refused to show an emptying.

### Misstep 1 (mine, caught in ~4 minutes) — I read a shared component as the WRONG component

Listing the tool components on loanzy I saw `tool-eligibility-checker` and
`tool-credit-health-check` both pointing at function **`loans-credit-health-check`**, and
`tool-is-a-loan-right-for-me` pointing at **`loans-damage-checker`** — a name with nothing to do
with loans-vs-savings, borrowed from `loanandmortgagecalculator.co.uk`. I was one step from
filing "two loanzy pages serve the wrong tool".

**What stopped me:** the estate rule that shared actions/components are *design*, not a smell
(owner 08-14) — so the discriminating check is the **content**, not the function name. Dumping
both pages' `content_data` settled it in one query: eligibility asks *"Your income and regular
outgoings" / "The loan you're considering" / "Your eligibility snapshot"*; credit asks *"Payment
history" / "Credit utilisation" / "Length of credit history"*. And `loans-damage-checker` on
"is a loan right for me" is a correct affordability calculator (*take-home pay, existing debt
repayments, essential costs, the repayment you're considering, what you'd have left over*) —
**a badly-named template doing exactly the right job.**

**The cheap check that would have saved the detour:** read the `content_data`, not the
`function`. A shared template is identified by its function; what it *is* on this page is its
content. I had the query written before I had the suspicion — I just ran the wrong one first.

### Misstep 2 (mine, and this one nearly re-opened a closed bug) — a probe whose POSITIVE control also failed

Verifying the 337 fix was still live after the 16:03Z roll, the `build provenance` log line had
scrolled past `--tail=3000` (as CLAUDE.md warns for `agent-chassis` specifically). So I fell
back to the binary probe — and ran both controls, which is the only reason this is a note and
not an incident:

```
grep -ac 2dbe12f1d5a1…  /proc/1/exe   -> 0     # must be PRESENT
grep -ac 013d8040…      /proc/1/exe   -> 0     # control, must be ABSENT
```

**Both zero. The positive control failed, so the instrument is blind** — and a blind instrument
reporting 0 is indistinguishable from a true absence. Read at face value it says *"the 337 fix
is not deployed"*, which would have re-opened a bug I was about to close. **Discarded the
measurement rather than the conclusion.**

What answered it: **probe the capability, not the commit.** `aspect_paths` is the field the fix
adds to the advisory; it reads **10,292 chars on a row 20 minutes old**, on pods started 16:03Z
running v1.0.1330. That is "live *now*", which is strictly more than "was built from the right
sha".

### The finding that cancelled my own §3b

All 40 empty values across 11 loanzy pages are `stat_*` keys — *no other key is ever empty* —
and they empty in label/value pairs, so non-empty 11/9/7/5 is just 3/2/1/0 of three optional
stat slots filled. Then the direction check killed it outright: across successive writes the
stat count goes **0→3, 0→0→3, 0→2** and only once **1→0**. **A mechanism that empties values
cannot fill them.** There is no writer to census.

**And the refutation was already in `bugs_open/253`, immediately above my own contribution.**
The 305 lane wrote it on 08-22 — *"per-run variance in generated output, not a settled renderer
change"* — including a note that its own error had been letting three agreeing samples outvote
one disagreeing one. **I then wrote the same generalisation from five agreeing samples the next
day, in the same file, below the correction.** Reading a file is not reading the corrections in
it. That is the misstep worth keeping from today.

### And the page was never blocked

Last floor refusal for `tool-credit-health-check`: **14:03:06Z**. Successful save of all four
slots: **14:23:29Z**, `build_status='deployed'`. The handoff declaring it blocked was written at
**17:15Z** — 2h50m *after* the repair. The live page serves 200 / 30,756 B with 18 function refs,
9 instance-scoped ids and 13 buttons: the same repaired profile the bug file already certifies
for the sibling page.

**The general shape, and why it is now trap #4 in the handoff:** the refusal was a *retryable*
failure, and a snapshot of a retryable failure is indistinguishable in a status table from a
permanent blocker. Nothing in the row says "this will probably clear on its own in 20 minutes".
The lane had even been told this once already — `bugs_closed/029` self-heals at ~40 minutes, and
that trap is written in this same handoff.

### Disposition

- `bugs_open/337` → **`bugs_closed/337`**, with a close-out section and the demand control.
- `bugs_open/253` → correction filed at the foot, cancelling one open question and adding none.
  I explicitly did **not** propose the "measure non-empty fields instead of classes" change I
  had floated: on this evidence it would refuse the same saves for a better-worded reason, and
  that is not worth a shared-guard contract change across nine writers.
- §3c (does Arm A raise the orphan rate?) **run, not carried**: 0/48 pre-fix vs 1/10 post-fix,
  which decides nothing at n=1. Recorded with the sample size that *would* decide it, so the
  next session can skip it knowingly instead of re-discovering that it is underpowered.

---

## 2026-08-24 — closing the 345 loop surfaced a wrong call of mine from yesterday: the build-tripwire I gave them was blind, and I had the disconfirming rows in my own query when I vouched for it

Asked to confirm nothing was still open with the 345 lane, I re-checked the tripwire they had
recorded on my word before sending a closing message — and it cannot fire. Full account:
`bugs_closed/345` (dated correction at the foot), `WRONG_CALLS.md` 2026-08-24, commit
`1d693b72e`. The compressed version for this lane's record:

- The tripwire watched the two **store-side** check literals — which have genuinely never fired —
  and missed the class's real face: `execute_llm_prompt` failing at `generate_template` with
  `response truncated: stop_reason=max_tokens`. That face fired 08-19 00:24Z and its record sat
  in **both** swept tables when we published 0/0.
- Both of the class's work items died in **my own 12:33:13Z supersede batch** on 08-23 — the
  truncation population's drain was 100% administrative, 0% success. (Their file half-knew this;
  now fully attributed.)
- **My compounding error**: I told them "my independent sweep agrees — the only truncat hits are
  RENDER_AUDIT_TRUNCATED", while the generate_template truncation rows sat in my own result set
  under `error_code=UNKNOWN`. I summarised a sweep by the rows I recognised. That is the same
  failure the 305 lane named in `bugs_open/253` — a disconfirming sample present in my own data —
  which I have now committed **twice in one week**, once by writing below their correction and
  once here. The check that stops it: classify every distinct `(error_code, step)` in a sweep
  result before writing "the only".
- The subtler half, pushed to memory (`a-post-fix-zero-needs-a-demand-control.md` §5): their
  demand control validated the **channel** (14 rows of the :477 shape in the column), never the
  **predicate** (the ILIKE patterns). The one-row test — run the instrument's own WHERE against a
  known-real occurrence — fails instantly and costs one query.
- Also narrowed the closed file's "reversal is an hour" claim: moving the recorder wires the
  store-side face only; the firing face needs a writer at the step-failure path.

**The don't-build decision stands** on the corrected baseline (dormant since 08-19, with the 337
work in between). WII-026's relations line corrected in step. Messaged the 345 lane with the
one-row test as their re-verification handle; nothing open between the lanes.
