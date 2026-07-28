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
