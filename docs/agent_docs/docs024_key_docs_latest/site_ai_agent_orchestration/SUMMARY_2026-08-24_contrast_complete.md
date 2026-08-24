# SUMMARY — ai-agent-orchestration.com, 2026-08-24: the readability work is finished

Written to be read aloud. First summary for this lane; the milestone is that the thing it was
opened for is done.

## What we're trying to do

Make ai-agent-orchestration.com a site that demonstrates the platform rather than embarrasses it.
It is the shopfront for a multi-agent build capability, so anything visibly broken on it is an
argument against the product. The immediate job was readability — text that could not be read
because it was drawn in nearly the same colour as the thing behind it — and then two things the
owner asked for on top: carousels for the case studies, and real pictures instead of broken ones.

## Where we've come from

The site had **44 places where text was effectively invisible**, across four pages. Fourteen of
them measured 1.00:1 — text painted in exactly its own background colour.

The cause was not one bad decision. It was three unrelated ones that happened to land on the same
pages:

- Shared component templates asked for a colour by a name that, on this particular site, resolved
  to the same value as the surface behind it. Only two sites in the estate have a palette where
  that is fatal, and this is one of them.
- Two components had **no theme support at all** — six hard-coded colours between them — in a
  library where the component next door does it correctly. On a dark site they were never going to
  work.
- One page, `pricing`, had lost its stored content back in April, so it could not be re-rendered at
  all. It was frozen with its faults intact.

## What we've done

Four database migrations and one code fix, each applied on its own, rehearsed first, and reversible
byte-for-byte.

- **456 / 457** repointed the shared templates. 457 exists because 456 introduced a fault — it
  changed a button's label colour without noticing the button had a coloured fill behind it. That
  was caught by re-measuring, not by the overall improvement, which looked like a clean win.
- **469** gave the two unthemed components the site's own colours — all six values moved together,
  because moving only the backgrounds would have left dark text on newly-dark cards.
- **557** fixed the reason `pricing` could not be rebuilt, and this is the one worth understanding:
  the site's own instruction sheet told the writer to use a phrase that the site's own fact-checker
  could not accept. Following the instruction guaranteed rejection.
- **559 / 560** built the carousel and generated and attached nine diagrams.
- **`bugs_open/364`** fixed a genuine flaw in the platform's honesty checker, which was reading the
  "2" in "2am" as a business statistic.

## Where we are now

**Zero unreadable elements on all four pages.** The arc was 44 → 32 → 8 → 0.

The carousels are live on two pages and switched off for the two other sites that share the
component — verified on their live pages, not just in the database. All ten case-study pictures
load. The broken markdown link that used to show raw `[text](link)` syntax to visitors is gone.

Two findings remain and both are text sitting on top of an image, where the measuring tool says
plainly that its own reading is approximate. Every figure quoted above excludes them.

**The last eight failures were not fixed by us in the room.** The migration cleared the blocker;
the page then rebuilt itself hours later, unattended. That is worth saying because the honest
version of this story is that the fix was a source change, not a heroic repair, and the thing that
made it finally work was the writer happening to produce different wording on the retry.

## Where we're going

Nothing is blocked and no work is outstanding on the original ask. What remains is small:

- A batch of 17 old readability items sits parked. They are not live defects; they can only be
  cleared by an automatic audit that has not run on this site since 10 August. Now that every page
  genuinely passes, that audit should clear them — and if it clears none, that tells us something
  about the clearing mechanism, which another team already owns.
- There is a mechanism that would keep the site's instruction sheet up to date automatically. We
  have deliberately **not** switched it on, because as built it would silently delete the site's
  "never claim this" rules. That gap needs closing first, and it would benefit every site.
- The carousel setting is not durable against a full page rebuild. It nearly got wiped once and
  survived by luck. If it ever disappears, the fix is to set it again, not to go hunting.
