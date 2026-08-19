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
