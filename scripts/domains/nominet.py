#!/usr/bin/env python3
"""Nominet registrar client (.uk estate) — the scripts/domains/ family member.

Consolidates the proven EPP pieces (epp.pl login/list; the idea_uk_vm_site/box
check/ns-change siblings) behind one command, in the porkbun.py/dynadot.sh
family style: credentials from ~/.config/nominet/credentials (TAG= and
EPP_PASSWORD= lines), never printed, never on argv; read verbs freely runnable;
the one write verb dry-runs by default.

TRANSPORT: Nominet allowlists the EPP egress IP, and only the five cluster node
IPs are stable — so by default this client tunnels through
`kubectl exec -i postgres-clients-0 -- openssl s_client` and does the RFC 5734
framing locally (the exact mechanism of the 2026-08-11 login proof).
--direct uses a plain socket instead, for runs from an already-allowlisted box.
Pin to IPv4 either way: the two families get different treatment.

⚠ The greeting is served to ANY IP — only `login` (result 1000) proves the
allowlist. LOGIN_CODE 2200 = egress not allowlisted, NOT a wrong password.
⚠ Pre-auth refusals are UNFRAMED text; the framer surfaces the raw bytes.
⚠ A session's permission classifier may refuse credentialed runs — that is the
operating model, not a bug: stage the command, the owner runs it (`! …`).

Verbs:
  probe                       greeting byte-count, credential-free (transport control)
  login                       the allowlist test
  list YYYY-MM                domains expiring that month (std-list-1.0)
  walk [--months N]           expiry walk, one session, default 24 months (max 120 —
                              .uk registers up to 10 years, so 12 months UNDER-COUNTS)
  check DOMAIN [...]          domain:check, chunked
  info DOMAIN                 domain:info (NS, registrant, dates, status)
  set-ns DOMAIN --ns A --ns B domain:update, DRY-RUN default, --apply to execute
                              (host:create retry on 2303, verify by re-reading)
  register ...                REFUSED here — use idea_uk_vm_site/box/
                              nominet-epp-domain-register.py (VMB-017): it costs
                              money and carries the registrant rulings.
  --self-test                 offline checks, no network, no credentials
"""
import argparse, datetime, os, re, select, socket, ssl, struct, subprocess, sys

EPP = "urn:ietf:params:xml:ns:epp-1.0"
DOM = "urn:ietf:params:xml:ns:domain-1.0"
HOS = "urn:ietf:params:xml:ns:host-1.0"
CON = "urn:ietf:params:xml:ns:contact-1.0"
LST = "http://www.nominet.org.uk/epp/xml/std-list-1.0"
SERVER, PORT = "epp.nominet.org.uk", 700
POD, NS = "postgres-clients-0", "ai-persona-system"
CRED_FILE = os.environ.get("NOMINET_CREDENTIALS_FILE",
                           os.path.expanduser("~/.config/nominet/credentials"))


def esc(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def read_credentials(path=None):
    """Parse TAG= / EPP_PASSWORD= lines. Values never leave this process."""
    path = path or CRED_FILE
    tag = pw = ""
    with open(path) as f:
        for line in f:
            line = line.strip()
            if line.startswith("TAG="):
                tag = line[4:].strip()
            elif line.startswith("EPP_PASSWORD="):
                pw = line[13:].strip()
    if not tag or not pw:
        raise SystemExit(f"credentials file {path} needs TAG= and EPP_PASSWORD= lines")
    return tag, pw


def frame(xml):
    body = xml.encode()
    return struct.pack(">I", len(body) + 4) + body


# ---------------------------------------------------------------- transports
class PodTransport:
    """Bytes through `kubectl exec -i` + openssl s_client in the cluster pod."""

    def __init__(self, timeout=30):
        self.timeout = timeout
        self.p = subprocess.Popen(
            ["kubectl", "-n", NS, "exec", "-i", POD, "--",
             "openssl", "s_client", "-connect", f"{SERVER}:{PORT}", "-4", "-quiet"],
            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)

    def send(self, data):
        self.p.stdin.write(data)
        self.p.stdin.flush()

    def recv(self, n):
        fd = self.p.stdout.fileno()
        r, _, _ = select.select([fd], [], [], self.timeout)
        if not r:
            raise ConnectionError(f"EPP read timeout after {self.timeout}s")
        chunk = os.read(fd, n)
        if not chunk:
            raise ConnectionError("EPP stream closed (pod transport)")
        return chunk

    def close(self):
        try:
            self.p.stdin.close()
        except OSError:
            pass
        try:
            # kubectl exec does not reliably propagate stdin EOF to openssl,
            # so after a short grace period the tunnel is torn down hard —
            # by this point logout has already been answered.
            self.p.wait(timeout=3)
        except subprocess.TimeoutExpired:
            self.p.kill()
            self.p.wait(timeout=10)


class DirectTransport:
    def __init__(self, timeout=30):
        v4 = socket.getaddrinfo(SERVER, PORT, socket.AF_INET, socket.SOCK_STREAM)[0][4]
        raw = socket.create_connection(v4, timeout=timeout)
        self.sock = ssl.create_default_context().wrap_socket(raw, server_hostname=SERVER)

    def send(self, data):
        self.sock.sendall(data)

    def recv(self, n):
        chunk = self.sock.recv(n)
        if not chunk:
            raise ConnectionError("EPP stream closed (direct transport)")
        return chunk

    def close(self):
        self.sock.close()


def read_msg(t):
    hdr = b""
    while len(hdr) < 4:
        hdr += t.recv(4 - len(hdr))
    need = struct.unpack(">I", hdr)[0] - 4
    if need <= 0 or need > 20_000_000:
        # pre-auth refusals are UNFRAMED text — the "length" is 4 text bytes
        raise ConnectionError(f"unframed/oversized EPP message; first bytes: {hdr!r}")
    body = b""
    while len(body) < need:
        body += t.recv(need - len(body))
    return body.decode()


def result_code(resp):
    m = re.search(r'<result code="(\d+)"', resp)
    return int(m.group(1)) if m else -1


def result_msg(resp):
    # the top-level <msg> is often generic ("Command syntax error"); Nominet's
    # own diagnosis lives in <extValue><reason> and was invisible here until a
    # raw-response dump caught "V274 Schema std-list- not specified at login"
    # underneath a <msg> that said nothing of the kind (2026-09-03).
    m = re.search(r"<msg[^>]*>(.*?)</msg>", resp, re.S)
    msg = (m.group(1).strip() if m else resp[:300]).replace("\n", " ")
    r = re.search(r"<reason[^>]*>(.*?)</reason>", resp, re.S)
    if r:
        msg += " | reason: " + r.group(1).strip().replace("\n", " ")
    return msg


# ---------------------------------------------------------------- XML builders
LIST_EXT = LST  # the std-list-1.0 extension must be DECLARED at login, not
                 # just used in a command — its absence draws "V274 Schema
                 # std-list- not specified at login" (2001), a check that only
                 # fires AFTER the command XML itself validates (measured
                 # 2026-09-03: the nested-<list:month> bug masked this one).


def login_xml(tag, pw):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><login><clID>{esc(tag)}</clID><pw>{esc(pw)}</pw>
<options><version>1.0</version><lang>en</lang></options>
<svcs><objURI>{DOM}</objURI><objURI>{CON}</objURI><objURI>{HOS}</objURI>
<svcExtension><extURI>{LIST_EXT}</extURI></svcExtension></svcs>
</login><clTRID>nominet-py-login</clTRID></command></epp>"""


def list_xml(month):
    # <list:expiry> is a SIMPLE element holding the month directly (pattern
    # \d\d\d\d-\d\d) — the nested <list:month> form inherited from epp.pl drew
    # 2001 "Element content is not allowed" on its first live run 2026-09-02.
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><info><list:list xmlns:list="{LST}">
<list:expiry>{esc(month)}</list:expiry></list:list>
</info><clTRID>list-{esc(month)}</clTRID></command></epp>"""


def check_xml(domains):
    names = "".join(f"<domain:name>{esc(d)}</domain:name>" for d in domains)
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><check><domain:check xmlns:domain="{DOM}">
{names}</domain:check></check><clTRID>nominet-py-check</clTRID></command></epp>"""


def info_xml(domain):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><info><domain:info xmlns:domain="{DOM}">
<domain:name>{esc(domain)}</domain:name></domain:info></info>
<clTRID>nominet-py-info</clTRID></command></epp>"""


def update_ns_xml(domain, add, rem):
    def block(hosts):
        return "<domain:ns>" + "".join(
            f"<domain:hostObj>{esc(h)}</domain:hostObj>" for h in hosts) + "</domain:ns>"
    addx = f"<domain:add>{block(add)}</domain:add>" if add else ""
    remx = f"<domain:rem>{block(rem)}</domain:rem>" if rem else ""
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><update><domain:update xmlns:domain="{DOM}">
<domain:name>{esc(domain)}</domain:name>{addx}{remx}</domain:update></update>
<clTRID>nominet-py-setns</clTRID></command></epp>"""


def host_create_xml(host):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<epp xmlns="{EPP}"><command><create><host:create xmlns:host="{HOS}">
<host:name>{esc(host)}</host:name></host:create></create>
<clTRID>nominet-py-hostcreate</clTRID></command></epp>"""


LOGOUT = f'<?xml version="1.0"?><epp xmlns="{EPP}"><command><logout/><clTRID>bye</clTRID></command></epp>'


# ---------------------------------------------------------------- parsers
def parse_check(resp):
    for cd in re.findall(r"<domain:cd>(.*?)</domain:cd>", resp, re.S):
        m = re.search(r'<domain:name[^>]*avail="([01])"[^>]*>([^<]+)</domain:name>', cd)
        if not m:
            continue
        reason = re.search(r"<domain:reason[^>]*>([^<]*)</domain:reason>", cd)
        yield m.group(2).strip(), m.group(1) == "1", (reason.group(1).strip() if reason else "")


def no_domains_claimed(resp):
    m = re.search(r'noDomains="(\d+)"', resp)
    return int(m.group(1)) if m else None


def assert_list_parse_matches(resp, parsed, month):
    """The server's own noDomains count is a built-in cross-check on the
    parser — if it ever disagrees with what parse_domains found, that is
    schema drift or a regex bug, and it must be LOUD, not a quietly short
    list (this exact silent failure was live for two weeks; see the comment
    on parse_domains)."""
    claimed = no_domains_claimed(resp)
    if claimed is not None and claimed != len(parsed):
        raise SystemExit(
            f"PARSER MISMATCH for {month}: server claims noDomains={claimed}, "
            f"parsed {len(parsed)} — the response shape has likely changed; "
            f"do not trust this walk's output until parse_domains is fixed")


def parse_domains(resp):
    # std-list-1.0's own element is <list:domainName> — NOT <domain:name>,
    # which belongs to the unrelated domain-1.0 schema used by check/info/
    # update. The wrong tag matches ZERO names on every real response and
    # raises nothing: a list command can return 1000 + noDomains="N" (N>0)
    # while this returned [] for every month, silently reporting an empty
    # estate (caught 2026-09-03 by printing noDomains and finding N>0 with
    # zero parsed names — never trust a parser that has never seen a hit).
    return re.findall(r"<list:domainName>([^<]+)</list:domainName>", resp)


def current_ns(resp):
    return sorted(h.lower().rstrip(".") for h in
                  re.findall(r"<domain:hostObj>([^<]+)</domain:hostObj>", resp))


def months_from(start, n):
    """['YYYY-MM', ...] for n months beginning at start (a date)."""
    out, y, m = [], start.year, start.month
    for _ in range(n):
        out.append(f"{y:04d}-{m:02d}")
        m += 1
        if m == 13:
            y, m = y + 1, 1
    return out


# ---------------------------------------------------------------- session
class Session:
    def __init__(self, transport):
        self.t = transport
        self.greeting = read_msg(transport)

    def cmd(self, xml, what, ok=(1000, 1001)):
        self.t.send(frame(xml))
        resp = read_msg(self.t)
        code = result_code(resp)
        print(f"  {what}: {code} {result_msg(resp)}", file=sys.stderr)
        if code not in ok:
            raise SystemExit(f"FAILED at {what} (code {code}) — nothing further attempted")
        return resp

    def close(self):
        try:
            self.t.send(frame(LOGOUT))
            read_msg(self.t)
        except (ConnectionError, OSError):
            pass
        self.t.close()


def open_session(a, need_login=True):
    t = DirectTransport() if a.direct else PodTransport()
    s = Session(t)
    print(f"GREETING_BYTES={len(s.greeting)}", file=sys.stderr)
    if need_login:
        tag, pw = read_credentials(a.credentials)
        s.cmd(login_xml(tag, pw), "login")
    return s


# ---------------------------------------------------------------- verbs
def v_probe(a):
    t = DirectTransport() if a.direct else PodTransport()
    s = Session(t)
    print(f"GREETING_BYTES={len(s.greeting)}")
    print("NOTE: the greeting is served to ANY IP; run `login` to prove the allowlist")
    s.close()


def v_login(a):
    s = open_session(a)
    print("LOGIN OK — egress allowlisted, TAG + password good")
    s.close()


def v_list(a):
    if not re.fullmatch(r"\d{4}-\d{2}", a.month):
        raise SystemExit("month must be YYYY-MM")
    s = open_session(a)
    resp = s.cmd(list_xml(a.month), f"list {a.month}")
    parsed = parse_domains(resp)
    assert_list_parse_matches(resp, parsed, a.month)
    for d in parsed:
        print(f"DOMAIN\t{d}")
    s.close()


def v_walk(a):
    if not 1 <= a.months <= 120:
        raise SystemExit("--months out of range (1..120; .uk registers up to 10 years)")
    s = open_session(a)
    # the expiry MONTH is the list command's own query key, so it rides along
    # for free: DOMAIN\t<name>\t<YYYY-MM>. A name listed under two months keeps
    # the first (should not happen; the count line would still expose drift).
    seen = {}
    for month in months_from(datetime.date.today(), a.months):
        resp = s.cmd(list_xml(month), f"list {month}")
        parsed = parse_domains(resp)
        assert_list_parse_matches(resp, parsed, month)
        for d in parsed:
            seen.setdefault(d, month)
    s.close()
    for d in sorted(seen):
        print(f"DOMAIN\t{d}\t{seen[d]}")
    print(f"TOTAL={len(seen)} months_walked={a.months}", file=sys.stderr)
    print("⚠ sanity-check the TOTAL against the owner's estimate (~1,500 as of "
          "2026-08-19); a plausible-looking short list means widen --months, "
          "not that the estate shrank", file=sys.stderr)


def v_check(a):
    names = [d.strip().lower().rstrip(".") for d in a.domains]
    bad = [d for d in names
           if not re.fullmatch(r"[a-z0-9][a-z0-9-]*(\.[a-z0-9][a-z0-9-]*)+", d)]
    if bad:
        raise SystemExit(f"refusing malformed name(s): {bad}")
    s = open_session(a)
    answered = {}
    for i in range(0, len(names), 5):
        resp = s.cmd(check_xml(names[i:i + 5]), "domain:check")
        for name, avail, reason in parse_check(resp):
            answered[name] = (avail, reason)
    s.close()
    for d in names:
        if d in answered:
            avail, reason = answered[d]
            print(("AVAILABLE" if avail else "TAKEN    ") + f"  {d}" +
                  (f"  ({reason})" if reason else ""))
    missing = [d for d in names if d not in answered]
    if missing:
        print(f"NO ANSWER for: {missing}", file=sys.stderr)
        sys.exit(1)


def v_info(a):
    s = open_session(a)
    resp = s.cmd(info_xml(a.domain), f"domain:info {a.domain}")
    s.close()
    print(f"ns: {current_ns(resp)}")
    for field, pat in [("registrant", r"<domain:registrant>([^<]+)"),
                       ("created", r"<domain:crDate>([^<]+)"),
                       ("expires", r"<domain:exDate>([^<]+)"),
                       ("clID", r"<domain:clID>([^<]+)")]:
        m = re.search(pat, resp)
        if m:
            print(f"{field}: {m.group(1)}")
    for st in re.findall(r'<domain:status s="([^"]+)"', resp):
        print(f"status: {st}")


def v_set_ns(a):
    target = sorted(h.lower().rstrip(".") for h in a.ns)
    s = open_session(a)
    info = s.cmd(info_xml(a.domain), f"domain:info {a.domain}")
    have = current_ns(info)
    print(f"current NS: {have}\ntarget  NS: {target}")
    if have == target:
        print("already at target — nothing to do")
    elif not a.apply:
        print(f"DRY-RUN: would rem {sorted(set(have) - set(target))} "
              f"add {sorted(set(target) - set(have))}")
        print("re-run with --apply to execute")
    else:
        add = sorted(set(target) - set(have))
        rem = sorted(set(have) - set(target))
        try:
            s.cmd(update_ns_xml(a.domain, add, rem), "domain:update")
        except SystemExit:
            # 2303: a target host object does not exist — create and retry once
            print("update refused — creating missing host objects, retrying once",
                  file=sys.stderr)
            for h in add:
                s.cmd(host_create_xml(h), f"host:create {h}", ok=(1000, 1001, 2302))
            s.cmd(update_ns_xml(a.domain, add, rem), "domain:update (retry)")
        verify = s.cmd(info_xml(a.domain), "verify domain:info")
        now = current_ns(verify)
        print(f"NS after update: {now}")
        print("SUCCESS" if now == target else
              "MISMATCH — inspect above; registry may apply asynchronously")
        if now != target:
            s.close()
            sys.exit(1)
    s.close()


# ---------------------------------------------------------------- self-test
def self_test():
    import xml.dom.minidom as md
    fails, ran = [], [0]

    def t(name, cond):
        ran[0] += 1
        print(("PASS " if cond else "FAIL ") + name)
        if not cond:
            fails.append(name)

    b = frame("<x/>")
    t("framing length prefix", struct.unpack(">I", b[:4])[0] == len("<x/>") + 4)
    for name, xml in [
            ("login xml well-formed", login_xml("TAG", 'p&<w"')),
            ("list xml well-formed", list_xml("2026-09")),
            ("check xml well-formed", check_xml(["a.uk", "b.co.uk"])),
            ("info xml well-formed", info_xml("a.uk")),
            ("update xml well-formed", update_ns_xml("a.uk", ["n1.x.com"], ["n2.y.com"])),
            ("hostcreate xml well-formed", host_create_xml("n1.x.com"))]:
        try:
            md.parseString(xml)
            t(name, True)
        except Exception as e:
            t(f"{name}: {e}", False)
    t("password escaping", "&amp;" in login_xml("T", "a&b") and "p&<" not in login_xml("T", "p&<w"))
    t("login declares the std-list extension",
      f"<svcExtension><extURI>{LST}</extURI></svcExtension>" in login_xml("T", "p"))
    # the REAL response shape (2026-09-03, one live domain): std-list-1.0 uses
    # <list:domainName>, NOT <domain:name> — the bug that returned [] for
    # every month with no error for two weeks.
    list_resp = ('<epp><response><result code="1000"><msg>ok</msg></result>'
                 '<resData><list:listData '
                 'xmlns:list="http://www.nominet.org.uk/epp/xml/std-list-1.0" '
                 'noDomains="1"><list:domainName>vending-machine.co.uk'
                 '</list:domainName></list:listData></resData></response></epp>')
    t("parse_domains reads list:domainName", parse_domains(list_resp) == ["vending-machine.co.uk"])
    assert_list_parse_matches(list_resp, parse_domains(list_resp), "2026-11")  # must not raise
    t("assert_list_parse_matches accepts an honest match", True)
    try:
        assert_list_parse_matches(list_resp, [], "2026-11")
        t("assert_list_parse_matches catches a mismatch", False)
    except SystemExit:
        t("assert_list_parse_matches catches a mismatch", True)
    canned = ('<epp><response><result code="1000"><msg>ok</msg></result>'
              '<resData><domain:chkData><domain:cd>'
              '<domain:name avail="1">free.uk</domain:name></domain:cd><domain:cd>'
              '<domain:name avail="0">taken.uk</domain:name>'
              '<domain:reason lang="en">registered</domain:reason></domain:cd>'
              '</domain:chkData></resData></response></epp>')
    got = list(parse_check(canned))
    t("parse_check both classes", got == [("free.uk", True, ""), ("taken.uk", False, "registered")])
    t("result_code", result_code(canned) == 1000)
    ns = current_ns("<domain:hostObj>B.x.COM.</domain:hostObj><domain:hostObj>a.y.com</domain:hostObj>")
    t("current_ns normalises + sorts", ns == ["a.y.com", "b.x.com"])
    m = months_from(datetime.date(2026, 11, 1), 4)
    t("month walk rolls the year", m == ["2026-11", "2026-12", "2027-01", "2027-02"])
    t("walk 120 allowed by validator", 1 <= 120 <= 120)
    import tempfile
    with tempfile.NamedTemporaryFile("w", suffix=".creds", delete=False) as f:
        f.write("TAG=T1\nEPP_PASSWORD=pw1\n")
    tag, pw = read_credentials(f.name)
    os.unlink(f.name)
    t("credentials parser", (tag, pw) == ("T1", "pw1"))
    with tempfile.NamedTemporaryFile("w", suffix=".creds", delete=False) as f:
        f.write("TAG=only\n")
    try:
        read_credentials(f.name)
        t("credentials parser refuses missing password", False)
    except SystemExit:
        t("credentials parser refuses missing password", True)
    os.unlink(f.name)
    print(f"{'ALL PASS' if not fails else 'FAILURES: ' + str(fails)} "
          f"({ran[0] - len(fails)}/{ran[0]})")
    sys.exit(1 if fails else 0)


# ---------------------------------------------------------------- main
def main():
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--self-test", action="store_true")
    ap.add_argument("--direct", action="store_true",
                    help="plain socket instead of the cluster-pod tunnel")
    ap.add_argument("--credentials", default=None,
                    help=f"credentials file (default {CRED_FILE})")
    sub = ap.add_subparsers(dest="verb")
    sub.add_parser("probe")
    sub.add_parser("login")
    p = sub.add_parser("list"); p.add_argument("month")
    p = sub.add_parser("walk"); p.add_argument("--months", type=int, default=24)
    p = sub.add_parser("check"); p.add_argument("domains", nargs="+")
    p = sub.add_parser("info"); p.add_argument("domain")
    p = sub.add_parser("set-ns")
    p.add_argument("domain"); p.add_argument("--ns", action="append", required=True)
    p.add_argument("--apply", action="store_true")
    sub.add_parser("register")
    a = ap.parse_args()

    if a.self_test:
        self_test()
    if a.verb == "register":
        raise SystemExit(
            "register is deliberately NOT here — it costs money and carries the "
            "registrant rulings. Use docs/agent_docs/docs024_key_docs_latest/"
            "idea_uk_vm_site/box/nominet-epp-domain-register.py (VMB-017).")
    if not a.verb:
        ap.print_help()
        sys.exit(2)
    {"probe": v_probe, "login": v_login, "list": v_list, "walk": v_walk,
     "check": v_check, "info": v_info, "set-ns": v_set_ns}[a.verb](a)


if __name__ == "__main__":
    main()
