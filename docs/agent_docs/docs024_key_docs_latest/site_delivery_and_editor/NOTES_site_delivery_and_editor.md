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
  1. Bump IMAGE_TAG, `make build-agent-chassis` from committed HEAD, roll —
     then verify at the binary (provenance stamp, per SERVICE).
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
