#!/usr/bin/env python3
"""
site-locale-unset-check — the watchdog half of bugs_closed/252 (og/lang slug).

WHY THIS EXISTS. 252 moved the document language out of Go and into the head
component: the template carries a gated `{{if .lang}}` attribute, the value comes
from site_specs `site_config.locale.lang`, and assemblePage reads it back to stamp
`<html lang>`. Migration 508 set it for every site that existed on 2026-08-20 —
naming each domain EXPLICITLY rather than deriving from the TLD, because
relojistas.com is Spanish and a blanket en-GB would have been false metadata
stated more confidently than the `en` it replaced.

That migration was a ONE-OFF. The next site created gets no locale at all, falls
back to the Go default `en`, and — before this job — nothing anywhere would say
so. The silent default is exactly the fault 252 was about, met one level up.

It caught a real case on its first outing: 508's own fail-closed guard ABORTED the
first apply because indoorplanters.co.uk had been created that same day. This job
is that guard, made recurring.

TWO FINDINGS, because the value has two ways of not reaching a page:

  A. UNSET — a real site with no `site_config.locale.lang`. It will declare `en`
     regardless of what it is. This is the case above.

  B. UNREACHABLE — a site that HAS a locale, whose head component's template has
     no `{{if .lang}}` gate to render it into. The value is set, correct, and
     cannot reach the page. This is not hypothetical: webdesign.co.uk's head was a
     hand-authored FRAGMENT with no `<head>` open tag at all (bugs_closed/347),
     so 508 set its en-GB and 117 pages kept serving `en` until 529 wrapped it.
     A new hand-authored head component reproduces that exactly.

B is the half worth having. A is visible the moment anyone looks at a new site;
B looks *finished* — the config is right there — and only the served page disagrees.

Deliberately NOT checked: whether a site's declared language matches its content.
"What language is this site?" is a product decision (owner ruling 2026-08-20:
non-English sites must not be en-GB, and that generalises to future language
sites). This job surfaces the question; it must never answer it.

`*.internal` pool domains are excluded — they serve no visitor.

Writes ONE doc_notes row per run — on findings AND on a clean result — so a
missing row means THE JOB DID NOT RUN, which must never read as "nothing is
wrong". Exits non-zero on findings. Modelled on site-discovery-staleness-check
(same image, same secret, same doc_notes convention, same direct-Postgres
constraint — no pods/exec RBAC here).
"""

import json
import os
import subprocess
import sys


def psql(sql, password, host):
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-tA", "-v", "ON_ERROR_STOP=1", "-c", sql],
        env=env, check=True, capture_output=True, text=True,
    )
    return out.stdout.strip()


# One round trip, one JSON document.
#
# The head-slot join is LEFT on purpose: a site with no head component row at all
# is neither finding — it has no head to carry anything and assembles from
# buildDefaultHead. Reported as context so the counts add up for a reader.
STATE_SQL = """
SELECT jsonb_build_object(
  'unset', (
    SELECT COALESCE(jsonb_agg(d ORDER BY d), '[]'::jsonb) FROM (
      SELECT s.domain AS d
      FROM sites s
      LEFT JOIN site_specs ss
             ON ss.site_id = s.id AND ss.aspect = 'site_config' AND ss.is_current
      WHERE s.domain IS NOT NULL AND s.domain <> ''
        AND s.domain NOT LIKE '%.internal'
        AND COALESCE(ss.data #>> '{locale,lang}', '') = ''
    ) u),
  'unreachable', (
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
             'domain', x.domain, 'lang', x.lang, 'component', x.cname) ORDER BY x.domain),
           '[]'::jsonb) FROM (
      SELECT s.domain, ss.data #>> '{locale,lang}' AS lang, cc.name AS cname
      FROM sites s
      JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'site_config' AND ss.is_current
      JOIN site_components sc ON sc.site_id = s.id AND sc.slot_name = 'head'
      JOIN content_components cc ON cc.id = sc.component_id
      WHERE s.domain IS NOT NULL AND s.domain <> ''
        AND s.domain NOT LIKE '%.internal'
        AND COALESCE(ss.data #>> '{locale,lang}', '') <> ''
        AND cc.html_template NOT LIKE '%{{if .lang}}%'
    ) x),
  'no_head_component', (
    SELECT COALESCE(jsonb_agg(d ORDER BY d), '[]'::jsonb) FROM (
      SELECT s.domain AS d FROM sites s
      WHERE s.domain IS NOT NULL AND s.domain <> '' AND s.domain NOT LIKE '%.internal'
        AND NOT EXISTS (SELECT 1 FROM site_components sc
                         WHERE sc.site_id = s.id AND sc.slot_name = 'head'
                           AND sc.component_id IS NOT NULL)
    ) n),
  'real_sites', (SELECT count(*) FROM sites
                  WHERE domain IS NOT NULL AND domain <> '' AND domain NOT LIKE '%.internal'),
  'langs_in_use', (
    SELECT COALESCE(jsonb_object_agg(lang, n), '{}'::jsonb) FROM (
      SELECT ss.data #>> '{locale,lang}' AS lang, count(*) AS n
      FROM sites s JOIN site_specs ss ON ss.site_id = s.id
                                     AND ss.aspect = 'site_config' AND ss.is_current
      WHERE s.domain NOT LIKE '%.internal' AND ss.data #>> '{locale,lang}' IS NOT NULL
      GROUP BY 1) l)
);
"""


def build_body(st):
    unset = st["unset"]
    unreachable = st["unreachable"]
    no_head = st["no_head_component"]
    lines = []

    if not unset and not unreachable:
        lines.append(
            f"CLEAN — all {st['real_sites']} real sites declare a document language, and every "
            "head component that serves one can render it."
        )
    else:
        lines.append(
            f"FINDINGS — {len(unset)} site(s) with no declared language, "
            f"{len(unreachable)} with a language their head cannot render."
        )

    if unset:
        lines.append("")
        lines.append(
            "A. UNSET — these sites will declare `en` on every page regardless of what they are. "
            "Set site_specs aspect 'site_config' key locale.lang, EXPLICITLY per site. Do NOT derive "
            "it from the TLD: .com sites on this estate are mostly British, and relojistas.com is "
            "Spanish (owner ruling 2026-08-20 — a non-English site must not be en-GB, and that "
            "generalises to future language sites). If the site has no content yet, say so when you "
            "set it rather than recording a guess as a measurement."
        )
        for d in unset:
            lines.append(f"  - {d}")

    if unreachable:
        lines.append("")
        lines.append(
            "B. UNREACHABLE — these sites HAVE a language and their head component has no "
            "`{{if .lang}}` gate to render it into, so the value is set, correct, and cannot reach "
            "the page. This is bugs_closed/347's shape (a hand-authored head component). Fix the "
            "TEMPLATE — add the gated attribute to its <head> open tag, and a map-valued `lang` "
            "input_schema entry with source config.locale.lang; migration 529 is the worked example. "
            "A scalar schema entry is silently skipped by the resolver."
        )
        for r in unreachable:
            lines.append(f"  - {r['domain']} (lang={r['lang']}, component={r['component']})")

    if no_head:
        lines.append("")
        lines.append(
            "CONTEXT, not a finding — sites with no head component row at all. They assemble from "
            "buildDefaultHead, which carries no language and is not something this check can fix:"
        )
        for d in no_head:
            lines.append(f"  - {d}")

    lines.append("")
    lines.append(f"Languages in use: {json.dumps(st['langs_in_use'], sort_keys=True)}")
    lines.append(
        "Verify at the ARTEFACT, never at the config: "
        "curl -s https://<domain>/<inner-page> | grep -oE '<html[^>]*>'. "
        "A page shows the change only after it re-assembles (bugs_open/346)."
    )
    return "\n".join(lines)


def write_doc_note(body, password, host):
    # source is the CHECK's name, identical to the CronJob and service-directory
    # name on purpose, so the doc_notes.source landmine (script name vs CronJob
    # name diverging) cannot fire. Query by categories ? 'site-locale-unset'.
    tag = "slucbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'site-locale-unset', ${tag}${body}${tag}$, "
        "'[\"site-locale-unset\"]'::jsonb, 'site-locale-unset-check');"
    )
    path = "/tmp/site-locale-unset-note.sql"
    with open(path, "w") as f:
        f.write(sql)
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-v", "ON_ERROR_STOP=1", "-f", path],
        env=env, check=True,
    )


def main():
    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("CLIENTS_DB_PASSWORD not set", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    try:
        st = json.loads(psql(STATE_SQL, password, host))
    except subprocess.CalledProcessError as e:
        print(f"query failed: {e.stderr}", file=sys.stderr)
        sys.exit(2)

    body = build_body(st)
    print(body)
    write_doc_note(body, password, host)
    print("\ndoc_notes row written (subject_type='pipeline', subject_key='site-locale-unset').")

    findings = bool(st["unset"]) or bool(st["unreachable"])
    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main()
