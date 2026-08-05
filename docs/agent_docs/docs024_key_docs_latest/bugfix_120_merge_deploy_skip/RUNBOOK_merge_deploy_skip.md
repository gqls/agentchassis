# RUNBOOK — bugfix 120

## Read the live workflows (never trust a local clone without fetching)

```bash
git -C ~/projects/sites fetch origin master && git -C ~/projects/sites show origin/master:.github/workflows/deploy-to-b2.yml
git -C ~/projects/vm-sites fetch origin && git -C ~/projects/vm-sites show origin/main:.github/workflows/deploy-to-vm.yml
# gotcha: gqls/sites default branch is MASTER (no main); vm-sites is MAIN.
```

## Watch a deploy run

```bash
gh run list -R gqls/sites -L 5            # Deploy to B2, branch master
gh run view <run-id> -R gqls/sites --log | grep -A3 "Changed domains"
# gotcha: do NOT preview the changed-domain grep locally — this machine's grep
# is a ugrep wrapper whose ERE engine does not match '^[^/]+\.[^/]+/' (LANDMINES).
```

## Induce the merge-race (verification for this bug)

```bash
# clone B in scratchpad
git clone git@github.com:gqls/sites.git "$SCRATCH/sites-b" --depth 50
# A: commit probe WITHOUT pushing; B: push its own probe first; A: push (rejected),
# git pull --no-rebase (merge), push. PASS = A's run names A's domain and
# curl serves A's probe file. Cleanup: one non-merge push deleting both probes.
# gotcha: the racing commits must each touch a DOMAIN dir — a root-level file
# produces empty CHANGED and triggers the deploy-ALL fallback, wrecking the test.
# gotcha 2: probe cleanup relies on deletions, which b2 sync --delete performs
# regardless of --skip-newer; content REVERTS are the thing --skip-newer eats.
```
