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

> **CORRECTED 2026-08-16, same day — "written by exactly one thing" is FALSE for
> `original_pipeline`.** The `090` run named three writers (see the UNVERIFIABLE entry
> below); the other two hardcode the literal `"build"`. The conclusion survives, on the
> VALUE rather than the key — every one of these rows reads `design` or `content`,
> which only `to_jsonb(pipeline)` can produce. Caught by the diagnosis loop, not by me.

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

## 2026-08-16 — council round 1: REVISE, and two of the objections were RIGHT about my evidence

Verdict `revise`, gated by **guardian** (high), with objections also from
`editquality`, `reuse_agent` and `prior_art_librarian`; 4 seats abstained. None of
it disputed the mechanism. All of it disputed what I had *shown*.

**The two that were straightforwardly right:**

- **`prior_art_librarian`: `spec->>'original_pipeline'` cannot name which CHECK
  wrote a row.** It is the row's own pipeline — one label over several producers,
  the `audit_source` landmine shape — so my 15-design/3-content split does not
  attribute anything to `check_palette_contrast` or `check_content_duplication`. It
  is right. Re-measured on the marker each check writes into its OWN spec:

  ```sql
  SELECT spec->>'check', count(*), count(DISTINCT site_id) FROM site_work_items
  WHERE item_type='capability_gap' AND status='blocked' GROUP BY 1;
  -- content_duplication | 9 | 9
  -- palette_contrast    | 9 | 9
  ```

  Exactly the two files I edited, nothing else. The pipeline split is still
  evidence — but only of *which mechanism* wrote the key, never of which producer.

- **`editquality`: "three of six producers got it wrong" was loose.** The accurate
  enumeration is **six** producers filing an empty-handler flag row at `detected`,
  and it gained one I had not found: `check_site_structural_validity.go:1148-1163`
  (`head_essentials_missing`, 36 live rows) — which **also omits the field**, so
  that is TWO omitters, not one, and my value-grep was blind to both.

**The gating one, and why I resubmitted rather than backed out.** `guardian` (high)
said the claim-path refactor rests on an unverified equivalence: I wrote
"semantically identical" and never quoted the query claim actually runs. Fair — and
answerable exactly, which is the difference between an evidence objection and a
scope veto. At `7027a2801^`:

```
SELECT EXISTS(SELECT 1 FROM agent_definitions WHERE type = $1 AND deleted_at IS NULL)
```

No `is_active`, no `is_snapshot` — in the before OR the after. The seat's own
answered check closed the practical half in round 1: **zero** live `handler_agent`
values point at an inactive or snapshot definition.

**`reuse_agent` asked the best question**: why not make the existing
`detected-item-promoter` scheduled task the mechanism, rather than a second
routability predicate? Answer, measured: it promotes ≤20 rows per 900s fleet-wide,
asynchronously, and demands ≥1 lifetime `complete` for the pair — right for an
unattended sweep, wrong inside a loop that promotes its own site's findings and
dispatches them in the same run. Its second question found something better than
an objection: a repo-wide search for the predicate returns a **third** consumer,
`remit.go:230` `HandlerStepConfig`, whose own comment already names
`claim_work_item_action.go` as its source of truth and keeps the coupling **in
prose**. Three consumers of one predicate, two coupled by a comment — which is the
estate's own argument for extracting it. Not migrated here (the guardian objected
to blast radius); recorded as the named next adopter.

**Cost of the round: ~11 minutes, and it bought a corrected attribution, a sixth
producer, and a third consumer I had not enumerated.** The `a-revise-round-is-
cheaper-than-the-defect-it-finds` lesson, again.

## 2026-08-17 — LIVE on v1.0.1305, and the repair was done by another lane

The guard shipped. Two independent proofs, and the better one is not mine:

- **Another session verified it at the artefact and wrote the repair**
  (`docs/agent_docs/sql_for_agents/442_repair_flag_only_rows_blocked_by_the_old_promoter.sql`,
  uncommitted in the tree at the time of writing — **not mine to commit**). Its
  header records chassis AND core-manager on `v1.0.1305`, image digests matching the
  local images, OCI label **`revision=6a782274b`**, and
  `git merge-base --is-ancestor 7027a2801 6a782274b` exiting 0. That is the
  definitive form — a tag is not evidence, an ancestry proof is.
- **My own three-needle binary probe**, both replicas: the promoter's log literal,
  `workItemHandlerRegisteredSQL`'s `EXISTS (SELECT 1 FROM agent_definitions ad WHERE
  ad.type = `, `workItemRoutableSQL`'s `, '') <> '' AND `, and
  `countUnroutableDetected`'s `SELECT wi.item_type, count(*)` — all PRESENT, with a
  long-lived positive control PRESENT and a nonsense negative control absent in the
  same exec.

**Repair state, measured 2026-08-17 11:0x — all handler-less blocked rows are gone:**

```sql
SELECT count(*) FROM site_work_items WHERE status='blocked' AND handler_agent='';  -- 0
```
`capability_gap` 18 → `deferred`, `image_url_404` 40 → `detected`,
`needs_experience_plan` → `deferred`, `page_rerender` → **cancelled** (a ten-day-old
verification rerender for another lane) — one transaction, 11:01:26. The two
hand-inserted rows were judged individually exactly as this lane's RUNBOOK asked,
including the owner-raised one being deferred rather than cancelled. My own
`REPAIR_*.sql` did **not** run (`repair_284_backup` does not exist) and is now
redundant; it stays in the lane as the rollback-bearing version if a future repair
of this class is needed.

### ⚠ THE GUARD IS LIVE AND UNEXERCISED, and a zero here is not proof it works

> **CORRECTED 2026-08-17, ~1h after writing it — THIS SECTION'S CENTRAL CLAIM IS FALSE.**
> The guard **has** been exercised, and the section below says it "cannot currently run
> at all". What caught it: the lane that closed this bug ran a **single-step
> `triage_detected_items` dispatch** on `leopardessconsulting.co.uk` (36 flag-only
> `detected` rows and nothing routable — a manufactured demand control), correlation
> `a5be3dea-3f2c-490a-9922-22993662bc95`, result **`promoted: 0`,
> `not_promotable: 36`, `not_promotable_by_type: {"head_essentials_missing": 36}`**,
> rows left `detected`. That is the guard working, in production, on the real binary.
> **My reasoning error:** I read "the only SCHEDULED driver is disabled" as "the code
> path cannot run", forgetting that a step can be dispatched directly — which is
> routine here and is in my own memory index as *"single-page deploy bypasses stalled
> queue"*. A disabled cron bounds what runs BY ITSELF; it says nothing about what can
> be run ON PURPOSE. Everything below about `improvement-sweep` being `enabled=false`
> and about the `detected-item-promoter` being the live promoter is still true and
> still worth knowing — it is the conclusion drawn from it that was wrong.

```sql
SELECT name, enabled, last_triggered_at FROM scheduled_tasks
WHERE name IN ('improvement-sweep','detected-item-promoter');
-- improvement-sweep      | f | 2026-08-14 16:34
-- detected-item-promoter | t | 2026-08-17 10:46
```

`improvement-loop` is the **only** live carrier of `triage_detected_items`
(measured, `SingleOwner`), and its only driver `improvement-sweep` is **disabled**.
So the code path my guard sits in **cannot currently run at all** — the same
"LIVE but NEVER YET EXECUTED" shape as WII-017/WII-018.

Worse for anyone reading a dashboard: the thing actually promoting rows today is the
`detected-item-promoter` scheduled task, created **2026-08-15**, which already
refused handler-less rows from birth. **So "0 blocked rows" today is explained by
the disabled driver plus a promoter that was never broken, NOT by my fix.** What the
fix buys is that re-enabling `improvement-sweep` — or any future agent gaining the
step — cannot re-create the class. Say it that way and nothing is overclaimed.

### Misstep 1 — a `kubectl exec` loop that times out reports "absent"

My first needle probe ran one `kubectl exec` per needle in a shell loop with a 120s
tool timeout. It was killed during the third needle, and that needle printed
**`absent`** — because the killed `exec` exits non-zero and my `if` read that as
"not found". The loop's output looked complete: three lines, one of them a clean
negative. Re-run as a SINGLE exec with all needles and a negative control inside, it
was **PRESENT**. A timeout and a real absence are the same reading at the call site;
put every needle in one exec, include a control that must come out the other way,
and check the exit status.

### Misstep 2 — "zero runs in five weeks" was measuring the PRUNER

Chasing who promoted the 58 rows, I found **0** `orchestration_states` rows for
`improvement-loop`, `design-audit-agent` and `site-review-agent` across a table
whose oldest row is 2026-07-13, and briefly took that as "the promoter never ran, so
the attribution is wrong". It is not:

```sql
SELECT status, count(*), min(created_at)::date, max(created_at)::date
FROM orchestration_states GROUP BY 1;
-- COMPLETED | 3264 | 2026-08-16 | 2026-08-17   ← nothing older than ~24-48h survives
```

**`COMPLETED` rows exist only for the last two days.** The table cannot answer "did X
run on 08-09", so an empty result there is not evidence of anything. The 5-week span
comes entirely from a handful of stuck `RUNNING`/`CANCELLED` stragglers, which is
exactly what makes the window look long enough to trust. Before reading an absence
out of `orchestration_states`, ask what its retention actually is — for terminal
rows, and separately from the oldest row of any status.

### And a new bug came out of the verification: `bugs_open/291`

The other arm of the same claim guard is bleeding while this one is paused:
`tool-auditor` files `item_type='needs_human_review'` at `handler_agent='hitl-review'`,
an agent that has never existed — 14 rows, 2 sites, **5 yesterday → 14 today**, born
dispatchable so neither this lane's guard nor its proposed CHECK constraint covers
them (a CHECK cannot subquery `agent_definitions`). Filed with the producer named
and a `090` as its next step.

## 2026-08-17 — the repair ran, the guard is proven live, and I wrote a duplicate of this lane's own repair file

**The miss first, because it is the useful part.** I (the 279 lane, picking this
up after the v1.0.1305 roll) wrote and applied
`sql_for_agents/442_repair_flag_only_rows_blocked_by_the_old_promoter.sql`
**without opening this directory**, where `REPAIR_2026-08-16_blocked_flag_only_rows.sql`
was already written — to the same council standard, with the same gate. No damage:
the two reach the same end state, 442 is the one in `schema_migrations`, and the
lane file now carries a SUPERSEDED banner so nobody runs it and concludes the
repair never happened. **The check I skipped costs one command:**
`ls docs/agent_docs/docs024_key_docs_latest/bugfix_<n>_*/` before writing any SQL
for a bug that has a lane. `who-owns.py` names the lane; it does not list what the
lane has already built. Logged in `WRONG_CALLS.md`.

**What actually happened, in order:**
1. Roll verified per SERVICE at the artefact — chassis AND core-manager both
   v1.0.1305, running digests matched the local images, OCI `revision=6a782274b`,
   `git merge-base --is-ancestor 7027a2801 6a782274b` → 0.
2. Census re-measured: 60 blocked / 4 item_types / 58 with `original_pipeline` —
   identical to this lane's measurement, so nothing had moved.
3. **A contradiction that had to be walked past**: both `capability_gap` producers
   read `Status: "deferred"` at HEAD, which would mean a `detected`-only promoter
   could not have touched them. `git log -S` settled it — `deferred` arrived in
   `7027a2801`, this lane's OWN fix. The rows were born `detected`. Now a
   LANDMINE, because the next reader hits the same apparent refutation.
4. Repair applied (442). The two hand-inserted rows judged individually; the
   `needs_experience_plan` one is **owner-raised** and went to `deferred` (roadmap
   view), NOT cancelled.
5. CHECK constraint `swi_no_handlerless_promotable` (mig 443), `NOT VALID` +
   `VALIDATE`, induced with two negative controls.
6. **Demand control for the guard** — the part a passing zero cannot give you:
   `leopardessconsulting.co.uk` holds 36 flag-only `detected` rows and nothing
   routable, so a `triage_detected_items` run there could only hold them back or
   promote them. Result: `promoted: 0, not_promotable: 36,
   not_promotable_by_type: {"head_essentials_missing": 36}` (corr `a5be3dea`).
   Under the old binary all 36 would have been promoted and then blocked.

> **CORRECTED 2026-08-17 (same session, minutes later):** the entry above was
> written dated 2026-08-16 — the repair, the constraint and the demand-controlled
> probe all happened on **2026-08-17**. I anchored to the roll (2026-08-16 22:07
> UTC) rather than to the clock; the neighbouring lane's commit timestamp is what
> exposed it. Migrations `442`/`443` keep their 08-16 prose headers because they are
> recorded and must not drift their checksums; their `applied_at` is correct.
>
> **Also superseding the previous entry's headline:** it says the guard is
> "UNEXERCISED because improvement-sweep is disabled". That was true when written
> and is not now — a single-step `triage_detected_items` run against
> `leopardessconsulting.co.uk` (36 flag-only rows, nothing routable) exercised it
> directly: `promoted: 0, not_promotable: 36`. A disabled sweep does not mean the
> guard cannot be tested; it means you have to drive the step yourself.

## 2026-08-17 (later) — the owner's three rulings, and the same-tag deploy that shipped nothing

**Rulings taken: (1a) unify the third rendering · (2b) fix `227` before routing the
parked row · (3) "add the images".**

### The deploy shipped no new code — probed, not assumed

Pods restarted 14:42Z as `agent-chassis-5bd56bdd9b-*`, image **`v1.0.1305` unchanged**,
digest `sha256:f90a7e88…`. Single-exec probe with controls on the running binary:

| needle | expected | got |
|---|---|---|
| `"patch the three call sites the council saw"` (from `5b30a831b`, POST-stamp) | present if new code shipped | **absent** |
| `"No handler_agent set — item cannot be routed to any agent"` (long-lived) | present | present |
| `"held back detected items nothing can route"` (my 284 guard) | present | present |
| `ZZZ_never_present_ZZZ` | absent | absent |

So the binary is still the `6a782274b` build and `git rev-list --count 6a782274b..HEAD`
= **215 commits unshipped**. `IMAGE_TAG` was not bumped, so the rebuild served the node's
cached image — the trap CLAUDE.md's build section names. Another lane measured the same
thing independently (commit `4c77496e9`) and MEMORY now carries it as a banner. **Every
Go change from today, including the approved unification, is committed and inert.**

### (1a) was already done by another lane — I verified and repaired one defect

`10fc61184` implements it correctly: `discovery_checks.HandlerRegisteredSQL` is now the
single definition (that package, not `actions`, because `actions` imports it and never the
reverse — the cycle decides the home), with `actions.workItemHandlerRegisteredSQL`
delegating and `HandlerStepConfig` rendering from it. Builds and passes at HEAD.

**What it got wrong, now fixed (`cab28cfbe`):** the new function was inserted INTO
`HandlerStepConfig`'s doc comment, so the nav-link-fixer/`077` rationale became
`HandlerRegisteredSQL`'s documentation and `HandlerStepConfig` was left with an orphaned
`Returns (config, agentExists, error)` tail. Comment-only, but comments are load-bearing
here and godoc showed the wrong prose on the wrong function.

### (3) "add the images" — measured, and it does not fit most of the findings

Census of the 30 open `unbacked_path` findings, joined per finding against the site's own
active assets by **exact `asset_key`** and separately by `purpose`:

| basename | findings | exact `asset_key` exists | any asset of that purpose |
|---|---|---|---|
| `hero` | 10 | 4 | 10 |
| `case-study-*` | 10 | 5 | 0 |
| `favicon` | 4 | 0 | 0 |
| `og-card` | 4 | 0 | 0 |
| `logo` | 2 | 2 | 2 |

**11 need a DEPLOY, not a generation** (asset present under the exact referenced key —
`deploy_image_asset_action.go:383` documents `asset_key=hero → assets/images/hero.jpg`, so
the page's path is right and the artefact is absent). **6 need a REPOINT**: e.g.
lendzy.co.uk holds 9 active heroes keyed `hero_home`/`hero_about`/`hero_price_cap`… and
none keyed plain `hero`, so nothing deploys to the base path a page still asks for.
**5 genuinely need generating** (`case-study-*`, 2 sites). **8 are favicon/og-card and
belong to `bugs_open/131`** (owned — left alone).

**Live blocker on the 5, and the obvious cause is REFUTED.** `needs_imagery` today: **12
failed / 4 complete**, every failure `step call_asset_deployer failed: … failed to get
latest commit/base tree for branch "master": … 404`. All four sites involved have an empty
`sites.github_repo` — **but so do the sites whose items COMPLETED in the same window**;
the same two domains sit on both sides of the same day, so the config cannot be the
discriminator. Not diagnosed further: it is those lanes' site work.

**Nothing generated, nothing dispatched, no rows touched.** The census and the blocker
went into `bugs_open/114` (its subject is this class from the other end) and the
dependency into `bugs_open/227`; both are owned lanes, so both are contributions, not
competing fixes. `who-owns.py` was run on 114, 131 and 227 before writing a word into any
of them.
