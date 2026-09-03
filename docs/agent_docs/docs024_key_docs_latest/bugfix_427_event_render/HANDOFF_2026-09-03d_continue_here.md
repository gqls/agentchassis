# HANDOFF — bugs_open/427 (open) + 428 (contributed, owned by "428" session) + bugs_closed/454, continue here

Written 2026-09-03, mid-afternoon, **after the SECOND (real) chassis roll and the fix's live
verification.** Supersedes every earlier handoff in this directory
(`HANDOFF_2026-09-03c_continue_here.md` and before) — read only this one.

**The one-line state: 454 is CLOSED and moved to `bugs_closed/`. 427's own defect is proven
fixed and deployed to the real artefact. What remains is not code — it is one owner/council
decision and one overnight detector check.**

## 0. Where each thing actually stands

| item | state | what's left |
|---|---|---|
| **`bugs_closed/454`** | **CLOSED** — fixed, council-approved, proven live through a real production save + deploy on TWO independent pages | nothing. Watch for `bugs_open/457` (component_id NULL rows) as a named, separate, smaller residual. |
| **`bugs_open/427`** | Its own defect (empty calendar) **fixed and deployed** — `event-list` items 0→1, real git commit `0cc6da28`, real GitHub Actions upload | (1) **top open item**: 3 migrations are TRANSIENT (§below) — council/owner decision needed. (2) overnight `experience_loop` reclassification — the real closing signal. |
| **`bugs_open/428`** | Now actively owned by a session named **`428`** (resumed ~14:00Z) | Not this lane's job any more. Its 687 residual is `bugs_open/460`, unowned. |

## 1. Decisions/actions that need a person, not a session

1. **`site_plan_sections` immutability, blocking a durable fix for the transient-migrations
   problem.** `pages.sections` is a CACHE; `site_plan_sections` is the AUTHORITY. `sync_pages`
   (`site_db_actions.go:1277-1283`) overwrites the cache from any non-empty plan proposal — and
   the current plan for `boxingonline.com/tool-fight-calendar` still names the OLD composition
   (`generic-text-block`, `advertising`). **The next re-plan silently reverts migrations 719,
   727 AND 728 in one write** and re-arms `check_unresolved_sections`. NOT fixed here because
   `reconcile_site_plan_action.go`'s `decideEmit` relies on `site_plan_sections` being per-plan
   **immutable** to distinguish an unchanged page from a re-composed one — correcting it site-wide
   to fix one page's array is a bigger call than this lane should make unilaterally. Same class
   as `bugs_open/443`. Full account: `bugs_open/427` §19.2–19.4.
2. **A human uses `bugs_open/428`'s release surface** on a real flagged verdict — live since
   2026-09-02, nobody has clicked it.
3. **`bugs_open/460`** (why `blog-content-planner` stopped 2026-04-24) — unowned, unrelated to
   closing 427.
4. **Still open from earlier**: whether `news_feed_ingestion` extraction should run beyond
   boxingonline.com; arming `event_fixture_completeness` in a live `run_checks.config.checks`.

## 2. What was proven this session — do NOT re-derive or re-test any of this

**The chassis carries both fixes.** Second roll, `v1.0.1359`, commit
`3043885191b20a0e9b83594b2002e8805fbe95ec`. `merge-base --is-ancestor` clean for both
`9831e9ab4` (454's fix) and `29b40e8bc` (450's guard fix, which had blocked the save twice
today).

⚠ **The FIRST "fresh build" reported this session was NOT a roll.** Same standing pods, same
commit, started BEFORE the fix's own commit timestamp — recorded as a dated negative before
being superseded. If a build is ever reported again, run the four checks in the RUNBOOK before
believing it; a dozen freshly-started pods on the SAME image is spawned agents, not a roll (this
session was fooled by exactly that shape once).

**The re-render works and the save now completes.** Dispatch `53f08444`: `COMPLETED` through
`save_sections`/`render_page`/`deploy_page`, no `__step_error`. `event-list`: items 0→1, HTML
1,813→2,498 bytes. **The deploy is real, traced past the DB row**: `commit_sha 0cc6da28`, GitHub
Actions run `33771117580` **success**, its own "Sync to B2" step showing the real
delete/upload of `tools/fight-calendar/index.html`.

**The public domain and preview subdomain not yet showing it is EXPECTED, checked via `WebFetch`,
not a new problem.** `boxingonline.com` is pre-handover (`handed_over_at IS NULL`) — not DNS-live
at all. The preview (`.ugg2.com`) is a separate `site-publisher` reconciliation pipeline on its
own tick, established 2026-09-02/03, not chased.

**`bugs_closed/454`'s closure is corroborated by TWO independent lanes, both verified
first-hand, not taken on report:**
- The `components`/`bugs_open/425` lane's two-day-old regression turned out to BE this
  mechanism (three prior `UNVERIFIABLE` diagnosis runs). Their `garden-tools.uk /care`
  experiment is dispositive because the before-state provably lacked the resolved value.
- The `bugs_open/384` lane's `designblog.co.uk` canary **repaired at 12:54:41–43Z** and was
  **independently re-verified in this session, all the way to the served page**: DB row (4/4
  image field populated, 3,327 B) AND `WebFetch` of the live domain (4 real `<img src>`, zero
  broken) both agree. This is the only evidence in the whole file checked past the database.
- **The run's own counts are, again, worthless as evidence** — `rerendered:4 carried:0
  escalated:false` was byte-for-byte what the BROKEN runs on that same page reported. §3's
  argument, demonstrated a third time.

**One enumerated exception survives the close**: `page_components` rows with `component_id
NULL` structurally miss `resolveComponent` and take a carry branch forever — 8 rows / 3 pages,
filed as `bugs_open/457`, someone else's bug.

**A new LANDMINES class was found and documented** during the close itself: a THIRD PARTY's
ordinary, correctly-formed `git commit -- <path>` was turned into a silent deletion by a
concurrent `git mv` invalidating its target seconds earlier. `HEAD` held zero copies of a file
for several minutes. Distinct from the existing same-file-passenger entry (that one is about a
mover dropping half of their OWN commit). Recovery recipe now in the RUNBOOK. The peer who caught
it did the right thing: diagnosed precisely, declined to fix it unilaterally, flagged it instead.

**`ff91e666` (427's three migrations) is APPROVED**, round 3, 12:41:09Z. Answering its
advisories found that **five council seats repeated a FALSE premise** — that
`save_page_sections_action.go` is the typed writer for `pages.sections`. It contains no
`UPDATE pages` at all; it writes `page_components`. `ReconcileSitePlanAction` is site-scoped and
explicitly not for this. This lane had quoted the premise into a submission **without grepping
it**, so four more seats objected on its own quotation — logged as a lesson (`427` handoff §
traps, below).

## 3. Traps this session hit — the full list, so the next one does not re-earn any of them

- **A reported "fresh build" can be no roll at all.** Check pod start time vs. fix commit
  timestamp by arithmetic, not by counting pods.
- **A dozen fresh pods on the same image is spawned agents, not a roll.**
- **`rerendered`/`carried`/`escalated` cannot tell you data moved** — proven wrong twice more
  today, on two different pages, by two different lanes.
- **Capture a control BEFORE dispatching**, or "it looks populated" has nothing to measure
  against.
- **A citation is not a read.** Quoting an unread objection into a submission propagates it —
  four more seats objected on this lane's own quotation of a false premise.
- **Grep `/bugs_open/` when you FORM a hypothesis, not when you file.** Two lanes hit 454 from
  opposite ends 90 minutes apart.
- **Copy a mechanism's predicate; never paraphrase it.** This lane sized another lane's guard
  with the wrong query and reported a floor as a total.
- **A pathspec commit can be made wrong by SOMEONE ELSE'S concurrent `git mv`** — new landmine
  class this session, `git ls-tree -r HEAD` returning zero rows is the tell.
- **`curl` to a live customer domain can `ETIMEOUT` for a legitimate reason** (pre-handover, not
  DNS-live) — check `sites.handed_over_at` before concluding a fix "isn't showing"; `WebFetch`
  succeeded where `curl` failed on the same resolvable-elsewhere URL.
- **A guard whose harm is masked by an unrelated defect looks free until the defect is fixed.**
  The general hazard both lanes converged on today.

## 4. Where everything lives

- `bugs_closed/454_HANDOFF_2026-09-03_the_light_rerender_computes_a_section_plan_and_drops_it_so_every_page_is_rendered_from_its_own_stored_data.md` — the closed case, §15/§15a the closure + corroboration.
- `bugs_open/427_HANDOFF_2026-09-02_no_writer_populates_dated_correctable_event_facts_so_boxingonlines_fight_calendar_shipped_empty.md` — §14–20 are today; §19 is the council/transience finding; §20 is the deploy proof.
- `bugs_open/428_...md` — not this lane's job any more; contributed sections at the foot.
- This directory: `NOTES_bugfix_427_event_render.md` (five entries today), `README_where_we_are.md`
  (owner's plain prose), `RUNBOOK_bugfix_427_event_render.md` (all today's recipes),
  `SUMMARY_2026-09-03_...md` and `SUMMARY_2026-09-03b_...md` (two milestones, same day, genuinely
  different read-outs).
- Fleet: `docs024_key_docs_latest/LANDMINES.md` (two new entries today), `WRONG_CALLS.md` (two),
  `016b_debugging_guide_8_consolidated.md` §9 + §10 index (one new pattern + index row).
