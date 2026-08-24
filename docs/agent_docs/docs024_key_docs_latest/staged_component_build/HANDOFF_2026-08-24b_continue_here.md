# HANDOFF — 2026-08-24b, fresh chat starts here: **the lane's deliverable is still DONE. What changed today is that `bugs_open/353`'s forward fix WENT LIVE, and a council round found the one gap that mattered.**

**Supersedes `HANDOFF_2026-08-24_continue_here.md`** (and transitively 08-22 / 08-21).
**DOES NOT supersede `HANDOFF_2026-08-20_continue_here.md`** — that file still holds the gate's
terms, baseline, interim reads and the CLOSE-OUT (§2.11). Read it for the evidence trail; do not
re-derive it.

**Read this file only for what MOVED.** §1–§3 of the 08-24 file (the RFC_029 D2 Phase 2 close-out,
the improvement-loop finding, and what leaves this lane) are unchanged and still correct.

---

## 1. WHAT CHANGED TODAY — three things, and the middle one is the one to internalise

### 1.1 `bugs_open/353`'s forward fix is **LIVE** — open item (a) is discharged

The 08-24 file and `bugs_open/353` §11 both said "inert until a roll". **That expired between
writing and reading.** Pods `agent-chassis-8bbb57765-{6q6vp,j5gdd}` are on **v1.0.1332**, rolled
**09:39Z today**, and the binary carries the new arm.

⚠ **The `build provenance` startup line had already scrolled out of `--tail=3000`** — so an empty
grep there means "not in range", not "unstamped". What settles it is the **capability probe**,
which has no shelf life, run with BOTH controls:

| literal | expected | got |
|---|---|---|
| `emitted_ungated_build_enqueued_by_caller` | present iff shipped | **PRESENT** |
| `tool_page_will_not_go_live` (control **+**) | must be present | **PRESENT** |
| `zzz_synthetic_literal_that_cannot_exist` (control **−**) | must be absent | **ABSENT** |

**Standing lesson for this lane's cold-starts: an "inert until the roll" line makes the correct
next action look premature.** Re-probe; never inherit a deploy status from prose.

### 1.2 ⚠ LIVE IS NOT PROVEN — and the clean numbers are NOT the result they look like

Since the roll: **0** withholds (against **5** on 08-23) and **0** rows of the new INFO code.

**Do not bank that zero.** Demand control: `tool-generator` ran **3** times since the roll, all
COMPLETED — and **all three recorded `no_related_pages`**, which returns *before* Guard 2 is ever
reached. Nothing since the roll could have exercised the new arm whatever it does.

**The honest state is live-and-unexercised.** The first non-zero
`emitted_ungated_build_enqueued_by_caller` row closes it, and nothing else does. The INFO row
exists precisely so this stays measurable rather than inferred.

### 1.3 Council round 2 = **REVISE** (editquality), round 3 submitted — corr `642ecc3c` throughout

**The high-severity objection was to MY TEXT, not the code.** Edit 3's sketch still showed
`crossLinkEmitDecision(false, …)` — the round-1 defect — while the committed code has passed the
real `pageLive` since round 2. I fixed the code and left the sketch describing the old code; from
outside those are indistinguishable, which is why the objection was sound. Logged in
`WRONG_CALLS.md`. **The check is nearly free: paste the sketch FROM the committed file
(`git show HEAD:<file>`), never from memory of the change you intended.**

**The `missing` item was the substantive one, and it is now a test.** Nothing pinned the CALL
SITE's arguments, so a literal there was invisible to every test in the file — 353's own failure
mode recurring. `TestCrossLinkCallSitePassesTheRealPageLive` (`027461e3d`) drives the emitter
through sqlmock with the only discriminating setup (**page SERVED, opt-in OFF**) and asserts the
**effect**, never the absence of a query. **Mutation-proved** in a `git archive HEAD` copy:
restoring the literal fails exactly this test and leaves **both** older tests PASSING.

---

## 2. `bugs_open/353` — what it is still open on (§12.6 of the bug file is the authority)

- **(b)** the new arm is **live but UNEXERCISED** (§1.2) — first non-zero INFO row closes it;
- **(c)** `tool-deployer` still has **0 runs in retained history** — the emitter's second caller
  remains an unexercised path, untouched by any of this work;
- **(d)** the **regeneration residual**: `replace_existing` returns before the emitter, so a
  regeneration whose spec ADDS a related page never emits for it. **Unowned.**
- round 3's verdict.

**CLOSED and not to be reopened without new evidence:** the damage half (§11 — 74 items, 61
complete, **12/12 sampled pages serving** at their DB-read URLs), and item (a), the forward fix
pending a roll (§12.1).

> ⚠ **DO NOT RUN THE 51-PAGE REDEPLOY.** An earlier section asked for it and drew owner approval on
> a premise I later retracted: I had **constructed** page URLs instead of reading `pages.url`, so
> every zero — control included — was a miss on a URL that does not exist. One page was redeployed
> (dartsonline `barrel-shapes`, harmless); the other 50 were **not**. `bugs_open/353` §11.

---

## 3. TRAPS EARNED (superset of the 08-24 file's list; 1–6 there still stand)

7. **A pure function's test table proves the FUNCTION, never the CALL.** Extracting a decision to
   make it testable *moves* the untested seam onto its arguments. If a fix's whole value is "this
   branch is now reachable", something must assert reachability THROUGH the caller — and assert the
   **effect**, never the absence of a query (that shape passes vacuously).
8. **A sketch is a claim about what the code says, and it decays the moment you edit the code.**
   Paste it from the committed file.
9. **A post-fix zero needs a DEMAND control before it means anything.** "0 failures since the roll"
   and "nothing since the roll could have reached the changed code" look identical in the count.

---

## 4. SESSION-START CHECKLIST

1. `kubectl -n ai-persona-system get pods -l app=agent-chassis` — one tag, one replicaset? **The
   fix is live as of v1.0.1332; do not re-derive that from prose, re-probe** (§1.1's three-row
   method) if the tag has moved.
2. **The gate is CLOSED — do not re-run a window.** Result: `HANDOFF_2026-08-20_continue_here.md` §2.11.
3. **Read the round-3 verdict** for corr `642ecc3c` (run `53e3812f`, submitted 11:28Z today) and act
   on it. If APPROVED, commit the trailer against it; the code is already committed and live, so
   there is nothing to ship.
4. **Check whether the new arm has fired yet** — this is the only outstanding *measurement* on 353:
   ```sql
   SELECT count(*), min(occurred_at) FROM agent_error_log
    WHERE error_code = 'tool_crosslink_not_emitted:emitted_ungated_build_enqueued_by_caller';
   ```
   Non-zero ⇒ item (b) closes. Still zero ⇒ check the demand control (§1.2) before concluding
   anything at all.
5. Council/ownership: `5ae2147d`, `26186633`, `e05ea6f9` APPROVED + live; `642ecc3c` at round 3.
   Nothing else owed.

---

## 5. ⚠ ADDED LATER THE SAME DAY — round 3 is **APPROVED**, and its advisory objections corrected §1.1 and §2 above. **Read this section before acting on either.**

Run `53e3812f` → `complete_approved`, *"approved with 3 advisory objection(s) — none high-severity"*.
**The forward fix is approved AND live: there is nothing left to ship on it.** Full record:
`bugs_open/353` §13.

**Two of the three objections found real defects in my evidence, and one refutes a claim that had
stood in the bug file since filing. `APPROVED` is exactly the moment those stop being read — so:**

### 5.1 §1.1's probe was too narrow. Method wrong, conclusion survives.

`-l app=agent-chassis` matches **2** pods. The namespace holds **159**, of which **68** run this
binary as **per-run `agent-*` pods** — and `tool-generator` is one of those (it spawns per run; no
such pod exists at rest). "One replicaset" on a two-pod selector proves nothing about the pod that
executes the code.

**Re-checked and it holds:** 68/68 per-run pods on `v1.0.1332`, and a per-run pod **on a different
node** probes clean with both controls. **`-l app=<service>` answers "which pods carry this label",
never "which pods run this binary" — here they differ by 34×.**

### 5.2 ⚠ **"`tool-deployer` has never run" is FALSE** — §2's item (c) is withdrawn

`orchestration_states` retention is **24 h, sliding** (measured: oldest `COMPLETED` **24.5 h**). The
"0 rows" meant *"none in the last day"*. The "retention reaches 2026-07-19" figure was the
documented **false floor**.

**Re-sourced from `agent_error_log` (no reaper): 10 `tool-deployer` rows, 08-03 → 08-15**, each
*"workflow completed but its result could not be delivered to the parent"*. It ran.

> **(c) is REPLACED by (c′), which is a better question:** the re-emission path **ran at least 10
> times inside the damage window and the 30 withheld tools still have zero cross-links.** Did it run
> for other tools, or for these and withhold again? **[UNMEASURED]** — the discriminator is whether
> any of those 10 runs names a withheld tool. **This is the most interesting thing still open on 353.**

The damage figures are untouched — they came from durable rows, never from that table.

### 5.3 §2's item (d) has MOVED — it is now `bugs_open/379`, and it is UNOWNED

The regeneration residual (a `replace_existing` regeneration that ADDS a related page never emits
cross-links for it) was parked in 353's open list, where it would have closed silently when 353
closed. Filed separately at the council's direction. **Its size is `[UNMEASURED]` on purpose** — the
first task there is counting how many regenerations changed their `related_pages` set, not guessing.

### 5.4 Corrected open list for 353

- **(b)** new arm live but **UNEXERCISED** — first non-zero INFO row closes it (§1.2 still stands).
- **(c′)** `tool-deployer` ran ≥10× and the withheld tools still have nothing — **[UNMEASURED]**.
- ~~(d)~~ → `bugs_open/379`, unowned.

### 5.5 Trap 10, and it is the one worth carrying

**An ADVISORY objection in an APPROVED round is where the real defects were.** Two of three landed,
including the only thing that refuted a filed claim — raised by the one seat that *could not query
the table* and judged the claim's SHAPE instead. Read them before filing the verdict away.

**Trap 11: grepping `LANDMINES.md` for footprint A does not deliver entry B.** I had the file open
this session and still wrote the retention false floor. A TABLE-footprinted entry cannot reach you
through a hook that matches dirty FILE paths, and `kubectl exec … psql` touches no file — so the
delivery moment is the query itself. **Grep the table name before any `count(*)`, `min()`, or the
word "never".** (Third recorded instance; appended to the entry's own tally.)

---

## 6. FINAL UPDATE, 2026-08-24 after the `v1.0.1334` roll — **item (c′) is CLOSED. Only (b) remains, and it is a wait, not a task.**

### 6.1 Deploy state re-probed on the new build

`v1.0.1334` (chassis pods rolled 15:39Z; 58 of 61 per-run agent pods on 1334, 3 still on 1333).
Probed with both controls: fix arm **PRESENT**, control **+** present, control **−** absent.
**The 353 forward fix is live on the current build.**

### 6.2 ⚠ `tool-deployer` is NOT broken and NOT unexercised — §5.2's question is answered, and the answer removes work rather than adding it

**It ran, it reached the emitter, it emitted successfully 12 times, and it never once withheld.**
`site_work_items` where `item_key LIKE 'tool_crosslink:%'` by `source`: `tool-deployer` **12 items,
5 tools, 08-03 → 08-19**. `agent_error_log` has **no** crosslink skip row attributed to it at all.

**And one of those 5 tools is in the withheld set** — `tool-automation-savings-estimator`, 08-11.
That is one of the two "exceptions" `bugs_open/353` §3 had recorded since filing as coming from
"a later rebirth". Wrong attribution: **the path the bug file called dead had already repaired one
of the bug's own casualties, eight days before the bug was filed.**

**So there is nothing to fix in the deployer.** `deploy_tool_action.go` emits on both arms and its
early-return arm documents itself as the supported backfill route. What was missing was a TRIGGER —
nothing re-runs the deployer for a tool whose birth withheld — and both ends of that are already
closed (forward fix stops new withholds; the backfill repaired the historical 30).

**The 10 `agent_error_log` rows were never a deployer defect:** all are *"workflow completed but its
result could not be delivered to the parent"*, i.e. `bugs_closed/274`, a CLOSED fleet-wide class
(`page-rerender` alone has 5,869). ⚠ And because only DELIVERY-FAILED runs leave a row, **10 is a
floor, not a total** — do not quote it as how often the deployer has run.

### 6.3 Corrected open list for 353

- **(b)** new arm **live but still UNEXERCISED**. Re-checked post-roll: **4** births since the first
  roll, **all 4** `no_related_pages`, **0** cross-link items created today — Guard 2 has not been
  reached once. **A wait, not a task.** First non-zero `emitted_ungated_build_enqueued_by_caller`
  row closes it.
- ~~(c′)~~ **CLOSED** — `bugs_open/353` §14.
- ~~(d)~~ → `bugs_open/379` (unowned).

### 6.4 The one thing worth someone's morning

`no_related_pages` has **12** rows all-time since 07-31 and **4 are today** — a third of every
occurrence ever, on the same day tool births stopped reaching Guard 2 entirely. Pairs with §3.4's
tool-birth refusal rate. **[UNMEASURED]** as a rate (births/day not established), and still unfiled,
but it has outgrown "two points is not a trend".

### 6.5 Trap 12

**A path's own OUTPUT identifies its caller; its RUNS may not.** Three separate attempts to answer
"did `tool-deployer` do this?" via `orchestration_states` (24 h retention), log descriptions
(truncated at 150 chars, matching no component) and site/time overlap (7 of 8 sites overlapped —
suggestive and worthless) all failed or misled. The emitter stamps `source` and `emitted_by` on
everything it writes, and **one `GROUP BY source` answered in a single query what three joins could
not.** Before reconstructing who did something, check whether the artefact already says.

---

## 7. ADDED 2026-08-24 evening — §6.4 is ANSWERED, and the answer removes it from the open list while adding a better question underneath

**Read this before acting on §6.3 or §6.4: both change.**

### 7.1 The `no_related_pages` rise is migration 516 working, not a trend

§6.4 flagged "4 of 12 occurrences in one day" as having outgrown *two points is not a trend*. With a
denominator it is neither a trend nor a defect. Tool births come from `content_components`
(`component_level='tool'`, one row per tool per site), and for 08-22 → 08-24 the accounting closes
**exactly**: **13 births, 13 accounted** — 8 stopping at `no_related_pages`, 5 at Guard 2 — and
`tool-generator` has created **zero** `tool_crosslink:%` items since 08-21.

The boundary is **migration 516, applied 2026-08-21 16:55Z** (recorded in the lane NOTES, because the
runner refuses `--record-only` on a `_HOLD` sidecar). The first row of the new regime is 08-21
18:49Z. Pre-516 the whole-tree search substituted `suggestions[0]`'s pages into any spec that lacked
the key — `bugs_open/330`'s damage. **516 traded wrong cross-links for none, which is the right
trade and was the stated design.** Nothing to file on the count itself.

⚠ **Do not date the apply from `agent_definitions.updated_at`.** It read **08-24 15:38:27Z** on both
tool agents — three hours *after* today's occurrences, which would have refuted 516 as the cause. It
dates the last write to the row, not the write you are asking about; a different migration touched
both agents this afternoon. I made this mistake and the correct date was in two files I already had.

### 7.2 The real finding: `related_pages` has exactly ONE producer

[MEASURED 2026-08-24 ~17:1xZ] `add_tool` items since 08-17, by `source`: `tool-suggester` **11 of
11** carry the key; `owner-request` **0 of 58**, `automated_check` **0 of 7**, `operator` **0 of 1**.
A clean split by producer, not a compliance rate. The suggester's prompt asks for the field and its
one LLM reply since 08-20 returned it — the LLM half works. Every other route writes the spec by
hand from a five-key template (`complexity, description, function, name, priority`) that has never
carried it, with no default, no validation and no warning. Filed: `bugs_open/330` §12, and a `090`
diagnosis (intake `5dbead0b`, run `0b5695a4`).

### 7.3 ⚠ This corrects §6.3: item (b) is NOT "a wait"

§6.3 called 353's remaining item *"a wait, not a task"* — first non-zero
`emitted_ungated_build_enqueued_by_caller` row closes it. **On the current producer mix that row
cannot arrive.** Every birth since the forward fix rolled is `owner-request`, so it returns at
`no_related_pages` *before* Guard 2 is reached; the new arm is not merely unexercised but
**unreachable**. Waiting is not a strategy that terminates.

**What would exercise it:** a `tool-suggester`-sourced `add_tool` item (so the spec carries
`related_pages`) whose tool page is not yet live and has no gate item. dartsonline's five births on
08-23 were exactly that shape — one day before the fix rolled. So either wait for the next suggester
run on a fresh site, or hand-author one `add_tool` spec **with** `related_pages` and let it build.

### 7.4 Corrected open list for 353

- **(b)** live, unexercised, and **unreachable on the current producer mix** (§7.3) — needs a
  suggester-sourced birth, not patience.
- ~~(c′)~~ CLOSED (§6.2).
- ~~(d)~~ → `bugs_open/379`, unowned.
- **NEW:** the producer split — `bugs_open/330` §12, `090` run `0b5695a4` pending.

### 7.5 Trap 13

**A denominator is what turns a count into a reading, and the good ones close.** "4 today out of 12
all-time" and "13 of 13 births emitted nothing" are the same fleet-day described with and without
the denominator, and they point at opposite conclusions — the first at a spike to chase, the second
at a design change working as intended plus a producer nobody had audited. The test that the
denominator is sound is that the accounting leaves **no remainder**: births = skips + emissions. A
proxy that leaves one is measuring something else.

---

## 8. CONSUMER NOTICE FROM THE `bugs_open/333` LANE (received 2026-08-24 evening) — the 353 acceptance census gains a THIRD bucket after the next roll

Recorded here because it arrived in chat and would otherwise be lost. **No action owed by this
lane**, but it changes how the backfill's numbers must be read.

The 333 lane (owned-page routing) reports that this lane's `backfill-353` run on 08-23 filed **8**
cross-link items at `page-build-handler` on pages whose `rebuild_policy` is `'owned'` (17:15–17:17Z).
`page-build-handler` is forbidden to touch an owned page, so today it refuses, the item terminates
`wont_fix`, and the detector re-files it later. That is the mechanical half of the caveat
`bugs_open/353` already carries at its lines 128 and 213.

**Their change (`6ab0b3434`, INERT until the next chassis roll):** `emitToolCrossLinkItems` now goes
through `withWorkItemTx` → `writeWorkItem`, so it sits inside the owned-page door's coverage. From
the roll, a cross-link finding onto an owned page is **parked** rather than routed —
`status='deferred'`, no handler, `error` leading `OWNED_PAGE_GUARD`, keeping its own `item_type` and
`item_key`. Nothing about the emitter, the 029 URL guard or the 353 gate changed.

⚠ **The trap for whoever re-runs the acceptance census:** "74 created, N completed" becomes three
buckets, and a ratio computed as `complete / (complete + failed)` will silently exclude the deferred
ones and **read higher than the work actually done**.

```sql
SELECT status, count(*) FROM site_work_items
 WHERE item_key LIKE 'tool_crosslink:%' GROUP BY 1;   -- expect a 'deferred' bucket post-roll
```

Whether an owned tool page *should* receive a cross-link at all is an open design question. 333
deliberately does not take it; it stops the answer being "silently refused and forgotten". If this
lane picks it up, that is where it belongs — not in 333.

---

## 9. ⚠ ADDED SAME EVENING — the `090` in §7.2 returned **NO VERDICT**. §7.2 stands on first-hand verification, and says so.

Run `0b5695a4-a09d-4895-b4ab-652ce88a991a` (intake `5dbead0b`) went `COMPLETED`, work item
`complete`, **4 bundles, no `iteration_note`, no verdict, no `doc_note`.** It neither confirmed nor
refuted the producer split. **Do not quote "090 filed" as "090 confirmed"** — the 2026-07-31 owner
ruling allows declared first-hand verification instead, and `bugs_open/330` §12's correction block is
that declaration.

⚠ **It is NOT the truncation case `LANDMINES.md` tells you to predict, so do not re-file narrower.**
The target file is **32,928 bytes** (under the ~60,000 bar the entry says to check with `wc -c`), and
iterations 3 and 4 rendered **12 of 12** in-scope symbols with `truncated: false`. Only iteration 2
hit the budget (840 chars omitted at 59,882 of 60,000). Iterations 3 and 4 are identical in
`body_chars` and `symbol_count` — the loop re-requested the same scope and stopped: **an iteration cap
without convergence**, a second and previously undocumented route to the same silence. Addendum
appended to that landmine entry.

**Trap 14: `orchestration_states` COMPLETED plus bundles and no verdict is a KNOWN silent failure of
the diagnosis loop, and its documented predictor (`wc -c` the target file) does not cover all of it.**
Check `metadata->>'truncated'` per iteration before concluding the budget was the cause — all-`false`
with no verdict means narrowing the scope is wasted effort.

### 9.1 Also owed, and unfixable forward-only: the 353 code commits carry NO council trailer

The round-3 verdict for corr `642ecc3c` (run `53e3812f`) was **APPROVED** (§5), but none of
`323b63a00`, `8ae7aceae`, `027461e3d` carries `Council-Reviewed:` or `Council-Submitted:`, so the
`098` coverage report lists `027461e3d` under **UNREVIEWED** (checked 2026-08-24, 2-day window).
**That is a false negative caused by a missing trailer, not a missing review.** Forward-only forbids
an amend, and writing the trailer onto some unrelated later commit would be a false claim — so the
record is here and in `bugs_open/353` §13. **Do not resubmit `642ecc3c`**; the verdict exists and is
approved.

---

## 10. OWNER RULING 2026-08-24 (evening) — the picker is BUILT. One half live-inert, one HELD. **This is the live work; read it before §7.**

Owner's answer to §7.2's open question: *"Do the third and the second as a stop gap please."*

### 10.1 State of each half

| half | commit | state | what it needs next |
|---|---|---|---|
| stopgap: `related_pages` in the hand-order recipe | `d5dafd6a7` | **LIVE** (a doc) | nothing — the owning lane has it |
| reader: `related_pages_fallback` on both tool actions | `0fb94a7dd` | **live-inert** once a build ships it | a chassis roll |
| wire: migration 602 | `c64bbbd03` | **HELD** | apply AFTER the reader is live |

Council `c962abd1-87e4-473f-9990-3985322050af`, submitted 19:1xZ, verdict not yet read.

### 10.2 The apply is NOT "after the next roll" — it is after the reader is PROVEN live

Full procedure, with the binary probe and both controls:
`RUNBOOK_staged_component_build.md`, "Applying migration 602". ⚠ Two traps it carries:
`-l app=agent-chassis` matches **2** pods of the ~68 running this binary, and `tool-generator` is a
**per-run** pod, not one of the two; and there will be **no `schema_migrations` row** for 602, because
the runner refuses `--record-only` on a `_HOLD` sidecar — so record the apply in the lane NOTES, and
do not later read that absence as "it never ran". (That is exactly the trap that made me date 516 from
`agent_definitions.updated_at`; see §7.1.)

### 10.3 The verification is a DEMAND problem, and the demand has been offered

The apply verifies itself in-transaction. What it cannot verify is that the picker ever supplies
pages. The discriminating case is a hand-filed `add_tool` **without** the key from a real producer —
**the `webdesign-tool-rebuilds` lane has offered to supply one on request.** Take them up on it.

**PASS = at least one row carrying `related_pages_source='suggested'`**, in either
`agent_error_log.context` or the emitted items' `spec`. "Cross-mentions resumed" is not the test, and
an empty `suggested` bucket beside a non-empty `spec` one means the picker was never REACHED (the
requester named pages and won), not that it failed.

### 10.4 This also gives 353's item (b) its route

§7.3 established that item (b) is unreachable on the current producer mix, because every birth stops
at `no_related_pages` before Guard 2. **Once 602 is applied, hand-ordered births will carry pages
again and will therefore REACH Guard 2** — so the first non-zero
`emitted_ungated_build_enqueued_by_caller` row becomes reachable as a side effect. It is still a wait,
but it is now a wait on something that can actually happen.

### 10.5 Trap 15, learned by doing it wrong today

**On a two-commit change, SUBMIT TO THE COUNCIL BEFORE THE FIRST COMMIT.** I committed the reader,
then submitted, then committed the migration — so only the migration carries `Council-Submitted:`, and
`0fb94a7dd` will list as unreviewed for ever even when this round approves. Forward-only forbids an
amend, and a trailer on a later unrelated commit would be a false claim. The submission asserts
nothing and costs nothing until the verdict, so there is no reason to commit first. Second trailer gap
this lane has recorded today (§9.1 is the other).
