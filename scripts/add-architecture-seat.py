#!/usr/bin/env python3
"""add-architecture-seat.py — seat `review_architecture` on feature-designer.

WHY HERE, AND WHY ADVISORY (decisions D1/D2/D3, owner-approved 2026-07-27).

The council already has a seat that argues the conservative side: the guardian,
the only hard-veto holder, told by its charter to protect long-stable
infrastructure and to prefer a fix at a higher, less-foundational layer. Nothing
argues the other side — that a design will not carry us where we are going, and
that not changing has a cost too.

D1: this seat argues the FORWARD side only. A single seat holding both remits
collapses into the conservative one, because "this is battle-tested, don't touch
it" is always cheaper to evidence than "this won't carry us in three months".

D2: it sits on feature-designer, not the fix gate. That is the earliest point
platform code takes shape, and the place where the guardian currently sits with
the fewest counterweights — four colleagues, none of them looking forward.

D3: advisory, NO veto. It is deliberately absent from hard_veto_from. Its output
is an RFC trigger (needs_rfc / point_fix), not an approve/object, so it cannot
deadlock against the guardian; two veto-holders would make the gate unpassable.

D5 is why it exists at all: measured 2026-07-27, coordinator.go/ProcessResponse
was deflected upward by six distinct submissions in seven days while the file
itself moved nine times in sixty days. Pressure on the core is high, change to
the core is near zero, and the difference is landing as workarounds above it.

Wiring: review_guidelines -> review_architecture -> review_guardian. The seat is
inserted immediately BEFORE the guardian so the forward argument is on the record
in the same round the veto is considered.

Dry run by default; --apply writes /tmp/acm/feature-designer-SEATED.json for the
owner to push (writing to live config is not available to this process).
"""
import argparse
import json
import pathlib
import subprocess
import sys

SEAT = "review_architecture"
PROMPT = """# Council reviewer: ARCHITECTURE (forward fitness) — ADVISORY, NO VETO

You are the only seat that argues the COST OF NOT CHANGING. Every other seat's
default is caution; yours is not. You change nothing; you judge one question:

  Is the architecture this plan builds on SUFFICIENT for what we are trying to do
  next — and if it is not, is this plan the moment to say so?

You are not a second guardian. The guardian protects battle-tested infrastructure
and holds the only veto; it is right far more often than not. But it has no
instrument for benefit, and a system where only caution has a voice ossifies:
pressure to change the core gets deflected upward and re-emerges as workarounds in
the layer above. Measured on this platform 2026-07-27 — one orchestrator file was
sent "fix it higher up" by SIX independent submissions in seven days while the file
itself changed nine times in sixty days. That is the failure mode you exist to catch.

JUDGE, in this order:

(a) FORWARD FITNESS. Does the design this plan extends carry the work we can
    already see coming, or does the plan add another accommodation to something
    that has stopped fitting? Name the specific future load you think it will or
    will not carry. Vague "this may not scale" is not an argument.

(b) COST OF NOT CHANGING. If this plan takes the contained route, what is the
    running cost — repeated workarounds, a defect class that stays reachable, a
    capability that stays blocked? Quantify it if the DB can (failure counts,
    recurrence), and say plainly when you cannot.

(c) IS THIS A POINT FIX OR AN ARCHITECTURE CHANGE? Apply the architecture-review
    track's trigger test: a shared contract (dedupe key, delivery guarantee, a
    state machine consumed by more than one package, a wire/message shape); an
    exported symbol other packages depend on; coordinated edits across three or
    more top-level platform packages; or a change needing a staged rollout because
    the change and its verification cannot fit in one deploy step.

(d) HAS THIS SITE BEEN DEFLECTED BEFORE? If the plan touches a core site, check
    how many times the council has already sent that site upward. A site that keeps
    returning is evidence the higher-layer fixes are NOT holding — say so, because
    the guardian cannot see its own history unless someone names it.

YOUR VERDICT IS A ROUTING DECISION, NOT AN OBJECTION:
- "point_fix"  — constrained; the existing design carries it; proceed normally.
- "needs_rfc"  — this is an architecture change, whether or not it is dressed as a
                 fix. Say which trigger condition it meets. This does NOT block:
                 it says the change deserves an RFC in the architecture-review track
                 (architecture_review/PROCESS_architecture_review.md), with blast
                 radius, staged rollout and rollback written down.
- "insufficient" — the plan is fine but the architecture underneath it is not, and
                 you want that on the record even though this plan should proceed.

Be concrete or be quiet. An unevidenced "we should redesign this" is worse than
silence — it spends the one voice arguing for change on nothing. If the contained
fix is genuinely right, say "point_fix" and move on; that is a real answer and the
most common correct one.

CHECKS: if your verdict hinges on a fact a read-only SQL query could settle — how
often a failure actually occurs, whether a work-item class is still accumulating,
how many past council rounds touched this file — put it in checks as
{"sql": "SELECT ...", "why": "..."}. SELECT/WITH only, never writes. Write checks
ONLY against the tables in the Schema section below.

PRIOR COUNCIL HISTORY — your own minutes. `diagnosis_artifacts` (in your Schema
list) holds every past verdict as kind='council_report'; `body` is JSON with
{"decision","decided_by","reviews":[{"reviewer","verdict","objections":[...]}]}.
This is how you answer (d) with a number instead of an impression:
  SELECT count(*) FROM diagnosis_artifacts WHERE kind='council_report'
   AND body ILIKE '%<the core symbol>%' AND body ILIKE '%higher%layer%';
Treat an empty result as "no precedent found", NOT as "this is novel".

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The owner-approved capability spec this design serves
{{.spec_row.spec_json}}

## The author's stated rationale
{{.input_data.rationale}}

## The plan
{{.plan_persisted.plan_json}}

## Output — ONLY this JSON
{"reviewer": "architecture", "verdict": "point_fix|needs_rfc|insufficient", "trigger": "which trigger condition, or null", "future_load": "the specific coming work you judged against", "cost_of_not_changing": "...", "deflection_count": "n or unknown", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "checks": [{"sql": "SELECT ...", "why": "..."}], "code_checks": [{"kind": "symbol|content|ls", "query": "pattern", "why": "..."}], "notes": "..."}

CODE QUESTIONS (code_checks): when your verdict hinges on a fact about the CODEBASE
— does another implementation exist, which files carry symbol X, does anything still
reference Y — attach a code_checks entry rather than judging blind. kind "symbol"
matches symbol names, "content" searches source bodies, "ls" lists indexed paths.
Answered from the code_symbols index next round. SQL cannot see the codebase and
code_checks cannot see the database — use each for its own tier."""


def psql(sql: str) -> str:
    r = subprocess.run(
        ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
         "psql", "-U", "clients_user", "-d", "clients_db", "-A", "-t", "-c", sql],
        capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr}")
    return r.stdout


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true", help="write the patched config file")
    a = ap.parse_args()

    cfg = json.loads(psql(
        "SELECT default_config::text FROM agent_definitions WHERE type='feature-designer' "
        "AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;").strip())
    steps = cfg["workflow"]["steps"]

    if SEAT in steps:
        sys.exit(f"{SEAT} already seated — refusing to double-add")
    for req in ("review_guidelines", "review_guardian", "council_decide"):
        if req not in steps:
            sys.exit(f"missing expected step {req} — chain has changed, re-read before patching")

    # Model/plumbing copied from the guardian so the new seat cannot drift from
    # the council's own conventions (provider, temperature, error routing).
    tmpl = json.loads(json.dumps(steps["review_guardian"]))
    tmpl["config"]["prompt_template"] = PROMPT
    tmpl["config"]["input_fields"] = sorted(set(
        tmpl["config"].get("input_fields", []) + ["input_data", "plan_persisted", "schema_hint", "spec_row"]))
    tmpl["description"] = ("Council reviewer — ARCHITECTURE (forward fitness): is the design "
                           "sufficient for anticipated work, and is this a point fix or an "
                           "architecture change? ADVISORY, no veto; verdict routes to the RFC track.")
    tmpl["output_field"] = SEAT
    tmpl["next_step"] = "review_guardian"
    steps[SEAT] = tmpl

    old_next = steps["review_guidelines"].get("next_step")
    steps["review_guidelines"]["next_step"] = SEAT

    dec = steps["council_decide"]["config"]
    fields = dec.get("review_fields", [])
    if f"{SEAT}.result" not in fields:
        fields.insert(fields.index("review_guardian.result"), f"{SEAT}.result")
    dec["review_fields"] = fields

    veto = dec.get("hard_veto_from", [])
    assert "architecture" not in veto, "the architecture seat must NEVER hold a veto (D3)"

    print(f"chain:  review_guidelines -> {SEAT} -> {steps[SEAT]['next_step']}"
          f"   (was review_guidelines -> {old_next})")
    print(f"seats:  {len(fields)} review_fields, hard_veto_from={veto}")
    print(f"prompt: {len(PROMPT):,} chars")
    print(f"steps:  {len(steps)}")

    if not a.apply:
        print("\nDRY RUN — nothing written. Re-run with --apply.")
        return
    out = pathlib.Path("/tmp/acm/feature-designer-SEATED.json")
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(cfg, ensure_ascii=False, separators=(",", ":")))
    print(f"\nwrote {out} ({out.stat().st_size:,} bytes)")


if __name__ == "__main__":
    main()
