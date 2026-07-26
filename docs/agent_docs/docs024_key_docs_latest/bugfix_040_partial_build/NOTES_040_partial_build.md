# NOTES — bugs_closed/040-partial-build

Running technical record. **Append-only, newest at the bottom.** Missteps are the
point, not an appendix — record what was tried, what the system actually said, and
every wrong turn including my own earlier claims in this file.

Case file: `bugs_closed/040_HANDOFF_2026-07-20_failed_page_build_leaves_page_deployed_and_partially_composed.md`
Entry point: `HANDOFF_2026-07-26_continue_here.md` (same directory).

---

## 2026-07-26 21:30–22:00 UTC — candidate 2 behavioural verification

Picked up from the handoff, whose §4 described an **unsolved dispatch problem**: five
publishes of the scratch probe had produced zero `orchestration_states` rows, and it
listed three hypotheses to work through, starting with a column-by-column diff of the
`scratch-cand2-probe` agent row against the known-working `council-gate` row.

### The dispatch problem was already solved by §5 of the same document

Before spending anything on the hypotheses I checked the probe's **current** state — the
cheapest possible query, and the one that turned out to answer everything:

```sql
SELECT item_key, status, attempt_count, COALESCE(NULLIF(error,''),'<<BLANK>>'), updated_at
FROM site_work_items WHERE item_type='scratch_cand2_probe' ORDER BY item_key;
-- scratch_cand2:a | failed   | 2 | step boom failed: failed to execute action
--                                  update_work_item_status: work_item_id not found at
--                                  input_data.does_not_exist and skip_if_missing=false
--                                | 2026-07-26 21:16:16+00
-- scratch_cand2:c | complete | 2 | <<BLANK>>
--                                | 2026-07-26 21:16:16+00
```

**The probe had run — twice — at 21:16:09 and 21:16:16 UTC, one minute after the handoff
was written at 21:15.** Two `orchestration_states` rows, both `COMPLETED` at step `done`:
`2fe04703-df0d-4cd3-bac2-90f5eca44dce` and `a0257bd3-8a6b-4709-9798-54194fa8b701`.

There was never a dispatch defect. The cause is documented in **§5 of the handoff
itself**: the `generic-requests-group` consumer sat at a frozen committed offset of
**105196** for ~25 minutes, and the handoff's own backlog dump shows the probe messages
queued at offsets **105197, 105202 and 105204** — i.e. immediately behind the stall. When
the lane cleared they were consumed and ran normally. The document diagnosed the stall
and listed the stuck offsets, but did not connect them to its own "unsolved problem" two
sections earlier.

> **Lesson, and it is the cheap check that would have saved five attempts and three
> hypotheses:** when a dispatch appears dropped, re-read the *outcome* table before
> theorising about the *cause*. A queue that is merely stalled looks exactly like a
> dispatch that was eaten, and the handoff had already proved the stall.
> Logged in `WRONG_CALLS.md`.

### Owed item 1 — the fix firing: VERIFIED LIVE

`scratch_cand2:a` ended `failed` with
`step boom failed: failed to execute action update_work_item_status: work_item_id not
found at input_data.does_not_exist and skip_if_missing=false`.

Pre-fix this column would have been blank: `mark_failed` configures **no**
`error_message` literal, and the literal was the only thing that ever wrote it.

### Owed item 3 — the `complete` exclusion holding: VERIFIED LIVE

`scratch_cand2:c` ended `complete` with `error` still **BLANK**, and it ran *after*
`boom` had set `__step_error` — which is never cleared. This is the load-bearing negative
control: without the `newStatus != "complete"` guard it would have inherited a stale
failure. The fleet census found exactly one literal-less `complete` step in production
(`image-build-handler`), so this is a live path, not a hypothetical.

### MISSTEP — I claimed the prefix branch had fired, then the evidence refuted it

Reading `step boom failed: …` on the item, I wrote that the `step X failed: ` prefix
sub-branch had fired and was therefore now verified live, contradicting the handoff's
`[UNVERIFIED LIVE]`. **That was wrong.** I inferred it from the *output* instead of
reading the *input*. The stored `__step_error` says:

```sql
SELECT collected_data->'__step_error'->>'failed_step',
       collected_data->'__step_error'->>'message'
FROM orchestration_states WHERE orchestration_id='a0257bd3-8a6b-4709-9798-54194fa8b701';
-- failed_step | boom
-- message     | step boom failed: failed to execute action update_work_item_status: ...
```

The routed message **already** carried the prefix, so
`!strings.HasPrefix(errorMessage, "step ")` was false and the prefix branch was correctly
**skipped**. The fallback copied the message verbatim.

So the handoff's `[UNVERIFIED LIVE]` on the prefix branch **stands**, and this run is
positive evidence for the code comment's claim that action errors already carry the
prefix — only the awaited-request-timeout shape (`Request <id> timed out after 3
retries`) is bare, and that still needs a real `call_agent` timeout to exercise.

Cheap check that caught it: read `__step_error` in `orchestration_states` rather than
inferring the branch from `site_work_items.error`. **The output of a
prefix-if-absent branch is indistinguishable from the output of no branch at all.**

### Deployment re-confirmed on the current pod

The pod rolled again (v1.0.1171, `agent-chassis-5b4456686c-s5fkc`, started
2026-07-26T21:02:56Z) after the handoff's grep, so I re-ran it with both controls in one
command — a removed/added-literal grep is unfalsifiable without a positive control
alongside it:

```
new-string (created by cand2)  : 1     # "no error_message literal"
positive control (040 guard)   : 1     # "build is short of its plan", live since v1.0.1146
negative control (must be 0)   : 0     # "candidate two placeholder xyzzy"
```

(The command exits 1 — that is the negative control's `grep -c` returning no match, which
is the intended result, not a failure.)

### Owed item 2 — the literal-wins control needed a NEW probe

The handoff's owed item 2 was "a `needs_human_review` park must still read its configured
literal". History alone cannot discharge it. The live rows that *do* carry the literal:

```sql
SELECT COALESCE(NULLIF(error,''),'<<BLANK>>'), COUNT(*), MIN(updated_at), MAX(updated_at)
FROM site_work_items WHERE status='needs_human_review' AND updated_at > now() - interval '3 days'
GROUP BY 1 ORDER BY 2 DESC;
-- 'page-build-handler no-op: no sections ready to build …'  | 4 | 07-25 01:35 | 07-25 19:29
```

All four **predate** the roll that made the fallback live (v1.0.1170, 2026-07-26 18:35Z).
They prove a literal was written when there was no fallback to compete with it — which is
not the question. The question is whether a configured literal still wins **now that a
fallback exists and `__step_error` is set**. That needs an induced run.

So I extended the harness rather than re-using history: added **PROBE B**
(`b0b0b0b0-1111-4222-8333-444444444444`) and a `mark_park` step —
`needs_human_review`, *with* a literal — spliced between `mark_failed` and
`mark_complete`, so it executes while `__step_error` is still set from `boom`.
All three items reset to `detected`/blank/`attempt_count=0` first so the assertion is
unambiguous (they were at 2 of 3 max_attempts).

Workflow now: `boom` –(error_step)→ `mark_failed` → `mark_park` → `mark_complete` → `done`.

### MISSTEP — the kcat container's entrypoint swallowed the shell

First fire attempt returned kcat's usage text and `-b <broker,..> missing`. Cause:
`kubectl run --image=edenhill/kcat:1.7.1 ... -- sh -c '…'` passes `sh -c …` as
**arguments to the image's kcat entrypoint**, not as a replacement command. Nothing was
published. Fix: add `--command`:

```bash
kubectl -n kafka run "kcat-cand2-$(date +%s)" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach --quiet --command -- \
  sh -c "printf '%s' '$PAYLOAD' | kcat -P -c 1 -b … -t … -H … && echo PUBLISH_OK"
```

This is a *distinct* trap from the known `kubectl run -i --rm | kcat -P` one (which exits
0 having published nothing). Both are avoided by the same discipline: **put the payload in
the container command and make the container print `PUBLISH_OK`** — a publish with no
positive confirmation is not evidence of a publish. Second attempt printed `PUBLISH_OK`.

### Owed item 2 — the literal wins: VERIFIED LIVE, 22:00Z

The extended probe ran (orchestration `e5907364-7dfa-44be-8f45-9ed85da32df4`, `COMPLETED`
at `done`). All three arms in one run:

```
PROBE A | failed             | step boom failed: failed to execute action update_work_item_status:
                               work_item_id not found at input_data.does_not_exist and skip_if_missing=false
PROBE B | needs_human_review | cand2 probe literal: this park reason was configured, not routed
PROBE C | complete           | <<BLANK>>
```

**B is the one this run was for.** It executed *after* `boom`, so `__step_error` was set —
the same `collected_data` shows `failed_step: boom` and the full routed message — and it
still recorded **its own literal**. The `errorMessage == ""` guard holds: a configured
literal is never overwritten by the fallback. A and C reproduced their earlier results.

The prefix branch was **again correctly skipped** — `__step_error.message` already started
`step boom failed: `, so `HasPrefix` was true. Two independent runs now agree that the
action-error shape arrives pre-prefixed, which is what the code comment claims. The
branch remains `[UNVERIFIED LIVE]` for the bare awaited-request-timeout shape only.

**All three owed items are now discharged.** Candidate 2 carries no verification debt.

### Ruled out on the way — the 175 blank `needs_human_review` parks are NOT this defect

The 3-day census showed 182 `needs_human_review` items, 175 with a blank `error` — 8×
the population candidate 2 was built for, and superficially the same
"triage cannot see why" symptom. It is not the same class:

```sql
SELECT source, item_type, COALESCE(result::text,'<<NULL>>'), COUNT(*)
FROM site_work_items WHERE status='needs_human_review' AND COALESCE(error,'')=''
  AND updated_at > now() - interval '3 days' GROUP BY 1,2,3 ORDER BY 4 DESC;
-- internal-link-resolver | unresolved_cta                | {} | 63
-- section-planner        | needs_section_data            | {} | 44
-- discovery              | required_fields_missing       | {} | 25
-- discovery              | cta_names_unknown_destination | {} | 22
-- reconcile_site_plan    | owned_page_review             | {} |  6
```

`result` is `{}` — the column default — and `update_work_item_status` **always** writes a
result payload carrying `completed_by_orchestration_id`/`completed_by_step`. So these
items were **created** as `needs_human_review` by detectors and never transitioned by that
action at all. Their reason lives in `summary`/`spec`; `error` is for a *failure* reason,
and a detection is not a failure. No bug, no residual — recorded so the next reader does
not re-open the same question.

### The case file's "the blank count must stop growing" is NOT dischargeable by census

Step 2 of "Verify once live" says to re-run the blank census and confirm the count stops
growing. It cannot do that job, for two independent reasons — recorded so nobody spends
time on it:

```sql
-- 2026-07-26 22:00Z, against the 2026-07-25 baseline in the case file
--            2026-07-25          2026-07-26
-- failed     75 / 21 blank       52 / 14 blank
-- cancelled  51 / 43 blank       52 / 43 blank
-- rejected    5 /  0 blank        5 /  0 blank

SELECT count(*), count(*) FILTER (WHERE COALESCE(error,'')='') FROM site_work_items
WHERE status='failed' AND updated_at > '2026-07-26 18:35:07+00';   -- the roll
-- 0 | 0
```

1. **The population shrinks.** `failed` went 75 → 52 items between the baseline and now,
   so the blank count fell 21 → 14 without a single new item being written. A count that
   goes *down* cannot demonstrate "stopped growing" — the denominator moved. Same trap as
   the retention-clock landmine in `bugs_open/003`: **record a rate, not a count**, and
   take the baseline at the moment you need it.
2. **There is no traffic to measure.** **Zero** `failed` items have been stamped by any
   real handler since the 18:35:07Z roll that made the fix live. The census is quiet
   because nothing has run, not because the fix is working. That is exactly what the
   handoff said and it is still true — which is precisely why the induced probe, not the
   census, is the thing that discharges this.

### Lane health during this session

`generic-requests-group` was healthy throughout, unlike the stall in the handoff's §5, but
**slow**: sampled twice 30 s apart the committed offset moved 105267 → 105270 → 105272
with lag 5–6, i.e. seconds per message on a single in-order partition. That is the known
`bugs_closed/030` residual shape (slow-but-progressing), *not* the frozen-offset stall
(§5). The discriminator is whether `CURRENT-OFFSET` moves between two samples — one
reading cannot tell "stalled" from "busy".
