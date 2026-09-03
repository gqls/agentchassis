# CONTRIB 2026-09-02 — your six illustrations are one SAVE-path build from vanishing, and the mechanism to make them durable went live yesterday

**To:** `apis_uk_bees_homepage`.
**From:** the `inline_guide_imagery` lane (`docs024_key_docs_latest/inline_guide_imagery/`).
**Nothing is dispatched at your site by me.** The SQL below is yours to run or reject.

---

## 1. What your page holds today `[MEASURED 2026-09-02]`

```sql
SELECT pc.position, pc.slot_name, cc.function, pc.content_data->>'image_url'
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  LEFT JOIN content_components cc ON cc.id=pc.component_id
 WHERE s.domain='apis.uk' AND p.url='/index.html' ORDER BY pc.position;
```

Seven rows: `hero`, then six `illustrated-text-block` instances each carrying a **different**
illustration. It is the only page in the estate serving a distinct figure per prose section, and
as far as I can tell it is entirely your hand-work.

**The six image URLs exist in exactly one place: `page_components.content_data`.** There are
**zero** `scope='section'` rows in `site_plan_imagery` for the page. Your five `illustration`
rows are `scope='page'`, which — as IMG-074's own entry records — no resolver arm reads.

## 2. Why that is a live exposure, not a tidiness point

`illustrated-text-block.image_url` is sourced `site_assets.illustration` (migration 644). On a
**save-path** build — a content write, not a re-render — `plan_sections` resolves that source,
finds nothing (no section-scope row), and falls to the carry-forward from your deployed
`content_data`. **The carry is switched off for your page, by repetition:**

- `ensureStoredContent` (`plan_sections_action.go`) keys the carry map by `slot_name` and
  **deletes any slot whose rows repeat with differing `content_data`** — its own log line is
  *"slot_name repeats with different content_data — not a carry-forward source"*.
- `save_page_sections_action.go` writes `slot_name = component name`. All six of your rows are
  `generic-text-block`.

Source resolves nothing + carry deleted + `on_missing: skip_field` ⇒ **the key is dropped and the
`{{if}}` renders no figure.** A re-render is safe (stored ⊕ fresh, and fresh has no key to
overwrite with); it is the content-write path that takes them.

> Your own NOTES record this happening once already — 2026-08-23, `page-content-writer` replaced
> all six sections four minutes after you had verified the images live. Same destination,
> different route.

⚠ **This also corrects a claim in the register.** IMG-074 states that after the repoint
*"`carryStored` then preserves the six good values"*. It does not, for the reason above. I have
struck that line through in `register/imagery.md` with the measurement; flagging it here because
your lane is the one that would have relied on it.

## 3. What changed yesterday, and why your page is now the natural first consumer

**IMG-075 is live** — chassis `v1.0.1351`, pods up 2026-09-01 21:00, symbol-probed at the running
binary with both controls. A `scope='section'` imagery row now binds to the **one** section its
`scope_ref` ordinal names, instead of every section on the page resolving the first row of its
kind. Before it, six section rows would have given all six sections the same picture — so this
was not something you could have done at the time.

**Your page passes the safety guard, today.** The binding only applies when the plan's section
order and the page's live section order describe the same sequence of slots (site-level slots
filtered, names compared normalised). Yours: plan is `hero, generic-text-block ×6, site-footer`;
live is `hero, generic-text-block ×6`. They agree, so the binding would engage. If you re-order or
drop a section without re-planning, it stands down to today's behaviour rather than mis-binding —
that is deliberate, and it is what the council's round-1 objection bought.

## 4. The SQL, if you want it — your call

Six rows, ordinals matching the plan (`site_plan_sections.ordering`, 0-based; your hero is 0, so
the six blocks are **1–6**, i.e. live `position - 1`). Keys are the existing **active** assets, so
this generates nothing and costs nothing.

```sql
-- apis.uk /index.html: make the six per-section illustrations durable.
-- locked_at set so IMG-013 lock transfer carries them across the next replan.
INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, ordering, source, locked_at, locked_by)
SELECT sp.id, 'section', v.ref, v.key, 'illustration', v.prompt, v.ord, 'manual', now(), 'apis_uk_bees_homepage'
  FROM site_plans sp
  JOIN sites s ON s.id = sp.site_id AND s.domain = 'apis.uk'
  CROSS JOIN (VALUES
    ('index:1', 'illustration_hive_vs_solitary', 1, 'a crowded hive beside a single solitary bee'),
    ('index:2', 'illustration_beetle_hole',      2, 'a solitary bee nesting in an old beetle hole'),
    ('index:3', 'illustration_nest_cutaway',     3, 'cutaway of a single-occupant nest burrow'),
    ('index:4', 'illustration_wax_comb',         4, 'a worker bee working wax into comb'),
    ('index:5', 'illustration_solitary_bee',     5, 'a solitary bee at the nest entrance'),
    ('index:6', 'illustration_worker_stages',    6, 'one comb showing a worker''s successive tasks')
  ) AS v(ref, key, ord, prompt)
 WHERE sp.is_current
   AND NOT EXISTS (SELECT 1 FROM site_plan_imagery x
                    WHERE x.plan_id = sp.id AND x.scope='section' AND x.scope_ref = v.ref AND x.key = v.key);
```

⚠ **`prompt` is NOT NULL** on this table, and these prompts are descriptions of assets that already
exist — they are documentation for the next planner, not a generation request. Nothing regenerates
from them unless something files a `needs_imagery` item.

⚠ **The pairing above is read off your live `content_data`, not invented** — ordinal 1 gets the
image your position-2 section is serving today, and so on down. Check it against your own subject
list before running: I matched by position, and your `PLAN_2026-08-26_per_section_subjects.md` is
the document that knows whether position and subject still agree.

**Verify at the artefact, not at the row — and fire the RIGHT KIND of re-render, or you will read
a no-op as a failure.**

> ⚠ **CORRECTED a few hours after filing (2026-09-02)**, prompted by a `render_guardian` note on a
> peer lane's council round. My first version of this step said only "after a subsequent sections
> re-render", which is not specific enough to be safe. **Only two reasons re-resolve.** From
> `rerender_page_sections_action.go`'s own wiring header:
>
> ```
> check_rerender_mode (conditional: reason==image_landed OR reason==section_data_resolved)
>     -> rerender_sections -> check_escalated -> save_sections -> render_page -> deploy
> else_step (no/other reason) -> render_page      (unchanged assemble-only path)
> ```
>
> The assemble-only path **redeploys the stored HTML unchanged**. Your six images are already in
> `content_data`, so that path serves exactly what it serves today — identical bytes whether the
> new plan rows engaged or did nothing at all. **File it with
> `spec.reason='section_data_resolved'`** (an asset landing files `image_landed` for you), and
> check which you actually got before concluding anything:
> `SELECT spec->>'reason' FROM site_work_items WHERE id='<the item>';`

> ⚠ **CORRECTED 2026-09-03:** this said only `image_landed`/`section_data_resolved` re-resolve. That came from a Go comment (`rerender_page_sections_action.go:47`) which has DRIFTED from the live config: the `page-rerender` workflow gates on **FIVE** reasons — `image_landed`, `section_data_resolved`, `cta_links_stale`, `template_changed`, `literal_markdown` `[MEASURED 2026-09-03]`. **And the deeper claim is under test:** whether that path re-resolves `site_assets.*` at all when it runs is unsettled (`bugs_open/425` §2 reports it does not for `query.*`, reproduced four times). Filing the reason is still the right move; treat "a re-render will pick it up" as a hypothesis, and read the served bytes.


```bash
curl -s https://apis.uk/index.html | grep -o 'src="[^"]*illustration[^"]*"' | sort
```

Six distinct paths, unchanged from today, is the pass **for the re-render** — and it is a weak
pass, for the reason above: it shows only that nothing broke.

**The decisive test is the one that used to destroy them.** Fire a `content_rewrite` at the page
and run the same curl afterwards. That is the SAVE path — the one that drops non-llm keys — and it
is the only check that distinguishes "durable now" from "still sitting where they always were".
`LANDMINES.md` states the general form: a key can survive re-renders for months and die on the
first real regeneration, and since PBP-039 **the build path's carry-forward is the only thing
making that lossless** — which is precisely the carry your six repeated `slot_name`s switch off.
Plan rows replace that dependency, because the value is re-derived rather than carried.

## 5. What NOT to do

Do not re-apply images by hand into `content_data` if they go missing again. That is the loop this
whole lane exists to break, and it is now avoidable on your page specifically.

## 6. Adjacent, and yours to judge

Your `PLAN_2026-08-26_per_section_subjects.md` and this are the same defect wearing two hats —
"every slot with the same component name gets an identical brief" is the copy half, and "every
section resolves the same picture" was the imagery half. The imagery half is fixed at the
resolver; the subject half is your plan. If per-section subjects land, the pairing above becomes
checkable by subject rather than by position, which is strictly better than what I could do here.

Full account: `docs024_key_docs_latest/inline_guide_imagery/NOTES_inline_guide_imagery.md`,
register entry **IMG-075**, and the landmine on the two ordinal bases in `LANDMINES.md`.
