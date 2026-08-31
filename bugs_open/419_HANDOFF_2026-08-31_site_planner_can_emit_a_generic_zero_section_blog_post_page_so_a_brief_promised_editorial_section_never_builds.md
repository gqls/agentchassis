# 419 — the site planner can emit a GENERIC, ZERO-SECTION blog-post page, so a brief-promised editorial section never builds — and the link validator then hides the damage

**Filed 2026-08-31** by the delivery-lane session, from the FIRST PAID CUSTOMER BUILD
(boxingonline.com, order BR-9AUZ59, site `d2aa5206-73bc-4707-a69c-2702c1eb9152`).
**Root cause: UNDIAGNOSED — a 090 needs_diagnosis run is being filed alongside this
file; do not quote this file as a mechanism finding.** Everything below is symptom,
census and impact, measured at the live DB on the date in the heading.

## Symptom, end to end (all legs measured 2026-08-31)

1. The paid brief promises, verbatim: *"There is an editorial section featuring
   article titles as clickable links. Six article slots are included with placeholder
   editorial articles to show how the section works."* (build_queue.direction,
   order_reference BR-9AUZ59).
2. The site planner (needs_site_plan completed 12:37Z, plan
   `bba66eda-2eae-459f-9e37-896efc9d079c`) emitted SIX pages, five with planned
   sections — and ONE page `article` (role blog-post, /blog/article.html) with
   **ZERO rows in site_plan_sections**.
3. page-build-handler then no-ops on it, correctly: orchestration
   `7c9b803e-9324-4894-93c5-6924225cf5d4` → `load_spec_sections` returned
   `{"count": 0, "source": "none"}` → `mark_no_ready_sections` → the needs_page item
   parks at needs_human_review ("Build article page (not_built)"). The page row stays
   `build_status='planned'` for ever (which is also bug 349's live-predicate food).
4. Three `unresolved_cta` items park at needs_human_review (index ×1,
   articles-index ×2): CTAs whose destinations were meant to be articles have no
   real-page destination.
5. **The wrong result then looks exactly like the right one:** validate_page_content's
   link repair rewrote the dead editorial CTAs to point at existing pages
   (fight-calendar, contact), so the SERVED articles-index (checked at the published
   copy, 2026-08-31 ~14:45Z) renders a clean page with **zero article links** and no
   visible defect. Every page validated `valid=true, issues=0`. Nothing about the
   served site says "the brief's core editorial promise is missing" — only the diff
   between build_queue.direction and the page inventory says it.

## Census — this is a repeating shape, not a one-off (counted 2026-08-31)

```sql
SELECT s.domain, p.name, count(sps.id) AS sections
FROM site_plan_pages p
JOIN site_plans sp ON sp.id=p.plan_id
LEFT JOIN sites s ON s.id=sp.site_id
LEFT JOIN site_plan_sections sps ON sps.plan_id=p.plan_id AND sps.page_name=p.name
WHERE p.role='blog-post' GROUP BY 1,2;
```

**2 plans** (as of 2026-08-31) carry a GENERIC zero-section blog-post page:
`adversecreditmortgage.co.uk` (`blog-post`) and `boxingonline.com` (`article`).
Every healthy plan instead emits **named, concrete articles with sections**:
dartsonline.com 9 × 3 sections, agritec.uk 6 × 3, farmerinsurance.uk several × 3.
So the planner CAN do this and usually does; under some condition it emits a
placeholder-shaped generic page instead. (Note the generic page's NAME differs
between the two hits — `blog-post` vs `article` — so a name-based detector is the
wrong shape; zero-sections × role=blog-post is the signature.)

## Why this matters more now than it did

Estate builds had no customer to disappoint. This fired on the FIRST paid,
customer-attested build: the one deliverable feature the customer described in most
detail is the one the plan silently dropped, and every downstream honesty mechanism
(validator, link repair) actively tidied the absence into a clean-looking page. The
shape recurs for every future editorial/content-hub order.

## Open questions for the diagnosis run (do NOT treat as findings)

- What distinguishes the two zero-section plans' inputs from the healthy ones?
  [UNVERIFIED candidates: objective wording that *describes* placeholder articles;
  planner prompt treating "article slots" as a template concept; a truncation or
  section-emission failure specific to blog-post roles.]
- Is `adversecreditmortgage.co.uk`'s hit the same mechanism? (That site was never
  built at all — bug 349 classed it "legitimate whole-site-unbuilt"; its plan shape
  may predate other planner changes.)

## Remediation applied for boxingonline (this lane, 2026-08-31)

Six sequential `content-gap-planner` dispatches (approach B: one new blog-post page
each, framework-chosen topics, placeholder-quality per the brief), then rerender of
articles-index/index, then the stranded `article` page + parked items surfaced at
the owner's pre-delivery review. See
`docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md`
2026-08-31 entries for measurements and outcomes.

## How to verify a fix

Build (or re-plan) a site whose objective promises an editorial section with N
placeholder articles; assert the plan contains N role='blog-post' pages EACH with
≥1 site_plan_sections row, and zero role='blog-post' pages with none. The census
query above is the detector; run it against the new plan_id.
