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
