# Where we are — Gemini for the content-producing agents

*Owner's running log. Append-only, newest at the bottom. Plain prose.*

---

## 2026-07-27 — why it was reversed, and why it's worth another go

You asked why we backed out of moving the content agents from Claude to Gemini,
and whether we can get it done this time. Here's the short version.

We tried it over two days, 23rd and 24th July. Two agents were involved and
they're configured in completely different places, which is part of why the story
is muddled. One is the **content creator**, which writes blog posts and social
posts — it's its own service with its settings in a config file. The other is the
**page content writer**, which writes the actual copy on every site we publish —
that one lives in the database. The blog-post one got switched over on the 23rd.
The page-copy one got switched on the afternoon of the 24th and switched back six
minutes later.

Two things went wrong, and both were real.

**First, the model we'd pinned didn't exist for us.** We'd asked for
`gemini-2.5-pro`. Google replies with a flat "no longer available to new users"
for our API key. It turns out Google closes off older model generations to keys
issued after a certain point, so this isn't about the model being retired — it's
about our key being too new. There's a floating name, `gemini-pro-latest`, that
does work, and that's what got used instead. This part was straightforward: a
clear error saying what was wrong.

**Second, and this is the one that killed it, the model barely wrote anything.**
On the short formats it produced nothing at all — literally zero words for a
tweet — and about 85 characters where we wanted a couple of paragraphs. The
thread investigating it found that the newer Gemini models "think" before they
answer, that the thinking couldn't be switched off, and concluded that the
thinking was eating the space where the writing should have been. The only Gemini
model that worked properly was a smaller, cheaper one that's a clear step down in
quality from what we'd picked pro for. You were asked whether to take that
downgrade, said no, and everything went back to Claude. On what was known at the
time, that was the right call.

**Here's what nobody spotted.** The diagnosis was three-quarters right and the
last quarter changes the answer. The thinking *was* eating the space — but it was
eating space *we failed to give it*. When we ask Gemini for a piece of writing we
send a number saying how much output we'll allow. With Claude, that number is
purely the writing. With Gemini, the same number covers the thinking *and* the
writing, and the thinking is spent first. Every one of those numbers in our system
was chosen years ago for Claude. So when we asked for a tweet in 100 tokens, we
were asking the model to fit its reasoning and its tweet into a space barely big
enough for the tweet. It thought, ran out of room, and said nothing. That's not
the model being incapable — it's us handing it a budget with no room in it.

Our own code never mentioned thinking anywhere. It passed the Claude-sized number
straight through, and it didn't even read back the figure that would have shown
what was happening — the count of thinking tokens was sitting in Google's
response the whole time and we threw it away. So the error message just said "ran
out of tokens", which looks exactly like a model that wanted to write more.

**The good news is that fixing it costs nothing.** That output number is a
ceiling, not a purchase — Google charges for what's actually produced. So we can
ask for a generous allowance for thinking on every call, and if the model only
thinks a little, we only pay for a little. I've changed our Gemini code to add
that headroom on top of whatever the caller asked for, and to record the thinking
count so this can never be invisible again. If it ever does run out of room, the
error now says so in plain terms and names the setting to turn up. I also fixed a
couple of related traps while I was in there: a dead model name now fails
immediately with the working alternative in the message rather than failing on
every call at run time, and there was a latent problem where the model's own
reasoning could have been pasted into published page copy, because Google returns
the thinking and the answer in the same list and we were concatenating everything
in it.

**Two things I want to be straight about.**

The first is that I have not proved this works. I've proved our code now sends
sensible numbers, and there are tests pinned to the exact settings that produced
nothing in production. But whether Gemini actually writes well for us is a
separate question, and it needs a real run. It's entirely possible the mechanism
explanation is correct and pro still turns out to be impractical — if it thinks
for tens of thousands of tokens on our prompts, that's expensive regardless. The
first thing to do is measure that, and I've written a script that does exactly
that in one go: which models our key can actually reach, how much thinking versus
writing we get at each of our real settings, and which of Google's thinking
controls this model accepts. Those answers were worked out by hand last time,
mid-incident, and then lost inside a commit message.

The second is that I can't finish it from here. The cluster login has expired in
this session — everything I try against it comes back "Unauthorized". That blocks
the remaining work: I can't build and roll the new code, can't verify it landed,
and can't run the probe, because the Gemini API key only exists inside the
cluster and there's no local copy. Everything up to that point is done and
committed.

**Worth knowing about the page-copy writer specifically:** it was never actually
tested on Gemini. It got switched over at 16:53 on the 24th and switched back at
16:59, and the page rebuild that was supposed to be the test was still sitting in
the backlog when the revert happened. So for the agent that writes our sites —
the one that actually matters commercially — we have no evidence about Gemini in
either direction. That's an open experiment, not a settled question.

**One thing I need from you.** Why did we want Gemini in the first place? The
repo doesn't record it, and it changes what I should configure. If it was about
cost, the cheap model already worked on the 24th and needs none of this headroom
machinery. If it was about quality or about not having everything on one
provider, then pro is the candidate and the headroom fix is what makes it
possible — but pro-with-a-proper-budget has genuinely never been run, so it's a
real experiment. Either way it's now a one-line config change to try it and a
one-line change to go back, so trying pro first costs very little.

## 2026-07-27 (later) — you've answered: pro, for quality and diversity

Noted, and that's the interesting choice rather than the safe one. It means the
headroom fix isn't a side improvement — it's the whole reason to expect a
different outcome, because pro is precisely the model that produced nothing last
time.

Two things I want on the record before we run it, so neither arrives as a
surprise.

The first is that I picked the headroom figure (8192 tokens) as sensible-looking
generosity, not from measurement. Nobody has ever watched how much pro actually
thinks on our prompts. The probe prints that number per setting, and the largest
one it reports is what the headroom should be set from. It's plausible the current
default is already comfortable; it's also plausible pro thinks well past it on a
long page section, in which case we'd see the same empty output as before and it
would look like the fix hadn't worked when in fact it just needs a bigger number.

The second is that the risk has moved. Truncation was the problem last time; with
headroom in place, **cost** is the thing to watch, because Google bills thinking
tokens as output. A model that thinks for twenty thousand tokens to write three
paragraphs is working exactly as designed under my fix and is still the wrong
answer for us. That's now measurable in advance — the probe reports it — and
trackable afterwards, since I added the thinking count to what we log per call.
If those numbers come back ugly, the honest thing is to put the flash-lite trade
back to you with real figures attached, rather than the way it was put on the 24th
with no numbers on either side.

Still blocked on the cluster login before any of that can run.

## 2026-07-27 (evening) — it works. Blog and social are on Gemini pro, and the tweet that used to come back empty now comes back written

The new build went out, and I can tell you the fix is real rather than
theoretical.

First I checked the running pods actually contain the new code rather than
trusting the version number — both the chassis and the content-creator service are
on v1.0.1173, and I grepped five separate phrases that only exist because of this
change out of the live binaries. All five present. I also checked two phrases that
should have *disappeared*, and they had. (One of my three "should be absent" checks
was useless, because I picked a phrase that exists elsewhere in the codebase
anyway — I've noted that so nobody repeats it.)

Then I switched the blog-and-social agent over to Gemini pro and put two real jobs
through it.

The first was the one that mattered: a tweet, at the exact setting that returned
**nothing at all** on the 24th. It came back with a proper 264-character tweet.
The second was a long blog post — 1,292 words, nothing cut off, thirty-five
seconds. So the thing we concluded Gemini couldn't do, it does, and the reason it
couldn't before was us.

One small bonus that confirms a second fix: the cost figure now comes back at
Gemini's rate rather than silently using Claude's. That's a detail, but it means
the cost numbers we look at from here are about the model we're actually calling.

**What's left, and why I've stopped short of it.** The page-copy writer — the one
that writes the sites, and the one that has never actually run on Gemini — is a
change to the live database rather than a config file, and my tooling declined to
make that write. That's a sensible guard, so rather than argue with it I've written
the statement out in full as a script you can run in one command. It's in the
workstream folder as `P6_FLIP_page_content_writer.sql`. It backs the row up first,
refuses to proceed if another session has edited the row since I read it, and —
this is the important bit — it checks its own work afterwards and undoes itself if
anything is wrong.

**Writing that script caught a mistake of mine that would have been genuinely
expensive.** My original instructions for this step replaced the whole settings
block with the three provider lines. The writer's output limit — 8,000 tokens —
lives inside that same block, so replacing it would have silently dropped the limit
to the default of 2,048. Less than a third. Nothing would have looked wrong: the
switch would have "worked", and then page sections would have started coming back
cut short some days later, and I'd have gone hunting in completely the wrong place.
I only caught it because I read the row before writing to it. The script now
preserves everything it isn't deliberately changing, and refuses to commit unless
the 8,000 is still there afterwards.

So: run that script when you're ready, then rebuild a single page and read the copy
before letting it near the whole estate. That last step is the real test of the
thing you actually asked about — whether Gemini writes well enough — and it's the
one question none of today's work answers.

---

**2026-07-27, later the same day — I was wrong about what was blocking us, and the
real answer is both simpler and more urgent.**

Earlier today I told you the last step was blocked by a separate fleet-wide fault
that had stopped all site builds since the 19th. That was wrong, and I should have
caught it: the builds had not stopped. They had been running all along, including
three the same day. I took the headline of another bug report and repeated it
without reading the correction that had been added to that same report six days
earlier, which said the fault is a brief window after a software update, not a
standing outage. One database query would have shown me sixty-two successful runs
the day before. I have recorded that properly as a wrong call.

The test page I queued did run, about half an hour ago. It failed, but for a dull
reason: that particular page had never been given a list of sections to write, so
there was nothing for the writer to write. Bad choice of target on my part, nothing
more.

**The real problem is something nobody had spotted, and it is the reason none of
this could ever have worked.** The two pieces of software we switched to Gemini get
their Google API key by two completely different routes. The social-content writer
runs as a permanent service with the key handed to it directly, which is why that
half worked and produced real posts. The page writer is different: it is started up
on demand as a short-lived worker, and those workers are given their credentials
from a hand-written list in our code. That list has the Anthropic key and the Grok
key on it. It has never had the Gemini key. The key itself is sitting there in our
secret store, perfectly fine. It simply never gets handed to the worker that now
needs it.

So the page writer cannot talk to Gemini at all. Not badly, not expensively. At all.

**This needs a decision from you today, and there is a clock on it.** The switch to
Gemini is already live in the database, and the site that rebuilds itself
automatically every six hours is next due at about half past eight this evening. When
it runs, the writer will reach for a key that is not there, and because that step has
no fallback, the whole page build will fail. Nothing is damaged and nothing is
published wrongly, but it will fail, and it will keep failing every six hours until
someone changes something.

There are two ways to go, and they are genuinely different decisions:

**Put the key on the list.** It is a small, obvious code change of about ten lines,
sitting right next to the two keys already there. It needs a rebuild and a restart to
take effect, so it will not be done before this evening's run. I have written it up
as a proper bug report so whoever picks it up has the evidence.

**Or turn the page writer back to Claude for now.** That is a database change, it
takes effect immediately, and it removes the clock entirely. Nothing we built today
is lost or wasted by doing this: the underlying fix that started this whole
investigation, the one that stopped us starving Gemini of room to answer, is built,
reviewed, shipped and confirmed running in production. Only the *choice of model* for
the page writer would go back.

**And there is a reason to think the second option is the right one anyway.** The
model comparison I ran this afternoon, before any of this came up, found that Gemini
spends around ten times as many billable tokens per section as Claude, because it
thinks at length before answering and we pay for that thinking. The page writer runs
once per section, on every page, across the whole estate. So the honest summary of
today is that we proved the mechanism works and we proved Gemini writes to our house
style at least as well as Claude does, and separately we found that it would cost
roughly ten times as much to use it for the job that matters most. That is a
commercial question rather than a technical one, and it is yours.

My recommendation: revert the page writer to Claude today to stop the clock, keep the
key fix on the list as a small job so the option stays open, and treat "is Gemini
worth ten times the cost for page copy" as the real question we now have to answer.

---

## 2026-07-27, later that evening — the page writer has now actually written a page, on Gemini, and it is live

Short version: it works. The thing we had proved in probes and bake-offs has now run
the whole way through the real machinery, unattended, and put readable English on a
public page. Nothing in the chain needed a nudge.

**But the job we were told to do first could never have worked, and it is worth
saying why, because the reason is not about Gemini at all.**

The handoff I picked up said: re-run the `grip-styles` build, because it failed
earlier today only because the writer could not reach Google, and that has since been
fixed. That was wrong. I read the failure rather than assuming it, and the build had
never got anywhere near Google. It stopped much earlier, with "no sections ready to
build". A page on this system is a list of blocks — a banner, a product list, a
closing pitch — and *nobody had ever decided what blocks `grip-styles` should have*.
There was nothing to write. The same is true of every article-style page on that
site: not one of them has a block list. Five earlier attempts to build such pages are
still sitting in the review queue from a week ago, all with the same complaint. So
the suggested fallbacks in the handoff would have failed identically, and I would
have added a sixth.

**What I did instead**, with your go-ahead: built the `sale` page, which does have a
block list. It is live now at `dartsonline.com/sale.html`.

**The copy is good, and it avoided the exact trap I was worried about.** A sale page
is where a machine invents "up to 40% off", and inventing numbers is a live, open
problem for us. It invented nothing. No percentages, no fake deadlines, no
manufactured urgency. It also cleared every style check we care about — no em dashes,
no exclamation marks, no filler, and it uses contractions like a person does. The
line I would point at is *"It's easier to test different weights and grip profiles
when the gear costs less"*, because that is the page explaining why you should care,
in plain words, which is the hardest thing to get out of these models. And it is
recognisably about darts — tungsten barrels, flights, grip profiles, tightening your
grouping — rather than generic sale language with the word "darts" dropped in.

**Two honest caveats.**

The first is mine to report and yours to weigh: the two blocks say much the same
thing twice. The banner says "Find Your Next Set on Clearance" and the closing pitch
says "Find your next setup in the clearance", and both go on to talk about discounted
tungsten and finding the right weight. Each block is written well; they were written
without sight of each other. That is a structural thing, not a Gemini thing, and I
have not tried to fix it.

The second is worse and has nothing to do with the writing. **The sale page has no
products on it.** The block list says banner, product list, closing pitch — and the
product list was dropped during the build because there was no product data to fill
it. So we have published a page that says, twice, in good English, that things are
marked down, with nothing to buy underneath. The system did the defensible thing at
each step: better to drop an empty grid than render an empty grid. The result is
still a shop page that cannot sell anything. I have flagged it rather than fixed it,
because the fix is product data, not copy.

**One thing I nearly got wrong, and the check that caught it.** When I first looked
for the new page it came back "not found", and my instinct was that our deploy had
silently failed — we have an open bug that does exactly that. Before believing it I
checked a different page that I had not touched at all. It came back "not found" too.
That is not something my change could have caused, so the fault had to be in how I
was looking, and it was: this site serves addresses ending in `.html`. The page had
been live and correct the whole time. Checking something you did not touch is a
thirty-second habit that keeps turning a false alarm into a non-event.

---

## 2026-07-28 — the decision: we keep Gemini, and not for the reason I'd been assuming

The owner has ruled: **we keep Gemini, because being able to run different models in the
same pipeline is part of the story we're telling.**

I want to record that carefully, because I had the question framed wrongly and it's
worth someone knowing that later.

All week I'd been treating this as a commercial decision waiting on a number. Was Gemini
worth roughly ten times another model's output tokens for the writing job? I built the
measurement precisely so that question could be answered, and I kept describing the
answer as "the owner's call" in the sense of *someone has to weigh the money*.

**The actual answer wasn't about the money at all.** It's that a business selling AI
work is in a stronger position demonstrating a pipeline that uses several different
model providers than one wired to a single supplier. That's a product judgement, and no
amount of measuring on my side could have produced it — I could have measured for
another week and still not had the answer, because the answer wasn't in the data.

**The measurement wasn't wasted, and the order it happened in matters.** We found out
what the thinking costs — about eight and a half words of invisible reasoning for every
word of copy we keep — and *then* the decision was made. So this isn't "we carried on
without checking". It's "we checked, and decided it's worth it". If someone reads the
ratio in six months and thinks it looks bad, they're disagreeing with a judgement
somebody made on purpose, not catching something nobody noticed.

**What the meter is for now.** It was built to inform this decision. With the decision
made, its job changes: it's the thing that will tell us if the ratio quietly gets worse
— if a prompt change doubles the thinking, or a model update starts costing more for the
same copy. Nobody needs to watch it daily. Somebody should look at it occasionally, and
now somebody can.

**One consequence worth flagging.** "We use different models in our workflows" is only a
story for as long as it stays true. Right now it is: the two content writers run on
Google, most of the estate runs on Anthropic, and the vet price scraper runs on a local
model. If a future tidy-up ever proposes consolidating everything onto one provider for
simplicity, this decision is the thing it would be overriding — and whoever proposes it
should know that the variety is deliberate rather than accidental.
