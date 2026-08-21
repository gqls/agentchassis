# 347 — webdesign.co.uk's head component is a bare fragment with no `<head>` element, so 117 pages opt themselves out of head features

**Filed AND closed 2026-08-21** by the `bugfix_252_og_lang_assembly` lane, on the owner's instruction
("fix webdesign.co.uk live"). Filed straight into `bugs_closed/` because it was found, fixed and proven
live within one session — there was never an interval in which it was an open case for anyone else to
pick up. Found incidentally while fixing `bugs_closed/252`; recorded first as FINDINGS §A1 in that lane.

## The defect

`content_components` row `14cf6193-c8f0-4640-9cf1-f8b5347e6885` (`webdesign.co.uk Document Head`,
its own `function` `webdesign-couk-head` — so a `WHERE function='head'` query never sees it) was a
hand-authored **fragment**: it began `<meta charset="utf-8">` and ended without a closing tag. **No
`<head>` open tag and no `</head>` close tag.** The served page began:

```html
<!DOCTYPE html>
<html lang="en">
<meta charset="utf-8">
```

Browsers imply the element, so every page rendered and looked fine.

**Why that is not cosmetic.** Every per-page head helper anchors on `</head>` —
`injectCanonicalLink`, `injectPageJSONLD`, `injectRobotsNoindex`, `injectComponentCSS` and
`spliceOpenGraph`, all via `strings.LastIndex(head, "</head>")` — and each carries its **own private
fallback** for the missing marker. **They do not agree.** Most append, which happens to work.
`injectBrandHeadTags` returns the head **untouched**. So the site silently opted out of head features
while every helper reported success.

**Scale: 117 assembled pages — the most of any site in the fleet.**

**How it surfaced:** `bugs_closed/252` moved the document language into the head component as a gated
attribute. This component had no open tag to carry an attribute, so that fix could never reach this
site — which is what made a latent shape into a visible gap.

## The fix

Migration `docs/agent_docs/sql_for_agents/529_webdesign_couk_head_gains_a_head_element.sql`
(+ `_ROLLBACK`), applied and recorded 2026-08-21. Wraps the existing fragment in
`<head{{if .lang}} lang="{{.lang}}"{{end}}>` … `</head>` and adds the map-valued `lang` schema entry
(WRAPPED shape, `source: config.locale.lang`) — the same contract migration `507` gave the two shared
head templates. **No content removed, nothing else changed.** md5 drift guard plus two `DO`/`RAISE`
assertions, one of which pins the hand-authored contents (`port-compat.css`, `cf_analytics_token`,
`<title></title>`) so a wrap cannot quietly replace the body, and one which refuses if the template
ends up with anything other than exactly one `</head>`.

## Verified at the artefact

Chrome re-rendered for the site (`rerender-chrome`), then one assembled page rerendered
(`049b_deploy_single_page.sh`, assemble-only) and read on the wire:

```
before: <!DOCTYPE html> / <html lang="en">    / <meta charset="utf-8">     (no head element)
after:  <!DOCTYPE html> / <html lang="en-GB"> / <head lang="en-GB">        (…) </head>
```

Confirmed there is **no area-component override** for the page (`area_components` head rows = 0), so
the site component is genuinely what serves.

⚠ **Check the orchestration reached `COMPLETED` before reading the page.** My first read showed the old
bytes because the run was still `AWAITING_RESPONSES` at `deploy_page` — which looks exactly like a fix
that did not work.

⚠ **Pick an ASSEMBLED page to verify.** `webdesign.co.uk/about.html` already served a `<head>` element
before this change: it is not built by `assemblePage`. Verifying there would have shown "already fine"
and hidden the defect entirely.

## What this deliberately does NOT fix

**`injectBrandHeadTags` still skips this site wholesale**, because its guard trips on the `rel="icon"`
this template carries. Wrapping the fragment does not change that. So webdesign.co.uk still gets no
`og:image` and no derived favicon tags — **expected after this fix, not a regression.** That guard is
`bugs_open/322` item 4, and it is the general mechanism: any future per-page tag added to that block
reproduces `bugs_closed/252` exactly.

## Residual

The 117 pages gain the head element only as each re-assembles; ~1 rebuild/hour fleet-wide and this
site is not among the busiest. **Owner ruling 2026-08-21: do not force page rebuilds.** Tracked with
the rest of that residual in **`bugs_open/346`** — note that file lists `webdesign.uk` (8 pages), a
**different site**; `webdesign.co.uk` is this case.

## See also

- `bugs_closed/252_…og_tags_and_hardcodes_html_lang_en…` — the parent fix and full evidence
- `docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/FINDINGS_2026-08-21_errors_caught.md` §A1 — where this was first written up
- `docs026_concept_register/register/seo.md` **SEO-005**
- `bugs_open/322` item 4 — the untouched guard
