# HANDOFF 2026-08-26 (evening) — the council is fully live at the binary, farmerinsurance.uk is dispatched and queued, and three proofs are in flight

**Lane:** `loanzy_uk_example_site` — plus the platform-seat work this session accreted (RFC_056,
the promoter's doors, the record-mode lifecycle). **Supersedes
`HANDOFF_2026-08-25c_376_approved_the_council_has_seats_and_the_evolutionary_switch_awaits_one_word.md`**
(bannered; its §1–§9 are the fuller history). **Read in this order on a cold start:** this file →
`PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md` (executed) →
`architecture_review/RFC_056_…seats_are_the_benchmark.md` (CORRECTION + ADDENDA + FOLLOW-UPS are the
live design record) → `bugs_open/405` → `RUNBOOK_loanzy_uk_example_site.md` ("The verdict queue" +
the re-fire recipe). `NOTES_…` is the full running record, newest at bottom, missteps included —
they are the point.

⚠ **CLOCKS:** the DB is UTC; git log shows local **BST (+01:00)**. Two sessions (this one and the
webdesign platform seat) each manufactured a phantom 1–3h from this. Read `now()` from the DB
before ANY staleness claim.

---

## 1. What is LIVE, at which layer `[all verified 2026-08-26, DB clock]`

| thing | state | proof |
|---|---|---|
| chassis `b34c24f4c` (rolled 20:45:57Z) | carries the ENTIRE RFC_056 stack | three-control ancestry via `service_binary_capabilities` + `git merge-base --is-ancestor` (`f51d3cf5e` in; old-commit in; post-build commit NOT in). **Never the `build provenance` log grep — refuted; never `strings`** |
| migrations **618** (376 floor) · **623/624** (council) · **625** (stamp reset) · **629** (origin door) | ALL APPLIED | each probe now refuses re-apply; 624's post-conditions verified (48 steps, 6 record-mode write steps) |
| `improvement-sweep` | ON since 08-25 21:18:19Z | first firing verified clean; **first night: 209 verdict rows / 9 sites / 0 dispatched** |
| record mode (`filing_mode: record`) | LIVE, exercised | 271+ verdict rows; **0 ever left `deferred` except by hand**; release = RUNBOOK recipe, a HUMAN verb |
| record-mode retraction (`recordModeSilenceRule` + `recurrenceExpected`) | in the binary since the roll; **behaviour unproven** | proof recipe §3.3 below; council trail `04a3ce1f` APPROVED r4 (after 3 REVISE — the trail is the design record) |
| origin stamp + promoter door 5 | stamp in the binary; door live in the pre_query (`origin_ok` ×4) | **§6 proof RUNNING at session end — see §3.1 for the takeover** |
| `farmerinsurance.uk` | site `99cae989-2413-430d-b026-59dfeeb638c0`; `needs_domain_research` **triaged, unclaimed** since 19:03:59Z; zone `ccb2ecd19e653f2b36795bfe066226fb` **active** 20:09Z, NS alexis+leah set by owner | deliberate NO-PROMPT build (owner's word, FCA-adjacency flagged and chosen) |

## 2. The owner's rulings this session — do not re-ask

Apply 618 ✓ · apply 623 + sweep on ✓ · render-audit rotation stays LIVE, css-patch fix routed to
`bugfix_390_cascade_attribution` ✓ · promoter thread found → `bugs_open/405` filed ✓ ·
farmerinsurance.uk: registered, **no-prompt deliberate**, NS set ✓ · sequencing: **garden-tools
re-plan AFTER farmer clears hop two; homegarden.uk stays the UNTOUCHED control.**

## 3. IN FLIGHT at session end — a new chat must TAKE THESE OVER (background watches die with the session)

**3.1 The 405 §6 origin-door proof (direction 1).** A synthetic stamped row is LIVE in the table:
site `99cae989…`, `item_key = 'content_rewrite:405-door-verification'`, status `detected`,
`spec.origin='model_opinion'`, proven pair, inserted just after promoter tick 20:45:47Z. The watch
asserts it is STILL `detected` after ≥2 promoter ticks (~21:16Z DB), then cancels it with a result
note. **If the session died first, finish by hand:**
```sql
-- assert (expect 'detected' and promoter last_triggered_at > 21:15Z):
SELECT status FROM site_work_items WHERE item_key='content_rewrite:405-door-verification';
SELECT last_triggered_at FROM scheduled_tasks WHERE name='detected-item-promoter';
-- then ALWAYS clean up (a stranded synthetic row is somebody's confusing census tomorrow):
UPDATE site_work_items SET status='cancelled', updated_at=now(),
  result = COALESCE(result,'{}'::jsonb) || '{"verification":"405 §6 d1 — held across ticks; cancelled"}'::jsonb
 WHERE item_key='content_rewrite:405-door-verification' AND status='detected';
```
Direction 2 (ticks ran) = the promoter's `last_triggered_at` advancing; direction 2 of the FULL
recipe (a NATURAL promotion — never a synthetic promotable) can ride any later window. Record the
outcome in `bugs_open/405` §6 and, if both directions pass, move 405 toward closed-when-live.

**3.2 The farmer hop-two capture.** Watch armed on the `vertical-exemplar-researcher`
orchestration for site `99cae989…`; on reaching synthesise/floor/terminal it dumps
`collected_data` → `EVIDENCE_2026-08-26_farmerinsurance_hop_two.json` and prints the draw +
`formatted_N.source_count`. **Re-arm in a new chat** (orchestration_states reaps in ~25h — capture
BEFORE the reap): poll
`SELECT current_step,status FROM orchestration_states WHERE collected_data->'input_data'->>'site_id'='99cae989-2413-430d-b026-59dfeeb638c0' AND collected_data::text LIKE '%crawl_exemplar%'`.
This is **376 §11e's first natural test**: a refused host must reach `create_next_item` with
`content_quality:'none'` counted; a below-floor draw must FAIL loudly via `record_exemplar_floor`
→ `insufficient_exemplars` with the three counts in the item's `error`. Test 3 (induced
below-floor) is still owed separately.

**3.3 The record-retraction behavioural proof** (owed; slow by nature): over coming audits, record
rows of a silent seat accrue `result.retraction.silent_runs` and retract at 3 — watch:
```sql
SELECT count(*) FILTER (WHERE result ? 'retraction') || ' streaked / ' || count(*)
FROM site_work_items WHERE spec->>'filing_mode'='record';
```
Full induced recipe: RFC_056 ADDENDUM 3 "Post-roll verification". Also confirm the origin stamp
writes on the next model-seat filing:
`SELECT spec->>'origin' FROM site_work_items WHERE spec ? 'audit_source' ORDER BY created_at DESC LIMIT 5;`
(expect `model_opinion` on post-roll rows).

**3.4 Farmer's QUEUE WAIT — a route finding, and an owner option.** At 20:48Z the classifier item
had waited 1.75h behind **27 sites** with older eligible items; dispatch healthy (277 claims/h,
269 completions/h) but **net drain ~15/h** (producers refill). A fresh site's first item waits
behind the fleet's backlog AGE, not its own priority (the 391 lane's between-sites finding —
`dispatch_throughput/CONTRIB_2026-08-26_from_bugfix_391_…`). **Option put to the owner, not
taken:** a one-item direct dispatch for farmer's classifier (single-item bypass precedent). If he
says go, fire it; otherwise report the measured wait as part of the canary's truth.

## 4. Owed next, in order

1. Finish 3.1 (door proof close-out + 405 update).
2. Farmer: hop-two evidence (3.2) → then the build runs ~10h → then `after_test.sh
   farmerinsurance.uk` (parked-domain control FIRST) + read the council's verdict rows on it — the
   first site BORN under the council.
3. **garden-tools.uk re-plan** (owner-sequenced after farmer's hop two). Baseline to beat is in
   the 08-24 handoffs: 7 serve / 5 × 404 / 0 tables / 0 lists. The structure seat should light it.
4. RFC_056 FOLLOW-UPS 1–6 (the file's block is the list): #1 needs the owner's word (per-site
   growth refusal — apis.uk worked case); #4 is §3 above; #5 design-audit child fail-open; #6 a
   verdict-release surface before record rows accumulate further.
5. 376 §11e test 3 (induced below-floor) — after farmer's natural draws are read.

## 5. Traps this session paid for (beyond LANDMINES; each has a NOTES entry with the check)

- **BST-vs-UTC** (header ⚠). · **The roll kills in-flight loop runs** — seat-failure rows +
  attempt stamps absorb it by design. · **`deferred` holds its dedup key** — a park suppresses
  re-filing until released (works FOR you: apis.uk; and AGAINST you if you forget a release).
- **Verify the payload AT dispatch** — a dry-run validates what is on disk NOW, not what your
  generator meant (one council round burned; the at-dispatch assert is now standard).
- **Watch verdicts at `doc_notes`, never `orchestration_states`** (reaped in minutes; the printed
  RUN_ORCH_ID is not the row id).
- **Read the ROSTER, not the posture** (gates map / checks array / exclusion list say who is
  ENROLLED; status text says who is ELIGIBLE) — the ADDENDUM-1 false claim's root.
- **Say N out loud before interpreting; account for all N** (391 lane) — correct instruments read
  partially were the day's worst failure class.
- **A stopped thing has not yet told you WHY it stopped** — three stopped things, three causes
  (an owner ruling; a dead account; a dead provider). Ask the provider's status page before
  diagnosing its idle consumers.

## 6. Peer seats (live sessions, coordinate not compete)

`offer analyser…` (vigilant) — WII-033, gate 1c, sweep-switch history · `bugs_open/277`-named
session = the **391 lane** · `webdesign-tool-rebuilds` = the OTHER "platform seat"; **their**
continuation handoff is `webdesign_tool_rebuilds/HANDOFF_2026-08-26_platform_seat_continue_here.md`
(the owner has this path too — do not confuse the two seats) · `copy quality two stage` (tone-test
sign-off on record) · `apis.uk` (the growth-refusal worked case; their park holds by dedup).

## 7. Falsifiers

- The synthetic §3.1 row PROMOTED → door 5 is not holding; read the promoter pre_query origin_ok ×4.
- A record row outside `deferred` not released by hand → a door changed; the filing-mode tests name both.
- Post-roll model filings WITHOUT `spec.origin` → the stamp is not writing; check ancestry again per service.
- Farmer's classifier still unclaimed after ~24h → the queue-age finding is worse than measured; escalate with numbers.
- The council filing nothing on farmer once built → `REFERENCE` §11 first bullet still applies.
