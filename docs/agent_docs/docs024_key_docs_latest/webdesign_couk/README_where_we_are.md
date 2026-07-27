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
