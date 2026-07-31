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

---

## CORRECTIONS + decisions, 2026-07-26 (P4 delivered)

**P4 is DONE and LIVE.** Steps 0, 1, 2 and 3 are complete and verified against
the deployed pages (72 of 73 checks, desktop + mobile — the one failure is
upstream, `bugs_open/083`). Three corrections to the plan as it stood:

> **CORRECTED — Step 1 did not need doing.** The plan and the handoff both
> treated the homepage provocation-card CTAs as broken. They were not: the feed
> already carried a correct `today.primary_cta.url`, and `provocation-card-loader`
> already sets both hrefs at runtime. Step 1 collapsed into Step 0, which means
> **the homepage was never edited** — so the `rebuild_policy='generic'` clobber
> risk that both documents spend paragraphs managing was avoided outright rather
> than mitigated. Journey A was verified live instead, and the
> `cta_names_unknown_destination` item it came from is closed. *What caught it:
> re-reading the live feed and the loader source before building, rather than
> trusting the handoff's summary of them.*

> **CORRECTED — council advisory 1 was already obsolete when it was written.**
> It asks for the "Enter today's Arena" CTA on provocations-index to be
> re-pointed at `/tools/arena/index.html`. It already points there. What that CTA
> actually needs is for the arena page to *deliver* — see the deferral below.

**Step 4 (tool-arena-interface) is DEFERRED to LATER, explicitly**, per the
plan's own §4 gate and the council mvp seat's advisory (defer outright rather
than attempt conditionally). The gate's precondition is now satisfied — the
source has been pulled and read — so this is a decision on the evidence, not on
its absence:

- `html_template` is 38,705 B; **`js_content` IS NULL**. The permanent
  "Loading…" at line 578 (`<div class="provocation-text" id="provocation-text">`)
  is template text that no JS was ever written to fill.
- Mount points are `id`-based, not `data-`-based, and are clean and sufficient:
  `#provocation-text`, `#provocation-day`, `#provocation-date-label`,
  `#take-input`, `#take-submit-btn`, `#your-take-display`, `#your-take-text`,
  `#floor-takes`, `#refile-btn`, `#char-count-label`, `#remix-root`,
  `#take-block`.
- The template carries its own inline `<script>` holding a **hardcoded
  provocation array** (`day: "Round 01"` …) meant to be selected by day-of-year.
- **Why defer rather than do it:** wiring the display to the feed is easy, but
  the page's real substance is a "take" submission flow (`#take-submit-btn`,
  `#floor-takes`) with **nothing behind it**. Making the text load while leaving
  submission going nowhere would replace a visibly-broken page with a
  convincingly-broken one — precisely the failure mode this workstream exists to
  remove. It needs its own scoping round.
- **One contained defect located while reading, for whoever picks it up:**
  `.reaction-chip.active[data-reaction="Delusional"]` sets `background`,
  `border-color` *and* `color` all to `var(--color-primary)`, so the active chip
  is invisible. The three sibling rules each use a distinct accent colour; only
  this one is wrong. It is the open `improve_tool` work item.

**P5 acceptance cannot run as specified, and that is a harness defect.**
`browserrunner/run_checks_action.go:200` waits `stepDelay = 300ms` between an
interaction step and its assertion. `gauntlet_position_flow` and
`gauntlet_defend_flow` assert on AI output measured at 8–23 s, so they would fail
a correct page. Two acceptable routes: extend the runner with a wait/poll
(`criteriaExpect` would need a timeout field), or replace those two checks with
assertions that are true at 300 ms. **Not acceptable:** making the page paint
optimistic placeholder text to satisfy them — that would make the checks pass
with the engine switched off. Meanwhile the journeys ARE verified, by
`p4_sources/verify_live.py` driving the deployed pages in Chromium.

---

## OWNER DIRECTION — 2026-07-28 (supersedes the blocking order in `bugs_open/131`)

Recorded verbatim in substance because it changes both the sequencing and the
long-run shape of this product, and neither is derivable from the code.

### Sequencing ruling: design FIRST, premise later

> *"I'm happy to improve the design before deciding on the whole premise as the
> latter may take longer."*

**This unblocks `bugs_open/131` items C, E and F**, which that file had gated
behind item H (the "why argue with an AI when Perplexity is free" question).
**H no longer blocks them.** Do the design work now; the premise question runs on
its own, longer clock.

### The premise question is not unanswered — it has a hypothesis

The owner's own answer to H, which should be treated as the working theory rather
than an open void:

> *"It might be that when there are more people and perhaps with different,
> predefined, interesting categories of provocations — maybe current affairs,
> celebrity, films, news, finance, as well as the type of provocations you are
> doing already — that it will seem lively and serve better purpose."*

So the differentiator from a chat window is **not** the argument itself; it is
**liveliness and breadth** — many people, many subjects, a place with something
going on. That reframes several things already built: the Gauntlet is currently
one provocation per day in one register (abstract/cultural), which is the
narrowest possible version of the thing the owner thinks would work.

**Implication for `131` item G** (recording a won verdict): it is not a vanity
feature, it is a *first step toward liveliness*. Treat it as such when scoping.

### Direction for LATER — not now, but do not lose it

1. **Provocation categories.** Predefined, interesting, named: current affairs,
   celebrity, films, news, finance — *alongside* the existing abstract register,
   not replacing it. This is the single most concrete lever named.
2. **Group opinion statistics, with graphs as the first step.** *"There could be
   value in the stats of group's opinions and decisions at a later stage, so
   graphs might be the first step for that."*
   **RAIL — this is the sharpest fabrication risk this workstream has faced.**
   A graph of group opinion on a site with no participants is a fabricated crowd
   with a chart drawn on it. The existing rule stands and applies with force: a
   number appears only if it is true by construction or measured. **Graphs come
   after participants, never before them.** [Recorded 2026-07-28 so nobody builds
   the chart first because it is the easy part.]
3. **Visitor-suggested provocations, to their own groups.** *"People suggesting
   their own provocations to their own groups might be another angle to explore."*
   Note this implies identity, groups and moderation — none of which exist. It is
   an exploration, not a queued feature.

### Priority, stated plainly

> *"But to start with, let's get what we have done, working well."*

**Fix what exists before extending it.** `bugs_open/131` B–F is the near-term
list; categories, graphs and groups are all downstream of that and of real
participants.

## OWNER DIRECTION — 2026-07-29 (the H ruling)

Asked to choose between the four H options (HANDOFF_2026-07-29 §7), the owner
ruled:

> *"3 leading to 2"* — **the distribution experiment first, feeding the arena
> thesis.** The owner does the distribution leg himself (the share card and
> the daily provocation are the travelling artefacts); real behaviour then
> informs the arena build (categories, one provocation per day per category,
> group views only after participants exist).

And one feature direction, in the owner's words:

> *"I think a (dated) personal history of your opinions might be a goldmine
> idea"* — **a dated personal ledger of what YOU argued and when.** Each
> played round already contains everything needed: the day's provocation, the
> position the visitor committed to, the verdict, the date. Accumulated, that
> is a diary of your opinions — where you stood on X, dated — which composes
> directly with the arena thesis (your stance beside the eventual communal
> split) and gives a visitor a reason to RETURN, which the single-round page
> never had.

**Design constraint for whoever builds it (do not lose this):** the current
round store is deliberately `sessionStorage` — tab-scoped, because a round is
20 minutes and should not outlive the tab. A HISTORY is a different artefact
with a different deliberate scope: client-side `localStorage` of the visitor's
own completed rounds (dated provocation + their position + verdict) is honest
by construction — facts of their own rounds, on their own device, no accounts,
no server identity — and is the right first form. Server-side history implies
identity and is a separate, later decision. The same rail applies as
everywhere: entries are created ONLY as the consequence of a real /defend
response, never synthesised or backfilled.

## OWNER DIRECTION 2026-07-31 — the share card: **option 3, staged via option 1**

Answering `HANDOFF_2026-07-30_A_share_card_and_the_full_debate.md`. Three
decisions, taken after seeing all three options mocked with a real round
(decision page: `https://claude.ai/code/artifact/2cb2166e-ba5e-406d-a6b2-aabfa5fb8d45`;
sources + measurement in `p4_sources/share_card_options_2026-07-30/`).

**1. The card links through to a record of the full debate (option 3), reached
through the exchange card (option 1).** Not either/or: step 1 ships an exchange
card needing no new URL, endpoint or migration; step 2 adds the per-round page
and the same card becomes the hook by gaining one line. Nothing in step 1 is
discarded by step 2.

**Why the choice was largely settled by measurement, not preference** — keep
this, because it will be re-litigated otherwise:

- a complete round averages **3,109 characters** (51 complete rounds on the
  island, 25–30 Jul; min 2,396, max 5,073);
- a 1200×630 card holds **~700 characters legibly** once a timeline downscales
  it (≈2.38× at X's ~504px in-timeline width — that width is `[ASSUMED]`, but the
  conclusion survives any plausible figure);
- therefore **the whole debate cannot go on a card** (it auto-fits at 11px,
  ~4.6px in a feed), and **two cards carry only ~46% of one round**.
- So options 1 and 2 were both *excerpting* strategies. Only option 3 carries
  the round. **Option 2 was rejected**: ~2× option 1's cost, still not a debate,
  and its real cost is a decision the owner makes every time he posts.

**2. Publication model: pressing share publishes the round.** One action —
downloads the PNG, makes that round public, and the card carries the permalink.
Rationale: no second control, and no dead links (every shared card has a page).
**Two requirements that follow, and neither is optional:**
- **the button must say plainly that sharing publishes** (owner's own words).
  It is therefore rewritten in step 2, once — which is why step 1 deliberately
  leaves the label alone rather than editing it twice;
- **consent is inferred from a press, so the wording IS the consent mechanism.**
  Publishing by default was explicitly rejected: it would put a stranger's
  writing online without asking and would have seeded the record from the 51
  existing **harness** rounds (`count(DISTINCT client_ip_hash)` = 1 over all 95
  rows — no stranger has ever argued here). **The public record starts empty.**

**3. Whether the private opinion ledger links out to a published round:
DEFERRED** until the record page exists and can be looked at. Reversible, one
line of markup; not worth guessing now.

### Status

- **Step 1 is LIVE** (2026-07-31, `js_content` only, served asset verified
  byte-identical, md5 `64dbfb8c…`). The card carries vonc's challenge, the
  visitor's defence and the ruling, auto-fitted (26px on the measured round).
  It carries **no per-round URL by design** — there is no page yet and a 404
  link is worse than none.
- ~~**Step 2 is not started.**~~ **Step 2 BACKEND IS BUILT (`28cf5ceb3`, PUB-004,
  council `a24a754b`).** `POST /publish` + `GET /round/:slug` + `published_at` and
  `public_slug` on `gauntlet_rounds`. Migration 276 **applied and ledgered on the
  island**; all four SQL statements PREPAREd against the live schema; router
  registration asserted by test (gin panics at *registration* on a route
  conflict, which would be a service that will not boot, found only after a
  swap); slug tests mutation-verified. Image **`v1.0.1216` built from committed
  HEAD, shipped to the island and binary-verified** (4 new strings present, 2
  positive controls present, negative control absent — the first grep returned
  all zeros *including* the controls, because binary grep needs `-a`; the control
  is the only reason that was caught).
  **Known limitation to state up front:** a client-fetched page cannot emit a
  per-round `og:image` (crawlers do not run JS), so a shared *link* previews
  with the site's generic card. This does not affect the actual plan, where the
  owner posts the PNG and it travels as an image.

### Step 2 — what remains, in order

1. **Swap the island to `v1.0.1216`.** One line in `/opt/island/docker-compose.yml`
   (currently pins `v1.0.1207` at line 50) plus `docker compose up -d tools-api`.
   The image is already loaded on the box and `1207` is still present, so rollback
   is a tag change and a restart. **BLOCKED: the permission classifier refused the
   compose edit** — a sensible place for a human check, so it needs the owner's go
   rather than a workaround. Until then the endpoints exist in the binary and are
   unreachable.
2. **The record page.** A new page at `/tools/gauntlet/round.html` — `pages` +
   `content_components` + `page_components` + `pages.sections`, following
   `sql_for_agents/275_oufe_tool_relevant_alternative.sql`, which is a complete
   worked template from the day before. **Its URL shape is a real decision** and
   the pretty form in the mock is not available: the site serves only exact
   `.html` paths with no directory index, so `vonc.com/r/<slug>` **cannot** be
   served. The honest options are `/tools/gauntlet/round.html?r=<slug>` (sits with
   the tool) or a shorter root-level `/round.html?r=<slug>`.
3. **The JS**: publish-on-share, the permalink line on the card, and the button
   rewording that announces publishing. Sequencing that matters — **the publish
   call must succeed BEFORE the card is drawn with a permalink**, or a failed
   publish yields a card carrying a URL that 404s. On failure: draw the card
   without the link and say so. That keeps the rail (no control changes state
   except as the consequence of a real API response).
4. **Verify end-to-end** by driving a real round, pressing share, and fetching the
   permalink — `p4_sources/drive_exchange_card_2026-07-31.py` is the harness to
   extend, and `~/.venvs/vonc_pw` is the interpreter that has playwright.
- ~~**Outstanding verification** on step 1: the card has not been pressed on the
  live page (the live driver was blocked by the permission classifier). The
  renderer is proven against a real round offline, and that the required DOM
  values are present at press time is `[INFERRED]` from the code, not observed.~~
  **CLOSED 2026-07-31, owner authorised the drive** (`ea9051dfc`). One real round
  on the live page — real `/round`, `/position`, `/defend`, pressed the real
  control, captured the real PNG (`p4_sources/live_card_2026-07-31.png`, 147,606
  bytes, 1200×630). The inputs **are** live at press time: the defence textarea
  still held all 198 characters and the challenge element all 469 after `/defend`
  returned. The `[INFERRED]` marker is discharged — it was right, but it was
  reading, and this was the failure that would have been invisible.
  - Incidental real-world evidence for the auto-fit: that round's challenge ran
    to **469 characters against a 305 average** and the card fitted it with no
    overlap. That is what the drawn-layout binary search buys over the character
    budget which produced the overlapping first mock.
  - **Misstep worth keeping**: the driver's first run printed `SKIP  PIL
    unavailable` for its three image assertions and still ended in **ALL LIVE
    CHECKS PASSED**. The summary was reporting on a rule that had not run. A
    missing Pillow is now a *failed* check; Pillow is installed in
    `~/.venvs/vonc_pw`. The three assertions were run separately against the
    captured PNG and pass (95/210 sampled rows marked, lowest 609). The hardened
    script has **not** been re-run end-to-end — that would drive another real
    round.
