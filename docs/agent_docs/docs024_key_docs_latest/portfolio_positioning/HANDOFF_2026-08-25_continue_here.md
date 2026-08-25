# HANDOFF — portfolio_positioning — 2026-08-25. **START HERE. First task is in §1.**

Supersedes `HANDOFF_2026-08-24b_continue_here.md` (all of whose tasks are DONE).
Owner read-out: **`SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md`** — new today, and the
current one. `SUMMARY_2026-08-21_the_machine_that_writes_the_brief.md` covers the brief-writer
half and is still accurate.

**Counts carry the date they were counted** (owner ruling 2026-08-22). Re-run anything
load-bearing before quoting it: `git log --since=<date> --diff-filter=A -- <dir>` is what the date
buys you.

---

# 0. THE ONE-PARAGRAPH STATE

Sitemaps are **done and live**. `590`'s rotation swept the fleet unattended for 13.2 hours and
took coverage from **8 of 28 → 26 of 28** live sites, every one a perfect match against its own
`pages` rows. Two defects in the generator were found and fixed along the way. The first is live and closed out —
**zero domains fleet-wide still emit a non-canonical homepage.** The second (`54ba65b25`) is
**APPROVED by the council with zero objections** and **committed but inert until the next chassis
roll** — confirming it is §1 and the only outstanding item on this lane. The rest of the work is the
deliberately-deferred second half (deploy-path wiring) and one recorded observability gap.

---

# 1. ⏭ START HERE — confirm the redirect fix at the artefact, AFTER the next chassis roll

**The council verdict is already in and needs nothing from you: `25157bab-4b6d-40c5-a218-98148b60daf6`
— APPROVED 2026-08-25 10:16:29, "all reviewers approve", ZERO objections.** `54ba65b25` carries
`Council-Submitted:`, so `098` credits it automatically; **do not amend, and do not hand-write
`Council-Reviewed:` on it.**

**The one thing outstanding on this lane: `54ba65b25` is Go, so it is INERT until the next chassis
roll.** It is in HEAD, so any `make build-*` picks it up. Once rolled, the proof is `webdesign.uk`:

```sql
-- force it round rather than waiting for its 3-day slot
DELETE FROM site_discovery_rotation r USING sites s
 WHERE r.site_id=s.id AND r.agent_type='sitemap-refresh' AND s.domain='webdesign.uk';
```

Then wait for the next tick — ⚠ **the task fires every 30 MINUTES**, so check
`scheduled_tasks.last_triggered_at` before concluding anything.

**Expected AFTER the fix:** `url_count` **0** (not 7), `rendered: false`, and **no commit** — the
conditional routes `url_count=0` to `complete`. That is correct for a domain that 302s every path
away. **Before the fix it reported `url_count: 7, probe_dropped: 0`** and committed a sitemap
advertising seven URLs that redirect elsewhere.

```sql
SELECT (collected_data->'sitemap_render_result'->>'url_count'),
       (collected_data->'sitemap_render_result'->>'rendered'),
       (collected_data->'sitemap_render_result'->>'reason')
FROM orchestration_states WHERE owner_agent_type='sitemap-refresh'
  AND collected_data->'sitemap_render_result'->>'domain'='webdesign.uk'
ORDER BY created_at DESC LIMIT 1;
```

# 2. WHAT IS LIVE AND PROVEN

| | evidence |
|---|---|
| **Fleet coverage 8 → 26 of 28** | census by BODY 2026-08-25 10:00; every one of the 26 scores a **perfect n/n** `<loc>`↔`pages` match |
| **Sweep ran unattended 13.2 h** | 2026-08-24 15:32:13 → 2026-08-25 04:44:38, one site per 30-min tick |
| **27 sites, 27 orchestrations, ALL `COMPLETED`** | zero FAILED, zero partial |
| **Zero dropped dispatches** | every rotation stamp reconciles to `runs = 1` (query in §5) |
| **Canonicalisation fix live** | natural experiment: the roll fell BETWEEN tick 1 and tick 2, so `robot-hands.com` (pre-roll) emitted `/index.html` and all 26 later sites emitted `/` |
| **…and the one stale artefact is CLOSED OUT** | `robot-hands.com`'s stamp was cleared, it re-ran at 10:19:48 on the current binary, and now serves `<loc>https://robot-hands.com/</loc>` — 35 locs preserved, **35/35 matching**, file 10 bytes smaller, which is exactly the length of `index.html`. **Fleet re-checked live 2026-08-25 10:22: ZERO domains still emit a non-canonical homepage.** |

**The two sites NOT covered, both understood — do not "fix" either without reading this:**

- `adversecreditmortgage.co.uk` — **correct by design.** Under the owner HALT of 2026-08-18; the
  pre_query excludes it via `locked_at IS NULL`. What it serves is the parking provider's
  1-`<loc>` `/lander` file, which is not ours and never was.
- `webdesign.uk` — 302s every path to `webdesign.co.uk`. Its committed sitemap is wrong, which is
  what exposed the redirect defect. §1b is its fix.

**Steady state, so a quiet rotation does not read as a stall:** with all 27 stamped inside the
3-day threshold, nothing is due until **2026-08-27**. Ticks between now and then correctly find
no work and stamp themselves complete.

# 3. COMMITS THIS LANE, 2026-08-24/25

| commit | what | state |
|---|---|---|
| `5c9acf1bd` | canonicalisation: site root `/index.html` → `/` + 2 tests | **LIVE**, proven at the artefact |
| `0bce1db39` | migration `590`: `sitemap-refresh` agent + rotation | **APPLIED** 2026-08-24 15:31 |
| `ff55133ac` | register **SEO-007** + index row | — |
| `5f67b977a` | 2 LANDMINES (the two 200-returning sitemap traps) | — |
| `78e980876` | WRONG_CALLS: confirmed a fix from one page | — |
| `948a5a975` | handoff 08-24b + NOTES + owner log | superseded by this file |
| `ea1406a71` | council round 2 follow-through; `DO NOT ADD repo_name` comment | — |
| `5ca817a34` | SEO-007 status: proven at the artefact | — |
| `d958a01f3` | handoff correction: a bare `--apply` would never reach `590` | — |
| **`54ba65b25`** | **redirect fix: `CheckRedirect` + mutation-proven test** | **committed, INERT until the next roll** |
| `911bceb1c` | 2 LANDMINES (both instrument faults) + WRONG_CALLS | — |

Council: `8a004aab-…` **APPROVED** (round 2, the wiring — 7 advisory objections, all run down in `NOTES` 2026-08-24 (c)). `25157bab-…` **APPROVED, all reviewers, ZERO objections** (the redirect fix).

# 4. WHAT IS OPEN

- **The deploy-path half of SEO-002's question — now the main remaining item.** A newly published
  page waits up to 3 days for the rotation. Wiring `render_sitemap` into the page-deploy /
  rerender path closes it. **Deliberately deferred until the sweep was proven, which it now is.**
  ⚠ Cost is the reason it was second, and it has not gone away: the probe is one GET per URL, so
  re-probing a whole site on every page change means **135 requests for `webdesign.co.uk`, every
  edit**. Consider debouncing, or probing only the changed URL and merging.
- ⚠ **`check_has_urls` collapses two opposite cases into one silent no-op** (council
  `bug_historian`, medium, ACCEPTED and NOT fixed). `url_count = 0` routes to `complete` whether
  the site **opted out** or the pages query **unexpectedly returned nothing** — and sites carry
  26–135 pages, so the second would be a fault with no work item and no error row.
  **Not "permanent"**: the rotation retries every 3 days, so the cost is a missing SIGNAL, not a
  stuck state. **Fix designed, ready to implement:** give `render_sitemap` a machine-readable
  `skip_reason` (`opted_out` | `no_listable_urls`) and branch on that — **not** on the prose
  `reason` string, which is its own trap. Go change + follow-up migration.
- **`scripts/site-discovery-files.py:132` still emits the non-canonical `/index.html`.** The
  action no longer does. It is driven by nothing (0 `scheduled_tasks`, no CronJob — checked
  2026-08-24), so there is no two-writer race; it is only wrong when hand-run. Small fix.
- **Whether 3 days is the right threshold**, now that real probe cost is observable rather than
  projected.
- **The 22 hosted-site remakes** (`DECISION_2026-08-20_remake_the_hosted_sites.md`). 3 protected:
  `leopardess.co.uk`, `leopardess.uk`, `cartoon.co.uk`. **Do not start with
  `businessinsurancequotation.co.uk`** — insurance, so it inherits the compliance layer, and it is
  the largest.
- **The Christmas card sender** (register G3/G4) — design the delivery half FIRST; an open
  "send to any address" form is a spam relay. Read `bugs_open/283`.
- **`adversecreditmortgage.co.uk` stays halted** — owner's call, nothing technical blocks it.
- **21 portfolio domains have no register row** (as of 2026-08-21).

# 5. TRAPS — the four that cost time here

1. **A 200 on `/sitemap.xml` is not evidence the site has YOURS.** Three shapes fool a status-code
   census: a parking provider's file, a `text/html` homepage served for any path, and a 302.
   **Match every `<loc>` path against that site's `pages` rows** — ours score n/n, the parking file
   scores 0/1.
2. **⚠ …and canonicalise BOTH SIDES of that join, or the check scores your own fix as a
   regression.** After `/index.html` → `/` shipped, 26 domains scored exactly **n−1** and
   `apis.uk` (whose only page IS the homepage) read **0 of 1, NOT OURS** with a perfect sitemap.
   **The tell is uniformity: one miss on 26 domains is one systematic difference, not 26 faults.**
   The working census is in `NOTES` 2026-08-25; it moved the figure 25 → 26.
3. **A probe proves FETCHABILITY, never CANONICALITY** — twice over. `/index.html` returns 200
   whether or not it is canonical, and **Go's HTTP client follows redirects by default**, so
   `probeOK` reported 200 for URLs that 302 away and `probe_dropped` read 0. **Both rules were
   correctly STATED in the header and implemented by neither.** A doc comment enforces nothing.
4. **The rotation stamps BEFORE it fires** (fire-and-forget), so a dispatch dropped in the ~300s
   post-chassis-restart dead-zone leaves a site marked done and unre-selected for 3 days. Reconcile:

   ```sql
   SELECT s.domain, r.last_selected_at,
          (SELECT count(*) FROM orchestration_states o
            WHERE o.owner_agent_type='sitemap-refresh'
              AND o.created_at BETWEEN r.last_selected_at - interval '2 min'
                                  AND r.last_selected_at + interval '10 min') AS runs
   FROM site_discovery_rotation r JOIN sites s ON s.id = r.site_id
   WHERE r.agent_type='sitemap-refresh' ORDER BY r.last_selected_at DESC;
   ```
   `runs = 0` is a dropped dispatch. Remedy — clear the stamp and it returns to the front of the
   queue (`ORDER BY last_selected_at ASC NULLS FIRST`), re-running within one tick:
   ```sql
   DELETE FROM site_discovery_rotation
    WHERE agent_type='sitemap-refresh' AND site_id='<site with runs=0>';
   ```
   ⚠ Only for a stamp you have PROVEN had no run. **Verified clean for all 27 on 2026-08-25.**

**Operational notes that cost time:**
- ⚠ **DO NOT add `repo_name` to `590`'s `commit_sitemap` config.** Its ABSENCE is what routes each
  site correctly; `resolveGitRepoNameDB` tries explicit config FIRST. 4 of 28 live sites are
  `vm-sites`. The comment is in the migration.
- **Scope `MIGRATIONS_DIR` on every apply.** A bare `--apply` aborts at `562` (whose guard refuses)
  and never reaches later files — the runner stops on first failure.
- **The rotation task fires every 30 MINUTES**, not every scheduler tick (30s). A newly-cleared
  stamp waits for the task's own interval. Check `last_triggered_at` before concluding anything.
- **`grep -aq` over `/proc/1/exe` gives FALSE ABSENCES on BusyBox images while both controls pass.**
  Use `tr '\0' '\n' < /proc/1/exe | grep -Fc`. The `build provenance` startup line scrolls within
  hours. See `LANDMINES.md`.

# 6. FILES OF RECORD

**Cold start:** this file → **`SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md`** →
`SUMMARY_2026-08-21_the_machine_that_writes_the_brief.md` →
`PLAN_2026-08-19_one_flow_three_brief_sources.md` → `README_where_we_are.md` (owner's log,
appended 08-25) → `NOTES_portfolio_positioning.md` (evidence; entries "2026-08-24 (b)/(c)" and
"2026-08-25").

**Sitemaps:** `platform/orchestration/actions/render_sitemap_action.go` (+ `_test.go`) ·
`docs/agent_docs/sql_for_agents/590_wire_render_sitemap_into_a_rotation.sql` (+ `_ROLLBACK`) ·
register **SEO-007** and **SEO-002** in `docs/agent_docs/docs026_concept_register/register/seo.md`.

**Traps written this round:** `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` — 4 entries
(two 200-returning sitemap traps; Go redirect-following; canonicalisation-breaks-your-own-check).
`WRONG_CALLS.md` — 2 entries (one-page generalisation; trusting an unchecked instrument).

**Decisions:** `DECISION_2026-08-20_remake_the_hosted_sites.md` · `REGISTER_positioning.md` ·
`RFC_037` (binding collision check, still open).
**Domains:** `RUNBOOK_domain_inventory_and_classification.md` · `RESERVED_test_domains.md` ·
`scripts/domains/`.
