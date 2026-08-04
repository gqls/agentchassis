# Where we are — per-substep error tolerance (bug 173)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04, morning — what this is and why it was worth doing

The job was to pick up the next open bug nobody else was working on and fix it properly —
at the framework level, not just for the one site that tripped over it.

**The bug, in plain terms.** A lot of our work happens in loops: build twelve pages, one
after another. Each pass through the loop does several things — write the content, save the
sections, extract the links, and so on. Today the platform has exactly one switch for what
happens when one of those things fails, and the switch covers the **whole loop**. Either
every failure inside the loop is fatal and the entire build stops, or every failure is
shrugged off and the item is quietly skipped. There is nothing in between.

That is a genuinely awkward position to be in, and it has already forced a bad decision. A
few days ago another thread needed one small step — recording links into a table that is
regenerable and currently empty — to be allowed to fail without killing an entire site
build. With no way to say "just this step", they made the step itself pretend it had
succeeded. Four separate reviewers on our review council rejected that, and they were right:
the honest word for it is that the code was made to lie about its own outcome. The reviewers
said, in effect, *the switch you actually need doesn't exist — go and add it, don't work
around it*. That is this bug.

**Why it matters beyond that one case.** When I went looking, the same shape turns up in
three more loops — the main page builders. In each, one page failing to save could take down
a build that had already produced a dozen good pages. So the missing switch is not an
oddity of one lane; it is a gap four live loops are standing on.

**What I'm doing about it.** Letting an individual step inside a loop declare its own
tolerance, and inheriting the loop's setting when it says nothing. It is a small change in
one file. The important properties are what it *doesn't* do:

- **Nothing changes for anything running today.** I checked the live configuration: of the
  79 steps sitting inside loops across the fleet, exactly **zero** currently use the new
  setting. So on the day this ships, every existing build behaves precisely as it does now.
  The new capability is invisible until somebody deliberately switches it on.
- **The unsafe direction is the one you have to ask for.** "Ignore failures" is the risky
  setting, and you only get it by writing it down, in the place a reviewer of that step will
  see it. Doing nothing keeps you exactly as strict as you are today.
- **It works in both directions.** Just as useful as making one step tolerant inside a strict
  loop is making one step *strict* inside a tolerant loop — several of our dispatch loops
  currently swallow everything, and this lets a step opt back out of that.

**One thing I want to flag as a small trap I found on the way.** It turns out you could
already write this setting on an individual step today. It just did nothing — the code
overwrote it a few lines later without a word. So anyone who tried this would have written
something reasonable, seen no error, and got no effect. I've recorded that as a landmine so
the next person doesn't lose an afternoon to it.

**Honest status.** The code and its tests are being written now. I can commit it, but I
can't make it live: rolling a new build out is a whole-fleet operation the owner runs, not
something one session does unilaterally. Our own rule is that a bug stays open until the fix
is genuinely running in production, not merely committed — so unless another session's roll
happens to carry it out today, this will stay open with the remaining step written down
precisely. I would rather leave it accurately open than tidily closed and wrong.
