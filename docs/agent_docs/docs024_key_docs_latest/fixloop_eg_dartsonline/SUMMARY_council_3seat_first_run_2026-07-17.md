# The council's first three-seat run — what we're doing, where we are, where we're going

*2026-07-17, "diagnosis fixloop 3" thread. The read-through account of the
council's first live run with the bug-historian seat (added 2026-07-16 by the
concept-register thread, fix-proposer v6). Companion detail:
`NOTES_running_fixloop(10).md` turns 34–36; the case under review is
`bugs_open/008` (stop_reason undecoded), diagnosis correlation `e505f70f`.*

---

## What we're doing

The fix loop's council exists so that no machine-authored code change reaches a
PR on one model's say-so: a proposed edit plan is argued by independent
reviewers with different briefs, and a deterministic router — not a model —
combines their verdicts. Until this week the council had two seats: right-edits
(editquality) and platform-safety (guardian, who holds the only hard veto). The
concept register's stage 3 exists to widen that roster with seats chosen by
evidence — concepts the platform's history shows are rediscovered over and over.
The first such seat is the **bug-historian**: a reviewer whose brief is to check
every proposed fix against the platform's documented bug history and recurring
failure patterns. This run was its first live vote, on the first real-case
CONFIRMED diagnosis the loop has produced (BUG A: `GenerateText` discards
`stop_reason`, so max_tokens truncation returns as silent success).

## Where we are now — what actually happened

The proposer drafted a correct, tightly-scoped plan: decode `StopReason`, fail
loud on `"max_tokens"`. Three rounds followed, and the same shape held in each:

| seat | round 1 | round 2 | round 3 |
|---|---|---|---|
| editquality | approve | approve | approve |
| guardian (hard veto) | approve | approve | approve |
| **bug-historian** | **object** | **object** | **object** |

Revise cap reached → decision **`exhausted`** → a full escalation package
persisted on the diagnosis correlation (all three plan revisions, all council
reports, `decided_by: "objection from bug_historian — revise cap reached"`).

**The objection was materially right.** The historian argued from pattern
history: the platform has reproduced this silent-drop shape repeatedly, and its
recorded pattern says patching ONE call site of a generic mechanism leaves the
class open. Its blocking question — *are there other LLM provider adapters?* —
was answered from source after the run: **yes, `platform/aiservice/ollama.go`
carries its own `GenerateText`**, untouched by the plan. Two approving seats
missed a real scope gap; the new seat caught it on its first vote, kept it
advisory rather than veto (exactly its designed weight), attached runnable SQL
checks, and explicitly told the human what would dissolve its objection.

**The escalation is therefore a SUCCESS terminal, not a failure.** The tool
refused to bless a partial class-fix and handed a complete, honest package to a
human. The 008 handoff now instructs the fixing thread to start from that
package and cover the ollama sibling.

**What the run exposed about the tool** — the reason the loop couldn't resolve
the objection itself: the historian's blocking question was **code-shaped**
("how many adapters exist?"), and the council's verify step (`run_checks`) can
only execute SQL against the platform DB. The proposer had no way to answer a
question about the codebase, so three revise rounds burned without the one fact
that would have settled the round. This is the long-noted F2.3b(c) residual
("code-shaped questions need a lookup_code_symbols check tier") — now
demonstrated on a real case instead of a benchmark footnote. The irony recorded
for the record: the `code_symbols` index (3,723 symbols, source bodies, trigram
search) could answer the exact question in one query —
`SELECT path, symbol FROM code_symbols WHERE symbol LIKE '%GenerateText%'`
returns both adapters. The evidence existed; the council had no hand to reach it.

A second sample, free of charge: a mis-fired trigger also ran the 3-seat council
against the darts benchmark, and behaved identically (historian object,
exhausted, escalate) — consistent behaviour across two cases in one afternoon.

## Where we're going

1. **The code-lookup check tier (F2.3b(c)) — being built now, in this thread.**
   Reviewers gain a `code_checks` channel beside their SQL `checks`: structured
   code questions (symbol lookup, content grep, path listing) answered from the
   `code_symbols` index — a DB read, so it works in the chassis pod without the
   GitHub token (which only spawned pods hold, by design). Had it existed today,
   round 2's repropose would have carried "two adapters: anthropic.go,
   ollama.go", widened the plan, and plausibly earned approval within cap.
2. **The seat stays.** First-vote evidence says the bug-historian adds judgement
   the other seats lack. Its objections are advisory by design; the router's
   `exhausted → escalate` handled the disagreement honestly.
3. **Roster model policy is an open owner decision** — proposer and reviewers
   still run `claude-sonnet-4-6`; only diagnose-agent moved to `claude-sonnet-5`.
4. **The council-gate thread** inherits all of this directly: the same roster,
   the same verify machinery (now including code lookups), opened as a service
   for every thread's fixes. Its seed (`0NN_council_gate.sql`) already exists.
5. **BUG A ships via the escalation package** — the fixing thread (bugs_open/008)
   extends the plan to both adapters and PRs it; the council can re-review the
   widened plan if the owner wants the loop to close its own loop.

## The one-line state

Three seats argued a real fix; the newest seat was the only one that saw the
class behind the instance; the router escalated honestly; and the missing
capability that would have let the loop settle it alone is now under
construction.
