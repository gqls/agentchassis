# HANDOFF — loancalculator.co.uk · continue here (2026-08-04)

**Supersedes `HANDOFF_2026-08-03_continue_here.md`.** That file's §4 ("what will
mislead you") and §4a (retiring a page) are still accurate and NOT repeated here in
full — read them. This file carries current state, the open items, and everything
learned since.

Read order for a cold start: this file → `HANDOFF_2026-08-03` §4/§4a → `NOTES` tail
→ `RUNBOOK` §§ "Changing a LOCKED calculator" (including the ⛔ correction and the
COMMIT-BEFORE-APPLY rule), "⛔ NAV: do NOT run nav-updater on this site".

---

## 1. State in one block

```
site            loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
pages           26 active (was 27 — /tools/standard-calc.html RETIRED 08-03)
component rows  57 = 46 ported-prose + 11 tool   (6 rows went with the retired page)
locks           11/11 tool rows   permanent, locked_by=decompose_20260802_proven_calculators
                3/3  site_components (head/header/footer) permanent,
                     locked_by=loancalculator_authored_chrome_20260803
                46 prose rows unlocked, deliberately
live            26/26 HTTP 200 · retired page 404 · sitemap 26 entries
calculators     11/11 match GOLDEN_2026-08-03b_after_orphan_retired.json
chassis         v1.0.1250, rolled 2026-08-04T10:29Z
```

**Everything the owner asked for is done.** Four calculator defects fixed and live;
the orphan retired; withholding ruled and kept. No work is half-finished.

## 2. Open items, in the order I would take them

**(1) The header nav — blocked on a decision, not on effort.** See
`HANDOFF_2026-08-03` §3a, still accurate. In short: **do not run `nav-updater`** —
all 26 pages declare neither `in_header` nor `in_footer` (explicitly `false`, not
NULL), every tool page is never-primary by type and every `/guides/` page by URL
shape, so `classifyPagesForNav` omits all of them and `populate_nav_tables` opens
with a `DELETE`. The chrome lock blocks the second half of that damage, not the
first. The precondition is declaring `in_header`/`in_footer` to match the authored
chrome — **which is a write to 26 live rows and `in_header` has a second consumer
(`buildServicesHTML`, `render_site_components_action.go:1156`)**, so it wants its
own change and its own verification.

**(2) The footer's corrected comment has not propagated.** `site_components.footer`
is correct in the DB (2504 b, engineering prose removed). The 26 live pages still
serve the OLD footer until each next re-renders. Self-heals; 26 forced deploys for
an invisible comment was judged disproportionate. If you re-render pages for any
other reason, this rides along.

**(3) `bugs_open/182` is owned by another thread** (owner said so 08-03). Do not
work it. It is why `rerender_sections` is a no-op here — read it before trying to
re-render anything by the ordinary route.

**(4) Nothing else is owed.** The four defects, the tidy, the retirement and the
re-baseline are all closed and verified.

## 3. The one live procedure you must not get wrong

**`rerender_sections` DOES NOTHING on this site and reports success** (`bugs_open/182`).
Slots are positional (`prose-0`, `tool-2`); the component lookup keys on
`slot_name`; nothing resolves; every section is carried. Measured `rerendered: 0,
carried: 4` with fixes already live.

**The route that works** — and the one all pages shipped through:

```bash
cd /home/ant/projects/agentchassis
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk

# 0. COMMIT THE TEMPLATE FIRST. The apply tools write the DB from the FILE and
#    cannot see it is uncommitted. Missed on 08-03 — see §5.
python3 $LANE/rewrite/verify_rewrite.py            # nothing moved that should not have
python3 $LANE/rewrite/defect_vectors.py --both     # the thing that SHOULD move, did
python3 $LANE/decompose/update_component.py --check|--apply <function>
python3 $LANE/decompose/backfill_content_data.py --check|--apply <function>   # ⛔ see §5
python3 $LANE/decompose/render_tool_row.py --check|--apply <function>
#    then an ASSEMBLE-ONLY page_rerender (spec with NO `reason`), page_id in the
#    spec AND the column, status 'triaged' (never 'detected')
python3 $LANE/rewrite/defect_vectors.py --live     # drives the SERVED pages
```

`render_tool_row.py --check` runs a CONTROL — re-renders from a baseline ref and
requires the CURRENTLY STORED bytes back — and refuses to write if that fails.

## 4. Verification assets, and which is which

| file | what it is |
|---|---|
| `GOLDEN_2026-07-31c_tool_values.json` | **the evidence.** Only record of what the HAND-BUILT site computed. Never delete. |
| `GOLDEN_2026-08-02_decomposed.json` | the decomposed site pre-defect-fixes |
| `GOLDEN_2026-08-03_defects_fixed.json` | post-fixes, **now stale** — names the retired page, will fail |
| `GOLDEN_2026-08-03b_after_orphan_retired.json` | **the forward baseline.** 11 tools, self-verified 11/11 |

- `rewrite/verify_rewrite.py` — equivalence gate. Baseline PINNED to sites-repo ref
  `b4302e22b` via `git archive` (§5).
- `rewrite/defect_vectors.py` — drives the defect CONDITIONS the golden vectors
  cannot reach. `--both` scores PROVEN / CONTROL / VACUOUS against `PRE_FIX_REF =
  6e8098022`. `--live` drives production.
- `decompose/render_tool_row.py`, `decompose/backfill_content_data.py` — added 08-03.

## 5. What was learned the hard way — the short list

Full write-ups in `NOTES`; the fleet-wide ones are in `LANDMINES.md`.

1. **A new schema field renders EMPTY.** The render context is
   `base ⊕ content_data ⊕ resolved_data`; `input_schema`'s `fallback` is **never
   consulted**, and the `<no value>` marker is explicitly stripped. Schema +
   template is two thirds of a change. Run `backfill_content_data.py --check`.
2. **`toolgolden` only visits neighbourhoods of the shipped DEFAULTS** (×1, ×2,
   ×0.5). A boundary defect — 0% APR, a blank field — is unreachable, so a green
   gate can be silent about your fix. Two of four fixes were invisible to it.
3. **A baseline that names a MOVING thing stops being a control, silently.** Twice
   in one day: `git show HEAD:` stopped being "before" the moment I committed; and
   `verify_rewrite` read `~/projects/sites`, **the checkout the platform's own
   deploys write into**, so each decomposed page replaced its own baseline. Both
   now pinned to absolute shas.
4. **"Live" and "committed" are independent facts.** `tool-loan-repayment` ran in
   production most of a day while HEAD held the previous template. The apply tools
   write from the FILE and cannot see it is uncommitted. **Sweep file-vs-live before
   calling a session done** — the no-argument script is in the RUNBOOK.
5. **A `LIKE '%…%'` probe cannot tell markup from a comment.** My inbound-link audit
   found two links that were both comments, one written by me three hours earlier.
   Use `href="<url>"`, across bodies + chrome + nav, with a positive control.
6. **Verify a push at `origin/master`, never at your local ref.** `gqls/sites` takes
   concurrent pushes; mine was rejected twice while my own command printed success.
7. **Authored chrome has no tool-doc sentinel** — anything written in it is
   published on every page. The footer's engineering note shipped site-wide and went
   FALSE the moment the orphan was retired.

## 6. Post-roll check for v1.0.1250 (2026-08-04) — IN FLIGHT AT HANDOFF TIME

The chassis rolled to **v1.0.1250 at 2026-08-04T10:29Z** (previous pods ~08-03 21:00).

- Site verified healthy immediately after: **26/26 HTTP 200**, retired page 404.
- **The render-path diff is NOT empty**, so this roll could not be waved through:
  `12ae5824f fix(187)` touches `rerender_page_sections_action.go`. Reading it, the
  change is confined to the **escalation** branch (it guards a false-alarm
  `needs_page` and returns a disposition) and touches neither assembly nor template
  rendering — and it cites `bugs_open/182` as its reason for naming a no-op. So it
  should not move bytes here. **Expected is not measured**, hence the check below.
- Baseline captured from the OLD image's output, then three pages re-rendered on the
  new one (`created_by='postroll-1250'`):

```
index.html                  bd4ea8c5ba39   26338 b
tools/consolidation.html    b48c4a05cabb   34521 b
tools/loan-vs-savings.html  06f9347699d5   27623 b
```

**TO FINISH THIS CHECK:**

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c \
 "SELECT spec->>'page_name'||' '||status FROM site_work_items WHERE created_by='postroll-1250';"
# when all three are complete (allow for deploy propagation, ~2 min behind 'complete'):
for u in index.html tools/consolidation.html tools/loan-vs-savings.html; do
  n=$(echo "$u" | tr '/' '_')
  printf '%-30s %s\n' "$u" \
    "$(curl -s https://loancalculator.co.uk/$u | md5sum | cut -c1-12) vs $(md5sum /tmp/decomp-work/postroll/$n | cut -c1-12)"
done
```

**Identical md5s = the new image renders these pages the same, measured.** If one
differs and the others match, **re-run before believing it** — propagation lag is
per-batch and clears; a real fault is per-page-shape and reproducible.

Then re-confirm the calculators:

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 $LANE/toolgolden.py --compare $LANE/acceptance/GOLDEN_2026-08-03b_after_orphan_retired.json \
  https://loancalculator.co.uk/index.html \
  https://loancalculator.co.uk/tools/{overpayment-calculator,consolidation,compare-loans,car-finance-calculator,loan-vs-savings,settlement-calculator,interest-rate-stress-test,credit-health-check,damage-checker,application-tracker}.html
# PASS = all 11 tools reproduce their golden values exactly
```

⚠ If `/tmp/decomp-work/postroll/` is gone (it does not survive a session), the
baseline is gone with it and this particular check cannot be completed — capture a
fresh baseline and re-render a page again. Do not substitute "the pages look fine";
that proves the site is unchanged, not that the new binary renders the same.

## 7. Backups taken 2026-08-03 (all still present)

```
page_components_bak_20260803_toolfix        pre-fix tool rows
page_components_bak_20260803_backfill       pre-backfill content_data
page_components_bak_20260803_orphan_retract the retired page's 6 rows
pages_bak_20260803_orphan_retract           the retired page row
site_components_bak_20260803_chromelock     chrome before the lock
content_components_bak_20260802_decomp      pre-fix templates
```

Restore a retired page: re-insert from the two `orphan_retract` tables, set
`pages.status='active'`, restore the file + its sitemap `<url>` block in `gqls/sites`.

## 8. Owner rulings on record

- **2026-08-03: retire `/tools/standard-calc.html`.** Done, verified 404.
- **2026-08-03: consolidation keeps WITHHOLDING** — *"we need honesty at any cost."*
  No code change; already live. **Settled — do not re-open.**
- Still with the owner: `GITHUB_READ_TOKEN` cannot see `gqls/sites` (needs GitHub
  admin, off the critical path); whether to create a `site_plans` row.

## 9. Commits (2026-08-03 → 04)

`b819b2af6` four fixes + defect_vectors · `31188f4f1` pin the defect baseline ·
`c11c8b6f9` backfill_content_data · `06362ad2a` NOTES/RUNBOOK/2 landmines ·
`22cbadf24` README · `23acb5e3f` render_tool_row + cache fix · `08b716e98` shipped +
`bugs_open/182` · `46d93e15c` RUNBOOK correction · `dcdfff230` 016b §9 ·
`0766dfaf4` SUMMARY + 182 CONFIRMED · `a6b9083b0` pin verify_rewrite's baseline ·
`4cb891a7f` sites-checkout landmine · `4d30dfa7f` chrome lock + nav survey ·
`a56e197a7` orphan retired · `ae4aab839` NOTES · `af0cd1eb4` handoff ·
`3cd500e9f` contribution to the 098 lane · `a32fc2406` README ·
`ea4fbf651` golden re-baseline · `edfbb91f3` the live-but-uncommitted template ·
`a9c74992d` commit-before-apply + the file-vs-live sweep

Sites repo: `dd91b2aa7` (retirement — file + sitemap entry).

**No `Council-Reviewed:` trailers, and none owed** — every file is under `docs/` or
`bugs_open/`, which the gate refuses client-side.
