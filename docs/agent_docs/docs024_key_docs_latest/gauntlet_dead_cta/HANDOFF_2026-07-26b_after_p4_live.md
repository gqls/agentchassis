# HANDOFF — gauntlet_dead_cta after P4 went live (2026-07-26 evening)

**Supersedes `HANDOFF_2026-07-26_gauntlet_p4_frontend_rebuild.md`**, which was
written before P4 and whose Steps 0–3 are now done. Keep it for its API contract
(§6) and its landmine list (§7) — both still accurate — but **do not work from
its step list**.

## 0. Re-read CLAUDE.md first, and re-verify anything load-bearing

Multiple sessions work this repo and cluster. Commit per task by explicit
pathspec, forward-only. And take this document's own advice, which is the lesson
of the last session: **a handoff's identity section survives; its diagnosis
section decays.** Three of the previous handoff's premises were stale within 90
minutes of it being written (see `WRONG_CALLS.md`, 2026-07-26). Re-check before
you plan around anything here.

## 1. State — one paragraph

**The experience is live and works.** A visitor on vonc.com can read today's
provocation, press a button, get a real twenty-minute round against a live AI
opponent, file a position, receive a written counter-argument and a challenge,
send a defence, and get a judged verdict with reasons — all against
`https://tools.apis.uk`. The archive opens full cases at deep-linkable URLs.
Every fabricated statistic is gone from the feed, the component and its schema.
Verified on the deployed pages, desktop and mobile: **72 of 73 checks pass**
(`p4_sources/verify_live.py`, log alongside). The one failure is upstream and
filed as `bugs_open/083`.

## 2. What is live, and how it got there

| piece | state | delivered by |
|---|---|---|
| `/data/provocations.json` | LIVE — honest stats, `slug`+`detail_body`, deep-link urls | `gh api` PUT into `gqls/sites` (commit `0044cc700`) |
| `gauntlet-interface` template + JS + schema | LIVE | `deliver_section_edit.sh` → `republish_gauntlet_js.sh` → reset `pc.build_status` |
| `provocations-archive-loader` | LIVE | `UPDATE js_snippets` → `083_trigger-asset-renderer-vonc.sh` |
| Journeys A, B, C, D | verified live | `p4_sources/verify_live.py` |

Sources, harnesses, logs and the pre-change DB backups: `p4_sources/`.
Commands and every gotcha: `RUNBOOK_gauntlet_dead_cta.md` §§8–9.

## 3. Next actions, highest value first

1. **`bugs_open/083` — make the engine's failures diagnosable, then fix them.**
   This is the only thing between the page and "just works". Both handlers throw
   the upstream error away, so nobody can say why the 503s happen. Start with the
   one-line log; do not act on the truncation theory, which 083 records as
   *not* fitting the evidence. `internal/` → council gate, and the island must be
   rebuilt (RUNBOOK §5) or the fix is inert.
2. **P5 acceptance — fix the harness, not the page.**
   `browserrunner/run_checks_action.go:200` allows `stepDelay = 300ms` before
   asserting; the AI calls take 8–23 s, so the plan's own
   `gauntlet_position_flow` / `gauntlet_defend_flow` would fail a correct page.
   Either add a wait/poll to `criteriaExpect`, or rewrite those two checks to
   assert what is true at 300 ms. **Do not** make the page paint placeholder text
   to satisfy them — that would pass with the engine switched off. Also fold in
   the council's three extra checks (deep-link direct load, the two missing
   selectors, static-entry non-interactivity); the first and third are already
   covered by `verify_live.py` and can be lifted from it.
3. **`claimscan` + a `dead_controls` re-check on `tool-gauntlet`.** Not yet run
   this round. `verify_live.py` asserts no fabricated strings and no dead anchors
   *within the tool*, but that is not the same as the platform's own sweeps.
4. **Step 4, the Arena** — deferred on evidence, with the groundwork done. See
   the PLAN's 2026-07-26 corrections block: mount points enumerated, the inline
   hardcoded-provocation script found, and the invisible "Delusional" reaction
   chip located at its exact CSS rule. **Read the deferral reasoning before
   starting** — the display is easy, but the take-submission flow behind it goes
   nowhere, and making the page load without addressing that trades a visibly
   broken page for a convincingly broken one.

## 4. Two open items that are NOT this workstream's

- **`bugs_open/082` — the clients database is BestEffort and drifted from its own
  manifest.** It crash-looped for ~20 minutes today and blocked this delivery.
  Filed with evidence and a one-command remedy; **deliberately not patched** —
  shared prod infra, owner's call. If you hit unexplained agent failures, check
  this first: `kubectl -n ai-persona-system get endpoints postgres-clients`.
- The `cta_names_unknown_destination` items on `about`, `catalyst` and
  `tool-archetype-taster-quiz` are other components' and remain open by design.

## 5. Landmines learned today (the RUNBOOK has the full list)

- **A spawn fired within ~300 s of a chassis restart is silently dropped** — no
  row, no error. Check the chassis pod's `startedAt` before concluding "queued".
  That is what ate the first dispatch today.
- **`apply_section_edit` does not republish `js_content`**, and **section-editor
  leaves `pc.build_status='approved'`**. Both confirmed again today, in that
  order. Between the two steps the page serves NEW html with OLD js — a real
  broken window, so do them back to back.
- **`.gitignore` silently swallows `build/` and `*.log`.** A workstream directory
  named `build` cannot be committed, and evidence logs vanish from a commit
  without a warning. `git check-ignore -v <path>` before you rely on either.
- **A 403 from `tools.apis.uk` has two senders**: the API's own
  `{"error":"origin not allowed"}`, and Cloudflare's plain-text
  `error code: 1010` for a non-browser fingerprint. Send a browser `User-Agent`
  from any script.
- **`innerText` on a `display:none` element falls back to `textContent`** — so a
  hidden-but-populated region reads as non-empty, and an acceptance check can
  pass without the interaction having done anything. This bit a real check today.

## 6. The one thing worth protecting

Nothing on these pages now claims a number that isn't true by construction, and
no control changes state except as the consequence of a real API response. That
is the whole point of the workstream, and it is easy to lose to a well-meaning
re-render: `input_schema`'s `llm_guidance` now says so explicitly at every field,
which is the only durable defence. If a future run reintroduces a "win rate" or a
leaderboard, the schema is where to look first.
