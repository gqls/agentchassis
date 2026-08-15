# HANDOFF 2026-08-15c — PLAN steps 1–4 of 6 DONE; next build work is STEP 5 (tool-deployer backend half, proven on a SECOND site); 285 fix moved to a SEPARATE lane — SUPERSEDES HANDOFF_2026-08-15b

**Start here cold.** Read order: this file → NOTES tail (all 2026-08-15
evening/night entries) → PLAN_2026-08-11 §4–§5 (the step you are about to
build) → the approved EXPERIENCE_PLAN (query in §3 below) → the sibling
lane's newest `ai_site_selling_automation/HANDOFF_*`.

## 0. State in one paragraph

PLAN_2026-08-11 ("chat box as a framework capability") is four-sixths done
and all of it is verified live: the chat-input-box is a registered library
tool (step 1), the `requires-backend` gate is live in tool-suggester
(step 2, migration 406), the "site chat intake" **EXPERIENCE_PLAN is
council-approved and persisted** (step 3, corr
`8b0f77bf-592e-4280-a167-12113311ca98`, doc_plans is_current 11,152 chars),
and the facts relay is live, DNS-named, and has now survived **two** fleet
rolls unattended (step 4 — the 18:44Z fresh-build roll re-proved both the
terraform-owned token AND the new named FACTS_URL). The Stripe webhook proxy
is built and proven keyless end to end through the public edge. Box
cluster-DNS is fixed and durable; both ClusterIP pins are retired. The
lock-blind-assembler defect this lane diagnosed (two-round 090 trail) is
filed as `bugs_open/285` *(section_list_assembly slug — AMBIGUOUS NUMBER,
two 285s exist; resolve by slug)* and its **fix is being implemented by a
SEPARATE lane (owner, 2026-08-15)** — this lane keeps the chat-box lock ON
and runs the acceptance when their fix lands.

## 1. Owner decisions PENDING (unchanged answers change nothing below)

1. **Contact email** — the live `contact` fact says
   `webdesign@contactforsales.com` (faithfully sourced from the `sites`
   row); domain doesn't match webdesign.uk. It is the address every chat
   fallback journey points at, so the approved plan gates journeys C/D on
   it (its Step 0), and an open `content_rewrite`/`needs_human_review` item
   flags it. **Deliberate sales inbox, or leftover?**
2. **Stripe webhook URL** — apex + www 302 every path to webdesign.co.uk at
   the Cloudflare edge; Stripe treats 3xx as failed delivery. Register
   `https://preview.webdesign.uk/stripe/webhook`, or add an edge path
   exception (owner-side, Cloudflare dashboard).
3. Standing: Stripe keys later (via 047-base-configs terraform, NEVER
   kubectl); build-duration copy UNDECIDED (do not touch); CTA ink vs
   egg-gold UNDECIDED (owner checking himself).

## 2. Next work for THIS lane, in order

1. **PLAN step 5 — extend `tool-deployer` for the backend half.**
   `deploy_tool_to_site` already does the frontend (component-fork +
   tool-page + page_component link). What it cannot do, for any tool, is
   provision/register the BACKEND a `requires-backend` tool needs. Design
   anchors: PLAN_2026-08-11 §4 (Option A is the ruled shape — the SAME box
   binary parameterised per site, facts from each site's own
   `evidence_base` via the relay; §4's option table has the full
   reasoning), §5 (tool-auditor/improver already cover the frontend half;
   the backend half has NO audit mechanism — out of scope, don't pretend
   otherwise), and the approved EXPERIENCE_PLAN's data contract (§3 below:
   site identity + relay URL + token pair + contact fallback + per-site
   control values). **The acceptance is a SECOND real site sharing
   webdesign.uk's box, served by the same binary with different
   parameters** — not a hypothetical. Platform code → council gate;
   register the seam (concept register) in the same commit per CLAUDE.md.
   Note: the chat-service source lives on the box AND in this directory's
   `box/chat-service/` — keep them in sync like the nginx config.
2. **PLAN step 6 — wire `tool-suggester` to cite the approved
   EXPERIENCE_PLAN** when recommending chat-input-box (only after step 5
   proves deployment works, so suggestions don't outrun delivery). The 406
   gate already scopes WHO can be offered it; step 6 adds the WHY/HOW
   citation. Likely a config/prompt migration in the 406 idiom.
3. **EXPERIENCE_PLAN MVP verification round** (no rebuilding — the plan's
   §4): Step 0 is owner decision 1 above and GATES the rest; Steps 1–4 are
   re-verifications (locked contact sections survive; `/api/chat` live
   same-origin; four controls configured for webdesign.uk specifically;
   `deploy_config.capabilities` still `backend`); Step 3 includes ONE
   controlled manual trip of cap/rate-limit/relay-down (never scripted
   against live traffic controls).
4. **Watch items** (not this lane's build work): the separate lane's 285
   fix (when live: run the five-step acceptance in the 285 file, then
   answer work item `a4cd5dc8` — the lock comes off only per owner);
   `bugs_open/282` (tool-resolver eats tool-level names — co-requisite for
   the chat-box case, 407 thread); `bugs_open/275` (LIMIT 30 hides 38/68
   tools — unclaimed, any thread may take it); migrations 418/419/420
   (bugs_open/276 class, another thread, planner surfaces).

## 3. What is LIVE and mine, verified 2026-08-15 night

- **EXPERIENCE_PLAN**: `SELECT body FROM doc_plans WHERE
  subject_type='experience' AND subject_key='site-chat-intake' AND
  is_current;` — journeys, promise ledger (four controls + fail-closed as
  contract), data contracts, MVP cut, acceptance criteria. Brief:
  `BRIEF_2026-08-15_site_chat_intake.sql` (doc_notes `6bf8f9a4-…`).
  Council verdict note: "approved with 1 advisory objection(s) — none
  high-severity" (the advisory = the contact-email gate).
- **Facts relay**: live mode on
  `http://core-manager.ai-persona-system.svc.cluster.local:8088/api/v1/site-facts/webdesign.uk`,
  token terraform-owned, proven across the 13:5xZ AND 18:44Z rolls. Health
  read: box journal `grep -i 'refresh failed'` — logs every 5 min on
  failure, silence is success.
- **Webhook proxy**: `location = /stripe/webhook` → auth-service by
  cluster-DNS name, proven through `preview.webdesign.uk` (503
  `billing provider not configured` = the keyless truth). Repo copy
  `box/webdesign.uk.nginx` == live file.
- **Box DNS**: wg0 PostUp attaches kube-dns `10.21.0.10`, routing domain
  `~cluster.local` (only cluster names cross the tunnel). Backups of all
  three changed box files: `*.bak-20260815`.
- **chat-input-box lock**: ON (permanent), stays on per owner until the
  separate lane's 285 fix is live and accepted.

## 4. Traps for this lane's next session (all hit + documented this session)

- Facts relay auth header is **`X-Facts-Token`** (facts.go:109), NOT
  `Authorization: Bearer` — a Bearer probe 401s exactly like a dead relay.
- `/etc/webdesign-chat.env` is a systemd EnvironmentFile, **not
  shell-sourceable** — grep/cut single keys.
- nginx: `systemctl reload` + immediate curl races the OLD config (404) —
  retry before diagnosing; named `proxy_pass` resolves at CONFIG LOAD (a
  Service recreate costs a reload, not an edit).
- Diagnosis plumbing: verdicts live in the diagnose-agent orchestration's
  `collected_data->'verdict'`; `diagnosis_artifacts` rows are iteration
  INPUT bundles. A 090 refusal listing only page-NAME matches on other
  sites + terminal `failed` rows from closed lanes is clearable with
  FORCE=1 — but read them first.
- **Both 285s**: refer by slug; `git log` the file path, never the number.

## 5. Falsifiers / re-check before trusting this file

- A newer handoff here; the separate lane's 285 fix may have landed (grep
  their commits for the slug + `who-owns.py 285`, then read BOTH files).
- The plan row: any later approved run supersedes `is_current`.
- `10.21.0.10` still kube-dns; names still resolving from the box
  (`getent hosts core-manager.ai-persona-system.svc.cluster.local`).
- Whether the owner has ruled on the contact email / webhook URL / keys
  (checks 1–3 in §1) — each unblocks work listed in §2.
- The sibling lane has moved — read their newest handoff.
