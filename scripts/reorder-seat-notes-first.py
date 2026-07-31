#!/usr/bin/env python3
"""reorder-seat-notes-first.py — move `notes` to the head of a seat's output schema,
for the seats whose REMIT makes notes load-bearing. bugs_open/138 fix candidate 4.

WHY THIS IS PER-SEAT AND NOT A RULE. Candidate 4 asked for the load-bearing field
first in every seat's schema, because truncation eats the tail. Measured across all 51
live templates, that is already satisfied and going further would HURT:

  * `reviewer` and `verdict` are first in 51 of 51 — which is why salvage works.
  * `severity` is last inside each objection, which looks like the same bug one level
    down. 0 of 2,713 stored objections lack a severity. REFUTED.
  * `notes` is what truncation actually destroys (present in 2 of 30 degraded reviews
    vs 3,067 of 3,076 complete ones) — and for almost every seat it is the RIGHT
    thing to lose, because moving it forward pushes `objections` into the tail, and
    objections carry both the severities the gate reads and the content the proposer
    revises against.

So this applies ONLY where a seat's own prompt makes `notes` carry something the round
cannot do without. Today that is exactly one remit: `review_architecture`, whose
mandated ARCHITECTURE_SIGNAL lives in `notes`. When that seat truncated, the field
that made it MEASURABLE was the field destroyed — and the result was indistinguishable
from "this seat is noise, retire it". fix-proposer and council-gate were fixed by hand
on 2026-07-29; feature-designer was reached by nothing, because 099 mirrors only those
two councils.

WHY JSON PARSING AND NOT STRING SURGERY. The output line is syntactically valid JSON
even though it is a template (`"problem": "..."`, `"edit": 1`). So the reorder is
parse -> rebuild with the key order I want -> dump, which cannot mangle an escape,
a `|` inside a value, or a nested object. Cutting `, "notes": "…"` out with a regex
can, and the notes value here contains `\\n\\n`, `|` and angle brackets.

IDEMPOTENT: a seat whose `notes` is already in position 3 is reported and skipped.

Usage:
    ./scripts/reorder-seat-notes-first.py                 # dry run: show before/after
    ./scripts/reorder-seat-notes-first.py --apply         # snapshot, then write
"""

import argparse
import json
import subprocess
import sys

NS = "ai-persona-system"
POD = "postgres-clients-0"
ANCHOR = "## Output"

# (council, seat) -> why this seat's remit makes `notes` load-bearing.
TARGETS = [
    ("feature-designer", "review_architecture",
     "its mandated ARCHITECTURE_SIGNAL lives in `notes`; fix-proposer and council-gate "
     "got this on 2026-07-29 and 099 mirrors only those two, so this council was "
     "reached by nothing. Owner approved propagating it 2026-07-31"),
]


def psql(sql, stdin=False):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--",
           "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1", "-t", "-A"]
    r = subprocess.run(cmd if stdin else cmd + ["-c", sql],
                       input=sql if stdin else None, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}")
    return r.stdout


def live_prompt(council, seat):
    # prompt_template is a SIBLING of ai_service under config; max_tokens is INSIDE
    # ai_service. Both wrong depths return NULL for every row rather than erroring.
    out = psql(
        "SELECT s.value->'config'->>'prompt_template' "
        "FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s "
        f"WHERE a.type='{council}' AND a.is_active AND COALESCE(a.is_snapshot,false)=false "
        f"AND a.deleted_at IS NULL AND s.key='{seat}';")
    return out if out.strip() else None


def find_schema_line(prompt):
    """The output schema is the first line after the ## Output heading that parses as
    a JSON object. Returning the line ITSELF (not an index into a regex match) keeps
    the replacement exact."""
    if ANCHOR not in prompt:
        return None, None
    tail = prompt[prompt.index(ANCHOR):]
    for line in tail.splitlines():
        s = line.strip()
        if s.startswith("{") and s.endswith("}"):
            try:
                return line, json.loads(s)
            except json.JSONDecodeError:
                continue
    return None, None


def reordered(obj):
    """reviewer, verdict, notes, then everything else in its original order."""
    head = ["reviewer", "verdict", "notes"]
    out = {k: obj[k] for k in head if k in obj}
    for k, v in obj.items():
        if k not in out:
            out[k] = v
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true")
    args = ap.parse_args()

    todo = []
    for council, seat, why in TARGETS:
        prompt = live_prompt(council, seat)
        if prompt is None:
            print(f"  [MISSING  ] {council}/{seat}")
            continue
        line, obj = find_schema_line(prompt)
        if obj is None:
            print(f"  [NO SCHEMA] {council}/{seat} — no JSON object line after '{ANCHOR}'")
            continue
        if "notes" not in obj:
            print(f"  [NO NOTES ] {council}/{seat} — schema has no `notes` key; nothing to move")
            continue
        keys = list(obj)
        if keys[:3] == ["reviewer", "verdict", "notes"]:
            print(f"  [ALREADY  ] {council}/{seat} — notes already third")
            continue
        new_line = json.dumps(reordered(obj), separators=(", ", ": "), ensure_ascii=False)
        print(f"  [REORDER  ] {council}/{seat}")
        print(f"              why: {why}")
        print(f"              before: {' -> '.join(keys)}")
        print(f"              after : {' -> '.join(reordered(obj))}")
        todo.append((council, seat, prompt, line, new_line))

    if not todo:
        print("\nNothing to do.")
        return
    if not args.apply:
        print("\nDRY RUN. The new schema line, in full:\n")
        for _, _, _, _, nl in todo:
            print("   " + nl)
        print("\nRe-run with --apply to snapshot and write.")
        return

    for council in sorted({c for c, _, _, _, _ in todo}):
        psql(f"SELECT snapshot_agent('{council}', "
             f"'pre-update: notes-first output order, bugs_open/138 candidate 4');")
        print(f"  snapshot taken: {council}")

    for council, seat, prompt, line, new_line in todo:
        # Replace the schema line only. count=1 and an equality assert, because a
        # prompt that contained the line twice would otherwise be silently mangled.
        assert prompt.count(line) == 1, f"{council}/{seat}: schema line appears {prompt.count(line)} times"
        new_prompt = prompt.replace(line, new_line, 1)
        if "$PT$" in new_prompt:
            sys.exit(f"refusing: {council}/{seat} contains the dollar-quote tag")
        out = psql(
            "UPDATE agent_definitions SET default_config = jsonb_set(default_config, "
            f"'{{workflow,steps,{seat},config,prompt_template}}', to_jsonb($PT${new_prompt}$PT$::text), false), "
            "updated_at = now() "
            f"WHERE type='{council}' AND is_active AND COALESCE(is_snapshot,false)=false "
            "AND deleted_at IS NULL "
            f"AND default_config #>> '{{workflow,steps,{seat},config,prompt_template}}' IS NOT NULL "
            "RETURNING type;", stdin=True)
        rows = [l for l in out.splitlines() if l.strip() == council]
        print(f"  applied: {council}/{seat} (rows updated: {len(rows)})")
        if len(rows) != 1:
            sys.exit(f"expected exactly 1 row, got {len(rows)} — stopping")

    print("\nVerifying against the live rows:")
    for council, seat, _, _, _ in todo:
        _, obj = find_schema_line(live_prompt(council, seat) or "")
        print(f"  {council}/{seat}: {' -> '.join(list(obj)[:4])} …" if obj else f"  {council}/{seat}: UNREADABLE")


if __name__ == "__main__":
    main()
