#!/usr/bin/env python3
"""install_fences.py — put emitted `computed_values` criteria into `doc_plans`.

WHAT THIS IS FOR. `--emit-criteria` writes JSON files; nothing executes a file.
The platform's Tier-4 browser runner executes a ```criteria fence inside
`doc_plans.body`, looked up by (subject_type='tool', subject_key), and that row
is the only thing that makes these assertions part of the system rather than a
lane's private harness.

THREE THINGS IT DOES TO THE EMITTED JSON, each for a measured reason:

 1. `"profiles": ["desktop"]` on every computed_values check. The acceptance
    agent's request_run carries profiles ["desktop","mobile"], so an ungated
    check runs TWICE. These fences carry three vectors each and the whole
    request has a 120s deadline (TL-036 hit it at 36 evaluations in-cluster
    while the same fence took 10.6s locally). Arithmetic does not differ by
    viewport; layout checks are what mobile is for.

 2. DROPS CONTAINER SELECTORS. The emitter asserts every id that moved under
    driving, which includes wrappers like `#col-a` whose text is
    "Option A Amount (£) APR (%) ... Monthly Payment £160.04 ...". Pinning that
    makes every COPY EDIT a failed calculator, and the failure would name the
    calculator rather than the label that changed. A selector is treated as a
    container when another asserted selector's text is a substring of its own,
    so the leaf keeps the assertion and the wrapper loses it.

 3. Adds `page_status_ok`. A computed_values check on a page that 404s fails
    with a confusing selector error; this makes the real cause the first line.
    Nothing else is added: these fences exist to defend ARITHMETIC, and a red
    fence must mean the numbers moved. Console/overflow checks belong to the
    layout concern and would make a red result ambiguous.

Usage:
    python3 install_fences.py                 # dry run, prints the SQL
    python3 install_fences.py --apply         # supersede + insert, one tx each
    python3 install_fences.py --only standard-calc [--apply]
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
CRIT = os.path.join(HERE, "acceptance", "criteria")
SITE_ID = "ed633ada-f8af-424b-b4d4-8af79160dbcd"
DOMAIN = "loanandmortgagecalculator.co.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]

APPLY = "--apply" in sys.argv
ONLY = None
if "--only" in sys.argv:
    ONLY = sys.argv[sys.argv.index("--only") + 1]


def psql(sql, sep="\t"):
    r = subprocess.run(PSQL + ["-tA", "-F", sep, "-v", "ON_ERROR_STOP=1", "-c", sql],
                       capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit((r.stderr or r.stdout).strip()[:800])
    return r.stdout.strip()


def strip_containers(expect):
    """Drop any selector whose text CONTAINS another asserted selector's text."""
    kept = {}
    for sel, text in expect.items():
        others = [t for s, t in expect.items() if s != sel and t]
        if any(o in text and o != text for o in others):
            continue
        kept[sel] = text
    return kept


# pages.name is the subject_key the acceptance agent resolves by
# (tool_acceptance_actions.go: name IN ($key, 'tool-'||$key)); this site's pages
# are named e.g. 'loans-standard-calc', so the URL tail alone would not resolve.
pages = {}
for line in psql("SELECT url, name FROM pages WHERE site_id='%s' AND status='active';"
                 % SITE_ID).splitlines():
    url, name = line.split("\t")
    pages[url] = name


def subject_key_for(slug):
    hits = [(u, n) for u, n in pages.items()
            if u.rsplit("/", 1)[-1] == slug + ".html"]
    if len(hits) != 1:
        return None, "%d page(s) match %s.html" % (len(hits), slug)
    return hits[0][1], hits[0][0]


installed = skipped = 0
for fn in sorted(os.listdir(CRIT)):
    if not fn.endswith(".criteria.json"):
        continue
    slug = fn[: -len(".criteria.json")]
    if ONLY and slug != ONLY:
        continue
    key, url = subject_key_for(slug)
    if not key:
        print("SKIP  %-34s %s" % (slug, url))
        skipped += 1
        continue

    doc = json.load(open(os.path.join(CRIT, fn)))
    # NO page_status_ok — REMOVED 2026-08-09, and the reason is not cosmetic.
    #
    # Tier 2 (`check_tool_acceptance`) reads the SAME fence and, unlike Tier 4,
    # does NOT honour no_auto_fix. It skips every `computed_values` check
    # ("not statically checkable (Tier 4)", check_tool_acceptance.go:467) and
    # raises an improve_tool item only when something FAILED
    # (`if len(ev.failed) == 0 { continue }`, :222). `page_status_ok` is the one
    # check type in these fences that Tier 2 CAN evaluate — so it was the single
    # way a fence of ours could hand a page to tool-improver.
    #
    # That matters because Tier 2's improve_tool carries `component_id`, which
    # for these adopted pages is the shared `ported-page` shell used by ~154
    # pages across THREE sites — and tool-improver rewrote that shell once
    # already (webdesign.co.uk, 2026-08-04), after which it was flagged
    # component_template_corrupted and the repair fanned needs_rerender across
    # all three sites. Losing `{{.body}}` there empties 154 pages.
    #
    # With only computed_values in the fence, Tier 2 finds nothing it can fail,
    # so it can never raise improve_tool for these pages. That is a GUARD; the
    # thing holding it off otherwise was `build_status != 'deployed'`, which is
    # a data condition that any deploy would clear.
    # Nothing is lost: a page that does not serve fails its computed_values
    # checks in Tier 4 anyway (missing selectors), and verify_site.py covers
    # serving directly.
    checks = []
    dropped = reloaded = 0
    for c in doc["checks"]:
        expect = strip_containers(c["expect_values"])
        dropped += len(c["expect_values"]) - len(expect)
        if not expect:
            continue
        steps = list(c["steps"])
        # The runner opens ONE page per (url, profile) and runs every check
        # against it — there is no reload between vectors, while the capture
        # these values came from used a fresh page each time. A check whose
        # clicks come BEFORE its fills is building structure (consolidation adds
        # debt rows and removes one), so vector 2 inherits vector 1's leftovers
        # and its selectors no longer exist. Measured, not guessed: three of
        # consolidation's four vectors failed on `#d-name-2` while the same
        # steps drive the live page perfectly from a fresh load.
        # A click AFTER the fills is just the tool's own Calculate button and
        # needs no reload — those tools pass as they are, and a reload each
        # would spend the 120s whole-request budget for nothing.
        first_click = next((i for i, s in enumerate(steps) if s["action"] == "click"), None)
        first_fill = next((i for i, s in enumerate(steps) if s["action"] == "fill"), None)
        if first_click is not None and (first_fill is None or first_click < first_fill):
            steps.insert(0, {"action": "reload"})
            reloaded += 1
        checks.append({"id": c["id"], "type": "computed_values",
                       "profiles": ["desktop"], "steps": steps,
                       "expect_values": expect})
    n_assert = sum(len(c.get("expect_values", {})) for c in checks)
    # no_auto_fix is NOT optional on an arithmetic fence. A failing Tier-4 verdict
    # normally raises an `improve_tool` item and hands the tool to tool-improver,
    # which edits `content_components.html_template`. The only way an automated
    # rewriter can turn a failing COMPUTED-VALUES fence green is to change the
    # arithmetic — the exact thing the fence exists to protect, on pages quoting
    # consumer credit and tax. With this flag the failure escalates as
    # `acceptance_stuck` at `needs_human_review` and no improve_tool item is
    # created (platform/orchestration/actions/tool_acceptance_actions.go:850-930).
    fence = json.dumps({
        "profiles": ["desktop", "mobile"],
        "no_auto_fix": True,
        "no_auto_fix_reason": (
            "computed_values pinned to an independent oracle (oracles.py) on a "
            "consumer-credit and tax site; a failure means the ARITHMETIC moved, "
            "and the only way an automated rewriter could make it pass is by "
            "changing the numbers. A human decides what may change. See "
            "bugs_open/224 and bugs_open/225."),
        "checks": checks,
    }, indent=2, ensure_ascii=False)

    body = f"""# PLAN — {key}

Live tool: https://{DOMAIN}{url}

## Acceptance criteria

These values were **emitted from the live page** by
`loancalculator_couk/toolgolden.py --emit-criteria` on 2026-08-09, and then
**re-derived from the published formulae** by
`loanandmortgagecalculator_couk/verify_criteria.py`, which recomputes each one
from `oracles.py` — the independent module written from the annuity definition
and HMRC's published bands, never from this page's script. That second step is
the point: an emitted value is only "expected" because the tool currently
prints it, and pinning a wrong answer into the acceptance record is how a
defect gets defended rather than found (`bugs_open/224`, `bugs_open/225`, both
of which this estate shipped for months while a consistency-only golden
certified them green).

`computed_values` is desktop-only here on purpose: arithmetic does not vary by
viewport, and the whole request has a 120s deadline that three vectors × two
profiles can exceed.

```criteria
{fence}
```
"""
    print("%-8s %-30s -> %-34s %2d check(s), %2d assertion(s)%s"
          % ("INSTALL" if APPLY else "would", slug, key, len(checks), n_assert,
             (", %d container assertion(s) dropped" % dropped if dropped else "")
             + (", %d reload(s) added" % reloaded if reloaded else "")))

    if APPLY:
        sql = f"""BEGIN;
UPDATE doc_plans SET is_current=false, superseded_at=now(), updated_at=now()
 WHERE subject_type='tool' AND subject_key=$k${key}$k$ AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, source_agent, created_by)
VALUES ('tool', $k${key}$k$, $plan${body}$plan$,
        'loanandmortgagecalculator_couk', 'toolgolden+verify_criteria',
        'operator:bugfix224-session');
COMMIT;"""
        r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                           capture_output=True, text=True)
        if r.returncode != 0:
            sys.exit("INSTALL FAILED for %s:\n%s" % (key, (r.stderr or r.stdout)[:800]))
        installed += 1

print("\n%d installed, %d skipped%s"
      % (installed, skipped, "" if APPLY else "  (dry run — pass --apply)"))

if APPLY:
    print("\nverifying the rows read back:")
    print(psql("SELECT subject_key, position('```criteria' in body) AS fence_pos, "
               "(body LIKE '%computed_values%') AS has_cv FROM doc_plans "
               "WHERE subject_type='tool' AND is_current AND created_by="
               "'operator:bugfix224-session' ORDER BY subject_key;"))
