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

