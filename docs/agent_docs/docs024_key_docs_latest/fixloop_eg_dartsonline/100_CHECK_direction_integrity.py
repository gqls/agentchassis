#!/usr/bin/env python3
"""100_CHECK_direction_integrity.py — D3 of the direction drift guard
(PLAN_2026-07-20_direction_reach_and_drift_guard.md). READ-ONLY, always.

Checks the three surfaces the direction can drift on:
  1. FILES: sha256 of each blessed doc vs the DIRECTION_LEDGER hash — a mismatch
     means an edit got past/around the D2 commit gate (or the ledger wasn't
     updated in the same commit, which is the same offence).
  2. COPIES: every sanctioned copy byte-identical to its canonical.
  3. SEATS: the constitution/mission (and librarian) seat prompts in BOTH live
     councils still carry their blessed rule anchors — catches live DB edits to
     the seats (agent_definitions is writable by any thread; this session has
     WATCHED a seat prompt change under it mid-query).

Exit 0 all green; exit 1 on any drift (report printed either way).
Run: manually after ANY seat migration (runbook: right after the 099 step), and
whenever the ledger changes. No DB writes, no filesystem writes, ever.
"""
import hashlib
import json
import subprocess
import sys
import os

REPO = os.path.dirname(os.path.abspath(__file__))
for _ in range(4):
    REPO = os.path.dirname(REPO)

LEDGER = {
    "docs/agent_docs/docs024_key_docs_latest/adoption/docubundle/thin_slice_constitution.md": "18453e8cac84bdfe",
    "docs/agent_docs/docs024_key_docs_latest/028_platform_mission_and_pipeline_direction(2).md": "c6aa949edab8e44a",
}
COPIES = {
    "scripts/documentation_project/02/thin_slice_constitution.md":
        "docs/agent_docs/docs024_key_docs_latest/adoption/docubundle/thin_slice_constitution.md",
    "docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/thin_slice_constitution.md":
        "docs/agent_docs/docs024_key_docs_latest/adoption/docubundle/thin_slice_constitution.md",
    "docs/agent_docs/docs014_documentation_collection/028_platform_mission_and_pipeline_direction(2).md":
        "docs/agent_docs/docs024_key_docs_latest/028_platform_mission_and_pipeline_direction(2).md",
}
# Seat prompt anchors: stable rule fragments that MUST be present. If a seat is
# reworded legitimately, update these anchors in the same reviewed change.
SEAT_ANCHORS = {
    "review_constitution": ["FIX THE CAUSE, NOT THE SYMPTOM", "REUSE BEFORE RECREATE"],
    "review_mission": ["THE REVENUE MODEL SHAPES THE SITE", "SILENT OVERRIDE IS THE FAILURE MODE"],
    "review_prior_art": ["ASSERTED-ABSENCE", "DORMANT-MACHINERY"],
}
COUNCILS = ["fix-proposer", "council-gate"]

def sha16(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        h.update(f.read())
    return h.hexdigest()[:16]

def psql(sql):
    r = subprocess.run(
        ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
         "--", "psql", "-U", "clients_user", "-d", "clients_db", "-tA", "-c", sql],
        capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed (read-only check could not run): {r.stderr.strip()}")
    return r.stdout.strip()

def main():
    failures = []

    for rel, blessed in LEDGER.items():
        p = os.path.join(REPO, rel)
        if not os.path.isfile(p):
            failures.append(f"FILES: blessed doc MISSING: {rel}")
            continue
        actual = sha16(p)
        status = "ok" if actual == blessed else "DRIFT"
        print(f"  [{status}] {rel}  ledger={blessed} actual={actual}")
        if actual != blessed:
            failures.append(f"FILES: {rel} hash {actual} != ledger {blessed} — unsigned edit or stale ledger")

    for copy, canonical in COPIES.items():
        pc, pk = os.path.join(REPO, copy), os.path.join(REPO, canonical)
        if not os.path.isfile(pc):
            failures.append(f"COPIES: sanctioned copy MISSING: {copy}")
            continue
        same = open(pc, "rb").read() == open(pk, "rb").read()
        print(f"  [{'ok' if same else 'DRIFT'}] copy {copy}")
        if not same:
            failures.append(f"COPIES: {copy} differs from canonical {canonical}")

    for council in COUNCILS:
        out = psql(
            "SELECT default_config->'workflow'->'steps' FROM agent_definitions "
            f"WHERE type='{council}' AND is_active "
            "AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;")
        if not out:
            failures.append(f"SEATS: no live {council} row")
            continue
        steps = json.loads(out)
        for seat, anchors in SEAT_ANCHORS.items():
            prompt = (steps.get(seat, {}).get("config", {}) or {}).get("prompt_template", "")
            if not prompt:
                failures.append(f"SEATS: {council}.{seat} MISSING or has no prompt")
                print(f"  [DRIFT] {council}.{seat}: missing")
                continue
            gone = [a for a in anchors if a not in prompt]
            print(f"  [{'ok' if not gone else 'DRIFT'}] {council}.{seat}")
            if gone:
                failures.append(f"SEATS: {council}.{seat} lost anchor(s): {gone}")

    print()
    if failures:
        print("DIRECTION DRIFT DETECTED:")
        for f in failures:
            print(f"  - {f}")
        print("\nIf the change was owner-approved: update DIRECTION_LEDGER.md (same")
        print("commit, Direction-Approved trailer) or the seat anchors above, via a")
        print("reviewed change. If it was not: treat as drift — restore from the")
        print("canonical file / agent_definitions_backup snapshot and investigate.")
        sys.exit(1)
    print("ALL GREEN: files match ledger, copies match canonicals, seat anchors present in both councils.")

if __name__ == "__main__":
    main()
