# HANDOFF — bug sweep, 2026-09-03c (session **bugsweep4**)

> ⚠ **THIS IS A DIFFERENT SESSION'S SWEEP FROM `HANDOFF_2026-09-03b_continue_here.md`.**
> That file belongs to another concurrent bugsweep and covers 442/338/320/404/407. **Nothing
> here supersedes it and nothing here contradicts it** — we worked disjoint bugs. If you are
> resuming *that* thread, read that file; this one is bugs **361, 366, 400**.

## 0. STATE IN ONE TABLE

| bug | what it was | state now | who must act |
|---|---|---|---|
| **361** render-check ratchet | red 25 days, 478 manufactured "NEW" findings | **FIXED, committed, rides the next release.** 478 → **18 real + 460 unbaselined** | **OWNER: gate or bank the 18** (§2) |
| **366** corpus admission | unreported usage read as "finished normally" | **FIXED, committed.** Hole measured **empty**; a blanket fix would have deleted 161 real rows | nobody — ratify §4 if you disagree |
| **400** news goto URLs | 1,378 Google redirect links served as article links | **DIAGNOSED + HANDED OFF, no code.** Bug **inverted** — intake stopped, served damage live | next session: `bugfix_400_news_goto_urls/HANDOFF_2026-09-03_start_here.md` |

**Commits:** `051c73d1e`, `d716c837a` (361 code) · `29eee3bc6` (366 code) · plus docs commits.
Both fixes are in HEAD ahead of the pending `v1.0.1358` build. `component-render-check` is in
`RELEASE_IMAGES`, so it ships with the release.

## 1. START HERE IF YOU ARE FRESH

**The one piece of unfinished technical work is 400**, and its lane handoff is written to be read
cold: `docs/agent_docs/docs024_key_docs_latest/bugfix_400_news_goto_urls/HANDOFF_2026-09-03_start_here.md`.
Diagnosis is complete; no code exists. It is **in council scope**, so it wants a round — unlike 361
and 366, which are not.

Everything else on this page is either done or waiting on the owner.

## 2. ⚖ OWNER DECISION — the 18 render-check findings: gate or bank?

The check now fails on **18 findings in 5 components**, and stays red until they are dealt with.
All five were **edited after the baseline** — two of them *the same day* the check reported them —
so these are **rewrites, not decay**, and they are **one class**: a rewrite added label/heading/cell
fields without gating them, so an absent field renders an empty element.

- **Gate** (recommended) — edit the 5 templates so an absent field renders nothing. Fixes the real
  defect; one pattern applied 18 times.
- **Bank** — regenerate the baseline. Green immediately, holes stay on the pages. ⚠ **Blocked**:
  `--write-baseline` refuses today because **2 templates fail to parse**, and that guard is correct.

**Why I did not just do it:** the five are live components belonging to other lanes and sites.
Full list in `bugs_open/361` §2026-09-03 (later). One is `conversion_rate_aria_label` — an
accessibility label rendering empty, which no visual check would ever surface.

## 3. ⚖ OWNER DECISION — a gap I opened deliberately, currently owned by nobody

361's fix narrows what a red means: it now fails only on components the baseline **covered**, so
**a hole in a component born after the baseline fails nothing.** That is what stops the job going
red as the library grows (282 → 497 components).

`bugs_open/361` says that debt "belongs to birth-time gating (CGV-029)". **I checked — it does
not**: `component-fallback-check` sees only fields declared `on_missing:"skip_field"`, which is
precisely the blindness CGV-030 exists to close. So it is an **open gap, not a delegation**, and it
is recorded as such in CGV-030's `verify-later` and in the bug file.

Current population: **460 findings across 62 components**, listed by name in the daily `doc_notes`
row. I did **not** build a gate — a new gate on a shared mint arriving inside a bug patch is the
`bugs_closed/124` veto shape. The `components` lane confirmed they are not on it.

**A third blindness, unrelated to the ratchet**, was contributed by that lane and verified by me:
a per-item slot inside a `{{range}}` can hold a hole this check **cannot synthesise**, because it
probes by removing a *declared* field and per-item declarations describe shape only (253 flat
declarations, **zero** carrying `on_missing` or `source`). **Do not read a green CGV-030 as covering
it.**

## 4. What 366 decided, so nobody re-opens it

366 reserved "what should an unknown-usage row do?" for whoever owns the corpus. **I took it on
measurement, not preference**, and the measurement inverted the obvious answer:

- the filed hole (`success` + usage unreported) is **ZERO rows** over the corpus this tool actually
  reads — a latent hazard, which 366 §4 itself says is "a fine outcome to record";
- a blanket "exclude everything unverifiable" would have deleted **161 real rows**, 8% of the
  corpus, averaging 3,092 output tokens.

So only the genuinely anomalous shape is excluded. Reversible in a line if the owner wants stricter.
⚠ **366's own supplied census query returns 882 and is measured against the wrong population** — it
counts all of `llm_call_log`, not the step_names the tool ingests. The file now says so.

## 5. Also surfaced, not taken

**`bugs_open/257`** has been waiting on the owner since 2026-08-16 — its file ends *"Owner call
needed on both: in scope for this bug, or separate lanes?"*. I evaluated it as a candidate and left
it for exactly that reason. Two items: a duplicated precedence rule, and direct LLM calls writing no
`llm_call_log` row, so **every truncation instrument the estate has is blind to that population**.

**`bugs_open/349`** — OPEN, UNOWNED, live and growing (42→**58** rows, 14→**18** sites since filing).
Not taken because its obvious fix is explicitly gated: the filing lane's handoff says *"do not narrow
`PageWantedLivePredicateFor` without architecture"*. A real candidate for someone with an RFC appetite.

## 6. Traps banked this session (all in the fleet ledgers)

- **`LANDMINES`** — a jsonb absence filter written with `jsonb_typeof` returns **zero rows**, and
  zero reads as a finding (`NOT NULL` is NULL, `WHERE` discards). Use `?`, which never returns NULL.
  Verifier dispatched.
- **`WRONG_CALLS` ×5** — a mutation-proof test that passed under the mutation it was written to
  catch (two guards in series); a fix whose distinguishing branch today's data cannot execute, so
  the live run agreed before *and* after; a census run against a wider population than the tool
  reads; relaying a peer's arithmetic into the register unverified; and **twice** a figure that was
  wrong because **nothing in the argument depended on it**.
- The last of those is the transferable one: **a number that does no work is never tested by the
  argument carrying it**, however carefully taken and however well marked.

## 7. Cross-lane state, so nobody re-treads it

- **`components`** — told about the narrowed guarantee; replied that they are **not** on the birth
  gate and will not be. They corrected a claim of their own as a result, and contributed the
  per-item finding in §3.
- **`bugs_open/437`** — routed the legacy-dialect evidence to them directly (the components lane
  could not resolve their session; it is live and named `bugs_open/437`). They settled 240 vs 437 as
  **distinct defects**, and used a population we handed them to falsification-test a claim their
  whole submission rests on. Their `PBP-052` now carries the stronger form.
- **`idea_uk_vm_site`** filed 400 and explicitly is not fixing it; tell them when it lands.
