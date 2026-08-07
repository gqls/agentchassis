# SUMMARY — 2026-08-07: the structure is built, the copy goes back to the framework

## What we're trying to do

Take loanandmortgagecalculator.co.uk — a live site of 41 pages, 23 of them
interactive calculators — and get it writing in the register the owner chose
(the "gentle explanatory" voice, trial H, approved with four sample rewrites).
The site must end up **maintainable by the platform**, not by whoever last
edited it, and no calculator may stop computing correctly along the way.

## Where we've come from

The site was hand-built in this repo on 31 July, went live, and was then
adopted into the framework as **41 frozen whole documents** — one stored file
per page, shipped byte-for-byte, which the platform never looks inside. That
freeze is why the copy could not be changed: no writer agent could reach a
word of it. The chosen voice was seeded into the site's `content_direction` on
5 August and sat completely inert for the same reason.

So the first question was never editorial. It was structural: how do you open
41 frozen pages to the platform without destroying the 23 calculators living
inside them?

## What we've done

**Recorded what every calculator computes, before touching anything.** A real
browser drives each one and writes down every answer. That baseline is the
only defence against the failure that actually matters here — rewriting words
around a calculator and silently breaking its arithmetic, which looks fine on
screen. One calculator could not be certified by the standard harness at all,
and it turned out to be the instrument, not the page: `mortgages/investor`
computes only ratios, and the harness scales every input by the same factor,
under which no ratio can change. It needed its own staggered-input capture.

**Installed the site's shared furniture** — one head, header and footer,
locked, with an installer that refuses to write them unless every asset it
references really resolves and every navigation link matches a real page. This
mattered more than it sounds: with no shared furniture, the first page opened
to the platform would have assembled against a default that links a stylesheet
this site does not have, with no header and no footer at all, and reported
success.

**Built and proved the decomposition.** All 41 pages break cleanly into
editable prose plus a locked, byte-original calculator. Two pages were taken
all the way through as tests — one guide, one calculator — and both came back
from the live site **byte-for-byte identical to what we predicted offline**,
which is what licenses doing the rest. The consolidation calculator's numbers
matched their baseline exactly afterwards, down to the pound.

**Then the owner ruled that the copy itself must come from the framework, not
from this session** — and that ruling is right, for the same reason the
hand-built shopfront was wrong: a site whose product is framework-built pages
cannot have its words typed in by hand. Thirty-nine pages of copy had been
written and machine-checked by that point. That work is now **superseded**. It
stays on disk as evidence of what the register looks like in practice, and the
loading tool has been changed to default to the *original* copy so the
superseded version cannot be shipped by accident.

**Verified the question that decides how the framework writes here.** All 41
pages are marked as owned by their source, and the platform's ordinary writer
refuses to touch an owned page — so opening them up means either changing that
marking or driving the writer a different way. The risk in changing it is that
the ordinary rebuild deletes and re-inserts every part of a page, which is
exactly how interactive tools have been destroyed before. **Measured, not
assumed:** the deletion is itself lock-aware, and against a real decomposed
page it removed both prose parts and left the locked calculator standing. The
same run proves the statement was live rather than a no-op, because it did
delete something.

## Where we are now

The structure is built and proven; the words are not written, deliberately.

Two pages are still live carrying this session's hand-written copy — they need
restarting so the framework can write them, and the command to do that is with
the owner along with the command that opens all 41 pages up. Neither has been
run yet. Everything else on the site is untouched and serving exactly as it
was.

**One thing the owner asked for is not done, and it is the sharpest question
anyone has asked in this lane.** Our calculator checks prove that a page still
computes *what it computed before*. They do not prove it computes the *right
answer*. If a calculator has been wrong since the day it was written, the
baseline faithfully records the wrong answer and every future check certifies
it. Validating the arithmetic properly means recomputing it independently —
from the standard repayment formula, from the published stamp duty bands, from
the definition of loan-to-value — and comparing. That is a real piece of work,
it is not built, and roughly half the calculators are scoring tools or
checklists where no such independent answer exists to compare against.

## Where we're going

Three things, in order.

Restart the two hand-written pages and open all 41 to the framework — the
structural change, with the original words intact.

Then settle how the framework writes here. There are two routes: change the
ownership marking and use the ordinary pipeline, which we have now shown keeps
the locked calculator but would move it to the bottom of its page unless the
page's parts are named to match; or drive the targeted section editor, which
respects ownership and position by design but needs an instruction per block
rather than per page. That choice is the owner's, and it is the difference
between "the framework writes the copy" being true and being nominal.

Then build the arithmetic validation, and be honest in it about which
calculators can be independently checked and which cannot.
