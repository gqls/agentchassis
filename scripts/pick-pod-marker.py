#!/usr/bin/env python3
"""pick-pod-marker — choose pod-grep markers a binary can actually contain.

Why this exists (2026-08-03): a lane probed a deployed chassis for a phrase that
lives only in a Go `//` comment — 0 matches against every binary ever built —
hours AFTER the landmine describing exactly that trap was written. Prose was not
enough, so this is the mechanical version. The source diff only NOMINATES
candidates; the verdict comes from a binary built from `git archive <commit>`,
which no comment can pass and no working-tree WIP can pollute.

    scripts/pick-pod-marker.py <commit> [--package ./cmd/agent-chassis]

Output:
  * ADDED literals that ARE in the built binary  → positive markers, ranked;
  * ADDED literals that are NOT in the binary    → the trap, shown explicitly
    (comment-only, test-only, unused const, or build-tag-excluded);
  * a REMOVED literal absent from the binary     → the preferred negative
    control ("a string it REMOVED, expect 0" — proves the roll, not just the
    pipeline), plus a synthetic control that exists nowhere;
  * ready-to-paste kubectl greps for EVERY replica, per the
    roll-is-not-evidence rule.

Exit 0: at least one verified positive marker. Exit 1: none survived the
binary. Exit 2: could not resolve the commit or build the probe binary.
"""

import argparse
import os
import re
import shlex
import subprocess
import sys
import tempfile

NS = "ai-persona-system"

# A line's worth of Go string literals. Backslash-bearing literals are dropped
# rather than unescaped: decoding Go escapes here would be a second compiler to
# keep correct, and clean candidates are never in short supply.
DQ = re.compile(r'"((?:[^"\\])+)"')
RAW = re.compile(r'`([^`]+)`')


def run(argv, **kw):
    return subprocess.run(argv, capture_output=True, text=True, **kw)


def literals(line):
    for pat in (DQ, RAW):
        for m in pat.finditer(line):
            yield m.group(1)


def harvest(commit):
    """(added, removed) literal lists from the commit's non-test .go lines."""
    out = run(["git", "show", "--format=", "--unified=0", commit, "--", "*.go"])
    if out.returncode != 0:
        sys.exit("git show failed: %s" % out.stderr.strip())
    added, removed, path = [], [], None
    for line in out.stdout.splitlines():
        if line.startswith("+++ b/"):
            path = line[6:]
            continue
        if path is None or path.endswith("_test.go"):
            continue
        if line.startswith("+") and not line.startswith("+++"):
            body, bucket = line[1:], added
        elif line.startswith("-") and not line.startswith("---"):
            body, bucket = line[1:], removed
        else:
            continue
        comment = body.lstrip().startswith("//")
        for lit in literals(body):
            # Printable ASCII only: the probe pipeline is `strings | grep`, and
            # `strings` emits ASCII runs — a multibyte char (…, £, —) SPLITS the
            # line, so a marker spanning one matches the binary's bytes and still
            # greps 0 in the pod. Found live: this script's own first suggested
            # marker contained U+2026 and probed 0/0 against a binary that had it.
            if len(lit) >= 8 and lit.strip() and lit.isascii() and lit.isprintable():
                bucket.append({"lit": lit, "path": path, "comment": comment})
    return added, removed


def build_probe(commit, package, scratch):
    """go build <package> from a clean archive of <commit>; return binary bytes."""
    tree = os.path.join(scratch, "tree")
    os.makedirs(tree)
    ar = subprocess.Popen(["git", "archive", commit], stdout=subprocess.PIPE)
    tar = run(["tar", "-x", "-C", tree], stdin=ar.stdout)
    ar.wait()
    if ar.returncode != 0 or tar.returncode != 0:
        sys.exit("git archive %s failed" % commit)
    probe = os.path.join(scratch, "probe")
    out = run(["go", "build", "-o", probe, package], cwd=tree)
    if out.returncode != 0:
        print(out.stderr, file=sys.stderr)
        sys.exit("go build %s failed at %s — the commit itself may not compile "
                 "standalone; try a descendant commit that does" % (package, commit))
    with open(probe, "rb") as fh:
        return fh.read()


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("commit", help="the commit whose deployment you want to prove")
    ap.add_argument("--package", default="./cmd/agent-chassis",
                    help="go package to build as the probe (default ./cmd/agent-chassis)")
    args = ap.parse_args()

    rev = run(["git", "rev-parse", "--verify", args.commit + "^{commit}"])
    if rev.returncode != 0:
        print("cannot resolve %r as a commit" % args.commit, file=sys.stderr)
        return 2
    sha = rev.stdout.strip()
    svc = os.path.basename(args.package.rstrip("/"))

    added, removed = harvest(sha)
    if not added:
        print("no added non-test .go string literals (>=8 chars) in %s" % sha[:12],
              file=sys.stderr)
        return 1

    with tempfile.TemporaryDirectory(prefix="pick-pod-marker-") as scratch:
        blob = build_probe(sha, args.package, scratch)

    seen = set()
    verified, dead = [], []
    for c in added:
        if c["lit"] in seen:
            continue
        seen.add(c["lit"])
        n = blob.count(c["lit"].encode())
        (verified if n else dead).append((n, c))
    # Rank: fewest occurrences first (most distinctive); prose beats
    # backslash-soup (a regex literal is a terrible thing to paste into a
    # grep); longest first within.
    verified.sort(key=lambda t: (t[0], "\\" in t[1]["lit"],
                                 " " not in t[1]["lit"], -len(t[1]["lit"])))

    synthetic = "neg_control_%s_xyzzy" % sha[:12]
    assert synthetic.encode() not in blob
    neg = next((c["lit"] for c in removed
                if c["lit"] not in seen and blob.count(c["lit"].encode()) == 0), None)

    print("probe: %s built from git archive %s\n" % (args.package, sha[:12]))
    if dead:
        print("NOT in the binary — these are the trap, do not probe with them:")
        for _, c in dead:
            why = "comment" if c["comment"] else "unused/excluded"
            print("  [%s] %-60r %s" % (why, c["lit"][:60], c["path"]))
        print()
    if not verified:
        print("no added literal survived the binary — pick a marker by hand from a "
              "compiled format string, or probe a different package", file=sys.stderr)
        return 1

    print("VERIFIED markers (count in probe binary — expect the same shape in the pod):")
    for n, c in verified[:8]:
        print("  %2d× %-60r %s" % (n, c["lit"][:60], c["path"]))
    print("\nnegative controls (expect 0):")
    if neg:
        print("  removed by this commit:  %r  — proves the ROLL carried the change" % neg[:60])
    print("  synthetic:               %r  — proves only the PIPELINE" % synthetic)

    best = verified[0][1]["lit"][:60]
    negmark = neg[:60] if neg else synthetic
    # Trailing `:`— the negative grep EXITS 1 when it (correctly) finds 0, and
    # "command terminated with exit code 1" reads as failure when it is the pass.
    payload = ("strings /app/{svc} | grep -cF {pos}; "
               "strings /app/{svc} | grep -cF {neg}; :").format(
                   svc=svc, pos=shlex.quote(best), neg=shlex.quote(negmark))
    print("\nrun on EVERY replica (a roll is not evidence your fix shipped):")
    print("  for p in $(kubectl get pods -n %s -l app=%s -o name); do" % (NS, svc))
    print('    echo "$p"; kubectl exec -n %s "${p#pod/}" -- sh -c %s' %
          (NS, shlex.quote(payload)))
    print("  done")
    print("\n(first number: your marker, expect >=1 · second: the negative control, expect 0)")

    # A marker set is proved ONCE by the loop above and then AGAIN after every
    # roll, because a proof carries the tag it was taken on. Rebuilding a probe
    # binary each time is wasted work once the markers are known, so hand the set
    # to the verifier and give it a name. It asserts counts on every replica,
    # refuses a non-ASCII marker outright, and separates "could not look" (2)
    # from "looked and it is not there" (1).
    # PRESENCE, not an exact count. The counts above are this PROBE's, built at
    # <commit>; the deployed binary is a descendant, and a later commit can add or
    # drop a call site while the string survives. Measured 2026-08-06: the probe at
    # 87ea0a5e7 holds the contact-detail marker once, the live chassis twice — so a
    # count harvested here and saved as an assertion fails on arrival, and the
    # failure looks like a missing fix. Tighten to `=N` yourself once you have read
    # the count off the running binary, which is the only place it is authoritative.
    save = " \\\n      ".join(
        ["scripts/prove-live-markers.py --save <name> --deployment %s" % svc] +
        ["--expect %s" % shlex.quote(c["lit"][:60]) for _, c in verified[:3]] +
        (["--negative %s" % shlex.quote(neg[:60])] if neg else []))
    print("\nto re-prove this after EVERY future roll, save it once:")
    print("  " + save)
    print("  scripts/prove-live-markers.py <name>        # every roll thereafter")
    print("  (markers saved as PRESENCE — the counts above are the probe's, and a "
          "later commit\n   may add a call site. Tighten to '<marker>=N' from the "
          "LIVE count if you want that.)")
    if not neg:
        print("  (no removed string in this commit, so the saved control would be "
              "synthetic — it proves the pipeline, not the roll)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
