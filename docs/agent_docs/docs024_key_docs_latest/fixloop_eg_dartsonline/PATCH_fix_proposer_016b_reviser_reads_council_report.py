#!/usr/bin/env python3
"""PATCH_fix_proposer_016b_reviser_reads_council_report.py

bugs_open/016, SECOND finding: the reviser sees 6 of 13 seats.

THE CHOICE (owner's call, delegated to this thread by the reasoning-dataset
audit, 2026-07-18): list all thirteen seats in the prompt, or have the reviser
read the council_report artifact ONCE. This patch takes the second, because the
first recurs on seat 14 — and the gap arrived by seat growth, not by a code
defect, so a fix that scales with the roster is the only one that closes it.

WHAT IT DOES
  1. Inserts `load_council_reviews` (query_database, no LLM) that reads the
     newest council_report body for this fix correlation — every seat that
     voted, verbatim, whatever the roster is that day.
  2. Routes it between `council_decide` and `check_approved`, so BOTH the
     revise path (-> repropose) and the veto path (-> reframe) have it.
  3. Rewrites `repropose` and `reframe` prompts: the per-seat sections become
     one "## The council's reviews" section fed by the artifact.
  4. Drops the per-seat `review_*` entries from both steps' `input_fields`
     (they are what had to be edited on every seat add) and adds
     `council_reviews`.

Net effect: seat 14 reaches the reviser with no prompt edit at all. reframe
gains eleven seats it never saw (it referenced only edit-quality + guardian).

IDEMPOTENT: if `load_council_reviews` already exists, it reports and exits
without writing — so two threads applying it concurrently cannot duplicate
anything (the reason the "list all thirteen" option was rejected).

KNOWN CAVEAT, stated not hidden: `query_database` resolves params only from
collected_data, which does not carry the orchestration id, so the query keys on
correlation + newest row. Within a run that IS this round's report (council_decide
wrote it seconds earlier). Two proposer runs racing the same correlation could in
principle cross wires — narrow, and the 090 coverage check already discourages
concurrent work on one target. The clean end-state is a small Go change:
`diagnose_council_decide` already holds the parsed reviews in memory and could
return them, removing both this query and the caveat. That needs an image; this
does not.

USAGE:  python3 PATCH_...py           # dry run, prints the new prompts
        python3 PATCH_...py --apply   # snapshot, then write
"""
import json
import re
import subprocess
import sys

NS, POD = "ai-persona-system", "postgres-clients-0"
DB = ["-U", "clients_user", "-d", "clients_db"]

SECTION = """## The council's reviews — EVERY seat that voted this round, verbatim
{{.council_reviews.body}}

This is the council_report artifact: a JSON object with `decision`, `decided_by`
and `reviews[]`, one entry per seat ({reviewer, verdict, objections[], missing[],
notes}). Address EVERY objection from EVERY seat, not only the ones you find
familiar — the roster grows, and a seat you have not seen before is not
advisory noise. The guardian holds the only hard veto; the rest are advisory,
but an unaddressed advisory objection is what sends this round back again."""


def psql(sql=None, stdin_payload=None):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--", "psql", *DB, "-tA", "-v", "ON_ERROR_STOP=1"]
    cmd += ["-c", sql] if stdin_payload is None else ["-f", "-"]
    r = subprocess.run(cmd, input=stdin_payload, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}")
    return r.stdout.strip()


def rewrite_prompt(prompt, first_header, end_anchor):
    """Replace the run of per-seat review sections with the single artifact
    section. first_header is the first '## ... reviewer ...' line; end_anchor is
    the header that follows the last one (or None = to end of prompt)."""
    start = prompt.find(first_header)
    if start == -1:
        return prompt, False
    end = prompt.find(end_anchor, start) if end_anchor else len(prompt)
    if end == -1:
        return prompt, False
    return prompt[:start] + SECTION + "\n\n" + prompt[end:], True


def main():
    apply = "--apply" in sys.argv
    wf = json.loads(psql(
        "SELECT default_config->'workflow' FROM agent_definitions WHERE type='fix-proposer' "
        "AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"))
    steps = wf["steps"]

    if "load_council_reviews" in steps:
        print("load_council_reviews already present — already patched. Nothing to do.")
        return

    seats = len(steps["council_decide"]["config"]["review_fields"])
    before_repro = sorted(set(re.findall(r"\{\{\.(review_[a-z_]+)\}\}", steps["repropose"]["config"]["prompt_template"])))
    before_ref = sorted(set(re.findall(r"\{\{\.(review_[a-z_]+)\}\}", steps["reframe"]["config"]["prompt_template"])))
    print(f"seats seeded: {seats} | repropose referenced: {len(before_repro)} | reframe referenced: {len(before_ref)}")

    steps["load_council_reviews"] = {
        "action": "query_database",
        "description": ("Load the newest council_report for this correlation — every seat's verdict, "
                        "verbatim — so the reviser reads the whole council from one artifact instead of "
                        "each seat threaded through the prompt (bugs_open/016 second finding: 7 of 13 "
                        "seats were invisible to the reviser, a gap that arrived by seat growth)."),
        "output_field": "council_reviews",
        "next_step": "check_approved",
        "config": {
            "output_format": "object",
            "error_step": "check_approved",
            "params": ["input_data.fix_correlation_id"],
            "query": ("SELECT body FROM diagnosis_artifacts "
                      "WHERE correlation_id = $1 AND kind = 'council_report' "
                      "ORDER BY created_at DESC LIMIT 1"),
        },
    }
    steps["council_decide"]["next_step"] = "load_council_reviews"

    p, ok = rewrite_prompt(steps["repropose"]["config"]["prompt_template"],
                           "## Edit-quality reviewer said", "## Verification results")
    if not ok:
        sys.exit("REFUSING: could not locate repropose's reviewer sections")
    steps["repropose"]["config"]["prompt_template"] = p

    p, ok = rewrite_prompt(steps["reframe"]["config"]["prompt_template"],
                           "## Edit-quality review", "## Output")
    if not ok:
        sys.exit("REFUSING: could not locate reframe's reviewer sections")
    steps["reframe"]["config"]["prompt_template"] = p

    for name in ("repropose", "reframe"):
        cfg = steps[name]["config"]
        cfg["input_fields"] = [f for f in cfg["input_fields"] if not f.startswith("review_")] + ["council_reviews"]
        left = re.findall(r"\{\{\.(review_[a-z_]+)\}\}", cfg["prompt_template"])
        if left:
            sys.exit(f"REFUSING: {name} still references per-seat fields: {left}")
        print(f"\n--- {name}: input_fields -> {cfg['input_fields']}")
        print(cfg["prompt_template"])

    if not apply:
        print("\nDRY RUN — nothing written. Re-run with --apply.")
        return

    payload = json.dumps({"workflow": wf})
    tag = "$fp016b$"
    if tag in payload:
        sys.exit("REFUSING: payload contains the dollar-quote tag")
    psql(stdin_payload=(
        "BEGIN;\n"
        "SELECT snapshot_agent('fix-proposer', 'pre-update: 016 second finding — reviser reads the "
        "council_report artifact instead of per-seat prompt sections');\n"
        f"UPDATE agent_definitions SET default_config = {tag}{payload}{tag}::jsonb, updated_at = now() "
        "WHERE type='fix-proposer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;\n"
        "COMMIT;\n"))
    print(f"\nAPPLIED. The reviser now reads all {seats} seats from the artifact; seat 14 needs no prompt edit.")


if __name__ == "__main__":
    main()
