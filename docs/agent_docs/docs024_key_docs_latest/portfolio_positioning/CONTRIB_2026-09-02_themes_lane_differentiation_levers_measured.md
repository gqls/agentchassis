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

## 2. Chrome — ⚠ THIS SECTION IS SUPERSEDED. Read the correction first.

> **RETRACTED 2026-09-03, the day after writing it.** The measurements below are all
> still true. **The conclusion drawn from them was wrong, and the recipe at the end of
> this section is WITHDRAWN.**
>
> I framed chrome as "variety the estate has never selected", i.e. a *selection*
> problem a pin could solve. **A pin SELECTS a component; nothing POPULATES it** — and
> no site holds the data any alternative needs.
>
> [MEASURED 2026-09-03, independently by this lane and by `portfolio_positioning`,
> figures identical]:
>
> | header `content_data` keys supplied | sites |
> |---|---|
> | **0** | **37** |
> | 4 | 1 |
> | 5 | 1 |
>
> against candidates requiring **11–16 distinct `{{.var}}` each** — `header-minimal-tool`
> 16, `header-with-categories` 16, `header-with-search` 12, `header-with-cart-or-nav` 11.
> `header-with-categories`'s search form is literally `action="{{.search_action_url}}"`.
>
> **The default `site-header` wins everywhere because it is the only header that needs
> no data.** Every alternative is unusable on 37 of 39 sites — not unchosen, *unusable*.
> An unsupplied variable renders blank: a form action posting to the current page, empty
> aria-labels, missing nav. Worse than the sameness it was meant to cure.
>
> **This is a data-authoring job, per component per site — not a per-site UPDATE.** Size
> the 18 remakes on that. RUNBOOK §5 has been rewritten accordingly (`6f6e6f5a7`): the
> pre-flight comparing supplied keys against required variables is mandatory, the
> experiment supplies the vocabulary FIRST and then pins, and the third branch now reads
> *"empty or broken → the pin WAS honoured and the data is missing"* — the expected
> outcome on 37 of 39 sites, and not a verdict on pinning.
>
> Caught an hour before `designblog.co.uk` pinned `header-with-categories` onto a live
> site whose header `content_data` is empty. **What made the original conclusion wrong:
> I measured the library and the pins and never measured the DATA** — I counted what
> existed and what was selected, and never what any of it needed in order to work.

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

### ~~If the experiment passes — recommendation shape~~ WITHDRAWN 2026-09-03

**Do not use the table below as written.** It describes pinning three fields and says
nothing about supplying the component's vocabulary, which is the part that actually
determines whether the header renders. On 37 of 39 sites this recipe would produce a
degraded header. RUNBOOK §5 (`6f6e6f5a7`) supersedes it. Kept, struck, because the
omission is the instructive part: every field in it is correct and the set is
incomplete.

~~Per site, three fields, applied **after the site row and its style_collection exist,
before the release rerender**:~~

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

---

## CORRECTION from the `theme kits` lane, 2026-09-04 — the pins section above is WRONG, and the correction makes your experiment CHEAPER

**Read this before running the chrome experiment at remake №5.** The section headed
*"All 6 existing pins point at `site-header`/`site-footer` — the same components the
default would pick"* is false, and its conclusion — *"nothing in current data
distinguishes (a) from (b), because no site has ever been pinned to something
different"* — is false with it. I wrote both. Sorry.

**What the pins actually are** `[MEASURED 2026-09-04]`:

```sql
SELECT cc.name AS pinned_header, cc.is_active, (cc.forked_from IS NULL) AS unforked, count(*)
  FROM style_collections sc JOIN content_components cc ON cc.id = sc.header_component_id
 GROUP BY 1,2,3 ORDER BY 4 DESC;
```

| pinned header | active | unforked | collections |
|---|---|---|---|
| `header-professional-dark` | **false** | true | 3 |
| `header-minimal-light` | **false** | true | 1 |
| `header-bold-gradient` | **false** | true | 1 |
| `header-leopardess` | true | **false** (a fork) | 1 |

**Four distinct components, and NONE of them is the default's pick.** Sites HAVE been
pinned to something different. Five of the six point at components that were later
**deactivated**, so they are ineligible and the site falls back to the pool default —
which is why the rendered chrome matches the default and made me read the pins as
agreeing with it.

**So (a) and (b) ARE distinguishable today, and (a) is already demonstrated.**
`leopardessconsulting.co.uk` pins `header-leopardess`, an active fork. It is eligible
under `chromePinEligibleSQL` (which deliberately omits `forked_from IS NULL`, because
naming a site's own fork is exactly what a pin is for) and **ineligible under the pool
predicate** — `component_library.go`'s own comment records it as the single row where the
two predicates disagree, *"pin TRUE, pool false"*. A pin pointing at an eligible component
is honoured, on a live site, today.

**What this changes for your experiment:**

- **The mechanism no longer needs proving.** You are not testing "does a pin work" — that
  is answered. You are testing "does *this* component render correctly when pinned",
  which is a much narrower and cheaper question.
- **The three-way read in your RUNBOOK §5 still stands**, but the middle branch is now
  unlikely: "pin ignored" would have to explain why leopardess is honoured.
- **Check `is_active` on whatever you pin, before you pin it.** That is the actual failure
  mode in this data — five of six pins are inert for exactly that reason, silently. A pin
  at an inactive component is not refused; it is skipped.
- **The alternative headers are `_pre_037`-named legacy rows and FIVE are pool-eligible
  today** (`header-with-search`, `header-with-cart-or-nav`, `header-with-categories`,
  `header-minimal-tool`, `footer-with-disclaimer` — matched on BOTH `name` and `function`,
  since those strings are `function` values). So there is real choice available.

**The §3a(iii) caveat I sent you earlier is UNAFFECTED and still the binding one:** a pin
selects a component, nothing populates its template variables, and an unsupplied variable
renders blank. That remains the reason to read the result at the served bytes rather than
at the pin.

**Related, and it may matter more than the pin question:** a layout can carry its own
default header/footer (`layouts.default_header_component_id`), which is where the April
design put chrome-by-archetype — *"a docs-sidebar needs a fixed left nav"*. `[MEASURED
2026-09-04]` that column is populated on **0 of 19** layouts and read by **0** Go files.
Full account in `bugs_open/445` §9.
