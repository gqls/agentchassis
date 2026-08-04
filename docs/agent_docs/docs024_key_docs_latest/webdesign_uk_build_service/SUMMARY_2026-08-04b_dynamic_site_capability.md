# SUMMARY — can the framework build dynamic sites? State as at 2026-08-04

*Written in answer to a direct question. Current state only, measured today
against live code and the live database, not read off the concept register.*

---

## The short answer

Yes, more than expected, and the boundary is not where the documentation says it is.

The framework already builds and deploys websites onto real servers, in
production, today. What it does not do is **write the server-side code**. That is
the actual line, and it is much further along than "static sites only".

## What is genuinely built and running

**Sites can be deployed to a machine instead of to file storage.** The switch is
per site and already exists: a site whose repository is set to `vm-sites` gets
deployed to a box rather than to the static bucket. Two sites use it today,
idea.uk and relojistas.com. Relojistas has twenty pages built by the framework and
was deployed this morning, so this is a live path and not a leftover experiment.

**There is a recognised class of "site with a backend".** Such a site is marked
with a capability flag, and the platform knows the difference: there is a health
check that probes those sites and deliberately does nothing for static ones.

**Interactive, in-the-browser behaviour is properly supported and proven.** The
calculators and tools across the estate are framework generated. There is a
working mechanism for sections that fill themselves from a data feed after the
page loads, proven on three separate components, with a library of shared scripts
and guards that run at generation time to stop known mistakes being baked in.

So "static" is the wrong word for what the framework produces. It produces sites
that are interactive in the browser, and it can put them on a server.

## What is not built

**The framework does not generate backend code.** The server-side piece, called
the site engine, is a single hand-written Go service, the same binary on every
machine, offering a fixed handful of endpoints. Nothing generates it per site and
no agent writes server code. The register describes agent-generated per-site
backends as a future direction, and that remains accurate.

**A component cannot declare that it needs a backend.** There was a design for
this: components would carry a tag saying they require a server, and the planner
would refuse to place them on a static site. Checked today, the column that tag
would live in does not exist, and no active agent configuration mentions it. So
the planner can neither be told a component needs a backend nor stopped from
putting one somewhere it will not work.

**There is no tooling to provision or repair a machine.** The health check that
notices a box is down has no ability to act on it, and its own comments name a
future adapter as the thing that would eventually fix that.

## Why this matters for webdesign.uk

It changes the plan I wrote this morning, which assumed the site had to stay
static with the chat bolted on at a separate address. The better shape is to do
what relojistas already does: let the framework build the pages and deploy them to
the box, and have the chat live on the same machine as another path on the same
site. That removes a whole class of fiddly cross-address problems and copies
something already working rather than inventing.

The chat service itself still has to be written by hand. That was always true; the
difference is that it now sits inside an existing pattern instead of beside one.

## One thing found along the way that needs someone's attention

The health check for backend sites is built but has never been switched on, which
another team already knows and is tracking.

What they did not have, and what I have added to their file, is that switching it
on will not be enough. The check only looks at sites explicitly flagged as
backends, and of our two machines only relojistas carries that flag. **idea.uk does
not, and idea.uk is the one taking card payments.** So the first time that check
runs it will report that backends are healthy while the only machine earning money
is not being looked at, and because a skipped site returns an empty result rather
than an error, nothing will say so.

It is a one-row fix, but it needs doing at the same time as switching the check
on, otherwise the first clean run becomes the evidence that everything is fine.

## A note on how this was established

Not from the concept register. That file is frozen at the thirteenth of July and
its entry on this subject says nothing has been built beyond the basics, which is
now wrong. Its own banner warns that absence from it is not evidence of absence in
the platform, and this is a clean example: the VM deployment path, the backend site
class and the health check all postdate the freeze.

Everything above was checked against the running code and the live database today.
The one place I ran a check whose control came back empty, I noticed and re-ran it
a different way before drawing a conclusion, because a query that finds nothing and
a query that asks the wrong question look identical.
