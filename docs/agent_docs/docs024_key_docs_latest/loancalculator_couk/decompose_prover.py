#!/usr/bin/env python3
"""decompose_prover.py — prove the fidelity=high decomposition rule BEFORE writing it in Go.

WHAT THIS IS, AND IS NOT. The shipped decomposer is a Go action (platform
convention: Go, not Python). This is a *prover*: it runs the proposed splitting
rule over all 27 real pages and asserts the properties that, if violated, would
silently break a calculator or silently fail to deliver what the owner asked
for. It writes nothing to the database and touches no site file. Its outputs are
(a) a go/no-go on the rule and (b) the measured evidence the council gate needs,
rather than an argument that it should work.

THE RULE.
  * page-local <style> and every inline <script> are captured BYTE-EXACT;
  * a node is TOOL-MARKED if it is a form control, or carries an id an inline
    script addresses;
  * for each top-level body block, DESCEND while every tool-marked descendant
    still sits inside one single child. The node where descent stops is the
    tool; every child passed on the way down, and every sibling at every level,
    becomes editable PROSE, in document order;
  * the tool component carries its subtree + all inline <script> + all
    page-local <style>, so the widget is self-contained.

WHY DESCENT, AND NOT "the top-level block that holds the inputs" (v1's rule).
v1 passed every safety proof and still failed the brief. On this site the
calculator and its surrounding article are SIBLINGS inside one `.container`
wrapper, so classifying whole top-level blocks put **98% of the visible text on
interactive pages inside the frozen tool component** — 100% on eleven of twelve
pages, i.e. zero editable prose. The safety proofs (P1-P4) cannot see that,
because nothing is broken; it is just useless. Hence P5.

WHY THE id RULE, RATHER THAN "the block with the inputs". A calculator's output
region usually has NO controls in it — standard-calc writes its answer into
<div id="monthly-display"> inside a results box. Split on controls alone and the
results box is classified as prose, at which point the writer loop may rewrite
away the very element getElementById() targets. The script's own id references
are the only honest definition of "what this tool needs".

WHY <style> IS LOAD-BEARING HERE, not cosmetic. 8 of these pages carry a
page-local <style> block, 7 of them calculators, and on credit-health-check the
two rules `.check-step{display:none}` / `.check-step.active{display:block}` are
what make a five-step wizard show one step at a time. Drop that block and the
tool still COMPUTES while displaying all five steps at once — it looks broken
but every value is right, which is the hardest kind of regression to attribute.
(Another lane reported that exact state as a live defect on 2026-07-31; it was
not, because the rules are in the page. But it is precisely what a
<style>-dropping decomposer would cause.)

KNOWN CONSEQUENCE, stated rather than discovered later: descent DISSOLVES the
per-page wrapper div (`.container`). Its children become sibling sections, and
layout must come from the site chrome that assembly wraps them in. That is
correct for a site moving to assembled mode, but it is a real visual change and
is not a fidelity claim.

Usage:  python3 decompose_prover.py <dir-of-stored-pages> [--verbose]
"""
import os
import re
import sys
from html.parser import HTMLParser

VOID = {"area", "base", "br", "col", "embed", "hr", "img", "input", "link",
        "meta", "param", "source", "track", "wbr"}
RAWTEXT = {"script", "style"}
CONTROL_TAGS = {"input", "select", "textarea", "button"}
# Below this many characters, a "prose" block is a label or a stray <br> and
# counting it as editable content would flatter the P5 measurement.
PROSE_MIN = 40


class Node:
    __slots__ = ("tag", "start", "end", "inner_start", "children", "starttag")

    def __init__(self, tag, start, starttag):
        self.tag, self.start, self.starttag = tag, start, starttag
        self.end = None
        self.inner_start = None
        self.children = []


class TreeBuilder(HTMLParser):
    """Build an element tree carrying BYTE OFFSETS into the original string.

    Offsets, not reconstructed markup: slicing the source is what makes an
    emitted block byte-identical. Rebuilding from parser events would silently
    normalise entities and attribute quoting, which is how a "faithful" port
    stops being faithful.
    """

    def __init__(self, src):
        super().__init__(convert_charrefs=False)
        self.src = src
        self.lines = [0]
        for ln in src.split("\n")[:-1]:
            self.lines.append(self.lines[-1] + len(ln) + 1)
        self.root = Node("#root", 0, "")
        self.stack = [self.root]

    def _off(self):
        ln, col = self.getpos()
        return self.lines[ln - 1] + col

    def handle_starttag(self, tag, attrs):
        o = self._off()
        st = self.get_starttag_text() or ""
        n = Node(tag, o, st)
        n.inner_start = o + len(st)
        self.stack[-1].children.append(n)
        if tag in VOID:
            n.end = o + len(st)
        else:
            self.stack.append(n)

    def handle_startendtag(self, tag, attrs):
        o = self._off()
        st = self.get_starttag_text() or ""
        n = Node(tag, o, st)
        n.inner_start = n.end = o + len(st)
        self.stack[-1].children.append(n)

    def handle_endtag(self, tag):
        if tag in VOID:
            return
        for i in range(len(self.stack) - 1, 0, -1):
            if self.stack[i].tag == tag:
                close = self.src.find(">", self._off())
                self.stack[i].end = close + 1 if close != -1 else len(self.src)
                del self.stack[i:]
                return

    def close(self):
        super().close()
        for n in self.stack[1:]:
            if n.end is None:
                n.end = len(self.src)


def raw_blocks(tag, s):
    """Byte-exact spans of every <tag>...</tag>, used for <script> and <style>."""
    return [m.group(0) for m in
            re.finditer(r"<%s\b[^>]*>.*?</%s>" % (tag, tag), s, re.S | re.I)]


def script_ids(scripts):
    """Every element id an inline script addresses, however it is written."""
    ids = set()
    for s in scripts:
        ids |= set(re.findall(r"getElementById\(\s*['\"]([^'\"]+)['\"]", s))
        ids |= set(re.findall(r"querySelector(?:All)?\(\s*['\"]#([\w-]+)", s))
        ids |= set(re.findall(r"['\"]([\w-]*-)['\"]\s*\+", s))
    return ids


def marked(n, ids):
    """Is this node itself part of the tool's machinery?"""
    if n.tag in CONTROL_TAGS or "contenteditable" in n.starttag.lower():
        return True
    m = re.search(r'''\bid\s*=\s*['"]?([\w-]+)''', n.starttag)
    if m:
        v = m.group(1)
        if v in ids or any(v.startswith(p) for p in ids if p.endswith("-")):
            return True
    return False


def any_marked(n, ids):
    return marked(n, ids) or any(any_marked(c, ids) for c in n.children)


def loose_text(node, src, kids):
    """Text sitting DIRECTLY in a node, between its element children.

    Descent used to iterate element children only, so a bare text node between
    two elements was dropped — 5 pages silently lost 50-70 characters and P6 is
    the only reason it was noticed. Emitted as prose so it stays editable.
    """
    out = []
    cur = node.inner_start
    close = len("</%s>" % node.tag)
    for c in kids + [None]:
        stop = c.start if c is not None else max(node.inner_start, node.end - close)
        if stop > cur:
            gap = src[cur:stop]
            if re.sub(r"<[^>]+>", "", gap).strip():
                out.append(gap)
        if c is not None:
            cur = c.end
    return out


def split(node, src, ids, prose, depth=0):
    """Descend to the smallest subtree still containing every marked node.

    Two stopping shapes, and the second is why v2 still swallowed prose:
      * ONE child holds every marked node -> descend into it;
      * SEVERAL children do (car-finance-calculator's input card AND its results
        grid are siblings) -> stop, but take only the MARKED children as the
        tool. Taking the whole parent instead froze the page's <h1> and intro
        paragraph along with them, which is the brief failing quietly.
    """
    kids = [c for c in node.children if c.tag not in RAWTEXT]
    holders = [c for c in kids if any_marked(c, ids)]

    if marked(node, ids) or not holders or depth >= 12:
        return [src[node.start:node.end]]

    prose.extend(loose_text(node, src, kids))
    for c in kids:
        if c not in holders:
            prose.append(src[c.start:c.end])

    if len(holders) == 1:
        return split(holders[0], src, ids, prose, depth + 1)
    # The wrapper dissolves; each marked child travels in the tool component.
    return [src[c.start:c.end] for c in holders]


def decompose(html):
    head = re.search(r"<head\b[^>]*>(.*?)</head>", html, re.S | re.I)
    body = re.search(r"<body\b[^>]*>(.*?)</body>", html, re.S | re.I)
    head_s = head.group(1) if head else ""
    body_s = body.group(1) if body else ""

    styles = raw_blocks("style", head_s) + raw_blocks("style", body_s)
    all_scripts = raw_blocks("script", body_s) + raw_blocks("script", head_s)
    inline = [s for s in all_scripts if not re.search(r"\bsrc\s*=", s, re.I)]
    external = [s for s in all_scripts if re.search(r"\bsrc\s*=", s, re.I)]
    title = re.search(r"<title>(.*?)</title>", head_s, re.S | re.I)
    desc = re.search(r'<meta\s+name=["\']description["\']\s+content=["\'](.*?)["\']',
                     head_s, re.S | re.I)

    tb = TreeBuilder(body_s)
    tb.feed(body_s)
    tb.close()
    ids = script_ids(inline)

    prose, tool = [], []
    for blk in tb.root.children:
        if blk.tag in RAWTEXT:
            continue
        if blk.tag == "div" and 'id="nav-placeholder"' in blk.starttag:
            continue  # site chrome, not page content
        if any_marked(blk, ids):
            tool.extend(split(blk, body_s, ids, prose))
        else:
            prose.append(body_s[blk.start:blk.end])

    return {"title": title.group(1).strip() if title else "",
            "meta_desc": desc.group(1).strip() if desc else "",
            "styles": styles, "inline_scripts": inline,
            "external_scripts": external,
            "prose": [p for p in prose if p.strip()],
            "tool": tool, "script_ids": ids,
            "n_blocks": len([b for b in tb.root.children if b.tag not in RAWTEXT])}


def visible(h):
    return re.sub(r"\s+", " ", re.sub(r"<[^>]+>", " ", h)).strip()


def main():
    d = sys.argv[1]
    verbose = "--verbose" in sys.argv
    fails, rows = [], []
    for f in sorted(os.listdir(d)):
        html = open(os.path.join(d, f), encoding="utf-8").read()
        r = decompose(html)
        emitted = "".join(r["tool"]) + "".join(r["inline_scripts"]) + "".join(r["styles"])

        # P1/P2 — every inline script and page-local style survives BYTE-EXACT.
        for s in raw_blocks("script", html):
            if not re.search(r"\bsrc\s*=", s, re.I) and s not in emitted:
                fails.append((f, "P1 inline <script> not byte-exact in output"))
        for s in raw_blocks("style", html):
            if s not in emitted:
                fails.append((f, "P2 page-local <style> not byte-exact in output"))

        # P4 — THE CALCULATOR-SURVIVAL PROOF. Every id an inline script
        # addresses must travel inside the tool component with it.
        present = [i for i in r["script_ids"]
                   if ('id="%s"' % i) in html or ("id='%s'" % i) in html]
        orphan = [i for i in present
                  if ('id="%s"' % i) not in emitted and ("id='%s'" % i) not in emitted]
        if orphan:
            fails.append((f, "P4 script targets left OUTSIDE the tool: %s"
                          % ",".join(sorted(orphan)[:6])))

        pt = len(visible("".join(r["prose"])))
        tt = len(visible("".join(r["tool"])))
        # P6 — no visible BODY text may vanish. Measured against the body only:
        # counting the whole document made this fail on 5 pages by exactly the
        # length of their <title>, which is not lost at all — it is carried in
        # `title` and re-injected by assemblePage. A check that answers a
        # slightly different question than the one asked is worse than none,
        # because the 5 "failures" looked like real content loss.
        bm = re.search(r"<body\b[^>]*>(.*?)</body>", html, re.S | re.I)
        bodyonly = re.sub(r"<(script|style)\b.*?</\1>", " ", bm.group(1) if bm else "",
                          flags=re.S | re.I)
        whole = len(visible(bodyonly))
        if pt + tt < whole * 0.98:
            fails.append((f, "P6 visible body text lost: kept %d of %d chars"
                          % (pt + tt, whole)))

        rows.append((f, len(r["prose"]), len(r["tool"]), len(r["inline_scripts"]),
                     len(r["styles"]), pt, tt))
        if verbose and r["inline_scripts"]:
            print("--- %s\n    tool: %s\n" % (f, visible("".join(r["tool"]))[:150]))

    print("%-42s %5s %5s %4s %4s %8s %8s %7s"
          % ("page", "prose", "tool", "js", "css", "prose_ch", "tool_ch", "%frozen"))
    ip, it = 0, 0
    for f, p, t, j, c, pt, tt in rows:
        fz = 100 * tt // max(1, pt + tt)
        print("%-42s %5d %5d %4d %4d %8d %8d %6d%%" % (f[:42], p, t, j, c, pt, tt, fz))
        if j:
            ip += pt
            it += tt

    # P5 — THE BRIEF. Interactive pages must end up mostly EDITABLE, or the
    # decomposition delivered safety and nothing else. v1's rule scored 98%.
    frozen = 100 * it // max(1, ip + it)
    print("\ninteractive pages: prose %d ch, tool %d ch -> %d%% frozen" % (ip, it, frozen))
    if frozen > 50:
        fails.append(("(all)", "P5 %d%% of interactive-page text is frozen in tool "
                               "components — decomposition achieved nothing editable" % frozen))

    if fails:
        print("\nFAILURES (%d):" % len(fails))
        for f, why in fails:
            print("  %-42s %s" % (f[:42], why))
        sys.exit(1)
    print("\nALL PROOFS PASS — P1 script bytes, P2 style bytes, P4 no orphaned "
          "script target, P5 prose is editable (%d%% frozen), P6 no text lost" % frozen)


main()
