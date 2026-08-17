# 299 — the home page's second CTA names the Website Brief Starter and its href DIALS THE PHONE; the tel: URI is malformed too, and the section was written AFTER the 268 fleet fix

**Filed 2026-08-17** by the `webdesign_uk_build_service` lane, **owner-reported**
("this text links nowhere"). It is worse than a dead link: it goes somewhere, and
somewhere wrong.

## The defect, verbatim from the served page

`https://preview.webdesign.uk/index.html`, the call-to-action section:

```html
<a href="tel:+44 (0) 7934 524 911" class="cta-btn cta-btn-secondary">Or answer a couple of
quick questions first with the Website Brief Starter, a tool that helps you set out what you
need before we talk.</a>
```

Three separate faults in one element:

1. **The copy names a TOOL; the href dials a PHONE.** A visitor who clicks the
   sentence about answering questions gets their dialler. This is the
   `cta_names_unknown_destination` / misdirected-CTA class (the `bugs_closed/268`
   family) — copy promising X, link delivering Y.
2. **The destination exists and is correctly linked elsewhere.**
   `/tools/website-brief-starter/index.html` is live and is linked properly from
   the nav and the footer on both `index` and `contact` (measured 2026-08-17).
   So this is not a missing page; it is a wrong href on one element.
3. **The `tel:` URI is itself malformed** — `tel:+44 (0) 7934 524 911` contains
   spaces and parentheses. Even read as a phone link it is not a valid `tel:` URI
   (`tel:+447934524911`).

Whole sentence as a button label is a fourth, milder problem: a 130-character CTA
is not a button, and no button copy in the house voice looks like this.

## Why this is a PLATFORM finding, not just bad copy

**The section was written AFTER the 268 fleet fix was live.** `page_components` for
`index`: `call-to-action` `updated_at = 2026-08-16 16:12:45`; the 268 CTA-destination
fleet fix shipped in `v1.0.1298+` and the sibling lane recorded it live before that.
So a chassis carrying the fix produced this element. Either the guard does not
recognise this shape (copy naming a TOOL, href a `tel:`), or it does not run on this
path. **That is the question the fixing thread must answer first** — the copy can be
corrected in one UPDATE and will be regenerated wrong again by the next rebuild if
the producer is not fixed.

Note the fleet already carries `cta_names_unknown_destination` rows for this site on
other pages (`what-you-get`, `how-it-works`) sitting in `needs_human_review`, and a
`misdirected_cta:index-rejected-v1-20260806` item that is `failed`. **No open item
covers this element**, which is why it reached the owner's eye rather than a queue.

## How to verify (and the control that stops a false pass)

```bash
curl -s https://preview.webdesign.uk/index.html | grep -o '<a[^>]*>[^<]*Website Brief Starter[^<]*</a>'
```
Expect, after a fix, an href of `/tools/website-brief-starter/index.html`.
**Control:** the SAME page's nav and footer already link that tool correctly, so a
check that merely greps the page for the correct URL PASSES TODAY, while the broken
button is untouched. Assert on the anchor whose TEXT contains "Brief Starter" and a
verb ("answer", "questions"), not on the presence of the URL anywhere in the page.

## Scope note — do not fix this in isolation right now

The owner is finalising a new plan for this site (2026-08-17, other session
`webdesign live web builder project`) under which **the whole site will be rewritten**,
the chat box moves to the home page, and the positioning copy is rewritten. This CTA
sits in exactly the section that work will touch. **File, do not patch**: a surgical
copy fix now is discarded by the rewrite, and the producer question above is the part
that survives it. Whoever does the rewrite must confirm the regenerated CTA points at
the tool page, using the control above.
