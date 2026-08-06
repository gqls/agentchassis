#!/usr/bin/env python3
"""show_blocks.py — print a page's decomposed blocks, for authoring overlays.

  DECOMP_WORK=<dir> python3 show_blocks.py <page-name> [...]
  DECOMP_WORK=<dir> python3 show_blocks.py --list

Prose blocks print in full (that is what an overlay replaces). Tool blocks
print only their index, kind and size — a tool block can never be overlaid,
so its contents are noise here, and printing widget markup into an authoring
context invites someone to edit it.
"""
import json
import os
import sys

work = os.environ.get("DECOMP_WORK")
if not work:
    sys.exit("set DECOMP_WORK")
doc = json.load(open(os.path.join(work, "manifest.json"), encoding="utf-8"))
pages = doc["pages"]

if "--list" in sys.argv:
    for n, p in sorted(pages.items()):
        kinds = "".join("T" if b["kind"] == "tool" else "P" for b in p["blocks"])
        print("%-46s %-52s %-8s %s" % (n, p["url"], kinds,
                                       "tight" if p["tight"] else ""))
    sys.exit(0)

for name in [a for a in sys.argv[1:] if not a.startswith("--")]:
    if name not in pages:
        sys.exit("no such page: %r (try --list)" % name)
    p = pages[name]
    print("=" * 78)
    print("PAGE %s   url=%s   tight=%s   tool_page=%s"
          % (name, p["url"], p["tight"], p["is_tool_page"]))
    print("TITLE %s" % p["title"])
    print("DESC  %s" % p["meta_desc"])
    for i, b in enumerate(p["blocks"]):
        if b["kind"] == "tool":
            print("\n--- block %d: TOOL (%d bytes) — NOT overlayable ---"
                  % (i, len(b["html"])))
            continue
        print("\n--- block %d: PROSE (%d bytes) ---" % (i, len(b["html"])))
        print(b["html"])
