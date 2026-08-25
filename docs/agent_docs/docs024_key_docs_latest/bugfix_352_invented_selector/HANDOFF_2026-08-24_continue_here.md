# HANDOFF 2026-08-24 — `bugs_open/352`, the invented selector: cold-start and what is left

**Read this first, then `NOTES_invented_selector.md` (newest at the bottom) and
`RUNBOOK_invented_selector.md`.** The bug file itself carries a status banner at the top —
`bugs_open/352_HANDOFF_2026-08-22_contrast_findings_name_a_selector_that_matches_nothing.md`.

---

## 0. One-paragraph state — UPDATED 2026-08-24 19:20 UTC

**Arm 1 is FIXED, council-APPROVED, LIVE ON BOTH IMAGES, and now PROVEN ON A LIVE PAGE.** Migration
**587 was applied by hand at 2026-08-24 19:11:22 UTC** (`UPDATE 73`), so both of the items this
handoff was written to hand over are **DONE**. Arm 2 is untouched and reproducible, which is why 352
stays in `/bugs_open/` — **it is now the only reason.** Nothing here is blocked and nothing is owed
to another lane. What a new session can usefully pick up is in §10 at the bottom, which is the part
of this document still in the future tense.

> ⚠ **§4 and §5 below are SUPERSEDED and kept for the trail.** §4 asked for two things that have
> since happened; §5 is a census taken on the near side of 587 and every one of its open-population
> figures is now **0 by design**. Read §3b for what replaced §4(a), and RUNBOOK §10 for which side
> of 587 you are on and the query that keeps returning 73 for ever.

## 1. What the bug is, in three sentences

The render audit recorded a class-less element's **tag name** in a field called `Class`, so the
orchestrator composed selectors like `H3.H3`, `P.P`, `A.A` — which select elements carrying
`class="H3"`, of which there are none. `css-patch-agent` faithfully wrote a CSS rule against that
selector, deployed it, and marked the work item `complete`; the text stayed unreadable.
**[MEASURED 2026-08-24] 181 of 452 `contrast_failure` rows carried such a selector, and 108 were
already `complete`** — repairs recorded that could never have applied.

## 2. ⚠ THE THING THAT MAKES THIS NOT A ONE-LINE FIX — read before touching anything

The bug file's own candidate (1) says to omit the class component so `H3.H3` becomes `h3`. **That
is a REGRESSION.** Today `p.P { … }` matches nothing and is therefore *harmless*. Corrected to `p`
it matches — and css-patch-agent's own live prompt says *"The platform APPENDS your rules to the END
of the stylesheet"*, one stylesheet per site — so the rule recolours **every paragraph on the site**.
`P.P` (77) and `A.A` (44) were **121 of the 181**.

So the shipped fix composes the selector **in the page** (class → own id → nearest ancestor with an
id or class → bare tag) and **asserts it selects the very element that was measured**. A bare tag is
refused and counted. The invariant is *"prove it"*, not *"stop lying"*.

## 3. What is LIVE, and how it was proved (do not re-derive this — but do re-check the dates)

`v1.0.1334`, all three overlays, pods started **15:39 UTC 2026-08-24**.

| service | half | evidence |
|---|---|---|
| `browser-runner-adapter` | producer | own startup line, `git_commit 70fd163c2` |
| `render-audit-adapter` | producer (shares the image, makefile:107) | same line, same commit |
| `agent-chassis` | consumer | startup line had scrolled; **binary probe** found `70fd163c2` |

`git merge-base --is-ancestor ffa6e1c3d 70fd163c2` → **YES** (fix 13:45 UTC, build commit 15:11 UTC,
pod start 15:39 UTC). **Capability probed too**, which is stronger than the commit: chassis carries
`skipped_unverified_selector` / `skipped_unanchored_selector` / `selector_scheme`; browser-runner
carries `verified/v1` / `indexOf.call(nodes,el)` / `selectorVerified`; and an invented control string
is absent from both, so the probe is not over-matching.

⚠ **My first negative control was worthless** — I chose a commit on the assumption it postdated the
build and it did not. See `WRONG_CALLS.md` 2026-08-24. If you re-run any of this, print the
timestamps of control, subject and build side by side **before** interpreting either result.

## 3b. THE ARTEFACT-LEVEL PROOF (this replaced §4(a), and it is better than what §4(a) asked for)

**The driven canary never ran.** Correlation `c2fce02e-…` has no orchestration row of any kind —
checked by correlation, by `site_id`, and by substring over `initial_request_data` and
`collected_data` — **2 h 15 min** after publish. It was NOT re-dispatched: by then the proof existed
and a second audit would only have cost credits. (Context: `render-audit-agent`'s only run ever
against that site was `16781a84-…` at 02:23 UTC, which ended `complete_error`.)

**What replaced it, unstaged.** `render-audit-agent` runs ~hourly across the estate, cycling sites.
Two runs straddle the 15:39 UTC roll:

| | run `6dc00a26`, **15:31:50** (old image) | run `e0bd33d0`, **17:33:16** (new image) |
|---|---|---|
| rows filed | 47 | 10 |
| invented `TAG.TAG` | **3** | **0** |
| `spec.selector_scheme` | absent on all 47 | `verified/v1` on all 10 |
| `spec.matches` | absent | present on all 10 |

⚠ Different sites, so **not** a controlled A/B — but the pre-roll arm is what makes the post-roll
zero mean anything: a bare `still_invented = 0` could not have come out otherwise if no class-less
element happened to be measured. The post-roll selectors are `.ported-page-content A` — ancestor
anchor, bare-tag leaf — i.e. **exactly** the class-less case the old code turned into `A.A`.

**Settled in the page, because `spec.matches` is the producer vouching for itself.** Fetched the
live pages over HTTPS (invented-path control per domain → 404, so a 200 is a real page) and counted
with an independent stdlib `HTMLParser` that walks the open-element stack:

| selector | page | producer said | independently measured |
|---|---|---|---|
| `.ported-page-content A` | `loancash.co.uk/guides/index.html` | 15 | **15**, all class-less |
| `.ported-page-content A` | `loancash.co.uk/guides/jargon-buster.html` | 8 | **8**, all class-less |
| `SPAN.SPAN` (pre-roll row, already `complete`) | `loanzy.uk/tools/loan-repayment-calculator/` | — | **0**, against 22 real `<span>`s |
| `LABEL.LABEL` (pre-roll row, already `complete`) | `loanzy.uk/tools/loan-comparison-calculator/` | — | **0**, against 6 real `<label>`s |

Parser controls: non-existent class → 0; same selector on the 404 body → 0; `class="A"` and
`class="H3"` occur **nowhere** in the markup. Script kept at `scratchpad/sel_check.py`.

**587 then applied** — `_VERIFY` arms 1–3 re-run fresh first (66 distinct keys, all `<path>#TAG.TAG`;
false-positive arm 0 of 31 with the positive control at 166 of 173), then
`UPDATE 73` at **19:11:22 UTC**, then arms 4–5: `open_invented = 0`, `withdrawn = 73`,
`withdrawn_without_prior_status = 0`, `falsely_completed = 0`.

⚠ **The fleet re-rolled at 18:32 UTC to `v1.0.1335` (`48f55f218…`) while this was going on, and the
fix survived** — `merge-base --is-ancestor ffa6e1c3d 48f55f218` YES with `HEAD` as a control that
correctly returns NO, plus a capability probe of the running binary (three symbols present, an
invented control string absent, build sha present, nonsense sha absent). Timestamps were printed
side by side **before** interpreting either result, which is the correction §3's warning is about.

## 4. ~~THE TWO OPEN ITEMS~~ — **BOTH DONE 2026-08-24 19:11 UTC. Superseded by §3b; kept for the trail.**

### (a) READ THE CANARY — in flight, this is the real proof

- **Correlation:** `c2fce02e-2fe7-489f-bc22-edcfa75b0761`
- **Site:** `ai-agent-orchestration.com`, `2a8ebf9c-20a2-4c39-b191-840b012371da`
- **Dispatched:** 2026-08-24 ~16:52 UTC via `kafka_publish_checked` (receipt confirmed: `PUBLISHED`)
- **Why that site:** it is **`bugs_open/211`'s own site** — its `[UNRESOLVED]` §4 item is written
  around six class-less `<h3>`s mislabelled `.H3` — and it holds 8 invented rows.

⚠ **A missing orchestration row is LATENCY, not a dropped dispatch** (~29 min publish→start measured
under fleet load). **Do not re-dispatch on that evidence** — it costs a duplicate audit. Find it by
payload:

```sql
SELECT current_step, status, created_at FROM orchestration_states
 WHERE collected_data->'input_data'->>'site_id' = '2a8ebf9c-20a2-4c39-b191-840b012371da'
 ORDER BY created_at DESC LIMIT 3;
```

**The measurement that decides it** — and it can come out either way, which is the point:

```sql
SELECT count(*) AS rows_since_roll,
       count(*) FILTER (WHERE item_key ~ '#([A-Z][A-Z0-9]*)\.\1$') AS still_invented,  -- MUST be 0
       count(*) FILTER (WHERE spec ? 'selector_scheme')             AS scheme_stamped,  -- should equal rows
       count(*) FILTER (WHERE spec ? 'matches')                     AS carries_matches
  FROM site_work_items
 WHERE item_type='contrast_failure' AND created_at > '2026-08-24 15:39:00+00';
```

Then take **one** fresh row and settle it at the artefact rather than in the database — open the
affected page and ask the browser whether the filed selector actually matches:

```js
document.querySelectorAll("<spec->>'selector'>").length   // must be >= 1
```

`still_invented > 0` refutes the fix. `rows_since_roll = 0` after the audit completes means the
audit found nothing on that page (possible — check `summary.pages_audited`), **not** that the fix
works.

### (b) APPLY MIGRATION 587 — gate already satisfied, hold for the canary

`docs/agent_docs/sql_for_agents/587_retire_invented_contrast_selectors_HOLD.sql` (+ `_ROLLBACK`,
+ `_VERIFY`). It **withdraws** the 73 open invented-selector rows as `cancelled` — **withdrawal, not
resolution** — freeing their dedup slots so still-failing pairings return under verified selectors.

- **Ordering gate: MET.** Both images confirmed at the artefact (§3). The `_HOLD` suffix means
  applied by hand; it is not waiting on anything else.
- **Why hold for the canary anyway:** applying early is *churn, not corruption* (the old code would
  refill the freed slots) — but seeing one good row first costs nothing and turns an argument into a
  measurement.
- **Run `_VERIFY` arms 1–3 first.** They have been executed read-only against the live DB already:
  66 distinct keys listed, all `TAG.TAG`; the false-positive arm found **0 of 31** tag tokens
  existing as real classes, **with a working positive control** (154 of 161 genuine class tokens
  found by the same predicate). Arms 4–5 are for after.
- ⚠ `run-migrations.sh --apply` takes **every** pending file including other lanes' backlog. Scope
  it, or apply this one by hand.

## 5. ~~State of the data RIGHT NOW~~ — **SUPERSEDED: this is the NEAR side of 587**

> ⚠ **Two corrections to the table below, and one of them is about the date on it.**
>
> 1. **It is pre-587.** Every open-population figure in it now reads **0**, by SUBTRACTION, which
>    looks like "this never happened" rather than "we fixed it". RUNBOOK §10 has the which-side
>    query and the recovery query that keeps returning **73** for ever.
> 2. **`452` was never true at the time this table claims.** The label carried the time the handoff
>    was *written*, not the time the census was *run*, and a scheduled audit filed 47 rows in
>    between. `509 − 10 post-roll − 47 = 452`, so 452 was the total before 15:31:50; at 16:55 it was
>    499. → `WRONG_CALLS.md` 2026-08-24. **Date a figure when you measure it, not when you write it.**
> 3. **The damage figure is 111, not 108** [MEASURED 2026-08-24 19:10 UTC] — the 15:31 audit filed
>    three more invented rows and they closed `complete`. 111 is the permanently-quotable number;
>    587 never touches it.

### The superseded table [measured ~15:30 UTC, mis-labelled 16:55]

| | |
|---|---|
| `contrast_failure` total | **452** |
| invented-selector rows | **181** (`complete` 108 · `deferred` 58 · `unresolved` 15) |
| **open** invented rows (587's population) | **73**, across **13** sites |
| rows filed since the roll | **0** (no audit has run yet) |
| `withdrawn_by_587` | **0** — 587 is committed and **NOT applied** |

⚠ **After 587 applies, every census above returns ZERO by design** — staleness by *subtraction*,
which reads as *"this never happened"* rather than *"we fixed it"*. RUNBOOK **§10** carries the
which-side-of-587 query and the recovery query that keeps returning 73 for ever. **The 108 is never
touched by 587 and is the permanently-quotable damage figure.**

## 6. What was built (so you do not re-derive it)

| file | what changed |
|---|---|
| `internal/adapters/browserrunner/render_audit_action.go` | in-page composition + verification; `selector`/`matches`/`selector_verified`; `summary.selector_scheme` stamped unconditionally. **The `cls` tag-name fallback is KEPT as a frozen legacy echo — do not "clean up"**, an un-rolled chassis reads it |
| `platform/orchestration/actions/write_render_audit_findings_action.go` | `filingSelector` (prefers verified, falls back to today's exactly), two categorical refusals with counters, `selectorLockTokens`, retraction **alias keys** + **scheme guard**, reworded retraction reason |
| the two test files | 7 tests, **every guard mutation-proven** |
| `scripts/render_audit.py` | mirrors the composition; prints the selector instead of `.<cls>` — this is the line that misled 211 |
| `docs/agent_docs/sql_for_agents/587_*` | the withdrawal + `_ROLLBACK` + `_VERIFY` |

**Commits:** `ffa6e1c3d` (code), `587` migration, `4741758d6` (prose/register/probe), `7d18c8c83`
(index row), `61b84c937` + `ffdca67fd` (landmine + its amendment), `07f8f3c21` (post-roll notes),
`1ef3f602c` (WRONG_CALLS). **Council `acadbe8b-f131-4d4b-b4de-5b61f0898f93` — APPROVED round 1.**

Register: **VIZ-016** (selector contract + the shared `item_key` shape) and **WII-016** (why alias
keys were needed).

## 7. ARM 2 — still live, and NOT designed. Do not let the file read as fixed

Even with a correct selector, the appended rule can be inert: for the `~1.0x:1` family the offending
declaration lives in **page-level component CSS emitted AFTER the stylesheet the agent edits**, so an
equal-specificity rule loses on source order however right it is. `bugs_open/296` §10.5 reaches the
same finding from the other end.

Sketch only: css-patch-agent's workflow gains a **measurable precondition** — grep `css_themes` for a
declaration governing the filed selector's property; if the offending declaration is not in the file
the agent can edit, **refuse and park** with a `parked_by` marker (198's `mark_base_unsafe` shape)
rather than append a rule that cannot win. And completion should consult the spec's own
`acceptance_test` at the `checks.GetVerifier` / `verifyBeforeComplete` choke point — which
`write_audit_findings_verifier_join_test.go:85` confirms **nothing reads today**.

## 8. Owed, and honestly not done

- **The two new counters have no automated reader.** The council's `bug_historian` seat raised this
  (medium) and I did **not** close it: `skipped_unverified_selector` / `skipped_unanchored_selector`
  ride the action's result map and its log line, and nothing surfaces them. That is honest
  bookkeeping with no consumer, which is a weaker position than it sounds — a genuine finding
  withheld from a fixer is practically invisible if the counter is never read.
- **The locked-component interaction is narrowed, not proven closed.** [UNPROVEN] I never
  established that the old substring check dropped a real finding; the mechanism permitted it.

## 9. Other lanes — told, nothing owed back

- **`brochure_component_library`** (`bugs_open/296`): 73 of their durable 171, on 13 of 15 sites,
  were unexecutable by the fixer. They have an **open owner decision** about releasing them — CONTRIB
  filed in their directory. **Their count will step-change when 587 applies; that is this lane, not
  drift.**
- **`bugfix_122_contrast_ink_slots`** (`bugs_open/211`): the "six `.H3` headings" correction. CONTRIB
  filed and a dated line added to 211 §4. ~~**Their site is the canary.**~~ **The canary on their
  site never ran** (§3b) and the proof came from the scheduled rotation on two other sites instead.
  Their `.H3` correction is unaffected — it never depended on the canary.
- **`bugs_open/198`**: filed this bug and released it; eight corrections passed between us and every
  one was found by the other side. Closed out.

## 10. WHAT IS ACTUALLY LEFT — the only part of this document still in the future tense

Written 2026-08-24 19:25 UTC, after §4's two items closed. In rough order of value:

1. **Arm 2** (§7). Still live, still reproducible, **not designed**. This is the reason 352 is open
   and the only thing here that changes what a user sees. §7 is a sketch, not a plan — the
   measurable precondition and the `checks.GetVerifier` choke point both need a real design pass,
   and `bugs_open/296` §10.5 reaches the same finding from the other end, so read it first.
2. **Give the two counters a reader** (§8). They now certainly *fire* — the composition path is
   live and producing — but nothing surfaces them, and the 18:32 fleet roll demonstrated the cost:
   the 17:33 audit's `write_render_audit_findings: complete` line, counters and all, was gone from
   the chassis logs within an hour. **A counter whose only sink is a log line on a service that
   restarts is not bookkeeping, it is a hope.**
3. **Re-check the withdrawal actually re-detects.** 587 freed 73 dedup slots on 13 sites on the
   promise that still-failing pairings return under verified selectors. ~~within ~14 days … from
   **2026-09-07**~~ — **CORRECTED 2026-08-25: the check date is 2026-08-28.** The fortnight came
   from `contrast_failure.created_at` (when a finding was last FILED, which only happens if an audit
   found something); the real cadence is the rotation stamp, and the live `pre_query` window is
   **3 days**. [MEASURED 2026-08-25 09:40 UTC] all 13 sites last selected BEFORE 587 applied, **0**
   audited since, earliest due **2026-08-26 21:20 UTC**, all 13 by ~**2026-08-27 21:30 UTC**. So from
   **2026-08-28**, any of those 13 with no re-filed `contrast_failure` and a visible contrast fault
   is a defect in this promise. The recovery query in RUNBOOK §10 gives the 73 to check against.
4. **Not ours. THE GREP HAS NOW BEEN DONE — it is unfiled, and I did not file it.** `render-audit-agent`
   fails more often than it succeeds — [MEASURED 2026-08-24 19:08 UTC] **11 of 20** runs
   ~~over 7 days~~ **in ONE DAY** (corrected 2026-08-25: `orchestration_states` is pruned to ~24 h,
   so the `interval '7 days'` filter excluded nothing and the window was never 7 days; the rate is
   real for that day and **is not reproducible**, because those rows have since been pruned) ended
   `complete_error`, every one on `Request timed out (code: TIMEOUT)` at almost exactly 3 minutes,
   and that rate **predates this lane's change**. It is also the clock on item 3: a
   re-detection window measured in audits is only as good as the audits landing.
   ⚠ And the post-roll sample is **3 runs (2 errored)** — that cannot distinguish 55% from 67%, so
   "no regression from our change" is **unproven, not established**. Re-check after ~20 post-roll
   runs (≈ a day at the current cadence).
   **Prior-art search, 2026-08-24 19:40 UTC, so nobody repeats it:** nothing in `/bugs_open/` or
   `/bugs_closed/` covers it — the nearest is `bugs_open/296`, which mentions both the render audit
   and a timeout but is about parked contrast findings, not the audit's own failure rate. The
   `needs_diagnosis` queue holds **no open item at all** (49 complete / 8 failed / 5 cancelled, none
   naming the render audit). **So it is genuinely unfiled**, and I left it that way deliberately: I
   have a symptom count and no cause, and this estate's rule is that a cross-cutting root cause goes
   through `090` *before* it is asserted, not after. Whoever picks it up should run
   `090_TRIGGER_needs_diagnosis_v1.sh` rather than write a mechanism into a bug file from these
   numbers. Re-measure first — the figures above are 7-day and will have moved.
