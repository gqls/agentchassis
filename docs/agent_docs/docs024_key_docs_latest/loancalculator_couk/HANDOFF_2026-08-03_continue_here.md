# HANDOFF — loancalculator.co.uk · continue here (2026-08-03)

**This supersedes `HANDOFF_2026-08-02_continue_here.md`.** That file's §3 queue is
now item (1) done, items (2)–(4) untouched. Its §4 "what will mislead you" is all
still true and still worth reading — this file ADDS to it rather than replacing it.

Read order for a cold start: this file → `NOTES` tail (the 2026-08-03 section) →
`RUNBOOK` §§ "Changing a LOCKED calculator" **including the ⛔ correction inside
it**, "Chrome for an assembled site", "After a chassis roll".

---

## 1. State in one block

```
site            loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
pages           27, ALL rebuild_policy='generic'
component rows  63 total = 51 ported-prose + 12 tool, ZERO verbatim
locks           12 tool rows lock_type='permanent'  (verified 12 of 12 after the work)
                51 prose rows unlocked, deliberately
live            27/27 HTTP 200; 4 tool pages re-rendered 2026-08-03
calculators     12/12 match GOLDEN_2026-08-03_defects_fixed.json  ← NEW baseline
defect fixes    4 of 4 LIVE and proven on the SERVED pages
backups         page_components_bak_20260803_toolfix    (pre-fix tool rows)
                page_components_bak_20260803_backfill   (pre-backfill content_data)
                content_components_bak_20260802_decomp  (pre-fix templates)
                + the 08-02 decomposition backups, untouched
```

**The four queued calculator defects are fixed, live and verified.** The site is
still open to the improvement loop for text and closed for arithmetic.

## 2. Do this first

**a. Read `bugs_open/182` before you try to re-render anything on this site.**
`rerender_sections` is a **no-op here** and reports success. It resolves components
by `slot_name`; this site's slots are positional (`prose-0`, `tool-2`) and match
nothing. Measured: `rerendered: 0, carried: 4` with the fixes already live.

**Use `decompose/render_tool_row.py` + an assemble-only rerender instead** — the
route all 27 pages were originally shipped through. Full procedure in the RUNBOOK
under "✅ THE ROUTE THAT WORKS HERE".

**b. Check the 090 diagnosis verdict**, filed for the 182 mechanism:
```sql
SELECT current_step, status FROM orchestration_states
WHERE correlation_id='834a24b0-4d3e-4ce7-a17c-6b270493bfd6';
```
It was still on `verdict` when this was written. **If it REFUTES the mechanism,
correct `bugs_open/182` in place — do not delete it** — and note what caught it.

**c. If the scratchpad is gone, rebuild it** (unchanged from the 08-02 handoff):
```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/decompose
./prepare_work.sh /tmp/decomp-work        # ~2 min
export DECOMP_WORK=/tmp/decomp-work
```

## 3. The queue of real work, in the order I would take it

> **UPDATED later on 2026-08-03.** Item (2) is DONE and live. Item (1) was surveyed
> and is a different, larger job than it looked — read §3a before touching nav.

**(1) ⛔ Header nav — NOT a scripting job, and there is a demolition charge under
it.** See §3a. The precondition is declaring nav membership on 27 live rows, and
the orphan decision (item 3) is an INPUT to it. **Do not run `nav-updater`.**

**~~(2) A tidy owed on `tool-loan-repayment`.~~ DONE 2026-08-03, live and verified.**
It was not cosmetic: the HTML comment explaining why the homepage must not carry two
dated factual claims was **publishing both figures into the homepage's own source**
(HTML comments are not stripped; only the tool-doc sentinel block is). Moved inside
the sentinels. Served pages shrank — `/index.html` 27039→26338 b,
`/tools/standard-calc.html` 26389→25702 b — with `nobody asked to publish`,
`3.75% base rate` and `7.9% market average` all now **0** on both, and both tools
still MATCHING the golden.

**~~(3) `/tools/standard-calc.html` is an orphan.~~ RETIRED 2026-08-03 — owner
decision, verified on the wire: 404, sitemap 27→26, other 26 pages all 200.** It was
unreachable from the site yet in the sitemap at priority 0.8, self-canonical, no
noindex — and it was the page carrying two dated rate claims the homepage
deliberately omits. Backups: `pages_bak_20260803_orphan_retract`,
`page_components_bak_20260803_orphan_retract`. **No longer blocks item (1).**

**~~(4) EXCLUDE vs WITHHOLD on consolidation.~~ RULED 2026-08-03: keep WITHHOLDING**
— *"we need honesty at any cost."* No code change; the behaviour was already live
and proven. Do not re-open it.

## 3a. ⛔ NAV — read before touching item (1)

**This site is one `nav-updater` run away from a nav of roughly ONE link on all 27
pages, shipped immediately.** Measured 2026-08-03, not inferred:

```
pages                27  (13 guide, 12 tool, 1 content, 1 landing)
in_header = true      0
in_footer = true      0
declaring NEITHER    27  (explicitly false — NOT NULL, which is a third state)
site_nav_items        1  (against 25 links in the authored header)
```

`classifyPagesForNav` omits a page that is never-primary AND declares neither flag.
Every `tool` page is never-primary by type; every `/guides/` page by URL shape. That
is all 27. `populate_nav_tables` opens `DELETE FROM site_nav_items WHERE site_id=$1`,
and `nav-updater` then re-renders chrome and re-assembles every deployed page.

The fleet landmine for this was NARROWED on 2026-07-31 to "only a page declaring
neither flag is still lost" — which reads reassuring until you measure a site where
that is **every page**.

**Mitigation applied: the chrome is now LOCKED.** `head`/`header`/`footer`,
`lock_type='permanent'`, `locked_by='loancalculator_authored_chrome_20260803'`,
backup in `site_components_bak_20260803_chromelock`. Verified in order: the chrome
re-render honours the predicate (three write sites in
`render_site_components_action.go`), the lock bites here (`agent_may_write=f` ×3),
and the predicate DISCRIMINATES (45 chrome rows across 15 sites still read `t`).
Nothing on the wire moved.

That blocks the chrome overwrite, **not** the `site_nav_items` rebuild. And there is
no drift to fix today: 25 header links, **zero dead**, the only two pages absent from
the header being `/legal.html` (deliberately in the footer) and the orphan.

**The real item (1), in order:** declare `in_header`/`in_footer` to match the
authored chrome → decide the orphan (owner) → only then generate, via
`nav-link-fixer` (refreshes chrome from EXISTING tables, no populate step), never
`nav-updater` → unlock, load, re-lock. ⚠ `in_header` has a second consumer,
`buildServicesHTML` (`render_site_components_action.go:1156`), so the flag change
needs its own verification rather than a bulk UPDATE.

## 4. What the 2026-08-03 work added to "what will mislead you"

Everything in the 08-02 handoff §4 still stands. These are new, and the first
three are in the fleet `LANDMINES.md`.

- **A component's `input_schema` fallback is NEVER consulted at render time.** The
  render context is `base ⊕ page_components.content_data ⊕ resolved_data`.
  `RenderTemplate` resolves an unknown key to the empty string, and the `<no value>`
  marker that would have made it visible is **explicitly stripped**. So adding a
  schema field + a template placeholder is only TWO THIRDS of a change. Use
  `decompose/backfill_content_data.py --check` before any re-render.
- **`toolgolden.py` only ever drives NEIGHBOURHOODS OF THE SHIPPED DEFAULTS**
  (×1, ×2, ×0.5 of each field's own default). A boundary defect — 0% APR, a blank
  field — is unreachable, so a green `verify_rewrite.py` can be entirely silent
  about your fix. Two of these four fixes were invisible to it, measured both ways.
  `rewrite/defect_vectors.py --both` is the answer.
- **`rerender_sections` carries every section and reports success** — `bugs_open/182`.
- **`render_component`'s cache key omitted `rewrite_dir`** until 2026-08-03, so
  rendering one component from two directories in one process returned the first
  render twice and compared them EQUAL. Fixed. It was wrong exactly where a wrong
  answer reads "identical — nothing to do, your fix is already live".
- **`carryStoredSection` → `save_page_sections` is not byte-preserving**: it trims
  the trailing `\n` after `</script>` (8893 → 8892 on an untouched row). Harmless,
  but it will show up as a 1-byte drift in any byte comparison, and
  `render_tool_row.py` names that tolerance explicitly rather than comparing loosely.
- **The locks did not need lifting.** The 08-03 unlock was for the
  `rerender_sections` route that turned out not to work. The route that works writes
  by SQL — the deliberate act the lock exists to force, not automation — and
  assemble-only `render_page` never touches `page_components`.

## 4a. Retiring a page — what the platform does NOT do for you

`retract_page_deployment` is **live** in the chassis (pod-grep 6, positive control
20, negative control 0) but wired into **no agent**, so it is unreachable today, and
its own acceptance target still served 200 when checked. Do not assume it is a
usable route.

⚠ **Whatever route you use, the SITEMAP is a separate act on this site.** Its
`sitemap.xml` was last written by the **adoption commit** — the platform has never
regenerated it — and `retract_page_deployment` deliberately excludes `sitemap.xml`
as a file `pages` does not model. So retracting a page via the platform primitive
leaves the sitemap advertising a 404, which is the very defect `bugs_open/098`
exists to fix. Delete the file and its `<url>` block in one commit.

The order that worked: archive in the DB first (`pages.status='archived'`, so
nothing re-publishes it) → remove file + sitemap entry in one sites-repo commit →
verify at `origin/master`, **never at your local ref** (the push was rejected twice
by concurrent pushes while a naive echo reported success) → wait for the 404.

## 5. The one judgement call, flagged rather than buried

`tool-consolidation-risk`: a debt row with a balance but no usable rate or term now
**withholds the comparison** — the balance still shows (it is a fact, independent of
any rate), the three interest figures show `—`, and the verdict box says what is
missing.

The other candidate was to **exclude** the incomplete row from the balance too,
restoring like-for-like arithmetic. Rejected because it silently answers a different
question from the one the reader asked: a confident verdict computed over a subset
of their debts, with nothing on the page saying so. Withholding cannot state a wrong
comparison at all, which is the property worth having on a page about mis-selling.

**RULED 2026-08-03 by the owner: keep WITHHOLDING** — *"we need honesty at any
cost."* Settled; do not re-open. No code change was needed.

> **CORRECTION carried into the template.** The old note said the rate-less row made
> consolidation look **better** than it is. Backwards — and its own stated reason
> contradicted it in the same sentence. Omitting a debt's interest *understates* the
> old side, so the verdict is biased toward **WORSE**. Measured at the pre-fix sha:
> a £5,000 blank-rate debt produced *"Consolidating will cost you £1,359.36 MORE in
> total interest"* off an old side of £0.00.

## 6. How the four fixes were proven, so you can re-run it

Two gates, asking opposite questions. **Neither alone is sufficient** and that is
the whole lesson of this session.

```bash
cd /home/ant/projects/agentchassis
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk

python3 $LANE/rewrite/defect_vectors.py --both   # did the thing that SHOULD move, move?
python3 $LANE/rewrite/defect_vectors.py --live   # ...on the SERVED pages
python3 $LANE/rewrite/verify_rewrite.py          # did anything move that should NOT have?
```

Recorded results, 2026-08-03:

- `--both`: **4 PROVEN, 3 CONTROL, 0 vacuous.** Each defect case reads differently
  against the pinned pre-fix sha `6e8098022`; each control is unmoved.
- `--live`: **8 of 8 pass against production.**
- `verify_rewrite`: **9 of 11 MATCH**; the two divergences are only `save-display`
  (NUMBER) and four loan-vs-savings `text/display` keys, with `loan-benefit` and
  `save-benefit` **unmoved** — so no arithmetic moved anywhere on the site.
- Live `toolgolden --compare`: same shape, 10 of 12 MATCH.
- Per served page, a positive AND a negative control: the string the fix ADDED is
  present, the string it REMOVED returns 0.

⚠ **Keep `GOLDEN_2026-07-31c`** — still the only record of what the HAND-BUILT site
computed. `GOLDEN_2026-08-02_decomposed` is the pre-fix decomposed baseline.
`GOLDEN_2026-08-03_defects_fixed` is the forward one.

## 7. Still with the owner (unchanged from 08-02)

- **`GITHUB_READ_TOKEN` cannot see `gqls/sites`** — fine-grained PAT, 404 while
  authenticated. Needs GitHub admin. Off the critical path.
- **Whether to create a site PLAN.** Zero `site_plans` rows, so the reconciler has
  nothing to act on. A plan is what would let the loop add and reshape pages rather
  than only improve existing ones. His call.
- **NEW: whether `bugs_open/182` gets fixed at the platform.** Until it does, this
  site — and 5 others, partially — cannot receive a template-driven re-render, and
  every tool change here must go the offline route.

## 8. Commits (this session, `docs/` + one `bugs_open/` file — no platform code)

`b819b2af6` the four fixes + defect_vectors, proven ·
`31188f4f1` pin the defect baseline to a sha ·
`c11c8b6f9` backfill_content_data.py ·
`06362ad2a` NOTES + RUNBOOK + two fleet landmines ·
`22cbadf24` README entry ·
`23acb5e3f` render_tool_row.py + the render_component cache fix ·
`08b716e98` shipped + verified live + `bugs_open/182` ·
`46d93e15c` RUNBOOK correction + the two missteps

**No `Council-Reviewed:` trailers, and none owed** — every file is under `docs/` or
`bugs_open/`, which the gate refuses client-side. Note the 08-02 handoff's warning:
two of ITS commits carry `Council-Submitted: pending-…`, a placeholder and a
mistake. None of this session's do.
