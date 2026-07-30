# loancalculator.co.uk — adopting a hand-built site, and what it tells us about the framework

*Written 2026-07-30, at the start of the work. Plain prose, meant to be read aloud.*

## What we're trying to do

We own a small, genuinely useful website — loancalculator.co.uk — that was built by
hand back in March and has been sitting outside the platform ever since. It has
twelve working loan calculators and thirteen plain-English guides to UK borrowing.
Nothing about it is managed: the platform does not know it exists, no checks run
against it, nothing rebuilds it, and the only copy of its "source" is a one-off Go
programme in a folder on this machine.

The job is to bring it inside. Two things have to be true when we finish. The
calculators must still work — all twelve of them, doing the same sums they do now.
And the web addresses must not change; anything anyone has ever linked to or
bookmarked has to keep working, with a forwarding page if it truly cannot.

Then the owner added the part that makes this interesting. He does not just want
this one site rescued. He wants **the framework itself able to adopt sites like
this one** — so anything the platform's own adoption process would have got wrong,
we report, and where we sensibly can, we fix in the platform rather than working
around it. This site becomes the test case that proves the road is open for the
next one.

## Where we've come from

The site was built in March in the way you build something quickly: a Go programme
stamped some data into HTML templates, the result was copied into the shared
deployment repository, and that was that. It was never wired into any of the
machinery the rest of our sites use. The last time anything touched it was the
20th of March.

Meanwhile the platform grew a proper adoption route — you point it at a live
website, it crawls it, works out what the site is, and rebuilds it inside the
system. That route has been used successfully. But it was built for a particular
purpose: taking over a site and *regenerating* it in our own style. That is a very
different thing from what this site needs.

## What we've done

So far, this has been an investigation, and it turned up more than expected.

First, the site was completely dead. Not slow — dead. The domain name resolved, the
security certificate was valid, but no page ever came back; the connection just
hung until it gave up. The cause turned out to be one missing piece of
configuration at Cloudflare: every healthy site on the platform has a small
programme at the network edge that fetches pages from our storage bucket, and this
domain had never been connected to it, so requests were still being sent to a
storage location at Amazon that was switched off long ago. The owner added the
missing route this afternoon and **the site is serving again** — I confirmed it
returns pages normally at 16:11 our time.

Second, the site has real defects of its own that nobody had noticed, because
nobody was looking. Four pages are not really pages at all — they are loose
fragments with no page structure, no styling and no navigation, so they arrive in
the browser as unstyled text. Ten of the twenty-eight pages have no navigation
menu whatsoever, including the flagship loan calculator, so a visitor landing there
has no way to reach anything else. The navigation menu that does exist points at a
calculator page that does not exist, on every page that shows the menu. And one
page loads its stylesheet from the wrong place, so it has been serving unstyled
since March.

Third — and this is the part that matters for the owner's wider question — I read
the platform's adoption process closely, and **it would have damaged this site in
three specific ways.** It rewrites every web address into a different shape, so all
twenty-eight addresses would have changed. It marks every page for regeneration by
a language model, which for twelve hand-written calculators means twelve rewrites
of working, tested arithmetic. And it has a setting that is supposed to control
exactly this — how faithfully to preserve what it finds — which is accepted, and
recorded, and then read by absolutely nothing. I checked: the setting has no
implementation behind it anywhere in the codebase.

## Where we are now

The site is up. Nothing has been changed in the platform yet, and nothing has been
adopted yet.

We have a plan in four parts, and the shape of it is this. First, repair the site
where it sits, as ordinary file changes to the deployment repository — the broken
pages, the missing navigation, the dead links, the address list search engines
read. This is worth doing on its own merits, and it also matters for what follows,
because the platform learns the site by reading it, so the site has to be correct
before it is read.

Then, mend the framework. The honest fix is not to work around the adoption
process for this one site; it is to build the missing faithful-preservation mode
that the settings already promise — so that "adopt this site exactly as it stands"
becomes something the platform can actually do, for this site and every one after
it. That is two focused changes, both opt-in, so no existing site can be affected
by them. They will go through the platform's own review council before they count
as done.

Then adopt the site through the real pipeline, using that new mode, and watch it
land — with a check partway through that compares what the platform captured
against what the site actually serves, character for character, so we find out
immediately if anything was altered in transit.

## Where we're going

The immediate destination is a site that looks exactly as it does now to a
visitor, but is fully inside the system: known, tracked, checked, rebuildable, and
safe from being quietly rewritten.

The larger destination is the one the owner asked for. When this is finished, the
answer to "can we adopt a hand-built site without wrecking it?" stops being "not
really" and becomes "yes, and here is the site we did it to". The gaps we cannot
close in this pass — and there are a few, including the fact that the platform's
redirect feature is a database table nobody ever wrote the code for — get written
down as findings rather than left as folklore, so the next person meets them as
warnings instead of surprises.

One thing worth flagging for the owner now, because it is a judgement call rather
than a technical problem: the site's written content quotes interest rates and a
"last updated" date from March. Fixing those figures is not part of adopting the
site, and I am deliberately leaving them alone in this pass — but once the site is
managed, keeping numbers like that current is exactly the sort of thing the
platform could be doing for us, and it is the obvious next conversation.
