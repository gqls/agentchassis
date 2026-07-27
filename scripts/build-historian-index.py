#!/usr/bin/env python3
"""build-historian-index.py — generate the council historians' case index.

WHY THIS EXISTS. The `bug_historian` seat's "documented history" was seven
narrative items hand-typed into its prompt, frozen at whenever someone wrote
them. Meanwhile the real corpus — the debugging guides, the bug files, and
WRONG_CALLS.md — is ~3.3 MB across 124 files and, being markdown, is invisible
to every seat (`code_symbols` is Go-only; measured 2026-07-27: 4,535 symbols,
0 markdown). So the seats reason about history from a stale constant.

The corpus cannot be inlined. But its HEADINGS can: every entry in 016b §9,
016, and WRONG_CALLS is already a one-line dated assertion — "A mistyped
routing key produces silence in every gate at once, not one loud failure
(2026-07-18)". The headings alone are ~27 KB, which fits a prompt.

So the seats get an INDEX, not the corpus: enough to recognise a shape and name
where it is written up, with the body left for a human to open. The seat is told
this explicitly, because an index that pretends to be the evidence is worse than
no index.

Run with --emit to print the two index blocks; --patch writes them into
fix-proposer's two historian prompts (then mirror with 099). Re-run it whenever
the corpus grows — that is the point: the list is generated, never typed.
"""
import argparse
import json
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
D024 = ROOT / "docs/agent_docs/docs024_key_docs_latest"

G016B = D024 / "016b_debugging_guide_8_consolidated.md"
G016 = D024 / "016_debugging_guide_v2_58_consolidated.md"
WRONG = D024 / "WRONG_CALLS.md"

MARK_START = "<<<HISTORIAN_INDEX_START>>>"
MARK_END = "<<<HISTORIAN_INDEX_END>>>"


def headings(path: pathlib.Path, level: str, after: str | None = None,
             keep: re.Pattern | None = None) -> list[str]:
    """Headings at `level`. `after` starts collecting only past that heading —
    016b's ### are §9 patterns but the file has other sections, and scoping by
    line number would rot the moment someone edits above. `keep` filters to real
    entries: WRONG_CALLS.md opens with structural sections ("Row shape",
    "Entries") that are not wrong calls and must not be indexed as cases."""
    if not path.exists():
        print(f"  WARN missing {path}", file=sys.stderr)
        return []
    out, armed = [], after is None
    for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        if after and not armed:
            if line.startswith("#") and after in line:
                armed = True
            continue
        if line.startswith(level + " "):
            h = re.sub(r"\s+", " ", line[len(level) + 1 :].strip())
            if h and (keep is None or keep.search(h)):
                out.append(h)
    return out


# A wrong-call entry is dated; the file's own scaffolding sections are not.
DATED = re.compile(r"\d{4}-\d{2}-\d{2}")


def bug_slugs() -> list[str]:
    rows = []
    for d in ("bugs_open", "bugs_closed"):
        for f in sorted((ROOT / d).glob("*.md")):
            if f.name.upper().startswith("README"):
                continue
            rows.append(f"{'OPEN ' if d == 'bugs_open' else 'CLOSED'} {f.stem}")
    return rows


PREAMBLE = """
{title}

This index is GENERATED from the platform's written record (scripts/build-historian-index.py).
It is a list of case TITLES, not the cases. You cannot read the bodies — they are markdown and
markdown is not in any corpus you can query. So use it for exactly one thing: recognising that a
shape has been seen before, and naming WHERE it is written up so a human can open it.

Rules, because an index that pretends to be evidence is worse than no index:
- If a title matches the plan's failure shape, cite it BY TITLE and say which file it lives in.
  Do not paraphrase what you imagine the case said — you have not read it.
- Absence from this list is NOT evidence the shape is novel. The list is titles only, the corpus
  starts mid-2026, and plenty of real history was never written up at all.
- Prefer a cited near-match over an uncited confident claim. "This resembles <title>, in 016b §9 —
  worth checking before merging" is a useful objection. "This is unprecedented" almost never is.
{sources}
{body}
"""


def block(title: str, sources: str, sections: list[tuple[str, list[str]]]) -> str:
    body = ""
    for name, items in sections:
        if not items:
            continue
        body += f"\n### {name} ({len(items)})\n"
        body += "\n".join(f"- {h}" for h in items) + "\n"
    return PREAMBLE.format(title=title, sources=sources, body=body).strip()


def build_blocks() -> tuple[str, str]:
    bug = block(
        "PLATFORM CASE INDEX — bug files and specific failure patterns",
        "Sources: /bugs_open/, /bugs_closed/ (case files, numbered; note 016 and 017 are each used "
        "by TWO different cases — resolve by slug, never by bare number), and "
        "016b_debugging_guide_8_consolidated.md §9 (transferable patterns).",
        [
            ("Specific failure patterns — 016b §9",
             headings(G016B, "###", after="Specific Failure Patterns")),
            ("Bug case files — /bugs_open/ and /bugs_closed/", bug_slugs()),
        ],
    )
    dbg = block(
        "DEBUGGING LORE INDEX — incident back-catalogue and our own wrong calls",
        "Sources: 016_debugging_guide_v2_58_consolidated.md (back-catalogue) and WRONG_CALLS.md "
        "(fleet-wide, append-only: claims WE wrote down that turned out false, what caught each "
        "one, and the cheap check that would have). The second is about how WE fail, not how the "
        "system does — read a match there as a warning about the PLAN's reasoning, not its code.",
        [
            ("Back-catalogue patterns — 016", headings(G016, "###")),
            ("Wrong calls — WRONG_CALLS.md", headings(WRONG, "##", keep=DATED)),
        ],
    )
    return bug, dbg


def psql(sql: str, capture=True) -> str:
    cmd = [
        "kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-A", "-t", "-c", sql,
    ]
    r = subprocess.run(cmd, capture_output=capture, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr}")
    return r.stdout


def splice(prompt: str, index_block: str) -> str:
    """Insert (or replace) the index just before the Schema section."""
    payload = f"{MARK_START}\n{index_block}\n{MARK_END}"
    if MARK_START in prompt:
        return re.sub(
            re.escape(MARK_START) + r".*?" + re.escape(MARK_END),
            lambda _: payload, prompt, flags=re.S,
        )
    mark = "## Schema (the ONLY tables available to checks)"
    if mark in prompt:
        return prompt.replace(mark, payload + "\n\n" + mark, 1)
    return prompt.rstrip() + "\n\n" + payload + "\n"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--emit", action="store_true", help="print the blocks and exit")
    ap.add_argument("--patch", action="store_true", help="write into fix-proposer prompts")
    ap.add_argument("--out", default=None, help="write patched config JSON here instead of stdout")
    a = ap.parse_args()

    bug, dbg = build_blocks()
    print(f"bug_historian index:   {len(bug):,} chars", file=sys.stderr)
    print(f"debug_historian index: {len(dbg):,} chars", file=sys.stderr)

    if a.emit:
        print(bug + "\n\n" + "=" * 70 + "\n\n" + dbg)
        return

    if not a.patch:
        ap.error("give --emit or --patch")

    cfg = json.loads(psql(
        "SELECT default_config::text FROM agent_definitions WHERE type='fix-proposer' "
        "AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
    ).strip())
    steps = cfg["workflow"]["steps"]
    before = {k: len(v["config"].get("prompt_template", "")) for k, v in steps.items()
              if k.startswith("review_")}

    for seat, blk in (("review_bug_historian", bug), ("review_debug_historian", dbg)):
        if seat not in steps:
            sys.exit(f"missing seat {seat}")
        steps[seat]["config"]["prompt_template"] = splice(
            steps[seat]["config"]["prompt_template"], blk)

    after = {k: len(v["config"].get("prompt_template", "")) for k, v in steps.items()
             if k.startswith("review_")}
    changed = [k for k in before if before[k] != after[k]]
    print(f"seats changed: {changed}", file=sys.stderr)
    for k in changed:
        print(f"  {k}: {before[k]:,} -> {after[k]:,} chars", file=sys.stderr)

    out = a.out or "/tmp/acm/fix-proposer-INDEXED.json"
    pathlib.Path(out).parent.mkdir(parents=True, exist_ok=True)
    pathlib.Path(out).write_text(json.dumps(cfg, ensure_ascii=False, separators=(",", ":")))
    print(f"\nwrote {out} ({pathlib.Path(out).stat().st_size:,} bytes)", file=sys.stderr)
    print("apply it, then mirror:", file=sys.stderr)
    print("  SRC=%s /tmp/acm/APPLY_council_memory.sh   # (script honours SRC)" % out, file=sys.stderr)


if __name__ == "__main__":
    main()
