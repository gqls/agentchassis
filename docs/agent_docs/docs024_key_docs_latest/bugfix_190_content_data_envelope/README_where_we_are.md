# Where we are — bugs_open/190, the poisoned content_data rows

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04 — what this bug actually is, in ordinary words

Every page on our sites is stored twice. There is the HTML a visitor sees, and there is the
structured data underneath it — the headline, the body text, the list of features, each in
its own named field. The second one is the important one, because whenever we rebuild or
refresh a page, we rebuild it *from that structured data*. The HTML is the output; the data
is the source.

Two live pages have rubbish in the source. Not wrong content — the wrong *shape* entirely.
Instead of "headline: …, body: …", they hold the delivery envelope the AI writer's reply
arrived in: a little wrapper saying "this is text" with the entire real reply crammed into a
single field as one long string. It is roughly the difference between a filing cabinet with
labelled folders and a sealed envelope with everything inside it shoved through the letterbox.

**Both pages look completely fine today**, which is the whole problem. Their HTML is real and
serves properly. Nothing is visibly broken, no alarm has ever fired, and the pages have been
like this since 15 July and 3 August. But the moment either page is rebuilt, the rebuilder
goes looking for "headline" and "body", finds an envelope instead, and produces an empty or
garbled section. It is a landmine, not a fire.

We already fixed the *cause* of this back in July — the AI writer no longer produces these
envelopes. What we never fixed is that nothing stops one being *saved*. So the two that got
through in the spring are simply immortal: every rebuild reads the envelope, writes the
envelope straight back, and moves on.

## 2026-08-04 — two things the original bug report got wrong, and one I nearly got wrong myself

The report was written yesterday evening by another session, and it is good work, but two of
its specifics had already gone stale or were mistaken. I want both on the record because they
are the kind of thing that gets copied forward.

**First**, it describes this as leftover historical residue. It is not. One of the two rows it
names no longer exists — the page was rebuilt at 22:35 on 3 August, *an hour after the report
was written*, and the rebuild produced a brand-new row with exactly the same rubbish in it.
So this is live and recurring, not a museum piece.

**Second**, and more consequential, the report tells whoever fixes it to detect these rows by
looking for "exactly two fields". That test would have missed one of the only two examples we
have. The finetuning.uk row has *three* fields — the envelope's two, plus a real one that got
attached later. Had I followed the instruction in the file, I would have shipped a guard that
was blind to half the known population and reported success.

**And the mistake I nearly made myself.** There is a history table recording every time one of
these pages was overwritten, and it holds 65 such records, the most recent yesterday. I
drafted the sentence "the system has written 65 of these" — a much more alarming bug, and one
that would have justified dropping other work. It is wrong. That table stores what was
*replaced*, not what was written, so 65 counts rebuilds of already-poisoned pages. The thing
that settled it was reading four lines of code inside the save routine; no amount of further
database querying would have told me, because every query returns a confident number under
either reading. Doing it properly was worth it twice over: it showed the problem has touched
**25 pages across 6 sites** historically — far more than the two live ones — and that fresh
poisoning genuinely stopped in mid-July, so what we are fixing is the survival of old poison,
not the creation of new.

## 2026-08-04 — what I have built, and the one judgement call in it

A guard that sits at the point where page content gets written to the database, and looks at
the shape before letting it through. If it sees an envelope, it does one of two things.

**If the envelope can be opened losslessly, it opens it** and stores the proper labelled
fields. The finetuning.uk page is this case — its envelope contains a perfectly good article,
just wrapped up.

**If opening it would lose anything, it refuses the save outright.** This is the judgement
call and it is worth explaining, because it means one of our pages will now start failing
loudly. The gaswholesalers.com page *can* be opened mechanically — but what comes out is a
131-character fragment, while the actual page copy sits outside the wrapper in a form no
machine can reliably attribute to the right fields. Opening that one would silently replace a
real page with a stub. So the guard refuses, the page keeps serving what it serves now, and a
human has to look at it. That page already has a "needs human review" ticket against it.

I want to be plain that this is a deliberate trade: **we are choosing a page that fails
noisily over a page that quietly guts itself.** The failure will recur on every automated
rebuild attempt of that one page until someone fixes it by hand. If that noise becomes
annoying the answer is to fix the page, not to loosen the guard — loosening it by one line is
exactly what would destroy the content.

## 2026-08-04 — the state of play, and what is not done

The guard and its tests are committed and have gone to the review council, which is running
now. I have deliberately **not** committed one of the two wiring points yet, and the reason is
a good illustration of what this shared workspace is like: another session is part-way through
its own change in the very same file, and its work-in-progress calls a file it has not
committed yet. If I committed that file now I would drag their half-finished work in with
mine and break the build for everyone. So that one line waits until they land theirs.

Two things are still owed after that. The change is inert until the next chassis rebuild —
it is Go code, so nothing happens until an image ships. And once it is live, the finetuning
page should be repairable by simply rebuilding it through the normal machinery, which is the
outcome I want: no hand-written database surgery, the framework repairing its own page.

Fourteen tests, and I ran four deliberate sabotages of my own code to check the tests could
actually detect a broken guard. All four failed the right test and passed again when I put the
code back. That mattered here more than usual, because this guard's normal behaviour is *to do
nothing at all* — and a guard that is completely broken also does nothing at all. A green test
run proves very little unless you have watched it go red.

## 2026-08-05 — it is live, and the machine repaired its own page

The new chassis went out overnight and the guard is in it — checked in both running copies of
the service rather than trusting the version number.

The first look was alarming and turned out to be nothing. The guard had recorded no activity at
all, the count of bad pages was unchanged, and a rebuild of one of the two bad pages had
*completed* a few hours earlier. That is exactly what a broken guard looks like. It was timing:
the new copies of the service only started this morning, so everything I was looking at had run
on the old one. Worth remembering — "it is live and it did nothing" is usually a question about
*when*, not about the code.

Then the repair, which is the part I am pleased about. The finetuning page's envelope contained
a perfectly good article, so rather than editing the database by hand I backed the row up and
asked the platform to rebuild the page in the ordinary way. Two minutes later the page's stored
source had gone from the sealed-envelope shape to proper labelled fields, the guard had logged
what it did and why, and the live page still serves correctly. **No hand surgery — the
framework repaired its own page**, which is the whole point of fixing things at the seam rather
than patching rows.

One honest wobble on the way: I nearly talked myself out of that repair. Two database
measurements suggested the article inside the envelope did not match the one beside it, which
would have meant the guard refusing rather than repairing, and I had started writing that the
row needed manual surgery after all. Running the actual code on the actual data took three
minutes and said the opposite — they matched exactly. The measurements had compared two
different things in two different units. It is in the wrong-calls log, because the tempting
move was to trust a quick database query over the code that actually makes the decision.

**Where that leaves us: the bug is closed.** One page still holds the bad data — the
gaswholesalers pricing page — and that is deliberate and expected, not unfinished work. Its
content genuinely cannot be recovered by machine, it has had a human-review ticket since April,
and the guard now stops it spreading. I have written down in the ticket that a count of one is
the *finished* state, so nobody later reads it as a fix that half-worked.

The one loose thread is a separate, narrower question — whether the same envelope can reach the
page *rendering* path, which my fix does not cover — and that is filed as its own ticket with
the measurement that decides it. The review council specifically asked for it to be a tracked
ticket rather than a paragraph in my reasoning, and they were right.
