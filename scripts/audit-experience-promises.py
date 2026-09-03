#!/usr/bin/env python3
"""audit-experience-promises.py — two checks on promises a page makes and does not keep.

WHY THIS EXISTS
---------------
Round two of the owner's review of the first paid customer build (boxingonline.com,
2026-09-02) produced three more defects of one shape: **the page keeps the letter of its
build and breaks the promise it makes.** Everything validated clean; every work item that
was going to complete, completed. Two of the three reduce to a mechanical test, and those
two are here. The third does not, and is documented as refused rather than approximated —
see REFUSED below, because a check that quietly does 60% of a semantic job is worse than
no check, and this lane has already shipped one detector that had to be narrowed twice.

RULE A — TWO DOORS, ONE NAME
    Two active header entries carrying the SAME nav label and pointing at DIFFERENT pages.
    boxingonline's header reads: Home · News · Fight Calendar · About · Contact · News ·
    Get Started. `/articles/index.html` (blog-index, nav_order 2) and `/news/index.html`
    (news-index, nav_order 100) are both labelled "News". A visitor cannot know which is
    which, and no journey through that nav is well defined.
    Note the shape, because it is why a journey check is not the same as a nav check: the
    nav DECLARATION pipeline was proved end-to-end on 2026-08-31 and is working — it put
    the right five in the right order. It simply had nothing to say about a sixth entry
    duplicating one of them. A mechanism can be correct and the experience still broken.

RULE B — A TOOL PAGE WITH NOTHING THE READER CAN USE OR READ
    A `page_type='tool'` page that serves no interactive element, no inline data, and no
    runtime fetch. boxingonline's `/tools/fight-calendar/index.html` — the paid brief's
    core deliverable ("a boxing calendar showing upcoming fights … populated with real
    upcoming events") — serves 6,640 characters explaining HOW THE CALENDAR IS MAINTAINED
    ("How entries get added", "What each listing gives you", "We correct entries … rather
    than leaving a stale date up") above no calendar, no events, and no empty state. A
    reader cannot tell whether there are no fights or whether the page is broken.
    The owner has since ruled the general form (2026-09-02): "The research agent should
    have researched what's on and that is what should have appeared on this page", and on
    the fighter comparator, that it "should contain detailed, fact checked information that
    prefills the form". So "name the reader input and the site-supplied data separately; a
    tool whose site-supplied set is empty is a form, not a tool" is a ruling, not a
    proposal.

REFUSED — "does this page contain the thing its title asserts?"
    `/blog/last-nights-result-underdog-shocks-the-champion.html` contains no result (it is
    an essay on why underdogs win, citing 1990 and 2019); `/blog/saturday-fight-card-
    preview-whos-fighting-who.html` names no fighters. Meanwhile `/news/index.html` on the
    SAME SITE carries "Filip Hrgovic beats Moses Itauma by stoppage", 31 August — the story
    the article's own title promised. That is a real defect and it is NOT mechanical: it
    needs a judgement about whether a body delivers what a title asserts. Approximating it
    with proper-noun or date counting would fire on every well-written general piece and
    stay silent on a specific-sounding essay. Route it to a seat that can read, not to a
    regex. Recorded here so the next reader knows it was considered and declined.

    THE CONCRETE DISPROOF, so nobody re-derives it: that underdog essay names Buster
    Douglas, Mike Tyson, Andy Ruiz, Anthony Joshua, Tokyo and Madison Square Garden, and
    dates 1990 and 2019 — nine proper nouns and two dates in an article containing no
    news. A proper-noun-or-date rule would score it ABOVE a correct, short report of the
    actual result. The approximation does not merely miss the case; it inverts it.

SUBSTRATE: WHY RULE B READS STORED MARKUP AND NOT THE SERVED PAGE - SETTLED 2026-09-02
    This looked like an open risk (this estate has a documented stored-vs-served gap: a
    phantom-link check was caught reading stored html while the served page differed), so
    it was measured rather than assumed. The boxingonline session probed all five of that
    site's tool pages, cache-busted, three signals each, stored vs served: four agree, one
    differs - and THE ONE THAT DIFFERS IS THE ONLY TRUE POSITIVE.

    The served fight-calendar page has exactly one control, `<button class="mobile-menu-
    toggle">`, inside `<header>`; outside header and footer it has none. Every site carries
    that toggle - measured 2026-09-02, 30 of 30 sites whose header component mentions it
    store it as a `<button>` tag. `page_components.rendered_html` excludes chrome (header
    and footer live in `site_components`), so the stored substrate correctly sees zero
    controls on that page.

    So reading served bytes would give every page in the estate a control, Rule B would
    never fire, and the fleet result would be a clean zero - the exact shape this lane
    keeps writing landmines about. STORED IS THE RIGHT SUBSTRATE, because the question is
    "does the page BODY offer the reader anything", and chrome is not the page. If a future
    check does want served bytes, strip the header and footer elements first, or the chrome
    answers the question for you.

    While confirming that, I first measured "0 of 33 headers carry a control" and nearly
    wrote it up as a sharper mechanism. It was my own regex: BACKSLASH-b IS BACKSPACE IN
    POSTGRES, not a word boundary - '<button\b' matches nothing; backslash-y is the word
    boundary. Every rule in this file runs in Python, where it behaves, so the checks were
    never affected - but any census written in SQL with backslash-b silently under-matches
    and reads clean. See LANDMINES.md, 2026-09-02.

CONTROLS (printed on every run; a run without them is not evidence)
    demand   : rule B is only meaningful if tool pages CAN pass it — 126 of 320 carry an
               inline data array and 305 carry an interactive element (measured 2026-09-02),
               so the check can come out either way.
    escape   : vonc.com/tools/gauntlet/round.html has no inputs and no inline data and is
               NOT a finding, because it fetches its round from the live API. Only 6 of 320
               tool pages fetch at runtime, so this escape is narrow, not a blanket.
    separate : a tool page with NO rendered html at all is a different defect (never built,
               not mis-built) and is counted in its own bucket, never mixed into rule B.

⚠ THE MOTIVATING PAGES ARE BEING REPAIRED WHILE YOU READ THIS. boxingonline is pre-delivery
and the owner has ruled "build and fix everything before approval", so both of rule A's and
rule B's live cases will disappear — correctly. That is why every control here is a FIXTURE
in --self-test and no live page is named as one. The lane learned this on 2026-08-31 when a
positive control was repaired inside twelve minutes and the detector reported FAIL on
itself.

USAGE
    scripts/audit-experience-promises.py [--json] [--site DOMAIN] [--write-note]
    scripts/audit-experience-promises.py --self-test        # no cluster access needed
"""

import argparse
import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone

# Signals, all read off the artefact rather than off a taxonomy column. The lesson behind
# that choice is written up in LANDMINES.md: `pages.page_type` records how the PLANNER
# classified a page, and the fleet files guides as blog-posts, so a class question answered
# from a taxonomy column returns a clean zero on the very case it was written for. Here
# page_type only SELECTS the corpus (which pages claim to be tools); every judgement below
# is made on what the page actually serves.
RE_INTERACTIVE = re.compile(r"<(input|select|textarea|button)\b", re.I)
RE_DATA_ARRAY = re.compile(r"(const|let|var|window\.[A-Za-z_]+)\s*[A-Za-z_]*\s*=\s*\[\s*[\{\"']")
RE_RUNTIME_FETCH = re.compile(r"fetch\(|XMLHttpRequest|axios|\.json\b", re.I)


def norm_label(s):
    return re.sub(r"\s+", " ", (s or "")).strip().casefold()


def judge_nav(entries):
    """entries: [{url, nav_label, page_type, nav_order}] for ONE site's active header.

    A finding is one label carried by more than one DISTINCT url. Same label on the same
    url is a duplicate ROW, not a duplicate door — a data tidy-up, not an experience defect.
    """
    by_label = {}
    for e in entries:
        lab = norm_label(e.get("nav_label"))
        if not lab:
            continue
        by_label.setdefault(lab, {})[e.get("url")] = e
    return [{"label": e_first(v)["nav_label"], "destinations": sorted(v.keys()),
             "entries": [v[u] for u in sorted(v)]}
            for lab, v in sorted(by_label.items()) if len(v) > 1]


def e_first(d):
    return d[sorted(d.keys())[0]]


def judge_index(row):
    """An index page that lists NONE of the pages sitting in its own directory.

    This closes the blind spot the sibling listing-class check states in its own docstring:
    that check catches guides/tools appearing under an editorial promise, and CANNOT catch
    the reverse, because `/blog/` is where this estate files guides and no class inference
    survives that. Rule C needs no class inference at all. It asks one question the site
    answers itself: this index sits at /guides/, so does it show what is in /guides/?

    THE ESCAPE IS `pages_in_dir`, and it is doing the heavy lifting. On 2026-08-31 I looked
    at dartsonline's /guides/index.html listing twelve /blog/ items and dismissed it as
    correct, reasoning that this estate files guides under /blog/. That reasoning was
    sound and the conclusion was WRONG: /guides/ on that site holds NINE tool guides, all
    orphaned, none of them listed. The convention is real; it simply does not settle the
    question, and I never asked whether the directory had anything in it.

    Restricted to INDEX-ROLE pages for the other escape. Measured 2026-09-02 without that
    restriction, the rule fires on every tool page and individual guide carrying a "related
    content" block — a page at /tools/x/ listing six items elsewhere is doing its job. The
    unrestricted population is large; the restricted one is two.
    """
    urls = [u for u in (row.get("item_urls") or []) if u]
    d = row.get("dir") or ""
    own = [u for u in urls if u.startswith(d)]
    detail = {"n_items": len(urls), "listed_from_own_dir": len(own),
              "pages_in_dir": row.get("pages_in_dir") or 0}
    if not urls:
        return "empty", detail
    # ESCAPE — nothing of its own to list, so listing elsewhere is the right behaviour.
    if detail["pages_in_dir"] == 0:
        return "ok", detail
    if own:
        return "ok", detail
    return "ignores_own_section", detail


def judge_tool(page):
    """One tool page -> verdict.

    'not_built'    : nothing rendered at all — a different defect, kept in its own bucket
                     so it can never inflate rule B's count.
    'nothing_usable': rendered prose, but no control, no data, and nothing fetched. This is
                     the fight-calendar shape: a page ABOUT a tool, standing in for one.
    'ok'           : anything the reader can operate or read.
    """
    html = page.get("html") or ""
    if not html.strip():
        return "not_built", {}
    detail = {
        "chars": len(html),
        "interactive": bool(RE_INTERACTIVE.search(html)),
        "inline_data": bool(RE_DATA_ARRAY.search(html)),
        "runtime_fetch": bool(RE_RUNTIME_FETCH.search(html)),
    }
    if detail["interactive"] or detail["inline_data"]:
        return "ok", detail
    # ESCAPE — the data arrives at runtime. Narrow by measurement: 6 of 320 tool pages
    # fetch at all (2026-09-02), and the one it rescues is vonc's API-backed gauntlet.
    if detail["runtime_fetch"]:
        return "ok", detail
    return "nothing_usable", detail


NAV_SQL = r"""
SELECT COALESCE(jsonb_agg(jsonb_build_object(
         'domain', s.domain, 'url', p.url, 'nav_label', p.nav_label,
         'page_type', p.page_type, 'nav_order', p.nav_order)), '[]'::jsonb)
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.in_header AND p.status = 'active' AND COALESCE(p.nav_label,'') <> '' __SITE__;
"""

# ⚠ CHUNKED ON PURPOSE. The first cut asked for every tool page's rendered_html in one
# statement and the kubectl exec stream died mid-transfer ("unexpected EOF") — ~320 pages
# of markup is several MB through a pipe not built for it. The tempting fix was to push
# the three regexes into SQL so only booleans cross the wire; that was REFUSED, because
# then --self-test would be exercising a Python mirror of a rule that Postgres actually
# applies, and the two dialects disagree. The rules stay in one place and the transport
# gets fixed instead.
TOOLS_SQL = r"""
SELECT COALESCE(jsonb_agg(jsonb_build_object(
         'domain', domain, 'url', url, 'html', html,
         'repair_not_served', repair_not_served)), '[]'::jsonb) FROM (
  SELECT s.domain, p.url,
         (SELECT string_agg(pc.rendered_html, ' ') FROM page_components pc
           WHERE pc.page_id = p.id) AS html,
         -- WRITTEN-NOT-SHIPPED (2026-09-03): this rule reads STORED rendered_html, so a
         -- repair that has landed in the DB and has NOT been deployed makes it report
         -- clean while the visitor still gets the broken page. p.build_status is no help
         -- — it records that a deploy once happened, not that THESE components were in
         -- it. Measured on seotools: 7 pages rebuilt 09:34-09:54 behind a 00:08 deploy,
         -- all serving 0 controls while stored html carried them.
         -- Narrowed 2026-09-03 after ground-truthing: the bare
         -- max(updated_at) > deployed_at flagged 38 pages and 5 of 5 sampled from
         -- other sites were serving fine. Two false-positive classes, both benign:
         --   * build_status='needs_rebuild' — the page is HONESTLY labelled pending
         --     (gaps of 13-44 days), and the previously deployed copy serves fine;
         --   * sub-second gaps — the component write lands microseconds after the
         --     deploy in the same transaction (finetuning.uk: 0.047s).
         -- The real defect is a page CLAIMING to be current while carrying newer
         -- components (seotools serp-snippet-previewer: build_status='deployed',
         -- deployed_at 00:08:16, newest component 09:43:18, serving 0 controls).
         (p.build_status = 'deployed' AND p.deployed_at IS NOT NULL
          AND (SELECT max(pc2.updated_at) FROM page_components pc2
                WHERE pc2.page_id = p.id) > p.deployed_at + interval '1 minute'
         ) AS repair_not_served
  FROM pages p JOIN sites s ON s.id = p.site_id
  WHERE p.status = 'active' AND p.page_type = 'tool' __SITE__
  ORDER BY s.domain, p.url LIMIT __LIMIT__ OFFSET __OFFSET__) t;
"""

TOOLS_COUNT_SQL = r"""
SELECT count(*) FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status = 'active' AND p.page_type = 'tool' __SITE__;
"""

# Small on purpose. The `kubectl exec` stream truncated a 948 KB response mid-string, so
# the ceiling is the workstation transport, not the query. In the CronJob this script dials
# postgres directly and the chunking is merely harmless. Chosen by bisection, not by taste.
CHUNK = 8

# Rule C's corpus. `pages_in_dir` counts the ACTIVE pages living under the index's own
# directory, excluding the index itself — that count is the whole escape clause: an index
# whose directory is empty has nothing of its own to list and must not be reported.
INDEX_SQL = r"""
WITH inst AS MATERIALIZED (
  SELECT p.id, p.site_id, s.domain, p.url AS host, p.page_type,
         '/' || split_part(trim(leading '/' from p.url),'/',1) || '/' AS dir,
         COALESCE(pc.content_data->'articles', pc.content_data->'items') AS arr
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
  WHERE p.status = 'active'
    AND p.page_type IN ('section-index','blog-index','news-index','entity-directory')
    AND jsonb_typeof(COALESCE(pc.content_data->'articles', pc.content_data->'items')) = 'array'
    AND p.url ~ '^/[^/]+/' __SITE__
), sized AS MATERIALIZED (
  SELECT * FROM inst WHERE jsonb_array_length(arr) > 0
)
SELECT COALESCE(jsonb_agg(jsonb_build_object(
         'domain', domain, 'host', host, 'page_type', page_type, 'dir', dir,
         'item_urls', (SELECT jsonb_agg(e->>'url') FROM jsonb_array_elements(arr) e),
         'pages_in_dir', (SELECT count(*) FROM pages q
                           WHERE q.site_id = sized.site_id AND q.status = 'active'
                             AND q.url LIKE sized.dir || '%' AND q.url <> sized.host)
       )), '[]'::jsonb) FROM sized;
"""


def _psql_argv(sql):
    """kubectl exec from a workstation; a direct dial from inside the cluster."""
    host = os.environ.get("PG_CLIENTS_HOST")
    pw = os.environ.get("CLIENTS_DB_PASSWORD")
    if host and pw:
        return (["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
                 "-At", "-v", "ON_ERROR_STOP=1", "-c", sql],
                {**os.environ, "PGPASSWORD": pw})
    return (["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
             "psql", "-U", "clients_user", "-d", "clients_db", "-At",
             "-v", "ON_ERROR_STOP=1", "-c", sql], None)


def _q(sql):
    argv, env = _psql_argv(sql)
    out = subprocess.run(argv, env=env, capture_output=True, text=True, timeout=300)
    if out.returncode != 0:
        print(out.stderr.strip(), file=sys.stderr)
        sys.exit(2)
    return out.stdout.strip()


def fetch(site=None):
    clause = f"AND s.domain = '{site}'" if site else ""
    nav = json.loads(_q(NAV_SQL.replace("__SITE__", clause)) or "[]")
    total = int(_q(TOOLS_COUNT_SQL.replace("__SITE__", clause)) or "0")
    tools = []
    for off in range(0, total, CHUNK):
        page = _q(TOOLS_SQL.replace("__SITE__", clause)
                  .replace("__LIMIT__", str(CHUNK)).replace("__OFFSET__", str(off)))
        tools.extend(json.loads(page or "[]"))
    # A silent short read would understate every count below, so it is an error, not a note.
    if len(tools) != total:
        print(f"fetch incomplete: {len(tools)} of {total} tool pages returned", file=sys.stderr)
        sys.exit(2)
    indexes = json.loads(_q(INDEX_SQL.replace("__SITE__", clause)) or "[]")
    return {"nav": nav, "tools": tools, "indexes": indexes}


def analyse(data):
    nav_by_site = {}
    for e in data.get("nav", []):
        nav_by_site.setdefault(e["domain"], []).append(e)
    nav_findings = []
    for domain, entries in sorted(nav_by_site.items()):
        for f in judge_nav(entries):
            nav_findings.append({"domain": domain, **f})

    tool_findings, not_built, unserved, ok = [], [], [], 0
    demand = {"interactive": 0, "inline_data": 0, "runtime_fetch": 0}
    for page in data.get("tools", []):
        verdict, detail = judge_tool(page)
        for k in demand:
            if detail.get(k):
                demand[k] += 1
        if verdict == "nothing_usable":
            tool_findings.append({"domain": page["domain"], "url": page["url"], **detail})
        elif verdict == "not_built":
            not_built.append({"domain": page["domain"], "url": page["url"]})
        elif page.get("repair_not_served"):
            # It LOOKS usable in the database and may not be on the site. Refusing to
            # call this clean is the whole point: the window where a repair is written
            # and not shipped is exactly when a clean result is most misleading.
            unserved.append({"domain": page["domain"], "url": page["url"], **detail})
        else:
            ok += 1

    index_findings, index_ok = [], 0
    for row in data.get("indexes", []):
        verdict, detail = judge_index(row)
        if verdict == "ignores_own_section":
            index_findings.append({"domain": row["domain"], "host": row["host"],
                                   "page_type": row["page_type"], "dir": row["dir"], **detail})
        elif verdict == "ok":
            index_ok += 1

    n_tools = len(data.get("tools", []))
    return {
        "measured_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "nav_entries_scanned": len(data.get("nav", [])),
        "tool_pages_scanned": n_tools,
        "rule_a_two_doors_one_name": nav_findings,
        "rule_b_nothing_usable": tool_findings,
        "tool_pages_never_built": not_built,
        "tool_pages_repair_not_served": unserved,
        "tool_pages_ok": ok,
        "index_pages_scanned": len(data.get("indexes", [])),
        "rule_c_ignores_own_section": index_findings,
        # DEMAND CONTROL: rule B is only evidence if a tool page can pass it. If these go
        # to zero the regexes have drifted, and every "clean" tool page is unexamined.
        "demand_control": demand,
    }


def note_body(r):
    """Built apart from the write so a fixture can reach it — a self-test cannot vouch for
    a path it never calls, which this lane learned from a CronJob NameError on 2026-08-31."""
    lines = [f"experience promise check - {r['measured_at']}",
             f"{r['nav_entries_scanned']} header nav entries and {r['tool_pages_scanned']} "
             f"tool pages scanned.",
             f"Rule A (two doors, one name): {len(r['rule_a_two_doors_one_name'])}",
             f"Rule B (tool page with nothing usable): {len(r['rule_b_nothing_usable'])}",
             f"Rule C (an index listing none of its own directory): "
             f"{len(r.get('rule_c_ignores_own_section', []))}",
             f"Separately, tool pages never built (no rendered html): "
             f"{len(r['tool_pages_never_built'])}",
             f"Tool pages whose repair is written but NOT SERVED (stored html has a "
             f"control, newest component postdates deployed_at): "
             f"{len(r['tool_pages_repair_not_served'])} — NOT counted clean. TRIAGE ONLY: "
             f"ground-truthed 2026-09-03, 7 of 11 were genuinely unserved and 4 served fine; "
             f"no DB column separates them, so curl the page before acting.",
             ""]
    for f in r["rule_a_two_doors_one_name"]:
        lines.append(f"[A] {f['domain']}: nav label {f['label']!r} points at "
                     f"{len(f['destinations'])} different pages: "
                     f"{', '.join(f['destinations'])}")
    if r["rule_a_two_doors_one_name"]:
        lines.append("    a rule A finding names two doors; it does NOT say which to "
                     "close. Read what each page HOLDS before removing either — on the "
                     "case this rule was built from, the newer duplicate held the site's "
                     "only real dated sourced results and the older one held disposable "
                     "essays, so the nav symptom pointed at the wrong page.")
    for f in r["rule_b_nothing_usable"]:
        lines.append(f"[B] {f['domain']}{f['url']}: {f['chars']} chars rendered, no control, "
                     f"no inline data, no runtime fetch — a page about a tool, not a tool.")
    for f in r["tool_pages_repair_not_served"]:
        lines.append(f"[!] {f['domain']}{f['url']}: stored html looks usable but its newest "
                     f"component postdates the page's deployed_at — the repair may not be "
                     f"served. This rule reads STORED html; do not read it as clean.")
    for f in r["tool_pages_never_built"]:
        lines.append(f"[--] {f['domain']}{f['url']}: no rendered html at all (never built; "
                     f"a different defect from rule B, not mixed into it)")
    for f in r.get("rule_c_ignores_own_section", []):
        lines.append(f"[C] {f['domain']}{f['host']} ({f['page_type']}): {f['pages_in_dir']} "
                     f"active pages live in {f['dir']} and the listing shows NONE of them "
                     f"({f['n_items']} items, all from elsewhere).")
    if r.get("rule_c_ignores_own_section"):
        lines.append("    a completed page_rerender does NOT prove the listing was "
                     "re-resolved: some reasons degrade to assemble-only when the item "
                     "carries no component_id (platform/livespec/rerender_reasons.go:85), "
                     "re-shipping the stored array byte for byte and completing anyway. "
                     "Check the item's reason AND its component_id before reading a "
                     "persistent finding as a failed fix.")
    d = r["demand_control"]
    lines += ["",
              f"Demand control: of {r['tool_pages_scanned']} tool pages, "
              f"{d['interactive']} carry a control, {d['inline_data']} inline data, "
              f"{d['runtime_fetch']} fetch at runtime. If these reach zero the check has "
              f"gone blind and its clean result means nothing.",
              "Detector: scripts/audit-experience-promises.py (--self-test carries the "
              "fixtures). Origin: experience_loop, boxingonline owner review round two, "
              "2026-09-02."]
    return "\n".join(lines)


def write_doc_note(result):
    """ONE row per run, findings or not: a missing row means the job did not run, and must
    never read as 'nothing is wrong'."""
    body = note_body(result)
    sql = ("INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
           "VALUES ('pipeline', 'experience-promise', $epc$" + body + "$epc$, "
           "'[\"experience-promise\"]'::jsonb, 'experience-promise-check');")
    argv, env = _psql_argv(sql)
    subprocess.run(argv, env=env, check=True, capture_output=True, text=True, timeout=120)


def run(args):
    r = analyse(fetch(args.site))
    if args.write_note:
        write_doc_note(r)
    if args.json:
        print(json.dumps(r, indent=2))
        return
    print(f"experience promise check — {r['measured_at']}")
    print(f"  {r['nav_entries_scanned']} header nav entries · "
          f"{r['tool_pages_scanned']} tool pages ({r['tool_pages_ok']} fine)\n")
    print(f"  RULE A — two doors, one name: {len(r['rule_a_two_doors_one_name'])}")
    for f in r["rule_a_two_doors_one_name"]:
        print(f"    {f['domain']}: {f['label']!r} ->")
        for e in f["entries"]:
            print(f"        {e['url']}  ({e['page_type']}, nav_order {e['nav_order']})")
    if r["tool_pages_repair_not_served"]:
        print(f"\n  ⚠ REPAIR WRITTEN BUT MAYBE NOT SERVED "
              f"({len(r['tool_pages_repair_not_served'])}) — NOT counted clean; this rule "
              f"reads STORED html and these pages' newest component postdates their deploy.\n"
              f"    TRIAGE LIST, NOT A FINDINGS LIST: ground-truthed 2026-09-03, 7 of 11 were "
              f"genuinely serving 0 controls; the other 4 served fine (deployed_at simply was\n"
              f"    not updated). No DB column separates the two — CURL THE PAGE before acting, "
              f"and never 'fix' one of these on the strength of this line alone:")
        for f in r["tool_pages_repair_not_served"]:
            print(f"    {f['domain']}{f['url']}")
    print(f"\n  RULE B — a tool page with nothing usable: {len(r['rule_b_nothing_usable'])}")
    for f in r["rule_b_nothing_usable"]:
        print(f"    {f['domain']}{f['url']}  — {f['chars']} chars, no control, "
              f"no inline data, no fetch")
    print(f"\n  RULE C — an index listing none of its own directory: "
          f"{len(r.get('rule_c_ignores_own_section', []))}"
          f"  (of {r.get('index_pages_scanned', 0)} index pages)")
    for f in r.get("rule_c_ignores_own_section", []):
        print(f"    {f['domain']}{f['host']}  ({f['page_type']}) — {f['pages_in_dir']} pages "
              f"in {f['dir']}, {f['listed_from_own_dir']} of {f['n_items']} items from it")
    if r["tool_pages_never_built"]:
        print(f"\n  SEPARATELY — tool pages with no rendered html "
              f"({len(r['tool_pages_never_built'])}); never built, a different defect:")
        for f in r["tool_pages_never_built"]:
            print(f"    {f['domain']}{f['url']}")
    d = r["demand_control"]
    print(f"\n  DEMAND CONTROL — of {r['tool_pages_scanned']} tool pages: "
          f"{d['interactive']} interactive, {d['inline_data']} inline data, "
          f"{d['runtime_fetch']} runtime fetch.")
    if r["tool_pages_scanned"] and not any(d.values()):
        print("    FAIL — every signal is zero; the check is blind and its clean "
              "result means nothing.")


def self_test():
    failures = []

    # RULE A — the real header, as measured 2026-09-02.
    boxing = [
        {"url": "/index.html", "nav_label": "Home", "page_type": "landing", "nav_order": 1},
        {"url": "/articles/index.html", "nav_label": "News", "page_type": "blog-index", "nav_order": 2},
        {"url": "/tools/fight-calendar/index.html", "nav_label": "Fight Calendar", "page_type": "tool", "nav_order": 3},
        {"url": "/about.html", "nav_label": "About", "page_type": "content", "nav_order": 4},
        {"url": "/contact.html", "nav_label": "Contact", "page_type": "content", "nav_order": 5},
        {"url": "/news/index.html", "nav_label": "News", "page_type": "news-index", "nav_order": 100},
    ]
    got = judge_nav(boxing)
    if len(got) != 1 or got[0]["destinations"] != ["/articles/index.html", "/news/index.html"]:
        failures.append(f"    rule A missed the two-News header: {got}")
    else:
        print("  PASS  rule A finds two 'News' doors on the measured boxingonline header")

    if judge_nav([e for e in boxing if e["url"] != "/news/index.html"]):
        failures.append("    rule A fires on a header with no duplicate label")
    else:
        print("  PASS  rule A silent on the same header without the duplicate")

    dupe_row = [{"url": "/news/index.html", "nav_label": "News", "page_type": "news-index", "nav_order": 2},
                {"url": "/news/index.html", "nav_label": "News", "page_type": "news-index", "nav_order": 100}]
    if judge_nav(dupe_row):
        failures.append("    rule A fires on a duplicate ROW (same url) — a tidy-up, not a defect")
    else:
        print("  PASS  rule A silent on a duplicate row pointing at the SAME page")

    if not judge_nav([{"url": "/a.html", "nav_label": " news ", "page_type": "x", "nav_order": 1},
                      {"url": "/b.html", "nav_label": "News", "page_type": "x", "nav_order": 2}]):
        failures.append("    rule A missed a case/whitespace variant of the same label")
    else:
        print("  PASS  rule A normalises case and whitespace before comparing")

    # RULE B — each fixture is a real page shape measured on 2026-09-02.
    cases = [
        ("fight-calendar: prose about a calendar, no calendar (the motivating case)",
         {"html": "<h2>How we build the fight calendar</h2><p>The calendar above pulls "
                  "together the fights worth building your weekend around.</p>"
                  "<h3>What each listing gives you</h3><p>Date and start time.</p>"},
         "nothing_usable"),
        ("vonc gauntlet round: no control and no inline data, but it FETCHES (escape)",
         {"html": "<div id='round'></div><script>fetch('/api/v1/tools/gauntlet/round')"
                  ".then(r=>r.json())</script>"},
         "ok"),
        ("fighter comparator: 18 inputs — a form, but the reader can operate it",
         {"html": "<form><input name='a'><input name='b'></form>"},
         "ok"),
        ("a tool shipping its own data",
         {"html": "<script>const FIGHTERS = [{\"name\":\"Usyk\"}];</script>"},
         "ok"),
        ("a tool page with nothing rendered is NOT rule B — it was never built",
         {"html": ""},
         "not_built"),
        ("whitespace-only html is still 'never built', not a finding",
         {"html": "   \n  "},
         "not_built"),
    ]
    for name, page, expected in cases:
        got, _ = judge_tool(page)
        if got != expected:
            failures.append(f"    {name}: expected {expected}, got {got}")
        else:
            print(f"  PASS  {name}")

    # RULE C — both fixtures are pages measured on 2026-09-02.
    c_cases = [
        ("boxingonline /guides/index.html: 4 guides in /guides/, listing shows 0 of them",
         {"dir": "/guides/", "pages_in_dir": 4,
          "item_urls": ["/blog/a.html", "/blog/b.html", "/blog/c.html"]},
         "ignores_own_section"),
        ("dartsonline /guides/index.html: 9 orphaned tool guides in /guides/, 0 listed "
         "(the case I wrongly cleared on 2026-08-31)",
         {"dir": "/guides/", "pages_in_dir": 9,
          "item_urls": ["/blog/x.html"] * 12},
         "ignores_own_section"),
        ("ESCAPE: an index whose directory is empty has nothing of its own to list",
         {"dir": "/news/", "pages_in_dir": 0, "item_urls": ["/blog/a.html"] * 20},
         "ok"),
        ("an index listing its own section is fine",
         {"dir": "/guides/", "pages_in_dir": 14,
          "item_urls": ["/guides/a.html", "/guides/b.html"]},
         "ok"),
        ("a MIXED index (some own, some elsewhere) is fine — the rule fires only on zero",
         {"dir": "/guides/", "pages_in_dir": 6,
          "item_urls": ["/guides/a.html"] + ["/blog/b.html"] * 6},
         "ok"),
    ]
    for name, row, expected in c_cases:
        got, _ = judge_index(row)
        if got != expected:
            failures.append(f"    {name}: expected {expected}, got {got}")
        else:
            print(f"  PASS  {name}")

    # The note path, which the cluster had to find for this lane once already.
    sample = analyse({"nav": boxing and [{**e, "domain": "boxingonline.com"} for e in boxing],
                      "tools": [{"domain": "boxingonline.com",
                                 "url": "/tools/fight-calendar/index.html",
                                 "html": "<p>How we build the fight calendar</p>"}],
                      "indexes": [{"domain": "boxingonline.com", "host": "/guides/index.html",
                                   "page_type": "section-index", "dir": "/guides/",
                                   "pages_in_dir": 4, "item_urls": ["/blog/a.html"]}]})
    try:
        body = note_body(sample)
    except Exception as exc:                       # noqa: BLE001 — report, never raise
        failures.append(f"    note_body raised {exc!r}")
    else:
        if "$epc$" in body:
            failures.append("    note body contains the dollar-quote tag it is wrapped in")
        elif "[A]" not in body or "[B]" not in body or "[C]" not in body:
            failures.append("    note body drops a rule's findings")
        else:
            print("  PASS  doc_notes body builds and carries both rules")
    try:
        note_body(analyse({"nav": [], "tools": []}))
    except Exception as exc:                       # noqa: BLE001
        failures.append(f"    note_body raised on a CLEAN run: {exc!r}")
    else:
        print("  PASS  doc_notes body builds for a clean run too")

    # PARITY with the CronJob's copy. Two copies of a detector is the drift class this
    # estate keeps paying for, so the copy is pinned rather than trusted.
    here = os.path.dirname(os.path.abspath(__file__))
    twin = os.path.join(here, "..", "deployments", "kustomize", "services",
                        "experience-promise-check", "base", "check.py")
    if os.path.exists(twin):
        with open(os.path.abspath(__file__), "rb") as a, open(twin, "rb") as b:
            if a.read() != b.read():
                failures.append(
                    "    PARITY: scripts/audit-experience-promises.py and "
                    "deployments/.../experience-promise-check/base/check.py DIFFER. The "
                    "cluster runs the copy — `cp` one over the other, then re-apply the "
                    "overlay or the ConfigMap keeps the old script.")
            else:
                print("  PASS  cronjob copy is byte-identical to this script")
    else:
        print("  SKIP  cronjob copy not present (running outside the repo)")

    if failures:
        print("\nSELF-TEST FAILED:")
        print("\n".join(failures))
        sys.exit(1)
    print("\nself-test: all cases pass")


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--site", help="restrict to one domain")
    ap.add_argument("--self-test", action="store_true", help="fixtures only, no cluster")
    ap.add_argument("--write-note", action="store_true",
                    help="write one doc_notes row for this run (clean runs included)")
    args = ap.parse_args()
    if args.self_test:
        self_test()
        return
    run(args)


if __name__ == "__main__":
    main()
