# PLAN — repairing webdesign.co.uk's 63 tools, one at a time, through the framework

**Started 2026-07-29** on the owner's instruction, verbatim:

> *"none of the tools actually work please work through them one at a time and fix
> them via the framework using the tools we have for this. search the docs for the
> tools tools and the doc traveller and start afresh for each tool documenting as
> you go."*
> *"let's fix the tools before deleting them."*
> *"we need to aim this site to what helps them - tools, helpful guides linking to
> the tools."* (ruling D15, `webdesign_couk/PLAN_2026-07-25` §D15)

This is now the site's **top priority**. Everything else on webdesign.co.uk waits.

---

## 1. What the framework is, and where it lives

The platform already has a self-verifying tools system — **do not build another
one.** Entry point: `travelling_docs/OVERVIEW_self_verifying_tools.md`; operating
manual: `travelling_docs/RUNBOOK_travelling_docs(39).md` (take the highest
numbered copy — the directory is full of numbered duplicates).

Two halves:

- **Travelling docs.** Every tool carries a versioned **PLAN** in Postgres
  (`doc_plans`, keyed `(subject_type='tool', subject_key=<function>)`) stating its
  aim, its delivery mechanism, its *"deliberate decisions — do not re-fix"*, and a
  machine-readable ```criteria fence; plus an append-only **NOTES** stream
  (`doc_notes`). Agents write these themselves: the generator at tool birth, each
  fixer after each repair.
- **The verification ladder.** Tier 0 generation integrity → Tier 1 structural
  (`check_tool_health`) → Tier 2 static contract-presence (`check_tool_acceptance.go`,
  under the anchor rule: static checks CONFIRM, never refute) → Tier 3 LLM audit
  (`tool-auditor`) → **Tier 4 real headless Chromium** (`browser-runner-adapter`,
  `internal/adapters/browserrunner/run_checks_action.go` — real click/fill/select,
  console capture, screenshots). A failure writes an `acceptance-fail` note and
  raises an `improve_tool` work item for `tool-improver`, which loads PLAN+NOTES,
  fixes, notes, re-renders, re-verifies.

Manual single-tool trigger (`travelling_docs/087_TRIGGER_tool_acceptance.sh`):

```bash
SPEC_FUNCTION=<tool-function> DOMAIN=webdesign.co.uk \
SITE_ID=6b49db8e-d447-4467-8277-4f3018af9897 SEND=1 \
./docs/agent_docs/docs024_key_docs_latest/travelling_docs/087_TRIGGER_tool_acceptance.sh
```

## 2. THE BLOCKER, measured before planning anything

**webdesign.co.uk's 63 tools are invisible to that entire ladder, and always have
been.** Two independent reasons, both checked against the live DB on 2026-07-29:

| gate | what it requires | what this site has |
|---|---|---|
| `check_tool_acceptance_due.go:51` | `cc.component_level = 'tool'` | **all 63 are `'section'`** (they are `ported-page` components) |
| PLAN criteria fence | a `doc_plans` row per tool | **zero** rows for any of the 63 |

```sql
-- both, verbatim, 2026-07-29
SELECT cc.component_level, count(*) FROM page_components pc
  JOIN content_components cc ON cc.id=pc.component_id
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE s.domain='webdesign.co.uk' AND p.page_type='tool' GROUP BY 1;
--  section | 63

SELECT count(*) FROM doc_plans WHERE subject_type='tool' AND subject_key IN
  (SELECT p.name FROM pages p JOIN sites s ON s.id=p.site_id
    WHERE s.domain='webdesign.co.uk' AND p.page_type='tool');
--  0
```

**This is why 63 tools could ship broken without one alarm.** They were PORTED
from website-design.com / websitedesign.com as static HTML blobs, never born
through the generator, so no PLAN was ever written and the acceptance sweep's
eligibility query has never returned one of them. Related, already filed:
`bugs_open/084` candidate 3 (widen Tier-4 past `component_level='tool'`).

**Consequence for this programme: "start afresh for each tool" is not a slogan,
it is the required bridging step.** A repair that does not also give the tool a
PLAN leaves it invisible again the moment we look away.

## 3. The per-tool loop (the unit of work)

One tool at a time. Each tool gets, in this order:

1. **Measure first, in a real browser.** `scratchpad/toolprobe.py` (this
   workstream's own witness; see RUNBOOK) loads the live page in headless
   chromium, records console errors, counts controls, drives the first control
   and reports whether anything changed. **A tool is "broken" only when the
   browser says so** — the static read disagreed with the browser on at least
   two tools (`community-growth`, `csp-builder` looked script-less and work).
2. **Write the PLAN** (`doc_plans`, `subject_type='tool'`): aim, delivery
   mechanism, deliberate decisions, and the ```criteria fence that Tier 2/4 will
   test it against. This is the "start afresh, documenting as you go" step and it
   is what makes the tool permanently visible to the ladder.
3. **Fix**, smallest change that satisfies the criteria.
4. **Verify in the browser again** — same probe, plus the specific behaviour the
   criteria name. Never verify by re-reading the source.
5. **Append a NOTES entry** (`doc_notes`) saying what was wrong and what changed.
6. **Ship**: file to `gqls/sites` + DB (`content_data` AND `rendered_html`, per
   the standing artefact rule), then confirm on the wire.

**Build up in steps, not big-bang** — the platform's own prior ruling, and the
owner's: `FOCUS_interactive_content_generation(4).md` §"Path C — work breakdown"
(*"six incremental steps, each independently verifiable. After each step we pause
and check the output before moving on"*) and `022_dynamic_applications.md` §5
*"Start with the simplest version that works… Each step is a separate work item,
not a big-bang rewrite."* Applied here: a tool that needs a rebuild gets its
simplest working version first, verified, then enrichment — not a rewrite
attempted in one pass.

## 4. Order of work

By the visitor's benefit, not by the file listing: **fix what is most broken and
most linked-to first.** The 31 guides link out to tools, so a broken tool behind a
guide's call-to-action is the worst case (a reader is sent to it deliberately).

Phase A — the browser census (§5) partitions the 63.
Phase B — pilot ONE tool end-to-end through §3, to prove the loop and time it.
Phase C — the rest, in benefit order, one at a time, each fully documented.

## 5. Census (Phase A)

Live-browser results are recorded in `NOTES_webdesign_tools_repair.md` as they
land, with the four verdicts the probe reports:

- **OK** — a control was driven and the page changed.
- **DEAD** — controls exist, driving one changes nothing (no JS, or unbound).
- **BROKEN** — the console threw (missing element, missing canvas, syntax).
- **NO-CONTROL** — nothing to drive; likely not interactive at all.

**The count in the owner's message ("none of the tools work") is the symptom
report, not the measurement** — the measurement is the census, and where it
disagrees with the report it is recorded honestly in NOTES either way.
