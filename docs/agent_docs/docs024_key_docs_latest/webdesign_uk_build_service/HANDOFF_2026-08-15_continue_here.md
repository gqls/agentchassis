# HANDOFF 2026-08-15 — relay release-proof (PROVEN on a real roll), chat box is a gated library tool, PLAN step 3 is next — SUPERSEDES HANDOFF_2026-08-13

**Start here cold.** Read order: this file → NOTES tail (2026-08-13 evening
through 2026-08-15) → PLAN_2026-08-11 §3 (the step you are about to run) →
the sibling lane's newest `ai_site_selling_automation/HANDOFF_*` (they move
fast and their surface is shared with ours).

## 0. State in one paragraph

The webdesign.uk chat bot reads live £149 facts from `site_specs.evidence_base`
through the core-manager site-facts relay, and that pipe is now **proven
release-proof**: `SITE_FACTS_TOKEN` lives in terraform (047-base-configs), and
the 2026-08-15 fleet roll — the same event class that deleted the
kubectl-patched token on 08-13 and broke refresh for 18 hours — left the relay
untouched (token in secret + pod, zero failed refreshes on the box). PLAN
steps 1+2 are done and council-approved: `chat-input-box` is a real library
tool (`component_level='tool'`, `category='interactive'`) and tool-suggester
only offers `requires-backend` tools to sites with
`deploy_config.capabilities:['backend']` (migration 406, trail c78ed496
APPROVED round 3). The chat→pay→build loop is unchanged from 08-13: payment
(PAY-009) built but keyless; build trigger and transcript→brief still the
sibling lane's P4/design items. The owner has instructed both webdesign
threads to stay mutually aware — coordination flows through the sibling's
NOTES file.

## 1. What is LIVE and mine, verified 2026-08-15

- **Facts relay**: live mode since 08-14 08:12Z, durable across releases
  (measured, not assumed — NOTES 08-15 entry). Token pair: terraform
  `site_facts_token` (tfvars.secret, local, gitignored) ↔ box
  `/etc/webdesign-chat.env` `FACTS_TOKEN`. Rotate both together — RUNBOOK
  § "Restoring or rotating the facts-relay token".
- **chat-input-box** library row `d6a8f57b-…`: tool-level, gated. The locked
  page instance on `contact` is untouched and serving; an improvement sweep's
  08-11 attempt to REMOVE it was blocked by the lock (owner-review item
  `a4cd5dc8`, still open).
- **WireGuard tunnel** box↔cluster: up. `FACTS_URL` pins core-manager's
  ClusterIP `10.21.127.41` because box cluster-DNS is unresolved — the lane's
  one remaining fragility (survives pod restarts, not a Service recreate).

## 2. Next work, in order

1. **PLAN step 3 — `experience-planner`, once, for "site chat intake"**
   (PLAN §3): produces the approved `EXPERIENCE_PLAN` — journeys, promise
   ledger (per-IP rate limit, turn cap, spend ceiling, fail-closed to real
   contact details — already built+mutation-tested in Go, now stated as
   contract), data contracts (which per-site parameters a deployment needs),
   MVP cut. The four-critic council with honesty hard-veto is the gate.
   **Owner gave no explicit go in-session — ask, or treat "carry on" as the
   go if the queue is healthy.** Steps 5–6 (tool-deployer backend extension,
   tool-suggester wiring) cite this plan; do not start them first.
2. **Box DNS durability** — make cluster names resolve from the box (wg0
   `DNS=10.21.0.10` line is inoperative), then move `FACTS_URL` off the
   pinned ClusterIP. Read-only diagnosis first (`resolvectl status wg0`,
   `dig @10.21.0.10` over the tunnel); the tunnel-health recipe is in RUNBOOK.
3. **Bugs for pickup (either lane, or a fresh thread):** `bugs_open/275`
   (tool-suggester LIMIT 30 hides 38/68 tools — fix candidates ranked in the
   file) · `bugs_open/276` (section-level `requires-backend` components
   ungated in generic planning — VMB-010's section half; the council demanded
   this be tracked, it is now the concrete follow-up).
4. **Migration dry-run** (`./scripts/migration/run-migrations.sh`, no args) —
   per-session practice, and a roll just happened. It can take >2 min; run it
   in the background. Pending files seen 08-14 belonged to other threads —
   do not `--apply` unscoped.

## 3. Owner decisions OPEN (mine to surface, not to make)

1. **Step 3 go/no-go** (above — credits, one run).
2. **The blocked chat-box removal** (`lock_blocked_change`, `a4cd5dc8`,
   `needs_human_review`): keep the chat box on contact (dismiss the sweep's
   change) or release the lock. Recommendation: keep — it is the product demo.
3. *(Sibling lane's, listed for the whole-chain view — decide with them:)*
   **Stripe keys** (test mode; ⚠ via 047-base-configs terraform vars, NEVER
   kubectl — proven twice this week that a kubectl-added key dies at the next
   release) · **webhook exposure path** (their (a) = proxy over MY tunnel —
   exists, proven) · **pay-first vs build-first** (a copy migration, all five
   pages say pay-after-approval today; my 08-13 money-first recommendation
   stands for the automated path) · **"three or four days"** build-duration
   fact (re-attest or replace — the bot speaks the new value within 5 min of
   the register change) · **CTA buttons ink vs egg-gold** (their evening
   note awaits preference; their honest case is for ink).

## 4. Coordination (owner instruction, 2026-08-14 — standing)

Both webdesign threads stay mutually aware. Their lane =
`ai_site_selling_automation` ("the live web builder project"): £149
take-it-or-leave-it, kraft brand COMPLETE + palette served 08-14, working the
build-duration copy when last observed. **At session start and at natural
breaks: read their NOTES tail; write cross-lane notes there** (heading "from
the webdesign_uk_build_service lane"). A live session is invisible to git —
grep the active `~/.claude/projects/*/…jsonl` transcripts if it matters now.
Memory: `track-the-ai-site-selling-thread`.

## 5. Falsifiers / re-check before trusting this file

- Relay still healthy: box journal `grep -i facts` — failures log every 5
  min, silence is success; ask the bot the price (£149, pay-after-approval).
- Token still in secret AND pod (0-byte check + `TOKEN-IN-POD` probe —
  NOTES 08-15 has both commands).
- The sibling lane has moved — read their newest handoff, not §4 above.
- `10.21.127.41` still core-manager's ClusterIP.
- The 406 gate still in the live row (`load_library_tools` query carries
  `requires-backend`; params `["input_data.site_id"]`).
- Whether another thread picked up 275/276 (`who-owns.py` + live transcripts).
