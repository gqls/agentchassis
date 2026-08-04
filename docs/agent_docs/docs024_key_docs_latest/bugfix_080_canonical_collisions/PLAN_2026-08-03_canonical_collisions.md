# PLAN 2026-08-03 — bugs_open/080: close the canonicalisation class, detect the collisions

Owner thread: this lane (opened 2026-08-03, evening). Bug: `bugs_open/080` — the gap-planner's
`new_page` bypassed `datahelpers.CanonicalisePage`. The *arm* is fixed and live (v1.0.1177); the
class is not closed and the damage is not detected. `who-owns.py 080` → unowned; no live session
on it (checked transcripts, not just commits).

## Decisions and their reasons

1. **Detect and file only; no live mutation** (user ruling, 2026-08-03). robot-hands.com's two
   live duplicate pairs are a *decision* (which row survives), and retraction is council-VETOED
   (`bugs_open/098` → RFC_011). The check files `needs_human_review` items carrying the decided
   section-index-family convention, the 081 idiom.
2. **"Every un-canonicalised writer" — honoured per-surface, not literally** (user asked for a
   recheck before implementing; the recheck found two surfaces would be harmed):
   - `create_blog_posts_action.go` → canonicalise. Measured: 0 live names change (`name ~ [A-Z ]
     / %/% / %.html` → 0 rows), so no re-keying; byte-identical for clean blog-post slugs.
   - `deploy_tool_action.go` → canonical for NEW tools only; existing row looked up under
     either name (`<bare>`, `tool-<bare>`) and reused. Naive canonicalisation would mint
     duplicates on the 12 live legacy-shape rows — the bug's own shape. Implements TL-010's
     registered proposal.
   - `create_tool_component_action.go` → the silent flat-URL fallback (empty canonical triple →
     `/tools/<function>.html`) becomes fail-loud; the companion guide routes through the helper
     (measured byte-convergent — enforcement only).
   - `adopt_verbatim.go` → EXEMPT: preserving crawled URLs is the feature (LANDMINES.md:2476 —
     canonicalising adopted pages is the documented trap). Gets a convergence TEST pinning its
     name mirror to `CanonicalisePage` so it cannot drift silently.
   - `cmd/webdesignport/import.go` → EXEMPT, manifest is the identity authority; detector covers.
   - `create_report_page_action.go` → out of class: `report-<uuid>` names are unique by
     construction, one producer, so divergent identity cannot occur.
3. **Detector = 080's fix candidate 3**, built as discovery check `page_canonical_collision` on
   `completeness-discovery-agent`, with a mechanical verifier. Two grouping signals both needed —
   verified against live data: canonical-name catches `/news` (both rows → `news-index`) but NOT
   `/gripper-catalog` (orphan typed `content` canonicalises to itself); URL path-key catches both.
   File an item only when ≥2 rows in the group are `status='active'`.
4. **Liveness predicates**: never raw `build_status='deployed'` — `bugs_open/185`'s class. Use
   `datahelpers` predicates.

## Phasing

1. Lane docs (this commit) → 2. Surfaces A–D + detector F with tests → 3. build/test against
`git archive HEAD` → 4. council submission (one run for the coherent task; commit may carry
`Council-Submitted:`) → 5. narrow pathspec commits; concept-register entry in the SAME commit as
the detector seam (ordering-exemption condition 2) → 6. image, pod-verify (positive AND negative
grep, every replica), seed, induce on robot-hands (expect exactly 2 items; re-run → 0 new) →
7. update 080 / 016b §10+§9 / WRONG_CALLS / memory; close 080 → `bugs_closed/` naming BOTH paths
on the commit, verify at HEAD with `git ls-tree`.

## Measurements this plan rests on (all 2026-08-03, live DB/wire)

- Collision census (URL shape, both liveness predicates agree): 6 groups / 2 sites; 2 fully live
  (robot-hands `/news`, `/gripper-catalog` — the second is NOT in the bug file); 4 with one
  non-active side (dartsonline ×3, robot-hands learning-center).
- Wire: all four robot-hands URLs serve HTTP 200; `/news.html` carries a self-referential
  `rel=canonical`; `robot-hands.com/sitemap.xml` is 404.
- Tool shapes (`page_type='tool'`): 102 canonical (86 live) / 14 double-prefixed-url (14 live) /
  12 deploy_tool-shape (12 live) / 20 other.
- Blog: 79 `blog-post` rows, 78 canonical-URL-shape (1 legacy outlier predating the action);
  0 page names fleet-wide would be altered by `normaliseSlug`.

The full command set with gotchas → RUNBOOK. Running record → NOTES.
