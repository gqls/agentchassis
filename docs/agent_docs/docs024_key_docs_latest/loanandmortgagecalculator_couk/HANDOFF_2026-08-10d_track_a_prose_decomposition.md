# HANDOFF — TRACK A: decompose the 17 LMC prose pages. Ready to run.

**Written 2026-08-10 (evening).** Owner: *"please go ahead with track A."*
Parent brief: `HANDOFF_2026-08-10c_continue_here.md` (read §2, §2b, §3 first —
they hold the corrections this file assumes).

**Nothing in this file is planned-but-untried.** The tooling was run end to end
against the live DB in read-only mode while writing it: the manifest builds clean,
all 17 pages pass prediction, and the one guard I could induce was induced. What is
left for the executing session is the `--apply` half, which is the only part that
writes.

---

## 0. Status line

**Track A is 17 pages. No page in it carries a calculator, and that is proven, not
assumed.** The tooling runs today. The pinned ref is safe *for these 17*. Start at §3.

---

## 1. The worklist — 17 pages

Derived from the live DB, not from a doc. `pages.rebuild_policy='generic'` AND
`sections = ["ported-page"]`:

```
guide-car-finance-and-your-mortgage          guide-the-fees-nobody-quotes
guide-consolidating-debt-into-your-mortgage  guide-total-cost-of-borrowing
guide-credit-file-before-a-mortgage          guide-when-repayments-are-a-struggle
guide-deposit-or-clear-the-debt              guides-index
guide-fixed-vs-variable-on-both              index
guide-jargon-buster                          legal
guide-remortgaging-with-other-debt           loans
guide-secured-vs-unsecured-what-changes      mortgages
guide-stress-testing-the-whole-budget
```

`guide-how-loans-affect-mortgage-affordability` is the 18th generic page and is
**already decomposed** (`["prose-0"]`). It is not in the worklist; it is your
reference for what "done" looks like.

### ✅ The safety property, measured — no calculator can be in this set

`decompose_lmc.py` carries a hand-authored `CALCULATOR_URLS` set of 23 pages. Against
the live `pages` table (41 rows), the partition is exact:

```
calculators NOT owned                 : []      (must be empty)
generic pages that ARE calculators    : []      (must be empty)
owned pages that are NOT calculators  : []      (must be empty)
```

All 23 calculators are `owned`; all 18 generic pages are non-calculators. **Use the
`CALCULATOR_URLS` list as the authority, never a regex.** Migration 367 classified
with `onclick=|addEventListener` and missed six calculators that bind
`oninput=`/`onsubmit=`/`onchange=` or keep their listeners in the external
`calculators.js` — and **367's negative control used the same expression as its
filter, so it agreed with itself** (10c §2b, migration 377). I re-ran that same blind
expression while preparing this file and it returned a clean "no tool on any generic
page" — which was *true only because 377 had already fixed it 90 minutes earlier*.
Two checks blind the same way agree with each other.

**Independent confirmation:** `decompose_lmc.py` over these 17 reports
`with a tool block: 0`. A calculator in the set would produce a tool block.

---

## 2. Pre-flight facts, each measured 2026-08-10 evening

### The pinned ref is safe for Track A, and NOT for Track B

`decompose_lmc.py` pins sites-repo ref **`b318a8fad`**. The 08-09 re-baseline moved
`load_lmc.py` to a baseline taken at `b26fdc81b`, and the bugfix-224 work has since
changed the repo. `git diff b318a8fad HEAD` over the site touches **19 files** — and
every one is a calculator page or `assets/js/calculators.js`.

**None of the 17 Track A pages appears in that diff.** So for Track A the pinned ref
and sites-repo HEAD are byte-identical, and the pin needs no attention. **Track B
cannot use this pin unchanged** — its pages are exactly the ones that moved.

### The tooling runs clean, today

```
decompose_lmc.py --pages <the 17>   ->  pages: 17   with a tool block: 0   ->  wrote manifest
load_lmc.py --check --all           ->  17/17 predicted, EXIT=0, no REFUSE
```

Every Track A page decomposes to exactly **one** prose block, so each becomes
`["prose-0"]` — the same shape as the already-done reference page. That is the
design (`decompose_lmc.py` merges consecutive prose into one run), not a defect.

### Chrome is already loaded — do not re-run `load_chrome.py --apply`

`site_components` holds `head` / `header` / `footer`, all `build_status='rendered'`,
all `lock_type='permanent'`, loaded 2026-08-05. The "once per site, before the first
page" precondition in RUNBOOK §12 step 0 is **already satisfied**.

---

## 3. The sequence, per page

```bash
cd /home/ant/projects/agentchassis
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
export DECOMP_WORK=<a scratch dir of your own>      # must be YOURS; see §4 trap 2

# 1. build the manifest for the 17 (seconds, read-only — reads git, writes a file)
TRACK_A="guides-car-finance-and-your-mortgage,guides-consolidating-debt-into-your-mortgage,\
guides-credit-file-before-a-mortgage,guides-deposit-or-clear-the-debt,\
guides-fixed-vs-variable-on-both,guides-index,guides-jargon-buster,\
guides-remortgaging-with-other-debt,guides-secured-vs-unsecured-what-changes,\
guides-stress-testing-the-whole-budget,guides-the-fees-nobody-quotes,\
guides-total-cost-of-borrowing,guides-when-repayments-are-a-struggle,\
index,legal,loans-index,mortgages-index"
python3 $LANE/decompose_lmc.py $DECOMP_WORK/manifest.json --pages "$TRACK_A"
#   EXPECT: "pages: 17   with a tool block: 0"  then "wrote …"
#   A tool block > 0 means a calculator got into your list. STOP.

# 2. predict, whole set (read-only; refuses if a destination row moved)
python3 $LANE/load_lmc.py --check --all
#   EXPECT: 17 lines, each "1 row(s) -> /url (predicted N bytes)" + "0 prose-0 … b", exit 0

# 3. WRITES A LIVE PAGE — one page first, then verify before continuing
python3 $LANE/load_lmc.py --apply legal
#   'legal' is the smallest (2,721 b prose, 9,248 b predicted) and the least trafficked.

# 4. deploy that page: assemble-only page_rerender (spec shape in RUNBOOK §8;
#    NO spec.reason; page_id in the spec AND the column; status 'triaged')
#    then, ~90s after the item reports complete:
diff $DECOMP_WORK/predicted/legal.html <(curl -s -A Mozilla/5.0 https://loanandmortgagecalculator.co.uk/legal.html)

# 5. only then, the remaining 16 — still one at a time, diffing each
python3 $LANE/load_lmc.py --apply <name>

# rollback, per page, from the pre-change rows (--apply takes a backup first)
python3 $LANE/load_lmc.py --restore <name>

# after ANY repo edit on this site
python3 $LANE/gate_component_bytes.py --repair && python3 $LANE/gate_component_bytes.py
```

**Order:** `legal` → `guides-index` → the 12 guides → `loans-index` → `mortgages-index`
→ `index`. Smallest and least-linked first; the homepage last, because it is the one
page where a bad render is seen immediately.

---

## 4. Traps — the first three are new, found preparing this file

### 1. ⚠ `--pages` takes the MANIFEST slug, which is NOT `pages.name`

`decompose_lmc.py` derives its name from the file path (`guides/jargon-buster.html`
→ `guides-jargon-buster`). The database row is `guide-jargon-buster` — **singular**.
Likewise `loans/index.html` → `loans-index` but `pages.name` is `loans`; same for
`mortgages`. **14 of the 17 differ between the two spellings.**

Feeding DB names to `--pages` matches nothing and writes a manifest with **zero
pages** — and `decompose_lmc.py` skips its "expected 23 tool pages" assertion whenever
`--pages` is given, so **an empty manifest exits 0 and prints `pages: 0`**. That is a
silent no-op that looks like a successful run. Always read the `pages: N` line.

This mismatch is *known and deliberately handled downstream*: `load_lmc.py:page_ids()`
joins on **URL, not name**, and its docstring says a bare-name join "was a real defect
caught at --check design time". **Do not 'fix' the naming** — the URL is the shared
identity.

### 2. ⚠ `--manifest <file>` also parses the filename as a page name

`load_lmc.py` builds its page list as `[a for a in argv if not a.startswith("--")]`,
which picks up the value of `--manifest` too. `--check --manifest manifest_mutant.json legal`
fails with `no manifest entry for 'manifest_mutant.json'`. **Use a separate
`DECOMP_WORK` directory per manifest instead of `--manifest`.**

### 3. ⛔ `--check` is a DESTINATION guard, not a CONTENT guard — INDUCED

I mutated a manifest's prose block (appended an HTML comment to `legal`) and re-ran
`--check`. **It passed**, reporting 9,273 predicted bytes against the clean run's
9,248. No REFUSE.

Reading the three guards (`load_lmc.py:281–305`), that is correct behaviour and worth
knowing exactly:

| guard | compares | catches |
|---|---|---|
| title / meta_desc | manifest vs `pages` row | assembly silently changing the title |
| baseline md5 | **stored row** vs `BASELINE_2026-08-09_…txt` | another session having touched the page since 08-09 |
| dropped sections | assembly output | a section assembly would discard |

**None of them compares the manifest's prose HTML to the stored page.** So content
fidelity rests entirely on `decompose_lmc.py`'s own assertions at manifest-build time
(P-visible, P-stranded, P-script-bytes, …), which refuse to write the manifest at all
on failure. That is a strong control — it refuted the tool's own first draft on real
pages — but it is a *different* control from `--check`, and it is upstream. If you
hand-edit a manifest (which the voice pass does by design), `--check` will not catch a
content error you introduce.

`[NOT INDUCED: decompose_lmc.py's own assertions.]` They ran clean over all 17, but I
did not mutate a page to watch one fire. If you want the content guard proven rather
than inherited, that is the induction to do, and it is cheap.

### 4. The baseline guard IS live and did pass

For each of the 17, `--check` asserted the page has exactly **1** component row whose
`md5(rendered_html)` equals the 08-09 baseline. All 17 passed, so nothing has moved
those rows since 08-09. This guard exists to catch another session's write — if it
REFUSEs on a page, **do not regenerate the baseline to make it quiet**; find out who
wrote (10c and the 08-10 handoff §6.3 both record why).

### 5. Carried over from RUNBOOK §12 — still true

- **`build_site.py` is DEAD for any page you decompose.** The DB becomes the render
  source; rebuilding from the build scripts and pushing would fight the next rerender,
  silently.
- **Fetch the served page ~90 s after the item says `complete`.** Inside the deploy
  window B2 returns a `NoSuchKey` JSON at HTTP 200 and every grep against it returns 0,
  which reads as a clean pass. Guard on byte count and a leading `<!DOCTYPE`.
- **Tab field separator with psql on this site** — every title contains
  `" | LoanAndMortgageCalculator.co.uk"` and the default `|` splits inside the data.

---

## 5. What Track A actually delivers — state this plainly, do not oversell it

Measured, so the owner gets the real shape:

- **One prose component per page, not paragraph-level sections.** All 17 collapse to a
  single `prose-0`. The framework can rewrite that block; it cannot yet address
  "paragraph 3" independently.
- **It reduces shared-template blast radius but does not remove it.** Today all 40
  verbatim pages point at ONE `content_components` row — *"Ported Page
  (webdesign.co.uk)"*, used by **154 pages across 3 sites**, and corrupted once
  already (08-04). After decomposition a prose row points at *"Ported Prose Block"*
  (`ported-prose`), used by **29 pages across 2 sites**. Better, not isolated. Only
  tool components get a page-scoped definition of their own (consolidation's is 1 page,
  1 site).
- **It makes the page genuinely rebuildable.** `["prose-0"]` resolves through the build
  path now that `bugs_open/204` is fixed (10c §2). A `["ported-page"]` verbatim page
  does not participate meaningfully in the generic pipeline.
- **It does not seed a site plan.** `site_plans` is still 0 rows for both sites.
  Decomposition gives per-component editing; wholesale rebuild-from-plan is a separate
  decision (10c §3).

---

## 6. Definition of done for Track A

1. All 17 pages `sections = ["prose-0"]`, one `page_components` row each, function
   `ported-prose`.
2. Each page's served HTML matches its `predicted/<name>.html` byte for byte.
3. `python3 $LANE/verify_site.py` clean (link graph, canonicals, sitemap, ld+json).
4. `gate_component_bytes.py` clean.
5. The arithmetic estate untouched — Track A touches no calculator, so
   `oracle.py` should be unchanged. Run it once at the end anyway with its controls in
   the same session; a moved number means something reached a page it should not have.
6. NOTES + `README_where_we_are` updated as you go, not at the end.

**Then stop.** Track B is a separate brief and a separate risk class — it needs the
pin re-pointed (§2), the tool row locked, and the §6 re-slot question in 10c settled
on one page before it goes wide.
