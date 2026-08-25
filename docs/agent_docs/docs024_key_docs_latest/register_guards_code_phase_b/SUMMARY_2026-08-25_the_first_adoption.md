# SUMMARY 2026-08-25 — the first adoption: something is actually being watched now

## What we are trying to do

Every one of our calculator pages has a small store of facts sitting beside it — the tax
thresholds, the rates, the legal figures, each one written down with a source and checked
against that source every day. The calculators themselves have the same figures buried in
their code. **Nothing connected the two.** You could correct a figure in the register,
watch the check go green, and the calculator beside it would carry on using last year's
number for as long as nobody happened to look.

That is not hypothetical. It is what `bugs_closed/225` was: a first-time-buyer cap that
expired sixteen months earlier, still being used to quote people a tax bill, with the
correct figure written down two feet away and every check we owned passing the page.

So the job is to make the register reach the code — to have a calculator say which
registered facts it uses, and to have something notice, the day one of those facts moves,
that this calculator needs looking at.

## Where we have come from

The machinery was built over the last ten days and went live a day ago. It ran across the
whole estate properly for the first time yesterday morning: nineteen sites, no errors. It
worked, and it immediately found a mistake of mine, which is roughly what you want a new
check to do.

But it was watching almost nothing, because **watching is opt-in.** A calculator has to
declare which facts it uses, and on an estate of a hundred and seventy-eight calculator
pages, exactly one had ever declared. The check was in place and pointed at almost nothing,
and there was no mechanism to change that except asking people.

So the last piece built was a suggester: rather than asking someone to hand-write a hundred
and seventy-eight declarations, the daily sweep looks for calculators whose code already
contains figures from their own register, and writes them a ready-to-paste suggestion. On
its first run it produced five, across three sites.

## What we have done

**Today the first of those suggestions was taken up, and it is our second stamp duty
calculator — the one in exactly the position the original bug describes.** It now declares
the register facts it uses, and the daily sweep will name it the day one of them moves.

Three things about how that went are worth saying out loud.

**The machine asked for seven and the right answer was thirteen.** The suggester finds a
figure by looking for it, literally, in the calculator's code. Our register writes tax
rates as percentages — 2, 5, 10, 12 — and the calculator writes them as decimals — 0.02,
0.05, 0.10, 0.12. Those can never match, and a two-digit number is too common to search for
safely in any case. So six facts were invisible to it, and they happened to be *every rate
in a stamp duty calculator*, which is precisely what a Budget changes. Taking the machine's
list at face value would have left them unwatched behind a declaration that looked
complete — the original bug again, inside the thing built to prevent it. Reading the
calculator's code took ten minutes and found all thirteen.

**The obvious way to install it would have deleted itself.** Our own note told that team to
install through their lane's own installer rather than editing the database row by hand.
That advice was right, and it had one failure mode nobody had checked: their installer
rebuilds the whole file from scratch every time it runs, and knew nothing about
declarations. The paste would have vanished the next time anyone used it — no error, no
warning, everything still looking correct. It is fixed properly now, in the installer, so
the declaration is rebuilt every time instead of surviving until the next run.

**And I had warned them about the wrong file.** The note I sent them described a *different
lane's* installer that happens to share the filename. One command on a path I had already
typed into my own note would have refuted the paragraph. It is corrected, and logged.

## Where we are now

The mechanism is live, it has run for real, and for the first time it is watching something
that matters. Two small fixes from yesterday morning are committed but not yet in the
running system, so a build is owed — the tag is bumped and the command is ready.

**One number should not be over-read.** Of the thirteen facts declared, the check could
confirm seven are present in the calculator's code and declined to guess about six, and it
found nothing missing. That reads like a clean bill of health, and it is not one: this
declaration was written *from the code*, so of course the code contains it. The measurement
we actually need — how often a declared figure turns out to be **absent** from the
calculator — cannot come from a sample built this way. That is stated plainly in the
handoff so the next person does not inherit it as good news.

**And one thing was deliberately not done.** Three calculators on the games site have the
same suggestion waiting. Their plans are written by an automated tool-builder, not a
person, and that builder rewrites the whole plan when it runs — one of them was rewritten
three weeks ago. Anything added by hand would quietly disappear. That needs a fix inside
the builder, which is real code and a review, not something to slip in. Nobody currently
owns that site either.

## Where we are going

Three things, in order.

**The build**, so yesterday's two fixes are actually running.

**More adoptions, of a particular kind.** We now need a declaration written from the
register rather than from the code, because that is the only kind that can tell us how
often a calculator has quietly drifted away from its own facts. Until we have that number
we cannot honestly turn this check from something that reports into something that acts —
and turning it on without the number is the mistake an earlier version of this plan was
rejected for.

**Then the builder.** As long as automated tool-generation silently drops these
declarations, adoption only sticks on the calculators a person maintains by hand. That is
the difference between a mechanism that covers a handful of pages and one that covers the
estate.

The honest summary of the day: the thing we built is no longer only working, it is being
used — and both of the real obstacles we hit were the same shape, which is something
generating a file and quietly discarding the part a human added.
