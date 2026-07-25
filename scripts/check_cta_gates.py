#!/usr/bin/env python3
"""check_cta_gates.py — ADVISORY standing lint for bugs_open/023 (CTA label/destination pairing).

Flags any ACTIVE content_component that can render a button whose destination is
absent — the defect the owner reported as "I don't understand what these buttons
are and what they do, I think they are broken" (leopardessconsulting.co.uk,
2026-07-19). Four findings, all decidable from the template + schema alone:

  UNGATED     <a href="{{.x_url}}"> with no enclosing {{if}} on that same field.
              An unset field renders href="" — a control that looks live and goes
              nowhere. Platform invariant LNK-005 says render nothing instead.
  PLACEHOLDER <a href="#"> or href="{{if .x}}{{.x}}{{else}}#{{end}}" — the same
              dead control wearing a '#'. The {{else}}# form is worse than an
              ungated anchor because it survives every check that looks for "".
  LLM_URL     a *_url / *_link field declared source:llm AND required:true. That
              pair instructs a model to author a value it cannot look up, so it
              invents one: leopardess.contactforsales.com was the site's own
              contact address with @ swapped for '.', and finetuning.ai (a real
              third party's page) was the different-TLD variant of the same move.
  NO_VALUE    a template containing the literal '<no value>' — a Go render
              artefact saved back as a template, so the component ships a broken
              href to every site that adopts it.

WHY A SCRIPT AND NOT scripts/pattern-check.py: content_components live only in
the database, never in the repo, so a diff-based file linter cannot see them.
This is an OPERATIONAL check — run it by hand, after any component seed, and from
a sweep. Same reasoning as scripts/check_list_empty_states.py, same exit codes.

ADVISORY: exits non-zero on findings so a caller can decide; it is NOT wired into
the pre-commit hook (a DB round-trip has no place blocking a shared commit).
Exit 0 = clean, 1 = findings, 2 = could not reach the DB.

DELIBERATE EXCLUSIONS — each one cost a wrong measurement to learn:

  * Anchors inside {{range}} are ITEM links, not CTAs: the field belongs to the
    ranged item ({{range .items}}<a href="{{.url}}">), fed by a query-provided
    list. Different class, different owner, and gating them can delete a whole
    card rather than one control (image-hover-card-grid's anchor wraps the card
    image and title). Reported separately under --show-item-links, never as a
    failure. 17 of them existed at the time of writing; a blanket "no ungated url
    anchor" rule trips on them and rolls back correct migrations — that is
    exactly what migration 181's first draft did.
  * href="#" on an element that is hidden or carries a data-*template attribute
    is a JS clone source (provocations-archive-list), not a control. Gating it
    breaks the feature it belongs to.

THE PARSE IS REAL, NOT A REGEX. It tokenises each template and maintains an
{{if}}/{{range}}/{{with}}..{{end}} block stack; an anchor counts as gated only
when an enclosing block's condition references the SAME field. The heuristic it
replaced (RUNBOOK R9's 60-char lookback) undercounted by 2.4x, because
regexp_matches(...,'g') returns non-overlapping matches and the greedy prefix
eats the previous anchor — so in runs of adjacent anchors, which is exactly where
CTAs cluster, every other one vanished from the count.

TWIN: docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/parse_gates.py
holds the same block-stack parse for ad-hoc worklists. If you change the parse
here, change it there — two hand-maintained copies of one rule is the drift class
this bug's own council reviews for.
"""
import argparse
import collections
import json
import re
import subprocess
import sys

QUERY = """
SELECT jsonb_agg(jsonb_build_object(
         'name', name, 'function', function,
         'tpl', COALESCE(html_template, ''),
         'fields', COALESCE(input_schema->'fields', '{}'::jsonb)))
  FROM content_components WHERE is_active;
"""

PSQL = ["kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-tAc"]

ACTION = re.compile(r"\{\{-?\s*(.*?)\s*-?\}\}", re.S)
HREF = re.compile(r'<a\b[^>]*?href="\s*\{\{-?\s*\.([A-Za-z0-9_]+)\s*-?\}\}\s*"', re.S)
FIELDREF = re.compile(r"\.([A-Za-z0-9_]+)")
HASH_HREF = re.compile(r'<a\b([^>]*?)href="#"', re.S)
IF_HREF = re.compile(r'<a\b[^>]*?href="\{\{\s*if\s+\.([A-Za-z0-9_]+)', re.S)
URL_FIELD = re.compile(r"(^|_)(url|link)$")


def blocks(tpl):
    """(start, end, condition) for every if/range/with block in the template."""
    stack, out = [], []
    for m in ACTION.finditer(tpl):
        body = m.group(1)
        head = body.split()[0] if body.split() else ""
        if head in ("if", "range", "with"):
            stack.append((m.start(), body))
        elif head == "end" and stack:
            start, cond = stack.pop()
            out.append((start, m.end(), cond))
    return out


def scan(components):
    """Return (findings, item_links). findings = list of (kind, component, detail)."""
    findings, item_links = [], []
    for c in components:
        tpl = c.get("tpl") or ""
        blks = blocks(tpl)
        for m in HREF.finditer(tpl):
            field, pos = m.group(1), m.start()
            enclosing = [cond for (s, e, cond) in blks if s < pos < e]
            if any(field in FIELDREF.findall(cond) for cond in enclosing):
                continue
            if any(cond.split()[:1] == ["range"] for cond in enclosing):
                item_links.append((c["name"], field))
                continue
            findings.append(("UNGATED", c["name"], field))
        for m in HASH_HREF.finditer(tpl):
            attrs = m.group(1)
            if "hidden" in attrs or re.search(r"data-[a-z-]*template", attrs):
                continue          # JS clone source, not a control
            findings.append(("PLACEHOLDER", c["name"], 'href="#"'))
        for m in IF_HREF.finditer(tpl):
            findings.append(("PLACEHOLDER", c["name"],
                             'href="{{if .%s}}...{{else}}#" — gate the anchor instead' % m.group(1)))
        if "<no value>" in tpl:
            findings.append(("NO_VALUE", c["name"],
                             "template is a saved render artefact (%d occurrence(s))" % tpl.count("<no value>")))
        for field, spec in (c.get("fields") or {}).items():
            if not isinstance(spec, dict):
                continue
            if (URL_FIELD.search(field) and spec.get("source") == "llm"
                    and str(spec.get("required")).lower() == "true"):
                findings.append(("LLM_URL", c["name"], field))
    return findings, item_links


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--json", help="read components from a file instead of the cluster "
                                   "(same shape as the QUERY above)")
    ap.add_argument("--show-item-links", action="store_true",
                    help="also list range-scoped item links (never a failure)")
    args = ap.parse_args()

    if args.json:
        components = json.load(open(args.json))
    else:
        try:
            out = subprocess.run(PSQL + [QUERY], capture_output=True, text=True,
                                 timeout=90, check=True).stdout.strip()
        except Exception as exc:                                  # noqa: BLE001
            print("could not reach the database: %s" % exc, file=sys.stderr)
            return 2
        components = json.loads(out) if out else []

    findings, item_links = scan(components)
    print("checked %d active components" % len(components))

    if args.show_item_links:
        by = collections.Counter(n for n, _ in item_links)
        print("\nrange-scoped item links (separate class, NOT a failure): %d across %d components"
              % (len(item_links), len(by)))
        for name, n in sorted(by.items()):
            print("   %2d  %s" % (n, name))

    if not findings:
        print("\nCLEAN — no active component can pair a rendered label with an absent destination.")
        return 0

    by_kind = collections.defaultdict(list)
    for kind, name, detail in findings:
        by_kind[kind].append((name, detail))
    print("\n%d finding(s):" % len(findings))
    for kind in ("UNGATED", "PLACEHOLDER", "LLM_URL", "NO_VALUE"):
        rows = by_kind.get(kind)
        if not rows:
            continue
        print("\n  %s (%d)" % (kind, len(rows)))
        for name, detail in sorted(rows):
            print("    %-38s %s" % (name, detail))
    print("\nFix shape: docs/agent_docs/sql_for_agents/211_cta_gate_every_active_anchor.sql")
    return 1


if __name__ == "__main__":
    sys.exit(main())
