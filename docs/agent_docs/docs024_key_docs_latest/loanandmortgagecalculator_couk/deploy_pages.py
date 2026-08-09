#!/usr/bin/env python3
"""deploy_pages.py — file assemble-only rerenders for decomposed pages, wait,
then prove the served bytes equal the offline prediction.

ONE PAGE AT A TIME IS THE POINT OF THE PREDICTION FILE; this just does the
same thing for a batch without losing any of the checks. Per page:

  1. file a `page_rerender` work item — status 'triaged' (a 'detected' row is
     NEVER dispatched; nothing promotes it), page_id in the SPEC and the
     COLUMN (this item type does not resolve by page_name), and NO
     `spec.reason`, which is what selects ASSEMBLE mode. Assemble mode never
     enters `save_page_sections`, so bugs_open/189's locked-section
     duplication cannot bite.
  2. poll to a terminal status.
  3. wait out the CDN, then fetch with a browser UA and REFUSE to grade a
     response that is not a page: inside its own deploy window B2 serves a
     ~7-line {"error":…"NoSuchKey"} blob at HTTP 200, against which every
     grep returns 0 — which reads as a clean pass. Byte floor + DOCTYPE.
  4. diff against <work>/predicted/<name>.html, byte for byte.

`complete` is the work item's status, not the CDN's. Both canaries landed
~50s after filing and were correct on the wire ~90s later.

Usage:
  DECOMP_WORK=<dir> python3 deploy_pages.py --tag <run-tag> <name> [...]
  DECOMP_WORK=<dir> python3 deploy_pages.py --tag <run-tag> --all-applied
    (--all-applied deploys every page whose DB rows are already decomposed
     and whose prediction exists — i.e. exactly what load_lmc.py has written)
"""
import json
import os
import subprocess
import sys
import time
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
SITE_ID = "ed633ada-f8af-424b-b4d4-8af79160dbcd"
DOMAIN = "loanandmortgagecalculator.co.uk"
POLL_SECONDS = 15
POLL_MAX = 60
CDN_SETTLE = 100
MIN_PAGE_BYTES = 5000

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]


def psql(sql, sep="\t"):
    r = subprocess.run(PSQL + ["-tA", "-F", sep, "-v", "ON_ERROR_STOP=1", "-c", sql],
                       capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:600])
    return r.stdout.strip()


def fetch(url):
    req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
    with urllib.request.urlopen(req, timeout=30) as resp:
        return resp.read().decode("utf-8", "replace")


def main():
    work = os.environ.get("DECOMP_WORK")
    if not work:
        sys.exit("set DECOMP_WORK")
    if "--tag" not in sys.argv:
        sys.exit("--tag <run-tag> is required (it namespaces item_key)")
    tag = sys.argv[sys.argv.index("--tag") + 1]
    manifest = json.load(open(os.path.join(work, "manifest_voiced.json"),
                              encoding="utf-8"))["pages"]

    names = [a for a in sys.argv[1:] if not a.startswith("--") and a != tag]

    # url -> (id, name); pages.name is adoption's identity, not the manifest slug
    pages = {}
    for line in psql("SELECT url, id, name FROM pages WHERE site_id='%s';" % SITE_ID
                     ).splitlines():
        p = line.split("\t")
        if len(p) >= 3:
            pages[p[0]] = (p[1], p[2])

    if "--all-applied" in sys.argv:
        # decomposed == no row carries deploy_mode 'verbatim' any more
        decomposed = set(psql(
            "SELECT p.url FROM pages p WHERE p.site_id='%s' AND NOT EXISTS ("
            "SELECT 1 FROM page_components pc WHERE pc.page_id=p.id "
            "AND pc.content_data->>'deploy_mode'='verbatim');" % SITE_ID).splitlines())
        names = sorted(n for n, pg in manifest.items()
                       if pg["url"] in decomposed
                       and os.path.exists(os.path.join(work, "predicted", n + ".html")))
        print("--all-applied: %d decomposed page(s) with a prediction" % len(names))
    if not names:
        sys.exit("name at least one page, or pass --all-applied")

    filed = []
    for name in names:
        page = manifest[name]
        pid, dbname = pages[page["url"]]
        pred = os.path.join(work, "predicted", name + ".html")
        if not os.path.exists(pred):
            print("SKIP    %s: no prediction file (run load_lmc.py first)" % name)
            continue
        spec = json.dumps({
            "domain": DOMAIN,
            "page_id": pid,
            "page_name": dbname,
            "filename": page["url"].lstrip("/"),
        })
        item_key = "page_rerender_%s_%s_%s" % (dbname, SITE_ID[:8], tag)
        wid = psql(
            "INSERT INTO site_work_items (site_id, source, item_type, severity, "
            "summary, spec, priority, handler_agent, status, created_by, page_id, "
            "item_key) VALUES ('%s', 'lmc_decompose_voice', 'page_rerender', "
            "'medium', 'Assemble-only rerender after decomposition + voice pass', "
            "$sp$%s$sp$::jsonb, 90, 'page-rerender', 'triaged', "
            "'claude-session-lmc-voice-20260805', '%s', '%s') RETURNING id;"
            % (SITE_ID, spec, pid, item_key))
        # psql -tA can print the INSERT command tag after the RETURNING row;
        # keeping both lines poisoned the poll query's IN-list (2026-08-08).
        wid = wid.splitlines()[0].strip()
        print("filed   %-40s %s" % (name, wid))
        filed.append((name, wid, page["url"], pred))

    print("\npolling %d item(s)…" % len(filed))
    pending = {w: n for n, w, _u, _p in filed}
    for _ in range(POLL_MAX):
        if not pending:
            break
        rows = psql("SELECT id, status FROM site_work_items WHERE id IN (%s);"
                    % ",".join("'%s'" % w for w in pending))
        for line in rows.splitlines():
            wid, status = line.split("\t")
            if status in ("complete", "failed", "rejected", "cancelled"):
                print("  %-10s %s" % (status, pending.pop(wid, "?")))
        if pending:
            time.sleep(POLL_SECONDS)
    if pending:
        print("STILL PENDING after %ds: %s"
              % (POLL_SECONDS * POLL_MAX, ", ".join(pending.values())))

    print("\nwaiting %ds for the CDN (complete is the item's status, not B2's)…"
          % CDN_SETTLE)
    time.sleep(CDN_SETTLE)

    bad = 0
    for name, _wid, url, pred in filed:
        try:
            served = fetch("https://%s%s" % (DOMAIN, url))
        except Exception as e:  # noqa: BLE001
            print("FETCH-FAIL %-38s %s" % (name, e))
            bad += 1
            continue
        if len(served) < MIN_PAGE_BYTES or not served.lstrip().startswith("<!DOCTYPE"):
            print("NOT-A-PAGE %-38s %d bytes — B2 deploy-window blob? re-check later"
                  % (name, len(served)))
            bad += 1
            continue
        want = open(pred, encoding="utf-8").read()
        if served == want:
            print("IDENTICAL  %-38s %d bytes" % (name, len(served)))
        else:
            print("DIFFERS    %-38s served %d vs predicted %d"
                  % (name, len(served), len(want)))
            bad += 1

    print("\n%d of %d page(s) not proven" % (bad, len(filed)))
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
