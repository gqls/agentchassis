# 252 — assembly drops per-page `og:*` and hardcodes `<html lang="en">` on 503 pages

**Filed 2026-08-11** at the owner's direction, escalating what has until now been
carried as a *stated, accepted loss* in one lane's PLAN. **Platform, fleet-wide.**
Not a regression: this is how assembly has always behaved. What changed is the
scale — the loss applied to 2 pages on 2026-08-06, 19 after LMC Track A, and
**503 assembled pages already exist fleet-wide**.

> **On the 2026-07-31 ruling: `090` substituted, and here is the why.** No part of
> this is inferred. Both behaviours are read directly from source (line numbers
> below), the affected population is a single `COUNT`, and the seam the fix should
> use is an existing, working mechanism in the same function. There is no
> not-where-you-are-looking cause for the loop to find. What *is* uncertain is a
> product decision (which locale, and whose), and no diagnosis run settles that.

---

## Two separate defects, filed together because they share a fix site

### A. Per-page `og:title` / `og:description` / `og:url` are lost on every assembled page

The shared `<head>` component is one row per site, so it cannot carry per-page
values. A hand-built page carries 5 `og:` tags; an assembled one carries the 2
generic tags the shared head holds (`og:type`, `og:site_name`). Measured on LMC:
verbatim page `og:5`, assembled page `og:2`.

**This is a solved problem in the same function.** Assembly already splices two
per-page values into the shared head:

- `rerender_single_page_action.go:618` — regex-replaces `<title>…</title>`
- `rerender_single_page_action.go:625` → `spliceMetaDescription` (:1017) — fills
  a **blank placeholder** `<meta name="description" content="">` left in the
  site-level head, and *removes* the placeholder when the page has no description.

So the pattern is established and proven: **the shared head carries a blank
placeholder; assembly fills it per page.** `og:` wants exactly this, and nothing
more novel. Confirmed the LMC head component already has the blank-description
placeholder (`t`) and no `og:title` (`f`), so the two are directly comparable.

**Why this shape is the right one, and is safe:** a site whose head component has
no `og:` placeholder gets no `og:` tags — i.e. **unchanged behaviour**. The
capability is opt-in per site, reachable by nothing until a head component names
it. Per the OWNER RULING of 2026-07-29 §1 that makes it a normal council-gate
change rather than an RFC: it adds an opt-in capability without changing what the
shared mechanism *guarantees* to anyone not using it.

### B. `<html lang="en">` is hardcoded, in four places, on a British estate

```
platform/orchestration/actions/rerender_single_page_action.go:670
platform/orchestration/actions/rerender_pages_actions.go:527
platform/orchestration/actions/multipage_actions.go:1024, :1053
platform/orchestration/actions/ai_actions.go:1250        (template)
```

Every assembled page on every site declares `en`, regardless of locale. The
hand-built LMC pages declared `en-GB`. **`sites` has no language, locale, country
or region column** — checked; nothing to read from. So this is not "wire up the
existing field", it is "decide where locale lives, then wire it".

**Scale: 503 assembled pages fleet-wide** — this is not an LMC issue, and fixing
it in one lane's chrome would be the wrong place.

Consequence is modest but real: screen readers pick voice and pronunciation from
`lang`, and `en` on a UK finance site gives an American reading of currency and
place names. It is also simply false metadata on a British estate whose own
platform conventions specify British English.

## Blast radius, measured 2026-08-11

| | value |
|---|---|
| assembled pages fleet-wide (`sections <> ["ported-page"]`) | **503** |
| LMC pages affected by A | 19 (was 2 on 08-06) |
| sites with an assembled homepage | 23 |
| `sites` columns matching lang/locale/country/region | **0** |

## Fix candidates

**For A — follow `spliceMetaDescription` exactly.**
Add `spliceOpenGraph(head, page)` beside it, filling blank placeholders
`<meta property="og:title" content="">` etc., and removing any placeholder the
page cannot fill (that removal behaviour is already the established convention —
copy it, do not invent a different one). Then add the placeholders to the head
components of sites that want them. `og:url` should use the **same** normalised
URL as the canonical, which means it inherits `bugs_open/251` — **fix 251 first
or A will faithfully reproduce the `/index.html` error into `og:url` as well.**

**For B — decide where locale lives before writing code.** Options, cheapest first:
1. A `sites.language` column defaulting to `'en'`, read at the four sites. Honest,
   small, and the default preserves today's behaviour exactly.
2. Derive from the domain TLD (`.uk` → `en-GB`). Rejected: `.com` sites in this
   estate are British too, so it guesses wrong in the direction that matters.
3. Put it in the head component and stop emitting `<html lang>` from Go. Rejected
   for the same reason `og:` cannot live there — except it *can*, because `lang`
   is genuinely per-site not per-page. **This is worth considering seriously**: it
   needs no schema change and no Go change on three of the four sites.

**Do not fix B in one lane's chrome file.** With `<html lang="en">` emitted from
Go, a chrome-level change would be overwritten; with option 3 it would not. Settle
the mechanism first.

> **OWNER DECISION 2026-08-11: option 3.** Put `lang` in the head component, stop
> emitting `<html lang="en">` from Go. No schema change; behaviour unchanged on
> three of the four affected sites (they'll get the same default via the head
> component that Go emits today); the fourth needs its head component to actually
> name a language once this lands. Mechanism now settled — B is unblocked.

## Scope note

Shared render-path Go on the `page-rerender` path → **council gate**. Name the
other consumers in the submission: `AssemblePageAction` (`multipage_actions.go`),
used by three active agent types, is a **second head producer that does not share
these code paths** — the standing landmine recorded in
`docs026_concept_register/register/seo.md` is that `injectCanonicalLink`,
`injectPageJSONLD` and `injectRobotsNoindex` all live on one producer only. A fix
to A that lands on `assemblePage` alone will leave the rebuild path emitting no
`og:` at all, which is the same divergence a third time.

## See also

- `bugs_open/251` — the canonical names `/index.html`; **A depends on it**, since
  `og:url` should agree with the canonical.
- `loanandmortgagecalculator_couk/PLAN_2026-08-05` §6 — where these were first
  recorded as accepted losses, and correctly so at 2 pages.
- `docs026_concept_register/register/seo.md` — the two-head-producer landmine.
