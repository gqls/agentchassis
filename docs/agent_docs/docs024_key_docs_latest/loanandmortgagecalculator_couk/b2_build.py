#!/usr/bin/env python3
"""b2_build.py — turn each page's tool block into a parameterised component seed.

TRACK B2 (bugs_open/263, 2026-08-13 entry; owner: "all text and widgets need to be
editable so in future we can reuse them"). Per page this produces:

  html_template   the tool block with every clean copy span replaced by {{.field}}
                  — machinery (ids, scripts, attributes, options) stays literal
  input_schema    one field per span: required, fallback = the original copy,
                  llm_guidance per kind (labels carry the load-bearing warning)
  content_data    the original copy, so the first render is the current page
  render check    EVERY page's template is rendered with Go's own text/template
                  (missingkey=zero, the engine call_agent.go actually uses) and
                  must equal the manifest's tool block BYTE FOR BYTE, or the page
                  is refused. A python re-implementation of the engine would be
                  a second opinion agreeing with itself — this is the real one.

WHAT COUNTS AS A COPY SPAN (v1, deliberately conservative):
  <h1..h4>text</h>      heading fields
  <label>text</label>   label fields (load-bearing warning)
  <button ...>text</    button fields
  <a ...>text</a>       link-text fields (href stays in the template)
  <p ...>text</p>       advisory fields — only when the paragraph holds no element

A span is only extracted when its inner text is TAG-FREE and appears at that exact
byte sequence a countable number of times (occurrence-indexed replacement keeps
duplicates safe). Anything else stays literal in the template: that copy is not yet
field-editable, and the per-page report SAYS SO rather than letting an unextracted
span read as extracted. v1 leaves <select> options, placeholder attributes and
value attributes as machinery on purpose — scripts read them.

Usage:
  python3 b2_build.py <manifest.json> <outdir> [--pages a,b,c]
Exit 1 if any page in scope fails its render check.
"""
import json
import os
import re
import subprocess
import sys
import tempfile

LABEL_WARNING = ("LOAD-BEARING COPY: this is the visible label a reader uses to decide "
                 "what number to type, and the arithmetic oracle asserts against it. "
                 "Changing what it MEANS (not just its wording) silently changes what "
                 "users believe the calculator computes. Reword for voice only; never "
                 "change the quantity it names, its unit, or its currency marker.")
GUIDANCE = {
    "heading": "A heading inside the calculator panel. Name what this part of the tool "
               "does in plain words. No negation, no aphorism, sentence case.",
    "label": LABEL_WARNING,
    "button": "Button text. An imperative verb is correct here — buttons are "
              "instructions. One or two words.",
    "link": "Link text. Say what the reader gets at the destination. Keep any trailing "
            "arrow entity. The destination itself is not yours to change.",
    "para": "Advisory copy inside the calculator panel. Plain UK English, framed as "
            "something the reader can act on. Keep every figure and factual claim "
            "unless the brief for the edit says otherwise.",
}

SPAN_RES = [
    ("heading", re.compile(r"<h([1-4])(\s[^>]*)?>([^<]{2,120})</h\1>")),
    ("label",   re.compile(r"<label(\s[^>]*)?>([^<]{1,120})</label>")),
    ("button",  re.compile(r"<button\s[^>]*>([^<]{1,60})</button>")),
    ("link",    re.compile(r"<a\s[^>]*>([^<]{2,120})</a>")),
    ("para",    re.compile(r"<p(\s[^>]*)?>([^<]{40,700})</p>")),
]

GO_RUNNER = r"""package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

func main() {
	tmplStr, _ := os.ReadFile(os.Args[1])
	var data map[string]interface{}
	raw, _ := os.ReadFile(os.Args[2])
	json.Unmarshal(raw, &data)
	t, err := template.New("component").Option("missingkey=zero").Parse(string(tmplStr))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := t.Execute(os.Stdout, data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
"""


def extract(block):
    """Return (template, fields, schema, skipped). Occurrence-indexed replacement:
    each chosen span is replaced at its exact position, never by global replace."""
    fields, schema, skipped = {}, {"fields": {}}, []
    matches = []
    for kind, rx in SPAN_RES:
        for m in rx.finditer(block):
            text = m.group(m.lastindex)
            if "{{" in text or "}}" in text:
                skipped.append((kind, text[:40], "brace"))
                continue
            # the span's inner text must sit inside THIS match only once
            inner_start = m.start() + m.group(0).index(text)
            matches.append((inner_start, len(text), kind, text))
    # de-overlap (a <p> containing only an <a> matches both): keep the innermost/first
    matches.sort()
    chosen, last_end = [], -1
    for start, ln, kind, text in matches:
        if start < last_end:
            skipped.append((kind, text[:40], "overlap"))
            continue
        chosen.append((start, ln, kind, text))
        last_end = start + ln
    counters, out, pos = {}, [], 0
    for start, ln, kind, text in chosen:
        counters[kind] = counters.get(kind, 0) + 1
        name = "%s_%d" % (kind, counters[kind])
        out.append(block[pos:start])
        out.append("{{.%s}}" % name)
        pos = start + ln
        fields[name] = text
        schema["fields"][name] = {"type": "text", "source": "llm", "fallback": text,
                                  "required": True, "llm_guidance": GUIDANCE[kind]}
    out.append(block[pos:])
    return "".join(out), fields, schema, skipped


def go_render(tmpl, fields, workdir):
    tp = os.path.join(workdir, "t.html")
    fp = os.path.join(workdir, "f.json")
    open(tp, "w").write(tmpl)
    json.dump(fields, open(fp, "w"), ensure_ascii=False)
    r = subprocess.run(["go", "run", "render.go", tp, fp],
                       capture_output=True, text=True, cwd=workdir)
    if r.returncode != 0:
        return None, r.stderr.strip()[:200]
    return r.stdout, None


def main():
    man_path, outdir = sys.argv[1], sys.argv[2]
    only = None
    if "--pages" in sys.argv:
        only = set(sys.argv[sys.argv.index("--pages") + 1].split(","))
    man = json.load(open(man_path))
    os.makedirs(outdir, exist_ok=True)

    work = tempfile.mkdtemp(prefix="b2render")
    open(os.path.join(work, "render.go"), "w").write(GO_RUNNER)
    open(os.path.join(work, "go.mod"), "w").write("module b2render\n\ngo 1.23\n")

    failures = 0
    for name, entry in sorted(man["pages"].items()):
        if only and name not in only:
            continue
        tools = [b for b in entry["blocks"] if b["kind"] == "tool"]
        if len(tools) != 1:
            print("REFUSE %-36s %d tool blocks (expected 1)" % (name, len(tools)))
            failures += 1
            continue
        # THE SCRIPTS KEY IS LOAD-BEARING AND THE FIRST VERSION OF THIS FILE DROPPED IT.
        # decompose_lmc puts body-level inline scripts and the calculators.js tag in a
        # separate b["scripts"], and load_lmc.py:240 appends it to the row's html. This
        # file read only b["html"], so 10 of 15 batch-1 pages shipped WITHOUT their
        # scripts — dead calculators live (2026-08-13, restored same hour). Scripts are
        # machinery: they join the template LITERALLY, after field extraction, exactly
        # as load_lmc joins them ("\n" separator), and the render check below now
        # asserts against html+scripts so this cannot silently regress.
        block = tools[0]["html"]
        if tools[0].get("scripts"):
            block = block + "\n" + tools[0]["scripts"]
        html_only_len = len(tools[0]["html"])
        tmpl_head, fields, schema, skipped = extract(block[:html_only_len])
        tmpl = tmpl_head + block[html_only_len:]
        rendered, err = go_render(tmpl, fields, work)
        if err or rendered != block:
            print("FAIL   %-36s render check: %s" % (name, err or "differs from block"))
            failures += 1
            continue
        json.dump({"page": name, "pinned_ref": man["pinned_ref"],
                   "template": tmpl, "fields": fields, "schema": schema,
                   "prose_blocks": [b["html"] for b in entry["blocks"] if b["kind"] == "prose"],
                   "block_order": [b["kind"] for b in entry["blocks"]],
                   "unextracted": ["%s: %s (%s)" % s for s in skipped]},
                  open(os.path.join(outdir, name + ".json"), "w"), ensure_ascii=False)
        print("ok     %-36s %2d fields  render==block  %s" %
              (name, len(fields),
               ("(%d span(s) left literal)" % len(skipped)) if skipped else ""))
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
