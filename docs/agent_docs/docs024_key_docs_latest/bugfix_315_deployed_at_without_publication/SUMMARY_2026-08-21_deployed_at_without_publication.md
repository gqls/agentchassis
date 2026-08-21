# SUMMARY — 2026-08-21 — the last piece is built, and it found a hole in its own foundation

*(A new file, not an edit of the 08-20 one. The series is the record.)*

## What we're trying to do

Make the platform able to answer one question honestly: **is the website actually serving the page we
think we published?** For most of this system's life it could not. It knew a page had been *asked* to
publish, and it recorded that as if it were the same thing.

## Where we've come from

A tool page on a customer site carried on serving its old version for about six hours, through four
completed rebuilds, while every internal signal said everything was fine. The investigation found
that `deployed_at` — the timestamp the whole estate read as "this page has shipped" — was written
whether or not anything had been sent. Two of the five agents that wrote it stamped it *before* the
deploy was even dispatched.

The deeper finding reframed the bug: the platform could not tell **"this page never needed
republishing"** from **"this page failed to republish"**. Those two are identical in every signal it
produced, which is why nobody noticed for six hours.

Over the previous two days we fixed the writing side: the deploy step now reports what it actually
committed, and the page records a fingerprint of the exact bytes that went out. That shipped and is
live.

## What we've done

Built the reading side — the piece the whole bug was pointing at. A check that fetches the live page,
fingerprints what comes back, and compares. Committed, registered, and reviewed by the council, which
approved it.

Before writing any of it, we ran the comparison by hand across every page that has a fingerprint:
**228 pages, 12 sites, 228 matches.** Every page we believe we published is serving exactly the bytes
we sent. So the check ships as a smoke alarm, not a repair — and those 228 agreements are themselves
the proof that the fingerprint mechanism works end to end.

We also measured how long publishing actually takes, because the obvious design — check the moment a
page goes out — would have been wrong. Watching every freshly published page for nearly three hours
(1,099 measurements), publishing turned out to be batched: three times a page genuinely did not match,
all three within fourteen seconds of being published, all three healthy and correct within about two
minutes. An eager check would have raised **three false alarms in under three hours**. It now ignores
anything published in the last thirty minutes, which costs us nothing against a fault that lasted six.

## Where we are now

**The check is built and deliberately not switched on**, for two reasons — one routine, one not.

The routine one: it needs the next fleet release before the configuration that enables it can be
applied, or it would break the checks already running.

The one that matters: **the review caught a false claim at the centre of our own safety argument.**
We had written, in six places, that every route which marks a page as published also records its
fingerprint — "three routes, all correct". A council reviewer said that number had almost certainly
been measured with a query that only looks one level deep, and asked for it to be re-run. It was
right. There are **six** such routes and **three of them do not record a fingerprint** — and those
three are the main page-building paths, the ones that actually produce new content. Left alone, this
check would eventually have accused perfectly healthy pages of being broken, convincingly and
permanently.

Nothing is broken today: no page is in that state, and we can see that because the 228-page sweep
would have shown it. But that is an observation about one moment, not a property of the system.

So the enabling step now **refuses to run** while any of those routes is still unfixed. It is not a
warning in a document that someone has to remember; it is a gate that stops the change, and we
watched it stop the change before committing it.

The reviewer worked this out from the *shape* of our claim — a confident small number about "every
route that does X" — without ever seeing the query. The verdict was still "approved". Had we read the
decision and skipped the objections, we would have shipped it.

## Where we're going

1. **Fix the three routes** so they record what they send. It is a small, well-understood change of
   the same shape as one we made two days ago — but it touches the busiest path in the system, and
   the last time we changed that path we took page publishing down for thirty-three minutes. It gets
   its own review and its own checks, not a footnote in this one.
2. **Then switch the check on**, after the next release. The gate will let it through once step 1 is
   done, and not before.
3. **Then close the bug.** Of its four proposed fixes, two are delivered and live, this is the third,
   and the fourth cannot be answered from inside our own systems at all — the publishing runner lives
   in a repository we do not control, which is precisely why this check was designed to detect the
   failure from our side instead of explaining it from theirs.

The honest summary of today: we finished the thing we set out to build, and the review process caught
us making, inside the fix, a smaller version of the same mistake the bug was about — trusting a
measurement that could not have come out any other way.
