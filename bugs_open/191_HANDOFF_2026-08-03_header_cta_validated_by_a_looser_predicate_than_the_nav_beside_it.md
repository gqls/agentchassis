# 191 — the header CTA is validated by a LOOSER predicate than the nav rendered beside it, so chrome ships a 404 button the nav itself would have refused

**Filed 2026-08-03 by the `mortgagecalculator_couk_adoption` lane.**
**FIXED IN CODE 2026-08-04 (`d32692882`) — STILL OPEN, because it is not live.**
**See "Fix record" at the foot before reading the candidates: candidate 2 shipped,
and candidate 1's first-build recommendation was deliberately REFUSED.**
**LIVE on one site, confirmed at the wire.**
**Severity:** medium. Low blast radius *today* (1 site), but it fires on exactly the
condition every adoption passes through — a site with pages planned and not yet built —
and chrome ships on **every page**, so one bad button scales with the next batch.
**Class:** two call sites answer one question differently (same family as `118`, which
fixed the *component-eligibility* spelling of this and did not touch the *link-target* one).

## The defect

`RenderSiteComponentsAction` builds the header's nav items and the header's CTA button in
the same run, from the same `pages` table, and validates them with **two different
eligibility predicates**:

| element | helper | predicate |
|---|---|---|
| nav items | `loadFetchablePageSet` — `nav_tables.go:258` | `status NOT IN ('deleted','archived')` **AND NOT (`datahelpers.NeverDeployedPagePredicate`)** |
| header CTA | `loadResolverPageSet` — `resolve_internal_links_action.go:486` | `status NOT IN ('deleted','archived')` — **no deployment check** |

The CTA fallback (`render_site_components_action.go:172-187`) runs when there is no
`contact` nav item — which, note, is itself *more likely* on a young site, because the
contact item was already dropped by the stricter nav filter. It then asks
`chooseCTATargets` for an "interactive page" and validates that pick against the **loose**
set:

```go
headerPages := loadResolverPageSet(ctx, params, siteID, params.Logger)   // :171
if ctaURL == "" || !headerPages.Contains(ctaURL) {                        // :172
    …
    primary, _ := chooseCTATargets("", "", interactive, hubs)             // :181
    if primary.URL != "" && headerPages.Contains(primary.URL) {           // :182  ← loose set again
        ctaURL = primary.URL
```

So a page that has never been deployed passes, and its URL is written into
`site_components.header`. The file's own comment at `:149-150` states the opposite
intent — *"so it points at an existing page instead of the hardcoded phantom
/contact.html"* — and `:166-170` says the gated template "renders no CTA button rather
than a phantom". Both are true of the *contact* path and false of the *fallback* path.

## Evidence — the two elements inside ONE rendered header

`mortgagecalculator.co.uk`, `site_components.slot_name='header'`, rendered 2026-08-03 11:01 UTC:

```
<li><a href="/index.html"                                    ← nav: filtered to the 1 deployed page
href="/tools/stamp-duty/index.html" class="header-cta"       ← CTA: never deployed
```

```
$ curl -s -o /dev/null -w '%{http_code}' https://mortgagecalculator.co.uk/tools/stamp-duty/index.html
404
```

`tool-stamp-duty` is `build_status='planned'`, `deployed_at IS NULL`. The nav in the same
component was reduced to a single link precisely *because* the other targets are unbuilt;
the button beside it was not.

## Blast radius — MEASURED, and the DB over-reports it

Extract the CTA from `rendered_html` (it is **not** in `content_data` — see the trap below)
and join to `pages`:

```sql
SELECT s.domain,
       substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"') AS cta_href,
       p.build_status, p.deployed_at IS NOT NULL AS ever_deployed
  FROM site_components sc
  JOIN sites s ON s.id = sc.site_id
  LEFT JOIN pages p ON p.site_id = sc.site_id
       AND p.url = substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
 WHERE sc.slot_name = 'header' AND p.deployed_at IS NULL;
```

→ 2 rows: `lendzy.co.uk` → `/tools/price-cap-checker/index.html`, and ours.

> **The DB answer is wrong by one, and the check that catches it is one `curl`.**
> `lendzy.co.uk/tools/price-cap-checker/index.html` returns **HTTP 200**. `deployed_at IS
> NULL` means "no recorded deploy", **not** "does not serve" — the same conflation as
> `bugs_open/098`'s durable half, in the opposite direction. **12 of 14 sites point their
> CTA at a deployed page; 1 is a false positive; the confirmed live 404 is 1 — ours.**
> Do not repeat the DB-only measurement and report 2.

## Diagnosis-loop verdict

Run `1cdd9556-637d-4dd8-8bc6-c1db3f4196cd` (seeded, 2026-08-03): **UNVERIFIABLE, stopped by
iteration-cap.** Not a refutation — and its trail is worth reading, because it
**independently reproduced the static mechanism**, citing the same three lines quoted
above without being shown them.

What it could not close was *occurrence*, and it diagnosed its own failure correctly:

> "cta_url is not a stored column anywhere in the schema shown (it lives inside
> `site_components.rendered_html` as literal HTML) — that query could not have found a real
> row even if the mismatch exists, so its 0-row result is not evidence of absence."

Its `DataRequests` queried `sc.content_data->>'cta_url'`. **There is no such key.** The
occurrence evidence above is that query re-aimed at `rendered_html`, plus the wire check.
An earlier unseeded run of the same symptom failed outright at `assemble_bundle`
(`no scope`); pass `SEED_SCOPE` — see the correction on `bugs_closed/174`.

## Fix candidates, ordered by what closes the door

1. **Give the CTA the nav's predicate.** Validate `primary.URL` against
   `loadFetchablePageSet` (the set `applyNavVisibility` already built) instead of
   `loadResolverPageSet`. Smallest diff; makes the two elements in one component agree by
   construction. **Caveat that must be handled:** `applyNavVisibility` deliberately
   *disables* the filter when a site has **zero** deployed pages (`nav_tables.go:194-207`,
   so a first build does not freeze an empty nav into chrome). The CTA needs the same
   escape or a first build renders no button — which is the *correct* outcome here, since
   the gated template already omits it on empty `cta_url`.
2. **One named predicate, one resolver** — the shape `118` applied to component
   eligibility (`b052249d8`), applied to link targets. Closes the door for the next
   caller, not just this one. Costlier; the right structural answer.
3. Detect-only: extend the phantom-link checks to `site_components.rendered_html`.
   **Weakest** — `bugs_open/071` is the standing record that detection here is already
   discarded, so this adds a finding nobody consumes.

**Do NOT "fix" it by building the target page.** That clears our symptom and leaves the
predicate wrong for the next site.

## How to verify a fix

```sql
-- must return 0 rows AFTER a nav-updater run on a site with unbuilt pages
SELECT s.domain, substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
  FROM site_components sc JOIN sites s ON s.id=sc.site_id
  LEFT JOIN pages p ON p.site_id=sc.site_id
       AND p.url = substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
 WHERE sc.slot_name='header' AND p.deployed_at IS NULL;
```

Then **curl each surviving CTA href** — the query alone over-reports (see above).
`mortgagecalculator.co.uk` is the live reproduction: it is locked, has 25 unbuilt pages and
1 deployed one, so re-running `nav-updater` there re-creates the condition on demand.

## Ownership note

`render_site_components_action.go` is actively worked by the chrome lane
(`117`/`118`/`166`/`167`/`170`). This bug is **distinct** from `118` — that one fixed
"which component template serves this slot", this one is "which page may a chrome link
name" — but the fix lands in the same file, so **contribute into that lane rather than
starting a competing change.** Filed here, not routed at them, per `scripts/who-owns.py`.

---

# Fix record — 2026-08-04

**Commit `d32692882`** (code + registration), `dd9c5ab07` (the standing five and
three wrong calls). Lane: `docs024_key_docs_latest/bugfix_191_chrome_link_policy/`.
Council: **APPROVED, round 4** (`78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30`) — 3 advisory
objections, none high. Rounds 1-3 were REVISE and **every one of them changed the code**;
the trail is worth reading before you touch this area (see "What the council caught").

**STATUS: FIXED IN CODE, NOT LIVE. This file stays in `bugs_open/` deliberately.**
A Go change is inert until an image is rebuilt and rolled, and the bar for
`bugs_closed/` is *fixed AND live* — the defect is still reproducible on
`mortgagecalculator.co.uk` as you read this. Move it only when §"What is owed"
below has been run and passed.

## What shipped — candidate 2, not candidate 1

`ChromeLinkPolicy` (`platform/orchestration/actions/chrome_link_policy.go`,
registered as **LNK-030**): one named decision for *"which page may a piece of
CHROME link to?"*, consumed by **both** `applyNavVisibility` and the header-CTA
fallback. It calls the existing `loadFetchablePageSet`, so
`datahelpers.NeverDeployedPagePredicate` keeps exactly one definition — **no new
spelling of the predicate was introduced.**

**The diagnosis in this file was right about the symptom and one layer short on the
cause.** The two predicates were the visible half. The half that explains how it
happened, and why candidate 1 alone would not have held: **the escapes were inline.**
`applyNavVisibility` disables the deployment filter on a lookup error and on a
first build, and both branches lived inside the nav function, so the policy they
encode was unreachable from any other caller. The CTA's author reached for the
nearest other helper because there was nothing else to reach for. So the unit
extracted is not the predicate — it is the **decision**.

Boundary written where the next reader stands: `loadResolverPageSet` keeps its
loose predicate for its two page-CONTENT callers and now carries a doc comment
saying which set is which and why, plus a source-scanning allow-list naming each
entitled caller **with its reason**.

## The one recommendation in this file that was REFUSED, and why

§"Fix candidates" 1 says the CTA should render **no button** when the site has zero
deployed pages, calling that "the correct outcome here". **It does the opposite: the
CTA takes the nav's escape and goes unfiltered.**

The nav's reason for that escape is at `nav_tables.go:194-202` — chrome is
idempotence-gated, so what the first build writes may never be re-rendered, and a
nav emptied then persists indefinitely. That argument applies to the button
identically. A buttonless header on a site about to deploy 25 planned pages is
*permanently* wrong; a planned-target button is wrong only for the window in which
the platform has already ruled the fully unfiltered **nav** acceptable. Answering the
freeze one way for the list and the other way for the button beside it is a
miniature of this very bug, inside its own fix.

The same argument carries the error case, which is a **real behaviour change**: the
CTA previously vanished on a lookup error. That was never a considered policy — it
was a side effect of `loadResolverPageSet` returning an empty set on error. It is
now unfiltered, like the nav. Named explicitly in the council submission's `risks`
block as the thing to disagree with.

## ⚠ CORRECTION to this file's own measurements — the verify SQL over-reports TWICE

§"How to verify a fix" and §"Blast radius" share a query that **cannot** give the
number they want.

1. **The `LEFT JOIN` on a regex-extracted href inflates it.** A header with **no**
   `header-cta` produces a NULL `p.*` row, which satisfies `p.deployed_at IS NULL`.
   Run it today and you get **6 rows, 4 of them with an empty `cta_href`**. Those
   four are sites with no button, not sites with a broken one. Add
   `AND substring(sc.rendered_html from '…') IS NOT NULL`.
2. The file already warns about `lendzy.co.uk` (200, not 404) and it is right —
   `deployed_at IS NULL` means "no recorded deploy", not "does not serve".

**So the confirmed live 404 was 1 — not 6, and not 2.** The file's own instruction
("do not repeat the DB-only measurement and report 2") was correct in spirit and its
query still reports 6. Corrected query in the lane `RUNBOOK` R1.

## Escape population — measured, and the first measurement pointed the other way

The design rests on "the unfiltered escape is a first-build path, not a routine one".
The obvious query says **19 of 38 sites have zero shipped pages** — which reads as
the disconfirming result. It is the wrong question: **18 of those have no pages at
all**, so chrome renders nothing either way. Split properly: 19 strict, 18
never-built, and **1** (`webdesign.uk`) actually takes the escape. Query in `RUNBOOK`
R2; the near-miss is logged in `WRONG_CALLS.md`.

## Tests — mutation-proven, not merely green

`chrome_link_policy_test.go`: four behavioural tests plus two source scans (each
with a comment-skip, a gone-blind `Fatal`, a Fatal when an allow-listed file
produces zero matches, and a synthetic-line test proving the matcher still fires).

Proven by breaking the code in a `git archive HEAD` copy, so the shared tree was
never dirtied: revert the CTA call site → both scans RED; drop the deployment
predicate from the fetchable SQL → 5 RED; disable the first-build escape → 2 RED;
make `Allows` permanently true → 2 RED. **`nav_visibility_test.go` passes UNEDITED**,
which is the behaviour-preservation proof for the `applyNavVisibility` refactor.

## What the council caught, in order — the trail is the useful part

1. **`bug_historian` (high, gating):** the fix was **inert for every header already
   rendered**. Chrome is idempotence-gated, so correcting the predicate reaches only
   sites whose chrome happens to re-render for some other reason. I had written the
   repair as a manual step in my own verify-later list, which is not a mechanism.
   *Answered with* `markStaleChromeLinkSlot`.
2. **`render_guardian` (high, gating):** show that a corrected `site_components` row
   actually reaches an already-deployed page. *Answered with evidence, not code:*
   `assemblePage` (`rerender_single_page_action.go:532-537`) opens by calling
   `getSiteComponents`, which is `SELECT slot_name, rendered_html FROM site_components
   WHERE site_id = $1` — chrome is re-read on **every** assembly, never baked in. The
   residual dependency (a page must still be re-assembled) is **`bugs_open/117`'s**, and
   is named as a dependency rather than absorbed into this fix.
3. **`editquality` (high, gating):** a duplicated edit in the submission. My artefact
   defect, not the code's. Collapsed.
4. **`reuse_agent` + `guardian` (medium each) — the two that mattered most.** Had I
   checked whether `bugs_open/166` left a mechanism to extend? I had not. It did:
   `repointRetiredChromeSlot`, which signals a needed re-render with
   `build_status = 'pending'` **under `pageComponentAgentWritableSQL`**. My design at
   that point computed staleness above the loop and OR'd it into `force` — inventing a
   second force channel **and bypassing the lock guard**, so a human-locked chrome slot
   would have been forced to re-render. **No test of mine would have caught that.**
   Rewritten to extend 166's mechanism instead.
5. **Advisory, at approval:** the `architecture` seat noted this is the **second** fix of
   this defect class by bespoke named-type-plus-scan (CLC-013 was the first, four days
   earlier, in this same action) and asked that the pattern be flagged as an RFC
   candidate rather than treated as solved. Recorded in LNK-030. `prior_art_librarian`
   objected that the blast-radius claims rested on grep — so they were **re-verified by
   the compiler** (rename each helper, read the build errors: a method that cannot miss
   a caller). `debug_historian` objected that a curl proves a runtime effect but not
   **which binary** produced it — hence the pod-grep leading the list below.

## What is owed before this file may move to `bugs_closed/`

1. **Pod-grep both replicas FIRST** after the next chassis roll — before any curl,
   because a wire check proves an effect but not that the intended binary served it
   (a same-tag rebuild ships a stale one, `bugs_open/153`). This change *removes* the
   string `headerPages := loadResolverPageSet`, so a **real** negative control exists —
   use it rather than inventing one:
   `LoadChromeLinkPolicy` > 0, `markStaleChromeLinkSlot` > 0,
   `headerPages := loadResolverPageSet` **== 0**, `loadFetchablePageSet` > 0 (positive).
2. **Re-run `nav-updater` on `mortgagecalculator.co.uk`** — locked, 25 unbuilt pages
   and 1 deployed, so it recreates the condition on demand.
3. **Re-run the CORRECTED query** (`RUNBOOK` R1) and **curl every surviving href.**
   The DB alone cannot close this bug.
4. **Read the council verdict** for `78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30` and act on
   a REVISE/REJECTED — the code is already on the shared branch.

## Ownership note — discharged

This file asked the fixer to contribute into the chrome lane
(`117`/`118`/`166`/`167`/`170`) rather than start a competing change. Checked before
starting: seven live sessions had *read* `render_site_components_action.go`, none had
uncommitted edits in it, and the change there is two swapped lines plus a comment.
The new mechanism went into its own file to keep that footprint minimal.
