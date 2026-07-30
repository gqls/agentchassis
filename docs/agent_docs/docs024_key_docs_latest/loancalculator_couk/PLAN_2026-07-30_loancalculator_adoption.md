# PLAN — adopt loancalculator.co.uk through the framework, mending the adoption path

Created 2026-07-30. Owner brief, in two parts:
1. "please can you adopt this site into the framework so we can manage it properly.
   Please try to make sure all the tools will work and are still on the same url
   paths if possible, if not the add some redirects."
2. "I'd like to be able to adopt this type of site through the framework, so
   anything you find that the framework adoption workflow would have failed with,
   please report and or mend it in the framework."

Part 2 is the load-bearing one: the deliverable is a mended adoption path, proven
by this site — not a bespoke port of this site.

## The site

28 HTML files in `~/projects/sites/loancalculator.co.uk` (the shared deploy repo
`gqls/sites`, one top-level dir per domain, branch `master`): `index.html`,
`legal.html`, `404.html`, 12 tools at `/tools/*.html`, 13 guides at `/guides/*.html`.
Source generator (now to be retired, out of scope) is
`~/projects/domains/loancalculator.co.uk` — Go `main.go` stamping `data/*.json` into
`templates/` → `public/`.

Tools are self-contained: inline `<script>` at end of body, 17–89 lines each, no
runtime `fetch()`. Shared deps are `/assets/css/style.css` and `/assets/js/nav.js`
(+ `global.js` on index). `application-tracker.html` persists via localStorage.

## Decisions and their reasons

**D1 — Adopt through the real pipeline, not a bespoke CLI.** `cmd/webdesignport`
exists and would work, but a second bespoke porter is exactly the "two
hand-maintained things that must stay identical" drift class this platform keeps
being bitten by. The owner asked for the framework to be able to do this. So the
framework gets the capability. `webdesignport` stays as the reference for row
shapes.

**D2 — Preserve verbatim, do not recreate.** [VERIFIED]
`apply_adoption_plan_action.go:517,625,633` sets `mode:"recreate"` for every page
and routes interactive pages to `needs_tool_recreation` (`:627`) — an LLM rewrite.
Twelve working calculators must not be regenerated. Verbatim preservation is the
requirement, so `fidelity=locked` becomes real (M1) and gets a byte-exact deploy
path (M2).

**D3 — Repair the site before adopting it.** The pipeline learns the site by
crawling what it serves, so defects would be captured and frozen. Also: the nav is
injected client-side by `nav.js`, so HTML link discovery alone may under-crawl —
`sitemap.xml` is the crawler's discovery net and must be correct first.

**D4 — Restore serving first rather than build an adopt-from-files source.**
[VERIFIED live 2026-07-30 15:11Z] The site had to come up anyway; the repair was
one Cloudflare worker route. An adopt-from-files crawl source is a real gap (G1)
but a follow-up, not a blocker.

**D5 — Fix the site's own defects during Phase A, not "faithfully" preserve them.**
Four pages are chrome-less fragments and ten of 28 have no nav at all. Preserving
those verbatim would be preserving breakage. Verbatim applies to the *repaired*
site.

## The gap inventory (the report deliverable)

| # | Gap | Evidence | Disposition |
|---|-----|----------|-------------|
| G1 | Crawl needs the source live; no adopt-from-files/repo source | `091_site_adoption_agent.sql` crawl step | Report; serving restored instead (D4). Follow-up feature. |
| G2 | `--fidelity` accepted, recorded, consumed by **nothing** | [VERIFIED] `082_submit_domain_unified.sh` NOTE ~line 50 says so; `grep -rn fidelity --include=*.go platform/ internal/ pkg/` = 10 hits, **all unrelated** (vet-med parse guard, a gemini comment) | **MEND M1** |
| G3 | Interactive pages forced to `needs_tool_recreation` (LLM rewrite) | [VERIFIED] `apply_adoption_plan_action.go:627` | **MEND M1** (locked bypasses) |
| G4 | Original URLs discarded — `CanonicalisePage` synthesises `/tools/<slug>/index.html` | [VERIFIED] `apply_adoption_plan_action.go:446` → `datahelpers/page_canonical.go` | **MEND M1**; downstream is fine once `pages.url` is right ([VERIFIED] `rerender_single_page_action.go:381` `Filename = TrimPrefix(URL,"/")`) |
| G5 | `redirects` table is dead schema | [VERIFIED] 0 rows fleet-wide; no Go consumer | Report + static meta-refresh stub. Emitter = follow-up. |
| G6 | No sitemap/robots/404 machinery in the chassis | `cmd/webdesignport/generate.go` ("ours to generate") | Report; stays repo-managed for adopted sites. |
| G7 | Adoption provisions DB rows but never the zone→worker→B2 serving path | [VERIFIED] this site's outage: zone up, no worker route, requests fell to a dead legacy S3 origin | Report + owner checklist in RUNBOOK. |
| G8 | (a) JS-injected nav starves crawl link discovery ⇒ sitemap dependency; (b) crawl `limit: 30` vs 28 pages is tight [UNVERIFIED — confirm in live agent_definitions]; (c) CSS/JS assets are not crawled or DB-managed; (d) full classifier→design cascade would restyle an adopted site; (e) `sectionHasVisibleContent` <10-char floor; (f) owned pages sit outside browser verification (bugs_open/084) | mixed | Report; (d) handled by M1 skipping the classifier handoff; (e) moot under M2 (no assembly) |

## The mends (council-gate scope: `platform/`)

Both are **opt-in**: absent `fidelity=locked` and absent `deploy_mode:"verbatim"`,
behaviour is byte-identical to today. Per the owner ruling of 2026-07-29, an
additive-and-inert capability is normal-gate, not architecture-scope — but the new
seam (`deploy_mode` reserved key on ported components) **must be registered in the
concept register in the same commit that ships it**.

**M1 — `fidelity=locked` in `ApplyAdoptionPlanAction`**
(`platform/orchestration/actions/apply_adoption_plan_action.go`):
- Read `input_data.fidelity`. If `locked`, drive the page list from the crawl index
  (`buildCrawlPageIndex` — keyed by real URL, carries `rawHtml`) rather than the LLM
  plan: deterministic and complete.
- `pages.url` = crawled path **verbatim**. Name from path (`/tools/x.html` →
  `tool-x`); `page_type` from prefix; `status='active'`,
  `build_status='deployed'`, `rebuild_policy='owned'`, `sections=["ported-page"]`.
  Row shape per `cmd/webdesignport/import.go:177-236`.
- One `page_components` row per page: slot `ported-page`, `component_id` =
  **the existing** passthrough row (see RUNBOOK — reuse, never seed a second),
  `rendered_html` = crawled `rawHtml` full document, `build_status='approved'`,
  `content_data` = `{schema, deploy_mode:"verbatim", source_url, sha256, generator}`.
- Emit one `page_rerender` item per page via `insertPageRerenderItem`
  (`create_rerender_items_action.go:120`) instead of
  `needs_content_page`/`needs_tool_recreation`; skip the `needs_domain_research`
  classifier handoff so no restyle cascade runs (G8d).

**M2 — verbatim deploy bypass in `RerenderSinglePageAction`**
(`platform/orchestration/actions/rerender_single_page_action.go`):
- Before assembly ([VERIFIED] `assemblePage` at `:101`, `StripToolDocHeader` `:153`,
  `repairOutboundPageLinks` `:161`): if the page is `rebuild_policy='owned'` and its
  sole component has `content_data->>'deploy_mode'='verbatim'`, ship
  `files = {Filename: rendered_html}` unmodified.
- This is what makes preservation real (keeps `lang="en-GB"`, no `<main>` wrapper,
  no chassis-chrome dependency) *and* rebuild-safe: any later sweep re-ships
  identical bytes.

Fallback if council REJECTS: `cmd/loancalcport` reusing `import.go` row shapes,
fed from the repo files (cmd/ is outside gate scope). Recorded, not preferred.

## Phasing

- **A** — repair the site in the deploy repo; push (b2 sync + CF purge). Exit gate:
  28 URLs + assets 200 live.
- **B** — M1 + M2 + register entry; council gate; build + roll; pod-grep verify.
- **C** — `082_submit_domain_unified.sh loancalculator.co.uk --from https://loancalculator.co.uk --fidelity locked`.
- **D** — 28 `page_rerender` items drain → git-adapter → sites repo → B2.
- **E** — verification + gap report (bugs_open entries for what stays unmended).

## Out of scope (follow-ups)

Stale market-rate copy ("3.75%", "18 March 2026") — needs an owner call on managed
content; retiring the `~/projects/domains` generator; a redirects→stub deploy
emitter; chassis sitemap/robots machinery; adopt-from-files crawl source; `www`
record; meta descriptions; PLAN/acceptance criteria per tool.

---

## REVISED DIRECTION (owner, 2026-07-30 late): adopt from our own files, and let the site EVOLVE

Owner's words: *"adopt from our own files. I want the site to be completely editable
and evolve and improve like the other sites will, just as long as it starts similarly
enough with working tools."*

This **supersedes the `fidelity=locked` posture for this site** and resolves the two
high objections the council raised (adoption_guardian's NO-BYPASS / WRITE-THEN-RELAY
and mission's classifier-skip) — because the site will now go through the classifier
cascade like any other, so nothing is bypassed.

### What the owner is asking for, in platform terms

Two requirements that pull in opposite directions under the current build:

1. **Starts faithfully, tools working** — the calculators must work on day one and the
   URLs must not move.
2. **Then behaves like every other site** — editable, replanned, redesigned, improved
   by the loops.

`rebuild_policy='owned'` + `deploy_mode='verbatim'` delivers (1) by making (2)
IMPOSSIBLE: the pipeline is contractually forbidden from touching an owned page, and
`save_page_sections` refuses it. So this site must come OFF that posture.

### The blocker that makes this a real build, not a config flip

**A whole stored document cannot go through normal assembly.** `assemblePage`
concatenates chrome + section HTML; feeding it a complete `<!DOCTYPE html>…</html>`
blob produces nested `<html>`, i.e. invalid output. So "editable like other sites"
requires the content **decomposed into real sections/components** — one blob per page
is precisely what cannot evolve.

**Do NOT flip `rebuild_policy` to `generic` before decomposition exists.** That would
turn a working site into malformed pages on the next rerender. The current state
(owned/verbatim, 27/27 byte-exact, zero queued items) is SAFE and is the right place
to wait.

### The design: map onto doc 028's existing dial rather than invent a mode

- **`fidelity=locked`** — frozen for ever. BUILT (ADO-037). Stays, and is the right
  answer for a site adopted as an archive. Not what this site wants.
- **`fidelity=high`** — *seed faithfully, then evolve.* This is what the owner is
  describing, and doc 028 already reserves the word for it. TO BUILD.

`high` = source the starting content from **our own files** (exact bytes, no
browser-rendered DOM — G10), decompose it, then hand the site to the normal cascade.

### Build order

1. **File source** (replaces the crawl for our own sites). The deploy repo already
   holds exactly the bytes we publish — verified: origin bytes == repo bytes for all
   27 pages. Read via git-adapter (platform-side, reusable) rather than a local CLI,
   so any site in `gqls/sites` can be adopted this way. For a genuinely external
   site a plain no-browser HTTP GET is the equivalent; firecrawl `rawHtml` is not
   (G10).
2. **Decompose each page into sections.** Prose → ordinary content sections the
   writer/improvement loops can rewrite. **Each calculator → a
   `component_level='tool'` component with its inline JS preserved BYTE-FOR-BYTE.**
   That is the platform's first-class mechanism for interactive widgets and the only
   way a tool survives a page rebuild.
3. **⚠ The interactivity-regression guard will NOT protect these tools as written.**
   `save_page_sections_action.go:292,449` keys on `<canvas`, `game-container` and
   `tool-page`. Our calculators contain **none** of those — they are `<input>` +
   `getElementById`. So either mark the extracted widgets with a `tool-page` class
   (cheap, makes the existing guard apply) or extend the guard to recognise
   `<input>`+`<script>` widgets. **Without one of the two, the first page rebuild
   silently drops every calculator** — the exact failure class this lane has been
   documenting all day.
4. **`rebuild_policy='generic'`** + a **timed** adoption lock (the platform's existing
   faithfulness mechanism — `FOCUS_adoption_faithfulness_via_locks(2).md`,
   `locked_by='adoption'`, expiring) so the seeded state is protected briefly and
   then evolves. Not a permanent lock.
5. **Queue `needs_domain_research`** so classifier → strategist → briefing → planner →
   composition → design all run. This is what makes the site "like the others", and it
   discharges the council's two high objections.
6. **Acceptance bar (the owner's own):** *"starts similarly enough with working
   tools."* After the first full build: all 27 URLs still resolve, and every
   calculator still computes in a real browser. Regenerated prose is EXPECTED and
   acceptable; a dead calculator is not.

### Watch items carried forward

- **ADO-034** (in the register): the planner invents differently-slugged duplicate
  pages for topics already adopted (`economy-basics` vs `guide-economy-basics`). With
  27 adopted pages and the cascade now running, expect this — check for bare-name
  duplicates after the planner runs.
- URL preservation must survive the PLANNER too, not just adoption:
  `CanonicalisePage` is what the planner uses, and it produces
  `/tools/<slug>/index.html`. Adoption can seed the flat URL, and the planner may
  still propose the directory form for the same page. That is the next place URLs can
  silently move.
- The checksum gate (G10's lesson) should be built into whichever source is used —
  it is the only reason this lane caught the DOM substitution.
