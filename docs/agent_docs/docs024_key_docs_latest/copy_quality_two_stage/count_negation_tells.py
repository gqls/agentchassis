#!/usr/bin/env python3
"""count_negation_tells.py — count define-by-negation constructions on a page.

WHY THIS IS AN OBSERVATION AND NOT A GATE. The v2 house voice says a matched
contrasting pair is "earned once or twice per page at most", and `voicetells.go`
codifies the shape (`strawmanCommaRe`, the defining-by-negation check at :212). That
makes the count meaningful — but it is a LEXICAL measure over prose, and this lane's
standing rule is that acceptance gates compare declared sets, types and structure,
never prose (`bugs_open/278` §8: same generator, same inputs, 2 of 4 card bodies
diverged with nothing wrong). So this reports a number for a human to weigh; it never
returns a pass/fail that anything should automate against.

It also cannot tell an EARNED contrast from a lazy one — "in days, not months" may be
the clearest way to say a true thing. That judgement is stage 2's and the reviewer's.

Counts, over visible text only (script/style/nav/header/footer stripped):
  * "X, not Y"      — comma followed by `not` (strawmanCommaRe's shape)
  * "rather than"   — the same move in a longer coat
  * "instead of", "not just", "isn't a" — the near neighbours

Run:  count_negation_tells.py <url|file> [<url|file> ...]
      count_negation_tells.py --component <page_component_id>
"""
import re, sys, html, subprocess, json

PATTERNS = [("X, not Y", r',\s+not\s+\w'), ("rather than", r'\brather than\b'),
            ("instead of", r'\binstead of\b'), ("not just", r'\bnot just\b'),
            ("not a/an (predicate)", r'\bnot an?\s+\w+')]


def visible(t):
    body = re.sub(r'(?s)<(script|style|nav|header|footer)\b.*?</\1>', ' ', t)
    return re.sub(r'\s+', ' ', html.unescape(re.sub(r'(?s)<[^>]+>', ' ', body))).strip()


def fetch(src):
    if src.startswith("http"):
        return subprocess.run(["curl", "-s", src], capture_output=True, text=True, timeout=60).stdout
    return open(src).read()


def report(label, raw):
    txt = visible(raw)
    total = 0
    print(f"\n{label}")
    for name, pat in PATTERNS:
        n = len(re.findall(pat, txt, re.I))
        total += n
        print(f"  {name:22} {n}")
    words = len(txt.split())
    per1k = (total / words * 1000) if words else 0
    print(f"  {'TOTAL':22} {total}   ({words} words, {per1k:.1f} per 1,000)")
    return total, words


if __name__ == "__main__":
    args = sys.argv[1:]
    if not args:
        sys.exit(__doc__)
    if args[0] == "--component":
        sql = ("SELECT COALESCE(string_agg(rendered_html,' '),'') FROM page_components "
               f"WHERE id = '{args[1]}';")
        raw = subprocess.run(["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
                              "--", "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A", "-c", sql],
                             capture_output=True, text=True, timeout=120).stdout
        report(f"component {args[1]}", raw)
    else:
        for a in args:
            report(a, fetch(a))
