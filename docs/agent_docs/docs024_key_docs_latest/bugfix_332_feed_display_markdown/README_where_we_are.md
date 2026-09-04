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

## 2026-09-04, late morning — it is live, it works, and checking it found two more things

The build went out last night at 22:06. I checked the actual pages this morning rather than
assuming, and it was worth doing twice over.

**The good news, and it is properly established rather than hopeful.** The fix is running and
doing its job. The proof is one site: dartsonline. Its stored news items still contain the
broken markdown, its page rebuilt itself at 10:40 this morning, and the page came out **clean**.
That combination — bad data going in, a rebuild after the new code, good page coming out — is
the only thing that actually proves anything. Three other sites also look clean, and I am
deliberately not counting them: their stored items had nothing wrong in the first place, so a
clean page there tells us nothing at all. Boxingonline, the site you saw the problem on, has
gone from five instances to **none**.

**The first thing checking found.** One site, idea.uk, rebuilt itself this morning *after* the
new code and still showed a problem. That is exactly the alarm I set up yesterday, going off as
intended. The cause is a shape I had not covered: a chopped-off **image** rather than a
chopped-off link. Fixed this morning.

**The second thing is the one I am least comfortable about, so I want to be direct.** While
writing the test for that, I found that one of the two halves of yesterday's change had **never
been connected**. The half that *cleans* the pages was working — that is why your pages are
getting better. The half that *detects* the problem, so our automated checks can spot it coming
back, was written, documented, described to the review board as connected, and then never
actually plugged in. So for a day our checker would have declared a bad page clean.

Nothing was harmed by it — the cleaning worked throughout, which is why the pages improved
anyway — but I told the review board something that was not true, and the board could not have
caught it, because it reviews the plan and my plan said it was done. Nor could any of our
existing tests: they all run through the very function the pattern was missing from, so the
missing piece was invisible to every one of them. It is written up in full. Also fixed this
morning.

**What that means for you: we need one more build.** Both of today's fixes are written and
committed but do nothing until the chassis is rebuilt and rolled. Until then, chopped-off images
will still appear on the news pages and our checker stays half-blind.

**And a reminder of the two things still waiting on a person, not on code.** The small database
change that fixes the script security issue is approved and ready, and deliberately held back so
that someone applies it and then *looks at a news page in a browser* — because none of our
automated checks can see what a browser sees. And the ESPN-menu problem, where a site's own
navigation gets scraped in as if it were an article summary, is filed and its owning team has
decided not to take it yet. **The news pages will still read oddly in places until that one is
done**, and no amount of markdown cleaning will change it.

## 2026-09-04, 16:05 — everything is live now, and one of my explanations was wrong

Two updates, and the second is a correction to something I told you this morning.

**Everything is live.** A build went out at 16:00 today carrying this morning's two fixes — the
chopped-off-image case and the detector that had never been connected. So the answer to "do we
need another build?" is **no, it already happened**, on a roll another team was running anyway.
Nothing is waiting on a build.

What is still waiting is simply time: the pages have not rebuilt themselves since that roll yet,
so the newest fix is running but has not had a chance to show itself. That is hours, not days,
and it needs no decision from you.

**The correction.** This morning I told you I could not tell which version was running from the
version tag, and had to prove it another way. That reason was wrong. I compared a timestamp from
one tool against a timestamp from another without noticing they were in **different time zones**
— British summer time against UTC — so a commit that was actually half an hour *before* the
software started looked like it came half an hour *after*. The version tag would have told me
straight away.

The answer I gave you was still right — the fix is live, and I had proved that a second,
independent way, which is why the mistake did not reach your pages. But I published the wrong
reasoning, and if I had left it there it would have taught the next person to distrust a check
that works perfectly well. It is corrected everywhere I wrote it, and another team spotted it,
having made the opposite version of the same mistake themselves.

**So the two things still needing a person are unchanged**, and both are yours to call: applying
the small database change for the script security issue, with someone then looking at a news
page in a browser; and whether to prioritise the ESPN-menu problem, which is filed and which
nobody has picked up yet.

## 2026-09-04, 16:30 — the database change is applied, and one thing about it you should know

I applied the small database change (the script security fix). It went in cleanly at 16:28 and
I checked the result immediately: both components now build the page properly instead of
gluing text together, and both handle links safely.

Before applying I re-checked the things I had assumed yesterday rather than trusting them —
that there are exactly two components and no per-site copies of them, which would have meant
patching one site and silently leaving nine broken. Still true. I also rehearsed the whole
change against the live database and rolled it back, twice, before doing it for real.

**The one thing you should know, because it changes what the "hold for a browser check" was
protecting.** I had held this change back so a person could look at a news page before it
reached customers. Applying it is what you asked for, and it was the right call — but applying
and publishing are not the same step here. The updated script does **not** go out immediately:
each site picks it up the next time that site rebuilds its news page, which happens on its own
schedule, a few times a day. Nothing had rebuilt as of 16:31.

So the browser check has changed character. It is no longer a gate *before* customers see it;
it is now a check *after*, with a one-command undo ready if it looks wrong. I want to be
straight that this is a slightly weaker position than the one I described yesterday, and it
follows from applying the change rather than from anything going wrong.

**What I checked instead, and what it does and does not prove.** I ran a structural check over
both scripts as they now sit in the database: brackets and braces all balanced, both ending
correctly, and — the one that would have been nasty — neither contains the marker that would
have cut the script in half while still appearing to save successfully. That rules out the
clumsy failures. It does **not** rule out a subtle one like a missing comma. Only a browser
can, and this machine has none; the browser service we have needs the updated script to be
published first, which brings us back to the same wait.

**What to watch for, in one sentence:** if a news page comes up with an empty list where the
articles should be, that is this change, and the undo is ready and takes seconds.
