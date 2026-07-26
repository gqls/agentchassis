# SUMMARY — webdesign.co.uk is live

**2026-07-26.** Written to be read aloud.

## What we're trying to do

The owner runs two hand-built static sites that overlap in purpose and split his
audience: `website-design.com`, the larger one, with fifty-five small
browser-based tools and twenty-three articles in a stark black-and-electric-blue
Swiss style; and `websitedesign.com`, much smaller, with ten more tools and ten
guides, whose homepage wears a warm sage-and-terracotta look that none of its own
sub-pages ever received.

The job was to merge them into one home at `webdesign.co.uk` — a premium generic
domain he already owns — in the warm design, carrying every feature except one:
`websitedesign.com`'s client-side LLM builder, which downloads an 800MB model
into the browser and which he asked to skip. And to do it as a
*chassis-managed* site, so the platform owns the design, the navigation and the
publishing from here on rather than the files being hand-maintained forever.

## Where we've come from

The starting position was worse than it looked. The two sites shared no code at
all — not a stylesheet, not a script, not an image. The larger site's design
tokens existed but were routinely ignored: one colour appeared as a hardcoded
literal a hundred and eighty-three times. The smaller site was half-migrated, its
warm palette present on exactly one page and a legacy dark-terminal skin on the
other twenty. Three things were simply broken: a tool that had never worked
because it loaded a file that does not exist, a guide page that was a
byte-for-byte duplicate of another with the wrong title on it, and fourteen
finished pages that were live but linked from nothing.

Choosing the warm design meant applying the *smaller* site's look to the
*larger* site's content — eighty-six pages to reskin rather than twenty. That
was the owner's call and the right one aesthetically; the cost was the shape of
the whole engineering problem.

## What we've done

We built a converter that reads both source trees and produces finished pages in
the new design, and refuses to finish if any source page is neither converted nor
explicitly listed as dropped with a reason. It strips each page's copy-pasted
navigation, keeps its own styles and scripts, and rewrites colours
property-by-property — because the same literal means "ink" in one place and "a
dark panel" in another, and a blind find-and-replace would have got that wrong
everywhere.

We walked the platform's onboarding through one step at a time with everything
else held back, reading what each step decided before allowing the next. That is
how the design got pinned before the design agent's first run — an agent that
invents its own palette whenever it is not given one in a specific structured
form, and then re-invents it on every subsequent run. It is also how a wrong tool
count was caught before it could be written into the home page.

The site went live with ninety-eight pages: sixty-three tools, thirty-one
articles, an about page, two index pages generated from the real page list rather
than hand-maintained, and a home page. Everything runs in the visitor's browser;
nothing is uploaded, nothing is stored, no account is required.

## Where we are now

Live and working. Every page returns 200. The header, navigation and search work.
The tools have their JavaScript. Every colour on the site is one of the owner's,
verified in the published stylesheet rather than in the configuration that was
supposed to produce it.

Three defects found along the way are fixed rather than carried: the dead tool
now works and its dead Copy button does too, the duplicate guide is gone with the
feature it documented, and the fourteen orphaned pages are linked.

Two mistakes of our own are worth stating plainly, because both were caught late
and both are now guarded.

The first: a bug in our converter was silently discarding **every tool's
JavaScript**, because both source sites put their scripts after the main content
and the converter looked only inside it. Sixty tools would have shipped as dead
markup while the build reported ninety-seven pages and zero warnings. Nothing
about the output looked wrong. It is now impossible to repeat — the converter
refuses to build if any page loses a script — and that guard was proved by
deliberately reintroducing the original bug and watching it produce sixty errors.
The wider gap it exposed is filed for the platform: nothing anywhere checks that
a published page's JavaScript actually works.

The second: we reported the final publishing step as blocked by a platform bug.
It was not. The queue takes about twenty minutes to pick up the first job and
then runs at roughly two minutes a page — three and a half hours for
ninety-eight — and we gave up eight minutes before the first one started. The
supporting evidence, that other work was completing concurrently, was an
artefact: that work had been claimed before the page jobs existed and stopped
dead the moment they began, because we had just told the queue to prioritise
pages. We mistook the queue obeying our own instruction for the queue being
broken.

## Where we're going

The remaining work is small and none of it is blocking. About sixteen tools —
the ones that use the camera roll, canvas or clipboard — need clicking through in
a real browser, which no automated check can substitute for. The home page's
hero has just been swapped from a dark full-width banner to the two-column layout
the brief asked for, and that needs confirming live. The verification command in
the port tool is still a stub; its checks currently live as recipes in the
runbook and deserve to be code.

One question for the owner remains open: the two source sites are untouched and
still live, and at some point the duplicate content across three domains needs a
decision — redirect, canonical tags, or leave them.
