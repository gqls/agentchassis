I surveyed the whole platform rather than just the fix loop, and it reframes the project. Written up in reasoning_dataset/PLAN_capture_gaps_and_volume.md (628467b7d).

We were sizing this off the wrong thing

The fix loop is 445 rows. The platform has logged 40,785 LLM calls and holds 5,292 work items with terminal outcomes. The reasoning is already being produced at roughly 100× the rate we planned to harvest it.

What's missing isn't reasoning. It's the join between a decision and what happened next — and it's missing for a boring, fixable reason.

┌──────────────────────────────────────────┬──────────────────────────────────┐
│                                          │              today               │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ LLM calls logged                         │ 40,785                           │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ …carrying a work_item_id                 │ 2,894 (7.1%)                     │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ …joinable to a terminal outcome          │ 1,165 — already 15× the fix loop │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ items ever independently verified        │ 9                                │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ human decisions with the reason recorded │ 0 of 316                         │
└──────────────────────────────────────────┴──────────────────────────────────┘


The gap that matters most

llm_call_log already has a work_item_id column and the logger already populates it when it's present. It usually isn't:

- content-quality-auditor — 7,119 calls, 0%
- visual-design-auditor — 4,032 calls, 0%
- site-review-agent — 3,987 calls, 0%
- feed-triage — 423 calls, 0%
- tool-recreation-handler — 181 calls, 100%

That last row is the tell: it's plumbing, not physics. ~15,500 calls where an agent formed a verdict and we can't tell whether it was right. One field threaded through four agents, no migration, no new LLM spend,
order-of-magnitude more labelled data. Everything else on the list multiplies whatever this produces, so it goes first.

Five more, briefly

- Human decisions keep the status and throw away the reason. 316 items resolved by a human; approved_by 0, resolution_path 0. Both columns already exist. "Rejected" teaches nothing; "rejected because it named a
  destination that doesn't exist" teaches a lot.
- complete is self-reported. 4,583 complete, 9 ever verified — all one item type, none since 14 July. Training on complete trains on the agent's own say-so, which is exactly the bugs_open/012 failure.
- Free signals unused. attempt_count: 60 items hit 3 attempts and 44 of those are stuck. Hard cases with ground-truth negatives, already recorded, currently invisible to any export.
- feed-triage is the sleeper. It already emits {score, reason, credibility, credibility_reason, source_tier, flagged} and batches many items per call — 423 calls carrying thousands of judgements, running daily since
  March. Best-shaped non-loop source on the platform and it wasn't on anyone's list. The three big auditors return prose instead; porting the loop's cite-or-abstain contract to them would make ~15k calls/month
  training-shaped.
- The log is lossy. LogLLMCall is fire-and-forget with a 5s timeout — rows vanish under load, and load correlates with the interesting runs. That's undetectable corruption after the fact.

The honest ceiling

Even with all of it, this platform makes hundreds of decisions a day, not millions. Six-month realistic scale is tens of thousands of outcome-labelled steps — fine-tuning and eval scale, not pretraining. Worth doing,
because a few thousand outcome-verified domain traces with real negatives is genuinely scarce; but the goal should be stated as a high-quality specialist corpus, not a big one.

Cheapest volume lever once the plumbing's fixed: replay /bugs_open/ — 20 cases with documented root causes and verification steps, so N bugs × M models generates graded trajectories with the answers already written
down.

One correction I should flag: I tried to measure "fix didn't hold" via recurrence and my query was wrong — a naive self-join reported 387,301 recurrences for page_rerender, which is the join exploding, not a finding.
It's in the doc as an open lead with the bad number explicitly not quoted.

Gaps 1 and 2 are platform/ changes, so not mine to make — they'd go to the owning threads through the council gate. Want me to draft those two submissions?

---