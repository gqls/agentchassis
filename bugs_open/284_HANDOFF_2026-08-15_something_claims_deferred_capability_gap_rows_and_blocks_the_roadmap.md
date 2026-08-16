# 284 — something claims `deferred` capability_gap rows, turning parked roadmap entries into `blocked` — and it is still happening

> **DIAGNOSED 2026-08-16 — and the title is wrong, so read this box first.**
> **Nothing claims a `deferred` row.** The blocked rows were never `deferred`: they
> were born **`detected`**, which is a promotable queue, and `TriageDetectedItemsAction`
> promotes every `detected` row on a site without looking at `handler_agent`. The
> class is also **three times the size recorded below** — 60 rows across **four**
> item_types, not 18 across one; `image_url_404` alone has 40.
> **Fix committed `7027a2801`** (routability guard at the promoter), council
> `c22998e8-41df-4145-a7b9-f132a7c77426`. **Stays OPEN**: inert until the next
> chassis roll, the 60 rows are not repaired yet, and one of the two paths is not
> closed. Full account: `docs024_key_docs_latest/bugfix_284_flag_only_items_promoted/`.
> The evidence below is sound and is left unedited — only its conclusion moved.

**Filed 2026-08-15** by the `bugs_open/279` owner-decision session, spun out of the
`bugfix_213` lane's contribution into 279 (its "candidate 3: stop capability_gap
being claimed at all — that is arguably the root"). **Status: OPEN, evidence
measured, ROOT CAUSE NOT DIAGNOSED — the claimer is unidentified.** Per the
2026-07-31 owner ruling: this file does NOT assert a structural root cause; it
records the measured symptom and the one mechanism read first-hand, and names the
`090` run as the next step. Grep found no existing bug on this mechanism
(`deferred.*claim` hits only 255/279, both different defects).

## The design being violated (first-hand read, both files)

`capability_gap` rows are the platform's "found work I have no handler for" shape
(`bugs_closed/077`). Two producers file them deliberately **non-dispatchable**:

- `discovery_checks/remit.go` `CapabilityGapItem`: status `'deferred'`, **empty**
  `handler_agent`, with the reason IN the spec: *"naming a real agent on an
  undispatchable row is an invitation for someone to promote it — which re-creates
  077 exactly"*.
- `write_audit_findings` (since `d6d56e540`): same pair, for unrouted audit
  categories.

The dispatch loader (`load_work_item_actions.go:651`) selects
`status IN ('triaged','approved')` — `deferred` is not loadable there, and the
`detected`→`triaged` promoter (`triage_detect_items_action.go`) touches
`status='detected'` only. **By every mechanism read so far, a deferred row should
be unreachable by the claim path. It is not:**

## `[MEASURED]` 2026-08-15 — 18 blocked capability_gap rows, and the bleed is ongoing

```sql
SELECT count(*), count(DISTINCT site_id) FROM site_work_items
WHERE item_type='capability_gap' AND status='blocked';           -- 18 rows, 14 sites
SELECT DISTINCT left(error,60) FROM site_work_items
WHERE item_type='capability_gap' AND status='blocked';
-- "No handler_agent set — item cannot be routed to any agent"
```

That error string is `claim_work_item_action.go` (~:165) — the claim path's
empty-handler branch. So each of these rows was **claimed**, found handler-less
(as designed), and stamped `blocked`. The filed→blocked timeline shows it is a
standing mechanism, not an incident: rows filed 07-28 through 08-10 were blocked
on 08-02, 08-03, 08-04, 08-05, 08-08, 08-09, 08-10 and 08-11 — new blocks every
few days, `created_by` spanning `completeness-discovery-agent`,
`design-discovery-agent` and `generic`.

## Why it matters (three costs, each measured or read)

1. **The roadmap loses its entries' meaning.** `deferred` + empty handler is the
   honest "parked, awaiting a builder" state that `diagnose_triage`'s roadmap view
   groups; `blocked` means "work stopped by an error". 18 of the estate's
   capability gaps now read as failures.
2. **Until commit `d6d56e540`'s producer-scope fix, these 18 rows armed a
   site-wide mute** on 14 sites: `write_audit_findings`' broader blocked check
   collapsed to "ANY blocked capability_gap on this site" for its own new
   capability_gap filings (279's contribution section has the full mechanism).
   The mute is now fixed at that one reader; the blocked rows that armed it are
   this bug.
3. **claim burns an attempt on a row designed never to be claimed** — pure waste,
   repeated every few days.

## What is NOT established — the diagnosis gap, stated plainly

**What invokes ClaimWorkItemAction against a `deferred` row is unknown.** The two
live claimers by config census (`agent_definitions LIKE '%claim_work_item%'`) are
`build-dispatch-loop` and `diagnose-dispatch-loop`; the build loader's status list
excludes `deferred` at :651, but the file has other UPDATE arms (:1056, :1072,
:1089) not yet read, and the diagnose loop's loader has not been read at all.
Candidate mechanisms (all `[UNVERIFIED]`): a second loader with a wider status
list; a promoter arm that flips `deferred` under some condition; a workflow
passing item ids to claim directly.

**Next step is a `090` diagnosis run** (symptom: the mechanism above; point at
`site_work_items` rows `item_type='capability_gap' AND status='blocked'`,
`claim_work_item_action.go`'s empty-handler branch, both dispatch-loop agent
definitions and `load_work_item_actions.go`). Filed here first because the queue
check comes first — check `needs_diagnosis` open items before firing.

## DIAGNOSIS 2026-08-16 — the mechanism, end to end, each link read first-hand

**The rows were born `detected`, not `deferred`.** Four discovery checks build a
`WorkItemSpec` with an empty `HandlerAgent` and `Status: "detected"`, each with a
comment asserting the row is deliberately non-dispatchable:
`check_palette_contrast.go:120-132`, `check_content_duplication.go:232-248`,
`check_site_unreachable.go:254-264`, `check_backend_unreachable.go:99-108` — and
`check_image_url_404.go:256-278, 297-310, 330-348`, which **omits the field
entirely** so Go zero-values it to `""` (invisible to any grep for
`HandlerAgent: ""`, which is how a census of this class misses its largest member).

`TriageDetectedItemsAction` (`triage_detect_items_action.go:161-173`, live in
`improvement-loop`) then promotes **every** `detected` row on the site —
`WHERE site_id = $1 AND status = 'detected'`, no item_type filter, no handler
filter. The loader selects `status IN ('triaged','approved')`, claim
(`claim_work_item_action.go:96-105`) claims on the same pair, and its
empty-handler branch (:159-180) stamps `blocked` and NULLs `claimed_by` /
`claimed_at` — which is exactly why all 18 rows below show a NULL claimer and
`attempt_count = 0`.

`ClaimWorkItemAction` never touches a `deferred` row: its UPDATE is
`WHERE status IN ('triaged','approved')`, so a `deferred` row returns
`sql.ErrNoRows` and exits at :107. **This file's premise is refuted; its evidence
is not.**

### `[MEASURED]` 2026-08-16 — 60 rows, four item_types, not 18 and one

| item_type | blocked | sites |
|---|---|---|
| `image_url_404` | 40 | 15 |
| `capability_gap` | 18 | 14 |
| `needs_experience_plan` | 1 | 1 |
| `page_rerender` | 1 | 1 |

A further **37** rows sit at `detected` with an empty handler today
(`head_essentials_missing` 36, `image_url_404` 1) and would join them on the next
triage of those sites.

### Attribution, by a value that discriminates rather than a key that does not

58 of the 60 carry `spec.original_pipeline` = **`design`** or **`content`**. The
promoter writes `to_jsonb(pipeline)` — the row's own pipeline — whereas the only
two other writers of that key (`site_admin_handlers.go` `HandleApproveWorkItem`,
`tool_acceptance_actions.go` `routeChromeFailures`) both hardcode the literal
`"build"`, so neither can have produced these rows. **The `090` run
(`d1477c1d-bca4-4ac9-806d-da860eb0014a`) is what found those two writers**; it
returned UNVERIFIABLE and named them in `next_scope`, which is what turned a
claim of "one writer, therefore proven" into a check that could have failed.

**The remaining 2 rows are a different path**: no `original_pipeline`, no
`triaged_at`, `created_by` naming sessions (`bugfix-189-verify`,
`contrast-front-113`) — hand-inserted, born dispatchable, with no handler. A
promoter guard cannot see that one.

### And nothing legitimate depends on promoting a handler-less row

Enumerated, not asserted: of the **150** handler-less rows that have ever reached
`complete` (11 item_types), **zero** carry `spec.original_pipeline` and **zero**
have a non-null `triaged_at`. Every one completed through retraction or
revalidation — never through dispatch.

## FIX — committed `7027a2801`, council `c22998e8-41df-4145-a7b9-f132a7c77426`

`TriageDetectedItemsAction`'s UPDATE now carries the claim path's own routability
test, rendered from one shared helper (`workItemRoutableSQL` /
`workItemHandlerRegisteredSQL` in `work_items_common.go`) that claim itself now
calls — so the two cannot drift into disagreeing. Held-back rows are counted and
logged by item_type (`not_promotable`), because a filter that quietly promotes
fewer rows reads exactly like a site with less work. The two `capability_gap`
producers now file `deferred`, the status that type means everywhere else.

Three mutation proofs, run 2026-08-16: delete the guard → the promotion test
fails; hand-write claim's predicate → the coupling test fails; drop the
empty-handler half → two tests fail.

**Behaviour change, stated:** a site whose only `detected` findings are flag-only
now reports `site_dispatchable=false`, so `improvement-loop`'s
`check_has_findings` takes `notify_scheduler_clean` rather than
`insert_rerender_item`. That is the honest answer, and it replaces a rerender plus
a row that could only ever be blocked. The branch reads `site_dispatchable` (read
from the live `agent_definitions` row), not the call-scoped `has_items`, so
`bugs_closed/150` is not reintroduced.

## COUNCIL — APPROVED at round 2 (`c22998e8-41df-4145-a7b9-f132a7c77426`), verdict READ

**Round 1: REVISE**, gated by `guardian`. Nothing disputed the mechanism; the
objections were about what had been *shown*, and two were right:

- `prior_art_librarian`: `spec->>'original_pipeline'` cannot name which CHECK wrote
  a row — it is the row's own pipeline, one label over several producers (the
  `audit_source` landmine shape). **Correct.** Re-measured on the per-producer
  marker: `spec->>'check'` → `content_duplication` 9 rows/9 sites,
  `palette_contrast` 9 rows/9 sites. Exactly the two files edited, nothing else.
- `editquality`: "three of six producers got it wrong" was loose — and the precise
  enumeration found a **sixth** producer, `check_site_structural_validity.go`'s
  `head_essentials_missing` (36 live rows), which **also omits the field**. Two
  omitters, not one.
- `guardian` (gating): "semantically identical" was asserted without quoting the
  query claim runs. At `7027a2801^`: `SELECT EXISTS(SELECT 1 FROM agent_definitions
  WHERE type = $1 AND deleted_at IS NULL)` — no `is_active`, no `is_snapshot`, in
  the before or the after; and the seat's own answered check found **zero** live
  `handler_agent` values pointing at an inactive or snapshot definition.

**Round 2: APPROVED**, 4 advisory objections, none high-severity. Answered rather
than waved through:

- *`debug_historian`: 016's "Missing handler agents" may overlap.* **Ruled out** —
  that case is a handler NAMED but unregistered (claim's OTHER block branch). This
  is a handler deliberately absent. Same family, different member; the guard covers
  both branches, so it is a superset of that case, not a duplicate of it.
- *`debug_historian`: the 60-row repair is prose with no needle-gate discipline.*
  **Written properly**: `REPAIR_2026-08-16_blocked_flag_only_rows.sql` +
  `ROLLBACK_…sql` in the lane directory — a `DO`/`RAISE` gate refusing to run until
  the pod-verified commit is named, a mechanically-counted pre-state, a row-level
  backup, idempotent per-type UPDATEs, a verify that RAISES, and an induction
  recipe to prove the verify could fail.
- *`debug_historian`: no pod-verification step named.* It is the repair's own
  precondition and the RUNBOOK's first block: the `build provenance` line per
  SERVICE, then `git merge-base --is-ancestor 7027a2801 <stamp>`.
- *`tooling_provenance`: the deferred work needs a `doc_notes` row so the next
  session does not re-derive it.* **Written** — `subject_key`
  `site_work_items:flag_only_rows_promoted_and_blocked`.
- *`reuse_agent`: does the scheduled promoter already report held-back counts?*
  **No** — its `pre_query` returns `promoted` and the promoted pairs only.
  `countUnroutableDetected` duplicates nothing.
- *`guardian`: what if that second query errors?* It logs `Warn` and the promotion
  result stands; a failure to COUNT is not a failure to promote.

### ⚠ OWNER DECISION — two seats point in opposite directions on the same edit

`reuse_agent` (medium) objects that a **third** copy of this predicate —
`discovery_checks/remit.go` `HandlerStepConfig`, coupled to claim only by a prose
comment that names claim as its source of truth — was identified and **not**
migrated onto the new shared renderer, leaving three renderings of one test.

`guardian` (medium) objects that touching `claim_work_item_action.go` **at all** was
more than the bug required: the round-2 byte-diff shows the refactor is inert, which
argues it is *safe*, not *necessary*, and its stability preference is to restore the
inline literal and keep the helper only in the promoter.

One seat says the unification did too little; the other says it did too much.
Per the 2026-07-28 ruling — *"let a human break it, especially when seats disagree
with each other"* — **nothing was done unilaterally.** The approved code stands as
committed. Either direction is a one-file forward commit whenever you rule.

## WHY THIS STAYS OPEN (three things, in order)

1. **Inert until the next chassis roll.** Go changes do not ship on a commit.
2. **The 60 rows are not repaired**, deliberately — repairing before the guard is
   live means they re-block on the next claim. Then: `capability_gap` → `deferred`,
   `image_url_404` → `detected`, and the 2 hand-inserted rows judged individually.
3. **The hand-insert path is NOT closed.** What closes it is a CHECK constraint
   making `(handler_agent = '' AND status IN ('triaged','approved','claimed'))`
   unrepresentable — every writer, including the ~20 raw `INSERT INTO
   site_work_items` sites that bypass the shared door, and any manual insert.
   **It must land AFTER this binary rolls**: DB config is live immediately while Go
   is inert until the roll, so a constraint arriving first makes the OLD promoter's
   blanket UPDATE error on any site holding a flag-only `detected` row — breaking
   `improvement-loop` fleet-wide. There are currently **0** rows in that state, so
   it has no existing violators to repair first.

## Fix candidates, ordered by what closes the door (pending the diagnosis)

1. **Make `capability_gap` unclaimable at the claim path** — a guard in
   ClaimWorkItemAction (or the loaders) that refuses the type, or refuses any
   `deferred` row, rather than blocking it. Closes the door whatever the claimer
   turns out to be; the 213-lane contribution in 279 proposed exactly this.
2. **Repair the 18 rows** back to `deferred` once (1) ships — until then they
   re-block on the next claim. Do not repair before the mechanism is closed.
3. NOT here: the write-time vocabulary residual (`create_work_item` accepts any
   item_type from workflow config) — recorded in 279's status updates; different
   door.
