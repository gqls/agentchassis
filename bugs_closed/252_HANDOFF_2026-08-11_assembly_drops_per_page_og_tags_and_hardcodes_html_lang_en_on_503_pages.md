# 252 — assembly drops per-page `og:*` and hardcodes `<html lang="en">` on 503 pages

> ## ✅ CLOSED 2026-08-21 — FIXED, LIVE and PROVEN AT THE ARTEFACT
>
> **Live on chassis `v1.0.1320`, still present on `v1.0.1321`** — `spliceOpenGraph`, `headLangAttr`
> and `htmlDocumentOpen` probed PRESENT in the running binary on **both replicas**, with a positive
> control that fired and a fabricated negative control that did not. Council **APPROVED round 1**
> (`3b6712d4-4565-4bfe-87f6-c47ecefd6a93`), all four seat-requested checks run and answered.
>
> **Proven on real pages, in both directions.** `ai-agent-orchestration.com/about.html` went from two
> `og:title` tags (one blank, one the site name) and an `og:url` naming the **homepage**, to one
> `og:title` carrying the page title and an `og:url` naming the page — **agreeing with its canonical**
> — plus `<html lang="en-GB">`. The homepage is the discriminating control: its `og:url` came out as
> the bare `/`, not `/index.html`, so `preferredPageURL`'s root normalisation carries through.
>
> **Fleet stored heads fully repaired** (24 rows): fingerprint drift 22 → **0**, heads carrying a
> language 0 → **22**, heads baking a homepage `og:url` 22 → **0**, blank `og:title` duplicates
> 4 → **0**. `relojistas.com` carries `<head lang="es-ES">` — the owner's ruling that non-English sites
> must not be `en-GB` (and that this generalises to future language sites), working end to end.
>
> **Residual, tracked not ignored: `bugs_open/346`.** A repaired stored head is not a repaired PAGE —
> a page picks it up when it next re-assembles. 502 of 727 pages had not, and **twelve real sites are
> at zero**, with a natural rebuild rate of ~1 page/hour. **Owner ruling 2026-08-21: do not force those
> rebuilds** (≈500 items into `bugs_open/083`'s queue is other lanes' exposure). The defect is dead —
> it cannot recur, and any rebuild from now produces a correct page. The damage decays on its own.
>
> **Where everything lives:** lane
> `docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/` (PLAN with decisions D1–D7,
> RUNBOOK, NOTES with the full misstep log, README owner log, FINDINGS, DECISIONS) · register
> **SEO-005** · pattern `016b` §9 · `LANDMINES.md` (the 090 loop cannot see served bytes) ·
> `WRONG_CALLS.md` (three entries of mine) · migrations `507_head_components_carry_lang.sql` and
> `508_site_specs_locale_lang.sql`, both applied and recorded.
>
> **Spun out, still open:** `bugs_open/346` (the residual above) · `bugs_open/347`
> (webdesign.co.uk's head component has no `<head>` element — 117 pages, our largest site, and it will
> never gain a language until that is fixed) · `bugs_open/322` item 4 — **the guard that made this bug
> possible is untouched, so a future per-page tag added to that block reproduces 252 exactly.** That
> is the mechanism; this bug was one symptom of it.
>
> ⚠ **SEO-005 carries a binding threshold set by the council's architecture seat:** this was the
> FOURTH head fix to land on one of the two head producers. **A fifth raises an RFC on SEO-003 rather
> than taking a fifth patch.**

**Filed 2026-08-11** at the owner's direction, escalating what has until now been
carried as a *stated, accepted loss* in one lane's PLAN. **Platform, fleet-wide.**
Not a regression: this is how assembly has always behaved. What changed is the
scale — the loss applied to 2 pages on 2026-08-06, 19 after LMC Track A, and
**503 assembled pages already exist fleet-wide**.

> **On the 2026-07-31 ruling: `090` substituted, and here is the why.** No part of
> this is inferred. Both behaviours are read directly from source (line numbers
> below), the affected population is a single `COUNT`, and the seam the fix should
> use is an existing, working mechanism in the same function. There is no
> not-where-you-are-looking cause for the loop to find. What *is* uncertain is a
> product decision (which locale, and whose), and no diagnosis run settles that.

---

## Two separate defects, filed together because they share a fix site

### A. Per-page `og:title` / `og:description` / `og:url` are lost on every assembled page

The shared `<head>` component is one row per site, so it cannot carry per-page
values. A hand-built page carries 5 `og:` tags; an assembled one carries the 2
generic tags the shared head holds (`og:type`, `og:site_name`). Measured on LMC:
verbatim page `og:5`, assembled page `og:2`.

**This is a solved problem in the same function.** Assembly already splices two
per-page values into the shared head:

- `rerender_single_page_action.go:618` — regex-replaces `<title>…</title>`
- `rerender_single_page_action.go:625` → `spliceMetaDescription` (:1017) — fills
  a **blank placeholder** `<meta name="description" content="">` left in the
  site-level head, and *removes* the placeholder when the page has no description.

So the pattern is established and proven: **the shared head carries a blank
placeholder; assembly fills it per page.** `og:` wants exactly this, and nothing
more novel. Confirmed the LMC head component already has the blank-description
placeholder (`t`) and no `og:title` (`f`), so the two are directly comparable.

**Why this shape is the right one, and is safe:** a site whose head component has
no `og:` placeholder gets no `og:` tags — i.e. **unchanged behaviour**. The
capability is opt-in per site, reachable by nothing until a head component names
it. Per the OWNER RULING of 2026-07-29 §1 that makes it a normal council-gate
change rather than an RFC: it adds an opt-in capability without changing what the
shared mechanism *guarantees* to anyone not using it.

### B. `<html lang="en">` is hardcoded, in four places, on a British estate

```
platform/orchestration/actions/rerender_single_page_action.go:670
platform/orchestration/actions/rerender_pages_actions.go:527
platform/orchestration/actions/multipage_actions.go:1024, :1053
platform/orchestration/actions/ai_actions.go:1250        (template)
```

Every assembled page on every site declares `en`, regardless of locale. The
hand-built LMC pages declared `en-GB`. **`sites` has no language, locale, country
or region column** — checked; nothing to read from. So this is not "wire up the
existing field", it is "decide where locale lives, then wire it".

**Scale: 503 assembled pages fleet-wide** — this is not an LMC issue, and fixing
it in one lane's chrome would be the wrong place.

Consequence is modest but real: screen readers pick voice and pronunciation from
`lang`, and `en` on a UK finance site gives an American reading of currency and
place names. It is also simply false metadata on a British estate whose own
platform conventions specify British English.

## Blast radius, measured 2026-08-11

| | value |
|---|---|
| assembled pages fleet-wide (`sections <> ["ported-page"]`) | **503** |
| LMC pages affected by A | 19 (was 2 on 08-06) |
| sites with an assembled homepage | 23 |
| `sites` columns matching lang/locale/country/region | **0** |

## Fix candidates

**For A — follow `spliceMetaDescription` exactly.**
Add `spliceOpenGraph(head, page)` beside it, filling blank placeholders
`<meta property="og:title" content="">` etc., and removing any placeholder the
page cannot fill (that removal behaviour is already the established convention —
copy it, do not invent a different one). Then add the placeholders to the head
components of sites that want them. `og:url` should use the **same** normalised
URL as the canonical, which means it inherits `bugs_open/251` — **fix 251 first
or A will faithfully reproduce the `/index.html` error into `og:url` as well.**

**For B — decide where locale lives before writing code.** Options, cheapest first:
1. A `sites.language` column defaulting to `'en'`, read at the four sites. Honest,
   small, and the default preserves today's behaviour exactly.
2. Derive from the domain TLD (`.uk` → `en-GB`). Rejected: `.com` sites in this
   estate are British too, so it guesses wrong in the direction that matters.
3. Put it in the head component and stop emitting `<html lang>` from Go. Rejected
   for the same reason `og:` cannot live there — except it *can*, because `lang`
   is genuinely per-site not per-page. **This is worth considering seriously**: it
   needs no schema change and no Go change on three of the four sites.

**Do not fix B in one lane's chrome file.** With `<html lang="en">` emitted from
Go, a chrome-level change would be overwritten; with option 3 it would not. Settle
the mechanism first.

> **OWNER DECISION 2026-08-11: option 3.** Put `lang` in the head component, stop
> emitting `<html lang="en">` from Go. No schema change; behaviour unchanged on
> three of the four affected sites (they'll get the same default via the head
> component that Go emits today); the fourth needs its head component to actually
> name a language once this lands. Mechanism now settled — B is unblocked.

## Scope note

Shared render-path Go on the `page-rerender` path → **council gate**. Name the
other consumers in the submission: `AssemblePageAction` (`multipage_actions.go`),
used by three active agent types, is a **second head producer that does not share
these code paths** — the standing landmine recorded in
`docs026_concept_register/register/seo.md` is that `injectCanonicalLink`,
`injectPageJSONLD` and `injectRobotsNoindex` all live on one producer only. A fix
to A that lands on `assemblePage` alone will leave the rebuild path emitting no
`og:` at all, which is the same divergence a third time.

## STATUS 2026-08-20 — TAKEN, and the defect has MUTATED since filing

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/`
(PLAN / RUNBOOK / NOTES / README_where_we_are). `090` diagnosis run corr
`af31ec22-5662-4798-91b9-b12132ebca70`. Both owner decisions below are executed by that lane.

> **CORRECTION to §A of this file, measured 2026-08-19/20.** §A says an assembled page carries
> "the 2 generic tags the shared head holds (`og:type`, `og:site_name`)" — i.e. the per-page values
> are ABSENT. **That is no longer the defect.** Commit `d3f73a724` (imagery I1, AFTER this file was
> filed) added `injectBrandHeadTags` (`render_site_components_action.go`), which bakes site-level
> `og:title`/`og:description`/`og:image`/**`og:url`** into the stored per-site head — and builds
> `og:url` from the site origin.
>
> So the live defect is **worse in kind**: not a silence but a false statement. Every assembled
> subpage declares the HOMEPAGE's og:url, beside a canonical that (since `bugs_open/251` shipped)
> correctly names the page. Verified at the artefact, `https://ai-agent-orchestration.com/about.html`:
> `og:url` = `https://ai-agent-orchestration.com/`, canonical = `…/about.html`, and **two `og:title`
> tags** — one `content=""` from the `head-seo-standard` template, one filled with the site name by
> the injector, whose idempotency guard (`rel="icon"` OR `og:image`) cannot see the blank.
>
> **Consequence for this file's fix candidate:** "add blank `og:` placeholders and fill them at
> assembly" would now fix almost nothing. Placeholders exist on 4 of 24 heads, and on those 4 they
> are already SHADOWED by the injector's filled duplicates. The design is therefore
> **remove-then-inject** at assembly (strip `og:title`/`og:description`/`og:url` from the stored
> head, inject one per-page set) — which self-heals all 22 wrong-og:url heads AND the 4 duplicated
> ones with no chrome rebuild, and is idempotent by construction.

**Blast radius re-measured** (the table above is 08-11 and stale): assembled pages **700** (was 503),
sites **26** (was 23), head rows 24 of which **22 carry og:url**, 4 carry a blank+filled duplicate
pair. `sites` lang/locale columns: still **0**.

**§B is unblocked and unchanged in substance:** all five real `<html lang="en">` emitters were
re-confirmed present (this file's cited lines are stale but the code is not fixed — a literal
`lang="en"` grep cannot see `lang=\"en\"`, logged in `WRONG_CALLS.md`). Only `assemblePage` is being
made lang-aware; the other four are left hardcoded, each for a reason recorded in the lane PLAN (D6).

**`bugs_open/251`, named below as A's blocker, is DISCHARGED** — `61abbdbd0` (`preferredPageURL`),
council corr `33fb41cb` APPROVED r1, verified live. `og:url` calls it, so og:url == canonical ==
JSON-LD `@id` by construction.

### What is now BUILT (2026-08-20) — committed, not yet live

**Go** (`Council-Submitted: 3b6712d4-4565-4bfe-87f6-c47ecefd6a93`):
- NEW `platform/orchestration/actions/head_assembly.go` — `spliceOpenGraph` (remove-then-inject:
  strips `og:title`/`og:description`/`og:url` from the stored head, injects one per-page set;
  `og:url` via `preferredPageURL`, so og:url == canonical == JSON-LD `@id` **by construction**;
  correct-or-absent), `headLangAttr`, `htmlDocumentOpen`.
- `rerender_single_page_action.go` — both call sites in `assemblePage`. **The og splice runs AFTER
  `spliceMetaDescription`**, which looks backwards; see the code comment and the correction note in
  the lane PLAN D3.
- `render_site_components_action.go` — one line: `injectBrandHeadTags` no longer emits `og:url`.
- 11 claims in `head_assembly_test.go`, **every one mutation-proven** (M1–M9 ledger in the lane NOTES).
  M1 is worth knowing: the mutation reproducing **this file's own fix candidate** fails five of them.

**Config, both `_HOLD`** — `docs/agent_docs/sql_for_agents/507_head_components_carry_lang_HOLD.sql`
(+`_ROLLBACK`) and `508_site_specs_locale_lang_HOLD.sql` (+`_ROLLBACK`). 507 gives both shared head
templates a gated `lang` attribute (map-valued schema entry; the two components use DIFFERENT schema
shapes) and removes `head-seo-standard`'s two blank-rendering og lines. 508 sets `site_config.locale.lang`
for the 25 real sites. **⚠ Do NOT apply either before the binary carrying `head_assembly.go` is proven
running on every replica** — DB config is live on apply, Go is inert until the roll, and the old code
would consume the staleness edge, restamp `render_inputs`, and leave the pipe quiet with the fleet
still wrong.

Registered as **SEO-005**; 016b §9 pattern filed; `LANDMINES.md` gained the 090-cannot-see-served-bytes
entry. Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/`.

**⚠ ONE SITE IS NOT ENGLISH.** `relojistas.com` is a Spanish-language publication (identity location
`España`, Spanish tagline, Spanish headings on the served page). It gets `es-ES`; a blanket en-GB
would have been false metadata stated more confidently than the `en` it says today — the exact defect
class this file is about. This is why 508 names every domain explicitly instead of deriving from the
TLD, which §B of this file had already rejected.

**⚠ `webdesign.co.uk` will keep serving `lang="en"` and that is NOT a failure.** Its head component
(`webdesign-couk-head`, its own `function`, 1 site, **117 assembled pages — the most in the fleet**)
is a bare fragment with no `<head>` open tag to carry the attribute. Its pages still gain per-page og
from the Go half; the append path is pinned by a test.

### Still owed before this can close

1. Fleet roll, then **prove the binary per service** (stamp + `merge-base --is-ancestor`, or a binary
   probe with BOTH controls).
2. Canary two pages on a duplicated-tag site (the og half needs no migration), verified at the artefact.
3. Apply 507 then 508; let the `stale_chrome` pipe fan out; sweep all 26 sites across **all four head
   families**.
4. **Retire the premises this fix falsifies, which will NOT fail loudly:** `verify_site.py`'s
   `OG_PER_PAGE` accepted-loss exemption (`…/loanandmortgagecalculator_couk/verify_site.py:71`) and
   the og:url-exclusion rationale in
   `discovery_checks/check_site_structural_validity.go` (~`:55`, `:1029`). Both currently document
   "the shared `<head>` cannot carry a per-page value" as settled fact.

### `090` outcome, recorded rather than buried

Run corr `af31ec22-5662-4798-91b9-b12132ebca70`: **UNVERIFIABLE, stopped at the iteration cap — not
REFUTED.** It confirmed the static mechanism from source (18 symbols, citations quoting
`injectBrandHeadTags`' own `WriteString`) and could not reach an occurrence, for two structural
reasons it states itself: `pages.rendered_head` returns 0 rows fleet-wide (those columns are
VESTIGIAL — `bugs_closed/270`), and every `site_components.rendered_html` row it fetched was truncated
before the `</head>` tail where the block sits. **The gap it names is closed by the `curl` evidence
above, which predates the run.** Filed as a LANDMINE so the next session does not read an UNVERIFIABLE
about served markup as doubt about their claim.

**Division of work with `bugs_open/322`** (filed 08-19, the emitter-side twin): this lane takes 322's
item 1 *at the assembly end* (per-page og identity) plus the one-line removal of the injector's
`og:url` emission. **322 keeps** item 2 (og:title/description fallback quality), item 3 (og:image
emitted whether or not the file exists — landmine: do NOT gate on an assets row), item 4 (the
wholesale idempotency skip that opts webdesign.co.uk out entirely), item 5 (favicon source). Producer
convergence stays the register's open architecture question (SEO-003); this lane adds
`head_assembly.go` as the named seam for it without deciding it.

## See also

- `bugs_open/251` — the canonical names `/index.html`; **A depends on it**, since
  `og:url` should agree with the canonical.
- `loanandmortgagecalculator_couk/PLAN_2026-08-05` §6 — where these were first
  recorded as accepted losses, and correctly so at 2 pages.
- `docs026_concept_register/register/seo.md` — the two-head-producer landmine.
