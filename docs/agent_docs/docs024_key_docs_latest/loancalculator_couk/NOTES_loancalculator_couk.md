# NOTES — loancalculator.co.uk adoption

Append-only, newest at the bottom. Evidence, commands, what the system actually
said — and every misstep.

---

## 2026-07-30 — session start: find the site, find out why it is dead

Owner asked to look at `~/projects/domains/loancalculator.co.uk`, then
`~/projects/sites/loancalculator.co.uk`, then to adopt it into the framework.

**Two copies, different roles.** `domains/` is the source: Go `main.go` +
`templates/` + `data/*.json` → `public/`. `sites/` is the deploy checkout of
`gqls/sites` (branch `master`, one dir per domain). `diff -rq` of
`domains/public` vs `sites/loancalculator.co.uk`: identical but for `sitemap.xml`
living one level up in `domains/`. So the deployed copy is not stale relative to
source — the whole thing is just frozen at 2026-03-20.

**The site was dead, and the signature misled.** `curl -sI https://loancalculator.co.uk/`
→ **exit 28, timeout, no status line at all.** DNS resolved to Cloudflare proxy IPs,
TLS handshake completed, HTTP/2 request sent, nothing returned. Compared against
`gamesdesign.co.uk` (200, with `x-amz-*` headers ⇒ B2 origin behind Cloudflare) and
against `/worker-health`, which healthy zones answer `200 "Worker is running!"` from
`scripts/cloudflare/worker.js`. loancalculator.co.uk did not answer it ⇒ **no worker
route bound on the zone**, so requests fell through to the zone's configured origin —
the legacy `s3://loancalculator.co.uk` from the old
`domains/.../deploy/Dockerfile`, long switched off.

Owner bound the route. Verified serving at **15:11:50Z**: `/worker-health` 200, `/`
200 `content-type: text/html`, `/tools/standard-calc.html` 200. The site is up.

**No platform record at all.** `SELECT * FROM sites WHERE domain ILIKE '%loancalc%'`
→ **0 rows**. Not inactive, not pooled — absent. (32 rows exist: 14 deployed
domains, 17 `.internal` pool rows, 1 system.)

## 2026-07-30 — reading the adoption path, and what it would have done to this site

The interesting finding, because it is the owner's actual question.

- **`--fidelity` is decoration.** The script says so itself
  (`082_submit_domain_unified.sh`, NOTE at ~line 50: "RECORDED in input_data … does
  not yet modulate the build … forward-looking, not active"). I did not take the
  comment's word for it — `grep -rn fidelity --include='*.go' platform/ internal/
  pkg/` returns **10 hits, every one unrelated**: a vet-med price parse guard
  (`vet_med_price_scrape_action.go`) and a comment in `platform/aiservice/gemini.go`.
  Zero consumers. [VERIFIED myself, not inherited.]
- **URLs would all have changed.** `apply_adoption_plan_action.go:446` runs every
  page through `datahelpers.CanonicalisePage`, which synthesises
  `/tools/<slug>/index.html`. This site's URLs are flat `/tools/<slug>.html`. The
  crawled URL — which the analyze LLM even emits in its plan — is discarded.
- **The calculators would have been rewritten.** `mode:"recreate"` at `:517`,
  `:625`, `:633`; interactive pages route to `needs_tool_recreation` at `:627`.
  Twelve working, tested calculators handed to an LLM to regenerate.

So the mend is to make `fidelity=locked` mean something (M1) and give it a
byte-exact deploy path (M2). Both opt-in.

**Reference implementation, with a correction to my own earlier read.** I had
assumed `cmd/webdesignport` was a full-page verbatim porter and could simply be
pointed at this site. It is not: it strips each source page's chrome and stores a
**body fragment** (`<section class="ported-page">`), which then relies on the
chassis assembling `site_components` head/header/footer around it. Fine for
webdesign.co.uk, which got a full LLM design stand-up; wrong here, where the point
is that the served bytes do not change. That is why M2 exists at all — without it,
"owned" pages still go through `assemblePage` ([VERIFIED] `rerender_single_page_action.go:101`,
then `StripToolDocHeader` `:153`, `repairOutboundPageLinks` `:161`) and come out
re-wrapped. The useful half of the precedent is the **row shapes**
(`import.go:177-236`: `rebuild_policy='owned'`, one `ported-page` component,
`build_status='approved'`).

Good news downstream: `rerender_single_page_action.go:381` sets
`Filename = TrimPrefix(URL,"/")`, so once `pages.url` holds the flat path, the
deployed filename follows. URL preservation is a one-place fix, not a chase.

**One passthrough component exists, and it is a trap to duplicate.**
`SELECT … FROM content_components WHERE function='ported-page'` → exactly one row,
`a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef`, "Ported Page (webdesign.co.uk)", active.
Lookups in the precedent take `ORDER BY created_at DESC LIMIT 1`, so seeding a
second would silently repoint every future port. Reuse it. → RUNBOOK §3.

## 2026-07-30 — the site's own defects (found by looking, not by symptom)

Nobody had been looking at this site, so nothing had been reported. Counted from the
files:

- **4 of 28 pages are bare fragments** — no doctype at all, therefore no head, no
  stylesheet, no nav: `tools/credit-roadmap.html` (28 lines),
  `tools/damage-checker.html` (32), `guides/fixed-vs-variable-loans.html` (15),
  `guides/hidden-loan-fees.html` (16). Verified by testing every file for
  `<!DOCTYPE>`.
- **10 of 28 pages have no navigation** (no `#nav-placeholder`, no `nav.js`) —
  the 4 above plus `tools/standard-calc.html`, `tools/overpayment-calculator.html`,
  `guides/car-finance-explained.html`, `guides/secured-vs-unsecured.html`,
  `guides/uk-lending-landscape.html`, and `404.html` (fine to leave bare). The
  flagship calculator and the real overpayment tool are both in that list: a visitor
  landing there has no route into the rest of the site.
- **`nav.js` points at a page that does not exist** —
  `/tools/tool-overpayment-calculator.html` (real file:
  `overpayment-calculator.html`). Because the nav is injected on every page that
  loads it, that is a dead menu entry site-wide. `sitemap.xml` carries the **same**
  phantom URL, and omits three real pages (`overpayment-calculator`,
  `interest-rate-stress-test`, `credit-roadmap`).
- **One page has genuinely broken styling since March**:
  `tools/standard-calc.html` loads `assets/css/style.css` page-relatively from inside
  `/tools/`, i.e. `/tools/assets/css/style.css` → 404. `index.html` uses the same
  style but sits at the root, so it works by accident.
- Dead weight: `assets/js/pdf-gen.js` and `assets/js/search.js` are referenced by
  **no** page; `assets/css/style.css.1`, `main.go.{1,2,3}` are leftovers.
- `global.js` reads an optional `DYNAMIC_MARKET_AVG` global with a 8.2 fallback;
  **nothing defines it** on any page (0 hits) — harmless, but it means the
  "rate context" wording always uses the fallback, not the site's own 7.9% figure.

Decision recorded in PLAN as D5: repair these in Phase A **before** adopting.
Preserving them verbatim would be preserving breakage, and since the pipeline learns
the site by crawling it, defects captured now would be frozen in the DB.

## 2026-07-30 — misstep to record

I initially planned this as a webdesignport-style bespoke port and had a full plan
built around that before the owner's second message reframed it. The plan was not
wrong so much as **answering a smaller question than the one that mattered**: a
bespoke port would have rescued this site and left the framework exactly as unable
to adopt the next one. Worth noting because the pull toward "just port it" is
strong — it is less work and it looks like progress. The second message is what
turned a site rescue into a framework mend; a session that had not asked would have
shipped the smaller thing and called it done.

Also: my first read of `webdesignport` (full-page verbatim) was wrong, as recorded
above. Caught by actually reading `transform.go` rather than trusting the row shapes
in `import.go` to imply what was stored in them.
