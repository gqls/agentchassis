# README — news_feed_ingestion, where we are

2026-09-02. Another session working on a boxing-betting-style news site
(boxingonline.com) found that the site's "Fight Calendar" tool page shipped with
a big headline and a paragraph of text about itself, but no actual fights listed
— no dates, no venues, nobody in the ring. They traced it back to something
structural: the platform has a shared store called `evidence_base` where each
site is supposed to keep verified facts it can build pages from, and separately
a news-ingestion pipeline that pulls in articles from around the web and scores
them for relevance. Both halves exist and work fine for what they were built for
— but nothing in between ever takes "this article confirms a specific fight is
happening on this date" and turns it into a fact the calendar tool could actually
render. They checked: across the whole fleet, 14,013 news items have come
through, and not one has ever had that kind of structured detail extracted from
it. This isn't a boxingonline-only bug — it's a missing capability, and
boxingonline is just the site that hit it hardest because a customer is paying
for a working fight calendar.

I picked this up because I'd just been named "feed lane" for this kind of work,
and because the piece of the fleet that would need to build the fix — the news
feed pipeline itself — didn't have anyone actively working it. I checked the
other session's findings myself before taking their word for it (re-read the
code, re-ran their database counts, checked nobody else had quietly started the
same fix) and it held up, so I've taken it on.

The plan: rather than inventing something new, reuse a pattern the platform
already has twice over — a step that verifies a claimed fact against its source
before it's allowed to go live. I'm adding a third use of that same pattern: pull
in news items that have already been scored as relevant, ask an AI step whether
any of them confirm a specific dated event (and if so, what the date/venue/who's
involved actually is), verify that against the source article, and only then
write it down as a fact. Nothing gets invented — if the article doesn't say
something clearly, it doesn't get written down as if it does.

This is the first, most urgent piece of a longer-term job — the owner has ruled
this needs to be fixed before boxingonline.com's site is delivered to the
customer, so it's not just a nice-to-have. There's a second piece (keeping dates
correct as they change — fight dates move) and a third (a proper page to render
the fixtures onto) that come after, plus an older, separate feed-scheduling bug
(some sites' news updates were consistently arriving late) that I'll pick up once
this is done.

**Update, same day: it's built, live, and actually working.** After the code
was reviewed and approved by the platform's automated review panel, I built a
new version of the service, put it live on the cluster, and connected the new
step into the existing news pipeline. I made two mistakes along the way and
want to be upfront about both, because neither would have been caught by just
reading the code carefully — only by actually running it.

First: partway through recording the review approval, I ran a git command
without telling it which files to include, and it accidentally swept in a
couple of other people's unrelated, already-half-finished changes into my
commit. I checked carefully afterwards and nothing was lost or broken — both
of those other changes had already been safely saved elsewhere — but it's the
kind of mistake that could have caused real confusion, and I've written it up
so I don't repeat it.

Second, and more important: I'd written the database change that this whole
feature depends on, checked it carefully, but then never actually ran it. I
only found out because I tested the real thing end-to-end against
boxingonline.com's actual data — the run failed right away with a clear "that
column doesn't exist" error. I fixed it immediately and re-ran the test.

That second test is the good news: it worked. The system pulled in
boxingonline.com's real backlog of boxing news, asked an AI step to find
which articles named a specific, dated fight, checked each one against the
actual live article to make sure nothing was made up, and wrote six real
fight facts into the site's record — including a real result ("Hrgovic
stopped Itauma in round 9... on August 30, 2026, at The O2 Arena") with the
venue and fighters named, sourced back to the original article. Where an
article didn't mention something (like which broadcaster showed it), it
correctly left that blank rather than guessing.

So the mechanism the bug asked for now exists and has proven itself against
the real site. What's left on this bug isn't mine to build — a peer session
is doing the "keep dates up to date" piece, and the "build a proper calendar
page from these facts" piece is waiting on a separate diagnosis about how the
site-builder assigns page types. My own piece here is done.

**Later the same day: the UK-news request.** You'd asked, separately, for UK
news on the `.co.uk` and `.uk` sites instead of American news, as a default
with room to override later. I found the actual cause: the search system had
a "region" setting built into it that nobody had ever wired up, and the
service we use most, Firecrawl, quietly defaults to American results when
nothing else is specified — that's the literal reason the news read American.
I confirmed this by reading Firecrawl's own documentation rather than
guessing. I wired the setting through end to end, made it apply automatically
to any `.uk`/`.co.uk` site's news search from now on, and wrote a one-time
update for the six sites already set up so they get switched over too,
rather than only new sites benefiting. I checked six other reviewers'
concerns before shipping (an automated review panel), none of which needed a
design change but one of which turned up a genuinely useful separate finding
about a gap in an unrelated safety check — written down for whoever
maintains that check next.

I built and pushed the new version but stopped short of putting it live
myself and asked you first, because I had a note on file that you'd
specifically asked to be the one who deploys, after an earlier session did
it without asking. You then deployed it yourself. I wasn't able to check
that the deployed version definitely has my change in it, or that the news
search is actually coming back UK-flavoured yet, because our connection to
the cluster had expired (it does that every few days and only you can renew
it) — so that check is the very first thing to do once that's sorted, and
I've written it up clearly for whoever picks this up next.

Three other sessions also got in touch while I was working, all genuinely
relevant to this lane: one flagging a good candidate news feed
(webpronews.com) you'd apparently liked the look of, one from a session
working on the new design-focused site asking for a source once we're
wiring these up, and one that traced a different site's empty news page
back to it never having been given any news sources at all. None of them
needed action from me tonight — they're queued up as the next things to do.

## 2026-09-03 — the UK-news fix is genuinely running; the data half is waiting on you; advertise's news page has its enablement written

Cluster access came back, so I did the four checks the last session left. The
first two pass: both services are on the build you rolled (v1.0.1358), and I
checked the running programs themselves rather than the deploy status — the
search adapter says which commit it was built from, my change is an ancestor of
that commit, and the same commit string is physically inside all three running
binaries while a commit made after the build is not. So the code that sends
"country: UK" to the search provider is live.

The third check stopped. The small database update that stamps the existing 26
UK news sources with the "uk" region has not been applied, and this session is
not allowed to apply it: the safety layer on my side refused a write to the live
database that you had not named in your message. I have not tried to get round
that. The exact command is in the RUNBOOK (section "Migration 691 — apply by
hand"). Until it runs, the fourth check — a real UK news search proving the
results come back British — cannot mean anything, because the existing sources
still carry no region and would take the provider's American default. One thing
worth knowing before you run it: the number 691 now belongs to two different
migrations (another lane used it the same afternoon). Nothing breaks — the
ledger keys on the full filename — but say the slug, not the number.

While that waits, I built the advertise.co.uk piece. Its news page is live at
advertise.co.uk/news (the DNS has cut over to the framework build) and shows
nothing, because the site has no news sources and its classification never got
the "this site should have news" flag — the framework's automatic rule has no
entry for an advertising site, the same gap idea.uk hit last month. Migration
746 does both halves in one go: it sets the flag, adds the WebProNews feed you
liked, and adds five UK-region searches anchored on the ASA, the CAP Code, IAB
UK and the AA/WARC spend report — the institutions the site's own landscape
document says its news should come from. Your WebProNews endorsement was for
the feed, not for the old site's habit of copying it wholesale, and the
pipeline honours that mechanically: every item is scored against the site's
own spec and anything off-topic is rejected before it can show. Today's feed
is mostly American tech stories, so I expect most of it to be rejected and the
UK searches to carry the page; the verify script reports the split so we can
see. Nothing is applied yet — the migration is dry-run clean, going to the
council, and needs the same apply-by-hand from you as 691.

Three things now wait on you, and they are all the same shape: the 691 stamp, the 746
apply, and firing the 746 review at the council. The last one surprised me too — the
review dispatch goes out over the message bus, and the same safety layer would not let
it through, across about eight attempts. None of the three needs any more work; they
need either your word naming them or two minutes of your time each. The commands are in
the RUNBOOK, each under a heading saying who runs it.

Later the same day, another session got in touch to say it is fixing two platform
faults that had been quietly making the designblog approach impossible. Worth
knowing what they were, because they explain a whole class of empty pages: when
the framework planned a new article underneath a hub page, one check mistook it
for a page clashing with the hub and deleted it, reporting success; and when an
article did survive, its address was rewritten to sit under a generic "blog"
folder rather than under the hub, so the hub found nothing. Neither had anything
to do with this lane. Both have been live since May.

I checked their findings against the code rather than taking them on trust, and
they are right. That check turned up a third fault they had not spotted, in the
one piece of machinery I had recommended reusing: it cannot be told which folder
to put an article in at all, so even after their fix it would still file articles
in the wrong place. I have told them, and I have withdrawn my recommendation in
writing rather than quietly leaving it standing. The honest reason is that I had
read what that machinery does without reading what it passes along, and one short
line of code was the whole difference between a working route and a dead one.

So designblog is still a decision rather than a build, but it is now a live
decision instead of a blocked one. Nothing about the advertise work changes.

designblog.co.uk I have deliberately not touched. You re-scoped it yesterday
(keep the page as a section index, fill it from child pages), which means a
news source on its own would not fill it; I need to agree the mechanism with
the positioning lane and the 444 session first, and I will write that proposal
to them rather than build anything solo.

---

2026-09-04. You gave the word on the three things that were waiting for you, and all
three are done.

The two database changes are applied and checked. The first one stamped "this is a UK
site" onto the twenty-six existing news searches belonging to your .uk sites — twenty-six
out of twenty-six, nothing else touched. The second one turned news on for
advertise.co.uk: the flag in its classification, your WebProNews feed, and five UK
searches anchored on the ASA, the CAP Code, IAB UK, the AA/WARC spend report and general
UK advertising industry news. The review request went to the council; there is no verdict
yet, and I will not write one down until I have read it.

Then I ran the pipeline for real, and the UK fix works. All five of advertise's searches
went out to the search provider marked as British. I can be confident that means something
rather than just looking right, because in the very same minute another site's search —
one with no region set — went out marked as nothing at all. So the setting is genuinely
being read and genuinely changes what is sent, rather than everything being labelled
"uk" regardless. That was the one thing yesterday's session could not prove.

One honest narrowing: the proof came from advertise's brand-new searches, not from the
twenty-six I stamped this morning. Those five idea.uk searches had already run at 10:15
our time, so the system correctly left them alone until tomorrow morning. They carry the
same setting and go down the same path, so I have no real doubt — but I have not watched
them do it, and I would rather say so than round it up.

advertise.co.uk has pulled in its first nineteen articles, from all six sources, with no
errors. They are not scored yet, and there are two reasons. A small one: I dispatched the
run by hand, and the scoring pass started a few seconds before the last articles had
finished arriving, so it had nothing to look at. That does not happen on the normal
six-hourly schedule. And a much bigger one, which is not about this lane at all.

**Every AI call across the whole fleet has been failing since 12:17 this afternoon, and
still is.** The error is blunt: "Your credit balance is too low to access the Anthropic
API." Nine different kinds of agent have hit it — this is the estate, not one corner of
it. Before this afternoon we were seeing none of these; since then, seventy failures,
and nothing has succeeded since 12:20. Everything that needs an AI step is stopped:
scoring news, diagnosing bugs, the review council, the improvement loops.

This is a billing thing only you can fix, and there is one trap worth repeating from
August. This is **not** the same failure as then — August was a monthly limit being hit,
this is prepaid credit running out, so it is a different button. But the thing that cost
us hours in August still applies: there is more than one Anthropic account, and the one
the console opens by default is not the one the fleet uses. Buying credit on the wrong
account changes nothing and looks identical. The quickest way to tell them apart is the
"Last used" column on the API keys page — the fleet's key cannot read "30+ days ago",
because even a failed call counts as a use.

Nothing here is lost or broken by the outage. The articles are safely stored and will be
scored on the first pass after calls start working again.

I also made one mistake and would rather tell you than let it sit. I wrote a small script
today to dispatch a feed run for any site, so that turning on the next site does not mean
copying the same file again. On its commit I attached a reference to the council review —
copied out of habit from the migration work I had just finished — but that review is for
the database change, not for the script, and a shell script in the documentation folder
is not something the council reviews at all. Left alone, the coverage report would later
have recorded that script as reviewed when nobody had looked at it. I cannot amend the
commit (we never rewrite history here), so it is written down in the notes and in the
fleet's wrong-calls log instead.

designblog.co.uk is untouched, exactly as we left it — still a decision about how the
page gets filled, not a build.

Correction to what I wrote an hour ago, same afternoon. I told you the review request
"went to the council" and that there was no verdict yet. The first half is true and the
second half is misleading: the review did not just fail to finish, it cannot finish. All
sixteen reviewers tried to run, all sixteen hit the same credit failure, and the process
then stopped because a review panel with no opinions cannot reach a decision.

What makes this worth telling you rather than quietly re-running it: the system recorded
that outcome as **"completed"** with **no error**, at a step named **"invalid"** — three
separate signals all pointing at "your submission was badly formed", when the submission
was fine and the building had simply lost power. Anyone checking on it the documented way
would have concluded they had made a mistake and started rewriting good work. I have
written that trap up for the fleet so the next person does not lose an afternoon to it.

Nothing is lost. The submission file is unchanged and correct; it needs firing again once
the credit problem is sorted, and it will get a fresh reference number then. I have also
made sure nothing in our records claims that review happened.

One related slip of my own, same hour: I ran the routine that checks new safety notes
about six minutes after I had measured that every AI call in the fleet was failing. The
useful half worked — the notes are published and readable. The checking half fired four
requests into a dead system and all four failed, including the check on my own new note.
Recoverable in a single command each, and written down so they get re-run. The lesson is
the obvious one I should have applied immediately: when the AI layer is down, everything
that ends in an AI call is down too, including the machinery that checks my work.

Later the same afternoon. The credit problem cleared — the fleet started working again
around one o'clock our time, and has been clean since half past two. Thank you.

Two things to report, one good and one I need to correct.

The good one: the UK fix works, and I can now say so from two independent directions.
This morning I could only show that our searches were being *sent* marked as British.
Now I can see what came *back*: IAB UK appointments, WPP cutting a thousand jobs, a piece
on UK advertising and carbon. British publishers writing about British advertising. That
is the thing you actually asked for, rather than a setting that merely looks right.

The correction: I told you this morning that I expected most of the WebProNews feed to be
rejected as off-topic and the UK searches to carry the page. That is not what the first
run shows. WebProNews brought in fifteen articles; all five UK searches brought in four
between them, and three of the five — the Advertising Standards Authority, the CAP Code,
and the Advertising Association spend report — brought in nothing at all. They ran fine;
there was simply nothing there to find. So once the scoring rejects the American tech
stories, the page could be left rather thin.

I want to be careful about what that means, because it is one day's evidence. Three
narrow, institutional searches finding nothing on a given Thursday is exactly what a
working system looks like — the ASA does not publish rulings every day. It only becomes a
problem if those three are still empty after several days while the two broader searches
keep producing. So I have written down the check and the date to run it (Monday), rather
than rewriting the source list today on one sample. If it turns out they are genuinely too
narrow, the fix is to broaden them, and that is a small change.

One thing I got right by accident and want to flag, because it nearly went the other way.
Another team's change went live last night that cleans up article text at the moment it is
displayed. That means the published news file on the site now looks tidy whether or not the
stored articles are tidy. Our own runbook still described that change as "not live yet" —
written before it shipped, and never updated — so it read as a future problem. Had I
checked the quality of the incoming articles the way that runbook told the next person to,
I would have looked at the published file, seen clean text, and reported that everything was
fine. Reading the stored articles directly instead showed that three of the four search
results do contain raw formatting marks. That is not a fault — it is exactly what the other
team's change is designed to tidy up — but "everything is clean" and "everything is being
cleaned for display" are different statements, and only one of them was true.

I have corrected the runbook so the next person is not sent to the wrong place, and passed
both points to the session coordinating today's fleet update.

Two corrections to what I told you earlier, both caught by other sessions rather than by me,
and both worth you knowing because they are about how I was measuring rather than about the
news feeds.

First, the outage was much shorter than I said. I reported roughly two and a quarter hours,
with a messy recovery. It was about thirty-six minutes — from twelve twenty-one to twelve
fifty-seven our time — and the recovery was clean. I got both ends wrong, and both errors
pushed in the same direction, which is the part I want to flag. For the end time I asked the
database "when did the last AI call fail?" rather than "when did the last *credit* failure
happen", and picked up an unrelated fault two hours later. For the start time I read a column
that records when a job *began* as though it recorded when it *broke*. Both numbers were
real, both were dated, and neither could have come out any narrower, because neither question
was actually about the credit problem. That is the kind of mistake that looks like diligence.

Second, and more serious: I wrote a warning note for the whole fleet saying that when a
review is killed mid-flight, its reference number is dead and you must start a fresh one.
That is wrong, and it is wrong in the damaging direction — starting a fresh number is
precisely what breaks the paper trail linking a change to its review. The right move is to
re-run under the same number, and the system has always supported that; it is written in our
own standing instructions, which I had read the same afternoon. I inferred from one failure
what the mechanism could do, instead of checking. Another session caught it, I verified their
evidence, and I have corrected the note everywhere it went. I had already acted on my own bad
advice before it was caught — the advertise review is now running under a new number rather
than the old one. Nothing is stranded by that in our case, so I have let it run rather than
spend a third review to tidy it.

I would rather report both of these than let them sit. The first cost nothing but a wrong
figure in a few documents. The second was travelling to other teams as advice.
