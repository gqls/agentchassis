# Where we are — the CTA buttons that describe one page and link to another

Plain-prose log, append-only, newest at the bottom.

## 2026-08-26 — what this is, and what we found

The owner reported four broken buttons on dartsonline.com. Rather than fix four things, another
session looked for what they had in common and filed bug 399. The common thread is unusual, and it
is the reason this was worth doing properly.

When the framework builds a page it works out where each button should go, and it writes down two
things side by side: the link, and **the title of the page that link leads to**. It writes the title
down deliberately, so that the part of the system that composes the button's wording can say
something true about where the button goes. Then nobody ever compares the two. The words and the
destination sit in adjacent slots of the same record, and no piece of code has ever read them
together.

So a button can say *"Catch up on this week's darts news"* while the record immediately beside it
says the link goes to *"All Brands"* — and everything reports success.

**The first thing I checked was whether the writer is even told.** If it were not, the fix would be
plumbing. It is told, twice over: once in the button field's own instructions, and once at the
moment of writing, in a sentence that reads *"the destination is fixed — write this text for that
destination; never promise a different one."* Three-quarters of a thousand prompts carried that
sentence in the last three days.

And it still gets it wrong about **one time in seven**. Of the buttons written since that
instruction started reaching the writer, 155 out of 1,060 describe somewhere they do not go. That
number could have come back near zero and it did not, which is what makes it evidence rather than
reassurance. Telling the writer more firmly is not going to fix this. Something has to check.

## 2026-08-26 — what I decided not to build, and why it matters

The bug file proposed the obvious repair: when the words and the destination disagree, rewrite the
words to match the destination. I have not done that, and I think it would have made things worse.

A neighbouring investigation found that the system already picks a lot of destinations badly — it
ranks candidate pages by an ordering that has nothing to do with what the page is about. When it
picks badly and *then* has the wording rewritten to match, the bad choice stops looking like a bad
choice. It becomes a sentence a human wrote, describing a page, and the cheap fix for the ranking
can no longer reach it. Rewriting the words to match the link would have quietly converted the
easy-to-fix cases into hard-to-fix ones, at a rate of about 150 a week.

I also checked whether the system could simply **move the link** to the page the words describe. It
mostly cannot. Of the 186 disagreements live today, the wording clearly names exactly one other page
in **13** of them. In 78 it names two pages equally well, and in 95 it names nothing that exists.
So an automatic correction would reach about 7% while risking damage to the rest — and we already
have one dated case, on 24 August, of an automatic CTA repair turning a *correct* contact link into
a wrong one.

That left recording it. Which sounds like a retreat, and is not: **173 of those 186 buttons need a
human or a writing pass, and nothing anywhere currently tells anyone they exist.** Two other pieces
of work are stalled precisely because they have no list. One is trying to fix the ranking and needs
to know which buttons its fix cannot reach; the other is a commissioned writing pass that has never
run because nobody could say which buttons needed it. The record is the thing both are missing.

## 2026-08-26 — two near-misses worth knowing about

I nearly put the check in the wrong place. The natural spot is where the words are first written —
but it turns out the *repair* path, the one that runs when something has already gone wrong,
doesn't go through there at all. A check in the obvious place would have watched the calm half of
the system and missed the half that actually churns. Both paths do meet, one step later, at the
point where the page is saved. That is where it went.

And I nearly broke a working check while trying to be tidy. My design had the shared piece of logic
reject anything that isn't an ordinary page link — sensible-looking, since a phone number is not a
page. But there is a live check for phone and email buttons that reuses that exact logic and feeds
it phone numbers on purpose. Had I "tidied" it, that check would have gone quiet without anyone
editing it, and it would have looked fine. Caught it by reading the callers before writing the code
rather than after.

## 2026-08-26 — where this goes next

The mechanism is built, tested and committed, and it is inert until the next fleet image rolls —
which is deliberate, not a delay. Once it rolls I need to confirm the recording actually fires from
**both** paths, because seeing it from only one would mean the coverage is failing quietly, and that
is the exact failure this design exists to avoid.

The one thing I want to be honest about: a record nobody reads is not a fix. The session that filed
the bug made that point back to me sharply, and it is right. **What matters is the rate, not the
individual rows** — nobody should read 155 records, but somebody should notice if one-in-seven
becomes one-in-three, or drops to one-in-fifty after a change to how the writer is instructed. I
have put the query and a standing obligation to read it monthly in the runbook, and named it in the
register. That is a promise on paper, and paper promises are the weak point here.
