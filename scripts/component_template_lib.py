#!/usr/bin/env python3
"""component_template_lib.py — shared plumbing for the content_components lints.

WHY IT EXISTS: check_list_empty_states.py and check_card_slot_guards.py both
read every ACTIVE html_template out of the live cluster and both need to know
where a {{range}} body starts and ends. Two hand-copied psql invocations and two
hand-rolled template scanners is the drift class this estate keeps paying for
(see queryresolve's shared SQL fragments, and bugs_open/425 itself, which is two
producers disagreeing about one projection). One definition, both callers.

THE RETURNCODE IS CHECKED, and that is deliberate. LANDMINES.md records that 7
of the 22 helpers under scripts/ that use capture_output=True never test
returncode — check_list_empty_states.py was one of them — so a psql that DIES
returns "" and the caller reports a clean corpus. An empty corpus and a healthy
one are indistinguishable, and the failure is in the reassuring direction.
"""
import re
import subprocess

# Openers that must be matched by a {{end}}. `else`/`else if` are NOT openers.
_OPENER = re.compile(r"\{\{-?\s*(if|range|with|block|define)\b")
_END = re.compile(r"\{\{-?\s*end\s*-?\}\}")
_ACTION = re.compile(r"\{\{.*?\}\}", re.S)

RANGE_START = re.compile(r"\{\{-?\s*range\s+(?:\$\w+\s*,\s*\$\w+\s*:=\s*)?\.([A-Za-z_]\w*)\s*-?\}\}")


class DBUnreachable(RuntimeError):
    """psql could not be run, or ran and failed. NOT the same as 'no rows'."""


def psql(query, namespace="ai-persona-system", pod="postgres-clients-0",
         user="clients_user", db="clients_db", timeout=60):
    """Run a query and return its rows as a list of field-lists.

    Raises DBUnreachable on any non-zero exit, so a dead connection can never
    present as an empty corpus.
    """
    cmd = ["kubectl", "exec", "-n", namespace, pod, "--",
           "psql", "-U", user, "-d", db, "-tAF\x1f", "-v", "ON_ERROR_STOP=1", "-c", query]
    try:
        p = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout)
    except Exception as exc:  # noqa: BLE001
        raise DBUnreachable(str(exc)) from exc
    if p.returncode != 0:
        raise DBUnreachable(
            f"psql exited {p.returncode}: {(p.stderr or '').strip()[:400]}")
    return [line.split("\x1f") for line in p.stdout.splitlines() if line.strip()]


def active_templates():
    """Every ACTIVE component that ranges over a collection: (name, template).

    Uses a record separator psql will not find inside a template, because an
    html_template contains newlines and pipes and both are ordinary content.
    """
    rows = psql(
        "SELECT name, replace(html_template, chr(10), chr(1)) "
        "FROM content_components "
        "WHERE is_active AND html_template ~ '\\{\\{-? *range' "
        "ORDER BY name;")
    return [(r[0], r[1].replace("\x01", "\n")) for r in rows if len(r) >= 2]


def range_bodies(template):
    """Yield (collection_name, body) for each {{range .X}}…{{end}} block.

    Nesting-aware: {{if}}/{{with}} inside the body do not end it. A range whose
    {{end}} is missing yields the rest of the template, which is the honest
    reading — the template would not parse, and that is a finding either way.
    """
    for m in RANGE_START.finditer(template):
        coll = m.group(1)
        depth = 1
        pos = m.end()
        body_end = len(template)
        for act in _ACTION.finditer(template, pos):
            if _END.match(act.group(0)):
                depth -= 1
                if depth == 0:
                    body_end = act.start()
                    break
            elif _OPENER.match(act.group(0)):
                depth += 1
        yield coll, template[pos:body_end]


def mentions_in_condition(fragment, field):
    """True if some {{if}} or {{with}} in `fragment` names `.field`.

    COARSE, and deliberately the same coarseness bugs_closed/054's lint admits
    to: it proves a conditional NAMING the field exists in the fragment, not
    that the conditional encloses the element. Good enough to flag a candidate;
    read the template before editing. Stated here so a caller does not quote it
    as proof of enclosure.
    """
    return re.search(
        r"\{\{-?\s*(?:if|with)\b[^}]*\.%s\b" % re.escape(field), fragment) is not None
