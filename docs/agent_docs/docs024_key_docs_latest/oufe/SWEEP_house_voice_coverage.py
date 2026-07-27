#!/usr/bin/env python3
"""
SWEEP_house_voice_coverage.py — apply the house voice to every prose producer.

WHY THIS EXISTS
  Migration 228 put the house voice into `page-content-writer` and
  `content-writer`. 230 added `grounded-explainer`. Then a coverage count:

      with_voice | total
               3 |    26

  Three of twenty-six. The other twenty-three are content-creator-hero,
  -about, -contact, -cta, -features, -testimonials, blog-content-planner,
  simple-content-writer-with-approval and the rest — every one of which writes
  copy a reader sees.

  Each of the three was wired by hand, one at a time, because it was the one in
  front of me. That is the same failure this workstream has now recorded four
  times in a week: **a rule written down changes nothing that already exists.**
  The register froze and nothing swept. A deferral lapsed and nothing swept. Bug
  095 was filed and the prepared SQL still carried the defect. The writer rule
  went in and the existing copy still broke it.

  So this is the sweep, rather than a fourth hand-patch.

WHAT IT DOES
  Finds every active agent that produces reader-facing prose, locates each
  `prompt_template` under a generate-shaped step wherever it lives in the
  workflow tree (the paths genuinely differ: `generate_content`,
  `generate_hero_content`, `generate_about_content`, `generate_draft`, and some
  nested under `process_sections_loop.sub_workflow`), and appends the voice
  block once.

  Idempotent. An agent already carrying the marker is skipped, so re-running is
  safe and the count is the report.

WHAT IT DELIBERATELY DOES NOT TOUCH
  Anything that is not reader-facing prose: classifiers, planners that emit
  JSON structure, reviewers, auditors, extractors. A voice instruction on a
  JSON-emitting prompt is noise at best and corrupts the output at worst. The
  allow-list is by agent type and the step name must look like generation.

USAGE
  ./SWEEP_house_voice_coverage.py            # dry run — prints what it would do
  ./SWEEP_house_voice_coverage.py --apply    # writes
"""

import argparse
import json
import subprocess
import sys

MARKER = "HOUSE VOICE"

VOICE = """

## HOUSE VOICE — follow this unless the site's own voice spec says otherwise

Write the way a knowledgeable person explains something out loud to one other person.

Start with the fact. Never open by saying what something is NOT before saying what it is. Fold any genuine contrast in afterwards, as a trailing clause.

One idea per sentence. Do not use em dashes: use two sentences, or a comma, or brackets.

Contractions in ordinary sentences: it's, isn't, doesn't, don't.

Match the word to the size of the fact, in both directions. Neither "critical" and "transformative" nor "nothing fancy" and "surprisingly simple".

Cut: crucially, genuinely, exactly, "which is the point", "what matters here is", at its core, in essence, seamless, robust, leverage, delve, furthermore, moreover.

Vary paragraph length. Never open consecutive sentences with "It is", "This is" or "There is". No exclamation marks, no hype adjectives.
"""

# Reader-facing prose producers only. Reviewers, classifiers, planners that emit
# JSON, extractors and auditors are excluded on purpose.
TYPE_PATTERNS = ("content-creator", "content-writer", "page-content-writer",
                 "grounded-explainer", "blog-content", "simple-content-writer")
EXCLUDE = ("planner", "classifier", "auditor", "reviewer", "extractor",
           "researcher", "validator")

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db", "-tAc"]


def q(sql: str) -> str:
    return subprocess.run(PSQL + [sql], capture_output=True, text=True, timeout=120).stdout


def find_prompt_paths(node, path=None):
    """Every ('a','b','prompt_template') path under a generate-shaped step."""
    path = path or []
    out = []
    if isinstance(node, dict):
        for k, v in node.items():
            if k == "prompt_template" and isinstance(v, str):
                # only under a step that looks like generation
                if any("generat" in str(p).lower() or "draft" in str(p).lower() for p in path):
                    out.append(path + [k])
            else:
                out.extend(find_prompt_paths(v, path + [k]))
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true")
    args = ap.parse_args()

    like = " OR ".join(f"type ILIKE '%{p}%'" for p in TYPE_PATTERNS)
    notlike = " AND ".join(f"type NOT ILIKE '%{e}%'" for e in EXCLUDE)
    rows = q(f"""SELECT type FROM agent_definitions
                  WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                    AND ({like}) AND ({notlike}) ORDER BY type;""").strip().splitlines()
    types = [r.strip() for r in rows if r.strip()]
    print(f"candidate prose producers: {len(types)}")

    todo, skipped = [], 0
    for t in types:
        cfg = q(f"SELECT default_config FROM agent_definitions WHERE type='{t}' "
                f"AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL LIMIT 1;").strip()
        if not cfg:
            continue
        if MARKER in cfg:
            skipped += 1
            continue
        paths = find_prompt_paths(json.loads(cfg).get("workflow", {}).get("steps", {}))
        if paths:
            todo.append((t, paths))
        else:
            print(f"  no generate-shaped prompt found: {t}")

    print(f"already carry the voice: {skipped}")
    print(f"to update: {len(todo)}")
    for t, paths in todo:
        for p in paths:
            print(f"    {t}  ->  workflow.steps.{'.'.join(p)}")

    if not args.apply:
        print("\nDRY RUN — nothing written. Re-run with --apply.")
        return 0

    done = 0
    for t, paths in todo:
        for p in paths:
            jpath = "{workflow,steps," + ",".join(p) + "}"
            sql = (f"UPDATE agent_definitions SET default_config = jsonb_set(default_config, "
                   f"'{jpath}', to_jsonb((default_config#>>'{jpath}') || $hv${VOICE}$hv$), false), "
                   f"updated_at=NOW() WHERE type='{t}' AND is_active "
                   f"AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL "
                   f"AND (default_config#>>'{jpath}') NOT LIKE '%{MARKER}%';")
            subprocess.run(PSQL + [sql], capture_output=True, text=True, timeout=120)
            done += 1
    print(f"\nupdated {done} prompt(s) across {len(todo)} agent(s).")
    print("Verify on OUTPUT, not on the prompt. The rule being present is not")
    print("evidence the copy followed it — that is exactly how oufe shipped 19")
    print("negative-frame constructions with the rule already live.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
