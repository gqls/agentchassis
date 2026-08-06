#!/usr/bin/env python3
"""prove-live-markers — run a marker set against EVERY replica and assert the counts.

The other half of scripts/pick-pod-marker.py. That one CHOOSES markers a binary
can actually contain (expensive: it builds a probe from `git archive <commit>`).
This one RUNS a known marker set against the live pods and gives a verdict, which
is what you need after every roll — and this fleet rolls every few hours.

Why it exists (2026-08-06): the bugfix_140 lane re-proved one fix by hand FOUR
times in three days (v1.0.1250, 1251, 1252, 1257, 1259) because a proof carries
the tag it was taken on and a rolled fleet retires it. Hand-running the same four
greps across two replicas is exactly the kind of repetition that gets shortened
to one replica, or to a positive marker with no negative control, on the fifth
day when someone is busy. So it is mechanical now, and the rules the lane learned
the hard way are enforced by the tool rather than remembered:

  * EVERY replica, never one. `kubectl logs deploy/X` reads one pod of N, and so
    does a habit of grepping "the" pod.
  * A NEGATIVE CONTROL is mandatory. A positive marker alone proves the grep
    pipeline works, not that your build shipped. Supply a string your change
    REMOVED (the strong form: it proves the roll); if you have none, this tool
    synthesises one and says plainly that a synthetic control is the weak form.
  * NON-ASCII MARKERS ARE REFUSED. The probe is `strings | grep`, and `strings`
    emits ASCII runs — one multibyte character (…, —, £) SPLITS the run, so a
    marker spanning it greps 0 against a binary that provably contains it. That
    trap has fired on this tree more than once; here it is a startup error rather
    than a mysterious zero.
  * Pods come from the DEPLOYMENT'S OWN SELECTOR, not from a guessed
    `-l app=<name>`. On this cluster one image serves several subsystems and a
    label can select the wrong service's pods.
  * The verdict is stamped with the IMAGE TAG it was taken on, because that is
    the fact that expires.

Usage:
    scripts/prove-live-markers.py rfc009                 # a saved profile
    scripts/prove-live-markers.py --list
    scripts/prove-live-markers.py --deployment agent-chassis \\
        --expect "some added text=2" --expect "another" --negative "text you removed"
    scripts/prove-live-markers.py --save myfix --deployment agent-chassis \\
        --expect "..." --negative "..."          # save, then run it

`--expect "TEXT=N"` asserts exactly N matches per replica; `--expect "TEXT"`
asserts at least one. Exact counts catch a call site being deleted while the
string survives elsewhere, which presence alone cannot see.

Exit: 0 every marker agrees on every replica · 1 a mismatch · 2 setup failure
(no pods, unreadable profile, a refused marker).
"""

import argparse
import json
import os
import shlex
import subprocess
import sys

NS = os.environ.get("NAMESPACE", "ai-persona-system")
PROFILES = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                        "live_marker_profiles.json")
# Exists nowhere by construction. Used only when the caller supplies no removed
# string — see the weak/strong distinction in the docstring.
SYNTHETIC = "zzz_marker_that_exists_in_no_build_prove_live"


def die(msg):
    """Setup failure — exit 2, NEVER 1.

    Written after the first round of arm-testing caught it: bare `sys.exit("msg")`
    exits 1, so a refused marker and a missing deployment both reported the code
    that means "the fix is not in the binary". A caller — or a session reading an
    exit code in a hurry — could not tell "I could not look" from "I looked and it
    is not there", which is the same could-not-look-vs-clean ambiguity the checks
    in this family exist to remove. Same convention as
    check_placeholder_fallbacks.py: 2 means the probe never ran.
    """
    print(msg, file=sys.stderr)
    sys.exit(2)


def run(argv, **kw):
    return subprocess.run(argv, capture_output=True, text=True, **kw)


def load_profiles():
    if not os.path.exists(PROFILES):
        return {}
    try:
        with open(PROFILES) as fh:
            return json.load(fh)
    except (OSError, ValueError) as exc:
        die("cannot read %s: %s" % (PROFILES, exc))


def save_profile(name, spec):
    profiles = load_profiles()
    profiles[name] = spec
    with open(PROFILES, "w") as fh:
        json.dump(profiles, fh, indent=2, sort_keys=True)
        fh.write("\n")
    print("saved profile %r to %s" % (name, os.path.relpath(PROFILES)))


def parse_expect(raw):
    """"TEXT=N" -> (TEXT, N); "TEXT" -> (TEXT, None) meaning >= 1.

    Split on the LAST '=' so a marker containing '=' survives. A trailing
    '=<non-integer>' is treated as part of the text, not as a broken count.
    """
    if "=" in raw:
        head, _, tail = raw.rpartition("=")
        if head and tail.strip().isdigit():
            return head, int(tail)
    return raw, None


def refuse_non_ascii(markers):
    bad = [m for m in markers if not (m.isascii() and m.isprintable())]
    if bad:
        for m in bad:
            offenders = {c for c in m if not (c.isascii() and c.isprintable())}
            print("REFUSED marker (non-ASCII/unprintable %s): %r"
                  % ("".join(sorted(offenders)) or "char", m), file=sys.stderr)
        die("`strings` splits its output at a non-ASCII byte, so this marker "
            "would grep 0 against a binary that contains it. Shorten the marker "
            "to the ASCII run BEFORE the offending character.")


def resolve(deployment):
    """(image, [pod names]) — pods via the Deployment's own selector, not a guess."""
    d = run(["kubectl", "-n", NS, "get", "deploy", deployment, "-o", "json"])
    if d.returncode != 0:
        die("cannot read deployment %s: %s" % (deployment, d.stderr.strip()))
    spec = json.loads(d.stdout)
    image = spec["spec"]["template"]["spec"]["containers"][0]["image"]
    match = spec["spec"]["selector"].get("matchLabels") or {}
    if not match:
        die("deployment %s has no matchLabels selector to resolve pods with"
            % deployment)
    sel = ",".join("%s=%s" % kv for kv in sorted(match.items()))
    p = run(["kubectl", "-n", NS, "get", "pods", "-l", sel,
             "--field-selector=status.phase=Running",
             "-o", "jsonpath={.items[*].metadata.name}"])
    if p.returncode != 0:
        die("cannot list pods for %s: %s" % (sel, p.stderr.strip()))
    pods = p.stdout.split()
    if not pods:
        die("no Running pods matched %s — nothing to prove against" % sel)
    return image, pods


def count_in_pod(pod, binary, text):
    """Matches of `text` in `binary` on `pod`, or None if the probe itself failed.

    grep -c EXITS 1 on a zero count and prints '0', so a non-zero return is not
    an error here — which is precisely why the count is parsed rather than the
    exit code trusted. Anything unparseable is None (a broken probe), never 0,
    because a broken probe that reads as 0 is a false negative.
    """
    cmd = "strings %s | grep -c -F -- %s" % (shlex.quote(binary), shlex.quote(text))
    out = run(["kubectl", "-n", NS, "exec", pod, "--", "sh", "-c", cmd])
    body = out.stdout.strip()
    if body.isdigit():
        return int(body)
    print("  probe failed on %s: %s" % (pod, (out.stderr.strip() or body)[:200]),
          file=sys.stderr)
    return None


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("profile", nargs="?", help="a saved profile name")
    ap.add_argument("--list", action="store_true", help="list saved profiles and exit")
    ap.add_argument("--deployment", default="agent-chassis")
    ap.add_argument("--binary", default="", help="path in the pod (default /app/<deployment>)")
    ap.add_argument("--expect", action="append", default=[], metavar="TEXT[=N]",
                    help="marker that MUST be present (=N for an exact count)")
    ap.add_argument("--negative", action="append", default=[], metavar="TEXT",
                    help="marker that must be ABSENT — ideally a string your change removed")
    ap.add_argument("--save", metavar="NAME", help="save these flags as a profile, then run")
    args = ap.parse_args()

    profiles = load_profiles()
    if args.list:
        if not profiles:
            print("no saved profiles in %s" % os.path.relpath(PROFILES))
            return 0
        for name, spec in sorted(profiles.items()):
            print("%-12s %-16s %s" % (name, spec.get("deployment", "agent-chassis"),
                                      spec.get("note", "")))
            for e in spec.get("expect", []):
                print("               expect   %s" % e)
            for n in spec.get("negative", []):
                print("               negative %s" % n)
        return 0

    deployment, binary = args.deployment, args.binary
    expect, negative, note = list(args.expect), list(args.negative), ""
    if args.profile:
        if args.profile not in profiles:
            die("no profile %r — try --list (file: %s)"
                % (args.profile, os.path.relpath(PROFILES)))
        spec = profiles[args.profile]
        deployment = spec.get("deployment", deployment)
        binary = spec.get("binary", binary)
        expect += spec.get("expect", [])
        negative += spec.get("negative", [])
        note = spec.get("note", "")
    if not expect:
        die("nothing to prove: give --expect or a profile")

    binary = binary or "/app/%s" % deployment
    pairs = [parse_expect(e) for e in expect]
    refuse_non_ascii([t for t, _ in pairs] + list(negative))

    synthetic = not negative
    if synthetic:
        negative = [SYNTHETIC]

    if args.save:
        save_profile(args.save, {"deployment": deployment, "binary": binary,
                                 "expect": expect,
                                 "negative": [] if synthetic else negative,
                                 "note": note or "saved by --save"})

    image, pods = resolve(deployment)
    print("deployment %s  image %s" % (deployment, image))
    print("binary %s  replicas %d%s" % (binary, len(pods),
                                        ("  [%s]" % note) if note else ""))
    print()

    failures = []
    width = max([len(t) for t, _ in pairs] + [len(n) for n in negative] + [20])
    width = min(width, 62)
    for pod in pods:
        print("== %s ==" % pod)
        for text, want in pairs:
            got = count_in_pod(pod, binary, text)
            ok = got is not None and (got == want if want is not None else got >= 1)
            print("  %-*s  want %-8s got %-6s %s"
                  % (width, text[:width], want if want is not None else ">=1",
                     "err" if got is None else got, "OK" if ok else "*** FAIL ***"))
            if not ok:
                failures.append((pod, text, want, got))
        for text in negative:
            got = count_in_pod(pod, binary, text)
            ok = got == 0
            print("  %-*s  want %-8s got %-6s %s  (negative control%s)"
                  % (width, text[:width], 0, "err" if got is None else got,
                     "OK" if ok else "*** FAIL ***", ", synthetic" if synthetic else ""))
            if not ok:
                failures.append((pod, text, 0, got))
        print()

    if synthetic:
        print("NOTE: the negative control is SYNTHETIC, so it proves the grep pipeline "
              "runs — not that your build shipped. The strong form is a string your "
              "change REMOVED: pass it with --negative and the control then fails if "
              "the pods are still on the old image.")
    if failures:
        print("\nFAILED: %d mismatch(es) across %d replica(s)." % (len(failures), len(pods)))
        print("A COUNT that differs (rather than dropping to 0) usually means the SOURCE "
              "changed — a call site added or removed — not that the fix is gone. "
              "Re-derive markers with: scripts/pick-pod-marker.py <commit>")
        return 1

    print("PASS: %d marker(s) + %d negative control(s) agree on all %d replica(s)."
          % (len(pairs), len(negative), len(pods)))
    print("Proved on %s. This verdict expires at the next roll — it is a statement "
          "about that tag, not about the repo." % image)
    return 0


if __name__ == "__main__":
    sys.exit(main())
