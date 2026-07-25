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

All 63 slugs unique (recorded here as 64 on first writing — corrected 2026-07-25 once the catalogue was actually built; see the wrong-calls list below). Read the actual functionality of every near-miss pair before
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

---

## 2026-07-25 — session 1, the build

### The cascade, leg by leg

Submitted fresh (`SUBMISSION_CORR=85878c43-5b8a-4a49-8256-470b798f3bea`), site row
`6b49db8e-d447-4467-8277-4f3018af9897`. Legs ran in this order, each released
through the airlock and read before the next:

| leg | agent | duration |
|---|---|---|
| `needs_domain_research` | domain-research-classifier | ~4 min |
| `needs_vertical_research` | vertical-exemplar-researcher | ~6.5 min |
| `needs_strategy` | domain-strategist | ~2 min |
| `needs_briefing` | build-briefing-agent | ~1 min |
| `needs_site_plan` | build-site-planner | released 16:2x |

**`needs_vertical_research` was a leg the plan did not know about.** It sits
between the classifier and the strategist, crawls three exemplar sites and
writes a `vertical_landscape` spec. Released it after reading its workflow —
it only writes that one aspect and then emits `needs_strategy`, so skipping it
would have meant hand-emitting the strategist's item.

**The classifier honoured the mission brief better than expected.** It wrote all
eight palette `reference_values` at exactly the owner's hexes, plus a 12-entry
`avoid` list, from prose alone. The phase-3 pin was therefore a *hardening*
rather than a correction — it added the `character`/`guidance` prose blocks that
webdesign-agent's prompt actually renders, the full typography block, spacing,
and an explicit `dark_light`.

### Wrong calls, in the order they happened

> **1. The watcher died on its own success path.** `set -o pipefail` plus a
> `grep` over an *empty* allowlist — the normal state — returned 1 and `set -e`
> killed the loop seconds after it went active. Fixed with `|| true`. Lesson:
> the failure mode of a guard script is usually its idle path, not its busy one.

> **2. "A blocked item is excluded from `idx_swi_dedup`."** Wrong, and it was in
> the approved plan. `blocked` is absent from that index's exclusion list, so a
> parked item still holds its dedup slot. Caught by reading `\d site_work_items`
> instead of paraphrasing it. Corrected in PLAN and here.

> **3. Releasing a leg races the watcher.** Append to the allowlist and flip to
> `triaged` in the same breath and the watcher's UPDATE — built from the
> allowlist it read at the *top* of that cycle — re-parks the item instantly.
> `needs_briefing` sat `blocked` with its id sitting in the allowlist, and
> neither fact explained the other. The procedure now says: append, wait one
> interval, then flip.

> **4. psql's command tag logged as a park.** `-t -A` still prints `UPDATE 0`,
> so every idle cycle printed `PARKED UPDATE 0`. Harmless in itself and
> genuinely dangerous: a real park would have been invisible in the noise.

> **5. Every DOM round-trip added a `<div>`.** `parseFragment` wrapped content in
> a div (necessary — bare fragments get relocated by HTML tree construction) but
> returned `renderChildren(body)`, which *includes* the wrapper. Three passes
> (scripts, styles, links) meant `<div><div><div>` on every page. Exit code 0
> throughout; only reading a fragment showed it.

> **6. Three tools would have shipped unstyled.** `micro-cms`,
> `blueprint-compiler` and `vibe-equalizer` keep their CSS in sibling
> `style.css` files. The transform kept `<style>` blocks and dropped every
> `<link>`, so their entire skin vanished — and the manifest looked perfect,
> because nothing else about those pages was wrong. Now inlined, which also puts
> them through the colour sweep they needed most.

> **7. Two hover states collapsed into their base colour.** `#f4f4f5` and
> `#e4e4e7` both mapped to `--surface`; so did `#333` and `#555`. The hover
> simply stopped doing anything. Found by eye on `tool-aspect-ratio`, not by any
> gate — worth remembering that "0 warnings" measures the rules you wrote, not
> the ones you should have.

> **8. I invented a statistic while removing invented statistics.** The about
> page's stat grid had two facts (0 servers, 0kb frameworks) and two
> measurements nobody has taken for this domain (100 Lighthouse, 0.1s FCP). I
> replaced the measurements with counts — and hand-typed "64 Tools" when the
> real number is 63. Counts are now substituted from the catalogue via
> `{{TOOL_COUNT}}`, so the figure cannot be typed. **Then the substitution ran
> before the rewrite that introduces the placeholder**, so a literal
> `{{TOOL_COUNT}}` reached the page. Exit code 0 both times.

> **9. The same 64 was already loose in the live database.** It came from my own
> mission brief, written before the catalogue existed, and eight specs had
> repeated it in prose the site will show — `identity.about_us`,
> `strategy.value_proposition`, the briefing the planner reads. Corrected by
> supersede+merge before releasing the planner (`SQL_p4_fix_tool_count.sql`), so
> the home page never opens by advertising a tool that does not exist.
> **Root cause both times: a count typed by hand rather than derived.**

> **10. The data-modifying-CTE ordering trap.** `SQL_p4` first put the supersede
> `UPDATE` in a CTE the `INSERT` did not reference. All CTEs see one snapshot and
> an unreferenced one has no ordering guarantee, so the insert hit
> `idx_site_specs_current` with the old rows still current. The fix is to make
> the `INSERT` select **from the `UPDATE`'s own `RETURNING`** — which is exactly
> what `robot_hands/SQL_2026-07-17_r1b` does, and now it is clear *why* it does.

### Traps for the next person

- **`content_components.name` is NOT NULL with no default.** Every seed example
  in the docs omits it; the insert fails on it. (Seed 208.)
- **`build_status='planned'` on an owned page is a mistake.** `write_build_items`
  sweeps planned pages into the generic pipeline — the one place an owned page
  must never go. Import writes `'deployed'`.
- The dispatch queue was fast today: legs started within ~1–5 minutes of release,
  not the ~29 minutes the runbook warns about. Do not plan around either figure.

### Numbers as they actually are

63 tools, 31 learn pages, 1 about = 95 catalogued; plus 2 generated indexes = 97
pages; 28 assets; `search.json` 95 entries; `sitemap.xml` 98 URLs.

---

## 2026-07-25 — session 1, end state (BLOCKED on dispatch)

### Live and verified

- **Homepage responds 200** at `https://webdesign.co.uk/`.
- **Static assets serving from B2**: `styles.css`, `port-compat.css`, `search.json`,
  `sitemap.xml`, `robots.txt`, `404.html`, and every tool's sibling JS
  (`/tools/bayesian-rank/bayes.js` → 200), plus the header's own JS at
  `/tools/assets/webdesign-couk-header.js` → 200. 43 files.
- **The design pin held end to end.** The committed `assets/css/styles.css`
  contains `#f9f8f6`, `#5c6b5d`, `#d4a373`, `#2b2b2b`, `#edece9`, `#717171`,
  both font families, **zero** `prefers-color-scheme` blocks and **zero**
  occurrences of the old `#0055ff`. The four colour copies agree; layout is
  `tool-portal-light`, scheme `light`.
- **Chrome rendered**: header 3900 chars with Tools/Learn/About, the search pill,
  its `<script src>`, and zero empty hrefs; footer 1196; head 570.
- **97 pages imported** as owned rows with their component instances.
- **Planner honoured the scoping**: one page only.

### BLOCKED: the 98 page assemblies will not dispatch

`rerender-pages` ran (twice, by direct kcat) and did its job: it rendered chrome
and created **98 `page_rerender` work items**. Those items are all `status='triaged'`,
priority 5, and **nothing claims them**.

Evidence that this is the dispatch layer and not our items:
- No `claimed` rows on the site at all, site `locked_at IS NULL`.
- Only 4 claimed rows fleet-wide, so `max_concurrent=8` is not saturated.
- Earlier items on this same site *did* dispatch and complete (`needs_page`,
  4 of 12 `needs_imagery`), so the site was dispatchable minutes earlier.
- A **direct** `page-rerender` kcat dispatch for `tools-index`
  (page_id `6fd84cb9-7f8d-455d-ac29-350f24b17cf0`) produced **no
  `orchestration_states` row at all** — the spawn was dropped, which is the
  `bugs_open/003` signature, not a queue delay.
- The chassis pod is 116 minutes old, so the "no dispatch within ~300s of a pod
  restart" rule does not explain it.

Note the asymmetry worth chasing: `rerender-pages` dispatched fine by direct
kcat, twice. `page-rerender` did not. Same envelope shape, same topic.

**The work items are correct and are left in place.** When dispatch recovers they
assemble with no further action. Nothing needs re-importing.

### Also open

- **The homepage was assembled BEFORE chrome existed**, so the published
  `index.html` carries the default head, no header and no footer, and does not
  link `port-compat.css`. Its re-assembly is one of the 98 queued items.
- **The hero is wrong for the brief.** The planner chose a full-bleed `hero`
  component that paints `linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6))` over
  a background image with `--hero-ink: #fff`. The brief asked for a two-column
  hero (copy left, image right) and the `design_intent.avoid` list forbids dark
  backgrounds. This is component SELECTION, not the palette — the pin is doing
  its job; the chosen component simply is not the one the brief describes.
  Needs either a different section component or a recompose of that one section.
- 8 of 12 `needs_imagery` items still queued.
