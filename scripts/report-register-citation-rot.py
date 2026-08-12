#!/usr/bin/env python3
"""report-register-citation-rot.py — which concept-register citations no longer resolve,
and WHERE the file went.

Signal 2 and 3 of the staleness survey (docs026_concept_register/
FINDINGS_2026-08-10_staleness_survey.md): "`sources:` paths that no longer resolve" and
"bug references that have moved". Like DOC-077 this is a REPORT, not a checker. It never
says an entry is wrong. It says **this citation does not resolve as written**, and where
git says the file is now.

THE QUESTION THIS LANE CARRIES INTO EVERY STALENESS SIGNAL: is there a key that does not
require reading prose? For version lag the answer was the register's own FIELD vocabulary.
Here the field key does NOT separate stale from current — unresolved rates sit at 10-37%
across every field with no clean break. What works instead is a different structural key:
**what git can say about the cited target.** At HEAD / moved-but-present / deleted /
never-existed is a total, mechanical classification, and the middle two name their own
repair.

The field key still earns its place, but as SEVERITY rather than as a filter. A path that
does not resolve means something different depending on the field it sits in:
    sources:       → a grounding claim whose evidence cannot be opened. The sharp case.
    status-evidence: → the proof of a current-state claim cannot be opened.
    verify-later:  → a thing to check, named wrongly or not yet built. Mild by design.

WHAT IT REFUSES TO JUDGE, and why each refusal is load-bearing:

  * TOKENS WITH NO EXTENSION ("build/operate", "internal/adapters/vmhost"). A slash used
    as English conjunction and a directory citation are the same shape. 137 of them here,
    and separating them needs the sentence — the one thing this check will not read.
  * UNROOTED SHORTHAND ("running_notes_15.md", "NOTES.md", "docs006/007"). Extraction-era
    citations into a documentation tree that predates this repo's layout, plus runtime
    site artefacts ("assets/js/snippets.js", "robots.txt") that are object-store files and
    were never in git. Unjudgeable from this repo, permanently. Counted, never listed as
    defects.
  * BRACE NOTATION AND ELLIPSIS ("{PLAN,NOTES}_x.md", "docs021.../025_x.md"). The 08-10
    survey's own misstep: a naive regex made 187 "broken" citations of which 92 were
    artefacts of the extraction, not defects in the register.

THREE BUGS IN THIS SCRIPT'S OWN ANCESTOR, each of which produced confident wrong output.
They are why --self-test exists:

  1. `git rev-list --objects --all` DEDUPS BY OBJECT. Content-identical files share one
     blob and only ONE of their paths is ever printed: 791 of 9,301 HEAD paths were absent
     from its output, 791/791 content-identical duplicates. A path-existence check built on
     it reports live files as never having existed. Use `git log --name-only`.
  2. THE `(N)` SUFFIX IS AMBIGUOUS IN THIS ESTATE. It is an extraction-unit id in some
     citations ("PLAN_tool_widget_clobber(9).md") and genuinely part of the filename in
     others ("002e_concept_spark(6).md" is what the file is called). Stripping it
     unconditionally MANUFACTURED 27 of 34 "never existed" findings — including both paths
     the draft report led with, one of them 15 entries citing the debugging guide. Resolve
     what was written before resolving a guess about what was meant.
  3. TAKING THE LAST VARIANT'S VERDICT INSTEAD OF THE BEST. A citation that resolved
     exactly as written was then overwritten by the stripped form's failure, and 15 entries
     read as "never existed" while the file sat in git under precisely the cited path.

USAGE:
    scripts/report-register-citation-rot.py              # census + the dead citations
    scripts/report-register-citation-rot.py --worklist   # + every citation with a named repair
    scripts/report-register-citation-rot.py --list DELETED
    scripts/report-register-citation-rot.py --self-test  # the resolver against known cases
"""
import collections
import glob
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
REG = os.path.join(ROOT, "docs/agent_docs/docs026_concept_register/register")

ENTRY = re.compile(r"^### ([A-Z]{2,4}-\d{3})\b")
# A field is a bullet whose bold label ENDS with a colon INSIDE the asterisks. Without
# that anchor, bolded prose mid-entry ("**landmine, found 2026-07-29 by the owner**")
# parses as a field name and the field column fills with sentences.
FIELD = re.compile(r"^\s*-\s+\*\*([a-z][^:*]{1,38}):\*\*", re.I)

CODE_EXT = {".go", ".py", ".sh", ".sql", ".js", ".ts", ".tsx", ".yaml", ".yml",
            ".json", ".tf", ".mod", ".css", ".html", ".tmpl"}
DOC_EXT = {".md", ".txt", ".csv"}
REPO_ROOTS = ("platform/", "internal/", "cmd/", "pkg/", "scripts/", "deployments/",
              "build/", "migrations/", "frontend/", "web/", "test/", "tests/",
              "bugs_open/", "bugs_closed/", "docs/")
BUGREF = re.compile(r"^bugs_(open|closed)/(\d{3})$")
GO_SYMBOL = re.compile(r"/\w+\.[A-Z]\w+$")          # platform/kafka.MockProducer
UNJUDGEABLE = re.compile(r"[{}]|\.\.\.|…|\*")       # brace notation, ellipsis, glob
UNIT_SUFFIX = re.compile(r"\((\d+)\)(?=\.[A-Za-z0-9]+$|$)")

# Best-first. Every variant of a token is resolved and the BEST verdict wins (bug 3 above).
RANK = ["AT-HEAD", "AT-HEAD-DIR", "BUG-MOVED", "DELETED", "DELETED-DIR",
        "MOVED-AT-HEAD", "MOVED-AMBIGUOUS", "MOVED-THEN-DELETED", "BUG-MISSING",
        "NEVER-REPO-PATH", "UNJUDGED-DIRSHAPE", "NEVER-UNROOTED"]
RESOLVED = ("AT-HEAD", "AT-HEAD-DIR")
REPAIRABLE = ("MOVED-AT-HEAD", "BUG-MOVED", "DELETED", "DELETED-DIR", "MOVED-THEN-DELETED")
# Severity by the field the citation sits in — the field key, used as an ordering rather
# than as a filter. `sources:` is grounding; `verify-later:` is a to-do list.
SEVERITY = {"sources": 0, "status-evidence": 1, "status": 2, "relations": 3}


def git_paths(args):
    p = subprocess.run(args, cwd=ROOT, capture_output=True, text=True)
    return {l for l in p.stdout.splitlines() if l}


def basename_index(paths):
    d = collections.defaultdict(list)
    for p in paths:
        d[p.rsplit("/", 1)[-1]].append(p)
    return d


def clean(tok):
    """Normalise one whitespace-delimited token into a candidate path, or None."""
    if tok.startswith("`"):
        tok = tok[1:]
    if "`" in tok:                       # a backtick CLOSES the citation; the rest is prose
        tok = tok[:tok.index("`")]       # "adapter.go`'s", "position.go:67`/`defend.go"
    prev = None
    while tok != prev:                   # wrappers nest: (`x.go`), "`y.md`,"
        prev = tok
        tok = tok.strip().strip("`'\"*()[]<>,;:+%").strip()
    tok = re.sub(r"['’]s$", "", tok)                     # possessive: bugs_open/197's
    tok = re.split(r"[#§]", tok)[0]                      # anchors
    tok = re.sub(r":L?\d+([-,]L?\d+)*$", "", tok)        # line refs, incl. ":151,227"
    tok = re.sub(r"(\.[a-z]{2,4}):\w+$", r"\1", tok)     # file.go:SymbolName
    tok = tok.rstrip(".,;:)]}>'\"`")
    return tok.strip("/") or None


def variants(tok):
    """As-cited FIRST, then with the (N) suffix stripped. See bug 2 in the header."""
    yield tok
    stripped = UNIT_SUFFIX.sub("", tok)
    if stripped != tok:
        yield stripped


def is_path(tok):
    if GO_SYMBOL.search(tok):
        return False
    ext = os.path.splitext(tok)[1].lower()
    if ext in CODE_EXT or ext in DOC_EXT or BUGREF.match(tok):
        return True
    return "/" in tok and tok.startswith(REPO_ROOTS)


def klass(tok):
    ext = os.path.splitext(tok)[1].lower()
    if tok.startswith(("bugs_open/", "bugs_closed/")):
        return "bug-ref"
    if ext in CODE_EXT or tok.startswith(REPO_ROOTS[:-1]):
        return "code/config"
    return "doc"


class Resolver:
    def __init__(self, head, ever):
        self.head_idx, self.ever_idx = basename_index(head), basename_index(ever)
        self.head_dirs, self.ever_dirs = self._dirs(head), self._dirs(ever)
        self.bugs = collections.defaultdict(list)
        for p in head:
            m = re.match(r"^bugs_(?:open|closed)/(\d{3})_", p)
            if m:
                self.bugs[m.group(1)].append(p)

    @staticmethod
    def _dirs(paths):
        """EVERY ancestor, not each file's immediate parent. Without this,
        'deployments/kustomize/services/analyser-adapter' reads as never-existed while
        four of its descendants sit at HEAD."""
        out = set()
        for p in paths:
            parts = p.split("/")[:-1]
            for i in range(len(parts)):
                out.add("/".join(parts[:i + 1]))
        return out

    def _suffix(self, tok, idx):
        """Citations are routinely abbreviated to a SUFFIX of the real path
        ('fixloop_eg_dartsonline/0NN_x.sql'), so equality would report the whole
        register as broken.

        But a token with NO directory component carries no evidence about WHICH file it
        means, so a suffix match on a bare basename is not a resolution — it is the first
        candidate in an arbitrary order. 'NOTES.md' matched a/b/NOTES.md here and read as
        cleanly resolved. Bare basenames must be unique to count."""
        cands = idx.get(tok.rsplit("/", 1)[-1], ())
        if "/" not in tok:
            return cands[0] if len(cands) == 1 else None
        for cand in cands:
            if cand == tok or cand.endswith("/" + tok):
                return cand
        return None

    def resolve(self, tok):
        """(verdict, where). Always about the CITATION, never about the entry."""
        m = BUGREF.match(tok)
        if m:
            cited, num = m.group(1), m.group(2)
            hits = self.bugs.get(num, [])
            if not hits:
                return "BUG-MISSING", None
            if any(h.startswith(f"bugs_{cited}/") for h in hits):
                return "AT-HEAD", hits[0]
            return "BUG-MOVED", hits[0]

        hit = self._suffix(tok, self.head_idx)
        if hit:
            return "AT-HEAD", hit
        if tok in self.head_dirs or any(d.endswith("/" + tok) for d in self.head_dirs):
            return "AT-HEAD-DIR", tok
        hit = self._suffix(tok, self.ever_idx)
        if hit:
            return "DELETED", hit
        if tok in self.ever_dirs or any(d.endswith("/" + tok) for d in self.ever_dirs):
            return "DELETED-DIR", tok
        base = tok.rsplit("/", 1)[-1]
        if self.head_idx.get(base):
            # Only a UNIQUE basename match names a repair. "NOTES.md" matches dozens of
            # files and printing the first would be a confident wrong path.
            if len(self.head_idx[base]) == 1:
                return "MOVED-AT-HEAD", self.head_idx[base][0]
            return "MOVED-AMBIGUOUS", f"{len(self.head_idx[base])} candidates at HEAD"
        if self.ever_idx.get(base):
            return "MOVED-THEN-DELETED", self.ever_idx[base][0]
        if not os.path.splitext(tok)[1]:
            return "UNJUDGED-DIRSHAPE", None
        if tok.startswith(REPO_ROOTS):
            return "NEVER-REPO-PATH", None
        return "NEVER-UNROOTED", None


def scan(R):
    rows, skipped = [], 0
    for fn in sorted(glob.glob(os.path.join(REG, "*.md"))):
        if os.path.basename(fn).startswith("000_"):
            continue
        entry, field = None, "(prose)"
        for line in open(fn, errors="replace"):
            m = ENTRY.match(line)
            if m:
                entry, field = m.group(1), "(prose)"
            f = FIELD.match(line)
            if f:
                k = f.group(1).strip().lower()
                field = ("status-evidence" if k.startswith("status-evidence")
                         else "status" if k.startswith("status") else k)
            for raw in line.split():
                if UNJUDGEABLE.search(raw):
                    t = clean(re.sub(r"[{}*]|\.\.\.|…", "", raw) or "")
                    skipped += bool(t and is_path(t))
                    continue
                tok = clean(raw)
                if not tok or not is_path(tok):
                    continue
                verdict, where = min((R.resolve(v) for v in variants(tok)),
                                     key=lambda r: RANK.index(r[0]))
                rows.append((entry, field, klass(tok), tok, verdict, where,
                             os.path.basename(fn)))
    return rows, skipped


def self_test():
    """Synthetic cases for the three bugs that produced confident wrong output, plus the
    citation forms. A resolver that cannot fail is not evidence, so each case names the
    wrong answer it is guarding against."""
    head = {"docs/social001/002e_concept_spark(6).md", "platform/orchestration/x.go",
            "bugs_closed/158_HANDOFF_2026-07-30_eight_reply_sites.md",
            "bugs_open/193_HANDOFF_x.md", "a/b/NOTES.md", "c/d/NOTES.md",
            "deployments/kustomize/services/analyser-adapter/overlays/production/k.yaml"}
    ever = head | {"docs/016b_debugging_guide_merged(3).md"}
    R = Resolver(head, ever)
    best = lambda t: min((R.resolve(v) for v in variants(clean(t))),
                         key=lambda r: RANK.index(r[0]))[0]
    cases = [
        ("docs/social001/002e_concept_spark(6).md", "AT-HEAD",
         "(N) stripped unconditionally → NEVER; it is part of the filename here"),
        ("`docs/016b_debugging_guide_merged(3).md#orientation`", "DELETED",
         "last-variant-wins → NEVER, while git holds the exact cited path"),
        ("platform/orchestration/x.go:151,227", "AT-HEAD", "line refs not stripped → NEVER"),
        ("`bugs_open/193`'s", "AT-HEAD", "backtick/possessive glue → NEVER"),
        ("bugs_open/158", "BUG-MOVED", "bug refs are by NUMBER, not filename"),
        ("deployments/kustomize/services/analyser-adapter", "AT-HEAD-DIR",
         "only leaf parents in the dir set → NEVER for a live directory"),
        ("NOTES.md", "MOVED-AMBIGUOUS", "picking the first of many → a confident wrong path"),
        ("build/operate", "UNJUDGED-DIRSHAPE", "an English slash judged as a path"),
        ("platform/orchestration/sweeper.go", "NEVER-REPO-PATH", "a real dead citation"),
        ("running_notes_15.md", "NEVER-UNROOTED", "extraction-era shorthand judged as rot"),
    ]
    bad = 0
    for tok, want, guard in cases:
        got = best(tok)
        ok = got == want
        bad += not ok
        print(f"   {'ok  ' if ok else 'FAIL'} {tok:58s} {got:18s} guards: {guard}")
    print(f"\n   {len(cases) - bad}/{len(cases)} pass")
    return 1 if bad else 0


def main():
    if "--self-test" in sys.argv:
        return self_test()

    head = git_paths(["git", "ls-tree", "-r", "--name-only", "HEAD"])
    # NOT `git rev-list --objects --all` — see bug 1 in the header.
    ever = git_paths(["git", "log", "--all", "--no-renames", "--pretty=format:", "--name-only"])
    if not head <= ever:
        print(f"CONTROL FAILED: {len(head - ever)} HEAD paths absent from the ever-set.")
        print("Every 'never existed' verdict below would be unfalsifiable. Nothing reported.")
        return 1
    R = Resolver(head, ever)
    rows, skipped = scan(R)
    if not rows:
        print("no path citations found — check REG path")
        return 1

    by = collections.Counter(r[4] for r in rows)
    resolved = sum(by[v] for v in RESOLVED)
    print(f"{len(rows)} path citations across {len({r[0] for r in rows if r[0]})} entries")
    print(f"{resolved} resolve as written ({100*resolved//len(rows)}%)\n")
    for v in RANK:
        if by[v]:
            note = ("  ← resolves" if v in RESOLVED else
                    "  ← git names the repair" if v in REPAIRABLE else
                    "  ← not judged" if v.startswith("UNJUDGED") or v == "NEVER-UNROOTED" else
                    "  ← the file exists; the citation does not say which one"
                    if v == "MOVED-AMBIGUOUS" else
                    "  ← no file, ever, under that name")
            print(f"   {by[v]:6d}  {v:20s}{note}")
    print(f"   {skipped:6d}  (excluded: brace notation / ellipsis / glob)")

    print("\n⚠ BUG REFERENCES ARE ONE-DIRECTIONAL. The owner ruled 2026-08-06 that a fixed")
    print("  bug STAYS in bugs_open/, so a bug that has NOT moved proves nothing. BUG-MOVED")
    print("  is a subset of the drift by construction and can never be a clean bill of health.")

    dead = sorted([r for r in rows if r[4] in ("NEVER-REPO-PATH", "BUG-MISSING")],
                  key=lambda r: SEVERITY.get(r[1], 9))
    print(f"\nNO FILE, EVER, UNDER THAT NAME ({len(dead)}) — ordered by the field it sits in,")
    print("because a dead path in `sources:` is a grounding claim nobody can open, while one")
    print("in `verify-later:` is a to-do named wrongly:")
    for r in dead:
        print(f"   {r[0] or '?':10s} {r[1]:16s} {r[3]}")
        print(f"              in {r[6]}")

    if "--worklist" in sys.argv:
        fix = sorted([r for r in rows if r[4] in REPAIRABLE and r[5]],
                     key=lambda r: (SEVERITY.get(r[1], 9), r[0] or ""))
        print(f"\nWORKLIST — citations git can still locate ({len(fix)}). The entry is not")
        print("wrong; the path is. Each line names its own repair:")
        for r in fix:
            print(f"   {r[0] or '?':10s} {r[1]:16s} {r[4]:14s} {r[3]}")
            print(f"              → {r[5]}")

    if "--list" in sys.argv:
        want = sys.argv[sys.argv.index("--list") + 1]
        sel = [r for r in rows if r[4] == want]
        print(f"\n--- {want} ({len(sel)}) ---")
        for r in sel:
            print(f"   {r[0] or '?':10s} {r[1]:16s} {r[3]}" + (f"   → {r[5]}" if r[5] else ""))

    print("\nCONTROLS")
    print(f"   {len(head)} paths at HEAD, {len(ever)} ever, HEAD ⊆ ever: yes")
    print(f"   resolution rate {100*resolved//len(rows)}% — near 100% means the resolver is")
    print("     too generous to find anything; near 0% means it does not speak the register's")
    print("     citation forms. Both failure modes have happened while building this.")
    print(f"   {by['UNJUDGED-DIRSHAPE'] + by['NEVER-UNROOTED']} citations declared unjudgeable "
          f"and NOT counted as defects")
    print("   the register reads a WORKING TREE, not a ref — an uncommitted entry is included")
    print("     here and invisible to the drift harness, which reads a ref")
    return 0


if __name__ == "__main__":
    sys.exit(main())
