# SUMMARY — 2026-08-10: the Gauntlet round page is out of the search index

First summary for this lane, written at the milestone that earns one: the fix is live and
proven. Plain prose, written to be read aloud.

## What we're trying to do

Stop published Gauntlet rounds turning up in search results, without changing anything else
about how they work.

The Gauntlet is the tool on vonc.com where a visitor argues a position against the AI and
gets a verdict. If they like the outcome they can press share, and their round becomes a
permanent public page. That page carries two kinds of writing that need care: the visitor's
own prose, and a verdict our service wrote *about* it. Today those rounds are almost all our
own test traffic. The risk is what happens the first time a stranger publishes a round that
argues about a real, named person — because then that person's name becomes a search term
that returns our page.

The reason this was worth doing quickly is the shape of the trade: it costs us **nothing**.
A shared link keeps working exactly as it did. All we give up is being *findable by
strangers searching*, which for this page we never wanted.

## Where we've come from

Another thread found it on 9 August and wrote it up as bug 232, deliberately splitting it
out of a much larger open question about third-party harm so that the cheap, obviously-right
part could be done straight away. That was a good call and it's why this took two days
rather than joining a queue behind a design debate.

We picked it up the same day. The first surprise was that **both of the obvious fixes are
impossible here**, and establishing that was most of the first afternoon. The tidy fix is to
have the web server attach a "don't index" instruction to the response — but there is no web
server of ours in front of that page; it's a static file in cloud storage with a CDN in
front. The second fix is to put the instruction in the page's own template — but that
template is only the middle of the document, and the instruction has to live in the header,
which is assembled somewhere else entirely. There *are* server config files in our repo that
look like the answer; they belong to a different service on a different machine, and using
them would have produced a fix that never ran.

We also found something uncomfortable on the way, which we did not go and fix: there are
**two** separate pieces of code that build a page's header, and only one of them has been
getting the improvements. Three in a row now — structured data in July, canonical links a
week ago, and this one — have each landed on one path and not the other.

## What we've done

We put the instruction where the header is actually assembled, switched on by a new per-page
flag that is **off everywhere by default**. One page out of 630 has it on. The mechanism is
reusable: any page can be excluded from search now, which is a framework capability rather
than a patch to one page. We also added the proper server-side header to the API that serves
the round data, since that endpoint *can* take one.

Three things we'd point to as the work being done properly rather than just done:

- **The migration's safety check was deliberately broken first** to prove it could actually
  fail, before it was trusted. A check that cannot come out false isn't evidence.
- **The tests were sabotaged to confirm they'd notice.** Two different deliberate breakages,
  each failing the cases it should.
- **The proof is a pair, not a single page.** We rebuilt the flagged page and it carries the
  tag; we then rebuilt a *different, unflagged* page on the same new code and it correctly
  does **not**. Without that second half we couldn't tell "the switch works" from "the new
  code tags everything".

We put it through the review council, which came back asking for changes. The gating
objection was a fair one and we'd missed it: we had proved the two-code-paths problem in
general but never proved *which path builds this particular page* — and if it were the other
one, the whole fix would have been inert for the one page it targets. Answering that
confirmed the fix is on the right path, and **the same query refuted a claim we'd already
published in three places**: we'd said the other path would silently strip the tag, when in
fact it refuses to touch this class of page at all. We'd checked what that code *does* and
never read what it *declines to start*. Corrected everywhere, visibly, and logged.

## Where we are now

**Live and verified at the real page**, on the build that went out overnight. Verified
against the actual running program on every machine, not the version label — and in both
directions, as above.

One thing is outstanding, and it is not ours to do unasked: the API half runs on a separate
machine with its own deployment process, so the overnight rebuild did nothing for it. The
code is written and committed and needs someone to deploy it deliberately. It's the smaller
half — the page a person actually lands on from a search is done.

Two traps got paid for and written down. The procedure our own runbook prescribes for
rebuilding this page **hung and failed**, having never started the job, on a completely idle
queue; a more direct route did it in twenty-five seconds. And that failure left **no trace in
the table we tell each other to look in** — which matters beyond this fix, because an open
bug about those hangs partly rests on that table reading clean.

## Where we're going

Nothing further is needed for this bug beyond the API deployment. The open thread we're
leaving deliberately is the bigger one: **two pieces of code build page headers and only one
keeps up.** That's now three features deep, and the review council's architecture seat said
the same thing independently — it wants a tracking item, not another comment in the margin.
We've recorded it rather than quietly widening a small bug fix into an architectural change,
which is the thing we've agreed not to do. Someone should scope that convergence properly,
and when they do, our correction matters: the exposure is narrower than we first wrote, and
knowing exactly how narrow is what makes it possible to argue about honestly.
