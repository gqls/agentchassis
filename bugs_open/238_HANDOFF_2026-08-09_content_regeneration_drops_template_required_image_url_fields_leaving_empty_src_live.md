# 238 — a content regeneration drops the template's required image-URL fields, shipping `src=""` to the live page

**Status: OPEN. Live on finetuning.uk's homepage right now** — five
`<img class="csg-card-image" src="">` on `/index.html`, verified at the served
page 2026-08-09 ~16:10Z.

**I caused this run to happen** (see §Provenance) — the item was promoted and
dispatched by an improvement-loop firing I made for an unrelated reason. The
defect is in the handler, not in the firing, but the honest ordering is that the
homepage was intact at 14:47Z and is broken now.

## Mechanism

A `tone_shift` work item (`design-audit_tone_shift_index_…`, handler
**`page-build-handler`**, claimed by `build-dispatch-loop`, completed
**2026-08-09 15:18:10Z**) regenerated the `case-studies-grid` section data on
`/index.html`. The page's components were rewritten at 15:17:19Z.

The regenerated `content_data` contains, for all five cards:

- `card1_image_alt` … `card5_image_alt` — **present**
- `card1_image_url` … `card5_image_url` — **absent entirely**

The component template still requires them. `content_components` row
`3f946437-1dc7-4164-987d-620933589076` (`case-studies-grid`) emits, five times:

```html
<img class="csg-card-image" src="{{.card1_image_url}}" alt="{{.card1_image_alt}}" loading="lazy" />
```

A missing key renders empty (`missingkey=zero`), so the live page serves
`src=""` five times with the alt text fully intact — which is what makes it read
as a working image block in every check that looks at markup shape.

**The regeneration also rewrote the case studies themselves**, not just their
tone: card 1 went from "Cutting Quote Turnaround for a Facilities Management
Firm" to "Cutting the Manual Work Out of Quote Requests", and card 4 changed
subject from a private-AI deployment to "coordinating agent processes". So the
five images that *do* exist no longer map 1:1 onto the five cards — this cannot
be repaired by pasting the old URLs back.

## Why this is a class, not one site

The failure needs only two things, both common: a component whose template
sources an image from `content_data`, and any handler that regenerates that
section's data. The generator is not told that certain keys are structural
rather than editorial, so it reproduces the ones that look like copy
(`*_image_alt`) and drops the ones that look like plumbing (`*_image_url`).

Worth measuring before sizing a fix (NOT done here — stated as the open
question, not as a finding): how many content-driven components fleet-wide have
a `{{.*_image_url}}` in their template, and how many of their current
`content_data` rows are missing at least one of those keys.

## Evidence

```sql
-- the template requires the keys
SELECT substring(html_template from '<img class="csg-card-image"[^>]*') FROM content_components
WHERE id='3f946437-1dc7-4164-987d-620933589076';

-- the content no longer has them (returns 0)
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND p.url='/index.html' AND pc.slot_name='case-studies-grid'
  AND pc.content_data ? 'card1_image_url';

-- the item that did it
SELECT item_type, handler_agent, completed_at FROM site_work_items
WHERE site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND item_key='design-audit_tone_shift_index_1368e337-dd1d-4799-bbb3-8221a1b79bcc';
```

At the artefact:
`curl -s https://finetuning.uk/index.html | grep -c 'csg-card-image" src=""'` → 5.

## Why no detector caught it

The site's own `empty_src` checker exists and is live (shipped 2026-08-03,
pod-verified). It has not fired because **site discovery has no recurring
driver** — the last discovery pass on this site ran at 13:51Z, roughly 90
minutes *before* the regression. That is `bugs_open/230`, and this is a concrete
instance of its cost: a defect introduced by the repair pipeline itself sits
unseen because nothing re-examines the site afterwards.

**The transferable shape:** a repair loop that regenerates content and a
discovery pass that only runs on demand means the loop can introduce exactly the
class of defect the loop exists to remove, and report success.

## Fix candidates, ordered by what closes the door

1. **Make structural keys non-regenerable.** Have the content generator receive
   the existing `content_data` and merge rather than replace, or mark
   `*_image_url` (and siblings) as preserved fields the LLM cannot emit or drop.
   Only this makes the bad state unrepresentable.
2. **Fail the write.** Validate regenerated section data against the template's
   required keys before persisting; refuse and raise a work item on a missing
   key. Cheaper, and turns a silent live regression into a queued item.
3. **Detect after the fact** — give discovery a recurring driver (`bugs_open/230`)
   so `empty_src` catches this within a cycle. Necessary regardless, but it is
   detection, not prevention.

## Immediate state of the affected site

- **The five case-study images now exist and serve** (`/assets/images/case-study-*.jpg`,
  all HTTP 200, `image/jpeg`, 52–94KB, distinct). That work is done and is not
  what is broken.
- **`/index.html`** — five empty `src`. Needs new content that carries image
  URLs matching the *rewritten* case studies. Not a paste-back.
- **`/case-studies.html`** — unaffected by the rewrite (its `case-studies-list`
  template **hardcodes** the five paths, unchanged since March). But its
  `rendered_html` dates from 2026-04-23 and predates the assets, so the page
  does not yet show them. **A rerender of that page alone should light up all
  five images with no content decisions required** — the cheapest real win here,
  and the one to do first.

## Provenance

Found while verifying `bugs_open/233`-adjacent work: the finetuning imagery
Phase 1 task (`docs/agent_docs/docs024_key_docs_latest/finetuning_uk_repair/PLAN_2026-08-04_imagery_then_visual_designer.md`).
The images were generated successfully; checking whether the *pages* referenced
them is what exposed this. Verifying at the asset URL alone would have reported
Phase 1 complete with the homepage newly broken.
