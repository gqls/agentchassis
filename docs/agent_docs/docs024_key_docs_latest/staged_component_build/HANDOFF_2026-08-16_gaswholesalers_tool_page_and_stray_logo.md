# HANDOFF — 2026-08-16: gaswholesalers' empty tool page + the stray `logo.jpg` (owner-directed, two tasks)

> **⚠ THIS DOES NOT SUPERSEDE `HANDOFF_2026-08-15c_continue_here.md`.**
> 15c is the lane's cold-start for the **RFC_029 Phase-1 revision**, and that task was
> **actively being built by another live session** when this file was written (2026-08-16
> ~11:00Z: `resolver_findings.go` untracked + `agent_error_log`/`SetResolverFindingRecorder`
> work dirty in the tree). **Do not touch `platform/orchestration/datahelpers/` or
> `platform/agentbase/agent.go` from this file's tasks** — you will collide on the
> same-file-passenger trap, which no hook can prevent.
> This file is a **separate, self-contained task pair** handed down by the owner. Read 15c
> for the lane; read this for these two jobs.

## 0. The owner's asks, and the one question that was answered

Owner, 2026-08-16: *"please remove the logo.jpg, I'd like the tool-gas-unit-converter tool to
exist so either rebuild it (with a guide) or fix this one. Will the full build-pipeline affect
other pages too?"*

**The blast-radius answer, which is the load-bearing part of the decision — ANSWERED, measured:**

| mechanism | scope | evidence |
|---|---|---|
| **082** (`082_submit_domain_unified.sh`) | **WHOLE SITE — all 36 active pages** | its own header flow: `build-site-planner → (cascade) needs_composition → needs_design → needs_content_page ×N → rerender`. It re-plans the site. **Too broad for one tool page.** |
| **`page-rebuild`** (agent type) | **ONLY pages you flag** | `get_pages_to_rebuild` filters `build_statuses=["needs_rebuild"]`. The site's own worked example is `scripts/initial_messages/001_assemble_all_pages_rerender/081_trigger_page_rebuild_gaswholesalers.sh`, which promotes exactly the pages it wants first. LLM-cost is **per page**, not per site. |

**So: yes, 082 would affect the other 35 pages. Do not use it for this.** The narrow route is
`page-rebuild` over a one-page selection — or, better if it works, a tool-component build that
does not re-run any writer at all (§2.2).

## 1. Verified state (2026-08-16 ~10:00–11:05Z — re-verify, this tree moves in hours)

- Fleet **`agent-chassis` + `browser-runner-adapter` v1.0.1303**, pods up 2026-08-15 18:45Z.
- **The guide page ALREADY EXISTS and is fine** — `tool-gas-unit-converter-guide`,
  `/guides/tool-gas-unit-converter-guide.html`, **3 sections**, `build_status=deployed`,
  deployed **2026-08-15**. So "rebuild it *with a guide*" is **half already done**; the
  remaining gap is the interactive tool page only.
- **The tool page is the empty shell** — `tool-gas-unit-converter`,
  `/tools/tool-gas-unit-converter.html`, page id **`7e576bc4-fb8b-46a4-b035-2842c481f35a`**,
  `status=active`, `build_status=deployed`, **`sections` = 0**.
- Site `gaswholesalers.com` = **`5fe15466-4e2e-4ff2-981e-98c1b7074002`**, **36 active pages,
  6 of them zero-section** (so this page is not unique — see §3 "do not widen").
- **No `doc_plans` row exists for this tool** (`subject_type='tool'`, subject_key matching
  `%gas%unit%` → **0 rows**). That matters: the acceptance-fence machinery keys on a tool PLAN,
  and installing one switches Tier-2 checking ON (standing landmine — do not install one as a
  side effect of this work).
- The three work items I unparked on 08-10 are still `needs_human_review`
  (`e4844153` needs_page, `261631b2` empty_section, `483fb749` required_fields_missing —
  9 llm-sourced fields never written). **`required_fields_missing` has no repair handler
  anywhere in the fleet** — it only ever closes by revalidation when something else writes
  the content. Re-verified today.
- **`/assets/images/logo.jpg` → 200** (the stray). `/assets/images/logo.png` → **200** (the
  real, correct logo). The stray is referenced by nothing.

## 2. TASK 1 — remove the stray `logo.jpg`. **There is no framework path today. This is the finding.**

Owner has authorised the removal. I traced the mechanism fully and stopped before improvising:

- `retract_page_deployment` is the ONLY action that reaches `delete_file`, and it derives every
  path from **`pages.url` via `datahelpers.PageFilePathFromURL`** (`:173`). It is **page-keyed
  and cannot name an asset path.** It also refuses any path an active page owns.
- The generic caller **refuses `delete_file` by allowlist, client-side**:
  `git_adapter_request_action.go:89` — *"adapter_action %q is not allowed (want one of: commit,
  create_branch, create_pull_request; delete_file is deliberately NOT reachable here — see
  RFC 011, use retract_page_deployment)"*. The exclusion is a **decision, not an oversight**
  (`:47-66`).
- **But the git-adapter itself DOES implement the verb** — `internal/adapters/git/adapter.go:363`
  `case "delete_file"`, parsing `paths` + `domain` (`:486–506`). And the deploy chain reconciles
  on the far side: gqls/sites' deploy-to-b2 runs `b2 sync --delete` + a Cloudflare purge on every
  push, so removing the file from the repo removes it from the edge (documented in
  `retract_page_deployment_action.go`'s own header).

**Three routes, pick deliberately — this is the decision the next session must make, not skip:**

1. **Direct Kafka dispatch to git-adapter with `delete_file`** (paths=`assets/images/logo.jpg`,
   domain=`gaswholesalers.com`). Reachable today, uses the adapter's own verb, same mechanism
   retraction uses. **It routes around RFC 011's deliberate client-side control** — defensible
   for one inert, owner-authorised file, but say so in the commit message and do not make it a
   habit. Worked dispatch example to copy: how `retract_page_deployment` builds its request.
   `[UNVERIFIED]` — I did not build or send the envelope; confirm the exact payload shape
   against `adapter.go:486–506` before firing.
2. **Extend the platform** with a narrow asset-retraction capability (the honest structural fix
   if this recurs). Platform-scope → council gate, concept register, the usual. Heavier than one
   file warrants unless bucket-style asset litter is expected again.
3. **Owner removes it by hand** from `gqls/sites` (one file, one commit).

**Recommendation: (1) with the reasoning recorded, or (3) if you'd rather not touch the control.**
Whichever: verify at the wire afterwards — `logo.jpg` must 404 **and** `logo.png` must still be
200. A check that only asserts the first would pass if you deleted the wrong one.

## 3. TASK 2 — make the tool exist. Narrow first; `page-rebuild` as the fallback

**Framework rule applies in full: the framework writes the content, not you** (owner ruling
2026-08-04). No hand-authored tool HTML, however small.

### 2.1 What is actually missing
The page has a component placement but no content: `483fb749`'s spec names the **9
schema-required, `source: llm` fields** that `content_data` never received —
`reference_table_heading`, `section_heading`, `section_subheading`, `table_note`, and five
`table_row_*_desc`. The template renders them as empty strings, which is the "empty-slotted
tool" symptom.

### 2.2 Preferred route — build the tool component, do not re-run a site writer
`create_tool_component` exists as an action (`registry.go:1465`). **Unverified and the first
thing to check:** whether it can populate this page's existing placement stand-alone, and
whether it needs a tool `doc_plan` (there is none — and creating one has the Tier-2 side effect
noted in §1). Read the action end to end before dispatching; two standing landmines sit on it
(`create_tool_component_action.go` is a **deliberately non-allow-listed** unrepaired
`page_components` writer, and `RepairPageLinks` cannot tell an anchor from JS that builds one).

### 2.3 Fallback — scoped `page-rebuild`, ONE page
If 2.2 cannot fill the fields, promote **only** page `7e576bc4` to `build_status='needs_rebuild'`
and fire `page-rebuild` for site `5fe15466…`. The 081 script is the worked envelope; **read its
pre-step note** — it documents exactly this promote-then-fire pattern, with a `RETURNING` clause
so you confirm **1 row** changed.

> **⚠ DO NOT WIDEN.** Five OTHER zero-section pages exist on this site. `page-rebuild` sweeps
> **every** page in `needs_rebuild`, so promoting more than one — or leaving an old promotion
> lying around — silently rebuilds pages nobody asked for, at LLM cost. **Check what is already
> in `needs_rebuild` before you promote** and, if anything unexpected is there, stop and ask.

### 2.4 Verify at the artefact, never at the status
`complete` is not proof. Afterwards: curl `/tools/tool-gas-unit-converter.html` and confirm the
9 fields render as real text (not empty strings); then re-check `483fb749` — it should close by
**revalidation**, which is the only way that item type ever closes. If the item is still open,
the content did not land whatever the run said.

## 4. Traps specific to these two tasks

- **`kubectl exec -i` with nothing piped to stdin HANGS** until timeout (from 15c §4, and it
  bit this session too). Drop `-i` unless you are piping a heredoc.
- **`pages` has no `page_name` column** — it is `name`; and `content_components` has no
  `template`/`html_content` — it is `html_template` / `js_content`. Both cost a round trip here.
- **The number 248 is ambiguous** — two unrelated bugs. The asset one is now in `bugs_closed/`;
  the CTA one is still open in `bugs_open/`. Resolve by filename, never by number.
- A **tool `doc_plan` switches Tier-2 acceptance checking ON** for that tool, which can then fail
  the page for three reasons unrelated to your fence. Do not create one as a side effect.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file **and** 15c from disk.
2. **Check whether the RFC_029 revision session is still live** before touching anything under
   `platform/` — `git status --short platform/orchestration/datahelpers/ platform/agentbase/`.
   Dirty means occupied; stay out.
3. Re-verify §1's state (the guide page, the tool page's 0 sections, both logo paths at the wire).
4. Confirm nothing unexpected already sits in `build_status='needs_rebuild'` on site
   `5fe15466…` before promoting anything.
5. `scripts/who-owns.py` on anything you are about to write to, and re-check the LLM cap
   (`llm_call_log.success`) before any content-writing dispatch — it was capped 08-14 and had
   recovered by 08-15; a monthly cap nominally runs to 2026-09-01.
