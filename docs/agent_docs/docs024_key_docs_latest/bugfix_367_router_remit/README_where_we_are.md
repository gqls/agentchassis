# Where we are — bug 367 (plain prose, append-only, newest at the bottom)

## 2026-08-23, evening — a queue was quietly marking its own failures as successes

There is a piece of the system whose only job is to sort. When something notices that a
chunk of a page is missing required text, it writes a note. This sorter picks the note up,
works out what kind of problem it is, and then does one of three things: hands it to
whoever can fix it, parks it for a person, or closes it as no longer true.

To do any of that it first has to find the chunk of page the note is about. It was looking
only at chunks that had already been published. If it found nothing, it closed the note —
"this can't be found on the live site any more" — and the note went away clean: marked done,
no error, nothing to see.

That was a reasonable way to look, right up until August, when we added a second thing that
writes these notes. The new one exists **specifically** to catch problems on chunks that
never get published — that was its whole point. So every note it wrote was about something
the sorter had been told not to look at, and the sorter closed them all as imaginary.

I checked one for real before changing anything. The sorter said the chunk could not be
found. The chunk was sitting right there, unpublished, with about nine thousand characters
of content in it, and the two missing bits of text really were missing. The note was true
and it had been filed away as resolved.

**The dangerous part is the tidiness.** Before we fixed a related bug last week, these notes
failed noisily — three attempts, then parked where someone would see them. Now they close
cleanly. If you asked "are we acting on our missing-text findings?", the answer would come
back "yes, all of them", and it would be wrong.

## What I nearly did, and why I did something else

The obvious fix is to widen what the sorter looks at. The bug report suggested that, and I
started there. Two things stopped me.

First, widening on its own doesn't actually fix anything. If the sorter finds the chunk and
hands it on for repair, the repair step immediately falls over, because the new note-writer
doesn't fill in four pieces of information the repair step needs. So you'd swap a silent
wrong answer for a loud crash. Better, but not a fix.

Second — and this is the one that mattered — the repair it would have handed off to is not
a gentle edit. It deletes every chunk on the page and rebuilds them. I looked at what has
actually happened to those repair jobs: **28 out of 31 have already failed**, because most
of these pages are ones we've marked as hand-owned and refuse to rebuild. That's someone
else's open bug, and the page in front of me is exactly that kind of page. So sending this
work down that road would have meant piling more failures into another team's dead end.

## What I did instead

I made the sorter honest rather than clever. It may now only close a note when it has
**positive proof** that the thing is gone: the page itself has been deleted, or the chunk
has been explicitly retired, or a person has locked it as "leave this alone". Anything else
— including simply not finding it — now goes to the parked pile for a human, with the facts
written on it: what state the chunk is in, and the three things that would resolve it
(publish it, lock it, or retire it).

We already had this exact rule written down elsewhere in the system, in a neighbouring piece
of code, in almost these words: not finding something might mean it's gone, or it might mean
you looked in the wrong place, and those are not the same, so don't close on it. The sorter
now follows the same rule.

## How confident I am

More than usual, because I tested it without touching the live system. I applied the change
inside a database transaction, asked the changed sorter to classify three real cases, then
threw the transaction away. Only after all three came out right did I apply it properly, and
then I re-ran the same checks against the real thing.

The three cases were chosen so it could have failed:

- the real note that started this → now correctly parked, with the chunk found
- a chunk we genuinely *did* retire → **still closes**, as it should
- a page that genuinely doesn't exist → **still closes**, as it should

That middle one matters most. It would have been easy to "fix" this by making the sorter stop
closing things, which would just break a route that was working. It didn't.

I also re-ran every one of the 65 notes of this kind that has ever been filed, through both
the old and the new sorter, and compared. **Exactly one changed** — the one this bug is
about. Everything else sorts identically.

## What is still not fixed, and I'd rather say so

The notes from the new writer are now **visible and honestly labelled. They are not
repaired.** They sit in the parked pile waiting for a person, because there is no safe
automatic repair for an unpublished chunk today. Making them repairable needs the other
team's bug fixed first. Nobody should read this as "we now fix those".

Separately, while tracing this I found something bigger and left it alone deliberately. The
system has a safety net whose whole job is to stop a job reporting success when it did
nothing. There are three places in the code that can mark work as complete, and **one of
them never consults that safety net** — the one this sorter uses. That is why the silent
close was possible at all. It affects far more than this bug, so changing it inside a bug fix
would be the wrong way to do it. I've written it up as its own thing for someone to take a
proper decision on.
