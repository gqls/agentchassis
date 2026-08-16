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
        # COMMENTS ARE STRIPPED FOR THE OFFENCE SEARCH ONLY, AND WITH A SHELL-SPECIFIC
        # STRIPPER (added 2026-07-31). Two bugs were fixed here, the second introduced
        # while fixing the first:
        #
        #  1. Without any stripping the rule fires on a comment that WARNS about the
        #     trap. CHECK_naming_contract.sh carries "do NOT `while read` over a
        #     here-string here" directly above the mapfile+for loop that is the correct
        #     fix, and this check reported that file as the defect. A detector that
        #     flags its own warning text teaches people to ignore the whole script.
        #
        #  2. The shared strip_comments() CANNOT be used here, and its docstring's
        #     claim that it "can only ever suppress a finding, never invent one" does
        #     NOT hold for this check. It treats `--` as a comment start, and `--` is
        #     kubectl's argument separator, so
        #       kubectl exec -i pod -- psql -c '…' </dev/null
        #     strips down to "kubectl exec -i pod" and THE GUARD DISAPPEARS. A
        #     correctly-guarded loop then gets flagged. The monotonicity claim holds
        #     only for checks that search the stripped text for the OFFENCE; a check
        #     that searches it for a GUARD inverts the direction.
        #
        # Hence: `#`-only stripping (and only where `#` starts a word, so `${row#tool-}`
        # survives), the offence searched in the stripped body, and the guard searched
        # in the RAW body. Line count is preserved either way, so reported line numbers
        # are unaffected. Both directions are covered by controls — a genuine unguarded
        # eater must still fire, and a guarded one must not.
        raw_lines = content.splitlines()
        lines = [SH_COMMENT.sub("", l) for l in raw_lines]
        depth_stack = []
        for i, line in enumerate(lines):
            if re.search(r"\bwhile\b.*\bread\b", line):
                depth_stack.append(i)
            if re.match(r"^\s*done\b", line) and depth_stack:
                start = depth_stack.pop()
                body = "\n".join(lines[start:i])              # comment-free: the OFFENCE
                body_raw = "\n".join(raw_lines[start:i])       # verbatim: the GUARD
                m = STDIN_EATERS.search(body)
                if m and "</dev/null" not in body_raw and "< /dev/null" not in body_raw:
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

# Shell-only comment stripper. Deliberately NOT the COMMENT regex above: `--` is
# kubectl's argument separator, not a shell comment, and treating it as one deletes
# the `</dev/null` guard that check_stdin_eater looks for. `(?<!\S)` keeps
# `${row#tool-}` and `"a#b"` intact by requiring whitespace or line-start before `#`.
SH_COMMENT = re.compile(r"(?<!\S)#.*$")


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


# ── runtime-fill marker: one predicate, or say which scope you meant ────────
#
# bugs_open/137. The exemption "this control is exempt because it hydrates
# client-side" was written NINE times as a bare test against whatever the caller
# passed — so its blast radius followed caller chunking, not the markup: a
# section-shaped input made it right, a page-shaped input made it exempt every
# unrelated section. No test pinned it because at each site the line reads as
# obviously correct.
#
# The reason a tenth copy was free is that adding one told nobody. This fires at
# the moment the author can still choose, and it does NOT judge which scope is
# right — only the author can. It requires a NAMED predicate so the intent is
# visible in review:
#   HasRuntimeFillMarker(html)  "is this SECTION a shell?"      (whole input)
#   RuntimeFillSpans(html)      "is this CONTROL alive?"        (per element)
#   InRuntimeFillShell(sel)     the same, for a goquery selection
#
# ALLOWLIST = sites deliberately left raw, each with its reason. An entry is a
# decision on the record; a gate with no escape hatch only pressures the next
# author into weakening the pattern. Two are WRITERS (they act on what they see,
# so a wide exemption is fail-safe), the rest ask the section question per row.
#
# Matches the SQL spelling too, because four of the copies are SQL strings
# embedded in Go — a Go-only pattern reported the tree clean while they sat in
# it, which is this bug's own defect one level up.
RUNTIME_FILL_MARKER_RE = re.compile(
    r'(?i)((strings\.)?(Contains|HasPrefix|HasSuffix|Index)\s*\([^)]*"data-runtime-fill"'
    r"|regexp\.MustCompile\s*\(\s*[`\"][^`\"]*data-runtime-fill"
    r"|LIKE\s+'%data-runtime-fill%')")

RUNTIME_FILL_OWNER = "platform/orchestration/datahelpers/runtime_fill.go"

RUNTIME_FILL_ALLOWED = {
    "rerender_single_page_action.go":
        "section question; the tree's only (?i) test — normalising it would silently "
        "change the page assembler (bugs_open/137)",
    "check_empty_sections.go":
        "section question, per component (Go verdict + its SQL twin)",
    "render_site_components_action.go":
        "WRITER — DropDeadURLControls removes the control, so a wide exemption is "
        "fail-safe; shared chrome, so the safest edit is none",
    "check_required_fields_missing.go":
        "SQL, per row: is this component a shell, so missing fields are by design?",
    "check_component_standards.go":
        "SQL, per row: is this template a shell, so '<no value>' is the mechanism?",
    "check_component_template_corrupted.go":
        "SQL, per row: is this template a shell, so build-time emptiness is intended?",
    "rendercheck.go":
        "named predicate componentIsRuntimeFillShell with the scope stated beside "
        "it — the input is ONE component's template, never a page, so the "
        "containment test cannot exempt an unrelated section (140 plan item 1). "
        "The file is deliberately NOT main.go: this map keys on basename, and "
        "allowing 'main.go' would exempt every cmd/ tool at once",
}


def check_runtime_fill_marker(files, ref, findings):
    """bugs_open/137 — a raw marker test has no scope in its name; name the predicate."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        if path == RUNTIME_FILL_OWNER:
            continue          # the predicate's own file owns the literal
        if os.path.basename(path) in RUNTIME_FILL_ALLOWED:
            continue
        content = file_content(path, ref)
        if not content:
            continue
        for i, line in enumerate(strip_comments(content).splitlines(), 1):
            if RUNTIME_FILL_MARKER_RE.search(line):
                findings.append((
                    "runtime-fill-scope", f"{path}:{i}",
                    f"raw {BOLD}data-runtime-fill{RESET} test — its scope is whatever the caller passed",
                    "bugs_open/137 — this exemption was written nine times as a bare "
                    "string test, so a page-shaped input exempted every unrelated section "
                    "and nothing recorded it. Use a NAMED predicate and say why beside it: "
                    "HasRuntimeFillMarker (is this SECTION a shell?), RuntimeFillSpans / "
                    "InRuntimeFillShell (is this CONTROL alive?). If it must stay raw, add "
                    "it to RUNTIME_FILL_ALLOWED with the reason.",
                ))


# An item_type CONSTRUCTED at runtime is a label no consumer can know
# (bugs_open/279): it inserts into site_work_items cleanly, reports success, and
# nothing ever claims it — the rows die open (bugs_open/115: 100% of one
# auditor's output, for weeks). The work-item vocabulary lives only in Go source,
# so the ban is on construction itself. strip_comments is safe here per its own
# docstring rule: this check searches ONLY for the offence, so stripping can only
# ever suppress, never manufacture.
#
# This is the commit-time ADVISORY half; the BLOCKING half is
# TestNoDynamicallyConstructedItemTypes (platform/orchestration/actions/
# work_item_type_minting_ratchet_test.go), same pattern with its own
# must-match/must-not-match self-test. Change the two patterns TOGETHER.
DYNAMIC_ITEM_TYPE_RE = re.compile(r'\b[Ii]temType\s*[:=][^,\n]*("\s*\+|\+\s*"|fmt\.Sprintf)')


def check_dynamic_item_type(files, ref, findings):
    """bugs_open/279 — a constructed item_type is a label nothing consumes."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        content = file_content(path, ref)
        if not content:
            continue
        for i, line in enumerate(strip_comments(content).splitlines(), 1):
            if DYNAMIC_ITEM_TYPE_RE.search(line):
                findings.append((
                    "dynamic-item-type", f"{path}:{i}",
                    f"work-item {BOLD}item_type{RESET} built from string parts — a label no consumer knows",
                    "bugs_open/279 — the only construction site ever found "
                    "(write_audit_findings' audit_finding_ + category) filed items that "
                    "died open for weeks because no verifier, workflow or handler names a "
                    "constructed label. Use a literal from the routing vocabulary, or file "
                    "the finding as capability_gap (the platform's 'work I have no handler "
                    "for' shape, bugs_closed/077). The blocking twin of this advisory is "
                    "TestNoDynamicallyConstructedItemTypes — change both patterns together.",
                ))


# A writer of page_components.rendered_html with no link repair (bugs_open/136).
#
# bugs_open/079 put the dead-internal-link repair at the full-page section save,
# which is where LLM-authored body prose normally enters. The council's
# bug_historian seat objected that this was never shown to be the ONLY writer of
# the column — only the only one with a 'save_page_sections' STEP NAME — and the
# check it asked for found four more. Between the filing and the fix a FIFTH
# appeared (adopt_verbatim.go), which is the argument for a mechanical check
# rather than a list in a bug file: the set grows, and it grows in files nobody
# is thinking about link repair in.
#
# This does NOT assert that every writer must repair. Two of them must not —
# byte-preserving adoption would be corrupted by it. It asserts that a writer
# either repairs or is NAMED here with the reason, so the decision is visible at
# the moment of the edit instead of being rediscovered by a council round.
COMPONENT_HTML_WRITE_RE = re.compile(
    r"(?is)(INSERT\s+INTO\s+page_components\b[^`\"]{0,400}?rendered_html"
    r"|UPDATE\s+page_components\b[^`\"]{0,400}?SET[^`\"]{0,400}?rendered_html\s*=)")

# The GUARD is searched for as a CALL on a NON-COMMENT line of the RAW body, and
# both halves of that are a landmine (LANDMINES.md, 2026-07-31, strip_comments):
#
#   - not the stripped body, because strip_comments can DELETE a guard and
#     manufacture a finding. Its docstring's "can only ever suppress" holds for a
#     check searching for the OFFENCE; this one also searches for the guard.
#   - not the raw body either, because a comment merely MENTIONING the seam would
#     then silence the check on a genuinely unguarded writer — control (c) of that
#     landmine's own four, and the one that catches a detector narrowed into
#     inertness. Hence: a real call, on a line that is not a comment.
COMPONENT_REPAIR_SEAM_RE = re.compile(
    r"(?:repairComponentHTMLBeforePersist|repairSectionsBeforePersist|repairOutboundPageLinks)\s*\(")


def _calls_repair_seam(raw):
    for line in raw.splitlines():
        bare = line.lstrip()
        if bare.startswith("//") or bare.startswith("*"):
            continue                      # prose about the seam is not a call to it
        if COMPONENT_REPAIR_SEAM_RE.search(line):
            return True
    return False

# Reasons, not exemptions. Each says why repairing this writer would be WRONG —
# not that it was inconvenient. A writer whose reason is "not done yet" belongs
# in the bug file and in this check's output, NOT in here: an allow-list that
# absorbs the open cases silences the detector on exactly what it was written to
# catch (the tool-markup writers are deliberately absent for that reason).
COMPONENT_WRITE_ALLOWED = {
    "adopt_verbatim.go":
        "byte-preserving adoption — MEASURED, not assumed (council editquality seat, "
        "corr 0275f9c2, asked for the citation): it writes content.RawHTML verbatim "
        "(:514, :533) and stores sha256(RawHTML) in content_data (:487), so a repair "
        "would silently invalidate the hash the file exists to keep. It is reachable "
        "ONLY under a strict binary — apply_adoption_plan_action.go:426, "
        "`if fidelity := adoptionFidelity(...); fidelity == fidelityLocked` — so there "
        "is no milder mode in which this writer emits LLM-authored prose "
        "(cf. the '--fidelity high is not a milder locked' landmine, which is about "
        "which PATH runs, not about what this file writes)",
    "import.go":
        "one-off port CLI, same byte-preservation reasoning as adopt_verbatim",
    "fix_harcoded_colours_action.go":
        "colour-only rewrite of existing html — cannot introduce an href",
    "fix_forced_text_colours_action.go":
        "colour-only rewrite of existing html — cannot introduce an href",
    "rebuild_blog_listing_action.go":
        "hrefs come from pages.url (blogPostsQuery -> articles[].url) and the live "
        "content-listing template carries exactly one anchor, href=\"{{.url}}\" — the "
        "same table the repair index is built from, so repair could only no-op "
        "(measured 2026-08-02)",
    "page_admin_handlers.go":
        "human-driven admin API — a person edited this deliberately",
}


def check_unrepaired_component_write(files, ref, findings):
    """bugs_open/136 — a writer of rendered_html that never repairs its links."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        if os.path.basename(path) in COMPONENT_WRITE_ALLOWED:
            continue
        content = file_content(path, ref)
        if not content:
            continue
        # OFFENCE in the stripped body (a comment describing an UPDATE is not one),
        # GUARD in the raw body via _calls_repair_seam — see its note above.
        if not COMPONENT_HTML_WRITE_RE.search(strip_comments(content)):
            continue
        if _calls_repair_seam(content):
            continue
        findings.append((
            "unrepaired-component-write", path,
            f"writes {BOLD}page_components.rendered_html{RESET} with no dead-internal-link repair",
            "bugs_open/136 — 079's repair guards the full-page section save; a sibling "
            "writer that persists rendered_html on its own bypasses it, and an invented "
            "/pricing ships as a 404 with a green status. Call "
            "repairComponentHTMLBeforePersist(...) immediately before the write (it wraps "
            "the same seam, fail-open, same CONTENT_LINK_REPAIR_DETAIL record). If repair "
            "would be WRONG for this writer, add it to COMPONENT_WRITE_ALLOWED with the "
            "reason — and if the honest reason is 'not yet', leave it here and say so in "
            "the bug file instead.",
        ))


# ── a component template rendered with no per-instance token bound ──────────
#
# bugs_open/283. A component's element ids are namespaced by {{.InstanceID}}, and
# a render path that never binds one gets missingkey=zero — an empty string — so
# every instance on the page lands back on IDENTICAL ids. The failure is silent
# and plausible: the second calculator renders, accepts typing, responds to its
# button, and answers from the FIRST one's fields.
#
# Why a check rather than a list in the bug file: measured 2026-08-16, EIGHT
# non-test files hold FOURTEEN call sites. The council's bug_historian seat named
# five files it believed were call sites; four of them do not call these helpers
# at all, and the two that were the actual defect (the section editor's) were on
# nobody's list. My own first census missed cmd/component-render-check entirely,
# because it grepped platform/ and internal/ and the file lives in cmd/ — the
# broad sweep below is what found it. An enumeration written by reading is
# exactly what goes stale; same shape as the rendered_html writer census above.
#
# It matches the CALL, not the argument's name. An earlier version required the
# template argument to be spelled HTMLTemplate/htmlTemplate/headTemplate, which
# is the same staleness one rung down: a new call site passing `tpl` would render
# an unnamespaced page and the check would say nothing. Measured 2026-08-16, the
# broad form costs nothing — exactly eight non-test files call any RenderTemplate*
# helper, and every one is either bound or allow-listed below.
COMPONENT_RENDER_RE = re.compile(r"\bRenderTemplate\w*\s*\(")

# Searched as a CALL on a non-comment line of the RAW body, for both the reasons
# in _calls_repair_seam's note above: the stripped body can DELETE a guard and
# manufacture a finding, and the raw body lets a comment merely MENTIONING the
# seam silence a genuinely unbound writer.
INSTANCE_BIND_SEAM_RE = re.compile(
    r"(?:BindInstanceToken|BindSingleSectionInstanceToken)\s*\(")


def _binds_instance_token(raw):
    for line in raw.splitlines():
        bare = line.lstrip()
        if bare.startswith("//") or bare.startswith("*"):
            continue                      # prose about the seam is not a call to it
        if INSTANCE_BIND_SEAM_RE.search(line):
            return True
    return False


# Reasons, not exemptions — and every one here is a MEASURED claim about the
# slot, not a guess about the file. A component rendered into <head>, or into the
# site header/footer chrome, occurs once per document by construction, so there
# is no second instance to collide with. A writer whose honest reason is "not yet"
# belongs in the finding list and the bug file, not in here.
INSTANCE_TOKEN_ALLOWED = {
    "component_library.go":
        "RenderHeader/RenderFooter/RenderHead (:1943, :2012, :2285) render CHROME, "
        "resolved through ResolveChromeComponent — one header, one footer and one "
        "<head> per document, so no second instance exists to collide with. This "
        "file also HOSTS the shared reporting seam (RenderTemplateReportingMissing), "
        "which is what covers every other caller",
    "render_site_components_action.go":
        "site chrome slots (header/nav/footer), one instance per page by construction",
    "rerender_pages_actions.go":
        "renders the <head> template only (:532) — a document has exactly one",
    "rendercheck.go":
        "cmd/component-render-check is an offline LINT: it renders every active "
        "component to look for empty-element shapes and writes its report to "
        "doc_notes (:507), never to page_components.rendered_html and never to a "
        "served page, so no page can inherit an id from it. It also synthesises a "
        "unique marker for EVERY referenced field (:315-350), so it supplies its "
        "own InstanceID; binding a real token would defeat the absence arm it "
        "exists to run",
}


def check_unscoped_component_render(files, ref, findings):
    """bugs_open/283 — a component template rendered with no instance token bound."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        if os.path.basename(path) in INSTANCE_TOKEN_ALLOWED:
            continue
        content = file_content(path, ref)
        if not content:
            continue
        # OFFENCE in the stripped body (a comment quoting a render call is not
        # one), GUARD in the raw body via _binds_instance_token.
        if not COMPONENT_RENDER_RE.search(strip_comments(content)):
            continue
        if _binds_instance_token(content):
            continue
        findings.append((
            "unscoped-component-render", path,
            f"renders a component template with no {BOLD}InstanceID{RESET} bound",
            "bugs_open/283 — the template's {{.InstanceID}} then renders EMPTY under "
            "missingkey=zero and every instance on the page takes identical element "
            "ids, so getElementById hands each lookup to the first copy: the second "
            "calculator answers from the first one's inputs, with no error anywhere. "
            "Bind it before rendering — BindInstanceToken(rc, counter.Next(fn)) if you "
            "walk the whole page in order, BindSingleSectionInstanceToken(rc, fn) if "
            "you render one section and cannot see the rest. If this slot genuinely "
            "occurs once per document, add it to INSTANCE_TOKEN_ALLOWED with the "
            "measured reason.",
        ))


# ── a page upsert that names page_type in the INSERT and not in the SET ─────
#
# bugs_open/175, and bugs_closed/081 before it. The statement:
#
#   INSERT INTO pages (site_id, name, url, title, page_type, sections, ...)
#   ...
#   ON CONFLICT (site_id, name) DO UPDATE SET url = ..., title = ..., sections = ...
#
# reads as a create with an idempotent re-run. It is not. On a name collision it
# is a PARTIAL UPDATE: this arm's content lands under the EXISTING row's role, no
# error is raised, and `RETURNING id` yields an id either way, so the caller
# cannot tell which happened. 081 measured one such arm looping for three months.
#
# Five arms were written with this shape before anyone noticed, in four files, by
# different sessions — which is the definition of a class rather than a bug. The
# fix seam is UpsertPageForRole (platform/orchestration/actions/page_role_upsert.go);
# this check is what stops a SEVENTH being written, because knowing the pattern is
# not what fires it — something at the moment of the edit has to.
#
# PRECISION: it fires only when page_type is in the INSERT column list AND absent
# from the SET list. An arm that deliberately carries `page_type =
# EXCLUDED.page_type` (adoption, blog posts, site sync, verbatim adoption) is a
# DIFFERENT decision — 175 says explicitly not to make the two camps identical —
# and is not flagged. MEASURED over all 1,120 .go files at HEAD on 2026-08-02:
# exactly 4 hits before the 175 fix (create_report_page, create_tool_component,
# deploy_tool ×2 — 175's census, nothing else), 0 after, and no hit on any of the
# five arms in the other camp. The first run of this measurement was against a
# tree that already carried the fix and reported 0/0; a check that has not been
# seen to FIRE is not a check (LANDMINES — "a gate's 0 findings has two causes").
PAGE_UPSERT_RE = re.compile(
    r"INSERT\s+INTO\s+pages\b(?P<insert>[^;`]*?)"
    r"ON\s+CONFLICT\s*\(\s*site_id\s*,\s*name\s*\)\s*DO\s+UPDATE\s+SET(?P<set>[^;`]*?)"
    r"(?:RETURNING|$)",
    re.I | re.S)


def check_partial_page_upsert(files, ref, findings):
    """bugs_open/175 — an upsert that drops page_type turns a CREATE into a partial update."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        content = file_content(path, ref)
        if not content:
            continue
        for m in PAGE_UPSERT_RE.finditer(strip_comments(content)):
            if not re.search(r"\bpage_type\b", m.group("insert"), re.I):
                continue          # the arm does not state a role; nothing to drop
            if re.search(r"\bpage_type\b", m.group("set"), re.I):
                continue          # deliberately re-types on collision — a different decision
            findings.append((
                "partial-page-upsert", path,
                f"`ON CONFLICT (site_id, name) DO UPDATE` writes the row without "
                f"{BOLD}page_type{RESET}, which the INSERT names",
                "bugs_open/175 — on a name collision this is not a create, it is a PARTIAL "
                "update: your content lands under the existing row's role, nothing errors, "
                "and RETURNING id gives you an id either way. bugs_closed/081 measured that "
                "loop running three months. If the role is a constant of this arm, use "
                "UpsertPageForRole (platform/orchestration/actions/page_role_upsert.go), "
                "which makes the collision an explicit branch. If re-typing on collision is "
                "genuinely right here (the adoption paths argue it is), say so by writing "
                "`page_type = EXCLUDED.page_type` and this stops firing.",
            ))


# ── a hand-rolled "has this page shipped" test ──────────────────────────────
#
# bugs_open/185. `pages.build_status = 'deployed'` reads as "this page is live" and
# is not: a `needs_rebuild` page HAS deployed and is still serving its previous
# artefact (bugs_closed/037), and 28 active pages fleet-wide sit in that state. The
# estate's one definition is datahelpers.NeverDeployedPagePredicateFor / its
# negation PageHasShippedPredicateFor.
#
# WHY A CHECK AND NOT JUST THE HELPER: the council's bug_historian seat put it
# exactly right — a shared predicate stops nobody re-typing the string by hand, and
# hand-re-typing is verbatim how this bug arose. The helper removes the EXCUSE (an
# unaliased constant could not be used by an aliased query); this removes the
# opportunity to do it silently.
#
# PRECISION: `pages` only, and only where the test decides liveness. It does NOT
# fire on `page_components.build_status` (a different table and a different
# question — component deploy state), on writes (`SET build_status = 'deployed'`),
# or on the two checks that legitimately keep the narrow form, which are listed
# with their reasons. MEASURED over the tree on 2026-08-03: 11 raw matches, of which
# 3 are false positives now allow-listed with their reasons (an ORDER BY ranking in
# create_tool_cross_link_items, twice, and queryresolve's correct disjunct — the last
# retired 2026-08-15 when that disjunct became a derivation, see below), leaving
# **8 genuine hits — every one of them in the tranche-2 holdout set named in
# bugs_open/185**, and 0 elsewhere. The first version of this comment claimed "7"
# from counting by eye before running anything; the rule also fired on the 3 false
# positives until that run showed them.
HANDROLLED_SHIPPED_RE = re.compile(
    r"(?<![.\w])(?:[a-z]\.)?build_status\s*=\s*'deployed'")

# Each entry is a DECISION on the record, not an exemption granted by convenience.
SHIPPED_PREDICATE_ALLOWED = {
    "check_page_component_status_drift.go":
        "correct as-is: drift means 'the page claims full deployment but a component "
        "does not'. A needs_rebuild page having non-deployed components is the EXPECTED "
        "state, so the shared predicate would manufacture false positives (measured: 2)",
    "check_unresolved_sections.go":
        "not a liveness test: it flips deployed -> needs_rebuild, and a page already "
        "needs_rebuild is already flagged, so converging adds updated_at churn and no "
        "information",
    "check_news_feed.go":
        "bugs_closed/015's stranded-page set, deliberately the NEGATIVE direction "
        "(build_status <> 'deployed'); bugs_closed/081 records why it is correct",
    "adopt_verbatim.go": "WRITES the value, does not test it",
    "import.go": "WRITES the value, does not test it",
    "fix_component_template_action.go": "page_components write, not a pages liveness test",
    "save_page_sections_action.go": "page_components predicate, different table",
    "maintenance_actions.go":
        "DECIDED 185 tranche 2: findStalePages flags pages for refresh — a needs_rebuild "
        "page is already flagged, so converging double-queues it; findPagesWithNoContent "
        "overlaps check_componentless_pages (PBP-025), which covers the shipped-not-"
        "deployed case with its own deliberate deployed_at predicate",
    "store_generated_component_action.go":
        "DECIDED 185 tranche 2: markPagesForRebuild flips deployed -> needs_rebuild; a "
        "page already needs_rebuild is already flagged, so converging is updated_at churn "
        "(same reason as check_unresolved_sections)",
    "component_library.go":
        "DEAD CODE: GetHeaderNavFromPages / GetFooterNavFromPages — both call sites are "
        "commented out (component_library.go:2064,2114), superseded by nav_tables.go "
        "which already uses the shared predicate. Delete rather than converge when the "
        "nav lane next touches this file",
    "create_tool_cross_link_items.go":
        "ORDER BY (p.build_status = 'deployed') DESC — a ranking PREFERENCE, not a filter. "
        "Nothing is excluded, so no page is missed; converging would only reorder ties",
    # "queryresolve.go" was listed here until 2026-08-15: FetchablePageEligibilitySQL
    # was a hand-written `deployed_at IS NOT NULL OR p.build_status = 'deployed'` (the
    # correct disjunct, not a narrow test). bugs_open/185 fix candidate 2 derived it
    # from datahelpers.PageHasShippedPredicateFor("p"), so the literal is gone and the
    # entry with it — a dead allow-list entry would silence this rule on the one file
    # where a hand-respelled floor is now exactly the drift being guarded against.
}


FLEXLESS_TOGGLE_RE = re.compile(
    r"\.(?:mobile-menu-toggle|hamburger|nav-toggle|menu-toggle)[^{}]*\{[^{}]*\}")


def check_flexless_hamburger(files, ref, findings):
    """bugs_closed/200 — a menu-toggle rule that goes flex without a direction."""
    for path in files:
        if not path.endswith((".css", ".sql", ".go", ".html")):
            continue
        content = file_content(path, ref)
        if not content or ("menu-toggle" not in content and "hamburger" not in content):
            continue
        for m in FLEXLESS_TOGGLE_RE.finditer(content):
            block = m.group(0)
            if re.search(r"display\s*:\s*(?:inline-)?flex\b", block) and "flex-direction" not in block:
                line = content[:m.start()].count("\n") + 1
                findings.append((
                    "flexless-hamburger", f"{path}:{line}",
                    f"a menu-toggle rule sets {BOLD}display:flex{RESET} with no flex-direction",
                    "bugs_closed/200 — the toggle's three <span> bars become ROW flex items and "
                    "fuse into one thin line the width of all three: the mobile menu goes "
                    "visually invisible on every site rendered from the layout. Add "
                    "flex-direction: column (plus centering) in the SAME rule — a base-rule "
                    "declaration far away is what let 18 layouts ship this for four months "
                    "(fixed by seed 314).",
                ))


def check_handrolled_shipped_predicate(files, ref, findings):
    """bugs_open/185 — 'has this page shipped' spelled by hand misses 28 live pages."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        base = os.path.basename(path)
        if base in SHIPPED_PREDICATE_ALLOWED or base == "links.go":
            continue
        body = strip_comments(file_content(path, ref))
        if not body or "build_status" not in body:
            continue
        # Only the lines this commit touched, so an untouched legacy site is not nagged.
        touched = changed_lines(path, ref)
        for i, line in enumerate(body.splitlines(), start=1):
            if i not in touched or "page_components" in line:
                continue
            if HANDROLLED_SHIPPED_RE.search(line):
                findings.append((
                    "handrolled-shipped-predicate", f"{path}:{i}",
                    f"`build_status = 'deployed'` on {BOLD}pages{RESET} — that is not "
                    f"\"is this page live\"",
                    "bugs_open/185 — a needs_rebuild page HAS deployed and is still serving its "
                    "previous artefact (bugs_closed/037: 35 of 46 carry a deployed_at), so this "
                    "test silently skips 28 active pages. Use "
                    "datahelpers.PageHasShippedPredicateFor(\"<alias>\") — it takes the alias, "
                    "which is the reason this kept being re-typed by hand. If the narrow form is "
                    "genuinely right here (drift checks and write paths are), add the file to "
                    "SHIPPED_PREDICATE_ALLOWED with the reason.",
                ))
                break


# ── "unsupported figure" — TRIED and DECLINED, 2026-08-03, on the measurement ──
#
# bugs_open/185 lane. The idea: flag a count or "N of M" claim, newly added to a
# bug/NOTES/PLAN/HANDOFF/SUMMARY file or a Go comment, that carries no date, no
# [MEASURED]-family marker, and no adjacent code block — the "typed from memory"
# shape this session hit twice (a count written from recollection rather than
# from the command sitting next to it).
#
# CALIBRATED, NOT SHIPPED. Two scopes were built and measured against real
# history, and both fired on ordinary, well-evidenced work far past this file's
# own bar ("anything that fired on ordinary work was cut"):
#   - markdown docs (bugs_open/bugs_closed/NOTES/PLAN/HANDOFF/SUMMARY), evidence
#     scoped to the enclosing paragraph: 175 historical commits sampled, 92 fired
#     (53%), 226 total findings.
#   - Go comment blocks fleet-wide, evidence widened to any date OR marker OR a
#     command/query signal in the block: 323 figure-bearing comment blocks found,
#     103 (32%) lacked all three.
#
# READING THE SAMPLES IS WHAT KILLED IT, not the raw rate alone. The false
# positives were not sloppy edge cases — they were this codebase's NORMAL, GOOD
# house style: a section states "**Measured, <date>:**" once, then several
# paragraphs of prose cite counts from it without repeating the marker each
# time. Requiring evidence in every paragraph would make the annotation absurd
# (a tag before every sentence with a number) — and this session's OWN true
# figures were written exactly the same way, so the check could not have told
# my true claims from my false ones. The shape and the defect are lexically
# identical; only the epistemic status differs, and that does not show up as
# text. See TALK_2026-08-03b_the_ratchet_from_judgement_to_mechanical.md.
#
# What DOES stay mechanical from this investigation: the trailer-shape check
# below (a real syntactic property, not a judgement call) and the CLAUDE.md
# doctrine line asking the question a regex cannot: could this measurement have
# come out otherwise? That question is answered by the author, at the moment of
# writing, or not at all.


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


# ── a concept entry that ships without its index row ────────────────────────
# The register is two halves in two files: the ENTRY (`### ABC-001 — name` in a
# category file) and the INDEX ROW (`| ABC-001 | … |` in 000_concept_index.md).
# The index is what a session searches, so an entry with no row reads as **does
# not exist** — and the whole point of the register is to be the thing you
# consult before concluding a capability does not exist.
#
# WHY A GATE AND NOT (ONLY) THE DAILY WATCHER. `concept-register-drift-check`
# (DOC-074) has reported this class every morning since 2026-08-05 and it keeps
# recurring: SCH-024 (08-08), BIZ-031 + WFA-012 (08-10), DIAG-042 (08-10) — four
# concepts in three days, from four different lanes, none of them careless. The
# 08-04 backfill of 34 rows was not a backlog being cleared; it was the first
# reading of a leak. A report can only ask somebody to fix it afterwards, and
# this lane already measured what that is worth: a headline mismatch sat
# uncorrected for three consecutive days *while the watcher named it every
# morning*. This fires in front of the one person who can close it in ten
# seconds, at the moment they are writing the entry — the same argument
# check_register_coverage makes one level up.
#
# IT DRIFTS ONE WAY ONLY, which is why the reverse (a row with no entry) is not
# checked here: adding a concept is two edits in two files and only the first is
# load-bearing for the author, so the row is the half that gets skipped. The
# comm pair in the index header has never once found a row without an entry.
#
# IT READS THE COMMIT, NOT THE WORKTREE. On this tree the house rule is
# `git commit <pathspec>`, which takes the named paths and ignores the index —
# so an author who edited both files but named only the category file in the
# pathspec ships the entry alone. That is the same half-a-move failure the
# LANDMINES file records for `git mv`, and reading the worktree would call it
# clean. `git show :<path>` reads the temporary index git builds for the commit.
REGISTER_ROOT = "docs/agent_docs/docs026_concept_register/register/"
CHECK_PY_REL = "deployments/kustomize/services/concept-register-drift-check/base/check.py"
# The same shape as the watcher's ENTRY_RE, for `git grep`, which cannot be handed
# a compiled Python pattern. Every MATCH is re-parsed with the watcher's own
# ENTRY_RE below, so this string only ever narrows what is read — it can miss a
# heading, it can never define one.
ENTRY_HEADING = "^### [A-Z]{2,4}-[0-9]{3}"


def committed_content(path, ref=None):
    """The file as this COMMIT will contain it — never the worktree. See above."""
    if ref:
        return sh("git", "show", f"{ref[1]}:{path}")
    r = subprocess.run(["git", "show", f":{path}"], capture_output=True, text=True)
    return r.stdout if r.returncode == 0 else file_content(path)


def check_register_entry_without_row(files, ref, findings):
    """A commit adding `### ABC-001` to a register file without its index row.

    IMPORTS THE WATCHER'S OWN PARSER rather than re-deriving the two regexes.
    Two hand-maintained copies of one matching rule is the drift class this
    platform keeps filing bugs about (idx_swi_dedup / workItemTerminalStatuses),
    and it would bite here in a specific way: if this gate's idea of an entry
    heading ever diverged from the watcher's, a commit could pass the gate and
    then be reported by the CronJob for ever, with nothing to tell the author
    which of the two was wrong.

    MEASURED BEFORE INCLUSION, per this file's bar — 390 register-touching
    commits over the 14 days to 2026-08-10. See the fire rate in
    docs026_concept_register/RUNNING_NOTES_concept_register.md (2026-08-10).
    """
    touched = [f for f in files
               if f.startswith(REGISTER_ROOT)
               and f.endswith(".md")
               and "/" not in f[len(REGISTER_ROOT):]]      # direct children only
    if not touched:
        return

    import importlib.util
    repo = sh("git", "rev-parse", "--show-toplevel").strip()
    check_path = os.path.join(repo, CHECK_PY_REL)
    if not os.path.exists(check_path):
        return                      # watcher moved or removed — stay silent, never break a commit
    spec = importlib.util.spec_from_file_location("register_drift_check", check_path)
    watcher = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(watcher)                       # import-time side effects: none

    index_rel = REGISTER_ROOT + watcher.INDEX_NAME

    # Concept ids this commit ADDS as entry headings. A `-` line is ignored on
    # purpose: an entry MOVED between category files shows as one add and one
    # delete, and its row already exists, so it must not fire.
    added = []
    for path in touched:
        if path == index_rel:
            continue
        for line in raw_diff(path, ref).splitlines():
            if not line.startswith("+") or line.startswith("+++"):
                continue
            m = watcher.ENTRY_RE.match(line[1:])
            if m:
                added.append((m.group(1), path, m.group(2).strip(" —-")))
    if not added:
        return

    # ── arm 2: the id is already taken ──────────────────────────────────────
    # This is arm 1's OWN BLIND SPOT, not a separate nice-to-have. Two lanes
    # claimed `LNK-031` three hours apart on 2026-08-08; the second added no row,
    # and arm 1 stays silent on it — because a row for that id already exists,
    # written by the FIRST claimant. So the one shape arm 1 cannot see is the one
    # where the entry is not merely invisible but wrong, and the daily watcher
    # only caught it after both had landed. Renumbering afterwards is permanent
    # damage of a small kind: the originating commit and bugs_open/228 still say
    # LNK-031, so the register carries a note explaining the discrepancy for ever.
    # Ten seconds before the commit, it is a different three digits.
    #
    # Counted over the COMMITTED corpus, not the diff, so an entry MOVED between
    # category files (one add, one delete, still exactly one occurrence) does not
    # fire — the case that would have made this noisy.
    #
    # ⚠ THE ARGUMENT ORDER IS LOAD-BEARING AND FAILS SILENTLY IF YOU GET IT
    # WRONG. `git grep <pattern> --cached` is not a synonym for
    # `git grep --cached <pattern>`: git reads the trailing word as a REVISION
    # and dies with "unable to resolve revision: --cached". sh() captures stdout
    # only, so that fatal arrives here as an empty string — an empty corpus, no
    # collisions, no findings, no error. The first version of this arm was
    # written that way: it passed every --commit audit (where the sha legally
    # follows the pattern) and was INERT in the staged mode the hook actually
    # runs. What caught it was a positive control that staged a known-duplicate
    # id and required a finding; nothing else could have.
    if ref:
        heads = sh("git", "grep", "-E", ENTRY_HEADING, ref[1], "--", REGISTER_ROOT)
    else:
        heads = sh("git", "grep", "--cached", "-E", ENTRY_HEADING, "--", REGISTER_ROOT)
    where = {}
    for line in heads.splitlines():
        rest = line.split(":", 1)[1] if ref else line        # audit output is sha:path:text
        path_hit, _, text = rest.partition(":")
        if os.path.basename(path_hit) == watcher.INDEX_NAME:
            continue
        m = watcher.ENTRY_RE.match(text)
        if m:
            where.setdefault(m.group(1), []).append(os.path.basename(path_hit))

    # The corpus must contain the very headings this commit adds — they are in it
    # by construction. If it does not, the read above is broken rather than the
    # register empty, and arm 2's silence would be a lie rather than a verdict.
    # Say so and drop the arm; never let a broken read pass as "no collisions".
    if any(cid not in where for cid, _, _ in added):
        print(f"{DIM}pattern-check: register id-collision arm skipped — the corpus read "
              f"did not contain headings this commit adds, so it cannot be trusted{RESET}",
              file=sys.stderr)
        where = None

    rows = watcher.parse_rows(committed_content(index_rel, ref))
    for cid, path, name in added:
        holders = where.get(cid, []) if where is not None else []
        if len(holders) > 1:
            elsewhere = sorted(set(holders) - {os.path.basename(path)})
            others = ("also in " + ", ".join(elsewhere)) if elsewhere else "twice in this same file"
            findings.append((
                "register-id-collision", path,
                f"{BOLD}{cid}{RESET} is already used — it appears {len(holders)} times, {others}",
                "Two lanes claimed LNK-031 three hours apart on 2026-08-08 and the collision was "
                "only found the next morning, by which time renumbering left a permanent note in "
                "the register explaining why the originating commit names the wrong id. Pick the "
                "next free number in that prefix now: "
                f"`grep -hoE '^### {cid.split('-')[0]}-[0-9]{{3}}' {REGISTER_ROOT}*.md | sort -u | tail -1`. "
                "Advisory.",
            ))
            continue        # one finding per id: the row it needs depends on which id it ends up with
        if cid in rows:
            continue                # row already present, whenever it landed — not a defect
        findings.append((
            "register-entry-without-row", path,
            f"{BOLD}{cid}{RESET} gets an entry but no row in 000_concept_index.md — {name[:70]}",
            "The index table is what a session searches, so an entry with no row reads as "
            "DOES NOT EXIST and the next lane builds a second one (bugs_open/106). Add "
            f"`| {cid} | name | status | one-line summary | {os.path.basename(path)} |` to the "
            "table in the SAME commit — the register's one rule that ever mattered. If you "
            "commit by pathspec, name the index file too. Advisory.",
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

def check_silent_reply_drop(files, ref, findings):
    """016b §9 / bugs_closed/062 + bugs_closed/133 — a reply that cannot be
    delivered must become a deliverable error, never silence.

    The caller is listening on the reply topic, not reading this pod's logs. A
    producer error on a REPLY path that is only logged leaves the caller to time
    out, and the timeout names no cause. `platform/kafka.DeliverReply` (ADP-017)
    is the built, tested, live mechanism: it answers `FailedUndeliverable` so the
    site can send its own error response instead.

    Measured 2026-08-03 (bugs_open/158): the rule held at **2 of 9** sites — the
    two webscrape paths that adopted it — while reasoning, contentcreator (x2),
    websearch (x2) and thunder (x2) still log-and-return. Widening adoption is an
    RFC, because it changes four services' caller-observable failure behaviour
    (council architecture seat, 7478233b). **This check deliberately does NOT
    change any behaviour** — it makes the remaining sites visible and stops a
    TENTH being written while that RFC is pending.

    Scoped to REPLY produces on purpose. Producing to a request/work topic and
    logging the error is a different decision and not this rule; the reply
    context is what makes silence a broken promise to a waiting caller.

    ⚠ Note for whoever adopts DeliverReply: `FailedUndeliverable` is a RETURN
    VALUE A CALLER CAN IGNORE. A site that calls it and drops the outcome
    compiles, passes every platform/kafka test, and reintroduces the silent
    starve unchanged — so this check also fires when DeliverReply's outcome is
    assigned to `_`.
    """
    go = [f for f in files if f.endswith(".go") and not f.endswith("_test.go")]
    if not go:
        return

    produce_re = re.compile(r"\.Produce(WithValidation)?\s*\(")
    # Reply context. Widened after a positive control caught the first version
    # missing 3 of the 4 known sites: websearch names it `responseTopic`
    # (SINGULAR), so a pattern keyed on `responsesTopic` saw nothing.
    reply_re = re.compile(r"repl(y|ies)|response[s]?Topic|reply_to", re.I)
    logcall_re = re.compile(r"\.(Error|Warn|Errorf|Warnf)\s*\(")
    # Something that actually answers the caller, rather than swallowing.
    answers_re = re.compile(r"DeliverReply|return\s+\S|\.Produce")

    for path in go:
        try:
            with open(path, encoding="utf-8", errors="replace") as fh:
                lines = fh.read().splitlines()
        except OSError:
            continue

        adopted = any("DeliverReply" in ln for ln in lines)

        # A site that adopted DeliverReply but threw the outcome away is back to
        # square one — the landmine bugs_open/158 names explicitly.
        if adopted:
            for i, ln in enumerate(lines):
                if re.search(r"_\s*,\s*\w+\s*:?=\s*\w*\.?DeliverReply\s*\(", ln) or \
                   re.search(r"^\s*_\s*=\s*\w*\.?DeliverReply\s*\(", ln):
                    findings.append((
                        "silent-reply-drop", f"{path}:{i+1}",
                        f"{BOLD}DeliverReply{RESET}'s outcome is discarded",
                        "FailedUndeliverable is a return value a caller can ignore — dropping it "
                        "compiles, passes every platform/kafka test, and reintroduces the silent "
                        "starve unchanged. Answer it with this service's own error response, and "
                        "assert that in an ADAPTER-package test (bugs_closed/133's shape), not by "
                        "trusting platform/kafka's own tests.",
                    ))
                    break
            continue

        for i, ln in enumerate(lines):
            if not produce_re.search(ln):
                continue
            # `return producer.Produce(...)` PROPAGATES the error to the caller.
            # That is the opposite of swallowing it, and a first version of this
            # check reported it — caught by the fleet-wide sweep, not by the
            # positive control, which only ever looks at known-bad files.
            if ln.lstrip().startswith("return "):
                continue
            # Gate on the CALL'S OWN ARGUMENTS, not on a window of surrounding
            # lines. The window version flagged two produces to REQUEST topics
            # (DispatchRequestsTopic, childRequestsTopic) purely because the word
            # "response" appeared a few lines away. Accumulate the call text until
            # its parens balance.
            call, depth = [], 0
            for t in lines[i: min(len(lines), i + 12)]:
                call.append(t)
                depth += t.count("(") - t.count(")")
                if depth <= 0 and len(call) > 0:
                    break
            call_text = "\n".join(call)
            if not reply_re.search(call_text):
                continue                        # not a reply path — different decision
            # How is the produce error handled? Three swallowing shapes, all found
            # in the live tree and all missed by a first version that only looked
            # for "log then bare return":
            #   (a) `if err := ...Produce(...); err != nil { log }`  — falls through
            #   (b) the same with a bare `return`                     — returns nothing
            #   (c) `a.producer.Produce(...)` with the error not captured at all
            tail = lines[i: min(len(lines), i + 16)]
            err_at = None
            for j, t in enumerate(tail):
                if "err != nil" in t and t.rstrip().endswith("{"):
                    err_at = j
                    break
            if err_at is None:
                # (c) — no error check anywhere in this statement. Only report when
                # the statement plainly does not capture err, so a multi-line call
                # whose check sits further out is not misread.
                stmt = " ".join(tail[:6])
                if not re.search(r"\b\w*[eE]rr\w*\s*:?=", stmt):
                    findings.append((
                        "silent-reply-drop", f"{path}:{i+1}",
                        f"a REPLY produce's error is {BOLD}not checked at all{RESET}",
                        "016b §9 / bugs_closed/062: the caller is listening on the reply topic, so "
                        "an undelivered response is a silent timeout with no cause named. This site "
                        "does not even log it. platform/kafka.DeliverReply (ADP-017) is built, "
                        "tested and live — answer FailedUndeliverable with this service's own error "
                        "response. Adoption beyond webscrape is an RFC (bugs_open/158 item 1).",
                    ))
                    break
                continue
            # Body of the error branch, to the closing brace at the `if`'s indent.
            indent = len(tail[err_at]) - len(tail[err_at].lstrip())
            body = []
            for t in tail[err_at + 1:]:
                if t.strip() == "}" and (len(t) - len(t.lstrip())) <= indent:
                    break
                body.append(t)
            joined = "\n".join(body)
            logged = any(logcall_re.search(t) for t in body)
            answers = answers_re.search(joined) is not None

            # An error RETURNED after the block is propagated, not swallowed —
            # the caller still learns the reply failed. Missed by the first
            # version, which only read the block body: processor.go's
            # `if err != nil { log } else { log }` is followed two lines later by
            # `return err`, and it was reported as a silent drop. Scan from the
            # end of the block to the end of the enclosing function.
            err_var = "err"
            m = re.search(r"(\w+)\s*!=\s*nil", tail[err_at])
            if m:
                err_var = m.group(1)
            abs_end = i + err_at + 1 + len(body)
            for t in lines[abs_end: min(len(lines), abs_end + 60)]:
                if t.startswith("}"):                 # end of the enclosing func
                    break
                if re.match(r"\s*return\b", t) and re.search(rf"\b{re.escape(err_var)}\b", t):
                    answers = True
                    break

            if logged and not answers:
                findings.append((
                    "silent-reply-drop", f"{path}:{i+1}",
                    f"a REPLY produce error is {BOLD}logged and swallowed{RESET}",
                    "016b §9 / bugs_closed/062: a response that cannot be delivered must become a "
                    "deliverable error, never silence — the caller is listening on the reply topic, "
                    "not reading this pod's logs, so it times out with no cause named. "
                    "platform/kafka.DeliverReply (ADP-017) is built, tested and live; answer "
                    "FailedUndeliverable with this service's own error response. Adoption beyond "
                    "webscrape is an RFC (bugs_open/158 item 1) — if this is one of the seven "
                    "known sites, do not fix it casually; if it is a NEW site, it should not be "
                    "written this way at all.",
                ))
                break                           # one finding per file is enough to act on


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
                  check_new_capability_surface, check_register_coverage,
                  check_register_entry_without_row,
                  check_runtime_fill_marker, check_unrepaired_component_write,
                  check_unscoped_component_render,
                  check_dynamic_item_type,
                  check_partial_page_upsert, check_silent_reply_drop,
                  check_handrolled_shipped_predicate, check_flexless_hamburger):
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
