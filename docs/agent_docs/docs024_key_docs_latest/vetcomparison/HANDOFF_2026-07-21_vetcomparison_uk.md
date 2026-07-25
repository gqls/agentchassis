# HANDOFF — vetcomparison.uk

**Written 2026-07-21. Supersedes `HANDOFF_2026-07-19_vetcomparison_uk.md`** (still worth reading
for the deeper background on the fabrication remediation and the legal/consultation record).
Every figure below was re-checked against the live system today — pod, DB, live pages — not
carried forward.

**Read in this order:** this file → `HANDOFF_2026-07-19_...` (background) →
`README_where_we_are.md` (owner's plain-prose history, append to it) →
`NOTES_vetcomparison.md` (technical log, newest at the bottom, includes my missteps) →
`PLAN_2026-07-15_rebuild.md` → `RUNBOOK_vetcomparison.md`. Fleet missteps this session are in
`docs024_key_docs_latest/WRONG_CALLS.md`.

**Obey repo `CLAUDE.md`, and RE-READ IT from disk first** — it is co-edited and moves daily. Two
clauses bit me this session and are now load-bearing here: **council dispatch queues ~30 min, a
missing orchestration row is latency not a drop** (do not resubmit — find the run by payload); and
**trust the rendered artefact, not the status** (every `complete` here has lied at least once).

---

## State, verified 2026-07-21

**Live site — clean, and the first natural render has now landed.** The homepage flipped from the
6.1 KB hand-authored page to the **46.7 KB chassis render** (verified: `curl` → 200, 46,656 bytes).
It carries the real search (`hero`: region filter, pagination, `fetch('/data/vet-full-index.json')`),
the info-card grid, **server-rendered news** (3 `<article class="news-card">` in the raw HTML, no JS),
and the CTA. **Zero fabrication** — `Mulberry32`/`makePostcode`/`representative sample`/`ownership
data` all absent from the live HTML. 2,109 practices still serving.

**The news feed now WORKS end to end.** `/data/latest-news.json` → 200, `item_count: 3`, all real
CMA items with genuine gov.uk URLs (CMA Impact Assessment; Vets market investigation draft funding
Order; Veterinary services for household pets). This was unblocked by the `042` fix landing in
v1.0.1144 (see below).

**Build v1.0.1144, verified against the pod** (not the tag):
- `took literal scalar config value` → 1 — **my `042` fix is live** (numeric step config now reaches
  actions; the `max_age_hours: 720` window is finally real, which is what let the 460h-old CMA items
  through).
- `persistNewsSectionHTML` → 0 — **my original `027` server-render is GONE.** Another session
  reworked `027` through the council (REVISE → "queryresolve resolvers over content_feed_items"),
  and the v1.0.1142 sweep (`2d529d6dc`) removed the function. News is nonetheless server-rendered
  now, so `027`'s goal is met by their approach. **`027` is theirs — do not touch it;** run
  `scripts/who-owns.py 027` before any move.

**Platform.** `042` CLOSED (`bugs_closed/042`, fixed AND live). The two about-page queue items were
cancelled 07-20 (stale premise). Owner review queue is **5**: three 0-section pages
(`directory-index`, `guides-index`, `practice`), the contact email fact, and the
`tool-compliance-deadline-calculator` review — all genuinely need the owner.

---

## ⚠️ Two things to watch first

### 1. The directory exporter is still failing — and my fix does NOT cover it
`directory-export-json` last ran **2026-07-19 20:25 and FAILED**: *"directory export requires an
explicit domain; refusing to export without one"* — although `scheduled_tasks.input_data` plainly
carries the domain. It runs on a 48 h cycle, so **the next run is ~tonight 2026-07-21 20:25** — a
natural checkpoint. Consequence: `/data/vet-full-index.json` has stopped refreshing (still serves
the last good copy, 2,109 practices, so nothing looks wrong on the page).

**Why my `042` fix probably does NOT save it.** `042` fixed *numeric* config not reaching actions;
the exporter's problem is a *literal string* (the domain) not reaching the action — a sibling in
the same family that I deliberately left **out of scope** (taking unresolved strings as literals
would mask real broken references). Filed as diagnosis `55dc0fa4-116c-40d6-90b2-bfad9ad73692`, but
**that run hung at `route` and produced no verdict** (see `bugs_open/043`). So this one likely
needs a human read of `directory_export_action.go:123-136` and `vet_med_export_action.go:152`
(identical shape), not a wait for the loop. **Watch tonight's 20:25 run; if it fails again, the
literal-string plumbing is the cause.**

### 2. A render restored the dead search grid and stripped the anti-fabrication locks
> **RESOLVED 2026-07-24 (durably, at the source).** The grid returned a third time (via the
> 07-23 20:36 render). Root cause: it was in the **current site plan** (`site_plan_sections`,
> plan `9d9c601d`, `ordering=1`), so every render faithfully reproduced it — a plan-content
> issue, **not** the `bugs_open/001` re-plan-clobber this note guessed (nothing re-planned).
> Fixed by removing it from the plan itself (+ `pages.sections` + `page_components`, all
> snapshotted); reconciles can no longer regenerate it. Full account:
> `NOTES_vetcomparison.md` 2026-07-24 entry.
>
> **CONFIRMED CLOSED 2026-07-25.** Two full renders have since run (07-25 01:49:55 and 13:51:20)
> and both produced **four** sections, not five — the fix survives a rebuild, which is what the
> two earlier hand-deletes did not. Live: 42,051 bytes, `filtered-result-grid` = 0.
> *(One correction to the line above, which I wrote: it did NOT flush on a ~6 h
> `content-feed-refresh` cycle. That cycle no-ops when there is no news change; it flushed on the
> next render, ~1 day later.)*
>
> ⚠️ **But the same page has a worse, unrelated defect found 07-25: all six links in
> `info-card-grid` are 404**, plus 3 chrome links to never-built pages = 9 live 404s on the
> homepage. One fixed (`/guides/cma-compliance` → `.../index.html`); the other five need an owner
> decision on what should exist. Contributed to `bugs_open/023` (OWNED — do not fork a fix).
> See `NOTES_vetcomparison.md` 2026-07-25 entry §3.
The 08:08 render re-materialised all five index components — which (a) brought back the dead
`filtered-result-grid` I had deleted on 07-19, and (b) stripped every `bug-020` `permanent` lock
(verified delete-and-recreate: the `hero` row's id changed). **I re-removed the grid today, this
time from `pages.sections` as well as `page_components`** (the manifest is the regeneration source —
my 07-19 note that "sections is a no-op" was true only of the *assembly* path, not the *regenerate*
path). **The DB is clean; the live page still shows the grid until the next render.** If it returns
*again*, the source is the site plan, and that is a `bugs_open/001` re-plan-clobber problem — do not
just re-delete in a loop. The lock-stripping is filed to `bugs_open/020`: content regenerated clean
this time, so the exposure is latent, but **no lock-based mitigation on this platform survives a
rebuild.**

---

## Open items, priority order

1. **Watch tonight's exporter run (~20:25)** and, if it fails again, fix the literal-string input
   plumbing (`directory_export_action.go`). This is the one thing actively degrading (stale
   directory data). Highest value.
2. **CMA funding consultation closes 30 July 23:59 — ~9 days.** *(Carried from 07-19; re-verify the
   date against the Notice before relying on it.)* Draft ready at
   `CONSULTATION_RESPONSE_funding_DRAFT_2026-07-16.md`; owner verifies levy figures and submits via
   connect.cma.gov.uk (case team VetsMI@cma.gov.uk). Now that the site carries the CMA feed, the
   consultation item will surface on the homepage automatically.
3. **The substantive draft Order** *(status as of 07-19; re-check the case page)* was timetabled for
   July, must be made by 23 September, and was still unpublished. When it lands: clause-referenced
   response from `CONSULTATION_2026-07-16_briefing.md` §3. Nothing submitted without owner sign-off.
4. **Owner review queue (5 items)** — the three 0-section pages need a purpose-or-cancel decision
   (`directory-index` is the interesting one: does the directory want its own page?); the contact
   item needs the email/phone fact; the deadline-calculator needs a build-or-cancel call (**if
   built it must not invent data** — same class as `020`).
5. **Phase 5 — provenance-first scraping** (unchanged from 07-19): re-verify the 176 wheree
   practices; scrape practices' own price pages with `source_url` + `observed_at` per price;
   extend the discovery deny-list; normalise website host before upsert.
6. **`bugs_open/043` — the diagnosis loop returns no verdicts** (runs hang at `route`). Not this
   site's bug, but it blocks the auditable-diagnosis workflow CLAUDE.md now makes the default. Five
   runs this session, zero verdicts. Worth a platform thread.

---

## Hard rules (do not relax)

- **No price without provenance or consent.** `seed_import` rows never publish. No default domains.
- **News is RSS-only, by design.** `content_features.news_feed.source_types = ["rss"]` on this site
  is load-bearing: `seed_content_sources_action` runs every cycle and would auto-create an
  xAI/Grok `api_news` source (LLM-*authored* news) for any other type. The reason is written into
  the spec's `reason` field — read it before changing it. This site was remediated for fabricated
  content; LLM-written news is exactly the wrong thing here.
- **Never re-plan this site to fill gaps** (`bugs_open/001`); **never let a generative step invent
  records** (`bugs_open/020`); locks do not protect you (see above) — provenance in the data source
  does.
- **Council submissions:** the fix-plan allowlist is `modify | remove | config_change` — **NOT
  `create`.** A submission that adds a file is rejected at intake (cost me two wasted rounds this
  session, WRONG_CALLS 07-20). Budget ~30 min per run; find it by payload, never resubmit on a
  missing row.
- Sites-repo deploys: detached worktree off `origin/master`; verify and push separately;
  rebase-retry when render bots race; never force-push; never touch the stale local clone.

## Pointers

- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
- Verify build against the POD: `kubectl exec -n ai-persona-system <agent-chassis pod> -- sh -c
  'strings /app/agent-chassis | grep -c "<symbol>"'`
- News window lives in `content-feed-orchestrator`'s `render_news_json` step config **and** in seed
  `sql_for_agents/090_content_feed_orchestrator.sql` — change both or a re-seed reverts it.
- Snapshots taken this session: `_vetcomparison_bak_20260719_index_components`,
  `_vetcomparison_bak_20260721_index_components` (the 08:08 render's components, pre-grid-removal).
- Find a council/diagnosis run by payload, not the printed id:
  `... WHERE collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';`

## What changed this session (2026-07-19 → 21), in one place

- **Built the CMA news feed** (RSS-only, the keyword-filtered GOV.UK feed); it now publishes.
- **Found & fixed `042`** — numeric step config never reached actions (CLOSED, live in v1.0.1144).
- **Removed the dead search grid** (twice; durably now).
- **Filed** `bugs_open/043` (diagnosis loop route-hang), `bugs_open/020` addendum (locks don't
  survive rebuilds), and three `WRONG_CALLS` entries (a wrongly-declared fleet outage, a
  wrongly-declared council drop + resubmit, a wrongly-reported "filed").
- **Corrected** my own `027` addendum (superseded by another session) and `043` (the council was
  never stalled — it was ~1 h latency I misread).
