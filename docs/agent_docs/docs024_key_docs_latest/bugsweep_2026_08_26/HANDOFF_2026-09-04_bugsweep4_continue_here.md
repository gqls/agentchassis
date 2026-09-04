# HANDOFF — bug sweep, 2026-09-04 (session **bugsweep4**)

> **Read this instead of `HANDOFF_2026-09-03c_bugsweep4_continue_here.md`.** That file is this
> lane's previous state and is still correct about 366 and 400. **What changed: the v1.0.1360
> roll landed, 361's fix is LIVE and PROVEN, and reading the live row found one more defect of
> mine (fixed).** Where 09-03c is now out of date I say so here rather than editing it.
>
> ⚠ **`HANDOFF_2026-09-03b_continue_here.md` is a DIFFERENT session's sweep** (442/338/320/404/407).
> Disjoint bugs. Nothing here supersedes it.

## 0. STATE IN ONE TABLE

| bug | state | blocked on |
|---|---|---|
| **361** render-check ratchet | ✅ **FIXED + LIVE + PROVEN AT THE ARTEFACT** (v1.0.1360). 478 manufactured → **18 real + 460 unbaselined** | **OWNER §3** — the 18 are undispositioned, so the job stays red |
| **366** corpus admission | ✅ **FIXED, committed.** No image, no roll — "live" = the next corpus build runs it | nobody. Ratify §5 only if you disagree |
| **400** news goto URLs | 🔶 **DIAGNOSED, HANDED OFF, NO CODE.** Intake still stopped (7 days), **served damage still live today** | next session — lane handoff is written cold |

**Commits:** `051c73d1e`, `d716c837a`, `c0b59ba30` (361) · `29eee3bc6` (366) · docs commits alongside.

## 1. ✅ WHAT THE ROLL PROVED (361), and how — because the method matters more than the result

`component-render-check` is on `v1.0.1360` and ran itself at 06:55Z. **The proof is not the tag.**
It is the shape of the row the job wrote:

| day | first line |
|---|---|
| **09-04** | `… 459 of 504 active …, **18 REGRESSION, 460 unbaselined across 62 new component(s)**, 56 fixed, 0 UNCOVERED` |
| 09-03 | `… 425 of 490 active …, **478 NEW**, 60 fixed, 0 UNCOVERED` |

That vocabulary exists only in the new binary, so no git ancestry or `strings` probe is needed.

**The control that could have come out otherwise:** the library grew **490 → 504** overnight and
the regression count stayed at **18**. Under the old ratchet those 14 new components would have
manufactured more findings. *Growth with a flat regression count* is the disconfirmable form of
"the scoping works". Body verified too: **536 lines — 18 / 460 / 56** listed by name.

## 2. ⚠ ONE DEFECT OF MINE THAT ONLY THE LIVE ROW COULD FIND (fixed, `c0b59ba30`)

The row read `"…close that blind spot., 3 inherited from an identical template (…)"`. The
clone-suppression count had landed on the **legacy-warning line** instead of the summary line,
because the warning starts with `\n` and I appended `inherited` after it. The daily series query
reads `split_part(body, E'\n', 1)`, so a count that exists *precisely so a filter cannot hide its
own effect* (owner ruling 2026-08-05) was invisible to the only query anyone runs.

**No test caught it, and none should have** — the tests pin the classifier; this is report
assembly, a string built in `main()`. **The check that works is reading the live row after a
roll**, now in the lane RUNBOOK with both the broken and correct SQL forms.

## 3. ⚖ DECISION ONE — the 18 findings: gate or bank?

**What a ratchet is.** It holds a list of holes it already knows about and fails only when a *new*
one appears in a component it had vouched for. **The rule:** it stays red until the new ones are
dealt with, and a permanently-red check is one people stop reading — that is how this went
unnoticed for 25 days. **This case:** 18 findings, 5 components.

All five were **edited after the baseline** — two of them the same day the check reported them — so
these are **rewrites, not decay**, and they are **one class**: a rewrite added label/heading/cell
fields without gating them, so an absent field renders an empty element.

- **Gate** *(my recommendation)* — edit the 5 templates so an absent field renders nothing. Fixes
  the real defect; one pattern applied 18 times. Clears the class and lets a later regeneration
  land on a settled state.
- **Bank** — regenerate the baseline. Green immediately, holes stay on the pages.
  ⚠ **Blocked today**: `--write-baseline` refuses because **2 templates fail to parse**, and that
  refusal is correct (baselining a blind run bakes the blindness in as "clean").

**Why I have not done it:** the five are live components owned by other lanes and sites. Full list
in `bugs_open/361` §2026-09-03 (later). One is `conversion_rate_aria_label` — an **accessibility**
label rendering empty, which no visual check would ever surface.

**Consequence of not deciding:** the job stays red, `lastSuccessfulTime` stays at 2026-08-09, and
the check drifts back to being ignored — the exact failure it was just rescued from.

## 4. ⚖ DECISION TWO — who owns the gap the fix deliberately opened?

**What changed.** The check used to fail on any hole it did not recognise; it now fails only on
holes in components it had **covered**. That is what stops it going red as the library grows.
**The consequence:** a hole in a component born *after* the baseline **fails nothing**.

**The rule the bug file assumed:** that debt "belongs to birth-time gating (CGV-029)".
**The case:** I checked — **it does not.** `component-fallback-check` sees only fields declared
`on_missing:"skip_field"`, which is precisely the blindness CGV-030 exists to close. So this is an
**open gap, not a delegation**. Population today: **460 findings across 62 components**, listed by
name daily.

**Why I did not build a gate:** a new gate on a shared mint arriving inside a bug patch is the
`bugs_closed/124` veto shape. The `components` lane confirmed they are not on it and will not be.

**A third blindness, unrelated, verified independently:** a per-item slot inside a `{{range}}` can
hold a hole the check **cannot synthesise** — it probes by removing a *declared* field, and per-item
declarations describe shape only (253 flat declarations, **zero** carrying `on_missing`/`source`).
**Do not read a green CGV-030 as covering it.**

## 5. What 366 settled, so nobody re-opens it

366 reserved "what should an unknown-usage row do?" for whoever owns the corpus. **Taken on
measurement, and the measurement inverted the obvious answer:** the filed hole is **ZERO rows** over
the corpus this tool actually reads, while a blanket "exclude everything unverifiable" would have
deleted **161 real rows** — 8% of the corpus, averaging 3,092 output tokens. So only the genuinely
anomalous shape is excluded. Reversible in a line.

⚠ **366's own supplied census query returns 882 and is measured against the wrong population** (all
of `llm_call_log`, not the step_names the tool ingests). The file says so now.

**No roll applies** — `cmd/reasoningset` has no image and is not in `RELEASE_IMAGES`. "Live" means
the next corpus build. Whoever runs one should confirm `usage_unreported` appears zero times and the
row count does not drop by ~161.

## 6. 🔶 THE ONE PIECE OF UNFINISHED TECHNICAL WORK — 400

**→ `docs/agent_docs/docs024_key_docs_latest/bugfix_400_news_goto_urls/HANDOFF_2026-09-03_start_here.md`**

Written to be read cold. Diagnosis complete, **no code**. Re-checked this morning:

- intake **still stopped** — 7 days now, zero goto rows against 57–289 items/day
- **served damage still live** — idea.uk serving **2 of 6** today
- 1,378 stored rows, 11 sites

**In council scope** (unlike 361 and 366), so it wants a round. Order in its handoff: detector
first, then unwrap at the bridge, then the backlog repair last — the repair collides with
`idx_cfi_dedup`, a partial UNIQUE on `source_url`, so measure the collision set before the UPDATE.

## 7. Evaluated and deliberately NOT taken

- **`bugs_open/257`** — waiting on an owner call since 2026-08-16; its file ends *"Owner call needed
  on both: in scope for this bug, or separate lanes?"*. Two items, the sharper being that direct LLM
  calls write no `llm_call_log` row, so **every truncation instrument the estate has is blind to
  that population**.
- **`bugs_open/349`** — OPEN, UNOWNED, live and growing (42→58 rows, 14→18 sites). Its obvious fix is
  gated: the filing lane says *"do not narrow `PageWantedLivePredicateFor` without architecture"*.
  A good candidate for someone with an RFC appetite.

## 8. Traps banked (fleet ledgers, all committed)

- **`LANDMINES`** — a jsonb absence filter written with `jsonb_typeof` returns **zero rows**, and
  zero reads as a finding. Use `?`, which never returns NULL. Verifier dispatched.
- **`WRONG_CALLS` ×6** — a mutation-proof test that passed under the mutation it was written to
  catch; a fix whose distinguishing branch today's data cannot execute; a census against a wider
  population than the tool reads; a peer's arithmetic relayed into the register unverified;
  **twice** a wrong figure that survived because *nothing in the argument depended on it*; and
  today, a `LIMIT` applied after `unnest` expansion that returned `0` and **accused my own
  just-shipped fix** — audited last, not first, because it confirmed a fear.
- The transferable one: **a number that does no work is never tested by the argument carrying it.**

## 9. Cross-lane state, so nobody re-treads it

- **`components`** — not on the birth gate, said so explicitly. Corrected a claim of their own after
  our exchange and contributed the per-item finding in §4.
- **`bugs_open/437`** — settled 240 vs 437 as **distinct defects**, and used a population we handed
  them to falsification-test a claim their whole submission rests on. Their `PBP-052` carries the
  stronger form.
- **`idea_uk_vm_site`** filed 400 and explicitly is not fixing it — tell them when it lands.
