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
- **16:11Z — NO-DRIFT PROVEN, and checked for the RIGHT REASON.** The second
  tick's `published_at` was unchanged — but a FAILED tick would look
  identical, so the result payload was read: orchestration COMPLETED with
  `{"published": false, "skipped": true, "reason": "no drift", "tree_hash":
  "th1:05a06351…"}` — and that hash equals the stored `published_hash`
  exactly. **PHASE 2 IS COMPLETE**: publish on drift, don't publish without
  it, and never record success the served bytes don't support.
- **Operational finding + change made the same hour**: every tick spawns a
  `site-publisher` pod EVEN WHEN THE ANSWER IS NO-DRIFT — the drift check
  needs bucket credentials, so it cannot run on the chassis. Measured: 2
  lingering pods after 2 ticks (11m-old one still Running), i.e. ~144
  pod-spawns/day per opted-in site to answer "nothing changed". **Interval
  raised 600s → 3600s** (live UPDATE + the seed file's value and rationale
  updated together, so a fresh apply matches production). A built tree
  changes a few times a day at most; an hour of mirror latency on a
  delivered site is immaterial. If a future backend can answer drift without
  credentials, this becomes a chassis-side pre-check and the interval can
  come back down.
- **Canary LEFT ARMED** (noted.co.uk → noted.ugg2.com): it is now the seam's
  continuous live proof at one no-op tick/hour. One `UPDATE sites SET
  publish_target=NULL WHERE domain='noted.co.uk';` reverses it.

## 2026-08-17 — DELIVERY ARCHITECTURE DECIDED (five discussion rounds with the owner); full record in PLAN_2026-08-17_delivery_architecture_decisions.md

- **The v1.0.1305 roll (14:43Z today) is a NO-OP for this lane, checked not
  assumed**: no inert code awaiting a roll (`b4981634d` shipped+proven in
  1304); canary undisturbed (`published_hash` still the 08-16 16:01:40Z
  value ⇒ the reconciler correctly found no drift across the roll). Do not
  re-verify. ⚠ Fleet memory the same day: a "fresh build" CAN ship no new
  code (same-tag rebuild serves the cached image) — this lane's check did
  not depend on the roll shipping anything.
- **OWNER DECISIONS (each with its reason):**
  1. **Fees SEPARATE** — domain "keep this domain — £10/mo"; hosting priced
     deliberately HIGH because the owner does not want the hosting
     business; always offered beside FREE connect-your-own-Netlify (and
     possibly other third parties). Supersedes this session's earlier
     one-combined-subscription sketch.
  2. **No free custom-domain serving** — the custom domain serves a
     choose-a-home page until the customer picks Netlify / paid hosting /
     ZIP-elsewhere; the site stays visible on our preview subdomain
     (our brand). Reason: bounds the free-rider commitment to zero, and
     with a FREE door beside it the page converts rather than punishes.
  3. **Netlify connect happens IN THE REQUEST PHASE** — on the
     request-confirmation page + email, i.e. the build-wait window
     ("usually ready the next day"), so the hosting choice is a non-event
     by delivery: most sites are BORN into the customer's own account.
     NEVER on the request form itself (top-of-funnel stays zero-friction);
     always SKIPPABLE (a stalled signup never blocks a build or sale; the
     link repeats in approval + delivery emails); tokens for
     never-converting requests expire after N days.
  4. **Account surface** — v1 IS the delivery email (ZIP, connect link,
     hosting payment link, domain subscription link, Stripe's hosted
     customer portal); Phase 6's editor login home later becomes the
     account hub (Edit / Domain / Hosting / Billing). No standalone page.
  5. **Own authoritative DNS: GO — as part of the DOMAIN programme**, not
     the scale review. Reason: every domain we register needs DNS wherever
     the site lives (ours / their Netlify / choose-a-home), so DNS is the
     domain product's backbone, and owning it keeps every state a zone-file
     template + keeps the emergency exit ours. Early domains may ride
     zone-per-domain; migration later = one proven EPP call each (VMB-015).
     Our-hosted sites then take CF for SaaS custom hostnames for TLS/CDN.
- **Facts that reshaped the advice** (measured this session, cited in the
  PLAN snapshot): Netlify-via-customer-OAuth was already in the 08-14
  matrix as the real-ownership row; no provider has an account-creation API
  and temp-email account minting is the bulk-suspension shape; EPP login
  already PROVEN from the cluster + a deployed EPP client exists + the
  second customer TAG is already applied for (`domain:create` is the only
  missing verb); `mode=payment` is HARDCODED in the billing provider (£10/mo
  needs Stripe Payment Links or the PAY-005 methods; never trust the
  PAY-007 scaffold); forms are already ruled mailto-with-paid-tier-later.
- **TODO/backlog recorded at the owner's request:**
  (a) **Domain-per-site programme** — framework-chosen .uk name,
      `domain:create` on the proven EPP client under the SECOND tag (gated
      on Nominet granting it), own-DNS pair with three zone templates,
      choose-a-home page, £10/mo retention Payment Link. Also gated on
      Stripe keys + webhook edge exception.
  (b) **Whole-architecture scale review incl. own cluster(s)** — AFTER the
      working site. Seeded agenda: CF zone cap (unpublished), CF-for-SaaS
      vs zones, runner-per-repo, monorepo build throughput, pod-per-tick
      economics, cluster capacity, per-hostname metering (Workers Analytics
      Engine; TRF-006 beacon pattern), abuse/takedown posture.
  (c) **Busy-site payment thread** — metering threshold → upsell email
      offering BOTH our paid tier AND free Netlify-connect (the pressure
      valve: the costliest customers are invited to take hosting home).
  (d) **DEFERRED by owner, do not build**: newsfeeds/editorial pieces as a
      paid-hosting perk.
- Superseded from the 08-14 PLAN: Part 2d's "ownership reverts to the ZIP"
  — ownership is now **domain + their-own-Netlify + ZIP**. Phases 3–6
  mechanics otherwise stand; Phase 4 gains the email links; new Phase 4b =
  netlify-oauth publisher backend (one const + one case on the proven seam;
  needs a new Deps field — Deps carries only ObjectStore today).

## 2026-08-17 (later session) — PHASE 3 BUILT + COMMITTED: zip_deliverable (commit e1a7f1935, register DGH-011)

- **Falsifiers checked before writing a line**: no newer handoff; `archive/zip`
  absent from the repo (nobody had started); canary intact (`noted.co.uk`,
  `published_hash` still the 08-16 16:01:40Z value). The "webdesign.uk live
  webdesign" session was not in the peer list, but its owner-terms work landed
  as commit `84202f061` (payment before build; preview ~1 month beside the
  PERMANENT ZIP) — context for Phase 4's email copy, no conflict with this
  phase.
- **The handoff's open dispatch question, decided: own agent type, not
  site-publisher's workflow.** Evidence: one type = one workflow (spawned
  agents run their definition's workflow; the group-spawn `workflow` override
  does not apply to single spawns, spawn_actions.go:1592 vs :2259), and
  coupling the cut to the hourly reconciler would produce ~24 unread archives
  a day. So: `zip-deliverer` on `isStorageEnabledAgent` (the sanctioned
  per-type grant) + a `zip-deliverable-dispatch` spawn→call shim, seeded by
  **459_zip_deliverer_agent_HOLD.sql** — deliberately NO scheduled task, and
  the seed's verify block asserts zero schedules target the new types.
  Topics need no Go change: the spawn→call handshake creates per-JOB topics
  at spawn time (spawn_actions.go:1180–1210); 422's `publish-reconciler` (a
  seeded new type) already proved the path.
- **The action** (`zip_deliverable_action.go`): reuses `publish.S3Source` +
  the `newPortfolioStore` idiom; composes to a TEMP FILE (seekable `*os.File`
  → SDK sends Content-Length; the b2worker whole-buffer pattern explicitly
  not copied — its own comment forbids it for this size class); verifies at
  the artefact BOTH sides of the upload (entries == listing count; archived
  `index.html` byte-equal to a fresh origin read; remote size == local, read
  from the destination LISTING, never the upload return); key =
  `deliverables/<domain>/<domain>-<treehash12>.zip` so a re-cut of the same
  tree overwrites its equivalent; oversize ALERTS and completes (threshold is
  an input so an induced oversize can prove the alert fires). 3 unit tests
  green, incl. the induced-oversize demand control; the package fakes refuse
  non-seekable bodies, so the wrong upload shape fails in test.
- **Seed subtlety worth keeping**: the workflow config maps
  `size_alert_bytes`/`expiry_minutes` as EXPLICIT dot-paths because a field
  with a spec default can only be overridden by a Strategy-0 path that
  resolves (bugs_open/248 finding (b), action_inputs.go); an unresolved path
  leaves the default standing (verified at Strategy 0's `value != nil` guard).
- **RFC_022 worked as designed**: the parity test FAILED on first run
  (`check.py OPTIONAL_KEY_COUNTS["zip_deliverable"] = 0, registry declares 3`),
  literal regenerated with the command in check.py's own comment (120 → 121
  lines, only our action added), all four BudgetCron tests green, and the
  kustomize overlay re-applied post-commit (new configmap
  `optional-key-budget-check-script-t5849k8k98`, cronjob reconfigured).
  Reader-side proof arrives with tomorrow's 06:50Z run's `doc_notes` row —
  remember a MISSING row means the job did not run, not "clean".
- **Same-file passenger, recorded not lamented**: the `000_concept_index.md`
  DGH-011 row reached HEAD on ANOTHER session's commit (`8e051f16d`) between
  my edit and my commit — the known shared-tree shape; forward-only holds,
  the row is correct, and e1a7f1935's message says so.
- **Council**: submitted alongside the commit, corr
  `4cc887b9-6c4a-4165-ae21-6c69bbefccfd`, `Council-Submitted:` trailer on
  e1a7f1935. Verdict may land ~30 min after submission (dispatch queues
  behind the fleet) — READ it; a REVISE must be acted on, the code is already
  on the shared branch.
- **HEAD proven, not assumed**: `git archive HEAD` → clean build of
  `./platform/... ./cmd/...` + Zip tests green from the archive.
- **What Phase 3 still owes (post-roll, next session or later this one)**:
  (1) verify the running chassis carries e1a7f1935 (ancestry vs the stamp,
  with controls — recipe in 459's header); (2) apply seed 459; (3) production
  acceptance on the canary: dispatch the shim (PUBLISH_OK recipe in the seed
  header), `unzip -l` count == 8 B2 objects, `index.html` sha256 == origin,
  presigned 200 in-expiry AND 403 after expiry (both directions), and the
  size-alert demand control (`size_alert_bytes: 1` dispatch → alert fires,
  cut completes).
- **Council verdict READ: APPROVED round 1** (corr `4cc887b9`, all seats, 2
  advisory objections, none high). Objection outcomes: (a) *input_contract
  undeclared on zip-deliverer* — ACTED ON: the seed now declares it and the
  verify block asserts it; **chasing this caught a real defect the council did
  not name** — in call_agent's `input_mapping` an UNMARKED field whose source
  path does not resolve ERRORS the step (ResolveInputMapping — a DIFFERENT
  rule from the executor step config's Strategy 0, which silently skips), so
  the shim's `size_alert_bytes`/`expiry_minutes` mappings would have failed
  every plain `{domain}` dispatch; both now carry the `?` optional marker.
  (b) *seed omits description column* — REFUTED: a sketch-elision artefact
  (the `...` in the council sketch), the committed file always carried the
  column; no change. (c) *459 collision risk* [low] — apply-time re-list
  instruction added to the seed header. The precedent rows (site-publisher,
  publish-reconciler) both carry NULL input_contract — the new row is now
  stricter than the 422 pair it mirrors, deliberately.
- **Landmine filed** for the `?`-marker trap (call_agent input_mapping vs
  executor step config — same dotted path, opposite failure rules), verifier
  ran STILL_VALID at 22:22Z. The entry itself reached HEAD as a same-file
  passenger on `d157f6714` (another session's LANDMINES correction) — the
  second passenger event of this session, one in each direction; recorded,
  nothing lost. Also: a first grep for the entry reported a FALSE ABSENCE
  (case-sensitive grep vs the heading's capitals) — trust `git log -S`, not
  one grep spelling.
