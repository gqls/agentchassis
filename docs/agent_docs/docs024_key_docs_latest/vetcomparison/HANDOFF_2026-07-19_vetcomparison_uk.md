# HANDOFF — vetcomparison.uk

**Written 2026-07-19. Every figure below was re-checked against the live system today**
(live page fetched, DB queried, CMA case page fetched) — not carried forward from another doc.

**Read in this order:** this file → `SUMMARY_2026-07-18_readout.md` (what/why, plain prose) →
`PLAN_2026-07-15_rebuild.md` (phases + statuses) → `RUNBOOK_vetcomparison.md` (commands) →
`NOTES_vetcomparison.md` (running log, newest at the bottom, includes my own corrected mistakes).
Also `LEGAL_2026-07-15_*` (factual record — update it whenever publication-relevant facts change)
and `SUMMARY_2026-07-18_bugs_journey.md` (what went wrong, incl. the one that recurred).

**Obey repo `CLAUDE.md`.** Load-bearing here: concurrent sessions (pathspec commits per task,
forward-only, re-check `git status`), builds from committed HEAD, **verify against the pod not
the tag**, queue-check before dispatch, the standing-four docs, and above all —
***trust the rendered artefact, not the status***. That rule caught the worst defect in this
workstream; every `complete` status here has lied at least once.

---

## State, verified 2026-07-19

**Live site — clean and working.** `https://vetcomparison.uk` serves a real, searchable directory
of **2,109 verified practices**, three sourced CMA guides, area price statistics, and claim +
opt-out routes. Verified today: 0 fabrication markers, 0 unsupported claims, real
`/data/vet-full-index.json` wired, claim CTA present. The only monetary figures anywhere on the
site are the CMA's own (£21 / £12.50 / £500).

**Data.** 2,109 published; 238 practices `pending` (176 of them real practices parked because
their recorded website was a wheree.com mirror — they need genuine re-verification); **0**
fabricated price rows current (997 quarantined as `source='seed_import'`, never publish);
**0 claims and 0 claimed practices** — the claim flow is built and proven but nobody has used it
yet. Per-practice prices published: **none**, and that is correct — all historical price rows
have empty source URLs, so under our own rule they can only feed anonymous area aggregates.

**Exporter.** `directory-export-json` enabled, 48h cycle, last completed 2026-07-17 20:25Z, so
the **next run is due ~tonight 2026-07-19 20:25Z** — a good natural checkpoint to watch.

**Platform.** Site adopted onto the chassis 2026-07-17; full build cascade ran. Phases 0–4 of the
PLAN are done and deployed; Phase 5 (provenance-first scraping) is the remaining build.

---

## ⚠️ The one thing to watch first: the render path is UNVERIFIED

Background (full story in `SUMMARY_2026-07-18_bugs_journey.md` §9, bug case `/bugs_open/020`):
the chassis's tool-recreation agent rebuilt our practice search as a **client-side synthetic
practice generator** — invented names and postcodes — and shipped it live, four days after we
remediated this site for exactly that. All its work items said `complete`.

Both layers are now fixed:
- **Published file** — restored 2026-07-18, live-verified clean.
- **Source (`page_components.rendered_html`)** — the generator actually lived in the **`hero`**
  slot, not `filtered-result-grid`. Its data layer was rewritten to `fetch('/data/vet-full-
  index.json')`, keeping the chassis's better UI (region filter, pagination). Demo-sample
  disclaimer, price-sort controls, "pricing information / ownership data" claims, and a false
  about-page differentiator all removed. Four components carry `lock_type='permanent'`
  (index: hero, filtered-result-grid, info-card-grid; about: differentiators).

**But no render has run since.** Confirmed today: no commit has touched `vetcomparison.uk/` since
the restore. I could not dispatch one manually — `rerender-pages` is `experimental` and neither
`system.agent.site-builder.requests` nor `system.agent.page-rerender.process` produced an
orchestration state from a kcat trigger. **So the fix is verified in the database and on the live
page, but the render that joins them has never been exercised.**

When the first natural render happens the homepage will visibly **change** — the live page is
currently our simpler hand-authored version (6.1 KB, no region filter); the DB holds the richer
chassis component (11.3 KB, region filter + pagination, real fetch). That change is expected and
is an improvement. Verify it the moment it lands:

```
curl -s "https://vetcomparison.uk/?cb=$RANDOM" \
  | grep -ciE 'Mulberry32|makePostcode|PREFIXES|representative sample|ownership data|Price: Low to High'   # must be 0
curl -s "https://vetcomparison.uk/?cb=$RANDOM" | grep -c 'vet-full-index'                                  # must be >= 1
```
If fabrication returns, the locks did not hold — that is a platform finding; add it to bug 020.

---

## Owner review queue — 7 items (needs you in the admin UI)

**Your judgement:**
- `owned_page_review` — `tool-compliance-deadline-calculator`, unbuilt, 0 sections. A vet enters
  their number of sites and gets their CMA deadlines. Good idea on merits (it drives claims), but
  **if built it must not invent data** — same class as bug 020.
- Three planned pages with **0 sections**, so the builder correctly refused to invent content:
  `directory-index`, `guides-index`, `practice`. Decide purpose or cancel. `directory-index` is
  the interesting one — the planner evidently wanted the directory on its own page, which is
  worth settling deliberately (homepage, own page, or both). `practice` would eventually be the
  per-practice page where claimed prices appear.

**A fact only you have:**
- `needs_section_data` on contact — the identity spec has no email or phone. The site already
  publishes `vetcomparison@contactforsales.com` in its claim/opt-out links; confirm that (and
  whether to publish a phone number at all).

**Technical, a new session can clear:**
- `needs_page` about — **failed**, "claim timed out, handler pod likely died". Transient;
  about.html is live. Retry or cancel.
- `needs_page` about re-render — `validate_content` reported "1 blockers"; read the blocker.

---

## Open items, priority order

1. **Watch the first render** (above). Highest value per minute in the whole workstream.
2. **CMA funding consultation closes 30 July 23:59 — 11 days.** Draft ready at
   `CONSULTATION_RESPONSE_funding_DRAFT_2026-07-16.md`; owner verifies the levy figures against
   the Notice PDF and submits via connect.cma.gov.uk. Case team VetsMI@cma.gov.uk.
3. **The substantive draft Order is still not published** (case page checked today; latest entry
   remains 30 June 2026). It was timetabled for July and the Order must be made by 23 September,
   so it is overdue and imminent. When it lands: clause-referenced response from
   `CONSULTATION_2026-07-16_briefing.md` §3 — pro-independent, express reuse right,
   machine-readable lists, no selective blocking. **Nothing is submitted without owner sign-off.**
4. **Phase 5 — provenance-first scraping** (the remaining build, and what unlocks per-practice
   prices): re-verify the 176 wheree practices; scrape practices' own price pages persisting
   `source_url` + `observed_at` **per price** (without them the exporter refuses to publish — by
   design); extend the discovery deny-list (wheree.com, bestlocalrated, yelp.*, starofservice,
   threebestrated, allvets.co.uk, calmshops, rated.club, digifarm.uk); **normalise website host
   before upsert** — that omission caused 280 duplicate practices.
5. **Fix bug 020 properly** (platform, not this site): the tool-recreation path has no
   data-dependency contract, and its prompt's "no fake data" rule is scoped to arithmetic. Fix
   candidates ranked in the case file. Fleet-wide risk, not just ours.
6. Solicitor review (LEGAL §8): the factual record and the database-right position. Owner
   decision 2026-07-16 was that attributed publication proceeds meanwhile under stated conditions.
7. Optional: classifier `content_features` patch if a news feed is wanted (the May 005 patch was
   lost and never landed anywhere).

---

## Hard rules (do not relax)

- **No price without provenance or consent.** `seed_import` rows never publish. We do **not** own
  `vetcomparison.co.uk` — no default domains anywhere (fail-closed in Go, blanked in DB config).
- **Never re-plan this site to fill gaps** — the first-plan branch is the only faithful path and
  re-plans clobber built pages (`/bugs_open/001`).
- **Never let a generative step invent records.** If a component needs data it cannot reach, it
  renders an empty state and fails to review. This is the whole reason the site exists.
- Sites-repo deploys: detached worktree off `origin/master`; **verify and push as separate
  steps**; rebase-retry when render bots race you; never force-push; never touch the stale dirty
  local clone (~1,700 commits behind).
- Aggregates: min_n 3, always publish the n. Claimed supersedes scraped. Action opt-outs promptly.

## Pointers

- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
  (password from secret `postgres-clients-secret`). Migrations 006–010 in this dir, all applied,
  all idempotent.
- Large HTML into a DB column: generate **dollar-quoted SQL locally and pipe via stdin**.
  `\set x \`cat file\`` runs `cat` *inside the pod* and silently blanks the column (it reported
  `UPDATE 1` while writing 0 bytes — caught only by re-querying).
- Live checks with `curl`; python-urllib gets a Cloudflare 403. Mailto bodies are URL-encoded, so
  grep the encoded form.
- Exporter config lives in `scheduled_tasks.input_data` for `directory-export-json`.
- Adoption re-trigger if ever needed: `bash 082_submit_domain_unified.sh vetcomparison.uk --from
  https://vetcomparison.uk --fidelity locked`.

## Other files in this directory (not written by this workstream)

- **`README_where_we_are.md` — STALE, do not act on it.** It is a paste of chat messages from
  15 July: it still says the strip needs pushing and quotes 2,579 practices. Superseded entirely
  by this handoff. `README_vet_legal.md` is a similar paste of the copyright discussion — its
  reasoning still stands, but `LEGAL_2026-07-15_*` is the authoritative version.
- **`RCVS_mismanagement.md`, `RCVS2_mismanagement.md`, `RCVS2_who_should_manage_it.md`,
  `RCVS3_report_and_discussion`** — owner-supplied research (~150 KB) on RCVS institutional
  efficacy, regulatory overreach, financial stewardship and its software-project record.
  **Strategically live, and not yet folded into any plan:** the RCVS is building the official
  "Find a Vet" comparison tool that will be our main competitor (~Sept 2027), and it sets the
  approval criteria for third parties (~June 2027). If its delivery record is weak, that bears
  directly on the consultation argument for low-barrier third-party approval, and on our own
  positioning. Worth a session to read them and decide what, if anything, we say publicly.
