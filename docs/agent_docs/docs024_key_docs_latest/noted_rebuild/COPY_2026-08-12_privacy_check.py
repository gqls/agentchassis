#!/usr/bin/env python3
"""
Check the draft privacy copy against noted.co.uk's LIVE evidence_base bans.

Three things that make this a check rather than a gesture:

 1. The patterns are read from the DATABASE, not copied into this file. A ban the
    owner adds tomorrow is enforced here tomorrow, with no edit.
 2. It tests ONLY the section after "## THE DRAFT". The commentary above it quotes
    banned phrases on purpose ("we can't see your notes", "your notes are safe"),
    and a checker that read the whole file would fail on the explanation of why
    those phrases are forbidden.
 3. It runs a POSITIVE CONTROL first — the old site's real sentence, which MUST be
    caught. A zero from a checker that has never fired is not evidence.

Usage: python3 COPY_2026-08-12_privacy_check.py
Exit 0 = draft is clean AND the control fired.
"""
import re
import subprocess
import sys
from pathlib import Path

DRAFT = Path(__file__).with_name("COPY_2026-08-12_privacy_DRAFT_for_owner.md")

PSQL = [
    "kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
    "psql", "-U", "clients_user", "-d", "clients_db", "-tA", "-c",
    """SELECT b->>'pattern' FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
       jsonb_array_elements(ss.data->'banned_claims') b
       WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""",
]

# The old site's actual sentence (guides/about.html:34), the one the first ban exists
# for. If this is not caught, the checker is broken and a clean draft means nothing.
CONTROL = "We can't see your notes, read your text, or listen to your recordings."


def live_patterns():
    out = subprocess.run(PSQL, capture_output=True, text=True, timeout=60)
    pats = [p.strip() for p in out.stdout.splitlines() if p.strip()]
    if not pats:
        print("FATAL: read no patterns from the database — refusing to report a pass")
        print(out.stderr[:400])
        sys.exit(2)
    return pats


def draft_text():
    body = DRAFT.read_text(encoding="utf-8")
    if "## THE DRAFT" not in body:
        print("FATAL: no '## THE DRAFT' marker — cannot isolate the copy")
        sys.exit(2)
    section = body.split("## THE DRAFT", 1)[1]
    return section.split("## Verification", 1)[0]


def hits(text, pats):
    found = []
    for p in pats:
        try:
            rx = re.compile(p, re.I)
        except re.error:                      # evidence_base rule: bad regex degrades
            rx = re.compile(re.escape(p), re.I)   # to a literal, never silently drops
        for m in rx.finditer(text):
            found.append((p, m.group(0)))
    return found


def main():
    pats = live_patterns()
    print(f"{len(pats)} live banned_claims patterns\n")

    ctrl = hits(CONTROL, pats)
    print("POSITIVE CONTROL (the old site's real sentence — must be caught)")
    if ctrl:
        print(f"  CAUGHT by: {ctrl[0][0]}\n         on: {ctrl[0][1]!r}\n")
    else:
        print("  NOT CAUGHT — the checker is not working; a clean draft proves nothing")
        return 2

    text = draft_text()
    bad = hits(text, pats)
    print(f"DRAFT ({len(text.split())} words)")
    if bad:
        for p, m in bad:
            print(f"  BLOCKED by {p}\n    matched: {m!r}")
        return 1
    print("  clean against all patterns")

    # Non-blocking style flags from the writer_block.
    style = ["seamless", "effortless", "powerful", "revolutionise", "unlock"]
    found = [w for w in style if re.search(rf"\b{w}", text, re.I)]
    nums = re.findall(r"\b\d[\d,.]*\s*(?:%|users|customers|notes)\b", text, re.I)
    print(f"  writer_block style words: {found or 'none'}")
    print(f"  figures (there are no registered facts, so any number is invented): {nums or 'none'}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
