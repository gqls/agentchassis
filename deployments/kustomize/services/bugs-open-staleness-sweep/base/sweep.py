#!/usr/bin/env python3
"""bugs-open-staleness-sweep — RFC_005 §3.3, cheap first pass only.

Runs weekly (K8s CronJob). For every file under bugs_open/*.md at a pinned
git ref, extracts `path:line`/`path:line-range`/`path:Symbol` citations and
checks whether the cited path still exists at that ref and, where a line
number was cited, whether the file still has that many lines, or, where a
bare name was cited, whether that text still appears in the file at all.

This is deliberately the CHEAP pass: no LLM, no judgement about whether a
bug is actually fixed. It writes ONE doc_notes row per run (subject_type
'pipeline', subject_key 'bugs-open-staleness' — an existing subject_type,
so this needs no migration and no change to validDocSubjectTypes) listing
what it found. It NEVER edits a bugs_open/*.md file and NEVER closes a bug
— a citation going stale is a flag for a human/thread to interpret (moved,
refactored, or genuinely fixed elsewhere), not a verdict.

Stdlib only — no pip install, so a fresh postgres:16-alpine container needs
nothing but `apk add python3` to run this.

Ref discipline (matches 090_TRIGGER_needs_diagnosis_v1.sh's own rule, "never
silently fall back to a stale ref"): SWEEP_REF has no default here. The
CronJob manifest pins it explicitly and a human updates it deliberately when
the platform's live working branch changes — the same manual-bump discipline
IMAGE_TAG already uses. resolve_ref_sha() below fails loudly (the whole run
exits nonzero, nothing is written) if the ref does not resolve on GitHub,
rather than falling back to anything.
"""
import json
import os
import re
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

REPO_OWNER = os.environ.get("SWEEP_REPO_OWNER", "gqls")
REPO_NAME = os.environ.get("SWEEP_REPO_NAME", "agentchassis")
API_ROOT = f"https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}"

CITED_EXTS = r"(?:go|py|sh|sql|md|yaml|yml|js|ts|tsx|jsx)"
# `path:loc` — loc is a line number, a line range, or a bare symbol/text token.
CITATION_WITH_LOC_RE = re.compile(
    r"`([A-Za-z0-9_./-]+\." + CITED_EXTS + r"):([A-Za-z0-9_-]+)`"
)
# `path` alone, no colon — existence-only check.
CITATION_BARE_RE = re.compile(
    r"`([A-Za-z0-9_./-]+\." + CITED_EXTS + r")`"
)


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
            "User-Agent": "bugs-open-staleness-sweep",
        },
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        return resp.read()


def resolve_ref_sha(ref):
    # Fail loudly rather than fall back — see module docstring.
    try:
        data = json.loads(gh_get(f"{API_ROOT}/commits/{urllib.parse.quote(ref)}"))
    except urllib.error.HTTPError as e:
        print(f"REFUSING TO RUN: ref '{ref}' does not resolve on GitHub ({e}).",
              file=sys.stderr)
        sys.exit(2)
    return data["sha"]


def fetch_tree(sha):
    data = json.loads(gh_get(f"{API_ROOT}/git/trees/{sha}?recursive=1"))
    if data.get("truncated"):
        print("WARNING: GitHub truncated the tree listing — some paths may be "
              "missed this run.", file=sys.stderr)
    return [item["path"] for item in data["tree"] if item["type"] == "blob"]


def fetch_raw(path, ref):
    url = f"{API_ROOT}/contents/{urllib.parse.quote(path)}?ref={urllib.parse.quote(ref)}"
    return gh_get(url, accept="application/vnd.github.raw+json").decode(
        "utf-8", errors="replace"
    )


def extract_citations(text):
    seen = set()
    out = []
    for path, loc in CITATION_WITH_LOC_RE.findall(text):
        key = (path, loc)
        if key not in seen:
            seen.add(key)
            out.append((path, loc))
    for path in CITATION_BARE_RE.findall(text):
        key = (path, None)
        if key not in seen:
            seen.add(key)
            out.append((path, None))
    return out


def resolve_path(cited_path, tree_set, basename_index):
    """Returns (kind, resolved): kind is 'resolved' (resolved is the single
    real path, never None), 'ambiguous' (resolved is the list of candidate
    paths), or 'not-found' (resolved is None) — every caller can rely on
    'resolved' being a usable path if and only if kind == 'resolved'."""
    if "/" in cited_path:
        if cited_path in tree_set:
            return ("resolved", cited_path)
        return ("not-found", None)
    matches = basename_index.get(cited_path, [])
    if len(matches) == 1:
        return ("resolved", matches[0])
    if len(matches) > 1:
        return ("ambiguous", matches)
    return ("not-found", None)


def parse_loc(loc):
    m = re.match(r"^(\d+)(?:-(\d+))?$", loc)
    if m:
        end = int(m.group(2)) if m.group(2) else int(m.group(1))
        return ("line", end)
    return ("symbol", loc)


def run_sweep(ref):
    sha = resolve_ref_sha(ref)
    tree_paths = fetch_tree(sha)
    tree_set = set(tree_paths)
    basename_index = {}
    for p in tree_paths:
        basename_index.setdefault(p.rsplit("/", 1)[-1], []).append(p)

    bug_files = sorted(
        p for p in tree_paths if p.startswith("bugs_open/") and p.endswith(".md")
    )

    content_cache = {}
    flagged = []  # (bug_file, citation_display, reason, category)
    citations_checked = 0

    # Categories, so a reader's eye goes to genuine staleness first:
    #   'stale'      — the cited evidence itself is gone/changed (the real signal)
    #   'ambiguous'  — citation is a bare filename matching >1 path; a citation-
    #                  quality issue, not evidence that anything moved
    #   'fetch-error' — our own check couldn't complete; not a claim about the bug

    for bug_path in bug_files:
        try:
            text = fetch_raw(bug_path, ref)
        except urllib.error.HTTPError as e:
            flagged.append((bug_path, "(whole file)", f"could not fetch this bug file itself: {e}", "fetch-error"))
            continue
        time.sleep(0.1)

        for cited_path, loc in extract_citations(text):
            citations_checked += 1
            display = f"`{cited_path}:{loc}`" if loc else f"`{cited_path}`"
            kind, resolved = resolve_path(cited_path, tree_set, basename_index)

            if kind == "not-found":
                flagged.append((bug_path, display, "cited path no longer found in the tree", "stale"))
                continue
            if kind == "ambiguous":
                flagged.append((
                    bug_path, display,
                    f"bare filename matches {len(resolved)} paths at this ref — "
                    f"cannot verify without a full path ({', '.join(resolved[:5])})",
                    "ambiguous",
                ))
                continue

            real_path = resolved  # kind == "resolved" here — real_path is never None
            if loc is None:
                continue  # existence confirmed, nothing further to check

            if real_path not in content_cache:
                try:
                    content_cache[real_path] = fetch_raw(real_path, ref)
                except urllib.error.HTTPError as e:
                    content_cache[real_path] = None
                time.sleep(0.1)
            body = content_cache[real_path]

            if body is None:
                flagged.append((bug_path, display, f"{real_path} resolved in the tree but content fetch failed", "fetch-error"))
                continue

            loc_kind, val = parse_loc(loc)
            if loc_kind == "line":
                total_lines = body.count("\n") + 1
                if val > total_lines:
                    flagged.append((
                        bug_path, display,
                        f"cites line {val} but {real_path} now has only {total_lines} lines",
                        "stale",
                    ))
            else:
                if val not in body:
                    flagged.append((
                        bug_path, display,
                        f"`{val}` not found anywhere in {real_path} (heuristic text match — a "
                        f"rename or non-literal reference can also produce this)",
                        "stale",
                    ))

    return {
        "ref": ref,
        "sha": sha,
        "files_checked": len(bug_files),
        "citations_checked": citations_checked,
        "flagged": flagged,
    }


def render_report(result):
    lines = []
    lines.append(f"**bugs-open-staleness-sweep** — cheap pass, RFC_005 §3.3.")
    lines.append(
        f"Checked {result['files_checked']} file(s) under `bugs_open/`, "
        f"{result['citations_checked']} citation(s), at `{result['ref']}` "
        f"(commit `{result['sha']}`)."
    )
    lines.append("")
    if not result["flagged"]:
        lines.append("No citations flagged this run.")
        return "\n".join(lines)

    by_category = {"stale": {}, "ambiguous": {}, "fetch-error": {}}
    for bug_path, display, reason, category in result["flagged"]:
        by_category[category].setdefault(bug_path, []).append((display, reason))

    def render_group(by_file):
        out = []
        for bug_path in sorted(by_file):
            out.append(f"- `{bug_path}`")
            for display, reason in by_file[bug_path]:
                out.append(f"  - {display}: {reason}")
        return out

    stale_count = sum(len(v) for v in by_category["stale"].values())
    ambiguous_count = sum(len(v) for v in by_category["ambiguous"].values())
    error_count = sum(len(v) for v in by_category["fetch-error"].values())

    lines.append(
        f"**{len(result['flagged'])} citation(s) flagged** — "
        f"{stale_count} likely-stale, {ambiguous_count} ambiguous citations, "
        f"{error_count} check errors. A flag means the cited evidence has moved "
        f"or gone (or, for 'ambiguous', was never fully qualified) — it is NOT a "
        f"verdict that the bug is fixed; a human/thread still decides that."
    )
    lines.append("")

    if by_category["stale"]:
        lines.append(f"### Likely stale ({stale_count}) — the cited evidence itself has moved or gone")
        lines.extend(render_group(by_category["stale"]))
        lines.append("")
    if by_category["ambiguous"]:
        lines.append(
            f"### Ambiguous citations ({ambiguous_count}) — bare filename, not "
            f"necessarily stale, just not verifiable without a full path"
        )
        lines.extend(render_group(by_category["ambiguous"]))
        lines.append("")
    if by_category["fetch-error"]:
        lines.append(f"### Check errors ({error_count}) — this sweep's own fetch failed, not a claim about the bug")
        lines.extend(render_group(by_category["fetch-error"]))
        lines.append("")

    return "\n".join(lines).rstrip()


def write_doc_note(body):
    tag = "swbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'bugs-open-staleness', ${tag}${body}${tag}$, "
        "'[\"bugs-open-staleness\"]'::jsonb, 'bugs-open-staleness-sweep');"
    )
    sql_path = "/tmp/bugs-open-staleness-note.sql"
    with open(sql_path, "w") as f:
        f.write(sql)

    env = dict(os.environ)
    env["PGPASSWORD"] = CLIENTS_DB_PASSWORD
    subprocess.run(
        [
            "psql", "-h", PG_CLIENTS_HOST, "-p", "5432",
            "-U", "clients_user", "-d", "clients_db",
            "-v", "ON_ERROR_STOP=1", "-f", sql_path,
        ],
        env=env, check=True,
    )


if __name__ == "__main__":
    GITHUB_TOKEN = env_required("GITHUB_READ_TOKEN")
    SWEEP_REF = env_required("SWEEP_REF")
    PG_CLIENTS_HOST = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")
    CLIENTS_DB_PASSWORD = env_required("CLIENTS_DB_PASSWORD")

    result = run_sweep(SWEEP_REF)
    report = render_report(result)
    print(report)
    write_doc_note(report)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='bugs-open-staleness').")
