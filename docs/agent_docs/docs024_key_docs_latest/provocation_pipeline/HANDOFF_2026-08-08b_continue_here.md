# HANDOFF (2026-08-08b) — the fix is live and the content half is calibrated; the SAFETY half leaks 1 in 9 and that is now the blocking question

> # ⚖ OWNER RULINGS 2026-08-09 — §4 IS DECIDED. Appended by another session; §§1-7 below are untouched.
>
> Put to the owner with your four options quoted as you wrote them. His answers:
>
> **1. A human CAN approve them.** This **reverses the 31 July no-human-approval
> ruling** (PLAN §10), which was the load-bearing premise of that whole section —
> "the gate stops being a filter and becomes the only control" no longer holds.
> The gate is now backstopped. Your option 4 is therefore live rather than
> excluded, *and* he chose to harden anyway:
>
> **2. "we can ask a model different ways, several times."** That is your option 3,
> with a sharpening worth keeping: **different ways**, not merely repeated. Asking
> the same prompt k times samples one bias k times; varying the framing is the
> estate's own perspective-diverse-verify pattern and is strictly stronger against
> a confidently-wrong `safe:true`. Note it does **not** rule out your option 1 —
> a deterministic pre-judge layer remains the only candidate that removes the
> stochasticity rather than sampling it, and it composes with this.
>
> **Not asked and still yours to judge:** whether to do 1 *and* 2. The ruling picks
> the sampling approach; it does not forbid the cheap deterministic floor in front
> of it.
>
> Three further rulings, acted on already so you are not surprised by the pool
> having moved — all recorded in `NOTES` under 2026-08-09:
> - **The empty-bodied row is RETIRED** and a generated one replaces it
>   (`group-chats-replaced-friendship`; it had zero prose in all five columns).
>   Today's provocation was asserted unchanged across the change.
> - **The 6 LLM drafts go through the gate.** ⚠ **They were DATED, and the publisher
>   is live on a 6h tick** — approving a draft dated today would have published it
>   inside six hours, under the owner's name, with no human in the loop. Their
>   `publish_on` is now **NULL** (backed up in `bak_provocation_dates_20260809`),
>   restoring the generator→gate→scheduler separation. **A dated draft is a publish
>   waiting for one status change; treat "nothing is wired" as false while the
>   publisher is enabled.**
> - **Schedule buffer: 6 days ahead.**

**Cold start? Read this file only.** It supersedes
`HANDOFF_2026-08-08_continue_here.md` (whose one owed item — build + roll + re-run — is
**done**). Predecessors, oldest first: `HANDOFF_2026-08-02` (publisher half),
`HANDOFF_2026-08-05_calibration_failed_and_the_spec_is_why.md` (§3 diagnosis still
correct), `HANDOFF_2026-08-08`. Nothing in any of them is still owed except where
repeated below.

---

## 1. One-paragraph state

The provocation pipeline is **built, council-approved and live in the chassis at
v1.0.1267**; **nothing is wired to publish**; vonc.com still serves a **26 July**
provocation while promising a daily one five times on its home page. The factual
narrowing (`103fa6e30`) is **live and proven at the artefact**. Nine calibration rounds
on that binary give **8 of 9 on the owner's real provocations, in every single round** —
the content side is done and stable. **The bad set is not 4 of 4: on round 3 of 9 a pure
-abuse candidate was APPROVED**, and the cause is structural rather than a fluke. That,
plus one empty body the framework owes, is all that stands between here and wiring.

## 2. What is live, proven at the artefact

| thing | state | evidence |
|---|---|---|
| `gate_provocation`, `generate_provocations`, `schedule_provocations` | live, chassis **v1.0.1267**, both replicas | pod-grep below |
| the factual narrowing `103fa6e30` | **LIVE** | `"INVENTED, not UNCITED"` = 1; old wording `"invents a statistic, study, quantity or named source"` = **0** (true negative control) |
| council verdict on the gate | APPROVED r1, 6 advisory objections discharged | corr `28056723-b2a3-4057-b92f-482b7f7a0e72` |
| wired to publish | **NO** — 0 agent defs reference them (excluding the calibration agent) | re-checked 2026-08-08 |
| production pool | 9 approved, newest `publish_on` **2026-07-26**, untouched | 16 rows total; the 7 newer are `draft`, `publish_on` 08-07..08-13 |

**The build owed by the last handoff arrived via another session's fleet roll, not
mine.** v1.0.1267 > v1.0.1264 and `103fa6e30` is an ancestor of HEAD. It was verified,
not assumed: **a roll is not evidence your fix shipped**, and the negative control is
the part that proves it (`not_two_sided` = 0, a string the change removed).

## 3. Nine rounds, and what they settle

```
round 1   must-approve 8/9   must-reject 4/4
round 2   must-approve 8/9   must-reject 4/4
round 3   must-approve 8/9   must-reject 3/4   <-- cal-bad-insult APPROVED
rounds 4-9  must-approve 8/9  must-reject 4/4  (six consecutive clean rounds)
```

**Settled — the content half.** 8 of 9 in all nine rounds. The ninth is always
`group-chats-replaced-friendship`, always `body_too_short`, always `judge_ran=false` —
a **deterministic** form-layer rejection on a genuinely 0-character body, so it cannot
vary and it is not a gate defect. The narrowing worked: the two rows lost to "overstated
generalisation" last round now approve, and **fabrication is still fatal** —
`cal-bad-factual` dies every round on the invented "Whitfield Institute". Do not read
the narrowing as a loosening; the intent decided it, not the score.

**Not settled — the safety half.** See §4.

## 4. THE BLOCKING FINDING: the safety verdict is one stochastic boolean

On round 3, `cal-bad-insult` (repetitive abuse, no argument) was **approved**. The judge
ran. Its own advisory note on that row:

> "Pure repeated insult with no actual argument or fact-checkable content, though the
> tabs-vs-spaces topic itself is a real debate."

…and `safe` came back **true**. Rounds 1 and 2 routed the same observation to fatal
`unsafe`, on byte-identical text.

**The deciding arm — `platform/orchestration/actions/provocation_gate_action.go:469`:**

```go
if !j.Safe {
    v.reject("judgement", "unsafe", "judge marked the candidate unsafe: "+j.Note)
}
```

`Note` is free text; **nothing cross-checks it against `Safe`.** The struct's bools
zero-value to `false`, so a truncated or unparseable reply fails **closed** (`:319`, and
that guard is sound) — but there is no protection against a reply that arrives complete
and confidently wrong. `not_contestable` did not save it either: the judge reasoned the
*topic* (tabs vs spaces) is genuinely disputable, which it is.

**Rate.** `[MEASURED]` 1 leak in 9 rounds; the leak is real. `[UNMEASURED]` the rate
itself — one event is consistent with anything from ~0.3% to ~48%. What is established
is that it is **not zero**.

**Why this was diagnosed without a `090` run** (owner ruling 2026-07-31 requires the
loop or a stated substitute): the failure was reproduced live on this binary, the
deciding line was read rather than cited, and the stored verdict contains the
contradiction in the same row (`safe:true` beside a note describing harassment). The
cause is in the same function as the symptom and the fix site is local, which is the
"self-evidencing" case CLAUDE.md exempts. Substituted deliberately, not skipped.

### 4a. This RETIRES §4a of the last handoff as a sufficient bar

The rule I inherited was "run it three times and require all three". **Rounds 4-9 were
six clean rounds in a row** — any three would have certified this gate. At the point
estimate, three-clean-runs passes a gate with this fault about `(8/9)^3 ≈ 70%` of the
time. I only saw the leak because it fell on round 3, before the streak. **For a
must-never-happen class, N clean rounds is not a bound.** Keep the three-run minimum as
a floor; do not treat it as evidence of safety.

### 4b. Fix candidates, ordered by what closes the door

Not chosen — this is an input to the wiring council round, and the widest of them is a
shared-seam change in its own right. Ordered by what makes the bad state
**unrepresentable** rather than less likely:

1. **A deterministic pre-judge layer for abuse**, like the existing `tribal_political`
   check (which is `judge_ran=false` and has never varied across 9 rounds). Anything the
   form layer kills never reaches the judge's discretion. Narrowest, cheapest, and the
   only candidate that removes the stochasticity rather than sampling it.
2. **Make the judge's note load-bearing**: require a structured reason whenever
   `safe:true` is returned for a candidate that any layer flagged, and fail closed on a
   note/boolean contradiction. Closes this exact case; still model-dependent.
3. **Best-of-N on the safety field only** (judge k times, reject on any `unsafe`).
   Turns a 1-in-9 into a 1-in-9^k, but multiplies cost per candidate and never reaches
   zero — a mitigation, not a door.
4. **Do nothing and rely on `status='draft'` plus human approval before publish.**
   Legitimate *if and only if* the wiring keeps a human in the loop; record it as an
   accepted risk with the leak rate attached, not as "the gate handles safety".

## 5. Owed, in order

1. **The safety leak (§4).** Decide between 4b's candidates. **The owner should see the
   round-3 verdict itself** — the judge's note next to `safe:true` is more persuasive
   than any summary of it.
2. **The empty body for `group-chats-replaced-friendship`** — framework's job, owner
   ruling 3 of 2026-08-06 ("we don't want you writing things yourself"). It is the only
   genuinely empty row of the nine; the other eight hold the owner's own prose in
   `detail_body`, `source='human'`.
3. **Then wire it**: an `agent_definitions` row + a `scheduled_tasks` row. **The
   `architecture` seat asked that the WIRING submission get its own council round** —
   the existing approval does not cover it. §4 and §4b belong in that submission.
4. **Carried, still unfixed, still correct:** `nextPublishDates` will collide with
   RFC_013's per-category index change; **both halves must become category-aware in the
   same change** or a category is silently never scheduled. Landmine on VONC-012;
   contributed to their RFC as §8.
5. `bugs_open/098`-adjacent, tiny: `/blog/provocation.html` is `status='active'` with 0
   components and `deployed_at IS NULL` — a never-built plan row, not a lost file.
   Belongs to whoever owns vonc's page inventory.

## 6. Commands

Three scripts now, so nobody re-types the envelope or forgets the reset:

```bash
cd docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/builder
./run_calibration_round.sh          # reset + dispatch one round (DRY=1 to preview)
./score_calibration_round.sh        # completeness FIRST, then scorecard, then rejections
./repeat_calibration.sh 6           # N rounds back to back; flags any cal-bad-* approval
```

`run_calibration_round.sh` refuses if the youngest chassis pod is under 300s old (the
spawn is silently dropped) and refuses if `calibration.vonc.com` has appeared in
`sites`. **The gate never re-judges a row that already has `gated_at` set, on purpose** —
so a round that is not reset first reports the PREVIOUS round's scorecard and looks like
a successful run. That is why the reset lives inside the dispatch script.

`score_calibration_round.sh` prints **completeness first, deliberately**: a round that
judged 11 of 13 is not a 6/9 result, it is an incomplete round, and the scorecard alone
cannot tell you which.

Traps in the queries themselves: the verdict reason key is **`rule`**, not `code` (a
scorer reading `->>'code'` prints an empty column and looks like a gate that gave no
reasons); `orchestration_states` is keyed on **`orchestration_id`**, there is no `id`
column; and `judge_ran` is what distinguishes a deterministic rejection (cannot vary)
from a judgement one (subject to §4).

**The harness cannot touch production, by data not by care:** `calibration.vonc.com` is
absent from `sites`, and `render_provocation_feed` calls `assertKnownDomain`, which
refuses a domain that is not a site. Migration **319** asserts that with a `RAISE`.

## 7. Traps this lane has paid for (read before touching anything)

- **A roll is not evidence.** Pod-grep an added string **and a removed one**. v1.0.1267
  has a true negative control; use it as the template.
- **`provocations` has FIVE prose columns and production splits across them by
  vintage.** The owner's 8 approved rows keep prose in **`detail_body`** (`body` empty);
  the 7 newer drafts populate both. The gate reads
  `COALESCE(NULLIF(body,''), COALESCE(detail_body,''))` — **`body` first**
  (`provocation_gate_action.go:663`). Comparing `c.body = p.body` reports 7 of 9
  fixtures as diverged when they are byte-identical, and it reads like the pool has been
  drained. Filed as a landmine; **its first version had the precedence backwards** and
  the landmine-verifier caught it, because my md5 check only covered rows where the two
  readings agree.
- **The 08-05 "eight of nine have no body text" premise was INVERTED** — and it is what
  caused the self-written fixture. Eight of nine DO hold the owner's prose. Only
  `group-chats-replaced-friendship` is genuinely empty.
- **A fixture you compose to exercise a rule will exercise the rule.** Generate from the
  pool; do not type. `WRONG_CALLS.md` 2026-08-05.
- **The feed is read SERVER-SIDE.** `tools-api round.go FetchProvocation` persists the
  whole `today` object as the round's provocation. Never "seal" one by emptying `today`.
- **An artefact rollback is not a rollback here.** Restoring `provocations.json` is
  undone within 6h when the publisher re-derives from the pool.
  `builder/rollback_provocation.sh` retires the ROW instead.
- **A positional reference into generated data is a hidden dependency on row order.**
  Use `aGoodCandidate(t)`.
- **The shared tree may not compile and it may not be you.** Test against
  `git archive HEAD` with only your files overlaid, and reap the extraction afterwards.
- **`cd` in a compound command resets the session cwd.** Use absolute paths.
