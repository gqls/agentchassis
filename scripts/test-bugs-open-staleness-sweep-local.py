#!/usr/bin/env python3
"""Local dry-run of bugs-open-staleness-sweep's logic against the real
working tree, with every GitHub API call replaced by an equivalent local
`git` command (ls-tree / show / rev-parse). No network access, no
credentials, no doc_notes write — this only exercises citation extraction
and path resolution against real bugs_open/*.md content.

Run from the repo root:
    python3 scripts/test-bugs-open-staleness-sweep-local.py [ref]
(ref defaults to HEAD)
"""
import importlib.util
import subprocess
import sys

SWEEP_PATH = (
    "deployments/kustomize/services/bugs-open-staleness-sweep/base/sweep.py"
)


def git(repo_root, *args):
    return subprocess.run(
        ["git", "-C", repo_root] + list(args),
        capture_output=True, text=True, check=True,
    ).stdout


def main():
    repo_root = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        capture_output=True, text=True, check=True,
    ).stdout.strip()
    ref = sys.argv[1] if len(sys.argv) > 1 else "HEAD"

    spec = importlib.util.spec_from_file_location("sweep", f"{repo_root}/{SWEEP_PATH}")
    sweep = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(sweep)

    sweep.resolve_ref_sha = lambda r: git(repo_root, "rev-parse", r).strip()
    sweep.fetch_tree = lambda sha: [
        line for line in git(repo_root, "ls-tree", "-r", "--name-only", sha).splitlines() if line
    ]
    sweep.fetch_raw = lambda path, r: git(repo_root, "show", f"{r}:{path}")

    result = sweep.run_sweep(ref)
    print(sweep.render_report(result))


if __name__ == "__main__":
    main()
