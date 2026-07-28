#!/usr/bin/env python3
"""memory-index.py — budget checker and reconciler for the auto-memory index.

WHY THIS EXISTS
---------------
`MEMORY.md` is loaded into EVERY session automatically. That is its whole value —
the lessons most worth heeding are the ones nobody thinks to look up — and also its
whole problem: past a hard read limit it stops loading at all and every session
loses its map. It reached 24.0KB against a 24.4KB limit on 2026-07-28, roughly
400 bytes from that failure, after ~2.4KB of growth in one evening.

Six hand-compactions in a fortnight is not a maintenance history, it is a design
signal: thirty-odd concurrent sessions edit one document with no schema, no budget,
and no way to say no at the moment a line is added. This tool is the "say no".

RECONCILE, NEVER REGENERATE
---------------------------
The memory instructions tell every session to hand-add a pointer line to
`MEMORY.md`. A tool that regenerated the file from topic-file frontmatter would
silently destroy those edits — turning a helpful build step into a data-loss
machine, which is the exact failure class the memory git hook was written for.

So: this tool NEVER deletes a line it does not understand. `--sync` copies index
lines DOWN into topic-file frontmatter (making the topic file the source of truth
going forward). `--rebuild` copies them back UP, but only for entries that already
carry frontmatter, and leaves every hand-added line untouched and in place.

USAGE
  memory-index.py                 # check: sizes, budgets, over-length lines  (advisory)
  memory-index.py --strict        # same, but exit 1 on a breach (for a hook)
  memory-index.py --sync          # back-fill index_line frontmatter into topic files
  memory-index.py --rebuild       # rewrite index lines from topic-file frontmatter
"""
import argparse
import os
import re
import sys

MEM_DIR = os.path.expanduser(
    "~/.claude/projects/-home-ant-projects-agentchassis/memory")
INDEX = "MEMORY.md"

# ── budgets (C) ────────────────────────────────────────────────────────────────
# HARD is the point past which the index does not load and every session loses its
# map. SOFT is the tooling's nag. The section budgets exist so that the answer to
# "what do we cut" is decided when a line is ADDED, not in a panic at the ceiling.
HARD_KB, SOFT_KB = 24.4, 17.1
SECTION_BUDGET_KB = {
    "practice entries": 8.0,   # protected: these must stay auto-loaded
    "bug entries": 5.0,
    "status banners": 1.5,     # volatile, rewritten constantly
    "header/ruling": 2.5,
}
# The markdown link prefix is a fixed cost (median 80ch, max 107ch on this file),
# so a flat LINE cap is unfair to long slugs and unachievable without gutting them.
# Cap the CONTENT after the em-dash instead.
CONTENT_CAP = 90

FRONTMATTER_KEY = "index_line"


def classify(line):
    if line.startswith(">") or line.startswith("# "):
        return "header/ruling"
    if not line.startswith("- "):
        return None
    if re.match(r"- \[?\*{0,2}[Bb]ug", line):
        return "bug entries"
    if line.startswith("- **"):
        return "status banners"
    return "practice entries"


def split_entry(line):
    """-> (head, content, topic_file) or None. head includes the ' — ' separator."""
    m = re.match(r"(- \[[^\]]*\]\(([^)]+\.md)\)\s*[—-]\s*)(.*)", line)
    return (m.group(1), m.group(3), m.group(2)) if m else None


def read_frontmatter(path):
    try:
        text = open(path, encoding="utf-8").read()
    except OSError:
        return None, None
    m = re.match(r"^---\n(.*?)\n---\n", text, re.S)
    return (m.group(1), text) if m else (None, text)


def check(strict=False):
    lines = open(INDEX, encoding="utf-8").read().split("\n")
    total = sum(len(l.encode()) + 1 for l in lines) / 1024
    sizes, breaches = {}, []

    for l in lines:
        k = classify(l)
        if k:
            sizes[k] = sizes.get(k, 0) + (len(l.encode()) + 1) / 1024

    print(f"{INDEX}: {total:.1f}KB   (soft {SOFT_KB}KB · HARD {HARD_KB}KB)")
    if total > HARD_KB:
        breaches.append(f"TOTAL {total:.1f}KB EXCEEDS THE HARD LIMIT — the index "
                        f"will not load; sessions are running blind")
    elif total > SOFT_KB:
        headroom = HARD_KB - total
        print(f"  over soft budget, {headroom:.1f}KB of headroom before it stops loading")

    print("\n  section                    size   budget")
    for k, budget in SECTION_BUDGET_KB.items():
        got = sizes.get(k, 0.0)
        flag = "  OVER" if got > budget else ""
        print(f"  {k:24s} {got:5.1f}KB  {budget:5.1f}KB{flag}")
        if got > budget:
            breaches.append(f"{k} is {got:.1f}KB against a {budget:.1f}KB budget")

    # over-length content
    longs, missing = [], []
    for l in lines:
        e = split_entry(l)
        if not e:
            continue
        head, content, topic = e
        if len(content) > CONTENT_CAP:
            longs.append((len(content), topic, content[:60]))
        if not os.path.exists(topic):
            missing.append(topic)

    if longs:
        print(f"\n  {len(longs)} entr{'y' if len(longs)==1 else 'ies'} over the "
              f"{CONTENT_CAP}-char content cap:")
        for n, topic, snippet in sorted(longs, reverse=True)[:12]:
            print(f"    {n:4d}ch  {topic}\n            {snippet}…")
        breaches.append(f"{len(longs)} entries over the {CONTENT_CAP}-char content cap")

    if missing:
        print(f"\n  {len(missing)} entr{'y' if len(missing)==1 else 'ies'} point at a "
              f"topic file that does not exist:")
        for t in missing:
            print(f"    {t}")
        breaches.append(f"{len(missing)} dangling topic-file links")

    # orphans: topic files nothing points at (informational only — recall still finds them)
    linked = {split_entry(l)[2] for l in lines if split_entry(l)}
    on_disk = {f for f in os.listdir(".")
               if f.endswith(".md") and f not in
               {INDEX, "MEMORY_closed.md", "MEMORY_workstreams.md"}}
    orphans = on_disk - linked
    if orphans:
        print(f"\n  {len(orphans)} topic files are not linked from any index "
              f"(reachable by recall, invisible to a cold start)")

    print()
    if breaches:
        print("BREACHES:")
        for b in breaches:
            print(f"  · {b}")
        if strict:
            return 1
        print("\n(advisory — nothing was blocked)")
    else:
        print("within budget on every axis.")
    return 0


def sync():
    """Copy index lines DOWN into topic-file frontmatter. Never edits MEMORY.md."""
    lines = open(INDEX, encoding="utf-8").read().split("\n")
    added = skipped = 0
    for l in lines:
        e = split_entry(l)
        if not e:
            continue
        _, content, topic = e
        fm, text = read_frontmatter(topic)
        if fm is None:
            skipped += 1
            continue
        if re.search(rf"^{FRONTMATTER_KEY}:", fm, re.M):
            skipped += 1
            continue
        safe = content.replace('"', '\\"')
        new_fm = fm + f'\n{FRONTMATTER_KEY}: "{safe}"'
        open(topic, "w", encoding="utf-8").write(
            text.replace(f"---\n{fm}\n---\n", f"---\n{new_fm}\n---\n", 1))
        added += 1
    print(f"index_line frontmatter: {added} added, {skipped} already present or no frontmatter")
    return 0


def rebuild():
    """Copy frontmatter index lines back UP. Hand-added lines are left alone."""
    lines = open(INDEX, encoding="utf-8").read().split("\n")
    out, changed, untouched = [], 0, 0
    for l in lines:
        e = split_entry(l)
        if not e:
            out.append(l)
            continue
        head, content, topic = e
        fm, _ = read_frontmatter(topic)
        m = re.search(rf'^{FRONTMATTER_KEY}:\s*"(.*)"\s*$', fm or "", re.M)
        if not m:
            out.append(l)          # no frontmatter — this is someone's hand-added line
            untouched += 1
            continue
        want = m.group(1).replace('\\"', '"')
        if want != content:
            out.append(head + want)
            changed += 1
        else:
            out.append(l)
    open(INDEX, "w", encoding="utf-8").write("\n".join(out))
    print(f"rebuilt {changed} line(s) from topic frontmatter; "
          f"{untouched} hand-added line(s) left untouched")
    return 0


if __name__ == "__main__":
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--sync", action="store_true")
    p.add_argument("--rebuild", action="store_true")
    p.add_argument("--strict", action="store_true")
    p.add_argument("--hook", action="store_true",
                   help="silent unless the index is over a budget; for a PostToolUse hook")
    a = p.parse_args()
    os.chdir(MEM_DIR)
    if a.hook:
        # Runs after every Write/Edit, so it must cost nothing and say NOTHING
        # unless it matters. A hook that chatters gets ignored, and an ignored
        # hook is worse than none because it looks like coverage. Two filters:
        #
        #  1. Only speak if the edit actually touched the memory directory.
        #  2. Only speak for SIZE breaches — the failure that stops the index
        #     loading. Over-length lines are guidance for the full report, not
        #     grounds to interrupt someone editing an unrelated file.
        try:
            import json
            payload = json.load(sys.stdin)
            path = str(payload.get("tool_input", {}).get("file_path", ""))
        except Exception:
            path = ""
        if "/memory/" not in path.replace("\\", "/"):
            sys.exit(0)
        total = os.path.getsize(INDEX) / 1024
        if total <= SOFT_KB:
            sys.exit(0)
        head = ("WILL NOT LOAD — sessions are running blind"
                if total > HARD_KB else
                f"over soft budget; {HARD_KB - total:.1f}KB before it stops loading")
        sys.stderr.write(
            f"\n── memory index {total:.1f}KB: {head} ──\n"
            "  python3 scripts/memory-index.py   for the section breakdown\n")
        sys.exit(0)          # advisory: never block a write
    sys.exit(sync() if a.sync else rebuild() if a.rebuild else check(a.strict))
