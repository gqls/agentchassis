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
