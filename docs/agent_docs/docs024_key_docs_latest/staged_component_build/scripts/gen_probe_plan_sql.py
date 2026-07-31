#!/usr/bin/env python3
"""Generate the SQL that writes (or rolls back) the throwaway component PLAN
used by PROBE_doc_subject_go_gate.sh.

Why a generator and not a heredoc: the PLAN body contains triple backticks (the
```criteria fence). In a double-quoted bash string those are COMMAND
SUBSTITUTION, and the mangling is silent — you would discover it as a corrupted
contract in production, not as an error here. RUNBOOK §9.

Usage:
  gen_probe_plan_sql.py            -> dry run (BEGIN ... ROLLBACK)
  gen_probe_plan_sql.py --apply    -> BEGIN ... COMMIT
"""
import sys

SUBJECT_TYPE = "component"
SUBJECT_KEY = "teaser-reveal-panel"
SOURCE = "handoff-goproof"

BODY = """# PLAN — component `teaser-reveal-panel` (PROBE ROW — DELETE ME)

**This is not an authored contract.** It exists only to give `load_doc_context`
a row to return, so the Go-side `subject_type` gate can be *observed* accepting
`component` in the running chassis binary rather than inferred from a build
date. Written and deleted by the same session (staged_component_build,
2026-07-31), per HANDOFF_2026-07-31b §3.

If you are reading this in the database, the cleanup did not run:

    DELETE FROM doc_plans WHERE source = 'handoff-goproof';

The real contract for this component, when someone authors one, must be proved
with `try_fence.go` and `prove_fence_can_fail.go` first (register TL-036) — an
unproven fence is the defect this lane exists to stop.

## Acceptance criteria

```criteria
{"probe": true, "checks": []}
```
"""

apply = "--apply" in sys.argv
body_len = len(BODY)

print(f"-- PROBE PLAN row for {SUBJECT_TYPE}/{SUBJECT_KEY} — "
      f"{'APPLY' if apply else 'DRY RUN (rolls back)'}")
print("BEGIN;")
# idx_doc_plans_current is a partial UNIQUE index on (subject_type, subject_key)
# WHERE is_current, so supersede before inserting even though the current count
# is 0 — another session could have written one between the count and this run.
print(f"""UPDATE doc_plans SET is_current=false, superseded_at=now(), updated_at=now()
 WHERE subject_type='{SUBJECT_TYPE}' AND subject_key='{SUBJECT_KEY}' AND is_current;""")
print(f"""INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('{SUBJECT_TYPE}','{SUBJECT_KEY}', $planbody${BODY}$planbody$,
        '{SOURCE}', 'operator:staged_component_build');""")
# Assert the stored length equals the length built here. That single check is
# also how you prove psql did not interpolate a :name inside the literal.
print(f"""SELECT subject_type, subject_key, source, is_current,
       length(body) AS stored_chars,
       length(body) = {body_len} AS length_matches_generator,
       position('```criteria' in body) > 0 AS has_criteria_fence,
       (SELECT count(*) FROM regexp_matches(body, '```criteria', 'g')) AS criteria_fence_count
  FROM doc_plans WHERE source='{SOURCE}';""")
print("COMMIT;" if apply else "ROLLBACK;")
print(f"-- generator-side body length: {body_len} chars")
