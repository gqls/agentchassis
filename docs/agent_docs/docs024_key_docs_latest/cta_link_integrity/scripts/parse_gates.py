#!/usr/bin/env python3
"""Real template parse for bugs_open/023 P2.1.

Replaces RUNBOOK R9's 60-char lookback heuristic. Tokenises each active
component template into text + template actions, maintains a block stack
({{if}}/{{range}}/{{with}} .. {{end}}), and for every <a href="{{.X}}">
records the enclosing conditions. An anchor is GATED iff some enclosing
block's condition references the same field X.
"""
import json, re, sys, collections

ACTION = re.compile(r'\{\{-?\s*(.*?)\s*-?\}\}', re.S)
# href="{{.field}}" possibly with spaces inside the action
HREF = re.compile(r'<a\b[^>]*?href="\s*\{\{-?\s*\.([A-Za-z0-9_]+)\s*-?\}\}\s*"', re.S)
FIELDREF = re.compile(r'\.([A-Za-z0-9_]+)')

def blocks_for_offsets(tpl):
    """Return list of (start,end,cond) for every if/range/with block."""
    stack, blocks = [], []
    for m in ACTION.finditer(tpl):
        body = m.group(1)
        head = body.split()[0] if body.split() else ''
        if head in ('if', 'range', 'with'):
            stack.append((m.start(), body))
        elif head == 'end':
            if stack:
                start, cond = stack.pop()
                blocks.append((start, m.end(), cond))
        elif head in ('else',):
            pass
    return blocks

def main(path):
    comps = json.load(open(path))
    rows = []
    for c in comps:
        tpl = c.get('tpl') or ''
        blocks = blocks_for_offsets(tpl)
        for m in HREF.finditer(tpl):
            field = m.group(1)
            pos = m.start()
            enclosing = [cond for (s, e, cond) in blocks if s < pos < e]
            gated_by = [cond for cond in enclosing
                        if field in FIELDREF.findall(cond)]
            rows.append({
                'component': c['name'],
                'function': c.get('function'),
                'field': field,
                'offset': pos,
                'gated': bool(gated_by),
                'gated_by': gated_by[0] if gated_by else None,
                'enclosing': enclosing,
                'anchor': tpl[m.start():m.start()+160].replace('\n', ' '),
            })
    json.dump(rows, open('parsed_anchors.json', 'w'), indent=1)

    ung = [r for r in rows if not r['gated']]
    gat = [r for r in rows if r['gated']]
    print(f"TOTAL href=\"{{{{.X}}}}\" anchors : {len(rows)}")
    print(f"  GATED   : {len(gat):3d} anchors / {len(set(r['component'] for r in gat))} components")
    print(f"  UNGATED : {len(ung):3d} anchors / {len(set(r['component'] for r in ung))} components")
    print()

    # An anchor inside a {{range}} is an ITEM link (the field is a property of
    # the ranged item, e.g. {{range .items}}<a href="{{.url}}">), not a CTA
    # label/url pair. Different class, different fix, different owner — the CTA
    # worklist is the NOT-in-range set. Counting them together overstates P2.1
    # and, in a migration post-condition, a blanket "no ungated url anchor"
    # regex will trip on them and roll back a correct change.
    in_range = [r for r in ung
                if any(c.split()[:1] == ['range'] for c in r['enclosing'])]
    not_range = [r for r in ung if r not in in_range]
    print(f"  of the UNGATED:")
    print(f"    inside {{{{range}}}} (item links, separate class) : {len(in_range):3d} "
          f"/ {len(set(r['component'] for r in in_range))} components "
          f"-> fields {sorted(set(r['field'] for r in in_range))}")
    print(f"    NOT in a range (the CTA worklist)              : {len(not_range):3d} "
          f"/ {len(set(r['component'] for r in not_range))} components")
    print()
    # url-suffixed only (the CTA class R9 measured)
    ung_url = [r for r in ung if r['field'].endswith('url')]
    print(f"UNGATED with *_url field : {len(ung_url)} anchors / "
          f"{len(set(r['component'] for r in ung_url))} components")
    ung_nonurl = [r for r in ung if not r['field'].endswith('url')]
    print(f"UNGATED non-url field    : {len(ung_nonurl)} anchors / "
          f"{len(set(r['component'] for r in ung_nonurl))} components "
          f"-> fields: {sorted(set(r['field'] for r in ung_nonurl))}")
    print()
    print("Per-component UNGATED counts:")
    cnt = collections.Counter(r['component'] for r in ung)
    for name, n in sorted(cnt.items(), key=lambda kv: (-kv[1], kv[0])):
        fields = sorted(set(r['field'] for r in ung if r['component'] == name))
        print(f"  {n:2d}  {name:42s} {','.join(fields)}")

if __name__ == '__main__':
    main(sys.argv[1])
