# 266 — an `archived` page is rebuilt and re-stamped `deployed` by at least four independent producers, none of which reads `pages.status`

**Filed 2026-08-12** by the `bugs_open/215` quiet-mode lane, closing that file's §"Two
defects found while doing this" item 1 and the 090 loose end in
`brochure_component_library/HANDOFF_2026-08-12_215_quiet_mode_continue_here.md` §7.

**Status: OPEN. The damage is LIVE and RECURRING — the most recent re-deploy was
today, 2026-08-12 14:25Z.** Nothing has been changed in the tree by this filing.

## The symptom, measured

Two pages on `fundamentallyai.com` were hand-archived on 2026-08-08 by the
fundamentallyai sweep front (`deployed_at IS NULL`, zero components). They are still
`status='archived'` and they are serving:

```
 name                       | status   | build_status | deployed_at
 tool-llm-cost-calculator   | archived | deployed     | 2026-08-11 19:05:36.493684+00
 ai-readiness-checker-guide | archived | deployed     | 2026-08-12 14:25:21.495362+00
```

Artefact check, 2026-08-12 (a fabricated URL is the control, so the check can come out
negative):

```
200  25861b  https://fundamentallyai.com/blog/ai-readiness-checker-guide.html
200  35331b  https://fundamentallyai.com/tools/llm-cost-calculator/index.html
404   2697b  https://fundamentallyai.com/blog/definitely-not-a-real-page-control.html
```

Note the `deployed_at` values: **they are not the 08-11 10:34 / 11:13 stamps the 215
file recorded.** Both pages have been re-deployed *again* since, the latest four hours
before this filing. This is not a historical incident with a residue; it is a loop that
is still running.

## Root cause — it is not one missing predicate, it is four producers and no shared gate

The 090 diagnosis (`38099787-c7f9-46d4-b75e-3a1867fcaf41`, 5 bundle iterations,
2026-08-11 13:03–13:25) reached this, and **I re-verified every claim of it first-hand
against `site_work_items` before recording it here**:

| work item | producer (`created_by`) | `reason` | completed | matching `deployed_at` |
|---|---|---|---|---|
| `2840024e` `needs_page` | `reconcile_site_plan` | `not_built` | 08-11 10:34:24.51 | 10:34:21.99 ✓ |
| `051d46eb` `owned_page_review` | `reconcile_site_plan` | `not_built` | **never — `needs_human_review`** | — |
| `ac21f8d6` `needs_page` | **`image-build-handler`** | `image_landed` | 08-11 11:13:39.01 | 11:13:25.64 ✓ |
| `75981275` `page_rerender` | `completeness-discovery-agent` → `page-rerender` | `cta_links_stale` | 08-11 19:05:44.05 | 19:05:36.49 ✓ |
| `851f7114` `section_edit` | `claude-ideauk-copy-20260812` → `section-editor` | — | 08-12 14:25:31.77 | 14:25:21.50 ✓ |

**The load-bearing observation is the second row.** For `tool-llm-cost-calculator`,
`ReconcileSitePlanAction` did *not* emit a build — it correctly routed the page to
`owned_page_review` / `needs_human_review`, where it still sits, uncompleted. The gate
worked. `image-build-handler` then rebuilt and deployed that same page **sixteen minutes
later** through a completely unrelated path (`needs_imagery` → image lands → `needs_page`
with reason `image_landed`).

So the mechanism the 215 file recorded — "plan still names the page → reconcile emits
`needs_page` → build → deploy", PLAN-017's documented regeneration trap — is **real but
accounts for only one of the two pages**, and for neither of the two subsequent
re-deploys. A fix in `reconcile_site_plan_action.go` would have closed the one door that
was already shut and left the three that were actually used standing open.

> **This corrects the 215 file's own correction.** That file talked itself *down* from
> "distinct defect" to "the documented regeneration trap, not a new one". The
> regeneration trap explains page one; three further producers explain the rest. The
> narrowing was wrong in the safe-looking direction.

### What no reader checks

- `loadRealisedPages` (`platform/orchestration/actions/reconcile_site_plan_action.go:459-464`)
  — `SELECT ... FROM pages WHERE site_id = $1`, no status predicate. [VERIFIED, read]
- `UpdatePageStatusAction` (`platform/orchestration/actions/v3_site_actions.go:648`), deploy
  branch at `:866-874` — `UPDATE pages SET build_status=$2, deployed_at=NOW(), ... WHERE id = $1`.
  No status predicate. It is the **only** writer of `pages.deployed_at` in the estate
  (`grep -rn "deployed_at = NOW()" --include=*.go platform/ internal/` returns this line and
  two `sites` writers). [VERIFIED]
- That function already carries three refusals — `pageHasComponents`, `pageSectionShortfall`,
  and the assembly-skip refusal added by `bugs_open/210` (fixed, live `v1.0.1268`). **None
  reads `pages.status`.** So the idiom and the place for a refusal both already exist.

## Fix candidates, ordered by what they make unrepresentable

1. **Refuse at the commit seam (`deploy_page` / `git_commit`,
   `platform/orchestration/actions/git_deployer_actions.go`).** For an archived page there
   is **no legitimate deploy path at all**, so the correct seam is the one the owned-page
   guard deliberately avoids. Closes all four observed producers and any fifth.
2. **Refuse at `assemble_page`, copying `owned_page_guard`'s placement — DO NOT DO THIS
   BY REFLEX; it closes only two of the four doors.** `owned_page_guard.go:29-36` states
   why it chose `assemble_page`: `git_commit` "is also how owned pages LEGITIMATELY
   deploy", because `page-rerender` (`rerender_single_page`) and `section-editor`
   (`apply_section_edit`) commit pages without passing through `assemble_page`. Those two
   are precisely the producers behind the 19:05 and 14:25 re-deploys above. **`archived`
   is not the same shape as `owned`** — owned means "not the generic pipeline's to
   rebuild", archived means "nothing may deploy this", and that difference moves the seam.
   [Placement rationale INHERITED from that file's doc comment, measured 2026-08-06, not
   re-measured by me — re-measure which loops call `assemble_page` before building.]
3. A predicate in each producer — N doors, and the fifth producer arrives unguarded.
4. "Operators must retract after archiving" — this is the current de facto state and it is
   a defect, not a remedy. See the `098` note below.

## Relation to `bugs_closed/098` — this defeats a closed bug's remedy

`098` ("archiving a page does not retract it from the deployed site") was closed
2026-08-06 with population zero. Its resolution was deliberate: archiving does **not**
auto-retract; a two-step `page-retraction` procedure is the mechanism. That is not what
is failing here.

What this bug shows is that **the retraction primitive is not durable against a
rebuild.** 098's acceptance test was "0 new `page_rerender` rows for any retracted page
since dispatch", measured over pages that happened to have no active producers. These two
pages have four. Retract them today and the next `section_edit`, image landing, discovery
sweep or reconcile pass puts them back. **098 should not be reopened on this file's
evidence** — its own population is still zero — but anyone relying on retraction being
durable should read this first.

## How to verify a fix

1. Pick an archived page with an open producer; do **not** use a page with no work items,
   which is the shape 098's acceptance accidentally measured.
2. Induce each of the four paths and assert the page stays `status='archived'`,
   `deployed_at` unchanged, and the URL 404s — **with a live control page that must still
   deploy in the same run**, or a guard that refuses everything passes.
3. Standing detector for the class, no fix required to run it:

```sql
SELECT s.domain, p.name, p.status, p.build_status, p.deployed_at
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status = 'archived' AND p.deployed_at IS NOT NULL
ORDER BY p.deployed_at DESC;
```

**Population is NOT yet measured fleet-wide.** The 215 file records open work items
sitting on archived pages across **8 domains**, so this is very unlikely to be a
fundamentallyai quirk, but that figure counts work items, not re-deployed archived pages,
and it is not the same measurement. Run the query above before quoting a scope.

## Provenance

- 090 diagnosis run `38099787-c7f9-46d4-b75e-3a1867fcaf41`; artefacts in
  `diagnosis_artifacts` (`kind='bundle'`, iterations 1–5), **expire 2026-09-10, unpinned**.
  Its findings are reproduced above so this file does not depend on them surviving.
- Every table row quoted here was re-queried first-hand on 2026-08-12, in the hour before commit `f71019552`.

---

## Population MEASURED, 2026-08-12 (shortly before the 19:46Z fix commit) — and the detector this file shipped is BLIND

> **CORRECTION to my own "How to verify a fix" §3 above.** The query I filed
> (`status='archived' AND deployed_at IS NOT NULL`) returns **18 rows across 5 domains**.
> **Only 5 of those 18 are actually serving.** `deployed_at` is a *historical build stamp*
> and retraction does not clear it — which is `016b`'s own standing rule, *build columns are
> history, not liveness*, written by `bugs_closed/098`, the very bug I cross-referenced.
> Thirteen of the eighteen are 404: mostly 098's ten retracted leopardess pages, still
> carrying stamps from April and May. **A fix measured against my query would have looked
> 3.6× worse than reality, and a "population reduced from 18 to 13" claim could be earned
> without changing anything.**

**The detector must be two-step: the SQL selects candidates, the HTTP decides.**

```sql
-- step 1: candidates only. This number is NOT the population.
SELECT s.domain, p.url, p.deployed_at::date
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status = 'archived' AND p.deployed_at IS NOT NULL
ORDER BY s.domain, p.url;
```
Then curl each one, **with a fabricated URL per domain as a control** (all five returned 404,
so the check could come out negative). **A `000` is not a `404`** — one page returned `000`
(connection failure) on the first pass and `200` on three straight retries; recording that
row as "not serving" would have undercounted the live population by 20%.

**The live population, verified at the artefact:**

| domain | page | stamp |
|---|---|---|
| fundamentallyai.com | `/blog/ai-readiness-checker-guide.html` | 2026-08-12 |
| fundamentallyai.com | `/tools/llm-cost-calculator/index.html` | 2026-08-11 |
| leopardessconsulting.co.uk | `/our-approach.html` | 2026-07-17 |
| robot-hands.com | `/gripper-catalog.html` | 2026-08-11 |
| robot-hands.com | `/news.html` | 2026-08-11 |

**5 pages, 3 domains — so it is NOT a fundamentallyai quirk**, which is what the filing above
suspected but could not assert. Three of the five carry stamps from the last two days, so this
is an active process, not a backlog. `leopardessconsulting.co.uk/our-approach.html` is the
useful outlier: stamped 2026-07-17 and still serving while archived, which means the condition
survives for weeks unnoticed and is not specific to the recent replan work.

**Two consumers should be told** (RFC-style, per the 2026-07-29 ruling that a shared
mechanism's other consumers must be told, not merely measured): the **leopardess** lane and
the **robot-hands** lane each own one or two of these pages and neither knows the page is
archived-and-serving. Their `/bugs_open/` reading will not surface it — the row looks retired.

**What the 18 vs 5 gap does NOT mean.** The 13 non-serving rows are not evidence of a second
defect; they are the expected residue of 098's deliberate design (archiving does not
auto-retract, retraction removes the file and leaves the stamp). Do not "fix" them.

---

## FIX BUILT AND COMMITTED 2026-08-12 19:46Z — `580af7ff0`. **Still OPEN: inert until the chassis rolls, and the 5 live pages are untouched.**

**What shipped** (register entry **PBP-042**, same commit per the 2026-07-28 ruling):
two refusals and one shared read, Go only — no migration, no seed, no config key,
no opt-in flag.

| seam | file | stops |
|---|---|---|
| `GitCommitAction` (new step 0b) | `git_deployer_actions.go` | the page **serving** |
| `UpdatePageStatusAction` (before the components check) | `v3_site_actions.go` | the row **claiming** a deploy |
| `pageIsArchivedForGuard` / `resolveDeployTargetPage` | `archived_page_guard.go` (new) | — |

Both seams are needed and that is not belt-and-braces: without the second, a refused
commit still writes `build_status='deployed'` + a fresh `deployed_at`, which is the
state every downstream selector reads.

**Fix candidate 1 was taken, and candidate 2's trap was real.** Placing this at
`assemble_page` — the neighbour's seam — would have been **inert, not merely
partial**: measured against live, active, non-snapshot `agent_definitions`, **no
workflow has a step whose action is `assemble_page`**. All four producers reach
`git_commit` (page-rerender and section-editor each have their own
`deploy_page`→`git_commit`; page-build-handler's `deploy_page` is a `call_agent` to
`target_role: page_renderer`, i.e. into page-rerender's).

**Retraction was checked, not reasoned about.** `page-retraction`'s step is
`retract_page_deployment`, described as *"Audit the retraction graph, retire nav rows,
dispatch delete_file"* — a different path from `git_commit`. Had the guard gone on the
file-writing adapter instead, it would have made `098`'s remedy unusable.

**Fails OPEN**, matching `owned_page_guard`'s posture and for its stated reason (failing
closed would halt fleet-wide deployment on one transient DB error). The window is
countable, not silent:

```sql
SELECT error_code, count(*), max(occurred_at) FROM agent_error_log
WHERE error_code LIKE 'ARCHIVED_PAGE_%' GROUP BY 1;   -- _DEPLOY_REFUSED | _GUARD_UNCHECKED
```

**Tests proven in both directions**, not merely green: the refusal tests were re-run
against a clean `git archive HEAD` overlay **with the guard file added but the two
wirings left out**, isolating the wiring as load-bearing. Both fail there for the right
reason — and `TestUpdatePageStatus_ArchivedPageIsNotStampedDeployed`'s failure output is
sqlmock reporting the unexpected `UPDATE pages SET build_status=$2, deployed_at=NOW()`
firing with `deployed` on an archived page, i.e. this bug, caught by its own test.
`TestGitCommit_LivePageStillDeploys` is the control without which a guard that refused
everything would satisfy every other assertion in the file.

**The untouched-twin question** (raised by `pattern-check`, answered not waved through):
`UpdatePageComponentsStatusAction` writes `UPDATE page_components …`
(`v3_site_actions.go:4205`) and never touches `pages` — **not a fifth door**.

**Council: submitted, verdict pending** — corr `2da9d905-25d8-4916-9b76-bc096679c6ab`,
commit carries `Council-Submitted:`. **Still owed: read the verdict and act on a
REVISE/REJECTED**, because the code is already on the shared branch.

### What is still OPEN, and why this file does not close

1. **Inert until the chassis image rolls.** Go changes do nothing until rebuilt and
   deployed. The bar for `bugs_closed/` is fixed AND live.
2. **The 5 archived-and-serving pages are untouched.** Whether each should be
   un-archived or retracted is a content decision, and two of the three domains belong
   to other lanes, which were told in their own handoffs. **The guard stops the state
   RECURRING; it does not undo it** — and note the guard now also prevents an accidental
   *repair-by-rebuild* of those five, which is correct but means retraction is the route.
3. **Verification after the roll**, in this order: pod-probe the literal
   `ARCHIVED_PAGE_GUARD` on both replicas → re-run the two-step population check (SQL
   candidates, curl verdict, fabricated control) → confirm no NEW `deployed_at` on any
   archived page. A zero on the counters needs a demand control, exactly as
   `bugs_open/215` §4 does: no archived page has been dispatched at means nothing fired.

---

## 2026-08-13 — council **APPROVED**, fix is **LIVE on `v1.0.1295`**, and the objections found two real things

**Verdict:** APPROVED round 1, corr `2da9d905-25d8-4916-9b76-bc096679c6ab`, 4 advisory
objections, **none high-severity**. Landed 10 minutes after submission, not the ~30 the
runbook budgets.

**LIVE, artefact-verified 2026-08-13 on chassis `v1.0.1295`** (rolled 13:53Z): literal
`ARCHIVED_PAGE_GUARD` **present on both replicas**, one-letter near-miss `…GUARE`
**absent**, pre-lane control `OWNED_PAGE_GUARD` present; provenance stamp
`69612d692a4a…` on both, and `git merge-base --is-ancestor 580af7ff0` confirms the fix
is in the build. This answers `debug_historian`'s objection, which was that the plan
named no pod-verification step.

**Behaviourally UNEXERCISED, and the bug stays OPEN.** `ARCHIVED_PAGE_%` counters are
`0` — and that zero is want of demand, not evidence of function: **zero work items
targeting any archived page have run since the roll** (`site_work_items` joined to
`pages` on `status='archived'`, `updated_at > 13:53Z`). Neither fundamentallyai page has
been re-deployed since. **Do not read the zero as "the guard works" until a build is
dispatched at an archived page.**

### The objections, answered with queries rather than with my own word

1. **"The sole-writer claim rests on a literal grep"** (`bug_historian`, `guardian`,
   `prior_art_librarian`, all medium). **Fair, and the method was wrong even though the
   claim survived.** Re-audited structurally — every `UPDATE pages` in `platform/` and
   `internal/`, not the string `deployed_at = NOW()`:
   - **`deployed_at`: the claim HOLDS.** No other statement writes it. Every other
     `UPDATE pages` writes `needs_rebuild`, `page_type`, `sections`,
     `suppressed_sections` or `built_from_plan_version`.
   - **`build_status`: the claim would NOT have held**, and I never had to make it.
     `UpsertPageForRole` writes it via `Col("build_status", …)` — **a helper-constructed
     column name that no literal grep for `build_status` in a SQL string can find.** It
     is safe for this bug only because those callers write `'planned'`, never
     `'deployed'`.
   - **Hardened:** `deployed_at` added to `reservedPageColumns` (`page_role_upsert.go`),
     so the sole-writer property is now **enforced** rather than merely true. Reserving
     `build_status` too was tried and **reverted — it broke three live callers' tests**,
     which is exactly why the distinction above matters.
2. **"A git_commit that batches multiple pages bypasses the guard"** (`editquality`,
   `guardian`, medium). **Such a path EXISTS and the guard is blind to it — but it is
   unexercised.** Census of all **19** live `git_commit` steps: the only multi-page one is
   `site-deployer / deploy_to_git`, `files_field=input_data.site_files.wrap_multipage.files`.
   Its producer `multipage-wrapper` names action `wrap_multipage`, which **is not in the
   Go action registry** (the registered name is `assemble_multipage_site`), and
   **neither agent has any orchestration row** — against a control in the same query
   showing `page-rerender` 214, `section-editor` 130, `deployer-agent` 126, all with runs
   today. Recorded as a **named residual, not a closed hole**: if anyone wires
   `wrap_multipage` up, this guard does not see it.
3. **A second residual found by the same census, which no seat named:**
   `deployer-agent / commit_to_git` commits a single `index.html` by filename with **no
   page identity in its payload**, so `resolveDeployTargetPage` returns `ok=false` and the
   guard is silent. It ran 126 times, most recently today. An archived `index` is
   pathological, so this is not a fix candidate — but it is the honest edge of the
   guard's coverage and it is now written down rather than implied.
4. **"You didn't check the existing status-predicate family"** (`reuse_agent`,
   `prior_art_librarian`, medium/low). Checked now. The family is
   `linkablePageStatusPredicate` (`status NOT IN ('deleted','archived')`) and
   `PageWantedLivePredicateFor` (`status = 'active'`) — **SQL fragments for filtering a
   SET inside a WHERE clause.** This guard asks a different question: is THIS one
   already-identified page archived. Reusing `linkablePageStatusPredicate` would also
   annex `'deleted'`, which this fix **deliberately scopes out** and pins with a test.
   The divergence stands; the criticism that the submission never showed the family was
   considered is accepted.
5. **"The fail-open window is unmeasured going forward"** (`bug_historian`, medium).
   Accepted, unclosed. There is a count but no sweep. **Follow-up, unowned:**
   `SELECT count(*), max(occurred_at) FROM agent_error_log WHERE error_code = 'ARCHIVED_PAGE_GUARD_UNCHECKED';`
6. **`architecture` (approve, `point_fix`)** flagged for the record that `git_commit` now
   hosts two independent guards with a third deliberately excluded one function away —
   *"a second guard on the same seam in two bug-fix cycles is a trend worth someone
   tracking before a third one lands"*. Not actioned; recorded here and in PBP-042 as the
   cross-reference that mitigates it.

---

## Note from another lane, 2026-08-14 — the CONSUMER side of the same seam, and it corroborates your central finding

Added by the `bugfix_168_deployed_asset_path` lane (the `claims_unverified` retraction sweep).
**Contributing, not competing:** `who-owns.py` names this file as yours, the producer-side fix is
APPROVED and LIVE on `v1.0.1295`, and nothing here asks for a change to it. This is a second
consumer reaching your population from the other end, recorded because your §"tell the two other
consumers" already names both domains involved.

**Your central claim reproduces independently, on a third page and by a different route.** The
claims audit (`ScanDeployedClaims`) has no page-status filter, so I measured how much of my own
queue sits on non-active pages: **3 of 30 revalidated `claims_unverified` items**, on pages your
lane has already met.

```
 domain                     | page                       | status   | build_status  | verdict
 robot-hands.com            | gripper-catalog            | archived | deployed      | resolved
 leopardessconsulting.co.uk | for-engineering-teams      | archived | deployed      | still_holds
 webdesign.uk               | index-rejected-v1-20260806 | archived | needs_rebuild | still_holds
```

I set out to file "the audit wrongly judges archived pages" and the artefact check **refuted my own
framing**, which is why it is worth your seeing. Fetched with a fabricated-URL control per domain so
the check could come out negative:

```
200 30997b  robot-hands.com/gripper-catalog.html                      ← archived AND SERVING
404  2886b  robot-hands.com/definitely-not-a-real-page-control.html   (control)
404  2711b  leopardessconsulting.co.uk/for-engineering-teams.html     ← genuinely absent
404  2711b  leopardessconsulting.co.uk/…-control.html                 (control, same size)
302   143b  webdesign.uk/index-rejected-v1-20260806.html              ← never deployed
```

**`robot-hands.com/gripper-catalog.html` is `status='archived'` and serving 31KB to the public**, with
the same-domain control returning 404 — so the 200 is real and not a catch-all. That is your bug's
damage, still visible today, from a producer you may not have enumerated: this page's
`deployed_at` is **2026-08-11 13:14:44Z**, inside the window your table covers, on a domain your
table does not.

**What this adds for whoever picks your file up:**

1. **Do NOT let anyone "fix" the audits by filtering on `pages.status`.** It looks like the obvious
   companion fix and it is wrong in the same direction your bug is: an archived page can be live, so
   filtering it out would stop auditing a page that really is asserting unsupported claims to the
   public. The audit is right to look. (`check_unverified_claims.go` already says "NO PAGE-STATUS
   FILTER, and that is deliberate" for a parity reason; this is a second, better reason.)
2. **The discriminator nobody reads is "is it served", not "is it archived".** My three closing gates
   read `pages.build_status` and `pages.deployed_at` and neither of those separated the serving page
   from the two dead ones here: all three were `deployed`/`needs_rebuild` in exactly the pattern you
   would expect of live pages. `status` did not discriminate either — it was `archived` for all three,
   including the one that serves.
3. **Residual cost on my side, stated so it is not double-counted as a new defect:** two of those
   items are parked in `needs_human_review` for ever, asking a human to correct copy on pages that
   return 404 and 302. So my gate-reachable population is **16, not 18**. That is a consequence of
   your seam, not a separate bug, and I am not filing one.

Cross-references: `NOTES_deployed_asset_path.md` (2026-08-14, with the corrected framing recorded as
a correction), `RUNBOOK_deployed_asset_path.md` § "Is a page the audit flagged actually SERVED?" for
the control-paired curl recipe, concept register CQ-021.

---

# 2026-08-15 — THE GUARD IS BEHAVIOURALLY PROVEN. 20 refusals, 3 pages, 2 producers, BOTH seams.

**This discharges the "Behaviourally UNEXERCISED" block above**, which said: *"Do not read the
zero as 'the guard works' until a build is dispatched at an archived page."* One was. Several
were. Found by the `bugs_open/215` O2 lane while re-checking counters after the `v1.0.1300` roll
— **not by contriving a test**, which is what §10 of the 215 handoff predicted would happen.

## What fired

`SELECT … FROM agent_error_log WHERE error_code='ARCHIVED_PAGE_DEPLOY_REFUSED'` —
**20 rows, 2026-08-14 18:34:39Z → 19:53:12Z**, all robot-hands.com.

| refused page | id | status | serving? | rows |
|---|---|---|---|---|
| `gripper-catalog` | `64fab29e…` | archived | **200** (one of the five archived-and-serving) | 2 |
| `news` | `18d681af…` | archived | **200** (one of the five) | 9 |
| `learning-center-index` | `5a3b27db…` | archived since 08-03 | **404** (already retracted) | 9 |

**Both seams refused, which is the part that matters** — the fix built two, and both were needed:

- `step_name=deploy_page`, `action=git_commit` → *"committing it would re-publish a retired page"*
- `step_name=update_status`, `action=update_page_status` → *"refused deploy stamp"*

**Two independent producers**, which answers the "does it close all four doors" question from the
other side: `page-rerender` **and** `page-build-handler`. Neither is `assemble_page`, so this is
also live confirmation that copying `owned_page_guard`'s placement would have closed the wrong
doors (trap 1 in this file).

**Corroborating negative:** none of the 20 names the `215` lane's pair-5 loser
(`48d52965…`, archived 16:36Z the same day) — archiving removed it from the rerender wave's
population, so it attracted no demand. The proof came from the pre-existing archived-and-serving
population, not from anything this lane staged.

## ⚠ A SECOND DEFECT, visible only because the guard now refuses — the failure MISNAMES ITS CAUSE

The three refused pages were driven by three work items. Two of them, both `literal_markdown`
filed by the 15:17Z improvement sweep, ran to **`failed`, `attempt_count=3/3`**, and each records
this as its reason:

> `completion blocked: post-fix verification found the defect still present: 6 finding(s) still
> present across N component(s)`

**That sentence blames the fixer. The fixer worked.** [INFERRED — the co-occurrence is
[MEASURED]: same `page_id`, same minute, guard refusal at `update_page_status` immediately before]
the repair was written, the guard refused the commit and the stamp, so the deployed artefact never
changed, so the post-fix verification re-read an unrepaired page and reported the defect present.
A reader of that work item concludes the markdown fixer is broken. **The check that settles it:**
confirm whether the verifier reads the served artefact or `content_data`, and whether the refusal
precedes the verification within each orchestration —
`SELECT current_step, status FROM orchestration_states WHERE …` for the two item ids
(`858bb33e-20ec-46f8-99b1-949d59602a7f`, `e7f98894-0fe6-4ab7-a443-c4c572e5712c`).

Two costs, both real and neither yet owned:

1. **Wasted LLM work.** Three full repair attempts per item against a page that cannot deploy.
   The guard is doing its job at the last possible moment; nothing upstream declines to *start*.
2. **A misleading failure.** This is the estate's own *"a REJECTION does not name the RULE that
   rejected"* shape. The honest message is "refused: page is archived", which the guard already
   knows and writes to `agent_error_log` — it just does not reach the work item.

**Cheapest fix candidate, stated not taken:** have the checkers skip archived pages when filing
(`literal_markdown` was filed against three archived pages), or have the repair pipeline
pre-check `pages.status` and close the item as `wont_fix` with the real reason. The first is
narrower; the second catches every item type. **Neither is this lane's to choose.**

## Observability note for anyone querying these rows

`context->>'page_id'` is present on **all 20** rows and is the reliable key.
`context->>'page_name'` is populated on **1 of 20** (`gripper-catalog`, the `git_commit` arm) and
**empty on the rest** — the `error_message` for those reads *"page  is status=archived"* with a
blank where the name should be. **Do not filter or group these rows by `page_name`.** Also note
`domain` is empty on every `page-rerender` row and populated on every `page-build-handler` row —
the `COALESCE(domain,'') = ''` rule for this table applies exactly as its landmine states.

## Why this file is NOT being moved to `bugs_closed/` today

The CLAUDE.md bar (fixed AND live, restored by the owner 2026-08-12) is met, and this file's own
stricter self-imposed bar — behavioural proof — is now met too. **It stays open on the two named
residuals recorded above**, because both are paths by which the original defect is still
reachable: the **multi-page `git_commit`** path that bypasses the guard, and **`deployer-agent`'s
`index.html` commit**, which carries no page identity for the guard to check. Closing on "the
guarded paths are proven" would retire a ticket whose defect two unguarded doors still admit.
**Stated so the next session knows this was a decision, not an oversight.**

## Contribution 2026-08-16 (leopardess services-restore session) — writes INTO an archived page's components, not just rebuild/redeploy

`leopardessconsulting.co.uk/for-engineering-teams.html` (archived, 404 live, `build_status`
still 'deployed', `deployed_at` 2026-07-17 — already in the table above): its
`generic-text-block.content_data` was rewritten with fresh LLM prose at **2026-08-15
10:46:04Z** ("Most AI projects stall at the same point…"), and a new
`needs_internal_links:for-engineering-teams` item was filed **2026-08-16 10:01:59Z**
(`triaged`). Neither was this session. Producer not traced (out of lane) — recorded because
it is a live example of the write side of this defect on a page nobody can reach, and it
also blocks the `claims_unverified` revalidator (`bugs_open/262`'s deployed_at gate can never
be satisfied by a page that never redeploys), which is why that item had to be closed by
hand. The 90,790 figure it carried has been cleared from `features` so a revival cannot ship
it.
