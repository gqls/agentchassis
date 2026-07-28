# PLAN 2026-07-26 — making vetcomparison.uk worth visiting

**Status: ACTIVE. Supersedes nothing** — `PLAN_2026-07-15_rebuild.md` still holds for Phases 0–5
of the rebuild; this plan is what to do now that those landed and the site is still thin.

**Origin.** Owner, 2026-07-26: *"I don't think the site is strong enough yet"* — and a decision to
let the CMA consultation deadlines go (funding 30 Jul, substantive 20 Aug). **Both consultation
responses are dropped. Do not spend further time on them.** Owner then chose all four candidate
directions; this plan sequences them.

---

## The real diagnosis — it is not a missing-pages problem

The instinct is to fill the homepage's nine 404s with pages. That would make the site *look*
complete without making it *stronger*, and it is the wrong read. The site is thin because
**nearly everything useful we already hold is unpublishable under our own provenance rule**:

| asset | held | why it cannot publish |
|---|---|---|
| practice prices | 762 current rows | `[VERIFIED]` **0** carry a source URL — unrecoverable, they predate the rule |
| ownership (`group_name`) | 1,102 practices | no evidence trail; **contradicts `is_independent` on 870 practices** — all 55 Medivet, 27 CVS Vets, 34 Vets4Pets flagged "independent" |
| practice detail (species, emergency cover, accreditations, accepting-new-clients) | 2,781 rows, last touched 2026-03-18 | no source URL, no observed-at; same Feb–Mar scrape |
| Companies House companies | 5,798 (SIC 75000, all active) | only **158** matched to a practice deterministically (tier1/tier2 ≥0.91); 542 more at 0.50–0.81 by postcode-proximity or LLM — not publishable as fact *on this site* |

So the live site is: 2,109 practice names + postcode + an outbound website link, three sourced CMA
guides, and a working CMA news feed. Everything else is either absent or barred.

`[VERIFIED]` none of the unpublishable material is live — the directory export publishes exactly
six fields (`id, name, postcode, location, website, is_claimed`) and no ownership or price field.
The problem is under-publishing, not mis-publishing. **Nothing here is a correctness emergency.**

> **CORRECTED before execution, 2026-07-26.** My first version of this plan said the unlock was
> "apply seed `082_company_number_scraper`, which was never applied". **That was wrong twice over
> and would have wasted a cycle.** Reading the seed to the end (the landmine check) shows it
> *creates the agent and then deletes it*:
> ```sql
> -- forget it, we'll use the regex in the verifier
> DELETE FROM scheduled_tasks   WHERE name = 'ch-scrape-company-number';
> DELETE FROM agent_definitions WHERE type = 'ch-company-scraper';
> ```
> So applying it is a no-op, the standalone scraper was **deliberately retired**, and its
> replacement *was* built — `StoreBusinessVerification` takes the LLM-extracted
> `registration_number` and falls back to a **regex** over the scraped content
> (`business_intel_actions.go:350-372`, `updateBusinessFields` line 909).
> **The real reason nothing is populated is below, and it is much simpler.**
> The check that caught it: read the whole seed before applying it, per the seed-037 landmine.

## The actual unlock — the data pipeline has been switched off since March

`[VERIFIED]` **Every vet data-collection task is disabled, and nothing has been verified since
2026-03-18:**

| task | agent | enabled | last ran |
|---|---|---|---|
| `vet-sweep-continue` | vet-pipeline-orchestrator | **false** | 2026-03-11 → 03-17 |
| `vet-batch-verify` | vet-batch-processor | **false** | 2026-03-19 |
| `ch-vet-collect` | ch-collector | **false** | 2026-03-19 → 03-29 |

`SELECT max(last_verified_at)` = **2026-03-18**; verified in the last 30 days = **0**.

They were switched off — correctly — during the July fabrication remediation. **Everything since
has been publishing-side work** (exporter, adoption, guides, news feed). Nobody restarted
collection. That single fact explains the whole table above: the data is thin and stale because
it stopped being collected four months ago, not because the machinery is missing.

**The machinery is present, wired, and deployed:**
- `StoreBusinessVerification` already captures the company number — LLM-extracted
  `registration_number` first, then a deterministic **regex** fallback over the scraped page
  (`business_intel_actions.go:350-372`). Regex means no fabrication surface, which matters more
  here than anywhere (`bugs_open/020`).
- The same path already refreshes `group_name` and the practice-detail facts.
- It already writes `data_observations` **including a `source_url` column**
  (`business_intel_actions.go:328-334`) — the provenance hook exists in the write path.
- `ch_local_match` is *"No API calls — pure SQL + Go scoring. Safe to re-run."*
- `[VERIFIED against pod `agent-chassis-5df8868c9f-j4j6s`]` `ch_scrape_company_number` → 1,
  `ch_local_match` → 1, `company_number_scraped` → 8, bogus control → 0.

**The one thing we do not know, and it decides everything downstream:** all 2,970 existing
`data_observations` rows have an **empty** `source_url`. The column is in the INSERT, so either it
was added later or `sourceURL` resolves empty at runtime. **If a live run still records no source
URL, restarting the pipeline refreshes the data but leaves it exactly as unpublishable as it is
today** — and the fix becomes a Go change (council gate, image build, roll) before any restart is
worth doing. This is the first thing to measure, on a handful of practices, before anything else.

> **ANSWERED 2026-07-26 ~22:45 BST (session "bugfix 061") — and neither of the two hypotheses
> above was the cause.** `source_url` is **structurally guaranteed empty**; no live pilot was
> needed to establish it. The writer reads `source_type/source_name/source_url` out of
> `verResult` — the **LLM's own output object** (`business_intel_actions.go:322-324`, with
> `verResult := extracted["verification_result"]` at line 180). The `extract_and_reconcile`
> prompt requests six sections (`business`, `vet_details`, `vet_staff`, `prices`,
> `confidence_score`, `extraction_notes`) and **no source field**. And `store_results`'
> `input_fields` are `["business_id","verification_result","task_id"]` — **`scraped_data` is
> excluded**, so the URL `scrape_web` actually fetched cannot reach the writer at all.
> Empirically: `raw_data` *is* `json.Marshal(verResult)`, and **0 of 2,970 rows carry a
> `source_url` or `source_type` key** — the fields are absent, not blank.
>
> **So the branch is decided: "Provenance empty → Go change before any restart."** Note the fix
> must **not** be "ask the LLM for the source URL" — that makes provenance a model claim, the
> exact class this site was remediated for. Thread the fetched URL through instead.
> Mechanism filed to the diagnosis loop before being asserted:
> `SUBMISSION_CORR = e6580fe5-7537-4eba-a3aa-7863ce4dbfc7`.
>
> **Step 1's pilot was also run, read-only, and step 4's question is partly answered** — see
> `NOTES_vetcomparison.md` (2026-07-26 §2 and §6). Company-number hit rate, deterministic sample
> of **100** (the 25-run figures of 16%/28% are superseded): **22/100 (22%) homepage-only**,
> **30/100 (30%) if the scrape also read legal/terms pages** (~95% intervals ~14-31% and ~21-40%).
> **16 of 20 distinct numbers resolve to a real `ch_vet_companies` row.** The ownership signal is
> the prize: 30 hits are only 20 companies — **8 of the 100 sampled practices are VetPartners**
> (`10084952`), 3 are CVS (`03777473`), so 13 of 30 belong to three groups. That is evidenced
> ownership against P2's 870-practice `is_independent` contradiction.
>
> ⚠️ **`[UNSETTLED]` whether production's extraction sees what the probe saw.** Firecrawl's
> `onlyMainContent` is set to `false` in code but only *sent* when true (`firecrawl.go:77-111`),
> so with no `scrape_config` the key is omitted and Firecrawl's own default applies. If it strips
> footers, the real rate is lower and adding page fetches will not help. Ambiguous against 2,452
> stored samples (75% retain footer nav; 0 contain registration text). **One real verification
> settles it — do that before implementing `bugs_open/101` candidate 2.**
>
> > **CORRECTED within the hour, same session — I first wrote that the 28% was "a config-only
> > change" by widening `follow_links`. That is FALSE and it is the more important finding.**
> > `follow_links` **is not read by any Go code in this repo** (`grep -rn follow_links --include=*.go`
> > → no hits). `WebscrapeAction` (`webscrape_actions.go:27-147`) honours only `url_field` /
> > `url` / `action` / `upload_results` / `scrape_config`, and dispatches **exactly one URL** to
> > the webscrape adapter. So **four of the six keys on the `scrape_website` step are inert**:
> > `max_pages: 3`, `follow_links: [fees, prices, about, team, contact, services]`,
> > `extract_mode: "text"` and `fallback_url_field: "search_results.results.0.url"`.
> > The step reads as a six-page crawl with a search-result fallback; it is a single homepage GET.
> >
> > Consequences: (a) the **16% is exact**, not an approximation — homepage-only *is* production;
> > (b) reaching 28% needs a Go change or extra workflow steps, so it should be **bundled with
> > the provenance fix** into one council round, one build, one roll — not treated as a free win;
> > (c) `fallback_url_field` being dead means the intended "no website → use the top search
> > result" path silently never runs (moot today: all 3,419 rows have a `website_url`).
> > **The cheap check that would have caught it before I wrote it: grep the config key in the Go
> > source before calling any config change a win.** Logged in `WRONG_CALLS.md`.

---

## Phases

### P1 — Prove provenance on a pilot, THEN restart collection — DO FIRST

> ## ⛳ 2026-07-28 (from the `bugfix_100_101_scrape_provenance` lane) — **YOUR P1 BLOCKER IS LIFTED. Step 2's branch is already decided, and step 3's check is already done.**
>
> Left here rather than in a handoff because this is the file that branches on it.
> Nothing below is asked of you — it is all *already done*, and it changes what P1 is.
>
> **"Provenance is structurally impossible" is no longer true.** Step 2's second
> branch — *"Provenance empty → a Go change to thread the source URL … before any
> restart"* — was **taken and completed by another lane**. Platform code, council
> gate (APPROVED at round 3, corr `f4cf0aab-5a08-4475-91ea-fa831cff323c`), image
> built, rolled. `bugs_closed/101` and `bugs_open/100`.
>
> - **Provenance is recorded from the FETCH, not asked of the model.** That was the
>   hard requirement in your own §"must **not** be ask the LLM for the source URL" —
>   it was honoured. The fetcher was already returning the answer; nothing read it.
> - **The DB now refuses an observation that cannot cite its source.**
>   `data_observations_provenance_not_empty` CHECK on
>   `(source_url, source_type)`, applied 2026-07-28 18:30Z, **proven to enforce by
>   negative control** (an insert with empty provenance errors), and `NOT VALID` —
>   so your **2,970 historical rows are untouched** and no backfill is owed.
> - **Your step 3 is DONE and its figure was stale.** The doc says
>   `vet-practice-verifier` is on `v1.0.1169`; checked 2026-07-28 19:2xZ, all three
>   relevant agents carry the fix:
>   ```
>   business-intel v1.0.1192 · domain-research-classifier v1.0.1192 · vet-practice-verifier v1.0.1192
>   ```
>   So the spawned worker pods get the provenance code, not just the deployment.
>   **This mattered more than it looks:** had `image_tag` still been pre-fix while
>   the CHECK was live, every insert from your first restarted crawl would have
>   *hard-failed* — a silent data defect converted into an outage. It was checked
>   for exactly that reason. Re-check it after any roll rather than trusting this line.
>
> ### What P1 now is
>
> **Your pilot is the acceptance test for `bugs_open/100`, and running it closes
> that bug.** No such run has ever happened — collection has been off since
> **2026-03-18**, which is still the newest `collected_at` as of tonight. Both
> columns matter:
>
> ```sql
> SELECT source_url, source_type, raw_data ? 'source_url' AS llm_claimed, collected_at
> FROM business_intel.data_observations ORDER BY collected_at DESC LIMIT 5;
> ```
>
> **`source_url` non-empty AND `llm_claimed` still FALSE.** A populated column alone
> proves nothing about *where it came from* — if `llm_claimed` is true, the fix in
> place is the explicitly-rejected "ask the model" candidate and must be reverted.
>
> **If provenance comes back empty**, do not guess: grep the chassis log for
> `no fetch provenance available`. That warning was added so a *shape mismatch* is
> distinguishable from a genuine absence, and it names the field it looked in. The
> live shape of `scraped_data` is `[UNVERIFIED]` — traced through code as `data.url`,
> never observed, because no run carrying it survives the retention clock. Your
> pilot is the first opportunity to observe it.
>
> ### Two things your pilot will be the first to exercise
>
> 1. **The `bugs_closed/062` payload watch has a zero denominator.** Three steps now
>    receive full pages for the first time (`only_main_content: false` became
>    expressible). The post-roll watch reads **0 errors over 0 scrape attempts** —
>    measured tonight, `grep -c "Starting scrape"` = 0 — so the clean log is
>    currently unfalsifiable, not reassuring. **Count attempts before reading the
>    error count**, and watch during your pilot:
>    ```bash
>    kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h \
>      | grep -c "Starting scrape"                       # denominator FIRST
>    kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=1h \
>      | grep -i "Message Size Too Large\|Failed to produce"
>    ```
>    The 062 failure is **silent to the caller** (~12 min of timeout retries), so an
>    absent workflow error proves nothing. Mitigation if it bites is config-only, no
>    roll: set `scrape_config.formats` on the offending step.
> 2. **Your `max_pages: 3` / `follow_links` finding was right, and is now VISIBLE
>    rather than fixed.** Those keys are read but honoured **only on a crawl**;
>    `vet-practice-verifier/scrape_website` is a single-page scrape, so it still
>    fetches one page. It now *warns at runtime* and appears in
>    `scripts/audit-config-keys.sh`'s CONDITIONALLY HONOURED section. **The owner
>    ruled 2026-07-28 that it STAYS that way** — switching it to `action: "crawl"` is
>    a deliberate behaviour change for whoever owns the agent, not a side effect of a
>    bug fix. So your 16% figure remains exact, and reaching 28% is still a real
>    change, not a config flip. (`fallback_url_field` and `extract_mode` *were*
>    implemented and now work.)

**Do not simply re-enable the three tasks.** Re-enabling is one UPDATE and it is the wrong first
move: it would spend a fleet-wide crawl to produce data we still could not publish, and it would
overwrite the current rows while doing so.

1. **Pilot: verify ~10 practices and inspect what lands.** The question is narrow and decisive —
   *does a live verification write a non-empty `source_url` into `data_observations`, and does it
   populate `company_number_scraped`?* Nothing else in this plan can be sized until that is known.
2. **Branch on the answer:**
   - **Provenance recorded** → the pipeline is publishable as-is. Restart collection (staged, not
     fleet-wide at once), then P2/P3 proceed on fresh, sourced data.
   - **Provenance empty** → a Go change to thread the source URL and an as-at date through the
     verification write path, **before** any restart. Platform code → council gate → image build →
     roll → re-pilot. Slower, and the correct order: a crawl that cannot produce publishable data
     is a wasted crawl.
3. **`image_tag` check before dispatch** — spawned worker pods run `agent_definitions.image_tag`,
   **not** the deployed image. `vet-practice-verifier` is on `v1.0.1169`; confirm that carries the
   regex path before trusting a pilot result.
4. Then `ch_local_match` (free, re-runnable, no API) over the newly-numbered practices → measure
   how many gain an **authoritative legal identity**.

**Success = a measured number, not a built thing.** Specifically: what fraction of practices end
up with a company number *and* a source URL. If that fraction is poor, P2 shrinks and P3's
ownership panel shrinks with it — and we want to know that before building either.

**Outbound-request note:** this crawls real businesses' websites. The pilot is ~10 fetches. Scaling
to 2,109 is an owner call, not a silent escalation — the existing task config is already polite
(100 per batch, 1 req/sec, 10s timeout).

### P2 — Ownership, derived not asserted
- Rebuild ownership from the P1 evidence chain: practice → self-declared company number → CH record.
- **`is_independent` becomes derived, never asserted.** Today it contradicts `group_name` on 870
  practices; a derived field cannot contradict its own source. Retire or recompute the flag.
- **Normalise groups.** Currently the same owner is split several ways — Linnaeus ×4 variants
  (`Linnaeus Veterinary Ltd (A Mars Company)` 26, `Linnaeus (Mars)` 17, `Linnaeus Group` 16,
  `Linnaeus Veterinary Ltd` 13 = 72 practices), `CVS Vets` 27 + `CVS Group` 15, `Vets4Pets` 38 +
  `Vets for Pets` 27 (both Pets at Home). `Independent` (155) is a category masquerading as a group.
- **Publish only what is evidenced, and disclose the method.** A postcode-proximity or LLM match
  is not "who owns this practice"; it is a guess, and on this site a guess must not ship as a fact.
- Widening the directory export to carry ownership is a **platform** change (the record struct is
  hardcoded Go, `directory_export_action.go:255-280`, shared across verticals) → must be
  config-driven and must go through the council gate.

### P3 — Practice pages
- **Reuse, do not build**: the entity-page machinery is proven live — relojistas.com has 8
  `entity-page` + 1 `entity-directory` deployed, vonc.com 8. vetcomparison already has both rows
  scaffolded (`practice`, `directory-index`) at `build_status='planned'`, 0 sections, sitting in
  the owner review queue since 07-17. URL derivation is handled (`page_canonical.go:200-214`).
- **Do NOT build 2,109 pages in one go.** No site on this fleet has more than 8. Start with a
  small proving set, measure build cost and render-queue impact, then scale.
- Page content = evidenced facts only: contact/address/website (already published), ownership from
  P2, and practice detail **only where re-verified** with a source URL and an as-at date.
- The March-2026 detail (species, emergency cover, accreditations) is *not* publishable as-is —
  it needs re-verification in the same crawl pass, or it stays off the page.

### P4 — Compliance deadline calculator (independent — can run in parallel)
- Buildable, but **relative periods only**: 3/6/9/12 months from the Order being made, split Large
  vs Small business. `[VERIFIED 2026-07-25 from the draft Order PDFs]` every absolute date and
  every monetary cap is a **bracketed placeholder** — `[£21]`, `[£12.50]`, `[X March 2027]`,
  `[X December 2026]` — explicitly "adjusted for inflation … before the Order is made".
- **Emitting a calendar date would assert something the CMA has not fixed** — precisely the class
  of claim this site was remediated for. Takes a "date the Order is made" input; ships no date
  until the Order is made.
- Serves the practice-side audience, which is the commercial side (claim listings, B2B guides).

### P5 — The homepage's dead promises (finish last, but land the cheap half early)
`[VERIFIED live 2026-07-25]` 9 broken internal links; six are *every anchor* in `info-card-grid`.
- Cheap and immediate, do early: an honest `/about-pricing` explaining plainly **why we publish no
  prices yet** is a genuinely good page, not an apology. Same for `/about-ownership-disclosure`.
- The rest repoint as P2/P3 make their destinations real. `/search` → the directory index;
  `/claim-listing` → the claim flow (Phase 3 machinery already exists: `claim_requests`, routes
  live at `58e2a837`).
- **The platform half is not ours.** `bugs_open/023` is OWNED and active; the finding that
  `ctaFieldNames`'s `[2]string` cannot address `cards[].link_url` is contributed there. Do not fork.

---

## Constraints that bind every phase

- **No claim without provenance.** `seed_import` never publishes. This is the rule the whole plan
  is shaped around, not an obstacle to route past.
- **Platform changes must be generic** (owner standing constraint) and go through the **council
  gate** (`platform/`, `internal/`, `pkg/`). Budget ~30 min per run.
- **Go changes are inert until an image is built and rolled; DB config is live immediately.** Seeds
  and migrations are the fast path — prefer them.
- **Never re-plan this site** (`bugs_open/001`).
- **Renders delete-and-recreate components** (`save_page_sections_action.go:498`) — `content_data`
  survives an assemble-only render but the **row id does not**, and a full content re-render
  regenerates copy. So site content must be fixed at its source, never in `page_components`. This
  is what defeated two attempts at the dead search grid.
- **`med_*` is another thread's** (med-retailer arm, `bugs_open/061`). Do not touch.
- **Check the queue before dispatching.** Coverage check run 2026-07-26: clean — only the 5
  standing owner-review items, no other session mid-flight on this site.

## Verification

- **P1**: `SELECT count(*) FROM business_intel.businesses WHERE company_number_scraped <> ''` —
  from 0. Then the join to `ch_vet_companies` for authoritative identities.
- **P2**: zero rows where a published ownership claim lacks an evidence row; the 870-row
  contradiction goes to 0 **by derivation, not by UPDATE**.
- **P3**: pages `build_status='deployed'`; `curl` the live URL, not the row —
  *trust the rendered artefact, not the status*.
- **P5**: `curl -s https://vetcomparison.uk/ | grep -oE 'href="/[^"]*"'` → every target 200.
- Throughout: verify Go against the **running pod**, never git, never the tag.

## Owner decisions still outstanding

1. The 5 standing review items (3 zero-section pages, contact email fact, deadline-calculator
   build-or-cancel — **P4 answers the last one: build it, relative-only**).
2. Whether to scale the P1 crawl to all 2,109 after the pilot hit-rate is known.
3. `/claim-listing` and `/search` destinations — resolved by P3/P5 rather than decided up front.
