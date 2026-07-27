#!/usr/bin/env python3
"""patch-feature-designer-context.py — close the gap the 2026-07-27 audit found.

THE GAP. D8a′ patched `fix-proposer`, and the 099 mirror carried it to
`council-gate`. But the mirror only spans those two — `feature-designer` has its
own roster and is mirrored by nothing. So after both builds landed, the audit read:

    feature-designer | review_bug_historian | minutes=f | case_index=f
    feature-designer | review_guardian      | minutes=f | case_index=f

That is incoherent with the argument this whole workstream rests on. D2 seats the
new architecture reviewer at `feature-designer` precisely BECAUSE it is the
earliest point platform code takes shape and the place the guardian sits with the
fewest counterweights. Having argued the design stage matters most, leaving its
guardian without its own minutes — and without the deflection check that is the
entire operational payload of the D5 measurement — equips the most important
council worst.

So this gives feature-designer's guardian and bug historian the same context their
fix-proposer counterparts now have: the council's own 259 verdicts, the guardian's
deflection count, and the generated case index.

Dry run by default; --apply writes /tmp/acm/feature-designer-CONTEXT.json.
"""
import argparse
import importlib.util
import json
import pathlib
import subprocess
import sys

MINUTES = """
PRIOR COUNCIL HISTORY — your own minutes.
Every verdict this council has ever reached is stored in `diagnosis_artifacts` (already in
your Schema list) as `kind='council_report'`. `body` is JSON:
{"decision","decided_by","reviews":[{"reviewer","verdict","objections":[{"edit","problem","severity"}]}]}.
Sibling rows carry `kind='fix_plan'` (the plan as submitted). Use a check to ask whether this
file, symbol or mechanism has been before the council already and what was said — precedent you
agree with is the cheapest possible objection, and precedent you contradict is worth naming.
  SELECT created_at, left(body,4000) FROM diagnosis_artifacts
   WHERE kind='council_report' AND body ILIKE '%<symbol_or_file>%'
   ORDER BY created_at DESC LIMIT 5;
Treat an empty result as "no precedent found", NOT as "this is novel" — the corpus starts
2026-07 and the ILIKE only matches text a past reviewer happened to quote.
"""

GUARDIAN_EXTRA = """
STABILITY PREFERENCE — CHECK YOUR OWN DEFLECTIONS BEFORE YOU REPEAT ONE. Measured 2026-07-27
across 259 past reports: the guardian seat has issued 437 objections, 29 of them invoking this
preference, and they concentrate on a few sites — `coordinator.go`/`SagaCoordinator.ProcessResponse`
was deflected upward across SIX distinct submissions in seven days, `spawn_actions.go` four. A core
site that keeps returning is evidence the higher-layer fixes are NOT holding, and repeating the
deflection a seventh time is not caution, it is a loop. So when you are about to prefer a fix at
a higher layer, first check how often this site has already been sent upward:
  SELECT count(*) FROM diagnosis_artifacts WHERE kind='council_report'
   AND body ILIKE '%<the core symbol>%' AND body ILIKE '%higher%layer%';
If the count is non-trivial, say so in notes and either name a higher-layer alternative that has
NOT already been tried and refuted, or concede that the core is the right place and object on
containment instead.
"""

MARK = "## Schema (the ONLY tables available to checks)"


def psql(sql: str) -> str:
    r = subprocess.run(
        ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
         "psql", "-U", "clients_user", "-d", "clients_db", "-A", "-t", "-c", sql],
        capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr}")
    return r.stdout


def splice(prompt: str, block: str) -> str:
    if MARK in prompt:
        return prompt.replace(MARK, block.strip() + "\n\n" + MARK, 1)
    return prompt.rstrip() + "\n\n" + block.strip() + "\n"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true")
    a = ap.parse_args()

    # Reuse the index builder rather than duplicating the corpus logic — the two
    # must never drift, or feature-designer's historian sees a different history
    # from fix-proposer's. (Loaded by path because the filename is hyphenated.)
    p = pathlib.Path(__file__).resolve().parent / "build-historian-index.py"
    spec = importlib.util.spec_from_file_location("bhi", p)
    m = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(m)
    bug_index, _ = m.build_blocks()

    cfg = json.loads(psql(
        "SELECT default_config::text FROM agent_definitions WHERE type='feature-designer' "
        "AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;").strip())
    steps = cfg["workflow"]["steps"]

    plan = {
        "review_guardian": MINUTES + GUARDIAN_EXTRA,
        "review_bug_historian": MINUTES + "\n" + bug_index,
    }

    for seat, block in plan.items():
        if seat not in steps:
            sys.exit(f"missing seat {seat} — read the roster before patching")
        p_before = steps[seat]["config"]["prompt_template"]
        if "council_report" in p_before:
            print(f"  {seat}: already carries its minutes — skipping")
            continue
        steps[seat]["config"]["prompt_template"] = splice(p_before, block)
        print(f"  {seat}: {len(p_before):,} -> "
              f"{len(steps[seat]['config']['prompt_template']):,} chars")

    if not a.apply:
        print("\nDRY RUN — nothing written. Re-run with --apply.")
        return
    out = pathlib.Path("/tmp/acm/feature-designer-CONTEXT.json")
    out.write_text(json.dumps(cfg, ensure_ascii=False, separators=(",", ":")))
    print(f"\nwrote {out} ({out.stat().st_size:,} bytes)")


if __name__ == "__main__":
    main()
