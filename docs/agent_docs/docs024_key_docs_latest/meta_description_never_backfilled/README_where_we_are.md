# Where we are — meta descriptions

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-08-19 — I went to fix five pages and found 407

I picked up the job that was handed on from the last lane. It was meant to be small. The
Platform Log index on fundamentallyai.com lists six articles as cards, and none of the
cards is clickable — you can see the article titles but you cannot get to the articles.
The repair for that was already designed, approved by you, and applied on 08-18. It did
not go live because the system refused to write the new version of the page, and the
note left for me said why: five of the eight articles have no "meta description" — the
one-sentence summary that sits under a page's title in a Google result and, on this
page, is what each card shows as its blurb. With five blurbs empty the new card list came
out 42% the length of the old one, and there is a safety guard that refuses to replace a
page section with something less than half its previous size. That guard is right, and I
have not touched it: it is the thing that stops a bad render silently hollowing out a
page, and it applies to every page rebuild on the estate.

So the job was: get those five descriptions written. The instruction attached to it was
yours, from 08-06 — the framework writes the content, not me. I agree with that and I
have stuck to it.

**The first thing I found is that the note's explanation was wrong.** It said the
descriptions were missing because of a known queue problem: the system detects the
fault, but the detection has no one assigned to fix it, so it sits there. That is a real
problem elsewhere, but it is not this one. The check it named looks at three things on a
page — the title, the "skip to content" link, and the footer. It has never looked at
meta descriptions at all. So fixing that queue would not have filled these five, and I
would have spent the day on the wrong thing. It took one look at the check's own source
to see it.

**The second thing I found is that it is not five pages.** Across the whole estate,
**407 of 731 live pages — 56% — have no meta description.** Twenty-six of our
twenty-seven sites are affected. Three of them have none at all, on any page:
loancalculator.co.uk, adversecreditmortgage.co.uk and loanzy.uk. Every one of those
pages is currently being served to Google with nothing to show under its title.

**The third thing is why, and this is the part worth your attention.** I traced it to a
single omission. When the planner agent designs a site, it is given a template of what
to return for each page — name, title, page type, nav label, ordering, and the list of
sections. **That template has no field for a meta description.** The code that writes
the page into the database then asks the plan for one anyway, and takes an empty string
when it does not find it. So every page the planner creates is born with no description,
and always has been.

And nothing afterwards ever fixes it. I checked every route:

- Nothing in the codebase updates that column on a page that already exists. All seven
  places that write it do so only when the page is first created.
- None of the 58 automated site checks looks at it, so we do not even detect it.
- It is not a field a writing agent can reach — the writers work on sections inside a
  page, and this is a property of the page itself.

I also tested the one route that looked promising, because there are already jobs in the
queue complaining about missing meta descriptions, assigned to a handler. Two of them
have completed. **Neither actually wrote a description** — both pages are still empty
today, and one of them was demonstrably touched by the handler on the way through. So
the system can file the complaint, mark it done, and change nothing. If I had filed jobs
for our five pages, they would have come back green and the page would still be broken.

**The last thing, and it changes what the fix has to be.** I assumed this was a backlog
to be cleaned up. It is not. Of the pages created this month, **53% were born with no
description** — the same rate as every month before it. The tap is still running. So
filling in the 407 without fixing the planner's template would just mean doing it again
in a month.

### Where that leaves us

The five pages that block your Platform Log fix are ordinary members of a much larger
problem, and I have deliberately not papered over it by writing the five descriptions
myself. Your 08-06 instruction covers this exactly: if the generator does not exist yet,
that is the finding to report, not a gap for me to fill personally. It does not exist
yet.

What I have not done, and want your steer on, is build the fix. There are two halves —
stopping new pages being born empty (a small change to what the planner is asked for),
and filling the 407 that already exist (which needs something built that does not exist
today, and which writes public copy that appears under your name on 26 sites). The
second half is not a change I should make on my own judgement, so I have written up the
options rather than picked one.

I am putting the evidence through the system's own diagnosis loop before I write any of
it down as settled — that is the standing rule for a claim this broad, and it is cheap
insurance against my being confidently wrong in a document other people then believe.
