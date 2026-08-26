# 407 — a site cannot put its OWN most important page in its own header: membership is decided by a hardcoded fleet-wide page-NAME list, `nav_order` only sorts within it, and the cap is fleet config

**Filed 2026-08-26** by the `finetuning_uk_service` lane, **at the owner's explicit direction**,
after his instruction to promote one page into a header could not be carried out as given.
**Status: OPEN, unowned. Severity: MEDIUM** — nothing breaks, no error is raised, and the page
simply is not there. It costs a site its own conversion route and it costs an operator's
instruction its meaning.

> **OWNER, 2026-08-26, verbatim:** *"Please submit a bugs_open to fix the miss eg perhaps label
> slots and page names rather than having to search from names it considers important"* — the
> proposed direction is recorded as candidate 1 below because it is his, and because it is right.

## 1. What happened, which is the cleanest statement of the defect

The owner asked for `/your-own-model.html` — finetuning.uk's **£99 offer page, its primary
conversion route** — to be moved into the header, and named four pages he was happy to displace:
About, Case Studies, How We Work, Contact.

Contact was displaced. **The offer page did not take the slot — `Pricing` did.** Displacing a
second page from his list was required before the page he actually wanted could appear.

Nothing warned. The page row said `in_header = true` the whole time.

## 2. The mechanism

`platform/orchestration/actions/populate_nav_tables_action.go`:

- **`navPriorityTier(nameLower, pageType)`** assigns every header candidate a tier from a
  **hardcoded, fleet-wide list of page NAMES**:
  - tier 1 — `index, services, tools, about, contact`
  - tier 2 — `blog, news, case-studies, use-cases, pricing, how-we-work, portfolio, products, solutions, industries, model-directory, adoption-tracker, protocol-tracker`
  - tier 3 — everything else, i.e. **every page whose name the platform has never heard of**.
- `classifyPagesForNav` sorts **tier ascending, THEN `nav_order`** — so `nav_order` cannot move a
  page past a tier boundary. A site's own page at `nav_order = 1` sits behind every tier-2 page at
  `nav_order = 100`.
- **`max_header_items` (default 8) lives in `nav-updater`'s step config**, i.e. **fleet-wide**
  `[MEASURED 2026-08-26]`. Raising it for one site raises it for all 31.

So the only three ways to promote a site's own page today are: **rename the page** to a name on the
list; **edit the fleet-wide Go list**; or **displace enough tier-1/2 pages** that it reaches the cap
by elimination. All three are wrong for the same reason — the site cannot express its own priority.

## 3. THE EVIDENCE THAT SETTLES IT: the existing workaround is inside the broken thing, and it does not work

`navPriorityTier`'s tier-2 map contains this, with a comment explaining itself:

> `"model-directory": true, "adoption-tracker": true, "protocol-tracker": true,`
> *"The directory registers … Ranked explicitly because the alternative is not neutral: as tier 3
> they sort behind every tier-2 page and are the first thing dropped when max_header_items
> truncates, so a site that deliberately promoted its directory into the header would silently lose
> it again the next time any other page gained in_header. That is a real sequence —
> ai-agent-orchestration.com's header was exactly full on 2026-07-25 and the directory only fitted
> because the owner moved Pricing down to make room."*

**Three site-specific page names were hardcoded into a fleet-wide list to fix this once. `[MEASURED
2026-08-26]` `adoption-tracker` and `protocol-tracker` are STILL ABSENT from that site's header** —
along with `news-index` — because ai-agent-orchestration.com now sits at the 8-slot cap again.

So the documented remedy was applied, is in the source today, and the pages it was written for are
still missing. That is the argument for a structural fix rather than a longer list.

## 4. Damage `[MEASURED 2026-08-26]`

- **8 pages across 5 sites** declare `in_header = true`, are active and deployed, are not child
  pages, and **do not appear in their site's primary nav**.
- **5 of 31 sites sit at the 8-slot cap**, where any newly promoted page is silently excluded.
- Confirmed instances of *this* mechanism: `ai-agent-orchestration.com` — `news-index`,
  `protocol-tracker`, `adoption-tracker` (header full, all tier ≥2 losers);
  `gaswholesalers.com` — `pricing-transparency`; `finetuning.uk` — `approach`.

> ⚠ **A SECOND CAUSE IS MIXED INTO THAT 8 AND I COULD NOT CLEANLY SEPARATE IT.** Some absences are
> simply a **stale nav** — `loanzy.uk`'s nav was last rebuilt **2026-08-18** and its `index` (tier
> **1**) and `glossary` are absent while its header holds only **3 of 8**, which the tier mechanism
> cannot explain. I tried to discriminate with `pages.updated_at > max(site_nav_items.updated_at)`
> and **that discriminator is no good**: `pages.updated_at` is bumped by any re-render, so it does
> not tell you the *flags* changed. **Whoever picks this up needs a real discriminator before
> quoting a split.** The 8 is a true upper bound on this defect, not a measurement of it.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **THE OWNER'S: declare the slots, per site.** A site names its own header — an ordered list of
   page ids/labels — and the nav builder renders that list. The hardcoded tier table degrades to a
   **fallback for sites that have not declared one**, which is what it is actually good at
   (a sensible default for a fresh build). This makes the defect unrepresentable rather than rarer:
   there is no longer a fleet-wide opinion that can outrank a site's own. It also makes the
   operator instruction *"put X in the header"* mean exactly one thing.
   ⚠ Design note for whoever builds it: the declaration must survive a nav REBUILD, which today
   `DELETE`s and re-derives `site_nav_items` from `pages` — so it belongs in a site-scoped spec or
   on the page rows, **not** in `site_nav_items`, which is a derived table.
2. **Per-site `max_header_items`.** Cheap, and strictly worse: it does not let a site say what
   matters, only how many things fit. Useful *with* candidate 1, not instead of it.
3. **Keep extending the fleet-wide name list.** This is today's approach. §3 is the measurement
   that it does not work: the last extension is in the source and its pages are still missing.

## 6. How to verify a fix

```sql
-- must return 0: a page that declares header membership and is not in the header
SELECT s.domain, p.name FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE p.in_header AND p.status='active' AND p.build_status='deployed'
   AND p.url NOT LIKE '/tools/%' AND p.url NOT LIKE '/blog/%' AND p.url NOT LIKE '/guides/%'
   AND NOT EXISTS (SELECT 1 FROM site_nav_items ni JOIN site_nav_groups ng ON ng.id=ni.group_id
                    WHERE ni.site_id=p.site_id AND ng.group_type='primary'
                      AND ni.status='active' AND ni.url = p.url);
```

⚠ **Verify at the SERVED page, not at the nav tables.** A nav rebuild's last step only FILES
re-render items — the tables can be correct for an hour while every served header is stale (52
items on finetuning.uk, 2026-08-25). And `pages.rendered_header` is NULL site-wide on some sites,
so a column check reads "never shipped" for ever (LANDMINES 2026-08-25).

## 7. Sources

`platform/orchestration/actions/populate_nav_tables_action.go` — `navPriorityTier` (the name lists),
`classifyPagesForNav` (tier-then-nav_order sort), `max_header_items` (step config, default 8) ·
`agent_definitions` row `nav-updater`, step `populate_nav_tables` · lane
`docs024_key_docs_latest/finetuning_uk_service/HANDOFF_2026-08-25b_continue_here.md` (the trap as
this lane hit it) · `bugs_open/149` A2 (the adjacent nav-membership family).
