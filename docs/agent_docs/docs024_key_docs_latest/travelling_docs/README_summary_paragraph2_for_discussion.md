● This is a milestone worth stating plainly: the tool verification ladder is now whole and self-driving.

What just happened

tool-acceptance-agent drove tool-xp-curve-designer from end to end with no human in the loop, and wrote the first machine-authored acceptance-run note: "Tier-4 acceptance PASSED — all 3 evaluated checks passed in
headless Chromium (2 skipped: mobile-fit, curve-switch)," verdict passed=3, failed=0, no improve_tool item. The orchestrator produced results byte-identical to the T6 hand-produced smoke, which confirms the full
chain works as one system:

PLAN criteria → tool-acceptance-agent → request_browser_run (Kafka) → browser-runner-adapter (Chromium on the live page) → reply → judge_acceptance_results → doc_notes.

On the deploy gate: I flagged that the commit-message ancestry said v1.0.1108 predated the actions — but this repo reuses version tags, and the pod started at 16:11 versus the 12:37 commit, so the tag had been
rebuilt from HEAD. The run itself was the decisive test (it would have failed at request_run in seconds otherwise), and it passed. I've banked "verify by pod-start-time, not commit-message ancestry" for this repo.

Where the ladder stands

- Tier 0 — generation-time completeness ✓
- Tier 1 — structural (check_tool_health) ✓
- Tier 2 — static acceptance (tool_acceptance check, the anchor rule) ✓ live
- Tier 4 — behavioural, headless browser ✓ now self-driving

The arc that began with "the composer invented #xpTableBody" ends with a machine that renders the real tool in a real browser and records whether it works — the exact capability that would have caught the
economy-simulator bugs with no one watching.

Open, for future sessions

1. The fail path (acceptance-fail note + improve_tool item → tool-improver loads PLAN+NOTES → fix → re-run) is unit-tested but not yet exercised live; the first genuinely-failing tool will close that loop and
   demonstrate the full iterate-until-criteria-pass cycle.
2. Trigger wiring — fire acceptance automatically after tool creation/recreation deploys, after any improve_tool completes, and on a periodic sweep.
3. P1 mobile / P2 interactions — the interaction tier is the one that asserts a tool calculates correctly, not just boots.

All docs are current: RUNBOOK position line updated to "Tier-4 self-driving," running notes carry the proof, handoff turn log at T8, memory updated. Uncommitted in your tree for the next release:
tool_acceptance_actions.go + tests, the registry edit, migration 145, and the 087 trigger.
----

The summary doc

OVERVIEW_self_verifying_tools.md — a plain-language explainer written to be talked from, not to be a spec. It opens with a one-sentence version, states the two failure modes this solves (lost intent; "deployed ≠
works"), explains the two mechanisms (travelling docs, the verification ladder) with the tier table, walks the autonomous loop as a diagram, lists what's actually proven vs designed, and ends with the design
principles worth repeating and where it's headed. It name-checks the concrete wins (the game recreated with both bugs fixed, the hallucinated selector caught on the first sweep, the self-driving browser run)
without drowning in codenames. If you want a shareable web version for a talk, I can render it as an Artifact — just say the word.

The fail path — proven live, then cleaned up

This was the last piece of the story: what happens when a tool actually fails. Using the runbook's own smoke-test precedent (prove, then clean up), I temporarily added one genuinely-failing criterion to
drop-rate-tuner's PLAN — a selector I confirmed absent from the live page — and fired acceptance. The verdict came back failed=1, failing_checks=[failpath-proof], improve_tool_created=true, and both artifacts were
exactly right:

- an acceptance-fail note: "Tier-4 acceptance FAILED — 1 of 4 evaluated checks failed in headless Chromium: failpath-proof: no element matches #zzz-failpath-proof in the live DOM after settle"
- an improve_tool ticket routed to tool-improver, carrying the criteria as acceptance_test and listing the failing check

Then a complete, verified teardown: ticket cancelled (it was still detected, so tool-improver never touched it), PLAN reverted to the exact 3046-char / 5-check original with no asset check, the test note deleted,
zero orphan references left. I checked the fence precisely afterward to be sure nothing was corrupted.

So the loop is now demonstrated end-to-end both ways: pass → acceptance-run note; fail → acceptance-fail note + a fix ticket the existing pipeline already knows how to dispatch. That's the headline you can now say
with a straight face: the platform documents its own tools, tests them in a real browser against their own definition of working, and files its own repair ticket when they fall short — no human in the loop.

What's genuinely left

1. A real failure, end to end — let a non-manufactured failure flow into tool-improver and back through repair → re-verify. That's the one link I proved with a synthetic input rather than a real one.
2. Continuous triggering — fire acceptance automatically after every creation/recreation/repair, plus a periodic sweep, so it runs without a hand on the trigger.
3. P1 mobile / P2 interactions — the tier that asserts a tool computes the right answer, not just that it boots.

All docs are current (overview written, RUNBOOK position line, running notes, handoff at T9, memory). Uncommitted for your next release remain the Tier-4 agent code, migration 145, the trigger, and now the overview
doc.

-----

Continuous acceptance is built and wired — the top tier of the ladder no longer needs a hand on the trigger.

What landed this turn

The tool_acceptance_due discovery check — the periodic trigger that makes Tier 4 continuous. For every active tool with a deployed page and a current PLAN carrying criteria, it queues one acceptance_run work item for
tool-acceptance-agent, with three guards: a 7-day cooldown after any verdict (a recent fail already has its fix ticket in flight — re-running would just duplicate the verdict), a skip when a run is already open, and
the criteria-fence precondition (no criteria is Tier-2's needs_criteria concern).

Three design calls worth knowing:
- Post-creation/post-improve hooks deliberately not used. Both fire before the page actually redeploys — creation leaves the page planned, and the improver merely queues a rerender — so they'd test the old page. The
  sweep only ever sees deployed pages, so the timing is right by construction. Event-driven hooks can come later, placed after deploy completes.
- Items emit straight to triaged. We observed detected items sitting unswept on this site for days; acceptance needs no human judgment, and create_rerender_items set the precedent.
- Priority 90 places acceptance after builds and rerenders in the queue, so it tests the new page, not the one about to be replaced.

I also verified the dispatch loop's input_mapping passes spec whole — so the handler receives input_data.spec.function, exactly the 145 contract — and applied the recorded follow-up while in the file: the Tier-2
checker's cooldown now excludes cancelled items (a cancelled item means the finding was resolved another way, so the tool is re-checkable immediately).

Migration 146 applied through the runner (snapshot taken, checks list extended, pipeline note written). Safe pre-deploy as always — the running binary warn-skips the unknown check name.

The one gate, and the proof that follows

The check and the cooldown fix ride the next chassis image. Once it deploys, the full-autonomy proof is simply: trigger (or wait for) a design-discovery sweep on gamesdesign → an acceptance_run item appears → the
dispatch loop drives tool-acceptance-agent → a fresh acceptance-run note lands in the travelling docs — with zero manual triggers anywhere in the chain. That will be the first completely unattended run of the whole
mechanism, discovery to verdict.

Docs current: running notes, handoff at T10 with the gate and proof recipe, memory updated. Uncommitted for your next release: the new check, the cooldown fix, and migration 146's file. Remaining open after this: a
real failure flowing through tool-improver and back (waits for reality to break something), then P1 mobile / P2 interactions.


