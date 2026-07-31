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
