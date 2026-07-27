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

D3: advisory, NO veto — and the mechanism for that is NOT the hard_veto_from list.

  CORRECTED 2026-07-27, before this ever ran. The first draft asserted the seat
  was advisory because it is absent from `council_decide.hard_veto_from`, and gave
  it the verdict vocabulary point_fix|needs_rfc|insufficient. Both were wrong, and
  the second would have broken the feature-build lane outright:

    * `diagnose_council_decide_action.go:13-14` — ANY reviewer's veto rejects the
      round; `hard_veto_from` only changes the audit label. Absence from that list
      buys nothing. What actually makes a seat advisory is that its prompt never
      offers `veto` — the same reasoning already written down for the bug historian
      in docs026_concept_register/PILOT_bug_historian_reviewer.md §2.
    * `:160` recognises exactly {approve, object, veto}. `:397` records anything
      else UNREADABLE, and `:446` downgrades any would-be approval that has an
      unreadable seat to REVISE. So the invented vocabulary would have forced every
      feature-designer round to revise, exhaust max_rounds=3, and fail — a seat that
      breaks the lane it was added to help.

  Caught by the owner pointing at the concept register, whose pilot doc states the
  veto behaviour plainly. The cheap check skipped: read the decider before inventing
  a verdict vocabulary for it.

So: verdict is approve|object, never veto. The forward argument is carried by
objection SEVERITY (high gates a round, low/medium are recorded but advisory).

  CORRECTED AGAIN 2026-07-27, after the seat was already live. The fix above put
  the RFC routing in a separate `architecture_signal` JSON field. That field is
  DISCARDED: `councilReview` (same file, :84-95) marshals only
  {reviewer, verdict, objections, missing, notes, degraded}, so every custom field
  a seat emits is dropped when the report is persisted — confirmed against 2,138
  stored reviews, where `checks` and `code_checks` do not appear either.
  The seat would have run correctly and its distinctive output would have gone
  nowhere. The signal now goes in the FIRST LINE of `notes`, in a fixed greppable
  shape, which is both what survives and what makes the seat measurable.
  Caught by building the adoption report — i.e. by asking "how would I ever know
  if this worked?", which is a question worth asking before shipping, not after.

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

YOUR VERDICT VOCABULARY IS "approve" OR "object". NEVER "veto".

The council's decider recognises exactly three verdicts, and treats ANY reviewer's
veto as an outright rejection regardless of which seats are nominally the veto
holders. So a veto from you would make you a second gatekeeper, which is precisely
what this seat must not be — the guardian holds the block; you hold the argument.
Anything outside approve/object is recorded as UNREADABLE and downgrades the whole
round to revise, so it is not an option either.

Severity is what carries your meaning, because HIGH gates a round and low/medium
are advisory-but-recorded:
- "approve" — the existing design carries this work. The most common correct answer.
- "object", severity MEDIUM — this is an architecture change whether or not it is
  dressed as a fix, OR the architecture underneath is insufficient even though the
  plan itself should proceed. Recorded and returned to the author without blocking.
  Say which of the two you mean in the ARCHITECTURE_SIGNAL line of your notes.
- "object", severity HIGH — reserve this. Use it only when proceeding would make
  the architecture materially harder to change later: a contract about to be
  depended on, a workaround about to be built on top of a workaround. This DOES
  force a revise round, so spend it rarely and say exactly what becomes irreversible.

YOUR ROUTING SIGNAL MUST GO IN `notes`, AS THE FIRST LINE, IN THIS EXACT SHAPE:

  ARCHITECTURE_SIGNAL: point_fix|needs_rfc|insufficient | DEFLECTIONS: <n or unknown>

then a blank line, then your prose. This is not decoration. The council persists
only {reviewer, verdict, objections, missing, notes} — every other field you emit is
DISCARDED when the report is stored, so a signal placed anywhere else is written to
nothing and no one will ever read it. The first line is what makes your verdict
findable later, and what lets us measure whether this seat earns its place.

- "point_fix"    — constrained; proceed normally.
- "needs_rfc"    — meets the architecture-review trigger test; deserves an RFC
                   (architecture_review/PROCESS_architecture_review.md) with blast
                   radius, staged rollout and rollback written down.
- "insufficient" — the plan is fine, the architecture under it is not, and you want
                   that on the record.

DEFLECTIONS is the count from (d), or `unknown` if you did not or could not check.
Write `unknown` honestly rather than guessing a number — a fabricated count here is
worse than no count, because it will be read as measurement.

After that first line, use the prose in `notes` to carry what the schema will not:
the specific future load you judged against, the cost of not changing, and which
trigger condition fired if any. Objections carry anything you want the author to act
on; notes carry your reasoning.

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
{"reviewer": "architecture", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "..."}], "code_checks": [{"kind": "symbol|content|ls", "query": "pattern", "why": "..."}], "notes": "ARCHITECTURE_SIGNAL: <point_fix|needs_rfc|insufficient> | DEFLECTIONS: <n|unknown>\\n\\n<your reasoning: the future load you judged against, the cost of not changing, which trigger fired>"}

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

    # What actually makes this seat advisory is that its prompt never offers a
    # veto, and that every verdict it CAN emit is one the decider recognises.
    # hard_veto_from is only an audit label (diagnose_council_decide_action.go:13),
    # so asserting on it would be theatre. Assert the two things that bite:
    veto = dec.get("hard_veto_from", [])
    assert '"verdict": "approve|object"' in PROMPT, \
        "prompt must offer ONLY approve|object — an unrecognised verdict is read " \
        "as UNREADABLE and downgrades the round to revise (decide action :160/:397/:446)"
    assert "NEVER \"veto\"" in PROMPT, "the seat must be told explicitly never to veto (D3)"

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
