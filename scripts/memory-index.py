#!/usr/bin/env python3
"""memory-index.py — budget checker and reconciler for the auto-memory index.

WHY THIS EXISTS
---------------
`MEMORY.md` is loaded into EVERY session automatically. That is its whole value —
the lessons most worth heeding are the ones nobody thinks to look up — and also its
whole problem: past a hard cap the harness TRUNCATES it. It reached 24.0KB against
a 24.4KiB limit on 2026-07-28, roughly 400 bytes from that, after ~2.4KB of growth
in one evening.

THE CAP, READ FROM THE CLI BUNDLE — NOT INFERRED
------------------------------------------------
    var DS = "MEMORY.md", cie = 200, Ixe = 25000, ...

  * byte cap  25,000 bytes  (= 24.41 KiB, the "24.4KB" everyone quotes)
  * LINE cap  200 lines     (a second axis, easy to miss)
  * warn at 0.8x cap, "compact to" target 0.7x cap — so 19.5KiB and 17.1KiB were
    never independent judgements, just fractions of one constant.

Not raisable: both are hardcoded constants, no setting and no environment variable
(CLAUDE_COWORK_MEMORY_* covers path/guidelines/injected content, not caps).

*** IT TRUNCATES — IT DOES NOT STOP LOADING. *** The harness's own over-cap message:

    "The write succeeded, but everything past the limit is silently dropped each
     time the index is loaded — entries at the end are already invisible to readers."

This corrects what this file, MEMORY.md and memory-index-how-it-works.md all used to
say. It matters because the failure is SILENT and PARTIAL, and it eats the TAIL —
so the newest entries go first, and ENTRY ORDER IS A SAFETY PROPERTY: whatever must
survive belongs at the top. (Same lesson as bugs_open/138 one layer down: put the
load-bearing field first in any structured output that can be truncated.)

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

NEWEST WRITER WINS (the direction guard, 2026-07-29)
----------------------------------------------------
`--rebuild` alone is still not safe, because "frontmatter is the source of truth"
silently loses whichever edit landed most recently — and sessions are *instructed*
to hand-edit `MEMORY.md`. Observed live on 2026-07-29: a session rewrote the
architecture-review line minutes after a `--sync`, and a blind rebuild would have
reverted it. So `--rebuild` compares mtimes per entry:

    topic file newer  -> someone edited index_line deliberately   -> rebuild it
    MEMORY.md newer   -> someone hand-edited the index            -> KEEP it,
                         and it is listed so `--sync` can push it down

Both directions therefore converge instead of fighting, and neither can clobber a
concurrent session. Two invariants back it: the link SET may never change across a
rebuild (a lost link is a memory no cold start reaches again), and `--sync` matches
`index_line:` at ANY indent — it is conventionally nested under `metadata:`, and an
anchored `^index_line:` created a SECOND key in 8 files before this was fixed.

USAGE
  memory-index.py                 # check: sizes, budgets, over-length lines  (advisory)
  memory-index.py --strict        # same, but exit 1 on a breach (for a hook)
  memory-index.py --hook          # silent unless over budget; for a PostToolUse hook
  memory-index.py --sync          # DRY RUN: index lines -> topic frontmatter
  memory-index.py --sync --apply  # ...write it
  memory-index.py --rebuild       # DRY RUN: topic frontmatter -> index lines
  memory-index.py --rebuild --apply

ROUTINE: `--sync` after you hand-edit the index; `--rebuild` after you edit a topic
file's `index_line`. Running both in either order is safe and idempotent.
"""
import argparse
import os
import re
import sys

MEM_DIR = os.path.expanduser(
    "~/.claude/projects/-home-ant-projects-agentchassis/memory")
INDEX = "MEMORY.md"

# ── budgets ────────────────────────────────────────────────────────────────────
# The two hard caps, taken from the CLI bundle (`cie=200, Ixe=25000`), plus the
# fractions the harness derives its warn/target from. Past the byte cap the file is
# TRUNCATED and the tail is silently dropped — see the header. The section budgets
# exist so that the answer to "what do we cut" is decided when a line is ADDED, not
# in a panic at the ceiling.
BYTE_CAP = 25_000
LINE_CAP = 200
WARN_FRAC, TARGET_FRAC = 0.8, 0.7

HARD_KB = BYTE_CAP / 1024                    # 24.4
SOFT_KB = BYTE_CAP * TARGET_FRAC / 1024      # 17.1
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

# ── per-entry byte budgets ─────────────────────────────────────────────────────
# From MEMORY.md's own preamble (measured 2026-07-31, 9th pass). They are in BYTES
# because the cap is, and because sizing entries with bash `${#var}` — which counts
# CHARACTERS — made a whole round of trims 2–16B short apiece: every `—`, `⚠`, `≡`
# is 1 char and 3 bytes.
#
# A FAMILY line is billed PER MEMBER, not per line. That is the unit sessions
# actually add: the 2026-07-29 MERGE-ON-ADD ruling tells everyone to merge a new
# practice into a family line rather than append a bullet, so growth arrives as a
# new member on an unbounded line. Measured 2026-08-06, three days after the
# delivery fix: 49 surviving entries grew +2,392B while 5 retired and 2 arrived,
# and four family lines held 28% of the whole file. Everyone had complied.
MEMBER_BUDGET = 150          # one member of a family line
ENTRY_BUDGET = 400           # a whole line carrying a single link

# The slot count is sized to the HARNESS NAG LINE (0.8x = 20,000B), not to the cap.
# Sizing it to the cap would declare the file healthy right up to the point the tail
# starts vanishing, which is the failure this whole tool exists to prevent — and it
# is what the per-entry budgets do today: at 115 units they permit 30,394B, or 121%
# of the cap. A budget that permits a file bigger than the file can be is not a
# budget. This one leaves 5,000B of genuine headroom.
SLOT_TARGET_BYTES = int(BYTE_CAP * WARN_FRAC)

FRONTMATTER_KEY = "index_line"

# Sibling indexes. They are NOT topic files, and the links they hold are what
# keeps a topic file off the orphan list.
SIBLINGS = ("MEMORY_closed.md", "MEMORY_workstreams.md")

LINK_RE = re.compile(r"\]\(([^)]+\.md)\)")


def links_in(text):
    """Every topic-file link in a string. A FAMILY line carries several."""
    return LINK_RE.findall(text)


def entry_units(text):
    """{slug: (bytes, budget)} for every BILLABLE unit in an index.

    A single-link line is one unit billed at ENTRY_BUDGET. A FAMILY line is one
    unit PER MEMBER, split on the ` · ` separator and billed at MEMBER_BUDGET,
    because a member is what a session adds and therefore what a session can be
    asked to keep small. Billing the whole family line instead reports a breach
    nobody can act on — the line is long BY DESIGN, it is the compaction working.

    Keyed by slug so two snapshots of the index can be joined to find which units
    a given write actually GREW. Byte lengths, never character lengths.
    """
    units = {}
    for line in text.splitlines():
        if not line.startswith("- "):
            continue
        members = line.split(" · ") if len(links_in(line)) > 1 else [line]
        budget = MEMBER_BUDGET if len(members) > 1 else ENTRY_BUDGET
        for m in members:
            got = links_in(m)
            if got:
                units[got[0]] = (len(m.encode("utf-8", "replace")), budget)
    return units


def classify(line):
    if line.startswith(">") or line.startswith("# "):
        return "header/ruling"
    if not line.startswith("- "):
        return None
    # FAMILY lines (2026-07-29 ruling): one line, several related memories, e.g.
    #   - **A claim about behaviour is NOT the behaviour** — [a](x.md) · [b](y.md)
    # They open with `- **` like a status banner but are the OPPOSITE of volatile —
    # they are the compaction working, and they are long BY DESIGN. Billing them to
    # the 1.5KB status-banner budget reported a 5.5KB breach that was mostly the
    # merge everyone was asked to do. Count them as practice entries, which is what
    # they are.
    if len(links_in(line)) > 1:
        return "practice entries"
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

    nbytes = sum(len(l.encode()) + 1 for l in lines)
    nlines = len(lines)
    print(f"{INDEX}: {nbytes:,}B of {BYTE_CAP:,} ({nbytes*100//BYTE_CAP}%) · "
          f"{nlines} of {LINE_CAP} lines ({nlines*100//LINE_CAP}%)")

    if nbytes > BYTE_CAP:
        breaches.append(
            f"TOTAL {nbytes:,}B EXCEEDS THE {BYTE_CAP:,}B CAP — the tail past the cap "
            f"is SILENTLY DROPPED on every load; the last entries are already invisible")
    elif total > SOFT_KB:
        print(f"  over soft budget, {BYTE_CAP - nbytes:,}B before the tail starts "
              f"being truncated")

    # The SECOND axis. Slack today, but it is the opposite lever to bytes: families
    # trade lines for bytes, and "split the long lines up" trades bytes for lines.
    # Unmonitored, someone fixes one cap by breaching the other.
    if nlines > LINE_CAP:
        breaches.append(f"{nlines} lines EXCEEDS THE {LINE_CAP}-line cap — same "
                        f"silent tail truncation as the byte cap")
    elif nlines > LINE_CAP * WARN_FRAC:
        breaches.append(f"{nlines} lines is over {int(WARN_FRAC*100)}% of the "
                        f"{LINE_CAP}-line cap")

    # Truncation eats the TAIL, so what sits at the bottom is what disappears first.
    # Naming them turns an abstract budget into "these three entries, by name".
    if nbytes > BYTE_CAP * WARN_FRAC:
        tail = [l for l in lines if l.startswith("- ")][-3:]
        if tail:
            print("\n  first to be silently dropped (truncation eats the tail):")
            for l in tail:
                e = split_entry(l)
                print(f"    {e[2] if e else l[:72]}")

    print("\n  section                    size   budget")
    for k, budget in SECTION_BUDGET_KB.items():
        got = sizes.get(k, 0.0)
        flag = "  OVER" if got > budget else ""
        print(f"  {k:24s} {got:5.1f}KB  {budget:5.1f}KB{flag}")
        if got > budget:
            breaches.append(f"{k} is {got:.1f}KB against a {budget:.1f}KB budget")

    # ── the axis that actually binds ───────────────────────────────────────────
    # Per-entry budgets bound the WORST entry, never the total. Measured 08-06:
    # 115 units at FULL budget would be 30,394B = 121% of the cap, so the file fits
    # only because most entries sit well under (members 120B of 150, singles 242B
    # of 400). Enforcing every budget perfectly would reclaim 802B — about three
    # days. COUNT is the constraint, so report it as a slot count with the
    # displacement rule attached, and recompute the sustainable number from the
    # LIVE mean rather than trusting the constant as the corpus changes shape.
    body = "\n".join(lines)
    pre_bytes = 0
    for l in lines:
        if l.startswith("- "):
            break
        pre_bytes += len(l.encode()) + 1
    units = entry_units(body)
    if units:
        ubytes = sum(s for s, _ in units.values())
        mean = ubytes // len(units)
        room = (SLOT_TARGET_BYTES - pre_bytes) // max(mean, 1)
        flag = "  OVER" if len(units) > room else ""
        print(f"\n  entries                  {len(units):5d}  {room:5d} slots{flag}"
              f"   (mean {mean}B; slots = ({SLOT_TARGET_BYTES:,}B target "
              f"- {pre_bytes:,}B preamble) / mean)")
        if len(units) > room:
            print(f"    {len(units) - room} unit(s) over — an ARRIVAL MUST DISPLACE ONE. "
                  f"Closed-and-live bugs belong in MEMORY_closed.md,")
            print( "    which is already the sanctioned destination; retiring one is not "
                   "a loss, it is the rule working.")
            breaches.append(f"{len(units)} entries against {room} sustainable slots")

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

    # orphans: topic files NO index points at (informational — recall still finds
    # them, but a cold start never will).
    #
    # Two bugs lived here until 2026-07-29, and together they overstated this by
    # 30x — 129 reported against 4 real:
    #   1. `split_entry` returns only the FIRST link on a line, so every member of
    #      a FAMILY line after the first counted as unlinked;
    #   2. the sibling indexes were excluded from `on_disk` but never READ, so the
    #      ~89 memories they link looked orphaned.
    # Scan every link in every index instead.
    linked = set(links_in("\n".join(lines)))
    for sib in SIBLINGS:
        try:
            linked.update(links_in(open(sib, encoding="utf-8").read()))
        except OSError:
            print(f"\n  warn: sibling index missing: {sib}")
    on_disk = {f for f in os.listdir(".")
               if f.endswith(".md") and f not in {INDEX, *SIBLINGS}
               and not f.startswith("MEMORY.")}          # MEMORY.backup-*.md
    orphans = sorted(on_disk - linked)
    if orphans:
        print(f"\n  {len(orphans)} topic file(s) are not linked from ANY index "
              f"(reachable by recall, invisible to a cold start):")
        for o in orphans[:15]:
            print(f"    {o}")
        if len(orphans) > 15:
            print(f"    … and {len(orphans)-15} more")

    # a link owned by two entries means two places to update, i.e. drift
    from collections import Counter
    entry_links = [l for ln in lines if ln.startswith("- ") for l in links_in(ln)]
    dupes = {k: v for k, v in Counter(entry_links).items() if v > 1}
    if dupes:
        print(f"\n  {len(dupes)} memory(ies) linked from more than one entry:")
        for k, v in sorted(dupes.items()):
            print(f"    x{v}  {k}")
        breaches.append(f"{len(dupes)} duplicate memory link(s) — one memory, one home")

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


# The key is conventionally NESTED under `metadata:` (two-space indent), but some
# files carry it at column 0. Both are "present"; an anchored `^index_line:` sees
# only the second. That mismatch made --sync append a SECOND key to 8 files on
# 2026-07-29 — two conflicting index lines in one file, which is the exact
# duplicated-state problem this tool exists to remove. Match either form, and
# write back in the form the file already uses.
INDEX_LINE_RE = re.compile(rf"^(\s*){FRONTMATTER_KEY}:\s*(.*)$", re.M)


def sync(apply=False):
    """Copy index lines DOWN into topic-file frontmatter. Never edits MEMORY.md."""
    lines = open(INDEX, encoding="utf-8").read().split("\n")
    added = updated = skipped = 0
    for l in lines:
        e = split_entry(l)
        if not e:
            continue
        _, content, topic = e
        fm, text = read_frontmatter(topic)
        if fm is None:
            skipped += 1
            continue

        safe = content.replace('\\', '\\\\').replace('"', '\\"')
        m = INDEX_LINE_RE.search(fm)
        if m:
            # already present (at whatever indent) — refresh it in place, keeping
            # the file's own indentation so we never create a second key
            new_line = f'{m.group(1)}{FRONTMATTER_KEY}: "{safe}"'
            if m.group(0) == new_line:
                skipped += 1
                continue
            new_fm = fm[:m.start()] + new_line + fm[m.end():]
            updated += 1
        else:
            # absent — nest it under `metadata:` if that block exists, matching the
            # established convention rather than inventing a top-level key
            mm = re.search(r"^metadata:\s*$", fm, re.M)
            if mm:
                insert = f'\n  {FRONTMATTER_KEY}: "{safe}"'
                new_fm = fm[:mm.end()] + insert + fm[mm.end():]
            else:
                new_fm = fm + f'\n{FRONTMATTER_KEY}: "{safe}"'
            added += 1

        if apply:
            open(topic, "w", encoding="utf-8").write(
                text.replace(f"---\n{fm}\n---\n", f"---\n{new_fm}\n---\n", 1))

    verb = "" if apply else "would be "
    print(f"index_line frontmatter: {added} {verb}added, {updated} {verb}refreshed, "
          f"{skipped} already current or no frontmatter")
    if not apply:
        print("DRY RUN — re-run with --apply to write.")
    return 0


def rebuild(apply=False):
    """Copy frontmatter index lines back UP. Hand-added lines are left alone."""
    original = open(INDEX, encoding="utf-8").read()
    lines = original.split("\n")
    out, changed, untouched, stale = [], 0, 0, []
    for l in lines:
        e = split_entry(l)
        if not e:
            out.append(l)
            continue
        head, content, topic = e
        fm, _ = read_frontmatter(topic)
        # same either-indent match as sync(); an anchored pattern here would treat
        # every conventionally-nested key as "no frontmatter" and rebuild nothing
        m = re.search(rf'^\s*{FRONTMATTER_KEY}:\s*"(.*)"\s*$', fm or "", re.M)
        if not m:
            out.append(l)          # no frontmatter — this is someone's hand-added line
            untouched += 1
            continue
        want = m.group(1).replace('\\"', '"').replace('\\\\', '\\')

        # DIRECTION GUARD. Sessions are instructed to hand-edit MEMORY.md, so a
        # blind frontmatter-wins rebuild reverts whichever edit landed most
        # recently — the data-loss machine this file's header warns about, and it
        # was live on 2026-07-29: another session had just rewritten the
        # architecture-review line and the frontmatter still held the older text.
        # Newest writer wins, in both directions:
        #   topic file newer  -> someone edited index_line deliberately  -> rebuild
        #   MEMORY.md newer   -> someone hand-edited the index           -> keep,
        #                        and --sync will push it down
        if want != content:
            try:
                if os.path.getmtime(INDEX) > os.path.getmtime(topic):
                    out.append(l)
                    stale.append(topic)
                    continue
            except OSError:
                pass
        if want != content:
            out.append(head + want)
            changed += 1
        else:
            out.append(l)
    result = "\n".join(out)

    # THE invariant: a rebuild may reword an entry, never lose one. A dropped link
    # is a memory that no cold start will ever reach again — and this file is edited
    # by many concurrent sessions, so the damage would not be noticed for days.
    from collections import Counter
    before, after = Counter(links_in(original)), Counter(links_in(result))
    if before != after:
        lost = before - after
        gained = after - before
        sys.stderr.write(
            "REFUSING to write: the rebuild changed the link set.\n"
            f"  lost:   {dict(lost) or '{}'}\n"
            f"  gained: {dict(gained) or '{}'}\n")
        return 2

    size = len(result.encode()) / 1024
    print(f"rebuilt {changed} line(s) from topic frontmatter; "
          f"{untouched} hand-added line(s) left untouched")
    if stale:
        print(f"  {len(stale)} entr(y|ies) KEPT because MEMORY.md is newer than the "
              f"topic file (a hand-edit wins; run --sync to push it down):")
        for t in stale:
            print(f"    {t}")
    print(f"  result {size:.1f}KB (was {len(original.encode())/1024:.1f}KB), "
          f"link set unchanged ({sum(after.values())} links)")
    if apply:
        open(INDEX, "w", encoding="utf-8").write(result)
        print("  written.")
    else:
        print("  DRY RUN — re-run with --apply to write.")
    return 0


if __name__ == "__main__":
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--sync", action="store_true")
    p.add_argument("--rebuild", action="store_true")
    p.add_argument("--strict", action="store_true")
    p.add_argument("--apply", action="store_true",
                   help="with --sync: actually write (default is a dry run)")
    p.add_argument("--hook", action="store_true",
                   help="silent unless the index is over a budget; for a PostToolUse hook")
    a = p.parse_args()
    os.chdir(MEM_DIR)
    if a.hook:
        # Runs after every Write/Edit, so it must cost nothing and say NOTHING
        # unless it matters. A hook that chatters gets ignored, and an ignored
        # hook is worse than none because it looks like coverage. Three filters:
        #
        #  1. Only speak if the edit touched the INDEX itself (or the index is
        #     already over the hard cap, which is everyone's problem). Writing a
        #     topic file is the behaviour we want — do not tax it.
        #  2. Only speak for SIZE breaches — the failure that stops the index
        #     loading. Over-length lines are guidance for the full report, not
        #     grounds to interrupt someone editing an unrelated file.
        #  3. Carry the DELTA this write caused, so the message is a live gauge
        #     rather than a standing nag. A trim that lands reads as a negative
        #     number, which is the confirmation a compaction pass otherwise has
        #     to hand-roll with `wc -c`.
        #
        # DELIVERY — THE DEFECT THIS BLOCK WAS WRITTEN WITH (fixed 2026-08-03).
        # It used to `sys.stderr.write(...)` then `sys.exit(0)`, reasoning
        # "advisory: never block a write". Both halves of that are wrong for
        # PostToolUse:
        #   * The tool has ALREADY RUN, so nothing this hook does can block the
        #     write. The safety the exit code was buying did not exist.
        #   * On exit 0 the docs route stdout to the debug log and say nothing
        #     about stderr; only exit 2 "shows stderr to Claude". So the warning
        #     reached NOBODY — not the model, not the transcript. It was captured
        #     in the session .jsonl as an `attachment.stderr` record, which looks
        #     exactly like delivery and is not.
        # Measured: 15 hand-compactions of MEMORY.md between this tool being
        # written (2026-07-28) and the defect being found, every one of them
        # under an armed, correct, wired, MUTE detector.
        # The documented path is `hookSpecificOutput.additionalContext`, which is
        # what the two sibling hooks in this repo already use
        # (`landmines-session-start.py`, `.claude/hooks/psql_readonly_gate.py`).
        # Copy them, not this file's history.
        try:
            import json
            payload = json.load(sys.stdin)
            path = str(payload.get("tool_input", {}).get("file_path", ""))
        except Exception:
            path = ""
        path = path.replace("\\", "/")
        if "/memory/" not in path:
            sys.exit(0)
        nbytes = os.path.getsize(INDEX)
        if nbytes <= SOFT_KB * 1024 or (
                not path.endswith("/" + INDEX) and nbytes <= BYTE_CAP):
            sys.exit(0)

        # Delta against the previous snapshot of the index. The sibling hook
        # commits every write, so history is dense; which of the last two commits
        # is "before" depends on hook ordering, so pick by content rather than
        # assuming. Failure here must never cost the warning — hence the bare
        # except and the empty-string default.
        delta, grown = "", []
        try:
            import subprocess
            shas = subprocess.run(
                ["git", "log", "-2", "--format=%H", "--", INDEX],
                cwd=MEM_DIR, capture_output=True, text=True, timeout=5
            ).stdout.split()
            blobs = []
            for s in shas:
                blobs.append(subprocess.run(
                    ["git", "show", f"{s}:{INDEX}"], cwd=MEM_DIR,
                    capture_output=True, timeout=5).stdout)
            before = next((b for b in blobs if len(b) != nbytes), None)
            if before is not None:
                # "since the last snapshot", NOT "your write". ~30 sessions share
                # this file; the sibling hook commits per write, so the window is
                # usually one edit, but a concurrent session's bytes land inside it
                # and get attributed to whoever writes next. Observed within minutes
                # of shipping this: a member another lane had just merged was
                # reported to me as "+153B this write". Naming the wrong author is
                # how a true signal gets distrusted, so claim only what is measured.
                delta = f", {nbytes - len(before):+,}B since the last snapshot"
                # WHICH unit grew, not just by how much. The file-level number
                # names no one and is therefore nobody's job; this names the entry
                # the writer just touched, while they still have the context to
                # spend those bytes better. Only entries THIS write grew, and only
                # those now over budget — a standing list of every over-length
                # entry is the full report's job and would be pure chatter here.
                was = entry_units(before.decode("utf-8", "replace"))
                now = entry_units(open(INDEX, encoding="utf-8").read())
                for slug, (size, budget) in now.items():
                    old = was.get(slug, (0, budget))[0]
                    if size > old and size > budget:
                        grown.append((size, budget, slug, size - old))
                grown.sort(reverse=True)
        except Exception:
            pass

        if nbytes > BYTE_CAP:
            head = (f"OVER CAP by {nbytes - BYTE_CAP:,}B — the tail is being SILENTLY "
                    f"DROPPED on every load, so the newest entries are already "
                    f"invisible to every cold start")
        else:
            head = (f"{BYTE_CAP - nbytes:,}B of headroom before the TAIL starts being "
                    f"truncated (newest entries go first, silently)")
        over = ""
        if grown:
            rows = "\n".join(
                f"  {s:,}B against a {b:,}B budget (+{d:,}B this write)  {slug}"
                for s, b, slug, d in grown[:4])
            over = ("\n%d entr%s grew past budget since the last snapshot (yours, or a "
                    "concurrent session's — the file has ~30 writers):\n%s\n"
                    "A FAMILY member is billed at %dB and a single-link entry at %dB — "
                    "merging into a family line still costs the bytes, it just hides "
                    "them on a line that is long by design.\n"
                    % (len(grown), "y" if len(grown) == 1 else "ies", rows,
                       MEMBER_BUDGET, ENTRY_BUDGET))
        msg = (
            f"memory index: MEMORY.md is {nbytes:,}B of {BYTE_CAP:,} "
            f"({nbytes * 100 // BYTE_CAP}%){delta} — {head}.\n"
            f"{over}"
            "The index carries the TRIGGER only: enough to recognise you are in this "
            "situation. Push the remedy, the numbers and the war story DOWN into the "
            "topic file — confirm they landed there (`grep -F`) before cutting, because "
            "a cap-driven trim targets exactly the lines nobody has re-read lately.\n"
            "`python3 scripts/memory-index.py` for the section breakdown and the list of "
            "entries that get dropped first."
        )
        print(json.dumps({
            "hookSpecificOutput": {
                "hookEventName": "PostToolUse",
                "additionalContext": msg,
            },
            "suppressOutput": True,
        }))
        sys.exit(0)          # advisory: PostToolUse cannot block — the write is done
    sys.exit(sync(a.apply) if a.sync else rebuild(a.apply) if a.rebuild else check(a.strict))
