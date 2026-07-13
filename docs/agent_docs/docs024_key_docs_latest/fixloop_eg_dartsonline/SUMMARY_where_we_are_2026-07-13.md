# Where we are — the self-fixing loop, in plain language

*2026-07-13. Written to be read calmly, start to finish, with no code needed.
For your use. Companion docs, if you ever want the detail:
`RUNBOOK_diagnosis_fix_loop(10).md` (how it works + every gotcha),
`PLAN_fixloop_pilot.md` (what's built and what's next),
`NOTES_running_fixloop(10).md` (the turn-by-turn story),
`HANDOFF_CURRENT_fixloop.md` (the living state, updated every session).*

---

## The short version

We set out to build something that could take a plain-English report that
"something is wrong" with the platform, work out the real cause on its own,
propose a careful fix, have that fix argued over by reviewer agents, and — only
if it survives all of that — turn it into a pull request for you to approve.

**Today it did the whole thing, end to end, for the first time.** It found a
real (small) defect in our own code, wrote the fix, put it through every check,
and opened pull request #1. You read it, approved it, and it merged. Nothing
happened without your yes.

That's the milestone: not that it fixed a hard bug, but that the *entire
assembly line* now works, in production, with every safety gate switched on and
a human holding the final say.

## What it actually does, told as the journey of one bug

Think of it as a small, careful organisation that a bug passes through.

1. **Someone reports a symptom.** In plain words — "the logs say the wrong
   thing here" — with no need to point at the code.

2. **A diagnostician works out the cause.** It reads the real code and the live
   database, forms a theory, and is *forbidden from guessing*: every claim has
   to quote its evidence, or it must honestly say "I can't prove this." When it
   is sure, it writes a conclusion that explains the whole symptom.

3. **A planner proposes a fix.** Not code yet — a short, constrained list of
   exactly which files to change and why, each point tied back to the evidence.
   It is pushed to keep the change as small as it can be.

4. **A council reviews the plan.** Two specialists argue with it: one on whether
   the edits are the *right* edits, one guarding the rest of the platform from
   side-effects. A plain rule — not a third opinion — combines their verdicts
   into *approve*, *revise*, or *reject*. If they want a fact checked ("does
   this column really exist?"), the system runs that query and feeds the answer
   back, so arguments are settled with evidence instead of more opinion. If the
   plan can't be made safe, it stops and hands you a tidy package explaining
   why — that "I need a human" is treated as a proper, successful outcome, not
   a failure.

5. **An implementer writes the code — carefully caged.** Only once the council
   approves. It runs in its own throwaway workspace, rewrites only the files the
   plan named (a strict list it cannot step outside), and makes its changes
   through a single, separate service that holds the keys — so the part of the
   system driven by the AI never holds the credentials to write to your repo.

6. **A build gate checks it compiles.** Before any pull request exists, the
   proposed change is built in a clean container. If it doesn't compile, there
   is simply no pull request — just a note explaining what failed. Broken code
   never reaches your review queue dressed up as ready.

7. **You decide.** A pull request opens, its description carrying the whole
   story — the diagnosis, the plan, the council's verdict. It waits. You merge
   it, or you don't. **The platform never merges its own work.**

## What today proved, concretely

The bug we used was deliberately real and deliberately tiny: a stray debug line
in the image code that printed the *wrong function name* — genuinely misleading,
completely safe to fix. We hand-wrote its diagnosis (the diagnosis stage was
already proven separately; today was about everything *after* it), then let the
rest run for real.

- The council **approved** a plan for the first time — earlier, on a much harder
  bug, it had rightly kept refusing, which is exactly what a good reviewer does.
- The implementer produced a change that was **precisely the plan and nothing
  else**: one file, two lines removed, no incidental edits, no reformatting.
- The build gate went **green**, and pull request #1 opened with the full story
  attached.
- You approved; it merged; the misleading line is gone.

There was even a small bonus that shows the gates are real: on an earlier
attempt the build gate went *red* — and the thing it caught was a genuine,
pre-existing bug elsewhere in the codebase that stopped it compiling. The system
had, in effect, found a second real problem while trying to fix the first, and
refused to open a pull request until it was sorted. That is the safety working,
not failing.

## Why you stay in control (the part worth trusting)

Every design choice bends toward one worry: an AI that sounds equally confident
whether it's right or wrong. So the whole thing is built to be *honest about
what it knows* rather than clever:

- It must cite evidence or abstain.
- Small, deterministic checks — written in ordinary code, not AI judgement —
  stand between every AI step and anything that matters: the "is this really
  confirmed?" gate, the "is this plan valid?" check, the "are these exactly the
  approved files?" list.
- The credentials that can write to your repository live in one isolated place,
  never in the AI-driven parts.
- Nothing merges itself.
- When it isn't sure, it escalates to you with its homework attached.

## Honest caveats

- This has run cleanly on **small, contained** bugs. On the hard training bug
  (a genuinely platform-wide problem) the council keeps — correctly — saying
  "this is too big for a constrained fix; a human should decide the
  architecture." That's the right answer, but it means big changes still land
  on your desk, by design.
- A couple of details are currently set up for our working branch rather than
  the long-term default — noted in the handoff, easily tidied when the moment
  comes.
- The system is only as good as the mission and guidelines it's checking
  against. Keeping those correct is the human job that doesn't go away.

## Where we are, and the next step

**Built and proven, end to end:** report → diagnosis → plan → council (with
self-checking and honest escalation) → caged implementation → build gate →
pull request → your approval.

**The next thing we agreed to build — before making the system any more
autonomous — is an "awareness surface":** a plain, regular digest of what the
loop has been doing and deciding, so you can stay comfortably informed without
going digging. We deliberately put this *ahead* of widening the reviewer council
or adding the legacy-migration agents, because the right order is: make sure you
can always see what it's doing, *then* let it do more.

After that: more reviewer perspectives (a guidelines checker, a "have we already
built this?" checker, a bug historian), and pointing the whole pipeline at a
real bug you choose.
