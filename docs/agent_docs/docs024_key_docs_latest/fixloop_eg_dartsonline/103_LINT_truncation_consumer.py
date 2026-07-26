#!/usr/bin/env python3
"""103_LINT_truncation_consumer.py — ADVISORY. Reads the LIVE agent definitions and
flags any step that opts into keeping a truncated LLM response (`tolerate_truncation:
true`) inside a workflow where NOTHING reads the `__truncated` marker.

WHY THIS EXISTS (bugs_closed/076, residual R1 — the origin is the point).

076 was: `ExecuteLLMPromptAction` may keep a response the model cut at max_tokens,
stamping `__truncated` beside the result instead of failing the step. Sound only
while a consumer READS the marker — otherwise the step succeeds and nothing
downstream can tell a complete answer from a fragment. Tolerance was already
opt-in per step, so the PRODUCER half failed closed correctly; the CONSUMER half
was enforced by nothing at all.

The shipped fix put a guard at the call site: refuse to tolerate unless the
workflow contains a truncation-aware consumer. That guard is a **floor that fires
late** — it only speaks when a response is ACTUALLY cut, in production, and it
speaks by failing that run. The bad config can sit in the fleet for weeks first.
The council seat `guardian` said exactly that when reviewing the fix (medium, corr
470678f4): validate the config itself, before any run exists. This is that check.

WHY IT IS A LIVE-DB LINT AND NOT A REGISTRATION GATE.
There is no registration or deploy step to hook. CLAUDE.md's own invariant is that
DB config is LIVE IMMEDIATELY — a seed can add `tolerate_truncation: true` to a
live workflow with no build, no deploy, no restart. So the literal thing the
council asked for ("validate at workflow registration/deploy time") has nowhere to
attach. What exists instead is two cheap checks from different sides:
  * THIS script — reads the real fleet, so it catches config however it arrived
    (seed, jsonb_set patch, or a hand-run UPDATE). Run it after any seeding.
  * pattern-check.py `truncation-tolerance-no-reader` — fires at commit time on a
    SQL seed that embeds a whole workflow and arms tolerance inside it. Free, no
    cluster, but blind to anything that reaches the DB without passing a commit.
Neither blocks. A blocking gate here would let a bad seed take the fleet down, and
that needs an owner ruling nobody has asked for.

NO SECOND COPY OF THE REGISTRY. Which actions count as readers is stated once, in
Go (`truncation_guard.go: truncationAwareActions`), and parsed out of that file by
`scripts/truncation_registry.py`. Hard-coding the two names here — the obvious way
to write this query, and the way the 076 handoff's own measurement query is
written — would create exactly the drift class this platform keeps paying for
(the council gate's 099 roster mirror; 102_LINT's own origin story). A list that
falls behind Go does not fail loudly: it reports a CLEAN fleet that is not clean.
If the parse fails, this script exits 2 and reports nothing.

WHAT A CLEAN RUN IS AND IS NOT EVIDENCE OF. On 2026-07-26 the fleet had 37
tolerating steps (council-gate 16, fix-proposer 16, feature-designer 5) and every
one of them sat in a guarded workflow, so this reports CLEAN and always did. A
clean report therefore proves nothing about the check — `--self-test` runs the
predicate against fixtures that include the offending shapes, and P4 of the plan
induced a real one in the live fleet and deleted it afterwards.

USAGE (read-only; no writes, no credits, no LLM):
    python3 103_LINT_truncation_consumer.py             # the check
    python3 103_LINT_truncation_consumer.py --verbose   # + every guarded step and its guard
    python3 103_LINT_truncation_consumer.py --self-test # predicate vs fixtures, no cluster
    python3 103_LINT_truncation_consumer.py --strict    # exit non-zero on findings

ADVISORY by default (exit 0 on findings). --strict for a caller that wants to gate.
"""
import json
import os
import subprocess
import sys

NS = "ai-persona-system"
POD = "postgres-clients-0"
DB = ["-U", "clients_user", "-d", "clients_db"]

# Only ExecuteLLMPromptAction reads `tolerate_truncation` at all, so the flag on
# any other action is inert — reported separately, and NOT as a hazard: an inert
# flag means truncation is not tolerated, i.e. it still fails closed.
PRODUCING_ACTION = "execute_llm_prompt"

YELLOW, RED, DIM, BOLD, RESET = "\033[1;33m", "\033[1;31m", "\033[2m", "\033[1m", "\033[0m"


def repo_root():
    r = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True)
    return r.stdout.strip() if r.returncode == 0 else ""


sys.path.insert(0, os.path.join(repo_root(), "scripts"))
try:
    from truncation_registry import (RegistryUnreadable, accepts_truncated_key,
                                     truncation_aware_actions, unknown_actions)
except ImportError as e:                       # noqa: BLE001 — the message IS the handling
    sys.exit(f"cannot import scripts/truncation_registry.py ({e}) — run from inside the repo")


def psql(sql):
    cmd = ["kubectl", "-n", NS, "exec", "-i", POD, "--", "psql", *DB, "-tAX",
           "-v", "ON_ERROR_STOP=1", "-c", sql]
    r = subprocess.run(cmd, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()}")
    return r.stdout.strip()


def load_rosters():
    """{label: {step_name: step_dict}} for every LIVE definition — the same
    population the chassis itself loads (active, not a snapshot, not deleted).
    One agent per output line: jsonb::text escapes every newline inside the blob.

    KEYED BY id, NOT BY type, and that is not fussiness: five live types have TWO
    active rows each (2026-07-26 — chief-strategist, content-creator,
    content-creator-contact, multipage-website-builder, site-component-architect).
    The first draft keyed a dict by `type` and silently dropped one row of each
    pair, so 171 live workflows were reported as 166 scanned. A duplicate row is
    exactly where a stale or hand-edited config hides, so dropping it is dropping
    the population most likely to offend. The label disambiguates only when it has
    to, so ordinary output stays readable.
    """
    out = psql(
        "SELECT id::text || E'\\t' || type || E'\\t' || (default_config->'workflow'->'steps')::text "
        "FROM agent_definitions "
        "WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL "
        "AND jsonb_typeof(default_config->'workflow'->'steps')='object' ORDER BY type, id;"
    )
    rows = []
    for line in out.splitlines():
        parts = line.split("\t", 2)
        if len(parts) != 3:
            continue
        agent_id, agent_type, steps_json = parts
        try:
            steps = json.loads(steps_json)
        except json.JSONDecodeError:
            continue
        if isinstance(steps, dict):
            rows.append((agent_id, agent_type, steps))
    counts = {}
    for _id, agent_type, _s in rows:
        counts[agent_type] = counts.get(agent_type, 0) + 1
    return {(t if counts[t] == 1 else f"{t}#{i[:8]}"): s for i, t, s in rows}


def cfg(step):
    c = step.get("config") if isinstance(step, dict) else None
    return c if isinstance(c, dict) else {}


def is_true(value):
    """Mirrors datahelpers.GetBoolField (data_helpers.go:1570) EXACTLY: a real JSON
    boolean true, and nothing else.

    Checked, because the first draft of this file guessed that a string "true"
    counted too and wrote a fixture asserting it. It does not: GetBoolField type-
    asserts `m[key].(bool)` and returns the default on anything else. A step
    configured `"tolerate_truncation": "true"` therefore gets NO tolerance — it
    fails closed, which is safe, but it is not what its author asked for. Reported
    as inert rather than as an offence. All 37 live flags are real booleans
    (jsonb_typeof, 2026-07-26); `jsonb_set(..., 'true'::jsonb)` — how 177 and the
    PATCH seeds write it — produces a boolean, so the string form is a hand-seed
    hazard, not a current one.
    """
    return value is True


def is_stringy_true(value):
    """A flag an author plainly meant as true that Go will read as false."""
    return isinstance(value, str) and value.strip().lower() in ("true", "t", "yes", "1")


def find_consumer(steps, producer, registry, hatch_key):
    """The Python mirror of findTruncationAwareConsumer (truncation_guard.go).

    Returns (step_name, why) of the first guarded consumer, or (None, None).
    Sorted for a stable answer, and the PRODUCER is excluded even if it carries
    the hatch flag — a step cannot certify its own truncation.
    """
    for name in sorted(steps):
        if name == producer:
            continue
        step = steps[name]
        if not isinstance(step, dict):
            continue
        action = step.get("action", "")
        if action in registry:
            return name, f"action {action}"
        if is_true(cfg(step).get(hatch_key)):
            return name, f"{hatch_key}: true (a CLAIM about that action, verified by nothing)"
    return None, None


def scan(rosters, registry, hatch_key):
    """(offenders, guarded, inert, hatch) — one pass, no cluster calls.

    inert entries carry their own reason: a flag Go never reads on that action, or
    a flag Go reads as false because it is a string. Both fail CLOSED, so neither
    is an offence — but both are a config claim that reads as protection and is
    not, which is the same misreading 076 was about, one layer up.
    """
    offenders, guarded, inert, hatch = [], [], [], []
    for agent in sorted(rosters):
        steps = rosters[agent]
        for name in sorted(steps):
            step = steps[name]
            if not isinstance(step, dict):
                continue
            action = step.get("action", "")
            config = cfg(step)
            if is_true(config.get(hatch_key)):
                hatch.append((agent, name, action))
            elif is_stringy_true(config.get(hatch_key)):
                inert.append((agent, name, action,
                              f"{hatch_key} is the STRING {config.get(hatch_key)!r}; GetBoolField "
                              f"reads only a real boolean, so this step does not declare a guard "
                              f"at all — any producer relying on it will be refused at run time"))
            flag = config.get("tolerate_truncation")
            if is_stringy_true(flag):
                inert.append((agent, name, action,
                              f"tolerate_truncation is the STRING {flag!r}; GetBoolField reads "
                              f"only a real boolean, so no tolerance is in force — the step still "
                              f"hard-fails on a cut response"))
                continue
            if not is_true(flag):
                continue
            if action != PRODUCING_ACTION:
                inert.append((agent, name, action,
                              f"tolerate_truncation is only read by {PRODUCING_ACTION}; on this "
                              f"action it does nothing"))
                continue
            consumer, why = find_consumer(steps, name, registry, hatch_key)
            (guarded if consumer else offenders).append((agent, name, consumer, why))
    return offenders, guarded, inert, hatch


# ── fixtures ────────────────────────────────────────────────────────────────
# The live fleet is clean and has been since the guard shipped, so this check
# passing against it is not evidence the check works. These are the shapes that
# must separate. Each fixture is (label, steps, expect_offenders, expect_inert).
FIXTURES = [
    ("offender: tolerance, nothing reads the marker",
     {"ask": {"action": "execute_llm_prompt", "config": {"tolerate_truncation": True}},
      "done": {"action": "complete_workflow", "config": {}}},
     ["ask"], []),
    ("guarded by the Go registry",
     {"ask": {"action": "execute_llm_prompt", "config": {"tolerate_truncation": True}},
      "decide": {"action": "diagnose_council_decide", "config": {}}},
     [], []),
    ("guarded by the config hatch on another step",
     {"ask": {"action": "execute_llm_prompt", "config": {"tolerate_truncation": True}},
      "sink": {"action": "complete_workflow", "config": {"accepts_truncated": True}}},
     [], []),
    ("self-certifying: the producer's OWN hatch flag must not count",
     {"ask": {"action": "execute_llm_prompt",
              "config": {"tolerate_truncation": True, "accepts_truncated": True}},
      "done": {"action": "complete_workflow", "config": {}}},
     ["ask"], []),
    ("a LATER guarded step counts, not just an earlier one (order is not the test)",
     {"ask": {"action": "execute_llm_prompt", "config": {"tolerate_truncation": True}},
      "zz_verify": {"action": "verify_report_prose", "config": {}}},
     [], []),
    ("string 'true': Go reads only a real bool, so this is INERT, not tolerance",
     {"ask": {"action": "execute_llm_prompt", "config": {"tolerate_truncation": "true"}},
      "done": {"action": "complete_workflow", "config": {}}},
     [], ["ask"]),
    ("a string hatch flag guards nothing — the producer is an offender",
     {"ask": {"action": "execute_llm_prompt", "config": {"tolerate_truncation": True}},
      "sink": {"action": "complete_workflow", "config": {"accepts_truncated": "true"}}},
     ["ask"], ["sink"]),
    ("no tolerance anywhere — the ordinary case, must stay silent",
     {"ask": {"action": "execute_llm_prompt", "config": {}},
      "done": {"action": "complete_workflow", "config": {}}},
     [], []),
    ("tolerance on a non-LLM step is INERT, not an offence",
     {"fetch": {"action": "query_database", "config": {"tolerate_truncation": True}},
      "done": {"action": "complete_workflow", "config": {}}},
     [], ["fetch"]),
]


def self_test(registry, hatch_key):
    failures = 0
    for label, steps, want_off, want_inert in FIXTURES:
        offenders, _guarded, inert, _hatch = scan({"fixture": steps}, registry, hatch_key)
        got_off = sorted(step for _a, step, _c, _w in offenders)
        got_inert = sorted(step for _a, step, _act, _why in inert)
        ok = got_off == sorted(want_off) and got_inert == sorted(want_inert)
        failures += 0 if ok else 1
        print(f"  {'PASS' if ok else BOLD + 'FAIL' + RESET}  {label}"
              + ("" if ok else f"\n        offenders: expected {sorted(want_off)}, got {got_off}"
                               f"\n        inert:     expected {sorted(want_inert)}, got {got_inert}"))
    print(f"\n{len(FIXTURES) - failures}/{len(FIXTURES)} fixtures pass")
    return 1 if failures else 0


def main():
    strict = "--strict" in sys.argv
    verbose = "--verbose" in sys.argv
    try:
        registry = truncation_aware_actions()
        hatch_key = accepts_truncated_key()
        unknown = unknown_actions()
    except RegistryUnreadable as e:
        print(f"{RED}REGISTRY UNREADABLE{RESET}: {e}", file=sys.stderr)
        print("Reporting nothing rather than guessing — a stale list reports a clean fleet "
              "that is not clean.", file=sys.stderr)
        return 2

    print(f"{DIM}readers (parsed from platform/orchestration/actions/truncation_guard.go): "
          f"{', '.join(sorted(registry))} | config hatch: {hatch_key}{RESET}")
    if unknown:
        print(f"{RED}registry names an action registry.go does not have: {', '.join(unknown)}{RESET} "
              f"— that entry guards nothing.")

    if "--self-test" in sys.argv:
        return self_test(registry, hatch_key)

    rosters = load_rosters()
    if not rosters:
        sys.exit("no live agent definitions found — wrong cluster/namespace?")
    offenders, guarded, inert, hatch = scan(rosters, registry, hatch_key)
    print(f"{DIM}scanned {len(rosters)} live agent definitions | tolerating steps: "
          f"{len(offenders) + len(guarded)} ({len(guarded)} guarded){RESET}")

    if verbose:
        for agent, step, consumer, why in guarded:
            print(f"   {DIM}ok{RESET}  {agent}.{step} {DIM}-> guarded by {consumer} ({why}){RESET}")

    if hatch:
        print(f"\n{DIM}── {hatch_key} declarations ({len(hatch)}) — each is an UNVERIFIED claim "
              f"that the action handles a partial (bugs_closed/076 R2) ──{RESET}")
        for agent, step, action in hatch:
            print(f"   {agent}.{step}  action={action}")

    for agent, step, action, why in inert:
        print(f"\n{DIM}inert-flag{RESET}  {agent}.{step}  action={action}")
        print(f"     {why}.")
        print(f"     {DIM}Not a hazard — it fails closed. But it is a config claim that reads as "
              f"protection and is not.{RESET}")

    if not offenders:
        print(f"{BOLD}clean{RESET} — every step that tolerates a truncated response sits in a "
              f"workflow that reads the marker.")
        print(f"{DIM}Clean is the expected state and is NOT evidence this check works: run "
              f"--self-test for that.{RESET}")
        return 0

    print(f"\n{YELLOW}── truncation consumer: {len(offenders)} step(s) tolerate a cut response "
          f"with NO reader ──{RESET}")
    for agent, step, _c, _w in offenders:
        print(f"   {BOLD}truncation-tolerance-no-reader{RESET}  {agent}.{step}")
        print(f"     sets tolerate_truncation but no other step in this workflow reads "
              f"__truncated, so a fragment would be indistinguishable from a complete answer.")
        print(f"     {DIM}Fix one of: raise max_tokens; drop tolerate_truncation; or set "
              f"{hatch_key}: true on the step that genuinely handles a partial. The runtime "
              f"guard will FAIL this step's run when a response is actually cut "
              f"(bugs_closed/076).{RESET}")
    print(f"   {DIM}Advisory — this never blocks. --strict exits non-zero.{RESET}")
    return 1 if strict else 0


if __name__ == "__main__":
    sys.exit(main())
