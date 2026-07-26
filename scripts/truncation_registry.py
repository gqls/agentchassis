#!/usr/bin/env python3
"""truncation_registry.py — reads the truncation-consumer registry OUT OF THE GO
SOURCE, so that no checker anywhere holds a second copy of it.

WHY THIS EXISTS (bugs_closed/076, residual R1).

`ExecuteLLMPromptAction` may keep an LLM response the model cut at max_tokens,
stamping `__truncated` beside the result instead of failing the step. That is only
sound while something downstream READS the marker, so the shipped guard refuses to
tolerate unless the workflow contains a truncation-aware consumer. Which actions
count is stated once, in Go:

    platform/orchestration/actions/truncation_guard.go
        var truncationAwareActions = map[string]string{...}
        const acceptsTruncatedConfigKey = "accepts_truncated"

Every static check over `agent_definitions` has to ask the same question, and the
obvious way to write one is to paste those action names into the query. The 076
handoff flagged that as the landmine before anything was built, and it is right:
two hand-maintained lists that must agree is the drift class this platform keeps
paying for — the council gate's 099 roster mirror exists for it, and
102_LINT_council_seat_parity.py exists because of another instance of it. A copy
that falls behind the Go registry does not fail loudly; it reports a CLEAN fleet
that is not clean, which is worse than no check at all.

So there is no copy. This module parses the source.

FAILS LOUD, NEVER FALLS BACK. If the registry is renamed, restructured, or
rewritten in a form these patterns cannot read, every function here raises. A
remembered default would be indistinguishable from a correct parse at the point
where it matters, so there is deliberately no default to fall back to.

USAGE:
    from truncation_registry import truncation_aware_actions, accepts_truncated_key
    python3 scripts/truncation_registry.py     # print what it parses, and validate
"""
import os
import re
import subprocess
import sys

GUARD_FILE = "platform/orchestration/actions/truncation_guard.go"
REGISTRY_FILE = "platform/orchestration/actions/registry.go"

# var truncationAwareActions = map[string]string{ ... }
# Closing brace at column 0 ends the literal — gofmt guarantees that for a
# top-level declaration, and gofmt is itself enforced by pattern-check.py.
_MAP_BLOCK = re.compile(
    r"^var\s+truncationAwareActions\s*=\s*map\[string\]string\{(.*?)^\}",
    re.M | re.S,
)
# Entries are found by their KEY, and the value is everything up to the next key.
# The first version of this matched key AND value on one line, and silently dropped
# any entry whose mechanism string gofmt had wrapped across two lines with a `+`
# — three entries in the file, two parsed, exit 0. That direction of failure is
# the nastier one for the callers: a MISSING reader makes every workflow guarded
# only by that action look like an offender, so the lint cries wolf on a clean
# fleet. Caught by a second session reading this file on 2026-07-26; reproduced
# against a copied tree before being fixed, and the count backstop below now
# raises if any entry is present but unparsed.
_MAP_KEY = re.compile(r'^\s*"([a-z0-9_]+)"\s*:', re.M)
_GO_STRING = re.compile(r'"((?:[^"\\]|\\.)*)"')
_ACCEPTS_KEY = re.compile(r'^const\s+acceptsTruncatedConfigKey\s*=\s*"([a-z0-9_]+)"', re.M)
# registry.go entries: `"action_name": {` one indent in.
_REGISTERED = re.compile(r'^\t"([a-z0-9_]+)":\s*\{', re.M)


class RegistryUnreadable(RuntimeError):
    """The Go registry could not be parsed. Fix the parser — do not guess."""


def repo_root():
    if os.environ.get("REPO_ROOT"):
        return os.environ["REPO_ROOT"]
    r = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True)
    if r.returncode != 0:
        raise RegistryUnreadable("not inside the repo (git rev-parse failed)")
    return r.stdout.strip()


def _read(rel):
    path = os.path.join(repo_root(), rel)
    try:
        with open(path, encoding="utf-8") as fh:
            return fh.read()
    except OSError as e:
        raise RegistryUnreadable(f"cannot read {rel}: {e}")


def truncation_aware_actions():
    """{action_name: stated mechanism} — the Go registry, parsed.

    The value is the mechanism the Go entry claims, so a caller can print WHY an
    action counts as a reader without anyone re-deriving it.
    """
    src = _read(GUARD_FILE)
    m = _MAP_BLOCK.search(src)
    if not m:
        raise RegistryUnreadable(
            f"truncationAwareActions map literal not found in {GUARD_FILE} — it was "
            "renamed or restructured. Fix this parser; a stale hard-coded list would "
            "report a clean fleet that is not clean (bugs_closed/076 R1)."
        )
    body = m.group(1)
    keys = list(_MAP_KEY.finditer(body))
    actions = {}
    for i, km in enumerate(keys):
        end = keys[i + 1].start() if i + 1 < len(keys) else len(body)
        value = body[km.end():end]
        # A wrapped value is `"part one " +\n\t\t"part two"` — join the literals.
        # Go string escapes survive the regex; unescape so the printed mechanism
        # reads as the source does (entries quote symbol names and error strings).
        parts = [p.replace('\\"', '"').replace("\\\\", "\\") for p in _GO_STRING.findall(value)]
        if not parts:
            raise RegistryUnreadable(
                f'registry entry "{km.group(1)}" in {GUARD_FILE} has no readable mechanism '
                "string — the entry format changed. Fix this parser rather than letting an "
                "entry be dropped silently."
            )
        actions[km.group(1)] = "".join(parts)
    if len(actions) != len(keys):
        raise RegistryUnreadable(
            f"{len(keys)} entries present in truncationAwareActions but {len(actions)} parsed "
            f"({GUARD_FILE}) — duplicate or unreadable key. Refusing to report on a partial "
            "registry: a MISSING reader makes correctly guarded workflows read as offenders."
        )
    if not actions:
        raise RegistryUnreadable(
            f"truncationAwareActions parsed EMPTY from {GUARD_FILE}. An empty registry "
            "would mark every tolerating step in the fleet as an offender; refusing to "
            "report rather than guess."
        )
    return actions


def accepts_truncated_key():
    """The per-step config key that declares 'my action handles a partial'."""
    src = _read(GUARD_FILE)
    m = _ACCEPTS_KEY.search(src)
    if not m:
        raise RegistryUnreadable(
            f"acceptsTruncatedConfigKey not found in {GUARD_FILE} — fix this parser."
        )
    return m.group(1)


def registered_action_names():
    """Every action name registry.go knows. Used only to cross-check the parse:
    a truncationAwareActions entry naming an action that does not exist is a
    typo that silently guards nothing, and it is the one error the Go-side
    lockstep test cannot catch (it scans for a READER, not for registration)."""
    return set(_REGISTERED.findall(_read(REGISTRY_FILE)))


def unknown_actions():
    """Parsed registry entries that name no registered action. Empty on a healthy
    tree; non-empty is a real defect, not a parser problem."""
    known = registered_action_names()
    if not known:
        raise RegistryUnreadable(f"no action entries parsed from {REGISTRY_FILE}")
    return sorted(a for a in truncation_aware_actions() if a not in known)


def main():
    try:
        actions = truncation_aware_actions()
        key = accepts_truncated_key()
        unknown = unknown_actions()
    except RegistryUnreadable as e:
        print(f"REGISTRY UNREADABLE: {e}", file=sys.stderr)
        return 2
    print(f"source: {GUARD_FILE}")
    print(f"config hatch key: {key}")
    print(f"truncation-aware actions ({len(actions)}):")
    for name, why in sorted(actions.items()):
        print(f"  {name}\n      {why}")
    if unknown:
        print(f"\nWARNING: registered as truncation-aware but unknown to "
              f"{REGISTRY_FILE}: {', '.join(unknown)}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
