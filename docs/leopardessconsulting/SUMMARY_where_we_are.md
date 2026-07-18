# Leopardess Consulting website — where we are

*A plain-language status of the rebuild. Last updated 2026-07-18. All figures below were
checked against the live site and database on that date, not carried forward from earlier
notes.*
*Site: leopardessconsulting.co.uk*

---

## In one paragraph

We are rebuilding the Leopardess Consulting site so that everything on it is true, the
branding is coherent, and it reads like a person wrote it. The old site was fluent but full of
fabrications — invented staff, invented client case studies, capabilities that don't exist. The
engineering it describes is largely real; the framing was not. The site is now honest, has a
consistent voice, and has just gained its first proper graphics. Two automated checkers now
watch it: one for unverifiable claims, one for machine-sounding prose. One significant problem
remains, and it is not on the website itself — something else in the platform keeps overwriting
the pages we fix.

**The rule that governs all of it:** no claim ships without a verified fact behind it. We check
by artifact — a live page, a database row, an image we have actually looked at — never by a
"done" status.

---

## Where we've been

1. **Audit.** Every claim checked against the real code and database. Fabrications catalogued
   and removed.
2. **Rebrand.** Real logo, favicon and social card; an accessible palette of warm light reading
   surfaces against dark charcoal chrome with antique-gold accents.
3. **Honest rewrite.** The main pages now describe real systems in concrete terms.
4. **Voice.** A first pass stripped the marketing tells. A second pass, after your review, made
   the register plainer and friendlier: short sentences, contractions, no literary flourishes.
5. **This week: graphics, checkers, and one hard blocker found.**

---

## Where we are now

**The site now has real graphics.** Four infographics, each generated and then reviewed by eye
before going anywhere near the site:

| Page | What it shows |
|---|---|
| Homepage | Three columns: what we've built, what we could build with you, how an engagement starts |
| How it works | The six stages a job passes through, with "a person decides" drawn largest |
| Technical architecture | The agent hierarchy sitting on Kubernetes, Kafka and Postgres |
| What we've built | **The Leopardess Line** — an Underground-style route map of the three running systems |

Every figure in them (2,767 records verified, 937 enriched, 5,652 items collected, 4,672
scored, 8 sites) comes from our own database. Each is placed with a full written description
attached, so the pages still make sense to a screen reader or with images turned off.

**Three pages now have proper hero images** — the homepage, who we help, and how we work. All
are text-free abstract illustrations in the house style.

**Two automatic checkers are live.** One flags claims that can't be traced to the evidence
base. The other flags copy that reads machine-written, measured against the site's own voice
rules — including the words you asked us to stop using ("trust", "honest", "earns its keep").
Neither ever rewrites anything: they raise it for a person to rule on. There are currently 25
voice findings waiting, and they double as the to-do list for finishing the rewrite.

**The tools work, mostly.** Four of the five interactive tools are functional and are now
linked from the footer, which they weren't. One — the LLM cost calculator — is genuinely
broken: it loads the wrong file. That's diagnosed and written up.

---

## The one significant problem

**Something else in the platform keeps rebuilding this site's pages and undoing our work.**

It happened twice in twenty-four hours. On Friday it rebuilt the homepage and reinstated a
fabricated statistic and invented case-study titles. On Saturday it rebuilt the services page
and *invented a link to a tool that doesn't exist* — which is the blank page you clicked.

This matters more than any single defect, for two reasons. It doesn't just lose our
corrections, it actively puts fabrications back. And it means anything we fix by hand has an
undefined shelf life until it's addressed. It is written up as a platform bug with the fresh
evidence attached.

One useful discovery: images attached properly to the site plan **survive** these rebuilds,
while page copy does not. So imagery work is durable in a way copy work currently isn't — which
is why we've prioritised the graphics.

---

## Where we're going

1. **Fix the rebuild problem.** Everything else is provisional until this lands.
2. **Finish the imagery.** Five pages still have no image at all: about, services, use cases,
   contact and the blog index. Two of them need a small platform change first, because their
   page templates have nowhere to put an image.
3. **Replace one bad image.** A garbled picture is still the fallback on six pages. Replacing
   that single file fixes all of them at once.
4. **Finish the wording pass** using the 25 findings the checker produced.
5. **Repair the LLM cost calculator.**
6. **The build-out** — more tools, illustrated guides, a news surface.

---

## Two things worth knowing about how this now works

**We were wrong about image generation, and corrected it.** We had concluded that AI image
tools simply cannot render readable text, and had planned to build a whole drawing system to
work around it. You showed us two Gemini infographics with perfectly clear text, and the
correction turned out to be embarrassing in a useful way: the capable model was *already
wired into our platform* and had never been used, because one setting sent hero images to an
older model that genuinely can't do text. The fix was a better instruction, not a new
subsystem. The wrong turn is recorded rather than quietly deleted.

**The instruction matters more than the model.** The same model produced gibberish from a vague
prompt and publication-quality work from one that specified the layout, the exact wording, the
permitted figures and the colours. The prompts that produced these four are kept, so this is
repeatable.

---

## Where to look for more

- **`HANDOFF.md`** — the working document; open a fresh engineering session from it.
- **`RUNNING_NOTES.md`** — the full turn-by-turn record, including the wrong turns.
- **`AUDIT_verified_facts.md`** — the evidence base behind every claim.
- **`PLAN_imagery_and_design_2026-07-18.md`** — the plan for the graphics work.
- **`/bugs_open/`** — the platform defects: `001` (the rebuild problem), `003` (an
  infrastructure flake that stalls image generation), `011` (the image-routing bug).
