#!/usr/bin/env python3
"""Render the gauntlet component exactly as the chassis would, into a local
page, so the real template + real JS can be driven against the real API before
anything is delivered to production."""
import json, re, pathlib

here = pathlib.Path(__file__).parent
tmpl = (here / "gauntlet_new.html").read_text()
data = json.loads((here / "field_updates.json").read_text())

missing = []


def sub(m):
    k = m.group(1)
    if k not in data:
        missing.append(k)
        return "<no value>"
    return str(data[k])


rendered = re.sub(r"\{\{\.(\w+)\}\}", sub, tmpl)
if missing:
    raise SystemExit("template keys absent from content_data: " + ", ".join(sorted(set(missing))))

# The chassis serves the component's JS as a sibling asset; mirror that path.
page = f"""<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Gauntlet local harness</title>
<style>
  :root {{
    --color-primary: #dc2626; --color-primary-text: #ffffff; --color-accent: #fbbf24;
    --color-background: #0b0b0d; --border-radius: 6px; --container-max-width: 1200px;
  }}
  body {{ margin: 0; background: #0b0b0d; font-family: system-ui, sans-serif; }}
</style>
</head><body>
{rendered}
</body></html>
"""
(here / "test.html").write_text(page)

# Serve the component JS at the same path the live page uses.
assets = here / "tools" / "assets"
assets.mkdir(parents=True, exist_ok=True)
(assets / "gauntlet-interface.js").write_text((here / "gauntlet_new.js").read_text())

datadir = here / "data"
datadir.mkdir(exist_ok=True)
(datadir / "provocations.json").write_text((here / "provocations.json").read_text())

print("rendered test.html (%d bytes), assets in place" % len(page))
