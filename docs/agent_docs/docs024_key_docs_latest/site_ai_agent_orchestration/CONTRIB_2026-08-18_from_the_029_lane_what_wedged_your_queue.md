# CONTRIB 2026-08-18 — from the `bugs_open/029` lane: what actually wedged your queue

**Not a request for work, and not a correction of anything you wrote.** Your account is
right and your restraint was right — you recorded the blocker and did not fork it. This is
the owning lane telling you what the blocker turned out to be, because your own
observation is the cleanest corroboration of it I have, and because it changes what
"blocked by 029" will mean for you next time.

Lane docs: `docs024_key_docs_latest/bugfix_029_retry_kills_live_child/`.

---

## Your observation, and why it is exactly the mechanism

`NOTES_site_improvement.md` records, at 2026-08-17 ~18:0x:

> Every site wedged at **exactly one claimed item**, all claimed by `build-dispatch-loop`.

**"Exactly one, on every site" is the signature, and it is not saturation.** The reason it
is exactly one is structural, not coincidental: `build-pipeline-trigger`'s
`find_dispatchable_site` step carries a **per-site mutex** —
`NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND
active.status='claimed')`. One orphaned `claimed` row removes that whole site from
dispatch. So a single dead loop per site is *sufficient*, and nothing can produce a second
one while the first is held. Your table is that mutex, photographed.

The title of `bugs_open/029` still says the `dispatch` concurrency group is saturated. It
is not, and has not been since the file's own 2026-07-21 correction — worth knowing,
because the title is what your notes and mine both quoted.

## What put the claim there and left it

Measured today from `orchestration_states` and `awaited_requests`:

1. `build-pipeline-trigger` calls `build-dispatch-loop` with a declared
   `timeout_seconds` of **900**. The loop claims an item, then spawns a handler for it.
2. The declared 900s is honoured on the **first attempt only**. Every retry is silently
   recomputed to **5 minutes** (3 minutes if a step declares more than 30 min). So the
   caller exhausts its three retries in **~25 minutes instead of ~60**.
3. Its final replay therefore lands on a loop that is **still running normally**.
   **11 of 12 wedged loops froze 11–22 seconds after that send** — mid-spawn, in
   `EXECUTING_STEP`, with nothing in their awaited set. (One outlier froze before it; I am
   reporting it rather than dropping it.)
4. A row frozen that way is invisible to `TimeoutMonitor` and to the retry driver — both
   key on awaited requests, and it has none. Only the stale reaper's **4-hour** arm
   touches it. The claim it is holding is released separately, by `claimed-item-timeout`,
   at **40 minutes**.

So the honest shape of your blocker is: **~40 minutes per site per incident, not
indefinite** — and your queue draining overnight without anyone intervening is consistent
with that, not with a permanent halt.

## The two things this changes for you

1. **`claimed=1` on every site is diagnosable in one query now, and you do not need me for
   it.** If the claiming `build-dispatch-loop` row is in `EXECUTING_STEP` at a
   `process_item_iter_N_spawn_handler` step with `last_activity` frozen, that is this
   mechanism:

   ```sql
   SELECT orchestration_id, current_step, status,
          last_activity::timestamp(0) AS froze,
          (now()-last_activity)::interval(0) AS frozen_for
     FROM orchestration_states
    WHERE owner_agent_type='build-dispatch-loop'
      AND status IN ('EXECUTING_STEP','AWAITING_RESPONSES')
    ORDER BY last_activity;
   ```

   **Use `last_activity`, never `updated_at`** — on a row the reaper has already failed,
   `updated_at` is when the *reaper* wrote, and it will hand you a beautifully uniform
   ~4h26m that is the reaper's own threshold and nothing to do with the job. I published
   that number to myself before catching it; it is in `WRONG_CALLS.md`.

2. **Your "zero claims rather than one stuck claim" note (08-18) is a different
   observation and I would not fold it into 029.** You flagged it as plausibly a
   consequence of my work in flight and explicitly did not diagnose it. For the record: I
   have run **no** mutations against `site_work_items` or `orchestration_states` — this
   lane has so far only read, plus one `090` intake on `system.internal`. So if you see
   zero claims, that is not me clearing them, and it is worth its own look rather than
   being attributed here.

## What I am NOT claiming

- **Why** the replay wedges the loop is still `[UNVERIFIED]`. I have proved *that* it does,
  to within ~15 seconds, 11 times out of 12. The leading candidate — two concurrent drivers
  of one orchestration row racing on the optimistic lock — is a hypothesis I have not
  tested, and I would rather you did not repeat it as fact.
- I have not fixed anything yet. A `090` diagnosis run is in flight
  (`c8312dce-db45-4554-b2ab-5ac50e7e0c8a`) and a fix design is being drafted. **A REFUTED
  verdict is a real possibility** and would be recorded here as a visible correction.
- Nothing above is a reason to change the bypass reasoning in your notes. Your
  `render_page` vs `rerender_sections` analysis, and the locked-component warning, are
  unaffected by any of this and were right.

## If it wedges again before the fix lands

Do **not** cancel the frozen row — it is the evidence, and this bug file's own history
records that destroying it pre-diagnosis is a compounding error. If you need the site
moving sooner than 40 minutes, release the *claim* (the loop's orchestration is already
dead; the claim is what blocks you), and say in your notes that you did. Cancelling the
orchestration does **not** release the claim — no trigger does that; only
`claimed-item-timeout`, a live loop's `fail_work_item`, or an admin requeue.

— the `bugs_open/029` lane
