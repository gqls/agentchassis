# PLAN 2026-08-02 — the unpublish primitive (`bugs_open/098`)

**Bug:** `bugs_open/098` — archiving a page removes it from every derivation but not
from the deployed site, so its frozen listing keeps advertising a 404.
**Second caller:** `bugs_closed/125`, whose council round asked that "the fix stops new
orphans and retracts none" be logged as a required follow-up rather than absorbed. It was
logged on 098. This is that follow-up.

**The one-line statement of the gap, taken from 125's filing and verified here:**
*the platform can publish a page but has no implemented way to unpublish one.*

---

## What was verified first-hand before any code was written

Not carried forward from the bug file — each re-run 2026-08-02.

| claim | check | result |
|---|---|---|
| the population is real | `SELECT count(*) FROM pages WHERE status='archived' AND deployed_at IS NOT NULL` | **13** (leopardess 10, robot-hands 2, relojistas 1) — unchanged since the 07-27 re-measure |
| the live instance still reproduces | `curl -o /dev/null -w '%{http_code}' https://robot-hands.com/learning-center/index.html` | **200**; its listed target `/blog/learning-center-article.html` **404** |
| the artefact is really in the deploy repo | `gh api repos/gqls/sites/contents/robot-hands.com/learning-center/index.html?ref=master` | present, 41,785 bytes, sha `8394345f…` |
| the adapter has no deletion verb | `internal/adapters/git/adapter.go:357-366` | 5 verbs: commit, create_repo, delete_repo, create_branch, create_pull_request. `delete_repo` returns *"not yet implemented"* (`:641`) |
| nothing in the codebase archives a page | grep `archived` across Go + all four frontends for a write | **no hits** — archiving is a hand-run SQL operation, which is *why* there is no retraction hook to add one to |

### The finding that decides the whole design

**The B2 sync is already reconciling.** `gqls/sites/.github/workflows/deploy-to-b2.yml`
runs `b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"` on every
push to `master`, then purges the Cloudflare zone for that domain.

So a file removed from the `sites` repo **is** removed from B2 and **is** purged from the
edge. Bug 098's fix candidate 1 ("make the deploy reconciling rather than additive")
turns out to be half-built already: the *sync* reconciles; it is the *git tree* that
nothing ever removes from. That reduces the whole bug to one missing capability rather
than a deploy-pipeline redesign, and it means the primitive below is sufficient — no
B2-side work, no cache-invalidation work.

`--skip-newer` governs *overwrites*, not deletions, so it does not blunt `--delete`.
The changed-domain detector is `git diff --name-only HEAD~1 HEAD`, which lists deletions,
so a deletion-only commit still selects its domain.

---

## Design — deletion is a KIND OF COMMIT, not a second write path

The whole of the fix hangs on one decision, so it is stated first.

GitHub's Git Data API expresses a deletion as a tree entry whose `sha` is `null`. That
means a deletion can travel through **`CommitToRepo` unchanged** instead of through a
bespoke `DELETE /repos/{o}/{r}/contents/{path}` call. Reusing the commit path buys, for
free and by construction rather than by remembering:

- **the ref-race retry loop** (`github_client.go:147-184`, the `bugs_open/120` fix). A
  retraction racing a concurrent deploy of the same repo re-bases exactly as an addition
  does. A bespoke Contents-API delete would be a *second* writer to the same branch with
  none of that, which is the drift class this platform reviews for;
- **the `{domain}/{path}` prefixing convention** — one implementation, so a deletion can
  never target a path a publish would not have written;
- **atomicity with additions.** A move (write the new path, delete the old) becomes ONE
  commit, so a page is never absent from both paths. `bugs_closed/125`'s orphan class is
  exactly a half-move; this makes the whole move expressible;
- **the credential boundary.** The write token still never leaves the adapter.

The alternative — a `deleteFile` method calling the Contents API — is smaller to write and
worse in every one of those four ways. It is not taken.

## The four pieces

**1 — `TreeEntry.SHA` becomes `*string`.** `nil` marshals to JSON `null`, which is the
API's deletion signal. An empty Go `string` marshals to `""`, which GitHub rejects — so
the pointer is load-bearing, not stylistic.

**2 — `GitCommitData.Deletions []string`**, domain-prefixed identically to `Files`, turned
into null-SHA tree entries. `CommitToRepo`'s guard relaxes from `len(Files) == 0` to
`len(Files)+len(Deletions) == 0`, so a deletion-only commit is legal and a
nothing-at-all commit still is not.

**3 — `delete_file` verb** on the adapter, plus the generic chassis caller's allowlist.

**4 — `retract_page_deployment` chassis action** — the page-level caller, so the primitive
is *reachable by the platform* and not just by a hand-written payload.

## Idempotency, and why existence is checked per path

A retraction must be safely re-runnable. **This turned out to be required rather than
defensive, and it was probed rather than assumed** — POST `/git/trees` against
gqls/sites' live master tree, 2026-08-02 (a tree object is created unreferenced, so the
probe moves no ref and fires no workflow):

| request | result |
|---|---|
| null sha, path **present** | 201, new tree sha `0d8dab50…` |
| null sha, path **absent** | **422 `GitRPC::BadObjectState`** |

So without the filter, the second run of a repair fails. The client filters requested paths
against what is at the branch head and treats absent as a success — the state the caller
asked for.

**Existence is checked with the Contents API, one call per path, not with a recursive tree
listing.** A recursive listing carries a `truncated` flag and a truncated listing reports
present files as absent, which would make a real retraction skip silently and report
success.

> **CORRECTED 2026-08-02, before submission — I wrote this justification as though
> truncation were a present hazard, then measured it and it is not.** The sites repo is
> **1,847 entries, `truncated: false`** — nowhere near GitHub's limit. So truncation is
> *headroom*, not the reason. The actual reason a per-path check is right is that it is
> what makes the present/absent split reportable at all. Logged in `WRONG_CALLS.md`; the
> code comment says the measured thing now.

## Guards on the page-level action

It deletes files, so the guards are the design, not decoration:

1. **Paths are only ever derived from `pages.url` through the shared
   `datahelpers.PageFilePathFromURL`** — the same function the publish uses since
   `bugs_closed/125` (PBP-0xx). So a retraction can only ever name a file a publish could
   have written: never an asset, never a stylesheet, never `.github/`.
2. **The action re-reads the page rows itself and never trusts a caller-supplied path.**
   A caller names pages; it does not name files.
3. **It refuses a page whose derived file path is also derived by an `active` page on the
   same site.** Two urls can collide onto one file (`/foo/` and `/foo/index.html`), and
   retracting the archived one would delete the live one's artefact.
   **Measured 2026-08-02: 0 such collisions fleet-wide** — so this guard is inert today.
   It ships anyway because it is the difference between "no collision exists" and "a
   collision cannot cause damage", and RFC_010 is open on exactly this class.
4. **Only pages the platform says must not be served are eligible** — `status <> 'active'`.
   Never "delete every file with no matching page row": the repo legitimately holds files
   the `pages` table does not model.

## What this deliberately does NOT do, and the measurement behind it

**`deployed_at` is NOT cleared on retraction.** 098's candidate 2 asks for it, and it is
the candidate that would protect the shared predicate family from a future consumer who
reaches for `deployed_at IS NOT NULL` alone.

It is not done here because **the census contradicts the premise that the column is only
ever read as a boolean.** 49 non-test references; three read it as *history*:

- `reconcile_superseded_reviews_action.go:98` — `p.deployed_at > GREATEST(wi.created_at, …)`,
  i.e. "was the page redeployed after this review was raised?";
- `maintenance_actions.go:725` — `findStalePages`, `deployed_at < NOW() - interval`;
- `page_admin_handlers.go:101` — displayed to a human.

Clearing the stamp changes what all three return. That is a change to what a shared column
*guarantees*, which by the owner ruling of 2026-07-29 §1 is architecture-scope — and
bundling it into a bug patch is precisely the ground `bugs_closed/124` drew a REJECTED
verdict on. It is recorded on the bug as measured-and-deferred, with the census, so the
next thread starts from the numbers rather than from the recommendation.

## Acceptance

The bug names its own failing branch, and it is directly reproducible:

```sh
curl -sS -o /dev/null -w '%{http_code}\n' https://robot-hands.com/learning-center/index.html
# before: 200.  after a real retraction: 404
```

A green run on a site with no archived pages proves deployment, not correctness — 098 says
so explicitly, and the acceptance is the live instance or nothing.

## Landmines found while doing this (all appended to LANDMINES.md)

- **`sites.github_branch` says `main`; the repo has no `main`.** `gqls/sites` carries
  `master` (default) and `750start`. The B2 workflow triggers on `master`. Deploys work
  only because `GitCommitAction` never passes a branch and `CommitToRepo` falls back to
  the repo default. **A retraction that helpfully passes `sites.github_branch` commits to
  a branch that does not exist and never deploys.**
- **a recursive tree listing can be `truncated`** — see above.
