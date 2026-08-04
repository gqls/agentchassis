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
chassis         v1.0.1251, rolled 2026-08-04T19:19Z  (v1.0.1250 was 10:29Z — TWO
                rolls today; both post-roll-checked, see §6 and §6b)
```

*(§1 re-measured 2026-08-04 evening: 26/26 HTTP 200, retired page 404, 11/11
calculators match `GOLDEN_2026-08-03b`. Locks not re-counted this session.)*

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

**(3) ~~`bugs_open/182` is owned by another thread~~ — CORRECTED 2026-08-04 (evening).**

> **This entry was already false when I wrote it, and the NOTES tail said so.**
> `182` is **FIXED, LIVE and CLOSED** — `a43be1e70`, shipped on chassis v1.0.1240,
> and it now lives in **`bugs_closed/`**, not `bugs_open/`. Pod-grepped on
> v1.0.1251 this evening, both replicas: `id_resolved_component` = 1, positive
> control 1, negative control 0. `rerender_sections` resolves by
> `page_components.component_id` first now, so this site's 57 positional slots are
> **exactly the population that fix reaches** — it is no longer a no-op here.
>
> **The offline route is still mandatory — for a NEWER and MORE DANGEROUS reason.**
> The old reason was benign: the re-render did nothing. The live reason is
> `bugs_open/189`, which 182's fix *newly reaches*: firing `section_data_resolved`
> on a previously-unresolvable **locked** section **DUPLICATES it on the page**
> (measured on `tool-loan-vs-savings` 08-03: 5 rows where there should be 4, the
> calculator rendered twice on the served page; remediated live). Still unfixed —
> `extractSectionsFromMetadata` is unchanged since 06-17, verified this evening.
> **Do NOT fire `section_data_resolved` at any of this site's 11 locked tool
> sections** (or oufe.com's 2) until 189 is fixed. Each will reproduce it the first
> time it is touched.
>
> ⚠ **`189` is an AMBIGUOUS NUMBER — `bugs_open/` holds TWO unrelated files.**
> This lane's is `189_HANDOFF_2026-08-03_resolving_a_locked_positional_slot_duplicates_it_on_the_page.md`.
> The other (`189_HANDOFF_2026-08-04_siblingsignatures…`) is a `path:Symbol`
> parser duplicate from the 163 lane and has nothing to do with this site — so
> every commit message saying "189" since 08-04 means the *other* one. Resolve by
> slug, per CLAUDE.md.
>
> **What caught it:** reading the `NOTES` tail on cold start, which carried both
> facts under a `2026-08-03 (platform thread)` heading. I wrote this handoff hours
> later and carried the stale premise forward from the 08-03 handoff without
> re-reading my own notes file. The lesson is not "182 moved" — it is that a
> handoff's §2 was contradicted by its own lane's NOTES at the moment of writing.

**(4) Nothing else is owed.** The four defects, the tidy, the retirement and the
re-baseline are all closed and verified.

## 3. The one live procedure you must not get wrong

> **CORRECTED 2026-08-04 (evening) — the danger INVERTED, the instruction did not.**
> The paragraph below describes **pre-v1.0.1240 behaviour** and is kept only so the
> measurement is not lost. Since `bugs_closed/182` shipped, `rerender_sections`
> resolves by `component_id` and **does reach this site's positional slots**. It no
> longer fails silently — it fails *loudly and destructively* on a LOCKED row, by
> duplicating the section (`bugs_open/189`, open). So the rule is unchanged and the
> stakes are higher: **use the offline route below.** See §2 item (3).

~~**`rerender_sections` DOES NOTHING on this site and reports success** (`bugs_open/182`).
Slots are positional (`prose-0`, `tool-2`); the component lookup keys on
`slot_name`; nothing resolves; every section is carried. Measured `rerendered: 0,
carried: 4` with fixes already live.~~ *(true until chassis v1.0.1240, 2026-08-03)*

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

## 6. Post-roll check for v1.0.1250 (2026-08-04) — ✅ DONE, PASSED

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

### RESULT — the roll moved nothing, and my baseline was contaminated

All three pages re-rendered on v1.0.1250 and all three came back with a **different
md5**. That is the alarm signal, and it was **my own doing**, not the roll's:

```
index.html                  pre=bd4ea8c5ba39  post=05dee40c992d
tools/consolidation.html    pre=b48c4a05cabb  post=be7bd8586779
tools/loan-vs-savings.html  pre=06f9347699d5  post=fcf803a8e1e8
```

Diffing rather than guessing: the **only** change on each page is the footer comment
— the 10-line engineering note replaced by the 2-line pointer, 12 lines changed,
identical on all three, nothing else moved. That is open item (2) propagating
exactly as designed.

**So: v1.0.1250 renders these pages identically. Measured.** And `toolgolden
--compare` against `GOLDEN_2026-08-03b`: **all 11 tools reproduce exactly.**

> ⚠ **MISSTEP IN THE METHOD, worth more than the result.** I captured the pre-roll
> baseline from pages carrying a **known pending change of my own** — the footer fix
> was sitting in `site_components` waiting for each page's next re-render, and I had
> written that down in this very file as item (2) before running the check. So the
> baseline could never have matched, and "DIFFERS" was guaranteed before the roll
> was involved at all.
>
> It survived only because the diff was unambiguous and attributable. A subtler
> pending change would have produced a mismatch I could not have attributed — and
> the standing advice for a mismatch ("re-run first, propagation lag clears") would
> have been wrong here, because re-running would have reproduced it every time.
>
> **Before a post-roll byte check, establish that there is NOTHING pending.** Either
> let every pending change propagate first, or predict the expected diff and assert
> it exactly. A baseline is only a baseline if nothing else is in flight.

### Item (2) status: 3 of 26 pages now carry the corrected footer

`index`, `tool-consolidation`, `tool-loan-vs-savings` picked it up in this check.
The other 23 still serve the old comment until they next re-render.

**HOW TO RUN THIS CHECK NEXT TIME:**

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

## 6b. Post-roll check for v1.0.1251 (2026-08-04 evening) — ✅ DONE, PASSED, clean baseline

**The chassis rolled TWICE on 2026-08-04**: v1.0.1250 at 10:29Z (§6) and
**v1.0.1251 at 19:19Z**. This is the second check.

**The render-path diff is again NOT empty, and this time it is chrome** —
`chrome_link_policy.go` (new, `bugs_open/191`), `render_site_components_action.go`,
`nav_tables.go`, `resolve_internal_links_action.go`,
`load_current_section_content_action.go`. On this site, chrome is the most sensitive
surface there is (25 authored header links, all three slots locked), so it was read
rather than waved through.

**It cannot reach an assembled page.** `assemblePage`
(`rerender_single_page_action.go:532`) loads chrome through `getSiteComponents`,
which is a plain `SELECT slot_name, rendered_html FROM site_components` — it **reads
stored chrome and never renders it**. `LoadChromeLinkPolicy`'s only callers are
`render_site_components_action.go:179` and `nav_tables.go:193`, neither on the
assembly path. The RUNBOOK's named render-path files had **no commits at all** in the
window.

**Measured anyway, because expected is not measured:**

```
index.html                  05dee40c992d -> 05dee40c992d  IDENTICAL
tools/consolidation.html    be7bd8586779 -> be7bd8586779  IDENTICAL
tools/loan-vs-savings.html  fcf803a8e1e8 -> fcf803a8e1e8  IDENTICAL
```

`toolgolden --compare` vs `GOLDEN_2026-08-03b`: **11 of 11 exact.** 26/26 pages 200,
retired page 404. `tool-loan-vs-savings` still has **4** `<section>` blocks (not 5),
so `bugs_open/189` did not bite — expected, as assemble-only never enters
`save_page_sections`.

### The baseline was clean this time, and here is why — READ THIS BEFORE THE NEXT ROLL

§6's warning was *"establish that there is NOTHING pending"*. There was a third
option beyond §6's two, and it is the cheap one:

> **Use the pages that have ALREADY absorbed the pending change as your baseline
> pages.** After §6, exactly three pages on this site were in a settled state —
> `index`, `tool-consolidation`, `tool-loan-vs-savings` — because §6's own check had
> propagated the footer to them. Those three were therefore the *only* correct choice
> of baseline page, and needed no prediction at all.

Anchor it twice before firing: `cmp` the live bytes against the previous check's
`.post` files, so the baseline is provably both **current** and **produced by the old
image**. Both held here.

⚠ **After the footer propagation below completes, all 26 pages are settled**, so this
constraint disappears and any page may be used as a baseline. If a new pending change
is introduced, this paragraph applies again.

Baselines for this check: `/tmp/decomp-work/postroll1251/` (`*.post1251` are the
v1.0.1251 renders). Work items `created_by='postroll-1251'`.

⚠ Two of the three work items hit the known **spawn→call handshake race**
(`workflow completed but its result could not be delivered to the parent
(failed_transient)`) and succeeded on attempt 2 of 3. That is fleet-wide and
expected; do not diagnose it here, and **do not cancel a failing row** — it retries.

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
