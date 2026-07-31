# Where we are — the deploy path bug (bugs_open/125)

*(plain prose, append-only, newest at the bottom)*

---

**2026-07-31, afternoon.**

The job was "take the next open bug nobody else is working on". I picked one, spent
twenty minutes reading the code around it, started writing the fix — and discovered
another session was already twenty minutes into exactly the same fix. We had both,
independently, written a new shared helper into the same file, in the same minute. I
threw mine away and gave them my notes.

Worth saying plainly, because it is the interesting part: **I did run all the checks the
rules ask for, and they all said the bug was free.** They were not wrong. They just all
answer the question "has anyone finished this?" — and nobody had. There is no signal
anywhere that answers "is anyone in the middle of this right now", except reading what
the other sessions are actually doing minute by minute. So I started doing that, and it
is now written down as the first thing to run before picking up a bug.

The bug I then took is a good one. When the platform rebuilds a single page and commits
it to the site's repository, something has to decide *which file* to write. There is a
function that makes that decision, and it works out the filename from the page's short
name — "about" becomes "about.html". That is fine for a page that lives at the top
level. It is wrong for every page that lives in a folder. A page whose real address is
`/tools/password-entropy.html` gets written to `/password-entropy.html`.

The damaging part is that this does not *move* the page. The real page stays exactly
where it was, untouched, and a second copy of it appears at the wrong address — a real,
live, fetchable page that nothing links to, that the platform has no record of, and
that no clean-up job will ever remove. It happened for real on finetuning.uk on 28 July
and the owner had to delete the file by hand, because we have no way to unpublish
anything.

I re-counted how many pages this would affect. The bug was filed saying 280 of 431. It
is now **316 of 472** — two thirds of every page we have. Three days moved it by 41
pages, which is a decent argument for never quoting a number without re-running it.

The fix itself turned out to be more interesting than "add the missing line". Instead of
looking at the broken function, I went looking for everywhere in the codebase that turns
a page into a filename — and found five of them. Four already do it correctly. The one
that doesn't is the one the page-building pipelines actually use. So the problem is not
a missing feature; it is the same rule written out five times and one copy drifting,
which is a failure mode we have a standing rule about. I wrote one shared version and
moved all five onto it, so it cannot drift again.

Two things the earlier investigation had written down turned out to be wrong, and both
were the kind of wrong that only shows up when you check rather than read.

The first: there is one page whose address is a *section* of another page —
`/tools.html#audience-check`. The earlier note said the fix should just strip the `#...`
part off. I nearly did. But `/tools.html` is a **different page's** address, so stripping
it would make rebuilding one page overwrite another page's file. An address with a `#`
in it doesn't name a file at all, so the right answer is to refuse it and fall back, not
to tidy it into something that looks valid. Making a bad input *valid* and making it
*correct* are not the same operation.

The second: page addresses all start with a `/`, and the code that talks to GitHub
sticks the site's name and another `/` on the front. Pass the address straight through
and you get a double slash and a broken path. Nobody had mentioned it.

I also got two things wrong myself today, both caught quickly and both written up: I
searched for callers of a function while accidentally excluding the file it lives in and
concluded it was dead code (it has two callers), and the bug report names a function
that does not exist in the repo under that name.

Where it stands: the fix is written, all the tests pass, and it has gone to the review
council — that usually takes about half an hour because of the queue. Once it comes back
I will commit it, build a new image and roll it, and then prove on the running pods that
the change is actually in the binary rather than trusting the version tag, which is a
thing that has bitten people here before.

One caveat I want on the record now rather than later: **this stops new wrong copies
being created; it does not remove the ones already out there.** We still have no way to
delete a published page, which is its own open bug.

---

**2026-07-31, later.**

Done. The fix is live and the bug is closed.

The review council sent it back the first time, and it was right to. Five of the twelve
reviewers spotted the same thing: our landmine file names a function `getPageInfo` in the
file I was changing, and my submission said I had changed a function called
`loadPageInfo`. Their worry was the serious version of it — that I had patched the wrong
function and left a sixth copy of the very rule I was claiming to have de-duplicated, in
the one file where that would be most embarrassing.

The truth was more boring and still my fault: **there is no `loadPageInfo` in this
codebase.** I had read the file starting from partway down, needed a function name for
the paperwork, and typed a plausible one without scrolling up to check. The line I
changed is inside `getPageInfo`, so the fix was right; the description of it was not. In
a submission whose entire argument is "I went and counted all the copies", a name I never
checked is not a typo — a reviewer can't tell it apart from a copy I missed.

Two good things came out of being sent back. The landmine entry they were quoting turns
out to be **wrong about its own mechanism** — it says a guard "keys on the page being
named index" when in fact the clause immediately before it already covered the case. I
corrected it and credited the round. And a reviewer asked about a second function I had
skimmed past: it turns out to be the *mirror image* of mine — one builds the address we
store, the other builds the file path we write, and they deliberately disagree about the
leading slash. Anyone who "tidies" those two into one helper will quietly break half the
system, so that's written down now too.

Round two came back approved.

Then I pushed and rolled the image, waited until no other review rounds were running so I
didn't kill anyone else's work, and checked the fix on the actual running servers rather
than trusting the version number — including checking that the old server *didn't* have
it, which is the bit that proves the check can tell the difference at all.

Two things I want to be straight about rather than leave in the small print.

**This stops the problem happening again; it does not clean up what already happened.** We
still have no way to delete a published page. That's logged against the other bug that
needs the same missing piece.

**And nobody knows how big the mess is.** I said two thirds of pages were affected — that
is how many were *exposed* to the bug, not how many actually got a duplicate published.
Those are different numbers and I had been sliding between them, as had the bug report.
When I tried to count the real one, the obvious query came back zero for the whole of the
last eighteen days, which cannot be right, because we know one happened three days ago. So
the query is looking in the wrong place. I have written that down as a failed measurement
rather than reporting the zero, because a confident zero here would be worse than no
number at all.
