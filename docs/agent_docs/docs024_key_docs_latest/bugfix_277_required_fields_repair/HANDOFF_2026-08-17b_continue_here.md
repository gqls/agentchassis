# HANDOFF — 2026-08-17b (evening), fresh chat starts here: both trails APPROVED, both owner decisions implemented, 444 self-corrected by 454 — and THE FLEET IS 252 COMMITS BEHIND

**Supersedes `HANDOFF_2026-08-17_continue_here.md`** (same day, earlier). Measured 2026-08-17
14:40–17:00Z. Read from disk; then `NOTES_required_fields_repair.md` from the bottom.

## 0. READ THIS FIRST — the running chassis is NOT built from HEAD

**The afternoon roll shipped no new code.** It rebuilt and deployed at the **same `IMAGE_TAG`**
as the morning (`v1.0.1305`), so the node served its cached layer. Verified on three instruments:

| instrument | reading |
|---|---|
| local image at `v1.0.1305` | `revision=89a0cbeb7`, `created=2026-08-17T14:30:02Z` — a real new build |
| running binary `/proc/1/exe` | `6a782274b` **PRESENT**, `89a0cbeb7` **ABSENT** |
| pod `imageID` vs local repo digest | `sha256:f90a7e88…` vs `sha256:6039e19c…` — **different images** |

Pods restarted 14:42/14:43Z, *after* the 14:30 build, so it is not a timing race. **252 commits
unshipped, 26 touching `platform/`/`internal/`/`pkg/`/`cmd/`.** Contributed to `bugs_open/153`
(which owns the trap). **Consequence for any session: "did my Go fix ship?" is currently NO for
anything after `6a782274b`, and that is not evidence your change is missing from the build.**
Fix is owner-run: bump `IMAGE_TAG` (still `v1.0.1305`) then a whole-fleet release.
**DB config is unaffected** — `agent_definitions`/`scheduled_tasks` are live at COMMIT.

> ⚠ Use a **real but different commit sha** as the negative control. A 40-zeros needle is present
> in every binary (LANDMINES, 2026-08-17) and makes a sound probe look broken.

## 1. What is LIVE

| thing | state |
|---|---|
| council `05a3d1c8` (the promoter) | **APPROVED** round 2, 12 seats, 2 advisories both answered |
| council `7b0e2833` (the router) | **APPROVED** round 5, 9 seats, 3 advisories; trail closed |
| `required-fields-missing-handler` (410 v3, CQ-023) | live; **only 3 of its 8 routes have ever been taken**, convert arms never fired, zero child items |
| `detected-item-promoter` (430 + **444** + **454**, SCH-026) | live, 900s; door-closers + `verified` fix |
| `held-pair-canary-escalation` (**453**) | live, daily, 3-day limit; **proved itself on its first tick** |
| migrations 444 / 453 / 454 | applied, ledger-recorded, each with a `_ROLLBACK.sql` |

## 2. OWNER DECISIONS — both answered 2026-08-17, do NOT reopen

1. **Gate the router's convert arms?** → **LEAVE AS IS.** No change made. `bug_historian` raised it
   at rounds 4 and 5 and a second seat backed a fail-loud guard; the owner ruled anyway. Recorded so
   it is not re-litigated. The arms remain **never-exercised** — first run will be in production.
2. **Owner per type + a time limit?** → **DONE, migration `453`.** Daily task; a pair held by the
   canary rule >3 days moves `detected` → `needs_human_review` carrying its owner, `days_waiting`,
   and the exact by-hand promote. **The owner map enriches, never gates** — an unmapped type
   escalates with `(UNASSIGNED - claim this)`, i.e. louder. Owners named from evidence:
   `placeholder_contact` → `bugs_open/201` lane; `page_component_status_drift` → **deliberately
   unassigned** (its check was added 2026-07-10 and never touched; naming a plausible owner would be
   a fabrication that reads as a decision).
   **First tick 12:57:43Z: escalated the 4 seven-day rows, correctly left the 1-day pair alone.**

## 3. `444` WAS WRONG AND `454` FIXES IT — read this before trusting any handler-reliability number

`site_work_items.status` has **TWO** terminal success states. `verified`
(`complete_work_item_verification.go:218`) is a completion that *also passed verification* —
`idx_swi_completed`'s predicate already lists both. 430's known-good rule and 444's 25% floor both
counted only `complete`, so **a pair's apparent success rate FELL as its work got verified** — a
metric that degrades as the platform improves.

Caught by a pair moving underneath: `empty_section → page-build-handler` read **11/13 = 46%** at
11:00Z and **3/12 = 20%** at 16:30Z with nothing regressed (a sweep moved 9 completes to
`verified`); counting both it is **12/12 = 50%**, and 444 was holding 2 live rows for no reason.
`454` fixes **both** predicates, with controls flipping opposite ways in one run.

**The latent case is why it could not wait:** a pair whose successes have *all* been verified has
zero `complete` rows, reads as never having worked, and is held **for ever** — `bugs_open/083`'s
own disease. Scope when found: `verified` was 9 rows across 1 pair.

## 4. Owed work, in priority order

1. **Tell the owner the fleet is 252 commits behind** (§0) — nothing else here is blocked by it, but
   other lanes are, and they cannot see it.
2. **Watch `453`'s next ticks.** `placeholder_contact` should escalate on ~2026-08-19 (created
   08-16, 3-day limit) — that is the clock's second discriminating test. `empty_section` should NOT
   escalate: after `454` its pair passes the floor and the promoter should take it.
3. **`277` → `bugs_closed/`** on ~**2026-08-22**: churn guard (at day 2: one new row, born-detected →
   promoted → routed → parked, **zero `unresolved`**) plus the two cancelled conversions re-raising.
   **Both paths on the commit** — LANDMINE.
4. **`083` → `bugs_closed/`**: criterion 3 MET at the served page; criterion 1 holds under its
   corrected wording; criterion 2 met but **non-discriminating** (already true 6 days before the fix
   — do not bank it). Named open residual: a new pipeline value silently stops being promotable; the
   cheap control is to have the pre_query return a count of rows each door held.
5. **Start the `router_engine` lane (RFC_030)** — its own cold-start handoff exists. Now the largest
   real work here, and the council seats' residual ("a hard gate on a 4th router, not merely a lane")
   is waiting on it.
6. **Raise the council gate's config blind spot**: `097` scopes on `platform/`/`internal/`/`pkg/`, so
   a mechanism shipping as `agent_definitions`/`scheduled_tasks` config cannot be submitted at all.
   Round 2 needed `FORCE=1`; round 5 passed only because it was anchored on a Go file. Another lane
   has filed a LANDMINE on it (`landmine(297)`); it wants a real fix.

## 5. Landmines this lane hit (all in LANDMINES.md / WRONG_CALLS.md)
- **`status` has two success states** (`complete`, `verified`) — §3. `GROUP BY status` before
  filtering on it.
- **`failed` rows carry NO `completed_at`** (0 of 265) — a pair-health query keyed on that column
  returns a uniform 100% and cannot come out otherwise.
- **These two are ONE class and both hit in one session:** a status column whose values were
  *assumed rather than enumerated*. Neither was unmeasured — both were **measured against an
  incomplete definition**, which no marker and no council round detects (12 seats approved 444).
- **A same-tag rebuild ships the cached image** — §0.
- **A verify criterion is a figure and goes stale like one** (083's criterion 2 was carried forward
  twice while already satisfied).
- **A zero from a detector you just wrote needs a demand control.**
- **Fetch the URL the row literally names** (`/guides/index.html`, not `/guides/`).
- **Check whether YOUR OWN guard strands YOUR OWN lane** — 444's allow-list passes
  `(build, content, design)`; this lane's producer files `content` (`:163`). Clean by measurement,
  not by design.
- **A pathspec commit still takes a same-file passenger** — `b6bca52fc` carried another session's
  LANDMINES improvement. Nothing lost; named in NOTES.

## 6. Session-start checklist
`git log --oneline -10` · re-read this file from disk · **verify the chassis revision before
trusting any Go-code claim (§0)** · `scripts/who-owns.py 277` and `083` (by SLUG — 083 is an
ambiguous number) · re-measure §1 · then §4 item 1.
