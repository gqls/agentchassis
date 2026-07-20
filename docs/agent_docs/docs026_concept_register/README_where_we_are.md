Stage 1 of the concept register is complete. Here's the full picture:

What was built

docs/agent_docs/docs026_concept_register/ — a new directory, nothing else touched:

- 1,627 concepts across 107 category register files, each entry with status (deployed/partial/aspirational/superseded/abandoned/unknown), dated evidence, sources, relations, and stage-2 verify-later pointers
- Consolidated from 2,185 raw concept blocks, extracted by 34 agent runs sweeping essentially all ~4,111 files under docs/
- Master index: register/000_concept_index.md
- Final taxonomy comparison: 005_TAXONOMY_final.md

How the taxonomy landed

Started from the docs024 spine (~30 categories) but stayed open by design — extraction agents proposed 65 distinct NEW: categories, which consolidation collapsed into a settled 107. The biggest surprises versus the
seed: the diagnosis-loop/fix-loop machinery, imagery, and the site-build/work-item pipeline all turned out far richer than the docs024 index suggested; finetuning-flywheel and a long tail of abandoned early-era
agent-org experiments weren't anticipated at all.

Directly useful to your fix-loop project

- FIX-036 (in the register) is explicitly the vision of a wider council roster — flagged by a consolidator as "the seam this concept register itself is meant to help fill." That's stage 3.
- Several concepts were independently rediscovered 4-6 times across different documentation eras (e.g. "adoption writes first, classifier consumes," the wrapper-orchestrator pattern) — strong signals for which
  concepts deserve a dedicated council seat first.

Loose ends worth your attention

- Two credential leaks found in docs: a Thunder API bearer token and what looks like a real AWS password, both noted in memory for rotation.
- A couple of evidence tensions flagged for stage 2: whether multi-cluster dispatch is actually wired into any live workflow, and whether the trained LoRA adapter was ever wired into production inference.

Stage 2 (verifying each concept against real code/DB) and stage 3 (building council agents per concept area) are next, whenever you want to pick them up.

--

● Stage 2 is now fully complete. Summary of what landed:

The final sweep (100 agents, 174 concepts checked): 18 corrections confirmed, 9 overturned.

- A scheduler cluster (SCH-007/008/009): three concepts, same shape — real fixes landed in cmd/scheduler/main.go (ownership gap, starvation prevention, fire_message handling) but the docs never caught up.
  Three independent confirmations of the same drift.
- A new failure-mode class: DOC-064 was tagged abandoned because extraction's search was scoped to one doc subtree and never reached a sibling folder holding a byte-identical live copy — evidence that was
  never found, distinct from the present-tense-plan misreading batches 1-2 found.
- Several "bundled" concepts were half-superseded — a claimed clean replacement, but part of the old mechanism was still live and wired.

Stage 2 grand total, all 1,627 concepts now checked at least once: 124 corrections confirmed, 106 overturned by the adversarial pass (~7.6% error rate). Final distribution: 853 deployed / 257 partial /
290 aspirational / 90 superseded / 67 abandoned / 21 unknown / 49 convention.

All three running docs, the master index, and memory are updated to reflect completion. Two items remain, both explicitly yours to call: credential rotation (RUNBOOK B1) and whether/when to implement the
stage-3 council design against the live fix-loop workflow (RUNBOOK B4) — that's a cross-workstream production change I've deliberately left for your sign-off rather than proceeding on unilaterally.


---

● Everything's reconciled and verified. Here's the summary of this turn's work:

Doc sweep since last visit (2026-07-14 → -16): 62 commits across many concurrent workstreams. Confirmed via file mtimes that the concept register itself was untouched by anyone else — no drift to reconcile there.

Coordinating with the fixloop thread: Read its own journey doc (SUMMARY_where_we_are_2026-07-16.md) — it independently names this workstream as the answer to its council-widening question, and states the same
boundary already in this register's plan: wiring stage-3 seats into the live workflow is reserved for the owner's explicit sign-off.

Register gap closed: fixloop's entire triage/escalation subsystem (4 phases, all live) shipped after extraction froze, so none of it existed in the register. Added FIX-051/052/053 and updated FIX-034, each
independently verified against live code, then cross-checked by a dedicated research pass that additionally confirmed every commit against git log. Also added STY-049 for the missingkey=zero structural defect
behind fixloop's active real-case — and while writing its relations, surfaced a cross-cutting failure family (the same "schema says required, renderer says silently empty" shape recurring across TL-001,
PBP-012/019, STY-004/019, and CLC-003).

A finding worth your attention, not something I acted on: fixloop chose the image-landing/article-body trap (004) as its first real-case dispatch. A separate concurrent session resolved the underlying data loss the
same day (005, confirmed via file mtime and an independent go test re-run) — 2 of 004's 3 "open items" are actually done. Only the structural defect remains open. Fixloop's dispatch decision on case 004 may need
revisiting — flagged in the plan and runbook (B6), left for the owner to decide.

Stage-3 pilot seat, data-driven: computed rediscovery frequency across all 1,631 concepts per fixloop's own suggested method. Two strong candidates, both matching FIX-036's original roster: reuse-agent
(tool-lifecycle.md, the single most rediscovered category) and bug-historian (the silent-content-loss family above, directly tied to fixloop's active work — recommended first). Neither is built; the choice is
yours.

Register integrity double-checked end to end: 1,631 concepts, perfectly matched between category files and index, zero loss. All three running docs, README.md, memory, and MEMORY.md updated to match.

---

SUMMARY_where_we_are_2026-07-16.md — a plain-language, read-aloud version (matching the same "calm companion to the technical doc" convention fixloop itself uses): where the register came from, where it is now, and
where it's going, no jargon or citations, readable aloud in a couple of minutes.

PILOT_bug_historian_reviewer.md — the complete pilot design for the bug-historian seat. I went with bug-historian over reuse-agent: the recurring failure pattern (7 independent occurrences across the platform's
history) is durable evidence even though the specific triggering incident just closed — a historian's value is catching the next one. Before designing the patch I read the actual live workflow file and the
council-decision Go code directly, which surfaced something not spelled out anywhere in the docs: any reviewer's veto triggers outright rejection regardless of hard_veto_from — it only affects the audit label. That
shaped the design directly: the new seat's prompt offers only approve|object, never veto, so it's genuinely advisory rather than an accidental second gatekeeper.

The document contains the charter, the exact 7-concept curated-context digest, a full prompt template matching the existing two reviewers' style precisely, and the exact 4-edit patch needed in 0NN_fix_proposer.sql.

I haven't touched the live file or database — per the boundary established earlier (and independently corroborated by fixloop's own docs), that step needs your explicit go-ahead. Section 6 of the pilot doc lays out
exactly what applying it would involve. Say the word and I'll do it — as drafted, or with whatever changes you want first.
--

Done. Here's what happened:

Your memory update flagged that fixloop hit its first real-case CONFIRMED diagnosis — directly relevant to the bug-historian's still-open "not yet exercised" item, so I checked in on it (filesystem only, no DB
access needed).

Two new platform bugs found, both independently verified against live source and added to the register:
- MDL-038 (BUG A): GenerateText in platform/aiservice/anthropic.go never decodes stop_reason — confirmed by reading its response struct directly (only Content/Usage decoded). A max_tokens-truncated LLM call
  silently looks like a complete success; the loop's own diagnosis found 17 live instances via llm_call_log.
- MDL-039 (BUG B): an agent's root-level ai_service config silently shadows its step-level config — confirmed by reading ExecuteLLMPromptAction directly. 17-agent fleet blast radius.

The bug-historian itself: still live, still unexercised, and confirmed fine either way.
- BUG A is CONFIRMED but its fix dispatch to fix-proposer "awaits owner go" per fixloop's own notes — the council hasn't actually run yet, so there's nothing new to report on the reviewer's real-world behavior. Not
  stalled, just not dispatched.
- I specifically checked whether BUG B affects the bug-historian, since it also uses a step-level ai_service block. It doesn't — I re-read fix-proposer's default_config and confirmed it has no top-level ai_service
  key, so it's not among the 17 affected agents.

Register is now at 1,633 concepts, integrity-checked clean (no duplicates, files and index perfectly matched). All docs and memory updated, committed in one narrow commit (173bb297e) alongside everything already in
place from last time.

Nothing left pending on my end right now — the next real update depends on fixloop's owner dispatching BUG A's fix, which will be the bug-historian's actual first outing.
---


Clean. The only remaining uncommitted files are the two the other thread owns (their addendum in the reuse-agent pilot, and their gate file) — both left for them, exactly as they should be.

Both tasks are done. Summary:

Guidelines-agent (seat #3) is live — council is now 5 sequential reviewers: editquality → bug_historian → reuse_agent → guidelines → guardian. Its distinctive design handles FIX-036's two clauses: a rule violation
→ object (triggers revise), but a guideline-gap (the fix is right, the rule is wrong — like BUG B's backwards max_tokens runbook rule) → approve + a note, never an objection, so a correct fix isn't punished for
exposing a bad rule.

The relevance filter is designed (DESIGN_relevance_filter.md). The key finding: unlike the pure-SQL seat additions, it needs a chassis-image Go change — council_decide hard-fails on any absent reviewer, so skipping
a seat requires teaching it to treat absent as abstention, plus a select_review_panel action. That's a bigger, cross-thread change, so I've presented it as a build decision rather than proceeding.

Two things happened along the way worth flagging:

1. A coordinating thread caught a real defect in my own seats. They reviewed the reuse-agent pilot and noticed run_checks only executed 2 of the 5 reviewers' checks[] — so the three advisory seats could request a
   fact-check that was silently dropped. I verified it (true, and safe to fix), fixed it in v9, and applied it. A genuine gap I'd missed, caught by exactly the kind of cross-review the council concept is built on.
2. I nearly collided with that thread's live edit and backed off. I started syncing their council-gate file to keep the two rosters in lockstep — but my exact-match check failed because the file changed under me
   mid-write: they're actively editing it and already syncing the roster themselves. My script failed cleanly before writing, so nothing was corrupted. Correct call was hands-off — their file, their task. (This
   collision also sharpened the relevance-filter design: two hand-maintained rosters that must stay identical is itself the drift-failure these seats exist to catch, which argues for the shared-Go-action approach.)

All committed narrowly — my docs, the v8 and v9 migrations as separate commits — leaving the other thread's two files untouched.

---

Done — the diagnosis-loop guardian (candidate #4) is live. Council is now 8 reviewers: edit-quality + guardian always-on, six specialists gated behind the relevance filter.

The seat: grounded in the register's highest hot-concept-density category. Its job is to defend the diagnosis machinery's hard-won honesty gates when a fix touches the loop itself — cite-or-abstain (DIAG-001), the
two-evidence-family citation standard that correctly withheld CONFIRMED on BUG B (DIAG-008), the three-layer read-only SQL enforcement (DIAG-009), observability-never-fails-a-diagnosis and skip-never-guess notes
(DIAG-028), the config-level error_step trap (DIAG-030 — step-level placement is silently inert, a recurring class), and token/pod isolation (DIAG-019/022). There's a fitting self-reference in it: the loop
reviewing fixes to itself is exactly why these guards need a defender — they were earned from benchmark runs that produced CONFIRMED verdicts a fixer must never have acted on.

The execution followed the settled routine: drift-check first (the row had been touched by another thread at 09:21 this morning — all my anchors verified intact, which is precisely what the surgical pattern is
for), grounding from the register, gated + surgical migration (8 chained jsonb_set, atomic, idempotent, snapshot), pre-flight (no active runs), apply, verify (8 seats, chain tail converging correctly, all arrays at
8, proviso + code_checks byte-intact), document, commit narrowly.

Remaining candidates: #5 improvement-loop, #6 compliance, #7 render, #8 LLM-reliability, #9 debugging. Say the word and I'll keep going down the list.

---

All done — clean tree. Three deliverables this turn:

1. The summary is written — SUMMARY_where_we_are_2026-07-18.md, same read-aloud style: what we set out to achieve (grow a 2-member review panel into an evidence-chosen council), where we are (9 reviewers live, the
   relevance filter routing them, your stability preference on the guardian), and what's next.

2. Your multi-model idea: a different subproject — and here's the honest distinction. The debugging seat is a reviewer: it reads a proposed fix and asks "has anything like this happened before?" against the
   platform's 74 documented debugging lessons — cheap, advisory, wakes only when relevant. Your idea generates diagnoses that don't exist yet: when the diagnosis loop honestly gives up on a bug (exhausted or
   unverifiable — both currently hand to you), fan the same evidence bundle out to a diverse panel of models — different vendors, different strengths — and see if any of them cracks it. That belongs at the diagnosis
   loop's escalation terminal, not in the review chain. I've recorded it in the PLAN as the "multi-model diagnosis gauntlet" with the register-verified groundwork (per-step model routing already exists; the platform
   has Anthropic + Ollama clients, other vendors need new ones) and one prerequisite worth fixing first: BUG A (MDL-038) — an ensemble comparing model outputs must know when an output was silently truncated. It's
   ready to start whenever you green-light it.

3. Per your instruction, the list continued as stated — seat #5, the improvement-loop guardian, is live (council now 9). It defends the termination guards born from a real incident: the audit→fix→re-audit loop once
   ran unbounded, generating 845+ findings in ten days. It objects to anything that bypasses the pass cap, auto-unlocks sections, enables a discovery check with no handler, or lets a check insert its own work items
   around the dedup machinery. Same routine as always: drift-check, gated + surgical, snapshot, verified intact.

Remaining: #6 compliance, #7 render, #8 LLM-reliability, #9 debugging — then the council side is essentially complete.

---

The live run is progressing — the panel woke bug_historian and reuse_agent this time, and it's mid-review. The monitor will notify me when it reaches a terminal state; I'll report the full outcome then.

Interim status while it runs:

- I did not dispatch — a concurrent session (acting on your same instruction) fired the run at 10:16, about a minute before I would have. Dispatching again would have double-run the correlation, so my job switched
  to monitoring.
- Premise checks all passed first: BUG A still unfixed in HEAD and in the running pod's binary, no competing work items in the queue, the CONFIRMED diagnosis and its evidence bundles intact, pod well past the
  dispatch-quiet window.
- The big discovery: fix-proposer already ran against BUG A three times yesterday evening (by the fixloop thread) — and the 20:13 run was approved by the filtered council, with the relevance filter selecting only
  the reuse-agent and council_decide correctly recording 3 abstentions for the skipped seats. The filter and abstention mechanism have already worked in production; today's run is the first on the complete 13-seat
  council.

---

The council's first real outing produced a genuinely sound fix — and notably, the revise loop improved it: the final plan covers both provider clients (anthropic.go and ollama.go), because a code lookup during
reproposal found the identical silent-truncation gap in Ollama that the first-round plan had excluded. Let me confirm the durable artifact was persisted and check what the implementation stage requires.

Ran 1 shell command

● Now I have the complete picture. Let me record this milestone durably before reporting.

Committed 5b0e2b5f0, ran 1 shell command

● The council's first real outing is complete — and it worked.

What happened

I did not re-dispatch: my six premise checks passed (BUG A still absent from HEAD and the running pod's binary, no competing work items, diagnosis and evidence intact), but a concurrent session fired the run at
10:16 — about a minute before I would have. Re-firing would have double-run the same correlation, so I monitored it through instead.

Run 53da3a30 ran the full 13-seat council and ended APPROVED after 3 revise rounds (10:16 → 10:37). The relevance filter woke 5 specialists — bug-historian, reuse-agent, diagnosis, LLM-reliability, and
debug-historian — plus the two always-on seats; council_decide correctly recorded the other 6 as abstentions, exactly as designed. A typical fix waking ~5 of 13 seats is precisely the scaling behaviour the filter
was built for.

The two seats built this week earned their place

Both approved — with grounded, specific review notes rather than rubber stamps:

- LLM-reliability confirmed no ai_service config was touched (so the BUG B root-shadows-step trap was irrelevant), verified the error guard sits after the token write-back so llm_call_log still records
  output_tokens on the error path, and noted no model swap was implicated.
- Debug-historian cleared all four lore dimensions (no SQL surgery, no informational-column blast radius, log-state not git/tag deploy verification, no template surgery) and independently flagged the same
  write-back ordering constraint for code review.

The revise loop demonstrably improved the fix

Round 1 touched only anthropic.go. During reproposal a code lookup found finish_reason absent from the entire codebase — surfacing the identical silent-truncation gap in the Ollama client. The final approved plan
covers both providers: add StopReason + a max_tokens hard-error guard in anthropic.go, and a parallel DoneReason == "length" guard in ollama.go. It's grounded in 23 real silently-truncated llm_call_log rows (the
output_tokens == max_tokens fingerprint, across 5 agent types and 2 models). That's a genuinely sound fix that would close the bug where you were about to build the multi-model gauntlet on top of it.

The decision point

The approved fix_plan is persisted. The next stage — the fix-implementer (092 trigger) — is a distinct, outward-facing dispatch: whole-file LLM implementation → allowlist → a fix/* branch → commit → gofmt+build
gate in a k8s Job → on green, a real PR into main. I've held there rather than assume, since "the fix dispatch" you asked for was the council run, and opening a PR against production LLM-client code is a bigger
call.

Want me to dispatch the implementer and let it open the PR (build-gated, so a red build leaves the branch + log and opens nothing), or would you rather review the approved plan first?

※ recap: The concept-register council is complete and passed its first real test: the BUG A fix ran through all 13 reviewer seats and was approved. Next action is your call on whether to dispatch the implementer,
which opens a real PR into main.

--

Re-read done, and it caught two process gaps in my recent work. Here's what I found and corrected.

The real miss: the council-gate roster

CLAUDE.md's "If you add or change a council seat" rule (lines 70–78) changed while I was working, and I violated it. There are two rosters that must stay identical — fix-proposer and the advisory council-gate (the
one any thread uses via 097). The rule is now: seat fix-proposer, then run the mirror 099_SYNC_gate_roster.py — don't hand-patch the gate. I seated fix-proposer five times (v14–v18) and never ran the mirror.

I checked, and the rosters are in sync — 13 seats, footprint maps byte-identical, review-step sets identical, confirmed by the sanctioned 099 dry run (added/removed: none). But that's because another thread's 099
run closed the gap, not because I did my half. Nothing to --apply; the correction is that I've now recorded the mechanical method so I (and future sessions) run it after seating rather than relying on luck.

The second miss: a durable diagnosis needs a guide entry, not just a case file

CLAUDE.md (lines 165–168) says when you diagnose something durable you file the bugs_open/ case and add the transferable pattern to the debugging guide's §9. I filed bugs_open/013 but skipped the guide. Now fixed
(8035e39d2): added the §9 pattern ("a generator feeding a downstream verifier must satisfy its contract at write-time; an unformattable body is a truncation signal"), indexed 013 in §10, and — since I committed the
BUG A fix — corrected 008's index state from "fix not shipped" to "committed f32b208e5, not yet deployed".

One thing I got right, that I'd earlier worried about

My Council-Reviewed: 53da3a30 trailer on the BUG A fix is valid — lines 53–58 were updated to explicitly accept a fix-proposer council RUN_ORCH_ID, and 098 resolves it by prefix. So no false audit join.

Memory updated to replace my stale "hand-patch both councils" note with the current mirror-script method. Clean tree on everything I touched.

✻ Worked for 4m 7s

※ recap: We completed the 13-seat fix-loop council and proved it end-to-end: BUG A's truncation fix was diagnosed, council-approved, and committed to branch 085. Next action is building and deploying the chassis
image so the fix goes live.


---

> **Note 2026-07-19:** this file was renamed from `README_state_of_play.md` to
> `README_where_we_are.md` — the canonical name in CLAUDE.md's new "standing five"
> working-docs directive, so you can find it where you'd expect. Same file, full
> history preserved; nothing above was edited. Also filling a cadence gap: the two
> turns below (the implementer run and the doc-compliance turn) happened but were
> never logged here as they went — which is exactly the habit the directive is
> pushing against. Caught up now.

**Landing BUG A's fix (2026-07-18, filling the gap above).** You said to dispatch
the implementer, and noted the deploy system currently builds from the working
branch (085) rather than main. I ran the implementer against the approved plan. It
did the right thing and it also showed me a real rough edge:

- The implementer spawned a child that generated *logically correct* code for both
  guards (anthropic + ollama) and pushed a branch — but the build gate then failed
  it on `gofmt`, because the model had added a new struct field without re-aligning
  the neighbouring one (two whitespace characters) and left a trailing blank line.
  So: no PR. **The gate did its job** — it refuses to open a PR for code that isn't
  clean, even when the fault is cosmetic. Better a wasted run than a messy PR.
- Rather than re-roll the dice on the model (it tends to make the same alignment
  slip), I finished the last mile by hand: applied the two approved guards as small
  edits to the working tree, ran `gofmt`, confirmed it vets and builds, and committed
  it narrowly to 085 as `f32b208e5`. The fix is now on the branch that deploys — but
  **not yet built or rolled**, so it isn't live; that image build is still yours /
  the deploy thread's call. Both guards sit *after* the token write-back, so the log
  table still records the token count even on the new error path — a detail two of
  the council seats had specifically flagged.
- You asked me to file the rough edge as a bug. It's `bugs_open/013`: the implementer
  should run `gofmt` on its own generated files *before* the gate, so trivially-
  unformatted output doesn't burn a whole run. Filed with the exact fix location, and
  the transferable lesson added to the debugging guide. I also deleted the stale
  branch the implementer had left behind.

**Two CLAUDE.md re-reads (2026-07-18/19) — process housekeeping.** You twice asked me
to re-read CLAUDE.md, and both times it had genuinely changed under me (it's edited
constantly by other threads; the copy I load at startup goes stale). What that caught:

- I'd added the last five council seats to the `fix-proposer` roster but hadn't run
  the mirror that keeps the *second* roster (the advisory gate any thread submits to)
  in step — a rule that had just changed to "run the mirror script, don't hand-patch."
  They happened to be in sync because another thread ran the mirror, but I'd only done
  my half. Recorded the method so it's not left to luck next time.
- A durable bug diagnosis is supposed to leave *two* traces — the case file and a
  pattern entry in the debugging guide — and I'd only left the case file. Added the
  pattern.
- This turn: the docs directive itself. I had all five standing docs but this one under
  the wrong name and one turn behind — renamed and caught up, as above. The honest
  lesson for me is the cadence: this log should grow while the work happens, not get
  reconstructed when you ask where things are.

**Where we are now.** The council is complete and proven (13 seats, first real case
passed). BUG A's fix is committed to 085, awaiting an image build to go live. The
open threads are unchanged: build+deploy the fix (yours/deploy thread), and the
proposed multi-model diagnosis gauntlet whenever you want to start it.

---

**Constitution & mission became gatekeepers (2026-07-20).** You gave three
directions: fixes must tackle the root cause, not work around a bug; the
constitution needs a gatekeeper; and everything should follow the mission. I read
the actual documents first, and the finding reframed all three into one: the
platform already *has* a written constitution and a written mission, both of which
say in their own text that changes should be checked against them — but nothing
ever did. Your first point was literally already a constitutional rule ("fix
structural problems, not symptoms") that no reviewer enforced.

So rather than invent new principles, I pointed the council — the gatekeeping
machinery we'd just finished — at the two documents that matter most. Two new
always-on reviewer seats now sit on every fix and every submitted platform change:
a **constitution** seat (root-cause-not-workaround first, then reuse-before-
recreate and the other always-on rules) and a **mission** seat (best site per
domain; the revenue model shapes the site; never let an agent silently override the
strategic direction). Always-on rather than occasional, because that's what a
constitution and a mission *are*. They're advisory like the other seats — an
objection forces the fix to be reworked, it doesn't kill it — which is the right
behaviour: a workaround gets sent back to become a real fix, not rejected outright.

Both councils (the fix loop's own, and the gate any thread submits to) now carry
the two seats, verified identical. Two things I flagged rather than decided: I've
started only on the fix/change councils — extending this to the feature and
experience councils, and to the site-build pipeline itself, is the natural next
reach but a larger, separate piece. And one part of "the direction is fixed" that
these seats don't cover: protecting the constitution and mission *documents* from
drift — that they change only with your sign-off, not on a passing vote. Worth
deciding how you want that guarded.

(Also noted in passing: my earlier BUG A fix was quietly replaced by a better,
more structural one from another thread — it now carries the truncated text back
to the caller instead of discarding it. A live example of exactly the root-cause-
over-patch principle you're now having the council enforce.)

**Two more seats' worth of teeth, and the two plans you asked for (2026-07-20,
later).** Your rerender trap — "page-rerender re-deploys the existing HTML, it
does not regenerate it from content_data" — is now in the render reviewer's
charter, worded to match the corrected understanding another thread nailed down
yesterday, and the reviewer now objects when a plan cites "the rerender
completed" as proof a content or template edit is live (that false green has
already burned us — it's the mechanism behind an open bug). Your suggested
prior-art librarian is live too, on both councils: it fact-checks a plan's
claims that something "doesn't exist" or "needs to be built" by actually looking
— code search, agent records, run history — because we've twice watched a plan
sail through every reviewer on a false premise they all inherited. The council
is now sixteen reviewers: five that always sit (craftsmanship, constitution,
mission, the librarian, safety) and eleven specialists woken by relevance.

And the two plans are written, not started, awaiting your word. Reaching the
mission into the site-building pipeline itself: first let the strategic-brain
classifier actually see the platform mission, then an observe-only reviewer on
its output that files findings somewhere they'll actually be read (we checked —
one earlier lane died precisely because nothing consumed its findings), then
enforcement only once a week of numbers shows it isn't crying wolf, then a fleet
audit for already-built sites. Guarding the direction documents themselves:
bless one canonical copy of each, a ledger of their fingerprints, a commit gate
that blocks edits to them unless the message carries your explicit sign-off
mark (same tier as the existing secrets gate — it fires only when someone
touches the constitution, which is exactly when friction is wanted), a checker
that watches all three places the words live (the files, the copies, the
reviewer prompts in the database), and eventually the long-promised move of the
constitution into proper database rows. Decisions that are yours in it: whether
the commit gate blocks or just shouts, and when observation is allowed to
become enforcement.
