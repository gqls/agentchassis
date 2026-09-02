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

## 2026-09-02 (later still) — phase 1a verdict READ: APPROVED, 2 advisories none high, 4 abstained

Corr `55def842`. Dispositions, so nothing dangles:

- **editquality [medium] "nothing calls the resolver; the defect is unchanged by this commit"** —
  TRUE AND BY DESIGN, conceded in the submission's own first paragraph ("INERT foundation only").
  This round never claimed to change production behaviour; the flip is RFC_062 phase 3. Their
  MISSING items (gate untouched, creator untouched) are the stated phase plan, not omissions.
- **reuse_agent [medium] + prior_art_librarian [medium], the real ask: the three claimed reuse
  targets were not SHOWN to exist** — fair, and answered at the artefact `[MEASURED 2026-09-02]`:
  `platform/livespec/rerender_reasons.go` — `var RerenderSectionReasons` (symbol area ~:129),
  `func CheckRerenderModeConditionClause()` (~:157), `func RerenderSectionReasonByName` (~:168);
  and structurally, `rerender_routing_key.go` CALLS the first and third in the same package and
  the package compiles with all four tests green — a compile is an existence proof grep cannot
  fake. The gate-config read prior_art_librarian wanted attached is the RUNBOOK's own query; its
  output (single conditional, five-value `==` disjunction, then `rerender_sections` / else
  `render_page`) was captured in this session before RFC_062 was written and is quoted in the RFC.
  Lesson for r-next submissions: attach the grep/query OUTPUT, not the claim that it was run —
  the same submission-accuracy failure class 404's r3/r4 just paid for.
- **bug_historian [low]** (missing-vs-empty deferred): stated as a BLOCKING caveat in the file
  header, RFC_062's open list, and the submission risks; phase 3 cannot pass its own round
  without discharging it. Deliberate, not overlooked.
- **guardian [low]** (name that livespec is shared cross-pipeline; future wiring out of this
  round's scope): correct, and now explicit here — any wiring into `agent_definitions` workflow
  steps is phase 3, RFC_062, its own round.
- **architecture [low]** (register must forbid a second producer of `RoutingReasonSpecKey`
  pre-RFC): ACTIONED — REB-008 amended in place, this commit.
- **architecture MISSING ("is this the Nth deflection of the gate to a higher layer?")** —
  answered: it is the FIRST raise, not a deflection. The deferral was stated ONCE (the 404
  lane's livespec header, through their approved rounds) and this lane's response to it was to
  CREATE the RFC with an owner and build the foundation the same day. A deflection leaves the
  higher layer unwritten; RFC_062 exists.

Commit `a3758c399` carries `Council-Submitted:` and is auto-credited now the verdict is approved.
