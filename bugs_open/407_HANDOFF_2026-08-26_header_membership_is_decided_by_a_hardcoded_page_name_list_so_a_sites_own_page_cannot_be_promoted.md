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

---

## 2026-08-26 (evening) — TAKEN, FIXED, COUNCIL-APPROVED r1, COMMITTED. **Stays OPEN: inert until 654 applies after the roll — and ONE OWNER DECISION is owed (§B).**

Taken by the `bugfix_407_site_declares_its_own_header` lane. `who-owns.py 407` returned
**`likely OWNING workstream(s): (none identified)`** before starting. Commit `74e92e961`;
council `cb67cc71-b420-4399-9a52-f306d0f4bccf`, **APPROVED round 1**, 6 advisory objections,
none high.

### A. §4's damage count corrected — the discriminator is BUILT, and it is not a timestamp

§4 warned that a second cause was mixed into the 8 and that its own discriminator
(`pages.updated_at > max(site_nav_items.updated_at)`) was no good. It is not, and the way past
it is not a better timestamp: **`classifyPagesForNav` is deterministic Go over `pages`, so it
can be REPLAYED in SQL** and its expected primary rank diffed against the stored nav. A page
absent from stored primary is then exactly one of three things:

| expected | verdict |
|---|---|
| rank **>** cap | **excluded by the tier/cap — this bug** |
| rank **≤** cap | a merely **STALE nav** (a rebuild would place it) |
| not a candidate at all | **another cause** — barred by type or URL |

`[MEASURED 2026-08-26]` the §6 population is 6 today, and the replay splits it **5 / 1**:

```
ai-agent-orchestration.com  adoption-tracker      tier 2  rank 10  EXCLUDED BY TIER/CAP
ai-agent-orchestration.com  news-index            tier 2  rank 11  EXCLUDED BY TIER/CAP
ai-agent-orchestration.com  protocol-tracker      tier 2  rank  9  EXCLUDED BY TIER/CAP
finetuning.uk               approach              tier 3  rank  9  EXCLUDED BY TIER/CAP
gaswholesalers.com          pricing-transparency  tier 3  rank 12  EXCLUDED BY TIER/CAP
idea.uk                     report                page_type=tool   barred by neverPrimaryTypes
```

**So the honest population of THIS bug is 5 pages across 3 sites, not the filed 8 across 5.**
`loanzy.uk`'s two — §4's own suspected stale-nav case — have gone from the population entirely
since filing. `idea.uk/report` is not stale either: it is `page_type='tool'`, and **no rebuild
would ever place it**. A cruder cap-only screen agrees on the 5/1 split but cannot say why; both
queries are in the lane RUNBOOK, and the cap one must read `max_header_items` from live config
rather than hardcode 8, or it asserts the very fleet constant this bug is about.

### B. ⚠ ONE OWNER DECISION IS OWED, and it is deliberately not assumed here

The council's `guardian` seat objected [medium] that the fix lets a site's declaration override
**three independent membership guards at once** — `pages.in_header`, `neverPrimaryTypes`
(blog-post / tool / entity-page) and the child-URL bar — and that this is a semantic widening
beyond "fix the tier order" which deserves explicit sign-off rather than being cited by analogy
to the 2026-08-02 §2 ruling. **The seat is right that it is a widening, and it is right that it
should be your call.**

The case for it, so the decision can be made on the merits: every one of those three is a
**fleet-wide DEFAULT**, and the whole point is that a default may not outrank a site's own word.
`idea.uk/report.html` is the worked example — the site set `in_header`, set `nav_order` 3, and
gets nothing, because the platform has decided that pages of type `tool` are never in a header.
The undeclared path keeps all three bars unchanged, and the two guards that are **correctness**
rather than preference — the system-page exclusion and the legal group — are NOT overridable: a
site declaring `privacy` or `404` is told so rather than obeyed.

What it costs if you say no: the fix still solves finetuning.uk and gaswholesalers (tier and
same-tier ties), and `idea.uk/report.html` stays unpromotable.

### C. What shipped

`site_specs.data -> 'chrome' -> 'header_slots'` — an **ordered array of `pages.name`** — plus
optional `chrome.max_header_items`, read by `populate_nav_tables`. Declared pages take the
leading primary slots in the site's order ahead of every undeclared candidate; undeclared
candidates fill what remains by the existing tier/`nav_order` sort; overflow still goes to
utility so nothing vanishes. Registered **NAV-014**.

**It is an ORDERED list because half the damage is same-tier ties** — gaswholesalers' absent page
and the three that beat it are all tier 3 at `nav_order` 100, so a design that only RAISED a
page's tier would have fixed finetuning.uk and left that site untouched. §5 candidate 2
(per-site `max_header_items`) is IN, as the same key.

**Where it lives changed after the council.** The first draft used `sites.settings->'nav'`, on a
rationale asserting *"there is NO `site_specs` table"* — which I had read off a `\dt site*`
listing truncated at twenty rows. `site_specs` sorts on line 21. The `reuse_agent` and
`prior_art_librarian` seats both caught it, and the measurement settles it: **`site_config.chrome`
already carries per-site HEADER config** (`header_cta_url`/`header_cta_label` on oufe.com,
`compliance_lines` on two more). Header slots belong beside header CTAs. The declaration now
inherits versioning, `pinned`, provenance and a real writer, none of which the settings column
had. Recorded in `WRONG_CALLS.md`.

**§5 candidate 3 (extend the fleet list) is refused on this file's own §3 evidence**, now proven
at the served page: of the three names hardcoded into tier 2 to fix this once, **`model-directory`
is in ai-agent-orchestration.com's header and its two siblings are not.**

### D. Opt-in, and the no-op is a query rather than an argument

`[MEASURED 2026-08-26]` **0 of 51 sites declare `header_slots`** (3 have a `chrome` object at
all), so the first roll changes nothing for the fleet. And `classifyPagesForNav` keeps its exact
signature as a wrapper, so `nav_membership_test.go` — which pins `bugs_open/149` A2's invariant —
passes with a **ZERO DIFF**, all 8 tests green. That zero diff is the evidence, not the claim.

Eight guards mutation-proven, each by a named failing test, and **re-run in full after the table
moved** — a mutation proof is about the code that shipped, not the code that was planned.

### E. What remains

1. A chassis roll carrying `74e92e961`.
2. Probe the RUNNING binary for `header_slots` with a positive AND a negative control in the
   same exec.
3. **Take the pre-fix replay reading BEFORE applying 654**, or the after has nothing to be
   compared with. That is what 654's hold exists for — unlike a checks-array name, an early
   apply here is inert rather than fatal.
4. Apply 654 (seeds ai-agent-orchestration.com only — finetuning.uk's ordering is the owner's
   decision, §B), trigger `nav-updater`, and read `nav_declaration_source` / `declared_missing`
   from the step's own result before asking whether it worked.
5. **Verify at the SERVED page** after the re-render items drain, per §6's own ⚠ — and anchor on
   `<nav>` or the chrome's class, **never on a bare `<header>` tag**: a rendered page carries
   several, and matching the wrong one gave a 420-byte "header" with no links during this
   investigation.

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_407_site_declares_its_own_header/`.

---

## PROVEN IN PRODUCTION 2026-08-31 — the declaration mechanism's FIRST use, on the first paid customer site

**The fix this file asked for (`nav_declaration.go`, `site_config` →
`chrome.header_slots`) was exercised for the first time tonight and worked end to
end**: declaration → nav tables → header re-render → rolling redeploy of all 19
pages → mirror → **served order matches the declaration exactly**
(Home · News · Fight Calendar · About · Contact). Verified independently by two
sessions (delivery lane + boxingonline critique session), 19/19 pages, both
sweeps agreeing; measurements in
`docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md`
2026-08-31 ~18:1x–19:0x entries.

Two transferable findings from the exercise, which matter more than the outcome:

1. **What the declaration beat was NOT a broken row — it was the TYPE BAR.**
   boxingonline.com's fight-calendar row was fully populated and correctly
   ordered the whole time (`in_header=true`, `nav_label='Fight Calendar'`,
   `nav_order=3`), and the site sat far under the 8-slot cap — so neither of
   this file's §2 mechanisms applied. `page_type='tool'` sits in
   `classifyPagesForNav`'s `neverPrimaryTypes`, so a perfect row was classified
   into the utility group and rendered in the footer. **Anyone debugging a
   missing nav item will check the row first, find it perfect, and stall** — the
   declaration is the sanctioned override for exactly that state, and this is
   its worked case. (The page was the paid brief's NAMED core deliverable, which
   is what made the footer placement a delivery blocker rather than a quirk.)

2. **The verification shape that made the close safe**, each element of which
   caught something real the same evening that a looser version would have
   missed: enumerate the page set from `pages WHERE deployed_at IS NOT NULL`
   (never from memory); scope the assertion INSIDE the `<header>` block (an
   occurrence-count of 3 would also have passed if the header were the one
   place the link was missing — the footer and mobile menu carry it too); a
   must-be-present control per page so a zero is an absence rather than a
   broken fetch; and two independent sweeps required to agree before closing.
