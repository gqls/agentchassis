# NOTES — webdesign.co.uk

Append-only, newest at the bottom. Evidence, commands, what the system actually
said — and every misstep, including claims in this file that later turned out
false.

---

## 2026-07-25 — session 1, orientation

### The two source sites, measured not assumed

`website-design.com`: 86 HTML pages, 55 tool dirs, 23 articles, one shared
`assets/css/global.css` (170 lines), one shared `assets/js/search.js` (63 lines)
plus `search.json` (65 entries). **Zero external runtime resources** — `grep` for
`<script src="https` across all 86 pages returns exactly one hit, and that is
escaped example code inside `learn/security/cdn-risks.html`.

`websitedesign.com`: 21 pages, 10 tool dirs, 10 guides, `css/main.css` (111 lines).
Four external deps: Google Fonts (`@import` in main.css), unpkg Lucide, Unsplash
hero photo, `esm.run/@mlc-ai/web-llm`.

`md5sum` across both trees: **no byte-identical files**. The sites share nothing —
no CSS, no JS, no images, no symlinks.

### Things that are broken in the sources (found before porting, not after)

- **`websitedesign.com/tools/vibe-equalizer/` is dead as shipped.** Its
  `index.html` does `<script src="../../js/state.js">` and
  `websitedesign.com/js/` does not exist, so `StateManager` is undefined and
  `loadState()` throws on load.
- **`guides/hosting-economics.html` is a byte-identical copy of
  `guides/local-ai.html`** (confirmed by `diff` — same `<title>`, same `<h1>`
  "Understanding the Local AI Builder"). The hosting-economics content was never
  written. Both are dropped with the LLM tool anyway.
- **`websitedesign.com` is half-migrated.** The warm palette exists on
  `index.html` only; all 20 sub-pages carry the legacy dark-terminal skin. Count
  proof: `#333` appears 59×, `#111` 25×, `#0f0` 17× — against `#5c6b5d` **1×**.
- The local-builder card's inline styles on `index.html` contain HTML entities
  (`var(&#45;&#45;primary)`) that never resolve — already broken before we touched
  anything.
- **`website-design.com`'s dark mode is declared but defeated.** `global.css` has
  a `prefers-color-scheme` block, but 71 of 86 pages hardcode `background:#fff`
  in their inline header style, plus 183× `#fff`, 142× `#666`, 99× `#555` in
  inline styles that ignore the tokens.
- Orphans: 4 finished tool dirs are in `search.json` but absent from
  `tools/index.html` (index links 51 of 55); 10 articles absent from
  `learn/index.html` (index links 13 of 23).

### Duplicate-tool check (asked for explicitly, 2026-07-25)

All 64 slugs unique. Read the actual functionality of every near-miss pair before
concluding — titles alone would have been misleading in both directions:

- `seo-schema` offers Article/Product/FAQ JSON-LD (`generator.js` has
  `forms = {article, product, faq}`); `seo-injector` builds LocalBusiness JSON-LD
  wrapped in an AI-builder prompt (its inputs: Business Name, Type, Website URL,
  Image/Logo URL, Phone). Different schemas, different output mode. **Both stay.**
- `shadow-stacker` labels are `X:/Y:/Blur:/Spread:/Opacity:/Color:` with an
  `addLayer` function — a manual layer editor. `smooth-shadow` labels are
  `Alpha (Opacity)/Distance/Sharpness` driving 1–7 auto-eased layers — parametric.
  **Both stay.** This was the closest call.
- `text-sanitizer` cleans AI text after the fact; `insight-injector` constrains
  generation up front ("Force the AI to build copy around hard facts, and
  permanently ban generic LLM vocabulary"). **Both stay.**
- `css-variables` = "Define your design tokens once… 1.5 ratio scale";
  `vibe-equalizer` = mood sliders → live preview. **Both stay.**
- Five prompt tools, five distinct jobs (image prompts / prompt trees /
  permutation matrix / sequential site-build deck / refactor prompts). **All stay.**

### Platform facts verified this session

- **`webdesign.co.uk` had no DB presence at all** — no `sites` row, no
  `site_specs`, no `pages`, no `site_work_items`. Confirmed by
  `SELECT * FROM sites WHERE domain ILIKE '%webdesign%'` → 0 rows.
  Worth flagging because everywhere else in this repo "webdesign" means the
  **`webdesign-agent`** (the CSS renderer), not a site. The palette-churn landmine
  in memory is about that agent, and it damaged **robot-hands.com**, not a site
  of this name.
- **The domain is already live-wired.** `curl https://webdesign.co.uk/` returns
  the B2 worker's `NoSuchKey` JSON from a CF edge IP — registration, DNS, zone
  and worker route all exist; only bucket content is missing.
- **`create_work_item` defaults to `status='triaged'`**
  (`create_work_item_action.go:141`) — immediately dispatchable. This is why an
  airlock is needed at all.
- **The site lock is NOT a park.** `build-pipeline-trigger.pre_query` counts only
  `locked_at IS NULL` sites, so the lock decides whether the trigger *fires*. But
  the trigger's `find_dispatchable_site` step runs
  `SELECT DISTINCT ON (wi.site_id) … WHERE wi.status IN ('triaged','approved') …
  ORDER BY wi.site_id, wi.priority ASC LIMIT 1` with **no `locked_at` filter** —
  so a trigger firing for another site's backlog can select ours. Read both SQL
  texts to establish this; the `pre_query` alone would have given the wrong answer.
  `LoadWorkItemsAction` (the dispatch loader, line 511) has no lock check either;
  the `locked_at` check at line 126 belongs to `WriteBuildItemsAction`, a
  different action that happens to live in the same file.
- **`build-pipeline-trigger` interval = 120s, `max_concurrent` = 8.** Watcher at
  5s therefore has a comfortable margin, but the window is not zero.
- **Dispatch gate** (`load_work_item_actions.go:559-568`):
  `status IN ('triaged','approved') AND attempt_count < max_attempts AND
  (approval_mode='auto' OR status='approved') AND (depends_on IS NULL OR all deps
  complete/verified)`. So `approval_mode='manual'` is a second, independent park
  if ever needed.
- **`blocked` is a respected status**: it has its own index
  (`idx_work_items_blocked`) and `CompleteWorkItemAction` explicitly refuses to
  overwrite it (line 808).

> **CORRECTED 2026-07-25 — a claim in the approved plan was wrong.** The plan said
> a `blocked` item is *"inert and excluded from the `idx_swi_dedup` live set"*.
> Half right. It is inert (the dispatch gate only takes triaged/approved), but it
> is **not** excluded from the dedup index. Read the index definition:
> `idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND status
> <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved',
> 'cancelled'])` — `blocked` is not in that exclusion list, so a parked item
> **still holds its dedup slot**. That is benign here (it stops a duplicate item
> appearing behind a parked one) but it means a forgotten park blocks re-emission
> of that item_key. Caught by reading `\d site_work_items` rather than trusting
> the plan text. Cheap check that would have caught it earlier: read the index
> definition instead of paraphrasing it.

### Latest seed number

`docs/agent_docs/sql_for_agents/` currently tops out at **207**
(`207_experience_gauntlet_api_liveness_evidence.sql`, another thread's, landed
today). Note two `206_*` files already collide by number. Re-check at commit time
— seed numbers are taken by whoever commits first.
