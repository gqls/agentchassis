# README — bugfix 247, plain-prose log (append-only, newest at the bottom)

## 2026-08-11

Picked up bug 247 from the open-bugs backlog. It's about some dead code in the message
processor: there's a leftover function called `processRequest` that nothing calls any more
(the real message-handling path moved on and left this one behind), and it uses a small
in-memory cache that has no lock protecting it. Because nothing calls it today, that's not
actually causing a crash right now — but if some future change accidentally wires it back
up, it would cause a real concurrency bug (a "data race"), and in the meantime it's
dangerous in a quieter way: a session investigating "how does this system decide what an
agent should do?" can stumble on this well-named, plausible-looking code and waste time
reasoning about it, when it's actually inert. There's a second, similar leftover
(`selectWorkflowOLD`) that has the exact same shape of bug that was already fixed in the
*live* version of that function — left broken on purpose in the dead copy, which is
arguably worse, because now it looks freshly maintained.

Checked nobody else was already fixing this (nobody was — confirmed via the ownership
script and by checking other sessions' activity), and double-checked the bug's claims were
still true by re-running the greps myself rather than trusting the write-up. They held up,
and I found independent corroboration in an unrelated architecture document that happened
to mention the same dead function as a side note.

Had a plan drawn up (using the lighter "fable" model, as asked) for exactly what to delete
and what to leave alone — the tricky part is that the *type* this dead code lives on
(`AgentConfigLoader`) is used elsewhere for real, so the fix has to be surgical: delete the
specific unused method and its cache, not the whole file. The plan also turned up two more
dead functions in the same area that aren't part of this bug — noting those for a follow-up
rather than pulling them into this change.

Next: make the edit, build and test it, put it through the advisory review process, and
commit.

Update, same day: done. Made the deletion, ran the build and the full test suite (including
a race-detector run, since a race was the whole point of the bug) — all clean. Sent it
through the advisory review process; it came back approved unanimously on the first round,
quickly, because the queue happened to be empty. Committed. The code change itself is small
(a pure deletion, nothing added) and safe, because the only thing that changed is that some
already-inert code is now gone rather than sitting there waiting to confuse the next person
who goes looking in that area, or to become a real bug if something ever accidentally started
calling it again.

One thing still owed, not urgent: this kind of change only takes effect once the service gets
rebuilt into a new image and rolled out, which hasn't happened yet. Until then the bug stays
open on paper even though the fix is written and reviewed — that's a deliberate house rule
here, not an oversight.

## Later the same day

A fresh build of the whole fleet went out. Checked directly against the running service
(not just "a build happened, so it's probably in there") and confirmed the fix really is in
the version that's live now — found the exact commit the image was built from and confirmed
my fix commit is one of its ancestors. So this one is genuinely done: written, reviewed,
built, deployed, and now verified. Marked it closed in the bug file itself rather than
moving the file to the closed-bugs folder — apparently that's a deliberate house preference
here (review it in place before it leaves the open list), which I'd started to write over out
of habit before catching myself.
