# NOTES — loanandmortgagecalculator.co.uk

Append-only, newest at the bottom. Evidence, commands, what the system actually
said, and every misstep. The missteps are the point.

---

## 2026-07-31 — session 1

### The source of truth was not where it looked

`~/projects/domains/mortgagecalculator.co.uk/` has both a top level and a
`gemini/02/` subdirectory holding what appear to be the same 23 pages. **The live
site serves `gemini/02`, not the top level.** Established by sha256, not by
looking at dates:

```
c2144f3e…  live https://mortgagecalculator.co.uk/
be66f725…  domains/mortgagecalculator.co.uk/index.html          <- NOT live
c2144f3e…  domains/mortgagecalculator.co.uk/gemini/02/index.html <- live
```

Then all 23 pages: **23/23 byte-identical to live, 23/23 HTTP 200.** Had I taken
the top-level tree I would have ported a different site and every later check
would have passed against the wrong baseline.

### First provenance run: 14 of 14 calculators BROKEN. All 14 were the harness.

`webdesign_tools_repair/toolaudit.py` reported `NO-CONTROL` — "nothing a visitor
can touch" — for every interactive page on a site whose calculators all work.

The tell was the shape: **14 of 14 is not a site, it is an instrument.** A site
with fourteen genuinely dead calculators would not have fourteen *identical*
failures.

Cause: every check scoped its queries to `main …`; the site has no `<main>`
element (`hasMain: false` via `evalpage.py`, 3 inputs and 1 button present in the
DOM). Fixed with an asymmetric `SCOPE_FN` fallback and committed to the owning
lane (`f38f5bf7f`); full write-up is in that lane's NOTES as harness fault ten.

**The lesson I want to keep: I nearly wrote "the source site's calculators are
broken" into this file as a finding.** What stopped me was the count, not
diligence. The generalisable check is the one `evalpage.py` performs in twenty
seconds — *did the harness do the thing it says it did?* — and it should be the
first move on any adverse verdict, not the last.

### Baseline, once the instrument worked

13/13 mortgage calculators `RESPONDS`. The 14th, the homepage, is `NO-CONTROL`
and that is **correct**: its 14 "buttons" are all `<a>`-wrapped navigation. A hub
page has nothing to compute.

### Defects on the source site, which the blind harness had hidden

Found by walking the link graph rather than by the audit:

- 6 of 9 guides link `Home` to `index.html` from inside `/guides/`, resolving to
  `/guides/index.html` → **live 404**
- the homepage links `guides/mortgage-scorecard.html`; the file is
  `your-mortgage-scorecard.html` → **live 404**
- 2 guides are orphans
- `sitemap.xml` is 404, and `robots.txt` still carries
  `# Sitemap location (replace with your actual domain)`
- no `favicon.ico`, so every page load logs a 404 (this was the one real console
  error in the baseline, and it belongs to the browser, not the page)

All fixed on the new site. **None fixed on the old one** — out of brief, and worth
offering separately.

### The framework could already do source ≠ destination. The doc said it couldn't.

`FUTURE_adoption_source_destination_separation.md` (dated 2026-04-20) describes
`destination_domain` as unbuilt future work with a Go change list. It is built:

```
ensure_site_record.config.domain_override_field = "input_data.destination_domain"
   consumed at platform/orchestration/actions/site_db_actions.go:131
crawl_site.config.url_field = "input_data.target_url"
```

Three months of drift. I checked because of the "prior-art searches go stale"
rule, and it saved building something that exists. **A `FUTURE_` filename is not
evidence about the present.**

### And `fidelity=locked` is live, but its byte source is wrong

Live in the pod (`v1.0.1211`), pod-grepped on three added strings plus a control:
`1 / 1 / 2`. Note that `grep -c fidelity` would have "confirmed" this on a binary
without the feature — there were ~10 pre-existing unrelated hits.

But the loancalculator lane had already found, hours earlier, that firecrawl's
`rawHtml` is the **serialised post-JavaScript DOM**, not the origin's bytes. It
mutated 3 of their guides in production before they caught it — `md5` matched on
**0 of 27** pages, each larger by a near-constant 8,900–9,060 bytes, which is a
signature rather than noise.

Their conclusion, which I have adopted: **the deploy repo is the byte source, not
the crawl.** So this site is built in the repo and verified there, and adoption
comes after.

I found this by reading `git log` on the sites repo before copying their files. The
top commit was `revert loancalculator: restore 3 guides to served source after a
verbatim-adoption deploy mutated them`. **Had I not read it I would have run
`--fidelity locked` and mutated 23 calculators.**

### Building: the mistake that the second assertion now catches

`build_site.py` asserts every inline `<script>` block is byte-identical to source.
It passed on all 24 pages. The browser audit then said
`mortgages/bridging-loan.html` — `ReferenceError: formatGBP is not defined`.

**The assertion was true and the page was broken.** `bridging-loan`,
`equity-release` and `fee-analyser` keep `<script src="js/calculators.js">` in the
`<head>`; the builder rebuilds `<head>` from scratch, so their only dependency was
silently discarded. The inline blocks were untouched — that is exactly why the
assertion passed.

`bridging-loan` threw because its code calls `formatGBP`. **`equity-release` and
`fee-analyser` scored `RESPONDS` with the same dependency missing** — they respond
without needing the helper, so nothing complained. Two of the three would have
shipped broken-ish and green.

Fixed structurally, not per-file: dependencies are now scanned from the **whole**
source and a second assertion fails the build if any goes missing. **A
byte-identical script with a missing dependency is still a broken page**, and a
safety property that only covers the code and not what the code imports is
half a property.

### Three more adverse verdicts; two were the harness again

| tool | verdict | truth |
|---|---|---|
| `loans/damage-checker` | DEAD, "driven and pressed, nothing changed" | **works.** Four checkboxes, no buttons; `pick()` had no checkbox branch, so `target` was null and nothing was ever driven. Proven by hand: ticking `#dmg-1` takes `#damage-verdict` `display:none → block`. Harness fault 11 |
| `loans/credit-health-check` | DEAD | **real, and not mine.** A 5-step wizard relying on `.check-step` / `.active` CSS that **neither source stylesheet ever defined**, so all five steps rendered at once. Fixed in CSS. Then it *stayed* DEAD — `snap()` reads a fixed selector list and cannot see which region is visible. Harness fault 12 |
| `loans/credit-roadmap` | NO-CONTROL | **correct.** 1,816 bytes, zero controls, zero script. Not a tool. Not ported; reason recorded in the builder rather than dropped silently |

That `credit-health-check` case is the most interesting thing in this session. **A
tool can be broken by CSS that was never written, and no amount of reading the
JavaScript finds it** — the script does exactly what it should; the missing half is
in a stylesheet, in a class name that appears nowhere except as a string in a
`classList.add` call. It was broken on the live source site and nobody had noticed.

That is also how I found that **36 classes the loan tools use are undefined in
their stylesheet.** `.check-step` was merely the one where the absence was
load-bearing rather than cosmetic.

### Wrong turns worth recording

- **My own link checker reported 60 dead links, all false.** It did not resolve
  `href="/"` to `index.html`. I very nearly "fixed" the site. The check to apply to
  a checker is the same one I applied to the harness: if the failures are uniform,
  suspect the instrument.
- **My leakage grep reported the old brand on every page.** `grep -v
  loanandmortgagecalculator` is case-sensitive, and
  `LoanAndMortgageCalculator.co.uk` **contains** `MortgageCalculator.co.uk`. The
  exclusion has to be `-i`. Two instrument errors in five minutes, both producing
  confident output.
- **I pushed and it was rejected**, three other pushes having landed in the
  interim. `git pull` was the obvious response and would have been wrong:
  `bugs_open/120` means a merge commit makes the deploy green and empty. Used
  `--rebase` and asserted `parent count == 1` before pushing.

### Verified state at end of session

| check | result |
|---|---|
| calculators in a real browser, served locally | **23 / 23 RESPONDS** |
| dead internal refs | **0** |
| orphan pages | **0** |
| sitemap vs disk | exact both ways |
| old-domain / old-brand / beacon / `nav.js` leakage | **0** |
| head essentials (title, canonical, og:url, skip link, footer) on every page | **43 / 43** |
| B2 upload | **52 / 52 files**, run `30618165416`, domain named in `Changed domains` |
| harness drift from my 3 changes (4 pages × 9 fields, twice) | **0 / 36** |
| build assertions mutation-tested | both go red on a mutant |

**Not live.** The domain is parked at its registrar; the Cloudflare zone and worker
route are owner-only and there are no Cloudflare credentials on this machine. The
B2 keys are already correct, so it serves the moment the route exists.

### One thing I have NOT verified, and am marking as such

**[UNVERIFIED] the 11 ported loan calculators compute the same numbers as their
originals.** What is verified is that their inline scripts are byte-identical to
source and that all 23 pages respond to being driven. Byte-identical logic plus a
present dependency is strong, but it is not the same as asserting equal output for
equal input, and I have not done that. A per-tool acceptance check —
`features_open/015`'s criteria-fence idea, one measurable claim per tool — is the
right shape for it and is the obvious next piece of work.

---

## 2026-07-31 — session 2: the site went live, and three of my own defects came with it

The owner added the Cloudflare zone and the Workers Route. Verified before doing
anything else, comparing against a healthy zone **in the same breath** so a
misconfigured zone could not be mistaken for a slow network:

```
https://loanandmortgagecalculator.co.uk/worker-health  ->  "Worker is running!"
https://mortgagecalculator.co.uk/worker-health         ->  "Worker is running!"   (control)
NS: betty/ivan.ns.cloudflare.com on both
```

Then all 52 files fetched and digested: **51 byte-identical to the repo.** The 52nd
is `robots.txt` (198 B repo → 2,034 B live) and it is **not a defect**: Cloudflare's
zone-level *Managed robots.txt* prepends content-signal directives. The control
domain does the same, and my own rules and `Sitemap:` line survive at the tail.
`verify_site.py` now carries that as its one sanctioned byte exemption, with the
reason, so nobody re-investigates it.

### Defect 1 — three hub URLs 404'd live, and my instruments had been taught not to see it

`/mortgages/`, `/loans/` and `/guides/` all returned **HTTP 404**. The worker maps
`{hostname}{path}` straight to a B2 object key and rewrites **only** `/` →
`/index.html` (`scripts/cloudflare/worker.js:8-11,27`), so `/loans/` asks B2 for an
object literally named `loans/`. Fleet-wide property (`bugs_open/116`), confirmed on
`loancalculator.co.uk/tools/`, `mortgagecalculator.co.uk/guides/` and
`webdesign.co.uk/tools/`. **Mine was the only site in the sites repo that LINKED to
such a path** — 0 distinct directory-path hrefs on loancalculator.co.uk, 3 on mine.

Measured blast radius: **42 of 42 pages** carried all three links, **4 of 49**
distinct internal references 404'd, **3 of 41** sitemap URLs 404'd, and **3 pages'
own canonicals** pointed at a 404. That last one is the worst of it — a canonical in
the sitemap naming a URL that does not exist.

**Two instrument faults, and the second is the one worth keeping.** First, I verified
against `python3 -m http.server`, which **does** resolve directory indexes — a *more
forgiving* server than production. Second, and worse: when my own link checker
flagged `/loans/` as dead, I recorded it in this file as a false positive and "fixed"
the checker to resolve `/loans/` → `loans/index.html`. **I taught the instrument the
same forgiveness, converting a true positive into silence.** That is
`narrowing-a-detector-can-make-it-inert`, self-inflicted, and the session-1 entry
above ("My own link checker reported 60 dead links, all false") is **partly wrong** —
of those 60, three were real.

> **CORRECTED 2026-07-31:** the session-1 claim that all 60 link-checker hits were
> false is false. The `/loans/`, `/mortgages/` and `/guides/` hits were **correct**,
> and the "fix" to the checker is what hid them for a day. Caught by fetching every
> reference from the live origin instead of resolving it locally.

### Defect 2 — all 13 guides shipped JSON-LD that no parser could read

`build_pages.py:161` built the headline as
`html.escape(repr(title).replace("'", '"'))`. `html.escape` defaults to
`quote=True`, so it escaped the quotes it had just inserted:

```
"headline":&quot;Loan and Mortgage Jargon, Translated&quot;,
```

**13 of the 14 `ld+json` blocks on the site failed `json.loads`.** Google discards
invalid structured data silently, so nothing complained and all 13 guides lost
rich-result eligibility.

Same root cause as defect 1, stated plainly because it is the transferable part:
**every pre-launch check asserted PRESENCE where it needed to assert VALIDITY.** The
session-1 table above says "head essentials (title, canonical, og:url, skip link,
footer) on every page — 43/43". That was true and it was worthless: the canonical was
*present* on every page and *resolved* on 39 of 42.

### Defect 3 — the copy claimed 24 calculators, and there are 23

Three places said "24 free UK calculators"; two said "12 loan calculators" and one
"Twelve calculators" for a section holding 11. Dropping `credit-roadmap` (correctly —
it is 1,816 bytes with no controls) never propagated into the prose. A false number
in copy is the `bugs_open/161` failure mode: the artefact asserts a fact and then
vouches for it. Counts are now **derived** from the tool tables via `len()` and a
`word()` helper, so prose cannot drift from reality again.

### What actually fixed these, and how the guards were proven

Structural, not per-file:

- one `hub(section)` helper defines the hub URL shape for all 13 emission sites;
- counts derived from `MORTGAGE` / `LOAN` / `GUIDES`, never typed;
- **`write()` — the one function both builders funnel through — gained two
  assertions**: no emitted `href`/`src` may name a directory, and every `ld+json`
  block must parse. There are now four build properties, and **all four were
  mutation-tested red** (`features_open/027` S2), including the two pre-existing ones
  I had changed the builder around:

| mutant | result |
|---|---|
| `hub()` returns `/{section}/` again | `ABORT … reference "/mortgages/" names a directory` |
| JSON-LD headline re-escaped to `&quot;` | `ABORT … ld+json does not parse` |
| `parseFloat` → `parseFloatX` in output | `ABORT … inline script blocks changed` |
| every external `<script src>` dropped | `ABORT … dependency /assets/js/calculators.js lost` |

`verify_site.py` is new and **defaults to `--live`**, because the local server is the
instrument that hid defect 1. Its `--disk` mode models the worker's *single* `/` →
`/index.html` rewrite and **deliberately refuses to resolve any other directory
path**, so the two modes agree about what a valid reference is.

It earned its place immediately: on its first run it caught **a fourth defect the
build assertions could not see** — `href="{hub('guides')}"`, an unexpanded f-string
placeholder, because that segment of `HOME_BODY` was a plain string not an f-string.
Valid-looking markup, not a directory path, invisible to property 3.

Then the red→green, which is better evidence than a synthetic mutant: 4 failures →
fix → 0 failures. And the dead-reference check was separately proven on the
motivating case by injecting `href="/loans/"` into a built page (`FAIL … 1 FAILURE`),
then restored by rebuild and confirmed byte-identical with `cmp`.

### Verified live after the fix

| check | result |
|---|---|
| distinct internal references resolving live | **48 / 48** (was 45/49) |
| sitemap URLs resolving live | **41 / 41** (was 38/41) |
| canonicals that resolve AND self-name | **42 / 42** |
| `ld+json` blocks that parse | **14 / 14** (was 1/14) |
| files byte-identical live | **51** + 1 sanctioned (`robots.txt`) |
| calculators RESPONDS in real chromium, live | **23 / 23** |
| build assertions red on a mutant | **4 / 4** |

One transient to record so it is not mistaken for a regression: the first live audit
returned `HARNESS-ERROR … [Errno 32] Broken pipe` on `mortgages/overpayment.html`.
Re-run alone: `RESPONDS`. **A harness error is not a site verdict** — the same lesson
as session 1's three harness faults, now at 4 of 5 adverse verdicts on this site
being the instrument.

### Adoption: what the pipeline can and cannot do, read before running it

`adopt_verbatim.go` has **exactly one** possible byte source — firecrawl's `rawHtml`
(`apply_adoption_plan_action.go:873`, DB fallback `:912-940` reading `raw_html`). No
local-file or repo path exists anywhere in the platform; the gap is inventoried as
**G1**, and `diagnose_read_repo_files_action.go`'s token **cannot see `gqls/sites`**
(404 while authenticated — a fine-grained PAT scoped to selected repos).

So I measured my own exposure rather than inheriting the sibling lane's number:

| page | served | post-JS DOM | `<option>` |
|---|---|---|---|
| `loans/credit-health-check` | 6,990 B | 6,983 B | 0 → 0 |
| `loans/application-tracker` | 9,138 B | 9,148 B | 0 → 0 |
| `mortgages/fact-finder` | 15,859 B | 15,954 B | **32 → 32** |

Divergence is cosmetic in scale (−7/+10/+95 B), not the sibling's 8,900 B inflation,
because this site has no `nav.js`. Identical `<option>` counts mean **no script
injects DOM on load**, so a stored-DOM round trip would not duplicate controls. But
**3 of 3 still differ**, so the byte gate fails everywhere and repair is mandatory.

> **[UNVERIFIED]** that is *Chromium's* serialisation, not firecrawl's. firecrawl
> additionally absolutises URLs, rewrites `<meta charset>` and escapes `&` — measured
> by the sibling lane, not by me. Treat the figures as a **lower bound**.

### Two things I checked that turned out NOT to be true

- **`sites.locked_at` does not hold dispatch.** Migration
  `213_dispatch_gate_matches_dispatcher.sql:106` contains
  `AND s.locked_at IS NULL` in the dispatcher's selector, and I was about to use it as
  a clean site-level hold. **The LIVE `build-pipeline-trigger` row has no such
  clause.** Read the live `agent_definitions` row, not the migration — "the seed is
  not the system", and a migration file is no better evidence than a seed.
- **The race is real and tight.** `scheduled_tasks` shows `build-pipeline-trigger`
  **enabled at `interval_seconds=120`**, last fired a minute before I looked. So the
  window between `adopt_verbatim` creating `page_rerender` items (already
  `status='triaged'`, inside the adoption transaction) and mutated bytes deploying is
  **under two minutes** — not something to check by hand afterwards. Hold first,
  then look.

The hold is a poller that flips new `page_rerender` items to `deferred`, chosen
because `deferred` is **not** in `workItemTerminalStatuses`, so the row still holds
its `idx_swi_dedup` slot (no duplicate can appear behind it) and the release is a
plain `UPDATE` back to `triaged`.

---

## 2026-08-05 — session 3 (portfolio_positioning handoff): voice rebuild lane opened

The owner-directed task (portfolio_positioning handoff, updated 08-05): apply the
chosen "gentle explanatory" voice to the whole site. Design + reasons in
`PLAN_2026-08-05_voice_rebuild_and_decomposition.md`. Recon, all measured today:

- **Stored rows are clean: 41/41 md5-identical to `origin/master` `b318a8fad`.**
  The 07-31 worry (crawl-DOM divergence, "repair is mandatory") does not describe
  what is stored now — whatever repaired it worked. Checked stored
  `page_components.rendered_html` md5 against `git show` per page, not inherited.
- All 41 pages still `1|1` verbatim; no open work items; no site plan.
  `content_direction` v2 (19 rules, 22,980-char formatted, 08-05 row) is current;
  both 07-31 generations preserved superseded.
- Site serves (HTTP 200). Page classes: 13 guides + guides index; home; legal;
  loans index + 11 self-contained calculators (inline scripts); mortgages index
  + 12 calculators that also load `/assets/js/calculators.js` (shared arithmetic
  + localStorage portfolio). `site.js` is chrome-only (mobile menu). **Zero
  page-local `<style>` blocks site-wide** — the sibling lane's style-placement
  complexity does not exist here.
- Writer-path survey (code + live agent_definitions): `save_page_sections`
  refuses owned pages (:150); section-editor has no writer step (applies
  pre-authored edits); lendzy's copy came from needs_page → page-build-handler,
  which needs a plan and writes fresh. Hence: TRANSFORM the existing copy per
  the approved voice, keep pages owned, decompose replacing the verbatim row
  in-transaction, widgets byte-original in locked rows. Full reasoning in the
  plan.
- A near-miss worth recording: I created the plan file believing this lane
  directory was NEW — the Write tool then refused my NOTES "creation" because
  the file existed. **The lane had two full 07-31 sessions and two CONTRIB
  files I had not read.** The read-before-write rule caught it; the CONTRIB
  chrome warning (buildDefaultHead's plural `styles.css` + no header/footer on
  the first decomposed page) is now Phase-1-blocking in the plan.

### 2026-08-05 — Phase 0 done: 23/23 calculators baselined, three-way byte alignment

- `acceptance/GOLDEN_2026-08-05_prechange.json` — 22 tools via the sibling's
  `toolgolden.py`, then **self-verified: all 22 reproduce exactly on --compare
  against live.**
- **`mortgages/investor.html` is toolgolden-UNCERTIFIABLE, and it is the
  instrument, not the page — harness fault: a ratio-only calculator is
  invariant under toolgolden's uniform x1/x2/x0.5 vectors.** Its two functions
  compute gross yield (rent*12/price) and LTV (loan/price); scaling every
  field by the same factor moves no ratio, so the inert-tool guard fired
  ("output is identical for every input value") and refused the whole golden.
  Arithmetic read and hand-checked: 1200*12/250000 = 5.76% on defaults. That
  makes it 5 of 6 adverse verdicts on this site being the instrument (session
  1's series). Remedy: `investor_golden.py` — staggered vectors, ONE field
  moves per vector (rent x2 → 11.52%, price x2 → 2.88%, loan x2 → 150% + the
  high-LTV comment branch). `acceptance/GOLDEN_2026-08-05_investor.json`,
  self-verified 5/5 on --compare. Inherits toolgolden's settle() (mid-parse
  trap) and storage-clearing reload discipline verbatim.
- `acceptance/BASELINE_2026-08-05_stored_md5_at_b318a8fad.txt` — md5+length
  of all 41 stored rows, pinned to the sites-repo sha in the filename.
  **stored == repo (41/41) and live == repo (3-page spot check, browser UA).**
  All three surfaces agree; nothing is pending anywhere. A clean baseline in
  the §6/§6b sense of the sibling's handoff — any page can be a canary.

### 2026-08-05 — chrome installed (locked, inert); decomposer proven over all 41; canary applied

Chrome: 3 locked `site_components` rows via `load_chrome.py` (INSERT-guarded,
asset checks with a proven negative control, nav+footer links joined against
`pages`, splice literals asserted). **Its balance check caught a literal
`<div` inside my own comment on the first run — the sibling's exact
documented first-draft failure, reproduced independently.**

Decomposer (`decompose_lmc.py`): first draft assumed whole-child tool marking
and byte-slice prose runs; **the assertion suite refuted it the same hour on
real pages** — all-41 "header contains a script-addressed element" (uniform
failure = instrument: `marked()` counts ANY `<button>`, and the header's
mobile-menu button is chrome), six mortgages pages "decomposed to zero prose"
(intro copy shares an <article>/card wrapper with the widget), and
loans/standard-calc keeps two compliance boxes OUTSIDE its .container.
Rewrote around the sibling's proven `split_ordered` descent; all 41 pages now
decompose clean, 23/23 tool blocks, every assertion green. Geometry: only
`container-tight` needs a wrapper (the head-chrome shim makes `<main>` the
container) — tight pages get their wrapper INSIDE `content_data.content` so a
future re-render from the shared `ported-prose` template reproduces stored
bytes exactly.

Three more instrument faults caught before they cost anything, logged here
because each wore a real-defect costume:
- **voice_apply F1/F2 flagged the same figure as invented AND lost** ("2022")
  — the number regex swallowed a trailing comma. The contradiction is the
  signature: one token, two opposite verdicts.
- **load_lmc reported adoption had truncated every page title** — my psql
  parser split on `|`, and every title on this site contains
  " | LoanAndMortgageCalculator.co.uk". Tab separator; re-measured: 0 of 41
  titles or descriptions differ between DB and the original heads.
- **The sibling's assemble_mirror predates seam 2** and predicts no
  canonical; the live Go assembler emits one. Added `inject_canonical` to
  load_lmc's prediction path — without it the first real render would have
  differed from the prediction by exactly one line, blaming the decomposition.

Canary (guides-how-loans-affect-mortgage-affordability, the page whose intro
the owner approved as sample #2): overlay authored, all voice guards pass
(figures subset of the page's own, links/anchors multiset-equal, contractions
present, banned performatives absent). Verbatim row REPLACED in one txn
(owner ran the --apply after the auto-mode classifier declined it for me;
backup `page_components_bak_20260805_lmc` holds all 41 originals).
`page_rerender` c64c273e filed triaged 20:53Z, assemble mode (no reason),
page_id in spec AND column.

### 2026-08-06 — CHECKPOINT PASSED: both canaries byte-identical to prediction

Deliberately two canaries, not one — the sibling lane's two DISAGREED and a
single agreeable one would have licensed a confident wrong prediction. Here
they were a guide (prose only) and a calculator (prose + locked widget).

| canary | rerender | served vs predicted |
|---|---|---|
| guides-how-loans-affect-mortgage-affordability | complete in ~50s | **byte-identical** (13,051 b) |
| loans-consolidation | complete in ~50s | **byte-identical** (12,865 b) |

**So the mirror is validated against the real Go path** — the sibling's
"hypothesis with a scheduled test" has now had its test, on this site, on
both page shapes. Predictions for the remaining 39 can be trusted to the same
degree, and every one is still diffed anyway.

**The widget survived: `golden_compare_post.py` MATCHES on consolidation —
all 11 money/control fields byte-equal to the pre-change golden** (including
`old-int £1,664,666.67` and the full verdict string).

⚠ **A known structural divergence affects all 23 calculator pages, and the
raw `toolgolden --compare` reports it as a red result.** The golden
fingerprints every id-bearing element; the hand-built wrapper carried
`id="content"` (the skip-link target) and therefore held the entire page text,
while the decomposed page's `content` is the site header's empty
`<span id="content" tabindex="-1">`. Only that one field moves —
11 of 12 matched exactly on the canary. `golden_compare_post.py` encodes this
as a SHAPE ASSERTION rather than an ignore-list: `content` must be exactly
`|inline` (empty, inline), anything else fails, and every other field must
match exactly. **Its positive control is run, not assumed** — `--self-test`
compares one tool's golden against a different tool's live page and must
report FAIL; it found 90 problems, so the comparator can still fail.
A full post-rebuild golden re-baseline is owed once the site is finished (the
sibling lane owed the same after its decomposition).

### 2026-08-06 — OWNER RULING: the framework writes the copy, not this session

*"I want all the copy to be done through the framework and not through this cli
session. We'll need to restart all those that have been written through this
cli."* Correction block filed at the head of the PLAN; decision 1 struck
through, not deleted.

**What I got wrong, precisely.** I asked "which write path can a session drive
against owned pages today?" and answered it well — the writer refuses owned
pages, there is no site plan, so transformation was the only route a *session*
could take. That reasoning is sound and the conclusion was still wrong, because
the question was never which route a session can drive. CLAUDE.md answers the
real one outright (EVERY SITE GOES THROUGH THE FRAMEWORK, owner 08-04, raised
by a hand-built shopfront on the lane whose product is framework-built sites),
and I read that ruling, cited the sibling lane's practice, and still filed the
in-session route as decision 1. **The tell I walked past: my own guards check
that the copy preserves facts and links — nothing in them could ever check that
the platform was capable of producing it.** A verification suite that cannot
express the property that matters is a sign the property was never in scope.

39 overlays, all passing every guard, are superseded. Kept on disk as evidence
of the register; `load_lmc.py` now defaults to `manifest.json` so the
superseded copy cannot ship by accident (a flag would have been enough, but the
DEFAULT is what a future session actually runs).

### 2026-08-06 — MEASURED: a locked tool row survives the generic rebuild's DELETE

The owner asked for this to be verified before choosing a route, and it is the
fact that decides Route A.

Read first: `save_page_sections_action.go:708` deletes with
`pageComponentAgentWritableSQL`, i.e. **lock-aware**. Three defences in series
(which is why a passing mutation here would need care — a pass could be any one
of them): the lock-guarded DELETE, the Layer-1 `INTERACTIVITY REGRESSION
BLOCKED` guard (`:580`), and the Layer-2 interactive carry-forward
(`:375-443`).

Then induced, against the real decomposed `/loans/consolidation.html`, inside
`BEGIN … ROLLBACK`, running the EXACT predicate the action uses:

```
BEFORE        prose-0 | tool-1 (locked) | prose-2
DELETE 2
AFTER-DELETE  tool-1 (locked)                  <- widget stands
AFTER-ROLLBACK 3 rows                          <- nothing kept
```

**`DELETE 2` is the control**: the statement removed rows, so it was live. A
delete that removed nothing would have "proved" the same thing while proving
nothing — the no-op case, which is the one worth checking.

⚠ **THE CAVEAT IS THE FINDING.** Surviving is not the same as staying put.
`matchLockedRow` repositions a locked row by matching `slot_name` against the
incoming section name; our slots are positional (`tool-1`), so a writer's
sections will not match, and `:855` moves an unmatched locked row to
`len(sections)+1` — **the calculator lands at the BOTTOM of the page**, under
all the new prose, with a `lock_blocked` item raised and no other signal.
Silent unless someone looks at the rendered page. [UNMEASURED end-to-end: read
from code plus the SQL test above; no full writer run has been driven against a
live page.] Remedy if Route A is chosen: re-slot tool rows to names the plan
will emit, BEFORE flipping any page.

### 2026-08-06 — OPEN GAP the owner named: our calculator checks prove CONSISTENCY, not CORRECTNESS

*"check that we have a comprehensive check on the calculators that they produce
validated output"* — we do not, and the distinction is exact:

`GOLDEN_2026-08-05_prechange.json` records **what each calculator did on
2026-08-05**. Every later comparison asks "does it still do that?". If a
calculator has been wrong since the day it was written, the golden records the
wrong answer faithfully and every future check certifies the bug — the same
shape as [[a-pass-from-a-blind-check-outlives-the-blindness]]. This site has
already produced one such case: `credit-health-check` was broken on the SOURCE
site by CSS that was never written (07-31 session 1), and nothing computational
noticed.

What "validated" needs is an INDEPENDENT ORACLE — recompute from the standard
amortisation formula, the published SDLT bands, the definition of LTV — and
compare. NOT BUILT. Scope honestly when it is: roughly half the 23 are analytic
(repayment, simple, standard-calc, overpayment ×2, consolidation, stamp-duty,
investor, loan-vs-savings, settlement, compare-loans, rate stress ×2,
fee-analyser, bridging, equity-release roll-up, portfolio aggregate); the rest
(credit-health-check, damage-checker, application-tracker, fact-finder) are
scoring tools or checklists with no external right answer to check against, and
saying so is part of the deliverable.

### 2026-08-08 — the open gap is CLOSED: an independent oracle, and it found two defect families

The 08-06 entry above ended "NOT BUILT". It is built. Full account:
`REPORT_2026-08-08_arithmetic_validation.md`; commands in RUNBOOK §13.

**176 oracle checks against the live site — 143 PASS, 27 FAIL, 6 CONVENTION;
21 class-C invariant checks, all passing.** Two bug files: `bugs_open/225`
(SDLT), `bugs_open/224` (zero rate). 8 of 23 tools affected.

**How independence was kept, since it is the whole value.** `inventory.py`
reads each page the way a user does — the visible `<label>` bound to each
control, the button text, the caption above each result box — so the per-tool
specs could be authored without opening a single calculation body.
`oracles.py` computes from the annuity formula, the gov.uk SDLT tables and
arithmetic identities. **No page's arithmetic was read until a check had
already FAILED.** Both SDLT defects and all seven zero-rate defects were found
before any of that source was opened; reading it afterwards is diagnosis, not
authorship.

**Defect family 1 — `mortgages/stamp-duty` is running an expired tax rule.**
Its FTB branch is gated `price <= 625000`, which was the cap between
2022-09-23 and 2025-03-31. Since 2025-04-01 relief is unavailable above
£500,000 and standard rates apply to the whole price. Flat **−£5,000** for
every FTB purchase in (£500,000, £625,000]. The page's own prose says the
temporary period ended and its band table is current — only the arithmetic is
16 months stale. Separately, it charges the 5% surcharge below the £40,000
higher-rate floor: +£2,000 at £39,999.

**Defect family 2 — a 0% rate breaks six of the seven `loans/*` calculators.**
`assets/js/calculators.js` has an explicit `if (rate === 0)` branch; every
`mortgages/*` payment tool calls it and all five pass. Every `loans/*` tool
re-implements the formula inline and none of the private copies has the
branch. Three print `£NaN`; three are gated `if (rate > 0)` and so write
NOTHING, leaving the previous answer on screen. `compare-loans` compares two
NaNs, falls to its `else`, and **declares a 0% loan the more expensive
option**. `consolidation` quotes **£0.00/month** for an interest-free loan.

⚠ **The sharpest statement of family 2 needed no source at all.** Driving the
same final vector by two different routes:

```
standard-calc, 0% APR:  '£143.47' by one route and '£429.81' by another
car-finance,   0% APR:  '£501.78' by one route and '£1222.56' by another
settlement,    0% APR:  '£5,158.11' by one route and '£5,023.84' by another
```

Same numbers in the same boxes, different answer. That check (`determinism`)
exists because the first attempt at detecting staleness — compare against a
single primed reading — MISSED it: the stale figure was the answer to an
intermediate state created halfway through typing the new vector, a
combination the user never entered and the harness never recorded.

**A boundary suite that tested 0% still nearly missed `consolidation`.** Driven
first with a 0% *debt*, it passed — the guarded branch returns 0 and 0 is the
right answer for "interest remaining on a 0% debt". Only a 0% *new consolidation
loan* exposes it, where returning 0 means a £0.00 monthly payment. Testing the
case where a broken guard's output coincides with the correct one produces a
green tick and no information.

**Four things the harness got wrong (all in `WRONG_CALLS.md`):**

1. **My oracle was WRONG about `rate-forecaster`; the page was right.** I
   asserted each window's payment on the FULL original principal over the FULL
   original term and filed 4 FAILs. It amortises the balance REMAINING over the
   term REMAINING — the correct model — and recomputing that way reproduces
   £1,526 and £1,286 to the pound. Kept in the checks as a named wrong answer.
2. **My reporting mechanism downgraded the biggest finding to an advisory.** One
   bucket served both "defensible alternative convention" and "named wrong
   answer", so the first run labelled an EXPIRED TAX RULE a CONVENTION. Split
   into `alt` and `defect_alt`.
3. **`--selftest-parse` found a gap in my own parser**: `£0 / mo Rent`, a live
   `portfolio` reading, was refused — that check would have come back N/A, and a
   check quietly not made reads exactly like a check made.
4. **Two "class C defects" were my harness.** The credit-health walk clicked the
   result panel's "Start Over" (`location.reload()`) and reported a
   non-deterministic wizard; the tracker round-trip reloaded before the notes
   field's 1-second debounce fired and reported that notes do not persist.
   `toolgolden.PRESS_JS` already excludes reset-ish buttons AND SAYS WHY — I
   reused the browser harness and left its hard-won exclusion behind.

**Classification disagrees with the brief's 14/3/6 — mine is 15/3/5**, recorded
rather than silently re-bucketed. `car-finance-calculator` C→A (measured: it
uses the discounted-balloon PCP convention, so it has one right answer);
`portfolio` stays C but its aggregate is checked as arithmetic. **And the brief
is wrong that `loans/overpayment-calculator` shares `/assets/js/calculators.js`
— it does not load that file at all**, which is exactly why it is a zero-rate
casualty while `mortgages/overpayment` is not.

**Controls.** Four, all red on demand; two needed their own criterion corrected
first. `--mutate parse` came out N/A and my first exit rule called that inert —
wrong: the field rejects `£200,000`, `set()` refuses to drive on, and that IS
the "silent 0" the control forbids. Criterion is now "no check may PASS under
mutation". `--mutate crosstool` left 4 PASSes that are NON-TESTS (adjacent
boundary vectors expecting the same figure — £1,500,000 and £1,500,001 differ by
12p) and one where the mutation DID NOT BITE (at £39,999 the borrowed
expectation equals what the buggy page prints). Both are excluded **by name and
printed**, not by loosening the bar.

**Filed to the diagnosis loop** per CLAUDE.md, the structural claim in 224 being
cross-cutting: intake `fe69a7b8-d364-4e12-8039-f93f42a4170c`, run correlation
`3e18a949-8732-4603-b19b-f0c159860fa5`.

### 2026-08-08 (later) — the 090 run finished with NO VERDICT, and the reason is a fleet-wide blind spot

The run above (`3e18a949-8732-4603-b19b-f0c159860fa5`) reached
`current_step='complete'`, `status='COMPLETED'`, work item `complete`, and wrote
**5 `bundle` artifacts and no verdict artifact, no `doc_notes` row**, in ~9 min.

Measured, not inferred:

```sql
SELECT DISTINCT repo FROM code_symbols;   -- gqls/agentchassis   (ONE row)
SELECT DISTINCT substring(path from '\.[a-zA-Z0-9]+$') FROM code_symbols;  -- .go  (ONE row)
SELECT count(*) FROM code_symbols;                                          -- 5755
SELECT count(*) FROM code_symbols WHERE path LIKE '%calculators.js%'
    OR path LIKE '%standard-calc%' OR path LIKE '%loanandmortgage%';        -- 0
```

**The diagnosis agent could not open one file the symptom named.** They are
`.html`/`.js` in the `sites` repo. Its five bundles fetched `page_sections`
rows — the DB half — so it looked, found the ported page records, and never
reached the JavaScript the claim is about.

⚠ **A 090 run on a non-Go artefact terminates as a SUCCESS.** Nothing separates
"diagnosed, found nothing" from "structurally could not look". Same Go-only
index behind `bugs_open/223` (another lane, same day, landmine verifier) — one
index, two consumers, silence wearing a finding's clothes. Landmine filed.
`bugs_open/224` takes CLAUDE.md's stated escape hatch explicitly and says what
was substituted; the determinism evidence needs no source reading at all, which
is why the filing does not depend on the loop.

### 2026-08-08 (night) — bugfix 224 session: the six private annuity copies are gone, fix candidate 1 executed

Owner directed a session ("bugfix 224") at `bugs_open/224`. What shipped, in
order, with the checks that were run before each irreversible step:

- **Six verbatim `loans/*` pages** now load `/assets/js/calculators.js` and
  call the shared engine — `calculateAmortization` (standard-calc,
  compare-loans, stress-test), `calculateOverpayment` (overpayment-calculator),
  and a NEW additive helper `calculateBalloonAmortization` (car-finance; PCP =
  annuity on the balloon-discounted principal, so the 0% branch lives in
  exactly one place). `settlement-calculator` is linear in the rate, so its fix
  is `apr >= 0` + always-write, not a shared call. Every submit now writes the
  DOM — answer or cleared state; the stale mode is dead as a class.
  Display conventions preserved exactly at non-zero rates (standard-calc keeps
  billed-rounded totals; at 0% it shows the exact figures — no £0.20 "interest"
  on an interest-free loan).
- **Pre-flight before deploy** (new `scratchpad/preflight_224.py` pattern —
  local http.server + the vonc_pw venv's Playwright against the EDITED files):
  0% vectors from the bug table asserted, plus each page's default vector
  asserted against the OLD formula computed independently in Python. 24/25,
  and the 25th was the harness: I formatted an expectation to 2 dp where the
  page's `toLocaleString(minimumFractionDigits: 2)` prints up to 3
  (`£448.024`). Live-vs-local on that vector: LIVE == LOCAL byte-for-byte, so
  the page behaviour is unchanged; the 3-dp display is pre-existing and out of
  224's scope.
- **Second harness misstep, same shape**: my tool-1 test wrapper had no
  `<meta charset>`, so the fragment's ✅ decoded as mojibake and a correct
  verdict read as a failure. On this site the red result being the harness is
  the PRIOR (handoff §5 said so); both misseps confirm it.
- **`gate_component_bytes.py` would have destroyed the decomposition.** Its
  --repair compares EVERY page_components row against the whole repo file;
  consolidation's prose-0/prose-2 (writable) would have been overwritten with
  the full 12,865-byte document. Fixed: rows are only comparable when
  `mode='verbatim' AND components=1` (the same predicate as
  `loadVerbatimPageHTML`); assembled rows are SKIPPED loudly. ADO-038's
  "re-run --repair after any builder change" is only safe WITH this fix.
- Gate → --repair (6 rows, UPDATE 1 × 6, none lock-suppressed) → gate green.
  Sites repo `ea72609d6`, Actions run 31282369282, changed-domains line shows
  only this domain.
- **Consolidation tool-1 (locked row)**: both inline copies in `calcRisk`
  replaced with shared calls via direct SQL (deliberate operator arithmetic
  correction through the permanent lock; lock RETAINED; provenance under
  `_provenance.bugfix_224`). Pre-flighted in a wrapper page first (8/8).
  Assemble-only rerender filed via deploy_pages.py; served page byte-identical
  to the substitution prediction; pipeline committed `Rerender:
  loans/consolidation.html` (5b55a1ca4).
- **deploy_pages.py had a latent crash**: psql -tA printed the INSERT command
  tag after the RETURNING row, and the two-line "id" poisoned the poll's
  IN-list. The item WAS filed; the script died polling. Fixed
  (`wid.splitlines()[0]`).
- **Oracle, live, same session**: the seven tools went 23 FAIL → **PASS 77
  FAIL 0 CONV 6 N/A 0** (CONV = standard-calc's pre-existing billed-rounding,
  matched by the comparator). Controls: `--selftest-parse` OK; `--mutate
  expectation` CONTROL OK (16 FAIL, 0 passed under mutation); crosstool/parse
  + full sweep recorded below when done.

### 2026-08-09 (small hours) — 224 verification complete, all green, with the controls

- Remaining controls: `--mutate crosstool` CONTROL OK (28 FAIL, 5 refused),
  `--mutate parse` CONTROL OK (0 passed, 4 refused — the legitimate N/A the
  handoff documented). Nothing passed under any mutation.
- **Full sweep: PASS 166 / FAIL 4 / CONV 6.** All four FAILs are
  `mortgages/stamp-duty.html` — `bugs_open/225`, untouched by this fix. The 11
  calculators.js consumer pages: zero FAILs, so the appended helper regressed
  nothing.
- **Golden (GOLDEN_2026-08-05_prechange.json), self-test passed first**:
  `consolidation` **MATCHES — arithmetic exact** (the decomposed page, the one
  the comparator is FOR). The six verbatim pages each report a single
  divergence on the `content` FIELD SHAPE — the comparator expects the
  decomposed chrome's empty `#content` span, and a verbatim page's `#content`
  is the full wrapper. Control: `loans/loan-vs-savings.html` (untouched
  tonight) reports the IDENTICAL divergence and nothing else — comparator
  scope, not regression. **No numeric element diverged on any page.**
- `verify_site.py`: 3 FAILs, all pre-existing and none mine — `/` flagged dead
  by object-store resolution while Cloudflare serves it at 200 (the known
  directory-index class), and missing `og:url` on the two ASSEMBLED pages
  (the 08-06 render already had zero og:url; my diff changed only the tool-1
  span, proven by served == substitution-prediction byte identity).
- Deploy proof at the artefact: live standard-calc == repo bytes; live
  calculators.js contains `calculateBalloonAmortization` (positive) and the
  served page greps 0 for `Math.pow` (negative); consolidation served ==
  predicted, `Rerender: loans/consolidation.html` = `5b55a1ca4`.

### 2026-08-09 — `--emit-criteria` RUN. 7 tools covered, 10 REFUSED, and the refusal is the finding

Owner asked for the emission the PLAN had gated on "224 and 225 both fixed".
Pre-condition re-measured in-session rather than carried forward from the 225
lane's run: full estate **PASS 170 FAIL 0 CONVENTION 6**.

**Scope, deliberately narrow.** Emitted for the 17 tools the oracle certifies,
on THIS site only. Excluded, with reasons:
- `mortgages/investor.html` — toolgolden structurally cannot certify a ratio
  tool (uniform scaling leaves a yield invariant; the existing LANDMINE). It
  would have tripped the INERT gate and aborted the whole emission, so leaving
  it in would have produced nothing at all.
- **loancalculator.co.uk — NOT emitted, and must not be.** The 08-09 addition to
  `bugs_open/224` records 5 of its 8 pages carrying the live 0% defect. Pinning
  a tool's current answers is exactly what the emit-criteria landmine forbids
  while it is unfixed. (Their lane already has its own criteria dir from
  07-31; those vectors never reach 0%, so they pin correct non-zero answers —
  blind to the defect, not asserting it. Left alone.)

**Result: 7 emitted, 10 skipped.**

```
emitted: standard-calc, compare-loans, interest-rate-stress-test,
         overpayment-calculator, car-finance-calculator,
         settlement-calculator, loan-vs-savings   (52 pinned assertions)
skipped: simple, repayment, overpayment, rate-forecaster, stamp-duty,
         affordability, fee-analyser, bridging-loan, equity-release
             — "pressed button 'Calculate …' has no id"
         consolidation — debt-row inputs are class-selected, no ids
```

⚠ **The refusals are the substantive finding, not noise.** Nine mortgage tools
and consolidation cannot be given platform coverage because the button the user
presses has no `id`, so a criteria step cannot name it — and toolgolden refuses
rather than emitting steps that drive the tool differently from the capture.
**`mortgages/stamp-duty` is on that list**: the tool that was wrong for 16
months is precisely the one that cannot yet be watched unprompted, and the
remedy is one `id` attribute per button. A skipped tool looks identical to a
covered one in the acceptance record — which is why this is written down here
and not left in a terminal.

**Install gate: PASS.** `INSTALL_GATE.sh` (sibling lane's, reused) —
`computed_values: 1`, control `no_horizontal_overflow: 1` in the same exec, pod
`browser-runner-adapter-5479844658-hwvvr`. So a fence would EXECUTE, not skip.
**Not installed** — that is a separate decision, and the gate's step 2 is still
owed per fence (one in-cluster run each, skip list free of "not implemented",
inside the 120s deadline).

#### The pinned values were re-derived from the definitions, and the first attempt was MINE that was wrong

`--emit-criteria` pins whatever the tool currently shows, at toolgolden's
x1/x2/x0.5 vectors — which are **not** the oracle's boundary vectors. "The
oracle is green" therefore does not by itself certify the numbers going into the
acceptance record. New `verify_criteria.py` (lane) recomputes every pinned value
from `oracles.py`.

First run: **6 MISMATCH** of 52, all on the `asym` vector, all on
fractional-term inputs. **The tools were right and my recomputation was wrong** —
same shape as the lane's earlier `rate-forecaster` refutation. toolgolden's
asym vector drives 6.9 / 11.5 / 1.8 / 4.4 YEARS, and these pages compute
`months = years * 12` unrounded, so a 6.9-year term is **82.8 payments**. I had
rounded to whole months. Corrected to the pages' convention: **52 of 52 agree
(±£0.02)**. The six failures are retained above as the evidence that the
verifier can fail — it is not a checker written to agree.

⚠ **Recorded as an OPEN CONVENTION QUESTION, because the fence now pins it.**
Whether 6.9 years should mean 82.8 payments (a smooth interpolation) or 83 (a
real schedule with a smaller final payment) is not settled by the oracle, and
this behaviour is **pre-existing and unchanged by 224** — the old inline copies
and the shared `calculateAmortization` both do `years * 12`. But a future,
reasonable "round the term to whole months" improvement would now FAIL these
criteria and read as a regression. Whoever makes that change must re-emit.

### 2026-08-09 — button ids, re-emit 17/17, and the fences are INSTALLED and proven in-cluster

**Button ids (owner-directed).** Nine mortgages action buttons got
`id="btn-calculate"` (equity-release computes two tools on one page, so
`btn-calculate` + `btn-project`); consolidation's debt-row inputs, its add
button and its generated remove buttons got ids too. Markup only — no handler,
no arithmetic, no displayed value moved. Sites `9bf26db81`; DB rows repaired
(gate 39/39 PASSES); consolidation deployed through its tool-1 row + an
assemble-only rerender, served == prediction.

> **MISSTEP, caught by its own post-check.** My first edit script read each file
> fresh per edit, so equity-release's TWO edits each computed from the original
> and the second write would have silently discarded the first — one button left
> with no id and the emission still refusing the tool with no sign why. Fixed by
> accumulating edits per file, plus an on-disk assertion that every intended id
> is present. **A per-file post-check is what caught it, not the diff.**

> **MISSTEP 2 — my comment was the bug.** `addDebtRow` numbered new rows from
> `children.length + 1`, and I wrote in the comment that ids "stay stable even
> after a row in the middle is removed". The test disproved the comment in the
> same minute: remove row 2, add a row, and you get **two elements with
> `id="d-bal-3"`** — ambiguous selectors, which is the exact thing ids were being
> added to fix. Now a counter that only ever increases.

**Re-emit: 17 of 17, 0 skipped** (was 7 of 17). Every previously-refused tool
came back the moment its button could be named — stamp-duty and consolidation
included.

**Pinned values re-verified: 72/72 agree with `oracles.py`**, now including
stamp-duty's eight (the corrected FTB figures) and simple/repayment.
> **MISSTEP 3, and it is the funniest one.** `verify_criteria.py` collected only
> `fill` steps, so stamp-duty's `#buyerType` **select** was dropped and every FTB
> vector was graded as a standard buyer. It reported a **£5,000 mismatch** — the
> same £5,000 `bugs_open/225` was actually about — against a tool that was
> correct. A checker bug wearing the defect's clothes. Fixed to collect
> `select` too.

**Installed: 17 `doc_plans` fences** (`subject_type='tool'`, `subject_key` =
`pages.name`, e.g. `loans-standard-calc`), via new `install_fences.py`. It makes
three deliberate changes to the emitted JSON, each recorded in its docstring:
`profiles:["desktop"]` (the 120s deadline), DROP container selectors (pinning
`#col-a`'s text would make every copy edit a failed calculator), and add
`page_status_ok`.

**Proven in-cluster, not merely installed** (the gate's owed step):
`loans-standard-calc` 6 passed / 0 failed, `mortgages-stamp-duty` 6/0,
`loans-consolidation` 6/0. Zero `not implemented` anywhere — the only skips are
`@mobile`, which is the desktop gating working as intended.

#### ⚠ TWO FINDINGS THAT OUTLIVE THIS TASK

**1. The runner shares ONE page per (url, profile) across every check — emitted
criteria assume a fresh page per vector.** Consolidation failed 3 of 4 vectors
on `#d-name-2` while the identical steps drove the live page perfectly. Cause:
each vector ends by removing a debt row, and with ids that are never reused the
next vector's adds are rows 4 and 5, so the selectors captured from a fresh page
no longer exist. Fix: `install_fences.py` prepends `{"action":"reload"}` to any
check whose clicks come BEFORE its fills (structure-building), and to no others
— a click after the fills is just the Calculate button, and a reload each would
spend the 120s budget for nothing. Re-run: **6/0**.
**I first blamed a race between the clicks and the inline script parse; driving
the live page at `wait_until="commit"` REFUTED that** before I changed anything.

**2. These fences will only ever run when fired BY HAND, and that is not
obvious.** `tool_acceptance_due` selects on `cc.component_level='tool'` OR
`p.page_type='tool'` (`discovery_checks/tool_eligibility.go:71-92`). Measured on
this site: page_types are `content` (26), `guide` (13), `landing`, `section-index`
— **no `tool`** — and every component is `ported-page`/`ported-prose` at
`component_level='section'`. So **no unattended acceptance run will ever select
these 17 tools.** The fences are correct, installed, and dormant until the site
is decomposed into per-tool components (what the sibling lane did) or its pages
are re-typed. Anyone reading "17 fences installed" without this paragraph would
reasonably believe the calculators are now watched. They are not — they are
*checkable*, in one command per tool.

### 2026-08-09 (evening) — the sibling site, and a fourth checker-was-wrong

Owner asked for the 0% defect fixed "in all the calculators", so this session
took loancalculator.co.uk too. Full account in `bugs_open/224` and the new
`HANDOFF_2026-08-09_continue_here.md`; only the transferable parts here.

- **Told the lane before touching it** (`CONTRIB_2026-08-09b_*` in their dir),
  per the 07-29 ruling that a shared mechanism's other consumers must be told,
  not merely measured. Checked their threads first: copy/voice ("site DONE") and
  `bugs_open/227` — neither arithmetic, no work item on the defect.
- **No shared engine there, deliberately.** Their only shared JS plumbing
  (`assets/js/snippets.js`, already on every page) is generated from the
  **fleet-wide `js_snippets` table — no `site_id`**. Adding a row changes a
  shared mechanism for every site: architecture scope, an RFC, not a bug patch.
  So each tool got its own zero branch, following their 08-03 precedent, and the
  door-closing version is written down as an open decision rather than done
  quietly.
- **`render_tool_row.py`'s control refused to write, and it was RIGHT to.** Its
  default `--control-ref` is `6e8098022` (pre-08-03), which can no longer
  reproduce rows written since. Passing `--control-ref 767681e0d^` — the commit
  that actually produced the stored rows — REPRODUCES on all seven. The lesson
  is the lane's own MISSTEP 2 recurring: **a pinned baseline expires when the
  thing it baselines moves**, and the failure reads as "the renderer drifted".
- **Verified before shipping by driving the DB rows themselves**
  (`probe_zero_rate_rows.py`, 18/18) rather than waiting for the deploy. On a
  consumer-credit site a wrong number should cost an edit, not a live page.

> **MISSTEP 4 — my "no NaN on the page" check matched my own comment.** The
> probe asserted `"NaN" not in page.content()` and failed on three tools whose
> displayed values were all correct: the **fix comments explain the NaN defect**,
> so the detector found its own explanation. This is
> [[prompt-text-poisons-its-own-detector]] in a new costume. Now it reads the
> textContent of `[id]` elements — the things a user actually sees.
> **Four times this session the red result was my harness, not the site.** The
> lane handoff said that prior was high; it is higher than I believed.

### 2026-08-09 — UNATTENDED acceptance enabled (owner decision), by re-typing 16 pages rather than decomposing

The route: `toolEligibilityWhere` accepts EITHER a tool-level component OR
`page_type='tool'` + no tool component + **exactly one** active component. Sixteen
of the seventeen calculators already had exactly one, so a single column change
made them eligible — no decomposition, and the subject keys stay `pages.name`,
which is what the fences are already installed under. `UPDATE 16`; the predicate
now returns exactly those 16.

**Blast radius measured BEFORE the update, not after.** Three things could have
made this a live-site change, and each was checked:

1. **Nav.** `page_type='tool'` IS in `neverPrimaryTypes`
   (`populate_nav_tables_action.go:328`), so typing a page `tool` bars it from
   primary nav — and `/loans/` and `/mortgages/` are **not** in the child-URL
   prefix list, so these pages were primary-ELIGIBLE before. That looked like a
   real nav change. It is not, and the reason is the flags: all 25 pages have
   `in_header=false, in_footer=false`, and both branches end the same way —
   never-primary with no flag is "omitted from nav", and primary-eligible with
   `!InHeader && !InFooter` hits `continue` at :407. Same outcome either way.
   (The site also serves baked-in chrome and has 1 `site_nav_items` row.)
2. **Rendering.** `page_type` is not plumbed into rendering
   (`rerender_single_page_action.go:642` says richer per-type markup "needs
   page_type plumbed through PageInfo first"), and these pages take the verbatim
   short-circuit anyway. Verified after the update: three served pages
   byte-identical to the repo, and zero work items raised in the following 20
   minutes.
3. **⚠ THE AUTO-REWRITER, which is the one that mattered.** A failing Tier-4
   verdict normally raises an `improve_tool` item and hands the tool to
   `tool-improver`, which edits `html_template`. On a fence pinned to an
   independent oracle, **the only way an automated rewriter can go green is to
   change the arithmetic** — on pages quoting consumer credit and tax. So all 17
   fences now carry top-level **`no_auto_fix: true`** plus a reason; a failure
   escalates as `acceptance_stuck` at `needs_human_review` and creates NO
   improve_tool item (`tool_acceptance_actions.go:850-930`). Verified 17 of 17.
   The OTHER auto-fix route, `check_tool_health`, selects
   `WHERE cc.component_level='tool'` and this site has none — so re-typing does
   not expose the calculators to it.

Cadence, so nobody is surprised by the bill: `check_tool_acceptance_due` requires
a current PLAN with a criteria fence, skips any tool with a `tool-acceptance`
doc_note in the last **7 days**, and skips any with an open `acceptance_run`. So
each tool is looked at weekly at most. The three fired by hand today are in
cooldown; the other 13 are due.

**`loans-consolidation` is NOT included** — it has 2 active components, so it
fails the "exactly one" clause. It keeps its fence and stays manual-only
(`tool_acceptance_run.sh … loans-consolidation`) until the page is properly
decomposed with a tool-level component. Worth stating because "16 of 17" is the
kind of gap that silently reads as 17.

### 2026-08-09 (late) — the fence had a SECOND auto-rewrite path, and it was the one I added myself

A blast-radius review after the change (not before — recorded honestly) found a
route I had missed. `no_auto_fix: true` protects **Tier 4 only**. Tier 2
(`check_tool_acceptance`) reads the SAME fence and has **no `no_auto_fix`
support at all** — grep returns nothing. It was held off from these pages solely
by `build_status != 'deployed'` (:189), which is a data condition any deploy
clears, not a guard.

**What made the fence able to fail in Tier 2 was my own `page_status_ok`.** Tier 2
skips every `computed_values` check (:467, "not statically checkable (Tier 4)")
and raises `improve_tool` only when something FAILED (:222). So a fence of pure
computed_values is inert there — and the one convenience check I added for a
clearer error message was the entire fuse.

The blast radius on the other end is not this site: the improve_tool spec carries
`component_id`, which here is the shared `ported-page` shell — **one component
across ~154 pages on three sites** (webdesign 97, this site 39, loancash 18).
`tool-improver` rewrote that shell once already (2026-08-04, off a Tier-2 failure
on webdesign.co.uk); it was later flagged `component_template_corrupted` and the
repair fanned `needs_rerender` across all three sites, which is why 39 of our 41
pages sit `needs_rebuild`.

**Fixed by removing `page_status_ok` from all 17 fences** (reinstalled; 0 of 17
now carry it, 17 of 17 carry `no_auto_fix`). Tier 2 can now find nothing in these
fences it is able to fail, so it can never raise `improve_tool` for these pages —
a guard rather than a data condition. Nothing is lost: a page that does not serve
fails its computed_values checks in Tier 4 anyway, and `verify_site.py` covers
serving directly. Re-proved after the trim: `mortgages-stamp-duty` 4 passed / 0
failed.

Landmine filed. **The transferable shape: a fence is read by TWO tiers with
different rules, and the safety flag only binds one of them.** Ask which check
types the OTHER tier can evaluate before adding any, and ask what `component_id`
the page would hand over if it failed.

### 2026-08-10 — the unattended sweep fired overnight and FOUND A NEW DEFECT on its first run

The sweep selected this site at **03:20:59** and raised **14** `acceptance_run`
items — the 16 eligible tools minus the 2 still in cooldown from yesterday's
manual runs. All 14 completed.

**13 passed. 1 failed: `mortgages-equity-release`** — and the failure was real.

**The safety held, behaviourally, not just on paper.** `improve_tool` items
raised against this site: **0**. The doc_note says it in terms: *"NOT auto-fixed
— this fence declares no_auto_fix"*. Yesterday that flag was an assertion; now it
has been asked a question and answered it correctly.

**The defect.** `calcEquityRelease` had `if(age < 55) { alert(...); return; }` —
a bare return that never touches the DOM. Measured live before changing
anything:

```
fresh page, untouched         dispAge=65   erMaxCash=£0        <- INITIAL MARKUP
age 65 (valid)                dispAge=65   erMaxCash=£124,000
age 32.5 after a valid calc   dispAge=65   erMaxCash=£124,000  <- STALE
age 32.5 on a FRESH page      dispAge=65   erMaxCash=£0        <- markup again
```

That is `bugs_open/224`'s mode 2 exactly — an ineligible input silently leaving
the previous answer on screen — in a tool the 0% work never touched, **because
this guard is on AGE, not on rate**. My 224 note said the stale mode was "dead as
a class"; that was true of the six tools I converted and **false as a general
claim about the site**. Corrected here.

**Two things the failure exposed at once**, which is why it is worth reading
twice:

1. The tool is wrong (fixed: validates, then always writes — £0 against the age
   actually entered; bands 55/65/75/85 re-checked unchanged).
2. **The fence had pinned the page's INITIAL MARKUP as if it were an answer.**
   `--emit-criteria` records whatever the tool displays, and for the half/asym
   vectors (ages 32.5 and 39) the tool displayed nothing — so the capture stored
   `dispAge 65 / £0`, which is the untouched DOM. On a fresh page per vector that
   looked self-consistent; under the runner's shared page it collided with the
   previous vector's real answer and the check failed. **A captured expectation
   from a tool that did not write is not an expectation.** Re-emit for this tool
   once the fix is live, or the fence keeps asserting markup.

⚠ **The transferable check: when emitting criteria, ask whether the tool actually
RESPONDED to each vector.** `--emit-criteria` already refuses a wholly inert tool
(react/vary gates); it has no equivalent gate for a tool that is inert on ONE
vector, and that is what happened here.

### 2026-08-10 (later) — consolidation joined the sweep: 17 of 17. Done by giving it a TOOL-LEVEL component, not by restructuring the page

`loans-consolidation` could never satisfy eligibility branch (b) — that clause
needs **exactly one** active component and this page has two prose rows plus the
tool row. Branch (a) is the one it can satisfy: `cc.component_level='tool'`.

**The choice that made it cheap:** the new component's `function` is
**`loans-consolidation`, identical to `pages.name`.** `toolSubjectKeyExpr`
returns `cc.function` for a tool-level component, so the subject key is unchanged
— the fence installed yesterday still applies, and the acceptance agent's page
lookup (`name IN ($key, 'tool-'||$key)`) still resolves. Had I named the function
anything else (the sibling site uses `tool-consolidation-risk`) I would have
orphaned the fence and broken page resolution in one step.

What was done, in one transaction: created the component
(`Debt Consolidation Risk Checker (loanandmortgagecalculator.co.uk)`,
`component_level='tool'`, template = the tool's own markup **plus a tool-doc
header**), pointed the `tool-1` row at it, set `page_type='tool'`.

Verified: eligibility returns **17 rows**; subject key is `loans-consolidation`
and a fence exists under it; the `tool-1` row is **still `lock_type='permanent'`
with its 5,720 bytes untouched**; the served page is byte-identical to the repo;
a fired run reads **4 passed / 0 failed**; and **0** `improve_tool` /
`needs_tool_recreation` items exist for this site.

⚠ **A tool-level component is a NEW exposure, and `no_auto_fix` does not cover
it.** `check_tool_health` selects `WHERE cc.component_level='tool'` — which is
precisely why the other 16 are invisible to it — so this page is now the one that
IS audited, and that check raises `improve_tool` for `tool-improver`, which
rewrites `html_template`. `no_auto_fix` is a *fence* flag read by the acceptance
tiers; it has nothing to do with this path. Three things make it safe here, and
they should be re-checked rather than assumed:
1. the component is this page's OWN — blast radius one page, not the 154-page
   shared `ported-page` shell;
2. the page is `rebuild_policy='owned'`, and `save_page_sections` hard-refuses an
   owned page, so a rewritten template cannot reach it;
3. the `tool-1` row is `lock_type='permanent'`, so a rerender's computed output
   is discarded in favour of the stored bytes.
The doc header states all of this inline, including that the external
`<script src="/assets/js/calculators.js">` is DELIBERATE and must not be
"fixed" by inlining the annuity back — `check_tool_health` flags it as a
self-containment warning, and that warning is wrong for this tool.

**The chassis build deployed today changes nothing here**: every change this
session was site content, DB rows or lane tooling — no Go — so nothing was
waiting on a build.

### 2026-08-10 — OWNER: unlock both sites. 39 prose pages DONE; the 20 tool pages REFUSED, with the reason measured

Owner: *"unlock them both and make their components and tools fully editable and
upgradable … all through the framework."* Full account + the remaining recipe:
`HANDOFF_2026-08-10_unlock_and_upgrade.md`. RUNBOOK §14.

**The lock is `pages.rebuild_policy`** ('generic'|'owned', migration 164 CHECK),
enforced in four places (build-queue exclusion, reconcile → `owned_page_review`,
`save_page_sections` refusal, `owned_page_guard` on `assemble_page`). Both sites
were 100% `owned`: 41 pages (lmc) + 18 (loancash). Only ONE component lock existed
site-wide (`consolidation`'s `tool-1`, `lock_type='permanent'`).

⚠ **CORRECTION to how this gets described.** "Locked" did NOT mean uneditable.
Migration 164 states that re-assembly of existing `page_components` is
deliberately NOT gated — *"it is how owned pages deploy"* — so `page-rerender` and
`section-editor` already worked. `owned` blocks the GENERIC pipeline rebuilding a
page wholesale. I nearly wrote "the content was locked and is now editable", which
would have been false in both halves.

**Migration 367 applied by hand** (not `--apply`: other threads had pending
files), then `--record-only` with the verification note. 39 prose pages
`owned → generic` (24 lmc / 15 loancash). Stamps every changed row into
`_mig367_unlocked_prose_pages` so the ROLLBACK re-locks exactly those, not
"everything on these domains" — a concurrent thread may flip a page after 367.

**Both verify assertions were INDUCED before applying**, because a verify block of
`SELECT`s cannot stop a COMMIT (`ON_ERROR_STOP` ignores a non-empty result):

```
39 -> 40 population assertion        -> ERROR, aborted
tool pages allowed into the target   -> ERROR "NEGATIVE CONTROL FAILED — a
                                        calculator page has been unlocked and
                                        would be clobbered by the next generic build"
after both inductions: stamp table ABSENT, 59 still owned   <- nothing leaked
```

**Why the 20 tool pages were NOT flipped, and this is the load-bearing part.**
19 of 20 are single-component verbatim (`slot_name='ported-page'`) — ONE row
holding the whole page, calculator `<script>` included. The three composition
loops run `assemble_page → deploy_page(git_commit) → save_sections`, so freshly
LLM-written HTML is **committed to the sites repo the site deploys from, one step
BEFORE** the DB guard refuses (`owned_page_guard.go` header documents this
ordering). `rebuild_policy='owned'` is therefore the ONLY thing between those
pages and a generic rebuild that replaces the calculator with prose — the vonc
arena clobber (TL-001) that migration 164 exists to prevent. Two of them are the
calculators `bugs_open/224`/`225` were just fixed on.

The route is decomposition, not unlocking: prose components + a tool row at
`lock_type='permanent'`, **re-slotted before the flip** (`matchLockedRow` matches
`slot_name`; positional `tool-1` never matches, and `:855` moves an unmatched
locked row to `len(sections)+1` — the calculator lands at the page BOTTOM,
silently). One page at a time with a check between.

**A second precondition nobody had recorded: NEITHER SITE HAS A SITE PLAN.**
`site_plans`/`site_plan_pages`/`site_plan_sections` = 0 rows for both. So 367
removed the refusal and created no demand — the 39 pages are eligible and
undriven. `bugs_open/204` is the trap to read before seeding one.

**Checks re-run after the unlock** (it changes no rendering, but measured rather
than assumed): `oracle.py` full sweep **PASS 170 / FAIL 0 / CONV 6**; controls in
the same session all green; `zero_rate_sweep` 0 of 6 on this site and **0 of 3 on
loancash.co.uk** — the first time loancash's tools have been checked by anything.

**NEW GAP, highest-value next: loancash.co.uk has NO arithmetic oracle**, and two
of its three tools are REGULATORY — `price-cap-checker` and
`true-cost-calculator` both carry the FCA cap literals (`0.8`/day, `£15` default
fee, `100%` total cost) with no external check. That is the exact shape of
`bugs_open/225`: a dated regulatory number in a page nothing verifies. Verify
against CONC 5A, not the page.

⚠ **MISSTEP, this session, worth the line because the output lied.** I appended
this entry with `cd <lane> && cat >> NOTES <<EOF …` in a block whose next command
was a second heredoc for the RUNBOOK. The `cd` failed (I was already in the lane
dir, so the relative path did not resolve), `&&` short-circuited, **the NOTES
write silently never happened — and the RUNBOOK one did**, because it was a
separate command. The block ended `echo done`, which printed. So: one of two
appends landed, exit 0, "done" on screen, and the only tell was that `tail` showed
another session's text where mine should have been. **Never chain a write behind
`cd &&` in a multi-write block; `cd` to an absolute path on its own line, or write
to absolute paths — and verify each append by grepping for its own distinctive
string, not by reading the tail** (on an append-only doc a concurrent session's
entry sits exactly where yours would).

---

## 2026-08-10 (evening) — orientation for the decomposition task, and a four-day-stale blocker

Owner asked for both sites' components to be decomposed so the framework controls
them, and for the handoff to be updated for a fresh thread. Wrote
`HANDOFF_2026-08-10c_continue_here.md` as the single entry point; banners on both
earlier 08-10 handoffs.

**The measurement that defines the job.** 57 of the 59 pages across the two sites
are still `pages.sections = ["ported-page"]` — a single verbatim blob. Only two are
decomposed: `loans-consolidation` (owned, `prose-0/tool-1/prose-2`, one permanent
lock) and `guide-how-loans-affect-mortgage-affordability` (generic, `["prose-0"]`).
So migration 367's 39 "unlocked" pages are unlocked **and still verbatim** —
permission and structure are different things, and the 13:24 handoff did not
separate them. Split: 38 prose + 19 tool pages to convert. `site_plans` still 0.

**⛔ `bugs_open/204` and `bugs_open/189` are FIXED AND LIVE — and were already fixed
four days before the 13:24 handoff cited 204 as the blocker to read before seeding.**
Both were fixed, rolled and behaviourally verified 2026-08-06 (204 at v1.0.1257,
189 proven at v1.0.1259 by 204's own canary). They sit in `bugs_open/` only because
of the owner direction of 2026-08-06 to leave found bugs there. **A file's directory
is not its status.** Re-verified by me at the *current* binary rather than inherited
from the record — chassis **v1.0.1280**, both replicas, started 2026-08-10T15:45:06Z:

```
stored_slot_name                                  1 / 1     (189)
load page slot identities                         1 / 1     (204)
slot_name repeats with different component_ids    1 / 1     (204)
zzz_cannot_exist  (negative control)              0         instrument proven
page-content-writer default_config LIKE '%slot_name_from%'  t   (189 config half)
```

Consequences, both of which change the plan: a decomposed page **can** now be
rebuilt from a plan (build path resolves by `page_components.component_id` first),
and 189's prohibition — *"never fire a build-path run on a page holding locked
rows"* — is lifted, which matters because prose-rows-plus-a-locked-tool-row is
exactly the shape this work creates.

**A jsonb path read said the config key was absent when it is present.**
`jsonb_path_query_array(default_config, '$.workflow.steps.*.config.slot_name_from')`
returned `[]`; `default_config::text LIKE '%slot_name_from%'` returned `t`. Same
row, same moment. This is the "a jsonb PATH read cannot see the shape change
underneath it" trap — had I stopped at the `[]` I would have reported the 189 config
half as reverted and blocked the whole task on a phantom. **Enumerate or text-match
before believing a path read's absence.**

**The re-slot precondition is narrower than recorded.** RUNBOOK §14 and the 13:24
handoff both say a positional `tool-1` never matches `matchLockedRow` and the
calculator is silently moved to the page bottom. Reading the function
(`save_page_sections_action.go:1043`, not the recorded `:855` — the line numbers
have drifted; reposition is `:928`, lock-aware DELETE `:757`): it matches the
incoming section name exactly first, then kebab-normalised, and consolidation's live
`pages.sections` **already contains `tool-1`**. So it matches. The trap fires when
the incoming composition *omits* the tool slot — i.e. a seeded plan with semantic
section names. Precondition restated as *"the composition must name the tool slot"*.
`[INFERRED from code + the live row; UNMEASURED end to end]` — flagged in 10c §6
with the disconfirming result named (locked row at `len(sections)+1` plus a
`lock_blocked` item).

**Smaller correction:** `pages.name` is hyphenated (`loans-consolidation`), not
slashed as the 13:24 handoff's §3 table lists it. My first query used the slashed
form and returned 0 rows — which reads as "no such page", not as a typo.

**Not touched, still owed:** loancash has no arithmetic oracle and two of its three
tools carry undated FCA cap literals (0.8%/day, £15, 100%) — same shape as the SDLT
bug. Independent of decomposition and arguably higher value.

---

## 2026-08-10, evening (~19:20–19:50Z) — migration 377: six live calculators had been unlocked, and 367's own negative control could not have seen them

Picked the lane up from `HANDOFF_2026-08-10c_continue_here.md` with the owner's
"yes, decomposition was what I wanted". Ran 10c §5's three session-start checks
first: `stored_slot_name` / `load page slot identities` / `slot_name repeats with
different component_ids` = **1 / 1 / 1 on BOTH replicas** (`m9fbr`, `swzhc`) with
the `zzz_cannot_exist` negative control at **0**, chassis **v1.0.1280**, and
`page-content-writer`'s `default_config::text LIKE '%slot_name_from%'` = **t**. So
§2's lifted-prohibition claim re-verified at the current binary rather than
inherited. `[MEASURED 2026-08-10 19:20Z]`

**Then the inventory query disagreed with the handoff, and the handoff was wrong.**
Listing all 41 LMC pages by policy and shape, six of the 24 pages migration 367 had
unlocked are in `decompose_lmc.py`'s hand-authored `CALCULATOR_URLS`:
`loans/compare-loans`, `loans/interest-rate-stress-test`, `loans/loan-vs-savings`,
`loans/settlement-calculator`, `loans/damage-checker`, `mortgages/fact-finder`.
367's whole design was to refuse calculator pages. It refused 20 and missed these six.

**Mechanism.** 367 classified with `bool_or(rendered_html ~ 'onclick=|addEventListener')`.
Measured per page: all six are `f` on that expression and `t` on
`oninput=|onsubmit=|onchange=|onkeyup=`; `compare-loans` and
`interest-rate-stress-test` are also `t` on `calculators.js`, the shared external
script — which is where their `addEventListener` calls live, i.e. **outside** the
column the detector read.

**The part worth carrying off this lane** (now in `LANDMINES.md`): 367's negative
control asserted "17 + 3 tool pages still `owned`" **using the same expression as
its filter**. It was written deliberately and it *was* induced — the NOTES entry for
12:22Z records it firing on the induction. It was nevertheless blind to precisely the
population the filter was blind to. Inducing a control proves the `RAISE` fires; it
says nothing about whether the classifier was right. A control has to disagree with
its subject *somewhere* to be a control.

**Why it was live damage waiting rather than an untidy flag.** All six were, at once:
`rebuild_policy='generic'`; `build_status='needs_rebuild'` (what `get_pages_to_build`
selects on — its only ownership filter is `ownedPageExclusionSQL`,
`COALESCE(rebuild_policy,'generic') <> 'owned'`); still a single verbatim
`["ported-page"]` row with the calculator inline; and each carrying an open
`page_rerender:detected` item. And the generic rebuild path has **already run at
these pages**: `needs_page:loans-compare-loans`, created by `page-rerender`
2026-08-08 22:24Z, `attempt_count=1 max_attempts=3`, `error` =

> step save_sections failed: … page loans-compare-loans is rebuild_policy=owned
> (tool/widget-owned): a generic section save would clobber it … Refusing to overwrite.

19 siblings the same. That refusal is the only reason those six calculators still
exist, and 367 removed it for them at 12:22Z. Automated page-level runs against
these exact pages are frequent (`orchestration_states`: `loans-compare-loans`
03:46Z, `loans-interest-rate-stress-test` 03:47Z, `loans-settlement-calculator`
03:50Z today), so this was not a theoretical window.

**One thing that lowered the urgency and one that did not.** `claim_work_item_action.go`
claims only `status IN ('triaged','approved')`, so the 57 `failed` `needs_page` rows
are not themselves re-claimable — the exposure was a *fresh* item or a build run, not
a retry of those. That is a mitigation, not a guard.

**An inherited claim that did NOT survive, recorded because Track B leans on it.**
Both earlier handoffs and RUNBOOK §14 state the loop *"commits LLM-written HTML to
the sites repo one step BEFORE the DB guard refuses"*. Checked the sites repo for the
window in which 20 runs reached `save_sections` — `git log --since '2026-08-08 20:00'
--until '2026-08-09 03:00' -- loanandmortgagecalculator.co.uk/` returns exactly two
commits, the 224 APR fix (23:43) and `Rerender: loans/consolidation.html` (02:34).
**No clobbering commit.** So on the `page-build-handler` path the DB guard fired
before `deploy_page` wrote anything. `[MEASURED for that path and that window only]`
— the other two composition loops are unmeasured, the guard is still what saved the
pages, and nothing should be relaxed on this.

**Fix: `377_relock_six_verbatim_tool_pages_missed_by_367.sql`** (+ ROLLBACK), applied
by hand with `ON_ERROR_STOP=1`, then `--record-only` with the note. Detector ORs
three independent spellings (handlers/listeners/`calculators.js`; form controls;
`getElementById|querySelector`); over all 38 generic verbatim pages on the two sites
**the six match all three and the other 32 match none**, and assertion 1 compares the
stamped set to `CALCULATOR_URLS`, which never read this SQL.

Both controls induced *before* applying:

```
induce 1: expected set claims a 7th page      -> ERROR "stamped set disagrees with
                                                 decompose_lmc.py's CALCULATOR_URLS
                                                 … missing from stamp=[…jargon-buster]"
induce 2: sweep 2 prose pages in AND name them
          in `expected` so only the over-lock
          control can catch it                 -> ERROR "OVER-LOCKING CONTROL FAILED
                                                 … should be 17/15, got 16/14"
state after 3 aborted runs: stamp table absent, 24/15 generic unchanged (no leak)
after apply: the six = owned/needs_rebuild/["ported-page"]; LMC 18 generic / 23 owned,
             loancash 15 / 3 unchanged
```

**My own misstep, caught by the induction and not by re-reading the file.** The first
version of assertion 1 compared **`url` alone**. `pages.url` is not unique across
these two sites — both have `/guides/jargon-buster.html` and `/legal.html` — so
inducing an over-lock of "one" page stamped **two** rows, and the assertion reported
`missing=[-] unexpected=[-]`: it passed on a set it had never actually matched. Now
keyed on `domain || '|' || url`. This is the second time on this lane that the
induction, not the review, found the defect in the checker.

**Net effect on the plan:** Track A is **17** pages, not 23; Track B is **22**, not
16. The six moved tracks, not shapes — nothing was decomposed or undecomposed by this.

---

## 2026-08-10 (late evening) — Track A pre-flight: tooling proven, three new traps, one induced

Owner said go on Track A. Wrote `HANDOFF_2026-08-10d_track_a_prose_decomposition.md`.
Everything below was run read-only against the live DB and the sites repo; the
`--apply` half is deliberately left to the executing session.

**I ran the blind check and it lied to me, exactly as it lied to 367.** Before
learning of migration 377 I checked whether any `generic` page carried a calculator
using `rendered_html ~ 'onclick=|addEventListener'` — the same expression 367 used.
It returned a clean "no tool on any generic page". That was **true only because 377
had re-locked the six 90 minutes earlier**; had I run it at 18:00 it would have
returned the same reassuring answer while six calculators sat unprotected. **Two
checks blind the same way agree with each other**, and I reached for the blind one
by default. The authority is `decompose_lmc.py`'s hand-authored `CALCULATOR_URLS`.

**The safety property for Track A, done properly.** Drove `CALCULATOR_URLS` (23
entries) against all 41 live `pages` rows: calculators not owned = `[]`, generic
pages that are calculators = `[]`, owned pages that are not calculators = `[]`. An
exact three-way partition. Independent corroboration: `decompose_lmc.py` over the 17
reports `with a tool block: 0`.

**The pinned ref is safe for Track A and not for Track B.** `decompose_lmc.py` pins
`b318a8fad`; `load_lmc.py`'s baseline is at `b26fdc81b`. `git diff b318a8fad HEAD`
over the site touches 19 files and **every one is a calculator page or
`calculators.js`** — none of the 17. So Track A needs no re-pin; Track B's pages are
precisely the ones that moved, so it does.

**Tooling proven today:** manifest builds clean (`pages: 17  with a tool block: 0`);
`load_lmc.py --check --all` → 17/17 predicted, exit 0, no REFUSE. Chrome is already
loaded (3 `site_components`, permanent, 08-05), so RUNBOOK §12 step 0 is satisfied
and must not be re-applied.

**⛔ `--check` is a DESTINATION guard, not a CONTENT guard — INDUCED.** I appended an
HTML comment to `legal`'s prose block in a copied manifest and re-ran `--check`: it
**passed**, predicting 9,273 bytes against the clean run's 9,248. Reading
`load_lmc.py:281–305`, the three guards compare title/meta (manifest vs row), stored
row md5 vs the 08-09 baseline, and dropped sections. **None compares the manifest's
prose HTML against the stored page.** Content fidelity comes entirely from
`decompose_lmc.py`'s upstream assertions, which refuse to write the manifest at all.
That is a strong control but a *different* one, and a session that treats `--check`
as a content check will be wrong. `[NOT INDUCED: decompose's own assertions — they
ran clean over all 17 but I did not mutate a page to watch one fire.]`

**Two naming traps.** (1) `--pages` takes the MANIFEST slug, derived from the file
path, which differs from `pages.name` on **14 of the 17**: `guides-jargon-buster` vs
`guide-jargon-buster`, `loans-index` vs `loans`, `mortgages-index` vs `mortgages`.
Feeding DB names in matches nothing, and because `--pages` also suppresses the
"expected 23 tool pages" assertion, an **empty manifest exits 0** printing
`pages: 0` — a silent no-op that reads as success. The mismatch is deliberately
handled downstream (`page_ids()` joins on URL, docstring says a bare-name join was a
real defect); do not "fix" it. (2) `--manifest <file>` also parses its own value as a
page name, so it fails with `no manifest entry for 'manifest_mutant.json'`. Use a
separate `DECOMP_WORK` dir per manifest.

**What Track A actually buys, measured — less than the word "decomposition" suggests.**
All 17 collapse to a single `prose-0` block (by design: consecutive prose merges into
one run), so this is one editable component per page, **not** paragraph-level control.
And it reduces shared-template blast radius without removing it: today every verbatim
page points at one `content_components` row, *"Ported Page (webdesign.co.uk)"*, used
by **154 pages across 3 sites**; after decomposition a prose row points at *"Ported
Prose Block"*, used by **29 pages across 2 sites**. Only tool components get a
page-scoped definition (consolidation's: 1 page, 1 site). The real gains are
per-instance content ownership and genuine rebuildability now that 204 is fixed.

### 2026-08-11 — TRACK A EXECUTION: pre-flight re-measured, and two real defects in the tooling

Picked up `HANDOFF_2026-08-10d`. Everything in its §0–§2 re-measured against the live
system before acting; all of it held. What follows is what the handoff did **not**
have right, or did not have at all.

**Pre-flight, re-measured (all 2026-08-11 morning).** Pod-greps `stored_slot_name` /
`load page slot identities` / `zzz_cannot_exist` = **1 / 1 / 0** on
`agent-chassis-7c9d5f74b9-6j5xn` (v1.0.1284, both replicas, started 09:23:45Z).
`page-content-writer.default_config LIKE '%slot_name_from%'` = **t**. Page census:
17 generic verbatim, 1 generic decomposed, 22 owned verbatim, 1 owned decomposed —
exactly the post-377 table in 10c §1.

> **CORRECTION to 10d §2.** It says the 19 files that moved in the sites repo since
> the pinned ref `b318a8fad` are "every one a calculator page or
> `assets/js/calculators.js`". Measured: **17 calculators + `calculators.js` +
> `guides/how-loans-affect-mortgage-affordability.html`** — the last is a prose page,
> the already-decomposed one. The *conclusion* is unaffected (no Track A page moved,
> so the pin is byte-identical to HEAD for these 17) and the positive control fired
> (`mortgages/stamp-duty` IS in the moved set), so the check was not blind. But the
> stated reason was wrong, and a Track B session reading "every one is a calculator"
> would mis-scope the re-point.

> **MY OWN WRONG CALL, caught by a control I nearly left out.** My first run of the
> calculator/prose partition check printed all three "must be empty" lists as `[]` —
> the exact pass 10d reports — **for the wrong reason**: `CALCULATOR_URLS` holds stems
> (`loans/compare-loans`) and `pages.url` holds `/loans/compare-loans.html`, so the two
> sets could never intersect and every difference was trivially empty. What caught it
> was the third line, `owned pages that are NOT calculators`, which printed **23
> entries instead of `[]`** — an expected-empty that came out full. Had I printed only
> the two lists that mattered to me, I would have recorded a clean partition proof
> built on a string-form mismatch. After normalising, the partition is genuinely
> exact (23 owned == 23 calculators, 18 generic == 17 + the done one), and I induced
> it: dropping one entry from the set moves `owned NOT calculators` to exactly that
> page. **An empty set is only evidence when something could have put a member in it.**

**The content guard 10d left `[NOT INDUCED]` — done, and the answer is better than
inducing it.** Induced five mutations of `legal.html` through `decompose_page`
directly (control: clean). All five refused: footer altered → *"footer differs from
chrome"*; body `<style>` → hard refusal; executable `<head>` script → hard refusal;
inline body script → *"inline script on a non-calculator page"* + *"P-script-bytes"*;
skip-link altered → *"skip-link missing or differs"*. So the refusal machinery is
live, not inert.

**But P-visible cannot fire on ANY Track A page, and that is fine once you see why.**
With no script-addressed ids and `is_tool_page` false, `decompose_page` takes the
`else` branch and emits `inner_html` **whole** as one prose block. `orig_vis` and
`got_vis` are then computed from the same bytes — the assertion compares a value with
itself. Deleting a `<p>` from `#content` produced **NO PROBLEM**, as it must.
So rather than inherit "content fidelity rests on P-visible", I proved the stronger
property directly, for all 17: the manifest's prose block is **byte-identical** to the
source `#content` inner, and those exact bytes are a **substring of the live stored
`rendered_html`**. Decompose is the *identity* on content here — nothing can be lost
inside it. Both arms induced (mutate the source → first arm False; mutate the stored
copy → second arm False).

**Reader-visible text: 17 of 17 IDENTICAL, served-today vs predicted.** Fetched all 17
live pages as a "before" corpus and compared visible text against `predicted/`.
First run said all 17 DIFFER — **that was my comparator**, not the site: `visible()`
does not strip HTML comments and the chrome components carry long authoring comments
the hand-built pages never had. With comments stripped: 17/17 identical, and the
comparator induced (inject a `<p>` into one prediction → DIFFERS; restore → identical).
This is the sixth time on this estate that a red from a same-day checker was the
checker.

**The +2,903 bytes on `legal` (6,345 → 9,248), accounted for**: 1,291 chrome authoring
comments + 566 injected JSON-LD + ~1,046 the layout-shim `<style>` (mostly its own CSS
comment). The structural diff also shows `og:title`/`og:description`/`og:url` gone
(5 og tags → 2) and `lang="en-GB"` → `lang="en"`. **Both are already-stated accepted
losses** — `PLAN_2026-08-05` §6 lists per-page `og:*`, nav `aria-current` and the
`lang` change, "each is visible in the assemble-mirror diff and none is silent". I
re-derived them independently, which is the record working as intended. Confirmed
live: the two already-decomposed pages serve `og:2 / lang="en"` today, a verbatim
calculator serves `og:5 / lang="en-GB"`. Track A does not introduce this; it takes it
from 2 pages to 19.

#### ⛔ DEFECT 1 — `backup_everything()` could not run at all (schema drift)

`--apply legal` died on its first call with *"INSERT has more expressions than target
columns"*. `page_components` has gained **`rendered_html_digest`** since the backup
table was cloned `LIKE page_components` on 08-05 — 28 columns into 27. **It failed
before writing anything** (the backup runs first), which is the only reason this was
an inconvenience rather than a half-applied page.

The asymmetry is worth keeping: the **restore** direction was never broken. Fewer
expressions than target columns is legal SQL — the trailing ones take their defaults —
so 27-into-28 succeeds. Proved it rather than assumed it, inside a `BEGIN … ROLLBACK`,
and `legal` came out untouched at the baseline md5. **Drift breaks the backup loudly
and the restore not at all**, so a session that only ever tested `--restore` would have
found nothing wrong.

Fixed in `load_lmc.py`: the INSERT now names the columns the backup table actually
has (qualified `pc.`-side — the source joins `pages`, so a bare `id` is ambiguous and
that cost a second failed run), so a future added column degrades to "not captured"
instead of "nothing can be backed up". `ALTER TABLE … ADD COLUMN rendered_html_digest`
so the backup is complete going forward.

#### ⛔ DEFECT 2 — the backup table was poisoning its own rollbacks, silently

The old guard was **per-ROW** (`NOT EXISTS … b.id = pc.id`). So once a page is
decomposed, the *next* `--apply` of **any** page sweeps that page's new `prose-0` row
into the backup **beside** its original `ported-page` row. `--restore` then replays
both, giving one page a verbatim blob *and* a prose section — the nested-`<html>`
corruption this file's own docstring warns about, arriving through the mechanism meant
to be the safety net.

**Already materialised**: `/guides/how-loans-affect-mortgage-affordability.html` held
**2 rows** in the backup (`ported-page` + `prose-0`) before I touched anything. 42 rows
over 41 pages. Applying Track A one page at a time would have poisoned ~16 more
rollbacks, and nothing would have said so until someone tried to roll back.

Guard is now **per-PAGE**: a page with any row in the backup is already snapshotted and
is never re-captured. Repaired the existing damage in one transaction with a `DO`/
`RAISE` verify block — stray preserved to `page_components_bak_strays_20260811` first,
never just deleted — leaving **41 rows over 41 pages, one verbatim generation each**.

**Then proved the rollback end to end on `legal`, before the other 16**: apply →
`["prose-0"]`, restore → `["ported-page"]` at md5 `7265952aac43b36361cf6aebf0e6580b`,
**equal to the 08-09 baseline file byte for byte**, re-apply → `["prose-0"]`. The
safety net is now measured rather than assumed, on the cheapest page to be wrong about.

**Also fixed: `deploy_pages.py` read `manifest_voiced.json`,** which the 2026-08-06
owner ruling superseded and which `load_lmc.py` no longer writes predictions from.
Deploying from a different manifest than the predictions came from is how a byte-diff
passes against a prediction it does not correspond to. It now prefers `manifest.json`
and falls back to the voiced file only if that is absent. No `--manifest` flag on
purpose — this tool derives page names as "argv entries not starting with `--`", so a
flag value would be parsed as a page name (the trap 10d §4.2 records for `load_lmc.py`).
