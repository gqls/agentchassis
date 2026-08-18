# 184 — LLM-emitted markdown reaches the rendered page as literal `**asterisks**`

**Filed 2026-08-03** by the `mortgagecalculator_couk_adoption` lane, found on the
first page it built. **OPEN, unowned. Low severity, but it is live copy on
production sites and it is trivially detectable.**

## Symptom

A content writer emits markdown emphasis inside a text field. The renderer treats
that field as plain text (correctly — it is not a markdown field), so the asterisks
reach the visitor verbatim:

> Banks evaluate your application using a `**Decision Engine**` (an automated
> algorithm that grades your financial history).

Live at `https://mortgagecalculator.co.uk/guides/first-time-buyer/index.html`
(hero slot) as of 2026-08-03.

## Scope — small, and cross-site, which is the point

Three components fleet-wide, on **three unrelated sites and three different slot
types**, so this is not one agent or one template misbehaving:

```sql
SELECT s.domain, p.url, pc.slot_name,
       substring(pc.content_data::text from '\*\*[A-Za-z][^*]{2,40}\*\*') AS sample
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.content_data::text ~ '\*\*[A-Za-z][^*]{2,40}\*\*';
```

| domain | url | slot | sample |
|---|---|---|---|
| mortgagecalculator.co.uk | /guides/first-time-buyer/index.html | hero | `**Decision Engine**` |
| gaswholesalers.com | /how-pricing-works.html | pricing | `**Recommended next steps:**` |
| webdesign.co.uk | /news/index.html | news-listing | ``**the `animation`**`` |

Note the third: it carries a **backtick code span as well**, so a fix that only
strips `**` leaves that one still wrong.

## Why it is worth a file despite being three rows

It is the cheapest possible class to detect and it is **silent** — every existing
check passes. The page renders, the HTML is valid, the component is structurally
complete, `build_status` reads `deployed`, and nothing in the discovery-check layer
looks at it. The only reason it was found is that a human read the prose.

## Root cause (candidate — NOT yet verified in code)

> `[UNVERIFIED]` I did not trace which writer produced these three, and the three
> come from different agents, so a single culprit is unlikely. The general shape:
> nothing on the write path normalises or rejects markdown syntax in fields that
> are rendered as plain text, and prompts do not forbid it. **Do not quote this
> paragraph as a diagnosis** — it is where to start looking, not a finding.

## Fix candidates, ordered by what closes the door

1. **Detect it.** A discovery check in the `check_*` family, matching
   `\*\*[^*]+\*\*`, `` `[^`]+` `` and `^#{1,6} ` in rendered text slots. Cheap,
   offline, no LLM. This is the one that generalises — it catches the next writer
   that does it, including one that does not exist yet.
2. **Normalise on write** — convert `**x**` → `<strong>x</strong>` for slots whose
   schema says they accept inline HTML, and strip otherwise. Needs care: the
   renderer's escaping rules differ per slot, so this is not a blanket
   `strings.ReplaceAll`, and doing it wrong turns a cosmetic bug into an injection
   surface.
3. **Forbid it in the prompts.** Weakest — it is an instruction, not a control, and
   `LANDMINES`/`WRONG_CALLS` are full of cases where a prompt instruction was
   treated as an enforcement mechanism. Do this *as well as* 1, never instead.

## How to verify a fix

Re-run the query above; expect 0 rows. Then confirm at the **artefact**, not the
DB — `curl` the page and grep the visible text, because `content_data` and
`rendered_html` are separate copies and a repair to one does not imply the other
(see `bugs_open/097`).

## Progress — 2026-08-03/04

**Scope decided**: detect (candidate 1) + prompt hardening (candidate 3).
Normalise-on-write (candidate 2) deliberately deferred — the render path has
zero HTML escaping anywhere (`text/template`, not `html/template`), so a
markdown→HTML converter at write time would be writing into an unescaping
pipe, and mutating the shared `SavePageSectionsAction` choke point changes what
that save guarantees for every writer. Named as future work in the register
entry (CQ-019), not silently dropped.

**Built, not yet enabled/applied/live:**
- `platform/orchestration/actions/discovery_checks/check_literal_markdown.go` —
  new discovery check, dual-surface (`content_data` + `rendered_html`, the
  `check_unverified_claims`/093 precedent), letter-guarded regex patterns
  (bold/code-span/heading) that do not fire on `3 * 4`, `#fff`, `#1 rated`,
  JS `` `${x}` ``. Routes to `page-content-writer` for auto-repair (the
  `check_placeholder_contact` precedent — this is a definite mechanical
  defect, not a judgement call). Retracts via `CheckResult.Resolved`
  following `check_required_fields_missing`'s shape (no hand-rolled status
  filter; `resolveWorkItems` alone owns `workItemClosedStatuses`).
  Unit-tested (`check_literal_markdown_test.go`), `go build`/`go vet`/`go test`
  clean for the package (one pre-existing, unrelated failure in the same
  package — `TestRegisteredVerifiersMatchClaimTimeoutExclusion` on
  `page_canonical_collision` — confirmed via `git stash` to predate this
  change and belong to a different concurrent thread's work).
- `docs/agent_docs/sql_for_agents/303_enable_literal_markdown_check.sql` —
  enables the check on `quality-discovery-agent`. **Apply AFTER the image is
  live** (unregistered check names fail loudly since bugs_open/149 B4).
- `docs/agent_docs/sql_for_agents/304_forbid_markdown_in_text_fields.sql` —
  extends live STRICT RULE 9 of `page-content-writer`'s `generate_content`
  prompt in place (scoped `replace()`, fail-loud verification, backup table),
  measured live 2026-08-03 that `content-writer` and
  `simple-content-writer-with-approval` don't carry this prompt block or the
  `save_page_sections` write path, so they are not touched.
- Concept register `CQ-019` added (`content-quality.md`, `000_concept_index.md`).

**Still to do before this bug can close**: submit to the council gate; build +
push + deploy an agent-chassis image carrying the new check (pod-verify the
symbol); apply migration 303 (image first), then 304; let a discovery run fire
on the three founding sites and confirm `page-content-writer` repairs the three
rows; re-run this file's own SQL (expect 0 rows) AND curl the three live URLs
(artefact-level, per the note above — `content_data` and `rendered_html` are
separate copies). Bug stays OPEN until fixed AND live AND the three founding
rows are verified clean at the artefact, not merely at the DB.

## Progress — 2026-08-04 (continued) — everything BUILT is now LIVE; repair cycle not yet run

**Commits**: `0dd08d6a5` (fix: check + tests + both migrations + register +
this file), `de62a2c63` (gofmt follow-up on the test file). Both carry
`Council-Submitted: eb8f9cc0-3a28-437a-8725-1a785f3d12b5`.

**Council**: still `EXECUTING_STEP` at `review_improvement_guardian` as of
2026-08-04 ~09:15 (last checked). Poll:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='eb8f9cc0-3a28-437a-8725-1a785f3d12b5' AND kind='council_report';
```
If APPROVED, the trailer is already `Council-Submitted:` so `098` credits it
automatically — no amend needed. If REVISE/REJECTED, read the objections
(`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
created_at DESC LIMIT 1`) and judge whether they change anything below; the
code is already on the shared branch and live either way.

**Image: LIVE, verified at the artefact, not just the roll.** A concurrent
whole-fleet `make release` (started 08:53, unrelated to this bug) deployed
`v1.0.1247` at 09:06 — pod-grepped and it carried ZERO occurrences of
`literal_markdown`, because its build read HEAD *before* `0dd08d6a5` landed
(09:05:28) — a live instance of the fleet's own "a roll is not evidence your
fix shipped" landmine, caught rather than assumed. Built `v1.0.1248` from HEAD
myself (`make build-agent-chassis IMAGE_TAG=v1.0.1248` — note another
concurrent session independently built the *same* tag number at the same
time, since "highest seen + 1" is deterministic from shared state; harmless,
just wasted a duplicate build), verified locally
(`docker run --rm --entrypoint sh ... strings /app/agent-chassis | grep -c
literal_markdown` → 11), pushed, bumped the kustomization `newTag`. A further
concurrent deploy cycle (not mine) then moved the fleet to `v1.0.1250`.
**Re-verified 2026-08-04 on the CURRENT pods**
(`agent-chassis-88cf8787-*`, both replicas): `literal_markdown` → 11 on each,
a negative-control nonsense string → 0 on each. **The fix is live fleet-wide,
confirmed at the binary, right now.**

**Migrations 303 and 304: APPLIED and verified live** (2026-08-04):
```sql
-- both return true:
SELECT default_config #> '{workflow,steps,run_checks,config,checks}' ? 'literal_markdown'
  FROM agent_definitions WHERE type='quality-discovery-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
SELECT (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}') LIKE '%Plain string also means NO markdown syntax%'
  FROM agent_definitions WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
303's idempotent guard and 304's fail-loud verify both passed cleanly (no
RAISE fired; 304 printed `OK: rule 9 extended, rules 10 and 14 intact`).

**NOT yet done — this is the actual remaining work:**

1. **No discovery run has fired against the three founding sites yet.**
   `SELECT * FROM site_work_items WHERE item_type='literal_markdown'` returns
   **0 rows** — the check is enabled but nothing has invoked
   `quality-discovery-agent` for these sites since. Checked
   `scheduled_tasks`: every `quality-discovery-agent`/`*-discovery-agent` row
   is a **one-shot, already fired and `enabled=false`** (per-site,
   `oneshot-<agent>-<site-slug>-<date>` naming) — there is no live recurring
   schedule to just wait for. Site ids for the three founding sites:
   `mortgagecalculator.co.uk` = `62b5978e-4271-4589-8e00-4baebfc0447c`,
   `gaswholesalers.com` = `5fe15466-4e2e-4ff2-981e-98c1b7074002`,
   `webdesign.co.uk` = `6b49db8e-d447-4467-8277-4f3018af9897`.
   **Next action**: dispatch `quality-discovery-agent` for each site id —
   either insert a one-shot `scheduled_tasks` row (`fire_message=true,
   enabled=true, target_agent_type='quality-discovery-agent', input_data:
   {"site_id": "<id>"}` — **read what actually polls/fires this table and
   at what cadence before assuming "insert = immediate"**, I did not verify
   that path) or fire a direct kafka message to
   `system.agent.generic.requests` (`action=orchestrate`,
   `agent_type=quality-discovery-agent`) mirroring the envelope pattern in
   `scripts/trigger-landmine-verifier.sh` / `097_TRIGGER_council_review_v1.sh`
   (topic and `input_mapping` will differ — read a live `quality-discovery-agent`
   `agent_definitions` row's `workflow.start_step`/expected `input_data` shape
   first, don't guess it from the council script).
2. **Whether `page-content-writer` can actually service a
   `literal_markdown`-shaped work item is UNVERIFIED.** The risk was named at
   plan time (council submission `risks` field) but never checked against the
   live handler's routing table. Before trusting auto-repair, grep how
   `page-content-writer` dispatches by `item_type` today (it already handles
   `placeholder_contact` per that check's precedent — confirm `literal_markdown`
   either falls into the same generic path or needs an explicit addition).
3. Once items are filed and (auto- or manually-) repaired: re-run this file's
   own SQL (§ "Scope" above, expect 0 rows) **and** curl the three live URLs
   for the literal string (artefact-level — `content_data` and `rendered_html`
   are separate copies, bugs_open/097). Confirm the check's `Resolved`
   retraction actually closes the three items on the next discovery pass
   (`SELECT item_key, status, result->>'reason' FROM site_work_items WHERE
   item_type='literal_markdown'`).
4. **Close-out, once fixed AND live AND verified at the artefact**: move this
   file to `bugs_closed/` — **the number 184 is AMBIGUOUS** (shared with
   `bugs_closed/184_..._three_more_detectors_...md`, an unrelated closed
   case) — `git mv` **both** the old and new paths and name **both** on the
   `git commit` pathspec in one go (the LANDMINES entry on this exact trap:
   a pathspec commit after a bare `git mv` can silently ship a COPY, leaving
   the file in both dirs at HEAD — verify with `git ls-tree -r --name-only
   HEAD -- bugs_open/ bugs_closed/ | grep 184` returning exactly the ONE new
   path). Consider whether this defect class earns a 016b §9 entry (a
   transferable pattern: an LLM writer's markdown syntax leaking into a
   plain-text render surface with zero escaping) — not yet written.
5. No `WRONG_CALLS.md` entry needed from this session: the one design
   correction made (the retraction query's shape, initially copied a
   hand-rolled `status NOT IN (...)` filter from the fable-model plan, caught
   by reading `check_required_fields_missing`'s own header warning against
   exactly that) was caught and fixed **before** committing, not after —
   doesn't meet the bar of "a claim that turned out to be false after being
   acted on".

**Handoff note**: this bug's code/build/migration work is DONE; what remains
is dispatch-and-verify, which needs fresh research into the live dispatch
mechanism (item 1 above) rather than more of what's already been established
here. A fresh session can start directly at item 1 without re-deriving
anything above.

## Progress — 2026-08-04 (continued 2) — discovery DISPATCHED, real findings confirmed, repair not yet triggered

**Item 1 from above is SOLVED — the dispatch mechanism, found and used.**
`system.agent.generic.requests` is the WRONG topic for a direct one-off
dispatch of `quality-discovery-agent`: it resolves to a stub `agent_config`
(`workflow.start_step: "complete"`, description *"No-op — scheduled task
pre_query already did the work"*) and the orchestration fails at
`ensure_site_record` with an empty error — three wasted dispatches before I
caught this by reading `collected_data->'agent_config'` directly rather than
trusting the `FAILED` status alone. **The correct topic is
`system.agent.scheduled.requests`** — found by reading a spent one-shot
`scheduled_tasks` row's own `target_topic` column rather than guessing from
the council-gate trigger script's topic. Same envelope shape otherwise
(`action: orchestrate`, `config: {agent_type: "..."}`,
`input_data: {site_id, domain}`, same kcat headers). Confirmed working: all
three `quality-discovery-agent` dispatches on this topic reached
`status=COMPLETED` and produced real `site_work_items` rows.

**Real findings, not a drill.** Detected `literal_markdown` items:
- `mortgagecalculator.co.uk` — 1 item (the founding hero row).
- `gaswholesalers.com` — 1 item (the founding pricing row).
- `webdesign.co.uk` — **10 items across 10 distinct pages**, not the single
  founding row. Verified genuine, not a check false-positive, by reading the
  `spec->'findings'` payload directly: three `learn-*` coding-tutorial pages
  carry backtick code terms (`` `true` ``, `` `ease` ``, `` `fetch()` ``) and
  the `news` page alone carries **20 findings** — real `#`/`##` headings,
  `**bold**` phrases, and code spans, on both `content_data` and
  `rendered_html`. The bug's original manual query only ever sampled ONE row;
  this site's blast radius was never fully measured before the check existed.
  **This is the check doing exactly what it was built for — finding the class,
  not just the three known instances — and it should be stated as such when
  this bug closes, not narrowed back down to "3 rows".**

Dispatch orchestration ids (for `orchestration_state_audit` lookups if
needed): mortgagecalculator `04c928f4-ea47-4895-b606-fbf5ae6f7b22`,
gaswholesalers `35057952-fc5e-405f-94e8-c7abe6826184`, webdesign.co.uk
`228eff66-7136-43f4-aee2-48dd296e4b7c`.

**NOT yet done — the new item 1, replacing the old one:**

1. **All 12 items sit at `status='detected'`, which is NOT dispatchable.**
   `run_discovery_checks`/`quality-discovery-agent`'s own workflow has only
   three steps (`ensure_site_record` → `run_checks` → `complete`) — no
   triage. Promotion from `detected` to `triaged`/`approved` (the only
   statuses `workItemDispatchableStatuses` in `work_items_common.go` accepts)
   happens via the `triage_detected_items` action, which lives in the
   **`improvement-loop`** agent type, not in `quality-discovery-agent` itself.
   I dispatched `improvement-loop` for all three sites the same way (same
   topic/envelope), but it is a much heavier orchestrator — it spawns its OWN
   design/completeness/quality discovery sub-agents first
   (`spawn_quality_discovery` / `spawn_completeness_discovery` /
   `spawn_design_discovery`, all still `EXECUTING_STEP` at last check, ~20s
   in) before it presumably reaches triage. **Did not wait for these to
   finish** — orchestration ids: mortgagecalculator
   `8fb8d1ef-80c1-4431-8a71-623441c021ce`, gaswholesalers
   `952ee804-78d5-4bc7-a246-8c255a2c6b34`, webdesign.co.uk
   `cb71eca2-cde2-48e2-9fc3-5c25208c0b74`. **Next action for whoever
   continues**: poll those three orchestrations (or just re-poll
   `SELECT status, count(*) FROM site_work_items WHERE
   item_type='literal_markdown' GROUP BY status` for a status other than
   `detected` appearing) until they reach `complete`. If `triaged`/`approved`
   items appear, check whether `page-content-writer` claims and repairs them
   on its own normal cadence, or whether that too needs a manual dispatch —
   **unverified, same caveat as before**: I have still not confirmed
   `page-content-writer`'s routing table accepts `literal_markdown` items
   (grep its `item_type` dispatch/routing before assuming — the check's
   `HandlerAgent: "page-content-writer"` field is a claim, not a proof it will
   be honoured).
2. Once repaired (auto or manual): re-run the bug's own SQL (0 rows expected
   for the ORIGINAL narrow pattern — but given webdesign.co.uk's real extent,
   also re-run `check_literal_markdown`'s full pattern set, not just the
   bold-only sample query, before declaring clean) **and** curl the affected
   URLs (artefact-level, `content_data` ≠ `rendered_html`, bugs_open/097).
   `webdesign.co.uk/news/index.html`, `/learn/code-regex-visualized`,
   `/learn/design-physics-of-ui`, `/learn/index` and 6 more (see
   `site_work_items.summary` for full page names) all need checking now, not
   just the one hero/pricing/news row each site started with.
3. Confirm the check's own retraction (`Resolved`) closes items on the NEXT
   discovery pass once repaired, per its design — don't hand-close them.
4. Close-out steps (move to `bugs_closed/`, ambiguous-number `git mv` caution,
   optional 016b §9 entry) unchanged from the previous progress note.

**Council**: still not resolved as of this update (poll query in the previous
progress section) — not blocking, trailer already correct either way.

## Progress — 2026-08-04 (continued 3) — the real blocker was the dispatch backlog, not routing

**Item 2 from the previous note is answered: routing is generic, not
item_type-specific.** `claim_work_item_action.go:94-105` claims by
`status IN ('triaged','approved')` alone; the handler-existence check
(`:126-214`) only verifies `agent_definitions` has a row for
`handler_agent` — confirmed via `handler_coverage_test.go`'s
`knownHandlerAgents` map, which already lists `page-content-writer` (it
covers `check_placeholder_contact` too, same handler). `page-content-writer`'s
own workflow (`load_work_item_actions.go` grep) has **no `item_type` branching
anywhere** — it works from `spec` alone, so it is structurally
item-type-agnostic. Nothing here was the blocker.

**The actual blocker: `build-dispatch-loop` has no scheduled cadence, and
these sites have large triaged backlogs ahead of the markdown items by
priority.** `SELECT * FROM scheduled_tasks WHERE target_agent_type
='build-dispatch-loop'` → **0 rows, fleet-wide** — nothing fires this
dispatcher on its own; it only runs when something spawns/calls it directly
(a scheduled discovery task, or `improvement-loop`'s own
`spawn_dispatch`/`call_dispatch` steps). The 2026-08-04(continued 2)
`improvement-loop` dispatch for all three sites DID reach and complete
`call_dispatch` (confirmed via `processing_history`, ~6 min runtime for
mortgagecalculator's run), but touched **zero** `literal_markdown` rows —
verified by `attempt_count=0, claimed_by=NULL` on all 12 afterward. Root
cause: `load_work_items`'s current live config caps at **`max_items: 5`**
per invocation, ordered by `priority` (lower = sooner), and these items sat
at priority 40 behind **9 items** (mortgagecalculator), **40 items**
(gaswholesalers), and **113 items** (webdesign.co.uk) of higher-priority
triaged work — `site_work_items` totals 23 / 84 / 251 triaged rows on those
three sites respectively. None of that is `literal_markdown`-specific; it is
a standing fleet-wide gap (no recurring `build-dispatch-loop` schedule exists
at all) that this bug's repair step happened to walk into. **Worth its own
bug filing** — not done here, out of scope for 184, but flagged so it isn't
lost: sites accumulate triaged work indefinitely unless something manually
dispatches `build-dispatch-loop`.

**Action taken**: bumped `priority` to `5` on exactly the 12 `literal_markdown`
rows (`UPDATE site_work_items SET priority=5 WHERE item_type='literal_markdown'
AND status IN ('triaged','approved')` — 12 rows). This reorders only these
rows within the existing, already-verified dispatch mechanism; it does not
touch the mechanism itself and does not affect any other site's queue.

**Dispatch mechanism, confirmed and now scripted.** `system.agent.generic
.requests` is still wrong for this agent type (same stub-resolution issue as
the 2026-08-04(continued 2) note). Working topic is
`system.agent.scheduled.requests`, envelope mirrored exactly from
`cmd/scheduler/main.go`'s `fireTrigger()` (`action=orchestrate`,
`config.agent_type`, `input_data:{site_id,domain}`, and its header set
including `from_agent_type=kafka-scheduler`). **Published via the container
COMMAND, not piped stdin** — `kubectl run -i --rm | kcat -P` is the
`kcat-publish-silently-drops` landmine (loses ~4/5 messages at exit 0);
used `--command -- sh -c "... && echo PUBLISH_OK"` instead and confirmed
`PUBLISH_OK` printed for all three fires. Correlation ids: mortgagecalculator
`446fe9bb-401d-4cf1-a838-d749fde11af3`, gaswholesalers
`96c323a0-a198-4e96-8b11-2a5f43a0fc4b`, webdesign.co.uk
`eb3732c9-7663-47b3-8526-3ae0643548fd`. webdesign.co.uk has 10 markdown items
against `max_items: 5`, so this first round will only clear half of them —
expect to need a second dispatch there.

**Still to do**: poll the three orchestrations to completion, check
`attempt_count`/`error`/`status` on all 12 items, re-dispatch webdesign.co.uk
if items remain `triaged`, then verify at the artefact per the standing
note (`content_data` ≠ `rendered_html`, `bugs_open/097`) and confirm the
check's `Resolved` retraction closes items on the next discovery pass.

## Progress — 2026-08-04 (continued 4) — auto-repair does NOT work: 0 of 12 items actually fixed. Filed as bugs_open/201, this bug is BLOCKED on it

**Do not re-dispatch these items — it will not help.** All three orchestrations
reached a terminal state. Final tally:

```
domain                    | status   | count
gaswholesalers.com        | complete |     1
mortgagecalculator.co.uk  | triaged  |     1   (attempt_count=1, will fail identically)
webdesign.co.uk           | failed   |    10   (attempt_count=3, exhausted)
```

**11 of 12 hard-failed**, all with the identical error
`page-content-writer planned its own sections and none are ready — no page
can be written … see section_plan.reason and bugs_open/087`. Root cause
traced and written up in full at `bugs_open/201`: `build-dispatch-loop`
dispatches `page-content-writer` directly with no `section_plan` in its
`input_mapping`, which is exactly the untested "self-plan, falsy branch"
`bugs_closed/087` added *today* (migration 309) and explicitly flagged as
unproven in its own closing note. `plan_sections` returns `ready_count: 0`
for an already-built page (there is nothing left to *plan*, only to *edit*),
so the workflow hard-fails every time, for any page, regardless of item type.

**The 1 item that reached `status='complete'` is a false positive, not a
success.** Checked the actual artefact:
`page_components` for that page (gaswholesalers `pricing` slot) is stamped
`updated_at = 2026-08-03 22:35:17` — **before today's dispatch** — and still
contains the literal `**Recommended next steps:**` string the item was filed
for. The writer's workflow reached `complete` without writing anything to
this slot. `result` carries no `error` key, so nothing at the work-item layer
would have caught this — it needed a direct artefact check
(`bugs_open/097`'s standing warning, proven right again).

**Net: repair candidate 1 (detect) works and found real findings; the
auto-repair leg of candidate 3's plan (`HandlerAgent: "page-content-writer"`)
does not currently work at all — not for any of the 12 items.** This is not
literal_markdown-specific: `bugs_open/201` shows `check_placeholder_contact`
and `check_component_standards` route to the same handler the same way and
have **never had a single item reach `complete` or `failed`** in production,
so this gap predates 184 and was simply never exercised until now.

**This bug cannot close via its planned repair path until `bugs_open/201` is
fixed** (or 184 is re-routed to a different handler — see 201's fix
candidates). Detection stays live and correctly finds real defects; that part
of the scope decision stands. Left the 11 failed / 1 triaged items as-is
(re-arming them now would just re-fail the same way) — 201's fix section
names how to verify once a real fix lands: require the specific slot's
`content_data` to change, not just the work item to reach `complete`.

Priority bump (to 5) on the 12 items from the previous note is now moot —
left in place, harmless, does not need reverting.

**Correction 2026-08-05**: the council submission for this bug's original fix
(`eb8f9cc0-3a28-437a-8725-1a785f3d12b5`) did not stay `EXECUTING_STEP` as the
last note said — re-checked and it is `status='FAILED'` at
`review_improvement_guardian` (errored, not a REVISE/REJECTED verdict; no
`council_report` artifact was produced, only a `fix_plan` one from
2026-08-04 07:58). Not blocking — advisory only, and irrelevant to the actual
blocker (201). Noted so nobody waits on a verdict that isn't coming from this
run; a fresh submission would be needed if council sign-off is wanted.
Fleet moved to chassis `v1.0.1252` (2026-08-05 09:10) in the interim — checked
`git log` since this bug's last commit (`49e8e3048`): nothing touched
`page-content-writer`, `section_plan`, or `build-dispatch-loop`, so 201's
finding is unaffected by the new build.

---

## CONTRIB 2026-08-07, `bugfix_201` lane — **201 is no longer your blocker, and the repair still does not work. Different reason.**

`bugs_open/201` is fixed, live and proven on both symptoms, so **184's auto-repair leg is
unblocked**. It is also **still not curing this defect**, and that is now demonstrated rather
than suspected.

**The run.** Item `efaa39a2…` (this bug's own `webdesign.co.uk` / `news` finding), re-armed onto
`page-build-handler` per 201's fix, dispatched 2026-08-07 08:28Z on chassis `v1.0.1262`.

**It dispatched cleanly** — no `fail_no_ready_sections`, which was 201 symptom 1 and used to kill
11 of 11 attempts. **The handler rebuilt the page for real**: all three components rewritten at
08:37:26Z (`hero` 304→347 B, `news-listing` 10 232→10 157 B, `call-to-action` 331→308 B,
every md5 changed against a baseline taken beforehand).

**And the markdown came straight back.** Completion was blocked by 201's new completion verifier:

```
completion blocked: post-fix verification found the defect still present:
18 finding(s) still present across 3 component(s); first: slot "news-listing"
field "items[1].summary" pattern bold in content_data — "**the `animation`**"
```

So `page-content-writer`, spawned behind the build handler, **wrote markdown syntax back into the
very text field it was dispatched to clean** — 18 findings across 3 components after a full
regeneration.

**What this means for 184.** Routing was never the whole problem. The detector is right, the
dispatch now works, and the *writer* is the remaining defect: it emits markdown into
`type: text` fields, which is this bug's original diagnosis. Repairing by regeneration cannot fix
that while the regenerator has the same habit — a rebuild is not a repair here. The likely lever
is the writer's own prompt/rules (migration `304_forbid_markdown_in_text_fields.sql` exists and is
**pending**, per the runner's dry run on 08-06 — worth checking whether it is applied and whether
it says what this needs).

**The item is now terminally `failed` at 3/3** and routed to human review, which is correct: three
attempts, defect not cleared. A further attempt needs `attempt_count` reset as well as `status`.

**No action requested of this lane** — flagging that the blocker is gone, the failure has moved,
and the evidence is artefact-level rather than status-level. Full trail:
`docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch/NOTES` (2026-08-07) and
`HANDOFF_2026-08-07_continue_here.md`.

## Contribution 2026-08-17 (`bugfix_277_required_fields_repair` / `bugs_open/083` lane) — the lifetime record of the repair pair, and a consumer notice: the promoter now HOLDS it

Not a new diagnosis — I arrived here from the other end, sizing why a scheduled promoter kept
feeding work to a handler that was not completing it. Three things this lane can add.

### 1. The lifetime record, which nobody had put in one place

`(item_type='literal_markdown', handler_agent='page-build-handler')`, all history:
**1 `complete`, 28 `failed`** — a 3% success rate. For context, that is the worst pair in the
fleet by a distance: the 28 `(item_type, handler_agent)` pairs holding at least one lifetime
complete run **3%**, then 41, 42, 46, 50, 67, 79, 80, 86, 86, 88, 89, 94, 96, 98, 99, 100×12.
There is a clean gap between this pair and every other one.

(There is also a second, separate pair: `literal_markdown → page-content-writer`, 1 complete /
0 failed — the path `bugs_open/201` is about.)

### 2. The single success is REAL, and the failures are real too — both measured at the page

Checked because a council seat asked whether that one `complete` might be a
`bugs_closed/028`-style hollow completion. It is not:

| page | item status | literal-markdown hits in visible text |
|---|---|---|
| `gaswholesalers.com/how-pricing-works.html` (the one complete, 2026-08-15) | complete | **0** (of 8,120 visible chars) |
| `ai-agent-orchestration.com/news.html` | needs_human_review | **9** (6 ATX headings, 2 md links, 1 bullet) |
| `robot-hands.com/gripper-catalog/index.html` | failed | **5** |
| `fundamentallyai.com/news/index.html` | failed | **13** (11 ATX headings, 2 md links) |

Five patterns over the served page's visible text: `**bold**`, `^#{1,6}\s`, `[text](url)`,
leading `-`/`*` bullets, `_italic_`. The three failing pages are the demand control that makes
the zero on the repaired page mean something.

**So this bug's symptom has widened since it was filed.** 184 was filed on `**asterisks**` in a
hero slot. What is live today on these pages is predominantly **ATX headings** (`# A…`, 6 and 11
occurrences) and **markdown links**, on `/news` listing pages. Worth confirming the detector and
the repair both cover the heading and link forms, not just emphasis.

**And a failure here is a real failure, not `201`'s verifier correctly refusing a bad repair.**
The `fundamentallyai.com` item carries `deploy_result.success = true` while the page it deployed
still shows 13 hits: the repair ran, deployed, and did not fix the defect. `robot-hands.com`
shows `iter_1`/`iter_2`/`iter_3` — the full three-strikes path. (Read the outcome at the page,
not at `result`: on these rows `result` is the SPAWN record, not the outcome.)

### 3. CONSUMER NOTICE — this pair is now HELD, so the failures will stop arriving on their own

Migration `444` (2026-08-17, council-approved on corr `05a3d1c8`) added a success floor to
`detected-item-promoter`: a pair with ≥5 terminal outcomes must be succeeding at ≥25% before the
promoter will promote more of its findings. **`literal_markdown → page-build-handler` is held by
that floor** — it was the pair that motivated it, after the promoter fed it 6 items of which 5
failed.

What that means for this lane, stated plainly so it is not discovered as a surprise:

- **New `literal_markdown` findings will sit at `status='detected'` instead of being dispatched.**
  They are not lost, not `unresolved`, and carry no error — they are parked upstream of dispatch.
- **The pair unholds itself automatically** once its ratio recovers past 25%. From 1/28 that
  needs about **9** further successes, so in practice: fix the handler, then promote a few rows
  by hand to re-earn the ratio. The by-hand canary path is unchanged and is how every new pair
  bootstraps.
- **To unhold it deliberately**, either promote rows by hand (`status='triaged'`,
  `pipeline='build'`, stamp `spec.original_pipeline`) or apply
  `sql_for_agents/444_detected_item_promoter_door_closers_ROLLBACK.sql`, which restores the
  pre-444 rule for every pair.
- **This lane is not claiming the bug.** Whether `page-build-handler` is the right repairer for
  `literal_markdown` at all is 184/201's call, and the floor deliberately does not answer it —
  it only stops the queue spending dispatches on the answer.

## Progress — 2026-08-18 (`bugfix_184_literal_markdown` lane) — the bug is CLAIMED; repair goes mechanical

**Taking the bug** (ownership checked: who-owns says "recently active" but every recent
commit is a contribution that explicitly declined to claim it; transcript grep found the
only live session touching `literal_markdown` is the bugfix_277 promoter-floor lane,
working the pair as data). Re-validated first: 71 open items across 6 sites (34
unresolved / 24 failed / 10 detected — newest **2026-08-18**, parked by the 444 floor /
3 needs_human_review), and `fundamentallyai.com/news/index.html` serves 11 raw `#`
headings + 2 raw md links at the artefact, curl-verified today. Migration 304's prompt
rule re-verified live — present, and the defect still occurs. A prompt is not a control.

**The design decision, from the accumulated evidence in this file: stop using an LLM to
fix a mechanical defect.** Every repair attempt so far regenerated content with a writer
that has the same habit (08-07 proof above: 18 findings written back). Removing markdown
markers from a plain-text field is a deterministic string operation, so the repair is now:

1. **One shared primitive** — `datahelpers/literal_markdown.go`: the check's letter-guarded
   patterns (single-sourced; the check now imports them) + `StripLiteralMarkdown`
   (strip-only: `**x**`→x, `` `x` ``→x, `# H`→H, `[text](url)`→text; never inserts a
   character, so CQ-019's injection objection to normalise-on-write does not apply).
   Property-tested: scan(strip(x)) == ∅ — detector, verifier and repair cannot drift.
2. **Detector widened to `md_link`** — the live symptom outgrew the filing: 9 md-link
   components fleet-wide (the largest raw bucket) and `## [title](url)` composites in
   open items. Bullets/italic measured ZERO live — deliberately not added.
3. **Repair = a page rerender.** HandlerAgent re-points to `page-rerender` (already a
   proven work-item handler: 5,044 lifetime completes; the check_misdirected_cta
   `reason` precedent), item spec gains `reason: "literal_markdown"`, migration 473
   opens `check_rerender_mode` to it and sets `strip_literal_markdown: true` on
   `rerender_sections` — so the existing no-LLM rerender loads stored content_data,
   strips it, re-renders and redeploys. Both surfaces heal in one pass; the completion
   verifier (unchanged, item_type-keyed) certifies it.
4. **Prevention** — the same strip, opt-in default OFF, at RenderComponentAction
   (LLM content at birth, both surfaces) and section-editor's two pre-render points;
   migration 474 enables it for page-content-writer and section-editor.

**State right now:** plan + council submission `060bcc0a-1ba5-4525-8fea-03de021e26f5`
(verdict pending); part 1 committed (`019fb0616` + gofmt `5fbe549f7`): the datahelpers
primitive + the inert render-seam hooks. Part 2 (check re-route + rerender strip hook)
is deliberately held: the rerender file currently carries the 299 lane's uncommitted
KEEP #3 hunk whose helpers are not at HEAD, and the re-route must ship in the SAME
image as the strip hook (re-route alone would burn item attempts on an unequipped
handler). Sequenced with that session by direct message: they commit (taking my strip
block as a named passenger), I land part 2 immediately after. Migrations 473/474
(+ROLLBACKs) are authored and committed; both are safe to apply before the image
(flags unread, reason unemitted) but intended to be applied with it.

**Rollout after part 2 + image + 473/474:** two-page canary by hand-promotion
(`status='triaged'`, `handler_agent='page-rerender'`, `attempt_count=0`,
`spec = spec || '{"reason":"literal_markdown"}'` — the 444 consumer notice's documented
bootstrap path), artefact-level verification, then batch promotion; the check's
retraction closes the leftovers on the next discovery pass. The old held pair is left
alone — no 444 rollback; it simply stops receiving items. Consumer notice for the
bugfix_277 lane: expect the held pair to go quiet and a fresh
`literal_markdown → page-rerender` pair to appear and earn its ratio through canaries.

Full plan with evidence and risks:
`docs024_key_docs_latest/bugfix_184_literal_markdown/PLAN_2026-08-18_mechanical_markdown_repair.md`.
