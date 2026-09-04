#!/usr/bin/env python3
"""
audit-logo-legibility.py — can a visitor actually SEE the site's logo?

    ./scripts/audit-logo-legibility.py                 # sweep every site with a logo asset
    ./scripts/audit-logo-legibility.py --site X --site Y
    ./scripts/audit-logo-legibility.py --json out.json # machine-readable, for a filer

    ./scripts/audit-logo-legibility.py --self-test     # prove the arms fire, offline

    exit 0 = every logo measured and legible · 1 = anything else · 2 = usage failure
    A BLIND, NOT-DISPLAYED or 4.5 row always prints and never counts as a pass:
    "could not measure" must not read as "fine".

WHY THIS EXISTS — bugs_open/462. A logo can be generated, matted, guard-checked,
stored, deployed, referenced and rendered, pass every one of those gates, and be
invisible on the header it sits on. Nothing in the estate measures the one property
that matters: whether the mark contrasts with what is behind it. The render audit's
remit is text (`request_render_audit_action.go:4-5`); its findings writer files
`contrast_failure` at `css-patch-agent`, which repaints a CSS class and cannot fix a
pale PNG (`write_render_audit_findings_action.go:12-13`); images are handled
broken/not-broken only (`registry.go:1257`). This is the gap.

⚠ THIS IS THE FAST PATH, NOT THE PERMANENT HOME (462 §7a, owner ruling 2026-09-03 =
candidate 2, "report it afterwards"). It reads the header colour from the site's own
DECLARED theme token, which is a SNAPSHOT. Colour churn is documented and live on this
estate (the `generic_theme` landmine; `bugs_open/396` rewrites the theme row), so a
pass recorded here decays into a FALSE PASS — the one direction 462 is already about.
The version that stays correct measures the backdrop from the render (462 §7a option
(a), the render audit). Until that exists:
  * every row records `header_bg` AND `measured_at`, so a later reader can tell
    "passed against a palette that no longer exists" from "passed";
  * the thresholds live in ONE place (see THRESHOLDS below) so (a) can reuse them.

⚠ BINDING CONSTRAINT (462 §6): measure AFTER matting, against the HEADER, never
against the keyed generation ground. A pre-matte check sees a high-contrast
white-on-magenta image and passes it happily. This script measures served bytes, so
it is after matting by construction.

⚠ THE BACKDROP IS PER SITE AND SOMETIMES DARK. dartsonline.com's header is #111520.
A check hardcoding white would invert the verdict on every dark-header site. The
backdrop is read per site, from that site's own served CSS, and recorded in the row.

CONTROLS THIS SCRIPT RUNS ON ITSELF (a measurement that cannot come out otherwise is
not a measurement):
  - an invented-path 404 control per site — a parked domain 200s every path, so a
    200 on nonsense means the whole site's readings are meaningless -> BLIND;
  - a `</html>` control on the fetched page — a truncated or error page parses as
    "no token found" and would otherwise read as BLIND-for-the-wrong-reason;
  - the byte count of every fetched asset is asserted against Content-Length, and
    the magic bytes are read: the extension LIES (417's RUNBOOK; 12 of 12 logo
    objects sampled 2026-09-02 were JPEG under a .png key), and a truncated PNG
    still reports its header correctly to PIL — the loss only shows on pixels;
  - the header token is collected from BOTH the inline <style> blocks and the linked
    stylesheet, and the declarations must AGREE. Disagreement is the cascade
    surprise this cheap check cannot resolve -> BLIND, not a guess;
  - the token must be USED, not merely declared: a custom property that is defined
    and never applied reads identically in source (462 §1b) -> usage is required.

TWO POPULATIONS, AND ONLY ONE IS JUDGEABLE HERE. `[MEASURED 2026-09-04]` of the 34
live logo assets, 29 were fetched and **22 of those 29 have NO ALPHA CHANNEL at all**
— a background is baked into the image (SITE_DEFECT_CATEGORIES 4.5, the class
`bugs_closed/424` fixed for new generations). For those, "contrast against the header"
is the wrong question: what a visitor sees is a coloured box, and the mark inside it
reads against the BAKED background, not the header. Judging them by this rule would
flag ~20 sites whose marks are perfectly legible — verified by eye on 2026-09-04
(farmerinsurance.uk and apis.uk both read cleanly and both would have failed). So:
  * alpha-backed marks are MEASURED AND JUDGED (7 of the 29 today);
  * baked-background marks are measured against their own border colour and REPORTED
    WITHOUT A VERDICT, tagged 4.5. That is a stated blind spot, not a pass — nothing
    in the estate measures whether such a mark reads against its own baked box.

⚠ MEASURE THE URL THE PAGE REFERENCES, NOT `assets.url`. Found the hard way on
2026-09-04: `fundamentallyai.com`'s asset row holds a presigned B2 link minted
2026-08-10 with `X-Amz-Expires=604800`, so it 401s — while the served page references
`/assets/images/logo.png`, which is 200 and 157,165 bytes. Trusting the DB row
produced a BLIND row for a site whose logo is fine. The visitor loads what the HEADER
MARKUP says; the asset row is a record, and records go stale.

METHOD. Contrast is WCAG 2.x sRGB relative luminance. Each pixel with any ink in it
(alpha >= ALPHA_INK) is COMPOSITED over the header colour and the composite is
compared with the header colour — which is what a viewer's eye is actually doing.
462 §1's recorded figures used opaque-only (alpha > 200) raw pixels against white;
those numbers are reproduced in `legacy_*` fields so the two are comparable, but the
verdict is taken on the composited statistics, because a mark drawn entirely at 50%
alpha has NO opaque pixels and the legacy method cannot see it at all.
"""

import argparse, json, os, re, subprocess, sys, hashlib, random, io, time
from datetime import datetime, timezone

try:
    from PIL import Image
except ImportError:
    print("FATAL: PIL/Pillow is required (pip install pillow)", file=sys.stderr)
    sys.exit(2)

import urllib.request, urllib.error

# ─────────────────────────────────────────────────────────────────────────────
# THRESHOLDS — ONE PLACE. 462 §7a: keep the measurement and its threshold in one
# place so the render-audit version (option (a)) can reuse them rather than
# re-deriving a second, silently different rule.
# ─────────────────────────────────────────────────────────────────────────────
MIN_CONTRAST   = 3.0    # WCAG 2.x SC 1.4.11 non-text contrast floor.
ALPHA_INK      = 8      # below this a pixel contributes nothing a viewer can see.
LEGACY_OPAQUE  = 200    # 462 §1's alpha cut, kept only to reproduce its figures.
INVISIBLE_MAX  = 1.5    # contrast below this = indistinguishable from the backdrop.

# CALIBRATED 2026-09-04 against the whole live population (34 logo assets, 7 of them
# alpha-backed and therefore judgeable here). TWO ARMS, because each catches a real
# artefact the other misses — neither is a rounded-off guess:
#
#   arm A  max < MIN_CONTRAST — NO pixel anywhere in the mark clears the floor.
#          462 §1 calls this "the load-bearing row". Fires on the PRE-regeneration
#          websitepromotion mark (max 2.55:1) and on mortgagecalculator.co.uk
#          (max 2.39:1, measured 2026-09-04).
#   arm B  legible_frac < LEGIBLE_INK_MIN_FRAC — too little of the mark reads, even
#          though some sliver of it does. Fires on the POST-regeneration
#          websitepromotion mark, which arm A CANNOT see: it reaches 20.75:1 on a
#          magenta despill fringe while 86% of it is white on a white header
#          (462 §6). A max-only rule passes the motivating case.
#
# LEGIBLE_INK_MIN_FRAC sits in a wide empty gap in the measured distribution —
# 0.0%, 6.7% | (nothing) | 26.4%, 29.3%, 50.0%, 75.5%, 88.1%. Worst passing artefact
# is gamedesign.uk at 26.4% (1.8x above); worst failing is websitepromotion at 6.7%
# (2.2x below). ⚠ n=7. This is calibrated, not proven: re-derive it when the alpha
# population is bigger, and NEVER move it to make one artefact pass.
LEGIBLE_INK_MIN_FRAC = 0.15

# ─────────────────────────────────────────────────────────────────────────────
# DB — TWO PATHS, chosen by environment, because this runs in two places.
#
#   from a session on this machine  -> kubectl exec into postgres-clients-0
#   from INSIDE the cluster (the CronJob) -> psql straight at PG_CLIENTS_HOST
#
# The in-cluster path is NOT a convenience. `ai-persona-app` has no pods/exec
# RBAC in this namespace — the constraint single-owner-carriers-check and
# bugs-open-staleness-sweep both hit — so `kubectl exec` cannot work there at
# all, and a scheduled run would fail on a permission error that reads like a
# broken query. PG_CLIENTS_HOST being set is the fleet's own switch for this
# (stated in component-render-check's cronjob).
#
# ⚠ BOTH paths must carry -tAF| . Every parser below splits on "|"; a default
# psql returns ALIGNED, boxed output, which does not error — it parses as
# nonsense, i.e. it would fail in the one direction this check must never fail.
# The flags live in one list so the two paths cannot drift apart.
PSQL_FLAGS = ["-tAF|", "-v", "ON_ERROR_STOP=1"]


def _psql_argv():
    """(argv-without-the-SQL, in_cluster)."""
    host = os.environ.get("PG_CLIENTS_HOST")
    if host:
        return (["psql", "-h", host, "-p", "5432", "-U", "clients_user",
                 "-d", "clients_db"] + PSQL_FLAGS, True)
    return (["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
             "--", "psql", "-U", "clients_user", "-d", "clients_db"] + PSQL_FLAGS, False)


def _psql_env(in_cluster):
    if not in_cluster:
        return None
    env = dict(os.environ)
    pw = os.environ.get("CLIENTS_DB_PASSWORD")
    if not pw:
        # Refuse rather than fall through to the kubectl path: in-cluster that
        # path has no RBAC, so the fallback would turn a missing secret into a
        # confusing permission error instead of this sentence.
        print("FATAL: PG_CLIENTS_HOST is set but CLIENTS_DB_PASSWORD is not.",
              file=sys.stderr)
        sys.exit(2)
    env["PGPASSWORD"] = pw
    return env


NAMED = {"white": (255, 255, 255), "black": (0, 0, 0), "transparent": None}


def q(sql):
    argv, in_cluster = _psql_argv()
    r = subprocess.run(argv + ["-c", sql], capture_output=True, text=True,
                       stdin=subprocess.DEVNULL, env=_psql_env(in_cluster))
    if r.returncode != 0:
        print("FATAL: psql failed: " + r.stderr.strip()[:400], file=sys.stderr)
        sys.exit(2)
    return [ln for ln in r.stdout.splitlines() if ln.strip()]


DOC_NOTE_TAG = "logolegbody"


def write_doc_note(body):
    """ONE doc_notes row per run — on a clean result too.

    The reason is the fleet's, and it is the whole point of a scheduled check: a
    MISSING row must mean "the job did not run", never "the job ran and found
    nothing". A check that only speaks when it finds something is
    indistinguishable from a check that has silently stopped — which is 462 one
    level up.
    """
    if DOC_NOTE_TAG in body:
        # Dollar-quoting is what keeps arbitrary report text out of the SQL
        # grammar; a body containing the tag would end the literal early.
        print("FATAL: report body contains the dollar-quote tag; refusing to write.",
              file=sys.stderr)
        sys.exit(2)
    sql = ("INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
           "VALUES ('pipeline', 'logo-legibility', "
           "${t}${b}${t}$, '[\"logo-legibility\"]'::jsonb, 'logo-legibility-check');"
           ).replace("{t}", DOC_NOTE_TAG).replace("{b}", body)
    argv, in_cluster = _psql_argv()
    r = subprocess.run(argv + ["-c", sql], capture_output=True, text=True,
                       stdin=subprocess.DEVNULL, env=_psql_env(in_cluster))
    if r.returncode != 0:
        print("FATAL: doc_notes write failed: " + r.stderr.strip()[:400], file=sys.stderr)
        sys.exit(2)
    print("doc_notes row written (subject_type='pipeline', subject_key='logo-legibility')")


# ── colour ──────────────────────────────────────────────────────────────────
def parse_colour(s):
    """A CSS colour -> (r,g,b), or None if this check cannot resolve it (gradient,
    url(), currentColor, an unknown name). None means BLIND, never a guess."""
    s = s.strip().rstrip(";").strip()
    if not s:
        return None
    m = re.fullmatch(r"#([0-9a-fA-F]{3,8})", s)
    if m:
        h = m.group(1)
        if len(h) in (3, 4):
            h = "".join(c * 2 for c in h[:3])
        if len(h) >= 6:
            return tuple(int(h[i:i + 2], 16) for i in (0, 2, 4))
        return None
    m = re.fullmatch(r"rgba?\(([^)]*)\)", s, re.I)
    if m:
        parts = [p.strip() for p in re.split(r"[,\s/]+", m.group(1)) if p.strip()]
        if len(parts) >= 3:
            try:
                out = []
                for p in parts[:3]:
                    out.append(round(float(p[:-1]) * 255 / 100) if p.endswith("%") else int(float(p)))
                return tuple(max(0, min(255, v)) for v in out)
            except ValueError:
                return None
    return NAMED.get(s.lower())


def lum(rgb):
    def f(c):
        c = c / 255.0
        return c / 12.92 if c <= 0.03928 else (((c + 0.055) / 1.055) ** 2.4)
    r, g, b = rgb
    return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b)


def contrast(a, b):
    la, lb = lum(a) + 0.05, lum(b) + 0.05
    return (la / lb) if la >= lb else (lb / la)


# ── fetch ───────────────────────────────────────────────────────────────────
def fetch(url, timeout=30, tries=3):
    """-> (status, body_bytes, note). note is non-empty when the fetch is suspect.

    ⚠ RETRIES ARE LOAD-BEARING. Measured 2026-09-04: two hosts that had answered 200
    minutes earlier returned a 404 and a connection error inside one sweep — a burst
    of three requests per site is enough to trip a CDN. Without a retry those land as
    BLIND rows, which a reader then has to tell apart from a real one."""
    last = (0, b"", "")
    for attempt in range(tries):
        st, body, note = _fetch_once(url, timeout)
        if st == 200:
            return st, body, note
        last = (st, body, note)
        if attempt < tries - 1:
            time.sleep(1.5 * (attempt + 1))
    return last


def _fetch_once(url, timeout):
    req = urllib.request.Request(url, headers={"User-Agent": "agentchassis-logo-legibility/1"})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            body = r.read()
            declared = r.headers.get("Content-Length")
            note = ""
            # ⚠ assert the byte count: a truncated image still reports its header
            # correctly to PIL and only fails on pixel access (417 RUNBOOK).
            if declared is not None and declared.isdigit() and int(declared) != len(body):
                note = "TRUNCATED: Content-Length %s, got %d" % (declared, len(body))
            return r.status, body, note
    except urllib.error.HTTPError as e:
        return e.code, b"", ""
    except Exception as e:
        return 0, b"", "fetch error: %s" % (type(e).__name__,)


MAGIC = [(b"\x89PNG\r\n\x1a\n", "PNG"), (b"\xff\xd8\xff", "JPEG"),
         (b"GIF8", "GIF"), (b"RIFF", "WEBP?"), (b"<svg", "SVG"), (b"<?xml", "SVG?")]


def magic_of(b):
    for sig, name in MAGIC:
        if b.startswith(sig):
            return name
    return "UNKNOWN(%s)" % b[:4].hex()


# ── the header backdrop, from the site's own served CSS ─────────────────────
DECL_RE  = re.compile(r"--color-header-bg\s*:\s*([^;}]+)[;}]")
USE_RE   = re.compile(r"background(?:-color)?\s*:\s*([^;}]*var\(\s*--color-header-bg[^;}]*)[;}]", re.I)
FALLBK   = re.compile(r"var\(\s*--color-header-bg\s*,\s*([^)]+)\)")
VAR_RE   = re.compile(r"var\(\s*(--[a-zA-Z0-9_-]+)\s*\)")


def decl_values(css, prop):
    rx = re.compile(re.escape(prop) + r"\s*:\s*([^;}]+)[;}]")
    return [m.group(1).strip() for m in rx.finditer(css)]


def resolve_header_bg(page_html, extra_css):
    """-> (rgb, provenance_string, blind_reason). Collects from BOTH the inline
    <style> blocks and the linked stylesheet and requires them to AGREE; a
    disagreement is the cascade surprise this cheap check cannot resolve."""
    corpus = page_html + "\n" + extra_css

    uses = USE_RE.findall(corpus)
    raw = [v.strip() for v in DECL_RE.findall(corpus)]
    if not uses:
        if not raw:
            return None, "", ("no --color-header-bg anywhere in the served CSS — this site is not "
                              "on the token-based theme, so its header colour cannot be read "
                              "without a render (462 §7a option (a))")
        return None, "", "--color-header-bg is declared but never applied to a background (462 §1b)"

    distinct = sorted({" ".join(v.split()) for v in raw})

    if len(distinct) > 1:
        return None, "", "--color-header-bg declared with %d different values %s — the cascade decides, and only a render can (462 §7a option (a))" % (len(distinct), distinct)

    if distinct:
        rgb = parse_colour(distinct[0])
        if rgb is None:
            return None, "", "--color-header-bg is %r, which this check cannot resolve to a flat colour" % distinct[0][:60]
        return rgb, "--color-header-bg: %s" % distinct[0], ""

    # Declared nowhere: the usages may still carry a fallback, e.g.
    # var(--color-header-bg, var(--color-surface)).
    for u in uses:
        fb = FALLBK.search(u)
        if not fb:
            continue
        val = fb.group(1).strip()
        v2 = VAR_RE.fullmatch(val)
        if v2:
            inner = sorted({" ".join(x.split()) for x in decl_values(corpus, v2.group(1))})
            if len(inner) == 1:
                rgb = parse_colour(inner[0])
                if rgb:
                    return rgb, "fallback %s: %s" % (v2.group(1), inner[0]), ""
            continue
        rgb = parse_colour(val)
        if rgb:
            return rgb, "fallback literal: %s" % val, ""
    return None, "", "--color-header-bg is used but never declared, and no fallback resolves"


# ── which image the visitor actually loads ─────────────────────────────────
IMG_RE = re.compile(r"<img\b[^>]*>", re.I)
SRC_RE = re.compile(r'src=["\']([^"\']+)["\']', re.I)


def logo_src_from_page(page):
    """The <img> the site's own header markup points at, which is what a visitor
    loads. -> (src, provenance) or (None, why-not).

    ⚠ THE FIRST <header> IN THE DOCUMENT IS OFTEN NOT THE SITE HEADER. Measured
    2026-09-04: cookly.uk, webdesign.co.uk and ai-agent-orchestration.com all open
    with a CONTENT header (`<header class="info-card-grid__header">`) long before the
    site chrome. Taking the first one found no logo, fell back to `assets.url`, and
    produced a confident verdict about an image those pages never load."""
    blocks = re.findall(r"<header\b[^>]*>.{0,6000}?</header>", page, re.I | re.S)
    site = [b for b in blocks if re.search(r'<header\b[^>]*class=["\'][^"\']*(site-header|-header\b)', b, re.I)]
    ordered = [b for b in site if "logo" in b.lower()] + \
              [b for b in blocks if "logo" in b.lower()] + site + blocks
    seen, scan = set(), []
    for b in ordered:
        if id(b) not in seen:
            seen.add(id(b)); scan.append(b)
    for b in scan:
        for tag in IMG_RE.findall(b):
            if "logo" in tag.lower():
                m = SRC_RE.search(tag)
                if m:
                    return m.group(1), "logo <img> in a <header> block"
    for b in scan:
        for tag in IMG_RE.findall(b):
            m = SRC_RE.search(tag)
            if m:
                return m.group(1), "first <img> in a <header> block (no 'logo' in the tag)"
    for tag in IMG_RE.findall(page):
        if "logo" in tag.lower():
            m = SRC_RE.search(tag)
            if m:
                return m.group(1), "'logo' <img> outside any <header>"
    why = "the served header has no logo image"
    if re.search(r'class=["\'][^"\']*logo-text', page, re.I):
        why += " — it renders class=\"logo-text\" instead (417 RUNBOOK: 'a site has a logo asset but the header still shows TEXT')"
    return None, why


# ── the measurement ─────────────────────────────────────────────────────────
def border_modal(im):
    """The modal colour of the image's 1px border ring — for a baked-background mark
    this is its own backdrop. Same idea as the keyed-ground statistic the generator
    already computes on the border (`keyground.go`)."""
    w, h = im.size
    ring = {}
    for x in range(w):
        for y in (0, h - 1):
            ring[im.getpixel((x, y))[:3]] = ring.get(im.getpixel((x, y))[:3], 0) + 1
    for y in range(h):
        for x in (0, w - 1):
            ring[im.getpixel((x, y))[:3]] = ring.get(im.getpixel((x, y))[:3], 0) + 1
    return max(ring.items(), key=lambda kv: kv[1])[0] if ring else (255, 255, 255)


def measure(img_bytes, bg):
    """Composite every inked pixel over `bg` and compare with `bg`. Returns a dict
    of statistics, or {'error': ...}."""
    try:
        im = Image.open(io.BytesIO(img_bytes))
        im.load()                       # force pixel access: a truncated file dies HERE
        has_alpha = im.mode in ("RGBA", "LA", "PA") or "transparency" in im.info
        im = im.convert("RGBA")
    except Exception as e:
        return {"error": "PIL: %s: %s" % (type(e).__name__, str(e)[:120])}

    cols = im.getcolors(maxcolors=1 << 24)
    if cols is None:
        cols = [(1, p) for p in im.getdata()]

    br, bgc, bb = bg
    ink = []            # (contrast, count) composited
    legacy = []         # (contrast, count) 462 §1 method: opaque, raw, vs white
    n_ink = n_total = 0
    n_invisible = n_legible = 0
    for count, px in cols:
        r, g, b, a = px
        n_total += count
        if a < ALPHA_INK:
            continue
        n_ink += count
        f = a / 255.0
        comp = (round(r * f + br * (1 - f)), round(g * f + bgc * (1 - f)), round(b * f + bb * (1 - f)))
        c = contrast(comp, bg)
        ink.append((c, count))
        if c < INVISIBLE_MAX:
            n_invisible += count
        if c >= MIN_CONTRAST:
            n_legible += count
        if a > LEGACY_OPAQUE:
            legacy.append((contrast((r, g, b), (255, 255, 255)), count))

    if n_ink == 0:
        return {"error": "no inked pixels at all (alpha < %d everywhere) — the asset is blank" % ALPHA_INK,
                "n_total": n_total, "has_alpha": has_alpha, "size": list(im.size)}

    def pct(pairs, p):
        pairs = sorted(pairs)
        tot = sum(c for _, c in pairs)
        want, seen = tot * p, 0
        for v, c in pairs:
            seen += c
            if seen >= want:
                return v
        return pairs[-1][0]

    out = {
        "size": list(im.size), "has_alpha": has_alpha,
        "n_total": n_total, "n_ink": n_ink,
        "min": round(min(v for v, _ in ink), 2),
        "median": round(pct(ink, 0.5), 2),
        "p95": round(pct(ink, 0.95), 2),
        "max": round(max(v for v, _ in ink), 2),
        "legible_px": n_legible,
        "legible_frac": round(n_legible / n_ink, 4),
        "invisible_frac": round(n_invisible / n_ink, 4),
        "ink_frac_of_frame": round(n_ink / n_total, 4),
    }
    if not has_alpha:
        # A baked-background mark: the header is NOT its backdrop. Measure it against
        # its own border colour instead and report WITHOUT a verdict (see the
        # docstring's "TWO POPULATIONS").
        bb = border_modal(im)
        inner = []
        for count, px in cols:
            c = contrast(px[:3], bb)
            inner.append((c, count))
        out["baked_bg"] = "#%02x%02x%02x" % bb
        out["baked_median"] = round(pct(inner, 0.5), 2)
        out["baked_max"] = round(max(v for v, _ in inner), 2)
        out["baked_legible_frac"] = round(
            sum(c for v, c in inner if v >= MIN_CONTRAST) / max(1, sum(c for _, c in inner)), 4)
    if legacy:
        out["legacy_opaque_median_vs_white"] = round(pct(legacy, 0.5), 2)
        out["legacy_opaque_max_vs_white"] = round(max(v for v, _ in legacy), 2)
        out["legacy_n_opaque"] = sum(c for _, c in legacy)
    else:
        out["legacy_n_opaque"] = 0
    return out


def verdict(m):
    """-> (ok, reason, arm). Only ever called for an alpha-backed mark: see
    `judgeable()`. `arm` names which rule fired, so a finding says why."""
    if m["max"] < MIN_CONTRAST:
        return False, ("no pixel anywhere in the mark reaches %.1f:1 against its header "
                       "(max %.2f:1, median %.2f:1)" % (MIN_CONTRAST, m["max"], m["median"])), "A"
    if m["legible_frac"] < LEGIBLE_INK_MIN_FRAC:
        return False, ("only %.1f%% of the mark's ink clears %.1f:1 against its header "
                       "(floor %.0f%%); median %.2f:1, %.1f%% of it is indistinguishable "
                       "from the backdrop" % (m["legible_frac"] * 100, MIN_CONTRAST,
                                              LEGIBLE_INK_MIN_FRAC * 100, m["median"],
                                              m["invisible_frac"] * 100)), "B"
    return True, "%.1f%% of ink clears %.1f:1 (median %.2f:1, max %.2f:1)" % (
        m["legible_frac"] * 100, MIN_CONTRAST, m["median"], m["max"]), ""



# ─────────────────────────────────────────────────────────────────────────────
def self_test():
    """Prove the arms fire, offline. Two of the six cases are the REAL artefacts
    462 is about — the lane preserved websitepromotion's logo either side of its
    2026-09-03 regeneration, and they are the only copies that exist (462 §6: a
    regeneration UPSERTs the row and mints a fresh key, so there is no rollback).
    Cases 5 and 6 are the same white mark against a dark and a white header: they
    are here to prove the backdrop operand is LOAD-BEARING rather than decorative —
    a check that ignored it would give both the same verdict."""
    D = "docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy"
    W, K = (255, 255, 255), (17, 21, 32)     # white header, dartsonline's #111520

    def synth(mark_rgb, alpha=255, frac=0.25):
        im = Image.new("RGBA", (100, 100), (0, 0, 0, 0))
        for y in range(int(100 * frac)):
            for x in range(100):
                im.putpixel((x, y), mark_rgb + (alpha,))
        b = io.BytesIO(); im.save(b, "PNG"); return b.getvalue()

    cases = [
        ("preserved pre-regeneration websitepromotion  (462 §1: max 2.55:1)",
         D + "/PRESERVED_websitepromotion_logo_pre_regeneration_2026-09-03.png", W, False),
        ("preserved post-regeneration websitepromotion (462 §6: 85% white on white)",
         D + "/PRESERVED_websitepromotion_logo_post_regeneration_2026-09-03.png", W, False),
        ("synthetic black mark on a white header", synth((0, 0, 0)), W, True),
        ("synthetic near-white mark on a white header", synth((250, 250, 250)), W, False),
        ("synthetic WHITE mark on a DARK header", synth((255, 255, 255)), K, True),
        ("synthetic WHITE mark on a WHITE header  (same bytes as above)", synth((255, 255, 255)), W, False),
    ]
    ok_all, ran = True, 0
    for name, src, bg, want_ok in cases:
        if isinstance(src, str):
            try:
                data = open(src, "rb").read()
            except OSError as e:
                print("  SKIP  %-62s (%s)" % (name, e.__class__.__name__))
                ok_all = False
                continue
        else:
            data = src
        m = measure(data, bg)
        if "error" in m:
            print("  FAIL  %-62s measure error: %s" % (name, m["error"])); ok_all = False; continue
        got_ok, reason, arm = verdict(m)
        ran += 1
        mark = "PASS" if got_ok == want_ok else "FAIL"
        if got_ok != want_ok:
            ok_all = False
        print("  %s  %-62s want %-8s got %-8s %s" % (
            mark, name, "legible" if want_ok else "FINDING",
            "legible" if got_ok else "FINDING(%s)" % arm, reason[:70]))
    print("\nself-test: %d case(s) run, %s" % (ran, "all as expected" if ok_all else "MISMATCH — the check is wrong"))
    if ran < len(cases):
        print("⚠ a SKIPPED case is not a pass: the preserved artefacts are the only real "
              "known-bad inputs this check has, and without them the arms are proven only "
              "against synthetics.")
    return 0 if ok_all else 1


def render_report(results, measured_at, findings, blind, undisplayed, unjudged,
                  clean, partial):
    """The doc_notes body. Written to be read by someone who was not here.

    Every counted bucket appears even when it is zero, because the shape of this
    check's answer is the finding (462 §8a): the headline is not how many logos
    fail, it is how few can be JUDGED. A report that printed only failures would
    hide that 22 of 34 marks are unjudgeable, which is the thing a reader most
    needs to know before quoting a clean run as "the estate's logos are fine".
    """
    L = ["LOGO LEGIBILITY SWEEP (bugs_open/462) — measured %s" % measured_at, ""]
    if partial:
        L += ["⚠ PARTIAL RUN — limited to named sites with --site. Not a fleet census.", ""]
    L += ["logo assets examined:      %d" % len(results),
          "  FINDING (below floor):   %d" % findings,
          "  measured legible:        %d" % clean,
          "  not judged (baked bg):   %d   <- SITE_DEFECT_CATEGORIES 4.5, not a pass" % unjudged,
          "  not displayed at all:    %d   <- asset exists, served header never loads it" % undisplayed,
          "  BLIND (unmeasurable):    %d   <- NOT a pass; nobody has measured these" % blind,
          "",
          "thresholds: floor %.1f:1 (WCAG 2.x non-text), arm B needs >= %.0f%% of ink over it"
          % (MIN_CONTRAST, LEGIBLE_INK_MIN_FRAC * 100),
          ""]
    fs = [r for r in results if r.get("ok") is False]
    if fs:
        L.append("FINDINGS")
        for r in fs:
            L.append("  %s  [arm %s]" % (r["domain"], r.get("arm")))
            L.append("    %s" % r.get("reason"))
            L.append("    header %s [%s] · %s %dB md5 %s · %s"
                     % (r.get("header_bg"), r.get("header_bg_provenance"), r.get("magic"),
                        r.get("asset_bytes") or 0, r.get("asset_md5"), r.get("fetched")))
        L.append("")
    else:
        L += ["No logo is below the floor in this run.", ""]
    bl = [r for r in results if r.get("blind")]
    if bl:
        L.append("BLIND — measure these by hand; a blind row outlives its blindness")
        for r in bl:
            L.append("  %s: %s" % (r["domain"], r["blind"]))
        L.append("")
    L += ["⚠ Each row was measured against the header colour DECLARED at the timestamp above.",
          "  That token is a SNAPSHOT. Theme rows get rewritten (bugs_open/396), so a pass here",
          "  decays into a false pass — the one direction this bug is already about. The version",
          "  that stays correct reads the backdrop from the render (462 §7a option (a)); unbuilt.",
          "",
          "⚠ This check REPORTS. Nothing files, by design as of 2026-09-04 — see 462 §9 for why",
          "  every automatic destination is destructive, dead or manual, and §9e for the fork."]
    return "\n".join(L)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--site", action="append", default=[], help="limit to these domains (repeatable)")
    ap.add_argument("--json", help="write the full result set here")
    ap.add_argument("--quiet", action="store_true", help="print findings and BLIND rows only")
    ap.add_argument("--report", action="store_true",
                    help="write ONE doc_notes row for this run — on a clean result too. "
                         "For the scheduled job: a MISSING row must mean the job did not run, "
                         "never that it ran and found nothing.")
    ap.add_argument("--self-test", action="store_true",
                    help="prove the arms fire against the preserved real artefacts + synthetics, offline")
    args = ap.parse_args()

    if args.self_test:
        return self_test()

    rows = q("""SELECT s.domain,
                       coalesce(nullif(s.publish_project,''), s.domain),
                       a.url,
                       to_char(a.updated_at,'YYYY-MM-DD')
                FROM assets a JOIN sites s ON s.id = a.site_id
                WHERE a.status='active' AND a.purpose='logo'
                ORDER BY s.domain;""")
    pop = [r.split("|") for r in rows]
    if args.site:
        want = set(args.site)
        pop = [p for p in pop if p[0] in want]
        missing = want - {p[0] for p in pop}
        for m in sorted(missing):
            print("  ⚠ %s has no active logo asset — nothing to measure" % m)
    if not pop:
        print("FATAL: no logo assets in the population", file=sys.stderr)
        return 2

    measured_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    print("logo legibility sweep — %s — %d logo asset(s), floor %.1f:1 (WCAG 2.x non-text)\n"
          % (measured_at, len(pop), MIN_CONTRAST))

    results, findings, blind, unjudged, undisplayed = [], 0, 0, 0, 0
    for domain, host, url, asset_updated in pop:
        rec = {"domain": domain, "host": host, "asset_url": url,
               "asset_updated": asset_updated, "measured_at": measured_at}
        base = "https://%s" % host

        # control first: a parked domain 200s every path, which would make every
        # subsequent reading meaningless.
        ctl, _, _ = fetch("%s/zzz-not-real-%d.html" % (base, random.randint(10**6, 10**7)), timeout=20)
        rec["control_404"] = ctl
        if ctl == 200:
            rec["blind"] = "invented-path control returned 200 — this host answers everything"
        else:
            code, html, note = fetch("%s/index.html?cb=%d" % (base, random.randint(1, 10**6)))
            rec["index_status"], rec["index_bytes"] = code, len(html)
            page = html.decode("utf-8", "replace")
            if code != 200 or len(html) == 0:
                rec["blind"] = "index.html returned %s (%d bytes)%s" % (
                    code, len(html), (" — " + note) if note else "")
            elif "</html>" not in page.lower():
                rec["blind"] = "index.html has no </html> — truncated or an error page"
            else:
                css = ""
                for href in re.findall(r'<link[^>]+rel=["\']?stylesheet["\']?[^>]*href=["\']([^"\']+)["\']', page, re.I) \
                          + re.findall(r'<link[^>]+href=["\']([^"\']+)["\'][^>]*rel=["\']?stylesheet', page, re.I):
                    if href.startswith("http"):
                        cu = href
                    elif href.startswith("/"):
                        cu = base + href
                    else:
                        cu = base + "/" + href
                    sc, sb, _ = fetch(cu + ("&" if "?" in cu else "?") + "cb=%d" % random.randint(1, 10**6))
                    if sc == 200:
                        css += "\n" + sb.decode("utf-8", "replace")
                rec["css_bytes"] = len(css)
                bg, prov, why = resolve_header_bg(page, css)
                if bg is None:
                    rec["blind"] = why
                else:
                    rec["header_bg"] = "#%02x%02x%02x" % bg
                    rec["header_bg_provenance"] = prov

                    # ⚠ the URL the PAGE references, not assets.url (see docstring)
                    src, where = logo_src_from_page(page)
                    rec["asset_url_db"] = url
                    if src:
                        rec["src_from"] = where
                        if src != url:
                            rec["src_differs_from_db"] = True
                    else:
                        rec["not_displayed"] = where
                        src = None
                    if src is None:
                        pass                    # nothing a visitor loads: never measured
                    else:
                        iu = src if src.startswith("http") else base + (src if src.startswith("/") else "/" + src)
                        rec["fetched"] = iu.split("?")[0]
                        if rec["fetched"].lower().endswith(".svg"):
                            rec["blind"] = "the header logo is an SVG (%s) — this check measures raster pixels" % rec["fetched"]
                        else:
                            ic, ib, inote = fetch(iu + ("&" if "?" in iu else "?") + "cb=%d" % random.randint(1, 10**6))
                            rec["asset_status"], rec["asset_bytes"] = ic, len(ib)
                            if ic != 200 or not ib:
                                rec["blind"] = "logo fetch returned %s (%d bytes)%s at %s" % (
                                    ic, len(ib), (" — " + inote) if inote else "", rec["fetched"])
                            elif inote:
                                rec["blind"] = inote
                            else:
                                rec["asset_md5"] = hashlib.md5(ib).hexdigest()[:12]
                                rec["magic"] = magic_of(ib)
                                m = measure(ib, bg)
                                if "error" in m:
                                    rec["blind"] = m["error"]
                                    rec.update({k: v for k, v in m.items() if k != "error"})
                                else:
                                    rec.update(m)
                                    if m["has_alpha"]:
                                        ok, reason, arm = verdict(m)
                                        rec["ok"], rec["reason"], rec["arm"] = ok, reason, arm
                                    else:
                                        # baked background: 4.5's class, not this
                                        # check's verdict. Reported, never passed.
                                        rec["ok"] = None
                                        rec["reason"] = ("baked background (no alpha) — not judged here: "
                                                         "SITE_DEFECT_CATEGORIES 4.5. Against its own box "
                                                         "%s the mark is max %.2f:1, %.1f%% of ink over %.1f:1"
                                                         % (m["baked_bg"], m["baked_max"],
                                                            m["baked_legible_frac"] * 100, MIN_CONTRAST))

        results.append(rec)
        if rec.get("not_displayed"):
            undisplayed += 1
            print("  ⚠ NOT DISPLAYED  %-28s %s" % (domain, rec["not_displayed"]))
            print("      the site holds an active logo asset (%s) that the served page never loads, "
                  "so its legibility is moot until the header is rebuilt" % rec["asset_url_db"])
        elif rec.get("blind"):
            blind += 1
            print("  ⚠ BLIND  %-34s %s" % (domain, rec["blind"]))
        elif rec.get("ok") is False:
            findings += 1
            print("  ⚠ FINDING (arm %s)  %s" % (rec["arm"], domain))
            print("      %s" % rec["reason"])
            print("      header %s [%s] · %s %s %dB md5 %s · %s · measured %s"
                  % (rec["header_bg"], rec["header_bg_provenance"], rec["magic"],
                     "x".join(map(str, rec["size"])), rec["asset_bytes"], rec["asset_md5"],
                     rec["fetched"], measured_at))
        elif rec.get("ok") is None:
            unjudged += 1
            if not args.quiet:
                print("  4.5  %-38s header %-8s baked box %-8s mark-vs-box max %7.2f  ink over floor %5.1f%%  [not judged]"
                      % (domain, rec["header_bg"], rec["baked_bg"],
                         rec["baked_max"], rec["baked_legible_frac"] * 100))
        elif not args.quiet:
            print("  ok   %-38s header %-8s ink %7d  min %6.2f  med %6.2f  max %7.2f  legible %5.1f%% (%d px)  invisible %5.1f%%"
                  % (domain, rec["header_bg"], rec["n_ink"], rec["min"], rec["median"], rec["max"],
                     rec["legible_frac"] * 100, rec["legible_px"], rec["invisible_frac"] * 100))

    clean = len(results) - findings - blind - unjudged - undisplayed
    print("\n%d logo asset(s): %d FINDING(s), %d blind, %d not displayed at all, %d not "
          "judged here (baked background, 4.5), %d measured legible"
          % (len(results), findings, blind, undisplayed, unjudged, clean))
    if blind:
        print("⚠ a BLIND row is NOT a pass — it is a logo nobody has measured.")
    if unjudged:
        print("⚠ a 4.5 row is NOT a pass either — nothing in the estate measures whether a "
              "baked-background mark reads against its own box. Stated blind spot.")
    print("⚠ every row was measured against the header colour DECLARED at %s. That token is a "
          "snapshot; re-run after any theme change (462 §7a)." % measured_at)

    if args.report:
        write_doc_note(render_report(results, measured_at, findings, blind,
                                     undisplayed, unjudged, clean, bool(args.site)))

    if args.json:
        with open(args.json, "w") as fh:
            json.dump({"measured_at": measured_at,
                       "thresholds": {"min_contrast": MIN_CONTRAST,
                                      "legible_ink_min_frac": LEGIBLE_INK_MIN_FRAC,
                                      "alpha_ink": ALPHA_INK, "invisible_max": INVISIBLE_MAX},
                       "results": results}, fh, indent=2)
        print("wrote %s" % args.json)

    # EXIT CODE. Two different questions, and conflating them makes a check
    # nobody reads.
    #
    #   hand-run  -> "is the estate clean?"  1 if anything was not measured-and-
    #                legible. A BLIND row is not a pass.
    #   --report  -> "did the RUN work?"  The findings are the doc_notes row's
    #                job, not the exit code's.
    #
    # MEASURED 2026-09-04, first in-cluster run: without this split the job exits
    # 1 on the two standing findings — one of which the owner has RULED PERMANENT
    # (462 §7, websitepromotion stays) — so the CronJob is red for ever by design,
    # retries the whole 5-minute sweep on backoffLimit, and writes a SECOND
    # doc_notes row for one run, breaking the one-row-per-run rule this flag
    # exists to keep. "A permanently-red job is a job everybody learns to ignore"
    # is written into component-render-check's own header; this is that mistake.
    #
    # What DOES still fail a scheduled run is the sweep being broken rather than
    # the estate being imperfect: nothing measured at all, or every row blind
    # (egress gone, theme tokens unreadable). An operational failure — no
    # password, psql refusing — has already exited 2 long before here.
    if args.report:
        if not results:
            print("FAIL: no logo assets examined — the sweep measured nothing.",
                  file=sys.stderr)
            return 1
        if blind == len(results):
            print("FAIL: every row is BLIND — the sweep is broken, not the estate.",
                  file=sys.stderr)
            return 1
        return 0
    return 1 if (findings or blind or undisplayed) else 0


if __name__ == "__main__":
    sys.exit(main())
