# PLAN — vonc gauntlet dead CTAs + the generic dead-control detector hole

**Started:** 2026-07-22 · **Branch:** `085_debug_and_feature_loops` · **Owner-directed.**
Site: vonc.com (`9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`), page `tool-gauntlet`
(`ecb637c1-845f-46bf-b174-9c92a43f9586`), component `gauntlet-interface`
(`content_components` `5da50747-7936-4b8f-a66d-c1ea98919c75`; page_component
`1048b344-f1fa-44ea-b936-951bc7eafc59`).

## Symptom (owner report)
`vonc.com/tools/gauntlet/index.html#` — "the link doesn't work … not sure we have a
working gauntlet yet." Reproduced: the page serves 200 and its widget (timer,
checkable objectives, progress, animated stat counters) is fully wired and works.
The **only** broken controls are the two hero buttons:
- **Enter the Gauntlet** → `<a href="#" data-gi-enter-btn>` → clicking appends `#`
  (the URL the owner pasted). Dead.
- **Preview Rules** → `<a href="#">`. Dead.

## Root cause (evidenced)
1. Both `href="#"` are **hardcoded in the component's `html_template`**. The
   `input_schema` parameterises the button *labels* (`cta_enter_label`,
   `cta_preview_label`) but has **no URL field** — dead by construction; no
   content_data/resolver/render could give them a destination. The primary CTA
   even carries a `data-gi-enter-btn` hook the JS never binds.
2. The stats (`12,847` competitors / `94,210` completed / `38%` win rate / `7` day
   streak) and the 5-name leaderboard (AxonFury, ZeroRush, NexVoid, Skorch,
   Proxima) are **fabricated placeholders** — stats slotted from `static` fallbacks,
   leaderboard hardcoded in the template. There is no real gauntlet behind
   "Enter the Gauntlet".

## Why nothing caught it (the generic gap)
- `misdirected_cta` (live, enabled) scans the page but **skips `href="#"`**:
  `ClassifyLinkScope("#")` = `LinkScopeAnchor`, and the check only inspects
  page/empty scopes (`check_misdirected_cta.go:234`). Documented in `bugs_open/023`.
- `dead_controls` (live in binary, enabled on completeness-discovery-agent) is the
  detector **built for exactly this** — its header names *"the vonc gauntlet … both
  href="#", live for weeks"* as its proof case. But it **never fires on the
  gauntlet**, because its query filtered `p.build_status = 'deployed'`
  (`check_dead_controls.go:65`) and the gauntlet page is `build_status =
  'needs_rebuild'` while serving 200 (`pc.build_status='deployed'`). It is one of
  ~34 fleet pages that serve live as `needs_rebuild` (`bugs_open/049/052/053`).
  **The detector missed its own proof case.**

## Owner decisions (2026-07-22)
1. **Make the gauntlet genuinely work** (not a mock): wire the CTAs to real on-page
   behaviour; strip the fabricated stats + leaderboard so nothing is simulated.
2. **Fix the generic detector** (`dead_controls` build_status predicate) so any new
   site's live-but-needs_rebuild tool page gets its dead CTAs flagged — via the
   council gate, coordinating with the owning thread.
3. **Backend:** deliberately NOT building a full competitive-gaming backend (accounts
   + live leaderboard) now — no real competitors exist, so a live leaderboard would
   be a *new* fabrication. Reuse the existing form-delivery backend (bugfix 006) for
   the one real action ("file your Position" → delivered to the owner). Full backend
   is an explicit follow-on once real traffic exists.
4. **Council directive** (owner, verbatim intent): *we shouldn't be creating
   placeholders like that, that don't work.* Carried to the council as the rationale
   of the detector-fix submission (the fix enacts the directive).

## Phases
- **P1 — generic detector fix (Go, image-gated, council-gated).** `check_dead_controls.go`:
  gate liveness on `pc.build_status='deployed'` (the component that actually serves),
  not the drifting page-level `p.build_status`. DONE (edit + local build green);
  council submission carrying the owner directive next; commit on APPROVED with
  trailer; ship on next chassis image; verify the gauntlet is flagged post-roll.
- **P2 — gauntlet component honesty + function (config/content, live via section-editor).**
  Rewrite `gauntlet-interface` template + js_content + input_schema: CTAs do real
  on-page things; remove fabricated stats/leaderboard. Because the page is
  `rebuild_policy='owned'`, deliver ONLY via `section-editor`/`apply_section_edit`
  (generic rerender is forbidden — `bugs_closed/024`). Verify live by curl (match the
  component's OWN rule, never a generic property — the 024/046 trap).
- **P3 — real action (optional, owner-gated).** "File your Position" submits via the
  existing contact/lead form delivery. Modest, real, honest. Follow-on.

## Coordination
Dead-control detection is guard-rail-3 of the experience loop, actively owned by the
`bugs_open/054` (chrome dead-control) / `cta_link_integrity` (`bugs_open/023`) threads.
Do NOT fork: the P1 fix goes through the council gate; the finding is contributed into
their record. `who-owns.py 054` confirms active ownership (cqls).

---

## 2026-07-23 — PHASE 2: the real build (owner-approved plan; supersedes "P2/P3 optional" above)

The 2026-07-22 fix was CORRECTED as cosmetic (see NOTES + WRONG_CALLS: buttons wired
to invisible-in-context effects; checkboxes theatre). Owner directed the real build.

**Owner decisions (2026-07-23, all on record in the approved session plan):**
- **D-A. Debate opponent**: file a Position on today's provocation → AI files a real
  opposing Position + challenge → defend on the clock → honest AI verdict with
  reasons. Objectives = real self-checking steps. "AI competitor" labelling while no
  human traffic. Degraded mode honest, never a mock.
- **D-B. Backend via the feature-builder** (first fire of its implementer = platform
  milestone B4). Work item `capability_gap:tools-api-gauntlet-debate`
  (`9ed684bc-864a-4aa1-b17a-7ed061e08f2a`); designer corr `cff7ff61-…`.
- **D-C. Experience loop unstuck**: contracts-rule greenfield split (migration 196).
  New requirement injected via compose-prompt decisions block (migration 197,
  D1-REVISED). Re-plan fired: corr `4d3d89fa-…`.
- **D-D. Architecture**: engine in-cluster (`tools-api`, ClusterIP, no ingress);
  public path = Cloudflare (`<SUB>.apis.uk`, owner names) → Tunnel → bastion VM
  (Caddy allowlist `/api/v1/tools/*` only, caps, rate limit, no k8s creds) →
  WireGuard → service. Drafts in `infra/`. Sites stay static.
- **D-E. Credit policy**: blanket go for this workstream's paid runs (designer,
  implementer + shakeout, contingency 092 re-fire); each spend reported as it
  happens. Owner's hard gate = the PR merge.

**API contract (FIXED — pinned in 197 and the capability_gap spec; do not drift):**
`POST /api/v1/tools/gauntlet/round` → `{round_id, provocation:{headline, body}}`
(provocation fetched server-side from the calling site's live feed);
`POST …/position {round_id, position_text}` → `{counter_position, challenge}`;
`POST …/defend {round_id, defence_text}` → `{verdict, reasons}`.
Caps ≤2000 chars; CORS from sites table; per-IP rate limit; LLM via aiservice.

**Sequence + status:** P0 done (196 applied+ledgered) · P1 fired (197 applied,
092 corr 4d3d89fa in flight — accept only approved + abstained:0 + reviewers:5) ·
P2 designer in flight (corr cff7ff61) → implementer B4 on approval → owner merges
PR → image → migration → deploy · P3 blocked on owner infra tasks
(infra/README_bastion_exposure.md) · P4 front-end via section-editor +
assemble-only JS republish · P5 Tier-4 journey acceptance + claimscan +
dead_controls re-check · P6 docs/close-out.

---

## 2026-07-24/25 — CORRECTIONS to D-D and the phase sequence above (status as of 2026-07-25 evening)

> **CORRECTED: D-D's architecture (in-cluster tools-api behind a bastion +
> WireGuard) was never built.** A concurrent thread re-decided the exposure
> route on 2026-07-24: **Route B1, a standalone Mythic Beasts VM ("the
> island")** — Cloudflare Tunnel → Caddy path-allowlist → tools-api container
> → the island's OWN Postgres, with the production cluster appearing NOWHERE
> in the public path (stronger isolation than the bastion draft, and the
> WireGuard-to-cluster premise was separately refuted — masquerade defeats
> ipBlock policies). `infra/README_bastion_exposure.md` and the WireGuard
> drafts are DEAD; as-built truth is `infra/island/RUNBOOK_island.md`. Public
> URL: `https://tools.apis.uk`. P3's *goal* (a secured public path, no k8s
> creds exposed) was achieved by a different, better structural route than
> planned — record the destination reached, not the route drafted.

**Actual status, P0 through the experience re-plan (all DONE as of
2026-07-25 ~16:45):**
- P0/P1 (196/197): DONE, as planned.
- P2 (B4 designer→implementer): DONE. Designer converged corr `c379f7b7`
  (3 council rounds). Implementer's first-ever complete run (`af286d2c`)
  produced **PR #3**, merged by the owner 2026-07-25 09:19Z. Cost:
  the B4 shakeout also surfaced and fixed 5 durable platform bugs
  (bugs_closed/065/067, bugs_open/071 with residuals, migrations 199-202)
  — see `fixloop_eg_dartsonline/NOTES_running_feature_builder.md`.
- Build+deploy: DONE. Image built from the 086 branch (tools-api source
  carried onto it, verbatim + 5 post-merge fixes found by deploy-and-smoke —
  none catchable by the implementer's stage gates). Deployed to the island;
  DB prepped (minimal `sites` table + corrected migration 198, ledgered in
  `island_migrations`, NOT clients_db).
- **Real liveness proven** 2026-07-25 ~15:00Z: a full `/round`→`/position`→
  `/defend` round-trip through the public internet, genuine AI-generated
  content, two complete rounds persisted with real verdicts ("opponent
  wins" both times — honest judging, not a pushover).
- P3 (exposure): DONE via Route B1 (see correction above), not the drafted
  bastion.
- Experience re-plan: DONE. Carried the liveness evidence into the
  planner's compose channel (migration 207); first re-fire ran 5 genuine
  REVISE rounds then hit its round cap and escalated (a designed circuit
  breaker — surfaced a real platform defect along the way, reviewer-seat
  token truncation, fixed in migration 208); folded the escalation's own
  named objections back into the compose channel (migration 209); second
  re-fire converged in ONE round. **APPROVED 2026-07-25 ~16:45Z**, full bar
  (`approved`+`abstained:0`+`reviewers:5`+`unreadable:0`), corr `5316e79c`.
  `is_current` doc_plan for `vonc-spark-game`, 13971 bytes, is now the build
  target for P4.

**NEXT = P4** (front-end rebuild against the approved plan) → **P5**
(Tier-4 journey acceptance + claimscan + dead_controls re-check) → **P6**
(close-out).
