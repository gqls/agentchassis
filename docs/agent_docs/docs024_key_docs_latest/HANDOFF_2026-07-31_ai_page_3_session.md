# SESSION HANDOFF — "AI page 3", 2026-07-30 → 07-31

Written 2026-07-31 ~08:30Z for a fresh chat. This session ran two lanes: the
**consolidation programme** (`features_open/024`) and the **gripper dossier**
(`robot_hands_gripper_dossier/`). Both are at a clean stopping point; nothing is
half-applied.

**Read these two first, not this file's history:**
- `consolidation/HANDOFF_2026-07-30_continue_here.md` — consolidation state (§1 rewritten, §4b = what to do on each branch of the owner's decision)
- `robot_hands_gripper_dossier/HANDOFF_RESUME_gripper_dossier.md` — gripper cold-start (switch positions corrected 07-31: **dispatch is ON**)

---

## 1. TWO THINGS WAIT ON THE OWNER — everything else is done or filed

**(a) Adoption sequencing (consolidation).** `platform/httpguard` + `platform/mailer` are
built, council-approved and have **zero importers**. Both remaining adopters live inside
`internal/tools-api`, which belongs to the **gauntlet lane** — the session titled
**`vonc 6`** (`c4daed6f-5514-49f1-be6a-7bbf6bbd3c98`). They **have** picked it up: read our
note 07-30 14:12Z, and at 15:16Z put it to the owner as item 3 of 4, recommending it
**before** their distribution leg. **Do not re-notify them and do not apply the patch
yourself** — `bugs_open/083` (by slug) is open against that service. The decision is which
order the owner wants; the recommendation given was "limiter fix first", because
`client_ip_hash` is the only column in `gauntlet_rounds` that can distinguish one visitor
from another and it has held **one value in 83 of 83 rows**, so it is the instrumentation for
the distribution experiment rather than hygiene.

**(b) Fixture cleanup (gripper).** The owner has now seen the two 07-27 pages. Cleanup is
owed and covers **four** things: the `manual-test` work items, the two 07-27 pages, fixture
3's `…edd863e8….json`, and fixture 4's page once seen. **Do not clean up unasked.**

## 2. WHAT IS LIVE AND VERIFIED (not inherited — checked 07-31 08:07Z)

**Chassis `v1.0.1213`, both pods.** The report-chart fix (`f8e7c31ce`) is in the running
binary:

| marker | count |
|---|---|
| `estTextWidth` — a symbol the fix ADDED | 1 |
| `Capacity headroom against your requirement` — positive control | 1 |
| `nonexistent_marker_xyz` — negative control | 0 |

**Keep this marker lesson:** my other candidate marker was a *comment* and greps **0** —
comments do not survive compilation. A comment used as a deploy marker manufactures a false
negative that reads exactly like "the fix did not ship".

**Council `60d05267-a671-4b98-9b87-6a97e16d78a0`: APPROVED round 1** (2 advisory objections,
none high; `architecture` → `point_fix`). All four checkable items discharged — see the
gripper NOTES for the evidence. Trailer is on the docs commit `e640a2529`; the code shipped
in `f8e7c31ce`, which predates the verdict and cannot be amended under forward-only.

**`report-dispatch` is ENABLED** (owner instruction, 07-30 22:13Z). `report-request-pull`
stays off. An idle ON is free: the task's `pre_query` ends `HAVING count(*) > 0`, so with an
empty queue the scheduler logs *"Pre-query found no rows — task ran with nothing to do"* and
publishes no message. **Do not "fix" that query, and never read it through
`left(pre_query,120)`** — that cuts the `HAVING` off and makes it look unfiltered
(`WRONG_CALLS.md`, 07-30, is me making exactly that mistake).

## 3. IN FLIGHT RIGHT NOW — FIXTURE 4

Work item `4ccc73d7-c467-480f-9a39-0b327b383870`, `request_id`
`bf3765d6-befe-43a8-b1cd-ca5c210f39e9`, re-running **fixture 1's exact spec** so the chart
before/after is the same inputs. Expected page
`https://robot-hands.com/reports/bf3765d6-befe-43a8-b1cd-ca5c210f39e9.html`.

At the time of writing: **`reporting`, attempt 0/2** — claimed and building. For scale,
fixture 1 took **27 minutes** on 07-27, so `reporting` is not a hang.

```sql
SELECT status, attempt_count||'/'||max_attempts AS attempts, left(coalesce(error,'-'),120)
FROM site_work_items WHERE id='4ccc73d7-c467-480f-9a39-0b327b383870';
```
```bash
curl -s -o /dev/null -w '%{http_code} %{size_download}\n' \
  https://robot-hands.com/reports/bf3765d6-befe-43a8-b1cd-ca5c210f39e9.html
```

**If it FAILED again with a `verify_prose` violation, that is `bugs_open/160` and not a new
problem** — read §4. If it failed some other way, the cause is in the **child**
orchestration's `collected_data->'__step_error'`, never the work item's `error` column, which
carries only the wrapper `gripper dossier build failed — see the step error`.

**If it COMPLETED:** the page is the first artefact rendered by the fixed chart code. Look at
it — the whole point of the fix was that no automated check could see the defect. Check three
things: value labels beside capped bars are **whole** (`6.42× (Insufficient data)`, not
`6.42× (Insufficien`), the two reference captions are on **separate lines**, and capped bars
end in a **point**. Then tell the owner and hold for the cleanup decision.

## 4. THE BLOCKER TO SEEING IT ON A FRESH PAGE — `bugs_open/160` (filed this session)

The first fixture-4 attempt produced **no page** because `verify_prose` rejected the writer's
summary:

> `summary_html names model-like token "IP54-or-better" not in the candidate set or fact block`

`modelNumberRe` (`verify_report_prose_action.go:241`) classifies `IP54-or-better` as
SKU-shaped, and the clearance test is verbatim containment in the fact block — which has
`IP54`, not the composed phrase. **The gate is right and must stay strict**; its classifier
cannot tell a fabricated sibling SKU from a fact-block token recombined into English. The
step is fail-closed, so the whole report dies.

**Worst property: it is intermittent.** The identical spec passed this gate on 07-27, because
the trigger is the writer's phrasing. Expect it to read as a flaky pipeline rather than a
rule. Four fix candidates, ordered, are in the bug file — with the hole to close named first
(a prefix rule would clear an invented sibling, which is exactly what the guard exists to
stop). **Verification needs both halves**: the legitimate phrase clearing *and* an invented
sibling still rejected.

## 5. OWED, in the order they will bite

1. ~~**`016b` §9 entry for 160's pattern**~~ **DONE 07-31** — *"A strictness gate's CLASSIFIER
   rejects a legitimate recombination, and fail-closed then destroys the artefact instead of
   degrading it"*, filed in §9 with the both-halves verification rule and the note that
   fail-closed itself was established by a council HIGH and must not be relaxed to a warning.
2. ~~**Council debt (consolidation):** the contracts-gap claim, single-sourced~~
   **DISCHARGED 07-31 — CONFIRMED, and it must be cited in its CORRECTED form.** Measured:
   `input_contract`'s only runtime read is `call_agent.go:988–1011`, gated behind a step
   declaring an input mapping; the `OutputContract` type has **zero references** in
   `platform/`/`internal/`/`cmd/`; and **0 of 184** active agents declare either contract.
   Three sharpenings that change what it may be used to argue: (a) say *"no **runtime**
   reader"* — the offline `workflow_validator` does read `output_contract`; (b) that validator
   is **wired into nothing** — no makefile target, no hook, no CI — and exists as two
   near-identical copies; (c) therefore cite this as *"the mechanism exists and is inert"* and
   **never** as *"contracts are enforced narrowly"*, because there is no narrow enforcement,
   there is none. Full working in the gripper NOTES, 07-31.
3. ~~**`[UNMEASURED]`:** whether bare `.(string)` assertions on LLM-parsed maps recur~~
   **MEASURED 07-31 — REFUTED as stated, no bug filed.** 1,734 `.(string)` occurrences; 40
   survive a two-value filter; of those, 12 are the safe `x, _ =` form, 4 are comments (**two
   documenting a bare assertion already REMOVED**), leaving **24** genuinely panicking. Of the
   24: **10** are `r.Context().Value("user_id").(string)` in `auth-service/project/handlers.go`
   (middleware-injected, not LLM), ~10 read startup/DB/step config, and only **2** plausibly
   touch model output — `v3_site_actions.go:3424` and `:3721`, both inside `zap.String(...)`,
   so a panic would occur while logging. **Residue, a different risk class and not what was
   asked:** those 10 auth-service assertions panic the handler if middleware is ever reordered
   or bypassed on that route. Real fragility; **not filed on the strength of a grep.** Method
   limits and the two filter mistakes are in the gripper NOTES, 07-31.
4. **Structurally unclosable, stated rather than implied:** `httpguard`'s peer-gate reversion
   needs a connection from a genuine *public* peer. Every address a dev box can bind is
   loopback or RFC1918. **Do not let a future thread "close" it with another unit test.**

## 6. LANDMINES EARNED THIS SESSION (all filed with footprints in `LANDMINES.md`, synced to `doc_notes`)

- **An SVG viewBox CLIPS**, so a label that does not fit reads as corrupted *content*, not as
  a layout bug — in a report of computed figures the natural diagnosis is a scoring bug.
  Includes the check that found it: render it and **look**. ⚠ chromium here is a **snap** —
  it cannot screenshot into `/tmp/claude-*` or any dot-directory under `$HOME`; use a plain
  `~/dir` or it fails in a way that reads as "tool unavailable".
- **A passing mutation test may mean a SECOND guard absorbed the mutation**, not that the
  guard is redundant. Worked case: reverting the chart's computed gutter left "nothing is
  clipped" green because a truncation fallback silently shortened the label. Assert the
  outcome the caller asked for, not the absence of the symptom.
- **A hand-made `report_request` needs `handler_agent` on the ROW.** The loop claims it
  correctly, then dies at `spawn_handler` with `agent_type is required`, naming neither the
  row nor the column. Diff your row against a working one over **all** columns with
  `to_jsonb(w)` + `jsonb_object_keys` instead of reading the error.
- **Prose in another lane's live HANDOFF carries no timestamp** — `git log -S'<line>'` before
  quoting it as their position. A stale line in a cold-start path is indistinguishable from a
  considered decision.
- **A quiet `git log` is not silence** — a lane that reads, decides and reports to the owner
  commits nothing. The artefact is their transcript (`~/.claude/projects/<proj>/<id>.jsonl`;
  `customTitle` on line 1).

## 7. WRONG CALLS THIS SESSION (both in `WRONG_CALLS.md`)

- **"The patch was not delivered."** Measured uptake by commits and their directory; the
  owning lane had read it, decided, and reported to the owner — producing neither.
- **"`report-dispatch`'s `pre_query` has no status filter."** Reasoned from a
  `left(pre_query,120)` truncation I applied myself; the `HAVING` that refutes it sits at
  ~char 250. **If you truncate a field for display you may quote what you SAW, not what it
  MEANS.**

Both were caught by evidence contradicting a prediction I had already stated. That is the
cheap place to be wrong; the expensive one is a handoff.

## 8. COMMITS THIS SESSION

`1f780faba` `a0a22cfbf` `31c684124` `171ff677c` `c1e583267` `df7f918b8` (consolidation:
139 refutation, owner ruling, the `FrontEnd` seam + PUB-002/003, adoption patch, verdict,
peer-gate discharge) · `bd4bb06d6` `d736ce6ff` `99f9ff531` (handoff/summary + the delivery
finding, sharpened) · `8f78d96a7` (WRONG_CALLS) · `f8e7c31ce` (the chart fix + geometric
tests) · `6c5650146` `e640a2529` (landmines + gripper docs, council trailer) ·
`8f8e9d7be` (`bugs_open/160`).
