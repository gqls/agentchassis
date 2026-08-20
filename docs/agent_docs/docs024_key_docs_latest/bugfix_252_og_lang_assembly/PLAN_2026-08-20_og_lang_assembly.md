# PLAN — bug 252 (og/lang half): per-page Open Graph at assembly, and `<html lang>` from the head component

**Lane opened 2026-08-20.** Target:
`bugs_open/252_HANDOFF_2026-08-11_assembly_drops_per_page_og_tags_and_hardcodes_html_lang_en_on_503_pages.md`

> **252 is an ambiguous number.** The other 252 (disk invisible to the scheduler) is CLOSED in
> `bugs_closed/`. Refer to this one by slug: `assembly_drops_per_page_og_tags_and_hardcodes_html_lang_en`.

---

## 1. What we are trying to do

Two defects that share one fix site, both on the page-assembly path:

- **A.** An assembled page cannot state its own Open Graph identity. The shared per-site `<head>`
  is one row per site, so every page it serves inherits site-level `og:*` values.
- **B.** `<html lang="en">` is hardcoded in Go on a British estate.

## 2. Correction to the originating brief — the bug MUTATED between filing and pickup

> **CORRECTION 2026-08-20 to the bug file's own description (filed 2026-08-11).** The file says an
> assembled page carries "the 2 generic tags the shared head holds (`og:type`, `og:site_name`)",
> i.e. that the per-page values are ABSENT. That was true when written. It is **no longer the
> defect**. Commit `d3f73a724` (imagery I1) added `injectBrandHeadTags`
> (`platform/orchestration/actions/render_site_components_action.go`), which bakes site-level
> `og:title` / `og:description` / `og:image` / **`og:url`** into the stored per-site head at chrome
> render time. `og:url` is built from the site origin — `https://<domain>/`.
>
> So the live defect is **worse in kind, not just in scale**: it is no longer a silence, it is a
> false statement. Every assembled subpage now declares the **homepage's** og:url. Absence would
> have been a missing share preview; this is 700 pages asserting the wrong identity, beside a
> canonical that (since `bugs_open/251`'s fix) correctly names the page. The page contradicts itself.
>
> **Why this matters beyond bookkeeping:** the fix candidate in the bug file — "add blank `og:`
> placeholders to the head components, fill them at assembly" — would now fix almost nothing.
> Placeholders exist on only 4 of 24 heads, and on those 4 they are already SHADOWED by
> `injectBrandHeadTags`' filled duplicates appended later in the same head. The design had to
> change shape as a result: see §5.

### Measurements taken at pickup (2026-08-19/20)

| what | value | how |
|---|---|---|
| head rows in `site_components` (`slot_name='head'`) | 24 | `count(*)` |
| of those, carrying `og:url` (= homepage) | **22** | `rendered_html LIKE '%og:url%'` |
| of those, carrying BLANK `og:title` placeholder + a filled duplicate | **4** | `head-seo-standard` sites: finetuning.uk, leopardessconsulting.co.uk, ai-agent-orchestration.com, gaswholesalers.com |
| assembled pages fleet-wide | **700** (was 503 at filing) | `pages` with any `page_components` row |
| sites with assembled pages | **26** (was 23) | `count(DISTINCT site_id)` same predicate |
| `sites` columns matching lang/locale/country/region | **0** (unchanged) | re-confirmed |

**Verified at the artefact, not only in the DB** (`https://ai-agent-orchestration.com/about.html`):
two `og:title` tags — one `content=""`, one the site name — plus `og:url` =
`https://ai-agent-orchestration.com/` while `<link rel="canonical">` correctly says `/about.html`;
and `<html lang="en">`.

## 3. Where we came from

- Filed 2026-08-11 by the LMC Track A lane, escalating what its own PLAN had carried as a *stated,
  accepted loss* (correctly so, at 2 pages).
- **Owner decision, same day** (commit `f666408ed`): for B, **option 3** — lang lives in the head
  component, Go stops emitting it. No `sites.language` column.
- **Owner decisions, 2026-08-20** (this lane): rollout is **canary then fleet-wide rerender waves**;
  **all UK sites opt into `en-GB` now** (a mechanism shipped with zero consumers rots unexercised —
  this estate has been bitten by that before).
- `bugs_open/251` (the canonical naming `/index.html`) was A's stated blocker. It is **fixed, live
  and council-APPROVED** (`61abbdbd0`, corr `33fb41cb`, round 1) — `preferredPageURL` normalises the
  site root only. og:url must call it, so og:url == canonical == JSON-LD `@id` by construction.

## 4. Decisions, and their reasons

**D1 — REMOVE-THEN-INJECT, not fill-a-placeholder.** Strip every `og:title`/`og:description`/`og:url`
from the stored head at assembly, then inject one per-page set. Reasons: (a) fill-blank-only fixes 0
of the 22 wrong-og:url heads, which have no placeholders; (b) on the 4 that do, the blank is
shadowed by a filled duplicate, so filling it changes nothing a scraper reads; (c) remove-then-inject
is idempotent by construction; (d) it self-heals both defect populations **at assembly time, with no
chrome rebuild** — which matters because Go changes regenerate no stored head (see D5). It does
change every assembled page, and that is the point: the value being removed is false.

**D2 — og:type / og:site_name / og:image / twitter:\* are NOT touched.** They are genuinely
site-level and their defects belong to `bugs_open/322`. Narrow blast radius, and it keeps this
task's claim reviewable.

**D3 — ordering: og splice runs BEFORE `spliceMetaDescription`.** That function's legacy fallback
fills the FIRST `content="">` anywhere in the head when the exact blank description tag is absent.
With blank og placeholders present, running it first could put the page description into an og tag.
Stripping first removes the hazard. The order is pinned by a test, not by a comment.

**D4 — lang travels as a `<head>` attribute, sourced from `site_specs`.** Head template emits
`<head{{if .lang}} lang="{{.lang}}"{{end}}>`; the value comes from `site_config.locale.lang` via the
existing schema-driven config resolution (the STY-050 mechanism, worked precedent migration `339`);
`assemblePage` reads the attribute off the head it already holds and stamps `<html lang=…>`, default
`en` when absent. Reasons: no schema change (owner's option 3); **zero Go on the value path**, so a
site opts in with a spec row; and the default keeps every non-opted site byte-identical.

**D5 — the fleet propagation is the EXISTING staleness pipe, deliberately.** `site_components.render_inputs`
hashes the head **template and site_specs by value**, so the two migrations move the fingerprint and
`StaleSiteComponentsCheck` files its own `stale_chrome` item per site → chrome re-render + per-page
fan-out. We are not building a new sweep. (Go code is NOT a fingerprint input — this is why the og
half had to work without a chrome rebuild.)

**D6 — the other four `<html lang>` emitters stay hardcoded, each for a named reason.** Not laziness:
`AssemblePageAction`'s head is `RenderFallbackHead` fleet-wide, so no configured value can reach it
(wiring it would ship inert plumbing); `buildPageHTML`'s output is re-headed by that path anyway;
about/contact are static stubs; `rerenderSinglePage` is unregistered, DEPRECATED and was mid-refactor
by another session. Recorded so the next audit does not read the omission as an oversight.

**D7 — producer convergence is NOT attempted here.** `assemblePage` and `AssemblePageAction` are two
independent head producers, and the register (`docs026_concept_register/register/seo.md`, SEO-003)
holds "should they converge" as an OPEN architecture-scope question. This lane adds
`platform/orchestration/actions/head_assembly.go` and registers it as the **named seam** for that
convergence, so the next head fix has a home — but it does not decide the question, and it does not
move the existing injectors (churn without behaviour change, across files other sessions hold).

## 5. Shape of the change

New `platform/orchestration/actions/head_assembly.go`: `spliceOpenGraph`, `headLangAttr`,
`htmlDocumentOpen`. Wired into `assemblePage` only. `injectBrandHeadTags` loses its og:url line
(a per-site artefact must not assert a page URL). Two migrations: head templates carry lang (and
`head-seo-standard` loses its blank og lines), and `site_specs` gains `locale.lang = en-GB` for the
UK estate. Register entry SEO-005 in the same commit as the Go.

Out of scope and routed: `bugs_open/322` items 2/3/4/5 (og fallback copy, og:image 404 gating —
**landmine: never gate on an assets row**, the wholesale idempotency skip, favicon source); producer
convergence (register); the one site with no stored head, which keeps `buildDefaultHead`'s no-og
behaviour.

## 6. Phasing

0. **Openers** — this lane's five docs; `090` diagnosis dispatched (run corr `af31ec22`); coordination
   notes committed into `bugs_open/252` and `bugs_open/322`.
1. **Go + tests** — the helper file, the two `assemblePage` call sites, the emitter one-liner
   (separate commit), every test mutation-proven.
2. **Council gate** — one submission covering the Go and both migrations (**scope widened
   2026-08-19 to include appliable migrations, `bugs_open/314`** — a migration is the running system).
3. **Roll** — owner's whole-fleet release cadence; prove the binary per service before anything else.
4. **Canary** — 2 pages on a duplicate-tag site, verified at the artefact.
5. **Migrations** — only after the binary is proven live (see the trap in RUNBOOK).
6. **Fleet waves + per-site verification** across all four head families.
7. **Close** at fixed AND live, with the 016b §9 pattern and the LANDMINES entry.

## 7. Risks

- **Same-file passenger.** `render_site_components_action.go` was held dirty by the bug-260 session
  at planning time. Re-check before every touch; keep that edit to one line in its own commit.
- **A shared template serves 18 sites.** Editing "Document Head" is a fleet action, not a site
  action. Per-site values must go through `site_specs`, gated `{{if .field}}`, and the schema entry
  must be **map-valued** (a scalar is silently skipped by the resolver).
- **Migration-before-image inverts the rollout** — see RUNBOOK's named trap. Both migrations are
  `_HOLD` for this reason.
- **`MetaDesc` is empty on ~56% of pages** (`bugs_open/320`, backfiller now scheduled), so
  og:description will be legitimately absent on many pages at first. Correct-or-absent, by design —
  do not read those as failures.
