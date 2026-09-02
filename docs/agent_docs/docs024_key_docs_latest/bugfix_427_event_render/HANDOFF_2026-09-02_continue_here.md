# HANDOFF — bugs_open/427 + bugs_open/428, continue here

Written 2026-09-02, end of session, for a fresh chat to pick up cold. Covers BOTH bugs —
one session worked both, they share boxingonline.com as the motivating site, and the
fresh chassis build discussed below carries all the code from both. **Read the two bug
files' own status sections first — they are the live record, this document is the
orientation and the decision list, not a duplicate of them:**

- `bugs_open/427_HANDOFF_2026-09-02_no_writer_populates_dated_correctable_event_facts_so_boxingonlines_fight_calendar_shipped_empty.md`
  — §9, §10, §11 are today's account.
- `bugs_open/428_HANDOFF_2026-09-02_site_planner_llm_knowingly_defers_strategy_named_entity_roles_citing_its_own_final_say.md`
  — §9, §10, §11 are today's account.
- Full design/reasoning docs: `docs024_key_docs_latest/bugfix_427_event_render/` and
  `docs024_key_docs_latest/bugfix_428_planner_deferral/` (PLAN/NOTES/README/RUNBOOK each).

## 0. One paragraph on each bug

**427** — boxingonline.com's fight-calendar tool page shipped with a hero banner and
prose describing a calendar, but zero actual fixtures: nothing on the estate turned a
confirmed real-world event into a dated, evidenced fact, and nothing rendered one even
where it existed. Three sessions split the fix: `news_feed_ingestion` built the
populator (an extraction step that registers a dated fact from a news article), this
session built the correction path (mostly already covered by existing machinery) and the
render target (a new `query.upcoming_events` resolver). All of it is now live in a fresh
chassis build, and boxingonline.com has 6 real registered facts. **What's not done: no
component on the actual page declares the new source yet** — the resolver has nothing to
render onto.

**428** — the site-planner LLM correctly reads a strategy's recommended `entity-page`/
`entity-directory` roles and deliberately skips them ~76% of the time, citing its own
"final say" license with a generic justification. The bug's own first-proposed fix
(auto-dispatch the 13 already-detected verdicts) would have reversed a recent, deliberate
owner safety ruling (RFC_056, built after an earlier auto-dispatch of this exact finding
class destroyed live content) — caught and put to the user, who chose a safer path: two
prompt fixes (both live, both council-approved) and a human-reviewed release surface
(backend live and approved; **frontend not yet redeployed — this is the main open item**).

## 1. Decisions that need a person (not a session) to make

These are judgement calls this session deliberately did not make unilaterally, in order
of how much they block:

1. **Deploy the admin-dashboard frontend?** `make admin-dashboard` (Docker build, no
   local node/npm needed — pushes and applies the kustomize overlay). The backend release
   endpoint has been live and council-approved for a while; the button to use it from the
   UI is committed but sitting in a 170-day-old, unrebuilt frontend deployment. This is a
   production deploy action — confirm before running it, per this repo's own norm on
   actions that touch shared/live systems. Low risk (additive UI, no schema/data change)
   but it's still a deploy.
2. **Should anyone actually release any of bug 428's 13 record-mode verdicts?**
   (boxingonline's own row: `e3c2b440-c006-40ec-be7a-88d0b689ed1e`, would route to
   `page-build-handler`.) The tool exists (pending #1); *using* it on a specific row is a
   content/business call — is that gap worth building now, and does the person reviewing
   it agree with the LLM-audit seat's finding? Not a code question.
3. **Build Phase B of 427 now, or hand it to the next session?** Everything it depends on
   is live (resolver, 6 real facts). This is genuinely the next piece of work, not a
   pending council item — named as a "decision" only in the sense of "whose turn is it":
   this session ran out of the natural stopping point rather than out of blocking
   dependencies. See §3 below for exactly what it involves.
4. **Should `news_feed_ingestion`'s extraction prompt run beyond boxingonline.com?**
   Vertical-agnostic by design, tested against one site. A cost/quality call for that
   lane's owner — named here because it determines whether OTHER sites in bug 428's
   76%-omission sample could ever have real data to populate an eventual entity-directory
   page with. Not this session's call, and not urgent.

## 2. What was actually verified this session (don't re-derive these)

- **Fresh chassis build, checked at the artefact, not assumed from "a build was
  deployed":** `agent-chassis` (via `service_binary_capabilities`, every current pod
  uniform) and `core-manager` (via its own startup log — low volume, not scrolled) both
  report `git_commit = ebf27c60377f984fd2847a1d5d88ff87ae01ebf7`. `git merge-base
  --is-ancestor <commit> ebf27c60377f...` confirms every commit from both bugs through
  today is an ancestor — i.e. **it's really in there**, not just a same-tag relabel.
- **The admin-dashboard frontend is NOT part of that build** — it deploys separately
  (Docker + its own kustomize overlay) and its pods are 170 days old. Checked via
  `kubectl -n ai-persona-system get deployment admin-dashboard` (age column). This is the
  gap named in decision #1.
- **Council verdicts, read in full, not just the decision column:**
  - `d0442d50-e383-477f-9ed8-19eaaeea3d93` (composeWriterBlock fix) — APPROVED.
  - `4849c95f-2594-48e6-87b9-acee6341b0f8` (news_feed_ingestion candidate #1) — APPROVED.
  - `38be9226-d5b5-48b7-9b87-20efbaf3dec3` (428's release surface) — APPROVED.
  - `3f9cdfea-7287-4ab3-afad-9c386fbb7365` (428's prompt migration 687) — APPROVED.
  - `08f56b7e-61e4-42d1-a3b6-13d700dd833c` (427's `query.upcoming_events` resolver) —
    **REVISE** (compliance HIGH: no evidence gate before rendering real-world scheduling
    claims), fixed same session (commit `987ed3b3b`: requires citation url+quote,
    travelling disclaimer, settled naming, documented the `evidence-chart` prior-art
    check), **resubmitted on the same correlation — verdict not yet back as of this
    handoff.** Check it first thing in the new chat:
    ```sql
    SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
    WHERE correlation_id='08f56b7e-61e4-42d1-a3b6-13d700dd833c' AND kind='council_report'
    ORDER BY created_at;
    ```
    If it's REVISE again, read the objections in full before touching anything (see
    RUNBOOK for the exact query — `body` column, not `content`).

## 3. Phase B of 427, spelled out (the main remaining code task)

Nothing here is blocked; it just wasn't started. From `bugfix_427_event_render/PLAN`'s
§5 (R1-R3), still open:

1. Check whether an existing `content_components` row already accepts a `query.*`-sourced
   `items` array with event-shaped columns (candidates: anything in the `bugs_closed/381`
   sibling family), or whether a small new `event-list` component is needed. Query:
   `SELECT name, function, input_schema::text FROM content_components WHERE
   input_schema::text LIKE '%query.%';` and read what each one's template actually
   iterates over.
2. Find boxingonline's actual `/tools/fight-calendar/index.html` page row (`pages` table),
   its `rebuild_policy`, and which path minted it (`tool-deployer` vs `tool-generator` —
   they declare `pages.sections` differently, per `tool_content_item.go`'s header).
3. Write the migration placing the component (precedent: migration `267`, which inserted a
   component into an already-shipped tool page by hand). **The template MUST `{{if}}`-guard
   every optional field** (venue/broadcaster/participants) — this was council
   `bug_historian`'s flagged risk (Go's `missingkey=zero` renders an unguarded absent key
   as silent empty text with no error) and is a real, stated precondition, not decoration.
4. Submit that migration to council separately (DB migrations are their own council scope).
5. Verify: `experience_loop`'s nightly check should reclassify
   `/tools/fight-calendar/index.html` out of "no control, no inline data, no runtime
   fetch" — that's this bug's own original measurement instrument, and the actual closing
   signal, not a code diff or a green test.

## 4. Quick-reference commands

```bash
# Council verdict for the pending resubmission
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='08f56b7e-61e4-42d1-a3b6-13d700dd833c' AND kind='council_report'
ORDER BY created_at;"

# Confirm what commit a service is ACTUALLY running (not the roll, the artefact)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
SELECT DISTINCT git_commit FROM service_binary_capabilities WHERE service='agent-chassis';"
kubectl -n ai-persona-system logs -l app=core-manager --tail=1000 | grep -m1 'build provenance'

# Frontend deploy state
kubectl -n ai-persona-system get deployment admin-dashboard

# Boxingonline's site_id and its 6 real facts
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
SELECT id FROM sites WHERE domain='boxingonline.com';"
```

## 5. Everything else you need is in the bug files

Both `bugs_open/427...md` and `bugs_open/428...md` carry the full diagnosis, every
correction made along the way (several — re-verify before quoting an old figure from
either file), the fix candidates considered and why each was accepted/refused, and the
council submission correlations. The `bugfix_427_event_render/` and
`bugfix_428_planner_deferral/` directories carry the RUNBOOKs (exact commands for
editing a live agent prompt safely, mutation-testing a guard, checking a deploy
artefact) and the plain-English `README_where_we_are.md` for each. Start from the bug
files' own status sections (§9 onward in each); they're newer than this handoff for
anything that happens after it's written.
