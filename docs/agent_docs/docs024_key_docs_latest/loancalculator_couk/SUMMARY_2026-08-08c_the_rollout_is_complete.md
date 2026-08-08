# loancalculator.co.uk — the voice rollout is complete

*Written 2026-08-08, evening. Third read-out today, and the last for this phase: the
morning's summary said 23 of 26 pages and named a platform bug as the blocker; the
midday one carried the first owner review. This one closes the phase — all 26 pages
are in the new voice, the blocker is fixed and live, and the calculators are proven
untouched. Series, not replacement: `SUMMARY_2026-08-08…` and `…08-08b…` stand.*

## What we are trying to do

Take a site that was hand-built by a person, bring it fully inside the framework, and
then prove the framework can change its writing without damaging what makes it useful.
The useful part here is arithmetic: eleven working calculators whose numbers a reader
is entitled to trust. The test of the whole exercise is whether we can rewrite every
word on the site and leave every number exactly where it was.

## Where we have come from

The site was adopted at the end of July, decomposed into framework components in early
August, and given a house voice ("voice H") that the owner reviewed on a single canary
page. The rollout began on 7 August: batch by batch, every page's prose rewritten by
the framework — never by hand, which is a standing ruling — and graded afterwards.

Two things bit along the way, and both are on the record because neither was obvious.
Eight of the site's fifty-one "prose" rows were not prose at all: they held the page's
stylesheet, and rewriting one silently deleted the CSS that lays a calculator out while
every guard in the platform said yes. And three pages could not be rebuilt at all —
they failed validation every time, with an error blaming the model.

## What we have done

**Fixed the platform bug that blocked the last three pages, and it was not the bug the
file said it was.** The page checker scans for phrases a machine writes when it talks
about its task instead of doing it. It was reading the entire file, including the code
and the notes a programmer leaves in the code. Our three calculators each carry a
careful note explaining their arithmetic, and those notes name the thing they explain —
which is the word the checker objects to. So the checker refused to save the page, and
blamed the machine for a human's writing from five days earlier.

The morning's bug file recorded those notes as being in one kind of comment and
proposed the obvious remedy: ignore that kind. They are in a different kind. Two of the
three files contain none of the sort named. **That fix would have passed review,
shipped, changed nothing, and been recorded as done.** What exposed it was asking the
database to extract just the comments and count: none of the offending word, where
counting the whole file found three. Two counts that had to agree, and did not.

The real fix was smaller and reuses machinery a council review had already settled for
the neighbouring check that morning: read only the words a visitor actually sees. It
was reviewed and approved first time, and proven on the exact bytes that had failed —
the system still held this morning's failed inputs, so the new checker could be run
against the real thing rather than an example we invented.

**Finished the site.** All three remaining pages went through the framework, the
homepage included — the owner ruled against holding it back, because the framework's
output is what he is judging. All three built cleanly and passed grading.

**Proved the calculators survived**, in that order: compare first, re-record second.
All eleven tools reproduce their reference values exactly, including the two whose
pages had just been rewritten. Only then was a new reference recording taken.

**Found a second, unrelated instance of the same class** on webdesign.co.uk, where a
page cannot currently be rebuilt because its own product copy says "as an AI-builder
prompt". Our fix does not help it — that text really is visible copy — so it was filed
on its own evidence and that lane was told, rather than quietly loosened into our
change or worked around by rewording someone else's page.

## Where we are now

- **26 of 26 pages in the new voice.** 26 of 26 serving HTTP 200, swept live.
- **11 of 11 calculators identical to golden**, re-baselined afterwards as
  `GOLDEN_2026-08-08_voice_h_complete.json`, now covering a fourth test vector the
  previous baseline predated.
- **The blocker is fixed, live on `v1.0.1266`, and verified in the running binary** on
  every replica with a control pod proving the check could have failed.
- **The grader was proven able to fail** in the same session by deliberate mutation,
  which is what makes its passes worth quoting.
- Its record is kept open rather than filed away, by owner direction, with the evidence
  inside it.

## Where we are going

Two open questions, neither blocking, both belonging to someone else to answer.

The first is the owner's and has been waiting since yesterday: several calculator pages
had near-empty prose stubs, and the framework did not restyle them — it **filled** them,
with 800–1,900 bytes of new explanatory copy each. It reads well and trips no claim
gate, but the brief said "voice only, preserve every fact, add nothing", and this is new
substance on a finance site. Both readings are defensible. If the answer is "trim", the
original bytes are in the backup table.

The second belongs to webdesign.co.uk: the `as an AI` pattern needs tightening to the
first-person disclosure it was written for, and their tools index cannot be rebuilt
until it is.

What this phase actually demonstrated, beyond a site in a new voice: the framework can
rewrite every word of a working site without moving a single number — and we can show
it, because the checks were built to be capable of failing and were made to fail on
purpose before their passes were believed.
