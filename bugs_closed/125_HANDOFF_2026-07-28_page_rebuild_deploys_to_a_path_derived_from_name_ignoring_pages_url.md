# 125 — `page-rebuild` deploys to a path derived from the page NAME, ignoring `pages.url`, creating an orphaned duplicate at the wrong URL

**Filed** 2026-07-28 by the bug-sweep thread, **caused and observed live** while running
`bugs_open/087`'s acceptance test · **Status** OPEN, unowned ·
**Severity** HIGH — **280 of 431 pages (65%) would deploy to the wrong path**, and each
one publishes a real, fetchable duplicate of a live page ·
**The one live instance was REMOVED by the owner 2026-07-28 ~07:2x UTC; the cause is untouched**

---

## Symptom, observed

Running `page-rebuild` for `finetuning.uk` (one armed page,
`ai-agent-roi-estimator`) rebuilt the page and committed it to:

```
"file_path": "/ai-agent-roi-estimator.html"
"files":     ["/ai-agent-roi-estimator.html"]
"domain":    "finetuning.uk"      "success": true
```

But that page's canonical URL, in the same page object the step was handed, is:

```
"url": "/tools/ai-agent-roi-estimator.html"
```

Measured before and after, same session:

| URL | 07:00 UTC | 07:12 UTC |
|---|---|---|
| `/ai-agent-roi-estimator.html` | **404** (302 b) | **200 (29,521 b)** ← created by the rebuild |
| `/tools/ai-agent-roi-estimator.html` | 200 (35,129 b) | 200 (35,129 b), **byte-identical** |

So the real page was untouched and a **second, orphaned copy of the same tool now
serves at a different URL with different content**.

## Root cause

> **CORRECTED 2026-07-31 (bugfix-8 session): the function is `determinePageFilename`
> at `git_deployer_actions.go:374`. There is no `resolveFilePath` in this repo** — the
> name below is wrong and a grep for it returns nothing, which reads as "the code has
> been refactored away" rather than "the bug file has the wrong name". The bug-sweep
> handoff (`bug_backlog_clearing/HANDOFF_2026-07-28_bug_sweep_continue_here.md` §4a)
> has the correct name. The line range is also off by ~40. Everything else in this
> section is accurate.

`resolveFilePath`, `platform/orchestration/actions/git_deployer_actions.go:414-445`.
Given the page object it tries, in order:

```go
if slug, ok := p["slug"].(string);        ok && slug != ""     { return ensureHTMLExtension(slug) }
if name, ok := p["name"].(string);        ok && name != ""     { return ensureHTMLExtension(name) }      // ← matched
if pageName, ok := p["page_name"].(string); ok && pageName != "" { return ensureHTMLExtension(pageName) }
if filename, ok := p["filename"].(string); ok && filename != "" { return filename }
if id, ok := p["id"].(string);            ok && id != ""       { return ensureHTMLExtension(id) }
```

**`url` is not in the list.** It is never consulted, at any priority. The page
object supplied by `get_pages_to_build` carries `url` — the canonical, correct
path — and the deployer discards it in favour of a path synthesised from `name`.

For a page whose URL happens to be `/<name>.html` the two agree by coincidence.
For anything in a subdirectory they do not.

## Blast radius — this is not one odd page

```sql
SELECT count(*) FILTER (WHERE url <> '/'||name||'.html') AS wrong_path,
       count(*) FILTER (WHERE url =  '/'||name||'.html') AS right_path,
       count(*) AS total
FROM pages WHERE url IS NOT NULL AND url <> '';
--  wrong_path | right_path | total
--         280 |        151 |   431
```

**65% of pages carrying a URL would be deployed to the wrong path** by this
resolver. Every `/guides/…`, `/blog/…` and `/tools/…` page across the estate:

```
ai-agent-orchestration.com | password-entropy          | /tools/password-entropy.html
ai-agent-orchestration.com | ai-readiness-quiz-guide   | /guides/ai-readiness-quiz-guide.html
ai-agent-orchestration.com | multi-agent-state-management-distributed-systems | /blog/…
```

## Why it has stayed hidden until today

`page-rebuild` never reached `deploy_page`. It died several steps earlier at
`process_sections_loop` with `key 'sections_ready' not found` — that is
`bugs_open/087`, filed 2026-07-26. Fixing 087 (migration 246, applied
2026-07-27) let the workflow run to completion for the first time, and the very
first complete run exposed this.

**A defect behind a defect.** Nothing here is new code — `resolveFilePath` is
unchanged. 087 was simply masking it, so the fleet has been one working rebuild
away from this since the resolver was written. It matters for the other lanes
too: **any** caller of `git_commit` with `page_field` inherits the same rule.

## Damage done, and the cleanup this thread could not do

`https://finetuning.uk/ai-agent-roi-estimator.html` — **created 2026-07-28
~07:09 UTC by this test, live now.** It is an orphaned duplicate: nothing links
to it, it is not in `pages`, and it will not be regenerated or cleaned by any
existing sweep.

I could not remove it. There is no credentialed path to `github.com/gqls/sites`
from this environment (`git ls-remote` → *could not read Username*), and the git
adapter's only deletion verb is `delete_repo`, which returns
*"delete_repo action not yet implemented"* (`internal/adapters/git/adapter.go:641`).
**Removing that one file needs the owner's git access, or a `delete_file`
capability that does not exist yet.**

Note the wider gap this exposes: **the platform can publish a page but has no
implemented way to unpublish one.** That is the same shape as `bugs_open/098`
(archiving does not retract from the deployed site).

## Fix candidates, ordered by what closes the door

**1 — Prefer `url` in `resolveFilePath`, above `slug` and `name`.** Three lines:
try `p["url"]` first, fall through to the existing chain when absent. Makes the
canonical path the default and the synthesised one the fallback, which is the
right precedence — `url` is the field the rest of the platform treats as
authoritative. Go, so it needs a build and a roll. **Check first** whether any
caller depends on the synthesised path, and whether `url` is ever stored without
a leading `/` or without `.html` (`ensureHTMLExtension` handles the latter).

**2 — Refuse to deploy when `url` is present and disagrees with the resolved
path.** Fail loudly rather than writing to a path nobody asked for. Smaller
behavioural surface than 1 and turns a silent duplicate into a visible error, but
it stops rebuilds working until 1 lands.

**3 — A `delete_file` verb on the git adapter**, so orphans of this class can be
retracted at all. Does not fix the cause; it is what makes the consequence
recoverable, and `bugs_open/098` needs the same primitive. Worth doing once for
both.

**4 — A fleet sweep for orphaned duplicates**: files in the sites repo with no
matching `pages.url`. Detection only — but this defect has probably fired before
today wherever anything else called `git_commit` with `page_field`, and nothing
would have reported it.

**1 + 3 together** is the pairing that fixes the cause and makes today's instance
removable.

## How to verify a fix

Re-run the `087` acceptance test on the same page — `finetuning.uk`,
`ai-agent-roi-estimator`, `url = /tools/ai-agent-roi-estimator.html` — and assert
the **path**, not the success flag: `collected_data->'page_deployed'->'response'
->'data'->>'file_path'` must equal `/tools/ai-agent-roi-estimator.html`. The run
already reports `"success": true` while writing to the wrong place, so the
success flag is exactly the thing that must not be trusted here.

Then confirm no NEW file appears at `/ai-agent-roi-estimator.html`, and remove
the existing one.

**Related:** `bugs_open/087` (the defect that masked this; its fix is what
exposed it), `bugs_open/080` (gap planner bypasses canonicalisation — same
duplicate-page class, arrived at from another direction), `bugs_open/098`
(archiving does not retract a deployed page — same missing unpublish primitive).

---

## Cleanup CONFIRMED 2026-07-28 — the orphan is gone, the real page never moved

The owner deleted the file from the sites repo. Verified against the **live
site**, not the repo, because a deletion in git is not a retraction from a CDN:

```
/ai-agent-roi-estimator.html                       -> 404 (302 b)
/ai-agent-roi-estimator.html?cb=<epoch>            -> 404 (302 b)   cache-busted
/tools/ai-agent-roi-estimator.html                 -> 200 (35,129 b)
```

The cache-busted request matters: an unbusted GET is a cache's opinion of a page,
so a plain 404 could have been an edge holding a stale negative. It is genuinely
gone.

And the real page is **byte-identical to the pre-test capture** — 35,129 bytes,
sha `b1f0afe6c03e67bf`, taken at 07:07 UTC before the rebuild committed. So the
test left **no residue at all**: it neither altered the canonical page nor left
the duplicate behind.

**The CAUSE is untouched.** `resolveFilePath` still never consults `url`, and the
next `page-rebuild` of any of the 280 mismatched pages recreates an orphan
exactly like this one. This section records that the *instance* was cleaned, not
that the bug was fixed — fix candidates 1–4 above all still stand.

**Still true and worth keeping in view:** removing that one file needed the
owner's own git access, because the platform has no implemented way to unpublish
a page (`delete_repo` is the git adapter's only deletion verb and it is a stub).
Candidate 3 remains the one that makes this class recoverable without a human
with repo credentials.

---

## Severity refined 2026-07-28 — LATENT, not firing. It is a landmine, and 087 just re-armed it

Measured after filing, because the original text left the impression this might be
happening fleet-wide right now. **It is not.** The distinction matters for
prioritisation and it does not reduce the eventual blast radius.

### Only THREE steps use the buggy resolution, and none of them runs

`determinePageFilename` (`git_deployer_actions.go:368-400`) has **four**
priorities, not the three I first read — I missed `filename_field` at priority 1.
Classifying every live `git_commit` step by which one it hits:

| priority | steps | verdict |
|---|---|---|
| P1 `filename_field` | 1 — `section-editor::deploy_page` | safe |
| P2 `file_path` (static) | 2 — the two `deploy_css` steps | safe |
| **P3 `page_field`** | **3** — `pageflow-builder`, `page-rebuild`, `site-work-orchestrator` (all `…_loop/deploy_page`) | **this bug** |
| P4 default `index.html` | 16 | see caveat below |

And the three P3 agents have not run. Checked by **structural signature** rather
than by agent-type extraction, which leaves 570 of 1,913 rows unresolved and
would have produced a false zero:

```sql
SELECT count(*), max(created_at) FROM orchestration_states
WHERE workflow_plan::text LIKE '%build_pages_loop%';   -- 1, 2026-07-28 07:03  (MY test)
WHERE workflow_plan::text LIKE '%build_items_loop%';   -- 0
-- retention floor: 2026-07-13 → 2026-07-28, 1,913 rows
```

**Across fifteen days of retained history the buggy path has executed exactly
once: my own acceptance test.** So there are no other orphans from this cause in
the retention window, and candidate 4 (a fleet sweep for orphaned duplicates) is
lower priority than filed — though orphans predating 07-13 cannot be ruled out
this way.

### Why this still matters, and matters more than "latent" suggests

`page-rebuild` is dormant **because it was broken** — that is `bugs_open/087`.
Fixing 087 makes it work, and a working `page-rebuild` on any of the 280
mismatched pages publishes an orphan. **087's fix is what converts this from a
dead code path into a live one.** The two must ship together, or 087 should not
be routed traffic until 125 is fixed.

That is the real finding of this pair: *unblocking one defect armed another*, and
neither file could see it alone.

### ⚠️ P4 is NOT claimed as a bug — I checked myself before writing it down

Sixteen steps configure neither `filename_field`, `file_path` nor `page_field`,
which by the priority list defaults to `index.html`. That reads alarmingly —
`page-rerender::deploy_page` alone has **49 runs** — and it is almost certainly
fine: `determinePageFilename` serves the SINGLE-file commit path, and there is a
separate multi-file path (`filesMap`, same file) those steps likely take. **I have
not traced it, so it is recorded as unverified rather than filed.** If the estate
were overwriting `index.html` 49 times it would be extremely visible, and it is
not. `[UNVERIFIED — needs the filesMap path read before anyone acts on it.]`

---

## FIX COMMITTED 2026-07-31 (bugfix-8 session) — candidate 1, widened to the whole class

Lane docs: `docs024_key_docs_latest/bugfix_125_deploy_path_from_url/`.
Commit `5dc177f97`. Council `Council-Submitted: 758f6e62-99b8-4f33-a81b-7143351ecd69`.

### Re-measured before touching anything — the filed figure had moved

| | filed 07-28 | measured 07-31 |
|---|---|---|
| wrong path | 280 | **316** |
| right path | 151 | 156 |
| total with a url | 431 | **472** |

Still 2/3 of the estate, and 41 pages larger. Same query as the one in this file.

### What shipped, and why it is not the three lines candidate 1 describes

Candidate 1 ("try `p["url"]` first") closes the instance and leaves the class.
Grepping the **derivation** instead of the symptom
(`grep -rn 'TrimPrefix(.*[Uu][Rr][Ll], "/")' platform/ internal/`) found **five**
places that turn a page into a deploy path — and **four of them already consult `url`
first and get it broadly right**:

| site | consults `url`? |
|---|---|
| `datahelpers/file_extractor.go:194` `determineFilename` | **yes** — comment reads *"Try url field first"* |
| `rerender_single_page_action.go:521` | yes |
| `get_pages_for_rerender_action.go:176` | yes |
| `rerender_pages_actions.go:324` | yes |
| **`git_deployer_actions.go:374` `determinePageFilename`** | **NO — this bug** |

So this is a **duplicated classifier that drifted**, not a missing feature — and the
correct implementation sitting eleven characters away in the name (`determineFilename`
vs `determinePageFilename`) bought nothing, because the wrong copy is the one the three
build pipelines reach. Shipped as one definition —
`datahelpers.PageFilePathFromURL` / `PageDeployFilename` — with all five call sites
moved onto it, plus the lockstep tests 016b §9 asks for (56 cases).

### Two things in this file's pre-work that did not survive checking

1. **"The fix must strip `#…`" (sweep handoff §4a) is the wrong repair and would be
   destructive.** The single fragment row is `idea.uk` / `tool-audience-check` →
   `/tools.html#audience-check`, and **`/tools.html` is a different page's canonical
   url** (`idea.uk`/`tools`, measured). Stripping the fragment aims one page's rebuild
   at another page's file — strictly worse than the bug. A url with a fragment names no
   file of its own, so the helper **declines** it (`ok=false`) and the caller falls back
   to its own chain. Making an input *valid* and making it *correct* are different
   operations.
2. **The leading slash is not mentioned anywhere in the pre-work and is load-bearing.**
   `pages.url` is site-absolute on **472/472** rows and `CommitToRepo` builds
   `data.Domain + "/" + path` (`internal/adapters/git/github_client.go:69`), so a
   passed-through url yields `example.com//tools/x.html` — a `//` and an empty segment
   in the GitHub tree. Every existing path here is repo-relative (`assets/css/styles.css`).

### Blast radius of the change itself, measured before submission

- **471 of 472** live urls resolve byte-identically to what the three rerender call
  sites produce today ⇒ swapping them is inert for every page but one.
- The exception is the fragment row, which those copies today turn into a file literally
  named `tools.html#audience-check.html`. **Not present on the live site (404)** — so
  that copy's defect is latent too.
- **0** pages named `index`/`home` carry a non-`/index.html` url ⇒ dropping that special
  case from the rerender copies is inert.
- **0** urls with a query string, `..`, `//`, whitespace, or a multi-dot final segment.

### The `[UNVERIFIED]` in §4d of the sweep handoff is now resolved

*"16 `git_commit` steps appear to default to `index.html`"* — read the path rather than
acting on it, as instructed. Of the 19 live top-level `git_commit` steps, **none carries
`page_field`**: they use `files_field` (the multi-file map, keyed by the renderer) or
`file_path`/`filename_field`. The three that DO carry `page_field` are
`pageflow-builder`, `page-rebuild` and `site-work-orchestrator`, and their `deploy_page`
steps sit **inside a loop step's sub-steps**, which is why a `jsonb_each` over
`workflow.steps` finds zero of them. So the "16 defaulting to index.html" reading is an
artefact of that query shape, not a real population. Query that finds them:
`WHERE default_config::text LIKE '%page_field%'`.

### Still open after this fix, deliberately

**This stops new orphans; it does not retract existing ones.** Candidate 3 (a
`delete_file` verb on the git adapter) and candidate 4 (a sweep for repo files with no
matching `pages.url`) are untouched and remain the right pairing — `bugs_open/098` needs
the same primitive. Candidate 2 (refuse on disagreement) was deliberately NOT taken: it
would block 316 pages on a disagreement that is expected until each is next deployed.
The disagreement is logged instead, which also makes the fix observable in the pod.

---

# CLOSED 2026-07-31 — fixed, council-APPROVED at round 2, and LIVE on v1.0.1217

**Council `758f6e62-99b8-4f33-a81b-7143351ecd69`: round 1 REVISE → round 2 APPROVED**
("approved with 2 advisory objection(s) — none high-severity"; 12 reviewers, 5 abstained,
0 unreadable). Commits `5dc177f97` (fix), `0a203af54` (round 2), `fb9fcbce5` + `305427012`
(docs). Lane: `docs024_key_docs_latest/bugfix_125_deploy_path_from_url/`.

**Live, verified at the pod and not at the tag** (`bugs_open/153`: a roll is not evidence).
`determinePageFilename` and `PageFilePathFromURL` are unexported, so the discriminating
marker is the log string the change adds, grepped on **every** pod with a negative control
in the same exec:

```
agent-chassis-867fc4f77c-9pcww  v1.0.1217   marker 1   fallback-warn 1   NEG CTL 0
agent-chassis-867fc4f77c-wd2cg  v1.0.1217   marker 1   fallback-warn 1   NEG CTL 0
agent-chassis-744d68cd7-76s55   v1.0.1215   marker 0   fallback-warn 0   NEG CTL 0   ← draining, now gone
```
The old 1215 pod scoring 0/0/0 on the same greps is the control that proves the marker
discriminates. Both 1215 pods are **gone, not draining** — re-checked after the rollout.

## What the round-2 council left as advisory, and what was done about it

- **`prior_art_librarian` (medium): "deferring scope on an unverified absence is the exact
  ASSERTED-ABSENCE shape".** Right, and the check found a near-miss worth recording.
  Verified: the only deletion verb in the git adapter is `delete_repo`
  (`adapter.go:361`, unimplemented at `:641`); no `delete_file`/`unpublish`/`retract`
  action exists anywhere in `platform/`, `internal/` or `cmd/`. **But
  `discovery_checks/check_orphan_pages.go` DOES exist** — and it is not this sweep: it
  starts from `pages` rows and finds pages with no inbound links, so it structurally
  cannot see a **file with no `pages` row at all**, which is what this defect produces.
  The absence claim survives; it is now checked rather than asserted.
- **`prior_art_librarian` (low): round 2's resolution rested on self-reported grep output.**
  Re-attached: `grep -rn 'func getPageInfo\|func loadPageInfo' platform/ internal/ cmd/`
  returns exactly one line — `rerender_single_page_action.go:493 func getPageInfo`.
- **`bug_historian` + `guardian` (medium): the existing orphan population is uncounted, and
  after this ships a fixed page has a correct copy AND a stale one.** True, and it stays
  true — candidate 3/4 are logged on `bugs_open/098` (2026-07-31), not absorbed here.
  **An attempt to bound the population FAILED and is recorded as a failure, not as a zero:**
  `SELECT owner_agent_type, count(*) FROM orchestration_states WHERE owner_agent_type IN
  ('pageflow-builder','page-rebuild','site-work-orchestrator')` → **0 rows across the whole
  18-day retention window (2026-07-13 → 07-31)** — which **contradicts** the known 07-28
  `page-rebuild` run that created the finetuning.uk orphan, so the query is not seeing
  those runs and does **not** bound anything. Anyone doing candidate 4 should start by
  finding out where those runs are actually recorded. **316/472 is the number of pages
  EXPOSED, not the number damaged** — the two have been used interchangeably in this file
  and in the council round, and they are not the same quantity.
- **`guardian` (low): a defect in the shared helper now fans out to 7 pipelines.** Accepted.
  That is the intended trade (one definition instead of five), and it is why the helper
  ships with 60 test cases including a never-empty postcondition.
- **`guardian` (medium): a pod-grep proves the code shipped, not that the first live build
  behaved as measured.** Correct, and the acceptance run is still owed — see below.

## STILL OWED (deliberately, and named rather than quietly dropped)

1. **The `bugs_open/087` acceptance re-test**, on a page that is **not**
   `rebuild_policy=owned`, asserting the **path** and not `success: true`. The sweep
   handoff says 087 and 125 should ship together; 125 is now live, so 087's re-test is
   unblocked and would close both. SQL to pick a target is in the lane RUNBOOK.
2. **Candidates 3 and 4** — the `delete_file` verb and the orphan census. Logged on
   `bugs_open/098`, which needs the same primitive. **Neither bug has ever counted the
   orphan population**, and the obvious query does not do it (above).

**Closing on the "fixed AND live" bar**: the defect — a single-page deploy resolving its
path from the page name — is no longer reproducible on the running binary. What remains
open is the *consequence* of past occurrences (retraction), which is `098`'s primitive and
was never in this bug's cause.
