# SUMMARY 2026-08-03 — the unpublish primitive, and the veto that split it

## What we're trying to do

Make it possible for the platform to **take a page down**. When we retire a page we set
`pages.status='archived'`, which correctly stops it being re-derived and stops it being
listed — but nothing removes the file that was already published, so the page carries on
being served for ever, frozen at whatever it last said. Thirteen pages are in that state.
The one this was filed on serves a "learning centre" index whose article list was written
on 3 July and links to an article that has since gone, so a visitor gets a working-looking
page full of dead links.

## Where we've come from

Two bugs had walked into the same wall from opposite sides. `bugs_open/098` is "an
archived page keeps serving". `bugs_closed/125` was "a page published to the wrong path
can't be removed" — the owner deleted that one by hand. 125's council round asked that the
shared gap be written down rather than absorbed, and it was written down on 098. The gap
is one sentence: **the platform can publish a page but has no implemented way to unpublish
one.** The git adapter's only deletion verb was `delete_repo`, which returns *"not yet
implemented"*.

## What we've done

**Built the missing capability.** A deletion is expressed as *a kind of commit* — in
GitHub's API a removal is a tree entry with a null SHA — so it rides the existing commit
path and inherits the concurrency retry, the path-naming rules, and atomicity with writes
(a *move* becomes one commit rather than a delete and a write that can half-fail). Plus a
`delete_file` verb, and `retract_page_deployment` as the page-level caller with guards:
paths only ever derived from `pages.url` via the shared helper, a page still linked from
live content refused, nav rows retired, newly-stranded pages reported.

**Found that the bug's own mechanism was only half right.** Re-checking the target before
deleting it showed its `deployed_at` had moved to *today*. The page was not frozen — it
was being re-rendered and re-published **twice a day**, because the query that picks pages
for a news refresh asks "has this shipped?" (`build_status`) and never "does the platform
still want this?" (`status`). That has been true since 31 July, five days after 098 was
filed. It also meant a retraction would have been **self-undoing**: delete the file, and
the next refresh puts it back — while a single post-delete `curl` reports success. Fixed,
live, and its self-declared cousin had the identical defect and is fixed too.

**Widened retraction to a graph operation**, per the owner's directive that we should also
fix links pointing in, nav entries, and pages left unreachable. That is the same finding
098 records from the other end: a page elsewhere was correctly deleted and every page on
that site then advertised the resulting 404 from its own footer, because the nav row was
still live.

## Where we are now

**The council rejected it — a hard veto — and the veto is about packaging, not design.**
The architecture seat said plainly that the mechanism is right: *"expressing
delete-as-null-sha inside the existing CommitToRepo path is the right reuse… the plan is
sound and I'd carry it."* What it vetoed is that a **destructive verb was added to a
shared adapter's vocabulary inside a bug fix** — the same shape a previous change was
vetoed for. The guardian named the safest path: ship the small predicate fix now, take the
unpublish machinery to architecture review on its own.

So the work is split, and the split is honest about what is live:

- **Live and settled:** the twice-daily republishing of a retired page is fixed, on both
  chassis replicas, verified by image digest rather than by tag.
- **Live but vetoed:** the unpublish capability exists in the running binaries — the shared
  tree makes that unavoidable, and the platform's own ruling says review here is after the
  fact by design. It is registered with its veto, and routed to **RFC 011** for a human.
- **Not done, deliberately:** **no page has been retracted.** The owner approved retracting
  one page before the verdict existed. Firing a vetoed capability at a live customer site
  on an approval given under a different premise is precisely what the veto is about.

Four of the objections are worth having regardless of the packaging argument, and all four
land on the graph audit — the part added fastest and reviewed least. The sharpest: links
into a page also live in structured content fields, not just in `href=` markup, so the
"is anything still linking here?" check can miss one and let a retraction through — the
exact case the owner's directive was written to prevent.

## Where we're going

1. **A human breaks RFC 011.** The question generalises past this change: does a
   destructive verb on a shared adapter differ in kind from an inert field, such that
   "additive and inert" doesn't cover it? My own preference is recorded rather than hidden
   — keep the verb, remove it from the generic allowlist, so it is reachable by the
   retraction action and not by any future workflow author.
2. **Pay the four correctness debts** on code that is already live, whichever way the RFC
   goes.
3. **Then, and only then, the retraction** — with the two-part acceptance the bug file now
   carries: 404 immediately, and *still* 404 after the next news refresh. The second check
   is the one that tests anything.
