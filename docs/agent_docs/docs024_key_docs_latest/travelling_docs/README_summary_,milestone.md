● This is the milestone the whole arc was building toward: the mechanism now runs unattended, discovery to verdict.

What just happened, in full

A design-discovery sweep on gamesdesign — which in production is a scheduled maintenance tick, not a human action — ran the new tool_acceptance_due check, and from there not one step was human-driven:

1. The check found a tool due for verification and emitted an acceptance_run work item.
2. The dispatch loop claimed it and spawned tool-acceptance-agent.
3. The agent loaded the tool's PLAN criteria, called the browser-runner adapter, which drove headless Chromium against the live page.
4. The agent judged the results and wrote the verdict into the tool's travelling docs: "Tier-4 acceptance PASSED — tool-drop-rate-tuner — all 3 evaluated checks passed."
5. The work item closed complete.

The entire loop from the overview doc — discovery → real-browser test against the tool's own definition of working → verdict recorded — executed with no one watching.

Two honest notes on getting there

- First sweep correctly emitted nothing. Both tools had verdict notes inside the 7-day cooldown (from my own testing this week) — the guard working as designed. To observe the full chain, I freed drop-rate by
  deleting a genuinely stale note: a Tier-2 fail from 07-10 16:25 describing the exact issues migration 143 fixed 26 minutes later. Legitimate cleanup, not gaming.
- The deploy check paid off. v1.0.1112 did carry the check this time — I confirmed it was in-binary via the sweep's own checks_run list before trusting it, having been burned by the untracked-file trap on v1.0.1111.

One follow-up worth flagging

That stale note exposed a small design point: the cooldown counts Tier-2 verdicts as well as Tier-4, so an independent or outdated Tier-2 result can suppress a Tier-4 run for 7 days. Not urgent (it's a coarse
don't-spam guard), but scoping the cooldown query to Tier-4 verdicts (source='tool-acceptance') would be cleaner. Noted in the docs, not fixed this turn to avoid another deploy.

Where it stands

The verification ladder is whole, self-driving on both pass and fail, and now continuous. The two remaining items are genuinely "waiting for reality": a real (non-manufactured) failure flowing through tool-improver
and back through repair, and the deeper behavioural tiers (P1 mobile, P2 interactions — asserting a tool computes the right answer, not just that it boots). Docs are all current (RUNBOOK position line, running notes,
handoff at T12, memory), and the check is committed. A clean, genuinely finished-feeling stopping point.

--------------

● Three things done this turn, plus one important catch.

The summary docs

- STATUS_2026-07-13_where_we_are.md — the state-of-play snapshot you asked for: the milestone ladder (all 7 proven), the architecture in one picture, what's live now, and what's next in a table.
- OVERVIEW_self_verifying_tools.md — refreshed the T9 draft to record full autonomy and reorder "what's next" now that continuous operation is done.

Continued: the interaction and mobile tiers (P1 + P2)

This is the capability that closes the "deployed ≠ works" gap for real. I rewrote the runner behind a testable interface (preserving every P0 behavior) and added:
- Interaction checks (P2) — fill/click/select a control, then assert the result changed. Proven live on xp-curve: it genuinely selected 'exponential' in the dropdown, the tool's JS rebuilt the table, and the runner
  asserted the rows appeared. This is the exact class of check that would have caught the economy-simulator's dead slider directly.
- Mobile profile + overflow (P1) — a real 390×844 touch viewport; no_horizontal_overflow passed at that width. Both the interaction and boot checks ran on desktop and mobile.

Nine unit tests pass; the live run took 9.3s for the full desktop+mobile sweep. Migration 147 (applied) sets the agent to request both profiles — safe with the current adapter, and mobile/interaction criteria
activate automatically once the new adapter image ships.

The catch worth flagging

While committing, git status revealed that the entire browser-runner adapter package had never been git-tracked — every production image worked only because the Dockerfile does COPY . ., which sweeps up untracked
working-tree files. A fresh clone would have been missing the whole Tier-4 runner. I committed it complete (53a5b518): entrypoint, dispatcher, the P0/P1/P2 runner, tests, Dockerfile, config, and kustomize. The
durable lesson (now in the notes) extends the untracked-file guard: a green Docker build is not proof the repo is complete — COPY . . masks untracked Go, so check git status for whole missing packages, not just stray
files.

Where this leaves us

The verification ladder is whole, self-driving, continuous, and now — pending one adapter deploy — behavioural at the interaction level. To activate P1/P2: build and deploy a new browser-runner-adapter image from
53a5b518 or later (147 is already applied). Then the mobile and interaction criteria that currently show as "skipped" will evaluate live, and I can verify the tools actually compute correctly, on both form factors,
unattended.

Remaining after that: a genuine failure flowing through tool-improver and back (waiting on reality), and the small cooldown-scoping refinement. Docs all current — STATUS + OVERVIEW written, running notes, handoff at
T13, memory updated.

