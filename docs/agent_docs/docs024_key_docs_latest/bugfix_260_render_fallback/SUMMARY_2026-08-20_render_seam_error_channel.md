# SUMMARY 2026-08-20 — the render seam now has an error channel

Written to be read aloud. Five parts, in order.

## What we're trying to do

Stop a single badly-shaped piece of content from silently wrecking a whole page section, and make
the failure say what is wrong.

Our page sections are built from reusable templates: HTML with instructions in it — "if there is a
heading, print it here", "for each step in this list, print one of these". The instructions are
meant to be carried out and then disappear, leaving finished HTML. When one field arrives in the
wrong shape — a list written as a sentence — the template engine cannot carry out the instruction
that walks the list, and everything downstream of that depends on what the code does next.

## Where we've come from

For months, what it did next was hand the job to a much older renderer that speaks a **different
dialect** of instruction. That renderer filled in the individual words it recognised and left every
instruction it did not understand sitting in the page. The output was well-formed HTML with the
values already resolved, so nothing noticed — until a final check refused the whole page for
containing template gibberish, reporting up to twenty "blockers" that were really a cap of ten
matches, not a measurement.

The bug had been filed since 12 August with the headline "no live damage", and that phrase was
hiding the cost. It was true of *stored* content: nothing broken has ever been saved, because the
final check refuses first. But the defect's entire effect is that **nothing is saved** — so
counting saved rows reports it as harmless by construction. Meanwhile the page never exists and the
pages that did build keep linking to it. Three lanes hit it independently in one week; loanzy.uk
serves a 404 where its consumer-rights page should be, and remortgagecalculator.uk carries two dead
links in the navigation of every page it serves. By 18 August it was twenty-six occurrences across
seven domains and still accelerating.

## What we've done

**Deleted the older renderer, and given the render step a way to say "I could not do this".** A
component render now either executed or errored — there is no third outcome. That deliberately
broke every piece of code that asked for a page and had no way to be told it failed; there were
fifteen, and each now makes a decision you can read in the code: the page builder stops and names
the field; the repair path keeps the good page it already has and asks the writer to fix the
content; the header/footer code falls back to a plain version; the two paths that edit a page
already published refuse the edit and leave the live page untouched.

**Measured before deleting, three ways, each with a deliberate wrong answer built in** so the test
could have failed: not one of 251 live templates fails to load, not one of 1,778 stored sections
fails to render against its own content, and not one of 253 components is even written in the
dialect the old renderer spoke. We also re-added the old behaviour on purpose and confirmed five
tests fail — and recorded the one test that *passed*, because it has a second line of defence and
therefore does not prove what it appeared to.

**Added a plain-English diagnosis.** A new check reads a component's declared field shapes against
the content and reports, for example, `steps[2].branches: declared array (items: object), got
string` — with the position, because the real case is nested two levels down. It runs always as an
explanation when a render has already failed, and optionally as an earlier refusal that is switched
off by default and stays off until after the next release.

**Put it through the reviewer council, which earned its place.** The first round came back
"revise" in fifteen minutes and found three things we would otherwise have shipped: three of the
fifteen sites were still unwritten while the plan said they were done; a known-unsafe piece of Go
behaviour had been described in our own rationale as if it were a safety feature; and a "nobody had
ever noticed this" claim was simply untrue — three earlier pieces of work had noticed it, one had
half-fixed it. All corrected; the second round approved with no serious objections.

**And writing the switch-on file caught a mistake nothing else would have.** Asking "what would the
new check refuse if we turned it on today?" returned one live, healthy, serving page where five
items are simply blank. The check was treating blank as the wrong type. It is the only page on the
estate shaped that way, so no test we would have thought to write would have found it; the check
now shares one definition of "blank" with the check that already existed.

## Where we are now

The code is committed and approved, and **it is not live**. Go changes do nothing until the next
fleet release builds and ships them, so nothing has changed on any site today. The release tag is
prepared and the release itself is the owner's to run. The optional early check stays off, and its
switch-on file is deliberately held back until after the release.

The mechanism is registered so other work can find it, the contract change has an architecture
write-up, two traps are recorded where the next person will trip over them, and the three lanes
waiting on this have been told what to expect — including one warning that matters: a brand-new
site whose header or footer cannot render will now **fail the build** rather than publish gibberish
and report success.

One thing is honestly still open and we have said so in four places rather than one: this makes the
*wrong-shape* case loud, and leaves the *missing* case exactly as it was. A field that is absent
rather than mistyped still renders as empty with no error, and the check that catches that runs at
only two of the fifteen places a page can be rendered. It needs an owner and does not have one.

## Where we're going

Next is the release, and then the real test, which is not ours: the loanzy lane runs a clean build
from scratch and reports either way. After that, one item on the locked specimen site is re-armed
to confirm the new error names the right field on a page nobody has touched since it broke. Only
then does the optional early check get switched on, and its file is written so that whoever applies
it must first re-run the population query and read the rows, not the count.

**Success is not "the page builds."** The twenty-four pieces of work sitting in the review queue
still hold content of the wrong shape, and making that content correct belongs to the writing lane.
Success is that the failure is honest, immediate, names one field, and can no longer reach a
published page through the two routes that had nothing guarding them.
