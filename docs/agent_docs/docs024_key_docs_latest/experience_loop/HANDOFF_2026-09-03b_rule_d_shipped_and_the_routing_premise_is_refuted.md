# HANDOFF — experience_loop, continue here (2026-09-03b)

**Full path of this file:**
`docs/agent_docs/docs024_key_docs_latest/experience_loop/HANDOFF_2026-09-03b_rule_d_shipped_and_the_routing_premise_is_refuted.md`

**SUPERSEDES `HANDOFF_2026-09-03_experience_loop_continue_here.md`.** Read this alone. The
older file stays for the trail; **its §5c contains an instruction that the code refutes — see
§4 here before you act on any of it.**

---

## 1. Where this stands in one paragraph

The owner asked for `content-quality-auditor` to be put into the new build path. That turned
out to be three jobs. **(1) Give it sight — DONE 09-02 (migration 694, council-approved).
(2) Teach it the promise-keeping questions — DONE, and they fire.** Both verified again today
against chassis `v1.0.1358`. **(3) Route it into the build path — STILL OWED, and its stated
method does not work; §4 has the evidence and the corrected options.** Two side-quests that
became work of their own are now also DONE: **rule D** (the empty-index gap a peer lane found)
is built, ground-truthed and live, and both detectors' scoped-run control bugs are fixed. Both
outstanding peer promises are **discharged**.

---

## 2. What shipped today

### Rule D — a collection page with nothing in it (SQ-005)

| | |
|---|---|
| commits | `95f891a84` (detector + cronjob twin) · `6ad5e6c41` (register, 2 landmines, notes) · `1d96fd609` (peer reports + a third landmine) |
| live | ConfigMap `experience-promise-check-script-fh4ck725kb`, `cronjob configured`, triggered job **exitCode=0**, receipt in `doc_notes` 15:14:47Z |
| council | **NOT in scope** — `council-scope.sh` read, not assumed: `scripts/pattern-check.py` is the only in-scope script, and this is neither that, platform code, nor a migration |

**Fleet at ship, 2026-09-03:** 126 collection candidates — **71 list something, 28 bare
section-indexes skipped, 18 rule D findings across 9 sites, 8 never built, 1 render-vs-data
divergence.** Demand control 62 structured / 2 runtime / 7 own-links / **0 prose**.

**Ground truth, both directions, sampled away from the motivating site:** all **18** findings
confirmed EMPTY at the served body with a per-domain invented-URL control; **11 of 12** cleared
pages confirmed to serve items; the 12th (vonc's runtime-fill archive) checked at its source
instead — `https://vonc.com/data/provocations.json` serves **21 entries**.

**Three false positives the curl caught and the rule read cleanly past** — all three are now
fixtures, two are landmines: the inner join to `content_components` (hid two 20-odd-article
blog listings, because `rebuild_blog_listing` writes `component_id` NULL); matching the whole
URL path (made every page under `/guides/` a guides index); matching the whole title (imported
the site name — "Contact Us — Garden Tools UK").

### The other two

- **SQ-004 + SQ-005 scoped demand controls** now print `n/a` instead of `FAIL` when the
  control case or every clean page is outside `--site`. (SQ-004's half shipped this morning as
  `e535fc4f0`; I wrote the same bug into SQ-005 hours later and fixed it in `95f891a84`.)
- **`render_data_divergence`**, a bucket for the page that is neither empty nor clean.

---

## 3. Peer promises — BOTH DISCHARGED, do not re-promise them

| to | owed | state |
|---|---|---|
| `designblog.co.uk` | build the empty-index rule, re-run over their four pages, report | **DONE** — all four fire, re-confirmed at the served body; reported in their `NOTES_designblog_couk.md`, 2026-09-03 ~16:30Z. Both my caveats to them WITHDRAWN |
| `vetcomparison` | re-run both detectors post-701; fire the widened auditor (record mode) | **DONE** — 701's remainder verified at the ARTEFACT (it is NOT in `schema_migrations`, hand-applied); both detectors clean with non-vacuous controls; auditor RAN, correlation `bd349a6c-c7a6-432b-a2c1-f6f7a6b8d6fa`, COMPLETED in 51s, 8,235 input tokens, **4 findings, all record-mode**. Reported in their `NOTES_vetcomparison.md` |

---

## 4. ⚠ §5c's METHOD IS REFUTED BY THE CODE — read this before routing anything

The superseded handoff says, in bold: *"Use `call_agent`, not `spawn_agent` — the spawn→call
handshake fails ~half the time fleet-wide, and `apply_site_design` in that same workflow is
already a `call_agent` to copy verbatim."*

**A bare `call_agent` naming `content-quality-auditor` will fail on every build.** Read
`platform/orchestration/actions/call_agent.go`:

- `findTargetAgent` (~:225) has exactly three arms. With no `target_role` it tries
  `findAgentByType`, which **scans `CollectedData` for a SPAWN RESULT** of that type; then
  `isStandardAgent`, which (~:446) is the two-element list `{"search", "image"}`; then returns
  `fmt.Errorf("no spawned %s agent found")`.
- `apply_site_design` works **only because `spawn_webdesign_agent` ran earlier in the same
  workflow** and left a `webdesigner` role in collected data. It is not a copyable template
  for an agent that is never spawned.
- The two other candidate actions are not escapes: **`dispatch_agent`**
  (`dispatch_actions.go`) is a *remote-cluster alternative to spawn* whose own usage block
  shows `"next_step": "call_researcher"`, and **`StartOrchestrationAction`**
  (`spawn_actions.go` ~:1685) begins with `findSpawnData(params)`.

**So: every in-workflow path to another agent appears to require a prior `spawn_agent`.** That
is a structural claim, so per the 2026-07-31 owner ruling it is **filed for independent
diagnosis rather than asserted**: `RUN_CORRELATION_ID=ebb08a21-4460-427c-b2ca-dea66ef40c04`
(queue was empty; prior art checked — `bugs_open/210` is a *different* `call_agent` failure,
an `input_mapping` path miss, not this arm). **Read that verdict before you build.**

### The options, once the verdict lands

1. **`spawn_agent` + `call_agent`**, the `site-review-agent` pattern. Honest cost: the
   handshake's ~50% failure. Mitigation worth designing rather than assuming: the audit is
   advisory and record-mode, so a failed handshake should not fail the build — make the step
   non-fatal.
2. **A cold-dispatch action.** What I did by hand today proves the *transport* works: one
   `orchestrate` message to `system.agent.generic.requests` naming
   `content-quality-auditor` with `{site_id, domain}` ran the whole audit in 51 seconds. There
   is no workflow action that does this. Adding one is **platform code and architecture-scope**
   (a new shared mechanism on a shared seam — see CLAUDE.md's 2026-07-28 ruling), so it needs
   the concept register in the same commit and a council round.
3. **Leave it out of the workflow** and let the sweep cover it, which it already does — 13
   distinct sites audited today, roughly one every 15–20 minutes, all COMPLETED. This is the
   option to argue against explicitly rather than by omission: the owner asked for the build
   path specifically, and the sweep does not gate delivery.

**Unchanged and still correct from the old §5c:** insert between `apply_site_design` and
`update_site_status` (verified live today — the tail is still
`fix_items_loop → apply_site_design → update_site_status → complete`); the auditor needs **no
storage client**; it is **not redundant with `content-reviewer`** (that is per-PAGE inside
`build_items_loop`, this is a GROUP/site auditor, and boxingonline is the argument — every page
passed individually and the SITE was wrong); a migration under `sql_for_agents/` **is** council
scope.

---

## 5. ⚠ NEW LANDMINE — do NOT run `001_audit_agent_trigger.sh`

`scripts/initial_messages/010_audit_triggers/001_audit_agent_trigger.sh` has a clean usage
header, worked examples, and a `SAFETY:` paragraph promising *"Each agent operates on ONE site
only (the site_id you pass). It will NOT affect other sites."* **All of that is false.** The
file is **895 lines, mode 755, with ELEVEN `kcat -P` publish blocks and no `exit` between
them.** Lines 36–38 require the three arguments; every later block reassigns
`AGENT_TYPE`/`SITE_ID`/`DOMAIN` to hardcoded literals before publishing. Running it as
documented dispatches **11 runs at six OTHER lanes' sites and none at yours** — including
`blog-content-planner` and `image-build-handler`, which write rather than audit.

**Use this instead** (it worked today, correlation `bd349a6c`):
```bash
. scripts/kafka-publish-lib.sh
CORR=$(cat /proc/sys/kernel/random/uuid)
kafka_publish_checked --topic system.agent.generic.requests \
  --payload '{"action":"orchestrate","config":{"agent_type":"content-quality-auditor"},"input_data":{"site_id":"<uuid>","domain":"<domain>"}}' \
  --correlation "$CORR" --header correlation_id=$CORR --header orchestration_id=$(cat /proc/sys/kernel/random/uuid) \
  --header request_id=$(cat /proc/sys/kernel/random/uuid) --header message_id=$(cat /proc/sys/kernel/random/uuid) \
  --header message_type=request --header client_id=demo_client --header action=orchestrate \
  --header sender_agent_type=cli --header sender_agent_id=cli-user \
  --header responses_topic=system.agent.generic.responses --header timestamp=$(date -u +%FT%TZ)
# then VERIFY BY YOUR OWN CORRELATION, never a time window:
#   SELECT count(*) FROM orchestration_states WHERE correlation_id='<CORR>';
#   SELECT count(*) FROM llm_call_log        WHERE correlation_id='<CORR>';
```
Target the **group agent** directly; that is what avoids the spawn→call handshake by hand.

---

## 6. Per-roll verification (unchanged, still cheap, still do it)

Chassis was `v1.0.1358` today (pods up 12:06/12:07Z); the seat's `updated_at` moved to
12:06:08 with every 694 marker intact — that is now **five rolls survived**. The three queries
are in the superseded handoff §2 and still correct. **Use `llm_call_log`, never
`orchestration_states`** — that table is a **24-hour rolling window** (§4 of the old file) and
a zero there means "reaped", not "never ran".

Post-694 behaviour today: **~7,000 input tokens average** against a pre-694 average of
**1,744**, 0 failures.

---

## 7. What is left

1. **The routing (§4)** — read the diagnosis verdict `ebb08a21`, pick an option, argue it.
2. **CONTRIB ask 3, planner half** — refusing to SELECT a tool at planning time when we hold
   no data to put in it. The mechanical half shipped as SQ-005 rule B; this is the
   planner-side door-closer.
3. **Nothing else is owed to anyone.** Both peer promises are closed (§3).

---

## 8. The five standing docs

| doc | path |
|---|---|
| PLAN | `…/experience_loop/PLAN_experience_loop.md` |
| RUNBOOK | `…/experience_loop/RUNBOOK_experience_loop.md` |
| NOTES | `…/experience_loop/RUNNING_NOTES_experience_loop.md` — rule D entry at the foot |
| README (owner's) | `…/experience_loop/README_where_we_are.md` — **append only, never rewrite** |
| SUMMARY | `…/experience_loop/SUMMARY_where_experience_loop_is_now.md` — **deliberately not rewritten again.** The milestone is the routing going live; the five headings would still repeat the last one. Write a NEW dated file when §4 lands |

Register: **SQ-004** `listing-class-promise-check` (`25 7 * * *`) · **SQ-005**
`experience-promise-check` (`40 7 * * *`, now rules A–D).

**Refused deliberately, do not re-attempt as a regex:** *"does this page contain the thing its
title asserts?"* — disproof in SQ-005's docstring. Rule D is decidable because it counts items.
**Also refused:** widening rule D's collection nouns to catch designblog's `/criticism/` — it
would catch homegarden.uk's twelve month pages, which are articles. Stated gap, not an oversight.
