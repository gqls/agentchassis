# 072 — component markup ships without matching CSS on some sites (news cards bare on 2 of 5)

**Filed:** 2026-07-25, from the model_directory_pipeline session, which hit it in its own
components first and then found the same shape in the news feed's.
**Severity:** Medium — no data loss; owner-visible on live pages. Cards render as bare
unstyled markup on an otherwise designed page.
**Class:** structural (a component's markup and the CSS it depends on are produced by
different mechanisms, and nothing checks they both arrived).
**Status:** OPEN, cause NOT diagnosed. The measurement below is solid; the explanation
is not attempted here.

---

## Symptom — measured live 2026-07-25, over HTTPS, not from stored HTML

For each site: does its homepage emit `class="news-card"`, and does anything style it —
either a rule in `/assets/css/styles.css` or an inline `<style>` block on the page?

| site | uses `.news-card` | rules in styles.css | inline block | verdict |
|---|---|---|---|---|
| ai-agent-orchestration.com | 6 | **0** | 0 | **UNSTYLED** |
| relojistas.com | 6 | **0** | 0 | **UNSTYLED** |
| gaswholesalers.com | 6 | 11 | 0 | styled |
| robot-hands.com | 6 | 11 | 0 | styled |
| idea.uk | 0 (no news on `/`) | 11 | 0 | n/a |
| vonc.com, fundamentallyai.com, dartsonline.com | 0 | 0 | 0 | n/a |

So **2 of the 5 sites that emit the markup have nothing that styles it**. This is not
fleet-wide and not unique to one site — which is the interesting part, because the same
component row produced the markup in all five cases.

Reproduce:

```bash
for d in ai-agent-orchestration.com relojistas.com gaswholesalers.com robot-hands.com; do
  html=$(curl -s "https://$d/"); css=$(curl -s "https://$d/assets/css/styles.css")
  printf '%-28s uses=%s css_rules=%s inline=%s\n' "$d" \
    "$(printf '%s' "$html" | grep -c 'class="news-card"')" \
    "$(printf '%s' "$css"  | grep -c 'news-card')" \
    "$(printf '%s' "$html" | grep -c 'news-card {')"
done
```

## What is established, and what is not

**Established (measured):**
- `content_components` has **no CSS column** — `html_template`, `js_content`, and that
  is all (`\d content_components`). A component can only ship CSS by inlining a
  `<style>` block inside its `html_template`.
- Some components do exactly that and are therefore styled everywhere: `hero` and
  `call-to-action` both have `html_template LIKE '%<style%'`. `latest-news`,
  `news-listing` (and, until this session, `model-directory` /
  `model-directory-listing`) do not.
- No `css_themes` row contains a `news-card` rule (12 active themes checked,
  `css_content LIKE '%news-card%'` false for all). So the styled sites' rules did **not**
  come from the theme.

**NOT established — do not repeat these as fact:**
- [UNKNOWN] where the 11 `news-card` rules on gaswholesalers/robot-hands/idea.uk came
  from. A per-site CSS write, a design pass, or a session doing it by hand are all
  candidates; none has been checked.
- [UNKNOWN] whether the two unstyled sites ever had the rules and lost them, or never
  had them.
- [UNKNOWN] how many other component families this affects. Only `news-card` was
  surveyed; `.model-card` was the case that started it, and the survey stopped there.

## The workaround already applied (and why it is not the fix)

The `model-directory` / `model-directory-listing` components were given their own
inlined `<style>` block on 2026-07-25 (DB config, live immediately; page re-render
needed to reach the HTML), using the site's own custom properties with literal
fallbacks — `var(--color-primary, #1e40af)` — so they match whatever palette a site
has. That follows the `hero`/`call-to-action` precedent and makes those two components
correct on every site regardless of what generates `styles.css`.

It is a workaround because it fixes two components, not the class. If the right answer
is that every self-contained component should carry its own styles, then `latest-news`
and `news-listing` want the same treatment and the fleet wants a check that a
component's markup and its styling arrive together. **That was deliberately not done in
this session**: quietly restyling a live news section across several customer sites is
not a change to make as a side effect of a different workstream's investigation. It
belongs to whoever owns the news feed components.

## Fix candidates (unranked — the cause is undiagnosed)

1. **Give `latest-news`/`news-listing` their own `<style>` block**, as done for the
   directory pair. Smallest, immediately correct on all sites, no image roll. Risk: two
   sources of truth for styling on the three sites that already have the CSS rules —
   check specificity before shipping.
2. **Find and fix whatever writes component CSS into `styles.css`**, so it covers every
   component a site actually uses. Correct fix if that mechanism exists; the first step
   is establishing that it does, which this file does not.
3. **A discovery check**: for each `page_components` row on a deployed page, does any
   class its rendered markup emits have a matching rule in the site's CSS or an inline
   block? Catches the whole class rather than instances, and is the kind of thing that
   would have caught this the day it appeared rather than three months later.

## How to verify a fix

Not "the CSS row exists" — the rendered page. Re-render the page, fetch it over HTTPS,
and confirm the cards have a matching rule (either in `styles.css` or inline). The
component change is live in the DB the moment it is written and invisible on the site
until the page's sections are re-rendered, so "I updated the component" is not evidence
of anything (`trust the rendered artefact, not the status`, 016b).

## Related

- `bugs_open/027` is **not** this. Its "unstyled" is about imagery direction for
  generated `content_hero` images, a different mechanism entirely — the word collides,
  the defect does not.
- Owning workstream for the news half: the news-feed pooling / content-feed workstream.
  Filed here rather than routed at them directly because the cause is unknown and the
  measurement is what is worth having.
