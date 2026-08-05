# Summary — 2026-08-05 — the check that asks whether a page's JavaScript is really there

## What we're trying to do

Close a blind spot. A web page can tell a browser to load a JavaScript file or a
stylesheet, that file can be missing, and nothing we own would ever notice. The
page still renders. Every status in the database reads "built and deployed".
Nothing looks wrong until a visitor clicks a button and nothing happens.

JavaScript is uniquely exposed to this, because its absence changes nothing you
can see. An image that fails leaves a broken-image icon; a script that fails
leaves a page that simply doesn't work. That is the whole of bug 084.

## Where we've come from

The bug was filed on 26 July and had sat untouched since — one commit, nothing
after it. It listed five possible fixes. Two turned out to be closed to us before
any code was written, and finding that out was most of the value of the research:

- **The obvious fix was already ruled out, months ago.** We have a check whose
  name promises exactly this — it is called "asset loads" — and it does not load
  anything; it looks for the filename as text in the page. Changing that seems
  like a one-line improvement. Another team proposed precisely that and the review
  council ruled it off-limits: the check is shared vocabulary, and changing what
  it *means* changes the verdict on every document already using it. They paid a
  real price for the ruling — a legitimate check of theirs had to be marked
  "impossible for now" and parked rather than smuggled through.
- **Another of the five was already done** and nobody had updated the bug.
  Widening the browser-testing tier past tool pages happened on 29 July. Two open
  bugs still describe the world as it was before that.

## What we've done

Built the remaining high-value item: a check that takes every script and
stylesheet a live page asks for, works out the real address, and requests it.

The care went into what it does when the answer is ambiguous. It reports a
problem *only* for a definite "this file does not exist", and it asks twice before
believing it. A refusal, a rate-limit, a timeout, a server error — all recorded as
"couldn't tell", reported as nothing. That is not timidity: someone previously ran
a simpler version of this against one of our sites and its security layer refused
all 63 requests. Under a looser rule that is 63 false alarms about a site where
everything was fine. This check can be blinded, and says so in its logs; it cannot
be tricked into lying.

We also measured the problem honestly before building anything, and the answer was
uncomfortable: **there is nothing broken right now.** All 541 live pages, 854
script references, 96 distinct files, every one loading correctly. So this is a
smoke alarm, not a repair — and everything we wrote says so, rather than letting
the work read as if it fixed live damage.

Because there is nothing live to prove it against, we proved it by deliberately
breaking each of its safety rules in turn and confirming a specific named test
caught each one. Six of those, including the one added after review. A guard you
cannot make fail on purpose is decoration.

The review council approved it first time, with five advisory comments. One was a
genuine catch: our database query used a home-made test for "is this page live",
where a shared one exists precisely because home-made versions keep missing pages
— 28 of them. Ours would have quietly skipped those, and a check that skips pages
reports "all clear" for the wrong reason. That is the exact failure this work
exists to prevent, so it was a pointed thing to have got wrong. Fixed, and wired
so that repeating the mistake breaks 21 tests immediately.

## Where we are now

Written, tested, reviewed, approved and committed — and deliberately switched off.
If you name a check the running software doesn't know about, the entire scan
fails, so it can only be enabled after a build carrying it goes out. We verified
this evening's build does not contain it yet, which is expected: the commit came
after it.

Two missteps are on the record, and the second is the one worth carrying. First,
our initial measurement matched a script reference that turned out to be a comment
inside a tool's own code — a text search cannot tell a page loading a file from a
page mentioning one, and tool pages are the worst possible place to make that
mistake. Second, and sharper: the experiment meant to prove the new safeguard
worked reported "nothing failed", which would mean the safeguard was useless. It
hadn't run at all — the edit silently didn't apply, so a blank result got read as
an answer. That is the same error the entire check exists to prevent, made by us,
in our own tooling, on the same day.

One objection we could not answer is left standing for a person: this adds another
kind of alert that nobody is assigned to act on, and we are creating those faster
than we drain them. Our rules permit it — the repair here is a human judgement —
but the reviewer is right about the trend, and it is recorded rather than argued
away.

## Where we're going

Four steps, in this order, and the order is not negotiable: a build that carries
it; a check of the running software with both a positive and a negative control;
switching it on, together with the test fixture that records what is switched on;
and then deliberately breaking one page to watch it catch a real fault and clear
itself when repaired.

Only then can the bug close — and even then, not entirely. Two of its five
original suggestions are untouched, and one of them overlaps work other threads
are already doing. The honest end state is a narrowed bug, not a closed one.
