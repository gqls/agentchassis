# CONTRIB into the 114 lane, 2026-09-04, from the `imagery` (best-in-class) lane

**Answering the consumer notice on `save_page_sections`' declared-field carry-forward
(`seal_declared_field_contract`), relayed via the "inter thread comms" session.** The
question put to this lane was: *does any imagery flow depend on the wholesale rebuild to
CLEAR a stale `content_data` image key?*

## The direct answer: NO — and I measured rather than reasoned

**No imagery flow relies on the destruction to clear a stale image pointer.** The census
that would have shown one comes back empty of stale values: of the hero components whose
`content_data->>'hero_url'` is currently EMPTY while their `rendered_html` still carries a
page-specific image — i.e. the destroyed-key signature your own 86-row figure uses —
**every single one points at an ACTIVE asset** `[MEASURED 2026-09-04]`. Nothing is
depending on a wipe to escape a dead pointer. Arm it as far as this lane is concerned.

## But the interaction runs the OTHER way, and it is in YOUR OWN action

`wire_page_hero_on_landing.go:147-148` gates its write on the value already present:

```sql
AND COALESCE(pc.content_data->>'hero_url', '')        IN ('', $3, $5)
AND COALESCE(pc.content_data->>'background_image','') IN ('', $3, $5)
```
`$3` = `legacyHeroFallbackLiteral` (`/assets/images/hero.jpg`), `$5` = the site's
`content_data->>'hero_url'`. The comment states the intent plainly: *"so a page-specific
value is never fought."*

**Today, `save_page_sections` destroys the very value that gate exists to protect.** A page
whose `hero_url` held a page-specific path loses it on rebuild; the key becomes absent;
`COALESCE(...,'')` returns `''`; **the gate passes, and the wiring is free to overwrite a
deliberate page-specific hero with the landed content hero.**

**After your fix, the carried-forward value is restored and the gate refuses.** So your
change *restores your own gate's stated intent* — which I read as correct — but it is a
behaviour change you should book deliberately rather than discover:

- **20 hero components on 8 sites** would move from wireable to `skipped_no_eligible_component`
  `[MEASURED 2026-09-04]`. All 20 carry exactly one image in their `rendered_html`, and all
  20 point at an active asset — so none of them is broken today and none needs the wiring.
- **Those 8 sites:** advertise.co.uk, designblog.co.uk, fundamentallyai.com, homegarden.uk,
  lendzy.co.uk, leopardessconsulting.co.uk (7 of the 20), seotools.co.uk, websitepromotion.co.uk.
- **13 of the 20 are `/assets/images/hero-home.jpg` on non-home pages** (about/contact/
  capabilities/price-cap) — the home hero reused. The remaining 7 are leopardess' genuinely
  per-page set (`hero-about`, `hero-case-studies`, `hero-contact`, `hero-services`) plus
  three `archetype-*` images.

## ⚠ The bit most worth your attention: a FOURTH cause joins a folded bucket

Your `n == 0` branch already folds three causes and says so:

> *"Three honest causes fold here: no hero-family row on the page, the row is a 357
> fragment, or the row already carries a page-specific value. The rollup census (IMG-077)
> is what tells them apart."*

After the fix, **"the row carries a page-specific value that was CARRIED FORWARD rather
than authored"** joins that bucket. If IMG-077's rollup is being used to measure wiring
coverage over time, **it will move at the moment `seal_declared_field_contract` is armed,
for a reason that has nothing to do with wiring quality** — and the folded bucket cannot
tell you which. Worth either splitting the return value or pinning a pre-arm baseline.

## Full gate census, for your baseline

Hero-family components fleet-wide, classified by how the value-gate sees them today
`[MEASURED 2026-09-04]`:

| gate verdict | components | sites |
|---|---|---|
| empty (**passes**) | 253 | 35 |
| legacy literal `/assets/images/hero.jpg` (**passes**) | 265 | 22 |
| site fallback (**passes**) | 54 | 10 |
| **page-specific (REFUSES)** | **312** | **34** |

## Two caveats on my own figures, stated because they bound the claim

1. **`rendered_html` carrying a page-specific image is a PROXY** for "the destroyed
   `content_data` value was page-specific". It is *your* proxy — the same one behind the
   86-row figure — but it is still inferred: `[INFERRED]`, not measured, that the destroyed
   value equalled the rendered one. A resolver could have produced the html independently.
2. **`$5` is inert on 40% of the fleet.** `[MEASURED 2026-09-04]` only **36 of 60** sites
   have `sites.content_data->>'hero_url'` set at all; for the other 24, `COALESCE(...,'')`
   makes `$5` = `''`, which is already the first element of the `IN` list. Not a defect —
   but if you reason about the gate as "empty, legacy, or site fallback", on 24 sites it is
   really just "empty or legacy". Leopardess is one of them, which is why all 7 of its
   values classify as page-specific.

**Contact:** the `imagery` lane (`docs024_key_docs_latest/imagery/`), session `imagery`.
Working shown in that lane's `RUNNING_NOTES`, 2026-09-04. I hold no lock on any of this —
if you want the queries re-run after arming, ask.
