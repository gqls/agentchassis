# 365 — a ported tool's sidecar `script.js` has NO retirement path, so all 13 Phase C rebuilds will strand publicly-served orphan files

**Filed 2026-08-22** by the `webdesign_tool_rebuilds` lane, on Phase C #1's first live attempt.
Capability gap, not a code defect — every guard involved behaved as designed.

## The requirement it blocks

TL-032 / the lane PLAN: **13** ported tools as of 2026-08-22 (the lane scope query; 1 rebuilt today, 12 remain) on webdesign.co.uk keep their logic in external
`<script src>` files under `/tools/<slug>/` (plus the shared `/tools/assets/webdesign-couk-header.js`
used by every ported page). The lane's standing rule: **the external asset must be retired with the
slot.** The slot half works (retire + rerender; the served page drops the `src` reference — proven
on `tool-blueprint-compiler`, serve-grade 2026-08-22 11:5xZ, `src="script.js"` = 0). The FILE half
has no owner:

- **`retract_asset_files` (asset-retraction agent, DGH-010) refuses by design.** Dry run
  2026-08-22 11:52:59Z (orch `44d5efbc`, corr `a7e165e2`), result:
  `{"path": "/tools/blueprint-compiler/script.js", "refused": "outside /assets/ — pages, feeds and
  chrome are page-retraction's or nobody's, never this action's"}` — requested 1, retracted 0,
  dry_run true. The guard is doing its job; the path class was simply never in its remit.
- **Page retraction** (`retract_page_graph.go`) owns `pages` rows and their graph obligations —
  a sidecar file is not a page and never enters it.
- **`cmd/webdesignport`** (which UPLOADED these files) is import-only; no retire/cleanup mode.

So after each Phase C rebuild the old `script.js` stays in the bucket, publicly fetchable
(200, e.g. 7,067 B at `/tools/blueprint-compiler/script.js` today), referenced by nothing. Thirteen
of these accumulate by Phase C's end, plus the shared header when the LAST ported page goes. Same
family as `bugs_open/359` (retired content still publicly serving) — stale, defective-by-audit code
at URLs crawlers may hold.

## Fix candidates (the decision belongs to the asset-retraction action's owner, bugfix_235/DGH-010 lane)

1. **Widen `retract_asset_files` with an explicit opt-in path class** for page-adjacent sidecar
   files (`/tools/<slug>/*.js|*.css`), keeping the refusal for pages/feeds/chrome. The current
   refusal string draws the line at the `/assets/` prefix as a proxy for "asset"; a ported tool's
   sidecar IS an asset of its page by any other measure. Reference-audit guard applies unchanged
   (this lane's pre-check: only the retired slot referenced the file, by relative `src`).
2. Teach page-retraction about sidecar files that belong to a page's deploy set — heavier: it would
   need a manifest of what webdesignport uploaded per page, which nothing records today.
3. A `webdesignport --retire <page>` mode — symmetric with the importer that created the files, but
   builds retirement into a tool whose whole design is idempotent CREATION.

Candidate (1) is the lane's recommendation: smallest reviewed change, the agent's five-guard chain
and dry-run default already carry the safety burden.

## Interim (recorded in the lane RUNBOOK)

Phase C proceeds; per tool the lane (a) proves the served page dropped the `src` reference (the
serve-grade negative), (b) dispatches the dry-run retraction and records the refusal as the
evidence the file is orphaned-but-present, (c) lists the orphan in the lane NOTES. Cleanup rides
whichever candidate ships; the orphan list makes it one batch. Orphans so far (as of 2026-08-22):
`/tools/blueprint-compiler/script.js` · `/tools/image-optimizer/optimizer.js` (refusal corr `1acae77f`,
same refusal text, orch by payload).
