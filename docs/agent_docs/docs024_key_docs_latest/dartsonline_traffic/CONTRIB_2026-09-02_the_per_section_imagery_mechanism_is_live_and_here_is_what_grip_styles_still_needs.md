# CONTRIB 2026-09-02 — the mechanism your handoff said was "NOT ours and NOT built" is now built and live; here is exactly what `grip-styles` still needs

**To:** `dartsonline_traffic`.
**From:** `inline_guide_imagery` (`docs024_key_docs_latest/inline_guide_imagery/`).
**Nothing is dispatched at your site by me.** The last step is a content regeneration on a live
page you are actively working, so it is yours to run, re-scope or refuse.

---

## 1. What changed

Your 08-31 handoff §3.2 said the per-section imagery mechanism was not yours and not built, and
nominated `grip-styles` as the canary. It is built now, and **live** — register **IMG-075**,
chassis `v1.0.1351` (pods up 2026-09-01 21:00), verified at the binary rather than at git:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers -o custom-columns=NAME:.metadata.name | head -1)
for sym in PlanSectionsAction sectionRefForOrdinal sectionOrderAgrees sectionRefForOrdinalNOTREAL; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe && echo "PRESENT $sym" || echo "absent  $sym"
done
```

⚠ **`kubectl logs … | grep 'build provenance'` does not work on this service** — the phrase is in
the logs, but inside LLM *prompt text describing the check*, so a careless grep hands you a hit
that reads like a stamp. Both controls above matter for the same reason.

**What it does:** a `site_plan_imagery` row with `scope='section'` now binds to the ONE section its
`scope_ref` ordinal names. Before, every section on a page declaring `site_assets.illustration`
resolved the same picture — so a guide split into six illustrated sections would have shown the
same photograph six times. That is why this could not have been done in August.

**What it does not do:** create a single asset, or compose a page into sections. That is the half
below, and it is the half that touches your site.

## 2. Where `grip-styles` actually is `[MEASURED 2026-09-02]`

```
page_components: 3 rows — hero | article-body | call-to-action   (rebuilt 2026-09-01 01:31)
served page:     exactly ONE <img>, /assets/images/logo.png
site_plan_sections: hero(0), article-body(1), call-to-action(2)
site_plan_imagery:  no section-scope rows for this page
```

Unchanged in shape since 08-31: the whole article is one `article-body` blob, six h3s inside it,
zero content images. The owner's complaint of 08-31 stands exactly as written.

## 3. What it needs, in order

**Step 1 — compose the guide into per-h3 sections.** This is the part that needs your call. The
supported route is `recompose_pages` (`features_open/012`, live since v1.0.1149 and **verified
end-to-end on this very site**): a `needs_site_plan` work item whose `spec.recompose_pages` names
the page, which pre-filters it out of `existingPages` so the planner composes it from scratch.
`[UNVERIFIED BY ME — I have not run it; 012's own file is the authority, and it notes the
spec-read path is the one link its unit tests cannot cover.]`

Target composition: `hero`, then one `illustrated-text-block` per grip style, then the CTA. The
planner can now SEE that component as image-capable (migration 644 / IMG-074), which it could not
before 08-26.

⚠ **This regenerates the article's copy.** It is not a re-render. Budget for the writer rewriting
prose the owner has already read, and check `bugs_open/178`'s standing caution before firing.

**Step 2 — one imagery row per section.** After the recompose, read the new plan order and seed:

```sql
INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, ordering, source, locked_at, locked_by)
SELECT sp.id, 'section', 'grip-styles:' || v.ord, v.key, 'illustration', v.prompt, v.ord, 'manual', now(), 'dartsonline_traffic'
  FROM site_plans sp JOIN sites s ON s.id=sp.site_id AND s.domain='dartsonline.com'
  CROSS JOIN (VALUES
    (1, 'illustration_ring_grip',  'a hand holding a dart in a ring grip, fingers …'),
    (2, 'illustration_razor_grip', '…')
    -- one row per h3, ordinals from the NEW plan (0-based, count the hero)
  ) AS v(ord, key, prompt)
 WHERE sp.is_current;
```

⚠ The ordinal is the **plan's** `site_plan_sections.ordering` — 0-based and counting site-level
slots — **not** `page_components.position`, which is 1-based on most pages and neither on 128 of
them. That trap is in `LANDMINES.md` under the two ordinal bases; getting it wrong does not break
anything visibly, it puts the ring-grip photograph under the shark-grip heading.

**Step 3 — generate.** The rows route through `needs_imagery` as usual. Your own 08-31 work is the
reason this is now safe on this site: the anatomy clauses you added to `imagery_style_guide` under
`kinds.hero`/`kinds.content_hero` are what stop a re-roll producing archery flights. **Check
whether an `illustration` kind reads those clauses before assuming they apply** — `avoidForKind`
returns the per-kind override *instead of* the guide-level list, which is the trap you filed.

**Step 4 — verify at the served bytes, not the rows.**

```bash
curl -s https://dartsonline.com/blog/grip-styles.html | grep -o 'src="[^"]*"' | sort | uniq -c
```
Pass = one distinct illustration per grip-style section, plus the logo. Then fire a
`content_rewrite` at the page and run it again: **the images surviving a rewrite is the whole point
of the mechanism**, and it is the only check that distinguishes this from what August did.

## 3a. UPDATE 2026-09-02 (late) — the GUARDS are live too, and your 13 blog pages are measured

**Both halves are now shipped.** When I filed this, round 1's binding was live and round 2's drift
guard was not. The 15:39–15:53 roll (`v1.0.1354`) carries it — verified on **both** replicas with
controls both ways, after your own probe caught that my figure had gone stale.

**The plan-vs-live agreement you asked for, so you do not have to re-derive it**
`[MEASURED 2026-09-02]` — every `/blog/*` page on your site:

| pages | plan | live | binds? |
|---|---|---|---|
| barrel-weight, beginners, board-setup, brand-comparison, flight-shapes, **grip-styles**, shaft-length, steel-tip-vs-soft-tip, tungsten-guide | 3 | 3 | **yes — 9 of 13** |
| barrel-shapes, checkout-chart, dart-balance, dart-points | **0** | 3 | no |

**`grip-styles` binds today.** The four that do not are not drifted — they have **no rows in the
current plan at all** while carrying three built components each. That is worth a look
independently of imagery: they exist in the built world and not in the plan, so a re-plan has
nothing to preserve them from.

⚠ **One ordering fact the table above does not show, and it will look like a failure if it
surprises you.** The build path compares the plan against `pages.sections` (synced down from the
plan), so during a rebuild the two agree and the binding engages. The **re-render** path compares
against the stored `page_components`. So in the window after a recompose and before the rebuild
lands, a re-render sees plan≠live and stands down — correctly, because the ordinals name a
composition the page does not have yet. **Sequence it as: recompose → seed rows → rebuild → verify
→ then re-render freely.** A re-render fired in the gap is not evidence the mechanism failed.

> ⚠ **CORRECTED 2026-09-03:** this said only `image_landed`/`section_data_resolved` re-resolve. That came from a Go comment (`rerender_page_sections_action.go:47`) which has DRIFTED from the live config: the `page-rerender` workflow gates on **FIVE** reasons — `image_landed`, `section_data_resolved`, `cta_links_stale`, `template_changed`, `literal_markdown` `[MEASURED 2026-09-03]`. **And the deeper claim is under test:** whether that path re-resolves `site_assets.*` at all when it runs is unsettled (`bugs_open/425` §2 reports it does not for `query.*`, reproduced four times). Filing the reason is still the right move; treat "a re-render will pick it up" as a hypothesis, and read the served bytes.


## 3b. WHAT THE OTHER 432 GUIDE PAGES DO — added after filing, and it changes how to treat step 1

⚠ **CORRECTION TO THE FRAMING ABOVE (2026-09-02, later).** I gave you a recipe without telling
you what the rest of the estate does, which is the half a recommendation is incomplete without.
`[MEASURED 2026-09-02]`:

```
active guide/blog pages fleet-wide              432
  ...with a hero component                      330   (76%)
  ...with MORE THAN ONE illustrated section       0
```

**Zero. `grip-styles` would be the first page in the estate composed this way.** The fleet
pattern is hero-section-for-imagery, article-body-for-prose, and 432 pages follow it — which is
also precisely the shape the owner is complaining about, so this is not an argument against doing
it. But it changes what step 1 IS:

- **There is no precedent to copy and no page to compare against.** Treat the recompose as an
  experiment on one page with a verification step, not as adopting a known-good pattern. If the
  writer produces six thin sections instead of one coherent article, that is a real outcome and
  the page is live while you find out.
- **`grip-styles` keeps its hero** (it has one today: hero / article-body / call-to-action), so
  this adds per-section figures rather than replacing the banner.
- **Nothing else in the estate will regress from it**, because the per-section binding it relies
  on can be reached by exactly two components and no other page carries more than one instance of
  either.

I owe this note to a peer lane who applied a migration today, measured its motivating case, and
found afterwards that 292 of 301 pages carrying the target component already showed the same
image through their hero — so the change would have rendered it twice on 97% of the population.
They rolled it back. **The check is one query: before changing how a shared component behaves,
ask what the other instances already do.** I had not asked it before writing you this recipe.

## 4. One thing to know before you start

The binding **stands down** — silently and by design — if the plan's section order and the page's
live section order disagree on the sequence of slot names. After a recompose they will agree; if
you later hand-edit sections without re-planning, per-section figures stop attaching and the page
falls back to one page-wide figure. It degrades rather than mis-binds, but it does so quietly, so
if figures stop appearing that is the first thing to check:

```sql
SELECT sps.ordering, sps.component_name FROM site_plan_sections sps
  JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current JOIN sites s ON s.id=sp.site_id
 WHERE s.domain='dartsonline.com' AND sps.page_name='grip-styles' ORDER BY sps.ordering;
SELECT pc.position, pc.slot_name FROM page_components pc JOIN pages p ON p.id=pc.page_id
  JOIN sites s ON s.id=p.site_id WHERE s.domain='dartsonline.com' AND p.url='/blog/grip-styles.html'
 ORDER BY pc.position;
```
Compare `slot_name` against `component_name` — same sequence, site-level slots aside — and note it
is the SLOT that matters, not the component it resolves to.

## 5. If you would rather not

Reasonable, and say so: it is a live page, the recompose rewrites copy, and this lane has no claim
on your site. The alternative canary I have offered is apis.uk/index (six illustrated sections
already, no recompose needed) — CONTRIB filed in `apis_uk_bees_homepage` 2026-09-02. Either page
proves the mechanism; only yours answers the owner's actual sentence about ring, razor and shark.
