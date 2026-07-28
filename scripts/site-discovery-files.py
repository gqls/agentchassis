#!/usr/bin/env python3
"""
site-discovery-files.py — generate robots.txt, sitemap.xml and llms.txt for ANY site.

WHY THIS EXISTS
    On 2026-07-28 relojistas.com was found to be publishing curated content daily while
    being very nearly invisible to search: real crawlers spent 78% of their budget on a
    dead forum's URLs, the site had no sitemap, and Cloudflare's managed robots.txt was
    turning ClaudeBot away at the door. None of that is specific to relojistas — every
    site in the estate has the same three files missing or wrong.

    Full evidence: docs024_key_docs_latest/traffic_probe/EVIDENCE_2026-07-28_crawl_budget*

THE THREE RULES THIS TOOL ENFORCES, because each was learned the expensive way:

  1. PROBE BEFORE LISTING. A sitemap advertising a 404 is worse than no sitemap. Every
     URL is fetched and only 200s are emitted. Use --no-probe at your peril.

     BUT THE PROBE IS POINT-IN-TIME, and that cuts both ways. On 2026-07-28 it dropped
     oufe.com/cases/thames-water.html as 404 — correctly, at that moment: the page was
     deployed 1.5 hours later. Run this when the target site is NOT mid-build, and treat
     a dropped URL as "not fetchable right now", never as "broken".

  2. llms.txt IS BUILT *FROM* THE PAGES, NOT WRITTEN *ABOUT* THEM. Each entry is the
     page's own <h1> and its own first sentence. Nothing here invents a description of
     a site — that is how unsupported claims get published.

  3. CHECK WHO IS ACTUALLY SERVING robots.txt. Cloudflare's managed file is PREPENDED to
     the origin's, not replaced by it — so shipping your own changes nothing until the
     dashboard setting is off. This tool detects the merge and says so.

USAGE
    ./scripts/site-discovery-files.py <domain> [--out DIR] [--write] [--no-probe]

    Default is a DRY RUN printing a summary and what would change. --write emits files
    into --out (default: the site's repo checkout if one is found, else ./out/<domain>).
    Committing and pushing is left to you, deliberately — this tool does not deploy.
"""
import argparse, html, json, os, re, subprocess, sys, urllib.parse
from datetime import datetime, timezone

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A", "-F", "\x1f", "-c"]
UA = "Mozilla/5.0 (compatible; site-discovery-files/1.0; +internal tooling)"


def q(sql, width=None):
    """Rows as lists. psql emits NOTHING for trailing empty columns, so a row of
    (uuid, '', '') comes back as one field — pad to `width` rather than unpack blind.
    Cost of learning this: two tracebacks on the second site it was ever run against."""
    r = subprocess.run(PSQL + [sql], capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed: {r.stderr.strip()[:400]}")
    rows = [ln.split("\x1f") for ln in r.stdout.strip().splitlines() if ln.strip()]
    if width:
        rows = [row + [""] * (width - len(row)) for row in rows]
    return rows


def fetch(url, timeout=20):
    """curl, not urllib — Cloudflare rejects urllib's default UA and you get a silent
    empty result that looks like 'the site has no content'. Learned 2026-07-28."""
    r = subprocess.run(["curl", "-s", "-A", UA, "--max-time", str(timeout),
                        "-w", "\n%{http_code}", url], capture_output=True, text=True)
    body, _, code = r.stdout.rpartition("\n")
    return body, (code.strip() or "000")


def site_row(domain):
    rows = q(f"SELECT id, COALESCE(github_repo,''), COALESCE(deploy_config->>'target','') "
             f"FROM sites WHERE domain = '{domain}'", width=3)
    if not rows:
        sys.exit(f"no sites row for {domain}")
    return rows[0]


def live_pages(site_id):
    return q(f"""SELECT url, to_char(GREATEST(updated_at, COALESCE(last_built_at, updated_at)),'YYYY-MM-DD')
                 FROM pages WHERE site_id='{site_id}' AND status='active'
                   AND deployed_at IS NOT NULL ORDER BY url""", width=2)


def page_facts(domain, path):
    """The page's OWN h1 and OWN first sentence. Never a description we invent."""
    body, code = fetch(f"https://{domain}{path}")
    if code != "200":
        return None, None, code
    b = re.sub(r"(?is)<(script|style)[^>]*>.*?</\1>", " ", body)
    m = re.search(r"(?is)<h1[^>]*>(.*?)</h1>", b)
    title = html.unescape(re.sub(r"<[^>]+>", "", m.group(1))).strip() if m else None
    main = re.search(r"(?is)<main.*?</main>", b)
    text = re.sub(r"\s+", " ", html.unescape(re.sub(r"<[^>]+>", " ", main.group(0) if main else ""))).strip()
    if title:
        text = text.replace(title, "", 1).strip()
    first = next((s for s in re.split(r"(?<=[.!?])\s+", text) if len(s) > 40), "")[:190]
    return title, first, code


def robots_txt(domain, signal):
    return f"""# {domain}
#
# Served from this site's deploy repo, NOT from Cloudflare's managed robots.txt.
#
# TRAP: Cloudflare PREPENDS its managed file to this one rather than yielding to it.
# Shipping this file does not by itself unblock anything — the dashboard settings
# "Block AI training bots" AND "Set your preference to block training in robots.txt"
# must both be off. Verify PER AGENT, never with a single curl: the file is served
# conditionally, so one fetch proves nothing about any particular crawler.

User-agent: *
Allow: /

# contentsignals.org — what this site permits, stated explicitly.
Content-Signal: {signal}

Sitemap: https://{domain}/sitemap.xml
"""


def sitemap_xml(domain, entries):
    out = ['<?xml version="1.0" encoding="UTF-8"?>',
           '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">']
    for path, lastmod in entries:
        out.append(f"  <url><loc>https://{domain}{path}</loc><lastmod>{lastmod}</lastmod></url>")
    out.append("</urlset>")
    return "\n".join(out) + "\n"


def llms_txt(domain, summary, groups, extras):
    out = [f"# {domain}", ""]
    if summary:
        out += [f"> {summary}", ""]
    for name, items in groups:
        if not items:
            continue
        out.append(f"## {name}")
        for path, title, first in items:
            out.append(f"- [{title}](https://{domain}{path})" + (f": {first}" if first else ""))
        out.append("")
    if extras:
        out.append("## Otros" if any(ord(c) > 127 for c in summary or "") else "## Other")
        out += extras + [""]
    return "\n".join(out)


def group_for(path):
    seg = [s for s in path.split("/") if s]
    return seg[0].replace("-", " ").replace(".html", "").title() if len(seg) > 1 else "Site"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("domain")
    ap.add_argument("--out")
    ap.add_argument("--write", action="store_true")
    ap.add_argument("--no-probe", action="store_true",
                    help="skip the 200 check. A sitemap listing a 404 is worse than none.")
    ap.add_argument("--signal", default="search=yes, ai-input=yes, ai-train=yes",
                    help="Content-Signal value. Refuse training with ai-train=no.")
    ap.add_argument("--summary", default="", help="one-line site summary for llms.txt")
    a = ap.parse_args()

    site_id, repo, target = site_row(a.domain)
    pages = live_pages(site_id)
    print(f"site {a.domain}  id={site_id}  repo={repo or '(none)'}  target={target or '(none)'}")
    print(f"live pages in DB: {len(pages)}")

    # --- rule 3: who is actually serving robots.txt? ------------------------------
    served, code = fetch(f"https://{a.domain}/robots.txt")
    if "BEGIN Cloudflare" in served:
        print("  !! robots.txt: Cloudflare's MANAGED block is being merged in. Shipping "
              "your own file will NOT unblock crawlers until the dashboard setting is off.")
        blocked = re.findall(r"(?im)^User-agent:\s*(\S+)\s*$\s*Disallow:\s*/\s*$", served)
        if blocked:
            print(f"     currently disallowed: {', '.join(sorted(set(blocked)))}")
    elif code == "404":
        print("  robots.txt: absent (404) — Cloudflare may inject a managed one at the edge.")
    else:
        print(f"  robots.txt: {code}, no Cloudflare block detected (good)")

    # --- rules 1 & 2: probe, then take the page's own words -----------------------
    entries, facts, dropped = [], [], []
    for path, lastmod in pages:
        if a.no_probe:
            entries.append((path, lastmod)); continue
        title, first, code = page_facts(a.domain, path)
        if code != "200":
            dropped.append((path, code)); continue
        entries.append((path, lastmod))
        if title:
            facts.append((path, title, first))

    print(f"probed: {len(entries)} fetchable, {len(dropped)} dropped")
    for p, c in dropped:
        print(f"  DROPPED {c} {p}   (would have been a 404 in your sitemap)")
    if not entries:
        sys.exit("nothing fetchable — refusing to write an empty sitemap")

    groups = {}
    for path, title, first in facts:
        groups.setdefault(group_for(path), []).append((path, title, first))
    ordered = sorted(groups.items(), key=lambda kv: (kv[0] == "Site", kv[0]))

    feed, fcode = fetch(f"https://{a.domain}/feed.xml")
    extras = [f"- [Sitemap](https://{a.domain}/sitemap.xml)"]
    if fcode == "200":
        extras.insert(0, f"- [RSS](https://{a.domain}/feed.xml)")

    files = {
        "robots.txt": robots_txt(a.domain, a.signal),
        "sitemap.xml": sitemap_xml(a.domain, entries),
        "llms.txt": llms_txt(a.domain, a.summary, ordered, extras),
    }

    outdir = a.out or (f"/home/ant/projects/vm-sites/{a.domain}" if target == "vm"
                       else f"./out/{a.domain}")
    if not a.write:
        print(f"\nDRY RUN — would write into {outdir}:")
        for n, c in files.items():
            print(f"  {n:<12} {len(c):>6} bytes")
        print("\nre-run with --write to emit. Committing and pushing is yours.")
        return
    os.makedirs(outdir, exist_ok=True)
    for n, c in files.items():
        open(os.path.join(outdir, n), "w", encoding="utf-8").write(c)
        print(f"wrote {os.path.join(outdir, n)} ({len(c)} bytes)")
    print("\nNOT committed and NOT deployed — deliberately. Review, then commit by pathspec.")


if __name__ == "__main__":
    main()
