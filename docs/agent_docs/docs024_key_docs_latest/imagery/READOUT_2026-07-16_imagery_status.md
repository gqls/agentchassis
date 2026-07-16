# Imagery best-in-class — status read-out (2026-07-16)

*A spoken-word briefing: what we've done, where we are, where we're going.
Companion detail lives in the PLAN, RUNNING_NOTES, RUNBOOK and HANDOFF in this
same folder.*

---

## In one breath

We're raising the whole fleet's visual quality on a live testbed (robot-hands.com).
Three phases are finished and verified on the live site: a clean rebuild with
per-page hero images, a consistent brand layer, and sprite-sheet bullets. Along
the way the work surfaced three pre-existing platform bugs that were silently
losing content — those are written up and being fixed in their own threads. Next
up is content-linked card imagery, but the immediate priority is the content-loss
fix, not more imagery.

---

## What we've done

**Phase I0 — rebuild and per-page imagery. Done and live.**
The testbed was rebuilt into a coherent 33-page site with live news, and every
page now carries its own hero image committed straight into the site's git repo —
sixteen distinct heroes, no expiring image links. The dark "tool-portal" look was
put right, and corrupted page templates now heal themselves.

**Phase I1 — brand consistency. Done and live.**
A per-site brand guide now drives every image the platform generates, so the look
stays consistent. The logo is approved and permanently locked, and the favicon and
social-share card are derived from it automatically.

**Phase I2 — sprite-sheet bullets. Complete this week.**
Instead of loading a separate little image for every list bullet, the site now
uses one small sheet of nine glyphs, sliced by stylesheet. List bullets render as
crisp themed icons — a check, a gauge, a warning triangle and so on — from a single
75-kilobyte download. You gave the visual sign-off on the glyphs. We also built two
things that make this durable rather than a one-off:
- A self-checking mechanism that notices if a site has the sprite sheet but not the
  stylesheet that uses it, and repairs it automatically — and, importantly, does
  that exactly once and then stays quiet rather than churning.
- A "house style" mode so that ordinary article text lists pick up the themed
  bullets automatically, with no hand-editing of each page. You chose a neutral
  arrow as the default bullet, with the check mark reserved for where it's meant.

---

## Where we are right now

**Phase I2 is complete and live.** The one refinement still in flight is the
arrow default you chose: the code is written and committed, but it goes live on the
next deploy — the self-repair mechanism will swap the default from check to arrow
on its own once that deploy is out. Nothing to do but let it roll and confirm.

**Three platform bugs came to light while doing this work, and this matters more
than the imagery itself.** They were all the same shape — *a background process
reporting success while quietly failing* — and they were losing real content on
live pages across several sites:
- **Product pages shipping empty**, with the "fix" loop marking the problem solved
  while the page stayed blank.
- **Article bodies stored as unparsed data**, so five pages show raw scaffolding
  text to readers and nine more have silently lost their article entirely.
- **An image landing can wipe an article** — the very act of adding a picture to a
  page can trigger the blank-out above.

Each of these is fully documented with root cause, the affected pages, and a
recovery recipe, and each is being handled in its own dedicated thread. The
article-body and image-landing pair is the one you're now driving separately.
None of them were caused by this workstream, though our image work is what exposed
(and can trigger) the last one.

---

## Where we're going

**First, and above imagery: land the content-loss fix.** Nine live pages across
five sites are currently missing their article body, and more are one image-landing
away from the same. That repair — already in progress in its own chat — is the
highest-value open item on the board. The imagery workstream has paused any
image-landing on the affected pages until it's safe.

**Then, Phase I3 — content-linked card imagery.** Cards and listings get images
that actually match their article or entity, drawn from the same brand guide and
generation pipeline we've already built. This is the next imagery phase proper.

**Further out** sit data-accurate charts (rendered from real numbers, never
AI-drawn), news imagery, copyright-safe product sketches, performance budgets, and
an automated quality-audit loop — the remaining phases in the plan.

---

## The one-sentence version

Three phases of visual upgrade are done and live on the testbed; the same work
uncovered a class of silent content-loss bugs that are now the priority fix; and
content-matched card imagery is the next step once that's safe.
