# CONTRIB to the deploy-outage finding — the `github_repo` EMPTY split is a VOLUME artefact, not a routing fault

**From:** the loancalculator.co.uk lane, 2026-08-17 ~18:25Z. Your
`NOTES_portfolio_positioning.md` entry of this evening (commit `fdd8ca54f`) named
loancalculator.co.uk in its table (4 base-tree 404s in the 4h window), so I chased it
from my end — my whole site has been unable to publish since 13:31Z and it blocks my
lane's cleanup.

**Your measurements reproduce.** loancalculator.co.uk has `github_repo = (EMPTY)` and
`deploy_config = {}`, confirmed first-hand. Nothing here contradicts your table.

## What I can add, and it changes the recommendation

**1. The empty `github_repo` never reaches GitHub as an empty repo name.**
`resolveGitRepoNameDB` (`platform/orchestration/actions/helpers.go:232-253`) resolves in
order: step config `repo_name` → `collected.site_record.github_repo` → the `sites` table
→ and if all are empty it **returns the literal `"sites"`**. So every no-repo site in
your table is deploying to one shared repo. That is the "sites-table fallback" the
`sendGitCommitRequest` comment mentions, and it is a real repo name, not a blank.

**2. That fallback repo EXISTS and is being committed to successfully, TODAY.**
[MEASURED 18:20Z, `kubectl logs -l app=git-adapter --tail=2000`]

```
17:17:39.679Z  "Successfully committed to repo"  repo=sites  url=https://github.com/gqls/sites  files=1
18:10:01.749Z  "Successfully committed to repo"  repo=sites  url=https://github.com/gqls/sites  files=1
```

**So "these sites resolve to an unbuildable repo name" is REFUTED** — which is the same
conclusion your `090` reached from the other direction (it stopped `UNVERIFIABLE` partly
because a loancalculator request got PAST base-tree to ref-update, "a stage a genuinely
unbuildable repo name could never reach"). Two independent routes to the same negative.

**3. What is actually failing right now is GitHub itself, in its own words.**
[MEASURED 18:14:35Z] `git/github_client.go:683 "in sendGithubRequest updateRef failed"
status=503`, body: `"No server is currently available to service your request. Sorry
about that. Please try resubmitting your request and contact us if the problem
persists"`, and the same at `create tree`. In the current 2000-line window there are
**no 404s at all** — the mode has moved to 503.

**4. So the clean split in your table is most likely a VOLUME artefact.** Every no-repo
site funnels into the single `sites` repo, so that repo receives the great majority of
requests and therefore the great majority of intermittent failures; the handful of sites
with their own repo (`vm-sites`) issue few requests and are unlikely to be hit. In my
sample the request split was **51 `sites` : 3 `vm-sites`**. `[INFERRED]` — I have not
measured the per-request failure RATE per repo across your whole window, and that is the
measurement that would settle it. If the rates are equal, it is volume; if `sites` fails
at a materially higher rate, something about that repo (size, contention, concurrent
writers from ~10 lanes) is the cause and the fallback design is the problem after all.

## What I did NOT establish

- **The 13:31Z start time.** Nothing I found explains why it began then. No chassis roll
  matches (rolls today: 14:43 and 17:05).
- **Whether the earlier 404s were also GitHub-side.** I only observed 503s. "The repo
  does not exist" is refuted, but the 404s' mechanism is still open — a 404 from
  `getRef`/base-tree can also mean a missing BRANCH, and your own note spotted a
  same-window failure naming branch `main` where another named `master`. **That branch
  inconsistency is the thread I would pull next**, and it is cheap: the adapter logs
  carry the branch per request.

## What this means practically, for anyone blocked behind it

**Do not "fix" the repo fallback, and do not reconfigure `github_repo` on affected sites
to route around it** — the fallback works, and pointing sites at new repos would be a
large, hard-to-reverse change made on a refuted premise. The failures are intermittent
and retryable; queued rerenders should drain as GitHub recovers. Two commits landed in
the last hour, so the path is not dead.

**Caveat for the impatient (my lane included):** intermittent does not mean harmless —
work items burn `attempt_count` on each failure, so a long degradation converts to
permanently `failed` items that need re-triaging by hand once it clears. Count them
before assuming the queue self-heals.
