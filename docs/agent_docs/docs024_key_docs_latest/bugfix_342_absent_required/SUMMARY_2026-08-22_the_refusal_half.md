# SUMMARY 2026-08-22 — bugs_open/342: the refusal half

## What we're trying to do

Stop pages losing content silently. When a page section is built from a template, the template
names the pieces of text it needs — a headline, a body, a call to action. If the writer never
produces one of them, Go's template engine does not complain: it renders that spot as an empty
string. The page assembly step then throws away any section that looks visually empty. So the
content does not arrive broken and get noticed; it does not arrive at all, and nothing anywhere
says so. This is the mechanism behind an earlier fleet-wide incident where article bodies went
blank across the estate.

## Where we've come from

A previous lane fixed the loudest part of this in the days before we picked it up. The render
seam — the single place every render passes through — now checks the component's declared
contract and reports, at error level, which required fields came out empty. Nine of the fifteen
render call sites pass it the contract to check against; the other six have mechanism reasons
not to, each verified individually. Two paths that write straight onto a live page also file a
work item when it happens.

That work was approved and shipped. What the bug file said it stayed open for was blunt: *"no
refusal was added anywhere"*. Everything built so far notices and reports. Nothing declines.

## What we've done

First, we corrected the record. The bug file still said one half of the previous lane's work
was waiting for a deployment; it had shipped, and we proved that at the running binary rather
than at the version tag. We also re-measured the populations the file quoted, and one of them
had been badly misread by two separate documents: the file warned that ~75 components with no
declared contract would be "the hard part", but 95 of the 100 such components today are
self-contained tools, which have no contract by design. The genuinely uncovered set is five
components, one page each. That correction is now a landmine entry so the next reader does not
inherit the same wrong number.

Then we built the refusal. The page-editing routes now decline to save a section whose required
fields rendered empty, leaving the live page exactly as it was; the work item is still filed
first, so refusing never costs us the record. It sits at the single point where every edit is
saved, so future edit types inherit it rather than having to remember. It is switched off by
default and turned on per agent by a migration held back until the code has actually deployed,
which is this estate's rule for giving the system new authority.

The review council pushed back twice, and both rounds improved the change. The first found that
arming *detection* on the site-header-and-footer path while only the page editor got
*protection* would reproduce this very bug on the sibling path — so that path now has the same
protection, sharing the identical decision code, deliberately switched off because nothing can
trigger it today, with the signal that says it is time written down. The second round approved
it and still found three things worth fixing, including a property of our own gate we had not
stated: it does not run on the no-op refresh path, which is right for refusing and a genuine
blind spot for detecting.

## Where we are now

Approved and committed, across seven commits. The site-chrome reporting half is armed in
production and verified by reading the live configuration rather than trusting the migration's
own check, with a negative control confirming the refusal remains armed nowhere. The refusal
code is on the branch and inert until the next chassis build.

## Where we're going

Three things close this bug. When the next build rolls, apply the held-back migration and run
the canary — one edit that must be refused with the live page byte-identical, one clean edit
that must still save, and a reading of the job's own reported status, which we expect to be
wrong in a known way until a separate bug is fixed. Someone owns the five contract-less
components, either by writing minimal contracts or by recording an explicit decision not to.
And the site-chrome refusal gets switched on if and when its named signal fires — not on a
schedule, because until then it is a switch with nothing behind it.
