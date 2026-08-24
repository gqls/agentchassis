# HANDOFF 2026-08-24 — `bugs_open/352`, the invented selector: cold-start and what is left

**Read this first, then `NOTES_invented_selector.md` (newest at the bottom) and
`RUNBOOK_invented_selector.md`.** The bug file itself carries a status banner at the top —
`bugs_open/352_HANDOFF_2026-08-22_contrast_findings_name_a_selector_that_matches_nothing.md`.

---

## 0. One-paragraph state

**Arm 1 is FIXED, council-APPROVED, COMMITTED, and LIVE ON BOTH IMAGES as of 2026-08-24 15:39 UTC
(`v1.0.1334`).** Arm 2 is untouched and reproducible, which is why 352 stays in `/bugs_open/`.
**Two things remain and neither is blocked:** (1) a canary render audit is IN FLIGHT and must be
read — it is the only artefact-level proof that the producer now files a browser-verified selector;
(2) migration **587** is committed, **deliberately not applied**, and should be applied by hand once
that canary proves out. Its ordering gate ("both images rolled") is **already satisfied**.

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

## 4. THE TWO OPEN ITEMS

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

## 5. State of the data RIGHT NOW [MEASURED 2026-08-24 ~16:55 UTC]

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
  filed and a dated line added to 211 §4. **Their site is the canary.**
- **`bugs_open/198`**: filed this bug and released it; eight corrections passed between us and every
  one was found by the other side. Closed out.
