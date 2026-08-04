# NOTES — the unpublish primitive (`bugs_open/098`)

Running record, append-only, newest at the bottom. Missteps are the point.

---

## 2026-08-02 — picking the bug

Surveyed `bugs_open/` against the live `.jsonl` transcripts of ~30 concurrent
sessions (`who-owns.py` reads COMMITS, so it is blind to a session mid-fix —
grepping live transcripts is the check that sees those). 098 came out unowned,
still reproducible, and named a **missing framework primitive** two bugs share.

`bugs_closed/125`'s council round had asked that the same gap be logged as a
required follow-up rather than absorbed; it was logged on 098. So this is that
follow-up, not a fresh choice.

## 2026-08-02 — the finding that shrank the job

098's fix candidate 1 is "make the deploy reconciling rather than additive",
described as the biggest-blast-radius option needing its own survey. It is
**half-built already**: `gqls/sites/.github/workflows/deploy-to-b2.yml` runs
`b2 sync --delete --skip-newer` per changed domain and then purges the Cloudflare
zone. The *sync* reconciles; only the *git tree* is never removed from. That
turned a deploy-pipeline redesign into one missing capability.

Reading the workflow before designing anything is what found this. It was not in
any doc.

## 2026-08-02 — probed rather than assumed

Wrote the existence filter as "defensive", then probed GitHub to see whether it
was actually needed. POST `/git/trees` against the live master tree (a tree
object is created unreferenced — no ref moves, no workflow fires):

```
null sha, path PRESENT -> 201, tree 0d8dab50…
null sha, path ABSENT  -> 422 GitRPC::BadObjectState
```

So it is **required**: without it, re-running a repair fails. Good outcome from
a cheap check — and the opposite of what I had written in the comment.

## 2026-08-02 — MISSTEP: I justified a design choice with a hazard I had not measured

I chose a per-path existence check over one recursive tree listing and wrote the
reason as though it described the estate: *"a recursive listing can come back
`truncated: true`"*. Then I measured: `gqls/sites` recursive is **1,847 entries,
`truncated: false`**. Nowhere near the limit.

The decision was right; the stated reason was fiction, written in the same voice
as the 422 probe two paragraphs above it. Logged in `WRONG_CALLS.md`. The tell
is the word **"can"** — it smuggles a hypothetical into a paragraph of facts.

## 2026-08-02 — MISSTEP: a pod-grep that read 0 for everything, including its control

Verified the built image with `strings /app/git-adapter` — got `0` for every
positive check. The binary is at `/root/git-adapter`; `strings` with no readable
file read stdin and matched nothing. **The control also read 0**, which is what
flagged it as a broken check rather than a bad image. Without the control I would
have concluded the build had failed and rebuilt for nothing.

In-pod grep then failed on permissions (container runs as uid 1000, binary owned
by root). Settled on the stronger check anyway: extract the binary from the image
locally, grep it with a **negative control** (a string the change REMOVED, expect
0), then compare the **pushed digest to the running pod's `imageID`**. That is a
complete provenance chain and it does not depend on in-pod tooling.

## 2026-08-02 — live proof on a scratch path before touching production

`unknown-domain/` exists in the sites repo and does NOT match the deploy
workflow's changed-domain regex (`^[^/]+\.[^/]+/` — no dot), so it is an in-repo
scratch area with no B2 or CDN side effects. Wrote a file through the adapter,
deleted it via the new verb (404 confirmed), re-deleted it:

```
CommitToRepo: no-op — every requested deletion is already absent
```

head unmoved. Write → delete → idempotent re-delete, all three arms live.

## 2026-08-03 — THE BUG'S STATED MECHANISM IS ONLY HALF RIGHT

> **CORRECTION to 098's root cause, found by re-checking the target before
> retracting it** (CLAUDE.md: other threads change things beneath us; a figure
> carried forward unchecked is how a stale premise gets diagnosed as a bug).

098 says an archived page is **frozen**: "the last HTML that page ever rendered
is frozen and it keeps being served". For the retraction target that is **false
as of 2026-07-31**. `robot-hands.com/learning-center/index.html` is re-rendered
and re-committed **twice a day**: 08-01 08:07, 08-01 20:06, 08-02 08:05, 08-02
20:15, 08-03 08:15.

Traced work item → orchestration → commit → SQL. Cause is
`queueNewsPageRerenders` (`render_news_section_html.go:64`):

```sql
WHERE p.site_id = $1
  AND cc.function IN ('latest-news','news-listing')
  AND p.build_status = 'deployed'     -- and NEVER p.status
```

`build_status` records whether the page ever shipped; `status` records whether
the platform still wants it served. Archiving sets the second and leaves the
first, so a selector keyed on the first keeps choosing a retired page for ever.
6 `page_rerender` items raised for it since 07-31.

**This makes a retraction self-undoing** — delete the file and the next news
refresh republishes it. So the retraction could not have been proven without
fixing this first, and a session that had retracted and walked away would have
recorded a fix that quietly reverted overnight.

Blast radius, measured: fleet-wide this selects exactly **one** non-active page,
this one. Only two page statuses exist (`active` 557, `archived` 25). Filed
through the diagnosis loop as a structural claim (`5bdec8cf`).

098 was accurate when filed on 07-26; it was overtaken five days later. The bug
file gets the correction, not a rewrite.

## 2026-08-03 — the council's REVISE was worth having

Round 1 → **revise**, gating objection from `editquality`, echoed by
`bug_historian` and `reuse_agent`: my retraction derives the deploy path via
`PageFilePathFromURL`, and a landmine warns two functions eleven characters apart
decide a page's deploy path and disagree. If mine is not the one that WROTE the
file, the deletion silently targets nothing.

The landmine predates the 125 fix (all five derivations now delegate to the
shared helper, `git_deployer_actions.go:436`), but the seats were right that I had
**asserted** rather than shown it. So I showed it — ran the real helper (not a
re-implementation, which would be the very drift under test) over all 13
candidates and looked for the file in the actual deploy repo:

| | |
|---|---|
| `sites`-repo pages with a file at exactly the derived path | **11 of 12** |
| the 12th (`blog/learning-center-article.html`) | genuinely absent, serves 404 — it is the dead link the frozen page advertises |
| the 13th (relojistas `contacto`) | deploys to `vm-sites`, correctly absent there |

**Zero mismatches.** Refuted with evidence, not argument.

## 2026-08-03 — MISSTEP: I checked the wrong repo for one of the 13

My first sweep queried `gqls/sites` for every candidate. relojistas.com has
`github_repo='vm-sites'`. The row came back with a sha and I nearly recorded it
as "file present but 404" — an interesting-looking contradiction that was
entirely my own error. `gqls/sites` does hold a stale
`relojistas.com/contacto.html`, left over from before that site moved to the VM,
and it serves nothing because the domain is served from the VM.

The production code was never wrong here — `resolveGitRepoNameDB` picks the
per-site repo. **The verification script did not, and a verification that does
not model the thing it verifies is a coin toss.**

## 2026-08-03 — the owner's directive, and what it changed

> "When we retract a page we should look through the whole site to change any
> links inward, nav links, and any places that it linked to but now noone links
> to."

This is 098's second surface reached from the other end, and it is a stronger
statement of it. Implemented in `retract_page_graph.go` with the three
obligations split by KIND (editorial refuses / nav deactivates / stranded
reports) — the argument is in the file header and in DGH-006.

Audited for the real target before building: 0 editorial, 0 chrome, 0 nav,
0 stranded. Clean retraction.

## 2026-08-03 — MISSTEP: my first orphan census was blind, and said so confidently

The first pass at "what would be orphaned" looked only at `page_components` and
reported that **10 of 16** outbound targets had no other referrer — which would
have meant retracting one page stranded most of the site. Implausible on its
face, and wrong: the nav and the site chrome are where most links live, in
`site_nav_items` and `site_components`, and my query could not see either.

Re-run across all three sources: **every one of the 16** has chrome and nav
referrers. Nothing is stranded.

The lesson is the one 016b keeps recording: a link census that enumerates one
table answers a question about that table, not about the site. It is also
exactly why the shipped code copies the orphan check's three sources rather than
inventing its own.

## 2026-08-03 — the tests caught a real runtime bug the compiler could not

`[]string` is not a driver-supported parameter type here (database/sql over pgx
stdlib, with neither `pq.Array` nor pgx array types imported). Every one of my
new queries would have failed at runtime with *"unsupported type []string"*. The
sqlmock tests failed first; the fix is the project's existing
`datahelpers.PGTextArrayLiteral` with an explicit `::text[]` cast.

Separately, `go build` cannot parse SQL, so the queries were **PREPAREd against
the live schema** — which caught `site_components.component_type`, a column that
does not exist (it is `slot_name`). Two real defects, neither visible to the
compiler, both found before shipping.

Then ran them for real, with a **positive control**, because all four guards
returning empty for the target is also what a broken query looks like:

```
target /learning-center/index.html : 0 body, 0 chrome, 0 nav, 0 stranded
control /contact.html              : 4 body, 2 chrome, 1 nav
```

The control is the only thing that makes the zeros mean something.

## 2026-08-03 — the 090 diagnosis run produced NO VERDICT, and that is stated rather than glossed

Filed the resurrection claim through the diagnosis loop as CLAUDE.md requires for a
structural claim whose cause is not where the symptom is
(`5bdec8cf-24cc-419f-8d9d-b3d7a8df6dbb`). The dispatch loop **claimed it, ran, and
COMPLETED at 10:30:47** — and wrote **three `bundle` artifacts and no diagnosis**.
No `verdict`, no confirmation, no refutation:

```sql
SELECT kind, count(*) FROM diagnosis_artifacts
 WHERE correlation_id='5bdec8cf-24cc-419f-8d9d-b3d7a8df6dbb' GROUP BY 1;
--  bundle | 3        (and nothing else)
```

**So the loop did not corroborate anything, and this file does not claim it did.**
The owner ruling of 2026-07-31 allows a filing session to substitute equivalent
first-hand verification provided it says so plainly rather than omitting it. Doing
that here, and naming the chain:

1. `pages.deployed_at` for the page moved to **today** (10:05 read) — the trigger
   for looking at all;
2. the sites repo's commit list for that exact path shows a **single-file modify**
   at 08-01 08:07, 08-01 20:06, 08-02 08:05, 08-02 20:15, 08-03 08:15;
3. `orchestration_states.collected_data->'input_data'` for the most recent one
   carries `source: render_news_section`, `reason: section_data_resolved` and the
   page's own id;
4. `site_work_items` holds **6** `page_rerender` rows for that page with
   `source='render_news_section'`, first 07-31, latest 08-03 08:05;
5. the function those rows come from selects on `p.build_status='deployed'` with
   no `p.status` filter (read, not inferred), and the page is
   `status='archived', build_status='deployed'`;
6. fleet-wide that selector matches exactly **one** non-active page, and only two
   statuses exist (`active` 557, `archived` 25).

Each step is a fact about a different table or artefact, and they agree. That is
what is standing in for the loop's verdict.

**Prior art found while chasing it:** `queueNewsPageRerenders` has been through the
council before — correlation `320878ca-5c25-4fed-99ae-82b52b095aba`, routing it down
the no-LLM `page_rerender` path, live in v1.0.1155. That round changed HOW it emits;
neither round changed WHICH pages it selects, which is where the defect was.

## 2026-08-03 — COUNCIL ROUND 2: **REJECTED**, hard veto from `guardian`

Correlation `4a7f0877-4149-4431-97d4-318d093570a4`, round 2. Not a surprise — I named
this exact risk as #1 in my own submission — but the verdict is the verdict, and
CLAUDE.md is explicit: **"A veto on SCOPE is not answered by resubmitting with better
measurements."** So this is not being resubmitted. It is being split.

**The veto, verbatim on the substance:**

> "This is eight files spanning a shared wire struct, the core git-commit function every
> deploy-writing pipeline depends on, a new verb on a shared adapter, and its allowlist
> gate — plus three more new/changed action files. That is architecture-change-dressed-as-fix
> by this council's own definition… The one edit I would approve standing alone is
> `render_news_section_html.go` (add `p.status='active'` to `queueNewsPageRerenders`) —
> that is the actual root-cause fix for the resurrection bug, small, contained, and
> matches sibling query conventions. Safest path: ship that predicate fix now, and take
> `delete_file`/`retract_page_deployment`/`retract_page_graph` through an architecture
> review as a self-contained RFC on the git adapter's action vocabulary, separate from
> the urgent bug. **Splitting does not require re-deriving any of the measurement work
> already done — it only changes sequencing.**"

The `architecture` seat agreed and signalled `needs_rfc`, while saying the design itself
is right — *"expressing delete-as-null-sha inside the existing CommitToRepo path is the
right reuse… On that axis the plan is sound and I'd carry it"*. So the objection is to
**how the capability reached production, not to the capability**. That distinction is the
whole of 124's precedent and it is why more evidence would not have helped.

**What follows, and what does not.** The code is already on shared HEAD and forward-only
forbids removing it; the 2026-07-29 ruling says review here is after the fact by design
and that a thread must not pretend it could have held the change back. So: the predicate
fix stands and is live, the seam is registered (DGH-006) and now carries the veto, and
the RFC goes to architecture review on its own merits for a human to break.

**THE RETRACTION IS NOT BEING PERFORMED.** The owner approved retracting one page before
this verdict existed. Firing a vetoed capability at a live customer site, on the strength
of an approval given on a different basis, would be the exact thing the veto is about.
It waits for the RFC.

### The four objections that are RIGHT ON THE MERITS regardless of the veto

These are real defects in code that is already live, so they matter more than the
packaging dispute:

1. **`debug_historian`, HIGH, edit 8.** My link census does not go through
   `linkablePageStatusPredicate`. The landmine it cites is in my own memory index: *"an
   offline census MUST use `linkablePageStatusPredicate` — an ARCHIVED page makes a
   CORRECT rewrite look wrong."* Building a fresh census over `pages` with a hand-rolled
   status condition is precisely the shape that landmine guards. **Owed: verify and fix.**
2. **`editquality`, MEDIUM, edit 8.** Inbound references also live in **structured
   `content_data` fields** (`link_url`, `cta_url`, hero-tool/tool-cta `*_url`), not only
   in `href=` markup. An href-only scan can miss a real inbound reference and let a
   retraction proceed — *the exact case the owner's directive was written to prevent.*
   **Owed, and it is the sharpest catch in the round.**
3. **`reuse_agent` + `guardian`, MEDIUM, edit 8.** The inbound-source logic is *copied*
   from `check_orphan_pages.go` with a comment warning of drift, rather than extracted and
   shared. I argued the two ask different questions (what IS unreachable vs what BECOMES
   unreachable); the seats are right that a comment is not a mechanism. **Owed.**
4. **`bug_historian`, MEDIUM, edit 7.** The predicate fix invents a third spelling instead
   of consolidating onto a canonical helper, and I presented no fleet-wide census of other
   `build_status`-as-liveness selectors. **Half-answered after submitting**: I ran that
   census and found `queueDirectoryPageRerenders` — which calls itself "cousin of
   queueNewsPageRerenders" — carrying the identical defect (latent, 0 live rows), fixed in
   `8f73e7279`. The consolidation half is still owed and is a fair hit.

The pattern across all four: **every one is about edit 8**, the graph audit I added
fastest and reviewed least. The primitive, which took the most care, drew no correctness
objection at all — only scope.

## 2026-08-03 (evening) — RFC 011 decided, debts paid, and the retraction DONE

**Owner ruled OPTION B.** `delete_file` keeps its place on the adapter, loses its place in
the generic allowlist. Two lines of code, one test, and it concedes the seats' real
complaint (unbounded reachability) without discarding a capability two bugs needed.

**Debt 2 was a REAL defect, not a theoretical one — and the proof is the nicest measurement
of the lane.** Re-running the widened inbound scan on live data, `/contact.html` on
robot-hands went from **4 referrers to 5**: `gripper-cycle-time-estimator/hero` carries its
CTA url in `content_data`, invisible to an `href=`-only scan. The council seat that raised
it could not have known that; it reasoned from a landmine and was right.

I deliberately did NOT enumerate field names to fix it. `ctaFieldNames` exists and was the
obvious tool, and it is exactly the wrong one: a field-name allow-list is blind one level
down, which is what `bugs_open/097` was. Matching the url as a **quoted JSON string value**
finds it at any depth, under any key, including keys nobody has invented yet.

**The retraction ran through the platform**, not by hand — `page-retraction` agent,
orchestration `fc00db29…`. 200 → 404, file gone from the repo, deploy workflow green, and
six neighbouring live pages still 200. The graph audit ran on the live path and agreed with
the hand measurement exactly: `editorial_refs 0, nav_refs 0, stranded 0`.

## 2026-08-03 — MISSTEP, and it is the one I would most want caught: the audit's findings are DISCARDED

Reading the orchestration record afterwards, `collected_data.retraction` holds only the
adapter's reply — `{paths, success, repo_url, …}`. My action's candidates, refusals,
`editorial_inbound`, `nav_inbound`, `stranded_targets` and `nav_retired` are **gone**.

Cause: the step awaits the adapter response, and the await machinery **overwrites
`output_field`** with that response. The findings exist only in pod logs, which this repo
documents as ephemeral across rollouts.

This is `bugs_open/071`/`083`/`091`'s class — detected then discarded — and I wrote the
file header that says *"every refusal is RETURNED, not swallowed"*. It is true of the
function and false of the system. **A retraction that refuses a page would today refuse it
silently**, which is precisely the failure the owner's directive was meant to prevent.

Found only because I read the durable record instead of trusting the green result — the
same habit that found the resurrection. **The check that would have caught it earlier: for
any action that both computes findings and awaits a response, ask where the findings live
AFTER the await, not before it.** An action's return value is not a record.

Owed on the bug as item 5. The fix is to persist the audit before dispatching (or emit a
work item), rather than relying on `output_field` surviving.

---

## CONTRIBUTION 2026-08-03 — loancalculator lane: the sitemap gap, found by retiring a page by hand

I retired `loancalculator.co.uk/tools/standard-calc.html` on an owner decision, used
your audit procedure and your runbook throughout, and hit one thing your primitive
does not cover. Recording it here rather than editing your PLAN.

### `retract_page_deployment` is live but unreachable — your acceptance has not run

- Pod-grep on `agent-chassis-6c76bf964f-4tlrw`: `retract_page_deployment` → **6**,
  positive control `rerender_page_sections` → 20, negative control → 0. It is in the
  running binary.
- But `SELECT ... FROM agent_definitions WHERE default_config::text LIKE
  '%retract_page_deployment%'` → **0 rows**, and the same for `delete_file` → 0. No
  agent can reach either.
- `curl https://robot-hands.com/learning-center/index.html` → **200** at 2026-08-03
  ~21:00, so the live-instance acceptance your PLAN insists on ("the live instance or
  nothing") has not been taken.

I did **not** seat an agent to reach it — that is your platform change to make, and
riding it would have been the scope-veto shape the 2026-07-28 ruling describes.

### The gap: on an ADOPTED site, retraction leaves the sitemap advertising a 404

Your action's guard 4 comment lists `sitemap.xml` among the files "the repo
legitimately holds that `pages` does not model", and excludes it. That is right for a
site whose sitemap the platform generates and will regenerate.

**It is not right for an adopted site.** `loancalculator.co.uk/sitemap.xml` was last
written by the adoption commit `b4302e22b` (2026-07-30) and **the platform has never
regenerated it** — `git log` over that path shows only the adoption and the original
import. It is a static repo artefact listing all 27 urls with `<lastmod>` and
`<priority>`.

So `retract_page_deployment` on this site would have deleted the page and left the
sitemap pointing at it: **the frozen-listing-advertises-a-404 shape that 098 is
about**, reintroduced by 098's own fix. I removed the file and its `<url>` block in
one commit instead (sitemap `<loc>` 27 → 26, XML re-parsed before writing).

**Worth measuring before you decide anything:** how many sites have a sitemap the
platform has never written? That decides whether this is one site's quirk or a class.

### Two smaller things, both confirming your runbook

- **Your `href="…"` audit is the right one and a substring probe is not.** Mine
  reported two inbound links that were both HTML **comments** — one a previous
  session's footer note, one my own from three hours earlier. Your insistence on
  bodies + chrome + nav *with a positive control* is what makes it trustworthy;
  the control returned 1 body / 2 chrome for a page I knew was linked.
- **`gqls/sites` takes concurrent pushes constantly.** My push was rejected twice
  (`non-fast-forward`, then `cannot lock ref … is at <x> but expected <y>`) before
  landing on the third attempt. Anything scripted against that repo needs a
  pull-rebase-retry loop, and must verify at `origin/master` rather than at the
  local ref — a naive `git log -1` after a failed push reports success.

## 2026-08-03 (late) — DEBT 5 PAID: the audit now outlives the await, two sinks, both tested

Committed (Council-Submitted: `5a965452-a9a0-40a6-a990-410f14ac32b0`), **not yet live** —
a chassis build was already in flight from a HEAD that predates the commit, so the next
build after this one is the first that carries it. Pod-grep string for that check:
`retraction refused for page` (a literal the change ADDS; negative control: none removed —
use `conditions_recorded` as a second positive).

**The fix, and why this shape.** The await machinery replaces the step-name and
`output_field` keys wholesale when the adapter reply lands (`applyResponseToState`,
default branch — read, not inferred: `state.CollectedData[stepName] = normalisedData`,
then the same for `output_field`). So:

1. the whole audit is attached to `collected_data.retraction_audit` — a top-level
   SIBLING key, outside the response handler's write set, persisted when the step parks
   to await. Writing side keys into `params.CollectedData` is an established action
   pattern (`image_result`, `final_html`, `__spawn_input_data__`), and the map is the
   coordinator's live state map by reference (`coordinator.go:1682`).
2. refusals become `agent_error_log` rows (`RETRACTION_REFUSED`, one per refused
   candidate, editorial referrers attached; `RETRACTION_STRANDED_TARGETS` as a batch
   summary; severity `warning`) — on REAL runs only, BEFORE the empty-batch return (an
   all-refused run is the loudest silent case) and BEFORE dispatch (a failed send cannot
   unrecord the audit). The orchestration record is retention-clocked (~24h terminal), so
   the record alone was never a durable sink for a refusal; `agent_error_log` is where
   the immune sweep already looks.

**Two design points a reviewer should know:**
- `orchestration.LogAgentError` ("ONE INSERT against agent_error_log") is unreachable
  from the actions package — `coordinator.go:23` imports it, so the call would be a
  cycle. The direct INSERT mirrors its column list exactly, which is also what
  `recordComponentWriteRejection` (this package's precedent for "a refusal must be
  queryable, not pod-log-only") already does.
- The recorder is best-effort by contract, which made "no rows on a dry run"
  untestable through a swallowing recorder — so the audit itself carries
  `conditions_recorded`/`conditions_lost`. That is not only for tests: a lost row must
  not read as a recorded one in the durable record.

**Verified**: 4 action-level tests (audit survives a verbatim overwrite simulation;
dry-run writes nothing durable; all-refused still records; failed dispatch does not
unrecord) — action-level rather than helper-level so deleting a call site fails them.
Mutation-tested both ways: disabling the attach fails 4, disabling the recorder fails
the 3 real-run tests and correctly leaves dry-run green. Built + tested against
`git archive HEAD` + only my two files, because the working tree was churning under
two other sessions (both of whose test-package breakages healed while I watched —
`create_blog_posts_canonical_test.go`, then `toolPageDeclaredSections`, then
`realisedPageIsBuilt`; none were mine to fix).

**NOT done, deliberately:** no work-item emission (a new item_type with no consumer is
queue-noise per 033/071; the error-log route already has a fleet-wide reader), no
coordinator change (merging responses instead of replacing them is a shared-mechanism
redesign — architecture scope), debts 3 and 4 still open.

## 2026-08-04 — the debt-5 council round was KILLED BY THE ROLL; resubmitted under the same trail

Round 1 of `5a965452…` was created at 22:56:40 and FAILED at `review_editquality` with an
EMPTY `__step_error` — both chassis pods restarted 22:56:19/22:56:41, i.e. the run was
claimed in the exact window of the fleet roll the owner had announced. Fresh evidence for
the standing memory line "a roll KILLS an in-flight council"; the empty step error is the
tell that distinguishes a kill from a seat-level failure. Resubmitted 08-04 with
`RESUBMIT_CORR=5a965452…` — the script makes the old correlation the new round's
`fix_correlation_id` (097 line 75), so the commit trailer on `e35e549a8` still resolves.
Round-2 run orchestration: `9e5352cf-7b1f-4d4f-ae83-09923d2baf30`.

## 2026-08-04 — round 2 APPROVED (8 seats, no high-severity), three checks run, RFC 012 filed

`decided_by: "approved with 2 advisory objection(s) — none high-severity"`, 9 abstained.
The commit's `Council-Submitted:` trailer resolves automatically — no amend, per the
07-30 rule. What the seats asked for, and what came of it:

- **editquality (low)**: "no edit corrects the 'refusals are RETURNED, not swallowed'
  comment". **It was corrected in the shipped diff** — the submission's edit list
  under-described the change, the code has it. Lesson: sketch the comment edits too.
- **editquality (low)**: confirm `severity='warning'` matches table conventions. **Run:**
  fatal 1694 / error 1264 / warning 396 / info 1 — 'warning' is an established tier. ✓
- **reuse_agent (missing)**: `RETRACTION_*` code collision. **Run:** no such codes in DB
  or Go; nearest stem is `FIX_PLAN_VALIDATION_REFUSED`. ✓
- **guardian (low)**: fleet-wide confirmation nothing names `retraction_audit`. **Run:**
  0 rows across live `agent_definitions`, no seed mentions it. ✓
- **prior_art_librarian (medium)**: "is the import cycle real, and is
  recordComponentWriteRejection generic enough to take these payloads?" **Answers:** the
  cycle is real (`coordinator.go:23` imports actions); recordComponentWriteRejection is a
  SIBLING, not a duplicate-target — it extracts site/domain/work-item from `input_data`
  paths only, while the retraction resolves its site from config/`site_record`/several
  locations and has the values in hand; extending its signature touches its 3 existing
  callers. The new recorder passes resolved values. Recorded here as the argued case the
  seat asked for.
- **guardian + debug_historian (medium/low)**: the duplicated 13-column INSERT is a real
  running cost (~15 hand-copies in the package now). Folded into RFC 012 §3(c).
- **architecture (approve, with signal)**: the point fix is correctly scoped, AND the
  coordinator's overwrite semantics deserve their own RFC — this is the third-plus action
  to hand-roll the sibling-key escape. **Filed:**
  `architecture_review/RFC_012_the_await_overwrite_destroys_action_findings.md`
  (options costed; filing thread recommends B — named helper + reserved namespace +
  coordinator guard test + shared error-log writer).

## 2026-08-04 — debts 3 and 4 PAID, at a different altitude than the objections assumed

Committed (`Council-Submitted: 37593214-a02a-4f12-b91c-c0704c47037a`), not live until the
next roll — though every substitution renders byte-identical SQL, so the roll changes
nothing observable; what ships is the mechanism that stops the NEXT drift.

**Debt 3 (share the inbound-source logic) — the honest shared seam is the SOURCE LIST,
not the queries.** Reading both censuses settled it: the orphan check substring-matches
(`LIKE '%url%'`) because over-matching only under-reports orphans; the retraction audit
matches quote-delimited + content_data because its unsafe direction is refusing
legitimate retractions (`/index.html` is a substring of `/learning-center/index.html`).
One parameterised query would erase that reasoning. So: `datahelpers.InboundLinkSurfaces`
declares the three tables, each census's SQL is a package-level var, and a lockstep test
on each side asserts every declared surface appears. **Mutation-proven**: appending
`phantom_surface_mutation` to the list failed both packages' tests (4 each), restore →
green. The estate precedent is the dedup-index↔terminal-statuses contract.

**Debt 4 (canonical status predicate) — the census REFRAMED the consolidation.** Three
combinations exist legitimately: lifecycle alone (load_site_pages, plan_sections);
lifecycle+shipped (request_render_audit — 185 t2 chose `PageHasShippedPredicateFor`
precisely because `build_status='deployed'` was the wrong shipped-test); lifecycle+the
exact 'deployed' STATE as a state-machine arm (markPagesForRebuild's
deployed→needs_rebuild transition). A merged "deployed-and-active" helper would
misdescribe two of the three — the consolidatable unit is the LIFECYCLE ARM alone.
`datahelpers.PageWantedLivePredicateFor(alias)` (same one-function-two-forms shape as
`NeverDeployedPagePredicateFor`), seven call sites migrated, both rendered forms pinned
by test.

**Deliberately NOT migrated, and the helper's comment says why:** `loadRetractionCandidates`'
`COALESCE(status,'')<>'active'` — NOT interchangeable with `NOT(status='active')` on a
NULL status (COALESCE-form selects the row, NOT-form drops it; checked before nearly
"tidying" it); and `linkablePageStatusPredicate` (`NOT IN ('deleted','archived')`) which
admits statuses the lifecycle predicate rejects — identical row sets only while exactly
two statuses exist.

Registered as **PBP-029** (+ index row; headline count re-measured 1,717 → 1,718, exactly
my row). HEAD-isolation check: `git archive HEAD` + the 12 changed files builds and
passes all three packages' suites.

## 2026-08-04 — round 37593214 APPROVED (4 advisory, none high) — and its census objection was RIGHT

`bug_historian` (medium + missing): the seven migrated sites were drawn from prior
findings, not a stated audit — "worth a human grep sweep before calling debt 4 fully
paid". **The sweep found five more `pages`-table lifecycle selectors I had missed**, all
now migrated in the follow-up commit: `findStalePages` (maintenance:724 — sitting 26
lines above the one I did migrate), `findOrphanNavPages` (:778), the name-list
needs_rebuild UPDATE (:980), `loadActivePagesForLinkContext` (:1048), and
`render_news_section_action.go:216` (a different file from render_news_section_html.go —
near-namesakes are exactly how a census by memory fails). Also migrated: my own
`loadActivePageFilePaths` COALESCE `=` form (NULL-identical in the `=` direction; only
the `<>` complement differs and keeps its spelling).

**Deliberately SKIPPED, and why:** `tool_acceptance_actions.go` (2 sites) is DIRTY in the
tree — another session's WIP; editing it makes my commit take their half-finished work
(the same-file-passenger rule). `check_page_canonical_collision.go:390` belongs to the
080 lane's fresh PLAN-047 code, committed today — their lane's file, their call. Both are
recorded here so the next sweep knows they are known, not missed.

**Answers to the other advisory objections, so nobody re-derives them:**
- `guardian` (init-order, low): the hoisted package vars call datahelpers functions at
  package init — safe by the Go spec: an imported package is fully initialised before the
  importer's package-level vars; `actions` imports `datahelpers`.
- `guardian` (markPagesForRebuild WHERE, medium): the resulting clause is
  `WHERE status = 'active' AND build_status = 'deployed' AND EXISTS (…sections…)` —
  byte-identical to before; the 'deployed' arm is the state-machine source state and
  stays hand-written.
- `debug_historian` (pod-grep step, medium): debts 3+4 render byte-identical SQL, so the
  binary carries NO new behavioural string to grep — the roll-proof for this lane is debt
  5's added literal `retraction refused for page`, which ships in the same build.
- `editquality` (mediums): real bookkeeping fault in my submission — multi-file edits
  collapsed into single edit entries, and the test file wasn't listed as its own edit.
  Nothing was actually skipped (the commit shows all files), but the lesson stands: ONE
  edit entry PER FILE, and list the test files.
- `architecture` (low): the "drift alarm, not semantic proof" caveat is now IN both
  lockstep tests' doc comments.

## 2026-08-04 — debt 5 verified LIVE at the pod, and the lane is handed off

Fresh chassis deployed (digest `0e99ace…`, both replicas started 10:29). Pod-grep on BOTH
replicas: `retraction refused for page` = 1, `retraction_audit` = 1, and the pre-existing
control `delete_file sent, awaiting response` = 1 (proves the pipeline; the change was
purely additive so no removed-string negative exists — stated, not skipped). Debts 3+4
render byte-identical SQL, so their liveness is unobservable by design and needs no pod
proof. RFC_012 has meanwhile gained an addendum from the bugfix_192 lane — the same
unguarded write on the `storeActionResult` side, with a fleet-wide outage attached — so
the RFC now documents both faces of the class and is the natural place for the owner's
ruling. Session handed off; the refreshed STATE block at the top of the HANDOFF is the
cold-start.
