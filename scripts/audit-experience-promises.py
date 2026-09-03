#!/usr/bin/env python3
"""audit-experience-promises.py — checks on promises a page makes and does not keep (rules A–D).

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

RULE D — A COLLECTION PAGE WITH NOTHING IN IT (added 2026-09-03)
    A page that promises a list of things and lists none. designblog.co.uk served four
    the day after its build: /glossary.html (0 terms), /inspiration/ (0 showcases),
    /the-design-feed/ (0 items), /uk-studios-directory/ (0 studios) — each carrying
    prose about what the page WILL hold ("What this glossary is for", "How the entries
    are written"). Owner's words: "the glossary has text about the brief and is not a
    glossary… the directory is empty." Rule C could not see any of them: it asks
    whether an index lists things from OUTSIDE its own directory, which presupposes it
    lists something, and both its corpus filters (an `articles`/`items` key that is a
    NON-EMPTY array) drop an empty index before the rule runs. Second independent
    instance of the "a page ABOUT X standing in for X" shape rule B was built from.

    WHAT COUNTS AS A COLLECTION PAGE (the corpus) — three signals, any one suffices:
      * the planner typed it as one: blog-index, news-index, entity-directory, or a
        bespoke directory type (model-directory, mortgage-lenders, health-insurers,
        protocol-tracker, adoption-tracker) — the same four index roles the plan-time
        gate in listing_item_sources.go (bugs_open/444) calls listing-family;
      * its own name, its own url segment, or its title with the SITE NAME REMOVED
        names a collection — COLLECTION_NOUNS below, derived 2026-09-03 from the
        section directories this estate actually plans. /glossary.html is
        page_type='content' on all 7 sites that have one, so page_type alone cannot
        see the case that funded this rule. ⚠ Both narrowings here are false positives
        this rule actually produced and that ground truth caught: matching the whole
        URL PATH made every page under /guides/ a guides index (idea.uk's guide at
        /guides/feedback-loops/), and matching the whole TITLE imported the site name —
        garden-tools.uk titles pages "Contact Us — Garden Tools UK", so "Tools" put its
        contact page and its affiliate disclosure in the corpus. Splitting the title on
        "|" does not fix it: that site separates with an em-dash, and em-dashes appear
        INSIDE legitimate titles ("Farm Insurance Glossary — Insurance Terms in Plain
        Farming Language"). strip_site_name() drops a trailing segment only when its
        letters match the domain's, so the separator does not have to be guessed;
      * its directory holds other active pages (pages_in_dir > 0): there IS something
        to list, whatever the page is called.
    A bare `section-index` carrying none of the three is NOT a collection page.
    Measured 2026-09-03, that shape is homegarden.uk's twelve month pages ("April —
    Garden and Home Jobs for This Month"), two contact pages and an about page —
    articles misfiled as indexes. Counted as `not_a_collection`, never reported.

    WHAT COUNTS AS LISTING SOMETHING (the ladder; first match wins):
      1. a body component carries a non-empty array AND renders at least one item tag
         (a link, h3/h4, article, li, tr or dt). The AND is measured, not cautious:
         dartsonline's /brands/ holds nav_items=4 in content_data and renders 315
         bytes containing nothing — served body h3=0, article=0.
      2. the body fills at runtime (`data-runtime-fill`, `fetch(`) — vonc's provocations
         archive, which also carries an honest empty state.
      2b. NEITHER LISTS NOR EMPTY — the markup repeats an item class while the data holds
         nothing: its own bucket, `render_data_divergence`, never a rule D finding and
         never counted clean. Found by ground-truthing, on the one page of 19 whose
         served body disagreed with this rule. farmerinsurance.uk/guides/index.html
         renders guide cards — titles, descriptions, a "Read guide" label — while
         `content_data.items` is `[]` and carries an `empty_state_text` ("More guides are
         being added") that is never shown because the markup predates it. Measured
         2026-09-03: stored 4 cards, served 3, and **0 anchors in either** — every card
         is a `<span>`, so a reader sees four guides and can click none. Calling it a
         rule D finding would be false (it renders items); calling it clean would be
         false too. As of 2026-09-03 this is 1 page of 126 candidates, and the two
         classes separate cleanly — every other candidate with an all-empty array set
         repeats an item class 0 or 1 times, never 2+.
      3. the body links to at least one page in its OWN directory (its own url excluded)
         — the ported indexes (webdesign /learn/ lists 31 that way, in markup only) and
         oufe's /cases/ (one case, one link).
      4. the body carries >=5 <dt> or >=10 <h3>/<h4>: a list written as prose headings.
         Every brief-echo page measured carries 0–5 headings; the smallest real prose
         list carries 13 (loanandmortgagecalculator /guides/).
      otherwise → empty_listing.
    Body = every page_component that is not hero / CTA / navigation chrome.

    ⚠ NEVER JOIN content_components TO DECIDE WHAT A PAGE CARRIES. rebuild_blog_listing
    INSERTs its row with component_id NULL (slot 'blog-listing' / 'article-grid'), and
    the first census here inner-joined the library and could not see finetuning.uk's
    22-article and ai-agent-orchestration.com's 18-article listings — two false
    positives, caught only because the ground-truth curl was sampled AWAY from the
    motivating site. LEFT JOIN, and fall back to slot_name for the name.

    IT WILL FIRE ON PAGES THAT CANNOT FILL THEMSELVES, AND THAT IS CORRECT. Per
    bugs_open/444 the remakes' feeds have 0 content_sources rows, glossary and showcase
    have no item producer anywhere in the estate, and the studio-directory kind does not
    exist. The plan-time gate (listing_item_sources.go) now refuses to PLAN such a page;
    this rule holds that door shut for the pages already built and for anything that
    gets round the gate. Do not "fix" it for over-reporting — an empty glossary is an
    empty glossary whatever the reason, and the reader cannot tell why.

    STORED-vs-SERVED: reads STORED html like rule B, with the same repair_not_served
    triage bucket. Ground-truthed 2026-09-03 at the served page, sampled away from the
    motivating site: 7 of 7 findings served empty (websitepromotion /guides/, seotools
    /blog/, loanzy /glossary.html, farmerinsurance /guides/, gamedesign /articles/,
    dartsonline /brands/, advertise /glossary.html), each with a parked-domain control
    (an invented path on the same host 404s), plus designblog's four.

    KNOWN MISS, stated so nobody re-derives it: designblog's /criticism/ ("Criticism &
    Commentary", 0 pieces) is the same defect and is NOT reported — no collection noun,
    empty directory, no listing component, so nothing structural separates it from a
    misfiled article. Widening the noun list to catch it would catch the month pages.

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

# RULE D. The planner-declared collection types (the four index roles the 444 gate calls
# listing-family, plus the bespoke directory page types), and the collection vocabulary.
STRONG_COLLECTION_TYPES = {"blog-index", "news-index", "entity-directory", "model-directory",
                           "mortgage-lenders", "health-insurers", "protocol-tracker",
                           "adoption-tracker"}
# Derived 2026-09-03 from the section directories the estate plans (guides, blog,
# articles, news, tools, games, cases, brands, shop, insights, catalog, feed, glossary,
# directory, showcases, archive) plus their Spanish forms on relojistas.com. Matched with
# word boundaries against name + url + the title BEFORE its "| Site Name" suffix — the
# suffix is why the title cannot be matched whole: "| Design Blog" and "| SEO Tools" would
# put every page of those sites in the corpus. The same alternation is used verbatim in
# the SQL pre-filter (\y there, \b here — equivalent on this ASCII vocabulary); SQL only
# SELECTS, the judgement is made here.
COLLECTION_NOUNS = ("glossary|glosario|directory|directories|feed|showcases|archive|archives|"
                    "catalogue|catalog|listings|articles|posts|guides|guias|news|noticias|blog|"
                    "tools|games|cases|case-studies|brands|shop|reviews|resources|insights|library")
RE_COLLECTION_NOUN = re.compile(r"\b(" + COLLECTION_NOUNS + r")\b", re.I)
RE_TITLE_SEP = re.compile(r"\s*[|\u2014\u2013\u00b7]\s*")
RE_RUNTIME_FILL = re.compile(r"data-runtime-fill|fetch\(", re.I)
RE_ITEM_TAG = re.compile(r"<(a\s|h3\b|h4\b|article\b|li\b|tr\b|dt\b)", re.I)
RE_DT = re.compile(r"<dt\b", re.I)
RE_H34 = re.compile(r"<h[34]\b", re.I)
RE_HREF = re.compile(r'href="([^"]+)"', re.I)
RE_CLASS = re.compile(r'class="([^"]*)"', re.I)
RE_ITEM_CLASS = re.compile(r"(card|item|entry|tile|result|term|post|article)", re.I)
CHROME_CATEGORIES = {"hero", "cta", "navigation"}


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


def is_chrome(c):
    return ((c.get("category") or "") in CHROME_CATEGORIES
            or (c.get("name") or "").endswith("hero") or c.get("name") == "call-to-action")


def own_dir(url):
    m = re.match(r"^(/[^/]+/)", url or "")
    return m.group(1) if m else None


def strip_site_name(title, domain):
    """Drop trailing title segments that are just the site's name.

    Matched on LETTERS, against the domain with and without its TLD, so the separator
    never has to be guessed: "Contact Us — Garden Tools UK" on garden-tools.uk loses
    "Garden Tools UK", while "Farm Insurance Glossary — Insurance Terms in Plain Farming
    Language | Farmer Insurance UK" loses only the last segment and keeps its glossary.
    """
    letters = lambda s: re.sub(r"[^a-z0-9]", "", (s or "").lower())
    dom = letters(domain)
    dom_no_tld = letters(re.sub(r"\.[a-z.]+$", "", domain or ""))
    parts = RE_TITLE_SEP.split(title or "")
    while len(parts) > 1 and letters(parts[-1]) in {dom, dom_no_tld} and letters(parts[-1]):
        parts.pop()
    return " ".join(parts)


def page_slug(url):
    """The page's OWN segment: 'glossary' for /glossary.html, 'the-design-feed' for
    /the-design-feed/index.html. Matching nouns against the whole path declared every
    page under /guides/ a collection — the self-test caught it on a guide at
    /guides/feedback-loops/index.html before it shipped."""
    parts = [p for p in (url or "").split("/") if p]
    if not parts:
        return ""
    if parts[-1] == "index.html" and len(parts) >= 2:
        return parts[-2]
    return re.sub(r"\.html?$", "", parts[-1])


def judge_listing(row):
    """One collection-page candidate -> verdict. The ladder is documented under RULE D.

    row: {domain, url, page_type, name, title_head, pages_in_dir,
          components: [{name, category, html, arrays: {key: length}}], repair_not_served}

    'not_a_collection': nothing declares this page a list (a bare section-index) — skipped.
    'never_built'     : no component renders anything at all — its own bucket, like rule B.
    'ok'              : lists something, and detail['listed_by'] says how it passed.
    'empty_listing'   : promised a list; renders none.
    """
    ident = " ".join([row.get("name") or "", page_slug(row.get("url")),
                      strip_site_name(row.get("title_head"), row.get("domain"))])
    detail = {"pages_in_dir": row.get("pages_in_dir") or 0}
    declared = (row.get("page_type") in STRONG_COLLECTION_TYPES
                or bool(RE_COLLECTION_NOUN.search(ident))
                or detail["pages_in_dir"] > 0)
    if not declared:
        return "not_a_collection", detail
    comps = row.get("components") or []
    if not any((c.get("html") or "").strip() for c in comps):
        return "never_built", detail
    body = [c for c in comps if not is_chrome(c)]
    html_all = " ".join((c.get("html") or "") for c in body)
    d = own_dir(row.get("url"))
    own_links = 0
    if d:
        own_links = sum(1 for m in RE_HREF.finditer(html_all)
                        if m.group(1).startswith(d) and m.group(1) != row.get("url"))
    # The repeated-item-class signal for 2b. Counts how many times the SAME class token
    # appears, so a container ("guide-list-grid") scores 1 and cannot reach the threshold;
    # only a repeated unit ("guide-card-title" x4) can.
    class_tokens = {}
    for cv in RE_CLASS.findall(html_all):
        for t in cv.split():
            if RE_ITEM_CLASS.search(t):
                class_tokens[t] = class_tokens.get(t, 0) + 1
    arrays_all = [n for c in body for n in (c.get("arrays") or {}).values()]
    detail.update({
        "body_components": [c.get("name") for c in body],
        "item_units": max(class_tokens.values() or [0]),
        "body_anchors": len(RE_HREF.findall(html_all)),
        "declared_arrays": len(arrays_all),
        "arrays": {f"{c.get('name')}.{k}": n for c in body for k, n in (c.get("arrays") or {}).items()},
        "own_links": own_links,
        "dt": len(RE_DT.findall(html_all)),
        "headings": len(RE_H34.findall(html_all)),
        "runtime_fill": bool(RE_RUNTIME_FILL.search(html_all)),
    })
    # 1. structured items that are actually RENDERED (dartsonline /brands/: 4 in data, 0 on the page)
    for c in body:
        for k, n in (c.get("arrays") or {}).items():
            if n and RE_ITEM_TAG.search(c.get("html") or ""):
                detail["listed_by"] = f"structured:{c.get('name')}.{k}={n}"
                return "ok", detail
    # 2. ESCAPE — fills at runtime (vonc provocations archive). Checked BEFORE 2b: a
    # runtime page legitimately holds an empty array and one hidden template element.
    if detail["runtime_fill"]:
        detail["listed_by"] = "runtime"
        return "ok", detail
    # 2b. NEITHER — repeated item units in the markup, nothing in the data. See RULE D.
    if detail["item_units"] >= 2 and arrays_all and not any(arrays_all):
        return "render_data_divergence", detail
    # 3. ESCAPE — links into its own directory (ported indexes, oufe /cases/)
    if own_links:
        detail["listed_by"] = f"own_links:{own_links}"
        return "ok", detail
    # 4. ESCAPE — a list written as prose headings
    if detail["dt"] >= 5 or detail["headings"] >= 10:
        detail["listed_by"] = f"prose:dt={detail['dt']},h3h4={detail['headings']}"
        return "ok", detail
    return "empty_listing", detail


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


# Rule D's corpus. ⚠ This is a deliberately LOOSE pre-filter and must stay a SUPERSET of
# the Python declaration test in judge_listing: it matches the raw title (site name and
# all) and the whole url path, and the judge then re-decides on the page's own segment
# and the title with the site name stripped, marking what it drops `not_a_collection`.
# Narrow it here and the judge never sees the page to skip it — a silent loss, invisible
# in every count the run prints.
# Index-shaped pages (a directory index, or a top-level /x.html) that
# are either typed as an index/directory or NAMED as a collection. Nested items
# (/glosario/tourbillon.html), tools, posts and landing pages are out: they are the things
# a collection lists, not collections. The Python judge applies the full declaration test
# (adding pages_in_dir > 0); this WHERE is a superset pre-filter and must stay one.
LISTING_WHERE = r"""
  p.status = 'active'
  AND (p.url ~ '/index\.html$' OR p.url ~ '^/[^/]+\.html$')
  AND COALESCE(p.page_type,'') NOT IN ('tool','blog-post','guide','entity-page','game','report','landing')
  AND (p.page_type IN ('section-index','blog-index','news-index','entity-directory','model-directory',
                       'mortgage-lenders','health-insurers','protocol-tracker','adoption-tracker')
       OR (p.name || ' ' || p.url || ' ' || split_part(COALESCE(p.title,''), '|', 1)) ~* '\y(__NOUNS__)\y'
       OR EXISTS (SELECT 1 FROM pages q
                   WHERE q.site_id = p.site_id AND q.status = 'active' AND p.url ~ '^/[^/]+/'
                     AND q.url LIKE '/' || split_part(trim(leading '/' from p.url), '/', 1) || '/%'
                     AND q.url <> p.url))
  __SITE__
""".replace("__NOUNS__", COLLECTION_NOUNS)

LISTING_COUNT_SQL = r"""
SELECT count(*) FROM pages p JOIN sites s ON s.id = p.site_id WHERE __WHERE__;
""".replace("__WHERE__", LISTING_WHERE)

# ⚠ LEFT JOIN on content_components, name falls back to slot_name — see RULE D. Chunked
# like TOOLS_SQL because the html crosses the wire; the regexes run in Python so that
# --self-test exercises the rule that actually runs. TWO transport-only measures, both
# forced by a real "unexpected EOF" on the first fleet run (2026-09-03): the chunk is 2
# (a directory listing runs to 30–50 KB per component, three per page) and <style>
# blocks are stripped server-side — 42.8% of rendered_html fleet-wide is stylesheet
# (migration 694's measurement) and no rule here reads CSS. The pattern is 601's:
# '<style[^>]*?>.*?</style>' — Postgres takes the greediness of the FIRST quantifier, so
# the lazy '*?' in the tag makes the whole match lazy; the greedy '[^>]*' form deletes
# prose between two style blocks (LANDMINES 2026-09-02). <script> is kept: the runtime
# escape reads 'fetch(' from it.
LISTING_CHUNK = 2
LISTING_SQL = r"""
SELECT COALESCE(jsonb_agg(row_to_json(t)), '[]'::jsonb) FROM (
  SELECT s.domain, p.url, p.page_type, p.name,
         split_part(COALESCE(p.title,''), '|', 1) AS title_head,
         (SELECT count(*) FROM pages q
           WHERE q.site_id = p.site_id AND q.status = 'active' AND p.url ~ '^/[^/]+/'
             AND q.url LIKE '/' || split_part(trim(leading '/' from p.url), '/', 1) || '/%'
             AND q.url <> p.url) AS pages_in_dir,
         (p.build_status = 'deployed' AND p.deployed_at IS NOT NULL
          AND (SELECT max(pc2.updated_at) FROM page_components pc2
                WHERE pc2.page_id = p.id) > p.deployed_at + interval '1 minute'
         ) AS repair_not_served,
         (SELECT COALESCE(jsonb_agg(jsonb_build_object(
                   'name', COALESCE(cc.name, pc.slot_name),
                   'category', COALESCE(cc.category, ''),
                   'html', regexp_replace(COALESCE(pc.rendered_html, ''), '<style[^>]*?>.*?</style>', ' ', 'gi'),
                   'arrays', (SELECT COALESCE(jsonb_object_agg(kv.key, jsonb_array_length(kv.value)), '{}'::jsonb)
                                FROM jsonb_each(COALESCE(pc.content_data, '{}'::jsonb)) kv
                               WHERE jsonb_typeof(kv.value) = 'array')
                 ) ORDER BY pc.position), '[]'::jsonb)
            FROM page_components pc LEFT JOIN content_components cc ON cc.id = pc.component_id
           WHERE pc.page_id = p.id AND pc.build_status <> 'removed') AS components
  FROM pages p JOIN sites s ON s.id = p.site_id
  WHERE __WHERE__
  ORDER BY s.domain, p.url LIMIT __LIMIT__ OFFSET __OFFSET__) t;
""".replace("__WHERE__", LISTING_WHERE)


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
    n_listings = int(_q(LISTING_COUNT_SQL.replace("__SITE__", clause)) or "0")
    listings = []
    for off in range(0, n_listings, LISTING_CHUNK):
        page = _q(LISTING_SQL.replace("__SITE__", clause)
                  .replace("__LIMIT__", str(LISTING_CHUNK)).replace("__OFFSET__", str(off)))
        listings.extend(json.loads(page or "[]"))
    if len(listings) != n_listings:
        print(f"fetch incomplete: {len(listings)} of {n_listings} listing pages returned",
              file=sys.stderr)
        sys.exit(2)
    return {"nav": nav, "tools": tools, "indexes": indexes, "listings": listings}


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

    listing_findings, listing_never_built, listing_unserved = [], [], []
    listing_divergence = []
    listing_ok, listing_skipped = 0, 0
    listing_demand = {"structured": 0, "runtime": 0, "own_links": 0, "prose": 0}
    for row in data.get("listings", []):
        verdict, detail = judge_listing(row)
        base = {"domain": row["domain"], "url": row["url"], "page_type": row.get("page_type")}
        if verdict == "empty_listing":
            listing_findings.append({**base, **detail})
        elif verdict == "never_built":
            listing_never_built.append({**base, **detail})
        elif verdict == "render_data_divergence":
            listing_divergence.append({**base, **detail})
        elif verdict == "not_a_collection":
            listing_skipped += 1
        elif row.get("repair_not_served"):
            listing_unserved.append({**base, **detail})
        else:
            listing_ok += 1
            listing_demand[detail.get("listed_by", "").split(":")[0]] += 1

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
        "listing_pages_scanned": len(data.get("listings", [])),
        "rule_d_empty_listing": listing_findings,
        "listing_pages_never_built": listing_never_built,
        "listing_pages_render_data_divergence": listing_divergence,
        "listing_pages_repair_not_served": listing_unserved,
        "listing_pages_ok": listing_ok,
        "listing_pages_not_a_collection": listing_skipped,
        # DEMAND CONTROL: rule B is only evidence if a tool page can pass it. If these go
        # to zero the regexes have drifted, and every "clean" tool page is unexamined.
        "demand_control": demand,
        # Rule D's demand control: how the clean pages passed. All four at zero means no
        # collection page in the corpus could pass, i.e. the ladder is blind.
        "listing_demand_control": listing_demand,
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
             f"Rule D (a collection page with nothing in it): "
             f"{len(r.get('rule_d_empty_listing', []))} of "
             f"{r.get('listing_pages_scanned', 0)} collection-page candidates "
             f"({r.get('listing_pages_ok', 0)} list something, "
             f"{r.get('listing_pages_not_a_collection', 0)} bare section-indexes skipped, "
             f"{len(r.get('listing_pages_never_built', []))} never built, "
             f"{len(r.get('listing_pages_render_data_divergence', []))} render-vs-data divergence, "
             f"{len(r.get('listing_pages_repair_not_served', []))} repair-not-served triage)",
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
    for f in r.get("rule_d_empty_listing", []):
        lines.append(f"[D] {f['domain']}{f['url']} ({f['page_type']}): promises a list and "
                     f"renders none — body {'+'.join(f.get('body_components') or []) or '(chrome only)'}, "
                     f"arrays {f.get('arrays') or {}}, {f.get('own_links', 0)} own-directory links, "
                     f"{f.get('headings', 0)} h3/h4, {f['pages_in_dir']} pages in its directory.")
    if r.get("rule_d_empty_listing"):
        lines.append("    a rule D page CANNOT usually fill itself (bugs_open/444: no content_sources "
                     "row, no glossary/showcase producer, no directory kind) — the fix is upstream, "
                     "at plan time, and the plan-time gate now refuses to plan one. Do not narrow "
                     "this rule for over-reporting: an empty glossary is empty whatever the reason.")
    for f in r.get("listing_pages_render_data_divergence", []):
        lines.append(f"[~] {f['domain']}{f['url']}: renders {f['item_units']} repeated item units "
                     f"while every declared array is empty, and the body carries "
                     f"{f['body_anchors']} links — the markup predates the data, so the reader "
                     f"sees items that the site no longer has and may not be able to click any. "
                     f"NOT a rule D finding (it renders items) and NOT clean.")
    for f in r.get("listing_pages_repair_not_served", []):
        lines.append(f"[!] {f['domain']}{f['url']}: stored html lists something but its newest "
                     f"component postdates deployed_at — not counted clean (rule D reads STORED html).")
    for f in r.get("listing_pages_never_built", []):
        lines.append(f"[--] {f['domain']}{f['url']}: collection page with no rendered html at all "
                     f"(never built; {f['pages_in_dir']} pages in its directory)")
    d = r["demand_control"]
    ld = r.get("listing_demand_control", {})
    lines += ["",
              f"Rule D demand control: of {r.get('listing_pages_ok', 0)} clean collection pages, "
              f"{ld.get('structured', 0)} passed on rendered structured items, "
              f"{ld.get('runtime', 0)} on runtime fill, {ld.get('own_links', 0)} on own-directory "
              f"links, {ld.get('prose', 0)} on prose headings. All zero = the ladder is blind.",
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
    print(f"\n  RULE D — a collection page with nothing in it: "
          f"{len(r.get('rule_d_empty_listing', []))}"
          f"  (of {r.get('listing_pages_scanned', 0)} candidates: {r.get('listing_pages_ok', 0)} list "
          f"something, {r.get('listing_pages_not_a_collection', 0)} bare section-indexes skipped)")
    for f in r.get("rule_d_empty_listing", []):
        print(f"    {f['domain']}{f['url']}  ({f['page_type']}) — body "
              f"{'+'.join(f.get('body_components') or []) or '(chrome only)'}; arrays "
              f"{f.get('arrays') or {}}; {f.get('own_links', 0)} own-dir links; "
              f"{f.get('headings', 0)} h3/h4; {f['pages_in_dir']} pages in dir")
    if r.get("listing_pages_render_data_divergence"):
        print(f"    ~ {len(r['listing_pages_render_data_divergence'])} page(s) RENDER items their "
              f"DATA does not have (markup predates the data; neither a finding nor clean):")
        for f in r["listing_pages_render_data_divergence"]:
            print(f"      {f['domain']}{f['url']}  {f['item_units']} repeated item units, "
                  f"{f['body_anchors']} links in the body, {f['declared_arrays']} declared "
                  f"array(s), all empty")
    if r.get("listing_pages_repair_not_served"):
        print(f"    ⚠ {len(r['listing_pages_repair_not_served'])} collection page(s) list something "
              f"in STORED html but their newest component postdates deployed_at — not counted clean:")
        for f in r["listing_pages_repair_not_served"]:
            print(f"      {f['domain']}{f['url']}")
    if r.get("listing_pages_never_built"):
        print(f"    separately, {len(r['listing_pages_never_built'])} collection page(s) with no "
              f"rendered html at all (never built):")
        for f in r["listing_pages_never_built"]:
            print(f"      {f['domain']}{f['url']}  ({f['pages_in_dir']} pages in dir)")
    ld = r.get("listing_demand_control", {})
    print(f"    demand control — clean pages passed by: structured {ld.get('structured', 0)}, "
          f"runtime {ld.get('runtime', 0)}, own-dir links {ld.get('own_links', 0)}, "
          f"prose headings {ld.get('prose', 0)}")
    if r.get("listing_pages_scanned") and not any(ld.values()):
        # A --site run can legitimately contain only findings (designblog: 4 of 4). That is
        # not blindness, and printing FAIL there is the SQ-004 control-scope bug in a new
        # coat (fixed there 2026-09-03). The demand control is a FLEET question.
        if args.site:
            print("    n/a — no candidate in --site scope passed; the demand control needs the "
                  "fleet corpus (a scoped zero here is not evidence of blindness)")
        else:
            print("    FAIL — no collection page passed by any step; rule D is blind.")
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
        if args.site:
            print("    n/a — every tool page in --site scope failed every signal; the demand "
                  "control needs the fleet corpus (a scoped zero is not evidence of blindness)")
        else:
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

    # RULE D — every fixture is a page shape measured 2026-09-03 (stored components), and the
    # verdict on each was confirmed at the SERVED page with a parked-domain control.
    PROSE = ("<h2>What this glossary is for</h2><p>A design glossary is only useful when it "
             "explains why a term exists.</p><h3>How the entries are written</h3><p>…</p>"
             "<h3>What's covered</h3><ul><li>layout</li><li>typography</li><li>colour</li>"
             "<li>UX</li><li>brand</li><li>a11y</li></ul><h3>Where the definitions lead</h3>"
             "<p>…<a href=\"/tools/css-unit-converter/index.html\">CSS Unit Converter</a>…</p>")
    HERO = {"name": "hero", "category": "hero", "html": "<section class=\"hero\"><h1>Design "
            "Glossary</h1><a href=\"/glossary.html\">Browse</a></section>", "arrays": {}}
    CTA = {"name": "call-to-action", "category": "cta",
           "html": "<section class=\"cta\"><a href=\"/contact.html\">Get in touch</a></section>",
           "arrays": {}}
    d_cases = [
        ("designblog /glossary.html: page_type CONTENT, hero + prose about the glossary, 0 terms "
         "(the case that funded the rule — invisible to any page_type selector)",
         {"url": "/glossary.html", "page_type": "content", "name": "glossary",
          "title_head": "Design Glossary ", "pages_in_dir": 0,
          "components": [HERO, {"name": "Generic Text Block", "category": "content",
                                "html": PROSE, "arrays": {}}]},
         "empty_listing"),
        ("designblog /inspiration/: section-index with 0 pages in /inspiration/, declared a "
         "collection only by its title 'Inspiration Showcases'",
         {"url": "/inspiration/index.html", "page_type": "section-index", "name": "inspiration-index",
          "title_head": "Inspiration Showcases ", "pages_in_dir": 0,
          "components": [HERO, {"name": "Generic Text Block", "category": "content",
                                "html": PROSE, "arrays": {}}, CTA]},
         "empty_listing"),
        ("designblog /uk-studios-directory/: entity-directory whose directory-listing carries a "
         "headline and no entries key (bugs_open/444 mechanism 2)",
         {"url": "/uk-studios-directory/index.html", "page_type": "entity-directory",
          "name": "uk-studios-directory-index", "title_head": "UK Design Studios Directory ",
          "pages_in_dir": 0,
          "components": [HERO, {"name": "directory-listing", "category": "content",
                                "html": "<section class=\"directory-listing\"><h2>Find UK design "
                                        "studios</h2></section>", "arrays": {}}]},
         "empty_listing"),
        ("advertise /news/: news-listing with items=[] and 0 content_sources (444 mechanism 1)",
         {"url": "/news/index.html", "page_type": "news-index", "name": "news-index",
          "title_head": "UK Advertising News ", "pages_in_dir": 0,
          "components": [HERO, {"name": "news-listing", "category": "content",
                                "html": "<section class=\"news-listing\"><p>Latest news</p></section>",
                                "arrays": {"items": 0}}]},
         "empty_listing"),
        ("dartsonline /brands/: nav_items=4 IN THE DATA, 315 bytes rendering nothing — the "
         "rendered-item AND in step 1",
         {"url": "/brands/index.html", "page_type": "section-index", "name": "brands-index",
          "title_head": "All Brands ", "pages_in_dir": 0,
          "components": [HERO, {"name": "category-listing", "category": "content-site",
                                "html": "<section class=\"category-listing\"><div class=\"grid\">"
                                        "</div></section>", "arrays": {"nav_items": 4, "services": 0}}]},
         "empty_listing"),
        ("a blog-index that is hero + CTA and nothing else — chrome-only body is empty, not "
         "'never built'",
         {"url": "/blog.html", "page_type": "blog-index", "name": "blog", "title_head": "Blog ",
          "pages_in_dir": 0, "components": [HERO, CTA]},
         "empty_listing"),
        ("websitepromotion /guides/: 8 pages in /guides/, guide-list with no items key, hero links "
         "only to the page ITSELF (self-link excluded from own_links)",
         {"url": "/guides/index.html", "page_type": "section-index", "name": "guides-index",
          "title_head": "Website Promotion Guides ", "pages_in_dir": 8,
          "components": [{"name": "hero", "category": "hero", "arrays": {},
                          "html": "<a href=\"/guides/index.html\">Guides</a>"},
                         {"name": "guide-list_pre_037", "category": "content", "arrays": {},
                          "html": "<section><h2>Guides for getting your website found</h2>"
                                  "<p>More guides are on the way.</p></section>"}]},
         "empty_listing"),
        ("finetuning /blog.html: the listing row has component_id NULL (slot 'article-grid'), "
         "22 articles rendered — the inner-join false positive, must be OK",
         {"url": "/blog.html", "page_type": "blog-index", "name": "blog", "title_head": "Blog ",
          "pages_in_dir": 0,
          "components": [HERO, {"name": "article-grid", "category": "",
                                "html": "<section class=\"section--articles\">" +
                                        "<article class=\"article-card\"><h3>x</h3></article>" * 22 +
                                        "</section>", "arrays": {"articles": 22}}, CTA]},
         "ok"),
        ("ESCAPE runtime: vonc /provocations/ fills at runtime and says so (data-runtime-fill)",
         {"url": "/provocations/index.html", "page_type": "section-index", "name": "provocations-index",
          "title_head": "Provocations Archive ", "pages_in_dir": 0,
          "components": [{"name": "provocations-archive-list", "category": "content", "arrays": {},
                          "html": "<section data-runtime-fill=\"true\"><a data-archive-template hidden "
                                  "href=\"#\"></a><p>Nothing filed yet.</p></section>"}]},
         "ok"),
        ("ESCAPE own links: webdesign /learn/ is a ported page listing 31 /learn/ pages in markup "
         "only (no content_data arrays)",
         {"url": "/learn/index.html", "page_type": "section-index", "name": "learn-index",
          "title_head": "Learn ", "pages_in_dir": 31,
          "components": [{"name": "Ported Page (webdesign.co.uk)", "category": "", "arrays": {},
                          "html": "".join(f"<h3><a href=\"/learn/p{i}.html\">p{i}</a></h3>"
                                          for i in range(31))}]},
         "ok"),
        ("ESCAPE own links, the smallest case: oufe /cases/ links to its one case among tool links",
         {"url": "/cases/index.html", "page_type": "section-index", "name": "cases-index",
          "title_head": "Case studies ", "pages_in_dir": 1,
          "components": [{"name": "Generic Text Block", "category": "content", "arrays": {},
                          "html": "<h2>Cases</h2><p><a href=\"/cases/thames-water.html\">Thames "
                                  "Water</a> <a href=\"/tools/x/index.html\">tool</a></p>"}]},
         "ok"),
        ("ESCAPE prose: a glossary written as one h3 per term (12 terms)",
         {"url": "/glossary.html", "page_type": "content", "name": "glossary",
          "title_head": "Glossary ", "pages_in_dir": 0,
          "components": [HERO, {"name": "Generic Text Block", "category": "content", "arrays": {},
                                "html": "".join(f"<h3>Term {i}</h3><p>def</p>" for i in range(12))}]},
         "ok"),
        ("NEITHER: farmerinsurance /guides/ renders guide cards (stored 4, served 3, ZERO "
         "anchors in either) while content_data.items is [] with an unshown empty_state_text",
         {"domain": "farmerinsurance.uk", "url": "/guides/index.html", "page_type": "section-index",
          "name": "guides-index", "title_head": "Farm Insurance Guides ", "pages_in_dir": 0,
          "components": [HERO, {"name": "guide-list_pre_037", "category": "content",
                                "arrays": {"items": 0},
                                "html": "<section class=\"guide-list-section\"><div class="
                                        "\"guide-list-grid\">" +
                                        "".join("<span class=\"guide-card-badge\">g</span>"
                                                "<h3 class=\"guide-card-title\">g</h3>"
                                                "<p class=\"guide-card-desc\">d</p>"
                                                "<span class=\"guide-card-link-label\">Read guide"
                                                "</span>" for _ in range(4)) + "</div></section>"}]},
         "render_data_divergence"),
        ("a CONTAINER class alone cannot reach the threshold: one 'listing-card-grid' wrapper "
         "round nothing is still an empty listing",
         {"domain": "x.uk", "url": "/directory/index.html", "page_type": "entity-directory",
          "name": "directory-index", "title_head": "Directory ", "pages_in_dir": 0,
          "components": [HERO, {"name": "directory-listing", "category": "content",
                                "arrays": {"entries": 0},
                                "html": "<section class=\"listing-card-grid\"><h2>Find one</h2>"
                                        "</section>"}]},
         "empty_listing"),
        ("runtime fill still WINS over divergence: vonc's hidden template element is not items",
         {"domain": "vonc.com", "url": "/provocations/index.html", "page_type": "section-index",
          "name": "provocations-index", "title_head": "Provocations Archive ", "pages_in_dir": 0,
          "components": [{"name": "provocations-archive-list", "category": "content",
                          "arrays": {"items": 0},
                          "html": "<section data-runtime-fill=\"true\">"
                                  "<a class=\"archive__item\" data-archive-template hidden></a>"
                                  "<a class=\"archive__item\" data-archive-template hidden></a>"
                                  "</section>"}]},
         "ok"),
        ("NEGATIVE CONTROL: homegarden /april/ — section-index, empty directory, no collection "
         "noun, 4 h3 of prose; an article misfiled as an index, NOT a finding",
         {"url": "/april/index.html", "page_type": "section-index", "name": "april-index",
          "title_head": "April — Garden and Home Jobs for This Month ", "pages_in_dir": 0,
          "components": [HERO, {"name": "Generic Text Block", "category": "content", "arrays": {},
                                "html": "<h3>a</h3><h3>b</h3><h3>c</h3><h3>d</h3><p>…</p>"}]},
         "not_a_collection"),
        ("NEGATIVE CONTROL: garden-tools /contact.html — the site name IS a collection noun "
         "('Contact Us — Garden Tools UK'), and an em-dash separator defeats a '|' split",
         {"domain": "garden-tools.uk", "url": "/contact.html", "page_type": "content",
          "name": "contact", "title_head": "Contact Us — Garden Tools UK", "pages_in_dir": 0,
          "components": [{"name": "contact-hero", "category": "hero", "html": "<h1>Get in touch</h1>",
                          "arrays": {}},
                         {"name": "contact-form", "category": "content", "arrays": {},
                          "html": "<form><input name='name'><button>Send</button></form>"}]},
         "not_a_collection"),
        ("NEGATIVE CONTROL: garden-tools /affiliate-disclosure.html, same site-name leak",
         {"domain": "garden-tools.uk", "url": "/affiliate-disclosure.html", "page_type": "content",
          "name": "affiliate-disclosure", "title_head": "Affiliate Disclosure — Garden Tools UK",
          "pages_in_dir": 0,
          "components": [{"name": "Generic Text Block", "category": "content", "arrays": {},
                          "html": "<h2>How we make money</h2><p>We earn commission.</p>"}]},
         "not_a_collection"),
        ("but an em-dash INSIDE a real title is kept: farmerinsurance's glossary still counts",
         {"domain": "farmerinsurance.uk", "url": "/glossary.html", "page_type": "content",
          "name": "glossary",
          "title_head": "Farm Insurance Glossary — Insurance Terms in Plain Farming Language | Farmer Insurance UK",
          "pages_in_dir": 0,
          "components": [HERO, {"name": "Generic Text Block", "category": "content", "arrays": {},
                                "html": "<h3>a</h3><h3>b</h3><h3>c</h3><p>…</p>"}]},
         "empty_listing"),
        ("NEGATIVE CONTROL: mortgagecalculator /contact/ typed section-index",
         {"url": "/contact/index.html", "page_type": "section-index", "name": "contact-index",
          "title_head": "Contact us", "pages_in_dir": 0,
          "components": [{"name": "contact-hero", "category": "hero", "html": "<h1>Contact</h1>",
                          "arrays": {}},
                         {"name": "Generic Text Block", "category": "content", "arrays": {},
                          "html": "<p>Write to us.</p>"}]},
         "not_a_collection"),
        ("title suffix is NOT matched: 'Claim Your Practice Listing' is singular and the site "
         "suffix '| VetComparison.uk — Directory' is cut before matching",
         {"url": "/entities/practice.html", "page_type": "content", "name": "practice",
          "title_head": "Claim Your Practice Listing ", "pages_in_dir": 0,
          "components": [HERO, {"name": "Generic Text Block", "category": "content", "arrays": {},
                                "html": "<p>Claim your listing.</p>"}, CTA]},
         "not_a_collection"),
        ("garden-tools /brand-directory/: entity-directory with no components at all",
         {"url": "/brand-directory/index.html", "page_type": "entity-directory",
          "name": "brand-directory-index", "title_head": "UK Garden Tool Brands ",
          "pages_in_dir": 0, "components": []},
         "never_built"),
        ("loanzy /guides/: 7 pages in /guides/ and every component renders empty html",
         {"url": "/guides/index.html", "page_type": "section-index", "name": "guides-index",
          "title_head": "Loan Guides ", "pages_in_dir": 7,
          "components": [{"name": "hero", "category": "hero", "html": "  ", "arrays": {}}]},
         "never_built"),
    ]
    for name, row, expected in d_cases:
        got, detail = judge_listing(row)
        if got != expected:
            failures.append(f"    {name}: expected {expected}, got {got} ({detail})")
        else:
            print(f"  PASS  {name[:96]}")
    # The declaration must read the page's OWN segment: idea.uk's guide at
    # /guides/feedback-loops/index.html is neither a feed nor (by its parent dir) a guides index.
    got, _ = judge_listing({"url": "/guides/feedback-loops/index.html", "page_type": "guide",
                            "name": "feedback-loops", "title_head": "Feedback loops: improving on "
                            "what users actually tell you", "pages_in_dir": 0,
                            "components": [HERO, {"name": "Generic Text Block", "category": "content",
                                                  "html": "<p>…</p>", "arrays": {}}]})
    if got != "not_a_collection":
        failures.append(f"    a guide under /guides/ read as a collection page: {got}")
    else:
        print("  PASS  a guide under /guides/ is not a collection page (own segment, not the path)")
    if RE_COLLECTION_NOUN.search("feedback"):
        failures.append("    collection noun matched inside 'feedback' — word boundary lost")
    else:
        print("  PASS  'feedback' does not read as a feed")

    # The note path, which the cluster had to find for this lane once already.
    sample = analyse({"nav": boxing and [{**e, "domain": "boxingonline.com"} for e in boxing],
                      "tools": [{"domain": "boxingonline.com",
                                 "url": "/tools/fight-calendar/index.html",
                                 "html": "<p>How we build the fight calendar</p>"}],
                      "indexes": [{"domain": "boxingonline.com", "host": "/guides/index.html",
                                   "page_type": "section-index", "dir": "/guides/",
                                   "pages_in_dir": 4, "item_urls": ["/blog/a.html"]}],
                      "listings": [{"domain": "designblog.co.uk", "url": "/glossary.html",
                                    "page_type": "content", "name": "glossary",
                                    "title_head": "Design Glossary ", "pages_in_dir": 0,
                                    "components": [HERO, {"name": "Generic Text Block",
                                                          "category": "content", "html": PROSE,
                                                          "arrays": {}}]},
                                   {"domain": "garden-tools.uk", "url": "/brand-directory/index.html",
                                    "page_type": "entity-directory", "name": "brand-directory-index",
                                    "title_head": "Brands ", "pages_in_dir": 0, "components": []}]})
    try:
        body = note_body(sample)
    except Exception as exc:                       # noqa: BLE001 — report, never raise
        failures.append(f"    note_body raised {exc!r}")
    else:
        if "$epc$" in body:
            failures.append("    note body contains the dollar-quote tag it is wrapped in")
        elif any(tag not in body for tag in ("[A]", "[B]", "[C]", "[D]", "[--] garden-tools.uk")):
            failures.append("    note body drops a rule's findings")
        else:
            print("  PASS  doc_notes body builds and carries all four rules")
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
