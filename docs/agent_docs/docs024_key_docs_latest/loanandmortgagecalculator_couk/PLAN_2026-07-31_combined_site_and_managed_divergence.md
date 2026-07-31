# PLAN — loanandmortgagecalculator.co.uk: a combined site, and divergence the framework manages

Created 2026-07-31. Owner brief, in three parts (the third arrived mid-session and
is the one that shapes the design):

1. "adopt through the framework the domain mortgagecalculator.co.uk and put it onto
   my domain loanandmortgagecalculator.co.uk"
2. "use the tools and provenance workflows … to make sure that the calculators and
   tools are fully working"
3. "change or update the guides to be different and better than the
   mortgagecalculator.co.uk guides" … and then: **"I'd like them both to evolve in
   different directions toward different target markets to avoid duplicate
   penalties managed by the framework"**

Part 3 is load-bearing. The deliverable is not a copy of a site onto a second
domain — it is **two sites with different audiences, whose divergence is recorded
in the framework as spec rather than maintained by hand.**

## Owner decisions taken this session

| # | decision | consequence |
|---|---|---|
| D1 | **Combined loan + mortgage site**, matching the domain name | 24 calculators, not 12: 12 mortgage + 12 loan |
| D2 | **Both old sites stay up, unchanged**; no redirects | divergence must be real, not cosmetic — see D3 |
| D3 | The two sites **evolve toward different target markets, managed by the framework** | distinct `identity` / `content_direction` / `audience` specs per site; the divergence is a spec the content pipeline reads, not an edit I make once |

## The market split (D3 made concrete)

- **mortgagecalculator.co.uk** — stays the narrow, single-subject mortgage
  authority. Audience: someone with one question about a mortgage.
- **loanandmortgagecalculator.co.uk** — the **whole-borrowing-picture** site.
  Audience: someone whose unsecured borrowing and their mortgage *interact* —
  car finance counting against affordability, consolidating debt into a
  remortgage, whether the next £1,000 should go on the deposit or clear a loan.

This is what makes the sites non-duplicate on the merits rather than by paraphrase:
**no single-subject site can answer a question that spans both**, and every new
guide here is about the crossing point. The calculators overlap; the questions the
site answers do not.

## Sources, and what is verified about them

| source | what | verified |
|---|---|---|
| `~/projects/domains/mortgagecalculator.co.uk/gemini/02/` | 23 HTML + css/js/images | **all 23 byte-identical to what the live site serves** (sha256, 2026-07-31). The top-level dir is NOT what is live — `gemini/02` is |
| `~/projects/sites/loancalculator.co.uk/` | 12 tools + 13 guides + assets | live, all 29 URLs 200. **Another lane owns this site — I copy its files, I do not touch its directory** |

## Provenance baseline (part 2 of the brief), before any porting

Ran `webdesign_tools_repair/toolaudit.py` — a real-browser audit against what each
tool *claims* — over every interactive page.

**First run: 14 of 14 `NO-CONTROL`. That was the harness, not the site.** Every
check scoped its queries to `main …`; this site has no `<main>` element. Fixed
(`f38f5bf7f`), and the fix measured both ways: 13 of 14 now `RESPONDS`, and the
original binary re-run over 4 `webdesign.co.uk` tools drifted **0 of 36 result
fields**, so no existing verdict moved. Full account in that lane's NOTES —
harness fault ten.

**Baseline: 13/13 mortgage calculators RESPONDS.** The 14th is the hub page, whose
"buttons" are all `<a>`-wrapped navigation — `NO-CONTROL` is correct there.

## Defects found in the sources (fix here, do not port)

Verbatim preservation applies to the *repaired* site — the same reasoning the
loancalculator lane recorded as its D5.

| # | defect | evidence |
|---|---|---|
| S1 | 6 of 9 mortgage guides link `Home` → `index.html` from inside `/guides/`, resolving to `/guides/index.html` | live **404** |
| S2 | mortgage homepage links `guides/mortgage-scorecard.html`; the file is `your-mortgage-scorecard.html` | live **404** |
| S3 | 2 mortgage guides are orphans (nothing links to them) | link graph |
| S4 | no `sitemap.xml`, and `robots.txt` still carries `# Sitemap location (replace with your actual domain)` | live 404 + the placeholder comment |
| S5 | no `favicon.ico` — every page load logs a 404 | live 404 |
| S6 | **36 CSS classes used by the loan tools are undefined in their stylesheet** (`.fca-style-warning`, `.market-context-box`, `.comparison-grid`, `.score-meter`, …) — those components render unstyled | class inventory vs `style.css` |
| S7 | mortgage pages carry a **Cloudflare Analytics beacon token belonging to the old site** | `index.html:137` et al |

## Design decisions for the build

**B1 — Static nav in every page; no JS-injected nav.** The loan tools use
`<div id="nav-placeholder"></div>` + `nav.js`. Two independent reasons not to carry
that over: it starves crawl link discovery (G8a), and it is the direct cause of the
verbatim-adoption mutation below (nav.js had run before the crawl serialised the
DOM, baking ~9KB of nav into every page). Static nav is more robust *and* makes the
page safe to preserve.

**B2 — Root-relative asset and link paths (`/assets/…`, `/mortgages/…`).** S1 is
what relative paths cost.

**B3 — One unified stylesheet covering both class vocabularies**, so S6 is fixed
rather than inherited. Base is the mortgage design system (navy + gold); the loan
components are restyled into it.

**B4 — The calculator logic is copied byte-for-byte and the build asserts it.**
Only `<head>`, nav, and internal links are rewritten. The builder extracts every
`<script>` block from source and destination and fails if any differs. This is the
guarantee behind "make sure the calculators are fully working" — it is checked
mechanically, not by inspection.

**B5 — Guides are NEW, and are the only wholly-new content.** 10 guides on the
crossing point between unsecured borrowing and a mortgage, plus a hub.

**B6 — No invented market figures.** The source sites hard-code "3.75% base rate"
and a March 2026 date, which is how copy goes stale and starts lying. The new
guides explain *mechanisms* (which do not go stale) and send the reader to the
calculator for numbers. Where a rule is structural it is stated as structural.

## The framework path, and the live defect in it

Verified live this session, **contradicting `FUTURE_adoption_source_destination_separation.md`**,
which describes source/destination separation as unbuilt. It is built:

- `agent_definitions` → `site-adoption-agent` → `ensure_site_record.config.domain_override_field = "input_data.destination_domain"`, consumed at `site_db_actions.go:131`
- `crawl_site.config.url_field = "input_data.target_url"`
- so `082_submit_domain_unified.sh <destination> --from <source>` genuinely separates the two
- `fidelity=locked` verbatim adoption is **live in the pod** — `v1.0.1211`, pod-grepped on three strings the change added plus a positive control
- `crawl_site` requests `formats: ["markdown","rawHtml"]`, `limit: 60`; exactly one active `ported-page` component (`a7daa5c5…`) — reuse it, never seed a second

**But `fidelity=locked` is NOT safe as a byte source, and this is already
diagnosed** (loancalculator lane, LANDMINES, their G10): firecrawl's `rawHtml` is
the **serialised post-JavaScript DOM**, not what the origin sent. It mutated 3 of
their guides in production — baked nav, absolutised URLs, `href="#"` → a self-link
that reloads the page, `<meta charset>` swapped, whitespace collapsed, `&` →
`&amp;` — before they caught it and cancelled the remaining 24 items. They then
proved the *mechanism* sound: with the **served bytes** loaded into the components,
a `page_rerender` redeployed **byte-identically** (empty diff across the commit).

So: **the deploy repo is the byte source, not the crawl.**

## Phasing

- **A — build** the combined tree in `~/projects/sites/loanandmortgagecalculator.co.uk`
  (assets, 24 ported calculators, hub + 2 section hubs, 10 new guides + hub,
  sitemap, robots, 404, legal, favicon). Exit gate: B4's byte-identical assertion
  passes for all 24.
- **B — verify locally** in a real browser: serve the tree, run the fixed
  `toolaudit.py` over all 24, compare per-tool against the baseline. A tool that
  passed before the port must pass after it. **This gate does not need DNS.**
- **C — publish**: commit to the shared `sites` repo with an explicit pathspec,
  push, and let the Action `b2 sync` populate `b2://portfolio-sites/loanandmortgagecalculator.co.uk/…`.
  The bucket keys are correct before DNS exists, because the worker maps
  `{hostname}{path}`.
- **D — OWNER, blocking, not scriptable here**: add the Cloudflare zone
  (nameservers at the registrar) and a Workers Route
  `loanandmortgagecalculator.co.uk/*` → the same worker every other site uses.
  There are **no Cloudflare credentials on this machine** — the only token is a
  GitHub Actions secret scoped to cache purge. See RUNBOOK §1.
- **E — adopt into the framework**, with a **per-page policy split**:
  - the 24 calculators → `rebuild_policy='owned'`, verbatim components loaded from
    the **deploy-repo bytes** with a sha256 gate, so no rebuild can rewrite working
    arithmetic;
  - the guides → **framework-managed**, so they can evolve.
  This split is the point: D3 asks for the site to *evolve*, and a wholly-verbatim
  site cannot. `rebuild_policy` is per-page, so the framework already supports it.
- **F — record the divergence as spec** (D3): distinct `identity` /
  `content_direction` / `audience` on both sites.
- **G — report** what stays unmended, as `/bugs_open/` entries.

## Out of scope

Retiring `~/projects/domains/mortgagecalculator.co.uk`; redirects (D2 says no);
`www` records; the loancalculator lane's own adoption; fixing S1–S5 **on the old
site** (the brief is the new domain — worth offering separately).
