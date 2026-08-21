# PLAN — agritec.uk rebuilt inside the framework

Started 2026-08-21. Workstream directory:
`docs/agent_docs/docs024_key_docs_latest/agritec_uk/`.

Companion files: `SUBJECT_LEDGER.md` (the completeness contract — read it first),
`RUNBOOK_agritec_uk.md` (the commands), `NOTES_agritec_uk.md` (running record, missteps),
`README_where_we_are.md` (the owner's plain-prose log).

## 0. Why this exists

`agritec.uk` is live now — HTTP 200, on our own Cloudflare nameservers
(`leah`/`alexis.ns.cloudflare.com`, the same pair as oufe.com), served from a Backblaze B2
bucket, `last-modified: 2026-01-22`. It is entirely hand-built and has never been touched by the
framework: zero hits for "agritec" anywhere in this repo, and **no `sites` row** (checked against
all 45 rows, 2026-08-21).

That is the case the owner ruling of 2026-08-04 addresses. A hand-built site silently opts out of
every control the pipeline applies — evidence-base claims gating, banned-claim sweeps, the
discovery checks, imagery style, rerender. This site's condition is what that looks like after
seven months: unsourced figures throughout, four pages missing from their own indexes, a market
ticker fed by a random-number generator, and a data layer no page reads.

The goal is a full rebuild as a framework-managed site — tools and explainers decomposed into
managed components the improvement loop can evolve — back on `agritec.uk`, replacing what is
there, with richer imagery and infographics, and news/editorial/directory capability later.

## 1. Decisions (owner, 2026-08-21)

**D1 — Scope and phasing.** Agri/CEA cluster first (6 calculators + 6 explainers). The
IoT/machine-vision cluster (7 tools + 7 deep dives) is Phase 2.

**D2 — Everything migrates.** *"I would like eventually to have all the content from the existing
site into the new site."* Phasing is sequencing, not selection. `SUBJECT_LEDGER.md` carries a row
for all 26 live pages and is the mechanism that makes this checkable rather than hoped-for.

**D3 — Fresh Tier-3 submission, not crawl-adopt.** Both `082` paths end in a rebuild, but
`--from` seeds `identity`/`design_reference`/`design_intent` from the crawled site, which
anchors the new design to the old look — directly against the ask for more imagery and
infographics. The inventory `--from` would have fetched is already captured in the ledger by
hand-crawling the live site.

**D4 — Source everything before any page is written.** No page asserts a figure until it exists
as a registered fact with a source. *This is a rule about what may be published, not a promise
that every number will be found.* Where no citable primary source exists, the figure is not
published: it becomes a user input with no asserted default, or the page says plainly that we
have not verified it. "We have not verified this" is always publishable; a plausible guess never
is.

**D5 — Clean framework URLs, redirect map built after.** Take the framework's canonicalisation,
then read the real `pages.url` values and map the 26 old paths on. Redirects go in the Cloudflare
Worker — **not** as hand-written HTML stubs, which the 2026-08-04 ruling forbids.

**D6 — The copy is written afresh.** *"written afresh — it doesn't have to be the same copy."*
The old pages are a subject inventory and a depth floor, not a source text. This is the
platform's standing position (owner ruling 2026-08-06, the framework writes the content) and here
it is also the point: the existing copy is precisely what carries the unsourced numbers we are
removing. Porting it would import the problem.

**D7 — And it must come out deeper.** *"but to the same level of detail and greater."* Treated as
a measured floor per subject, not a judgement call — see §3.

**D8 — No cannabis content.** Owner ruling. Live today in a help-text hint under the DLI input of
`/tools/vertical-energy-calc.html` and in two rows of the publicly-served
`/data/crop-dli-table.json`. Both go at cutover. Enforced as a `banned_claims` pattern rather
than prompt wording, because the framework's writers could reintroduce it from general knowledge
of CEA crops without ever seeing the old page.

## 2. The two measurements this plan rests on

Both taken 2026-08-21. Method and re-run commands in `RUNBOOK_agritec_uk.md` §2.

**Measurement 1 — the existing depth.** Live explainers run 315–453 words (mean 377); deep dives
434–605 (mean 533). Technical density is real — correct units, genuine equations, specific
mechanisms — but the pages are short, and across all six explainers there is exactly **one**
diagram (`dli-vs-ppfd.svg`). That absence is what the request for imagery and infographics is
aimed at.

**Measurement 2 — what the framework produces at each page type.** Word counts of rendered
components across four mature framework sites:

| page_type | Sites | Pages | Avg words | Range |
|---|---|---|---|---|
| `blog-post` | 4 | 40 | 1,568–1,876 | 899–2,771 |
| `guide` | 1 | 5 | 511 | 483–537 |

**`page_type` alone decides whether a technical explainer lands at ~1,600 words or ~500.** The
old agritec explainers sit at the `guide` shape. Building them as `guide` would reproduce exactly
the depth we were asked to exceed. So: **explainers are `page_type='blog-post'`**, hubbed under a
`section-index` at `/guides/`.

> `[MEASURED, but note the sample]` The `guide` figure is one site, n=5. It is a strong signal,
> not a fleet-wide law. Re-measure before leaning on it again.

## 3. The depth contract

Per subject, the rebuilt page must meet or exceed the live page's measured **words, H2 count,
table count and equation count** (the floor columns in `SUBJECT_LEDGER.md`), and additionally
gain what the old one lacked:

- a sourced figure for every number, and
- at least one code-rendered infographic.

Target is ~1,600 words, not the floor. The floor is the failure line; the target is the job.

## 4. Phases

**Phase 0 — workstream + ledger.** This directory, the standing five, and `SUBJECT_LEDGER.md`.

**Phase 1 — seed before submitting.** `SEED_2026-08-21_agritec_site_and_specs.sql`, modelled on
`../oufe/SEED_2026-07-25_oufe_site_and_specs.sql`. Three aspects must exist before the first page
is written, for reasons recorded in that file's own header:
- `sites` row **with a real contact email** — the hallucinated-email check *fails open* on a site
  with no email, so a fabricated address can reach production.
- `evidence_base`, even with `facts: []` — the whole claims layer gates on this aspect's
  *presence*; `loadEvidenceBase` returns nil and every honesty lane silently no-ops without it
  (`validate_page_content.go:727-746`). Seed the agri `banned_claims` here, including D8.
- `imagery_style_guide` — `content_hero` generates unstyled art on any site that has none.

Never `UPDATE` a spec row in place: supersede (`is_current=false`) then insert, one transaction.
Write `content_direction` as one complete object, not incremental patches (`bugs_open/327`).

**No figures in any spec.** A number written into a spec is a *given*, and a given outranks every
anti-fabrication rule we have — a re-render once wrote invented spec numbers back over grounded
values while both the evidence register and the fleet prompt rule were live.

**Phase 2 — evidence first.** This is now the long pole, and deliberately so (D4). Six data
domains: SFI/ELMS payment rates (GOV.UK collection), UK energy prices and carbon intensity
(Ofgem, NESO), LED efficacy by fixture class (datasheets/DLC), crop DLI ranges (may not be
citably sourceable — likely becomes user input), BSF bioconversion rates (published studies),
seaweed dry-matter and carbon fraction (published studies). Run `evidence-researcher` per domain;
it web-searches, extracts atomic claims with verbatim quotes, re-fetches and re-matches the quote
before registering.

Two traps bite exactly here:
- **`bugs_open/161` — the register ratifies the claim it was built to catch.** The evidence base
  is simultaneously the whitelist the writer is instructed from and the authority every gate
  checks against. A false fact in it is self-ratifying: it causes the claim, then vouches for it,
  and no gate in the layer can object. A registered fact needs the source actually read.
- **`bugs_open/288` — the register guards COPY, not CODE.** A figure a calculator *encodes* is
  checked by nothing; an SDLT calculator ran an expired threshold for 16 months with a clean
  scanner throughout. **This is agritec's central risk**, because the SFI rates, LED efficacies
  and carbon fractions all live inside calculator JavaScript. Control: every encoded constant is
  also a registered fact **and** is asserted in the tool page's visible copy, where the scanner
  can see it. A constant that appears only in JS is ungoverned by design.

**Phase 3 — Tier-3 submission.** `082_submit_domain_unified.sh` has no `--roadmap-file` flag, so
hand-roll the envelope from `../oufe/TRIGGER_submit_tier3.sh`. Both briefs are `{"text": "..."}`
objects — a bare string renders `<no value>` and is silently ignored. The roadmap brief is
authoritative to the planner ("build only these pages, do not invent additional pages"), which is
how Phase 1 stays a deliberate 6+6 slice. `--fidelity` other than `locked` is recorded but wires
to nothing; do not use it to control scope. **Turn the news feed off deliberately after
classification** — the classifier will read this as agriculture and auto-seed generic farming-news
keywords, which spends credits per fetch and dilutes a calculator-led site.

**Phase 4 — the six calculators.** Via `tool-generator`/`create_tool_component`: self-contained
HTML with inline `<script>`, `component_level='tool'`, `function` prefixed `tool-`. JS **must** be
inline in `html_template` — the `js_content` lane publishes the file and injects no `<script>`
tag, published-but-inert (`bugs_open/041` class). Share the maths: `bugs_open/224` records one
site with seven private copies of the same formula, only one handling a zero-rate edge case;
several agritec calculators share conversions (PPFD↔DLI, EC↔ppm, mass balance). Check identity
agrees across `pages.name`, `content_components.function` and the acceptance doc's `subject_key`
(`bugs_open/311`).

> **ADDED 2026-08-21, from measurement (NOTES, misstep 5).** Put the honesty constraint **in the
> tool brief itself**, in the brief's own words. Do not delegate it to the evidence register.
> Measured across the four writing agents: `page-content-writer` and `tool-recreation-handler`
> both name `writer_block` in their prompts; **`tool-generator` and `tool-improver` name neither
> it nor evidence, fact, source or invent** — `tool-generator`'s 5,189-character template mentions
> none of them. They *receive* the register (their `load_brand_context` omits `aspect`, which is
> all-aspects mode, and `site_specs` is in `input_fields`), but nothing in the instruction points
> at it. Combined with `bugs_open/288`, the tool path has neither an instruction going in nor a
> check coming out, while prose has both. `tool-improver` — the agent that would *evolve* these
> calculators, which is the owner's actual goal — has the thinnest coverage of the four.

**Phase 5 — explainers, indexes, and the linking guarantee.** This is where we fix what the old
site got wrong, deliberately rather than by luck. Build the hubs with an explicit list-component
section in the plan **from day one** — re-typing pages does not populate a hub; a `guide-list`
section resolving against the page type has to be in the hub's own `sections` array, and a hub
with `sections=[]` renders nothing (`bugs_open/309` is a live case of an index rendering six
cards with zero anchors). Set `in_header`/`in_footer` explicitly on every page: URL shape may
decide *where* a page appears in nav, never *whether* it appears; a page with no nav flags at all
is still invisible, the still-open "A3" gap in `../bugfix_149_nav_membership/`.

**Phase 6 — imagery and infographics.** Planned into `site_plan_imagery`
(`kind` ∈ logo|hero|illustration|icon|**infographic**|sprite_sheet), styled by the seeded
`imagery_style_guide`. **Charts are code-rendered from registered facts — a diffusion model never
draws data.** `evidence-chart` resolves every point by `fact_id`; `evidence-timeseries` carries a
citation per observation; `mechanism-flow` has no numeric field at all, and the absence of the
slot is the control. The Go template funcmap has **no arithmetic**, so anything numeric is
pre-computed, never templated. `<svg><text>` is invisible to the claims scanner, so diagrams use
real HTML text with CSS furniture. Watch `bugs_open/214` and `bugs_open/114`: imagery gets
planned, generated, paid for, and never referenced.

**Phase 7 — cutover.** Commit → `git-adapter` → shared `sites` repo under `agritec.uk/` → Actions
runner → `b2 sync --delete` to `b2://portfolio-sites/agritec.uk/` → Cloudflare purge. `--delete`
means the old hand-built files are retired, not merely overwritten — which is also how the
cannabis JSON goes. DNS needs nothing; the zone is already ours. Confirm the `portfolio-sites`
Worker route is bound to `agritec.uk/*` and `*.agritec.uk/*` — a zone-name mismatch makes the
purge **silently** skip, so verify with a cache-busted `curl`, not a green Action. Then build the
redirect map from the real `pages.url` values (the estate has three competing tool-URL shapes
live, so read them, do not predict them). `bugs_open/315`: `pages.deployed_at` is stamped whether
or not the object write succeeded — verify at the served artefact.

**Phase 8 and later.** IoT cluster as ledger Phase 2. Then: **news** is a mature pipeline gated on
`classification.content_features.news_feed.recommended`; **editorial** is an assembly recipe
(`blog-post` feature + `section-index` hub), not new machinery; **directory listings** are the
real work — the global `directory_entities` registry has no agricultural kind, so a
supplier/product directory is new-vertical work, not a switch.

## 5. Council review

> **CORRECTED 2026-08-21, same day, before acting on it.** This section originally read *"The
> seed SQL is an appliable DB migration and therefore **in scope** for the council gate."*
> **It is not, and a session following that line would have had its submission refused.**
> Caught by reading `scripts/council-scope.sh` — the single source, which CLAUDE.md says not to
> re-derive — and then testing the path mechanically against the same `jq` predicate `097` runs:
>
>     COUNCIL_SCOPE_CODE_RE      ^(platform|internal|pkg)/
>     COUNCIL_SCOPE_MIGRATION_RE ^docs/agent_docs/sql_for_agents/[0-9]{3}_[A-Za-z0-9_]+\.sql$
>     -> false for docs/agent_docs/docs024_key_docs_latest/agritec_uk/SEED_*.sql
>
> The error was reasoning from the *widening* (migrations are in scope now) to *this file*,
> without checking that this file is a migration. It is not: it is per-site setup applied out of
> band with `psql -f`, exactly as the oufe seed was, and it is out of scope by path **and** by
> intent. What the widening covers is `sql_for_agents/NNN_name.sql` — the runner's own appliable
> shape — plus `_HOLD.sql`.

**So: nothing in this workstream is in council scope today.** The seed, the briefs, the ledger and
the site content are all out, and would be refused client-side without spending credits.

That changes when the lane touches platform code — and two of the ratchet line's re-check
conditions would do exactly that: extracting the depth-floor measurement or the reachability
crawl into a shared check, or adding an agricultural kind to the directory registry. At that
point submit via `097_TRIGGER_council_review_v1.sh` before or alongside the commit, use
`Council-Submitted: <corr>` if the verdict has not landed, and never write `Council-Reviewed:` on
a verdict not read.

**Test admission, don't infer it.** `DRY_RUN=1` on the trigger answers "would this be admitted?"
for free, and sourcing `scripts/council-scope.sh` lets you test a path in one line.

## 6. Open questions

None outstanding. D8 (cannabis) was raised and ruled the same day.
