#!/usr/bin/env python3
"""Generate the supersede-then-insert SQL that persists a BATCH of component
PLANs (each carrying its proven ```criteria fence) into `doc_plans`.

WHY A GENERATOR AND NOT A HEREDOC (RUNBOOK §9). The PLAN body contains triple
backticks — the ```criteria fence itself. Inside a double-quoted bash string
those are COMMAND SUBSTITUTION, and the mangling is SILENT: you would discover
it as a corrupted contract in production, not as an error at authoring time.
Batches 1-5 hand-rolled this generator in a session scratchpad each time and
lost it to tmp-cleanup each time (the 2026-08-05 handoff predicted exactly
that). This is the committed version.

WHAT IT ASSERTS, INSIDE THE TRANSACTION, BEFORE COMMIT
  1. `length(body)` equals the length Python built — the only check that proves
     psql did not interpolate a `:name` inside the literal.
  2. exactly one `is_current` row per (subject_type, subject_key) afterwards.
Both are DO/RAISE, not bare SELECTs: `ON_ERROR_STOP` ignores a non-empty result
set, so a verify block made of SELECTs cannot stop the COMMIT (LANDMINES,
RFC_006). A single `%` in a RAISE format string is a format directive — write
`%%` if you ever need a literal one.

SUBJECT_TYPE IS PER-ENTRY AND IT DECIDES WHETHER THE PLAN IS EVER READ.
The two consumers look under different types, and a PLAN filed under the wrong
one is not an error anywhere -- it is silently invisible:
  * components -> `component`, read by the component S6 dispatch;
  * TOOLS      -> `tool`, read by `tool-acceptance-agent`'s `load_docs`, whose
    live config is {"subject_type":"tool","subject_key_field":
    "input_data.spec.function"} (read out of `agent_definitions` 2026-08-10).
    A tool PLAN written as `component` makes `request_browser_run` skip with
    reason=needs_criteria -- "no fake pass", so it reads as a clean run that
    asserted NOTHING (tool_acceptance_run.sh's own trap #1).
The value is validated against the running binary's own vocabulary
(`validDocSubjectTypes`, doc_subjects_common.go:71) because it is interpolated
into SQL as a literal; so is `function`, hence the same treatment.

MANIFEST (JSON array, one object per subject):
  {
    "function":     "news-listing",
    "subject_type": "component",   (OPTIONAL, default "component"; use "tool"
                                    for anything component_level='tool')
    "kind":         "section component",  (OPTIONAL prose label for the title;
                                    defaults to "section component", or
                                    "tool" when subject_type is "tool")
    "batch":        "batch 7 -- the INTERACTIVE stock",   (OPTIONAL prose)
    "aim":        "one or two sentences",
    "contract":   ["- bullet", "- bullet"],
    "fence":      "fence_component_news_listing.json",   (relative to this dir)
    "proof_url":  "https://robot-hands.com/news/index.html",
    "proof_note": "single-instance; assets probed 200",
    "mutants":    "8/8 caught, every check watched red, baseline green",
    "accommodations": "serve_local /data/news-archive.json (same-origin fetch)",
    "notes":      ["- any extra PLAN paragraph"]            (optional)
  }

USAGE
  gen_component_plan_sql.py <manifest.json>            -> dry run (ROLLBACK)
  gen_component_plan_sql.py <manifest.json> --apply    -> COMMIT
Pipe the output to psql; never let the shell see the body.
"""
import json
import os
import re
import sys

DIR = os.path.dirname(os.path.abspath(__file__))
SOURCE = "staged_component_build"
CREATED_BY = "operator:staged_component_build"
AUTHORED = "2026-08-10"

# Mirrors validDocSubjectTypes (doc_subjects_common.go:71) as read at HEAD on
# 2026-08-10. Kept as a literal rather than derived: this is a belt-and-braces
# check on a value that becomes a SQL literal, and a drifting Go list should
# make this generator refuse, not quietly widen.
VALID_SUBJECT_TYPES = {
    "tool", "pipeline", "experience", "action",
    "experience-pattern", "component", "landmine", "decision",
}
# Both values below are interpolated into single-quoted SQL literals, so the
# only safe policy is a strict allow-list of shapes that cannot carry a quote.
SUBJECT_KEY_RE = re.compile(r"\A[a-z0-9][a-z0-9-]*\Z")


def subject_type_for(s: dict) -> str:
    st = s.get("subject_type", "component")
    if st not in VALID_SUBJECT_TYPES:
        raise SystemExit(
            f'subject_type {st!r} for {s.get("function")!r} is not one of '
            f'{sorted(VALID_SUBJECT_TYPES)} -- the Go gate (docResolveSubject) '
            f"would reject it and the DB CHECK would too"
        )
    return st


def subject_key_for(s: dict) -> str:
    key = s["function"]
    if not SUBJECT_KEY_RE.match(key):
        raise SystemExit(
            f"function {key!r} is not a bare kebab-case key; it is interpolated "
            f"into a SQL literal, so it is refused rather than escaped"
        )
    return key


def body_for(s: dict) -> str:
    fence_path = os.path.join(DIR, s["fence"])
    with open(fence_path, "r", encoding="utf-8") as fh:
        fence = fh.read().rstrip("\n")
    # Re-serialise so the stored fence is exactly what the evaluator parsed.
    fence = json.dumps(json.loads(fence), indent=2)

    st = subject_type_for(s)
    kind = s.get("kind") or ("tool" if st == "tool" else "section component")
    batch = s.get("batch", "batch 7 -- the INTERACTIVE stock")
    # The PLAN cites its own mutants file by name, so the default must not be a
    # guess: batches 1-7 all used mutants_component_<fn>.json and that stays the
    # default, but a subject whose file is named otherwise says so rather than
    # letting the PLAN point at a path that does not exist.
    mutants_file = s.get(
        "mutants_file", f'mutants_component_{s["function"].replace("-", "_")}.json'
    )
    if not os.path.exists(os.path.join(DIR, mutants_file)):
        raise SystemExit(
            f'{s["function"]}: mutants file {mutants_file!r} does not exist beside '
            f"the generator -- the PLAN would cite a path no reader can open. "
            f'Set "mutants_file" in the manifest.'
        )

    parts = [
        f'# PLAN -- {s["function"]} ({kind})',
        "",
        f'**Authored {AUTHORED} by lane `staged_component_build`** under D10 (exhaustive',
        f'backlog clearance), production-line {batch}. Mutation',
        f'proofs use `prove_fence_mutants_file.go` + `{mutants_file}`.',
        "",
        "## Aim",
        "",
        s["aim"],
        "",
        "## Behaviour contract",
        "",
        "Read from the live `content_components` row (html_template AND js_content) and the",
        f'served proof placement on {AUTHORED}:',
        "",
    ]
    parts += s["contract"]
    parts += ["", "## Acceptance criteria", "", "```criteria", fence, "```", ""]
    parts += [
        "## Proof (S1+S2 discipline, " + AUTHORED + ")",
        "",
        f'- Proof placement: {s["proof_url"]}. {s["proof_note"]}',
        "  Placements move -- re-verify the row AND the served markup before any future run.",
        "- `try_fence.go` against the LIVE url: all evaluations passed; arithmetic reconciled.",
        f'- Mutation-proven with `prove_fence_mutants_file.go`: {s["mutants"]}.',
    ]
    if s.get("accommodations"):
        parts += [
            f'- Declared harness accommodation, uniform across baseline and every mutant:'
            f' {s["accommodations"]}.',
        ]
    parts += [
        "- Fence vocabulary pod-grepped in the DEPLOYED browser-runner before authoring, not",
        "  inferred from HEAD (the offline harnesses run HEAD's evaluator -- vocabulary newer",
        "  than the deployed binary passes offline and skips live).",
    ]
    if s.get("notes"):
        parts += ["", "## Notes"] + [""] + s["notes"]
    return "\n".join(parts) + "\n"


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    apply = "--apply" in sys.argv[2:]
    with open(sys.argv[1], "r", encoding="utf-8") as fh:
        subjects = json.load(fh)

    out = ["BEGIN;"]
    for s in subjects:
        body = body_for(s)
        if "$planbody$" in body:
            sys.stderr.write("dollar-quote tag collides with body text\n")
            return 2
        key = subject_key_for(s)
        st = subject_type_for(s)
        out.append("")
        out.append(f"-- ---------- {st}/{key} ----------")
        out.append(
            "UPDATE doc_plans SET is_current=false, superseded_at=now(), updated_at=now()\n"
            f" WHERE subject_type='{st}' AND subject_key='{key}' AND is_current;"
        )
        out.append(
            "INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)\n"
            f"VALUES ('{st}','{key}', $planbody${body}$planbody$,"
            f" '{SOURCE}', '{CREATED_BY}');"
        )
        out.append(
            "DO $$\nDECLARE n int; c int;\nBEGIN\n"
            f"  SELECT length(body), count(*) OVER () INTO n, c FROM doc_plans\n"
            f"   WHERE subject_type='{st}' AND subject_key='{key}' AND is_current;\n"
            f"  IF n IS DISTINCT FROM {len(body)} THEN\n"
            f"    RAISE EXCEPTION 'body length for {st}/{key} is %, expected {len(body)}', n;\n"
            "  END IF;\n"
            f"  IF c IS DISTINCT FROM 1 THEN\n"
            f"    RAISE EXCEPTION 'expected exactly 1 current row for {st}/{key}, found %', c;\n"
            "  END IF;\n"
            f"  RAISE NOTICE 'OK {st}/{key}: % bytes, 1 current row', n;\n"
            "END $$;"
        )
    out.append("")
    out.append("COMMIT;" if apply else "ROLLBACK;  -- dry run: pass --apply to commit")
    print("\n".join(out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
