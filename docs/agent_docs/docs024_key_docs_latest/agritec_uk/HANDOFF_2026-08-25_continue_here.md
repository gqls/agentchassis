# agritec.uk — continue here (2026-08-25)

Cold start for a fresh session. Read this, then `SUBJECT_LEDGER.md`, then `NOTES_agritec_uk.md`
from the bottom up. Everything below was verified at the artefact on the date given, not inferred.

---

## 1. What this lane is

Rebuilding `agritec.uk` — a hand-built static site of calculators and technical articles — inside
the framework, so tools and guides become managed components the improvement loop can evolve.
Owner ruling 2026-08-04 requires it: a hand-built site opts out of evidence gating, banned-claim
sweeps, discovery checks, imagery style and rerender, and this one had visibly suffered all five.

**Workstream directory:** `docs/agent_docs/docs024_key_docs_latest/agritec_uk/`

| file | what it is |
|---|---|
| `PLAN_2026-08-21_agritec_uk.md` | decisions D1–D8 and the phase plan |
| `SUBJECT_LEDGER.md` | **the completeness contract** — all 26 live pages, depth floors, gate results |
| `NOTES_agritec_uk.md` | technical log, append-only, **includes nine recorded missteps** |
| `README_where_we_are.md` | the owner's plain-prose log |
| `SUMMARY_2026-08-24_agritec_uk.md` | milestone read-out |
| `TOOL_SPEC_sfi_stacker.md` | the SFI calculator specification |
| `SEED_*.sql` | fourteen applied migrations, each with its reasoning in the header |
| `install_fence_facts.py`, `fix_fence_selectors.py` | lane installers (dry run by default) |

## 2. State as of 2026-08-25, measured

**Live and correct.** 13 pages, all `deployed`. The old hand-built site is **gone** —
`/tools/elms-calculator.html` returns 404, retired by `b2 sync --delete` at cutover.

- Six explainers at **1,400–1,803 words** against the retired site's 315–453 (**depth gate: passed**)
- **Zero orphans** — all six listed on `/guides/index.html` (**reachability gate: passed**)
- **104 evidence facts**, every one carrying a clickable source URL and a verbatim quote
- The **SFI26 Revenue Stacker** is built, deployed and correct at
  `/tools/sfi26-revenue-stacker/index.html`

**The SFI calculator's guard rails, verified by the `bugs_open/288` lane on 2026-08-25:**

    PLAN fence declarations    24 of 24 file, zero fact_declaration_broken
    artifact_check fences       4 of 4 fresh
    citation arm              104 of 104 fresh, zero errors, zero drifted

## 3. ⚠ THE ONE OPEN ITEM — the light palette is 90% in; the CSS is the last step

**Owner instruction, 2026-08-25:** a lighter palette with dark text, as the default on all domains
and on this one.

### Done and verified

| | |
|---|---|
| `design_intent.palette.reference_values` | `#F7F8F5` bg / `#1A202C` text ✓ |
| `design_intent.style_direction` | `modern-light` ✓ |
| `classification.suggested_style` | `modern-light` ✓ (was `professional-dark`) |
| `imagery_style_guide` | moved to light grounds ✓ |
| **the site's theme** | **replaced** — `site-design-planner` installed `collection-agritec-uk-819c6c8c` at 15:55, light, and repointed `sites.style_collection_id` to it ✓ |
| fleet default | migration **613** applied + verified; council `ca1d0f70-602d-4908-9098-632fc89bdb61` ✓ |

### The one thing left

**The `head` slot of `site_components` is still stamped 2026-08-24 19:20 and carries `#12151F`.**
It is the last dark artefact and it is what the pages embed.

A `needs_design:regen-css-on-light-theme` item is **queued now** for `webdesign-agent`. The
earlier attempt re-rendered dark because `render_css_from_spec` merges `comp.Palette` (the
THEME's) with the spec's, and the theme was still dark. **The theme is light now**, so this run
should produce light CSS. If it does not, read the merge in
`render_css_from_spec_action.go:125` and `enforceLayoutScheme` at `:390` — the LAYOUT also
declares a scheme and the layout may still be a dark variant.

**Verify at the served page, never at the specs.** The specs have read "light" for hours while the
site served dark; that gap is the whole reason this took four seams:

    curl -sS -L "https://agritec.uk/tools/sfi26-revenue-stacker/?cb=$(date +%s)" | grep -c '#12151F'

Must return 0.

**Then:** the 17 generated images were made against the *dark* imagery guide and will look wrong
on a light site. They need regenerating once the CSS lands.

### The seam map — FIVE seams, and four of them report success while changing nothing

This is the most useful thing in this file. Getting a palette onto a live page touches five
distinct mechanisms, and **only one of them ever refuses**.

| want | seam | reality |
|---|---|---|
| regenerate CSS into the theme | `needs_design` → `webdesign-agent` | reads the **THEME's** palette, not just the spec — useless while the theme is stale. **And it never writes the chrome at all.** Completed 3× changing nothing |
| replace the site's theme / composition | `needs_composition` → `site-design-planner` | **the only owner of installation**. Needs `allow_reinstall: true` if a collection already exists — **the one seam that refuses, and it names its own remedy** |
| contribute a theme to the library | `fork_theme` | **NOT this.** Its header: the site's `style_collection_id` is *not* modified |
| write the chrome (head/header/footer) | **`rerender-chrome`**, dispatched directly by Kafka message — no work item type routes to it | `render_site_components` is the only writer of `site_components` |
| get it onto the served page | `page_rerender` per page | stored components can be clean while the deployed file is stale |

**The lesson worth carrying past this site:** four of the five completed successfully while
changing nothing that mattered. A silent success cost three wrong seams; the single explicit
refusal was solved in thirty seconds. When a chain of agents all report success and the artefact
does not move, stop trusting statuses and diff the artefacts at each hop.

## 4. Other open items, in priority order

1. **The companion-guide stubs.** The tool pipeline auto-created
   `/guides/tool-sfi26-revenue-stacker-guide.html` — "Understanding the SFI26 Revenue Stacker" —
   beside the real subject-led explainer. Reconcile deliberately: retire the stub or repoint the
   tool's CTA. **Do not solve it by writing explainers as companion guides.**
2. **`/tools.html` hub.** The tools index exists as `page_type='content'` at `/tools.html`;
   `/tools/index.html` is a 404. Check it lists the calculator.
3. **Phase 2 of the ledger:** the IoT/machine-vision cluster — 7 tools and a 7-part engineering
   series on distributed optical crop monitoring. Owner: everything migrates eventually.
4. **Remaining Phase 1 tools:** five agri calculators beyond the SFI one (vertical energy, VPD,
   nutrient dosing, BSF converter, blue carbon). Evidence largely registered; specs not written.
5. **Later:** news, editorial, directory. Directory is genuinely new work — the global
   `directory_entities` registry has no agricultural kind.

## 5. ⚠ DO NOT TIDY THESE — they are working machinery

- **24 `unreconciled_declaration` items** are expected, low severity, handler-less and
  self-quieting. Each records the value that becomes its own baseline.
- **The `fact_binding_suggested` note disappearing** from future sweeps is the suggester skipping
  tools that declare. Its absence is the mechanism working, not a regression.
- **`misplaced_artifact_checks` is absent from the payload**, not empty, because the reporter is
  not yet in a running binary. An absent key can only mean "the code isn't running"; an empty
  array would be ambiguous. Not evidence of anything yet.
- **Seven `sitemap_entry_dead_live` items** describe the *retired* site and self-clear.
- The **nine equal-value fact pairs** (UPL10/CNUM2 at £102, etc.) are **different actions sharing
  a rate**, not duplicates. Collapsing them destroys real facts.

## 6. Traps this lane paid for — read before touching these things

- **`pages.sections` is DERIVED.** The authority is `site_plan_sections`. Editing the projection
  appears to work and is silently reverted by the next rebuild. **When a fix is reverted rather
  than rejected, you edited a projection.**
- **Rerender ASSEMBLES; rebuild RE-RESOLVES.** A listing component caches its resolved item set in
  `content_data`. No number of rerenders will pick up a page that deployed later.
- **A `complete` work item is not a built artefact.** `tool-generator` reported complete with
  `error` NULL while the orchestration ended at `complete_error`; the truth was in
  `collected_data.__step_error` (it had truncated at 32,000 output tokens).
- **`artifact_check` goes INSIDE `source`.** At the top level of the fact it is live in the
  register and invisible to the mechanism — silently, for ever.
- **A citation with no `quote` can never be re-proved, and never escalates** (`citationDateStale`
  returns false when `staleness_days` is unset). 83 facts sat like that for a day.
- **Compose a quote from the EXTRACTOR's output, not the rendered page.** Two failed on a stray
  space before a closing bracket. Call `datahelpers.VisibleTextFromHTML` and `QuoteFoundInText`
  directly rather than re-implementing extraction.
- **Test a banned-claim pattern on BOTH arms.** Two of mine blocked pages from building — one
  matched the citation "Carbon Brief, May 2025", the other the honest past-tense sentence it was
  written to permit.
- **The acceptance fence must name INSTANCE-SCOPED ids** (`c-tool-<function>-<id>`). The
  auto-generated fence named unscoped ones and could never pass.
- **No backticks in a `git commit -m` message** — they execute. Cost a word twice in one session.

## 7. The `bugs_open/288` collaboration

Live and productive. That lane's mechanism reads a tool's *code* for registered values; ours is
the first site where a tool and a register were known to disagree, so we are their live proof.
Contact them via `SendMessage` to `bugs_open/288`.

Standing facts from that exchange: their probe has a **measured distinctiveness floor of 1000**,
so **69 of our 72 SFI rates are refused by it** and read `not_probed` — that is honest and
expected. Our four `artifact_check` fences are the per-fact control that reaches below the floor.
Their `misplaced_artifact_checks` reporter is committed but unrolled.

## 8. Owner decisions on record

- **D1–D8** are in `PLAN_2026-08-21_agritec_uk.md` §1. Most consequential: **source everything
  before any page is written**; **the copy is written afresh** (never ported); **all content
  migrates eventually**; **no cannabis content**.
- **2026-08-24:** cite every figure with a visible link. Note the framing deliberately avoids
  overclaiming — *we cite everything so you can check us, not so you can stop checking* — on the
  oufe precedent that a citation proves provenance, not correctness.
- **2026-08-25:** lighter palette with dark text, as the default on all domains.
