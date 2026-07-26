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
