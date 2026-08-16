# HANDOFF 2026-08-15 — contrast front. Start here in a fresh chat

**Supersedes `HANDOFF_2026-08-10_contrast_front_continue_here.md`** (7 addendums, title now
stale). That file is history; nothing in it needs reading unless a link below sends you.

**Everything below was re-measured on 2026-08-15, not carried forward.** Where something is
unmeasured it says so.

---

## The one-paragraph state

**This front is finished bar one decision.** The bug it started on (`bugs_open/113` — a dark
site's palette omitting a slot, so the layout's light literal shipped and text went
invisible) is **repaired on every dark site on the fleet**, proven at served stylesheets,
and its council trail reached APPROVED. The platform gained a way to re-compose a site that
already has the wrong colour scheme (`allow_reinstall`), which is **live and was used once**
— that is what repaired `ai-agent-orchestration.com`. As of 2026-08-15 the owner has closed 122 and ruled the approval
default stays loose, so what is left is **one small revise and one decision about a bug
number**. No engineering is blocked and nobody is waiting on us.

## Live facts, each with how it was checked

| fact | value | how |
|---|---|---|
| chassis build | `0115f2b4528b0063fd01e7af275ccefe9c5a991d` | binary probe on `agent-chassis-7779f5d998-96lpf`, `deadbeef` control absent |
| `bugs_open/213` | **CLOSED**, in `bugs_closed/` | `ls` |
| `bugs_closed/122` | **CLOSED 2026-08-15** on the owner's ruling; moved in `d22a6c875` | `git ls-tree HEAD` |
| `bugs_open/113` | open | `ls` |
| `allow_reinstall` (DES-082) | **LIVE and USED** (the site repair) | merge-base + served artefact |
| approval recording (DES-084) | **LIVE, NEVER RUN** — 0 rows | `result ? 'reinstall_approved_by'` → 0 |
| `improvement-sweep` | **`enabled = false`** | `scheduled_tasks` |
| queue | **76 detected**, **225 parked `contrast_failure`** | `site_work_items` |
| approval default | **owner 2026-08-15: do NOT tighten yet** | standing ruling |

> **Do not verify a deploy with `strings`** — retired 2026-08-11 (it is absent from the
> debian-slim images, so its failure is indistinguishable from "not stamped").
> **"Did my fix ship?" is:** `git merge-base --is-ancestor <your-commit> <the stamp>`, and the
> stamp is **per service**, not per fleet.

## THE ONE THING OWED: council `9767969e` came back REVISE, and the objection is right

**Correlation `9767969e-92fa-44d0-b416-d7187c869531` — REVISE, gating objection from
`editquality`.** The code is **already live**, so this is a revision of shipped behaviour,
not a proposal you can decline.

**The objection, and it is correct.** `resolveReinstallApprover`'s fourth source reads
`approved_by` from the **work-item `spec` JSON**, while my rationale claimed it was *"the
column a real HITL approval flow already fills, so wiring one later needs no change here."*
`site_work_items.approved_by` is a **real first-class column** — reading the spec does not
read it, so wiring a HITL flow **would** need a change. The claim was false.

**Measured today, and it makes the impact nil but the claim no less wrong:**

```sql
SELECT count(*) FILTER (WHERE approved_by IS NOT NULL AND approved_by <> '') AS column_filled,
       count(*) FILTER (WHERE spec ? 'approved_by')                          AS spec_key_present,
       count(*)                                                              AS total
FROM site_work_items;
--  0 | 0 | 8314
```

**Nothing fills either one.** There is no HITL approval flow on this platform yet. So today
the bug is in the *claim*, not in the behaviour — but the claim is what a future implementer
would trust.

**The fix, when someone takes it:** read the **column** (or read both, column first), correct
the rationale, and resubmit with `RESUBMIT_CORR=9767969e-92fa-44d0-b416-d7187c869531`. It is
a small change. **I did not make it** — see "why I stopped" below.

## Decisions — TWO ARE NOW RULED (2026-08-15). One is still open.

**~~D1 — `bugs_open/122` wants your look.~~ RULED 2026-08-15: "122 is now closed."**
**DONE** — moved to `bugs_closed/122_…` (`d22a6c875`), both paths named on the commit and
**verified at HEAD**, not at the tree. The bar was met by that lane's own `700f06bfc`
(WIDENED AND COMPLETE, all 97 article-body placements across 20 sites), `63d42fcda`
(migration 415 applied) and `c9049cb2c` (derivation fix live on v1.0.1298).
*Worth knowing why it was still open:* its status block said *"stays per the owner's
2026-08-06 ruling"* — **superseded on 08-12**, when the owner restored CLAUDE.md's bar.
True when written; nothing joined it back up to the later ruling. **Not re-verified at the
artefact by me** — the closure rests on the owner's ruling plus that lane's commits, and the
file says so.

**~~D3 — tighten the approval default?~~ RULED 2026-08-15: "don't tighten yet."**
Matches the evidence: DES-084 has **never run** (0 rows), so the query that would show who
*would* have been refused has no population at all. **Nothing to do — this is a standing
"leave it".** Revisit only after real composition replaces have happened; **do not force
one to generate data.** The one-line flip, when it ever comes, is in
`resolveReinstallApprover` (return `""`, have the caller refuse).

**D2 — STILL OPEN: should `bugs_open/113` close?** Its own mechanism is fixed and proven
fleet-wide, and with 122 now closed, **113 is the last file this front holds**. What still
sits under its number is the *primary-used-as-ink* family (`features_open/026`), which was
122's territory — and 122 is closed. **Recommend: close 113**, and if the ink family needs a
home, give it its own file rather than leaving it under a number whose headline defect is
fixed. Leaving 113 open makes its status unreadable: a reader cannot tell which of two
mechanisms is unresolved.

## What NOT to do

- **Do not unpark the 225 `contrast_failure` rows, and do not write a contrast verifier.**
  `bugfix_122_contrast_ink_slots` costed that fork (`b2fca2f8f`), found the exemption is on
  the record at `verifier_coverage_test.go:156` justified by an argument RFC_017 refuted, and
  chose **discovery-path retraction** instead. That lane is active and ahead.
- **Do not enable `improvement-sweep` casually.** Its pre-query is `LIMIT 1` — **one site per
  900s tick** — and it *discovers* as well as triages, so `detected` can rise after a run.
  That is correct behaviour, not a failure. It is off; leave it off unless you want the spend.
- **Do not reuse a `Council-Submitted:` correlation for a later, different change** — see the
  lesson below.
- **Do not re-derive ownership from memory.** `./scripts/who-owns.py <bug>` **before** filing,
  not after.

## Three traps this front actually hit (all cost real work)

1. **`palettes.source_domain` is stamped only on a per-site FORK.** Asking it about a site on
   a shared seed palette returns 0 rows, which reads as "no palette". The real chain is
   `sites → style_collections → css_themes → palettes`. In `LANDMINES.md`.
2. **`jsonb_each(default_config->'workflow'->'steps')` is TOP-LEVEL ONLY.** The dispatch
   loop keeps its real work in `process_item.config.sub_workflow.steps`. I filed the
   resulting absence as a structural finding; it was false. There is a landmine for exactly
   this, footprinted on a **table** — so the SessionStart hook cannot surface it and you must
   grep it yourself.
3. **One trailer, one plan.** DES-084 shipped under `Council-Submitted: b8e341b9`, a
   correlation that later approved a plan **not containing** the approval feature. `098` then
   credited it *"by correlation, via submitted"* for a review that never looked at it. Not a
   MISMATCH by the report's rule, but misleading — which is why `9767969e` now exists.

*All three are the same shape, and it is worth internalising:* **an absence produced by a
query is a claim about the query first and the world second.** Two of the three returned zero
rows, and neither zero was about the platform.

## Why I stopped here

The revise is small and I could have written it, but it changes **shipped** behaviour on a
seam that no live consumer exercises, and the honest version needs one thing I do not have:
whether a future approval flow should write the **column**, the **spec**, or both. That is a
design call for whoever builds the approval queue, not a detail to guess at while patching a
rationale. Everything needed to make it is above, measured.

## Cold-start checklist

1. Read this file. Ignore `HANDOFF_2026-08-10_*` unless a link sent you there.
2. `grep -n "<table|symbol>" docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` before any
   census — the hook only matches **file** footprints.
3. Confirm `improvement-sweep` is still `enabled=false`.
4. If you touch the approval seam, read verdict `9767969e` first — it is REVISE and the code
   is live.
