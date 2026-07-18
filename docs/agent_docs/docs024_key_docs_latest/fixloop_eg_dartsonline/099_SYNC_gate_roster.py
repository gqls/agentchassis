#!/usr/bin/env python3
"""099_SYNC_gate_roster.py — mirror the LIVE fix-proposer council roster onto
the LIVE council-gate, deterministically.

WHY THIS EXISTS. The council-gate runs the same council as the fix-proposer, so
every seat added to one must be added to the other. That was hand-work, and the
roster changed FIVE times in 18 hours (3 -> 5 -> 7 -> 9 -> 13 seats); each
hand-mirror is a chance to drift, and two hand-maintained rosters that must stay
identical is precisely the failure family the reuse-agent and guidelines seats
review for. So the mirror is mechanical now: read the live fix-proposer row,
transform, write the live council-gate row.

WHAT IT MIRRORS: every `review_*` and `gate_*` step, plus select_panel's
`footprints` map, plus council_decide.review_fields and run_checks.check_fields.

THE FOUR TRANSFORMS (a submission is not a diagnosis):
  1. prompt context: '## The diagnosis' + '{{.diagnosis_row.conclusion}}'
     becomes '## The author's stated rationale' + '{{.input_data.rationale}}'.
  2. input_fields: 'diagnosis_row' becomes 'input_data'.
  3. error_step: 'complete_refused' becomes 'complete_invalid' (the gate's
     no-verdict terminal).
  4. chain endpoints: review_editquality -> the FIRST gate; the last gate's
     else and the last reviewer's next -> review_guardian (always-on, last).

WHAT IT NEVER TOUCHES (gate-specific machinery, deliberately divergent):
  persist_submission, select_panel's plan_field/extra_text_fields, route_checks,
  compose_verdict*, append_verdict, route_*, complete_*. And it does not mirror
  fix-proposer's code_lookup / repropose / reframe / escalate: those serve the
  blind reproposer, which the gate has no equivalent of (its authors read the
  objections themselves).

USAGE (read-only by default — always dry-run first):
    python3 099_SYNC_gate_roster.py            # dry run: report the delta
    python3 099_SYNC_gate_roster.py --apply    # snapshot, then write

Safety: --apply runs snapshot_agent('council-gate', ...) first, so the prior row
is restorable from agent_definitions_backup, and the UPDATE is parameterised.
"""
import json
import subprocess
import sys

NS = "ai-persona-system"
POD = "postgres-clients-0"
DB = ["-U", "clients_user", "-d", "clients_db"]

ALWAYS_ON = {"review_editquality", "review_guardian"}


def psql(sql, stdin_payload=None):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--", "psql", *DB, "-tA", "-v", "ON_ERROR_STOP=1"]
    cmd += ["-c", sql] if stdin_payload is None else ["-f", "-"]
    r = subprocess.run(cmd, input=stdin_payload, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}")
    return r.stdout.strip()


def live_workflow(agent_type):
    out = psql(
        "SELECT default_config->'workflow' FROM agent_definitions "
        f"WHERE type='{agent_type}' AND is_active "
        "AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
    )
    if not out:
        sys.exit(f"no live {agent_type} row found")
    return json.loads(out)


def transform_step(name, step):
    """Apply the four transforms to one mirrored reviewer/gate step."""
    s = json.loads(json.dumps(step))  # deep copy
    cfg = s.get("config", {})

    if cfg.get("error_step") == "complete_refused":
        cfg["error_step"] = "complete_invalid"

    if "input_fields" in cfg:
        cfg["input_fields"] = ["input_data" if f == "diagnosis_row" else f
                               for f in cfg["input_fields"]]

    p = cfg.get("prompt_template")
    if isinstance(p, str):
        p = p.replace("## The diagnosis", "## The author's stated rationale")
        p = p.replace("{{.diagnosis_row.conclusion}}", "{{.input_data.rationale}}")
        cfg["prompt_template"] = p

    s["config"] = cfg
    return s


def main():
    apply = "--apply" in sys.argv
    fp = live_workflow("fix-proposer")
    gate = live_workflow("council-gate")
    fp_steps, gate_steps = fp["steps"], gate["steps"]

    mirrored = {k: v for k, v in fp_steps.items()
                if k.startswith("review_") or k.startswith("gate_")}
    if not mirrored:
        sys.exit("fix-proposer has no review_/gate_ steps — refusing to wipe the gate")

    before = sorted(k for k in gate_steps if k.startswith("review_"))
    after = sorted(k for k in mirrored if k.startswith("review_"))
    added, removed = sorted(set(after) - set(before)), sorted(set(before) - set(after))

    new_steps = {k: v for k, v in gate_steps.items()
                 if not (k.startswith("review_") or k.startswith("gate_"))}
    for name, step in mirrored.items():
        new_steps[name] = transform_step(name, step)

    # Chain endpoints: the gate enters the reviewer chain from persist_submission
    # -> select_panel -> review_editquality, exactly as fix-proposer does from
    # persist_plan, so the internal chain carries over verbatim. Only re-assert
    # the entry and exit, in case fix-proposer's differ.
    first_gate = fp_steps["review_editquality"].get("next_step")
    if first_gate:
        new_steps["review_editquality"]["next_step"] = first_gate
    new_steps["select_panel"]["next_step"] = "review_editquality"
    new_steps["review_guardian"]["next_step"] = "council_decide"

    # Mirror the footprint map (relevance signal), keeping the gate's own
    # plan_field / extra_text_fields (a rationale, not a diagnosis).
    fp_panel = fp_steps.get("select_panel", {}).get("config", {})
    if "footprints" in fp_panel:
        new_steps["select_panel"]["config"]["footprints"] = fp_panel["footprints"]

    new_steps["council_decide"]["config"]["review_fields"] = \
        fp_steps["council_decide"]["config"]["review_fields"]
    if "run_checks" in new_steps and "run_checks" in fp_steps:
        new_steps["run_checks"]["config"]["check_fields"] = \
            fp_steps["run_checks"]["config"]["check_fields"]

    # Integrity: every routing target must exist.
    targets = set()
    for s in new_steps.values():
        for key in ("next_step",):
            if s.get(key):
                targets.add(s[key])
        for key in ("then_step", "else_step", "error_step"):
            if s.get("config", {}).get(key):
                targets.add(s["config"][key])
    dangling = sorted(t for t in targets if t not in new_steps)
    if dangling:
        sys.exit(f"REFUSING: mirrored workflow has dangling targets: {dangling}")

    seats = len(new_steps["council_decide"]["config"]["review_fields"])
    print(f"fix-proposer seats: {len(after)} | council-gate seats before: {len(before)}")
    print(f"  added:   {added or '(none)'}")
    print(f"  removed: {removed or '(none)'}")
    print(f"  footprints: {sorted(new_steps['select_panel']['config'].get('footprints', {}))}")
    print(f"  resulting steps: {len(new_steps)}, review_fields: {seats}, routing: OK")

    if not apply:
        print("\nDRY RUN — nothing written. Re-run with --apply to write.")
        return
    if not added and not removed:
        print("\nAlready in sync; --apply would be a no-op. Nothing written.")
        return

    gate["steps"] = new_steps
    payload = json.dumps({"workflow": gate})
    # Dollar-quoting, not \set: the payload holds quotes, backslashes and
    # newlines that psql's \set mangles. The tag must not occur in the payload.
    tag = "$cfgsync$"
    if tag in payload:
        sys.exit("REFUSING: payload contains the dollar-quote tag")
    sql = (
        "BEGIN;\n"
        "SELECT snapshot_agent('council-gate', 'pre-update: 099 roster sync from live fix-proposer');\n"
        f"UPDATE agent_definitions SET default_config = {tag}{payload}{tag}::jsonb, updated_at = now() "
        "WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false "
        "AND deleted_at IS NULL;\n"
        "COMMIT;\n"
    )
    psql(None, stdin_payload=sql)
    print(f"\nAPPLIED: council-gate now mirrors {len(after)} seats (snapshot taken).")


if __name__ == "__main__":
    main()
