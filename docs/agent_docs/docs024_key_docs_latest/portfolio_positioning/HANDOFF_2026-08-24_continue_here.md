# HANDOFF — portfolio_positioning — 2026-08-24. **START HERE. First task is in §1.**

Supersedes `HANDOFF_2026-08-20_continue_here.md`. Owner read-out:
`SUMMARY_2026-08-21_the_machine_that_writes_the_brief.md`.
Design: `PLAN_2026-08-19_one_flow_three_brief_sources.md` (+ its 08-19b addendum).

**Counts in this file carry the date they were counted** (owner ruling 2026-08-22): a census does
not go wrong, it goes stale by ADDITION and reads as current for ever. Re-run anything load-bearing
before quoting it — `git log --since=<date> --diff-filter=A -- <dir>` is what the date buys you.

---

# 1. ⏭ START HERE — wire `render_sitemap` into a workflow

**The action is built, registered and callable. Nothing calls it**, and that is the whole problem.

**The council said so, and it is right** (correlation `8a004aab-be85-4d6d-bdb1-4fb114f1d64b`,
REVISE, high severity):

> *"The whole rationale for this change is that a working generator which nothing invokes is 'not a
> mechanism'… The plan creates render_sitemap as an action but the risks section itself admits 'the
> action is not yet wired into any workflow' — so after this plan lands, exactly zero sites gain an
> automatically-generated sitemap. **A registered-but-uncalled action reproduces the diagnosed
> defect in a new form.**"*

That is exactly the owner's ruling turned back on the fix: *"All future sites should have
sitemaps."* A callable action is the script-nobody-runs wearing different clothes.

## The scouting is done — here is the pattern to copy

`content-feed-orchestrator` already does this exact thing for RSS, in three steps
(verified 2026-08-24):

```
render_rss_xml   → render_rss_feed, output_field: rss_render_result
check_has_rss    → conditional on rss_render_result.rendered
commit_rss       → git_commit { files_field: "rss_render_result.files",
                                domain_field: "rss_render_result.domain",
                                commit_message: "Update RSS feed" }
```

`render_sitemap` returns the **same `{files: {...}, domain, rendered}` shape on purpose**, so the
same trio works unchanged with `sitemap.xml` in place of `feed.xml`.

**`git_commit` reaches B2 sites as well as git-hosted ones** — checked, don't re-derive it:
`resolveGitRepoNameDB` (`helpers.go:236-246`) reads `sites.github_repo` and **defaults to `sites`
when it is empty**, which is what every B2 domain has. `loanandmortgagecalculator.co.uk` is B2,
has an empty `github_repo`, and serves a sitemap today — so the path demonstrably works.

## Where to wire it — a real decision, not a detail

| candidate | what it gives you | what it costs |
|---|---|---|
| the **page-deploy / rerender** path (`page-rerender`, `rerender-pages` handle `needs_rerender`/`page_rerender`) | the file refreshes whenever a page changes — closest to "always current" | fires very often; the probe is a GET per URL, so a 98-page site is 98 requests each time |
| a **scheduled sweep** | cheap, bounded, catches retracted pages dropping out | a new page is missing from the sitemap until the next run |
| **both** | correct in both directions | two callers to keep honest |

SEO-002's own note leans to both, and says why: a new page and a retracted page are different
events. **Start with one, prove it at the artefact, then add the other** — and whichever you pick,
the first thing to check is the probe's cost at fleet scale, which is the risk the submission
itself flagged.

## Then resubmit

`RESUBMIT_CORR=8a004aab-be85-4d6d-bdb1-4fb114f1d64b` with the wiring edit added. The submission
JSON is at
`~/.claude-scratch/.../scratchpad/council_sitemap.json` — if that scratchpad is gone, the plan is
reconstructible from `render_sitemap_action.go`'s header, which carries the whole argument.

## Definition of done, so it cannot drift

Not "the action is wired". **A site that did not have a sitemap has one, fetched and read.**
As of 2026-08-24 only **8 of 25** live sites serve a sitemap of ours — that number is the
before-figure, and re-measuring it after is the proof.
⚠ **Judge the body, not the status code**: `adversecreditmortgage.co.uk` returns 200 on
`/sitemap.xml` carrying a single `<loc>` for `/lander` — the **parking provider's** file, with no
matching `pages` row. That one cost me a wrong count already.

---

# 2. ✅ `bugs_open/311` IS CLOSED — the fleet precondition is GONE

Verified at the artefact 2026-08-24, not taken from the label: `remortgagecalculator.uk` serves
**6 `<input>` elements, 69,704 bytes**, with real calculator fields (`calc-loan-amount`,
`calc-interest-rate`, `calc-term-months`, `calc-monthly-output`). On 2026-08-18 the same page was
40,726 bytes with **no calculator at all** — the missing tool the owner spotted. The symptom is
healed, not just the file moved.

The closing commit (`6e2d21a70`) records seven diversions, every incumbent component
byte-identical afterwards, both halves re-verified through v1.0.1322. RFC_036's tool half is live
too, so the pair that had to ship together did.

**⚠ This corrects the 08-21 summary and the 08-20 handoff**, which both say the halt turns on one
unfixed defect. It does not any more.

**Residuals were deliberately NOT swept under the close** (checked 2026-08-24): `bugs_open/345`
open (fix live, demand unproven) · `bugs_open/337` open · `bugs_open/283` open · `bugs_open/315`
closed. **`283` matters more than it did**: it is the design question about reusing an interactive
component on one page, which is the nearest existing work to the Christmas card-sender (§5).

# 3. THE HALT — half lifted

`remortgagecalculator.uk` is **unlocked**. `adversecreditmortgage.co.uk` is **still locked** with
its queued items held (as of 2026-08-24). Nothing technical blocks it now; it is the owner's call.

# 4. WHAT IS BUILT AND LIVE

- **Brief writer** (`brief-writer`, migration `510`, register **BLD-024**) — researches a subject
  from the domain name, writes a comprehensive aspirational `mission_brief`, then **holds** at
  `status='needs_human_review'` so nothing builds until a person releases it. Fire with
  `scripts/fire-brief-writer.sh <domain> [direction]`. Proven twice.
- **The register is a database** (`511`/`512`/`513`, **BLD-025**) — **189 rows as of 2026-08-21**.
  `brief-writer` reads this domain's entry and its family siblings. **PROVEN USED, not merely
  received**: the `buytoletcalculator.uk` brief wrote *"deliberately does not serve the general
  homebuyer (M2 territory), does not compare rates (M3 territory), and does not go deep on
  company/SPV structures (M12 territory)"* and escalated the one boundary it could not settle.
  ⚠ **`raw_md` is AUTHORITATIVE**; the typed columns are an index over it.
  ⚠ **`REGISTER_positioning.md`'s fate is undecided** — do not edit both.
- **Regulated-identity guard** (**CGV-033**) — live; all **3 persist branches as of 2026-08-24**
  guarded, with `TestEveryPersistingEditBranchIsGuarded` so a fourth fails at birth.
  ⚠ Three stated blind spots: **`pages.title` and JSON-LD are scanned by nothing**, chrome is
  ungated. `meta_description` IS covered. Council round 4 pending on `aac38d5b`.
- **`render_sitemap`** — built, registered, **uncalled**. See §1.

# 5. NOT STARTED

- **The 22 hosted-site remakes** (`DECISION_2026-08-20_remake_the_hosted_sites.md`). 3 protected:
  `leopardess.co.uk`, `leopardess.uk`, `cartoon.co.uk`. Best first picks are the single-pagers with
  strong generic names; **do not start with `businessinsurancequotation.co.uk`** — it is insurance,
  so it inherits the compliance layer, and it is the largest.
- **The Christmas card sender** (register G3/G4) — one component, two skins. **The estate's first
  mechanism that takes a stranger's input and delivers it to a third party**; design the delivery
  half first, because an open "send to any address" form is a spam relay. Read `bugs_open/283`.
- **Domain search-and-buy for third-party customers.** Nothing here spends money yet.
- **21 portfolio domains have no register row** (as of 2026-08-21) — `.uk` siblings the document
  itself never names.

# 6. TRAPS

- **A registered-but-uncalled action is not a mechanism.** §1 is this, caught by the council.
- **A binary carries ONE commit stamp** — grepping it for YOUR sha always fails. Use
  `git merge-base --is-ancestor`. The `build provenance` log line scrolls within hours.
- **A source-scanning test finds its own words in the file's COMMENTS and passes while the code is
  wrong.** Extract the query/switch first. Mine failed exactly this way and had to be rewritten.
- **Mutation-prove with a change that still COMPILES** — a build failure is not the test catching
  anything. Mine did that too.
- **On this tree a transient build failure is often another session's half-written file.** Verify
  from the committed tree: `git archive HEAD | tar -x -C $(mktemp -d)`.
- **One client's view of a CDN is not the site's state**; a fresh `Last-Modified` on a stale body
  means the delivery path, not the page.
- **`rm` the temp file before `curl`** — a failed fetch leaves the previous body in place and it
  reads as a fresh result.
- A parked domain returns 200 on every path — read the BODY.
- Discovery queries must stay **< 200 bytes** or `web_search` drops them and blames config keys.
- A migration NUMBER identifies nothing — ask the ledger by exact filename.

# 7. Files of record

**Cold start:** this file → `SUMMARY_2026-08-21_…` → `PLAN_2026-08-19_one_flow_three_brief_sources.md`
→ `README_where_we_are.md` (owner's log) → `NOTES_portfolio_positioning.md` (evidence).
**Decisions:** `DECISION_2026-08-20_remake_the_hosted_sites.md` · `REGISTER_positioning.md` ·
`RFC_037` (binding collision check, still open).
**Sitemaps:** `render_sitemap_action.go` (+ test) · register **SEO-002** · `HOSTED_domains_for_owner_decision.md`.
**Regulated:** `claims_regulated.go` · `section_editor_regulated_guard.go` · `cmd/regcheck` ·
`scripts/regulated/record_attestation.py` · **CGV-033**.
**Domains:** `RUNBOOK_domain_inventory_and_classification.md` · `RESERVED_test_domains.md` ·
`scripts/domains/`.
