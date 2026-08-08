# NOTES — operator bulk page rebuild (`features_open/021`)

## 2026-08-05 — picked up, scoped, built, tested; one live-fire test still owed

Picked this up after a `features_open/` survey (done while `bugs_open/178`'s
workstream ran dry) found #021's blocker (`bugs_open/070`) had closed
2026-07-27 with nobody following up. User confirmed: go ahead.

**Re-verified every claim in the feature file against the LIVE system before
building anything** (11 days since the filing, ~1500 commits/wk on this repo —
per this repo's own memory practice, an unrefreshed prior-art claim is not
evidence):
- `maintenance_queue`: still exactly 2 rows, both `complete`, max `created_at`
  2026-02-18 — unchanged, confirms nobody has used this since the filing.
- `maintenance-triage` agent: still `is_active=true`.
- `scheduled_tasks` targeting it: still 0 rows.
- `stale-work-item-reaper`'s `pre_query`: **now keys on `updated_at`**, not
  `created_at` — confirms `bugs_closed/070`'s fix is genuinely live, not just
  claimed in its own file.
- `build-pipeline-trigger`: still 120s, enabled.

**Read the actual mechanism in full before designing anything** —
`maintenance_actions.go`, both agents' full workflow JSON from
`agent_definitions`. Found the correction recorded in `PLAN`: this path never
touches `site_work_items`, only `pages.build_status` and `maintenance_queue`
directly. `stale-work-item-reaper` cannot reach it. `bugs_open/070` is not
actually a prerequisite for this path (it is for the operator's ORIGINAL
by-hand workaround, which this feature makes obsolete). This is a genuine
correction to the feature file's own stated reasoning, not just new
information — recorded in `PLAN`, not silently.

Also read `recompose_pages` (`features_open/012`) to check whether its intent
vocabulary already applies here. **It does not** — it lives in
`v3_site_actions.go`'s site-PLAN validation pipeline, a different mechanism
entirely. `page-rebuild`'s own `build_pages_loop` has no re-render-only branch;
it always calls `plan_sections` + a content-writer step. So "intent" (point 4
of the feature file) is currently un-wireable without new Go code in
`page-rebuild` itself — recorded as explicitly deferred in `PLAN`, not silently
dropped.

**Built `scripts/rebuild_pages.sh`** — INSERT a `maintenance_queue` row +
direct kcat dispatch to `maintenance-triage`, modelled on the existing
`090_TRIGGER_needs_diagnosis_v1.sh` conventions (envelope shape, correlation
handling, "don't trust a clean dispatch" caveat).

**Tested it — twice, and it was wrong both times before it was right:**

1. **First run (DRY_RUN=1, but the code path inserted a row regardless):**
   `TASK_ID` captured as `6f454f1c-...\nINSERT 0 1` — the INSERT's own command
   tag leaked into the captured value, because `psql -t` suppresses a SELECT's
   header/footer but NOT a non-SELECT's completion tag. **This is the exact
   landmine `090_TRIGGER_needs_diagnosis_v1.sh` already documents** (its own
   header explains wrapping an UPDATE in a CTE + SELECT for precisely this
   reason) — I wrote a bare `INSERT ... RETURNING` anyway and only caught it by
   actually running the script, not by reading 090's script first as closely
   as I should have. Fixed: wrapped in `WITH ins AS (INSERT ... RETURNING id)
   SELECT id::text FROM ins`, plus a UUID-shape assertion on the captured value
   before trusting it (same discipline 090 uses for its claim check). Cleaned
   up the stray test row by hand (`DELETE FROM maintenance_queue WHERE
   id='6f454f1c-...'`) since nothing else would have.

2. **Second issue, found by the same test run, not by re-reading code:** the
   dispatch used `dry_run:true` in `input_data`, matching what I'd assumed the
   safe/preview path meant. But `maintenance-triage`'s `check_dry_run` step's
   dry branch (`complete_dry_run`) skips straight past
   `prepare_rebuild_dispatches` — the ONLY step that ever reads a
   `maintenance_queue` row. So the "dry run" I dispatched never looked at my
   inserted row at all; it only ran `scan_and_queue`'s own independent
   stale/missing/orphan scan. A `DRY_RUN=1` invocation of my script therefore
   inserted a real row that would sit unpreviewed and unclaimed until some
   LATER real dispatch happened to claim it — a genuine design gap, not just a
   labelling problem. **Corrected: `DRY_RUN=1` now does no DB write and no
   dispatch at all — it prints a local report of what WOULD happen and exits.**
   Only `DRY_RUN=0` touches the database or Kafka.

**Re-tested after both fixes**: `DRY_RUN=1` against `gaswholesalers.com`
(chosen because it already carries 7 pre-existing `needs_rebuild` pages, a good
case to prove the sweep-in warning fires) — output showed the correct 7-page
warning, then the local dry-run report, and a follow-up query confirmed **0
rows** in `maintenance_queue` for that site afterward. Clean.

**Did not fire a real (`DRY_RUN=0`) dispatch this session.** A real dispatch on
`gaswholesalers.com` right now would sweep in those same 7 pre-existing
`needs_rebuild` pages, whose history I do not know, into a 90-minute,
content-regenerating, real-deploy run — the kind of hard-to-reverse,
production-affecting action this repo's own operating norms say to slow down
for, not a unilateral call for a first proof-of-mechanism run. Left as the
explicit next step; see `HANDOFF`.

**Not yet done, not attempted, correctly deferred** (see `PLAN` for the full
reasoning): intent (recompose vs re-render) wiring; a dedicated Kafka topic
for this dispatch type (page-rebuild runs are long, same shape of concern that
got council-gate its own topic — worth measuring after real usage, not
designing for zero data points); the first real live-fire test itself.

## 2026-08-06 — first real (`DRY_RUN=0`) dispatch: CLEAN, mechanism proven end to end

Target chosen deliberately per `HANDOFF`'s own checklist: `vetcomparison.uk`,
page `index` (the site had **zero** pre-existing `needs_rebuild` pages —
confirmed live before firing — so nothing rode along uninvited). Real reason
(operator-supplied, via chat): homepage's vet list read alphabetical/plain,
components looked clunky, page didn't clearly state what the site is for.

`CORRELATION_ID=093164d1-0c31-4033-855c-4e042bfe4e3d`,
`TASK_ID=9758bd8a-b08e-411a-9ef1-af0c9c78b20b`,
`SITE_ID=72b9e3a6-872f-4528-a6d6-7f205ea60f4d`.

**Pre-flight (`DRY_RUN=1`) matched expectations**: no sweep-in warning, no
existing pending tasks. Fired for real immediately after.

**End-to-end result — every link in the chain checked, not just "it
dispatched cleanly":**
- `kcat` did not silently drop it: `orchestration_states` showed
  `EXECUTING_STEP`/`spawn_rebuilder` within 12s of firing.
- Full run: claimed 08:32:16, completed 08:35:48 — **~3.5 minutes**, well
  inside the 5400s per-site step timeout (this was one page, not a batch).
  All 6 `orchestration_states` rows sharing the correlation id (parent
  `maintenance-triage` + the `spawn_rebuilder`/`rebuild_loop` sub-orchestrations)
  ended `COMPLETED`/`complete`. No `error_message` on the `maintenance_queue`
  row.
- `pages.build_status` for `index` flipped to `deployed`,
  `updated_at`=08:35:27 — matches the run.
- **Deployed artefact checked directly, not just the status** (this repo's
  own recurring lesson): `curl`'d the live URL. Response `last-modified:
  Thu, 06 Aug 2026 08:35:49 GMT` — matches the run's completion almost to the
  second, so this is genuinely this run's output, not a cached/unrelated
  deploy. The hero headline changed from "Find a UK Veterinary Practice" /
  a long generic CMA-disclosure paragraph to "Find a UK veterinary practice,
  and see what it discloses before you call." — materially addresses the
  "doesn't clearly state what the site is for" complaint.

**One honest gap in the verification, worth recording rather than glossing
over**: I could not confirm the "alphabetical vet list" complaint against
anything that actually changed. Neither the pre-rebuild nor post-rebuild
`page_component_history` rows for this page show a raw list of practice
names anywhere on the homepage — it has always been `hero` +
`info-card-grid` (3 cards, text byte-identical before/after) + `latest-news`
+ `call-to-action`. The card linking to "Search the directory" points at
`/directory/index.html`, i.e. the `directory-index` page — which is
`build_status='planned'`, **not built**. So a browsable/orderable list of
practices does not exist live anywhere on this site yet; the homepage never
had one to reorder. Worth surfacing to whoever asked for this, in case what
they actually want is the directory page built (a different, larger piece of
work, out of scope for this script).

**Also worth recording as [OBSERVED, unexplained]**: the *live-fetched* hero
headline does not match *either* the before-rebuild or after-rebuild
`content_data` stored in `page_components`/`page_component_history` for the
`hero` slot (both DB records read "Find a UK Veterinary Practice..." — the
old copy; a third, different-again variant "UK veterinary practice
directory..." also appears in the history but is not what's live either).
The `last-modified` match proves the served file is this run's output, so
the discrepancy means the final rendered/deployed HTML is not simply a
template render of `page_components.content_data` as stored — some later
step in the pipeline produces the copy that actually ships. Not chased
further here (out of scope for proving the mechanism); flagging because it
means **do not trust `page_components.content_data` alone to describe what a
page says** — check the deployed artefact, per the same standing lesson.

**Conclusion: the mechanism is proven live.** This is the first time
`maintenance_queue` → `maintenance-triage` → `page-rebuild` →
`mark_maintenance_complete` has run for a real operator request, and it
worked cleanly on the first attempt. See `HANDOFF`'s "after a clean first
real run" section for the follow-on decisions (concept register entry,
feature-file status line — done same session) and the two explicitly
deferred questions (intent wiring, dedicated topic) — neither is triggered by
one clean run of one page.

---

## 2026-08-06 (evening) — INBOUND from the `bugfix_208` lane: your entry point's behaviour on OWNED pages changes at the next roll

Not my workstream — appending because the owner ruling of 2026-07-29 §3 says a shared
mechanism's other consumers must be **told**, not merely measured, and `page-rebuild` is the
mechanism I changed. Nothing here needs action from you; it is a guarantee change you should
know about before you next read a dispatch's output.

**What happened.** Your lane's own pre-flight (the `ai-agent-orchestration.com` rebuild) filed
`bugs_open/208`: `page-rebuild` selects pages with no ownership filter, and its loop order is
`assemble_page → deploy_page (git_commit) → save_sections`, so a `rebuild_policy='owned'`
tool page swept into a rebuild was recomposed by an LLM and **committed over the live tool** one
step before the ownership guard refused the database write. I have taken 208 and fixed it
(`cb7b4d759`, registered as PBP-036, council corr
`5d1dcb10-7929-431e-b9e5-496992ce3229`). Go-only, so it is **inert until the chassis rolls**.

**What changes for `rebuild_pages.sh`, concretely.**

- An `owned` page is now **excluded at selection**. Your script's documented sweep-in behaviour
  (note 2 — a dispatch takes every `needs_rebuild` page on the site regardless of the page list)
  is unchanged for generic pages, but owned pages no longer come with it. **This is the sweep-in
  becoming safe rather than being removed.**
- If an operator names an owned page explicitly, it is **refused, not rebuilt**. That is
  deliberate: an owned page's route is the tool pipeline (`tool-generator` /
  `create_tool_component`) or `apply_section_edit`, never the generic builder.
- The refusal is **visible, not silent**: each excluded page gets a `site_work_items` row of
  `item_type='owned_page_review'`, `status='needs_human_review'`, keyed
  `owned_page_review:<page>` (reconcile's existing namespace, so repeated dispatches converge on
  one row rather than piling up). `get_pages_to_build`'s result also carries
  `owned_pages_excluded` / `owned_pages_excluded_count`.
- **A run that previously died now finishes.** `continue_on_error` is unset on
  `build_pages_loop`, so the old hard refusal at `save_sections` failed the **whole workflow** —
  and since selection is `ORDER BY nav_order, name`, every page after the owned one silently
  never rebuilt. If you ever saw a bulk rebuild stop part-way through a site, that is a
  candidate explanation worth checking against your own runs.
- An owned page that is skipped **stays at `needs_rebuild`** rather than being stamped
  `deployed`. Intentional: the stamp also writes `built_from_plan_version`, which would make
  reconcile treat the page as built and permanently suppress the review item.

**What did NOT change:** nothing about your script, the `maintenance_queue` →
`maintenance-triage` → `page-rebuild` path, or generic-page rebuilds. No agent config was
touched — the refusal reuses the `assembled_page.skipped` protocol `git_commit` already honours.

**If you disagree with the default**, the escape hatch is a step-config key `include_owned:true`
on `get_pages_to_rebuild` — but please read PBP-036 first: its unsafe value is the non-default on
purpose, and the intended answer for a genuinely pipeline-owned page is to change
`pages.rebuild_policy`, which is the auditable statement.

**Live population you may care about:** 14 owned pages sat at `needs_rebuild`/`planned` across 6
domains when I measured (2026-08-06), 13 of them serving working tools. All 13 are intact —
this was fixed before it fired, not after. Baseline with per-body sha256:
`docs024_key_docs_latest/bugfix_208_owned_page_commit_before_guard/BASELINE_2026-08-06_owned_pages_served.txt`.

---

## 2026-08-08 — cross-lane notice from the bugfix_210 lane (not this lane's author)

**Your entry point's guarantee changed** (bugs_open/210 fix, committed 2026-08-08, inert until
the next chassis roll; register PBP-038). Previously, a bulk rebuild whose CONTENT GENERATION
failed for a page was silently stamped `deployed` — the operator's request was forgotten and
the page read as current. Now `update_page_status` refuses that stamp: the page goes honestly
to `needs_rebuild` (retryable), every refusal is countable
(`agent_error_log.error_code='DEPLOY_STAMP_REFUSED_ON_SKIP'`), and after 3 failures since the
last successful deploy the page is PARKED behind an open `page_build_failed` work item
(`needs_human_review`, holds the page's `needs_page:<name>` slot) — so a re-fired bulk rebuild
skips it rather than paying for a fourth LLM failure. A successful deploy auto-closes the park.
Operationally: after a bulk run, pages that failed content now show up in
`SELECT ... FROM site_work_items WHERE item_type='page_build_failed'` instead of vanishing
into a false `deployed`. — bugfix_210 lane
