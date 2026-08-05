# PLAN — bugfix 120: a merge commit silently skips the deploy for one side

**Lane opened 2026-08-05 ~11:30 BST.** Session "bugfix 201" (redirected: 201 was
picked up by its filing lane minutes before this session started — verified in
that session's live transcript, not inferred from who-owns).

Bug: `bugs_open/120_HANDOFF_2026-07-27_merge_commit_silently_skips_the_deploy_for_one_side.md`

## Validity re-check (2026-08-05, first-hand)

- `gqls/sites@origin/master` (`bbd7703a4`, fetched today): `deploy-to-b2.yml`
  still has `fetch-depth: 2` and `git diff --name-only HEAD~1 HEAD`. **Defect live.**
- `gqls/vm-sites@origin/main` (`bec162b`, fetched today): `deploy-to-vm.yml`
  line 42, same line verbatim. **Defect live.**
- No 090 diagnosis run filed for this: the root cause is a documented git
  semantic (`HEAD~1` on a merge = first parent), self-evidencing and reproducible
  locally on any merge commit; the bug file already carries a fully measured
  production instance (run `30296304596`, curl evidence). This substitutes
  first-hand verification for the loop per the 2026-07-31 owner ruling, stated
  here deliberately.

## Decision — fix candidate 1 (range diff) + candidate 4's spirit (loud coverage), applied to the CLASS

The framework-level defect class is: **"which domains changed" is derived from a
single-commit diff, but a push is a RANGE.** `HEAD~1..HEAD` is wrong for merge
commits (sees only the other side; the pusher always loses) and for multi-commit
pushes (sees only the last commit). The push event carries the true range:
`github.event.before .. github.sha`.

Three files carry the defective line; all are fixed in this lane so they cannot
drift apart:

1. `gqls/sites` `.github/workflows/deploy-to-b2.yml` (branch `master`) — the live
   B2 deploy for ~29 domains.
2. `gqls/vm-sites` `.github/workflows/deploy-to-vm.yml` (branch `main`) — the
   "faithful sibling"; identical "Get changed domains" step, rsync target,
   allowlisted to `relojistas.com` only today.
3. `agentchassis` `docs/agent_docs/docs024_key_docs_latest/034_github_action.md`
   — the transcription sessions read; refreshed to the fixed workflow with a
   dated note. (The `traffic_probe/deploy_setup/working_dir/deploy-to-vm(*).yml`
   copies are point-in-time snapshots of a scratch dir, not a live surface —
   left alone, noted in the bug file at close.)

### The new "Get changed domains" step (both workflows, identical semantics)

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0        # range diff needs the 'before' commit; runner workspace persists, so full history is a one-time cost (pack = 86.67 MiB today)

- name: Get changed domains
  id: changed
  env:
    BEFORE: ${{ github.event.before }}
    AFTER: ${{ github.sha }}
  run: |
    if [ "$BEFORE" = "0000000000000000000000000000000000000000" ] \
       || ! git cat-file -e "$BEFORE^{commit}" 2>/dev/null; then
      echo "No usable 'before' (first push to branch, or force-push discarded it) — falling back to ALL domains"
      CHANGED=""
    else
      echo "Push range: $BEFORE..$AFTER"
      if git rev-parse -q --verify "$AFTER^2" >/dev/null; then
        echo "Tip is a merge commit — the range diff spans BOTH sides (bugs_closed/120)"
      fi
      CHANGED=$(git diff --name-only "$BEFORE" "$AFTER" | grep -E '^[^/]+\.[^/]+/' | cut -d'/' -f1 | sort -u | tr '\n' ' ' || echo "")
    fi
    if [ -z "$CHANGED" ]; then
      CHANGED=$(ls -d */ 2>/dev/null | grep -E '^[^/]+\.[^/]+/$' | tr -d '/' | tr '\n' ' ' || echo "")
      echo "Falling back to all domains"
    fi
    echo "domains=$CHANGED" >> $GITHUB_OUTPUT
    echo "Changed domains: $CHANGED"
```

Sync loops additionally get an `else` branch so a domain in the changed set with
no directory is echoed as a skip instead of vanishing silently (candidate 4's
"convert silent to loud", applied where it is not circular).

### Why this shape

- **Correct for all three push shapes.** Fast-forward: `before..sha` = the one
  commit. Multi-commit: whole range (also broken today, same line). Merge:
  `before` is the remote tip the merge absorbed, so the diff is exactly the
  pusher's un-deployed side — the other side already deployed on its own push.
- **`git diff A B` is a tree comparison** — no ancestry needed — so the range is
  valid even across odd histories, provided both objects exist; the `cat-file`
  guard routes "before does not exist" (first push, force push) into the existing
  deploy-all fallback, which already handles empty-`CHANGED` today. No new
  fallback mechanism, no new failure mode.
- **A hard assertion (candidate 4 verbatim) was considered and dropped as
  circular**: any in-run recomputation of "domains touched in the push" is the
  same range diff. The non-circular residue is loud logging (range, merge-ness,
  fallback reason) and the explicit skip branch — cheap, and it converts every
  future silent state into a visible one.
- **"Rebase before pushing" stays retired** as an operator-memory rule; the
  memory file `merge-commit-skips-the-site-deploy.md` gets corrected at close so
  the workaround folklore does not outlive the defect.

### Blast radius, measured (not asked of the reviewer)

- Cadence: ≥40 deploy runs in the last 24 h on `gqls/sites` (gh run list, capped
  at 40). Every one runs the changed step.
- Strictly additive selection: the new range is a superset of `HEAD~1..HEAD`
  exactly when the old one was dropping domains (merges, multi-commit pushes);
  identical on single-commit fast-forward pushes — the overwhelmingly common
  case (Rerender jobs are single-commit).
- Cost: `fetch-depth: 0` = one-time ~87 MiB history fetch per runner workspace;
  self-hosted runner persists its checkout, so subsequent runs fetch deltas as
  today. Recent runs take 42–54 s; no per-run regression expected after the
  first.
- Consumers told (owner ruling 2026-07-29 §3): every session pushing to
  `gqls/sites` — via the bug-file close, the corrected memory entry, and
  `034_github_action.md`. The guarantee CHANGES for them in one way: after a
  push race, `git pull` (merge) now deploys correctly, so the
  "always `git pull --rebase`" workaround is no longer load-bearing.

## Verification — induce the failing branch (per the bug's own recipe)

On `gqls/sites` (the B2 side, full induction):
1. Clone B (scratchpad): push a probe commit touching `oufe.com/deploy-verify-120-b.txt`.
2. Clone A (`~/projects/sites`): commit probe `webdesign.co.uk/deploy-verify-120-a.txt`
   BEFORE fetching B's push; push → rejected; `git pull --no-rebase` → merge; push.
3. PASS = A's run log names `webdesign.co.uk` in "Changed domains" AND
   `curl https://webdesign.co.uk/deploy-verify-120-a.txt` serves the probe string.
   (Pre-fix, the measured instance shows the pusher's domain absent — that is the
   negative control, already on record as run `30296304596`.)
4. Cleanup: one non-merge push deleting both probes; `b2 sync --delete` removes
   them from the bucket (deletion is not blocked by `--skip-newer`, which gates
   only content updates); curl → 404.

On `gqls/vm-sites`: the fixed step is byte-identical logic; its live blast
surface is one mapped domain (`relojistas.com`) whose webroot is a real VM. The
workflow-fix push itself is `paths-ignore`d (`.github/**`), so the change is
proven by the sites-repo induction plus the next natural push's green run and
its logged "Push range:" line — a full VM induction spends a live-site touch to
re-prove logic already proven an hour earlier on the sibling. If the owner wants
the VM side induced too, the same probe recipe works against
`relojistas.com` with rsync `--delete` cleanup.

Local-preview trap, pre-read (LANDMINES): this machine's `grep` is a ugrep
wrapper whose ERE engine does NOT match `'^[^/]+\.[^/]+/'` the way GNU grep
does — never "test" the pipeline locally with bare `grep`; use `command grep`
or trust the runner's own log.

## Council

Submitted before/alongside the commits (advisory). Note the edited paths live
outside `platform/`/`internal/`/`pkg/` — two of the three files are in sibling
repos and the third is docs — so the relevance-gated seats may mostly not fire,
or the client-side filter may refuse the submission; whichever happens is
recorded in NOTES verbatim.

## Commit / push order

1. agentchassis: lane docs (this file, NOTES, RUNBOOK) — pathspec commit.
2. `gqls/sites`: workflow fix, single-file pathspec commit, push (rebase if the
   push races — ironically this very bug's window; our commit is `paths-ignore`d
   either way so no deploy run fires for it).
3. `gqls/vm-sites`: same, single-file commit to `main`.
4. Induction (above), then `034_github_action.md` + bug-file close + memory
   corrections in agentchassis, `git mv` with BOTH paths named on the commit.
