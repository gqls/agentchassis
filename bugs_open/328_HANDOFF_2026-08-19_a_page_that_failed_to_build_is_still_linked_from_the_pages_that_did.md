# 328 — a page that failed to build is still linked from the pages that did, so one blocked page becomes a visibly broken site

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane, from a greenfield one-shot build.
**Status: OPEN, UNOWNED. Live on `loanzy.uk` today.**

> **On the 090 loop:** not run. This is not a hypothesis about a mechanism — it is two URLs,
> one serving and one 404, and the anchor between them, all read from the live site.

## The one-paragraph version

When a page fails to build, the build knows. The item is `needs_human_review`, the `pages` row
never reaches `deployed`, and the failure is recorded. **Nothing tells the pages that link to
it.** They are built, deployed and served with anchors pointing at a page that does not exist,
so a single blocked page turns into a live site with dead navigation — and the reader gets a
404 with no indication that anything is unfinished.

## Evidence (measured 2026-08-19, live)

`https://loanzy.uk/` serves 200 and contains:

| link on the live home page | target status |
|---|---|
| `/your-rights.html` | **404** — build refused at `validate_content` (`bugs_open/260`, 20 blockers) |
| `/guides/index.html` | **404** — builder no-op'd honestly: *"no sections ready to build"* |
| `/tools/loan-comparison-calculator/index.html` | 200, **but empty of its tool** (`bugs_open/311`) |

Two of the three defects behind those targets are already filed and owned. **This bug is about
the third fact: the links shipped anyway.** The home page was built AFTER `your-rights` had
already failed, so the information was available at build time and simply not used.

## Why it is its own defect and not part of 260/311

Fixing 260 removes this instance; it does not remove the class. **Any** page can fail to build —
truncation, a validation blocker, a missing component, an exhausted retry budget, an
infrastructure burst (`bugs_open/307`) — and every one of them produces the same shape: a live
site advertising a page that is not there. The route needs one rule, not one rule per cause.

It is also the difference between the two possible readings of a build. A site with four good
pages and no links to a fifth reads as *small*. The same site with two dead links reads as
*broken* — which is the judgement a customer makes about the product, not about the page.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Resolve links against `build_status` at deploy time.** A link whose target is not
   `deployed` is dropped (or rendered as plain text) by the deploy step. The bad state stops
   being representable because the anchor cannot outlive the missing page. ⚠ must consult
   `pages.build_status` **and** the artefact, given `bugs_open/315` (`deployed_at` is stamped
   whether or not the object was written).
2. **Hold the first publish until the corpus is complete**, then publish atomically. Strongest,
   and the most disruptive: it changes the route from incremental to all-or-nothing, and it
   would also have prevented the stale-nav window this build served for ~1 hour.
3. **File a visible item per dangling link and re-render the referring pages** once the target
   builds. Weakest — it is repair after publication, and it depends on the repair path
   actually running (see `bugs_open/313`, where `plan_links` has never run at all).
4. **At minimum, a post-deploy assertion**: after a build completes, fetch every internal link
   on every deployed page and fail the run if any 404s. This does not prevent the defect, but it
   converts "nobody noticed" into "the build reported it", which is what happened here only
   because a human went looking.

## How to verify a fix

Build a site in which one page is *forced* to fail (a deliberately invalid section is enough),
and require that no deployed page links to it — asserted **at the served HTML**, not at the
plan, the sections, or `page_components`. Positive control in the same run: a page that DID
build must still be linked, or the fix has simply stopped emitting internal links (which
`bugs_open/313` shows is a state this platform can reach and not notice).
