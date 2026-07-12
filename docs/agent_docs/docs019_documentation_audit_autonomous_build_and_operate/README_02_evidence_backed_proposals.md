
1. What we have actually built (honest scope)

Strip away the phase labels and it's this: a pipeline that turns "something is wrong" into "here is a reviewable, evidence-backed proposal" without a human in the middle — but with a human always at the end.
Concretely: symptom → cited diagnosis (forbidden from guessing) → constrained edit plan → adversarial review by two specialists → revision informed by real database queries the reviewers themselves asked for →
either an approved plan, or an honest escalation package saying "this needs you."

What it is not yet: a feature-builder, a migration engine, or anything that acts from mission documents. It has been exercised on exactly one bug. The write step (plan → PR) is half-built.

But — and this is the key point for your question — the valuable thing isn't the bug-fixing. It's the pattern, which is general:

- Cite or abstain — no claim without quoted evidence; honest "UNVERIFIABLE" beats confident wrongness
- Hard deterministic gates between every LLM step — CONFIRMED gate, plan validator, file allowlist. The models propose; plain Go code decides what proceeds
- Adversarial multi-agent review with deterministic aggregation — reviewers argue, a rule (not a third model) combines them
- Self-verification instead of self-belief — when a reviewer's objection hinges on a fact, the loop runs the query rather than letting the proposer argue
- Escalation as a first-class success — "this is beyond my mandate" is a correct output, packaged for you
- Everything durable, one correlation id — the entire story is reconstructible after the fact

2. The direction I have been applying

Two explicit commitments, both recorded in the docs as decisions:

The value proposition (decided 2026-07-09): not "the loop finds what humans can't" — on this platform, bugs are legible to anyone with schema access and patience. The proposition is unattended, cited, consistent —
the 3am diagnosis with a paper trail, handing you a package instead of a hunch. Every design choice flows from that.

The single enemy is confident wrongness. Runs 1–2 of the benchmark produced CONFIRMED verdicts that were wrong, stated with full confidence. Everything since — the closure gate, the citation-backed coverage, the
council, the verify step — is aimed at that one failure mode. Your word for it is hallucination-minimisation; the mechanism is: never let one model's assertion pass unchecked into consequence.

And one structural commitment: the human gate never moves. More autonomy upstream (diagnose, plan, revise, even commit-to-branch) — never past the PR. Nothing merges itself.

1. What we have actually built (honest scope)

Strip away the phase labels and it's this: a pipeline that turns "something is wrong" into "here is a reviewable, evidence-backed proposal" without a human in the middle — but with a human always at the end.
Concretely: symptom → cited diagnosis (forbidden from guessing) → constrained edit plan → adversarial review by two specialists → revision informed by real database queries the reviewers themselves asked for →
either an approved plan, or an honest escalation package saying "this needs you."

What it is not yet: a feature-builder, a migration engine, or anything that acts from mission documents. It has been exercised on exactly one bug. The write step (plan → PR) is half-built.

But — and this is the key point for your question — the valuable thing isn't the bug-fixing. It's the pattern, which is general:

- Cite or abstain — no claim without quoted evidence; honest "UNVERIFIABLE" beats confident wrongness
- Hard deterministic gates between every LLM step — CONFIRMED gate, plan validator, file allowlist. The models propose; plain Go code decides what proceeds
- Adversarial multi-agent review with deterministic aggregation — reviewers argue, a rule (not a third model) combines them
- Self-verification instead of self-belief — when a reviewer's objection hinges on a fact, the loop runs the query rather than letting the proposer argue
- Escalation as a first-class success — "this is beyond my mandate" is a correct output, packaged for you
- Everything durable, one correlation id — the entire story is reconstructible after the fact

2. The direction I have been applying

Two explicit commitments, both recorded in the docs as decisions:

The value proposition (decided 2026-07-09): not "the loop finds what humans can't" — on this platform, bugs are legible to anyone with schema access and patience. The proposition is unattended, cited, consistent —
the 3am diagnosis with a paper trail, handing you a package instead of a hunch. Every design choice flows from that.

The single enemy is confident wrongness. Runs 1–2 of the benchmark produced CONFIRMED verdicts that were wrong, stated with full confidence. Everything since — the closure gate, the citation-backed coverage, the
council, the verify step — is aimed at that one failure mode. Your word for it is hallucination-minimisation; the mechanism is: never let one model's assertion pass unchecked into consequence.

And one structural commitment: the human gate never moves. More autonomy upstream (diagnose, plan, revise, even commit-to-branch) — never past the PR. Nothing merges itself.

3. Your aspirations, mapped to what exists

┌────────────────────────────────────┬────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│              You want              │                                                                              Where it stands                                                                               │
├────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│                                    │ This is the council pattern — and the roster of 2 (edit-quality, guardian) is explicitly a skeleton. The runbook's F2 design already names the full bench: guidelines      │
│ Many agents on every task, each a  │ agent (does this conform to 000-0xx — or did the guidelines fall short?), reuse agent (are we rebuilding something that exists — checking code and docs), bug-historian    │
│ different lens                     │ (has this class recurred?), compliance eye, pipeline guardians one per master workflow, specialist knowledge agents. Adding a reviewer is a seed change + prompt + curated │
│                                    │  context. The aggregation rule already scales                                                                                                                              │
├────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Patches consistent with structural │ Partly proven: the guardian vetoed an architecture change dressed as a fix — that's exactly this instinct, working. What's missing is the memory: today the guardian       │
│  decisions, changing only when     │ reasons from the diagnosis; it should reason from your decision record (the constitution, the DECISIONS sections, guideline docs) as curated reviewer context. That's open │
│ completely right                   │  question Q-G, and it's the single highest-value extension for your goal                                                                                                   │
├────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Legacy-migration agents            │ Not built, but it's the same machinery with a different intake: instead of a bug symptom, the input is "pattern X supersedes pattern Y" — a scanner finds Y-shaped code,   │
│                                    │ the proposer writes constrained migration plans, the council reviews, PRs flow. The allowlist/gate/council scaffolding transfers unchanged                                 │
├────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Features from specs/mission docs   │ Honestly furthest away. The diagnosis loop's guards are built around causal grounding in evidence ("cite the mechanism"). Feature work needs different grounding — "cite   │
│                                    │ the spec clause / mission line this serves" — which is the same shape but a new evidence tier. Plausible; not designed                                                     │
└────────────────────────────────────┴────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

4. . Your awareness concern — I think it's the right concern

You've named the real risk: not that the framework does something wrong, but that it does things you don't know about, and drift compounds silently. Two honest observations:

- Everything already leaves a trail (artifacts, PR bodies, escalation packages) — but a trail you must pull is not awareness. You'd need to go looking.
- The missing organ is a push surface: a periodic digest — what ran, what was decided and by which rule, what was escalated, what changed, what the council almost approved. F0.3 (per-iteration notes) was parked
  long ago; this is its grown-up form.

My concrete recommendation, and I'd treat it as a gate: before autonomy widens — before councils multiply or migration agents exist — the awareness surface gets built first. The system's own principle applied to
itself: it must explain what it's doing, or it doesn't get to do more.

5. The fork

I think it's the right move, for a reason beyond comfort: the strongest isolation is that the write surface points only at the fork. The git-adapter's credential and repo config would target the fork exclusively —
this repo becomes physically unwritable by the loop, not just protected by review discipline. The loop then develops the fork; you pull from it into here only what you've reviewed and want. Your awareness surface
and the PR list are the membrane between the two.

What it needs (a designed slice, not huge): fork the repo; point the loop's intake defaults (owner/repo/ref are already parameters), corpus indexing, and adapter write config at it; seed the agents. I'd fold
"mission and objectives correct in the first place" into the same slice — the fork's constitution/mission docs become the councils' curated context, so conformance is checked against documents you authored, not the
loop's own drift.

6. The escalate decision, explained

The write step activates only on decision == 'approved' — that gate is load-bearing and I won't soften it. But our only test bug never produces an approved plan, because the council keeps ruling — correctly, three
runs running — that its proper fix is architecture-level, beyond a constrained plan's mandate. The loop isn't failing; the bug is simply the wrong size for it. So to test the write step end-to-end you must choose
how to obtain an approvable plan:

1. Seed a small bug (my recommendation) — plant a deliberate, contained, single-file defect on a test surface (or pick a genuinely trivial real one). The council can genuinely approve its fix, and the first PR the
   system ever opens will have earned its approval honestly — which is also the demo worth having.
2. Hand-approve an existing plan — fastest, but the plans on record were judged incomplete/risky, so you'd be pushing a known-flawed plan through the write path; it tests the plumbing while contradicting the
   reviewers we built.
3. Wait for an organic small bug — purest, unbounded timing.

  ---
That's the shape of it. For the second turn you offered, the threads I'd most value your reaction on: (a) the build gate (A/B/C from the summary doc), (b) seeded-bug vs hand-approved for the first write-step run,
(c) whether "awareness surface before wider autonomy" becomes a standing rule, and (d) when you want the fork slice designed — before or after the write step completes. Push back anywhere you think I've got the
direction wrong.
