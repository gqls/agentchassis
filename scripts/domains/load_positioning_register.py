#!/usr/bin/env python3
"""Load REGISTER_positioning.md into the `positioning_register` table (migration 511).

OWNER RULING 2026-08-19 (RFC_037): the data goes in a database and the database is
the source of truth. This is the one-time-and-repeatable move.

THE THING THIS SCRIPT IS CAREFUL ABOUT. The register's value is its REASONING, and
that reasoning lives in prose no schema anticipates — 49 entries use 18 different
labelled field names, and `owns:` appears as a label in exactly ONE of them while
the same idea appears as prose in dozens. So:

    raw_md is ALWAYS stored, in full, and is authoritative.
    The typed columns are a convenience index over it.

A lossy parse of a hand-written document is not a migration, it is a deletion, and
"the database is the source of truth" is only safe if nothing is lost on the way
in. When a typed column disagrees with `raw_md`, `raw_md` wins.

IDEMPOTENT: upserts on `lower(domain)`, so re-running after editing the markdown
updates rows rather than duplicating them. That is also why the unique index in
511 is load-bearing rather than decorative.

WHAT IT REPORTS, and why the report is the point: every entry it could not fully
parse, and every field it left NULL. A loader that silently produced 49 thin rows
would look exactly like a loader that worked.

Usage:
    load_positioning_register.py                 # dry run: parse and report
    load_positioning_register.py --apply
    load_positioning_register.py --apply --exclude-file /tmp/the50.txt
"""
import argparse
import json
import os
import re
import subprocess
import sys

REGISTER = os.path.join("docs", "agent_docs", "docs024_key_docs_latest",
                        "portfolio_positioning", "REGISTER_positioning.md")
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]

DOMAIN_RE = re.compile(r"\b([a-z0-9][a-z0-9-]*\.(?:co\.uk|org\.uk|me\.uk|uk|com|net|org|io|ai))\b")


def psql_stdin(sql):
    r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                       capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}\n{r.stdout.strip()}")
    return r.stdout


def lit(s):
    """Single-quote a value for SQL, or NULL. Everything goes through here."""
    if s is None or (isinstance(s, str) and s.strip() == ""):
        return "NULL"
    return "'" + str(s).replace("'", "''") + "'"


def field(entry, name):
    """Pull a `**name:** value` field, stopping at the next `**field:**`, a `·`
    separator, or end of line. The register writes several fields per line
    separated by `·`, which is why that is a terminator."""
    m = re.search(r"\*\*" + name + r":\*\*\s*(.+?)(?=\s*·\s*\*\*|\n\s*[-*]\s*\*\*|\n\n|$)",
                  entry, re.S)
    if not m:
        return None
    v = re.sub(r"\s+", " ", m.group(1)).strip()
    v = re.sub(r"^[·\s]+|[·\s]+$", "", v)
    return v or None


def parse(md):
    """Split on `### ` headings; each becomes one entry with one or more domains."""
    family = None
    out = []
    # Walk the document so a heading inherits the most recent `## Family:` line.
    chunks = re.split(r"\n(?=(?:##|###) )", md)
    for ch in chunks:
        if ch.startswith("## Family:"):
            family = ch.split("\n", 1)[0][len("## Family:"):].strip()
            # A family block may contain its entries inline; fall through so the
            # `### ` chunks inside it are handled below.
        for e in re.split(r"\n### ", "\n" + ch)[1:]:
            title = e.split("\n", 1)[0].strip()
            code = None
            mcode = re.match(r"([A-Z]+\d+[a-z]?)\s*[—-]", title)
            if mcode:
                code = mcode.group(1)
            primary = field(e, "primary")
            twins = field(e, "twins")
            domains_f = field(e, "domains")

            prim_doms = DOMAIN_RE.findall((primary or "").lower())
            twin_doms = DOMAIN_RE.findall((twins or "").lower())
            other_doms = DOMAIN_RE.findall((domains_f or "").lower())

            neighbours = []
            nb = field(e, "neighbours")
            if nb:
                for d in DOMAIN_RE.findall(nb.lower()):
                    neighbours.append({"domain": d, "rule": nb[:400]})
                for c in re.findall(r"\b([A-Z]+\d+[a-z]?)\b", nb):
                    neighbours.append({"code": c, "rule": nb[:400]})

            row = {
                "entry_code": code,
                "family": family,
                "title": title,
                "proposition": title.split("—", 1)[1].strip() if "—" in title else None,
                "audience": field(e, "audience"),
                "stage": field(e, "stage"),
                "mode": field(e, "mode"),
                "stance": field(e, "stance"),
                "status": field(e, "status"),
                "neighbours": neighbours,
                "raw_md": "### " + e.strip(),
            }
            seen = set()
            for d in prim_doms:
                if d in seen:
                    continue
                seen.add(d)
                out.append(dict(row, domain=d, is_primary=True, primary_domain=None,
                                attribution="field"))
            base = prim_doms[0] if prim_doms else None
            for d in twin_doms + other_doms:
                if d in seen:
                    continue
                seen.add(d)
                out.append(dict(row, domain=d, is_primary=False, primary_domain=base,
                                attribution="field"))
            # PROSE SWEEP. The labelled fields alone left 82 of the 152 portfolio
            # domains with no row (measured 2026-08-20) — every one of them named
            # in an entry's body rather than in primary:/twins:/domains:. Now that
            # the database is the source of truth, a domain with no row is
            # invisible to anything that asks the table, so they are swept in and
            # tagged `prose` so the weaker attribution stays visible. See 512.
            for d in DOMAIN_RE.findall(e.lower()):
                if d in seen:
                    continue
                seen.add(d)
                out.append(dict(row, domain=d, is_primary=False, primary_domain=base,
                                attribution="prose"))
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--register", default=REGISTER)
    ap.add_argument("--apply", action="store_true")
    ap.add_argument("--exclude-file",
                    help="file of domains to mark exclude_from_build (the test set)")
    ap.add_argument("--exclude-reason", default="test domain (owner 2026-08-20): keep out of the build list")
    args = ap.parse_args()

    md = open(args.register, encoding="utf-8").read()
    rows = parse(md)

    excluded = set()
    if args.exclude_file:
        with open(args.exclude_file, encoding="utf-8") as f:
            excluded = {l.strip().lower() for l in f if l.strip() and not l.startswith("#")}

    # The report IS the deliverable of a dry run: a loader that silently produced
    # thin rows would look identical to one that worked.
    missing = {k: 0 for k in ("audience", "mode", "stance", "family", "entry_code", "proposition")}
    for r in rows:
        for k in missing:
            if not r.get(k):
                missing[k] += 1
    entries = len({r["entry_code"] or r["title"] for r in rows})
    print(f"entries parsed: {entries}")
    print(f"domain rows:    {len(rows)}  ({sum(1 for r in rows if r['is_primary'])} primary, "
          f"{sum(1 for r in rows if not r['is_primary'])} twin/other)")
    print(f"neighbour links: {sum(len(r['neighbours']) for r in rows)}")
    by_attr = {}
    for r in rows:
        by_attr[r.get("attribution")] = by_attr.get(r.get("attribution"), 0) + 1
    print("attribution:    " + "  ".join(f"{k}={v}" for k, v in sorted(by_attr.items(), key=lambda kv: -kv[1])))
    print("\nfields left NULL (raw_md still carries the whole entry, so nothing is lost):")
    for k, n in sorted(missing.items(), key=lambda kv: -kv[1]):
        print(f"  {k:<14} {n}/{len(rows)} rows")
    if excluded:
        hit = sum(1 for r in rows if r["domain"] in excluded)
        print(f"\nexclude_from_build: {len(excluded)} domains supplied, {hit} of them already in the register")

    if not args.apply:
        print("\n--dry-run (default): nothing written. Sample row:")
        if rows:
            s = dict(rows[0]); s["raw_md"] = s["raw_md"][:160] + " …"
            print(json.dumps(s, indent=2)[:900])
        # Domains in the exclusion list with NO register entry still need a row,
        # or the build dispatcher has nothing to check.
        if excluded:
            orphans = sorted(excluded - {r["domain"] for r in rows})
            print(f"\n{len(orphans)} excluded domains have no register entry — they will be "
                  f"inserted as exclusion-only rows (no proposition), e.g. {orphans[:5]}")
        return

    stmts = ["BEGIN;"]
    for r in rows:
        stmts.append(f"""
INSERT INTO positioning_register
 (domain, entry_code, family, is_primary, primary_domain, proposition, audience, stage, mode,
  stance, neighbours, raw_md, status, exclude_from_build, exclude_reason, attribution, source_file, parsed_at, created_by)
VALUES ({lit(r['domain'])}, {lit(r['entry_code'])}, {lit(r['family'])}, {str(r['is_primary']).lower()},
        {lit(r['primary_domain'])}, {lit(r['proposition'])}, {lit(r['audience'])}, {lit(r['stage'])},
        {lit(r['mode'])}, {lit(r['stance'])}, {lit(json.dumps(r['neighbours']))}::jsonb,
        {lit(r['raw_md'])}, {lit(r['status'])},
        {str(r['domain'] in excluded).lower()},
        {lit(args.exclude_reason) if r['domain'] in excluded else 'NULL'},
        {lit(r.get('attribution'))},
        {lit(args.register)}, now(), 'scripts/domains/load_positioning_register.py')
ON CONFLICT (lower(domain)) DO UPDATE SET
  entry_code=EXCLUDED.entry_code, family=EXCLUDED.family, is_primary=EXCLUDED.is_primary,
  primary_domain=EXCLUDED.primary_domain, proposition=EXCLUDED.proposition,
  audience=EXCLUDED.audience, stage=EXCLUDED.stage, mode=EXCLUDED.mode, stance=EXCLUDED.stance,
  neighbours=EXCLUDED.neighbours, raw_md=EXCLUDED.raw_md, status=EXCLUDED.status,
  exclude_from_build=EXCLUDED.exclude_from_build OR positioning_register.exclude_from_build,
  exclude_reason=COALESCE(EXCLUDED.exclude_reason, positioning_register.exclude_reason),
  attribution=EXCLUDED.attribution,
  source_file=EXCLUDED.source_file, parsed_at=now(), updated_at=now();""")
    for d in sorted(excluded - {r["domain"] for r in rows}):
        stmts.append(f"""
INSERT INTO positioning_register (domain, exclude_from_build, exclude_reason, attribution, source_file, parsed_at, created_by)
VALUES ({lit(d)}, true, {lit(args.exclude_reason)}, 'exclusion-only', {lit(args.register)}, now(),
        'scripts/domains/load_positioning_register.py')
ON CONFLICT (lower(domain)) DO UPDATE SET
  exclude_from_build=true, exclude_reason=EXCLUDED.exclude_reason,
  -- COALESCE, not overwrite: a domain that has BOTH a register entry and an
  -- exclusion keeps its real attribution. Without this the first run's rows kept
  -- a NULL attribution for ever, because the upsert never set it — which is how
  -- 51 rows ended up unlabelled after 512 added the column.
  attribution=COALESCE(positioning_register.attribution, EXCLUDED.attribution),
  updated_at=now();""")

    stmts.append("""
DO $$
DECLARE n_rows int; n_raw int; n_excl int;
BEGIN
  SELECT count(*) INTO n_rows FROM positioning_register;
  SELECT count(*) INTO n_raw  FROM positioning_register WHERE raw_md IS NOT NULL AND length(raw_md) > 40;
  SELECT count(*) INTO n_excl FROM positioning_register WHERE exclude_from_build;
  IF n_rows = 0 THEN RAISE EXCEPTION 'loader wrote nothing'; END IF;
  RAISE NOTICE 'positioning_register: % rows, % with raw_md, % excluded from build', n_rows, n_raw, n_excl;
END $$;
COMMIT;""")
    print(psql_stdin("\n".join(stmts)))
    print("WRITTEN. raw_md is authoritative — verify a row you know:")
    print("  SELECT entry_code, domain, audience, left(raw_md,120) FROM positioning_register "
          "WHERE domain='remortgagecalculator.uk';")


if __name__ == "__main__":
    main()
