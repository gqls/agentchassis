# Where we are — the dispatch and database-pool lane

Plain prose, append-only, newest at the bottom. This is the owner's log. Add to it; don't
rewrite it or reorder it.

> Created 2026-08-14, later than it should have been. This lane ran for four days on handoff
> documents alone. The history before today is in `HANDOFF_2026-08-11/-12/-12b/-13_continue_here.md`
> and in `NOTES_dispatch_fail_closed.md`.

---

## 2026-08-14 — we deleted a database handle that never worked, and four dead features fell out with it

The chassis talks to the database through a connection handle. It turned out to be carrying
**two** of them. One is real: it is opened by the layer underneath and handed in, and everything
that uses it works. The other was opened by the message processor itself, from a setting called
`DATABASE_URL` — and nothing in production sets that. It was set to nothing, on every pod, from
the day it was written.

That would be untidy but harmless, except that four pieces of real functionality had been built
*behind* it, each one written as "if that second handle exists, do this". It never existed, so
none of them had ever run. Not once, on any pod. We had already found and fixed a fifth one last
week — the one where a failed dispatch was supposed to leave a record in the database and never
did — and at the time we wrote down that the two handles were not interchangeable. Nobody then
went and checked the rest of the file. Today's work is that check.

**What the four dead things were, and why none of them is a loss.** This is the part worth being
careful about, because "delete four features" sounds alarming and the reassurance has to be
different for each one.

The first was a check on whether a child workflow had finished, with a log line saying it was
"sending response to parent". It sent nothing. Both paths through that code did exactly the same
thing — returned without doing anything — and the only difference between them was which log
line got written. Deleting it provably changes nothing; the log line was the only reason it
looked important.

The second was a function that built a workflow's final result. It had a genuine bug in it: it
always substituted a placeholder where the real data should have gone. But the function has **no
callers anywhere in the codebase**, so nothing ever received the placeholder. This one is worth
dwelling on, because the tempting fix — point it at the working handle — would have been the
dangerous option. It would have brought a dead code path back to life and changed what every
parent workflow receives, which is a real behavioural change smuggled in under a tidy-up.

The third was a duplicate-message check. This is the one that would matter if we got it wrong,
because switching duplicate detection off would be serious. It isn't off: the layer underneath
does exactly the same check on the working handle, and we can watch it doing it — 449 records
written by 82 different workers in a single hour. I also checked something the earlier
investigation hadn't: the specific rule about handing a claim *back* when a dispatch fails
temporarily also exists in that lower layer, so we lose no part of it. This was a second,
never-run copy of a decision the platform already makes correctly one level down.

The fourth and fifth were smaller and were not in the original bug report at all — a field and a
function, both built to use the dead handle, both never read by anything. I found them by
re-checking the file rather than trusting the write-up, which is the argument for re-checking.

**A judgement call I want on record.** One test had to be deleted, because the thing it tested
was the dead handle itself. I could have quietly let the test count drop; instead the deletion
is stated in the commit message and in the test file, where a future reader will find it. A
second test I redirected rather than deleted, and then I deliberately broke the production code
to confirm the redirected test still notices — it does. A test that passes after you move it is
not evidence that it still checks anything, and that felt worth ten minutes.

**Where this leaves us.** The change is committed and it is in review by the automated council;
I committed without waiting for the verdict, which is the standing practice here, and the
verdict still needs reading and acting on. Nothing is live yet: this is Go code, so it does
nothing until the next fleet roll. That is also why the bug stays on the open list — until it
ships, the defect is still there on every pod. The honest test after the roll is not "does it
compile" but the two paths that *are* alive: failed dispatches must still leave their database
record, and dropped validation errors must still leave theirs. Both of those could come out
either way, which is what makes them worth running.

**One thing I found that isn't ours.** A test in an unrelated part of the codebase (the thunder
adapter) does not compile at the moment — it refers to a field that no longer exists. It is
committed, so someone shipped it that way; it has nothing to do with this work and I have left
it alone. It is worth knowing about because the usual quick build check doesn't compile test
files, so it is invisible unless you go looking, and the next person to run the full test suite
may think they broke it.

---

## 2026-08-14, later — the new build carries it, and the checks came out clean

You rolled a fresh chassis and the deletion is live on both pods. Confirming that turned out to be
the interesting part, because the two obvious ways to check were both the wrong tool.

The normal way to ask "did my change ship?" is to ask the running service which commit built it.
That line is printed once when the service starts, and on a busy service it has scrolled out of
reach within the hour — it was long gone. The second instinct is to search the running binary for
my own commit id, and that would have been actively misleading: the release is built from
whatever was current at build time, which was a *later* commit than mine, so searching for mine
would have come back empty **on a perfectly correct deploy**. I would have gone looking for a
broken deployment that didn't exist.

What worked plays on the fact that this change *removes* things. Two log messages that only
existed in the deleted code are now absent from the running binary — and, importantly, a third
message that I deliberately kept is still there. That last part is what makes it evidence rather
than wishful thinking: if my search method were simply broken it would find nothing at all, so the
message that *must* be present is what proves the method works. I checked first that the deleted
message existed exactly once before my change and nowhere after it, so its absence can't be
credited to anyone else.

Then the behaviour, which matters more than the binary. The strongest check was available because
the old bug had a very specific signature: a failed dispatch used to leave *no database record at
all*. So I sent a deliberately malformed request naming an agent that doesn't exist, and the
record appeared — correctly attributed to the agent that was *asked for* rather than the pod that
happened to receive it, which was the whole point of the earlier fix. That check could only pass
on repaired code. Duplicate-message handling also carried on across the roll untouched.

**One check I could not complete, and I want to be straight about it rather than round it up.**
The third thing worth watching writes a row whenever a particular kind of error is discarded. No
such row has appeared since the roll — but that function only fires about twice a day, and the
last time was six hours *before* the roll went out. So the expected number of rows in the window
was under one, and finding none tells us nothing at all about whether it works. I nearly recorded
that as a failure, and separately I nearly identified the wrong function altogether: I was
filtering on a severity label that four different parts of the system share.

The reason I've still closed the bug is that this particular edit is provable without observing
it. The old code preferred the dead handle and fell back to the live one; since the dead handle
was *always* dead in production, the old code always ended up using exactly the handle the new
code uses directly. Same value, by construction — not "probably fine". And the one assumption
underneath it, that the live handle works on the new build, is precisely what the dispatch test
above measured. I've left the query in the file for whoever wants the direct sighting when it
next occurs.

The bug has moved to the closed pile. That leaves this lane with no outstanding bug work.
