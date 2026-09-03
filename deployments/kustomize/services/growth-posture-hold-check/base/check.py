#!/usr/bin/env python3
"""
growth-posture-hold-check — the report migration 722 owed.

WHY THIS EXISTS. WDS-020 lets a site hold growth: `sites.settings->maintenance_profile
->>growth_posture = 'hold'` makes the tool chain file its growth as deferred, handler-less
records instead of dispatching it. Migration 722 (owner ruling 2026-09-02) makes every NEW
site born holding, released by a human. Both are silent by design — a held site errors
nowhere, its loop runs report clean, its growth items sit `deferred` — so a site that
nobody remembers to release simply stops growing and nothing anywhere says so. 722's own
header names that residual, and the council's guardian seat named it again. This job is
the answer: once a day, list every held site with how long it has been held, and raise a
finding for any LIVE site held longer than HOLD_DAYS.

WHY A CRONJOB AND NOT AN IMPROVEMENT-LOOP CHECK: a per-site discovery check would file its
finding as a handler-less `detected` row — which is precisely the class of finding the
improvement_loop lane opened on (1,385 of them nothing could act on and nobody could see).
The watchdog for a silent state must report somewhere a person reads, and must be driven by
something outside the mechanism it watches.

THE CLOCK. Migration 752 makes the born-held trigger stamp `growth_posture_set_at` (and
`_set_by`, `_reason`) beside the posture. A hand hold is asked to do the same (register
WDS-020 recipe). A hold with NO set_at — the two hand-holds that predate 752, or a lane that
skipped the recipe — is reported with the age UNKNOWN and bounded below by the first day
THIS check saw it, read back from its own earlier doc_notes rows (each held site gets one
line starting `HELD <domain> `, which is the token the lookup anchors on). That is honest,
needs no write to anybody's site, and self-corrects the day the lane stamps the row. Stated
caveat: a release-and-re-hold between two runs would make the lower bound too old; the
remedy is the stamp, not a smarter guess.

WHAT COUNTS AS LIVE. `status IN ('active','deployed') AND locked_at IS NULL`. A held site
that is still locked or pre-release (`test`, `pool`) is LISTED but is not a finding: it is
not growing for a different reason, and its lane will release the lock deliberately. The
hold becomes the silent one only once the site is otherwise live.

THE TRIGGER IS WATCHED TOO. This report's premise is that new sites are born held; if the
722 trigger is dropped or disabled, no new hold is ever created and a clean report would be
a lie. So the trigger's presence is a finding of its own kind (`trigger_missing`).

Writes ONE doc_notes row per run — findings or clean — so a MISSING row means the job did
not run, which must never look like "nothing is wrong". Exit 0 clean, 1 findings, 2 refused
to look (no password, or an empty fleet returned — a failed query must not read as clean).
Modelled on site-discovery-staleness-check (same image, same secret, same doc_notes
convention, same direct-Postgres constraint — no pods/exec RBAC in-cluster).

Hand runs use the SAME file, so there is no second copy to drift:
    scripts/audit-growth-posture-hold.sh              # kubectl route, no doc_notes write
    scripts/audit-growth-posture-hold.sh --days 14    # a different threshold for this run
    scripts/audit-growth-posture-hold.sh --write      # also record the row, as the cron does
    python3 check.py --self-test                      # fixtures only, no cluster
    python3 check.py --stdin < state.json             # classify a captured state document
"""

import json
import os
import re
import subprocess
import sys
from datetime import datetime, timedelta, timezone

# [ASSUMED 2026-09-03] a default pending an owner ruling. Every held site is LISTED on
# every run regardless of N — N only decides which of them are FINDINGS (exit 1).
HOLD_DAYS = 7

LIVE_STATUSES = ("active", "deployed")
SUBJECT_KEY = "growth-posture-hold"
SOURCE = "growth-posture-hold-check"          # == the CronJob and service-directory name
HELD_TOKEN = "HELD "                          # a line per held site starts with this
NOTE_TAG = "gphbody"                          # dollar-quote tag for the doc_notes INSERT

# One round trip, one JSON document. first_seen anchors on a NEWLINE + the token + the
# domain + a SPACE so `apis.uk` cannot match inside `xapis.uk` or `apis.uk.example`.
STATE_SQL = f"""
SELECT jsonb_build_object(
  'now', now(),
  'site_count', (SELECT count(*) FROM sites),
  'live_count', (SELECT count(*) FROM sites
                  WHERE status IN ('active','deployed') AND locked_at IS NULL),
  'trigger', (SELECT jsonb_build_object('present', count(*) > 0,
                                        'enabled', COALESCE(bool_or(tgenabled <> 'D'), false))
                FROM pg_trigger WHERE tgname = 'trg_sites_born_holding_growth'),
  'held', (SELECT COALESCE(jsonb_agg(jsonb_build_object(
              'domain',     s.domain,
              'status',     s.status,
              'locked',     s.locked_at IS NOT NULL,
              'created_at', s.created_at,
              'set_at',     s.settings->'maintenance_profile'->>'growth_posture_set_at',
              'set_by',     s.settings->'maintenance_profile'->>'growth_posture_set_by',
              'reason',     s.settings->'maintenance_profile'->>'growth_posture_reason',
              'first_seen', (SELECT min(n.created_at) FROM doc_notes n
                              WHERE n.subject_type = 'pipeline'
                                AND n.subject_key = '{SUBJECT_KEY}'
                                AND n.body LIKE '%' || E'\\n' || '{HELD_TOKEN}' || s.domain || ' %')
            ) ORDER BY s.created_at), '[]'::jsonb)
             FROM sites s
            WHERE s.settings->'maintenance_profile'->>'growth_posture' = 'hold')
)::text;
"""


# ---------------------------------------------------------------- transport

def run_sql(sql, local):
    """Run SQL and return trimmed stdout. Direct Postgres in-cluster; kubectl by hand."""
    if local:
        cmd = ["kubectl", "-n", os.environ.get("NAMESPACE", "ai-persona-system"), "exec", "-i",
               "postgres-clients-0", "--", "psql", "-U", "clients_user", "-d", "clients_db",
               "-tA", "-v", "ON_ERROR_STOP=1"]
        out = subprocess.run(cmd, input=sql, check=True, capture_output=True, text=True)
        return out.stdout.strip()
    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set (use --local for a kubectl run).",
              file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-tA", "-v", "ON_ERROR_STOP=1"],
        input=sql, env=env, check=True, capture_output=True, text=True)
    return out.stdout.strip()


# ---------------------------------------------------------------- pure half

def parse_ts(value):
    """Postgres/JSON timestamps ('2026-09-03T17:04:45.347469+00:00', '…Z', '… +00')."""
    if not value:
        return None
    v = str(value).strip().replace(" ", "T", 1)
    if v.endswith("Z"):
        v = v[:-1] + "+00:00"
    if re.search(r"[+-]\d\d$", v):
        v += ":00"
    try:
        dt = datetime.fromisoformat(v)
    except ValueError:
        m = re.match(r"(\d{4}-\d\d-\d\d)", v)
        if not m:
            return None
        dt = datetime.fromisoformat(m.group(1))
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt


def classify(state, hold_days=HOLD_DAYS):
    """Turn the state document into (rows, findings). No I/O — this is what --self-test drives."""
    now = parse_ts(state["now"])
    rows = []
    for h in state.get("held", []):
        live = h.get("status") in LIVE_STATUSES and not h.get("locked")
        since = parse_ts(h.get("set_at"))
        if since is not None:
            basis = "stamped"
        else:
            since = parse_ts(h.get("first_seen"))
            basis = ("first seen by this check — a LOWER BOUND, the row has no set_at"
                     if since is not None else
                     "unknown — no set_at on the row and this is the first run that sees it")
        age = (now - since).total_seconds() / 86400.0 if since is not None else None
        rows.append({
            **h, "live": live, "since": since, "basis": basis, "age_days": age,
            "overdue": bool(live and age is not None and age > hold_days),
        })

    findings = []
    trig = state.get("trigger") or {}
    if not trig.get("present"):
        findings.append({"kind": "trigger_missing",
                         "detail": "trg_sites_born_holding_growth is ABSENT from pg_trigger — no "
                                   "new site is born held; this is migration 722 undone."})
    elif not trig.get("enabled"):
        findings.append({"kind": "trigger_missing",
                         "detail": "trg_sites_born_holding_growth is present but DISABLED — no "
                                   "new site is born held until it is re-enabled."})
    for r in rows:
        if r["overdue"]:
            findings.append({"kind": "held_overdue", "detail": (
                f"{r['domain']} held for {r['age_days']:.1f} days (threshold {hold_days}d; "
                f"basis: {r['basis']}) — status {r['status']}, unlocked; "
                f"set_by: {r.get('set_by') or 'NOT RECORDED'}; "
                f"reason on the row: {r.get('reason') or 'NONE RECORDED'}")})
    return rows, findings


def should_refuse(state):
    """A failed query must never read as a clean fleet."""
    if not state.get("site_count"):
        return "0 sites returned — the query failed or the fleet is empty; refusing to report a clean fleet over it."
    return None


def fmt_age(r):
    if r["age_days"] is None:
        return "unknown"
    return f"{r['age_days']:.1f}d"


def render_report(state, rows, findings, hold_days=HOLD_DAYS):
    live_rows = [r for r in rows if r["live"]]
    pending_rows = [r for r in rows if not r["live"]]
    trig = state.get("trigger") or {}
    lines = [
        "GROWTH-POSTURE HOLD CHECK (WDS-020; the report migration 722 owed)",
        "",
        f"sites:                 {state.get('site_count')} total, {state.get('live_count')} live (active/deployed, unlocked)",
        f"born-held trigger:     {'present' if trig.get('present') else 'ABSENT'}, {'enabled' if trig.get('enabled') else 'DISABLED'}",
        f"held sites:            {len(rows)} ({len(live_rows)} live, {len(pending_rows)} not yet live)",
        f"threshold:             {hold_days} days [ASSUMED default — owner to rule]",
        f"findings:              {len(findings)}",
        "",
    ]
    for r in rows:
        # The HELD line is the token the next run's first_seen lookup anchors on. Keep
        # `HELD <domain> ` at the start of the line, verbatim.
        lines.append(
            f"{HELD_TOKEN}{r['domain']} age={fmt_age(r)} basis={r['basis'].split(' — ')[0]} "
            f"status={r['status']}{' LOCKED' if r.get('locked') else ''} live={'yes' if r['live'] else 'no'} "
            f"set_at={r.get('set_at') or '-'} set_by={r.get('set_by') or '-'} "
            f"reason={r.get('reason') or '-'}")
    if rows:
        lines.append("")
    if not findings:
        lines += [
            "No live site has been held longer than the threshold, and the born-held trigger",
            "is in place.",
            "",
            "This row exists on a clean run ON PURPOSE: a MISSING row means the job did not",
            "run, which is not the same as 'nothing is wrong', and the two must not look alike.",
        ]
    else:
        for f in findings:
            lines.append(f"  [{f['kind']}] {f['detail']}")
        lines += [
            "",
            "Remedies (register WDS-020 carries the recipes):",
            "  release: UPDATE sites SET settings = jsonb_set(settings, '{maintenance_profile,growth_posture}', '\"open\"') WHERE domain = '<domain>';",
            "           — a STATED 'open' is kept by the trigger and is greppable; then release the site's",
            "           held items from the growth_release_recipe on their spec.",
            "  keep holding, on the record: stamp growth_posture_reason / _set_by / _set_at on the row",
            "           so the next run reports an exact age and a reason instead of a guess.",
            "  trigger_missing: re-apply migration 722 (the trigger) and 752 (the record).",
        ]
    return "\n".join(lines)


# ---------------------------------------------------------------- doc_notes

def write_doc_note(body, local):
    if f"${NOTE_TAG}$" in body:
        raise RuntimeError("doc_notes body contains its own dollar-quote tag")
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', '{SUBJECT_KEY}', ${NOTE_TAG}${body}${NOTE_TAG}$, "
        f"'[\"{SUBJECT_KEY}\"]'::jsonb, '{SOURCE}');"
    )
    run_sql(sql, local)


# ---------------------------------------------------------------- self-test

def self_test():
    failures = []

    def ok(cond, label):
        if cond:
            print(f"  PASS  {label}")
        else:
            failures.append(f"    FAIL  {label}")

    now = datetime(2026, 9, 10, 8, 0, tzinfo=timezone.utc)
    iso = lambda dt: dt.isoformat()
    base = {"now": iso(now), "site_count": 60, "live_count": 40,
            "trigger": {"present": True, "enabled": True}, "held": []}

    def held(domain, days, **kw):
        row = {"domain": domain, "status": "deployed", "locked": False,
               "created_at": iso(now - timedelta(days=30)), "set_at": None, "set_by": None,
               "reason": None, "first_seen": None}
        if days is not None:
            row["set_at"] = iso(now - timedelta(days=days))
        row.update(kw)
        return row

    # 1. clean fleet
    rows, findings = classify({**base, "held": [held("a.uk", 2)]})
    ok(findings == [] and rows[0]["age_days"] is not None and abs(rows[0]["age_days"] - 2) < 0.01,
       "a live site held 2 days is listed and is not a finding")

    # 2. the threshold is load-bearing: 9 days is overdue at 7, not at 10
    _, f7 = classify({**base, "held": [held("a.uk", 9)]}, hold_days=7)
    _, f10 = classify({**base, "held": [held("a.uk", 9)]}, hold_days=10)
    ok([f["kind"] for f in f7] == ["held_overdue"] and f10 == [],
       "9 days held is a finding at N=7 and not at N=10 (threshold load-bearing)")

    # 3. a held site that is not yet live is listed, never a finding
    rows, findings = classify({**base, "held": [held("b.uk", 40, status="test", locked=True),
                                                 held("c.uk", 40, status="pool"),
                                                 held("d.uk", 40, status="deployed", locked=True)]})
    ok(findings == [] and all(not r["live"] for r in rows),
       "locked / test / pool sites held 40 days are listed as not-live and are not findings")

    # 4. no set_at: unknown on first sight; the check's own first_seen becomes a lower bound
    rows, findings = classify({**base, "held": [held("e.uk", None)]})
    ok(rows[0]["age_days"] is None and findings == [] and rows[0]["basis"].startswith("unknown"),
       "an unstamped hold with no history reads age unknown and is not a finding")
    rows, findings = classify({**base, "held": [held("e.uk", None, first_seen=iso(now - timedelta(days=9)))]})
    ok([f["kind"] for f in findings] == ["held_overdue"] and "LOWER BOUND" in rows[0]["basis"],
       "an unstamped hold first seen 9 days ago is overdue on the lower bound, and says so")
    rows, _ = classify({**base, "held": [held("e.uk", 1, first_seen=iso(now - timedelta(days=9)))]})
    ok(rows[0]["basis"] == "stamped" and abs(rows[0]["age_days"] - 1) < 0.01,
       "a stamped set_at wins over the check's own history")

    # 5. the trigger is watched
    _, findings = classify({**base, "trigger": {"present": False, "enabled": False}})
    ok([f["kind"] for f in findings] == ["trigger_missing"], "an absent trigger is a finding")
    _, findings = classify({**base, "trigger": {"present": True, "enabled": False}})
    ok([f["kind"] for f in findings] == ["trigger_missing"], "a disabled trigger is a finding")
    _, findings = classify({**base, "trigger": None})
    ok([f["kind"] for f in findings] == ["trigger_missing"], "a null trigger object is a finding, not a crash")

    # 6. the body: builds clean and dirty, carries the HELD token in the exact shape the
    #    first_seen lookup anchors on, and never contains its own dollar tag
    st = {**base, "held": [held("apis.uk", None, reason="owner: go ahead"), held("f.uk", 9)]}
    rows, findings = classify(st)
    body = render_report(st, rows, findings)
    ok(re.search(r"(^|\n)HELD apis\.uk ", body) is not None and re.search(r"(^|\n)HELD f\.uk ", body) is not None,
       "each held site gets a line starting 'HELD <domain> ' (the first_seen anchor)")
    ok(f"${NOTE_TAG}$" not in body, "body does not contain the dollar-quote tag")
    ok("[held_overdue] f.uk" in body and "NONE RECORDED" not in body.split("[held_overdue]")[0],
       "findings render with the domain named")
    clean_body = render_report(base, *classify(base))
    ok("ON PURPOSE" in clean_body and HELD_TOKEN not in clean_body,
       "a clean run's body explains why the row exists and lists nothing as held")

    # 7. refusal on an empty fleet
    ok(should_refuse({**base, "site_count": 0}) is not None and should_refuse(base) is None,
       "an empty fleet is refused; a populated one is not")

    # 8. timestamp shapes Postgres actually emits
    ok(parse_ts("2026-09-03T17:04:45.347469+00:00") is not None
       and parse_ts("2026-09-03 17:04:45+00") is not None
       and parse_ts("2026-09-03T17:04:45Z") is not None
       and parse_ts(None) is None,
       "parse_ts accepts to_jsonb(now()), psql text with a bare +00, and Z")

    # 9. the SQL and the renderer agree on the anchor: the LIKE pattern in STATE_SQL is
    #    newline + token + domain + space, and the renderer emits exactly that.
    ok("|| E'\\n' || '" + HELD_TOKEN + "' || s.domain || ' %'" in STATE_SQL,
       "STATE_SQL's first_seen anchor is newline + HELD_TOKEN + domain + space")

    if failures:
        print("\n".join(failures))
        print(f"\nself-test: {len(failures)} FAILED")
        return 1
    print("\nself-test: all cases pass")
    return 0


# ---------------------------------------------------------------- main

def main():
    args = sys.argv[1:]
    if "--self-test" in args:
        sys.exit(self_test())

    hold_days = HOLD_DAYS
    if "--days" in args:
        hold_days = int(args[args.index("--days") + 1])
    local = "--local" in args
    write = ("--write" in args) or (not local and "--dry-run" not in args)

    if "--stdin" in args:
        state = json.load(sys.stdin)
        write = False
    else:
        state = json.loads(run_sql(STATE_SQL, local))

    refusal = should_refuse(state)
    if refusal:
        print(f"REFUSING TO RUN: {refusal}", file=sys.stderr)
        sys.exit(2)

    rows, findings = classify(state, hold_days)
    report = render_report(state, rows, findings, hold_days)
    print(report)
    if write:
        write_doc_note(report, local)
        print(f"\ndoc_notes row written (subject_type='pipeline', subject_key='{SUBJECT_KEY}').")
    else:
        print("\n(dry run — no doc_notes row written)")
    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main()
