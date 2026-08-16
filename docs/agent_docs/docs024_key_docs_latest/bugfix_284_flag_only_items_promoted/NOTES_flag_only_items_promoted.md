# NOTES — bugs_open/284, flag-only items promoted and blocked

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-08-16 — session open, ownership check

`scripts/who-owns.py 284` returned **"VERDICT: OWNED or recently active"** — but the
only evidence it had was the commit that FILED the bug (`36aca20bc`, 2026-08-15).
That is the tool's known lag: it reads commits, so the filing session and an owning
session look identical. Checked the live transcripts instead
(`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`, modified in the last
5 h, grepped for `capability_gap|bugs_open/284`): the 279 lane's session had 104 hits
and its closing message says, verbatim, *"Open and unowned — needs a fresh session,
not this one"*. No other session had substantive hits. Taken on.

## 2026-08-16 — the bug is still live

```sql
SELECT status, count(*) FROM site_work_items WHERE item_type='capability_gap' GROUP BY status;
--  deferred 19 | blocked 18 | needs_human_review 1
```

18 blocked rows still present, unchanged from 284's count. `needs_diagnosis` queue
was empty (`status='awaiting_diagnosis'` → 0 rows), so no duplicate filing.

**Nothing has blocked since 2026-08-11 21:51.** Worth stating because it could be read
as "the bug has stopped": it has not. The last `capability_gap` row of any kind was
created 08-12, so there has been nothing new for the promoter to catch on that type.
The 37 `detected` rows with empty handlers sitting on the fleet today are the standing
exposure, and `head_essentials_missing` (36 rows, one site) was touched **today**.

## 2026-08-16 — root cause, and where my first read was too narrow

284 pointed at the claim path and asked "what claims a `deferred` row?". The answer is
**nothing does** — and that is the finding. `ClaimWorkItemAction`'s UPDATE
(:96-105) is `WHERE status IN ('triaged','approved')`, so a `deferred` row returns
`sql.ErrNoRows` and exits at :107 without touching anything. The blocked rows were
never `deferred`. They were born `detected`.

The evidence that settles it, and that could have come out otherwise:

```sql
SELECT status, count(*),
       count(*) FILTER (WHERE spec ? 'original_pipeline') AS via_triage_action,
       count(*) FILTER (WHERE spec ? 'not_dispatchable')  AS via_CapabilityGapItem
FROM site_work_items WHERE item_type='capability_gap' GROUP BY 1;
--  blocked   18 | 18 | 0
--  deferred  19 |  0 | 18
```

`spec.original_pipeline` is written by exactly one thing (`TriageDetectedItemsAction`,
:165-169); `spec.not_dispatchable` by exactly one thing (`CapabilityGapItem`,
`remit.go:184`). The two populations are disjoint and complementary. The blocked rows
came through the promoter; the deferred rows never did.

**So 284's framing — "something claims deferred rows" — is refuted, and its title is
wrong.** The bug is real and its evidence is sound; the mechanism is one step upstream
of where it looked. Corrected in the bug file rather than silently.

## 2026-08-16 — the class is bigger than the type 284 named

Chasing `capability_gap` alone would have fixed 18 of 60 rows. The census that caught
it asked about the ERROR, not the type:

```sql
SELECT item_type, status, left(error,55), count(*) FROM site_work_items
WHERE status='blocked' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

`image_url_404` has **40** blocked rows — more than twice `capability_gap`'s 18 — and
its producer does not even set the field: `check_image_url_404.go:256-278` OMITS
`HandlerAgent`, so Go's zero value supplies `""`.

**Misstep, recorded because it nearly set the scope:** my first census was
`grep -B12 'HandlerAgent: *"",'`, which finds only the sites that write the empty
string EXPLICITLY. It found 16 producers and missed the one with the most damage. A
zero-valued struct field is invisible to a grep for its value. The check that works is
to ask the DB which `item_type`s actually hold `handler_agent = ''`, and only then go
looking for their producers — the data knows about fields the source does not mention.

## 2026-08-16 — the platform already had the correct predicate

Found while reading the concept register (`register/scheduler-and-tasks.md:231`), then
read live rather than trusted:

```sql
SELECT pre_query FROM scheduled_tasks WHERE name='detected-item-promoter';
```

The live `detected-item-promoter` (enabled, 900s, `fire_message=false` — the CTE is
the worker) promotes only where `COALESCE(wi.handler_agent,'') <> ''`, the handler is a
live active non-snapshot `agent_definitions` row, and the `(item_type, handler_agent)`
pair has ≥1 lifetime `complete`. So the estate has TWO promoters with different rules,
and the Go one — the one inside `improvement-loop` — is the permissive one.

That reframes the fix: this is not "invent a guard", it is "one of two promoters is
missing the predicate the other one already states".

## 2026-08-16 — `090` filed before asserting

Per the 2026-07-31 owner ruling, a cross-cutting root cause is not filed until it has
been through the diagnosis loop or the session states why it substituted first-hand
verification. Filed rather than argued the exception: intake correlation
`ce78493f-0e53-4805-bb99-a9a75dd307aa`, run correlation
`d1477c1d-bca4-4ac9-806d-da860eb0014a`. The symptom names the mechanism and points at
the tables and symbols; it asserts no counts, and it asks the one question I could not
answer from reading alone — whether a SECOND path also reaches the claim.

## 2026-08-16 — the `090` came back UNVERIFIABLE, and it was worth every minute

Outcome `UNVERIFIABLE` — it ran out of iterations before confirming, and its
`next_scope` named ten symbols. That is not a wasted run: **it refuted a claim I
had already written down.**

I had written, in this file and in the bug file, that `spec.original_pipeline` is
*"written by exactly one thing"*. The loop's citations name **three** writers:
`TriageDetectedItemsAction`, plus `internal/core-manager/admin/site_admin_handlers.go`
(`HandleApproveWorkItem`) and `platform/orchestration/actions/tool_acceptance_actions.go`
(`routeChromeFailures`). Read both: each hardcodes the literal `"original_pipeline": "build"`
into a spec it constructs. So the KEY does not discriminate.

The **value** does, and the corrected check is stronger than the one it replaces:

```sql
SELECT item_type, spec->>'original_pipeline', count(*) FROM site_work_items
WHERE status='blocked' AND handler_agent='' GROUP BY 1,2;
-- capability_gap|design|15 · capability_gap|content|3 · image_url_404|design|40
-- needs_experience_plan|(null)|1 · page_rerender|(null)|1
```

The promoter writes `to_jsonb(pipeline)` — the row's OWN pipeline — so `design`
and `content` can only have come from it; a literal writer would always say
`build`. 58 of 60 attributed, on a value that could have come out otherwise.

**And the two rows with no key at all are a second path I had not found.** Their
`created_by` names sessions (`bugfix-189-verify`, `contrast-front-113`), their
`triaged_at` is NULL: hand-inserted, born dispatchable, no handler. The promoter
guard cannot see those. Recorded in the bug file as the sequenced follow-up (a
CHECK constraint, after the roll) rather than quietly left out of the story.

**Reading the verdict was itself a trap.** The work item's `result` column held
`{"role":"handler","agent_id":…,"agent_type":"diagnose-orchestrator"}` — the SPAWN
record, not the verdict (the `bugs_open/287` shape). `diagnosis_artifacts` held two
`bundle` rows and nothing else, which looks exactly like the >60KB no-verdict
failure — but its budget line was absent, so that was not it. The verdict lives in
`orchestration_states.collected_data->'verdict'`, on the THIRD orchestration row of
the correlation. Recipe now in the RUNBOOK.

## 2026-08-16 — near miss: the scratchpad root is a shared git repo

Running the isolated-tree test I typed `cd $SCRATCH && git checkout 2>/dev/null`
to reset something. `git rev-parse --show-toplevel` from there returns
**`/home/ant/.claude-scratch/claude-1000`** — one repo holding **every** session's
scratchpad, 55 files dirty across ~40 sessions at that moment. Bare `git checkout`
takes no action without a pathspec, so nothing was lost (verified after: still 55
dirty). `git checkout .` or `git restore .` from that directory would have
destroyed 55 other sessions' working files, and `git stash`'s hook ban does not
cover either. Logged in LANDMINES.

## 2026-08-16 — Fable was asked for the plan and could not run

The owner's instruction was to prepare the plan with Fable. The agent terminated on
`You've reached your Fable 5 limit. Run /usage-credits to continue`. No plan was
returned — recorded here so nobody later reads the plan in this directory as Fable's
work. Planning was done in-session; re-runnable when the limit resets.

## 2026-08-16 — what the guard actually does to today's fleet, with a demand control

Predicate run against the live table before the roll, so the effect is measured
rather than predicted. **Held back: 37 rows, all of them handler-less**
(`head_essentials_missing` 36 on one site, `image_url_404` 1) — and **0 rows held
back for "handler not registered"**, so no real job is stranded by the second half
of the test today.

The control matters more than the finding: **4 `detected` rows still promote**. A
guard that held back everything would produce the same "37 held back" line and look
identical in the logs, so the number that proves it discriminates is the one that
still gets through. (`a-post-fix-zero-needs-a-demand-control`.)

```sql
-- held back, by reason
SELECT CASE WHEN COALESCE(wi.handler_agent,'')='' THEN 'no handler named'
            ELSE 'handler not registered: '||wi.handler_agent END, wi.item_type, count(*)
FROM site_work_items wi WHERE wi.status='detected'
  AND NOT (COALESCE(wi.handler_agent,'') <> '' AND EXISTS (
    SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent AND ad.deleted_at IS NULL))
GROUP BY 1,2;
-- the control: still promotable  -> 4
```

Full build (`go build ./...`) and vet clean against `git archive HEAD` + my files;
the one `vet` complaint (`load_component_library_actions.go:207 unreachable code`)
is another lane's and pre-exists this change.
