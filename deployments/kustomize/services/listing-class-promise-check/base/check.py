#!/usr/bin/env python3
"""audit-listing-class-promise.py — does a listing's CONTENT CLASS match the class its
heading promises?

WHY THIS EXISTS
---------------
The first paid customer build (boxingonline.com, BR-9AUZ59, 2026-08-31) shipped with the
single most prominent editorial block on its home page — headed **"Latest from the ring"**,
subtitled *"Fresh news, previews and results…"* — populated by **four explainer guides about
the site's own tools**. Every page validated `valid=true, issues=0`. No artefact was
malformed: the guides were well-formed guides, the listing was a well-formed listing. The
PROMISE was broken and nothing in the fleet could see it, because every check we own looks
at one artefact in isolation.

That defect was routed to the experience_loop lane
(`docs024_key_docs_latest/experience_loop/CONTRIB_2026-08-31_…four_experience_defects…md`)
with a suggested cheap first cut:

    "a listing whose heading/subtitle says news/articles/latest must not be populated by
     pages whose `page_type`/role is `guide` or `tool`."

**THAT RULE IS REFUTED, AND THIS SCRIPT CARRIES THE REFUTATION AS AN ARM YOU CAN RE-RUN.**
`page_type` is not a content-class signal on this estate. Measured 2026-08-31:

    pages that are guide-shaped by url/name/title:  blog-post 246 across 30 sites
                                                    guide      72 across  9 sites

Guides-as-`blog-post` is the FLEET CONVENTION. Boxingonline's four guides were typed
`blog-post` at build time (retyped to `guide` by the webdesign lane at 17:30 that day, after
the owner complained), and leopardessconsulting's eight are `blog-post` today. So the
proposed rule returns **zero on both real cases** — including the one that motivated it. It
is the estate's most expensive shape: a check that could not have come out otherwise
(`a-post-fix-zero-needs-a-demand-control`).

WHAT THIS READS INSTEAD
-----------------------
The item's OWN DISPLAYED IDENTITY — the url the reader clicks, the title they read, the
component's own item name. That is the same evidence the reader has, and it does not depend
on a taxonomy column that three lanes populate differently.

THE ESCAPE CLAUSE, AND WHY IT IS NOT OPTIONAL
---------------------------------------------
`no_horizontal_overflow`'s fix (`5042d5ecb`) is the worked example every check on this lane
inherits: a check without an escape clause reports CORRECT pages as broken for ever. Here
the escape is that **a heading naming the other class is honest**. robot-hands.com's
"Gripper Technologies & Integration **Guides**" lists guides — that is a kept promise, and
it must PASS. So a finding requires the promise to name an editorial class and NO other.

CONTROLS (this script prints all of them; a run without them is not evidence)
----------------------------------------------------------------------------
  positive : boxingonline.com /index.html  and  leopardessconsulting.co.uk /blog.html
             must both be FOUND. If either goes quiet, the classifier has drifted (or the
             site was fixed — check which before believing a clean run).
  negative : robot-hands.com /learning-center-hub.html must NOT be found (escape clause).
  naive arm: the same corpus under the refuted page_type rule. Expected: 0 on the historic
             cases. If it ever rises, the estate's typing convention has changed and this
             header's measurement needs re-dating.

USAGE
    scripts/audit-listing-class-promise.py [--json] [--site DOMAIN] [--all]
    scripts/audit-listing-class-promise.py --self-test      # no cluster access needed

Exit 0 always (advisory). `--json` prints one object; findings are in `.findings`.
"""

import argparse
import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone

# --------------------------------------------------------------------------------------
# Vocabularies. GROUNDED, not invented: the url segments below are the ones that actually
# occur as the first path segment of a listing item fleet-wide (measured 2026-08-31:
# tools 557, guides 132, blog 67, mortgages 54, games 10, archetypes 8, glosario 8,
# insights 4, guias 4, plus 278 absolute https: links). Words are matched case-insensitively
# on word boundaries so "newsletter" does not read as "news".
# --------------------------------------------------------------------------------------

PROMISE_LEXICON = {
    # English, plus the Spanish/Portuguese terms that actually occur in live headings
    # (noticias/notícias, guías/guias, calculadora) — 9 of 159 listing instances are not
    # in English. Any other language is a STATED BLIND SPOT, not a clean result.
    "editorial": r"\b(latest|news|articles?|blog|stor(y|ies)|updates?|recent|coverage|editorial|insights?|opinions?|briefings?|dispatch(es)?|analysis|notici(a|as)|not\u00edcias?|\u00faltimas)\b",
    "guide": r"\b(guides?|how[- ]to|explainers?|tutorials?|walkthroughs?|advice|tips|learn|gu\u00ed?as?|gu\u00edas)\b",
    "tool": r"\b(tools?|calculators?|estimators?|checkers?|quiz(zes)?|converters?|finders?|simulators?|planners?|calculadora)\b",
}

ITEM_URL_CLASS = {
    "guides": "guide",
    "guias": "guide",
    "tools": "tool",
    "blog": "editorial",
    "news": "editorial",
    "articles": "editorial",
    "insights": "editorial",
}

# Item NAME/TITLE tells, applied only when the url segment does not decide it.
RE_NAME_GUIDE = re.compile(r"(^|-)guide(-|$)|guide$", re.I)
RE_NAME_TOOL = re.compile(r"^tool[-_]", re.I)
RE_TITLE_GUIDE = re.compile(r"^\s*understanding\s|\|\s*guide\s*$|\bguide\b\s*$", re.I)

ITEM_ARRAY_KEYS = ("articles", "items")  # measured: the only link-bearing listing keys
# (`nav_items`, 61 instances, is navigation chrome, not a listing promise — out of scope)


def promise_text(inst):
    """The promise as the READER meets it, and where we read it from.

    content_data wins when it carries a heading; otherwise the rendered markup does. This
    fallback is load-bearing, not a nicety: measured 2026-08-31, only 19 of 159 listing
    instances put their heading in content_data, so a content_data-only reading reported
    88% of the corpus as "promises nothing" — a clean result that could not have come out
    otherwise.
    """
    if (inst.get("heading") or "").strip():
        return inst.get("heading"), inst.get("subtitle") or "", "content_data"
    if (inst.get("html_heading") or "").strip():
        return inst.get("html_heading"), inst.get("html_subtitle") or "", "rendered_html"
    return "", "", "none"


def classify_promise(heading: str, subtitle: str, site_name: str = ""):
    """Which content classes does this heading NAME? Returns (classes, suppressed).

    SUBJECT-VOCABULARY GUARD. A class word that is part of the SITE'S OWN NAME is subject
    matter, not a content-class promise. garden-tools.uk heads a block "Caring for your
    garden tools" and lists care guides: nothing is broken, "tools" is what the site sells.
    Without this guard that reads as a tool promise kept by guides, for ever.

    Suppressions are RETURNED, not swallowed — the caller prints them, so a reader can see
    what the guard silenced and disagree with it.
    """
    text = f"{heading or ''} {subtitle or ''}"
    hay = re.sub(r"[^a-z]", " ", (site_name or "").lower())
    classes, suppressed = [], []
    for k, pat in PROMISE_LEXICON.items():
        m = re.search(pat, text, re.I)
        if not m:
            continue
        word = m.group(0).lower()
        if word in hay or word.rstrip("s") in hay:
            suppressed.append({"class": k, "word": word, "site_name": site_name})
        else:
            classes.append(k)
    return sorted(classes), suppressed


def classify_item(url: str, name: str, title: str):
    """The class this item DECLARES ITSELF to be, from what the reader can see.

    ⚠ ASYMMETRIC BY DESIGN, and this is the whole of the check's honesty.

    `guide` and `tool` are SELF-DECLARED on this estate: the page's own title says
    "Understanding X | Guide", its item name is `tool-x-guide`, its url sits under
    `/guides/` or `/tools/`. Those are assertions the artefact makes about itself.

    "Editorial" is NOT self-declared and this check does not claim it. A `/blog/` url
    proves nothing here, because THIS ESTATE FILES GUIDES UNDER `/blog/` — measured
    2026-08-31, 246 guide-shaped pages across 30 sites are typed `blog-post`, against 72
    typed `guide` across 9. The first cut of this detector did assert it, and reported
    dartsonline's "All guides" and agritec's "Technical explainers" as broken for listing
    `/blog/` items that are, on the fleet's own convention, the guides they promise.

    CONSEQUENCE, STATED RATHER THAN HIDDEN: this check catches guides/tools showing up
    under an editorial promise (the boxingonline and leopardess shape). It CANNOT catch
    articles showing up under a guide promise. Closing that half needs a class signal the
    artefact does not currently carry, not a cleverer regex.
    """
    url = (url or "").strip()
    if re.match(r"^https?://", url, re.I):
        return "external"
    seg = url.lstrip("/").split("/")[0].lower()
    if seg in ("guides", "guias"):
        return "guide"
    if seg == "tools":
        return "tool"
    if name and RE_NAME_GUIDE.search(name):
        return "guide"
    if title and RE_TITLE_GUIDE.search(title):
        return "guide"
    if name and RE_NAME_TOOL.search(name):
        return "tool"
    return "unspecified"


def judge(heading, subtitle, items, site_name=""):
    """One listing instance -> (verdict, detail).

    The rule, stated generally: a listing may show any class its HEADING names, and no
    class it does not. So a "Guides for every part of the farm" block full of calculators
    is the same finding as a "Latest from the ring" block full of guides — one rule.

    ⚠ THE SUBTITLE IS NOT PART OF THE PROMISE, and that is a measured decision, not
    fastidiousness. The first cut read heading+subtitle. Of its 8 fleet findings, **4 came
    from the subtitle naming a tool or a guide IN PASSING** — homegarden's section prose
    ends "…if you want a shorter list, the Garden Jobs Finder"; lampenkap's is an item
    excerpt reading "the companion guide sets out the method". Worse, for the 139 listings
    whose promise lives only in rendered markup, "the first <p>" IS AN ITEM'S OWN EXCERPT.
    I had built a promise-reader that reads the items and calls it the promise. The
    heading is the block's label, it is short, and it is what the reader takes as the
    promise — so it alone carries it.

    Verdict: 'mismatch' | 'ok' | 'no_promise' (the heading names no class, so there is
    nothing to break — "Start here", "Run the numbers before you decide", and every
    heading in a language this lexicon does not read).
    """
    promised, suppressed = classify_promise(heading, "", site_name)
    classes = [classify_item(i.get("url"), i.get("name"), i.get("title")) for i in items]
    counted = [c for c in classes if c in ("guide", "tool")]
    detail = {
        "promised_classes": promised,
        "item_classes": {c: classes.count(c) for c in sorted(set(classes))},
        "n_items": len(items),
        "n_self_declared": len(counted),
        "suppressed_as_site_subject": suppressed,
    }
    # ESCAPE 1 — a heading naming no class promises nothing to break. Most of the corpus.
    if not promised:
        return "no_promise", detail
    # ESCAPE 2 — a heading that names the class it lists is KEEPING its promise, however
    # many classes that is (robot-hands' "…& Integration Guides"). Without this clause the
    # check reports correct pages as broken for ever — the lesson `no_horizontal_overflow`
    # paid for in 5042d5ecb.
    off = [c for c in counted if c not in promised]
    if not off:
        return "ok", detail
    detail["off_class"] = len(off)
    detail["ratio"] = round(len(off) / len(items), 3) if items else 0
    detail["severity"] = "high" if detail["ratio"] >= 0.5 else "advisory"
    return "mismatch", detail


# --------------------------------------------------------------------------------------
# Cluster access. Same idiom as the other audit scripts: kubectl exec into the client pod.
# --------------------------------------------------------------------------------------

SQL = r"""
WITH inst AS MATERIALIZED (
  SELECT pc.id, p.site_id, s.domain, COALESCE(s.company_name,'') AS company,
         p.url AS host_page, cc.name AS comp,
         COALESCE(pc.content_data->>'section_title', pc.content_data->>'title',
                  pc.content_data->>'heading', '')                       AS heading,
         COALESCE(pc.content_data->>'section_subtitle',
                  pc.content_data->>'subtitle', '')                      AS subtitle,
         -- 140 of 159 listings fleet-wide carry their heading ONLY in the rendered
         -- markup (measured 2026-08-31). Reading content_data alone made 88% of the
         -- corpus look as though it promised nothing.
         COALESCE((regexp_match(pc.rendered_html,
                   '<h[1-3][^>]*>\s*(.{0,120}?)\s*</h[1-3]>', 'i'))[1], '')  AS html_heading,
         COALESCE((regexp_match(pc.rendered_html,
                   '<p[^>]*>\s*(.{0,200}?)\s*</p>', 'i'))[1], '')            AS html_subtitle,
         COALESCE(pc.content_data->'articles', pc.content_data->'items') AS arr
  FROM page_components pc
  JOIN pages p  ON p.id = pc.page_id
  JOIN sites s  ON s.id = p.site_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE jsonb_typeof(COALESCE(pc.content_data->'articles', pc.content_data->'items')) = 'array'
    __SITE__
), sized AS MATERIALIZED (
  SELECT * FROM inst WHERE jsonb_array_length(arr) > 0
)
SELECT jsonb_agg(jsonb_build_object(
         'domain', domain, 'company', company,
         'host_page', host_page, 'component', comp,
         'heading', heading, 'subtitle', subtitle,
         'html_heading', html_heading, 'html_subtitle', html_subtitle,
         'items', (SELECT jsonb_agg(jsonb_build_object(
                            'url',   e->>'url',
                            'name',  e->>'name',
                            'title', COALESCE(e->>'title', e->>'name'),
                            'page_type', (SELECT tp.page_type FROM pages tp
                                           WHERE tp.site_id = sized.site_id
                                             AND tp.url = e->>'url' LIMIT 1)))
                     FROM jsonb_array_elements(arr) e WHERE jsonb_typeof(e) = 'object')
       )) FROM sized;
"""


def _psql_argv(sql):
    """Two ways in, chosen by environment — the SAME query either way.

    A session on the workstation reaches the database through `kubectl exec`; a CronJob
    inside the cluster has no pods/exec RBAC and dials postgres directly. Doing both here
    is what lets one file be the thing a session runs by hand AND the thing the clock runs
    (the sibling checks keep a second copy under `deployments/` and pin the pair with a
    parity test — the drift risk is real, and this avoids it by having no second copy).
    """
    host = os.environ.get("PG_CLIENTS_HOST")
    pw = os.environ.get("CLIENTS_DB_PASSWORD")
    if host and pw:
        return (["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
                 "-At", "-v", "ON_ERROR_STOP=1", "-c", sql],
                {**os.environ, "PGPASSWORD": pw})
    return (["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
             "psql", "-U", "clients_user", "-d", "clients_db", "-At",
             "-v", "ON_ERROR_STOP=1", "-c", sql], None)


def fetch(site=None):
    sql = SQL.replace("__SITE__", f"AND s.domain = '{site}'" if site else "")
    argv, env = _psql_argv(sql)
    out = subprocess.run(argv, env=env, capture_output=True, text=True, timeout=300)
    if out.returncode != 0:
        print(out.stderr.strip(), file=sys.stderr)
        sys.exit(2)
    body = out.stdout.strip()
    return json.loads(body) if body and body != "" else []


def naive_page_type_arm(instances):
    """The REFUTED rule, kept runnable so its zero stays visible rather than assumed."""
    hits = []
    for inst in instances:
        h, sub, _ = promise_text(inst)
        promised, _ = classify_promise(h, "",
                                       f"{inst.get('domain','')} {inst.get('company','')}")
        if not promised:
            continue
        types = [(i.get("page_type") or "") for i in (inst.get("items") or [])]
        if any(t in ("guide", "tool") and t not in promised for t in types):
            hits.append(f"{inst['domain']}{inst['host_page']}")
    return hits


def run(args):
    instances = fetch(args.site)
    findings, ok, na, no_promise, suppressed = [], 0, 0, 0, []
    sources = {"content_data": 0, "rendered_html": 0, "none": 0}
    for inst in instances:
        items = inst.get("items") or []
        h, sub, src = promise_text(inst)
        sources[src] += 1
        if src == "none":
            no_promise += 1
            continue
        verdict, detail = judge(h, sub, items,
                                f"{inst.get('domain','')} {inst.get('company','')}")
        detail["promise_source"] = src
        for sup in detail.get("suppressed_as_site_subject", []):
            suppressed.append({"domain": inst["domain"], "host_page": inst["host_page"],
                               "heading": h, **sup})
        if verdict == "mismatch":
            findings.append({
                "domain": inst["domain"], "host_page": inst["host_page"],
                "component": inst["component"], "heading": h,
                **detail,
                "off_class_items": [
                    {"url": i.get("url"), "title": i.get("title"),
                     "class": classify_item(i.get("url"), i.get("name"), i.get("title")),
                     "page_type": i.get("page_type")}
                    for i in items
                    if classify_item(i.get("url"), i.get("name"), i.get("title"))
                    in tuple(c for c in ("guide", "tool")
                             if c not in detail["promised_classes"])],
            })
        elif verdict == "ok":
            ok += 1
        else:
            na += 1  # heading named no class

    found = {f"{f['domain']}{f['host_page']}" for f in findings}

    # A control whose own case is outside --site's corpus cannot be evaluated, and must
    # say so rather than resolve to a boolean (2026-09-03). Before this, every scoped run
    # against another domain printed "positive_leopardess: FAIL — classifier drifted",
    # which is not a failure and not a finding: --site had simply filtered the control's
    # page away. It trained the reader to ignore the CONTROLS block, which is the only
    # thing telling them whether a zero is trustworthy. The negative control needs the
    # same treatment for the opposite reason — out of scope it passes VACUOUSLY, since a
    # page that was never fetched can hardly be reported.
    def _scoped_out(domain):
        return "n/a (control case not in --site scope)" if (
            args.site and args.site != domain) else None

    _leo = _scoped_out("leopardessconsulting.co.uk")
    _rh = _scoped_out("robot-hands.com")
    controls = {
        # LIVE positive control. It has a shelf life: see HISTORIC_FIXED below.
        "positive_leopardess":
            _leo if _leo else "leopardessconsulting.co.uk/blog.html" in found,
        # ESCAPE-CLAUSE control: a heading that names guides must never be reported.
        "negative_robot_hands_escape":
            _rh if _rh else "robot-hands.com/learning-center-hub.html" not in found,
        # The refuted rule, re-run every time so its zero is a measurement, not a memory.
        "naive_page_type_rule_hits": naive_page_type_arm(instances),
    }
    result = {
        "measured_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "instances_scanned": len(instances),
        "promise_read_from": sources,
        "verdicts": {"mismatch": len(findings), "ok": ok,
                     "no_class_promised": na, "no_promise_text_found": no_promise},
        "suppressed_as_site_subject": suppressed,
        "controls": controls,
        "findings": sorted(findings, key=lambda f: -f.get("ratio", 0)),
    }
    if args.write_note:
        write_doc_note(result)
    if args.json:
        print(json.dumps(result, indent=2))
        return
    print(f"listing-class promise audit — {result['measured_at']}")
    print(f"  {len(instances)} listing instances scanned "
          f"({ok} kept the promise, {na} promised no class, "
          f"{no_promise} carried no promise text, {len(findings)} MISMATCH)")
    print(f"  promise read from: {sources}\n")
    for f in result["findings"]:
        print(f"  [{f['severity'].upper()}] {f['domain']}{f['host_page']}  ({f['component']})")
        print(f"      heading   : {f['heading']!r}  [{f['promise_source']}]"
              f" -> promises {f['promised_classes']}")
        print(f"      items     : {f['item_classes']}  "
              f"({f['off_class']}/{f['n_items']} off-class, ratio {f['ratio']})")
        for i in f["off_class_items"][:4]:
            print(f"        - {i['class']:8s} {i['url']}  (page_type={i['page_type']})")
        if len(f["off_class_items"]) > 4:
            print(f"        … {len(f['off_class_items']) - 4} more")
        print()
    if suppressed:
        print(f"  SUPPRESSED by the subject-vocabulary guard ({len(suppressed)}) — "
              "a class word that is part of the site's own name:")
        for x in suppressed:
            print(f"    {x['domain']}{x['host_page']}  {x['heading']!r}  "
                  f"({x['word']!r} reads as {x['class']}, but the site is named for it)")
        print()
    print("  CONTROLS")
    for k, v in controls.items():
        if k == "naive_page_type_rule_hits":
            print(f"    refuted page_type rule would find: {len(v)} {v}")
        elif isinstance(v, str):
            print(f"    {k}: {v}")
        else:
            print(f"    {k}: {'PASS' if v else 'FAIL — classifier drifted or site changed'}")
    print("  " + HISTORIC_FIXED)


HISTORIC_FIXED = (
    "note: the motivating case (boxingonline.com/index.html, 'Latest from the ring' over four\n"
    "        tool guides — read at 16:45Z on 2026-08-31) was REPAIRED by another lane's\n"
    "        rerender while this detector was being written: gone by the first fleet scan at\n"
    "        16:57:20Z, row updated_at 16:57:33Z. That is why it is a FIXTURE in --self-test\n"
    "        and not a live control. A positive control naming a live page has a shelf life\n"
    "        measured in minutes here."
)


def note_body(result):
    """The doc_notes body, built separately from the write so the SELF-TEST CAN REACH IT.

    It is split for a reason worth keeping: the first cut of this file lost `write_doc_note`
    to a bad patch and `--self-test` still reported all-pass, because no fixture ever
    called it. The CronJob found it in the cluster, with a NameError, on its first run. A
    self-test cannot vouch for a path it never calls.
    """
    f = result["findings"]
    lines = [f"listing-class promise check - {result['measured_at']}",
             f"{result['instances_scanned']} listing instances scanned; "
             f"{len(f)} promise mismatch(es).",
             ""]
    if f:
        for x in f:
            lines.append(f"[{x['severity']}] {x['domain']}{x['host_page']} "
                         f"({x['component']}): heading {x['heading']!r} promises "
                         f"{x['promised_classes']}, but {x['off_class']} of {x['n_items']} "
                         f"items are self-declared "
                         f"{sorted(set(i['class'] for i in x['off_class_items']))}.")
    else:
        lines.append("No mismatch. A listing may show any class its heading names.")
    lines += ["",
              "What this means: the block's heading promises the reader one kind of thing "
              "and the items are another kind. Neither artefact is malformed, which is why "
              "nothing else sees it.",
              "Blind spot, stated: guides/tools under an editorial promise are catchable; "
              "articles under a guide promise are NOT, because this estate files guides "
              "under /blog/ and types them blog-post.",
              "Detector: scripts/audit-listing-class-promise.py (--self-test proves the "
              "escape clauses). Origin: experience_loop CONTRIB 2026-08-31, boxingonline."]
    return "\n".join(lines)


def write_doc_note(result):
    """ONE row per run, on findings AND on a clean result.

    The convention its sibling checks established, and the reason for it: a MISSING row
    means the job did not run, and must never read as "nothing is wrong". A check whose
    silence is ambiguous is a check nobody can rely on.
    """
    body = note_body(result)
    sql = ("INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
           "VALUES ('pipeline', 'listing-class-promise', $lcp$" + body + "$lcp$, "
           "'[\"listing-class-promise\"]'::jsonb, 'listing-class-promise-check');")
    argv, env = _psql_argv(sql)
    subprocess.run(argv, env=env, check=True, capture_output=True, text=True, timeout=120)



# --------------------------------------------------------------------------------------
# Self-test. Fixtures, no cluster. Every case is a real shape taken from the live estate.
# --------------------------------------------------------------------------------------

def self_test():
    cases = [
        # (name, heading, subtitle, items, expected verdict)
        ("boxingonline home slot (the motivating case)",
         "Latest from the ring", "Fresh news, previews and results from around the boxing world",
         [{"url": "/guides/tool-boxing-trivia-quiz-guide.html",
           "name": "tool-boxing-trivia-quiz-guide",
           "title": "Understanding Boxing Quiz — Test Your Knowledge | Guide"}] * 4,
         "mismatch"),
        ("leopardess blog listing (7 of 13 off-class)",
         "Latest Articles", "",
         [{"url": "/guides/tool-gdpr-ai-risk-assessment-guide.html", "name": "g", "title": "Understanding GDPR"}] * 7 +
         [{"url": "/blog/can-you-trust-ai-with-your-data.html", "name": "b", "title": "Can You Trust AI"}] * 6,
         "mismatch"),
        ("robot-hands — heading NAMES guides, so listing them is honest (ESCAPE)",
         "Gripper Technologies & Integration Guides",
         "Structured technical content: deep-dives, application engineering guides by industry",
         [{"url": "/guides/x.html", "name": "x", "title": "X"}] * 6,
         "ok"),
        ("dartsonline 'Start here' — promises no class at all",
         "Start here", "",
         [{"url": "/guides/x.html", "name": "x", "title": "X"}] * 3,
         "no_promise"),
        ("an honest news listing",
         "Latest news", "What happened this week",
         [{"url": "/blog/a.html", "name": "a", "title": "A"},
          {"url": "/news/b.html", "name": "b", "title": "B"}],
         "ok"),
        ("STATED BLIND SPOT: a guides index listing /blog/ items cannot be judged, "
         "because this estate files guides under /blog/",
         "All guides", "",
         [{"url": "/blog/barrel-shapes.html", "name": "barrel-shapes", "title": "Barrel shapes"}] * 12,
         "ok"),
        ("a heading in a language the lexicon does not read promises nothing "
         "(it must not guess a class)",
         "Rekeninstrumenten voor de fotometrie van uw ruimte", "",
         [{"url": "/tools/kleurtemperatuur-gids/index.html", "name": "t", "title": "T"}] * 6,
         "no_promise"),
        ("the subtitle naming a tool in passing is NOT a promise "
         "(4 of the first cut's 8 findings came from this)",
         "Garden and home jobs for January",
         "If you want a shorter list for your own plot, the Garden Jobs Finder will do it",
         [{"url": "/guides/tool-garden-jobs-finder-guide.html", "name": "g", "title": "G"}] * 3,
         "no_promise"),
        ("a single guide among real articles is advisory, not high",
         "Latest Articles", "",
         [{"url": "/blog/a.html", "name": "a", "title": "A"}] * 9 +
         [{"url": "/guides/g.html", "name": "g", "title": "G"}],
         "mismatch"),
        ("a GUIDE heading listing tools is the same defect (rule is class-general)",
         "Guides for every part of the farm", "",
         [{"url": "/tools/gas-unit-converter/index.html", "name": "tool-gas", "title": "Gas converter"}] * 3 +
         [{"url": "/guides/soil.html", "name": "soil", "title": "Soil"}],
         "mismatch"),
        ("SUBJECT-VOCABULARY GUARD: garden-tools.uk heading 'Caring for your garden "
         "tools' is about its product, not a content class",
         "Caring for your garden tools", "",
         [{"url": "/guides/tool-brand-comparator-guide.html", "name": "g", "title": "G"}] * 4,
         "no_promise"),
        ("external links are not evidence of an off-class listing",
         "Latest news", "",
         [{"url": "https://elsewhere.example/x", "name": "", "title": "X"}] * 3,
         "ok"),
    ]
    failures = []
    for name, h, s, items, expected in cases:
        site = "garden-tools.uk" if "garden tools" in (h or "").lower() else ""
        got, detail = judge(h, s, items, site)
        if got != expected:
            failures.append(f"    {name}: expected {expected}, got {got}  {detail}")
        else:
            extra = ""
            if got == "mismatch":
                extra = f"  (ratio {detail['ratio']}, {detail['severity']})"
            print(f"  PASS  {name}{extra}")

    # The refuted arm, proven refuted on fixtures rather than asserted in prose:
    # both real cases carried page_type='blog-post' on every off-class item.
    hist = [{"domain": "boxingonline.com", "host_page": "/index.html",
             "heading": "Latest from the ring", "subtitle": "Fresh news",
             "items": [{"url": "/guides/tool-x-guide.html", "name": "tool-x-guide",
                        "title": "Understanding X | Guide", "page_type": "blog-post"}] * 4}]
    naive = naive_page_type_arm(hist)
    if naive:
        failures.append(f"    naive arm should find NOTHING on the historic shape, found {naive}")
    else:
        print("  PASS  refuted page_type rule finds 0 on the historic shape (as measured)")

    # Severity discrimination must actually discriminate.
    _, hi = judge("Latest Articles", "", [{"url": "/guides/g.html", "name": "g", "title": "G"}] * 4)
    _, lo = judge("Latest Articles", "",
                  [{"url": "/blog/a.html", "name": "a", "title": "A"}] * 9 +
                  [{"url": "/guides/g.html", "name": "g", "title": "G"}])
    if not (hi.get("severity") == "high" and lo.get("severity") == "advisory"):
        failures.append(f"    severity does not discriminate: {hi.get('severity')} / {lo.get('severity')}")
    else:
        print("  PASS  severity discriminates (4/4 high, 1/10 advisory)")

    # The note path, which the cluster had to find for me once already.
    sample = {"measured_at": "T", "instances_scanned": 159, "findings": [
        {"severity": "high", "domain": "d", "host_page": "/p", "component": "c",
         "heading": "Latest Articles", "promised_classes": ["editorial"],
         "off_class": 7, "n_items": 13,
         "off_class_items": [{"class": "guide"}]}]}
    for probe in (sample, {"measured_at": "T", "instances_scanned": 159, "findings": []}):
        try:
            b = note_body(probe)
        except Exception as exc:                       # noqa: BLE001 - report, do not raise
            failures.append(f"    note_body raised {exc!r}")
        else:
            if "$lcp$" in b:
                failures.append("    note body contains the dollar-quote tag it is wrapped in")
    if not any("note_body" in f for f in failures):
        print("  PASS  doc_notes body builds for a finding run AND a clean run")

    # PARITY. base/check.py is a byte copy of this file, mounted into the CronJob as a
    # ConfigMap. Two copies of a detector is exactly the drift class this estate keeps
    # paying for, so the copy is PINNED here rather than trusted: edit one, this fails.
    here = os.path.dirname(os.path.abspath(__file__))
    twin = os.path.join(here, "..", "deployments", "kustomize", "services",
                        "listing-class-promise-check", "base", "check.py")
    if os.path.exists(twin):
        with open(os.path.abspath(__file__), "rb") as a, open(twin, "rb") as b:
            if a.read() != b.read():
                failures.append(
                    "    PARITY: scripts/audit-listing-class-promise.py and "
                    "deployments/.../listing-class-promise-check/base/check.py DIFFER. "
                    "The cluster runs the copy. `cp` one over the other, then re-apply "
                    "the overlay or the ConfigMap keeps the old script.")
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
