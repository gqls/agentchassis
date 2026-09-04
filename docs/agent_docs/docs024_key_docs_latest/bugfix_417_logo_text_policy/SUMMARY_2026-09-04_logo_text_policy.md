# SUMMARY — 2026-09-04 — the logo text policy lane

*Written to be read aloud. Previous in the series: `SUMMARY_2026-09-03_logo_text_policy.md`.*

## What we're trying to do

Two things now, and they became separable yesterday.

The first is the original job: **stop the system inventing brand names.** Our planner used to send
image models an instruction that quietly permitted a logo made of letters, without ever saying what
those letters should be, so the model made a name up. We want logos that are pictures, with the real
name set in ordinary text beside them where we control it.

The second arrived out of the first and is now the bigger one: **nothing in the system checks
whether a logo can be seen.** A logo can be generated, have its background cut away, pass a
safety check, be stored, published and placed on every page — every gate green — and be invisible
to a human looking at the page.

## Where we've come from

The wordmark fix is live and has been verified by looking at the pictures rather than at any status.
The transparency fix from another team is live too. Yesterday four sites generated new logos under
both, and none of the four has a single letter on it.

The visibility problem was filed yesterday, and then reproduced live within hours: a site's
regenerated logo came back **white on a white header**, with the only visible part being leftover
traces of the temporary background colour used during cutting. Every measurable signal improved and
the logo became less visible. You ruled that we should detect this after the fact rather than refuse
to save it, and that the site in question keeps the logo it has.

## What we've done

**Built the check, and run it over every logo we own.** It fetches the logo the page actually loads,
reads that site's own header colour, and measures how much of the mark stands out against it.

It needs two tests, not one, and the reason is worth knowing because it is the whole difficulty in
miniature. The obvious test — "is any part of this dark enough to see?" — is passed easily by the
worst logo we have, because the thin coloured outline left around it is very high contrast, while
eighty-six per cent of the mark is white on white. So the second test asks *how much* of the mark
reads, not whether its best pixel does. One test looks at the best part; the other at how much of it
is any good.

It can prove itself without touching the cluster. We kept copies of that site's logo from either
side of its regeneration — the only copies that exist — and the check is wired to run against both,
plus a pair of made-up cases that are the same white shape on a dark header and a white one:
identical files, opposite answers, which is what proves it is really reading each site's own colour
rather than assuming white.

## Where we are now

**The check works, it catches the logo it was built to catch, and it found a second one — but the
number that matters is not the number of bad logos.**

We have thirty-four logos. Two fail. **Only seven could be judged at all.** Twenty-two of them have
a background baked into the picture — the older ones, made before the transparency fix — and for
those, "does it stand out against the header" is the wrong question, because what you see is a
coloured box and the mark reads against that box. I looked at several and they are fine. So this is
not twenty-two hidden problems; it is the reason I cannot yet tell you how widespread this is. The
check can currently speak for seven sites, and it will speak for more only as old logos get remade.

The second failure is mortgagecalculator.co.uk, and I want to under-claim it deliberately: no part
of that mark reaches the accessibility contrast standard against its header, but I opened the image
and it is a gold house-and-key on cream that you *can* see. "Below the standard" is the honest
sentence; "invisible" is not.

> **CORRECTED later the same day, and it changes what can be DONE about it rather than the finding
> itself.** I checked where that logo came from, after writing the above. **Nobody generated it —
> a person uploaded it**, and there is no prompt recorded against it at all. So "whether that is
> worth regenerating is your call", which is what I say further down, quietly assumes a remedy that
> may not exist for this one: there is nothing to regenerate *from*, and remaking it would replace
> someone's deliberate choice with a model's guess, permanently. **The honest question for this
> logo is not "regenerate it?" but "does someone want to supply different bytes?"** The measurement
> stands; what I implied about the fix does not. It also means the two failures are not two
> failures of the same thing: only one of them came out of our own pipeline.

Three other sites have a logo we made that their pages never show — they still print the site's name
as text in the header. That is a different problem, and the check now names it instead of quietly
measuring a picture nobody sees.

**On the original wordmark question, one thing changed that yesterday's summary got wrong.** It said
the one site that would really test our override had never produced a logo. It had — eleven minutes
before that summary was written. The result was clean, and it is the best artefact of the week. So
the test case exists now and it passed; it is a single run, which refutes "the override never wins"
and settles nothing about how often it does.

## Where we're going

**One decision is yours and it is the only thing blocking the rest.** The check reports; it does not
yet hand anything to something that can act. The existing route for contrast problems sends them to
a tool that repaints stylesheets, and no amount of restyling fixes a pale picture. Naming the right
destination is the next step and it needs a person.

**One piece of work is already scoped.** What we built is deliberately the cheap version: it trusts
the colour the site *says* its header is. That is true today and goes stale silently the moment a
site's colours change — and we have a live mechanism that rewrites site colours without anyone
asking. So a logo that passes today can quietly keep passing after it has stopped being legible.
The version that stays right reads the colour off the rendered page. We wrote the cheap one so its
measurements and thresholds sit in one place for the proper one to reuse.

**And one thing keeps being true.** Every real defect this lane has found, on three separate days
now, was found by opening the picture — and every one of them had a green status beside it. Today
that included three of my own measurements: each was careful, correct, and taken of the wrong image.
