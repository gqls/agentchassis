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

**The direction guard is real now, and the mission reaches the pipeline
(2026-07-20, later still).** You said go ahead, so: the constitution and mission
files now have a ledger naming the one true copy of each (with fingerprints),
and a commit gate that physically blocks any edit to them unless the commit
carries a "Direction-Approved" mark — which only your word earns. The gate's
very first act was to challenge the commit that created it, since the ledger is
itself protected; it passed on the approval you'd just given, which is exactly
how it should work. A checker script watches all three places the direction
lives — the files, their copies, and the reviewer seats in the database — and
its first run came back all green.

On the pipeline side: the classifier that decides what every site should be now
reads the platform mission as part of its brief, and right after it writes its
decision, a new observe-only reviewer judges that decision against the mission —
did it argue a revenue model from evidence or just default to "consultancy
site"? It can't block anything; it writes its objections to a log, and there's a
weekly report to read them. After a week or so of real numbers we'll know
whether it's finding real drift or crying wolf — that's the evidence for your
promotion decision later.

One honest correction from the build: the plan originally had the reviewer's
findings going into the work-item queue. Reading the actual code showed the
queue's triage would have swept them into the build pipeline toward a handler
that doesn't exist — the exact "findings die with no consumer" failure we were
designing against. They go to the notes log instead, with the report as their
reader. The plan carries the correction visibly, as ever.

**Handing over (2026-07-20, end of day).** Tonight's fresh build went out and I
checked everything against the actual running system: the truncation fix family
is live in the binary, and all of today's council and guard work survived the
deploy untouched. One of the open bugs (the original truncation case) was
closed by the thread that owned it, with their own verification. For the next
conversation: the cold-start file is `HANDOFF_2026-07-20_council_continuation.md`
in this directory — what's live, what's proven, the open decisions (mission
reviewer's probation verdict, the two remaining councils, the fleet audit, the
standards table, the multi-model gauntlet), and the landmines written down so
nobody pays for them twice. Today's milestone summary is
`SUMMARY_where_we_are_2026-07-20.md`, written to be read aloud.

**Bringing the register up to date, and clearing out the duplicate documents
(2026-08-04).** You asked what the concept register is, how it was collected, what
to do next — and then to bring it up to date and delete the out-of-date duplicate
documents.

The register is a complete list of every nameable thing this platform does — every
mechanism, contract, agent, tool and idea — with a status on each saying whether
it's real. It was built in July by sweeping all ~4,100 documents under `docs/`,
pulling out 2,185 raw concepts, merging them down to 1,627, and then checking every
single one against the actual code and database, which corrected 124 of them. It's
grown since to 1,756, one entry at a time, added by whichever thread built the
thing. The point of it is that a session about to build something can find out
whether it already exists — and the council seats that review our changes are
seeded from it, so a wrong entry isn't just misleading prose, it's false evidence
in a machine review.

What "up to date" turned out to mean wasn't another sweep. The sweep froze on 13
July and 129 of the 155 workstreams on disk are newer than that, so re-sweeping
would be a treadmill. What keeps it current is that every thread registers what it
builds, and a coverage check that shouts when a workstream appears that the
register has never heard of. So the real question was whether that machinery still
works. It didn't, in two places, and both failures were the same kind — **a check
that could not have failed.**

The first: 34 concepts had a proper entry in the register and no row in the master
index. Among them the entire first half of the claims-verification layer — the
thing that stops our sites making claims we can't back up. The index is what
everyone actually looks in, so those 34 were invisible in exactly the lookup they
exist for; anyone searching would conclude they didn't exist. It had survived about
twenty careful re-measurements because of how people were measuring: count the rows,
compare with last time's row count, confirm your own addition landed. That tells you
nothing about a row nobody ever wrote. I've backfilled all 34, and put the check
that *can* see it into the file's header — compare the entries against the rows,
both directions. They now agree exactly, 1,756 each way.

The second: the coverage check's "known and accepted" list wasn't accepting
anything that had been annotated. Sessions had been writing a note beside each line
explaining why that lane doesn't need an entry — the most useful thing in the file
— and the note stopped the line working, so those lanes were reported as brand new
every single run. Twelve of the seventeen "new" items that morning were things
already decided weeks ago. That's the exact failure the check's own comments say it
exists to avoid: a report nobody reads because it's mostly noise. Fixed, and the
annotations now survive a rebuild of the list instead of being silently wiped.

Then the seven genuinely new lanes: one had built something worth registering (a
pre-commit detector that stops a tenth service being written with the same
undeliverable-reply bug — it exists because *fixing* the other seven properly needs
a formal review, so they shipped a guard instead), one had switched an existing
mechanism on for a second consumer (recorded on that mechanism's own entry, where
you'd look for it), one had already done its own paperwork correctly, and four are
site builds or lanes that haven't built anything yet.

On the duplicates: there were 441 documents that existed in multiple saved versions
— `RUNBOOK_travelling_docs.md` plus thirty-nine numbered copies of itself, and one
running-notes file with fifty-seven. 1,973 files in total for 441 actual documents.
I've deleted 1,339 of them, keeping the newest of each, and everything is still in
git if we ever want it back. Two things stopped this being a clean sweep and are
worth knowing: the *unnumbered* copy is sometimes the newest one, not the oldest
(so a naive "keep the highest number" would have deleted a live document and kept a
six-week-old copy), and the numbering doesn't always run in date order. I kept both
copies wherever there was any doubt, and I checked afterwards that every document
path referenced by our scripts still resolves — seventeen didn't, and all seventeen
were already broken before today.

The cost, which I'd rather say than have you find: 43 of the deleted files are cited
as sources in register entries, so those particular citations now only resolve
through git. There's a note in the landmines file telling whoever hits one how to
pull it back, and warning them that a missing file is not the same as a fabricated
citation.

**The watcher, and why it isn't a hook (2026-08-04, later).** You asked whether the
watcher could live in the framework rather than in a session's local tooling. It
can, and there was already a working example to copy: the sweep that checks open
bug files for stale citations runs as a scheduled job in the cluster, reads the
repository through GitHub, and writes what it finds into the notes table. Two other
jobs in the same family run daily. So the register's watcher is now one of those —
a daily job that reads the register, compares its two halves, and writes a verdict.
It reports and never repairs: filling in a missing row needs a sentence written by
someone who understands the concept, and a generated one would be worse than the
gap because it would look authored.

Worth saying why it isn't a commit-time check, since that's the obvious place. A
commit-time check only runs for the person editing the register — the one person
most likely to have got it right — and never for what accumulates in between. The
problem was never that people skipped a check. Around twenty sessions did exactly
what the file told them to do, and none of them could see the gap, because the
check they were told to run compares this week's count with last week's count.

Two honest notes. First, I got an assertion wrong while writing the test: I assumed
that if the register was broken, every number in it would look wrong. It didn't —
the headline was perfectly accurate while thirty-four concepts were missing. The
test now checks for that agreement, which is a better test than the one I meant to
write. Second, the very first run against the current state wasn't clean: another
session, three hours earlier, had copied the wrong number into the headline — there
are three counts sitting in adjacent lines of that file and they answer different
questions. That's now corrected, the misleading one is labelled, and it is the
best argument for the watcher I could have asked for: the rule had been tightened
that same morning, in that same file, and it still went wrong before the day was
out.

It's built, tested and committed, but **not deployed** — you said you'd run that.
One thing worth knowing: this afternoon's `make release` won't have picked it up.
That target deploys a fixed list of services and a new scheduled job isn't on it,
so it needs its own one-line deploy.

**Retiring the counts (2026-08-09).** You said retire the hand-maintained
headlines, so they are gone — from the index, from all but one of the 109 category
files, and from the README.

Two numbers made the case. The index had **four** different commands that count it,
all four giving different and individually-correct answers, sitting in adjacent
lines of the same file: twice in four days a careful person took the wrong one and
published it. And the per-file counts, which nobody had ever checked, were wrong in
**32 of 109 files** — ninety concepts of drift in total, one file claiming five
concepts while holding none. That is not carelessness; it is what a number written
down beside its own near-identical rivals will always do.

The count is now worked out on demand — by the daily job, or by one command — and
written down nowhere. The old history of "previously it was N" is kept in full at
the bottom of the index, because how it went wrong is worth more than the numbers
ever were. The one rule that actually mattered is now the only one left: when you
add a concept, add its index row in the same commit.

The watcher's own check for this had to be turned around rather than deleted. With
no number left to compare against, it would have found nothing and said nothing —
which looks exactly like everything being fine. It now reports any count that comes
*back*, and there is a test that puts one back to prove it still fires.

Two things worth telling you from its first four days. It caught a wrong number
three days running and **nobody acted on it** — the report was right and unread,
which is a caution about reports in general. And the original problem came straight
back: a concept was filed on the 8th with no index row, four days after we
backfilled thirty-four of them. So this isn't a backlog that got cleared, it's a
leak, and now we know roughly how fast it drips.

One file was left alone deliberately: another session has had it half-edited in the
shared tree since the 4th, and tidying my one line would have swept five days of
their work into my commit. It's noted as owed, and the watcher will keep naming it
until someone clears it.

The job's script changed today, so it needs a re-deploy to pick up the new check —
one command, same as before.

---

**2026-08-10, later in the day.** The leak is now stopped where it starts.

Yesterday's note ended by saying the register had become self-monitoring but was
still dripping: concepts arriving with an entry and no row in the index, roughly
one every day and a half, each one invisible to anybody searching for it. The
question left open was whether to build something at the moment of writing, or
carry on watching. That's now decided and built: when a commit adds a concept and
forgets its index row, the commit says so, in front of the person making it, before
it lands. It also catches the other half of the same problem — claiming an ID
number another lane has already used, which happened three hours apart between two
sessions on the 8th and cost us a permanent explanatory note in the register.

Before wiring it in I checked it against the last fortnight, one commit at a time,
because a check nobody trusts gets switched off. Of the commits that added a
concept in those two weeks, **84% did it correctly and the check stays completely
silent on them**. The 16% that didn't produced 34 findings and, on inspection, not
one false alarm.

**But the number that actually settles it isn't the hit rate — it's how long each
one sat there.** A hit rate can't tell you whether people were fixing these
themselves five minutes later, in which case the check is just nagging. So I
measured the wait: **the median concept sat for 93 hours — nearly four days —
before its row appeared**, and twenty-one of them were eventually swept up in one
go by a different session doing a clean-up, not by the person who wrote them. That
is the difference between an untidy habit and a cost somebody else pays.

**One correction to what I told you yesterday.** I said the leak was about one
every day and a half. It's about one a day. Not because anything was measured
carelessly — because the daily watcher can only report what is still missing at
ten to seven in the morning, so anything fixed the same afternoon never appears in
its report at all. Two of them in the last week were exactly that. The instrument
was quietly showing us a fraction of the problem, and nothing in the report said
so. It's the same lesson as the stored counts, in a different costume: a number is
only as honest as the thing that could have made it come out differently.

**And one mistake of my own, which is the more useful story.** The second half of
the check — the one that spots a reused ID — was dead when I first wrote it. Not
broken in a way that errors; dead in the way that looks exactly like working. One
of the git commands had its arguments in the wrong order, git refused it, and
because the tool only listens to a command's output and not its complaints, the
refusal arrived as "nothing found". No matches, no findings, no error — which is
also precisely what a clean register looks like. It passed every test I ran across
four hundred commits of real history, because those tests happened to exercise the
one mode where the argument order is legal, and not the mode that actually runs
when you commit. What caught it was deliberately staging a case that HAD to
produce a finding and refusing to accept silence. I've written it up as a trap for
everyone, since seven of our twenty-two little maintenance scripts listen to
output and ignore complaints in exactly the same way.

Two loose ends from yesterday, both now checked rather than assumed. The re-deploy
I said was owed has in fact already happened — the job running in the cluster is
byte-for-byte the current script, so it has been doing the full set of checks all
along. And the one file still carrying a stored count is still somebody else's
half-finished work in the shared tree; they last touched it on the 8th, so it's
live work rather than something abandoned, and tidying my one line out of it would
sweep their changes into my commit. Left alone again, deliberately, for the third
time.

Where that leaves us: the register's two halves now agree exactly — 1,817 concepts,
1,817 index rows — and the thing that kept pulling them apart is caught at the
moment of writing rather than reported the next morning. What still isn't answered,
and is now the only real question left, is whether the 1,817 entries are still
**true**. Nothing checks that. It's the next piece of work and it deserves its own
session.

---

**2026-08-10, end of the day — I checked staleness, as you asked.**

The short answer: the register is complete and agrees with itself, and **its
evidence is quietly ageing**. Nothing we have built so far would ever have told
us that, because coverage, drift and today's new gate all ask the same question —
does the register agree *with itself*. None of them asks whether it agrees with
**the platform**. That took an afternoon to find out and it is the real result.

Three numbers worth having.

**A hundred and twenty-nine entries name a specific version of the platform as
their evidence — and eighty of those name a version more than fifty rebuilds
old.** Two of them cite v1.0.407. We are on v1.0.1280. That is eight hundred and
seventy-three rebuilds ago. Those entries may well still be correct; the point is
that nobody reading one can tell, and the council seats we feed from these entries
certainly can't.

**About twenty entries say a thing is "not live yet" when it has since shipped.**
I proved two of them properly rather than assuming: I went into the running
containers and looked for the code. In one case the entry had been saying "not yet
live" for thirteen days and twenty-two rebuilds after it stopped being true. Both
are corrected, with the evidence written into them. The other eighteen I have
listed but deliberately **not** touched — each needs someone to go and look at the
specific thing that lane built, and writing "live" on an entry I hadn't checked
would manufacture exactly the false confidence I was sent to measure.

**Ninety-six citations point at files that no longer exist**, mostly the old
numbered documents we deleted on the 4th.

**Two mistakes of mine, both instructive.** The broken-citation figure was 187
before I sampled it — a good third were my own text-matching picking up fragments
rather than real paths, and the number looked perfectly publishable while being
wrong. And the count of "not live" entries came out at 38 from pattern-matching,
against about 20 when I actually read them. That second one changed the
recommendation, so it was worth the time: the reason no pattern can do it is that
the status field is doing four different jobs at once. One entry says "inert **by
design**", which is permanent and looks identical to a stale claim. One quotes its
own out-of-date wording inside the correction that fixed it. Several are half-live
— one thing, two states, two different clocks. And one is a chain of three
preconditions in order.

**So my recommendation is narrower than the brief.** Do not build something that
reads the status text and judges it; it will be confidently wrong in both
directions. Build something that keys only on things with no ambiguity — a version
number, a file path, a date — and have it say *"this entry's evidence has
expired"*, never *"this entry is wrong"*. That is exactly the restraint that makes
the daily drift check worth reading, and it is why people trust it.

And the cheapest useful thing isn't a checker at all: **make the version lag
visible.** A hundred and twenty-nine entries already carry the number, and what the
fleet is running is one command away. That alone would have surfaced everything
above.

Full detail, including the twenty-entry worklist for whoever picks this up, is in
`FINDINGS_2026-08-10_staleness_survey.md`.

---

**2026-08-10, later — the new build landed, and it settled this afternoon's list.**

The chassis that just rolled is the first one that **says which version of the code
it is**. That was BLD-019, built this morning and sitting in the register marked
"not live yet"; it is live now, and I read the answer straight out of both running
containers to be sure.

It matters more than it sounds. This afternoon I gave you a list of about twenty
register entries claiming "not live yet" about things that had almost certainly
shipped, and said each one needed somebody to go and look inside the running
system for a specific piece of code. That whole exercise is now a single command:
the container tells you which commit it was built from, and you ask git whether the
entry's own commit came before it. I checked the command could actually say "no"
before trusting it to say "yes" — it can.

Every one came back shipped. So nineteen entries are corrected, and I was careful
about what the correction claims: it says **the code is in the running system**, and
it deliberately does *not* say the feature is working. For several of them it isn't
— they are waiting on a database change or a configuration switch, which is a
different thing on a different clock, and blurring the two is exactly the mistake I
found in another entry this afternoon.

One new thing worth knowing, because it is cheap to fix going forward: **thirteen of
the entries I looked at never wrote down which commit they came from**, so even the
new stamp can only guess at them from a date. That is nine characters at the time of
writing that turn an unanswerable question into a one-line check. I'd suggest it
becomes part of the same commit-time check we added this morning rather than another
daily report — same reasoning: catch it where it is written.

What is still open is the other three-quarters of the staleness picture, none of
which a new build touches: eighty entries citing a platform version fifty or more
rebuilds old, ninety-six citations pointing at deleted files, and a hundred and
fifty-six pointing at bugs that have moved.

---

**2026-08-11 (evening) — the check we shipped yesterday was working perfectly and
nobody could hear it.**

Yesterday I left a note at the top of the handoff saying the new commit-time check
had fired within three hours of going in, and had been ignored — an entry went into
the register with no index row anyway. I also said the first thing to work out was
*which* kind of failure that was: did that session never see the warning, or see it
and decide the row could wait? Those need opposite fixes, and I didn't think
anything on record could tell them apart.

Something could. Every session on this machine keeps a transcript of what it ran and
what came back, and I had never thought to look in one. The commit in question ended
with `| tail -8` — a perfectly ordinary way to trim a noisy command down to the bit
you care about — and what came back was exactly eight lines long. The warning had
been printed and thrown away before anyone could read it. Nobody ignored anything.

Then it got worse, in the useful way. The pre-commit warnings print *first* and
git's own "here is your commit" summary prints *last*, so trimming to the last few
lines keeps the receipt and discards the warning, every single time. I swept every
commit made through the tool since we added the scope report on 18 July: **of 2,669
commits touching more than one file, 1,199 — forty-five per cent — never showed the
session a word of it. Ninety-five per cent of those were cut by the session's own
`tail`, across 258 different sessions.** This was never one lane's bad habit.

And it was not only my register check going missing. The same trimming removes the
"commit scope" block, which is the main thing standing between us and one session
quietly committing another's half-finished work — the exact problem that whole
section of CLAUDE.md exists for — plus all seventeen of the automated code checks,
plus the architecture warning. All computed correctly, all binned. The one hook that
*did* survive is the council-trailer nudge, purely because it prints later, which is
why everything has looked healthy and quiet from the outside.

The fix is in and it is small: after any commit, a hook re-runs those two reports
against the commit that was just made and hands the result to the session directly,
outside the command's output, where no amount of trimming can reach it. I checked
that no ordinary git hook could do this — I tried the obvious one in a scratch
repository and its output gets cut by the same trim.

One thing I want to be straight about, because it cuts against what I wrote
yesterday. My note said this looked like the check being ignored and that we should
watch for a week before anyone argued for giving it teeth. **That reading was wrong,
and the argument for teeth is now weaker, not stronger.** The evidence that looked
like people ignoring a warning was an artefact of a pipe. Nothing here says a
blocking check is needed; it says we were measuring the wrong thing.

I also got the diagnosis wrong on the first pass, which is worth recording. My first
theory was elegant — that our own house rule for committing (naming files explicitly)
was hiding those files from the checks. It took about a minute in a scratch
repository to prove that git handles it properly and my theory was false. The real
cause was duller and was sitting in plain sight in the command itself. I went for the
interesting explanation before I read what had actually been typed.

Still open, and untouched by any of this: the other three-quarters of the staleness
picture from yesterday — eighty entries citing a platform version fifty or more
rebuilds old, ninety-six citations pointing at deleted files, a hundred and
fifty-six pointing at bugs that have moved.

---

**2026-08-12 — the thing I built to check yesterday's fix was measuring the wrong pipe**

Yesterday I found that about half of all commits never showed the session its warnings, because
people trim command output down to the last few lines and the warnings print at the top. I fixed
it by having the warnings delivered separately, outside the command's output, where trimming
can't reach them. And I wrote down a command to run the next day to check it had worked.

I ran that command this morning. It said delivery had **fallen** — thirty-eight per cent, against
the fifty-five per cent I started with. On the face of it my fix had made things worse.

It hadn't. The checker was looking in the wrong place. It counted warnings that appeared in the
command's own output — and the whole point of the fix was to stop sending them there. Everything
the new route delivered was being counted as a failure. When I taught it to look at both routes,
the real number for today is **thirty-six commits out of thirty-six, a hundred per cent** — and
twenty-three of those thirty-six were reached *only* because of yesterday's fix. It is working
exactly as intended.

I want to be blunt about how bad that near-miss was. The wrong number didn't just fail to confirm
the fix — it pointed confidently the other way, on a day the fix was flawless. If I hadn't
happened to distrust it, the next person to look would have read "the fix failed", or worse,
"we clearly need to start blocking people's commits" — which is the exact conclusion I spent
yesterday arguing *against*, on good evidence. A broken instrument doesn't go quiet; it keeps
producing numbers, and they keep looking like findings.

What made it slip through is worth remembering. I had four separate safeguards on that checker,
and every one of them passed. All four were checking the *old* delivery route. Nothing was
watching the new one. The one-line version of the check I should have run takes a second: search
the checker for the name of the thing the fix writes to. It wasn't in there at all.

Two smaller stumbles inside the repair, both of which briefly looked like real problems. The
first: the two halves of the data write the commit reference at different lengths, so my initial
match found nothing — which looks identical to "the fix never ran". The second: one of my own
safeguards reported twenty-four deliveries from *before* the fix existed, which reads as a broken
count. That was me comparing British time against the recording's UTC. An hour's difference.

One more thing I nearly got wrong, and this one is about honesty rather than plumbing. Now that
delivery works, the obvious next question is whether being *told* actually changes behaviour. The
signal looks perfect — not one missing register row across all seventeen relevant commits since
the fix. I was about to report that as good news. It isn't news at all: only **four** commits in
that window even added the kind of entry that could have gone wrong, and at our historical error
rate you'd expect a clean sweep half the time by luck alone. It needs about fourteen before it
means anything, which is days away. So: delivery is proven, and the behaviour question is open
and stays open. I've written that on the record in place of the all-clear I nearly gave.

Untouched, and still the real backlog: the eighty entries citing a platform version fifty or more
rebuilds old, the ninety-six citations pointing at deleted files, and the hundred and fifty-six
pointing at bugs that have moved.

---

**2026-08-12 (later) — the version-lag job, and why measuring it first saved building the wrong thing**

Yesterday's note left three jobs. The cheapest was supposed to be this: lots of register entries
say "this was true as of platform version 1283", the platform is now on 1290, so show me the
entries whose number is old. Eighty of them are fifty or more versions behind. It sounds like a
morning's work and an obviously useful list.

I measured it before building it, and the obvious list would have been mostly noise. The problem
is that a version in an entry means one of two opposite things, and they look identical. "Shipped
in version 1029" is a fact about history — it will be true for ever. "Both servers on version
1218 gave the right answer" is a *check*, and checks go out of date. I tried to tell them apart by
the words around them and failed on three quarters of the cases. A list of a hundred and eleven
"stale" entries where most are permanent facts is a list nobody reads twice.

What worked was already sitting there. The register writes each entry in labelled parts, and two
of those labels — the status, and the evidence for the status — are *by convention* claims about
how things are right now. The rest is description and history. Keying on the label instead of the
sentence needs no cleverness at all, and it cuts the list by a quarter straight away. It also
turned up something I didn't expect: the *evidence* lines are far staler than the *status* lines.
We update what we claim; we don't go back and re-check why we claimed it.

Then one genuinely sharp result. Some entries quote a container version as proof — "this agent is
running version X". I checked what the live system says: all 187 agent records carry the current
version, uniformly. The release rewrites them. So quoting one of those numbers only ever records
the day you looked; it expires at the next release, guaranteed. That gave me a short, precise list
instead of a long vague one — and two entries got fixed off it the same afternoon.

The nicest illustration is a pair. Two entries both cite version 407, from last November, which is
883 releases ago. One is **wrong** — it says a live record "still references" that version, and
that record now says 1290. The other is **perfectly fine** — it describes what a setup file in the
repository says, and that file does still say 407. Same number, same day, opposite verdicts, and
the only way to tell is to know which thing is being described. So the tool names the category and
refuses to pass judgement, and prints the one-line check instead. I'd rather it stayed trusted
than became clever.

Three things I got wrong on the way, and the pattern in them is the same. First, I asked the
database the wrong question and got a clean zero — the version isn't stored where I looked. My
safety check also returned zero, which is the only reason I noticed. Second, I searched for the
demo agent by the word "HITL" and reported here that it wasn't installed. It is installed; its
name simply doesn't contain "HITL", and the group it belongs to is filed under a human-readable
title rather than the code name. I searched for the name the *documentation* used instead of the
name the *system* uses. Third, and the one I'd most like to remember: my own tool's output made me
think it was broken. It was printing the start of each line rather than the bit it had actually
tested, so three correct hits looked like mistakes, and I was about to loosen a detector that was
right nine times out of nine. Show the evidence you actually tested.

Two jobs left from the three: ninety-six citations pointing at deleted files, and a hundred and
fifty-six pointing at bugs that have moved. The question I'll take into both is the one that
worked here — is there something to key on that doesn't involve reading English?

---

**2026-08-12, later the same day — the last two jobs on the staleness list, and the answer was no.**

The question I said I'd carry into these two was whether there's something to key on that
doesn't involve reading English. For the version numbers, the answer was the register's own
labelled parts: two of those labels are, by convention, claims about how things are right now,
and keying on the label rather than the sentence is what made that check trustworthy.

**It doesn't transfer, and the reason is simple once you say it out loud.** Whether an entry's
citation still works has nothing to do with which part of the entry it sits in. A file gets
renamed and every citation of it breaks, wherever it was written. The numbers say the same
thing — between one in ten and one in four citations fail to resolve in every part of the
entry, with no clean line anywhere.

**What works instead is asking git, and it turns the whole picture around.** Instead of "does
this file exist right now", ask "can git still find this file" — at its stated place, somewhere
else today, deleted but recoverable, or never there at all. That's four answers, it needs no
English, and two of them tell you exactly where the file went.

Put that way, the register's citations are in far better shape than the survey suggested. Of
nearly eight thousand, three quarters work exactly as written. Eight hundred point at something
git can still find, usually because a big documentation folder was reorganised on the 4th and
took a lot of citations with it. Another seven hundred and sixty-nine give only a filename with
no folder — several files share that name, so the citation is vague rather than wrong. **And
four — four in the whole register — name a file that has never existed under that name.** Three
of those four are in the part of an entry that lists things to go and check later, which is the
mildest place for it. The fourth cites a bug report by the right number but the wrong folder,
date and title, which is the one I'd want its author to see.

**Now the part I got wrong, because it is the more useful half.** Earlier today I told you two
paths accounted for most of the dead citations, and that fifteen entries pointed at our
most-read debugging document at an address that never existed. That was false, and my own tool
invented it. These files have a number in brackets in their names — `(3)`, `(6)` — and in some
citations that bracket is a reference marker rather than part of the name. I stripped it
everywhere. So I took correct citations, altered them, found the altered version missing, and
reported the register as broken. Because those findings came out sorted by how often they
appeared, the invented ones went straight to the top.

Then I fixed it and made the same mistake again one level down: the corrected code checked both
spellings but kept the *second* answer instead of the *better* one, so the same fifteen entries
came out wrong a second time. **The count didn't budge between two completely different bugs** —
which is worth remembering, because a number that comes out the same twice feels confirmed, and
both runs were only agreeing with my own instrument, not with the world.

There was a third, caught earlier and more cheaply: the command I first used to ask git "has
this path ever existed" quietly leaves out files whose contents are identical to another file's.
Seven hundred and ninety-one files at the current version were missing from its answer, every
one of them a duplicate. That one never reached a number, because I'd written a check that
compared its answer against a list it should have contained — and the check failed on the first
run. It's written up as a trap for anyone else asking git that question.

**So the pattern in all three, and it is the same pattern as yesterday's blind checker.** Each
time, the tool changed the thing it was checking before checking it, and the change is invisible
in the output. Yesterday the instrument was looking at the wrong channel; today it was looking
at a tidied-up version of the evidence. The defence is the same one that saved the version-lag
report last time: show the evidence you actually tested, not your cleaned-up restatement of it.

Both remaining jobs are now closed, and the tool that closed them says what it can't judge as
loudly as what it can. What I have deliberately not done is repair the eight hundred fixable
citations — an automatic rewrite of eight hundred lines across a hundred files is precisely the
change nobody can review, and those citations were correct when they were written. If in a
month nobody has repaired any of them by hand, then the honest answer isn't a louder report, it's
a small rule at the point where citations are written: cite a path, not just a filename.

**One thing you should know that isn't about the register.** While I was writing the three
documents above, another session ran a command that throws away every unsaved change in the
whole project at once, and it took them with it — along with everyone else's. I had my own text
in front of me so I simply wrote it again, but the wider loss is real: half a dozen other files
were being worked on, one of them since the 30th of July and another since the 3rd of August,
and all of those edits are gone. There is no way to recover them, because they were never
committed — that is exactly what "committed" buys you here. Nothing is broken and nothing needs
undoing, and I'd rather you heard it from me than found a colleague's work missing tomorrow. The
one silver lining is unpleasant: a file four handoffs in a row have described as "someone else's
half-finished work, leave it alone" is now clean, so the last outstanding item on the register's
own health check can finally be closed — by a route nobody would have chosen.

> **Correction, 2026-08-12, an hour later — I was wrong about the lost work, and it is not lost.**
> I told you above that another session threw away everyone's unsaved changes and that there was
> no way to get them back. What actually happened is that a session ran `git stash`, which is the
> command for *setting work aside safely* — it clears the desk but keeps everything in a drawer.
> All of it is still there: thirty-eight files, including the ones I rewrote and the one file
> four handoffs have described as someone else's half-finished work. Nobody has lost anything,
> and the colleague whose work it is can take it back whenever they return.
>
> Why I got it wrong is worth a sentence, because it is the same failure as the rest of today.
> Git keeps a log of what happened to the project, and stashing writes a line in that log which
> is indistinguishable from the line written by the genuinely destructive command. I read the
> line, recognised the destructive meaning, and reported it — without running the one command
> that would have shown me the drawer was full. Another session had already worked it out
> correctly and written it down; I saw their note in passing while checking my own work had
> saved.
>
> So: nothing to worry about, nobody to apologise to, and one more instance of the thing I keep
> saying today — I checked a proxy for the claim instead of the claim.

---

**2026-08-16, register lane — three things since the 12th, in plain terms.**

First, the stash. On the 12th one session ran `git stash` on our shared tree and it quietly
swept up two days of everybody's unsaved work — thirty-eight files across about ten lanes —
including the release manifests, which it put back a hundred releases while leaving the tree
looking perfectly clean. I put everything back and wrote it up. On the 14th you asked me to
forbid it, and to ban it in git itself if we could. We can't — git has no way to intercept a
stash before it happens — so the ban lives in the harness every session runs through instead:
a small gate that refuses any stash that would change the tree, and lets through the two
read-only forms we need for recovering from the one that already happened. I proved it live by
trying a stash and watching it get refused. Since then, in over six hundred commits, nobody has
managed to make a new one. That is the ban working.

Second, the landmine keys. When I filed the stash landmine I tested whether it would actually
reach anyone, and it wouldn't: the tool that reads our landmine file was splitting the "which
files does this concern" line on commas, and most of the file uses a little dot separator
instead. So dozens of entries were being delivered under a key nothing could match — present,
synced, and inert. You asked me to check the fix was right and then do it. Checking it made it
bigger: the dot problem had grown from 59 entries to 63 in two days, and there was a second,
larger problem of the same kind hiding behind it (commas inside brackets), so 185 of 482
entries have now been re-keyed. The check also caught something in the sync tool itself: it
decided whether an entry needed rewriting by counting rows, and six of those entries changed
their keys without changing their count — they'd have stayed wrong forever, with every sync
reporting clean. That's fixed too. And my own first attempt at the fix was wrong in a way the
new self-test caught before it touched anything.

Third, today. I fired the verifier at the three landmines I owed it — two of them had been
armed since the 12th and I never sent them, which is exactly the trap you had corrected in
CLAUDE.md the day before. Reading back the verdicts is the next session's first job.

One thing to know: the register has a new drift finding as of this morning — a row promising a
concept whose entry hasn't been committed yet. It's the exact failure the 12th's handoff
predicted (the row rides out inside someone else's commit), it belongs to a lane that is live
right now, and it should close itself when they commit. I've left it to them and flagged it.

---

**2026-08-25.** Picked this lane up cold and re-ran every measurement in the handoff before
touching anything, which turned out to matter — the tree had moved about fifteen hundred commits
since that document was written, and roughly a thousand more while I worked. Nearly everything was
still healthy: the landmine keys are clean, nobody has run the forbidden `git stash` in over two
thousand commits, and the fleet is on one uniform version.

One number had moved the wrong way. The citation report has a category it states very plainly —
"no file, ever, under that name" — and it had gone from four to seven. That is the report's
sharpest signal, the one the last three handoffs have quoted as proof the register isn't rotting,
so three new ones was worth a look.

None of the three was rot.

Two of them were the same thing in disguise. An entry cited a Go file that genuinely exists on
disk, but the lane that wrote it hadn't committed it yet — and the report resolves paths against
git's history while reading entries from the working folder. So a brand-new file reads as one that
has never existed. I predicted on the 20th that it would clear itself when that lane committed, and
it did, in exactly the way expected. Nothing to fix.

The third was a real bug, in our own report. When an entry cites two line numbers in one file,
there are two ways people write it, and the code only understood one of them. Given the second
form it mangled the path into something git has never held, then reported that as a file that
never existed. So the tool was manufacturing a small amount of the very rot it exists to detect.

The fix is one character. What took the time was making sure the test I added could actually fail:
I put the old broken code back on a scratch copy, kept the new test, and watched it fail in exactly
the way the bug behaved. This lane has been caught before by a test that passed because it wasn't
really testing anything, and a test that has never been seen to fail is only a comment.

Worth saying plainly, because it would be easy to overclaim: the count went from seven back to
four, but only one of those three was my doing. The other two fixed themselves. And the honest size
of this fix is a single citation in the whole register today — I did it because a decisive-sounding
number was quietly capable of being wrong, not because much was broken.

Still outstanding, and now for the eighth session running, is the small stored count in one
register file that can't be committed without dragging another lane's unfinished edit along with
it. It is once again the only thing the drift check complains about. I checked it properly rather
than assuming a clean-looking status meant it had gone away.

---

**2026-08-31.** The small job that had been top of this lane's list for four handoffs is
done. The idea is simple: when a register entry says "this is built but won't actually work
until the next deploy", it should say *which* change it is talking about. Nine characters.
Without them, anyone reading later has to guess, and there is now a one-line command that
answers "did this ship?" exactly — but only if the entry named its change.

So there is now a check that says so at the moment someone writes such a line, rather than a
report that mentions it the next morning to nobody in particular.

Two things about building it are worth your time, because both were near misses.

The first is that I measured it wrong and got a beautifully clean answer. I ran the check
across forty-five recent commits and it found nothing at all. That is exactly what you would
hope for — a quiet check that does not nag. It was completely false: my counting script was
looking for the wrong thing and would have reported zero no matter what. I only caught it
because I ran a case I already knew *should* trip it through the same counting script, and
watched it come back clean too. A zero that cannot tell "nothing is wrong" from "I am not
looking" is worth nothing, and this one looked like good news.

The second is worse and I nearly shipped it. The check lets you off if you have cited your
change, and it recognised those by their shape — a string of a certain length made of certain
characters. It turns out the council review system stamps its own reference numbers in exactly
that shape. So entries that had merely quoted a *review* reference were being waved through as
though they had named their change. Seven of the eight real cases were being let off by this.
The check would have sat there looking quiet and useful while catching almost nothing. The fix
is to ask git whether the thing actually *is* a change, which takes one command and settles it.

Both near misses have the same shape, which is why I am writing them down rather than just
fixing them: **a check that is broken tends to report good news.** Silence is what you expect
from something working. So the only way to trust quiet is to prove the thing can still shout.

The work is committed, it is registered so other lanes can find it, and it has gone to the
review council — which it now needs, because that file came into the council's remit a week
ago. I have not held the code back waiting for the verdict; the estate has a specific marker
for "submitted, not yet judged" precisely so nobody sits on work.

---

**2026-09-03.** The review council approved the check from last week, first time round, with two
comments. Both were worth reading and neither was quite right, which is itself the useful part.

The first asked me to make the check safer, on the grounds that if it ever crashed it would stop
everyone in the building from committing anything. That would be serious if true. It is not — the
script is deliberately run in a way that ignores its failures, and the reviewer said in its own
notes that it had not checked and was going on a hunch. But it was still worth doing, for a
different reason: a crash would silently switch off the twenty-three *other* checks that run
alongside it, and they would report nothing rather than reporting a problem. So I made the change
and wrote down that the reason given was wrong, so nobody argues it again from the wrong end.

The second said one of my supporting facts was contradicted by our own records. I went and read
them. They say the opposite of what the reviewer quoted — it had picked up a warning about an
unrelated *technique* and read it as a warning about the *thing*. Our records in fact state my
argument almost word for word. So I have noted that in the register too, because an objection
left lying around gets re-made.

Three of the reviewers made the same fair criticism: I told them my change plugged into existing
machinery without showing them that the machinery was there. I had checked — I built and tested
the thing before submitting — I just did not put the evidence in front of them. Cheap lesson.

**Your roll is the interesting part.** Now that a fresh build has gone out, a hundred and ten
entries in the register still say "this is built but won't work until the next deploy". The next
deploy has happened. Eighty-one of them say which change they mean, so each can be settled in one
command — in or out, no judgement required. **Twenty-nine do not say, and those cannot be settled
at all.** That is exactly the gap last week's check now prevents growing, and it is the clearest
demonstration I could have asked for of why it was worth doing.

Clearing those is the first job in the new handoff. It is reading work rather than clever work,
and it is bounded — which makes it a good one to pick up cold.
