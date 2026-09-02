# HANDOFF — bugs_open/427 + bugs_open/428, continue here

Written 2026-09-02, updated later same session after further work. Covers BOTH bugs —
one session worked both, they share boxingonline.com, and one fresh chassis build carries
all the code from both. **Read the two bug files' own status sections first — they are
the live record, this document is the orientation and the decision list:**

- `bugs_open/427_HANDOFF_2026-09-02_...md` — §9, §10, §11, §12 are today's account.
- `bugs_open/428_HANDOFF_2026-09-02_...md` — §9, §10, §11 are today's account.
- Full design/reasoning: `docs024_key_docs_latest/bugfix_427_event_render/` and
  `docs024_key_docs_latest/bugfix_428_planner_deferral/`.

## 0. One paragraph on each bug, current as of this update

**427** — boxingonline.com's fight-calendar page had a hero banner and prose describing a
calendar, but zero fixtures. Populator (`news_feed_ingestion`), correction path, and
render target (`query.upcoming_events`) are ALL built, tested, **fully council-approved**
(three REVISE rounds total, all fixed with real code — an evidence gate, a disclaimer, a
whole new discovery check for visibility — not argued past), and live in the current
chassis build. 6 real facts exist for boxingonline.com. The actual visible component
(`event-list`) is built, template-tested, and live IN THE LIBRARY — but deliberately
**not yet attached to boxingonline's live page**, because doing so goes through the full
page-build-handler pipeline on an already-deployed paid customer's page, and the
carry-forward safety of that pipeline for existing approved sections was not verified
this session. That's the one thing standing between this bug and closed.

**428** — the site-planner LLM deliberately skips strategy-named `entity-page`/
`entity-directory` roles ~76% of the time. The bug's first-proposed fix would have
reversed a recent owner safety ruling (RFC_056) — caught, escalated, and replaced with two
prompt fixes (both live, both approved) and a human-reviewed release surface (backend live
and approved; **frontend still not redeployed — 170 days old, this is the one open item**).

## 1. Decisions that need a person, not a session

1. **Attach `event-list` to boxingonline's live page (427's Phase B step 2)?** The
   component exists (migration `712`, library-only, zero live effect today). Attaching it
   means either (a) verifying first that `check_unresolved_sections` → `needs_rebuild` →
   `page-build-handler` carries forward a generic page's EXISTING sections untouched
   (read `plan_sections_action.go`'s handling of an already-deployed page, or test on a
   non-customer site first), or (b) accepting that risk knowingly on a paid customer's
   page. The exact SQL to run once decided is commented at the foot of migration `712`.
2. ~~**Deploy the admin-dashboard frontend**~~ **DONE — confirmed 2026-09-02 (`gap
   planner` session), no longer an open decision.** Owner said go-ahead; before running
   `make admin-dashboard`, checked current state first and found it already deployed —
   both pods 22 minutes old, `v1.0.1355`, no tag drift between the kustomize overlay and
   the running pods. Checked at the artefact rather than trusting rollout status: `kubectl
   exec` into both pods and grepped the served bundle directly
   (`grep -l 'Review Queue' /usr/share/nginx/html/assets/*.js`) — found in both, same
   content-hashed filename as this session's own pre-commit `vite build`, so this is
   confirmed to be the actual committed code, not a stale image under a reused tag. Not
   established which process triggered this deploy — most likely the same fresh-build
   event that put `ebf27c60377f984fd2847a1d5d88ff87ae01ebf7` on agent-chassis/core-manager
   (§2 below), rather than a separate admin-dashboard-only release. The "Deferred" filter,
   "Record verdicts only" checkbox, "Review & Release" button, and the "Review Queue" nav
   tab (added after this handoff's first draft — `7c359649f`,
   `frontends/admin-dashboard/src/App.tsx`) are all live now. Decision #3 below is
   therefore unblocked.
3. **Should anyone release any of 428's 13 record-mode verdicts** (boxingonline's own:
   `e3c2b440-c006-40ec-be7a-88d0b689ed1e`)? The tool exists (pending #2). Using it on a
   specific row is a content call, not a code one.
4. **Should `news_feed_ingestion`'s extraction prompt run beyond boxingonline.com?**
   Another lane's cost/quality call, named because it gates whether other sites in 428's
   sample could ever get real data.
5. **Arm `event_fixture_completeness`** (the new discovery check) in a live
   `run_checks.config.checks` array — low-risk config addition, not done this session,
   not urgent (it has nothing to check yet — zero incomplete/unevidenced facts today).

## 2. What was actually verified this session (don't re-derive)

- **Fresh chassis build, checked at the artefact**: `agent-chassis` and `core-manager`
  both report `git_commit = ebf27c60377f984fd2847a1d5d88ff87ae01ebf7` (via
  `service_binary_capabilities` and core-manager's own startup log respectively).
  `git merge-base --is-ancestor` confirmed every commit from both bugs through the point
  of that check is an ancestor. **Re-check this if picking up later** — more has been
  committed since (the event-list component migration, the completeness check); confirm
  those are in whatever's running before assuming so.
- **The admin-dashboard frontend is NOT part of any backend build** — separate pipeline,
  170 days stale, confirmed via `kubectl get deployment admin-dashboard`.
- **Council verdicts — ALL current submissions from this session, final state:**
  - `d0442d50…` (composeWriterBlock fix) — APPROVED.
  - `4849c95f…` (news_feed_ingestion candidate #1) — APPROVED.
  - `38be9226…` (428's release surface) — APPROVED.
  - `3f9cdfea…` (428's prompt migration 687) — APPROVED.
  - `08f56b7e…` (427's `query.upcoming_events` resolver) — **APPROVED**, after two REVISE
    rounds (round 1: compliance HIGH, evidence gate — fixed; round 2: bug_historian HIGH,
    visibility for incomplete/unevidenced facts — fixed with a new discovery check). Full
    objection-by-objection account in bug 427 §10-§12.
  - `ff91e666…` (the `event-list` component, migration 712) — submitted, **verdict not
    yet back as of this writing**. Check first:
    ```sql
    SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
    WHERE correlation_id='ff91e666-608d-4b26-9c41-d97d23a21437' AND kind='council_report'
    ORDER BY created_at;
    ```

## 3. What Phase B actually needs now (narrower than it looked this morning)

R1 (does an existing component fit) and R3 (which path minted the page, its
rebuild_policy) are ANSWERED: no existing component fits (surveyed all 32 `query.*`
components — none has an event/date/venue shape), a new one (`event-list`) is built and
council-submitted, and boxingonline's page is `rebuild_policy='generic'`,
`status='active'`, `build_status='deployed'`, sections
`["hero-tool","generic-text-block","advertising"]` (no function-specific tool slot was
ever declared — the widget was simply never planned in). What's left is purely the
attach-safely question in decision #1 above — not a design question anymore, a safety
verification one.

Once decided, closing the bug needs: run the sections-update SQL (foot of migration 712)
or let `check_unresolved_sections` find it on its own next sweep; verify
`experience_loop`'s nightly check reclassifies `/tools/fight-calendar/index.html` out of
"no control, no inline data, no runtime fetch" (the bug's own original measurement
instrument, and the actual closing signal — not a code diff).

## 4. Quick-reference commands

```bash
# Both pending/recent council verdicts
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT correlation_id, created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id IN ('08f56b7e-61e4-42d1-a3b6-13d700dd833c','ff91e666-608d-4b26-9c41-d97d23a21437')
  AND kind='council_report' ORDER BY created_at;"

# Confirm what commit a service is ACTUALLY running (re-check — more has shipped since)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT DISTINCT git_commit FROM service_binary_capabilities WHERE service='agent-chassis';"
kubectl -n ai-persona-system logs -l app=core-manager --tail=1000 | grep -m1 'build provenance'

# Frontend deploy state
kubectl -n ai-persona-system get deployment admin-dashboard

# The event-list component + boxingonline's page, current state
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT name, function, section_type FROM content_components WHERE name='event-list';
SELECT sections FROM pages WHERE site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND name='tool-fight-calendar';"

# The attach step, once decided (also at the foot of migration 712)
# UPDATE pages SET sections = (SELECT jsonb_agg(DISTINCT x) FROM jsonb_array_elements(
#   COALESCE(sections,'[]'::jsonb) || '["event-list"]'::jsonb) x)
# WHERE site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND name='tool-fight-calendar'
#   AND NOT (sections @> '["event-list"]'::jsonb);
```

## 5. Everything else is in the bug files

Both bug files carry the full diagnosis, every in-place correction (several — re-verify
before quoting a number from either), every fix candidate and why it was accepted/
refused, and every council correlation. The two `bugfix_*` directories carry RUNBOOKs
(editing a live agent prompt safely, mutation-testing a guard, the exact migration
dry-run pattern, checking a deploy artefact, validating a Go html_template standalone
before writing it into a migration) and plain-English `README_where_we_are.md` files.
