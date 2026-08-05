# 120 — a merge commit silently skips the deploy for one side of the merge

**Filed 2026-07-27** from `webdesign.co.uk`, caught by accident. ~~**Unowned.**~~
**OWNED 2026-08-05 ~11:30 BST by the `bugfix_120_merge_deploy_skip` lane** (session
"bugfix 201", redirected here — 201 was taken by its filing lane minutes before).
Re-verified today against `gqls/sites@origin/master` (`bbd7703a4`): the workflow
still ships `fetch-depth: 2` + `git diff HEAD~1 HEAD` — the defect is live. Taking
fix candidate 1 (+4 as complement), across all three files named in Related.
Affects the `gqls/sites` deploy for **every domain in the fleet**, and the same
line is present in the VM deploy workflow, so it is not B2-specific.

---

## Symptom

A push lands on `gqls/sites@master`, the GitHub Action reports **success**, and
the files are **never copied to the bucket**. The site keeps serving the previous
version. Nothing anywhere reports a failure — the run is green, the commit is in
`master`, and `git log` shows exactly what you expect.

Measured instance, 2026-07-27:

- `7237eb851` added the related-link block to **63** `webdesign.co.uk` tool pages.
- Push rejected (another session had pushed); `git pull --no-rebase` produced merge
  `55f136229`; push succeeded.
- Run `30296304596` — **success**, 26 s.
- Live check afterwards: `curl https://webdesign.co.uk/tools/touch-target/index.html`
  returned `last-modified: Sat, 25 Jul 2026 20:40:31 GMT` and **zero** occurrences
  of the new block. Cache-busted with a query string; still absent. Not a cache.

## Root cause

`.github/workflows/deploy-to-b2.yml`, "Get changed domains":

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 2
...
CHANGED=$(git diff --name-only HEAD~1 HEAD | grep -E '^[^/]+\.[^/]+/' | cut -d'/' -f1 | sort -u ...)
```

**On a merge commit, `HEAD~1` is the FIRST parent** — the branch you were on, i.e.
*your own commit*. So `git diff HEAD~1 HEAD` returns **only the other side of the
merge**. Your own changed files are, by construction, absent from that diff.

In the instance above the diff yielded `fundamentallyai.com oufe.com
robot-hands.com` (the other session's work, which deployed correctly) and
`webdesign.co.uk` never entered `$CHANGED`, so the `for domain in ...` loop never
synced it and the Cloudflare purge never named it.

**Both sides never lose — it is always the pusher who loses**, because the pusher's
commit is what becomes parent 1. So the session that did the work is the one whose
work silently does not ship, which is the worst possible assignment of the failure.

### Why it stays invisible

- The run is **green**. Nothing in it asserts that the domain you touched was synced.
- `paths-ignore` and the `grep -E '^[^/]+\.[^/]+/'` filter mean an empty result is a
  normal, expected state (a docs-only push), so "no domains changed" cannot be
  distinguished from "your domain was dropped".
- The empty-`CHANGED` fallback (`ls -d */`, deploy everything) **does not** rescue
  this case: `CHANGED` is not empty, it is merely missing your domain.
- `b2 sync --skip-newer` means even a later unrelated sync will not repair the file
  if the bucket copy has a newer mtime.

### Frequency is structural, not rare

`gqls/sites` is pushed by many concurrent sessions and by automated `Rerender:`
jobs — 4 pushes landed in the ~10 minutes around this one. A push race is therefore
the **normal** case, and `git pull` (merge) is the documented, obvious response to a
rejected push. Every one of those produces a merge commit and drops a domain.

## Fix candidates — ordered by what closes the door

1. **Use the push event's own commit range, which spans BOTH sides.** GitHub
   supplies it, so no git reasoning is needed:
   ```yaml
   - uses: actions/checkout@v4
     with: { fetch-depth: 0 }        # a range diff needs more than depth 2
   ...
   CHANGED=$(git diff --name-only ${{ github.event.before }} ${{ github.sha }} | ...)
   ```
   This makes the bad state unrepresentable for a push of any shape — merge,
   fast-forward, or several commits at once (which `HEAD~1` **also** mishandles
   today: a 3-commit push only ever examines the last one). Guard
   `github.event.before` being all-zeros on a first push by falling back to
   deploy-all.
2. **Diff against the merge base when the tip has two parents** —
   `git diff --name-only $(git merge-base HEAD^1 HEAD^2) HEAD`. Correct for merges,
   still wrong for multi-commit pushes.
3. **Deploy every domain when `git rev-parse HEAD^2` succeeds.** Crude, safe,
   one line; costs a full-fleet sync on every merge.
4. **Assert instead of infer** — after the sync loop, fail the run if any domain
   directory touched in the push is absent from `$CHANGED`. Does not prevent the
   miss, but converts a silent skip into a red run.

Candidate 1 is the one to take; 4 is a cheap complement. **"Rebase instead of
merging before you push" is not a fix** — it is an operator-memory rule standing in
for a defect, and it will be forgotten by the next session under push-race pressure.

## Workaround until it is fixed

Any commit that touches a file under `<domain>/` causes `b2 sync` to ship **the
whole domain directory**, not just the changed file. So a subsequent non-merge push
touching that domain carries the earlier orphaned files with it. That is how the 63
pages above eventually reached the bucket (`df51bfd91`, run `30296511843`) — and
note it required a **rebase** so that the tip commit's own diff named the domain;
a second merge would have failed identically.

## How to verify a fix

Induce it, do not infer it:

1. On a clean `master`, commit a change to `<domain>/` **without** pushing.
2. From a second clone, push an unrelated commit so your push is rejected.
3. `git pull --no-rebase`, push the resulting merge.
4. **Before**: the run is green and `curl` shows the old `last-modified` for your
   file. **After**: the run names `<domain>` in "Changed domains" and `curl` shows
   the new content.

Do **not** verify by pushing a non-merge commit — that is the path that already
works, and it will pass whether or not the fix is present.

## Related

- Same `git diff HEAD~1 HEAD` line appears in the VM deploy workflow
  (`docs/agent_docs/docs024_key_docs_latest/traffic_probe/deploy_setup/working_dir/deploy-to-vm(4).yml:35`)
  and is transcribed in `docs024_key_docs_latest/034_github_action.md:25`. Fix all
  three together or they drift.
- Family: `bugs_open/098` (archiving does not undeploy) and the standing note that
  **`deployed_at IS NOT NULL` means a deploy happened once, not that the page is
  fetchable**. This is the same shape one layer lower — the platform's model says
  shipped, the bucket disagrees, and nothing compares them.
- `bugs_open/116` (the link checks have never run) is why this class survives: there
  is no post-deploy assertion that what is live matches what was built.

---

### CONTRIBUTED 2026-07-28 ~11:30 (work-item parallelisation thread) — a SECOND concurrent-writer mechanism on gqls/sites, caught in a controlled burst

Five simultaneous `page-rerender` runs (vonc.com, distinct pages) each ended
in a `git_commit` to gqls/sites. Four deployed; the fifth got a prompt,
explicit failure from the git-adapter:

    failed to update ref for branch "master": github API request failed with
    status: 422 Unprocessable Entity - {"message":"Update is not a fast forward"}

So this file's mechanism (a HUMAN-side `git pull` merge breaking the deploy
diff) has a machine-side sibling: the git-adapter's own concurrent commits
race read-base→update-ref at the GitHub API, and the loser is refused. Today
the chassis processed work strictly serially, so the race was nearly
unreachable; the worker pool (chassis_replica_scaling CS-2, live 2026-07-28)
makes concurrent same-repo commits ROUTINE. The class is now: gqls/sites has
multiple concurrent writers and no serialisation anywhere — one fix shape
(per-repo commit serialisation or 422-retry-with-rebase in the git-adapter)
would close both the machine race and narrow this file's window. Evidence
corr: `6aedced7-490d-466a-ba5e-163616bdce45`, 11:26:28Z. Not filed as a new
number — same repo, same missing serialisation, one account.

---

### CLOSED 2026-08-05 — fixed AND live on both deploy repos, proven by induced merge race (bugfix_120_merge_deploy_skip lane)

**Fix shipped** (candidate 1 + candidate 4's non-circular residue), council
**APPROVED round 1** (`Council-Reviewed: f029e06f-cfe9-4e01-84fd-1940349fc6d5`,
2 medium advisories, both acted on — below):

- `gqls/sites@efe046c6d` (master): `deploy-to-b2.yml` now diffs
  `github.event.before..github.sha` with `fetch-depth: 0`; unusable `before`
  (40-zero first push, or force-push that discarded it — literal byte-verified
  at exactly 40 zeros, per the editquality advisory) falls into the
  pre-existing deploy-all fallback; the range, merge-ness, fallback reason and
  every skip are logged.
- `gqls/vm-sites@30e447a92` (main): same step, byte-identical `run:` body.
- `agentchassis`: `034_github_action.md` and `scripts/github/deploy_to_b2_action.yaml`
  transcriptions refreshed with authority pointers.

**Verification — induced, both controls, artefact-level:**

- Probe A committed unpushed on `webdesign.co.uk`; probe B pushed first from a
  second clone (`8d1ae74a6`); A's push rejected; `git pull --no-rebase` → merge
  `750e8ecd4` (first parent A `3c68252dd`, absorbed remote tip B); pushed.
- Run `30998151789` (the merge push) logged `Push range: 8d1ae74a6..750e8ecd4`,
  `Tip is a merge commit…`, **`Changed domains: webdesign.co.uk`** — the
  pusher's side, which pre-fix `HEAD~1` structurally dropped (parent-1 diff
  would have yielded only `testllmlog.example.com`).
- Artefact: `curl https://webdesign.co.uk/deploy-verify-120-a.txt` served the
  probe token; after the cleanup push, HTTP 404. A file that never existed
  before has no stale-baseline ambiguity (the debug_historian advisory).
- Control run `30998135192` (B's fast-forward push): range diff, no merge note,
  `Changed domains: testllmlog.example.com` — identical semantics to the old
  behaviour on the common path.
- `gqls/vm-sites` run `30998426033`: `Push range: ce3acff..651811a`,
  `Changed domains: idea.uk`, then the allowlist's explicit skip (no mapped
  host) — the fixed step proven on the sibling with zero live-VM impact.

**Third-copy census** (the bug_historian advisory — run, not asserted):
`HEAD~1` changed-set derivations fleet-wide are: the two live workflows (fixed);
`~/projects/sites2` + `~/projects/sites_orig` (stale local clones of gqls/sites
— their checkouts update on pull, not independent surfaces);
`traffic_probe/deploy_setup/**` copies (dated snapshots of a scratch dir, left
as history); `scripts/github/deploy_to_b2_action.yaml` (maintained reference —
refreshed in this close); `scripts/test-concept-register-drift-local.py`
(a comment warning AGAINST relative refs — not a selector). **No third live copy.**

**Residual risks, distinct from this bug, named for the record:**

- `b2 sync --skip-newer` can eat a content REVERT (guardian advisory; already a
  LANDMINE keyed to this file). Not implicated here: the losing run never wrote
  the pusher's files, so no newer bucket copy exists on the repaired path — and
  the induction proved live propagation end-to-end.
- A domain DIRECTORY deleted from the repo is skipped by the sync loop (now
  loudly), so whole-domain retraction still never propagates — `bugs_open/098`
  family, not this mechanism.
- The two live workflows carry the fix as byte-identical copies rather than a
  shared composite action (reuse_agent advisory) — a future factoring if the
  sibling count grows.

Transferable pattern added to 016b §9 ("A CI selector that diffs ONE COMMIT…").
Lane docs: `docs024_key_docs_latest/bugfix_120_merge_deploy_skip/`.

---

**UPDATE same day ~14:45 — the machine half is FIXED and LIVE (owner-ruled).**
`CommitToRepo` now re-bases and retries at GitHub's ref CAS on a
non-fast-forward (commit `7dc876795`, adapter v1.0.1187 on both replicas,
verified via `/proc/1/exe` — `/root/git-adapter` is unreadable by the
container user). Proof: 16/16 concurrent deploys completed where the morning
burst went 4/5; the failing branch is pinned by an induced fake-GitHub test
(`1602dcd95` — verbatim 422, head moves mid-flight, second attempt builds on
the winner's base, blobs created once). **This also absorbs this file's human
half in the window sense**: a human push landing between the adapter's read
and update now causes a re-base, not a failed deploy. What it does NOT fix is
this file's original mechanism — the deploy WORKFLOW diffing `HEAD~1` on a
merge commit — which remains open. Council review of the fix died on the
fleet's LLM spend cap (`complete_invalid` at seat 1, corr `bf2bef0a…`;
access returns 08-01) — resubmit after that date if the trailer is wanted.
— work-item parallelisation thread
