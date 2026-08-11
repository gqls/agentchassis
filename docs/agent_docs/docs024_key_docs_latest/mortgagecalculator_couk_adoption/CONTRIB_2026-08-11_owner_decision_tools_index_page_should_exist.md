# CONTRIB — owner decision (2026-08-11): mortgagecalculator.co.uk SHOULD have a `tools-index` landing page

**From the `bugfix_214_imagery_scope_ref` lane, routed here because this site is yours**
(who-owns + live-transcript check both say so; several sessions active in this estate
today). We are not dispatching anything at your site. This file hands you an owner
decision plus the evidence and the traps around executing it — sequencing is yours.

## The decision

Asked 2026-08-11 (in the 214 lane's session, options presented explicitly): *what should
happen to the `tools-index` imagery reference that names a page existing under no
spelling?* **The owner chose: ADD THE TOOLS LANDING PAGE** — over deleting the orphan
imagery row, and over leaving it undecided.

## The evidence that framed it (all measured live 2026-08-11)

- The current plan (2026-08-02) holds **ten `tool-*` pages** — the site's core product —
  and **every other family has its section-index page** (`guides-index`, `about-index`,
  `contact-index`, `investor-index`). **Only the tools family has no landing page**, in
  the plan and in `pages`.
- Two artefacts already point at the missing page, i.e. the planner *intended* it:
  - `site_plan_imagery` row `scope='page', scope_ref='tools-index', key='hero_tools'` —
    the **one** remaining consumer-invisible imagery ref fleet-wide (the 214 backfill
    repaired the other nine and was ordered to leave exactly this one).
  - work item `needs_imagery:page:tools-index:hero_tools`, **deferred** — ⚠ if your
    queue is ever undeferred **before** the page exists, this generates a **paid image
    referenced by nothing** (the `bugs_open/114` waste, pre-booked). No `hero_tools`
    asset exists today (checked).
- Live: `/tools/affordability/index.html` serves 200; `/tools/index.html` 404s — and so
  does `/guides/index.html`, so the index-page gap is wider than tools (consistent with
  your own commit note that "the port lane's 114 finding extends to our tools").

## What the executor needs to know (from our lane's scars, not to teach yours)

1. **`recompose_pages` can silently no-op** (LANDMINES ~§8441): seed 362 makes the
   planner re-emit built pages verbatim and it is never told the recompose list — put
   the intent in **both** the spec AND the briefing prose, and verify by the error-code
   query, not by eyeballing.
2. **The imagery side now takes care of itself.** Bugfix 214 is live (v1.0.1286,
   council-approved corr `46a50b4c`): on any replan, imagery refs resolve against the
   plan's own page names — proven in production on fundamentallyai's unforced replan
   (`imagery_refs_canonicalised: 2`). Once a plan contains `tools-index`, the hero ref
   resolves with no action from anyone; the fleet census of invisible refs should then
   read **0**.
3. **Do NOT re-run `sql_for_agents/373`** — already applied; its guard asserts exactly
   1 unresolvable row *remains* and will abort (in the right direction) once your page
   makes that 0.
4. Two more of your deferred items (`about`/`contact` heroes) carry pre-canonical keys
   (`about` vs the plan's `about-index`) — the ItemKey-drift note in
   `imagery/RUNNING_NOTES_imagery_best_in_class.md` (2026-08-11) covers re-key vs
   cancel; it interacts with whatever replan route you choose.

## Status in the 214 lane

This was our last open item. Our bug file (`bugs_open/214…`, STATUS 2026-08-11 section)
now records the decision as MADE and routed here; the item closes for the fleet when a
`tools-index` page exists and the invisible-ref census reads 0. Questions →
`docs024_key_docs_latest/bugfix_214_imagery_scope_ref/`.
