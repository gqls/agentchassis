# SUMMARY 2026-08-08 — webdesign.uk build service

Third in the series (2026-08-04, 2026-08-06, this one). Written the night the
rebuilt site went up on the preview.

## What we're trying to do

webdesign.uk will sell complete websites to small and medium UK businesses: one
fixed price paid once, the whole site built and put live on the customer's own
domain, full refund at any point until they accept it. The shopfront has to be
built by the very machinery it sells — the site is the portfolio piece — hosted
on our own machine behind our own controls, and eventually carrying a chat box
that opens a conversation with a visitor instead of a form. Then, and only on
the owner's say-so, the domain switches over and the shop is open.

## Where we've come from

By the sixth of August every piece of machinery was in place and proven: the
server, the private tunnel, the preview address, the redirect that keeps the
public domain pointing at the owner's other site in the meantime. But the first
proper build of the site itself was rejected on sight — one page instead of a
site, no styling, copy that read brittle, and no way in for a customer. All four
faults were traced: the pipeline's first look at the domain had studied a
since-deleted hand-made page and concluded a one-page site was wanted; the
styling had been filed to the wrong repository because the site's record was
pointed at the right one only after the styling was made; the copy inherited
both problems; and the chat box was always a later phase. The plan out was a
clean resubmission, constrained by an explicit page list, with all copy
rewritten under the owner's improved writing rules.

## What we've done

Tonight the rebuild went from first submission to a finished site on the
preview, and the road there taught us as much as the destination.

Before pressing go we checked the thing we said we'd check: the crawler that
studies a domain at the start of a build does follow the public redirect, and
would have studied the owner's other site — a hundred pages of the wrong
business. We switched the redirect off for the window. Then we caught a second,
subtler trap: the crawler keeps a copy of what it last saw, and the copy it held
was one our own checking had planted minutes earlier. We told the pipeline's
first-look step to always fetch fresh — a permanent fix, kept — and watched the
classifier read the domain correctly this time: no existing site, five named
pages from the plan, nothing invented. Every trace of the earlier contamination
was regenerated clean.

The build itself had to be walked through by hand, stage by stage, because the
shared machinery that is supposed to notice queued work and start it runs on
nobody's clock — a fortnight of other sites' work sits ahead of ours in a queue
that only moves when pushed. The walking worked: research, strategy, briefing,
the plan (exactly the five pages asked for, nothing more), then five pages and
four pictures built in parallel, then styling and logo — which the machinery had
skipped because it remembered making them in August's first build, not knowing
the files had gone to the wrong shelf — and finally the pass that stitches the
menus so every page links every other.

The checks earned their keep on the way. The writer tried "we usually reply the
same day" on the contact page; the owner's rules blocked the save and the
rewrite came back clean. The what-you-get page described the product as "not a
single template page" — breaking the rule about not repeating a frame even to
deny it — and the automatic gate missed it, because it turns out to inspect only
a small slice of page types; our own end-of-build sweep caught it, the page was
rebuilt by the machinery, and the gap in the gate is written up for the platform
people. The home page title had picked up a long dash through the same side door
as last time; fixed the same way as last time. Final tally, read from the served
pages themselves: zero banned-claim hits in anything a visitor reads, zero
dashes, every link and picture and stylesheet resolving.

## Where we are now

The five-page site — home, how it works, what you get, the questions page with
the price and VAT answers, and contact — is serving on the private preview,
built end to end by the framework, checked against every rule the owner has set,
in the right repository, on our own machine. The rejected first version is
archived out of the way. Nothing public has changed: the domain still forwards
to the owner's other site, and the preview is the only window in.

Open, and known: the five "get started" buttons have nowhere final to point,
because the thing they should point at is the chat box, which is the next phase
— nine review items hold that question for the owner. The site has no favicon
(neither build ever made one), the pages carry no search descriptions yet, and
two platform faults found tonight — the queue that doesn't drive itself and the
content gate that inspects less than it appears to — are documented with
evidence for whoever owns them.

## Where we're going

The owner looks at the preview and says what he thinks. In parallel or after:
the chat service gets built behind its spending controls once the owner provides
the scoped key, the input box then replaces the email link as the way in, and
the review items about the buttons resolve themselves in the process. When the
owner is happy with what he sees, cutover is one DNS change and the deletion of
two redirect rules. The queue and gate findings go to the platform side rather
than being patched from this lane.
