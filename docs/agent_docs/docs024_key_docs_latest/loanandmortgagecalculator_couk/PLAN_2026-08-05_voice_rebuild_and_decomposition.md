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

> **⛔ CORRECTED 2026-08-06 — OWNER RULING: DECISION 1 BELOW IS SUPERSEDED.**
> *"I want all the copy to be done through the framework and not through this
> cli session. We'll need to restart all those that have been written through
> this cli."*
>
> Decision 1 chose to TRANSFORM the existing copy in-session, on the reasoning
> that the pipeline writer refuses owned pages and there is no site plan to
> regenerate from. **Those facts are still true; they were the wrong thing to
> optimise for.** The reasoning treated "which route can a session drive
> today?" as the question, when the question was "who is allowed to write this
> site's words?" — and CLAUDE.md already answers that: EVERY SITE GOES THROUGH
> THE FRAMEWORK (owner ruling 2026-08-04, raised by a hand-built shopfront on
> the lane whose product is framework-built sites). A hand-authored rewrite is
> the same error wearing better checks: the guards I built verify that the copy
> preserves facts and links, not that the platform could ever reproduce it.
>
> **What this changes.** Everything structural in this plan STANDS — decompose,
> keep owned, widgets byte-original in locked rows, chrome, the verification
> ladder. Only the source of the WORDS changes. `load_lmc.py` now defaults to
> `manifest.json` (original copy); `voice_overlays/` (39 pages) and
> `manifest_voiced.json` are superseded and retained only as evidence of the
> register. The two pages already live with in-session copy are to be restored
> and re-decomposed with their original words.
>
> **What is now the open question, and it is a real one:** the platform's
> writer path REFUSES `rebuild_policy='owned'` pages
> (`save_page_sections_action.go:150`), so decomposition alone does not make
> the framework able to write here. See "Routing the framework's writer" below —
> it needs an owner decision, and the measurement that informs it is done.

## Decisions, with reasons

1. ~~**TRANSFORM the existing copy; do not regenerate.**~~ **SUPERSEDED — see
   the correction block above. Kept unedited because the reasoning is the
   record of how the wrong call was reached.** The owner approved
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

## Routing the framework's writer (OPEN — owner decision, measurement done)

Decomposition makes a page *editable*; it does not make the framework *able to
write it*. `save_page_sections_action.go:150` refuses any
`rebuild_policy='owned'` page outright. Two routes, and the risk in each is
now measured rather than argued.

**Route A — flip pages to `generic`, use the ordinary build pipeline.**
The stated risk is TL-001: `save_page_sections` does DELETE-then-reinsert over
`page_components`, which is how tool widgets have been destroyed before.
**MEASURED 2026-08-06 — the widget SURVIVES, with a caveat that matters:**

- The DELETE is itself lock-aware (`:708`, `pageComponentAgentWritableSQL`).
  Induced against the real decomposed `/loans/consolidation.html` inside a
  rolled-back transaction: `DELETE 2` — both prose rows went, **the locked
  `tool-1` row stood**. The statement removed rows, so it was live, not a
  no-op; ROLLBACK restored all three.
- Three defences in series, which is why a passing mutation here needs care:
  the lock-guarded DELETE, a Layer-1 "INTERACTIVITY REGRESSION BLOCKED" guard
  (`:580`) that fails a save whose new content has no interactive section when
  the old page had one, and a Layer-2 carry-forward (`:375-443`) that
  re-appends a dropped interactive section even when unlocked.
- **THE CAVEAT, and it is the deciding fact:** a locked row is repositioned by
  matching `slot_name` against the incoming section name (`matchLockedRow`).
  Our slots are positional (`tool-1`), so a writer's sections will NOT match,
  and an unmatched locked row is moved to `len(sections)+1` — i.e. **the
  calculator survives but lands at the BOTTOM of its page**, below all the new
  prose, with a `lock_blocked` work item raised. On a calculator page that is
  a serious layout regression, and it is silent unless someone looks at the
  page. [UNMEASURED end-to-end: this is read from the code plus the SQL-level
  test above, not from a full writer run against a live page.]

**Route B — drive `apply_section_edit` / section-editor.** Respects owned
pages and locked rows by construction, and edits a named section in place, so
position is never disturbed. Cost: it is per-section, so it needs a work item
per prose row rather than per page, and the section-editor workflow applies a
pre-authored edit — the *writing* step still has to come from somewhere.

**Recommendation, stated as a recommendation:** Route A is viable ONLY if the
tool rows are re-slotted to names the plan will emit, so `matchLockedRow`
matches and the widget keeps its position. That is a small, testable change to
the decomposition (name the slot after the tool, not after its index) and it
should be made BEFORE any page is flipped, not after a page has shipped with
its calculator at the bottom.
