# CONTRIB from the `bugs_open/283` / RFC_032 lane, 2026-08-24 — four guide pages serve `<section id="">`; the owner has asked for them to be FIXED, on this lane

**Owner ruling, 2026-08-24 (relayed verbatim in intent): "We should fix this even if it means
content generation."** Filing here because these are idea.uk guide pages and this lane owns
their content path; the cause was ours and is already dead, the casualties are yours.

## What is broken, measured (2026-08-24, live DB + served pages)

Four `page_components` rows, all `generic-text-block`, all with `rendered_html` last written
**2026-08-12**, each serving exactly one `<section id="">` on the live page (fetched,
cache-busted 2026-08-23):

| page | slot |
|---|---|
| `/guides/creating-ideas/index.html` | generic-text-block |
| `/guides/patents/index.html` | generic-text-block |
| `/guides/testing-it/index.html` | generic-text-block |
| `/guides/user-acceptance/index.html` | generic-text-block |

`id=""` is invalid markup, invisible to the fleet's collision detector (`reElementID` requires
a non-empty id — a fix for that blindness is now owner-ruled and being planned on the RFC_032
track), and it blocks any future anchor/script use of those sections.

## The cause — dead since 2026-08-23, so this cannot recur while you repair

The `generic-text-block` template used to declare `id="{{.ComponentID}}"`. The path that wrote
these rows on 08-12 bound no value for it; Go's missingkey=zero rendered `<no value>`, which
`component_library.go:1170` strips to `""`. On 2026-08-23 the template was converted to
`{{.InstanceID}}` (owner-ruled RFC_032 §8) and every render path binds that token, so a fresh
render of these sections produces `id="c-generic-text-block"` — correct here, since none of the
four pages repeats the component.

## What repair needs — and it is probably NOT content generation, try the cheap step first

**Their `content_data` is intact**: all four rows carry `content` + `heading`, exactly what the
template needs. This is not the missing-section-data class this lane usually handles. So:

1. **First try one plain `page_rerender`** (`spec.reason='template_changed'`, and include
   `page_name` — `save_page_sections` silently skips without it, reporting success). ⚠ The
   open question you'll hit: the 2026-08-23 fixer-driven rerender of these pages read
   `complete` and deployed files, **yet the stored rows kept their 08-12 bytes** — something
   carried these sections rather than re-rendering them, and I did not extract which arm
   (candidates: the pre-check's escalate/carry logic in `rerender_page_sections_action.go`
   ~:390-420). One canary page, then read the row's `updated_at` — the item status will say
   `complete` either way and proves nothing.
2. **Only if it carries again**: the content-generation route the owner has sanctioned —
   regenerate the section through the writer path, which re-renders as it saves.
3. Verify at the artefact per this estate's bar: cache-busted fetch, assert
   `id="c-generic-text-block"` PRESENT (a positive assertion — "no empty id" also passes on a
   page that lost the section).

## Existing queue state you already own (don't double-file)

- `/guides/creating-ideas/` has `content_rewrite` **needs_human_review** (2026-08-23) and
  `placeholder_contact` needs_human_review (08-16).
- `/guides/feedback-loops/` and `/guides/testing-it/` have `content_rewrite` **unresolved**
  since 08-05.
Whatever unsticks those items can carry this repair with it; the four pages above are the
complete list for the empty-id defect (formal corpus census 2026-08-24: 6 rows fleet-wide with
`id=""`, these 4 plus 2 on dartsonline.com — the dartsonline pair is a *different* cause,
`category-listing`'s `id="{{.category_slug}}"` rendering empty, and is being routed via the
RFC_032 detector work, not this lane).

— 283/RFC_032 lane session, 2026-08-24. Trail: `bugs_closed/283_CONTINUE_HERE_2026-08-24.md`,
`architecture_review/RFC_032…md` §9, lane NOTES sessions 10–12.
