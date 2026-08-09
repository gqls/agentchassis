# PLAN — bugfix 230: give site discovery a recurring driver (fair rotation, observe-only)

**Started 2026-08-09.** Owning thread: this one (bug was filed OPEN, UNOWNED by the
`bugfix_201_page_content_writer_dispatch` lane, which closed out; ownership checked via
`who-owns.py` + live transcript grep — the only session mentioning 230 substantively is the
brochure lane, whose handoff claims **215**, not 230).

**Bug:** `bugs_open/230_HANDOFF_2026-08-09_site_discovery_has_no_recurring_driver_so_a_site_stops_being_examined.md`
**Sibling (explicitly NOT acted on here):** `bugs_open/083` — the *drain* side (promotion of
`detected` findings). Its owner decision is recorded there as PENDING and this plan does not
touch it.

---

## 1. The problem, re-verified 2026-08-09 (all live `clients_db`, this session)

- `count(*) FILTER (WHERE enabled)` = **0 of 5** scheduled rows targeting the three
  site-discovery agents; all five are `oneshot-*`, fired once, left disabled. `[MEASURED]`
- `improvement-sweep` (the designed driver, register IMP-016: *"intentionally paused during
  core build"*) still `enabled=f`, last fired 2026-05-02. `[MEASURED]`
- Detection demonstrably follows attention: every discovery orchestration in the 24h
  retention window (improvement-loop cycles at vetcomparison 08-08 14:33, leopardess 17:12,
  webdesign 18:08, dartsonline 08-09 08:57) is un-prefixed / irregular — hand-fired by lanes
  at the sites they were already working. `[MEASURED]`
- The worked case stands: `finetuning.uk` pages `ai-guides` + `insights`, slot
  `featured-content`, 334-byte components — `findEmptySections`' own predicate returns both
  pages today and **no `empty_section` item exists for that slot**. `[MEASURED]`

## 2. What already exists (reuse, not invent)

- **The scheduler** (`cmd/scheduler/main.go`) supports everything needed with **zero Go
  changes**: `pre_query` produces dynamic `input_data` (first row as JSON); a no-rows
  pre_query is a clean stamped no-op; data-modifying CTEs in `pre_query` are an established
  pattern (`database-cleanup`, `stuck-task-reaper`).
- **The agents**: all three discovery agents start at `ensure_site_record` with
  `input_fields ["site_id","domain"]`; the proven trigger shape is the 08-03 oneshot
  (`target_topic='system.agent.scheduled.requests'`, `input_data={site_id,domain}`).
- **Observe-only is a designed mode**, not an improvisation: register `improvement-loop.md`
  — *"Findings insert with status='detected' (unclaimable), so a check can run observe-only
  while improvement-sweep (the triager) stays disabled."* Insertion is dedup-guarded
  (`idx_swi_dedup`), so re-examination is idempotent for open findings.
- **The watchdog pattern**: RFC_006's `single-owner-carriers-check` CronJob — stock
  `postgres:16-alpine` + configmap script, one `doc_notes` row per run **on findings and on
  clean** (a missing row = the job did not run), exit non-zero on findings.

## 3. Why NOT simply re-enable `improvement-sweep`

Three reasons, each measured or on record:

1. **It is the whole detect→triage→fix loop**, and the owner decision on draining findings
   is pending (`bugs_open/083` §"What this changes about the fix": *"Decision pending — do
   not act on this section until it is recorded here"*). Detection cadence is 230's scope;
   repair cadence is 083's owner call. IMP-016's gated re-enable sequencing also requires
   handler-existence checks that are exactly 083's open question.
2. **Its site selection starves** (register IMP-010, still live at HEAD): `ORDER BY
   s.updated_at ASC NULLS FIRST` — a site whose `updated_at` never moves wins every tick;
   nothing the sweep does advances the key it sorts on.
3. **Its backlog cap permanently exits the busiest sites**: the live pre_query skips sites
   with ≥50 open build items. Measured 2026-08-09: webdesign.co.uk **85**, dartsonline.com
   **79** — the two most-worked sites in the fleet would never be examined by the old driver
   even if flipped on today. Combined with 083 (queues that never drain), the cap converts
   "has findings" into "stops being examined", which is this bug's mechanism embedded in its
   own designed remedy.

## 4. The fix (candidate 1 of the bug file, made fair and observe-only)

### 4a. Rotation state — migration `346_site_discovery_rotation.sql`

```sql
CREATE TABLE site_discovery_rotation (
    site_id          uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    agent_type       text NOT NULL,
    last_selected_at timestamptz NOT NULL,
    PRIMARY KEY (site_id, agent_type)
);
```

Why a table and not an existing column: COMPLETED orchestrations are purged at **24h**
(`database-cleanup` step 3, read this session), so orchestration history cannot key a
7-day rotation; `sites.updated_at` is IMP-010's defect; `sites.content_data` is live site
content. The stamp is **selection**, not completion — deliberately, so a site whose run
fails cannot pin the rotation head and starve the fleet (the `bugs_open/048` shape the
scheduler's own comments warn about). Completion is the watchdog's question (§4c).

### 4b. Three `scheduled_tasks` rows (same migration)

One per discovery agent, `site-discovery-rotation-{quality,design,completeness}`:

- `target_topic='system.agent.scheduled.requests'`, `interval_seconds=3600`,
  `concurrency_group='site-discovery'`, `max_concurrent=1`, `enabled=true`.
- `pre_query` (per agent): pick the least-recently-selected deployed/active site whose
  stamp is older than **7 days** (or absent), skipping sites with a `claimed` build item
  *this tick only* (no backlog cap — the skip costs the site nothing because its stamp
  stays oldest); stamp it via data-modifying CTE; return `{site_id, domain}`.

Properties: a new site (no stamp) is picked first; a busy site is deferred, not excluded;
no site can fall out of examination while the task runs — the invariant 230 asks for.
Steady state ≈ 20 sites × 3 agents / 7 days ≈ **9 runs/day**; cold start drains in <1 day
at hourly ticks. Measured LLM cost of one full cycle (dartsonline, 08-09 08:57): **2 calls,
~4.2k in / ~2.1k out tokens** (visual + content audits; the deterministic checks are SQL).
Worst case ≈ **6 LLM calls/day** fleet-wide. `[MEASURED, single cycle — may vary by site]`

### 4c. Watchdog — `site-discovery-staleness-check` CronJob (daily 06:35 UTC)

Two questions, because the stamps are fire-and-forget (LANDMINES: *"last_triggered_at keeps
advancing while nothing runs"*):

1. **Coverage**: any deployed site with no stamp, or a stamp older than 2× period, for any
   agent → finding. Catches: tasks disabled (the quiet grave this bug is about), a
   perpetually-skipped site, a new site the rotation misses.
2. **Closers vs producers**: stamps advanced in the last 24h with **zero** matching
   discovery orchestrations in the same window (retention permits exactly this) → finding.
   Catches silent fire-drops (chassis restart window, topic breakage).

One `doc_notes` row per run, findings or clean; exit non-zero on findings. Same
image/secret/reporting as `single-owner-carriers-check`; applied with `kubectl apply -k`
like its siblings (stock image — not a fleet-release artefact).

## 5. Decisions and their reasons

- **Enabled=true at seed, conservative defaults.** Seeding disabled rows recreates the
  `oneshot-*` grave this bug documents. The cost question 230 §4 deferred is now measured
  (§4b) and small. The knobs (period, interval) are plain UPDATEs, named in the RUNBOOK, and
  the owner can turn any of it off with one UPDATE. Council reviews before anything is
  applied.
- **Detection only.** Findings land exactly where each check already puts them. This plan
  neither promotes nor drains; 083 keeps its pending owner decision. Consequence stated
  plainly: the `detected` pile will grow with true findings for types whose promoter is
  parked. That is the honest state of a fleet whose examination works and whose repair
  cadence is a pending decision — the record exists instead of silence.
- **No fourth `triage_detected_items` carrier** (RFC_006 SingleOwner; the `has_items`
  landmine). This plan adds no carrier at all.
- **`improvement-sweep` row untouched.** It is the owner's paused mechanism; its two
  defects (§3.2, §3.3) are recorded in 083/230 and the register rather than "fixed" inside
  an unrelated change.

## 6. Verification (bug 230 §6, unchanged)

The `featured-content` slot on finetuning.uk's two pages is an outstanding, undetected true
positive. After the completeness rotation reaches finetuning.uk (≤20 hourly ticks after
apply), an `empty_section` item for that slot must appear **without anyone dispatching it**.
Negative control: the stamp table must show every deployed site selected within the first
day; the watchdog's first run must write its doc_notes row.

## 7. Council + registration

- Council submission before/alongside commit (correlation recorded in NOTES + commit
  trailer `Council-Submitted:`).
- Concept register: entry in `scheduler-and-tasks.md` (the rotation mechanism + table) and
  a status correction in `improvement-loop.md` (IMP-016 gains the detection-cadence half;
  IMP-010's starvation gets its 2026-08-09 measurement), same commit — the ordering
  exemption's condition (2).
- Consumers told, not merely measured (owner ruling 2026-07-29 §3): contribution notes into
  `bugs_open/083` (its "second-order effect" — stranded shipped detectors — starts moving
  again) and `bugs_open/093` (its live-but-never-run stat audit gains a runner), and the
  finetuning lane's worked case is the verification canary.
