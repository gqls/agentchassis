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

## 2026-08-18 — ROLL LANDED (v1.0.1308): deploy proven, seed 459 APPLIED, acceptance dispatched

- **Deploy proven at the artefact, not assumed**: startup provenance line
  already scrolled from both pods (2h of chatter — expected, it is a startup
  line), so the image-LABEL recipe was used instead:
  `docker inspect ...:v1.0.1308 → org.opencontainers.image.revision =
  e7e5e4d53`. Ancestry: `git merge-base --is-ancestor e1a7f1935 e7e5e4d53`
  exit 0 (and `5a2faf018` also in); discrimination control: first post-stamp
  commit (`72c5d8a3b`) correctly NOT an ancestor; same-tag trap excluded by
  digest equality — pod imageID `sha256:32e10fa2…` == local labelled image.
- **Seed 459 applied ~09:52Z** (re-listed first — no collision): both INSERTs
  1 row, verify block passed (action name, input_contract assertion, 3 shim
  steps, zero schedules). STATUS line updated in the header per the sidecar
  convention (no runner record exists for UPPERCASE sidecars, by design; 422
  did the same — and sidecars are outside the migration-checksum trap).
- **Acceptance dispatch fired ~09:55Z** with the PUBLISH_OK recipe:
  corr `c9ee1ca8-a7bc-4168-bb29-527fd3c5e49a`, input `{domain: noted.co.uk}`.
  Watcher armed; a missing row is latency (budget 30 min), and a FAILED
  spawn→call handshake is re-dispatched, never cancelled pre-diagnosis.
- **ACCEPTANCE COMPLETE ~09:55Z — every check at the artefact, and PHASE 3 IS
  DONE.** Cut 1 (corr `c9ee1ca8`, input `{domain}` ONLY — which live-proves
  the `?`-marker fix, since the pre-fix mapping would have errored this exact
  dispatch): COMPLETED first try, result
  `{files: 8, zip_size_bytes: 18996, total_source_bytes: 60466, tree_hash:
  th1:05a06351…}` — the tree hash EQUALS the canary's stored published_hash.
  Verified independently of the action's self-checks: presigned GET → 200,
  18,996 bytes (== result); `unzip -l` → **8 files, 60,466 bytes** (== the 8
  B2 objects, == result); extracted `index.html` sha256 `b4416c32…` ==
  served `noted.ugg2.com/index.html` fetched cache-busted in the same minute
  == the origin value recorded in NOTES on 08-16 BEFORE any publish existed
  (three independent sources). Cut 2 (corr `8e3ad6e7`, input
  `{expiry_minutes: 1, size_alert_bytes: 1}` — both optional mappings
  exercised RESOLVING this time): COMPLETED in ~48s, `size_alert=true` AND
  `zipped=true` — the demand control: induced 1-byte threshold vs 60,466
  source bytes alerted and still completed, never truncated. Same zip_key as
  cut 1 (same tree ⇒ same key: overwrite-not-accumulate proven live).
- **Expiry proven BOTH directions, with one provider nuance**: in-expiry =
  200 (cut 1's 7-day URL, fetched twice). After expiry: **B2's S3 gateway
  returns 401 `UnauthorizedAccess`, NOT AWS's 403** — body says `Request has
  expired given timestamp: 20260818T095232Z and expiration: 60`, and the
  7-day URL 200'd in the SAME BREATH (the control that separates a real
  expiry refusal from a mangled URL). The PLAN's "403 after expiry" is
  satisfied in substance; any future check that asserts the literal 403
  against B2 will fail wrongly — assert "non-200 with the expiry message",
  and DGH-011 now says so.
- Nothing was written on any box: all verification via presigned HTTP + the
  served site + local scratch. Spawned pods: 2 zip-deliverer pods total, one
  per cut, on-demand only — no schedule exists (asserted by the seed verify).

## 2026-08-18 (~12:20Z) — delivery-architecture addition from the owner (recorded here because Phase 4 consumes it)

- **The domain fee now has a BUY-OUT**: owner directive (chat, 2026-08-18,
  while directing webdesign.uk FAQ copy): domain rental £10/month via a
  subscription link carried IN THE DELIVERY EMAIL (confirms Phase 4's email
  is the vehicle), OR a one-off £200 purchase after which the customer is
  free to transfer the domain to their own registrar or host. The 08-17
  architecture PLAN's decision 1 knew only the £10/mo retention; the £200
  transfer-out path is new and belongs in Phase 4's email links and the
  domain programme's Stripe products. Recorded in webdesign.uk's register
  as facts domain_rent_monthly (value 10) and domain_buy_once (value 200).

## 2026-08-19 — the seeded "whole-architecture scale review" is being executed elsewhere

The scale review this lane seeded (this file, 2026-08-17, "(b) whole-architecture scale
review incl. own cluster(s) — AFTER the working site") has been brought forward at the
owner's direct request and is being executed by the `dispatch_throughput/` workstream —
deliverable `dispatch_throughput/RESEARCH_2026-08-18_throughput_to_thousands_of_domains.md`.
Its seeded agenda items (CF zone cap, CF-for-SaaS vs zones, runner-per-repo, monorepo build
throughput, pod-per-tick economics, cluster capacity, per-hostname metering, abuse posture)
are all carried there. One addition that concerns THIS lane directly: the owner now expects
promotion-driven onboarding BURSTS (many domains/day), which makes the request-phase build
wait UX and a priority lane for paying builds first-class inputs to Phase 4+ — see the
research doc's burst section and decisions D0b/D2/D15. Nothing in this lane's Phase 4 plan
is changed by that work; this is a pointer, not a handoff.

### 2026-08-20 — Phase 4 STATE is built: the stamps, and ONE token shape for every customer link

Register entry **DGH-014**. Migration **511** applied and recorded; `platform/delivery`
in the build. Council submitted, `Council-Submitted: 905d9078-86c2-47a7-af0a-781723a46c08`.

**What made this a token table rather than a config change.** The owner ruled the ZIP
download link should last "the longest time we have, which I think is 6 weeks". It
cannot: a presigned URL is capped by the SigV4 protocol at 604,800 s, **the cap is
enforced by the object store and NOT by the SDK** (`aws-sdk-go-v2 v1.25.1`'s
`aws/signer/v4` and `service/s3 v1.51.0` both sign any duration and hand back a
well-formed URL), so a six-week link mints cleanly, the action reports success, and it
fails only in the customer's browser — as `SignatureDoesNotMatch`, which reads as
broken credentials. `[MEASURED 2026-08-20]`, key deliberately absent so the status is
the whole answer: **`604800` → HTTP 404 `NoSuchKey`** (the control — signature
accepted), `604801` → 403, `3628800` → 403. Exact to the second.

So the customer holds **our** token for six weeks and each click mints a fresh short
presign server-side. The six weeks lives in ONE constant (`LiveLinkWindow`) instead of
being a number two systems have to agree on, and `MaxPresignWindow` + `PresignWindowFor`
make the ceiling impossible to overrun by accident.

**Schema:** `sites.handed_over_at`, `sites.live_link_expires_at`,
`sites.transfer_confirmed_at`, plus `customer_access_tokens` (hashed, expiring,
optionally single-use, `ON DELETE CASCADE`). **One table, not one per link** — the
producer set is named (`zip_download`, `confirm_transfer`, and Phase 5's
`editor_session`) and `purpose` is a CLOSED `CHECK`, so a fourth customer link costs a
migration and stays visible in the ledger. That is the 2026-08-02 §1 condition met
explicitly rather than assumed.

**The plaintext is never stored** — sha256 hex only — so a leaked database yields no
working links, and there is no resend-the-same-link path: re-issuing mints a new token,
which is the right behaviour anyway.

#### How it was verified, and the one thing the unit tests cannot do

**sqlmock asserts the SQL this package SENDS; it cannot tell you what Postgres DOES
with it.** Every property that actually protects a customer here is SQL-level. So the
Go tests cover Go-level behaviour only — and say so at the top of the file — and the
semantics were run against the real database inside a rolled-back transaction:
**10 checks, all passed** (idempotent stamping; the closed purpose vocabulary; purpose
isolation — a download token must not redeem as a confirmation; single-use spending
once and missing on the second click; a download token NOT being single-use; expiry
evaluated at redemption time; revocation; the confirmation date not moving on a second
click; hash uniqueness estate-wide; cascade delete leaving no orphaned links).

**Then the harness was mutation-proved twice, and one mutation taught me something:**

| mutation | what caught it |
|---|---|
| drop the `used_at IS NULL` predicate from the confirm redemption | **the CHECK constraint fired BEFORE my own assertion did** |
| turn the idempotent `COALESCE` stamp into a plain assignment | `FAIL 1c: a second stamp MOVED the handover date` |

The first is the interesting one. The `single_use → use_count <= 1` CHECK was written as
a backstop "in case some future writer forgets the predicate", which is the kind of
claim that normally sits in a comment unexercised for ever. Removing the predicate
proved it: the backstop is **measured**, not asserted. It is also why the predicate and
not the CHECK is the primary control — the CHECK alone would turn an ordinary second
click into a constraint error instead of a clean miss a handler can report.

The migration's own guard **induces** both CHECK constraints rather than asserting they
exist, and that was mutation-proved too: a `purpose` CHECK that exists, is correctly
NAMED, and permits anything is caught (`511: the purpose CHECK did not refuse an
unknown purpose`). A constraint that has never refused anything is untested.

#### Applied by hand, deliberately, and recorded

`--apply` takes **every** pending file, and the pending list is other lanes' work
(several probes come back "inconclusive — the live config has drifted"). So 511 went in
by hand and was then registered with `--record-only` plus a note saying why. Do the same
for the next one.

#### Answering the commit-msg architecture signal

It fired: *"migration + platform code in one commit — needs a staged rollout order."*
The three RFC_022 conditions hold — opt-in, the **safe** side is the default (an
unstamped site is one the editor gate refuses, so a missed stamp closes a door rather
than opening one), and **zero live consumers, enumerated across the whole repo rather
than asserted**: every hit for all four new names is inside this change.

And the ordering worry has nothing to bite on **yet**: the schema is purely additive so
an old binary ignores it, and the new code has no caller so a new binary would never
execute a statement naming these columns. **That changes the moment the HTTP surface
lands** — at which point 511 must already be applied, and it is.

#### What is NOT built, stated in three places so it cannot be inferred away

Nothing enforces `live_link_expires_at` (nothing takes delivered sites down today);
nothing mints or redeems these tokens in production; there is no HTTP surface and no
delivery email; no chase email. **The helper has no live caller.** Its only dependency
is the database and the tests exercise that, which is why landing it uncalled is
defensible — but the deployment contract is unverified until something real calls it,
and the register entry says so in those words.

**Next, in order:** the HTTP surface (`/d/<token>` minting a clamped presign and
redirecting; `/c/<token>` recording the confirmation), then the delivery email through
`platform/mailer`, then the weekly chase, then the retraction job that gives
`live_link_expires_at` teeth.

### 2026-08-20, later — the council said REVISE, and the round was my fault twice over

Verdict on `905d9078-86c2-47a7-af0a-781723a46c08`: **revise**, decided by a gating
objection from `editquality`. **Five of ten seats raised the SAME point** — editquality
(gating), `reuse_agent`, `guardian`, `constitution`, `prior_art_librarian` — while four
approved on their own grounds (`bug_historian`, `render_guardian`, `debug_historian`,
`mission`, `architecture`).

The objection: *"the Schema already lists `sites.handed_over_at`,
`live_link_expires_at` and `transfer_confirmed_at` as existing columns, so either the
`ALTER TABLE` fails on apply, or they pre-existed and the premise is false."*

**They were right about what they could see, and the cause is my sequencing.** I applied
511 at **17:24:07Z** and submitted at **~17:30Z**. The council prompt's Schema section is
generated from the **live** database, so the reviewers were shown my own six-minute-old
change back as pre-existing state, with nothing to date it. From inside a seat those two
readings are the only ones available and both are sound inferences. `reuse_agent` filed
it as a STEP ZERO miss — *"no evidence the author ran `\d sites`"* — and on the evidence
in front of it, that was the correct call.

**Two real defects, both in the submission, neither in the migration:**

1. **I ran the schema check and never showed it.** The information_schema query went in
   early in the session and returned nothing; the migration's own opening guard raises
   `511: already applied` on either the column or the table existing, and it **passed on
   the real apply** — which is only possible if neither existed. That IS the check, and
   it was compiled into the file. It was not in `grounded_in`, so it did not exist as far
   as the council was concerned. **A check you ran but did not cite is a check you did
   not run.**
2. **My sketch elided the guard the reviewers then asked for.** editquality's suggested
   fix was `IF NOT EXISTS` — protection the file already had in a stronger form. **DGH-011's
   register entry records this exact class** ("a sketch-elision artefact, the file always
   carried the column"), so this lane has now paid for it twice. Sketch the guards
   verbatim and abbreviate the boring part instead.

**What I did NOT do: take the suggested fix.** `IF NOT EXISTS` would satisfy the
objection by making a double apply a silent no-op, which on a shared tree is worse than
the loud refusal it would replace. Resubmitted with the guard in the sketch and an
invitation to overrule me if the seats still want it having seen it.

**The measurement that closes the 'dormant machinery' reading** — `guardian` and
`prior_art_librarian` both asked whether something already writes these columns with
other semantics, and that cannot be settled by argument:

```
44 sites | handed_over_at populated on 0 | live_link_expires_at on 0 | transfer_confirmed_at on 0
customer_access_tokens: 0 rows
```

And the prior-art search the previous submission **asserted rather than ran** (upheld
objection): `information_schema.tables` filtered on `%token%`/`%magic%`/`%access%`/
`%customer%` returns exactly one row — this table. The estate's only other token table is
`auth_tokens`, in the auth-service's own database, which `PLAN_2026-08-14` already rules
out for customer space in as many words.

Round 2 submitted on the same trail correlation, so the commit's existing
`Council-Submitted:` trailer stays valid and `098` will credit it if this one approves.

**Filed in `LANDMINES.md`, because the trap is general and it feeds the other one:**
`--apply` takes every lane's pending file, so you apply yours by hand and `--record-only`
it — which is precisely what puts you in "applied before submitted". The two traps hand
you to each other. **A REVISE round is cheaper than the defect it finds, and this one
found two — they were just both in the submission.**

### 2026-08-20 — council APPROVED at round 3, and the approval's ADVISORY notes found one more defect

`905d9078-86c2-47a7-af0a-781723a46c08`, 18:06:02Z, **all reviewers approve**, 5 advisory
objections all `low`. Trailer `Council-Reviewed:` written on `fa3b665ed` after reading it.

**Three rounds, three findings, none of them in the state layer.** Worth recording as the
argument for submitting at all — the change under review was fine every time:

| round | what it caught | where the defect was |
|---|---|---|
| 1 (5 of 10 seats) | I applied the migration 6 min before submitting, so the council's live Schema section showed my own change back as pre-existing state; and my defence rested on a check I had run and cited nowhere | the **submission** |
| 2 (`editquality` high) | `applied_by='record-only'` cannot prove the file's SQL ran | the **evidence** |
| 2 (`reuse_agent`, `prior_art_librarian`) | **the presign clamp was in the wrong place** | the **design** |
| 3 (advisory, post-approval) | **my call-site census said SIX; it is NINE** | the **measurement** |

**Round 2's design finding is the one that paid for the whole exercise.** I had put
`MaxPresignWindow` in `platform/delivery` — the one package that could not yet get it wrong —
leaving every existing presign caller unprotected. The ceiling now lives in `platform/storage`
and is enforced inside both presigners, so every caller inherits it.

**Round 2's evidence finding was the sharpest.** I cited the `schema_migrations` row as proof the
file ran; the row says `applied_by='record-only'`, which is the estate's marker for *not run by
the runner*. The seat was right that it proves nothing. The answer that rests on nothing I
remember: the file leaves a fingerprint a bare `ALTER TABLE` cannot — four `COMMENT`s, two
**partial** indexes, three named CHECKs. All present. **Nobody hand-types that.**

**Round 3's advisory then caught the census.** `reuse_agent`: *"the plan's search was grep-based
and could miss a call site."* It did — three of them, in `thunder/data_url_actions.go`, invisible
because they pass a **variable** (`req.ExpiryMinutes`, a caller-supplied JSON field) rather than
one of the magic numbers I grepped for. **Enumerate by the interface, not the value.** And the
claim I actually wanted was one grep away and I never reached for it: `NewPresignClient` /
`PresignGetObject` / `PresignPutObject` match **only** inside `platform/storage/s3.go`, so there
is no bypass path at all. Full entry in `WRONG_CALLS.md`.

**The miss was hiding a defect the clamp itself introduced.** Thunder reports an `ExpiresAt`
computed from the same `expiry` variable it passes down, so clamping only inside the helper would
have made the reply **advertise a lifetime the URL did not have** — a fresh inconsistency from a
change whose entire purpose was to stop links promising what they cannot deliver. The three
thunder sites now clamp locally too. **That defect was unreachable while I believed thunder was
not a caller**, which is the real cost of an undercount: not the number, the blind spot.

**Two advisories deliberately NOT acted on, recorded in the register so they are not lost:**
- the clamp is **silent** (`S3Client` carries no logger, only its constructor takes one), so a
  caller that starts requesting above-ceiling *by design* would be invisible. Revisit with a
  metric hook then, not now;
- the clamp is **inert until each service rolls** — nine call sites across ~6 services, so a
  same-tag rebuild could leave an old unclamped binary serving. Verify per SERVICE at the
  artefact: `logs -l app=<service> | grep 'build provenance'`, then
  `git merge-base --is-ancestor 882622629 <stamp>`.

**Housekeeping note:** the commit-scope pattern check flagged one removed line from
`LANDMINES.md`. Verified before moving on — `git diff --numstat` says `2 1`, and the removed line
is my own bullet from this morning, replaced by its struck-through corrected form. Not another
session's entry.

## 2026-08-21 — owner ruling relayed from the scale review: DNS plan B starts NOW

In the dispatch_throughput scale-review discussion the owner ruled **"let's start on plan
B"** (own authoritative DNS + Cloudflare-for-SaaS custom hostnames) — the trigger is no
longer "near 500–1,000 domains": with promotion bursts of up to 50 domains/day expected,
the ~1k zone cap is ~3 weeks of promotion away, so plan B readiness should PRECEDE the
first big promotion. Execution belongs to this lane / the domain programme (own-DNS was
already ruled GO on 08-17; the new part is CF-for-SaaS + the calendar urgency). Also
relevant to Phase 4+: the owner wants a human review gate before each client site goes
out (mechanism in another thread), and clients-first dispatch priority is now ruled.

## 2026-08-25 — the delivery-only listener BUILT (RFC_054 Q2), and the defect it nearly shipped with

Picked the lane up cold. **No thread had it open** — no session carries its name — but it is
joint with webdesign.uk since 08-18 and four webdesign sessions were alive, three busy, so
the shopfront copy/design work is theirs. Two other lanes hold pieces: `web_admin_console`
built the second-click `/c/` page today and owns its close-out; `dispatch_throughput`
explicitly routed **DNS plan B** here on 08-21 ("do not build here; check they picked it
up") and nobody had.

**State re-measured rather than carried forward.** `4c996e1b5` is the running core-manager
(09:26Z); `fa3b665ed` (Phase 4 state) and `882622629` (presign clamp) are ancestors,
reversed control passes; `24b63120d`/`d1a4bdcdf` (second-click, committed 12:34/12:40Z) are
**not** — so it is built and not live, and this change lands on the same roll. Counters:
51 sites, `handed_over_at`/`transfer_confirmed_at`/`live_link_expires_at` all **0**,
`customer_access_tokens` **0** [MEASURED 2026-08-25].

**What was built:** `/c/` (and later `/d/`) moved off `core-manager:8088` onto a second
listener on `:8090` registering those routes and nothing else. Commit `d30917150`, register
**SYS-095**, council `Council-Submitted: 25cd3044-23e0-4902-9686-692a42779170`. Plan:
`PLAN_2026-08-25_delivery_only_listener.md`.

### The misstep worth recording: I nearly shipped an opt-in that nothing could switch on

The listener is opt-in via `server.delivery_port`, set **only** as
`SERVICE_SERVER_DELIVERY_PORT` in the production overlay and named in no service's YAML.
That combination does not work on this loader, and it fails **silently**: viper's
`Unmarshal` populates from viper's known key set, and `AutomaticEnv` alone does **not** add
a key absent from both the config file and the defaults. So the overlay would have read
perfectly correct, `DeliveryPort` would have been `""`, no listener would have started, and
every customer link would have 404'd at the box — with nothing wrong in the cluster and no
error anywhere to find.

I caught it only because I wrote the test **before** trusting the wiring, and it FAILED:

```
--- FAIL: TestDeliveryPortIsReadFromTheEnvironment
    Server.DeliveryPort = "", want "8090"
```

Fixed with an explicit `v.BindEnv("server.delivery_port")` in `platform/config/loader.go`.
**The general trap: any env-only config key on this loader has it**, and the symptom is
indistinguishable from "the key is set to empty". Filed to `LANDMINES.md`.

That is also the second time this lane has been reminded that a check only counts if it can
come out the other way. A `[MEASURED]` marker on "the overlay sets the port" would have been
true and worthless; the disconfirming result was the point.

### What the change actually buys — and what it does not, stated here because it is easy to overstate

Widening the box's `location` regex now reaches `/c/` and `/d/` and stops. But **8088 stays
on the WireGuard egress fence**, because the chat bot's facts relay
(`/api/v1/site-facts/:domain`) needs it and the fence's own comment records that removing it
"stops the bot STARTING, by its own design". So **the fence is not the containment**: a NEW
vhost written straight to `:8088` would still reach the whole admin API. Making the fence
the control means moving the facts relay onto a box-facing listener too — a change touching
a live customer-facing bot, and outside what the owner ruled. Recorded in the register entry,
in the fence's own comment, and in the vhost header, so the next reader inherits the residual
rather than the reassurance.

### Guards, mutation-proved (a passing test proves nothing until a mutation kills it)

| mutation | killed by |
|---|---|
| `assertNoDeliveryRoutes` returns nil early | the assert test **and** the main-router test |
| empty port defaults ON instead of OFF | `TestDeliveryListenerIsOptInAndDefaultsOff` |
| an extra `/health` added to the delivery engine | the exact-route-count test **and** the 404 sweep |
| delivery engine registers nothing at all | the count test **and** the reach control |

The fourth matters most: without `TestDeliveryEngineReachesTheHandler`, an engine serving
**nothing** would have passed every 404 assertion and read as perfect containment.

### Owed next, in order

1. **After the next core-manager roll:** ancestry with reversed control, then verify **from
   OUTSIDE** — the box's anchored regex means a cluster-internal curl proves nothing about
   the customer path (LANDMINES, footprint `links.webdesign.uk.nginx`). Containment check is
   `:8090/api/v1/admin/work-items` → 404 **paired with** `:8088` → not-404; the 404 alone
   passes just as well against a typo'd hostname.
2. **Then** the box vhost apply (owner box steps) — its header carries the
   DO-NOT-APPLY-BEFORE-THE-ROLL gate.
3. Read the council verdict and act on a REVISE; the code is already on the shared branch.

## 2026-08-25 (later) — Phase 4's CLAIM built, and DNS plan B picked up and scoped

Continuing the same session. Three pieces taken in order at the owner's direction:
the delivery-only listener (above), the delivery email, DNS plan B.

### The listener's council round 2 — both gating objections were changes I HAD made

`25cd3044` came back **REVISE**, decided by a `guardian` HIGH. It asked whether
`cmd/core-manager/main.go` actually calls `StartDelivery`, warning that otherwise *"the
listener struct is built, the config is read, the tests pass, and no delivery traffic is
ever served in production"*. `editquality` raised the same shape about the box vhost.

**Both files were already in the commit.** I omitted them from an 8-edit plan and my
rationale then asserted them as done. That is this lane's round-1 lesson repeating —
*a check you ran but did not cite is a check you did not run* — except this time it was a
**change I made but did not show**. And when I first re-cut the plan to restore them, the
cap silently pushed out the **opt-in** edit instead, which is the same failure a third
time. Fixed by merging the three same-file `server.go` edits into one.

**One real code defect came out of it** (`guardian`, medium): `assertNoDeliveryRoutes`
compared paths only, so a root catch-all would have served `/c/` by prefix dispatch while
the route table held no `/c/` entry — the guard would have called the admin port clean
while it answered customer links. It now refuses a wildcard whose static prefix sits at or
above a delivery path, and **discriminates** rather than banning wildcards (`/api/*path`
cannot reach `/c/` and is still allowed, asserted by a control). Mutation-proved.

Two objections answered by running the check rather than arguing:
- `reuse_agent` was **right that prior art existed** and I had not looked. `platform/health/server.go`
  IS a shared second-listener helper, and five adapters hand-roll a `healthServer`. It does
  not fit, structurally: gorilla/mux (the handler takes a `gin.IRouter`), it **always**
  mounts `/health` and `/ready` — which breaks the delivery-only property my own mutation
  proves the tests catch — it retains no `*http.Server` so it cannot drain a customer
  mid-click, and `Start()` also launches a metrics listener. The house convention for
  route-table assertions exists as **tests** (`internal/tools-api/api/server_test.go`,
  including an opt-in "registers no routes when config is nil" assertion) and this change
  matches it without having known.
- `prior_art_librarian` flagged my `ea99befa` citation as possibly fabricated. It is real:
  approved 2026-08-25 12:36:14Z, seat `guardian`, severity low, edit 5, and the objection
  text is what I said it was.

`architecture`'s residual is now **RFC_054 §6**, status OPEN and unowned.

### Phase 4: the CLAIM (`platform/delivery/prepare.go`, `6fee5b7ce`, register DGH-017)

Built the delivery email's **precondition layer** and deliberately not its copy or its send.
The reason is dated: the owner's copy brief landed in the register **today at 14:51:36Z**
and two of his product questions are undecided, so wording written now gets rewritten —
while the gate, the claim and the links are settled by rulings. Two of the email's six
links also cannot exist yet (`/d/` unbuilt, no Stripe keys), so a "finished" email today
would be missing a third of its content.

The ordering is the whole safety argument: review first → **the stamp IS the claim**
(idempotent, `AlreadyHandedOver` treated as a refusal) → mint last. At-most-once therefore
belongs to the stamp, not the send, so no retry can email a customer twice. The accepted
cost is a stamp outliving a failed send — chosen because that fails **loud** (the work item
errors, a human sees it) where send-then-stamp duplicates **silently**.

**The mutation that PASSED, which is the finding.** Four guards, four mutations; the fourth
— deleting `validate()` — passed. With validation gone, `Claim` ran on to the gate query,
sqlmock refused a call it was never told to expect, and the test saw *an* error and was
satisfied. `ExpectationsWereMet()` does not catch it either: there were no expectations to
meet. **A bare `err != nil` could not distinguish a config refusal from a database
failure.** The test now identifies the error AS the config refusal, and the mutation kills
it. Exactly the class this file's own header warns about, found in my own test.

### DNS plan B (`PLAN_2026-08-25_dns_plan_b.md`, `2321d2542`)

`dispatch_throughput` routed this here on 08-21 — *"do not build here; check they picked it
up"* — and nobody had, for four days.

**Measured today rather than carried forward:** **41** Cloudflare zones (39 on 08-18, so
~2/week), **41 of 41 on the Free tier**, 51 sites all `build_status='pending'`. At 2/week
the cap is ~9 years out; at the promotion scenario's 50/day it is ~19 days out. So the
trigger is **"promotion is scheduled"**, not a date — and promotion cannot start while the
shopfront is parked.

**⚠ The finding: the cap is UNVERIFIED.** Every statement of "~1,000 zones" in this estate
traces to the single 2026-08-18 research doc. I could not verify it — the Cloudflare token
is **zone-scoped**, so `/accounts` returns nothing and the account's own limits are
unreadable with it. It is the entire justification for the programme's timing. Same for
whether CF for SaaS needs a paid tier. **Both are prerequisites, not formalities**, and
both are one dashboard look for a human.

Nothing built, deliberately, and §7 of the plan says why: unverified premise, a nameserver
shape that is architecture-scope by this repo's own test (blast radius = every domain we
manage, including ones we do not host), and a trigger that is not met.

### Owed next

1. Read both verdicts (`25cd3044` round 2; `0b84970d` the claim) and act on a REVISE.
2. After the next core-manager roll: ancestry + reversed control, then verify **from
   outside**; then the box vhost apply.
3. The `needs_delivery_review` **producer** — and it must go through `actions.writeWorkItem`.
4. Owner: the two DNS commercial facts (§3a, §4 D-e), then D-a/D-b.

## 2026-08-26 — design rotation back ON (peer heads-up, verified); my gate is immune by construction; label risk relayed

`webdesign-tool-rebuilds` messaged: `site-discovery-rotation-design` re-enabled after 15
days off (`bugs_open/401`). **Verified at the live rows rather than relayed:** both it and
`detected-item-promoter` are `enabled=t`, last triggered 09:20:36Z / 09:13Z today.

Impact on this lane: **none on the shipped work** — `Reviewed()` filters on
`item_type='needs_delivery_review'` (nothing produces it), and the safety counters live in
tables design traffic never touches. The one real cross-lane consequence: webdesign.uk is
in the rotation's reach (19 historical design-adjacent items) and an unprompted design
repair on `index` would strip the hand-placed "Not active yet" label at the parked preview
— probability moved from ~zero to live this morning. Relayed with the cheap standing check:
`NOTE_2026-08-26_design_rotation_resumed_label_strip_risk.md` in the webdesign dir.

## 2026-08-26 — verdicts read; Fable adversarial review found five real defects; all fixed or tracked; both trails resubmitted

**Verdicts from last night's resubmissions:** the claim layer **APPROVED** (0b84970d,
21:21); the listener **REVISE a third time** (25cd3044, 21:25) — and round 3's gating
objections were both MY SUBMISSION again: I appended the round-3 `newDeliveryEngine`
sketch after the round-2 one (two conflicting signatures side by side), and the wiring
test my argument leaned on never appeared in the test edit. Third occurrence of the
submission-not-reflecting-code class. Round 4's sketches are rewritten wholesale from the
committed files, never accumulated.

**The user asked for a fresh look "using fable"** — an adversarial reviewer with no
session context, read-only, briefed to find what the council and I both missed. It did.
No fatal findings; five real-but-bounded, three of which refute things I told the council:

| # | finding | disposition |
|---|---|---|
| 1 | `AlreadyHandedOver = (handed_over_at <> $2)` **double-delivers** on a same-microsecond collision or ANY reused-`now` retry (pgx = microsecond precision) | **FIXED** `d547b54cc`: the claim is now `WHERE handed_over_at IS NULL` — at most one winner by construction. **Proven on real Postgres rolled back, with the control**: second same-T claim → `UPDATE 0`; the OLD discriminator on the same state → "NOT already → would DOUBLE-DELIVER" |
| 2 | The filing contract was half-stated: `HandleApproveWorkItem` ALSO 400s without `spec.checkpoint=true`, and its error steers the owner to RESOLVE — writing the key the gate ignores. Silent stall, every press "succeeds" | **FIXED**: `ReviewItemRequiredSpec()` pins the spec half; test asserts the flag and map freshness |
| 3 | The pod stays **Ready with a dead customer door**: both `cancel()` calls were dead code (main blocked on bare `<-sigCh>`), and my round-2 answer to guardian's HIGH asserted the opposite | **FIXED** `67f1794f9`: shutdown is `select{sigCh, ctx.Done()}` — both cancels real, incl. the main listener's pre-existing dead one. The false claim corrected on the council record |
| 4 | `assertRoutesAreBoxServable` **vouched for `/d/`** while the vhost has no `/d/` block — the exact invisible-404 failure it exists to prevent, deferred | **FIXED**: `boxServablePrefixes` (= /c/) split from `deliveryRoutePrefixes`; registering `/d/:token` now refuses startup until the vhost block lands in the same commit. The /d/ test's expectation FLIPPED |
| 5 | My "approve path never exercised" claim was **false by the archive mistake the gate itself corrects** — one archived `approved_by` row exists (2026-03-17, created_by='test') | **CORRECTED** in the doc comment + WRONG_CALLS (the census-reused-for-a-second-question mechanism) |

**Deliberately NOT patched, recorded in DGH-017:** the unguarded reopen
(`UpdateWorkItemStatusAction` non-complete branch merges `result` forward — approval
survives; gate has no revocation) and the admin one-PATCH re-type. Shared machinery;
patching them in this lane's round is the `bugs_closed/124` bundling.

**Mutations:** 2 new (revert the WHERE → 3 tests fail via pinned regex; regain /d/ →
flipped test fails), full suite + `verify-head-builds` green at HEAD `1e37f282a`.

**Resubmitted:** listener round 4 (25cd3044) with wholesale-rewritten sketches + both
listener fixes; claim round 3 (0b84970d) as post-approval hardening so the trail reflects
what shipped. Also this morning: design rotation re-enable verified at the live rows,
label-strip risk relayed to the webdesign lane; the six-listeners observation gained its
trigger (an eighth forces an RFC).

## 2026-08-26 (midday) — THE ROLL LANDED: listener LIVE and verified at the pod; the box apply is now the owed half

A core-manager roll at ~11:55:08Z picked up everything: pod stamp `e7f1045fd`, and ALL of
`24b63120d` (second-click), `d1a4bdcdf`, `d30917150` (listener), `67f1794f9` +
`d547b54cc` (this morning's Fable fixes) are ancestors — control (`fa3b665ed`) passes.

**Verified at the pod, 2026-08-26 midday** (in-cluster half of the plan's table; the
outside half stays owed):
- opt-in present (`SERVICE_SERVER_DELIVERY_PORT=8090` in pod env) and the listener's own
  startup lines at 11:55:08Z ("Delivery listener configured/Starting … customer routes only");
- **the containment triple**: `:8090 /api/v1/admin/work-items` → **404** ·
  `:8088` same path → **401** (the not-404 control — proves the 404 is the listener, not a
  dead probe) · `:8090 /c/<43-char token shape>` → **200** (the reach control — proves the
  404s are not a dead port). That is checks #5/#6 + reach from `PLAN_2026-08-25`.

**The state to not misread:** from OUTSIDE, `/c/` now 404s at the box — the box still
proxies `:8088`, where the routes no longer are. This is the plan's predicted zero-impact
window (`customer_access_tokens` 0, handovers 0/0, re-measured today), and the vhost's
DO-NOT-APPLY-BEFORE-THE-ROLL gate is **now satisfied**: the box apply is UNBLOCKED and
owed at the owner's next box step. After it: the outside verification table, then the
second-click page's outside check, which is what releases the delivery-email block.

Relayed all of the above (dated) to the webdesign.uk lane for their whole-launch owner
status, including the design-rotation label risk and the two bounded admin-only residuals.

## 2026-08-26 (afternoon) — BOX APPLIED; OUTSIDE TABLE COMPLETE; SYS-095 fully live end-to-end

The webdesign.uk lane applied the repointed vhost (~13:2xZ; old vhost backed up to
`/root/links.webdesign.uk.nginx.bak-2026-08-26`, `nginx -t` clean, reloaded) and measured
the outside table: GET `/c/<43-char shape>` **200** (1,021 bytes, the render-only page,
from the internet) · POST same path **200** · `/c/x` **404** · `/other` **404** · preview
200 · admin 302 · apex parked-302.

**I ran the control their table didn't name** — the landmine's exact case: POST and GET
`https://links.webdesign.uk/c/<43>/confirm` both → **404** [MEASURED 2026-08-26 from
outside]. With theirs, that completes `PLAN_2026-08-25`'s outside table verbatim.

**Their one question — POST 200 on a nonexistent token — is the DESIGNED answer**
(`delivery.go:182-188`): 200 on success and on every failure, one undifferentiated page,
because the status being identical across causes IS the no-oracle property. Confirmed to
them with the instruction not to "fix" it.

**So SYS-095 is LIVE END-TO-END as of 2026-08-26 afternoon:** rolled ~11:55Z →
pod-verified (containment triple 404/401/200) → box applied ~13:2x → outside-verified
(their table + my suffix control). The second-click page is outside-verified, and the
delivery-email gate of DECISION_2026-08-24 is met on that front.

**Delivery email's remaining gates, stated so nobody re-derives them:** the
`needs_delivery_review` producer (unbuilt; must file via `actions.writeWorkItem`, at
`needs_human_review`, with `ReviewItemRequiredSpec()`'s `checkpoint:true`); the copy +
send (deliberately unwritten pending the owner's two open product questions); the owner
review itself per DECISION_2026-08-21e. Nothing else on this lane blocks it.

## 2026-08-26 (late afternoon) — OWNER RULES BOTH OPEN PRODUCT QUESTIONS; mail diagnosis; ZIP state answered

**OWNER RULING (decision 1, verbatim intent):** *"I give myself the opportunity to edit the
site, they don't have to know. As far as they are concerned it is one-shot with no approval
stage. But I need to be sure I'm not selling rubbish so I need this step. Three or four
days is fine."* — So: pre-send review EXTENDS to owner EDITS, internal only; the attested
customer position (`one_shot_no_approval`, "no changes are included") is UNCHANGED and no
copy or register fact moves; `build_duration` ("three or four days") absorbs the step. This
is exactly DECISION_2026-08-21e's design (internal gate, invisible to the customer; the
one risk is the gate leaking into copy, and the claims layer still cannot catch approval
language — that ban is still owed). The `needs_delivery_review` mechanism fits as built:
owner edits via the admin surface, then approves; a rejection routes to repair.

**OWNER RULING (decision 2):** *"For launch the customer doesn't get any edits. But I want
to build this voice edit next."* — So: launch ships with NO customer editing (attested
terms stand); **customer VOICE EDIT is the next build on this lane after launch** —
Phases 5–6 unfreeze POST-LAUNCH with voice as the input method. Both 08-25
reconsiderations are now RULED; the webdesign lane's SUMMARY_2026-08-25 resume point
applies and their launch sequencing unfreezes. Relayed to their session.

**Mail diagnosis (owner: "email not working... not from another email client"), all
MEASURED 2026-08-26 from outside:** contactforsales.com's DNS is NOT on Cloudflare
(NS = dns1/dns2.uk-noc.com — the hosting provider; our CF account has 0 matching zones),
so there is nothing to fix in Cloudflare. Inbound WORKS: MX = mx1/mx2.email-cluster.com
(Proxmox Mail Gateway), and an SMTP RCPT probe with a resolvable sender got **250 Ok for
webdesign@contactforsales.com on both** — external mail is accepted. Client ports ALL
OPEN from the internet (465/587/993/995/2080), the TLS certificate is valid for
mail.contactforsales.com, and 465 offers AUTH PLAIN LOGIN over implicit SSL. **So the
server side is healthy end-to-end; the failure is in the connecting client's settings or
credentials** — the classic mismatch being security mode: port 465 must be "SSL/TLS"
(implicit), port 587 must be "STARTTLS"; crossing them fails exactly like this. Residuals
for deliverability once sending works: **DKIM absent** (no default._domainkey record) and
**DMARC absent** — both enabled at the host (cPanel → Email Deliverability), NOT
Cloudflare. ⚠ mail.contactforsales.com:25 does not answer externally — fine (the PMG
cluster fronts inbound), noted so nobody reads it as the fault.

**Mailer config for the delivery email when it ships** (RUNBOOK'd; no secret held):
host mail.contactforsales.com, port 465 (platform/mailer's UsesImplicitTLS(465) already
handles implicit TLS), username webdesign@contactforsales.com, password via secret ref
only. The owner setting this address up implies the contact-address question (old D3)
resolves as KEEP contactforsales.com — observed, not ruled; worth one explicit confirm.

**ZIP state, answered:** the ZIP MECHANISM works — DGH-011 live since 08-18, canary 8/8
byte-verified, cut on demand via zip-deliverable-dispatch, presign ≤7 days. The
CUSTOMER-FACING 30-day download link (/d/<token>) is NOT built: pre-mint+refresh design
settled (DECISION_2026-08-21b), vhost has no /d/ block, and since 67f1794f9 core-manager
structurally refuses a /d/ route until vhost + boxServablePrefixes land together. So: we
can hand a ZIP today; the durable emailed link is the unbuilt half.

## 2026-08-26 (evening) — webdesign lane confirms: copy hold RELEASED; ban ownership settled; the email's critical path is now entirely this lane's

The webdesign session recorded both rulings with this lane's trail reference and confirmed
the net effect: terms/register/copy stand, their 08-25 resume point applies, and **the
delivery email's copy hold is RELEASED** — the two product questions were its stated
reason. Division of labour now explicit:

- **Theirs:** the approval-language OFFER-SHAPE ban (08-19 round-of-changes narrowing as
  precedent, both-halves probe via claimscan) — armed AFTER their in-flight three-page
  prominence wave lands, so their own rebuilds don't trip it. The contact-address KEEP
  observation goes to the owner in their next status for one explicit confirm.
- **This lane's, in dependency order:** (1) the `needs_delivery_review` PRODUCER —
  contract fully pinned in code (via `actions.writeWorkItem`, filed at
  `needs_human_review`, `ReviewItemRequiredSpec()`'s `checkpoint:true`); (2) the email
  compose + send step (`platform/mailer`, settings runbooked, port 465 implicit TLS);
  (3) `/d/<token>` + vhost block + `boxServablePrefixes` entry, one commit.

**Design note for the producer, so the next session doesn't rediscover it:** the admin
console's `HandleCreateWorkItem` cannot substitute — it hardcodes `status='triaged'`
(verified 08-25 during the Fable round) and the approve handler requires
`needs_human_review`, so a console-filed review could never be approved. The real open
question is the TRIGGER — what event marks a site "ready to deliver". For the first
customers the honest answer is probably an explicit operator dispatch (no build-completion
seam exists for customer sites yet, 0 having ever been built); wire the action so any
later automatic trigger can call the same thing.

**DKIM/DMARC** stay owed at the HOST before any real customer send (NOTES entry above).

## 2026-08-26 (night) — THE REMAINING PATH BUILT: producer + /d/ + the email. Phase 4's code is COMPLETE.

Owner: "please go ahead with the remaining path." Done, in dependency order, one commit +
one council round (`Council-Submitted: 5a33a174`), everything mutation-proved and green at
HEAD `98aff4d7e`. Register: **DGH-018**; DGH-017's key shape corrected
(`delivery_review_<domain>` — what `create_work_item` actually files, not what my helper
wished for).

**The producer is ZERO new Go.** `create_work_item` already documents `needs_human_review`
+ no handler as the HITL idiom and takes `spec_literal` — so `delivery-review-filer` is a
seed (651, _HOLD). The admin console could NOT substitute (hardcodes `triaged`, which the
approve button 400s).

**/d/ is DB→302 with no standing credentials.** Migration 650 stores zip-deliverer's
presign ON the token row; `LookupZipURL` is one statement; stale NEVER redirects (an
expired presign 403s as SignatureDoesNotMatch — the landmine) — honest refresh page +
`ZIP_LINK_STALE` persisted to `agent_error_log`, the swept channel. **Stated deviation**
from 21b §4's literal "file a work item": that rides the REFRESHER, which is the one
remaining unbuilt half. The vhost `/d/` block + `boxServablePrefixes` + the route landed
in ONE commit, as the SYS-095 guard demands — the /d/ test flipped back to accepted.

**The email's order is the design.** Sender constructed BEFORE the claim; template
validated against producible links BEFORE the claim (the gap my own dead test line
exposed: `{{zip_link}}` with no presign FILLS AS EMPTY — "Your files:" then nothing —
invisible post-fill because the fill succeeded); then Claim (gate + once-only stamp +
minting); then fill the closed vocabulary, refuse survivors with an error that names the
recovery recipe. Copy is DB config in seed 651, owner-editable without a roll, figures
from the attested register.

**Owed before the first real send, in order:** apply 650 → roll → apply 651 →
`DELIVERY_SMTP_*` env + secretKeyRef on the chassis (owner holds the password) → DKIM +
DMARC at the HOST (still absent tonight) → box vhost re-apply (the /d/ block, gated on the
roll per its own header) → then the flow in 651's header: file review → owner edits +
approves → cut ZIP → dispatch the email. **The refresher is the follow-up build.**

## 2026-08-26 (late night) — 650 APPLIED; chain APPROVED r1; REFRESHER built; the mail question answered

**Migration 650 APPLIED + recorded** at the owner's direction (guard passed, verify
passed, columns confirmed nullable at the schema). 651/652 stay **_HOLD** — they name
`send_delivery_email`/`refresh_zip_link`, which exist only from the next roll
(image-before-seeds).

**Council `5a33a174` (the delivery chain) APPROVED at 17:00, round 1.**

**The REFRESHER is built** (`refresh_zip_link` + seed 652, council
`Council-Submitted: c618d189`): 6h schedule, re-stamps anything dying within 48h; the
action's WHERE cannot touch revoked or expired tokens (mutation-proven); zero rows is a
benign-race WARN. The pre_query's NULL arm is load-bearing — recovery-recipe tokens with
no stored URL would otherwise be permanently invisible. Spawn→call's ~50% handshake
failure is inherited and self-heals at the next tick; the honest health meter is the
outcome query in the seed header, never run status.

**The mail question:** the owner hit cPanel's "not authoritative" wall on DKIM and asked
whether to change email host. **No** — the server is healthy (measured yesterday and
today) and a different host would face the same wall: DKIM/DMARC must land at
`dns1/dns2.uk-noc.com`, the domain's authoritative NS. Three routes in the RUNBOOK
(Zone Editor first; host ticket second; move DNS to our CF account third — mailbox
unmoved either way). I verify from outside with dig once records are added.

## 2026-08-26 (night, addendum) — domain_programme key shape reviewed AS ITS FIRST READER; two additions asked for pre-first-row

The webdesign lane recorded two owner rulings (interim TAG = DESIGNCONSULT; the domain is
a SEVERABLE opt-in per-site layer, `site_config.domain_programme`, default absent/ugg2)
and built the EPP register + CF zone scripts (both dry-run-default; the zone script now
lives in THIS dir as `cf_customer_domain_zone.sh`; brief =
`BRIEF_2026-08-26_domain_find_register_point_service.md`). They asked for shape objections
BEFORE anything reads the key. P5 wiring (hostname choice + email domain section) is mine.

**My review, sent 2026-08-26 night — bones right, two gaps named while they cost one line:**
1. **`zone_live_at`** — `registered_at` stamps the MONEY event; serving needs the ZONE
   event. Between EPP create and the CF zone answering at the assigned NS there is a real
   window where hostname=domain breaks a serving site. The zone script's verify step is
   the right writer. P5 reader rule: hostname = domain ONLY when `zone_live_at` present.
2. **`commercial: rent|bought`** — the email's domain section, the weekly chase and the
   post-sale Registrant Transfer (owner 08-21 D1) all branch on it; reserve the name now
   or the first writer invents another.
Plus the reader rule I will encode regardless: **an unrecognised mode fails safe to slug
serving** — mode is an open vocabulary ("transferred" is clearly coming), and a reader
that errors on mode #3 takes a live site down, while ugg2-fallback keeps it up.

**Addendum, same night: ALL THREE ADOPTED in full** (webdesign lane, `0bc658c0c`):
`zone_live_at` split from `registered_at` — and the zone script now emits the spec-write
instruction ONLY on its verified re-read path, with the mismatch path explicitly saying
do-NOT-stamp, so the serving stamp is mechanically tied to verification;
`commercial: rent|bought` blessed as the reserved name; the unrecognised-mode-fails-safe
rule written into the BRIEF as agreed-before-any-reader-exists. **The P5 contract is now
SETTLED — whoever builds P5 wiring: read `BRIEF_2026-08-26_domain_find_register_point_service.md`
(RULINGS + shape) and encode: hostname = domain ONLY when `zone_live_at` present;
unknown mode → slug serving; email domain section branches on `commercial`.**

## 2026-08-26 (very late) — THE ROLL LANDED AND THE WHOLE CHAIN IS NOW LIVE CONFIG

**The fresh build verified the hard way, and the empty-stamp trap bit again:** chassis
pods (started 20:24Z) had rotated the provenance line out of log range in 17 minutes, and
an empty `$STAMP` made ancestry read "NOT in chassis" — the failed control caught it. My
first binary probe was also WRONG (grepping for MY commit's sha — the binary carries ONE
stamp, the build commit, not its ancestry). Correct route: probe `/proc/1/exe` against
tonight's CANDIDATE commits. **Stamp = `b34c24f4c`** — same commit as core-manager, both
delivery commits (`10a963da2`, `aca0afe1d`) proven ancestors.

**Applied tonight, verify-then-record:** seeds **651 + 652 LIVE** — all three agents
active (`delivery-review-filer`, `delivery-email-sender`, `zip-link-refresher`) and the
`zip-link-refresh` schedule enabled at 6h. First attempt FAILED (my invented
`ON CONFLICT (type)` — no such constraint; 459's real pattern is `WHERE NOT EXISTS`) and
I had batched the ledger INSERT with the applies, briefly recording failures as applied —
retracted within a minute, logged in WRONG_CALLS (apply → verify → record, three acts).

**Also wired: `DELIVERY_SMTP_*` into the chassis overlay** (`6d76dab1e`) — with
`optional: true` on the secret ref as the load-bearing line: without it a missing mail
password would CreateContainerConfigError the ~46-pod fleet; with it the cost lands where
it belongs (send fails loud, pre-stamp). ⚠ The env reaches pods at the NEXT `apply -k` +
restart, not tonight's — sender construction fails cleanly until then.

**Learned from tonight's commit stream:** a DIFFERENT `651_...sql` exists (robot-hands
gripper page — the number-collision class; resolve by filename); **Stripe keys are LIVE
in the cluster** (webhook keyed, `ad8a9b596`) though no Payment Links exist yet, so the
email's reply-to-arrange stands with config slots ready.

## 2026-08-26 (late night, fresh session) — cold-start verification pass: the hardening verdict is APPROVED, its advisory already closed by 651; every remaining step is the owner's

All the handoff's falsifiers re-checked ~21:0x–21:2xZ [MEASURED 2026-08-26]:

- **The 0b84970d post-approval hardening round LANDED and is APPROVED** — 08-26 10:29,
  1 advisory objection, none high-severity (`doc_notes` council-gate row; full report in
  `diagnosis_artifacts` kind=council_report). 25cd3044 round 4 likewise APPROVED 10:26,
  all reviewers. So handoff §1.5 is done and there is no REVISE to act on. (The handoff's
  "submitted 08-26 night" was off — the resubmission and verdict were the same MORNING.)
- **The one advisory** (editquality, medium: `ReviewItemRequiredSpec()` added but no
  producer shown calling it — a producer omitting checkpoint:true would still stall) was
  **closed the same night by the producer itself**: verified at the LIVE row, not the seed
  file — `delivery-review-filer`'s `file_review` config carries
  `spec_literal = {"checkpoint": true}` (`agent_definitions`, queried tonight). No
  follow-up owed.
- **Mail secret `delivery-smtp-secrets`: ABSENT** (`kubectl get secret` → NotFound).
  Handoff §1.1 stands, owner-only.
- **DELIVERY_SMTP env: NOT on the live deployment** (env-name grep over the deploy spec →
  0 matches; `last-applied-configuration` → 0 matches) — exactly as the handoff predicted:
  the running pods ARE the 20:24Z roll (ReplicaSet timestamps 20:23–20:24Z) and
  `6d76dab1e` was committed 20:47Z, after it. **Deliberately NOT forcing a one-service
  `apply -k`**: releases are whole-fleet here, and a ~46-pod chassis restart kills
  in-flight council runs (one was live at 21:04 tonight). The env rides the next release;
  send fails loud pre-stamp until then, which is the designed cost.
- **Box `/d/`: NOT re-applied** — from outside, GET `/d/<43-char junk>` → **404** (146 B)
  with the `/c/` → 200 reach control in the same run. Post-apply the same probe must
  return the uniform "no longer active" 200. §1.3 stands, unblocked, owner box step.
- **Nothing-customer-facing invariant holds**: `customer_access_tokens` 0 rows,
  `sites.handed_over_at` non-null 0.
- Working-tree note: the ~18 dirty `uk_001` kustomization tag bumps (1239→1345) are the
  release session's uncommitted work (live chassis image IS v1.0.1345) — left untouched.
- Stale line spotted, not edited: 651's header still says DKIM/DMARC "absent as of
  2026-08-26" — **closed that evening** (dkim/spf/dmarc all PASS at Gmail, `e38760737`).
  Not correcting the applied seed file in place: it is recorded in `schema_migrations`
  and a comment edit risks the drift guard; this note is the correction.

Net: **no session-side work remains before the rehearsal.** The critical path is entirely
the owner's three small steps — mail secret, box vhost re-apply, and the APPROVE press
when we rehearse (plus the env riding the next fleet release).

## 2026-08-27 (morning) — the 502 was the BOX's nginx, dead since an unattended upgrade at 06:22; fixed, hardened, and items 2+3 of the handoff CLOSED in passing

Owner reported preview.webdesign.uk 502 while asking for the go-live walkthrough. Chain
of evidence, all [MEASURED 2026-08-27]:

- **Not just preview**: links.webdesign.uk 502'd too — including `/other`, the path
  nginx answers LOCALLY without touching the cluster — in ~60 ms, body `error code: 502`
  (cloudflared's origin-refused response). Cluster healthy throughout (no bad pods;
  the delivery listener answered 200 in-cluster on the new pods). So: box-side, between
  tunnel and nginx.
- **Root cause at the box journal**: unattended-upgrades ran 06:21–06:22Z; the nginx
  restart it triggered died at start — `[emerg] host not found in upstream
  "admin-dashboard.ai-persona-system.svc.cluster.local"` — a transient cluster-DNS
  (WireGuard leg) failure at the moment of startup. `nginx -t` passed at diagnosis time
  (DNS back), so the config was innocent. nginx dead 06:22→08:32Z, cloudflared healthy,
  hence uniform fast 502s.
- **Fix**: `systemctl start nginx` → all vhosts back (preview 200, /c/ 200 outside).
  **Hardening**: drop-in `/etc/systemd/system/nginx.service.d/retry-on-failure.conf`
  (Restart=on-failure, RestartSec=15s, StartLimitIntervalSec=0) — this class now
  self-heals in ≤15 s. LANDMINES entry appended (the 'error code: 502' signature reads
  as cluster-down and is never cluster-side) + verify dispatched.
- **HANDOFF ITEM 2 CLOSED BY THE OVERNIGHT ROLL**: the fleet rolled to v1.0.1346
  (~23:30Z 08-26) and the chassis pods NOW CARRY `DELIVERY_SMTP_HOST/PORT/USER/FROM`
  (printenv on a live pod). `DELIVERY_SMTP_PASS` absent — secret still not created
  (re-verified NotFound). ⚠ secretKeyRef env resolves at container START: creating the
  secret does NOT reach running pods; the PASS arrives at the next roll/restart after
  the secret exists. Daily rolls (1344→1345→1346 on consecutive days) make this
  automatic within a day; a deliberate `rollout restart deploy/agent-chassis` is the
  faster option but kills in-flight councils + ~300s no-dispatch — owner's call.
- **HANDOFF ITEM 3 DONE (by this session — the runbook's 08-25 correction stands:
  sessions CAN ssh the box)**: backed up to `/root/links.webdesign.uk.bak-2026-08-27`,
  applied the repo vhost with the /d/ block, `nginx -t` clean, reloaded. Outside table:
  `/d/<43-junk>` → **200 "no longer active"** (the uniform page) · `/d/x` → 404 ·
  `/c/<43>` → 200 · `/other` → 404 · apex still parked-302. Box apex vhost re-verified
  `/c/`-free (grep 0).
- **Go-live gate 2 had FAILED and is FIXED**: the served preview carried ZERO
  "Not active yet" labels — vm-sites `ba44c5c` (Rerender: index.html) landed after the
  last re-placement, exactly the strip the 08-26 rotation note predicted. Re-placed at
  both insertion points (vm-sites `b72c608`), sitesync triggered, served page verified
  ×2 (`grep -c 'hand-placed 2026-08-25'`).
- **Go-live gates re-checked**: safety counters 0|0|0 (runbook query verbatim); edge
  known-safe (apex+www 302→webdesign.co.uk, preview /c/x 404, control 200); apex vhost
  /c/-free. Stripe restored per the webdesign handoff's DONE block (08-26 late night).

Remaining before first delivery email: (1) owner creates `delivery-smtp-secrets`;
(2) one chassis restart/roll AFTER the secret exists; (3) the rehearsal. The shopfront
unpark (Cloudflare page rule, owner dashboard-only) is independent of all three.

## 2026-08-27 (mid-morning) — mail secret CREATED (owner); terraform question answered NO; Lovable removal filed through the framework

- **Owner created `delivery-smtp-secrets`** and will run the fleet deploy in a quiet
  moment — that deploy is what carries `DELIVERY_SMTP_PASS` onto the pods (secretKeyRef
  resolves at container start).
- **"Should the secret live in terraform?" — NO, and here is the mechanism** [checked at
  the config 2026-08-27]: `047-base-configs/main.tf` manages exactly two resources — the
  `personae_prod_config` ConfigMap and `kubernetes_secret.personae_platform_secrets`
  (whose data map it owns WHOLESALE — that ownership is what wiped the Stripe keys on
  08-26). `delivery-smtp-secrets` appears nowhere in terraform, so a release cannot
  touch it: the Stripe failure mode needs terraform to OWN the secret, and nothing owns
  this one. Moving it in would put the password into `terraform.tfvars.secret` and tf
  state — against the stated posture ("exists in no file and no session", 651 header,
  overlay comment). If belt-and-braces is ever wanted, the Stripe pattern (required var,
  value only in the owner's local tfvars.secret) is the shape — owner's call, not
  recommended while nothing owns the secret.
- **Blueprint Compiler / Lovable (owner instruction, through the framework):** the tool
  (webdesign.uk `/tools/blueprint-compiler/`) directs visitors to paste its prompt deck
  "into v0, then Lovable" — description line, two deck-block titles, one embedded
  prompt opener (4 places, read from the component template). The component is
  `render_mode='template'`, no input fields, so this is TEMPLATE copy — no field edit
  can reach it, and a vm-sites HTML edit is both forbidden (owner) and futile (next
  rerender regenerates). **The framework's own change mechanism for an existing tool is
  `improve_tool` → `tool-improver`** (319 completions all-time, several today — the
  same route the acceptance/audit checks use). Filed TWO items with the instruction
  (remove Lovable AND v0 — same recommendation class; keep the three-phase deck,
  no replacement product names): `b611b4cd` for the webdesign.uk instance component
  (`5b0110c3`), dispatchable now; `c51317f4` for the webdesign.co.uk library original
  (`ad0cda73`) with `depends_on={abc6ac7c}` — the already-open acceptance-failure
  improve_tool on that same component (loader honours depends_on,
  `load_work_item_actions.go:763`), so two writers never race one component. Open
  audit_tool items on the instance are read-only; the pending `section_edit`
  (component-template-fixer propagation) writes the PAGE, not the component.
  **Verify at the served page once complete:** `curl -s
  https://preview.webdesign.uk/tools/blueprint-compiler/index.html | grep -ci
  'lovable\|v0'` → 0 (and the .co.uk page likewise after its chain clears).

## 2026-08-27 (late morning) — OWNER REVERSAL: the v0/Lovable references STAY; both improve_tool items cancelled unclaimed

The owner reversed the morning instruction before anything ran: *"leave the v0 and
Lovable references... We can make much more of third party services in the future rather
than less. We offer a different service to them and we may in future recommend them if
our service doesn't suit."* Both items (`b611b4cd`, `c51317f4`) were cancelled with a
guarded UPDATE (`status='triaged' AND claimed_at IS NULL` — both matched, so neither was
ever claimed; `retry_feedback` carries the do-not-re-file note). The monitor was stopped.
Served page verified untouched: Lovable still present ×2 [MEASURED 2026-08-27].

**Standing position for future sessions: the third-party builder mentions in Blueprint
Compiler are DELIBERATE (owner ruling 2026-08-27)** — do not file removal on sight of
them; the direction of travel is MORE third-party positioning, not less. (Gotcha logged:
`retry_feedback` is jsonb — a bare string errors; wrap with jsonb_build_object.)

## 2026-08-27 (~10:00Z) — the shopfront UNPARKED and verified; this lane's launch gate is now open

webdesign.uk went LIVE (owner toggled both page rules off; full account + the 522/DNS
trap in the webdesign lane's NOTES 08-27 and the LANDMINES page-rule entry's dated
UPDATE). For THIS lane that closes "delivery waits on the shopfront launch": what stands
between us and the first real delivery is now only the owner's quiet-moment fleet deploy
(carries DELIVERY_SMTP_PASS — secret exists since this morning) and THE REHEARSAL.
Safety counters re-read before the unpark: 0|0|0.

## 2026-08-27 (afternoon) — the owner's deploy LANDED: email fully armed; terraform Stripe fix PROVEN on its first release

All [MEASURED 2026-08-27 afternoon], post-deploy (fresh chassis generation `7df947c88b-*`,
~2.5h old at check):
- **`DELIVERY_SMTP_*` COMPLETE on pods including PASS** (presence + byte-length only —
  the value was never read into the session, per the 08-23 owner rule). Every
  prerequisite of the first delivery email is now live.
- **Stripe webhook → 400 keyed** — the first whole-fleet release since `0cdc9e2d9`
  did NOT wipe the keys: the required-terraform-variable fix held. (This was the
  post-roll check the 08-26 revert made mandatory.)
- Listener routes re-verified from outside post-roll: `/c/<43>` 200 · `/d/<43>` 200
  uniform page · `/other` 404 · apex 200 · label ×2 · counters 0|0.
- Fleet healthy (one transient ContainerCreating research-agent spawn, seconds old).

**Monday's pickup = label re-check + THE REHEARSAL, nothing else.** Site ruled
remortgagecalculator.uk; email destination info@designconsultancy.co.uk (unobjected
proposal — confirm in passing).

## 2026-08-27 (mid-afternoon) — chat "outage" = the ANTHROPIC USAGE LIMIT, not the deploy; brief-starter's old flow diagnosed and filed

**The chat fallback the owner saw was the API limit, and the timing was a pure
coincidence with the deploy.** Evidence chain: /api/chat answered a REAL reply at
~09:55Z; by 13:32Z every question (including that morning's verbatim) drew the
fail-closed contact line. `webdesign-chat` journal on the box has the truth from
11:39:39Z onward: `anthropic 400 ... You have reached your specified API usage limits.
You will regain access on 2026-09-01 at 00:00 UTC` on every call. The fail-closed arm
(chat.go contactLine) behaved exactly as designed. **The FLEET shared the key's fate**:
`llm_call_log` went 0-failures (10:00 hour) → 100% failures (12:00+). **The owner raised
the limit ~13:50Z ("I had spent too much") and chat was verified BACK within a minute**
(real reply, from outside). Dark-window damage checked against a baseline before
worrying: items going terminal 11:35–14:00 today = 71 vs SAME WINDOW YESTERDAY = 129 —
below normal churn, **no casualty sweep owed**. Lesson (the instrument-doubt family): a
fallback message styled as helpful copy IS an error arm — grep the service's journal
before blaming the deploy that happened to precede it.

**Brief-starter "old flow" (owner report, screenshot):** the tool's ending still says
copy the summary into "our contact form ... before we speak" — three places in fork
component `852886be` (description line, summary-step label, tool-doc purpose comment).
**Verified at the live site: NO contact form exists** — contact.html's only `<form>` IS
the chat (`data-chat-form` → /api/chat), which is the order intake; and "before we
speak" contradicts the one-shot pay-before-build model. The NOTES trial-loop design
already names "brief-starter intake" as the flow head. Framework route again:
`improve_tool` **be0bdf28** filed (destination = the chat; no conversation/approval
promises; register facts govern; mechanics untouched; no quoted exemplar copy — the
improver words it). No open items on the component at filing. Monitor armed; verify at
the served page when complete: `grep -ci "contact form"` → 0 on
/tools/website-brief-starter/index.html, and the guide stays clean (it already greps 0).

**Addendum (post-limit-raise "Something went wrong", ~13:37Z):** a THIRD distinct arm,
and the diagnosis chain matters: the client-side error (frontend generic text) masked a
**429 from the chat service's own per-IP limiter** — `newChatIPLimiter`: **5 new
conversations/hour, 15/day per visitor IP** (ratelimit.go; the 35-byte body "too many
requests, try again later" in access.log at 13:37:13/13:39:00 was the tell — nginx's
limit_req was innocent: no "limiting requests" in error.log, and it returns 503 not
429). The owner's own testing burned the band — every attempt today was a NEW
conversation, including the fallback-answered ones during the outage. Counters are
in-memory BY DESIGN (the spend ledger is what survives restarts) → `systemctl restart
webdesign-chat` cleared them; verified answering. **Operator recipe: owner testing
heavily from one IP WILL trip 5/hour — restart the service to reset, or ask for the
band to be widened.** Real-visitor traffic cannot realistically brush it. Three chat
failure shapes now distinguished for the next reader: contactLine as the REPLY =
server-side Claude-call failure (read `journalctl -u webdesign-chat`) · generic
"Something went wrong" under the input = HTTP failure, check access.log status (429 =
the service limiter, 503 = nginx limit_req) · silence = the box/tunnel (the 08-27
morning landmine).

**be0bdf28 COMPLETE and verified at the component (13:55Z):** template 20,742 → 20,781 B
(copy-sized, no truncation), "contact form"/"before we speak" GONE, all three passages
now point at the chat with order framing ("paste straight into the chat, found on our
home page and contact page, to start your order" · "Copy this into the chat" · purpose
comment matches). Mechanics byte-identical around the edits. Served page follows when
component-template-fixer files its rerender (it files continuously — six today);
monitor armed on the served URL. This is the second successful owner-directed
improve_tool through the framework today — the route is proven for copy-level tool
changes.

## 2026-09-03 (evening) — 466 fixed, and the census made the root cause worse than the bug file's

Picked up from `HANDOFF_2026-09-03_continue_here.md`. Items 2, 3 and 5 of its NEXT list are with the
owner or other lanes; item 1 is a monitor; item 4 (the 651 rehearsal) needs the owner's say-so
because it sends a real email and burns a once-only handover stamp. So: `bugs_open/466`, this lane's
own bug, filed today and owned here (`scripts/who-owns.py 466` → this workstream).

### The monitor for item 1, and one thing the handoff's own check would have got wrong

Re-armed the served-page watcher (the previous session's `bw8thrkta` died with its session; monitors
are session-local). Baseline `last-modified: 15:15:32Z`, i.e. the edits were still unpublished.

**The handoff's check 3 says "zero `cta-subtitle` elements". A bare grep for that string cannot
express it** — the served page carries **two** occurrences and they are different things: one
rendered `<p class="cta-subtitle">` and one CSS rule `.cta-subtitle { margin-bottom: 2rem; }` inside
the page's own `<style>` block. The CSS rule survives the edit, so a bare-string check will read
**1** for ever and be misread as "the edit did not land". The check is `class="cta-subtitle"` = 0,
with the bare count ≥1 kept as the **liveness control** that the section still rendered at all. This
is `MEMORY[a grep occurrence-count conflates a CSS rule with a rendered element]`, met head-on.

Read the template before arming, rather than after being surprised by it: `call-to-action` gates the
element on `{{if .subheadline}}`, and Go treats `""` as false, so an empty subheadline omits the
`<p>` entirely — there is no checklist-5.1 empty-`<p>` risk from the owner's "cut it". Watcher
foreground-tested against the pre-publish page first: it reports FAIL on exactly the three checks
that must change and PASS on the one that must not, which is what makes a later PASS mean anything.

### 466 defect 1 is not what the bug file says, and a census is what showed it

The file said the approval flow and `include_fields` "disagree about where approved content lives".
Read `checkpoint_for_review_action.go:157-180` and that is too kind: the review item's spec is built
from a **fixed literal** — `review_data`, `checkpoint`, `source_agent`, `correlation_id`, then
optionally `domain`, `spec_aspect`, `on_approve`. No arbitrary field can ever be at its top. The
approve handler looked its `include_fields` names up in exactly that object.

```sql
WITH ck AS (SELECT id, spec, jsonb_array_elements_text(spec->'on_approve'->'include_fields') AS fld
            FROM site_work_items WHERE spec->'on_approve' ? 'include_fields')
SELECT count(*) AS mentions, count(*) FILTER (WHERE spec ? fld) AS resolvable FROM ck;
-- 42 | 0     [MEASURED 2026-09-03, all history, 21 items, first 2026-08-24]
```

**Zero of 42, ever.** So `include_fields` is not half of a broken join — it is a mechanism that never
copied anything for any consumer, including the two names in `checkpoint_for_review`'s own header
comment (`reviewed_brief`, `site_record`). The measurement is disconfirmable: a producer writing the
named field at spec top level shows as `resolvable > 0`.

### Two things I nearly missed, both found by looking at the population rather than the case

1. **The addresses rot.** `LANDMINES` (copy_quality_two_stage, 2026-08-18) says a rerender REPLACES
   the `page_components` row with a new id. `[MEASURED 2026-09-03]` of the **31** edits parked in
   `needs_human_review`, **3** point at a row that no longer exists; 0 are stale by `updated_at`.
   That is a consequence *of the fix*: making fan-out work unblocks 16 parked proposals, and three of
   their edits would file `section_edit` items guaranteed to die at `load_edit_context` with nothing
   said to the approver — i.e. the fix would manufacture more of the bug it fixes. Hence the
   dead-address refusal.
   ⚠ **The staleness arm of that query is artefactual on `complete` rows** (4 of 4 "stale") because
   applying an edit is what bumps `updated_at`. It discriminates only for unapplied items. Do not
   quote the `complete` figure.
2. **Prior art the bug file missed.** `copy_quality_two_stage` hit this the day before: review
   `be23d897`, approved in chat 2026-09-02, released **by hand as two `section_edit` items**, its own
   `result` saying *"replicating the dashboard `on_approve` contract in the proven `section_edit`
   spec shape"*. Two lanes independently hand-built the fan-out before anyone proposed it as a fix.
   That is the strongest evidence for candidate 1 and it was sitting in a `result` column.

### What the migration's shape was decided by, not guessed at

`[MEASURED 2026-09-03]` of the 41 proposed edits across those 21 items: **41** carry
`page_component_id`, **0** carry `edit_type`, **0** carry `page_name`. So `defaults` must supply
`edit_type` and must **not** supply `page_name` — every edit is addressed by `page_component_id`,
which alone satisfies `load_edit_context`, and a defaulted `page_name` would be a guess applied to
every page on every site. `include_fields` becomes `["domain"]`, a key the spec genuinely holds and
what the two proven hand-filed items carried.

### The tests were mutation-tested, because green is not evidence

Four mutations, each restored immediately and the file checksummed back to identical: remove the
dead-address guard · remove the approved-body fallback · make fan-out unconditional · drop the
element-field merge. **All four killed their own test.** The third is the one that matters longest —
it is what stops a later tidy-up making fan-out the default for every checkpoint consumer.

`33dfeed3a`, council `d04c1bc1-b9a3-41bb-b144-1d101e68e542` (submitted, verdict pending — trailer is
`Council-Submitted:`, which asserts nothing). HEAD `a26cc1313` verified green with the change in it.
**Inert until a roll**: `git merge-base --is-ancestor 33dfeed3a <admin-dashboard stamp>`.

### My own wrong call today, logged in full at WRONG_CALLS

Session `332` showed that `sweep_site_defects.sh` §1.4 is blind by construction: it greps the served
HTML, and every news page ships JS that overwrites that HTML with `/data/news-archive.json`, which
carries 7 unstripped ATX headings against the server HTML's 0. Fleet-wide, five hosts. They own the
fix; I told them to use the script's own `blind()` helper rather than printing the word, because
`blind()` is what makes the **exit** non-zero — printing "blind" while exiting 0 would reproduce the
very shape one layer down.
