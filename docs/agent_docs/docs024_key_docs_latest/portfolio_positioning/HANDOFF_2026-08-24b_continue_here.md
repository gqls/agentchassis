# HANDOFF — portfolio_positioning — 2026-08-24b. **START HERE. First task is in §1.**

Supersedes `HANDOFF_2026-08-24_continue_here.md` (whose §1 is now done except for one
step). Owner read-out: `SUMMARY_2026-08-21_the_machine_that_writes_the_brief.md` — still
current; no new summary today, because "where we are now" has not yet changed (see §6).

**Counts carry the date they were counted** (owner ruling 2026-08-22). Re-run anything
load-bearing: `git log --since=<date> --diff-filter=A -- <dir>` is what the date buys you.

---

# 1. ⏭ START HERE — apply migration `590`. It is the ONLY thing between here and done.

Everything else in the previous handoff's §1 is built, tested, committed and submitted.
The live-DB write was refused by a session tool-permission classifier — **not by any check
of ours** — so it is the one step left.

```bash
SCR=$(mktemp -d) && cp docs/agent_docs/sql_for_agents/590_wire_render_sitemap_into_a_rotation.sql "$SCR"/
MIGRATIONS_DIR="$SCR" ./scripts/migration/run-migrations.sh          # dry run first
MIGRATIONS_DIR="$SCR" ./scripts/migration/run-migrations.sh --apply
```

⚠ **Scope the dir.** A bare `--apply` takes EVERY pending file, and other lanes have
`_HOLD` migrations queued (`587`, `588`) that are held deliberately.

**Already rehearsed, do not redo:** scoped dry run clean; the probe transaction ran to its
own COMMIT and rolled back; and all four verify guards were **induced** to fail —
typo'd action name, a conditional default that never reaches the commit step, the
`locked_at` guard removed, and `files_field` pointing at a field `render_sitemap` never
writes. Each raised the right exception.

## Then prove it at the artefact — the close condition, unchanged

Not "the migration applied". **A site that did not have a sitemap has one, fetched and
read.** First tick lands within 30 minutes; watch a site come round:

```sql
SELECT s.domain, r.last_selected_at FROM site_discovery_rotation r
JOIN sites s ON s.id = r.site_id WHERE r.agent_type='sitemap-refresh'
ORDER BY r.last_selected_at DESC LIMIT 5;
```

Then fetch that domain's `/sitemap.xml` and **read the body** (§5, trap 1). The
before-figure to beat is **8 of 28** live sites serving a sitemap of ours, as of
2026-08-24.

⚠ **A commit to the repo is not a file on the CDN.** Confirm the fetch, not the commit —
`sitemap_commit_result` says the git step returned, nothing more.

# 2. WHAT SHIPPED TODAY (all committed to `087_towards_multiple_domains`)

| commit | what | live? |
|---|---|---|
| `5c9acf1bd` | canonicalisation fix in `absoluteURL` + 2 tests | **Go — inert until the next fleet roll** |
| `0bce1db39` | migration `590` (+`_ROLLBACK`): `sitemap-refresh` agent + rotation task | **authored, NOT applied — §1** |
| `ff55133ac` | register **SEO-007** + index row | n/a |
| `5f67b977a` | two `LANDMINES.md` entries (verifier armed, 3 dispatched) | n/a |
| `78e980876` | `WRONG_CALLS.md` row | n/a |

The action itself is **already live** in chassis **v1.0.1333** — binary-probed with
controls (registry description PRESENT, invented string ABSENT, known older action
PRESENT). **So §1 needs no image roll.** Only the canonicalisation fix does, and the
rotation regenerates every site on a 3-day cycle, so it self-heals once the roll lands.

# 3. THE COUNCIL ROUND — read the verdict, it was still running when this was written

Round 2 submitted on the same correlation via `RESUBMIT_CORR`. Last seen at
`review_architecture`, EXECUTING_STEP.

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='8a004aab-be85-4d6d-bdb1-4fb114f1d64b' AND kind='council_report'
ORDER BY created_at;                      -- the 2026-08-22 'revise' row is ROUND 1
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

All five commits carry `Council-Submitted: 8a004aab-...`, so `098` credits them
automatically **once it turns approved** — no amend needed, and none is allowed. **If it
comes back REVISE, act on it: the code is already on the shared branch.**

# 4. WHAT ROUND 1 OBJECTED TO, AND WHAT ANSWERS IT

> *"A registered-but-uncalled action reproduces the diagnosed defect in a new form."*

Correct, and `590` is the answer. Four steps, deliberately the **same shape** as
`content-feed-orchestrator`'s RSS trio (render → conditional → `git_commit`), because
`render_sitemap` returns `render_rss_feed`'s `{files, domain}` contract on purpose.

**No `ensure_site_record` step**, unlike `render-audit-agent`: that action can **create** a
site row, and a sweep must never bring a site into existence as a side effect. The
pre_query already yields `site_id` (`cmd/scheduler/main.go:223-229` merges it into
`input_data`).

**Rotation, not the deploy path, and the cost was measured rather than left as a risk for
reviewers** — which is what round 1 got wrong. **[MEASURED 2026-08-24]** 735 listable pages
across 28 live sites (avg 26.3, max 135 on `webdesign.co.uk`). One site per tick ≈ 26 GETs,
steady state ≈ 245/day. The deploy path would re-probe a whole site on **every page change**.
The rotation is also the only half that reaches the **20 sites with no sitemap at all**.

# 5. TRAPS — the two new ones both return 200

1. **A 200 on `/sitemap.xml` is not evidence the site has YOURS.** Three of 28 fool a
   status-code census: `adversecreditmortgage.co.uk` serves the **parking provider's** file
   (171 B, one `<loc>` for `/lander`); `noted.co.uk` serves **27,414 B of `text/html`** — its
   own homepage, for any path; `webdesign.uk` **302s** away. Status codes say 11, the truth is
   **8**. Discriminate by matching `<loc>` paths against that site's `pages` rows — ours score
   17/18 to 98/98, the parking file 0/1.
2. **A probe proves FETCHABILITY, never CANONICALITY.** `/` and `/index.html` served
   byte-identical 92,822-byte bodies, so the duplicate passes the probe exactly as well as the
   canonical. ⚠ **And the naive fix is worse than the bug:** the ROOT canonicalises to `/`
   but a SECTION index keeps the extension (10 of 10 sites agree, all three builders), so
   `TrimSuffix` fixes **27** rows and breaks **228**. Whole-path match only.
3. `locked_at IS NULL` in the pre_query is **load-bearing**. `adversecreditmortgage.co.uk`
   is under the owner HALT of 08-18. **A sitemap commit IS a deploy** — do not remove it.
4. `rm` the temp file before `curl`; a failed fetch leaves the previous body in place.
5. A binary carries ONE commit stamp — use `git merge-base --is-ancestor`, never a grep for
   your own sha. The `build provenance` log line had already scrolled out of `--tail=3000`.

Both new entries are in `LANDMINES.md` with their footprints; the near-miss on trap 2 is in
`WRONG_CALLS.md`.

# 6. STILL OPEN, unchanged from the previous handoff unless noted

- **The deploy-path half of SEO-002's question** — a new page waits up to 3 days for the
  rotation. Now the *only* remaining half. Do it after §1 is proven, not before.
- **Whether 3 days is right**, once real probe cost is observed rather than projected.
- **`scripts/site-discovery-files.py:132` still has the canonicalisation defect** the action
  no longer has. Left deliberately: it is a hand-run script and fixing it was out of scope
  for this round. Small and worth doing.
- **The 22 hosted-site remakes** (`DECISION_2026-08-20_remake_the_hosted_sites.md`). 3
  protected. Do not start with `businessinsurancequotation.co.uk` (insurance → compliance
  layer, and the largest).
- **The Christmas card sender** (register G3/G4) — design the delivery half first; an open
  "send to any address" form is a spam relay. Read `bugs_open/283`.
- **`adversecreditmortgage.co.uk` remains halted** — owner's call, nothing technical blocks it.
- **21 portfolio domains have no register row** (as of 2026-08-21).
- **No SUMMARY today, deliberately.** The five headings would read as they did on 08-21 —
  the sitemap work is not a milestone until a site actually serves one. Write it when §1 is
  proven at the artefact; that IS the inflection.

# 7. Files of record

**Cold start:** this file → `SUMMARY_2026-08-21_…` → `PLAN_2026-08-19_one_flow_three_brief_sources.md`
→ `README_where_we_are.md` (owner's log, appended today) → `NOTES_portfolio_positioning.md`
(evidence, appended today as "2026-08-24 (b)").
**Sitemaps:** `platform/orchestration/actions/render_sitemap_action.go` (+ test) ·
`docs/agent_docs/sql_for_agents/590_wire_render_sitemap_into_a_rotation.sql` (+ `_ROLLBACK`) ·
register **SEO-007** and **SEO-002** in `docs026_concept_register/register/seo.md`.
**Decisions:** `DECISION_2026-08-20_remake_the_hosted_sites.md` · `REGISTER_positioning.md` ·
`RFC_037` (binding collision check, still open).
