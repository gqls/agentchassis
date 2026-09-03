# 465 — RETRACTED the day it was filed. Not a bug: a duplicate with an overstated headline

**Filed and retracted 2026-09-03** by the gripper dossier lane (session: robot hands).
**Nothing here is biting production.** Filed into `bugs_open/` for ~2 hours; moved here so
that directory keeps answering "what is biting prod right now".

**Kept, not deleted**, because the way it went wrong is more useful than the finding was —
it is a clean instance of the trap named in the landmine it duplicated.

---

## What I claimed

That **8 of 18 `js_snippets` rows bind zero listeners "in production"**, because the bundle
is a synchronous `<head>` script and self-guarding is a per-snippet convention.

## What is actually true

The **mechanism** is real and correctly described. The **headline was false**, and the
finding was **already on record** — filed the same day, by the previous session on this same
lane, in better form.

| | my filing | the record |
|---|---|---|
| how the census was done | `js_content LIKE '%DOMContentLoaded%'`, then reading 8 bodies at source | **by execution**, under a head-parse DOM stub, with the pre-fix widget as a negative control |
| `is_active` | **never queried** | queried: **ACTIVE 9 rows, 0 exposed; INACTIVE 9 rows, 8 exposed** |
| headline | "bind ZERO listeners **in production**" | "**Activating** a row ships a widget that can never render" |
| where | a new `bugs_open/` file + a new LANDMINES entry | `LANDMINES.md` (canonical entry), lane NOTES, `doc_notes` |

`[MEASURED 2026-09-03]`
```sql
SELECT is_active, count(*) FROM js_snippets
 WHERE js_content NOT LIKE '%DOMContentLoaded%'
   AND (js_content LIKE '%querySelector%' OR js_content LIKE '%getElementById%')
 GROUP BY is_active;
--  f | 8      <- and no 't' row at all
```

**All eight are dormant.** The live bundles are clean. The exposure converts to a live
defect the moment somebody flips `is_active` — which is exactly what the canonical landmine
says, and is a materially different claim from the one I filed.

## The three errors, each one command from being avoided

1. **I appended to `LANDMINES.md` without grepping it.** CLAUDE.md says to grep it by
   footprint before touching anything unfamiliar. `grep -n js_snippets LANDMINES.md` returns
   the canonical entry. My duplicate went in ~600 lines below it.
2. **My census used the method that entry forbids, in bold, in the same file:** *"DO NOT USE
   A SOURCE SCAN FOR THIS — `js_content LIKE '%DOMContentLoaded%'` passes any snippet that
   merely mentions it."* The previous session had already thrown this method away once, and
   written down why.
3. **I never queried `is_active`.** One column, in the table I was already querying, in the
   query I had already written.

I also did not read the lane's own `NOTES` to the end before filing — the audit, the
`is_active` split and the two thrown-away methods were all sitting in it, written hours
earlier. The lane handoff's top banner says to go straight to the bottom block; I did, and
then treated the council objection as unanswered when NOTES showed it had been closed by
measurement.

## The transferable lesson

**A correct mechanism plus an uncounted population reads exactly like a finding.** Reading
all eight bodies at source *was* real verification — of the mechanism. It could never have
surfaced the error, because the discriminating column was not in the query at all. The
check that would have caught it is not "read more carefully"; it is *"name the population
this claim is about, and put it in the WHERE clause."*

Related: this is the same family as the lane's own `[MEASURED]` discipline note — a marker
proves a measurement was *claimed*, not that it *could have come out otherwise*. My census
returned the same 8 rows whether or not they were live.

## Where the real record is

- **`LANDMINES.md`** — *"Activating a `js_snippets` row ships a widget that can NEVER
  render — 8 of 9 inactive snippets query the DOM during `<head>` parse and bail silently"*.
  Carries the execution-based check, the goja runner, the two refuted scan methods, and the
  guard shape to copy.
- **`LANDMINES.md`** — my retracted duplicate, kept in place with its correction, directly
  below the corrected heading.
- **`robot_hands_gripper_dossier/NOTES_robot_hands_gripper_dossier.md`**, 2026-09-03 — the
  audit, its negative control, and the council round it answered.
- **`WRONG_CALLS.md`**, 2026-09-03 — this misstep.
- Council corr `5775dc10-c791-4285-9f4c-249a055b5aa3`, seat `bug_historian` — the objection
  that prompted the audit. **It was already CLOSED by measurement before I re-opened it.**

## The one thing that is genuinely still open

The **structural fix is unmade and unowned**: the renderer could emit `defer` on the bundle
`<script>` tag, or wrap every snippet at render time, and then no author could reintroduce
this. The canonical landmine records it as **architecture-scope — not something to slip into
a bug fix**, wanting its own round. That routing stands. It does not need a `bugs_open/`
entry to stay visible, and this file is not one.
