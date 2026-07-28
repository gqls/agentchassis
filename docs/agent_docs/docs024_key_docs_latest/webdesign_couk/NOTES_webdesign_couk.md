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

## 2026-07-25 — session 1, end state (I called this BLOCKED. It was not — see the correction at the foot of this file.)

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

---

## 2026-07-26 — session 2

> ### CORRECTED: the section above is wrong. There was no dispatch bug.
>
> Yesterday I wrote that 98 `page_rerender` items "will not dispatch", called it
> the `bugs_open/003` spawn-loss signature, and committed that claim. **Every
> part of it was wrong.** The queue was working normally the whole time.
>
> What actually happened, measured this morning:
>
> | | |
> |---|---|
> | items created | 17:12:22 |
> | **first claim** | **17:33:03 — 20m40s later** |
> | last completion | 20:40:20 |
> | total for 98 pages | **3h28m**, ~2.1 min each |
>
> The 20-minute wait is the platform's documented publish→run-start latency.
> The 3.5 hours is simply what single-flight-per-site costs for 98 pages.
> **I gave up 8 minutes before the first claim.** By morning every page was live.
>
> **The reasoning error matters more than the delay.** My evidence that
> `page-rerender` was *specifically* broken was that `needs_imagery` items on the
> same site were completing while the page items were not — apparent
> concurrency, therefore a page-rerender-specific fault. They were never
> concurrent. Every imagery item I watched finish had been **claimed at
> 17:04–17:16, before the page items existed at 17:12** — I was watching the tail
> of work already in flight. Imagery then stopped dead at 17:18:54 and did not
> resume until **20:40:38, eighteen seconds after the last page finished**.
>
> So the priority change I had made minutes earlier (`page_rerender` → 5,
> imagery → 90) had worked *perfectly*, pre-empting imagery entirely for three
> and a half hours. **I read the starvation I had deliberately caused as
> evidence of a bug.**
>
> **The cheap check that would have settled it in one query:**
> ```sql
> SELECT item_type, min(claimed_at), max(completed_at)
> FROM site_work_items WHERE site_id = '<site>' GROUP BY 1;
> ```
> Items claimed *before yours existed* say nothing about yours. And
> `attempt_count = 0` means *not yet tried*, which is indistinguishable from
> *never will be* until the documented latency window has actually elapsed.
> CLAUDE.md says this in as many words — "a missing orchestration row is almost
> always latency, not a dropped dispatch — do not retry on that evidence" — and I
> both ignored it and paid the predicted cost: a duplicate direct dispatch for
> `tools-index`. (Harmless, as it happens: the importer keys the instance on
> `(page_id, slot_name)`, so `tools-index` still has exactly one component.)
>
> Logged fleet-wide in `WRONG_CALLS.md`. Note it sits directly beneath a row from
> the *opposite* failure — waiting 62 minutes because "absence means queued".
> Same ambiguity, opposite conclusions, and in both cases the resolving evidence
> was a timestamp comparison nobody ran.

### The real operational finding

Not a bug, but worth knowing before anyone plans a big rerender:

- **A site-wide rerender of ~98 pages takes about 3.5 hours** and shows *no*
  progress at all for the first ~20 minutes. Both numbers are normal.
- **Priority is absolute across item types on a site.** Setting `page_rerender`
  to 5 starved priority-90 imagery for the full 3.5 hours. That is correct
  behaviour and useful — but if you want two kinds of work interleaved, do not
  separate their priorities.
- `rerender-pages` dispatched fine by direct kcat while `page-rerender` appeared
  not to. That asymmetry was also an artefact: `rerender-pages` was picked
  up promptly because it went in when the site's queue was near-empty; the
  page items went in behind 98 siblings.

### The hero: right palette, wrong furniture

The published home page carried a full-bleed hero painting
`linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6))` over a background image with
`--hero-ink: #fff`. Dark — against both the brief ("two-column hero, copy left,
image right") and `design_intent.avoid` ("Dark backgrounds of any kind").

**The palette pin was not at fault and could not have prevented it.** The pin
governs colour *values*, and it held perfectly — every colour in the committed
`styles.css` is a pinned one. The darkness was a literal `rgba()` baked into the
chosen component's own template, drawn from no palette at all. `design_intent.avoid`
is prose; the planner's component *selection* does not consult it.

That is a real gap worth stating plainly: **we can pin what colours a site uses,
but not which components it picks, and a component can carry its own darkness.**

Fixed with a per-site `webdesign-couk-hero` (`SQL_p6`), reproducing
websitedesign.com's own hero: 1fr/1fr grid, 4rem gap, copy left, 4:3 image right
at 12px radius with the large soft shadow, collapsing to one column and 16:9
under 900px. Forked rather than editing the shared `hero`, which six other sites
use. Its verify block fails if the template ever regains `rgba(0,0,0` or
`hero-ink`.

Also fixed while there: the old instance had `cta_text` and `secondary_cta` but
**no URLs**, so the previous template gated both buttons away and the hero shipped
with an empty action row. Both now point at pages that exist.

### The JavaScript loss, made impossible to repeat

`checkScriptParity` (`cmd/webdesignport/transform.go`) now counts scripts in the
source, subtracts the site-wide engines the chassis supplies, and **refuses** to
emit a fragment holding fewer. It fails rather than warns, because a warning is
precisely what the original failure produced.

**Proven by inducing the fault, not by watching it pass.** Re-introducing the
original defect in a scratch build produced **60 `script parity` failures** — one
per tool that would have shipped dead. A gate only ever seen passing has not
been tested; it has been observed not complaining.

The platform-wide version of this gap is filed as **`bugs_open/084`**: nothing
anywhere asserts that a published page's JavaScript works. Every check we have
tests *presence* (a row exists, a file exists, a status is `complete`); none
tests *integrity*. JS is uniquely exposed because its absence changes nothing
visible until a human clicks something.

### End state, verified live 2026-07-26

Every one of these was checked with `curl`, not inferred:

| check | result |
|---|---|
| `/`, `/tools/index.html`, `/learn/index.html`, `/about/index.html` | 200 |
| sample tool + article pages | 200 |
| tool engine JS (`/tools/bayesian-rank/bayes.js`) | 200 |
| header JS (`/tools/assets/webdesign-couk-header.js`) | 200 |
| `port-compat.css`, `search.json`, `sitemap.xml`, `robots.txt` | 200 |
| chrome on a tool page (header markup, search box, nav) | present |
| old Swiss blue `#0055ff` anywhere | 0 |
| all 12 generated images incl. `hero-home.jpg` | 200 |

### Swapping a section's COMPONENT is not something any rerender path handles

Landed the two-column hero, and hit a genuinely non-obvious platform behaviour
worth writing down.

Repointing `page_components.component_id` at a different component changes which
template the section *should* use — but **no rerender path regenerates the HTML**:

- **assemble-only** (`page_rerender` with no `reason`) republishes the stored
  `rendered_html` verbatim. It happily re-published the old dark hero.
- **`reason='section_data_resolved'`** takes the `rerender_page_sections` branch,
  which **re-resolves `query.*` data fields**. Our hero has none, so it ran,
  completed successfully, and correctly changed nothing.
- **`page-build-handler`** would re-render — but it also re-runs `plan_sections`,
  which could re-select the very component we are trying to replace.

So the rerender family assumes *data* changed, never that the *template* did.
The gap is real: there is no "this section's component changed, re-render it"
signal.

**What I did**: executed the component's `text/template` over its stored
`content_data` (with `missingkey=zero`, as the platform does) and wrote the
result to `rendered_html`, then fired an assemble-only rerender to publish.

**Why that is not the "never hand-write rendered_html" anti-pattern**: that rule
exists to stop *dynamic* data being baked into HTML instead of flowing through a
query resolver. This component has no `query.*` fields at all — it is a pure
template over static content_data — so executing it produces exactly what
`RenderComponentAction` would, byte for byte. The one-off refused to write unless
the output contained `wd-hero-inner` and contained no `rgba(0,0,0`/`hero-ink`,
and it honoured the lock predicate.

**If this recurs often it deserves a proper fix** — a rerender reason like
`component_swapped` that re-renders a section from its current template. Worth a
`bugs_open` entry if a second workstream hits it.

### The second dispatch stall — and this one WAS real, for a known reason

While publishing the hero, an item sat `triaged` for 26 minutes with **zero
items claimed fleet-wide in 15 minutes** and only 2 queued in total. That is a
different evidence pattern from yesterday's false alarm, and it had a cause:
**the chassis pod had restarted 30 minutes earlier** (another session rolled the
image). This is `bugs_open/029`'s documented post-roll signature — the dispatch
pool saturates after every chassis roll.

The distinguishing check, worth keeping:
```sql
SELECT count(*) FROM site_work_items WHERE claimed_at > now() - interval '15 minutes';
```
```bash
kubectl -n ai-persona-system get pods | grep agent-chassis   # AGE column
```
Zero claims fleet-wide **plus** a young pod is a real stall. Your own items
queued behind siblings while other work drains is not. Yesterday I had the
second and called it the first.

### Live end state, 2026-07-26 (after the hero fix)

`/` → 200, two-column hero present (`wd-hero-inner` ×3), hero image, both CTA
buttons, header chrome. **Zero dark overlay inside the hero section** (verified
by extracting the section and counting). The four remaining `rgba(0,0,0` on the
page are `box-shadow` *fallbacks* in the card grids (`var(--shadow, …)`), not
backgrounds — cosmetically neutral-grey rather than warm, which is a nit not
worth a 3.5-hour site-wide rerender.

### CORRECTION: bugs_open/084's central claim was false

Filed 084 asserting *"there is no point in the pipeline where 'this page's
JavaScript works' is asserted"*. **Wrong.** The platform has a four-tier
verification ladder whose Tier 4 drives the deployed page in **real headless
Chromium** — `internal/adapters/browserrunner/run_checks_action.go`, Playwright,
real `click`/`fill`/`select`, post-interaction DOM assertions, `console.error`
and uncaught page errors via `OnConsole`/`OnPageError`, desktop + mobile
profiles, failure screenshots. **Live in production at v1.0.1167.** Plus live
`dead_controls`, `truncated_component` and `tool_health` checks.

Verified personally rather than on trust: `Chromium.Launch` at
`run_checks_action.go:547`; the Tier-4 gate `AND cc.component_level = 'tool'` at
`check_tool_acceptance_due.go:51`; `asset_loads` as `strings.Contains(html, ch.Path)`
at `check_tool_acceptance.go:375`.

**What survives, narrowed and now accurate:**
1. Tier 4 only visits `component_level='tool'` pages with a declared criteria
   fence — so owned pages, chrome, sections and `js_snippets` get no browser run.
   That is all 97 pages of this site.
2. `asset_loads` checks a script path is *mentioned*, never that it *loads*. A
   `<script src>` pointing at a 404 passes.
3. `<button>` with no handler is descoped at Tier 2 by design and unbuilt at
   Tier 4 (`RUNBOOK_experience_loop.md` T5.1) — that class has no owner.

**The mechanism of my error, which is the part worth keeping:** I generalised
from *my* population to *the platform*. My pages genuinely have no browser
coverage; that is a **coverage boundary**, not an absent capability, and the two
demand completely different work. The first version would have sent someone to
build a headless browser tier that has been running for weeks. A true statement
about a subset, promoted to a universal, reads identically in the file and costs
far more.

**Real opportunity this surfaces for THIS site:** Tier 4's gate is one predicate.
Widening it (or letting an owned tool page carry criteria in `content_data`)
would give the ~16 canvas/clipboard tools genuine browser verification instead of
manual clicking. That is now fix candidate 3 in 084 and the single most useful
follow-up here.

### Misstep: I swept another session's work into my commit

The 084 correction was committed with `git add bugs_open` — a **directory**, not
a pathspec. It picked up two files belonging to another session: the deletion of
`bugs_open/006` (which they had just moved to `bugs_closed/`, legitimately) and
their new `bugs_open/088`. Both are intact and nothing was lost, but 751 lines of
someone else's work now sit under a commit message about JavaScript verification,
where `git log --follow` and `git bisect` will not expect them.

This is the exact hazard CLAUDE.md opens with, and the exact rule I broke: *"Never
`git add *`, `git add -A`, `git add .`"* — a bare directory is the same mistake
wearing different clothes, because in a shared tree a directory is not a unit of
work, it is a shared namespace. Forward-only, so it stands; recorded here instead.

**The habit that would have caught it:** the yellow commit-scope block prints
exactly this and I did not read it. `git status --short bugs_open/` before adding
would have shown two files I did not touch.

---

## 2026-07-27 (later session) — the news feed was never going to fire

Picked up `HANDOFF_2026-07-27_phase2_uk_authority.md`, whose step 1 is **"Wait for
the feed."** I nearly did. Checking the clock first would have burned the hour and
told me nothing, because the feed could not have fired at any tick, ever.

### What the handoff said, and where it was wrong

> *"`content-feed-refresh` runs every 6h; sources were primed with
> `next_fetch_at = now()`, next task tick was due 13:49 UTC."*

Two errors in one sentence, one harmless and one not:

- **Harmless:** the sources were *not* primed with `next_fetch_at = now()`.
  `SQL_p8` never sets the column, so all five are **NULL**. That turned out not
  to matter — both due-queries read
  `(next_fetch_at IS NULL OR next_fetch_at <= NOW())`
  (`dispatch_feed_sources_action.go:91`, `feed_actions.go:1004`), so NULL *is*
  due, and sorts first under `NULLS FIRST`. I checked this before assuming it was
  a bug; it very nearly went into this file as one.
- **Not harmless:** the tick was irrelevant. **Creating `content_sources` rows is
  not what arms a feed.** The site is enumerated from the *classification spec*,
  and ours did not qualify.

### The actual blocker

`content-feed-refresh` → agent `content-feed-trigger` → first step
`find_news_sites`, whose query is:

```sql
JOIN site_specs ss ON ss.site_id = s.id
   AND ss.aspect = 'classification' AND ss.is_current = true
   AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
```

webdesign.co.uk **has** a current classification spec (written by
`domain-research-classifier`, 2026-07-25) but it has **no `content_features` key
at all** — news was never evaluated for it. `NULL::boolean = true` is NULL, not
false, and the row is dropped. Measured before the change:

```
 ai-agent-orchestration.com | gaswholesalers.com | relojistas.com
 robot-hands.com            | vetcomparison.uk           -- 5 rows, ours absent
```

`SQL_p8`'s verify block asserted 5 sources, 1 page, 4 nav items — all true, all
passing, and **none of them the thing that decides whether the feed runs.** The
verification checked what the script wrote, not what the consumer reads.

### Fix

`SQL_p9_news_feed_classification_flag.sql` — applied. New superseding
classification version adding `content_features.news_feed`, following the shape
robot-hands.com was armed with on 2026-07-08 (`source='manual-recovery'`). The
old `data` is carried forward with `||` rather than retyped, and the verify block
runs the **trigger's own predicate verbatim** rather than restating it.

Confirmed after applying, by running `find_news_sites` unfiltered and unmodified:

```
 relojistas.com | vetcomparison.uk | webdesign.co.uk    -- 3 of 5 slots
```

Run it **unfiltered**. Filtering that query to `WHERE domain='webdesign.co.uk'`
answers a different question and hides the `LIMIT 5` below.

### The trap I had to design around

`content-feed-orchestrator` runs `seed_sources` **before** `dispatch_sources`, and
`seed_content_sources` creates one source **per vertical keyword**, named
`fmt.Sprintf("News Search: %s", keyword)` with `config.query = keyword`
(`seed_content_sources_action.go:252-292`). Left naive, arming the flag would have
silently added five *more* sources whose queries were the bare keywords —
quietly overriding the editorial queries SQL_p8 wrote a page of justification for.

The insert is `ON CONFLICT (site_id, name) DO NOTHING` against `idx_cs_site_name`.
So `vertical_keywords` is set to **exactly the five name suffixes** SQL_p8 used,
making the auto-seeder a no-op by collision. `source_types` is `["news_search"]`
only — `api_news` would have added an xAI/Grok-backed `LLM News: webdesign.co.uk`
source nobody chose. **Change one character in those keywords and the sixth
source appears.** Verified after the change: still exactly 5.

### [MEASURED] Latent starvation in find_news_sites — not filed, not yet biting

The step ends `ORDER BY s.domain LIMIT 5` with no rotation or fairness, and the
sub-workflow loop caps at `max_iterations: 5`. This change makes webdesign.co.uk
the **sixth** recommended site and, alphabetically, the **last** of the six
(`w` > `v`). Whenever five or more are due in one tick, ours is the one dropped —
deterministically, every time, in silence.

It is not biting now, and the reason is an accident worth writing down.
`UpdateSourceTimestamps` sets `next_fetch_at = NOW() + fetch_interval` at
*ingestion*, which is seconds-to-minutes **after** the trigger fired, while the
scheduler's next tick is `last_triggered_at + interval`. So a fetched site's
sources come due *just after* the following tick and miss it:

```
 content-feed-refresh   last_triggered_at = 07:49:09   (next tick 13:49:09)
 ai-agent-orchestration.com   next_fetch_at = 13:49:46   <- misses by 37 SECONDS
 gaswholesalers.com           next_fetch_at = 13:51:40
 robot-hands.com              next_fetch_at = 13:57:40
```

[INFERRED, one cycle observed — not two] the consequence is that each site is
effectively refreshed every **12h**, not the configured 6h, and that the
alternation is what keeps enough `LIMIT 5` slots free for us. Both the starvation
and the doubled interval are fleet-wide properties of a query in
`agent_definitions`, not of this site. Recorded here rather than filed as a bug
because neither has yet produced an observable failure — but if this site's feed
goes quiet for a day, look here first, not at the sources.

### Misstep in my own reading, caught before it cost anything

My first instinct on seeing `next_fetch_at IS NULL` was "the SQL forgot to prime
them — that's the bug." It was wrong, and it was the *satisfying* answer: it
matched a claim in the handoff, it was one line, and it explained the symptom.
Reading the two due-queries took ninety seconds and refuted it. Had I "fixed" it,
I would have set five timestamps, watched the next tick do nothing, and had no
idea why — with a plausible-looking fix already committed to argue against the
real cause. **The cheap check was: grep the column, read the consumer.**

### The first tick fired, dispatched correctly, and ingested nothing — 029's shape

SQL_p9 worked exactly as intended and the feed still produced zero items. Both
halves of that sentence matter, and separating them took one query.

**What worked (the fix is verified, not assumed).** At 13:49 the site was
enumerated for the first time and all five sources were dispatched with correct,
well-formed payloads — right `site_id`, right `source_id`, the editorial query
intact (`"web design visual trends typography colour"`). `next_fetch_at` moved
from NULL to 19:58 on all five, which is the dispatcher's optimistic stamp
(`dispatch_feed_sources_action.go:271`). None of that could have happened before
SQL_p9.

**What then failed, and is not ours.** All five orchestrations died at
`spawn_ingester`:

```
spawn_ingester | FAILED | Request <id> timed out after 3 retries | elapsed ~00:08:07
```

`error` was populated here, but `collected_data->>'__step_error'` was **empty** —
worth knowing, since the standing advice is to reach for `__step_error` first. On
this failure class the plain `error` column is the one carrying the message.

**The discriminating check, which is the transferable part.** My feed had been
armed that same hour, so the overwhelmingly natural reading was "SQL_p9 is
wrong". The check that settled it in one query was to look at **the same tick on
a site I had never touched**:

```
 vetcomparison.uk | failed 1 | new_items 0     <- not mine
 webdesign.co.uk  | failed 5 | new_items 0
```

Two sites, two unrelated threads, one tick, zero items. Not ours.

**Cause: `bugs_open/029` (hung spawns — resolve BY SLUG, `bugs_closed/029` is the
unrelated phantom-links case).** Its corrected diagnosis says roll-adjacent, and
that is exactly what this was:

```
agent-chassis pod startTime = 13:45:31Z
content-feed-refresh fired   = 13:49:09     <- 218 s later, inside the ~300 s window
```

That file is **heavily owned** — six council rounds, active through 07-26 — so I
contributed the occurrence into the bug file and started no competing fix.

> **CORRECTED — my earlier `[INFERRED]` note in this file about a 12-hour
> effective cadence is now MEASURED, and it holds.** I marked it inferred from one
> cycle and said so. Three days of ingestion history confirm it: items land only
> in the 07:xx and 19:xx–20:xx hours, never 01:xx or 13:xx, across four sites.
> The marker did its job — it stopped me repeating a one-cycle guess as fact, and
> the check that upgraded it was three lines of SQL.
>
> The consequence is sharper than I first wrote. **The 13:49 tick is structurally
> the quiet one**, because established sources always come due just after it. So
> the tick our newly-armed site landed in was the one carrying *only* never-fetched
> sources — which is why a single roll took out 100% of that tick's work.

**Recovery.** The five sources sat at `next_fetch_at = 19:58`, **9 minutes past**
the 19:49 tick — the same staggering trap, which would have deferred them to
01:49. Reset to NULL, guarded on `last_fetched_at IS NULL` so a success could not
have been clobbered. Due again at 19:49.

**Still true and unchanged:** the page must have items before it builds, and must
be built before chrome is re-rendered. Nothing about this failure changes that
order; it only delays it.

### 2026-07-27 — every content link on the home page was a 404, and the audit had never run

The owner clicked a tool link on the live home page and it 404'd. Measured: **10 of
13 hrefs dead**, only the three nav links alive. All 12 cards across the two
`info-card-grid` components. The site had been declared live, "98 pages, all 200",
since 07-26 — true, and irrelevant, because nobody had asked whether the pages
*link* to each other. **"All pages return 200" and "the site works" are different
claims, and I had only ever checked the first.**

**Two independent faults**, which is why no single substitution covered them:

1. **Invented slugs.** `colour-contrast-checker` and `css-layout-generator` name
   pages that do not exist (real: `smart-contrast`, `layout-generator`).
   `spacing-scale-calculator` and `typography-scale` name tools that exist in **no**
   form among the 63 built. Four category links point at category pages that were
   never built. The slugs are absent from `cmd/webdesignport`, so this is
   generation, not the port — `bugs_open/092`'s mechanism.
2. **Wrong path shape.** `/tools`, `/guides` and the category links carry no
   `/index.html`. The sites are served from an **S3-compatible bucket behind
   Cloudflare** (`x-amz-*` headers) and an object store cannot resolve directory
   indexes. Measured: `/tools/` 404, `/tools` 404,
   `/tools/smart-contrast/index.html` 200 — and `/about/` + `/about` 404 on
   relojistas, robot-hands and gaswholesalers too. **Fleet-wide and inherent.**

**Scope, checked before fixing rather than after:** page_components with dead
links = the home page only; the other 97 pages match none of the patterns;
`site_components` (header/footer/head) are clean, every href already a full
`/index.html`. Two rows.

**The fix.** `SQL_p10` corrects `content_data` **and** `rendered_html` with one
shared replacement list. Both, deliberately: `content_data` is what a future render
reads, `rendered_html` is what is served, and the standing landmine is that
assemble republishes *stored* HTML — so fixing the data alone changes nothing a
visitor sees.

**Copy was corrected too, where a card described a tool we do not have.**
Repointing a card titled "Spacing scale calculator" at a design-token tool would
have fixed the 404 and left the card lying about the destination — the same defect,
quieter. Replacement descriptions were taken from the live tools, not invented.

**A deploy detail I did not know and would have got wrong:** the site is **not**
served from the DB. It is published as files into `gqls/sites` (branch `master`,
not `main`), and a GitHub Action ships changed `<domain>/` dirs to B2. My local
clone was **394 commits behind** with no `index.html` at all — so "the repo looks
empty" meant "you have not fetched", not "the site is not in the repo". Fixed in
both places, then verified live.

**Verified against the live URL, every link, not a sample:**

```
/about/index.html 200   /index.html 200   /learn/index.html 200
/tools/index.html 200   /tools/css-variables/index.html 200
/tools/fluid-typography/index.html 200  /tools/layout-generator/index.html 200
/tools/smart-contrast/index.html 200
```

Then a full-site sweep of all **99** deployed HTML files against the artefacts on
disk: **the only unresolved internal link left anywhere is `/favicon.ico`**, on all
98 pages. No favicon exists on any fleet site; choosing a brand mark is not mine to
invent, so it is flagged, not fabricated.

### Why nothing flagged it — the answer is worse than a missed check

`phantom_internal_links`, `dead_controls` and `misdirected_cta` are enabled on
exactly one agent, `completeness-discovery-agent`, and **that agent has never run —
on any site**, across the full 13-day retention of `orchestration_states`. The only
discovery agent ever to execute is `design-discovery-agent` (8 runs), which carries
none of them. Its only recurring caller is `improvement-loop`, which the owner
confirms is off; the only scheduled task pointing at it is disabled, a one-shot,
and scoped to a different site.

So the site *was* checked — by the agent that does not look at links. Filed as
`bugs_open/116`. **A detector improved but never scheduled produces exactly the
same outcome as no detector**, which is what makes this worth a bug of its own
rather than a line in 071.

> **CORRECTION to my own approved plan.** Step 2 was "change `NormalizePagePath` so
> the platform stops treating `/tools` as equivalent to `/tools/index.html`". I
> wrote that before checking who owned the adjacent bug. `bugs_open/071` **already
> reasons about that exact function** and concludes the repair belongs at the
> writer, not the normaliser — and 071 is owned by two workstreams with 68 and 60
> commits in 14 days, while `092` is owned by `bugfix_079`, active the same day.
> `scripts/who-owns.py` answers this in 0.3s and I ran it *after* writing the plan
> rather than before.
>
> My finding is still new and still real — 071's analysis covers the flat-file
> shape (`/about.html`), where the mismatch is flagged correctly; the
> `dir/index.html` shape **inverts** it into a false match. But the change is
> theirs to make, especially as `rerender_page_sections_action.go:429` compares
> normalised paths and the strip may be load-bearing there. **Contributed to 071,
> changed no platform code.**

**The most useful thing in this entry, from 071 and confirmed here:** the repair is
an **artefact, not a property**. I fixed `content_data`, `rendered_html` and the
published file, and all three will be overwritten the next time that page is
generated, because nothing upstream changed. The home page is link-sound this
afternoon. The *site* is not link-sound.

---

## 2026-07-27 evening — two corrections, both to claims in the handoff I inherited

Session 3. Owner said he thought Cloudflare analytics was enabled ("it shows 1
visit (mine)"). Checking that led into both corrections below.

### CORRECTION 1 — the "re-rendering chrome ships a 404 News link" blocker is GONE

`HANDOFF_2026-07-27b` §1 says, of building the news page:

> only after it builds, re-render chrome to publish the News nav link. That order
> is not optional — the nav row already exists in the DB, so re-rendering chrome
> early puts a 404 in the header of all 98 pages (`bugs_open/049`'s exact shape).

**That is no longer true, and I nearly worked around a constraint that does not
exist.** The platform now drops the item by itself. Evidence, in the order I got it:

- The News nav row is real and `active` — `site_nav_items` has `News →
  /news/index.html` at position 40, and `/news/index.html` is `build_status =
  'planned'`, `deployed_at IS NULL`. So the *premise* holds.
- But **every chrome path asks for fetchable items only**:
  `render_site_components_action.go:103,104,119,194`, `section_editor_actions.go:475,476`
  and `v3_site_actions.go:968` all pass `NavFetchableOnly` to `GetNavItems`.
- `applyNavVisibility` (`nav_tables.go:152-249`) drops any item whose URL is not in
  `loadFetchablePageSet`, which excludes
  `NeverDeployedPagePredicate = deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed'`
  (`datahelpers/links.go:210`). The news page matches that predicate exactly.
- The kept set is non-empty (Home/Tools/Learn/About are all deployed), so neither
  the `deployedPages == 0` nor the "every item unfetchable" fallback fires — those
  are the two branches that would serve the UNFILTERED nav and re-arm the trap.
- **It is in the running binary**, not just the tree: chassis pod
  `agent-chassis-566bf56b78-jtjnj` runs `v1.0.1175` and
  `strings /app/agent-chassis | grep -c "dropped nav items whose target page has never been deployed"` → **1**.
  The guard arrived in `a9083d51b` / `759cb2b77` (the 049 fix).

**[UNVERIFIED — no production execution trace.]** I could not show the branch
actually firing in production: the pod started `18:00:40Z` and I looked at `18:35`,
so `--since=24h` is really 35 minutes of pod life, and no chrome render happened in
it. Zero occurrences of the drop log is therefore *uninformative*, not confirming —
the same narrow-window trap as `narrow-filter-defines-the-conclusion`. The static
chain above is complete and the binary check is real, but the first chrome render
after this note is the one that proves it. **Watch for the
`dropped nav items whose target page has never been deployed` line naming
`/news/index.html`.**

Why it matters beyond the news page: three `needs_rerender` work items are already
sitting in `site_work_items` for this site (header, footer, head, all `detected`),
so a chrome re-render is going to happen whether or not anyone sequences it. Under
the old belief that was a live hazard. It is not.

### CORRECTION 2 — W2's "American spellings" is right about the count and wrong about the method

The handoff sizes W2's starting item as *"American spellings in body copy on 23 of
98 pages"*, excluding `color`/`center`/`gray`/`behavior` as CSS tokens. The count is
close — I measure **38 prose occurrences across 22 files** — but **the stated
exclusion list is not sufficient, and a sweep built on it would break the site.**

Measured on the served files (`~/projects/sites/webdesign.co.uk`), stripping
`<script>` and `<style>` blocks and then all tags, so only real prose is counted:

```
7 visualize   6 optimized   3 optimizing   3 organize   3 organization
3 defense     3 optimizer   2 prioritize   2 visualizing 1 optimizes
1 recognizing 1 specialized 1 fiber        1 optimize    1 recognize
```

**Three traps a naive `replace()` over the HTML walks straight into:**

1. **The same letters are JavaScript identifiers.** Un-stripped, the files yield
   `resize` ×23, `minsize` ×9, `textsize` ×6, `fontsize` ×4, `filesize` ×3,
   `maxsize`, `brushsize`, `originalsize`, `optimizedsize`, `initialize`,
   `tokenizer`, `sanitizer`. These are camelCase variables and the CSS `resize`
   keyword inside the tools' inline scripts. `optimizedSize` contains `optimized`;
   rewriting it silently breaks an interactive tool, and the tool still renders, so
   nothing looks wrong.
2. **Four live slugs contain the American form** — `/tools/image-optimizer/`,
   `/tools/svg-optimizer/`, `/tools/text-sanitizer/`,
   `/learn/code/regex-visualized.html`. Rewriting an `href` re-creates the exact
   404 class this site spent yesterday fixing. **Britishise the title, never the
   URL** — which means accepting a visible mismatch between a title reading
   "Image Optimiser" and a URL reading `image-optimizer`. That is the correct
   trade, not an oversight.
3. **`meter` is not an error.** It shows up 5 times over 3 pages
   (`/tools/entropy-meter/`, `/learn/security/entropy-physics.html`,
   `/tools/index.html`). An *entropy meter* is a measuring device, and British
   English keeps "meter" for the device — "metre" is only the unit of length.
   Changing it would introduce a mistake while claiming to fix one. Dropped from
   the list.

**So the safe shape of the fix is prose-only**: edit text nodes outside
`<script>`/`<style>` and outside tag attributes, plus `<title>` and the meta
description, on all three surfaces (`page_components.rendered_html` +
`content_data`, `pages.title` + `meta_description`, and the served file). Not
written yet — see the open question below.

### One thing the sizing exercise turned up that changes the framing

There are **zero** `-ise` forms anywhere on the site and 15 pages carrying `-ize`.
So this is not "stray Americanisms in otherwise British copy" — the site is
**uniformly** American on these words, and converting it is a deliberate house-style
change to a fifth of the live pages, not a defect cleanup. Worth the owner's
sign-off rather than my assumption, especially as Oxford British English legitimately
permits `-ize`. My view: on a `.co.uk` pitched at UK buyers, `-ise` is right, because
most UK readers read `-ize` as American whatever the OED allows.

### Cloudflare analytics — the beacon is NOT live, and Route B is already built

Measured `18:20 UTC` against `https://webdesign.co.uk/`: **no beacon**, under a
plain curl, under a desktop-Chrome UA, and with a cache-busting query string.
`cf-ray` present and `cf-cache-status: DYNAMIC`, so the response really is passing
through the Cloudflare proxy and could have been transformed. `cache-control:
public, max-age=3600` carries no `no-transform`, which is the one documented
condition that blocks automatic edge injection. So **Route A (Automatic Setup) is
not injecting**, whatever the dashboard shows.

`SQL_p7` already wired **Route B** and it is live and correctly gated: the
`webdesign-couk-head` template contains the beacon behind
`{{if .cf_analytics_token}}`, and `site_components.head.content_data` has **no
token**, so the gate is closed and nothing renders. That is the intended resting
state. **One token string turns it on** — the commented UPDATE at the foot of
`SQL_p7`, then a chrome re-render, which correction 1 above has just established is
safe.

**The distinction to put to the owner:** Cloudflare shows two different things and
only one of them needed enabling. *Analytics & Logs → Traffic* is server-side, on
by default for any proxied zone, and has been counting since the site went live —
seeing "1 visit" there is not evidence Web Analytics is on. *Web Analytics* is the
beacon product, and it is the one that answers "which pages are popular", which is
what the deferred ordering-by-popularity decision is waiting for.

---

## 2026-07-27 evening (cont.) — every tool page was a dead end; fixed on both surfaces

Owner ruled on W2: **designer pages stay for now** — the site should "cover design
from several angles" and evolve toward the buying-design goal — and **the pages are
not worth a rewrite for Americanisms alone**, but *"if we can positively add value
or change the subjects to be more useful etc then please go ahead."* So the spelling
conversion is **dropped**, and this is what replaced it.

### The measurement, and the two false starts before it

`tools/` n=63, median **102** words; 30 under 100. `learn/` n=31, median **303**;
none under 200. So the tools are the thin half — but thinness is not the defect,
because a tool page's value is the tool.

The real finding is link topology, and **I got it wrong twice before I got it right**:

1. **First count said "63 tools link to a guide, 0 guides link to a tool"** — an
   artefact of counting `/learn/index.html`, which is in the **footer of every
   page**. Chrome, not content. Stripping `<nav>/<header>/<footer>` was required.
2. **Second count said 0 everywhere** — my "exclude landing pages" filter dropped any
   URL ending `index.html`, and **every tool URL is `/tools/<slug>/index.html`**. It
   excluded the entire tool namespace. Caught only because the debug sample printed
   `learn/data/cors-scraping.html` with three tool links in it, contradicting the 0.

Corrected, body-only, chrome stripped:

```
TOOL  (n=63): -> a guide  0   -> another tool  0     ALL 63 are dead ends
GUIDE (n=31): -> a tool  23   -> another guide 0
```

The asymmetry is **the opposite way round from my first answer**. The guides already
feed the tools well (23 of 31). The tools feed nothing at all. Confirmed by eye on
`tools/smart-contrast/index.html`: every `href` in it is chrome or a stylesheet.

> **The lesson is the one already in `[[narrow-filter-defines-the-conclusion]]`,
> twice in five minutes.** Both wrong answers were *confident, tidy numbers*. What
> caught both was printing a sample of the underlying rows next to the aggregate —
> the aggregate cannot contradict itself, but a sample can contradict the aggregate.
> **Never report a link-topology count without printing example rows beside it.**

### What shipped

Each tool page now carries **"The thinking behind this tool"** (the guide that
explains its subject) and **"Related tools"** (2–3 siblings). 63 pages, 252 links,
89 distinct targets.

Deliberate choices worth keeping:

- **No copy was written.** Titles and descriptions are lifted **verbatim** from the
  site's own `tools/index.html` and `learn/index.html` cards. Writing 189 fresh
  descriptions is exactly the surface where `bugs_open/043`-class invented claims
  appear, and there was no need — the site had already described itself.
- **No CSS was added.** The block uses `.grid-cards` and `.card`, the ported pages'
  own classes (`port-compat.css:118,129`), so it inherits the design pin's hover
  signature. **`.cta-grid` was deliberately NOT reused** — the guide pages use it and
  it has **no rule in either stylesheet**, so those cards were stacking in one
  column. Fixed separately (one page, `learn/data/cors-scraping.html`).
- **Inside `<section class="ported-page">`, not after it** — that section supplies
  the max-width and padding, so a block placed after `</section>` would run
  full-bleed.
- **Idempotent** — keyed on a `wd-related` sentinel class, so a re-run cannot
  double-insert.

Verification, before and after shipping: **1,370 internal hrefs across the whole
site resolve to a real file, 252 of them new — 0 broken**; 63/63 structurally
balanced (main/body/html/div counts); 819 insertions, **0 deletions**; 63/63 live;
18 sampled targets fetched live, all 200.

### Both surfaces, because one is not enough

Files: `gqls/sites` `7237eb851` + `df51bfd91`. DB: `SQL_p11_tool_related_links.sql`,
63/63 components updated. The DB half is not optional — SQL_p10's standing landmine
is that **assemble republishes STORED `rendered_html`**, so a file-only fix is one
assemble away from vanishing. It is still an **artefact fix**: a full regeneration
of a tool page rebuilds it from upstream and the block goes. The durable version is a
generation-time related-links step, which is not mine to build.

### The accident that found a fleet-wide deploy bug — `bugs_open/120` (NEW, unowned)

The first deploy **reported success and shipped nothing.** `curl` showed the file
still dated 25 July, cache-busted, with no block.

Cause: the workflow computes changed domains with `git diff --name-only HEAD~1 HEAD`.
**On a merge commit `HEAD~1` is the first parent — your own commit — so the diff
returns only the OTHER side of the merge.** My push had been rejected (another
session pushed first), I ran `git pull`, and the resulting merge made my 63 files
invisible to the deploy while the other session's work shipped fine.

**It is always the pusher who loses**, because the pusher's commit becomes parent 1.
And the run is green, so nothing signals it. Filed with an induced-fault verification
recipe. The same line is in the VM deploy workflow, so it is not B2-specific.

Fixed for this push by **rebasing** rather than merging, so the tip commit's own diff
named the domain. That is a workaround, not the fix — the fix is to diff
`${{ github.event.before }}..${{ github.sha }}`, which spans both sides of a merge
and also fixes the untested multi-commit-push case.

> **`git push` succeeding and the run being green are two facts that together still
> do not mean the site changed.** Add it to the "all pages return 200 ≠ the site
> works" family from this morning. The only check that settles it is `curl` for a
> string your change created — which is the same discipline as pod-grepping for a
> symbol, one layer out.

---

## 2026-07-27 ~20:10 — the beacon, the feed's first ingestion, and a correction to my own correction

### CORRECTION to correction 1 above — right conclusion, WRONG mechanism

I wrote that re-rendering chrome is safe because `applyNavVisibility` drops the
News item at render time. **The conclusion held — no 404 shipped — but that is not
what fired.** Observed on the real run (orchestration `d12940a9`, 19:58):
`refresh_nav_tables` ran BEFORE the render, and `populate_nav_tables_action:146`
does `DELETE FROM site_nav_items WHERE site_id = $1` and repopulates from
**deployed pages**. The news page is `planned`, so the News row was simply **not
recreated**. It is now GONE from `site_nav_items` — deleted, not deactivated:

```
About /about/index.html active 0     Home  /index.html   active 0
Tools /tools/index.html active 1     Learn /learn/index.html active 2
```

`nav_refreshed` reported `"items": 4`. So two independent mechanisms would each
have prevented the 404 and I asserted the one that did not run.

**Consequence for the news sequence, which is now BETTER than the handoff says:**
the premise *"the nav row already exists in the DB"* is **false as of 19:58**. When
the news page is built and deployed, the next nav refresh will recreate the row by
itself, because the table is derived from deployed pages. Nothing to re-arm.

> **The lesson, and it is the same one twice in a day:** a static read of the code
> told me *a* mechanism that would produce the observed outcome, and I reported it
> as *the* mechanism. Two code paths can guarantee the same result; the log says
> which one ran. I had even marked the claim `[UNVERIFIED — no production execution
> trace]`, which is what made this cheap to correct — the marker did its job.

### THE FEED INGESTED — 50 items, first time ever

`content_feed_items` for webdesign.co.uk: **50 rows, 0 duplicates, 50 distinct
`source_url`**, all five editorial sources firing 10 each. The 19:49 tick worked.

**And I was wrong at 19:56 when I said it was failing fleet-wide.** I measured zero
items across all sites in two hours and read it as a stall of the same class as
13:49. It was simply **slow** — the run processes sites one at a time and each
spawned worker spends ~6 minutes validating its workflow before doing anything.
webdesign.co.uk was 5th of 5. **"Nothing has arrived yet" and "nothing is going to
arrive" look identical while a slow job is still running**, and the discriminator I
had (`updated_at` frozen at 19:49:47) was ambiguous — the parent legitimately does
not tick while awaiting a child. What settled it was simply waiting.

Also corrected en route: I claimed the run had "no children, so the spawn never
took". Wrong — children here are **spawned pods**, not rows carrying
`parent_orchestration_id`. `agent-content-feed-orchestrator-21becfc9-qhtqn` was
running the whole time. Check `kubectl get pods` before concluding a spawn was lost.

**Content quality, by source — this is an editorial decision, not a bug:**

| source | verdict |
|---|---|
| `CSS and browsers` | **strong** — Can I Use, MDN, "new CSS features 2026". On-brief. |
| `web accessibility UK` | **strong** — WCAG 2.2, UK public-sector duty, legal obligations. Also feeds the buyer section's accessibility-duty angle. |
| `AI web design tools` | **weak** — mostly vendor landing pages ("Free AI Website Builder"), i.e. competitor marketing, not news. |
| `web design trends` | **weak** — generic "2026 trends" listicles, some graphic-design rather than web. |
| `UK web design` | **RISKY** — almost entirely *"Top Web Design Agencies in the UK 2026"* ranking listicles. |

⚠️ **`UK web design` collides with a standing RAIL.** The Phase 2 brief says
**never publish comparative rankings of named agencies** — different risk class, and
it destroys the neutrality the buying-design section depends on. Nine of its ten
items are exactly that. Curating them would either republish competitors' rankings
or read as us entering the ranking business. **Raise before the news page is built.**
`relevance_score` is NULL on all 50 — nothing has triaged them yet.

### THE BEACON: SQL_p7's Route B COULD NEVER HAVE WORKED

Owner supplied the token; SQL_p12 set it in `site_components.content_data` exactly
as SQL_p7 instructed. Work item `complete`, `render_site_components` returned
`"rendered": {"head": true, "footer": true, "header": true}, "success": true`, and
the chrome **was** genuinely rebuilt — `updated_at` 19:58:21, head 570→571 bytes.
**No beacon.**

Cause: **the chrome renderer never reads `site_components.content_data`.**
`renderCtx.ContentData` (`render_site_components_action.go:227`) is a fixed
hardcoded vocabulary off the `sites` row. The bugs_open/018 schema fill then reaches
`cf_analytics_token`, sees `source: 'static'`, and at line 622 takes the branch that
fills from `fallback` **or appends to `unresolved`**. No fallback ⇒ unresolved ⇒
`{{if .cf_analytics_token}}` shut ⇒ nothing renders. The gate worked; the supply
never existed.

**Fixed in SQL_p13** by putting the token in the schema's `fallback` — this system's
own idiom for a declared static literal (the comment at line 624 says so, citing
`nav_aria_label`). Checked, not assumed: the component is referenced by **exactly
one** `site_components` row, so no other site can inherit the token.

> **Why two verify blocks passed over a path that never worked.** SQL_p7 asserted
> the template is present and gated; SQL_p12 asserted the token is in content_data.
> **Both assert the WRITE.** Neither exercised the READ, and nothing rendered the
> component and grepped the output for the tag. Straight
> [[writes-the-field-is-not-reads-the-field]]. A verify block that only re-reads
> what you just wrote proves the UPDATE ran, nothing more.
>
> Second trap in the same incident: **`"rendered": true` meant "did not error",
> not "changed anything"** — the idempotence gate at line 483 returns
> `(true, false)` on an already-populated slot, and the caller reports that as
> rendered. `force_rerender: true` (which `nav-updater` does set) got past it, and
> the artefact STILL disagreed with the status. Trust the artefact.

**HELD: the forced chrome re-render.** Owner says a new chassis is about to deploy,
and a dispatch inside ~300s of a pod restart is silently dropped — the exact thing
that cost the 13:49 feed tick. SQL_p13 is DB config, so it is already live and
roll-independent. After the roll settles: queue a fresh `nav_drift`, then grep
**`site_components.rendered_html`** for `beacon.min.js` before looking at any page.

---

## 2026-07-27 ~20:25 — news queries retuned to the industry and real developments (SQL_p14)

Owner: *"aim the news queries at the industry and any new big developments in
design and web design."*

**The first ingestion was the experiment that told us how.** 50 items, 10 per
source, and the split was clean:

```
WORKED  "CSS new features browser support"            -> Can I Use, MDN, feature news
WORKED  "web accessibility WCAG UK regulations"       -> WCAG 2.2, UK public-sector duty
FAILED  "AI website builder design tools"             -> vendor landing pages
FAILED  "UK web design industry"                      -> 9/10 agency ranking listicles
FAILED  "web design visual trends typography colour"  -> "Top 10 Trends 2026"
```

> **The rule: a query that names a TECHNOLOGY or an AUTHORITY returns journalism;
> one that names a MARKET CATEGORY returns marketing.** "builder", "tools",
> "trends" and the exact phrase "UK web design" are the terms vendors and agencies
> buy and optimise for, so the engine hands back their own pages. The two that
> worked named CSS/browsers and WCAG/UK-regulations — things standards bodies and
> journalists write about. This is worth keeping for any future feed on any site.

**One of the three was a compliance risk, not just weak.** The standing rail is
*never publish comparative rankings of named agencies*, and `UK web design` was
returning almost nothing else — straight into the table a news page curates from.

### The new set

| keyword (= source name suffix) | query |
|---|---|
| CSS and browsers | `CSS new features browser support` **(kept)** |
| web accessibility UK | `web accessibility WCAG UK regulations` **(kept)** |
| web platform standards | `web platform standards W3C WHATWG specification` |
| design industry moves | `design agency acquisition merger industry report` |
| typography and design systems | `typeface release design system open source` |

Each new one names a thing that *publishes* — a standards body, a corporate event,
a release — rather than a product category.

**Renamed in place, not deleted and recreated**, because
`content_feed_items.source_id` is a foreign key and the 20 good items had to
survive. **The 30 items from the retired queries were deleted** — agency rankings
and vendor landing pages sitting in the exact table a news page reads from. They
are a regenerable cache, not a record; leaving them would have meant trusting a
later triage step to refuse material we had already decided against.

**Both tables moved together, and that is the load-bearing part.** `vertical_keywords`
exist to COLLIDE with the editorial source names so `seed_content_sources` no-ops.
Rename a source without its keyword and the seeder adds a **sixth** source carrying
the bare keyword as its query — silently reintroducing the generic phrasing this
change exists to remove. The verify block asserts the two sets are identical.

### My verify block failed its own first run, and that was the block working

`array_agg(expr ORDER BY 1)` — **`1` is the constant 1, not a positional
reference**, so the names never sorted while the keys did, and two identical sets
compared unequal. `ON_ERROR_STOP` + the transaction rolled the whole thing back;
re-checked afterwards and all 5 original names and all 50 items were untouched.
Fixed by repeating the expression in the ORDER BY. **Worth having: a verify block
that can fail closed is worth more than one that always passes** — this one cost a
re-run and would have caught a genuine drift identically.

### Verified after applying (independently, not from the block's own NOTICE)

5 sources; 3 carrying the new queries with `next_fetch_at IS NULL` (re-armed, so
they refetch on the next tick); 2 kept sources due 02:05; keywords match names
exactly; `recommended: true` still set, so the site stays in `find_news_sites`;
20 items remain, all from the two good sources.

**NOT verified: whether the new queries actually return journalism.** That cannot
be known until the next ingestion — and note the measured behaviour that the
6-hourly task effectively runs 12-hourly, so the realistic next chance is the
~07:49 tick. **Read the titles before building the news page.**

---

## 2026-07-28 09:10 — the beacon is in the chrome, and the news retune found a much bigger bug

### Beacon: SQL_p13's fix is CONFIRMED

Forced chrome rebuild (`nav_drift:...:cf-beacon-2`) on chassis v1.0.1180, 11h old
so well clear of the ~300s dispatch window. Stored chrome now carries, verbatim:

```html
<script type="module" src="https://static.cloudflareinsights.com/beacon.min.js"
        data-cf-beacon='{"token": "633f794e53dc4f718e91be595d7037ff"}'></script>
```

`site_components.head.updated_at` 09:10:12, beacon `t`, token `t`. So the
schema-`fallback` route works and the `content_data` route never could — settled.

**Not live yet, and that is expected.** nav-updater queued **99 `page_rerender`
items** itself (exactly 117's recipe — do not queue them yourself). A ~98-page
rerender runs ~3.5h and shows nothing for ~20 min, so the live page will not carry
the beacon for a while. `curl | grep -c cloudflareinsights` on the home page is
still 0. **The stored artefact is the thing that proves the fix; the live page
proves the rerender.**

### The retuned queries: partly better, and the reason is NOT the wording

Second ingestion (07:50 tick) fetched all five retuned sources, 53 items. Honest
verdict — the retune only partly worked:

- `design industry moves` — **2 of 10 useful.** It did surface *"Anticipated
  acquisition by Adobe Inc. of Figma, Inc."*, which is exactly the target. But the
  rest were `Design Agencies Market Research Report 2034`, `Creative Agency Market
  Trends & Forecast Data 2035`, `Merge: Trusted M&A Advisor to Buy and Sell
  Agencies` — **report vendors and M&A advisory marketing.** I replaced one market
  category with another.
- `typography and design systems` — **poor.** `28 Best Free Fonts for Modern UI
  Design in 2026`, `Favourite free (open source) software? : r/typography` (twice,
  so dedup missed it). "free"/"open source" pulled resource round-ups.
- `web platform standards` — **poor, differently.** `HTML5 specification`, `The web
  standards model — MDN`, `What is the difference between W3C and WHATWG? :
  r/javascript`. Naming authorities got me the authorities' **evergreen
  documentation**, not events.

> **My "technology/authority beats market category" rule was half right and I
> should have distrusted it sooner.** It explains why the two originals worked, but
> it cannot explain 2034-dated forecast pages. The honest reading is that ALL of
> these look like web-search results — ranked by authority and SEO, undated — and
> that is a property no amount of phrasing changes.

### Which turned out to be literally true — `bugs_open/127` (NEW, unowned, FLEET-WIDE)

**Every "news search" feed on every site is a plain WEB search.** The chain is
intact until the last hop, which is exactly why it reads as correct from the inside:

1. `FetchNewsSearchAction` (`feed_fetch_async_actions.go:158`) **forces**
   `params.StepConfig.Config["search_type"] = "news"`, and its doc comment at :119
   states the guarantee.
2. `WebSearchAction` reads it (`:62`) and puts it in the Kafka payload (`:161`).
3. The adapter unmarshals it (`adapter.go:46`) and **logs it** (`:199`).
4. Then: `provider.Search(attemptCtx, query, numResults)` (`adapter.go:371`) — and
   `SearchProvider` (`providers/provider.go:7`) is
   `Search(ctx, query string, numResults int)`. **There is no parameter for it to
   travel in.** Unmarshalled, logged, dropped.

`grep -c search_type providers/*.go` → duckduckgo **0**, scrapingbee **0**,
firecrawl **0**. And `providers.SearchOptions{SearchType, Language, Region,
TimeRange, SafeSearch, Domains}` is defined and **never constructed anywhere** — so
`TimeRange` is dead too, i.e. **no recency constraint exists on any feed**. That is
the direct explanation for market-forecast pages dated 2034–2035 arriving in a news
feed, and for `source_published_at` being NULL on all 53 rows.

> **I nearly filed the opposite claim.** My first reading was "the sources don't set
> `search_type`, so they default to web" — I had the config in front of me and it
> was absent. Reading one hop further showed the *action* forces it, which would
> have made that filing wrong in a way a reviewer could not have caught without
> opening the same file. **Grepping for the key in the CONSUMER is what settled
> it** — and the sharpest form of the standing lesson: a dead key normally looks
> like a live one, but this one looked *better* than live, because it is known,
> typed, logged, and structurally unable to arrive.

**Consequence for this site: keep SQL_p14 but stop expecting it to fix the feed.**
The retune was still worth doing on its own compliance grounds — it retired the
agency-ranking query that collided with the "never publish comparative rankings of
named agencies" rail. But **the news page cannot be built on this material yet.**
The honest options are (a) wait for 127, or (b) switch this site to `api_news`
(`FetchLLMNewsAction`, which supports `hours_lookback` and genuinely searches for
recent items) — that works TODAY with no code change, but costs LLM calls per fetch
and the spec deliberately chose `news_search` to avoid an unchosen xAI source.
**That is an owner call, not mine.**
