Git (commit-per-task): A fresh git status showed the multi-session reality the file warns about had already played out — all my claims-verification code had been swept into concurrent sessions' commits (core under
87d13b864 "claims verification further development", the rest under d076c3c8e/f51a7accc). Forward-only, nothing lost — and it had a silver lining: it's why the deployed image already carries my final code including
the precision fix. The one remaining piece of my task (the deployment-verification notes entry) I committed narrowly by pathspec as ab0b1e452 — HEAD moved again seconds before, and the pathspec commit excluded
everything else exactly as the rule intends. I did not touch SPEC_voice_tells_check.md / datahelpers/voicetells.go — that's another session's in-flight WIP in the same directory (looks like a sibling check being
developed; V3-adjacent territory).

Production verification (completing the deploy follow-up):
- Verified against the running pod, never git: strings /app/agent-chassis contains all the claims symbols — the V1a gate is live for any evidence-base site.
- unverified_claims is enabled on quality-discovery-agent (alongside its sibling placeholder_contact); workflows are re-read per message so it took effect immediately.
- Two production runs on leopardess: the improvement-loop run executed the check (40 log mentions in the discovery pod) with zero claims_unverified items — the correct result on the cleaned site; the loop's later
  failure was a pre-existing call_site_review timeout, downstream of and unrelated to discovery. A second single-agent discovery run COMPLETED with zero items of any type. One behaviour worth knowing: a clean check
  leaves no positive trace by design (checks emit findings only), so "ran clean" is evidenced by COMPLETED + zero items — now recorded in the notes.

V1 is done end to end: evidence base live → gate in the build path → post-deploy check in the discovery cycle → HITL routing → clean production baseline. Next phases when you want them: V2 (writer whitelist
injection — the direct answer to the spec-instructed leaks), V3 (LLM claims auditor for prose assertions), V4 (freshness), and vetcomparison as the second evidence-base site.

----

V3 is complete, proven on both paths, and committed (0540698a4). Summary of this stretch:

What just happened

The claims-auditor (V3, the judgement lane) is live. One LLM pass per site: loads the evidence-base fact register, takes tag-stripped text of every deployed unlocked page, and reports only unsupported prose
assertions — with the audit's hard-won semantics in the prompt (a fact supports its wording, not its topic; honest "we could build" framing is fine; a true number under a false label is unsupported; named entities
must trace to the allowed list). Findings become one needs_human_review item per pass (claims_llm:<domain>) with the human-review pseudo-handler — human-terminal, no auto-fix path existing at all.

Both behaviour paths verified by artifact:
- leopardess (opted in): run completed, one LLM call, and the model's entire response was the literal [] — four output tokens — so no item was created. Clean passes cost one call and produce zero queue noise.
- robot-hands (no evidence base): run completed with zero LLM calls — the opt-in gate short-circuits before any cost. The fleet-safety rule holds by construction.

A platform bug surfaced on the first dispatch: checkpoint_for_review — the documented HITL checkpoint action — turned out to have never been registered in the action registry; its file header describes a
registration that doesn't exist, so any workflow referencing it fails validation with "requires a topic". Nothing had ever used it. I worked around it with the registered, dedup-aware create_work_item (same HITL
shape), and added the missing registry entry in Go — that rides your next image build. Lesson banked in the notes: a header's "Registration:" block is an intention, not a fact.

Honest boundary, on the record: the auditor's catch path isn't yet demonstrated — I wasn't going to plant fabrications on a live site to test it. Like V1b (which proved itself within hours on the invented service
name), it will be proven by the first real drift. And it's not wired into any cadence yet — that's your call, at one LLM call per site per pass.

Where the workstream stands

V0 (evidence base) → V1a (build gate) → V1b (post-deploy check) → V2 (writer whitelist) → V3 (prose auditor): all built, all live, all production-proven. What remains is V4 — freshness: a scheduled task that
re-runs the SQL-sourced facts, updates values and dates, raises stale_evidence items on tolerance drift (underclaiming included), and regenerates the writer's whitelist so it can never go stale — that's a Go action
plus a scheduled_tasks entry, riding a future image. After that, vetcomparison as the second evidence-base site with its price-claims lane. Say the word and I'll build V4.


--

My council submission was fully reviewed and then thrown away by a platform bug — which I've filed rather than worked around.

What happened to the V4 review

All seven seats reviewed my change and produced verdicts. Then review_guidelines came back at 8,000 output tokens against a max of 8,000 — the CLAUDE.md rule verbatim: output_tokens == max_tokens means the completion
was cut, not finished. Its JSON was truncated, diagnose_council_decide hard-errors on unparseable reviewer output, and the run went to complete_invalid — discarding six complete, well-formed reviews, including the
guardian's (4,856 tokens) and the bug-historian's (5,548). No verdict, no report, all seats' credits spent.

So I still don't have a verdict on V4, and it's not because anything is wrong with V4.

Filed, not patched

The action isn't mine — it belongs to the fixloop/council threads, and the runbook warns against touching those seeds without diffing against the live row. So per CLAUDE.md I filed the case and the pattern:

- bugs_open/019 — evidence table, root cause at diagnose_council_decide_action.go:99–126, and fix candidates. The best one follows the code's own reasoning: it already treats an absent reviewer as a principled
  abstention (relevance-filter skips), so a present-but-unreadable one should be a loud abstention too — counted separately, logged, carried into the decision — with the safety property preserved by never letting an
  unreadable seat upgrade a verdict to approve. That keeps six good opinions instead of destroying the round. Secondary: salvage via the existing repairTruncatedJSON, since verdict and reviewer come early in the object
  and usually survive.
- 016b §9 — the transferable pattern, including how to tell the two complete_invalid causes apart (failed_step: persist_submission = your JSON is malformed; council_decide = a reviewer was cut and your plan was fine).

This is the fourth instance of the same truncation family (005 article bodies, 008 stop_reason, 012 component wreckage) — the platform keeps detecting truncation after acting on the fragment. I noted that in both files,
because a shared guard would be worth more than a fifth local check.

Also worth flagging: four of seven seats used more than half the ceiling on this submission, so the margin is thin precisely on substantial changes. My own long rationale contributed — if I resubmit I'll tighten it, but
that's a workaround for a defect that will bite the next thread too.

Where V4 itself stands, unchanged by any of this: code built and tested (guard refusals, all four tolerance modes, under-claiming, whitelist composition), registry entries added, writer_line populated on all 18
leopardess facts, seed written and deliberately unapplied pending an image. It's uncommitted, pending either a resubmission or your call to commit it unreviewed given the gate is currently broken for submissions this
size.


---
Written and committed: SUMMARY_2026-07-19_claims_verification_complete.md (77970a92d), in the claims_verification folder alongside the two earlier dated snapshots it supersedes.

It's structured for reading aloud, in five movements:

Where we've been — the gap: every quality layer we had asked whether a page was well-made, none asked whether it was true. Then the concrete cost: invented case studies, an invented founder with a 404 photograph, "2,767 Awards
Won", an eight-department taxonomy that exists nowhere, fabricated vet prices with legal exposure. And the detail that decided the design — the one fabrication class with a deterministic checker is the one class that was ever
caught.

What we set out to do — the three principles: evidence is data; deterministic first; truth decisions are human.

What we've done — the five components in plain language, including both things you asked be covered explicitly: the whitelist that flips "never invent statistics" into "these are the only numbers you may assert", and the
reasoning for keeping the banned claims out of the prompt (don't-think-of-an-elephant — the blacklist is enforced by the deterministic gate instead, so each mechanism does what it's actually good at). Freshness gets the same
treatment, including the humans-own-the-words / machine-owns-the-numbers split and the two safety properties (untrusted SQL, and compare-and-swap so a person's edit is never lost).

The results — history as a test suite; the first scan falsifying our own "clean site" assumption; the poisoned specs that were instructing the fabrication; the writer citing a whitelisted figure with honest dating unprompted;
the stranger caught within hours and ruled on; and the two platform faults found on the way past, including the council defect that voided its own review.

Where we are and where we're going — four of five live, freshness awaiting one image build, the two honest boundaries stated plainly (the auditor's catch is undemonstrated because we won't stage a fabrication; its cadence is a
cost decision that's yours), then vetcomparison, then the fleet by invitation.

✻ Brewed for 1m 12s

※ recap: We built a claims-verification layer so the platform can't publish unsupported claims; V0–V4 are done, with V1–V3 live and V4 committed but inert. Next action is one chassis image build, then switch on the freshness
scheduled task.


----

**2026-07-19 — freshness built, the council broke on us, and the docs caught up**

Since the last entry, three things happened.

**Freshness (the last component) is built and committed.** It re-runs every fact
whose source is a live database query, updates the number and the date it was
checked, and raises anything that has drifted for a human to rule on — including
the case where the site is *under*-claiming, saying 2,767 when the database now
says more. Not a lie, but stale, and worth knowing. It also regenerates the
writer's whitelist, and the split there matters: humans own the words, the machine
owns the numbers. Each fact carries a human-written line with a gap where the
figure goes, so a caveat like "this is a catalogue count, never present it as a
running fleet" survives regeneration word for word. A fact nobody has worded is
left out entirely rather than auto-worded.

Two safety details worth knowing about, because both were deliberate. The queries
live in a data column, so they are treated as untrusted — read-only, one statement,
must begin with SELECT, nothing that could write. And because the pass rewrites a
whole record that a human owns, it checks nobody edited it in the meantime; if
somebody did, it writes nothing and says so. Losing a refresh costs one scheduled
run. Losing your edit costs trust.

**The platform's own review council broke while reviewing it.** I put the change
through the advisory council as the house rules now ask. All seven reviewers read
it and produced verdicts — and then the round was thrown away, because one
reviewer's answer was cut off at its length limit, and the code that tallies
verdicts refuses to work if any single reviewer is unreadable. Six complete
reviews binned. That is a real defect, so I wrote it up as a bug with a proposed
fix rather than working around it. Within a day another thread hit exactly the same
thing on an ordinary submission and added their evidence to the file.

I also got something wrong that is worth recording. When my own council run was
slow to appear, I went looking for it by "most recent run" — and found another
thread's run instead, and reported its progress and its verdict as if it were
mine. Every council run looks the same in that table, so recency cannot tell them
apart. The right handle is the submission's own correlation id. I corrected it as
soon as the other monitor fired.

**And the documentation caught up with the rules.** Re-reading the house
instructions, this workstream was missing two of the five standing documents — a
plan recording decisions *and their reasons*, and a runbook of commands. Both now
exist. The runbook is the one that will save time: every query that was fiddly to
get right, with its trap written next to it. The plan records four places where the
original design turned out to be wrong, marked as corrections rather than quietly
fixed — including that its stated baseline ("the site is currently clean") was
false on day one.

**Where that leaves us.** Four of the five components are live. The fifth needs one
image build to activate, and its scheduled task is written and waiting. Nothing is
imposed on any site that hasn't asked for it. The one open choice for you is
whether to put the prose auditor on a schedule, which costs one model call per site
per pass.
