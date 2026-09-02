# NOTES — bugfix 440 (append-only, newest at the bottom)

## 2026-09-02 — lane opened; evidence gathered; one wrong call caught

Spun out of 410's candidate 1 (owner decision). Evidence base assembled first-hand — census,
probe, migration reads — all `[MEASURED 2026-09-02]`, recorded in the bug file. Missteps:

- **Wrong call, caught before filing but after saying it in-session**: read "2 rows contain the
  warning string" as "warning fired twice in production". Both rows were the 404 lane's council
  runs QUOTING the string in their submission payload. Caught by reading one member row
  (`current_step = complete_revise/complete_approved`). Logged in WRONG_CALLS 2026-09-02. The
  corrected finding (zero production firings + prose minted via migrations that bypass the
  creator) became the load-bearing "many doors" argument — the error, chased, was worth more
  than the number.
- Side effect of the same read: learned 404's r4 is `complete_approved` — their design is
  through. Their session has not yet read/recorded the verdict; nothing of theirs touched here.

## 2026-09-02 (later) — phase 1a built, mutation-proved, submitted (r1 corr 55def842)

- `rerender_routing_key{,_test}.go`: three-state resolver + two paste-target clause renderers.
  All four tests green; both named mutations run and killed by exactly their named test
  (unknown→assemble mutation → UnknownRefuses red; absent→refuse mutation → AbsentAssembles
  red). Refusal proven with REAL census unknowns (`tool_retirement`,
  `light_palette_chrome_replaced`), not synthetic values alone.
- ⚠ `platform/livespec` package run FAILS regardless of this change:
  `TestNoNewMigrationFileReadersOutsideTheAllowList`, the 405 lane's committed breakage
  (`ffa1707b3`, 7 days, documented in 404's NOTES). Reproduced identically without my files.
  Not touched — their file, their lane.
- 097 gotchas hit, for the next submitter: `operation` is an enum (learned 08-31, remembered);
  NEW one — a sketch whose every line starts `#`/`//`/`--` is refused as COMMENT-ONLY, which a
  markdown `###` heading triggers. Sketch the content lines, not the heading.
- REB-008 registered in the same commit (status: BUILT AND INERT BY DESIGN, with the
  do-not-overread warning).
