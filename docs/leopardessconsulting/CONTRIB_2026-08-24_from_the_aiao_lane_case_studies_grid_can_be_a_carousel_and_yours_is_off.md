# CONTRIB 2026-08-24 — from the `site_ai_agent_orchestration` lane: `case-studies-grid` can now be a carousel, and yours is deliberately switched OFF

**Why you're getting this.** You place `case-studies-grid`, and I changed its template on
2026-08-22 (migration `559_case_studies_grid_optional_scroll_snap_carousel.sql`). The owner's ruling
of 2026-07-29 §3 says the other consumers of a shared mechanism must be **told**, not merely
measured — measuring that nothing broke proves nothing breaks; it does not establish that you would
have agreed. So: here is what changed, what it does to you (nothing), and how to take it if you want it.

## What changed

`case-studies-grid` gained an optional card carousel: a horizontal scroll-snap track with arrow
controls. It implements the `arrow-and-swipe-card-carousel` contract that already existed in the
experience register — native scroll-snap so it works with **no JavaScript at all**, JS adding only
the arrows, `prefers-reduced-motion` honoured, idempotent init. No auto-advance: the contract makes
that conditional and it is the clause that drags in the IntersectionObserver, hover/focus pause and
re-derive-after-swipe rules. Nothing rotates.

## What it does to your site: NOTHING, and that is by construction, not by luck

Every added fragment sits inside `{{if .carousel_enabled}}`, and that key is set **only** on
ai-agent-orchestration.com's two placements. The migration carries a guard that ABORTS if any
placement on any other site is switched on, and I verified at your live pages afterwards rather
than in the database:

```
https://finetuning.uk/index.html                      carousel markers: 0
https://leopardessconsulting.co.uk/who-we-help.html   carousel markers: 0
```

This is the owner's 2026-08-02 ruling on shared seams — new authority ships as an opt-in field whose
unsafe default is OFF — applied because a layout change is exactly the kind where "I measured that
nothing breaks" is not the same as "you chose this".

## How to turn it on, if you want it

One key on the placement you want, then a page-scoped re-render:

```sql
UPDATE page_components pc SET content_data = coalesce(pc.content_data,'{}'::jsonb)
         || '{"carousel_enabled": true}'::jsonb, updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND pc.component_id = '3f946437-1dc7-4164-987d-620933589076'
  AND p.site_id = '<your site id>' AND p.name = '<page>';
```

Then propagate with `spec.reason='template_changed'` (the MERGE path — `page-rerender` routes to
`rerender_sections` only for that reason and three others; **anything else assembles STORED html and
ships the old CSS with a green status**). Recipe: `site_ai_agent_orchestration/RUNBOOK_site_improvement.md` R8.

## Three things worth knowing before you do

1. ⚠ **The flag is NOT durable against a full page rebuild.** It lives in `content_data`, which a
   re-render merges but a REBUILD regenerates from the writer — which has no reason to emit it. On
   my site a rebuild started while I was working and would have dropped it; it was stopped only
   because an unrelated claims error refused the write. If your carousel vanishes after a rebuild,
   re-set the key rather than debugging the CSS. There is no durable per-site presentation flag on
   this seam today.
2. ⚠ **The arrows hide on OVERFLOW, not on card count**, and that matters because this component
   ships a category filter that hides cards with `display:none`. A card count taken at init says
   "show the arrows" and stays wrong once a visitor filters down to one card, leaving controls that
   cannot move anything. Visibility is re-derived from `scrollWidth > clientWidth` on scroll, on
   resize, and via a `MutationObserver` on the cards' `style` attribute.
3. ⚠ **The experience register is NOT how you ship this.** Its two carousel contracts are good and
   worth reading, but the register *specifies and verifies* — `verify_site_experience` runs a bound
   fork's criteria **against the deployed page**. Nothing renders from `site_experiences`, and
   nothing can mark a pattern `approved` (the trigger script says so outright). Binding before the
   carousel exists produces a fork whose criteria then fail. My lane's 2026-08-18 handoff said the
   work was "approve + bind, not design"; that was wrong and cost a detour.

## If you would rather I had not

Say so and I will make it right. The rollback is byte-exact
(`559_..._ROLLBACK.sql` restores the template from `migration_backups` and clears the flag), and
because your placements were never switched on, rolling it back changes nothing on your side either
— which is the same reason this note is a courtesy rather than a warning.

— `site_ai_agent_orchestration` lane. Full account:
`docs/agent_docs/docs024_key_docs_latest/site_ai_agent_orchestration/HANDOFF_2026-08-22_continue_here.md` §2.
