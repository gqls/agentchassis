# SUMMARY — 2026-08-18: five imported tools rebuilt by the framework, and three of them were broken before we started

*Written for the owner to read aloud. Second summary in this lane; the first (2026-08-16) marked the
pilot working. The milestone here is different: the process is now routine, and it has started
finding real faults rather than merely replacing markup.*

## What we are trying to do

webdesign.co.uk carries 63 tools that were imported rather than built — dropped onto the site as
finished lumps of HTML sharing one wrapper. Because of that the framework has never been able to see
them: it cannot check them, improve them or rebuild them the way it does everything it made itself.
We are replacing them with tools the framework builds and owns, at the same web addresses.

## Where we have come from

The pilot proved it was possible at all, after we found and fixed the reason it never had been: the
step that saves a newly built tool could only ever create a new page, never attach to an existing
one. Since every rebuild is "same address", that had blocked the entire idea.

Since then the process has been through one real failure and one real mistake, and both made it
better. The failure: a rebuild died because tool names are unique across the whole estate, not per
site, and a library template nobody had placed anywhere was quietly holding the name. The mistake was
mine — the written check that would have caught it says "this must come back empty", it came back
with three rows, and I reasoned past it using the logic of a different check I had just been reading.
Both are now designed out: every filing asserts both checks inside the transaction before it writes,
so a blocked rebuild fails in milliseconds instead of after the AI has done its work.

## What we have done

**Five tools are rebuilt, live, and checked at the actual served page**: the aspect-ratio calculator,
the Markdown table architect, the HTML minifier, the SVG code stripper and the SRI hash generator.
Each one was judged before the old version was switched off, and each old version was retired rather
than deleted, with its content fingerprinted before and after to prove it can be put back in one step.

**The unexpected result is that three of the five were already broken in production.** Two of them —
the HTML minifier and the SVG stripper — had a checkbox advertised on the page that did nothing at
all: the line of code implementing it had been swallowed by its own comment, almost certainly when
someone hit a syntax error and commented it out to make the page load. The minifier had a second
fault that would quietly corrupt anyone's page, flattening the contents of `<pre>` and `<script>`
blocks it should have left alone. None of this is visible from looking at the page; we only found it
because we now read the live tool's code before writing the rebuild brief. The rebuilds fix all of it.

## Where we are now

Five tools live and proven. Two — the A/B test calculator and the meme generator — are deliberately
parked: the only cheap way to unblock them is to retire the shared library entry other sites fork
from, and both of those have a live copy on another site. That question is written up as `RFC_036`
and waiting on the mechanism you preferred, where a rebuild records itself as a fork.

One thing went wrong yesterday evening and is worth stating plainly. Each rebuild has a short window
where the old version must be switched off before the site automatically republishes the page. I lost
that window on the SVG tool: it republished with both tools on the page, and stayed that way for
about an hour before I caught it. Nothing was lost and the repair was one step, but the honest lesson
is that the step was never really under control — it had simply been lucky three times. The rule now
is not to start a rebuild we cannot sit with; the tool built after that one was switched over 46
seconds after it finished.

## Where we are going

Roughly 55 tools remain. The process is repeatable and each takes minutes of machine time, so the
limit is review attention rather than throughput. The recommendation stands that the hand-built
applications — the mind-map studio, meme studio, logic architect, the mini-CMS, the pasteboard — go
last and one at a time, since those are reimplementations rather than copies and are the ones most
worth looking at properly.

The five live pages, if you want to try them:
`/tools/aspect-ratio/`, `/tools/markdown-tables/`, `/tools/html-minifier/`, `/tools/svg-optimizer/`
and `/tools/sri-generator/` — each with `index.html` on the end, because the bare directory is a 404.
