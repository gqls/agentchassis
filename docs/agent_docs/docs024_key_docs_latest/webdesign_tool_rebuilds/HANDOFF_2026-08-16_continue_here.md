# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-16 ~10:15Z. Supersedes `HANDOFF_2026-08-15_continue_here.md`.

Read `PLAN_2026-08-15_…` (design), `RUNBOOK_…` (commands, incl. the new "Is the adopt path live?" block),
`NOTES_…` (evidence + missteps, newest at the bottom), `bugs_open/286` (the pilot's real cause + fix).

## Verified state (2026-08-16 ~10:10Z)

| thing | state |
|---|---|
| fleet binary | **still `v1.0.1303`** at 09:52Z (pods 18:45Z 08-15; deploy image 1303; no newer local image). The owner said a fresh chassis build was deployed — nothing visible corresponded. **Probe before believing any roll** (RUNBOOK). |
| pilot root cause | `create_tool_component` has no attach path → `pages_site_id_name_key`, component self-deleted. 090 CONFIRMED (`3050effc`). NOT the handshake (WRONG_CALLS). |
| fix | committed `88897190e` (Go + 3 tests + `bugs_open/286` + register TL-044 + seed `435_…_HOLD.sql`), council APPROVED `27d0f428`. **INERT** until roll + 435. |
| seed 434 | APPLIED + committed `bc4cd65e7`: 4 producers moved off map-valued `spec_data`; 4 items backfilled; 3 ran clean 22:37–22:40Z (new fork versions + `section_edit` deliveries complete). Served-page grade of those three still owed. |
| ab-test | REVERTED to the ported tool 10:0xZ (fork hollow: 0 visible chars; page was serving 47 raw tags). `page_rerender:owner-gate:tool-ab-test-calculator:revert-to-ported` **queued `triaged`** (expect ~80 min behind the serial dispatcher). |
| 285 | its owning lane is running the induce-a-refusal close-out NOW (fence refused at 09:59:06Z, wrapper intact). Not this lane's. |
| aspect-ratio page | untouched: ported slot `deployed`, no native slot; item key `add_tool_novel_webdesign.co.uk` is free (item `complete`). |

## Next actions, in order

1. **Grade the ab-test revert at the served page** once its item completes: `curl` the URL cache-busted;
   `grep -c '{{\.'` = 0, `grep -c 'ported-page-section'` ≥ 1 (the ported tool IS the tool now),
   `grep -ci 'A/B Test Significance'` ≥ 1, `abc-container` = 0. If the item fails on "assembled to
   nothing", read the ported slot's visible chars (RUNBOOK) before touching anything.
2. **Grade the three re-armed audit fixes at the served pages** (`tool-css-unit-converter`,
   `tool-css-specificity-calculator`, `tool-llm-cost-calculator`): the improver GREW css-unit-converter
   8,257→14,261 chars — read the artefact, not the status; visible-chars check + the tool still boots.
3. **Wait for a roll that carries `88897190e`** (RUNBOOK probe with controls). Then un-HOLD + apply seed
   435, confirm the flag, and **refile the aspect-ratio pilot** (RUNBOOK INSERT — description written
   from the LIVE tool's behaviour). Grade the generator run (`page_adopted=true`, one new
   `page_components` row on the EXISTING page id, no new `pages` row), then retire the ported slot,
   rerender, grade at the artefact. Only after that succeeds ONCE: the batch (PLAN §2), ab-test second.
4. **Do NOT file section-edits at tool slots to "fix a raw tag"** — LANDMINES (a fork with `{{.}}`
   copy fields and `content_data={}` renders hollow). If a rebuilt tool ever serves raw tags, that is a
   generator-output defect: rebuild, don't edit.
5. Owed to others, unchanged from the 08-15 handoff §Owed elsewhere (122 ink lane's audit checks
   ~08-18; mindmap junk text = owner's localStorage, nothing to do).

## Traps this session paid for (all in LANDMINES / WRONG_CALLS / NOTES)

- An empty `final_result` is the workflow's `output_fields` shape, not a failure; read `agent_error_log`
  for the window and check whether the thing EXISTS before naming a mechanism.
- `spec_data` as a MAP is silently unread — census `jsonb_typeof` on any copied `create_work_item` step.
- A tool slot can be 13 KB and have zero visible text; the class-attr floors will not tell you.
- Guard "open items on this page" on the dispatchable statuses only; `unresolved`/`failed` pile up.
- The 4 audit_fix items I re-armed were the SAME class as the 233 dead ones — when a "two-item cleanup"
  turns out to be a class, say so in the seed and the bug file, and fix the producer, not the rows.
