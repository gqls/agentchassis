# HANDOFF 2026-08-09 — bugfix 229 page component archive: continue here

> **LANE COMPLETE 2026-08-09 (late evening). Nothing is owed.** Steps 1-4 all
> landed: council APPROVED round 4; the release arrived as v1.0.1276 and both
> main replicas pod-verified (4/1/2 + chrome control 2); the e2e protocol ran
> and passed on dartsonline "beginners" — WARN + exact-key item +
> md5-identical hand_patched ledger row, negative control silent, delete-arm
> archiving proven in production. Probe item cancelled with a note. Bug file
> header says DONE IN SUBSTANCE (stays in bugs_open per the 08-06 ruling).
> If you are picking this up cold: there is no task here — only the STY-056
> open-review watches ((a) volume growth, (e) the unsurfaced-writer sweep,
> whose driver is the 230 rotation once 083 drains).

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

1. ~~**Read the council verdict**~~ **DONE 2026-08-09 ~22:15Z: APPROVED,
   round 4 of the trail** — "approved with 1 advisory objection(s), none
   high-severity", 3 abstained, verified at the report body. The advisory
   (bug_historian: unenumerated writers archive but never surface) is STY-056
   open-review (e), query written out, driver = the 230 rotation once 083
   drains. Trail: r1 REVISE reuse_agent (the submission's own applied column
   read as pre-existing — answered, approved r2) · r2 REVISE editquality
   (edit list dropped a file my script split away) · r3 REVISE editquality
   (risks field was round-1 prose) · r4 APPROVED. Two real catches executed
   along the way: migrations 351+357 ledger-recorded; rollback drilled
   (BEGIN…ROLLBACK, guard NOTICE, triggers intact after).
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
