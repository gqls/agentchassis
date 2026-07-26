# 072 — component markup ships without matching CSS on some sites (news cards bare on 2 of 5)

**Filed:** 2026-07-25, from the model_directory_pipeline session, which hit it in its own
components first and then found the same shape in the news feed's.
**Severity:** Medium — no data loss; owner-visible on live pages. Cards render as bare
unstyled markup on an otherwise designed page.
**Class:** structural (a component's markup and the CSS it depends on are produced by
different mechanisms, and nothing checks they both arrived).
**Status:** OPEN. **Cause DIAGNOSED 2026-07-26 and the fix committed — but not yet live**,
so this stays open (the `/bugs_closed/` bar is fixed AND live). See "Diagnosis" and
"What has been done" at the bottom; everything above them is the original filing and is
unchanged except where marked.

> **CORRECTED 2026-07-26:** this line used to read *"cause NOT diagnosed"*. It is now
> diagnosed. The measurement below held on re-measurement — every figure reproduced
> exactly, 24 hours later.

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
- ~~[UNKNOWN] where the 11 `news-card` rules on gaswholesalers/robot-hands/idea.uk came
  from.~~ **ANSWERED 2026-07-26:** from `css_snippets`, appended into `styles.css` by
  `render_css_from_spec`. See Diagnosis below.
- ~~[UNKNOWN] whether the two unstyled sites ever had the rules and lost them, or never
  had them.~~ **ANSWERED 2026-07-26: never had them.** Neither stylesheet has ever
  contained a component snippet of any kind.
- ~~[UNKNOWN] how many other component families this affects.~~ **SURVEYED 2026-07-26:**
  of the 94 component functions in use on active pages, **86 already carry their own
  `<style>` block** in `html_template`; the 8 that do not are `generic-text-block`,
  `latest-news`, `faq`, `news-listing`, `content-listing`, `category-listing`,
  `ported-page`, `pricing`. Only `latest-news` and `news-listing` of those have a
  `css_snippets` row, i.e. only they depend on the frozen stylesheet. Query in the
  workstream RUNBOOK.

## The workaround already applied (and why it is not the fix)

The `model-directory` / `model-directory-listing` components were given their own
inlined `<style>` block on 2026-07-25 (DB config, live immediately; page re-render
needed to reach the HTML), using the site's own custom properties with literal
fallbacks — `var(--color-primary, #1e40af)` — so they match whatever palette a site
has. That follows the `hero`/`call-to-action` precedent and makes those two components
correct on every site regardless of what generates `styles.css`.

**Verified on the rendered page, 2026-07-25 ~09:35** (not just the DB row): the live
`/model-directory.html` now carries a 2,492-char style block with 18 directory rules,
27 rendered `.model-card` articles, 50 citation links. The directory INSTANCE of this
bug is closed; the bug stays open for the class — the two unstyled news-card sites and
whatever the survey hasn't covered.

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
  > **NOTE 2026-07-26:** that workstream is parked behind an owner gate (last commit
  > 2026-07-20, "everything downstream parked, on purpose"), so nobody was going to pick
  > this up there. Taken by the `bugfix_072_component_css` workstream instead.

---

# Diagnosis (2026-07-26)

**A site's `assets/css/styles.css` is a whole-file artefact written ONLY by a
webdesign-agent design run, and nothing ever re-renders it. It is frozen at the moment
of that run, while the site's component set keeps changing underneath it.**

`RenderCSSFromSpecAction` (`platform/orchestration/actions/render_css_from_spec_action.go:76`)
renders the layout template, then appends the `css_snippets` whose `applies_to` overlaps
the site's component list **at that instant** (`loadComponentCSSSnippets`, `:586`), then
the `--section-*` defaults, then the token aliases. Page rerender never touches the
stylesheet. So a component added after the last design run has markup on the page and its
CSS written nowhere.

Evidence, all measured:

- `css_snippets` holds exactly two rows carrying `.news-card`/`.news-list-item`:
  `Latest News Grid` (`applies_to ["latest-news"]`) and `News Listing Page`
  (`["news-listing"]`). Both `applies_to` values are correctly populated — this
  **refutes** the natural hypothesis that an empty `applies_to` was the cause.
- Both unstyled stylesheets contain **zero** component snippets of any kind — no
  `fade-in-up`, no `responsive-grid`, and no `/* === Component-specific styles === */`
  block at all. Both styled ones carry that block with every matching snippet. So this
  was never about the news snippet specifically; those two stylesheets never received
  any component CSS.
- Last commit touching each stylesheet (`~/projects/sites`): ai-agent-orchestration
  **2026-05-02** (`52242272`), relojistas **2026-07-16** (`593fbec6`), gaswholesalers
  2026-05-18 (`382c8096`), robot-hands 2026-07-20 (`6f316b90`).
- ai-agent-orchestration's `index.html` first carried `news-card` on **2026-07-21**
  (`e4c4a895`) — **80 days after** its stylesheet was written. relojistas gained
  `latest-news` on its homepage **2026-07-26**; the 2026-07-25 checkout of its
  `index.html` has no `news-card`. In both cases the stylesheet predates its own markup.
- The zero-snippet shape matches the pre-2026-05-16 code path: `loadPagesWithComponents`
  returned empty (wrong `pages.status` filter), so `all_component_functions` was an empty
  **non-nil** array; `extractCSSComponents` (`:493`) falls back only when the value is
  `nil`; `loadComponentCSSSnippets` early-returns `""` on a zero-length list. Recorded at
  `design_actions.go:340-345` and in
  `docs/agent_docs/docs024_key_docs_latest/js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md`.

**[INFERRED, not measured]** relojistas' own immediate cause. Its `orchestration_states`
rows for the 2026-07-16 run are pruned, so whether that run hit the empty-list path or
simply had no matching component yet cannot be checked. The class-level cause does not
depend on which.

**Why the obvious repair is wrong:** re-running webdesign-agent to regenerate the
stylesheet re-rolls the site palette (the `generic_theme` colour-churn problem). Fixing
CSS that way would restyle two live customer sites.

## What has been done (commit `7821ad7f5`, 2026-07-26)

1. **`collectComponentCSS` / `injectComponentCSS`** (`rerender_single_page_action.go`) —
   the matching `css_snippets` are now collected at **page assembly** time and injected
   before `</head>`, so whatever assembles the page also styles it and the two cannot
   drift apart. Wired into **both** assembly paths: `assemblePage` and the bulk
   `rerenderSinglePage` (`rerender_pages_actions.go`), because otherwise a bulk rerender
   would strip the CSS back off a page the single-page path had just styled. It skips any
   component whose stored `rendered_html` already carries its own `<style>`, per component
   function, so the component-owned pattern and this injection never both ship the rules.
2. **Empty-list hardening** (`render_css_from_spec_action.go`) — an empty component list
   now resolves from the DB via `loadSiteComponentFunctionsForJS` (the helper the JS
   sibling action already calls for exactly this case) and warns, instead of silently
   writing a stylesheet with no component CSS.
3. **Migration `222_news_components_carry_their_own_css.sql` — APPLIED and recorded.**
   `latest-news` and `news-listing` now carry their own `<style>`, copied from the
   `css_snippets` row at apply time so the three already-styled sites are byte-identical
   and see **no visual change**. This brings them in line with the 86-of-94 house rule.

**Council:** submission `75d1a2af-afb8-492d-9587-4aa13bc440a2`.

## What is still outstanding — why this is still OPEN

- **The Go half is inert until the next image roll.** Nothing about the two failing sites
  has changed yet.
- **Migration 222 is live in the DB but invisible on every site**, because
  `rerender_single_page` concatenates stored `page_components.rendered_html` and does not
  re-render templates. "The component was updated" is not evidence of anything.
- **No page was re-rendered and no live site was touched** — the owner's call was to let
  the fix reach the sites on the next roll rather than hand-repair two live sites.

### Verify after the roll

First prove the code is live — pod-grep a string the change **created**:

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "data-component-css"'
```

Then re-render `index` on ai-agent-orchestration.com and relojistas.com, and measure the
rendered pages:

```bash
for d in ai-agent-orchestration.com relojistas.com gaswholesalers.com robot-hands.com; do
  html=$(curl -s "https://$d/"); css=$(curl -s "https://$d/assets/css/styles.css")
  printf '%-28s uses=%s css_rules=%s inline=%s\n' "$d" \
    "$(printf '%s' "$html" | grep -c 'class="news-card"')" \
    "$(printf '%s' "$css"  | grep -c 'news-card')" \
    "$(printf '%s' "$html" | grep -c 'news-card {')"
done
```

Pass = every row with `uses>0` has `css_rules>0` **or** `inline>0`, **and** gaswholesalers
and robot-hands are unchanged. That control is what proves this added styling rather than
restyling live customer sites — without it a green result is not distinguishable from
having overwritten three sites' news design.
