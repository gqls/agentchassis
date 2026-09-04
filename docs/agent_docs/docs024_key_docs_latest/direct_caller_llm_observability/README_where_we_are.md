# Where we are — direct-caller LLM observability

*Plain-prose log, append-only, newest at the bottom. The owner maintains this too.*

---

**2026-09-04 — the lane exists, and nothing has been done in it yet.**

This came out of the token-budget work (bug 257). While fixing where a step's size limit is read
from, it became obvious that we can only *see* what the model calls are doing for one part of the
system.

Here is the shape of it. When the platform asks a language model for something, one function writes
a line into a table — which agent asked, how much room it allowed for the answer, how long the
answer was, and whether the answer got cut off in the middle. Every question we can ask about "are
we cutting replies off?" is a question about that table. Two automatic checks run every six hours
and both of them are reading that table.

Calls made from outside the main orchestration layer do not write that line. Not a bad line — no
line. Which means they do not show up as healthy and they do not show up as broken; they are simply
missing, and a missing row looks exactly like a step that never ran.

Today that is: the two Gauntlet endpoints on the tools site, the gripper endpoint, the older
content-creator agent, and the reasoning agent. Six places. The tools site does log something, but
to its own console output, where no fleet-wide query will ever find it.

Nothing is on fire. Nobody has reported a problem caused by this. What it costs is that when we ask
"is anything being truncated?", the honest answer is "not in the part we can see", and we have never
been able to say how big the unseen part is.

The owner asked for this to be its own piece of work rather than staying inside bug 257, where it
had sat unstarted for three weeks. The bug number is **480**.

One warning, written down now because it is the trap this lane will walk into. Adding more logging
is not automatically more truth. In August, one step had a size limit of 2000 written in its
configuration and the number 2000 also hardcoded in the code — so the log said "2000" whether the
configuration was working or being thrown away, and no amount of querying could tell the two apart.
Whatever gets built here has to be able to tell the difference, or it will produce confident wrong
answers faster than the current silence does.
