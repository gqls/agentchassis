# HANDOFF — `bugs_open/315`, continue here (2026-08-20 ~17:00Z)

**Read this first, then `NOTES_deployed_at_without_publication.md` from the bottom up.**
The lane's core objective is **DELIVERED, LIVE and PROVEN AT THE ARTEFACT.** What remains is one
optional follow-on and some tidying. Nothing is broken and nothing is urgent.

---

## 1. The bug, in one paragraph

`pages.deployed_at` was read estate-wide as "this page has shipped" and could not mean that. Five
agents wrote it; **two wrote it before the deploy was even dispatched** and the other three wrote it
after a commit **whose result they discarded**. Underneath, delivery is asynchronous and batched
("commit is deploy" → a runner `b2 sync`s a whole changed domain directory minutes later), so no
synchronous stamp could honestly assert publication. And an unchanged file commits as an **empty
commit** with the adapter still reporting success, so neither "the step succeeded" nor "a commit
exists" implies bytes moved.

**The deep finding, which reframed the whole bug:** it is not that pages fail to publish. It is that
the platform could not tell *"this page never needed republishing"* from *"this page failed to
republish"* — identical in every signal it produced.

## 2. What is DONE and LIVE

| piece | state |
|---|---|
| **Migration 491** — drop the pre-deploy stamp from `page-build-handler` + `tool-recreation-handler` | **APPLIED 2026-08-19 15:20Z**, runtime-verified (31 pages, none stranded). "Stamps before any deploy" went 2-of-5 → **0** |
| **git-adapter half** (`0c5b94725`) — `CommitOutcome`, `commit_sha`, `files_sha256` | **LIVE.** Measured on a real orchestration: the reply now carries all three |
| **chassis half** (`086f9b7b7` + `f0dd97c71` + `460ff6b3d`) — `deploy_result_field` guard, `pages.content_hash` write, refuse-on-skip | **LIVE** in `v1.0.1320` |
| **Migration 494** — arms the guard on the 3 post-commit stampers | **APPLIED 2026-08-20 14:27:33Z.** 3 agents armed; survived the roll (config is DB-side) |
| **Council** | **APPROVED**, round 3 of trail `377167cd-6324-4bc7-a866-87ad8c435132` |
| **Register** | `DGH-013` + index row; `DGH-001` corrected (it falsely claimed commit SHAs were already recorded) |
| **RFC_038** | OPEN but its one blocking question (the 19-consumer survey) is **answered in §7** |

### The proof — do not take this on trust, it is re-runnable

`[MEASURED 2026-08-20 ~17:00Z]` `pages.content_hash` = **38 populated** (was **0 of 802**, all estate
history), spanning 14:38→16:57 across four domains. And the artefact check:

```
robot-hands.com/product-detail.html
stored: e9d7090facaaddd3733d11885982979b9710d855df97297c062099bb5b09940b
served: e9d7090facaaddd3733d11885982979b9710d855df97297c062099bb5b09940b    *** MATCH ***
```

So *"is this page serving what we sent?"* is now **one comparison**. On 2026-08-19 the same question
took four steps and a judgement call.

## 3. What is LEFT

**a. The divergence sweep — the only substantive item, and now buildable for the first time.**
`PLAN_2026-08-19…md` decision **D5** has the design. It compares `pages.content_hash` against the
sha256 of served bytes and files a work item on persistent mismatch. It was un-buildable until today
because the thing it compares against did not exist.

⚠ **Read D5's HARD CONSTRAINT before designing anything.** The bug file's own candidate 4 proposes
comparing `deployed_at` to origin `last-modified`. **That does not work and it is measured, not
argued:** run across 40 live pages it returned **40 of 40 "stale"**, all healthy, persisting **85
minutes** — past any usable settle window. A byte-identical rerender legitimately rewrites nothing.
**Only the hash separates the cases.** Predicate on `content_hash IS NOT NULL` and `status='active'`
(retracted/archived pages keep `deployed_at` by design). Prior art for the check shape:
`discovery_checks/check_componentless_pages.go`; for served-bytes acceptance:
`publish_site_action.go`.

**b. `RFC_038` can be closed or ratified** — its survey is done and the change shipped and is proven.
Someone other than this lane should make that call.

**c. `bugs_open/336`'s durable guard — deliberately NOT claimed by this lane.** A test that every key
an action READS is declared in ITS OWN spec, and that no spec declares a key its action never reads.
It would have prevented the outage in §5 outright. It belongs with whoever owns the spec/validation
seam; this lane should not grade its own homework there. If built, run it over
`RegisterActionInputSpec`'s registry, not per-file — the whole failure mode is two sibling specs in
one file each looking correct alone.

**d. `bugs_open/315` itself can probably be closed** once someone independently re-runs the §2 proof.
Its candidates 1 and 2 are delivered; 4 is designed-and-buildable; **3 is undiagnosable from here**
(the runner workflow lives in the private `gqls/sites` repo, so "why did one page fall out of the
batch" cannot be answered — the sweep is designed to *detect* it from this side instead).

**e. ANSWERED 2026-08-21 — a CONTRIB from the `staged_component_build` lane about `commit_sha`.**
`CONTRIB_2026-08-20_from_staged_component_build_commit_sha_resolves_by_guess.md` (in this directory)
asked which path is correct for `commit_sha` in `build-dispatch-loop`'s `complete_work_item`, because
RFC_029 Phase 2 will stop the whole-tree search resolving it. **Answered in full at the bottom of that
file**; the substance, because it is a fact about MY field that nobody else could have:

> **There is no single correct path.** `commit_sha` lands inside whatever the handler's `git_commit`
> step named its `output_field`, and the 19 live steps use **nine distinct names**. Sampling 8 real
> completed items already shows **two** paths (`response.deploy_result.…` ×5,
> `response.css_deployed.…` ×3). A single explicit mapping works for one handler family and silently
> resolves nothing for eight others.

Recommended fix (theirs to make): surface `commit_sha` in each deploying handler's
`complete_workflow.output_fields` so it lands at a stable `handler_result.response.commit_sha`.
**⚠ And: absence is CORRECT for 86 of 397 completions** — those handlers never deploy anything, so a
post-flip check treating a missing `commit_sha` as a regression convicts ~22% of healthy items.
**If they choose the resolver route instead, this lane owes them `collectUniqueValue` extracted into a
shared helper** — that is the one piece of follow-on work this lane has explicitly accepted.

## 4. Commands you will need

Full set in `RUNBOOK_deployed_at_without_publication.md` (two parts). The ones that matter most:

```sql
-- HEALTH. Read the DAMAGE query FIRST, always. This is not style; see §5.
SELECT count(*) FROM orchestration_states WHERE error ILIKE '%deploy_result_field%';  -- must stay 0
SELECT count(*) FROM agent_error_log WHERE error_code='DEPLOY_EVIDENCE_UNREADABLE';   -- 0 today
SELECT count(*), count(content_hash) FROM pages;                                      -- the benefit
```
```bash
# ARTEFACT CHECK — the only real proof. Always cache-bust.
curl -s "https://<domain><url>?cb=$RANDOM$RANDOM" | sha256sum   # compare to pages.content_hash
```
```sql
-- ROLLBACK (if the guard ever misbehaves): un-arms only, restores today's behaviour byte for byte
-- docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD_ROLLBACK.sql
```

## 5. TRAPS — every one of these was hit for real in this lane

1. **`deployed_at` vs origin `last-modified` is NOT a publish check.** The healthy fleet fails it
   40/40 for 85 minutes. Already a `LANDMINES.md` entry.
2. **After arming anything, the FIRST query is "what did I break?", not "did it work?"** I armed 494
   with the key declared on the **wrong sibling spec** (`RenderComponentInputSpec` instead of
   `UpdatePageStatusInputSpec`, 40 lines apart in one file) and **broke every page-publish in the
   estate for 33 minutes** — 8 items failed, 123 rerenders queued. `bugs_open/336`. I had verified
   the *config* and never asked whether a page could still publish. The zero I was watching read as
   "no traffic yet" and meant "nothing can run".
3. **A binary grep for a config-key literal is worthless here** — the string is in the chassis three
   times (the reader + two `zap` calls), so it reads PRESENT even when declared on the wrong spec.
   Check the LIST inside the RIGHT struct:
   `git show <stamp>:…/v3_site_actions.go | awk '/^var UpdatePageStatusInputSpec/,/^}/' | grep deploy_result_field`
   (expect **2** — one entry, one comment. Read the lines, don't trust the count.)
4. **`agent_definitions.updated_at` is NOT "when I armed it"** — it is whenever anyone last wrote the
   row, and it moved twice in one afternoon. Any "since X" query keyed on it under-reports. Pin from
   your own apply, or from the first hash.
5. **Don't date a change from when you looked at something related.** I used my build-check time as
   the arming baseline, was four hours out, and nearly filed a bug against working code.
6. **`_HOLD` migrations: pipe to psql by hand, and do NOT try `--record-only`** — the runner refuses
   sidecars. Record the apply in NOTES.
7. **Scope migration applies** (`MIGRATIONS_DIR=<scratch dir>` on the SAME LINE) — there were 129
   pending files belonging to other lanes.
8. **Commit the DEFINITION before the CALL SITE.** A function with no caller compiles; a caller with
   no definition breaks HEAD for every session. My "atomic hold" is what let HEAD break for 1.8 min.
9. **RFC_029 Phase 1 still RESOLVES conflicts**, it does not refuse them — the ruling in
   `findFieldRecursive`'s comment describes the intention, the paragraph below it describes the
   behaviour. Don't borrow the guarantee; `collectUniqueValue` implements it locally. **Delete
   `collectUniqueValue` when Phase 2 ships.**

## 6. Where the documents are

```
docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/
  PLAN_2026-08-19_…md       design + decisions D1–D5 (D5 = the sweep, with its hard constraint)
  NOTES_…md                 the technical log. Read from the BOTTOM. Missteps are the point
  RUNBOOK_…md               every command, with its gotcha attached (two parts)
  README_where_we_are.md    the owner's plain-prose log
  SUMMARY_2026-08-19_…md    what we believed at the first milestone
  SUMMARY_2026-08-20_…md    the second — read both; the distance between them is the record
  HANDOFF_2026-08-20_…md    this file
```
Elsewhere: `bugs_open/315` (contributions + 3 corrections to its own sizing) · `bugs_open/336` (the
outage, with this lane's acceptance) · `architecture_review/RFC_038` · register `DGH-013` and the
correction to `DGH-001` · `LANDMINES.md` (one new entry, two existing ones strengthened) ·
`WRONG_CALLS.md` (**six** entries from this lane — the most useful thing it produced).

## 7. The one thing worth reading even if you skip the rest

**This bug's exact defect was committed by the lane fixing it, at the moment of fixing it.** I
verified the config was armed and never asked whether a page could still publish — with my own notes
from ninety minutes earlier reading *"config being right is not the artefact — that is this bug's
entire lesson, and it applies to the fix as much as to the defect."*

Not from ignorance of the rule. From **looking for a change's benefit rather than its damage.** Three
other times this session I hit traps that were already documented in `LANDMINES.md`. The pattern in
all of them is the same: every check I ran was sound, and none was the one that mattered — always one
question sideways from the one I asked.
