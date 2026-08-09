#!/usr/bin/env python3
"""gate_component_bytes.py — assert every adopted component holds the bytes the ORIGIN
serves, and repair the ones that do not.

WHY THIS IS MANDATORY AFTER A LOCKED ADOPTION, NOT OPTIONAL.

`adopt_verbatim.go` has exactly one possible byte source: firecrawl's `rawHtml`
(apply_adoption_plan_action.go:843-975, reading page["rawHtml"] at :873), or its
`research_results` echo. `rawHtml` is the serialised POST-JAVASCRIPT DOM, not what the
origin sent. It absolutises URLs, rewrites <meta charset> to http-equiv, collapses
whitespace, escapes & — and bakes in whatever the page's own scripts did on load.

That is not a theory. A sibling site adopted 27 pages this way and matched the served
bytes on ZERO of them, each inflated by a near-constant 8,900-9,060 bytes; three of its
guides reached production mutated before anyone noticed. On THIS site the divergence is
much smaller (measured: -7/+10/+95 bytes on three pages, no nav.js to bake in) but it is
still 3 of 3, so the gate fails everywhere and the repair is required everywhere.

The `sha256` in content_data is NOT a gate: it is computed over the stored rawHTML, so
it is self-consistent by construction and says nothing about fidelity to the origin.
A council seat objected on exactly that ground and was right.

Two traps this encodes:

  * `sha256(text::bytea)` is INVALID Postgres. Use sha256(convert_to(col,'UTF8')).
    Digesting in the database also avoids shipping 42 pages of HTML over kubectl.
  * Every automated write to page_components must carry the SHARED lock predicate
    (lock_helpers.go:42-46), not a hand-rolled `locked_at IS NULL`. A write that is
    lock-suppressed reports success while changing nothing, so RowsAffected is read
    and asserted, never assumed.

Run:  python3 gate_component_bytes.py [--domain <d>] [--source <dir>]   # report only
      python3 gate_component_bytes.py --domain loancash.co.uk --repair  # repair

--domain defaults to this workstream's site; --source defaults to the deploy repo dir
for that domain (override it only when the byte source is NOT the deploy repo, e.g. the
mortgagecalculator gemini/02 tree — and then get the domain INTO the repo first, per that
lane's handoff §1).
"""
import argparse
import base64
import hashlib
import os
import re
import subprocess
import sys

_ap = argparse.ArgumentParser()
_ap.add_argument("--domain", default="loanandmortgagecalculator.co.uk")
_ap.add_argument("--source", default=None,
                 help="byte-source dir (default: ~/projects/sites/<domain>)")
_ap.add_argument("--repair", action="store_true")
_args = _ap.parse_args()
DOMAIN = _args.domain
SITE = os.path.expanduser(_args.source or f"~/projects/sites/{DOMAIN}")
REPAIR = _args.repair
if not os.path.isdir(SITE):
    sys.exit(f"byte-source dir does not exist: {SITE}")

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]

# lock_helpers.go:42-46 — the single source of truth for "may automation write this
# row?". Expiry-aware: a timed lock that has passed is writable again.
LOCK_OK = ("(pc.locked_at IS NULL OR (pc.lock_type = 'timed' "
           "AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()))")


def psql(sql, tuples_only=True):
    cmd = PSQL + (["-t", "-A", "-F", "\t"] if tuples_only else [])
    r = subprocess.run(cmd + ["-c", sql], capture_output=True, text=True, timeout=180)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}")
    return r.stdout


def repo_path(url, name):
    """Mirror datahelpers.PageFilePathFromURL: root and directory URLs resolve to
    index.html, everything else is the path with its leading slash stripped."""
    p = (url or "").split("?")[0].split("#")[0]
    if not p or p == "/":
        return "index.html"
    p = p.lstrip("/")
    if p.endswith("/"):
        p += "index.html"
    if not p.endswith(".html"):
        p += ".html"
    return p


rows = psql(f"""
SELECT p.id, p.url, p.name, p.rebuild_policy,
       COALESCE(pc.id::text, ''),
       COALESCE(pc.slot_name, ''),
       COALESCE(encode(sha256(convert_to(pc.rendered_html, 'UTF8')), 'hex'), ''),
       COALESCE(length(pc.rendered_html)::text, ''),
       COALESCE(pc.content_data->>'deploy_mode', ''),
       CASE WHEN pc.id IS NULL THEN '' WHEN {LOCK_OK} THEN 'writable' ELSE 'LOCKED' END,
       (SELECT count(*) FROM page_components x WHERE x.page_id = p.id)
FROM pages p
LEFT JOIN page_components pc ON pc.page_id = p.id
WHERE p.site_id = (SELECT id FROM sites WHERE domain = '{DOMAIN}')
ORDER BY p.url;
""").strip()

if not rows:
    sys.exit(f"no pages for {DOMAIN} yet — has the adoption finished writing them?")

match = mismatch = missing = locked = assembled = 0
repairs, problems = [], []

for line in rows.splitlines():
    f = line.split("\t")
    page_id, url, name, policy, pc_id, slot, stored_sha, stored_len, mode, lock, n_comp = f
    rel = repo_path(url, name)
    disk = os.path.join(SITE, rel)

    if not pc_id:
        missing += 1
        problems.append(f"NO COMPONENT   {url}  (page {name})")
        continue
    # Only a single verbatim row mirrors the whole repo file — the same predicate
    # loadVerbatimPageHTML applies (rerender_single_page_action.go). A decomposed
    # page's SECTION rows can never byte-match the assembled file, and a --repair
    # that "fixed" them would overwrite each section with the entire document
    # (nearly did, 2026-08-08, on consolidation's prose rows — bugfix 224 session).
    if mode != "verbatim" or int(n_comp) != 1:
        assembled += 1
        problems.append(f"ASSEMBLED      {url}  components={n_comp} mode={mode or '-'} "
                        f"slot={slot} — section rows do not mirror the repo file; skipped")
        continue
    if not os.path.isfile(disk):
        problems.append(f"NO REPO FILE   {url}  -> expected {rel}")
        continue

    want = open(disk, "rb").read()
    want_sha = hashlib.sha256(want).hexdigest()

    if want_sha == stored_sha:
        match += 1
        continue

    mismatch += 1
    delta = int(stored_len or 0) - len(want)
    problems.append(f"BYTES DIFFER   {url}  repo {len(want)}B vs stored {stored_len}B "
                    f"({delta:+d})  mode={mode or '-'} components={n_comp}")
    if lock == "LOCKED":
        locked += 1
        problems.append(f"               ^ LOCKED — automation must not overwrite; "
                        f"needs a human unlock")
        continue
    repairs.append((pc_id, url, want, want_sha, rel))

print(f"\nbyte gate for {DOMAIN}\n")
print(f"  components byte-exact against the repo   {match}")
print(f"  components whose bytes DIFFER            {mismatch}")
print(f"  pages with no component row              {missing}")
print(f"  mismatches blocked by an active lock     {locked}")
print(f"  assembled/decomposed rows skipped        {assembled}")
print()
for p in problems:
    print("  " + p)

if not mismatch and not missing:
    print("\n  GATE PASSES — every component holds the bytes the origin serves.")
    sys.exit(0)

if not REPAIR:
    print(f"\n  GATE FAILS on {mismatch + missing}. Re-run with --repair to load the "
          f"repo bytes in.\n  Expected after a locked adoption: the byte source is a "
          f"post-JS crawl, not the origin.")
    sys.exit(1)

# ── repair ───────────────────────────────────────────────────────────────────────
print(f"\n  repairing {len(repairs)} component(s) from repo bytes…\n")
stmts = []
for pc_id, url, want, want_sha, rel in repairs:
    b64 = base64.b64encode(want).decode()
    stmts.append(
        "UPDATE page_components pc SET "
        f"rendered_html = convert_from(decode('{b64}', 'base64'), 'UTF8'), "
        f"content_data = jsonb_set("
        f"  jsonb_set(COALESCE(pc.content_data, '{{}}'::jsonb), '{{sha256}}', '\"{want_sha}\"'::jsonb),"
        f"  '{{byte_source}}', '\"deploy-repo\"'::jsonb), "
        "updated_at = NOW() "
        f"WHERE pc.id = '{pc_id}'::uuid AND {LOCK_OK};")

out = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1", "--echo-errors"],
                     input="BEGIN;\n" + "\n".join(stmts) + "\nCOMMIT;\n",
                     capture_output=True, text=True, timeout=600)
if out.returncode != 0:
    sys.exit(f"repair transaction FAILED (rolled back):\n{out.stderr.strip()}")

# Read RowsAffected, never assume it. A lock-suppressed UPDATE reports success and
# changes nothing, which is indistinguishable from a real write unless you count.
ok = len(re.findall(r"^UPDATE 1$", out.stdout, re.M))
zero = len(re.findall(r"^UPDATE 0$", out.stdout, re.M))
print(f"  UPDATE 1 (row written): {ok}")
print(f"  UPDATE 0 (suppressed):  {zero}")
if zero:
    sys.exit(f"\n  {zero} update(s) affected NO row — lock-suppressed. Not repaired.")
if ok != len(repairs):
    sys.exit(f"\n  expected {len(repairs)} writes, saw {ok}. Stopping.")
print(f"\n  {ok} component(s) repaired. Re-run without --repair to prove the gate now "
      f"passes —\n  the re-run is the evidence, not this message.")
