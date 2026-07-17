# PILOT — "Reuse Agent" council reviewer (stage 3, seat #2 of the extended roster)

**Status: LIVE as of 2026-07-17.** Applied to `clients_db` via
`fixloop_eg_dartsonline/0NN_fix_proposer_v7_reuse_agent.sql`. Pre-flight
checked for in-flight fix-proposer/council orchestrations first (none found —
the council has never yet reached its review steps on a real case). Prior row
snapshotted (same source id as v6, `f9d90a2d-...`, new snapshot reason logged).
Verified live: `review_reuse_agent` present, wired
`review_bug_historian → review_reuse_agent → review_guardian`, both
`review_fields` arrays (`council_decide`, `escalate`) carry all four
reviewers, prompt content intact (2,906 chars). The council is now 4
sequential reviewers.

---

## 1. Why this seat, and a correction on its grounding

This was the first candidate identified (§ "candidate A" in `PLAN_concept_register.md`),
picked alongside the bug-historian on the strength of the register's own
rediscovery-frequency signal. On closer inspection while building this pilot,
one thing needed correcting: I'd cited `tool-lifecycle.md`'s high citation
density as the evidence for this seat. It isn't — that category's most-cited
concepts (`TL-001`, `TL-008`, `TL-014`, `TL-016`) are about tool-clobber
protection and the tool-verification ladder, a real and distinct theme
already folded into the bug-historian's curated context (`TL-001` is one of
its seven cited incidents). The actual charter for a reuse-checking reviewer
is `DEV-001` in `development-guide.md`, and FIX-036 itself names the founding
incident directly: "motivated by a real incident where a chat reinvented a
trigger+triage SQL pair that already existed." That's the real grounding —
noted here so the design record stays honest rather than silently keeping an
inconsistent citation.

## 2. Charter

**The reuse-agent judges one question only: "does this platform already have
something that does this, and did the plan check?"** Not edit quality, not
blast radius, not whether a pattern is recurring — specifically whether the
proposed plan introduces new code, a new table, a new function, or a new
migration where an existing mechanism already covers the need.

**Curated context (v1):**

```
KNOWN DISCIPLINE: "Reuse before create" (STEP ZERO)

This platform has a standing, explicitly-named discipline: before creating
any agent, action, function, or migration, search agent_definitions, the
action registry, Go code, and existing workflows for an equivalent — and
never create without first demonstrating no existing coverage exists.

Documented successes of following it: reused ssh_get_status as a monitor
probe rather than writing a new one; reused ListObjects for resume logic;
reused datahelpers.GetIntField over a custom helper; reused snapshot_agent()
instead of a new side-table migration for backups.

THE FOUNDING INCIDENT for this review seat: a prior session reinvented a
trigger+triage SQL pair that already existed elsewhere in the codebase —
duplicated work that a reuse check would have caught immediately.

A cautionary related case: this platform has at least one place
(tool-lifecycle: "two divergent tool-creation paths") where two different
code paths independently solve overlapping problems (a "novel" path and a
"fork" path for creating tool pages) with inconsistent side effects — not
because either was wrong on its own, but because nobody unified them once
both existed.

WHAT TO LOOK FOR in this plan: (a) does any edit ADD a new function, action,
table, or migration; (b) if so, does the plan's rationale or grounded_in
show evidence a search for an existing equivalent was done; (c) does the
plan's own diagnosis-stage evidence already reference an existing mechanism
that the fix could extend or call instead of duplicating; (d) is this
introducing a second way to do something the platform already has one way
to do (the same shape as the two-divergent-tool-creation-paths case).
```

**Verdicts:** `approve | object`, no `veto` — same advisory design as the
bug-historian, for the same reason (any reviewer's veto rejects outright
regardless of `hard_veto_from`, confirmed in `diagnose_council_decide_action.go`;
an advisory seat must not carry that option).

## 3. Prompt template (matches the existing reviewers' contract)

```
# Council reviewer: REUSE AGENT

You judge one thing: does this platform already have something that does
what this plan is about to build, and did the plan check? You change
nothing; you judge.

{{known discipline + founding incident + cautionary case, as in §2}}

Judge the plan: (a) does any edit ADD a new function, action, table, or
migration; (b) if so, is there evidence in the plan's rationale/grounded_in
that an existing-coverage search was done; (c) does the diagnosis's own
evidence already name an existing mechanism this plan should extend instead
of duplicating; (d) would this create a second way to do something the
platform already has one way to do.

Verdicts: approve (no reuse concern, or additions are genuinely novel),
object (this plan risks duplicating existing coverage — name what already
exists and where, in objections). You do NOT have a veto.

CHECKS: if a verdict hinges on whether a table, column, or action name
already exists, put that query in checks as {"sql": "SELECT ...", "why":
"..."} — SELECT/WITH only. Write checks ONLY against the tables/columns in
the Schema section below.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output — ONLY this JSON
{"reviewer": "reuse_agent", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "..."}], "notes": "..."}
```

## 4. Exact wiring (extends v6, becomes v7)

Chain: `persist_plan → review_editquality → review_bug_historian →
review_reuse_agent (NEW) → review_guardian → council_decide`.

Four edits, same shape as the v6 patch:
1. `review_bug_historian.next_step`: `'review_guardian'` → `'review_reuse_agent'`
2. New step `review_reuse_agent`, `next_step: 'review_guardian'`
3. `council_decide.review_fields` and `escalate.review_fields`: add
   `'review_reuse_agent.result'`
4. `repropose.input_fields` and its prompt: add `review_reuse_agent`

## 5. A scaling concern, surfaced before going further down the list

Going from 3 reviewers (v6) to 4 (this seat) is a modest, proportionate step
— roughly a 33% increase in per-round reviewer latency and LLM spend. Going
on to build all remaining 9 seats from the "next 10" list the same way — as
more always-on steps in the same sequential chain — would mean **12
sequential LLM reviewer calls before every single council decision**,
including every revise/repropose round (capped at `max_rounds`, but each
round re-runs the whole chain). That's roughly a 4x latency/cost increase
over today, on a council that's about to become the fleet-wide gate for
every platform commit (per the concurrent council-gate thread), not just
fix-loop's own runs.

This wasn't a problem worth pausing on for one more seat, but it is worth
raising before committing to nine more the same way. Three ways to go:
(a) build all 9 as more always-on sequential steps regardless, accepting the
latency/cost growth; (b) build a relevance-filtering activation mechanism
first — the "match the fix plan's touched files against each seat's relevant
categories" design already sketched in `PLAN_concept_register.md` §Stage 3 —
so only 2-5 relevant seats fire per run instead of all of them; (c) pace it —
build a couple more of the most broadly-applicable seats (this one, and
perhaps the guidelines agent) as always-on, and treat the narrower specialist
seats as candidates for (b) once it exists. Asked in the accompanying message,
not decided here.

---

## 6. ADDENDUM from the council-gate thread (2026-07-17, later the same day)

*Appended by the "fixloop council on every bugfix" thread — this section adds
two mechanical facts the §4 patch predates; nothing above is altered.*

**(i) There are now TWO council definitions; a seat migration must patch
both.** The council-gate thread built (files only, NOT applied — the owner's
roster ruling gates the launch on more stage-3 seats, i.e. on this very
pilot) an advisory review-service clone of the council:
`docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_council_gate.sql`. Its
reviewer steps are deliberately name-matched to fix-proposer's so seat
patches stay mechanical. The gate-side edits for this seat (same shape as §4,
rationale-context instead of diagnosis-context):
1. `review_bug_historian.next_step` → `'review_reuse_agent'`
2. New step `review_reuse_agent` — swap the prompt's `## The diagnosis` /
   `{{.diagnosis_row.conclusion}}` section for `## The author's stated
   rationale` / `{{.input_data.rationale}}`; `input_fields`
   `['input_data','plan_persisted','schema_hint']`; `error_step`
   `'complete_invalid'`
3. `council_decide.config.review_fields`: append `'review_reuse_agent.result'`
4. `run_checks.config.check_fields`: append `'review_reuse_agent.result.checks'`
While the gate seed is unapplied this is a file edit, not a migration —
cheap now, a silent-drift landmine later (two councils, different rosters —
the exact failure family this seat reviews for).

**(ii) v6's `run_checks.check_fields` omits the bug-historian — do not
inherit the omission.** The live v6 workflow runs only
`review_editquality.result.checks` + `review_guardian.result.checks` on a
revise round; the bug-historian's prompt solicits checks that are then never
executed. The §4 patch as written would leave the reuse-agent's checks
equally unrun. Recommended: the v7 migration's `run_checks.check_fields`
carries all four reviewers. (The gate seed already runs all seats' checks.)

**(iii) Context for §5's scaling question:** the owner's 2026-07-17 rulings
for the gate were: scope `platform/`+`internal/`+`pkg/`, advisory launch,
credits per submission, and **wait for more seats before launch** — so the
gate's cadence (per task/commit, fleet-wide) is exactly the load §5 warns
about. That strengthens option (b)/(c) over (a) for the wider roster.
Decision stays with the owner; recorded here so both threads argue from the
same facts.
