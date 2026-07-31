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
