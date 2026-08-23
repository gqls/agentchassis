#!/usr/bin/env python3
"""audit_writer_brief.py — audit the brief the WRITER ACTUALLY SEES, not the spec document.

WHY THIS EXISTS, AND WHY IT IS NOT `count_negation_tells.py` POINTED AT A SPEC.
On 2026-08-19 this lane published a fleet census of define-by-negation in
`content_direction` and had to withdraw the counts within hours: they were taken over
`data::text`, the whole spec DOCUMENT, and the writer does not read the document. It
reads a handful of named template fields. On `ai-agent-orchestration.com` that is 3,558
of 15,760 chars — so ~77% of the count was aimed at text with no consumer, and the
headline figure was roughly 2x reality (13 -> 2 for that site).

So the rule this tool enforces on itself: ESTABLISH WHICH TEXT REACHES THE CONSUMER
FIRST, then measure only that. The writer-visible surface is not hardcoded here — it is
derived at runtime from the live agent config by extracting `{{.site_specs...}}`
references from the prompt, so it cannot go stale the way a copied list does. Today it
resolves to five fields (content_direction.formatted, identity.key_differentiators,
identity.target_audience, evidence_base.writer_block, design_intent.imagery_direction).

WHAT IT REPORTS. Four things, in descending order of how well evidenced they are:

  1. SURFACE     — which writer-visible fields exist for this site, and their size.
  2. SILENT DROPS — `content_direction` keys that are in the stored document but reach
                   NO writer-visible field. `formatted` is computed from the incoming
                   partial BEFORE the deep merge (`site_spec_actions.go`), so a partial
                   write shrinks the brief to that partial's keys while the document
                   keeps growing. This is a MECHANICAL check: a key's humanised label is
                   either present in `formatted` or it is not.
  3. TELLS       — define-by-negation constructions, counted over the visible surface
                   only, attributed to the labelled block they sit in.
  4. SUPPLIED    — quoted phrases inside the visible surface that carry the construction.
                   These are the evidenced class: a phrase the brief HANDS the writer
                   transfers verbatim (measured: the canonical tagline of one site
                   appears in 1,348 rendered prompts and 408 responses). A supplied
                   phrase whose block also carries a mandate verb ("must appear",
                   "always use") is ranked highest, because the brief is not merely
                   modelling the construction, it is ordering it into a hero.

WHAT IT IS NOT. Never a gate. Instructional prose that merely uses a contrast to give
guidance ("use stack references naturally, not as buzzwords") is NOT evidenced to
transfer into output — that measurement has not been done and is designed to be able to
come out either way. This tool separates that class from the supplied class rather than
adding them up, which is the distinction the withdrawn census lacked.

Run:
  audit_writer_brief.py <domain> [<domain> ...]
  audit_writer_brief.py --fleet
  audit_writer_brief.py --transfer "<phrase>"   # prompt->response chain, the real test
  audit_writer_brief.py --self-test             # induced controls; MUST fail
"""
import json
import re
import subprocess
import sys

WRITER_AGENT = "page-content-writer"

# The construction, in the shapes `voicetells.go` codifies plus its near neighbours.
TELL_PATTERNS = [
    ("X, not Y", r",\s+not\s+\w"),
    ("rather than", r"\brather than\b"),
    ("instead of", r"\binstead of\b"),
    ("not just", r"\bnot just\b"),
    ("not a/an (predicate)", r"\bnot an?\s+\w+"),
]

# A brief that ORDERS a phrase into the page is stronger evidence than one that models it.
MANDATE_RE = re.compile(
    r"\b(must appear|must be used|always use|should appear|use this (?:exact|canonical)"
    r"|canonical tagline|verbatim)\b", re.I)

# Quoted spans: the shape a brief uses when it is handing over text to reuse.
#
# ⚠ The apostrophe is the trap and it produced junk on the first fleet run: a naive
# ['"] class treats the ' in "the client's own voice" as an opening quote and returns
# everything up to the next apostrophe as a "supplied phrase". So single quotes are
# only honoured when the opening mark is NOT preceded by a letter and the closing mark
# is NOT followed by one — which is exactly what separates a quotation from a
# possessive. Double quotes need no such guard.
QUOTED_RE = re.compile(
    r"[\"“”]([^\"“”]{12,220})[\"“”]"           # double-quoted, unambiguous
    r"|(?<![A-Za-z])['‘]([^'‘’]{12,220})['’](?![A-Za-z])")  # single, not a possessive


def quoted_spans(text):
    """Every quoted span, from either quoting style, with possessives excluded."""
    return [a or b for a, b in QUOTED_RE.findall(text)]


def psql(sql):
    """One psql round trip, tab-separated, no headers."""
    out = subprocess.run(
        ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
         "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A", "-F", "\t", "-c", sql],
        capture_output=True, text=True, timeout=180)
    if out.returncode != 0:
        sys.exit(f"psql failed: {out.stderr.strip()}")
    return out.stdout


def writer_visible_paths(agent=WRITER_AGENT):
    """Derive the writer-visible spec surface from the LIVE agent config.

    Never hardcode this list: the point of the tool is that the consumer decides what
    counts as the brief, and a copied list silently stops matching the consumer.
    """
    rows = psql(
        "SELECT DISTINCT m[1] FROM agent_definitions ad, "
        "LATERAL regexp_matches(ad.default_config::text, "
        "'\\{\\{[^}]*site_specs[^}]*\\}\\}', 'g') m "
        f"WHERE ad.type='{agent}' AND ad.is_active "
        "AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;")
    paths = set()
    for line in rows.splitlines():
        # {{if .site_specs.specs.x.y}} and {{.site_specs.specs.x.y}} both name the same field
        m = re.search(r"\.site_specs\.specs\.([A-Za-z0-9_.]+)", line)
        if m:
            paths.add(m.group(1))
    # An `{{if .site_specs.specs.content_direction}}` guard names the aspect, not a field;
    # drop it when a deeper path under the same aspect is also present.
    deeper = {p.split(".")[0] for p in paths if "." in p}
    return sorted(p for p in paths if "." in p or p not in deeper)


def humanise(key):
    """Mirror of datahelpers.HumaniseKey — snake_case to 'Title case' label."""
    k = key.replace("_", " ")
    return (k[:1].upper() + k[1:]) if k else k


def fetch_specs(paths, domains=None):
    """Every site's writer-visible aspects in ONE round trip, keyed by domain.

    Deliberately one query rather than one per site: 26 `kubectl exec` round trips is
    both fragile (one died mid-fleet-run with an EOF on stdin) and needless load on a
    database that may have a diagnosis run in flight — this lane has already had one
    loop time out under its own exploratory queries.
    """
    aspects = sorted({p.split(".")[0] for p in paths})
    inlist = ",".join("'" + a.replace("'", "''") + "'" for a in aspects)
    where = f"sp.aspect IN ({inlist}) AND sp.is_current"
    if domains:
        dlist = ",".join("'" + d.replace("'", "''") + "'" for d in domains)
        where += f" AND s.domain IN ({dlist})"
    # Returned as ONE json document, not delimited rows.
    #
    # ⚠ This is the second attempt and the first one silently LOST DATA. Splitting
    # psql's tab-separated output on Python's `splitlines()` dropped three sites from a
    # 25-site run and truncated a fourth to 2 chars — because `splitlines()` also breaks
    # on \r, \x0b, \x0c, \x1c-\x1e and \u2028, any of which can sit inside authored
    # spec prose, and a row that fails to parse is skipped in silence. The tell was a
    # site reporting `doc 2` (an empty object) where it had reported 37,606 minutes
    # earlier. Letting Postgres do the encoding removes the delimiter question entirely.
    raw = psql("SELECT COALESCE(jsonb_agg(jsonb_build_object("
               "'domain', s.domain, 'aspect', sp.aspect, 'data', sp.data)), '[]'::jsonb) "
               f"FROM site_specs sp JOIN sites s ON s.id=sp.site_id WHERE {where};")
    by_domain = {}
    for row in json.loads(raw):
        by_domain.setdefault(row["domain"], {})[row["aspect"]] = row["data"]
    return by_domain


def resolve_visible(specs, paths):
    """Pull the writer-visible fields out of one site's specs."""
    visible = {}
    for p in paths:
        cur, parts = specs, p.split(".")
        for part in parts:
            cur = cur.get(part) if isinstance(cur, dict) else None
            if cur is None:
                break
        if cur is not None:
            visible[p] = cur
    return visible


def blocks(formatted):
    """Split a `formatted` brief back into its (Label, body) blocks.

    FormatContentDirection joins top-level entries with a blank line and writes each as
    'Label: value' or 'Label:\\n- item'. Splitting on the blank line recovers them.
    """
    out = []
    for chunk in re.split(r"\n\s*\n", formatted or ""):
        chunk = chunk.strip()
        if not chunk:
            continue
        m = re.match(r"([A-Z][A-Za-z0-9 /'-]{0,60}):\s*(.*)", chunk, re.S)
        out.append((m.group(1), m.group(2)) if m else ("(unlabelled)", chunk))
    return out


def silent_drops(cd_spec):
    """Keys in the content_direction DOCUMENT that reach no writer-visible label.

    Discriminates a real loss from a benign one: an empty array/string formats to nothing
    and loses nothing, which is why the empty arm is reported separately rather than
    inflating the count. `compliance_rules: []` is the fleet's commonest such key.
    """
    formatted = cd_spec.get("formatted") or ""
    dropped, benign = [], []
    for key, val in cd_spec.items():
        if key == "formatted":
            continue
        # ⚠ The label is not always `Label:` exactly. A brief can carry
        # `Layout preservation (rerender rule):` — same key, qualified label — and an
        # exact-match test calls that key DROPPED when its 370 chars are right there.
        # Measured 2026-08-23 on loanzy.uk, where it produced a false "data is being
        # lost" reading about a write that had in fact worked correctly.
        # So: the label must START A LINE and reach a colon without crossing a newline.
        # Anchoring to the line start is what keeps this tight — a passing mention of the
        # words mid-sentence must NOT count, or the check stops seeing real drops, which
        # is the failure direction that matters.
        if re.search(r"^" + re.escape(humanise(key)) + r"[^\n]*:", formatted, re.M):
            continue
        chars = len(json.dumps(val, ensure_ascii=False)) if not isinstance(val, str) else len(val)
        empty = val in (None, "", [], {}) or (isinstance(val, (list, dict)) and not val)
        (benign if empty else dropped).append((key, chars))
    return sorted(dropped, key=lambda x: -x[1]), sorted(benign)


def as_text(v):
    """Flatten a spec value to the text the prompt will carry, for counting only."""
    if isinstance(v, str):
        return v
    if isinstance(v, list):
        return " ".join(as_text(x) for x in v)
    if isinstance(v, dict):
        return " ".join(as_text(x) for x in v.values())
    return ""


def count_tells(text):
    return [(name, len(re.findall(pat, text, re.I))) for name, pat in TELL_PATTERNS]


def has_tell(text):
    return any(re.search(pat, text, re.I) for _, pat in TELL_PATTERNS)


def supplied_phrases(label, value):
    """Phrases the brief HANDS the writer, as opposed to prose that instructs it.

    Two mechanically distinct routes, kept distinct because conflating them is how the
    withdrawn census went wrong — it added instruction and handover together:

      QUOTED   — the author put quote marks round it inside a prose block. That is a
                 handover in anybody's reading, and it is the shape the one proven
                 transfer chain has (a canonical tagline, 1,369 prompts -> 409 outputs).
      LIST ITEM — the field is a list that the writer's template injects verbatim
                 (`identity.key_differentiators` is rendered as its own text). Every
                 element is supplied by construction; no quote marks are involved.

    ⚠ The list arm was found by reading the FIRST fleet run's output rather than by
    design: the differentiators were being caught only because `json.dumps` had put
    quote marks round each element. Right answer, accidental mechanism — so it is now
    the stated rule instead of a side effect of serialisation.

    Reports, never judges. A regulatory disclaimer ("this is not a quote, offer, or
    financial advice") carries the construction and is required text; that is a human's
    call, and the tool's job is to put it in front of one.
    """
    hits = []
    if isinstance(value, list):
        for item in value:
            if isinstance(item, str) and has_tell(item):
                hits.append((item.strip(), False, label, "list item"))
        return hits
    if isinstance(value, dict):
        for k, v in value.items():
            hits += supplied_phrases(f"{label}.{k}", v)
        return hits
    if not isinstance(value, str):
        return hits
    mandated = bool(MANDATE_RE.search(value))
    for q in quoted_spans(value):
        if has_tell(q):
            hits.append((q.strip(), mandated, label, "quoted"))
    return hits


def audit(domain, paths, specs):
    visible = resolve_visible(specs, paths)
    if not specs:
        print(f"\n=== {domain} — NO current specs found (is the domain spelled as in `sites`?)")
        return None
    print(f"\n{'=' * 78}\n=== {domain}\n{'=' * 78}")

    print("\n1. WRITER-VISIBLE SURFACE (derived from the live agent config, not hardcoded)")
    total = 0
    for p in paths:
        v = as_text(visible.get(p))
        total += len(v)
        print(f"   {'PRESENT' if v else 'ABSENT ':8} {p:46} {len(v):>6} chars")
    cd = specs.get("content_direction", {})
    doc_chars = len(json.dumps(cd, ensure_ascii=False))
    print(f"   {'':8} {'TOTAL visible to the writer':46} {total:>6} chars"
          f"   (content_direction document alone is {doc_chars})")

    print("\n2. SILENT DROPS — in the content_direction document, absent from the brief")
    dropped, benign = silent_drops(cd)
    if not dropped:
        print("   none — every key with content reaches the brief")
    for key, chars in dropped:
        print(f"   DROPPED  {key:46} {chars:>6} chars never reach the writer")
    if benign:
        print(f"   (empty, nothing lost: {', '.join(k for k, _ in benign)})")

    print("\n3. TELLS in the visible surface only, by block")
    grand = 0
    for label, body in blocks(visible.get("content_direction.formatted") or ""):
        n = sum(c for _, c in count_tells(body))
        grand += n
        if n:
            print(f"   {n:>3}  {label}")
    for p in paths:
        if p == "content_direction.formatted" or not visible.get(p):
            continue
        n = sum(c for _, c in count_tells(as_text(visible[p])))
        grand += n
        if n:
            print(f"   {n:>3}  {p}")
    words = len(re.sub(r"\s+", " ", " ".join(as_text(v) for v in visible.values())).split())
    print(f"   {grand:>3}  TOTAL   ({words} words in the visible brief"
          f"{f', {grand / words * 1000:.1f} per 1,000' if words else ''})")

    print("\n4. SUPPLIED PHRASES carrying the construction — the evidenced transfer class")
    found = []
    for label, body in blocks(visible.get("content_direction.formatted") or ""):
        found += supplied_phrases(label, body)
    for p in paths:
        if p != "content_direction.formatted" and visible.get(p):
            found += supplied_phrases(p, visible[p])
    if not found:
        print("   none — the construction here is instructional only (transfer NOT evidenced)")
    for phrase, mandated, label, route in found:
        tag = "MANDATED" if mandated else f"supplied/{route.split()[0]}"
        print(f"   {tag:16} [{label}] \"{phrase[:110]}\"")
    if any(m for _, m, _, _ in found):
        print("   ^ MANDATED means the block also orders the phrase onto pages. Confirm the")
        print("     chain before acting:  audit_writer_brief.py --transfer \"<phrase>\"")
    return {"domain": domain, "visible": total, "doc": doc_chars,
            "dropped": len(dropped), "tells": grand, "supplied": len(found)}


def transfer(phrase):
    """The real test: does this phrase travel from rendered prompt into stored output?

    A lexical flag is a candidate; THIS is the evidence. Reports both halves because the
    interesting refutation is prompts=0 with responses>0 — the model's own phrasing, not
    transfer from the brief, which is how three flagged phrases were cleared on 08-19.
    """
    esc = phrase.replace("'", "''").replace("%", "\\%").replace("_", "\\_")
    rows = psql(
        "SELECT count(*) FILTER (WHERE prompt_rendered ILIKE '%" + esc + "%'), "
        "count(*) FILTER (WHERE response_text ILIKE '%" + esc + "%'), count(*) "
        f"FROM llm_call_log WHERE agent_type='{WRITER_AGENT}';")
    p, r, total = (rows.strip().split("\t") + ["0", "0", "0"])[:3]
    print(f'\nphrase: "{phrase}"')
    print(f"  rendered prompts containing it : {p}")
    print(f"  responses containing it        : {r}")
    print(f"  ({total} {WRITER_AGENT} calls in the log)")
    if int(p) == 0 and int(r) > 0:
        print("  => NOT transfer. It reaches no prompt, so the brief cannot be causing it;")
        print("     this is the model's own phrasing. Do not attribute it to the spec.")
    elif int(p) > 0 and int(r) > 0:
        print("  => TRANSFER EVIDENCED: supplied by the brief and emitted verbatim.")
    elif int(p) > 0:
        print("  => supplied to the writer, never emitted. The brief carries it; output does not.")
    else:
        print("  => absent from both. Check the phrase is quoted as it appears in the brief.")


SELF_TEST_SPEC = {
    "formatted": (
        "Voice: Direct and technical, not corporate.\n\n"
        "Emphasis: The canonical tagline 'Shipped in days, not months' must appear in "
        "the homepage hero.\n\n"
        "Blog strategy: Publish two posts a week with architecture diagrams."),
    "voice": "Direct and technical, not corporate.",
    "emphasis": "The canonical tagline ...",
    "blog_strategy": "Publish two posts a week ...",
    "things_to_avoid": ["hype vocabulary", "urgency language"],   # real content, dropped
    "compliance_rules": [],                                        # empty, benign
}


def self_test():
    """Every control, both arms. A check that cannot fail is decoration."""
    fails = []

    def check(name, got, want):
        ok = got == want
        print(f"  {'PASS' if ok else 'FAIL'}  {name}: got {got!r}, want {want!r}")
        if not ok:
            fails.append(name)

    print("\n-- silent_drops: discriminates real loss from an empty key")
    dropped, benign = silent_drops(SELF_TEST_SPEC)
    check("things_to_avoid is reported dropped", [k for k, _ in dropped], ["things_to_avoid"])
    check("empty compliance_rules is NOT a loss", [k for k, _ in benign], ["compliance_rules"])
    print("  (negative arm: a key present in `formatted` must not be reported)")
    check("voice reaches the brief, so is not dropped",
          "voice" in [k for k, _ in dropped] + [k for k, _ in benign], False)

    print("\n-- blocks: recovers the labelled blocks the formatter wrote")
    check("three blocks", [b[0] for b in blocks(SELF_TEST_SPEC["formatted"])],
          ["Voice", "Emphasis", "Blog strategy"])

    print("\n-- count_tells: fires on the construction, silent on clean prose")
    check("'not corporate' counted", sum(c for _, c in count_tells("Direct, not corporate.")), 1)
    check("clean prose counts zero",
          sum(c for _, c in count_tells("Publish two posts a week with diagrams.")), 0)

    print("\n-- supplied_phrases: separates a MANDATED handover from instructional contrast")
    emph = [b for b in blocks(SELF_TEST_SPEC["formatted"]) if b[0] == "Emphasis"][0]
    hits = supplied_phrases(*emph)
    check("the quoted tagline is caught", [h[0] for h in hits], ["Shipped in days, not months"])
    check("and is marked MANDATED", [h[1] for h in hits], [True])
    check("via the quoted route", [h[3] for h in hits], ["quoted"])

    print("\n-- supplied_phrases: the LIST arm (a field injected verbatim, no quote marks)")
    diffs = ["Verified, not hallucinated: agents check facts against official sources.",
             "We run the platform on our own sites first."]
    lh = supplied_phrases("identity.key_differentiators", diffs)
    check("only the element carrying the construction is flagged",
          [h[0] for h in lh], ["Verified, not hallucinated: agents check facts against official sources."])
    check("and it is attributed to the list route", [h[3] for h in lh], ["list item"])
    check("a clean list flags nothing", supplied_phrases("x", ["We ship on Kubernetes."]), [])
    voice = [b for b in blocks(SELF_TEST_SPEC["formatted"]) if b[0] == "Voice"][0]
    check("instructional contrast is NOT a supplied phrase", supplied_phrases(*voice), [])
    quoted_clean = ("Emphasis: use the tagline 'Built by engineers who run it' verbatim.")
    check("a quoted phrase WITHOUT the construction is not flagged",
          supplied_phrases(*blocks(quoted_clean)[0]), [])

    print("\n-- quoted_spans: an apostrophe is not a quote mark (the first fleet run's junk)")
    check("possessive does not open a span",
          quoted_spans("the client's own voice, not a long form and not a questionnaire"), [])
    check("a real single-quoted span is still caught",
          quoted_spans("use 'shipped in days, not months' verbatim"),
          ["shipped in days, not months"])
    check("a double-quoted span is caught",
          quoted_spans('use "shipped in days, not months" verbatim'),
          ["shipped in days, not months"])

    print("\n-- humanise: mirrors datahelpers.HumaniseKey")
    check("snake_case", humanise("things_to_avoid"), "Things to avoid")

    print(f"\n{'ALL CONTROLS PASS' if not fails else 'CONTROLS FAILED: ' + ', '.join(fails)}")
    return 1 if fails else 0


if __name__ == "__main__":
    args = sys.argv[1:]
    if not args:
        sys.exit(__doc__)
    if args[0] == "--self-test":
        sys.exit(self_test())
    if args[0] == "--transfer":
        transfer(" ".join(args[1:]))
        sys.exit(0)
    paths = writer_visible_paths()
    print(f"writer-visible spec surface for `{WRITER_AGENT}` (live config): "
          f"{len(paths)} fields")
    domains = None if args[0] == "--fleet" else args
    by_domain = fetch_specs(paths, domains)
    domains = domains or sorted(by_domain)
    rows = [r for r in (audit(d, paths, by_domain.get(d, {})) for d in domains) if r]
    if len(rows) > 1:
        print(f"\n{'=' * 78}\nFLEET SUMMARY — sorted by keys dropped from the brief\n{'=' * 78}")
        print(f"{'domain':34}{'visible':>8}{'doc':>8}{'dropped':>9}{'tells':>7}{'supplied':>10}")
        for r in sorted(rows, key=lambda r: (-r["dropped"], -r["tells"])):
            print(f"{r['domain']:34}{r['visible']:>8}{r['doc']:>8}"
                  f"{r['dropped']:>9}{r['tells']:>7}{r['supplied']:>10}")
