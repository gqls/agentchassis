#!/usr/bin/env python3
"""audit_prompt_demonstrations.py — phase 1 of the 2026-08-25 prompt audit.

WHAT THIS IS AND IS NOT. The owner's question is "is this prompt contributing to
good readable copy or encouraging AI styles of writing?" — a JUDGMENT question.
This script does not answer it. It answers the ORDERING question phase 1 exists
for: which prompts DEMONSTRATE the most, so the judgment pass reads the biggest
teachers first. The premise is measured, not assumed: an instruction is also an
example, and tell classes track their demonstration counts (REFRESH §2 — 3→0 when
demos removed, 6→8 when left standing). A LOW score here is NOT a clean bill: the
pattern list's ceiling is measured too (REFRESH §3), and the wider-register rows
below are lexical PROXIES for ordering, nothing more.

Populations scanned (PLAN_2026-08-25_prompt_audit §1), extracted live:
  A. agent_definitions.default_config — every string under a key matching
     (prompt|instruction|guidance)[a-z_]* at ANY depth (recursive walk in Python,
     because the sub_workflow census trap is a JSON-path depth trap: LANDMINES
     :7807, a steps-only walk reported 6 of a true 7).
  B. content_components.input_schema — every llm_guidance string, aggregated per
     component.
  C. site_specs content_direction — data->'formatted' ONLY (the writer's wire;
     the array forms are unread — REFRESH §5).
  D. Go prompt literals — backtick raw strings >200 chars in the files grep names
     as prompt-constructing (approximate BY DESIGN; sizes the previously-unsized
     population and scans what it finds; double-quoted literals not parsed).
  E. the three workflow JSON columns (task_workflow etc.) — same walk as A; this
     sizes the census's other unsized population.

Output: one markdown league table (stdout), sorted by demonstrations, with
per-1,000-char density and 30-day call volume per agent type where available.
Every row is dated by the run. Plumbing strings (<100 chars) are counted but
flagged, not scanned.
"""
import json
import re
import subprocess
import sys
from datetime import date

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A"]

KEY_RE = re.compile(r'(prompt|instruction|guidance)[a-z_]*$', re.I)

# The gate's five shapes as lexical proxies (negationtells.go is the authority in
# Go; these deliberately mirror count_negation_tells.py so the two tools agree),
# plus wider-register PROXIES from the owner's named tells (REFRESH §3). Proxy
# rows order the reading; they do not convict a prompt.
NEGATION = [
    ("x_not_y", r',\s+not\s+\w'),
    ("not_x_but_y", r'\bnot\s+[^.,;]{1,40}\s+but\s+'),
    ("rather_than", r'\brather than\b'),
    ("instead_of", r'\binstead of\b'),
    ("not_just", r'\bnot (?:just|only|merely)\b'),
    ("negative_reveal", r"\b(?:isn't|is not|aren't|are not)\s+(?:a|an|about|the)\b"),
]
PROXY = [
    ("scaffold", r'\bcomes down to\b|\bfor the same reason\b|\bin short[,:]|\bput simply\b|\bthe (?:key|important|real) (?:point|thing|decision|question)\b'),
    ("candour_beat", r'\bplainly\b|\bhonest(?:ly|y)?\b|\btransparent\b|\bno hype\b|\bwe don.t pretend\b|\bwon.t find\b|\bwill not (?:do|tell|sell)\b'),
    ("presumption", r'\bmost (?:visitors|readers|people|users)\b|\byou.re probably\b|\bchances are\b'),
    ("word_weight", r'\bcrucially\b|\bgenuinely\b|\bseamless\b|\brobust\b|\bdelve\b|\bat its core\b|\bin essence\b|\bfurthermore\b|\bmoreover\b|\bleverage\b'),
    ("em_dash", r' — '),
]


def psql_json_lines(sql):
    r = subprocess.run(PSQL + ["-c", sql], capture_output=True, text=True, timeout=300)
    if r.returncode != 0:
        print(f"  (query failed, population skipped: {r.stderr.strip().splitlines()[-1] if r.stderr else '?'})",
              file=sys.stderr)
        return []
    out = []
    for line in r.stdout.splitlines():
        line = line.strip()
        if line:
            try:
                out.append(json.loads(line))
            except json.JSONDecodeError:
                pass
    return out


def walk(node, path=""):
    """Yield (json_path, string) for every prompt-keyed string at any depth."""
    if isinstance(node, dict):
        for k, v in node.items():
            p = f"{path}.{k}" if path else k
            if isinstance(v, str) and KEY_RE.search(k):
                yield p, v
            else:
                yield from walk(v, p)
    elif isinstance(node, list):
        for i, v in enumerate(node):
            yield from walk(v, f"{path}[{i}]")


def scan(text):
    counts = {}
    for name, pat in NEGATION + PROXY:
        counts[name] = len(re.findall(pat, text, re.I))
    counts["_neg"] = sum(counts[n] for n, _ in NEGATION)
    counts["_proxy"] = sum(counts[n] for n, _ in PROXY)
    return counts


def main():
    rows = []  # (population, ident, chars, counts)
    plumbing = 0

    # A + E — agent_definitions, all four JSON columns
    cols = "default_config, task_workflow, orchestrator_workflow, orchestration_workflow"
    for rec in psql_json_lines(
            "SELECT json_build_object('type', type, 'cfgs', json_build_array(" + cols + ")) "
            "FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false "
            "AND deleted_at IS NULL;"):
        for ci, cfg in enumerate(rec["cfgs"] or []):
            col = ["default_config", "task_workflow", "orchestrator_workflow",
                   "orchestration_workflow"][ci]
            for path, text in walk(cfg or {}):
                if len(text) < 100:
                    plumbing += 1
                    continue
                rows.append((f"A:{col}" if ci == 0 else f"E:{col}",
                             f"{rec['type']} · {path}", len(text), scan(text)))

    # B — llm_guidance per component
    for rec in psql_json_lines(
            "SELECT json_build_object('name', name, 'schema', input_schema) "
            "FROM content_components WHERE COALESCE(is_active,true);"):
        strings = [t for _, t in walk(rec["schema"] or {}) ]
        guidance = " ".join(t for p, t in walk(rec["schema"] or {}) if "llm_guidance" in p)
        if guidance:
            rows.append(("B:llm_guidance", rec["name"], len(guidance), scan(guidance)))

    # C — content_direction.formatted (the writer's wire only)
    for rec in psql_json_lines(
            "SELECT json_build_object('domain', s.domain, 'formatted', sp.data->'formatted') "
            "FROM site_specs sp JOIN sites s ON s.id = sp.site_id "
            "WHERE sp.aspect='content_direction' AND sp.is_current;"):
        f = rec.get("formatted")
        if isinstance(f, str) and f:
            rows.append(("C:brief.formatted", rec["domain"], len(f), scan(f)))

    # D — Go backtick literals in prompt-constructing files
    grep = subprocess.run(
        ["bash", "-c",
         "cd /home/ant/projects/agentchassis && grep -rlE "
         "'buildPrompt|PromptTemplate|prompt :=|prompt =|systemPrompt|system_prompt' "
         "platform/ internal/ pkg/ cmd/ --include='*.go' | grep -v _test.go"],
        capture_output=True, text=True)
    go_chars = 0
    for fp in grep.stdout.split():
        try:
            src = open("/home/ant/projects/agentchassis/" + fp).read()
        except OSError:
            continue
        for lit in re.findall(r'`([^`]{200,})`', src):
            go_chars += len(lit)
            rows.append(("D:go_literal", fp.split("/")[-1], len(lit), scan(lit)))

    # F — the house voice block. NOT under any prompt-keyed path: it is one row of
    # agent_default_configs named voice_style_block, text at config->>'text'
    # (platform/voicestyle/voicestyle.go:35), injected into every template that
    # writes {{.voice_style}} — so it is in EVERY rendered writer prompt while
    # being invisible to a walk over agent_definitions. The first run of this
    # script missed it for exactly that reason.
    for rec in psql_json_lines(
            "SELECT json_build_object('t', config->>'text') FROM agent_default_configs "
            "WHERE config_name = 'voice_style_block';"):
        if rec.get("t"):
            rows.append(("F:house_voice", "voice_style_block (every {{.voice_style}} template)",
                         len(rec["t"]), scan(rec["t"])))

    # R — RENDERED prompts, the per-call truth: template + voice + brief + guidance
    # + facts assembled. The template scores are what we WROTE; these are what the
    # model READS. Latest 3 calls per writer-adjacent agent over 3 days (bounded —
    # llm_call_log is large and an unbounded scan times out). Writer step names
    # are process_sections_loop_iter_N_generate_content, hence the LIKE.
    for agent, step_like in [("page-content-writer", "%generate_content"),
                             ("copy-editor", "%"), ("content-gap-planner", "%"),
                             ("offer-analyser", "%"), ("build-site-planner", "%")]:
        for rec in psql_json_lines(
                "SET statement_timeout='100s'; "
                "SELECT json_build_object('created', created_at, 'step', step_name, 'p', prompt_rendered) "
                f"FROM llm_call_log WHERE agent_type='{agent}' AND step_name LIKE '{step_like}' "
                "AND created_at > now() - interval '3 days' AND prompt_rendered IS NOT NULL "
                "ORDER BY created_at DESC LIMIT 3;"):
            if rec.get("p"):
                rows.append(("R:rendered", f"{agent} · {rec['step']} @ {rec['created'][:16]}",
                             len(rec["p"]), scan(rec["p"])))

    # 30-day call volume per agent type, for the ordering column
    vol = {}
    for rec in psql_json_lines(
            "SELECT json_build_object('t', agent_type, 'n', count(*)) FROM llm_call_log "
            "WHERE created_at > now() - interval '30 days' GROUP BY agent_type;"):
        vol[rec["t"]] = rec["n"]

    rows.sort(key=lambda r: (r[3]["_neg"], r[3]["_proxy"]), reverse=True)

    print(f"# Phase-1 league table — prompt demonstrations, counted {date.today()}")
    print(f"\n{len(rows)} scanned strings ({plumbing} plumbing strings <100 chars counted, not scanned); "
          f"Go backtick literal volume: {go_chars:,} chars (population D, previously unsized).\n")
    print("| pop | prompt | chars | neg total | neg/1k | " +
          " | ".join(n for n, _ in NEGATION) + " | proxy total | " +
          " | ".join(n for n, _ in PROXY) + " | calls 30d |")
    print("|" + "---|" * (7 + len(NEGATION) + len(PROXY)))
    for pop, ident, chars, c in rows:
        agent = ident.split(" · ")[0]
        per1k = c["_neg"] / chars * 1000 if chars else 0
        print(f"| {pop} | {ident[:90]} | {chars:,} | {c['_neg']} | {per1k:.1f} | " +
              " | ".join(str(c[n]) for n, _ in NEGATION) + f" | {c['_proxy']} | " +
              " | ".join(str(c[n]) for n, _ in PROXY) + f" | {vol.get(agent, '')} |")


if __name__ == "__main__":
    main()
