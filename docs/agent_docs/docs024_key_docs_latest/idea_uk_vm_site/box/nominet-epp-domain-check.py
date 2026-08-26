#!/usr/bin/env python3
"""Check .uk/.co.uk domain availability at Nominet over EPP. Stdlib only.

Written 2026-08-26 for the webdesign.uk domain service (P1 of
site_delivery_and_editor/BRIEF_2026-08-26_domain_find_register_point_service.md):
the intake needs to offer AVAILABLE names, and EPP domain:check is the registry's
own answer. Sibling of nominet-epp-ns-change.py — same framing, login, IPv4 pin
and credential handling; read it for the connection caveats (IP allow-list,
unframed pre-auth refusals).

READ-ONLY BY NATURE: domain:check writes nothing, so there is no --apply here.
The dry-run-is-the-norm rule is satisfied by the command itself.

Names are checked in chunks (default 5 per command — conservative against
registry batch caps) and printed one per line:

  AVAILABLE  example.uk
  TAKEN      example.co.uk  (registered)

Exit code 0 if every name got an answer; 1 on any refusal mid-run.

Usage:
  ./nominet-epp-domain-check.py --tag DESIGNCONSULT \
      --password-file ~/.config/nominet/epp-password \
      example.uk example.co.uk another.uk
"""
import argparse, os, re, socket, ssl, struct, sys

EPP = "urn:ietf:params:xml:ns:epp-1.0"
DOM = "urn:ietf:params:xml:ns:domain-1.0"
HOS = "urn:ietf:params:xml:ns:host-1.0"


def frame(xml: str) -> bytes:
    body = xml.encode()
    return struct.pack(">I", len(body) + 4) + body


def read_msg(sock) -> str:
    hdr = b""
    while len(hdr) < 4:
        chunk = sock.recv(4 - len(hdr))
        if not chunk:
            raise ConnectionError("EPP server closed the connection")
        hdr += chunk
    need = struct.unpack(">I", hdr)[0] - 4
    body = b""
    while len(body) < need:
        chunk = sock.recv(need - len(body))
        if not chunk:
            raise ConnectionError(
                f"EPP server closed mid-message; received so far: {(hdr + body)!r}")
        body += chunk
    return body.decode()


def result_code(resp: str) -> int:
    m = re.search(r'<result code="(\d+)"', resp)
    return int(m.group(1)) if m else -1


def result_msg(resp: str) -> str:
    m = re.search(r"<msg[^>]*>(.*?)</msg>", resp, re.S)
    return (m.group(1).strip() if m else resp[:300]).replace("\n", " ")


def login_xml(tag, pw):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><login><clID>{tag}</clID><pw>{pw}</pw>
<options><version>1.0</version><lang>en</lang></options>
<svcs><objURI>{DOM}</objURI><objURI>{HOS}</objURI></svcs>
</login><clTRID>domain-check-login</clTRID></command></epp>"""


def check_xml(domains):
    names = "".join(f"<domain:name>{d}</domain:name>" for d in domains)
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><check><domain:check xmlns:domain="{DOM}">
{names}</domain:check></check><clTRID>domain-check</clTRID></command></epp>"""


def parse_check(resp: str):
    """Yield (name, available: bool, reason: str) per <domain:cd> block."""
    for cd in re.findall(r"<domain:cd>(.*?)</domain:cd>", resp, re.S):
        m = re.search(r'<domain:name[^>]*avail="([01])"[^>]*>([^<]+)</domain:name>', cd)
        if not m:
            continue
        reason = re.search(r"<domain:reason[^>]*>([^<]*)</domain:reason>", cd)
        yield m.group(2).strip(), m.group(1) == "1", (reason.group(1).strip() if reason else "")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tag", required=True)
    ap.add_argument("--password-file")
    ap.add_argument("--server", default="epp.nominet.org.uk")
    ap.add_argument("--port", type=int, default=700)
    ap.add_argument("--chunk", type=int, default=5,
                    help="names per check command (default 5)")
    ap.add_argument("domains", nargs="+", help="domain names to check")
    a = ap.parse_args()

    pw = os.environ.get("NOMINET_EPP_PW", "")
    if a.password_file:
        pw = open(os.path.expanduser(a.password_file)).read().strip()
    if not pw:
        sys.exit("no password: use --password-file or NOMINET_EPP_PW")

    names = [d.strip().lower().rstrip(".") for d in a.domains]
    bad = [d for d in names if not re.fullmatch(r"[a-z0-9][a-z0-9-]*(\.[a-z0-9][a-z0-9-]*)+", d)]
    if bad:
        sys.exit(f"refusing malformed name(s): {bad}")

    ctx = ssl.create_default_context()
    # IPv4 only — same reason as the sibling script: the allow-list is v4.
    v4 = socket.getaddrinfo(a.server, a.port, socket.AF_INET, socket.SOCK_STREAM)[0][4]
    answered = {}
    with socket.create_connection(v4, timeout=30) as raw:
        with ctx.wrap_socket(raw, server_hostname=a.server) as sock:
            print(f"connected to {a.server}:{a.port}", file=sys.stderr)
            read_msg(sock)  # greeting
            sock.sendall(frame(login_xml(a.tag, pw)))
            resp = read_msg(sock)
            code = result_code(resp)
            print(f"  login: {code} {result_msg(resp)}", file=sys.stderr)
            if code not in (1000, 1001):
                raise SystemExit(f"FAILED at login (code {code})")
            for i in range(0, len(names), max(1, a.chunk)):
                chunk = names[i:i + max(1, a.chunk)]
                sock.sendall(frame(check_xml(chunk)))
                resp = read_msg(sock)
                code = result_code(resp)
                if code not in (1000, 1001):
                    print(f"  domain:check refused: {code} {result_msg(resp)}", file=sys.stderr)
                    break
                for name, avail, reason in parse_check(resp):
                    answered[name] = (avail, reason)
            sock.sendall(frame(
                f'<?xml version="1.0"?><epp xmlns="{EPP}"><command><logout/><clTRID>bye</clTRID></command></epp>'))
    missing = [d for d in names if d not in answered]
    for d in names:
        if d in answered:
            avail, reason = answered[d]
            tag = "AVAILABLE" if avail else "TAKEN    "
            print(f"{tag}  {d}" + (f"  ({reason})" if reason else ""))
    if missing:
        print(f"NO ANSWER for: {missing}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
