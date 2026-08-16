# NOTES — site_delivery_and_editor — append-only, newest at the bottom

## 2026-08-14 — workstream created; plan approved; Phase 1 landed same evening

- Owner-approved plan (verbatim) in `PLAN_2026-08-14_site_delivery_and_editor.md`.
  Origin: the owner's Netlify+editor idea + a reference PDF
  (`~/Downloads/Automated_CMS_Architecture.pdf`), reviewed critically. Key
  rulings taken 2026-08-14: provider bar = COMPLETELY automated (Netlify not
  presumed → plan recommends Cloudflare Pages Direct Upload, seam keeps
  Netlify-OAuth available); ONE multi-tenant editor service; the framework
  KEEPS write access (locks are the two-writer referee); editor after
  handover only.
- Phase 1 (the "usually next day" rider) executed by the ai_site_selling
  lane the same evening: register supersede `SQL_2026-08-14b` (fact + wire +
  ban, ban proven non-inert 13 findings), three locked heroes moved
  surgically (href-asserted), five `nextday_*` content_rewrite items
  dispatched. Acceptance = served grep, not item status.
- Exploration facts the plan rests on (do not re-derive): zero Netlify code;
  zero CF-API automation; `*.ugg2.com` live via the portfolio-sites worker;
  `S3Client` binds one bucket and nothing binds portfolio-sites; no
  archive/zip anywhere; customer auth does not exist; no Ingress; the
  CTS-003 PATCH endpoint + auto-lock + history is the write path to extract;
  `apply_section_edit`'s build_status='approved' defect disqualifies it for
  customer edits (file that defect separately).
- Sharpest risks, ranked: (1) tenant scoping must be structural
  (session→site_id once; cross-tenant probe is the acceptance), (2) CF Pages
  partial-upload — acceptance is served-hash equality never API 200,
  (3) truncated ZIP = silent contractual failure — stream + alert.

## Coordination

- webdesign.uk's CTA components are permanently locked (SQL_2026-08-12k);
  the 268 fleet-fix thread owns unlocking. Phase 2's canary must be a quiet
  portfolio site, NOT webdesign.uk.
  > **CORRECTED 2026-08-15:** the locks are now OFF — 268's unlock ran
  > (verified: exactly 1 permanent lock on webdesign.uk, the sibling's chat
  > box). The canary rule stands anyway: webdesign.uk is the shop, not a
  > test surface.
- Council: one run per phase (2–6), submit before/alongside each shipping
  commit; register entries listed in the PLAN roll-up.

## 2026-08-15 — Phase 2 begun (session start after the v1.0.1301 roll)

- Chassis rolled to v1.0.1301 at 10:14Z (both replicas, image label read at
  the pod). Post-roll falsifier sweep from the ai_site_selling handoff §5,
  all HOLDING: webdesign.uk permanent locks = 1 (chat box only); no work
  items created since the roll target webdesign.uk (the post-roll items are
  mortgagecalculator/ai-agent-orchestration/webdesign.co.uk); link gate 5/5
  with self-test control correctly failing first; guide page serves 200 at
  `/guides/tool-website-brief-starter-guide.html` with `/what-you-get.html`
  present ×4 (the 09:22Z deploy was the monitor's third restore of that
  link); Stripe keys still ABSENT from `personae-platform-secrets` (owner
  gate unchanged); no `CF_PAGES_API_TOKEN` yet (expected — Phase 2 will need
  the owner to mint one; the B2+ugg2 backend needs no new credential);
  `bugs_open/271` still open (writer steering stays via writer_block).
- Migration state: latest number is 411 (`sql_for_agents/`), so Phase 2's
  migration takes 412 — **re-check at write time; creating the file does not
  reserve the number** (LANDMINES). Runner dry-run run this session; pending
  list is other threads' files — any apply must be scoped to our single file.
- Phase 2 build starting per PLAN Part 2e: `platform/publish/` seam +
  `publish_site_action` (no-op on NULL `sites.publish_target`) + scheduler
  reconciler on hash drift. CF Pages backend code ships now; its live
  verification is gated on the owner's token. Day-one verifiable backend is
  B2→ugg2 (no new credential).
- Landmine carried into the acceptance design: **hash comparisons must take
  bytes from the ORIGIN store, not through the CDN** — the edge rewrites
  some files (measured: robots.txt "Cloudflare Managed"); served-hash
  equality checks compare origin B2 bytes vs the pages.dev response, and any
  robots.txt mismatch gets checked against that landmine before being called
  a defect.

## 2026-08-15 (later) — Phase 2 BUILT: seam + action + reconciler seed; council submitted

- **What shipped** (one commit, `Council-Submitted: 21aba3f5`): `platform/publish/`
  (seam, `TreeHash`, `S3Source`, b2worker backend, cfpages as a LOUD refusal),
  `publish_site` action + registry entry, migration 412 (APPLIED — columns live,
  all 40 sites NULL=OFF, DO/RAISE guard asserts no default), seed 422 HELD
  (`_HOLD.sql`, apply commands in its header), register entry DGH-008.
  Tests green including the negative proofs: a corrupted destination copy fails
  publish; acceptance failure provably writes no `published_hash` (sqlmock
  ExpectationsWereMet with no UPDATE expected).
- **Design decisions and their reasons** (the PLAN's shape held; three calls made
  here): (1) **cfpages ships unarmed** — no owner token exists, the Direct
  Upload protocol is partly undocumented, and a blind protocol client behind an
  APPROVED verdict is the MDL-040 class; b2worker is the provable day-one
  backend, cfpages gets built verify-as-you-go at arming time. (2) **The
  execution pod is forced by credentials, not chosen**: standing chassis has no
  B2 creds (owner 2026-08-08, bugs_open/245 "do not re-add"), so the reconciler
  is scheduled_tasks → `publish-reconciler` (chassis) → spawn→call
  `site-publisher` (storage allow-listed) → `publish_site` constructs its own
  portfolio-sites client. (3) **site-publisher repurposed, not a new type** —
  it is a pre-070 fossil (`upload_to_s3` into nonexistent bucket "websites")
  already on the spawner allow-list + topic_manager; enumerated consumers
  before touching: 0 workflows / 0 schedules / 0 work items / 0 orchestrations
  all-history. 422 snapshots it and guards on the fossil's exact pre-state.
- **Misstep logged**: my first council JSON was invalid (an unescaped line
  break where a closing quote belonged — jq caught it before the trigger did).
  And mid-session the migration numbers moved 411→421 under me (the
  unreserved-number landmine, live demonstration): 412 was applied before the
  jump so it was safe; the seed took 422 after a re-list.
- **Still owed for Phase 2 completion (next session / post-roll)**:
  1. ~~Bump IMAGE_TAG, build, roll~~ > **CORRECTED same session:** releases
     are WHOLE-FLEET and the owner runs `make release` — a one-service apply
     at its own tag is the recorded trap. The code rides the NEXT fleet
     release; after it, verify at the binary (provenance stamp, per SERVICE:
     `git merge-base --is-ancestor 71e4d9736 <the stamp>`).
  2. Hand-apply 422 per its header (psql + `--record-only`), AFTER the stamp
     proves the image carries `publish_site`.
  3. Canary: pick a quiet portfolio site (webdesign.uk FORBIDDEN — it's the
     shop), set `publish_target='b2worker'`, `publish_project='<slug>.ugg2.com'`.
     ⚠ the `*.ugg2.com` wildcard route+DNS already exist (worker serves any
     hostname prefix, zero per-site config) — no Cloudflare work needed.
  4. Acceptance per PLAN: served-hash equality on the canary; second sweep with
     no change performs zero publishes (`skipped: no drift` in the result);
     every other site still NULL (`SELECT count(*) ... WHERE publish_target IS
     NOT NULL` = 1).
  5. Read the council verdict (corr `21aba3f5-ca44-4220-a680-d99f5ef0a90b`,
     budget ~30 min queue) — REVISE objections get worked, not defended.
- Council submission file kept at the session scratchpad
  (`council_publish_seam.json`); the artifact trail lives under the correlation
  either way.

## 2026-08-15 (evening) — council round 1: REVISE, objection CORRECT; seed fixed and resubmitted

- Verdict landed ~15 min after submission (dispatch was fast today, not the
  measured 29-min queue). **REVISE — gating objection from `debug_historian`,
  and it caught a real defect of mine, not a style point**: seed 422's
  "snapshot before repurpose" was a hand-rolled INSERT copying only 9 columns
  — a restore from it would have silently dropped the fossil's `topics`,
  `capabilities` and image fields. The estate already had the sanctioned
  mechanism (`snapshot_agent(type, reason)` → full-row copy into
  `agent_definitions_backup`) AND two documented landmines about exactly this
  (dual overloads writing to two different tables; backup rows copying
  id/created_at so only `snapshot_reason` + `snapshot_taken_at` identify a
  snapshot). I grepped LANDMINES for my file/table footprints but not for
  "snapshot" as an operation — the miss was in the verb, not the noun.
- Fixed in the held seed (all four objections):
  1. (HIGH) two-arg `snapshot_agent()` with reason `'422 pre-repurpose:
     upload_to_s3 fossil'`, then the snapshot is verified to hold the
     PRE-change config before anything mutates (exists ≠ restorable).
  2. (MEDIUM) re-application is a graceful NOTICE no-op on the applied state
     (no snapshot stacking); unknown pre-state still raises; the UPDATE gets
     a post-condition read-back.
  3. (LOW) the header now names the explicit pod-verification commands
     (provenance stamp + `git merge-base --is-ancestor 71e4d9736`, with the
     scrolled-stamp fallback) BEFORE the apply commands.
  4. (editquality, low) unarmed-cfpages opt-in stays possible by design —
     noted in the header with the reason a CHECK constraint would be worse.
- Rollback file rewritten to restore from `agent_definitions_backup` by
  reason + `snapshot_taken_at DESC`, refuse a snapshot not holding the
  pre-change config, and stamp `restored_at`.
- **Resubmitted on the SAME correlation** (round 2, run orch
  `507d6e87-cbf9-4ad8-9d13-725521416edb`); verdict-read still the owed step.

## 2026-08-15 (late evening) — round 2: APPROVED (3 advisory objections, none high); guards landed

- **APPROVED** on corr `21aba3f5` — the `Council-Submitted:` trailers on
  `71e4d9736` and `cd5490866` are credited automatically by the 098 report;
  no amend (forward-only). Advisories triaged:
  - **ACTED ON — publish_project uniqueness (editquality, medium)**: two
    sites sharing a `publish_project` would silently overwrite each other's
    hosted prefix. Migration `423_publish_project_unique.sql` APPLIED: partial
    unique index on `sites(publish_project) WHERE NOT NULL` — the bad state is
    now unrepresentable. ⚠ **number collision**: another session took 423
    concurrently (`423_finance_directory_researcher_named_firm_rule.sql`) —
    two unrelated migrations share the number; the full FILENAME is the
    identity (schema_migrations records it), do not "fix" by renaming.
  - **ACTED ON — re-verify enumeration at apply time (prior_art, medium)**:
    the 422 header now carries the four queries; two legs read tables outside
    some seats' schema tier, so the applier re-runs them, never trusts the
    rationale's figures.
  - **ACTED ON — pod-probe controls named (debug_historian, low)**: 422
    header now spells out the positive+negative control pair (zero-sha must
    NOT match — else the probe is reading Go's digit table).
  - **RECORDED, not reworked**: (reuse, medium) b2worker-in-Go vs shelling to
    `b2 sync` — the spawned chassis pods carry no b2 CLI, and adding a binary
    dependency + credential plumbing to the image for a copy the S3 API does
    in-process was judged the worse trade; noted here as the answer the
    submission should have carried. (editquality/guardian, low) B2
    read-after-write lag can only produce a false `accepted:false` (drift
    stands, next tick retries) — self-healing by design. (reuse, low)
    `contentTypeFor` duplication and a shared fetch/verify helper — candidates
    for the Phase 3+ passes. (architecture, low) **write up Direct Upload
    protocol decisions BEFORE arming cfpages** — added to the arming step's
    obligations, do not let the token's arrival shortcut it.
  - **tooling_provenance's doc_notes check, answered**: the only prior
    `site-publisher` notes are dormant-agents sweeps (07-22, 07-26) listing it
    as dormant — corroborating the fossil status, no conflicting context.
- Phase 2 session state: **everything session-local is done.** Remaining is
  the post-release sequence in the owed list above (release → verify pod →
  422 per its header → b2worker canary).

## 2026-08-15 (night) — the release landed: roll verified, 422 APPLIED, canary armed

- **Roll verified on v1.0.1303** (both replicas, started 18:45Z). The
  provenance log line had rotated out entirely (~1h45m of high-volume logs),
  so the binary probe carried it: stamp `5e075a6f9…` present on BOTH
  replicas, three sibling commit shas absent (the probe discriminates),
  random-hex negative control clean, and `git merge-base --is-ancestor
  71e4d9736 5e075a6f9…` holds — **the running image carries `publish_site`.**
- **Misstep → landmine**: my own 422 header prescribed the all-zeros sha as
  the negative control and it FIRED — forty zeros is git's null-sha CONSTANT,
  genuinely present in git-aware binaries. A false CONTROL-BROKEN that would
  have stalled the seed. Filed in LANDMINES (synced) and the header now
  prescribes a random 40-hex value.
- **Seed 422 APPLIED ~22:00Z** after the 4-leg enumeration re-ran = 0/0/0/0.
  Two apply-time discoveries, both fixed in the file:
  1. `agent_definitions.agent_category` is CHECK-constrained
     (`check_ad_category`: strategist/executor/analyst/integrator/
     coordinator/specialist or NULL) — my `'orchestrator'` was rejected,
     first apply rolled back CLEANLY (single transaction, snapshot included).
     Re-applied with `'coordinator'`; `category` (unconstrained) keeps
     `'orchestrator'`.
  2. The runner REFUSES `--record-only` on UPPERCASE-suffixed sidecars, by
     design — a `_HOLD` file is hand-run AND hand-tracked. The apply record
     is the file's own header STATUS line + this entry; the header's old
     record-only instruction is deleted.
  Post-apply state verified by the seed's own DO block: site-publisher
  repurposed (fossil snapshot in `agent_definitions_backup`, reason
  `'422 pre-repurpose: upload_to_s3 fossil'`), publish-reconciler inserted,
  schedule enabled at 600s.
- **Canary armed ~22:04Z**: `noted.co.uk` (chosen for 0 open work items vs
  ~24 on oufe/cookly) → `publish_target='b2worker'`,
  `publish_project='noted.ugg2.com'` (target prefix confirmed EMPTY in the
  bucket; source tree confirmed present; origin `index.html` sha256
  `b4416c32…` captured for independent acceptance). Exactly 1 site opted in
  fleet-wide. Timing note: the schedule's first tick (22:01:23Z) preceded the
  opt-in, correctly gate-skipped on zero rows, and stamped — so the first
  REAL pass runs at ~22:11:23Z, not immediately. A monitor watches the
  chain (stamp → orchestration → published_hash).
- **22:11–22:25Z — the canary RAN and caught a real defect.** Full chain
  executed (tick → stamp → publish-reconciler → spawned
  `agent-site-publisher-c08f7091-rl8hc` → publish_site) and failed at the
  first upload: **HTTP 411 MissingContentLength** — B2's S3 gateway refuses
  a bare stream; `copyOne` piped `Download()` straight into `PutObject`.
  The spawn→call handshake worked first try (no race), the spawned pod HAD
  its storage env — the seam's whole locality argument held; only the
  upload body shape was wrong. **Why the tests missed it**: my fake's
  `Upload` just read the stream — the defect lived at the environment
  boundary the double couldn't see, the same MDL-040 class the cfpages
  refusal exists to avoid. **Fix `b4981634d`**: buffer to `bytes.NewReader`
  (seekable ⇒ SDK computes Content-Length); BOTH test fakes now refuse a
  non-seekable body, and the guard is mutation-proven (reverting to
  streaming fails the copy test with the production failure shape). Zero
  objects landed (first file failed), `published_hash` never written —
  the acceptance gate held. Canary DE-ARMED (`publish_target=NULL`,
  project kept) until the fix rides the next owner release; re-arm recipe
  in the handoff §1. Phase 3 note: the ZIP action must NOT buffer whole
  objects this way — stream with known length or use multipart.

## 2026-08-16 (morning) — "a fresh chassis has been deployed": checked, and it is the SAME roll

- Session opened on the premise of a new roll. Measured before acting on it:
  every service and both chassis replicas still on **v1.0.1303**, started
  18:45Z 2026-08-15 (the roll verified last session); latest ReplicaSet
  `584b6fcf` unchanged, deployment revision 832 not advanced; no
  `v1.0.1304+` in the registry or the local docker cache; makefile
  `IMAGE_TAG ?= v1.0.1303`; local 1303 image built 18:31Z, i.e. BEFORE
  `b4981634d` (23:15Z). Binary probe on `agent-chassis-584b6fcf-9mtqd`:
  fix sha ABSENT, stamp `5e075a6f9` present (positive control), and
  `git merge-base --is-ancestor b4981634d 5e075a6f9` fails — the running
  binary predates the fix. **Re-arming now would reproduce yesterday's 411
  every 600s**, so the canary stays de-armed (confirmed: `publish_target`
  NULL, 0 sites opted in, 0 reconciler runs since 22:30Z).
- Nothing to update on the seam side; the handoff §1 recipe is still exact
  and still waiting on the first release that includes `b4981634d`. The
  check to run at the next "it has rolled" prompt is the one above — image
  label per service, then the ancestry test — BEFORE touching
  `publish_target`.

## 2026-08-16 (afternoon) — v1.0.1304 carries the fix; canary RE-ARMED

- **The roll is real this time** (the morning check refuted a premature one;
  this is the discipline paying off rather than a contradiction): both
  chassis replicas on **v1.0.1304**, started 10:41Z, image built 10:28Z,
  makefile `IMAGE_TAG ?= v1.0.1304`.
- **Verified at the binary, with the stamp identified rather than assumed.**
  The provenance log line had already scrolled (pods ~5h old). Probing the
  fix sha directly returned ABSENT — which is CORRECT and not a failure:
  **the stamp carries the HEAD the image was built from, not your own
  commit**, so the test is ancestry, not presence. Old 1303 stamp
  (`5e075a6f9`) also absent ⇒ genuinely a different binary. Walked recent
  commits against `/proc/1/exe` to find the actual stamp:
  **`5de6cddbe` (2026-08-16 11:25 BST)**. Then
  `git merge-base --is-ancestor b4981634d 5de6cddbe` → **TRUE: the running
  binary carries the Content-Length fix.** Controls both clean: random hex
  absent; a commit made AFTER the stamp absent (the probe discriminates).
- **Canary re-armed 15:56Z**: `noted.co.uk` → `publish_target='b2worker'`
  (`publish_project='noted.ugg2.com'` was retained from yesterday), exactly
  1 site opted in fleet-wide. Last reconciler tick 15:50:44Z ⇒ first real
  pass ~16:00:44Z. Monitor watching stamp → orchestration → published_hash.
- Acceptance unchanged (handoff §1b): served sha256 at
  `https://noted.ugg2.com/index.html` must equal origin
  `b4416c3208f9df047c044a526246f06c4fca03c4b02ec470e9e6af4e01f82ceb`;
  then a second tick must report `skipped: no drift` with `published_at`
  unchanged.
- **16:01Z — THE CANARY PASSED. The publish seam works end to end.**
  Chain, every link observed: tick 16:01:14Z → `site_publish_checks` stamped
  → `publish-reconciler-orchestrate-0816-1601` → spawn handshake succeeded
  FIRST TRY (pod `agent-site-publisher-bdcaf73e-krk56`, Running 16s in) →
  `publish_site` copied the tree → served-bytes acceptance passed →
  `published_hash = th1:05a06351…` written at 16:01:40.900765Z, orchestration
  COMPLETED. **Verified at the artefacts, not the status**: 8 objects at
  origin, 8 at `noted.ugg2.com/` (recursive count both sides); served
  `https://noted.ugg2.com/index.html` returns 200 and its sha256 is
  `b4416c32…` — **byte-identical to the origin hash captured BEFORE any
  publish existed**, which is the whole acceptance. Isolation holds: exactly
  1 site opted in fleet-wide.
  The 411 fix (`b4981634d`) is therefore proven in production, not just in
  test: yesterday this failed on the first file, today the same tree copies
  clean on the first attempt.
