# SUMMARY — 2026-08-19: the lane stopped fixing tools and started fixing the thing that builds them

*Written for the owner to read aloud. Third summary. The first (08-16) marked the pilot working; the
second (08-18) marked five tools rebuilt and the discovery that they were broken before we touched
them. This one marks a change of kind rather than a change of count: the fixes now live in the
framework, not in our instructions.*

## What we are trying to do

webdesign.co.uk carries 63 tools that were imported rather than built — dropped onto the site as
finished lumps of HTML sharing one wrapper. The framework cannot see inside them, so it has never been
able to check, improve or rebuild them the way it does everything it made itself. The goal, which the
owner confirmed this week, is that **all 63 become framework-owned**, at the same web addresses.

## Where we have come from

The pilot proved it was possible, after we found and fixed the reason it never had been. Then five
rebuilds turned up something we were not looking for: **the imported tools were already broken.** A
minifier whose "remove comments" checkbox did nothing because the code implementing it had been
commented out. The same fault in the SVG cleaner. A log cleaner that silently did nothing when you
typed a bad number, and destroyed your output if your JSON had a typo. None of it visible from the
page.

Then the owner found one of ours. A copy button that said "Copied" when nothing had been copied — on a
tool we had built and marked as passing hours earlier. That was the turn. It was not a mistake in that
tool; it was a mistake in **where our fixes were living.**

## What we have done

**Eleven of the 63 tools are now rebuilt, live, and checked at the actual served page.** Each old
version is retired rather than deleted, fingerprinted before and after, so any of them can be put back
in one step.

The more important thing is the second half of the owner's instruction: *fixes should extend into the
framework so the problems don't recur.* We had been fixing each defect by writing a requirement into
that one tool's instructions. That fixes one tool. **The next tool the system builds — for any site,
by any team — is born with the same fault**, because the requirement was in our prose and not in the
platform.

So six of those requirements are now **rules in the tool generator's own contract**: never claim
success you have not verified; wire up buttons properly; no blocking pop-ups; check what the user
typed before using it; never let an error destroy their work; and always show how much you changed, so
a tool that legitimately does very little cannot be mistaken for one that is broken. Every rule was
written from a fault we had measured in a live tool, not invented.

**It works, and we can show it works.** The two tools built since carry all six behaviours, and
neither instruction sheet mentioned any of them. Those two briefs are the shortest we have written and
produced the most complete tools. That is the whole argument, demonstrated rather than asserted.

## Where we are now

Eleven replaced and serving. Fifty-two to go, of which two cannot be reached at all until a small
platform change is made.

Along the way this work has uncovered **three faults in the platform itself**, all of which only appear
when you try to *fix* something rather than build something new — which is the worst place for a system
to be unreliable:

- A rule that means **a tool can never be rebuilt** once it exists, unless you know to rename the old
  one first. This is what parks the two tools.
- A safety check that **rejects any tool that handles web markup**, telling you the work was cut short
  when it was not. Working round it forces the generator to write deliberately obscure code, which
  makes the check worse at its actual job.
- A "published" timestamp that **is stamped whether or not anything was published.** One page served a
  six-hour-old version while every internal signal said it was live.

All three are written up with evidence and suggested fixes. None is ours to fix.

And a pattern worth naming: **three of the twelve tools we have read tell you, in their own output,
that they did something they did not.** One promised exact SQL and produced prose. One labelled a
fixed grey as "auto-generated". One had a colour picker wired to nothing. These are the dangerous
ones, because the output looks like the output of a tool that worked.

## Where we are going

The path is written down in four phases: the simple tools first, then the thirteen whose code lives in
separate files, then the larger hand-built apps last and one at a time so they get proper attention,
and the two blocked ones whenever the platform change lands. If that change is never made, the honest
finishing line is "61 of 63" and we should say so rather than call it done.

The piece with the most value left is not the remaining rebuilds. It is a **checker**: the new rules
stop the next tool being born broken, but nothing yet finds the faults in the fifty-two already out
there. Building that would turn a job someone has to remember to do into something the system notices
by itself — which is what the owner asked for, and the half we have not built yet.
