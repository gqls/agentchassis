#!/usr/bin/env python3
"""concept-register-drift-check — does the concept register still agree with itself?

Runs daily (K8s CronJob). Reads `docs/agent_docs/docs026_concept_register/register/`
at a pinned git ref and answers four questions that nothing else in the estate
asks, all of them about INTERNAL consistency rather than accuracy:

  1. entry with no index row   — a concept exists in a category file and is
                                 missing from 000_concept_index.md
  2. index row with no entry   — the reverse
  3. duplicate id             — the same concept id used by two entries, or
                                 listed by two index rows
  4. headline drift           — the "**N index table rows**" figure in the
                                 index header disagrees with the actual count

WHY THIS EXISTS, and why the obvious check does not catch it.
On 2026-08-04 the register held 1,756 concepts and the index listed 1,722: **34
concepts had a full register entry and NO index row**, including the whole first
half of the claims-verification layer. The index is the file a session or a
council seat searches to learn whether a mechanism exists, so those 34 were
invisible in exactly the lookup they exist for — a search reports them as not
existing, which is the failure the register was built to prevent.

It survived roughly twenty recorded re-measurements because of HOW the header
number was maintained: each thread counted index rows and compared the total to
the PREVIOUS index-row count ("1,720 -> 1,721, exactly my row"). That confirms
your own row landed. It is structurally blind to a row nobody ever wrote. Only a
comparison against an independent source — the category files themselves — can
see it, and until this check nothing compared the two.

Drift is one-directional and that is not luck: adding a concept is two edits in
two files, and only the first is load-bearing for the author, so the index row is
the half that gets skipped. Check 2 has never fired on real data; it is here
because a check that can only fail one way teaches you nothing the day the
failure inverts.

REPORTS, NEVER REPAIRS. Backfilling a row needs a one-line summary written by
someone who understands the concept; a generated one would be worse than the gap,
because it would look authored. This writes one doc_notes row per run and stops.

Advisory: it exits 0 whether or not it finds drift. A watcher that fails a job on
a backlog gets suspended within a week, and then watches nothing.

Stdlib only — a fresh postgres:16-alpine container needs `apk add python3` and
nothing else.

Ref discipline, copied from bugs-open-staleness-sweep (RFC_005 §3.3) which
learned it from 090_TRIGGER_needs_diagnosis_v1.sh: REGISTER_REF has NO default.
The CronJob manifest pins it and a human bumps it deliberately when the
platform's live working branch changes — the same manual attention IMAGE_TAG
needs. resolve_ref_sha() fails the whole run rather than falling back to
anything, because a silently stale ref makes every finding here unfalsifiable:
the register it describes would not be the register anyone is editing.
"""
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request

REPO_OWNER = os.environ.get("REGISTER_REPO_OWNER", "gqls")
REPO_NAME = os.environ.get("REGISTER_REPO_NAME", "agentchassis")
API_ROOT = f"https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}"

REGISTER_DIR = "docs/agent_docs/docs026_concept_register/register"
INDEX_NAME = "000_concept_index.md"

# `### ABC-001 — name` — the concept entry heading. Same shape the register's own
# documented count command uses, so this check and the header agree by
# construction rather than by two regexes that can drift apart.
ENTRY_RE = re.compile(r"^### ([A-Z]{2,4}-[0-9]{3})\b(.*)$", re.M)
# `| ABC-001 | name | status | summary | file.md |` — the index table row.
ROW_RE = re.compile(r"^\| ([A-Z]{2,4}-[0-9]{3}) \|(.*)$", re.M)
# The bolded headline figure in the index header, e.g. `**1,756 index table rows**`.
HEADLINE_RE = re.compile(r"\*\*([\d,]+) index table rows\*\*")

GITHUB_TOKEN = ""


def env_required(name):
    val = os.environ.get(name, "")
    if not val:
        print(f"REFUSING TO RUN: {name} is not set.", file=sys.stderr)
        sys.exit(2)
    return val


def gh_get(url, accept="application/vnd.github+json"):
    req = urllib.request.Request(
        url,
        headers={
            "Authorization": f"token {GITHUB_TOKEN}",
            "Accept": accept,
            "User-Agent": "concept-register-drift-check",
        },
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        return resp.read()


def resolve_ref_sha(ref):
    try:
        data = json.loads(gh_get(f"{API_ROOT}/commits/{urllib.parse.quote(ref)}"))
    except urllib.error.HTTPError as e:
        print(f"REFUSING TO RUN: ref '{ref}' does not resolve on GitHub ({e}).",
              file=sys.stderr)
        sys.exit(2)
    return data["sha"]


def list_register_files(sha):
    """Every *.md directly under the register directory, at this ref."""
    data = json.loads(gh_get(f"{API_ROOT}/git/trees/{sha}?recursive=1"))
    if data.get("truncated"):
        # Loud, and it matters more here than in a citation sweep: a truncated
        # tree silently shrinks the ENTRY side of the comparison, which would
        # manufacture "row with no entry" findings out of nothing.
        print("REFUSING TO RUN: GitHub truncated the tree listing, so the file "
              "list is incomplete and every finding below would be an artefact "
              "of the truncation.", file=sys.stderr)
        sys.exit(2)
    out = []
    for item in data["tree"]:
        p = item["path"]
        if item["type"] != "blob" or not p.endswith(".md"):
            continue
        if os.path.dirname(p) != REGISTER_DIR:
            continue  # direct children only — .index_fragments/ is scratch
        out.append(p)
    return sorted(out)


def fetch_raw(path, ref):
    url = f"{API_ROOT}/contents/{urllib.parse.quote(path)}?ref={urllib.parse.quote(ref)}"
    return gh_get(url, accept="application/vnd.github.raw+json").decode(
        "utf-8", errors="replace"
    )


def parse_entries(files, texts):
    """id -> [(file, heading tail)]. A list, so duplicates are visible."""
    entries = {}
    for path in files:
        if os.path.basename(path) == INDEX_NAME:
            continue
        for cid, tail in ENTRY_RE.findall(texts[path]):
            entries.setdefault(cid, []).append((os.path.basename(path), tail.strip()))
    return entries


def parse_rows(index_text):
    rows = {}
    for cid, tail in ROW_RE.findall(index_text):
        rows.setdefault(cid, []).append(tail.strip())
    return rows


def analyse(files, texts):
    index_path = f"{REGISTER_DIR}/{INDEX_NAME}"
    if index_path not in texts:
        print(f"REFUSING TO RUN: {index_path} not found at this ref.", file=sys.stderr)
        sys.exit(2)

    entries = parse_entries(files, texts)
    rows = parse_rows(texts[index_path])

    entry_ids, row_ids = set(entries), set(rows)
    findings = {
        "entry_without_row": sorted(entry_ids - row_ids),
        "row_without_entry": sorted(row_ids - entry_ids),
        "duplicate_entry": sorted(c for c, v in entries.items() if len(v) > 1),
        "duplicate_row": sorted(c for c, v in rows.items() if len(v) > 1),
    }

    headline = None
    m = HEADLINE_RE.search(texts[index_path])
    if m:
        headline = int(m.group(1).replace(",", ""))

    return {
        "files": len(files),
        "entries": len(entry_ids),
        "rows": len(row_ids),
        "headline": headline,
        "headline_drift": headline is not None and headline != len(row_ids),
        "findings": findings,
        "entry_detail": {c: entries[c] for c in findings["entry_without_row"]},
        "dup_detail": {c: entries[c] for c in findings["duplicate_entry"]},
    }


def render_report(res, ref, sha):
    total = sum(len(v) for v in res["findings"].values())
    lines = [
        "**concept-register-drift-check** — does the register agree with itself?",
        f"Read {res['files']} file(s) under `{REGISTER_DIR}/` at `{ref}` "
        f"(commit `{sha[:12]}`): **{res['entries']} concept entries**, "
        f"**{res['rows']} index rows**.",
        "",
    ]

    if total == 0 and not res["headline_drift"]:
        lines.append(
            "**Clean.** Every entry has an index row, every row has an entry, no id "
            "is used twice, and the index headline matches the row count."
        )
        lines.append("")
        lines.append(
            "This is the state to hold: the register is what a session searches to "
            "find out whether a mechanism already exists, and an entry with no index "
            "row reads as *does not exist*."
        )
        return "\n".join(lines)

    lines.append(
        f"**{total} drift finding(s)**"
        + (" plus a headline mismatch." if res["headline_drift"] else ".")
        + " Nothing here is a claim that an entry is WRONG — this check only asks "
        "whether the register's two halves describe the same set of concepts."
    )
    lines.append("")

    ew = res["findings"]["entry_without_row"]
    if ew:
        lines.append(
            f"### {len(ew)} concept(s) with a register entry and NO index row"
        )
        lines.append(
            "These are invisible to a search of the index — a session looking for "
            "them concludes they do not exist, and builds a second one. Fix by "
            "adding the row; the summary has to be written by someone who "
            "understands the concept, which is why this check does not write it."
        )
        for cid in ew:
            where = ", ".join(f"`{f}`" for f, _ in res["entry_detail"][cid])
            name = res["entry_detail"][cid][0][1].lstrip("— ").strip()
            lines.append(f"- `{cid}` — {name[:120]} (in {where})")
        lines.append("")

    re_ = res["findings"]["row_without_entry"]
    if re_:
        lines.append(f"### {len(re_)} index row(s) with NO register entry")
        lines.append(
            "The index promises a concept the category files do not define — either "
            "an entry was removed without its row, or the row's id is a typo. This "
            "direction had never fired on real data up to 2026-08-04."
        )
        for cid in re_:
            lines.append(f"- `{cid}`")
        lines.append("")

    de = res["findings"]["duplicate_entry"]
    if de:
        lines.append(f"### {len(de)} concept id(s) used by more than one entry")
        for cid in de:
            where = ", ".join(f"`{f}`" for f, _ in res["dup_detail"][cid])
            lines.append(f"- `{cid}` — defined in {where}")
        lines.append("")

    dr = res["findings"]["duplicate_row"]
    if dr:
        lines.append(f"### {len(dr)} concept id(s) listed by more than one index row")
        for cid in dr:
            lines.append(f"- `{cid}`")
        lines.append("")

    if res["headline_drift"]:
        lines.append("### Index headline disagrees with the index itself")
        lines.append(
            f"The header says **{res['headline']:,} index table rows**; there are "
            f"**{res['rows']:,}**. The headline is what threads quote when they do "
            "not re-measure, so it goes stale outward — into commit messages and "
            "handoffs — long after the file itself is fixed."
        )
        lines.append("")

    lines.append(
        "The local equivalent, if you want to reproduce this by hand: the `comm` "
        "pair in the index header (entry ids vs row ids, both directions)."
    )
    return "\n".join(lines).rstrip()


def write_doc_note(body, host, password):
    tag = "crdbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'concept-register-drift', ${tag}${body}${tag}$, "
        "'[\"concept-register-drift\"]'::jsonb, 'concept-register-drift-check');"
    )
    sql_path = "/tmp/concept-register-drift-note.sql"
    with open(sql_path, "w") as f:
        f.write(sql)

    env = dict(os.environ)
    env["PGPASSWORD"] = password
    subprocess.run(
        [
            "psql", "-h", host, "-p", "5432",
            "-U", "clients_user", "-d", "clients_db",
            "-v", "ON_ERROR_STOP=1", "-f", sql_path,
        ],
        env=env, check=True,
    )


def run_check(ref):
    sha = resolve_ref_sha(ref)
    files = list_register_files(sha)
    if not files:
        print(f"REFUSING TO RUN: no .md files found under {REGISTER_DIR}/ at "
              f"'{ref}' — an empty read would report every index row as an "
              f"orphan.", file=sys.stderr)
        sys.exit(2)
    texts = {p: fetch_raw(p, ref) for p in files}
    return sha, analyse(files, texts)


if __name__ == "__main__":
    GITHUB_TOKEN = env_required("GITHUB_READ_TOKEN")
    REGISTER_REF = env_required("REGISTER_REF")
    PG_CLIENTS_HOST = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")
    CLIENTS_DB_PASSWORD = env_required("CLIENTS_DB_PASSWORD")

    sha, result = run_check(REGISTER_REF)
    report = render_report(result, REGISTER_REF, sha)
    print(report)
    write_doc_note(report, PG_CLIENTS_HOST, CLIENTS_DB_PASSWORD)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='concept-register-drift').")
