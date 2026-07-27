# 097 — CTA integrity misses in-body card links to pages that were never built

**Filed** 2026-07-26 from the oufe.com workstream.
**Severity** medium — publicly visible broken links on a live site, including a
main navigation item.
**Status** OPEN.
**Related** `bugs_open/052` (listings re-advertise unbuilt pages) — same family,
different surface: 052 is about generated listings, this is about hand-planned
in-body cards. Also `bugs_open/049` (broken links fleet-wide).

## Symptom

oufe.com went live with **six broken content links on the homepage**, and the
header's own **Cases** nav item 404ing. The detector raised exactly two items:

```
unresolved_cta | Unresolved CTA on index ('hero'): no real-page destination for secondary_cta_url
unresolved_cta | Unresolved CTA on index ('call-to-action'): no real-page destination for secondary_cta_url
```

Those two were real and correctly found. The six that were missed:

| card link | target |
|---|---|
| `/restructuring-plan` | never planned |
| `/creditor-waterfall` | never planned |
| `/cases/thames-water` | page exists but its url is `/blog/thames-water.html` |
| `/tools` | never planned |
| `/framework` | never planned |
| `/cases` | page exists but its url is `/cases/index.html` |

## Cause

The check covers the *named CTA fields* of components that declare them —
`cta_url`, `secondary_cta_url`, `primary_cta_url`. It does not walk **arrays of
child objects** inside `content_data`, and `info-card-grid` keeps its
destinations at `content_data.cards[*].link_url`.

Two of the six are especially worth noting, because they show the failure is not
only "page missing":

- `/cases/thames-water` and `/cases` were written from **the plan's intent**, not
  from the page record. The pages exist; their urls are `/blog/thames-water.html`
  and `/cases/index.html`. So a link can be broken even when its target page is
  built and deployed — checking "does a page with this name exist" is not
  sufficient, the check has to resolve against `pages.url`.

## Why the templates did not save us

The component templates are already fail-closed and correct:

```
info-card-grid:  {{if .link_url}}<a href="{{.link_url}}">…</a>{{end}}
hero:            {{if and .cta_text .cta_url}}<a href="{{.cta_url}}">…</a>{{end}}
```

A card with **no** url renders as plain text — good. But a card with a url that
points nowhere renders a perfectly well-formed anchor. The template can only
protect against a missing destination, never against a wrong one. That check has
to live where page urls are known.

## Fix candidates, ordered by what closes the door

1. **Resolve every internal href in rendered component HTML against `pages.url`**
   at the same point the CTA check runs. This catches the whole class regardless
   of which field or nesting the link came from, and it catches wrong-url as well
   as missing-page. It is also the only candidate that would have caught all six.
2. **Walk arrays in `content_data`** for `*_url` / `link_url` keys, in addition to
   the top-level named fields. Cheaper, but it enumerates shapes and the next
   component with a new nesting reopens the gap.
3. Add `link_url` to the named-field list. Fixes this one component only.

Candidate 1 makes the bad state detectable wherever it appears; 2 and 3 both
require someone to remember to extend a list, which is the shape that produced
this bug.

## How to verify a fix

On a site with a card grid, point one card at a path with no page and another at
a page whose `url` differs from the path used. Run the check. Both must be
flagged. Then confirm a correct link is **not** flagged — a checker that fires on
everything is as useless as one that fires on nothing.

**Induce both faults.** A clean report on a healthy site proves only that the
check runs.

## Second instance, different site — found 2026-07-26 while closing `bugs_closed/052`

`robot-hands.com/learning-center.html` is live, linked from the site's own
navigation, and carries five in-body category links that all 404. Verified with
`curl`, not inferred:

```
404  https://robot-hands.com/learning-center/calculators
404  https://robot-hands.com/learning-center/technology-guides
```

(the other three — `/learning-center/application-guides`,
`/learning-center/comparisons`, `/learning-center/specification-workflow` — are the
same shape and were not individually fetched. `[UNVERIFIED]`)

Two things this instance adds to the oufe.com evidence above:

- **It is not confined to one site or one component library.** Same failure, a
  different site built at a different time, so the missed-nesting cause is
  platform-wide rather than an oufe-specific component choice.
- **These links have no `.html` extension**, while every real page on the site is
  served at `<name>.html`. So the destination was written from the plan's intent
  rather than from `pages.url` — the same sub-cause as `/cases/thames-water` above,
  and further evidence for fix candidate 1 (resolve rendered hrefs against
  `pages.url`) over candidates 2 and 3, which would not catch a link whose *target
  page exists under a different url*.

Not investigated further here — this file owns the mechanism; recording the instance
so the fix's verification set includes a second site.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — oufe is clean, robot-hands is not, and the detector gap is untouched

Verification sweep, not a fix. Both instances re-probed live over HTTPS.

**oufe.com's six are gone.** The homepage now emits only resolvable targets —
`/about.html`, `/cases/index.html`, `/contact.html`, `/index.html` (200 confirmed on the
two checked). `/restructuring-plan`, `/creditor-waterfall`, `/cases/thames-water`,
`/tools`, `/framework` and bare `/cases` are all absent from the served markup.

**robot-hands.com's are all still live, and there is a sixth this file missed.** Every one
individually probed, so the `[UNVERIFIED]` marker on three of them above is now discharged:

```
404  /learning-center/calculators            404  /learning-center/comparisons
404  /learning-center/technology-guides      404  /learning-center/specification-workflow
404  /learning-center/application-guides     404  /matchmatrix/methodology     <- NEW
```

`/matchmatrix/methodology` is the same shape from a different hub, and note the site *does*
serve `/matchmatrix-methodology.html` (200) — so it is the "target exists at a different
url" sub-class, not a pure phantom. That strengthens fix candidate 1 over 2 and 3 again.

### What changed underneath this bug, and what did not

`bugs_open/079`'s `RepairPageLinks` is now live and firing at the build gate
(`validate_page_content.go:356`), and it resolves **every** internal `<a href>` in the
assembled page against `pages.url` — which is literally this file's fix candidate 1, arriving
from another lane and at a different stage. On the numbers above it would have handled
oufe's six: bare `/cases` normalises to `/cases/index.html` and **rewrites**; the four true
phantoms **unlink**. `/cases/thames-water` (real page at `/blog/thames-water.html`) would
unlink rather than repoint — the 404 dies, the intended destination is still lost.

**Two reasons this file stays OPEN:**

1. **The detector is unchanged.** `ctaFieldNames`
   (`resolve_internal_links_action.go:98-105`) still lists six components, still has no
   `info-card-grid`, and nothing walks arrays in `content_data` — grep for `link_url` in that
   action returns nothing. So `unresolved_cta` still misses exactly what it missed, and the
   only thing now catching the class is a *repair* at a later stage that silently deletes the
   link instead of reporting it. Detection and repair are not substitutes.
2. **A write-path repair does not reach a deployed page.** robot-hands' six have been live
   since before the fix and will stay live until that page is rebuilt. There is no sweep that
   would find them (`bugs_open/083`: `improvement-sweep` off since 2026-05).

**Sharpest next action:** decide whether candidate 1 is now redundant at the gate and should
instead be implemented where this file said — at the CTA check — so the finding is *reported*
as well as repaired. Then a rebuild of `robot-hands.com/learning-center.html` clears the live
damage. Neither needs a diagnosis run; the cause is fully known.
