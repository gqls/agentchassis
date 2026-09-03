# Where we are — bugs_open/384 (plain prose, append-only, newest at the bottom)

## 2026-08-24 evening — picking the bug up

The bug: when an article gets its little listing picture (the "card"), the page that shows the list of articles is never told to go and look again. The list is a snapshot taken when the page was last built. A routine re-render of that page just re-ships the snapshot, so the card stays missing until something unrelated happens to rebuild the list properly. The dartsonline home page showed 4 of 12 cards without pictures for two days; the person who found it fixed that one page by hand this evening, so the site looks right now — but the mechanism is unchanged and the next card will do the same thing.

What I found on top of the filing: the same snapshot problem covers more than cards. The list picture falls back to the article's big hero image when there is no card, and a hero landing does not tell the list either. And the "tool" strips on several sites hold 14 stale entries right now — nobody sees it only because that strip doesn't display pictures.

The fix I'm building is one shared piece: "the data behind an article/tool list changed on this site — go and re-resolve every page that shows such a list". The card-maker and the hero-landing step both call it. A nightly-style check will also compare each list against a fresh answer and file a re-render if they differ, so any future producer we forget about is still caught. Pages marked as owned by a tool are skipped, because re-rendering those fails on purpose.

I've told the four other sessions whose files sit next to mine what I'm doing. The person who filed the bug is still online but working on something else; I've written into the bug file so they see it.

## 2026-08-24 late evening — the fix is written and tested; waiting on the review and the next build

The shared piece is built. When a card is made for an article, or a page's big picture lands and no card is coming, the code now finds every page on that site that shows a list of articles or tools and files a "re-resolve your list" request for each one. Those requests go through the same re-render machinery the news feed already uses, so nothing new has to be learned by the system — it just gets told, which it never was before. Pages that are owned by a tool are skipped because re-rendering those is refused on purpose and would just produce failures.

I proved the tests bite: I deliberately broke the code five different ways (declared a wrong source, un-declared a right one, removed the owned-page rule, deleted each of the two calls) and every time the right test went red. Then I put it back and everything is green.

Four other sessions checked my numbers while I worked, and three of the numbers I had written down were wrong or half-wrong — two of them in ways the other sessions themselves got wrong first. All corrected before anything was committed, and each mistake is written up with the check that would have caught it. The plain lesson: a number handed to you by another session is still just a number in a document; run it yourself before it goes in with a date on it.

Next: the reviewer council is looking at the change (budget half an hour), then it's committed so the next chassis build carries it. It does nothing until that build rolls. After the roll I'll prove it on a real card landing rather than trust the tests. The "safety net" sweep that would catch any future producer we forget is Phase 2, not yet built.

## 2026-08-24, later — reviewed, approved, and the safety net built too

The reviewer council approved the main fix (13 reviewers, none unreadable). Two of their advisory comments were worth acting on the same evening rather than filing:

- One reviewer asked how many pages a single card landing could send for re-render. I measured it: on one site, 26 — of which only ONE actually shows the picture. The other 25 are "tool strip" components that store the picture but never display it. So the shared lookup now only counts pages whose template actually renders the image (that brings every site down to 0–3), and there is a hard cap of 24 per landing that shouts in the log if it ever bites.
- Another reviewer said this "who consumes this data" lookup is the first general one in the platform and should be named as shared infrastructure rather than discovered later. I wrote a short RFC (052) stating the seam and the open question — whether the two older producers that hard-code their one consumer page should move onto it.

The safety-net sweep is also built now: a discovery check that compares each stored list against a fresh answer and files a re-render when a picture differs. It deliberately never "closes" anyone else's re-render request — I measured that no producer has ever retracted a re-render on this item type across 18,360 of them, so this would have been the first, judging 121 other producers' requests. Turning the sweep on is a held migration (603), to be applied by hand only after the build that carries the check has rolled.

Everything is committed (three commits) and the second review round is running. Nothing is live until the next chassis build rolls; after that, the proof is an induced card landing producing exactly the expected re-render requests — written up in the lane RUNBOOK.

## 2026-08-24, end of session — both reviews approved; nothing live until the next build

The safety-net sweep went through three review rounds tonight. Each round found something real: first, a place where an unreadable stored record could have been mistaken for "nothing to compare" (fixed and tested with a case that actually proves it); then two numbers I had written from memory instead of measuring — the migration now measures its own starting point when it is applied, and one status value I described as live does not in fact occur. Third round approved.

Where this leaves things: five commits on the branch, all reviewed, the build proven from a clean copy. None of it does anything until the next chassis build rolls. After that roll there are four things to do, written up in the lane's RUNBOOK: prove the roll, run the induced card-landing test, switch the sweep on by hand (migration 603 — it checks its own preconditions and tells you what to watch), and re-read the escalation number a week later.

Two open questions for you, neither urgent: RFC 052 asks whether the two older "tell my one listing page" mechanisms should move onto the new shared lookup; and the `rebuild_blog_listing` step still writes empty images into blog listings — harmless today (no blog index lists a page with a card) and now caught by the sweep when it fires.

---

## 2026-08-25 (afternoon) — you decided all four, and all four are done

You answered the four open questions: turn the sweep on, generalise the lookup now, fix the
action, and fix the tool-cta entries by changing the template. Then, when I measured what the
template change would actually look like, you made a second call: derive the missing images
first. Here is where each one landed.

**1. The sweep is on.** I applied the held migration by hand after re-checking that the running
software actually knows about the check — it does, on 301 pods. The sweep has run twice since
and, exactly as predicted, found nothing to fix. That is the expected result: it is insurance
against a future producer nobody wires up, not a repair of anything broken today.

One caveat I want to be straight about. To know the sweep is *working* rather than *blind*, I
need to see it report "I looked at N listings and all N were current". So far it has only
visited a site with an empty listing, where the honest answer is "nothing to compare". So the
proof is still outstanding — not failed, just not yet obtained. It should arrive as the rotation
reaches a busier site.

**2. The lookup is generalised.** This was the piece of shared machinery that answers "who is
using this data?". It could only answer that question about images. It can now answer it about
any kind of data — news items, directory entries, products — and the two older pieces of code
that had their own private answer now ask this one instead. Reviewed and approved first time.

Two things worth telling you. First, I found that a claim we had written down about those two
older pieces was **wrong**: we had recorded that each one had a single page name hard-coded into
it, and neither did. They were less broken than we had said. I have corrected that in the
document where it was written, because it had already been quoted onward into an architecture
proposal. Second, the test I wrote to keep this honest immediately caught an error in my own
work — and then turned out to be looking through a keyhole, reporting something correct as
wrong. I fixed the keyhole rather than excusing the case, because a test with a blind spot is
worse than no test.

**3. The action is fixed — and the review caught something real.** The blog listing rebuild was
blanking out the images. The reviewers sent it back once, and they were right to. My own code
comment described a way the thing could fail silently — if the data shape ever shifted, the page
would quietly keep its old contents and report success — and I had written that down as a known
risk without actually closing it. It is closed now: that case fails loudly instead.

They also asked why a query had no limit on it. Checking, I found something better than the
question: the same listing has two pieces of code that build it, one of them capped at 24 items
and the other uncapped. On one site with 40 articles they would have disagreed by 16. Both now
share one setting.

Two of the reviewers' own checks disagreed with numbers I had put in the submission. I re-ran
both rather than argue. One of them was right and I was too narrow — I said "47 images are
blank", meaning the pages this code writes; across the whole estate it is 55. The extra 8 are on
pages this code never touches, and I checked each one: they are blank because those articles
genuinely have no picture, which is correct. The other check was mine to stand behind, and I
have the row that proves it.

**4. The tool strips now show pictures.** This is the visible one. When you said "change the
template", I measured what would actually appear and found a problem worth putting back to you:
144 of the 228 pictures would have been full-width page banners squeezed into small cards — all
on loancalculator, whose tool pages had no proper thumbnail at all. You said derive the missing
ones first, so I did.

That worked for loancalculator — all ten pages now have a proper cropped thumbnail, and the
banner problem is gone entirely. It did **not** work for loanandmortgagecalculator, and the
reason is worth knowing: the thing that makes these thumbnails **crops an existing picture**. It
does not draw a new one. Those 19 pages have no picture of any kind to crop, so the job
completed honestly having produced nothing. Their strips will simply show text, which is the
designed fallback and looks fine.

Everything else is live and I have checked it on the actual pages, not just in the database: six
pages re-rendered so far, every one showing real thumbnails, and none showing an empty picture
box. The rest are working through the queue.

**One page refused to re-render, and I want to flag it rather than bury it.** A page on
ai-agent-orchestration was stopped by a safety guard: re-rendering it would have thrown away
half its layout, so the system refused and wrote nothing. That is the guard doing its job. It is
**not** caused by anything I changed — that same page already failed three times yesterday, and
those were the only other refusals anywhere in a fortnight. Something about that one page's
stored version cannot currently be rebuilt from its own source. It is a real problem, it is
someone else's to pick up, and nothing is damaged in the meantime.

**Something I bumped into that is not ours.** A test the house rules tell every author to run
before committing cannot be run at all — it refers to a name that was renamed two days ago and
never updated, so it fails to compile. It is another team's file and I have not touched it, but
it means that particular check has been silently unavailable to everyone since 23 August.

## 31 August — I picked this back up after five days, and I am NOT closing it

The last note here said the listing bug was fixed, proven four times, and ready to be closed. I
came back to do exactly that, re-checked everything first, and found a page still showing the
fault. So it stays open.

**What the bug is, in one line.** When a page gets its little thumbnail picture, the index page
that lists it is supposed to notice and start showing that picture. It wasn't noticing.

**What I found today.** On leopardessconsulting.co.uk, the blog index is still showing two
entries as plain text with no picture — and they are the first two in the list, so it is the most
visible spot on the page. The pictures for them were created on 27 August. It is 31 August. I
checked this on the actual live web page, not in the database, because the database is where I
would only be reading my own homework back to me.

**The part we built works.** The moment those two pictures arrived, our new code noticed within
about four hundredths of a second and correctly asked for the page to be refreshed — nine times
over, each request correctly addressed. So the detecting half is doing its job.

**The part that failed is the doing.** Two of those refresh requests were picked up and reported
success. They even published a new copy of the page. But the list of entries they were supposed to
rebuild was never actually rebuilt — I can see it was last touched an hour *before* the pictures
arrived. So the system told itself it had done the work, and had not. That is the shape of thing I
care most about, because it is invisible: everything reads green.

**I do not yet know why, and I have deliberately not guessed.** The detailed record of what those
runs did is only kept for about a day, so the evidence from 27 August is gone. I can rule out the
obvious suspect — I checked the live setting that decides which mode a refresh runs in, and it is
correct. Working out the rest properly needs one diagnosis run, which I have not fired yet.

**A large caveat I nearly missed, and a correction to myself.** The whole fleet was very quiet for
the last few days, and I first wrote that up as an unexplained problem, guessing it matched a bug
another team had filed. That guess was wrong and I have corrected it in the record. The truth is
simpler: we ran out of API credit. Another team had already measured it precisely — everything was
refusing from 28 August until about ten to nine this morning, when it recovered. No picture has
been produced anywhere for two days because of it.

**Why that matters for the decision.** Our fix looks healthy today partly because nothing has been
asking it to do anything. A test that passes because nobody rang the doorbell is not a test. That,
plus the page above, is why closing this today would have been the wrong call.

**The other thing, unchanged.** There are still fourteen missing pictures across three pages that
you own and edit yourself. Our fix structurally cannot reach those — it is not a bug in it, it is a
gap next to it, and it needs its own small piece of work. That has not moved.

**Where it stands.** Nothing is broken by any of this, nothing is in flight, and nothing needs a
decision from you this minute. What is owed is one diagnosis run on that single 27 August refresh,
and then a proper re-check now that the credit is back and the machines are working again.

## 2 September — I re-read the page as you asked, and it turned up a second, worse problem

You asked me to look at the page again before spending anything on a diagnosis run. That was the
right order, and it changed what I would spend it on.

**First, the page.** It is unchanged. Not "roughly the same" — the file I fetched today is
byte-for-byte identical to the one I fetched on Sunday, down to the last character. The same two
entries are still showing as plain text with no picture, still sitting at the top of the list. The
pictures for them were made on 27 August. That is six days.

**And this time there is no excuse available.** On Sunday I had to caveat everything, because the
whole system had been out of credit and nothing was running. It is running now, properly — about
nine in ten refresh jobs completed today, and pictures are being produced again. So the page had
every opportunity to fix itself and did not.

**Now the part I got wrong on Sunday, and the bigger thing it was hiding.** I told you the
detecting half of our fix "fired nine times". It did not. There were two different things filing
those requests and I lumped them together. Our card-landing detector filed six, and all six were
picked up and ran. The other nine came from the *sweep* — the safety net we added on 25 August to
catch anything the first detector misses.

**The safety net has never worked. Not once, since the day we turned it on.** Twelve requests in
its entire life, across three different customer sites, and every single one was dead on arrival —
filed into a state the system treats as already-closed, so nothing ever picks them up. It has been
correctly spotting this exact fault every day for five days and every one of those reports was
binned automatically, in silence.

**Why, and it is uncomfortable, because we caused it.** The system has a sensible rule: if a repair
has already been attempted twice for the same thing and the fault is still there, stop retrying and
park it for a human. The rule counts *completed* jobs as attempts. Our first detector's jobs keep
completing — they just are not actually fixing anything. So the safety net gets switched off by the
first detector's false successes. **Our own fix disabled our own backup, on precisely the pages
where the backup was needed.**

**One thing that has been quietly reassuring us and should not have been.** We set ourselves a
check for 1 September: look at how often the sweep escalates, expecting roughly one in thirty-six.
It has been reporting zero. Zero looks like good news. It means the sweep never ran at all. I have
marked that as worthless rather than leaving it to be read again.

**What I have done, and deliberately not done.** This turns out to be a known problem — another
team filed it on 25 August and their write-up describes this exact chain. So I have not opened a
competing case or touched their code. I have added our evidence to their file, including one thing
they did not have: in our case the "failed attempts" were all *successes*, and they came from a
different component than the one that got parked.

**Where that leaves the decision.** The original question — why does a refresh job complete without
doing anything — is still open and still unanswered, and it is now the only thing a diagnosis run
needs to look at. Nothing here is urgent or damaging: three pages on three sites are showing a
missing thumbnail. But this bug cannot be closed, and I would rather spend the run on the narrow
question now that everything around it has been cleared away.

## 2 September, later — the diagnosis run came back "can't confirm", and that turned out to be the useful answer

You said go ahead, so I fired it. It ran five rounds and finished in about a quarter of an hour.
Its verdict was **not confirmed** — it would not sign off on my explanation.

**It was right not to, and it caught me out.** My whole argument rested on one timestamp: the
listing's "last changed" time still read an hour before the pictures arrived, so I said the refresh
jobs had changed nothing. The run pointed out that this timestamp does not actually prove what I
was using it to prove — it cannot tell "never touched" apart from "touched, wrote the same thing
back". I had checked afterwards, and it turns out nothing in the database maintains that column
automatically. I had treated a habit as a guarantee.

**My conclusion did survive**, because there is a proper history table that records genuine changes,
and it has no entry for either job. So the jobs really did write nothing. But I want to be plain
that I was right by luck rather than by evidence, and I have written that up where we log this sort
of thing.

**Chasing the run's objection then found the real thing, and it is bigger.** That history table
records *which part of the system* made each change. Every change to this listing over the past
fortnight was made by one component — a blog-listing rebuilder. Across the whole estate, in
fourteen days, the machinery our fix actually triggers has never once rewritten a listing. Not
rarely. Zero times. And when I listed out the steps that our refresh job runs, there is indeed no
rebuild step in it — it re-renders the page and publishes it, which is why it reports success and
produces a real published version, while the underlying list is untouched.

**Which raises an uncomfortable question about our own evidence.** The report from 26 August said
the fix had "proven itself four times". I have now checked the first of those four against the
history table: the repair was made by the blog-listing rebuilder, not by our fix. Our fix files a
request of a different kind entirely. I have not yet checked the other three, and I have written
down clearly that they need checking rather than quietly assuming the worst — but if they match,
then the evidence we closed the book on was crediting our work with somebody else's repairs.

**So the question has changed shape, and it is a better question.** It is no longer "why did the
refresh skip this one section". It is "does that refresh path rebuild these lists at all, or did we
aim the fix at the wrong mechanism from the start". That is answerable, and cheaply — the run told
me exactly what it was missing, and both gaps were in how I set it up rather than anything
mysterious.

**Nothing here is on fire.** It is still three pages on three sites with a missing thumbnail. But
this bug is now further from closing than when I picked it up on Sunday, and I would rather tell you
that plainly than let the earlier "ready to close" stand.

## 2 September, end of the day — I was too quick to doubt our own work, and the real fault is now pinned down

Two corrections to what I told you an hour ago, and then a genuinely useful result.

**First, I owe our earlier work an apology.** I said the fix's four demonstrations might all be
crediting somebody else's repairs. I have now checked them properly and that was wrong: three of
them are real, and so is a fourth on another site that repaired slightly later than we thought. Only
one of the five — the leopardess one — was another component's doing. The fix works.

**Second, I had a bad measurement in between.** I ran a count that appeared to show our fix's
machinery had never rewritten a single listing anywhere, and I wrote that into the case file. It was
an artefact of how I joined two tables: one part of the system replaces rows rather than editing
them, and my query silently ignored 98% of the history as a result. What survived my filter was
exactly the component that edits in place, so the answer looked clean and pointed at the wrong
culprit. I caught it because a different check came back suspiciously empty, and I retracted it
about twenty minutes later.

**Now the result, which is worth the day.** The fault is specific to **blog listings**. On a tools
page or an archetypes page, our fix works: the refresh writes the list and the pictures appear. But
a blog listing's list is maintained by one dedicated component, and the refresh job we trigger does
not include that component in its steps. So on a blog page the job renders the page, publishes a
real new version, and never touches the list. That is why exactly one page in the whole estate is
stuck, and why every summary number looked fine.

**One question left, and it is a narrow one.** That dedicated blog-listing component last ran an
hour before the pictures arrived, and has not run since — six days. Something either triggers it on
a schedule that has stopped, or triggers it on an event we are not sending. Find that and this is
finished.

**Where we are.** The bug stays open, but it is no longer mysterious: we know which pages are
affected, why, and what the last question is. Separately there is still the safety-net problem from
this morning — that one belongs to another team's case and I have handed them the evidence.

## 2 September, evening — I asked the dispatch team, they answered, and they also caught me out

**Why the blog list stopped being rebuilt.** The component that rebuilds it lives inside one
particular job, and that job is only run for a site when the site has work waiting that qualifies.
Leopardess's work does not qualify any more: fifteen of its last eighteen such jobs were marked, at
the moment they were created, as "already tried twice, don't bother". They were never attempted.
The last one that did run was on 27 August at 21:33 — one minute before the last time the list was
rebuilt. The site simply dropped out of service that minute and has not been back.

**The reason is the one from this morning, biting a second time.** The system stops retrying
something after two previous attempts, which is sensible — except it counts *successful* attempts.
Our fix had been working repeatedly on that site in the days before, and each success counted
against it. So the site's own good run is what switched it off.

**I asked the team that owns dispatch, and it was worth doing.** They identified the scheduler
involved, ruled out two things I suspected, and — most usefully — ran a measurement I could not:
daily completion counts for the last three weeks. Everything dipped during the credit outage, but
their sites recovered afterwards and ours did not. Dartsonline did 268 jobs on Monday, ours did 2.
That is a much cleaner piece of evidence than anything I had, and it points squarely at our site
having a specific problem rather than sharing a general one.

**They also told me I was wrong about something, and they were right.** I had reported a
fleet-wide problem: over 1,300 items apparently stuck behind a filter. It turns out those items
are not stuck at all — they are flags, deliberately recorded with no automated handler, and the
place I found them is where they are meant to live permanently. My mistake was that I built my test
from their verbal description of the filter instead of reading the actual code, and my test could
not tell "refused by the filter" apart from "never shown to the filter". I checked their correction
against the source myself before accepting it, and they were right. That claim is now retracted in
both places, and marked so nobody re-opens it.

**One thing genuinely worth watching, tomorrow evening.** The "already tried twice" mark expires
after seven days. Working through the dates, it should lift at about half past nine tomorrow night.
If the site starts working again on its own after that, our explanation is confirmed. If it is still
broken on Thursday, our explanation is wrong and someone should start again. I have written that
into the handoff as the first thing to do, with both outcomes spelled out, so it does not matter who
picks it up.

**Where we are.** The bug stays open, the cause is understood well enough to test, and the fix for
it belongs to another team's case rather than ours. Nothing is damaged meanwhile: three pages on
three sites are missing a thumbnail.

## 3 September — it fixed itself overnight, exactly as predicted, on a date I got wrong

**The page is right.** All thirteen entries on the leopardess blog now show their picture, including
the two that had been blank since 27 August. It happened at twenty past eleven on Tuesday night,
without anyone touching it.

**And it happened for the reason we thought.** The "already tried twice, don't bother" mark expires
after seven days. The moment it expired, new jobs were created *without* the mark, they were picked
up normally, and the component that rebuilds the list ran for the first time in six days. I can see
the unmarked jobs in the record. That is the explanation confirmed, not merely consistent.

**I got the date wrong by about a day, and it is worth saying why.** I predicted Wednesday evening.
It happened Tuesday night. I had worked out the expiry from the wrong job: I used the one attached
to the blog list I kept looking at, when the jobs that actually control whether the site gets
serviced are different ones, whose clocks started a day earlier. Right mechanism, wrong stopwatch.
Worse, I had written into the handover that if nothing happened by Thursday the whole explanation
was wrong — which would have thrown away a correct answer. The thing that saved it was that I had
also written down a second, better test: look at whether the new jobs carry the mark. They did not.
That test cannot be wrong about clocks, and it is the one that settled it.

**What is left, and the one thing I need you to decide.**

Our fix does what it was built to do: when a picture arrives, it tells the listings to refresh, and
on most page types that refresh rebuilds the list correctly. **But blog listings are rebuilt by a
different component, and the refresh we trigger does not include it.** So a blog listing does not
get repaired by our fix — it gets repaired later, when the site's routine maintenance next runs.
That is what happened here, and it took six days only because the site had been frozen out of
maintenance by the separate problem.

So: **do we close this bug on "blog listings recover on their own within a normal maintenance cycle",
or do we keep it open and make our fix repair them directly?**

I lean towards closing it, with one condition: that we actually measure how long that normal cycle
takes, so "recovers on its own" is a number rather than an assumption. The alternative means
changing a shared piece of machinery and putting it through review, to fix a delay we have not yet
measured. I would rather measure first. **But it is your call, and the case for the other option is
real — a customer looking at a blog page during that window sees a gap.**

Two smaller things stay open regardless: the fourteen missing pictures on pages you own and edit
yourself, which our fix structurally cannot reach and which need their own small piece of work; and
our safety net, which still has never run because of the other team's problem, and will need
re-checking once they fix it.

## 3 September, later — you said keep it open, and the first check proved you right

I started working through the remaining items and the first one found the fault happening live, on a
different site, right now.

**designblog.co.uk.** Four guide pages got their pictures early this morning. The listing that shows
them refreshed twice afterwards — at six minutes past five and again at twenty-five past — and both
times the pictures came out blank. All four pictures exist, are active, and are correctly attached
to their pages. I checked every input the system needs and they are all correct.

**And this matters more than the original case, for two reasons.** First, this is not a blog listing
— it is the ordinary kind, on the path I told you yesterday was working. So my claim that the
problem was confined to blog listings was wrong. Second, because it happened four hours ago rather
than a week ago, the system's own record of what those refresh jobs did still exists, and it says
plainly that the section *was* rebuilt, not skipped. That kills the explanation I had been carrying
for two days.

**So I have stopped guessing and sent it to the diagnosis loop**, with the setup corrected for the
two mistakes I made last time. That is running now.

**One more correction I owe you.** Yesterday I said four of the five earlier successes were genuine.
What I actually checked was that a write happened — not that the write produced pictures. Those are
different things, and it is the same distinction that has caught me out repeatedly on this bug: a
job reporting success is not a repair, and a write is not necessarily a correct write. I have marked
that claim as needing re-checking rather than leaving it standing.

**Where this leaves us.** Keeping it open was the right call and I would not now argue for closing.
The fault is live and reproducible, which is the best position we have been in for diagnosing it —
far better than the six-day-old case I started with.
