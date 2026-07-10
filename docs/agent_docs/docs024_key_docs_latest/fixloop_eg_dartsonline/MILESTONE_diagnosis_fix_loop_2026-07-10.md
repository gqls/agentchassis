# Milestone — a self-diagnosing, self-proposing, self-reviewing fix loop

*2026-07-10. A plain-language account of what we built, how we got here, and
what's next. Written to be read aloud or lifted into a summary — no code
required to follow it.*

---

## The one-sentence version

We built a system that takes a plain-English bug report about one of our
websites, works out the true cause from the code and the live database on its
own, writes a concrete fix plan, and has a panel of specialist reviewers accept,
reject, or send that plan back for revision — all automatically, all leaving a
paper trail a human can audit.

## What it is

Four capabilities, chained, each producing a durable record:

1. **Intake.** A bug goes in through one documented command as a plain
   description of the symptom — no need to point at the code.
2. **Diagnosis.** A read-only loop forms a hypothesis, gathers evidence from
   three sources (the actual code, live database rows, and runtime logs), and
   either confirms a cause *with citations* or honestly says it couldn't. It is
   forbidden from guessing: every claim must quote its evidence, and a
   confirmation now has to explain **every** part of the original symptom, not
   just the convenient part.
3. **A fix plan.** Once — and only once — a diagnosis is genuinely confirmed, a
   proposer turns it into a small, constrained edit plan: which files, which
   functions, what change, and the evidence each edit rests on. It writes no
   code; it writes a reviewable proposal.
4. **A council.** Two specialist reviewers judge the plan — one on edit quality
   (are these the right changes, do they miss anything the diagnosis found?),
   one as a "pipeline guardian" watching for collateral damage to the rest of
   the platform. A deterministic rule aggregates their verdicts into *approved*,
   *revise*, or *rejected*. On "revise", the objections feed back and the plan
   is redrafted — up to twice — so it converges before anyone opens it.

Every step writes its output to one durable store, all joined by a single
correlation ID, so the whole story — symptom, evidence bundles, the confirmed
diagnosis, each plan, the review — is fetchable after the fact.

## How we got here — the story worth telling

We didn't build this by designing it on paper and shipping it. We built it by
**pointing it at one real bug and running it, over and over, fixing whatever
broke each time.**

The bug: our darts retail site published a "Guides" navigation link to a page
that was blank, with no guide content behind it — while a sister site on the
same platform had working guides. A link to something that was never built.

Before writing any tooling we diagnosed that bug *by hand*, completely, so we'd
have an answer key. Then we ran the loop against it five times — same symptom
every time, changing one thing between runs — and graded each result against the
key. The runs read like a controlled experiment:

- **Run 1** confirmed a cause — *the wrong one*, confidently. (The real
  mechanism lived in a database-stored workflow the loop literally couldn't
  see.)
- **Run 2** got half of it right and waved the other half away as "not a nav
  issue."
- **Run 3**, once we'd added a rule forcing it to explain the whole symptom,
  refused to over-claim — it stopped honestly and said what it couldn't yet
  prove.
- **Run 4** produced its first full, well-cited answer.
- **Run 5**, under the strictest checks, confirmed the cause *and* — for the
  first time — pulled the navigation-generation code into evidence, a corner it
  had been blind to for four runs because of a budgeting bug in our own tooling.

Each run surfaced a genuine defect in the loop, and each defect got a targeted
fix: durable evidence bundles you can re-read; keeping the original question in
front of the reasoner so it can't drift off-topic; pulling database-stored
workflow logic into the evidence; and three "guards" that stop the loop
declaring victory unless its answer is grounded on more than one kind of
evidence *and* accounts for the full symptom.

Then we turned on the fixer. Its **first plan missed the real fix and included
edits that changed nothing** — which told us, empirically, that the reviewer
council wasn't a nice-to-have but the missing organ. So we built it. On its
**first live run** the council did exactly what a good reviewer does: it did not
rubber-stamp. The second plan was much better — it now targeted the true broken
mechanism — but the reviewers still sent it back with specific, correct
objections (one edit touched a file every site shares; another raised a real
"could this cause an infinite retry?" safety question). The decision came back
*revise*, with reasons. That is the moment the system stopped being one
overconfident voice and became a small organization with checks and balances.

## Why it matters

The failure mode of an AI that debugs is not that it can't find bugs — it's that
it states wrong or partial answers with the same confidence as right ones. Every
piece of this milestone is aimed at that single problem: cite or abstain; explain
the *whole* symptom; ground every claim; and put a plan in front of adversarial
reviewers before anyone acts on it. The result isn't a system that's always
right — it's a system that is **honest about what it knows**, and that argues
with itself before it proposes changing anything.

There's a second, quieter payoff: the whole thing runs unattended and leaves an
auditable trail. The value was never "find a bug a human couldn't" — on this
platform bugs tend to be findable by anyone with database access and patience.
The value is doing it at 3am, with citations, consistently, and handing a human
a reviewable package rather than a hunch.

## What's next

- **The write step.** The council can now approve a plan; the last slice is
  turning an approved plan into an actual code branch and pull request — behind
  an isolated write credential, gated on the code compiling, with a human as the
  final approver. Nothing merges itself.
- **A learning record.** Categorize the bugs the loop confirms so recurring
  classes get caught earlier.
- **Retire the test bug.** The darts-site defect that trained all of this is
  still live (that's what made it a repeatable benchmark). The real fix has been
  known since day one and can be applied by hand whenever we choose — which also
  ends this particular experiment.

## The honest caveats

- Everything above ran on **one** bug. It's a deliberately strong test — a known
  answer, a sister-site control, three kinds of evidence — but it is one case.
- The loop still has a blind spot on that very bug: it reached the neighborhood
  of one sub-cause without quoting it exactly.
- The fixer's plans are only as good as what the diagnosis hands them, and the
  reviewers catch problems rather than guaranteeing correctness. The human at the
  end is load-bearing, by design.

*Full technical record: `NOTES_running_fixloop(10).md` (turn by turn),
`RUNBOOK_diagnosis_fix_loop(10).md` (how it works + every gotcha),
`PLAN_fixloop_pilot.md` (what's built and what's next).*
