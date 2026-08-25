# agritec.uk — continue here (updated 2026-08-25, end of day)

Cold start for a fresh session. Read this, then `SUBJECT_LEDGER.md`, then `NOTES_agritec_uk.md`
from the bottom. Every figure below was measured on 2026-08-25, not remembered.

---

## 1. What this lane is

Rebuilding `agritec.uk` — a hand-built static site of calculators and technical articles — inside
the framework, so tools and guides become managed components the improvement loop can evolve.
Required by the owner ruling of 2026-08-04: a hand-built site opts out of evidence gating,
banned-claim sweeps, discovery checks, imagery style and rerender, and this one had visibly
suffered all five.

**Directory:** `docs/agent_docs/docs024_key_docs_latest/agritec_uk/`

| file | what it is |
|---|---|
| `SUBJECT_LEDGER.md` | **the completeness contract** — all 26 original pages, depth floors, gate results |
| `PLAN_2026-08-21_agritec_uk.md` | decisions D1–D8 and the phase plan |
| `NOTES_agritec_uk.md` | technical log, append-only, **eleven recorded missteps** |
| `README_where_we_are.md` | the owner's plain-prose log |
| `SUMMARY_2026-08-24_agritec_uk.md` | milestone read-out |
| `TOOL_SPEC_sfi_stacker.md` | the SFI calculator specification |
| `RUNBOOK_agritec_uk.md` | commands, each with its gotcha — **§9 read an evidence run, §14 a stalled work item** |
| `SEED_*.sql` ×17 | applied migrations, reasoning in each header |
| `install_fence_facts.py`, `fix_fence_selectors.py` | lane installers (dry run by default) |

## 2. State — measured 2026-08-25

    pages 13 | deployed 13 | evidence facts 104 | specs 13 | assets 17 | work in flight 0

**The old hand-built site is gone.** `/tools/elms-calculator.html` → 404, retired by
`b2 sync --delete` at cutover. The abolished-subsidy calculator no longer serves.

**Both quality gates pass:**
- **Depth** — six explainers at 1,400–1,803 words against the retired site's 315–453
- **Reachability** — 6 of 6 listed on `/guides/index.html`, zero orphans

**The SFI26 Revenue Stacker** is live at `/tools/sfi26-revenue-stacker/index.html`, correct on
every figure the retired one had wrong (£224 not £382; no management payment; CHRW2 at £13 *for
one side*), with all four scheme caps and three action-level limits modelled.

**Its guard rails, verified by the `bugs_open/288` lane:**

    PLAN fence declarations   24 of 24 file, zero fact_declaration_broken
    artifact_check fences      4 of 4 fresh
    citation arm             104 of 104 fresh, zero errors, zero drifted

**The light palette is DONE**, verified at the served page on both arms (a "no dark hex" check
alone is equally true of a broken page):

    served 63,883 bytes | #12151F absent | cream/paper/dark-text/brand-green present
    tool intact: rate table, action codes, script, source link, capture date, £224 present, £382 absent

Site: four specs light, theme replaced (`collection-agritec-uk-819c6c8c`), chrome regenerated, all
13 pages re-rendered. Fleet: migration **613** applied and verified, council
`ca1d0f70-602d-4908-9098-632fc89bdb61`.

## 3. Open items, in priority order

1. **Regenerate the 17 images.** They were generated against the *dark* imagery guide and will
   look wrong on the light site. `imagery_style_guide` is already light; the assets are not.
2. **Run a Tier-4 acceptance pass on the calculator — it has NEVER run here** (zero
   `acceptance_run` items for this site, ever). The fence's selectors were wrong until today, so
   it could not have passed before. The tool is verified to *contain* the right figures; **nothing
   has verified that its arithmetic works.** Route: an `acceptance_run` item →
   `tool-acceptance-agent`.
3. **`claims_unverified` (1, `needs_human_review`, raised 2026-08-24 21:05)** — "unsupported prose
   assertions need a human ruling". This is precisely this lane's subject and deserves reading.
4. **`unresolved_cta` (9)** — CTAs with no real destination. Several will resolve as the remaining
   tools get built; check before treating any as a defect.
5. **The companion-guide stubs.** The tool pipeline auto-created "Understanding the SFI26 Revenue
   Stacker" beside the real subject-led explainer. Retire the stub or repoint the tool's CTA —
   **do not solve it by writing explainers as companion guides.**
6. **The five remaining Phase 1 calculators** (vertical energy, VPD, nutrient dosing, BSF
   converter, blue carbon). Evidence largely registered; specs not written.
7. **Phase 2 of the ledger:** the IoT cluster — 7 tools and a 7-part engineering series. Owner:
   everything migrates eventually.
8. **Later:** news, editorial, directory. Directory is genuinely new work — the global
   `directory_entities` registry has no agricultural kind.

## 4. ⚠ DO NOT TIDY — these look like backlog and are working machinery

- **24 `unreconciled_declaration` items** — expected, low severity, handler-less, self-quieting.
  Each records the value that becomes its own baseline.
- **`fact_binding_suggested` disappearing** from future sweeps is the suggester skipping tools
  that declare. Its absence is the mechanism working.
- **`misplaced_artifact_checks` absent from the payload** (not empty) because the reporter is not
  in a running binary. An absent key can only mean "the code isn't running"; empty would be
  ambiguous. Not evidence of anything yet.
- **7 `sitemap_entry_dead_live`** describe the *retired* site and self-clear.
- **The nine equal-value fact pairs** (UPL10/CNUM2 at £102, etc.) are **different actions sharing
  a rate**, not duplicates. Collapsing them destroys real facts.
- **1 `chrome_divergence_overwritten`** from 2026-08-24 19:20 — the original build overwriting a
  hand-patched head, archived to `site_component_history`. Predates the palette work.

## 5. The five-seam palette map — the most reusable thing here

Getting a palette onto a live page touches five mechanisms, and **only one ever refuses.**

| want | seam | reality |
|---|---|---|
| regenerate CSS into the theme | `needs_design` → `webdesign-agent` | reads the **THEME's** palette, not just the spec — useless while the theme is stale. **Never writes the chrome.** Completed 3× changing nothing |
| replace the theme / composition | `needs_composition` → `site-design-planner` | **the only owner of installation.** Needs `allow_reinstall: true` if a collection exists — **the one seam that refuses, and it names its own remedy** |
| contribute a theme to the library | `fork_theme` | **NOT this.** Its header: the site's `style_collection_id` is *not* modified |
| write the chrome | **`rerender-chrome`**, dispatched by Kafka message — no work item routes to it | `render_site_components` is the only writer of `site_components` |
| get it onto the served page | `page_rerender` per page | stored components can be clean while the deployed file is stale |

**Four of five completed successfully while changing nothing that mattered.** A silent success
cost three wrong seams; the single explicit refusal was solved in thirty seconds. **When a chain
of agents all report success and the artefact does not move, stop trusting statuses and diff the
artefacts at each hop.**

## 6. Traps this lane paid for

- **`pages.sections` is DERIVED** — the authority is `site_plan_sections`. **When a fix is
  reverted rather than rejected, you edited a projection.**
- **Rerender ASSEMBLES; rebuild RE-RESOLVES.** A listing component caches its resolved item set in
  `content_data`; no number of rerenders picks up a later page.
- **A `complete` work item is not a built artefact.** `tool-generator` reported complete with
  `error` NULL while the orchestration ended at `complete_error` — the truth was in
  `collected_data.__step_error` (truncated at 32,000 output tokens).
- **`artifact_check` goes INSIDE `source`.** At the top level it is live in the register and
  invisible to the mechanism, for ever.
- **A citation with no `quote` can never be re-proved and never escalates** (`citationDateStale`
  returns false when `staleness_days` is unset). 83 facts sat like that for a day.
- **Compose a quote from the EXTRACTOR's output, not the rendered page** — call
  `datahelpers.VisibleTextFromHTML` and `QuoteFoundInText` directly.
- **Test a banned-claim pattern on BOTH arms.** Two of mine blocked pages from building.
- **The acceptance fence must name INSTANCE-SCOPED ids** (`c-tool-<function>-<id>`).
- **A stalled item is usually a retry backoff** — see `RUNBOOK` §14. Ask the row what blocks it;
  do not re-type the loader's predicate, because a copy missing a clause can only say "eligible".
- **No backticks in `git commit -m`** — they execute. Cost a word twice in one session.

## 7. The `bugs_open/288` collaboration

Live and productive; contact via `SendMessage` to `bugs_open/288`. Their mechanism reads a tool's
*code* for registered values, and this is the first site where a tool and a register were known to
disagree, so we are their live proof. Standing facts: their probe has a **measured distinctiveness
floor of 1000**, so 69 of our 72 SFI rates are refused and read `not_probed` — honest and
expected. Our four `artifact_check` fences reach below that floor.

## 8. Owner decisions on record

- **D1–D8** in `PLAN_2026-08-21_agritec_uk.md` §1. Most consequential: **source everything before
  any page is written**; **the copy is written afresh**, never ported; **all content migrates
  eventually**; **no cannabis content**.
- **2026-08-24:** cite every figure with a visible link. The framing deliberately avoids
  overclaiming — *we cite everything so you can check us, not so you can stop checking* — on the
  oufe precedent that a citation proves provenance, not correctness.
- **2026-08-25:** lighter palette with dark text, as the default on all domains. Done both.
