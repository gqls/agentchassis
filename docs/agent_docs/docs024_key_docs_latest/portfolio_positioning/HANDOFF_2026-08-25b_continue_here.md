# HANDOFF — portfolio_positioning — 2026-08-25b. **START HERE. First task is §1.**

Supersedes `HANDOFF_2026-08-25_continue_here.md`. Owner read-out:
`SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md` (still current).

**Counts carry the date they were counted** (owner ruling 2026-08-22).

---

# 0. STATE IN ONE PARAGRAPH

Sitemaps are **live and self-maintaining**: the rotation swept the fleet unattended and now picks up
**new sites on its own** (three arrived yesterday and all three were swept without anyone asking).
**27 of 31 live sites serve a sitemap of ours**, every one a perfect match against its own pages.
Three defects were found and fixed across this work; the third (`622`) came from a council objection
that had been consciously deferred and turned out to be **live on two sites**. Nothing is blocked.

---

# 1. ⏭ START HERE — verify the two fixes that are now live but UNPROVEN at the artefact

Both landed within the last hour and neither has been confirmed on a real site yet. **Do these
before starting anything new** — they are quick and they close the loop.

## 1a. ✅ The redirect fix (`54ba65b25`) — **PROVEN 2026-08-25 19:59. Nothing to do.**

`webdesign.uk` re-ran on chassis v1.0.1339 and behaved exactly as predicted:

| | before (2026-08-24 20:06) | after (2026-08-25 19:59) |
|---|---|---|
| `candidate_count` | 7 | 7 |
| `url_count` | **7** | **0** |
| `probe_dropped` | **0** | **7** |
| `rendered` | true | **false** |
| committed | **yes** (7 redirecting URLs, to `vm-sites`) | **no** |

Same 7 candidates, and now all 7 are dropped by the probe because it finally sees the 302s.
`reason: "no listable URLs — refusing to publish an empty sitemap"`, and `check_has_urls` routed to
`complete` without committing. **A domain that redirects everything away correctly ends up with no
sitemap of its own** — so `webdesign.uk` will stay at 3 of 4 uncovered permanently, by design.

## 1b. ⏭ ONE SITE LEFT — `cv1.co.uk`, queued ~20:59

**`homegarden.uk` is DONE and verified 2026-08-25 20:30.** It was one of the two sites the `622`
defect had silently skipped:

| | before | after |
|---|---|---|
| `/sitemap.xml` | **HTTP 404** | **HTTP 200**, `application/xml`, 2,237 B |
| `url_count` | 0 (`candidate_count` 0) | **20** (`candidate_count` 20, `probe_dropped` 0) |
| committed | no | **yes** |

**20 of 20 locs match its `pages` rows**, and the homepage is listed as `/` — canonicalised, so the
`5c9acf1bd` fix is applying to newly-swept sites too.

**`cv1.co.uk` is the last one.** Stamp cleared, queued behind the other two, expected ~20:59:

```bash
rm -f /tmp/s.xml   # a failed curl leaves the PREVIOUS body in place
curl -s -o /tmp/s.xml -w '%{http_code}\n' https://cv1.co.uk/sitemap.xml; grep -c '<loc>' /tmp/s.xml
```
Expect **HTTP 200** and **3 locs** (it has 3 deployed pages). It served 404 as of 20:30.

### ⚠ BE PRECISE ABOUT WHAT THIS PROVES — and what it does NOT

`homegarden.uk` recovering proves **the manual remedy works** (clear the stamp → the site returns to
the front of the queue → it renders and commits). **It does NOT prove `622`'s guard fires**, because
`homegarden.uk` now HAS deployed pages, so the guard is not even consulted for it.

`622`'s guard is currently proven only by (a) its three induced verify failures against the live DB
and (b) the live `pre_query` text. **Its behavioural proof requires a site that has NO deployed
active pages at the moment the rotation would pick it** — i.e. the next newly-seeded site. When one
appears, the observable claim is that it is **NOT stamped** while it has no deployed pages:

```sql
-- a site with zero deployed pages must have NO row here until it is built
SELECT s.domain,
       (SELECT count(*) FROM pages p WHERE p.site_id=s.id AND p.status='active' AND p.deployed_at IS NOT NULL) AS deployed_pages,
       (SELECT r.last_selected_at FROM site_discovery_rotation r
         WHERE r.site_id=s.id AND r.agent_type='sitemap-refresh') AS stamp
FROM sites s WHERE s.status IN ('active','deployed') AND s.locked_at IS NULL
ORDER BY 2 ASC, s.domain;
```
**A site with `deployed_pages = 0` AND a non-null `stamp` means the guard did not hold.**

## 1c. Read the council verdict on `622`

`SUBMISSION_CORR=c88f5c0f-cca2-4753-bd6c-9fabc93b100e` (submitted 19:42).

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='c88f5c0f-cca2-4753-bd6c-9fabc93b100e' AND kind='council_report' ORDER BY created_at;
```
`689f0c5fa` carries `Council-Submitted:`, so `098` credits it automatically once approved. **Do not
amend.** Approval is not a reason to skip the objections — 2 of 7 were real on the wiring round, and
the one deferred there is what became `622`.

# 2. WHAT IS LIVE

| | evidence |
|---|---|
| **27 of 31 live sites serve a sitemap of ours** | census by BODY 2026-08-25 19:35; every one a **perfect n/n** `<loc>`↔`pages` match |
| **The fleet swept itself unattended** | 13.2 h, 08-24 15:32 → 08-25 04:44, 27 sites, **all COMPLETED** |
| **New sites are picked up automatically** | `homegarden.uk`, `lampenkap.com`, `cv1.co.uk` all arrived and were swept with nobody asking |
| **Canonicalisation fix (`5c9acf1bd`)** | **ZERO of 31 domains emit a non-canonical homepage** (re-checked live 2026-08-25) |
| **Redirect fix (`54ba65b25`)** | council **APPROVED, zero objections**; **PROVEN at the artefact 2026-08-25 19:59** — `webdesign.uk` went `url_count` 7→**0**, `probe_dropped` 0→**7**, no commit |
| **Selector guard (`622`)** | applied 19:45, three guards induced against the live DB — **behaviourally unproven until a site with zero deployed pages next appears; see §1b for the exact claim** |

**The four sites NOT covered, all understood:**

- `adversecreditmortgage.co.uk` — **correct by design**, excluded by `locked_at` (owner HALT). What
  it serves is the parking provider's 1-`<loc>` `/lander` file.
- `webdesign.uk` — 302s everything away. **Resolved: the correct end state IS no sitemap, and it now correctly produces none** (§1a). It stays uncovered permanently, by design.
- `homegarden.uk` — **RECOVERED 20:30**: 404 → 200 with 20 locs, 20/20 matching. No longer uncovered.
- `cv1.co.uk` — the last one, queued ~20:59. §1b.

# 3. WHAT IS OPEN

- **The deploy-path half of SEO-002's question — the main remaining item.** A newly published page
  waits up to 3 days. ⚠ The cost that made it second still applies: one GET per URL means **136
  requests for `webdesign.co.uk` on every edit**. Debounce, or probe only the changed URL and merge.
  ⚠ **`622` reduces the urgency but does not remove it**: a site now waits for its slot rather than
  being skipped, but it still waits.
- ⚠ **The residue `622` deliberately does NOT fix.** A site whose pages are **all** noindex or **all**
  expired still burns its slot silently. Needs the machine-readable `skip_reason`
  (`opted_out` | `no_listable_urls`) on `render_sitemap` — a Go change — and branching on **that**,
  never on the prose `reason` string. **Population today: zero.**
- **`scripts/site-discovery-files.py:132` still emits the non-canonical `/index.html`.** The action
  no longer does. Driven by nothing (0 `scheduled_tasks`, no CronJob, checked 08-24), so no
  two-writer race — only wrong when hand-run. Small fix.
- **Whether 3 days is right**, now real cost is observable.
- **The 22 hosted-site remakes** (`DECISION_2026-08-20_remake_the_hosted_sites.md`). 3 protected:
  `leopardess.co.uk`, `leopardess.uk`, `cartoon.co.uk`. **Do not start with
  `businessinsurancequotation.co.uk`** — insurance, so compliance layer, and the largest.
- **The Christmas card sender** (register G3/G4) — design the delivery half FIRST; an open
  "send to any address" form is a spam relay. Read `bugs_open/283`.
- **`adversecreditmortgage.co.uk` stays halted** — owner's call.
- **21 portfolio domains have no register row** (as of 2026-08-21).

# 4. TRAPS — read §4a before running ANY reconciliation

## 4a. ⚠ THE RECONCILIATION QUERY IN THE PREVIOUS HANDOFF GIVES FALSE POSITIVES. Use this one.

`orchestration_states` retains COMPLETED/FAILED for **~24 h** (database-cleanup step 3). The old
`runs = 0` query joins stamps to that table, so **any stamp older than the window reads as a dropped
dispatch.** It reported **six** false hits on 2026-08-25; all six had run and were serving correct
sitemaps. **And the documented remedy — "clear the stamp" — is destructive when fed a false hit:**
it re-probes healthy sites and reads as fleet-wide breakage.

```sql
-- Only stamps INSIDE the retention window can be judged. Older ones are UNKNOWABLE,
-- which is not the same as "no run" and must never be shown as 0.
SELECT s.domain, r.last_selected_at,
       CASE WHEN r.last_selected_at < now() - interval '20 hours' THEN 'unknowable (orchestration reaped)'
            WHEN EXISTS (SELECT 1 FROM orchestration_states o
                          WHERE o.owner_agent_type='sitemap-refresh'
                            AND o.created_at BETWEEN r.last_selected_at - interval '2 min'
                                                AND r.last_selected_at + interval '10 min')
            THEN 'ran'
            ELSE 'NO RUN — dropped dispatch' END AS verdict
FROM site_discovery_rotation r JOIN sites s ON s.id = r.site_id
WHERE r.agent_type='sitemap-refresh' ORDER BY r.last_selected_at DESC;
```

**For anything older than the window, the only reliable detector is the ARTEFACT**: stamped, has
pages now, and serves no sitemap of ours. That census is §4b.

## 4b. The artefact census — the one detector that does not go stale

1. Fetch `https://<domain>/sitemap.xml`. **`rm` the temp file first** — a failed curl leaves the
   previous body in place and it reads as a fresh result.
2. **Judge the BODY, never the status code.** Three shapes return 200-or-3xx and are not ours: a
   parking provider's file, a `text/html` homepage served for any path, and a redirect.
3. Extract every `<loc>`, strip the origin, and match the path against that site's `pages` rows.
   Ours score **n/n**; the parking file scores **0/1**.
4. ⚠ **Canonicalise BOTH sides of that join** (`/index.html` → `/`) or the check scores the
   canonicalisation fix as a regression — it reported `apis.uk` as NOT OURS with a perfect sitemap.
   **The tell is uniformity: one miss on 26 domains is one systematic difference, not 26 faults.**

## 4c. The rest

- ⚠ **`pages.deployed_at` is UPDATED on every redeploy, not set once.** Any query of the form
  `deployed_at > <some time>` measures rerender churn, not first publication. It made 24 of 24
  domains look affected by a defect that touched two.
- ⚠ **A probe proves FETCHABILITY, never CANONICALITY** — twice over. `/index.html` returns 200
  whether or not it is canonical, and Go's client follows redirects, so `probe_dropped` read 0 for a
  site where every URL redirects away. **Both rules were correctly STATED in the header and
  implemented by neither. A doc comment enforces nothing.**
- ⚠ **DO NOT add `repo_name` to `590`'s `commit_sitemap` config.** Its ABSENCE is what routes each
  site correctly; `resolveGitRepoNameDB` tries explicit config FIRST. 4 of 31 sites are `vm-sites`.
- ⚠ **DO NOT tighten `622`'s guard to mirror `render_sitemap`'s filter.** It is weaker on purpose,
  and the migration's verify block RAISES if `pg.noindex` or `pg.expires_at` appear. A mirrored
  guard needs permanent lockstep and its drift mode is a site **silently never selected**.
- **Scope `MIGRATIONS_DIR` on every apply.** A bare `--apply` aborts at the first refusing file and
  never reaches yours.
- ⚠ **`ls` the migrations directory immediately before naming a file.** This lane's `591` collided
  with another lane's inside one session; the tree had moved 591 → 621 while the file was written.
- **`grep -aq` over `/proc/1/exe` gives FALSE ABSENCES on BusyBox while both controls pass.** Use
  `tr '\0' '\n' < /proc/1/exe | grep -Fc`. The `build provenance` startup line scrolls within hours.

# 5. COMMITS, 2026-08-24/25

`5c9acf1bd` canonicalisation + tests · `0bce1db39` migration `590` (applied) · `ff55133ac` register
**SEO-007** · `5f67b977a` 2 landmines · `78e980876` WRONG_CALLS · `948a5a975` handoff 08-24b ·
`ea1406a71` council follow-through · `5ca817a34` SEO-007 proven · `d958a01f3` handoff correction ·
`54ba65b25` **redirect fix** · `911bceb1c` 2 landmines + WRONG_CALLS · `e0dc1678e` handoff 08-25 +
SUMMARY · `102a64b14` close-out · `689f0c5fa` **migration `622`** (applied).

**Council:** `8a004aab` APPROVED (wiring, 7 advisory — all run down in NOTES 08-24 (c)) ·
`25157bab` APPROVED, zero objections (redirect fix) · `c88f5c0f` **pending** (`622`).

# 6. FILES OF RECORD

**Cold start:** this file → `SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md` →
`SUMMARY_2026-08-21_the_machine_that_writes_the_brief.md` →
`PLAN_2026-08-19_one_flow_three_brief_sources.md` → `README_where_we_are.md` (owner's log) →
`NOTES_portfolio_positioning.md` (evidence: entries `2026-08-24 (b)/(c)`, `2026-08-25`, `(b)`, `(c)`).

**Sitemaps:** `platform/orchestration/actions/render_sitemap_action.go` (+ `_test.go`) ·
`docs/agent_docs/sql_for_agents/590_wire_render_sitemap_into_a_rotation.sql` ·
`docs/agent_docs/sql_for_agents/622_sitemap_rotation_skips_a_site_with_nothing_to_list.sql`
(both + `_ROLLBACK`) · register **SEO-007** / **SEO-002** in
`docs/agent_docs/docs026_concept_register/register/seo.md`.

**Traps written by this lane:** `LANDMINES.md` — 4 entries · `WRONG_CALLS.md` — 2 entries.

**Decisions:** `DECISION_2026-08-20_remake_the_hosted_sites.md` · `REGISTER_positioning.md` ·
`RFC_037` (open).
