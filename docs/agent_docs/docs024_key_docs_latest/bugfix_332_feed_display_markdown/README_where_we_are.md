# Where we are — feed markdown on the news pages (bug 332)

Plain-prose log for the owner. Append-only, newest at the bottom. No jargon where a plain word
will do.

---

## 2026-09-03, late afternoon — what this is, and what we found

You saw raw markdown on the boxingonline news page in your second review — things like
`[NLL (Lacrosse)](https://www.espn.com/boxin...` printed as text where a normal link should
be. That got routed to an existing bug, 332, which had been filed back in August as
"harmless, watch it". Nobody was working on it. We are now.

**Three things turned out to be true, and only the first was already written down.**

**First: the cleaner we already have works, but it can only see finished markdown.** Back in
August we taught the news page to strip out markdown symbols before showing anything to a
visitor. That fix is real and it is running. The proof is a small one: those pages carry *no*
stray `#` headings at all, even though over a thousand stored news items contain them. So the
cleaner is doing its job — it just cannot recognise a *half* of something. Every single
problem visible on the site today is a half: a link that was chopped in the middle, so the
closing bracket never arrives and the cleaner walks straight past it.

**Second: we are the ones chopping them in half.** The August bug file said "something
truncates these before we see them" and was honest that it had not gone and looked. I went and
looked. It is our own code, in the search adapter: it takes the summary the search provider
gives us and cuts it to 200 characters. If the cut lands in the middle of a link, we have
manufactured exactly the shape our own cleaner cannot handle.

There is a second cost to that which nobody had noticed. In the summaries that contain links,
**a third of the 200 characters is spent on the web address itself** — an average of 69
characters of `https://…` that no visitor ever sees. If we cleaned first and cut second, the
customer would get about 69 more characters of actual writing in every one of those summaries.
That is a content-quality win sitting in plain sight.

**Third, and this is the one that changes the size of the job: the page is not the only place
this text is shown, and the other place is worse.** Alongside every news page we publish a
data file — `news-archive.json` — that anyone can open in a browser. It goes through a
*different* piece of code that does no cleaning at all. It is currently serving seven stray
headings, nine broken links, and various other debris on a paying customer's site.

And it is not sitting there quietly. Every news page loads a small script that fetches that
file and **replaces the cleaned page with the uncleaned data**. So for any visitor with
JavaScript on — which is essentially all of them — the cleaning we did in August is cosmetic.
They are reading the dirty copy.

I got that last part wrong at first, and it is worth saying how. I searched the page's own
HTML for the name of that data file, found nothing, and concluded the script was dormant. But
the script lives in a separate file the page merely *points at*, so searching the page could
never have found it. One command fetching the script settled it. A clean "no results" felt
like an answer, and it was not one.

**And it is not one site.** Five sites are serving this right now: boxingonline,
fundamentallyai, robot-hands, ai-agent-orchestration and idea.uk.

## What we are going to do

The heart of it is one small piece of tidying with a large effect. Right now three different
bits of code take that same stored news text and each decides for itself how to prepare it for
display — and only one of the three cleans. We are replacing that with **one** shared piece
that all three call. After this, "read the news text for display" and "clean it properly" are
the same act, and a fourth reader added next year gets it for free.

Then we teach the cleaner to recognise chopped-off markdown, which is what fixes what you saw.

**One thing we are deliberately not doing, and I want to flag it because it is a change from
what you picked.** You chose to fix the truncation at source as well. I had two independent
reviews done of the plan, and both came back against touching that particular file — for
reasons I did not have when I asked you. The short version: cleaning at that point writes the
change permanently to our stored records, where cleaning at display time can be switched off
and undone at any moment. That same code also feeds our own AI models, so quietly changing it
would leave us with a training archive that means two different things either side of the
change, with nothing marking where. And another session was actively editing that file this
afternoon.

So instead I have written the finding up and handed it to the team that owns that code, with
the measurement attached, and offered to make the change with them. There is one part of it I
would happily do straight away because nothing is contested about it: the cut is done by raw
byte position, which can slice a letter like an accented character clean in half. Two records
already have the resulting corruption. That is a one-line fix using a helper we already have.

**One good piece of news on effort.** I checked whether the nine already-damaged pages would
need a repair campaign. They will not: every one of them was rewritten within the last
nineteen hours, three of them within the hour, because the news feed refreshes those sections
continuously. Fix the code and the pages repair themselves within about a day.

**Two things found along the way, both handed on.** The script that overwrites the page also
inserts that third-party text without escaping it, which is the shape a security problem takes
even though nothing dangerous is in there today — 14 records out of 5,863 contain any HTML at
all and none of it is executable. I asked the components team whether anything else in the
library does the same; they checked all 23 components that use that technique and found ours
are the only two, and pointed me at the safer pattern the other twelve already use. And the
summaries themselves sometimes contain ESPN's own site menu — `Tennis`, `NFL`, `MLB` —
scraped in as if it were article text. Cleaning the markdown will not fix that; it is a
different problem in how we capture the articles, and it has gone to the team that owns it.

## 2026-09-03, evening — done and committed, waiting on a roll

The work is finished and committed. It is **not live yet**: these are Go changes, and Go
changes do nothing until someone rebuilds and rolls the chassis image. That is the honest
status — the code is right, the pages are still wrong until the roll.

**What went in, in plain terms.** One shared piece of code now prepares news text for display,
and all three places that show it call that one piece. Before, three bits of code each decided
for themselves and only one of them cleaned anything up. That is the whole fix; everything else
is detail.

Alongside it, the cleaner learned to recognise markdown that has been **chopped in half**, which
is what all the visible damage actually was. And the off-switch we already had now reaches all
three places instead of one — which is what the original bug report asked for two weeks ago and
what the old arrangement simply could not do.

**The team that owns the search code fixed their half themselves.** I sent them what I'd found
rather than editing their file, and they came back having made the cut stop chopping links in
half at source, and made it safe for accented characters. They also checked my numbers against
the database themselves before acting, which I'd rather they did. They turned down the more
invasive change for the same reasons I'd given, having got there independently.

**Two independent reviews of the plan, and one of them saved us from something bad.** The
obvious way to clean up a chopped-off link also deletes the "…" that shows something was cut —
so a broken fragment would have become a smooth, complete-looking sentence that the source
never actually wrote. On a paying customer's page. And no test or automated check we have could
ever have spotted it. That is now fixed and specifically tested for.

**The formal review board approved it**, with four advisory comments, and two of those were
right. One asked a question I couldn't answer well, so I answered it in the code. The other
pointed out I'd checked my change against the news data but not against the *other* seven
places the same shared code is used — a fair hit, and exactly the kind of gap that causes this
sort of bug in the first place. I've now run it over all 40,318 pieces of text in that other
population: nothing was emptied, nothing broke, and the fifteen things that changed were all
the same defect being fixed.

**One thing I did differently from what you chose, and I want to be straight about it.** You
picked "fix it at source as well". Both reviews came back against touching that particular
file, for reasons I didn't have when I asked you — chiefly that cleaning there writes the
change permanently into records that also feed our own AI training, where cleaning at display
time can be switched off and undone at any moment. So I handed it to the team that owns it
instead, with the measurements. They took the safe parts and declined the rest. I think that
landed in the right place, but it was my call to make it, so it's flagged rather than buried.

**One thing needs a human before it goes live.** The small database change that fixes the
script security issue is written, rehearsed against the live system and rolled back cleanly,
and deliberately **held** rather than queued to apply automatically. The reason: it changes how
a page draws itself in the browser, and none of our automated checks can actually see a browser
— they'd all pass on a page that came out blank. Someone should apply it and then simply look
at a news page.

**What is left.** After the next chassis roll: re-check the five sites, both data files and the
feed. The affected pages should repair themselves within about a day without anyone doing
anything, because the news feed rewrites those sections every few hours — and if they don't,
that assumption was wrong, not the fix.

**What is still not fixed, and you will still see it.** Some summaries are ESPN's own site menu
scraped in as article text. Cleaning the markdown removes the symbols and leaves the words, so
the news page will still read a little oddly in places. That's a different problem in how we
capture articles, it's filed, and the team that owns it has acknowledged it.
