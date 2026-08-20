# HANDOFF — bug 252 (og/lang slug), 2026-08-20 · COLD-START, read this first

**State: code COMPLETE, council APPROVED, BLOCKED on a chassis build.** Nothing is live. Nothing has
been applied. No page has changed.

Target bug: `bugs_open/252_HANDOFF_2026-08-11_assembly_drops_per_page_og_tags_and_hardcodes_html_lang_en_on_503_pages.md`
⚠ **252 is an ambiguous number** — the other one (disk/scheduler) is CLOSED. Refer to this by slug:
`assembly_drops_per_page_og_tags_and_hardcodes_html_lang_en`.

---

## 1. The one thing blocking everything

**The deployed chassis (`v1.0.1319`, cut 2026-08-20T10:18Z) does NOT contain this fix.** The Go was
committed at 14:03Z — about four hours later. Proven at the binary, not inferred from the tag
(2026-08-20 14:35Z, all three arms):

| symbol | expectation | result |
|---|---|---|
| `spliceOpenGraph` | present iff the fix shipped | **absent** |
| `injectCanonicalLink` | must be PRESENT (positive control) | PRESENT |
| `zzQuiteImpossibleSymbol42` | must be absent (negative control) | absent |

**A chassis build cut from current HEAD is needed.** The owner runs releases (whole-fleet) — this lane
deliberately did not bump `IMAGE_TAG` or build.

⚠ **Do NOT apply migrations 507/508 before that build is live and probed.** DB config is live on
apply; Go is inert until the roll. Applied against the old binary, the old code consumes the chrome
staleness edge, re-stamps `render_inputs`, and **the detect→rebuild pipe goes quiet with the fleet
still wrong** — needing a manual re-fire. That is why both files are `_HOLD`.

## 2. What is committed (all on branch `087_towards_multiple_domains`)

| commit | what |
|---|---|
| `4a9a4d818` | lane opened, premise re-measured, coordination into 322 + LMC |
| `4abcd55a4` | **the Go**: `head_assembly.go` + tests + the two `assemblePage` call sites + SEO-005 |
| `028c8544c` | emitter one-liner: `injectBrandHeadTags` stops emitting `og:url` |
| `a7fc3f1af` → `acb2794af` | migrations, then renumbered 502/503 → **507/508** |
| `1caec47e3` | the 090 landmine + lane RUNBOOK |
| `66a1f9928` | 016b §9 pattern |
| `07994916b` | bug-file status block |
| `ec1d122c9` | **council answers + FINDINGS doc** (`Council-Reviewed: 3b6712d4-…`) |

**Council: APPROVED round 1**, corr `3b6712d4-4565-4bfe-87f6-c47ecefd6a93`, 5 advisories, none
high-severity. All four seat-requested checks were run and answered — see NOTES 2026-08-20 (f) and
`507`'s header.

## 3. What the fix does, in one paragraph

`site_components.rendered_html` holds ONE `<head>` row per site, reused by every page `assemblePage`
builds — so a page-scoped value baked into it is a claim about every page. 22 of 24 heads carried
`og:url` = the **homepage**, so all 700 assembled pages across 26 sites asserted the wrong identity
beside a correct per-page canonical. New `platform/orchestration/actions/head_assembly.go`:
`spliceOpenGraph` **removes** `og:title`/`og:description`/`og:url` from the stored head and injects
one per-page set (`og:url` via the existing `preferredPageURL`, so og:url == canonical == JSON-LD
`@id` **by construction**); `headLangAttr` + `htmlDocumentOpen` move the document language into the
head component per the owner's option-3 ruling, defaulting to `en`.

## 4. Resume here — the exact sequence

1. **Owner cuts a chassis build from HEAD and rolls it.**
2. **Probe every replica** for `spliceOpenGraph` with both controls (recipe in §1 and in
   `RUNBOOK_og_lang_assembly.md` §9). ⚠ Do **not** use
   `kubectl logs … | grep 'build provenance'` — it returned 2.4MB of another lane's payload today
   (the chassis logs whole council payloads, and they quote that phrase).
3. **Canary the og half — it needs NO migration.** Two pages on a duplicated-tag site:
   `ai-agent-orchestration.com` `/about.html` and `/index.html`, via direct dispatch
   (`049b_deploy_single_page.sh`), **not** the `spawn_agent`→`call_agent` wrapper, which has hung.
   ```bash
   curl -s "https://ai-agent-orchestration.com/about.html" \
     | grep -oE '<meta property="og:[^>]*>|<link rel="canonical"[^>]*>|<html[^>]*>'
   ```
   Expect: `og:url` == the page URL == the canonical; exactly ONE `og:title`, carrying the page title;
   zero `content=""` og tags; on the homepage `og:url` is the bare `/`.
4. **Take the backup** named in 507's header, then apply **507 then 508**; record with
   `run-migrations.sh --record-only`, scoping the directory.
5. The `stale_chrome` pipe fans out on its own (chrome re-render + per-page rerenders, per site).
   Direct lever if it stalls: the `rerender-chrome` agent (seed `351`).
6. **Sweep all 26 sites** with the curl above, across **all four head families** — `Document Head`
   (18 sites), `head-seo-standard` (4), the `webdesign.co.uk` fragment (1), and the site with no
   stored head.

## 5. Three results that will look like failures and are not

- **`webdesign.co.uk` will keep serving `lang="en"`.** Its head component is a bare fragment with no
  `<head>` open tag to carry the attribute. Its 117 pages still gain per-page og. See FINDINGS §A1.
- **Many pages will have no `og:description`.** Correct-or-absent by design — 55.7% of pages have no
  `meta_description` (`bugs_open/320`, backfiller now scheduled). An absent tag there is the mechanism
  working.
- **`relojistas.com` gets `es-ES`, not `en-GB`.** It is a Spanish-language publication (identity
  location `España`, Spanish headings live). **This is the one decision awaiting the owner's
  confirmation** — he asked for "all UK sites"; this corrects a non-UK one while in there.

## 6. Still owed after it goes live

1. **Retire the two checks whose premises this falsifies — neither fails loudly.**
   `verify_site.py:71` (`OG_PER_PAGE` accepted-loss exemption, LMC lane owns that file) and the
   og:url-exclusion rationale in `discovery_checks/check_site_structural_validity.go` (~`:55`, `:1029`).
   Both currently document "the shared `<head>` cannot carry a per-page value" as settled fact.
2. **Move the bug to `bugs_closed/` only at fixed AND live** — name **both** paths on the commit
   (`git mv` + a one-sided pathspec ships a copy) and verify at HEAD with `git ls-tree`, not `ls`.
3. **Tell `bugs_open/322`** when it lands; its item 1 shrinks but does not close.

## 7. Incidental findings, written up separately at the owner's request

`FINDINGS_2026-08-20_errors_caught.md` — six items. Highest value: **webdesign.co.uk serving 117 pages
with no `<head>` element**; the **untouched wholesale idempotency guard** (`322` item 4) being the
generic mechanism of which og:url was only a symptom; **migration numbering having no allocator**
(five consecutive numbers were taken by three lanes while I wrote two files); and the **090 diagnosis
loop's blindness to served bytes**.

⚠ **SEO-005 now carries a binding escalation threshold**, set by the council's `architecture` seat:
this is the **fourth** fix to land on one head producer only, and **a fifth instance raises an RFC on
SEO-003 rather than taking a fifth patch.**

## 8. Lane documents

- `PLAN_2026-08-20_og_lang_assembly.md` — design + decisions D1–D7 (D3 carries a correction)
- `RUNBOOK_og_lang_assembly.md` — every command with its gotcha
- `NOTES_og_lang_assembly.md` — evidence and every misstep, newest at the bottom
- `README_where_we_are.md` — the owner's plain-prose log
- `FINDINGS_2026-08-20_errors_caught.md` — incidental defects
- Register: `docs026_concept_register/register/seo.md` **SEO-005** · pattern: `016b` §9 ·
  traps: `LANDMINES.md` (090-cannot-see-served-bytes) · my errors: `WRONG_CALLS.md` (two entries)
