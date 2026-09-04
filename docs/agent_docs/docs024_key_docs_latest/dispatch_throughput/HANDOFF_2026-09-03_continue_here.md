# HANDOFF 2026-09-03 — dispatch throughput / D4 spend governor — CONTINUE HERE (supersedes HANDOFF_2026-09-02)

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` 2026-09-03 entries
(the enable, then the first post-enable watch) → `RUNBOOK_dispatch_throughput.md`
§"Spend governor (D4)" (new today — every governor query now lives there, not in prose) →
`SUMMARY_2026-09-03` (the arc read aloud) → `README_where_we_are.md` (owner's log, last two
entries). Paths relative to `docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## The one-paragraph state (2026-09-03 ~10:50Z)

**D4 IS LIVE. Every precondition is met and nothing is owed by the owner.** The governor was
enabled 10:14:32Z on the owner's word; the budget is $2,000/month (L1 $1,400 · L2 $1,700 ·
L3 $1,900); the owner raised the Anthropic console cap to **$3,000** at ~10:41Z, which closes
the last dependency — the account's hard wall now sits above every governor threshold, so the
staged brake always fires first. State at close: **enabled, level 0, ~$386 MTD, heartbeat
live.** The first post-enable watch is done and clean: dispatch unchanged either side of the
switch (0 refusals, 0 withheld), wiring intact at the right object after this morning's
release, Go halves capability-probed on BOTH current pods with an absent control, and the
**shed staircase proven at all four levels without withholding any live work** (rolled-back
synthetic-level probe: 0 / 51 / 112 / 112 items withheld, class ladder in the owner's ruled
order). **The ONE thing still unproven is the Go loader/claim reading a NON-ZERO level on live
traffic** — which is also option C's gate.

## NEXT — one open question, then the queue

1. **THE OPEN QUESTION, put to the owner 2026-09-03 ~10:50Z and awaiting one word:** prove the
   last link by **inducing** a shed in a controlled ~5-minute window (drop `monthly_budget_usd`
   so MTD crosses 70%, e.g. 500; watch the 120 s task move to L1; watch `governor_withheld_now`
   fill with maintenance/llm rows; watch a loop refuse with `spend_governor_shed`; watch the
   level-change `doc_notes` row; restore 2000; watch it drain) — **or** wait for the real
   crossing, which at ~$124/day arrives ~11 September. Nothing is lost either way: withheld
   items stay `triaged`, burn no attempt, and resume on restore. **Do not induce without the
   owner's word** — it deliberately throttles the live fleet, briefly.
2. **After one real or induced shed is observed: option C unlocks** (trigger interval ≤25 s) —
   its own migration editing VERIFY 2/7's lever in lockstep, its own council round.
3. Then the standing queue, unchanged: deploy batching (D8 interim) · clients-first lane (D2) ·
   Batch API (D6) · D16 retention · per-class maintenance LLM cost (RESEARCH §6) · DNS plan B.

## Daily habits for this lane (now three)

- **584 VERIFY** (zombie NOTICEs benign, TWO reaper spellings; runtime is DB-load-bound).
- **657 VERIFY** — pins the selector md5, now `fcbe8821a2a56512911955735796460e`.
- **NEW: the governor wiring check after EVERY release**, not only after touching the governor
  (RUNBOOK §"Spend governor" — two queries). A release rewrites all 208 live
  `agent_definitions` rows in one statement ~70 s before the pods start (measured 08:56:53Z
  today, no `schema_migrations` row — it is the release's seeding step). The hand-applied
  governor clause survived it, but 674 edited the LIVE row and no repo seed carries the
  clause, so a re-seed that DID overwrite would silently remove the governor's primary gate.

## Traps (new today; the 09-02 + 08-30 + 08-26 handoffs' all still stand)

- **"The selector" names TWO live objects.** The governor clause and 657's pinned md5 live in
  `agent_definitions` (`find_dispatchable_site`); `scheduled_tasks.pre_query` is a DIFFERENT
  query (the wake-up gate, 415/688's lane). Reading the clause off `pre_query` returns a
  confident `false` and an unfamiliar md5 — which reads as *someone reverted your migration*.
- **The selector ends in `LIMIT 1`** — `count(*)` over it is 1 at every shed level, true or
  false. Strip it with an assertable arm. LANDMINES entry + WRONG_CALLS row filed today.
- **A single stale `computed_at` is not a dead heartbeat** — 120 s interval + 30 s tick makes
  ~150 s ordinary; one 211 s read self-corrected to 25 s. Two reads over ~300 s is the signal.
- **Yesterday's pod probe proves nothing after a roll** — re-probe the CURRENT pods, and run
  the absent control SEPARATELY (combined, the exec times out).
- **Handler counts across a window boundary are right-truncated** — do not grade throughput
  on them; loops spawn handlers for minutes after they start.

## Coordination state (unchanged from 09-02 unless noted)

All cross-lane threads closed (413, 414, 314/426, 384). 415's migration 688 may apply any
time — no lockstep owed. Ordering decisions ROUTE TO THIS LANE (owner 09-02); the provisional
no-reorder ruling stands, mechanical revisit trigger = the pin census age tail.

## D4 reference card

Unchanged from HANDOFF_2026-09-02 §"D4 reference card" — read it there (objects, register
AGOV-013, the STANDING GATE on a second `honour_spend_governor` consumer, rollbacks). The one
correction: the selector md5 is now `fcbe8821a2a56512911955735796460e` everywhere, and
`governor_config.enabled` is **true**, so 671's rollback will refuse until it is set false.

---

## UPDATE 2026-09-03 ~11:45Z — the open question is ANSWERED and DONE; option C is unlocked; one new bug

The owner said **"induce it today"**. The window ran 11:14–11:33Z and is fully written up
(NOTES 2026-09-03 induced-shed entry, RUNBOOK §"The induced shed", README midday entry).

- **D4 is PROVEN end to end on live traffic.** L2 held 12m16s; 114–115 items withheld in the
  ruled class order; 3 loops handled 24 items, **all llm-free**, with 100+ llm-bearing items
  eligible and withheld; the Go claim backstop fired once; `build/llm` claims went 2 → 0 → 2
  across before/shed/after; restore clean. **Option C's gate ("one real or induced shed
  observed") is therefore MET — option C is unlocked.**
- **Two numbers that supersede the design's stated 120 s:** onset 156 s, **release 249 s**,
  task cadence ~250 s. Quote these, not 120, when sizing anything against the governor.
- **NEW BUG, and it is the recommended next job: `bugs_open/459` — the level-change alarm never
  fires.** `FOR UPDATE` on the `old` CTE (migration 673) races the same statement's `UPDATE`,
  so the note's INSERT selects nothing. Reproduced by one-token A/B on the live text with a
  third control arm; 672 proved the alarm, 673 silenced it and its verify could not see that.
  **Not fixed here on purpose**: it edits the live task that recomputes spend every ~250 s, and
  an appliable migration is council scope. Fix candidates are ordered in the bug file; whichever
  is taken, the verify must DRIVE a level change and assert the note (672's assertion), not
  check that a token is present (673's).
- **Scope fact to carry into any D4 conversation:** only `build-dispatch-loop` honours the
  governor. `diagnose-dispatch-loop`, `report-dispatch-loop`, `zip-deliverable-dispatch` do not,
  by design. ~1% of claims by count; **by spend unmeasured** and plausibly not small
  (`needs_diagnosis` drives whole diagnosis runs). Measuring that is a good small next task, and
  it is the evidence any second-consumer opt-in would need for the standing architecture gate.

**Revised NEXT:** (1) `bugs_open/459` fix + council round · (2) measure the ungoverned loops'
share of SPEND, not count · (3) option C (trigger interval ≤25 s), gate now met · (4) the
standing queue unchanged.

## UPDATE 2026-09-03 ~12:1xZ — NEXT-2 done, and it reframes D4: the governor reaches ~28% of spend

Measured NEXT-2 (the ungoverned loops' share of SPEND). Answer: 3.0% — but the measurement
turned up something far larger. **By orchestration lineage over 24 h** `[MEASURED 2026-09-03,
$319.67, control: all 4,620 calls' orchestration_ids resolve]`: governed `build-dispatch-loop`
**27.6%**, ungoverned `diagnose-dispatch-loop` 3.0%, **no dispatch-loop ancestor at all 69.4%**
— of which `council-gate` is **$198.38, 62% of all fleet LLM spend**, steady across days.

**So at L3 the governor removes at most ~28% of the burn. It cannot hold spend to a budget.**
It is a dispatch governor, not a spend governor. This corrects what the lane told the owner on
09-02; the plain-English correction is in `README_where_we_are.md` (09-03 early afternoon), the
evidence and the refuted first method in NOTES, and **register AGOV-013 + the concept index are
corrected** (their status still read "INERT (stage A)" today, which council seats read as truth).

⚠ **Do NOT re-derive this from `llm_call_log.work_item_id`** — that column is propagated one
generation and lost; 2,278 of 2,278 grandchild-of-a-loop calls carry none, and it yields a
confident, wrong 93.7%. Lineage is the method; retention (~24–27 h) is its bound. WRONG_CALLS
row filed.

**Revised NEXT, in order:**
1. **The owner's choice on coverage** (README states it): extend the governor to the
   council/verifier dispatch path — architecture-scope, own review round — or accept that it
   protects the site-building half only. **This is now the top D4 question, ahead of 459.**
2. `bugs_open/459` — the level-change alarm never fires. Fix + council round.
3. Option C (trigger interval ≤25 s) — gate met.
4. The standing queue unchanged.

## UPDATE 2026-09-03 ~17:2xZ — owner ruled "extend it"; D4b stage A LIVE (mig 751); council round dc6d2a54 in flight

- **D4b stage A applied 17:12:21Z, recorded, inert** — see NOTES 17:0x entry for the seven
  design decisions and the four proofs. `governor_admits` was REWRITTEN live (proven
  equivalent); post-apply canary clean at +4.5 min.
- **Council corr `dc6d2a54-bd73-4827-8267-49c5500467ac`** — architecture round, six edits incl.
  the stage-B sketch. Find it by payload; budget ~30 min. Read the verdict IN FULL before
  acting; a REVISE is cheaper than the defect it finds.
- **Stage B is NOT written.** Seam: `platform/messaging/processor.go:1798 executeWorkflow`,
  after `resolvedAgentType`, before `RecordAgentRun`. Flag `honour_spend_governor_run` on the
  agent's config, default OFF; on refusal INSERT `governor_withheld_runs`, log, `return nil`;
  fail-open. Four Go tests named in the submission. Config half = a `_HOLD` migration setting
  the flag on `council-gate` ONLY with a fleet negative control, applied after the roll. **Do not
  write stage B until the verdict is read** — the round may change its shape.
- **TWO OPEN OWNER QUESTIONS, put to them in README (evening entry):** (1) the council-gate
  shed LEVEL — seeded L3 ('research'); L1 is one UPDATE; (2) whether the other ungoverned agent
  types (landmine-verifier, auditors) get mapped — each is its own row and its own review.
- **`bugs_open/459` (alarm never fires) is unchanged and still the next fix after D4b.**
- **CONTRIB received** from the mortgagecalculator lane (starvation instance 3; two sites hold
  ranks 1–2 while being heavily served). Input to the no-reorder ruling's revisit trigger; not
  acted on. Read it before touching ordering.

**Revised NEXT:** (1) read dc6d2a54's verdict; act on it · (2) stage B Go + tests, after the
verdict · (3) the two owner questions · (4) 459 fix + round · (5) option C · (6) standing queue.

## UPDATE 2026-09-03 ~18:3xZ — dc6d2a54 APPROVED; stage B REDESIGNED (config-only); RFC_065 filed

- **Verdict: APPROVED, 4 advisories, none high.** Read it in full before doing stage B — the
  dispositions are in NOTES (17:43Z entry) and RFC_065 §3.
- **Stage B is now CONFIG-ONLY, inside council-gate's own workflow** (guardian's advisory,
  adopted): `gate_spend_governor` → `route_spend_governor` → `note_withheld` →
  `complete_withheld`, then `load_schema_hint` as before. Exact step shapes to mirror are in
  the live row (`load_schema_hint` = query_database, `gate_render` = conditional,
  `complete_invalid` = complete_workflow, `append_verdict` = append_doc_note). md5-guard the
  row like 674. **Ship with it:** DROP `governor_withheld_runs` + `_recent` (reuse_agent —
  `agent_error_log` exists; the orchestration row is the observable now); a daily
  start-step parity check in the 657-VERIFY shape (guardian — 099 `--apply` would erase the
  step); a line in the 097 runbook that `complete_withheld` means withheld, not queued
  (debug_historian). Its own lighter council round (editquality). **`processor.go` is NOT
  touched — the Go sketch is withdrawn.**
- **DO NOT ARM until the owner answers the LEVEL question** (architecture seat + this lane).
  Seeded `research` (L3). Owner asked twice in README; no answer yet.
- **RFC_065** carries the reasoning — cite it, don't excavate migration 751's header.

**Revised NEXT:** (1) owner's level answer · (2) stage B config migration per RFC_065 §4, its
round, then apply · (3) `bugs_open/459` fix + round · (4) option C · (5) standing queue. The
mortgagecalculator CONTRIB (ordering) stays parked against the no-reorder ruling's trigger.

## UPDATE 2026-09-03 ~19:0xZ — stage B WRITTEN + PROVEN + HELD (mig 752); round c400d333 in flight

- **752 HOLD / ROLLBACK / VERIFY committed; NOT applied.** Six proofs in NOTES (18:4x entry).
- **Council corr `c400d333-b117-4861-93bf-cb8dc71504fe`** — the lighter round. Watcher armed.
- **APPLY GATE = two owner facts:** the LEVEL answer (still open) AND an APPROVED verdict.
  Procedure: RUNBOOK §"D4b stage B — mig 752". After apply: daily 752 VERIFY joins the habit;
  first live submission is the canary; an induced L3 probe is the end-to-end proof.
- If the verdict is REVISE, revise 752 and resubmit with `RESUBMIT_CORR=c400d333…`.
**NEXT:** (1) c400d333 verdict · (2) owner's level → apply 752 → suffix drop + record → canary
→ induced-L3 proof · (3) `bugs_open/459` · (4) option C · (5) standing queue.

## UPDATE 2026-09-03 ~19:3xZ — c400d333 APPROVED; 752 REVISED for its advisories (still HELD); ready to apply on the owner's word

- 752 HOLD revised: fail-open on a SUCCESSFUL-but-empty gate query (`FROM (SELECT 1) LEFT JOIN`
  + `COALESCE(admitted,true)`), four refusal arms before the DROP, transaction-safety comment.
  Verify gained an induced-missing-row arm; daily VERIFY gained a one-row arm. Re-proven six
  ways incl. mutation 3d (RFC_065 §3b, NOTES 19:15Z).
- **APPLY GATE is now ONE owner fact: the LEVEL.** The round is approved. On the word: set the
  class if not 'research' → apply per RUNBOOK §"D4b stage B" → daily VERIFY green → suffix
  drop + record → canary → induced-L3 proof. Both trailers for this arc are earned
  (`Council-Reviewed:` dc6d2a54 and c400d333).
- Forward work noted, not built: a documented 4-step template before agent #2 is mapped.

## UPDATE 2026-09-03 ~21:4xZ — owner ruled "first, and loudly": ALL THREE PARTS LIVE; canary + induced-L1 proof in flight

- council-gate → **L1**; **752 APPLIED** 21:24Z (VERIFY green, recorded, renamed); **753 APPLIED**
  21:31Z (alarm fixed, bug 459 — round `83186fd9` in flight, which is also 752's L0 canary);
  **SessionStart banner** wired (`scripts/governor-session-start.py`).
- Daily habit now: 584 · 657 · wiring check · **752 VERIFY** · read the latest `level-change` note.
- **Watch for:** the canary (83186fd9's orchestration passing the gate); then the induced-L1 proof
  (`scratchpad/induced_l1.sh`, log alongside) — alarm 0→1, banner loud, probe → `complete_withheld`
  with 0 LLM calls, restore, alarm 1→0. If the canary FAILS at the gate (row stuck at
  `gate_spend_governor` or routed to withheld at L0) → roll 752 back and investigate the
  conditional's boolean evaluation (editquality's low, c400d333) before anything else.
- **⚠ my half-row (WRONG_CALLS):** 752's verify restored one column of `governor_state`; healed
  at the next tick. If a future verify deletes that row, restore EVERY column and assert it.
- **NEXT:** (1) canary + proof → docs · (2) 83186fd9 verdict · (3) option C (gate met) · (4) the
  4-step template before any second agent type is mapped · (5) standing queue.

## UPDATE 2026-09-03 ~21:5xZ — owner: a fresh chassis rolls within the hour; post-roll checks ARMED
`scratchpad/post_roll_checks.sh` armed against ReplicaSet `75b987cbd7` (the pods since 08:57Z).
The induced-L1 proof is HELD until the checks are green; round `83186fd9` may be killed by the
roll — resubmit from `scratchpad/council_753.json` if its row stops progressing.

## UPDATE 2026-09-03 ~21:5xZ — the gate FELL OPEN on its first live run (42P18); 754 applied; proof + 754's round HELD behind the owner's roll
- **754 applied 21:48:58Z** (`$1::text`); 752 daily VERIFY green incl. new arm 5b (PREPARE).
  Council runs 21:24–21:49Z were ungoverned (fell open) — stated in WRONG_CALLS.
- **Post-roll watcher re-armed against `85c4984f77`** (`scratchpad/post_roll2.log`; notifier
  bzgwek29j). The 13:28Z roll had already happened — current pods proven 3/3 + control.
- **After the roll goes green:** submit 754's round (`scratchpad/council_754.json`) — it is the
  next live canary: its row must show `gate_spend_governor` output (`collected_data ? 'governor'`)
  and NO `__step_error` on the gate. THEN the induced-L1 proof (`induced_l1.sh`).
- 753's round `83186fd9` is running (fell open at the gate — fine for a review); read its verdict.
- 22:1xZ: 753's round APPROVED (83186fd9); lock-wait property induced A/B (NOTES). 459 stays open
  until the induced-L1 proof writes a real `level-change` note. Decision note written (and fixed
  after a backtick mangled it — SQL bodies via file, never `-c "…"`).

## UPDATE 2026-09-04 ~08:0xZ — roll survived (22:06Z, v1.0.1360, both pods proven); 754's canary in flight (b93ca905); induced-L1 proof next
- Everything hand-applied survived the release (674 / 752 / 753) — measured. `post_roll_checks.sh`
  now re-reads the pod list after its 300 s wait (it had probed transient mid-rollout pods).
- **Order from here:** canary b93ca905 clean (gate OUTPUT, no gate error) → `induced_l1.sh` → close 459
  on the real `level-change` note → docs → 754's verdict.
