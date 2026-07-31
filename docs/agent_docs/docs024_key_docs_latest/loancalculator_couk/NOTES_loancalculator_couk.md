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
