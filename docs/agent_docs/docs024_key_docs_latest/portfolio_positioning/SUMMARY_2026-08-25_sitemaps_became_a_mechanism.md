# SUMMARY — 2026-08-25 — Sitemaps stopped being a script and became a mechanism

*Written to be read aloud. Series entry: follows `SUMMARY_2026-08-21_the_machine_that_writes_the_brief.md`.*

---

## What we're trying to do

Make the platform produce sites that can actually be found, without anyone remembering to do
anything. A sitemap is the file a site publishes listing its own pages, so a search engine can
find them all instead of guessing. The owner ruled on 20 August that all future sites should have
one.

The wider point behind this particular job: **the estate keeps building good tools that nobody
runs.** A capability that needs a person to remember it is not a capability, it is a to-do. This
was a test of whether we close that gap or keep widening it.

## Where we've come from

A generator existed — a script, written 28 July, that was genuinely good. It had learned two
expensive lessons already: check a page really exists before listing it, and read every column
that decides whether a page should be findable.

But it was a script someone had to run. By 20 August only 8 of the live sites served a sitemap,
and the missing ones included the newest pilot, built four days earlier with every other guard
applied. That is the clearest possible statement that a manual step is not a mechanism.

So on 21 August the script was promoted into a platform action. And then the review panel failed
it, correctly, on a single point: the action worked, was registered, and **nothing anywhere called
it**. A tool nobody runs, in a new costume.

## What we've done

We gave it a caller — a scheduled sweep that takes one site every half hour, regenerates its
sitemap and publishes it, coming back to each site every three days. We chose that over
"regenerate whenever a page changes" for a reason we measured rather than guessed: the sweep is
the only version that helps sites which have no sitemap *today*, and checking pages costs one
fetch each, so on our largest site the page-change version would mean 135 fetches every time
anyone edited anything.

Along the way we found and fixed two defects in the generator itself, neither of which anything
downstream would ever have complained about:

- It listed every site's front page by the wrong address. The sites declare their front page is
  plain `/`; the generator said `/index.html`. Same page, two names, and search engines want to be
  told which one is official.
- Its rule "don't list a page that forwards somewhere else" was written down clearly in the code
  and **had never actually worked**. The standard way of fetching a page follows the forwarding
  automatically, so the checker only ever saw the destination and passed everything.

The second one is the more instructive. The count of rejected pages — the number whose entire job
is to catch exactly this — read zero. It looked perfect precisely where it was blindest.

## Where we are now

The owner switched the sweep on at half past three on 24 August. It then ran unattended for
thirteen hours and finished at quarter to five the next morning.

**Twenty-seven sites swept. Twenty-seven successes. Nothing dropped** — checked one by one against
the record of what actually ran, because the system marks a site "done" just before sending the
work off, so a lost message would leave a site looking finished when nothing had happened.

**Sitemap coverage went from 8 of 28 live sites to 26 of 28**, and every one of the 26 is
complete — every address in every sitemap is a real page on that site. The two remaining are both
understood: one is the site under the owner's halt, which the sweep deliberately skips, and the
other forwards every address to a different domain and correctly ends up with no sitemap of its
own.

The front-page fix proved itself by accident, which is the nicest kind of proof. The chassis
rebuild went out between the first and second site of the sweep, so the first site was done on the
old code and the other twenty-six on the new — same job, same night, one change in between, and
the results differ exactly where they should.

**Both things that went wrong this week were in our measuring tools, not in the work.** One check
scored the fix as a regression because the fix deliberately changed the thing the check compared
against. Another was justified with reasoning that a warning written the same day says doesn't
hold. Both are written up where the next person will hit them.

## Where we're going

Three things follow, in order of how soon they matter.

**The other half of the original question.** A sweep every three days means a newly published page
waits up to three days to become findable. Wiring the generator into the publishing path as well
closes that, and was always the second half — deliberately deferred until the first half was
proven, which it now is.

**One known gap we chose to record rather than paper over.** When the generator finds nothing to
list, it stays silent, and it stays silent for two opposite reasons: the site opted out, or
something is broken. Given our sites carry between 26 and 135 pages, the second would almost
certainly be a fault, and nothing would tell us. The fix is designed and written down; it needs
the next change to the generator.

**And the thing this was always in service of** — the 22 hosted sites waiting to be remade, and the
domains with no positioning yet. Sitemaps were a precondition, not the point. That work is
unblocked and untouched.
