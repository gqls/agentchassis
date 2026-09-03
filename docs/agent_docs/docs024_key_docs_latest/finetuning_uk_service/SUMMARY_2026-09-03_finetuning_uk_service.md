# SUMMARY 2026-09-03 — finetuning.uk: the site becomes the playground

Milestone read-out, written to be said aloud. Previous: `SUMMARY_2026-08-18_finetuning_uk_service.md`.
This one exists because "where we're going" changed today, by the owner's direction.

## What we're trying to do

Sell a £99 fine-tuned model that a small business owns outright, and let people see it working
before and after they buy. As of today the owner has said the site should be **very much focused
around the playground tool**: the place where a visitor talks to a fine-tuned model. Other tools
can stay on the site, but the general "what else we do as a company" copy moves to
leopardessconsulting.co.uk or another of the owner's own sites. Over time the site should show
**real example after real example** of models we have trained, with before-and-after pairs, and
host those same models so a visitor can try them for **about a couple of pounds an hour, priced
at roughly five times our cost**. Details are to be talked through later; the direction is set.

## Where we've come from

July's plan designed the playground as a GPU-served chat in booked hours. Phase 0 in mid-August
proved every piece by hand: a 1.7B model fine-tuned in fifty minutes for about thirty cents, a
GPU box conversational three and a half minutes after being asked for, invoice-confirmed at $1.12
for the whole rehearsal. The two offer pages went live on 24 August, the terms, privacy and hero
images followed, and on 2 September the playground **booking** page went live. Building that page
exposed a platform defect: on sites without a plan, every repeated section got the same brief
and wrote the same section three times. That was filed as bug 443 and fixed the same evening.

## What we've done (since the last summary)

- The 443 fix is live in the chassis and **proven on this site today**: a rebuilt page reached the
  writer with a different subject for every section. The headings still repeat until the prompt
  change lands.
- The prompt change itself went through four drafts with the owner. He rejected the first three
  as instruction-shaped or AI-sounding and picked a version written in the page's own voice; then,
  seeing it rendered, asked for a more natural per-section sentence. That is now being reworked
  with the prompts thread he opened, which also received a cold-start on every prompt in the
  framework (141 of them; only 7 read the shared voice block).
- The homepage was rebuilt through the writer on his verdict that it read as AI-written. The
  negation gate repaired six of nine shapes; his one ruled sentence now serves exactly as he
  ruled it. Two calibration questions from that run went to him and are settled with the copy
  lane: cuts to a complete first clause are accepted, and the "so" clause is judged, not
  pattern-matched.
- A misspelt closing tag shipped live from the writer and nothing on the save path checks
  markup; filed as bug 456 with the structural fix first.
- Subjects are backfilled for three of the four affected pages; the playground link is on the
  offer page for its next rebuild.

## Where we are now

The site serves the £99 offer, the technical page, and a playground page that **describes**
the hour but has no playground in it. The owner has decided the playground is both a public
demo, always on, served from the small model on the cluster's own CPU at no model-fee cost, and
booked hours on a GPU at about 35 cents an hour real cost. None of the three pieces that make the
chat real exist yet: the chat box on the page, the route in the tools service, and the model
server behind it. The library has a chat box to copy, the tools service has the guards, and the
GPU provisioning already runs as workflow actions for training. The prompt change that makes
repeated sections distinct is one text decision away from applying.

## Where we're going

1. Build the playground chat: load and time the demo model, add the route, put the chat box on
   the page through the tool pipeline, turn booked-hour provisioning into a workflow, connect
   booking to session. Everything through the framework; nothing hand-authored.
2. Refocus the site around it: the tool prominent on the homepage; company-general copy moved
   to the owner's other sites; a growing set of real before-and-after examples, each with its
   model hostable for a paid hour.
3. Land the prompt change and rebuild the two offer pages so their sections read as three
   different sections.
4. Pricing for the hosted hours, once the details conversation happens: five times cost is the
   stated posture, and the measured cost is about 35 cents a GPU hour plus warm-up.
