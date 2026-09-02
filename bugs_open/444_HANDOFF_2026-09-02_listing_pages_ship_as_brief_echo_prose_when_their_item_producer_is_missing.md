# 444 — Remake listing pages ship EMPTY of their content type, filled with brief-echo prose

**Filed 2026-09-02 by portfolio_positioning**, from the owner's designblog.co.uk critique
("explaining the brief and not answering it") — verified same evening on advertise.co.uk, so it
is a CLASS across the day's four remakes, not a designblog instance. First-hand verification is
substituted for a 090 run per the 2026-07-31 owner ruling: every claim below is a direct
measurement of the served body, the live DB, or the concept register, made tonight; the one
absence claim is marked as such.

## Symptom (measured at served bodies, 2026-09-02 ~20:xx)

Listing-type pages serve 200 at full page weight but contain ZERO items of their own content
type, replaced by meta-prose about what the page WILL contain:

- advertise.co.uk `/channels-directory/index.html` — **0 entries** (plan said 15–20); one h2
  "Find UK advertising channels by type", then footer.
- advertise.co.uk `/glossary.html` — **0 terms**; h2/h3s are "What this glossary covers",
  "Where the terms come from", "Reading a definition".
- advertise.co.uk `/news/index.html` — **0 items**.
- designblog.co.uk: `/glossary.html` 0 terms · `/inspiration/` 0 showcases ·
  `/the-design-feed/` 0 items · `/uk-studios-directory/` 0 studios (that lane's own
  verification, `designblog_couk/CRITIQUE_2026-09-02_owner_site_review.md`).

The wrong result looks RIGHT: 200, ~60KB, plausible headings — every naive check passes.

## Root cause — three distinct mechanisms, one symptom (per listing type)

1. **Feed pages: the mechanism exists and is LIVE but per-site UNDRIVEN.** NEWS-001 fills feed
   pages from `content_sources` rows; idea.uk has 5 and a working feed. All four remake sites
   have **0 rows** (measured). Nothing in the build path creates a source row, and no sweep
   ever will — this never converges on its own.
2. **Directory pages: the mechanism exists and is fully wired but the KINDS do not.** DIR-001:
   one kind = one fleet-wide list; six kinds live (model/company/protocol/mortgage-lender/
   savings-provider/health-insurer — counts 51/36/33/28/12/5 tonight). "advertising-channel"
   and "uk-design-studio" are not kinds. Adding one is data-only across SEVEN places (two fail
   SILENTLY — see LANDMINES + directory-pipeline.md) plus a researcher run to populate. No
   sweep adds kinds.
3. **Glossary / inspiration-showcase pages: NO item producer found at all** [ABSENCE CLAIM —
   basis: no glossary/terms/showcase table in information_schema, no concept-register entry,
   grep of register + platform actions; a fixing thread should re-verify]. These were planned
   as `content` pages and the generic writer wrote prose.

**The copy half [INFERRED, writer prompt unread]:** the brief-echo prose is DOWNSTREAM of the
missing items — a section writer given a listing section with no items writes about the
intent. Pages with items (idea.uk news) do not show the pattern. The copy lane can confirm at
the prompt; fixing copy without items would produce nicer prose about an empty page.

## Related, same build-path family (measured tonight)

No remake got a TOOLS nav link or a `/tools/index.html` hub: advertise's nav is
index/guides/news/channels-directory/glossary with `nav_order` 4 conspicuously absent, while
seven tool pages serve. Tool pages arrive post-plan via tool-deployer; the nav rebuild ran
before they existed; the plan never planned a tools hub. (311 is CLOSED — the planner CAN see
library tools on sites that have them; the residue is ORDER: plan before tools exist.)

## Fix candidates, ordered by what closes the door

1. **Plan validation refuses (or explicitly degrades) a listing page whose item source resolves
   to zero** — kind missing, no content_sources row, no producer → a `capability_gap`/HITL item
   naming the missing producer, never a built page of meta-prose. Makes the bad state
   unrepresentable; the existing `capability_gap` carrier fits.
2. **Per-vertical enablement becomes part of the build path**: a brief that plans a directory
   page triggers the kind-creation checklist (7 places); a feed page triggers a
   content_sources seed (owner already flagged WebProNews for advertise — with the feed lane).
3. **A glossary/showcase item producer** — new build, smallest honest scope TBD by whoever
   takes it; until then briefs should keep those page types explicitly conditional.
4. (Weakest) copy-side: teach the writer to refuse listing sections with no items — treats the
   symptom at the last hop.

## How to verify a fix

The advertise pages above are the standing repro: a filled `/news/index.html` after a source
row + feed run; a filled `/channels-directory/` after the kind + researcher run; glossary needs
(3). Judge at the served body, item count > 0 — never at page status or byte size.

## Ownership / routing (as of filing)

designblog instances: the designblog.co.uk session. advertise instances: portfolio_positioning.
Class fixes: (1) build-pipeline/plan-validation owners; (2) kind additions per DIR-001's
runbook + feed lane for sources; nav/tools-hub: site-planner + the 149 nav-membership family.
