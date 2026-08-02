#!/usr/bin/env python3
"""decompose_pages.py — turn each stored verbatim document into an ORDERED manifest.

RELATIONSHIP TO decompose_prover.py. The prover established that the splitting
rule is safe (P1 script bytes, P2 style bytes, P4 no orphaned script target,
P5 the prose is actually editable, P6 no visible text lost) over all 27 real
pages. It answers "may we split this way?". It deliberately does NOT answer
"what rows do we write?", and it throws away the one thing the writer needs:

    THE PROVER LOSES DOCUMENT ORDER BETWEEN PROSE AND TOOL.

It returns two lists, `prose` and `tool`, each internally ordered and mutually
unordered. That is fine for proving text is not lost and scripts are not
orphaned — neither property depends on interleaving — and useless for emitting
page_components rows, where `position` IS the page. So this module re-walks the
same tree with the same rule and emits ONE ordered list of blocks instead.

The rule itself is unchanged and deliberately not re-stated here; read the
prover's docstring for why descent, why the id rule, and why <style> is
load-bearing. Where this file differs from the prover, it is because writing
rows needs a decision the prover never had to make:

  * TOOL SUBSTITUTION. The prover emits the ORIGINAL widget markup as the tool
    block. We do not ship that. The eleven components in content_components are
    proven numerically identical to those widgets across three input vectors
    (verify_rewrite.py) and are self-contained, so a page's contiguous run of
    tool blocks collapses to a single row pointing at the component. The
    original markup is still emitted, as `replaced_html`, so a reviewer can
    diff what was dropped rather than take it on trust.

  * PAGE-LOCAL <style> PLACEMENT. Eight pages carry one. In the original it sat
    in <head>, i.e. before every body rule in source order. Assembly has no
    per-page head, so it has to ride in a section — and WHERE it rides changes
    the cascade, because several of these rules are ID-based (#meter-fill,
    #progress-fill) and an id selector beats the component's own class-scoped
    rule. Equal-specificity ties then fall to source order. So the style block
    is prepended to the page's FIRST block, which reproduces the original
    ordering, and `first_block_is_tool` is reported per page so that the
    assumption is checked rather than assumed. (It is false on all 27 today.)

  * EXTERNAL SCRIPTS ARE READ, then dropped and named. This is a CORRECTION to
    the prover, not an embellishment of it, and it is the reason this file
    takes an --assets directory.

    The prover derives "what the tool needs" from the ids that INLINE scripts
    address, and that definition is right. Its INPUT was incomplete: it never
    opened the <script src> files, so a page whose logic lives in an external
    file has script targets the rule cannot see. On /index.html the whole of
    calculateLoan() lives in /assets/js/global.js, so the three ids it writes —
    monthly-display, total-interest, total-cost — were invisible, and the
    calculator's entire RESULTS BOX was classified as editable prose. It would
    have shipped as a second, dead results box beside the component's live one.

    P4 did not catch it, and could not: P4 asks whether every id an inline
    script addresses travels with the tool, and on that page the inline script
    addresses only the three inputs, all of which did travel. The proof passed
    because the question was narrower than the risk. So external scripts are
    now read from --assets and folded into the same id set, and a referenced
    script that cannot be read is a hard failure rather than a shrug — silently
    proceeding on a missing file is how this was introduced.

    They are still DROPPED from the output, and named. /assets/js/nav.js
    injects the site nav client-side into #nav-placeholder; assembly emits a
    server-side header instead, so carrying it would render the nav twice.
    global.js is replaced by the component's own arithmetic.

Usage:  python3 decompose_pages.py <stored-dir> <out.json> --assets <dir> [--verbose]

  <assets-dir> mirrors the site root, so /assets/js/global.js is read from
  <assets-dir>/assets/js/global.js.
"""
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from decompose_prover import (  # noqa: E402  (path set above)
    RAWTEXT, TreeBuilder, loose_text, marked, any_marked, raw_blocks,
    script_ids, visible,
)

# page name -> the content_components.function that replaces its widget.
# index and standard-calc are ONE calculator: identical steps and identical
# values across all three vectors, which is why there are eleven components for
# twelve interactive pages. tool-credit-roadmap is absent deliberately — it has
# no controls and no script, so it decomposes to pure prose.
TOOL_FOR_PAGE = {
    "index":                         "tool-loan-repayment",
    "tool-standard-calc":            "tool-loan-repayment",
    "tool-credit-health-check":      "tool-credit-health-check",
    "tool-interest-rate-stress-test": "tool-rate-stress-test",
    "tool-settlement-calculator":    "tool-early-settlement",
    "tool-overpayment-calculator":   "tool-overpayment-impact",
    "tool-loan-vs-savings":          "tool-loan-vs-savings",
    "tool-damage-checker":           "tool-return-damage-checker",
    "tool-compare-loans":            "tool-compare-loan-offers",
    "tool-car-finance-calculator":   "tool-car-finance-pcp-hp",
    "tool-consolidation":            "tool-consolidation-risk",
    "tool-application-tracker":      "tool-application-tracker",
}


def split_ordered(node, src, ids, out, depth=0):
    """Descent, identical to the prover's, emitting into ONE ordered list.

    `out` receives ("prose", html) and ("tool", html) tuples in document order.
    The prover's version appends prose to one list and returns tool blocks to
    another; the only change here is that both go to the same place, so a
    sibling that becomes prose keeps its position relative to the widget it
    sits beside.
    """
    kids = [c for c in node.children if c.tag not in RAWTEXT]
    holders = [c for c in kids if any_marked(c, ids)]

    if marked(node, ids) or not holders or depth >= 12:
        out.append(("tool", src[node.start:node.end]))
        return

    # Loose text sits before the first element child, so it leads.
    for t in loose_text(node, src, kids):
        out.append(("prose", t))

    for c in kids:
        if c in holders:
            if len(holders) == 1:
                split_ordered(c, src, ids, out, depth + 1)
            else:
                # Several siblings hold marked nodes (car-finance's input card
                # and its results grid). The wrapper dissolves; each marked
                # child travels as its own tool block, still in position.
                out.append(("tool", src[c.start:c.end]))
        else:
            out.append(("prose", src[c.start:c.end]))


def collapse_runs(blocks):
    """Collapse each contiguous run of tool blocks into one.

    A page has at most one calculator, so more than one run would mean the rule
    classified prose as tool somewhere in the middle. That is reported by the
    caller rather than merged away, because merging it would silently freeze
    whatever prose sat between the runs.
    """
    out, runs = [], 0
    for kind, html in blocks:
        if kind == "tool" and out and out[-1][0] == "tool":
            out[-1] = ("tool", out[-1][1] + "\n" + html)
            continue
        if kind == "tool":
            runs += 1
        out.append((kind, html))
    return out, runs


# Per-page overrides of the tool component's fields, written into that page's
# page_components.content_data.
#
# ONE ENTRY, AND IT EXISTS BECAUSE THE SCREENSHOT CAUGHT WHAT THE CHECKS DID NOT.
# tool-loan-repayment was built from /tools/standard-calc.html, which carries a
# risk warning and two market-context lines that /index.html does not. Reused on
# the homepage unchanged, the component ADDED all three.
#
# Every automated check passed it, and each for a defensible reason: the numeric
# fingerprint covers elements with ids and these three have none; the
# text-preservation check asks whether the ORIGINAL text survived, which it did.
# Nothing asked whether text had been ADDED. verify_assembled.py now does (check
# F), but the thing that actually found it was looking at the rendered page.
#
# THE FIRST FIX WAS WRONG AND THE SCHEMA REFUSED IT. Blanking all three was the
# obvious "reproduce the page as it is today" move, and render_tool.go rejected
# it: fca_warning is required:true, because a risk warning belongs alongside a
# consumer credit promotion and the field is `source: static` precisely so no
# loop can drop it. The homepage IS such a promotion, so what looked like
# faithfulness was removing a warning that ought to be there. It stays, and the
# homepage gains it — a deliberate change, flagged to the owner, not a slip.
#
# The two DATED FACTUAL CLAIMS are a different matter and are blanked: a market
# average and a base-rate note go stale, and duplicating them onto a second page
# doubles the number of places that have to be corrected when they do. Their own
# llm_guidance says as much and both are required:false for that reason.
CONTENT_DATA_FOR_PAGE = {
    "index": {"market_context": "", "rate_environment_note": ""},
}


def read_external(paths, assets_dir, name):
    """Read every <script src> the page loads, from a local mirror of the site.

    Hard-fails on a missing file. The alternative — carry on with the ids it
    could see — is precisely the failure this function exists to remove, and it
    is invisible: the decomposition still succeeds, still passes every proof,
    and quietly reclassifies part of a calculator as prose.
    """
    out = []
    for p in paths:
        if p.startswith(("http://", "https://", "//")):
            raise SystemExit("%s: external script %s is off-site; mirror it into "
                             "--assets before decomposing" % (name, p))
        local = os.path.join(assets_dir, p.lstrip("/"))
        if not os.path.exists(local):
            raise SystemExit("%s: references %s and it is not in --assets (%s). "
                             "Its getElementById targets would be invisible to the "
                             "splitting rule." % (name, p, local))
        out.append(open(local, encoding="utf-8", errors="replace").read())
    return out


def decompose_page(name, html, assets_dir):
    head = re.search(r"<head\b[^>]*>(.*?)</head>", html, re.S | re.I)
    body = re.search(r"<body\b[^>]*>(.*?)</body>", html, re.S | re.I)
    head_s = head.group(1) if head else ""
    body_s = body.group(1) if body else ""

    styles = raw_blocks("style", head_s) + raw_blocks("style", body_s)
    all_scripts = raw_blocks("script", body_s) + raw_blocks("script", head_s)
    inline = [s for s in all_scripts if not re.search(r"\bsrc\s*=", s, re.I)]
    external = [re.search(r'src\s*=\s*["\']([^"\']+)', s, re.I).group(1)
                for s in all_scripts if re.search(r"\bsrc\s*=", s, re.I)]
    title = re.search(r"<title>(.*?)</title>", head_s, re.S | re.I)
    desc = re.search(r'<meta\s+name=["\']description["\']\s+content=["\'](.*?)["\']',
                     head_s, re.S | re.I)

    tb = TreeBuilder(body_s)
    tb.feed(body_s)
    tb.close()
    # Inline AND external: see the module docstring. The rule is the prover's;
    # only the set of scripts it is allowed to read has widened.
    ids = script_ids(inline + read_external(external, assets_dir, name))

    ordered = []
    for blk in tb.root.children:
        if blk.tag in RAWTEXT:
            continue
        if blk.tag == "div" and 'id="nav-placeholder"' in blk.starttag:
            continue  # site chrome, not page content
        if any_marked(blk, ids):
            split_ordered(blk, body_s, ids, ordered)
        else:
            ordered.append(("prose", body_s[blk.start:blk.end]))

    ordered = [(k, h) for k, h in ordered if h.strip()]
    ordered, tool_runs = collapse_runs(ordered)

    # P4, re-asked against the WIDENED id set and against PROSE specifically.
    # The prover asks "is every script target inside the emitted output?", which
    # an id sitting in a prose block still satisfies. The question that matters
    # when writing rows is narrower: no script target may land in prose, because
    # prose is the part a writer agent is free to rewrite.
    prose_html = "".join(h for k, h in ordered if k == "prose")
    stranded = sorted(i for i in ids
                      if ('id="%s"' % i) in prose_html or ("id='%s'" % i) in prose_html)

    fn = TOOL_FOR_PAGE.get(name)
    blocks = []
    for kind, h in ordered:
        if kind == "tool":
            if not fn:
                # A widget with no replacement component would be frozen markup
                # in an "editable" page — the exact outcome this work exists to
                # remove. Refuse rather than emit it.
                raise SystemExit(
                    "%s: tool block found but no component mapped for it" % name)
            blocks.append({"kind": "tool", "function": fn, "replaced_html": h})
        else:
            blocks.append({"kind": "prose", "html": h})

    # Page-local <style> leads the page, as it did in <head>. See the module
    # docstring: this is a cascade decision, not a tidiness one.
    first_is_tool = bool(blocks) and blocks[0]["kind"] == "tool"
    if styles and blocks and not first_is_tool:
        blocks[0]["html"] = "\n".join(styles) + "\n" + blocks[0]["html"]

    return {
        "name": name,
        "title": title.group(1).strip() if title else "",
        "meta_desc": desc.group(1).strip() if desc else "",
        "blocks": blocks,
        "tool_function": fn,
        "tool_content_data": CONTENT_DATA_FOR_PAGE.get(name),
        "tool_runs": tool_runs,
        "first_block_is_tool": first_is_tool,
        "page_styles": styles,
        "inline_scripts_dropped": len(inline),
        "external_scripts_dropped": external,
        "stranded_script_targets": stranded,
        # visible() strips TAGS but keeps the text inside them, so a <style>
        # block counts its own CSS as prose. Left uncorrected this metric grew
        # by exactly the length of the stylesheet on the eight pages that carry
        # one (application-tracker read 1066 editable characters against a real
        # 77), which would have flattered the one number this work is judged on.
        "prose_chars": sum(len(visible(re.sub(r"<(script|style)\b.*?</\1>", " ",
                                              b["html"], flags=re.S | re.I)))
                           for b in blocks if b["kind"] == "prose"),
    }


def main():
    src_dir, out_path = sys.argv[1], sys.argv[2]
    verbose = "--verbose" in sys.argv
    if "--assets" not in sys.argv:
        raise SystemExit("--assets <dir> is required; see the docstring for why")
    assets_dir = sys.argv[sys.argv.index("--assets") + 1]
    manifest, warnings = {}, []

    for f in sorted(os.listdir(src_dir)):
        if not f.endswith(".html"):
            continue
        name = f[:-5]
        html = open(os.path.join(src_dir, f), encoding="utf-8").read()
        page = decompose_page(name, html, assets_dir)
        manifest[name] = page

        if page["stranded_script_targets"]:
            warnings.append("%s: script target(s) left in PROSE, where a writer "
                            "agent may rewrite them: %s"
                            % (name, ",".join(page["stranded_script_targets"])))
        if page["tool_runs"] > 1:
            warnings.append("%s: %d separate tool runs — prose may sit between them"
                            % (name, page["tool_runs"]))
        if page["first_block_is_tool"] and page["page_styles"]:
            warnings.append("%s: first block is the TOOL and the page has local "
                            "<style> — the style has nowhere safe to ride" % name)
        if page["tool_function"] and not any(b["kind"] == "tool" for b in page["blocks"]):
            warnings.append("%s: mapped to %s but no tool block was found"
                            % (name, page["tool_function"]))
        if not page["tool_function"] and any(b["kind"] == "tool" for b in page["blocks"]):
            warnings.append("%s: tool block with no component mapping" % name)

    print("%-36s %6s %6s %5s %9s %s"
          % ("page", "blocks", "prose", "tool", "prose_ch", "component"))
    for name, p in manifest.items():
        nt = sum(1 for b in p["blocks"] if b["kind"] == "tool")
        npr = sum(1 for b in p["blocks"] if b["kind"] == "prose")
        print("%-36s %6d %6d %5d %9d %s"
              % (name[:36], len(p["blocks"]), npr, nt, p["prose_chars"],
                 p["tool_function"] or "-"))
        if verbose:
            for i, b in enumerate(p["blocks"]):
                tag = b["kind"].upper()
                txt = visible(b.get("html") or b.get("replaced_html", ""))[:80]
                print("      %2d %-5s %s" % (i, tag, txt))

    mapped = set(TOOL_FOR_PAGE)
    withtool = {n for n, p in manifest.items()
                if any(b["kind"] == "tool" for b in p["blocks"])}
    print("\npages: %d  with a tool: %d  mapped: %d"
          % (len(manifest), len(withtool), len(mapped)))
    if withtool != mapped:
        warnings.append("mapping/reality mismatch: only-in-map=%s only-in-pages=%s"
                        % (sorted(mapped - withtool), sorted(withtool - mapped)))

    if warnings:
        print("\nWARNINGS (%d):" % len(warnings))
        for w in warnings:
            print("  " + w)

    with open(out_path, "w", encoding="utf-8") as fh:
        json.dump(manifest, fh, indent=1, ensure_ascii=False)
    print("\nwrote %s" % out_path)
    return 1 if warnings else 0


if __name__ == "__main__":
    sys.exit(main())
