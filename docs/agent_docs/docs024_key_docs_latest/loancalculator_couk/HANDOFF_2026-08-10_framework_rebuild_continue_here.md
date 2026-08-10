# HANDOFF — loancalculator.co.uk · **the FRAMEWORK-REBUILD thread** · continue here (2026-08-10)

> ⚠ **THIS DIRECTORY NOW HAS THREE LIVE THREADS.** This one (framework rebuild, opened
> 2026-08-09 evening) · the copy/voice thread (`HANDOFF_2026-08-09b`, its site work DONE,
> its §4a fleet base-prompt change still open) · the `bugs_open/227` platform thread
> (`HANDOFF_2026-08-09`, untouched, still owed). None supersedes another.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
serving   26/26, toolgolden 11/11 exact (last verified 2026-08-09 ~20:20Z)
chassis   fresh build rolled 2026-08-10 (owner said so; contains NOTHING of this thread —
          the FlatURLs commit 57a7fcbb4 postdates it and is inert anyway)
state     REBUILD NOT STARTED. Blocked, by owner choice, on the framework URL fix.
```

## 0. The mission, in the owner's words

Two live false claims were found (§2). Owner: **"please rebuild the site completely
through the framework so we can analyse what is actually wrong in the framework and not
what the cli wrote."** Then three explicit decisions (2026-08-10, AskUserQuestion):

1. **URL problem: "Fix the framework first, then rebuild"** — not accept-the-move, not
   avoid-the-planner.
2. **Target: "In place on loancalculator.co.uk"** — not a clean-room parallel domain.
3. **Locks: "Release everything"** — all 17 page_component locks AND the 3 chrome locks.
   The owner chose the fully pure artefact over preserving the proven calculators and
   their own approved homepage copy. Honour it, but sequence it so nothing is lost
   (§5 backups; §6 order).

## 1. What is DONE (all committed)

- **`bugs_open/241`** — the planner's canonicaliser moves every flat tool/guide URL
  (24/26 here). Mechanism, measurement, verification statement all in the file.
- **The representational half of the fix: commit `57a7fcbb4`** —
  `PageDescriptor.FlatURLs` bool, default false, `nestedOrFlatURL()` helper, five test
  cases; `go test ./platform/orchestration/datahelpers/` green, `go build
  ./platform/orchestration/...` green. **INERT: no caller sets the field.**
- **Council submission `6fdb9ce6-9ee2-4550-86ac-893ca0b44c3f`** — trailer
  `Council-Submitted:` is on the commit. **Verdict unread — reading it is owed** (§6.1).
- **Concept register BLD-018** (`register/build-pipeline.md`) — reached HEAD as a
  same-file passenger in another session's `4451b2a0a` (they committed the register file
  between my append and my commit; content intact, provenance noted here for the trail).
- Earlier same-thread work, committed 08-09: three-arm voice test
  (`fleet_copy_quality/RESULTS_2026-08-09_live_arm_test.md`, commits `2f181a689`,
  `ec1564503`).

## 2. The two false claims (both still LIVE — deliberate)

| where | claim | truth |
|---|---|---|
| footer, all 26 pages — **hand-authored chrome**, `site_components` locked `loancalculator_authored_chrome_20260803`, source `chrome/footer.html` | "Every calculator on this site shows its own arithmetic." | calculators emit 3 headline numbers (`monthly-display`, `total-cost`, `total-interest`), no working |
| `guides/how-loans-are-calculated` closing CTA — **framework-generated, passed the claims gate** | Main Loan Calculator "shows the month-by-month breakdown" | index.html has zero tables/schedule/per-month output |

Owner ruled the hand-writing itself is the error → rebuild. **The claims stay live until
the rebuild replaces them.** If the rebuild stalls more than a few days, propose the
two-line copy fix as an interim — but that is an owner call, do not just do it.
The second claim is the framework-attributable one: the claims gate can check a statistic
against evidence but has NO check for claims about our own software's behaviour. That is
a real gate gap for the audit's findings list.

## 3. Verified ground facts (all [MEASURED 2026-08-10] live unless dated otherwise)

- **Purity baseline is perfect:** all 63 page_components have `created_from='manual'`,
  `source_agent_type IS NULL`, `rendered_html_digest IS NULL`. Nothing framework-written
  exists today. Purity query for afterwards: count components where any of those three
  still hold — expect 0 on a pure rebuild.
- **Calculator arithmetic is NOT at unrecoverable risk** (corrects my own 08-09 claim):
  `js_content` is 0 bytes on all 11 tool components; the arithmetic is inline `<script>`
  inside `html_template`, which `component_versions` DOES snapshot and which exists on
  disk at `$LANE/rewrite/tool-*.html.tmpl`. The platform defect (no lock columns on
  `content_components`; `store_generated_component_action.go:1006` discards js on
  snapshot) is real but is NOT this site's exposure.
- **Planner cannot see the calculators**: they are `component_level='tool'`;
  `build-site-planner`'s load_components selects `component_level IN
  ('section','element')`. The regeneration reach is tool-improver / audit only.
- **16 open items target the tools**: 6 `improve_tool` + 10 `audit_tool` (plus ~40 other
  open items, see the full census in NOTES). Today the locks hold them off. **Park these
  BEFORE releasing locks** (§6.4).
- **No `site_plans` row, no `site_snapshots` row, `sites.style_collection_id IS NULL`,
  `built_from_plan_version` NULL on all 27 pages** → `decideEmit` returns "stale" for
  every page: a submission WILL emit needs_page for everything, and design/composition
  WILL fire.
- **The chrome backup table `site_components_bak_20260803_chromelock` is STALE** (footer
  3141 B vs live 2502 B). Do not treat it as a restore point; take fresh backups (§5).
- **Rebuild mechanics** (from source, 2026-08-10): `save_page_sections` DELETEs
  agent-writable rows and re-inserts (`save_page_sections_action.go:757,904`); locked
  rows survive but unmatched locked slots are RELOCATED to the page foot (`:923-950`).
  Chrome is never DELETEd anywhere in Go; locked chrome UPDATEs are refused. Mig 357
  archives component DELETEs to `page_component_history` EXCEPT when the parent page row
  is deleted (cascade archives nothing) or `rendered_html` is empty.
- **The archived-page-name trap**: `upsertPage` ON CONFLICT does not set `status` — an
  archived row keeping its name gets silently updated in place and never builds. If any
  page rows are archived out of the way, RENAME them (`index` →
  `index-pre-rebuild-20260810`), never just flip status.
- **`take_site_snapshot` post-mig-219 captures chrome + all four lock columns; `revert`
  re-applies CURRENT lock state, not the snapshot's** → reverting after releasing locks
  restores content UNLOCKED. Verify a fresh snapshot shows `locked_captured=17` before
  relying on it; if not, mig 219 is missing and the snapshot is worthless.
- The `improve_tool` blast-radius landmine (2026-08-09, LANDMINES): on adopted sites a
  shared ported-page component id can fan a rewrite across ~154 pages on 3 sites. This
  site's tools are per-page components, but re-check before any tool item is un-parked.

## 4. The one design constraint for the plumbing (next code step)

One site-level flag — recommendation `url_shape: "flat"` in the site's `structure` spec
aspect — read ONCE and passed to BOTH canonicalisation surfaces:
`SyncPagesToDBAction` (`site_db_actions.go:281`) and `write_site_plan_action.go:392`.
They diverged once before and it shipped a regression (comment at
`site_db_actions.go:245-254`). One flag, one read, both callers — or don't ship.
This is a second council submission when written. After building the image, pod-grep a
literal the change ADDS (e.g. `url_shape`) and a negative control, every replica.

## 5. Backups BEFORE anything destructive (none taken yet — rebuild hasn't started)

```bash
SITE=0162cde4-633e-45e9-8ca6-87a6b2fe1d26
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
PSQL() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db "$@"; }
# 1. artefact: tag the deploy repo + tar the site dir (record the sha in NOTES)
# 2. DB: CREATE TABLE loancalc_bak_20260810_pc  AS SELECT * FROM page_components pc ... (pages join on site)
#        CREATE TABLE loancalc_bak_20260810_pages AS SELECT * FROM pages WHERE site_id=...
#        CREATE TABLE loancalc_bak_20260810_sc  AS SELECT * FROM site_components WHERE site_id=...
#        CREATE TABLE loancalc_bak_20260810_cc  AS SELECT * FROM content_components WHERE id IN (site's component_ids)
# 3. off-cluster: pg_dump -t 'loancalc_bak_20260810_*' to a local file (a backup table in
#    the same DB is not a backup against a DB accident)
# 4. SELECT take_site_snapshot($SITE,'manual','<sha>','pre_framework_rebuild','operator');
#    → verify pages=27, chrome=3, locked_captured=17 or STOP (mig 219 check)
```

## 6. The ordered TODO (nothing below is started)

1. **Read the council verdict** for `6fdb9ce6` (budget ~30 min from submission,
   2026-08-10 ~09:5xZ; find by payload not printed id):
   `SELECT current_step,status FROM orchestration_states WHERE
   collected_data->'input_data'->>'fix_correlation_id'='6fdb9ce6-9ee2-4550-86ac-893ca0b44c3f';`
   REVISE/REJECTED → act on it (the code is already on shared HEAD).
2. **Write the plumbing** (§4), tests, council submission #2, commit, build image, roll,
   pod-grep positive + negative.
3. **Seed the flag**: `url_shape:"flat"` into this site's `structure` spec (supersede-
   then-insert pattern, see any SEED_*.sql).
4. **Park the 16 tool-targeting items** (improve_tool + audit_tool → status 'deferred',
   note why) BEFORE any lock is released.
5. **Backups** (§5), all four layers, verified.
6. **Release locks** — chrome 3 + page 17 (owner's "release everything"). Record the
   exact pre-release lock state in NOTES first (it is already in
   `page_components_bak_20260809_stearms` for the homepage 5; capture the rest).
7. **Re-submit through the framework**: `scripts/initial_messages/020_build_pipeline/
   082_submit_domain_unified.sh loancalculator.co.uk --email uk@websy.uk
   --mission-file <write one>` — ⚠ no mission text exists anywhere for this site
   (adopted, never submitted fresh). Draft it from the site's strategy spec + the owner's
   framing; show the owner before firing. Fresh mode runs domain-submitter →
   research → strategy → briefing → planner → needs_page cascade; dispatch is via the
   build-pipeline-trigger heartbeat, publish→start can be ~30 min.
8. **Monitor** (children via `parent_orchestration_id` — the printed id logs nothing,
   LANDMINES), then **verify**: purity query (§3), serving 26/26 with the SAME URL set
   (pre/post URL diff must be empty), toolgolden 11/11 — the golden set survives the
   rebuild only if the calculators are re-planned as tools with the same functions;
   expect this to be the hard part, and the audit's richest source.
9. **The audit itself**: diff what the framework produced against what was there —
   claims (does the gate write a software-behaviour claim again?), voice, structure.
   That's the owner's actual question: what is wrong in the framework.
10. Then the parked threads: fleet §4a base-prompt change (voice thread), self-check
    retest (superseded by the rebuild — the rebuilt writer config is where it lands).

## 7. Corrections this thread has already had to make (read before trusting me)

- 08-09: I claimed calculator js was unrecoverable → **wrong**, see §3. Caught by the
  Plan agent verifying live; confirmed by my own query.
- 08-09: I nearly reported the sentence-ceiling mechanism as proven on a page where it
  could not bind (homepage already compliant). The ceiling test needs
  `guides/can-i-overpay` — now moot pre-rebuild, relevant to the rebuilt writer.
- 08-10: my first STE audit run said 60.5% → 55.7% after removing false positives
  (proper nouns, -ise catch-all). Method notes in `fleet_copy_quality/ste_audit.py`.
