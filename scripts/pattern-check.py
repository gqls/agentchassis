#!/usr/bin/env python3
"""pattern-check.py — ADVISORY. Runs the MECHANICALLY CHECKABLE subset of the
debugging guide's section 9 patterns against what a commit actually changes.

WHY THIS EXISTS (2026-07-19, and the origin is the point):
016b section 9 has 28 patterns. On 2026-07-19 a session wrote entry 26 — "a fix
applied to one branch of a two-branch router reads as done" — and then committed
exactly that mistake eight hours later: it fixed `withPriorCodeRequests` and left
its near-identical twin `withPriorRequests` carrying the same defect. The pattern
was not forgotten. It had been WRITTEN THAT MORNING, by the same session. It was
simply never connected to the moment of editing.

That is the gap this closes, and it is narrow on purpose. Knowing a pattern does
not fire it; something at the moment of the edit has to. A council of reviewers
did eventually catch that instance — correctly, at full reading load — but it
cost two rounds and real credits to answer a question `grep` answers in
milliseconds. The platform's own doctrine already says this for bugs ("this
platform's bugs mostly dissolve under grep + a schema read"); the same holds for
review. Spend the LLM council on judgement, not on what a string comparison can
settle.

WHAT IT DELIBERATELY DOES NOT DO:
- It does not check all 28 patterns. Most are judgement ("trust the rendered
  artefact, not the status") and are not expressible as a diff test. Section 25
  is literally a pattern about a check that CANNOT be written. Pretending
  otherwise would produce noise, and noise is fatal here — see below.
- It does not block. See ADVISORY below.
- It does not replace the council gate. It runs BEFORE it, so the council's
  reading budget is spent on things only a reader can find.

ADVISORY, AND WHY (do not "improve" this into a blocker without reading this):
`.githooks/pre-commit` warns that "a stray non-zero exit here stops the whole
fleet committing." Several sessions share this tree; a false positive that blocks
is a fleet-wide outage, and a check that blocks on a bad day gets disabled
permanently. The failure mode we are guarding against is worth a warning, not a
stoppage. 016b also records what happens to enforcement that annoys people: an
agreed invariant with no mechanical teeth decayed to a comment and was violated
by 84% of the CTA anchors it governed. A check nobody reads is worth nothing, so
precision matters more than coverage — every check here was measured against real
history before being included, and anything that fired on ordinary work was cut.

Set PATTERN_CHECK_STRICT=1 to make findings exit non-zero (opt-in, per-session).

USAGE:
    scripts/pattern-check.py            # staged changes (what the hook runs)
    scripts/pattern-check.py --commit <sha>  # audit ONE past commit in isolation
    scripts/pattern-check.py --ref HEAD~5    # audit a range, for measuring
"""
import json
import os
import re
import subprocess
import sys

YELLOW, DIM, BOLD, RESET = "\033[1;33m", "\033[2m", "\033[1m", "\033[0m"


def sh(*args):
    return subprocess.run(args, capture_output=True, text=True).stdout


def changed_files(ref=None):
    """ref is (base, head) for an audit, or None for the staged commit."""
    if ref:
        out = sh("git", "diff", "--name-only", "--diff-filter=ACMR", ref[0], ref[1])
    else:
        out = sh("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
    return [f for f in out.splitlines() if f.strip()]


def file_content(path, ref=None):
    """Content as of the commit under test (HEAD for a ref audit, worktree for staged)."""
    if ref:
        return sh("git", "show", f"{ref[1]}:{path}")
    try:
        with open(path, encoding="utf-8", errors="replace") as fh:
            return fh.read()
    except OSError:
        return ""


def raw_diff(path, ref=None):
    if ref:
        return sh("git", "diff", "-U0", ref[0], ref[1], "--", path)
    return sh("git", "diff", "--cached", "-U0", "--", path)


def changed_hunk_text(path, ref=None):
    """Only the ADDED/REMOVED lines for this file — what this commit actually touched."""
    d = raw_diff(path, ref)
    return "\n".join(l for l in d.splitlines() if re.match(r"^[+-][^+-]", l))


HUNK = re.compile(r"^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@", re.M)


def changed_lines(path, ref=None):
    """1-based line numbers touched in the NEW file. A zero-length hunk (pure
    deletion) still marks the line it was removed from, so deleting a function's
    guts counts as touching it."""
    out = set()
    for m in HUNK.finditer(raw_diff(path, ref)):
        start = int(m.group(1))
        count = int(m.group(2)) if m.group(2) is not None else 1
        for n in range(start, start + max(count, 1)):
            out.add(n)
    return out


def functions_with_spans(content):
    """[(name, first_line, last_line)] — a Go func runs to the line before the
    next top-level func. Crude, and correct for this purpose: we only need to
    know which function a changed line falls inside."""
    marks = [(m.start(), m.group(1)) for m in GO_FUNC.finditer(content)]
    if not marks:
        return []
    line_of = {}
    line = 1
    for i, ch in enumerate(content):
        line_of[i] = line
        if ch == "\n":
            line += 1
    total = line
    spans = []
    for idx, (off, name) in enumerate(marks):
        start = line_of.get(off, 1)
        end = (line_of.get(marks[idx + 1][0], total) - 1) if idx + 1 < len(marks) else total
        spans.append((name, start, end))
    return spans


# ── the CamelCase twin rule ─────────────────────────────────────────────────
# Two identifiers are TWINS when one is the other with exactly ONE CamelCase
# segment inserted. withPriorCodeRequests -> [with,Prior,Code,Requests] and
# withPriorRequests -> [with,Prior,Requests] differ by inserting "Code": twins.
# Anything differing by two or more segments is a different function that merely
# reads similarly, and pairing those is where the false positives live.
def segments(name):
    return re.findall(r"[A-Z]+(?![a-z])|[A-Z][a-z0-9_]*|^[a-z0-9_]+", name)


# A test double is not a twin: changing a real function without changing its
# fake is normal and correct. Measured — ExecuteLLMPromptAction vs
# ExecuteLLMPromptActionFAKE was the only twin false positive in 150 commits
# (a3b606798), and it is this entire class.
DOUBLE = re.compile(r"(FAKE|Fake|Mock|Stub|Spy|Dummy|Noop|NoOp|Test)$")


def is_twin(a, b):
    if DOUBLE.search(a) or DOUBLE.search(b):
        return False
    sa, sb = segments(a), segments(b)
    if len(sa) == len(sb):
        return False
    if abs(len(sa) - len(sb)) != 1:
        return False
    longer, shorter = (sa, sb) if len(sa) > len(sb) else (sb, sa)
    for i in range(len(longer)):
        if longer[:i] + longer[i + 1:] == shorter:
            return True
    return False


RECENT_TWIN_COMMITS = 10

GO_FUNC = re.compile(r"^func(?:\s+\([^)]*\))?\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(", re.M)


def check_untouched_twin(files, ref, findings):
    """016b section 9 #26 — a fix applied to one of a near-identical pair."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        content = file_content(path, ref)
        if not content:
            continue
        spans = functions_with_spans(content)
        all_funcs = [n for n, _, _ in spans]
        # Attribute changed LINES to their enclosing function. Matching on "the
        # name appears in the diff" (the first implementation) silently missed
        # the commonest case of all — an edit INSIDE a function body, where the
        # name never appears in any changed line. Found by running the real hook
        # against a real staged edit rather than trusting the past-commit audit,
        # which happened to consist only of added functions and signature changes.
        touched_lines = changed_lines(path, ref)
        touched = {n for n, a, b in spans if any(a <= ln <= b for ln in touched_lines)}
        for t in sorted(touched):
            for other in all_funcs:
                if other in touched or not is_twin(t, other):
                    continue
                why = ("016b section 9 #26 — same shape, same file, one edited. If the change is a "
                       "fix, the twin probably has the same defect; if it is a feature, the twin "
                       "probably needs it too. Deliberate? Say so in the commit message.")
                # A twin corrected in a RECENT commit is a deliberate two-commit
                # sequence, not an oversight — measured: 03e86fc32 (the class fix)
                # tripped this because its twin had been fixed one commit earlier.
                # Still reported, because "I already did it" and "I forgot" look
                # identical from inside; but say which, so the reader can dismiss
                # it in one glance instead of re-deriving it.
                recent = sh("git", "log", "-1", "--format=%h %s", "-S", other,
                            f"-{RECENT_TWIN_COMMITS}", "--", path).strip()
                if recent:
                    why += f"  [{other} was last touched in: {recent[:64]} — if that was this same piece of work, this is expected]"
                findings.append((
                    "untouched-twin", path,
                    f"changed {BOLD}{t}(){RESET} but not its twin {BOLD}{other}(){RESET}",
                    why,
                ))


def check_gofmt(files, ref, findings):
    """016b section 9 #16 — un-gofmt'd Go fails the build gate and yields no PR."""
    if ref:
        return  # only meaningful for a commit about to happen
    gofiles = [f for f in files if f.endswith(".go") and os.path.exists(f)]
    if not gofiles:
        return
    out = sh("gofmt", "-l", *gofiles).split()
    for f in out:
        findings.append((
            "gofmt", f, "not gofmt-clean",
            "016b section 9 #16 — the build gate rejects un-gofmt'd code, so this "
            "reaches CI as a failed gate and no PR. Run: gofmt -w " + f,
        ))


STDIN_EATERS = re.compile(r"\b(kubectl\s+[^|]*\bexec\b[^|]*\s-\w*i|ssh\s|ffmpeg\s)")


def check_stdin_eater(files, ref, findings):
    """016b section 9 #20 — a stdin-reading command truncates the while-read loop."""
    for path in files:
        if not (path.endswith(".sh") or path.endswith(".bash")):
            continue
        content = file_content(path, ref)
        lines = content.splitlines()
        depth_stack = []
        for i, line in enumerate(lines):
            if re.search(r"\bwhile\b.*\bread\b", line):
                depth_stack.append(i)
            if re.match(r"^\s*done\b", line) and depth_stack:
                start = depth_stack.pop()
                body = "\n".join(lines[start:i])
                m = STDIN_EATERS.search(body)
                if m and "</dev/null" not in body and "< /dev/null" not in body:
                    findings.append((
                        "stdin-eater", f"{path}:{start + 1}",
                        f"`while read` loop calls {BOLD}{m.group(1).strip()}{RESET} without redirecting its stdin",
                        "016b section 9 #20 — that command consumes the rest of the loop's input, "
                        "so the loop silently processes only the first few items (it hid 90% of "
                        "the council coverage report). Add < /dev/null to the inner command.",
                    ))


# Pairs that must change together, where the relationship is real but invisible
# to the compiler. Kept SHORT on purpose: each entry is a documented incident,
# not a guess. A speculative pair here would fire on ordinary work and teach
# people to ignore the whole script.
DECLARED_PAIRS = [
    (r"idx_swi_dedup", r"workItemTerminalStatuses",
     "016b/memory — the dedup index's status set and the Go terminal-status list are ONE "
     "contract. Drift gives fleet-wide 42P10 on every keyed insert (bit 2026-07-16)."),
    (r"FROM code_symbols", r"codeIndexFreshness",
     "bugs_open/059 — code_symbols is an indexed SNAPSHOT: a stale index answers 'absent' "
     "identically to a genuine absence (it sat 3 weeks stale and fed false negatives to the "
     "code tier AND the prior_art seat). Any renderer reading it must carry codeIndexFreshness "
     "so the answer names its own staleness. If this is the indexer/writer side or a "
     "non-LLM-facing read (analyse_repo_local), say so in the commit message and carry on."),
]


CODE_EXT = (".go", ".sql", ".sh", ".py", ".ts", ".tsx", ".js")

COMMENT = re.compile(r"(//|--|#).*$")


def strip_comments(text):
    """Drop line comments so a comment ABOUT an invariant is not read as a change
    to it. Crude (it will also blank a // inside a string literal) and that is the
    safe direction here: it can only ever suppress a finding, never invent one."""
    return "\n".join(COMMENT.sub("", l) for l in text.splitlines())


def check_declared_pairs(files, ref, findings):
    # Docs are EXCLUDED, and this is not a detail: measured over 150 commits the
    # only unpaired-change hit was a docs commit that merely DISCUSSED
    # idx_swi_dedup in prose (41e3345b2). A file explaining an invariant is not a
    # change to it, and a check that cannot tell the difference trains people to
    # ignore it — which is the failure mode this whole script is written against.
    code = [p for p in files if p.endswith(CODE_EXT) and not p.startswith("docs/")]
    if not code:
        return
    # Strip comments before matching. Excluding docs FILES was not enough: a Go
    # comment explaining the invariant ("// Dedup is NOT waived by this flag —
    # idx_swi_dedup still refuses...") tripped this on f6e3f3166, a commit that
    # never touched the index. Naming a rule is not changing it — the same
    # distinction the docs exclusion already makes, one level down.
    joined = "\n".join(strip_comments(changed_hunk_text(p, ref) or "") for p in code)
    for a, b, why in DECLARED_PAIRS:
        has_a, has_b = re.search(a, joined), re.search(b, joined)
        if bool(has_a) != bool(has_b):
            present, missing = (a, b) if has_a else (b, a)
            findings.append((
                "unpaired-change", "(commit)",
                f"touches {BOLD}{present}{RESET} but not {BOLD}{missing}{RESET}", why,
            ))


# ── append-only doc integrity ───────────────────────────────────────────────
# Two docs in this repo are APPEND-ONLY by owner directive, and both have a
# documented incident behind the rule:
#   SUMMARY_*.md          each is a snapshot; the SERIES is the record. A
#                         same-day second summary takes a b-suffix, it does not
#                         replace the first. (Broken 2026-07-20 by the
#                         reasoning-dataset thread, which rewrote that morning's
#                         snapshot in place; restored from 4b8d2bca0.)
#   README_where_we_are   the owner's log. Append; never rewrite or reorder.
#                         (Broken 2026-07-19 by a session that mistook it for a
#                         stray file and overwrote it.)
# Both fire on DELETIONS, because an append has none. Thresholds measured over
# 300 commits: SUMMARY at >=20 deleted lines fires on 2.0% (the same bar the
# twin/gofmt checks were held to); README on ANY deletion fires on 0.7%. A
# plain "SUMMARY was modified" predicate was rejected at 4.3% — it fires on
# legitimate appends, which is how a check teaches people to ignore it.
SUMMARY_DELETION_FLOOR = 20

# Fleet-wide append-only ledgers: many threads, one file, no owner. Guarded with a
# floor of 1 (unlike SUMMARY) because there is no legitimate small edit — an entry
# is either appended or it is somebody else's.
FLEET_APPEND_ONLY = {"WRONG_CALLS.md", "LANDMINES.md"}


def _deleted_lines(path, ref):
    return sum(
        1 for ln in raw_diff(path, ref).splitlines()
        if ln.startswith("-") and not ln.startswith("---")
    )


# Migration idempotency (bugs_open/007, "Class C"). A migration whose INSERT
# into a durable table carries no guard cannot be replayed: if it is ever
# applied out of band and left unrecorded (three real events — 07-16, 07-20,
# 07-22 — and the trend was up), the runner replays it and dies on a raw 23505
# that is indistinguishable from broken SQL. 151 did exactly this and blocked
# the runner for 3 days. The runner's own dry run already warns
# (lint_idempotency in run-migrations.sh — SAME semantics, keep them in step),
# but only after commit, at whoever runs it next; this fires at the moment the
# author can still add the guard. Same allowlist: append-only log tables where
# a duplicate insert is harmless. Measured against the corpus 2026-07-25:
# 6 true hits in ~95 migrations >=124, no false fires (see bugs_open/007).
MIGRATION_DIR = "docs/agent_docs/sql_for_agents/"
MIGRATION_NAME = re.compile(r"^\d{3}_[a-z0-9_]+\.sql$")   # sidecars (_ROLLBACK etc.) excluded
IDEMPOTENT_SINKS = re.compile(r"(^|\.)(doc_notes|doc_plans|schema_migrations)$", re.I)


def check_unguarded_migration_insert(files, ref, findings):
    """bugs_open/007 Class C — a bare-INSERT migration halts the runner on replay."""
    for path in files:
        if not path.startswith(MIGRATION_DIR):
            continue
        if not MIGRATION_NAME.match(os.path.basename(path)):
            continue
        flat = strip_comments(file_content(path, ref))
        if re.search(r"ON\s+CONFLICT|WHERE\s+NOT\s+EXISTS", flat, re.I):
            continue          # any guard anywhere exempts the file — author is handling it
        if re.search(r"DO\s+\$", flat):
            continue
        risky = sorted({t for t in re.findall(r"INSERT\s+INTO\s+([A-Za-z_][A-Za-z0-9_.]*)", flat, re.I)
                        if not IDEMPOTENT_SINKS.search(t)})
        if risky:
            findings.append((
                "unguarded-migration-insert", path,
                f"INSERT into {BOLD}{', '.join(risky)}{RESET} with no ON CONFLICT / WHERE NOT EXISTS / DO $$ guard",
                "bugs_open/007 Class C — applied out of band and left unrecorded, this file's "
                "replay dies on a raw 23505 that reads as broken SQL (151 blocked the runner "
                "3 days). Guard the INSERT so a replay is a no-op; if the duplicate SHOULD "
                "fail loudly, say so in a comment and carry on.",
            ))


def check_append_only_docs(files, ref, findings):
    """Owner directive — SUMMARY snapshots and README_where_we_are are append-only."""
    for path in files:
        base = os.path.basename(path)
        if base.startswith("SUMMARY_") and path.endswith(".md"):
            deleted = _deleted_lines(path, ref)
            if deleted >= SUMMARY_DELETION_FLOOR:
                findings.append((
                    "summary-overwritten", path,
                    f"{deleted} lines removed from an existing {BOLD}SUMMARY{RESET} snapshot",
                    "Summaries are snapshots and the series is the record — write a NEW file "
                    "(b-suffix for a second on the same day) rather than editing the last one. "
                    "If this IS a deliberate restoration or a correction, say so in the message "
                    "and carry on; this never blocks.",
                ))
        elif base == "README_where_we_are.md":
            deleted = _deleted_lines(path, ref)
            if deleted > 0:
                findings.append((
                    "readme-not-appended", path,
                    f"{deleted} line(s) removed from {BOLD}the owner's log{RESET}",
                    "README_where_we_are.md is append-only: never rewrite, reorder, or edit the "
                    "owner's words — add a dated correction below instead. A session overwrote "
                    "one on 2026-07-19 after mistaking it for a stray file.",
                ))
        elif base in FLEET_APPEND_ONLY:
            # These two are the fleet-wide shared ledgers: EVERY thread appends to
            # them, so a deletion is almost never this thread's to make, and the
            # loss is silent — the next reader cannot tell a removed entry from one
            # that was never written. CLAUDE.md declares both append-only; until
            # 2026-07-29 nothing checked either, while SUMMARY/README (single-author
            # files, lower concurrency) were both guarded.
            deleted = _deleted_lines(path, ref)
            if deleted > 0:
                findings.append((
                    "shared-ledger-not-appended", path,
                    f"{deleted} line(s) removed from {BOLD}{base}{RESET}, a fleet-wide append-only ledger",
                    f"{base} is appended to by every thread — removed lines are most likely "
                    "another session's entry, and nothing downstream can tell a deleted entry "
                    "from one never written. Append below instead; correct in place with a "
                    "dated note rather than a rewrite. If this IS a deliberate consolidation, "
                    "say so in the commit message and carry on; this never blocks.",
                ))


# ── truncation tolerance with no reader (bugs_closed/076, residual R1) ──────
# A step may set `tolerate_truncation: true` to KEEP an LLM response the model cut
# at max_tokens instead of failing. That is only sound while some other step in the
# same workflow READS the `__truncated` marker — otherwise the step succeeds and a
# fragment is indistinguishable from a complete answer. The shipped guard enforces
# it at RUN time, which means the bad config is only discovered in production, by
# failing that run. This is the same predicate applied to a seed, at the moment the
# author can still fix it for free.
#
# SCOPED TO FILES THAT CAN ACTUALLY ANSWER, and this is the whole design. All three
# files in the repo that arm the flag today (sql_for_agents/177, and the two
# fixloop PATCH_ files) are `jsonb_set` patches: they name the flag and the target
# steps, but the WORKFLOW they patch lives in the database, so nothing in the file
# says whether a reader exists — and on all three the guess would be WRONG (their
# targets are guarded by diagnose_council_decide). So this only reads a `"steps"`
# object EMBEDDED in the file, where the answer is present. Measured 2026-07-26:
# 170 SQL files under docs/ embed a steps object, 62 of them contain
# execute_llm_prompt, and ZERO arm tolerance — the check fires 0 times on the whole
# corpus, and its true-positive case is the bug's next instance.
#
# The live-fleet half is 103_LINT_truncation_consumer.py, which resolves the real
# workflow and therefore catches the patch-style and hand-run-UPDATE paths this
# cannot. Both read the reader registry out of the Go source; neither holds a copy.
STEPS_OBJECT = re.compile(r'"steps"\s*:\s*\{')
LLM_ACTION = "execute_llm_prompt"


def _balanced_object(text, start):
    """The JSON object beginning at text[start] == '{', string-aware. Returns None
    if it never closes (a truncated or non-JSON block — skipped, never guessed at)."""
    depth, in_str, esc = 0, False, False
    for i in range(start, len(text)):
        ch = text[i]
        if in_str:
            if esc:
                esc = False
            elif ch == "\\":
                esc = True
            elif ch == '"':
                in_str = False
        elif ch == '"':
            in_str = True
        elif ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                return text[start:i + 1]
    return None


def _is_true(v):
    """Mirrors datahelpers.GetBoolField: a real bool and nothing else. A string
    "true" is read as FALSE by Go, so such a step is not tolerating anything."""
    return v is True


def check_truncation_without_reader(files, ref, findings):
    """bugs_closed/076 — a seeded workflow that tolerates a cut response, unread."""
    sql = [p for p in files if p.endswith(".sql")]
    if not sql:
        return
    # Imported lazily and inside the per-check try/except in main(): if the Go
    # registry ever becomes unparseable this check announces itself as skipped
    # rather than silently checking against a remembered list.
    sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
    from truncation_registry import accepts_truncated_key, truncation_aware_actions
    readers, hatch = set(truncation_aware_actions()), accepts_truncated_key()

    for path in sql:
        content = file_content(path, ref)
        if "tolerate_truncation" not in content:
            continue
        for m in STEPS_OBJECT.finditer(content):
            block = _balanced_object(content, m.end() - 1)
            if not block or "tolerate_truncation" not in block:
                continue
            try:                                   # the JSON sits in a SQL literal
                steps = json.loads(block.replace("''", "'"))
            except (json.JSONDecodeError, ValueError):
                continue                           # unparseable: 103_LINT still sees it live
            if not isinstance(steps, dict):
                continue
            guards = {
                n for n, s in steps.items()
                if isinstance(s, dict) and (s.get("action") in readers
                                            or _is_true((s.get("config") or {}).get(hatch)))
            }
            for name, step in steps.items():
                if not isinstance(step, dict) or step.get("action") != LLM_ACTION:
                    continue
                if not _is_true((step.get("config") or {}).get("tolerate_truncation")):
                    continue
                if guards - {name}:                # a step cannot certify its own cut
                    continue
                findings.append((
                    "truncation-tolerance-no-reader", path,
                    f"step {BOLD}{name}{RESET} sets tolerate_truncation, but no step in that "
                    f"workflow reads the __truncated marker",
                    "bugs_closed/076 — the step then SUCCEEDS on a fragment and nothing "
                    "downstream can tell it from a complete answer. The runtime guard will "
                    f"fail this step's run when a response is actually cut. Fix one of: raise "
                    f"max_tokens; drop tolerate_truncation; or set {hatch}: true on the step "
                    "that genuinely handles a partial. Patch-style seeds (jsonb_set) are not "
                    "checked here — run 103_LINT_truncation_consumer.py against the live fleet.",
                ))




# Surfaces a document can PROPOSE into existence: a binary, an image, a package.
# Deliberately only three, and deliberately only ones a filename can settle.
SURFACE_CMD = re.compile(r"\bcmd/([a-z0-9][a-z0-9._-]{1,40})")
SURFACE_DOCKERFILE = re.compile(r"\b([a-z0-9][a-z0-9._-]{1,40})\.dockerfile\b")
SURFACE_PKG = re.compile(r"\b(internal|platform|pkg)/([a-z0-9][a-z0-9._-]{1,40})/")
# How far back counts as "arrived after you last looked".
SURFACE_RECENT = "14 days ago"
# Repo top-level names. A doc enumerating the review scope writes "platform/,
# internal/, pkg/" and the package regex reads the second as a child of the first.
SURFACE_TOPLEVEL = {"internal", "platform", "pkg", "cmd", "docs", "scripts", "build"}


def _tree_names(prefix, ref=None):
    """Immediate child names under `prefix`, as of the commit under test."""
    rev = ref[1] if ref else "HEAD"
    out = sh("git", "ls-tree", "--name-only", f"{rev}:{prefix}")
    return {n.rstrip("/") for n in out.splitlines() if n.strip()}


def _recent_additions(prefix, ref=None):
    """Names first added under `prefix` in the recent window — one git call.

    These are the whole point: the failure this check exists for is a peer that
    arrived AFTER the author's prior-art search, so the peer list must say which
    ones are new rather than presenting a flat alphabetical wall.

    In audit mode the window is anchored to the AUDITED COMMIT's own date, not to
    today — otherwise auditing a commit from months ago labels half the tree "new",
    which is exactly the kind of confidently-wrong annotation this file exists to
    avoid producing.
    """
    rev = ref[1] if ref else "HEAD"
    since = SURFACE_RECENT
    if ref:
        stamp = sh("git", "log", "-1", "--format=%cI", rev).strip()
        if stamp:
            since = f"{stamp.split('T')[0]} -14 days"
    out = sh("git", "log", rev, f"--since={since}", "--diff-filter=A",
             "--name-only", "--format=", "--", prefix)
    names = set()
    for line in out.splitlines():
        rest = line.strip()[len(prefix):].lstrip("/")
        if rest:
            names.add(rest.split("/")[0])
    return names


def _peer_list(prefix, ref=None, limit=8):
    """'a, b (new), c …' — new-in-window first, then the rest, capped."""
    peers = _tree_names(prefix, ref)
    if not peers:
        return ""
    recent = _recent_additions(prefix, ref) & peers
    ordered = sorted(recent) + sorted(peers - recent)
    shown = [f"{BOLD}{p}{RESET} (new)" if p in recent else p for p in ordered[:limit]]
    more = len(ordered) - len(shown)
    return ", ".join(shown) + (f" … +{more} more" if more > 0 else "")


def check_new_capability_surface(files, ref, findings):
    """A DOCUMENT that proposes a binary, image or package which does not exist.

    THE INCIDENT (2026-07-27, WRONG_CALLS). A design doc committed 2026-07-24
    specified a new service `cmd/gripper-intake/` on the island VM. `cmd/tools-api`
    shipped to that same VM on 07-25 already doing all of it, multi-tool and
    multi-site. Caught on 07-26 by the OWNER asking an integration question. No
    mechanism caught it, and none could: the decision lived in markdown, and
    `097_TRIGGER_council_review_v1.sh:53` refuses docs client-side
    (`SCOPE_RE='^(platform|internal|pkg)/'`) — correctly, on cost grounds, since 72
    DESIGN/PLAN/SPEC docs were created in one month.

    WHY A GREP AND NOT A SEAT. The author's prior-art search on the 24th was
    exhaustive and CORRECT — tools-api did not exist that day. The failure class is
    a fact that was true at review time and false at build time. Every council seat
    is a one-shot evaluation against a snapshot; nothing re-validates. This check is
    free and idempotent, so it re-runs on every later commit that touches the doc —
    and on 07-25 and 07-26 it would have printed `tools-api` as a peer, newly
    arrived. That repetition is the entire value, not the first fire.

    MEASURED BEFORE INCLUSION, per this file's own bar — and re-measured against
    THIS predicate rather than carried over. Over the last 1,500 commits:
    **20 fires, 1.33%**, inside the accepted band (SUMMARY 2.0%, twin ~2%, README
    0.7%). An earlier draft measured 0.67%, but that was a cmd/-only predicate; the
    dockerfile and package clauses roughly double it, so the older figure does not
    describe what actually ships here.

    Precision, by inspection of the non-incident fires: `cmd/webdesignport`,
    `tools-api.dockerfile`, `internal/tools-api` and `cmd/assembler` were all genuine
    proposals-before-existence (the first three were subsequently built; the fourth
    never was, which is itself worth surfacing). One false-positive CLASS was found
    and removed — a doc enumerating the review scope writes "platform/, internal/,
    pkg/" and the package regex read the second as a child of the first; hence
    SURFACE_TOPLEVEL. Seven of the twenty fires are this incident's own write-up
    naming the path it tells you NOT to build, which is correct but will not recur.

    A whole-TREE scan is a different and unusable predicate: ~190 docs fire, almost
    all archived copies naming the long-retired `cmd/bundle`. Staged ADDED lines only.

    DELIBERATELY NOT CHECKED: new compose services (compose files live under docs/,
    and a service added inside an existing file is invisible to --name-only) and new
    route prefixes (gin, gorilla and stdlib idioms are all live in this tree at once).
    Both were dropped as unwritable rather than approximated.
    """
    docs = [f for f in files if f.endswith(".md")]
    if not docs:
        return

    for path in docs:
        added = "\n".join(l for l in raw_diff(path, ref).splitlines()
                          if l.startswith("+") and not l.startswith("+++"))
        if not added:
            continue

        # Prose punctuation rides along with a path ("build `cmd/foo`." → "foo.").
        # Strip it, or the check reports a surface nobody proposed.
        def _clean(n):
            return n.strip("._-")

        proposed = []  # (surface, prefix_for_peers)
        for name in {_clean(n) for n in SURFACE_CMD.findall(added)}:
            if name and name not in _tree_names("cmd", ref):
                proposed.append((f"cmd/{name}/", "cmd"))
        for name in {_clean(n) for n in SURFACE_DOCKERFILE.findall(added)}:
            if name and f"{name}.dockerfile" not in _tree_names("build/docker/backend", ref):
                proposed.append((f"{name}.dockerfile", "build/docker/backend"))
        for top, name in {(t, _clean(n)) for t, n in SURFACE_PKG.findall(added)}:
            # "platform/internal/pkg" in prose enumerating the review scope is not a
            # proposal — measured false positive, 2 of 6 non-incident fires.
            if name in SURFACE_TOPLEVEL:
                continue
            if name and name not in _tree_names(top, ref):
                proposed.append((f"{top}/{name}/", top))

        for surface, prefix in sorted(set(proposed)):
            peers = _peer_list(prefix, ref)
            findings.append((
                "new-capability-surface", path,
                f"proposes {BOLD}{surface}{RESET}, which does not exist",
                f"existing {prefix}/: {peers or '(none)'}\n"
                "     If one of these already does this, say in the doc why it does not — "
                "a second copy of a service, image or package is the shape that forked the "
                "VM estate. Peers marked (new) landed in the last fortnight and may postdate "
                "your prior-art search. Renaming something, or naming a path you have "
                "deliberately decided AGAINST, fires this too; that is expected and it never "
                "blocks.",
            ))


# A log sink: anything that writes to stdout/stderr or a logger.
LOG_SINK = re.compile(
    r"\b(?:log\.(?:Printf|Println|Print|Fatalf|Fatal)"
    r"|logger\.(?:Info|Warn|Error|Debug|Infof|Warnf|Errorf)"
    r"|fmt\.(?:Printf|Println|Fprintf))\s*\(")
# Identifiers that carry a model/user payload rather than a fact about one.
PAYLOAD_NAME = re.compile(r"\b(text|body|completion|response|raw|content|output|prompt|partial)\b", re.I)
# Wrapping the value in one of these emits a derived FACT, not the content.
SAFE_WRAPPERS = ("Fingerprint(", "TopLevelJSONObjects(", "len(", "IsTruncated(",
                 "utf8.RuneCountInString(", "Itoa(")


def check_register_coverage(files, ref, findings):
    """A commit that CREATES a workstream directory the concept register has never heard of.

    THE BUG (bugs_open/106). Concept-register extraction froze 2026-07-13. Three
    whole subsystems were later found missing — fixloop (07-16), model-directory
    (07-17), claims-verification (07-27) — and ALL THREE were found by coincidence,
    because somebody happened to be working beside the hole. The register is the
    instrument sessions are told to consult BEFORE concluding something does not
    exist, so a hole in it reads as "this does not exist" rather than "nobody
    looked". On 2026-07-26 a session concluded exactly that, wrote it into a live
    council seat's standing instructions, and was one step from building a
    redundant subsystem.

    The SENSOR already exists and works —
    `docs/agent_docs/docs026_concept_register/102_CHECK_register_coverage.py`,
    sensor + ratchet, built the same day 106 was filed. What it never got was a
    CADENCE: grep found it referenced in exactly two places, both its own
    documentation. It runs when a human remembers, which is the same
    detected-by-coincidence mechanism the bug is about, moved one step earlier.
    A fourth tool that must be invoked by coincidence does not retire it.

    WHY THIS TRIGGER AND NOT A PERIODIC SWEEP. A cron would report drift up to a
    week after it appeared, to nobody in particular. This fires at the moment the
    gap is CREATED, in front of the person creating it, who is the one person who
    can close it in ten seconds by adding a register line. Put the check where the
    error is made.

    WHY IT IMPORTS THE SENSOR RATHER THAN REIMPLEMENTING is_covered(). Two
    hand-maintained copies of one matching rule is precisely the drift class this
    platform keeps filing bugs about (idx_swi_dedup / workItemTerminalStatuses is
    the standing example). There is one implementation; this calls it.

    ONLY NEW DIRECTORIES FIRE. 43 existing workstreams are uncovered and on the
    ratchet — that is accepted backlog, and flagging active work on them every
    commit is how a check becomes wallpaper.

    MEASURED BEFORE INCLUSION, per this file's bar: see the fire rate recorded in
    bugs_closed/106.
    """
    ws_root = "docs/agent_docs/docs024_key_docs_latest/"
    base = ref[0] if ref else "HEAD"

    # Which workstream dirs does this commit touch?
    touched = set()
    for f in files:
        if not f.startswith(ws_root):
            continue
        rest = f[len(ws_root):]
        if "/" not in rest:
            continue                      # a loose file at the root, not a workstream
        touched.add(rest.split("/", 1)[0])
    if not touched:
        return

    # Of those, which are NEW — i.e. have no history before this commit?
    fresh = [d for d in sorted(touched)
             if not sh("git", "log", "--oneline", "-1", base, "--", ws_root + d).strip()]
    if not fresh:
        return

    # Reuse the sensor's own matching rule; never re-implement it here.
    import importlib.util
    repo = sh("git", "rev-parse", "--show-toplevel").strip()
    sensor_path = os.path.join(
        repo, "docs/agent_docs/docs026_concept_register/102_CHECK_register_coverage.py")
    if not os.path.exists(sensor_path):
        return                              # sensor removed or moved — stay silent, never break a commit
    spec = importlib.util.spec_from_file_location("register_coverage", sensor_path)
    sensor = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(sensor)          # read-only; __main__ guard keeps it quiet

    corpus = sensor.register_text()
    ratchet = sensor.read_ratchet()
    for d in fresh:
        if d in ratchet or sensor.is_covered(d, corpus):
            continue
        findings.append((
            "register-blind-spot", ws_root + d,
            f"new workstream {BOLD}{d}{RESET} — the concept register has never heard of it",
            "The register is what sessions consult before concluding a capability does not "
            "exist, so an unlisted subsystem reads as absent rather than unlooked-for "
            "(bugs_open/106: three subsystems went missing that way, each found by "
            "coincidence). If this directory will hold a reusable mechanism, add an entry "
            "to docs026_concept_register/register/<category>.md; if it is site content or a "
            "one-off, add it to 102_coverage_ratchet.txt and it stays quiet. Advisory.",
        ))


def check_logged_model_output(files, ref, findings):
    """bugs_open/083 + council corr e004fd81 — logging an LLM response verbatim.

    An LLM response echoes its prompt back, and on a debate, chat or support
    endpoint that prompt contains what the VISITOR wrote. Logging the response
    therefore publishes user-derived text to anyone who can read the container's
    logs. That is a content decision, not a code-review one — it took a whole
    council round to surface once (guardian, 2026-07-27: "cannot be closed by
    this council alone").

    It is usually the WEAKER diagnostic too. Every question an unusable
    completion raises is structural — prose wrapper? markdown fence? two
    objects? empty? — and a capped excerpt answers them worse than a fingerprint
    does: bugs_closed/088's second JSON object begins ~1,500 chars in, past any
    sane cap, so an excerpt cannot see the very case it exists for.

    Gated on the PACKAGE, not the file: the LLM call and the log sink usually
    live in sibling files (handlers/defend.go calls GenerateText, handlers/ailog.go
    does the logging). A file-level gate makes this check silently vacuous — the
    first version of it was, and only a positive control caught that.
    """
    go = [f for f in files if f.endswith(".go") and not f.endswith("_test.go")]
    if not go:
        return

    # Which directories are LLM-adjacent? Ask git, not the changed set, so a
    # commit touching only the logging file is still checked.
    llm_dirs = set()
    for path in go:
        d = os.path.dirname(path)
        if d in llm_dirs:
            continue
        for sibling in sh("git", "ls-files", os.path.join(d, "*.go")).splitlines():
            if "GenerateText" in file_content(sibling, ref):
                llm_dirs.add(d)
                break

    for path in go:
        if os.path.dirname(path) not in llm_dirs:
            continue
        lines = file_content(path, ref).split("\n")
        for i, line in enumerate(lines):
            if not LOG_SINK.search(line):
                continue
            stmt, depth = "", 0
            for l in lines[i:i + 6]:               # a log call often wraps
                stmt += l
                depth += l.count("(") - l.count(")")
                if depth <= 0:
                    break

            probe = stmt
            for w in SAFE_WRAPPERS:                # drop derived-fact calls
                probe = re.sub(re.escape(w) + r"[^()]*\)", "", probe)

            m = PAYLOAD_NAME.search(probe.split(",", 1)[-1])   # skip the format string
            if not m:
                continue
            findings.append((
                "logged-model-output", f"{path}:{i + 1}",
                f"log call passes {BOLD}{m.group(0)}{RESET} unwrapped, in a package that calls GenerateText",
                "an LLM response echoes the prompt back, so this can publish what the VISITOR "
                "wrote; prefer aiservice.Fingerprint(v) — it answers the shape questions without "
                "the text, and catches the two-object case a capped excerpt structurally cannot",
            ))
            break                                   # one finding per file is enough to act on

def main():
    ref = None
    if "--commit" in sys.argv:                      # audit ONE commit in isolation
        c = sys.argv[sys.argv.index("--commit") + 1]
        ref = (c + "~1", c)
    elif "--ref" in sys.argv:                       # audit a range base..HEAD
        ref = (sys.argv[sys.argv.index("--ref") + 1], "HEAD")
    files = changed_files(ref)
    if not files:
        return 0

    findings = []
    for check in (check_untouched_twin, check_gofmt, check_stdin_eater, check_declared_pairs,
                  check_unguarded_migration_insert, check_append_only_docs,
                  check_truncation_without_reader, check_logged_model_output,
                  check_new_capability_surface, check_register_coverage):
        try:
            check(files, ref, findings)
        except Exception as e:  # never let a check break a commit
            print(f"{DIM}pattern-check: {check.__name__} skipped ({e}){RESET}", file=sys.stderr)

    if not findings:
        return 0

    print(f"\n{YELLOW}── pattern check: {len(findings)} thing(s) worth a look ──{RESET}")
    for kind, where, what, why in findings:
        print(f"   {BOLD}{kind}{RESET}  {where}")
        print(f"     {what}")
        print(f"     {DIM}{why}{RESET}")
    print(f"   {DIM}Advisory — this never blocks. PATTERN_CHECK_STRICT=1 makes it exit non-zero.{RESET}")
    print(f"   {DIM}Full patterns: docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md section 9{RESET}\n")

    return 1 if os.environ.get("PATTERN_CHECK_STRICT") == "1" else 0


if __name__ == "__main__":
    sys.exit(main())
