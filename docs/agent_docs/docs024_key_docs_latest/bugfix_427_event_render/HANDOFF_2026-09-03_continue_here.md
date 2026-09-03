# HANDOFF — bugs_open/427 + bugs_open/428, continue here (supersedes HANDOFF_2026-09-02)

Written 2026-09-03. Continues directly from `HANDOFF_2026-09-02_continue_here.md` in this
same directory (read that first if you want the full arc from filing to yesterday's
council rounds — this document only covers what changed since). One session worked both
bugs again today; they still share boxingonline.com.

**Read the two bug files' own status sections first — they are the live record:**
- `bugs_open/427_HANDOFF_2026-09-02_...md` — §13 is today's account.
- `bugs_open/428_HANDOFF_2026-09-02_...md` — §12 is today's account.
- Full narrative + evidence: `docs024_key_docs_latest/bugfix_427_event_render/NOTES_*.md`
  (2026-09-02/03 entry, long, evidence-heavy) and `README_where_we_are.md` (plain prose).

## 0. One paragraph on each bug, current as of this update

**427** — `event-list` is now attached to boxingonline.com's live fight-calendar page and
deployed (git commit `007b3a7a1`, confirmed via the actual GitHub Actions "Sync to B2" log,
not just a DB status). Attached via `apply_section_edit`'s `component_swap` — a narrower,
already-safety-gated mechanism than the full-rebuild path yesterday's handoff was worried
about — after the council's `prior_art_librarian` flagged that path existed and this
session verified it in the actual Go source before using it. The `pages.sections` drift
this left behind is fixed (migration `719`). **One real defect remains, undiagnosed**: the
newly-attached component's query-sourced fixture data (`items`) never populates during a
light rerender, even though one genuinely qualifying fact exists (Canelo vs Mbilli,
2026-10-31). The visible page now shows the correct EMPTY state ("no confirmed fixtures
yet") instead of the OLD "no calendar mechanism at all" — real progress, not yet the full
fix. Reproduced 3×, root cause not found; needs a fresh diagnosis pass, not another guess.

**428** — Unchanged in substance since yesterday; the one open item (admin-dashboard
frontend deploy) is now DONE and independently re-confirmed at the artefact (live pods
serving the release-surface UI, checked by grepping the actual served JS). Nobody has yet
used the release button on a real flagged item — that's a human/content decision, not
code.

## 1. Decisions/actions that need a person, not a session

1. **Someone should sit down and actually diagnose the `query.upcoming_events`
   items-not-populating defect.** This is the one piece of user-visible value 427 still
   owes. Full reproduction recipe, what was ruled out, and two untried next steps are in
   `NOTES_bugfix_427_event_render.md`'s 2026-09-03 entry (search for "carrying stored" /
   "carry-vs-fresh-render"). This is a strong candidate for
   `090_TRIGGER_needs_diagnosis` per CLAUDE.md's own rule — cross-cutting-shaped (the
   query-resolve mechanism is shared by every `query.*`-sourced component), non-obvious
   after real effort, and a wrong guess here would ship visibly to a paid customer's page.
2. **Resubmit the `ff91e666` council round** (event-list component, migration 712) — it's
   sitting at REVISE. The gating objection is already answered (component_swap works,
   verified live); the submission text just needs to say so. Not a code change, a
   resubmission with updated rationale.
3. **A human should actually use bug 428's release surface** on a real flagged verdict —
   the tool is live end to end now, nobody has clicked it yet. Worked case ready to hand:
   boxingonline's own `e3c2b440-c006-40ec-be7a-88d0b689ed1e`.
4. **Same open decisions as yesterday's handoff, still open**: whether `news_feed_ingestion`
   extraction should run beyond boxingonline.com; arming `event_fixture_completeness` in a
   live `run_checks.config.checks` array.

## 2. What was verified this session (don't re-derive)

- **`event-list` is live on the served page's underlying data, not yet showing fixtures.**
  `page_components` for `tool-fight-calendar`: `hero-tool` + `event-list`, both
  `build_status='deployed'`. `pages.sections = ["advertising","hero-tool","event-list"]`
  (migration 719). Rendered HTML currently shows the guarded empty state
  (`event-list-empty`), not the Canelo/Mbilli fixture — that's the open defect.
- **The deploy chain, checked at every hop, not assumed from a status column**: DB row →
  git commit (`007b3a7a1`) → GitHub Actions run `33672753667`'s own "Sync to B2" step
  (`upload tools/fight-calendar/index.html`) → Cloudflare cache purge for boxingonline.com.
  `gh` was authenticated this session (`gh auth status`) — use it for this kind of check,
  it beats `kubectl logs -l app=github-actions-runner`, which interleaves every site's
  deploy jobs across 3 runner pods with no per-job body.
- **The preview subdomain (`boxingonline.ugg2.com`) is a SEPARATE reconciliation target**
  from the GH-Actions-driven `portfolio-sites/<domain>` sync — `site-publisher`/
  `publish_site_action.go`, needs a spawned storage-credentialed pod, not directly
  dispatchable from the standing chassis. Not chased this session (the artefact that
  matters — `portfolio-sites/boxingonline.com` — was already confirmed correct). If a
  future session needs the PREVIEW to show a change sooner than its own tick, that's the
  action to look at.
- **Fresh chassis build confirmed live, both directions**: `agent-chassis` (both pods
  uniform) and `core-manager`'s own startup log both show `git_commit =
  7bf1ff674021f2d57dfd0aa41324541070646c3a` — 650 commits ahead of the build the
  2026-09-02 handoff checked. `git merge-base --is-ancestor` confirms this bug's every
  commit, including the REVISE-round citation-gate fix (`987ed3b3b`) and the
  `event_fixture_completeness` check (`d6a952249`), are ancestors. **Note:**
  `service_binary_capabilities` can carry a STALE leftover row from a spawned pod that
  never returned — filter/order by `last_seen_at DESC`, don't trust a bare `DISTINCT
  git_commit` (this session was briefly confused by exactly that).
- **admin-dashboard is `v1.0.1356`, confirmed twice independently** (this session and, per
  bug 428 §11, `gap planner` at an earlier tag) by exec-ing into a live pod and grepping
  the served JS for `Record verdicts only` — not by trusting the deployment's image field
  alone.

## 3. Quick-reference commands

```bash
# Council round still needing resubmission
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='ff91e666-608d-4b26-9c41-d97d23a21437' AND kind='council_report' ORDER BY created_at;"

# The open defect's current state (re-run any time; costs nothing, changes nothing)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT pc.content_data->'items' AS items, length(pc.rendered_html)
FROM page_components pc WHERE pc.page_id='4b74ff1f-455a-4bb2-b81d-e1d0ec824f33' AND pc.slot_name='event-list';"
# 'items' should be non-empty once fixed — it is currently NULL/absent.

# The one fact that should be showing and isn't
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT jsonb_pretty(f) FROM site_specs s, jsonb_array_elements(s.data->'facts') f
WHERE s.site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND s.aspect='evidence_base' AND s.is_current
  AND f->>'id'='CIT-5b2cc9894bfc475f';"

# Confirm what commit agent-chassis/core-manager are ACTUALLY running (re-check — more ships constantly)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT pod_name, git_commit, last_seen_at FROM service_binary_capabilities
WHERE service='agent-chassis' ORDER BY last_seen_at DESC LIMIT 5;"
kubectl -n ai-persona-system logs -l app=core-manager --tail=1000 | grep -m1 'build provenance'

# admin-dashboard, checked at the artefact
kubectl -n ai-persona-system get pods -l app=admin-dashboard -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.startTime}{"\n"}{end}'
kubectl -n ai-persona-system exec <a-live-admin-dashboard-pod> -- sh -c \
  "grep -c 'Record verdicts only' /usr/share/nginx/html/assets/*.js"
```

## 4. Everything else is in the bug files and the workstream docs

Both bug files carry the full diagnosis, every correction, every fix candidate and every
council correlation. `bugfix_427_event_render/` carries the RUNBOOK (dispatch-a-single-
action-directly pattern, tracing a deploy to the real GH Actions log, applying one
migration by hand without sweeping the pending directory, verifying a frontend image's
content before pushing — all new this session) and the plain-English
`README_where_we_are.md`. `bugfix_428_planner_deferral/` is unchanged in substance this
session beyond its own README append confirming the frontend deploy.
