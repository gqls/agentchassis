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

**(1) The header's link list is hand-maintained.** 25 links lifted verbatim from
`nav.js`. A page added to `pages` does not appear; a page removed leaves a dead
link. Generating `site_components.header` from `pages` is the obvious next
mechanism, deliberately not bundled with the decomposition so that a regression
would not have two candidate causes. **Still the right next thing.**

**(2) A tidy owed on `tool-loan-repayment`.** Its template carries an HTML comment
that ships in the public source of `/index.html` and `/tools/standard-calc.html`,
and one clause of it editorialises. Not done on 08-02 because `index` was
mid-render; not done on 08-03 because it was out of scope for the defect fixes.
**It is now the cheapest thing on this list** — `render_tool_row.py` makes it one
`--apply` and one assemble-only rerender.

**(3) `/tools/standard-calc.html` is an orphan** — nothing links to it, and it
duplicates the index calculator. A content question for the owner, not a bug.

**(4) Consider whether the consolidation fix should EXCLUDE rather than WITHHOLD.**
See §5. This is the owner's call and it is small either way.

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

**Small change either way if the owner prefers exclusion.**

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
