# RFC 006 — a shared step whose effect is "take everything in state X" needs one owner, or every parent that branches on it is a latent `bugs_open/150`

**Status: RAISED 2026-07-31 by the bugfix_150 lane. Needs an owner decision.**
Raised because the council's `architecture` seat asked for it explicitly, on a
submission it otherwise approved (`757cc7be`, *"approved with 8 advisory
objection(s) — none high-severity"*):

> *"Recommend: proceed with this patch (correct, safe, tested), but open the RFC for
> 'one triage, one owner' rather than letting the next occurrence surface as another
> bespoke site-scoped field on this same action."*
> — `ARCHITECTURE_SIGNAL: needs_rfc`

The `bug_historian` seat reached the same place independently in the same round, and
named the shape: *"the exact shape of 016b §9's 'one call site of a shared judgement
gets the rigorous fix; the sibling stays heuristic'."* **Two seats, arriving from
different directions, on a patch neither of them opposed.** That agreement is the
reason this is an RFC rather than a note in a bug file.

Filed under the 2026-07-29 owner ruling: *review here is after the fact, by design* —
the patch is committed (`337fdd9af`) because HEAD is shared and holding code back is
not a capability this tree has. This RFC is about the class, not that patch.

---

## 1. The shape, stated generally

An action whose effect is **"take everything currently in state X"** has two
properties that are individually reasonable and jointly a trap:

- it is **idempotent by emptying** — a second run is not a failure, it reports success
  and zero;
- its result describes **what that call did**, which every workflow language then makes
  trivially available to a branch, as if it described **what is now true**.

Those two readings coincide only while exactly one caller exists. Add a second and the
loser reports an honest zero that the branch reads as "nothing to do". Nothing errors,
nothing logs, and the failure is a *correct value routed to the wrong branch*.

**The live instance.** `triage_detected_items` promotes `detected` → `triaged` for a
whole site (`WHERE site_id = $1 AND status = 'detected'` — no type filter, no ownership
filter) and is a step in **three** live agents: `improvement-loop`
(`triage_findings`), `design-audit-agent` (`triage`), `site-review-agent` (`triage`).
The improvement loop calls both children first, so its own copy always reports
`promoted: 0`, and its `check_has_findings` branch terminated the run on *"No issues
found — site is clean"*. Measured twice: orchestration `30692439` (67 findings) and
`911ecdd8` (27 findings, fired 2026-07-31 as a pre-fix control).

## 2. What was done, and what it deliberately did not do

`bugs_open/150` was fixed by making the shared action **also** return the site-scoped
answer (`site_dispatchable` / `site_dispatchable_count`), and pointing the one live
branch at it. That is the right fix for the bug: the decision no longer depends on
which copy ran, in what order, or on a fourth caller that does not exist yet.

**It does not close the class**, and the submission said so. The `architecture` seat's
statement of the residual is better than mine and is quoted rather than paraphrased:

> *"the identical defect class … stays reachable for the NEXT agent that adds a triage
> step, and now that agent inherits two candidate signals (`has_items`,
> `site_dispatchable`) with an undocumented-outside-this-diff scoping rule to pick
> between."*

That second clause is the part I would not have written and is the real cost: the fix
adds a **choice** where there was previously only a wrong answer, and the guidance for
making that choice currently lives in one file's header comment and one register entry.

## 3. The question for the owner

**Should a shared "take everything in state X" step have exactly one owning caller?**

Three answers are available and they are genuinely different:

**(a) One promoter, one owner — the structural fix.** Remove `triage` from
`design-audit-agent` and `site-review-agent`; the parent owns promotion. Then
`triage_result` describes the run, and the ambiguity disappears rather than being
documented.
*Cost, as stated on 2026-07-31:* both children are independently callable — every other
parent of either agent must be audited first, and a child fired on its own would then
promote nothing, which may be a behaviour change somebody depends on.
> **UPDATED 2026-08-02 after running the census in §5: that cost is now measured, and it
> is near zero.** There is no other parent (the definition scan returns two rows, both
> `improvement-loop`) and no standalone run of either child has ever been recorded. The
> audit that was the barrier is *done*, and it found nothing to audit. What remains is two
> `jsonb` step deletions and their re-wiring — a config-only change, reversible by
> `snapshot_agent`, with the caveat in §5 attached.
*Residual:* none for this class, once done.

**(b) Keep the fan-out; make the site-scoped signal the default and the call-scoped one
opt-in.** Invert which is easy to reach: a branch that names `has_items` on a *promoting*
step is almost always a bug, so make it the one you have to ask for.
*Cost:* touches the fleet-wide `has_items` convention (four live consumers across three
actions, three of them correct), which is exactly the shared-vocabulary change the
2026-07-29 ruling says needs this venue.
*Residual:* the two names still coexist; the trap is narrowed, not removed.

**(c) Leave it, and write the rule down properly** — a documented convention that a
shared step returns both, and that branches read the state-scoped one.
*Cost:* nearly nil.
*Residual:* the whole of it. This is the status quo plus a paragraph, and the honest
argument for it is that one instance in the fleet's history is a thin basis for a
refactor with an unmeasured blast radius.

## 4. What is NOT being asked

Not asking for a verdict on `bugs_open/150`'s patch. It is approved, committed, tested,
and its config half is held back pending the image. If (a) is later adopted, the
site-scoped field remains correct and simply stops being load-bearing.

## 5. Evidence a decision should rest on, and what is missing

**Measured, 2026-07-31:**
- three live agents carry the step; only `improvement-loop` branches on its result
  (`design-audit-agent` and `site-review-agent` both go `triage → complete`) — so the
  blast radius of (a) is *today* one branch, not three;
- all three leave `target_pipeline` unset, so all three promote to the same pipeline;
- four live conditions read a `has_items`, across three actions; three are correct;
- **0** live agent definitions declare `has_items` in an `output_contract` — so the
  `guidelines` seat's DECLARED CONTRACTS rule is unenforced for the existing key too,
  which is its own small finding and is relevant to (b).

~~**Not measured, and it decides (a):**~~ **MEASURED 2026-08-02** — the census this section
called for has now been run, at the owner's request, before the decision rather than after.
It comes back decisively in favour of (a) being cheap:

- **No other parent calls either child.** A scan of every live agent definition for a step
  naming `design-audit-agent` or `site-review-agent` returns **exactly two rows**, both
  `improvement-loop` (`spawn_design_audit`, `spawn_site_review`). There is no second parent
  to audit.
- **Neither child has ever been fired standalone.** `agent_run_stats` is a cumulative
  counter (78 agents tracked, oldest entry 2026-07-26) and reads
  **`improvement-loop` 3 / `design-audit-agent` 3 / `site-review-agent` 3**, identical
  counts over the identical window (2026-07-29 → 2026-08-01). A standalone child run would
  break that 1:1:1. Corroborated independently in `orchestration_state_audit` (3-day
  retention, identified by each agent's distinctive step name): 2 / 2 / 2.
- **No scheduler fires either child**: 0 rows in `scheduled_tasks` naming either type;
  `orchestration_requests` is empty.
- **Neither child branches on its own triage result** — both are
  `triage → complete` (live config), so removing the step changes no control flow inside
  them; it changes only whether promotion happens on that path.

`[CAVEAT — stated because it is the one soft edge]` The run counters begin 2026-07-26 and
the whole fleet's improvement-loop history is 3 runs, because `improvement-sweep` is
disabled (`bugs_open/083`) and every run so far was hand-fired by a session. So this is
**"no standalone caller exists today"**, firmly; it is not "no standalone caller has ever
existed" over the agents' whole life. The design question it answers — *is there a second
consumer whose behaviour we would change?* — is answered no by the definition scan alone,
which does not depend on run history at all.

## 6. Related

- `bugs_open/150` — the instance, and the patch.
- `bugs_open/171` — the *other* false-clean route in the same orchestrator
  (`check_audit_pass_limit`), filed the same day at the `bug_historian` seat's
  insistence. Different mechanism, same family: a terminal step asserting something the
  run never checked.
- Register `WDS-015` — the shipped seam, with the scoping rule and both landmines.
- 016b §9 *"One responsibility implemented in three agents"* + its 2026-07-31 addendum —
  the pattern write-up, including the two fixes that look obvious and are worse.
- `bugs_closed/124` — the precedent the `architecture` seat cited: a reserved-key
  addition to a shared mechanism arriving inside a bug patch. Same shape, and the reason
  this file exists rather than a paragraph in a commit message.
