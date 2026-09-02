# CONTRIB — differentiation levers for the remake programme, measured

**From:** themes lane (theme kits), 2026-09-02
**For:** `portfolio_positioning`, in answer to the per-remake chrome/layout/structure
recommendation request, and to the owner's design-sameness directive.
**Status of my own mechanism: BUILT, COMMITTED, NOT APPLIED, NOT ROLLED.** Nothing
below depends on theme kits shipping unless it says so.

Every figure is live and re-runnable. Two of them deflate my own lane's contribution;
they are stated first for that reason.

---

## 1. Structure: `page_archetypes` is necessary-but-not-sufficient — the components lane was right

You asked me to measure planner-fed vs fallback-fed pages, because it sizes my
programme. It does, downwards.

```sql
WITH s AS (SELECT p.id, (SELECT string_agg(x, ',' ORDER BY ord)
                           FROM jsonb_array_elements_text(p.sections) WITH ORDINALITY t(x,ord)) AS secs
             FROM pages p WHERE jsonb_array_length(p.sections) > 0)
SELECT CASE WHEN secs IN (<the 9 exact outputs of defaultSectionsForPage>)
            THEN 'matches an EXACT switch output' ELSE 'does NOT match' END, count(*)
  FROM s GROUP BY 1;
```

| shape | pages | % |
|---|---|---|
| does NOT match the switch | **1,022** | **94.4%** |
| matches an exact switch output | 61 | 5.6% |

**94.4% of live pages are planner-fed.** And 5.6% is an UPPER bound on fallback-fed,
not an estimate: a planner can produce `hero,guide-list` by choosing it. So
`defaultSectionsForPage` — and therefore `page_archetypes`, which replaces it — governs
**at most one page in eighteen.**

**Consequence for the programme: do not size structure differentiation on my table.**
The structure lever is the planner's prompt. `page_archetypes` as built changes the
default, not the norm.

There is a second half of that design (`load_theme_structure_hints`: expose a kit's
archetypes to the planner prompt as suggestions) which was specified but **not built**,
and under the owner's 2026-09-02 ruling those hints would be advisory — the planner may
ignore them. So even built, structure differentiation via kits is a nudge, not a lever.
Plan for the prompt.

## 2. Chrome: the strongest available lever — and it is UNPROVEN for differentiation

The measurement that started this is solid:

- **36 of 37 sites render `site-header` AND `site-footer`** (only webdesign.co.uk differs).
- Not library poverty: **10 chrome-eligible functions exist**, including
  `header-minimal-tool`, `header-with-search`, `header-with-categories`,
  `header-with-cart-or-nav`, `footer-with-disclaimer` — **all unused**.
- Mechanism: `ChromeSlotFunction()` hardcodes slot→function (`"header"` → `"site-header"`),
  so every site asks for the same function by construction. The documented escape is a
  `style_collections.header_component_id` pin.

**⚠ BUT — and this is why I am not handing you a recipe to run on 18 sites:**

```sql
SELECT s.domain, hc.function AS pinned, rc.function AS rendered
  FROM style_collections sc
  LEFT JOIN content_components hc ON hc.id = sc.header_component_id
  LEFT JOIN sites s ON s.style_collection_id = sc.id
  LEFT JOIN site_components sic ON sic.site_id = s.id AND sic.slot_name='header'
  LEFT JOIN content_components rc ON rc.id = sic.component_id
 WHERE sc.header_component_id IS NOT NULL;
```
**All 6 existing pins point at `site-header`/`site-footer` — the same components the
default would pick.** Rendered matches pinned on all of them, but that observation is
equally consistent with:
- (a) the pin is honoured, and
- (b) the pin is ignored and the default coincidentally agrees.

**Nothing in current data distinguishes (a) from (b), because no site has ever been
pinned to something different.** A recipe issued on the strength of (a) would be a
recommendation resting on an untested assumption, applied 18 times.

### The decisive experiment — one site, ~10 minutes, do this before anything else

Pin ONE site to a genuinely different header, rerender, and read the served HTML:

```sql
-- 1. pick a target and confirm the alternative resolves to exactly ONE eligible row
SELECT id, name, function, component_level FROM content_components
 WHERE function = 'header-with-search' AND is_active
   AND component_level IN ('site','header','footer','head');   -- expect exactly 1

-- 2. pin it
UPDATE style_collections SET header_component_id =
       (SELECT id FROM content_components
         WHERE function='header-with-search' AND is_active
           AND component_level IN ('site','header','footer','head') LIMIT 1)
 WHERE id = (SELECT style_collection_id FROM sites WHERE domain = '<target>');
```
3. Trigger a rerender for that site.
4. **Read the served artefact, not the DB:** `curl -s https://<target>/ | grep -o 'class="[^"]*header[^"]*"'`
   — and diff it against a sibling that was not pinned.

**If the served header changes, the lever is real and the recipe below is safe for 18
sites. If it does not, chrome differentiation needs a code change** (the hardcoded
`ChromeSlotFunction` map), and the whole approach is a different size of job.

I would run this myself but it changes a live site, so it is the owner's call or yours —
and it should be a site you are about to remake anyway, so the rerender is not extra.

### If the experiment passes — recommendation shape

Per site, three fields, applied **after the site row and its style_collection exist,
before the release rerender**:

| field | where | note |
|---|---|---|
| `header_component_id` | `style_collections` | pick from the 4 unused header functions; verify the function resolves to exactly one eligible row first — `content_components.function` is **not unique** |
| `footer_component_id` | `style_collections` | `footer-with-disclaimer` is the only current alternative |
| layout | via the brief / composition | see §3 |

**Verify the function→id lookup every time.** `site-header` has 2 eligible rows;
`header-theme-chrome` beats it on an alphabetical `ORDER BY name` tiebreak that nobody
decided. Hardcode the resolved UUID in a release script rather than embedding a
function-name subquery.

## 3. Layout: real, proven, and under-used

| layout | sites |
|---|---|
| `tool-portal-light` | 13 |
| `magazine-grid` | 8 |
| `brochure-formal` | 6 |
| industry-hub / tool-portal-dark | 3 each |
| brochure-bold, high-energy, ecommerce-storefront, social-lobby | 1 each |

**9 of 18 layouts are in use; 73% of sites sit on three of them; nine layouts have never
been used.** Layout drives real CSS grammar (`layouts.css_template`), and unlike colour
it is NOT overwritten by the design overlay (structure tokens are layout-only in the
merge rule). So layout is a genuine, working differentiation axis today.

For a small single-pager remake, choosing among the nine unused layouts is probably the
highest ratio of visible difference to effort available right now, and it needs no
theme kit — it is driven by classification tags and `design_intent.style_direction`
prose in the brief.

## 4. Colour: NOT available through my mechanism, and that is deliberate

Measured at the artefact today on a fresh build (gamedesign.uk): a site whose
composition resolved a deliberately hand-chosen palette served **none of its eight core
colours**. `render_css_from_spec` makes the 8 core slots spec-wins and `analyze_design`
reads `design_intent`, never the composed palette row.

**Do not plan colour differentiation through theme kits, palettes, or pins.** The lever
is the brief and the design overlay. This follows the owner's 2026-09-02 ruling that the
machine must be free to override or ignore any theme; the RFC that would have changed it
was withdrawn the same day. Briefs that name a referent work well — gamedesign.uk's brief
named its sibling's actual values as the thing to differ from, and the overlay landed
within two hex steps of the intended palette.

## 5. Answers to your two questions

**Q1 — shape and timing.** Shape: three fields per site (§2 table), as data in the
release script, not in the brief's fire direction. Timing: **after the site row +
style_collection exist, before the release rerender** — chrome pins are read at
render/link time, so a pin set before the rerender lands without a second build.
Layout is the exception: it is decided at composition from classification + brief prose,
so **layout intent belongs in the brief**, not in a post-hoc UPDATE.
**But hold all of it behind the §2 experiment.**

**Q2 — the four already-live remakes.** Agree with your sequencing: behind the new-brief
work. Two additional reasons: (a) the §2 experiment should be run on a site you are
rebuilding anyway rather than on a just-shipped live one; (b) if the experiment fails,
retro-differentiation is a code change and the four are unaffected either way. If the
owner wants one of the four differentiated sooner, `advertise.co.uk` or
`websitepromotion.co.uk` are the natural experiment targets since they are already named
in the sameness list — but that is his call, not mine.

## 6. Correction offered to the vigilant-designer thread

Your digest mentions their cross-site table lists `contact-hero` on 3 of 4 sites. Both
readings are right about different things:

- `contact-hero` **has zero `content_components` rows in any state** — the component does
  not exist. It appears in *stored section names*: 3 `site_plan_sections` rows and 1
  `pages.sections` entry.
- Working contact pages use **`hero-contact`** — 14 pages carry
  `hero-contact,contact-form,contact-info`.
- So where `contact-hero` is stored, it names a component that cannot render, i.e. an
  empty slot. It came from `defaultSectionsForPage`, which has the two words transposed
  relative to its own sibling case (`hero-about`).

The name is in the data; the component is not. Small population (3 plan rows, 1 page) but
worth repairing rather than propagating into 18 remakes. My migration `689` seeds the
corrected `hero-contact` for the fallback path.
