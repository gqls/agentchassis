# PLAN — provocation-card

**Tool/component function:** `provocation-card`
**Site(s) using it:** vonc.com (Spark) — index page, ordering 1
**Created:** 2026-06-29 (retroactively, from session history)
**Status:** Live and working via Path-2 loader; Path-1 migration pending.

---

## Aim
The daily hero provocation card — the centrepiece of the Spark v1 index ("the
product IS the landing page"). Shows a single contested claim, the AI's take, a
few engagement stats, a small set of "today's other provocations", and the two
primary actions (file a position / see all provocations). It must read as the
day's live provocation, refreshed daily.

## Source spec (Spark v1 roadmap)
- index purpose: "The product IS the landing page. Single provocation card
  filling screen. Text input beneath. Timer. No signup wall."
- index section_types include `provocation-card`.
- v1 feature set: `daily_provocation_generation_from_scraping`,
  `ai_take_per_provocation`, `shareable_provocation_cards`,
  `daily_static_regeneration`.
- AI role: producer not performer — AI frames/scores/curates; the take is the
  user's. The card presents the AI's framing + take, not AI "humour".

## Behaviour contract
Displays **today's** provocation. Content is **daily-regenerated** (not static):
- `eyebrow` — small label ("Today's Provocation")
- `headline` — the contested claim (may contain `<em>` emphasis)
- `body` — the AI's take / framing
- `primary_cta` / `secondary_cta` — label + url
- `stats` — 3 × {value,label} (e.g. positions filed, time to close, % disagree)
- mini-lobby — up to 4 × {icon,title,desc,url} ("today's other provocations")

## Delivery mechanism
- **Currently LIVE: Path 2** — a library `js_snippet` (`provocation-card-loader`)
  bundled into `/assets/js/snippets.js` by `render_js_snippets_for_site`,
  committed by `site-asset-renderer`. The loader `fetch()`es
  `/data/provocations.json` and fills the shell. Proven working 2026-06-29.
- **Durable target: Path 1** — move the fetch-and-fill logic into the component's
  own inline `<script>` (→ extracted to `content_components.js_content` →
  `/tools/assets/provocation-card.js`, auto-deployed on page rerender), matching
  how `latest-news` delivers its fetch. **Blocked** on the extraction bug (see
  NOTES) — provocation-card's inline `<script>` is not currently extracted to
  `js_content`.

## Data contract — `/data/provocations.json` (what Phase 3 must emit)
```json
{
  "generated_at": "ISO8601",
  "today": {
    "eyebrow": "...", "headline": "... <em>...</em> ...", "body": "...",
    "primary_cta": {"label":"...","url":"..."},
    "secondary_cta": {"label":"...","url":"..."},
    "stats": [{"value":"...","label":"..."}, ... x3]
  },
  "lobby": [{"icon":"...","title":"...","desc":"...","url":"..."}, ... x4]
}
```

## DOM contract (selectors the loader fills)
`.pc-eyebrow`, `.pc-headline#pc-headline` (innerHTML, for `<em>`), `.pc-body`,
`.pc-btn-primary` (href + label, inline SVG preserved), `.pc-btn-secondary`,
`.pc-stat-value` ×3, `.pc-stat-label` ×3, `.pc-card` ×4 (each `.pc-card-icon`,
`.pc-card-title`, `.pc-card-desc`, + click/keyboard nav from url).

## Dependencies
- `/data/provocations.json` — produced by the Phase 3 provocation pipeline
  (not built yet; interim file hand-committed for the proof).
- The loader — `js_snippet` now (Path 2); component `js_content` later (Path 1).
- `site-asset-renderer` — needed to deploy the Path-2 bundle (and currently
  needs manual triggering — see the js-bundle-stale issue).

## Deliberate decisions (do NOT "fix" these)
- **JS-required for the daily content, by design.** The card is a static shell
  filled client-side from a daily JSON. This is the intended v1 mechanism
  (doc 022 Tier 1: API-fetched data rendered client-side on static HTML), not a
  bug. Do not "fix" it by baking content at build time.
- The mini-lobby inside this card **will be trimmed** — decision CONFIRMED 2026-07-04.
  `lobby-grid` is the index's arena grid (live and browser-verified since 2026-07-04);
  this card's role is the single headline provocation + stats + CTAs. The overlap is
  currently live on the index. See "Trim plan" below.

## Known limitation
The underlying html_template is Mode-B broken (bare `<no value>`, names lost,
empty schema); the loader masks it by targeting selectors. If JS fails, the user
sees `<no value>`. A cleaner regeneration (sensible pre-JS defaults instead of
`<no value>`) is a logged refinement, not urgent.


---

## Trim plan (NEXT TASK, drafted 2026-07-09 — see HANDOFF_2026-07-09 §4)

**Three artefacts change, in this order:**
1. `content_components.html_template` (6163ff14-9f94-4962-aa19-d2718eabdeb1): remove the
   `div.pc-card-grid` block (four `article.pc-card` children, all loader-filled).
2. `js_snippets.js_content` where `name = 'provocation-card-loader'`: remove the
   `if (Array.isArray(data.lobby)) { … }` fill block. Keep eyebrow/headline/body/CTAs/stats.
3. `page_components.rendered_html` for this component on page b4d24f8e-fccd-49df-9dad-aa56a0b20a68 —
   an index rerender is ASSEMBLE-ONLY and would otherwise redeploy the old markup.

**Two deliberate consequences (handle in the same edit, do not discover later):**
- `.provocation-card-section .pc-container` is `grid-template-columns: 1fr 1fr` at `min-width: 900px`
  because the card grid was the second column. Change that rule to `1fr` or the section renders one
  column of content beside an empty one.
- The template's inline hover script (`.pc-card` activation) becomes a no-op over an empty NodeList.
  Harmless but dead: remove it in the same edit, or leave and note. (Extraction to `js_content` +
  a `script src` reference — the pattern already live on gauntlet-interface / latest-news /
  tool-archetype-taster-quiz — is the tidier long-term move; backlog.)

**Method.** Dump template + instance, edit OFFLINE, `UPDATE` with full new text, verify by length
delta (expect roughly −1,200 to −1,600 bytes). Do not attempt a multi-line SQL REPLACE of nested
markup. Dated backups first (`_vonc_*_backup_20260709`). Keep `data-runtime-fill` intact.

**Redeploy = two separate deploys.** Loader → `083_trigger-asset-renderer-vonc.sh` (bundle stays at
3 active snippets). Template + instance → `rerender-index-vonc.sh` (assemble-only).

**Acceptance.** Six sections still deploy; provocation-card fills; `curl … | grep -c 'pc-card-grid'`
= 0; lobby-grid's six cards unchanged; no console errors.

**Post-trim tidy.** The `lobby` key in `/data/provocations.json` becomes dead — safe to leave, remove
when Phase 3 emits the file.
