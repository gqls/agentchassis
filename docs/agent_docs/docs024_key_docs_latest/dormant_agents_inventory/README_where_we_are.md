# Where we are — dormant-agents inventory (bugs_open/044)

Plain-prose running log, newest at the bottom. Append only.

---

**2026-07-21 — first pass, built.**

The bug (044): the platform has no list of its own capabilities, and no way to
notice when one exists but nothing ever uses it. That is how, back in February,
we had a working "section editor" agent that had run three times in production,
and a thread still declared it didn't exist — twice, through a handoff and two
sign-offs. Nothing had failed, so nothing flagged it. It's the exact mirror of
the other open problem where findings pile up in a review queue nobody empties
(033): here it's a worker that no work is ever sent to.

I built the detector half (the bug explicitly splits off a second, "tidy up the
agents wrongly still marked active" half as an owner decision — I've only
*reported* that, not touched it). The detector is a small deterministic sweep
that sits next to the two we already have (triage and silent-check) and works the
same way: it looks at every active agent, works out which ones have never
actually been seen running, writes a plain report, and — once it's switched out
of preview mode — files one quiet, un-actioned marker per dormant agent for a
human to look at. No AI, nothing it can break, and it ships in preview
(dry-run) so the first thing it does is just show you the list.

How it tells whether an agent has ever run is the clever bit, and it's the bug
author's method, not mine: each agent's workflow has some uniquely-named steps,
and we can look for those names in the history of everything that's actually
run. (The obvious shortcut — a "who owns this run" column — is useless here,
because 95,000-odd of our 106,000 runs just say "generic".)

Two things I checked before trusting it:

- The bug's own write-up used the "feature-designer" agent as an example of one
  the method correctly sees as *having* run. Live, my version flagged it as
  *never* run, which stopped me cold — a wrong flag is the whole way this tool
  could embarrass us. Turns out the method is right and the write-up was loose:
  feature-designer's own workflow genuinely never ran; what people remember as
  "the designer worked on the 18th" was actually the review council approving its
  plan, which runs through different plumbing. So I've been careful to word the
  report as "never *seen* running" and to say plainly that it can miss things
  that run via the council path — it's a list to triage, not a verdict.

- Some agents share all their step names with others (the council-gate agent is
  the main one — we deliberately keep its steps identical to another agent's).
  For those the method simply can't tell, so it never guesses — it lists them
  under "can't measure these" and leaves them alone.

Right now: code, the database seed, tests, and these docs are written and
committed. The catch is the usual one — the code doesn't do anything until the
chassis image is rebuilt and rolled out, so 044 stays open until that happens.
After the roll: run it once, read the report, and if it looks right, flip it out
of preview. There are about 70 agents past the "give it two weeks to have run"
cutoff, most of them old retired ones — which is itself the tidy-up the owner
needs to decide on.

**A choice for you (owner), when you get to it:** the report groups the ~70 into
a rough guess at "current / older / paused workstream / legacy", but I can't
reliably tell "retired, should be switched off" from "paused on purpose" from
"real capability we forgot to wire up". That last group is the actual point of
the bug. The report surfaces all of it; deciding which agents to retire vs. wire
up is the half 044 reserved for you.
