#!/usr/bin/env python3
"""Register a .uk/.co.uk domain at Nominet over EPP. Stdlib only. DRY-RUN DEFAULT.

Written 2026-08-26 for the webdesign.uk domain service (P3 of
site_delivery_and_editor/BRIEF_2026-08-26_domain_find_register_point_service.md),
under the DESIGNCONSULT tag per the owner's interim ruling of 2026-08-26 (customer
domains move to the second TAG when Nominet grants it). Third sibling of
nominet-epp-ns-change.py / nominet-epp-domain-check.py: same framing, login, IPv4
pin, credential handling; read the first sibling for the connection caveats.

⚠ REGISTRATION COSTS MONEY (~GBP 4/year wholesale) and creates a real registry
object in the owner's name. DRY-RUN is the default and does everything EXCEPT the
create: login, domain:check (refuses if taken), registrant resolution, and prints
the exact create that would be sent. --apply performs it and verifies with
domain:info afterwards.

Registrant: per the owner ruling of 2026-08-21 the registrant is the OWNER until
an agreed sale. Pass --registrant <contact-id> explicitly, or
--registrant-from <domain> to reuse the registrant contact of a domain already on
the tag (e.g. idea.uk) — resolved live via domain:info, printed before anything
is sent, never guessed.

authInfo: the schema requires one; Nominet's .uk registry ignores it (transfers
are TAG/TAC-based). A random one is generated per run and NOT stored — do not
treat it as a credential.

Usage:
  ./nominet-epp-domain-register.py --tag DESIGNCONSULT \
      --password-file ~/.config/nominet/epp-password \
      --domain example.uk --registrant-from idea.uk \
      [--ns alexis.ns.cloudflare.com --ns leah.ns.cloudflare.com] \
      [--years 1] [--apply]
"""
import argparse, os, re, secrets, socket, ssl, struct, sys

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


def send(sock, xml: str, what: str, ok=(1000, 1001)) -> str:
    sock.sendall(frame(xml))
    resp = read_msg(sock)
    code = result_code(resp)
    print(f"  {what}: {code} {result_msg(resp)}", file=sys.stderr)
    if code not in ok:
        raise SystemExit(f"FAILED at {what} (code {code}) — nothing further attempted")
    return resp


def login_xml(tag, pw):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><login><clID>{tag}</clID><pw>{pw}</pw>
<options><version>1.0</version><lang>en</lang></options>
<svcs><objURI>{DOM}</objURI><objURI>{HOS}</objURI></svcs>
</login><clTRID>domain-register-login</clTRID></command></epp>"""


def check_xml(domain):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><check><domain:check xmlns:domain="{DOM}">
<domain:name>{domain}</domain:name></domain:check></check>
<clTRID>domain-register-check</clTRID></command></epp>"""


def info_xml(domain):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><info><domain:info xmlns:domain="{DOM}">
<domain:name>{domain}</domain:name></domain:info></info>
<clTRID>domain-register-info</clTRID></command></epp>"""


def create_xml(domain, years, registrant, ns, auth_pw):
    nsx = ""
    if ns:
        inner = "".join(f"<domain:hostObj>{h}</domain:hostObj>" for h in ns)
        nsx = f"<domain:ns>{inner}</domain:ns>"
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><create><domain:create xmlns:domain="{DOM}">
<domain:name>{domain}</domain:name>
<domain:period unit="y">{years}</domain:period>
{nsx}<domain:registrant>{registrant}</domain:registrant>
<domain:authInfo><domain:pw>{auth_pw}</domain:pw></domain:authInfo>
</domain:create></create><clTRID>domain-register-create</clTRID></command></epp>"""


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tag", required=True)
    ap.add_argument("--password-file")
    ap.add_argument("--server", default="epp.nominet.org.uk")
    ap.add_argument("--port", type=int, default=700)
    ap.add_argument("--domain", required=True)
    ap.add_argument("--years", type=int, default=1)
    ap.add_argument("--registrant", help="Nominet contact id for the registrant")
    ap.add_argument("--registrant-from", help="reuse the registrant of this existing tag domain")
    ap.add_argument("--ns", action="append", default=[],
                    help="nameserver host object (repeat); omit to register without NS and repoint later")
    ap.add_argument("--apply", action="store_true",
                    help="actually REGISTER (costs money). Default: dry-run everything short of the create")
    a = ap.parse_args()

    pw = os.environ.get("NOMINET_EPP_PW", "")
    if a.password_file:
        pw = open(os.path.expanduser(a.password_file)).read().strip()
    if not pw:
        sys.exit("no password: use --password-file or NOMINET_EPP_PW")
    if not (a.registrant or a.registrant_from):
        sys.exit("registrant required: --registrant <id> or --registrant-from <domain>")
    dom = a.domain.strip().lower().rstrip(".")
    if not re.fullmatch(r"[a-z0-9][a-z0-9-]*\.(co\.)?uk", dom):
        sys.exit(f"refusing '{dom}': the offer is .uk / .co.uk only (owner ruling 2026-08-21)")

    ctx = ssl.create_default_context()
    v4 = socket.getaddrinfo(a.server, a.port, socket.AF_INET, socket.SOCK_STREAM)[0][4]
    with socket.create_connection(v4, timeout=30) as raw:
        with ctx.wrap_socket(raw, server_hostname=a.server) as sock:
            print(f"connected to {a.server}:{a.port}", file=sys.stderr)
            read_msg(sock)  # greeting
            send(sock, login_xml(a.tag, pw), "login")

            resp = send(sock, check_xml(dom), f"domain:check {dom}")
            m = re.search(r'<domain:name[^>]*avail="([01])"', resp)
            if not m or m.group(1) != "1":
                reason = re.search(r"<domain:reason[^>]*>([^<]*)</domain:reason>", resp)
                raise SystemExit(f"REFUSING: {dom} is not available"
                                 + (f" ({reason.group(1)})" if reason else ""))
            print(f"AVAILABLE  {dom}")

            registrant = a.registrant
            if not registrant:
                ref = a.registrant_from.strip().lower().rstrip(".")
                info = send(sock, info_xml(ref), f"domain:info {ref} (registrant source)")
                rm = re.search(r"<domain:registrant>([^<]+)</domain:registrant>", info)
                if not rm:
                    raise SystemExit(f"could not read a registrant from {ref}")
                registrant = rm.group(1).strip()
            print(f"REGISTRANT {registrant} (owner-until-sale, ruling 2026-08-21)")

            auth_pw = secrets.token_urlsafe(12)
            xml = create_xml(dom, a.years, registrant, [h.lower().rstrip(".") for h in a.ns], auth_pw)
            if not a.apply:
                print("DRY-RUN: would send domain:create (~GBP 4/yr). Re-run with --apply to register.")
                print(re.sub(auth_pw, "<generated>", xml))
            else:
                send(sock, xml, f"domain:create {dom}")
                verify = send(sock, info_xml(dom), "verify domain:info")
                exp = re.search(r"<domain:exDate>([^<]+)</domain:exDate>", verify)
                print(f"REGISTERED {dom}"
                      + (f" expires {exp.group(1)}" if exp else " (no exDate in info — inspect above)"))
            sock.sendall(frame(
                f'<?xml version="1.0"?><epp xmlns="{EPP}"><command><logout/><clTRID>bye</clTRID></command></epp>'))


if __name__ == "__main__":
    main()
