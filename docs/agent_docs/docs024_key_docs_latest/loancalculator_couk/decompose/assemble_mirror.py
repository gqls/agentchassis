#!/usr/bin/env python3
"""assemble_mirror.py — a faithful Python mirror of RerenderSinglePageAction's
assembly, so the decomposition can be proven BEFORE any row is written.

WHY A MIRROR AT ALL, AND WHY THAT IS UNCOMFORTABLE. The real assembler is Go
(platform/orchestration/actions/rerender_single_page_action.go: assemblePage),
it is unexported, and it needs a database and a page id. There is no way to ask
it "what would you produce if I wrote these rows?" without writing the rows —
and writing them changes 27 live money pages from shipping stored bytes to being
assembled, which is the very thing being tested. So the choice is: ship first and
look afterwards, or mirror.

A mirror is a SECOND IMPLEMENTATION, which is exactly the thing this repo keeps
warning about: two implementations of the same rule agree with each other for the
reasons they are both wrong. That objection is right and is answered by never
trusting the mirror on its own word:

  THE FIRST PAGE SHIPPED VALIDATES THE MIRROR. One page is decomposed, the real
  Go path renders it, and its output is diffed against this file's prediction for
  that page. If they differ, the mirror is wrong and the remaining 26 pages do
  not move until it is fixed. The mirror is a hypothesis with a scheduled test,
  not an authority.

WHAT IS MIRRORED, function by function, with the Go line it tracks:

  assemblePage             DOCTYPE + <html lang="en"> + head + <body> + header +
                           <main> + sections + </main> + footer + </body></html>,
                           in that order and with those exact newlines.
  getPageSections          concatenation in `position` order, each section
                           followed by "\\n", skipping rows whose rendered_html
                           is empty or has no visible content.
  sectionHasVisibleContent keep if data-runtime-fill is present, else strip
                           <style>/<script>/tags/entities/whitespace and require
                           MORE THAN 10 characters left.
  title / meta injection   <title>...</title> replaced wholesale; the FIRST
                           literal `content="">` replaced with the description.
  injectPageJSONLD         one WebPage block before the LAST </head>, skipped if
                           the head already contains application/ld+json or the
                           page has no title, with Go's json.Marshal HTML
                           escaping of <, > and & reproduced.
  collectComponentCSS      deliberately a no-op — see the note in inject order.
  StripToolDocHeader       sentinel scan over the whole assembled page, removing
                           each block plus one trailing newline, idempotent, and
                           leaving a malformed block untouched.

WHAT IS DELIBERATELY NOT MIRRORED, and why that is safe here:

  repairOutboundPageLinks  It rewrites internal links that do not resolve against
                           pages.url. Not mirrored because it is a REPAIR: if it
                           fires, this site's links were already broken and the
                           mirror's job is to notice that, not to reproduce the
                           patch. verify_assembled.py therefore checks every
                           internal link against the real page list instead, and
                           a difference between mirror and production caused by
                           this would show up as a link the checker already
                           flagged.
  collectComponentCSS      Its query only injects a css_snippets row when the
                           matching component ships NO <style> of its own. All
                           eleven tool components ship their own; the prose
                           component's function name matches no applies_to entry
                           (measured: the 20 rows key on `card`, `hero`, `button`,
                           `section`, `latest-news` and similar generic names).
                           So it contributes nothing, and a no-op mirror of a
                           no-op is exact rather than approximate. If a snippet is
                           ever added for one of these functions this stops being
                           true, which is why the first-page diff exists.
"""
import html as htmlmod
import json
import os
import re
import subprocess
import tempfile

# ── the regexes getPageSections filters with ────────────────────────────────
RE_RUNTIME_FILL = re.compile(r"data-runtime-fill", re.I)
RE_STYLE_BLOCKS = re.compile(r"<style\b[^>]*>.*?</style>", re.S | re.I)
RE_SCRIPT_BLOCKS = re.compile(r"<script\b[^>]*>.*?</script>", re.S | re.I)
RE_HTML_TAGS = re.compile(r"<[^>]*>")
RE_HTML_ENTITIES = re.compile(r"&[a-zA-Z#0-9]+;")
RE_WHITESPACE = re.compile(r"\s+")
RE_TITLE = re.compile(r"<title>[^<]*</title>")

TOOL_DOC_OPEN = "/* === tool-doc ==="
TOOL_DOC_CLOSE = "=== /tool-doc === */"


def section_has_visible_content(s):
    """Mirror of sectionHasVisibleContent. The >10 is strictly greater."""
    if not s:
        return False
    if RE_RUNTIME_FILL.search(s):
        return True
    t = RE_STYLE_BLOCKS.sub("", s)
    t = RE_SCRIPT_BLOCKS.sub("", t)
    t = RE_HTML_TAGS.sub("", t)
    t = RE_HTML_ENTITIES.sub("", t)
    t = RE_WHITESPACE.sub("", t)
    return len(t) > 10


def strip_tool_doc_header(s):
    """Mirror of StripToolDocHeader, including the one swallowed newline."""
    while True:
        o = s.find(TOOL_DOC_OPEN)
        if o < 0:
            return s
        rest = s[o + len(TOOL_DOC_OPEN):]
        c = rest.find(TOOL_DOC_CLOSE)
        if c < 0:
            return s  # malformed — Go leaves it alone, so do we
        end = o + len(TOOL_DOC_OPEN) + c + len(TOOL_DOC_CLOSE)
        if end < len(s) and s[end] == "\r":
            end += 1
        if end < len(s) and s[end] == "\n":
            end += 1
        s = s[:o] + s[end:]


def _go_json(obj):
    """json.Marshal's escaping: <, > and & become \\u003c, \\u003e, \\u0026.

    Not cosmetic. Go escapes these so a title containing "</script>" cannot break
    out of the block it is embedded in, and the comment above the Go call says in
    terms not to switch it off. A mirror that emits a raw "<" would predict bytes
    the platform never produces, and the first-page diff would then blame the
    decomposition for a difference that is the mirror's own.
    """
    s = json.dumps(obj, separators=(",", ":"), ensure_ascii=False, sort_keys=True)
    return s.replace("<", "\\u003c").replace(">", "\\u003e").replace("&", "\\u0026")


def inject_page_jsonld(head, domain, url, title, meta_desc):
    if not head or "application/ld+json" in head:
        return head
    if not title or not domain:
        return head
    origin = "https://" + domain
    page_url = origin + url
    doc = {
        "@context": "https://schema.org",
        "@type": "WebPage",
        "@id": page_url,
        "url": page_url,
        "name": title,
        "isPartOf": {"@type": "WebSite", "url": origin, "name": domain},
    }
    if meta_desc:
        doc["description"] = meta_desc
    # Go builds a map[string]interface{} and marshals it; encoding/json sorts map
    # keys alphabetically at EVERY level, which is why _go_json sorts recursively
    # rather than preserving insertion order. So `description` lands between
    # @type and isPartOf, and isPartOf's own keys come out @type/name/url —
    # neither of which is the order they are written in above.
    block = '\n<script type="application/ld+json">%s</script>\n' % _go_json(doc)
    i = head.rfind("</head>")
    if i >= 0:
        return head[:i] + block + head[i:]
    return head + block


def assemble_page(head, header, footer, sections_html, domain, url, title, meta_desc):
    """Mirror of assemblePage + the outbound tool-doc strip.

    `sections_html` is the already-filtered, already-joined section string, so
    that the caller can report WHICH rows were dropped rather than have them
    disappear inside here.
    """
    if title:
        head = RE_TITLE.sub("<title>%s</title>" % title, head)
    if meta_desc:
        head = head.replace('content="">', 'content="%s">' % meta_desc, 1)
    head = inject_page_jsonld(head, domain, url, title, meta_desc)

    out = ['<!DOCTYPE html>\n<html lang="en">\n', head, "\n<body>\n"]
    if header:
        out.append(header)
        out.append("\n")
    out.append("<main>\n")
    out.append(sections_html)
    out.append("\n</main>\n")
    if footer:
        out.append(footer)
        out.append("\n")
    out.append("</body>\n</html>")
    return strip_tool_doc_header("".join(out))


def join_sections(rows):
    """Mirror of getPageSections' concatenation. Returns (html, dropped).

    `rows` is a list of (slot_name, rendered_html) in position order.
    """
    parts, dropped = [], []
    for slot, h in rows:
        if not h:
            dropped.append((slot, "unrendered"))
            continue
        if not section_has_visible_content(h):
            dropped.append((slot, "no visible content"))
            continue
        parts.append(h)
        parts.append("\n")
    return "".join(parts), dropped


def content_data_for(rewrite_dir, function, overrides=None):
    """Every field the component takes, with this page's overrides applied.

    THE FULL SET, not just the overrides. page_components.content_data is what a
    later re-render reads, so a row that stores only the two fields it changed
    depends on the renderer resolving the other thirteen from the schema
    fallback — and on that resolution never changing. Storing all of them makes
    the row self-describing: what shipped is what is written down.
    """
    schema = json.load(open(os.path.join(rewrite_dir, function + ".schema.json"),
                            encoding="utf-8"))
    data = {k: v.get("fallback", "") for k, v in schema["fields"].items()}
    for k, v in (overrides or {}).items():
        if k not in data:
            raise KeyError("%s has no field %r to override" % (function, k))
        data[k] = v
    return data


def render_component(rewrite_dir, function, overrides=None, cache={}):
    """Render a tool component with the REAL Go template engine.

    Same binary path verify_rewrite.py uses (render_tool.go), so a template that
    renders here renders identically there. Cached because eleven `go run`
    invocations per page would dominate the runtime of the verifier.

    Overrides are applied by writing a temporary schema whose fallbacks carry
    this page's values, rather than by teaching render_tool.go a second input
    format. The output is identical either way — render_tool.go's only source of
    values IS the schema — and it keeps the shipped renderer to one code path.
    """
    key = (function, tuple(sorted((overrides or {}).items())))
    if key in cache:
        return cache[key]
    tmpl = os.path.join(rewrite_dir, function + ".html.tmpl")
    schema = os.path.join(rewrite_dir, function + ".schema.json")
    if overrides:
        doc = json.load(open(schema, encoding="utf-8"))
        for k, v in overrides.items():
            if k not in doc["fields"]:
                raise KeyError("%s has no field %r to override" % (function, k))
            doc["fields"][k]["fallback"] = v
        schema = os.path.join(tempfile.gettempdir(), function + ".override.schema.json")
        json.dump(doc, open(schema, "w", encoding="utf-8"), ensure_ascii=False)
    out = os.path.join(tempfile.gettempdir(), function + ".assembled.html")
    r = subprocess.run(
        ["go", "run", os.path.join(rewrite_dir, "render_tool.go"), tmpl, schema, out],
        capture_output=True, text=True, cwd=rewrite_dir)
    if r.returncode != 0:
        raise RuntimeError("render %s failed: %s"
                           % (function, (r.stderr or r.stdout).strip()[:300]))
    cache[key] = open(out, encoding="utf-8").read()
    return cache[key]


PROSE_TEMPLATE = '<section class="ported-prose" data-component="ported-prose">%s</section>'


def rows_for_page(page, rewrite_dir):
    """Turn a manifest page into the rows to write.

    Returns (slot_name, rendered_html, function, content_data) in position order.

    Slot names are positional and kind-bearing (`prose-0`, `tool-3`) because the
    slot is what every log line, every dropped-section warning and every
    build-status query names. A row called `section` tells an operator nothing
    about which paragraph vanished.
    """
    rows = []
    for i, b in enumerate(page["blocks"]):
        if b["kind"] == "prose":
            # The prose row's content_data carries the raw block, so the row is
            # re-renderable from its own data rather than only from the HTML it
            # happens to be holding.
            rows.append(("prose-%d" % i, PROSE_TEMPLATE % b["html"], None,
                         {"content": b["html"]}))
        else:
            ov = page.get("tool_content_data") or {}
            rows.append(("tool-%d" % i,
                         render_component(rewrite_dir, b["function"], ov),
                         b["function"],
                         content_data_for(rewrite_dir, b["function"], ov)))
    return rows


def visible_text(s):
    """Visible text with entities decoded and whitespace collapsed.

    Entities are DECODED rather than deleted, unlike the prover's visible(),
    because this is used to compare an original page against its assembled
    successor: &amp; on one side and & on the other is not a content change, and
    a comparison that says it is would bury the real losses in noise.
    """
    t = RE_STYLE_BLOCKS.sub(" ", s)
    t = RE_SCRIPT_BLOCKS.sub(" ", t)
    t = RE_HTML_TAGS.sub(" ", t)
    t = htmlmod.unescape(t)
    return RE_WHITESPACE.sub(" ", t).strip()
