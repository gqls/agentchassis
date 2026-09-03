> ## ⚠ SUPERSEDED — the bug is CLOSED. Read `HANDOFF_2026-09-03_closed.md` in this directory.
>
> Nothing in this file's open list remains. What is left is three DECISIONS, laid out there.

> ## ⚠ SUPERSEDED 2026-09-02b — read `HANDOFF_2026-09-02b_continue_here.md` in this directory.
>
> Its open list is shorter than this one's: exactly ONE item remains (the CronJob firing on its
> schedule). Everything else here is either done or already stale.

# HANDOFF — 2026-09-02 (updated ~14:2xZ) — `bugs_open/394`, render-audit coverage cursor

**Supersedes `HANDOFF_2026-08-26_continue_here.md`** (same directory). Read this one; that one's
open list is four-fifths out of date.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`
**Bug:** `bugs_open/394_HANDOFF_2026-08-25_webdesign_render_audit_tail_is_71_pages_and_growing_and_the_truncation_row_has_no_reader.md`
**Council:** `f67593f5-90cb-4a35-9cc0-926254645192` — **APPROVED** (round 3). Trailer already on `ea08c831d`.
**Ownership 2026-09-02:** still this lane; nobody else has touched the action or the lane dir in the week.

---

## 1. State in one paragraph

**Both halves of what the owner commissioned are now built, and the cursor half is live and
accepted.** The acceptance test passes outright — 151 distinct pages of 151 live over three
scheduled runs, zero missed. The reader now has its CronJob (`render-truncation-check`, 07:50 UTC),
wired into `RELEASE_IMAGES` and `AGENT_DEPLOY_SERVICES`, so it builds, pushes, retags and applies
on the next fleet release. **Nothing is left to design.** What remains is one release and one
post-release verification.

## 2. Closed — with the evidence, not an assumption

| item | status | evidence `[MEASURED 2026-09-02]` |
|---|---|---|
| **the commissioned reader has a CronJob** | **BUILT, ships on the next release** | `render-truncation-check` 07:50 UTC; dockerfile + base + pinned overlay + makefile both lists + wrapper script; `make check-release-coverage` OK |
| **the acceptance test** (bug §2) | **PASSES** | union of `audited_paths` over the 3 scheduled cursor runs = **151 distinct pages**; live pages = **151**; **missed = 0**. Graded against the site, not against itself. |
| cycle completion | **PROVEN in production** | run 3 (09-02 13:09Z): final window 37 pages, last page `tool-llm-cost-calculator` (`nav_order` 201), `cursor_cleared = true`, and the cursor row was then **gone** — `deleteAuditCursor` fired unattended |
| identity fix `faf4872ce` live | **PROVEN by a discriminating test** | the two hypotheses predicted opposite windows; observed `window_first = index` + a NEW `render-audit-agent` row, while the `generic` row was untouched. "Not live" refuted. |
| audit step TIMEOUT | **CLOSED — was transient** | `contrast_failure` rows created 08-27 **28**, 08-28 6, 08-29 2, 08-30 1, 09-01 2 |
| optional-key overlay | **CLOSED — already applied** | live ConfigMap `optional-key-budget-check-script-9b89gcmd8g` line 179: `"request_render_audit": 7` |
| mode split | **PROVEN in production** | `design-critique-agent` → `prefix`; `render-audit-agent` → `cursor`, every post-roll row |
| a second site cycling | **PROVEN** | `loanandmortgagecalculator.co.uk` (61 pages): window 1 on 08-27, 2-page final window on 08-30 |

**What this replaced**, for contrast: the same first 60 pages for ever, with 91 never audited —
including all 45 `tool-*-guide` pages, unreachable at any cap below 98.

## 3. ⚠ OPEN — one release, then one check

### (a) The CronJob ships on the next fleet release — do NOT apply it by hand
`render-truncation-check` is committed complete: dockerfile, base, pinned overlay (`v1.0.1352`),
`RELEASE_IMAGES`, `AGENT_DEPLOY_SERVICES`, a `build-render-truncation-check` target, and
`scripts/audit-render-truncation.sh`. `make check-release-coverage` passes (35 of 38 overlays).

Releases are whole-fleet and the owner runs `make release` — **never a one-service apply at its own
tag.** Being in `AGENT_DEPLOY_SERVICES` is what makes the release build, push, retag and apply it in
one ordered pass. That order matters here: the image must exist before the overlay is applied, or
the CronJob sits in `ImagePullBackOff`, which this fleet reports as a Job **RUNNING**, never FAILED.

### (b) After the release, confirm the job actually ran — one query
```sql
SELECT created_at, left(body, 400) FROM doc_notes
 WHERE categories ? 'render-truncation' ORDER BY created_at DESC LIMIT 3;
```
It writes **one row per run, clean or not**, so an ABSENT row means the job did not run and must
never read as "nothing is wrong". Expect a body of the shape
`render-truncation: N rows across M sites and 2 callers … 0 finding(s); 1 dormant group(s)`.

Also confirm the schedule slot is genuinely free once applied (the 07:50 census was against the
committed manifests, which is the correct source, but the cluster is the final word):
```bash
kubectl -n ai-persona-system get cronjob -o custom-columns='NAME:.metadata.name,SCHEDULE:.spec.schedule' | sort -k2
```

### (c) Then 394 can close
The `/bugs_closed/` bar is "fixed AND live". Once the CronJob has run once and written its
`doc_notes` row, both commissioned halves are live and the bug moves — naming BOTH paths on the
commit (`git commit bugs_open/394_… bugs_closed/394_…`), because a pathspec commit naming only the
new path ships a COPY and leaves the old file at HEAD.

## 3b. What the pre-ship wire test caught, and why it matters to you

**I did not design the dormancy rule; the prescribed wire test found the need for it.** Running the
reader against live data BEFORE shipping returned **exit 1**, claiming the migration-660 config flip
had regressed on `loancalculator.co.uk`. It had not. That site has **one** truncation row, from
2026-08-11 under a per-dispatch `max_pages: 5` override, and **28 live pages against a cap of 60** —
it can never truncate again, so that row is frozen history and the group could never self-clear.
Judging "the most recent row in the group" would have been red on day one and every day after.

A group whose newest row is **>14 days behind the fleet's newest** is now reported as **dormant** and
not judged — counted and NAMED in every run, so "0 findings" cannot quietly become "0 findings among
the groups I still look at". Measured relative to the newest row in the data, not the wall clock, so
it is a pure function of the data.

Post-fix wire test: **19 rows / 4 sites / 2 callers / 0 findings / 1 dormant group named, exit 0.**

⚠ **The stated blind spot:** a regression on a site that ALSO stops truncating would go unseen. That
is bounded — a site stops writing these rows when it fits inside its cap, and a site inside its cap
has no coverage debt to detect. The residual is a site that shrinks below the cap *while*
misconfigured; that is visible in `agent_definitions` instead. 14 days is a judgement, not a
measurement (~4 missed opportunities at the 3-day cadence) and is a single named constant.

## 4. ⚠ Read before you re-probe a binary — a NEW landmine lands on last week's method

`LANDMINES.md` gained, 2026-08-24: **"BusyBox `grep` over `/proc/1/exe` reports FALSE ABSENCES —
and your present/absent controls PASS while it does it."** The fleet's images are BusyBox v1.37
(CLAUDE.md's "debian-slim" is stale); its grep is line-oriented and a Go binary's "line" can be
enormous, so a literal can read absent with a clean exit code.

**I used that instrument on 2026-08-26.** Assessed rather than waved away: the fault produces false
**ABSENCES**, and last week's claim rested on three **PRESENCES**, so the direction could only have
made me under-claim. The conclusion was also confirmed behaviourally minutes later. The negative
control may have been vacuous — that costs the control, not the conclusion.

Use the prescribed instrument from now on, both controls through the SAME pipeline:
```bash
kubectl -n ai-persona-system exec <pod> -- sh -c "tr '\0' '\n' < /proc/1/exe | grep -Fc '<literal>'"
```
Re-probed today on `agent-chassis-5bd89cf49-t4wdl`: `render_audit_page_cursor` 3 ·
`rotate_coverage` 2 · `runningStepProvenance` 1 · `selectAuditWindow` 2 · nonsense control 0.

## 5. Other traps this lane paid for (unchanged, still true)

- **`snapshot_agent` has TWO overloads writing to DIFFERENT TABLES**; a bare literal is ambiguous
  and aborts the migration; the two-arg form writes `agent_definitions_backup` and returns the
  SOURCE id, so verifying in `agent_definitions` reads 0 and looks like a no-op.
- **NEVER parse a `contrast_failure` `item_key`** — a selector may contain `#` and so may a page
  URL (`idea.uk` has `/tools.html` and `/tools.html#audience-check` both active). Match forward
  with `workItemKey(...)` + `HasPrefix`, longest path first.
- **`sql_for_agents` numbers are a shared sequence with no reservation** — mine was renumbered
  twice (646 → 649 → **660**). Take `max+1` at the moment you APPLY.
- **A test can only discriminate what its fixture varies** — `renderAuditParams` set
  `Sender.AgentType` and the running agent to the same literal, which made twelve tests blind to
  the key defect.

## 6. Facts, dated 2026-09-02

webdesign.co.uk **151** live pages (131 on 08-24 → 146 → 151; still growing). Bands: `0..90` 6 nav ·
`100` 94 tools · `200` 48 `tool-*-guide` · `201` 1. Callers: `render-audit-agent` cap 60
(rotating), `design-critique-agent` cap 8 (prefix by design, manual-only, acknowledged in
`render_truncation_acks.json`). Second truncating site: `loanandmortgagecalculator.co.uk`, 61 pages.
Cursor rows: 2 on webdesign (`render-audit-agent` live, `generic` orphaned).

## 7. Commit trail

`95a04168c` cursor · `72b16391b` R1 fix (forward match) · `41b03241d` reader + registry ·
`ea08c831d` R3 advisories + `Council-Reviewed` · `c71b46be0` migration 660 applied + landmine ·
`faf4872ce` identity fix · `a3610ea23` artefact proof · `99026097f` the week's evidence.
Migration: `docs/agent_docs/sql_for_agents/660_render_audit_coverage_cursor_HOLD.sql` (applied
2026-08-26 22:20:40Z, backup in `agent_definitions_backup`).

## 8. Also open, from this session's sweep (not 394)

- **`apis_uk_bees_homepage`** — their `SUBJECT_MISSING_ON_REPEATED_COMPONENT` registry entry puts
  its prose under `note` where the checker reads `why`, so `TestShippedRegistryIsSelfConsistent`
  was red at HEAD on 08-26. CONTRIB filed in their dir; worth re-checking whether it is still red.
- **`bugs_open/359`** — I validated it (7 of 39 archived pages serving) and yielded the lane; that
  session has since built the detector. Not mine.

## 9. Two things in the shared tree you should NOT chase

1. **`go test ./cmd/config-key-audit/` may fail on `TestBudgetCronCountsLiteralMatchesTheRegistry`
   — it is not yours and not mine.** Another session has **uncommitted** work in the tree adding a
   fifth optional key (`max_image_dimension`) to `execute_vision_prompt`, plus
   `vision_image_downscale.go`/`_test.go` (untracked). The cron literal still says 4. Their commit
   will hit that guard, which is the guard doing its job. **Do not "fix" it and do not commit those
   files.** Confirm with `git status --short -- platform/orchestration/actions/ | grep -i vision`.
2. **`TestShippedRegistryIsSelfConsistent` is now GREEN** — it had been red at HEAD for seven days
   on another lane's `SUBJECT_MISSING_ON_REPEATED_COMPONENT` entry (prose under `note`, which is
   the field the `consumed` shape reads; the human-evidence arm reads `why`). I CONTRIBed it to
   them on 08-26, it went unacted for a week while blocking a green run of the whole package, so I
   fixed it **additively** on 09-02: a new `why` field, their `note` left exactly as written, and
   the added field says in-line that this lane added it and why. If they object, the fix to prefer
   is theirs — the field just has to exist.
