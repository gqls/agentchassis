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
