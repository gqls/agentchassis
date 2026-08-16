#!/usr/bin/env python3
"""
instance-token-adoption-check — the expiry tripwire for bugs_open/283's RFC_022 exception.

WHY THIS EXISTS.

`bugs_open/283` gave components a per-instance element-id namespace, `{{.InstanceID}}`
(register CLC-014). The council's `architecture` seat approved it under RFC_022's NARROW
EXCEPTION, whose three conditions are: the field is opt-in, its unsafe side is the default,
and **no live consumer names it**. Its approval note is explicit about the third:

    "The moment the 22 templates start consuming InstanceID, condition 3 of the exception
     (zero live consumers) stops holding and this becomes a real load-bearing contract
     across the component library. That conversion PR, not this one, is where an RFC or at
     minimum a fresh architecture pass belongs."

So the exception is not permanent — it expires on an event. The seat asked for a WRITTEN
TRIGGER rather than a prose reminder, and this is it.

**A commit-time lint cannot do this job**, which is the whole reason it is a CronJob — the
same argument its two siblings make, and it is even more clear-cut here. A component's
`html_template` lives in a DATABASE COLUMN. It is written by the component-creator agent,
by hand-authored SQL, by migrations, and by the admin UI. Not one of those routes passes
through a commit, so `scripts/pattern-check.py` — which sees repo files — is structurally
incapable of noticing the first template that adopts the token. Only a clock against live
`content_components` can.

WHAT IT DOES. Counts active components whose `html_template` references `{{.InstanceID}}`.

  0  -> the exception still holds. 283's seam is live and inert. Nothing to do.
  >0 -> THE EXCEPTION HAS EXPIRED. The token is now a load-bearing contract across the
        component library, and `architecture_review/RFC_032` (or a fresh architecture
        round) is owed BEFORE more templates convert.

It writes ONE `doc_notes` row per run — on a trip AND on a quiet result — so that a MISSING
row means THE JOB DID NOT RUN, which is a different thing from "the exception still holds"
and must not look like it. Exits non-zero on a trip so the Job shows as failed.

⚠ THE POLARITY IS THE OPPOSITE OF ITS SIBLINGS, AND THAT IS DELIBERATE. A non-zero count
here is NOT a defect — adopting the token is the intended and desirable next phase of 283.
This job does not object to the conversion; it objects to the conversion happening without
anybody noticing that the architecture exception licensing the seam has lapsed. Read a trip
as "an owed review is now due", never as "someone broke something".

THE DEMAND CONTROL, and why this check would otherwise be worthless. Its normal, expected,
correct answer is ZERO — and a query that is broken, mis-escaped, or pointed at an empty
table also returns zero. A check whose healthy output is indistinguishable from its failed
output is not a check. So every run also counts templates referencing `{{.ComponentID}}`,
which is known to be non-zero (5 live components at the time of writing: faq,
generic-text-block, mechanism-flow, evidence-timeseries, pricing). If THAT count is zero,
the LIKE-matching itself is not working and the run REFUSES rather than reporting a
reassuring zero it has not earned.
"""

import json
import os
import subprocess
import sys

# One statement, three numbers, so a single round trip answers the question and its control.
CENSUS_SQL = """
SELECT json_build_object(
  'adopters',       (SELECT count(*) FROM content_components
                       WHERE is_active AND html_template LIKE '%{{.InstanceID}}%'),
  'control',        (SELECT count(*) FROM content_components
                       WHERE is_active AND html_template LIKE '%{{.ComponentID}}%'),
  'active_total',   (SELECT count(*) FROM content_components WHERE is_active),
  'adopter_names',  (SELECT COALESCE(json_agg(x.function ORDER BY x.function), '[]'::json)
                       FROM (SELECT function FROM content_components
                              WHERE is_active AND html_template LIKE '%{{.InstanceID}}%'
                              LIMIT 50) x)
);
"""


def psql(sql, password, host):
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-t", "-A", "-v", "ON_ERROR_STOP=1", "-c", sql],
        env=env, check=True, capture_output=True, text=True,
    )
    return out.stdout.strip()


def render_report(c):
    adopters = c["adopters"]
    head = [
        "instance-token-adoption-check — the RFC_022 expiry tripwire for bugs_open/283.",
        "",
        f"active components:                 {c['active_total']}",
        f"referencing {{{{.InstanceID}}}}:          {adopters}",
        f"referencing {{{{.ComponentID}}}} (control): {c['control']}",
        "",
    ]
    if adopters == 0:
        return "\n".join(head + [
            "QUIET — the exception still holds. 283's per-instance seam is live and inert:",
            "no live template consumes the token, so condition 3 of RFC_022's narrow",
            "exception (zero live consumers) is intact and no architecture round is owed.",
            "",
            "This row exists on a quiet run ON PURPOSE: a MISSING row means the job did not",
            "run, which is not the same as 'the exception still holds', and the two must not",
            "look alike.",
        ])

    return "\n".join(head + [
        "*** TRIPPED — THE RFC_022 EXCEPTION HAS EXPIRED. ***",
        "",
        "A live component template now references {{.InstanceID}}, so the per-instance",
        "identity is a load-bearing contract across the component library rather than an",
        "opt-in field nothing names. The council's architecture seat approved bugs_open/283",
        "on the explicit condition that this moment gets its own review:",
        "",
        '  "That conversion PR, not this one, is where an RFC or at minimum a fresh',
        '   architecture pass belongs."',
        "",
        "Adopting components:",
    ] + [f"      {n}" for n in c["adopter_names"]] + [
        "",
        "THIS IS NOT A DEFECT REPORT. Converting templates is the intended next phase of",
        "283 and the thing that actually fixes the bug. What is owed is the review, not a",
        "revert. Do NOT 'fix' this by reverting a conversion.",
        "",
        "What to do: take architecture_review/RFC_032 (three render context-builders",
        "disagree about what an instance is) to a round, or open a fresh architecture pass.",
        "Then retire this CronJob — a tripwire that has tripped is noise, and leaving it",
        "firing daily is how a real signal becomes one people mute.",
        "",
        "Background: bugs_open/283 §9-§10, register CLC-014, RFC_022's narrowing.",
    ])


def write_doc_note(body, password, host):
    tag = "itabody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'instance-token-adoption', ${tag}${body}${tag}$, "
        "'[\"instance-token-adoption\",\"rfc-022\",\"bugs-open-283\"]'::jsonb, "
        "'instance-token-adoption-check');"
    )
    path = "/tmp/instance-token-note.sql"
    with open(path, "w") as f:
        f.write(sql)
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-v", "ON_ERROR_STOP=1", "-f", path],
        env=env, check=True,
    )


def main():
    # --stdin renders a report from a census on stdin, so the two branches can be
    # exercised without a database (and so a trip can be SEEN before it happens).
    if "--stdin" in sys.argv:
        c = json.load(sys.stdin)
        print(render_report(c))
        sys.exit(1 if c["adopters"] else 0)

    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    raw = psql(CENSUS_SQL, password, host)
    if not raw:
        print("REFUSING TO RUN: the census returned nothing.", file=sys.stderr)
        sys.exit(2)
    c = json.loads(raw)

    if c["active_total"] == 0:
        print("REFUSING TO RUN: 0 active components — the query failed or the library is "
              "empty; refusing to report a quiet result over it.", file=sys.stderr)
        sys.exit(2)

    # THE DEMAND CONTROL. This check's healthy answer is zero, so it must prove it could
    # have seen a non-zero. {{.ComponentID}} is matched by the same LIKE, through the same
    # escaping, in the same statement — if it comes back empty, the matching is broken and
    # this run's zero says nothing about {{.InstanceID}}.
    if c["control"] == 0:
        print("REFUSING TO RUN: the {{.ComponentID}} control matched 0 templates. This "
              "check's expected answer is 0, so without a control that fires, a quiet "
              "result is indistinguishable from a broken query. Investigate the pattern "
              "matching before trusting any zero from this job.", file=sys.stderr)
        sys.exit(2)

    report = render_report(c)
    print(report)
    write_doc_note(report, password, host)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='instance-token-adoption').")
    sys.exit(1 if c["adopters"] else 0)


if __name__ == "__main__":
    main()
