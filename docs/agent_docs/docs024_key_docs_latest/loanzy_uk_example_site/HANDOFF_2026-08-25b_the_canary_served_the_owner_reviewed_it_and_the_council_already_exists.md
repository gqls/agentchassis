# HANDOFF 2026-08-25b — the canary built and served, the owner reviewed it, and the reviewer he wants already exists and is switched off

> **SUPERSEDED 2026-08-25 evening by `HANDOFF_2026-08-25c_376_approved_the_council_has_seats_and_the_evolutionary_switch_awaits_one_word.md`** — §6's to-do list is done (376 submitted AND approved; RFC written; four seats built); read 25c first.

**Lane:** `loanzy_uk_example_site` — the greenfield route (submit a domain, measure what the platform
produces). **Supersedes `HANDOFF_2026-08-25_the_route_is_gated_on_376_and_the_canary_is_the_only_thing_left.md`**
(same directory; bannered). That file's §3 (the `376` mechanism) and §5 (the harness) still stand;
everything about "the canary is the only thing left" is now history.

**Read in this order on a cold start:** this file → `REFERENCE_2026-08-25_site_acceptance_council.md`
→ `OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md` → `bugs_open/376` §10–11.
`NOTES_loanzy_uk_example_site.md` has the full running record with every misstep; `README_where_we_are.md`
is the owner's plain-prose log and he reads it.

---

## 1. What is LIVE and VERIFIED `[all MEASURED 2026-08-25]`

| thing | state | how verified |
|---|---|---|
| **`homegarden.uk`** — the owner-authorised canary, site `5904bd0f-33fd-4212-9c1b-50b28fe72fdb` | **21 pages planned, 20 serving over HTTPS**, 1 × 404 (`/blog/blog-post.html`, the one page with `sections_len = 0`, predicted before it failed) | invented-path control **404**; every page fetched |
| its Cloudflare zone `252c10abde85a6985392a084f68f9235` | **active**, NS `alexis`/`leah`, 2 A records + 2 worker routes **diffed identical** to `garden-tools.uk` | re-read the stored zone, not the POST receipts. **First zone this estate created via the API.** Owner set the NS himself |
| the `bugs_open/381` calendar | **real**: `<ol class="period-cal__list">`, 12 `<li>`, January–December, **12 non-anchor month names** on `/index.html` | element-scoped read; `381` CLOSED by its lane on two independent instruments |
| chassis `v1.0.1337` | `380`'s Go practice-claims family **present at the binary**, both replicas | capability probe with a present-control and an absent-control |
| `garden-tools.uk` | **unchanged**: 7 serve / 5 × 404, baseline intact (0 tables, 0 content lists, 0 `<strong>`) | re-run this morning; **owner: "wait with garden-tools.uk"** — re-plan AUTHORISED but HELD |

---

## 2. The owner's rulings today, verbatim in substance — do not re-ask any of these

| ruling | where recorded |
|---|---|
| **Retracted** the parked-row authorisation on garden-tools (`206`'s proof rides a build instead) | 08-24 handoff banner; CONTRIB to `206` |
| **Card composition: more than 4** card sections before something breaks them up | `OWNER_REVIEW` §0 |
| **Re-plan of garden-tools is fine — but WAIT** (his later word) | `OWNER_REVIEW` §0, §10 |
| **Switch on all the disabled agents** — *"we'll need to fix or further develop them as necessary"* | routed to `vigilant_designer_offer_analysis` |
| **Give the visual designer a dispatch path** | same |
| **Yes, an in-body image slot as default** | same |
| **Copy lane: deep refresh first, then audit EVERY prompt** in DB and code for AI register | routed to `copy_quality_two_stage`; ACKNOWLEDGED, in progress |
| **The site acceptance council runs AFTER THE FACT in the improvement loop**, routing to existing agents; **N = 6** structures; *"I like your suggestion"* | `REFERENCE_2026-08-25_site_acceptance_council.md` §2 |

---

## 3. The day's findings, each with its home

**3a. `bugs_open/376` — now has a verified fix DESIGN and a second failure mode.** Hop two on the
canary **passed** (refused host absent — one draw, not a rate), but exposed **(b): a crawl can
`success: true` with `source_count: 0` and the chain proceeds.** So a floor counted on step success
would ship a landscape from no research with every step green — worse than today. Design in §11 of
the bug: `error_step` per crawl → its own format step (verified: the format action returns a graceful
`content_quality: none` with a nil error); floor **2-of-3 on `source_count`**, written **without
parentheses** because `conditional_branch` strips them (LANDMINE filed), `fail_on_non_numeric: true`.
**NOT SUBMITTED to the council. That is this lane's next concrete job.**

**3b. The owner's review of homegarden** — all measured, all routed. Calendar with **0 links out**;
**1 `<img>` per page** (the logo); *"this April"* in August; about page **14 of 17 headings about our
own methodology**; comparisons page with no comparisons; every page carrying a **"Get Started" →
contact** button on a gardening site (mine, not his). `OWNER_REVIEW` §1–§7.

**3c. Three of his proposed agents already exist and NONE fired.** `visual-designer`: **no dispatch
path at all** — no live config names it, nothing schedules it, 0 LLM calls ever (positive control
24,217). `offer-analyser`: three carriers **disabled** since 08-14/15. `experience-planner`: same as
the designer. **The honest sentence: the designer did not do a bad job; it was never asked.**
`OWNER_REVIEW` §9, verified twice by two lanes.

**3d. No research ran, and the classifier could not have proposed a directory.** `research-agent` is
a complete 9-step pipeline with **0 LLM calls in its life**, because **0 of 21 pages request
research** — spawned, never given work. The classifier reads 15 layout categories at query time and
**none is `directory` or `news`**; and his own brief was anti-commercial, so two independent blockers.
`OWNER_REVIEW` §10.

**3e. The news feed is a BROKEN ROUTE with an exact cause.** Mechanism is healthy (7 sites, deployed
today, 20 outward links on dartsonline, carrier fired 14:45Z). But `find_news_sites` requires an
active `content_sources` row, the orchestrator's own `seed_sources` step runs *after* selection, so
**an unenrolled site can never become enrolled.** 9 of 51 sites enrolled; homegarden 0. CONTRIB to
`bugs_open/316` (same function, distinct defect). `OWNER_REVIEW` §11.

**3f. The pattern under 3b–3e:** **a greenfield site is born without the prerequisites that several
capabilities silently gate on** — evidence base (`380`), posts to list (`384`), feed sources (`316`),
research (nothing files it). Every failure is an absence, not an error.

**3g. The improvement loop IS the site acceptance council, dark since 08-17 12:30Z.** It already
carries brief-fidelity, completeness, quality, design-audit, offer-analyser, **and `enrich_news_feed`
+ `enrich_directory_features`** — his two asks as steps. Ran 08-13→08-22, 10 items on 7 sites. Two
traps: **both enrichment steps have `error_step`s that swallow failure**, and brief-fidelity judges
against the brief, which can itself be the problem. The route is **switch on + four seats**
(prerequisites, promise, structure ≥ N with recorded refusal, reader) — `REFERENCE` §3–§4, §10.

---

## 4. Traps found today — the ones that will bite the NEXT session, not the ones already in LANDMINES

- **`after_test.sh` now gates every HTTP reading on an invented-path control.** A newly built domain is
  parked by default and 200s every path — my harness would have reported homegarden **21 of 21
  perfect** while it served nothing. Merged into the existing 08-23 LANDMINE (I filed a duplicate
  first; `head -10` hid the original). **A `000` from the control is UNREACHABLE, not a pass** — for
  a window the site was live on port 80 and dead on 443 while the certificate issued.
- **Counts are chrome-stripped.** Every page carries all 12 month names as **menu links**, so a raw
  count read 12 on `/contact.html`. Strip anchor text; count elements bearing a class as a **whole
  token** (BEM inflates substring counts 9×); strip `<style>` **unconditionally** because the
  contaminant is site-dependent (inlined CSS on one site, external on another).
- **Two of three Cloudflare tokens are read-only.** `~/.config/cloudflare/portfoliotoken` is the one
  that writes. The runbook's "prove the token" check passes on a token that can write nothing.
  `ls` the directory before concluding anything about permissions.
- **The account uses TWO nameserver pairs** (29 × alexis/leah, 11 × betty/ivan). Take them from the
  create response; never quote the remembered pair.
- **`conditional_branch` ignores parentheses** — `(A OR B) AND C` evaluates as `A OR (B AND C)`,
  silently. 140 live steps use it. LANDMINE filed.
- **`sections_len = 0` is the class (b) predictor**, not `page_type`. I predicted 17 failures from a
  role; the build refuted me in four minutes, and the real predictor named the one page that failed.
  `capture_reconcile_mint.sh` prints the risk set.

---

## 5. Peer coordination — who has what

| lane | holds | state |
|---|---|---|
| `bugs_open/381` | CLOSED on the canary; 15 wrong calls tallied between us, **none caught by its author** | done |
| `bugs_open/206` | two CONTRIBs: the authorisation retraction; the canary cannot test them (0 entity-directory pages) + `sections_len=0` as their class (b) predicate | theirs |
| `bugs_open/384` | CONTRIB: 17 pages with a listing that renders nothing; positive control 12 cards / 1 section; **retracted** my link set-difference (tautology on a site whose nav links every page) | theirs |
| `bugs_open/316` | CONTRIB: the news bootstrap catch-22; closing it could take their candidate set from 9 toward 51 | theirs |
| `vigilant_designer_offer_analysis` | all three switch-on authorisations; the two fail-open `error_step`s; `site_reviewer` role unresolved | acting |
| `copy_quality_two_stage` | the deep refresh + the every-prompt audit; second owner escalation today (finetuning) — same shape | in progress |

---

## 6. What is next on THIS lane, in order

1. **Submit `bugs_open/376`'s fix to the council** — design is verified in §11 of the bug; write the
   migration and the submission JSON; `Council-Submitted:` trailer; run the three disconfirming tests
   (refused host, succeeds-but-empty host, below-floor run) when it applies.
2. **Draft the RFC for the acceptance council** — the `REFERENCE` is written to be lifted from; the
   seat definitions ARE the benchmark and the architecture seat should rule on them, on `380` D1, on
   the taxonomy gap, and on superseding BLD-006. Owner has approved the direction, not yet asked for
   the RFC in so many words — **confirm before spending the round.**
3. **When `improvement-sweep` is back on**, run `after_test.sh homegarden.uk` and compare what the loop
   files against what the harness finds. The harness is the promise seat's prototype.
4. **`garden-tools.uk`: HOLD.** Re-plan is authorised and deferred by the owner. Do not touch.
5. **Do not dispatch a new domain until `376` ships** — hop two is a 4-in-5 kill on the one vertical
   measured, and a no-brief run has to go through it.

---

## 7. Falsifiers

- `improvement-sweep` enabled and the loop filing nothing routable → `REFERENCE` §11 first bullet.
- `homegarden.uk` control path starting to 200 → DNS or worker regressed; the HTTP census is blind again.
- `bugs_open/376` taken by another lane → contribute, do not compete; `who-owns` only sees my commits.
- A new greenfield build appearing → it is a `376` draw and a council census candidate; watch hop two.
- `garden-tools.uk` `last_reconciled_at` moving off `2026-08-23 20:15:50` → someone re-planned it;
  the baseline is gone and every doc quoting it needs a date check.
