#!/usr/bin/env python3
"""install_fences.py — put the verified criteria into `doc_plans` as tool PLANs.

This is A4 of PLAN_2026-08-09. Until it runs, this site has NO tool PLAN for any
recreated tool, therefore no criteria, therefore no Tier 2 and no Tier 4, and
zero acceptance runs have ever happened here.

FOUR RULES, each measured rather than inherited. The neighbouring lane's
`loanandmortgagecalculator_couk/install_fences.py` is the ancestor of this file
and three of its four differ here — read the reasons before copying either.

1. THE SUBJECT KEY IS THE LADDER'S, NOT THE PAGE NAME. Both tiers derive it as
   `CASE WHEN cc.component_level='tool' THEN cc.function
         ELSE regexp_replace(p.name,'^tool-','') END`
   (discovery_checks/tool_eligibility.go). Our recreated pages carry a SECTION
   component, so `tool-stamp-duty` is keyed `stamp-duty` — not `tool-stamp-duty`,
   which is what this lane's handoffs assumed. A PLAN under the wrong key is a
   row nothing ever reads, and it fails silently and permanently: Tier 2 records
   `needs_criteria` and Tier 4 emits nothing, which is exactly what a site with
   no PLAN looks like. So the key is READ FROM THE ELIGIBILITY QUERY ITSELF
   here, never constructed from a page name in python.

2. ONLY ELIGIBLE TOOLS GET A ROW. The same query decides who the ladder can
   see. `tool-affordability` (two components, so it fails the sole-component
   clause), `game-fact-finder` (page_type 'game') and `investor-index`
   (page_type 'section-index') are NOT eligible — three of the "twelve recreated
   tools". Installing PLANs for them would create rows that look like coverage
   and are not.

3. ONLY RE-DERIVED ASSERTIONS ARE PINNED. Every expect_value must have been
   recomputed by verify_criteria.py from a published definition, from the
   evidence register, or from a stated tool convention. An emitted value that no
   model reproduces is exactly the thing PLAN §0 F3 warns about — "a golden
   captured from an already-wrong tool pins the wrong answer and then defends
   it" — so it is dropped, not pinned. This also handles containers and prose
   (`#resultsArea`, `#breakdown`) without a substring heuristic, and it is why
   `portfolio` installs nothing: toolgolden drove its term field to 1000/2000/
   500/450 years, the tool refused all four, and every emitted assertion is the
   validation message "Please enter a remaining term between 1 and 40 years."
   A fence built from that would certify the error message and call it a
   calculator.

4. computed_values ONLY, AND `no_auto_fix`. Not for the reason the neighbouring
   lane gives — ITS STATED GUARD IS INCOMPLETE, verified in the source and
   corrected here. Its comment says "with only computed_values in the fence,
   Tier 2 finds nothing it can fail, so it can never raise improve_tool". But
   check_tool_acceptance.go:478-500 appends THREE built-in shell failures
   (`shell-doc-header`, `shell-template-residue`, `shell-dead-controls`)
   *always*, independent of the fence, and any one of them creates an
   improve_tool item carrying `spec.component_id`. For these pages that id is
   the SHARED `hero` component: 252 pages across 18 sites — a wider blast radius
   than the ported-page shell that lane was protecting.
   So the guard that actually holds is `no_auto_fix: true` (Tier 4 routes a
   failure to acceptance_stuck/needs_human_review instead of tool-improver,
   tool_acceptance_actions.go), plus the measured fact that all twelve pages
   currently PASS all three built-in checks — run with the platform's own
   functions, positive control induced. That is a fact about today, not a
   guarantee: see NOTES 2026-08-10b and the LANDMINES entry.

Usage:
    python3 install_fences.py                 # dry run, prints what would land
    python3 install_fences.py --apply
    python3 install_fences.py --only stamp-duty [--apply]
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import verify_criteria as vc  # noqa: E402  — the models ARE the filter

CRIT = os.path.join(HERE, "criteria")
DOMAIN = "mortgagecalculator.co.uk"
CREATED_BY = "operator:mortgagecalculator-lane-a4"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]

APPLY = "--apply" in sys.argv
ONLY = sys.argv[sys.argv.index("--only") + 1] if "--only" in sys.argv else None

NO_AUTO_FIX_REASON = (
    "computed_values re-derived from published formulae and from this site's "
    "evidence register (13 SDLT facts, GOV.UK-cited, re-verified daily) — never "
    "from the page's own script. A failure means the ARITHMETIC or the LAW "
    "moved; the only way an automated rewriter could turn it green is by "
    "changing the numbers on a page quoting tax and consumer credit. A human "
    "decides what may change. See PLAN_2026-08-09_facts_into_tool_acceptance.md "
    "and bugs_open/225.")


def psql(sql, sep="\t"):
    r = subprocess.run(PSQL + ["-tA", "-F", sep, "-v", "ON_ERROR_STOP=1", "-c", sql],
                       capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit((r.stderr or r.stdout).strip()[:800])
    return r.stdout.strip()


# Rule 1 + 2: ask the ladder's own predicate who is eligible and what it calls
# them. Copied from tool_eligibility.go rather than paraphrased.
ELIGIBLE_SQL = """
SELECT CASE WHEN cc.component_level='tool' THEN cc.function
            ELSE regexp_replace(p.name,'^tool-','') END AS skey,
       p.name, COALESCE(p.url,''), COALESCE(p.build_status,'')
FROM pages p
JOIN sites s ON s.id = p.site_id
JOIN page_components pc ON pc.page_id = p.id
JOIN content_components cc ON cc.id = pc.component_id
WHERE s.domain = '%s'
  AND cc.is_active = true AND p.status = 'active'
  AND ( cc.component_level = 'tool'
     OR ( p.page_type = 'tool'
          AND NOT EXISTS (SELECT 1 FROM page_components pc_t
                          JOIN content_components cc_t ON cc_t.id = pc_t.component_id
                          WHERE pc_t.page_id = p.id
                            AND cc_t.component_level = 'tool' AND cc_t.is_active)
          AND (SELECT count(*) FROM page_components pc_s
               WHERE pc_s.page_id = p.id) = 1 ) );
""" % DOMAIN

eligible = {}
for line in psql(ELIGIBLE_SQL).splitlines():
    skey, name, url, build = line.split("\t")
    eligible[skey] = (name, url, build)

facts = vc.load_register_bands()
installed = skipped = 0

for fn in sorted(os.listdir(CRIT)):
    if not fn.endswith(".criteria.json"):
        continue
    slug = fn[: -len(".criteria.json")]
    if ONLY and slug != ONLY:
        continue
    if slug not in eligible:
        print("SKIP     %-18s not ladder-eligible on this site — a PLAN here "
              "would never be read" % slug)
        skipped += 1
        continue
    name, url, build = eligible[slug]
    model = vc.MODELS.get(slug)
    if not model:
        print("SKIP     %-18s no independent model — nothing may be pinned" % slug)
        skipped += 1
        continue

    doc = json.load(open(os.path.join(CRIT, fn)))
    checks, dropped = [], 0
    for c in doc["checks"]:
        want = model(vc.driven(c), facts)
        expect = {sel: text for sel, text in c["expect_values"].items() if sel in want}
        dropped += len(c["expect_values"]) - len(expect)
        if not expect:
            continue
        steps = list(c["steps"])
        # The runner opens ONE page per (url, profile) and runs every check
        # against it with no reload between vectors, while the capture used a
        # fresh page each time. A check whose clicks come BEFORE its fills is
        # building structure, so a later vector inherits the earlier one's rows.
        # None of these tools does that today; the rule stays so a future
        # row-adding tool cannot land here silently broken.
        first_click = next((i for i, s in enumerate(steps) if s["action"] == "click"), None)
        first_fill = next((i for i, s in enumerate(steps) if s["action"] in ("fill", "select")), None)
        if first_click is not None and (first_fill is None or first_click < first_fill):
            steps.insert(0, {"action": "reload"})
        checks.append({"id": c["id"], "type": "computed_values",
                       "profiles": ["desktop"], "steps": steps,
                       "expect_values": expect})

    if not checks:
        print("SKIP     %-18s every emitted assertion was unverifiable" % slug)
        skipped += 1
        continue

    n_assert = sum(len(c["expect_values"]) for c in checks)
    fence = json.dumps({"profiles": ["desktop", "mobile"], "no_auto_fix": True,
                        "no_auto_fix_reason": NO_AUTO_FIX_REASON,
                        "checks": checks}, indent=2, ensure_ascii=False)

    body = f"""# PLAN — {slug}

Live tool: https://{DOMAIN}{url}  (page `{name}`, subject key `{slug}`)

## Acceptance criteria

These values were **emitted from the live rebuilt page** by
`loancalculator_couk/toolgolden.py --emit-criteria`, and then **every one of
them was re-derived from a source that is not this page's script** by
`mortgagecalculator_couk_adoption/acceptance/verify_criteria.py`. An emitted
value is only "expected" because the tool currently prints it; pinning one that
nothing else can reproduce is how a defect gets defended instead of found
(`run_checks_action.go:775-781` says so in the code that does it, and
`bugs_open/225` is this estate shipping it for sixteen months).

Anything the verifier could NOT re-derive was **dropped rather than pinned** —
containers, prose breakdowns, echoed inputs, and one duration whose value turns
on a sub-penny rounding convention that nothing published settles.

Three strengths of evidence sit behind these numbers, and they are not equal:

- **DEFINITION** — the published formula (annuity, compound interest,
  amortisation run month by month), via the shared `oracles.py`.
- **REGISTER** — stamp-duty only, and it is the point of this lane: the bands
  are built from this site's **13 registered SDLT facts**, each a scalar
  carrying its own verbatim GOV.UK quote and re-verified daily by the
  `evidence-freshness` sweep. Not from a second hand-typed copy of the law.
- **CONVENTION** — a rule that is the tool's own design choice (the
  24/36-month phase split; what counts as "total cost"). Weaker: it catches a
  rewrite that changes the arithmetic, not a convention that was wrong to start
  with.

`computed_values` is desktop-only on purpose: arithmetic does not vary by
viewport, and the whole request has a 120s deadline that several vectors across
two profiles can exceed.

`no_auto_fix` is not optional here. A failing verdict must reach a human, not
`tool-improver` — the only way an automated rewriter can turn a red arithmetic
fence green is by changing the numbers.

```criteria
{fence}
```
"""

    print("%-8s %-18s -> key %-16s %d check(s), %2d assertion(s), %2d dropped   "
          "[build_status=%s]"
          % ("INSTALL" if APPLY else "would", slug, slug, len(checks), n_assert,
             dropped, build))

    if APPLY:
        sql = f"""BEGIN;
UPDATE doc_plans SET is_current=false, superseded_at=now(), updated_at=now()
 WHERE subject_type='tool' AND subject_key=$k${slug}$k$ AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, source_agent, created_by)
VALUES ('tool', $k${slug}$k$, $plan${body}$plan$,
        'mortgagecalculator_couk_adoption', 'toolgolden+verify_criteria', '{CREATED_BY}');
COMMIT;"""
        r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                           capture_output=True, text=True)
        if r.returncode != 0:
            sys.exit("INSTALL FAILED for %s:\n%s" % (slug, (r.stderr or r.stdout)[:800]))
        installed += 1

print("\n%d installed, %d skipped%s"
      % (installed, skipped, "" if APPLY else "   (dry run — pass --apply)"))

if APPLY:
    print("\nrows read back (fence position proves the fence survived quoting):")
    print(psql("SELECT subject_key, position('```criteria' in body) AS fence_pos, "
               "(body LIKE '%computed_values%') AS has_cv, "
               "(body LIKE '%no_auto_fix%') AS has_naf FROM doc_plans "
               "WHERE subject_type='tool' AND is_current AND created_by='"
               + CREATED_BY + "' ORDER BY subject_key;"))
