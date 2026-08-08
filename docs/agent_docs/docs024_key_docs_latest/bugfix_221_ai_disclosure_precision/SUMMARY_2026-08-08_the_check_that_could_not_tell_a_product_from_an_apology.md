# SUMMARY — the check that could not tell a product description from an apology

2026-08-08. Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md`, the evidence in `NOTES`.

## What we're trying to do

Stop a safety check from refusing correct pages. We run a gate over every page
before it is saved, looking for the sort of text a language model produces when
it gives up — *"As an AI, I cannot generate this listing."* That check is worth
having; it was built because exactly that apology once shipped as live copy. The
job here was to make it stop firing on ordinary English.

## Where we've come from

The check looked for the letters `as an ai` anywhere in a page's visible text.
On webdesign.co.uk's tools index there is a tool described, correctly, as
*"LocalBusiness schema, as an AI-builder prompt"*. The check found `as an ai`
inside `as an AI-builder` and refused the page.

Refusing is not a warning. It makes the step fail before the page is written, so
**the entire page could never be rebuilt** while that sentence was on it — not
the sentence, the page. Nobody had noticed, because it costs nothing until
somebody asks for a rebuild.

This was the second bug of its shape in a week. The first (`219`) was about
*where* the check looked — it read code comments and convicted a human's
changelog note. That was fixed. This one was about *what* it looked for, in text
that is genuinely visible prose, where no change of scope could help.

## What we've done

Made the check look for the **sentence shape** rather than the letters: an AI
noun phrase followed immediately by the first person. *"As an AI, I cannot…"*
still fails. *"as an AI-builder prompt"*, *"as an AI agent"*, *"as an AI
product"* no longer do.

The case that shaped the design is *"As an AI engineer, I built this"* — a
person's own bio, containing both the phrase and a first-person "I". It must not
be refused, and it is what forced the rule to be about the construction rather
than about the two things appearing near each other.

We wrote the test **before** the fix and ran it against the unfixed code: every
"must not refuse" case failed, every "must still refuse" case passed. That
ordering is what tells you the test can fail and that the change only narrows.
Then we deliberately broke the fix three ways to check each part was load-bearing
— including removing the first-person requirement, which immediately re-convicts
the human bio.

It went through the review council and was approved first time, with four
advisory comments. Two of them found real gaps and both are recorded.

## Where we are now

**Live and proven on chassis v1.0.1268.** The page's stored copy is unchanged and
still contains the sentence; the shipped code now passes it. The apology it was
built for still fails, tested against the real page with the apology injected
into it.

The proof needed some care. The house rule for showing a change reached
production is to find a string it added *and* one it removed — the second is what
proves you are looking at your own change. This change removes nothing; every
phrase it touches survives by design. Rather than nominate some string as
"removed" and print a reassuring zero, we started a throwaway container on the
*previous* image and showed the marker is absent there and present on both live
replicas — with a third string present in both, so that the absence is a real
absence and not a mistyped command.

Two things we found while proving it. The affected page is sitting at
`needs_rebuild`, so this was not a hypothetical trap — a rebuild is already
queued for the page that could not have been rebuilt. And the fleet is uniformly
on the new image, which had to be checked rather than assumed, because a mixed
fleet caught out another team the same day.

The team that owns the affected site has been told, in the terms that matter to
them: their page was unbuildable, it now is not, and their copy did not have to
change and should not.

## Where we're going

Nothing is outstanding on this bug. Three things sit beyond it.

**The wider question is untouched, on purpose.** One reviewer pushed hard on it:
fixing this one check leaves the general mechanism alone — any of our
blocker-severity text scans can still wedge a page's rebuild forever. There is a
near-identical bug open against a different checker right now, where a comment
saying *"no fabricated data"* is convicted for containing the words "fabricated
data". That belongs to another team who are working it, and how all these scans
should behave is a design decision this bug should not be allowed to settle
alone. The challenge is recorded rather than argued away.

**The other thirteen patterns are unchanged**, still matched as plain substrings
at the same severity. An article about data schemas could still trip one. The
trap is now written down where somebody adding a pattern will meet it.

**One loose end:** the automated check on that written-down trap could not run —
the fleet is out of API credit. The command to re-fire it is recorded, and its
verdict is weak evidence either way for this kind of entry, so it is a tidy-up
rather than a risk.
