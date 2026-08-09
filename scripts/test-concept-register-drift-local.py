#!/usr/bin/env python3
"""Run concept-register-drift-check's REAL logic against this working tree.

No network, no credentials, no cluster. Every GitHub API call is swapped for the
equivalent `git` command against this repo, and `write_doc_note` is replaced by a
print — so the analysis, the parsing and the report rendering under test are the
same functions the CronJob runs, not a copy of them. (A copy is how a harness
comes to pass while production fails: the two drift, and the harness keeps
testing the version nobody deploys.)

USAGE
  ./scripts/test-concept-register-drift-local.py            # HEAD
  ./scripts/test-concept-register-drift-local.py <ref>      # any ref
  ./scripts/test-concept-register-drift-local.py --self-test

THE SELF-TEST IS THE POINT. Running the check against today's tree prints
"clean", and a clean result proves nothing at all — a check that always returns
clean also returns clean. So --self-test runs it against `8f998e86b^`, the commit
immediately BEFORE the 34 missing index rows were backfilled, and requires it to
find exactly those 34. That is the positive control: a run where a broken check
must come out differently from a working one.
"""
import importlib.util
import os
import subprocess
import sys

REPO = subprocess.run(["git", "rev-parse", "--show-toplevel"],
                      capture_output=True, text=True, check=True).stdout.strip()
CHECK_PY = os.path.join(
    REPO, "deployments/kustomize/services/concept-register-drift-check/base/check.py")

# The commit that backfilled the 34 rows; its parent is the last tree that still
# carried the defect. Pinned as a sha, not HEAD~N — a relative ref means
# something different every time anyone commits (see MEMORY: "never cite HEAD~1").
FIX_COMMIT = "8f998e86b"
EXPECTED_MISSING = 34


def load_check():
    spec = importlib.util.spec_from_file_location("check", CHECK_PY)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def git(*args):
    return subprocess.run(["git", "-C", REPO, *args],
                          capture_output=True, text=True, check=True).stdout


def install_git_backend(check):
    """Swap the three GitHub calls for git, leaving the logic untouched."""
    check.resolve_ref_sha = lambda ref: git("rev-parse", ref).strip()

    def list_register_files(sha):
        out = git("ls-tree", "-r", "--name-only", sha, "--", check.REGISTER_DIR + "/")
        return sorted(
            p for p in out.splitlines()
            if p.endswith(".md") and os.path.dirname(p) == check.REGISTER_DIR
        )

    check.list_register_files = list_register_files
    check.fetch_raw = lambda path, ref: git("show", f"{ref}:{path}")
    return check


def run(ref):
    check = install_git_backend(load_check())
    sha, result = check.run_check(ref)
    return check, sha, result


def main():
    args = sys.argv[1:]
    if args and args[0] == "--self-test":
        failures = []

        # 1. POSITIVE CONTROL — the tree that carried the defect.
        _, sha, before = run(f"{FIX_COMMIT}^")
        missing = before["findings"]["entry_without_row"]
        print(f"[control] {FIX_COMMIT}^ ({sha[:12]}): "
              f"{before['entries']} entries, {before['rows']} rows, "
              f"{len(missing)} entry-without-row")
        if len(missing) != EXPECTED_MISSING:
            failures.append(
                f"positive control: expected {EXPECTED_MISSING} missing rows at "
                f"{FIX_COMMIT}^, found {len(missing)}")
        if "CLM-001" not in missing:
            failures.append("positive control: CLM-001 was one of the 34 and is absent")
        # The headline was HONEST at that ref — and that is the point of this
        # assertion, not a technicality. It said 1,721 and there were exactly
        # 1,721 rows; the 34 missing concepts were invisible to it, because a
        # row count cannot see a row nobody wrote. An earlier version of this
        # harness asserted headline_drift here, on the assumption that a broken
        # register must look broken from every angle. It does not: the check
        # that agreed with itself is precisely the one that let this run for
        # three weeks.
        if before["headline"] != before["rows"]:
            failures.append(
                f"positive control: the headline ({before['headline']}) is expected to "
                f"AGREE with the row count ({before['rows']}) at that ref — the whole "
                f"point is that agreement proved nothing while 34 entries were missing")
        if before["headline"] == before["entries"]:
            failures.append(
                "positive control: the headline must NOT equal the entry count at that "
                "ref, or there was no gap to find")

        # 2. THE FIX — same check, one commit later.
        _, sha2, after = run(FIX_COMMIT)
        total = sum(len(v) for v in after["findings"].values())
        print(f"[fixed]   {FIX_COMMIT} ({sha2[:12]}): "
              f"{after['entries']} entries, {after['rows']} rows, {total} finding(s)")
        if after["findings"]["entry_without_row"]:
            failures.append("fixed ref still reports entry-without-row")

        # 3. TODAY.
        _, sha3, now = run("HEAD")
        total_now = sum(len(v) for v in now["findings"].values())
        print(f"[HEAD]    {sha3[:12]}: {now['entries']} entries, "
              f"{now['rows']} rows, {total_now} finding(s), "
              f"stored counts={len(now['findings'].get('stored_count_returned') or [])}")

        # 4. MUTATION — prove the comparison, not the plumbing. Delete one row
        #    from the index text and require the check to name that id. Without
        #    this, an `analyse` that returned empty sets would pass everything
        #    above except the control, and the control could be satisfied by any
        #    function that happens to return 34 things.
        check = install_git_backend(load_check())
        files = check.list_register_files(git("rev-parse", "HEAD").strip())
        texts = {p: check.fetch_raw(p, "HEAD") for p in files}
        idx = f"{check.REGISTER_DIR}/{check.INDEX_NAME}"
        victim = "ADP-018"
        texts[idx] = "\n".join(
            ln for ln in texts[idx].splitlines() if not ln.startswith(f"| {victim} |"))
        mutated = check.analyse(files, texts)
        print(f"[mutation] removed the {victim} index row: "
              f"entry_without_row={mutated['findings']['entry_without_row']}")
        if mutated["findings"]["entry_without_row"] != [victim]:
            failures.append(
                f"mutation: deleting {victim}'s row must report exactly [{victim}], "
                f"got {mutated['findings']['entry_without_row']}")

        # 5. SECOND MUTATION — the retirement arm, which replaced the old
        #    headline comparison on 2026-08-09. With stored counts gone, the old
        #    check would simply find no headline and report nothing, which is
        #    indistinguishable from passing. So the arm now looks for a count
        #    that has COME BACK, and this proves it fires: re-add one.
        texts2 = {p: check.fetch_raw(p, "HEAD") for p in files}
        target = f"{check.REGISTER_DIR}/adapters.md"
        texts2[target] = "**9,999 concepts** in this file.\n" + texts2[target]
        mutated2 = check.analyse(files, texts2)
        got = mutated2["findings"].get("stored_count_returned") or []
        print(f"[mutation] re-added a stored count to adapters.md: "
              f"stored_count_returned={[g[0] for g in got]}")
        if [g[0] for g in got] != ["adapters.md"]:
            failures.append(
                f"mutation: a re-added stored count must be reported for exactly "
                f"['adapters.md'], got {[g[0] for g in got]}")
        elif got[0][1] != 9999:
            failures.append(
                f"mutation: the reported stated count should be 9999, got {got[0][1]}")

        # 6. AND THE INVERSE, which is the whole reason the arm was inverted:
        #    the CURRENT tree must report NO stored counts. If this passes while
        #    (5) fails, the arm is dead rather than satisfied.
        if now["findings"].get("stored_count_returned"):
            failures.append(
                f"HEAD still carries stored counts: "
                f"{[g[0] for g in now['findings']['stored_count_returned']][:5]}")

        print()
        if failures:
            for f in failures:
                print(f"FAIL: {f}")
            return 1
        print("SELF-TEST PASSED — the check finds the real historical defect, "
              "reports clean once it is fixed, and names the id when a row is "
              "removed under it.")
        return 0

    ref = args[0] if args else "HEAD"
    check, sha, result = run(ref)
    print(check.render_report(result, ref, sha))
    print()
    print(f"(local harness — no doc_notes row written; ref {ref} = {sha[:12]})")
    return 0


if __name__ == "__main__":
    sys.exit(main())
