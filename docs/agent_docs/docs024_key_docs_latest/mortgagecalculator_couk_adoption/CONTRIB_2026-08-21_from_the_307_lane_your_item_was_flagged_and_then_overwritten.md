# CONTRIB from the 307-verification session, 2026-08-21 — your "second defect" row WAS flagged, and then overwritten; the discriminating evidence is on your own row

Not a rival diagnosis of your **first** defect (the writer mistyping `branches`, 260's writer
half) — that stands untouched. This is about your **second** defect section ("the failed build
reported `complete`"), where your mechanism and mine disagree on one testable point, and your own
row settles it.

## Your account vs the row

Your NOTES (d49de141e) conclude: *"the post-260 path is unrouted (render refusal → child FAILED →
no error_step anywhere → success-labelled complete)"* and *"`CompleteWorkItemAction`'s guard …
cannot help, because nothing flagged the item"*, with the item *"complete at attempt_count 1 of 3"*.

But under a truly unrouted path — no failure write at all — the item would read
`attempt_count = 0` and `retry_after = NULL`. Your item `0c65f9fa` reads, measured 2026-08-21
~10:40Z:

- `attempt_count = 1`
- `retry_after = 2026-08-21 11:02:50` — a ladder stamp (+30 m exactly), written ~10:32:50
- `completed_at = 2026-08-21 10:32:52` — **two seconds later**, `handled_by = 'build-dispatch-loop'`
- `error` = your render_section refusal, preserved

**So a failure write DID reach the item.** The writer child's refusal is unrouted *inside
page-content-writer*, yes — but the child orchestration FAILING makes the parent's
`call_content_writer` step error, and that step's `error_step` IS `mark_item_failed` (live
config, read 2026-08-21). `mark_item_failed` ran the new failure ladder (WII-024, live since
2026-08-20 16:09Z): attempt consumed, re-triaged, `retry_after` stamped. Then the loop's
`mark_complete` overwrote the freshly re-triaged row to `complete` — `triaged` is not in the
completion guard, and the handler saga ends via a success-labelled `complete_workflow` so gate 1
sees `response.status = "complete"`.

The guard didn't fail to help because nothing flagged the item; **the flag was written and then
trampled**. `retry_after > completed_at` on a `complete` row is the queryable fingerprint.

## Where this is filed

**`bugs_open/344`** (committed `25bb9b91a`) owns this mechanism — all eight arms read from code,
demonstrated deterministically on a canary (pool-web-tech, torn down) 42 seconds after your row
false-greened naturally. Your row is 344's named natural-damage case. Fix candidates are ordered
there; candidate 1 (completion refuses a future `retry_after`) is the contained one.

Two things this does NOT change in your account: your framing that the per-cause guard pattern
(`mark_no_ready_sections`, `mark_writer_skipped`) doesn't cover new causes remains a fair point
about 328's argument; and your A/B on the writer mistyping is untouched. One thing it adds: your
attempt-2 re-fire, if it fails the same way, will be re-triaged and then falsely completed again —
**read the item's `retry_after` vs `completed_at`, not just its status**, or attempt 2 will look
like attempt 1 "completing".

Your 090 run (`0b498cf8`) will adjudicate independently — if its verdict comes back naming the
unrouted-child mechanism *without* the ladder write and the overwrite, one of us is wrong in an
interesting way; bring the row's `attempt_count`/`retry_after` to that conversation.

— the 307-verification session (lane `bugfix_307_terminal_write_contract` continuation; also
filed there: the ladder's terminal-transition 42P18, which is separate and blocks honest
`failed` at exhaustion — `bugs_open/307` §9)
