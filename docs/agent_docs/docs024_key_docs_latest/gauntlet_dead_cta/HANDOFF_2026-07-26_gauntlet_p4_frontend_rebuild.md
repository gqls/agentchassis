# HANDOFF — gauntlet_dead_cta: P4 front-end rebuild (+ everything else outstanding)

**Written:** 2026-07-26 · **For:** a fresh thread continuing this workstream.
**Read this first, then `README_where_we_are.md` for the plain-prose history and
`NOTES_gauntlet_dead_cta.md` for the full evidence trail if you need it.**

## 0. Re-read CLAUDE.md before touching anything

This repo is worked concurrently by multiple sessions. Commit per task by
explicit pathspec, forward-only, check `git log`/`site_work_items` before
assuming you own a target, check `scripts/who-owns.py` before routing at an
existing bug. This handoff is a snapshot — re-verify anything load-bearing
against the live DB before acting on it, especially anything more than a day
old by the time you read this.

## 1. Where this stands — one paragraph

The backend (`tools-api`) is built, merged, deployed to a standalone VM ("the
island," Route B1), and **proven live with real AI-generated content** — a
full `/round → /position → /defend` round-trip through the public internet at
`https://tools.apis.uk`. The experience-planning council has **approved a
concrete build plan** for the front-end (first-ever full approval on this
workstream, full bar met: `approved`+`abstained:0`+`reviewers:5`). **Nothing
on the front-end has been built yet.** That's the job: rewrite the actual
gauntlet page (and three adjacent components) against the approved plan,
deliver it live, then run acceptance. This is P4 → P5 of the original PLAN.

## 2. Identity — fresh as of 2026-07-26

Site: `vonc.com` = `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`.

| page | url | rebuild_policy | page build_status | page_component id | component id | function | delivery mechanism |
|---|---|---|---|---|---|---|---|
| tool-gauntlet | `/tools/gauntlet/index.html` | **owned** | needs_rebuild (serves live anyway — the P1 landmine, see §7) | `1048b344-f1fa-44ea-b936-951bc7eafc59` | `5da50747-7936-4b8f-a66d-c1ea98919c75` | gauntlet-interface | **component-owned JS** (`content_components.js_content`) — section-editor + assemble-only rerender |
| provocations-index | `/provocations/index.html` | **owned** | deployed | `9cf1d2f5-13ef-4af0-b9b5-597f80e4fdeb` | `70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37` | provocations-archive-list | **runtime `js_snippets` loader** `provocations-archive-loader` (`1cba8e17-4797-48ab-a7e5-abaf982f2223`) — DIFFERENT mechanism, see §4 |
| tool-arena | `/tools/arena/index.html` | owned | deployed | `68ffc838-3cc0-45c6-ab78-10148d33bf7d` | `faa69bcc-f94e-4ca0-ac76-14a03da4807c` | tool-arena-interface | Step 4, likely deferred — see §5.4 |
| index | `/index.html` | **generic** ⚠ | deployed | `50cfd1e9-609d-46f8-bcf6-5383f6709978` | `6163ff14-9f94-4962-aa19-d2718eabdeb1` | provocation-card | content_data only (Step 1, CTA url fix) — `js_snippets` loader `provocation-card-loader` (`dac802c9-4d43-40b5-99e9-391c572f1109`) untouched by the plan |
| index | `/index.html` | **generic** ⚠ | deployed | `d9f25bd6-25ed-403d-8081-ea651b1f50b2` | `9304f14d-e19b-4ce1-b3fd-f6a315aec6ed` | lobby-grid | NOT touched — verify-not-regressed only (Journey D) |

⚠ **`index` is `rebuild_policy='generic'`, not owned.** If anything triggers a
generic rebuild of the homepage while you're mid-edit, your `content_data`
change to provocation-card can be clobbered — see
[[replan-clobbers-built-pages]] / `bugs_open/049/052/053`. Re-check
`p.build_status`/recent `site_work_items` on `index` immediately before
editing it, and prefer the smallest possible edit (a `content_data` URL fix,
not a template/js change) since Step 1 doesn't need more than that.

**Re-run this query yourself before starting** — these rows can drift:
```sql
SELECT p.name, p.url, p.rebuild_policy, p.build_status, pc.id AS page_component_id,
       cc.id AS component_id, cc.function, pc.build_status AS pc_build, cc.is_active
FROM pages p JOIN page_components pc ON pc.page_id=p.id
JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND cc.function IN ('gauntlet-interface','provocations-archive-list','provocation-card','tool-arena-interface','lobby-grid')
ORDER BY p.name;
```

## 3. THE APPROVED PLAN — verify current, then read it

```sql
SELECT is_current, length(body), created_at FROM doc_plans
WHERE subject_type='experience' AND subject_key='vonc-spark-game'
ORDER BY created_at DESC LIMIT 3;
-- the is_current=true row should still be the one below (13971 bytes,
-- created 2026-07-25 16:40:10Z). If it's DIFFERENT, someone re-planned —
-- read whatever is_current says now, not this snapshot.
SELECT body FROM doc_plans WHERE subject_type='experience'
  AND subject_key='vonc-spark-game' AND is_current=true;
```

Full text of the approved plan as of 2026-07-25 16:40:10Z (for reference —
**re-fetch live before building**, don't build from this copy if the query
above shows drift):

<details>
<summary>Click to expand the full EXPERIENCE_PLAN (14KB)</summary>

# EXPERIENCE_PLAN — the Spark daily-provocation game

## 1. Journeys

**Journey A — Homepage daily provocation → real destination**
1. Page `index`, control `.pc-btn-primary` inside `[data-component="provocation-card"]`. Action: click. Outcome: navigates to `today.primary_cta.url` from `/data/provocations.json`, which must resolve to a real page (`/tools/gauntlet/index.html`) — never `href=""` or `#`.
2. Page `index`, control `.pc-btn-secondary`. Action: click. Outcome: navigates to `today.secondary_cta.url` (real page, e.g. `/provocations/index.html`).

**Journey B — Archive entry → client-side detail (D2)**
1. Page `provocations-index`, control `.provocations-archive__item--linked` (a rendered archive row whose feed entry HAS `detail_body`). Action: click (or Enter key). Outcome: URL updates to `/provocations/index.html?entry=<slug>` (history.pushState, no full reload), and `.provocations-archive__detail` becomes visible, populated with that entry's real `title`, `date`, and full `detail_body` text — not a class toggle on an empty node.
2. Page `provocations-index`, direct load of `/provocations/index.html?entry=<slug>`. Outcome: on `DOMContentLoaded`, the loader reads `location.search`, fetches the feed, and pre-populates `.provocations-archive__detail` for that slug without requiring a click — genuinely deep-linkable.
3. An entry whose feed record has no `detail_body`: rendered with class `.provocations-archive__item--static`, no `href`, no `tabindex`, no click handler — visibly not offered as openable.

**Journey C — The Gauntlet: a real timed debate (D1)**
1. Page `tool-gauntlet`, control `[data-gi-enter-btn]` inside `[data-component="gauntlet-interface"]`. Action: click. Outcome: `POST {API_BASE}/api/v1/tools/gauntlet/round` fires; on 200, `round_id` is held in memory, `[data-gi-challenge-body]` fills with the returned `provocation.body`/`headline`, the 20-minute clock (`[data-gi-timer]`) starts for real, and the challenge panel (`.gi-challenge-panel`) scrolls into view. On network/503 failure: `[data-gi-status]` shows "the AI opponent is offline — try again later," the clock does NOT start, and no other behaviour fires (the old simulate-a-round click handler is removed).
2. Control `[data-gi-position-input]` (visitor types a Position) then `[data-gi-position-submit]`. Action: fill + click. Outcome: `POST .../position {round_id, position_text}`; on 200, `[data-gi-opponent-position]` and `[data-gi-opponent-challenge]` fill with the returned `counter_position`/`challenge`; the first objective `[data-gi-obj]` (position filed) is programmatically marked `.is-complete` — never by manual click.
3. Control `[data-gi-defence-input]` then `[data-gi-defence-submit]`, before the timer reaches 0. Action: fill + click. Outcome: `POST .../defend {round_id, defence_text}`; on 200, `[data-gi-verdict]` and `[data-gi-verdict-reasons]` fill with the returned `verdict`/`reasons`; the second objective (defence sent) marks complete immediately, and the third objective (verdict received before clock expiry) marks complete only if `remaining > 0` at receipt — an honest, event-bound completion, not a checkbox anyone can tick by hand.
4. Manual toggling of `[data-gi-obj]` by click/keyboard is REMOVED in this rebuild; objectives only change state as a side-effect of steps 2–3.

**Journey D — Arena entry from homepage lobby (already working, verify not regressed)**
1. Page `index`, control `.lobby-grid-section__card[tabindex="0"]` inside `[data-component="lobby-grid"]`. Action: click/Enter. Outcome: navigates to `entry.url` from `data.arena.cards[]` — already wired by `lobby-grid-loader`; this journey must NOT be broken by the Gauntlet/archive rebuilds.

## 2. Promise ledger

| CTA copy | Destination must deliver |
|---|---|
| Provocation card primary CTA (today.primary_cta.label) | Real page load of `/tools/gauntlet/index.html` with a workable Gauntlet (Journey C) |
| Provocation card secondary CTA | Real page load of stated destination (e.g. Archive), no `#` |
| Archive entry title/row | `.provocations-archive__detail` populated with that entry's real body, at a deep-linkable URL |
| "Enter the Gauntlet" / `[data-gi-enter-btn]` | A real `POST /round` round actually starts: provocation shown, clock running, or an honest offline message |
| Position submit | A genuine AI `counter_position` + `challenge` appear, not templated text |
| Defence submit | A genuine AI `verdict` + `reasons` appear before/at clock expiry |
| Lobby-grid Arena card | Navigates to the named provocation's own destination (`entry.url`), already live via lobby-grid-loader |
| "Enter today's Arena" (provocations-index CTA, currently misdirected to homepage) | Fixed to point at `/tools/arena/index.html`; that page must show real (not "Loading… DAY 0") content — gated per §4 Step 4 |

No CTA in this experience may resolve to `href="#"`, an empty string, or a fabricated stat.

## 3. Data contracts

`/data/provocations.json`, authored/committed by the content build pipeline at deploy time (NOT client-generated, NOT a runtime emitter — D2 confirms no daily emitter this round):

- `today.primary_cta.url` / `secondary_cta.url` — must be real page paths (fixes the current empty-href and phantom-destination defects). `today.stats[3]` — real feed-authored numbers only, never invented at build time.
- `archive.entries[]` — existing fields (`date,title,teaser,stat,url`) PLUS two NEW fields required for D2: `detail_body` (string, full text shown in the detail region) and `slug` (URL-safe id used in `?entry=slug`). Entries lacking `detail_body` are rendered non-openable (Journey B.3) — the loader must check for its presence, not assume it.
- `arena{}` — unchanged shape already consumed by `lobby-grid-loader` (`eyebrow,title,subtitle,cta_label,cta{label,url},cards[]{icon,tag,title,desc,stat,url}`). Not touched by this plan except verified present.

**Client-side-only (never written to the feed):**
- Gauntlet round state (`round_id`, opponent position/challenge, verdict/reasons) is per-session and ephemeral, held in page memory only, sourced live from `{API_BASE}/api/v1/tools/gauntlet/{round,position,defend}`. The `provocation` object the `/round` endpoint returns is the verbatim `today` object — the client must render it as-is, never re-derive or invent fields.
- The Gauntlet's ONLY on-page number a visitor reads as a metric is the objective-progress percentage (existing `[data-gi-pct]`/`[data-gi-fill]`), and its meaning must stay exactly: `percentage = (count of [data-gi-obj].is-complete / 3) * 100`, rounded, where each of the 3 completions is bound strictly to a real API response per Journey C (never to manual clicking, never to elapsed time alone). Label it "X% Complete" — not "Score," not "Win Rate." No leaderboard, no win/loss tally, no fabricated participant counts anywhere in this experience (per D1: no leaderboard this round).

## 4. MVP cut + LATER

**Step 0 — DATA (gate: nothing below may proceed until this returns 200 with real content).**
Commit `/data/provocations.json` with: `today.primary_cta.url`/`secondary_cta.url` resolved to real pages; `today.stats[3]` real; `archive.entries[]` each carrying real `detail_body` + `slug` for every entry meant to be openable; `arena{}` intact. Verify via a direct GET returning HTTP 200, valid JSON, and non-placeholder strings in every field above — do not trust a prior "feed is live" claim without re-checking at build time.

**Step 1 — Homepage CTA fix (gate: Step 0 done).**
Re-point `today.primary_cta`/`secondary_cta` per Promise Ledger. Resolves the `cta_names_unknown_destination` / `dead_control` items scoped to `provocation-card` only — the unrelated `brief-explanation` dead controls ("Get Started"/"Learn More" → `#`) are a DIFFERENT component and OUT of this experience's scope; do not silently fix or claim resolved here.

**Step 2 — provocations-archive-loader modification (gate: Step 0 done; feed carries `detail_body`/`slug`).**
Add reads for `entry.detail_body` and `entry.slug`; add a `.provocations-archive__detail` region (new markup, built this round) that the loader populates on click and on initial `location.search` parse; add `--linked` class + interactivity only when `detail_body` is present, `--static` class and no interactivity otherwise. This is a genuine code change to the existing loader — the current source has no such read path and must not be assumed to already split entries.

**Step 3 — Gauntlet rebuild via tool pipeline (gate: Step 0 done; owned_page_review satisfied by tool builder, not generic builder).**
Implement Journey C in full: remove the manual-toggle objective handler and the enter-button's simulate-a-round side effects; wire the three exact fetches (`/round`, `/position`, `/defend`) against `API_BASE = https://tools.apis.uk`; add new DOM per §1/§5 selectors; implement the honest 503-offline path (never a 502-shaped guess, since Cloudflare rewrites 502→503 before the browser sees it); keep the existing real timer/progress-bar code paths, now driven by API events instead of clicks.

**Step 4 — tool-arena-interface fetch wiring (gate: Step 0 done AND the component's actual `html_template`/`js_content` has been pulled into build context and quoted — NOT available in this plan's source review).**
Because no source for `tool-arena-interface` was available to verify selectors, this plan pins ONLY the outer contract: the component must fetch `/data/provocations.json` and stop showing "Loading… DAY 0" forever. Concrete inner selectors are NOT specified here and must not be invented at build time; the build round must fetch and quote the real template/js before writing them. If that source cannot be obtained this round, Step 4 moves to LATER and is excluded from this round's acceptance gate — it is not silently marked resolved.

**LATER (explicitly out of this round):**
Daily emitter automation and static per-provocation pages (superseded by D2 this round); Gauntlet leaderboard, human-vs-human, persisted outcomes (excluded by D1); full `tool-arena-interface` inner-selector rebuild (pending Step 4 gate); unrelated open items — contact form delivery, archetype-page misdirected/unresolved secondary CTAs, `platform-comparison`/`/enter` phantom link, taster-quiz `cta_primary_url`, hardcoded colors, 16-page header/footer rerender, orphan blog post — all out of scope for this experience.

## 5. Acceptance criteria

```criteria
[
  {
    "profiles": ["desktop", "mobile"],
    "container": "[data-component=\"provocations-archive-list\"]",
    "checks": [
      {"id": "archive_list_exists", "type": "selector_exists", "selector": ".provocations-archive__list"},
      {"id": "archive_feed_loads", "type": "asset_loads", "path": "/data/provocations.json"},
      {"id": "archive_linked_item_exists", "type": "selector_exists", "selector": ".provocations-archive__item--linked"},
      {"id": "archive_detail_region_exists", "type": "selector_exists", "selector": ".provocations-archive__detail"},
      {"id": "archive_open_detail_populates", "type": "interaction", "steps": [{"action": "click", "selector": ".provocations-archive__item--linked"}], "expect": {"selector": ".provocations-archive__detail", "text_matches": ".+"}},
      {"id": "archive_no_overflow", "type": "no_horizontal_overflow", "profiles": ["mobile"]}
    ]
  },
  {
    "profiles": ["desktop", "mobile"],
    "container": "[data-component=\"gauntlet-interface\"]",
    "checks": [
      {"id": "gauntlet_page_ok", "type": "page_status_ok"},
      {"id": "gauntlet_enter_exists", "type": "selector_exists", "selector": "[data-gi-enter-btn]"},
      {"id": "gauntlet_status_exists", "type": "selector_exists", "selector": "[data-gi-status]"},
      {"id": "gauntlet_round_starts", "type": "interaction", "steps": [{"action": "click", "selector": "[data-gi-enter-btn]"}], "expect": {"selector": "[data-gi-challenge-body]", "text_matches": ".+"}},
      {"id": "gauntlet_position_input_exists", "type": "selector_exists", "selector": "[data-gi-position-input]"},
      {"id": "gauntlet_position_flow", "type": "interaction", "steps": [{"action": "fill", "selector": "[data-gi-position-input]", "value": "Daily provocations create more anxiety than growth."}, {"action": "click", "selector": "[data-gi-position-submit]"}], "expect": {"selector": "[data-gi-opponent-position]", "text_matches": ".+"}},
      {"id": "gauntlet_objective1_marked", "type": "selector_exists", "selector": "[data-gi-obj].is-complete"},
      {"id": "gauntlet_defence_input_exists", "type": "selector_exists", "selector": "[data-gi-defence-input]"},
      {"id": "gauntlet_defend_flow", "type": "interaction", "steps": [{"action": "fill", "selector": "[data-gi-defence-input]", "value": "My position holds because reflection requires discomfort."}, {"action": "click", "selector": "[data-gi-defence-submit]"}], "expect": {"selector": "[data-gi-verdict]", "text_matches": ".+"}},
      {"id": "gauntlet_progress_pct_exists", "type": "selector_exists", "selector": "[data-gi-pct]"},
      {"id": "gauntlet_no_overflow", "type": "no_horizontal_overflow", "profiles": ["mobile"]}
    ]
  },
  {
    "profiles": ["desktop", "mobile"],
    "container": "[data-component=\"provocation-card\"]",
    "checks": [
      {"id": "pc_feed_loads", "type": "asset_loads", "path": "/data/provocations.json"},
      {"id": "pc_primary_cta_exists", "type": "selector_exists", "selector": ".pc-btn-primary"},
      {"id": "pc_secondary_cta_exists", "type": "selector_exists", "selector": ".pc-btn-secondary"}
    ]
  },
  {
    "profiles": ["desktop"],
    "container": "[data-component=\"tool-arena-interface\"]",
    "checks": [
      {"id": "arena_container_exists", "type": "selector_exists", "selector": "[data-component=\"tool-arena-interface\"]"},
      {"id": "arena_page_ok", "type": "page_status_ok"},
      {"id": "arena_feed_loads", "type": "asset_loads", "path": "/data/provocations.json"}
    ]
  }
]
```
<!-- END EXPERIENCE_PLAN -->

</details>

## 4. Two DIFFERENT delivery mechanisms — do not conflate them

**A. `gauntlet-interface` (Step 3) — component-owned JS, `content_components.js_content`.**
Proven path this workstream already used successfully (2026-07-22, P2):
1. Dollar-quoted `UPDATE content_components SET html_template=…, js_content=…,
   input_schema=… WHERE id='5da50747-7936-4b8f-a66d-c1ea98919c75'` — **back up
   the row first** (`SELECT * FROM content_components WHERE id=…` into a file).
2. Deliver via section-editor `content_edit`, the **DIRECT orchestrator
   envelope** (086-pattern: `spawn_agent`+`call_agent`, `action=process`) —
   NOT the bare `action=orchestrate` envelope (049b), which has silently
   failed to ingest here before. `scripts/deliver_gauntlet_section_edit.sh`
   is the proven shape (adapt its payload for the new template/js).
3. **`apply_section_edit` does NOT republish `js_content`** — the JS asset
   stays stale until you follow with an assemble-only rerender:
   `scripts/republish_gauntlet_js.sh` (uses `rerender_single_page`, no
   `reason` field — assemble-only, does not re-select or re-plan sections).
4. **`section-editor` leaves `pc.build_status='approved'`** — an assemble
   path may drop non-'deployed' components. Set it back:
   `UPDATE page_components SET build_status='deployed' WHERE id='1048b344-f1fa-44ea-b936-951bc7eafc59';`
5. Verify live by curl, matching a string YOUR edit created (a new selector
   like `[data-gi-obj]` or the literal `https://tools.apis.uk` in the served
   JS) — never a generic CSS property (the `bugs_closed/024`/`046` trap).
   Cache-bust the request.

**B. `provocations-archive-loader` + `provocation-card-loader` (Steps 1–2) —
runtime `js_snippets`, a DIFFERENT table and (as far as this session verified)
a DIFFERENT delivery path. NOT exercised this session — you'll need to
confirm the mechanism before relying on it.**
- The row: `UPDATE js_snippets SET js_content=… WHERE id='1cba8e17-4797-48ab-a7e5-abaf982f2223'`
  (provocations-archive-loader) — back up first, same discipline as above.
- There is a registered action `render_js_snippets_for_site` (see
  `platform/orchestration/actions/registry.go:807`,
  `render_js_snippets_for_site_action.go`) that appears purpose-built for
  publishing a `js_snippets` change to a site's runtime bundle — **read that
  action's Go source and any agent/trigger that calls it before assuming
  it's the right (or only) path.** Do not guess; this table's publish
  mechanism hasn't been proven live by this workstream the way `apply_section_edit`
  + assemble-rerender has for `content_components`.
- Step 1 (provocation-card) is simpler: it's a `content_data`/feed change
  (`today.primary_cta.url` in `/data/provocations.json`, Step 0), not a
  template/js edit — the loader itself (`provocation-card-loader`) is
  explicitly untouched by the plan. Don't confuse "fix the CTA url in the
  data" with "edit the loader."

## 5. The council's own advisory notes — fold these into the build, they're free

The approval had 4 objections, all advisory (medium/low severity, none
gating) — the council still wants them addressed even though it didn't block
on them:

1. **(journeys, medium)** — no §1 journey names the "Enter today's Arena" CTA
   fix (provocations-index → `/tools/arena/index.html`), so nothing verifies
   it. Add the journey step + a click-through check when you do Step 1's
   sibling fix (this CTA is on `provocations-index`, not `index` — check
   which component owns it before editing).
2. **(feasibility, medium)** — add an explicit pre-flight check that
   `https://tools.apis.uk` is reachable before attempting Step 3's client
   work, rather than assuming it (it's a genuine external dependency — a
   quick health check at build/verify time, not a runtime gate on every
   visitor).
3. **(mvp, medium)** — Step 4 (tool-arena-interface) is scope creep beyond
   the core loop; the plan itself already gates it on source availability
   (§4 Step 4) — if you can't pull the real template/js this round, defer it
   to LATER outright, don't attempt it conditionally.
4. **(contracts, medium×2 + low)** — three acceptance-criteria gaps: no check
   for the deep-link `?entry=<slug>` direct-load path (Journey B.2); no
   selector assertion for `[data-gi-opponent-challenge]` /
   `[data-gi-verdict-reasons]` (only `-opponent-position`/`-verdict` are
   tested in §5's JSON — the plan's own criteria block is short two
   selectors relative to what Journey C promises); no check that
   `.provocations-archive__item--static` entries actually lack
   interactivity. Extend §5's criteria block with these three checks when
   you write the real acceptance run (P5), don't just build to the letter of
   what's already in the JSON.

## 6. The exact, VERIFIED API contract (first-hand — I built and tested this)

```
POST https://tools.apis.uk/api/v1/tools/gauntlet/round     Origin: https://vonc.com
  -> 200 {"round_id":"<uuid>","provocation":{"eyebrow":"...","headline":"...","body":"...",
          "primary_cta":{"label":"...","url":"..."},"secondary_cta":{"label":"...","url":"..."},
          "stats":[{"value":"...","label":"..."}]}}
  (provocation is the VERBATIM 'today' object from /data/provocations.json — render as-is)

POST .../position   {"round_id":"...","position_text":"..."}
  -> 200 {"counter_position":"...","challenge":"..."}   (exactly these two string fields)

POST .../defend      {"round_id":"...","defence_text":"..."}
  -> 200 {"verdict":"...","reasons":"..."}                (exactly these two string fields)

Errors (all JSON, all confirmed live from the public internet):
  403 {"error":"origin not allowed"}   — Origin header not in the sites table's deployed domains
  404 {"error":"round not found"}      — bad/missing round_id
  503 {"error":"gauntlet opponent unavailable"} / {"error":"gauntlet judge unavailable"}
                                        — engine offline; THIS is the status to build the
                                          honest degraded-mode UI against, NOT 502 —
                                          Cloudflare replaces raw origin-502 bodies with its
                                          own error page, so a 502-shaped client assumption
                                          will silently break in production even though it
                                          might look fine testing against the origin directly
  400  invalid/missing request body
  413  oversized input (>2000 chars per field)
  429  rate-limited (per-IP token bucket)
Preflight: OPTIONS on any of the 3 endpoints -> 204 with CORS headers, no body.
```

## 7. Landmines (do not relearn these)

- **`apply_section_edit` does NOT run `collectJSAssets`** — a template/HTML
  edit goes live, the JS asset does not, until an assemble-only rerender
  follows. This is the single most costly mistake to make twice on this page.
- **The bare `action=orchestrate` envelope (049b) has silently failed to
  ingest** on this specific page before (kubectl-run stdin race, no error, no
  orchestration row, no work item). Use the direct orchestrator envelope
  pattern.
- **`section-editor` drifts `pc.build_status` to `'approved'`** — reset to
  `'deployed'` or a later assemble path may drop the component.
- **`tool-gauntlet`'s `p.build_status` reads `needs_rebuild` while the page
  serves live** (the exact defect P1's `dead_controls` fix — commit
  `01e18019a`, now CONFIRMED live in the current chassis pod, v1.0.1165+ —
  was built to catch). Don't be confused by the page-level flag; the
  component (`pc.build_status`) is what actually serves.
- **`index` is `rebuild_policy='generic'`** (see §2 warning) — a concurrent
  generic rebuild can clobber your Step 1 `content_data` edit. Check for
  recent/in-flight `site_work_items` on `index` before editing, and re-verify
  after.
- **`/data/provocations.json` currently carries FABRICATED stats** from an
  earlier build (June-dated: "1,284 Positions Filed", "3h 12m Until Close",
  "62% Disagree" in `today.stats`) — `tools-api`'s `/round` endpoint passes
  these through verbatim since it just proxies the file. **Step 0 of the
  approved plan requires real feed-authored numbers only** — this file needs
  regenerating (or at minimum, the fabricated `stats` array needs removing/
  fixing) as part of Step 0, not left as-is. This has been flagged
  repeatedly across this workstream's docs; it has not yet been fixed.
- **`js_snippets` publish mechanism is UNPROVEN this session** — see §4B.
  Don't assume it works like `content_components` delivery; read
  `render_js_snippets_for_site_action.go` first.
- **Compose/load_context in the experience-planner read NO prior
  council_report/doc_plans history.** If you ever need to re-fire the
  092 trigger (e.g. the approved plan needs a real design change, not just
  an implementation detail), a bare re-fire starts blind. Fold the specific
  reason for re-planning into the compose Decisions channel first (see
  migrations 197/207/209 for the pattern — a `jsonb_set`+`replace()` on
  `agent_definitions` type `experience-planner`, anchored on a stable string
  in the current live prompt). **Common mistake**: `to_jsonb(replace(...))`
  needs its OWN closing `)` before the `, true)` that closes `jsonb_set` —
  hit this twice (migrations 207, 209) before it stuck.
- **All 5 experience-planner reviewer seats share ONE `max_tokens` (raised to
  16000 by migration 208 after a live truncation)** — if a future round
  truncates again on a big enough plan, that cap is the first thing to check
  (`default_config->'workflow'->'steps'->'review_<seat>'->'config'->'ai_service'->>'max_tokens'`
  for each of journeys/feasibility/honesty/mvp/contracts).
- **Migration ledger**: this workstream applied 196/197/199/200/201/202
  (backend/machinery) and 207/208/209 (experience-planner config) — all
  ledgered in `schema_migrations` on the CLUSTER db. Migration 198
  (gauntlet_rounds table) applies to the **ISLAND's own Postgres**, ledgered
  separately in `island_migrations` there — do not confuse the two DBs.
- **Applying just YOUR OWN pending migration**: `run-migrations.sh` (no
  flags) is a safe dry-run PROBE of everything pending; its `--apply` runs
  the WHOLE pending batch including other threads' files. To apply only
  yours: probe (confirms clean), apply by hand
  (`kubectl … psql -v ON_ERROR_STOP=1 < your_file.sql` — the file's own
  `BEGIN`/`COMMIT`/`DO $$ RAISE EXCEPTION` guards do the real verification),
  then `run-migrations.sh --record-only <file> --note '<what you verified>'`
  to ledger it.
- **Two council verdicts for the tools-api fixes were APPROVED but never got
  a `Council-Reviewed:` trailer added to their commits** (forward-only rules
  out amending after the fact) — see NOTES 2026-07-26 entry for the mapping.
  Not actionable, just don't be confused if the 098 coverage report lists
  `258444df1`/`76e9c44d2`/`b498df16b` as unreviewed; they were reviewed.

## 8. What's genuinely NOT this workstream's job (don't scope-creep into it)

- **`bugs_open/071`'s residuals** (pod-template `spawned-by` labels; topic-age
  tombstone gating) — this workstream diagnosed and fixed the live-breaking
  part (the cleanup cron's guard); a concurrent session appears to already be
  carrying the residuals forward (per a same-day memory update from another
  thread mentioning "071 residuals SHIPPED 07-25 pm, dedicated session"). Check
  `bugs_open/071`'s current file state before touching it.
- **`bugs_open/003`'s F2/F3** (durable retry machinery) — separate, larger
  workstream; this workstream only contributed evidence/sightings.
- **Everything in the approved plan's own LATER list** (§4 above) — daily
  emitter automation, Gauntlet leaderboard/human-vs-human, unrelated dead
  controls on other components, the 16-page header/footer rerender, etc.
  Explicitly out of scope; don't fix them opportunistically while in here.

## 9. Suggested order of work

1. Re-verify §2's table and §3's plan freshness (both can have drifted).
2. Step 0: fix `/data/provocations.json` (real CTAs, real stats, `detail_body`
   + `slug` on archive entries) — everything else gates on this.
3. Step 1: provocation-card CTA urls (content_data only).
4. Step 3: gauntlet-interface rebuild (the flagship piece — biggest single
   change, proven delivery path, do it while that context is freshest).
5. Step 2: provocations-archive-loader (js_snippets — resolve the delivery
   mechanism from §4B before starting, don't discover it live).
6. Step 4: tool-arena-interface — pull real source first; defer to LATER if
   unavailable, per the plan's own gate and the mvp seat's objection.
7. Fold in §5's four advisory notes as you go, not as an afterthought.
8. P5: run the plan's own §5 acceptance criteria (extended per §5.4 above) as
   a real Tier-4 journey acceptance pass — browser-runner against the LIVE
   page, not a synthetic check. Then `claimscan` for fabrication, and a
   `dead_controls` re-check on `tool-gauntlet` specifically (P1's fix is
   live in the pod but has never been behaviourally re-verified against this
   page post-rebuild — see §7).
9. Update the standing five as you go (NOTES at least once, README at
   natural checkpoints, a new dated SUMMARY only if the state genuinely
   moves — not on a clock).

## 10. Credit / owner visibility

No further paid council/designer/implementer runs are obviously needed for
P4 (it's DB content + config delivery, not new Go code) — but if you DO touch
`platform/`, `internal/`, or `pkg/` Go (e.g. while investigating the
`js_snippets` publish path, or if it turns out a new action is needed), route
it through the council gate per CLAUDE.md's standing norm, and report the
spend as it happens per this workstream's established "blanket go, report
each spend" policy. The owner's hard gate for the *backend* was the PR merge
(already done); for this front-end phase, treat "the live page now genuinely
works, verified by a real acceptance pass" as the natural point to report
back, the same way the backend milestone was reported.
