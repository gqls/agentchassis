# 028 — a page-build no-op reported `complete` and deployed a page built from other pages' components

**Filed 2026-07-19** (relojistas thread). **Status: OPEN.** Silent success — the most
expensive failure shape this platform has, because every status field says the work
happened.

> Numbering note: a concurrent session also used `027` on 2026-07-19
> (`027_..._content_hero_unstyled_on_sites_without_a_style_guide.md`). Per
> `bugs_closed/README.md` numbers are never reassigned and a bare number is ambiguous —
> resolve by slug. This case is `028`.

## Symptom

`page-build-handler` recorded a no-op for `glosario-tourbillon` on relojistas.com:

```
site_work_items.error:
  page-build-handler no-op: no sections ready to build
  (empty spec sections, or all sections deferred for missing data)
```

and the item still ended at **`status='complete'`**, the page at
**`build_status='deployed'`**, live at `/glosario/tourbillon.html`.

The handler composed nothing. The page shipped anyway, carrying two components that were
not its own:

- `hero` holding the **site homepage's** headline — "Relojería en español: noticias, guías
  y glosario" — not a word about tourbillons;
- `content-block-about` holding generic about-the-company copy.

A page titled "Tourbillon" published to a live site saying nothing whatsoever about a
tourbillon, and nothing anywhere reported a problem.

## Why this is distinct from `bugs_open/015`

015 produces the *same error string* from a different cause (a mistyped `page_type`
orphaning the page from its machinery). Crucially, **in 015 the work item correctly went to
`needs_human_review`.** That is the right outcome for a no-op.

This case is the failure of that outcome: the identical no-op reached `complete` and
deployed. So 015 is "the page never built"; 028 is "the page didn't build and said it did".
Fixing 015 does not fix this.

## What is actually wrong — two separable defects

1. **A no-op must not be `complete`.** Whatever marks the work item complete is not
   consulting the handler's own no-op result. 015 shows the `needs_human_review` path
   exists and works, so this is a routing inconsistency, not a missing capability.
2. **A page with no composed sections must not deploy** — and more seriously, it must not
   deploy *borrowed* components. The provenance of the hero and about block on that page is
   the open question: they belong to other pages of the same site. Something is falling back
   to site-level or sibling components when a page has none of its own, and that fallback is
   invisible in the output. Whatever that path is, it is capable of publishing a page that
   is confidently, fluently about the wrong subject.

## How it was found (worth repeating)

Only by reading the deployed page's `content_data` instead of its status. Every status field
— work item `complete`, page `deployed`, no error surfaced to any dashboard — said success.
The `016b` invariant *trust the rendered artefact, not the status* is what caught it, and
this case is a clean argument for that invariant.

The originating mistake was mine and is worth stating, because it is how anyone else will
hit this: I populated `pages.sections` and assumed that composed the page. The build reads
its spec sections from **`site_plan_sections`**, a different table. Setting one without the
other yields exactly this no-op. That is a genuine trap, but a trap should produce a stuck
item, not a published page.

## Reproduction

1. Create a `pages` row with `build_status='planned'` and a non-empty `pages.sections`.
2. Do **not** create matching `site_plan_sections` rows for it.
3. Queue `needs_page` for it (`handler_agent='page-build-handler'`, `status='triaged'`) and
   run `build-dispatch-loop`.
4. Observe: no-op logged in `site_work_items.error`; item `complete`; page `deployed`;
   `page_components` populated from components that are not that page's.

## Fix candidates

1. **Make the no-op terminal-but-unsuccessful.** Route it to `needs_human_review` (as 015
   demonstrably does) and leave `build_status='planned'`. Smallest change, closes the
   silent-success half.
2. **Gate deploy on composed sections.** Refuse to deploy a page whose `page_components`
   count is zero *or* whose components were not composed during this build. Closes the
   published-wrong-page half, which fix 1 alone does not.
3. **Find and name the borrowing fallback.** Whatever selects a sibling/site-level component
   when a page has none should either be removed or made explicit and logged. Until it is
   identified, fixes 1 and 2 are guards around an unknown behaviour.

Recommend 1 + 2 together; 3 as the follow-up that explains the mechanism.

## How to verify a fix

Run the reproduction. The item must **not** be `complete`, the page must **not** be
`deployed`, and `page_components` must be empty. Check the DB rows and the served URL —
not the job status, which is the thing that lied.

## Related

- `bugs_open/015` — same error string, different cause, *correct* routing. Read both.
- `bugs_open/026` — a different silent-acceptance defect on the same site: a `required`
  input field rendered empty and still saved. Same family: something that should have
  refused, didn't.
