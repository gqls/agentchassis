# HANDOFF 2026-08-15 — relay release-proof (PROVEN on a real roll), chat box is a gated library tool, OWNER DECISIONS IN (evening): step 3 GO + fix the improvement loop — SUPERSEDES HANDOFF_2026-08-13

**Start here cold.** Read order: this file → NOTES tail (2026-08-13 evening
through 2026-08-15) → PLAN_2026-08-11 §3 (the step you are about to run) →
the sibling lane's newest `ai_site_selling_automation/HANDOFF_*` (they move
fast and their surface is shared with ours).

**OWNER RULINGS 2026-08-15 (evening) — supersede §3's open list below:**
1. **Experience-planner run: GO.** Do it (work item 1 in §2).
2. **The improvement loop that tried to remove the chat box: CHECK AND FIX
   IT.** The box STAYS (lock stays on) until that fix lands — the
   `needs_human_review` row is answered by the fix, not by a dismissal.
   New work item 2 in §2.
3. Stripe keys: **later** — not now.
4. Webhook exposure: **proxy through the webdesign.uk box** (option (a),
   over this lane's tunnel) — decided "for now". Building the nginx block
   is sanctioned; it can be built and tested ahead of the keys.
5. Payment timing: **TAKE MONEY FIRST.** This triggers the sibling lane's
   pay-after-approval → upfront copy migration (relayed to their NOTES
   2026-08-15).
6. Build duration ("three or four days"): **UNDECIDED — do not change the
   copy.** Nuance worth preserving: the owner says he had already chosen
   "one or two", and is now considering letting the improvement loop run
   for a while before releasing the design; explicitly left open.
7. CTA buttons ink vs egg-gold: **UNDECIDED** — owner will check himself;
   do not act.

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

## 2. Next work, in order (reordered per the evening rulings)

1. **PLAN step 3 — `experience-planner`, once, for "site chat intake"
   (OWNER: GO).** PLAN §3: produces the approved `EXPERIENCE_PLAN` —
   journeys, promise ledger (per-IP rate limit, turn cap, spend ceiling,
   fail-closed to real contact details — already built+mutation-tested in
   Go, now stated as contract), data contracts (which per-site parameters a
   deployment needs), MVP cut. The four-critic council with honesty
   hard-veto is the gate. Steps 5–6 cite this plan; do not start them first.
   Dispatch mechanics: check the queue first (CLAUDE.md § Dispatching);
   no dispatch within ~300s of a chassis pod (re)start.
2. **Diagnose + fix the improvement-loop path that tried to REMOVE the
   locked chat box (OWNER: "check and fix"; the box stays until fixed).**
   Evidence to start from: work item `a4cd5dc8-ddf6-4d00-99ca-ab804d2ef6f9`
   (`lock_blocked_change`, `needs_human_review`) — "save_page_sections
   wanted to remove locked section chat-input-box on page contact",
   2026-08-11, i.e. an improvement/rebuild pass regenerated the section list
   WITHOUT the locked section and only the lock stopped the loss. Prior art
   to grep before filing: `bugs_closed/058` (rebuild path does not honour
   page_component locks — closed, this may be a NEW path), `bugs_open/268`
   (regeneration drops CTA URLs on unlocked components — CLOSED 08-14, read
   its record), `bugfix_149`'s "widening what REACHES a function breaks it
   unedited". Footprints: `save_page_sections` /
   `platform/orchestration/actions/lock_policy.go` / whatever improvement
   agent produced the change. **This is a cross-cutting root-cause claim →
   090 needs_diagnosis per CLAUDE.md's default** (the 090 trigger
   self-checks the queue). The fix's acceptance: an improvement pass over
   the contact page KEEPS the locked section in its proposed section list —
   not merely "the lock blocked it again".
3. **Webhook proxy on the box (OWNER: option (a) decided).** Small nginx
   `location` block on webdesign.uk receiving Stripe's public HTTPS webhook
   and `proxy_pass`ing over the tunnel to auth-service. Can be built + tested
   now (curl a dummy POST end to end); it goes live-meaningful only when the
   owner adds the keys (LATER — ruling 3). Coordinate with the sibling lane
   before touching their nginx surface; box nginx config is this lane's.
4. **Box DNS durability** — make cluster names resolve from the box (wg0
   `DNS=10.21.0.10` line is inoperative), then move `FACTS_URL` off the
   pinned ClusterIP. Read-only diagnosis first (`resolvectl status wg0`,
   `dig @10.21.0.10` over the tunnel); tunnel-health recipe in RUNBOOK.
5. **Bugs for pickup:** `bugs_open/275` (LIMIT 30 hides 38/68 tools) ·
   `bugs_open/276` (section-level requires-backend ungated — related to
   work item 2's class: both are "the loop rewrites what it should
   preserve/gate").
6. **Migration dry-run** (`./scripts/migration/run-migrations.sh`, no args)
   — per-session practice, and a roll just happened. >2 min, run in
   background; pending files seen 08-14 were other threads' — never
   `--apply` unscoped.

## 3. Owner decisions — mostly RULED (see the block at top); still open:

- **Stripe keys timing** — "a bit later"; when they come: via
  047-base-configs terraform vars, NEVER kubectl (a kubectl-added key dies
  at the next release — proven twice this week).
- **Build-duration copy** — UNDECIDED (ruling 6's nuance recorded at top);
  do not touch the "three or four days" wording, and note the owner is
  weighing letting the improvement loop run before releasing the design.
- **CTA buttons ink vs egg-gold** — owner checking himself; do not act.
- **Old subscription scaffold deprecation** — sibling's list, after the
  first real sale; unchanged.

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
