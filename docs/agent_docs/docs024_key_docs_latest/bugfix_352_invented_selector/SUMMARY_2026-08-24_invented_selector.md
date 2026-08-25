# SUMMARY 2026-08-24 — the invented selector: fixed, shipped, and proven on a live page

*First summary in this lane. Written the evening the fix was proven and the old tickets cleared,
because that is the first point at which the read-out differs from the bug report.*

---

## What we are trying to do

Our sites are checked automatically for text that is too faint to read against its background. When
the checker finds some, it files a ticket saying **which** text, and another agent writes a small
piece of CSS to darken it. The whole chain is meant to run without a person in it.

The thing that makes that chain work is the address in the ticket — the **selector**, the short
piece of CSS that says *"this heading, on this page"*. Get the address wrong and everything
downstream still functions perfectly: the rule gets written, deployed, and marked done. It just
never lands on anything.

## Where we have come from

The checker was writing addresses for elements that had no name of their own. When it met such an
element it fell back to using the element's *type* as if it were its name — so a plain heading was
filed as "the heading called `H3`". That address reads as sensible and matches nothing at all,
because no heading on any of our sites is called `H3`.

The consequences were entirely silent. Rules were authored against these addresses, deployed, and
their tickets closed as fixed. **111 repairs are recorded that could not have touched anything**,
and 73 more were sitting open, waiting for a fixer that would have wasted its run on them. Two
other teams had built work on top of these tickets: one was about to release the 73 to be fixed,
and one had written a piece of evidence around "six headings styled `.H3`" that were nothing of
the sort.

The obvious repair — drop the invented name and just say "heading" — is worse than the bug. Our CSS
is appended to a single stylesheet per site, so an address of "every paragraph" really does mean
every paragraph on every page. The two commonest broken addresses were exactly those: paragraphs
and links, 121 of the 181.

## What we have done

The checker now works out the address **inside the page**, walking outward until it finds something
nameable, and then **asks the browser whether that address actually selects the element it just
measured**. If it cannot produce an address it can prove, it refuses to file the ticket and counts
the refusal. The rule is not "stop guessing", it is "prove it" — which means the next address bug
of any kind announces itself instead of hiding.

That went through the review council and was approved first time. It reached both of the affected
services during the afternoon.

Today we proved it works, and not by reading our own logs. The system re-checks sites on a rotation
of about one an hour, and two of those runs happened to fall either side of the moment the new code
went live. The earlier one filed forty-seven tickets, three of them with impossible addresses. The
later one filed ten, **none** impossible, every one stamped as verified in the page. Then we
downloaded the live pages and counted for ourselves: the software claimed its addresses matched
fifteen things and eight things, and they matched fifteen things and eight things — every one of
them the sort of unnamed element that used to produce a broken address.

We also checked two of the morning's bad tickets the same way. Their addresses match **zero**
things on pages that plainly contain twenty-two and six of the elements concerned. Both tickets are
already marked fixed.

Finally, we withdrew the 73 stale tickets. Withdrawn, not resolved — the faults are still on the
pages, and each withdrawn ticket says so. Clearing them frees the system to find the same faults
again and re-file them with an address that works.

## Where we are now

The half of this problem we set out to solve is solved, live, and demonstrated on a real page
rather than in a database. Both other teams have been told, including the one whose evidence we
had to correct. Nothing is blocked and nothing is owed.

Three things are worth saying plainly rather than leaving in the technical notes.

**The damage figure moved while we were fixing it.** We have been quoting 108 impossible repairs;
it is 111. Three more were filed and closed in the eight minutes before the fix rolled out.

**One of our own numbers carried the wrong date.** A count in this morning's handover was labelled
with the time it was written up rather than the time it was measured — an hour and a half apart,
with a re-check in between. It looked completely current. Corrected, and written up where the next
person will meet it, because a wrongly-dated figure is more dangerous than an unmarked one: it
advertises a confidence it has not got.

**And a fault that is not ours.** The job that re-checks sites **fails more often than it
succeeds** — eleven of twenty runs ~~over the past week~~ timed out after three minutes, and that was
true before we changed anything. It matters here because that job is what will find the withdrawn
seventy-three again.

> **CORRECTED 2026-08-25.** Two things in the paragraph above are wrong, and the summary is left
> standing with the correction beneath it rather than rewritten. (1) It was eleven of twenty **in a
> single day**, not over a week — the table those runs live in only keeps about twenty-four hours, so
> the week-long window I asked for quietly became a one-day one. The failure rate itself is real.
> (2) The re-check job then went thirteen hours without running at all, which looked alarming and is
> not: it works through the sites one an hour and waits three days before returning to any of them,
> so it had simply finished the round. That also means the seventy-three withdrawn faults should
> come back **within days rather than a fortnight** — all thirteen sites are due to be re-checked by
> the twenty-seventh.

## Where we are going

The second half of this bug is untouched, and it is the reason the file stays open: even a
perfectly correct address can lose. Our fix's CSS is appended to one stylesheet, and some of the
offending colours are set in styles that load *after* it, so an equally-specific rule loses on
order alone. That has a sketch and no design. It is the next real piece of work.

Below that: the two new refusal counters have nowhere to be read — today's fleet restart wiped the
only place they appear within the hour, which rather makes the point. And from **7 September** the
withdrawal becomes checkable: any of those thirteen sites still showing a visible contrast fault
with no re-filed ticket means the promise we made when we cleared them has not been kept.
