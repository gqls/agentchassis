# HANDOFF — portfolio_positioning — 2026-08-26. **START HERE. First task is §1.**

Supersedes `HANDOFF_2026-08-25b_continue_here.md`. Owner read-out (still current — no new summary
today, deliberately: the five headings would read the same):
`SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md`.

**Counts carry the date they were counted** (owner ruling 2026-08-22).

---

# 0. STATE IN ONE PARAGRAPH

Sitemaps are **done, live, self-maintaining — and since today they FOLLOW THE DEPLOY.** Migration
`642` (applied 2026-08-26 ~10:30Z, council APPROVED round 1) makes a site whose pages changed since
its last render due EARLY after 30 quiet minutes; the 3-day line is now an unconditional floor. A
published or retracted page reaches its sitemap in ~30–60 min instead of up to 3 days. **Proven the
same day, 8 for 8**: nothing was age-due before 08-27 16:02, and 8 sites were selected between
10:42Z and 14:19Z, all COMPLETED, 0 dropped, all committed, the first verified at the served body.
The hand-run generator's non-canonical homepage is fixed too (`bcb4645ff`), so nothing in the estate
writes `/index.html` for a root any more. **Still unproven: `622`'s guard (§2a).** Nothing is blocked.

---

# 1. FIRST TASK — two cheap checks, then §3

## 1a. Is `642` still draining cleanly? (2 minutes)

```sql
-- every early selection: a stamp newer than the apply while the site was NOT age-due
SELECT s.domain, r.last_selected_at, o.status, o.collected_data->'sitemap_render_result'->>'url_count' AS urls,
       o.collected_data->'sitemap_render_result'->>'probe_dropped' AS dropped
FROM site_discovery_rotation r JOIN sites s ON s.id=r.site_id
LEFT JOIN orchestration_states o ON o.owner_agent_type='sitemap-refresh'
     AND o.created_at BETWEEN r.last_selected_at AND r.last_selected_at + interval '10 min'
WHERE r.agent_type='sitemap-refresh' AND r.last_selected_at > '2026-08-26 10:00+00'
ORDER BY r.last_selected_at;
```
Expect one row per 30-min tick, all COMPLETED, `dropped` 0 or explained. ⚠ `orchestration_states` is
reaped after ~24h (§4a) — for stamps older than that the join is NULL and means UNKNOWABLE, not failed.

```sql
-- the live due set: age / change-and-quiet / changed-but-busy
SELECT count(*) FILTER (WHERE age_due), count(*) FILTER (WHERE NOT age_due AND change_due AND quiet),
       count(*) FILTER (WHERE NOT age_due AND change_due AND NOT quiet)
FROM (SELECT COALESCE(r.last_selected_at,'-infinity'::timestamptz) < now() - interval '3 days' AS age_due,
        EXISTS (SELECT 1 FROM pages pu WHERE pu.site_id=s.id AND pu.updated_at > r.last_selected_at) AS change_due,
        NOT EXISTS (SELECT 1 FROM pages pq WHERE pq.site_id=s.id AND pq.updated_at > now()-interval '30 minutes') AS quiet
      FROM sites s LEFT JOIN site_discovery_rotation r ON r.site_id=s.id AND r.agent_type='sitemap-refresh'
      WHERE s.status IN ('active','deployed') AND s.locked_at IS NULL
        AND EXISTS (SELECT 1 FROM pages pg WHERE pg.site_id=s.id AND pg.status='active' AND pg.deployed_at IS NOT NULL)) x;
```
Readings so far: design time **0/28/2** · apply (~10:30Z) **0/20/—** · 14:20Z **0/14/11**. A large
change-and-quiet number that does NOT fall by ~2/hour means the rotation stopped ticking — check
`scheduled_tasks.last_triggered_at` for `sitemap-refresh-rotation` before anything else.

## 1b. `622`'s guard — still the falsifiable query from 08-25b

```sql
SELECT s.domain,
       (SELECT count(*) FROM pages p WHERE p.site_id=s.id AND p.status='active' AND p.deployed_at IS NOT NULL) AS deployed_pages,
       (SELECT r.last_selected_at FROM site_discovery_rotation r WHERE r.site_id=s.id AND r.agent_type='sitemap-refresh') AS stamp
FROM sites s WHERE s.status IN ('active','deployed') AND s.locked_at IS NULL ORDER BY 2 ASC, s.domain;
```
**`deployed_pages = 0` AND a non-null `stamp` means the guard did not hold.** As of 2026-08-26 09:30
BST every one of 31 sites has ≥1 (min `apis.uk`, `lampenkap.com` at 1), so the guard has never been
consulted. It gets its first real test on the next newly-seeded site.

# 2. WHAT IS LIVE, AND WHAT IS NOT YET PROVEN

| | evidence |
|---|---|
| **29 of 31 serve a sitemap of ours** | census by BODY 2026-08-25 21:02, every one n/n; the other 2 correct-by-design (`adversecreditmortgage.co.uk` HALT, `webdesign.uk` redirect-only) |
| **`642` — the sitemap follows the deploy** | applied 2026-08-26; council `6e448adb` APPROVED 10:35Z round 1, 2 advisories both checked (§5); **8 early selections 10:42Z–14:19Z**, all COMPLETED/0 dropped/committed; `loancalculator.co.uk` served body: 28 locs = 28 rows, two `2026-08-26` lastmods on exactly the pages that made it due |
| **Canonicalisation, both writers** | action `5c9acf1bd` (08-24); script `bcb4645ff` (08-26, exercised on `cv1.co.uk` 3/3) |
| **Redirect fix `54ba65b25`** | proven 08-25 19:59, `webdesign.uk` 7→0 |
| **Selector guard `622`** | applied 08-25, APPROVED — **behaviourally unproven (§1b)** |

## 2a. Not proven — say it plainly

- **`622`'s guard has never fired.** §1b.
- **`642`'s quiet gate has never been SEEN deferring a site.** The proof shows the early branch
  selects and renders; deferral is visible only in the due counts falling (28 → 20 → 14/11 at three
  instants), not in a selection that was withheld. Not a gap — but do not describe the quiet period
  as "proven" without that qualification.

# 3. WHAT IS OPEN

- **The `skip_reason` residue** (unchanged): a site whose pages are ALL noindex or ALL expired still
  burns its slot silently. Needs machine-readable `skip_reason` (`opted_out` | `no_listable_urls`) on
  `render_sitemap` — Go — and branching on THAT, never the prose `reason`. **Population 0** as of
  2026-08-25. With `642` the cost of a burned slot is now "waits for the floor" rather than "waits
  3 days regardless", so this is lower priority than it was.
- **"Is 3 days right?" — reframed, not answered.** The floor now only catches serving-side drift
  with no DB change (a URL that starts 404ing without any row moving). ~9 renders/day. Lengthen
  only if that cost ever matters; there is no latency argument left for shortening it.
- **Cloudflare's managed `robots.txt` is merged into `cv1.co.uk`'s** and disallows ClaudeBot, GPTBot,
  Google-Extended, CCBot, Amazonbot, Applebot-Extended, Bytespider, meta-externalagent,
  CloudflareBrowserRenderingCrawler [seen 2026-08-26 via the script's rule-3 detection]. Whether the
  estate wants AI crawlers blocked is the **owner's** call. Not measured fleet-wide; `traffic_probe`
  lane owns the Cloudflare question historically — check there before acting.
- **The 22 hosted-site remakes** (`DECISION_2026-08-20_remake_the_hosted_sites.md`). 3 protected:
  `leopardess.co.uk`, `leopardess.uk`, `cartoon.co.uk`. Do not start with
  `businessinsurancequotation.co.uk` (insurance → compliance layer, and the largest).
- **The Christmas card sender** (register G3/G4) — delivery half FIRST; read `bugs_open/283`.
- **`adversecreditmortgage.co.uk` stays halted** — owner's call.
- **21 portfolio domains have no register row** (as of 2026-08-21).

# 4. TRAPS

## 4a. The reconciliation window (carried from 08-25b, and `642` changes its usefulness)

`orchestration_states` keeps COMPLETED/FAILED ~24h. A stamp older than that with no matching run is
**UNKNOWABLE, not a dropped dispatch** — the old `runs = 0` query gave six false hits on 08-25 and
its remedy ("clear the stamp") re-probes healthy sites. With `642`, far more stamps are now recent,
so the in-window check in §1a is useful more often — and the artefact census (§4b) is still the only
detector that does not go stale.

## 4b. The artefact census — unchanged

Fetch `/sitemap.xml`, **`rm` the temp file first**, judge the BODY (a parking file, a `text/html`
homepage and a redirect all "succeed" by status), extract every `<loc>`, canonicalise BOTH sides
(`/index.html` → `/`) and match against `pages`. Ours score n/n; the parking file 0/1. One miss on
every domain is one systematic difference, not N faults.

## 4c. New today

- ⚠ **A council submission's SUMMARY must list EVERY pre-flight assertion.** `642`'s row-count
  guard was in the file and not in the summary; `editquality` objected (medium) to its absence.
  Reviewers cannot open the file — an unlisted guard is an absent guard to them.
- ⚠ **Do not run `landmines-verify-dispatch.sh` while another lane has a NEW uncommitted entry in
  `LANDMINES.md`** — the sync consumes their new-entry status and the verifier never checks theirs.
  Check `git diff --numstat` on the file first. (My own status correction there was swept into
  `b3bddba60` by the imagery lane; the `doc_notes` copy lags until the next lane syncs. Known.)
- ⚠ **`ls` the migrations dir immediately before naming a file** — third time on this lane: the tree
  moved 634 → 641 during one session and `635` was taken.
- ⚠ **`pages.updated_at` is bumped by EVERY writer by convention, not by a trigger.** `642` relies on
  that; a new writer that forgets `updated_at = NOW()` is invisible to the early branch (the floor
  still catches it in 3 days). If you add a `pages` writer, bump it.
- The rest of 08-25b §4c still holds: `deployed_at` is updated on every redeploy (never use it for
  "first published"); a probe proves fetchability, never canonicality; do NOT add `repo_name` to
  `590`'s `commit_sitemap`; do NOT tighten `622`'s guard; scope `MIGRATIONS_DIR` on every apply
  (today's pending set was 614–638, all other lanes'); `grep -aq` over `/proc/1/exe` false-absences
  on BusyBox.

# 5. COMMITS AND COUNCIL, 2026-08-26

`107327c6b` **migration `642`** + ROLLBACK + submission JSON (`Council-Submitted: 6e448adb…`) ·
`bcb4645ff` script canonicalisation · docs commit (this file, NOTES `2026-08-26`, README, `seo.md`).
My LANDMINES correction rode in `b3bddba60` (imagery lane).

**Council `6e448adb-1e03-4e2e-a3dd-42bc6857ff24` — APPROVED 10:35:39Z, round 1, 10 min after
submission.** `editquality` medium: assert one row by name — **already in the file's first `DO`
block**; live: 1 row, no sibling, did not fire. `debug_historian` low: concurrent ticks could
double-select before the stamp (583/584 family) — pre-existing, `kafka-scheduler` is 1 replica, 8/8
distinct today. 6 abstained, 9 approve; `architecture`: *"point_fix … exactly the shape 590
deferred toward"*.

# 6. FILES OF RECORD

**Cold start:** this file → `SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md` →
`README_where_we_are.md` (owner's log, entry 2026-08-26) → `NOTES_portfolio_positioning.md`
(entry `2026-08-26` (a)–(g); `2026-08-25` (c)–(g) for the previous day's proofs).

**Sitemaps:** `platform/orchestration/actions/render_sitemap_action.go` · migrations
`docs/agent_docs/sql_for_agents/590_…`, `622_…`, **`642_sitemap_rotation_follows_the_deploy.sql`**
(each + `_ROLLBACK`) · `scripts/site-discovery-files.py` · register **SEO-007** / **SEO-002** in
`docs/agent_docs/docs026_concept_register/register/seo.md` ·
`COUNCIL_SUBMISSION_642_sitemap_follows_the_deploy.json`.

**Decisions:** `DECISION_2026-08-20_remake_the_hosted_sites.md` · `REGISTER_positioning.md` ·
`RFC_037` (open).

---

# ADDENDUM 2026-08-26 ~19:00Z — §1 ran green; the remake programme has STARTED

**§1a**: 16 early selections 10:42Z–18:26Z, all COMPLETED, ~30-min cadence; due set 0/15/10 at
18:42Z (pool refills as it drains — 10 sites busy; rotation provably ticking, latest 18:26Z).
**idea.uk's `dropped=1` is PERMANENT and by design** — the idea_uk_vm_site lane 301s the static
`/privacy.html` to the tool's `/privacy` (their RUNBOOK, decided 2026-07-18). Expect candidate 23
→ kept 22 on every idea.uk render; do not re-chase. **§1b**: guard still unviolated, still never
consulted. The "31 vs 30 rows" is the locked halted site; not a departure. Evidence: NOTES (h).

**§3 — the remakes are no longer "not started":**

- Precondition re-confirmed gone (`bugs_open/311` closed 08-24; RFC_036 tool half live).
- **`advertise.co.uk` brief WRITTEN and HELD** (NOTES (i)–(k)): locked `test` sites row
  `d991a5b8-428f-44c1-b3eb-e50f44326fd9` (buytoletcalculator precedent), fired 18:48Z with a
  direction naming the remake ruling + the three estate neighbours, COMPLETED 18:51Z. Brief
  verified at the artefact: 15,915 B, 13 keys, confidence 0.78, differentiation stays off
  websitepromotion/seotools/webdesign BY NAME. `needs_brief_review` held at `needs_human_review`.
- **Deliberately fired ONE brief, not five.** Its Q2 asks whether advertise.co.uk is the HUB of
  the marketing cluster — the owner's answer shapes the websitepromotion.co.uk / seotools.co.uk /
  designblog.co.uk briefs. Firing those before the answer bakes in a guess. **Next session: check
  whether the owner has reviewed/released (`SELECT status FROM site_work_items … item_type=
  'needs_brief_review'` for advertise.co.uk); if released, the build flow takes over; if answered
  in prose, fold the answers into the next briefs' directions.**
- Before-snapshot saved: `salvage/advertise.co.uk/index.html` (20,453 B). The "single-pager"
  classification was an undercount — it is a Drupal 7 RSS aggregator with `?q=node/N` pages, but
  every node is a syndicated headline stub, so nothing original is at risk (NOTES (j)).
- ⚠ The advertise.co.uk row is status **'test'**, so it is invisible to the sitemap rotation and
  to §1b's census (both filter active/deployed) — `622`'s first real test is still pending on the
  next genuinely seeded site, and this row is NOT it.
- The owner's review queue is now THREE: indoorplanters (test 08-20), buytoletcalculator (test
  08-21), **advertise.co.uk (real)**. Flagged in README_where_we_are (entry "2026-08-26, evening").

**Late addition (same evening): `bugs_open/414` filed — lendzy's 08-02 acceptance marker is SERVED**
("checked against the FCA handbook, rule by rule", /about.html ×2 + guide ×1), and a held audit
item canonised it as the site's differentiator. Spec source FIXED live (row `81ddcc40`, guarded
strip, history kept); copy repair OPEN — ⚠ rerender cannot fix it (phrase is in `content_data`);
the held "differentiator" `content_rewrite` must be rejected/rewritten. NOTES (l); 016b §9+§10.

> **CORRECTED 2026-08-27:** the late-addition line "Spec source FIXED live" was REFUTED by the 414
> fixing session — `strategy` (aspect read by build-site-planner/webdesign-agent) carried
> `domain-strategist`'s 08-12 PARAPHRASE of the marker until they stripped it 08-27 (row `0326a892`).
> Fleet payload-census 0 across all aspects, re-verified by this lane. Audit item REJECTED (theirs);
> framework fix `fc588e445` + council `f4c144ad`; copy repair IN FLIGHT via framework (phrase still
> served until it lands — verify at the body). My census error: WRONG_CALLS 2026-08-27. Read 414 §7
> + its REFUTED block before citing anything from the late addition.
