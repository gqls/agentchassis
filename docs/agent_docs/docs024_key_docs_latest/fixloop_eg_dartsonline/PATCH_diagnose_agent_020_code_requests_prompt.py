#!/usr/bin/env python3
"""PATCH_diagnose_agent_020_code_requests_prompt.py

The CONFIG half of the diagnosis-side code-search tier. 2026-07-19.
Go half: commit 927b11ba0 (code_requests through wire -> route -> gather).
Plan: DESIGN_diagnosis_side_code_tier.md.

SEQUENCING — APPLY ONLY AFTER A CHASSIS IMAGE CARRYING 927b11ba0 IS ROLLED.
Verify first, against the POD binary and never the tag or git:

  kubectl exec -n ai-persona-system <chassis-pod> -- \
    sh -c 'strings /app/agent-chassis | grep -c code_requests_field'

Applying early is not destructive (a model emitting code_requests at an old
image has them ignored) but it is a wasted iteration every run: the verdicter
is invited to ask questions nothing will answer, and an unanswered question
reads back to it exactly like an EMPTY answer — i.e. "the mechanism is absent",
which is the wrong answer this tier most needs to avoid giving.

WHAT IT DOES (two edits to diagnose-agent's `verdict` prompt):
  1. Adds rule 10 — the code_requests channel, when to use it, and the two
     traps: answers are STATIC tier (they cannot satisfy the observed half of
     the two-evidence-family guard), and an empty answer is UNKNOWN, not ABSENT
     (the index is a snapshot; it can lag the fetched ref).
  2. Adds `code_requests` to the output JSON schema, beside data_requests.

WHY A RULE AT ALL: the channel is useless unless the verdicter knows it exists.
The council's equivalent tier needed the same prompt paragraph before reviewers
used it — the capability shipping and the capability being USED are two
different dates.

IDEMPOTENT: if rule 10 is already present it reports and exits without writing,
so two threads applying it concurrently cannot duplicate anything.

PATCH-STYLE, not a whole-object write (the config re-seed clobber landmine:
fix-proposer was re-seeded ~5x in 18h by other threads, each whole-object write
discarding concurrent work). This touches ONE json path and snapshots first.

USAGE:  python3 PATCH_...py           # dry run, prints the diff
        python3 PATCH_...py --apply   # snapshot, then write
"""
import json
import subprocess
import sys

PSQL = [
    "kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
    "psql", "-U", "clients_user", "-d", "clients_db",
]

RULE_ANCHOR = "## Output — return ONLY this JSON, nothing else"

NEW_RULE = """10. **`code_requests` search the CODEBASE — use them when the question is
   "does this exist elsewhere?"** `next_scope` follows the call graph: it reaches
   only code the current scope already touches. When your hypothesis turns on
   whether a mechanism exists SOMEWHERE ELSE — a second implementation, another
   call site, anything referencing a symbol — do NOT guess and do NOT abstain
   for want of the answer: emit a `code_request`. It is answered next iteration
   from the code_symbols index. Three kinds, and only these three:
   `symbol` (match a symbol name, e.g. `GenerateText`), `content` (match source
   text, e.g. `%stop_reason%`), `ls` (list indexed paths under a prefix, e.g.
   `platform/aiservice/`). You supply only a pattern — never SQL.
   TWO TRAPS, both of which have burned real runs:
   (a) **The answers are STATIC evidence.** They are code. They can show a
   mechanism EXISTS; they can never show it OCCURRED. A CONFIRM still needs a
   `state`/`runtime` citation showing it happening — cite a code_request answer
   at tier `static`, and if that is all you hold, use `data_requests` to fetch
   the observing rows (rule 9).
   (b) **An empty answer means UNKNOWN, not ABSENT.** The index is a snapshot
   and can lag the code the bundle was cut from. "No matches" is never evidence
   that something does not exist — never cite an empty result as proof of
   absence. If absence is load-bearing to your hypothesis, say so in
   `needed_evidence` and let a human settle it.

"""

SCHEMA_ANCHOR = """  "symptom_check": ["""

NEW_SCHEMA = """  "code_requests": [
    {"kind": "symbol | content | ls",
     "query": "the pattern to match — a symbol name, source text, or a path prefix. NOT SQL",
     "why": "what finding (or not finding) this would settle"}
  ],
"""


def fetch():
    out = subprocess.run(
        PSQL + ["-t", "-A", "-c",
                "SELECT default_config::text FROM agent_definitions "
                "WHERE type='diagnose-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"],
        capture_output=True, text=True, check=True).stdout.strip()
    return json.loads(out)


def main():
    apply = "--apply" in sys.argv
    cfg = fetch()
    step = cfg["workflow"]["steps"]["verdict"]
    tpl = step["config"]["prompt_template"]

    if "10. **`code_requests` search the CODEBASE" in tpl:
        print("ALREADY APPLIED — rule 10 present. No write.")
        return 0

    if RULE_ANCHOR not in tpl:
        print("REFUSING: rule anchor not found; the prompt has changed shape.\n"
              "Re-read the live prompt and re-anchor rather than forcing this.", file=sys.stderr)
        return 1
    if SCHEMA_ANCHOR not in tpl:
        print("REFUSING: schema anchor not found; the output block has changed shape.", file=sys.stderr)
        return 1

    new = tpl.replace(RULE_ANCHOR, NEW_RULE + RULE_ANCHOR, 1)
    new = new.replace(SCHEMA_ANCHOR, NEW_SCHEMA + SCHEMA_ANCHOR, 1)

    print(f"prompt: {len(tpl)} -> {len(new)} chars (+{len(new) - len(tpl)})")
    print("\n--- rule 10 inserted before the Output section ---")
    print(NEW_RULE)
    print("--- code_requests inserted into the output schema ---")
    print(NEW_SCHEMA)

    if not apply:
        print("DRY RUN — re-run with --apply to write (snapshots first).")
        return 0

    # DOLLAR-QUOTED, not a JSON-encoded string embedded in SQL text.
    #
    # The first version of this script did `{json.dumps(json.dumps(new))}::jsonb`
    # and psql died with `invalid command \` (2026-07-19). Two faults in one line:
    # the payload was double-encoded, and — the fatal one — psql reads piped SQL
    # line by line and treats ANY line beginning with a backslash as a
    # meta-command, so an escaped sequence landing at a line start is executed as
    # `\d`-style input rather than data. Dollar-quoting sidesteps the escaping
    # question entirely (the same shape used for the doc_notes insert that
    # worked), and to_jsonb() does the text->JSON-string conversion server-side
    # where it cannot be mangled in transit.
    #
    # The transaction did roll back cleanly on that failure (ON_ERROR_STOP + the
    # explicit BEGIN), leaving no stray snapshot — verified, not assumed.
    tag = "prompt020"
    if f"${tag}$" in new:
        print(f"REFUSING: the new prompt contains the dollar-quote tag ${tag}$", file=sys.stderr)
        return 1
    sql = f"""
BEGIN;
SELECT snapshot_agent('diagnose-agent', 'pre-update: 020 — verdict prompt gains the code_requests channel (diagnosis-side code tier)');
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{{workflow,steps,verdict,config,prompt_template}}',
       to_jsonb(${tag}${new}${tag}$::text))
 WHERE type='diagnose-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
COMMIT;
"""
    r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql, capture_output=True, text=True)
    print(r.stdout, r.stderr)
    return r.returncode


if __name__ == "__main__":
    sys.exit(main())
