# 140 — the `contact-info` component fabricates a phone number and office hours when the data is absent

**Filed 2026-07-29 (oufe workstream, found while closing the render-audit loop).
Owner lane for the component: none known — this file is the routing point.**

## The defect

The shared `content_components` row `function='contact-info'` renders three cards
— Email, Phone, Hours — **unconditionally**. Each card's value is a Go-template
fallback chain that, when the field is absent from `content_data`, **invents
one**:

```
{{if .phone}}…{{else}}+1234567890{{end}}          ← tel: href
{{if .phone_display}}…{{else}}+1 (234) 567-890{{end}}
{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}
```

So a site that supplies only an email publishes a fabricated US phone number and
fabricated office hours, styled identically to its real details. On a platform
whose whole claims apparatus exists to stop unsourced assertions reaching a page,
the component library itself asserts unverifiable business facts by default.

## Measured, 2026-07-29 (all six live uses)

```sql
SELECT s.domain, pc.content_data ? 'phone' AS has_phone, pc.content_data ? 'hours' AS has_hours,
       pc.rendered_html LIKE '%Monday – Friday, 9am – 6pm%' AS fake_hours
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
JOIN content_components cc ON cc.id=pc.component_id WHERE cc.function='contact-info';
```

| domain | has_phone | has_hours | renders fabricated hours |
|---|---|---|---|
| leopardessconsulting.co.uk | t | f | **t** |
| idea.uk | **f** | f | **t** |
| finetuning.uk | t | f | **t** |
| gaswholesalers.com | t | f | **t** |
| ai-agent-orchestration.com | t | f | **t** |
| fundamentallyai.com | t | f | **t** |

- **6 of 6 live uses render the fabricated hours**, and this is SERVED, not just
  stored: `curl` found `Monday – Friday, 9am – 6pm` on the live contact pages of
  idea.uk, leopardessconsulting.co.uk and gaswholesalers.com (3 of 3 sampled).
  Nobody stated those hours; no evidence register holds them.
- **idea.uk's rendered card shows `+44 (0) 7934 524 911` while its `content_data`
  has no `phone` key** — the stored artefact carries a value the data has lost
  (the `bugs_open/117` stored-artefact drift family, on a claim about how to
  reach a business). leopardessconsulting renders the SAME number and does have
  it in data. Whether that number is genuinely shared across those two
  businesses is an owner question, not something to "fix" — but idea.uk's copy
  of it is currently unbacked by any data row. [UNVERIFIED: whose number it is.]
- oufe.com does NOT use the component. Its contact page planned a `contact-info`
  section that was never built; on inspecting the template this workstream
  **removed the section from the plan rather than populate it**, because
  populating only the email would have published the fake phone + hours.
  (That unbuilt section had a second cost: the partial-build guard in
  `UpdatePageStatusAction` correctly refused the deploy stamp forever —
  2 of 3 planned sections — so the page sat `needs_rebuild` while live, and the
  render audit skipped it. Fixed by the plan edit; recorded in the oufe NOTES.)

## Why nothing caught it

- The fallbacks live in the component template, so every render "succeeds".
- claimscan targets business figures and evidence-register terms;
  a phone number and a weekday-hours string match none of its patterns.
  [INFERRED from its pattern set — not run against these pages to confirm.]
- `bugs_open/111` (footer Contact heading over nothing) already established the
  neighbouring rule for chrome: **contact furniture renders only when the datum
  exists** (`RenderFallbackFooter` gated 2026-07-28, `d4731109d`). The section
  component was never brought under that rule — same family, other surface.

## Fix candidates, ordered by what closes the door

1. **Make absent mean absent in the template** — wrap each card in
   `{{if .phone}}…{{end}}` / `{{if .hours}}…{{end}}`, delete the invented
   fallback values entirely (keep the email fallback only if `sites.email` is
   injected — the 111 house convention `<name>@contactforsales.com` is the
   legitimate source). This makes the bad state unrepresentable: a datum nobody
   supplied cannot render. **Blast radius: all 6 sites' contact pages change on
   their next rerender — they LOSE the fabricated hours card (and idea.uk its
   orphaned phone).** That is a correction, not a regression, but it is a
   shared-component change across six sites owned by other lanes: it needs a
   council run and/or the owner's nod, not a quiet patch from this lane.
2. A migration stamping real `hours`/`phone` into the six rows — only if the
   owner actually states them. Does not close the door for site seven.
3. Per-site stored-instance patches — the 253-precedent tool, wrong here: six
   instances and the template would re-fabricate on any rerender of an unlocked
   row.

## How to verify

Re-run the census above expecting `fake_hours = f` everywhere after the sites
rerender; curl the three sampled pages for the hours string; and check a site
supplying ONLY email renders exactly one card.

## Relations

`bugs_open/111` (contact furniture gating, footer surface) ·
`bugs_open/117`/`118` (stored artefacts drift) · `bugs_open/122` (hard-coded
ink family — same "component asserts what the site never said" shape, colours
instead of facts) · oufe NOTES 2026-07-29 (discovery context).
