# HANDOFF 2026-08-15b — EXPERIENCE_PLAN APPROVED (step 3 done), webhook proxy PROVEN, box DNS FIXED (both pins retired), 090 on the lock-blind planner IN FLIGHT — SUPERSEDES HANDOFF_2026-08-15

**Start here cold.** Read order: this file → NOTES tail (the two 2026-08-15
evening entries) → the approved plan itself (in the DB, query below) → the
sibling lane's newest `ai_site_selling_automation/HANDOFF_*`.

## 0. State in one paragraph

All four of the morning handoff's sanctioned work items moved today. The
"site chat intake" EXPERIENCE_PLAN is **approved and persisted** (council:
approved, one advisory objection, none high-severity; corr
`8b0f77bf-592e-4280-a167-12113311ca98`) — PLAN_2026-08-11 steps 1+2+3 are now
all done. The Stripe webhook proxy is **built and proven end to end keyless**
(public edge → tunnel → box nginx → wg0 → auth-service 503
`billing provider not configured`). Box cluster-DNS is **fixed and durable**
(kube-dns on wg0 with routing domain `~cluster.local`), and **both ClusterIP
pins are retired** — FACTS_URL and the webhook upstream use
`*.svc.cluster.local` names; the old "one remaining fragility" is closed. The
improvement-loop defect (owner: "check and fix") went through one full 090
round: run `c199c4bf-e433-4fa7-8bbf-c64b627e7373` **REFUTED** the first
hypothesis (PlanSectionsAction is a CONSUMER of an already-assembled
`sections` input, not the composer — NOTES "late" entry + WRONG_CALLS.md row
record the correction) and relocated the mechanism upstream to where the
section list is assembled (`load_page_sections_from_spec` / whatever writes
`pages.sections`); a second 090 is filed there (see §2.1). The chat box lock
STAYS ON until the fix lands (owner ruling 2, 2026-08-15).

## 1. Owner decisions NEEDED (new since the morning handoff)

1. **Contact email**: the live `contact` fact (and the `sites` row it is
   sourced from) says `webdesign@contactforsales.com` — domain does not match
   webdesign.uk. It is the address every chat fallback journey points at, so
   the approved plan gates its journeys C/D on this (**Step 0**), and an open
   `content_rewrite`/`needs_human_review` item already flags it. **Deliberate
   sales inbox, or leftover?** If deliberate, Step 0 can be marked passed
   as-is; if wrong, fix the `sites` row and re-attest the fact (sibling
   lane's register trail).
2. **Stripe webhook URL**: apex + www **302 every path to webdesign.co.uk at
   the Cloudflare edge, and Stripe treats 3xx as failed delivery.** Register
   `https://preview.webdesign.uk/stripe/webhook` with Stripe, or add an edge
   path exception for `/stripe/webhook` (Cloudflare dashboard, owner-side).
   Relayed to the sibling lane (their Phase 6 cutover review should pick the
   canonical hostname).
3. Standing, unchanged from the morning handoff: Stripe keys later (via
   047-base-configs terraform, NEVER kubectl); build-duration copy UNDECIDED
   (do not touch); CTA ink vs egg-gold UNDECIDED (owner checking himself).

## 2. Next work, in order

1. **Read the SECOND 090's verdict** (item_key
   `needs_diagnosis:section-list-assembly-lock-blind` — run correlation in
   the NOTES tail; the first round REFUTED the PlanSectionsAction framing,
   see §0) and FIX per the verdict. The refuting run's own next_scope:
   `load_page_sections_from_spec` (the authoritative source of the
   `sections` input per plan_sections_action.go's bugs_open/151 comment),
   `save_page_sections` (the write step), and whatever populated
   `pages.sections` for contact around 08-11/08-13. Reading the verdict
   bundle: the verdict lives in the diagnose-agent orchestration's
   `collected_data->'verdict'` (diagnosis_artifacts rows are the iteration
   INPUT bundles, not the output). Acceptance (owner ruling): an improvement
   pass over the contact page **KEEPS the locked section in its proposed
   section list** — not merely "the lock blocked it again". The `a4cd5dc8`
   needs_human_review row is answered by the fix, not dismissed. Prior art:
   bugs_closed/058 (the write guard IS its fix), bugs_open/189 (positional
   lock duplication — adjacent), bugs_open/204 (named in
   plan_sections_action.go's comment about the sections input), bugs_open/276
   + pending migrations 418/419/420 (another thread is gating
   requires-backend sections in the same planners — coordinate, don't
   collide: their pre-state probes already read "concurrent edit?").
   ⚠ Do NOT re-bundle the hero/call-to-action lock_blocked_change rows as
   evidence — the refuting run showed those are the CTA-destinations
   defect's locks (content overwrites), a different mechanism.
2. **PLAN steps 5–6** (extend tool-deployer for the backend half; wire
   tool-suggester to cite the approved plan) — both now unblocked by step 3.
   Step 5's real test is a SECOND site on the box; the approved plan's LATER
   list says the second site is not this site's build.
3. **Execute the approved plan's MVP steps 0–4** (verification-shaped, no
   rebuilding): Step 0 = owner decision 1 above; Steps 1–4 are re-checks
   (locked sections survive rebuild; /api/chat live; four controls
   configured; capabilities still `backend`).
4. **Bugs for pickup, unchanged**: bugs_open/275 (LIMIT 30 hides 38/68
   tools) · bugs_open/276 (now being worked by another thread via
   418/419/420 — check before touching).

## 3. What is LIVE and mine, verified 2026-08-15 evening

- **EXPERIENCE_PLAN**: `SELECT body FROM doc_plans WHERE
  subject_type='experience' AND subject_key='site-chat-intake' AND
  is_current;` — 11,152 chars, five sections (Journeys / Promise ledger /
  Data contracts / MVP cut + LATER / Acceptance criteria). The DOC-076 brief
  that grounded it: `BRIEF_2026-08-15_site_chat_intake.sql` (this directory),
  doc_notes id `6bf8f9a4-…`. Intake row completed with resolution.
- **Webhook proxy**: `location = /stripe/webhook` in the box nginx (repo
  copy `box/webdesign.uk.nginx` — file and live are in sync, committed).
  proxy_stripe.conf shape: NO rate limit, no body rewrite, no X-Real-IP.
  Upstream `auth-service.ai-persona-system.svc.cluster.local:8081`.
- **Box DNS**: wg0 PostUp `resolvectl dns wg0 10.21.0.10` + domain
  `~cluster.local` (runtime + durable; backups `wg0.conf.bak-20260815`,
  `webdesign-chat.env.bak-20260815`, `webdesign.uk.bak-20260815` on the
  box). Bot journal on the named relay URL: `fetched 15 facts` + `live mode`.
- **Facts relay**: unchanged, live, release-proof (terraform token proven
  across the 08-15 roll — morning handoff §1 still accurate on this).

## 4. Traps met this session (also in NOTES/RUNBOOK)

- The facts relay authenticates via **`X-Facts-Token`** (facts.go:109), NOT
  `Authorization: Bearer` — a Bearer probe 401s exactly like a dead relay.
- `/etc/webdesign-chat.env` is a systemd EnvironmentFile, **not
  shell-sourceable** (unquoted phone number breaks `source` mid-file, and
  everything after it silently stays unset) — grep/cut single keys out.
- nginx reload race: an immediate curl after `systemctl reload nginx` can
  hit the OLD config and 404 — retry before diagnosing.
- Static `proxy_pass` with a DNS name resolves at CONFIG LOAD: a Service
  recreate now costs `systemctl reload nginx`, not a config edit.

## 5. Falsifiers / re-check before trusting this file

- The 090 verdict may have landed after this file was written — read it
  before starting work item 1 (the NOTES tail records where it stood).
- Whether another thread took the lock-blind fix meanwhile
  (`who-owns.py`, live transcripts) — 418/419/420 prove planner surfaces
  are being edited concurrently RIGHT NOW.
- The plan row: `is_current` can be superseded by any later approved run.
- Relay healthy: box journal `grep -i facts` — silence is success; ask the
  bot the price (£149).
- The sibling lane has moved — read their newest handoff, not my summary
  of it.
