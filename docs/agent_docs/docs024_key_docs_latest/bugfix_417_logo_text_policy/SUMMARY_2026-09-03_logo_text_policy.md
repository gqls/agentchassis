# SUMMARY — 2026-09-03 — the logo text policy lane

*Written to be read aloud. Previous in the series: `SUMMARY_2026-08-31_logo_text_policy.md`.*

## What we're trying to do

Stop the system inventing brand names. When our planner asked an image model for a logo, the
instructions it sent quietly *permitted* a wordmark — a logo made of the company's name in
letters — without ever saying what that name should be. So the model made one up. Sites were
shipping logos with a fictional brand painted into the picture. We wanted logos that are pictures
only, with the real name set in ordinary text beside them, where we control it.

## Where we've come from

We found the licensing phrase, replaced the instruction that carried it, and added an override that
explicitly tells the model to produce a text-free mark and to ignore any earlier wording that
permits text. That change was reviewed, approved and went live. The bug stayed open on our own rule
that a fix is only closed once it is proven working in production, not merely shipped.

Alongside it, a second team found that logos were coming back with an opaque background instead of a
transparent one, because transparency is not something an image model can be asked for. Their fix
paints a known background colour and cuts it away afterwards. The two fixes only became live
together this morning.

## What we've done

Verified the fix at the artefact rather than at the status. Four sites generated new logos today
with both fixes running. **Three produced a picture, and none of the three has a single letter on
it** — a magnifying glass over a woven lattice, an abstract maze, and a chevron. In two of those the
permitting phrase was still sitting in the prompt next to our override, so the model had to choose,
and it chose correctly.

We also found and filed a new problem, and we found it only by looking at the pictures.

## Where we are now

**The thing we set out to fix appears to be fixed, and the honest version of that sentence is
weaker than it sounds.** Eight generations have come back clean and none has come back lettered.
But eight is a small number — it cannot tell "reliable" apart from "fails one time in five" — and
more importantly, only one site in the estate still has a permitting phrase in its plan, and that
site has never successfully produced a logo. So the case that would really test the override has
not run yet.

We now know the exposure honestly for the first time: **thirteen of thirty-three sites** will
compose a logo prompt that depends on the override winning an argument. Getting to that number took
three attempts, because twice we measured the rule by searching for a word the rule itself contains.

**And a second, separate problem is now the more urgent one.** Nothing anywhere in the system checks
whether a logo can actually be seen. A logo can be generated, cut out, safety-checked, stored,
published and placed on every page — passing every gate — and be invisible to a human. We filed that
today and then watched it happen live: a regeneration produced a white shape on a white header, with
the only visible part being leftover traces of the temporary background colour. Every measurable
signal improved; visibility got worse.

## Where we're going

Three things, in order.

The first is a decision, not work: whether to fix the visibility gap by refusing an illegible logo at
the moment it is created, or merely by reporting it afterwards. Only the first actually prevents it,
and it has to measure the mark against the header it will sit on, after the background is removed.

The second is to wait for the one site whose plan still carries a permitting phrase to produce a
picture. That single result is worth more than the eight we already have, because it is the only one
that tests what we changed.

The third is smaller: one site is currently serving a logo that is worse than the one it replaced,
and there is no undo — regenerating deletes the previous file. We hold the only surviving copy of
the old one, and putting it back needs a decision.
