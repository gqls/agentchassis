#!/usr/bin/env python3
"""Admit a client who has PROVEN they are a regulated firm.

THE POLICY THIS IMPLEMENTS (owner, 2026-08-19): *"if a client then says and proves
that they are regulated the system will let it through and do a good job for it —
that is a prime customer"*, and, for now, *"we reject a request until they write to
us by email with proof."*

So the flow is:

    site claims regulated status  ->  REFUSED by the fleet-wide guard (CGV-033)
    client emails proof           ->  a HUMAN checks it against the FS Register
    this script records it        ->  the site may state its status, and the FRN
                                      becomes a CITABLE FACT the claims layer checks

WHY A SCRIPT AND NOT AN AGENT. The decision "this firm really is authorised" is a
person reading an email and looking up the Financial Services Register. No agent
should be able to grant it, and nothing here calls a model. The script's whole job
is to make the human's decision durable, auditable, and correctly shaped.

WHAT IT WRITES. Two things, into the site's `evidence_base` aspect:

  1. `regulated` — the attestation itself: firm name, FRN, regulator, permissions,
     WHO checked it, WHEN, and WHAT they saw. All required. It is a record rather
     than a flag because "this site may call itself regulated" is a claim someone
     must stand behind later, and a boolean is unfalsifiable six months on.

  2. an evidence FACT of kind `attestation` carrying the FRN. This is the part that
     is easy to miss and is the reason the two are written together: the
     attestation PERMITS the claim, and the fact makes the claim CHECKABLE. Without
     it, an attested site could publish any FRN it liked — including a wrong one —
     and the claims layer would have nothing to compare against.

SAFETY PROPERTIES, stated because this grants authority:
  - validation mirrors RegulatedAttestation.Attested() exactly; anything it would
    reject is rejected here, so a recorded attestation always takes effect;
  - it REFUSES to overwrite an existing attestation without --replace, and prints
    the existing one first — silently replacing an audit record is the failure this
    guards against;
  - it supersedes rather than mutates (is_current=false + insert), so the previous
    state is recoverable;
  - --dry-run prints the exact JSON and writes nothing.

VERIFY AFTERWARDS with the same code the deploy gate runs:

    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user \\
      -d clients_db -At -c "SELECT data FROM site_specs ss JOIN sites s ON s.id=ss.site_id \\
      WHERE s.domain='<domain>' AND ss.aspect='evidence_base' AND ss.is_current" \\
      | go run ./cmd/regcheck -evidence - -claim "We are authorised and regulated by the FCA."

⚠ The guard itself is Go and is INERT UNTIL A CHASSIS ROLL. Recording an
attestation before then is safe and correct — it simply has nothing to exempt yet.
"""
import argparse
import json
import re
import subprocess
import sys
from datetime import date

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]

FRN_SHAPE = re.compile(r"^\d{6,7}$")


def psql(sql, tuples_only=True):
    cmd = PSQL + (["-At"] if tuples_only else [])
    r = subprocess.run(cmd + ["-c", sql], capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}")
    return r.stdout.strip()


def psql_stdin(sql):
    r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                       capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}\n{r.stdout.strip()}")
    return r.stdout.strip()


def sql_lit(s):
    return s.replace("'", "''")


def validate(a):
    """Mirror of RegulatedAttestation.Attested(). Anything Go would reject is
    rejected here, so a recorded attestation always takes effect — a script that
    accepted more than the guard would produce attestations that silently do
    nothing, which is worse than refusing."""
    problems = []
    if not a["firm_name"].strip():
        problems.append("firm_name is empty")
    if not FRN_SHAPE.match(a["frn"].strip()):
        problems.append(f"frn {a['frn']!r} is not 6 or 7 digits "
                        "(it is the Financial Services Register number, digits only)")
    if not a["attested_by"].strip():
        problems.append("attested_by is empty — who checked this?")
    if not a["evidence"].strip():
        problems.append("evidence is empty — what did you actually see?")
    try:
        date.fromisoformat(a["attested_at"].strip()[:10])
    except ValueError:
        problems.append(f"attested_at {a['attested_at']!r} is not YYYY-MM-DD")
    return problems


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--domain", required=True)
    ap.add_argument("--firm-name", required=True, help="the AUTHORISED entity's registered name")
    ap.add_argument("--frn", required=True, help="Financial Services Register firm reference number")
    ap.add_argument("--regulator", default="FCA")
    ap.add_argument("--permissions", default="", help="what the firm is actually authorised to do")
    ap.add_argument("--attested-by", required=True, help="the person who checked it")
    ap.add_argument("--attested-at", default=date.today().isoformat())
    ap.add_argument("--evidence", required=True,
                    help="what proof was seen, e.g. 'email 2026-08-19 from ops@firm + FS Register entry checked'")
    ap.add_argument("--replace", action="store_true", help="overwrite an existing attestation")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    att = {
        "firm_name": args.firm_name,
        "frn": args.frn,
        "regulator": args.regulator,
        "permissions": args.permissions,
        "attested_by": args.attested_by,
        "attested_at": args.attested_at,
        "evidence": args.evidence,
    }
    problems = validate(att)
    if problems:
        print("REFUSED — the attestation would not take effect:", file=sys.stderr)
        for p in problems:
            print(f"  - {p}", file=sys.stderr)
        sys.exit(2)

    site_id = psql(f"SELECT id FROM sites WHERE domain = '{sql_lit(args.domain)}'")
    if not site_id:
        sys.exit(f"no site row for domain {args.domain!r}")

    current = psql(
        f"SELECT COALESCE(data::text,'') FROM site_specs "
        f"WHERE site_id = '{site_id}' AND aspect = 'evidence_base' AND is_current")
    eb = json.loads(current) if current else {}
    if not isinstance(eb, dict):
        sys.exit("existing evidence_base is not a JSON object; refusing to touch it")

    existing = eb.get("regulated")
    if existing and not args.replace:
        print("REFUSED — this site already carries an attestation:", file=sys.stderr)
        print(json.dumps(existing, indent=2), file=sys.stderr)
        print("\nPass --replace only if you have re-checked the register yourself. "
              "Silently replacing an audit record is the thing this refusal exists to stop.",
              file=sys.stderr)
        sys.exit(3)

    eb["regulated"] = att

    # The FRN as a citable fact. The attestation PERMITS the claim; this makes it
    # CHECKABLE — without it an attested site could publish any number it liked.
    facts = [f for f in eb.get("facts", []) if f.get("id") != "regulated_frn"]
    facts.append({
        "id": "regulated_frn",
        "claim": f"{args.firm_name} is authorised and regulated by the {args.regulator}, "
                 f"firm reference number {args.frn}",
        "kind": "attestation",
        "source": {"type": "attestation",
                   "detail": f"attested by {args.attested_by}: {args.evidence}"},
        "verified_at": args.attested_at,
    })
    eb["facts"] = facts
    eb.setdefault("banned_claims", [])

    payload = json.dumps(eb, ensure_ascii=False)
    if "$evb$" in payload:
        sys.exit("payload contains the dollar-quote terminator; refusing to build unsafe SQL")

    print(f"site:   {args.domain}  ({site_id})")
    print(f"firm:   {args.firm_name}  FRN {args.frn}  [{args.regulator}]")
    print(f"by:     {args.attested_by} on {args.attested_at}")
    print(f"proof:  {args.evidence}")
    if args.dry_run:
        print("\n--dry-run: nothing written. The evidence_base that WOULD be written:\n")
        print(json.dumps(eb, indent=2, ensure_ascii=False))
        return

    sql = f"""
BEGIN;
UPDATE site_specs SET is_current = false, superseded_at = now()
 WHERE site_id = '{site_id}' AND aspect = 'evidence_base' AND is_current;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
VALUES ('{site_id}', 'evidence_base', $evb${payload}$evb$,
        'regulated-attestation', 'operator',
        'regulated attestation recorded by {sql_lit(args.attested_by)} on {sql_lit(args.attested_at)}; proof: {sql_lit(args.evidence)}',
        true, 'scripts/regulated/record_attestation.py');
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '{site_id}' AND aspect = 'evidence_base' AND is_current
     AND data->'regulated'->>'frn' = '{sql_lit(args.frn)}';
  IF n <> 1 THEN
    RAISE EXCEPTION 'attestation did not land: % current rows carry this FRN', n;
  END IF;
END $$;
COMMIT;
"""
    psql_stdin(sql)
    print("\nWRITTEN. Verify it at the guard (not just the row):")
    print(f"""  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \\
    -c "SELECT data FROM site_specs ss JOIN sites s ON s.id=ss.site_id \\
        WHERE s.domain='{args.domain}' AND ss.aspect='evidence_base' AND ss.is_current" \\
    | go run ./cmd/regcheck -evidence - -claim "We are authorised and regulated by the {args.regulator}." """)
    print("\n⚠ The guard is Go and is inert until the next chassis roll. This attestation is "
          "recorded correctly and will take effect then.")


if __name__ == "__main__":
    main()
