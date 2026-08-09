# HANDOFF 2026-08-09 — bugfix 229 page component archive: continue here

**Read PLAN (design + every measurement) and NOTES (evidence + the probe
misstep) beside this; this file is only what to DO next.**

## State in one paragraph

Owner ruled candidate 1 (extend the 344 shape) on 2026-08-09. The DB half is
LIVE: mig 357 applied and probe-verified on all four arms (no-op, machine_made,
hand_patched, delete-with-NULL-component) — every destruction of a page
component's artefact now archives to `page_component_history`
(`source='artefact_archive_trigger'`), fail-closed, from every writer
including raw psql. The Go half (same-statement stamps in four render/save
writers, divergence items from save_page_sections + rebuild_blog_listing,
ledger read-back fallback, 14 tests green with anchored pins) is committed —
`1930ef86f` — and **INERT until the next fleet release**. IMAGE_TAG bumped to
v1.0.1275 (`748a23af4`). Single-service rolls are forbidden (owner 08-03):
the owner runs `date; make release redeploy-agents ENVIRONMENT=production
REGION=uk001; date`. Council corr `eee2888b-20dc-46ba-9b1f-53e592374cba`
(run orch `52b1383b`), verdict pending when this file was written.

## Do next, in order

1. **Read the council verdict** — find by payload, never retry on a missing
   row: `SELECT current_step, status FROM orchestration_states WHERE
   collected_data->'input_data'->>'fix_correlation_id' =
   'eee2888b-20dc-46ba-9b1f-53e592374cba';` then the newest
   `doc_notes WHERE categories ? 'council-gate'`. REVISE → answer on the same
   trail (`RESUBMIT_CORR=eee2888b…`). Never write `Council-Reviewed:` on an
   unread verdict.
2. **After the owner's release: pod-verify** (the 153 discipline) — main
   `agent-chassis` deployment pods only (65 pods run the image; enumerate by
   image+owner kind, not label):
   `strings /app/agent-chassis | grep -c classifyPageComponentArtefacts`
   (expect >0) on each replica, plus a NEGATIVE control (a string the build
   should NOT contain — pick from an older commit if one was removed; at
   minimum confirm the tag AND date it).
3. **E2E protocol, one table over** (the bug file's verification section):
   (a) run a real save on a test page (dartsonline has 30+ pages) so its
   sections get stamped; (b) psql-patch one section's `rendered_html`
   (expect an immediate `machine_made` archive row — the raw-psql writer
   class proving itself); (c) rerender/save again → require the WARN
   (`were overwritten and archived (bugs_open/229)`), the
   `page_divergence_overwritten` item (key `page8:position:digest12`), the
   `hand_patched` ledger row (match by md5, not by time — same-transaction
   rows share created_at); (d) negative control: untouched byte-identical
   save → no new rows, no items; (e) DELETE recoverability: a full save's
   DELETE leaves op='delete' rows for every removed section. Arm chassis log
   followers BEFORE dispatching (retention is seconds).
4. **Then done in substance** — the bug STAYS in `bugs_open/` (owner 08-06);
   update its header, the memory topic file
   `bugfix-229-page-component-archive.md`, and the MEMORY_workstreams line.

## Standing landmines

- The archive is a DB TRIGGER PAIR — no Go grep shows it. Fail-closed: a
  page write erroring with `page_component_history` is the trigger refusing
  to destroy unarchived bytes; fix the ledger, never drop the trigger
  casually (ROLLBACK sidecar exists).
- **Same-transaction archive rows share `created_at`** — select by bytes.
  The mig 357 probe failed its own first apply on this.
- Delete-arm rows: `component_id NULL`, identity in slot_name/position.
  Full-page deletions archive nothing (structurally forced, stated).
- `rendered_html_digest` is the render/save paths' only; adopt_verbatim
  deliberately unstamped (test-pinned).
- A THIRD table adopting this shape needs the shared-abstraction RFC
  (architecture seat's recorded condition, in STY-056).
