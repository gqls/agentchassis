# HANDOFF — 2026-08-12, fresh chat starts here: 243 stays closed, batch-8 tail is the whole job

**Supersedes `HANDOFF_2026-08-11b_continue_here.md`.** That handoff's §1–2 (243 all three
candidates done, round 3 approved) is now further confirmed, not just claimed — see §1
below. Its §3 (batch-8 tail: `tool-llm-cost-calculator`, `tool-bayesian-ranking`) is
**unchanged and is now this lane's entire remaining job.** Still binding from further back:
the 08-09 handoff's §0 (shared-228) and §2 (rerender traps), the 08-08 handoff's §3
(interactive-fence line), the 08-10 handoff's ADDENDUM (batch-8 requalification,
`computed_values` corrections, two-session coordination traps).

**This handoff exists because the previous session spent most of its length on read-only
monitoring, not authoring** (~18h of a 15-minute cron watching corr `310dee45` and the
batch-8 tail; see NOTES `## 2026-08-12 — ~18h of read-only monitoring…`). Nothing new was
built. This file exists so the next session doesn't have to read that story to start work.

## 1. State (verified 2026-08-12 ~12:40Z)

- **243 is closed and STAYED closed for ~18h under independent watch.** Round 3
  (`2dfa8900…`, corr `310dee45`) approved 2026-08-11 18:16Z, committed `786bc6759`. No round
  4, no regression, confirmed by re-querying `diagnosis_artifacts` (indexed by
  `correlation_id` — fast; the `orchestration_states` JSONB scan for the same question hit
  `statement timeout` twice under fleet load, so use the indexed table for a specific
  correlation). **Nothing to do here.** If you were about to re-litigate the vision-findings
  feature, don't — read NOTES `## 2026-08-11 (fresh session, resuming from HANDOFF)` first.
- **Fleet: chassis + browser-runner both confirmed live and caught up.** Build stamp
  `fa078ab3d` (chassis pods up since ~21:53Z 08-11), confirmed a descendant of `786bc6759`
  (round 3) and `585e37dad` (the wrapper) via `git merge-base --is-ancestor`. **Re-grep at
  session start regardless** — do not trust this timestamp once you've been in the tree a
  while. **Do not grep the binary for an ancestor commit's own hash** — get the build's own
  stamp first (provenance log line if fresh enough, else extract the binary's actual 40-hex
  stamp and cross-reference against `git log --all --format=%H`), THEN
  `git merge-base --is-ancestor`. Two sessions independently made the naive-grep mistake on
  2026-08-11 and both caught it before asserting anything wrong — see `LANDMINES.md`,
  entry `grep -aq <ancestor-commit-sha> /proc/1/exe reads absent…`.
- **The whole-fleet tag bump (`kustomization.yaml` × ~18 services) may still sit
  uncommitted in the tree** — the owner's release, not this lane's to commit. Pathspec your
  commits around it, as always. Check `git status` fresh; don't assume it's still there.
- **Batch-8 tail confirmed genuinely open, not secretly claimed.** Both remaining subject
  names surfaced repeatedly in *other* sessions' live transcripts over the ~18h watch window,
  but **zero commits landed on either** (`git log -i --grep` and path-based checks, both
  clean). Read that as: several sessions considered the work and didn't commit — not as "safe
  because nobody's looked," but also not as "someone else has it." **Still do the live-
  transcript check yourself before your first edit** (recipe in §5 below) — it can go stale
  in the time it takes to read this file.
- **`vision_finding_revalidator` is a separate spun-off thread, not this lane's job unless
  reclaimed.** `docs/agent_docs/docs024_key_docs_latest/vision_finding_revalidator/
  HANDOFF_2026-08-11_pre_plan.md`. Don't touch it from here; if you want it, go there and
  claim it on its own thread.

## 2. The two remaining jobs — both fully scoped already, nothing left to investigate

### 2a. `tool-bayesian-ranking` (gamesdesign) — needs a two-row rename FIRST

The page is currently named `bayesian-ranking`; the Tier-4 resolver wants
`tool-bayesian-ranking` (or `tool-tool-bayesian-ranking`). This is the same prefix-strip
pattern as the `tool-review-council-simulator` case from 07-31, and gamesdesign already
names **15 other tool pages** with the `tool-` prefix — so the rename restores that site's
own convention, it isn't a workaround.

**Do the rename via RUNBOOK §11, exactly** — two UPDATE statements in one transaction
(`pages.name` AND `site_plan_pages.name`), or the page silently drops out of
`check_sectionless_pages`'s population with no error. The RUNBOOK section has the full
before/after proof queries (collision check, the resolver's own query red-then-green, the
sectionless-join check) — run them, don't paraphrase them. Take the served-page byte-size
baseline **immediately before** the rename, not from an earlier session — the arena page
moved 1.1KB from an unrelated redeploy in the two hours it sat unmeasured, once.

Once the page resolves, author the fence normally (RUNBOOK §8 onward) and prove it.

### 2b. `tool-llm-cost-calculator` (ai-agent-orchestration) — MUST be authored fork-aware

`content_components` has no `site_id`; `function` is unique only among
`is_active AND forked_from IS NULL`, so **forks share the function and therefore share the
PLAN** (`doc_plans` keyed on `(subject_type, subject_key)`, not per-placement). This tool has
**four forks** besides the canonical row — on fundamentallyai, webdesign.co.uk, finetuning,
leopardess — templates differing by up to 3.3KB from canonical. Today only the canonical
site's page resolves (the other four are named `llm-cost-calculator`, missing the same
lookup the rename in 2a fixes for gamesdesign) — so a fence written against the canonical
template alone would silently red the moment any fork's page gets renamed to convention.
**Author the PLAN stating explicitly that it covers the fork set**, not just the canonical
placement — don't let "batch-8 clean single" framing (which this tool doesn't actually
qualify for) leak into the PLAN's own scope statement.

One shared property with the other batch-8 subjects, already measured: `length(js_content)
= 0` — tools inline their JS in `html_template`, no `/tools/assets/*.js` sidecar. The
batch-7 interactive rule (a fence must carry one check a static render cannot satisfy) still
applies, but the inert-script-mutant technique from sections doesn't transfer — mutate the
inline script or the element it writes into, not a `<script src>`.

## 3. What is explicitly NOT this lane's job right now

- **The eight loancalculator tools** (`tool-car-finance-pcp-hp` and 7 others) — genuinely
  different page-slug words from their component functions, not a prefix-strip case.
  Renaming would be an owner-visible change on another lane's site. Migration 384
  (`url_field` route) is live and waiting on their lane's own golden-derived PLANs — that's
  their blocker to clear, not this lane's.
- `tool-fuel-budget-forecaster` (gaswholesalers, logo-404-blocked) and `tool-gas-unit-
  converter` (known-broken tool) — both routed to their own lanes already.
- `vision_finding_revalidator` — see §1 above.

## 4. Standing defect list

Unchanged from `HANDOFF_2026-08-11b` §4. Item 9 (243) closed and stayed closed. Item 10
(batch-8 naming gate) dissolved by migration 384 for loancalculator; gamesdesign's own
rename (§2a above) is wanted on its own merits, not because of that migration.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file FROM DISK — it is co-edited, and the tree
   carries other sessions' commits between your reads.
2. Pod-grep chassis + browser-runner using the provenance-log-line + `git merge-base
   --is-ancestor` method (§1's landmine pointer). No dispatch within ~300s of a chassis pod
   (re)start.
3. Re-run the census + `CHECK_naming_contract.sh` before quoting any batch-8 figure — the
   table in NOTES (`## 2026-08-10…`, the "9 of 17 resolve" table) is a snapshot, not current
   truth by construction.
4. `who-owns.py` for `243`/`245`/anything loancalculator-adjacent, PLUS grep live
   transcripts, not just `git log`:
   ```
   CUT=$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ)
   find ~/.claude/projects/-home-ant-projects-agentchassis/ -name '*.jsonl' -newermt "$CUT" \
     | xargs grep -lE 'tool-bayesian-ranking|tool-llm-cost-calculator' 2>/dev/null
   ```
   A hit means CHECK the session's actual recent content (mtimes in this environment can
   run on a clock skewed by up to ~1h from `date -u`'s output — compare in-content ISO
   timestamps, not file mtimes, before concluding something is "live right now"). A sanctioned
   task is not a claimed task; claim in this file before your first edit.
5. Pick 2a or 2b above and follow it — both are fully scoped, nothing left to investigate
   before starting.
