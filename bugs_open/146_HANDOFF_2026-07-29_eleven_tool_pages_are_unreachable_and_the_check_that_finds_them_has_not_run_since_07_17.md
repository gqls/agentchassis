# 146 — 11 deployed tool pages are unreachable, and the check that finds them has not run automatically since 2026-07-17

**Filed 2026-07-29 (oufe workstream, found while looking for where to link a second
tool). One instance fixed in-lane; the fleet residual and the two mechanisms below
are unowned.**

## The symptom

`oufe.com/tools/tool-recovery-waterfall.html` was live, Tier-4 verified 13/13, and
**no page on the site linked to it** — not the nav, not the chrome, not any page
body. A visitor could reach it only by knowing the URL. It had been like that since
it was deployed on 2026-07-28.

Fleet census, run 2026-07-29 with the platform's own orphan predicate (nav items by
url, nav items by page_id, `site_components.rendered_html`, other pages'
`page_components.rendered_html`):

| domain | unreachable tool pages |
|---|---|
| gaswholesalers.com | 4 |
| gamesdesign.co.uk | 4 |
| finetuning.uk | 1 |
| leopardessconsulting.co.uk | 1 |
| oufe.com | 1 (**fixed 2026-07-29**, mig 268) |
| **total** | **11 of 94 deployed tool pages, across 5 of 11 sites** |

Denominator checked: 94 deployed tool pages with a url, across 11 sites. The
all-page-types figure on the same predicate is **41 unreachable deployed pages**.

**Every one of the 11 carries a nav flag** (`in_header` or `in_footer` true) — so
none is excluded by `check_orphan_pages`' tool-page carve-out, and each is
`nav_drift` by the platform's own routing.

## Mechanism 1 — the check exists, is correct, and does not run

`check_orphan_pages` (`discovery_checks/check_orphan_pages.go`) implements exactly
this predicate and routes nav-flagged orphans to `nav_drift` → `nav-updater`. It is
registered in **one** agent:

```sql
-- run_discovery_checks steps, live definitions, 2026-07-29
design-discovery-agent        → 22 checks, orphan_pages NOT among them
completeness-discovery-agent  → 29 checks, INCLUDING orphan_pages
quality-discovery-agent       → 5 checks, orphan_pages NOT among them
```

`completeness-discovery-agent` has raised work items on **7 sites, most recently
2026-07-25** (vonc), and **has never run on oufe.com** (site created 07-25).
Its `needs_internal_links` items: 24 by the agent, all on/before 2026-07-17; the
9 since are `created_by='generic'` or a named thread — i.e. **sessions firing the
checks by hand**. Design-discovery *has* run on oufe (raised
`needs_brand_head_assets` 07-28), so "discovery runs on this site" is true and
misleading: the discovery agent that would catch this one does not.

This is the `zero-adoption-means-read-the-mechanism` shape: the capability is built
and correct, and the cadence is what is missing. **Do not write a check; schedule
the agent.**

## Mechanism 2 — the obvious remedy destroys site content, silently

The natural fix for `nav_drift` is `nav-updater` / `refresh_site_components`, which
re-renders chrome from `footer-theme-chrome`. On oufe that regeneration **would have
worked** for the link — `buildServicesHTML`
(`render_site_components_action.go:950`) already selects pages with
`in_header OR in_footer`, so the tool would appear in the footer "Explore" list
without any help.

**It would also have deleted the site's standing fallibility disclosure.** The
footer note —

> "OUFE publishes educational analysis of financial and legal mechanism. We make
> mistakes, and some of what is here is assembled with AI assistance that can invent
> convincing detail. Check anything that matters against the primary source."

— is **not in the chrome template and is produced by no Go code**. Verified
2026-07-29 by grepping the live `footer-theme-chrome` template for `footer-note`
(absent) and for the note's own words (absent). It exists only in this site's stored
`site_components.rendered_html`, hand-written by the workstream. A regeneration
overwrites the artefact and reports success.

So the two defects interlock: **the page that cannot be reached is fixed by the one
action that removes the disclosure explaining the site can be wrong.** Any site with
hand-patched chrome has the same exposure; oufe is simply the one where it was
looked at. `bugs_open/117`/`118` are the neighbouring cases (chrome is a stored
artefact; nothing rebuilds it) — this is the sharper version, because here the
rebuild is *harmful*, not merely absent.

## Mechanism 3 — a site-wide reassemble silently skips owned pages

`rerender-pages` on oufe fanned out 8 `page_rerender` items; **5 completed and 3 sat
at `triaged`, never claimed** — exactly the three `rebuild_policy='owned'` pages
(disclaimer, privacy, the tool). No error, no failed item, and the orchestration
reported `COMPLETED`. So a chrome change on a site with owned pages reaches only
part of the site, and the run that did it looks clean. The three had to be deployed
individually with the assemble-only path (`049b_deploy_single_page.sh`).

Practical consequence for anyone fixing the fleet residual above: **after any
chrome change, count the deployed pages, not the orchestration status.**

## What was done here

`docs/agent_docs/sql_for_agents/268_oufe_footer_links_the_tool.sql` — a targeted
`replace()` on the stored footer adding the tool to the Explore list, guarded
(no-op on replay), with a VERIFY block that counts occurrences and asserts the note
survives. Then `rerender-pages` **without** `refresh_site_components` (stored
chrome), plus three assemble-only deploys. Verified: 8 of 8 served pages carry the
link and the note; oufe no longer appears in the orphan census.

## Fix candidates, ordered by what closes the door

1. **Schedule `completeness-discovery-agent`** on the same cadence as
   design-discovery. Closes the detection gap for all 29 of its checks, not just
   this one. Needs someone to check why it stopped — the last automatic items are
   2026-07-17, which suggests a schedule that lapsed rather than one never written.
2. **Make the footer note survive a regeneration** — either promote it into the
   shared chrome template as an optional `{{if .footer_note}}` block fed from site
   data, or hold it in a component the regeneration does not own. This is a
   shared-template change (every site renders `footer-theme-chrome`), so it is
   architecture-scope and wants a council round on its own merits.
3. **Make the owned-page skip visible**: a `page_rerender` item that cannot be
   handled should fail loudly or be marked `wont_fix` with a reason, not sit at
   `triaged` forever inside a run that reports COMPLETED.
4. Route the 10 remaining fleet orphans — but note they belong to other lanes
   (gaswholesalers, gamesdesign, finetuning, leopardess), and each needs candidate 2
   settled first if its chrome has been hand-edited. **Check that before
   regenerating anyone's chrome.**

## How to verify

The census query is in this file's opening section; re-run it and expect the
per-site counts to fall. For candidate 2, the test is: regenerate a scratch site's
chrome and confirm a site-specific footer note survives.

## Relations

`bugs_open/117`/`118` (chrome is a stored artefact; selection ignores `is_active`) ·
`bugs_open/098` (deployed ≠ fetchable; this is fetchable ≠ reachable) ·
`bugs_open/140` (filed the same morning — a component asserting what the site never
said; same family of "the artefact and the data disagree") · oufe NOTES 2026-07-29.
