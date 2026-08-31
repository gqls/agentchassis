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

## 2026-08-26 evening — it is live, it is armed on a third of the estate, and it has not fired once

The new build went out and carries the code. I checked that at the binary rather than trusting the
version number, and I checked it the careful way: I asked for two strings that must be there and two
that cannot possibly be there. All four answered correctly, which is the only thing that makes the
first two mean anything — a search that matches everything gives the same answer as one that works.

Worth noting for anyone who tries this later: the service prints a line at startup saying which
version of the code it is, and on a busy machine that line had **already scrolled out of view ten
minutes after it started**. It is not something you can rely on. The binary itself has no such
expiry.

So I switched it on — but only on **two of the six** places that write page content, deliberately. The
reviewers were right that turning it on everywhere at once is a lot of exposure for one unproven
piece of code. The other four are written, held back, and will not run until the first two have shown
they work.

**And then nothing happened, which is the interesting part.**

No pages with buttons have been saved since I switched it on, so there has been nothing for it to
look at. Zero findings — but zero of nothing.

I nearly got this badly wrong. I had counted "how much work has gone past since this was switched
on?" and got 82 items, several of which should have triggered it. That looked like a broken
mechanism, and I started writing it up as one. Then I checked what time the switch actually flipped:
**22:17**. I had been counting from **20:40** — roughly when I sat down. Everything I was looking at
happened *before* the thing I was testing existed.

That is a nastier mistake than forgetting to check. I *had* built the safeguard that asks "was there
anything to find?" — I just pointed it at the wrong moment, and a safeguard aimed at the wrong moment
reports with exactly the same confidence as one aimed correctly. It manufactured a problem out of a
correct silence.

**So the honest state is: built, reviewed, approved, deployed, switched on in a third of the estate,
and completely untested in the wild.** The next person needs to watch for the first finding, confirm
it arrives from both of the two places rather than one, and only then switch on the remaining four.
If it only ever appears from one, the coverage is failing quietly — which is precisely the thing the
whole six-way count exists to catch.

**One improvement landed today from another team.** They pointed out a second thing this cannot see,
and it is a good catch. My check compares two things that both *name a page*. But a lot of button
copy does not name a page at all — it names a *kind* of destination: "book a discovery call", "write
to us". Ninety-five of the hundred and eighty-six problems are in that bucket, and I had described it
only as leftovers. They have a live example: a button that says "write to [an address]" which opens a
cost calculator. My check is silent on it, because neither side is naming a page.

They specifically asked me *not* to fix this inside my check — which was the right call, and I want
to record that they made it rather than me. Teaching one check to answer two different questions is
how these things rot. Instead I have made my check say *why* it stayed silent, so their check can pick
up exactly the cases mine puts down. Small change, and it turns a large pile of "don't know" into
something someone can act on.

**The thing I remain uncomfortable about is unchanged**, and I would rather say it twice than let it
pass: this records a problem, it does not fix one. What matters is whether anyone notices when the
number moves. Right now that is a note in a document asking someone to look once a month, which is
not a mechanism. Another team filed a bug today about exactly this class of thing — work that
completes, reports correctly, and lands somewhere nobody reads. If they build a general answer, this
is one query away from using it.

---

## 2026-08-31 — it fired, from both places, and it is now switched on everywhere

**The wait is over and the answer was the good one.** Five days ago this check was switched on in
two of the six places that save a page, and had never once run for real — there simply had not been
any page saves since the moment it was armed. Today there have been two hundred and fourteen, and
the check has recorded findings from **both** of the two places, not one.

That distinction was the whole point of waiting. If findings had come from only one of the two, it
would have meant the check was quietly missing half the pages it was supposed to cover — and looking
perfectly healthy while doing it. Both reported, so the coverage claim holds up.

**So I have switched on the remaining four.** As of this afternoon all six ways a page can be saved
run the check. That was the plan agreed at review, held back deliberately until the first two proved
themselves, and the condition it was waiting on has now been met.

**I also checked the thing that actually worried the reviewer**, which was not "does it record
something" but "does adding a new database read to every page-saving pipeline break any of them".
Seven page saves did fail in the last day. All seven failed for an entirely different and correct
reason — a separate safeguard refusing to let a routine save overwrite a page that a tool owns. None
of them involved this change.

**One genuinely encouraging number.** Back on the twenty-sixth, roughly one CTA button in six had
copy that did not match where it pointed. Today it is one in fourteen — and that is not because there
are fewer buttons: there are forty per cent *more* of them, and the count of bad ones still fell.
Two other teams have been repairing exactly this population all week, and it is working.

**One thing I chased down and want to record, because it looked alarming.** A related check that had
been finding ten to thirty problems a day stopped finding any at all four days ago. That is exactly
the shape of "your change broke the detector", and the previous session had explicitly asked whoever
came next to treat any such drop as our fault until proven otherwise.

It is not our fault, and I can show why rather than assert it. The system that runs that check is
still running — its sibling checks found seventy-four problems in the same four days. The pile of
problems it looks for has genuinely shrunk, per the numbers above. And ninety-nine of its earlier
findings are still sitting open awaiting a human, which by design stops it re-reporting the same
pages. Three independent reasons for the silence, none of them a broken detector.

I will not claim it is *proven* — the conclusive test would be to feed it a page I know it should
object to and watch it object. I did not do that. But every cheap explanation checks out and the
worrying one has nothing behind it.

**What has not changed is the thing I keep flagging.** This still records a problem rather than
fixing one, and the reading of it is still a promise in a document rather than a mechanism. The
difference today is that the number it reports is finally trustworthy — before this afternoon it was
measured in only a third of the estate, so any percentage taken from it would have been misleading.
From today it is a real fleet-wide figure. Somebody still has to look at it.
