# HANDOFF 2026-08-30 — dispatch throughput — CONTINUE HERE (supersedes HANDOFF_2026-08-26)

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` 2026-08-27 + 2026-08-30
entries (the whole arc: gate PASS → 658 applied → 657 applied → acceptance PASSED at +2h and
+3.3d → two LLM outages → resume-state read) → `README_where_we_are.md` (owner's log; the
2026-08-30 entry is the current ask). Paths relative to
`docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## The one-paragraph state (evening 2026-08-30, ~21:30Z)

**Everything this lane set out to ship is SHIPPED, LIVE, and MEASURED.** Ruling B (migration
637, one trigger @30s) since 08-26; Phase 3 (658, batch 5→8 both knobs) applied 08-27
09:15:06Z and capability-proven (loops load 8, claim keys to _7); the 413 starvation fix (657,
selector ranks sites by loadable work, K read live from 658's knob) applied by its session
08-27 13:18:19Z on this lane's all-clear. **Acceptance: PASS at +2h** (worst unserved-with-
old-work fell from 6–10h to ~68 min; commit `c5fc3016d`) **and PASS at +3.3d** (backlog fully
drained; six fresh rows fleet-wide; zero pins; zero stuck claims; lost claims **3.9%**
trailing-24h vs 58–60% pre-B; commit `10c14ff54`). Daily 584 VERIFY: all 7 hold (08-30 run).
413's closure is supportable and is the 413 session's call. **The blocking issue is not
dispatch:** a fleet LLM outage ("specified API usage limits" 400) has run since ~08-28 —
99%+ of calls failing 08-29/30, owner push-notified twice — so demand generation is idling.
That is D4's FOURTH live case.

## NEXT (in order)

1. **Nothing dispatch-side is owed until the account is topped up.** On recovery: sanity-read
   (fires/loops/claims per hour + floor with ALL THREE controls) once demand returns, and only
   then consider grading Phase-3's ~+7% — clean post-657 LLM-healthy windows have all aged out
   of retention (~24–27h); do not force the grading (NOTES 08-30 windowing note).
2. **D4 LLM spend governor — the first BUILD item, now with four measured cases** (08-17
   self-set limit; 08-25/26 credit balance; 08-27 11:30–13:35Z; ~08-28→now). Confirm at-cap
   shedding policy with the owner first. Gate for interval ≤25s (option C) once built.
3. The rest of the standing queue, unchanged: deploy batching (D8 interim) · clients-first
   lane (D2) · Batch API (D6) · D16 retention (it destroyed the 08-28 read — design
   consideration, not just cost) · per-class maintenance LLM cost (RESEARCH §6) · DNS plan B
   (domain-programme lane; check pickup).
4. **Daily habit:** 584 VERIFY (3–20 min under load, never a 2-min timeout; zombie NOTICEs
   benign — now TWO reaper spellings excluded, widened 08-27 `adebc2d11`).

## Traps (new since 08-26 first; the 08-26 handoff's all still stand)

- **`max(claimed_at)` is BLIND to claim-release cycles** — a release (deferral or timeout
  reset) clears it; a site can read "unserved for days" while taking 9 loops/2h. LOOPS are the
  service meter. RUNBOOK floor carries this + the LOCK control + the STUCK-CLAIM control —
  never quote a worst site without all three.
- **`claimed-item-timeout`** (scheduled_tasks, 120s tick; NOT named "reaper" — two sessions'
  censuses missed it): auto-completes evidenced claims >15 min, RESETS the rest >40 min with
  backoff. A dropped-spawn dark window is therefore BOUNDED ~40–42 min; a stuck claim OLDER
  than ~45 min means this task itself is broken. Full story: 413's 08-27 addendum (corrected
  same day, twice, with the 414 session).
- **A `failed`/reset item is not lost work** — 54 of 66 rerender resets on 08-27 had deployed;
  locate the failing step in workflow order before re-doing anything (shared memory topic
  `a-complete-work-item-is-not-a-repaired-artefact`, mirror section).
- **Kubeconfig token = 3-day cycle** — expired 08-27 19:11Z mid-protocol; `Unauthorized`
  everywhere = expiry, owner refreshes. Decode-check is in the memory topic.
- Timer discipline: `date -d 'today 09:00'` parses LOCAL time (BST cost the 08-27 read 23
  min); long background sleeps get reaped — use a persistent Monitor for waits >1h.
- `orchestration_states` retention ~24–27h — baseline the moment you need it; the 08-28
  Phase-3 window is gone for ever.
- 658's rollback restores 5 EXPLICITLY (never `#-` the keys — Go defaults 50/20); do NOT
  re-run `051_build_dispatch_loop.sql`; 657's rollback is its session's file.

## Coordination state (all quiet)

- **bugs_open/413 session:** fix applied + measured; closure suggested to them 08-30 ~21:25Z
  (their call). Residuals NOT theirs to close: candidate-2 age-floor policy (owner question,
  in README), **bugs_open/415** (fire-gate narrower than selector — untouched, unowned).
  > **CORRECTED 2026-09-02 (by the 413/415 session):** 415 was taken up and CLOSED same day —
  > migration `688` applied 13:28Z, gate ⊇ selector on BOTH rows (584 VERIFY 1/7 unaffected:
  > parity, not text; your lane acked). Now `bugs_closed/415`. The age-floor question was also
  > RULED (provisionally declined) — recorded in `bugs_closed/413`'s closing section.
- **bugs_open/414 session:** all threads closed 08-27 (jointly corrected each other; their
  §7k and WRONG_CALLS carry the sequence). Nothing owed either way.
- **396 lane:** their CONTRIB on 657's guard (paren-precedence presence tests) passed to the
  413 session — their call, non-blocking.
- Council: nothing owed (658 corr `95099f95` APPROVED + advisories done; 657 corr `ecf2e542`
  r2 APPROVED, theirs; 637 corr `69a04e0a` APPROVED).

## Session hygiene

Pathspec commits; grep LANDMINES for symbols you touch; who-owns before routing at any bug;
re-read CLAUDE.md from disk before acting on multi-session rules; the workstreams memory
points here (COLD-START = this file).

## UPDATE 2026-08-31 — D4 stage A SHIPPED (this file's NEXT-2 is in progress)

Owner ruled the shedding order (maintenance → builds → research; NOTES verbatim; supersession
of the 08-21 order marked in RESEARCH §10). Stage A LIVE + council-APPROVED (corr 80df0963,
r2): migrations **671** (meter/class-map/config/state/120s task; $2,113 August MTD measured)
+ **672** (advisory lock, level-change proof, doc_plans travelling design) + **673** (FOR
UPDATE snapshot close). All inert: enabled=false, budget NULL, no stage B. **NEXT = stage B**
(Go claim-step refusal reading governor_state, opt-in default OFF, register entry same
commit, own council round) **and the owner's monthly budget figure.** The lane PLAN §D4 and
doc_plans('pipeline','spend-governor') carry the design; prior art ruled out with evidence
(fuel.go = per-task depth; token-pressure tasks = truncation caps).
