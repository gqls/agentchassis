# NOTES — bugs_open/315, `pages.deployed_at` without publication

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

## 2026-08-19 ~10:15Z — session start, validity re-check

**Ownership.** `scripts/who-owns.py 315` returns **OWNED or recently active**, naming
`webdesign_tool_rebuilds` (58 commits/14d). Read before acting: that lane's own
`NOTES_native_rebuild_of_ported_tools.md` heads the section *"Three platform defects this lane
filed (all still OPEN, **none this lane's to fix**)"*, and the 08-19 contribution inside the bug
file says *"Neither of us owns this bug and neither is picking it up"*. So the lane owns the
**filing**, not the fix. Proceeding, and contributing INTO the bug file rather than a rival doc.

**Is it still valid?** Yes, re-measured first-hand today.

- `[MEASURED 2026-08-19]` active pages with zero `page_components`, unchanged from the 08-19
  contribution: `planned` **42** pages / 14 sites · `needs_rebuild` **11** / 6 · `deployed` **2** / 2.
- `[MEASURED 2026-08-19]` `SELECT count(*), count(content_hash), count(deployed_at) FROM pages`
  → **786 pages, 0 with `content_hash`, 689 with `deployed_at`.** So §5 candidate 1's premise is
  not merely "empty on those three pages" — **`pages.content_hash` is dead estate-wide, 0/786.**
- The live instance in §3 published itself at 02:42:54Z on 08-19, ~6 h after the fourth rerender.
  The filing lane recorded that and ruled the bug unchanged; agreed — a page that publishes six
  hours late by no action of ours is a description of the defect, not a refutation of it.

**Diagnosis loop filed** (CLAUDE.md: a `bugs_open/` file asserting a cross-cutting root cause is
not "filed" until it has been through the loop). Intake correlation
`d0789788-0501-4b27-a56b-1bb0402af867`, **run correlation `6f900e18-2106-4145-a84c-811baeceaa0d`**
(the run mints its own; the run one is the key the artefacts carry).

## 2026-08-19 ~10:20Z — reading the seam, first-hand

Queue check first: `site_work_items WHERE item_type='needs_diagnosis' AND status IN
('awaiting_diagnosis','diagnosing','pending')` → **0 rows** before filing. Nothing in flight.

**The publish chain for a rerendered page**, read at the live config and the source:

1. `page-rerender` `default_config->workflow->steps`: `render_page` → `save_sections` →
   `deploy_page` (`git_commit`) → **`update_status` (`update_page_status`)** → `complete`.
2. `GitCommitAction` (`platform/orchestration/actions/git_deployer_actions.go`) produces a Kafka
   request to `system.adapter.git.requests` with `AwaitResponse: true`.
3. `GitAdapter.handleCommitAction` (`internal/adapters/git/adapter.go:409`) calls
   `GitHubClient.CommitToRepo`.
4. `CommitToRepo` (`internal/adapters/git/github_client.go:68`) creates blobs, tree, commit, moves
   the ref — and **returns `repo.HTMLURL`**, discarding `newCommitSHA`.
5. `UpdatePageStatusAction` (`platform/orchestration/actions/v3_site_actions.go:679`) runs
   `UPDATE pages SET build_status=$2, deployed_at=NOW(), …`.

**`[MEASURED 2026-08-19]` the adapter's live success payload carries no commit identity.** Taken
from the most recently updated orchestration holding a `deploy_result`
(`webdesign.co.uk`, 2 files, 10:22:03Z):

```
deploy_result.response.data = {success, repo_url, repo_name, domain,
                               files, files_count, file_path,
                               commit_message, timestamp}
```

`repo_url` is `https://github.com/gqls/sites` — a constant, identical for every commit to that
repo. There is **no `commit_sha`**, so nothing downstream can name what was written, and nothing
can distinguish a commit that moved the ref from one that did not.

**`update_page_status` never reads that result.** Its `deployed` branch guards on, in order:
`upstreamAssemblySkipped` (which reads `collected_data["assembled_page"].skipped` —
`owned_page_guard.go:322`), `pageIsArchivedForGuard`, `pageHasComponents`, `pageSectionShortfall`.
The step's own `output_field` for the deploy is `deploy_result`, and that key appears nowhere in
`v3_site_actions.go`. So `GitCommitAction`'s own skip path —

```go
if len(filesMap) == 0 {
    return GitCommitResult{Success: true, AwaitResponse: false,
        Metadata: {"status": "skipped", "skip_reason": "no files to commit"}}, nil
}
```

— returns **success**, and the very next step stamps `deployed_at`.

## 2026-08-19 ~10:30Z — the finding that widens §2 beyond page-rerender

`[MEASURED 2026-08-19]` every step in every active, non-snapshot agent definition, joined on
`next_step`. **19 `git_commit` steps across 16 agents; 6 `update_page_status` steps across 6
agents.** Five of the six stamp `deployed`; here is what precedes each:

| agent | step preceded by | so the stamp is… |
|---|---|---|
| `page-build-handler` | `save_sections` (`save_page_sections`) | **BEFORE any deploy** — its own `update_status` → `spawn_rerender_agent` → `deploy_page` |
| `tool-recreation-handler` | `save_sections` (`save_page_sections`) | **BEFORE any deploy** — `update_status` → `spawn_rerender` → `deploy_page` |
| `page-rerender` | `deploy_page` (`git_commit`) | after a commit **whose result it does not read** |
| `report-builder` | `deploy_page` (`git_commit`) | after a commit **whose result it does not read** |
| `section-editor` | `deploy_page` (`git_commit`) | after a commit it does not read, and *then* `trigger_deploy` |
| `content-reviewer` | `handle_rejection` | writes `needs_attention`, not a deploy stamp — out of scope |

**This is the sharper version of the bug's §2.** The filing lane inferred from three pages that
the column "tracks a rerender ran, not bytes were written". At the config it is stronger and no
longer an inference: in **2 of the 5** deploy stampers the stamp is written *before the deploy is
dispatched at all*, and in the other 3 the deploy's result is discarded. There is no arrangement
of those six workflows under which `deployed_at` could be evidence of publication.

### The trap that would have made this table wrong

`jsonb_each` returns the steps in no meaningful order, and in `page-build-handler`'s row
`deploy_page` prints ABOVE `update_status`. Read by eye, that says "commit, then stamp" — the
opposite of the truth. The `preceded_by` column above is built by joining each step to whatever
step names it in `next_step`, which is the only ordering the engine actually follows. **Any
ordering claim about a workflow has to come from following `next_step`; the key order the row
prints in is not evidence.** Noted here rather than in `WRONG_CALLS.md` because the join was
written before the claim was, so nothing wrong was recorded — but it is a cheap way to be wrong
about a workflow and it is one query to avoid.

## 2026-08-19 ~10:40Z — the delivery path, and TWO dead columns that were built for exactly this

**The delivery path**, from the concept register (`docs026_concept_register/register/deployment-github.md`)
and confirmed against the code:

> commits per page are pushed via the git-adapter to a shared `sites` repo; a **self-hosted GitHub
> Actions runner** fires per commit, **detects changed root-level domain directories**, and runs
> **`b2 sync --delete --skip-newer`** to `b2://portfolio-sites/<domain>`, then purges the Cloudflare
> cache per zone. **There is no separate deploy step — "commit is deploy."**

That is the batch boundary the bug's §2 inferred from two pages sharing a `last-modified` to the
second: **one `b2 sync` per changed domain directory**, not one write per page. It also names the
mechanism for §5 candidate 3 — a page can be dropped by `b2 sync`'s own file-level decisions
(`--skip-newer` skips a destination object newer than its source) without the sync failing, and
without anything upstream hearing about it. **Not yet diagnosed** — the runner's workflow lives in
`gqls/sites/.github/workflows`, outside this repo, and I have not read it. `[UNVERIFIED]` as a
cause; `[MEASURED]` only as the architecture.

### The register asserts a traceability mechanism that does not exist

The same entry says: *"Commit SHAs are recorded on pages and work items for traceability."*

`[MEASURED 2026-08-19]` **All three halves of that sentence are false.**

| claim | check | result |
|---|---|---|
| recorded on `pages` | `information_schema.columns WHERE column_name ILIKE '%commit%' OR '%sha%'` | **no commit/sha column on `pages` at all** |
| recorded on work items | same query | **no commit/sha column on `site_work_items` at all** |
| recorded anywhere | `page_components.deploy_commit` **does** exist | **0 of 1,775 rows populated** |

And the column has no writer: `grep -rn "deploy_commit" --include=*.go .` (whole repo, **including
tests**) returns **zero lines**. It is a schema column that no code in this repository has ever
written.

`pages.content_hash` is the same shape: **0 of 786 rows populated**, and no Go statement writes
`content_hash` on `pages` anywhere in the repo.

**So the platform already carries two columns designed for precisely this question, and both are
inert.** That changes the character of the fix: it is not "invent a traceability mechanism", it is
"drive the one that was already designed and never wired up" — which is CLAUDE.md's *reuse existing
machinery before building new*, and it is why fix candidate 1 is cheaper than it reads.

⚠ **Corrected the register entry in place** (per the standing landmine: *"a concept-register STATUS
line is a snapshot that outlives its truth — and council seats read it as ground truth"*). A seat
reading `deployment-github.md` today would conclude commit-level traceability exists and object to
a proposal that adds it.

### Misstep, logged

Three greps in a row returned no output and I read them as absences. They were run from a
`docs/.../bugfix_315_…` working directory left over from an earlier `cd` in a compound command —
**the Bash tool's working directory persists between calls.** Two of the three ("no `portfolio-sites`
anywhere", "concept register has no publish entry") were flatly wrong, and the register entry that
one of them missed is the single most useful document in this investigation. Logged to
`WRONG_CALLS.md`.

## 2026-08-19 ~10:50Z — sizing the ACTUAL defect, and two corrections to the bug file's own sizing

### First, at the artefact (the bug's §6 method: cache-busted HEAD, `cf-cache-status: DYNAMIC`)

- §3's instance, `webdesign.co.uk/tools/seo-injector/index.html` — **200, `last-modified` Wed 19 Aug
  09:34:01 GMT.** Cleared, as the filing lane recorded.
- The contribution's instance, `vetcomparison.uk/tools/compliance-deadline-calculator/index.html` —
  **still 404 right now.** Live.

### Correction 1 — the contribution's second instance is NOT this bug

`[MEASURED 2026-08-19]` that page's row: `status=active`, `build_status='planned'`,
**`deployed_at` IS NULL**, `content_hash` NULL, 0 components.

**Nothing ever stamped it.** It is an active, link-eligible page that was never built and has never
been served — real, and already the target of `check_componentless_pages` and
`gatherNavLinkedNeverBuilt`. But it is not an instance of "`deployed_at` claims a publication that
did not happen", because no claim was ever made. Fix candidate 1 (stamp only after a confirmed
write) would not have touched it. The contribution called it "same shape as your four"; it is a
different shape that shares a symptom.

### Correction 2 — neither of the "2 at `deployed` with zero components" is this bug either

The contribution called those two "the sharper version of your bug — the estate believes those are
published and they have no components at all." `[MEASURED 2026-08-19]`, both at the DB and at the
artefact:

| page | `deployed_at` | served | what it actually is |
|---|---|---|---|
| `idea.uk` `/tools.html#audience-check` | **NULL** | `/tools.html` 200, LM 17 Aug 19:38:33 | a **phantom row whose `url` is a FRAGMENT** of another page. The real `/tools.html` row exists separately with 4 components and its own stamp. Nothing is unpublished |
| `ai-agent-orchestration.com` `/roi-estimator.html` | 2026-05-02 | **200, LM today 08:37:59** | serving, and rewritten **today**. Two newer roi-estimator pages exist with components; this row is a stale duplicate, not an unpublished page |

So the "42 / 11 / 2" table sizes *componentless active pages*, which is a real and overlapping
population — but **it does not size bug 315**, and the two rows it points at as the sharpest cases
are the two that are not cases at all. Worth saying plainly because that table is the only sizing
the bug file carries, and a fix planned against it would be aimed at the wrong population.

### Now the measurement that DOES size it — and what it says about candidate 4

40 pages, `status=active`, `build_status='deployed'`, ≥1 component, stamped in the last 4 days, not
on the `vm-sites` repo. Cache-busted `HEAD` on each; origin `last-modified` against `deployed_at`.

**`[MEASURED 2026-08-19 ~10:45Z]` 40 of 40 have an origin `last-modified` EARLIER than their own
`deployed_at`, by 50–57 minutes.**

40/40 is the shape that should make you distrust the instrument, not celebrate the finding, so I
looked at the raw values instead of the summary. They are decisive in a different way:

```
distinct origin last-modified across all 40 pages:
    18 × Wed, 19 Aug 2026 09:33:57 GMT
    13 × Wed, 19 Aug 2026 09:33:58 GMT
     9 × Wed, 19 Aug 2026 09:33:56 GMT
```

**Every page in the sample was written to the origin inside the same THREE SECONDS**, while their
`deployed_at` stamps are spread ~25–35 s apart across the hour that followed. That is the
`b2 sync`-per-changed-domain-directory batch, seen directly: the origin is rewritten **per domain,
in one pass**, not per page.

The instrument is sound, and the control is in the data: pages on *other* domains carry *different*
`last-modified` values (`idea.uk` 17 Aug 19:38:33, `ai-agent-orchestration.com` today 08:37:59), so
this is a per-object write time, not one global checkout mtime that would make every file look
identical.

**The consequence for fix candidate 4 is the important part, and it is negative.** At this instant
~40 `webdesign.co.uk` pages are stamped `deployed` with bytes not yet at the origin. Most of them
are almost certainly *fine* — they are inside the normal window between batches. **So a sweep that
flags `deployed_at > origin last-modified` would have flagged all 40, and would have been wrong
about nearly all of them.** Candidate 4 as written in the bug file is not viable: the comparison
has no way to tell "not synced yet" from "will never sync", because the only difference between
them is *elapsed time*, and the bug's own live instance took **six hours**.

⚠ **`[UNVERIFIED]` and the next thing to check:** whether this batch catches up. Re-probe the same
40 pages later in the session. If their `last-modified` advances past their stamps, this is normal
latency and the defect is only the *tail*; if a subset stays behind while others move, that subset
is the real population and the sweep's threshold should be set from it. **That re-probe is the
disconfirming test for everything in this section.**

## 2026-08-19 ~10:35Z — the publish lag, watched live; and a claim I stopped myself making

Following the batch finding, I watched the seam in real time on `webdesign.co.uk`.

**`[MEASURED 2026-08-19 10:34Z]` A confirmed successful commit, and an origin that had not moved
three minutes later:**

- `orchestration_states` — at **10:31:38** the git-adapter returned
  `deploy_result.response.data.success = true` with `file_path = /tools/css-variables/index.html`.
  Its page was stamped `deployed_at = 10:31:28`.
- Cache-busted `HEAD` on that exact URL at **10:34:49** — `last-modified: Wed, 19 Aug 2026
  09:33:58 GMT`. **58 minutes old.**
- And the commits are continuous, not isolated: eight `deploy_result` rows for `webdesign.co.uk`
  between 10:30:34 and 10:34:16, every one `success: true`.

**The runners are healthy and idle.** Two `github-actions-runner` pods, both `1/1 Running`, age
162 m, both `Connected to GitHub` / `Listening for Jobs` on runner 2.336.0. Their job history:

```
brbq6:  08:37:12, 08:40:12, 08:41:10, 08:43:06, 09:33:33, 09:59:38   (all "Succeeded")
tsktm:  08:35:34, 08:37:13, 08:38:37, 08:39:27, 08:43:03, 08:43:29, 09:34:04
```

**Last deploy job of any kind: 09:59:38–09:59:56.** Nothing since, against ~40 successful commits.

### The claim I was about to make, and why I did not

I had this written as a live incident — *"the runners have stopped receiving jobs while commits pile
up"*. **The job history refutes it, or at least removes its force.** Deploy jobs do not arrive
evenly; they arrive in **clusters separated by 25–50 minutes** (08:35–08:43, then nothing until
09:33, then 09:59). The gap from 09:59 to now is **36 minutes — inside that range.** So "60 minutes
of commits with no sync" is not distinguishable, from here, from the normal spacing of this seam.

`[UNVERIFIED]` — I could not settle it: `gqls/sites` is **private** (`api.github.com/repos/gqls/sites`
returns `Not Found` unauthenticated), so I cannot confirm from here that the 10:30–10:34 commits
reached the ref, and the chassis holds no B2 credentials to read the bucket. The background re-probe
now running is the test: if `last-modified` on those pages advances past 09:33:58 within the next
40 minutes, this is latency; if it does not, the gap is real and worth its own bug.

**Why this matters more than the incident would have.** Whichever way the re-probe falls, the
structural point is now measured rather than inferred: **the origin is rewritten in whole-domain
batches, tens of minutes after the commit that earned them, and `deployed_at` is stamped seconds
after the commit.** The stamp is not merely unverified — it is written at a moment when the
statement it makes is *reliably false*, and becomes true later if nothing goes wrong. That is why
this is a measurement defect rather than a race, and it is also why the bug's own §5 candidate 4
cannot work as written: at any given moment a large fraction of correctly-behaving pages are stale
against their stamp.

## 2026-08-19 ~10:40Z — the platform has already solved this exact defect ONE LEVEL UP

This is the most useful thing I have found for the fix, because it means the shape of the answer is
not a design question — it is a precedent to follow.

`platform/orchestration/actions/load_work_item_actions.go:945-956` and
`complete_work_item_verification.go:308-322`:

> *"the envelope's `response_status` only records DELIVERY — the saga's own verdict lives at
> `response.status`. A workflow that never ran at all (unregistered or mistyped action →
> WORKFLOW_INVALID) returns `response.status='failed'`, and used to be stamped 'complete' alongside
> the very error proving nothing was done. The `item_key` dedup then suppressed re-detection, so the
> loop believed the defect class was handled: **54 items across 6 sites** by the 2026-07-18 sweep."*

**That is bug 315, one layer higher.** Same three ingredients: a transport-level success mistaken
for an operation-level success; a status column that then asserts the operation happened; and a
dedup/skip mechanism downstream that treats the false status as settled and stops re-detecting.

It was fixed with `handlerReportedFailure` — a **deliberately narrow predicate keyed on an explicit
failure verdict rather than on the presence of an error string**, and the comment records that the
predicate was **measured against live data before being chosen** ("on the 2026-07-18 sweep, 'failed'
was the ONLY value…").

**So the fix for 315 has a house pattern to copy, not to invent:**

1. name the layer that actually knows (here: the git-adapter's own reply, and beneath it
   `CommitToRepo`'s `newCommitSHA`);
2. carry its verdict up rather than the transport's (`deploy_result.response.data`, which today
   carries no verdict a caller can use — `success: true` is the adapter's, and `repo_url` is a
   constant);
3. gate the status write on that verdict with a **narrow** predicate;
4. **choose the predicate against live data**, and say what values were observed.

Step 4 is the one I can already start: `deploy_result` is present in `collected_data` when
`update_status` runs (confirmed on a live orchestration), so the census of what values actually
appear there across recent runs is available now and is a prerequisite for writing the guard, not a
follow-up to it.

## 2026-08-19 ~10:45Z — the live census the guard has to be built from (house pattern, step 4)

`[MEASURED 2026-08-19]` every `orchestration_states` row holding a `deploy_result`, last 7 days —
**744 rows**, grouped by the four places a verdict could live:

| `metadata.status` | envelope `response_status` | `response.data.success` | action `success` | rows |
|---|---|---|---|---|
| (absent) | complete | true | true | **666** |
| (absent) | complete | **(absent)** | **(absent)** | **57** |
| (absent) | complete | true | (absent) | 19 |
| **`skipped`** | (none) | (none) | true | **2** |

Three things follow, and each of them constrains the fix.

**1. The skip path is live, and it is small.** The 2 rows at `metadata.status='skipped'` are
`GitCommitAction`'s *"no files to commit"* branch: it returns `Success: true`, and the next step
stamps `deployed_at`. So candidate 2's population is **2 in 7 days** — real, worth closing, and not
where the volume is. Anyone sizing the fix off the bug's §3 story would over-weight this arm.

**2. `deploy_result` has (at least) TWO SHAPES, and a naive guard is blind on 7.7% of runs.** The 57
rows with no verdict are not missing one — they are **doubly nested**, because the deploy was done by
a called sub-agent rather than inline:

```
direct :  deploy_result.response.data.success
nested :  deploy_result.response.deploy_result.response.data.success
```

(worked example: `webdesign.uk`, repo `vm-sites`, 10:18:03Z, `success: true` at the inner path.)

**A predicate written against the direct path alone would read `(absent)` on all 57 and — if it
fails open, as it must — would wave them through.** That is the failure mode this whole bug is
about, reintroduced inside its own fix. The guard has to resolve the verdict through the estate's
existing resolver rather than by indexing a literal path: `datahelpers`' whole-tree search and
`tryUnwrapMapPatterns`/`UnwrapDeep` (`unified_extractor.go:508,749`, `content_search.go:195`) exist
precisely for this envelope-unwrapping problem, and RFC_029's recent work made that search's
tie-break deterministic. **Reuse it; do not hand-roll a second unwrapper.**

**3. There is no verdict worth reading yet.** Note what the 666 healthy rows actually contain:
`success: true` from the adapter, and `repo_url` — which is a **constant per repo**. Nothing in any
of the four columns distinguishes *a commit that wrote a blob* from *a commit that wrote nothing*,
because `CommitToRepo` never returns the sha it computed. **So the census also says the guard cannot
be built from the data as it stands** — step 1 of the house pattern (make the layer that knows say
what it knows) is a genuine prerequisite, not a nice-to-have.

## 2026-08-19 ~10:55Z — the fix plan came back; I grounded its load-bearing claims before using them

A subagent's report is another doc with no seam showing where its measuring stopped, so I re-checked
every claim this plan rests on. **All held**, and two of them change the shape of the fix:

| claim | verified at | result |
|---|---|---|
| `UpdatePageStatusInputSpec` is `StrictConfig: true` | `v3_site_actions.go:636` | ✔ — so a NEW config key fails validation on a binary whose spec lacks it: **config must follow the image**, the inverse of the usual seed rule |
| a retired key `commit_from` already describes this feature | `v3_site_actions.go:615-621` | ✔ — *"recording which git commit a page's content was DEPLOYED in, from the git_commit step's output … unimplemented — pages has no such column. Implement it as a feature if wanted, do not re-add the key"* |
| the page-rerender seed already PROMISES a commit sha | `sql_for_agents/034_page_rerender_agent.sql:99` | ✔ — `"deploy_result": "git commit result with commit_sha"` |
| `CommitToRepo` has few callers | `adapter.go:438,518,710` | ✔ — **3 production callers, all in the adapter**; `CreateBranch`/`CreatePullRequest` do not call it. Tests are almost all `_, err :=` and compile unchanged |
| `section-editor`'s deploy output field is not `deploy_result` | live `agent_definitions` | ✔ — it is **`git_result`** |

### Two things this settles that the bug file left open

**1. The bug's §2 table is partly EXPLAINED, and for two of its three rows the behaviour is
correct.** Register `DGH-009`'s landmine (`deployment-github.md:101`):

> *"**`success:true` from the git-adapter is not evidence anything changed.** An unchanged file
> commits as an EMPTY commit and the adapter reports success with the file listed in
> `deploy_result`."*

A rerender that produces byte-identical output commits an empty commit; `b2 sync` then rewrites no
object; `last-modified` stays where it was. **So `tool-json-cleaner` and `tool-smooth-shadow` —
§2's two "stale but serving correctly" rows — are not defects at all.** Their bytes did not need
rewriting. §2 read all three rows as one finding; they are two findings, and only the
`seo-injector` row is the bug.

That does **not** weaken §2's conclusion, it sharpens it: `deployed_at` honestly means *"a rerender
ran and its output was committed"*. It has never meant *"the origin now serves these bytes"*, and
the gap only does damage when the bytes **did** change and still did not arrive.

It also kills the timestamp comparison for good: a check on `deployed_at` vs `last-modified` would
have convicted both healthy pages. **Only an intent-vs-reality content hash separates the three
rows**, which is why fix candidate 1 is load-bearing for candidate 4 rather than merely adjacent to
it.

**2. A guard keyed on a literal field name would be blind on most of the fleet.** The plan spotted
`section-editor`; the full census is worse. `[MEASURED 2026-08-19]` the `output_field` of all 19
live `git_commit` steps — **nine distinct names, and two steps set none at all**:

```
js_snippets_deployed   6      deploy_result          3   (page-rerender, report-builder, site-asset-renderer)
css_deployed           2      (none)                 2   (deployer-agent, site-deployer)
news_commit_result     1      rss_commit_result      1
directory_commit_result 1     sidecar_deployed       1     failed_sidecar_deployed 1   git_result 1
```

So `deploy_result` names only **3 of 19**. Any guard that hard-codes it inherits a 16/19 blind spot,
and — because it must fail open — would wave those through silently. **The guard has to take the
field name from its own step config**, which is also what makes it opt-in and default-OFF, and
therefore not architecture-scope under RFC_022.

### Status of the diagnosis loop, stated plainly

**It did not produce a verdict.** Two dispatches (`6f900e18-2106-4145-a84c-811baeceaa0d`, then
`f1433782-6ba7-4304-a7f9-8bd830dfb7c9`) both died at the `verdict` step on
`AI endpoint unavailable … "You have reached your specified API usage limits"`. The item is parked
at `triaged`.

Per CLAUDE.md's owner ruling of 2026-07-31, a cross-cutting root-cause claim is not "filed" until it
has been through the loop **or the filing session states plainly why it substituted equivalent
first-hand verification**. Stating it: I read every function named in this account at source
(`GitCommitAction`, `CommitToRepo`, `handleCommitAction`, `UpdatePageStatusAction`,
`upstreamAssemblySkipped`), measured the live workflow graph by joining on `next_step` rather than
trusting key order, censused 744 live `deploy_result` rows over 7 days, probed 40 pages at the
served artefact with cache-busters, and read the runner pods' own job logs. **The loop is still
queued and should be read if it ever completes** — and if it refutes any of this, the refutation
wins.

## 2026-08-19 ~10:50Z — the plan's first `[UNVERIFIED]` risk, closed (favourably)

The plan flagged as unverified whether removing the two handlers' early `update_status` step would
break an output contract, since that step's `output_field` is `status_updated`. `[MEASURED
2026-08-19]` at the live config — **it would not**:

| agent | `complete.config.output_fields` | names `status_updated`? |
|---|---|---|
| `page-build-handler` | `["sections_saved", "deploy_result"]` | **no** |
| `tool-recreation-handler` | `["tool_analysis", "sections_saved", "deploy_result", "training_data_saved"]` | **no** |

So the step can be removed without any consumer losing a declared field. **Risk 1 of the council
submission is closed** — recorded here rather than resubmitted, since it narrows a stated risk
rather than changing the plan.

And the replacement stamp is confirmed sound: `page-rerender.update_status` is configured
`{"status":"deployed","page_id_field":"rendered_page.page_id"}` — it stamps **by page id from the
render it just did**, which is a stronger identity than the two handlers' own steps, both of which
resolve the page by *name* (`site_id_field` + `page_name_field`). So the reorder does not merely
move the stamp later; it also moves it onto a more precise identifier.

### An incidental confirmation of the nesting finding

Both handlers declare `deploy_result` in their `complete` output fields while their own `deploy_page`
step is a **`call_agent`**, not a `git_commit`. That is exactly the shape that produces the
doubly-nested `deploy_result.response.deploy_result.…` envelope measured earlier at **57 of 744
rows** — the sub-agent's whole collected data comes back under `response`. The two findings were
made independently and agree, which is worth more than either on its own.

## 2026-08-19 ~10:50Z — the plan's fourth risk, closed at the register

Risk 4 asked whether anything downstream of the commit rewrites page bytes — because if it does,
hashing what the adapter was handed would not be comparable with served bytes, and the whole
divergence check would need to hash somewhere else.

**It does not.** Register `DGH` on the serving hop (`deployment-github.md:125`):

> *"The `portfolio-sites-router` worker maps `hostname + path` **straight onto a B2 object key**."*

A pure key mapping (its only logic is appending `index.html` to a path ending in `/`), and the hop
before it is a `b2 sync`, which copies rather than transforms. So the bytes committed are the bytes
served, and `files_sha256` taken at the adapter is directly comparable with a sha256 of the served
response. **Risks 1 and 4 of the council submission are now both closed**; risks 2, 3 and 5 stand as
stated (they are design choices, not unknowns).

## 2026-08-19 ~10:55Z — COUNCIL VERDICT: **REVISE**, and it was worth the round

`SUBMISSION_CORR 377167cd-6324-4bc7-a866-87ad8c435132` — **revise**, `decided_by: "gating objection
from editquality"`. Five seats approved (reuse_agent, diagnosis_guardian, render_guardian,
constitution, mission, debug_historian), five objected, five abstained.

⚠ **Reading trap, hit and caught:** `SELECT body FROM doc_notes WHERE categories ? 'council-gate'
ORDER BY created_at DESC LIMIT 1` — the query printed in the trigger's own output — returned a
**different lane's APPROVED verdict** (`e4840008-…`). The table is fleet-shared and a bare `LIMIT 1`
races every other submitting session. **Read the verdict by correlation**, from
`diagnosis_artifacts WHERE correlation_id='<yours>' AND kind='council_report'`. Had I not checked
the correlation I would have recorded an approval that belonged to someone else.

### The objections that are RIGHT and change the plan

**1. `editquality` (HIGH) — the summary overclaims, and it is my error.** `deploy_result_field` is
opt-in/default-OFF and this submission ships no migration setting it, so *"deployed_at only ever
written downstream of a real commit result"* is **false for 3 of the 5 agents** on merge. Only an
unused capability would exist. Correct, and not curable by better wording alone — the scope limit
has to be stated as a scope limit.

**2. `editquality` (MEDIUM) — edit 1 relocates the unguarded stamp, it does not gate it.** Dropping
the pre-deploy stamp hands the job to `page-rerender`'s own `update_status` — but `page-rerender` is
*itself* one of the three ungated post-commit stampers. **I missed this.** The edit is still an
improvement (post-commit on a page id beats pre-dispatch on a page name) but it is a smaller claim
than I made.

**3. `architecture` (needs_rfc) — edits 2–3 widen a SHARED wire shape and are not gated by the
opt-in key.** The adapter's reply is consumed by 19 live `git_commit` steps across 16 agents, and I
surveyed none of them for how they parse it. Recommendation: **ship edit 1 now; take the adapter
contract + payload through architecture review with the 19-step consumer list.** That is the
2026-07-28 platform-seams ruling applied to my own change, and it is right.

**4. `guidelines` (MEDIUM) — nested additions to an already-flowing shared object must be named in
the seam's concept-register entry IN THE SAME COMMIT.** I cited `deployment-github.md` for DGH-009
and then did not update it. A process omission, and the register entry is exactly what I corrected
earlier today for being stale.

**5. `debug_historian` (MEDIUM) — the migration needs a counted needle-gate and an explicit
`snapshot_agent()`,** not a post-hoc existence assertion as its only verify step.

### The objection I can ANSWER with evidence rather than revise

**`guardian` / `prior_art_librarian` — "is `CommitToRepo` behind an interface? the 3-caller claim is
asserted, and this package has a `TreeEntry` landmine for exactly that."** Fair challenge, and now
settled: `[MEASURED 2026-08-19]` `grep -rn "CommitToRepo(ctx context.Context" --include=*.go .`
returns **one line** — the concrete method on `*GitHubClient`. `interface.go` does not mention it,
and `grep -rln "GitClient\b" --include=*.go .` returns **nothing**: no such interface exists
anywhere in the repo. So it is a 3-caller change, and the answer is a citation, not a redesign.

### And the one that found a REAL DEFECT in my plan

`prior_art_librarian` objected that repurposing the two dead columns is "load-bearing and directly
checkable". Checking it found something worse than it suspected — in the estate's own SQL:

`sql_for_agents/356_retire_dead_config_keys_commit_from...sql:105-118`:

> *"`page_components.deploy_commit` EXISTS … **`pages.deploy_commit` was dropped by
> `sql_for_tables/003` as "belongs in page_components"** … NULL in it today means "never
> implemented", NOT "never deployed" — and **deciding whether to wire it or drop it is an owner
> call, not a bug fix.**"*

**Two things follow, and both change the plan.**

- **My edit 5 must go.** It proposed `ALTER TABLE pages ADD COLUMN deploy_commit`, justified as
  *"it once did and it was dropped — so restore it"*. It was dropped **deliberately, on a stated
  design ground**. Re-adding it is a reversal of someone's decision, not a restoration, and the
  column that survived (`page_components.deploy_commit`) is the one the estate chose. Logged to
  `WRONG_CALLS.md`.
- **Wiring `deploy_commit` at all is flagged in-tree as an OWNER CALL.** That is not mine to take
  inside a bug fix. It goes to the owner as a question.

Also found in the same pass: `sql_for_agents/291:24-26` independently measured `content_hash` dead
(0 of 1,183 rows, 2026-08-02) and **deliberately did not use it**, hashing `md5(rendered_html)`
instead. So content_hash's emptiness has now been discovered independently three times and acted on
zero times — which strengthens the case for wiring it, but again as a stated decision rather than a
silent one.

## 2026-08-19 ~11:00Z — the revised submission is OUT OF THE GATE'S SCOPE, and that is the correct outcome

Revised per round 1 and resubmitted with `RESUBMIT_CORR=377167cd-…`. The gate refused it
**client-side, before spending anything**:

> `REFUSED: no edit touches the review scope (platform/, internal/, pkg/ — owner ruling 2026-07-17).`
> `Docs and site content do not spend council credits. FORCE=1 to override.`

**I did not force it, and that is the point.** Once the architecture seat's ruling is honoured —
edits 2–3 out to architecture review, edit 4's guard out with them because it can only reference
fields they add, edit 5 withdrawn as an owner call — **what remains is a config migration and a
register update, which is not platform code.** The gate is telling me the narrowed change is not the
kind of thing it reviews, which is true.

So there is no round-2 verdict to obtain, and the honest position is:

- **Round 1 = REVISE**, and the record is `377167cd-6324-4bc7-a866-87ad8c435132`.
- The surviving change is the one the **architecture seat cleared in that round in so many words**:
  *"Edit 1 (config reorder) is a clean point fix — proceed."*
- The commit carries **`Council-Submitted:`**, never `Council-Reviewed:` — the verdict I read was
  `revise`, and writing `Council-Reviewed:` on it would be the MISMATCH the coverage report exists
  to catch.
- The register update the `guidelines` seat required ships **in the same commit** as the migration,
  per platform-seams condition 2.

Forcing would have spent credits re-reviewing a config change the gate declares out of scope, on a
round whose substantive review has already happened. Cost of not forcing: nothing. Cost of forcing:
a duplicate round and a weaker signal for the platform-code case the gate exists for.

## 2026-08-19 ~11:00Z — CORRECTION to my own 10:35Z entry: the 40 stale pages are BENIGN, and proving that is the argument for the fix

I wrote at 10:35Z that *"at this instant ~40 `webdesign.co.uk` pages are stamped `deployed` with
bytes not yet at the origin"*, hedged it, and set a re-probe. **The re-probe ran for 25 minutes and
`last-modified` never moved** (09:33:58 throughout, reaching 85 minutes of apparent staleness), which
looked like confirmation. It was not. Two checks settle it the other way:

**1. A deploy job DID run, and afterwards nothing changed.** `[MEASURED]` runner `tsktm` logged
`10:54:06Z: Running job: deploy` → `10:54:31Z: Job deploy completed with result: Succeeded`. My
10:56:17Z probe is *after* that job, and still reads 09:33:58. So the pipeline is not stalled — it
ran and wrote nothing.

**2. The origin is serving the CURRENT content.** `[MEASURED 2026-08-19 ~11:00Z]` — the decisive
check, and the one I should have run first:

```
DB: page_components.rendered_html for /tools/css-variables/index.html
    6,350 chars, md5 4cef84c7…, updated_at 2026-08-15 17:23:46
needle = substring(rendered_html, 200, 120)   -- a distinctive slice of the live component
curl the served page (cache-busted) -> 13,674 bytes; grep -F -c <needle> -> 2 occurrences
```

**The served page carries the current database content.** The component has not changed since
**2026-08-15**; the origin was last written at 09:33:57 today (a chrome/asset wave); every rerender
since has regenerated byte-identical output, committed an **empty commit** per `DGH-009`, and
`b2 sync` has correctly rewritten nothing.

So: **not a defect, not an incident, and `last-modified` is behaving exactly as it should.**

### Why this is the strongest evidence in the whole investigation

Look at what it took to answer *"is this page correctly published?"* — read the component out of the
database, cut a needle from its stored HTML, fetch the served page, and grep. **Four steps and a
judgement call, for one page.** And until I did it, the two hypotheses — *"the bytes never changed"*
and *"the bytes changed and never arrived"* — were **indistinguishable from every signal the platform
exposes**: same work-item status, same orchestration outcome, same `deployed_at`, same
`success: true` from the adapter, same unmoved `last-modified`.

That is `bugs_open/315` in one paragraph. **The bug is not that pages fail to publish; it is that the
platform cannot tell the difference between a page that did not need publishing and one that failed
to.** `pages.content_hash` collapses that four-step archaeology into one comparison, which is why
candidate 1 is the load-bearing fix and candidate 4 is worthless without it.

⚠ **And it is a warning about the sweep's design.** My own probe produced a confident-looking 40/40
"stale" result that was **entirely false**, from a method the bug file proposes as fix candidate 4.
A sweep built that way would file 40 work items today, all wrong, on the fleet's most active site.
The settle-window mitigation in the PLAN is **not sufficient on its own** — 85 minutes had elapsed
and the answer was still "fine". Only the hash separates them. Recorded in the PLAN as a hard
constraint, not a preference.

## 2026-08-19 15:20Z — MIGRATION 491 APPLIED (owner authorised). 2 of 5 → 0 of 4.

Applied **scoped**, because three other lanes had pending files (488 ×2, 489, 490) and an unscoped
`--apply` takes every one of them:

```bash
SCOPE=<scratch>/mig491; cp docs/agent_docs/sql_for_agents/491_*.sql "$SCOPE/"
MIGRATIONS_DIR="$SCOPE" ./scripts/migration/run-migrations.sh          # Pending (1) — probe ok
MIGRATIONS_DIR="$SCOPE" ./scripts/migration/run-migrations.sh --apply  # DO DO UPDATE 1 UPDATE 1 DO COMMIT
```

⚠ The assignment is on the **same line** as the command — on its own line it scopes nothing and the
run sweeps every other thread's pending migration. The scratch dir is the scoping mechanism.

### Verified at the live config, not at the migration's own say-so

```
type                     still_has_stamp   save_next               target_exists
page-build-handler       f                 spawn_rerender_agent    t
tool-recreation-handler  f                 spawn_rerender          t
```

And **the census re-run — the number the whole change exists to move:**

| agent | status | preceded by |
|---|---|---|
| `content-reviewer` | needs_attention | `handle_rejection` (not a deploy stamp) |
| `page-rerender` | deployed | **`deploy_page(git_commit)`** |
| `report-builder` | deployed | **`deploy_page(git_commit)`** |
| `section-editor` | deployed | **`deploy_page(git_commit)`** |

**Every remaining `deployed` stamper is now preceded by a `git_commit`. Zero stamp before a deploy —
down from 2 of 5.**

### The snapshot check the landmine demands

Not *"does a snapshot exist"* but *"does it hold the PRE-change config"* — a snapshot carrying the
post-change value restores nothing:

```
type                     snapshot_taken_at    snapshot_reason                                  holds_pre_change_step
page-build-handler       2026-08-19 15:20:41  491_drop_pre_deploy_deployed_stamp: pre-update   t
tool-recreation-handler  2026-08-19 15:20:41  491_drop_pre_deploy_deployed_stamp: pre-update   t
```

Found by `snapshot_reason` (the distinctive second argument) and ordered by `snapshot_taken_at`, per
the two `snapshot_agent` landmines — `agent_definitions_backup` copies the SOURCE row's `id` and
`created_at`, so `ORDER BY created_at` returns an arbitrary snapshot.

### What is NOT yet verified, and what to watch

**This is verified at the config; it has not yet been observed at RUNTIME.** No first build has run
through the rewired path since 15:20Z. The failure direction is the recoverable one — un-stamped
rather than falsely stamped — but the thing to check on the next `page-build-handler` run is that
`page-rerender`'s stamp actually lands:

```sql
-- a page built since 15:20Z should end up with a deployed_at from the RERENDER, not from the handler
SELECT p.name, p.build_status, p.deployed_at
FROM pages p WHERE p.updated_at > '2026-08-19 15:20:00+00' AND p.build_status='deployed'
ORDER BY p.deployed_at DESC LIMIT 10;
```

If pages start accumulating at `planned`/`needs_rebuild` with no `deployed_at` after a successful
build, the rerender's stamp is not firing and
`491_drop_pre_deploy_deployed_stamp_ROLLBACK.sql` restores both steps surgically (deliberately not a
blob restore — 488 is still pending against `page-build-handler` from another lane, and a whole-config
restore would silently revert it).

## 2026-08-19 ~15:40Z — 491's RUNTIME check PASSES, so the rollback is not needed

My 15:20Z entry left this open: *"verified at the config; it has not yet been observed at
RUNTIME … the thing to check on the next `page-build-handler` run is that `page-rerender`'s stamp
actually lands."* It landed.

`[MEASURED 2026-08-19 ~16:50 BST, re-run first-hand]` pages updated since 15:20Z:

```
build_status | count
-------------+------
deployed     |    31        <- ONE row. Nothing at planned, nothing at needs_rebuild.
```

and the most recent all carry a fresh `deployed_at` from the rerender, not from a handler:
`news-index` 16:00:31, `product-detail` 15:46:06, `services` 15:45:12,
`pneumatic-vs-electric-grippers` 15:44:11, `learning-center` 15:42:22.

**So removing the pre-deploy stamp has not stranded a single page**, and `491_…_ROLLBACK.sql` is
not needed on this evidence. ⚠ **What it does NOT show:** that the bytes arrived. It proves the
stamp still lands. That is §2, and it is exactly what `content_hash` is for.

### Credit, and the grounding I did anyway

This check was handed to me by the **`agentchassis-b0` session**, at the owner's request, with the
figures above (they measured 30; I measured 31 — the count moved between us, which is what a live
system does). They also re-measured `content_hash` **0 of 790** and `page_components.deploy_commit`
**0 of 1,789**, and confirmed **4** live `update_page_status` steps, down from 5.

**I re-ran the load-bearing one myself before recording it**, because a peer session's report is
another doc with no seam showing where its measuring stopped — the same rule as a subagent's. It
held. Their other offering — that §3's `seo-injector` instance has moved again (`last-modified`
now Wed 19 Aug **14:42:12**, `scriptOpenTag` 2, `ported-page` 0, `b-type` 0) — I have not re-run,
and they explicitly make no claim about *why* it moved. Marked `[UNVERIFIED BY ME]`, and it changes
nothing: the page clearing itself is a description of the bug, not evidence against it.

## 2026-08-19 ~15:45Z — the fix is BUILT (both halves), and one twin deliberately left alone

**Half 1 — the git-adapter says what it did** (`0c5b94725`). `CommitToRepo` returns
`CommitOutcome{RepoURL, CommitSHA, Branch, AbsentPaths}` instead of `repo.HTMLURL`, and the reply
carries `commit_sha` plus **`files_sha256`**.

The hashing lives in the *adapter*, not the git client, for two reasons that are not stylistic: the
keys there are still the **caller's own paths** (`CommitToRepo` prefixes `{domain}/` onto its private
copy, and the chassis looks a page up by the path it sent), and hashing is pure, so it has no
business inside the git plumbing.

⚠ **The base64 branch is the part that would have failed silently for ever.** A base64 file's
`content` string is a transport wrapper, not the file; hashing the wrapper yields a value that can
*never* equal a sha256 of what the origin serves, and nothing would ever report the mismatch. It is
not hypothetical — `derive_card_asset_action.go:202` and `derive_brand_head_assets_action.go:169`
send base64 PNGs today. **Mutation-proved:** removing the decode fails exactly that assertion;
restoring it passes.

Unusable input is **omitted, never hashed wrongly** — a missing key means "no fingerprint available"
and a reader can act on it; a wrong one means "this page is broken" and it cannot.

**No `NoChange` flag, and that is a deliberate narrowing of my own plan.** It needs the PARENT
commit's tree sha, and the parent tree is off the hot path — `getLatestCommitSHA` returns a *commit*
sha and `getBaseTreeSHA` (which does read a tree sha) is only its error fallback. So it would cost a
GitHub round-trip on **every** commit across 19 live `git_commit` steps, to populate a field the
council ruled **report-only**. The per-file hashes answer the same question better, at the grain the
site actually serves.

**Half 2 — the chassis reads it** (`086f9b7b7`). `update_page_status` takes one optional,
default-OFF key naming the deploy step's output field; refuses the stamp when that step reported a
skip; and writes `pages.content_hash` in the same UPDATE as the stamp, via
`COALESCE($3, content_hash)` so a stamp with no fingerprint leaves a good hash alone rather than
nulling the only evidence there is.

Resolution goes through `datahelpers.ExtractFields` **scoped to the named subtree**, not a path
index. **Mutation-proved:** swapping in a literal `<field>.response.data` read fails exactly the
nested-shape test and nothing else — which is the 7.7% of live runs, and failing there means failing
*open* on all of them.

Three outcomes kept distinct because they demand opposite responses: **skipped** → refuse;
**resolved** → stamp and fingerprint; **unreadable** → stamp anyway and write a
`DEPLOY_EVIDENCE_UNREADABLE` row. ⚠ I graded that row **`warning`, not `high` as the plan said**: the
chassis and the git-adapter are separate images, so a chassis carrying this key against an adapter
that predates RFC_038 resolves nothing on **every** deploy. That is an expected, bounded rollout
window, and grading it `high` would have made the fleet's error log useless for a day.

### The untouched twin — checked, and deliberately not changed

`pattern-check` flagged `UpdatePageComponentsStatusAction` as an unedited twin of the function I
changed. Checked rather than waved through: its only write is

```sql
UPDATE page_components SET build_status = $1, reviewed_at = $2, reviewed_by = $3, updated_at = NOW() ...
```

It never touches `pages.deployed_at` or `content_hash`, so **it makes no publication claim and does
not carry this defect.** Extending it would mean writing `page_components.deploy_commit` — which the
owner ruled out the same day. Not-changing it is right twice over.

## 2026-08-19 ~17:05Z — I hit a landmine that was ALREADY DOCUMENTED, twice

The council-verdict trap I recorded earlier today — reading `doc_notes … ORDER BY created_at DESC
LIMIT 1` and getting another lane's APPROVED verdict while mine was `revise` — is already in
`LANDMINES.md`, at **two** separate entries (lines ~667 and ~9064), the second titled almost exactly
what happened to me.

So I have not added a third copy. What is worth recording is **why the existing entries did not
reach me**: the `SessionStart` hook matches landmines against files already **dirty in the working
tree**, and this trap's footprint is a *query* and a *table*, which no path can match. The remedy is
the one the memory index already states and I did not follow —

> **grep `LANDMINES.md` for the SYMBOL, TABLE or COMMAND you are about to trust, not just the path.**

One `grep -n "council-gate" LANDMINES.md` before running the verdict query would have cost seconds.
This is the second time in this session the same class has bitten: the first was believing three
empty greps that were empty because of a stale working directory. Both are "the tool answered a
question I had not checked I was asking".

## 2026-08-19 ~17:10Z — what is DONE, and what is left

**Done and live:** migration 491 (pre-deploy stamp removed from the two handlers; verified at the
config and at runtime — 31 pages since, none stranded).

**Done, committed, NOT live** — both need an image:
- `0c5b94725` git-adapter: `CommitOutcome`, `commit_sha`, `files_sha256`.
- `086f9b7b7` chassis: `deploy_result_field`, the skip refusal, `pages.content_hash` written at the
  stamp.
- `494_stamp_reads_deploy_evidence_HOLD.sql` — **held on purpose**, must not be applied until the
  rebuilt chassis is running (StrictConfig).
- Registered as **DGH-013** with its index row, in the session it shipped.

**Blocked on the owner:** the build. Releases are whole-fleet and the owner runs them
(`MEMORY/releases-are-whole-fleet-make-release`), so building and rolling the git-adapter and
agent-chassis images is not mine to do.

**Designed, not built:** the divergence sweep (PLAN D5) — the piece that would actually have caught
this bug at 15:18 on the day it happened. It is gated on `content_hash` being populated, which is
gated on the roll. Until then this change is **provenance, not detection**, and DGH-013 says so.

## 2026-08-19 ~20:30Z — COUNCIL ROUND 2: **REVISE**, and it caught a false claim of mine

`377167cd` round 2 (the IMPLEMENTED code, submitted after the fact because round 1 judged a *plan*
and what shipped was narrower). **`revise`**, `decided_by: "gating objection from
prior_art_librarian"`, 7 abstained.

### The gating objection was right, and it is the most valuable thing the gate has done here

I wrote, in `deploy_evidence.go` and in the submission, that resolution is safe because
`datahelpers.ExtractFields` is **"collect-all / unique-or-nothing (RFC_029 §9)"** — so an agent with
several `git_commit` steps "resolves to nothing rather than to a guess". The seat checked that
against a `LANDMINES.md` entry stating the opposite (*"resolves its inputs by RANDOMISED recursive
search — the wrong sibling's id wins"*) and objected that the two claims cannot both be true.

**Reading `findFieldRecursive` TO THE END settles it against me.** The ruling *is*
unique-or-nothing — and then:

> *"**PHASE 1 (this build — instrument first, refuse second): conflicts still resolve**, to the
> STABLE shallowest-first winner, and emit the WARN below … **PHASE 2 (a later build) flips
> conflicts to refusal.**"*

Phase 2 has not shipped. **I quoted a comment I had not read to the end**, and built a safety
argument on the half that stated the intention rather than the half that stated the behaviour.

### Why it mattered here more than it would elsewhere

For most callers a shallowest-first guess is survivable. For this one it is the worst outcome
available: a fingerprint taken from the WRONG `git_commit` is **silently and permanently wrong**, and
every later comparison would report a healthy page as diverged. *No* fingerprint is recoverable;
*someone else's* fingerprint is not.

### The fix — make the property true rather than restate it

`resolveDeployEvidence` no longer borrows the guarantee. It collects candidates itself
(`collectUniqueValue`, a deliberately dumber walk — no unwrap patterns, no aliases, no ranking,
because ranking is only needed when you intend to PICK) and **refuses on conflict**. Agreeing
duplicates still resolve, or the guard would be useless on ordinary nested runs.

**Mutation-proved**, which is what the `editquality` seat asked for by name: restoring the Phase-1
guess (`return cur, true, false` on disagreement) fails exactly
`TestResolveDeployEvidence_AmbiguousSubtreeREFUSES` and `TestCollectUniqueValue_ConflictBeatsFound`
— and nothing else. Restored, green.

### A second defect the same reasoning exposed, which nobody objected to

`content_hash` was written with `COALESCE($3, content_hash)` — my own round-1 reasoning, *"a stamp
with no fingerprint must leave the previous one alone"*. **That is backwards.** A stamp means new
bytes went out, so any prior fingerprint describes an OLDER deploy and is stale by definition.
`COALESCE` preserves it, and the divergence check would then compare live bytes against a superseded
intent and **convict a healthy page** — the exact false-positive class I spent the morning proving
was fatal to candidate 4.

Now: the column is touched **only when the guard ran**, and then **assigned**, NULL included. NULL
means "we do not know what we sent", which is precisely what the check's `content_hash IS NOT NULL`
predicate is for. Guard off ⇒ the clause is not in the statement at all, so an unarmed path cannot
disturb a hash another path wrote.

### `bug_historian`'s coverage objection — answered with the measurement it asked for

The objection: the key is opt-in, 494 arms only 3 steps, so *"the defect remains fully live and
silent on every unconfigured step"* of the other ~16.

`[MEASURED 2026-08-19 20:35Z]` of the **19** live `git_commit` steps, exactly **3** are followed by
an `update_page_status`:

```
report-builder : deploy_page(deploy_result) -> update_status
page-rerender  : deploy_page(deploy_result) -> update_status
section-editor : deploy_page(git_result)    -> update_page_status
```

**Those are exactly the three 494 arms, with exactly those field names.** The other 16 deploy CSS,
JS snippets, RSS and directory files and have **no page stamp after them at all** — there is no
`deployed_at` claim to guard. So coverage of *this defect* is complete, not partial. (The objection
would be right about a different defect: those 16 also get no commit-evidence check. Nothing
currently claims anything about them.)

### Two seats I did NOT act on, and why

- `reuse_agent` (medium): should `deploy_result_field` extend `ActionInputSpec`/`input_fields`
  instead of a bespoke resolver? Fair in shape, but the *reason* for the bespoke path is now
  stronger, not weaker: the shared resolution ladder **resolves conflicts in this build**, and this
  caller must refuse them. Extending the shared mechanism to refuse is RFC_029 Phase 2's job and has
  a stated precondition (zero conflict WARNs over an observation window). Recorded rather than done.
- `prior_art_librarian` (medium): the absence claims (`content_hash` 0/790; "no step reads a field
  out of the reply") are *"stated as MEASURED but not independently checkable from this review"*.
  True and unavoidable — the seat cannot run queries. The queries are in the RUNBOOK and in
  RFC_038 §7 so the next reader can re-run them.

### ⚠ THE OPERATIONAL CONSEQUENCE

`v1.0.1316` (rolled 17:13Z) **carries the flawed resolver** — proven at the artefact, not the tag:
the git-adapter's own stamp is `git_commit: 07eeba4a1…`, and `git merge-base --is-ancestor` puts
both `0c5b94725` and `086f9b7b7` behind it.

It is **inert**, because `deploy_result_field` is set nowhere and 494 is still held. **So 494 MUST
NOT be armed until a build carrying `f0dd97c71` has rolled.** Arming it against 1316 would run the
version that guesses.

⚠ Also noted: the binary probe was **useless** here — `grep -a <full sha> /proc/1/exe` returned
*absent* for my commit AND for an older one, while the all-zeros control returned **PRESENT** (Go's
internal tables), exactly as `MEMORY/a-fresh-deploy-can-ship-no-new-code` warns. The git-adapter's
own startup log line is what answered it, because it is a quiet service and the line was still in
range; on the chassis it had already scrolled.

## 2026-08-19 ~21:00Z — COUNCIL ROUND 3: **APPROVED**

`377167cd` round 3 — **`approved`**, *"approved with 2 advisory objection(s) — none high-severity"*,
7 abstained. The trail is round 1 REVISE (plan) → round 2 REVISE (implemented code, found a real
false claim) → round 3 APPROVED.

`Council-Reviewed: 377167cd-6324-4bc7-a866-87ad8c435132` — written only now, having read an approved
verdict. Earlier commits on this trail carry `Council-Submitted:`, which asserts nothing and which
`098` resolves at report time.

### The advisory I ACTED on — and it was a real latent bug

`editquality` (medium): the stamp statement includes a `content_hash` clause **conditionally** and
appends its arg **conditionally**, while the placeholder was written as a literal `"$3"`. Those are
two facts that must agree and nothing forced them to.

Today it happens to be correct — with two base args, the hash is genuinely `$3`. **But the literal is
a trap set for the next person who adds an arg**, and the failure is a runtime `psql` error on the
deploy path, invisible to the compiler and to any test that exercises only one branch.

Fixed by **deriving** the index from the slice: `fmt.Sprintf("content_hash = $%d,", len(args))`.

⚠ **And my first test for it was worthless, which is the more useful lesson.** I wrote it to *mirror*
the construction rather than call it — so it would have passed happily while production was broken,
which is precisely the failure it was written to prevent. The construction is now extracted as
`buildPageDeployStampQuery` and the test calls the real thing.

**Mutation-proved that the test can fail at all** — a test that cannot fail is worth nothing, and this
one passes against BOTH the literal and the derived form today. So the mutation had to simulate the
future it protects against: add a third arg AND hardcode `$3`. Result:

```
guardRan=false: statement references up to $2 but 3 args are supplied
guardRan=true : statement references up to $3 but 4 args are supplied
```

Both branches fail; restored, green.

### The advisories I did NOT act on, and why

- `bug_historian` (medium): *"one call site of a shared judgement gets the rigorous fix; the sibling
  stays heuristic"* — `collectUniqueValue` gives THIS caller unique-or-refuse while
  `findFieldRecursive` still resolves conflicts for everyone else. **Correct, and it is RFC_029 Phase
  2's job, which carries its own stated precondition** (zero conflict WARNs over an observation
  window, or explicit mappings first). Recorded as a real follow-up in the submission's risk 2, in
  `DGH-013`, and in the strengthened landmine: **delete `collectUniqueValue` when Phase 2 ships.**
- `debug_historian` (medium): the "v1.0.1316 carries the flawed resolver" claim rests on the
  git-adapter's build-provenance LOG plus `merge-base`, not a binary grep. **That IS the pod, and the
  binary grep was measured unusable here** — absent for a commit the image demonstrably carries,
  PRESENT for a 40-zero control. Recorded in `494`'s header so the next person does not reach for it.
- `guardian` (low ×2): flags that the conditional clause touches the shared page-stamping path, and
  asks me to confirm `494` stays HELD. Confirmed — this round does not arm it, and `494`'s gate
  commit was moved to `f0dd97c71` precisely so it cannot be armed against the flawed build.

### ⚠ ONE COMMIT IS BLOCKED, and deliberately so

`platform/orchestration/actions/v3_site_actions.go` **cannot be committed right now.** Another
session is mid-flight in the SAME FILE: the working tree's copy now contains their
`refuseMistypedLLMFields(config)` call, whose definition lives in two files that are still
**untracked** (`actions/mistyped_llm_fields_gate.go`, `datahelpers/content_type_violations.go`).

A pathspec commit takes the working-tree file whole, so committing mine would carry their call
without its definition and **HEAD would not compile** — and `make build-*` builds from committed
HEAD, so that breaks builds for every session. This is the documented same-file-passenger landmine,
met in its worst form.

**Verified my earlier commit did NOT carry it:** `git show f0dd97c71 -- <file> | grep -c
refuseMistypedLLMFields` → **0**. Their call arrived afterwards.

So the extraction is proven and held. What is pending, as one atomic set:
`deploy_evidence.go` (the builder) + `deploy_evidence_test.go` (the test) +
`v3_site_actions.go` (the call site). **Splitting them is worse than waiting** — committing the
builder alone would leave HEAD with an unused helper beside a live inline copy, and the test would
then exercise the helper while production used the duplicate. That is the mirror problem relocated,
not solved.

Proven meanwhile against a clean `git archive HEAD` tree with my three files applied: build clean,
vet clean apart from the pre-existing `load_component_library_actions.go:207` warning, all 11
deploy-evidence tests green.

**Nothing is at risk from the wait:** the code is inert (`deploy_result_field` set nowhere, `494`
held), and the placeholder fix is protection against a future arg, not a live defect.

## 2026-08-19 22:26Z build → **494 ARMED 2026-08-20 ~06:50Z.** The fingerprint is switched on.

### The gate was checked per SERVICE, and the control discriminated both ways

`v1.0.1317` rolled 22:26:03–22:26:27Z (chassis and git-adapter). 494's gate is
`git merge-base --is-ancestor f0dd97c71 <stamp>`, and `f0dd97c71` is the **corrected** resolver —
`086f9b7b7` alone would have passed a check written before round 2 and armed the version that
guesses.

- **git-adapter**, from its own startup line: `build provenance git_commit=2d13d530d…`, and
  `f0dd97c71` **is** an ancestor of it.
- ⚠ **My first "negative control" was worthless and I caught it**: I used a commit of my own that
  turned out to predate the build, so of course it read as present. The second attempt was
  degenerate too — HEAD *is* the build stamp, and a commit is its own ancestor. The probe was only
  shown to discriminate by the **reverse** test: `--is-ancestor <stamp> f0dd97c71` is FALSE.
- **chassis**, probed per service because the landmine says per service, not per fleet — and its
  provenance line had already scrolled. Symbol probe with a **present/absent pair**:

  | symbol | in `f0dd97c71`? | probe result |
  |---|---|---|
  | `collectUniqueValue` | yes | **PRESENT** |
  | `resolveDeployEvidence` | yes | **PRESENT** |
  | `buildPageDeployStampQuery` | **no** — only in my still-blocked commit | **absent** ✓ control |

  This is the landmine's own prescription (*"verify a KNOWN value … always run a control in the same
  breath — a sha that must be absent, and one that must be present"*) and it worked where the
  40-hex-sha probe failed completely yesterday.

### Applied, and verified at the live config

```
type            armed_with
--------------  -------------
page-rerender   deploy_result
report-builder  deploy_result
section-editor  git_result
```

Three different field names, as measured — `section-editor` really does use `git_result`, and a
literal would have missed it.

⚠ **`--record-only` does not work on a `_HOLD` file** and my own header and runbook both told the
next person to run it: the runner refuses (*"is an UPPERCASE-suffixed sidecar … recording one is
meaningless"*). Harmless — a sidecar never appears in Pending so it cannot be double-applied, and
the file's `RAISE '494: already applied'` guard catches a human re-run. **Both instructions
corrected; the apply is recorded HERE instead, which is now the only place it exists.**

### Baseline before the first armed deploy

`[MEASURED 2026-08-20 06:50:18Z]` `SELECT count(*), count(content_hash) FROM pages` → **802 pages,
0 with a hash.** That is the number this whole lane exists to move, and it is the honest
before-reading: a watcher is now waiting for the first non-zero.

**Nothing is proven until it moves.** Config being right is not the artefact — that is this bug's
entire lesson, and it applies to the fix as much as to the defect.

## 2026-08-20 ~06:55Z — armed but UNEXERCISED, and the zero is traffic, not failure

`[MEASURED 2026-08-20 06:55Z]` after arming: `count(content_hash)` = **0**;
`agent_error_log WHERE error_code='DEPLOY_EVIDENCE_UNREADABLE'` = **0**; pages stamped since
06:50Z = **0**; and the most recent `page-rerender` orchestration of any kind completed
**2026-08-19 15:19:01Z**, roughly fifteen hours ago.

**So both zeros are consistent with "nothing has run", and neither is evidence the guard works or
that it doesn't.** A post-change zero with nothing driving it is the shape
`MEMORY/a-post-fix-zero-needs-a-demand-control` exists to warn about, and it is the same mistake
this bug is about — reading a green status that no traffic has tested. Recording it as unproven
rather than as a pass.

### Why I have NOT driven one myself

`docs/leopardessconsulting/scripts/rerender_page_safe.sh` is the documented safe trigger and the
chassis is well past its 300s post-restart window, so it is available. I did not use it, for a
reason worth stating rather than assuming:

**a rerender is not content-neutral.** That script sets `input_data.spec.reason` to
`section_data_resolved`, which its own header says makes page-rerender **REGENERATE** section HTML
from `content_data` rather than re-assemble the stored HTML. Regenerated output is not guaranteed
byte-identical to what is live — which is the whole point of
`MEMORY/repro-regenerated-from-source-is-destroyed-by-the-render`. Every candidate target is a live
customer page belonging to another lane (`webdesign.co.uk`'s tools are the
`webdesign_tool_rebuilds` lane's active work), so firing one to test **my** fix would risk changing
**their** page.

The content-neutral variant — a rerender WITHOUT that reason, which re-assembles the stored HTML and
so commits byte-identical output — would be the right instrument, and is a payload change rather
than a script that exists. Offered to the owner rather than taken unilaterally.

**Meanwhile a watcher is running** for the first non-zero `content_hash`, so organic traffic will
answer it: the fleet stamped 31 pages yesterday afternoon, so this is hours, not days.

### What each outcome will mean when it lands

| observation | reading |
|---|---|
| `content_hash` non-zero after a rerender | the guard resolved real evidence and the fingerprint is live — **the thing this lane exists for** |
| `DEPLOY_EVIDENCE_UNREADABLE` rows instead | the chassis is armed against something it cannot read. With `v1.0.1317` carrying BOTH halves this would NOT be the expected partial-roll window — it would be a real defect in the resolution path and the first thing to read is the field name per agent |
| a `deployed` stamp with neither | the guard did not run at all — check `deploy_result_field` is still set (another session edits these agents constantly) |

## 2026-08-20 07:25Z — INBOUND from the `webdesign_tool_rebuilds` lane: 494 IS ARMED, it broke the fleet's publish path, and I have rolled it back with YOUR rollback script. `bugs_open/336`.

Not a takeover — a report plus a restore, written where you will see it. Full evidence:
`bugs_open/336_HANDOFF_2026-08-20_deploy_result_field_is_declared_on_the_wrong_actions_spec_so_arming_it_hard_fails_every_workflow_that_stamps_a_page.md`.

**What happened.** 494 was applied at **06:49:49Z** (all three definitions: `page-rerender`,
`report-builder`, `section-editor`). At **07:01:50Z** the first item claimed after that died with

```
WORKFLOW_INVALID: … step 'update_status' (action 'update_page_status') has unrecognised
config keys [deploy_result_field] — this action declares its config contract as complete
```

and every claim after it did the same: `[MEASURED 07:20Z]` 8 items — 4 `page_rerender`, 2
`needs_content_page`, 1 `section_edit` — with **123 `page_rerender` queued fleet-wide and none
draining**. I ran
`docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD_ROLLBACK.sql` verbatim at
**07:22:40Z** (it snapshots all three first, so your arming is one command away again). All three
definitions verified clear.

**Your precondition was MET, and that is the part worth your attention.** Your note says 494 must not
be armed until a build carrying `f0dd97c71` has rolled. `v1.0.1317` (pods 2026-08-19 22:26Z, stamp
`2d13d530d`) has BOTH `086f9b7b7` and `f0dd97c71` as ancestors — so whoever armed it this morning was
following your instruction. **The precondition was about the READER shipping; the defect is in the
DECLARATION.** `deploy_result_field` sits in `RenderComponentInputSpec.ConfigKeys`
(`v3_site_actions.go:674`) — the spec for `render_component`, which never reads it — while the reader
is `UpdatePageStatusAction` (:982) and `UpdatePageStatusInputSpec` (:550-556) declares exactly five
keys and sets `StrictConfig: true`. Both specs are registered in the same `init()` forty lines apart,
so each looks right in isolation.

**Two instruments that will lie to you here, both tried on the way in:**
- `grep -aq "deploy_result_field" /proc/1/exe` on the chassis returns **PRESENT**. The literal is in
  the binary three times (the reader and two `zap.String` calls) regardless of which spec declares it.
  Presence of the literal is not membership of the right list.
- `git log -S'"deploy_result_field",'` names `086f9b7b7`, which reads like the declaration's commit.
  That match is `zap.String("deploy_result_field", field)`. I nearly concluded the declaration shipped
  in the live build on that basis; what settled it was reading the LIST inside the named spec at that
  commit.

**Also, for your outcome table.** Your third row says "a `deployed` stamp with neither → the guard did
not run at all — check `deploy_result_field` is still set (another session edits these agents
constantly)". That is now exactly the state, and the reason is this rollback rather than an editing
session — so when you next see it, read this entry before re-arming.

**Six items are still `failed` and dead to the dispatcher** having failed on the platform defect rather
than their own work: `9cb5d4e5`, `e291e4ea`, `5887736c`, `126c586a`, `a0015980`, `35972e9b`
(`a0015980` is webdesign.co.uk's `index` page; `5887736c` is a tool page of mine). I am flipping the
two of mine back; the others are listed here so nobody has to rediscover them.

## 2026-08-20 09:17–09:19Z — the held commit was taken from me, HEAD broke, and it is fixed

The peer symbols I was waiting on (`mistyped_llm_fields_gate.go`,
`datahelpers/content_type_violations.go`) landed in HEAD. But by then my hold had already been
defeated:

**`80b9c6235`** — the `bugs_open/260` lane, an unrelated commit about the component render seam —
took `v3_site_actions.go` from the working tree, and **my `buildPageDeployStampQuery` call site rode
along as a same-file passenger** while its definition was still uncommitted.

`[MEASURED]` a clean `git archive HEAD` build then failed:
`v3_site_actions.go:1062:17: undefined: buildPageDeployStampQuery`. **HEAD did not compile, and
`make build-*` builds from committed HEAD**, so builds were broken fleet-wide.

Fixed by `460ff6b3d`, landing the definition. **HEAD was broken for 1.8 minutes** (09:17:27 →
09:19:17) and no build started inside it — the chassis pods are still the 22:26Z ones. Verified after:
clean-archive build clean, `DeployEvidence|CollectUnique|StampStatement` green.

### This is my misjudgement, not the 260 lane's

CLAUDE.md predicts this exactly — *"committing per task stops **you** sweeping up others' WIP; it
cannot stop a session that still runs `git add -A` from sweeping up **yours**"* — and no hook can see
a same-file passenger. So the fault is not in their commit; it is in my having left a caller
uncommitted where a broad commit could find it.

**And I had a better option, which I reasoned my way past.** I wrote that splitting the set was
"worse than waiting" because it would leave a duplicated construction and a test pointed at the wrong
copy. True — and a *quality* wart, traded for a *breakage* risk. The asymmetry I missed is the
compiler's, not anyone's discipline:

- a function with **no caller** compiles;
- a caller with **no definition** does not, and takes the repo with it.

**So the definition should have gone first, alone.** Atomicity is not available on a shared working
tree; only ordering is. Logged to `WRONG_CALLS.md` with that as the rule.

### Current state of the code half

Everything is now in HEAD and consistent: the corrected resolver (`f0dd97c71`), the extracted
builder + its real test (`460ff6b3d`), and the call site (via `80b9c6235`). Nothing of mine is dirty.
`494` is armed. What remains is runtime evidence, which needs a deploy.

## 2026-08-20 06:49–07:22Z — **I BROKE THE FLEET'S PAGE-PUBLISHING PATH.** Corrected account.

My 07:52Z entry above says "armed but UNEXERCISED … both zeros mean 'nothing ran'". **That entry is
WRONG and I am leaving it in place with this correction beneath it**, because the way it was wrong is
the most useful thing in this file.

The zeros did not mean *nothing ran*. They meant **nothing could run.** Arming 494 made
`deploy_result_field` an **undeclared key on a `StrictConfig: true` spec**, which the validator turns
into a hard failure for the entire workflow:

```
WORKFLOW_INVALID: step 'update_status' (action 'update_page_status') has unrecognised
config keys [deploy_result_field] — this action declares its config contract as
complete, so an unknown key is a definition error, not a no-op
```

**Cause, and it is mine:** I appended the key to **`RenderComponentInputSpec.ConfigKeys`** instead of
`UpdatePageStatusInputSpec.ConfigKeys` — the same file, forty lines apart, two specs registered in one
`init()`. I anchored the insertion on the literal `"strip_literal_markdown"`, which belongs to the
other spec, and never checked which block I had landed in.

**Damage:** every workflow with an `update_page_status` step died at validation. 8 items carried it
(4 `page_rerender`, 2 `needs_content_page`, 1 `section_edit`; 6 left `failed`), with **123
`page_rerender` items queued fleet-wide and none draining.** Armed 06:49:49Z, first failure 07:01:50Z,
restored 07:22:40Z.

**Found and restored by the `webdesign_tool_rebuilds` lane** — as a blocked served-page grade, of all
things — who diagnosed it to the line, filed `bugs_open/336`, ran **my own rollback file**, and framed
it as a report rather than a takeover. Their rollback ran twice (07:22:39 and 08:16:00), which is why
the arming was gone when I looked.

### Why every check I ran passed. This is the part to keep.

- **My arming precondition was correct, met, and aimed at the wrong thing.** I had established 494
  must wait for `f0dd97c71` — and it had. But that condition was about the **reader** shipping.
  Nothing in it could see a declaration on the wrong spec.
- **The binary probe would have said PRESENT and been useless.** The literal is in the chassis three
  times (the reader and two `zap.String` calls). An hour earlier I had used a present/absent control
  pair on *function names* and been pleased that the probe discriminated. It does — for symbols.
  **Presence of a literal is not membership of the right list.**
- **I verified the config and never verified the artefact.** I ran the three-agent query, got the
  three expected field names, wrote "verified at the live config" — and it was already broken as I
  typed it. **That is `bugs_open/315`'s own defect, committed by the lane fixing `bugs_open/315`.**
  The entry directly above this one contains the sentence *"Config being right is not the artefact —
  that is this bug's entire lesson, and it applies to the fix as much as to the defect."* I wrote it
  and then did not do it.

### The rule, and it is not "be more careful"

**When you arm a switch, the FIRST query is "what did I break?", not "did it work?"** They hit
different tables and the second returning zero looks identical either way:

```sql
-- what did I break  (I never ran this)
SELECT count(*) FROM orchestration_states WHERE error ILIKE '%<your new key>%';
SELECT status, count(*) FROM site_work_items WHERE item_type='<the affected type>' GROUP BY status;
-- did it work      (I ran only this, and read its zero as "no traffic")
SELECT count(content_hash) FROM pages;
```

And at the edit, scope the grep to the block you believe you changed:
`awk '/^var UpdatePageStatusInputSpec/,/^}/' v3_site_actions.go | grep deploy_result_field`.
**Anchoring on a nearby literal is not anchoring in the right scope** — in a file with two sibling
structs it is a coin flip, and I lost it.

### State now

- Declaration **fixed at HEAD** by `daaa7541b` (not my commit): the key is in
  `UpdatePageStatusInputSpec` and absent from `RenderComponentInputSpec`; the "five ConfigKeys"
  census comment corrected.
- **Fleet healthy, checked rather than assumed:** `orchestration_states WHERE error ILIKE
  '%deploy_result_field%'` → **0**; `page_rerender` 5,189 complete / 1 claimed / 0 blocked on this.
- ⚠ **494 is UNARMED and MUST STAY UNARMED.** `[MEASURED]` `daaa7541b` is **not** an ancestor of the
  running build (`v1.0.1317`, stamp `2d13d530d`, built 2026-08-19 22:21:54). **Re-arming today
  reproduces the outage exactly.** 494's header now gates on `daaa7541b` and carries the post-arm
  damage queries.
- The durable guard `336` proposes — every key an action reads is declared on its own spec, and no
  spec declares a key its action never reads — is the right answer and I have deliberately not
  claimed it. My lane should not grade its own homework on this seam.

## 2026-08-20 — the content_hash watcher is MOOT while 494 is unarmed; do not restart it as-is

The background watcher waiting for the first non-zero `count(pages.content_hash)` has been stopped.
**It should not simply be restarted**, and the reason is the same mistake in miniature: with 494
unarmed, `deploy_result_field` is set nowhere, so the guard never runs and `content_hash` **cannot**
be written. A watcher on that column would sit at zero indefinitely and its silence would read as
"still waiting for traffic" — which is precisely the misreading that let the 33-minute outage stand
unnoticed for twelve minutes.

**Whoever arms 494 should start it then, not before**, and should pair it with the damage query, not
run it alone:

```sql
-- run BOTH, and read the damage one first
SELECT count(*) FROM orchestration_states WHERE error ILIKE '%deploy_result_field%';   -- must stay 0
SELECT count(content_hash) FROM pages;                                                  -- the benefit
```

A watcher on the second alone cannot distinguish *nothing has run* from *nothing can run*. That is
not a hypothetical — it is what happened.
