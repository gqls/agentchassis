# README — where we are (plain prose, append-only, newest at the bottom)

**2026-07-20** — Kicked off. You asked for more visually interesting components for
our consultancy-style brochure sites — the kind of thing Bain, BCG and McKinsey do:
a hero made of a few cards in a carousel that changes itself, each card with a photo
that gently zooms when you look at it instead of a video, a short title, a "read
more" link, and hardly any other text. Then further down the page, several different
kinds of blocks, a lot of them carousels you can swipe on a phone. And you want the
pages those links go to to look different from each other too, not one template
repeated.

You mentioned fundamentallyai.com as a possible new brand to try this on. I checked
it just now — right now that domain isn't yours. It's sitting on a "domain for sale"
marketplace page (Afternic), which means someone else currently owns the registration
and is trying to sell it. Before we design anything for that specific brand, you'd
need to either buy it or tell me to use a domain you already have. I haven't
assumed either way — just flagging it so we don't build a plan around a name we
don't hold yet.

I also looked at the consultancy site you already have running through this
platform — leopardessconsulting. Useful context: it deliberately uses flat
illustration-style images, not photos, partly because the current image generator
can't reliably render legible text in a photo-style image, and partly because that's
the house style you chose. If the new brand is meant to look like Bain/BCG/McKinsey
— real (or real-looking) photography, people-focused — that's a step further than
what leopardess does today, and it raises the exact question you already anticipated:
how do we show "people" imagery without our veracity checker (correctly) treating it
as inventing a person who doesn't exist. That's a real design question, not just an
engineering one, and I want your steer on it rather than picking for you.

Right now I've got two things running in the background: a deep external
research pass on what Bain/BCG/McKinsey and similar firms actually do
(specific techniques — hover-zoom instead of video, swipeable carousels, how
much text each card carries) and an internal pass mapping exactly how our own
system turns a site's spec into a rendered page, so any new component we design
slots into the real pipeline rather than being a one-off. I'll come back here once
both land with a concrete proposal for the first component or two to build, and the
open questions above for you to weigh in on.

**2026-07-20, later the same day** — Good news: you confirmed you do own
fundamentallyai.com after all (my check just caught it mid-redirect to a
"domain for sale" page — that's resolved now, hosting's coming shortly) and you
settled the imagery question yourself: line illustration for people, not
photography. That sidesteps the whole "don't invent a person" problem cleanly.
You also confirmed leopardessconsulting's content is genuinely factually
correct, which matters because you then asked for something bigger: don't just
make the new site look good, make it actually market what this platform can do
— embeddings-based private search for a partner's own database, instant
marketing/test/presentation sites, our fine-tuning work, the multi-agent
council review system, and real backend projects like idea.uk's Stripe
integration and the relojistas.com traffic-revival work — all as honest,
evidence-backed case studies rather than generic consultancy filler.

The external design research came back with real, useful detail even though
one part of it (the final "merge everything together" step) hit a system usage
limit partway through — I pulled the raw findings out directly rather than
losing them, so nothing was wasted. I now have a solid, concrete recipe for the
carousel and hover-zoom effects you described, with the accessibility rules
that come attached to them (a carousel that auto-advances has to be pausable,
for instance — that's not optional, it's a legal accessibility requirement in
most places).

I've also kicked off a second research pass, this time internal: what does
this platform actually, truly do well enough to put on a marketing page? A
quick check already turned up real code for both a fine-tuning-style
embeddings mechanism and a "RAG" (retrieve-and-generate search) mechanism, so
the private-search idea may be more real than aspirational — I'm having that
checked properly before we say so anywhere. Once that lands, I'll bring back a
grounded list of what's genuinely true and impressive, and a first design
proposal, before any actual page copy gets written — and to be clear, the
final copy will always be written by the platform's own content-writer from a
brief, not typed up by hand here, same as we agreed for leopardess.

**2026-07-20, later still** — The internal research came back, and it's good
material, with one important catch. The genuinely strong stuff: a real
13-reviewer AI council that's been live for a few days and already caught a
real production bug another team missed — I think this is the standout story.
A real fine-tuning run we evaluated honestly against a frontier model and
reported the result even though our model lost most of the time. A real Stripe
payment integration on idea.uk. And a clean, honest case study on
relojistas.com — we found a dead website's forgotten subscriber feed was still
being checked daily, rebuilt the site, and got that feed working again within a
day, and we were upfront in our own notes that most of that traffic is just
search-engine bots, not real people, rather than dressing it up.

The catch is about your embeddings idea. The actual search technology is real
and does work — but the "safely, without leaking to other organisations" part
isn't true yet on our own system. Right now everything sits in one shared
database with nothing stopping one site's data being visible to another; it's
a bit of a "the door isn't locked yet" situation. So I don't think we should
say we already offer that safely — I think we say we have the hard technical
part solved and can build a properly walled-off version for a client, which is
still a strong, honest pitch, just a different one than "we already do this."

I also found the original list of what got fabricated on leopardess last week
— a made-up founder, a made-up "70+ agents in 8 departments" structure, made-up
client case studies, a fake uptime number. None of that should ever show up on
the new site, even by accident, so I've written the exact list down as a hard
no.

There are a few calls only you can make before I go further: how you want the
embeddings pitch worded, whether you want to tell the story of us catching our
own past mistake as part of the pitch (it's compelling, but it means admitting
the mistake happened), and whether leopardess should be named directly as one
of "our sites" or just referred to in general terms. Once you've weighed in,
the next real step is turning all of this into a proper brief for the site's
content-writer to work from — not me writing the actual page text by hand.

**2026-07-20, decisions in** — You've answered all three: word the embeddings
pitch as "buildable, not already delivered," use the self-correction story, and
name leopardessconsulting.co.uk directly rather than keeping it vague. Those
last two fit together well — naming the actual site is what makes "we caught
our own mistake and fixed it" a real, checkable story rather than a
suspiciously vague anecdote, which is rather the point of telling it at all.
I've written all three into the plan as resolved, with one guardrail attached:
that story has to be built strictly from the real audit record of what was
fabricated and how it was fixed, not embellished — otherwise we'd be
undermining the very point we're making. There's nothing left needing your
input on this half of the work; the only thing still pending is your hosting
being pointed at fundamentallyai.com. Once that's live, the next real step is
turning this research into an actual content brief and firing the framework
for real — and separately, starting on the first new component (the hero card
carousel) so there's somewhere for that content to go.

**2026-07-20, trigger fired** — Ran the onboarding trigger for
fundamentallyai.com with the mission brief above. I confirmed directly (not
just trusting the script's "submitted" message) that it really did reach the
system's message queue. But it hasn't started processing yet — there's a
known, already-documented backlog on the shared queue every session's triggers
go through (single lane, and several of us are using it at once), so it's
sitting in a real queue, not lost. Past measurements put the wait anywhere from
about 25 minutes to a few hours depending on how busy things are right now.
I'm not going to resubmit it or force anything — that would just create a
duplicate and waste money — I'll check back on it.

**2026-07-21** — It ran. Overnight the pipeline built almost the whole site on
its own, and it came out genuinely on-brief. The palette is a smart dark navy
and amber, the design instructions correctly say people are line-drawn with one
consistent tint and never photographed, charts are to be drawn from real data
rather than generated as pictures, and — the bit I'm most pleased about — the
pages it created are named for exactly the things we decided to lead with: a
multi-agent review council page, a fine-tuning page, and a page that tells the
"we caught our own mistake" story and names leopardess directly, just as you
approved. So the brief carried all the way through the machine.

It's not live yet, and the two reasons why are actually the honest, useful part.
Five of the content pages, the homepage among them, are being held back by our
own content-checking gate — one issue flagged on each — while two other pages
passed fine. I can't yet see what the gate is objecting to, because the logs
that would tell me got wiped when the fresh build (v1.0.1144) restarted the
system. Rather than guess, the right move next thread is to rebuild one page and
watch it get rejected live. The second reason: nothing's actually being served
to a browser yet. The domain's DNS has now switched over to our setup, but the
pipe from our built pages to what a visitor would see isn't delivering, and I
haven't got to the bottom of why — part of it is probably a hosting step at your
end, like when idea.uk went live. Worth knowing: our database happily says two
pages are "deployed" when nothing actually loads, which is exactly why we never
trust the status and always check the real page.

I've written a proper handoff (HANDOFF_2026-07-21_start_here.md in this folder)
so you can open a fresh chat and carry straight on — it has the current state,
the two things to fix, and the still-to-come component build, all with the IDs
and commands needed. Nothing here is broken or half-done in a risky way; it's a
built site waiting on its last mile.

**2026-07-21, later — we now know exactly why the pages are blocked, and it's a
fixable own-goal.** The handoff assumed we'd have to rebuild a page and watch it
fail live to find out what the content checker was objecting to. It turned out we
didn't need to: the checker already writes the real reason into a database table
for exactly this situation, and the reason was sitting there all along — I'd just
been looking in the wrong place first time. (I've noted that as a lesson: check
the code before believing "it's unrecoverable.")

The reason is a bit of an irony given what this site is *about*. Our own
content-checker has a rule that flags a page if it mentions one of our other
sites by name — a sensible guard normally, meant to catch one site's text
accidentally bleeding into another's. But this new site's whole job is to talk
about our other work, and in particular to name leopardessconsulting.co.uk
directly as the "we caught our own AI's mistake and fixed it" story — which you
approved precisely because naming it is what makes the story credible. So the
guard is firing on the one reference we deliberately want there. Every blocked
page is blocked for that single reason.

There's a nastier side to it worth telling you straight: the two pages that *did*
get through only got through because, on a retry, the machine quietly rewrote
them to drop the leopardess mention. So right now not a single page on the site
actually contains the self-correction story — the pages that kept it are blocked,
and the pages that lost it "passed". That's the database-says-done-but-the-page-
isn't problem again, and it's exactly why we don't trust the status field.

The fix is small and honest: teach the checker that a site can have an approved
list of our own sites it's *allowed* to name, and put leopardess on that list for
this brand. It leaves the guard fully in place for every other site — if nothing's
on a site's allowed list, nothing changes for it. Because it's a change to the
platform's own code it needs to go through our review council first and then a
rebuild before it takes effect, so it's not instant, but it's straightforward and
low-risk. I've filed it properly as bug 055 with all the evidence. I'll take it
through the council, get the code rolled, then rebuild all the content pages so
the story is actually present this time, and check the real pages rather than the
status. The "nothing serves yet" problem is separate and still to sort — but
there's no point serving the pages until they've got the real content in them,
so this comes first.

**2026-07-21, later still — the fix is written and the review council has been
round it once already, which was worth doing.** You made two calls for me first:
go ahead and roll the fix out fleet-wide once it's approved, and put not just
leopardess but our whole intended portfolio (idea.uk, relojistas, finetuning) on
this site's "allowed to name" list so a future mention can't re-block it. I wrote
the fix, it passed its tests, and I sent it to the council rather than just
shipping it.

The council came back with "revise, not yet" — and it was a genuinely useful
"not yet", not a nitpick. Five of the eight reviewers were happy with the actual
code. The other three all made the same essential point in different ways: I'd
written the fix to the *mechanism* but hadn't written down the steps that
actually *close the problem* — flipping the switch on for this specific site
(with proper backup-and-verify care, not a casual database poke), checking the
new code is really running on the live server before trusting it, and rebuilding
the already-broken pages rather than assuming the fix alone heals them. Fair, all
of it. One reviewer also spotted something sharper and more important: the reason
the story vanished in the first place is a *separate* fault — when a page trips
the checker, the machine quietly regenerates it without the offending bit and
ships that, with nothing recording that something was dropped. My fix stops the
checker firing wrongly here, but that silent-drop behaviour is still lurking for
the next time any page trips any check. I've written that up as its own bug (056)
so it gets looked at properly rather than being quietly assumed "fixed" by this
one — because assuming exactly that is the trap.

So I've rewritten the plan to include the whole rollout, not just the code, and
sent it back round. I'm holding the code uncommitted until it's approved (it's
harmless sitting there — it does nothing until I switch it on for this site).
Once the council's happy, the sequence is: roll it, prove it's live, switch this
site on, rebuild every content page, and — the bit that matters — actually read
the finished pages to confirm the leopardess story is really on them this time,
not just that the database says "done".

**2026-07-22 — I have to correct myself, and it's an important one. You caught it.**
You said "double check your findings — v1.0.1146 is on production", and that made
me actually look at the live server's program instead of reasoning about dates. It
turns out the fix I'd spent the session writing and sending round the review
council **already existed and was already running in production** — someone (in
the big v1.0.1146 batch yesterday) had already built exactly the same thing. My
own change to that file was, in the end, one comment. So there was never any new
code to write or any server rebuild to do. I got that wrong by assuming, from the
timing, that the fix couldn't already be live, instead of just checking — and I
nearly did an unnecessary full-fleet rebuild off the back of that assumption. I've
written the miss down plainly in our shared "wrong calls" log so the next person
doesn't repeat it: before writing or rolling a fix, check whether it's already
there and already live.

The genuinely useful news: the *one* thing that was actually missing all along was
the simple switch-flip — telling this specific site that it's allowed to name
leopardess. No code, no rebuild; just a small, careful, backed-up database write.
I've now done that (it added the four of our own sites we agreed this brand may
cite as case studies). So the guard will now let the leopardess story through for
fundamentallyai and no other site is affected at all.

What's left is the part that actually puts the story on the page: rebuilding the
content pages that are still stuck, and then — the bit that matters most — reading
the finished pages myself to confirm the story is really there, rather than
trusting a "done" flag. There's a known wrinkle (the same silent-drop fault I
filed as bug 056) where a rebuild can still leave the bit out by chance, so I'll
check each page and re-run any that come back without it. The "nothing serves yet"
hosting step is still separate and still yours, but as before there's no point
wiring that up until the pages actually carry their content.

**2026-07-22 — the site is live, and I've done the last-mile cleanup.** You
deployed the new build and wired the hosting, so the site is now actually visible.
It doesn't look like the brief yet — that's the stage-2 job — but it works. You
spotted two broken menu links (a "Platform Log" and a "Decision Record" page). Those
turned out to be empty shells the pipeline created but never filled, and you asked
me to take them out of the menus. That was fiddlier than it should have been —
the menus are baked into a shared header/footer that every page reuses, and it took
me a few wrong turns to find where they actually live — but they're now gone from
every page (I checked each live page directly, not just the database). I've written
the wrong turns down so the next person finds it faster.

I've put your phone number on the site. And I've published the "our work" showcase
you approved — the three verified projects (the revived watch site, the idea.uk
payments work, and the leopardess honesty story), with "rebuilt same day" as you
said. The homepage is rebuilding now to show them; it's going through a backed-up
internal queue so it'll appear shortly rather than instantly.

One honest note: while fixing the menus I found a separate small broken link on the
council page — a "Review a sample record" button that points at a slightly wrong
address and 404s. It's a pre-existing content glitch, nothing to do with the menu
fix, and I've flagged it for the content pass rather than let it slip by unnoticed.

So: the site is live, the broken menu links are gone, your phone and the showcase
are in. What remains is genuinely stage-2 — making it look like the brief, and the
few empty pages (including the dedicated self-correction page) that need real
content built into them.

**2026-07-22 — it worked, and I've read the pages to prove it.** After the switch
was flipped, I rebuilt the four main content pages one at a time and then actually
opened each one's content to check. Three of them — the About page, the
Capabilities page, and the Multi-Agent Review Council page — are now built and
carry the self-correction story in their own words: "We caught our own platform
generating invented details on leopardessconsulting.co.uk. Our verification system
flagged it; we corrected it…". That's the exact thing the guard had been silently
stripping, now present on the page, checked by reading it rather than trusting a
flag. The homepage has the story in it too, but it's deliberately still waiting on
one thing — the "our work" showcase strip that needs the real list of our own
sites, which is one of the two bits I flagged earlier as correctly asking for real
data rather than inventing it. And across all of this, not a single new block from
the guard. So the original problem — our own checker refusing to let this site
tell the story it was built to tell — is genuinely fixed and live, and I've closed
that bug.

Three things remain before the site is truly done, and they're all separate from
what we just fixed: (1) the dedicated self-correction page, plus a couple of
others, came out of the build with no sections at all — that's a different,
planning-stage gap I haven't dug into yet; (2) the two "give me real data" asks —
the list of our sites for the homepage strip, and a real business email; and
(3) the hosting cut-over so the pages actually reach a browser, which is your step.
None of those is a nasty surprise; they're the remaining last-mile items now that
the story itself flows.

**2026-07-24 — the components are built, tweaked to your notes, and live across the
site.** This is the bit you originally asked for, delivered. The carousel you
approved got your three tweaks: the arrows now float over the images (not the
text), clicking anywhere on a card takes you through (not just the small link),
and the dead space under the text is gone. It also now sits still by default —
the movement you liked is there as an opt-in per carousel, but this one waits for
you to click.

Then the roll-out: each main page now has one of the new interactive pieces, so
the pages differ from each other the way the big consultancy sites do. The
homepage has a dark band of three real numbers that count up as you scroll to
them — 97% of a dead site's feed restored, 11 live sites, under a day to a
working site — all true, all checked. The capabilities page has the card
carousel. The about page pairs a line illustration with "review first, ship
second, correct openly". The review-council page has a swipeable strip of
what-independent-review-actually-looks-like cards (deliberately no numbers that
could go stale). The fine-tuning page has hover-reveal picture cards linking
onwards. I checked every page live after it deployed: all up, all rendering
cleanly, nothing half-baked.

Along the way we also found and worked around a real platform quirk (component
scripts were being published but never actually loaded on the page — now
documented for everyone), and everything — components, their behaviour, and the
exact content placed on each page — is version-controlled, so future tweaks are
small edits rather than archaeology.

What's still open, all smaller than what just shipped: the three empty pages
(the self-correction story's dedicated page chief among them), the contact page
getting a proper contact-details block so your phone number actually shows, one
stray link on the council page, and — if you want it — letting the site planner
choose these new components automatically for future pages, which I've built the
foundations for but not yet exercised.

---

**Saturday 25 July — all five pages rebuilt through the pipeline, and two useful
things fell out of it.**

You asked for the feature spec and to carry on with the site. Both are done, and
the second turned out to be a better story than expected.

All five main pages have now been rebuilt *by the platform itself* rather than by
me placing things by hand, and every one of them came back carrying its
interactive component automatically — the carousel on capabilities, the people
block on about, the swipeable strip on the council page, the hover cards on
fine-tuning, the counting numbers band on the homepage. That matters more than it
sounds: it means the components live in the site's own plan now, so the next
rebuild keeps them. Before this week, a rebuild would have wiped them.

**The rebuild queue kept eating my requests, and I had the reason wrong.** I told
you earlier that the likely cause was the system only letting one page build per
site at a time. That was wrong, and I want to correct it plainly. There is a
housekeeping job that tidies away build requests which have sat untouched for
48 hours. It decides "untouched for 48 hours" by looking at **how old the request
row is**, not how long it has actually been waiting. So when I reused an old
request to ask for a rebuild, the tidy-up job saw a five-day-old row and binned
it — while the request itself was minutes old. It even stamps the row with
"stale: triaged 48h+", which is how I eventually caught it: **the answer was
written on the thing I'd been staring at all along.** I'd been trimming that
field in my queries to keep the output readable, and trimmed off the explanation.
Logged as a wrong call, because the lesson is cheap and I paid full price for it.

It then did it again while I was writing the spec — parked a request twelve
minutes after I made it, labelled "48h+". Irritating, but it's the best possible
evidence, so it's now in the bug file. The workaround is simple: create a new
request instead of reusing an old one. That's what I did, and it works.

**Two pages the brief asked for were never in the site's plan at all.** The
self-correction story — the leopardess example, which your brief calls the
differentiating trust story — and the decision-record index page. Both have sat
empty since the site was created, not because anything failed, but because
nothing ever told the builder what sections those pages should contain. The
builder actually handled this well: it looked, found nothing to build, said so,
and stopped in 38 seconds without spending anything on writing. I've now given
the self-correction page a proper five-section shape and it's building.

**One real visible bug, caught by the fact-checking gate you asked for.** The
homepage's portfolio panel had a link reading "Visit Site \2192" — a
copy-and-paste error where a symbol code meant for stylesheets ended up in
readable text, so visitors saw the raw code instead of an arrow. What's pleasing
is *how* it surfaced: the new evidence checker flagged "2192" as an unregistered
number, which is technically the wrong reason — it's not a statistic — but it
stopped the page and made me look, and there was a genuine defect there. Fixed
and swept: it appeared exactly once across the whole fleet.

**Still open, and I want your call on one of them.** The em dashes are still
there — fewer than before, but present — and I've now tried tightening the
writer's instructions twice. The honest options are a third attempt with a
worked before-and-after example, or a mechanical pass that simply removes them
after the writing is done. The second definitely works and costs nothing to run;
the first is more elegant if it lands. Say which you'd prefer.

Beyond that: the contact-details block so your phone number shows, the
decision-record page, and a set of dead in-page links on two pages (the writer
creates "jump to this section" links but nothing creates the sections to jump
to). That last one is a detector gap rather than a one-off, so it belongs with
the existing link-checking work rather than being patched by hand here.

---

**Saturday 25 July, later — your phone number is on the site, and finding out why
it wasn't turned into the most useful bug of the day.**

The short version: the number was never lost. It was written into the site's
record in one place, and the block that displays contact details reads a
different place. Those two places differ by one level of nesting — the display
block looks for `email` and `phone` at the top of the record, the thing that
creates the record files them one level down under `contact`. So the lookup came
back empty, and because a missing email address is treated as "don't show this
block at all", the whole contact-details section was quietly left out of the
build. Three times. No error, no warning, nothing in the log naming it.

I checked whether that was just us, and it isn't. Across our thirteen live sites,
the five that have a contact-details block are **exactly** the five whose record
happens to have those fields at the top level. The eight that don't, don't. No
exceptions in either direction. The default way a new site gets created produces
the broken shape, which means **every new site is born without a
contact-details block** and nobody would notice unless they went looking for it.
That's `bugs_open/072`, with the fix options written up.

For your site I've added the fields where the block actually looks, kept the
existing ones, and rebuilt. Your number is now live on the contact page as a
tappable phone link.

**On the broken links, I have a correction to make about my own work.** I told you
earlier the internal links were fixed and that my check confirmed zero remaining.
That was wrong, and the reason is worth knowing. I found the links with a search
pattern, repaired them with the same pattern, then re-ran the same pattern to
confirm — and it came back clean. But the pattern had a blind spot: it couldn't
see links that jump to a section, like `/capabilities#approach`. So it missed
twenty-one of them, my repair missed the same twenty-one, and my confirmation
missed them too. All three agreed with each other and all three were wrong. What
found them was actually visiting the pages and following every link.

The lesson I've written down is simple: **a check that shares its logic with the
fix can't test the fix, it can only agree with it.** That's twice today — the
other time, my queries were trimming a text field to keep the output tidy, and
the trimmed-off part named the exact cause I then spent two wrong theories
guessing at. Both are now in the workstream's command book so the next session
doesn't repeat them.

The links are now genuinely fixed on six of seven pages, verified by crawling the
live site rather than by asking the database. The seventh, capabilities, is
republishing as I write, and I'm running a final sweep that visits every page and
tries every link three times before calling it broken — because I also learned
today that hammering the site with rapid requests makes it refuse connections,
which looks exactly like a broken link if you only ask once.

---

**Sunday 26 July — the site is verified sound, and you've green-lit the chart.**

Wrapping up with a proper handoff so this can continue in a fresh chat.
`HANDOFF_2026-07-26_continue_here.md` is the new starting point; the old one is
marked superseded so nobody cold-starts from a description of a site that wasn't
live yet.

The link fix is now genuinely confirmed: every internal link on all seven pages,
followed on the live site, three attempts each — **43 targets, none broken.** It
had been twenty-one of twenty-two. The last page published itself while I was
writing notes, via the queued request rather than either of the two direct
attempts I'd fired impatiently. Waiting was the right answer for the last half
hour and I got there late.

**On the decision-record page you asked about.** It's the page your brief calls for
to back up the strongest claim on the site — that every substantial change is
independently reviewed before it ships, with a real record. Right now the site
asserts that and offers nothing to look at, which is the one kind of gap this site
can't afford.

What it would show is real, and I checked rather than assumed: **156 review
decisions** recorded since 17 July, **41 changes** whose commit carries the
reviewer's verdict as a signed-off line, and a spread that includes **seven sent
back for revision, two vetoed outright, one escalated and one thrown out as
invalid** alongside nine approvals. That mix is the point. A review record showing
only approvals reads as marketing; one showing what got rejected reads as a review
record. It's the same argument as the leopardess self-correction story, but with
volume behind it.

The reason I haven't built it is that it isn't a technical decision. Publishing it
means putting our internal objection text and the reviewers' own criticisms of our
work on a public website. That's yours to weigh — you may want it whole, or
summarised, or with the objection text redacted to verdicts and dates only. Any of
those three works technically. Until you decide, the honest options are to build it
or to soften the sentence that currently promises it.

**The chart component: noted as green-lit and it's now the top build task in the
handoff.** Two things I've locked in so it doesn't drift. It gets built once, in
the framework, so leopardess can use it too rather than us building a second one.
And the numbers come from the verified-facts register, not from the model — the
model can choose labels and ordering, but it cannot supply a figure. A chart is the
most convincing place a number can sit, so on a site whose whole pitch is that
claims are sourced, an invented number in a chart would be the worst possible
failure. There's already a scoped design for this in the leopardess work, so the
next session should read that first and coordinate rather than start fresh.

---

## 2026-07-26 (later) — the chart is built and live, and it found three things on the way

**The chart is on the home page.** It shows the relojistas feed going from nothing
working to about 97% working within a day of relaunch, plus the news pipeline's
collected-versus-assessed counts and — your call, and you said yes — what our own
review council actually decides: 108 rounds sent back for revision, 37 approved,
9 rejected outright. I've checked the live page rather than the database: every
figure on it matches the register exactly, and the bars are drawn from those same
numbers rather than from anything anyone typed twice.

**It turned out not to need a code release at all.** The framework can already
feed a component straight from the verified-facts register, and whatever comes
from there overwrites whatever the model wrote. So "the model may not supply a
figure" isn't a rule we're trusting it to follow — there is no route by which a
model-written number can reach a bar. A chart definition names which facts to
draw and never repeats a value, so an unverified number has nowhere to live. That
also means it went live today rather than waiting on the next image build.

**Leopardess gets this free.** It's one component, in the framework, and their
side now needs nothing but a short list of which of their existing facts to plot.
I've written that into their own notes, along with two corrections: the charting
library their plan agonises over isn't actually in our codebase, and a chart
drawing routine already exists from another piece of work.

### Three things the build turned up, and one is on me

**A gap I'd like to fix later, not now.** I wanted the home page and the
capabilities page to show *different* charts. They can't yet: nothing tells a
section which page it's being drawn on, so every chart appears on every page
carrying the section. Rather than publish the same three charts twice, the
section is on the home page only. The "which page" information is already sitting
in the data, unused and correct, so this switches on with a one-line fix when we
next build an image. Filed as a bug so it doesn't get lost.

**A near miss on your dark site.** I'd styled the chart card with a colour
variable that reads as standard — and which no theme actually defines. On
leopardess that would have rendered a white card on a black page. Caught by asking
the database which variables exist rather than copying what the other components
do. Two of them have been quietly relying on the same non-existent names for
weeks; harmless there because it's spacing, not colour.

**This one is my fault and worth stating plainly.** Rebuilding the home page to
add the chart *reintroduced six broken links* on a page that was verified clean
yesterday. Four of them point at pages that don't exist at all. They came from the
model rewriting a card grid, the checker spotted every one and let the page deploy
anyway because it's classed as a warning, and the repair we did yesterday was
per-page — so a rebuild simply undid it. I've fixed the six by hand again, but the
honest read is that "the site is link-sound" describes an artefact, not the site:
it expires the next time any page is rebuilt. That belongs with the broken-link
case we already filed, and I've added the evidence there.

### The em-dash question, now measured

These were the first pages written since the third attempt at the voice fix. On
the home page the em dashes dropped by nearly half; on the capabilities page they
didn't move at all. Two components account for most of what's left. So: it half
worked, and I'd rather not spend a fourth round on the site-wide instruction.
When you want it finished, the cheaper options are a mechanical pass over the
finished text, or fixing the two specific components that produce most of them.
Your call; nothing is blocked on it.

## 2026-07-27 (afternoon) — the em-dash question, answered properly, and a bug I sized wrong

### The em-dash count was measuring two different things at once

You asked me to pick between the mechanical pass and fixing the two components.
Before recommending I re-measured, and the measurement changed the question, so
here is the honest version.

The count I gave you came from the finished page HTML. That HTML contains two
different kinds of em dash: the ones the writing model produced, and ones **typed
into the component template itself** and therefore reprinted on every render. No
instruction to the writer can ever move the second kind. Nor could the mechanical
pass I offered you, because that pass would work on the text, and these are not in
the text — they are in the shell the text is poured into.

Split properly, the site has **66 em dashes: 43 written, 23 baked into templates**.
That changes both halves of what I told you yesterday. The capabilities page,
which I reported as "no improvement at all", carries **four** template ones and only
**two** written ones — so the writer's output there may well have improved and I
had no way of seeing it. And one of the two components I named as the main
offenders, the card carousel, turns out to contribute nothing from the writer at
all. Its four are all template.

**So my recommendation is neither of the two options I gave you.** Leave the writer
alone — it is working, the home page halved. Do the per-component fix, but on the
templates rather than the copy: three components hold 21 of the 23, and two of
those three are *generated* tool pages, which means the same em dashes will be
reprinted into the next tool we generate unless the generator's own instructions
are fixed. That is the cheapest durable win here, and it is a small job.

One more thing, since I am being exact about it: **two of the three em dashes the
chart section added to the home page are mine**, not the model's. One is a comment
I left at the top of the chart's stylesheet, which — I had not realised — is
shipped to the browser on every render rather than stripped. The other is in a
caption I wrote.

### The page-identity bug: I said one line, it was three

Yesterday I filed the bug that stops a component knowing which page it is on, and
wrote in it, twice, that the fix was one line. That was wrong. Following the value
properly today, it is dropped at **three** separate points between where it enters
and where a component could read it. Each of the three looks perfectly reasonable
in isolation, which is why one pass found only one of them. Had I shipped just the
line I filed, nothing visible would have changed and the next person would
reasonably have concluded the diagnosis was bad.

It is fixed now, with a test that follows the value the whole way rather than
checking each piece — and I proved the test works by deleting each half of the fix
in turn and watching it fail differently each time.

The review council sent it back once, and the objection was a good one: *you have
fixed this one field, but not the mechanism that dropped it — the next field
someone adds will vanish the same way, silently.* They were right, and checking
turned out to make it concrete: **three other fields are in exactly that state
today**. So the fix now also carries a check that fails the moment anyone adds a
field the templates can see but the pipeline cannot deliver, and the wider
mechanism is filed as its own case rather than quietly bundled in.

None of this is live yet. Go code only takes effect when a new image is built and
rolled out, and I have not done that — it would also switch on a good deal of other
threads' finished work at the same time, which is your call rather than mine.

### The decision-record page

Separately below/in the handoff, but the short version: nothing technical is
blocking it. What is blocking it is a judgement only you can make, and I have
written out both what the page would contain and what the most quotable line
against us would be, so you can decide with the actual numbers in front of you.

---

## 2026-07-27 (later) — you looked at it on your phone and you were right about all of it

You said: nothing like the brief on mobile, no graph, one carousel and it doesn't
load its images, grey text you can't read on white, not enough imagery, not
exciting or professional yet. I went and measured rather than argued, and every
one of those is a real fault with a specific cause. Here is what was actually
wrong.

**The site was, in large part, invisible rather than badly designed.** I rendered
each page in a real browser and had it work out, for every piece of text on the
screen, what colour it was and what colour was behind it. That found **101 places
where the text could not be read against its own background** across five pages.
Not "a bit low contrast" — the worst were around 1.1 to 1.2 to 1, where 4.5 is the
readable threshold. Every card heading on the site. Every one of those little
uppercase labels above a heading. And the whole chart section, which is why you
couldn't see the graph: the bars were drawing correctly and their labels and
numbers were white text on a white card.

Your example — "every decision leaves a record" on the council page — measured
1.21 to 1. You were not being fussy; it is genuinely not readable.

**Why it happened, in plain terms.** A site's colours come from two places that
were never checked against each other. Our own palette for this site defines eight
colours. The page template it plugs into expects seventeen, and quietly fills in
the missing nine with its own defaults — which are all designed for a *white*
site. So on our near-black site, cards were being painted white while the text on
them stayed the pale colour chosen for a dark page. White on white. Nothing errors,
nothing warns, and if you read the finished stylesheet the white looks like
somebody's decision.

Two more of the same shape. Our "primary" brand colour was a navy so dark it was
almost the same as the background — and the component library uses that colour for
*text* in fifty-odd places, so all of it vanished, while the buttons filled with it
still looked fine. And the code that picks a default text colour was choosing the
palette's *dimmest* grey for body copy and headings alike, with the proper text
colour sitting unused right next to it. That last one is the "grey text" you saw,
and it affects every dark site we run, not just this one.

**The imagery was worse than "not enough".** It turns out we had already generated
twenty-one images for this site — six page headers, fifteen line-drawn icons, a
brand illustration — all in the style the brief asks for, all uploaded, all live on
the server. The pages referenced **three** of them. The carousel you saw was
pointing at a filename that does not exist on any site we run, which is why it
showed broken-image icons with the alt text next to them. Six more images on the
fine-tuning page pointed at a folder we have never used.

The reason the good images never got onto the pages is almost funny: when an image
finishes generating, the platform files a job that says "re-render this page now
the picture has landed". Five of those were filed for this site a week ago and all
five are still sitting in a queue marked "needs human review". Across all our
sites, **fourteen of those twenty-eight jobs are parked the same way**. So the
imagery stage of our pipeline is roughly a coin flip, everywhere.

**What is fixed and live right now.** All the imagery: six page headers including
proper background images on the capabilities, about and contact pages, the four
carousel cards, thirty card icons that are now real line drawings instead of
emoji, and the about-page illustration. Every page re-rendered and republished.
I re-measured afterwards: **zero broken images**, down from what the browser
initially reported as forty-one — of which, I should say, only six were genuinely
broken and the other thirty-five were our own server briefly refusing a burst of
requests. I built the re-check into the tool so it can't cry wolf like that again.

**What is fixed but not live.** The colour fix itself. The stylesheet is written,
tested and verified — I re-ran the measurement against it and the 101 failures drop
to 1, and that last one is a pre-existing problem on twelve of our sites that a
separate change fixes. But publishing a stylesheet is the one step I could not
complete: the only route the platform has to write that file also re-runs the AI
design pass, which re-rolls the colours it is meant to be fixing. I wrote a direct
publishing route instead, and my own permissions blocked me from running it. **It
needs one command from you, and the site changes the moment it runs.**

**On your bigger question — why didn't we catch this ourselves.** Three answers,
and the third is uncomfortable.

We run about fifty automated checks on a site. Every single one of them examines an
*ingredient* — a template, a colour list, a link, an image record. None of them
looks at the finished page. The three problems above are all invisible from the
ingredient side, because each ingredient is individually correct and only the
combination fails. So this was not an oversight in any one check; it is a missing
vantage point. I have written the missing piece — it renders the page and measures
it, the same way I found all this — and it is committed and working. Wiring it into
the build itself is now written up as the next piece of work.

The second answer: one check that *should* have caught the broken images says, in
its own source comments, that the half which would have caught them was deferred
for later and never done.

The third: **the platform had already found part of this and told us, and nobody
was listening.** On the 24th, our own brief-fidelity audit filed three findings
against this site. One of them reads: "Only 2 of 27 components contain images —
raising serious doubt that the illustration system is meaningfully present." That
is your complaint, in our own words, three days before you made it. It is still
sitting at status "detected". They are the only three findings of that type in the
entire database, and nothing anywhere reads them. Building more detectors is not
the fix if that is what happens to their output — so I have written that up as the
thing to do *first*.

**What I would do next, in order.** Publish the stylesheet (one command, needs
you). Then drain or stop filing the parked re-render jobs, because that is what is
starving every site of its own imagery. Then wire the page-measuring check into the
build. The deeper "make it exciting" work — more varied section types, richer use
of the illustrations we already have — is worth doing but it sits on top of these,
and there is no point styling a page whose text cannot be read.

---

**2026-07-28, afternoon (Claude).** Good news first: the spending cap you raised this
morning worked. The capabilities page rebuilt successfully and its chart section is
correct — two charts there, one on the home page, exactly as designed. And both missing
tool images have now been generated: the cost calculator page already shows its new
header picture, and the selector tool's should follow shortly (its re-render had
correctly refused to run while the picture didn't exist; I've set it going again now it
does).

The bad news is what I found while checking the links. That capabilities rebuild wrote
nine links to pages that don't exist, plus the four broken carousel pictures you
spotted. Here's the part worth your attention: the platform's link-repair safety net —
the one we built and closed as proven a couple of days ago — actually caught ALL nine
bad links during the build and fixed them. Then the very next step threw the fixed
version away and saved the broken one. The repair has never once worked on a real page
build; the test that "proved" it took a shortcut route that skipped the saving step.
I've reopened that bug with the full evidence and the fix options, corrected the other
documents that believed the safety net was working, and written up the lesson: check
what was actually saved, not what the fixer says it did.

One more thing to know: our own AI provider (Anthropic) has now hit ITS monthly spend
cap, until the 1st of August. The site writers and image generation use Google, so this
site is unaffected — but the review councils and the diagnosis loop are down until then.

Also confirmed: the `/assets/illustrations/` folder those carousel pictures point at has
never existed on any site we run — the writer simply invented it. So that regression
joins the invented-links problem as one and the same defect: the writer isn't told what
exists, and (we now know) nothing downstream actually stops the results shipping.

---

**2026-07-28, evening.** The spending cap on the Claude models came back mid-afternoon,
so the things we'd parked "until August" got done tonight instead.

Three outcomes. First, the model-selector tool now has its picture. It turned out the
image had been sitting published on the site all along — the page just never pointed at
it. We added the picture to the page directly and pushed that one page live; it's
showing now, and it's the right image for the job (a three-way fork illustration, no
lettering).

Second, the capabilities page is clean again. All thirteen broken references — the nine
made-up links and the four made-up illustration files — came from one bad rebuild this
morning, and the platform keeps a copy of what each section looked like before it gets
overwritten. So rather than repairing thirteen things we put back the two sections the
bad build had damaged, exactly as they were the day before, and re-rendered the page
without involving the writing AI at all. The live page now has no broken links, real
photographs on the carousel, and both charts still in place. One wart we inherited
rather than caused: several of the restored links point at spots *within* the page that
don't technically exist, so clicking them opens the page but doesn't scroll anywhere.
Cosmetic, noted for later.

Third, we asked the independent diagnosis system to check this morning's big claim —
that the platform's link-repair step does its work and then the work gets thrown away
before saving. Its answer was "could not verify" rather than "wrong": it agreed with
how the code is wired but couldn't see a complete example from where it looked. In
chasing the one new lead it gave us, we found a third site with the same signature —
a page whose repair has been recomputed as recently as this afternoon while the stored
page hasn't changed in a week. So the claim stands, now with three sites behind it,
and the real fix (make the save step apply the repair) is the biggest lever on the
board for whoever picks it up next.

One process note: we cleaned the capabilities page ourselves tonight, which means it's
no longer an example of the bug — the bug file now points at the other two sites for
anyone who needs to see it live.

**Later, same evening — the 079 fix is designed and handed off, not built.** You asked me
to go ahead with it and then to write it up so a fresh session can carry it. Both done.

The short version of the fix: the repair currently happens at the inspection stage, and
the page gets saved from a different copy that never sees it. So rather than trying to
keep two copies in step, we move the repair to the moment of saving — the last point
everything passes through. Whatever route a page took to get there, it cannot be stored
with a dead link. The reason that beats the alternative is concrete: one of the routes
that saves pages doesn't run the inspection step at all, so fixing the inspector could
never have covered it.

Two things make this cheaper than it sounds. Every piece already exists — the repair
function, the list of real page addresses, the logging — and they're all in the same
place in the code as the save step, so it's wiring rather than building. And we can prove
it works without spending anything on AI: there's a page on gamesdesign.co.uk whose
stored copy has carried a broken link since the 21st, and re-saving it after the fix
either cleans it or doesn't. That's the test.

There's also a safety catch built in: a switch that turns the repair off from the
database if it ever misbehaves, without waiting for a new software build.

Handoff doc is in the repo with everything needed to start cold; the bug file and my
notes both point at it.
