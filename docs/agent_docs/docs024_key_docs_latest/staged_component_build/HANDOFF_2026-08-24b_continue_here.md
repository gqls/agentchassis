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
