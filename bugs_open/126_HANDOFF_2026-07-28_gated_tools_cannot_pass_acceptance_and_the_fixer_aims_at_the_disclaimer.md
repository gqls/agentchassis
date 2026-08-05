# 126 — a tool with a consent gate cannot pass acceptance, and the auto-fixer is then aimed at its disclaimer

**Filed** 2026-07-28 from the oufe.com workstream, after the first Tier-4 run of
`tool-recovery-waterfall` failed 2 of 11 checks.
**Severity** medium now, high latently. Nothing is currently broken in
production. The reason to file it is the second half: the failure path routes an
automated code-rewriting agent at a legally load-bearing disclaimer, and it does
so silently.
**Status** OPEN. The one affected tool was fixed at the criteria level and now
passes 13/13; the platform behaviour is unchanged and will repeat on the next
gated tool.

## The mechanism

`internal/adapters/browserrunner/run_checks_action.go` runs a profile like this:

- `:563` — one browser context per profile
- `:569` — **one page**, created once
- `:584` — **navigated once**
- `:398` — then *every* check in the fence runs against that same page

So state accumulates across checks. That is fine for `selector_exists` and
`page_status_ok`, and it is a trap for `interaction` checks, because nothing
resets the page between them.

Now add a consent gate — the pattern the disclaimer doctrine actively pushes
toward, where a tool is unusable until the reader acknowledges it can be wrong:

```js
accept.addEventListener('click', function () {
  gate.hidden = true;
  gate.style.display = 'none';   // one-shot, by design
  body.hidden = false;
});
```

The first interaction check clicks the gate and passes. **Every later check that
clicks the gate fails**, because the element still resolves in the DOM but is no
longer visible. Playwright says so precisely, and the message is easy to misread
as a broken button:

```
locator resolved to <button type="button" id="rw-accept" class="rw-accept">…</button>
  - attempting click action
    2 × waiting for element to be visible, enabled and stable
      - element is not visible
```

Read quickly, that looks like "the tool's button is broken". It is not. The
button works; it has already done its job.

## Why this is worth a bug rather than a note

A failing acceptance run raises an `improve_tool` work item **carrying the fence
as `acceptance_test`**, handled by `tool-improver`, and the build loop dispatches
it. The fixer's task is therefore: *make this tool pass a test that requires the
consent gate to be clickable twice on one page load.*

The only ways to satisfy that are to stop the gate hiding itself, to make it
re-appear, or to remove it. **All three weaken or delete the disclaimer** — which
for this site is section B of the owner-approved wording and the proximate
placement that the negligent-misstatement analysis (Hedley Byrne) depends on. It
is the single element on the tool page that most needs to be immovable.

Nothing in the loop knows that. There is no notion of a protected region in a
tool, so an LLM rewriting for green checks has no reason to preserve it, and a
green run afterwards would look like success.

On this occasion the item was cancelled by hand (`wont_fix`, with the reasoning
in `resolution_path`) before it dispatched. That was luck and attention, not a
guard.

## Evidence

| what | value |
|---|---|
| failing run | 2 of 11 checks, both profiles, `zero-ev-breaks-at-top` at step 1 |
| after fixing the fence only — tool source unchanged | **13 of 13 passed**, both profiles |
| work item raised by the failure | `acceptance_fail:tool-recovery-waterfall:a0d7f1ae-…`, `item_type=improve_tool`, status `detected` |

The tool source was never edited between the two runs. That is what establishes
the fence, not the tool, as the defect.

## Fix candidates, ordered by what closes the door

1. **Give the runner a way to reset between interaction checks** — a `reload` or
   `navigate` step action (`Do()` at `:768` currently supports only `fill`,
   `click`, `select`), or an opt-in `fresh_page: true` on a check. This makes the
   coupling *unrepresentable* rather than merely documented: each interaction
   check becomes self-contained and order stops being load-bearing. Cheapest real
   fix — it is one `case` arm plus a page reset.
2. **Teach the tool-improver that some markup is protected.** A `<!-- acceptance:
   protected -->` region, or a rule that a tool carrying a consent gate never
   auto-dispatches and goes to human review instead. This is the half that
   matters even after 1 lands, because a *genuine* failure on a gated tool still
   points a rewriter at the disclaimer.
3. **Validate the fence when it is seeded** — reject a criteria set where more
   than one interaction check clicks the same selector, since on a shared page
   that is provably unsatisfiable. Catches the author error at write time with a
   comprehensible message instead of a 30-second Playwright timeout.
4. **Document the shared page in the criteria authoring guide.** Necessary, and
   on its own worth little: this fence was written by someone who had read the
   guide and still assumed per-check isolation.

1 and 2 are complements, not alternatives. 1 stops the false failure; 2 stops the
false failure from being *repaired into a real one*.

## How to verify a fix

Build a tool with a one-shot gate and a fence with two interaction checks that
both click it. Before the fix it must fail the second; after candidate 1 it must
pass both. Then make the tool genuinely broken (break the arithmetic, leave the
gate intact) and confirm the run still fails — **a fix that makes gated tools
always pass is worse than the bug**, because acceptance is the only tier that
tests behaviour rather than markup.

For candidate 2, confirm a failing gated tool routes to human review and that the
gate markup is byte-identical afterwards.

## The transferable lesson

Markup being present is not evidence a tool works, and *this* is the class of
thing that only a real run finds. But the sharper lesson is the second-order one:
**an automated repair loop inherits the authority of whatever test it is given.**
A wrong test does not merely fail — it becomes a specification, and the loop will
faithfully damage correct code to satisfy it. Any check whose failure triggers an
automated rewrite needs to be treated as a piece of production configuration,
because that is what it is.

## Related

- `docs024_key_docs_latest/oufe/` — the tool, the corrected fence (the superseded
  `doc_plans` row preserves the wrong one), and the NOTES entry.
- `bugs_closed/012` — an improver truncating and destroying a component. Same
  family: a rewriting agent with insufficient constraints on what it may not
  touch.
- `bugs_closed/024` — tool fixes never reaching the live page. The inverse
  failure: the loop being unable to act, rather than acting wrongly.

---

## CONTRIBUTION 2026-07-28 (086 per-handler audit) — `note_refusal`'s error handler is disabled, and the decision is yours

Not a change; a handoff of one finding into the workstream that owns this path.
`scripts/who-owns.py 126` names oufe as owner (ACTIVE, 50 commits/14d), so this thread
read the path and stopped.

**State.** `tool-improver.note_refusal` currently has **no live error handler**. It
declared `error_step: complete`, and seed 220 renamed that key to
`error_step_disabled_086` on 2026-07-26 as containment for `bugs_closed/086` (the plan
converter had never copied step-level `error_step`, so arming the fix armed 55 handlers
at once; ten routed to `complete` and were disabled pending per-handler review). Verified
live 2026-07-28: the step carries `error_step_disabled_086: "complete"` and no
`error_step`.

**Why it lands on you rather than on 086.** The refusal branch is 126's subject matter:

```
refuse_mangled_write  action=fail_work_item     next=note_refusal  err=note_refusal
note_refusal          action=append_doc_note    next=complete      err=complete (DISABLED)
append_note           action=append_doc_note    next=complete      err=-        (never had one)
```

**The two readings, both defensible — which is why it is a call and not a fix.**

- *Leave it disabled.* `append_note`, the success-path twin with the identical action,
  has never had an `error_step`. As things stand the two branches behave the same way
  (a failed note fails the run); re-enabling `note_refusal` alone makes them diverge.
- *Re-enable it.* `next_step` and `error_step` were **both `complete`** — the handler
  drew no distinction whatever, so the author's intent reads as "notes are best-effort".
  And it sits on the branch where the outcome is already settled: the work item has been
  failed, the refusal has happened. Failing the orchestration because the *note* about
  the refusal could not be appended replaces a clear signal with a generic one.

**The consideration that actually bites for 126.** The note is the branch's product —
*"record the refusal on the tool's travelling NOTES so the next agent…"*. If it fails and
the run dies loudly, the refusal still happened but **the next agent is not told why**,
which is close to the failure mode 126 is already about: a fixer aimed at the wrong
target for want of the record. Whichever way you rule, that is the thing to weigh.

**Revert is one rename**, and the pre-change snapshot is real — `agent_definitions_backup`,
`snapshot_reason LIKE '220_%'`, `2026-07-26 18:32:26.229Z`, `type='tool-improver'`.
Seed 220 carries the exact revert SQL in its header comment.

**Unrelated but noticed in the same pass, and it is yours:** `tool-improver.update_component`
**gained** a step-level `error_step → refuse_mangled_write` since 07-26 (`6e29d6d19`).
That is now a live handler, and step-level handlers only began working at all in
v1.0.1169 — so it has real routing behaviour that predates nothing. Worth a deliberate
test of its failing branch rather than assuming it.

### DECISION 2026-07-29 (oufe lane, as named owner): `note_refusal` STAYS DISABLED

Ruling on the contribution above: **leave it as it is** — no live `error_step`, a
failed note-append fails the run.

Reasoning, weighed against the consideration the contributor flagged:

- **The note is missing in BOTH outcomes.** The handler cannot save the note; it
  only chooses whether the run then reports COMPLETED or FAILED. So the real
  question is which report serves the next agent when the record is absent.
- **A quiet COMPLETED with a silently missing note is a success-shaped
  non-event** — this workstream's most-paid-for failure class (095's
  `complete_skipped`, the partial-build loop found this same morning, "complete
  is not proof the work happened"). A loud FAILED at least produces a failure
  row the immune system sweeps; a quiet completion buries the gap where nothing
  looks.
- **Symmetry**: `append_note`, the success-path twin running the identical
  action, has never had a handler. Re-enabling one branch alone makes twins
  diverge for no stated reason.
- The original author's `error_step: complete` read ("notes are best-effort") is
  defensible but predates step-level handlers working at all (v1.0.1169), so it
  was never an exercised design — there is no behaviour to preserve.

Revert path if this proves wrong is unchanged: seed 220's header carries the
rename-back SQL; snapshot `agent_definitions_backup` `snapshot_reason LIKE
'220_%'` (2026-07-26 18:32:26Z).

**Still owed from the same contribution (open):** a deliberate failing-branch
test of `tool-improver.update_component`'s now-live
`error_step → refuse_mangled_write` (gained `6e29d6d19`, live only since
step-level handlers work). Verify-the-failing-branch applies: induce a mangled
write in a sandbox item rather than assuming the route.

---

## FIX SHIPPED IN CODE 2026-08-05 — STILL OPEN, awaiting the fleet roll

**Status: candidates 1 and 2 (the platform mechanism) are built, tested and
committed. Bar per CLAUDE.md/`bugs_closed/README.md` is "fixed AND live" — this
stays OPEN until `browser-runner-adapter` and `agent-chassis` both roll and are
pod-verified.** Candidate 3 (write-time rejection) was evaluated and refuted
(below); candidate 4 (docs) partly done via this session's concept-register
entry, the criteria-authoring doc update is still owed.

### What shipped

Commit `67a4c50bd` (`087_towards_multiple_domains`), council-submitted
(`Council-Submitted: 479d747e-97a7-47d3-9c15-ccce0ee18014` — verdict not yet
read at commit time; do not add `Council-Reviewed:` until it is).

1. **`reload` criteriaStep action** — `internal/adapters/browserrunner/run_checks_action.go`,
   `chromiumPage.Do`. An interaction check whose `steps` begin
   `{"action":"reload"}` resets the shared page to its landing state before the
   rest of that check's steps run, so a one-shot consent gate is clickable
   again. Deliberately does not reset `status`/`navErr` (that is the ORIGINAL
   navigation's contract) or accumulated console errors (`no_console_errors`
   must see everything any interaction triggered across the whole page
   lifetime). `platform/orchestration/actions/experience_criteria.go`'s
   `experienceStepActions` table updated in the same commit — REQUIRED, not
   optional: `TestExperienceCheckCapabilities_LockstepWithCheckers` reads the
   runner's source as ground truth (its two named anchor functions,
   `applicableChecks` and `func (p *chromiumPage) Do(`, no longer match the
   real names `splitByProfile`/receiver `c`, so it silently falls back to
   whole-file scanning) and fails by name if the two drift. Induced the failure
   and reverted it to prove this before shipping.
2. **`no_auto_fix` (+ `no_auto_fix_reason`) fence flag** —
   `platform/orchestration/actions/tool_acceptance_actions.go`,
   `JudgeAcceptanceResultsAction`. A fence carrying it routes a FAILING verdict
   straight to the existing `acceptance_stuck` human-review escalation (same
   `item_type`, same `item_key` shape `acceptance_stuck:<function>:<siteID>`,
   same dedup/spec-merge machinery already built for the cycle-count case)
   instead of ever reaching `tool-improver`, regardless of cycle count.
   `parseNoAutoFix` fails OPEN: absent or malformed criteria mean today's
   behaviour, unchanged — proven by mutation-testing a fail-closed variant,
   which breaks three tests. This is candidate 2, and it was cheaper than this
   file originally assumed: the human-review escalation machinery already
   existed for the cycle-count case; the change is a flag parse plus widening
   one condition, not new design.

### Candidate 3 (write-time rejection) — evaluated and refuted, not built

This file's candidate 3 assumed "more than one interaction check clicking the
same selector on a shared page is provably unsatisfiable." That is false in
general: repeated clicks on an ordinary, still-visible button are perfectly
satisfiable (the runner clicks visible elements repeatedly without issue).
Unsatisfiability holds only for a ONE-SHOT element, and one-shot-ness is a
runtime property — nothing in a static criteria document declares it. A
validator built on the false premise would reject legitimate fences. Folded
into the `reload` fix's documentation instead (a later check simply needs to
lead with a reset).

### Verification done this session

- `go build ./...` clean; `go test ./internal/adapters/browserrunner/...
  ./platform/orchestration/actions/...` all green.
- **Real headless Chromium** (`BROWSER_RUNNER_IT=1`), against a local
  `httptest`-served fixture reproducing the bug's exact one-shot-gate JS:
  measured `#rw-accept` at `0x0` after one click (Playwright's own
  "not visible" state) and reproduced the bug's literal error text
  (`"2 × waiting for element to be visible, enabled and stable - element is not
  visible"`) when clicked again without a reload; a check leading with
  `{"action":"reload"}` then passed. A control fence against a genuinely broken
  tool (gate intact, wrong computed output) still FAILS — a fix that makes
  every gated tool pass unconditionally would be worse than the bug, per this
  file's own "how to verify a fix" section.
- Mutation-tested both guards: reverting the `stuck || noAutoFix` condition to
  `stuck` alone, and reverting `parseNoAutoFix` to fail-closed, each break the
  tests written against them.

### Owed before this can close

1. **The roll.** `make build-agent-chassis` and `make build-browser-runner-adapter`
   both build clean from this commit (verified locally, images not pushed).
   Per this repo's own `releases-are-whole-fleet-make-release` practice, a
   single-service build+push+deploy at its own tag previously fragmented the
   fleet and was corrected by the owner — this needs the whole-fleet release
   (`make release redeploy-agents ENVIRONMENT=production REGION=uk001`), not a
   one-service `kubectl apply -k`.
2. **Pod-verify both images** with `grep -ac` (the fleet's images carry no
   `strings` binary) + a positive control, every replica, once rolled — e.g.
   `grep -ac "reload navigation failed"` in `browser-runner-adapter` and a
   literal from the judge's new escalation wording in `agent-chassis`.
3. **Read the council verdict** on `479d747e-97a7-47d3-9c15-ccce0ee18014` and
   either add `Council-Reviewed:` on a follow-up commit (APPROVED) or resubmit
   with the objections answered (REVISE/REJECTED) — never add the trailer
   unread.
4. **The criteria-authoring documentation** (`docs024_key_docs_latest/tools/
   tool_acceptance_runner/PLAN_tool_acceptance_runner.md` §"Criteria contract")
   — add `reload` to the step verbs, the shared-page-state paragraph, the
   `runDeadline` cost note, and the `no_auto_fix`/`no_auto_fix_reason` keys.
   Not yet done.
5. Neither key has been adopted by a real tool's PLAN yet (the tool that
   originally hit this bug was already fixed at the criteria-authoring level,
   per the top of this file) — this ships the mechanism, not a retrofit.

Registered: `docs/agent_docs/docs026_concept_register/register/tool-lifecycle.md`
TL-040.
