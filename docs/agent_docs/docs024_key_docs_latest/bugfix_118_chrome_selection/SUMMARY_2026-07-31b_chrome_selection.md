# SUMMARY — 2026-07-31b — chrome selection: live, and the fleet repaired

*(A second summary the same day because the read-out genuinely changed: the first
was written with the fix inert and the fleet untouched. Both are now false.)*

## What we're trying to do

Make the platform give one answer to one question: *which component from the
library should serve a site's header, footer or head?* — and then make the fleet
actually use the answer.

## Where we've come from

`bugs_open/118`, filed 27 July: a fix applied to the footer component marked
*active* changed nothing on any page, because the code picking chrome never looked
at whether a component was switched on. It sat for four days on the belief that
fixing it would change every site's footer.

Measuring showed three separate bits of code asking the question, all three
answering differently and all three wrong — one ignoring active/inactive, one that
would hand a client's private forked header to every other site, one returning a
page-section component. And it showed the parked belief was wrong: the chooser only
runs for sites with no chrome assigned, and every real site had one.

## What we've done

The fix — one rule, used by both places that assign chrome, with a test that fails
if anyone writes a fourth copy — is **committed, council-approved at round one, and
now LIVE on v1.0.1219**, verified by grepping the running binary on both replicas
with a positive control rather than trusting the tag.

Then, on your call, we repointed the fleet: **21 assignments** moved off deactivated
components — 11 footers and 10 headers — under the human-lock guard, with the
previous mapping backed up first, followed by a chrome re-render on all 11 sites.
Leopardess keeps its own forked header, which is what a fork is for.

**28 of 28 header and footer slots across all 14 sites now render from an active
component.** None did this morning.

The proof is on the site the bug came from: relojistas' footer now emits
`<h4>Explore</h4>` instead of `<h4>Our Services</h4>`, and its Contact column is
gone — which is `bugs_open/111`'s fix finally working, on the component that
actually renders. That silent failure is what filed 118 in the first place.

## Where we are now

118 is **closed** and in `bugs_closed/`.

Doing the repoint by hand exposed two further gates that no amount of reading the
code would have shown, and both sharpen `bugs_open/166`: the repair handler only
touches chrome when explicitly asked to, and even then it **skips a slot that
already has HTML** — so a corrected assignment still does not reach the page. The
repair needs one more flag set on the detector's spec. That is now 166's cheapest
fix and it is written down.

**The honest half: stored chrome is correct everywhere; the deployed pages are
not.** They serve the old footer until the **206 page-rerender jobs** now queued
drain. `curl relojistas.com` still shows the old markup. That queue is
`bugs_open/149`'s lane, not this one, and "28 of 28 slots correct" must not be read
as "the fleet looks right".

## Where we're going

Nothing further is owed in this lane. Three things sit downstream, all filed:

- the 206 queued page re-renders need to drain before anyone sees the change;
- `bugs_open/166` — the routed repair still cannot clear a deactivated assignment
  on its own, so the next deactivation puts us straight back;
- `bugs_open/167` — the page-*building* path can still render a page-section
  component as site chrome, which is a fleet-visible change and remains your call.

And one library gap, which is data rather than code: **there is no active `head`
component at all**, so 13 head slots still point at deactivated ones and the
platform falls back to a hardcoded default. Activating one changes every page's
`<head>`, so it wants the same one-site-first treatment the footers just had.
