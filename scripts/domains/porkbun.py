#!/usr/bin/env python3
"""Porkbun registrar API helper — list domains, read/repoint nameservers, DNS CRUD.

Part of the domains_cloudflare_rollout lane's registrar tooling (see
RUNBOOK_domains_cloudflare_rollout.md, "registrar credentials" table). Porkbun was
one of the three registrar keys owed to that lane; this script is the client for it.

Credentials live in ~/.config/porkbun/credentials as API_KEY=… and SECRET_API_KEY=…
lines (mode 600). Values are whitespace-stripped on read — the RUNBOOK's gotcha: a
trailing newline in a credentials file becomes an auth failure that looks like a bad
key. This script never prints key material: not in output, not in error paths (the
request body carries the keys, so errors report the endpoint and the response only).

Endpoint reference: https://porkbun.com/llms-full.txt. All calls are POSTs with the
keys in the JSON body. Per-domain endpoints (dns/*, getNs, updateNs) require the
domain's own "API ACCESS" toggle to be ON in Porkbun's Domain Management — the error
for a disabled domain is explicit, so it cannot be mistaken for a bad key.

Usage:
  porkbun.py ping                    # auth check; prints your egress IP
  porkbun.py domains [--json]        # all domains, TSV (or raw JSON), paginated
  porkbun.py ns <domain>             # current nameserver delegation
  porkbun.py set-ns <domain> <ns1> <ns2> [...]
  porkbun.py dns <domain> [--json]   # all DNS records, TSV (or raw JSON)
  porkbun.py dns-create <domain> --type A --content 1.2.3.4 [--name www] [--ttl 600] [--prio N]
  porkbun.py dns-edit <domain> --id <id> --type A --content 1.2.3.4 [--name www] [--ttl 600] [--prio N]
  porkbun.py dns-delete <domain> --id <id>
  porkbun.py check <domain>          # availability/price (heavily rate-limited upstream)
  porkbun.py raw <endpoint> [--data '{"k":"v"}']   # escape hatch, e.g. raw /pricing/get
"""
import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request

BASE = "https://api.porkbun.com/api/json/v3"
CREDENTIALS = os.path.expanduser("~/.config/porkbun/credentials")
UA = "Mozilla/5.0 (compatible; agentchassis-domains/1.0)"
LIST_CHUNK = 1000  # documented page size for /domain/listAll


def load_credentials():
    if not os.path.exists(CREDENTIALS):
        sys.exit(f"porkbun: no credentials file at {CREDENTIALS} — see the "
                 "registrar credentials table in RUNBOOK_domains_cloudflare_rollout.md")
    values = {}
    with open(CREDENTIALS) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, _, val = line.partition("=")
            values[key.strip()] = val.strip()  # strip: trailing whitespace breaks auth
    api_key = values.get("API_KEY", "")
    secret = values.get("SECRET_API_KEY", "")
    if not api_key or not secret or "PASTE" in api_key or "PASTE" in secret:
        sys.exit(f"porkbun: {CREDENTIALS} still holds placeholder values — "
                 "create a key at https://porkbun.com/account/api and fill both lines in")
    return api_key, secret


def call(endpoint, extra=None):
    api_key, secret = load_credentials()
    body = {"apikey": api_key, "secretapikey": secret}
    if extra:
        body.update(extra)
    req = urllib.request.Request(
        BASE + endpoint,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json", "User-Agent": UA},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            payload = json.load(resp)
    except urllib.error.HTTPError as e:
        # Porkbun returns JSON bodies on 4xx/5xx; surface the message, never the request.
        try:
            payload = json.load(e)
        except Exception:
            sys.exit(f"porkbun: HTTP {e.code} on {endpoint} (no parseable body)")
    except urllib.error.URLError as e:
        sys.exit(f"porkbun: cannot reach {BASE}{endpoint}: {e.reason}")
    if payload.get("status") != "SUCCESS":
        sys.exit(f"porkbun: ERROR on {endpoint}: {payload.get('message', payload)}")
    return payload


def cmd_ping(_args):
    print(f"SUCCESS — authenticated; egress IP {call('/ping').get('yourIp', '?')}")


def cmd_domains(args):
    domains, start = [], 0
    while True:
        chunk = call("/domain/listAll",
                     {"start": str(start), "includeLabels": "yes"}).get("domains", [])
        domains.extend(chunk)
        if len(chunk) < LIST_CHUNK:
            break
        start += len(chunk)
        time.sleep(1)  # stay under the API's rate limit across pages
    if args.json:
        json.dump(domains, sys.stdout, indent=2)
        print()
        return
    print("domain\tstatus\texpires\tauto_renew\tlabels")
    for d in domains:
        labels = ",".join(l.get("title", "") for l in d.get("labels", []) or [])
        print(f"{d.get('domain')}\t{d.get('status')}\t{d.get('expireDate')}"
              f"\t{d.get('autoRenew')}\t{labels}")
    print(f"# {len(domains)} domains", file=sys.stderr)


def cmd_ns(args):
    for ns in call(f"/domain/getNs/{args.domain}").get("ns", []):
        print(ns)


def cmd_set_ns(args):
    call(f"/domain/updateNs/{args.domain}", {"ns": args.nameservers})
    print(f"SUCCESS — {args.domain} delegated to: {', '.join(args.nameservers)}")


def record_body(args):
    body = {"type": args.type, "content": args.content}
    if args.name is not None:
        body["name"] = args.name
    if args.ttl is not None:
        body["ttl"] = str(args.ttl)
    if args.prio is not None:
        body["prio"] = str(args.prio)
    return body


def cmd_dns(args):
    records = call(f"/dns/retrieve/{args.domain}").get("records", [])
    if args.json:
        json.dump(records, sys.stdout, indent=2)
        print()
        return
    print("id\ttype\tname\tcontent\tttl\tprio")
    for r in records:
        print(f"{r.get('id')}\t{r.get('type')}\t{r.get('name')}\t{r.get('content')}"
              f"\t{r.get('ttl')}\t{r.get('prio')}")


def cmd_dns_create(args):
    payload = call(f"/dns/create/{args.domain}", record_body(args))
    print(f"SUCCESS — record id {payload.get('id')}")


def cmd_dns_edit(args):
    call(f"/dns/edit/{args.domain}/{args.id}", record_body(args))
    print("SUCCESS")


def cmd_dns_delete(args):
    call(f"/dns/delete/{args.domain}/{args.id}")
    print("SUCCESS — record deleted")


def cmd_check(args):
    json.dump(call(f"/domain/checkDomain/{args.domain}").get("response", {}),
              sys.stdout, indent=2)
    print()


def cmd_raw(args):
    extra = json.loads(args.data) if args.data else None
    json.dump(call(args.endpoint, extra), sys.stdout, indent=2)
    print()


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = p.add_subparsers(dest="command", required=True)

    sub.add_parser("ping").set_defaults(func=cmd_ping)

    sp = sub.add_parser("domains")
    sp.add_argument("--json", action="store_true")
    sp.set_defaults(func=cmd_domains)

    sp = sub.add_parser("ns")
    sp.add_argument("domain")
    sp.set_defaults(func=cmd_ns)

    sp = sub.add_parser("set-ns")
    sp.add_argument("domain")
    sp.add_argument("nameservers", nargs="+")
    sp.set_defaults(func=cmd_set_ns)

    sp = sub.add_parser("dns")
    sp.add_argument("domain")
    sp.add_argument("--json", action="store_true")
    sp.set_defaults(func=cmd_dns)

    for name, func, needs_id in (("dns-create", cmd_dns_create, False),
                                 ("dns-edit", cmd_dns_edit, True)):
        sp = sub.add_parser(name)
        sp.add_argument("domain")
        if needs_id:
            sp.add_argument("--id", required=True)
        sp.add_argument("--type", required=True)
        sp.add_argument("--content", required=True)
        sp.add_argument("--name", default=None, help="subdomain; omit for the apex")
        sp.add_argument("--ttl", type=int, default=None)
        sp.add_argument("--prio", type=int, default=None)
        sp.set_defaults(func=func)

    sp = sub.add_parser("dns-delete")
    sp.add_argument("domain")
    sp.add_argument("--id", required=True)
    sp.set_defaults(func=cmd_dns_delete)

    sp = sub.add_parser("check")
    sp.add_argument("domain")
    sp.set_defaults(func=cmd_check)

    sp = sub.add_parser("raw")
    sp.add_argument("endpoint", help="e.g. /pricing/get or /domain/getNs/example.com")
    sp.add_argument("--data", default=None, help="extra JSON merged into the request body")
    sp.set_defaults(func=cmd_raw)

    args = p.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
