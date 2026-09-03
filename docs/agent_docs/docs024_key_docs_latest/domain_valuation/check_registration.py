#!/usr/bin/env python3
"""Ask the REGISTRY whether a domain is still registered, and when it expires.

Why RDAP and not DNS: a domain the owner registered but never delegated
answers NXDOMAIN — indistinguishable from a lapsed one by DNS alone (owner,
2026-08-19: "No nameserver usually means I never set a nameserver"). Only the
registry knows. 404 = not registered by anyone (expired/dropped or never held).

Usage: check_registration.py <domains.txt|csv> <out.csv> [--limit N]
Resumable: rows already in <out.csv> are skipped.
"""
import csv, json, os, subprocess, sys, time

ENDPOINTS = [
    ('.co.uk', 'https://rdap.nominet.uk/uk/domain/'),
    ('.org.uk', 'https://rdap.nominet.uk/uk/domain/'),
    ('.me.uk', 'https://rdap.nominet.uk/uk/domain/'),
    ('.uk', 'https://rdap.nominet.uk/uk/domain/'),
    ('.com', 'https://rdap.verisign.com/com/v1/domain/'),
    ('.net', 'https://rdap.verisign.com/net/v1/domain/'),
    ('.org', 'https://rdap.publicinterestregistry.org/rdap/domain/'),
    ('.us', 'https://rdap.nic.us/domain/'),
]


def endpoint(domain):
    for suffix, url in ENDPOINTS:
        if domain.endswith(suffix):
            return url
    return None


def query(domain, attempts=4):
    """Nominet throttles: a connection failure is NOT an answer, so retry with
    backoff and only ever report a state the registry actually gave us. This
    matters more than speed — an http-000 read as 'gone' would invent an
    expired domain, which is exactly the claim this file exists to test."""
    url = endpoint(domain)
    if not url:
        return 'unsupported-tld', '', '', ''
    for attempt in range(attempts):
        p = subprocess.run(['curl', '-sS', '--max-time', '30', '-w', '\n%{http_code}',
                            url + domain], capture_output=True, text=True)
        body, _, code = p.stdout.rpartition('\n')
        code = code.strip()
        if code in ('200', '404'):
            break
        time.sleep(2 * (attempt + 1))
    if code == '404':
        return 'NOT-REGISTERED', '', '', ''
    if code != '200':
        return f'http-{code or "error"}', '', '', ''
    try:
        d = json.loads(body)
    except json.JSONDecodeError:
        return 'unparsed', '', '', ''
    expiry = reg = ''
    for ev in d.get('events', []):
        if ev.get('eventAction') == 'expiration':
            expiry = (ev.get('eventDate') or '')[:10]
        if ev.get('eventAction') == 'registration':
            reg = (ev.get('eventDate') or '')[:10]
    registrar = ''
    for ent in d.get('entities', []):
        if 'registrar' in ent.get('roles', []):
            for item in ent.get('vcardArray', [[], []])[1]:
                if item[0] == 'fn':
                    registrar = item[3]
            registrar = registrar or ent.get('handle', '')
    status = ','.join(d.get('status', [])) or 'registered'
    return 'REGISTERED', expiry, registrar, status


def load(path):
    doms = []
    with open(path) as fh:
        first = fh.readline()
        fh.seek(0)
        if ',' in first:
            for r in csv.DictReader(fh):
                doms.append((r.get('domain') or '').strip().strip('"').lower())
        else:
            doms = [l.strip().lower() for l in fh]
    return [d for d in doms if d and '.' in d]


def main():
    src, out = sys.argv[1], sys.argv[2]
    limit = int(sys.argv[sys.argv.index('--limit') + 1]) if '--limit' in sys.argv else None
    doms = load(src)
    done = set()
    if os.path.exists(out):
        with open(out) as fh:
            kept = [r for r in csv.DictReader(fh) if not r['state'].startswith('http-')]
        # Rewrite without the failures so they are retried, never inherited.
        with open(out, 'w', newline='') as fh:
            w = csv.DictWriter(fh, fieldnames=['domain', 'state', 'expiry', 'registrar', 'status'])
            w.writeheader(); w.writerows(kept)
        done = {r['domain'] for r in kept}
    else:
        with open(out, 'w', newline='') as fh:
            csv.writer(fh).writerow(['domain', 'state', 'expiry', 'registrar', 'status'])

    # Control: a name nobody owns MUST come back NOT-REGISTERED, and a name we
    # know is held MUST come back REGISTERED. Without both, a run of 404s could
    # mean the endpoint is broken rather than the domains being gone.
    neg = query('thisdomaindoesnotexist9z8x7q.com')[0]
    pos = query('aakn.com')[0]
    print(f'control: unowned={neg} known-owned={pos}')
    if neg != 'NOT-REGISTERED' or pos != 'REGISTERED':
        sys.exit('CONTROL FAILED — refusing to run; results would be uninterpretable')

    n = 0
    with open(out, 'a', newline='') as fh:
        w = csv.writer(fh)
        for d in doms:
            if d in done:
                continue
            state, expiry, registrar, status = query(d)
            w.writerow([d, state, expiry, registrar, status]); fh.flush()
            n += 1
            if n % 50 == 0:
                print(f'  {n} checked')
            if limit and n >= limit:
                break
            time.sleep(1.0)
    print(f'done: {n} checked this run -> {out}')


if __name__ == '__main__':
    main()
