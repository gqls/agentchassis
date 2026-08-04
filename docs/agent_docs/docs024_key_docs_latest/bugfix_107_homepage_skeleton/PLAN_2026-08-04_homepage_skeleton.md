# PLAN — bugfix 107: every site gets the same homepage skeleton

**Opened** 2026-08-04, by the bug-sweep session that closed 121 the same evening.
**Bug:** `bugs_open/107_HANDOFF_2026-07-27_every_site_gets_the_same_homepage_skeleton.md`
**Owner-visible history:** `README_where_we_are.md` beside this file.

> **CORRECTED 2026-08-04, ~2 hours after opening — THIS LANE DOES NOT FIX THE
> BUG.** The claim below ("take on candidate 1") was wrong twice over, found by
> the P0 research itself: the owner PARKED this bug on 2026-07-27
> (`oufe/HANDOFF_2026-07-27_continue_here.md:78-82`), and the owner-approved
> `vigilant_designer_offer_analysis` programme (active, Phase 0 proven 08-04)
> carries candidate 1 as its Phase 4.1 and candidate 3 as Phase 4.2. Per the
> coordination rule ("if it says OWNED: contribute into the bug file, do not
> compete") this lane CONVERTED to a research contribution: mechanism map
> appended to the bug file, CONTRIB filed into the owning lane, wrong call
> logged in WRONG_CALLS.md. P1–P4 below are STRUCK — they are the owning
> lane's to run, on their schedule. What caught it and the cheap check are in
> the WRONG_CALLS entry of this date.

## Validity re-check (2026-08-04, before any work)

The bug is STILL VALID and got worse since filing. Fleet-wide homepage
composition query (page_components ordered by position, parent_instance_id IS
NULL, joined through pages/sites — RUNBOOK §1):

- **lendzy.co.uk (built 2026-08-02, newest framework-built site):**
  `hero > brief-explanation > info-card-grid > mechanism-flow > call-to-action`
  — the same skeleton the bug describes, five days after it was filed.
- oufe.com (07-25): `hero > brief-explanation > info-card-grid > call-to-action`.
- The sites that DIFFER (vonc, gamesdesign, dartsonline, loancalculator's
  ported pages) are the hand-directed or ported ones — i.e. divergence comes
  from humans overriding the planner, which is the bug's claim in one line.

Ownership checked: who-owns names the closed gemini lane only by citation;
live transcripts show one triage read (08-04 09:17) and directory listings.
No open `site_work_items` touch the planner/skeleton mechanism.

## Approach (bug file's candidate 1, + 3 as watcher)

1. **Give the planner an archetype that constrains shape, not just palette** —
   wire the existing classification vocabulary (`site_type` / the per_site_ai
   archetype×pattern grid) into the section-planning step as a structural
   constraint: required/forbidden/ordering rules per archetype, so
   "publication" cannot emit conversion furniture.
2. **A sameness watcher** (candidate 3) — a check that compares a new site's
   composition against the fleet and flags convergence; diagnostic, proves the
   fix worked.
3. Candidate 2 (brief specifies structure) is nearly free and worth doing if it
   falls out of 1 naturally; it alone cannot close the bug (new sites keep
   defaulting).

## Decisions & their reasons

- **2026-08-04 — diagnosis loop BEFORE the fix.** The root-cause claim ("the
  planning loop carries no structural constraint from site kind") is durable,
  structural, and the fix changes fleet-wide behaviour — squarely inside the
  "always file" list in CLAUDE.md, and the OWNER RULING of 2026-07-31 makes
  the loop (or a declared substitute) mandatory for structural claims. Filing
  090 after the code-map grounds the symbol pointers.
- **2026-08-04 — plan by Fable, implementation by Opus** (user directive for
  this lane), council gate before/alongside the commit.

## Phasing

- P0: research (code map + docs prior art, two read-only agents) — DONE, see NOTES
- P1: file 090 needs_diagnosis with mechanism-stated symptom; await verdict
- P2: finalise plan against the verdict; write the fix (Opus)
- P3: council gate; commit narrowly; register any new seam in the concept
  register IN THE SAME COMMIT if one ships
- P4: verify against a real build (two contrasting archetypes — the bug's own
  verification: compare section lists, not screenshots); close only at
  fixed-AND-live
