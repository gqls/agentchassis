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

## 2026-07-30 — Phase A done: site repaired and live

Commit `b4302e22b` in `gqls/sites`, pushed 15:22Z; `Deploy to B2` run 30556365792
green in 23s (B2 sync + Cloudflare purge both ran).

**Verified live, all of it:**
- **34/34 URLs return 200** — 27 sitemap pages + redirect stub + 404 + 3 assets +
  robots + sitemap.
- **The three deleted files now 404** (`pdf-gen.js`, `search.js`, `style.css.1`).
  This is the load-bearing check: a 404 on a deleted path proves `b2 sync --delete`
  actually propagated, where an all-200 sweep would have passed just as happily
  against a stale bucket.
- **nav.js: phantom marker 0, correct marker 1.** Checked both directions per the
  fleet practice — a DELETE-marker (the string the change removed) is the strongest
  evidence, because it cannot pass against the old file.
- **standard-calc.html**: absolute CSS ref present, old page-relative ref **gone**,
  nav present, `id="monthly-display"` still there. It has been serving unstyled
  since March; it is styled now.
- **credit-roadmap.html**: doctype and title present — it was a bare fragment.
- **sitemap**: 27 `<loc>` entries, phantom URL absent.

**Two things I got wrong on the way, both worth recording:**

1. **My first link audit under-reported the dead links.** I grepped
   `href="(\.\./|\./)?[a-z0-9-]+\.html"` — one path component only — so it could not
   see `../tools/overpayment.html` or `../tools/consolidation-calculator.html`. It
   reported 2 relative links and no dead ones; the truth was 3 dead targets across
   `404.html`, `guides/can-i-overpay.html` and
   `guides/debt-consolidation-explained.html`. **The regex answered a narrower
   question than the one I asked, and returned a clean-looking result** — exactly
   the shape that gets believed. Caught only because I re-ran it broadly (every
   `href` minus external/anchor) and then *resolved each target against the
   filesystem* instead of eyeballing the list.
2. **A `sed` that silently did nothing, twice.** `sed -i 's|(href|src)="..."|...|g'`
   — `|` was both my delimiter and the alternation character, so the pattern was
   truncated. Exit status was fine and 21 files were left untouched. My "verify"
   step then printed both old and new refs and I nearly read past it. Second attempt
   with `-E` failed the same way for the same reason; only using `#` as the
   delimiter worked. **The tell was the verification output contradicting itself**
   (old and new forms both present) — if I had only grepped for the new form I would
   have seen 23 hits and called it done. Check the *old* form is gone, not just that
   the new one exists.

**Caveat on the redirect stub — it is not an HTTP redirect.**
`/tools/tool-overpayment-calculator.html` returns **200**, then moves the browser via
`<meta http-equiv="refresh">` + `location.replace()`. `curl -L` does **not** follow
it (curl follows 3xx only), so a verification that expects `url_effective` to change
will read as a failure when the stub is working correctly. Search engines get
`noindex` + `rel=canonical`. A real 301 is not available to us: static B2 hosting has
no redirect layer and the platform's `redirects` table has no consumer (G5).

**Judgement calls made, flagged for the owner rather than buried:**
- Added an `<h1>` to `damage-checker.html` (it had only an `<h3>`) and a lead
  sentence, so it matches its sibling tool pages. That is content I wrote, not
  content I preserved.
- `overpayment-calculator.html` was self-styled with an inline `<style>` block and no
  shared stylesheet. I added the shared sheet **before** its inline block so its own
  rules still win — appearance should be unchanged where it defined something, and
  inherit site styling where it did not. Also fixed `shadow:` → `box-shadow:` (a
  no-op property, so the card had no shadow). **Not visually verified** — worth an
  eyeball [UNVERIFIED: rendered appearance].
- Fixed a markdown artefact in `hidden-loan-fees.html` (`**APR**` was rendering as
  literal asterisks).

## 2026-07-30 — Phase B: the mends built, tested, committed, submitted

Commit **`e6a8bb63b`**. Council submission **corr `f9eae63e-05fb-40c8-b60c-1670c5681cbe`**.

**M1** — `adopt_verbatim.go` (new) + the `fidelity == fidelityLocked` branch at
`apply_adoption_plan_action.go` §3a. **M2** — `loadVerbatimPageHTML` + the bypass in
`rerender_single_page_action.go`. `insertPageRerenderItem`'s parameter widened from
`*sql.DB` to a 2-method interface (precedent: `datahelpers.DocRekeyer`) so the one
canonical `page_rerender` INSERT could be called inside adoption's transaction —
the alternative was a second copy of that row shape, which is the exact drift its
own doc comment exists to prevent.

Registered in the SAME commit: **ADO-037** (new) + **ADO-011** updated to record
that its `locked` rung is no longer inert. Two **LANDMINES** appended and synced.

**Two discoveries while reading, both of which changed the design:**

1. **`cmd/webdesignport` is not a verbatim porter.** `transform.go:406` writes
   `<section class="ported-page">` — a body FRAGMENT that relies on the chassis
   assembling chrome around it. I had planned to reuse its path wholesale. If I
   had, the adopted pages would have been re-wrapped in platform chrome and
   `lang="en-GB"` would have been replaced — i.e. the one thing the owner asked
   for (unchanged tools, unchanged pages) would have quietly failed. M2 exists
   only because I read `transform.go` instead of inferring its behaviour from the
   row shapes in `import.go`.
2. **The crawl index stores one page under 2–3 keys.** `buildCrawlPageIndex`
   registers the same `*crawlPageContent` under the absolute URL, the path-only
   form and `sourceURL`, so `matchCrawlContent` can find it however the LLM plan
   spells it. Enumerating those keys — the obvious way to get "the list of pages"
   — would have created a page row per alias: ~60–80 rows for a 27-page site,
   every one valid-looking. Now deduped by content POINTER, keys sorted first
   because Go map order is random and an unsorted pick would write a different
   `pages.url` on a re-run. → LANDMINES.

**Mutation-verified rather than assumed green.** Both critical guards were broken
on purpose to confirm the tests catch them:
- removing the `componentRows > 1` refusal → `TestLoadVerbatimPageHTMLRefusals/
  verbatim_but_extra_components_attached` fails. Caught.
- forcing the root case to return `"/"` → `TestURLToDeployPath` fails on
  **`path "/" yields an EMPTY deploy filename`**, i.e. on the invariant, not on a
  spelling. Caught.
- **A first mutation attempt was WORTHLESS and I nearly counted it as a pass:** I
  removed `p == "/"` from the guard, the suite stayed green, and for a moment that
  read as a test gap. It was not — the directory-join fallback still produced
  `/index.html`, so the mutant was behaviourally identical to the original. A
  mutation that does not change behaviour proves nothing about the test. The
  sharper mutation (return `"/"` AND disable the join) is the one that counted.

**MISSTEP — I wrote `Council-Submitted: pending` in the commit trailer.** The
submission had not been made when I committed, so I had no correlation to write and
put a placeholder. `098` resolves that trailer by looking the correlation up, so
`pending` resolves to nothing: the commit will read as un-reviewed for ever, which
is precisely the hole `Council-Submitted:` was added on 2026-07-30 to close. And
forward-only forbids the amend that would fix it. **The correct order is: submit
FIRST, take the correlation, then commit with it.** The real correlation for
`e6a8bb63b` is recorded here and in the RUNBOOK; if the verdict is APPROVED it will
also go on whatever commit follows, so the trail exists somewhere findable even
though the join will miss.

**Not rolling the chassis yet, deliberately.** A roll kills an in-flight council
run, and this one has ~30 minutes to go. The code is committed (so any other
session's roll ships it — both paths are opt-in and inert, so that is safe), but I
will wait for the verdict before building, rather than destroying my own review.

## 2026-07-30 evening — the gap that would have made the whole mend silently useless

Two events, one of them a genuine find.

**1. Another session's roll killed my council run.** Chassis pods restarted
**17:33Z** onto `v1.0.1211`; my council orchestration
(`b9a6ac6c-610a-42ab-8a25-36a636ba1756`) last advanced at **16:57Z** and sat at
`EXECUTING_STEP / review_bug_historian` with **0 awaited requests** — the pod
executing the step died and nothing resumes it. I had deliberately not rolled for
exactly this reason; that protects the run from *me*, not from the fleet. Resubmitted
with `RESUBMIT_CORR` so the trail accumulates under the same submission correlation
(`f9eae63e-…`); the new run is `council-gate-orchestrate-0730-1930`.

**Silver lining, and it needed checking rather than assuming:** v1.0.1211 was built
*after* my commit, so my code is already live. Verified properly per the fleet
practice — **four strings the change ADDED** (`adoption-locked/1`,
`verbatim adoption found no crawled pages`,
`verbatim page, shipping stored bytes unmodified`, `refusing to deploy an empty page`)
each present, a **positive control** (`Built crawl page index`) present, and a
**negative control** at 0 — on **both replicas**. So no roll of my own is needed.

**2. THE REAL FIND — `--fidelity` never reaches the code that consumes it.**
I was about to fire the adoption when I checked whether `input_data.fidelity`
actually arrives at `apply_adoption_plan`. It does not.

`site-adoption-orchestrator` does spawn→call→complete, handing work to the spawned
`site-adoption-agent` through `call_agent`'s `input_mapping` — and **`input_mapping`
is an ALLOW-LIST, not a passthrough.** From `input_contracts/input_mapping.go`:
*"Key = destination field name (what child receives) / Value = source path in
CollectedData"*. The live `call_adopter` mapping enumerates exactly four fields:
`url?`, `domain?`, `target_url?`, `destination_domain?`. **`fidelity` is not one of
them.** So it reaches the orchestrator and is dropped at the spawn; the agent that
runs `apply_adoption_plan` never sees it, `adoptionFidelity` returns `""`, and the
run takes the **recreate** path — rewriting all 12 calculators and changing all 28
URLs — while reporting success.

**Had I not checked, this is exactly the failure I have been writing landmines
about all day.** Nothing errors. The flag is accepted by the script, recorded in the
submission, visible in the orchestrator's `collected_data`, and absent one hop
later. My Go code would have been correct, tested, mutation-verified, council-
reviewed — and inert. The verification I had planned (per-page sha256) would have
caught it eventually, but only *after* the site had been rebuilt.

Why it was invisible before today: **the value was inert at BOTH ends**, so a
missing hop between them changed nothing observable. `grep`ping for consumers found
none, which is what made the dial look like a pure "no consumer" problem. It was two
problems — no consumer *and* no plumbing — and fixing only the half you can see
leaves a mend that cannot fire. **[LESSON] When you make an inert parameter live,
trace its whole path from entry point to consumer, hop by hop. "Nothing reads it" and
"nothing carries it" look identical from the consumer end.**

Filed as **G9** and written as a migration rather than a hand-patch:
`docs/agent_docs/sql_for_agents/274_adoption_forward_fidelity_to_spawned_agent.sql`
(snapshot-first, asserts the snapshot holds the PRE-change mapping, asserts one live
row changed and that the key landed at the intended path — `create_if_missing=true`
is correct here because the key is new, which also means a mistyped path would
silently create a useless key instead of erroring).

**NOT YET APPLIED — blocked.** The direct `psql` write was refused by this session's
permission classifier, so the change is prepared and reviewable but not live. The
adoption run is held until it is applied: firing now would silently take the recreate
path, which is the one outcome the owner explicitly ruled out.

## 2026-07-30 late — the adoption RAN, the mechanism WORKED, and the SOURCE was wrong

The run completed and every structural claim held. Then the fidelity gate failed
27/27, which turned out to be the most valuable result of the day.

### What worked, verified

`fidelity=locked` reached the spawned agent (`locked`, confirmed live — migration
274 does its job) and `apply_adoption_plan` took the verbatim branch:

- **27 pages, every URL preserved EXACTLY** as its flat `.html` path
  (`/tools/standard-calc.html`, not `/tools/standard-calc/index.html`). Under the
  recreate path all 27 would have changed.
- **`rebuild_policy='owned'`, `build_status='deployed'`** on all 27; page types
  correctly derived (13 guide, 12 tool, 1 landing, 1 content); names mirror
  `CanonicalisePage` prefixes.
- **27 `page_rerender` items and ZERO `needs_content_page` / `needs_tool_recreation`.**
  The 12 calculators were never handed to an LLM.
- **27 `ported-page` components**, all `approved`, all
  `deploy_mode=verbatim`, `generator=adoption-locked/1`, 10–16KB each (whole
  documents, not fragments), exactly one component per page.

### What was wrong: the crawl is not a byte source

**`md5(rendered_html)` vs the served file: 0 of 27 matched.** Every stored page was
LARGER by **8,900–9,060 bytes** — and a near-CONSTANT delta across 27 different pages
is a signature, not noise. Diffing one page showed why: firecrawl's `rawHtml` is the
**serialised post-JavaScript DOM**, not the origin's bytes. `nav.js` had already run,
so ~9KB of generated nav sat inside `#nav-placeholder` with its `<style>` hoisted into
`<head>`; every relative URL had been rewritten absolute; the dropdown toggles'
`href="#"` had become `https://…/page.html#`, which **reloads the page on desktop
click** where the original did nothing; `<meta charset>` had been swapped for
`http-equiv`; whitespace collapsed; `&` → `&amp;`.

So the mend's MECHANISM is right and its INPUT was wrong. `formats: ["rawHtml"]`
sounds like it returns what the server sent. It does not. → **LANDMINES**, and
**G10**.

### The incident, and what caught it

Three items drained before I cancelled the rest, deploying the mutated form to the
live site: `car-finance-explained` 2,223→11,212, `debt-consolidation-explained`
3,430→12,465, `debt-help-uk` 2,954→11,932 bytes. Restored from `b4302e22b`, verified
byte-exact at the origin, all 27 URLs 200. Remaining 24 items cancelled with the
reason recorded on the row.

**The checksum gate is what caught this, and it is a MANUAL step, not code.** The
render_guardian seat objected in round 1 that my claimed sha256 gate was
*"aspirational, not built — only a sha256 field stored in content_data"*. **That
objection was correct, and running the check by hand is the only reason three pages
were damaged instead of twenty-seven.** It belongs in the pipeline.

### A second landmine found during the recovery

The 3-file restore deployed GREEN and printed only **2** `upload` lines.
`b2 sync --delete --skip-newer` **silently skipped `debt-help-uk.html` because the
bucket copy was newer** — written 15 seconds earlier by the very rerender I was
undoing. A cache-buster proved the ORIGIN was stale, not the CDN, so waiting would
never have fixed it. `gh run rerun` fixed it (fresh checkout ⇒ fresh mtimes ⇒
upload). **The revert case is exactly when this fires, because what you are undoing
is by definition the most recent write.** → LANDMINES.

### Repair, and the M2 proof

Loaded the **served bytes** from the deploy repo into all 27 components (base64 →
`convert_from(decode(...))` to avoid escape-mangling; the lock predicate respected).
Gate re-run: **27/27 byte-exact, and each component's recorded `sha256` independently
matches the file's real sha256.** `content_data.source` now records that the crawl
rawHtml was discarded and why.

> **First attempt failed and rolled back cleanly:** `sha256(text::bytea)` is invalid
> (`sha256` takes bytea; `decode(...,'base64')` already IS bytea). `ON ERROR STOP` +
> an uncommitted `BEGIN` meant nothing was written — the mismatch report I ran
> immediately after was measuring the UNCHANGED rows, which is what told me the
> transaction had rolled back rather than half-applied.

**Then the decisive end-to-end test of M2.** Raised one `page_rerender` for
`/tools/settlement-calculator.html` (spec shape copied from a completed row of the
same item_type, per the LANDMINES rule — `page_id` in the spec AND the column). It
completed, git-adapter committed `765c0e0d2 Rerender: tools/settlement-calculator.html`,
and **the diff across that commit is EMPTY**; the file is still 2,521 bytes. An owned
verbatim page redeploys **byte-identical** content. That is the rebuild-safety
property M2 exists to provide, demonstrated in production rather than argued.

### Where this leaves `fidelity=locked`

Sound, and **not yet safe to use on a site whose fidelity you have not checked by
hand**, because its byte source is a browser-rendered DOM. Three ways to close it,
for the owner to choose (G10 in the PLAN):
1. **Adopt from files/repo** — the G1 gap I deferred as optional. It is now revealed
   as the *natural* source, not a convenience: the deploy repo already holds exactly
   the bytes being served.
2. **A raw-fetch step for locked fidelity** — plain HTTP GET, no browser, instead of
   firecrawl.
3. **Build the checksum gate into the action** — compare stored bytes against a
   direct fetch before queueing any deploy, and fail loud. This one is worth doing
   whichever source wins, because it is the check that saved this run.

---

## 2026-07-31 — from the `loanandmortgagecalculator_couk` lane (not this lane's work)

I copied this site's 11 working tool files to seed a new combined site, so I read
this directory and your `git log` first. Three things you will want, none of which
needs anything from you.

**1. Your council verdict landed, and the roll did not kill it.** Your README says
you were holding off deploying to protect the run. A chassis roll to `v1.0.1211`
happened anyway at **2026-07-30 17:33Z** (another session), and your verdict came
through afterwards regardless:

```sql
SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='f9eae63e-05fb-40c8-b60c-1670c5681cbe' ORDER BY created_at;
--  2026-07-30 16:50:51  fix_plan
--  2026-07-30 19:30:03  fix_plan
--  2026-07-30 19:43:05  council_report   revise
```

**REVISE**, at 19:43Z — about two hours after the roll. So the "a roll kills an
in-flight council" rule in your RUNBOOK §9 did not fire here. Worth knowing before
you spend another run on the assumption that it did.

**2. `fidelity=locked` is confirmed live in the pod**, if you have not re-checked
since the roll. Pod-grepped on three strings `e6a8bb63b` added plus a positive
control, on `v1.0.1211`: `Verbatim adoption complete` 1, `verbatim_adoption_deploy`
1, `apply_adoption_plan` 2. (Note `grep -c fidelity` proves nothing — there were
~10 unrelated pre-existing hits.)

**3. Three live defects on loancalculator.co.uk, found by browser-auditing your
tools before copying them.** Evidence is a real-browser run of
`webdesign_tools_repair/toolaudit.py` against your live URLs, and `evalpage.py` by
hand where the verdict needed checking:

- **`tools/credit-health-check.html` is broken on the live site.** It is a five-step
  wizard: its script moves the class `active` between `<div id="step-N"
  class="check-step">` elements and relies entirely on CSS to hide the steps you are
  not on. **`assets/css/style.css` defines neither `.check-step` nor `.active`**, so
  all five steps render simultaneously and the tool is unusable. The script is
  correct — this is not findable by reading it. Two rules fix it:
  `.check-step{display:none}` and `.check-step.active{display:block}`. Verified on my
  copy: clicking step 1's button now takes `step-1` `block→none` and `step-2`
  `none→block`.
- **36 classes your tools use are undefined in that stylesheet** —
  `.fca-style-warning`, `.market-context-box`, `.comparison-grid`, `.score-meter`,
  `.stat-value`, `.progress-bar`, `.verdict-text`, `.type-btn`, `.debt-row` and 27
  others. Mostly cosmetic; `.check-step` was the one where the absence was
  load-bearing. Reproduce with:
  `for c in $(grep -oh 'class="[^"]*"' tools/*.html index.html | sed 's/class="//;s/"//' | tr ' ' '\n' | sort -u); do grep -q "\.$c\b" assets/css/style.css || echo ".$c"; done`
- **`tools/credit-roadmap.html` is not a tool.** 1,816 bytes, zero controls, zero
  script — a short article sitting in `/tools/`. The audit scores it `NO-CONTROL`,
  correctly. Either move it to `/guides/` or give it controls; as filed it fails its
  section's promise.

**And one that is NOT a defect, so you can ignore an adverse verdict if you see
it:** `tools/damage-checker.html` **works**. It scored `DEAD` because `toolaudit.py`
had no checkbox branch in its control picker, and that page has four checkboxes and
no buttons — so the harness drove nothing and reported that nothing changed. Ticking
`#dmg-1` takes `#damage-verdict` from `display:none` to `block`. Harness fixed
(`288e6e2be`); faults 11 and 12 are written up in the `webdesign_tools_repair`
NOTES. **If you audited your tools before 2026-07-31, re-run it** — the harness was
also blind to any page without a `<main>` element, which is all of yours.

I have not touched this site or this directory beyond appending here.

## 2026-07-31 — the acceptance baseline, and three premises of this lane corrected

Opening session of the decomposition build (`fidelity=high`). Before writing any
code I re-verified the handoff and then measured the site itself. The measurements
moved three things the handoff asserted.

### The handoff re-verified (all held)

- **27/27 URLs serve 200** (sitemap loop, §2 of the handoff).
- **27 pages, all `owned`/`deployed`**; zero `site_components`; one `ported-page`
  component per page.
- **M1/M2 code is in the running pod**, grepped on `adoption-locked/1` (1) and
  `page_id absent, resolved from` (1), with a **positive control**
  (`rerender_single_page` = 2) and a **negative control** (`xyzzy_not_a_symbol` = 0)
  in the same exec. `IMAGE_TAG` is now `v1.0.1214` — another session bumped it past
  the `v1.0.1213` the handoff names; my strings are still present.

### [CORRECTED 2026-07-31] This site has ELEVEN calculators, not twelve

Every prior doc in this lane says "12 inline-JS calculators". It is 11.
`tools/credit-roadmap.html` is a **static prose page that lives under `/tools/`** —
1,816 bytes, **zero** `<input>`/`<button>`/`<select>`/`onclick`/`addEventListener`,
and its only `<script>` is the shared `nav.js`.

Two independent methods agree, which is why I am confident enough to correct the
premise rather than flag it:
- **static**: the greps above, over all 28 files;
- **runtime**: the real-browser audit scores it `NO-CONTROL — nothing a visitor can
  touch`, while the other 11 score `RESPONDS`.

Why it matters: the acceptance bar is *"every calculator still computes"*. Measured
over 12, it can never pass — one of them cannot compute and never could. Measured
over 11 it is a real gate. `credit-roadmap` should be decomposed as **ordinary
editable prose**, which is what it already is.

### The acceptance baseline — captured, because it expires

`acceptance/BASELINE_2026-07-31_calculators.json`: **11 RESPONDS + 1 NO-CONTROL.**

This had to be taken *now*: the bar says every calculator *still* computes, and
"still" is unmeasurable without a before. Once decomposition starts, the pre-state
is gone.

Harness: `webdesign_tools_repair/toolaudit.py`, **sha256 `e7607680…`** — which is the
**working-tree copy, NOT HEAD** (HEAD is `1ea6740b…` at `f38f5bf7f`). I pinned a copy
into my scratchpad and ran from that, and saved the delta as
`acceptance/harness_wip_vs_f38f5bf7f.diff`, because:

- another session is **editing that file while I read it** (33 uncommitted lines) and
  was running it against my tools at 09:50, and
- **its uncommitted fixes were written against MY site's tools.** The diff's own
  comments name them: `damage-checker` scored DEAD because its only controls are
  checkboxes (setting `.value` on a checkbox is a no-op — a tick is a `click()`), and
  `credit-health-check` is a wizard that responds by moving a class so a different
  `<div id="step-N">` becomes visible, which `innerHTML` diffing cannot see.

So **HEAD's harness would have scored two of my working calculators DEAD.** A
baseline is only comparable against the same harness — re-pin from the diff (or the
commit, once that lane lands it) before re-running. Ports are randomised
(`--remote-debugging-port=9754`, `--user-data-dir=/tmp/toolaudit-<rand>`), so
concurrent audits do not collide.

> **Misstep, 30 seconds:** the pinned copy died on `ModuleNotFoundError: toolprobe`.
> The harness imports a sibling module from its own directory, so pinning it means
> pinning **both** files. `toolprobe.py` is clean at HEAD (`d36f020f…`).

### [CORRECTED 2026-07-31] The generic-flip blocker is BIGGER than "nested `<html>`"

PLAN §"The blocker" says feeding a whole document to `assemblePage` yields nested
`<html>`. True, but incomplete, and the missing half is worse.

```sql
SELECT slot_name, length(rendered_html) FROM site_components
WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26';   -- (0 rows)
```

**Zero `site_components`.** `assemblePage` resolves `head`/`header`/`footer` from
that table (`rerender_single_page_action.go:660`). Verbatim adoption never created
any, because it skips assembly entirely. So flipping `rebuild_policy='generic'`
today would ship every page **nested AND with no head, no nav and no footer** —
`buildDefaultHead` would substitute a generic head and the nav would simply be gone.
Decomposition must therefore **create the site-level chrome**, not only split the
bodies. That is a whole extra output of step 2, previously unstated.

### What the pages actually look like (this is a clean decomposition case)

Anatomy, uniform across 27 pages: `<head>` · `<div id="nav-placeholder"></div>` ·
body content · optional inline `<script>` · `<script src="/assets/js/nav.js">`.

- **`nav-placeholder` + `nav.js` on 27/27** — chrome is uniform, so ONE site-level
  `header` works.
- **`<footer>`: 0 of 28 files. `<main>`: 0 of 28.** There is no footer to extract;
  assembly will *add* both. That is new content on every page — acceptable under
  "evolve like the others", but it is a fidelity change and belongs in the owner's
  read-out, not buried here.
- **Head is effectively uniform**: by element inventory, 23 pages are
  `charset+viewport+stylesheet`, 4 are `charset+stylesheet`, 1 is the redirect stub.
  The 11 distinct title-normalised hashes are whitespace/title noise. `assemblePage`
  already injects per-page `<title>` and meta description from the `pages` row, so a
  single site-level `head` is sufficient.
- **A live defect that going generic FIXES**: `legal.html`,
  `guides/car-finance-explained.html`, `guides/secured-vs-unsecured.html`,
  `guides/uk-lending-landscape.html` have **no `<meta name="viewport">`** and so
  render at desktop width on a phone today. A shared head repairs all four.
- The tool itself is a contiguous `.card` (input-grid + results-box) plus a trailing
  inline `<script>` — genuinely separable from the sibling prose `<section>`.

### ⚠ `application-tracker.html` has TWO inline scripts, and uses localStorage

`<script>` at 75–128 and again at 135–169, plus `nav.js`; 9 `localStorage`
references. **An extractor that assumes one inline script per tool silently drops the
second and the tool half-works** — precisely the failure class this lane exists to
prevent. Preserve **all** inline scripts **and their order**.

### LANDMINE for the fidelity gate: `length()` is CHARACTERS, `octet_length()` is BYTES

The stored `standard-calc` component reports `length()` = **5,730** while the file is
**5,734 bytes**. That looks like a 4-byte fidelity failure and is not: the page
contains **four `£` signs**, each 2 bytes in UTF-8. `octet_length()` = 5,734 and
`md5(rendered_html)` = `14643b1f76ba4ee333d39a2ecfdf4352` = `md5sum` of the file.

So the handoff's "27/27 byte-exact" **holds, independently re-verified**. But the
gate the render_guardian seat asked for is about to be written into the pipeline, and
a gate built on `length()` would report a mismatch on a perfectly faithful page —
and, worse, could offset a real difference against a `£`. **Compare `md5`/`sha256`,
or `octet_length`. Never `length`.**

### Machinery that already exists — do not build it (step 4)

Both `site_components` and `page_components` carry `locked_at`, `locked_by`,
`lock_type` (CHECK: `permanent`|`timed`|`review`) and `lock_expires_at`, and
`site_components` has a partial index `idx_site_components_timed_lock` on
`lock_expires_at WHERE lock_type='timed'`. The **timed adoption lock is a config
choice, not a build.** `site_components` is also `UNIQUE (site_id, slot_name)`, so
chrome upserts cleanly.

The existing `ported-page` rows point at a **fleet-shared** `content_components` row
named *"Ported Page (webdesign.co.uk)"*, `component_level='section'`. Shared across
sites — the loanandmortgage lane independently recorded "reuse it, never seed a
second".

### A live adjacent lane, and it picked a DIFFERENT design

`docs024_key_docs_latest/loanandmortgagecalculator_couk/` — **untracked, created
today 09:19–09:42, active while I worked.** `git log` on it is empty; it is invisible
to every check that reads commits. Found only by listing the parent directory.

It builds `loanandmortgagecalculator.co.uk` (12 mortgage + 12 loan calculators) and
**copies my site's files**. It explicitly scopes out *"the loancalculator lane's own
adoption"* and states *"another lane owns this site — I copy its files, I do not
touch its directory"*, so there is **no collision**. It also independently confirmed
G10 and concluded "the deploy repo is the byte source, not the crawl", which is this
lane's step 1.

**But its Phase E chose a per-page split**: calculators → `owned` + verbatim with a
sha256 gate; guides → framework-managed. That is *not* what this lane is building,
and the difference is a genuine design question for the owner — see
`README_where_we_are`. Their split is far cheaper and freezes the tools for ever;
ours decomposes so a tool page's prose can evolve around a preserved widget. The
owner's word for this site was **"completely editable"**, which their split does not
deliver for 11 of 27 pages.

### Step 1 (file source) is OWNER-BLOCKED — and step 2 never needed it

The handoff's step 1 reads: *"Read our own bytes from the deploy repo (`gqls/sites`,
dir `loancalculator.co.uk/`) via **git-adapter**, so it is platform-side and reusable."*
Three things about that turned out to be wrong or blocked, in increasing order of
importance.

**1. git-adapter cannot read. It is write-only.** Its entire action surface is
`commit`, `create_repo`, `delete_repo`, `create_branch`, `create_pull_request`
(`internal/adapters/git/adapter.go:357`). There is no fetch/read/list operation, so
"via git-adapter" is a build, not a wiring job — and by the 2026-07-12 owner ruling
quoted in the code, the **write** credential lives in git-adapter deliberately while
reads go through a read token elsewhere. Adding a read op to the *write* adapter would
push against that separation.

**2. The read machinery already exists elsewhere — reuse it, don't add to git-adapter.**
`platform/orchestration/actions/diagnose_read_repo_files_action.go` already does
exactly this: GitHub **Contents API** with the `application/vnd.github.raw` media type
(no git binary in the chassis image, no base64 size quirks), authenticated by
`GITHUB_READ_TOKEN`. I exercised that identical code path as a positive control and it
works — `makefile` from `gqls/agentchassis`, **HTTP 200, 101,280 bytes**. What it is
*not* is reusable as-is: its input contract is the fix-plan schema (`plan.Edits`,
`modify`/`add`), and it reads files a plan names, not a directory listing. A site
source needs list-a-directory + fetch-each, which is a new action reusing that
transport.

**3. THE BLOCKER: the read token cannot see `gqls/sites`. Measured.**

| request | result |
|---|---|
| `GET /repos/gqls/agentchassis` (control) | **200** |
| `GET /repos/gqls/agentchassis/contents/makefile` raw (the real code path) | **200**, 101,280 bytes |
| `GET /repos/gqls/sites` **authenticated** | **404** |
| `GET /repos/gqls/sites` unauthenticated | **404** |

The token is a **fine-grained** PAT (`github_pat_`, 93 chars) and it *is*
authenticated — `x-ratelimit-limit: 5000` is the authenticated ceiling, not the
anonymous 60. So the 404 is **repo scope**, not auth failure and not a wrong path.
`gqls/sites` is private and simply outside this token's selected repositories.

> **LANDMINE (filed): GitHub answers "you may not see this" with 404, never 403.** An
> action built on this token would report *"loancalculator.co.uk not found in the
> deploy repo"* and the next session would hunt the path, the ref, the site-name
> spelling and the case — none of which is the cause. The positive control is what
> distinguishes them: same token, same media type, different repo.

Also required even once the token can see it: **`site-adoption-agent` is not on the
`isRepoCloningAgent` gate** (`spawn_actions.go:3066` — the members are `diagnose-agent`,
`code-indexer`, `fix-implementer`, `feature-implementer`), and that gate is what
injects `GITHUB_READ_TOKEN` into the pod. So the platform-side route needs, in order:
an owner-minted token scope → the agent added to the gate → the new list+read action.

**Only the first is blocked, and it is genuinely the owner's** (no GitHub admin
credential on this machine — the same shape as the loanandmortgage lane's Cloudflare
item). Recorded for the owner; not attempted.

#### The re-sequencing this forces, and it is an improvement

**Step 2 (decomposition) does not depend on step 1 at all, and the handoff's ordering
implied it did.** Decomposition needs faithful *bytes*, and we already hold them **in
the database**: `page_components.rendered_html`, 27/27 confirmed byte-exact against
the served files again today (md5 both sides, `14643b1f…` for `standard-calc`). The
previous session put them there precisely because the crawl could not be trusted.

So the correct order is:
- **now, unblocked:** decompose from the stored components — they are the verified
  faithful source for *this* site;
- **later, owner-gated:** the platform-side file-source action, whose value is **the
  next** site, not this one. It stops being a prerequisite and becomes a generalisation.

That also shrinks what has to be right on the first attempt: the decomposer reads from
the same table it writes to, inside one transaction, with no network and no new
credential.

## 2026-07-31 (later) — the decomposition rule, proved offline; and a rule that passed every safety test while failing the brief

Built `decompose_prover.py` in this directory: it runs the proposed splitting rule
over all 27 real pages, writes **nothing** to the DB and touches no site file, and
asserts the properties that would otherwise fail silently. The shipped decomposer
will be **Go** (platform convention); this exists so the rule is measured before it
is written, and so the council submission carries evidence instead of an argument.

Source of truth: the 27 stored components, exported from `page_components.
rendered_html`. **Re-verified byte-exact a third way** — DB → base64 → file, md5
against the repo files: **27/27**.

> **Misstep, and the check that caught it.** The first export produced 27 pages of
> **exactly 57 bytes each**. `encode(...,'base64')` **wraps output at 76 characters**,
> so each row spans many lines and my per-line parser read only the first. It
> "succeeded" — 27 files written, no error. What caught it was comparing each decoded
> page's md5 against the repo file, which is the only reason I did not go on to
> decompose 27 fragments and conclude the pages were empty. Fix:
> `translate(encode(...), E'\n', '')`.

### The rule, and why each clause exists

- page-local `<style>` and every inline `<script>` are captured **byte-exact**, by
  slicing the source at parser-reported offsets rather than re-serialising parser
  events (re-serialising normalises entities and attribute quoting — that is how a
  "faithful" port stops being faithful);
- a node is **tool-marked** if it is a form control or carries an `id` an inline
  script addresses;
- for each top-level body block, **descend** while every marked descendant sits
  inside one child; where descent stops, the **marked children** become the tool and
  everything passed on the way down becomes editable prose, in document order.

**Why the `id` rule rather than "the block with the inputs".** A calculator's output
region usually has no controls in it — `standard-calc` writes its answer into
`<div id="monthly-display">`. Keyed on controls alone, the results box is prose, and
the writer loop may one day rewrite away the very element `getElementById` targets.
The script's own id references are the only honest definition of what the tool needs.

### [FAILED, then fixed] v1 passed every safety proof and delivered nothing

v1 classified whole top-level blocks. It passed P1–P4 (bytes preserved, no block
lost, no orphaned script target) on all 27 pages — and put **98% of the visible text
on interactive pages inside the frozen tool component; 100% on eleven of twelve
pages.** Zero editable prose. The site would have been just as unable to evolve as it
is today, with a clean test run to say otherwise.

The cause is structural, not a coding slip: on this site the calculator and its
article are **siblings inside one `.container` wrapper**, so any whole-block rule
freezes both.

**The lesson is about the proofs, not the rule.** P1–P4 all measure *"nothing
broke"*. Not one of them measures *"anything became editable"*, which is the entire
brief. So I added **P5: the share of interactive-page text left editable**, and it is
the proof that failed. Safety proofs and purpose proofs are different axes, and a
green suite on the first says nothing about the second.

| rule | frozen text on interactive pages |
|---|---|
| v1 — whole top-level blocks | **98%** (100% on 11 of 12 pages) |
| v2 — descend while one child holds the markers | 49% |
| v3 — + take marked CHILDREN, not the parent; capture loose text | **25%** |

v2 still swallowed the `<h1>` and intro paragraph on 5 pages, because a tool's
machinery legitimately spans **two** siblings — `car-finance-calculator` has an input
`.card` and a separate `.comparison-grid` results region, so descent stopped at their
parent and took the prose with it. v3 stops there but takes only the marked children.
`standard-calc` went 90% frozen → **5%**; every interactive page now has editable
prose.

> **Misstep — my own P6 was measuring the wrong thing.** P6 ("no visible text lost")
> failed on exactly 5 pages, each by 45–60 characters, which looked precisely like
> real content loss and sent me hunting dropped text nodes. It was comparing
> body-derived output against **whole-document** text, so the shortfall was each
> page's `<title>` — which is not lost at all, it is carried as page metadata and
> re-injected by `assemblePage`. Corrected to body-vs-body and tightened to 98%.
> Same family as the fleet's `check-answers-the-question-you-encoded`: the figure was
> right, the question was not.

**A real bug that P6 did catch, before I broke it:** descent iterated element
children only, so **loose text between two elements was silently dropped**. Now
captured as prose via span arithmetic.

### Current result — all proofs pass

`P1` script bytes · `P2` style bytes · `P4` no orphaned script target · `P5` 25%
frozen · `P6` no body text lost, at a 98% threshold. Reproduce:
`python3 decompose_prover.py <dir-of-stored-pages>`.

### [CORRECTED 2026-07-31, correcting MY OWN correction earlier today] the homepage is a calculator too

This morning I corrected "12 calculators" to 11. That is right about `/tools/` and
**incomplete about the site**: `index.html` carries a **working hero calculator** —
`<input id="amount|interest|years">`, wired by an inline script, with the arithmetic
in **external** `/assets/js/global.js`. The audit scores it `RESPONDS`.

So the honest statement is: **12 interactive pages — 11 under `/tools/` plus the
homepage — and `credit-roadmap` is not one of them.** The original figure of 12 was
the right *number* by coincidence and the wrong *membership*, which is worse than
being wrong, because it agrees with the corrected count while pointing at a different
page.

**My acceptance baseline had the same hole**: it audited `/tools/*` only, so a
regression to the homepage calculator would not have been noticed. Baseline re-run
and extended — **12 RESPONDS + 1 NO-CONTROL across 13 pages**
(`acceptance/BASELINE_2026-07-31_calculators.json`).

Consequence for the decomposer: the homepage tool depends on an **external** script,
so preserving inline `<script>` is not sufficient — the tool component must also keep
its `<script src>` dependencies. The prover records these separately
(`external_scripts`) and the Go action must carry them.

### Corrections to the neighbouring lane's contribution (their §3, above)

They audited my tools before copying them and reported three live defects. Checked
each at the live bytes; two of the three do not hold.

- **`credit-health-check` is NOT broken on the live site. REFUTED.** Their evidence
  was that `assets/css/style.css` defines neither `.check-step` nor `.active`. True —
  and the page defines both **itself**, in a page-local `<style>` block in its
  `<head>`, along with `.score-meter` and `#meter-fill`. Fetched the live URL with a
  cache-buster: both rules are present in the served bytes. Their check looked only
  at the external stylesheet, so it could not see them.
- **"36 classes undefined in that stylesheet" is really 19.** Reproducing their
  method gives 44; counting page-local `<style>` too gives **19**. **25 classes are
  rescued by inline `<style>`** — including 7 of the 9 they named as evidence
  (`.check-step`, `.active`, `.score-meter`, `.comparison-grid`, `.stat-value`,
  `.progress-bar`, `.verdict-text`, `.type-btn`, `.debt-row`; only
  `.fca-style-warning` and `.market-context-box` are genuinely undefined). The
  underlying point survives at a quarter of the size: **19 classes do render
  unstyled**, and `.fca-style-warning` — the FCA compliance banner — is one of them.
- **`credit-roadmap` is not a tool: CONFIRMED**, independently of my own two checks.

Their remediation is unaffected: because their method **over**-reported, the classes
they restyled into their unified sheet are a superset of the real gap, and their copy
does define `.check-step`/`.active`/`.score-meter`. They dropped the inline `<style>`
blocks (0 `<style>` tags in their port) and covered the rules in CSS instead — safe,
by luck rather than by design. Worth telling them, because the same blind method will
under-count the next port.

**And the reason this matters far more to us than to them:** page-local `<style>` is
**load-bearing on 8 pages, 7 of them calculators**. A decomposer that preserves inline
`<script>` but drops inline `<style>` produces calculators that **compute perfectly
and display wrongly** — `credit-health-check` would show all five wizard steps at
once. That is exactly the state they mistakenly reported as live, and it is precisely
what my v1 would have shipped had P2 not been in the prover from the start.

### Design consequence, stated now rather than discovered later

Descent **dissolves the per-page `.container` wrapper**: its children become sibling
sections and layout must come from the chrome that assembly wraps them in. Correct for
a site moving to assembled mode, but it is a real visual change and **not** a fidelity
claim. It belongs in the owner read-out, and it is a second reason the site-level
`head`/`header`/`footer` components have to exist before any flip.

## 2026-07-31 (evening) — owner asks why only the ARTICLES are editable; and `RESPONDS` turns out to be vacuous

Owner: *"if necessary please rewrite the tools so we are able to build on them in the
future. Why are the articles editable?"*

### The answer to the question, recorded because it is a design correction

Three parts, in order of how much they matter.

1. **Nothing on this site is editable today.** All 27 pages are `owned` + `verbatim`.
   The "25% frozen / 75% editable" figure describes what the prover's split *would*
   produce, not the live site. My previous read-out could be read as describing a
   state that exists. It does not.
2. **Editability is a property of the ROW, not the markup.** A block is editable when
   it is a `page_components` row on a `generic` page, unlocked, and listed in
   `pages.sections` so the planner and writer loops can target it. No property of the
   HTML confers it.
3. **The split was not a judgement — it was a residue.** The only constraint I was
   given was "starts similarly enough with working tools", and byte-preservation was
   the only way I could *guarantee* a calculator survived. So I froze everything the
   script touches and freed the rest. **The articles are editable because they were
   what was left over.** That is a weak basis for a permanent architectural boundary,
   and the frozen 25% is precisely the part that can never improve: it cannot be
   restyled, cannot be made responsive, cannot be fixed when its CSS classes are
   missing (19 undefined today, incl. `.fca-style-warning`), is not a
   `content_components` tool row so no other page or site can reuse it, and is
   invisible to the tool-improvement loop.

So: **rewriting the tools is necessary**, not optional, if the goal is to build on
them. There are already **36 active `component_level='tool'` components fleet-wide**
(35 with an `html_template`) — that is the machinery these calculators should be
joining, not sitting beside as frozen blobs.

### The blocker that has to be closed FIRST: nothing verifies that a tool COMPUTES

Rewriting reimplements arithmetic about people's money. I went looking for what would
catch an error and the answer is nothing:

- **`check_tool_acceptance.go` (Tier 2)** validates that *selectors exist* — the
  "anchor rule" — and its own header says static checks *"CONFIRM, never refute"*.
- **`toolaudit.py`** reports `RESPONDS`: driving a control changed something.
- **No numeric or golden-value tool test exists anywhere in `platform/`.**

### [CORRECTION to my own baseline, PROVEN BY CONSTRUCTION] `RESPONDS` cannot tell a working calculator from a dead one

`toolaudit.py`'s reactivity test is `changed = snap() !== before`, and `snap()`
includes `'##' + all.map(e => e.value ?? '').join('|')` — **the value of every
control, including the one the driver just assigned.** So for any tool driven by
typing, `changed` is true *by construction*, whatever the page does or fails to do.

I did not want to assert that from reading the code, so I built the decisive case: a
page with **one number input, no script, no listener, nothing that can possibly
compute**, served locally.

```
RESPONDS    http://127.0.0.1:8791/inert.html
```

**A page with no logic at all scores `RESPONDS`.**

Being fair to the harness: `RESPONDS` also requires no console errors, no failed
subresources, every reference resolving and a non-empty main region — all real signal,
and my inert page passes those legitimately too. The vacuity is confined to the
*reactivity* component, and only for text/number/range/select-driven tools — which is
**11 of my 12**. For checkbox/radio tools it is not vacuous (a click does not change
`.value`), which is exactly why `damage-checker` scored DEAD until the `vis`
fingerprint was added.

**So my own headline this morning — "12 RESPONDS, 12 calculators working" — overstated
what was measured.** What the baseline actually establishes is that 12 pages are
structurally healthy and reactive-by-that-definition. It does **not** establish that
any of them computes correctly, and it would not have noticed a rewrite that returned
wrong money. Corrected here rather than quietly; the baseline file stays, because
verdict-to-verdict comparison is still a real regression signal.

### What I built instead: `toolgolden.py` — capture the ANSWERS

Drives each tool in a real browser with deterministic input vectors and records a full
behavioural fingerprint (every id-bearing element's text + `display`, every control's
value), so any rewrite can be required to reproduce it.

- **Vectors are derived from each page's OWN default values** (×1, ×2, ×0.5), clamped
  to each field's declared `min`/`max`/`step`. A positionally-fixed vector would put a
  rate of 12345 into field 2 and drive the arithmetic into `NaN`, where every
  implementation agrees and no real difference can surface. Scaling a field's own
  default keeps every value inside its intended domain for any tool, with no per-tool
  configuration.
- **The fingerprint is every id-bearing element**, not "the output field". Guessing
  which element holds the answer is what made `toolaudit` blind twice; these scripts
  address their regions by id because that is how they are written.

Verified against arithmetic I checked by hand — £10,000 at 7.9% APR over 5 years:

| vector | monthly | total interest | total repayable |
|---|---|---|---|
| defaults (10000 / 7.9 / 5) | **£202.29** | £2,137.40 | £12,137.40 |
| double (20000 / 15.8 / 10) | £332.54 | £19,904.80 | £39,904.80 |
| half (5000 / 3.95 / 2.5) | £175.31 | £259.30 | £5,259.30 |

The annuity formula gives 202.27 for the defaults (r=0.0065833, x=(1+r)^60≈1.48252),
and 202.29 × 60 = 12,137.40 exactly. **The harness is capturing real, correct
arithmetic, not just "something changed".**

> **⚠ A bug in this harness that would have been catastrophic, and the guard now
> written in.** My first version polled `document.readyState == 'complete'` and
> captured immediately. On `standard-calc` that returned a **mid-parse DOM**: the
> `.container` existed, so all three inputs were present and drivable, but the inline
> `<script>` at line 77 had not yet been parsed — so `calculateLoan` was undefined and
> **every vector recorded £0.00 while the capture reported success.** I spent a while
> believing the live calculator was broken; it is not (2 scripts in DOM, zero console
> errors, £202.29 → £404.57 on typing).
>
> **A golden file recorded from that state says every answer is £0.00 — and would then
> certify a completely broken rewrite as byte-perfect.** The exact inversion of the
> tool's purpose. Two guards now: `settle()` requires the script count *and* the
> serialised DOM length to stop moving across consecutive reads, and `dom_shape`
> (`"2:13804"`) is recorded in the golden file so a mid-parse capture is **visible in
> every later diff** rather than inferred.
>
> And the structural guard, which is this morning's lesson applied prospectively:
> **the harness REFUSES to write a golden file if any tool's output is identical
> across every vector.** A file like that certifies nothing and would mark an inert
> rewrite as correct — the same shape as "conservation proofs cannot fail on a no-op".
> It also refuses on a partial capture, because missing pages read as "nothing to
> check".

> **Second misstep, same run:** the 12-page capture died on `application-tracker` with
> `timeout waiting for Runtime.evaluate`. A modal dialog **blocks the renderer**, so
> CDP simply times out with no indication of the cause — its remove button calls
> `confirm()`. `window.confirm/alert/prompt` are now stubbed after settle. That is a
> real behaviour change (a tool gated on `confirm()` proceeds as if accepted) and it is
> stated in the file, because it applies identically to every implementation, which is
> what a differential test needs.

### The golden capture, and the two proofs that make it usable as a gate

`acceptance/GOLDEN_2026-07-31_tool_values.json` — all **12** interactive pages.

```
                                    fields  react  vary
index.html                              9      3     3
tools/application-tracker.html         13      1     0
tools/car-finance-calculator.html      13      2     2
tools/compare-loans.html               17      7     7
tools/consolidation.html               12      7     5
tools/credit-health-check.html         12      2     0
tools/damage-checker.html               8      1     0
tools/interest-rate-stress-test.html    9      3     3
tools/loan-vs-savings.html             12      5     5
tools/overpayment-calculator.html      12      2     2
tools/settlement-calculator.html        7      2     2
tools/standard-calc.html                9      3     3
```

`react` = fields that change when the tool is driven (gate A). `vary` = fields that
change when the input *values* change (gate B). **All 12 react; 9 vary.** The three
with `vary=0` — `application-tracker`, `credit-health-check`, `damage-checker` — are
exactly the three with no numeric field to scale (text, buttons, checkboxes), so gate B
does not apply to them **by construction** and `scaled_numeric()` exempts them. They
are still fully fingerprinted.

Two iterations were needed to get here, and both were the gate doing its job:

1. **First run refused**, flagging 4 tools. Three reported `controls driven: NONE` —
   my driver handled `number`/`range` only, so a button wizard, a checkbox tool and a
   text tracker were never driven at all. Extended to drive text, checkbox, radio and
   select.
2. **`consolidation` reacted but did not vary**, because it starts with **no debt rows**
   — there is nothing to compute until one exists. Added a bounded `SETUP_JS` that
   presses an add/new/+ button first. It now reads `react=7 vary=5`.

> **The gate was also wrongly FRAMED at first, and this is the more interesting error.**
> My only test was "outputs differ between the `defaults` and `double` vectors". For a
> tool with no numeric inputs the two vectors are **identical by construction**, so that
> test can never pass — it would have condemned three working tools for ever, the same
> unpassable-gate shape as the "12 calculators" denominator this morning. Split into two
> gates: **A** — does it react to being driven at all (compares a `before` snapshot,
> excluding control values); **B** — does the output depend on the input values (numeric
> tools only).

#### Proof 1 — it is deterministic (round-trip)

Re-ran `--compare` against the unchanged live site, an independent second capture:
**all 12 tools reproduce their golden values exactly.** Without this the harness would
be unusable as a gate regardless of anything else.

#### Proof 2 — it has TEETH (negative control), which the round-trip cannot show

A gate that always says MATCHES is worthless, and a round-trip alone cannot distinguish
one. So: served a local copy, captured its golden, then made one subtle arithmetic
change — the monthly-rate divisor, `(APR/100)/12` → `/11`. The page still loads, still
responds, still displays plausible money.

**Caught in every vector, flagged `NUMBER`, with exact values:**

| vector | golden | broken |
|---|---|---|
| defaults | £202.29 | **£205.74** |
| double | £332.54 | **£350.60** |
| half | £175.31 | **£176.10** |

Exit code **1** on divergence (verified without a pipeline swallowing it), so it is
usable in automation.

**And this is precisely the error nothing else would catch:** `toolaudit.py` scores the
broken version `RESPONDS` (its selectors all exist and its inputs still accept typing),
and `check_tool_acceptance.go` passes it (every anchor is present). A £3.45/month error
— £207 over the life of the loan — was invisible to every gate the platform had.

## 2026-07-31 (late) — two tools rewritten as platform-native components, both PROVEN numerically identical

Owner authorised the rewrite. Method, and it is the same for all 12:

1. Extract the widget into an `html_template` with `{{.field}}` slots + an
   `input_schema` of `static`/`llm` fields with fallbacks — the fleet convention,
   read off `tool-mortgage-overpayment` rather than invented.
2. Render offline with **`render_tool.go`**, which uses **Go `text/template`** (the
   real engine) with `missingkey=error`, and refuses to write output containing
   `<no value>` or a residual `{{`. A Python `str.replace` prover would leave a
   mistyped slot unreplaced and look fine — and `<no value>` in a stored template is
   the TL-030 corruption class, so the mistake must be caught offline.
3. Splice the rendered component into a **copy of the real page**, so only the widget
   differs, and run `toolgolden.py --compare` against `GOLDEN_2026-07-31`.

| # | tool | fields | result |
|---|---|---|---|
| 1 | `tool-loan-repayment` (standard-calc) | 15 | **MATCHES** — all 3 vectors, 9 id-fields |
| 2 | `tool-credit-health-check` | 18 | **MATCHES** — `react=2`, `vary=0` (correct) |

Tool 2 was chosen deliberately as the **structurally hardest** case, to test whether
the method generalises past a three-input calculator: a five-step wizard, no numeric
inputs at all, inline `onclick` handlers needing a global, and **page-local `<style>`
whose two rules are the entire one-step-at-a-time behaviour**. It does generalise.

### What the rewrite actually buys, per tool

- **Self-contained namespaced CSS with theme variables + literal fallbacks.** This
  **fixes a live defect by construction**: `.fca-style-warning`, `.market-context-box`
  and `.status-badge` are used by these pages and defined **nowhere** (19 such
  classes), so the FCA risk warning renders as unstyled body text today. It also fixes
  `#meter-fill { background: var(--accent) }` — **no fallback**, so on any site not
  defining `--accent` the meter is invisible and the tool looks like it failed.
- **Regulatory copy is `source: 'static'` with DO-NOT-REWRITE guidance.** Freezing the
  whole tool protected the FCA warning for the wrong reason; this protects it for the
  right one — a content loop can rewrite the prose around it and cannot touch it. Same
  for the hedged verdict wording ("lenders *likely* see you as…"), which must never
  become a lending decision.
- **Dated factual claims parameterised.** `3.75% Base Rate` and `7.9% market average`
  were hardcoded in markup and script. Now one static field each, per site.
- **`en-GB`/`GBP` parameterised**, which is what lets another site reuse the tool at
  all. The numeric result is independent of both — and the differential test is what
  establishes that rather than my assurance.
- **One namespaced global instead of two bare ones.** The wizard put `nextStep`,
  `showResult` and a mutable `totalScore` straight onto `window`; two tools on one page
  would silently overwrite each other's scoring.

### A scope decision I did NOT take, and why

The wizard's **11 option labels and their 11 point values stay as template literals.**
Flattening them into 22 unrelated text fields would be unreadable and trivial to
desynchronise (a reworded option whose score no longer matches). The right structure is
a `questions` array — but **no tool template in the fleet uses `{{range}}`, and every
`input_schema` field type is `"text"`** (measured: 60 tool components). Introducing an
array is therefore an **addition to a shared vocabulary**, which by the owner ruling of
2026-07-29 belongs in its own review rather than smuggled inside a tool rewrite —
precisely the `bugs_closed/124` precedent. Recorded as this component's next step.

### ⚠ A limitation in MY OWN harness that must be fixed before this pattern spreads

`toolgolden.py`'s `PRESS_JS` clicks *"the first button carrying an `onclick`
attribute"*. So I **kept the inline `onclick` attributes** on the wizard's buttons,
even though delegated listeners would be cleaner, because a modernised tool with no
`onclick` attribute is **undrivable by the gate and would read as broken while working
perfectly.** That is the gate dictating the code's style, which is backwards. Fix the
selector (click any enabled button) before rewriting the remaining tools, and
re-baseline. Not worked around silently.

> **Missteps — the backtick trap fired TWICE more, and an identifier check three
> times.** (1) `git commit -m` with `` `display` `` in backticks: bash executed it and
> the committed message of `13e6e2e52` reads "does NOT set  on the fingerprint
> elements". Forward-only forbids the amend; logged in `WRONG_CALLS.md` as the first
> recorded case of an **already-documented** landmine firing. (2) The same trap inside
> an **unquoted heredoc** (`<<EOF` expands backticks too) — fixed with `<<'PYEOF'` and
> `os.environ`.
> (3) Three splice attempts failed on assertions keyed on the identifier `nextStep`,
> because **my own component's comment explains that it replaced `nextStep`** — so the
> word legitimately survives, and an identifier-shaped check cannot tell code from a
> comment about that code. The assertion that works targets the **binding**
> (`onclick="nextStep`) and the original's unique line (`let totalScore`). Ordering
> matters too: strip the original's parts *before* splicing the component, or the
> removal regex matches the replacement.
>
> Each of these cost a cycle and none was dangerous, but the shape is worth keeping:
> **a check written in terms of a name will match the name wherever it appears,
> including in the sentence explaining that the name is gone.**

### Fixing the harness limitation — and my first fix made it WORSE, caught by a golden-to-golden diff

I said the onclick-only press selector had to be fixed before rewriting ten more tools,
because it forced inline handlers into every component. Fixing it took two attempts, and
the failure mode of the first is the useful part.

**Attempt 1: fall back to "any enabled button".** Re-baselined
(`GOLDEN_2026-07-31b`) and diffed against the original golden field by field —
**1,653 fields compared, 31 drifted.**

- **`pressed` went `'none'` → `'mobile-menu-btn'` on 9 of 12 pages.** The first enabled
  button on these pages is the **nav hamburger**, so the "fix" silently moved the press
  off the tool entirely and onto site furniture. The gate was no longer testing the tool
  at all — and the *fingerprint barely moved*, so nothing but a golden-to-golden diff
  would ever have shown it.
- **`application-tracker` leaked state between vectors.** Its `before` snapshot for
  vectors 2 and 3 showed five checkboxes already `true`: it persists to `localStorage`,
  the browser profile is reused across the three navigations, and once my extended
  driver began ticking checkboxes, vector 1's answers were still there when vector 2
  started. **A contaminated baseline is perfectly self-consistent**, so the round-trip
  test could never catch it.

**Attempt 2**, both fixed: exclude buttons inside `nav`/`header`/`[id*=nav]`/`[id*=menu]`
so the fallback cannot leave the tool, and clear `localStorage`/`sessionStorage` **then
reload** before each vector (clearing *after* the scripts have read storage leaves the
restored values on screen).

`GOLDEN_2026-07-31c` vs the original: **1,653 fields compared, 21 drifted — all of them
`application-tracker`, and every one an improvement.** The original golden pressed
**"Clear All Progress"**, which *wiped the state the driver had just set*, so it recorded
the **post-wipe** state as the tool's behaviour. `c` presses nothing (all its buttons are
reset/download) and records the driven state: boxes ticked, notes filled,
`save-status: "Typing..."`. Everything else — the other 9 calculators, the homepage and
the wizard — drifted **zero fields**, and `standard-calc`'s `pressed` returned to
`'none'`, which is the check that the fix is targeted rather than broad.

> **The transferable point: a round-trip proves DETERMINISM, not CORRECTNESS of the
> baseline.** Golden A round-tripped perfectly — twelve for twelve — while containing a
> destructively-wrong record for one tool and a state leak affecting another. Both were
> invisible to any single run, and both fell out immediately from diffing two goldens
> taken with different harness versions. **When you change a measuring instrument,
> re-measure and diff the whole corpus; do not just check the new run passes.** This is
> the prospective form of the landmine I filed this morning about harness versions.

`GOLDEN_2026-07-31c` is now the live baseline. `GOLDEN_2026-07-31` is superseded and kept
as the record of what the earlier harness saw (per the same landmine: a verdict is only
meaningful with the harness that produced it). `b` was a strictly worse intermediate and
was not kept.

**With the harness fixed, tool 2 was rewritten the right way round**: no inline
`onclick`, **no globals at all**, one delegated listener on the container reading
`data-chc-next` / `data-chc-points`. Both rewrites re-verified against `c`:
**MATCHES, 2 of 2.**

---

## 2026-07-31 late — the check moved into the platform, and all 11 tools were rewritten

Owner asked for three things: make the numeric check part of the tools workflow,
rewrite all the remaining tools, write a summary. All three done. The summary is
`SUMMARY_2026-07-31_calculators_that_prove_their_own_arithmetic.md`.

### 1. `computed_values` — the check, in the platform

New Tier-4 criteria check type (`internal/adapters/browserrunner/run_checks_action.go`,
`runComputedValues`). Drives with the EXISTING fill/click/select step vocabulary,
then asserts the EXACT text of every named output. Whitespace is the only latitude.
Registered at tier 4 in `experienceCheckTiers`, with `expect_values` in
`experienceCheckTypeFields`; both lockstep tests pass unmodified. Commit `c11efb91c`,
register entry TL-038.

**Tier 4 ONLY, and that is the load-bearing decision.** A Tier-2 form confirming
anchors would be four lines, exactly as `interaction` has. Adding it would mint the
same vacuity one rung higher under a name that reads like arithmetic was checked.

`toolgolden.py --emit-criteria <dir>` turns a capture into fences. 9 of 12 emitted;
**the 3 refusals were all real findings**, including one against my own tool 2.

### 2. All 11 tools rewritten, each proven — and it is 11, not 12

`index.html` and `tools/standard-calc.html` are the **same calculator**, sharing
`calculateLoan()` from `assets/js/global.js`: identical steps and identical expected
values in all three vectors. That file also holds `getRateContext()` and
`DYNAMIC_MARKET_AVG` with **no callers anywhere on the site**.

New harness `rewrite/verify_rewrite.py`: cuts the original widget out of the REAL
page, splices the rendered component in, serves the whole site locally, compares
against the golden for the live URL. Proves a rewrite BEFORE it ships.

> **`~/projects/sites2/loancalculator.co.uk` IS NOT THE SITE.** Two local copies
> exist. `~/projects/sites/` matches the served bytes exactly; `sites2/` differs on
> every page checked. Checked by md5 against live before using either — and only
> because the habit is cheap. Rewriting from `sites2` would have produced eleven
> components faithful to the wrong original, all passing their own review.

### 3. What the gate caught that review would not have

| # | caught | why nothing else would have |
|---|---|---|
| 1 | A schema fallback containing `"58-day"` **in double quotes** rendered into a JS string literal → syntax error → **whole calculator dead, showing £0.00 for every input** | It still shipped a `<script>` block (tool_health passes), still matched every selector (Tier 2 passes), still rendered normally. `render_tool.go` now REFUSES any template interpolating a quote-bearing field inside a `<script>`. |
| 2 | `display:flex` **blockifies its children**; computed display is fingerprinted. The same house-style row broke **two tools in OPPOSITE directions** — damage-checker needed non-flex, application-tracker needed flex | Their original stylesheets disagreed. Unknowable from reading either page; the two look identical. |
| 3 | A **one-character whitespace** difference: the original built rows with an `innerHTML` literal whose indentation became text nodes, so two remove buttons read `✕ ✕` not `✕✕` | `createElement` produces no such node. Invisible to a human, and to every other check. |

**The rule that came out of #1 and is now applied throughout: copy goes in the
MARKUP, the script writes only the number.** Retrofitted to tool 2, whose band
verdicts were in JS literals — safe today because none happens to contain a quote,
which is exactly why it was worth moving.

### 4. Defects found, and the line I drew

**Fixed** (provably unobservable today, gate confirms): `localStorage.clear()`
wiping the **whole origin** rather than its own keys; two tools counting checkboxes
**page-wide**; three property-assignment handlers (`window.onload`, `cb.onchange`)
that silently replace another component's; a leaked object URL per download; a
restore that died silently on a bad file.

**NOT fixed** (changes output, each owed its own re-baseline): three decimal places
on money (`£448.024`, `toLocaleString` with no `maximumFractionDigits`); 0% APR
computes nothing on car finance; consolidation counts a rate-less debt toward the
balance but not its interest; loan-vs-savings signals its verdict by colour alone.

> **THE LINE IS OBSERVABILITY, NOT SEVERITY** — and I had to apply it against
> myself. I wrote the accessibility badge for the colour-only verdict, the gate
> failed it on text content, and I **reverted it**. Mixing an improvement into a
> port means neither can be reviewed on its merits, and it is the same rule that
> keeps the `£448.024` bug in place for now.

### 5. Council: REVISE, and it was right

Corr `1056cf11-7693-4fb6-a9fe-f67ee9f28bca`. 5 approve / 4 object, gated by
`debug_historian` at **high**, and **four independent seats raised the same thing**:

> A criteria check type the running binary does not know is **SKIPPED**, and an
> **all-skipped fence PASSES**.

So a fence installed before the browser-runner rolls reports green having asserted
nothing — *the exact false-green this check exists to eliminate, reproduced by the
fix for it*. My risk note said "no fences are installed yet", which is a fact about
today, not a guard. Nothing stopped another lane installing one.

Answered with `acceptance/criteria/INSTALL_GATE.sh`: pod-grep with a positive
control in the same exec, plus the still-owed first-run skip assertion.

> **MISSTEP, caught by the gate's own control on its first run.** It reported
> `computed_values: 0, control: 0` and refused to conclude "not deployed". Cause:
> **`browser-runner-adapter`'s image has no `strings` binary**, so CLAUDE.md's
> `strings /app/X | grep -c` recipe fails silently and returns 0 for everything.
> Without the control that reads as "your change did not ship" and costs a
> pointless rebuild and roll. `grep -a` needs nothing installed. Second seat
> objection worth recording: `tooling_provenance` wanted a NOTES entry for the next
> author touching these files — this section is it.

### Tally: 11 of 11, proven

`application-tracker` · `car-finance-pcp-hp` · `compare-loan-offers` ·
`consolidation-risk` · `credit-health-check` · `early-settlement` ·
`loan-repayment` (serves BOTH standard-calc and index.html) · `loan-vs-savings` ·
`overpayment-impact` · `rate-stress-test` · `return-damage-checker`.

Three carry `allow_new_keys` (consolidation, credit-health-check,
application-tracker) because their controls **had no ids at all**, so the emitter
could not name anything to drive and they had **no numeric coverage available**.
Renames are paired BY VALUE and every pre-existing key is still compared strictly,
so a control that genuinely lost its value still fails. Re-baseline owed on ship.

`damage-checker` legitimately gets no `computed_values` coverage — it has no numeric
output. Its contract is a visibility change, so its PLAN wants `interaction` +
`has_visible_area`. **Do not manufacture a number so a check type fits.**

---

## 2026-08-02 — full decomposition, proven offline (owner chose full, not the split)

Owner answered the open question: **full decomposition**. This entry is the
technical log of getting it to the point where it is provable but not yet shipped.

### The baseline, re-established rather than assumed

27 stored documents dumped from `page_components.rendered_html` and checked
byte-exact against `octet_length` + `md5`, then three of them checked against the
LIVE bytes — identical md5. So the DB is the source of truth for what is served,
and an offline proof against the stored bytes is a proof about production.

> **MISSTEP, caught in 30 seconds and worth the line:** the first check compared
> file size against `length()` and reported 20 of 27 pages as mismatched by 1–15
> bytes. `length()` counts CHARACTERS; every `£` is 2 bytes and 1 character, so
> the deltas were exactly the count of non-ASCII characters per page. Nothing was
> wrong with the dump. `octet_length` is the operator; this is now in the RUNBOOK
> and in `prepare_work.sh`.

`decompose_prover.py` re-run over the fresh dump: **ALL PROOFS PASS**, 25% of
interactive-page text frozen in tool components (the rule's own P5 threshold is
50%).

### The prover cannot write rows, so `decompose_pages.py`

The prover returns `prose` and `tool` as two lists, each internally ordered and
mutually unordered. Fine for proving nothing is lost; useless for `position`.
The new module re-walks the same tree — importing `TreeBuilder`, `marked`,
`any_marked`, `loose_text` rather than forking them — and emits ONE ordered list.
`decompose_prover.py` gained an `if __name__ == "__main__"` guard so it can be
imported; two copies of a splitting rule that must stay identical is the drift
class this repo reviews for.

### BUG FOUND IN THE RULE: P4 is blind to external scripts

`/index.html` decomposed with its **entire results box classified as editable
prose** — `#monthly-display`, `#total-interest`, `#total-cost`, the three
elements the calculator writes into.

Cause: the rule derives "what the tool needs" from the ids that scripts address,
and only ever read INLINE scripts. On that page the whole of `calculateLoan()`
lives in `/assets/js/global.js`. Its `getElementById` calls were invisible.

**P4 passed, and could not have failed.** It asks whether every id an inline
script addresses travels with the tool; index's inline script addresses only the
three inputs, and all three did travel. The proof was narrower than the risk —
the definition was right, the input was incomplete.

Fix: `decompose_pages.py` takes `--assets` and folds every `<script src>` the
page loads into the same id set. A referenced script that cannot be read is a
**hard failure**, not a warning, because carrying on with the ids it could see is
precisely the silent-success shape that produced the bug. Post-fix, index emits
one tool block carrying all six ids.

Added alongside: `stranded_script_targets` — no script target may land in a PROSE
block, because prose is what a writer agent is licensed to rewrite. Zero on all
27 after the fix.

### The chrome was already wrong, and "are there rows?" said it was fine

The RUNBOOK's generic-flip precondition was "`site_components` must be non-empty".
It became non-empty at `2026-08-01 08:02:07`, written by something outside this
lane, followed five seconds later by 27 `page_rerender` items — all `complete`,
site unchanged by a byte, because `loadVerbatimPageHTML` returns before assembly.

Chrome that has never been exercised, and wrong three ways: stylesheet
`styles.css` (**404**, the real one is `style.css`), header `<ul>` with **no
links** (all 25 live in `nav.js`), and favicon + `og:image` both **404**.
Assembling any page would have shipped it unstyled and unnavigable.

The correction is recorded in the RUNBOOK where the wrong precondition was
written. **The check is not "are there rows" but "does the chrome resolve"** —
fetch every asset it references and require 200. Presence is not function; same
shape as the fleet landmine that a roll does not prove a deploy.

Replacements authored in `chrome/`: head links the real stylesheet and keeps
assembly's two literal injection points (`<title></title>`, `content=""`) intact;
header is the nav **extracted programmatically from `nav.js`'s template literal**,
not retyped; footer is an addition, justified by `/legal.html` having zero inbound
links from anywhere.

> **MISSTEP — a literal `<div class="container">` inside a CSS comment.** My first
> head explained the layout rule by quoting the markup, which is inert to a
> browser and unbalances the 5-pair structural predicate that `load_components.py`
> and the birth-write guard both use. Identical in class to the `{{` -in-a-comment
> trap from 07-31. All three chrome files are now balance-checked before loading.

### The mirror, and why it is a hypothesis rather than an authority

`assemble_mirror.py` reproduces `assemblePage` + `getPageSections` +
`sectionHasVisibleContent` + `injectPageJSONLD` + `StripToolDocHeader` in Python,
because the Go assembler is unexported and needs a page id — there is no way to
ask it "what would you produce?" without writing the rows, and writing the rows
is the thing being tested.

**A second implementation agrees with the first for the reasons they are both
wrong**, so the mirror has a scheduled test: the FIRST PAGE SHIPPED gets its real
rendered output diffed against the mirror's prediction, and the other 26 do not
move until they match. Two details that would otherwise diverge and be blamed on
the decomposition: Go's `json.Marshal` escapes `<`, `>`, `&` to `<` etc., and
sorts map keys at EVERY level (so `description` lands between `@type` and
`isPartOf`, and `isPartOf` comes out `@type`/`name`/`url`).

### `verify_assembled.py` — six assertions, two of which I got wrong first

A/B/C/D/E/F as tabulated in the RUNBOOK. The two failures were mine:

> **MISSTEP 1 — the text check measured the whole document, and the prover had
> already written down why not to.** P6's comment says in terms that measuring
> the whole document made five pages fail by exactly the length of their
> `<title>`, which is not lost at all. I scoped mine to the document and got 27
> failures, every one starting with the page title. Now body-only, like P6.
>
> **MISSTEP 2, in the same function and worse — the check had silently become
> vacuous-but-failing.** It split collapsed text on `\s{2,}|\n`, which after
> collapsing matches nothing, so every page produced ONE run: the entire body.
> The assertion was "is the whole original body a contiguous substring of the
> assembled page?", false by construction the moment a header sits between two
> blocks. **Too strict is not the safe direction** — 27 spurious failures look
> exactly like a broken check, and the only way to clear them is to weaken until
> it passes, which is how a real loss gets waved through.

### Check F exists because a screenshot caught what six checks did not

`tool-loan-repayment` serves both `/tools/standard-calc.html` and `/index.html`.
Built from standard-calc, it carries that page's FCA risk warning and two
market-rate lines. On the homepage, which never had them, it **added all three**.

Everything passed, each for a good reason: the numeric fingerprint covers `[id]`
elements and none of the three has an id; the text check asks what was LOST.
Nothing asked what was ADDED. So F now checks the other direction, per row —
prose rows adding text is always a failure, tool rows are listed for sign-off.

> **AND MY FIRST FIX FOR IT WAS WRONG, AND MY OWN SCHEMA GUARD REFUSED IT.**
> "Reproduce the page as it is today" meant blanking all three, and
> `render_tool.go` rejected the render: `fca_warning` is `required: true`,
> `source: static`, with `llm_guidance` saying it is a regulatory warning that
> belongs alongside a consumer credit promotion and must not be dropped. The
> homepage IS such a promotion. What I had called faithfulness was removing a
> warning that ought to be there. **The warning stays and the homepage gains it**
> — a deliberate change, flagged to the owner. Only the two DATED claims (a 7.9%
> market average, a 3.75% base rate) are blanked, because duplicating a
> stale-prone figure onto a second page doubles the places it has to be corrected;
> both are `required: false` and their own guidance says so.

Two further defects in F, both found by reading its output:

> **HTML comments leaked into text extraction.** `re.split(r"<[^>]*>")` tears a
> comment in half at any `>` inside it, and these components quote markup in their
> comments — F reported a paragraph of a component's own commentary as content
> added to the tracker page.
>
> **Whitespace-collapsed containment reported a formatting difference as content
> loss AND content addition at once.** The original marks up part of a sentence
> (`…Personal Loans is <strong>7.9%</strong>.`), which collapses to `… is 7.9% .`;
> the component holds it as one node. Compare with whitespace REMOVED.

F also separates **MOVED from ADDED**: a string the original held in a `<script>`
literal and wrote in at runtime is the same words relocated, which is exactly what
the rewrites did on purpose. 10 such strings on 5 pages. Without the split the
report is 12 lines of noise; with it, 2 lines that both matter — and both are
deliberate (the homepage warning, and the tracker's distinct restore-failure
message that replaced a success alert firing on a parse failure).

### Result

```
static checks pass on all 27 page(s): no text lost, no prose text added,
  no section dropped, no orphaned script target, no dead internal link
all 12 calculator(s) reproduce their golden values in the ASSEMBLED page
```

Screenshots confirm it: the guide page is pixel-identical above the fold, plus
the new footer. Nothing is written to the database yet.

### Shipping, and where it stopped (2026-08-02, later)

Three writes made, in order, each verified before the next:

1. **Chrome replaced** (`load_chrome.py --apply`). Its own validation refused the
   first run — reporting the two assets as **403** thirty seconds after curl had
   returned 200 for both. Cause: Cloudflare answers `Python-urllib` with 403
   regardless of method. Logged in WRONG_CALLS and LANDMINES; the discrimination
   test is that `style.css` must be 200 and `styles.css` 404 from the same code
   path. After the fix: all checks pass, chrome written, and **all 27 pages
   confirmed still byte-identical to their stored bytes** — the inertness claim
   measured rather than asserted.
2. **`tool-loan-repayment` updated** (`update_component.py --apply`, old row kept
   in `content_components_bak_20260802_decomp`). `d2ea795b` → `66a8e45d`.
3. **`guide-hidden-loan-fees` decomposed** — verbatim row replaced by one
   `ported-prose` row, in one transaction, after backing up all 27 pages' rows to
   `page_components_bak_20260802_decomp`. Guide first, deliberately: no calculator,
   so a mistake costs prose rather than arithmetic.

All 27 pages then dry-ran clean through the loader: **63 rows, no refusals.**

> **BLOCKED, and it is a queue-position problem rather than a fault.** Writing the
> rows deploys nothing; a render has to run. The `page_rerender` item was filed
> and — after being promoted out of `detected`, which is a status nothing ever
> dispatches — sits `triaged` behind **325 older items** across 12 sites. The
> selector orders `created_at ASC, priority ASC`, so priority only breaks an exact
> tie and a same-day item is a next-day deploy: items completing at 10:37 today
> were created 19 hours earlier.
>
> `dispatch-queue-depth.sh` says the lane is CLEAR, and it is right. **A clear
> lane and a 19-hour wait are not contradictory**, and treating this as bug 029 /
> 030 would be a misdiagnosis — this is depth, not a stall.
>
> The documented bypass (`049b_deploy_single_page.sh`, a direct `page-rerender`
> orchestrate) needs permission to run a kcat pod in the `kafka` namespace, which
> this session does not have. Escalated to the owner rather than worked around.

> **MISSTEP — `Council-Submitted: pending-see-next-commit`, and the RUNBOOK
> already calls this out.** §9 of that file records the identical mistake on
> `e6a8bb63b` ("a placeholder, and a mistake"), and I repeated it on both of
> today's commits. Worse, it was pointless: every file in both commits is under
> `docs/`, which the council gate refuses client-side, so **no submission was
> owed at all** and the right number of trailers was zero. Forward-only forbids an
> amend, so it stands as a wrong trailer in the log; `098` will not resolve it,
> and it asserts nothing false, but it is noise in exactly the field that exists
> to carry signal. The rule I should have applied: the trailer is for
> `platform/`, `internal/`, `pkg/` — check the pathspec before typing one.

### THE MIRROR IS VALIDATED — served bytes are byte-identical to the prediction

```
predicted: 16649 bytes  md5=80ea73c95365fc146953753e196063f0
served   : 16649 bytes  md5=80ea73c95365fc146953753e196063f0
cmp → BYTE-IDENTICAL
```

`assemble_mirror.py` predicted, before a single row was written, the exact bytes
the Go assembler would later serve for `guide-hidden-loan-fees` — including the
JSON-LD block's `<` escaping and its alphabetical key order at both levels,
the newlines around `<main>`, the tool-doc strip, and the chrome. **So the
offline "27/27 static, 12/12 calculators" can now be read as a result rather than
as a statement about my model of the assembler.** That was the one thing the
offline suite structurally could not tell itself.

> **CORRECTED — my queue estimate was wrong, and wrong in the direction that
> caused an unnecessary escalation.** I measured that items completing at 10:37
> had been created 19 hours earlier, saw 325 items ahead of mine, and projected a
> next-day deploy. It completed at **13:37 — about three hours**, not nineteen.
>
> The error: I read a single observed age as a queue latency. Items completing at
> a given moment are the OLDEST in the queue by construction — their age is the
> depth of the tail, not the wait a new arrival faces, because sites drain in
> bursts and most of those 325 items belonged to a handful of sites the
> dispatcher clears in one visit each. **An observed completion age is an upper
> bound on the backlog, not an estimate of your own wait.** The RUNBOOK section
> keeps the mechanism (created_at ASC ordering, `detected` never dispatched,
> priority nearly dead) because those are all true and load-bearing; only the
> "19 hours ⇒ next-day" projection was wrong, and it is struck through there.

All 27 pages are now decomposed: **63 rows, 51 prose + 12 tool, 0 verbatim**,
with every page's original row preserved in `page_components_bak_20260802_decomp`
(27 pages covered) and `load_decomposition.py --restore <page>` as the one-command
way back. 26 rerenders filed and draining.

### Checked: the per-page copy split landed correctly

```
name                | fca_warning | market_claim | market_context blanked
index               |      t      |      f       |        t
tool-standard-calc  |      t      |      t       |        f
```

Visible text of the index tool row, tags/script/style/comments stripped:
`Warning: Late repayment … moneyhelper.org.uk. Amount to Borrow (£) Interest Rate
(APR %) Term (Years) Monthly Repayment £0.00 Total Interest: £0.00 Total
Repayable: £0.00` — the FCA warning and the calculator, and nothing else. The
`{{if or …}}` guard removed the market-context block entirely rather than leaving
an empty styled panel; only its two CSS rules remain.

> **A `LIKE '%3.75%'` probe said the base-rate claim was still on the homepage,
> and it was not.** The only occurrence is inside the HTML comment I wrote in the
> template explaining why the guards exist. A substring test over `rendered_html`
> cannot tell markup from a comment — the check that answers the actual question
> is to strip `<script>`, `<style>` and `<!-- -->` first and read the visible
> text, which is what the block above does. Same family as the CSS-fallback
> landmine: present in the source, inoperative on the page.

**FOLLOW-UP, deliberately NOT done now:** that comment ships in the public page
source (HTML comments are not stripped — only the tool-doc sentinel block is), and
one clause of it editorialises rather than explains. `index` was `claimed` — mid
re-render — when I noticed, so rewriting the component then would have raced a
live deploy for a cosmetic gain. It rides with the next real touch of
`tool-loan-repayment`, alongside the queued defect fixes.

### End-to-end proof on the LIVE decomposed pages

`verify_shipped.py` over every page rendered so far: **10 of 10 byte-identical to
the mirror's prediction**, including `index` (5 rows, tool at position 3) and
`tool-application-tracker` (the structurally hardest tool). Not "close" —
`cmp` clean, same md5.

Then the calculators driven on the LIVE deployed pages with
`toolgolden.py --compare` against the golden captured from the hand-built site.
Raw output says "2 of 2 tools diverged", which is why it was classified rather
than eyeballed:

```
distinct diverging keys: 4
  clear-progress    APPEARED (absent in golden)
  download-backup   APPEARED (absent in golden)
  restore-backup    APPEARED (absent in golden)
  nav-placeholder   VANISHED (gone from live)

keys whose VALUE actually moved: NONE
```

Three APPEARED — the tracker's backup controls, which had no ids in the original
and therefore no way to be asserted on; that is the documented `allow_new_keys`
case and the re-baseline it owes. One VANISHED — the nav placeholder, because the
nav is server-rendered now. **Nothing computed a different number.**

> **A raw pass/fail from `--compare` would have read as a two-tool regression.**
> The tool answers "is the fingerprint identical", and the fingerprint legitimately
> gained and lost keys in this change. The question worth asking is narrower —
> "did any key that existed in both change its value" — and it has a different
> answer. Classify divergences by kind before believing a count of them.

### Six live calculators driven: 41 diverging keys, ZERO changed values

`toolgolden.py --compare` against the golden, over the six tool pages live at the
time. Raw verdict: "6 of 6 tools diverged". Classified:

```
diverging keys: 41
  APPEARED (28) debt-1-name … debt-3-remove, add-debt      (consolidation)
                chc-2-1 … chc-5-3, chc-restart             (credit-health-check)
                clear-progress, download-backup, restore-backup (application-tracker)
  VANISHED (13) input#0 … input#11                          (the SAME controls)
                nav-placeholder                             (chrome, now server-side)

keys whose VALUE moved: NONE
```

**The APPEARED and VANISHED sets are the same controls seen from two sides.** The
originals addressed these inputs positionally — the fingerprint keyed them
`input#0 … input#11` because they had no `id` to name them by. The rewrites gave
them ids, so they are now named. Nothing gained or lost a value; a naming scheme
changed. That is precisely the `allow_new_keys` case recorded when the three tools
were rewritten, and `verify_assembled.py` handles it by pairing renames BY VALUE —
which is why the offline run reported "+30 new key(s)" and passed while the raw
tool reports six regressions.

**So the re-baseline that has been owed since the rewrite is now actionable, and
better than before.** The reason those three tools had no numeric acceptance
coverage was that their controls could not be named; they can be now. Capture a
fresh golden from the live decomposed site once all 27 have rendered — that golden,
unlike the current one, can drive every control on every tool.

> **A pass/fail from `--compare` is the wrong instrument for a change that alters
> the fingerprint's SHAPE.** It answers "is the fingerprint identical", correctly,
> and the honest question here is "did any key present in both change its value".
> Six red lines and 41 keys look like a catastrophe and are a rename. Classify
> divergences by kind — APPEARED / VANISHED / VALUE CHANGED — before believing a
> count of them, and never before reporting one.

### A `complete` work item is not a propagated deploy — one page, one false mismatch

At 18 pages complete, `verify_shipped.py` reported **1 of 18 did not match**, with
a 432-line diff whose leading hunk was the header banner comment MISSING from the
served page — i.e. the page was still serving chrome from before the 10:29 chrome
write. Re-running the identical command minutes later: **19 of 19 EXACT.**

The page had been fetched inside its deploy window. `page_rerender` flips the work
item to `complete` when the render and the git commit succeed; the bytes then have
to go through the sites-repo Action, `b2 sync`, and a Cloudflare purge, which is a
minute or two behind. So for that window the wire serves the PREVIOUS page while
the database, the work item and the git commit all say the new one shipped.

**Why it did not become a false regression report:** the verifier prints
`READ THE DIFF and decide which side is wrong before writing any more rows` rather
than a bare fail, and prints the diff. The first hunk was chrome that is identical
for all 27 pages — so "one page lost the header comment" was not a coherent
failure mode, and the obvious next move was to re-fetch rather than to investigate
the decomposition. A boolean would have sent me looking for a bug in the page.

**The check:** when ONE page in a batch mismatches and the others pass, re-run
before believing it. A real decomposition fault is per-page-shape and reproducible;
a propagation lag is transient and clears on the next fetch. Do not add a sleep to
the verifier — that hides the distinction rather than reporting it.

### The re-baseline: coverage up 27%, and nothing computed differently

`GOLDEN_2026-08-02_decomposed.json` captured from the live decomposed site, all 12
interactive pages, no INERT flags and no capture errors. Compared key-for-key
against `GOLDEN_2026-07-31c` (chrome ids excluded):

| page | id fields | positional controls | shared-key values |
|---|---|---|---|
| /tools/consolidation.html | 9 → **19** | **8 → 0** | ok |
| /tools/credit-health-check.html | 9 → **21** | 0 → 0 | ok |
| /tools/application-tracker.html | 10 → **13** | 0 → 0 | ok |
| the other nine | unchanged | 0 → 0 | ok |

```
id fields: 94 -> 119   shared keys compared: 94   VALUES THAT MOVED: 0
```

Two claims, both measured rather than argued:

1. **Coverage is up 27%** and the gain is concentrated in exactly the three tools
   that had no numeric acceptance coverage. Consolidation's eight POSITIONAL
   controls are the sharpest case: `input#0 … input#7` cannot be asserted on in
   any durable way, because inserting a field silently renumbers every key after
   it. All eight are named now.
2. **Every one of the 94 keys present in both goldens has an identical value.**
   The decomposition changed how the site is stored and assembled, and changed no
   number it computes.

`GOLDEN_2026-07-31c` is KEPT, not replaced. It is the only record of what the
HAND-BUILT site computed, and every equivalence claim made during the rewrite is
stated against it; the new file is the forward baseline. Same rule as the summary
series — the old one being superseded is not the same as it being wrong.

The re-baseline owed by the three `allow_new_keys` tools since 2026-07-31 is
therefore **discharged**, and discharged better than it would have been on 07-31:
captured after decomposition, it names controls that were unnameable before.

### COMPLETE — 27/27 live, and every check green

```
rerenders                 27/27 complete, 0 failed
HTTP 200                  27 of 27
byte-identical to mirror  27 of 27  (verify_shipped.py)
footer present            27 of 27
stale #nav-placeholder     0 of 27
server-rendered nav        27 of 27
calculators vs new golden  12 of 12 MATCH — "all 12 tools reproduce their golden values exactly"
```

The last line is the one that took the most work to be able to say honestly. It is
a comparison against `GOLDEN_2026-08-02_decomposed.json`, captured from the
decomposed site — so on its own it only proves the site is self-consistent. What
makes it meaningful is the chain behind it: that golden was diffed key-for-key
against `GOLDEN_2026-07-31c`, captured from the HAND-BUILT site, and **0 of the 94
shared keys moved**. Self-consistency plus a clean diff against the original is
equivalence; either alone is not.

Final state: 27 pages, 63 rows (51 `ported-prose` + 12 tool), 0 verbatim, every
original preserved in `page_components_bak_20260802_decomp`, restore one command
per page.

### 2026-08-02 (later) — flipped to `rebuild_policy='generic'`, with the calculators locked

Owner asked for the flip. Done for all 27 pages, in one transaction with a
permanent lock on the 12 tool rows.

**What the flip actually removes**, read before doing it rather than after — two
named guard rails, both keyed on `rebuild_policy='owned'`:

1. `reconcile_site_plan_action` (guard rail 1 of the experience loop): an owned
   page that the plan says is missing or re-composed emits `owned_page_review`
   (needs_human_review, **no handler**) instead of `needs_page`. Without it the
   generic page builder rebuilds the page — documented as producing "a widget-less
   prose page where an interactive tool belongs (TP-004; the vonc arena clobber)".
2. `save_page_sections_action`: refuses a generic section save outright on an
   owned page, because that action DELETEs and re-inserts every `page_components`
   row — "exactly the TL-001 clobber".

**Measured before flipping, and the measurements changed the plan:**

- **`site_plans` for this site: 0 rows.** So guard rail 1 has nothing to guard —
  the reconciler iterates plan pages and there are none. That risk is LATENT, and
  becomes real the moment a plan is created (i.e. when the classifier is queued).
- **No scheduled tick walks sites.** Of 26 enabled `scheduled_tasks`, the only
  site-touching one is `build-pipeline-trigger`, which dispatches work items that
  already exist; it does not create them. So the flip, like every step before it,
  is **inert until something creates work for this site** — it is not a starting
  gun.
- **`save_page_sections`'s DELETE carries the lock predicate** —
  `DELETE FROM page_components WHERE page_id=$1 AND (locked_at IS NULL OR expired)`
  at line 566. **That is the fact the whole approach turns on.** A locked row
  survives a generic section save, and the blocked write emits a
  `lock_blocked_change` item for human review rather than disappearing.

So the tool rows are locked `permanent` (`locked_by =
'decompose_20260802_proven_calculators'`, matching the fleet convention of a
free-text reason — the 47 existing locks use e.g. `182_legal_pages`), and the
prose rows are deliberately left writable, because being editable is the whole
point of the decomposition.

```
rebuild_policy   generic  27 pages
component rows   section  51 rows   0 locked  51 writable
                 tool     12 rows  12 locked   0 writable
```

Proved the lock BITES rather than assuming it, using the exact predicate the
delete carries:

```
tool-standard-calc  prose-0..prose-3, prose-5  agent_may_delete = t
tool-standard-calc  tool-4                     agent_may_delete = f
```

Site byte-unchanged after the flip: **0 of 27** differ from the verified build. A
policy change is not a render.

**The cost of the lock, stated rather than discovered later:** the queued tool
defect fixes (overpayment's three-decimal money, car-finance at 0% APR,
consolidation's rate-less debt, loan-vs-savings' colour-only verdict) now need an
explicit unlock before the row can be rewritten. That is the intended behaviour,
not a snag — changing a calculator whose arithmetic is proven should be a
deliberate act, and the lock makes the attempt visible instead of silent.

### Post-roll: the new chassis renders identically, proved rather than inferred

A fresh chassis was deployed at `2026-08-02T21:39:20Z` (previous pods `18:39:14Z`).

Two checks, and the distinction between them is the point:

1. **All 27 served pages re-verified: byte-exact.** This proves the SITE did not
   change. It does **not** prove the new binary renders the same, because nothing
   had re-rendered — the wire was still holding output from the old image.
2. **So one page was re-rendered on the new image and checked.**
   `tool-standard-calc` (6 rows, tool at position 4 — the most structurally
   representative page) rendered at `21:53:41Z` and came back **EXACT** against a
   prediction written before the roll, with its calculator still MATCHING the
   golden.

The renderer source was also unchanged by this roll (`git log` since the previous
pods' start over `rerender_single_page_action.go`, `rerender_link_repair.go`,
`tool_doc_header.go` → empty), so EXACT was expected. **Expected is not measured**,
and the whole method here has been to prefer the measurement — the assembly path
is now proven stable across two chassis images.

The earlier roll counts too: the chassis rolled at `18:39:14Z` mid-rollout, 26 of
27 pages rendered after it and 1 before, all matching the same predictions.

---

## 2026-08-03 — the four queued calculator defects

Taken in the order the 08-02 handoff set. All four are *deliberate* behaviour
changes, which is why they were kept out of the port: equivalence was the
contract then, and a port that quietly printed £448.02 would have been changing
the page under cover of a rewrite.

| tool | defect | fix |
|---|---|---|
| `tool-overpayment-impact` | `£448.024` — `toLocaleString` with no `maximumFractionDigits`, whose default max is 3 | `maximumFractionDigits: 2` |
| `tool-car-finance-pcp-hp` | 0% APR computed nothing; the tool kept showing the last rate's figures | a linear branch at `r === 0`, the annuity formula's limit |
| `tool-consolidation-risk` | a debt with a balance but no rate counted toward the balance and contributed no interest | a blank rate is distinguished from a zero rate; an incomplete row WITHHOLDS the comparison |
| `tool-loan-vs-savings` | the verdict was carried by the `.winner` colour alone | a text badge in the winning panel, exactly one non-empty at a time |

### The finding worth carrying off this lane: HALF THESE FIXES ARE INVISIBLE TO THE GATE

`toolgolden.py` derives its vectors by scaling each numeric field's **own
default** (×1, ×2, ×0.5). That is a good default policy and it has a consequence
that is easy not to notice: **the gate only ever visits neighbourhoods of the
shipped defaults.** Two of these four defects live outside every such
neighbourhood —

- car finance breaks at APR **0**, and the APR default is 8.9. No scaling of 8.9
  is 0.
- consolidation breaks on a **blank** rate, and the driver fills every numeric
  field it can find, so every row it builds is complete.

Measured, not reasoned: `verify_rewrite.py` reports **MATCHES** for both tools
before *and* after the fix. A green gate said nothing whatever about either.

So `rewrite/defect_vectors.py` was written to drive the defect conditions
explicitly, with every expected value derived **by hand from the arithmetic** —
capturing them from the tool would only re-record the bug. It renders the
components from a pinned git sha as a negative control and scores each case on
whether it **DISCRIMINATES**, not on pass/fail.

Result: **4 PROVEN, 3 CONTROL, 0 vacuous.** And the equivalence gate answers the
complementary question — 9 of 11 tools MATCH; the two that diverge do so only on
the intended keys (`save-display`, and four `text/display` keys on
loan-vs-savings with `loan-benefit`/`save-benefit` **unmoved**, so no number
changed).

### MISSTEP 1 — my own arithmetic was wrong, and the harness caught it

The consolidation control case expected `£1,688.83` for £5,000 at 20% over 36
months. The tool said `£1,689.45`. **The tool was right.** I had rounded the
monthly payment to 185.80 before multiplying by 36; at full precision it is
185.817917 and the interest is 1689.445.

The lesson is not "be careful with arithmetic" — it is that an expectation
derived independently is worth writing precisely *because* it can disagree with
the code, and the disagreement has to be adjudicated rather than assumed in
either direction. Had I captured the expectation from the tool, this case would
have "passed" while asserting nothing.

### MISSTEP 2 — the negative control stopped being one the moment I committed

`defect_vectors.py` first read its baseline from `git show HEAD:`. That is
correct for exactly as long as the fixes are uncommitted. One commit later, HEAD
carried the fix, both sides of `--both` staged the *same* component, and every
case reported VACUOUS.

**It was only caught because the scoring had already been changed to compare
readings rather than pass/fail.** The earlier pass/fail version would have called
that same run **PROVEN** — a negative control that silently stops being one while
still printing a reassuring word. That is the exact failure this file exists to
argue against, reproduced by the file itself within an hour of writing it.

Fixed by pinning `PRE_FIX_REF = 6e8098022` (absolute sha, `--ref` to override).
**A baseline must name a commit that cannot move under it.**

### CORRECTION — the rate-less debt biased the verdict the OTHER way

The `tool-consolidation-risk` template carried this note:

> a half-filled row quietly makes consolidation look **better** than it is,
> because the new loan charges interest on a balance whose existing interest was
> never counted.

**That is backwards, and its own stated reason contradicts it in the same
sentence.** Omitting a debt's existing interest *understates* `oldTotalInterest`;
`newTotalInterest` still charges the new rate on the full balance; the verdict
tests `newTotalInterest > oldTotalInterest`. Understating the old side biases the
verdict toward **WORSE** — the tool talked the reader out of a consolidation that
might have been right for them.

Measured rather than argued. Driven at the pinned pre-fix sha, a £5,000 debt with
a blank rate produces:

```
old-int  £0.00
new-int  £1,359.36
verdict  "⚠️ Term Extension Risk: Consolidating will cost you £1,359.36 MORE in total interest."
```

What made this catchable was reading the reason clause against the conclusion,
not re-deriving the arithmetic: a sentence whose "because" does not support its
claim is worth stopping on, in your own writing and anyone else's.

### The design call on consolidation, stated because it was a call

Two candidates for a rate-less row:

- **exclude the row** from the balance too, restoring like-for-like arithmetic;
- **withhold the comparison** until the row is complete.

Withholding was chosen. Excluding silently answers a *different* question from
the one the reader asked — they would see a confident verdict computed over a
subset of their debts with nothing on the page saying so. Withholding cannot
state a wrong comparison at all, which is the property worth having on a page
about mis-selling risk. The balance is still shown (it is a fact, independent of
any rate); the three interest figures show `—`; the verdict box says what is
missing. **This is a judgement call, not a forced move** — if the owner prefers
exclusion it is a small change.

### THE TRAP THAT NEARLY SHIPPED: a new schema field renders EMPTY

Caught one step before it would have gone live, and it would have looked exactly
like success.

`content_components.input_schema` carries a `fallback` for every field and it is
natural to assume the renderer consults it. **It does not.** The render context in
`rerender_page_sections_action.go` is

```
base ⊕ page_components.content_data ⊕ plan.resolved_data
```

and `input_schema` is not in that list at all. `RenderTemplate` resolves an
unknown key to the **empty string** and logs a `Warn` — it does not fail, does not
fall back, and does not mark the page.

So adding a required field to a schema and a `{{.field}}` to a template is only
**two thirds** of a change. Without the backfill this would have shipped:

- `tool-loan-vs-savings` with an **empty accessibility badge** — the entire fix,
  rendering to nothing, on a page that would still have passed its acceptance
  check, because the element is present and no number moved;
- `tool-consolidation-risk` with an invisible notice in two empty inline colours.

`decompose/backfill_content_data.py` closes it: it adds schema fields the row does
not carry, using the schema's own fallback, and it merges `patch || stored` so it
can only ADD — a live value edited by the writer loop or a human must beat a
fallback, and the tool cannot tell an intentional edit from a stale one. Applied:
**4 fields across 2 rows**, exactly the 4 the schemas gained.

### Two things checked before firing `rerender_sections`, which had never run here

The decomposition shipped through the *assemble-only* branch (`render_page`, no
`reason`). These fixes need `reason=section_data_resolved`, which takes
`rerender_sections` and re-renders **every** section from `content_data` — a path
never exercised on this site.

1. **Is the prose re-render byte-stable?** The `ported-prose` template is
   `<section class="ported-prose" data-component="ported-prose">{{.content}}</section>`.
   All **51 of 51** prose rows reproduce their stored `rendered_html` exactly from
   `content_data.content` through it (checked in SQL, not inferred).
2. **Does the engine escape?** If `executeGoTemplate` used `html/template`,
   `{{.content}}` would be HTML-escaped and all 51 prose sections would turn into
   visible markup. It imports `text/template` (`call_agent.go:12`). Verified
   rather than assumed, because the failure would have been site-wide and
   irreversible in one pass.

### MISSTEP 3, and the biggest one — the documented re-render route does not work here

I wrote the "Changing a LOCKED calculator" RUNBOOK section, checked the two things
I thought could go wrong (prose byte-stability, `text/template` vs `html/template`),
unlocked the row, fired the re-render, and got back:

```
work item b0c2265d  status: complete
orchestration 439489b6  rerender_sections -> rerendered: 0, carried: 4, skipped: false
```

**Nothing re-rendered.** `rerender_sections` resolves each section's component by
passing `page_components.slot_name` to `loadComponentSchemas`, which matches
`content_components` by **name or function**. This site's slots are positional —
`prose-0`, `tool-2` — chosen deliberately (`assemble_mirror.py:270-273`) so a
dropped-section warning names which paragraph vanished. Nothing is called `tool-2`,
so every section took the `component not found, carrying stored HTML` branch.

The work item said `complete`. The page was byte-identical. **Every signal the
platform emits was consistent with a successful deploy.** Filed as `bugs_open/182`
with the fleet measurement (78 slots across 6 sites) and put through the `090`
diagnosis loop per the 2026-07-31 ruling.

**What I checked, and what I should have checked.** I verified the two ways the
re-render could DAMAGE the page and neither of the ways it could silently DO
NOTHING. Both of my checks were about the blast radius of a change that happened;
neither asked whether the change would happen at all. That asymmetry is the lesson:
*"what could this break"* and *"could this be a no-op"* are different questions, and
only the first one feels like diligence.

The tell was available before the deploy and I did not look for it — the whole
lane knew `slot_name` was positional, and `loadComponentSchemas` takes names.

### MISSTEP 4 — the fix for that found a defect in the mirror, by accident

`render_tool_row.py` renders the row twice: once from the working tree, once from a
baseline ref, so the second can act as a control. `render_component`'s cache key was
`(function, overrides)` and **omitted `rewrite_dir`** — harmless while every caller
passed one directory, wrong the moment two directories are in play. The second
render returned the first's output and the two compared EQUAL.

It failed for exactly the two tools whose fix touched only JavaScript (unchanged
overrides ⇒ colliding key) and worked for the two that had gained a schema field.
**It was wrong precisely where a wrong answer reads "identical — nothing to do,
your fix is already live".** Fixed; the key is now `(abspath(rewrite_dir), function,
overrides)`.

Also measured on the way past: `carryStoredSection` → `save_page_sections` is **not
byte-preserving** — it trims the trailing `\n` after `</script>` (8893 → 8892 on a
row whose content was untouched). The control names that tolerance explicitly
rather than comparing loosely.

### Shipped, and proved on the SERVED pages

The route that works here is the one all 27 pages were originally shipped through:
render offline with the same Go engine, write `rendered_html`, let the
**assemble-only** branch stitch it.

- `defect_vectors.py --live` drives the real production urls: **8 of 8 pass.**
- `toolgolden.py --compare` against the live site: **10 of 12 MATCH**; the two that
  diverge do so only on the intended keys — `save-display` (NUMBER) and four
  loan-vs-savings `text/display` keys, with `loan-benefit`/`save-benefit`
  **unmoved**, so no arithmetic moved anywhere on the site.
- Positive AND negative controls per served page: the string the fix ADDED is
  present; the string it REMOVED returns **0**.
- New golden `GOLDEN_2026-08-03_defects_fixed.json`, self-verifying (12/12).
- All 12 tool rows re-locked. The unlock window was ~20 minutes, nothing wrote to
  the row, and **in hindsight it was unnecessary**: the working route writes by SQL
  and assemble-only `render_page` never touches `page_components`.

---

## 2026-08-03 (later) — the owed tidy, and the harness that had quietly expired

### The tidy was not cosmetic after all

`tool-loan-repayment` carried a twelve-line HTML comment explaining why its three
copy blocks are conditional. HTML comments are NOT stripped from a served page —
only the tool-doc sentinel block is — so it shipped in the public source of both
`/index.html` and `/tools/standard-calc.html`.

It was queued as a cosmetic tidy ("one clause editorialises"). Reading it again
with fresh eyes, it is worse than that, and the reason is a small joke at our
expense:

> **The comment explaining why the homepage must not carry two dated factual
> claims was publishing both figures into the homepage's own source.**

Measured before the change — on the LIVE `/index.html`:

```
=== tool-doc ===          0     <- the sentinel block IS stripped
nobody asked to publish   1     <- the comment is NOT
3.75% base rate           1     <- and it carries the figure
7.9% market average       1     <- and the other one
```

This is the same defect the `{{if}}` guards exist to prevent, arriving by a route
the guards cannot see. It also explains the 07-31 misfire recorded above, where a
`LIKE '%3.75%'` probe said the base-rate claim was still on the homepage: **it was
right about the bytes and wrong about the page**, and the comment was the reason.

Fixed by moving the whole explanation inside the tool-doc sentinels — which the
`0` above proves are stripped — and leaving a two-line pointer in the markup with
no figures and no opinion.

> **MISSTEP: the validator caught a literal `{{if}}` in my prose.** The tool-doc
> block lives INSIDE the template, so writing "their `{{if}}` guards" in an
> explanation is real template syntax: `template: tool:161: missing value for if`.
> Second time today the same class caught me — earlier it was a literal `<div>` in
> a CSS comment, which `load_components.py` counts as markup when balancing tags.
> **Prose inside a template is not inert.** Spell the construct out in words.

### THE HARNESS HAD EXPIRED, and only luck made it loud

`verify_rewrite.py` — the gate the whole lane's equivalence claims rest on — takes
"the REAL hand-built page" from `~/projects/sites/loancalculator.co.uk`. **That is
the checkout the platform's own deploys commit re-rendered pages back into**
(`Rerender: tools/standard-calc.html`). So every page this lane successfully
decomposed replaced its own baseline.

It went unnoticed for a day only because the checkout was stale. Runs earlier
today were reading 2026-08-01 content and were genuinely valid. Another session
ran `git pull --rebase` at 10:19 and the very next run failed:

```
PREP-FAIL  standard-calc   opening-div regex matched 0 times (need exactly 1)
```

**It failed loudly by luck.** The cut patterns anchor on original markup that an
assembled page no longer has, and the harness requires each to match exactly once.
That rule was written to catch a mis-typed regex; it caught a moved baseline. With
looser patterns it would have spliced the component into a page that **already
contained it** and compared the rewrite against itself — passing, proving nothing,
on a gate this lane cites in every equivalence claim it has made.

Fixed by `git archive`-ing a pinned ref (`b4302e22b`, "repair the site before
adopting it into the platform", 2026-07-30 — the hand-built site as adopted, and
what `GOLDEN_2026-07-31c` was captured from). Verified file-for-file identical to
`803bf68c3`, the last pre-pull commit the earlier runs actually used, across every
page and asset the harness touches. Re-run on the pinned ref reproduces the
morning's result exactly: **9 of 11 MATCH, 2 DIVERGED on the intended keys only.**

**Same lesson as `PRE_FIX_REF`, three hours apart and in a different file: a
baseline that names a MOVING thing stops being a control, silently.** Once for a
git ref, once for a working copy. Fleet landmine written.

> **MISSTEP while diagnosing it, and it nearly sent me the wrong way.** I ran
> `git show <ref>:tools/standard-calc.html` — but the repo root is
> `~/projects/sites`, so the path is `<domain>/tools/...`. Every call failed with
> a `fatal:` that my own `2>/dev/null` swallowed, `grep -c` on the empty output
> returned `0`, and I read **four zeros across three refs as data** — concluding
> briefly that no ref anywhere held the original. The tell was that the numbers
> were *too* uniform: a working tree and three refs agreeing on 0 for a marker I
> could see with my own eyes in the file.

> **A same-file passenger, disclosed.** My `LANDMINES.md` commit `4cb891a7f`
> reports 3 deletions I did not write: another session was reformatting the
> `footprint:` line of the acceptance-callers entry from `·` separators to commas
> (the form `landmines-sync.py` parses). A pathspec commit cannot exclude a
> same-file edit. Checked before saying so: the change is complete, consistent
> with the other entries, and `--apply` ran clean after it. Nothing lost,
> forward-only holds.

### State at handover of this segment

- `tool-loan-repayment` template updated in `content_components`; **both** rows
  re-rendered offline and written (`index` 8571→9261 b, `tool-standard-calc`
  8796→9486 b — larger because the prose moved INTO the tool-doc block, which is
  stripped at assembly, so the SERVED pages shrink).
- Both controls REPRODUCED before either write.
- Assemble-only re-renders filed (`created_by='tidy-20260803'`). **Queued, not yet
  landed** — the queue went from 15 to 133 items deep while this was in flight.
  **Not verified live yet**; when it lands, expect `nobody asked to publish` → 0
  and both figures → 0 on `/index.html`.

## 2026-08-03 (later still) — the header-nav item, surveyed rather than built

Queue item (1) was "the header's link list is hand-maintained; generating
`site_components.header` from `pages` is the obvious next mechanism". **The survey
falsified the premise and then found something more urgent.**

### 1. There is no drift to fix. Yet.

```
header links: 25    pages: 27
IN HEADER, NOT A PAGE (dead links):  none
A PAGE, NOT IN HEADER:               /legal.html, /tools/standard-calc.html
```

Zero dead links, and both omissions are already known and deliberate: `/legal.html`
is in the FOOTER, and `/tools/standard-calc.html` is the orphan sitting on this
lane's own queue as a **content question for the owner**. So the hand-maintained
list is currently *correct*, and the risk it carries is latent, not realised.

That matters, because a generator would have had to decide the orphan question
mechanically — which is precisely the decision that is the owner's.

### 2. ⛔ THE MECHANISM ALREADY EXISTS, AND ON THIS SITE IT IS A DEMOLITION CHARGE

The platform's nav machinery is `nav-updater` → `populate_nav_tables` →
`site_nav_items` → chrome. Anyone acting on item (1) reaches for the agent whose
name says navigation. Read `classifyPagesForNav` at HEAD before you do:

```go
neverPrimary := neverPrimaryTypes[page.PageType] ||
    (isChildPageURL(page.URL) && !isSectionIndexType(page.PageType))
if neverPrimary {
    if page.InHeader || page.InFooter { utility = append(utility, page) }  // kept
    else { /* omitted — "the honest answer" */ }
    continue
}
```

Now this site, measured:

```
pages                     27   (13 guide, 12 tool, 1 content, 1 landing)
in_header = true           0
in_footer = true           0
declaring NEITHER         27   (explicitly false — not NULL, which is a third state)
site_nav_items rows        1   (against 25 links in the authored header)
```

Every `tool` page is `neverPrimaryTypes`. Every `/guides/` page is a child-path URL
whose `page_type` is not a section-index type. **All 27 declare neither flag, so all
27 take the "omitted" branch.** `populate_nav_tables` opens with
`DELETE FROM site_nav_items WHERE site_id = $1` and rebuilds — and `nav-updater`
then re-renders the chrome and re-assembles every deployed page, so the damage
ships immediately.

**This site is one `nav-updater` run away from a nav of roughly one link, on all 27
pages.** The fleet landmine for this is dated 2026-07-30 and was NARROWED on 07-31
to "only a page declaring neither flag is still lost" — which reads reassuring
right up until you measure a site where *that is every page*.

### 3. What was done instead: lock the authored chrome

The three `site_components` rows (head/header/footer) were authored by this lane and
were **unlocked**. Locking them is the same mechanism, precedent and rationale as
the twelve tool rows, and it is the protection that actually matches the threat.

Verified rather than assumed, in this order:

- **The chrome re-render honours the lock.** All three write sites in
  `render_site_components_action.go` (lines 569, 746, 913) carry
  `pageComponentAgentWritableSQL`. *(`CheckSiteComponentLock` itself has no
  production caller outside its test — a helper that looks like a finished
  refactor — but the predicate is in the WHERE of the writers, which is what
  actually bites.)*
- **The lock bites here**: `agent_may_write` = `f` on all three, evaluated with the
  platform's own predicate.
- **Negative control**: fleet-wide the same predicate returns `t` for **45 chrome
  rows across 15 sites** and `f` for exactly these 3. It is not constant-false.
- **Nothing on the wire moved**: 27 nav links on every page sampled, all HTTP 200.

**The stated cost is real and is already paid.** The guard's own file says "a
permanently locked header means new pages stop appearing in navigation on every
page of the site". On this site the header is hand-authored and never did update
itself — that IS item (1)'s complaint. So the lock costs nothing that was not
already true, makes it explicit, and converts a silent overwrite into a
`lock_blocked_change` for review.

### 4. So item (1) is now a different, larger job — and it has a precondition

Generating the header is no longer "write a script". It is:

1. **Declare nav membership**: set `in_header`/`in_footer` on the 25+2 pages to
   match what the authored chrome already does. This is the precondition for ANY
   platform nav mechanism working here, and it is what disarms the demolition
   charge in §2.
2. Decide the orphan (`/tools/standard-calc.html`) — **owner's call**, and it is an
   input to step 1, not an output of it.
3. Only then consider generating chrome from the nav tables, via `nav-link-fixer`
   (refreshes chrome from the EXISTING tables, no populate step) — never
   `nav-updater`, which repopulates.
4. Unlock the chrome for that deliberate act, then re-lock.

**Step 1 was deliberately NOT done in this session.** It writes to 27 rows on a live
money site, and `in_header` has consumers beyond the classifier — `buildServicesHTML`
(`render_site_components_action.go:1156`) selects pages with
`in_header = true OR in_footer = true` to build a footer "Our Services" column.
Today that is inert here (the authored chrome carries no placeholders and is now
locked), but "inert today" is a fact about today, not a guard — the same reasoning
this lane got wrong once already this session. It needs its own change with its own
verification, and the orphan decision first.

## 2026-08-03 (platform thread) — bugs_open/182 fixed and live; a second defect found inducing it

A different thread (`bugfix bugs_open/182`) picked up the platform bug this lane
filed and shipped the fix: `RerenderPageSectionsAction` now resolves by
`page_components.component_id` first, falling back to `slot_name`. LIVE on
chassis v1.0.1240, pod-verified both replicas. This site's positional slots
(`prose-0`, `tool-2`, etc.) are the population that fix repairs — 63 of 63
slots here, plus 2 on oufe.com.

**A second, DIFFERENT defect surfaced inducing the verification** — filed as
`bugs_open/189`. Firing `section_data_resolved` on `tool-loan-vs-savings`
(this site's own documented re-render, exactly as this NOTES file describes it
above) DUPLICATED the calculator on the page: the fresh render inserted
alongside the pre-existing locked row, instead of the lock guard discarding
the fresh copy as `bugs_closed/058` is supposed to guarantee. Root cause:
`save_page_sections` matches a locked row by name, and 182's fix means the
persisted name now silently changes (`extractSectionsFromMetadata` prefers the
resolved component's own identity over the stored slot_name) — so the lock
match misses. **Remediated live in the same session** (duplicate row deleted,
locked row repositioned back, prose slot_names restored, assemble-only
redeploy fired) — this page is back to 4 sections, content unchanged.

**Do NOT fire `section_data_resolved` on this site's other 12 locked
sections** (or oufe.com's 2) until `bugs_open/189` is fixed — each one will
reproduce the same duplication the first time it's touched. The four
calculator fixes this lane already shipped via the offline route
(`decompose/render_tool_row.py`) are unaffected — that route never goes
through `save_page_sections`' locked-row matching at all. Check
`bugs_open/189` for the current fix status before relying on the documented
re-render route again for a LOCKED section on this site.

## 2026-08-03 — the owner's two rulings, and what carrying them out turned up

**Ruling 1: delete the orphan.** **Ruling 2: keep withholding — "we need honesty at
any cost."** Ruling 2 needed no code: the withhold behaviour was already live and
proven (`defect_vectors.py --live`, 8 of 8 on the served pages). Recorded so it is
not re-litigated.

### Retiring `/tools/standard-calc.html`

Done and verified on the wire: **404**, sitemap **27 → 26** with zero mentions, the
other 26 pages all 200.

**The audit that authorised it — and the one that nearly misled me.** My first probe
was `rendered_html LIKE '%standard-calc%'` and it reported inbound links from
`/index.html` and the footer. **Both were HTML COMMENTS** — one a previous session's
footer note saying the page is "deliberately NOT linked", one my OWN tool-doc text
written three hours earlier. This lane already has that landmine written down (*"a
LIKE probe over rendered_html cannot tell markup from a comment"*) and I walked into
it anyway, in a query whose answer would have blocked a deletion.

The `bugfix_098` lane's audit is the correct one and I used it: match `href="<url>"`,
across **page bodies, chrome AND nav** — their note that a body-only census reports
most of a site as unreferenced is exactly right — and always with a positive control.
Result: 0 / 0 / 0 for the target; 1 body + 2 chrome for `/tools/consolidation.html`.

**Order matters, and the sitemap is the load-bearing half.** Archive in the platform
FIRST (`pages.status='archived'`) so nothing re-publishes it, then remove the file.
But this site's `sitemap.xml` was last written by the **adoption commit** — the
platform has never regenerated it — and `retract_page_deployment` deliberately
excludes `sitemap.xml` as a file `pages` does not model. So the platform primitive
alone would have left a sitemap advertising a 404, **which is precisely the defect
`bugs_open/098` exists to fix**. File and sitemap entry went in one commit.

> **`retract_page_deployment` is LIVE but UNREACHABLE.** Pod-grep: 6 hits, positive
> control 20, negative control 0. But it is wired into **no agent** — `SELECT` over
> `agent_definitions` for it and for `delete_file` both return 0 — so it is a
> capability with no caller, deliberately (its own PLAN says so). Its acceptance
> target `robot-hands.com/learning-center/index.html` still served 200 when I
> checked, so it has never been exercised. I did not seat an agent to reach it:
> that is the 098 lane's platform change to make, not mine to ride.

> **MISSTEP — I nearly reported a push that had not happened.** `git push` was
> REJECTED (non-fast-forward, then `cannot lock ref` — the sites repo takes
> concurrent pushes constantly), and my command still printed a cheerful
> "pushed: <local sha>" because it echoed `git log -1` of my LOCAL branch. Two
> rejections before it landed. **Verify at `origin/master`, never at the local ref**
> — `git ls-tree origin/master` for absence and `git show origin/master:` for the
> sitemap count.

### The footer's comment shipped on all 27 pages, and the retirement made it FALSE

`chrome/footer.html` opened with a 777-character engineering note explaining which
pages are orphaned and why the footer exists. **Authored chrome has no tool-doc
sentinel mechanism**, so unlike a component template, anything written in it is
published verbatim — on every page of the site. The moment the orphan went, that
note asserted the existence of a page that 404s, site-wide.

Cut to 138 characters pointing at the RUNBOOK, which already carried the reasoning.
Reloaded through the unlock → `load_chrome.py --apply` → re-lock procedure written
this morning (its first use; it worked). Footer 3147 → 2504 b, re-locked, no slot
mentions the retired page.

**Stated rather than hidden:** the 26 live pages still serve the OLD footer until
each next re-renders. The source is correct and it self-heals; 26 forced deploys for
an invisible comment is not proportionate. Third instance today of the same class —
engineering prose in a shipped artefact — after `tool-loan-repayment` twice.

## 2026-08-04 — post-roll check on v1.0.1250, and a baseline I contaminated myself

**Result: the roll moved nothing.** Three pages re-rendered on the new image; the
only difference on any of them is my own footer comment propagating (12 lines,
identical diff on all three). `toolgolden --compare`: **11 of 11 reproduce exactly.**

The render-path diff across this roll was **not** empty — `12ae5824f fix(187)`
touches `rerender_page_sections_action.go` — so it could not be waved through on the
usual "empty means it cannot move them" test. Reading it, the change is confined to
the **escalation** branch (it guards a false-alarm `needs_page` and returns a
disposition) and touches neither assembly nor template rendering. Pleasant to note:
it cites `bugs_open/182` as its reason for naming a no-op, so that finding is
already being used by another lane.

### MISSTEP — I captured a baseline from pages with a KNOWN pending change

All three pages came back DIFFERENT, which is the alarm signal. The cause was mine:
the footer fix had been sitting in `site_components` since 08-03 waiting for each
page's next re-render, and **I had written that down as an open item in the same
handoff, hours before running the check.** So "DIFFERS" was guaranteed before the
chassis was involved at all. I had built a check whose baseline could not hold.

It survived only because the diff was unambiguous and attributable — one contiguous
comment block, identical on all three pages. **A subtler pending change would have
produced a mismatch I could not have attributed**, and the standing advice for a
mismatch ("re-run first; propagation lag clears, a real fault reproduces") would
have pointed the wrong way here: re-running reproduces this every time, so it reads
as a real fault, and it is not.

> **A baseline is only a baseline if nothing else is in flight.** Before a post-roll
> byte check, establish that there is nothing pending — either let everything
> propagate first, or predict the expected diff and assert it exactly rather than
> asserting equality. The second is strictly better when something IS pending, and
> it is what I ended up doing after the fact instead of before.

Same family as the two baseline failures on 08-03 (`git show HEAD:`; the
`~/projects/sites` working copy) — but a different mechanism. Those two were
baselines that **moved**; this one was a baseline that was **wrong when taken**.
Three instances in two days says the class is worth naming: *know what your
"before" is a picture of, and what else is in flight when you take it.*

**Side effect worth recording:** 3 of 26 pages (`index`, `tool-consolidation`,
`tool-loan-vs-savings`) now carry the corrected footer; the other 23 still serve the
old comment until they next re-render.

## 2026-08-04 (evening) — v1.0.1251 post-roll: CLEAN baseline this time, and a handoff that contradicted its own NOTES

**A second roll landed the same day.** The handoff written this morning covers
v1.0.1250 (rolled 10:29Z); the pods now run **v1.0.1251, started 19:19Z**. So a
second post-roll check was owed, and this one could be done properly.

### The baseline problem from this morning solved itself, and I used it

This morning's misstep was a baseline taken from pages with a **known pending
change** (the corrected footer waiting on each page's next re-render). The fix I
wrote down was: *either let everything propagate first, or predict the diff and
assert it exactly.*

There was a third option available and I took it: **the three pages that already
picked up the footer this morning have nothing pending.** `index`,
`tool-consolidation` and `tool-loan-vs-savings` are, as of this morning, the only
three pages on the site in a settled state — so they are the only correct choice of
baseline page, and for once the check needed no prediction at all.

Doubly anchored before firing anything: today's live bytes are byte-identical to
the `.post` files v1.0.1250 produced this morning (`cmp`, all three), so the
baseline is both *current* and *attributable to the old image*.

### Reachability first — the render-path diff is NOT empty, again

Between the rolls (local 11:20 → 20:20) these changed under `platform/`:

```
chrome_link_policy.go              (new, bugs_open/191)
render_site_components_action.go
nav_tables.go
resolve_internal_links_action.go
load_current_section_content_action.go
```

Four of those are chrome/nav, which on this site is the **most sensitive surface
there is** — 25 authored header links, chrome locked. So this needed reading, not
waving through.

**It cannot reach an assembled page.** `assemblePage`
(`rerender_single_page_action.go:532`) loads chrome via `getSiteComponents`, which is
a plain `SELECT slot_name, rendered_html FROM site_components` — it **reads stored
chrome, it never renders it**. `LoadChromeLinkPolicy` has exactly three callers
(`render_site_components_action.go:179`, `nav_tables.go:193`, and a comment in
`resolve_internal_links_action.go:496`), none of them on the assembly path. And the
RUNBOOK's named render-path files (`rerender_single_page_action.go`,
`rerender_link_repair.go`, `tool_doc_header.go`) show **no commits at all** in the
window.

So: expected no movement. **Expected is not measured**, so:

### Measured — all three IDENTICAL

```
index.html                  05dee40c992d -> 05dee40c992d  IDENTICAL
tools/consolidation.html    be7bd8586779 -> be7bd8586779  IDENTICAL
tools/loan-vs-savings.html  fcf803a8e1e8 -> fcf803a8e1e8  IDENTICAL
```

`toolgolden --compare` against `GOLDEN_2026-08-03b`: **11 of 11 reproduce exactly**
(re-run after all three re-renders had landed, not during — the first run overlapped
two of them and I did not want a claim resting on that). 12 of 12 spot-checked pages
200. `tool-loan-vs-savings` still has **4** `<section>` blocks, not 5, so `189` did
not bite — as expected, since assemble-only never enters `save_page_sections`.

**Contrast with this morning worth keeping:** same site, same check, same three
pages. This morning it read DIFFERS and needed a diff to exonerate; this evening it
reads IDENTICAL and needs nothing. The only variable was whether the baseline had
something pending in it.

### The handoff I wrote this morning was contradicted by this file

`HANDOFF_2026-08-04` §2 item (3) said *"`bugs_open/182` is owned by another thread …
it is why `rerender_sections` is a no-op here"*, and §3 repeated the no-op claim as
the standing procedure. **Both were already false when written.** `182` was fixed
(`a43be1e70`), shipped on v1.0.1240, and moved to `bugs_closed/` on 08-03 — and the
`2026-08-03 (platform thread)` entry in *this file*, several screens above, says so
in full, along with `bugs_open/189`.

Pod-grepped rather than inherited, both replicas on v1.0.1251:
`id_resolved_component` = 1, positive control (`assemblePage: Complete`) = 1,
negative control = 0.

**The correction is not "182 moved".** It is that the danger **inverted** while the
instruction stayed the same. Before: firing `rerender_sections` here did nothing.
Now: it resolves this site's 57 positional slots, and on a **locked** row it
duplicates the section on the page (`bugs_open/189`, still open —
`extractSectionsFromMetadata` is unchanged since 06-17, checked). A reader who took
§3 at face value would have concluded the ordinary route was *harmless* here, which
is the most expensive possible way to be wrong about it.

> **What this says about handoffs.** The stale claim did not come from missing
> information — the correct version was in my own lane's NOTES, written by me, hours
> earlier. It came from carrying §3 forward verbatim from the 08-03 handoff while
> writing a new one. **A handoff assembled by copying its predecessor inherits its
> predecessor's expiry dates.** The cheap check is to re-read the NOTES tail
> *against* the handoff before publishing it, which is the one thing I did not do.

⚠ **`189` is now an ambiguous number** — `bugs_open/` holds two unrelated files
under it (this lane's 08-03 duplication bug, and the 163 lane's 08-04
`siblingSignatures` parser duplicate). Every commit message saying "189" since 08-04
means the other one. Resolve by slug; `git log` the file path.

### Footer propagation (open item 2) — measured, then acted on

Census over all 26 active pages, `THE HAND-BUILT PAGES HAD NO FOOTER` (old) against
`Site footer. Rationale` (new): **OLD 23, NEW 3, NEITHER 0.** Both arms non-zero and
they partition the site, so the probe discriminates rather than merely agreeing with
me.

Re-weighed the "disproportionate" judgement from 08-03 and reversed it, on two facts
that were not on the table then: the old note **asserts `/legal.html` is orphaned and
describes the page that now 404s**, so 23 live pages publish a claim that went false
on 08-03; and it is a **standing baseline contaminant** — it is precisely what broke
this morning's check, and it would break the next one on any of those 23 pages.

### …and then the dispatch lane stalled under me, so item (2) is HALF DONE

Fired two canaries first rather than 23 blind — `/legal.html` (prose) and
`/tools/overpayment-calculator.html` (tool) — the point being to assert that the
footer comment is the **only** thing pending on those 23 pages before bulk-firing.
That is the same discipline §6b's baseline lesson asks for, applied forwards.

Both are still `triaged` nine minutes later, and the reason is not mine:

```
page_rerender, last 3h:   complete 3 (mine, 19:41-19:44)   triaged 211
claims in the last 11 min: 0
of the 211:  webdesign.co.uk 156 · gaswholesalers.com 52 · loancalculator.co.uk 2 (mine)
```

**It is not a global stall, which is the interesting part.** In the same window
`page_component_status_drift` completed 12, `acceptance_run` 6, `needs_design_review`
4 — and the `build-dispatch-loop` orchestration itself is cycling normally
(19:46:26→19:48:39 COMPLETED, 19:48:55→19:52:12 COMPLETED, two more AWAITING). So
the loop runs and completes while `page_rerender` specifically does not get claimed.

**What I did NOT do, deliberately, and why:**

- **Did not fire the remaining 21.** Adding 21 unverifiable items to a lane that is
  not draining buys nothing and leaves the next thread unable to tell my items from
  the flood.
- **Did not re-fire the canaries.** The memory note on this is explicit and was
  written after someone got it wrong: *a missing/slow row is not a lost message*, and
  **the single-page bypass (049b, the 086 envelope) shares this very lane**, so it
  cannot beat a stalled dispatch. Re-firing would produce duplicates, not a deploy.
- **Did not fork a diagnosis.** `bugs_closed/030` is the closed fleet-wide version of
  this shape. This is not obviously the same thing — 030 was *everything* stalled,
  this is one item type while others drain — but establishing that properly is a
  `090` run and an owned-lane check, not a paragraph asserted from a bug's own lane.
  Recorded as a measurement, deliberately not as a root cause. **[MEASURED]** for
  the counts above; **[UNDIAGNOSED]** for the cause.

**So item (2) is genuinely half-finished and is written up that way** — 2 of 23 in
flight, 21 not filed, with the canary-diff assertion still owed before the bulk.
The next thread's first act should be to read the canary status, not to re-fire.

### The lane recovered — and the two canaries DISAGREED, which is why there were two

The stall cleared on its own after ~15 minutes (no intervention, nothing re-fired).
Both canaries ran. And the canary step did exactly the job it was there to do: **my
prediction was wrong, and it caught it before 21 pages shipped on it.**

I predicted "the only diff will be the footer comment block". True for one canary,
false for the other:

```
tools/overpayment-calculator   footer only            (10 lines out, 2 in)
legal.html                     footer + TWO MORE      (11 lines out, 4 in)
                               + <link rel="canonical" href="…legal.html">   ADDED
                               - <meta name="description" content="">        REMOVED
```

**The 23 pages are not homogeneous.** Each carries whatever platform improvements
have landed since *it* last re-rendered, so the pending set differs page by page. A
single canary — either one — would have licensed a confident and wrong prediction.

Attributed rather than guessed: both extra changes come from one commit,
`9c7a8e9e4 seo(assembly): emit canonicals + make the blank meta description
correct-or-absent` (2026-08-02). `injectCanonicalLink` and `spliceMetaDescription`
are both called from `assemblePage`, so any page not re-rendered since 08-02 has
neither.

**Census over all 26 active pages** (served, not stored): **20 of 26 had NO
`<link rel="canonical">` and carried an empty `<meta name="description" content="">`.**
Of the 21 remaining targets, **20 of 21**.

> **This re-frames open item (2) completely, and upward.** It was recorded on 08-03 as
> "an invisible comment, 26 forced deploys is disproportionate" — a fair call on the
> facts then known. The actual pending payload is **a canonical tag and a correct
> meta-description on 20 pages of a live money site**, absent for two days. That is
> not cosmetic, and the deploys are not disproportionate.
>
> **The general lesson is worth more than the instance:** *"only a cosmetic change is
> pending"* was an **inference about what had accumulated**, not a measurement of it.
> Nobody had asked what ELSE was waiting behind the same door. On a platform where
> improvements land continuously and apply on next render, the pending set on a stale
> page is **whatever shipped since it last rendered** — which is unknowable without
> looking, and grows every day the page sits.

**Prediction for the bulk, stated BEFORE firing** (so it can come out false): footer
on all 21; canonical added and empty-description removed on 20 of 21; the 21st
footer-only. Baselines in `/tmp/decomp-work/postroll1251/bulk/*.pre`.

> **CORRECTED within the minute — I named the wrong page, in the paragraph about not
> inferring.** I first wrote that the odd page out was `tools/credit-roadmap.html`.
> It is **`tools/car-finance-calculator.html`** — the only one of the 21 that already
> carries a canonical and has no empty-description tag, i.e. the only one re-rendered
> since 2026-08-02. I had the 20/21 count from a real loop but reached for the page
> NAME from memory instead of printing it, one command after writing that an
> inference had been stated as a measurement. What caught it: re-reading my own
> sentence and running the two-line loop that prints the exception by name.
> **A count you measured does not make the label you attached to it measured.**

### The bulk verification: my prediction held, but I refuted it wrongly first

Checking the first four completed pages, two read `canonical 0->1` and two read
`canonical 0->0`. All four had identical URL shape, title and domain, so I treated the
split as real and went looking for the mechanism: read `injectCanonicalLink`'s five
skip conditions, then `injectPageJSONLD`'s, then `getPageInfo`'s query — and found the
two injectors share a `page.Domain` input. Then checked JSON-LD as a discriminator:
**absent on exactly the two pages missing the canonical, present on the other two.**

That is a clean, corroborated story, and I was a step from writing up *"`PageInfo.Domain`
is unpopulated for some pages, silently suppressing both canonical and structured
data"* — a fleet-wide discoverability defect.

**There is no defect. The two files were not pages.** They were seven lines of
`{"error":"B2 returned error"…"NoSuchKey"}`, HTTP 200, because I fetched them inside
their own deploy window. Re-fetched after propagation, all four read
`canonical=1 jsonld=1 emptydesc=0 oldfooter=0`. **The prediction was right the whole
time.**

> **The corroboration was the trap, not the alarm.** I picked the JSON-LD check
> *because* it shares `Domain` with the canonical path, so agreement would be
> evidence — the right instinct. But both signals were computed **from the same
> 7-line file**, so they could not possibly have disagreed. A second measurement over
> a corrupt artefact is not a second measurement. Same family as the standing "two
> checks blind the SAME way AGREE" rule, walked into while trying to apply it.
>
> **And it defeated my own landmine, written 40 minutes earlier in this session.**
> That entry warns a zero has two causes — clean, or broken probe. **There is a
> third: the artefact is not there.** Every grep returned 0, *including the two I
> wanted at 0* (`oldfooter`, `emptydesc`), which is exactly why the row read as a
> partial success rather than a failure. Both the landmine and the handoff recipe now
> carry a `wc -c` + `DOCTYPE` guard as the first step, before any grep.
>
> `complete` is the WORK ITEM's status, not the CDN's. The lane's own guidance already
> said ~2 min; I queried on `status='complete'` and fetched immediately.

### Item (2) CLOSED — 23 of 23, and the acceptance test proved itself before it passed

All 21 bulk re-renders completed **first attempt, zero retries** — which is worth
noting given the lane spent the evening intermittent: the items were always fine, only
the claiming was slow. Nothing was re-fired.

**Acceptance census, all 26 served pages: `old footer 0 · no-canonical 0 ·
empty-description 0`.**

```
                    before      after
old footer          23 of 26      0
no canonical        20 of 26      0
empty <meta desc>   20 of 26      0
```

**The control was run FIRST and fired all three arms** against a known-stale baseline
page. That is the difference between a clean sweep and a silent probe — and after the
B2-blob episode an hour earlier, running it was not optional.

Re-confirmed after the 21 re-renders: **11/11 calculators exact** against
`GOLDEN_2026-08-03b`, 26/26 HTTP 200, retired page still 404, `tool-loan-vs-savings`
still **4** `<section>` blocks (so `bugs_open/189` did not bite — assemble-only never
enters `save_page_sections`, as predicted).

The site now carries everything the platform currently emits, so §6b's
"use-a-settled-page-as-baseline" constraint is lifted. ⚠ It will come back on its own:
**a platform improvement landing in `assemblePage` puts every un-re-rendered page back
into the pending state with nobody on this lane touching anything.** That is the whole
lesson of the day, and it is not a one-off condition to be cleared but a standing one
to be re-checked.

## 2026-08-05 — the site's copy goes back through the FRAMEWORK, and the reason it could not

**Owner direction (2026-08-05):** rerun loancalculator's copy through the framework
in the "gentle explanatory" (H) voice the `portfolio_positioning` lane developed with
him this morning — **explicitly not hand-written through this CLI**, per the standing
ruling of 2026-08-04. Scope ruled: **copy only, calculators kept.**

⚠ **Do not touch `loanandmortgagecalculator.co.uk`.** Session `fffe0948`
(portfolio_positioning) seeded the same voice there at **10:54:58Z today** and was
live while this was written. Its voice prompt is
`portfolio_positioning/VOICE_gentle_explanatory_v1.md`.

### The framework could not rewrite this site's prose, and it would have said it did

Before firing anything, traced the writer path end to end. Three links, each read
from live config or source rather than inferred:

```
content_components.ported-prose.input_schema.fields.content.source = "authored"
  -> plan_sections_action.go:1708   appends to llmFieldSpecs only `if source == "llm"`
  -> :777  `json:"llm_field_specs,omitempty"`  so an EMPTY list serialises as ABSENT
  -> live page-content-writer step check_render_mode:
       condition: "current_section.llm_field_specs != null"
       else_step: "render_from_template"      <- no LLM is called at all
```

So every one of the 51 prose sections would have taken the **template** branch: the
stored prose re-rendered verbatim, work item `complete`, bytes identical. **A clean
no-op that reports success**, which is worse than a crash — the natural reading would
have been "the model ignored the guidance" or "the seed never reached the prompt",
and both would have been wrong, because no model ran.

**There are TWO independent ways for a voice change here to be a silent no-op and they
present identically:** (1) `content_direction.formatted` not regenerated — the writer
reads *only* that field; (2) the field's `source` not being `"llm"`. Rule them out in
that order. Both are now in `LANDMINES.md`.

### The owner's ruling on the label, which is better than the argument I was building

I was preparing to justify flipping `source` as a contained config change. The owner
cut through it: **the prose is not authored.** `source: "authored"` asserts *a human
supplied this, do not regenerate* — and this prose was **another LLM's output**,
written outside the framework and without its checks, then lifted byte-for-byte by
the decomposer. So the label was simply **false**, and correcting it is a fix rather
than a loosening. That is a much stronger footing, and it is the one recorded.

**Measured before changing anything, not after:**

- `ported-prose` is used by **loancalculator.co.uk and no other site** — 51 rows,
  zero elsewhere.
- Exactly **2** fields fleet-wide carried `source: "authored"`. After the change: **1**.
  That count is the negative control proving exactly one field moved.
- The survivor is **`ported-page.body`, and it must stay `authored`** — that is the
  `--fidelity locked` byte-preserving adoption path (`adopt_verbatim.go`, ADO-037).
  Flipping it would let a writer rewrite a site adopted precisely to be preserved.
  Two authored fields; exactly one of them was wrong.

`llm_guidance` was rewritten at the same time to be writer-facing and to carry the
invariant the decomposer designed in: prose only, no form control, no element a script
addresses, no ids, no scripts — *"rewriting this cannot break a calculator"* — plus
"preserve every factual claim, figure and internal link: this is a rewrite for VOICE,
not a change of substance."

### Seeding the voice: the merge is NOT additive, and that is the whole difficulty

`seed_voice_h.py` (this lane). Two safety properties, both load-bearing:

1. **The formatter gate.** The writer reads one field, `content_direction.formatted`,
   produced by `datahelpers.FormatContentDirection`. The script ports that function to
   Python and, before writing, regenerates the CURRENT spec and asserts it reproduces
   the STORED `formatted` as a multiset of lines (not a string — Go map iteration order
   is random, so section order carries no meaning). **PASSED at 20,699 bytes / 141
   lines.** If the port ever drifts from the Go, it refuses to write rather than
   silently changing what the writer sees.
2. **Conflicts are REPLACED, not appended.** loancalculator's incumbent spec said
   *"Avoid contractions in declarative or authoritative statements"*; H rule 8 says
   contractions wherever they would be spoken. Appending would have handed the writer
   two opposing instructions and let the model choose. Three incumbent rules were
   replaced (contractions, rhetorical-question openers, "lead with the reader's most
   likely question"), eight added, `voice.register`/`formality`/`emotional_tone`
   rewritten. The script **refuses to run** if an expected incumbent is not found — a
   spec that has changed under it means the conflict may still be live under different
   wording. A final check asserts the retired no-contractions rule is absent from the
   regenerated `formatted`.

`writing_rules` 15 → 23; `formatted` 20,699 → 24,556 bytes; `has_H=true` verified on
the stored row. Exemplars are **this site's own copy** (consolidation, car finance,
overpayment), not another site's — per the voice doc's rule 2 for per-site reuse.

Backups first: `content_components_bak_20260805_prosesource` (2 rows),
`page_components_bak_20260805_framework_rewrite` (63 rows).

---

## 2026-08-05 — CONTRIB from the mortgagecalculator adoption lane: toolgolden's uniform vectors falsely convict RATIO tools; fixed with a fourth, asymmetric vector

Running your harness against mortgagecalculator.co.uk's 12 original tools, gate B
refused `investor.html`: *"reacts, but output is identical for every input value —
arithmetic ignores its inputs"*. **The conviction was false.** Both its calculators
are pure ratios — yield = rent×12/price, LTV = loan/price — and every vector scales
ALL fields by one shared factor (×1, ×2, ×0.5), which a quotient cannot see:
(2r×12)/(2p) ≡ (r×12)/p. The arithmetic uses its inputs perfectly; the harness's
question could not distinguish that from ignoring them.

**Fix, in `toolgolden.py`:** a fourth vector `("asym", "asym")` — each numeric field
scaled by a different deterministic factor (`[1.7, 0.6, 2.3, 1.1, 3.1, 0.45]`, cycled
by document order). Gate B now also compares defaults↔asym; ratios move. Guards added
so a pre-asym three-vector golden still loads, compares (with a printed NOTE) and emits.

**Non-regression proven per TL-038's own landmine discipline** (re-capture the corpus
and diff): `--compare` of your `GOLDEN_2026-08-03b_after_orphan_retired.json` across all
11 pages — **11/11 MATCHES** on the three shared vectors, so existing driving is
byte-identical. Result on the motivating case: investor.html vary 0→1, golden written;
your future re-captures also gain coverage (mortgagecalculator's portfolio.html went
vary 4→8 with asym in play).

Worth knowing for your corpus: any future loancalculator tool computing a pure ratio
(LTV checker, debt-to-income) would have hit this same false refusal.

### The canary: INDUCED, and it found a second blocker underneath the first

Fired one real `content_rewrite` (`created_by='voiceh-canary'`) at
`guide-how-loans-are-calculated` — 2 prose blocks, no calculator, so a failure could
not touch arithmetic. It did **not** rewrite, and the reason is not the one I had
just fixed:

```
page-build-handler no-op: no sections ready to build (empty spec sections,
or all sections deferred for missing data) — the target section was NOT rebuilt
```

**Credit where due: it refused LOUDLY.** No silent success. (That is the shape
`bugs_open/194`'s framework half shipped for.)

**The cause is `bugs_closed/182` again, in the sibling call site.** `pages.sections`
on this site is `["prose-0","prose-1"]` — **positional slot names**. `plan_sections`
resolves those against component **name/function**: `loadComponentSchemas`
(`plan_sections_action.go:1144`) indexes the map "by both name and function", and
`:918` does `components[sectionName]`. `prose-0` is neither a name nor a function, so
it misses, falls to the selector at `:937`, and the section defers.

**Measured, not inferred: 0 of 57 section names on this site resolve.** Fleet-wide:

```
loancalculator.co.uk          57 section names   57 unresolvable  (100%)
gaswholesalers.com           122                 11
finetuning.uk                152                 10
leopardessconsulting.co.uk   106                  6
oufe.com                      20                  2
```

> **The sharpest detail: `a43be1e70` (182's fix) EDITED THIS VERY FILE** — it factored
> `componentInfoFromRaw` so the truncation guard could not drift "across the three now
> shared conversion sites" — **and still only added `component_id`-first resolution to
> the re-render path.** The build path was refactored around and left heuristic. That
> is the documented "one call site of a shared judgement gets the rigorous fix, the
> sibling stays heuristic" shape (016b §9), caught here by induction rather than review.
> Re-checked at v1.0.1254: `plan_sections_action.go` untouched since 08-04, so this is
> still open and unowned.

**⚠ It does not just fail — it asks the fleet to manufacture junk.** The selector read
`prose-0` as an unknown *component type* and filed `needs_new_component` work items to
build components literally named `prose-0` and `prose-1`, plus two `needs_section_data`
HITL items. **All four cancelled** with an explanatory note before a component-creator
could act. Anyone pushing a decomposed site through the build path should expect this
and check for it afterwards.

### The base prompt is SEVEN prompts, and they have already drifted

For the owner's fleet-wide voice decision. The `## HOUSE VOICE` block is not one
shared prompt — seven live agents each carry their own copy, and no two are the same:

```
content-writer                        1046 b   563e678a…
content-creator-hero-without-research 1243 b   224a2008…
content-creator-about                 1250 b   8f117bcc…
content-creator-hero                  1252 b   ea4736ea…
simple-content-writer-with-approval   2334 b   cba1f868…
grounded-explainer                    2866 b   89221f73…
page-content-writer                   4657 b   d4b409e1…
```

All seven also carry the rule **"Start with the fact"**, which is in direct conflict
with H's rule 1 (*open where the reader is standing… before the first assertion*).
They agree on banning the negative-twist opener, so the reconciliation is to revise
the opening rule and keep the ban — but it is a REVISION of a live default, not an
addition, on seven divergent copies. Options for the council submission: seven edits
that will drift again, or one shared carrier both/all read at prompt-assembly time
(the `footer_compliance_lines` carrier the portfolio lane built is the local precedent).

### Post-roll, v1.0.1254 (rolled 2026-08-05T20:41Z)

26/26 HTTP 200; the 08-04 propagation holds — old footer 0, no-canonical 0,
empty-description 0 across all 26. Nothing has re-rendered since the roll, so served
bytes are necessarily unchanged; **the byte check proper (re-render on the new image,
compare) is OWED** — see the handoff.

---

## 2026-08-08 — the voice H rollout begins: owner approved the canary, and the blocker is gone

**Owner ruling:** *"yes voice H is much better (still not perfect but is a huge
improvement and we can look again later) please go ahead with the rewrites."*
So: roll the voice across the site's copy **as it stands**, do not tune the prompt
mid-run. Plan with the phasing and the grading table:
`PLAN_2026-08-08_voice_h_rollout.md`.

The 08-05 blocker (`bugs_open/204`) and the trap underneath it (`bugs_open/189`) were
both fixed and proven by the `bug_backlog_clearing` lane on 08-06. Nothing in this lane
was needed for that; it just became possible.

### Pre-flight, re-run at v1.0.1263 (a fresh build landed mid-setup)

Every precondition re-measured against the NEW image rather than carried over, because
a roll is not evidence and the writer's own row had been updated by another session at
08-07 05:46.

```
chassis            v1.0.1263, pods started 08:54:57Z / 08:55:18Z (14 min — past the ~300s drop window)
binary markers     "load page slot identities" 1 · "stored_slot_name" 1 · fabricated control 0
voice H            site_specs content_direction: formatted 24,556 b, 23 rules,
                   has_H marker true, retired "Avoid contractions" rule absent
ported-prose       input_schema.fields.content.source = "llm"   (the 08-05 correction holds)
slot_name_from     present on BOTH render_section and render_from_template
```

⚠ **The `slot_name_from` check read NULL the first time and it was MY path that was
wrong, not the config.** The keys are not on `workflow.steps.*` — they are nested at
`workflow.steps.process_sections_loop.config.sub_workflow.steps.*.config`. A jsonb path
read that misses returns NULL, which is indistinguishable from "another session reverted
it" — and that revert was a live, documented risk on this exact row, so the wrong
reading was extremely believable. **Before concluding a key is absent, prove the path
resolves at all**: `default_config::text LIKE '%slot_name_from%'` returned 1 in one
query and settled it. Going to `LANDMINES.md`.

Also checked, because it would have changed the output: `330_page_content_writer_prompt_v4_scoped_facts`
(bugs_open/151's writer half, per-section fact scoping) is **NOT applied** — the file
carries a `_HOLD` suffix, `agent_definitions_bak_330` does not exist, and the live
template does not contain its anchor string. So the writer prompt is the same one the
owner's canary was produced by.

### The route, and the two things that make it repeatable

`voiceh_rewrite.sh <page-name>` (this lane). It copies the prompt **by SQL from the
canary work item `2517bc4b`** and substitutes only page/page_id/page_name, so what the
owner approved is bit-for-bit what every page gets — no retyping, no drift. It files the
item `status='detected'` (the dispatcher selects `triaged`/`approved`, so nothing else
can pick it up) and the direct Kafka publish is the only dispatch.

**MISSTEP, caught on the first run:** `INSERT … RETURNING id` prints its `INSERT 0 1`
command tag on stdout *even under `-t -A`*, so the captured variable was
`"<uuid>\nINSERT 0 1"` and the next query died with `invalid input syntax for type
uuid`. Loud, cheap, fixed by wrapping the insert in a CTE (`WITH ins AS (INSERT …
RETURNING id) SELECT id FROM ins;`) — a SELECT prints no tag under `-t -A`. Worth
knowing because the failure mode is *not* always loud: had the variable been
interpolated into a `LIKE` or a text column it would have silently carried the tag.
The orphaned item from that run was deleted before re-firing.

### Predicted before firing: what happens to the 11 calculator pages

Traced, not assumed, so that a surprise would be a real signal:

- all 12 tool components have **0 `source='llm'` fields**, so `llm_field_specs`
  serialises absent → `check_render_mode` takes `render_from_template` → **no LLM runs
  on a tool section**;
- `save_page_sections` preloads actively-locked rows, holds them out of the DELETE and
  lets the locked copy stand, so the row keeps its id **and** its `updated_at`.

Protected twice. Expected side effect: one `lock_blocked_change` item at
`needs_human_review` per preserved lock — up to 12 — true about the mechanism, spurious
in substance (the "blocked" change was a byte-identical re-render). Cancel with a note
at the end rather than leaving them looking like real HITL work.

### Rollback assets, taken before anything fired

`page_components_bak_20260807_voiceh` (63 rows) · `baseline_20260807.json` (76 KB —
every row's full text, length, md5, row id, so it serves fact-by-fact comparison as
well as restore).

### The guides are done — 13 of 13 — and my own checker was wrong three times

Canary `guide-can-i-overpay` first, graded in full before anything else fired: row
replaced (new id — proof the save ran, not a carry), 3,012 → 3,045 b, every number,
link and heading preserved, no form control or script introduced, and on the served
page the new opening present **1** with the baseline opening **0**. Then batches of
4/4/3.

**The rewrite is on-spec, not merely different.** It opens where the reader is standing
(*"If you're thinking about paying extra off your loan, the short answer is yes,
usually"*), explains before naming (*"you'll need to ask your lender for the total
amount required to close the account today. Lenders call this the Settlement Figure"* —
plain words first, the term second), and it quietly retired an absolute claim:
*"Overpaying is almost always a smart financial move"* → *"For most people, overpaying
is worth doing."*

**Every batch threw flags, and 3 of the 4 flag CLASSES were defects in my checker, not
in the copy.** Worth recording because each one would have read as "the framework
damaged the page" and each was disprovable in one query:

1. **"new opening NOT on the served page"** (3 of 4 pages in batch 2a) — my `opening()`
   stripped tags from the whole document first, which welds the `<h1>` to the following
   subtitle into a pseudo-sentence that exists nowhere in the served HTML, because tags
   sit between them. Fixed by taking the opening from inside a single `<p>`.
2. **Same message again on a later page, different cause** — I compared tag-stripped DB
   text against the RAW served HTML, so an inline `<strong>PCP (Personal Contract
   Purchase)</strong>` inside the first sentence broke the substring match. Fixed by
   stripping tags from the served side too. **Two distinct bugs behind one identical
   symptom**; fixing the first made the second look like a real failure.
3. **"FACTS LOST"** — the H voice tells the writer to speak numbers as a person would,
   so it turns `4-5 years` into *"four to five years"* and `£25k+` into *"£25,000 or
   more"*. A digit-only comparison called every one of those a lost fact: 5 of the 6
   fact flags. The figures were all present, spelled out. `still_present()` now accepts
   the spelled-out and k/comma variants and reports only what is absent under all of
   them.
4. **The one real flag class**: heading wording. Two pages had a heading reworded — one
   `h4`, and one **`h1` page title** (`Typical Fees & The Power of APR` → `Hidden Loan
   Fees & The Power of APR`). The brief says keep the heading *structure*, and the
   structure is intact (same count, levels, order), so these are notes rather than
   failures — but an h1 is the page's title and that is the owner's call, so it is
   surfaced separately and has gone to `README_where_we_are`.

> **The lesson is not "my regex was loose".** It is that **a checker built from the
> shape of the OLD artefact will convict the new one of being different**, and every
> one of these failures pointed at the framework rather than at me. The tell each time
> was that the *content* probe passed while the *identity* probe failed — chunks of the
> new body were plainly on the served page while "the opening" was not. Where a check
> and a spot-read disagree, suspect the check. Going in `WRONG_CALLS.md`.

**The checker can come out otherwise, and I proved it rather than assuming it.** Run
against `guide-jargon-buster` before it was rewritten, and against `index` after the
guides were done, it correctly FAILED every prose row with *"row id UNCHANGED — the
save did not run"*. A grader that has never failed on a known-unrewritten page is not
evidence of anything.

**One accepted flag, adjudicated by hand rather than by the tool:**
`guide-document-checklist` dropped `100%` from *"Lenders need to be 100% sure you are
who you say you are"* → *"Lenders need to be sure you're who you say you are"*. That
`100%` is a rhetorical intensifier, not a figure about the world, and the voice spec
explicitly bans advertising emphasis. Accepted as correct behaviour. Left in the
checker as a FAIL on purpose — the tool should not be taught to swallow a missing
number, because next time it may be a real one.

**Also observed, and it is an addition rather than a preservation:** the writer expanded
*"under the Consumer Credit Act"* to *"under the Consumer Credit Act 1974"*. Correct —
that is the right statute for a settlement figure — but the brief said preserve, not
improve. Recorded, not reverted; flagged to the owner as the kind of thing to watch.

### The calculator pages: the framework refused one page, and a rewrite deleted a page's CSS

Batch 3a (`tool-application-tracker`, `tool-car-finance-calculator`,
`tool-compare-loans`, `tool-credit-roadmap`). **The lock prediction held exactly**: all
three locked tool rows kept their row id AND their 2026-08-02 `updated_at`, and the two
expected `lock_blocked_change` notices were filed. No calculator was touched.

**Two things happened that were not predicted.**

**1. `validate_page_content` REFUSED `tool-car-finance-calculator`.**
1 blocker, `meta_commentary` on the string `input_schema` — *"…ymbol come from the
input_schema. FIXED 2026-08-03 — 0% APR, as its own change…"*.

> **CORRECTED, same session — I first wrote that "the writer emitted this into the
> copy… the model narrated its own task", and that was WRONG.** It is what the
> validator's own message asserts (*"the model wrote about its task instead of doing
> it"*) and I repeated it without testing it. The writer never sees that text: the
> `page-content-writer` prompt interpolates
> `{{.current_section.existing_content_html}}` and the content-direction fields, never
> the tool component's template. What falsified it was a predictor: **exactly 3 of the
> 12 tool pages have a tool template containing `input_schema`, and exactly those 3
> failed while all 9 others passed** — a control group that could have come out
> otherwise. The real cause is that `checkMetaCommentary` substring-scans the WHOLE
> assembled page HTML, comments included, and the locked calculator's template carries
> a developer changelog comment that has shipped to readers since 2026-08-03. Filed as
> **`bugs_open/219`**. The lesson: **an error message names a cause it has not
> established**, and a confident one will be adopted whole. It cost about an hour. The page was NOT saved: both prose rows still
carry their original ids and bytes, and the live page is untouched. Detail is in
`agent_error_log.context.issues`, not in `collected_data` — that is where
`validate_page_content` persists its issue list.

**2. A rewrite DELETED a page's CSS, and every guard in the system said yes.**
`tool-compare-loans`' `prose-0` was never prose: it held the `<style>` block for the
comparison layout. The rewrite replaced it with 2,124 b of (good) copy and took
`.comparison-wrapper`, `.loan-column`, `.stat-label` and `.stat-value` off the served
page — **while the markup still referenced two of them**.

Nothing in the platform was wrong. The component's `llm_guidance` promises *"no element
addressed by any script, so rewriting this prose cannot break a calculator"* — which is
TRUE and silent about CSS. The locked-row guard protects the tool row, not the style
row. The validator did not object. The arithmetic still computed; only the layout
collapsed.

**Census after finding it: 8 of this site's 51 `prose-*` rows carry a `<style>` block**
(`guide-jargon-buster`, `tool-application-tracker`, `tool-car-finance-calculator`,
`tool-compare-loans`, `tool-consolidation`, `tool-credit-health-check`,
`tool-loan-vs-savings`, `tool-overpayment-calculator`). Of the four already fired, the
writer **kept** the style block on three and **dropped** it on one. So this is a coin
flip, not a rule — and a single spot-check of one page would have cleared the whole
class wrongly. That is the entry now in `LANDMINES.md`.

**Repaired**: exact row restore of `content_data`/`rendered_html`/`content_hash` from
`page_components_bak_20260807_voiceh` inside a transaction with a `DO … RAISE
EXCEPTION` verify block (a bare `SELECT` cannot stop a COMMIT), then an assemble-only
rerender, then confirmed **at the served page**: all four selectors back, `prose-1`'s
rewrite retained, 4 `<script>` blocks intact. prose-0 is left un-rewritten, which is
the correct end state — there was never any prose in it.

**The grader now fails on any lost CSS selector**, per page, never sampled. Checking for
the literal `<style` tag would not have been enough: a rewrite can keep the tag and drop
the rules.

### Final tally, and the site-wide gates

**23 of 26 active pages now carry the H voice** (22 this session + the 08-06 canary).
The 3 that do not are blocked by `bugs_open/219` and are unchanged, not damaged.

| gate | result |
|---|---|
| all 26 active pages | HTTP 200, >5 KB, real `<!DOCTYPE>` — **0 failures** |
| 12 locked calculator rows | **12/12** identical in row id, `updated_at` AND `rendered_html` |
| `toolgolden --compare` vs `GOLDEN_2026-08-03b` | **all 11 tools reproduce their golden values exactly**, exit 0 |
| per-page grading | facts, links, heading structure and CSS preserved; new copy present on the served page with the baseline opening absent |

**The legal page was read by hand, not graded by md5**, as the plan promised. Every
disclaimer survives: not a lender/credit broker/adviser, consult a professional or the
lender, all four UK GDPR points (no server storage, local-only tracker data, we can't
see or sell it, essential cookies), and the FCA statement with its *because* clause
intact. The one substantive change makes the page **more** conservative, not less: *"we
strive for 100% mathematical accuracy"* became *"We aim for mathematical accuracy, but
every result here is still an estimate"* plus an explicit clause about a figure not
matching a real agreement. Its `<h1>` gained `&amp;` for a raw `&`, which is correct
HTML and renders identically (checked on the wire: `&amp;amp;` returns 0).

**Bookkeeping:** 22 items completed after grading (never before — a direct dispatch
bypasses the loop's `mark_complete`, so completing on grade is the honest order), 7
`lock_blocked_change` notices cancelled with the evidence that they were benign, and
3 items cancelled pointing at 219.

**Still owed / deliberately not done:**
- The 3 pages blocked by 219 (`index`, `tool-car-finance-calculator`,
  `tool-interest-rate-stress-test`). Re-fire once it is fixed; the dispatcher and
  grader are in this lane and take one page name.
- **Re-baseline the golden.** `GOLDEN_2026-08-03b` still matches, so nothing is broken,
  but the prose around 10 calculators has changed and a future capture will diff on
  `dom_shape`. Not re-baselined here because 3 pages are pending — a partial capture
  certifies nothing, which is why the harness refuses one.
- **The expansion question is the OWNER's, not mine** (see `README_where_we_are`).
  Several calculator pages had near-empty prose stubs (32–156 b) which the framework
  filled with 800–1,900 b of new explanatory copy. That is more than a voice rewrite.
  Nothing invented a number the claims gate would reject, but it is new substance on a
  finance site and it should be looked at rather than assumed acceptable.

## 2026-08-08 (evening) — first owner review of the homepage, and two surprises

**Owner verdict:** likes the homepage copy except the opening paragraph, which is too
strong — *"exact"*, *"mathematically rigorous"*, *"true cost of credit"*. His reasoning
is the useful part: *"these are good points but just too strong, it is positioning us as
the authority in accurate tools but that's not usually a top reader request or need —
they already trust our tools, and everyone else's."* Routed the positioning half to the
`vigilant_designer_offer_analysis` lane as `CONTRIB_2026-08-08_a_true_well_evidenced_claim_can_still_be_the_wrong_thing_to_lead_with.md`.

### Surprise 1 — the paragraph he reviewed is NOT ours

`index`'s prose rows were last written **2026-08-02 18:46:05**, which is the
decomposition, not the voice rewrite. index is one of the three pages `bugs_open/219`
blocked, so it still carries the ORIGINAL hand-built copy.

**Both his praise and his complaint are about writing the framework never touched.**
Worth stating plainly before anyone reads this thread as feedback on the rewrite —
and worth remembering when 219 ships, because rebuilding index will replace the parts
he said he liked along with the paragraph he didn't.

### Surprise 2 — a voice fix would NOT have prevented it, and the DEFAULT is softer

Re-ran the exact source block through the live writer twice against the production key
and model (`gemini-pro-latest`, key pulled from `content-creator-agent`), same neutral
brief both times:

| | opening claim it produced |
|---|---|
| live copy (never rewritten) | "mathematically rigorous" · "exact" · "true cost of credit" |
| **A — platform default house voice only** | "These tools calculate the cost of credit" — all three gone |
| **B — this site's H voice spec** | "the true cost of borrowing" — two gone, one survives |

**Both dropped "mathematically rigorous" and "exact" unprompted, and the plain default
was the SOFTER of the two.** The house voice's "match the word to the size of the fact"
rule does this work; H's rules are about warmth and walk-in openings and police claim
strength less hard.

So: the over-claiming is not a voice defect, and a voice fix will not reliably prevent
the class. What no voice can decide is whether the sentence is worth having at all.

> **MISSTEP, caught and corrected mid-experiment.** My first brief to the model literally
> said the site *"takes no applications and sends nothing anywhere"*, and the output came
> back with "We take no applications and send nothing anywhere" — a negation pile I then
> nearly wrote up as the default voice's failure. **I had planted it.** Re-ran both arms
> with the facts stated positively ("It is independent. The calculators run in the
> reader's own browser."), and the negation pile did not reappear in either. A prompt that
> hands the model a phrasing and then measures whether it uses that phrasing is not a
> measurement of the voice.

⚠ Second-order: run A truncated mid-sentence at `maxOutputTokens: 3000`. Cause was
**`thoughtsTokenCount: 2769`** — thinking ate 92% of the ceiling, leaving ~230 for visible
text. Re-ran at 8000, `finishReason: STOP`, 119 visible tokens. This is the trap
`scripts/gemini-probe.sh` exists to answer; budget for thinking separately from output on
any `gemini-pro-latest` call.

### The two rewrites drafted for the owner

Both drop all three strong claims, lead with the reader's job rather than our rigour, and
let specifics carry authority instead of asserting it. Neither is applied.

**Candidate 1 — reader-first**
> Working out what a loan really costs you is harder than it should be. The monthly
> payment is the number lenders lead with, and it's the one that tells you least.
>
> These calculators show you the rest of it: what a personal loan repays over its full
> term, whether consolidating your debts actually saves anything once the term stretches,
> and what the interest inside a car finance deal adds up to. They're free, and your
> figures stay on your own screen.
>
> The standard calculator is just below. The Tools and Guides menus have the rest.

**Candidate 2 — plainer, quietly confident**
> Free UK loan calculators, with guides that explain what the numbers mean.
>
> You can work out what a personal loan repays over its full term, test whether
> consolidating your debts saves you anything once the term is longer, and see what the
> interest inside a car finance agreement really comes to. Your figures stay in your
> browser.
>
> Start with the standard calculator below, or use the Tools and Guides menus.

⚠ **`index` cannot be rebuilt through the framework until `219` ships**, so applying
either means an SQL row write plus an assemble-only redeploy — the offline route — which
is what the owner's standing "through the framework, not this CLI" instruction rules out.
**Wait for 219 and seed the wording as guidance to the writer** rather than hand-writing
the row. Also note the site's `content_direction` is where a lasting fix goes: neither
candidate is protected from being written back over by the next rebuild unless the
positioning judgement is in the spec.

---

## 2026-08-08 (afternoon) — 219 fixed, and the fix candidate I filed that morning would have shipped inert

**Fixed:** `744bfdb3d`, council APPROVED round 1 (`c9104844-b303-43dd-a426-73386ebbb25e`),
awaiting the roll at `v1.0.1265`. Full account in `bugs_open/219`; only the parts that
belong to this lane's log are here.

### The wrong turn, which is the useful part

This morning's bug file says the developer note lives in an **HTML comment**, and its
leading fix candidate follows from that: *"strip HTML comments before scanning"*. Both
wrong. The notes are **JavaScript `/* … */` comments inside `<script>`**:

```
        function         | hit_pos | last_script_open | next_script_close | last_html_comment_open
 tool-car-finance-pcp-hp |    8048 |             5379 |             11603 |   (none)
 tool-loan-repayment     |    8224 |             5019 |              9568 |     3560  <- closes before the script opens
 tool-rate-stress-test   |    5692 |             3884 |              7303 |   (none)
```

A comment-stripper would have unblocked **none** of the three pages, passed review,
and left the bug closed and live. What caught it was not re-reading — it was running
the extraction as a query: pulling only the `<!--…-->` text gave **0** hits where the
whole-template scan gave **3**. Two counts that should agree, disagreeing. Logged in
`WRONG_CALLS.md`; the shape is *a fix candidate is an untested claim wearing the
clothes of a conclusion*, sitting in the same file, same register, as a diagnosis that
was genuinely rigorous.

### What shipped

`checkMetaCommentary` now scans `datahelpers.ExtractAssertionText(html)` +
`headProseBlocks(html)` — the same seam `218` settled for the placeholder scan this
morning, after a council REVISE that said *reuse it, do not write a second stripper*.
Scope was the defect, not comment syntax, so this puts script/style/code/pre/head/
attributes/comments out of reach in one move rather than the one that bit us.

### Proving it, without composing a fixture

The three failing runs still held their exact input — `collected_data ->
'page_content' -> 'response' ->> 'page_html'`, unpruned inside the ~24h window. Ran
the real function over the real bytes:

| orchestration | page | bytes | old (production) | new (local) |
|---|---|---|---|---|
| `fbd3da9d…` | index | 14,436 | `1 blockers, 0 errors` | 0 issues |
| `0752258f…` | tool-interest-rate-stress-test | 10,475 | `1 blockers, 0 errors` | 0 issues |
| `01072bcf…` | tool-car-finance-calculator | 14,147 | `1 blockers, 0 errors` | 0 issues |

The old-code half is production's own error text, not a local replica of the old
function — a stronger control than anything I could have written, and free. **Worth
keeping as a habit: for a validator fix, the artefact that failed is usually still in
`collected_data` for a day.** The test also asserts each artefact still *contains* the
blocking string, so a truncated or wrong file cannot masquerade as a pass.

### Two things found while measuring, neither of which I went looking for

1. **`bugs_open/221`** — webdesign.co.uk's `tools-index` is blocked by the same check
   on *"LocalBusiness schema, as an AI-builder prompt"* (`as an ai` in genuinely
   visible copy). My fix does not help it; verified by running the new check over that
   page's real HTML. Filed separately and the lane has been told in their NOTES,
   because measuring that nothing breaks is not the same as telling the people whose
   page it is.
2. **The census that found it is the one that nearly did not run.** 1,244
   `page_components` rows, positive control 268. Had I only checked "does my change
   break anything on loancalculator", both the second instance and the JSON-LD answer
   below would have stayed invisible.

### The council objection, and why it was worth answering rather than noting

`bug_historian` (medium) + `guidelines`: `ExtractAssertionText` excludes JSON-LD, this
file already has a landmine about the banned-claims sweep missing JSON-LD, and I had
measured `<title>`/meta but **not** JSON-LD — so I was asking a human to sign off a
blind spot. Correct hit. Answered:

- 0 of the 37 assembled pages in `collected_data` contain `application/ld+json` at all
  (control: 19 of 37 contain `<script>`);
- because JSON-LD is appended to `<head>` at **render** time, after validation
  (`rerender_single_page_action.go:931`, `data_helpers.go:1505`) — corroborated by a
  dated note already in that file: *"measured 2026-07-28, ZERO of 14 live sites emitted
  any application/ld+json"*;
- and of the 27 `ld+json` blocks that do exist in stored component HTML, **0** carry
  any meta-commentary vocabulary (control: 25 match `schema.org`).

So there was no JSON-LD coverage to lose. **The seat was right that I had not
measured it** — the answer happening to be favourable does not retire the point.

### Still owed on this lane

Unchanged from the handoff except that the blocker is now a roll, not a bug:

- **The 3 pages**, once `v1.0.1265` is live. The recipe is unchanged: `voiceh_batch.sh`
  then `voiceh_grade.py`. ⚠ Prove the grader can still FAIL first (run it against a
  page that has not been rewritten) — a grader that has never failed is not evidence.
- **Re-baseline the golden**, after those 3 land and not before.
- **The expansion question is still the owner's.**

### CORRECTION (2026-08-08, same evening) — both candidates above contain bad English, owner-caught

The owner picked candidate 2 on tone and then found two faults in it. **Both faults are
in candidate 1 as well** — I wrote the same phrases into each and did not re-read them.

1. **"what a personal loan repays over its full term"** — a loan does not repay
   anything; the borrower does. The subject is simply wrong. It scans as fluent and
   means nothing precise.
2. **"once the term is longer" / "once the term stretches"** — longer than *what*? It
   silently compares against the terms of the debts the reader already holds, and never
   says so. The whole point of the consolidation trap is that comparison, and the
   sentence omits it.

> **The transferable bit, and it is uncomfortably on-topic.** I compressed for tone and
> broke the sense, and it read smoothly enough that I committed it and presented it. This
> is the same failure this lane has been documenting all week from the other side:
> **fluency covering for imprecision.** No rule in any voice spec would have caught it —
> it violates none of them. It needed a reader who knew what a loan actually does.

**Corrected candidates. Faults fixed, tone preserved, and the sentence openings varied
deliberately — my drafts had three consecutive "You can…" openings, which is precisely
the sameness the owner objected to in the first place.**

**2a**
> Free UK loan calculators, with guides that explain what the numbers mean.
>
> You can see what a personal loan costs in total, not just each month. If you're
> thinking about putting several debts into one, you can check whether that saves you
> money or just spreads the same cost over more years. There's one for car finance too,
> which shows how much of the deal is interest. Your figures stay in your browser.
>
> Start with the standard calculator below, or use the Tools and Guides menus.

**2b** — plainer, and the car finance line carries more
> Free UK loan calculators, with guides that explain what the numbers mean.
>
> Work out what a personal loan will cost you in total, not just each month. If you're
> considering consolidation, you can compare your current debts against a single loan and
> see whether it really saves anything. The car finance tool shows how much of a deal is
> interest rather than car. Your figures stay in your browser.
>
> Start with the standard calculator below, or use the Tools and Guides menus.

Changes: *"what a personal loan repays"* → *"what a personal loan costs you"* (cost sits
on the reader, where it belongs); *"once the term is longer"* → *"over more years"* with
the comparison named (one loan against the debts you hold now).

⚠ **Candidates 1 and 2 as first drafted are superseded — do not apply them.** They are
left above rather than edited away because the fault is the record.

### FINAL homepage opening (owner-approved shape, 2026-08-08) — and the spec rule that will fight it

Owner chose **2b** and cut the privacy line: *"neither needs the 'Your figures stay in
your browser' point here. It suddenly moves the reader from thinking about loans to
suddenly being technical."* Right, and the middle paragraph now stays on the reader's
money question from first word to last.

**FINAL:**
> Free UK loan calculators, with guides that explain what the numbers mean.
>
> Work out what a personal loan will cost you in total, not just each month. If you're
> considering consolidation, you can compare your current debts against a single loan and
> see whether it really saves anything. The car finance tool shows how much of a deal is
> interest rather than car.
>
> Start with the standard calculator below, or use the Tools and Guides menus.

⚠ **THE SPEC WILL PUT THE PRIVACY LINE BACK.** I did not invent that sentence — this
site's seeded `content_direction` contains, verbatim:

> "State facts positively, including privacy and cost: 'free', **'your numbers stay on
> your own screen'** — never a negation pile ('no sign-up, no credit check, nothing sent
> anywhere')."

So every framework rebuild of this page will reassert it. **Fixing the copy without
fixing the rule buys one render.**

The rule is half right and half wrong, and the halves are worth separating:

- **Right:** if privacy is mentioned, state it positively. The negation pile it forbids
  ("no sign-up, no credit check, nothing sent anywhere") is genuinely worse.
- **Wrong:** it reads as licence for the fact to appear anywhere, and supplies the exact
  wording, which is why it landed in an opening about money. **A rule about HOW to phrase
  something silently authorises WHETHER to include it.**

**Proposed amendment** (not applied — it is a change to the seeded spec and belongs with
the owner's positioning work): make placement part of the rule. Privacy answers a
question the reader has when they are **about to type their own figures into a box** —
so it belongs beside the calculator inputs or in the footer, not in an opening whose job
is to say what the site is for. Same shape as the `mathematically rigorous` finding: a
true, defensible, positively-stated fact, put where nobody asked for it.

**Third instance this session of one pattern**, which is now worth naming: over-claiming
accuracy, my own "what a personal loan repays", and now the privacy line. None of the
three violates any rule we have. **Two of the three were CAUSED by a rule** — the spec
told the writer to state privacy this way, and the H register's warmth rules kept "true
cost of borrowing" where the plain default dropped it. Rules here are not merely failing
to catch faults; they are generating them.

---

## 2026-08-08 (evening) — the rollout is COMPLETE: 26/26, and the golden re-baselined

`v1.0.1266` shipped the `219` fix. Everything below was done against that image.

### Pod verification, with a control that could have failed

Not "a roll happened". Both `agent-chassis` replicas: the string the fix ADDED = 1, the
string it REMOVED = 0. Then a pre-fix pod still on `v1.0.1264`
(`agent-page-rebuild-a104b2fc-tzk95`): ADDED = 0, REMOVED = 1 — the exact inverse, so
neither grep is one that always passes. ⚠ **The fleet was mixed at this point** (20
pods on 1264, 5 on 1266): the long-lived `agent-chassis` deployment had rolled, the
spawned agent pods had not. Check the pod that will actually run your action, not "the
fleet".

### The three pages

Fired all three (owner ruling: `index` goes through the framework like everything else,
no special handling). All reached `current_step = complete`; this morning the same
three ended `complete_error` with `1 blockers, 0 errors`.

`voiceh_grade.py` **3/3 PASS**:
```
[PASS] tool-car-finance-calculator     prose-0 683->2328b, prose-1 174->638b
[PASS] tool-interest-rate-stress-test  prose-0  34->2120b, prose-1 131->850b
[PASS] index                           prose-0 1101->1126b, prose-1 133->2210b,
                                       prose-2 117->295b,  prose-4 2803->2813b
```

**And the grader was proven able to fail, this session, not by reputation.** Mutated
the baseline so `index`'s recorded row ids equal its CURRENT ids, re-ran: `0/1 pass`,
4/4 rows flagged *"row id UNCHANGED — the save did not run"*. That is the mutation
test, not a spot-check: it proves the check can fire, which a pass alone never does.

### The calculators

`toolgolden.py --compare GOLDEN_2026-08-03b`: **all 11 tools reproduce their golden
values exactly** — including the two whose pages had just been rebuilt. Ran the compare
BEFORE re-baselining, deliberately: capturing first would have silently blessed
whatever the rebuild did, and the compare is the only step that could have said no.

Re-baselined as `acceptance/GOLDEN_2026-08-08_voice_h_complete.json` — 11 pages, and
now 4 vectors, since the old golden predates `asym` (added 2026-08-05 for the
ratio-calculator blind spot; it compared on 3 and said so each time).

### Work items completed AFTER grading

The three items were `detected` until the grade passed, then set `complete` with the
verdict and the falsifiability note in `result`. A direct dispatch bypasses the loop's
`mark_complete`, so `detected` honestly meant "built, not yet graded".

### A near-miss worth recording: my own filter nearly invented a missing page

Counting coverage with `max(prose.updated_at) >= '2026-08-07'` returned **25 of 26**,
and I was one keystroke from reporting a page the rollout had missed. It had not.
`guide-how-loans-are-calculated` is the **voice-H canary** (item `2517bc4b`, the page
the owner reviewed), rewritten `2026-08-06 11:53` — the day BEFORE the batches. My
filter date drew the line after the canary and the canary fell outside it.

The lane's own baseline should have tipped me off and didn't: the 08-07 backup already
contained that page's voice-H text. **The date in a coverage filter is a claim about
when the work happened, and it is exactly as fallible as any other claim** — here it
encoded "the rollout is the batches", which was false. Same family as the entries in
`MEMORY.md` about a filter defining its own conclusion; logged here because it fired on
a *completion* claim, which is the kind most likely to be repeated to the owner.

### Final state

26/26 active pages in voice H. 26/26 HTTP 200 (swept live, not inferred). 11/11 tools
identical to golden. `219` fixed, live, proven, and kept in `bugs_open/` per owner
direction with the evidence inside the file. `221` open and belongs to webdesign.co.uk.

## 2026-08-08 (late) — the rule-trim trial: mostly FAILED, and the failure is the useful part

**Owner question:** *"it's all too mechanical, either we extend the rules further or
remove some rules?"* Answer given: remove, and specifically remove **prescriptions**,
keeping prohibitions. The argument — a prohibition applied to 100 sections still permits
100 different openings; a prescription applied to 100 sections produces 100 openings of
the same shape, which is a template wearing a rule's clothes.

**The trial:** `trim_voice_rules.py` (gated on the same formatter port as
`seed_voice_h.py`; backup `site_specs_bak_20260808_ruletrim`; `--revert` provided).
23 rules → 22:

| rule | edit |
|---|---|
| "Open sections where the reader is standing… conditional or situational clause" | demoted to its prohibition half |
| "Lead every guide section with the practical bottom line…" | DELETED — duplicate mandate |
| "State facts positively, including privacy… 'your numbers stay on your own screen'" | demoted to "never a negation pile" |
| "Paragraphs are 2-4 sentences maximum" | relaxed to "vary their length" |

Then rebuilt three pages through the framework: `guide-debt-help-uk`, `tool-consolidation`,
`legal`. All 3 orchestrations COMPLETED, all 5 prose row ids changed (so the saves ran),
locked calculator rows untouched, CSS survived.

### RESULT — the tic SURVIVED, and I know why

`legal.html` still opens *"If you're using the calculators and guides…"*.
`debt-help-uk` still opens *"If you've missed a payment and you're starting to panic…"*.

**Because I removed the RULE and left the EXEMPLARS.** The spec's `voice_exemplars` —
which I wrote on 08-05 — demonstrate the pattern in three of four worked examples:

```
before_after_1_consolidation  AFTER: "If you're thinking about rolling several debts…"
before_after_2_car_finance    AFTER: "If you're looking at car finance…"
before_after_3_overpayment    AFTER: "If you have spare money each month…"
```

**And the VOICE doc says exactly why that wins:** *"A writer model follows exemplars more
reliably than rules — the rules explain the register, the pairs teach it."* So I deleted
the weaker teacher and left the stronger one standing. **A rule and its worked example are
not two statements of one instruction; the example is the instruction, and the rule is
commentary.** Any future trim must trim BOTH or it is theatre.

### And it introduced a NEW defect, from the rule I relaxed

`tool-consolidation` prose-0 went **1060 → 3306 bytes**. prose-0 is a **CSS-only row** (the
`<style>` block; 8 of 51 prose rows on this site are). Handed a row with no prose to edit,
the writer **invented a whole topic section** — and, unable to see prose-1, wrote what
prose-1 already says. The served page then carried *"the appeal is usually the monthly
figure"* and *"the appeal is usually the monthly payment"*, two disclaimers and two
"Need help?" blocks.

> **This is the cross-section collision I predicted on 08-06 and then RETRACTED when
> measurement found no repetition.** The retraction was correct for the copy as it stood:
> with the 2–4 sentence cap in force, no section had room to wander into a neighbour's
> territory. **The cap was suppressing the symptom of a defect I had concluded was absent.**
> Relaxing it released the symptom immediately, on the first page with a CSS row. So: the
> mechanism was real all along, the measurement was honest, and the conclusion drawn from
> it — "no repetition, therefore not a problem" — was wrong because it measured a system
> whose guard was still on.

Restored with `voiceh_restore_css_slot.sh tool-consolidation prose-0`.

### What the owner's three specific faults did

- **`debt-help-uk` order** — unchanged, as expected: not attempted in this trial. It is a
  **journey** judgement, not a copy one. Nearest existing home is `experience-planner`
  (its vocabulary is literally journeys and journey criteria). `visual-designer` could
  never have caught it — that agent "handles images, logos, and visual assets".
- **`tool-consolidation` "appeal"** — NOT fixed, and briefly doubled. Nothing in the spec
  addresses it, so nothing was going to.
- **`legal` "if"s** — NOT fixed, per the exemplar finding above.

### Standing conclusion, updated

The answer to "extend or remove" is still **remove** — but removal has to reach the
exemplars, and length rules turn out to be load-bearing in a way nobody intended: they are
the only thing stopping independently-written sections from colliding, because no writer
sees its siblings. Remove them and you need the page-level view first, not after.

### The exemplar fix ALSO failed — and the third layer is where the instruction actually lives

`fix_voice_exemplars.py`: five exemplars replacing four, five different opening shapes,
one still a conditional so the opposite tic is not taught instead. Also removed the
privacy line (still sitting in `register_anchor`, a second instance of the same oversight)
and the "appeal" wording the owner flagged. Gated and applied; `formatted` 24,361 →
24,779 b, conditional openings demonstrated in AFTER examples 3 → 0.

Rebuilt the same three pages. **The tic survived again**, including on
`tool-consolidation`, whose exemplar I had specifically rewritten to open with the plain
fact. So the exemplar hypothesis is refuted too.

**Both interventions were inert, and here is why.** `voiceh_rewrite.sh` copies the
rewrite guidance **by SQL from the canary work item the owner approved on 2026-08-06**,
and that guidance says, in the item's own `spec.suggestion`:

> "Apply the register: **open each section where the reader is standing (a conditional or
> situational clause such as 'If you are paying off a loan…')** rather than opening cold
> with the assertion…"

An explicit mandate **carrying its own inline example**, arriving as the task's own
instruction — the signal closest to the work and the strongest of the three. Editing the
spec's rules or its exemplars could never have reached it.

> **THREE LAYERS CARRY THIS VOICE AND ONLY ONE OF THEM WAS DRIVING.**
> 1. `site_specs.content_direction.writing_rules` — trimmed 08-08, no effect
> 2. `site_specs.content_direction.voice_exemplars` — fixed 08-08, no effect
> 3. **`site_work_items.spec.suggestion`** — the per-item rewrite guidance, pinned to a
>    canary. This is the one.
>
> **The pin is a deliberate safety property that became the blocker.** The script's own
> header explains it: the prompt is "COPIED BY SQL from the original canary item, never
> retyped, so it cannot drift from the spec the owner reviewed." That is good practice and
> it is exactly why two spec fixes did nothing — the reviewed prompt was frozen, including
> the instruction the owner had since rejected. **A prompt pinned for reproducibility stops
> tracking the decisions made after it was pinned.**
>
> **The general check:** before editing a prompt layer, establish which layer the model is
> actually reading. Diff what you changed against what the run received. I changed the spec
> twice without once looking at the work item's own `suggestion`, and the run text was one
> query away.

**Guidance v2** (`4a9edd45-…`, item_key `voiceh-guidance-v2-source`,
`voiceh_rewrite_v2.sh`): the conditional mandate replaced by its prohibition half plus an
explicit instruction to VARY, matching the spec change. ⚠ **This departs from the prompt
the owner reviewed on 08-06** — deliberately, because the clause removed is the one he has
since objected to twice. Old guidance kept at `2517bc4b-…` for comparison.

⚠ **Second-order trap hit on the way:** re-firing failed with
`duplicate key value violates unique constraint "idx_swi_dedup"`. The previous run's items
sit at `detected`, which is NOT in the index's excluded-status list
(`complete/verified/rejected/wont_fix/failed/unresolved/cancelled`), so the key blocks a
re-fire. That `detected` state is deliberate — the tool leaves items ungraded on purpose —
so **grading a batch is now a precondition for re-running it**, not just good manners.
Closed the three with their findings recorded in `error`.

### Guidance v2 WORKED — third intervention, first one aimed at the layer that was driving

`voiceh_rewrite_v2.sh` (guidance `4a9edd45-…`, conditional mandate replaced by its
prohibition half plus an explicit instruction to VARY). Same three pages, same everything
else. Measured on the served pages, blocks opening with a conditional:

```
legal.html            0 of 4    <- was "If you're using the calculators and guides…"
                                   now "The calculators and guides on loancalculator.co.uk
                                   are built for illustration and learning."
tools/consolidation   1 of 10
guides/debt-help-uk   2 of 10   <- and the one it kept is the right one:
                                   "If you've missed a payment and you're starting to panic"
```

**The tic is broken as a universal pattern.** It now appears where the reader's situation
genuinely is the subject and not where it isn't — which was the whole point. The legal
page, the clearest case of it being wrong, opens flatly. **Two spec-layer fixes: no
effect. One guidance-layer fix: immediate effect.** That ratio is the finding.

⚠ **The CSS trap fired AGAIN on `tool-consolidation`, and worse than last time.** prose-0
came back with `<style>` entirely **absent** (not appended-to, deleted) — served CSS class
matches 8 → 5. Restored from backup; verified 8 again, 35,206 b, no duplication, and
**toolgolden 11/11 exact**, 26/26 pages 200, zero locked rows touched.

> **This row has now been destroyed by two consecutive rewrites, and the guard is a
> RESTORE SCRIPT, not a preventer.** The lane's handoff says the rewrite "drops it about
> half the time"; today it was two for two. Every rewrite of this site pays a coin-flip on
> 8 of its 51 prose rows, and the only thing standing between that and a broken calculator
> layout is a human remembering to check. **The real fix is to stop handing a CSS-only row
> to a prose writer at all** — the row is identifiable (`content_data.content` starts with
> `<style>` and holds no sentence), so it can be excluded from the rewrite set rather than
> repaired after. Filed as the next thing to do in this lane.

**Where the three owner faults now stand:**

| fault | status |
|---|---|
| `legal` "if"s | **FIXED** — opens flatly |
| `tool-consolidation` "appeal" | **FIXED** — the word is gone from the live page |
| `debt-help-uk` ordering (experts first) | **NOT DONE** — a journey judgement, not a copy one; `experience-planner` is its home |

---

## 2026-08-08 (late) — OWNER RULING on the expansion question: keep the copy. Lane complete.

*"keep the explanatory copy."* No trim, no re-run, no work. The pages stand as the
framework built them, and with §1 and §4 closed earlier the same evening this lane is
**finished**.

Recorded here as well as in the handoff and `README_where_we_are` because the ruling is
worth more than its one-line content:

- **The framework was allowed to exceed its brief, and that is the finding.** The brief
  said "voice only, preserve every fact, add nothing"; the writer turned 32–156 byte
  stubs into 800–1,900 bytes of explanation. The instinct — mine — was to treat
  exceeding the instruction as a defect to be reverted. The owner's judgement is that
  the output is the thing being evaluated, which is the same reasoning that removed the
  `index` caution earlier today. **Two rulings, one principle**, and worth carrying into
  the next lane: on this estate, "the framework did more than asked" is a result to
  look at, not automatically a regression to undo.
- **The question was still right to ask and not to resolve unilaterally.** It was new
  substance on a finance site. Both readings were defensible, which is exactly the
  shape that belongs with the owner rather than with a thread's default.
- **Scope of the ruling is stated in the handoff, not assumed silently.** I read it as
  also covering the two smaller items filed underneath it (the `Consumer Credit Act
  1974` expansion and the two reworded headings) since they are the same act by the
  same writer, asked in the same question. That reading is written down where it can be
  contradicted rather than left implicit — if it is wrong it costs a two-heading trim.
- **The backup table's purpose has changed.** `page_components_bak_20260807_voiceh` was
  being held partly as the undo for a "trim" answer. There is nothing to undo now, but
  it stays: §3's CSS-in-a-prose-slot trap is what it really protects against, and
  `voiceh_restore_css_slot.sh` reads it.

Final lane state: 26/26 pages in voice H · 26/26 HTTP 200 · 11/11 calculators identical
to golden, re-baselined · `219` fixed, live, proven, kept open per owner direction ·
`221` open and owned by webdesign.co.uk · nothing owed here.

### The CSS trap turned into a PREVENTER — and it needed two different fixes, not one

Owner: *"exclude the `<style>` no sentence set from the rewrite set."* Measuring the set
first turned a one-line job into a two-part one.

**8 rows carry a `<style>` block, and they are not one population.** Prose characters
remaining after stripping style blocks and tags:

```
tool-compare-loans            prose-0    20   <- pure carrier
tool-credit-health-check      prose-0    20   <- pure carrier
tool-overpayment-calculator   prose-0    26   <- pure carrier
tool-consolidation            prose-0    32   <- pure carrier
tool-application-tracker      prose-0   170   <- MIXED: real prose + a style block
tool-car-finance-calculator   prose-0  1523   <- MIXED
tool-loan-vs-savings          prose-0  2293   <- MIXED
guide-jargon-buster           prose-0  2637   <- MIXED
```

A clean gap at 32→170. **Locking all eight would have frozen four rows of real copy**,
including `application-tracker`'s (which, incidentally, is a negation pile — *"There's no
account to create, and nothing gets sent to us or anyone else"* — so it is a row that
particularly needs rewriting, not protecting).

**Part 1 — lock the 4 pure carriers.** `locked_by='loancalculator_css_carrier_20260808'`,
`lock_type='permanent'`. Reuses the existing lock mechanism rather than inventing a
filter: these rows are layout in a prose slot, and the platform already has one honest
way to say "the writer may not have this". Site total 12 → 16 locked. Backup
`page_components_bak_20260808_csslock` (63 rows).

⚠ **Checked before locking, because `bugs_open/189` duplicates locked rows on the build
path:** 189's config half **IS applied** — `slot_name_from: current_section.name` is set on
BOTH `render_section` and `render_from_template`. My memory entry said UNAPPLIED and was
stale. Verified from live config, then again empirically below.

**Part 2 — guidance v3 for the mixed rows,** which must NOT be locked. Adds: *"if the
existing content contains a `<style>` block, reproduce that entire block byte-for-byte in
your output, unchanged and in the same position. It is the page's layout, not prose."*

**Both halves PROVEN on one run** (`tool-consolidation` = locked carrier,
`tool-loan-vs-savings` = mixed row):

```
tool-consolidation   prose-0  1060 b  style=YES  locked=YES  updated 18:00:28
                                                             ^ the RESTORE time, not the
                                                               rewrite's 18:14:37 — the
                                                               writer never touched it
tool-loan-vs-savings prose-0  3071 b  style=YES  locked=no   updated 18:12:46
                                                             ^ rewritten AND kept its CSS
```

Row counts 4 and 4 — **no 189 duplication**. Served: both pages carry 5 `<style>` blocks,
26/26 pages 200, **toolgolden 11/11 exact**.

> **The distinction worth keeping.** A lock says "this is not yours to write". A guidance
> clause says "write this, but carry that part through untouched". They are different
> instruments and the earlier restore-script approach was neither — it let the damage
> happen and then undid it, which only works while a human is watching. Deciding which
> instrument applies needed the population measured, not assumed: the handoff said "8 of
> 51 prose rows hold the page's `<style>` block", which is true and would have produced
> the wrong fix for half of them.

### The journey run: the instrument produced a confident plan for the WRONG SITE — `bugs_open/227`

Owner: *"can you run the journey to judge that ordering choice?"* Fired
`092_TRIGGER_experience_plan.sh loancalculator.co.uk debt-difficulty-help "getting help
when you cannot keep up with a loan repayment"` (corr `a30b0c5b-…`). Preconditions
checked first: `66d32477d` is an ancestor of HEAD and the image is built from today's HEAD
(pod-grep is impossible for that fix — it added a CONDITION, no production string literal,
so ancestry is the honest check available); planner active with 5 review seats; pods 2h
past restart.

**It returned a detailed, well-structured EXPERIENCE_PLAN about vonc.com** — four journeys
over `/provocations/index.html`, `/tools/arena/index.html`, `/tools/gauntlet/index.html`
and lobby cards. loancalculator has **0** such pages; vonc.com has 6.

**Inputs and context were both CORRECT.** The run's own `input_data` carried
loancalculator's site_id and my experience name; its `load_context` output contains
loancalculator's pages (`guide-hidden-loan-fees`) and **zero** occurrences of
`provocation`. Searching every `collected_data` key for where vonc entered: `compose`,
`reframe`, `proposal`, `review_mvp`, **`agent_config`**, `review_contracts`,
`review_feasibility` — and `agent_config` is the agent's own definition.

**Root cause: the shared `experience-planner` prompt hardcodes vonc's diagnosis as the
task** — *"## The diagnosis you are fixing (three broken surfaces, artifact-verified
2026-07-17) 1. /provocations/index.html …"* — immediately after correctly interpolating
`{{.experience_domain}}`. `provocation` ×24, `gauntlet` ×8, `arena` ×5 in the live config.

> **This is the week's lesson at platform scale.** Three times in this lane a worked
> example has beaten the instruction around it: the spec rules, the spec exemplars, and
> the pinned per-item guidance. Here a worked example beats **the actual data**. The model
> held loancalculator's real page list and wrote another site's plan, because the prompt
> said *"the diagnosis you are fixing"* and named vonc's surfaces in the imperative.

**Latent since 2026-07-18, and the reason is clean:** all-history, `doc_plans` holds 61
experience plans — **59 `vonc-spark-game`, 2 mine.** `debt-difficulty-help` is the first
non-vonc experience ever planned, so the agent had only ever run on the site it is
hardcoded for. Contamination on the only non-vonc subject: 2 of 2.

**What worked:** the council **rejected** it — 5 reviewers, `decided_by: "veto from
feasibility"`, run ended `complete_refused`. A feasibility seat reading journeys over
pages that do not exist correctly refused. The review layer is not the defect.

⚠ **Second, separable defect:** the vetoed plan was persisted `is_current=true` at
18:25:45 and the run refused at 18:25:53, so **a council-rejected plan became the plan of
record**. Demoted by hand; that is cleanup, not a fix. Both filed as `bugs_open/227`.

**And the owner's actual question is still unanswered.** The instrument for judging the
`debt-help-uk` ordering produced nothing usable about it. Options: fix 227 and re-run, or
judge the ordering directly — it is not a large question (*should a page for someone who
has just missed a payment lead with free expert charities rather than with negotiation
tactics?*), and the honest answer may be that it needed a person all along.

### debt-help-uk reordered — free expert advice FIRST. Owner ruling, done directly.

**Owner 2026-08-08: *"it didn't need the instrument and we can fix 227 later."*** So the
ordering judgement was made by him and executed as a content change, not discovered by the
experience loop. `bugs_open/227` stays filed and unworked.

Route: a one-off `content_rewrite` through the framework with **bespoke guidance**, not the
voice guidance — this is a structure change, and saying so in the prompt is what made it
work. The guidance stated the REASON, not just the order: *"this page is read by someone
who has just missed a payment, and this site is a calculator site, not a debt adviser."*
Source item `7933edd4-…`, run `9e94084f-…`, row `18cd4ec7` → `40c67de1` (3,103 → 3,428 b).

**Result on the served page:**

```
1. Where to get free, expert advice      <- StepChange, National Debtline, Citizens Advice
2. Talk to your lender immediately
3. The "Breathing Space" scheme
   A note on your credit score
```

**Facts preserved, all nine checked on the wire:** `stepchange.org`,
`nationaldebtline.org`, `citizensadvice.org.uk`, `60 days`, `Financial Conduct Authority`,
`Debt Respite Scheme`, `Reduced Payment Plan`, `Loan Extension`, `CCJ`. Opening paragraph
about legal rights kept in place as instructed. 26/26 pages 200, **toolgolden 11/11 exact**.

> **The writer did something better than it was told, and it is worth recording.** The
> guidance said Breathing Space now "follows naturally" from section 1. The writer instead
> wrote a bridge back to the lender section: *"That first option, a Breathing Space, also
> exists as a formal legal protection, not just something your lender might offer as a
> favour."* That sentence did not exist before and no instruction asked for it. It resolves
> a genuine ambiguity the original had — Breathing Space appears both as a lender
> concession and as a statutory scheme — which the reorder would otherwise have made worse
> by separating the two mentions further. **Given a reason rather than a rule, it made an
> editorial judgement.** That is the first thing all week that has looked like writing
> rather than compliance, and the difference in the prompt was stating WHY.

## 2026-08-08 (night) — 219 shipped, the last 3 pages rebuilt, ROLLOUT COMPLETE at 26/26

**Chassis v1.0.1269 (22:02Z) carries `bugs_open/219`'s fix**, so `index`,
`tool-car-finance-calculator` and `tool-interest-rate-stress-test` are unblocked. All
three rebuilt tonight. **No page on this site is left in the old voice** — prose last
written 08-08 on 25 pages, 08-06 on the remaining one (the 204 canary).

> **⚠ MY 219 PROBE WAS BLIND, TWICE, AND I REPORTED "NOT LIVE" ON IT.** I pod-grepped for
> `"does not name the matched pattern"` — a literal that lives in
> `validate_page_content_meta_scope_test.go`, **not in production code**. A test string is
> never in the binary, so the probe returned 0 whatever the truth was. `744bfdb3d` added
> **no unique production string literal**: its two candidate symbols both fail to
> discriminate — `ExtractAssertionText` predates it (2026-07-16) and `headProseBlocks` was
> added the same day by a *sibling* commit (`35889819c`). **Symbol archaeology could not
> answer this question at all.**
>
> **What settled it was firing one rebuild at a blocked page and watching it succeed.**
> When a fix adds no new string, the empirical test is not a fallback — it is the only
> instrument. Check that a probe string is in a non-test file *before* trusting its zero;
> `git show <sha> -- '<production file>'` rather than `git show <sha>`.

### The homepage now carries the approved opening — and my guidance duplicated it 3×

Fired with bespoke guidance carrying the agreed copy verbatim plus the reasoning (why
accuracy is the wrong thing to lead with; no privacy sentence; British English).

**It worked and it over-applied.** `prose-0`, `prose-1` AND `prose-2` each came back
containing the new opening, because my guidance said *"REPLACE the opening block"* — a
**page-level instruction delivered to a per-section prompt**, so every section decided it
was the opening block. `prose-1` (the Standard Calculator intro) and `prose-2` were
overwritten with it. Restored both exactly from backup; opening now appears once.

> **Third instance this week of the same root shape, and the sharpest.** The writer sees
> one section and never its siblings. A rule, an exemplar, a pinned guidance and now a
> *page-level instruction* have each been applied uniformly because uniform application is
> all a section-scoped prompt can do. **Guidance must be written in the second person to
> ONE section** — "if this section is the page opener, use this copy" — never "replace the
> opening block", which every section can believe it satisfies.

**Live now, verified on the wire:** opening appears once, `mathematically rigorous` = 0,
`true cost of credit` = 0, 26/26 pages 200, **toolgolden 11/11 exact**.

⚠ **RESIDUE, precisely located and NOT fixed:** `index` `prose-2` still reads *"Calculate
your exact monthly repayments and see the true total cost of borrowing."* It is unlocked
`ported-prose`, and it is there **because my restore brought the old copy back with it**.
The owner's named paragraph is fixed; the same register survives one line below it. One
targeted rewrite closes it — and per the lesson above, that guidance must address the
section, not the page.

### Guidance lineage (each pinned as a work item; the tool copies it by SQL, never retyped)

```
2517bc4b  canary, owner-reviewed 08-06   mandated the conditional opening  <- caused the tic
4a9edd45  v2                             mandate -> prohibition + "vary"   <- broke the tic
6d52beaf  v3                             + preserve <style> byte-for-byte  <- current
7933edd4  one-off                        debt-help section reorder
50c8ba5c  one-off                        index opening block
```

---

## 2026-08-08, late — picking up `bugs_open/227`: the fix is written and dry-run proven, not applied

The site is finished; this is the platform bug the lane filed on its way out. Owner asked
for it to be picked up and handed to a fresh session, so: design, SQL, dry run, and a
handoff — `HANDOFF_2026-08-09_continue_here.md`.

### What the bug file got wrong, and how much of it I got wrong first

**The bug file names `compose`. It is five prompts.** Case-insensitive census of
`provocation|gauntlet|arena|vonc|spark` over the live row `e0194bee-…`:

```
compose             41
review_feasibility   2   (veto)
review_honesty       2   (HARD veto)
review_mvp           2
reframe              1   (rewrites a vetoed plan)
                    48
```

**My own first census was case-sensitive and returned 37 hits across three steps** — I
told the owner "three prompts, not one" and drafted the fix, and its md5 drift guard,
against three. `reframe` and `review_honesty` hide because "Gauntlet" is capitalised in
both. Logged in `WRONG_CALLS.md` as a recurrence of the oldest rule in the file. The
verify block in 345 now asserts `!~*`, not `NOT ILIKE` — my first draft of the *check*
had the same blindness as the measurement it was checking.

### The finding that changed the fix

**Three of the four council seats hold vonc's criteria as their general judging rule.**
`review_feasibility` asks whether data is "in /data/provocations.json or
client-computable" and watches for "the daily emitter"; `review_mvp` is told the core
loop is "land on a provocation → file a position → …; enter a real timed Gauntlet round";
`review_honesty` is told "vonc's evidence_base has ZERO facts".

So the bug file's *"the review layer is not the defect here"* is too generous, and I have
corrected it in place. The verdict it praises was right, but **over-determined**: a
vonc-shaped plan on a non-vonc site fails a generic check *and* a contaminated one, and
the run cannot tell us which did the work. The practical consequence is the trap: **a
correct post-fix plan can still be objected to by a seat hunting for a feed and a timer
this site never had**, which would read exactly like "the fix didn't work".

### And the premise is stale, not just misplaced

`review_honesty` — a **hard veto** — asserts vonc's `evidence_base` has zero facts. True
when written 2026-07-18. False since **2026-08-08 08:58Z**: `site_specs` aspect
`evidence_base`, `is_current`, 4 facts (first is `vonc-archetypes`, a `count` with a SQL
source and `verified_at: 2026-08-08`). loancalculator has no such spec row at all.

A vonc plan run today is told by its own anti-fabrication seat that four verified facts do
not exist. **Nothing updates a premise pinned inside a shared prompt when the site moves
underneath it** — which is an argument for the brief being data that does not depend on
227 being true at all.

### The design, and why the brief had to be rehomed rather than deleted

D1/D2 are owner rulings carrying "do NOT relitigate", and 59 of 61 plans all-history are
vonc's. Deleting the text without rehoming it trades a contaminated plan for a de-briefed
one on the only site that has ever used this agent in anger. So 345 moves vonc's brief
verbatim into `doc_notes` (`subject_type='experience'`, category `experience-brief`) **in
the same transaction** that strips it from the prompt.

`doc_notes` because it is existing machinery: `append_council_note` already writes there
with `subject_type='experience'`, the CHECK constraint already allows it, there is a GIN
index on `categories`, and `query_database` takes N params with an `input_data.` fallback
(`database_actions.go:32-62`) — so the whole channel is config, no Go. Rejected: the
intake work item's `spec` (the trigger's `ON CONFLICT DO NOTHING` means a second run keeps
a stale spec, and work items get completed), and a trigger argument (a brief living in a
shell-script parameter is "operators must remember X", which this lane's own rule says to
rank last).

**The `load_brief` query returns a scalar `COALESCE`, so it always yields exactly one row.**
A miss returns a visible sentinel — "(no brief on file …)" — never NULL and never empty.
That is deliberate: an empty string would read exactly like a site that legitimately has
no brief, and the sentinel is what makes a mis-keyed brief disconfirmable.

### The trap in my own fix, recorded because it nearly shipped

`compose.config.input_fields` is `["experience_context","input_data"]`. A step only
receives the fields it lists. **Adding `load_brief` without adding `experience_brief` to
that list renders `{{.experience_brief.text}}` empty and errors nothing** — the migration
would look applied and do nothing. 345 sets it and the verify block asserts it, but the
runtime proof is the positive control in the handoff: vonc must come out `leaked=true,
sentinel=false` while loancalculator comes out the other way. **One direction alone cannot
distinguish a working channel from a channel that silently loads nothing.**

### Dry run

Ran the file with `COMMIT` → `ROLLBACK` against the live DB: both guards passed, snapshot
captured, `INSERT 0 1`, `UPDATE 1`, verify `DO` block silent (so all five `replace()`
calls matched and the whole-row census is clean), `ROLLBACK`. Re-queried after: 0
`load_brief` steps, 0 brief notes, 0 backup rows, row still contaminated — the dry run
left no trace. **It has never been run against a live orchestration; no plan has been
composed with it.**

Two paren errors were caught by that dry run and nothing else would have caught them:
`to_jsonb(($prompt$…)::text)` was one short of closing its `jsonb_set`, and I then
"fixed" the `replace(replace(…))` nest by adding a paren it did not need. Eight-deep
`jsonb_set` nesting is not reviewable by eye.

### Live facts worth dating

- Anthropic credits: exhausted fleet-wide 18:00–20:00Z today (20 failures in
  `llm_call_log`); **recovered** — 21:00Z and 22:00Z hours are 81 calls, 0 failures. A
  verification run is possible.
- `agent_definitions` was bulk-updated at **22:01:02.606329Z across 187 rows in a single
  statement, no snapshot taken**. Not `UpdateUsageCount` (`usage_count` still 0), not
  migration 338 (22:12:55Z, components/css_themes, "nothing else touched"). Mechanism
  **[UNIDENTIFIED]** — I looked in `sql_for_agents`, `scripts/`, and the Go writers and
  did not find it. It did not change the contamination counts. This is why 345 carries an
  md5 drift guard on all five prompts rather than trusting the row to sit still.

## 2026-08-09 (morning) — the one thing owed is closed, and the lock is what closed it

Picked up from `HANDOFF_2026-08-08b` §2: `index`/`prose-2` still read *"Calculate your
exact monthly repayments and see the true total cost of borrowing."* — unlocked
`ported-prose`, put back by the previous session's restore, carrying the register the
owner struck one line above it.

**Now live:** *"Enter your loan amount, rate and term below to see how the monthly figure
and the total cost move together."* Written by the framework (`content_rewrite` through
`page-build-handler`, corr `26648f55-7086-4a8e-b004-d688b615f3f4`), not by hand.

### State reconciliation first — two handoffs disagreed

`HANDOFF_2026-08-08b` (f0305cc50, 23:33) and `HANDOFF_2026-08-09` (a5c8bea7e, 23:47) both
say they supersede `HANDOFF_2026-08-08`, written 14 minutes apart by concurrent sessions.
08-09 states *"the site is DONE. Nothing is owed on the SITE"*; 08-08b's §2 records
exactly one thing owed. **08-08b was right** — the 227 session did not know about the
prose-2 residue. Both files now carry a correction block pointing at the other.

[MEASURED 2026-08-09] Fleet is on **chassis v1.0.1270**, not v1.0.1269 as both handoffs
record (`kubectl get pods -l app=agent-chassis`, pods started 08:49Z). Irrelevant to this
config/content change; noted so the next reader does not repeat a stale figure.

### Was it really only one line? — census before assuming

Before firing I swept all 26 active pages for the struck claim family, eight spellings
(`mathematically rigorous|true cost of credit|true total cost|true cost of borrowing|exact
monthly|exactly what you|precisely what|pinpoint accur`). **Three hits, one real:**
`index`/`prose-2` (two hits) and `guide-car-finance-explained`/`prose-0`'s *"understanding
exactly what you're signing"* — which is about the reader's contract, not a claim about
our arithmetic, and is fine. So §2's "small and precisely located" is corroborated by an
independent measurement rather than taken on trust. ⚠ This is an absence proven only for
those eight spellings.

### Method: lock the siblings, fire at the one writable row

§3's remedy (write the instruction conditionally, per-section) was applied **and it was
not enough on its own.** The mechanism that actually held was the lock:

1. Backed up the page → `page_components_bak_20260809_index_prose2` (5 rows).
2. Locked `prose-0` (durable, owner-approved copy), `prose-1` and `prose-4` (temporary op
   lock), leaving **`prose-2` as the only agent-writable row on the page**.
3. Fired the rewrite with a section-scoped conditional prompt.
4. Verified, then released the two temporary locks.

**`locked_at` is the load-bearing column, not `locked_by`** —
`AgentWritableSQLFor` is `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at <
NOW())`. A row with `locked_by` set and `locked_at` NULL is still fully writable. Checked
`page_components` for triggers first (there are none), so a lock write does not bump
`updated_at` and the pre-run timestamps survive as a baseline.

### The result, measured both directions

`save_page_sections` DELETEs every agent-writable row and re-inserts, so `prose-2` is a
**new row id** by design. The four locked/untouched rows must be identical in id,
`updated_at` **and** content hash — and were:

```
slot      id                  updated_at                       content_md5
prose-0   9e7cbaa2 unchanged  2026-08-08 22:25:00.540115 same  dcadc7ce unchanged
prose-1   2c363ac0 unchanged  2026-08-08 22:28:49.980291 same  7d69f637 unchanged
prose-2   7bd47f4a -> 6bf6ff7d  2026-08-09 09:15:05.173243     3a8068a3 -> dc8714f7
tool-3    993eda99 unchanged  2026-08-03 10:31:44.772743 same  56786308 unchanged
prose-4   f082c7d2 unchanged  2026-08-08 22:25:00.566302 same  8d553e76 unchanged
```

Live: struck register **0** across four spellings, owner's opening appears **exactly
once**, new strap line **exactly once**, 26/26 serving (200 + ≥2000 B + DOCTYPE guard),
**toolgolden 11/11 exact** against `GOLDEN_2026-08-08_voice_h_complete.json`.

### ⚠ THE FINDING: the conditional framing leaked anyway — §3's remedy is necessary, NOT sufficient

I compared what the writer **proposed** for every section against what was stored before
the run (`llm_call_log.response_text` vs the backup table). This is the check that
matters, and it says:

```
slot      proposed  stored   byte-identical
prose-0     1102     1102     TRUE   <- obeyed "leave your section alone"
prose-1      400      133     FALSE  <- LEAKED: kept its <h1>, APPENDED a strap line
prose-2      143      117     FALSE  <- the intended change
prose-4     2813     2813     TRUE   <- obeyed
```

`prose-1` is the `<h1>Standard Loan Calculator</h1>` section. It returned its heading plus
*"Enter your loan amount, rate and term below to see how the repayments break down over
the life of the loan…"* — **the exact job I had assigned to `prose-2`.** Unlocked, the page
would now carry that strap line twice, and this session would have repeated the previous
one's failure with a better-worded prompt.

**Why it leaked, precisely.** My condition read *"IF THIS SECTION IS the one-sentence
introduction sitting directly under the 'Standard Loan Calculator' heading"*. `prose-1`
**is** that heading — so it read the condition as "I should have an introduction under my
heading" and supplied one. **A conditional whose condition names a neighbouring landmark
is ambiguous to the neighbour.** The condition has to be decidable from the section's own
bytes alone. Quoting the sentence to be replaced (which I also did) was the part that
worked for the two distant sections; the landmark reference is what gave the adjacent one
a way in.

This is the **fifth** instance of the lane's one shape, and the first where the prescribed
remedy was followed and still failed. §3 should now read: write it conditionally **and**
lock every sibling you are not targeting.

> **MISSTEP, caught before it reached a claim.** Three `lock_blocked_change` items appeared
> (`index:prose-0`, `prose-1`, `prose-4`) and my first reading was "the guidance leaked and
> the locks caught it". **That inference is unfounded.** `emitLockBlockedChangeItem` fires
> whenever an incoming section matches a locked slot — it records `slot_name`, `locked_by`,
> `blocked_action`, and **no proposed content at all** — so it fires identically whether the
> writer rewrote the section or returned it byte-for-byte. The composer emits all five
> sections every run, so those three items were guaranteed before the writer wrote a word.
> The leak is real, but the evidence for it is `llm_call_log.response_text` compared against
> the backup table, which is a different query answering a different question. Filed as a
> landmine — the item *looks* exactly like a content-diff detector and is not one.

### Housekeeping

- `index-prose2-source` (`e5bf5c46`) left at `needs_human_review`, matching the lane's
  convention for pinned prompt sources (`6d52beaf`, `7933edd4`, `50c8ba5c`). The prompt
  text is also in git at `voiceh_index_prose2_source.sql`, but **the DB row is what drives**
  (§5).
- Fired item `e1bfffe1` and the three `lock_blocked_change` notices closed to `complete`
  with `resolution_path` set — a `detected` leftover blocks any future re-fire on
  `idx_swi_dedup` (§5).
- Locked rows on the site are now **17**: 12 tool + 4 CSS carrier + 1 owner-approved
  (`index`/`prose-0`, `loancalculator_owner_approved_20260809`).
- `voiceh_rewrite_v3.sh` gained `SRC_ITEM` / `KEY_PREFIX` / `SUMMARY` env overrides,
  defaults unchanged, so a one-off reuses the dispatch path instead of becoming a v4 copy.
- New: `check_site_serving.sh` — the 26-page check with the B2-blob guard baked in
  (200 **and** ≥2000 B **and** starts `<!DOCTYPE`), so it cannot be run the unguarded way.
- Landmine appended and `landmines-sync.py --apply` run (1504 owned `doc_notes` rows).
  **I did NOT run `landmines-verify-dispatch.sh`** — it fires a `landmine-verifier` per
  new/changed entry, and the sync reported two pending, the other belonging to another
  session. My entry substitutes first-hand verification, stated per the 2026-07-31 ruling:
  the code is cited (`lock_helpers.go:203-219` — the spec has no content field;
  `save_page_sections_action.go:769` — the guard matches on slot name alone) and the
  behaviour was measured on a real run with both arms (2 obeyed, 1 leaked). A verifier
  pass is still owed if anyone wants it independently confirmed.

### Second misstep, same session — I nearly "corrected" a figure that was right

Writing the next handoff I went to ground §7's *"the house voice block is SEVEN copies
across seven agents"* rather than repeat it unchecked. The obvious census —
`LATERAL jsonb_each(default_config->'workflow'->'steps')` reading
`step->'config'->>'prompt_template'` — returned **6**, and pointedly did **not** include
`page-content-writer`, the agent that writes this site's every section. That is a
striking result and I was one keystroke from writing "the figure is stale, it is 6, and
it does not even touch our writer".

**It was my query that was wrong.** A loop step's prompt lives at
`config->'sub_workflow'->'steps'->'generate_content'->'config'->'prompt_template'` —
below where the census looks. My first attempt to check that one agent directly used
`config->'steps'->…` (the key is `sub_workflow`), which returned NULL, i.e. it
*reproduced* the blindness and looked like confirmation. What broke it was asking for
`length()` as a positive control: NULL length = I measured my own typo. On the right
path the prompt is **12,813 chars and does carry the house voice**.

Nesting-proof census (`default_config::text ILIKE '%size of the fact%'`): **7 agents** —
`content-creator-about`, `content-creator-hero`, `content-creator-hero-without-research`,
`content-writer`, `grounded-explainer`, `page-content-writer`,
`simple-content-writer-with-approval`. **§7's figure is CONFIRMED, not stale.**

Two things carry forward. **(1)** The fleet-wide base-prompt job in §7 is exactly a
fleet-wide prompt census, so this is the first query it will write and the first way it
will be wrong. Filed as a landmine. **(2)** `page-content-writer` IS in scope for that
change — the agent this lane has spent a week driving is one of the seven, which is an
argument for the wide option the owner already chose, not against it.

---

## 2026-08-09 (afternoon) — `bugs_open/227`: 345 applied, proven both directions, and the prescribed proof turned out to be inert

### Applying it

Pre-checks first, because 345's drift guard pins five prompts by md5 and the row had been
bulk-touched once already with no snapshot. All five matched
(`8b05c372…`, `dfb8111e…`, `4a86a799…`, `fffda680…`, `ebe84e4f…`), `load_brief` absent,
`load_schema_hint.next_step` still `compose` — so no session had touched the prompts since
the file was composed at 22:45Z. Applied clean at ~11:50Z: `DO`, `DO`, `BEGIN`, snapshot
NOTICE, `INSERT 0 1`, `UPDATE 1`, `DO`, `COMMIT` — exactly the sequence the header predicts.

Structural verification straight after: case-insensitive census over the whole live row
**0** (from 48 hits across five steps), chain `load_schema_hint → load_brief → compose →
persist_plan`, `compose.config.input_fields` = `["experience_context","experience_brief",
"input_data"]`, vonc's brief rehomed at 7,908 b under `categories ? 'experience-brief'`.

### The verification I ran was inert, and the control is what showed it

§2 of the handoff, the bug file, and 345's own VERIFY header all prescribe the same pair on
`compose`: `prompt_rendered ~* 'provocation|gauntlet'` FALSE and
`prompt_rendered LIKE '%no brief on file%'` TRUE. loancalculator returned `f` / `t`, and I
recorded the fix as behaving exactly as designed.

Then the vonc control returned `leaked=t` (expected) **and `got_sentinel=t`, where all three
documents demand FALSE**. First reading: the fix is broken — the channel is handing vonc the
fallback instead of its brief. It is not. **The phrase "no brief on file" occurs once in the
static `compose` template 345 installs**, in the instruction covering the no-brief case. So
the assertion is TRUE on every run of every site — including one where `load_brief` was never
wired into `input_fields`, which is the precise silent failure the header calls "the single
most likely way for this migration to look applied and do nothing". **The check could not
come out false, and it had been reviewed as sound in three places.**

The disconfirmable forms, both cheap: the **count** of the phrase (2 = template + rendered
fallback, 1 = template only), or the substring only the `COALESCE` emits. With those:

| run | corr | phrase hits | COALESCE fallback | leaked | prompt |
|---|---|---|---|---|---|
| loancalculator `debt-difficulty-help` | `c3976aab` | 2 | TRUE | **FALSE** | 24,721 b |
| vonc `vonc-spark-game` | `72f540d3` | 1 | FALSE | TRUE (correctly) | 70,427 b |

Same step, opposite outcomes, keyed only on `subject_key`. Better still — and the thing I
should have reached for first, since it needs no sentinel reasoning at all —
`collected_data->'experience_brief'->>'text'` shows the loaded value **directly**: vonc's
opens `## The diagnosis you are fixing (three broken surfaces, artifact-verified
2026-07-17)`, 7,908 b, verbatim; loancalculator's is the sentinel. Filed to `WRONG_CALLS.md`
and as a landmine on `llm_call_log.prompt_rendered` (synced, 1,604 rows).

**Both councils returned `approved`** (2 advisory objections each, none high-severity),
which retires finding 2's worry: the de-contaminated seats did not object about a missing
feed or timer. The `debt-difficulty-help` plan of record is now clean — 11,442 b, names loan
and debt subjects, `body ~* 'provocation|gauntlet|arena|vonc|spark'` **false**. First
non-vonc experience plan in this system's history that does not describe vonc's pages.

### The 227 second defect bit this session's own verification

Firing the vonc control re-planned vonc: `persist_plan` wrote the new plan `is_current=true`
before the council voted, superseding `b6fdbc09`, vonc's plan of record since 2026-07-25.
Nobody asked for vonc to be re-planned — it was a test. **Restored by hand** in one
transaction (`idx_doc_plans_current` is UNIQUE where `is_current`, so demote before
promote): `b6fdbc09` current again, `superseded_at` cleared, and the reason written into the
`notes` column of all three rows. The new plan `2ec02a7e` was council-**approved**, so it is
kept and demoted rather than deleted — the vonc lane can promote it deliberately.
Worth stating plainly: **§3 is no longer only a hazard to a rejected plan. It fires on any
verification run, and it silently changed another lane's plan of record.**

### The bulk-writer the handoff asked about — bounded, not identified

The handoff flagged that the row had been bulk-updated with 186 others and offered ten
minutes of curiosity. What I can say from measurement: waves of **188 rows sharing one
microsecond timestamp** (183 distinct types) do happen — so a single statement, not
`UpdateUsageCount`, whose `WHERE type = $1` cannot span 183 types. One landed at 12:22:37Z,
**after** my apply, and **`load_brief` survived it intact** (re-censused post-wave: 0, chain
and `input_fields` unchanged). So whatever it writes, it is not replaying `default_config`
from a seed — which was the live risk to this fix and the reason I chased it at all.

Smaller waves at 13:22:15 (15 rows) and 14:00:07 (14 rows) **did** change `default_config`,
and those are identifiable: `content-creator-*`, `visual-designer`, `content_researcher`,
`fix-proposer`, then `council-gate` a minute later — other sessions' migrations 348/349/350
plus a `099_SYNC_gate_roster.py --apply`. Ordinary concurrent traffic, not a mystery.
**[UNIDENTIFIED]** the 188-row wave itself: no `schema_migrations` row at that timestamp.
Bounded, not solved.

**One trap for anyone re-reading this row's history: 345 does not set `updated_at`.** The
row's timestamp is whatever last touched it, so it does **not** record the apply — I read
12:22:37Z on a row I had changed at 11:50Z and briefly took the wave for a clobber.

## 2026-08-10 — post-roll verification, and the bulk-writer is IDENTIFIED (correcting yesterday)

**345 survived a full rebuild and roll.** Fleet is on **v1.0.1277** (chassis pods started
2026-08-09 21:34:53Z / 21:35:18Z). Re-verified this morning against the live row:
`load_brief` present, case-insensitive census **0**, chain `load_schema_hint → load_brief`,
`compose.config.input_fields` still carries `experience_brief`, vonc's 7,908 b brief note
intact, and both plans of record unchanged (`debt-difficulty-help` 4bfcb286 clean;
`vonc-spark-game` b6fdbc09 — **my hand-restore held across the roll**).

> **CORRECTED 2026-08-10 — yesterday I wrote `[UNIDENTIFIED]` against the 188-row wave and
> said it was "bounded, not solved". It is now solved, and it is not a mystery at all: the
> wave is the DEPLOY stamping `agent_definitions.image_tag`.**
>
> Measured across the roll with a fleet-wide column-by-column fingerprint diff (200 rows,
> yesterday 14:0xZ vs today):
>
> | column | rows changed |
> |---|---|
> | `image_tag` | **189** |
> | `updated_at` | **189** |
> | `default_config` | **4** |
> | `usage_count` / `idle_timeout_seconds` / `status` / `is_active` | **0** |
>
> 190 rows now read `v1.0.1277`. The wave is `scripts/deploy/update-agent-images.sh`, the
> deploy-time hygiene sync described in `platform/orchestration/actions/agent_image.go:21-24`
> and owned by `bugs_open/066` — a spawned agent pod takes its image from its
> `agent_definitions` row, so the deploy syncs those rows to the new tag. It leaves no
> `schema_migrations` row because **it is not a migration**, which is exactly why it read as
> trail-less. `usage_count` unchanged at 0 confirms the previous session's correct
> elimination of `UpdateUsageCount`.
>
> The four `default_config` changes are attributable and none are mine —
> `domain-strategist`, `image-build-handler`, `quality-discovery-agent`, `section-editor`,
> i.e. other lanes' migrations landing in the same window. **`now()` is transaction start
> time, so rows sharing one microsecond mean one TRANSACTION, not necessarily one
> statement** — which is what let me mistake a deploy sync plus concurrent migrations for a
> single 188-row rewrite.
>
> **The conclusion that matters for every config-only fix on this estate: a roll does NOT
> replay seeds over `default_config`.** DB config survives a rebuild. I chased this because
> if it had been a seed replay, 345 would have been erased by the next deploy — and the
> disconfirming result was available all along (a seed replay changes ~190 configs, not 4).

## 2026-08-10 — 227 second defect fixed: migration 363, persist only after approval

**Owner decision: route (a), the config-only rewire** — not the `write_doc_plan`
`set_current_when` seam. Dry-run first (guards + verify passed, rolled back, no trace
confirmed: graph still pre-363, 0 backup rows), then applied ~10:40Z.

Six edges. `compose`/`recompose`/`reframe` → `review_journeys`;
`check_approved.then_step` → `persist_plan`; `persist_plan.next_step` → `complete`;
`complete_escalated.output_fields` drops `plan_persisted`. Persist is now reachable **only**
from the approved branch.

**The verify block asserts the PROPERTY, not the writes.** It counts steps other than
`check_approved` that reach `persist_plan` and requires 0 — because `jsonb_set` on a wrong
path **adds a key instead of failing**, so "the value I wrote is there" would have passed
even if I had written it somewhere harmless. Reading back the six fields I set could never
have caught a mis-pathed edit; counting the reachable edges can.

**The assumption I checked before moving the write.** Persist reads
`plan_body_field: proposal.result`, and it now runs much later in the graph. If `recompose`
declared its own output field, the rewire would have persisted the **first draft** on every
revise round — a wrong-content failure that looks like success. It does not:
`compose`, `recompose` and `reframe` all declare `output_field: proposal`. Confirmed against
the two 08-09 runs, not read off the config alone —
`length(collected_data->'proposal'->>'result')` equals the length each run actually persisted
(11,442 and 13,840).

**Stated loss:** nothing is persisted on the escalated path now, so
`complete_escalated.output_fields` would have referenced a value that no longer exists. The
escalated plan survives in `collected_data->'proposal'->>'result'` and `llm_call_log`.
Persisting it *not-current* requires route (b). I rejected persisting under a derived
`subject_key` (`'<key>:escalated'`) — it invents a convention nothing else knows to read.

### The verification is PARTIAL and I am recording it as partial

Fired `9150dd54-6129-464b-8600-771e0a84408a` at 10:44:47Z, `debt-difficulty-help` plan-row
baseline **5**. It tests the **approved** arm through a signal that could not have come out
the same before: an approved run taking N compose rounds used to write N rows (08-09 wrote
**three** for one approval) and must now write **exactly one**.

**But 363 exists for the REJECTED arm, and that arm is unobserved.** Both 08-09 runs were
approved and a veto cannot be induced on demand. So: proven for the approved path, **owed**
for the vetoed one — either wait for a natural veto or seed a deliberately unbuildable
experience and assert no new row. Writing this down explicitly because this lane has already
produced one check that could not fail, on this same bug, two days ago. **An approved run is
not evidence about the rejected path**, however green it comes back.

### Same-day correction — my 363 verification signal was non-discriminating, and it passed

Run `9150dd54-6129-464b-8600-771e0a84408a` came back **COMPLETED / approved**, plan rows
**5 → 6, one current**, new plan `051af223` 10,075 b, `leaked=false`. Exactly the number I
predicted, and I was about to report it as proof of the rewire.

> **CORRECTED 2026-08-10 — "an approved run must now write exactly ONE row" only
> discriminates when the run takes TWO OR MORE compose rounds.** This run was approved on
> **round 1** — `compose` ×1, no `recompose`, no `reframe`. Under the OLD graph a
> single-round run also writes exactly one row, so "1 row" is the same answer before and
> after the fix. **The check I wrote into 363's header, the handoff and my own report could
> not have come out otherwise on the run I actually got.** The 08-09 run wrote three rows
> only because it took three rounds; I generalised from that without noticing the round
> count was doing the work.

**What actually proves it, and it discriminates on any run: the ORDERING.** The old edge was
`compose → persist_plan → review_journeys`, so under the old graph a plan row EXISTS by the
time the run is executing any review step. Sampled mid-flight at 10:4xZ, the run was
`EXECUTING_STEP|review_journeys` with the row count **still at the pre-run baseline of 5** —
past the point where it used to persist, having written nothing. I took that reading almost
by accident, while checking progress. **It is the only observation in this session that could
have come out the other way, and it is now the recorded proof for the approved arm.**

Two things to carry:
1. **A count-based check inherits its discriminating power from a property of the RUN
   (how many rounds), not from the fix.** Before trusting one, ask which property of this
   particular run made the number differ — and if the answer is "nothing", the check is
   inert for it. Same shape as the sentinel two days ago: an assertion that returns the
   expected answer for a reason unrelated to what is being tested. **Twice in three days,
   same lane, both on this bug.**
2. **The rejected arm is still unobserved** and is still what 363 exists for. Nothing above
   changes that.

---

## 2026-08-10 afternoon — the REJECTED arm, observed; and a fleet-wide API cap mid-session

Picking up the one thing `HANDOFF_2026-08-10_continue_here.md` left owed: 363 moved the plan
write onto the approved branch, and no vetoed run had ever been watched under the new graph.
Route taken: the handoff's second option — **seed a deliberately unbuildable experience**.

### How the probe was built, and why it cannot contaminate anything

`load_brief` selects `doc_notes` by **`subject_key`, not by `site_id`** (read the live step
config, don't infer it) — so a brief filed under a probe key is invisible to
`debt-difficulty-help` and `vonc-spark-game`. Fixture:
`probe_363_veto_arm_brief.sql` in this directory, deliberately NOT in `sql_for_agents/`
so the migration runner can never sweep it up.

The brief is a realistic owner brief for `live-lender-approval-race`: a live partner-lender
decisions API polled from the page with a key in the query string, a presence counter by
postcode, per-visitor state written server-side and read back cross-device, and an explicit
owner line that a coming-soon label is not acceptable. Every clause is something a real
client asks for and a static host cannot do. The test-artefact marking lives in `categories`
and `created_by`, **never in the body** — the body is what `compose` is handed, and a body
announcing itself as a test would steer the judgement under test.

**Pre-flight, learned from the sentinel trap two days ago:** ran `load_brief`'s *exact* query
for the probe key before firing and confirmed it returned the brief, not the no-brief
sentinel. A run where the brief silently failed to load would have planned from live context
alone, probably produced something buildable, been approved — and I would have read that as
"the council did not veto" rather than "the fixture never arrived".

### Run 1 — corr `d81aa5f4-a732-4fb3-b438-4ff496ef7ba2`, and this is the proof

| time (UTC) | step | `doc_plans` rows for the key |
|---|---|---|
| 14:39–14:40 | `compose` (returned 12,189 b at 14:40:33) | 0 |
| **14:40:48** | **`EXECUTING_STEP @ review_journeys`** | **0** |
| 14:42:42 | council round 1: **`veto from feasibility`** | 0 |
| 14:43–14:44 | `reframe` (returned 7,661 b) | 0 |
| 14:44:12 | `EXECUTING_STEP @ review_journeys` (round 2) | 0 |
| 14:45:51 | council round 2: `approved with 2 advisory objection(s)` | 0 |
| 14:46:26 | `COMPLETED @ complete` | **1** |

**Under the OLD graph this run writes TWO rows, and the first of them is the VETOED plan,
`is_current`.** That is bug 227's second defect verbatim — the 08-08 sequence was
compose-persist, veto, reframe-persist. Here the vetoed composition was never written at all:
the count was still 0 at 14:40:48, i.e. past the point where the old graph had already
persisted, and stayed 0 across the veto and the entire reframe round.

The feasibility veto is a real one, not a rubber stamp — four objections, three high:
"requires a POST write endpoint … this site is static with no server"; "cross-device
'return tomorrow on a phone' … a static host has no mechanism for this"; "a live API key
embedded in client JS … exposes the key in every page load".

**A check nobody had run, and it matters more than it looks.** The persisted body is
**7,661 b = the `reframe` response exactly**, not compose's 12,189 b. 363's header verified
the "compose, recompose and reframe all write `proposal`" assumption against compose+recompose
runs only; the **reframe** branch had never been measured. Had reframe written to its own
field, moving the write later would have persisted the **vetoed** draft on approval — a silent
wrong-content failure, and the ugliest possible version of it.

### What is STILL not observed, and why it is hard BY DESIGN

A run that **ends** non-approved (`complete_escalated` / `complete_refused`) leaving no row.
Reading the code rather than guessing at it:

- `reframe`'s prompt: *"If the vetoed feature admits no honest minimal-real version, demote it
  to a labelled coming-soon panel and move the real version to the LATER list — that is an
  acceptable honest MVP."* The reframe is **instructed to converge on something approvable**.
- `applyCouncilCaps` (`platform/orchestration/actions/diagnose_council_decide_action.go:663`):
  `shouldReframe := rejected && rejectedCount <= 1 && round < maxRounds`. Escalation needs a
  **second** rejection in the same run.

So the handoff's "wait for a natural veto" cannot deliver this arm: a veto is not terminal,
and the machinery exists to stop it being terminal. My unbuildable brief drew the veto it was
built to draw and the run still ended approved.

### Run 2 — the escalation probe, killed by a fleet-wide API cap

To make the terminal state reachable in one round I capped the council: `max_rounds` 5 → 1,
so **any** non-approved round-1 verdict routes to `complete_escalated` (`round < maxRounds`
is false at round 1). Deliberately a capacity knob, not the persist wiring under test.
Armed a detached 25-minute restore as a safety net first, then fired the same key at
14:51:30 — reusing the key on purpose, because `load_context` carries no prior plan (checked)
so there is no bias, and the observation becomes the damage claim itself: *the plan of record
must not change*.

It died 30 seconds later:

```
step compose failed: … API request failed with status 400:
{"type":"invalid_request_error","message":"You have reached your specified API usage limits.
 You will regain access on 2026-09-01 at 00:00 UTC."}
```

**Fleet-wide, account-level, not this lane's doing.** Same minute, another session's
`council-gate` run died on `review_architecture` with the identical message
(`request_id` req_011CduFpnewdmmf9ak88Ww2t). Every LLM-driven agent is blocked until the owner
raises the cap.

> **CORRECTED same day — I first wrote "last successful call across the whole estate:
> 14:51:15".** That figure was the **minimum** `created_at` in the window I happened to
> select, not the last success; I read a `min` off a summary row and reported it as a
> boundary. The real value is **14:51:45.067Z**:
> `SELECT max(created_at) FILTER (WHERE success) FROM llm_call_log`. **What caught it:**
> the `bugfix_236_site_availability` lane had independently diagnosed the same outage and
> appended it to `LANDMINES.md` with 14:51:45 — I saw their figure while checking my own
> commit for same-file passengers, and re-measured rather than assuming mine was right.
> **And their instrument is better than mine:** `llm_call_log` has a `success` **boolean
> column**; I was testing `response_text IS NOT NULL AND <> ''`, which agrees here but is a
> proxy for the thing that is recorded directly. Their entry also carries the sizing rule —
> on a quiet fleet the outage is only ~5 error rows, so **size it by the absence of any
> success, never by the error count**.

`max_rounds` restored to 5 and re-read from the live row immediately (window 14:50:20–14:53:10,
one run in it — mine). Safety-net task stopped.

### ⚠ THE MISSTEP — the third inert check in three days, and I nearly filed it

My first sight of run 2 was `COMPLETED @ complete_refused | rows=1 | current=6ebe06f5` —
a non-approved terminal state with the plan of record unchanged. **That is precisely the
sentence I came here to write, and it would have been worthless.** `compose` never returned,
so no plan existed to persist; the OLD graph writes nothing on that run either. The reading
is true and proves nothing.

What caught it: the run finished in 30 seconds instead of eight minutes, so I read
`collected_data->'__step_error'` instead of the status — the status says `COMPLETED` even
though the step failed (a known landmine, and it earned its place again today).

**Three for three on this bug now**: the `no brief on file` sentinel (08-09), the one-row
count on a single-round run (08-10 morning), and this. All the same shape — *an assertion
that returns the expected answer for a reason unrelated to what is being tested*. The common
factor is not carelessness, it is that **each check was designed against the run I hoped for
rather than the run I got**. The cheap habit that would have caught all three: before reading
a result, say out loud what the failing version of this run would look like — and if the
answer is "the same", the check is inert for this run whatever it returns.

### The structural backstop for the arm still owed

Not a substitute for observing it — but the escalation path cannot reach a write, and that is
checkable rather than arguable. Every step-target reference to `persist_plan` in the live row:

```sql
SELECT s.key AS step, y.k AS field
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s,
     LATERAL (SELECT k, v FROM (
        SELECT 'next_step' k, s.value->>'next_step' v
        UNION ALL SELECT 'config.'||c.key, c.value#>>'{}' FROM jsonb_each(COALESCE(s.value->'config','{}'::jsonb)) c
                   WHERE jsonb_typeof(c.value)='string'
     ) x WHERE v='persist_plan') y
WHERE d.type='experience-planner' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL;
```
→ exactly one row: `check_approved | config.then_step`. (Do **not** count occurrences of the
literal in `default_config::text` for this — 370 added the words to two descriptions, so the
raw count is now 3 and means nothing. Scan the target fields.)

Combined with run 1 — where the engine demonstrably honoured the rewired edges — the
escalated arm follows from the same mechanism just observed working. Recorded as
**[INFERRED]**, not as an observation, and the handoff says so in those words.

### Migration 370 — three strings 363 left describing the graph it replaced

Found while reading `check_reframe`/`complete_escalated` to design the probe:

- `complete_escalated.description`: *"The current (rejected) plan **stays is_current** but MUST
  NOT be built"* — asserts the exact state 363 made unrepresentable. A session reading it goes
  looking for a rejected plan of record, and the honest conclusion from not finding one
  ("someone already demoted it") is wrong in the other direction.
- `complete_escalated.config.success_message`: same claim, and this one travels to whoever
  reads the escalation.
- `recompose.description`: *"loops back to persist + review + decide"* — names a dead edge.

`370_experience_planner_escalation_descriptions_catch_up_with_363.sql` (+ ROLLBACK), config
only. Dry-run first (guards + in-transaction verify passed, rolled back, stale text confirmed
still live), then applied. Its verify block sweeps the WHOLE row for each retired claim rather
than checking the three paths it set — a `jsonb_set` on a wrong path silently adds a key, so
"my new string is there" is not "the old one is gone". Post-apply: 0 retired claims remain.

### Correction to my own commit message, minutes after writing it

Commit `3333b3299` says the `bugfix_236` lane's LANDMINES recurrence and the
`loanandmortgagecalculator` lane's WRONG_CALLS entry "ride here as a same-file passenger".
**They do not — it happened the other way round.** Between my `git diff` (which showed both
files dirty with their work and mine) and my `git commit`, that lane committed
`9fb7eee9c`, which swept **my** two entries into **their** commit. So my pathspec named two
files that were already clean, contributed nothing, and the commit landed with 9 files, not 11.

Nothing is lost — both entries are at HEAD and intact — and forward-only means the message
stands as written. Recording it here because the message is now the only thing that would
mislead someone doing archaeology: **looking for my landmine entry in my own commit finds
nothing.** It is exactly the hazard CLAUDE.md describes ("committing per task stops you
sweeping up others' WIP; it cannot stop a session that still runs git add -A from sweeping up
yours"), and the interval was about four minutes.

The cheap check I skipped: **re-run `git diff --numstat` on shared append-only files
immediately before the commit, not before writing the message.** I ran it, wrote a paragraph
about passengers off that reading, and by the time the commit executed the reading was stale.

---

## 2026-08-10 — framework-rebuild thread: URL defect found and half-fixed; rebuild NOT started

Owner directed a full framework rebuild of this site (to audit the framework, not the
CLI's hand-writing), after two false live claims were traced: footer "shows its own
arithmetic" (hand-authored chrome, locked, outside every claims control) and the
how-loans-are-calculated guide's "month-by-month breakdown" (framework-written, passed
the claims gate — the gate has no check for claims about our own software's behaviour).

Pre-flight exploration found the blocker BEFORE any planner run: CanonicalisePage has no
flat URL form for tool/guide/game, so a plan sync would move 24/26 live URLs in place
(bugs_open/241). Owner: fix the framework first, rebuild in place, release all locks.

DONE: FlatURLs opt-in field + tests, commit 57a7fcbb4, council corr
6fdb9ce6-9ee2-4550-86ac-893ca0b44c3f (verdict UNREAD), register BLD-018 (rode to HEAD in
another session's 4451b2a0a as a same-file passenger — content intact).
NOT done: plumbing, seeding, backups, lock release, submission — the full ordered list
with every measured ground fact is in HANDOFF_2026-08-10_framework_rebuild_continue_here.md.

Misstep logged: my page_canonical.go edit sat briefly on the shared tree referencing an
undefined helper (plan-mode interruption mid-edit) — the tree did not compile for that
package until this session resumed. The check that would have caught it sooner: go build
the package in the same turn as ANY code edit, before yielding.

## 2026-08-10 evening — the cap lifted, and the escalated arm closed on the first run

Two things had changed since the afternoon: the Anthropic cap was gone (successes resume at
**18:00Z**; the outage ran 14:51:45 → ~17:0xZ, ~5 failures in it, which is why the sizing rule
is *absence of success*, not error count), and the fleet had rolled to **v1.0.1283**.

**Re-verified before firing anything, because a roll is not evidence either way:** 345's
chain (`load_schema_hint → load_brief`), 363's rewire (`compose → review_journeys`,
`check_approved.then → persist_plan → complete`), 370's strings (0 stale claims) and
`max_rounds = 5` — all intact at the new tag. Third confirmation now that a config-only fix
survives a rebuild.

### The run — corr `c4127fe7-b6b0-4c44-9e26-fd869a09a873`

Pre-flight: 0 planner runs in flight, pods 19 minutes old (past the ~300s spawn window), brief
still seeded, baseline **1** row (`6ebe06f5`, 7,661 b, `is_current`). Capped `max_rounds` 5 → 1,
armed the 25-minute restore, fired the same key.

| time (UTC) | step | rows | is_current |
|---|---|---|---|
| 22:06:37 | `compose` **succeeds, 10,498 b** | 1 | 6ebe06f5 |
| 22:07:04 | `EXECUTING_STEP @ review_journeys` | 1 | 6ebe06f5 |
| 22:07–22:09 | all five seats, all successful | 1 | 6ebe06f5 |
| 22:09:26 | council: `rejected` / `veto from feasibility` / `should_reframe=false` / round 1 | 1 | 6ebe06f5 |
| 22:09:54 | **`COMPLETED @ complete_escalated`** | **1** | **6ebe06f5** |

`updated_at` on that surviving row is still **14:45:54** — it was not superseded, not touched.
No `__step_error`. `collected_data->'proposal'->>'result'` still held **10,498 b** at the end.

**Why this one discriminates where the 14:51 attempt did not: a write was POSSIBLE.** The
plan existed, in the field persist reads, at the moment the run took the escalation branch —
and no row appeared. Under the old graph that 10,498 b vetoed plan is persisted at ~22:07,
supersedes `6ebe06f5`, and the run ends escalated with a council-vetoed document as the plan
of record. That is the defect, and it is gone.

The `10,498 b` is the load-bearing figure, not the `rows=1`. I wrote the row count into the
poll loop this morning and it was inert twice; here it only means something *because* the
composed plan is non-empty. **The negative control has to be inside the observation, not in a
sentence next to it.**

Restored `max_rounds` to 5 and read it back from the live row; stopped the safety net.

### Cleanup

- Probe brief `5730bcf3` **deleted** — its job is done and the fixture SQL is committed, so it
  is reproducible without leaving a fabricated brief in `doc_notes` for a future reader to find.
- Probe plan `6ebe06f5` **demoted, not deleted** (`is_current=false`, reason in `notes`): it is
  the evidence for the veto arm — its 7,661 b is what proved persist reads the *latest*
  composition. `live-lender-approval-race` now has 0 current plans.
- Intake row cancelled again (the trigger re-inserts one per fire; the previous cancel does not
  block a new row, so this needs doing after every probe).

## 2026-08-11 (framework-rebuild thread) — round-1 verdict read, plumbing written, and a declared sweep

**Step 1 done.** Council verdict on the FlatURLs field (corr `6fdb9ce6`): **APPROVED, round 1**,
"2 advisory objection(s) — none high-severity", `complete_approved` at 19:21Z on 2026-08-10.
The advisories were real work for this session: editquality (medium) — the field is only a
prerequisite capability, the plumbing is the fix; prior_art_librarian (medium) — settle which
of the THREE pages-upsert helpers governs the planner path before trusting the layer.

**Step 2 written.** `siteUsesFlatURLs` (new `site_url_shape.go`): ONE reader of
`structure.url_shape`; both surfaces (`WriteSitePlanAction`, `SyncPagesToDBAction`) thread it
into `PageDescriptor.FlatURLs`; contract test `TestFlatURLFlagReachesBothCanonicalisationSurfaces`
pins the pair (anchored extraction, not a bare grep). Six sqlmock cases on the helper. Both
suites + `go build ./platform/...` green at HEAD **via `git archive`** (tree is dirty with
other lanes' WIP — one of which, per 7a066dba1's message, does not compile).

**Pre-write measurements that shaped it** (all [MEASURED 2026-08-11] live):
- pages table: 24 active tool/guide rows, ALL dir-flat (`/tools/<s>.html`, `/guides/<s>.html`)
  — byte-identical to `nestedOrFlatURL(dir, bare, true)`. The flag CAN express this site.
- upsertPage: `url = EXCLUDED.url` unconditional; sync loop always overwrites `page["url"]`
  from CanonicalisePage first → canonicalisation is the right layer for the plan-sync path.
- The "second flat mechanism" is verbatim adoption (fidelity=locked) — preserves crawl URLs,
  bypasses CanonicalisePage by design. Preserver, not rival synthesiser.
- `resolveToolPageIdentity` keeps an existing row's stored identity → tool redeploys can't
  move flat URLs. `nestedRoleFromURL` matches `parts[0]`, so flat URLs still corroborate
  ValidateRoles — no interaction.
- NOT wired (named, deliberate): apply_gap_plan/create_blog_posts/create_tool_component/
  companionGuideIdentity — new-page synthesis only, cannot move a live URL.

**The sweep.** My uncommitted edits + the untracked helper/test were carried to HEAD by the
215 lane's `7a066dba1` as a DECLARED same-file passenger (their fix edits the same two
surfaces; message states they could not separate the hunks and keep HEAD compiling). Exactly
the CLAUDE.md scenario: nothing lost, forward-only holds, no re-commit needed. Cost: the code
commit bears no `Council-Submitted:` trailer — the join lives in bugs_open/241 + BLD-018 +
here instead. Their realised-identity layer composes with the flag: synthesis (mine) vs
stored-identity override (theirs), complementary.

**Council round 2 submitted** (the plumbing): corr `70256656-4ada-465e-b959-096ae7225eb9`,
2026-08-11. Find by payload, not printed id. Budget ~30 min.

**Sequencing constraint honoured:** the chassis roll (step 2's tail: build v1.0.1289, deploy,
pod-grep `url_shape` + negative control) WAITS for the council verdict — council seats run on
the chassis and a roll kills an in-flight run. Also: no orchestration dispatch within ~300s
of the chassis roll.

**NOTIFY read:** owner decision 2026-08-11 (url_field over renames) — no URL/name changes on
this site; confirms, does not conflict.

### 2026-08-11 (same session, later) — steps 4+5 done; lock release deliberately HELD

**Step 4 — parked 17, not 16** [MEASURED, RETURNING ids]: 10 audit_tool + 6 improve_tool +
1 `evaluate_tools` (`9062d146`, "Evaluate tool needs for site") → status 'deferred', reason
+ prior status in `result` (deferred_reason/deferred_by/deferred_from_status). The 17th is a
deliberate deviation from the handoff's 16: evaluate_tools MINTS tool items, so parking the
16 while leaving the minter live would just refill the queue. Un-park = status back to
'detected' for these ids (they are listed in this file's pre-park census, and in the UPDATE's
RETURNING — ids stable).

**Step 5 — all four backup layers, each verified:**
1. Deploy repo: tag `loancalc-pre-rebuild-20260811` at `b685b74091dd63ab480570a2b1f0b77d03bc7a63`,
   pushed to origin; tar `~/backups/loancalculator/loancalculator.co.uk_pre_rebuild_20260811.tar.gz`
   (34 files, 28 html, tools/ + guides/ present).
2. DB: `loancalc_bak_20260811_{pages,pc,sc,cc}` = 27/63/3/12 rows (single transaction).
3. Off-cluster: `~/backups/loancalculator/loancalc_bak_20260811.sql` — 4 CREATE + 4 COPY,
   per-COPY row counts re-counted from the dump itself = 12/27/63/3; `monthly-display`
   present ×3 (calculator markup really in the dump, not just DDL).
4. `take_site_snapshot` → `0d1b55f0-529b-40bf-89ad-3d9fa4c62e46`, verified pages=27,
   chrome_components=3, locked_captured=17 (mig 219 IS in). ⚠ revert re-applies CURRENT
   lock state, not the snapshot's — a revert after lock release restores content UNLOCKED.

**Step 6 prerequisite — pre-release lock state (the full 20):**
- page_components (17, all `permanent`):
  - `decompose_20260802_proven_calculators` (12): index/tool-3 `993eda99`,
    application-tracker/tool-1 `2d45a024`, car-finance/tool-2 `f74f84f2`,
    compare-loans/tool-2 `4a5c8690`, consolidation/tool-2 `d54dd48e`,
    credit-health-check/tool-2 `c284de0b`, damage-checker/tool-4 `12fa13a6`,
    interest-rate-stress-test/tool-2 `6c1a1ac5`, loan-vs-savings/tool-2 `10be4f71`,
    overpayment-calculator/tool-3 `3e6fb1d2`, settlement-calculator/tool-2 `ab24cd19`,
    standard-calc/tool-4 `e8ce9cc0` — that is 12 rows; the homepage prose-0 below makes 17
    with the css carriers.
  - `loancalculator_css_carrier_20260808` (4): compare-loans/prose-0 `fe158218`,
    consolidation/prose-0 `ea49f2ba`, credit-health-check/prose-0 `959e220c`,
    overpayment-calculator/prose-0 `b03e254d`.
  - `loancalculator_owner_approved_20260809` (1): index/prose-0 `9e7cbaa2`.
- site_components (3, all `permanent`, `loancalculator_authored_chrome_20260803`):
  header `3099f1ee`, head `4a754431`, footer `8375feb6`.

**Step 6 itself HELD — deliberate sequencing choice, recorded so it is not read as an
omission:** 9 content_rewrite + 12 page_rerender + needs_design/needs_composition sit
'detected' on this site. Releasing locks now, with the mission file not yet owner-approved
(step 7 requires showing it first), opens an unbounded window in which the ordinary queue
could rewrite the pure baseline ahead of the rebuild. "Release everything" (owner, 08-10)
is FOR the rebuild; the locks come off immediately before the 082 submission fires, after
the owner has seen the mission. The release SQL is two UPDATEs over exactly the 20 rows
above (locked_at/locked_by/lock_type/lock_expires_at → NULL); the pre-state is recoverable
from this note, `loancalc_bak_20260811_{pc,sc}`, and snapshot `0d1b55f0`.

### 2026-08-11 (same session, evening) — round 2 REVISE, round 3, and two discoveries that reordered the plan

**Discovery 1 — the roll was already done.** The 215 lane's release put `v1.0.1288` up at
17:13Z, built from `038211dd8`, whose ancestor set includes `7a066dba1` (my plumbing).
Probed on both replicas with near-miss controls (`siteUsesFlatURLs`=3 vs `siteUsesFlatURLt`=0;
`url_shape`=2 vs `url_shapf`=0). So step 2's build+roll never needed doing — and the
council-kill worry with it. The startup provenance line had scrolled on both pods; the
literal probe is what settled it (their commit b8457a9fc records the same two dead ends).

**Step 3 done:** `url_shape:"flat"` seeded and verified (see SEED file; my 26-vs-27
expectation error is annotated there — the guard caught it, first exercise of the
DO/RAISE-in-seed pattern on this lane).

**Discovery 2 — round 2 came back REVISE, and the gate was a landmine with MY key on it.**
Four seats cited the 2026-08-11 LANDMINES entry (added by the 215 lane's round 2, HOURS
before my submission): adoption supersede+INSERTs the structure aspect, dropping every
key it does not write — `url_shape` included, so a re-adoption would resurrect bug 241.
I never grepped LANDMINES for my own new key/symbols before submitting — logged in
WRONG_CALLS (the check answered round 1's question, not my plan's footprint).

**Round 3 (committed `19acfc895`, trail corr `70256656`):** carryForwardStructureSpecKeys
(ALL unknown keys, fresh wins, same tx, fails open ×3 — protects PLAN-048's gates too;
215 lane notified); wired apply_gap_plan/create_tool_component/deploy_tool-new-arm;
per-file exact-count contract pin with the blog-post exclusions stated. prior_art's
"symbol already exists" HIGH answered by timeline (git log -S: first appearance is
7a066dba1 — the landmine describes the new code, it does not predate it). Reader
consolidation parked on BLD-018 — the 215 lane's file, not mine to refactor mid-lane.
Round-3 code in NO image yet; probe carryForwardStructureSpecKeys at the next roll.

**Pattern-check advisories on 19acfc895, dispositioned:** "migration+code in one commit"
— the SEED is per-site data (already applied), not schema; point fix under RFC_022's
narrowed trigger. logged-model-output (adoption:269) and unrepaired-component-write
(create_tool_component, deploy_tool) — pre-existing code in files my hunks touched
elsewhere; bugs_open/136's list is the right home, not this lane.

### 2026-08-12 — the four decisions answered, and decision 3 turned out to be two changes

Owner: approve brief today · keep the pin · keep the 12 calculator locks AND fix the
planner to see them · park the rest.

**Decision 4 done — and a correction to my own count.** I told the owner "twenty-one
other jobs". Wrong: 21 was the two types I happened to name (9 content_rewrite +
12 page_rerender). The full `detected` set was **45** (add 16 page_component_status_drift,
2 needs_brand_head_assets, and one each of needs_composition/needs_design/needs_rerender/
hardcoded_section_colors/deactivated_component/capability_gap). All 45 parked; 64 deferred
in total now. **The shape: I reported a filtered count as if it were the census** — the
exact `a-filtered-count-can-ship-inside-a-denominator` failure, on my own reporting to
the owner. Corrected to him in the same breath.

**Decision 3 split into two changes once measured, and the second one is the big one.**
- Half 1 (DONE, `f4820a877`, council `a625c326`): `matchLockedRow` gains the component
  identity arm `matchDecisionProtectedRow` always had. Without it, fixing half 2 alone
  would produce TWO calculators per page — plan's copy inserted in place, locked
  original exiled to the foot.
- Half 2 (DESIGNED, NOT APPLIED — `PLAN_2026-08-12_planner_sees_locked_tools.md`):
  three traps stopped me doing it today. `content_components` has **no site_id** (the
  library is GLOBAL — 81 tools, so a naive widening offers every planner another
  site's calculators); **21 sites** already place tools, so even site-scoped it is a
  fleet behaviour change; and `query_database` **hard-errors** on a nil param path, so
  a bad site-id binding stops every site planning. Self-gating SQL + opt-in flag
  designed; param path must be read from real planner runs, not guessed.

**The thing that made me stop and read rather than assume:** `grep matchLockedRow`
turned up `save_sections_positional_tool_slot_test.go` — the loanandmortgagecalculator
lane had already settled this on 08-10, and their framing corrected mine. I had it as
"positional names never match"; they proved a positional name matches FINE when the
composition carries it, and that the trap is a composition that OMITS the slot — i.e.
**"seeding a site plan is the dangerous act; rerendering is not."** My lane is the
dangerous act. CONTRIB filed back to them.

**Verified rather than asserted:** their mutation guarantee still fails under their own
stated mutation after my change (applied + restored atomically, residue-checked). Their
fixtures keep an empty `component_id` deliberately so they stay on the branch they test
— giving them a matching id would have made them pass while silently testing nothing,
which is `declaring-a-key-silences-your-own-detector` in test form.

**090 filed BEFORE any code was written** (`69109208-7ae2-400e-a2b5-57b72003677b`) —
still `diagnosing` at end of session. Verdict to be recorded either way, including if
it refutes me; the code stands regardless as a convergence of two sibling guards, but
the justification would need rewriting.

### 2026-08-12 (evening) — the two false claims are DEAD on the live site

Owner: "go ahead as you suggest" → wait for the planner half, cut the claims now.

**The cut.** Deletions, not rewrites, so no hand-authored copy was added to a site whose
product is framework-built pages:
- footer (`site_components`, locked chrome): dropped " Every calculator on this site
  shows its own arithmetic." leaving "Independent UK borrowing tools."
- `guide-how-loans-are-calculated` prose-1: dropped " It shows the month-by-month
  breakdown, so you can see how much of each payment goes to interest and how much chips
  away at what you actually borrowed." leaving "…use our Main Loan Calculator. Like any
  calculator, it can only work from the figures you give it…"

Both edited in **`rendered_html` AND `content_data`** (the claim was in both — fixing only
the HTML would have let any future regeneration restore it), inside one transaction with
`DO`/`RAISE` guards that required the claim to be gone AND the surrounding prose to have
joined up. Repo source `chrome/footer.html` fixed in the same task (`3d49ae8c2`).
`rendered_html_digest` left NULL deliberately — stamping it would falsely assert
"reproducible from content_data via the render path" and would corrupt the purity
baseline (63 rows, digest NULL) that the rebuild audit uses as its control.

**Why re-ASSEMBLE and not re-RENDER.** `assemblePage` (`rerender_single_page_action.go:560`)
reads stored `rendered_html` from `page_components` and stored chrome from
`site_components` — no content regeneration, no LLM call — and
`rerender_single_page_action.go` performs **zero DB writes** (grepped: no
ExecContext/UPDATE/INSERT anywhere in the file). So the purity baseline survives. A
regeneration would have stamped digests and pulled in every framework change since these
pages last rendered, contaminating the very before/after the rebuild exists to produce.
Checked first that nothing else had drifted: exactly one page had components newer than
its deploy — the guide I had just edited.

**A wrong turn worth keeping.** I first hand-INSERTed 26 `page_rerender` rows. They sat
`detected` with `attempt_count 0` and were never selected: real ones carry an `item_key`
and a `batch_id` and arrive as `triaged`, because they are created BY the `rerender-pages`
agent, not by hand. Then they vanished from the table entirely — not cancelled, deleted —
and I could not find any Go path that deletes `site_work_items` (only ad-hoc scripts
scoped to other domains). **Unexplained; recorded as unexplained rather than guessed at.**
Everything else of mine survived (64 deferred, url_shape=flat, 20 locks, the claim fix),
so it was targeted at those 26 rows. If a later session finds the cause, it belongs here.

**Then I nearly re-fired a dispatch that had already worked** — see `WRONG_CALLS.md`
2026-08-12: my poll used `now() - interval '15 minutes'` after a session interruption, so
it missed the successful run by 47 minutes. Query by correlation, not by clock.

**Verified at the artefact, all 26 live pages** [MEASURED 2026-08-12]: false-claim
occurrences **0**; positive control "Independent UK borrowing tools" present on **26/26**
(so the fetches were real and the footer is intact); pages under 2000 bytes **0** (no
truncation/deploy-window blob). Calculators still carry their inline arithmetic
(4 `<script>` blocks per tool page; `id="monthly-display"` on the homepage).

**The 090 on the matchLockedRow mechanism FAILED — it did not verify anything.**
`Request … failed: step verdict failed: … API request failed with status 529 overloaded_error`
after 4 retries. Not a refutation: an outage (which also explains this evening's psql
timeouts). So per the owner ruling of 2026-07-31 I state the substitute plainly: the
mechanism rests on (1) first-hand reading of the deciding arms, (2) a mutation test I ran
and restored, and (3) the loanandmortgagecalculator lane's independent 08-10 test, which
reached the same conclusion from the other direction. Re-run the 090 when the API is
healthy.

### 2026-08-13 — plan re-review before half 2; two premises verified, one new puzzle; handoff cut

Owner: "look over the plan once more and then carry on" → then "fresh chassis deployed,
update docs, handoff if heavy". Context heavy → handoff written:
**HANDOFF_2026-08-13_planner_half_continue_here.md** is now the continue-point.

**Deploy verifications (both at the artefact, with controls):** v1.0.1294 (09:48Z) and
v1.0.1295 (13:53Z, built from `69612d692a4a…`) BOTH carry `f4820a877` (identity arm) and
`19acfc895` (adoption carry-forward) — revision-label ancestry + stamp in `/proc/1/exe`
(3 hits, off-by-one control 0) + `carryForwardStructureSpecKeys` literal (2, near-miss 0).
LANDMINES addendum corrected to LIVE; synced.

**Council a625c326 (identity arm): APPROVED** round 1 — had landed 08-12 13:53Z; the 529
overload had hidden it from my queries. 10 advisories, none high, dispositioned in
bugs_open/241 (notable: bug_historian correctly matched the mechanism to CLOSED 058 —
same failure resurfacing through the identity gap; recorded so 058 isn't reopened).

**Plan re-review verdicts:**
1. Identity chain SOUND, measured: 12/12 locked tool rows point at MASTER components,
   each function exactly one active row fleet-wide, and enrichSectionsWithComponentIDs
   resolves `WHERE function=$1 AND is_active` with NO component_level filter. So a
   planned tool section resolves to exactly the locked row's id. (This was the review's
   biggest open risk — a fork/master mismatch would have made half 1 dead code.)
2. `$ctx.` namespace has NO site_id (execution_context_params.go:70-79) — the param
   must come from collected_data.
3. **NEW PUZZLE**: zero `build-site-planner` orchestrations retained ALL-TIME, zero
   retained workflow_plans mentioning load_components — while site_plans keep being
   written (noted.co.uk 08-12). Suspect short retention on completed orchestration_states
   (noted's 08-13 rows are only page-rerender/build-dispatch-loop; its 08-12 planner run
   already gone). Resolution steps written into the PLAN file; `site_record.site_id` is
   the likeliest param key but MUST be confirmed on a live run (trap 3: nil path =
   fleet-wide planner outage).

Half 2 NOT built this session — deliberately: the param-path evidence the plan requires
does not exist yet, and guessing it is the one forbidden move.

### 2026-08-14 — the param-path puzzle RESOLVED; control run PASSED; the site was busy while we were away (all benign, one page added)

**The 08-13 puzzle is retention after all — but of a shape the 08-13 queries couldn't
see.** `orchestration_states` is a ~2-day working set: 2,082 rows for 08-14, 599 for
08-13, then single digits per day back to 07-13 (stragglers). Completed rows are purged
on that horizon, and the planner is RARE — 3 `needs_site_plan` items all-time, latest
08-12 03:22 (noted.co.uk, which matches its `site_plans` row at 03:22:51 to the second).
So "zero build-site-planner rows all-time" was true every time anyone looked, and will
almost always be true; nothing exotic about the execution path. The planner is spawned
by build-dispatch-loop as a work-item handler (`spawn_agent` with
`agent_type_field: current_item.handler_agent`; the item is created by
build-briefing-agent's `create_next_item`).

**The param path is SETTLED: `site_record.site_id`.** Three independent legs, none
inference-only:
1. *Code chain:* the planner workflow is strictly linear (`start_step` read_specs →
   ensure_site → load_existing_pages → load_components → …; every step exactly one
   `next_step`, no conditionals — verified per-step against the live definition).
   `EnsureSiteRecordAction` returns `"site_id"` on EVERY path that lets the workflow
   continue — the DB path (site_db_actions.go:194) and the DB-nil placeholder
   (site_db_actions.go:970) — and the coordinator stores that map at
   `collected_data["site_record"]` (coordinator.go:1877). If ensure_site errors, the
   run dies two steps before load_components ever binds a param.
2. *Same-workflow precedent:* `write_site_plan` already binds the SAME path — its step
   config is `"target_site_id": "site_record.site_id"`, resolved by ExtractActionInputs
   Strategy 0 (explicit dot-path against collected_data, action_inputs.go:609-631).
   Every `site_plans` row therefore certifies one real run where the path resolved:
   10 rows since 08-02, covering a fresh build (noted 08-12) and repeat replans
   (1fcfa4f3 ×3).
3. *QueryDatabaseAction mechanics:* params resolve via ExtractNestedField against
   collected_data (database_actions.go:47) with the initialize short-circuit BEFORE
   param resolution (line 14) — so the initialize pass can never nil-fail, and the
   `input_data.` prefix fallback is irrelevant to this path.
   `input_data.site_id` (the plan's other candidate) is REJECTED: it is only as good
   as the dispatcher's input_mapping, and ensure_site's own `extractDomainFromInput`
   "aggressive search" exists precisely because input shapes vary by caller.

**Control run PASSED, live DB [MEASURED 2026-08-14]:** old query vs new self-gating
query md5-identical over the full ordered row set (`0ceba482…`, 129 rows) for noted
(unflagged), loancalculator (unflagged pre-seed), and a nonexistent site id; executed
via PREPARE with an UNTYPED $1 — which also proves the text-parameter→uuid coercion the
Go driver relies on. The disconfirming result existed: a widened row set for any
unflagged case. Preview of the flag-ON branch for loancalculator: exactly 11 components
— which is CORRECT against 12 locked rows, because `tool-loan-repayment` is placed
twice (index tool-3 + the inactive tool-standard-calc page's tool-4). First read I
called that "12 vs 11, one calculator would be exiled" — wrong, the count difference is
a shared component, not a gap. (tool-credit-roadmap has NO tool row at all — prose-only
page, nothing to pair.)

**Meanwhile the site was NOT idle (all verified benign at the artefact):** the post-1295
fleet-wide rerender wave (the predicted one-off `stale_chrome` wave, 440 page_rerender
items / 16 sites today) reached loancalculator: 9 pages rerendered incl. three
calculator pages and the claims-cut guide, ~14 more triaged. Verified at the REAL urls
(see WRONG_CALLS 2026-08-14, the name-vs-url 404 trap): calculators intact (4 script
blocks each), false claims still 0 everywhere, locked footer serving on rerendered
pages. The three chrome overwrite attempts were BLOCKED by the locks
(lock_blocked_change 14:56–15:12, needs_human_review — expected, leave for fire-time
release). Two deltas that change recorded premises:
- **The site is 27 active pages now, not 26**: content-gap-planner autonomously
  created `guide-loan-faqs` (/guides/loan-faqs.html, serving, correct chrome, matches
  the nested URL convention). The mission draft's keep-list and "26/26" figures predate
  it. A content_rewrite also added copy to `legal`.
- **The purity baseline (digest NULL) is eroding BY DESIGN**: 58/66 components still
  NULL, and the queued rerenders will stamp more. The rebuild audit's before/after
  control must lean on the 08-11 backups (bak tables / off-cluster dump / snapshot
  0d1b55f0), not on live digest-NULL.

Next: migration + flag seed (definition change is live-on-apply; control run proves it
inert unflagged), council submission naming the RFC_022 fifth flag and the 21
tool-placing sites, then the canary replan.

### 2026-08-14 (late) — council round 1: REVISE, and the gating objection was right about my evidence, wrong about the world

Round 1 on `508fe8eb` came back REVISE (gating: prior_art_librarian HIGH; 5 abstained;
guidelines/adoption_guardian/render_guardian/constitution/mission/debug_historian/
reuse_agent approved). Dispositions, all answered in round 2 (`SUBMISSION_..._r2.json`,
run orch `ae9e0873`):

- **librarian HIGH (gating):** I asserted the re-adoption flag-drop closed via
  `19acfc895` without attaching proof, while multiple landmine pointers still said the
  opposite. The world was as I claimed — the central LANDMINES entry was ALREADY
  corrected ("Do not re-fix", council 70256656) and in sync with doc_notes — but the
  claim in MY submission was bare. Fixed with three attached layers: the code quote
  (all-unknown-keys carry), a FRESH artefact probe (both replicas, literal 2 hits each,
  near-miss 0, `merge-base --is-ancestor` true), and the corpus correction: PLAN-048's
  stale ⚠ struck through in place, with the five-key census now living THERE as the
  single authority (reuse seat's ask folded into the same edit).
- **editquality/guardian/debug_historian (version ambiguity):** answered by
  enumeration — exactly ONE build-site-planner row exists (their own check showed it),
  and the migration's count(*)=1 guard would have aborted otherwise.
- **guardian (rollback placeholder):** the round-1 SKETCH abbreviated; the committed
  FILE was always verbatim — and now verified byte-identical to the snapshot's captured
  pre-state. Learned in passing: **the two-arg `snapshot_agent(type, reason)` writes to
  `agent_definitions_backup`**, not agent_definitions — a same-table search for the
  snapshot comes back empty and reads as "snapshot lost".
- **bug_historian MEDIUM (menu ≠ placement):** accepted, not defended — the canary
  (`canary_replan_407.sh`) is the placement-level proof and runs only after the verdict.
  Their 17/12 locked-rows check reconciled against my 12/11: superset (all locks) vs
  tool subset; `ported-prose` is the shared 12th component.
- **architecture MEDIUM:** follow-up filed as `features_open/033` (structure-aspect
  opt-in key counter; the 402-405 RFC_022 counter covers action input specs only).

Commits: `f3658c893` (the change + PLAN-049), `1c05b2178` (round 2 + canary script +
033), both `Council-Submitted: 508fe8eb`. My PLAN-049 index row travelled as a
same-file passenger in another session's `99fa0a3fb` — stated in both commit messages.

### 2026-08-14 (night) — round 2 APPROVED; canary run: mechanism PROVEN, placement UNTESTABLE by a bare replan, and two surprises contained

**Council round 2: APPROVED** (2 advisories, none high; librarian flipped to approve;
3 abstained). Advisories dispositioned: 033 got its hard number (review at 8 keys,
`a212b3470`, first commit entitled to `Council-Reviewed:` — verdict read first); the
editquality bookkeeping notes were about artifacts already committed; guardian's
seed-race note is moot (the seed applied under its exactly-1-current-row guard);
debug_historian's separate-verify-file convention noted for next time.

**Canary (corr `b23b19c7`, orch `7d9d9b6d`, COMPLETED 19:14Z), verdict in three parts:**

1. **PROVEN, at the live artefact:** `available_components` = **140** in the run's own
   collected_data — the widened menu fired, `site_record.site_id` resolved (the run's
   site_record carries this site's id), no step failure anywhere in the chain. Locked
   rows: 17/17 untouched, `tool-2` on loan-vs-savings still 11,845 B with its script.
   The 28 pre-existing pages rows: **byte-identical** on (name,url,page_type,status) —
   md5 `e6dd8fb8…` recomputed equal after excluding the newcomers. Serving 27/27.

2. **UNTESTED — and now known to be untestable this way:** the placement question
   (bug_historian's round-1 objection). A bare replan writes `site_plan_sections`
   ONLY for pages it invents: every realised page carried `sections: []` into the
   plan (read from collected_data, not inferred — index 0, tool-loan-vs-savings 0,
   about 2). Seed 362's "re-emit realised lists verbatim" did not manifest as plan
   sections on this input shape. So no composition for a built page was proposed,
   nothing could exile a calculator, and equally nothing exercised the identity arm.
   **Placement proof moves to the first run that composes a built page** — the
   mission-driven rebuild itself, or a targeted `recompose_pages` dispatch (intent in
   BOTH spec list and briefing prose, per the recompose no-op landmine).

3. **SURPRISES, both contained:** the planner INVENTED two pages (`about`,
   `guides-index`) despite converging on the 27 realised ones, and emitted 7 follow-on
   items (2 needs_page, 4 needs_imagery incl. a site logo, 1 needs_rerender whose
   assemble step would have DEPLOYED the two not_built pages as empty shells). All 7
   deferred with a stated reason (matching the owner's 08-12 parking); the two
   component-less rows archived under a guard (zero components + created-today only;
   one UPDATE to reverse). The serving checker briefly read 27/29 with two 404s —
   the PLAN-047 active-but-never-deployed shape, live for ~6 minutes.

**What this means for the fire sequence — TWO OWNER-FACING QUESTIONS, surfaced not
decided:** (a) the rebuild's regeneration mechanism must be chosen explicitly: a plan
that carries no compositions for built pages will not, by itself, regenerate them —
if the rebuild is to rebuild the 26 pages it must go through recompose_pages (both
places) or per-page content work, and the calculators' placement test happens THERE,
with the locks and the identity arm as the safety net; (b) page invention under the
mission brief — the canary ran bare, so whether the keep-pages pin suppresses
invention is untested; the fire runbook should check for new active rows immediately
after the planner phase. Neither question blocks the 407 mechanism, which is done and
approved; both shape how the fire is dispatched and judged.

### 2026-08-14 (late night, new session) — owner answered all three questions; fire sequence resized by four mechanism findings

**The owner's answers (this session, asked directly):**
1. **Q1 regeneration:** explicit recompose, all 26 existing pages.
2. **Q2 pin:** trust the mission's keep-pages pin, WITH the immediate post-planner
   new-active-rows check — invention gets archived before it can build.
3. **The invented pages:** BOTH wanted. Un-archive `about` + `guides-index`; the
   rebuild fills them in.

**Four mechanism findings tonight (code + live config, evidence inline) that resize
HOW the fire happens — the handoff's single-dispatch step 3 is superseded:**

1. **082 has no spec plumbing, so recompose cannot ride the rebuild dispatch.** The
   `needs_site_plan` item is minted by build-briefing-agent's `create_work_item` with
   NO spec key (live `agent_definitions` row read tonight), and 082's input_data
   carries only domain/fidelity/email/mission_brief (script read). Injecting the spec
   between mint and dispatch is a race: the three all-history `needs_site_plan` rows
   lifecycle to complete in 3–19 min, so the mint→dispatch window is likely seconds.
   ⇒ **The fire is TWO-PHASE:** phase 1 = 082 rebuild (mission installs, plan
   converges, the two new pages build, chrome/nav regenerate); phase 2 = a direct
   build-site-planner dispatch (`canary_replan_407.sh` is the template) carrying
   `input_data.spec.recompose_pages` = the 26 pre-fire built pages, fired once
   phase 1 settles. The intent PROSE lives in the mission (the planner's `read_specs`
   feeds it briefing from site_specs), so the recompose no-op landmine's both-places
   rule is satisfied in phase 2.
2. **The reconciler routes tool-role pages to a human gate.**
   `reconcile_site_plan_action.go` decision rule 3: tool/game role OR
   `rebuild_policy='owned'` → `owned_page_review` (needs_human_review, NO handler) —
   deliberate, the generic builder clobbers owned pages (TP-004/TL-001). Live: all 27
   pages are `rebuild_policy='generic'`, but the current plan carries **11
   role='tool' pages**. ⇒ phase 2 regenerates ~15 pages automatically and parks the
   11 tool pages' REBUILDS at human review. The plan-level placement test — does the
   planner keep the calculators in composition when free to recompose? — still fires
   for all 26. The 11 review items are EXPECTED OUTPUT, not a failure.
3. **`deferred` is open to both the dedup index AND the reconciler** (idx_swi_dedup
   terminal set read from pg_indexes; reconciler `NOT IN (...,'cancelled')` at :559).
   ⇒ the canary's deferred `needs_page:about`/`needs_page:guides-index` are the ONLY
   build path for the two pages — a fresh mint is dedup-blocked and the reconciler
   skips-as-queued. They must be un-deferred (status→'detected'), **but only AFTER
   phase 1's plan lands**: the build-pipeline scheduler ticks ~every minute
   (`sched-build-pipeline-trigger` observed 21:39Z with pending_sites=11), so
   un-deferring pre-fire would build both pages from the CANARY's bare plan and the
   old briefing.
4. **Never-shipped active pages are composable.** `realisedPageHasShipped`
   (v3_site_actions.go:6611): "a page that has never shipped is merely uncomposed" —
   the preserve guard does not snap it. ⇒ un-archiving the two rows to 'active' is
   sufficient; the planner can compose them.

**Also verified tonight:** `ensure_site_record` is create-or-find — resubmitting an
existing domain re-drives the pipeline on the SAME site row (no duplicate) · a THIRD
archived page exists, `tool-standard-calc` (archived 08-03, unrelated) — it stays
archived; only the two 08-14 rows get restored · fleet rolled **v1.0.1300** tonight
(both chassis pods started 20:36Z, AFTER the handoff was cut); serving re-verified
27/27 clean post-roll · the `RECOMPOSE_INTENT_NOT_REALISED` tell (`4e3c96e64`,
08-10) is an ancestor of 1295's build commit `69612d692a4a` (merge-base check) —
[INFERRED for 1300 by tag ordering on a forward-only tree; phase 2's primary check
is the direct SQL comparison of plan sections vs pre-fire realised compositions, so
nothing rests on this inference] · site quiet: open items are 13 lock_blocked_change
+ 6 content_rewrite (all needs_human_review) + 1 blocked page_rerender (the
bugs_open/189 verification item of 08-06) — nothing is actively working the site.

**Un-defer set is THREE items, not seven.** Of the canary's 4 deferred needs_imagery,
3 target the site or existing pages (site logo `003b98cc`, guide-loan-faqs hero
`15d6323f`, index hero_home `34b3ace1`) — they stay parked per owner 08-12. Un-defer
= `needs_page:about` (`222ecf94`), `needs_page:guides-index` (`a52e59d8`),
`needs_imagery hero_about` (`ad289c0e`). The stale `reconcile_rerender:8d7c…`
(`1fcb4772`) stays deferred permanently — it is keyed on the canary's plan and its
assemble would run against a superseded state; phase 1's reconciler emits its own.

**Fire sequence as being executed (supersedes handoff steps 2–5):**
1. Un-archive `about` (`b3f03e83`) + `guides-index` (`e31c71a8`) → 'active'.
2. Release the 8 non-calculator locks (page_components `fe158218` `ea49f2ba`
   `959e220c` `b03e254d` `9e7cbaa2`; site_components `3099f1ee` `4a754431`
   `8375feb6`). The 12 calculator locks STAY.
3. Fire 082 with `MISSION_2026-08-14_fire.txt` (mission draft + two dated additions:
   the page set now names about + guides index and pins "otherwise add no new
   pages"; a rebuild-scoped recompose-intent paragraph). Save CORR.
4. Post-planner checkpoint: Q2 new-active-rows check
   (`created_at > <fire time>` — the two restored rows are 19:14Z, they do not
   trip it); un-defer the three items; watch the two pages build.
5. Phase 1 settle: serving 29/29, then phase 2 recompose dispatch (26 pages).
6. Post-recompose: RECOMPOSE tell query + plan-sections-vs-realised SQL compare;
   expect ~15 needs_page + 11 owned_page_review + 1 needs_rerender; the 11 review
   items go to the owner.
7. Verify per handoff step 5 (purity of the 12 locked rows vs 08-11 backups,
   URL diff EMPTY on the 27, toolgolden 11/11, calculators in place).

**PHASE 1 FIRED 2026-08-14 ~21:55Z.**
> **CORRECTED 2026-08-15: the true fire time is `2026-08-15 07:54:41Z` — the DB's
> clock, which is the arbiter for every created_at comparison. The local machine
> clock is ~10h behind the cluster; the submitter's own spec rows caught it. Q2
> baseline tightened to `created_at > '2026-08-15 07:54:00Z'`. Same session, two
> further corrections: (1) my added mission sentence said the homepage keeps its
> "credit roadmap tool" — the homepage's locked row is `tool-loan-repayment`;
> caught minutes post-fire by the pre-fire baseline snapshot, corrected in the
> mission FILE and in the two spec rows that carried it (`mission_brief`,
> `submission`) BEFORE any downstream reader — the classifier's four output specs
> verified clean of the clause. Both in WRONG_CALLS.md 2026-08-15. (2) The
> pipeline is FAST tonight: classifier claimed the item 27s after mint and its
> outputs landed within ~3 min — the ~30-min publish→start budget did not apply.**
`CORR=2d950ecc-4919-441b-a4fb-e6aa47663ad9`,
printed orch `d03733c0` (find the run by CORRELATION, never the printed id). Q2
baseline for the new-active-rows check: `created_at > '2026-08-14 21:52:30Z'` (the
two restored rows are 19:14Z — they do not trip it). Pre-fire state at the moment of
dispatch: pages 29 active (27 built + 2 restored not-built), locks 12/12 calculators
only (8 released in one transaction, counts verified 17→12 pc / 3→0 sc), serving
27/27 clean, site queue quiet. Two schema-guess errors on the release SQL
(`section_id`, `component_type` — both aborted whole transactions cleanly before any
change persisted; "\d before SQL" exists for a reason and I skipped it twice).

**2026-08-15 ~09:10Z — THE PIPELINE STOPPED AT THE B2 GATE, BY DESIGN; manually
re-driven at the briefing seam.** Research 07:59 → vertical 08:27 → strategy 08:31,
then NOTHING: the strategist's `gate_next_item` (`site_state.is_deployed == true` →
complete WITHOUT chaining; migration 341, 2026-08-08) correctly refuses to enqueue
the briefing→plan rebuild chain on a site with deployed pages. Our rebuild is
exactly the wanted re-plan the gate cannot know about. Precedent found before
acting: webdesign.uk re-drove this same seam on 08-09 (`manual-redrive-2026-08-09-*`
rows, both complete). Filed the withheld item verbatim to the strategist's own
config shape: `needs_briefing`, key `briefing_loancalculator.co.uk`, handler
build-briefing-agent, priority 10, spec {}, source
`manual-redrive-2026-08-15-post-b2-gate`, id `00bd5eff`. The briefing agent's
final step mints `needs_site_plan` itself (read from its live workflow row), so the
designed flow resumes one seam downstream of the gate. **Consequence for anyone
re-running this fire on a deployed site: 082 fresh-mode alone will NEVER re-plan a
deployed site — the manual `needs_briefing` redrive is a REQUIRED step, not a
workaround.** Also noted: this site has ZERO all-history needs_briefing/
needs_site_plan items — adopted sites never walked this leg; tonight is its first.

**2026-08-15 ~10:55Z — the redrive item then sat UNCLAIMED for ~1h45m, and the
reason is a second finding: a hand-INSERTed work item must be born (or moved)
`triaged`, because the polling dispatcher cannot see `detected`.** Evidence chain:
the build-pipeline-trigger IS firing (~every 90s, kafka-scheduler logs, pending 2–3
sites) and its `find_dispatchable_site` query takes ONLY `status IN ('triaged',
'approved')`. The morning chain's items were claimed in ~30s NOT via that poller —
they were dispatched by the workflows that created them (`claimed_by=
build-dispatch-loop`, chain-created); a hand INSERT gets no such kick and waits for
the fleet triage sweep, whose `detected` backlog currently reaches back to 08-13
("detection works; SCHEDULE and DISPATCH do not", again). Moved `00bd5eff` to
`triaged` (guarded UPDATE, attempts 0/3); site-level `sites.locked_at` verified
NULL (the trigger also filters on it). **RUNBOOK rule for this seam: manual
redrive = INSERT with `status='triaged'`, not 'detected'.**

**Fleet roll v1.0.1301 landed 10:14Z** (owner-run; both chassis pods restarted).
Nothing of ours was in flight (the briefing item was still parked at detected);
the ~300s dispatch quiet window was long past before the item went dispatchable.

### 2026-08-15 ~11:05Z — PLAN LANDED (34b1b056, 10:59:33Z); checkpoint run; PHASE 1 EMITTED THE FULL REBUILD SET, resizing phase 2 to probably-unnecessary

Checkpoint results (script in this dir, three schema-guess fixes on the way —
`sps.ordering` not position, `agent_error_log.occurred_at` not created_at):
- **Q2 INVENTION: ZERO new page rows.** The keep-pages pin held under the mission
  brief — the canary's invention did not recur. Q2's answer is now MEASURED.
- **Plan: exactly 29 pages** (14 guide + 11 tool + 2 content + 1 section-index +
  1 landing). Identity md5 of the 27 pre-fire pages UNCHANGED (`e6dd8fb8…`).
- **`about` composed (hero,about-content); `guides-index` sectionless** — a
  section-index renders through the lister subsystem, sectionless is its shape.
- **Built pages again sectionless in the plan** (only guide-loan-faqs also carries
  sections: hero,faq,tool-cta) — BUT unlike the canary, the reconciler THIS TIME
  emitted the full rebuild set: **15 needs_page (all non-tool built pages,
  guide-can-i-overpay claimed within seconds) + 11 owned_page_review (every
  tool-role page — the TP-004 human gate arrived in PHASE 1, not phase 2) + 1
  needs_rerender keyed on the new plan.** [UNRESOLVED mechanism: why the canary
  skipped built pages and this run emitted for them — suspect sync_pages wrote
  the plan's empty sections onto pages.sections, flipping the reconciler's
  same-composition test; not chased tonight, the operative state is unambiguous.]
- Un-defer trio moved to detected (about, guides-index, hero_about). Calc locks
  12/12. RECOMPOSE tell rows: none (correct — no spec in phase 1).
- **PHASE 2 ON HOLD, probably unnecessary:** its purpose was to force build items
  for built pages; they exist. What it would still add — plan-level composition
  records + the RECOMPOSE tell — matters less than the REAL test now running:
  does a build-time composition keep the calculators? **The placement test has
  moved to BUILD time, ungated on `index`** (role landing → needs_page → rebuilds
  around locked `tool-loan-repayment`; identity arm + lock are the floor).
- **DESIGN REFRESH IS DEDUP-BLOCKED (new decision for the owner):** the 08-12
  owner-parked items hold site-wide keys — `needs_design`, `needs_composition`,
  `needs_brand_head_assets:favicon/og_card`, all deferred — so the planner's
  emit_design produced nothing (dedup). The 8 chrome/css locks are RELEASED but
  nothing will regenerate chrome/styles until those keys free. Un-deferring
  reverses an explicit owner parking, so it is D3-EXPANDED in the handoff, not
  a session action.

### 2026-08-15 ~14:30Z — build wave verified (14/15 + deploy), the homepage escalation decoded, the detected-trap repeated on my own trio, and PHASE 2 FIRED scoped to the 12 tool-carrying pages

**Build wave: 14 of 15 needs_page COMPLETE, zero failures, and VERIFIED AT THE
ARTEFACT, not the status:** sampled pages carry all-fresh `page_components` rows
post-plan (guide-can-i-overpay 3/3 touched 11:04, guide-jargon-buster 3/3 11:47,
legal 2/2 11:59); `reconcile_rerender:34b1b056` complete; serving **27/29** — the
two fails are exactly the not-yet-built about + guides-index (expected). Calc locks
12/12 throughout.

**The homepage did NOT rebuild, and the mechanism is now decoded** (orch `936f7ff9`,
completed_by_step `mark_no_ready_sections`, 11:57Z): `spec_sections: count 0,
source "none"` — the landing-page builder REQUIRES plan compositions and plan
34b1b056 carries none for index, so it escalated `needs_page:index` to
needs_human_review and touched NOTHING (index components 0/5 changed since 08-14 —
conservative and correct; the locked calculator never even came into play).
**Guide-type builders self-compose; landing/tool builders build only from plan
sections.** That is why the 14 guides+legal rebuilt and index+the 11 tool pages
cannot — and it reverses the phase-2 hold: **phase 2 is the mechanism that
supplies those compositions.**

**The detected-trap, repeated by ME, ~3 hours lost:** the checkpoint script's
un-defer wrote `status='detected'` — authored BEFORE the born-triaged rule was
discovered, and never re-checked after; the handoff note I added even generalised
it away. The trio sat invisible 11:05→14:21 while every sibling completed. Moved
to `triaged` 14:21 (script + handoff corrected; WRONG_CALLS entry — "a new rule's
first enforcement target is the tooling you wrote before you knew it").

**PHASE 2 FIRED 14:23:32Z (DB clock), corr `2f74a975-1a87-40a8-af88-a9bd2ecc1510`,
SCOPE REVISED 26→12** (index + the 11 tool pages; recorded in the script header):
the 14 guides+legal are already regenerated — recomposing them would only churn.
Expected outcome: plan compositions for the 12 (placement gate active,
plan_includes_tools seeded), RECOMPOSE tells for any no-op, NO new work items
(the reconciler skips-as-queued — needs_page:index at review + 11
owned_page_review are open and become the vehicles that realise the new
compositions). Phase 2 mutates NO live page: tool rebuilds stay human-gated
(D2), index waits at review until re-driven. Pre-state at fire: 30 rows/29
active, identity md5 `fd2c09c2…` (differs from `e6dd8fb8…` ONLY by the two
restored rows' status — expected), locks 12, plan 34b1b056.

### 2026-08-15 ~14:45Z — PHASE 2 JUDGED: the placement question is ANSWERED and the answer is 0/12; diagnosis filed

Plan `dcbae4df` landed 14:25:49 (2 min from dispatch). Zero RECOMPOSE tells
(correct — every released page got a genuinely new composition, nothing
no-opped). Locks 12/12. **THE LANE'S HEADLINE MEASUREMENT: the planner, with the
widened menu demonstrably in context (`available_components` = **151** in the
run's own collected_data), placed ZERO of the 12 locked tool component functions
in the recomposed compositions.** index → `hero,info-card-grid,tool-list,
guide-list,call-to-action` (no `tool-loan-repayment`); the 11 tool pages →
`hero,ported-prose,faq,tool-cta` (`tool-cta` is a CTA section, not the
calculator); `tool-credit-roadmap` alone carries `loans-credit-health-check`,
which matches NO locked function (nearest: `tool-credit-health-check`, a
DIFFERENT page's tool). So: **407's menu mechanism works; the planner's CHOICES
do not place tools.**
> **CORRECTED ~17:30Z same day — the planner's choices were RIGHT; the claim
> above inverts the mechanism.** The 090 loop refuted the attribution
> (`response_has_tool1=true`) and its named next_scope, walked first-hand,
> closed it: the raw plan_site response (`llm_call_log` `ca3c22f4…`) proposes
> `tool-loan-repayment` as index's SECOND section and the right `tool-<fn>` on
> the tool pages — then `loadComponentNameResolver`
> (v3_site_actions.go:3804-3809) whitelists ONLY
> `component_level IN ('section','element')`, so every proposed tool section
> resolves to nothing and validate drops it from the write. **407 widened the
> menu; nobody widened the resolver.** Filed as `bugs_open/282` with the full
> cited chain and ranked fix candidates. What caught it: the diagnosis loop's
> refutation plus reading the actual response text instead of trusting the
> persisted plan — the artefact one table upstream. The protective stack (tool-role review gate + 12 permanent
locks + identity arm) is currently the only thing between this plan and
calculator-less pages — all of it held, nothing live changed.
**⇒ D2 WARNING, prominent: do NOT realise the 11 owned_page_review /
needs_page:index tickets by naively applying plan `dcbae4df` compositions — they
would compose the calculators OUT (the lock floor should keep the rows, but that
is exactly the untested arm).** Diagnosis filed via 090 — RUN_CORRELATION
`4a02a4e1-3972-450a-8163-28d6bb0a79fd` (queue checked empty first); symptom
points at seed 362's converge arm + 407's placement wording vs the plan/locked
rows. Await the verdict before any prompt surgery.

**The rest of the phase-2 cascade, decoded:** 15 needs_page (14 guides + legal)
are STAMP-CONVERGENCE churn, not damage — the guides were BUILT this morning as
`hero,faq,tool-cta`, the new plan says the same, but they were stamped at plan
34b1 (which had no sections for them) so the version-compare emits; this build
re-stamps at `dcbae4df` and the loop ends. Let run. 16 needs_imagery = per-page
heroes for the rebuilt pages (different keys from the 3 owner-parked ones, so
dedup does not block them) — images will generate and land via rerender; this is
the pipeline's normal behaviour under the mission, distinct from the parked D3
set. 1 needs_rerender keyed on the new plan. Rebuild watcher re-armed over the
whole open set.

**15:37Z — one build-wave failure, dispositioned NO-ACTION:**
`needs_page:guide-how-loans-are-calculated` FAILED (attempt 1/3) at the
`complete_workflow` step — "failed to send response: Kafka write errors" — i.e.
AFTER the work: all 3 of the page's component rows are fresh at 15:33:22
(artefact checked, not the status). The known response-write flake class
(cf. spawn-call handshake races; "never cancel the failing row pre-diagnosis").
Not re-triaged — a retry would rebuild a healthy page a third time to repair
bookkeeping. If a future replan re-emits this one page for stamp reasons, that
single extra cycle is the accepted cost.

## 2026-08-16 — bug 282 is FIXED AND COMMITTED (from the bugfix_282 lane, not this one)

The blocker your D2 sequence names is done, and the sequence can resume once the
next chassis image rolls. Recording it here because the fix landed in another
lane and nothing in this directory would otherwise say so.

- Commit `5534e9f71`. Go half: `component_name_resolver_menu.go` (new) + two
  small hunks in `v3_site_actions.go` — **INERT until an image carries it**.
  Config half: migration `439`, **applied and recorded 2026-08-16**, setting
  `menu_field: "available_components"` on `build-site-planner.validate_plan`.
  Council submitted, corr `bbf49822-6704-4802-b3b5-1afed6777c88` (advisory).
- **What it does:** validate now accepts the component names its own planner
  step was OFFERED, instead of re-checking them against a hardcoded
  section/element whitelist. Your calculators stop being deleted between the
  LLM's response and `site_plan_sections`.
- **Your sequence is unchanged** and still correct: roll → re-fire
  `phase2_recompose_26.sh` (12-page scope) → verify the 12 locked tool functions
  in `site_plan_sections` on their own pages → then work the 11
  `owned_page_review` + `needs_page:index`. Baseline for the comparison is still
  plan `dcbae4df` (0 of 12).
- **Two things to expect that the old baseline will not show.** (1) A new log
  line, `"ValidateSitePlanAction: section resolved via the planner's menu"`, one
  per accepted tool section — that is the tell the arm fired; its absence with
  tools present in the plan means the image predates the fix, so check the build
  provenance before re-diagnosing. (2) LOCK-008 (`7d9b7334a`, live on v1.0.1304)
  now merges locked rows into the assembled list, so `pages.sections` on the 12
  tool pages carries positional slots — **verify at `site_plan_sections`, not at
  `pages.sections`**, exactly as your handoff says.
- **A correction to 282's ADDENDUM that touches this site's read of events.**
  `loans-credit-health-check` on `tool-credit-roadmap` was recorded as "a name
  matching NO component at all", evidence of a second drop branch. It is an
  ordinary **section-level** component (`824e3309…`, created 2026-08-13 for
  loanandmortgagecalculator.co.uk), which is why it survived validate. The
  `needs_new_component` item came later, from `plan_sections`, on a different
  path. There is no second branch; the account in step 3 of the bug file was
  complete. Full correction in the bug file and `WRONG_CALLS.md`.

## 2026-08-17 — the D2 sequence RESUMED and its headline measurement is 11/11: the calculators are back in the plan

Picked up `HANDOFF_2026-08-15_fire_in_flight_continue_here.md`. Its blocker
(`bugs_open/282`) was fixed by the bugfix_282 lane on 08-16; this session verified
the fix is live, ran the owed toolgolden battery, re-fired phase 2, and judged it.

### 1. The 282 fix IS live — and the check I reached for first was the WRONG one

[MEASURED] Running chassis is `v1.0.1305`, deployed digest
`sha256:f90a7e88…`, and the local docker image for that tag carries the same
RepoDigest, so its OCI label is authoritative for the running pod:
`org.opencontainers.image.revision = 6a782274b626c9f4977c9246d905deebb097cb1f`.
Positive control run in the same breath — that sha IS present in `/proc/1/exe`
on `agent-chassis-5657f446c7-q7b82`, so the label and the binary agree.
`git merge-base --is-ancestor 5534e9f71 6a782274b` -> **YES**. Config half also
live: `menu_field = available_components` on `build-site-planner.validate_plan`
(migration 439).

> **MY OWN WRONG TURN, worth the space because it looked rigorous.** My first
> probe grepped `/proc/1/exe` for the 282 fix's own sha (`5534e9f71…`) and got
> "absent" — and my negative control was also absent, so nothing flagged it. But
> **the binary carries exactly ONE sha, its own build commit, never its
> ancestors**, so that probe answers "was the image built AT my commit", which is
> almost never the question. Absent was the *expected* reading for a fix that
> shipped four commits later. CLAUDE.md already states the right test
> ("`git merge-base --is-ancestor <your-commit> <the stamp>`") and I reached past
> it for a grep. The tell I should have noticed: a probe whose negative control
> and subject both come back absent has not discriminated anything.
> Cheap check that fixes it: get the STAMP first (label or log line), then ask git.

### 2. The handoff named the WRONG golden file, and the name is a trap

The handoff's first step says the golden is
`acceptance/GOLDEN_2026-07-31b_tool_values.json`. **That file is the discarded
"attempt 1" baseline from 07-31**, not a golden. Evidence, from the file itself:
it holds **12** pages (including `tool-standard-calc`, archived since 08-03), its
`pressed` field is a bare STRING in a schema the current harness no longer emits
(now `{label, sel}`), and that string is `mobile-menu-btn` — the nav hamburger —
on **9 of 12** pages. That is exactly the failure this file's own NOTES entry of
07-31 describes: the press had moved off the tool and onto site furniture. It is
also **untracked in git** (`??`), so it exists only in this working tree; its
08-12 mtime is a copy or touch, not a re-capture.

The live baseline is the committed `GOLDEN_2026-08-08_voice_h_complete.json`
(11 pages, current schema, taken after the 08-08 compare passed). Corrected in
the handoff in place; WRONG_CALLS row added. **Cheap check that would have caught
it in one command: open the golden and count its pages.**

### 3. Toolgolden: exit 1, and the arithmetic is nevertheless UNCHANGED

`--compare` against the 08-08 golden exits 1 and prints "11 of 11 tools diverged".
That headline is true and misleading, so the useful reading is the breakdown
(diffed shared keys myself rather than trusting the summary):

```
shared fields compared, IDENTICAL : 1340
shared fields that CHANGED        : 0
ids only in golden (removed)      : 176   -> 2 distinct: nav-links-menu, mobile-menu-btn
ids only in capture (added)       :  80   -> 1 distinct: d91e7be1-… (the new FAQ block)
vectors where DIFFERENT inputs were driven : 0
```

[MEASURED] **Not one shared field changed value, and `drove` is identical in every
vector**, so the comparison really is apples-to-apples — the pages' own numeric
defaults are untouched, which is the thing that would silently invalidate it (the
toolgolden landmine). Every "divergence" is the rebuilt chrome: the old
hand-built nav's two id-bearing elements are gone and a FAQ section arrived. **No
tool's own field was added, removed or changed.** So the handoff's expected
"11/11" holds on arithmetic; only the exit code disagrees.

Also a real check in its own right: the harness **wrote** the new capture, and it
refuses to write when any tool is inert or input-independent. `vary=0` appears on
exactly the three tools the RUNBOOK says are exempt by construction
(application-tracker, credit-health-check, damage-checker).

Re-baselined as `acceptance/GOLDEN_2026-08-17_post_rebuild_tool_values.json`,
compare run BEFORE the capture, deliberately, per the 08-08 precedent.

### 4. NAV: the calculators are gone from the header, and Guides is stuck

[MEASURED] Served header on every tool page is now: logo, `Home`, `About`, and a
"Get Started" CTA to application-tracker. The 08-08 golden records what it
replaced — a `Tools` dropdown listing nine calculators (`nav-links-menu`).

- **The tool pages' absence is BY DESIGN, twice over.** `classifyPagesForNav`
  (`populate_nav_tables_action.go`) bars `page_type='tool'` from primary nav, and
  bars any URL under `/tools/` as a child page, "the parent /tools.html … represents
  them in navigation". All 11 tool rows are `in_header=false`. **But this site has
  no tools listing page** — only `guides-index`. So the framework's stated
  representative for the calculators does not exist here. Owner question, below.
- **`guides-index` is NOT barred** — it is `page_type='section-index'`, and the code
  has an explicit exception for exactly that case (`isSectionIndexType`), and
  `navPageScopeSQL` admits `status='active'`. Accordingly `site_nav_items` DOES
  carry `Guides -> /guides/index.html` at primary position 2, written
  2026-08-15 14:25:52 by the phase-2 planner. It is the SERVED chrome that omits
  it. [INFERRED, not chased] the chrome the pages carry predates or ignores that
  nav row; `pages.rendered_header` is NULL on index/about/tool-settlement, so the
  header is assembled from a site-level component, not per page.
- Calculators are still reachable: each tool page cross-links 8 of them in-body,
  and index links 8. Not orphaned, but not navigable from the header.

### 5. Serving state, unchanged from the handoff

28/29 active pages return 200; `guides-index` (`/guides/index.html`) is the only
404, still `build_status='planned'`. `tool-credit-roadmap` serves 200 despite
`build_status='needs_rebuild'`, so the cross-links to it are not dead.

### 6. PHASE 2 RE-FIRED — and this is the measurement the lane existed for

Fired `phase2_recompose_26.sh` (12-page scope, unchanged) at DB time
**2026-08-17 11:43:34Z**, corr `3584d962-d3de-415b-a468-64afab126534`,
orchestration `627ff71a-1af2-401b-a961-16a181916e71`. Pre-flight: no in-flight
orchestration for the site, pods 13h old (no 300s window), local clock within 2s
of the DB clock. New plan **`9463e31d-ee50-482e-94a9-7e186ef25543`** landed
11:46:22Z and is current.

[MEASURED] **11 of 11 locked calculators are now placed on their own page** —
against a baseline of **0 of 11** in plan `dcbae4df`. Verified as a join of the
locked rows' `content_components.function` against `site_plan_sections` on the
current plan, per page, not by eye:

```
index                          tool-loan-repayment          t
tool-application-tracker       tool-application-tracker     t
tool-car-finance-calculator    tool-car-finance-pcp-hp      t
tool-compare-loans             tool-compare-loan-offers     t
tool-consolidation             tool-consolidation-risk      t
tool-credit-health-check       tool-credit-health-check     t
tool-damage-checker            tool-return-damage-checker   t
tool-interest-rate-stress-test tool-rate-stress-test        t
tool-loan-vs-savings           tool-loan-vs-savings         t
tool-overpayment-calculator    tool-overpayment-impact      t
tool-settlement-calculator     tool-early-settlement        t
```

So `bugs_open/282` is not merely live, it is **PROVEN on the motivating case**:
the same script, same scope, same planner, one image later, 0/11 -> 11/11.
Supporting readings: **0** `RECOMPOSE_INTENT_NOT_REALISED` rows (every page
genuinely recomposed, no no-ops), locks **12/12** held.

⚠ **The script's own judge query #4 cannot show this.** It counts
`component_name LIKE 'tool-%'`, which also matches the `tool-cta` and `tool-list`
SECTION components — it returned **26** on the 0/11 baseline and would return 26
again either way. A query that gives the same answer whichever way the world is
must not be the test. Use the locked-function join above; runbook updated.

### 7. One placement to flag, and the wave

`tool-credit-roadmap` was given `tool-credit-health-check` — **another page's
calculator**, and the second copy of that function in the plan. That page has no
locked row of its own (it is a rebuild-era page), so nothing was displaced, but a
duplicated calculator across two pages is a content decision, not a win. Flagged
for the owner, not actioned.

Reconciler emitted, all born `triaged`: 15 `needs_page` (the 14 guides —
the same stamp-convergence churn as 08-15 — plus `needs_page:index`, "stale"),
5 `needs_imagery` (guides-index hero, compare-loans hero, 3 index icons), 1
`needs_rerender` keyed on the new plan. **No `owned_page_review` this time**: the
11 tool pages' review items from 08-15 are still open, so the reconciler
skipped-as-queued exactly as the script predicted.

**On letting `needs_page:index` run.** The handoff had it "STILL HELD" pending
282. Two things changed that: 282 is fixed and verified above, and the item the
handoff describes as held is in fact `complete` since 2026-08-16 08:45
(`build-dispatch-loop`) — index was auto-rebuilt a day ago, on an image WITHOUT
the 282 fix, and the locked calculator survived it (locks 12/12, and toolgolden
finds index's fields byte-identical). This new item is a fresh mint from the
reconciler, born triaged by the framework, on a `landing` page — TP-004's human
gate keys on tool-ROLE pages, which is why the 11 got `owned_page_review` and
index did not. So the floor has already held once unaided, and this rebuild is
strictly better informed than that one (the calculator is now IN the plan).
Let it run; verifying at the artefact afterwards with the fresh golden.

### 8. The build wave did NOT run, and the reason is `bugs_open/243` — already found by three other lanes today

All 21 items sat `triaged` and unclaimed. What I established before finding the
prior art (recorded because the re-derivation is itself evidence the tell is well
hidden, and it cost ~20 minutes):

- `build-pipeline-trigger` fires every 60s as configured; `build-dispatch-loop`
  pods spawn continuously (7 alive) and every run finishes **COMPLETED**, `error`
  NULL. 69 orchestrations completed fleet-wide in 25 minutes.
- The loop **does** load the item — `pending.has_items = true`, the full row
  present. So this is NOT `bugs_closed/176`'s selector/loader disagreement, which
  is what I first assumed; I had to correct that mid-investigation.
- The honest tell is one field: `collected_data->'claim_result'` =
  `{"claimed": false, "reason": "ai_endpoint_unavailable", "endpoint": "https://api.anthropic.com/v1/messages"}`.
- `claim_work_item_action.go` (~:218) reads `SELECT healthy FROM ai_endpoint_health`
  for the handler's endpoint and, on false, **releases the claim and puts the item
  back to `triaged`**. Fleet-wide, silently.
- [MEASURED] The `claude` row: `healthy=f`, `last_checked 11:09:53Z`,
  `last_healthy 11:07:15Z`, `check_interval_seconds 3600`, error = the API's own
  `status 400 … "You have reached your specified API usage limits. You will
  regain access on 2026-09-01 at 00:00 UTC."`
- [MEASURED] And the row is **stale**: `provider='anthropic'` calls are succeeding
  continuously right now — `claude-sonnet-5` and `claude-opus-4-6`, real
  `output_tokens`, 1–3 per minute, latest 11:58:08Z, from `council-gate` and
  `landmine-verifier`. The endpoint is serving while the instrument says it is not.
- Because `find_dispatchable_site` is `ORDER BY wi.created_at ASC … LIMIT 1` across
  ALL sites, the whole fleet queues behind the oldest unclaimable item. Six sites
  were behind it; loancalculator's items (11:46:36) were sixth.

**PRIOR ART — grep first, which I did only after diagnosing.** This is
`bugs_open/243` (the cap) / `bugs_open/244` (the spend), and **two** LANDMINES
entries already carry the whole mechanism, one titled "The BUILD QUEUE can be
fully stopped while every liveness check says the fleet is healthy … that row is
re-probed ONCE AN HOUR", filed today by the webdesign_uk lane, plus a 11:55Z
correction from the loanandmortgagecalculator lane establishing exactly what I
re-measured — that live traffic recovers within minutes while the row stays red.
Three lanes found this today before me. **Nothing filed; nothing to file.**

**What this lane can use, and is not in those entries: the clear time is
computable, not a guess.** `last_checked + check_interval_seconds` = **12:09:53Z**
is when the prober next becomes due (`check_endpoint_health_action.go` :87-95
gates on exactly that sum). And the current prober will clear it even if the cap
is still in force, because `pingClaude`'s switch sends a 400 to
`default: return true, ""` — only 401/402 return false. Added to the RUNBOOK.

⇒ **The 12 pages' PLAN is correct and proven; the pages themselves have not been
rebuilt.** Nothing about the placement result is contingent on the wave: the
measurement is at `site_plan_sections`, which is already written. The wave only
re-stamps the 14 guides and rebuilds index.

### 9. CORRECTION, and the outcome: the calculators are RIGHT, and I shipped 14 duplicate pages doing it

> **CORRECTED 2026-08-17 ~16:30Z. §6 and §7 above called the 15 `needs_page` items
> "the 14 guides — the same stamp-convergence churn as 08-15". THAT WAS WRONG, and
> it was wrong in the one direction that mattered.** They were builds for **14 NEW
> pages**. What caught it: running the handoff's own page-identity md5 — it had moved
> from the script's pre-fire `fd2c09c2…` to `da6908df…`, which is exactly what that
> check exists to detect. **I ran it ~35 minutes after the fire instead of in the
> minutes after the plan landed, and by then the fleet claim gate had re-opened.**
> The cheap check I skipped: the item keys themselves. `needs_page:can-i-overpay`
> is not `needs_page:guide-can-i-overpay`, and I read the list twice without
> noticing the missing prefix, because the 08-15 framing told me what to expect.
> **An inherited framing is an [ASSUMED] claim wearing a measurement's clothes.**

**What the re-fire's plan actually did.** `9463e31d` **retyped the whole guides
section**: 14 pages of role `blog-post` at `/blog/<slug>.html`, and **zero** pages
of role `guide`. Both earlier plans had 14 `guide` at `/guides/<slug>.html`
(measured: `34b1b056` guide=14, `dcbae4df` guide=14, `9463e31d` blog-post=14,
guide=0). So this was introduced by MY re-fire, not inherited — same script, same
12-page recompose scope, one image later.

**Damage, and it completed.** The 14 rows were created 11:46:26 with zero
components; their 14 `needs_page` items dispatched as soon as the `bugs_open/243`
gate re-opened at 12:10:17, and **all 14 duplicates are now `build_status=deployed`
and serving 200** alongside the 14 real guides, which are untouched and also serving.
Same content, two public URLs, 14 times. I attempted the guarded containment (cancel
the items, archive the zero-component rows — the owner's 08-15 answer #2) at 12:12
and the permission classifier refused the write; escalated to the owner and the
window closed while it sat.

**The platform caught it itself, which is worth recording as a credit.**
`orphan_blog_posts` fired — *"14 blog posts deployed but not linked from blog
listing page"* — and its remedy minted a **43-page rerender batch**
(`items_created: 43`), which is the origin of the 40 `page_rerender` rows still
`triaged`. So one bad plan cost a whole-site rerender wave on top of 14 builds.

**AND THE THING THE LANE EXISTED FOR CAME OUT RIGHT.** [MEASURED at the artefact,
16:17Z] `needs_page:index` rebuilt the homepage at 13:44:08 and the locked
calculator is now at **position 2** — properly composed as the second section, where
the plan put it — instead of appended at position 6 as the pre-fire row order had it.
Locks **12/12** held throughout. And `toolgolden.py --compare` against
`GOLDEN_2026-08-17_post_rebuild` returns **exit 0, "all 11 tools reproduce their
golden values exactly"**, index included. So the 282 fix, LOCK-008 and the re-fire
did together exactly what they were meant to: the calculator moved into its planned
slot and still computes identically.

**Serving now: 42 of 43 active pages 200.** `guides-index` is still the only
persistent 404 — its rerender came back `needs_human_review` with *"no sections
ready to build (empty spec sections)"*, so it remains uncomposed, unchanged from
this morning. ⚠ **`tool-damage-checker` read 404 once at 16:25 and 200 on three
immediate retries** — a mid-republish window during the rerender wave. A single 404
sampled during a wave is not evidence of a broken page; re-sample before believing it.

**Failure pattern to hand on: 8 of 13 failures are `failed to get latest commit/base
tree`** — the git deploy path, not the build. Includes `page_rerender:index`, which
failed to publish while the page itself deployed fine at 13:44 and passes toolgolden.
So these are late rerenders failing to republish, not lost work. With 40 rerenders
still queued against the same path, expect more.

### 10. The "fresh chassis build" is NOT what the cluster is running — the same-tag trap, live

The owner deployed a fresh chassis build. [MEASURED 16:15Z] It is not running:

```
cluster pod digest   sha256:f90a7e88…   from commit 6a782274b   (pods restarted 14:43Z)
local image v1.0.1305 digest sha256:6039e19c…  from commit 89a0cbeb7  built 15:30:44
```

`IMAGE_TAG` is still `v1.0.1305` in the makefile, so the rebuild pushed a **new image
under the same tag** and the node kept its cached binary — the exact failure CLAUDE.md
warns about ("a same-tag rebuild ships the node's stale cached binary"). The pods also
restarted at 14:43, *before* the 15:30 build, so they never had a chance to carry it.
**The running binary is 202 commits behind the image that was built** (`89a0cbeb7` is
an ancestor of HEAD). Needs an `IMAGE_TAG` bump + `make release` (owner runs releases;
they are whole-fleet). Nothing in this lane is blocked by it — 282 is live in
`6a782274b` — but no fix committed today is live.

**The tag is not the artefact and neither is a pod restart.** Two readings agreed on
"v1.0.1305" and disagreed on the binary; the digest is what settled it.

### 11. CORRECTION to §4 — the calculators ARE in the site navigation, in the footer, and the way I missed it generalises

> **CORRECTED 2026-08-17 ~16:40Z. §4 above, and what I told the owner at midday, said
> the rebuild had dropped the calculators from the navigation and that the framework's
> parent-listing page was missing. The second half stands; the first half is FALSE.**

[MEASURED 16:40Z] The framework placed all 11 calculators in the `utility` nav group —
the footer — which is precisely what `classifyPagesForNav` does with a never-primary
page that carries a nav flag ("Barred from the main menu, kept in the footer"). Three
readings, each at a different layer:

```
site_nav_items          11 rows, group_key='utility', one per calculator
footer site_component   all 11 /tools/ hrefs present, updated_at 13:47:45Z
served page             /guides/secured-vs-unsecured.html (deployed 16:21:45Z):
                        footer has 21 links, 16 to /tools/, all 11 calculators
```

**What fooled me, and it is a reusable trap.** I sampled the served footer on `index`
and `tool-settlement-calculator`, both deployed **13:44:08Z**. The chrome regenerated
at **13:47:45Z** — three minutes later. **A page carries the chrome that existed when
it was last rendered, so a chrome change is invisible on every page that has not
re-rendered since it.** I read "no /tools/ links in the footer" as a statement about
the site and it was a statement about one page's render age. Worse, the rerenders that
would have fixed it are the ones failing on the git deploy path, so the stale state was
stable enough to look permanent.

**The check, before concluding anything about served chrome:** compare
`site_components.updated_at` for the slot against the page's `deployed_at`, and sample
a page whose `deployed_at` is LATER. One query:
`SELECT name, deployed_at FROM pages WHERE site_id=… AND deployed_at > '<chrome updated_at>' ORDER BY deployed_at DESC LIMIT 3;`
If none exists, the served chrome is old everywhere and you cannot yet judge it.

**What genuinely remains** is narrower than §4 claimed: the **header** carries only
`Home`/`About` plus a CTA, because `page_type='tool'` and `/tools/` URLs are barred
from PRIMARY nav by design and this site has no `tools-index` parent. So the real
change is header-dropdown → footer-list. That is a UX judgement for the owner, not a
loss of navigation.

### 12. The fresh chassis IS live this time — and the lane is now publish-blocked by a fleet-wide deploy outage

**The build landed properly.** [MEASURED 18:14Z] Tag bumped to **v1.0.1307**, pods up
17:05:24/17:05:46Z, pod digest `sha256:8339bdbd…` matching the local image for that tag
exactly, whose label gives the stamp **`a6d1c53c068a5df421479cc9e8801f251f80d539`**.
Positive control PRESENT in `/proc/1/exe`, negative control (the old `6a782274b`) ABSENT
— so the probe discriminated. Ancestor of HEAD; **296 commits** gained over the binary
that ran this morning. Contrast §10: the earlier "fresh build" was a same-tag rebuild
that never reached the nodes. Another lane hit the same trap and fixed it at source
(`aa9c7b74f`, "bump IMAGE_TAG to v1.0.1306 — v1.0.1305 was reused … 24 code commits
across ~10 lanes are inert").

**But nothing can publish.** [MEASURED] **0 pages have `deployed_at` after the 17:05
roll.** The 8 `failed to get latest commit/base tree` errors recorded in §9 are the
local edge of a **fleet-wide deploy outage running since 13:31Z**, found and filed by
the portfolio_positioning lane (`fdd8ca54f`): ~832 base-tree 404s, every affected site
having `github_repo` EMPTY. **loancalculator.co.uk is named in their table (4 in a 4h
window), and I confirmed it first-hand: `github_repo = (EMPTY)`, `deploy_config = {}`.**

Two things they found that make it hard to see, worth carrying: it is logged under
`error_code = 'LLM_API_ERROR'` (a deploy fault under an LLM code, so grepping for a
deploy problem misses it), and 808 of the rows carry a NULL `site_id`, so per-site
triage under-reports it ~70×.

**Their `090` came back `UNVERIFIABLE` (stopped: scope-not-narrowing)** — corr
`75220928-935a-4e5d-8982-802992b0af34`, completed 16:41Z — and **this site is the
evidence that stopped it**: loancalculator has identical row state throughout the
window, yet some requests fail at branch/base-tree with 404 while a later one gets PAST
that stage and fails at ref-update with 503, "a stage a genuinely unbuildable repo name
could never reach". So "empty `github_repo` ⇒ 404" is a correlation, not the mechanism.
Named next step, not yet done by anyone: read the component that routes git vs bucket,
and what `resolveGitRepoNameDB` actually returns for an empty `github_repo` (the
`sendGitCommitRequest` comment mentions a sites-table FALLBACK repo name, not an empty
string). **Shared deploy infrastructure — not this lane's to fix.**

**Consequences for this lane, all downstream of that one outage:**
- The chrome fix (§11) is stuck at **10 of 43** pages carrying the new footer. The rest
  keep the pre-13:47:45 chrome until they can republish.
- 29 `page_rerender` still `triaged`, 12 failed, 1 re-claimed 18:01. Failure mode has
  shifted from the base-tree 404 to request timeouts since the roll.
- **Retracting the 14 duplicates would itself need a working deploy**, so even with the
  owner's answer in hand, the cleanup cannot execute until this clears.

⇒ **The lane is blocked on infrastructure, not on a decision, and the two blocks are
independent.** Nothing here changes the §9 result: the calculators are correct in the
plan and correct on the pages that did publish (toolgolden exit 0, locks 12/12).

## 2026-08-18 — the owner chose /guides/, and the framework already has the control for it

**Owner decision:** *"I would prefer /guides/ but I am happy to accept the most natural
fix for the code."* So: the guides keep `/guides/<slug>.html`, and the fix should be the
framework's own mechanism rather than anything bespoke.

### 1. The natural fix exists, was written FOR THIS SITE, and has never been on here

`normaliseRealisedToPlanPage` (`v3_site_actions.go:5615-5672`) stamps a realised-derived
plan page `identity_authority: "realised"` and carries `parent_section`, so
`CanonicalisePage` keeps the page where it is SERVING instead of re-deriving the role's
default hub. Its own comment names our incident before it happened: *"CanonicalisePage
re-derives a blog-post's URL under /blog/, which MOVES a live page that is serving from
/guides/ (the bugs_open/241 URL-move hazard)"*. `bugs_open/241` was filed **while planning
this site's rebuild**, 2026-08-10.

It is opt-in per site via the structure spec (`site_identity_policy.go:75-110`), unsafe
default off per the owner ruling of 2026-08-02. [MEASURED 2026-08-18] on this site:
`honour_realised_identity`, `twin_identity_snap`, `stem_twin_snap` **all NULL**;
`url_shape` correctly `flat` from the 08-11 seed. So the control for the exact hazard we
hit was present, built, documented — and off.

**Why the planner reached for `blog-post` at all** is the same 241 mechanism from the other
end: `CanonicalisePage` maps `role=guide` to `/guides/<slug>/index.html` and **no input
produces the flat `/guides/<slug>.html` this site actually serves**. The only roles that
emit a flat `/<dir>/<slug>.html` are `blog-post` and `entity-page`. So a flat guide is
unrepresentable to the planner, and the nearest expressible shape puts the dir at `blog`.

**Seed written:** `SEED_2026-08-18_identity_flags.sql` (all three flags, supersede-then-
insert, `DO`/`RAISE` verify that aborts if `url_shape` or the 27-entry pages list is lost).
All three because of a precondition established **2026-08-17 by the
loanandmortgagecalculator lane** (`96c83ebff`): `honour_realised_identity` is **inert
unless a snap or union re-stamped the page** — the marker is stripped from every LLM page
(`:6476`) — and that precondition is not stated where the flag is documented. They enabled
the flag alone, having measured the population first, and got the twins anyway.
`stem_twin_snap` is the layer for our shape (bare plan page vs prefixed realised, either
direction).

### 2. Why Pass C2 — which is written for our exact case — could not fire

`Pass C2` drops a plan entry that re-proposes an adopted item under a different
prefix/role/URL. Its comment's example is literally ours: *"'economy-basics' beside the
adopted 'guide-economy-basics'"*. It could not fire, and the reason is stated in the code:
`itemStemSets` is built from `noCurrentPlanPages`, and *"in practice that makes it
first-plan-only: noCurrentPlanPages is empty whenever the site has a current plan
(bugs_open/051)"*. This site had a current plan, so C2's index was empty. **A guard that
exists, matches your case by name, and is structurally unreachable on a re-plan.**

### 3. ⚠ CORRECTION to §6 — my "0 RECOMPOSE tells" was not evidence

> §6 above reads: "**0** `RECOMPOSE_INTENT_NOT_REALISED` rows (every page genuinely
> recomposed, no no-ops)". **Withdraw the parenthesis.** [MEASURED 2026-08-18]
> `agent_error_log` has, fleet-wide and all-history, **zero rows** for
> `RECOMPOSE_INTENT_NOT_REALISED`, `FACT_ASSIGNMENT_ABSENT`, `FACT_CARRY_MISS` and
> `PLAN_PAGE_IDENTITY_SNAPPED`. The only `PLAN_PAGE_*` code ever written is
> `PLAN_PAGE_MERGE_LOSSY` (2 rows, 2026-08-11) — and it is emitted from a **different
> file**. So the channel my check read has never carried a message, and my zero could not
> have come out any other way.
> **The control that makes this legible:** the table itself is written heavily (1,856
> `RESOLVER_CONFLICTING_CANDIDATES` in 12h), so the sink works — the silence is specific
> to these codes, not to logging. Part of it is by design: a `url_exact` snap
> deliberately does NOT append to `IdentitySnaps` (`:5885-5905`), so a run whose only
> snaps are Pass B renames records nothing. Our run's `pages_restamped: 2` is consistent
> with that, so I cannot say whether the twin layers have ever had demand.
> **I applied exactly this discipline to the endpoint-health row this morning and not to
> my own recompose check hours later.** A zero needs its demand control every time, not
> when it occurs to you.

### 4. And the evidence window closed while I was in it

The planner run's `collected_data` (orchestration `627ff71a`) held the identity-snap
detail. **It had purged before I could read it — gone within ~2 hours**, though this lane's
own cautions say "planner rows purge in ~2 days" and the code comment at `:3623` says the
orchestration row "expires in ~24h". I read `reconcile_result` from it at ~13:50 and the
whole row was absent by ~18:30. **Read a run's payload the moment you have a question about
it; the summary you already copied is all you will keep.** That is precisely why
`recordIdentitySnaps` writes durable rows — and why those rows being empty matters.

## 2026-08-20 — the duplicates are ARCHIVED; the file deletion is dispatched-but-blocked

Re-checked everything first, because the thread was two days old. **The lane was
untouched by other sessions** (zero commits to this dir since 07a44b2eb, though 686
commits landed fleet-wide), and — the check that mattered — **the identity flags I seeded
on 08-18 SURVIVED**: `honour_realised_identity` / `twin_identity_snap` / `stem_twin_snap`
all still true, `url_shape` still flat, 27-entry pages list intact. That was not a
formality: LANDMINES records that a re-adoption silently drops those opt-in keys.

**Chassis v1.0.1317 verified live** — pods 2026-08-19 22:26Z, pod digest
`sha256:64783665…` matching the local image for that tag, stamp
**`2d13d530d`** present in `/proc/1/exe` with the previous stamp `a6d1c53c0` absent
(so the probe discriminated). Ancestor of HEAD; **1091 commits** gained since v1.0.1307.
Nothing else about the site had moved: plan still `9463e31d`, locks 12/12, 43 active.

### The dry run this agent cannot express, done read-only instead

`page-retraction`'s step config passes only `site_id` and `page_ids`; `dry_run` is read
from **step config** (`:281`), so dry-running through the shared agent would mean editing
a definition other lanes use. Computed the same answers directly:

```
inbound  nav rows -> /blog/    0
inbound  chrome references     0     => no refusal expected
inbound  other page bodies     0
outbound newly stranded        1  /tools/interest-rate-stress-test.html
```

The one outbound hit is **reported, not orphaned**: it loses its only in-body inbound
link but stays in the footer nav (utility group, all 11 tools) and in the sitemap.

### Two guards found by reading, and one trap avoided

- `retract_page_deployment` **refuses an active page** — *"retracting a live page is not
  what archiving means"* (`:169`). So the archive must come FIRST; the order is not a
  preference.
- ⚠ **Its default selection is "every non-active page with a `deployed_at` stamp", and
  this site has one OTHER such page**: `tool-standard-calc` (archived 08-03, and it holds
  one of the 12 locked calculator rows). Its file already 404s so the delete would be a
  no-op, but it would have ridden along as an undecided change. **Explicit `page_ids` is
  mandatory here, not tidiness.**

### Done, and what is left

**Archived 14 rows** (guarded on `page_type='blog-post'` AND `status='active'` AND
`url LIKE '/blog/%'` AND created after the fire; `RETURNING` printed exactly the 14).
Site is back to **29 active pages**, 15 archived. Guides and tools untouched, locks 12/12.

**The file deletion did NOT run.** `retract_blog_duplicates.sh` (this dir, explicit ids)
was refused by the session's permission classifier — it creates a pod in the `kafka`
namespace to publish. So the 14 `/blog/` files still serve 200. That half-state is stable
and no worse than before: archived pages are refused by the deploy path
(`ARCHIVED_PAGE_DEPLOY_REFUSED`), so nothing will republish them; they simply persist
until the script is run by someone who can dispatch.

## 2026-08-22 — 227 moved to bugs_closed (contribution from an independent session)

Re-verified live before moving, all first-hand: active row `e0194bee` census
`~* 'provocation|gauntlet|arena|vonc|spark'` → FALSE (and this held AFTER a 198-row
fleet-wide agent sweep at 08:36Z the same morning — no `agent_snapshots` entry, bulk
touch, fix survived); `persist_plan` reachable only via `check_approved.then_step`
(target-field scan, not `::text`, per the 370 warning); `max_rounds` = 5;
`debt-difficulty-help` plan of record clean, both vonc-shaped rows still demoted;
`site-chat-intake` (08-15) is a second clean non-vonc plan. The owner-raised
fundamentallyai `needs_experience_plan` row closed satisfied 08-17 (other lanes'
rebuilds; see its `result.closed_2026_08_17`), so nothing queues behind 227. Move made
under the owner's 08-12 restored fixed-AND-live bar; commit `baa8102e0` names both
paths. Full closure evidence is in the bug file's CLOSED section.

## 2026-08-23 — state check (3-day gap): everything I did held; the retraction still has not run

Re-checked rather than assumed. **1,471 fleet commits** since 08-20; one touched this lane
(`b91dcaae3`, another session recording bug 227's closure in the lane docs — unrelated to
the /blog/ work).

**Held, all measured 2026-08-23:**

```
pages          29 active / 15 archived · 0 active blog-post · 14 active guide
identity flags honour_realised_identity, twin_identity_snap, stem_twin_snap all TRUE
               url_shape flat, 27-entry pages list intact  (survived a 5-day gap)
locks          12/12
plan           9463e31d still is_current — no replan has run
chassis        v1.0.1328, stamp 2dbe12f1d, pods 11:51Z — VERIFIED (positive control
               present, previous stamp 2d13d530d absent, ancestor of HEAD)
```

**NOT done: the file deletion.** `/blog/can-i-overpay.html`, `/blog/loan-faqs.html`,
`/blog/jargon-buster.html` all still serve **200**. `retract_blog_duplicates.sh` has not
been run — it needs a dispatch this session's permission classifier refuses.

**One page moved: `tool-credit-roadmap` re-deployed 2026-08-22 10:34Z** and now matches the
current plan exactly — `hero, tool-credit-health-check, ported-prose, faq, tool-cta`, i.e.
the calculator at **position 2**. Its `owned_page_review` ticket is still open, so something
else drove it.

> **CORRECTION to my 08-17 flag on this page.** I recorded that credit-roadmap "was given
> another page's calculator" and left the impression of damage. Checked at the served
> artefact today: it carries **its own instance** — zero element ids shared with
> `/tools/credit-health-check.html`, and identical interactive structure to it (13 buttons,
> 4 scripts, 3 onclick, 22 ids on both). So it is a *working* questionnaire of the same
> KIND, not a broken page and not a stolen instance. Two pages offering the same kind of
> tool is a content decision for the owner, which is where I should have left it.

**And credit-roadmap is the worked demonstration of what the 11 held tickets do.** Measured
on `tool-overpayment-calculator`, which is typical of the other ten:

```
live now :  hero, ported-prose, faq, tool-cta, [calculator LAST at position 5]
plan says:  hero, [calculator at 1], ported-prose, faq, tool-cta
```

**On all ten un-rebuilt tool pages the calculator sits at the BOTTOM**, below the prose, the
FAQ and a call-to-action. The plan puts it directly under the hero. That is the same move
the homepage rebuild already made (6 → 2) and credit-roadmap has now made. It is the
substantive remaining work on this lane, and it is gated only by the owner releasing the
11 `owned_page_review` tickets (TP-004's human gate).

### 2026-08-23 (later) — RETRACTION DONE: the 14 duplicates are gone, verified at the artefact

Owner authorised; the dispatch went through on this attempt (the earlier refusals were this
session's permission classifier, not the cluster). Corr
`d7f7f5b3-501b-4f14-b7a2-db6576acdf27`, orchestration `8045c4a9-2f3c-400b-94ff-0fb82d394277`,
**COMPLETED in 9 seconds**.

**The audit, which is the part worth keeping** — a `complete` status is not evidence, so
these are the action's own recorded numbers plus the served check:

```
considered        14
retracted         14      (zero refusals)
nav_retired        0      (nothing pointed at them — matches the pre-flight)
editorial_inbound  null   (no body/chrome links — matches the pre-flight)
stranded_targets   null
git                repo gqls/sites, commit a1508b92…, success:true, 14 paths deleted
durable row        RETRACTION_AUDIT in agent_error_log, 12:42:11Z
```

**At the artefact (the only thing that counts):** all **14 of 14** `/blog/` URLs now return
**404**; **28 of 29** active pages return 200, `guides-index` still the only 404. Controls
held — `/guides/can-i-overpay.html`, `/guides/loan-faqs.html`, `/index.html` and
`/tools/overpayment-calculator.html` all still 200.

> **One over-cautious call of mine, corrected by the run.** My read-only pre-flight predicted
> the action would REPORT one stranded target (`/tools/interest-rate-stress-test.html`, which
> loses its only in-body inbound link). It reported **none** — correctly, because the tool is
> in the site-wide footer, and the action asks the same reachability question the orphan check
> does, which includes chrome. The pre-flight's three inbound counts were exact; only this
> outbound prediction was wrong, and it was wrong in the harmless direction.

**⇒ The 2026-08-17 incident is closed at the artefact.** Site: 29 active pages, 14 guides at
`/guides/`, 11 tool pages, 12 locks held, identity flags on so a replan will not re-mint the
twins. What remains on this lane is the 11 held `owned_page_review` tickets (the calculators
sit LAST on ten tool pages) and `guides-index`.

## 2026-08-23 (owner's four instructions) — released, composed, and the duplicate-tool decision

Owner: *"1. deleted, 2. release the rebuilds. 3. build and restore the Guides link 4. we
only need one of them."*

### (2) "release the rebuilds" is NOT a status flip — the tickets have no handler

[MEASURED] All 11 `owned_page_review` items carry **`handler_agent = ''`**. They are review
MARKERS, not build jobs — TP-004's gate is "no handler by design". Setting them `triaged`
would leave them unroutable, not rebuild anything. **The rebuild is a separate item**, and
this site already has a proven one: `needs_page` / `page_rerender:<page>` /
`page-build-handler`, which is what rebuilt `tool-credit-roadmap` on 08-22 into plan shape
(created by the `bugfix_337_redrive` lane, not this one). Copied that row's exact shape
rather than inventing one — `pipeline=build, approval_mode=auto, priority=99,
spec={reason,page_name}`, and `created_by` is NOT NULL.

**CANARY FIRST, one page, deliberately.** Ten live money pages is too many to move on an
arm nobody has exercised. Two halves have precedent and their COMBINATION does not:
index rebuilt with a LOCKED calculator (08-17, moved 6 → 2), and credit-roadmap rebuilt as a
TOOL-ROLE page (08-22) — but credit-roadmap's calculator is **not locked**. The ten remaining
are tool-role AND locked. Canary = `tool-overpayment-calculator`
(`page_rerender:tool-overpayment-calculator`, born `triaged`). Baseline recorded:
`hero, ported-prose, faq, tool-cta, tool-overpayment-impact(locked, position 5)`.

### (3) The Guides 404 had a plain cause: the plan composes NOTHING for that page

[MEASURED] In plan `9463e31d`, `about` and `legal` each carry 2 sections and **`guides-index`
carries ZERO**. That is precisely what its build kept reporting — *"no sections ready to
build (empty spec sections)"*. The page is in the plan, its nav row exists, and nothing had
ever composed it.

Took the fleet convention rather than inventing: a `section-index` composes
`hero, <thing>-list` — `gamesdesign.co.uk/guides-index` is `hero,guide-list`, dartsonline and
idea.uk use `hero,content-listing`. **Composed `guides-index` as `hero, guide-list`** (0-based
ordering, matching this plan's own rows), and `guide-list` is already live on this site's
homepage, so it is a component this site demonstrably renders.

Then **re-drove the EXISTING item rather than creating one** — `idx_swi_dedup` refused a
duplicate `page_rerender:guides-index`, correctly: the item was sitting at
`needs_human_review`, attempt 1/3, with the error that the composition has now fixed.
⚠ **Side effect worth knowing: re-triaging preserved its `created_at` of 2026-08-17, which
made it the OLDEST triaged build item FLEET-WIDE** — so `find_dispatchable_site`
(`ORDER BY created_at ASC LIMIT 1`) now selects this site ahead of everyone. That is how a
re-drive jumps the queue, and it is worth knowing before doing it on a busy fleet.

### (4) "we only need one of them" — keep credit-health-check, and retiring the other is NOT clean

[MEASURED] The two pages carry the SAME component function, `tool-credit-health-check`:

```
tool-credit-health-check   deployed 08-19   tool-credit-health-check   LOCKED
tool-credit-roadmap        deployed 08-22   tool-credit-health-check   not locked
```

**So keep `tool-credit-health-check`** — its instance is the protected, golden-verified one;
`tool-credit-roadmap` holds an unlocked second instance minted in the rebuild era.

⚠ **But its retraction WILL BE REFUSED as things stand: 15 pages carry 16 links to
`/tools/credit-roadmap.html`, plus 1 active nav row.** That is the graph guard working —
inbound editorial links refuse the retraction and name the referrers, because repairing prose
is a content decision. Unlike the `/blog/` set, which had zero of all three.

**Sequence that makes it clean, and it composes with (2):** archive `tool-credit-roadmap`
FIRST, then run the remaining rebuilds — their cross-link sections regenerate from the ACTIVE
page set, so the links should clear themselves — then retract the file, then the nav row goes
with the retraction. **[INFERRED] that the rebuilds drop the links; the canary and the first
rebuilds are the test, and if the links persist the retraction stays refused and the owner
has a prose decision.** Not doing (4) until (2) has proven itself.

### (3) DONE — `/guides/index.html` is BUILT and SERVING; the LINK needs one more step

[MEASURED 2026-08-23] `page_rerender:guides-index` completed 13:14→13:16:47.
**`/guides/index.html` now returns 200**, 2 components, and the page links to **all 14
guides**. The site's last 404 is gone: 30 active pages, 30 serving.

**The fix was the composition, and the diagnosis is worth keeping**: the plan listed the page
and composed NOTHING for it, so every build honestly reported "no sections ready to build".
Two rows into `site_plan_sections` (`hero`, `guide-list`) and the existing item — re-driven,
not duplicated — built it first time on its second attempt.

⚠ **But the LINK does not ship yet, and the reason is a second mechanism.** The served
header still shows only Home/About + a CTA. [MEASURED] the chrome components (`header`,
`footer`) were last rendered **2026-08-20 17:42** and contain **no** `/guides/index.html`.
The nav ROW has existed since 08-15, so the chrome renderer excluded it deliberately —
**it was not linking a page that 404'd**, which is correct behaviour. Now that the page is
deployed, the chrome has to be rebuilt.

Raised the framework's own item for it rather than hand-editing chrome: **`nav_drift` →
`nav-updater`** (48 completed fleet-wide, so a well-exercised path), copying a live row's
shape — *"rebuild nav tables and re-render chrome so the link ships"*.
`nav_rebuild:e31c71a8…`, priority 30.

**So "restore the Guides link" is two mechanisms, not one**: build the page (done), then
regenerate the chrome (queued). A session that only did the first would see a working page
and a menu that still does not mention it.

### (2) canary in flight — and it REGENERATES PROSE, which is the thing to watch

The canary's handler is real work, not a stuck claim: orchestration `0648ce0f` sat at
`process_sections_loop_iter_4_generate_content` for minutes. **That means a tool-page rebuild
re-generates section CONTENT, not just the section ORDER** — so the page's tuned voice-H
prose is rewritten by the builder. Nobody has checked that on this lane: the 08-17 index
rebuild was verified with toolgolden, which records TOOL VALUES and would not notice changed
copy at all.
**Check after the canary lands** — `page_component_history` retains the previous rows, so
the before/after is recoverable: compare the `ported-prose` and `faq` content for this page
across the rebuild before releasing the other nine.

### (2) CANARY PASSED — and the evidence that matters is not the one I planned to use

`page_rerender:tool-overpayment-calculator` completed 13:24:20. [MEASURED]

```
served section order: hero · tool-overpayment-impact-section · ported-prose · faq · tool-cta
component rows      : 1 hero · 2 tool-overpayment-impact (LOCKED) · 3 ported-prose · 4 faq · 5 tool-cta
```

**The calculator moved from position 5 to position 2, on the page and in the rows.** Locks
12/12 sitewide.

**Two checks that make this more than "it completed":**

1. **The locked row was never written.** Its `updated_at` is still **2026-08-09** while every
   other component on the page shows 13:24:19. The rebuild REPOSITIONED it without touching
   its bytes — LOCK-008 doing exactly its job, on the arm nobody had exercised (a tool-role
   page whose calculator is locked; index proved locked-but-landing, credit-roadmap proved
   tool-role-but-unlocked).
2. **No copy was rewritten.** I feared the opposite: the handler sat in
   `process_sections_loop_iter_4_generate_content`, which reads like the builder rewriting
   prose. It did not. Every archived/saved pair at 13:24:19 is **md5-identical**:
   `674 acde538f` (tool-cta), `3264 5817941c` (ported-prose), `3280 aa8ccab8` (related
   items), `3941 f4006ce4` (faq). **A step named "generate_content" that produces identical
   bytes is not evidence of a no-op until you compare the bytes** — I nearly recorded the
   fear as a finding.

⚠ **THE ACCEPTANCE HARNESS IS DOWN, and I only know that because I ran a control.**
`toolgolden.py --compare` fails on the rebuilt page with `timeout waiting for
Runtime.evaluate`. That reads exactly like "the rebuild broke the calculator". **It is not:
the same harness times out identically on `settlement-calculator`, which has NOT been
rebuilt.** So it is the environment/chromium, not the change. Two consequences:
- **The lane's tool-verification instrument is unavailable right now.** Anyone quoting a
  toolgolden pass after 2026-08-23 should check it actually captured.
- For THIS question the lock evidence is stronger anyway: toolgolden proves the values
  still compute; `locked_at`/`updated_at` proves the bytes were never touched.

⚠ **Also noticed, and NOT caused by this rebuild:** the FAQ heading on the served page is now
*"Overpayment calculator: common questions"*, where the 08-17 golden recorded *"Questions
people ask about overpaying a loan"*. The faq content_data is md5-identical across today's
rebuild, so this changed EARLIER — the page re-deployed 08-17 19:00, after the golden was
captured at ~12:40. **The golden is stale relative to the pages**, which is a second reason a
compare would have looked alarming. Re-baseline once the harness works again.

**Released the remaining NINE** at priority 15 (`source='loancalc_owner_release_20260823'`),
excluding the canary and excluding `tool-credit-roadmap`, which the owner has chosen to
retire.

### (2)+(4) converging, and the link-clearing hypothesis is now MEASURED, not inferred

[MEASURED 2026-08-23 ~14:00] Three of the ten tool pages now carry their locked calculator at
**position 2** (`tool-overpayment-calculator`, `tool-compare-loans`,
`tool-credit-health-check`); seven remain at position 5 and are queued. **Every one of the
ten is still `locked=true`** — the rebuilds have not broken a single lock.

Verified the served artefact for a NON-canary page, `tool-compare-loans`:
`hero · tool-compare-loan-offers-section · ported-prose · faq · tool-cta`. So the canary
generalises.

**The (4) hypothesis is confirmed.** I recorded "[INFERRED] that the rebuilds drop the links
to credit-roadmap; the rebuilds are its test". They do: `tool-compare-loans` no longer
contains `/tools/credit-roadmap.html`, and the site-wide inbound count has fallen
**16 instances / 15 pages → 8 / 8** as pages rebuild.

**The 8 that remain are all `guide` pages**, which the tool-page rebuilds do not touch. They
clear via the OTHER queue: the 31 remaining `page_rerender` items ("Rerender page after
template fix", another lane's wave) target the guides too.
⚠ **A trap I nearly fell into here:** `guide-debt-consolidation-explained` shows a
**`complete`** rerender AND still links to credit-roadmap, which reads like "a rerender does
not clear the link". It does not read that way once dated: **that completion is 2026-08-17**,
and the archive happened **2026-08-23 13:42:43**. Its pending rerender from today's wave has
not run yet. **A completed job only tells you about the world at the moment it ran** — the
same staleness lesson as the golden, one table over.

⇒ So the retraction for (4) needs no prose decision after all, provided the guide rerenders
drain. Retract only once the inbound count reaches **0**; the action refuses otherwise, which
is the guard doing its job rather than an obstacle.

**Still queued:** 7 rebuilds (priority 15) and the `nav_drift` (priority 30, so it runs after
them) that ships the Guides link. Chrome still reads 2026-08-20 with no `/guides/` link.

### 2026-08-23 (evening) — ALL FOUR INSTRUCTIONS COMPLETE, verified at the artefact

```
(1) duplicates deleted    14/14 /blog/ URLs 404
(2) rebuilds released     10/10 tool pages: locked calculator at POSITION 2,
                          verified on the SERVED page for all ten, locks 12/12
(3) Guides                /guides/index.html 200 with all 14 guides;
                          chrome regenerated 14:50:24 WITH the /guides/ link
(4) duplicate tool        tool-credit-roadmap retracted -> 404;
                          tool-credit-health-check (the LOCKED instance) 200
site                      28 active pages, 28 serving 200 — ZERO 404s
```

**(2) generalised cleanly from the canary.** All ten served pages now render the calculator
as section 2 (`hero · <tool>-section · ported-prose · faq · tool-cta`), measured by parsing
each served page, not by trusting the rows. **Locks 12/12 throughout** — ten rebuilds, not
one lock lost. The 10 `owned_page_review` tickets are now closed as satisfied.

**(4) completed without the prose decision I had warned might be needed.** The sequence held:
archive first → the tool rebuilds and the template-fix rerender wave regenerated their
cross-link sections from the ACTIVE page set → inbound links fell **16 → 8 → 0** → then the
retraction ran clean (`considered=1, retracted=1`). `nav_retired=0` is CORRECT here and not a
miss: the 14:50 nav rebuild had already dropped the row, because by then the page was
archived.

**(3) has a tail worth stating precisely.** The chrome is right — regenerated 14:50:24 with
the Guides link and without credit-roadmap — but **a page only gains it when that page next
re-renders**. At the time of writing 15 of 28 pages carry the new chrome and 13 do not
(including the homepage). All 13 already have a queued rerender, so it ships without further
action. **Until it drains, the Guides link is live on some pages and absent on others**, and
sampling one page would misreport it either way — the same lesson as 08-17 §11, which I had
already been caught by once.

**Bookkeeping done:** 10 review tickets closed; credit-roadmap's 3 tickets cancelled with the
page.

## 2026-08-24 — the harness was never down, and the first thing it said when it worked was that a live calculator is broken

Sent here for two things: continue `HANDOFF_2026-08-23_continue_here.md`, and *"fix the
calculation verification harness that you said is broken"*.

### The harness: an environment fault wearing a broken-calculator costume

`toolgolden.py` failed today with **`RuntimeError: chromium did not start`** — not the
`timeout waiting for Runtime.evaluate` the 08-23 notes recorded. Both are start/attach-layer
faults; only the first is reproducible now, so only the first is measured below.

[MEASURED 2026-08-24] The cause is one line in `toolprobe.start_chrome`, which launches
chromium with `--user-data-dir=tempfile.mkdtemp(...)`. **`mkdtemp` honours `$TMPDIR`**, and
this session's `TMPDIR` is `/home/ant/.claude-scratch/gotmp`. The snap-confined chromium
cannot write under a **hidden** top-level directory of `$HOME` (snapd's `home` interface
grants `owner @{HOME}/[^.]*/**` only), so it aborts at ProcessSingleton before opening the
DevTools port. Discriminating launch, all three arms in one run (scratchpad `t1.py`):

```
HIDDEN-HOME  /home/ant/.claude-scratch/gotmp/tg-hidden-test  -> NO START in 30.1s, rc=21
SLASH-TMP    /tmp/tg-tmp-test                                -> UP in 3.0s
VISIBLE-HOME /home/ant/tg-visible-test                       -> UP in 0.5s
```

rc=21's own last line names it exactly — *"Failed to create a ProcessSingleton for your
profile directory"* — and **we were sending it to `DEVNULL`**. The poll also could not tell
"still starting" from "exited 200ms ago", so it burned the full 30 s and then said nothing
useful. Two blindfolds in six lines.

**This is a shared seam, which is why it matters beyond this lane.** `start_chrome` is
imported by `toolgolden.py`, `toolaudit.py`, `evalpage.py`, `defect_vectors.py`,
`investor_golden.py` and `oracle_driver.py` — **6** scripts across **4** lanes as of
2026-08-24 — so one `$TMPDIR` takes all of them down together, and each reports it in the
vocabulary of the *page it was pointed at*.

Fixed in `start_chrome` (commit `0aafce405`): fall back to a profile dir the confinement
permits (`/tmp`, then a non-hidden `$HOME`) saying which and why; fail fast on `poll()`
carrying chromium's own stderr; and **refuse a port that already serves DevTools**.
That last one is not tidiness — attaching to a browser this run did not start gives you a
browser whose targets are not yours, and the first `Runtime.evaluate` then hangs. `[INFERRED]`
that this is how the 08-23 wording arose; that environment is gone and I cannot re-measure it,
so it stays inferred.

### `--selftest`, because the 08-23 session was saved by a control it had to think of

That session only knew the failure was environmental because it re-ran against a page it had
**not** rebuilt and saw the identical failure. That is exactly right, and it should not depend
on someone thinking of it. `toolgolden.py --selftest` drives a fixture whose answer is known
in advance through the **same** `Runner.capture()` — same navigate, settle, `DRIVE_JS` scaling,
`SNAP_JS` read, press.

The expected values are computed **by hand from the driver's own rules** (each numeric field's
`value` attribute × the vector's factor; `ASYM` = 1.7 then 0.6), not read back from a recorded
run — a fixture whose expectations came from the harness could not fail if the harness were
wrong about everything. All four vectors land exactly: `1050.00 · 2200.00 · 512.50 · 1751.00`.

**Proven disconfirmable by mutation**, per [[mutate-the-code-to-prove-the-guard]]: perturbing
the fixture's arithmetic (`r/100` → `r/200`) turns all 8 value checks red and exits 1 — while
**gates A and B stay green**. That is the whole argument for asserting values: a wrong-but-
responsive calculator satisfies both gates, and the gates are what the platform's other
instruments already check.

### Then it worked, and immediately convicted a live page

Full 11-URL `--compare` against `GOLDEN_2026-08-17_post_rebuild`. Ten captured normally.
`/tools/loan-vs-savings.html` came back **`react=0` — INERT**, and the inert gate refused the
run before any diff printed (correct: it will not certify from a broken tool).

It is not a harness artefact. Three independent measurements, [MEASURED 2026-08-24]:

1. **The 08-17 golden for this page records it WORKING** — 4 controls, `react=5`
   (`results`, `loan-panel`, `save-panel`, `loan-benefit`, `save-benefit`), `vary=5`.
   Today: **8** controls, four ids each appearing twice, `react=0`.
2. **The served page** has every one of that tool's ids **twice** — `loan-rate`, `save-rate`,
   `spare-cash`, `tax-bracket`, `results`, `loan-panel`, `save-panel`, `loan-benefit`,
   `save-benefit` — plus `function compare` and `function copy` twice. A site-wide census of
   all 28 served pages found duplicate ids on **exactly this one**.
3. **The rows say why.** `page_components` for the page:

```
pos slot_name     component_id  locked  html_len  created
 2  tool-2        448422ce…     t       11845     08-02   <- the locked calculator
 6  tool-2        NULL          f       11845     08-23 14:14  <- an unlinked COPY
```

Both blobs md5 `be85284e7f61e452ea19178f4502713f`; both `content_data` md5
`f65a0b6e82cd5b1e43e44563d400f35e`. **Byte-identical, same slot name, one locked, one
orphaned.** It was written by the owner-released rebuild that finished 14:15:19 — the same
wave that got the other nine right.

**Why the served page is dead rather than merely doubled:** `SNAP_JS` keys the fingerprint by
`e.id`, so duplicate ids collapse and the **last** copy wins; the script's
`getElementById` writes to the **first**. The harness reads copy #2, which never changes. A
visitor sees the same thing — the lower calculator does nothing visible.

### A plausible cause, refuted in one query — which is why this went to `090` and not into a bug file

The obvious story was `bugs_closed/189` (*"resolving a locked positional slot duplicates it"*),
which was filed **on this very page** and closed 08-21 as fixed and live. Its population is
locked sections whose `slot_name` is positional rather than the component's function. So:
did loan-vs-savings duplicate because its locked slot is called `tool-2`?

**No.** [MEASURED] **all 11** locked tool sections on this site are positionally named and
none matches its function (`tool-1`…`tool-4`), ten of them went through the same wave, and
only this one duplicated. Positional naming is not the discriminator. 189's shape also
differs: its duplicate carried the **same** `component_id` as the locked row; today's carries
**NULL**.

I nearly wrote the 189 story down. It is exactly the shape CLAUDE.md sends to the loop —
cross-cutting, cause plausibly not where the symptom is, and a durable assertion about a
shared write path. Filed: intake `0a53b04e-e06e-48c8-ad11-4845d8ee96d5`, run correlation
**`b53c355b-7bfc-4202-b61d-89f16decffe2`**.

Fleet context, dated because a census goes stale by addition: `component_id IS NULL` rows on
**active** pages number **11 across 6 domains as of 2026-08-24**, and **two were created on
08-23** (`gamesdesign.co.uk/games/jelly-invaders`, slot `section`; this one). Not a
loancalculator-only shape, and not historic.

### The 08-23 handoff's open tail is CLOSED

[MEASURED 2026-08-24, curling all 28 active URLs] **28/28 serve 200**, **28/28 carry the
`/guides/index.html` link**, **0/28 still reference `credit-roadmap`**. The chrome drain the
08-23 session left in flight (15 of 28 at hand-off) has finished, so the "sampling one page
misreports it either way" caution no longer applies to this site.

### What the golden is NOT

**No re-baseline was written, deliberately.** The harness refuses to record a golden while any
tool is inert, and it is right to: a golden captured now would pin "loan-vs-savings answers
nothing" into the acceptance record and then defend it. `GOLDEN_2026-08-17` also remains stale
for a second, unrelated reason the 08-23 notes already found (FAQ headings changed in the
08-17 19:00 re-deploy). **Re-baseline after the page is repaired, not before** — the order is
not a preference, it is the gate.

### The repair landed, and the golden is re-baselined — with one wrong call of my own on the way

Owner ruled: repair by hand now, re-baseline after. Both done, both verified at the artefact.

**The delete.** One row (`3fd2639d…`), inside a transaction whose `WHERE` asserted every
distinguishing property (`component_id IS NULL AND locked_at IS NULL AND position=6 AND
slot_name='tool-2' AND md5(rendered_html)='be85284e…'`) so it could not reach the locked row,
followed by a `DO`/`RAISE` block — deliberately not a block of `SELECT`s, which
`ON_ERROR_STOP` ignores and which therefore cannot stop a `COMMIT`. Guards: 5 rows remain, the
locked row still locked, its bytes unchanged, its `updated_at` still `2026-08-02
23:01:02.947526+00`. `DELETE 1`, all guards passed.

**Recoverability was verified, not assumed:** `trg_page_component_artefact_archive_del` wrote
the row to `page_component_history` (`op=delete`, `source=artefact_archive_trigger`, md5
`be85284e…`).

**A check worth copying, made BEFORE the redeploy ran:** `pages.sections` is a materialised
cache that LOCK-008 merges locked rows into, so a stale sixth entry there would have let the
assemble re-materialise the duplicate and made the repair look done. [MEASURED] it held **5**.

**Then the assemble-only redeploy** (`98529d02…`, **no `spec.reason`** — `section_data_resolved`
is the route `bugs_closed/189` warns reproduces this, and it would rewrite 51 prose rows here).
Completed 19:04:33, commit `e1becb2a` to `gqls/sites`.

⚠ **AND HERE I GOT SOMETHING WRONG, recorded in `WRONG_CALLS.md`.** I curled at ~19:06, saw the
old bytes and `last-modified: 16:51:46`, and told the owner *"the bucket wasn't updated"*. It
had been: the `Deploy to B2` workflow logged `upload tools/loan-vs-savings.html` at
**19:04:57**. I sampled inside the propagation window. What made it persuasive is worth
keeping: every OTHER page read `19:04:5x` and this one alone read `16:51:46`, which looks like
*"the sync touched all of them and skipped mine"* — and is actually **position in the sweep**.
A one-shot comparison against peers who share the confound is not a control. The lane's own
standing caution (*"a single 404 sampled during a rerender wave proves nothing; re-sample"*)
covers this exactly; I had read it that morning and applied it only to status codes.

**Verified state, [MEASURED 2026-08-24]:**

```
served bytes  sha256 e3d2da2b… == the committed file, exactly     (was d30d112c…, 57,349 B)
duplicate ids 0        (was 11)
harness       react=5  vary=5  12 fields — identical to the 08-17 golden's own record
divergences   8, ALL the cosmetic c-faq container rename; zero controls, zero numeric
```

**Re-baselined:** `acceptance/GOLDEN_2026-08-24_post_385_repair_tool_values.json`, all 11 URLs,
written only after `--selftest` passed. **Proven to reproduce: a fresh `--compare` against it
returns `all 11 tools reproduce their golden values exactly`, exit 0.** A golden that has not
been compared against once is a file, not a baseline.

⚠ **The 08-17 golden is superseded but NOT wrong** — keep it. It is the only record of the
pre-rebuild values, and it is what proved `385` was a regression rather than a long-standing
fault (`react=5` there against `react=0` live).

## 2026-08-25 — post-roll acceptance: 11/11 exact, and 385 did not recur (but today's wave did not test the arm that broke)

Picked this lane back up after a fresh chassis build. Checked what moved first, because 750
commits landed under this thread overnight.

### What changed beneath us

```
chassis   v1.0.1339, build provenance git_commit a7459a44b — asked the SERVICE, not git
          (2 pods, both on that image, started 2026-08-25 19:07)
ancestry  my last commit (3973260c5) IS an ancestor of a7459a44b, and a7459a44b is an
          ancestor of HEAD — so the running binary post-dates this lane's work
my files  UNTOUCHED by any of the 750 commits: toolprobe.py, the whole loancalculator_couk
          lane, and bugs_open/385. 385 is still open, still unowned by anyone else
```

⚠ **`save_page_sections_action.go` is UNCHANGED between my session and the running build.**
One commit (`c735bfd9c`, the `bugs_open/375` verifier gate) touched neighbouring files but
`git show c735bfd9c -- save_page_sections_action.go` is **empty**. So **385's writer is still
live, exactly as filed** — the roll did not fix it and nobody claims to have.

### The acceptance answer, which is what this lane exists for

`--selftest` green first, then all 11 URLs against `GOLDEN_2026-08-24_post_385_repair`:
**`all 11 tools reproduce their golden values exactly`, exit 0, ZERO divergences** — not even
the cosmetic FAQ rename this time, because the 08-24 golden already records it.

That is the first clean post-roll acceptance run this lane has ever had with a working
instrument, and it covers a chassis roll AND a 24-page rerender wave on the same day.

### 385 has NOT recurred — and the ten tool pages rebuilt today

[MEASURED 2026-08-25] `byte_twins_on_page` (385's own discriminator, not a bare
`component_id IS NULL` count) is **0 for all 9 remaining orphan rows fleet-wide**. On this
site: 28 active pages, **11 locks held, 0 orphan rows**.

**And this is the part worth keeping:** all ten tool pages rewrote 4 of their 5 rows today
between 13:04 and 13:34 — **including `/tools/loan-vs-savings.html` at 13:11, the page that
duplicated** — locked row untouched on every one, and **none duplicated**.

⚠ **BUT THAT DOES NOT EXONERATE THE ARM THAT BROKE, and reading it as if it did is the trap
here.** [MEASURED] today's wave was **24 `page_rerender` items, `source='side_effect'`**, from
a `tool-cta` template change — the **rerender** arm (`rerender_sections`). The 08-23
duplication came from a **`needs_page`** item on the **build** arm (`page-build-handler` →
compile → `save_page_sections`). **Two different upstreams into the same INSERT.** So what
today proves is: *the rerender arm does not duplicate on a locked positional tool page*. The
build arm remains untested since the failure, and it is the one that failed.

### A lead I chased and CLOSED, negatively — do not re-chase it

`page_component_history` retains what the 08-23 rebuild saw. Its five
`source='save_page_sections_overwrite'` markers are a **pre-overwrite snapshot of the rows
that already existed** (`save_page_sections_action.go:830-843`, `SELECT pc.id …`), so they
describe the page BEFORE the rebuild, not the composition it received. The page had 5 rows:
faq, tool-cta, the locked calculator, hero, ported-prose.

Four of those five markers carry `component_id` NULL and one — the calculator's — carries
`10be4f71`, which [MEASURED] is a `page_components` **row id** (1 hit) and **not** a
`content_components` id (0 hits; the real component is `448422ce`). That looked like a
type-confusion lead pointing straight at an unresolvable component.

**It is not the discriminator.** Fleet-wide, **153 of 23,627** overwrite markers carry a row
id (0.6%) — and **101 of those are on THIS SITE, across 12 pages, from 08-03 to 08-25**,
against exactly **one** duplication in that whole window. The population it tracks is almost
certainly the locked/retained rows (this site has 11 locks; idea.uk, the next biggest, has
11 pages). **A 101-to-1 ratio is not a cause.**

I also checked the obvious "the code changed underneath the data" explanation before writing
any of this down: `git show ec653247f:…save_page_sections_action.go` (the commit HEAD was at
when the duplication happened) carries the **identical** snapshot SQL. So the anomaly is real
and it is simply not this bug's.

### Where 385 stands

Unchanged in substance from 08-24: damage repaired and verified, **cause still not
established**, writer still live. What today adds is three constraints — it has not recurred
anywhere; the rerender arm is clean on all ten locked tool pages including the victim; and
the overwrite-marker lead is closed. The next move is still §5b's, and it is now sharper:
**test the BUILD arm**, because that is the one that has not run on a locked positional tool
page since it failed.

## 2026-08-25 (second session) — 385's CAUSE ESTABLISHED: the Layer 2 re-append, matched by slot-name string alone

Full chain with inline evidence: `bugs_open/385` **§5c** (single source — this entry
records the investigation's shape and its wrong turn, not a second copy of the finding).

**The finding in one paragraph.** The writer is `save_page_sections_action.go`'s Layer 2
interactive-tool preservation block (`:484-608`): its preload (`build_status='deployed'
AND <interactive>`) selected the locked calculator — `[MEASURED 2026-08-25]` **the only
locked row in the fleet satisfying that predicate**; the other 11 locked rows on this site
are `'approved'`, which is the entire reason nine sibling pages in the same wave were
untouched — and its matcher (`:551-558`) pairs stored rows to incoming sections by EXACT
SLOT-NAME STRING only. The build arm names sections from the plan (function names), so
`tool-2` ≠ `tool-loan-vs-savings` → "slot dropped entirely" → append a verbatim copy of
the stored row with `ComponentID:''` (RFC_046 opt-in OFF). In the insert loop the plan's
tool entry consumed the lock via the identity arm (058/182/204 all working correctly), and
the appended entry — lock consumed, no id — became the byte-identical NULL-component
orphan at position 6. Third matcher of "is this section already in the set"; the other two
(`matchLockedRow`, `MergeLockedPageSlots`) were given identity arms, this one never was.

**The handoff's named candidate was WRONG, and refuting it first is what found the truth.**
HANDOFF_2026-08-25 pointed at LOCK-008's `MergeLockedPageSlots` vs `matchLockedRow`.
Read both, arm for arm: they DO mirror (merge arm 3 pairs the plan's function name to the
locked row's `cc.function`, which `[MEASURED]` is exactly `tool-loan-vs-savings` — and the
08-17 census in this file proved the match held before the failure). The merge inserted
nothing. What kept the investigation honest: the `llm_call_log` iteration census (iters
{0,2,3,4}, no iter 5 → five-entry composition), which killed every "the list had six
entries" story and forced the search downstream of the writer loop — where the only
appender between compile and INSERT is Layer 2.

**Dead ends, so nobody re-walks them:** orchestration_states for 08-23 are purged (expected,
~2d); generate_content prompts do not carry the section list (0 mentions of either name);
the 08-19 21:18–21:22 batch on 9 tool components that made `content_components.updated_at`
look suspicious was `component-template-fixer` (template repairs, one compose_note per fix
in llm_call_log — not function renames).

**Discriminating censuses run today** (queries in bug §5c): 12 locked rows on site — one
`'deployed'` (the victim's), 11 `'approved'`; armed set fleet-wide (locked + deployed +
interactive) = **1 row, the same one — STILL ARMED**: the next build-arm rebuild of
`tool-loan-vs-savings` duplicates it again. Pre-build state census via the delete-trigger
archive: the three single-pass pages (application-tracker, consolidation, loan-vs-savings)
all had writable rows at 1–4 with the lock at the tail; only the `'deployed'` one broke —
position/tail was a red herring, `build_status` is the whole discriminator.

**Fix path:** bug §8 candidate 0 — give the Layer 2 matcher `matchLockedRow`'s arms (all
the data is already in the preload's `preservedSection`). Data-side disarm available
without a roll: flip the one row to `'approved'` (owner call — it writes a locked row).
Harness: `--selftest` green before all of this, per the standing order.

### Same session, later — the fix is BUILT and COMMITTED; verdict owed

Commit `a799579fd`: `matchPreservedSectionIdx` with the siblings' arms + consumption;
six unit tests, a wiring scan, and a whole-action pin of the 08-23 shape. Mutation run
(slot-only revert): all six unit tests RED — **but the whole-action test stayed GREEN
under the mutation**, and the reason earns its line: the insert loop tolerates a failed
INSERT (Warn + continue), so sqlmock's unexpected-call error for the orphan INSERT is
swallowed and `sections_saved` still reads 1. A guard in series hiding the mutation —
the `a-mutation-that-passes-may-have-hit-a-guard-in-series` pattern, caught because the
memory rule says name WHICH guard; the test's own comment now records it, and the
mutation load sits on the unit tests + wiring scan. `verify-head-builds.sh` OK at
`a799579fd`. Council corr `ece638fb` (`Council-Submitted:` trailer) — verdict owed.
Pre-existing at HEAD and NOT mine: `findingcodes_scan_test` fails on
`WORK_ITEM_STATUS_OVERRIDE_REFUSED` (the 396 lane's `2b46afbe6`, today 20:48).

### Round 2 — the council's REVISE was RIGHT, and the unification is in

Round 1 drew a REVISE, gating objection from `reuse_agent`: my fix added a THIRD
hand-mirrored copy of the pairing arms, and three copies drifting apart is precisely how
385 was minted. Revised rather than defended (the estate's own rule, and the seat was
correct): the relation now lives ONCE in `datahelpers/slot_pairing.go` and all three
askers — `matchLockedRow`, `MergeLockedPageSlots`, `matchPreservedSectionIdx` — are thin
adapters over it. Equivalence for the two LIVE matchers is their existing suites passing
unchanged (6 matchLockedRow tests, the merge table); matchLockedRow gains NO arm (its
SlotIdentity views carry empty function/name — the loader doesn't join cc, and widening
would be its own change). New pins only the core can carry: arm PRIORITY across
candidates (mutation-verified — loop-swap went RED), plus wiring scans in BOTH packages
that go red if anyone re-inlines a private copy (the behaviour suites are equivalence
proofs and would stay green — that's why the scans exist).

Also answered from the verdict: guardian's consumer census (8 real dispatchers of
`save_page_sections`; council-gate/fix-proposer are prompt-text false positives — in bug
§9 now, dated) and debug_historian's liveness probe (stamp merge-base + symbol probe with
BOTH controls — in bug §9). Resubmitted on the SAME correlation
(`RESUBMIT_CORR=ece638fb`), so the trail accumulates.

⚠ One mechanical trap paid for here: `git checkout --` cannot restore an UNTRACKED
file after a mutation run — the mutated `slot_pairing.go` stayed on disk reading as
restored. Caught by the post-restore build+test pass; reverted by hand. Mutate tracked
files, or diff after "restoring".

### Round 2 verdict: APPROVED (2 advisories, none high) — and the residual dispositions

`ece638fb` round 2: **approved**, reuse gate lifted. Advisories dispositioned rather than
shelved, recorded in register **LOCK-009** (the new entry for `slot_pairing.go` — the
mechanism clears the "another workstream could call this and would not know it exists"
bar, so it is registered in the same arc that shipped it):

- **editquality's ordering concern is ANSWERED, not accepted**: both old closures were
  already arms-outer (matchLockedRow looped each arm across all rows before the next;
  the merge's pair() likewise), so the shared core PRESERVES the ordering — the seat
  could not see that from the sketches, and now the register says it where the next
  reviewer will look.
- **bug_historian [medium]**: matchLockedRow deliberately stays function-arm-less (its
  loader doesn't join cc); the widening is a candidate follow-up on its own evidence.
- **guardian [low]**: the merge passes componentID '' by convention — structural today
  (its input is `[]string`), named in LOCK-009 so a signature change trips over it.

Lane state at close: 385's cause established and fixed at the class level, council
approved, register + landmine + 016b §9 all carrying it. **Residue: (1) the roll** (the
fix is inert until an image ships — merge-base `b9d0f02be` vs the pod stamp, symbol
probe in bug §9), **(2) the owner's disarm decision** (flip the one armed row to
'approved'), **(3) build-arm verification post-roll** (bug §9 — do NOT fire a
needs_page at loan-vs-savings before (1) or (2)).

## 2026-08-26 — the fix is LIVE; the stale detection fired harmlessly; two waves inbound

**Fix LIVE `[MEASURED 2026-08-26]`.** Overnight roll: pods `6dd68888dc-*` started
2026-08-25 23:11Z (after `b9d0f02be` at 21:03Z). Symbol probe on BOTH replicas:
`matchPreservedSectionIdx` present, `PairStoredToIncoming` present, present-control
`matchLockedRow` found, absent-control `matchPreservedSectionIdxZZZ` absent. Triggered by
a peer heads-up (webdesign-tool-rebuilds: design rotation re-enabled 09:20Z) that made
"is the fix live yet?" urgent — the answer was already yes.

**The 08-24 stale detection resolved itself, harmlessly, and taught something.**
`content_duplication:tool-loan-vs-savings` — filed 08-24 14:49 against the REAL duplicate,
hours before our hand remediation; never retracted by it — dispatched 08-26 08:30 when
promotion resumed. The dedupe agent deleted nothing (its `remove_component_ids` named the
row already gone; `page_component_history` shows zero writes on the page today), completed
the parent, filed a benign assemble-only `page_rerender` (same op as §7 step 4).
**Lesson, now in bug §7b: a hand repair must grep the queue for detections of the damage
filed BEFORE the repair — they outlive it and fire when promotion resumes.** (Family:
"a closed blocker keeps being obeyed" / "checking the pod does not check the queue".)

**Health, all measured today:** page 5 rows, locked calc pos 2 with 08-02 bytes; 0 orphans
and 0 byte-twins fleet-wide; served page 0 duplicate ids; harness selftest green then
golden `--compare` MATCHES on the victim URL (one page deliberately — mid-morning sweeps
active; the full-11 run stays the post-wave check).

**Inbound (peer heads-ups, both recorded in the 08-26 handoff):** design rotation ramp
(~2-3 days to this site; findings born `detected`, promoter auto-dispatches) and the
`bugs_open/397` GTM repair (chrome + 44/47 pages re-render+redeploy — the rerender arm).
Neither can exercise 385's build arm; churn in `created_at`s and chrome is expected.

**385 stays OPEN on one criterion:** a clean build-arm rebuild (bug §7b). Everything else
about the bug is done: cause (§5c), class fix (LOCK-009), council APPROVED, fix live.

## 2026-09-02 — the evidence register is BUILT and LIVE (owner-directed, RFC_060 §4; migration 699)

**The instruction** arrived by peer relay (lendzy lane): the five register-less finance
sites build their registers; this site is one (RFC_060 §1b — zero `evidence_base` rows,
so the numeric scan never armed and the daily citation check was vacuously green).
Method = lendzy's RUNBOOK §8 / migration 695, followed step for step.

**What was done, all `[MEASURED 2026-09-02]`:**
- All 28 served pages crawled; 232 regulatory-shaped sentences extracted FROM THE
  ARTEFACT (scratch script; inventory in the session record). Unlike lendzy (CONC caps),
  this site's external assertions are dominated by **Consumer Credit Act 1974 rights** —
  so the register cites legislation.gov.uk (7 CCA sections + 2 SIs) alongside
  handbook.fca.org.uk (CONC 5.2A.5, CONC 7.3.4, DISP 1.3.1). The refresher re-fetches
  any citation URL, so both hosts get the daily quote re-check.
- Every provision FETCHED AND READ; every URL confirmed by `<title>` (both hosts 200 on
  wrong paths). **All 12 quotes verified through the production matcher**
  (`cmd/fcaquotecheck`), 12/12 true, absent control false in the same runs.
- **Migration 699 applied**: `699 OK ... COMMIT`; read-back shows 12 facts, pinned,
  current, `created_by='loancalculator_couk lane (migration 699)'`, 09-02 15:30.
  Rollback sidecar written. Council corr `1f259a95` (`Council-Submitted:`).

**The exercise caught the site TWICE (both `corrects_site_citation`, copy NOT touched —
bugs_open/320 §15, owner's call):**
1. `tools/settlement-calculator.html`: "usually within ten working days" — the
   prescribed period is **12 working days** (SI 1983/1564 reg 4, quoted verbatim).
2. `tools/overpayment-calculator.html`: "overpay up to 10% of your outstanding balance
   in any 12-month period without a charge" attributed to the CCA — **no 10% rule
   exists**; the statutory threshold is **£8,000 per 12 months** (s.95A(2)(a)), cap
   1%/0.5% (s.95A(4)). The 10% allowance is a mortgage-product convention.

**And it caught ME once — the near-miss is in the migration header:** the 12-day period
traced first to SI 1983/**1569**, whose schedule covers ss.77-79/103/107-110 and NOT
s.97; the right instrument is SI 1983/**1564** (reg 4 names s.97(1) explicitly).
Reading the schedule, not the title, is what caught it — the method working on its
own operator.

**Consequences now live:** the daily refresher re-checks 12 citations from its next
run; `HasScannableRegister` arms the numeric scan on future saves — measured-safe for
this site (RFC_060 §1c: zero findings on our 474-component export; fad209b92 live), and
the pending 397 GTM rerender wave passes the armed scan with the same zero expectation.
`rule` fields are human-verified (dated) until Q6's rule-span checker ships.

**For the owner (in README):** the two copy corrections are his prose decisions.

### 2026-09-02, later — where the traps landed, and the fleet picture

Lendzy relay: this lane's three findings (legislation.gov.uk 200s-on-wrong-paths; the
SI 1983/1569-vs-1564 near-miss, pointing at 699's header; run-the-method-expecting-errors)
are now **RUNBOOK §8c in `docs024_key_docs_latest/lendzy_co_uk/`** — the durable home;
future register lanes get them from there, not from this file. Fleet state: **4 of
RFC_060 §1b's 5 register-less sites populated the same day** (lendzy 695, loanzy 697,
farmerinsurance 698, this site 699) — the §1b census is stale by success. The fifth,
loancash, has NO seat (owner knows); it inherits §8/8b/8c, and the s.97 settlement-
deadline trap is live for it (payday proposition). Measured across the day, fleet-wide:
**five wrong claims found across three sites** — the register work is finding errors at
a steady rate, which is the strongest argument it should exist. Claims-verification
lane (register mechanism + Q6 rule-span checker) has our results and the 699 header.

### 699 verdict: APPROVED (1 advisory, medium) — and the advisory was verified, not filed

Council `1f259a95`: **approved**, one medium advisory from bug_historian — a raw-SQL
register INSERT risks a later TYPED writer (WriteSiteSpecAction / an EvidenceFact
struct) silently dropping fields the struct does not model (`corrects_site_citation`,
`rule`, `writer_line`) on re-serialisation.

**Disposition, verified in the code the same hour `[MEASURED 2026-09-02]`:** the daily
writer is SAFE — `refresh_evidence_base_action.go` unmarshals `data` into
`map[string]interface{}` (`:338`), mutates only the keys it manages, and
`writeRefreshedEvidenceBase` (`:1440-1500`) marshals THE SAME MAP back
(supersede-and-insert, CAS-guarded on the row id, `pinned` carried). Unknown keys
round-trip losslessly. The HAZARD STANDS for any future typed writer to this aspect —
that is the claims-verification lane's seam (they own the mechanism), and they have
been told with the verification attached. Note as designed: after the first refresher
pass the row's `created_by` becomes `evidence-refresher`, so 699's ROLLBACK guard will
refuse — correct, per its own header.

### 2026-09-02, later still — banned_claims armed (migration 707): the archetype constraints are enforced prose now

The bugfix_414 lane's relay named the residue precisely (their register census was
stale — 699 was live by the time it arrived — but the second half was real): 699
shipped `banned_claims: []`, and the site's `site_archetype.constraints` ("never appear
to give regulated advice" / "never recommend lenders" / "never reposition as a
lender/broker") was agent-written prose no gate reads. **707 translates it into the
enforced channel**: 8 regex patterns — 5 adapted from adversecreditmortgage's audited
set, 3 from our own constraints — live at the build gate, the persistence floor and the
post-deploy sweep. Council `99bd846e` submitted; applied (8 armed, 12 facts + both
`corrects_site_citation` fields carried through the supersede — asserted in the verify).

**Every admission decided by measurement `[MEASURED 2026-09-02]`:**
- All 8 compiled AND probe-fired through Go's own engine (8/8) FIRST — `claims.go`
  QuoteMeta-degrades a non-compiling pattern into a silently inert literal.
- Full-text census over all 28 served pages: **0 matches for all 8** — arming cannot
  refuse any current page, including the pending 397 GTM rerender wave.
- The sibling set's literal-%APR pattern **EXCLUDED on a 2-match census** —
  compare-loans' illustrative "7.9% APR loan and an 8.4% APR loan" is the pedagogy;
  banning it is §1c's convict-the-site-for-doing-its-job class, measured not argued.

**Two lessons inherited from loanzy's 702 (same ask, same day) instead of re-paid:**
- The no-credit-check pattern is their NARROWED form (bans the lending PROMISE, not the
  phrase — a calculator site truthfully says its tools involve no credit check).
  Re-censused separately here (the hyphenated alternative is NOT a subset of the broad
  form I first measured): 0 matches, planted control fires.
- **Supersede-and-merge, never an in-place register edit**: the refresher's write-back
  CAS is keyed on the row id it read, so an in-place jsonb_set between its read and
  write is a silent lost update (my first draft had exactly that shape). The supersede
  changes the id → the refresher skips → nothing lost.

Also: another session took migration number 700 while I authored (the 70x space filled
700-706 in ~an hour — loanzy's 702 header records the mirror-image collision on 699);
renamed to 707 before anything referenced it. `created_by` hands to 707, so 699's
rollback guard now refuses — correct per both headers. First-findings review after
arming is the 414 relay's standing instruction: nothing to review today (0 matches),
the next save through the claims floor is the first live exercise.

### 2026-09-02, close — the QuoteMeta parenthetical became a landmine correction

The 414 lane tested 707's header claim and found it CONTRADICTED the live banned-claims
escaping landmine ("a pattern that fails to compile is caught" — true only for the
Go-authored fleet set, pinned by a CI test that structurally cannot see DATA-authored
patterns; the per-site door at claims.go:348 is silent on both halves). Corrected as
`4f1ca1384`, with a fleet census: 239 live per-site patterns across 19 sites, 0 broken,
2 planted controls firing — clean baseline, stale by ADDITION on the next seed. **707's
compile+probe+census method is now the RECORDED remedy** for register authors. The
literal-%APR exclusion is a 3-of-3 measured convergence (us, loanzy 702,
adversecreditmortgage) going into RFC_060. My position on the unbuilt mechanical check —
fold a compile-and-report into the daily refresher's existing loop rather than leave the
remedy as prose or build new surface — sent to 414 to carry to claims-verification, who
own the seam. farmerinsurance's empty banned_claims flagged to the lendzy relay, not us.

### 707 verdict: APPROVED (2 medium advisories) — both answered at the ARTEFACT, same hour

Council `99bd846e`: **approved**. editquality + debug_historian [medium] made the same
correct catch: my 8/8 compile check ran on the SOURCE strings, not the bytes that landed
in jsonb after SQL-literal escaping — and the live landmine for this exact mechanism
("seeded through dollar-quoted JSON is DOUBLE-ESCAPED, compiles cleanly, matches
nothing") is precisely what a source-side check cannot see. The 414 lane's 239-pattern
fleet census predates 707's apply, so our 8 were uncovered.

**Closed by measurement `[MEASURED 2026-09-02]`:** extracted the 8 STORED pattern values
from the live row — `cat -A` confirms single backslashes, zero doubled — and compiled
each stored byte-string exactly as claims.go does (`regexp.Compile("(?i)"+p)`) with
probe + absent-control per pattern: **8/8 compile, 8/8 fire, 0/8 false-fire**. The
editquality [low] (all 12 facts intact, not just the 2 asserted corrections) also
verified: id list byte-exact vs 699's authored set, 12 substantive quotes, 12
verified_at stamps.

**Method upgrade for the RUNBOOK-§8c family, learned here: the compile+probe check must
run on the STORED bytes post-apply, not (only) on the source strings pre-apply.** The
pre-apply check catches authoring typos; only the post-apply check catches the escaping
layer. Both are one command each.

reuse_agent [low]: 699/702/707 are three same-day hand-copies of the identical
supersede-and-merge CTE — a fair drift observation; a stored function is the
claims-verification lane's seam (they own the register mechanism) and the council
report carries it to them. guardian lows: no in-flight writer (1 current row verified);
single-site data change, no architecture signal.

### Addendum, for the record's precision (lendzy relay)

The verification ladder has THREE distinct rungs, and different lanes were each missing
a different one: lendzy's original probe was already post-apply against stored values —
their gap was CONSUMER-FORM fidelity (bare pattern vs the `"(?i)"+pattern` the consumer
actually compiles, claims.go:346); ours was source-vs-stored (the escaping layer). The
full ladder is now lendzy RUNBOOK **§8f**, credited to 707's round: stored form at apply
→ stored bytes + consumer-form probe-fire post-apply → read the first real findings.
Lendzy re-verified 5/5 under it; claims.go:348 confirmed by them as the silent QuoteMeta
fallback. Fleet: farmerinsurance's empty banned_claims routed to the loanzy session
(their lane holds it) with 707 cited; **loancash remains the only true register
absence** (unowned, owner aware).

### Addendum 2: the refresher-loop argument CONCEDED and recorded (414 lane, f786ee271 + d38f07a7c)

The 414 lane conceded in writing: the landmine now carries their superseded sentence
struck through with the three counts under it, and RFC_060 gains **§3e** (deliberately a
note, not a new owner question — it is Q7's shape one field along: a guarantee enforced
on one write path only; if Q7 ships as a sweep, this rides the same loop). They VERIFIED
the load-bearing claim rather than take it: `resolveEvidenceSites`
(refresh_evidence_base_action.go:278) selects EVERY site with a current register
(input_data {}, pre_query NULL), :325 loads the whole document — "a few lines in a loop
that already runs over the right population" is exact. Two paid-for traps added to the
costing for whoever builds it: **ON CONFLICT DO NOTHING, never DO UPDATE** (a daily
re-write bumps `updated_at` and makes the row unreapable for ever — bugs_closed/213),
and **key the finding, not the site**. Routed to claims-verification as this lane's
argument; nobody is building it on spec.

**Operational fact for the NEXT session, from their verification:** the refresher's
`last_completed` was 09:08:58 TODAY — before 699 applied (15:30). **This register's
first live pass is tomorrow ~09:09.** The 414 relay's standing instruction applies then:
READ the first findings rather than treating silence as clean — 12 citations re-fetched
(two hosts), quotes re-matched; a `citation_lost` on a legislation.gov.uk URL may mean
revised statutory text, which is the design working.

## 2026-09-03 — the register's FIRST LIVE PASS, read as instructed: 12/12 green with demand

`[MEASURED 2026-09-03]` The evidence refresher ran 09:10:22 and wrote a new current
register row 09:10:31. Its own words: **"12 live-verifiable fact(s) checked, 12
updated, 0 drifted"** — every citation re-fetched from its live source (both hosts) and
its verbatim quote re-matched by the production pipeline, all 12 `verified_at` stamps
advanced to 2026-09-03. This is the demand-controlled green the whole exercise exists
for: the counter proves 12 rows did work, unlike the vacuous clean run over an empty
set that RFC_060 §1 called "the worst kind of green".

**And the 699-advisory verification is now proven ON DATA, not only by code-read:** the
refresher's supersede carried all 12 facts, both `corrects_site_citation` fields, and
all 8 banned_claims through its round-trip untouched. `created_by` is now
`evidence-refresher` — 699/707's rollback guards refuse from here, correctly.

**385:** no `needs_page`/`page_rerender` filed on this site since 09-02 noon, zero
orphan rows — the build-arm close criterion has not been exercised; the bug stays open
on it.
