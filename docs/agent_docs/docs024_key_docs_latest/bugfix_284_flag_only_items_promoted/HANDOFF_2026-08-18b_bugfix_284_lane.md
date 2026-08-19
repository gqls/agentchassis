# HANDOFF 2026-08-18b — bugfix_284 lane: what is finished, what is live, what is left

> **There are TWO handoffs in this directory and you want both.**
> `HANDOFF_2026-08-18_continue_here.md` is the **279/284/290 thread's** cold-start and
> covers four bugs, an RFC verification and a site item — read it first for the wider
> picture. **This** file is narrower: the `284` lane only, and it exists because I
> clobbered that file with a shell redirect and split the two apart on recovery (see
> WRONG_CALLS 2026-08-18). Nothing was lost; both are intact.

Written for a session that has never seen this lane. Read this file, then only the
parts of the five standing docs it points you at. **The bug this lane was opened for is
CLOSED and its fix is live and proven; everything still open belongs to OTHER lanes, and
this file's main job is to stop you re-deriving what was already measured.**

---

## 1. The one-paragraph version

`bugs_closed/284`: the platform files two kinds of finding — **jobs** (a named agent can
fix it) and **flags** (nothing on the platform can act, so the row names no handler *on
purpose*). `TriageDetectedItemsAction` promoted every `detected` row on a site without
looking at `handler_agent`, so flags were promoted, claimed, and stamped `blocked` with
*"No handler_agent set — item cannot be routed to any agent"*: a correct finding rewritten
as a routing failure, permanently (`blocked` is not terminal, so the row also held its
dedup slot and its check could never re-file). 60 rows, 4 item_types, 15+ sites. The fix
put the **claim path's own routability test at the promoter**, rendered from one shared
function, so the bad outcome is unreachable rather than merely unlikely.

## 2. State, with how each was proven (do not re-derive these)

| thing | state | proof |
|---|---|---|
| The promoter guard | **LIVE** since chassis `v1.0.1305`, still in `v1.0.1309` | image label `org.opencontainers.image.revision`, local `RepoDigests` matched to the pods' `imageID`, then `git merge-base --is-ancestor 7027a2801 <revision>` |
| Guard **exercised** in production | **YES** | single-step `triage_detected_items` on `leopardessconsulting.co.uk` (36 flag-only rows, nothing routable → could only hold or promote): `promoted: 0`, `not_promotable: 36`. Corr `a5be3dea-3f2c-490a-9922-22993662bc95` |
| The 60 damaged rows | **REPAIRED** | migration `442` (counted needles, `RETURNING` postcondition, `result.repair_284` stamped so none looks spontaneously fixed) |
| Born-dispatchable hole | **CLOSED** | migration `443`, CHECK `swi_no_handlerless_promotable`, `NOT VALID` then `VALIDATE`d, induced with two negative controls |
| Council | **APPROVED** round 2, verdict read | corr `c22998e8-41df-4145-a7b9-f132a7c77426`; all 4 advisories answered |
| Owner tie-break (predicate copies) | **RULED (a) and LIVE** | `10fc61184`; one definition fleet-wide; register **WDS-017** |
| Standing check, re-measured 2026-08-19 on `v1.0.1314` | **0 blocked FLEET-WIDE of any cause / 722 flag-only held** | see NOTES 2026-08-19 |
| **Lane status** | **CLOSED 2026-08-19 — nothing left on it; do not staff it** | §8 below |

**The standing check, and the control that makes it mean anything:**

```sql
SELECT count(*) FROM site_work_items WHERE status='blocked' AND handler_agent='';        -- must stay 0
SELECT count(*) FROM site_work_items WHERE handler_agent='' AND status IN ('detected','deferred'); -- 723 and rising
```

A zero alone proves nothing here — `improvement-sweep` (the only *scheduled* driver of the
only carrier) is `enabled=false` by the owner's `bugs_open/083` ruling of 2026-08-15, and
the promoter that runs unattended today (`detected-item-promoter`, created 2026-08-15)
never had the defect. The second number is why the zero is evidence.

## 3. What is LEFT — all of it in other lanes, none of it this lane's to execute

**(a) Five case-study images cannot be generated at all — a surface-list gap.**
`check_content_image_missing` is the framework's content-hero producer; its population is
hardcoded to `PageType: "blog-post"` (:131) and `"tool"` (:137), swept via
`AND p.page_type = $2` (:223). The case-study pages on `finetuning.uk` and
`ai-agent-orchestration.com` are **`page_type='content'`** — outside it by construction,
so no run of that check can ever emit their imagery, while the page markup references
`/assets/images/case-study-*.jpg`. Recorded in `bugs_open/114` with file:line.
**Do not hand-file `needs_imagery` rows for them** (the framework composes those prompts
from the page's own title/description — a session writing them is the 2026-08-06 ruling's
exact prohibition, and it hides the gap). **Do not widen the surface list unilaterally**:
two lines, but every `page_type='content'` page on every site enters the generator at
once. Count the population first.

**(b) 17 image findings need publishing or repointing, not new art.** Census (live table,
2026-08-17; still 41 open findings today, unchanged): 11 findings already have the asset
under **exactly** the referenced `asset_key` (4 `hero`, 2 `logo`, 5 `case-study-*`) — the
deploy to the canonical web path never ran; 6 are `hero` where the site's heroes are
page-scoped (`lendzy.co.uk` has 9, keyed `hero_home`/`hero_about`/…, none keyed plain
`hero`) so nothing lands at the base path a page requests. 8 more are `favicon`/`og-card`
and belong to **`bugs_open/131`** (owned). Full table in `bugs_open/114`.

**(c) One owner-raised row is queued behind `bugs_open/227`.**
`needs_experience_plan` on `fundamentallyai.com`, `spec.raised_by = "owner, reading the
live site 2026-08-12"`, parked at `deferred`. **`experience-planner` exists and is
active** — so routing it is a one-field change (`handler_agent`), not a build. The owner
ruled **(b): fix 227 first**, because that agent's prompt hardcodes one site's diagnosis
and would produce another site's plan. Dependency written into `bugs_open/227`.

## 4. The verification recipe that cost this lane a day — use it, do not re-invent it

**To answer "did my change ship?", read the image's own label. Do not probe for a needle.**

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.status.containerStatuses[0].imageID}{"\n"}{end}' | sort -u
docker image inspect aqls/agent-chassis:<tag> --format '{{json .RepoDigests}}'   # must contain that digest
docker image inspect aqls/agent-chassis:<tag> --format '{{json .Config.Labels}}' # → org.opencontainers.image.revision
git merge-base --is-ancestor <your-commit> <that revision> && echo SHIPPED
git rev-list --count <that revision>..HEAD                                       # what is still unshipped
```

Why: the `build provenance` log line is a STARTUP line and scrolls out of reach within
hours on this service, and binary needles go wrong in two ways this lane hit —
**a needle taken from source may be inside a `//` comment** (stripped at compile time, so
it can never fire: I published a "nothing shipped" claim on exactly that, and the
conclusion only survived because unchanged tag + identical digest carried it), and **a
`kubectl exec` loop killed by a tool timeout prints `absent`** for the needle it died on.
If you must probe: one `exec`, every needle inside it, a control that must come out the
other way, and draw the positive control from the **same commit** as the discriminating
needle.

**Also: a "fresh build" at an unchanged tag ships NOTHING** — a same-tag rebuild serves the
node's cached image. Measured 2026-08-17: pods restarted, tag still `v1.0.1305`, digest
byte-identical, **215 commits unshipped**. Bump `IMAGE_TAG` before building. The three
rolls since: 215 → 42 → 28 unshipped.

## 5. Traps this lane hit, each with the check (all in WRONG_CALLS/LANDMINES too)

- **A census by struct-field VALUE misses the site that omits the field.**
  `grep 'HandlerAgent: ""'` found 16 producers and missed the two with the most damage
  (`check_image_url_404`, `head_essentials_missing`) — Go zero-values an omitted field.
  Ask the DB which `item_type`s hold `handler_agent=''`, then find their producers.
- **A marker KEY does not prove authorship; a VALUE can.** `spec.original_pipeline` has
  three writers. Two hardcode `"build"`; the promoter writes `to_jsonb(pipeline)`, so
  `design`/`content` is its signature. Enumerate every writer before resting a case on a
  key's presence.
- **`orchestration_states` keeps `COMPLETED` rows for ~2 days.** "Zero runs in five weeks"
  measured the pruner. Check retention **for the status you are asking about**.
- **A disabled cron does not mean a path cannot run** — single-step dispatch exercises it,
  and that is also the cheapest proof available.
- **Reading the two `capability_gap` producers at HEAD looks like it refutes the
  diagnosis** (they say `Status: "deferred"`): that status was introduced by 284's own fix.
  `git log -S` on a status literal before believing a producer's present-day shape.
- **A pathspec commit still takes a same-file passenger.** `git diff --numstat <file>`
  before AND after appending to any fleet-wide file; I swept another lane's WRONG_CALLS
  entry under my message and had to add a provenance note.
- **A gate tested only on its failing side is indistinguishable from one that always
  fails.** My repair script's `\set guard_commit` (psql client var) was read as
  `current_setting('myvars.…')` (server GUC) — it aborted unconditionally. Induce both arms.

## 6. Where everything lives

- `bugs_closed/284_HANDOFF_2026-08-15_…` — the case, with the correction box explaining
  that its own title names the wrong mechanism (nothing claims a `deferred` row; the rows
  were born `detected`).
- This directory: `PLAN_2026-08-16`, `NOTES_…` (append-only, newest last — read the last
  three entries), `RUNBOOK_…` (the queries, each with its gotcha), `README_where_we_are.md`
  (the owner's plain-prose log), `SUMMARY_2026-08-16` + `SUMMARY_2026-08-17`,
  `REPAIR_…sql` (**superseded by migration 442 — do not run**), `ROLLBACK_…sql`.
- Migrations: `sql_for_agents/442` (repair) and `443` (the CHECK constraint), both with
  ROLLBACK sidecars.
- Register **WDS-017** (`register/work-dispatch.md`) — the mechanism, its landmines, and
  which way the seat disagreement was resolved.
- `016b` §9 (*"A status that means 'unclaimable' to its PRODUCER may be a promotable
  queue"*) and its §10 row for 284.
- Code: `work_items_common.go` (`workItemRoutableSQL`, `countUnroutableDetected`),
  `discovery_checks/remit.go` (`HandlerRegisteredSQL` — **the single definition**, in that
  package because `actions` imports it and never the reverse),
  `triage_detect_items_action.go`, `triage_routability_guard_test.go` (three tests, each
  proven load-bearing by a named mutation).
- `bugs_open/291` — the sibling defect this lane found and filed (`tool-auditor` filing at
  `hitl-review`, an agent that never existed); another lane fixed it, and its arm reads 0.

## 7. Do NOT do these

1. Do not re-open the `improvement-sweep` question — the owner ruled it stays paused
   (`bugs_open/083`, 2026-08-15); triage got its own scheduled task instead.
2. Do not read a clean `blocked` census as proof of this guard without quoting the
   flag-only population beside it (§2).
3. Do not write a fourth rendering of the agent-registration predicate. Call
   `discovery_checks.HandlerRegisteredSQL`.
4. Do not repair `image_url_404` rows expecting them to self-clear — that check has **no**
   `CheckResult.Resolved` arm (0 sites, versus 1/1/5 for its three flag-only siblings), so
   its rows stay open until a human acts. Not a fault of the guard.


## 8. LANE CLOSED 2026-08-19 — the conditions that would reopen it

Everything this lane existed to do is done, verified and live across **four** chassis
rolls (`v1.0.1305` → `1307` → `1309` → `1314`). No session should sit on it. The
directory stays as the record; the five standing docs are complete and the bug is in
`bugs_closed/`.

**Reopen ONLY on one of these, and each is a specific reading, not a hunch:**

1. `SELECT count(*) FROM site_work_items WHERE status='blocked' AND handler_agent='';`
   returns non-zero — quote the flag-only population beside it (§2) or the number means
   nothing.
2. `swi_no_handlerless_promotable` disappears or shows `convalidated = false`
   (`SELECT conname, convalidated FROM pg_constraint WHERE conname='swi_no_handlerless_promotable';`).
3. A fourth rendering of the agent-registration predicate appears — i.e. anything other
   than `discovery_checks.HandlerRegisteredSQL` computing "does this handler exist".
   `grep -rn "FROM agent_definitions" --include='*.go' platform/ internal/ | grep -i exists`
   should show only that function and the core-manager admin API (a different binary,
   different question).
4. `improvement-sweep` is re-enabled — not a defect, but it is the first time this guard
   runs on its own cadence rather than by single-step dispatch, so the first sweep is
   worth watching once (`not_promotable` should be non-zero on any site holding flag-only
   findings; today that is most of them).

**Everything else that came out of this lane belongs elsewhere and is written up there:**
`bugs_open/114` (the image census + the `page_type='content'` surface-list gap),
`bugs_open/131` (favicon/og-card), `bugs_open/227` (the owner-raised row's one-field
routing, held by the owner's decision until 227's prompt fix lands), and `bugs_open/291`
(filed by this lane, fixed by another, arm reads 0).
