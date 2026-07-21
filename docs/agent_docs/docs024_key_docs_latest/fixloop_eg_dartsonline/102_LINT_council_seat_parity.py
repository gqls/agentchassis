#!/usr/bin/env python3
"""102_LINT_council_seat_parity.py — ADVISORY. Reads the LIVE council rosters and
flags a seat that has drifted from its siblings: a config KEY the rest of the
family carries but it lacks, a critical VALUE (model / max_tokens / temperature)
it holds alone, or a seat wired into the workflow but absent from the decider's
roster (or vice versa).

WHY THIS EXISTS (2026-07-21, and the origin is the point).
`pattern-check.py` closed the GO half of one recurring class — "a fix whose scope
is a snapshot of a growing set" — but it reads git diffs of FILES, and TWO of the
five instances of that class in the two days before it was written lived in
`agent_definitions` JSON, which is in the DATABASE and never passes through a
commit:
  - bugs_open/019's last hole: `review_prior_art` (the 16th, always-on seat)
    shipped without `tolerate_truncation` while the other 15 seats had it. A
    truncated forensic row on THAT seat re-opened the void the whole 019 fix
    existed to close. Nobody noticed for a roster growth because nothing compared
    the new seat against its siblings.
  - 016 f2's reviser kept a hand-written seat list; the roster grew past it.
The shape is always the same: a family of seats that must stay parallel, one new
member added, one key/value forgotten on it. A string comparison across the
family answers in milliseconds what otherwise costs a voided council round and
real credits to discover.

WHY IT IS A LIVE-DB LINT, NOT A PRE-COMMIT CHECK. Council config lives in
`agent_definitions.default_config` and is LIVE the instant it is seeded — it does
not pass through git, so a pre-commit hook structurally cannot see it. The source
of truth is the live row. Run this AFTER any council/seat growth (the same moment
you would run 099), and it reads the rows 099 just wrote.

WHAT IT IS NOT, AND WHY THAT MATTERS.
- It is NOT 099. 099 mirrors fix-proposer -> council-gate, comparing the two
  councils against EACH OTHER; it is therefore BLIND to a key missing on BOTH
  (which is exactly how the 019 hole survived — prior_art lacked tolerate_truncation
  on both councils, so 099 read them as "in sync"). This lint compares each seat
  against its OWN family, one council at a time, and catches that.
- It does NOT compare councils to each other beyond 099's remit. Councils
  legitimately differ: experience-planner uniformly OMITS tolerate_truncation
  (owner ruling — the narrow-scope decision), and it has a different seat set.
  A cross-council parity rule would fire on that deliberate divergence, and a
  check that fires on intended state trains people to ignore it (pattern-check.py's
  header documents where that road ends).

MEASURED, like every check in pattern-check.py. Over the whole live fleet on
2026-07-21: every review_/gate_ family is internally uniform in both keys and the
critical values (distinct_keysets = 1; distinct model/max_tokens/temperature = 1
per family), so the default scope reports ZERO on a healthy fleet and fires only
on a genuine deviation. The general `--all-families` pass (every prefix family,
not just councils) was measured to surface ~11 findings fleet-wide, MOSTLY the
recognisable false-positive classes a council roster does not have — terminal
`complete_*`/`finalize_*` steps that legitimately differ (only the success
terminal carries `success_message`), and optional `query_database` params. That
pass is OFF by default and labelled advisory; it is an audit tool, not the check.

WHY PREFIX FAMILIES, NOT ACTION FAMILIES. Grouping by step `action` puts
fix-proposer's 16 reviewer seats in the same bucket as `propose`/`reframe`/
`repropose` (also execute_llm_prompt, but NOT reviewers and legitimately without
tolerate_truncation). tolerate_truncation would then read as 16/19 — missing from
three — and the odd-one-out rule would MISS the exact 019 case it is for. The
roster convention is the `review_`/`gate_` NAME PREFIX (099 relies on it too), and
that is the family boundary here.

USAGE (read-only; no writes, no credits):
    python3 102_LINT_council_seat_parity.py            # council rosters (default)
    python3 102_LINT_council_seat_parity.py --all-families   # + general audit pass
    python3 102_LINT_council_seat_parity.py --strict   # exit non-zero on findings

ADVISORY by default (exit 0). --strict makes findings exit non-zero, for a caller
that wants to gate on a clean roster.
"""
import json
import subprocess
import sys

NS = "ai-persona-system"
POD = "postgres-clients-0"
DB = ["-U", "clients_user", "-d", "clients_db"]

# The family boundary: seats whose name begins with one of these tokens are an
# intentional, parallel roster. Measured zero-noise on the live fleet.
ROSTER_PREFIXES = ("review", "gate")
MIN_FAMILY = 3  # below this a "family" is not a roster; do not compare

# Critical config paths that MUST be uniform across a reviewer family — a seat on
# a stale model or a smaller ceiling is the D1 drift class (099's own comment
# records the gate silently left on sonnet-4-6 @ 3000 while fix-proposer moved to
# sonnet-5 @ 8000). Compared only among seats that HAVE the path, so a family that
# uniformly omits one (experience-planner + tolerate_truncation) never fires.
CRITICAL_PATHS = [
    ("model", ("ai_service", "model")),
    ("max_tokens", ("ai_service", "max_tokens")),
    ("temperature", ("temperature",)),
]

YELLOW, RED, DIM, BOLD, RESET = "\033[1;33m", "\033[1;31m", "\033[2m", "\033[1m", "\033[0m"


def psql(sql):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--", "psql", *DB, "-tAX", "-v", "ON_ERROR_STOP=1", "-c", sql]
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}")
    return r.stdout.strip()


def load_rosters():
    """One row per live agent: {agent: {step_name: step_config_dict}}. The steps
    blob is emitted as jsonb::text (all newlines/tabs inside string values are
    escaped, so one agent is exactly one output line) prefixed by the agent type
    and a literal tab we split on."""
    out = psql(
        "SELECT type || E'\\t' || (default_config->'workflow'->'steps')::text "
        "FROM agent_definitions "
        "WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL "
        "AND jsonb_typeof(default_config->'workflow'->'steps')='object' ORDER BY type;"
    )
    rosters = {}
    for line in out.splitlines():
        if "\t" not in line:
            continue
        agent, steps_json = line.split("\t", 1)
        try:
            rosters[agent] = json.loads(steps_json)
        except json.JSONDecodeError:
            pass
    return rosters


def prefix(step):
    return step.split("_", 1)[0] if "_" in step else step


def dig(cfg, path):
    cur = cfg
    for p in path:
        if not isinstance(cur, dict) or p not in cur:
            return None
        cur = cur[p]
    return cur


def families(steps, prefixes):
    """{(prefix, action): [(step_name, config_dict), ...]} for families >= MIN_FAMILY.
    Keyed by prefix AND action: a coherent roster is same-prefix AND same-action —
    comparing config keys across different actions (a query_database vs a bespoke
    loader that merely share a load_ prefix) is meaningless and pure noise."""
    fams = {}
    for name, step in steps.items():
        if not isinstance(step, dict):
            continue
        p = prefix(name)
        if prefixes is not None and p not in prefixes:
            continue
        action = step.get("action", "")
        fams.setdefault((p, action), []).append((name, step.get("config", {}) or {}))
    return {k: m for k, m in fams.items() if len(m) >= MIN_FAMILY}


def check_key_parity(agent, fams, findings):
    for (pfx, _action), members in fams.items():
        n = len(members)
        counts = {}
        for _name, cfg in members:
            for k in cfg:
                counts[k] = counts.get(k, 0) + 1
        for k, c in counts.items():
            if c == n:
                continue  # universal — the healthy case
            missing = sorted(name for name, cfg in members if k not in cfg)
            present = c
            # Only the near-universal keys are a parity signal. A key held by a
            # minority is a sub-family distinction, not an omission.
            if present >= n - max(1, n // 8):
                findings.append((
                    "seat-missing-key", agent, f"{pfx}_* ({n} seats)",
                    f"config key {BOLD}{k}{RESET} is on {present}/{n} seats but MISSING from: {', '.join(missing)}",
                ))


def check_value_drift(agent, fams, findings):
    for (pfx, _action), members in fams.items():
        for label, path in CRITICAL_PATHS:
            vals = {}
            for name, cfg in members:
                v = dig(cfg, path)
                if v is not None:
                    vals.setdefault(json.dumps(v, sort_keys=True), []).append(name)
            if len(vals) <= 1:
                continue  # uniform (or nobody sets it) — healthy
            # Report the minority value(s) against the plurality.
            majority = max(vals.values(), key=len)
            maj_val = [k for k, v in vals.items() if v is majority][0]
            for v, seats in vals.items():
                if seats is majority:
                    continue
                findings.append((
                    "seat-value-drift", agent, f"{pfx}_* / {label}",
                    f"{', '.join(sorted(seats))} = {BOLD}{v}{RESET} while {len(majority)} sibling(s) = {maj_val}",
                ))


def check_decider_roster(agent, steps, findings):
    """Every review_ seat should be in council_decide.review_fields, and every
    review_field should have a seat. A seat added but not wired into the decider
    is silently ignored by the vote — the same growing-set drift, one table over."""
    decide = None
    for name, step in steps.items():
        if isinstance(step, dict) and step.get("action") == "diagnose_council_decide":
            decide = step
            break
    if not decide:
        return
    review_fields = (decide.get("config", {}) or {}).get("review_fields", [])
    if not isinstance(review_fields, list):
        return
    wired = {rf.rsplit(".", 1)[0] for rf in review_fields if isinstance(rf, str)}
    seats = {name for name, step in steps.items()
             if prefix(name) == "review" and isinstance(step, dict)
             and step.get("action") == "execute_llm_prompt"}
    seat_not_wired = sorted(seats - wired)
    wired_no_seat = sorted(wired - seats)
    for s in seat_not_wired:
        findings.append((
            "seat-not-in-decider", agent, "council_decide.review_fields",
            f"seat {BOLD}{s}{RESET} runs but is NOT in review_fields — the decider ignores its vote",
        ))
    for w in wired_no_seat:
        findings.append((
            "decider-ref-no-seat", agent, "council_decide.review_fields",
            f"review_fields names {BOLD}{w}{RESET} but no such seat step exists — dangling reference",
        ))


def main():
    all_families = "--all-families" in sys.argv
    strict = "--strict" in sys.argv
    rosters = load_rosters()
    if not rosters:
        sys.exit("no live agent rosters found")

    findings = []
    scope_prefixes = None if all_families else set(ROSTER_PREFIXES)
    council_agents = []
    for agent, steps in rosters.items():
        fams = families(steps, scope_prefixes)
        if fams:
            council_agents.append(agent)
        check_key_parity(agent, fams, findings)
        check_value_drift(agent, fams, findings)
        check_decider_roster(agent, steps, findings)

    scope = "ALL prefix families" if all_families else f"council rosters ({'/'.join(ROSTER_PREFIXES)}_*)"
    print(f"{DIM}scope: {scope} | agents with families: {len(council_agents)}{RESET}")
    if not findings:
        print(f"{BOLD}clean{RESET} — every family is internally uniform, every seat is wired into its decider.")
        return 0

    print(f"\n{YELLOW}── council seat parity: {len(findings)} thing(s) worth a look ──{RESET}")
    for kind, agent, where, what in findings:
        print(f"   {BOLD}{kind}{RESET}  {agent}  {DIM}{where}{RESET}")
        print(f"     {what}")
    print(f"   {DIM}Advisory. A near-universal key/value missing from one seat is the recurring "
          f"'snapshot of a growing set' class (bugs_open/019, 016 f2). Deliberate? Say so where the roster is defined.{RESET}")
    print(f"   {DIM}After fixing on fix-proposer, mirror with 099_SYNC_gate_roster.py --apply.{RESET}")

    return 1 if strict else 0


if __name__ == "__main__":
    sys.exit(main())
