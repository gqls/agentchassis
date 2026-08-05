# PLAN — actioning the 2026-08-05 code-review findings

Decisions and their reasons. Corrections to the originating brief
(`HANDOFF_2026-08-05_code_review_triage.md`) live here, marked as corrections.

---

## The decision that reopened the work

The handoff's stated policy was **"No fixes. Every cluster belongs to an active lane."** That
was correct when written and false an hour later: 194 and 195 closed at 11:01, and neither
lane's continue-here handoff mentions the review. So the policy's premise — that fixing would
be *competing* — no longer held for 10 of 15 findings.

**Decision: fix the findings belonging to closed lanes; contribute rather than compete only
where a lane is genuinely open (156).** Confirmed per-finding by `ls bugs_open/ bugs_closed/`
and a grep of each lane's directory, not by the handoff's table.

## Corrections to the handoff

> **CORRECTION 1 — the ownership table mis-assigns F11 and F12.** The handoff resolved
> ownership with `git log -1 -- <file>` (its §"How I triaged", step 1). That is
> file-granularity. `save_page_sections_action.go` was last committed by the **156** lane
> (`84b7d561c`), but `git blame -L 624,628` puts the lines both findings cite in the **194**
> lane (`47ee3ebce`), 26 minutes earlier. F11 and F12 are therefore unowned, not 156's, and
> were fixed rather than handed over.

> **CORRECTION 2 — F5 is a FALSE POSITIVE, the review's second after F4.** The handoff rated
> it "judgement call for the lane … defensible if the volume is genuinely low, but the volume
> has not been measured". Measured, the *write-side* facts hold (no dedupe, no rate limit, no
> `site_id`) but the predicted harm does not: `agent_error_log` is reaped hourly by
> `scheduled_tasks.database-cleanup` at 30 days, and the rows are excluded from both scoped
> diagnosis loads. See NOTES §3–4.

> **CORRECTION 3 — F10's remedy is impossible.** "Instead of calling `LogAgentError`" cannot
> be done: `platform/orchestration/coordinator.go:23` imports `actions`, so the reverse import
> is a cycle, and ~20 files in that package each carry a local INSERT for the same reason. The
> handoff repeated the finding's claim that "the package's own precedent forbids the third
> copy"; the precedent is the opposite.

## Decisions taken, and why

**1. Fix comments where the comment IS the defect (F1).** F1's code is correct and the
paragraph six lines above it argues for the behaviour. Only the closing claim is false. A
code change here would have broken a load-bearing early return to satisfy a sentence.

**2. Rename the NEW key, not the live one (F7).** Both directions remove the collision. The
live key is referenced by migration 219 and seeded on two agents, so moving it needs a
migration and a coordinated roll; the new key was seeded on nobody (RFC_010, default OFF), so
renaming it is free and measurable — `0` live `save_page_sections` steps carried it. Ranked by
what closes the door at lowest cost.

**3. Widen F9's predicate to match the DELETE, accepting more records.** The counter exists to
predict what the DELETE destroys, so its scope is defined by the DELETE and nothing else. This
makes a warning-level record fire more often, which is acceptable *because* the table's
retention was measured (§4 of NOTES) rather than assumed — the same measurement that refuted
F5 is what licensed this.

**4. Make F3 visible rather than fixing it.** Changing which path wins is a resolution-semantics
change on the fleet's highest-traffic save path, and it is unreachable today (all three live
callers explicit). Changing it would be authority taken on a prediction — the same reason
`writeContentDataRegressionLog` records instead of refusing. A warning names the ambiguity for
the future caller that would hit it.

**5. Do not fix F13.** Its file holds another session's uncommitted work, and the comment
carrying the wrong anchors is on the `+` side of their diff. Verified and handed back with the
corrected numbers (`:215`→`:235`, `:216`→`:237`) instead.

**6. Do not file F5 as a bug.** A bug file asserting unbounded growth on a demonstrably
bounded table would be a durable false claim. Recorded in NOTES and `WRONG_CALLS.md` instead.
The one real sub-finding it surfaced — `LogAgentError` writes `domain` as `''` where `site_id`
gets `NULLIF`, so 3,189 rows read as "has a domain" when they do not — is recorded, NOT fixed:
changing a shared writer's stored shape is a seam change and is not what F5 asked for.

**7. Three council submissions, not one.** The norm is one run per coherent task. F11/F12 were
discovered after the 194 submission was already dispatched, and were not in its plan, so they
went as their own submission rather than riding a trailer that did not cover them.

## What is deliberately still open

- **F6** — unactionable. Named in the handoff's ownership table, described nowhere, and the
  original `/code-review` output was not saved. Guessing at it would be worse than flagging it.
- **F13** — confirmed, corrected values recorded, owner not yet notified beyond these docs.
- **The `domain = ''` asymmetry** in `LogAgentError` — recorded, unowned, fleet-wide.
- **Two council verdicts** (`cb575682`, `d0d2c97a`) were pending at the time of writing. The
  code is already on the shared branch; a REVISE/REJECTED is still owed a response.
