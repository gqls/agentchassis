# HANDOFF — 2026-08-25, fresh chat starts here: **the lane's deliverable was done a week ago. Everything since is a bug tail, and today that tail went LIVE. What remains is three items, TWO of which nobody can act on — they are waits on demand.**

**Supersedes `HANDOFF_2026-08-24b_continue_here.md`** (and transitively 08-24 / 08-22 / 08-21).
**DOES NOT supersede `HANDOFF_2026-08-20_continue_here.md`** — that file still holds the RFC_029 gate's
terms, baseline, interim reads and the CLOSE-OUT (§2.11). Read it for the evidence trail; do not
re-derive it.

**If you are deciding whether to keep this lane open, read §7 first.**

---

## 1. STATE AT A GLANCE, 2026-08-25 ~10:0xZ

| thing | state | who acts next |
|---|---|---|
| RFC_029 D2 Phase 2 (the lane's actual deliverable) | **CLOSED** 08-24 | nobody |
| `bugs_open/353` forward fix | **LIVE + council-APPROVED** | nobody |
| `bugs_open/353` item (b) — new arm never exercised | **live, still 0** | *waits on a tool birth* |
| `bugs_open/330` §12 producer split → the picker | **LIVE as of today** | *waits on the demand case* |
| migration 602 (the picker) | **APPLIED 2026-08-25 ~09:5xZ** | nobody |
| picker-failure vs honest-"none" (council residual) | **OPEN, ~6 lines, mechanism verified** | **the only real task here — §4.3** |
| `bugs_open/379` (regeneration never emits) | **UNOWNED**, out of scope by design | somebody else |
| ~30-tool backfill | **UNOWNED** | somebody else |

---

## 2. WHAT HAPPENED TODAY

### 2.1 The reader is live — on BOTH tags in flight, probed, not inferred

Chassis rolled to **`v1.0.1337`** at 09:27Z. ⚠ **Two tags were in flight**: [MEASURED 2026-08-25]
**82** pods on `v1.0.1336`, **16** on `v1.0.1337`, **98** running this binary in total. A
`tool-generator` spawn could land on either, so both were probed, each with two controls:

| pod | tag | node | `related_pages_fallback` | ctl **+** | ctl **−** |
|---|---|---|---|---|---|
| `agent-chassis-67fd9c76f5-2g8kw` | 1337 | `…31336` | **present** | present | absent |
| `agent-build-dispatch-loop-4ccba5fc-nbnsh` | 1336 | `…36833` | **present** | present | absent |

The second is a **per-run pod on a different node** — the §5.1 requirement from the last handoff,
because `-l app=agent-chassis` matches **2** pods of the **98** running this binary.

### 2.2 Migration 602 is APPLIED, and this file is part of its ledger

Applied by hand ~09:5xZ with `ON_ERROR_STOP=1`. `COMMIT`, verify NOTICE fired. **There is
deliberately NO `schema_migrations` row** — the runner refuses `--record-only` on a `_HOLD` sidecar.
**Do not read that absence as "it never ran".** The record is here and in the lane NOTES
(`## 2026-08-25 (~09:5xZ)`).

Read back **independently** after the transaction closed:

```
type            anchor_next            picker_next   fallback_wire
tool-deployer   load_site_page_names   deploy_tool   suggest_related_pages.result
tool-generator  load_site_page_names   save_tool     suggest_related_pages.result
```
Carrier census, widened to the new key: **0 unmarked** carriers on every agent;
`tool-generator`/`tool-deployer` each **1 marked `related_pages?` + 1 marked
`related_pages_fallback?`**; `tool-recreation-handler` and `tool-suggester` read 0/0, confirming they
match only in prose.

### 2.3 What the picker actually is, in one paragraph

A tool built for a site is meant to gain **cross-mentions**: one sentence woven into a related page's
existing prose, mentioning the tool and linking to it (live example, `dartsonline.com`
`/blog/barrel-shapes.html`: *"…the tungsten percentage vs barrel diameter visualiser lets you compare
percentages against weight…"*). Which pages get one comes from `related_pages` on the `add_tool` item.
[MEASURED 2026-08-24] that field had **exactly one producer** — `tool-suggester` 11 of 11; every other
route **0 of 66** — because every hand-written spec came from a five-key template that never had it.
Migration 516 had removed the whole-tree search that used to mask this by substituting another tool's
list (`bugs_open/330`), so the omission had become **no cross-mentions at all**: 13 of 13 births
08-22→08-24 emitted zero. The owner ruled the system should ASK. 602 is the asking; `0fb94a7dd` is the
reader that accepts the answer, **only when the requester named nothing**.

---

## 3. THE COMMITS, so you do not have to reconstruct them

| commit | what |
|---|---|
| `3eecb0cd8` | `bugs_open/330` §12 — the producer split measured (11/11 vs 0/66) |
| `f656d6784` | LANDMINE: a hand-written `add_tool` spec builds a perfect tool with no cross-links |
| `4df5f4c63` | WRONG_CALLS: I dated migration 516 from `agent_definitions.updated_at` |
| `4680852ae` | the `090` returned NO VERDICT — recorded in all four places, plus the second route to that silence |
| `d5dafd6a7` | **stopgap**: the hand-order recipe gains `related_pages` |
| `0fb94a7dd` | **reader**: `related_pages_fallback` + the `relatedPagesSource` stamp |
| `c64bbbd03` | **wire**: migration 602 (HELD) + the council submission |
| `4432518b8` | lane docs: apply procedure, demand-control test, the trailer mistake |
| `a6ec50691` | council APPROVED r1 — all four advisory objections acted on (`Council-Reviewed: c962abd1-…`) |
| `22712fa31` | the demand-case protocol, recorded on both lanes' sides |
| `39ee790d6` | 602 applied; the exit-137 probe trap |
| `75e66aeb8` | LANDMINE addendum: the positive control cannot save the negative one |

---

## 4. WHAT IS LEFT — three items, in the order they should be picked up

### 4.1 ⏳ The demand case (waits on another lane — **do not chase it**)

The `webdesign_tool_rebuilds` lane has been **pinged** (the ping went out after 602 applied) and the
protocol is committed on both sides (`22712fa31`, and their own NOTES): **a real next-in-queue filing
from their recipe with `related_pages` deliberately omitted**, on `webdesign.co.uk` (37 non-tool
candidate pages as of 08-24). No synthetic spend — the filing was happening anyway, which is what
makes it a demand control rather than a staged one.

**PASS = at least one row carrying `related_pages_source='suggested'`:**
```sql
SELECT context->>'related_pages_source' AS src, context->>'related_pages_n' AS n, count(*)
  FROM agent_error_log
 WHERE error_code LIKE 'tool_crosslink_not_emitted:%' AND occurred_at > '2026-08-25 09:50Z'
 GROUP BY 1,2;
SELECT spec->>'related_pages_source' AS src, count(*) FROM site_work_items
 WHERE item_key LIKE 'tool_crosslink:%' AND created_at > '2026-08-25 09:50Z' GROUP BY 1;
```
⚠ **"Cross-mentions resumed" is NOT the test.** It cannot distinguish a working picker from a week in
which filers happened to fill the field in. The stamp exists for exactly this.

**Baseline as of 10:0xZ today, so you read the result against the right zero:** tool births today
**0**; `llm_call_log` where `step_name='suggest_related_pages'` **0 rows**. *Those are the same zero* —
nothing has asked, so nothing could have answered.

### 4.2 ⏳ `bugs_open/353` item (b) (waits on the same demand, and 602 unblocked it)

The new ungated-emit arm is live and has **never fired** — `emitted_ungated_build_enqueued_by_caller`
still **0**. Until today it was **unreachable**: every birth returned at `no_related_pages` *before*
Guard 2. **602 changes that** — a birth that now carries picked pages reaches Guard 2, so the arm
becomes reachable as a side effect of §4.1. Still a wait, but now a wait on something that can happen.

### 4.3 ⭐ **THE ONLY REAL TASK: a failing picker is indistinguishable from an honest "none"**

This is the council's one substantive objection (`bug_historian`, medium, round 1 of `c962abd1`) and it
is **correct**. Both new steps route `error_step` **to** the saving step — right, because a
cross-mention must never fail a tool build — but the emitter then records `no_related_pages` with an
empty source, which is **exactly what a picker that ran and honestly answered "none" produces**. The
stamp separates *spec* from *suggested*; it does **not** separate *declined* from *never ran*.

**The fix is ~6 lines and I have VERIFIED the mechanism exists** (`datahelpers.ActionInputs.Has`,
`action_inputs.go:554` — `ok && v != nil`, i.e. present-and-non-nil, which is precisely the
distinction). When the picker runs and answers `[]`, the key resolves to a present-but-empty value;
when the step errored to `error_step`, the key is **absent**. So in `relatedPagesFromInputs`
(`create_tool_cross_link_items.go`):

```go
if pages := relatedPagesFromSpec(inputs.GetRaw("related_pages_fallback")); len(pages) > 0 {
    return pages, relatedPagesSourceSuggested
}
if inputs.Has("related_pages_fallback") {
    return nil, relatedPagesSourcePickerDeclined // it RAN and named none — a real answer
}
return nil, ""                                   // nobody was asked, or the step died
```
plus a fourth constant and a fourth test case in `TestRelatedPagesPrecedenceAndSource` (the harness is
already there and already mutation-proved).

⚠ **Do not treat this as free.** The six lines are the small part: it is `platform/` code, so it needs
a council round, a build, a fleet roll and a re-probe — the same multi-hour cycle §2 just went through.
**That cost, not the difficulty, is the reason it is not done.** Until it is, the interim check is at
the picker's own surface, and it is written up in `RUNBOOK_staged_component_build.md` ("602 follow-up:
did the picker actually RUN?"):
```sql
SELECT date_trunc('day', created_at)::date, success, count(*) FROM llm_call_log
 WHERE step_name='suggest_related_pages' GROUP BY 1,2 ORDER BY 1;
```
⚠ That query **cannot see a failure of `load_site_page_names`**, which runs before any model call —
its signature is *skips with no picker calls at all*.

---

## 5. CLOSED — do not reopen without new evidence

- **RFC_029 D2 Phase 2** — the lane's deliverable. `HANDOFF_2026-08-20_continue_here.md` §2.11.
- **`353` (a)** the forward fix; **(c′)** the `tool-deployer` question — it ran, reached the emitter,
  emitted 12 times, never withheld, and had already repaired one of this bug's own casualties eight
  days before it was filed; **the damage half** (74 items, 61 complete, 12/12 sampled pages serving).
- **`330`'s resolver halves** — absence honoured and presence delivered, both proven.
- ⚠ **DO NOT RUN THE 51-PAGE REDEPLOY.** An earlier section asked for it on a premise since retracted:
  page URLs were **constructed** instead of read from `pages.url`, so every zero — control included —
  was a miss on a URL that does not exist. `bugs_open/353` §11.
- **Do not resubmit `642ecc3c`** (353's round 3) — APPROVED. Its three code commits carry no trailer
  and never can (forward-only); `098` lists `027461e3d` as unreviewed and **that is a false negative
  from a missing trailer, not a missing review**. §9.1 of the 08-24b handoff.

---

## 6. TRAPS EARNED (1–12 are in `HANDOFF_2026-08-24b`; these are new)

13. **A denominator turns a count into a reading, and the good ones CLOSE.** "4 today of 12 all-time"
    and "13 of 13 births emitted nothing" are the same fleet-day with and without a denominator, and
    they point at opposite conclusions. The test that a denominator is sound is that the accounting
    leaves **no remainder**.
14. **`orchestration_states` COMPLETED + bundles + no verdict is a known silent failure of the `090`
    loop, and its documented predictor (`wc -c` the target file) does not cover all of it.** Ours was
    32,928 bytes — under the bar — and iterations 3–4 rendered 12 of 12 symbols at `truncated:false`.
    Check `metadata->>'truncated'` per iteration; all-`false` with no verdict means narrowing scope is
    wasted effort.
15. **On a two-commit change, SUBMIT TO THE COUNCIL BEFORE THE FIRST COMMIT.** Submitting between the
    halves stranded `0fb94a7dd` with no trailer it can ever carry. The submission asserts nothing and
    costs nothing until the verdict.
16. **The SKETCH is the submission for anything the reviewer cannot open.** I truncated 602's sketch at
    900 characters, mid-`jsonb_set`, before the second `UPDATE` — and **three of the round's four
    objections were about what the truncation hid**, all false about the file and all correct about
    the submission. Ask of every sketch: *is the line that makes this work inside the excerpt?*
17. **The positive control cannot save the negative one, because `grep -q` makes them asymmetric.** A
    present literal exits early and returns in seconds; an absent one must scan the whole binary. So
    the negative control is the slow one — and under a deadline it returns **exit 137**, which every
    convenient shape (`[ $? -ne 0 ]`, `|| echo ABSENT`) reports as a clean absence. Print the raw code;
    treat anything but 0 or 1 as NO RESULT. And a per-run pod can complete between selection and use.
18. **An advisory objection in an APPROVED round is where the real defects are** (carried from trap 10,
    and it paid again today): of four advisories, one exposed a genuine design gap (§4.3) and one sent
    me to a live mechanism I had not found (`internal-linker`, now a landmine).

---

## 7. CAN THIS LANE CLOSE? — **yes, and here is the honest accounting**

**The lane's own deliverable closed on 08-24.** Everything in this file is a bug tail that grew out of
it. Of the three items left:

- §4.1 and §4.2 are **not tasks**. Nobody can make them happen; they resolve when another lane files a
  tool build, and the queries to read them are written down above. Keeping a workstream open to hold
  two queries is what produces the shelf of near-identical handoffs this repo's own rules warn about.
- §4.3 **is** a task, it is small, and its mechanism is verified — but its cost is a council round plus
  a fleet roll, not six lines.

**Recommendation: close the lane, and route the three items to homes that will actually be read** —
§4.1's queries into `bugs_open/330` (whose fix they verify), §4.2 into `bugs_open/353` §12.6 (already
its authority), and §4.3 into a `bugs_open/` file of its own so it is owned rather than parked in a
lane nobody re-opens. That last one matters: **a residual left in a closing lane's handoff closes
silently with it**, which is exactly what happened to 353's item (d) until the council made us file it
as `bugs_open/379`.

**Do not close it by declaring the residual done.** §4.3 is real, it is stated on the record in an
APPROVED council round, and the interim check is a workaround, not a fix.

---

## 8. SESSION-START CHECKLIST

1. **Do not re-derive the deploy state from this file** — re-probe (§2.1's three-row method, and read
   trap 17 first). `kubectl -n ai-persona-system get pods -l app=agent-chassis` sees **2** pods of ~98.
2. **602 is applied.** There is no `schema_migrations` row and there never will be. Confirm at the
   config, not the ledger: the read-back query in §2.2.
3. **Has the demand case landed?** The two queries in §4.1, and check the baseline zeros first.
4. **Has anything asked the picker yet?**
   `SELECT count(*), max(created_at) FROM llm_call_log WHERE step_name='suggest_related_pages';`
5. Council: `c962abd1` APPROVED + trailered. `642ecc3c` APPROVED, untrailerable. Nothing owed.
