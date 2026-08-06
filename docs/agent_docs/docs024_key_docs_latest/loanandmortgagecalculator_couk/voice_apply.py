#!/usr/bin/env python3
"""voice_apply.py — apply the gentle-explanatory voice overlays to the
decomposition manifest, with the checks that keep a rewrite honest.

INPUT:  <work>/manifest.json           (decompose_lmc.py output — original copy)
        voice_overlays/<name>.json     ({"blocks": {"<idx>": "<html>", ...}})
OUTPUT: <work>/manifest_voiced.json    (same shape; prose blocks replaced where
                                        an overlay names them; everything else
                                        byte-original)

A page with no overlay file passes through unchanged (legal.html is voice-
EXEMPT by rule 10 and must never gain one; mortgages/simple.html has nothing
to transform). Tool blocks can never be overlaid — an overlay naming a tool
block index is refused.

CHECKS, per overlaid block — the transformation moves REGISTER, never facts:

  F1  no invented figures: every numeric token in the transformed block must
      already occur somewhere in the ORIGINAL page's prose. (The approved
      sample transformations pull figures from elsewhere on the same page —
      that is allowed; a figure from nowhere is not.)
  F2  lost figures are REPORTED per block for human review, not auto-failed:
      a rewrite may legitimately drop a number whose sentence dissolved, and
      the reviewer decides. The report is the review surface; read it.
  F3  links preserved: multiset of href values, before == after, per block.
  F4  anchors preserved: multiset of id= attributes, before == after (guides
      carry heading anchors; a lost id breaks fragment links silently).
  F5  no <script>/<style> introduced; structural tags balanced.
  F6  compliance zones byte-identical: any element whose class contains
      fca-style-warning, fca-warning-box, market-context-box or status-badge
      must appear verbatim in the transformed block (voice rule 10, plus the
      market/status boxes carry dated claims that must not be paraphrased).
  F7  the voice is present: at least one contraction (rule 8) — crude, but
      censusable — and none of rule 4's banned performatives.
  F8  still visibly non-empty (assembly would silently drop an empty block).

Usage:
  DECOMP_WORK=<dir> python3 voice_apply.py [--pages name1,name2] [--dry-run]

--dry-run runs every check and writes NOTHING. Use it when several authors
are working at once: the real run rebuilds manifest_voiced.json from
manifest.json plus ALL overlays, so two concurrent writers would race on one
output file.
"""
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
OVERLAYS = os.path.join(HERE, "voice_overlays")

BALANCED = [("<div", "</div>"), ("<section", "</section>"), ("<p", "</p"),
            ("<ul", "</ul>"), ("<ol", "</ol>"), ("<table", "</table>")]
BANNED = ["arguably", "in many ways", "crucially",
          "the most important thing", "the single most"]
CONTRACTIONS = ["it's", "you're", "they'll", "they're", "that's", "what's",
                "don't", "can't", "won't", "isn't", "doesn't", "there's",
                "you'll", "we're", "here's", "didn't", "wouldn't", "aren't"]
EXEMPT_CLASS = re.compile(
    r'<(div|p|aside)\b[^>]*class="[^"]*(?:fca-style-warning|fca-warning-box|'
    r'market-context-box|status-badge)[^"]*"')
# commas only BETWEEN digits: "2022," must tokenise as 2022, or the same
# figure reads as invented-and-lost at once (caught on the first canary run —
# F1 and F2 flagged the same year, which is the contradiction signature of an
# instrument fault)
RE_NUM = re.compile(r"\d(?:[\d,]*\d)?(?:\.\d+)?")
VOICE_EXEMPT_PAGES = {"legal"}


def visible_text(s):
    t = re.sub(r"<(script|style)\b.*?</\1>", " ", s, flags=re.S | re.I)
    t = re.sub(r"<[^>]*>", " ", t)
    return re.sub(r"\s+", " ", t)


def numbers(s):
    return sorted(RE_NUM.findall(visible_text(s)))


def hrefs(s):
    return sorted(re.findall(r'href="([^"]*)"', s))


def anchors(s):
    return sorted(re.findall(r'id="([^"]*)"', s))


def exempt_spans(s):
    """Byte spans of compliance elements. Non-nested by construction on this
    site (censused); a nested one would fail the verbatim containment check
    anyway, which is the safe direction."""
    out = []
    for m in EXEMPT_CLASS.finditer(s):
        tag = m.group(1)
        depth, i = 0, m.start()
        open_re = re.compile(r"<%s\b|</%s>" % (tag, tag))
        for t in open_re.finditer(s, m.start()):
            if t.group(0).startswith("</"):
                depth -= 1
                if depth == 0:
                    out.append(s[m.start():t.end()])
                    break
            else:
                depth += 1
    return out


def check_block(name, idx, orig, new, page_prose_numbers):
    problems, notes = [], []

    if "<script" in new.lower() or "<style" in new.lower():
        problems.append("F5: overlay introduces script/style")
    low = new.lower()
    for op, cl in BALANCED:
        if low.count(op) != low.count(cl):
            problems.append("F5: unbalanced %s (%d/%d)"
                            % (op, low.count(op), low.count(cl)))

    new_nums = numbers(new)
    invented = [n for n in set(new_nums)
                if new_nums.count(n) > page_prose_numbers.count(n)]
    if invented:
        problems.append("F1: figure(s) not in the original page's prose: %s"
                        % ",".join(sorted(invented)))
    lost = [n for n in set(numbers(orig)) if n not in new_nums]
    if lost:
        notes.append("F2 review: figure(s) no longer in block %s: %s"
                     % (idx, ",".join(sorted(lost))))

    if hrefs(orig) != hrefs(new):
        problems.append("F3: hrefs differ\n      orig: %s\n      new:  %s"
                        % (hrefs(orig), hrefs(new)))
    if anchors(orig) != anchors(new):
        problems.append("F4: id anchors differ: %s -> %s"
                        % (anchors(orig), anchors(new)))

    for span in exempt_spans(orig):
        if span not in new:
            problems.append("F6: compliance element no longer byte-identical: %s…"
                            % visible_text(span)[:60])

    vt = visible_text(new).lower()
    if not any(c in vt for c in CONTRACTIONS):
        problems.append("F7: no contraction found — the register is absent")
    for b in BANNED:
        if b in vt:
            problems.append("F7: banned performative %r present" % b)

    if len(re.sub(r"\s", "", visible_text(new))) <= 10:
        problems.append("F8: block would be dropped by assembly as empty")

    return problems, notes


def main():
    work = os.environ.get("DECOMP_WORK")
    if not work:
        sys.exit("set DECOMP_WORK")
    only = None
    if "--pages" in sys.argv:
        only = set(sys.argv[sys.argv.index("--pages") + 1].split(","))

    doc = json.load(open(os.path.join(work, "manifest.json"), encoding="utf-8"))
    manifest = doc["pages"]

    all_problems, all_notes, overlaid = [], [], 0
    for name, page in sorted(manifest.items()):
        if only and name not in only:
            continue
        path = os.path.join(OVERLAYS, name + ".json")
        if not os.path.exists(path):
            continue
        if name in VOICE_EXEMPT_PAGES:
            all_problems.append("%s: is voice-EXEMPT and has an overlay" % name)
            continue
        overlay = json.load(open(path, encoding="utf-8"))["blocks"]
        page_prose = "".join(b["html"] for b in page["blocks"]
                             if b["kind"] == "prose")
        page_nums = numbers(page_prose)
        for idx_s, new_html in sorted(overlay.items(), key=lambda kv: int(kv[0])):
            idx = int(idx_s)
            if idx >= len(page["blocks"]):
                all_problems.append("%s: overlay names block %d, page has %d"
                                    % (name, idx, len(page["blocks"])))
                continue
            blk = page["blocks"][idx]
            if blk["kind"] != "prose":
                all_problems.append("%s: overlay names block %d which is a "
                                    "TOOL block" % (name, idx))
                continue
            problems, notes = check_block(name, idx, blk["html"], new_html,
                                          page_nums)
            for p in problems:
                all_problems.append("%s[%d]: %s" % (name, idx, p))
            all_notes.extend("%s: %s" % (name, n) for n in notes)
            if not problems:
                blk["html"] = new_html
                blk["voiced"] = True
        overlaid += 1

    print("pages with an overlay applied: %d" % overlaid)
    if all_notes:
        print("\nFOR HUMAN REVIEW (%d note(s)):" % len(all_notes))
        for n in all_notes:
            print("  " + n)
    if all_problems:
        print("\nREFUSING to write manifest_voiced.json (%d problem(s)):"
              % len(all_problems))
        for p in all_problems:
            print("  " + p)
        return 1

    if "--dry-run" in sys.argv:
        print("--dry-run: all checks passed, nothing written")
        return 0

    out = os.path.join(work, "manifest_voiced.json")
    with open(out, "w", encoding="utf-8") as fh:
        json.dump(doc, fh, indent=1, ensure_ascii=False)
    print("wrote %s" % out)
    return 0


if __name__ == "__main__":
    sys.exit(main())
