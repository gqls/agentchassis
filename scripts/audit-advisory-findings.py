#!/usr/bin/env python3
"""audit-advisory-findings.py — do `pattern-check.py`'s ADVISORY findings ever get ACTED ON?

WHY THIS EXISTS
`RFC_008` (a mandatory write seam for `page_components.rendered_html`) was closed on
2026-08-22 with NO seam, on the reasoning that this estate detects and attributes rather
than refuses. But that RFC named its own decisive question and nobody had answered it:

    "does the advisory check actually get read? ... pattern-check.py now hosts several
     advisory rules and NOBODY HAS EVER MEASURED whether any of them changes behaviour.
     If the answer is 'findings are ignored', that is an argument for mandatory seams
     across the board and a much bigger finding than this RFC."

The RFC's reopen trigger was first routed at `bugs_open/358`'s finding-code registry, and
that was WRONG (corrected same day): 358 can measure `agent_error_log` because codes are
ROWS. `pattern-check.py`'s findings are printed to a terminal at commit time and persist
NOWHERE, so "was this finding read?" had nothing to query. This script is the answer, and
it needs no new plumbing at all, because of one property of the detector:

    **THE FINDINGS ARE RECOMPUTABLE.** `pattern-check.py --commit <sha>` already replays
    any past commit in isolation. So the record does not have to be WRITTEN at commit time
    — it can be RECONSTRUCTED from git, for the whole of history, today.

That matters beyond convenience. The sketched alternative — write a durable row from the
pre-commit hook — would have put a network write (this estate reaches its DB through
`kubectl exec`, on a token that expires every three days) inside the hook that every
session runs on every commit, whose own header says "a stray non-zero exit here stops the
whole fleet committing. Keep it boring." And it would have measured only commits made
AFTER it shipped, i.e. answered the question in several weeks rather than now. Nothing in
this script runs at commit time; `pattern-check.py` is imported, never modified.

WHAT IT MEASURES
For every finding that fired in the window: was the condition subsequently REMOVED (a fix,
or an allow-list entry) or is it STILL TRUE at HEAD? A finding that fired and is still true
is one nobody acted on.

THE TRAP THIS IS BUILT AROUND, AND THE CONTROL THAT ANSWERS IT
Most of these checks are DIFF-SCOPED: they examine files a commit changed. Replayed
statically against HEAD there is no diff, so such a check emits nothing — **which is
byte-identical to "the problem was fixed".** A naive auditor would report every
diff-scoped finding as resolved and produce a triumphant, meaningless number. (That is
this estate's most expensive recurring shape: a zero that could not have come out
otherwise. See `a-post-fix-zero-needs-a-demand-control`.)

The discriminator is empirical and needs no hand-classification of the checks:

    Replay each check at ITS OWN FINDING'S COMMIT — content as of that commit, no diff.
    The condition was TRUE there; that is why the check fired. So the answer is known:
      * check fires  -> it reads STATE, so "does it still hold?" is a question it can
        answer, and its verdicts can be trusted.
      * check silent -> it reads the DIFF. It can never answer that question, and ALL
        its findings are reported UNDECIDABLE — never `acted`.

So each check earns the right to be believed, per run, from cases whose answer is known
in advance. A check that never fires on static replay is undecidable by default, never
resolved by default. The per-check evidence is printed, so a reader can see which checks
carried the run.

    ⚠ THIS REPLACED AN EARLIER CALIBRATION THAT WAS BIASED AGAINST FINDING ACTION, and
    the bias inverted the headline, so it is recorded rather than quietly fixed. Version
    one calibrated only on findings whose FILE WAS UNCHANGED since firing. But a check
    whose findings were ALL ACTED ON has no unchanged file, by definition — so its
    successes were filed as `undecidable` and it never reached the numerator. Measured
    2026-08-22: `unrepaired-component-write` fired on three writers, all three were fixed
    that day, and the run reported `acted: 0` while omitting the check from the control
    table entirely. **A control that cannot observe the outcome it exists to detect is not
    a control** — and it fails in the flattering direction, which is the dangerous one.

USAGE
    scripts/audit-advisory-findings.py [-n 150] [--json] [--write-note]

`--write-note` records ONE `doc_notes` row per run INCLUDING A CLEAN ONE, in the shape
every daily check in this estate uses: a missing row then means the job did not run, which
is not the same as "nothing is wrong". Requires `CLIENTS_DB_PASSWORD`, or falls back to
`kubectl exec postgres-clients-0`.

NOT BUILT, DELIBERATELY: the daily CronJob. Every check in that family runs
`postgres:16-alpine` + a script, with no git clone and no repo — and this auditor is
git-shaped to its core. Scheduling it means either shipping a clone into the image or
mounting the repo, which is a real decision and not a wiring detail. Named in the register
entry's verify-later rather than left to be discovered.
"""
import argparse
import importlib.util
import json
import os
import re
import subprocess
import time
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ANSI = re.compile(r"\033\[[0-9;]*m")
YELLOW, DIM, BOLD, RESET = "\033[1;33m", "\033[2m", "\033[1m", "\033[0m"


def load_pattern_check():
    """Import pattern-check.py by path — its name has a hyphen, so no plain import."""
    path = os.path.join(HERE, "pattern-check.py")
    spec = importlib.util.spec_from_file_location("pattern_check", path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def sh(*args):
    return subprocess.run(args, capture_output=True, text=True, cwd=HERE + "/..").stdout


def rc(*args):
    return subprocess.run(args, capture_output=True, text=True, cwd=HERE + "/..").returncode


def findings_for_commit(pc, sha, owner=None):
    """Replay every check against ONE commit, exactly as the hook would have seen it.

    Each check is run into its OWN list so the finding's `kind` can be attributed to the
    check that produced it (`owner`) in the same pass — the attribution map used to cost a
    second full sweep of history, which doubled the runtime of a 5,000-commit audit for
    information the first pass already had in hand.
    """
    ref = (sha + "~1", sha)
    files = pc.changed_files(ref)
    if not files:
        return []
    res = []
    for check in pc.CHECKS:
        out = []
        try:
            check(files, ref, out)
        except Exception:
            continue                  # a broken check must not break the audit
        for (k, w, _what, _why) in out:
            if owner is not None:
                owner.setdefault(k, check.__name__)
            res.append((k, ANSI.sub("", w)))
    return res


def still_true_at_head(pc, check_name, path, base):
    """Replay ONE check against ONE path with BASE's content and no diff.

    ref=(base, base) makes pattern-check read content via `git show <base>:<path>`,
    so this asks a question about STATE, not about a change.

    ⚠ `base` is a PINNED SHA, never the literal "HEAD", and that is load-bearing on this
    tree. A 5,000-commit sweep takes ~15 minutes; this repo takes ~4,967 commits in 14
    days, so HEAD MOVES UNDERNEATH THE RUN and findings judged early would be compared
    against a different tree from findings judged late. Not theoretical: on 2026-08-22 a
    sibling session fixed `create_tool_component_regenerate.go` DURING a sweep, so the
    same finding was "still true" at minute 2 and false at minute 14. A measurement whose
    baseline moves cannot be reproduced or defended.
    """
    check = next((c for c in pc.CHECKS if c.__name__ == check_name), None)
    if check is None:
        return None
    out = []
    try:
        check([path], (base, base), out)
    except Exception:
        return None
    return any(k == KIND_OF.get(check_name, k) for (k, _w, _a, _b) in out) or bool(out)


# A finding's `kind` string is not the check's function name; map what we see back to
# the check that produced it by re-running each check and recording which kinds it emits.
KIND_OF = {}


def kinds_by_check(pc, sample_commits):
    """Learn which check emits which `kind`, from real replays. No hand-maintained table."""
    owner = {}
    for sha in sample_commits:
        ref = (sha + "~1", sha)
        files = pc.changed_files(ref)
        if not files:
            continue
        for check in pc.CHECKS:
            out = []
            try:
                check(files, ref, out)
            except Exception:
                continue
            for (k, _w, _a, _b) in out:
                owner.setdefault(k, check.__name__)
    return owner


# DEMAND CONTROL, pinned. A sweep that reports "no findings" is worthless unless the
# replay machinery can be shown to produce one, and that distinction nearly went out
# undetected: the first run of this script returned 0 findings over 25 commits, which
# looked like a broken replay and was in fact a true zero — the commit I knew fired had
# already fallen out of the window (this tree takes >25 commits in an afternoon).
# So: a known-firing commit, asserted before any sweep is believed.
#   d795e10f5 — the RFC_008 ruling commit; corrected two LANDMINES.md lines in place,
#   which `check_append_only_docs` reports as `shared-ledger-not-appended`.
# Forward-only history (no resets/rebases, CLAUDE.md) is what makes a pinned sha safe.
# If the pin ever fails to RESOLVE, that is a hard failure, never a silent pass.
SELF_TEST_COMMIT = "d795e10f5"
SELF_TEST_EXPECT = ("shared-ledger-not-appended",
                    "docs/agent_docs/docs024_key_docs_latest/LANDMINES.md")


def self_test(pc):
    """Prove the replay can SEE a finding before any zero from it is believed."""
    if rc("git", "cat-file", "-e", SELF_TEST_COMMIT + "^{commit}") != 0:
        print(f"self-test REFUSED: pinned commit {SELF_TEST_COMMIT} does not resolve — "
              f"this audit cannot demonstrate it can see anything", file=sys.stderr)
        return 2
    got = findings_for_commit(pc, SELF_TEST_COMMIT)
    if SELF_TEST_EXPECT in [(k, w.split(":")[0]) for (k, w) in got]:
        # stderr, ALWAYS: this is diagnostic, and on stdout it corrupts --json for any
        # caller that pipes it (caught by actually piping it into a JSON parser rather
        # than by reading the code — the parse error was the only signal).
        print(f"{DIM}self-test PASS: replay of {SELF_TEST_COMMIT} yields "
              f"{SELF_TEST_EXPECT[0]}{RESET}", file=sys.stderr)
        return 0
    print(f"self-test FAIL: replay of {SELF_TEST_COMMIT} did NOT yield {SELF_TEST_EXPECT[0]} "
          f"(got {got!r}) — the sweep below cannot be trusted", file=sys.stderr)
    return 2


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("-n", type=int, default=150, help="commits to sweep (default 150)")
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--write-note", action="store_true")
    ap.add_argument("--self-test", action="store_true",
                    help="only run the pinned demand control, then exit")
    args = ap.parse_args()

    pc = load_pattern_check()

    # The control runs on EVERY invocation, not only under --self-test: a zero this
    # script prints must always come with proof that it could have printed non-zero.
    st = self_test(pc)
    if args.self_test:
        return st
    if st != 0:
        return st
    # PIN the comparison baseline once. Everything "at HEAD" below is judged against
    # this immutable sha, not against a HEAD that other sessions keep moving.
    base = sh("git", "rev-parse", "HEAD").strip()
    if not base:
        print("audit-advisory-findings: cannot resolve HEAD", file=sys.stderr)
        return 2
    commits = [c for c in sh("git", "log", "--format=%H", "-n", str(args.n), base).split() if c]
    if not commits:
        print("audit-advisory-findings: no commits in range — refusing to report a clean sweep",
              file=sys.stderr)
        return 2

    # 1. Replay the window, learning kind->check attribution in the SAME pass.
    fired = {}                                   # (kind, path) -> [shas]
    owner = {}
    for sha in commits:
        for (kind, where) in findings_for_commit(pc, sha, owner):
            path = where.split(":")[0]
            fired.setdefault((kind, path), []).append(sha)

    KIND_OF.update({v: k for k, v in owner.items()})

    # 2. For each finding, is the file unchanged since it fired, and does it still fire?
    rows = []
    for (kind, path), shas in sorted(fired.items()):
        first = shas[-1]                          # git log is newest-first
        exists = rc("git", "cat-file", "-e", f"{base}:{path}") == 0
        unchanged = exists and rc("git", "diff", "--quiet", first, base, "--", path) == 0
        check_name = owner.get(kind)
        fires = still_true_at_head(pc, check_name, path, base) if (check_name and exists) else False
        try:
            age = (time.time() - int(sh("git", "log", "-1", "--format=%ct", first).strip())) / 86400.0
        except (ValueError, TypeError):
            age = -1.0
        rows.append({"kind": kind, "path": path, "hits": len(shas), "first": first[:9],
                     "first_full": first,
                     "age_days": round(age, 1),
                     "exists": exists, "unchanged": unchanged, "fires_at_head": bool(fires),
                     "check": check_name})

    # 3. THE CONTROL — calibrate each check by replaying it at the finding's OWN commit.
    #
    # The condition was TRUE at that commit; that is why the check fired. So replaying the
    # check statically against that commit's content, with no diff, has a known answer:
    #   fires  -> the check reads STATE, so "does it still hold?" is a question it can answer
    #   silent -> the check reads the DIFF, and can never answer it. Undecidable, never acted.
    #
    # ⚠ THIS REPLACED A CALIBRATION THAT WAS BIASED AGAINST FINDING ACTION, and the bias
    # is worth stating because it inverted the headline. The first version calibrated only
    # on findings whose FILE WAS UNCHANGED since firing. But a check whose findings were
    # ALL ACTED ON has no unchanged file by definition — so its successes were recorded as
    # `undecidable` and it never appeared in the numerator. Measured 2026-08-22:
    # `unrepaired-component-write` fired on three writers, all three were fixed that day,
    # and the run reported `acted: 0` while listing the check nowhere at all. A control
    # that cannot observe the outcome it exists to detect is not a control.
    evaluable, control_evidence = set(), {}
    for r in rows:
        own = still_true_at_head(pc, r["check"], r["path"], r["first_full"]) if r["check"] else None
        r["fires_at_own_commit"] = bool(own)
        control_evidence.setdefault(r["check"], []).append(bool(own))
    for check_name, results in control_evidence.items():
        if any(results):
            evaluable.add(check_name)

    # 4. Verdicts.
    for r in rows:
        if not r["exists"]:
            r["verdict"] = "file_gone"
        elif r["check"] not in evaluable:
            r["verdict"] = "undecidable"
        elif r["fires_at_head"]:
            r["verdict"] = "unacted"
        else:
            r["verdict"] = "acted"

    tally = {}
    for r in rows:
        tally[r["verdict"]] = tally.get(r["verdict"], 0) + 1

    decided = tally.get("unacted", 0) + tally.get("acted", 0)
    body = render(commits, rows, tally, evaluable, control_evidence, decided, base)

    if args.json:
        print(json.dumps({"commits": len(commits), "findings": rows, "tally": tally,
                          "state_evaluable_checks": sorted(evaluable),
                          "baseline": base}, indent=2))
    else:
        print(body)

    if args.write_note:
        write_note(body)

    return 0


def render(commits, rows, tally, evaluable, control_evidence, decided, base):
    b = []
    b.append(f"advisory-findings audit — {len(commits)} commits swept "
             f"({commits[-1][:9]}..{commits[0][:9]})")
    b.append(f"  baseline PINNED at {base[:9]} — every 'still true?' verdict below is "
             f"judged against that one tree")
    b.append("")
    b.append(f"  distinct (check, file) findings replayed : {len(rows)}")
    b.append(f"    UNACTED  (condition still true at HEAD): {tally.get('unacted', 0)}")
    b.append(f"    ACTED    (fixed or allow-listed)       : {tally.get('acted', 0)}")
    b.append(f"    file_gone                              : {tally.get('file_gone', 0)}")
    b.append(f"    UNDECIDABLE (diff-only check)          : {tally.get('undecidable', 0)}")
    b.append("")
    # AGE IS A CONFOUND AND MUST BE SHOWN. A finding committed this morning has not been
    # ignored; it has not been given the chance. On this tree 400 commits is ~1.5 days,
    # so a naive sweep reports "0% acted on" about work nobody could yet have acted on.
    ripe = [r for r in rows if r["verdict"] in ("unacted", "acted") and r["age_days"] >= 2.0]
    green = [r for r in rows if r["verdict"] in ("unacted", "acted") and 0 <= r["age_days"] < 2.0]
    if decided:
        pct = 100.0 * tally.get("acted", 0) / decided
        b.append(f"  raw follow-through, ALL decidable findings: {tally.get('acted', 0)}/{decided} "
                 f"= {pct:.0f}% acted on")
        if ripe:
            ra = sum(1 for r in ripe if r["verdict"] == "acted")
            b.append(f"  ⤷ RIPE ONLY (first fired >=2 days ago, so action was POSSIBLE): "
                     f"{ra}/{len(ripe)} acted on  <-- THIS is the answer to RFC_008's question")
        else:
            b.append("  ⤷ RIPE ONLY: **NO decidable finding in this window is >=2 days old**, so "
                     "this run CANNOT answer the follow-through question. Widen -n. The raw "
                     "figure above is dominated by findings too recent to have been acted on "
                     "and MUST NOT be quoted as neglect.")
        if green:
            b.append(f"  ⤷ ({len(green)} decidable finding(s) are <2 days old and excluded above)")
        b.append("  ⚠ WINDOW ARTEFACT, in the other direction: a finding that first fired BEFORE")
        b.append("    this window and was fixed INSIDE it is never replayed, so `acted` UNDER-")
        b.append("    counts. Worked case 2026-08-22: `unrepaired-component-write` had fired on")
        b.append("    the two tool writers since 2026-08-02 (outside a 400-commit window), was")
        b.append("    fixed by `bugs_open/362`, and both files verify SILENT at HEAD — a real")
        b.append("    `acted` the sweep could not see. The arm is reachable; the window is short.")
    else:
        b.append("  FOLLOW-THROUGH: NOT MEASURABLE this run — no finding was decidable. "
                 "This is a BLIND result, not a clean one; do not read it as 'all well'.")
    b.append("")
    b.append("  the control — which checks earned the right to be believed this run")
    b.append("  (each check is replayed at its OWN finding's commit, where the condition is KNOWN")
    b.append("   to have held; silence there means the check reads the DIFF, not state, so it can")
    b.append("   never answer 'still true?' and all its findings are undecidable, never acted)")
    if control_evidence:
        for check_name, results in sorted(control_evidence.items()):
            mark = "state-evaluable" if check_name in evaluable else "DIFF-ONLY -> undecidable"
            b.append(f"    {check_name:38s} {sum(1 for x in results if x)}/{len(results)} "
                     f"own-commit replays fired  {mark}")
    else:
        b.append("    NONE — no finding had an unchanged file, so nothing could be calibrated.")
    b.append("")
    unacted = [r for r in rows if r["verdict"] == "unacted"]
    if unacted:
        b.append("  UNACTED findings (fired, condition still true at HEAD):")
        for r in sorted(unacted, key=lambda r: -r["hits"]):
            b.append(f"    {r['kind']:34s} {r['path']}  (fired in {r['hits']} commit(s), "
                     f"first {r['first']})")
    b.append("")
    b.append("  This row exists on a clean run ON PURPOSE: a MISSING row means the audit did")
    b.append("  not run, which is not the same as 'nothing is wrong'.")
    b.append("  Source: scripts/audit-advisory-findings.py (RFC_008's decisive question).")
    return "\n".join(b)


def write_note(body):
    sql_body = body.replace("$aafbody$", "")
    sql = ("INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
           "VALUES ('pipeline', 'advisory-findings', $aafbody$" + sql_body + "$aafbody$, "
           "'[\"advisory-findings\"]'::jsonb, 'advisory-findings-audit');")
    p = subprocess.run(
        ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
         "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1", "-c", sql],
        capture_output=True, text=True)
    if p.returncode != 0:
        print(f"{DIM}note NOT written: {p.stderr.strip()[:200]}{RESET}", file=sys.stderr)
        return 1
    print(f"{DIM}doc_notes row written (subject_key='advisory-findings').{RESET}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
