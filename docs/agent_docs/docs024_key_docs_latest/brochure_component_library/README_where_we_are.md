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
