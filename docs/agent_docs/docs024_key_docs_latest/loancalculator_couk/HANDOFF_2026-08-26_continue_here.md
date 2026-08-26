# HANDOFF — loancalculator.co.uk · the 385 fix is LIVE; two automated waves are inbound; one clean build-arm rebuild closes the bug (2026-08-26)

> Supersedes `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/HANDOFF_2026-08-25b_continue_here.md`.
> That file's arc (cause → fix → council APPROVED r2) is complete; what changed since:
> an overnight roll shipped the fix, the queue's stale detection of the 08-23 damage
> fired harmlessly, and two fleet mechanisms restarted that will churn this site.
> Evidence: `bugs_open/385` §7b · NOTES `## 2026-08-26`. Owner prose: `README_where_we_are.md`.

```
site        loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
fix         ✅ LIVE — matchPreservedSectionIdx + PairStoredToIncoming in /proc/1/exe on
            BOTH replicas (pods 6dd68888dc-*, started 08-25 23:11Z), controls both ways.
            Register LOCK-009. Council ece638fb r2 APPROVED.
page        tool-loan-vs-savings: 5 rows, locked calc at pos 2 (08-02 bytes), 0 orphans,
            0 byte-twins fleet-wide, served page 0 duplicate ids, GOLDEN MATCHES
            (--selftest green first, then --compare, 2026-08-26 ~11:12Z)
open        bugs_open/385 — CLOSE CRITERION: one clean BUILD-ARM rebuild (§7b/§9)
```

## What changed overnight and this morning

1. **The fix rolled.** Symbol-probed with both controls (bug §7b). The armed row can no
   longer duplicate; the owner's disarm question (flip `build_status` to `'approved'`)
   is downgraded to optional belt-and-braces.
2. **The 08-24 stale detection fired, harmlessly** — `content_duplication:
   tool-loan-vs-savings` (filed 08-24 14:49 against the then-real duplicate, never
   retracted by the hand remediation) dispatched 08-26 08:30 when promotion resumed:
   deleted nothing (zero history writes), completed, left a benign assemble-only
   `page_rerender` in the queue. Lesson recorded in §7b: **hand-repair must grep the
   queue for detections of the repaired damage.**
3. **Two automated waves are inbound — expect churn, do not read it as damage:**
   - `site-discovery-rotation-design` re-enabled 08-26 09:20Z after 15 days off
     (peer heads-up, webdesign-tool-rebuilds lane; `bugs_open/401`). ~1 site/3h,
     least-recently-visited first — this site inside ~2-3 days. Findings born
     `detected`; the 15-min promoter auto-dispatches known (item_type, handler) pairs.
   - The `bugs_open/397` GTM spec-key repair (analytics lane, owner go): writing the
     key triggers ONE `stale_chrome` → `needs_rerender` here at the next discovery
     pass — **chrome + 44/47 pages re-render and redeploy**. Served tag unaffected
     meanwhile. Not damage; no action.
   Both run the RERENDER arm (safe, and the fix is live regardless). Rows will get new
   `created_at`s and chrome will change — a mid-wave single sample proves nothing
   (WRONG_CALLS 08-24).
4. **Queue state this morning** (already non-empty BEFORE the design rotation's first
   visit — the 08:26–08:30 producers were the improvement-loop and
   acceptance-discovery): ~9 `heading_promise_unmet` `detected`, 4 `capability_gap`
   `deferred`, an improvement `needs_rerender` + the dedupe `page_rerender` `triaged`;
   site totals 76 detected / 72 deferred / 65 needs_human_review / 10 triaged / 1
   failed. Nothing 385-shaped. Owner may want a view on the backlog; not this lane's to
   drain unasked.

## The one thing that closes 385

A **build-arm** run (`needs_page` → `page-build-handler` → compile → `save_page_sections`)
on a locked positional tool page, coming out clean: byte-twins 0, locked row untouched,
golden `--compare` green (selftest first). The rerender waves above do NOT qualify (§9's
arm distinction). It can arrive naturally (a design finding promoting into a build) or be
owner-released as on 08-23; when it lands, verify per §9 and move
`385_…` to `bugs_closed/` — **name BOTH paths on the `git mv` commit** (LANDMINES: a
pathspec commit ships a COPY on a half-named move; verify with
`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 385` → exactly one line).

## Standing cautions (carried)

- `toolgolden.py --selftest` green FIRST, or nothing measured afterwards is quotable.
- Prove a deploy at the artefact; per SERVICE, not per fleet; symbol probes carry BOTH
  controls (bug §9 has the exact commands).
- ⚠ `UPDATE page_components SET position` does NOT touch `updated_at`.
- ⚠ Before any repair of 385's shape, check `pages.sections` for a stale sixth entry.
- Hand-filed / un-parked work items must be `triaged`; the dispatcher cannot see `detected`.
- `retract_page_deployment` refuses active pages; explicit `page_ids` always.
- Query runs BY CORRELATION, never `now()`-interval; collected_data can purge in ~2h.
- Before any planner run: the four cautions in `HANDOFF_2026-08-23_continue_here.md`.
