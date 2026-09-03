# HANDOFF — bugs_open/427 + 454 + 428, continue here

Written 2026-09-03, early afternoon, **after the `v1.0.1358` chassis roll**. Supersedes
`HANDOFF_2026-09-03b_continue_here.md` (and, through it, `HANDOFF_2026-09-03.md` and
`HANDOFF_2026-09-02.md`) in this directory. Read only this one to continue; the earlier ones are
the arc, not the state.

**The one-line state: 454's fix is VERIFIED WORKING on the live estate, and 427's page is
blocked at the very last step by another lane's guard that shipped in the same image. Neither
remaining item is this lane's code.**

## 0. Where each bug actually stands

| bug | state | why it is not closed |
|---|---|---|
| **454** | **Fixed, live, PROVEN at the artefact** | The *save* of the proven re-render is refused (see §2). CLAUDE.md's bar is fixed AND live at the artefact; the served page has not changed. |
| **427** | Its own defect **closed and verified**; everything the lane built is confirmed correct | Same blocker — **and `719`/`727`/`728` turn out to be TRANSIENT** (§below). `ff91e666` round 3 **APPROVED**. |
| **428** | Unchanged in substance | One item is a human decision (§1.2); its 687 residual spun out to `bugs_open/460`, unowned. |

## 1. Decisions/actions that need a person, not a session

1. **~~`bugs_open/450`'s scope decision~~ — RESOLVED SAME DAY, 13:32 BST. A CHASSIS ROLL IS NOW
   THE ONLY THING THIS LANE WAITS ON.** That lane removed the tool arm from
   `save_page_sections` (`29b40e8bc`); verified at the source, line 210 now gates on
   `refused && class == refusalOwned`, so `refusalToolPending` no longer fires there while the
   tool arm keeps firing at its other three call sites and migration 164's protection is
   untouched. **The live chassis (`d0252fd4d`) still carries the refusing version**, so 427's
   save keeps failing until an image containing `29b40e8bc` ships.
   ⚠ `DISABLE_TOOL_SHELL_REFUSAL` (`owned_page_guard.go:96`) disarms the arm fleet-wide with no
   build. **It is the owner's call, already put to them by the 450 lane — not this lane's to
   pull.**
2. **A human uses bug 428's release surface** on a real flagged verdict. Live end to end since
   2026-09-02, nobody has clicked it. Worked case: boxingonline's own
   `e3c2b440-c006-40ec-be7a-88d0b689ed1e`.
3. **`bugs_open/460` is unowned** — why `blog-content-planner` stopped on 2026-04-24. It decides
   whether 687's residual is an outage or a truth-telling defect. `[NOT ESTABLISHED]`.
4. **Still open from earlier handoffs**: whether `news_feed_ingestion` extraction should run
   beyond boxingonline.com; arming `event_fixture_completeness` in a live
   `run_checks.config.checks` array.

## 1b. ⚠ CHECKED 2026-09-03 12:54 UTC: NO NEW CHASSIS ROLL HAS HAPPENED

A fresh chassis build was reported as deployed. **It has not reached the cluster**, and this is
recorded as a dated negative so the next session does not re-derive it — or, worse, assume it.

`[MEASURED 2026-09-03 12:54 UTC]`, four independent ways, all agreeing:

| check | reading |
|---|---|
| standing pods | `agent-chassis-554857f96f-{kx69c,mdc6d}`, started **12:06:47Z / 12:07:16Z** — unchanged since the last roll |
| image on those pods | `v1.0.1358` |
| commit they report (grouped, standing pods only) | `d0252fd4d` — both agree |
| deployment spec / kustomize overlay / makefile `IMAGE_TAG` | all `v1.0.1358`, 2/2 updated, 2 ready |

**And the fix that matters cannot be in it, by arithmetic**: `29b40e8bc` was committed
**13:32:23 BST (12:32 UTC)**; the running binary's commit `d0252fd4d` is **12:18:32 BST
(11:18 UTC)** and its pods started at 12:06 UTC — before that commit existed.
`git merge-base --is-ancestor 29b40e8bc d0252fd4d` exits non-zero.

**So the fight-calendar save is still refused, and nothing has changed since §2.** The many pods
started 12:50–12:53Z are *spawned agents*, all on the same `v1.0.1358` image — which is exactly
the shape that reads as "lots of fresh pods, so it must have rolled". It has not.

**Re-run the four checks above before believing any future roll claim.** The landmine this
instantiates is already in `MEMORY.md`: *a "FRESH BUILD" CAN SHIP NO NEW CODE.*

## 2. What was proven this session — do NOT re-derive or re-test

**The chassis carries the fix.** `v1.0.1358`, standing pods `agent-chassis-554857f96f-{kx69c,mdc6d}`,
both reporting `d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`;
`git merge-base --is-ancestor 9831e9ab4 d0252fd4d` exits 0.

**The re-render works.** Dispatch `be75b209-1c52-4563-a7b3-bd00902a0367`,
`reason=section_data_resolved`, against a control captured minutes earlier:

| | before | after |
|---|---|---|
| `event-list` content keys | `content, heading` | `content, heading, `**`items`** |
| `event-list` items | 0 | **1** |
| `event-list` rendered_html | 1,813 B (`md5 ee2ec068…`) | **2,498 B** |
| `hero-tool` content_data | 11 keys | **+`hero_url`, +`background_image`** |

⚠ **The counts are NOT the evidence.** `rerendered:2 carried:0 escalated:false` is exactly what
the *broken* runs reported for a fortnight. Read `sections_metadata` (query in the RUNBOOK).

**The `hero-tool` row is the blast-radius claim demonstrated.** Nothing in 427 or 454 concerned
hero images; that section regained `planSection`'s hero aliasing from an unrelated non-`llm`
source in the same pass. One line, two sections, two sources.

**The save was refused** by `OWNED_PAGE_GUARD` — the predicate genuinely held, so this was never
a claim that the guard was wrong. **Fixed by the 450 lane the same day (`29b40e8bc`), riding the
next roll.**

```
OWNED_PAGE_GUARD: page tool-fight-calendar is page_type=tool with no tool component
```

⚠ **The reach census: use the guard's own predicate, and re-read it — the population DRAINS.**
`[MEASURED 2026-09-03 12:00–12:40 UTC, stable across two readings]` **66 pages / 15 sites**;
the 450 lane read 67/16 slightly earlier and watched a page leave between two of their own.
Tool attachments run at ~50 per 12 hours, and each landing on a refused page removes it. This
lane first reported **58 / 12**, which was a **floor read as a total**: `toolShellPredicateFor`
(`owned_page_guard.go:160-168`) carries `AND cc_g.is_active = true` and my census asked only
whether a `component_level='tool'` row existed at all, so it missed **exactly 9 pages on 5
sites** holding an INACTIVE tool component. The `53`-vs-`54` half was `deployed` versus any
non-`removed`, worth one page. Full account and the WRONG_CALLS entry: `bugs_open/427` §18.
**The check: when the thing you are measuring IS a mechanism, copy its predicate — do not
paraphrase it.** The direction was never in doubt, but nobody should quote 58.

**Council: both correlations APPROVED.** `075cfedd` (454's fix) round 1, advisories adjudicated
in `bugs_open/454` §11. `ff91e666` (427's migrations) **round 3 APPROVED** 12:41:09Z, "6 advisory
objection(s) — none high-severity"; adjudicated in `bugs_open/427` §19.

**⚠ THE TOP OPEN ITEM, found while answering those advisories: `719`/`727`/`728` ARE TRANSIENT.**
`pages.sections` is a CACHE; `site_plan_sections` is the authority. `sync_pages`
(`site_db_actions.go:1276`) writes `sections = EXCLUDED.sections` on any non-empty plan proposal,
and `[MEASURED 2026-09-03]` the current plan (`site_plans bba66eda`) still names
`generic-text-block` at ordering 1 and `advertising` at 2 for this page. **The next sync
overwrites all three migrations in one write and re-arms `check_unresolved_sections`.** NOT
fixed: `site_plan_sections` is relied on as per-plan **immutable** by
`reconcile_site_plan_action.go`'s `decideEmit`, so correcting it is a council/owner decision, not
a migration. Same class as `bugs_open/443`. **Do not close 427 believing the array is durable.**

**⚠ And a premise five council seats repeated is FALSE, so do not concede it again.**
`save_page_sections_action.go` is *not* the typed writer for `pages.sections` — it contains no
`UPDATE pages` and its own first line says it saves to **`page_components`**.
`ReconcileSitePlanAction` is site-scoped and its own comment says the comparison is "deliberately
NOT plan-to-`pages.sections`". The real writers are `apply_gap_plan_action.go` (×2),
`load_page_sections_from_spec_action.go`, `ensure_page_section_layout_action.go` and
`site_db_actions.go`'s upsert, none of which takes an arbitrary array for one page. This lane
quoted the objection into a submission **without grepping it** and four more seats then objected
on that quotation — see `bugs_open/427` §19.1.

**Migrations applied this session**, all by hand + `--record-only`, each rehearsed under
`BEGIN`/`ROLLBACK` **and** induced-failure-proven: `727` (restores the `pages.sections` position
order `719` lost) and `728` (drops the orphan `advertising` declaration that was arming
`check_unresolved_sections` to rebuild this page). Live array: `["hero-tool", "event-list"]`,
indexing exactly onto `hero-tool@1, event-list@2`, armed count 0.

**Fleet measurements taken for the council, worth keeping:**
- **Section index alignment:** 2,719 indexable live rows, **109 misaligned** / **68** pages /
  **21** sites. Disaggregated: 95 have their name elsewhere in the array (offset), 14 absent
  entirely (a different defect), 72 on pages whose declared count differs from their row count.
  ⚠ **The causal share attributable to the `jsonb_agg(DISTINCT)` idiom is UNMEASURED — do not
  read 109 as blast radius.** Containment: **0** of the 109 have an empty `slot_name`, so the
  positional consumer cannot fire on any of them today.
- **The anti-pattern is REUSED:** migrations **248, 252, 255, 266, 267** all rebuild
  `pages.sections` with `jsonb_agg(DISTINCT x)` and no `ORDER BY` — and **267's own header
  recommends it** as "naturally idempotent". Now a LANDMINES entry; nothing lints for it.
- **`advertising` is plan residue fleet-wide:** **ZERO** `page_components` rows anywhere join to
  `function='advertising'`; 3 active pages declare it (all boxingonline); **18** pages on **3**
  sites are armed by `check_unresolved_sections`' predicate.

## 3. The first thing to do when a chassis carrying `29b40e8bc` rolls

```bash
<scratchpad>/fire-page-rerender.sh 4b74ff1f-455a-4bb2-b81d-e1d0ec824f33 section_data_resolved
```
(shape and envelope in the RUNBOOK — it is not committed). Then, in order:

1. Confirm `save_sections` did **not** fail: `SELECT status, current_step, error FROM
   orchestration_states WHERE correlation_id='<CORR>';`
2. Read the **row**: `content_data->'items'` non-empty and `length(rendered_html)` ≈ 2,498.
3. Read the **served page** — `portfolio-sites/boxingonline.com`, then curl it. A DB row is not
   what a customer sees, and `deployed_at` is not proof the bytes changed. Trace the real deploy
   via `gh run list --repo gqls/sites` → `gh run view <id> --log | grep -E "upload |delete "`.
4. Then re-verify `experience_loop`'s nightly reclassification of
   `/tools/fight-calendar/index.html`. **That is the real closing signal for 427.**

## 4. Traps this session hit, so the next one does not

- **Asking which commit the chassis runs.** A bare newest-first read of
  `service_binary_capabilities` returned six rows for a **spawned** `agent-image-build-handler`
  pod still on the OLD commit, minutes after the roll. Filter to the standing pods by replicaset
  name and require the two to agree. Full recipe in the RUNBOOK.
- **`rerendered`/`carried`/`escalated` cannot tell you data moved.** `bugs_open/454` is the
  proof: healthy counts throughout a fortnight of delivering nothing.
- **Capture a control BEFORE dispatching** (content keys, item count, `length()` + `md5()` of
  `rendered_html`), or "it looks populated" has nothing to measure against.
- **`orchestration_states` spans ~24 HOURS** and **`site_work_items` is not the population** —
  closed rows archive into `site_work_items_archive`, so a *successful* mechanism erases its own
  evidence from the live table. A zero in either is not an all-history absence. Use
  `llm_call_log`, which keeps replies verbatim.
- **A pathspec commit takes a same-file passenger.** This session's one-line fix carried another
  lane's uncommitted rework into HEAD and broke the build. `git status --porcelain <file>` before
  committing is the cheap warning.
- **Reviewers judge the SKETCH.** An otherwise clean APPROVED verdict drew one advisory purely
  because the submission sketched one test where two had been written.
- **"It will work once X lands" is a prediction about a CHAIN.** This lane wrote that twice today
  and had verified only its own link both times.
- **A CITATION IS NOT A READ — and quoting an objection propagates it.** This lane conceded
  "you should have used the typed writer" twice, then quoted it into a council submission, and
  four further seats objected on that quotation. One grep refuted the premise. Check a named file
  before you concede to it, and before you repeat it.
- **Grep `/bugs_open/` when you FORM a hypothesis, not when you file.** Two lanes reached 454's
  mechanism from opposite ends 90 minutes apart. CLAUDE.md's rule is aimed one step too late —
  at filing time the duplicate work is already spent.
- **Copy a mechanism's predicate; never paraphrase it.** This lane sized another lane's guard
  with a query about "pages with a tool component" while the question was "pages this guard
  refuses" — a floor, reported in the same units as a total, sent to that lane as the basis for a
  scope decision. The two sentences are different and no result could have revealed it.
- **A guard whose harm is masked by an unrelated defect looks free until the defect is fixed.**
  Refusing those 53 saves cost nothing observable while `454` meant they were writing back
  unchanged bytes anyway. The guard's arrival and the repair vehicle's return to working landed
  in one image, so a latent cost became a real one in a single step — and neither lane could see
  it from its own side. The general form is worth more than this instance.

## 5. Named and deliberately NOT done

- **No fleet repair** of the 109 misaligned rows, and no attribution of them to any one cause.
  `727` fixed ONE page.
- **`728` touches one page.** The other two boxingonline pages carrying `advertising` residue are
  named in its header and untouched (`index` is already `needs_rebuild`).
- **`[NOT ESTABLISHED]`** whether a rebuild would ever realise an `advertising` row — i.e.
  whether those pages are on a permanent re-arm treadmill. Not claimed.
- **`reuse_agent`'s council objection is unanswered by design** — four hand-authored SQL
  migrations now touch a jsonb column that has a typed writer (`save_page_sections_action.go`),
  and `ReconcileSitePlanAction` was never checked as an alternative. Recorded as an objection
  not answered, rather than papered over.
- **Migration `719`'s header is unedited**, including its now-refuted paragraph — it is applied
  and the runner's drift guard hashes it. Corrections live in `bugs_open/427` §14–17.
- **No `090` diagnosis run on 454.** Substituted for, with the reason stated in
  `bugs_open/454` §7 per the 2026-07-31 owner ruling.

## 6. Where everything lives

- `bugs_open/454_HANDOFF_2026-09-03_the_light_rerender_computes_a_section_plan_and_drops_it_so_every_page_is_rendered_from_its_own_stored_data.md` — §11 verdict, **§12 the live proof**.
- `bugs_open/427_HANDOFF_2026-09-02_no_writer_populates_dated_correctable_event_facts_so_boxingonlines_fight_calendar_shipped_empty.md` — §14–17 are today.
- `bugs_open/428_HANDOFF_2026-09-02_site_planner_llm_knowingly_defers_strategy_named_entity_roles_citing_its_own_final_say.md` — §11, the CONTRIB and its addendum.
- This directory: `NOTES_bugfix_427_event_render.md` (four entries today), `README_where_we_are.md`
  (owner's plain prose), `RUNBOOK_bugfix_427_event_render.md` (dispatch + verification recipes),
  `SUMMARY_2026-09-03_bugfix_427_event_render.md` (milestone read-out).
- Fleet: `docs024_key_docs_latest/LANDMINES.md` (two entries added today),
  `WRONG_CALLS.md` (two), `016b_debugging_guide_8_consolidated.md` §9 + §10 index.
