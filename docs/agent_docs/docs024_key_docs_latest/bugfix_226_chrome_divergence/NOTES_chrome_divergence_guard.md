# NOTES — bugfix 226 chrome divergence guard (append-only, newest at the bottom)

## 2026-08-08 — session 1 (bug selection + research)

- Picked 226 from `bugs_open/` after ownership sweep: `who-owns.py` clean on
  223–227; live-transcript grep found the FILER (session `0c5a11f2`, closed out
  with a final summary) and a second bug-picker (`98b5904b`) that declined 226
  ("ink is wet") and took 209. 223 self-routes to `architecture_review`;
  224/225 were touched today by the SQAM-003 arithmetic-oracle commit. 226:
  unowned, no open `site_work_items` on the mechanism (only unrelated
  favicon/og-card `image_url_404` rows).
- Premise re-verified at HEAD: the store UPDATE
  (`render_site_components_action.go:938-943`) replaces `rendered_html`
  unconditionally behind the 069 lock predicate; no comparison, no archive.
- **Misstep avoided (and worth recording): candidate 1 in the bug file is not
  implementable as written.** It says "re-render with the row's stamped inputs
  (117 stores them)". `render_inputs` is a jsonb of md5 DIGESTS
  (`datahelpers/chrome_render_inputs.go`) — nothing can re-render from it. The
  file's own "What already exists" section describes the fingerprint correctly
  two paragraphs earlier. Caught by reading the helper before planning, i.e.
  the LANDMINES "fix candidate refuted by its own file" check. Recorded as a
  CORRECTED note in `bugs_open/226` rather than WRONG_CALLS (it was never
  asserted by me as true; the check is the point).
- DB state, measured: `site_components` 57 rows, 57 with HTML, **11 stamped**
  with `render_inputs` (the 117 fix is built but the wave rides the next roll —
  which is the urgency: 46 unstamped rows get rebuilt when it lands).
- Prior art found: `page_component_history` (14,396 rows) is the house archive
  shape — but it archives `content_data` only; `rendered_html` is unarchived
  even there. Scoped OUT of 226 (page side has its own lanes); noted as a
  possible follow-on.
- Writer inventory for `site_components.rendered_html`: 6 classes (render
  overwrite, relink-erase, set/replace, append, core-manager admin dynamic SQL,
  raw psql). This is what decided trigger-over-call-sites: a Go guard can never
  see the psql writer, and the psql writer (mig 268 style) is the bug's origin.
- sqlmock (`DATA-DOG/go-sqlmock v1.5.2`) is in go.mod and used by neighbouring
  tests — behaviour tests are possible without a live DB.
- Migration home confirmed: `docs/agent_docs/sql_for_agents/`, next free number
  **344** (343 is the highest on disk); ROLLBACK sidecar naming per 339/340/341.
- Register: entry goes in `styling-render-pipeline.md` as **STY-054**.
  `rebuild-cascade.md` is DIRTY in the shared tree (another session's edit,
  3 add / 3 del) — deliberately NOT touching that file; same-file-passenger
  risk.

## 2026-08-08 — session 1 (implementation)

- Council submission: corr `cffbfec4-3bec-4577-8844-d17c546ded3e` (8 edits,
  trigger-over-call-sites rationale, risks named incl. fail-closed and
  grep-invisibility). Committing with `Council-Submitted:` per the 07-30 rule;
  budget ~30 min for the verdict, find the run by payload not printed id.
- Migration 344 applied to live DB at ~23:30Z and RECORDED
  (`--record-only`, note names the probe). Verified live:
  `trg_site_component_archive` enabled (`O`), `site_component_history` = 0
  rows (probe rows self-deleted), 0 digests stamped (no backfill — by
  design). **Did NOT use `--apply`: dry-run showed the pending backlog
  contains other threads' files** (335, 337, 338, 340×3, 341×2, 342, 343 —
  several "LIKELY ALREADY APPLIED" per their own guards). The 98-pending
  landmine is real; single-file psql + record-only is the honest route.
- In-file probe design note: the probe exercises the NEGATIVE first (byte-
  identical rewrite must archive nothing — the check-the-no-op-case rule),
  then archive + restore, then deletes its own ledger rows. `updated_at` is
  untouched (probe UPDATEs do not set it).
- Go: `site_component_divergence.go` (classify + emitter mirroring the 069
  emitter), store statement gains `rendered_html_digest = md5($1)` same-
  statement. 4 tests green; `gofmt` clean.
- **Pre-existing red at HEAD, NOT this lane's:**
  `TestValidDocSubjectTypes_LockstepWithMigrationCheck` fails because
  `e1628f7df` (RFC_015 decision records, 2026-08-08 20:21) shipped migration
  `340_doc_notes_decision_subject_type.sql` (adds `'decision'`) without
  adding it to `validDocSubjectTypes` — the FOURTH instance of the
  both-halves-in-one-commit landmine (LANDMINES.md:646). Owning sessions were
  live at 23:26 (mtimes), so left to them; flagged in this lane's close-out.
  `go build` is unaffected; only `go test` reddens.

## 2026-08-08 — session 1 (council round 1 → REVISE → round 2)

- Round 1 verdict ~22:31Z: **REVISE**, decided by a gating objection from the
  `guardian` seat. Full report parsed from `diagnosis_artifacts.body` (note:
  the column is `body`, not `content` — schema-first would have saved one
  errored query; logged in WRONG_CALLS as the cheap check).
- **Objections REFUTED by measurement (no change made):**
  - guardian HIGH "item_key has no site_id → only the first site ever gets an
    item": `idx_swi_dedup` is `UNIQUE (site_id, item_key) WHERE non-terminal`
    — site-scoped by the index; the two-strike guard is also site-scoped.
  - guardian "no GRANT for core-manager's role → admin edits hard-error":
    `pg_roles` has exactly `clients_user` (owns both tables, is every chrome
    writer's role), `auth_user` (no chrome), `diagnose_ro` (read-only). Empty
    gap.
  - reuse_agent "extend the existing content_hash convention": DEAD COLUMN —
    0/1294 `page_components.content_hash`, 0/619 `pages.content_hash`
    populated. Live `content_hash` uses are insert-dedup identity, different
    semantics.
  - prior_art "is page_component_history trigger-fed?": application-fed (Go
    INSERTs at 4 sites, no trigger on page_components) — the raw-psql gap was
    unique to this case, as the plan claimed.
- **Objections CONCEDED and fixed:**
  - editquality: within-site repeat suppression — item_key now carries the
    patched digest (first 12 chars) AND `recurrenceExpected: true`.
  - render_guardian: emit could fire for a lock-refused store — WARN + emit
    moved AFTER `RowsAffected > 0`; classify stays before the store (it reads
    the outgoing bytes). Source-pinned (order: RowsAffected → hand_patched →
    emit; exactly one call site).
  - editquality: vacuous mock negatives — negatives moved to source pins.
  - bug_historian: page-side sibling — **filed as `bugs_open/229`** (228 was
    taken by another session between my ls and my write — re-check the number
    at write time, not at plan time).
  - architecture/debug_historian: staged rollout + fail-closed weighing now a
    dedicated PLAN section; work-item closure design stated (needs_human_review,
    no verifier registration — records an event, not a re-checkable state).
  - debug_historian: pod-grep close criteria added to the bug file (positive
    symbol + negative removed-string, every replica).
- Round 2 submitted on the SAME trail: `RESUBMIT_CORR=cffbfec4…`, run
  envelope `b3587307`, orch `6905d256`. All tests green after revisions.

## 2026-08-09 — session 2 (roll verified, round 2 read late, round 3 in)

- **Misstep (WRONG_CALLS'd): the round-2 verdict watcher could never fire** —
  its `created_at > 23:40 UTC` floor was my BST clock unconverted; the verdict
  landed 22:49 UTC. Found by querying unfiltered. Round 2 = **REVISE**, gated
  by `bug_historian` again (229 filed-not-fixed), 10 seats approve/3 abstain.
- **v1.0.1270 rolled and VERIFIED** (the 153 discipline, with round 2's own
  correction): 65 pods run the chassis image (the one-image-many-labels trap,
  measured — enumerate by IMAGE, not label); all 3 chassis deployments at
  1270; both main `agent-chassis` replicas grepped
  `classifySiteComponentArtefact`=2, round-1-only string=0, round-2 string=1.
  ~60 spawn pods still at 1269 (pre-roll spawns) — residual, noted.
- **The archive worked in production before we asked it to**: 4 rows in
  `site_component_history` from 08-08 evening — webdesign.uk and
  leopardessconsulting.co.uk, header+footer each, all `unstamped` (correct:
  pre-roll code, no stamps exist), zero trigger errors. Trigger re-verified
  SOLE + enabled at the exact boundary the round-2 `debug_historian` asked
  about (Go half now live).
- **Round-2 dispositions**: classify-error silent path → FIXED
  (`readBackDivergenceFromLedger`: the trigger's own DB-side verdict on the
  archive row is the fallback; ErrNoRows = byte-identical = nothing lost;
  both branches sqlmock-tested). item_key consumers → measured empty (0 rows
  of the type have ever existed; 0/57 stamped makes hand_patched impossible
  so far). Callers → exactly one call site (grep attached). Pods wording →
  close criteria corrected + executed. Trigger-sole re-check → done 08-09.
  Roles + dead-column queries attached verbatim for the prior_art seat.
- **The gating scope question is NOT re-answered with code**: bug_historian
  (fix page_components now) vs architecture (fine at two instances; a third
  needs an RFC) now disagree on the record. Per the 07-28 scope-veto
  guidance: the split is written into `bugs_open/229` as an OWNER CALL block,
  and this lane holds 226 to chrome. Round 3 submitted on the trail:
  run envelope `b7518808`, orch `50924d69`.
- Close criteria state: (1) pod-verify **DONE** for 1270; (2) end-to-end
  protocol **NOT RUN** — needs the two-step (first rebuild stamps, then
  hand-patch, then rebuild → item); (3) 117 wave **NOT FIRED** (0/57 stamped;
  the 4 archive rows are pre-wave partial evidence).

## 2026-08-09 — session 2 (verdict: APPROVED)

- **Round 3 APPROVED at 09:08:37Z** — "approved with 2 advisory objection(s),
  none high-severity", 5 abstained. Verified at the artifact row
  (orchestration `50924d69`), NOT from the watcher's echo: the stale round-2
  watcher finally fired on this report and its output literally said
  "ROUND2 VERDICT: approved" — the label was a hardcoded echo string; the
  filter, blind to round 2, could only ever have matched round 3. A claim
  about behaviour is not the behaviour, including your own tooling's labels.
- Advisories carried forward: (a) ErrNoRows-in-read-back trusts the WHEN
  gate's completeness (STY-054 open-review d — any WHEN edit must re-justify);
  (b) page-side exposure = the 229 OWNER CALL (bug_historian, standing);
  (c) guardian: watch per-slot item accumulation post-wave (open-review e);
  (d) guardian: keep close criteria 2+3 marked NOT DONE — they are.
- Trailer: this session's final docs commit carries `Council-Reviewed:` —
  the verdict was READ first. The three code commits keep their
  `Council-Submitted:`; 098 resolves them at report time by design.

## 2026-08-09 — session 3 (close criterion 2 run end-to-end; wave observed starting)

- **Found on arrival: the 117 wave had started** — `needs_rerender` /
  `render_inputs_drift` items complete for leopardessconsulting.co.uk
  (08-08 17:21Z) and webdesign.co.uk (08-08 18:15Z), both PRE-roll (rebuilt
  unstamped, correctly), then dartsonline.com at 08-09 09:08:30Z POST-roll —
  3/3 slots stamped, all byte-identical, zero archive rows (the WHEN gate's
  no-op path exercised by real traffic). **Timeline correction worth having:
  the four 08-08 archive rows (webdesign.uk 22:29Z, leopardess 23:43Z) did
  NOT come from those wave items** — the wave's two pre-roll sites predate
  the trigger (~22:30Z, NOTES said "~23:30Z" which was a BST slip — the
  webdesign.uk rows land seconds after application) and archived nothing;
  the rows are other rebuild traffic. Nearly asserted the wave as their
  cause; the item-vs-ledger timestamps refuted it before it reached a doc.
- **Criterion 2 protocol, dartsonline.com** (wave had stamped it ⇒ step (a)
  free). Pods 33m old (outside the 300s window); footer snapshotted byte-
  exact first (psql `-At` adds ONE trailing newline; `octet_length` 2313 +
  1 = file size — account for it before trusting a snapshot as recovery).
  Log followers armed on BOTH main replicas BEFORE any dispatch (chassis
  retention is seconds; a follower armed after the fact proves nothing),
  filter proven against the literal WARN string first — ASCII fragment
  `were overwritten and archived` deliberately avoids the em-dash.
- **The psql patch itself is evidence**: appending the probe comment drew an
  immediate `machine_made` ledger row, `application_name='psql'` — the
  raw-psql writer class (the 268 incident class, the one no Go guard can
  see) is provably visible to the trigger.
- **Dispatch**: kcat with payload in the container COMMAND + `PUBLISH_OK`
  receipt (the 07-26 stdin-EOF landmine pattern). Both publishes landed
  first time and were consumed in ~5s — no queue backlog at 09:25Z (the
  25-36 min figure is a concurrent-load ceiling, not a floor).
- **Positive case, all signals by row identity**: WARN fired exactly once,
  on the rendering pod (zhz2g), inside my orch `322b266e`; ledger row
  `hand_patched` with archived md5 == patched md5 (`2ed6dd06…`), stamp-at-
  archive == old machine digest (which IS the classification);
  `application_name` carries zhz2g's pod IP (10.20.39.19 — matches the
  WARN's pod); item key `…:footer:2ed6dd067c5f` (digest fragment as
  designed); footer restored byte-identical to pre-patch, re-stamped,
  probe string absent. Ordering: archive .907s → item .931s (archive atomic
  with overwrite; item after RowsAffected>0, as source-pinned).
- **Negative control by identity, not absence**: second forced rebuild
  (orch `453b2eb6`) — `updated_at` bumped on all 3 slots (the writes
  HAPPENED), md5s unchanged, zero new ledger rows, zero new items, no WARN.
- **Cleanup**: probe item `05fda19d` cancelled with a note naming this run
  (a deliberate probe is not a human's queue item). Both orchestrations
  COMPLETED, no error; dedup held — 0 new `page_rerender` items from my two
  dispatches (34 triaged wave items were the non-terminal blockers).
- **Wave-watch runbook (criterion 3, for whoever checks next):**
  - `SELECT count(*) FILTER (WHERE rendered_html_digest IS NOT NULL), count(*) FROM site_components;` — 3/57 at 09:30Z, should climb
  - `SELECT count(*) FROM agent_error_log WHERE occurred_at > '2026-08-08 20:00Z' AND error_message ILIKE '%site_component_history%';` — 0 so far (column is `occurred_at`, not `created_at`)
  - guardian ratio: hand_patched ledger rows vs `chrome_divergence_overwritten` items per site+slot — 1:1 (the probe's own pair)
