# SUMMARY 2026-08-08 — the site speaks in the new voice

## What we're trying to do

Make loancalculator.co.uk read like a person explaining something, rather than a
product describing itself — and do it **through the framework**, not by hand. The
second half is the point. Anyone can rewrite twenty-six pages in an afternoon by
editing them directly; the value is that the framework wrote them, which means every
page passed the same claim checks, banned-phrase sweeps and structural gates that any
future page will pass, and that the next voice change is a config change rather than
another afternoon.

## Where we've come from

The owner and the portfolio lane developed a "gentle explanatory" voice (H) at the
start of August. Seeding it on this site was quick. Actually using it was not: the
first attempt could not rewrite a single page, and the reason took two days and two
bug fixes to reach.

When this site was adopted it was broken into editable pieces named by position —
`prose-0`, `prose-1`. The build pipeline looked those names up against component
*types*, found nothing, and gave up on all 57 pieces; worse, it responded by asking
the fleet to manufacture new components with those names. That was `bugs_open/204`.
Fixing it armed a second trap underneath, `bugs_open/189`, where a save renamed a slot
and duplicated a locked row. Both were fixed, reviewed and proven on a real page on
6 August. The owner reviewed that page and said to go ahead.

## What we've done

Twenty-five pages dispatched through `page-build-handler`, in batches, each one graded
before the next batch fired. The prompt was copied by SQL from the very work item the
owner reviewed, so every page got bit-for-bit what he approved.

**23 of 26 active pages now carry the voice.** All 26 serve. The twelve locked
calculator rows are identical in row id, timestamp and rendered HTML, and the
equivalence harness reports all eleven calculators reproducing their golden values
exactly. The legal page was read line by line rather than graded mechanically: every
disclaimer, the FCA statement and all four UK GDPR points survive, and its accuracy
claim came out *more* conservative than it went in.

Two things were found that no amount of planning would have predicted:

- **Four pages had their calculator styling deleted by the rewrite**, because the
  decomposer had filed the page's `<style>` block into a slot called `prose-0`. Every
  guard in the platform passed it — the component's own guidance promises "rewriting
  this cannot break a calculator", which is true and silent about CSS. Caught on the
  first page, all four restored and confirmed on the live site, and written up as a
  landmine because the writer *kept* the style block on the other four such pages: a
  spot-check would have cleared the class wrongly.

- **Three pages cannot be rebuilt at all**, and it has nothing to do with the copy. The
  content validator substring-scans the whole assembled page, comments included, for
  words that suggest a model wrote about its task instead of doing it. Three of the
  site's calculators carry a developer's note containing one of those words. The
  validator blocks, every time, for ever. Filed as `bugs_open/219` with a control group
  as evidence: exactly the three pages with that word failed, all nine without it
  passed.

## Where we are now

The site reads in the new voice everywhere except the home page, the car finance
calculator and the interest rate stress test — which are unchanged and healthy, waiting
on a platform fix that is filed and small. The calculators are proven untouched. The
copy is on-spec rather than merely different: it opens where the reader is standing, it
explains before it names, and it has quietly retired the absolutes.

One open question belongs to the owner, not to this lane: several calculator pages had
near-empty intro slots, and the framework filled them with substantial new explanatory
copy rather than restyling what was there. It reads well and passes the claim checks,
but it is new substance on a finance site, so it is a decision rather than a detail.

## Where we're going

Fix `219` and the last three pages take minutes each. Re-baseline the calculator golden
once they are in — not before, because a partial capture certifies nothing.

Then the wider questions that this work was always the prerequisite for: the other
finance sites, held on the owner's review; and whether the H voice folds into the
fleet-wide base prompt, which is seven already-drifted copies and a genuine conflict
with a live default, and therefore an architecture round rather than a config edit.

The deeper result is not the copy. It is that a site adopted from hand-built HTML can
now be rewritten by the framework, end to end, with the calculators provably untouched
— which is the thing this whole adoption was for.
