# The instrument that was dark — loancalculator.co.uk, 2026-08-24

## What we are trying to do

Run a live loan-advice site whose product *is* its calculators, entirely through the
framework — planned, built, rebuilt and checked by machine — while proving, on demand and
at the artefact, that each calculator still computes what it is supposed to compute. The
proving part is the hard half. Nothing in the platform checks that a tool returns the
right *numbers*: the standard checks confirm that a tool's controls exist and that
something changes when you drive them. A rewritten calculator returning subtly wrong
monthly payments passes both.

## Where we have come from

The site was adopted hand-built and taken apart into framework components over several
weeks. Along the way it acquired a bespoke acceptance harness — `toolgolden` — which
drives every calculator in a real browser with deterministic inputs and records the
answers, so a rebuild can be proved equivalent rather than assumed to be. That harness is
the only instrument on the estate that checks arithmetic.

A week ago a re-plan invented fourteen duplicate pages; that was cleaned up and the cause
fixed. Yesterday the owner gave four instructions — delete the duplicates, release the
queued rebuilds, restore the Guides page and its menu link, and retire one of two
duplicate calculator pages. All four were completed and verified. But the session that did
them recorded, at the end, that the acceptance harness was down: it timed out identically
on pages that had been rebuilt and on pages that had not. It shipped ten live rebuilds
with its only value-level instrument dark, and said so plainly.

## What we have done

**Fixed the harness, at the seam where it was actually broken.** It was not broken at all,
in the sense of needing repair to its logic. It launches a private browser and hands it a
scratch directory taken from the environment; the sandboxed browser on this machine cannot
write into directories whose names begin with a dot, and the environment now hands out
exactly such a directory. The browser aborted before it could be spoken to, its own
explanation was being discarded, and the wait was blind — so a solvable configuration
problem surfaced as thirty seconds of silence and the words "chromium did not start". The
launcher is shared by six checking tools across four projects, so one setting takes all of
them out at once and each reports it as a fault in whatever page it happens to be holding.

**Given it a way to test itself.** It now builds a small calculator whose answer is known
in advance, drives it exactly as it drives the real pages, and states whether it is fit to
be quoted. The expected answers are worked out by hand from the driver's own rules rather
than recorded from a run, so it cannot pass by agreeing with a harness that is wrong about
everything — and we confirmed it can fail, by breaking the fixture's sums on purpose and
watching it go red.

**Used it, and found a real fault on a live page.** Ten of the eleven calculators
reproduce their recorded answers exactly; the only differences anywhere are one cosmetic
renaming of an FAQ container that every page shares. The eleventh, "Pay Off Loan or
Save?", had the calculator on it twice, with the lower copy dead. Yesterday's rebuild had
moved the protected calculator up the page as intended and *also* left an unattached,
byte-identical second copy at the bottom.

**Repaired it, and proved the repair at the artefact** — the page the public gets is now
byte-for-byte the file the framework generated, with no duplication, and the harness reads
it exactly as it read it before the damage.

**Re-recorded the reference snapshot** across all eleven calculators, and then compared
back against it: eleven of eleven exact. A snapshot nobody has compared against is a file,
not a baseline.

**Declined to guess at the cause.** The obvious explanation is a bug closed three days ago,
originally found on this very page. One query rules it out. The automated diagnosis loop,
which is the estate's route for turning a hunch into a finding, could not reach a verdict
on files this large. So the case is written up with the damage measured three ways and the
cause marked unknown — rather than a plausible story that would have read as fact.

## Where we are now

Twenty-eight pages, every one loading, no broken addresses, the Guides link on all of
them, and all eleven calculators verified. The instrument that checks them works, says so
before it passes judgement, and is now proof against the environment fault that blinded
it. The trap itself is written down where the other four projects that share the launcher
will meet it.

One open case: we know precisely what the framework did wrong on that page and can detect
it anywhere on the estate in one query — we do not yet know why it did it, and the writer
that did it is still live.

## Where we are going

Find that cause. The next move is written into the case file: the plan for the page
demonstrably does not list the calculator twice, so something between the plan and the
save added the extra copy, and the merge of protected rows into the page's cached section
list is the first place to look.

Beyond that, the wider prize is the one the harness was built for. The checks it performs
by hand can be handed to the platform, which would then run them unprompted on every
acceptance sweep, for ever. That is worth more than any single rebuild it has ever
verified — and today is the argument for it, because the fault it found had been live and
invisible for a day for want of an instrument that was working.
