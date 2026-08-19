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
