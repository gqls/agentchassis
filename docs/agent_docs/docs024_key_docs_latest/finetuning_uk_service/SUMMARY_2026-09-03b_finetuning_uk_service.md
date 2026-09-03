# SUMMARY 2026-09-03b — finetuning.uk service: the playground demo is live, end to end

Written to be read aloud. Current state only; the chronology is in NOTES and README_where_we_are.
Supersedes `SUMMARY_2026-09-03_finetuning_uk_service.md` (written this morning, before the route,
the model host or the widget existed).

## What we're trying to do

Make finetuning.uk the home of one thing: a fine-tuned model you can try. A public demo anyone can
chat to for free on the site, booked GPU hours for a customer's own model at a price that covers our
costs several times over, and, later, a catalogue of real examples with before-and-after pairs that
visitors can try too, including models submitted by third parties with a page of their own. The
owner's words, 2026-09-03: "I'd like the finetuning site to be very much focused around this tool."

## Where we've come from

Until this morning the playground page described a process and had no playground. The pieces existed
as designs: a route in the tools API (council-approved, unshipped), a model file from the Phase 0
training run, a library chat box we assumed we could reuse. Two offer pages had the same section
written three times under one heading, a defect (443) whose fix waited on a prompt change (641)
the owner had not yet read.

## What we've done today

- The demo model runs on a Hetzner box, three times faster than the cluster's CPU, reachable only from
  the tools API's island.
- The chat route is live on the island. The owner ran the deploy; one restart was wasted because my
  own recipe omitted the compose file's env block, and that trap is now written down for the next
  tenant.
- The library chat box turned out not to fit (one message at a time, same origin, no streaming). The
  owner chose the framework-generated path; the generator wrote a widget to the route's exact contract,
  refused once because the playground page was a content page and the framework will not re-type a
  live page on its own, then attached it as the page's second section after that one-column decision.
- The widget is live and proven in a real browser: a question typed on the served page came back as a
  streamed answer from the model. Nothing on the page calls anywhere but our route.
- Prompt change 641 went live after the owner read it; both offer pages were rebuilt. Every heading is
  now different. The three middle sections on each page still say one thing under three headings, and
  the cause is known and confirmed by the prompts lane: the whole page brief is rendered into every
  section's prompt as that section's own guidance. It is theirs to fix, as a diagnosis, not a patch.
- The technical-details brief no longer asks for the three-model listing the owner called unhelpful.
- The owner's four decisions are recorded: generated widget first, hand-written as fallback; pricing by
  GPU class as a choice, price to rise; the examples catalogue shape drafted for his line-by-line
  reaction; the company-general copy is likely heading to a family of narrower sites (proverb first),
  and leopardess has been told.

## Where we are now

A visitor to finetuning.uk/playground.html can talk to a fine-tuned model, for free, today. The page
around it is still yesterday's booking copy; a companion guide page exists but nothing links to it.
Two offer pages read better than they did and still repeat themselves in the body. The demo box
serves one small model on a CPU; the route can carry a `model` choice later without changing shape.

## Where we're going

Next, in order: a criteria fence so the tool is graded like every other tool; a multi-turn browser
test; the owner's read of the widget's words and of the guide page; a proper playground page brief
with the tool at its centre; then the catalogue's first shape (ours only, no accounts) and the
booking-to-session handoff with the GPU class as a choice. The repeated-body defect waits on the
prompts lane's diagnosis, and the two pages get rebuilt again when that lands.
