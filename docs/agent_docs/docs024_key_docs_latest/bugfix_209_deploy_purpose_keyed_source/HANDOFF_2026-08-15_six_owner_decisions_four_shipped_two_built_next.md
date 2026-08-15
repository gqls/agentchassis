# HANDOFF — 2026-08-15. Six owner decisions: **four shipped + approved, the TTL flip LIVE on v1.0.1301**, two pieces of build work left. COLD-START HERE.

Supersedes `HANDOFF_2026-08-14c_approved_and_live_on_1299.md`. Owner prose: `README_where_we_are.md`.
Evidence: `bugs_open/231` §POST-ROLL. Missteps: `NOTES_209…` and **three `WRONG_CALLS.md` entries dated
2026-08-15 — read those before submitting anything to the council** (§7).

## 1. What the owner ruled, and where each decision stands

The owner ruled on all six open questions from RFC_028 + the caching measurement on 2026-08-15.

| # | decision | state |
|---|---|---|
| D1 | the resolver gets an owner + one stated contract | **SHIPPED + APPROVED** (`260cb2393`, corr `5d491545`) |
| D2 | dot-discriminator expressed once in code | **SHIPPED + APPROVED**, same commit |
| D3 | an arm budget, "reasonably generous" | **SHIPPED + APPROVED**, same commit — budget 15, 10 used |
| D4 | alias collision: add the check, new name wins | **SHIPPED + APPROVED**, same commit + `1383d22cb` |
| D5 | use the existing scheduler, not a new CronJob | **DESIGN SETTLED, NOT BUILT** — §4 |
| D6 | add caching where we can | **TTL FLIP LIVE**; markers **NOT DONE** — §5 |

## 2. The one thing genuinely owed right now: finish the post-roll gate

`v1.0.1301` carries the cacheTTL flip. Verified 2026-08-15 10:42Z:

- **Stamp** `0115f2b45`, read from a pod whose log reached back to startup (`logs <pod> | head -1` showed the
  startup line — that discriminator is mandatory, the stamp scrolls within ~90s on busy pods).
- **Ancestry, two-sided:** `c5010ac26` (the flip) and `260cb2393` (D1–D4) are **ancestors**; `8bbedf5db` and
  `39aadb590` (both committed after the stamp) are **correctly absent**.
  > ⚠ **I got the control wrong first time** — I picked `0d1687108` as a must-be-absent control when it
  > *precedes* the stamp, so it read as a failed control. A control must POSTDATE the build. If you see
  > "CONTROL FAILED", check the commit order before concluding anything about the build.
- **No 400 spike:** zero failed calls in every hour of the last four (`llm_call_log`), which is the failure
  mode the whole change was gated on.
- **Caching works through the new binary:** one post-roll `council-gate` call, **119,721 cache reads, 0 writes,
  0 failures**.

### STILL OWED — and it is time-gated, not effort-gated

**Nothing so far proves the 1-hour TTL is actually in effect.** A cache read at a sub-5-minute gap is equally
explained by the old 5-minute behaviour. The artefact-level proof is **a cache hit at a gap > 5 minutes**, which
only appears with real traffic over time. Run this once a few hours of traffic have accumulated:

```sql
-- A non-zero count here is only possible with the 1h TTL. Under the old constant it is
-- structurally zero, so this is a disconfirmable check rather than a reassuring one.
WITH c AS (
  SELECT agent_type, created_at, cache_read_input_tokens,
         lag(created_at) OVER (PARTITION BY agent_type, md5(left(prompt_rendered,4000)) ORDER BY created_at) AS prev
  FROM llm_call_log
  WHERE created_at > '2026-08-15 10:41:00+00' AND prompt_rendered IS NOT NULL
)
SELECT agent_type, count(*) AS hits_beyond_5min, max(created_at - prev) AS widest_gap
FROM c WHERE prev IS NOT NULL AND created_at - prev > interval '5 minutes'
  AND coalesce(cache_read_input_tokens,0) > 0
GROUP BY 1;
```

Also still owed from the approved plan:
- **The roll was MIXED at check time** — 1 pod on `v1.0.1300`, 19 on `v1.0.1301`, across 10 services. Re-confirm
  a single tag fleet-wide before calling the gate done. **Enumerate by IMAGE, never `-l app=agent-chassis`**
  (that selector returns 2 pods of ~20):
  ```bash
  kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}{end}' | grep 'agent-chassis:'
  ```
- **Probe from a non-`agent-chassis` service** running the same binary (`vet-intel`, `business-intel`,
  `agent-page-content-writer` …) — the gate explicitly is not single-pod.
- **If the >5min-gap count stays zero over sustained traffic, or 400s appear:** revert `cacheTTL` to `""` in
  `platform/aiservice/anthropic.go`. The field-handling code is correct for both values; nothing else changes.

## 3. What was approved, and the honest story of how

Two councils, both **APPROVED**, both after the code was already correct:

- **RFC_028 D1–D4** — corr `5d491545`, 9 reviewers, 4 advisory objections. `editquality` found a real gap:
  `classifyDefaultedConfigValue` has **five** ways a canonical value can fail and I had tested **two**. All five
  now covered (`1383d22cb`). `guardian`'s coupling objection did not survive (the audit tool already imported
  `datahelpers`). `architecture` recorded `needs_rfc: insufficient` — the budget caps arm growth but does not
  answer why five documents of one contract drifted apart. **That stays open in RFC_028.**
- **cacheTTL 1h** — corr `176d921e`, **approved on round 3, "all reviewers approve", 12 clean approvals.**
  **The code was byte-identical across all three rounds.** Rounds 1 and 2 were gated by `prior_art_librarian`
  for the same reason both times: the blast-radius facts were narrated in the rationale and attached in
  `grounded_in` as nothing. See §7 — this is the single most important thing to carry forward.

## 4. D5 — the scheduler answer (design settled, build not started)

**There are two schedulers and the obvious one is the wrong one.** The six existing config-integrity checks
(`removed-config-keys-check`, `single-owner-carriers-check`, `optional-key-budget-check`, …) are k8s CronJobs;
four of those run on the stock `postgres:16-alpine` image because they are pure SQL. **This check cannot be
one of them** — its verdicts come from Go (`spec.Defaults` are compile-time `RegisterActionInputSpec`
declarations), which is why `scripts/audit-default-shadowed-keys.sh` pipes psql output into `go run`.

**Use `scheduled_tasks` (the platform scheduler), following four existing precedents** — `diagnose_silent_check`,
`diagnose_dormant_agents`, `revalidate_review_queue`, `reconcile_superseded_reviews` are all chassis actions
driven by `scheduled_tasks` rows. The payoff: **no new image, ever.** The check lives in the binary that already
holds the specs it checks, and scheduling is one DB row — which also sidesteps the ImagePullBackOff-reads-as-
RUNNING trap entirely.

Build shape:
1. New action wrapping the logic in `cmd/config-key-audit/defaultshadow.go` (`findDefaultShadowedKeys` +
   `--report`'s doc_notes write). **`cmd/config-key-audit/main.go` is CONTENDED** — another lane has 13 lines
   and untracked `optionalbudget*.go` there; put shared logic somewhere both can import rather than editing it.
2. Register it; seed an agent definition.
3. One `scheduled_tasks` row, `interval_seconds` daily, in the config-integrity window (06:20/06:25/06:35 are
   taken — 06:30 is free).
4. **Write one `doc_notes` row per run, clean or not** — a missing row must mean "the job did not run", which is
   different from "nothing is wrong". `optional-key-budget-check/base/check.py` states this rule and is worth
   reading in full before building; it is the closest sibling.

## 5. D6 — markers (NOT started), and the ordering constraint that bites

Measured 2026-08-15, three days of `llm_call_log`. **Break-even is a ~22% hit rate at the 5-minute TTL and
~53% at 1 hour** (write costs 1.25× base input at 5m, 2× at 1h; read 0.1×).

| agent | hit @5m | hit @1h | shared prefix | do what |
|---|---|---|---|---|
| `content-gap-planner` | 1.0% | **99.8%** | ~10k of 14.8k chars (**67%**) | **the prize — add the marker NOW that 1h is live** |
| `diagnose-agent` | 64.4% | 93.2% | ~10k of 75.7k (13%) | worth adding; was safe even at 5m |
| `page-content-writer` | 8.9% | 58.5% | ~4k-char **tail** | **DO NOT** — §6 |

- All four models involved have a **1024-token** minimum cacheable prefix (`sonnet-4-6`, `sonnet-5`). The minimum
  is model-dependent and **non-monotonic** — 512 on Opus 5, but **4096** on Opus 4.6/4.5 and Haiku 4.5. Re-check
  if any agent's model changes.
- **The marker is `<!--CACHE_BREAKPOINT-->` placed in the prompt template in `agent_definitions`** — DB config,
  **live immediately**, no roll needed. Place it at the shared/varying boundary.
- **⚠ A byte above the marker costs money and returns nothing, and it looks exactly like success.** After adding
  a marker, assert **non-zero `cache_read_input_tokens` on the 2nd+ call**. A zero is the failure mode, not the
  absence of one (register LCO-008's own landmine).
- Current adoption: **1 of 191 live agents** (council-gate only), **17 seats marked, 1 distinct shared prefix** —
  which is also the health check proving migration 377 has not been reverted by `099_SYNC_gate_roster.py --apply`.

## 6. `page-content-writer` — the largest spender, and why caching cannot help it

4.48M full-rate input tokens over three days, **zero cached** — the biggest single consumer. It is **not** a
caching candidate and adding a marker would make it worse:

- **208 distinct prefixes at 2,000 chars across 347 calls** — the prompts diverge almost immediately.
- Its shared block is at the **END**: `right(prompt,2000)` → **2 distinct**, `right(prompt,4000)` → 2 distinct,
  `right(prompt,5000)` → 33. So ~4,000 shared chars, in the wrong position (prefix caching matches from the front).
- Even reordered, ~4,000 chars ≈ **~1,000 tokens**, at or below the 1024-token floor for its model.

Making it cacheable is a **prompt-template restructure** (hoist the invariant instructions to the front, as
migration 377 did for council-gate), not a caching change. **Not costed. This is the largest remaining
cost-reduction opportunity in the fleet and nobody owns it.**

## 7. ⚠ READ THIS BEFORE SUBMITTING ANYTHING TO THE COUNCIL

**Three `WRONG_CALLS.md` entries on 2026-08-15, all one class: a figure quoted without its evidence attached.**

1. RFC_028 round 1 — the "27 council rounds" figure, unqueried. The **same seat had objected to the same number
   one submission earlier**, and it was recorded in this lane's own bug file.
2. cacheTTL round 1 — "marker adoption is council-gate only", asserted, never queried.
3. cacheTTL round 2 — the queries that fixed (2), annotated **"(query in grounded_in)"** and absent from
   `grounded_in`. **I noticed while assembling it and submitted anyway.**

Cost: two REVISE rounds on a change whose code never altered. **The rationale is where a submitter talks;
`grounded_in` is where a reviewer checks.** The mechanical check, which is not "try harder":

```bash
python3 -c "import json;d=json.load(open('sub.json'));print('\n---\n'.join(d['plan']['grounded_in']))"
# read that as a reviewer who has never seen the rationale. Every number and proper noun in
# the rationale must appear there with a runnable query beside it.
```

If this recurs a fourth time, the fix is probably to make `097_TRIGGER…` refuse a submission whose rationale
contains figures absent from `grounded_in`.

## 8. Other traps this lane hit today

- **`git archive HEAD | tar -x` into the shared scratchpad** — I left two 423M trees there; the council's
  `debug_historian` caught it. Removed, 846M reclaimed. Clean up after yourself; the scratchpad is shared.
- **The working tree does not compile** — another session's untracked
  `platform/orchestration/actions/work_item_retraction.go` redeclares `retractionCandidate`. Committed HEAD is
  fine. Build/test against `git archive HEAD` plus your own files (and then delete the extraction).
- **A trigger script can exit 0 having submitted nothing.** My first cacheTTL resubmit did. **Distinguish
  "never submitted" from "queued and slow" by artifact, not by orchestration row:** `persist_submission` writes
  the `fix_plan` artifact *before* the run starts, so `SELECT kind, created_at FROM diagnosis_artifacts WHERE
  correlation_id=…` tells you which. CLAUDE.md's "a missing orchestration row is latency, don't retry" is right
  for the normal case and would have been wrong here.
- **The register can contradict itself.** LCO-008's `what:` line claimed `ttl 1h` from 08-10 to 08-15 while
  production sent no ttl at all; the council's removal *was* recorded — nine lines lower, under the verdict
  trail. New LANDMINE added. **The top of a long register entry is the oldest text in it.**
- **Cluster was slow** (queries timing out at 100s+, dispatch queue items ~15 min old). Background long queries.

## 9. Not this lane's

- `3ba384c63` (another lane) carries a `Council-Submitted:` trailer whose correlation came back **rejected**, and
  the code is on shared HEAD. Honest trailer, but somebody should look at it.
- 209 Phase 3, 236, and the 96 `dotted_conditional` entries — unchanged, unowned.
