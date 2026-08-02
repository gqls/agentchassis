#!/usr/bin/env python3
"""Change a .uk domain's nameservers at Nominet over EPP. Stdlib only.

Written 2026-08-02 for the idea.uk Option B cutover (Hetzner NS -> Cloudflare
alexis/leah), but generic over any domain on the tag. DEFAULTS TO DRY-RUN:
connects, logs in, prints the domain's current state and what would change.
Nothing is written without --apply (migration-runner practice: the dry run is
the norm, the apply is the exception).

Credentials: the EPP password comes from --password-file or $NOMINET_EPP_PW,
never argv (argv is visible in ps). The client id is the registrar TAG.

⚠ Nominet EPP only accepts connections from IPs registered in Online Services
(Settings -> EPP -> IP addresses). A connection refused/reset/timeout here is
almost always the allow-list, not the credentials — fix it in the portal, or
run this script from a whitelisted box (e.g. over ssh with the password file
scp'd there and removed after).

⚠ Host objects: EPP references nameservers as registry host objects. If a
target host is unknown to the registry this script host:create's it and
retries once (result 2303 on update). Out-of-zone hosts need no glue.

Usage:
  ./nominet-epp-ns-change.py --tag DESIGNCONSULT --domain idea.uk \
      --ns alexis.ns.cloudflare.com --ns leah.ns.cloudflare.com \
      --password-file ~/.config/nominet/epp-password [--apply]
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
            raise ConnectionError("EPP server closed mid-message")
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
    print(f"  {what}: {code} {result_msg(resp)}")
    if code not in ok:
        raise SystemExit(f"FAILED at {what} (code {code}) — nothing further attempted")
    return resp


def login_xml(tag, pw):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><login><clID>{tag}</clID><pw>{pw}</pw>
<options><version>1.0</version><lang>en</lang></options>
<svcs><objURI>{DOM}</objURI><objURI>{HOS}</objURI></svcs>
</login><clTRID>ns-change-login</clTRID></command></epp>"""


def info_xml(domain):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><info><domain:info xmlns:domain="{DOM}">
<domain:name>{domain}</domain:name></domain:info></info>
<clTRID>ns-change-info</clTRID></command></epp>"""


def update_xml(domain, add, rem):
    def block(hosts):
        inner = "".join(f"<domain:hostObj>{h}</domain:hostObj>" for h in hosts)
        return f"<domain:ns>{inner}</domain:ns>"
    addx = f"<domain:add>{block(add)}</domain:add>" if add else ""
    remx = f"<domain:rem>{block(rem)}</domain:rem>" if rem else ""
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><update><domain:update xmlns:domain="{DOM}">
<domain:name>{domain}</domain:name>{addx}{remx}</domain:update></update>
<clTRID>ns-change-update</clTRID></command></epp>"""


def host_create_xml(host):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><create><host:create xmlns:host="{HOS}">
<host:name>{host}</host:name></host:create></create>
<clTRID>ns-change-hostcreate</clTRID></command></epp>"""


def current_ns(info_resp):
    return sorted(h.lower().rstrip(".") for h in re.findall(r"<domain:hostObj>([^<]+)</domain:hostObj>", info_resp))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tag", required=True)
    ap.add_argument("--domain", required=True)
    ap.add_argument("--ns", action="append", required=True, help="target nameserver (repeat)")
    ap.add_argument("--password-file")
    ap.add_argument("--server", default="epp.nominet.org.uk")
    ap.add_argument("--port", type=int, default=700)
    ap.add_argument("--apply", action="store_true", help="actually send the update (default: dry-run)")
    a = ap.parse_args()

    pw = os.environ.get("NOMINET_EPP_PW", "")
    if a.password_file:
        pw = open(os.path.expanduser(a.password_file)).read().strip()
    if not pw:
        sys.exit("no password: use --password-file or NOMINET_EPP_PW")

    target = sorted(h.lower().rstrip(".") for h in a.ns)
    ctx = ssl.create_default_context()
    with socket.create_connection((a.server, a.port), timeout=30) as raw:
        with ctx.wrap_socket(raw, server_hostname=a.server) as sock:
            print(f"connected to {a.server}:{a.port}")
            read_msg(sock)  # greeting
            send(sock, login_xml(a.tag, pw), "login")
            info = send(sock, info_xml(a.domain), f"domain:info {a.domain}")
            have = current_ns(info)
            print(f"  current NS: {have}\n  target  NS: {target}")
            if have == target:
                print("already at target — nothing to do")
            elif not a.apply:
                print(f"DRY-RUN: would rem {sorted(set(have)-set(target))} add {sorted(set(target)-set(have))}")
                print("re-run with --apply to execute")
            else:
                add, rem = sorted(set(target) - set(have)), sorted(set(have) - set(target))
                try:
                    send(sock, update_xml(a.domain, add, rem), "domain:update")
                except SystemExit:
                    # most likely 2303: a target host object does not exist yet
                    print("  update refused — creating missing host objects and retrying once")
                    for h in add:
                        sock.sendall(frame(host_create_xml(h)))
                        r = read_msg(sock)
                        print(f"  host:create {h}: {result_code(r)} {result_msg(r)}")
                    send(sock, update_xml(a.domain, add, rem), "domain:update (retry)")
                verify = send(sock, info_xml(a.domain), "verify domain:info")
                now = current_ns(verify)
                print(f"  NS after update: {now}")
                print("SUCCESS" if now == target else "MISMATCH — inspect above; registry may apply asynchronously")
            sock.sendall(frame(f'<?xml version="1.0"?><epp xmlns="{EPP}"><command><logout/><clTRID>bye</clTRID></command></epp>'))


if __name__ == "__main__":
    main()
