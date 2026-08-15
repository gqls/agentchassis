# HANDOFF 2026-08-15 — bugfix 213, continue here (supersedes 08-14b on STATE)

**This is the cold start.** Read this, then `SUMMARY_2026-08-15…md` for the read-aloud version.
The 08-14b handoff is superseded on state; its §"RECONCILIATION" (how 213/216/274 relate) still
stands and is the reference for that question.

---

> ## ⬆ UPDATE 2026-08-15 ~10:40Z — HALF TWO IS **LIVE**, AND §D IS NO LONGER UNIDENTIFIED
>
> Two things changed after this file was first written. Both are good news and neither
> creates work for this lane.
>
> **1. A fresh chassis build rolled at 10:14Z and it carries half two.** `v1.0.1301`, both
> replicas, **presence proven at the artefact**: three-needle single-pass binary probe returning
> this change's own literal `re-audited this site on` **present**, long-live control
> `NO_CHANGE_GATE_UNREADABLE_RESULT` **present**, nonsense needle **absent**.
> ⚠ **Two traps in doing that, both paid for:** the `build provenance` startup line was already
> out of `--tail=3000` on **twenty-minute-old** pods on this service, so
> `merge-base --is-ancestor` was not available (absence there means *not in range*, never
> *unstamped*); and the binary probe takes **~2 minutes per pod**, so a two-pod loop hits the
> default command timeout and **the second replica reads as silence — which looks exactly like a
> replica that does not carry the code.** Probe them one at a time.
> **LIVE IS NOT EXERCISED.** Re-checked at 10:40Z: both carriers still `enabled=false`, and the
> population is unchanged at 19 `complete` / 4 `failed`. Everything §"THE ONE THING A NEW READER
> WILL GET WRONG" says below still holds, word for word.
>
> **2. §D's open question is ANSWERED, and this lane's own claim about it was FALSE.** The
> re-fired `090` (`adecf408`) returned **`UNVERIFIABLE`** — but it cited a config quote that
> contradicted our first-hand claim that *"neither agent declares a `complete_work_item` step"*,
> and **verified today, the loop is right and we were wrong.** The claim was an artefact of the
> PATH we searched: `workflow->steps` genuinely has no such step, but a `$.**` search over the
> whole config hits one at
> **`workflow.steps.process_item.config.sub_workflow.mark_complete_step`**, binding
> **`result: handler_result`**. So "the site binding the `result` input is unidentified" is no
> longer true. Full note, with what is verified and what is only a lead, in
> `bugs_closed/213` §"POST-CLOSURE NOTE 2026-08-15". **The bug file stays closed.**

---

## THE ONE-PARAGRAPH STATE

**THIS LANE IS FINISHED AND ITS BUG IS CLOSED.** The owner ruled on 2026-08-15 that half two
could proceed and that 213 and 216 could close; all three are done. D1 half two is built,
council-**APPROVED round 1** (11 minutes), registered as **WII-018**, and it also extracted the
shared retraction helper WII-016's own council round asked the third adopter for — migrating
WII-016 onto it, with that lane's 14 tests passing unrewritten. `bugs_open/213` and
`bugs_open/216` are now in `bugs_closed/`, verified at HEAD. **Nothing is running and nothing is
spending.** Two things sit outside this lane: the §D diagnosis, **re-fired today and currently
running** (`adecf408-1e60-4293-8b22-351ddbb52a08`) after the 08-14 run died on a usage cap that
turned out to have lifted 90 minutes later; and `bugs_open/274`, which is the owner's with
another thread.

---

## WHAT IS DONE

| | evidence |
|---|---|
| **D3** — `verifier-remit-check`, the daily class detector | `WII-015`; CronJob live `25 7 * * *`, image `v1.0.1289`, deployed and run |
| **D1 half one** — gate 1b, refuses a completion the handler never earned | `WII-017`; live `v1.0.1299`, council APPROVED, **both arms proven on production traffic with a 1:1 accounting** |
| **D1 half two** — silence retraction, the way back OUT of `failed` | `WII-018`; built today, council **APPROVED round 1** (`54e3b698-3d18-4dd1-9d6f-badec7e331fa`), **8 of 8 mutations caught** |
| **213 closed** | `bugs_closed/213_…`, on option (a); the closure records the one branch that has never executed |
| **216 closed** | `bugs_closed/216_…`; it had met the bar since 08-08 and was held only by a superseded ruling |

Commits: `a620912f5` (half two + helper + WII-016 migration + register), `d103dfcea` (216),
`0c467cea3` (213), `0d40f25ad` (landmines + wrong call), `e3d61d7d4` (council objections
actioned), `dbe29bbd6` (register verdict), `c8c9677f1` / `e32dd5b6d` / `57fca16c7` (standing five),
`7dc57f7ef` (second wrong call).

---

## ⚠ THE ONE THING A NEW READER WILL GET WRONG

**Half two is INERT, exactly as gate 1b was, and for the same reason.** `improvement-sweep` is
`enabled=false` (off 2026-08-14 16:41Z on the owner's cost decision, measured **6.0x** baseline)
and `site-discovery-rotation-design` is disabled too. **They are the only carriers that dispatch
this audit.** So:

- **The false-green bleed is paused by the sweep being OFF, not by the gates.** The gates are
  what make re-enabling it safe. Any claim that they stopped the bleed is false.
- **When someone re-enables a carrier**, the runbook §8 has the two queries. The expected first
  sign is silence counters appearing in `result->'retraction'` and CLIMBING.
- **⚠ THE CORRECT OUTCOME ON THE 4 LIVE `failed` ROWS IS THAT THEY DO NOT RETRACT.** They are
  genuinely unrepaired, so a run that closes them is the bug, not the proof. This looks like
  failure and is not.

---

## WHAT IS STILL OPEN, AND WHOSE IT IS

**1. §D — a completed item carries a payload its handler never produced. ✅ UPDATED: the binding
is now IDENTIFIED and the remaining question is a small code read, not a diagnosis run.** 10 of
14 completed rows carry a foreign but well-formed payload; the abstain arm reproduced the split
live at 3:1, so it is systematic. The `090` re-fired today (`adecf408`) completed
**`UNVERIFIABLE`** — but its citations moved it a long way. What is now known:

- **[VERIFIED first-hand]** `complete_work_item` binds `result` from **`handler_result`**, in
  `build-dispatch-loop` at `workflow.steps.process_item.config.sub_workflow.mark_complete_step`.
  Our earlier "no such step exists" was an artefact of searching `workflow->steps` only.
- **[THE LOOP'S CITATION, not independently verified]** the two foreign shapes match
  `webdesign-agent`'s `design_spec` and `content-gap-planner`'s `gap_plan` `output_fields`
  **exactly** — so it names two specific producers rather than "something foreign".
- **✅ [READ AND VERIFIED AT SOURCE — the lead was right, and it is already somebody's RFC]**
  `complete_work_item` declares `result` as Optional → the config maps it `"handler_result"` →
  `IsDottedPathReference` is literally `strings.Contains(s, ".")` so **Strategy 0 skips a
  single-segment mapping** → Strategy 2's `ExtractFields` runs `findFieldRecursive` for **any
  key named `result`**, depth 20 → and Strategy 4, the arm that exists to resolve exactly that
  shape, **skips because the field already has a value.**
  **§D is an incident of `architecture_review/RFC_029`**, filed 08-14 by the
  `staged_component_build` lane and **RULED by the owner 08-15** (implementation OPEN).
  Contributed into their file, not filed separately, with two things they did not have: a
  **different entry condition** (RFC_029 frames it as a field the caller never mapped; here the
  caller DID map it and lost for want of a dot), and a warning that **"unique-or-nothing" may
  not cover this case** — a lone foreign `result` is *unique*, so the ruled remedy resolves it
  wrongly with full confidence. ⚠ Still **[NOT VERIFIED]** that it actually fired for those 10
  rows: the mechanism is *available*, not proven to have *run*.

**NOTHING FURTHER IS OWED BY THIS LANE ON §D.** Its answer now lives where the fix will be made.

**Do not borrow `bugs_open/274`'s mechanism as the answer** — a *delivered failure* predicts
errored items, not `complete` items carrying a well-formed foreign payload. Raised 08-14,
downgraded the same evening by its own evidence, and 274 §10 does not restore it.

**2. `bugs_open/274`** — the owner's, with another thread. ~15,000 events since 08-03, still
firing. Note that **216's fix is what makes 274's replays real**, so 274's traffic is inflated by
fictional failures; if it is fixed at the header, record 216's arm's replay volume *first* or the
before/after is gone.

**3. A handler that can actually repair a dark section does not exist and is nobody's task.**
Half two gives these items an honest exit when the fault goes away; nothing yet fixes one.

---

## TRAPS THIS SESSION PAID FOR (beyond the 08-14b list, which still stands)

1. **`trg_site_work_items_updated_at` bumps `updated_at` on EVERY write to `site_work_items`, and
   the `stale-work-item-reaper` keys on it.** Any periodic bookkeeping write to an open row makes
   `triaged` rows permanently unreapable, and the damage is an *absence* — a park that never
   happens, with nothing to inspect. Enumerate the readers before writing (runbook §1). Filed
   fleet-wide in `LANDMINES.md`.
2. **A test can pass because the MOCK refused the call, not because your guard did.** Mine did,
   with a confident comment on it. sqlmock cannot express "this must not happen" either. Test the
   FUNCTION directly with a bare mock and assert on the ERROR RETURN. Caught by the mutation
   matrix and by nothing else.
3. **A GREEN mutation means one of three things and they need opposite responses:** the test is
   vacuous (rewrite it), you hit a guard in SERIES (pass the condition down so each can be
   mutated singly), or the code is unreachable (DELETE it — do not write a test to justify it).
   All three occurred in one run.
4. **A vendor's stated reinstatement date is a promise, not a measurement.** "Down until
   2026-09-01" lasted ~90 minutes. `LANDMINES.md` already carried that correction **twice** — I
   grepped it for the cap's signature and not for its duration. Verify a lift on the SUCCESS
   side; failures stop appearing either way (runbook §7).
5. **Correcting a shared file does not correct the doc that quoted it.** The fleet-wide landmine
   was struck through on 08-14; this lane's own NOTES kept the false claim for a day.
6. **The council's best objection was one I could have run myself in a minute.** It asked whether
   a safeguard strands the population the change exists to serve. Before submitting, ask of every
   stated *risk*: does it actually apply to the rows in front of me? A disclosed risk is not a
   checked one.

---

## IF YOU ARE PICKING THIS LANE UP

There is no queued work. The useful things to do, in order of value:

1. ~~Read the §D verdict~~ ~~then read `ExtractActionInputs`~~ — **BOTH DONE.** §D is
   `RFC_029`'s, contributed into their file. Gate 1b's `NO_CHANGE_GATE_UNREADABLE_RESULT`
   stream is the free before/after when their fix lands. **Do not pick this up here.**
2. **When a carrier is re-enabled**, run runbook §8 and confirm the streaks climb and the four
   rows stay open. Nothing to do until then — the code is live and waiting.
3. **Do not re-open 213 to do either.** Both are recorded as post-closure follow-ups; re-opening
   a closed file to add a verdict is how a closed bug becomes ambiguous.
4. **If you probe a pod to check a deploy, probe ONE AT A TIME** — ~2 minutes per pod, and a
   loop that times out makes the second replica read as unstamped.
