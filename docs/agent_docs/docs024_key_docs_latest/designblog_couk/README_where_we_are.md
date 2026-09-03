# README — where we are (designblog.co.uk)

*Append-only, newest at the bottom. Plain prose for the owner.*

## 2026-09-02 — the lane opens on your critique

You reviewed designblog.co.uk the day after it went live and gave a critique
covering the design sameness, the missing tools link in the nav, four section
pages that talk about what they'll contain instead of containing it, copy that
sounds like AI, too few images, and an overall shortfall on the delight the
brief promised.

Before passing any of that on, I checked every point against the live site.
Every point is right. The glossary has no glossary entries, the inspiration page
has no showcases, the feed has no feed items, and the studios directory has no
studios — all four are essentially essays about the page they were meant to be.
Each page carries exactly one image. The nav really has no tools link even
though the contrast checker tool is live. The two sentences you quoted are on
the page word for word.

The pattern behind the four empty pages looks like one problem, not four: the
pipeline wrote the copy for each section from the brief, but nothing ever
produced the actual items (terms, showcases, entries, studios). That question —
"was anything ever going to fill these?" — has gone to the portfolio positioning
thread that built the site, and to the experience loop thread, whose new checks
exist for exactly this "page promises something it doesn't deliver" problem.

The design-sameness point has gone to all five threads you named (components,
experience loop, theme kits, site design planner, vigilant designer), each with
your "best in class, more impactful, sites should differ" message, and the copy
points to the copy quality two stage thread with your before/after examples.
I'm tracking who has acknowledged receipt and will chase any that haven't.

## 2026-09-02, same evening — six of the seven threads have answered, with measurements

Every thread except copy quality two stage has come back, and none of them
guessed — they all measured. The answers agree with each other, and the story
they tell is clean: **the machinery to make the sites different mostly exists
and is not being chosen**, and **the empty pages have no producer at all**.

On the sameness. Three separate threads measured three separate causes. The
header and footer are literally hardcoded — 36 of 37 sites render the same
pair, while ten alternative header designs sit in the library unused, because
the code picks the same one by name and only 6 of 40 style collections override
it. The page layout came out identical on all three new sites because the
layout library has exactly one professional editorial layout and no "content
hub with tools" shape at all — so three different briefs all landed on the same
answer honestly (now bugs_open/445). And about ten components carry roughly
four fifths of every slot on the estate, while two thirds of the component
library is unused or single-site. Nobody needs to build many new components;
something needs to start choosing differently, plus one new layout archetype
needs designing.

On the images: it's worse than you said. The one image per page is the header
logo. The content of designblog carries zero images across all fifty of its
component slots — the three new sites are the most image-poor on the estate.
Two already-decided things would help immediately and are simply not switched
on: the illustrated text block (still losing to the plain one everywhere but
one site) and the card-grid carousel you ruled default-on (on for 1 instance
of 49). The bigger half is supply — the machinery for placing images works,
but there are almost no images to place. That's now with the editorial design
uplift thread.

On the empty pages: the builder thread confirmed it on advertise too and filed
it as bugs_open/444. The feed pages have a working news mechanism behind them —
the four new sites just have no news sources wired (that's with the feed lane
now). The studios directory needs its kind added to the existing directory
machinery. But the glossary and inspiration pages have no item producer
anywhere in the estate — nothing was ever going to fill them. That's the
genuine gap.

Why nothing caught it: both of the experience loop's checks are blind to an
empty index by construction — the rule asks "does this list the right things?"
and an index that lists nothing never reaches the question. They've taken
building the empty-index rule and will run it over these four pages. Separately,
as of today the content quality auditor samples every page and now asks exactly
your question — "does an index page actually list its own items, or write ABOUT
itself instead" — recording verdicts for your approval rather than rewriting.

Decisions that are yours, gathered so far: whether theme kits' page_archetypes
mechanism gets applied and rolled (it's committed, inert, waiting on you);
whether the pre-flight gate (component-overlap + image-count checks on each
remake before it ships) gets assigned; and whether you want the cheap per-site
fixes now (tools nav link via a tools hub page; a pinned alternative header)
or only the class fixes. Copy answers still to come.

## 2026-09-02, late — one correction to what I told you above, and the last two answers

A correction first. I said the content of designblog carries zero images.
That's wrong, and another thread caught it: the big page-top images (heroes)
on this estate are painted as CSS backgrounds, which the counting we did
cannot see. Every designblog page does serve a real hero photo. What's
actually true — and it still supports your complaint: the site has heroes,
four small icons and a logo, and nothing else. No illustrations, no
infographics, nothing inside the text. And the six pages share only three
distinct hero photos — the feed page and the contrast tool reuse the
homepage's. I've recorded the wrong call in the fleet log with the check that
would have caught it.

The deeper imagery finding: across the whole estate, in all the time it's run,
planners have requested exactly ONE infographic, ever. The generating and
placing machinery for illustrations and infographics exists, is live, and has
been traced end to end — nothing ever asks for them. It's a vocabulary gap in
the planner, not broken machinery, and there's a proven precedent that fixing
the vocabulary works: when the planner's menu was taught the word for an
illustrated section in late August, planners started choosing it within days.
Also: every article page on the estate is one slab of prose plus trimmings, so
there's currently no structure to put pictures between — that's bugs_open/114.

And the copy answer arrived. Your build DID go through the copy gates, which
is the interesting part. The "starting point, not the final word" sentence was
seen and targeted by the repair step — the repair model just failed to answer
for that sentence, so the original shipped (that's now a counted failure class
with a candidate fix: re-ask once). The "says so plainly" wording is on a
banned-words list that currently only detects overnight — it has no repairer
yet. The "before your users have to" shape and the essay-in-a-button are
already-ruled classes queued in their lane. So the copy story is: gates exist
and ran; two repair arms are missing and one repair call failed silently.

All seven threads you named (plus three more the answers pointed at) have now
confirmed receipt of your best-in-class message, and every one answered with
measurements rather than promises.

## 2026-09-02, later still — a correction to the imagery story: it's not a missing word, it's an instruction

The thread that owns in-body imagery read the live planner prompt and
corrected the picture I gave you above. The planner's prompt already knows the
words "illustration" and "infographic" — the vocabulary isn't missing. What
keeps the numbers near zero is the prompt itself: it literally instructs "use
sparingly — most plans will have zero section-scope entries", its stated
minimum is logo-plus-heroes only, and its worked example shows only icons and
no infographic at all. Models copy the example's shape, so one infographic
ever planned is the planner doing as told, not failing.

That means the fix is a deliberate edit to the planner's instructions and
example — which costs real generated images on every future build, so it's a
decision, not a bug fix. Two strings attached: the one-image-per-entry rule
must survive the edit (or image generation produces botched collages), and
article pages currently have no structure to hold in-text images anyway — so
this edit alone gets pictures onto landing pages, and putting images inside
articles additionally needs the article-structure work (bugs_open/114). Those
are two separate asks and I'll keep them separate.

## 2026-09-02, night — you said go, and the planner prompt change is live

You ruled "go ahead with the planner prompt and exemplar changes", and it's
done — live in the running system as of just before 20:00 tonight.

What changed: the site planner's instructions no longer tell it to be sparing
with pictures. The "most plans will have zero" line is gone; in its place, the
prompt says content-carrying imagery is expected, sets a floor of at least one
illustration or infographic on the home page, and — because examples teach
these models more than instructions do — both of the prompt's examples now
show an illustration and an infographic being requested, written so the
pictures carry no text (image models make a mess of lettering; headings stay
in HTML). Two guard-rails from the earlier analysis were kept deliberately:
the one-image-per-request rule survives untouched (without it we'd get botched
multi-panel collages), and article pages are exempt — they have no structure
to hold pictures yet, so that stays a separate piece of work rather than
something forced.

It went through the full discipline: the exact live text read first, the
change rehearsed against the real database and rolled back before being
applied for real, a snapshot taken, a rollback file written and proven to
restore the original to the byte, the change submitted to the review council
(verdict pending — I'll read it and act if they object), and the result
re-checked in the live system afterwards.

Proof arrives with the next site build: the first new plan should ask for
illustrations and infographics on its own. The portfolio thread will check the
first one that runs. Cost-wise this means more generated images on every
future build, which you accepted when you ruled.

One more thing from the same evening: the owner critique of gamedesign.uk
turned out to be the same problem classes, that thread has joined the same
routes, and their site will re-plan under the new prompt — so your two reviews
are now feeding one set of fixes.

## 2026-09-03 — you ruled on everything, and here is what happened to each ruling

All eight answers landed and were carried out or routed the same morning. The
short version of where each one stands:

Page archetypes were already live — you'd told the theme kits thread directly.
One number to keep next to it: 94.4% of live pages get their structure from
the planner, not the fallback the archetypes replace, so this mechanism
governs roughly one page in eighteen. Real, but not the whole sameness fix.

The two switches (illustrated text block, card carousel) are accepted and
first in the vigilant designer's queue, with a proper before-and-after
measurement already prepared — including a page that already has the carousel
on, so "on" has a known signature to compare against.

Bug 445's thread confirmed it owes both the detector and the new layout
design, in that order, for a good reason: the measurement tells them what the
layout should be, rather than designing from three examples.

The glossary decision ("both") and the featured-content registration are
recorded in the bug that owns them; briefs stop promising glossaries and
inspiration pages until the producer exists.

The feed ruling is done: the page keeps its type and will fill from child
pages; the news-source work re-scoped itself around that within the hour.

The tools link: done through the framework. The hidden cause was tidy — the
four tool pages were always marked for the nav, but tool-typed pages are
deliberately barred from it; nothing had ever planned a hub. There's now a
Tools hub page in the site's plan with its build queued, and the nav link
appears when the page deploys and the chrome next re-renders.

The header pin I held back, and I want you to know why rather than have it
look like foot-dragging: the thread that owns that mechanism corrected its
own advice an hour before I would have pinned. The alternative headers are
old, never-rendered components that need about a dozen pieces of site data
designblog doesn't carry — the likeliest result was a visibly broken header
on the site you'd just critiqued. Worse, all three candidates are wrong in
kind for a blog: a search form with no search behind it, cart machinery, tool
labels. A genuinely different designblog header needs the chrome work, not a
pin of a stale part. The pin experiment still matters for the estate and
moves to the next remake instead.

And your seotools critique: right, and already in hand before it reached me.
The morning's automated check had caught all seven fake tool pages at 07:41;
the diagnosis existed (bug 450 — the plan named tool pages before any tools
existed, and a generic builder filled them with prose); you'd ruled "build
the tools" and all seven real tools were built by 09:54. They aren't visible
yet — the rebuilt pages are queued behind a stale deploy marker that reads
"deployed" while serving the old prose — and a watcher reports each page the
moment it actually serves its tool. Two lasting lessons went into the errors
file: the repair lands beside the promising prose rather than replacing it,
and a "deployed" status only records that a deploy once happened.
