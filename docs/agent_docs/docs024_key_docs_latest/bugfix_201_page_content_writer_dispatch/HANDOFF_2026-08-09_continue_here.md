# HANDOFF — `bugs_open/201` lane, 2026-08-09 · **start here.** Supersedes `HANDOFF_2026-08-08_continue_here.md`

## 1. Headline: this lane's original scope is CLOSED

**`bugs_open/201`'s two symptoms were fixed and proven live before 08-08** (see the superseded
file). **`RFC_017`** — the fail-open→fail-closed flip that symptom 2's work surfaced — was
measured, ruled on by the owner, built, council-approved, deployed, and as of **today its own
last open thread (behavioural proof) is closed too.** Nothing from this lane's original scope
remains unaddressed. If you were sent here to continue `201`, **there is nothing left to do on
201 itself** — read §2 for what fired today, then §4 for what spun off and is genuinely still
open (but not this lane's).

## 2. What happened today (2026-08-09), in order

1. **A fresh chassis build was deployed**, `v1.0.1276` (up from `v1.0.1268`). Checked at the
   pod, both replicas: `docker.io/aqls/agent-chassis:v1.0.1276`, RFC_017 marker
   (`fails closed (RFC_017)`) present, count 1 each. `[MEASURED]`
2. **RFC_017's fail-closed branch fired in production for the first time, unprompted.** Re-ran
   the 08-08 handoff's own proof query and got a third row:
   ```sql
   SELECT status, attempt_count, left(error,90) AS err, result->'_verification', updated_at
   FROM site_work_items WHERE result->'_verification'->>'status'='error'
   ORDER BY updated_at DESC;
   ```
   `dartsonline.com`, `empty_section:552bb99e-…:category-listing`, page `index`, item **created
   2026-07-14** (an old stuck item), re-triaged `09:02:52Z` by `bugs_open/230`'s new discovery
   rotation, burned all `max_attempts=3`, landed `status='failed'`, `error` beginning
   *"completion blocked: verification could not run, and this item type fails closed
   (RFC_017)"*, `_verification.fail_open=false`. Every element of the original proof
   requirement is met. **Full record:** lane `NOTES` 2026-08-09 entry; register `WII-011`
   (status line + its own "still unproven" bullet, both corrected in place).
3. **This is also the first live instance of the named cost**: 3 rebuilds burned before a
   human sees it — the mechanism working exactly as RFC_017 intended, at the cost the owner
   knowingly accepted over building option 3 (park/`Indeterminate`). One data point.
4. **Checked ownership before writing anything about the item**: `who-owns.py 083` — the
   number is ambiguous (two unrelated cases share it); one is an active lane
   (*"detected findings never reach a handler"*) already updated today by
   `bugfix_230_discovery_driver`. **This dartsonline item is that bug's SITE, not its
   MECHANISM** — it reached `page-build-handler` three times and was correctly refused, which
   is a different failure shape than "never reached a handler". Not acted on, not claimed.

## 3. This lane's two spin-off bugs — status, so you don't re-derive it

Both came out of chasing `201`'s original handoff §4 ("two live pages still serving empty
sections"), which itself needed a correction — logged in `WRONG_CALLS.md` 2026-08-08 (a
`item_key` grep on the wrong slot spelling gave a false "never detected" reading).

- **`bugs_open/230`** — site discovery had no recurring driver (every `scheduled_tasks` row
  targeting the 3 discovery agents was a disabled one-off). **Filed by this lane 2026-08-09,
  picked up same day by `bugfix_230_discovery_driver`, FIXED + LIVE + VERIFIED by that lane's
  own thread.** Migration `346` (rotation stamp table + 3 hourly observe-only tasks) + a daily
  watchdog CronJob, council APPROVED round 1. The file's own §6 proof criterion fired
  unprompted at `13:52:04Z` — a `featured-content` item on both previously-invisible
  `finetuning.uk` pages, raised by the rotation reaching that site in its normal order, with
  nobody dispatching it. **One figure in that lane's close-out was wrong and I corrected it on
  request** (a migration-apply timestamp, `10:47Z` claimed vs `09:49:52Z` actual per the
  ledger) — everything else in the close-out re-verified independently and held. Kept in
  `bugs_open/` per the owner ruling of 2026-08-06 (fixed bugs stay there). **Not this lane's
  to touch further** — owned by `bugfix_230_discovery_driver`.
- **`bugs_open/223`** — the landmine-verifier reports every non-Go footprint as non-existent
  (code index is 100% Go, 5,755 symbols, 0 `.sh`/`.py`/`.md`/`.sql`). **Filed 08-08 by
  `provocation_pipeline`, OPEN, UNOWNED.** This lane hit it twice more (two landmines filed
  08-08, one came back `STALE` for exactly this reason) and **contributed the recurrence
  rather than filing a duplicate** — see that bug's 2026-08-09 section for the measured index
  composition and a second, independent trap (`landmines-sync.py --apply`, run the way
  CLAUDE.md instructs, consumes its own `NEEDS_VERIFICATION` signal — recovery path is in
  `LANDMINES.md` under the sync-trap entry). **Still open and unowned** — pick it up if you
  have capacity, but it is not this lane's responsibility.

## 4. Open, and confirmed NOT this lane's to close (unchanged from 08-08, still true)

- **`empty_sections_loop_integrity` — the cheap fix.** `bugs_closed/032`'s own "stronger
  option": ask whether the page still declares the slot, return `Resolved:false`. Now that
  fail-closed is proven live (§2), this converts the 3-rebuild cost into one honest verdict.
  Told three times now across this lane's life; still their call.
- **`bugfix_071` lane's landmine** — the success UPDATE never clears `site_work_items.error`,
  so a row refused then completed keeps stale refusal text. Fail-closed makes it commoner, not
  caused by this lane. Plausibly a one-line fix; theirs, not ours.
- **`RFC_017` option 3** (`Indeterminate`/park outcome) stays open in the RFC. Now has one
  real data point (§2.3) rather than zero; still not enough to justify building it — the
  owner's stated trigger is the retry cost showing up in the numbers, and one occurrence is
  not a rate.

## 5. Lane files

`PLAN` · `RUNBOOK` (R8 = the verifier-error census + the three ways it lies — still accurate,
now with a fourth row in the table) · `NOTES` (evidence + every misstep, newest at bottom) ·
`README_where_we_are` (owner's plain prose) · `SUMMARY_2026-08-06`, `SUMMARY_2026-08-08` ·
three `SUBMISSION_*.json` (historical, all resolved). Outside the lane:
`architecture_review/RFC_017` (decision record, option 3 still open),
`docs026_concept_register/register/work-item-integrity.md` `WII-011` (now says PROVEN, not
unproven), `bugs_open/230` (fixed, live, owned elsewhere), `bugs_open/223` (open, unowned),
`WRONG_CALLS.md` (this lane's 08-08 entry), `LANDMINES.md` (three entries from this lane:
the slot-rename trap, the sync-signal trap — both `STILL_VALID`/dispatched correctly on
retry — plus the pre-existing `_verification.status='error'` era-inversion entry).

## 6. If you're picking this up cold

There is genuinely no queued task in this lane. Your options, roughly in order of value:
1. **Nothing** — the lane is closed. Move to whatever else needs attention.
2. **`bugs_open/223`** if you want an unowned, well-scoped bug with three independent
   confirmations already on record.
3. **Watch for a second RFC_017 fail-closed occurrence** (§2.3) — one data point doesn't
   establish a rate, and `bugs_open/230`'s rotation now examines every site roughly weekly, so
   more will surface over time. If they cluster, that's the evidence RFC_017 option 3 is
   waiting for. Not worth polling for; worth noticing if you're in the area.
