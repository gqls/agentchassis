# webdesign.co.uk — where we are

Plain-prose running log, append-only, newest at the bottom.

---

**2026-07-25 — starting out**

You asked for your two sites, website-design.com and websitedesign.com, to be
merged into webdesign.co.uk: very close to the existing design, keeping every
feature except websitedesign.com's client-side LLM builder.

First thing worth saying: the two sites are very different sizes and very
different designs. website-design.com is the big one — 86 pages, 55 little
browser tools and 23 articles, all sharing one stylesheet, all in a stark
black-and-electric-blue Swiss style with no web fonts and no external
dependencies at all. websitedesign.com is much smaller — 21 pages — and only its
homepage actually wears the warm sage-and-terracotta look; every one of its 20
sub-pages is still in an older dark "terminal" skin that was never migrated.

You chose the warm minimalist look for the merged site. That is the right call
aesthetically but it is worth being clear about the consequence: it means the
*smaller* site's design gets applied to the *larger* site's content, so roughly
86 pages need reskinning rather than 20. The way we make that affordable is a
single compatibility stylesheet that re-points the old design's colour names at
the new colours, so we sweep literal colours out of the pages rather than
rewriting thousands of references by hand.

Second thing: you asked me to check the two sites don't have duplicate tools.
They don't. All 64 tool names are unique, and I read the actual code of every
pair that looked like it might overlap. The closest call was two shadow
generators — but one is a manual layer-by-layer editor and the other is a
parametric one where you set softness and distance and it generates the layers
for you. Genuinely different tools. Same story with the two SEO tools (one does
Article/Product/FAQ markup, the other does local-business markup wrapped in an AI
prompt) and the two text tools (one cleans AI writing up afterwards, one
constrains it beforehand). Nothing needs dropping. The tools index will carry a
one-line subtitle on each of those pairs so a visitor isn't confused either.

Third thing, and this is good news: webdesign.co.uk is already wired up. It is
registered, it is on Cloudflare, and the edge already knows to serve it from the
storage bucket — when you visit it right now you get a "no such file" message
*from our own storage worker*, which is exactly what an empty-but-correctly-wired
site looks like. So there is no domain or DNS work for you to do. The one thing I
can't check from here is whether the deploy robot's Cloudflare token can see this
particular zone; if it can't, the only symptom is that the cache doesn't clear
itself after a publish, and it's a one-line fix on your side. I'll know on the
first deploy.

While poking at the sources I found three things already broken, which we'll fix
on the way through rather than carry over: the "vibe equalizer" tool on
websitedesign.com has been dead since it shipped (it loads a file that doesn't
exist); one of its guides is a byte-for-byte duplicate of another with the wrong
title on it; and four finished tools and ten finished articles on
website-design.com are live but not linked from anything, so nobody can find
them. Those get linked in.

The approach for building it: rather than let the platform's planner invent a
site and then fight it, I'm going to walk it through onboarding one step at a
time with everything else held back, so I can read what each step decided before
letting the next one run. The important moment is pinning the colours before the
design agent ever runs — that agent invents its own palette if it isn't told
one, and then re-invents it on every subsequent run, which is how another site of
yours ended up with its colours changing four times in a day.

---

**2026-07-25 — first build session**

The site now exists in the system and most of the machinery is built. Where things stand:

**The content is converted.** All 95 of your pages have been transformed out of the two old
sites into the new warm design — 63 tools, 31 articles and the about page — plus two index
pages that are generated rather than hand-written. Nothing was lost: the tool refuses to
finish if any page in either source folder is neither converted nor explicitly listed as
dropped with a reason. The only things dropped are the AI builder you asked to skip, its two
guide pages (one of which turned out to be a byte-for-byte duplicate of the other with the
wrong title on it), and an unfinished template page that was never linked from anywhere and
still had "[Site Name]" in its title.

**Three things that were broken are now fixed.** The vibe equalizer tool works for the first
time since it shipped — it was loading a file that doesn't exist, so every slider was dead.
Its Copy button had never had any code behind it either. And four tools and ten articles that
were live but unreachable are now properly linked, because the new index pages are generated
from the actual list of pages rather than maintained by hand.

**A number I got wrong, twice.** The old about page claimed a "100 Lighthouse score" and a
"0.1 second load" — figures nobody has measured for this domain, so I replaced them with
things that are simply true (no servers, no frameworks, no accounts, and the tool count).
Then I typed the tool count as 64 when it is actually 63. Worse, that same 64 had come from
the brief I wrote for the system before I'd counted properly, and it had already spread into
eight of the site's internal description documents — including the one the page planner
reads. So the home page would have opened by advertising a tool that doesn't exist. All
corrected, and the count is now calculated rather than typed, so it can't drift again.

**On the design.** The system read your brief and set the colours correctly by itself, which
I wasn't expecting — all eight of them exactly as specified. I've pinned them anyway, because
the design agent invents its own palette whenever it isn't given one in a specific structured
form, and then re-invents it on every subsequent run.

**What's left.** The site's header, footer and page-frame still need building, then the 95
pages get loaded in and published. I've deliberately walked the setup through one step at a
time rather than letting it run, checking what each step decided before allowing the next —
which is how the wrong tool count got caught before it reached a page.

**One thing that may need you.** The deploy robot needs permission to clear Cloudflare's cache
for this domain. If it doesn't have it, the only symptom is that changes take a while to show
up. I'll know on the first publish and will tell you if it needs a click from you.

---

**2026-07-26 — it's live, and I was wrong about the hold-up**

The site is up. Every page works: the home page, the tools index, the learn index, the about
page, and all 94 tool and article pages underneath them. The search box works, the tools have
their JavaScript, the colours are yours.

**The thing I got wrong.** Yesterday I told you the last step was blocked by a platform bug
and that 98 "publish this page" jobs would never run. That was wrong, and it was wrong in a
way worth explaining because it's the kind of mistake that wastes a lot of somebody's time.

The queue was working the whole time. It simply takes about twenty minutes before the first
job is picked up, and then roughly two minutes per page — so about three and a half hours for
98 pages. I gave up eight minutes before the first one started.

Worse than the impatience was my reasoning. I said "look, the image-generation jobs are
finishing but the page jobs never are — so something specific to page publishing is broken."
They weren't finishing at the same time. The image jobs I was watching had been picked up
*before* the page jobs even existed; I was watching the tail end of earlier work. And the
moment those finished, image work stopped completely for three and a half hours and only
resumed eighteen seconds after the last page was published. Which is exactly what should have
happened, because I had just told the system to prioritise pages over images. **I mistook the
queue obeying my own instruction for the queue being broken.** One query comparing timestamps
would have shown me that. I looked at counts going up and down instead.

I've written that up properly in the project's shared log of wrong calls, because the entry
immediately above mine is someone making the opposite mistake — waiting an hour for something
that really had vanished. Same ambiguity, opposite conclusions, and in both cases the answer
was a timestamp nobody checked.

**The hero is fixed.** You were right to flag it. It was a full-width dark banner with a
photo behind it, which is the opposite of what you asked for and what your own design rules
forbid. It's now the two-column layout from your brief: words on the left, image on the right,
stacking on a phone.

Worth being precise about why it happened, because it isn't the failure it looks like. The
colour pin did its job perfectly — every colour on the site is one of yours. But pinning
colours doesn't control which *building blocks* the system picks, and the block it chose had
darkness painted into it directly rather than drawn from your palette. Right paint, wrong
furniture. I checked that the planner only built the one page I asked for, and didn't check
what that page was made of. That's on me and it's now noted as a step to always do.

**The JavaScript problem can't happen again.** The bug where every tool's code was being
silently thrown away is now caught by a check that refuses to build at all if any page loses
its scripts. I proved it works by deliberately reintroducing the original bug — it produced
60 errors, one for each tool that would have shipped broken. A safety check you've only ever
seen pass isn't tested; it's just quiet.

I've also filed that as a wider issue for the platform, because the same blind spot exists
elsewhere: nothing anywhere checks that a published page's JavaScript actually works. Every
check we have asks "does this exist" rather than "does this function", and broken JavaScript
looks completely normal until someone clicks something.

**Still open, and none of it urgent:** the tools that need hands-on testing in a real browser
(the ones that use your camera roll, canvas or clipboard) haven't been clicked through yet —
that's about sixteen pages. And the deploy robot's permission to clear Cloudflare's cache
still hasn't been confirmed; the only symptom would be changes taking a while to appear.

---

**2026-07-27 — answering the duplication question, and what Phase 2 needs from you**

You asked whether your two sites duplicate each other. **They barely touch, and I should
correct something I said yesterday** — I wrote that "the same content now sits on three
domains", which was loose. What's actually true: the two old sites don't overlap; the new one
overlaps both, because it's their union.

The evidence: not one shared article subject. website-design.com writes engineering and maths
deep dives — the physics of UI, the end of hex codes, ambient occlusion in CSS, why you can't
scrape Google. websitedesign.com writes exclusively about AI website builders — v0, Lovable,
Bolt, AI slop, the 70% wall. No file on either site is byte-identical to any file on the
other. All 63 tool names are unique. Five pairs of tools sound similar and I read the code of
each: they're genuinely different jobs. The only real duplicate was *inside* websitedesign.com
— one guide was a byte-for-byte copy of another with the wrong title on it, and that's gone.

There's a useful consequence in that. The merge gave you a genuinely complementary library:
craft and visual depth from one site, current AI-builder practice from the other. Your
"renewed focus on AI" isn't a new direction — it's half of what you already own, currently
buried by the ordering.

**The one thing I need from you before I can do what you asked.** You want the tools and
guides ordered by popularity. **We have no popularity data at all.** The site is one day old,
it's served from storage behind Cloudflare so there are no server logs, and there's no
analytics beacon on it. I could invent a plausible ordering and present it as popularity, but
that's precisely the mistake this project already made twice with the tool count, and I'd
rather not make it a third time.

What I'd suggest: let me add Cloudflare's free analytics now — it's about ten minutes and one
line in the page template — and meanwhile order things by explicit editorial judgement,
labelled as such, with a note to revisit once there's a month of real data. And a question:
do you have Google Search Console on either of the old domains? That would give real search
and click data for this exact content, which is the best proxy going and costs nothing to look
at.

The good news is that reordering is cheap by design. The index pages are generated from a list
rather than hand-written, so reordering the entire site is an edit to that list and a re-run —
not a rewrite. The same is true of adding tools and articles, which is what "adding rather
than removing" needs.

**Two more things I found rather than built.** You want a news section — one already exists in
the platform, with the components and a refresh job that runs every six hours. It's parked
behind a deliberate gate that says don't switch it on without your say-so, which you've now
given, so it's a configuration job rather than a build. And there's existing machinery for
curated directories that the UK third-party tools idea should probably use.

Everything's written up in a Phase 2 handoff so you can pick this up in a fresh chat. It opens
with the popularity question and lists four decisions I need from you — including one I think
matters more than it sounds: you've described two different audiences, designers who want a
tool and people who want web design *done*. Those want quite different things from a homepage.

---

**2026-07-27 (later) — analytics wired, news switched on, ordering deliberately deferred**

Three things done, and one small job left with you.

**Analytics.** The Cloudflare beacon is now in the site's page template, but *gated* — it
only renders if there's a token, and there isn't one yet. I did it that way because the token
can only be created in your Cloudflare dashboard, and putting a fake one in would have meant
every page on the site making a request that always fails. So nothing broken has shipped, and
the moment a token exists it starts working.

**You need to do one of two things, and only one.** Easiest by far: in Cloudflare, go to Web
Analytics, add webdesign.co.uk, and choose **Automatic Setup**. Because the site already runs
through Cloudflare, it will inject the tracking itself and there's nothing for me to deploy.
Alternatively, if you'd rather it lived in our own code, send me the token from the Manual
Setup box and I'll set it. Until one of those happens we're collecting nothing, so it's
genuinely the highest-value five minutes on this project right now.

**Ordering.** You made the point that settles it: we're about to rewrite the tools and guides
anyway, so any stats we gathered now would be measuring content that's about to change. So the
sequence is instrument → improve → *then* measure → *then* reorder. I've deliberately not put
any ordering in, rather than putting in a guess and calling it popularity. When the time
comes it's a quick job — the index pages are generated from a list, so reordering the whole
site is an edit to that list and a re-run.

**News is on.** I want to be clear about what I did and didn't do here, because there was a
trap. There's an existing plan in the system to switch on news across thirty-seven *other*
domains, which you'd previously paused. That is not what you asked for, so I didn't go near
it. What I did was add a news section to this one site: five search topics (UK web design, AI
web design tools, CSS and browsers, UK accessibility rules, and design trends), a news page,
and a News nav item.

Two honest caveats. The five topics are my editorial guess — nobody has seen what they return
yet, and I'd expect to change them once we do. And the news page isn't live yet: the feed
refreshes on a six-hourly cycle, and the page has to have actual articles in it before it gets
built, otherwise it publishes as an empty shell. I've deliberately *not* published the News
link in the menu yet for the same reason — a menu link to a page that doesn't exist would put
a broken link on all 98 pages. The order is: articles arrive, page builds, then the link goes
live.

All of it is written into the handoff so you can pick it up in a new chat.

---

## 27 July, later — the news feed would never have started, and finding out took reading rather than waiting

Picking this back up, the first instruction I'd left myself was "wait for the feed to run at
13:49". I nearly did exactly that. It would have been a wasted hour, because the feed was
never going to run — not at 13:49, not at any time.

Here's the mistake, and it's an instructive one. I'd created the five search topics, the news
page and the menu entry, and I checked all three afterwards and they were all correctly
there. What I hadn't checked was the one thing that actually decides whether a site's news
gets fetched — and it turns out that isn't the search topics at all. The system keeps a
short profile for each site describing what kind of site it is, and there's a flag in that
profile saying "this site should have news". The scheduled job only ever looks at sites
carrying that flag. Ours didn't have it. The profile for webdesign.co.uk was written back on
the 25th when the site was first classified, and news simply wasn't considered at the time.

So I had built the whole thing correctly and left it disconnected from the switch. My own
verification passed cleanly because it checked everything I had written, and the flag was the
one thing I hadn't written. That's the lesson worth keeping: I checked my work, not the thing
that consumes my work.

It's fixed now. I've added the flag, and I confirmed it by running the scheduled job's own
site-selection query, unchanged, and watching webdesign.co.uk come back in the results where
it previously didn't appear.

**One thing I had to be careful about**, because it would have quietly undone the editorial
decisions. When a site is flagged for news, the system offers to generate search topics for
it automatically from a keyword list. If I'd filled that keyword list in casually, it would
have created a second set of five topics alongside the ones I'd hand-written, with cruder
search terms — and the careful choices about UK focus and accessibility would have been
diluted by machine-generated near-duplicates sitting next to them. It refuses to create a
topic whose name already exists, so I set the keywords to exactly match the names of the five
I'd already written, which makes the automatic step do nothing at all. It's a small thing but
it's the sort of detail that silently degrades a site over months. Whoever touches that list
next needs to know that changing a single word in it summons a sixth topic.

**Something I've noticed but deliberately not acted on.** The scheduled job only takes five
sites per run, in alphabetical order, and adding us makes six. "webdesign.co.uk" sorts last of
the six, so if all six ever come due together, we're the one that gets skipped — every time,
silently. Right now it isn't happening, because the sites drift out of step with each other by
a few minutes and rarely come due at once. But that's luck rather than design. I've written it
down in the technical notes rather than fixing it, because it hasn't actually caused a failure
yet and the fix touches machinery shared by every site on the fleet. If our news ever goes
quiet for a day, that's the first place to look.

**Where that leaves us.** The feed is now genuinely armed and the next run is due within the
hour. Nothing else has changed: the news page still won't be built until real articles arrive,
and the News menu link still stays unpublished until the page exists. That order still matters
for the same reason as before.

**And the Cloudflare analytics step is still sitting with you** — it hasn't moved since this
morning, and it's still the highest-value five minutes on the project.

## 27 July, mid-afternoon — the feed fired, dispatched properly, and still brought back nothing

Mixed news, but the good part is solid and worth stating first: **the fix works.** At 13:49
the site was picked up by the news job for the first time ever, and all five of our search
topics were sent off correctly, with the exact search wording we'd chosen intact. None of that
could have happened this morning. The thing that was broken is fixed and I've verified it
rather than assumed it.

**What then went wrong is not ours.** Every one of the five jobs died about eight minutes
later, all with the same message — the job was handed off and the thing meant to pick it up
never answered, three times over.

My first thought was that I'd broken it, because I'd armed this feed less than an hour
earlier and that's the obvious suspect. The way I checked took about a minute and I'd
recommend it to anyone in the same position: **look at a site you haven't touched, in the same
run.** vetcomparison.uk went through the identical failure in the identical run. Two different
sites, two different people's work, one scheduled run, nothing fetched by either. So it isn't
my change — it's the platform underneath.

**And it's a known problem that somebody else is already deep into fixing.** There's an open
case about jobs hanging when they're handed off shortly after the system restarts. Our run was
unlucky in a very specific way: the platform restarted at 13:45:31 and our news job fired at
13:49:09 — three minutes and thirty-eight seconds later, inside the window where this is known
to happen. I've added our example to their case file, because a fresh dated instance is
genuinely useful to them, and I've deliberately not started a rival fix — that's their work
and they're several rounds into it.

**One thing I found while digging that's worth telling you**, because it affects how you should
read this. The news job is set to run every six hours, so four times a day. In practice it only
ever brings anything back **twice** a day. I checked three days of history across four sites
and articles only ever arrive in the 7am and 7pm runs — never the 1am or 1pm ones. The reason
is a small timing quirk: a site's next fetch time is set when the articles arrive, a few
minutes after the run starts, which means it comes due *just* after the following run and
misses it by a whisker. On one site I measured the miss at thirty-seven seconds.

That's mildly wasteful in general, but it explains our bad luck exactly: the 1:49pm run is
structurally the quiet one, so it was carrying only brand-new sites like ours and nothing
established — which is why one restart wiped out the entire run.

**Where we are now.** I've reset our five topics so they're due again at the 7:49pm run, which
is one of the two that historically works. Left alone they'd have drifted to nearly 2am. There
was nothing to fix on our side.

The order still stands and hasn't changed: articles arrive, *then* the news page gets built,
*then* the News link goes live in the menu. I haven't published that link and won't until the
page genuinely exists.

**Separately, you answered the three open questions and I've recorded them properly.** You
chose to serve both audiences with fully separate tracks — that's the more expensive answer and
I think it's the right one, because a single buyer section inside a tool library treats buyers
as an afterthought, which is the exact problem the brief is trying to fix. I checked, and there
is genuinely no buyer content to build on: all 63 tools and all 31 guides are written for
practitioners. So that track is entirely new writing, and I've drafted a plan for it.

Two things in that plan I'd like you to look at. First, I've suggested calling the section
**Hire** — it's short and plain, but it's your call since it sits in the menu. Second, and more
importantly: the most obvious page to write is "how to judge a quote", and it's the most
dangerous page we've ever planned, because the natural way to write it is with price figures we
simply don't have. We've shipped made-up numbers twice on this project. So my rule is that page
ships with **no figures at all** unless you have real pricing data from your own work — it can
still be genuinely useful by explaining what a quote should itemise and what a cheap one has
quietly left out.

**And the Cloudflare analytics step is still with you.** Still five minutes, still the
highest-value five minutes on the project, still nothing being collected until it's done.

## 27 July, later — you moved the buyer section up a league, and it changes what the site is for

Your redirection landed a few hours after I'd planned the buyer track, and it's a much better
brief than the one I was working to. Writing down what I took from it, and the places I've
pushed back or added.

**You were right about `Hire` and I was wrong to propose it.** It sets a price expectation
before anyone reads a word, and the expectation it sets is Upwork. Going with **Buying design**,
your phrase. I did consider "Commissioning" — it has the nice property that a £5,000 buyer
doesn't use that word, so it filters the audience automatically — but it reads a bit
institutional, and I think the writing itself should do the filtering rather than a stiff label.

**You were also right about the quote page, for a better reason than mine.** I'd objected
because we have no verified pricing data. Your objection is stronger: it points the whole
section at money instead of design, and at £500,000 the buyer isn't price-shopping — they're
trying not to choose wrong on something they'll be held responsible for. That reframing is what
made the rest of the plan fall into place.

**The thing I think is the real opportunity, and I'd like you to push back on it.** Everyone
selling into this market has a reason to shade the truth about AI. An agency wants it to mean
little, because day rates depend on that. An AI vendor wants it to mean everything. We run one
of these systems in production across about a thousand sites, and we are not bidding for the
buyer's project — so we're the only party in the conversation with nothing riding on the
answer.

That means the most valuable thing we can publish is not what our system does well. It's a
specific, evidenced account of **where it goes wrong** — because that buyer's daily experience
is being sold a story, and the one thing they can't get anywhere is someone showing them the
failure modes honestly.

**But that needs a decision from you, and it's a business one, not mine.** There's a scale:
talk about failure types in general, publish anonymised cases, or publish our own named
failures with the evidence. The last is by far the most credible and it's also the most
exposing — we'd be putting our own mistakes in public where competitors can quote them. I'd
want that decided once, deliberately, rather than drifting into it page by page. I've not
written a word at that level and won't until you say.

**On the tools, you asked me to think and here's where I got to.** Your instinct — simple,
strong impact, returned to repeatedly during a supplier search — is right, and I think it has
a sharper test attached. The question isn't "is this useful", it's **"does the same tool get
used at three different stages?"** A buyer goes: build the case internally, write the brief,
draw up a shortlist, sit through pitches, sign a contract, then watch delivery. A tool that
helps at one of those is a gimmick. One that helps at three is why they keep the tab open for
four months.

The strongest by that test is a **side-by-side site comparison**: put your site next to five
others and get objective results. They'd use it to build the board case, again to check whether
an agency's portfolio work actually stands up rather than just the case study about it, and
again afterwards to see whether what they paid for is measurably better. Three visits, three
different reasons.

But the one I'd build **first** is an **accessibility check framed as legal duty**, because we
already own it. The contrast checker, the touch-target rule, the focus-visibility work — those
tools exist on the site today for designers. The buyer version is the same measurement with a
different frame: not "your contrast is 3.9:1" but "these elements would fail an audit, and
here's what that means for you". Same engine, two audiences. It's close to free and it carries
weight in a boardroom.

**One rail I want to state plainly.** These tools run on addresses the buyer types in, and the
results go to the buyer. We should never publish rankings of named agencies. Quite apart from
the legal exposure of publishing quality judgements about real firms, it would destroy the
neutrality that makes any of this worth reading. A buyer privately checking an agency's work is
research; us publishing the same thing is a hit list. It also means the directory idea should
be understood as **tools, not suppliers** — pointing at software is uncontroversial, ranking
agencies is not.

**On verified figures** — agreed, and no figures no page still stands. The most promising
honest source is one nobody uses in this context: **Companies House**. Agency size and turnover
from filed accounts is public and checkable. I should flag that the related machinery we have
elsewhere can't currently record where a fact came from, so treat that as a lead rather than
something I can lean on yet. Public performance data on real websites is the other solid one.

**The website creation form** — noted and recorded, not started. I've written down your
architecture point specifically, because it's the bit that tends to get quietly eroded later:
separate cluster, separate database. I've also noted that one of the tools I'm proposing (a
brief builder — answer questions, get a usable brief out) is the same shape as that form, so if
we build it that way it may simply become its first step.

**One tension I'd rather name than paper over.** You said we need a lot of web design traffic
any way we can get it, and also that this pitches at the multi-million pound buyer. Those are
different games. My reading: the 94 existing pages are the traffic engine — they rank, they're
already built, they cost nothing more to keep. The buying section is the positioning layer and
doesn't need volume to earn its place. The real risk isn't cost, it's **register collision**: a
brand director landing on an article about the mathematics of CSS Grid concludes this is a
developer site and leaves. So the buying section needs its own clear front door, and the home
page has to serve both without confusing either.

Related, and it makes the rewrite more urgent than I'd said: because the two old sites stay
live, our imported pages duplicate them. So rewriting the copy isn't only a quality job, it's
how we stop duplicating ourselves. And the buying section is the only part of the site with no
duplication problem at all, because none of it exists anywhere yet. Your instinct to build it
serves the traffic goal directly.

**Last question, and it decides real work.** You said ultimately we won't focus on designers.
I've argued they should stay as the traffic engine, but with no investment beyond the rewrite.
I need that either confirmed or corrected, because it's the difference between the practitioner
half of the rewrite being a proper job or a holding action.

## 27 July, evening — you were right, and the reason it wasn't caught is worse than the bug

**Fixed and live.** Every link on the home page now works. I checked all of them against the
real site, not the database, and then swept all 99 published pages for the same problem.

**What was actually wrong.** Ten of the thirteen links on the home page were dead. The only
ones working were the three in the top menu — which is exactly why it survived a look: the
menu works, so the site feels fine until you click a card.

Two separate problems had stacked up. First, four of the cards pointed at tools **that don't
exist** — two were real tools under different names, and two were tools nobody has ever built.
Another four pointed at category pages ("colour tools", "CSS tools") that were never created.
Second, the links were written in a form our hosting cannot serve: the site lives in a storage
bucket, and a bucket can't work out that `/tools` means `/tools/index.html` — it needs the full
address. That second one affects every site we run, not just this one; I checked relojistas,
robot-hands and gaswholesalers and they all behave the same way.

I also corrected the wording on two cards. One was titled "Spacing scale calculator" and we
have no such tool. Just repointing it somewhere would have fixed the broken link and left the
card promising something it doesn't deliver — which is the same problem wearing a disguise. So
those two cards now describe the tool they actually open.

**One thing I got wrong along the way, worth knowing.** I assumed the site was served from our
database. It isn't — the pages are published as files into a separate code repository and a
GitHub action ships them to storage. My copy of that repository was 394 commits out of date and
appeared to contain no home page at all, which for a moment looked like a much bigger problem
than it was. So I fixed it in both places and confirmed the deployment ran.

**Now the part that matters more than the bug.** You asked how this got live without being
checked or flagged, and the honest answer is that the checks exist, are well written, and
**have never run — on any site.**

We have three checks built specifically for this: dead internal links, dead buttons, and
misdirected calls-to-action. They're switched on for exactly one automated inspector, and that
inspector has not run once in the entire period our records cover, for any of our sites. The
only inspector that has ever run is a different one that looks at design and never looks at
links. So this site *was* inspected, the day before it went live, by the inspector that doesn't
check the thing that was broken.

The reason is mundane: the only thing that routinely triggers the link inspector is the
improvement loop, which you've got switched off, and the only scheduled job pointing at it is
disabled and was set up for a different site anyway.

**Why this is worse than a missed check.** A site with no link problems reported looks exactly
the same as a site nobody has looked at. I told you this site was live with all 98 pages
returning 200, and that was true — but "every page loads" and "the site works" are different
claims, and I'd only checked the first. I've written it up as its own fault to be fixed
(`bugs_open/116`) rather than burying it, because a checker that's been improved but never
scheduled is worth exactly as much as no checker.

Your instinct that the checkers should run after every build or change is, I think, the right
fix rather than just switching the loop back on. A periodic sweep can only tell you about
damage after it's shipped; running on every change means "no problems found" would actually
mean something.

**One thing I stopped myself doing.** My plan said I'd fix the address-format problem in the
platform code. Before editing I checked who owns that area, and it turns out two other active
workstreams are working in exactly that file and have already written up their reasoning about
it — and a third is working the related half. I'd have been shipping a competing fix into three
people's work. So I've handed them my evidence instead and changed no shared code. I've also
logged the process mistake: I ran the ownership check *after* writing the plan rather than
before, and it takes under a second.

**Two things left over.** The favicon is missing on all 98 pages — it's a genuine small 404 on
every page, but none of our sites has one and picking a brand mark is your call, not something I
should invent. And the home page currently has two near-identical "What's here" sections with
overlapping cards; that's a design question rather than a bug, but you'll notice it now that
the links work.

**And the honest caveat.** I repaired the stored page and the published file. Nothing upstream
changed, so if that page is ever regenerated it will come back broken. The home page is sound
today; the site is not yet sound as a rule. That's what `bugs_open/116` and the other two open
cases are for.
