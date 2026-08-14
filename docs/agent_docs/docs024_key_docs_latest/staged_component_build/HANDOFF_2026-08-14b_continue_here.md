# HANDOFF — 2026-08-14 (b), fresh chat starts here: `bugs_open/248`'s R4 routed to `RFC_029`, the backlog-drain job DESIGNED (not executed) — one production decision left before it can run

**Supersedes `HANDOFF_2026-08-14_continue_here.md`.** That handoff's own §3 named three things
left: (1) route R4 to a human/RFC, (3) design the backlog-drain job, (4) `bugs_open/256`
(explicitly not this lane's). This session did (1) and (3). (4) is still explicitly not this
lane's job — leave it.

## 1. State (verified 2026-08-14, ~17:00–17:30Z)

- **Both named symptom sites remain fixed** (gaswholesalers' logo, mortgagecalculator's hero) —
  nothing touched them this session, no reason to re-check unless a fresh symptom appears.
- **`RFC_029` filed and committed** (`439382985`):
  `architecture_review/RFC_029_the_aggressive_recursive_search_has_no_boundary_for_an_unmapped_field.md`.
  It routes R4's objection ("should the shared fallback strategy ever run for a field with no
  explicit `input_fields` entry, at all?") as its own architecture question, cross-referenced
  against the already-open `RFC_028` (same overall resolver — `ExtractActionInputs` — but a
  **different, nested arm** RFC_028's own census cannot see, because that arm
  (`findFieldRecursive`, `unified_extractor.go`) lives in a file RFC_028's SQL query never
  matches). Cites `WFA-009` as existing, owner-approved precedent for the shape of an answer
  (opt-in hard-fail, unsafe default OFF). **This now sits with the owner, same as RFC_028** —
  nothing to do here except wait, unless a fresh session wants to work the RFC track itself
  (read `architecture_review/` conventions first, it is not a normal task).
- **The backlog-drain job is DESIGNED, not executed.** Full detail in `bugs_open/248`'s own
  newest CONTRIBUTION (`## CONTRIBUTION 2026-08-14 (fresh session) — R4 routed…`) and
  `NOTES_staged_component_build.md`'s `## 2026-08-14 (drain job design)`. In short:
  - Re-measured live: **140 rows / 14 sites** carry the placeholder marker (down from 08-12's
    150/16 — organic repair + the fix stopping new instances, not a new gap).
  - **The flat number is the wrong shape to act on.** Bucketed by each row's own
    `undeployed_asset` work-item history:
    - **A (13 rows):** `unresolved`, needs an explicit promote-to-`triaged` — the exact
      one-line action §2 proved on gaswholesalers.
    - **B (26 rows):** already `triaged` — **self-draining**, `build-dispatch-loop` is alive
      and periodic (not stalled; 36 items sat `triaged` up to ~50 min with a real completion
      ~90 min prior — normal cadence, not a stuck queue). No action needed.
    - **D (64 rows, the largest bucket):** only terminal items exist — needs a **fresh cloned**
      item at `triaged`, the exact mechanism §2b proved on mortgagecalculator.
    - **E (30 rows):** no `undeployed_asset` item was ever filed. **Do not act blindly here** —
      reading `check_undeployed_assets.go`'s own query found this shape can mean the *page* was
      already fixed by a later, uncorrelated re-deploy while the *row* stayed stale — the code's
      own comment names this exact case for favicon/og_card and refuses to file against it
      ("a FALSE claim, which is worse than a missed finding"). **Curl the expected
      `/assets/images/<purpose>.*` path per row before deciding to act on any bucket-E item.**
      (Counts sum to 133 of 140 — a small reconciliation gap from rows with more than one
      matching item, flagged not resolved, doesn't change the shape of the plan.)
  - **The repair path itself (A/B/D) is confirmed LLM-free** — read `check_undeployed_assets.go`
    and `deploy_image_asset_action.go` end to end, neither calls an LLM; it re-deploys an
    already-stored image via DB + storage read + git commit. **Unaffected by today's fleet-wide
    LLM cap** (commit `8b897432a`, same day: every call 400 `invalid_request_error`, monthly
    cap, returns 2026-09-01). Only the hero/logo-**regeneration** subset that might live inside
    bucket E depends on that capped path — re-check `llm_call_log` live before touching it, this
    session's snapshot (0 ok/17 failed at 16:00Z, partial recovery by 17:00Z) is not a standing
    fact.
  - **Open, unresolved by this design:** the prior handoff's phrase "owned vs generic rebuild
    policy" does not match any `sites` column found on a direct check (`settings`,
    `locked_at`/`locked_by`, `status` — none of the 14 sites are locked). Flagged
    `[UNMEASURED]`, not guessed at. Confirm before batching across sites that might have
    different consent/ownership requirements.
- **Nothing was fired.** Buckets A+D together are ~77 rows of real git commits across 14 live
  sites, and the mechanism is proven at 2-row scale, not 77. This is a deliberate stopping
  point, not an oversight — matches the prior handoff's own instruction that this "deserves its
  own careful plan… rather than a one-off script improvised at the end of a session."

## 2. What's actually left

1. ~~**Execute (or get sign-off to execute) the drain, bucket by bucket, per the design
   above.**~~ — **BUCKET-A PILOT DONE, same day, user-approved, 11/11 clean** (2 skipped as
   live-referenced 200s, 1 canary + 10 waved, zero placeholder paths, all wire-verified,
   census 140 → **98 rows / 12 sites**). The pilot's revised design — **wire-check BEFORE
   promotion, gating it** — caught two would-be overwrites of served files and is now the
   proven shape for the rest of the drain. Full account: bug 248's newest CONTRIBUTION +
   NOTES `## 2026-08-14 (bucket-A pilot)`. **What remains of the drain:**
   - **Bucket B**: self-drains, confirmed live during the pilot (part of the 42-row census
     drop). Leave it alone.
   - **Bucket D** (~64 rows, now the bulk of the 98): needs the clone-to-triaged mechanism at
     scale. Inherit the wire-check-first gate verbatim — the pilot measured 2-of-13
     already-200 skips in bucket A, so assume a nonzero skip rate here too.
   - **Bucket E** (~30 rows): per-row wire check before any item is even filed, unchanged.
   - **New, small design question from the pilot**: rows that are "stale bookkeeping only"
     (already served correctly; 12 confirmed members) — should there be a bookkeeping-only
     correction that does not redeploy? Decide when designing bucket D, don't improvise.
2. **RFC_029 (and its sibling RFC_028) await an owner ruling.** Not a task for a normal session
   — if picked up, it is the `architecture_review/` process (RFC-shaped), read that directory's
   conventions first.
3. **The "owned vs generic" classification is still open.** Whoever executes the drain should
   resolve this — ask the owner, or find the actual mechanism — before touching a site that
   might need separate notice.
4. **`bugs_open/256`** (mobile screenshot exceeding the vision API's 8000px cap) — still
   explicitly **not this lane's job**. Belongs to whoever owns `run_checks_action.go`'s capture
   path.

## 3. What's explicitly NOT this lane's job

Unchanged from the prior handoff — see its §4 (mailer/PUB-003 work, `bugs_open/256`,
`bugs_open/235`). Nothing this session touched changes that list.

## 4. Landmines this session hit and are worth repeating

- **A flat row-count census is not a target list.** Bucketing by work-item state (and, for the
  bucket with no item at all, checking the live wire) turned "re-trigger 140 rows" into "26
  need nothing, 13 need one action, 64 need another, and 30 need a check before any action at
  all." The temptation to skip straight to a script from the round number is real — resist it
  when the number came from a single-condition marker query.
- **A periodic dispatch loop sitting on `triaged` items for tens of minutes is not evidence of a
  stall.** Check the last real `complete` timestamp before concluding the queue is stuck; this
  one had completed something ~90 minutes prior and was simply between ticks.
- **A discovery check's own exclusion logic (why it does NOT file against a shape) is exactly as
  load-bearing as its filing logic** — reading `check_undeployed_assets.go`'s comment on the
  favicon/og_card false-positive case is what caught that the same shape generalises to ~30
  other rows this design would otherwise have blindly re-triggered.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file FROM DISK; re-read `bugs_open/248`'s newest
   CONTRIBUTION for the full bucket detail (this file only summarises it).
2. Pod-grep chassis + browser-runner for the CURRENT build (method: §1 of the prior handoff —
   chassis's provenance line scrolls out fast; cross-check via a known-good sibling service's
   commit + a bogus negative control). Assume the fleet has rolled again since this was written.
3. Re-run the marker census (`SELECT count(*), count(DISTINCT site_id) FROM assets WHERE
   filename LIKE '%asset-key%' OR url LIKE '%asset-key%'`) and the bucket join before trusting
   this file's 140/14 and A/B/D/E counts — both are snapshots from ~17:00Z and will have moved
   if bucket B's self-drain (or a fresh discovery sweep) has continued since.
4. Re-check `llm_call_log` before touching anything that might call `image-build-handler` for
   actual regeneration (bucket E's non-brand-head subset) — the cap's live status this session
   was a partial recovery in progress, not a settled fact either way.
5. `who-owns.py 248` (both files will show — resolve by filename) before touching the drain
   execution — check for a concurrent session already mid-batch, given `site_work_items` showed
   very recent (within-the-hour) `triaged` promotions that turned out to be the normal discovery
   sweep, not a hidden competing thread, but that conclusion should be re-checked fresh, not
   assumed from this file.
