# PLAN — loanandmortgagecalculator.co.uk voice rebuild + decomposition (2026-08-05)

Owner directive (portfolio_positioning handoff, 08-05): apply the chosen
"gentle explanatory" voice (`portfolio_positioning/VOICE_gentle_explanatory_v1.md`,
trial H, owner-approved with four sample transformations) to the whole site —
*"do the whole site — I'll check it then."* The seeded `content_direction` v2
(2026-08-05 row) is the contract; this plan is how the 41 live pages actually
change.

**The site is LIVE** (HTTP 200 at https://loanandmortgagecalculator.co.uk).
Site id `ed633ada-f8af-424b-b4d4-8af79160dbcd`. All 41 pages are
`rebuild_policy='owned'`, exactly one `ported-page` component each,
`content_data.deploy_mode='verbatim'` (verified 1|1 on all 41, 2026-08-05).
No open work items on the site. No site plan exists (no `site_plan` spec, no
`site_plan_sections`).

**This lane pre-existed this plan** — the site was hand-built here (07-31,
`build_site.py`/`build_pages.py`/`guides_content.py`), went live, and was then
adopted verbatim. Read `NOTES_loanandmortgagecalculator_couk.md` (both 07-31
sessions) and BOTH `CONTRIB_*` files before touching anything. Standing
warnings inherited from there: this site has **no `site_components` rows**, so
the first decomposed page would assemble against `buildDefaultHead`, whose
stylesheet link is `styles.css` (plural) — a 404 here — with no header or
footer at all; hence chrome is Phase 1 and blocks everything. A `page_rerender`
filed `status='detected'` is never dispatched — always file `'triaged'`.
**Stored rows verified against the deploy repo 2026-08-05: 41/41 md5-identical
to `origin/master` `b318a8fad`** — the adoption-era crawl-DOM divergence the
07-31 notes flagged as needing repair is not present in what is stored today.
After decomposition the DB becomes the render source and git-adapter writes
the repo: the `build_site.py` era ends at that moment, per page.

## Decisions, with reasons

1. **TRANSFORM the existing copy; do not regenerate.** The owner approved
   *transformations* of four real copy blocks — that is the reviewed method.
   Fresh regeneration (needs_page → page-build-handler) has no site plan to
   drive it, would discard hand-built substance, and the owner queue holds an
   FCA citation pass owed "before any regeneration". The pipeline writer path
   is additionally REFUSED on owned pages (`save_page_sections_action.go:150`),
   and flipping 41 pages to generic for one rewrite would hand 23 calculator
   pages to the TL-001 clobber. The seeded content_direction v2 governs every
   *future* writer run; the transformation follows the same prompt, so the
   voice is consistent either way.
2. **Keep every page `rebuild_policy='owned'`.** Future copy maintenance goes
   through the sanctioned targeted-edit route (`apply_section_edit` /
   section-editor), which owned pages permit. This mirrors
   loancalculator.co.uk, which has run this way in production since 08-02.
3. **Decompose each page** — REPLACE the verbatim row with positional
   `prose-N` / `tool-N` rows **in one transaction** (LANDMINES: inserting
   alongside flips the page to assembly with the whole document as a section —
   nested `<html>`). Pages stay owned; assembly takes over from verbatim
   because the row count leaves the 1|1 state.
4. **Chrome extracted once** to `site_components` (head / header / footer),
   locked on creation, following the loancalculator pattern including the CSS
   shim (style `main` as the container; neutralise `.container` inside
   sections). This site's chrome is uniform across pages. site.js (mobile
   menu only — its own header comment says no calculator logic) rides at the
   end of the FOOTER component, as in the original body end.
5. **Widgets ship byte-original, in LOCKED tool rows.** No component
   templating (that was loancalculator's separate multi-session project).
   A calculator page's tool block = its cards + inline `<script>` +, on
   mortgages pages, the `<script src="/assets/js/calculators.js">` tag that
   precedes the inline script (11 loans pages are self-contained; 12
   mortgages pages need calculators.js folded into the prover's id set AND
   carried in the tool row). "Copy zones only on tool pages" = prose rows
   are the only thing the voice touches there.
6. **Accepted losses, stated:** per-page `og:*` tags (assembly has one shared
   head; the sibling site shipped the same way — JSON-LD + canonical are
   injected per page by assembly instead); nav `aria-current="page"` (shared
   header; `.nav-links a[aria-current]` styling simply never matches);
   `lang="en-GB"` becomes assembly's `lang="en"`. Each is visible in the
   assemble-mirror diff and none is silent.
7. **legal.html decomposes but its copy is UNTOUCHED** (voice rule 10).
   404.html is not a page row and is untouched. Facts, figures, links,
   element ids and anchors are preserved exactly in every transformed block —
   the transformation moves REGISTER, never content. Mechanical check:
   the multiset of numeric tokens and hrefs before == after, per page.

## Order of work

- **Phase 0 — baselines (before touching anything):**
  golden behavioural capture of all 23 calculators from the LIVE site
  (`toolgolden.py`, sibling lane, chromium via CDP); md5 census of all 41
  stored `rendered_html` rows; stored-vs-`origin/master` byte agreement per
  page (pin the sites-repo sha in the baseline file name).
- **Phase 1 — chrome:** extract, shim, verify resolution (every href/src
  fetches 200 with a browser UA; nav links join `pages.url`; head carries
  the literal `<title></title>` and `content=""` splice targets), load
  locked site_components.
- **Phase 2 — canary:** ONE guide (voice transform, full ladder) then ONE
  loans calculator (copy zones + locked widget + golden compare). Checkpoint:
  compare served shape against the handoff's expectations before batching.
- **Phase 3 — batches:** guides (12) → home + 3 section indexes → loans
  calculators (10) → mortgages calculators (12) → legal.html (decompose
  only). Per page: prover (script-target containment, no visible text lost
  pre-transform) → assemble-mirror diff (only intended deltas) → replace
  rows in txn → assemble-only `page_rerender` (NO spec.reason, page_id in
  spec AND column, status 'triaged') → served check (~2 min CDN lag; B2
  NoSuchKey guard from the sibling's acceptance probe).
- **Phase 4 — final census:** 23/23 golden exact; voice present on every
  transformed page; 41/41 HTTP 200; URLs unchanged (`git ls-tree` compare);
  update working docs + handoff; owner reviews the finished site.

## Verification inheritances (sibling lane, proven 08-02→08-04)

- decompose_prover P-checks with EXTERNAL scripts folded into the id set;
  unreadable referenced script = hard failure.
- assemble_mirror before any row is written.
- Post-deploy probe with the B2 NoSuchKey + DOCTYPE guards, induced non-zero
  first.
- Two canaries minimum before bulk (the sibling's two canaries DISAGREED).
- Dispatch discipline: nothing within ~300s of a chassis pod (re)start;
  find runs by payload, not printed id.
