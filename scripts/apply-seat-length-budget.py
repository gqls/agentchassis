#!/usr/bin/env python3
"""apply-seat-length-budget.py — put a LENGTH BUDGET into the council seats that are
measurably close to their token cap. bugs_open/138 fix candidate 4.

WHY A SCRIPT AND NOT SIX HAND EDITS. The block below is the single source of truth
for its own text. Six identical blocks typed into six live prompts is the drift class
099 (roster mirror) and 102 (seat parity) both exist to fight, and the seat list will
grow — the report that chose it (104_REPORT_seat_token_pressure_v1.sh) re-runs and
will flag more. Adding a target here is one line; re-running is idempotent.

SCOPE: EVERY ELIGIBLE SEAT (owner decision, 2026-07-31). Superseded the original
narrow scope, and the reason it changed is the useful part.

  It began as: "apply where the pressure is MEASURED, not everywhere it is
  imaginable", following the owner's criterion for the sibling change (cap raises,
  2026-07-29 — raise the seats that actually truncate, leave the rest until they do).
  The stated objection to going wider was that a blanket edit across 51 templates
  changes what every council asks its reviewers for, and that this was "not supported
  by the evidence".

  It is now supported by the evidence, which is why the scope changed rather than the
  argument being dropped. Measured against a real control arm on 2026-07-31, by ROUND
  SPAWN TIME: `review_editquality` @16000 went from peak 98.3% of cap (10 rounds
  spawned before the block, mean 9,848 tokens) to peak 55.0% (8 rounds after, mean
  6,569) — same seat, same afternoon, CAP UNCHANGED. That isolates the budget from
  the cap raise it originally shipped with, which the `review_architecture` result
  never could. A change with a measured effect and no observed cost is a different
  proposition from the same change argued for.

  Note what did NOT change: eligibility is still a claim about the mechanism (see
  ELIGIBILITY_SQL below), and the block still refuses to overwrite a hand-authored
  one. "All eligible" is not "all".

WHAT CANDIDATE 4 TURNED OUT TO BE, AND WHAT IT IS NOT — measured, not reasoned.
Candidate 4 was filed as "emit the load-bearing field FIRST in every seat's output
schema, because truncation eats the tail". The premise is right and the fleet-wide
prescription is already satisfied:
  * `reviewer` and `verdict` are FIRST in all 51 live templates, which is exactly why
    salvage works at all (salvageTruncatedReview's own comment says so).
  * `severity` sits LAST inside each objection, which looks like the same bug one
    level down — a cut mid-`problem` would lose the grade and an ungraded objection
    GATES. Measured: 0 of 2,713 stored objections, degraded or not, is missing its
    severity. The repair either keeps a whole objection or drops it entirely, so
    reordering the objection fields would fix nothing. REFUTED.
  * The field truncation actually destroys is `notes`: present in 2 of 30 degraded
    reviews (6.7%) against 3,067 of 3,076 complete ones (99.7%).
  * And for 49 of 51 seats `notes` is the RIGHT thing to lose. Moving it to the head
    would put `objections` in the tail instead — and objections carry both the
    severities the gate reads and the content the proposer revises against, surviving
    80% of truncations today. There is no ordering that saves everything; the current
    one already sacrifices the cheapest field. `review_architecture` is the exception
    BECAUSE its mandated ARCHITECTURE_SIGNAL lives in notes, which is what made that
    seat unmeasurable when it truncated — the seat's own remit decides, not a rule.
  * The guardian-veto case (its prompt requires the contained alternative in `notes`)
    is the one real ordering risk left: 15 vetoes all-history, 0 degraded, 0 with
    empty notes. Not worth reordering for zero observed instances; recorded so the
    next reader does not re-derive it.
So what generalises is NOT the reorder — it is the LENGTH BUDGET, the other half of
the architecture-seat fix, and the half the evidence credits: after it, that seat's
outputs got SHORTER (peak 4,443 tokens, 28% of its new cap) rather than merely
having a higher ceiling.

DELIBERATELY NOT CAPPING OBJECTIONS. The architecture seat's hand-written block says
"at most 3 objections". Generalising that would trade truncation risk for coverage
loss across every council — a seat that drops a real objection to fit the budget has
failed in a way nobody can see. This block budgets PROSE and says explicitly: cut
the words, never the findings.

Usage:
    ./scripts/apply-seat-length-budget.py              # dry run (default)
    ./scripts/apply-seat-length-budget.py --apply      # snapshot each agent, then write
    ./scripts/apply-seat-length-budget.py --verify     # report live state only
"""

import argparse
import subprocess
import sys

NS = "ai-persona-system"
POD = "postgres-clients-0"

# ---------------------------------------------------------------------------
# SCOPE, from 2026-07-31: EVERY ELIGIBLE SEAT, discovered from the live DB.
#
# Owner decision 2026-07-31, on evidence rather than argument: the block took
# review_editquality from 98.3% to 55.0% of an UNCHANGED cap (control arm by round
# spawn time, 10 rounds before vs 8 after). So it stops being a per-seat remedy and
# becomes the default for the seats the mechanism actually governs.
#
# DISCOVERED, NOT LISTED, and that is the point. A hand-written roster is the exact
# drift this platform keeps paying for: 102_LINT exists because a 16th seat was added
# and one key was forgotten on it, and 099 exists because two rosters had to be kept
# identical by hand. A list of 48 pairs would be stale the first time a seat is
# seated. This asks the database instead, every run.
#
# ELIGIBILITY IS A CLAIM ABOUT THE MECHANISM, NOT A CONVENIENCE. The block states
# that a DEGRADED `object` gates the round to REVISE regardless of severities. That
# is true only where `diagnose_council_decide` is the decider, so eligibility is
# defined by the council HAVING that step — measured, not assumed. Five councils
# qualify (council-gate, fix-proposer, feature-designer, experience-planner,
# experience-approval-council); `domain-research-classifier` has zero decide steps
# and a different output schema entirely, so the block would be a FALSE claim in its
# prompt. Putting text a reviewer will act on into a prompt where it does not hold is
# worse than leaving the seat uncovered.
ELIGIBILITY_SQL = """
SELECT a.type, s.key
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.key LIKE 'review_%'
  AND s.value->'config'->>'prompt_template' IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM jsonb_each(a.default_config->'workflow'->'steps') d
    WHERE d.value->>'action' = 'diagnose_council_decide')
ORDER BY 1, 2;
"""

# Seats deliberately left out, with the reason. An exclusion with no reason is how a
# gap becomes folklore; these are printed on every run so they stay arguable.
EXCLUSIONS = {
    ("domain-research-classifier", "review_mission_alignment"):
        "council has NO diagnose_council_decide step and a different output schema "
        "(objection_found/concerns/note) — the block's central claim would be false here",
}

# Seats whose pressure was MEASURED before the fleet-wide rollout. Kept because the
# evidence is the interesting part of the history, and 104_REPORT reprints it — not
# because these are the only targets any more.
MEASURED = [
    ("fix-proposer",     "review_guardian",
     "peak 99.2% of an 8000 cap over 278 calls (118 labelled with a holding council) — the highest-pressure seat with attributable evidence"),
    ("council-gate",     "review_guardian",
     "same seat, mirrored roster (099 keeps these two identical; a block on one only would BE the drift)"),
    ("feature-designer", "review_guardian",
     "third council holding this seat at 8000; synced by nothing, so it must be named explicitly"),
    ("fix-proposer",     "review_improvement_guardian",
     "peak 96.6% of 8000 over 89 calls (34 attributable)"),
    ("council-gate",     "review_improvement_guardian",
     "same seat, mirrored roster"),
    ("fix-proposer",     "review_debug_historian",
     "ADDED 2026-07-31 by the alert, not by hand: peak 99.8% of 8000 over 283 calls "
     "(128 attributable) at a p95 of only 62.2% — a long thin tail, and the case the "
     "near-miss threshold exists for. A p95 rule would never have seen it"),
    ("council-gate",     "review_debug_historian",
     "same seat, mirrored roster"),
    ("fix-proposer",     "review_editquality",
     "ADDED 2026-07-31: peak 98.3% of its RAISED 16000 cap over 52 calls, ALL 52 "
     "attributable, and rising — 13,115 -> 13,592 -> 15,721 -> 15,525 tokens across "
     "07-30/07-31, the last two within an hour. The 07-28 raise to 16000 bought this "
     "seat about three days. This is the clearest measurement anywhere that a cap "
     "raise MOVES the cliff rather than closing it, on the very seat it was applied to"),
    ("council-gate",     "review_editquality",
     "same seat, mirrored roster"),
    ("feature-designer", "review_editquality",
     "third council holding it; raised to 16000 by the same owner call on 2026-07-31, "
     "so it inherits the same ceiling and the same trajectory"),
]
MEASURED_WHY = {(c, s): w for c, s, w in MEASURED}


def targets():
    """Every eligible seat, from the live DB, minus the stated exclusions."""
    rows = [l.split("|") for l in psql(ELIGIBILITY_SQL).splitlines() if "|" in l]
    out = []
    for council, seat in rows:
        if (council, seat) in EXCLUSIONS:
            continue
        out.append((council, seat,
                    MEASURED_WHY.get((council, seat), "eligible seat (fleet-wide rollout 2026-07-31)")))
    return out

# The block. One copy, here. Inserted immediately before the "## Output" heading,
# the only anchor present in all 51 live templates.
BLOCK = """LENGTH — THIS IS A CORRECTNESS CONSTRAINT, NOT A STYLE NOTE. Your whole JSON
response must fit inside your token budget. If it is cut off part-way, the council
recovers the fragment and marks your review DEGRADED, and a DEGRADED `object` gates
the round to REVISE regardless of the severities you assigned — correctly, because a
high objection may have been lost with the tail. **So an over-long ADVISORY review
silently becomes a BLOCKING one, and the round is spent on your length instead of
your judgement.** It costs you too: the object rate that results is also how a seat
gets judged noisy and retired.

The fields are ordered so that what the round needs most is emitted first and prose
last. Keep that order. Then stay inside the budget: `notes` under ~250 words, no
preamble, and do not restate the plan back to the author who wrote it. Prefer fewer,
sharper objections — but if you are running long, shorten the `problem` texts, DO NOT
drop an objection you believe in. Cut words, never findings.
— end length budget —

"""

START_PHRASE = "LENGTH — THIS IS A CORRECTNESS CONSTRAINT"
END_SENTINEL = "— end length budget —"
ANCHOR = "## Output"


def psql(sql, tuples_only=True):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--",
           "psql", "-U", "clients_user", "-d", "clients_db"]
    if tuples_only:
        cmd += ["-t", "-A"]
    r = subprocess.run(cmd + ["-c", sql], capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}")
    return r.stdout


def psql_stdin(sql):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--",
           "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]
    r = subprocess.run(cmd, input=sql, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}\n{r.stdout}")
    return r.stdout


def live_prompt(council, seat):
    """Read one seat's live prompt. NOTE THE PATH: prompt_template is a sibling of
    ai_service under config, while max_tokens is INSIDE ai_service. Both wrong-depth
    paths return NULL for every row instead of erroring, which reads as a confident
    uniform answer — that trap has now been hit twice in this lane (016b:1385, and
    again on 2026-07-30 reading these very templates)."""
    out = psql(
        "SELECT s.value->'config'->>'prompt_template' "
        "FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s "
        f"WHERE a.type='{council}' AND a.is_active AND COALESCE(a.is_snapshot,false)=false "
        f"AND a.deleted_at IS NULL AND s.key='{seat}';")
    return out if out.strip() else None


def classify(prompt):
    if prompt is None:
        return "MISSING", "no such live seat, or prompt_template absent"
    if START_PHRASE in prompt and END_SENTINEL in prompt:
        return "APPLIED", "this script's block is present"
    if START_PHRASE in prompt:
        return "HAND-WRITTEN", ("a length block exists but has no end sentinel — a "
                                "hand-authored one (review_architecture has one). "
                                "Refusing to touch it: replacing prose someone wrote "
                                "deliberately is not this script's job")
    if ANCHOR not in prompt:
        return "NO-ANCHOR", f"no '{ANCHOR}' heading to insert before"
    return "NEEDS-BLOCK", "no length budget"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true", help="snapshot, then write")
    ap.add_argument("--verify", action="store_true", help="report live state only")
    args = ap.parse_args()

    TARGETS = targets()
    print(f"── seat length budget ── {len(TARGETS)} eligible target(s)")
    for (c, st), why in EXCLUSIONS.items():
        print(f"   excluded: {c}/{st}\n             {why}")
    print()
    todo = []
    for council, seat, why in TARGETS:
        prompt = live_prompt(council, seat)
        state, detail = classify(prompt)
        if state != "APPLIED" or args.verify:
            print(f"  [{state:12}] {council}/{seat}")
            if state != "APPLIED":
                print(f"                 {detail}")
                print(f"                 why targeted: {why}")
        if state == "NEEDS-BLOCK":
            todo.append((council, seat, prompt))
    applied_already = len(TARGETS) - len(todo)
    print(f"\n  ({applied_already} of {len(TARGETS)} already carry the block or were refused above)")
    print()

    if args.verify:
        return
    if not todo:
        print("Nothing to do — every target already carries a block (or was refused above).")
        return
    if not args.apply:
        print(f"DRY RUN: would insert the block into {len(todo)} template(s):")
        for council, seat, _ in todo:
            print(f"  {council}/{seat}")
        print("\nThe block, as it would be inserted:\n")
        print("    " + "\n    ".join(BLOCK.rstrip().splitlines()))
        print("\nRe-run with --apply to snapshot each agent type and write.")
        return

    # Write. One snapshot per agent type touched, before any change to it.
    for council in sorted({c for c, _, _ in todo}):
        psql(f"SELECT snapshot_agent('{council}', "
             f"'pre-update: seat length budget, bugs_open/138 candidate 4');")
        print(f"  snapshot taken: {council}")

    for council, seat, prompt in todo:
        at = prompt.index(ANCHOR)
        new = prompt[:at] + BLOCK + prompt[at:]
        if "$PT$" in new:
            sys.exit(f"refusing: {council}/{seat} prompt contains the dollar-quote tag")
        sql = (
            "UPDATE agent_definitions SET default_config = jsonb_set(default_config, "
            f"'{{workflow,steps,{seat},config,prompt_template}}', to_jsonb($PT${new}$PT$::text), false), "
            "updated_at = now() "
            f"WHERE type='{council}' AND is_active AND COALESCE(is_snapshot,false)=false "
            "AND deleted_at IS NULL "
            f"AND default_config #>> '{{workflow,steps,{seat},config,prompt_template}}' IS NOT NULL "
            "RETURNING type;"
        )
        out = psql_stdin(sql)
        rows = [l for l in out.splitlines() if l.strip() == council]
        # create_if_missing=false above means a wrong path is a silent no-op, not an
        # error — so the row count IS the check, not decoration.
        print(f"  applied: {council}/{seat}  (rows updated: {len(rows)})")
        if len(rows) != 1:
            sys.exit(f"expected exactly 1 row for {council}/{seat}, got {len(rows)} — stopping")

    print("\nVerifying against the live rows:")
    for council, seat, _ in todo:
        state, detail = classify(live_prompt(council, seat))
        print(f"  [{state:12}] {council}/{seat} — {detail}")
    print("\nNow run 099_SYNC_gate_roster.py (dry) — it should report no drift, because"
          "\nfix-proposer and council-gate were both targeted. Drift here means a"
          "\nmirrored seat was edited on one council only.")


if __name__ == "__main__":
    main()
