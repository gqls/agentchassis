# Leopardess Consulting website — where we are

*A plain-language status of the rebuild. Last updated 2026-07-16.*
*Site: leopardessconsulting.co.uk*

---

## In one paragraph

We are rebuilding the Leopardess Consulting site so that everything on it is true, the branding
is coherent, and it reads like a person wrote it rather than a language model. The old site was
fluent but full of fabrications — invented staff, invented client case studies, capabilities
that don't exist. The engineering it describes is largely real; the framing was not. Over the
last several working sessions we have audited every claim against the code and database,
rebranded, forked an accessible colour palette, removed the fabrications, and rewritten the
core pages in an honest voice. The structure and chrome are now solid. The site is honest. What
remains is depth — imagery, a few interactive tools and guides, and finishing the voice pass —
plus two genuine platform bugs that surfaced along the way and are blocking imagery and one page.

**The rule that governs all of it:** no claim ships without a verified fact behind it. We check
by artifact — a live page, a database row, a screenshot — never by a "done" status.

---

## Where we've been

The rebuild has moved through a clear arc:

1. **Audit.** Every claim on the old site was checked against the actual code and database.
   The fabrications were catalogued and removed. What survived is only what we can prove.
2. **Rebrand.** A real Leopardess logo, favicon and social card; a forked colour palette
   (warm light reading surface, dark charcoal chrome, antique-gold accents) that passes
   accessibility contrast checks.
3. **Honest rewrite of the main pages.** The homepage, about, services and the rest now
   describe real systems in concrete terms — the pipeline that checks business records against
   Companies House, the news pipeline that scores what's worth trusting, the platform that keeps
   this very site current.
4. **This session: integrity and voice.** We found and fixed the last dishonest page, cleaned
   up broken links and dead pages, fixed a platform bug that was silently breaking page builds,
   and rewrote four more pages into the site's plain, honest voice.

---

## Where we are now

**The site is honest and coherent.** Concretely, as of this session:

**Integrity — the important one.**
- The **use-cases page** was the worst remaining problem: five fully fabricated case studies
  with invented clients ("Revenue Operations at a Growth-Stage SaaS Company") and invented
  results ("latency drops from days to minutes"). It now presents five honest "here's a pattern
  we could build with you, and here's the real system it's based on" cards, each openly labelled
  *"Not yet done for a client."* Zero fabrications remain.
- Four separate **phantom links** to a page that didn't exist (an AI-readiness quiz) have been
  tracked down and removed across the site.

**Structure and chrome.**
- Consistent gold header and charcoal footer across every page.
- A dead "For Leaders" link (a blank duplicate page) removed from the footer of all 17 pages;
  the page itself archived.
- The blog now shows proper summaries and reading times.
- A guide page that was silently blank is rebuilt and live.
- The favicon, logo, and social card all serve correctly.

**Voice.**
- Four pages rewritten this session — services, how-it-works, our-approach, contact — to strip
  the tell-tale marks of machine writing (the neat three-item lists, the "not X but Y" framing,
  the corporate abstractions) and replace them with concrete, plain, honest sentences. The
  worst offender appeared twice: a stock phrase — "observability, fault isolation, and cost
  controls" — that we'd literally cited as the example of what to avoid. It's gone sitewide.
- The primary navigation journey (services → how it works → use cases → contact) now reads in
  one consistent human voice.

**Platform bugs found and fixed.**
- A shared form component carried a fake example email that a safety check kept flagging as a
  fabricated contact address — which silently failed the build of *every* page using that
  component, on this site and others. Fixed once, for the whole fleet.

Nothing is in a half-broken state. Every change this session either landed and was verified on
the live site, or was deliberately held back (see below).

---

## Where we're going

In rough priority order:

1. **Imagery.** The site has its logo and a shared hero texture, but no distinctive per-page
   images, card images, or illustrations. We proved a safe way to generate and place images
   without disturbing the carefully-fixed copy — but it's currently blocked by two things (see
   "the two blockers" below). Per-card and per-section images additionally need a piece of the
   platform ("Phase I3") that hasn't been built yet.
2. **The AI-readiness quiz page.** Still blank. The content is ready to generate and the bug
   that was blocking it is fixed — it's now held up only by the infrastructure flake below.
3. **Finish the voice pass.** Most pages are done. A couple remain, and there's a content
   decision to make: four pages currently say much the same thing in different words and should
   probably be merged rather than individually polished. That's a call for you to make.
4. **The build-out the brief actually asks for.** Beyond fixing what was broken: interactive
   tools and calculators (several already exist and can be adopted), illustrated guides, a news
   surface (the pipeline is real and running — it just needs a front end), and data charts
   drawn from real numbers.
5. **SEO and social polish.** Some page titles are still the old marketing versions, and there
   is currently no sitemap.

---

## The two blockers that aren't about the website

Both surfaced during this work, both are genuine platform issues, and both are written up in
detail for a separate engineering thread:

- **An infrastructure flake.** When the system spins up a helper process to write content or
  generate an image, that helper sometimes can't reach the message bus and its reply is lost, so
  the parent job hangs until a cleanup sweep fails it 30–90 minutes later. It's intermittent and
  tied to certain machines in the cluster, not to this site. It blocked the quiz build five
  times and stalled image generation. This is the single thing most in the way right now.
  *(Full technical write-up: `docs/HANDOFF_spawn_lost_child_response.md`.)*
- **Image routing.** The system sends "hero" images to a photographic image generator, but this
  site's entire visual language is flat gold-on-charcoal illustration. So a hero here can't come
  out on-brand under the current routing — it needs to be generated as an illustration instead.
  We generated one hero as a test, saw it come out wrong, and deliberately did **not** put it on
  the site. *(Details in `HANDOFF.md` §8/§9.)*

Neither is a reason the website itself is behind; they're the reasons the *imagery* is.

---

## Where to look for more

- **`HANDOFF.md`** — the working document; open a fresh engineering session from it.
- **`RUNNING_NOTES.md`** — the full turn-by-turn record of what was done and why.
- **`AUDIT_verified_facts.md`** — the evidence base; every claim traces back to a row here.
- **`specs/VOICE_REWRITE_PROMPT.md`** — the voice guide the rewrites follow.
- **`docs/HANDOFF_spawn_lost_child_response.md`** — the infrastructure bug, for a separate thread.
