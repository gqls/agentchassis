# Where we are — bug 145, the symbol reader that would read anything

Plain prose, append-only, newest at the bottom.

---

## 2026-07-31 — what this bug is, in ordinary words

The diagnosis loop works by showing an AI a bundle of code and asking it what is broken.
The AI answers, and part of its answer is *"next, show me these bits of code"*. Whatever
it names, we go and fetch, and paste into the next bundle under a heading that says
**"In-scope code"**, wrapped in the markup that means *this is Go source*.

The function that does the fetching had no opinion about what it was being asked for. If
the AI named a file rather than a function, the function handed back **the whole file**,
straight off the disk, without ever checking whether that file was source code at all.
So a request for `bugs_open/145.md`, or for a file shaped like a secrets file, would be
honoured — and the result would arrive labelled as our own source code.

The bug was found by our own council, not by a human and not by a failure. One of the
review seats was looking at a different change and asked the useful question: *you fixed
your own instance of this, but what about the next caller?* That is the seat doing exactly
what it is for, and it is worth saying so plainly.

## Two things the original report got wrong, and both matter

**First, it said the door was shut by luck.** The report worked out that the one producer
it looked at could never send a bare filename, and concluded the problem was theoretical.
It had followed the wrong producer. The main source of these requests is **the AI's own
reply**, and nothing between the AI and the file-reader checks anything — one function
trims whitespace and removes duplicates, and that is the whole of it. Worse, the bundle we
send the AI **literally invites this**: when a file has more functions than fit, we print
*"put the bare file path in next_scope to see it whole"*. We were asking for the thing
that wasn't safe to ask for.

That mattered practically, because the report's top-ranked fix was *delete the whole-file
feature*, and deleting it would have broken something we advertise. I was heading that way
before I read the line that invites it.

**Second, it said the leak was limited to files in the repository.** It isn't. I put a
file *outside* the checked-out code and asked for it with `../` in front, and it came
back. So the reach was "any file this program can read", not "any file we committed".
I only found that because I made the test create a real file — my first attempt asked for
files that didn't exist, which passes whether the guard works or not. A test that asks for
something absent tells you nothing.

## The fix, and why it is a small one

There is a list of every file our analyser parsed. It is already in memory, already used
by this same function two lines later, and it is by construction only real Go source —
no documents, no test files, no vendored copies. The fix is to consult that list
**first**, and refuse anything not on it.

That is it. Four lines moved, one error message improved. No new setting for anyone to
remember, nothing to keep in step, and if we ever teach the analyser a second language
the boundary widens by itself, correctly. Asking for a path outside the code is refused
by the same single check, because such a path cannot be on the list.

The best part is that **we are not inventing this rule — we are putting back one we
already had.** The tool this function was originally lifted from does exactly this check,
in the code that calls it. When the function was moved into our platform, the caller was
rewritten and the check was left behind. So the honest description of the bug is *a safety
step got lost in a move*, and the honest description of the fix is *put it somewhere a
future move can't lose it.* That is also precisely the generalisation the review seat
asked for, which is a good sign.

## Where we are now

The change is written, tested in both directions (it fails against the old code — six
files leaked, including the one from outside the checkout — and passes against the new),
and it does not disturb the rest of the package. It has gone to the council for review
and to the landmine verifier for an independent read. It is committed, so it will be in
the next build of the chassis whoever makes it.

One honest caveat: this **narrows** what the function will do. I have checked every place
that asks it for something and none of them needs the old freedom — and there is a
separate, purpose-built way to fetch arbitrary files when we genuinely want one. But I
have asked the council directly whether there is a case I have not thought of, rather than
assuming there isn't.

## Two things I found and deliberately did not fix

I would rather these were visible than tidy.

**A separate flaw sitting right next to this one.** When the code bundle gets too big, the
loop *stops* instead of *skipping*. So one oversized piece of code silently discards
everything the AI asked for after it. That is a different problem with a different reach,
and folding it in would have made this change hard to review. Flagged to the council; it
wants its own ticket.

**A trap in our own process, which cost me nothing this time and will cost someone
later.** Our standing instructions say that after writing a warning into the shared
landmines file you should run a sync command. I did. It told me the new entry needed
verifying. I then ran the tool that does the verifying — and it said *"nothing needs
verification"* and did nothing. The reason is that the signal is used up by the first
command, so the second one finds a clean slate. Everything reported success; no check
happened. I fired the verifier by hand instead, and wrote the trap down. **Our
instructions name the command that breaks it**, so I have flagged that rather than editing
the shared rule-book on my own initiative.

## The bit worth remembering

Both of the original report's mistakes point the same way. It was careful, written by a
component whose job is to be careful, and it was wrong twice — once about who could reach
the code, once about how far the leak went. In both cases the error was **stopping at the
first plausible answer**: the first producer, the first assumption about scope. Neither
needed cleverness to catch, just following the chain one more step and making the test use
a file that actually exists.

## 2026-07-31, end of the session — approved, shipped to the branch, and two new tickets

The council approved the fix first time round, with two advisory notes and nothing blocking.
One note said my two comment corrections were scope creep; I have kept them, because the bug
report specifically asked whoever took it to fix those comments, and one of them is the
comment that misled the report in the first place. The other note was better than my own
judgement: I had found a second flaw in the same function — the one where the code bundle
*stops* instead of *skipping* when it runs out of room, silently throwing away everything the
AI asked for after that point — and I had written it down as a footnote for a reviewer to
decide about. The reviewer said, correctly, that a known-shape flaw found while working on
the very same function should be a ticket, not a footnote. So it is now a ticket.

There was also a detour, and it is the most instructive part of the day. Writing the warning
note for this bug tripped our own automatic checker, which came back with "cannot confirm"
and a confident explanation: the code index is out of date. I checked, and it wasn't — every
symbol it claimed to be missing was right there, at the exact version it named. So I went
looking for why, and **I got the answer wrong twice.** Both times I found a real problem in
how the warning notes are written, measured it, wrote it up as the cause, and then re-ran the
checker — and both times it came back exactly as before. The real cause turned out to be two
halves of the checker that cannot agree: one half asks its question in a format the other
half is incapable of answering. It has never once succeeded at this kind of check, across
every run it has ever done, and no amount of care in how we write the notes can help.

What I want to record from that is not the finding but the mistake, twice repeated: **each
time, I read the code that builds the query and reasoned about it, instead of just running
the query.** Thirty seconds of actually executing it would have settled it before I wrote
anything down. The only reason it cost an hour instead of poisoning a handoff is that I made
a habit of re-running the thing after claiming to have fixed it — which is the cheap check
worth keeping.

The fix itself is committed, so it goes out with the next build of the system whoever makes
it. I have deliberately not built and rolled it myself: another session has a build already
in flight, and rolling the system mid-flight kills any review that happens to be running.
The ticket stays open until someone confirms it is actually running in the live system — a
commit is not a deployment here, and we have been bitten by treating it as one.
