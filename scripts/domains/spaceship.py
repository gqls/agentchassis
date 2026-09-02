#!/usr/bin/env python3
"""Spaceship registrar API helper — list domains, read/repoint nameservers, DNS CRUD.

Part of the domains_cloudflare_rollout lane's registrar tooling (see
RUNBOOK_domains_cloudflare_rollout.md, "Credentials" table and "Spaceship" section).
Spaceship was one of the three registrar keys owed to that lane; key landed and the
read path was proven 2026-09-02 (203 domains).

Credentials live in ~/.config/spaceship/credentials as API_KEY=… and API_SECRET=…
lines (mode 600), created in the dashboard's API Manager
(https://www.spaceship.com/application/api-manager/). Values are whitespace-stripped
on read — the RUNBOOK's gotcha: a trailing newline in a credentials file becomes an
auth failure that looks like a bad key. This script never prints key material: the
keys travel only in request headers (X-Api-Key / X-Api-Secret), and error paths
report the endpoint and response body only.

API base https://spaceship.dev/api/v1 (note: api.spaceship.dev does NOT resolve —
the API host is spaceship.dev itself, measured 2026-08-03). Errors arrive as JSON
with a "detail" field on a 4xx status. Upstream rate limits (docs, 2026-09-02):
domain list 300 req/300s per user; nameserver updates 5 per DOMAIN per 300s — a
bulk repoint must pace itself; most other ops 30/30s per user.

Usage:
  spaceship.py domains [--json]      # all domains, TSV (or raw JSON), paginated
  spaceship.py info <domain>         # full domain object, JSON
  spaceship.py ns <domain>           # current delegation (provider + hosts)
  spaceship.py set-ns <domain> <ns1> <ns2> [...]   # provider becomes "custom"
  spaceship.py set-ns <domain> basic               # back to Spaceship's own NS
  spaceship.py dns <domain> [--json] # all DNS records, TSV (or raw JSON)
  spaceship.py dns-put <domain> --items '<json array>' [--force]
  spaceship.py dns-delete <domain> --items '<json array>'
  spaceship.py raw <METHOD> <path> [--data '<json>']  # e.g. raw GET /domains?take=5&skip=0

DNS item shapes are Spaceship's own (field names vary by type — A/AAAA use
"address", not "content"): pass them through --items verbatim rather than trusting
a translation layer. dns-put with --force disables the API's conflict checker.
"""
import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request

BASE = "https://spaceship.dev/api/v1"
CREDENTIALS = os.path.expanduser("~/.config/spaceship/credentials")
UA = "Mozilla/5.0 (compatible; agentchassis-domains/1.0)"
LIST_CHUNK = 100   # /domains page size proven 2026-09-02
DNS_CHUNK = 500    # /dns/records documented take max


def load_credentials():
    if not os.path.exists(CREDENTIALS):
        sys.exit(f"spaceship: no credentials file at {CREDENTIALS} — see the "
                 "Credentials table in RUNBOOK_domains_cloudflare_rollout.md")
    values = {}
    with open(CREDENTIALS) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, _, val = line.partition("=")
            values[key.strip()] = val.strip()  # strip: trailing whitespace breaks auth
    api_key = values.get("API_KEY", "")
    secret = values.get("API_SECRET", "")
    if not api_key or not secret or "PASTE" in api_key or "PASTE" in secret:
        sys.exit(f"spaceship: {CREDENTIALS} still holds placeholder values — create a "
                 "key at https://www.spaceship.com/application/api-manager/ and fill both lines in")
    return api_key, secret


def call(method, path, body=None):
    api_key, secret = load_credentials()
    req = urllib.request.Request(
        BASE + path,
        data=json.dumps(body).encode() if body is not None else None,
        headers={"X-Api-Key": api_key, "X-Api-Secret": secret,
                 "Content-Type": "application/json", "User-Agent": UA},
        method=method,
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            raw = resp.read()
            return json.loads(raw) if raw.strip() else {}  # writes return 204/empty
    except urllib.error.HTTPError as e:
        # Spaceship returns a JSON body with "detail" on 4xx; surface it, never the request.
        detail = e.read().decode(errors="replace")[:2000]
        sys.exit(f"spaceship: HTTP {e.code} on {method} {path}: {detail}")
    except urllib.error.URLError as e:
        sys.exit(f"spaceship: cannot reach {BASE}{path}: {e.reason}")


def paginate(path_prefix, chunk):
    items, skip = [], 0
    while True:
        page = call("GET", f"{path_prefix}take={chunk}&skip={skip}").get("items", [])
        items.extend(page)
        if len(page) < chunk:
            return items
        skip += len(page)
        time.sleep(1)  # stay under the per-user rate limit across pages


def ns_summary(domain_obj):
    ns = domain_obj.get("nameservers", {}) or {}
    return ns.get("provider", "?"), ns.get("hosts", []) or []


def cmd_domains(args):
    domains = paginate("/domains?", LIST_CHUNK)
    if args.json:
        json.dump(domains, sys.stdout, indent=2)
        print()
        return
    print("domain\tstatus\texpires\tauto_renew\tns_provider\tns_hosts")
    for d in domains:
        provider, hosts = ns_summary(d)
        print(f"{d.get('name')}\t{d.get('lifecycleStatus')}"
              f"\t{(d.get('expirationDate') or '')[:10]}\t{d.get('autoRenew')}"
              f"\t{provider}\t{','.join(hosts)}")
    print(f"# {len(domains)} domains", file=sys.stderr)


def cmd_info(args):
    json.dump(call("GET", f"/domains/{args.domain}"), sys.stdout, indent=2)
    print()


def cmd_ns(args):
    provider, hosts = ns_summary(call("GET", f"/domains/{args.domain}"))
    print(f"provider: {provider}")
    for h in hosts:
        print(h)


def cmd_set_ns(args):
    if args.nameservers == ["basic"]:
        body = {"provider": "basic"}  # hosts must be omitted for "basic" (docs 2026-09-02)
        target = "Spaceship basic nameservers"
    else:
        body = {"provider": "custom", "hosts": args.nameservers}
        target = ", ".join(args.nameservers)
    call("PUT", f"/domains/{args.domain}/nameservers", body)
    # Rate limit is 5 NS updates per domain per 300s — verify by re-read, don't retry blind.
    provider, hosts = ns_summary(call("GET", f"/domains/{args.domain}"))
    print(f"requested: {target}\nnow: provider={provider} hosts={','.join(hosts)}")


def cmd_dns(args):
    records = paginate(f"/dns/records/{args.domain}?", DNS_CHUNK)
    if args.json:
        json.dump(records, sys.stdout, indent=2)
        print()
        return
    print("type\tname\tvalue\tttl")
    for r in records:
        value = r.get("address") or r.get("cname") or r.get("value") or \
            r.get("exchange") or json.dumps({k: v for k, v in r.items()
                                             if k not in ("type", "name", "ttl")})
        print(f"{r.get('type')}\t{r.get('name')}\t{value}\t{r.get('ttl')}")
    print(f"# {len(records)} records", file=sys.stderr)


def cmd_dns_put(args):
    call("PUT", f"/dns/records/{args.domain}",
         {"force": args.force, "items": json.loads(args.items)})
    print("SUCCESS — re-read with: spaceship.py dns " + args.domain)


def cmd_dns_delete(args):
    call("DELETE", f"/dns/records/{args.domain}", json.loads(args.items))
    print("SUCCESS — records deleted (matching is case-insensitive except TXT)")


def cmd_raw(args):
    result = call(args.method.upper(), args.path,
                  json.loads(args.data) if args.data else None)
    json.dump(result, sys.stdout, indent=2)
    print()


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = p.add_subparsers(dest="command", required=True)

    sp = sub.add_parser("domains")
    sp.add_argument("--json", action="store_true")
    sp.set_defaults(func=cmd_domains)

    sp = sub.add_parser("info")
    sp.add_argument("domain")
    sp.set_defaults(func=cmd_info)

    sp = sub.add_parser("ns")
    sp.add_argument("domain")
    sp.set_defaults(func=cmd_ns)

    sp = sub.add_parser("set-ns")
    sp.add_argument("domain")
    sp.add_argument("nameservers", nargs="+",
                    help="nameserver hosts, or the single word 'basic' to revert to Spaceship NS")
    sp.set_defaults(func=cmd_set_ns)

    sp = sub.add_parser("dns")
    sp.add_argument("domain")
    sp.add_argument("--json", action="store_true")
    sp.set_defaults(func=cmd_dns)

    sp = sub.add_parser("dns-put")
    sp.add_argument("domain")
    sp.add_argument("--items", required=True, help="JSON array of record objects")
    sp.add_argument("--force", action="store_true",
                    help="disable the API's conflict-resolution checker")
    sp.set_defaults(func=cmd_dns_put)

    sp = sub.add_parser("dns-delete")
    sp.add_argument("domain")
    sp.add_argument("--items", required=True, help="JSON array of record objects to match")
    sp.set_defaults(func=cmd_dns_delete)

    sp = sub.add_parser("raw")
    sp.add_argument("method", choices=["GET", "PUT", "POST", "DELETE", "get", "put", "post", "delete"])
    sp.add_argument("path", help="e.g. /domains?take=5&skip=0")
    sp.add_argument("--data", help="JSON request body")
    sp.set_defaults(func=cmd_raw)

    args = p.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
