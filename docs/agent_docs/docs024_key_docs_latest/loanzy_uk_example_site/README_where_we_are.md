# Where we are — loanzy.uk

Plain prose, append-only, newest at the bottom.

## 2026-08-18

Nothing has ever been built on loanzy.uk. The domain itself is in good shape — it has a
Cloudflare zone, a certificate, and both worker routes, so it answers instantly with a small
404 that means "no site here yet" rather than anything being broken. That puts it ahead of
most of the portfolio, where domains still point at registrar parking pages.

The FCA-rules site you remembered is real, but it is lendzy.co.uk, not loanzy.uk — one letter
apart, and built the day after loancash.co.uk as a second run of the same brief. Both are
live and were redeployed this morning.

The decision this session: loanzy.uk becomes webdesign.uk's first example site, built only
from a customer prompt, with no positioning entry written for it. That matters more than it
sounds. The webdesign.uk page lead approved today is "show the work, promise nothing" — the
promise being that you can see real sites and the exact prompt that produced each one. The
copy is currently forbidden from mentioning examples at all, because the four sites we could
point at were not built the way a customer's site gets built. Building loanzy.uk from a
prompt and nothing else produces the first honest pair.

What is needed next is the prompt itself, which is yours to write, because it gets published
next to the site. One caution worth a sentence: the domain sounds like a lender, and a
convincing fake lender on a live UK domain is a compliance problem rather than a demo. Either
the demo is not a financial firm — which also shows off the new "any sort of site" line — or
the page says plainly that it is an example.

## 2026-08-18, later

We tried it your way — the framework given nothing but the domain name — and it answered
confidently and dangerously. It decided loanzy.uk should be a loan comparison and matching
platform: a panel of FCA-regulated lenders, an eligibility checker, and money coming from
lenders paying us for each borrower we refer. That is credit broking, which is a regulated
trade. It even found two unrelated companies trading as "Loanzy" elsewhere in the world and
treated their business as evidence for what ours should be.

I stopped the build, but not cleanly, and one page got out before I did. It sat on the live
domain saying we search a panel of FCA-authorised lenders. The account of exactly how that
happened, including the three separate mistakes I made trying to stop it, is in the summary
document written this afternoon — you asked for it in full and it is in full.

Two things came out of it that are worth more than the site would have been. The first is
that the system already knew: two steps before any page was written, the briefing agent had
noted that an FCA authorisation number "must be obtained before launch" and that the lender
panel was not confirmed. It wrote that down and carried on, because nothing gates on a note.
The second is that our unpublish path cannot unpublish the last page of a site — the deletion
empties the folder, the deploy skips folders that do not exist, and both halves report
success. That is filed as a bug.

The classifier now carries the rule you asked for: a regulated business model is not an
available answer unless the brief explicitly asks for one, and a domain name tells it the
subject, never that it may broker or advise. That is live in configuration, though nothing has
been built through it yet, so it is not yet proven.

The page from this morning is still up. Both ways of removing it were blocked by my own
sandbox — one deletes production data, the other writes to the deploy repository — so that
one needs you.

## 2026-08-18, evening — it is live, and it is the opposite site

loanzy.uk is up. Given nothing but its own name — no brief, no facts, no contact details — the
framework has built a loan education site: calculators that run in your browser, a glossary, a
page pointing people at free debt advice, and a home page that says in its own words "we do not
arrange loans, introduce you to lenders, or take a cut when you borrow anywhere."

This morning the same domain with the same silence from us produced a credit broker with an
eligibility checker and a lender panel. The only thing that changed in between is the rule you
asked for, which now tells the classifier that a regulated business is not an available answer
unless the brief asks for one. I proved it by running the same thing twice and changing nothing
else, and kept both answers on file so the comparison can be checked rather than taken on trust.

Three things are unfinished and none of them is about that rule. The rights page did not build:
it hit a known platform bug that leaks raw template syntax into the page, which another team hit
on two pages today, so I added our case to it rather than filing a duplicate. The guides page
built nothing because there are no guides yet, and the builder refused to ship an empty shell,
which is the right call. And the menu is still the old one — it offers "Check Eligibility" and
"Lenders", pages that no longer exist — because the navigation rebuild sensibly refuses to run
while half the site is missing. That last one is the only thing on the live site that
contradicts itself, and it is next.

---

**2026-08-23, late afternoon — the garden-tools build is away, and two of our own claims turned
out to be wrong before it had produced a single page.**

I started the clean-domain test. garden-tools.uk went in at 17:17 with nothing attached to it — no
brief, no contact details, no seed, just the name, exactly as you asked. All the pre-flight checks
were clean first: no rows for the domain, the domain still serving the little "Not found" that
means the plumbing is right and the shelf is empty, and the chassis untouched for over an hour. I
then checked that the submission actually landed rather than trusting the script, because that
script still reports success whether or not it published anything. It landed.

Nothing has been built yet. Twenty minutes in, the job is sitting in a queue, and I want to be
clear that this is not a fault: I went and looked at why, and the machine that hands work to the
builders walks the sites one at a time, about ninety seconds apart, in a fixed order. Our new site
was created about forty seconds after the current lap had already started, so it missed this lap
and waits for the next one. I wrote that prediction down before the outcome so it can be wrong in
public — if it misses a second lap, that is a real defect and a much more interesting one. Either
way we have learned something worth having: **a new domain waits up to about three quarters of an
hour before anything starts, and the submit command returning instantly tells you nothing about
that.** Nobody had measured it.

The more useful part of the afternoon was two things we had written down that were not true.

The first was a trap I walked up to and did not step in. The instructions I inherited said that
when this build finishes I should re-check eight stored calculator templates and confirm they have
not changed, and to say so loudly if they had, because a build stealing another site's calculator
is the exact damage we are guarding against. I checked them **before** starting instead, and they
had already changed — three days ago, by a different team's tidy-up, which kept a copy of the
originals. Had I run that check only at the end, as written, I would have reported that our build
had damaged another site's live content. It would have been wrong, it would have been loud, and it
would have landed on someone else's bug. The general lesson is dull but it keeps biting: a
fingerprint taken days ago quietly stops being a baseline, and it goes on looking like one.

The second was ours outright. We filed a bug in this lane saying that once a build fails you can
never retry it, because the database refuses a repeat. Another session picked that bug up today to
fix it, read the database rule we had blamed, and found it says the opposite — it explicitly allows
a repeat after a failure. The real cause is a separate three-hour cooling-off period. I checked
both halves myself rather than take their word, and they are right. Worse, the recovery we
recommended and recorded as a success — renaming seventy-eight rows by hand — was probably
unnecessary; waiting would very likely have done the same job. I say probably, because renaming
those rows destroyed the only records that could have settled it. That is the uncomfortable one:
our own fix erased the evidence for the claim we then made about it.

I have corrected all four documents that carried the wrong version, written the whole thing up in
the fleet's log of wrong calls, and left the fix itself with the session that is doing it rather
than starting a competing account. The practical upshot for us: if this build fails tonight, we
wait, we do not start renaming things.

The build is still running. I am watching it and will report what it produces.

---

**2026-08-23, evening — the answer, and it is a clean one even though the website does not exist.**

The build is dead and we know exactly why. It got as far as deciding what the site should be — an
independent gardening-review hub, no regulated angle, twelve pages — and then stopped at the very
next step, which is the one where it studies three examples of the best sites in the field.

It chose Gardeners' World, The Spruce and Which?. Our scraping service flatly refuses to fetch The
Spruce. That one refusal throws away the entire step, including the sites it had already read
successfully. It tried three times over an hour and a half, picked the same three sites each time in
a different order, and died on The Spruce wherever it happened to land. Then it stopped for good,
because the step that hands the build to the next stage is the last one in that sequence and is the
only thing in the whole system that can start what comes after it.

So we have a site row, four pages of notes about what the site should be, and nothing else. Nothing
will ever pick it up again on its own.

I then did the thing an operator would do — submitted it a second time — and that produced two
useful results. The first is that another team's fix, shipped this afternoon, worked: the front door
accepted the resubmission where two hours ago it would have silently swallowed it and reported
success. They now have live proof of their fix on a real build, which they did not have before.

The second is more interesting for us. The second submission re-ran the "what is this domain"
step from scratch, and **it gave the identical answer** — same category, same style, same twelve
pages, and the same confidence score of 0.82 to two decimal places. Only the free-text list of
keywords came out differently worded. That is worth knowing: the machine's judgement about a domain
is stable, so we can use it as a fixed point when testing other things.

And it let me run the test that mattered. Because the second run produced genuinely different notes,
I could ask whether the three example sites were being chosen because of those notes or because of
the field itself. **It picked the same three sites again.** So this is not bad luck and it is not
something a retry will ever fix — the machine will keep nominating the same handful of famous
gardening sites, one of which we are not allowed to read, for ever.

I have not fixed it. Fixing it would have meant the build was no longer a measurement of what the
framework does unaided, which is the whole point of this exercise. It is written up as a new bug
with the evidence, and the fix worth having is not "try again" but "remember which sites we cannot
read, and stop suggesting them".

Two smaller things from today, both of them us being wrong rather than the system. A safety check we
were told to run after the build, I ran before it — and found the thing it checks had already
changed days ago for innocent reasons. Run in the order we were given, I would have loudly announced
that our build had destroyed another site's work. And a bug this lane filed four days ago blamed the
wrong mechanism entirely; another team found the real one, and I have corrected the four documents
that had been repeating it since.

---

**2026-08-23, 20:10 — correction: the build is not dead. It got past the wall half an hour after I
told you it could not.**

I need to correct what I wrote earlier this evening. I said the build had died permanently and that
trying again would never help, because it had chosen the same three example sites four times running
and been refused by the same one each time.

On the fifth try it chose differently. It dropped the site we cannot read, picked a British garden
tool maker instead, read all three successfully, and carried on. It has written up what it learned
about the field and has moved to the next stage, which is deciding the site's strategy. As I write
this it is working on that.

So the honest position is narrower than the one I gave you. The flaw is real: one unreadable example
still throws away the whole step including the examples it read fine, and if that happens three
times the build dies for good — which is exactly what happened to the first attempt. But it is a
four-in-five problem, not a certainty, and I described a run of bad luck as if it were a law.

**That is the second time today I have done that**, and it is the thing I would most want you to
take from the day. This morning I explained a delay with a theory built on fourteen consistent
observations, and it fell apart twenty minutes later. This evening I said "disproved" on the
strength of four, and it fell apart in thirty minutes. Both times the machine handed me the
counter-example for free, simply by carrying on running, and both times I had already written the
strong version down.

The rule I have written into the notes is simple enough to hold onto: **if what I have is a count,
the claim has to contain the count.** "Four times out of five" was always what I actually knew, and
it would still be true now.

I have corrected the bug report, the notes, and both handover documents. I have kept the wrong ones
rather than tidying them away — a handover whose title says "it died" and which was wrong within the
hour is worth more as a record than as an embarrassment.

The build continues. I am not touching it.

---

**2026-08-25, morning.** You retracted the authorisation to clear that parked row, which is what I
asked for yesterday evening after finding my own proposal was built on something I had not checked.
It is now written into the top of yesterday's handoff, so the next person reading that file cannot
pick the plan up again by accident. Nothing was done to the site.

**The new chassis went out at 09:27 and I checked what it actually contains rather than assuming.**
The claims check — the one that catches a site saying "we test these tools" when nobody has tested
anything — is now running in the live software. I confirmed that by asking the running program
directly, twice, and by asking it two questions I knew the answers to first so I could tell a real
answer from a broken check.

**One honest caveat on that.** It is live, but on garden-tools it is doing nothing, because it only
runs while a page is being built or rebuilt and that site is sitting still. Live is not the same as
firing. If I told you "no problems found on garden-tools" that would only mean nobody asked.

**The site itself has not moved.** Seven pages serve, five still 404, nine dead links, and the
seasonal planner still promises a month-by-month guide it does not contain. I re-ran the full check
this morning rather than trusting yesterday's numbers — and yesterday's page sizes turned out to
have been taken a few hours before a redeploy, so they were already slightly out of date when they
were written down. The important parts were unchanged.

**A number I decided not to send you.** I measured how much more structure the writer is producing
since yesterday's fix and got a modest-looking improvement. Then I noticed I had counted every piece
of writing on the estate, including all the parts the fix never touched — which drags the average
down and makes a good fix look mediocre. The lane that made the change measured only the writing it
actually applies to and found lists going from roughly one in ten to roughly seven in ten. Theirs is
the right number. Mine would have understated their work, and I would have been the one to publish
it first.

**Where this leaves the decision you were weighing last night.** The option of "just wait and watch
whatever gets built next" has quietly half-answered itself: hundreds of pages have been written
since the fix went in, so the *writing* half is proven in the wild. But the *planning* half — the
part that chooses what goes on a page, and the three new building blocks that lane created overnight
for exactly this — has not run once, because it only runs when a brand-new site is built. So the
question is no longer "does the fix work". It is narrower: **does the planner actually use the new
pieces when it lays out a page, and only a new site build can show that.**

**And there is a gate in front of that.** The bug I filed on Sunday — where one website refusing to
be read can kill an entire new-site build stone dead — is still nobody's job. On the gardening
vertical the refused site turned up in four draws out of five, and the system gives up after three
attempts. So if you authorise a new domain today, there is a real chance it dies half an hour in and
tells you almost nothing. That is the piece I would fix next, and it is this lane's own bug.

---

**2026-08-25, afternoon — your review of homegarden.uk, and where each part of it has gone.**

You gave two decisions and a long list of problems. The decisions are recorded: **more than four card
sections before something has to break them up**, and **a re-plan of garden-tools is fine**. Both are
written where the lanes that need them will find them.

**I checked every specific you raised against the live site rather than taking it on faith**, and all
of them hold. The calendar has twelve months and **no links at all** — so it cannot take anyone to the
monthly guides. Every page has exactly **one image**, which is the logo. The garden page really does
say *"Check what's due in the garden this April"* in August. And the about page has seventeen
headings of which **fourteen are about how the site works** rather than about homes or gardens, with
*"What this site will not do"* appearing **twice**.

**I added one you didn't mention**, because it is the purest example of the thing you were describing:
every page carries a **"Get Started"** button pointing at the contact page. On a gardening site that
is a software-company button that means nothing — get started with what? Nobody asked what a reader
wants there; a template supplied it.

**Where it has gone.** The copy lane has your instruction in full, including the two parts I did not
want softened: that they should **go and re-read the accumulated copy discussion before proposing
anything**, and that they should **audit every prompt in the database and the code** against the
question of whether it encourages AI writing. I flagged that the third one is a workstream rather
than a task. The design and offer lane has the imagery measurements, the button problems, and your
point that better imagery should be **the default for every site, not a fix to this one**.

**One thing I want to put to you rather than decide.** You proposed a user-experience agent, and
separately an agent to look at detail, and improving the experience loop. There are already eight
live agents in that territory — an experience planner, an experience approval council, an offer
analyser, a visual designer, a visual design auditor, a brand designer, a feature designer and a
design audit agent. So the question might not be whether to build another one, but **why the eight we
have produced this page**. I have asked the owning lanes that directly. If the honest answer turns
out to be that none of them runs on this path, that is worth you hearing plainly, and I would rather
find out than add a ninth.

**What I have not done:** touched the site, written any copy, or proposed design fixes. Those belong
to the lanes that own them, and I have handed each of them the evidence rather than my opinion of it.

---

**2026-08-25, late afternoon — the reviewer at the end of the build, and the surprise when I went
to write it up.**

You asked for a benchmark, then a structural floor of six, then said you liked the idea of a council
of checkers that judges a finished site the way the code council judges a change — running after the
fact in the improvement loop, routing its findings to the agents we already have. I sat down to
write the reference for that.

**Before writing I checked whether anything like it existed, and it does.** The improvement loop
*is* that reviewer. It already calls an offer analyser, a brief-fidelity auditor, two mechanical
checkers for empty sections and broken links, and a design audit — and it already has steps for a
**news feed** and a **directory**, the two things you asked for this afternoon. It ran for nine days
in August, filed real work on seven sites, and then its only timer was switched off on the 17th. It
has sat dark since.

So the route is much cheaper than I said an hour ago. Not "build a council" — **switch this one on,
which you have already authorised, and add the four lenses it lacks**: does the site have the things
it needs to have been born with (research, feed sources, an evidence base); does each page deliver
what its own headings promise; does it use at least six real structures or say why not; and would a
person who came to this kind of site actually want each page. That last one is the "happy user" you
have asked for three times today in three different words, and it is distinct from the auditor we
have, which checks whether a site is faithful to its brief — and the brief can be the problem, as
homegarden's anti-commercial one was.

**Two things in the loop's own configuration that will bite the moment it is switched on**, both
measured and now written where the lane switching it on will read them: if the news-feed step or the
directory step fails, the loop swallows the failure and carries on, so a site with neither passes
with no record of why. That is the same silent-skip shape as the claims auditor had until yesterday.

The reference document is in this directory, dated today, with every number and its date. The next
step is the RFC, because the definitions of those seats *are* the benchmark, and that is exactly the
kind of thing the architecture seat should rule on rather than a lane deciding quietly.

## 2026-08-25, evening — what the "evolutionary bit" actually is, and the plan to switch it off without losing it

You asked for four things this evening. Here is where each stands, in plain terms.

**The 376 fix is written and in front of the council.** It is one config migration: a refused crawl now
formats to "no content" instead of killing the whole research stage, a floor says "at least two of the three
exemplars must have actually delivered content" (counted on content, not on whether the step said success —
which.co.uk said success and delivered nothing), and below the floor the build fails loudly with the three
counts in the error rather than quietly writing a landscape from no research. I rehearsed it against the live
row inside a transaction I then threw away, twice, including the rollback. Not applied until the verdict.

**The "evolutionary bit" is the four LLM reviewers' opinions being turned into page rewrites.** The loop has
three mechanical reviewers (they find broken things) and four LLM ones (design audit, site review, offer
analyser, brief fidelity). The LLM ones do not fix anything — their findings go through a router that sends
"this page could be better" to the agent that regenerates the page. Over its life that one route produced
about a thousand page rewrites and four hundred rebuilds from a single reviewer. Bug 238 is what one did to a
page that was fine.

**Two things I did not expect.** First, switching the sweep off on the 17th did not switch the rewrites off:
a separate promoter, added on the 15th, dispatches any parked finding whose fix-type has ever succeeded, every
15 minutes. Twenty-six of these went through while the sweep was off. Second, seven of the eight reviewer
calls are wired so that a failure goes exactly where success goes — a site can be stamped "audited" when
nothing audited it. The other lane measured that; it is in the RFC.

**The plan (written up, nothing changed):** Phase 1, on your word — one edge in the loop takes the four LLM
reviewers off the path (they stay in the workflow, nothing deleted, rollback is the same edge back), then the
sweep goes on. Phase 2, Go, already written and tested — a "record, don't dispatch" mode for the router, so
a reviewer's opinion becomes a row a person can read and release, never a rewrite. Phase 3, after the next
roll — the four new seats (prerequisites, promise, structure with your N=6, and the reader seat you asked
for three times) plus the existing LLM reviewers switched back on in record mode. That is the acceptance
council, after the fact, with the rewrites removed.

**Two choices are yours:** whether to apply Phase 1 now (I recommend yes), and what to do about the render
audit — a separate hourly job that files contrast failures at the CSS patch agent; it is the other live
source of bad renders (239 in a fortnight) and this plan does not touch it.
