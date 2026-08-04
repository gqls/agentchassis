#!/usr/bin/env python3
"""Contact sheet of recent acceptance runs — the eye half of TL-035.

    contact_sheet.py                # last 8 runs, writes ~/acceptance_renders/contact_sheet.html
    contact_sheet.py --limit 20
    contact_sheet.py --out ~/somewhere/sheet.html

Why this exists. TL-035 closed the machinery half of "faults only an eye
catches": a passing acceptance run now files full-page renders as durable
s3:// URIs on its acceptance-run doc_note. But the URIs land inside a
technical note in a private bucket — no page, no digest, nothing that puts an
image in front of a person. A photograph nobody opens is worth the same as no
photograph. This script is the opening: one command, one HTML file, every
recent run's renders inline, PASSED/FAILED and skipped-check counts beside
them.

What it deliberately is NOT:
- Not a verdict. A render is a look. Nothing here asserts anything about the
  page; the sheet exists so a person can notice what no check asserts.
- Not a publisher. The sheet is a local file under $HOME. The bucket is
  private and stays private; do not wire this into a public site.

Traps this encodes (each paid for):
1. The bucket answers 401 for a key that EXISTS exactly as for one that does
   not — you cannot probe with unauthenticated HTTP. Credentials come from the
   cluster secret the adapter itself uses (personae-storage-secrets).
2. A note WITHOUT a "Rendered:" line is not a broken flag. A failing profile
   legitimately files no render, and agents other than tool-acceptance-agent
   never carry the flag (the tier-4 component acceptance path, for one). The
   sheet lists render-less runs too, greyed, so absence is visible rather
   than invisible — but read the run before calling it a defect.
3. The render photographs the page AFTER the checks have driven it (observed
   2026-08-03: the simulator's render shows the post-"Clear" empty state
   because cleared-panel-refuses-to-invent-a-number runs last). The sheet
   labels every image "state as driven by the checks, not the landing state"
   because the first person to forget that will file a false bug.
"""
import argparse
import base64
import io
import json
import os
import re
import subprocess
import sys

NS = "ai-persona-system"
PSQL = ["kubectl", "-n", NS, "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A"]
S3_HOST = "s3.us-east-005.backblazeb2.com"
S3_REGION = "us-east-005"


def psql(sql):
    r = subprocess.run(PSQL, input=sql, capture_output=True, text=True, timeout=60)
    if r.returncode != 0:
        sys.exit("psql failed: " + r.stderr[-500:])
    return r.stdout


def b2_creds():
    out = {}
    for key in ("B2_APPLICATION_KEY_ID", "B2_APPLICATION_KEY"):
        r = subprocess.run(
            ["kubectl", "-n", NS, "get", "secret", "personae-storage-secrets",
             "-o", "jsonpath={.data.%s}" % key],
            capture_output=True, text=True, timeout=30)
        if r.returncode != 0 or not r.stdout:
            sys.exit("could not read %s from personae-storage-secrets" % key)
        out[key] = base64.b64decode(r.stdout).decode()
    return out["B2_APPLICATION_KEY_ID"], out["B2_APPLICATION_KEY"]


def fetch_s3(uri, key_id, key):
    """s3://bucket/key -> PNG bytes, via curl's sigv4 (stdlib has no signer)."""
    m = re.match(r"s3://([^/]+)/(.+)", uri)
    if not m:
        return None, "unparseable uri"
    url = "https://%s.%s/%s" % (m.group(1), S3_HOST, m.group(2))
    r = subprocess.run(
        ["curl", "-sS", "--fail", "--aws-sigv4", "aws:amz:%s:s3" % S3_REGION,
         "--user", "%s:%s" % (key_id, key), url],
        capture_output=True, timeout=120)
    if r.returncode != 0:
        return None, r.stderr.decode()[-200:]
    return r.stdout, None


def thumb(png_bytes, width):
    """JPEG thumbnail at the given width; falls back to the raw PNG without PIL."""
    try:
        from PIL import Image
    except ImportError:
        return png_bytes, "image/png", None
    im = Image.open(io.BytesIO(png_bytes)).convert("RGB")
    ratio = width / im.size[0]
    im = im.resize((width, int(im.size[1] * ratio)), Image.LANCZOS)
    buf = io.BytesIO()
    im.save(buf, "JPEG", quality=72, optimize=True)
    return buf.getvalue(), "image/jpeg", im.size


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--limit", type=int, default=8)
    ap.add_argument("--out", default="~/acceptance_renders/contact_sheet.html")
    args = ap.parse_args()

    rows = psql(
        "SELECT json_build_object('created', created_at::text, 'subject', subject_key,"
        " 'body', body)::text FROM doc_notes WHERE categories ? 'acceptance-run'"
        " ORDER BY created_at DESC LIMIT %d;" % args.limit)
    runs = [json.loads(line) for line in rows.splitlines() if line.strip()]
    if not runs:
        sys.exit("no acceptance-run notes found")

    key_id, key = b2_creds()
    cards = []
    for run in runs:
        body = run["body"]
        verdict = "PASSED" if "PASSED" in body.split("\n", 1)[0] else \
                  ("FAILED" if "FAILED" in body.split("\n", 1)[0] else "?")
        skipped = re.search(r"\((\d+) skipped", body)
        # The tag is "(desktop)" on old notes, "(desktop 1366x900@1x)" with the
        # viewport field, and "(desktop 1366x900@1x, landing state)" once the
        # TL-035 (d) camera change ships — match anything in the parens, first
        # word is the profile, and the stage token decides the caption.
        uris = re.findall(r"(s3://\S+?\.png)\s*\(([^)]+)\)", body)
        imgs = []
        for uri, tag in uris:
            profile = tag.split()[0]
            stage = "landing state" if "landing" in tag else "as driven by the checks"
            data, err = fetch_s3(uri, key_id, key)
            if err:
                imgs.append("<p class='err'>%s: fetch failed — %s</p>" % (profile, err))
                continue
            width = 760 if profile == "desktop" else 380
            jpg, mime, size = thumb(data, width)
            b64 = base64.b64encode(jpg).decode()
            imgs.append(
                "<figure><figcaption>%s%s · %s</figcaption>"
                "<div class='frame'><img src='data:%s;base64,%s' alt='%s render, %s'></div>"
                "</figure>" % (profile, " %dx%d" % size if size else "", stage,
                               mime, b64, profile, stage))
        cards.append("""
<section class="card {cls}">
  <h2>{subject} — {verdict}</h2>
  <p class="meta">{created}{skip}{norender}</p>
  <div class="shots">{imgs}</div>
</section>""".format(
            cls="pass" if verdict == "PASSED" else "fail",
            subject=run["subject"], verdict=verdict, created=run["created"],
            skip=" · %s checks skipped" % skipped.group(1) if skipped else "",
            norender="" if uris else " · no render filed (pre-TL-035 run, failing "
                     "profile, or an agent that does not carry the flag — read the "
                     "run before calling it a defect)",
            imgs="".join(imgs) or "<p class='none'>no images</p>"))

    html = """<title>Acceptance renders — contact sheet</title>
<style>
:root{--bg:#f7f8fa;--card:#fff;--ink:#1c2330;--muted:#5b6575;--line:#dde2ea;
  --accent:#b45309;--pass:#1a7f37;--fail:#b42318;--frame:#eef1f5}
@media (prefers-color-scheme: dark){:root{--bg:#141821;--card:#1c222e;--ink:#e6eaf2;
  --muted:#98a2b3;--line:#2c3442;--accent:#e8a33d;--pass:#4ade80;--fail:#f87171;--frame:#10141c}}
:root[data-theme="dark"]{--bg:#141821;--card:#1c222e;--ink:#e6eaf2;--muted:#98a2b3;
  --line:#2c3442;--accent:#e8a33d;--pass:#4ade80;--fail:#f87171;--frame:#10141c}
:root[data-theme="light"]{--bg:#f7f8fa;--card:#fff;--ink:#1c2330;--muted:#5b6575;
  --line:#dde2ea;--accent:#b45309;--pass:#1a7f37;--fail:#b42318;--frame:#eef1f5}
body{font:15px/1.55 system-ui;margin:2rem auto;max-width:880px;padding:0 1rem;
  background:var(--bg);color:var(--ink)}
.banner{border:1px solid var(--line);border-left:3px solid var(--accent);
  background:var(--card);padding:.6rem 1rem;border-radius:6px}
.card{background:var(--card);border:1px solid var(--line);border-radius:8px;
  padding:1rem;margin:1.2rem 0}
.card.pass h2{color:var(--pass)}.card.fail h2{color:var(--fail)}
.meta{color:var(--muted);margin-top:-.5rem}
.shots{display:flex;gap:1rem;flex-wrap:wrap;align-items:flex-start}
.frame{max-height:520px;overflow-y:auto;border:1px solid var(--line);background:var(--frame)}
img{display:block;max-width:100%%}
figcaption{font-size:.85rem;color:var(--muted)}
.err,.none{color:var(--fail)}
.foot{color:var(--muted);font-size:.85rem}
</style>
<h1>Acceptance renders — contact sheet</h1>
<p class="banner"><strong>A render is a look, never a verdict.</strong> Images marked
<em>landing state</em> show the page as a visitor arrives; images marked <em>as driven
by the checks</em> were photographed after the tests interacted with the page (pressed
buttons, cleared panels) — on those, an odd-looking state is usually the tests' doing,
not the page's. Scroll inside each frame.</p>
%s
<p class="foot">Regenerated by <code>brochure_component_library/scripts/contact_sheet.py</code>
— weekly by cron, or on demand.</p>""" % "".join(cards)

    out = os.path.expanduser(args.out)
    os.makedirs(os.path.dirname(out), exist_ok=True)
    with open(out, "w") as f:
        f.write(html)
    print("wrote %s  (%d runs, %d KB)" % (out, len(runs), os.path.getsize(out) // 1024))


if __name__ == "__main__":
    main()
