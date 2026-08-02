# SUMMARY — 2 August 2026: the camera works, and nobody looks yet

*Successor to `SUMMARY_2026-07-31_a_third_tool_and_the_gap_that_only_eyes_close.md`.
That one named a gap. This one reports that half of it has closed, and is precise
about which half.*

---

## What we're trying to do

Build consultancy-grade interactive components and a site that demonstrates them, and
— the part that keeps turning out to be the harder problem — make the platform capable
of noticing when what it has built is wrong.

## Where we've come from

Every visual fault on this site that mattered was found by a human opening a page. Text
jammed flat against the edge of a card. A row of links sitting twenty-six pixels off
their shared baseline. Overlapping value labels on the council simulator's comparison
band. In each case our automated checks ran, and in each case they all passed, because
no check asserted anything about spacing or alignment.

That produced a specific and slightly absurd situation. The platform *could* take a
screenshot, and did — but only when a check failed. On a page where nothing failed,
nothing was photographed. The one class of defect that only an eye can catch was the
exact class for which no picture was ever kept. The last summary called this "the gap
that only eyes close" and treated it as the standing limitation.

## What we've done

Made the checker photograph a page that **passes**.

Mechanically that is one opt-in flag, and it took three rounds to land honestly. The
first round built the camera in the browser service. The second found that the camera
had no shutter release: the part of the system that requests a check had no way to ask
for the flag, and would have thrown away any photograph that came back. That second
finding is the one worth remembering, because "nobody has switched it on" and "nothing
*can* switch it on" look identical from the configuration table, and the obvious remedy
for the first — write the setting — would have silently done nothing and looked exactly
like success. Reading the calling code before writing the configuration is what
separated them.

The third round, today, turned it on. The order mattered: configuration takes effect
instantly and compiled code does not, so setting the flag before the new build shipped
would have produced a setting that read "on" while the running program quietly ignored
it. The new build went out this evening; we confirmed it was genuinely carrying the
change — by searching the running program for a phrase the change **deleted**, not just
one it added, since only the deletion is impossible to explain away as an old build —
and only then wrote the setting, as a filed, numbered, self-checking change rather than
a bare database edit.

Then we proved it end to end. A real acceptance run on the review-council simulator
passed all twenty-two of its checks and filed two full-page screenshots, desktop and
mobile, of the page exactly as it passed.

## Where we are now

The capability is live and demonstrated, and the proof carries its own control: the
same tool passed the same checks on 31 July, and that record has no photograph attached
to it. Same page, same checks, two days apart, one difference. We did not have to
construct that comparison — the earlier result was already there being the control.

Two limits are worth stating plainly rather than leaving to be discovered.

**We have not opened the image files.** They are in private storage. Asked for one over
the web, we got "not authorised" — but we also asked for a filename we invented, which
certainly does not exist, and got precisely the same answer. That test cannot tell the
two cases apart, so it proves nothing, and we are not going to describe it as if it did.
What we do have is a code-level guarantee: the upload function returns an error and no
file reference at all if the upload fails, so a reference existing at all means the
upload succeeded. That is a sound argument. It is still an argument rather than a
photograph anyone has looked at.

**Nobody looks at them.** This is the bigger of the two. The photographs land as storage
references inside a technical note. There is no page, no digest, no email — nothing that
puts an image in front of a person. A photograph nobody opens is worth exactly as much
as no photograph, so on the original problem — faults that only an eye catches — we have
closed the half that requires machinery and none of the half that requires a pair of
eyes and somewhere to put them.

## Where we're going

The next step is small and it is the one that decides whether any of this mattered:
somewhere for the pictures to go. That could be a line in a weekly digest, a contact
sheet of the last N passing runs, or a review page — the choice is editorial, not
technical, and it belongs to the owner.

Behind it, unchanged and still open: a guide page for the review-council simulator; two
dead call-to-action blocks on the cost calculator; and two stale page records serving
404s that nothing links to. And one open design question on the camera itself — a
full-page screenshot at a single width is not what someone reviewing the mobile
experience actually needs, so renders may want to carry their viewport.

The pattern this lane keeps rediscovering is worth naming once more, because it has now
appeared three times in a fortnight: **the platform's detection consistently outruns its
consumption.** We find things nothing reads, photograph things nobody sees, and file
findings that sit at status "detected". The build is rarely the bottleneck. The last few
feet to a human are.
